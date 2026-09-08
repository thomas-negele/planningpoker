package hub

import (
	"crypto/rand"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"de.thomasnegele.planningpoker/internal/game"
)

func TestRoomSurvivesJustUnderTheGracePeriodAndIsGoneJustOver(t *testing.T) {
	manager, clock := newTestManager(t)
	defer manager.Close()

	room, _ := manager.Create()
	id := room.ID()

	// Just under: the room is still there. This direction matters as much as the
	// other — a room expiring too eagerly looks like a bug in something else
	// entirely.
	clock.Advance(testGrace - time.Second)
	if discarded := manager.Sweep(); discarded != 0 {
		t.Errorf("sweep discarded %d rooms just under the grace period, want 0", discarded)
	}
	if _, ok := manager.Lookup(id); !ok {
		t.Fatal("the room was discarded before its grace period had elapsed")
	}

	// Just over: it is gone.
	clock.Advance(2 * time.Second)
	if discarded := manager.Sweep(); discarded != 1 {
		t.Errorf("sweep discarded %d rooms just over the grace period, want 1", discarded)
	}
	if _, ok := manager.Lookup(id); ok {
		t.Error("the room survived past its grace period")
	}
}

func TestARoomNobodyEverJoinedIsDiscardedOnTheSameRule(t *testing.T) {
	manager, clock := newTestManager(t)
	defer manager.Close()

	room, _ := manager.Create()
	id := room.ID()

	clock.Advance(testGrace + time.Second)
	manager.Sweep()

	if _, ok := manager.Lookup(id); ok {
		t.Error("a room nobody ever joined was not discarded")
	}
}

func TestAReloadDoesNotDestroyTheRoom(t *testing.T) {
	manager, clock := newTestManager(t)
	defer manager.Close()

	room, _ := manager.Create()
	id := room.ID()

	conn := attach(t, room, "token-a")
	room.Seat(conn, "Thomas")
	nextView(t, conn)
	room.Vote(conn, game.CardL)
	nextView(t, conn)

	// A reload: the connection closes, and for a moment the room has nobody in it.
	room.Detach(conn)
	nextViewOrClosed(t, conn)
	settle(t, room)

	clock.Advance(testGrace / 2)
	manager.Sweep()

	found, ok := manager.Lookup(id)
	if !ok {
		t.Fatal("a reload within the grace period destroyed the room")
	}

	// The same browser comes back with the token it already held.
	_, view := attachView(t, found, "token-a")
	if len(view.Participants) != 1 {
		t.Fatalf("the table shows %d participants after a reload, want 1", len(view.Participants))
	}
	if !view.Participants[0].Voted {
		t.Error("the vote was lost across the reload")
	}
	if view.Participants[0].Away {
		t.Error("the participant is still marked away after reconnecting")
	}
	if view.Participants[0].Name != "Thomas" {
		t.Errorf("name after reload = %q, want %q", view.Participants[0].Name, "Thomas")
	}
}

func TestAnOccupiedRoomIsNeverDiscarded(t *testing.T) {
	manager, clock := newTestManager(t)
	defer manager.Close()

	room, _ := manager.Create()
	id := room.ID()
	conn := attach(t, room, "token-a")
	room.Seat(conn, "Thomas")
	nextView(t, conn)

	// However long the game runs, somebody is connected throughout.
	for range 10 {
		clock.Advance(testGrace * 2)
		manager.Sweep()
		if _, ok := manager.Lookup(id); !ok {
			t.Fatal("an occupied room was discarded")
		}
	}

	drain(conn)
}

func TestADiscardedIdentifierReachesNothingAndIsNeverRecreated(t *testing.T) {
	manager, clock := newTestManager(t)
	defer manager.Close()

	room, _ := manager.Create()
	id := room.ID()

	clock.Advance(testGrace + time.Second)
	manager.Sweep()

	if _, ok := manager.Lookup(id); ok {
		t.Fatal("the discarded identifier still resolves")
	}

	// Creating a game must never hand an old identifier back. Reopen deliberately can
	// place a room at a used identifier — that is what lets an interrupted group
	// return to their link — but Create invents one, and inventing one that has been
	// seen before would put strangers in the same room by accident.
	for range 50 {
		next, err := manager.Create()
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if next.ID() == id {
			t.Fatal("a new game was created at a previously used identifier")
		}
	}
	if _, ok := manager.Lookup(id); ok {
		t.Error("the discarded identifier resolves again after other rooms were created")
	}
}

