package hub

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"de.thomasnegele.planningpoker/internal/game"
)

// ErrInvalidRoomID re-exports the domain error for invalid room ID syntax.
// Valid user-chosen names are accepted.
var ErrInvalidRoomID = game.ErrInvalidRoomID

// ErrAtCapacity means a new room would exceed the server's room limit.
var ErrAtCapacity = errors.New("the server is holding as many rooms as it can")

// ErrRoomAtCapacity means a room has reached its connection limit.
// It is distinct from the server room limit and the domain's seat limit.
var ErrRoomAtCapacity = errors.New("this room is holding as many connections as it can")

// Manager holds the rooms currently being played in.
type Manager struct {
	// mu serializes room lookup, admission and sweeping. Sweep also queries room
	// occupancy while holding it; rooms never acquire this lock.
	mu    sync.RWMutex
	rooms map[game.RoomID]*Room

	clock  Clock
	random io.Reader
	grace  time.Duration
	limits Limits
}

// NewManager uses clock for expiry and random for identifiers. grace retains
// empty rooms. Nonpositive capacity limits are clamped to one; startup
// configuration rejects them before constructing the manager.
func NewManager(clock Clock, random io.Reader, grace time.Duration, limits Limits) *Manager {
	if !limits.valid() {
		limits = Limits{
			Rooms:               max(limits.Rooms, 1),
			ConnectionsPerRoom:  max(limits.ConnectionsPerRoom, 1),
			ParticipantsPerRoom: max(limits.ParticipantsPerRoom, 1),
		}
	}

	return &Manager{
		rooms:  make(map[game.RoomID]*Room),
		clock:  clock,
		random: random,
		grace:  grace,
		limits: limits,
	}
}

// Create starts an empty room. Creation grants no seat or host privileges.
func (m *Manager) Create() (*Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check capacity and insert under one lock to prevent concurrent over-admission.
	// Check before constructing the room so refusals do not start a goroutine.
	if len(m.rooms) >= m.limits.Rooms {
		return nil, ErrAtCapacity
	}

	room, err := newRoom(m.clock, m.random, m.limits)
	if err != nil {
		return nil, err
	}

	m.rooms[room.ID()] = room
	return room, nil
}

// EnsureRoom returns an existing room or creates an empty one at a valid ID.
// Reusing an address after expiry or restart does not restore its former state.
func (m *Manager) EnsureRoom(id game.RoomID) (*Room, error) {
	// Validate before lookup or allocation; chosen and generated IDs share the grammar.
	if !game.ValidRoomID(id) {
		return nil, ErrInvalidRoomID
	}

	// Keep lookup and creation atomic so concurrent requests share one room.
	m.mu.Lock()
	defer m.mu.Unlock()

	if existing, ok := m.rooms[id]; ok {
		return existing, nil
	}

	// The ceiling restricts new rooms; existing rooms remain accessible.
	if len(m.rooms) >= m.limits.Rooms {
		return nil, ErrAtCapacity
	}

	room, err := newRoomWithID(id, m.clock, m.random, m.limits)
	if err != nil {
		return nil, err
	}
	m.rooms[id] = room
	return room, nil
}

// Lookup returns a room and reports whether it currently exists.
func (m *Manager) Lookup(id game.RoomID) (*Room, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	room, ok := m.rooms[id]
	return room, ok
}

// Len reports how many rooms exist.
func (m *Manager) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.rooms)
}

// Sweep discards rooms with no connections beyond the grace period and reports
// how many it removed. The manager lock serializes lookup and removal; room
// occupancy is queried through each room's command loop.
func (m *Manager) Sweep() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := m.clock.Now()
	discarded := 0

	for id, room := range m.rooms {
		occ, alive := room.Occupancy()
		if alive && (occ.connections > 0 || now.Sub(occ.lastOccupied) <= m.grace) {
			continue
		}

		delete(m.rooms, id)
		room.Stop()
		discarded++
	}

	return discarded
}

// sweepInterval uses one tenth of the grace period, bounded to 100ms–30s.
func sweepInterval(grace time.Duration) time.Duration {
	const (
		shortest = 100 * time.Millisecond
		longest  = 30 * time.Second
	)

	interval := grace / 10
	if interval < shortest {
		return shortest
	}
	if interval > longest {
		return longest
	}
	return interval
}

// Run sweeps until ctx is cancelled. Run it in a dedicated goroutine.
func (m *Manager) Run(ctx context.Context) {
	ticker := time.NewTicker(sweepInterval(m.grace))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.Sweep()
		}
	}
}

// Close removes all current rooms, signals them to stop and waits for their
// goroutines. It does not prevent subsequent calls from creating rooms.
func (m *Manager) Close() {
	m.mu.Lock()
	rooms := make([]*Room, 0, len(m.rooms))
	for _, room := range m.rooms {
		rooms = append(rooms, room)
	}
	m.rooms = make(map[game.RoomID]*Room)
	m.mu.Unlock()

	// Signal all rooms before waiting so they shut down concurrently.
	for _, room := range rooms {
		room.Stop()
	}
	for _, room := range rooms {
		room.Wait()
	}
}
