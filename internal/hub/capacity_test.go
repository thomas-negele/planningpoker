package hub

import (
	"crypto/rand"
	"errors"
	"runtime"
	"sync"
	"testing"
	"time"

	"de.thomasnegele.planningpoker/internal/game"
)

// newBoundedManager builds a manager under chosen ceilings, for the tests that are
// about what happens when one is reached.
func newBoundedManager(t *testing.T, limits Limits) *Manager {
	t.Helper()
	m := NewManager(newFakeClock(), rand.Reader, testGrace, limits)
	t.Cleanup(m.Close)
	return m
}

func TestTheRoomCeilingRefusesRatherThanCreates(t *testing.T) {
	m := newBoundedManager(t, Limits{Rooms: 2, ConnectionsPerRoom: 10, ParticipantsPerRoom: 10})

	for i := 0; i < 2; i++ {
		if _, err := m.Create(); err != nil {
			t.Fatalf("creating room %d of the two allowed: %v", i+1, err)
		}
	}

	if _, err := m.Create(); !errors.Is(err, ErrAtCapacity) {
		t.Fatalf("creating a third room returned %v, want ErrAtCapacity", err)
	}
	if _, err := m.EnsureRoom("team-alpha"); !errors.Is(err, ErrAtCapacity) {
		t.Fatalf("reaching a new name at capacity returned %v, want ErrAtCapacity", err)
	}

	if got := m.Len(); got != 2 {
		t.Errorf("the manager holds %d rooms after two refusals, want 2 — a refusal must "+
			"create nothing and discard nothing", got)
	}
}

func TestAnExistingRoomIsStillReachedAtCapacity(t *testing.T) {
	// A meeting in progress must not be shut out because the process is full. The
	// ceiling refuses the creation of a new room, never the finding of an old one.
	m := newBoundedManager(t, Limits{Rooms: 1, ConnectionsPerRoom: 10, ParticipantsPerRoom: 10})

	first, err := m.EnsureRoom("team-alpha")
	if err != nil {
		t.Fatalf("EnsureRoom: %v", err)
	}

	again, err := m.EnsureRoom("team-alpha")
	if err != nil {
		t.Fatalf("reaching an existing room at capacity was refused: %v", err)
	}
	if again != first {
		t.Error("reaching an existing room at capacity produced a different room")
	}
}

func TestRefusedRoomsAccumulateNothing(t *testing.T) {
	// Whatever a refusal costs, it must not be a cost that grows. If a refused
	// attempt left a room, a goroutine or a map entry behind, a client that keeps
	// trying would keep enlarging the process — which is the attack the ceiling
	// exists to stop, arriving through the ceiling itself.
	m := newBoundedManager(t, Limits{Rooms: 1, ConnectionsPerRoom: 10, ParticipantsPerRoom: 10})
	if _, err := m.Create(); err != nil {
		t.Fatalf("Create: %v", err)
	}

	before := settledGoroutines()
	for i := 0; i < 200; i++ {
		if _, err := m.Create(); !errors.Is(err, ErrAtCapacity) {
			t.Fatalf("attempt %d returned %v, want ErrAtCapacity", i, err)
		}
	}

	if got := m.Len(); got != 1 {
		t.Errorf("the manager holds %d rooms after 200 refusals, want 1", got)
	}
	if after := settledGoroutines(); after > before+2 {
		t.Errorf("200 refusals left goroutines behind: %d before, %d after", before, after)
	}
}

func TestTheRoomCeilingHoldsUnderSimultaneousArrivals(t *testing.T) {
	// Counting under one lock and inserting under another is the race this ceiling
	// would lose: several callers would each see room for one more and each take it.
	// Run under the race detector, this is what proves they share one critical
	// section.
	const ceiling = 5
	m := newBoundedManager(t, Limits{Rooms: ceiling, ConnectionsPerRoom: 10, ParticipantsPerRoom: 10})

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if room, err := m.Create(); err == nil {
				_ = room.ID()
			}
		}()
	}
	wg.Wait()

	if got := m.Len(); got != ceiling {
		t.Errorf("50 simultaneous creations produced %d rooms, want exactly the ceiling of %d",
			got, ceiling)
	}
}

func TestTheConnectionCeilingRefusesWithoutDisturbingAnybody(t *testing.T) {
	m := newBoundedManager(t, Limits{Rooms: 10, ConnectionsPerRoom: 2, ParticipantsPerRoom: 10})
	room, err := m.Create()
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	seated := attach(t, room, "token-one")
	second := attach(t, room, "token-two")

	// Somebody at the table is mid-game, so a refusal has something to disturb.
	room.Seat(seated, "Thomas")
	nextView(t, seated)
	nextView(t, second)
	drain(seated)
	drain(second)

	if got := room.Attach(NewConn("token-three")); got != AttachAtCapacity {
		t.Fatalf("attaching a third connection to a room of two returned %v, want AttachAtCapacity", got)
	}

	// The refusal must be silent to everyone else: no snapshot, nothing to notice.
	assertNoUpdate(t, seated)
	assertNoUpdate(t, second)

	// And the connections already attached must still work.
	room.Vote(seated, game.CardM)
	view := nextView(t, seated)
	if len(view.Participants) != 1 || !view.Participants[0].Voted {
		t.Errorf("after a refused attach the seated participant could not vote: %+v", view.Participants)
	}
}

func TestAStoppedRoomAndAFullRoomAreDistinguishable(t *testing.T) {
	// The page does opposite things with these: reconnect after a stopped room,
	// because reconnecting works; stop retrying at a full one, because it will not.
	// A single boolean could not tell them apart.
	m := newBoundedManager(t, Limits{Rooms: 10, ConnectionsPerRoom: 1, ParticipantsPerRoom: 10})

	full, err := m.Create()
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	attach(t, full, "token-one")
	if got := full.Attach(NewConn("token-two")); got != AttachAtCapacity {
		t.Errorf("a full room reported %v, want AttachAtCapacity", got)
	}

	stopped, err := m.Create()
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	stopped.Stop()
	stopped.Wait()
	if got := stopped.Attach(NewConn("token-three")); got != AttachRoomStopped {
		t.Errorf("a stopped room reported %v, want AttachRoomStopped", got)
	}
}

func TestAManagerIsNeverUnbounded(t *testing.T) {
	// Configuration refuses a non-positive limit long before it reaches here. This
	// is the guarantee that no path into this package can produce a manager without
	// a ceiling, whatever a future caller passes.
	m := NewManager(newFakeClock(), rand.Reader, testGrace, Limits{})
	defer m.Close()

	if _, err := m.Create(); err != nil {
		t.Fatalf("a manager built from zero limits could not create its first room: %v", err)
	}
	if _, err := m.Create(); !errors.Is(err, ErrAtCapacity) {
		t.Errorf("a manager built from zero limits accepted a second room (%v); it must be "+
			"bounded, not unbounded", err)
	}
}

// assertNoUpdate fails if a connection receives anything at all in a short window.
func assertNoUpdate(t *testing.T, conn *Conn) {
	t.Helper()
	select {
	case u := <-conn.Updates():
		t.Errorf("a connection received an update it should not have: %+v", u)
	case <-time.After(50 * time.Millisecond):
	}
}

// settledGoroutines counts goroutines after giving the scheduler a moment, because a
// raw count taken immediately is noise: goroutines that are already finishing have
// not been reaped yet.
func settledGoroutines() int {
	for i := 0; i < 10; i++ {
		runtime.Gosched()
		time.Sleep(5 * time.Millisecond)
	}
	return runtime.NumGoroutine()
}