func TestEnsureRoomAcceptsAChosenName(t *testing.T) {
	// A memorable address for a recurring meeting is the point of admitting names at
	// all, and it is not private — anybody who guesses it is in that room.
	manager, _ := newTestManager(t)
	defer manager.Close()

	room, err := manager.EnsureRoom("team-alpha")
	if err != nil {
		t.Fatalf("EnsureRoom: %v", err)
	}
	if room.ID() != "team-alpha" {
		t.Errorf("room id = %q, want the name that was asked for", room.ID())
	}

	// Asking again finds the same room rather than replacing it.
	again, err := manager.EnsureRoom("team-alpha")
	if err != nil || again != room {
		t.Errorf("asking a second time produced a different room (%v)", err)
	}
}

func TestTwoSpellingsAreTwoRooms(t *testing.T) {
	// Nothing is folded, so the address in the URL bar is the address.
	manager, _ := newTestManager(t)
	defer manager.Close()

	upper, err := manager.EnsureRoom("Team-Alpha")
	if err != nil {
		t.Fatalf("EnsureRoom: %v", err)
	}
	lower, err := manager.EnsureRoom("team-alpha")
	if err != nil {
		t.Fatalf("EnsureRoom: %v", err)
	}

	if upper == lower {
		t.Error("two spellings produced one room; they must be two")
	}
	if manager.Len() != 2 {
		t.Errorf("%d rooms exist, want 2", manager.Len())
	}
}

func TestOnlyEnsureRoomEverReusesAnIdentifier(t *testing.T) {
	// The distinction the test above rests on, stated directly: creating never reuses,
	// reopening is the one door that does.
	manager, clock := newTestManager(t)
	defer manager.Close()

	seed, _ := manager.Create()
	id := seed.ID()
	clock.Advance(testGrace + time.Second)
	manager.Sweep()

	if _, ok := manager.Lookup(id); ok {
		t.Fatal("the identifier was not discarded")
	}
	if _, err := manager.EnsureRoom(id); err != nil {
		t.Fatalf("EnsureRoom: %v", err)
	}
	if _, ok := manager.Lookup(id); !ok {
		t.Error("reopening did not put a room back at that identifier")
	}
}

func TestSweepIntervalStaysWellBelowTheGracePeriod(t *testing.T) {
	// The interval only trades a little idle work against a little imprecision, but
	// it must not be so long that a room noticeably outlives its grace period.
	for _, grace := range []time.Duration{time.Second, time.Minute, 5 * time.Minute, time.Hour} {
		interval := sweepInterval(grace)
		if interval > grace {
			t.Errorf("grace %s gives a sweep interval of %s, which is longer than the grace period itself",
				grace, interval)
		}
		if interval <= 0 {
			t.Errorf("grace %s gives a non-positive sweep interval %s", grace, interval)
		}
	}
}

func TestSeatingIdentityAndConnections(t *testing.T) {
	manager, _ := newTestManager(t)
	defer manager.Close()
	room, _ := manager.Create()

	// A browser with an unknown token takes a new seat.
	first := attach(t, room, "token-a")
	room.Seat(first, "Thomas")
	view := nextView(t, first)
	if len(view.Participants) != 1 {
		t.Fatalf("participants = %d, want 1", len(view.Participants))
	}
	thomas := view.Participants[0].ID

	// A second tab of the same browser presents the same token. One seat, two
	// connections.
	second, view := attachView(t, room, "token-a")
	if len(view.Participants) != 1 {
		t.Errorf("a second tab produced %d participants, want 1", len(view.Participants))
	}
	drain(first)

	// An action on one connection appears on the other.
	room.Vote(second, game.CardM)
	if v := nextView(t, first); !v.Participants[0].Voted {
		t.Error("a vote in the second tab did not appear in the first")
	}
	drain(second)

	// Closing one of the two connections must not mark the participant away.
	room.Detach(first)
	view = nextView(t, second)
	if view.Participants[0].Away {
		t.Error("closing one of two connections marked the participant away")
	}

	// Closing the last one does.
	third, _ := attachView(t, room, "token-c")
	drain(third)
	room.Detach(second)
	view = nextView(t, third)
	if !view.Participants[0].Away {
		t.Error("closing the last connection did not mark the participant away")
	}
	if view.Participants[0].ID != thomas {
		t.Error("the participant identifier changed")
	}
	if !view.Participants[0].Voted {
		t.Error("going away discarded the vote")
	}
}

