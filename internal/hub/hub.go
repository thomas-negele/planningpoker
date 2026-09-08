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

// Limits bounds room, connection and seat counts. Startup configuration requires
// positive values; NewManager clamps invalid values defensively.
type Limits struct {
	// Rooms limits rooms held by the manager.
	Rooms int

	// ConnectionsPerRoom counts sockets, including multiple tabs for one participant.
	ConnectionsPerRoom int

	// ParticipantsPerRoom counts all seats, including away participants.
	ParticipantsPerRoom int
}

func (l Limits) valid() bool {
	return l.Rooms > 0 && l.ConnectionsPerRoom > 0 && l.ParticipantsPerRoom > 0
}
