package hub

import (
	"io"
	"sync"
	"time"

	"de.thomasnegele.planningpoker/internal/game"
)

// commandBuffer bounds pending commands; senders wait when the queue is full.
const commandBuffer = 32

// Room serializes domain state and connection changes in its own goroutine.
type Room struct {
	id game.RoomID

	commands chan command
	stop     chan struct{}
	done     chan struct{}
	stopOnce sync.Once
}

// ID is immutable and safe to read outside the room goroutine.
func (r *Room) ID() game.RoomID { return r.id }

// roomState is accessed only by the room goroutine.
type roomState struct {
	game   *game.Room
	clock  Clock
	random io.Reader

	// A connection has an empty participant ID until seated.
	conns map[*Conn]game.ParticipantID

	// seats maps private credentials to public participant IDs; it is never broadcast.
	seats map[string]game.ParticipantID

	// lastOccupied starts at creation and updates when the last connection closes.
	lastOccupied time.Time

	// maxConnections counts sockets, including multiple tabs for one seat.
	maxConnections int
}

// newRoom creates a room with a generated ID.
func newRoom(clock Clock, random io.Reader, limits Limits) (*Room, error) {
	return newRoomWithDeck(clock, random, limits, game.TShirtDeckName)
}

// newRoomWithDeck creates a room with a generated ID and a supported deck.
func newRoomWithDeck(clock Clock, random io.Reader, limits Limits, deckName string) (*Room, error) {
	g, err := game.NewRoomWithDeck(random, limits.ParticipantsPerRoom, deckName)
	if err != nil {
		return nil, err
	}
	return startRoom(g, g.ID(), clock, random, limits), nil
}

// newRoomWithID creates a room at a validated ID.
func newRoomWithID(id game.RoomID, clock Clock, random io.Reader, limits Limits) (*Room, error) {
	g, err := game.NewRoomAt(id, limits.ParticipantsPerRoom)
	if err != nil {
		return nil, err
	}
	return startRoom(g, id, clock, random, limits), nil
}

// startRoom starts the sole owner of the room state.
func startRoom(g *game.Room, id game.RoomID, clock Clock, random io.Reader, limits Limits) *Room {
	r := &Room{
		id:       id,
		commands: make(chan command, commandBuffer),
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}

	state := &roomState{
		game:           g,
		clock:          clock,
		random:         random,
		conns:          make(map[*Conn]game.ParticipantID),
		seats:          make(map[string]game.ParticipantID),
		lastOccupied:   clock.Now(),
		maxConnections: limits.ConnectionsPerRoom,
	}

	go r.run(state)
	return r
}

// run processes commands until stopped.
func (r *Room) run(s *roomState) {
	defer close(r.done)

	for {
		select {
		case c := <-r.commands:
			c.apply(s)
		case <-r.stop:
			// Closing connections wakes their transport handlers.
			for c := range s.conns {
				c.Close()
			}
			return
		}
	}
}

// Stop signals shutdown and is idempotent. Use Wait to await completion.
func (r *Room) Stop() {
	r.stopOnce.Do(func() { close(r.stop) })
}

// Wait blocks until the room goroutine exits.
func (r *Room) Wait() { <-r.done }

// send queues a command or returns false if shutdown is observed.
func (r *Room) send(c command) bool {
	// Prefer an already completed shutdown over an available queue slot.
	select {
	case <-r.done:
		return false
	default:
	}

	select {
	case r.commands <- c:
		return true
	case <-r.done:
		return false
	}
}

// AttachResult distinguishes successful attachment, shutdown and capacity refusal.
type AttachResult int

const (
	// Attached means the connection was registered and its initial snapshot attempted.
	Attached AttachResult = iota

	// AttachRoomStopped means the room was discarded or the process is shutting
	// down. The page should reconnect; the next attempt creates the room again.
	AttachRoomStopped

	// AttachAtCapacity means no connection slot is available; existing clients remain
	// attached.
	AttachAtCapacity
)

// Attach registers a connection unless the room is stopped or at capacity.
func (r *Room) Attach(c *Conn) AttachResult {
	reply := make(chan AttachResult, 1)
	if !r.send(attachCommand{conn: c, reply: reply}) {
		return AttachRoomStopped
	}
	select {
	case result := <-reply:
		return result
	case <-r.done:
		return AttachRoomStopped
	}
}

// Detach removes a connection and marks its participant away only when no other
// connection remains for that seat.
func (r *Room) Detach(c *Conn) bool { return r.send(detachCommand{conn: c}) }

// Seat restores a known credential's seat or creates a new one.
func (r *Room) Seat(c *Conn, name string) bool { return r.send(seatCommand{conn: c, name: name}) }

// Vote plays a card.
func (r *Room) Vote(c *Conn, card game.Card) bool { return r.send(voteCommand{conn: c, card: card}) }

// Reveal shows the round's votes.
func (r *Room) Reveal(c *Conn) bool { return r.send(revealCommand{conn: c}) }

// NewRound starts a fresh hidden round.
func (r *Room) NewRound(c *Conn) bool { return r.send(newRoundCommand{conn: c}) }

// SetDeck changes the current or next-round deck under the domain rules.
func (r *Room) SetDeck(c *Conn, deckName string) bool {
	return r.send(setDeckCommand{conn: c, deckName: deckName})
}

// Rename changes a participant's display name.
func (r *Room) Rename(c *Conn, name string) bool { return r.send(renameCommand{conn: c, name: name}) }