func TestATokenFromAnotherRoomStartsANewSeat(t *testing.T) {
	manager, _ := newTestManager(t)
	defer manager.Close()

	first, _ := manager.Create()
	second, _ := manager.Create()

	a := attach(t, first, "token-shared")
	first.Seat(a, "Thomas")
	nextView(t, a)

	// The same token in a different room names nobody there, so it is a new arrival
	// rather than an error.
	b, view := attachView(t, second, "token-shared")
	if len(view.Participants) != 0 {
		t.Errorf("the second room already shows %d participants, want 0", len(view.Participants))
	}
	second.Seat(b, "Thomas")
	view = nextView(t, b)
	if len(view.Participants) != 1 {
		t.Fatalf("the second room shows %d participants after seating, want 1", len(view.Participants))
	}
}

func TestAnUnseatedConnectionMayNotAct(t *testing.T) {
	manager, _ := newTestManager(t)
	defer manager.Close()
	room, _ := manager.Create()

	conn := attach(t, room, "token-a")

	for name, act := range map[string]func() bool{
		"vote":     func() bool { return room.Vote(conn, game.CardM) },
		"reveal":   func() bool { return room.Reveal(conn) },
		"newRound": func() bool { return room.NewRound(conn) },
		"rename":   func() bool { return room.Rename(conn, "Mallory") },
	} {
		if !act() {
			t.Fatalf("%s: the room refused the command outright", name)
		}
		if err := nextRefusal(t, conn); !errors.Is(err, game.ErrUnknownParticipant) {
			t.Errorf("%s by an unseated connection: error = %v, want ErrUnknownParticipant", name, err)
		}
	}
}

func TestARefusalIsPrivateAndChangesNothing(t *testing.T) {
	manager, _ := newTestManager(t)
	defer manager.Close()
	room, _ := manager.Create()

	actor := attach(t, room, "token-a")
	room.Seat(actor, "Thomas")
	nextView(t, actor)

	observer, _ := attachView(t, room, "token-b")
	drain(actor)
	drain(observer)

	// A card the deck does not contain.
	room.Vote(actor, "XXL")
	if err := nextRefusal(t, actor); err == nil {
		t.Error("an invalid card was not refused")
	}

	// Nobody else hears about it, and no snapshot is sent on its account.
	select {
	case u := <-observer.Updates():
		t.Errorf("the observer received %+v; a refusal must reach only the connection that caused it", u)
	case <-time.After(200 * time.Millisecond):
	}
}

func TestConcurrentActivityKeepsTheRoomConsistent(t *testing.T) {
	// The whole design rests on room state having a single owner. A test with two
	// goroutines proves very little, so this one runs many connections doing
	// conflicting things at once, under the race detector.
	manager := NewManager(SystemClock{}, rand.Reader, time.Hour, testLimits)
	defer manager.Close()
	room, err := manager.Create()
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	const participants = 12
	const rounds = 25

	var wg sync.WaitGroup
	conns := make([]*Conn, participants)

	for i := range participants {
		conn := NewConn(string(rune('a'+i)) + "-token")
		conns[i] = conn
		if room.Attach(conn) != Attached {
			t.Fatalf("Attach: refused")
		}

		// Every connection reads continuously, so nothing is dropped for being slow and
		// the test exercises the room rather than the overflow path.
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case u, ok := <-conn.Updates():
					if !ok {
						return
					}
					if u.View != nil && u.View.Results == nil {
						// A hidden round must never carry results.
						for _, p := range u.View.Participants {
							_ = p
						}
					}
				case <-conn.Closed():
					return
				}
			}
		}()
	}

	deck := game.TShirtDeck().Cards
	for i, conn := range conns {
		wg.Add(1)
		go func() {
			defer wg.Done()
			room.Seat(conn, "player")
			for r := range rounds {
				room.Vote(conn, deck[(i+r)%len(deck)])
				if r%5 == 0 {
					room.Reveal(conn)
				}
				if r%7 == 0 {
					room.NewRound(conn)
				}
				if r%11 == 0 {
					room.Rename(conn, "renamed")
				}
			}
		}()
	}

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()

	// The writers finish; the readers end when Close shuts the room down.
	time.Sleep(500 * time.Millisecond)
	manager.Close()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("goroutines did not finish; something is blocked")
	}

	// The room ends in a consistent state: every seat still belongs to exactly one
	// participant, and no goroutine is left running.
	if manager.Len() != 0 {
		t.Errorf("%d rooms remain after shutdown, want 0", manager.Len())
	}
}

