package transport

import "time"

// RateLimit bounds messages per connection, independently of other participants.
type RateLimit struct {
	// PerSecond is the sustained message rate.
	PerSecond int

	// Burst is the maximum immediate allowance.
	Burst int
}

// bucket refills continuously up to capacity; each message consumes one token.
// Only the connection's reader goroutine accesses it, so it needs no lock.
type bucket struct {
	capacity float64
	refill   float64
	tokens   float64
	last     time.Time
	now      func() time.Time
}

// newBucket starts with the full burst allowance. now allows deterministic tests.
func newBucket(limit RateLimit, now func() time.Time) *bucket {
	if now == nil {
		now = time.Now
	}

	// Clamp invalid limits defensively; startup configuration rejects them.
	capacity := float64(max(limit.Burst, 1))
	refill := float64(max(limit.PerSecond, 1))

	return &bucket{
		capacity: capacity,
		refill:   refill,
		tokens:   capacity,
		last:     now(),
		now:      now,
	}
}

// burstSize supplies the threshold for closing a connection after repeated refusals.
func (b *bucket) burstSize() int { return int(b.capacity) }

// allow refills for elapsed time and consumes a token if available.
func (b *bucket) allow() bool {
	now := b.now()

	if elapsed := now.Sub(b.last); elapsed > 0 {
		b.tokens = min(b.capacity, b.tokens+elapsed.Seconds()*b.refill)
		b.last = now
	}

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}