// occupancy is a room's answer to "is anyone still connected, and if not, since
// when?".
type occupancy struct {
	connections  int
	lastOccupied time.Time
}

// Occupancy queries connection count and the time the room became empty.
// It reports false when the query cannot complete because the room stopped.
func (r *Room) Occupancy() (occupancy, bool) {
	reply := make(chan occupancy, 1)
	if !r.send(occupancyCommand{reply: reply}) {
		return occupancy{}, false
	}
	select {
	case o := <-reply:
		return o, true
	case <-r.done:
		return occupancy{}, false
	}
}

// command runs on the room's owning goroutine.
type command interface{ apply(*roomState) }

type attachCommand struct {
	conn  *Conn
	reply chan AttachResult
}

func (c attachCommand) apply(s *roomState) {
	// Check the connection limit before registering or broadcasting.
	if len(s.conns) >= s.maxConnections {
		c.reply <- AttachAtCapacity
		return
	}

	s.conns[c.conn] = ""

	// Restore the seat associated with this credential, if any.
	if pid, known := s.seats[c.conn.token]; known {
		if err := s.game.MarkPresent(pid); err == nil {
			s.conns[c.conn] = pid
		}
	}

	s.broadcast()
	c.reply <- Attached
}

type detachCommand struct{ conn *Conn }

func (c detachCommand) apply(s *roomState) {
	pid, wasSeated := s.conns[c.conn]
	delete(s.conns, c.conn)
	c.conn.Close()

	if len(s.conns) == 0 {
		s.lastOccupied = s.clock.Now()
	}

	// Only the participant's last connection marks them away.
	if wasSeated && pid != "" && !s.hasConnection(pid) {
		_ = s.game.MarkAway(pid)
	}

	s.broadcast()
}

type seatCommand struct {
	conn *Conn
	name string
}

func (c seatCommand) apply(s *roomState) {
	if pid, known := s.seats[c.conn.token]; known {
		// Rejoining also updates the display name.
		if err := s.game.Rejoin(pid, c.name); err != nil {
			s.refuse(c.conn, err)
			return
		}
		s.conns[c.conn] = pid
		s.broadcast()
		return
	}

	pid, err := s.game.Join(s.random, c.name)
	if err != nil {
		s.refuse(c.conn, err)
		return
	}
	s.seats[c.conn.token] = pid
	s.conns[c.conn] = pid
	s.broadcast()
}

type voteCommand struct {
	conn *Conn
	card game.Card
}

func (c voteCommand) apply(s *roomState) {
	s.act(c.conn, func(pid game.ParticipantID) error { return s.game.Vote(pid, c.card) })
}

type revealCommand struct{ conn *Conn }

func (c revealCommand) apply(s *roomState) {
	s.act(c.conn, func(pid game.ParticipantID) error { return s.game.Reveal(pid) })
}

type newRoundCommand struct{ conn *Conn }

func (c newRoundCommand) apply(s *roomState) {
	s.act(c.conn, func(pid game.ParticipantID) error { return s.game.NewRound(pid) })
}

type setDeckCommand struct {
	conn     *Conn
	deckName string
}

func (c setDeckCommand) apply(s *roomState) {
	s.act(c.conn, func(pid game.ParticipantID) error { return s.game.SetDeck(pid, c.deckName) })
}

type renameCommand struct {
	conn *Conn
	name string
}

func (c renameCommand) apply(s *roomState) {
	s.act(c.conn, func(pid game.ParticipantID) error { return s.game.Rename(pid, c.name) })
}

type occupancyCommand struct{ reply chan occupancy }

func (c occupancyCommand) apply(s *roomState) {
	c.reply <- occupancy{connections: len(s.conns), lastOccupied: s.lastOccupied}
}

// act checks that the connection has a seat before applying a domain action.
func (s *roomState) act(conn *Conn, do func(game.ParticipantID) error) {
	pid, ok := s.conns[conn]
	if !ok || pid == "" {

		s.refuse(conn, game.ErrUnknownParticipant)
		return
	}

	if err := do(pid); err != nil {
		s.refuse(conn, err)
		return
	}
	s.broadcast()
}

// broadcast shares one snapshot across connections. Recipients must treat it as
// read-only.
func (s *roomState) broadcast() {
	view := s.game.View()

	for conn, pid := range s.conns {
		if !conn.trySend(Update{View: &view, You: pid}) {
			// Drop slow connections instead of blocking the room on a full outbound
			// queue.
			s.dropLocked(conn, pid)
		}
	}
}

// refuse sends an error only to the affected connection.
func (s *roomState) refuse(conn *Conn, err error) {
	if !conn.trySend(Update{Err: err, You: s.conns[conn]}) {
		s.dropLocked(conn, s.conns[conn])
	}
}

// dropLocked runs on the room goroutine; the name does not imply a mutex.
func (s *roomState) dropLocked(conn *Conn, pid game.ParticipantID) {
	delete(s.conns, conn)
	conn.Close()

	if len(s.conns) == 0 {
		s.lastOccupied = s.clock.Now()
	}
	if pid != "" && !s.hasConnection(pid) {
		_ = s.game.MarkAway(pid)
	}
}

func (s *roomState) hasConnection(pid game.ParticipantID) bool {
	for _, other := range s.conns {
		if other == pid {
			return true
		}
	}
	return false
}