func TestEnsureRoomCreatesARoomAtTheGivenIdentifier(t *testing.T) {
	manager, clock := newTestManager(t)
	defer manager.Close()

	// A room that existed, was used, and then expired.
	original, _ := manager.Create()
	id := original.ID()
	conn := attach(t, original, "token-a")
	original.Seat(conn, "Thomas")
	nextView(t, conn)
	original.Vote(conn, game.CardM)
	nextView(t, conn)

	original.Detach(conn)
	nextViewOrClosed(t, conn)
	settle(t, original)
	clock.Advance(testGrace + time.Second)
	manager.Sweep()
	if _, ok := manager.Lookup(id); ok {
		t.Fatal("the room was not discarded")
	}

	reached, err := manager.EnsureRoom(id)
	if err != nil {
		t.Fatalf("EnsureRoom: %v", err)
	}
	if reached.ID() != id {
		t.Errorf("reopened room has identifier %q, want the one asked for %q", reached.ID(), id)
	}
	if found, ok := manager.Lookup(id); !ok || found != reached {
		t.Error("the reopened room is not the one the manager now has at that identifier")
	}

	// What comes back is the address, not the game that was at it.
	_, view := attachView(t, reached, "token-a")
	if len(view.Participants) != 0 {
		t.Errorf("the reopened room has %d participants, want 0", len(view.Participants))
	}
	if view.Revealed {
		t.Error("the reopened room starts revealed")
	}
}

func TestEnsureRoomReturnsARoomThatIsStillAlive(t *testing.T) {
	// Reopening something that never went away must be a join, not an error and not a
	// replacement — otherwise the second person to press the button would evict the
	// first.
	manager, _ := newTestManager(t)
	defer manager.Close()

	original, _ := manager.Create()
	id := original.ID()
	conn := attach(t, original, "token-a")
	original.Seat(conn, "Thomas")
	nextView(t, conn)

	reached, err := manager.EnsureRoom(id)
	if err != nil {
		t.Fatalf("EnsureRoom: %v", err)
	}
	if reached != original {
		t.Fatal("asking for a live room produced a different one, evicting whoever was in it")
	}

	_, view := attachView(t, reached, "token-b")
	if len(view.Participants) != 1 || view.Participants[0].Name != "Thomas" {
		t.Errorf("participants = %+v, want the one who was already seated", view.Participants)
	}
}

func TestEnsureRoomRefusesAnIdentifierThatMayNotNameARoom(t *testing.T) {
	manager, _ := newTestManager(t)
	defer manager.Close()

	// A name somebody typed is as good as one the generator issued, so the refusals
	// here are only about the rules a name must satisfy at all.
	for name, id := range map[string]game.RoomID{
		"empty":           "",
		"one character":   "a",
		"four characters": "abcd",
		"a space":         "team alpha",
		"a dot":           "team.alpha",
		"a slash":         "team/alpha",
		"an umlaut":       "gruppe-fünf",
		"far too long":    game.RoomID(strings.Repeat("a", 65)),
	} {
		before := manager.Len()
		room, err := manager.EnsureRoom(id)
		if !errors.Is(err, ErrInvalidRoomID) {
			t.Errorf("%s: error = %v, want ErrInvalidRoomID", name, err)
		}
		if room != nil {
			t.Errorf("%s: a room was returned anyway", name)
		}
		if manager.Len() != before {
			t.Errorf("%s: the number of rooms changed, so something was created", name)
		}
	}
}

