package transport

import (
	"testing"
	"time"
)

// steppedClock is a clock the test moves by hand, so a rate measured in seconds can
// be checked in microseconds. A test that proves a rate limit by sleeping is slow
// and flaky at the same time.
type steppedClock struct{ now time.Time }

func (c *steppedClock) Now() time.Time          { return c.now }
func (c *steppedClock) advance(d time.Duration) { c.now = c.now.Add(d) }

func newSteppedBucket(limit RateLimit) (*bucket, *steppedClock) {
	clock := &steppedClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	return newBucket(limit, clock.Now), clock
}

func TestABucketStartsFullSoAPromptClientIsNeverRefused(t *testing.T) {
	b, _ := newSteppedBucket(RateLimit{PerSecond: 10, Burst: 20})

	for i := 0; i < 20; i++ {
		if !b.allow() {
			t.Fatalf("message %d of the burst was refused; a bucket must start full", i+1)
		}
	}
	if b.allow() {
		t.Error("a message past the burst was allowed with no time having passed")
	}
}

func TestTimeBuysBackAllowanceContinuously(t *testing.T) {
	// Not stepped in whole windows: waiting half a second buys half the rate's worth,
	// which is what stops a client from having to wait out a window boundary.
	b, clock := newSteppedBucket(RateLimit{PerSecond: 10, Burst: 10})
	for i := 0; i < 10; i++ {
		b.allow()
	}

	clock.advance(500 * time.Millisecond)
	allowed := 0
	for b.allow() {
		allowed++
	}
	if allowed != 5 {
		t.Errorf("half a second at ten per second bought %d messages, want 5", allowed)
	}
}

func TestTheBurstIsAllowedOnceAndNotRepeatedly(t *testing.T) {
	// The failure a fixed window has: a full window's worth at the end of one and
	// again at the start of the next, which is twice the configured rate. Here the
	// long run is bounded by the refill no matter how the messages are spaced.
	b, clock := newSteppedBucket(RateLimit{PerSecond: 2, Burst: 10})

	allowed := 0
	for second := 0; second < 10; second++ {
		clock.advance(time.Second)
		for b.allow() {
			allowed++
		}
	}

	// Twenty-eight, and the arithmetic is worth spelling out because the obvious
	// answer is thirty. The bucket starts full, so the first second's refill lands in
	// an already-full bucket and is lost: that second yields the ten it was holding,
	// not ten plus two. The nine seconds after it yield two each. Ten plus eighteen.
	//
	// The point is the shape rather than the number: the burst is spent once, and
	// from then on the rate is the rate. A fixed window would have allowed twenty on
	// either side of a boundary.
	if want := 10 + 9*2; allowed != want {
		t.Errorf("ten seconds allowed %d messages, want %d — the burst must be a one-off, "+
			"not something earned again every interval", allowed, want)
	}

	if ceiling := 10 + 10*2; allowed > ceiling {
		t.Errorf("allowed %d messages, which is above burst plus rate times elapsed (%d)",
			allowed, ceiling)
	}
}

func TestABucketIsNeverUnbounded(t *testing.T) {
	// Configuration refuses a non-positive rate long before it reaches here. This is
	// the guarantee that no path produces a bucket that either bounds nothing or
	// blocks everything.
	b, clock := newSteppedBucket(RateLimit{PerSecond: 0, Burst: 0})

	if !b.allow() {
		t.Error("a bucket built from zero limits refused its first message; it must bound, not block")
	}
	if b.allow() {
		t.Error("a bucket built from zero limits allowed a second message with no time passing")
	}

	clock.advance(time.Second)
	if !b.allow() {
		t.Error("a bucket built from zero limits never refilled")
	}
}
