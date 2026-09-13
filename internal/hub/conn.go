package hub

import (
	"sync"

	"de.thomasnegele.planningpoker/internal/game"
)

// outboundBuffer bounds queued updates per connection. It accommodates one
// simultaneous-vote burst from the default twenty-seat table; a full queue still
// causes the room to drop that connection rather than block other participants.
const outboundBuffer = 32

// Update carries either a room snapshot or a refusal, with recipient identity.
type Update struct {
	// View is set for a room snapshot.
	View *game.View

	// Err is set for a refusal.
	Err error

	// You is the recipient's public participant ID, or empty before seating.
	You game.ParticipantID
}

// Conn is the transport-independent connection handle.
type Conn struct {
	// token is the private seat credential; it is never included in snapshots.
	token string

	updates chan Update
	throws  chan ThrowEvent

	closeOnce sync.Once
	closed    chan struct{}
}

// NewConn creates a connection with a bounded outbound queue.
func NewConn(token string) *Conn {
	return &Conn{
		token:   token,
		updates: make(chan Update, outboundBuffer),
		throws:  make(chan ThrowEvent, throwDeliverySize),
		closed:  make(chan struct{}),
	}
}

// Updates yields snapshots and refusals for the transport to send.
func (c *Conn) Updates() <-chan Update { return c.updates }

// Throws yields best-effort cosmetic events independently of reliable updates.
func (c *Conn) Throws() <-chan ThrowEvent { return c.throws }

// Closed is signalled when the connection should stop.
func (c *Conn) Closed() <-chan struct{} { return c.closed }

// Close signals shutdown; it is idempotent and safe for concurrent callers.
func (c *Conn) Close() {
	c.closeOnce.Do(func() { close(c.closed) })
}

// tryThrow drops cosmetic work on backlog without affecting the connection.
func (c *Conn) tryThrow(event ThrowEvent) {
	select {
	case <-c.closed:
	case c.throws <- event:
	default:
	}
}

// Refuse queues a refusal without going through a room and reports whether it was
// queued.
func (c *Conn) Refuse(err error) bool {
	return c.trySend(Update{Err: err})
}

// trySend attempts a nonblocking enqueue; closed or full connections may reject it.
func (c *Conn) trySend(u Update) bool {
	select {
	case <-c.closed:
		return false
	case c.updates <- u:
		return true
	default:
		return false
	}
}
