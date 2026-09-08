package hub

import (
	"crypto/rand"
	"sync"
	"testing"
	"time"

	"de.thomasnegele.planningpoker/internal/game"
)

// fakeClock is a clock the test moves by hand, so that a rule measured in minutes
// can be checked in microseconds.
type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

const testGrace = 5 * time.Minute

// testLimits are deliberately far above anything these tests reach, so that a test
// about expiry or reseating fails for its own reason and never because it quietly
// ran into a ceiling. The tests that are about the ceilings set their own.
var testLimits = Limits{Rooms: 1000, ConnectionsPerRoom: 1000, ParticipantsPerRoom: 1000}

func newTestManager(t *testing.T) (*Manager, *fakeClock) {
	t.Helper()
	clock := newFakeClock()
	return NewManager(clock, rand.Reader, testGrace, testLimits), clock
}

// attachView connects a fresh connection carrying the given seat token and returns
// it together with the snapshot that attaching always produces.
func attachView(t *testing.T, room *Room, token string) (*Conn, game.View) {
	t.Helper()
	conn := NewConn(token)
	if room.Attach(conn) != Attached {
		t.Fatalf("Attach to room %q was refused", room.ID())
	}
	return conn, nextView(t, conn)
}

// attach is attachView for the tests that do not care about the first snapshot.
func attach(t *testing.T, room *Room, token string) *Conn {
	t.Helper()
	conn, _ := attachView(t, room, token)
	return conn
}

// nextUpdate waits for one update, failing rather than hanging if none arrives.
func nextUpdate(t *testing.T, conn *Conn) Update {
	t.Helper()
	select {
	case u := <-conn.Updates():
		return u
	case <-time.After(2 * time.Second):
		t.Fatal("no update arrived within two seconds")
		return Update{}
	}
}

func nextView(t *testing.T, conn *Conn) game.View {
	t.Helper()
	u := nextUpdate(t, conn)
	if u.View == nil {
		t.Fatalf("expected a snapshot, got a refusal: %v", u.Err)
	}
	return *u.View
}

func nextRefusal(t *testing.T, conn *Conn) error {
	t.Helper()
	u := nextUpdate(t, conn)
	if u.Err == nil {
		t.Fatalf("expected a refusal, got a snapshot: %+v", u.View)
	}
	return u.Err
}

// drain consumes any updates already queued, so a later assertion sees only what a
// specific action produced.
func drain(conn *Conn) {
	for {
		select {
		case <-conn.Updates():
		default:
			return
		}
	}
}

func TestCommandChangesTheRoom(t *testing.T) {
	manager, _ := newTestManager(t)
	defer manager.Close()

	room, err := manager.Create()
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	conn := attach(t, room, "token-a")
	if !room.Seat(conn, "Thomas") {
		t.Fatal("Seat was refused")
	}

	view := nextView(t, conn)
	if len(view.Participants) != 1 || view.Participants[0].Name != "Thomas" {
		t.Fatalf("participants = %+v, want one named Thomas", view.Participants)
	}
	if view.Participants[0].ID == "" {
		t.Error("the seated participant has no identifier")
	}
}

func TestSnapshotArrivesOnAttachBeforeAnythingElse(t *testing.T) {
	manager, _ := newTestManager(t)
	defer manager.Close()
	room, _ := manager.Create()

	// Somebody is already seated and has voted before the second browser connects.
	first := attach(t, room, "token-a")
	room.Seat(first, "Thomas")
	nextView(t, first)
	room.Vote(first, game.CardM)
	nextView(t, first)

	late := NewConn("token-b")
	if room.Attach(late) != Attached {
		t.Fatal("Attach was refused")
	}

	view := nextView(t, late)
	if len(view.Participants) != 1 {
		t.Fatalf("the first message shows %d participants, want 1 — a connection must get "+
			"a complete picture immediately", len(view.Participants))
	}
	if !view.Participants[0].Voted {
		t.Error("the first message does not show that the seated participant has voted")
	}
}

func TestEveryChangeReachesEveryConnection(t *testing.T) {
	manager, _ := newTestManager(t)
	defer manager.Close()
	room, _ := manager.Create()

	conns := []*Conn{
		attach(t, room, "token-a"),
		attach(t, room, "token-b"),
		attach(t, room, "token-c"),
	}
	for _, c := range conns {
		drain(c)
	}

	room.Seat(conns[0], "Thomas")

	for i, c := range conns {
		view := nextView(t, c)
		if len(view.Participants) != 1 {
			t.Errorf("connection %d sees %d participants, want 1", i, len(view.Participants))
		}
	}
}

func TestAStuckConnectionDoesNotStallTheRoom(t *testing.T) {
	manager, _ := newTestManager(t)
	defer manager.Close()
	room, _ := manager.Create()

	stuck := attach(t, room, "token-stuck")
	healthy := attach(t, room, "token-healthy")
	drain(healthy)

	// The stuck connection never reads. Push far more than its buffer holds.
	room.Seat(healthy, "Thomas")
	nextView(t, healthy)
	for i := range outboundBuffer * 3 {
		card := game.TShirtDeck().Cards[i%len(game.TShirtDeck().Cards)]
		room.Vote(healthy, card)
		// The healthy connection keeps up throughout, which is the point: the room is
		// still serving it while the other one is hopelessly behind.
		nextView(t, healthy)
	}

	select {
	case <-stuck.Closed():
	case <-time.After(2 * time.Second):
		t.Fatal("the stuck connection was not dropped; the room would eventually block on it")
	}
}