func TestEnsureRoomGivesEverybodyTheSameRoom(t *testing.T) {
	// The case this feature exists for: several people holding the same link, all
	// pressing reopen at the same moment. They must converge on one room, not race
	// each other into several with one silently replacing the rest.
	manager, _ := newTestManager(t)
	defer manager.Close()

	seed, _ := manager.Create()
	id := seed.ID()
	manager.Sweep()
	seed.Stop()
	manager.Close()

	fresh := NewManager(newFakeClock(), rand.Reader, testGrace, testLimits)
	defer fresh.Close()

	const askers = 24
	var wg sync.WaitGroup
	rooms := make([]*Room, askers)
	errs := make([]error, askers)

	start := make(chan struct{})
	for i := range askers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			rooms[i], errs[i] = fresh.EnsureRoom(id)
		}()
	}
	close(start)
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("asker %d: %v", i, err)
		}
	}
	for i, room := range rooms {
		if room != rooms[0] {
			t.Errorf("asker %d got a different room; they must all converge on one", i)
		}
	}
	if fresh.Len() != 1 {
		t.Errorf("%d rooms exist at that identifier, want 1", fresh.Len())
	}
}

func TestAReopenedRoomExpiresLikeAnyOther(t *testing.T) {
	manager, clock := newTestManager(t)
	defer manager.Close()

	seed, _ := manager.Create()
	id := seed.ID()
	clock.Advance(testGrace + time.Second)
	manager.Sweep()

	if _, err := manager.EnsureRoom(id); err != nil {
		t.Fatalf("EnsureRoom: %v", err)
	}
	if _, ok := manager.Lookup(id); !ok {
		t.Fatal("the reopened room is not there")
	}

	// Its grace period starts fresh rather than inheriting the old one.
	clock.Advance(testGrace / 2)
	manager.Sweep()
	if _, ok := manager.Lookup(id); !ok {
		t.Error("the reopened room was discarded before its own grace period elapsed")
	}

	clock.Advance(testGrace)
	manager.Sweep()
	if _, ok := manager.Lookup(id); ok {
		t.Error("the reopened room outlived its grace period")
	}
}

func TestALongSessionWithoutAnyGameActionNeverExpires(t *testing.T) {
	// The owner's question, answered as a test: people talk for an hour and nobody
	// clicks anything. The room must survive and, just as importantly, the person must
	// still be shown as present rather than quietly given up on.
	//
	// TestAnOccupiedRoomIsNeverDiscarded covers the room. This adds the participant,
	// and it does deliberately nothing at all after sitting down: no vote, no reveal,
	// no rename, not one message. Time is driven rather than waited for, so an hour
	// costs microseconds.
	manager, clock := newTestManager(t)
	defer manager.Close()

	room, err := manager.Create()
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	id := room.ID()

	conn := attach(t, room, "token-a")
	room.Seat(conn, "Thomas")
	nextView(t, conn)
	drain(conn)

	// Twelve times the grace period, swept at every step, with not one intent sent.
	for range 12 {
		clock.Advance(testGrace)
		manager.Sweep()
	}

	if _, ok := manager.Lookup(id); !ok {
		t.Fatal("a room whose participant simply said nothing was discarded; nothing may " +
			"measure how long it has been since somebody last did something")
	}

	// And the seat is still occupied by somebody present. Asking the room is what
	// produces a fresh view.
	second, view := attachView(t, room, "token-b")
	if len(view.Participants) != 1 {
		t.Fatalf("the table holds %d participants after a long silence, want 1", len(view.Participants))
	}
	if view.Participants[0].Away {
		t.Error("a participant who said nothing for a long time was marked away")
	}
	if view.Participants[0].Name != "Thomas" {
		t.Errorf("the participant is shown as %q", view.Participants[0].Name)
	}

	drain(conn)
	drain(second)
}
