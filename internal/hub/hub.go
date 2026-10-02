// Package hub owns room lifetimes, connections and broadcasts. Each room
// serializes domain changes through one goroutine; the manager coordinates
// room lookup, capacity and expiry.
package hub

import "time"

// Clock supplies time for room expiry and deterministic tests.
type Clock interface {
	Now() time.Time
}

// SystemClock uses the process wall clock.
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now() }

// Limits holds what one server permits: room, connection and seat counts, and
// whether the pile of poo may be thrown. Startup configuration requires positive
// counts; NewManager clamps invalid values defensively.
type Limits struct {
	// Rooms limits rooms held by the manager.
	Rooms int

	// ConnectionsPerRoom counts sockets, including multiple tabs for one participant.
	ConnectionsPerRoom int

	// ParticipantsPerRoom counts all seats, including away participants.
	ParticipantsPerRoom int

	// PooThrows adds ThrowPoo to the objects every room accepts. The zero value
	// leaves it out.
	PooThrows bool
}

func (l Limits) valid() bool {
	return l.Rooms > 0 && l.ConnectionsPerRoom > 0 && l.ParticipantsPerRoom > 0
}