func TestOccupancy(t *testing.T) {
	manager, clock := newTestManager(t)
	defer manager.Close()
	room, _ := manager.Create()

	// A room nobody ever joined reports no connections, dated from its creation, so
	// it expires on the same rule as one everybody left.
	occ, alive := room.Occupancy()
	if !alive {
		t.Fatal("a new room reports itself stopped")
	}
	if occ.connections != 0 {
		t.Errorf("connections = %d, want 0", occ.connections)
	}
	if occ.lastOccupied != clock.Now() {
		t.Errorf("lastOccupied = %v, want the creation time %v", occ.lastOccupied, clock.Now())
	}

	conn := attach(t, room, "token-a")
	if occ, _ := room.Occupancy(); occ.connections != 1 {
		t.Errorf("connections with one attached = %d, want 1", occ.connections)
	}

	clock.Advance(time.Minute)
	room.Detach(conn)
	nextViewOrClosed(t, conn)

	occ, _ = room.Occupancy()
	if occ.connections != 0 {
		t.Errorf("connections after detaching = %d, want 0", occ.connections)
	}
	if occ.lastOccupied != clock.Now() {
		t.Errorf("lastOccupied = %v, want the time of the last detach %v", occ.lastOccupied, clock.Now())
	}
}

// nextViewOrClosed tolerates the connection having been closed, which is what
// detaching does to it.
// settle waits until the room's goroutine has worked through everything queued
// before this point.
//
// It exists because watching a connection is not a reliable signal that a command
// finished. Detaching closes the connection *before* it records when the room was
// last occupied, so a test that waits for the close and then moves the clock can
// move it into the very line it was waiting for — and the room then looks freshly
// occupied. Occupancy is a command like any other and is answered in order, so
// asking for it is a barrier: when the answer arrives, everything queued ahead of it
// has been applied.
func settle(t *testing.T, room *Room) {
	t.Helper()
	if _, alive := room.Occupancy(); !alive {
		t.Fatal("the room stopped while waiting for it to settle")
	}
}

func nextViewOrClosed(t *testing.T, conn *Conn) {
	t.Helper()
	select {
	case <-conn.Updates():
	case <-conn.Closed():
	case <-time.After(2 * time.Second):
		t.Fatal("neither an update nor a close arrived")
	}
}

func TestCreateAndLookup(t *testing.T) {
	manager, _ := newTestManager(t)
	defer manager.Close()

	first, err := manager.Create()
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	second, err := manager.Create()
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if first.ID() == second.ID() {
		t.Fatal("two games produced the same room identifier")
	}

	found, ok := manager.Lookup(first.ID())
	if !ok || found != first {
		t.Error("looking up a room that exists did not return it")
	}
	if _, ok := manager.Lookup("NEVEREXISTED"); ok {
		t.Error("looking up an identifier that never existed returned a room")
	}
}

func TestCreatingAGameSeatsNobody(t *testing.T) {
	manager, _ := newTestManager(t)
	defer manager.Close()
	room, _ := manager.Create()

	_, view := attachView(t, room, "token-a")
	if len(view.Participants) != 0 {
		t.Errorf("a freshly created game has %d participants, want 0", len(view.Participants))
	}
}

func TestVotesDoNotCrossBetweenRooms(t *testing.T) {
	manager, _ := newTestManager(t)
	defer manager.Close()

	first, _ := manager.Create()
	second, _ := manager.Create()

	a := attach(t, first, "token-a")
	first.Seat(a, "Thomas")
	nextView(t, a)
	first.Vote(a, game.CardM)
	nextView(t, a)

	_, view := attachView(t, second, "token-b")
	if len(view.Participants) != 0 {
		t.Errorf("the second room shows %d participants, want 0 — rooms must be independent",
			len(view.Participants))
	}
}

func TestAStoppedRoomRefusesCommandsRatherThanHanging(t *testing.T) {
	manager, _ := newTestManager(t)
	room, _ := manager.Create()
	conn := attach(t, room, "token-a")

	manager.Close()
	room.Wait()

	done := make(chan struct{})
	go func() {
		defer close(done)
		if room.Seat(conn, "Thomas") {
			t.Error("a stopped room accepted a command")
		}
		if _, alive := room.Occupancy(); alive {
			t.Error("a stopped room answered an occupancy query")
		}
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("commanding a stopped room hung instead of being refused")
	}
}

func TestShutdownClosesConnections(t *testing.T) {
	manager, _ := newTestManager(t)
	room, _ := manager.Create()
	first := attach(t, room, "token-a")
	second := attach(t, room, "token-b")

	manager.Close()

	for i, conn := range []*Conn{first, second} {
		select {
		case <-conn.Closed():
		case <-time.After(2 * time.Second):
			t.Errorf("connection %d was not closed by shutdown", i)
		}
	}
	if manager.Len() != 0 {
		t.Errorf("%d rooms remain after shutdown, want 0", manager.Len())
	}
}
