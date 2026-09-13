package hub

import (
	"crypto/rand"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"de.thomasnegele.planningpoker/internal/game"
)

type gatedReader struct {
	io.Reader
	mu      sync.Mutex
	entered chan struct{}
	release chan struct{}
}

func (r *gatedReader) blockNext() (<-chan struct{}, func()) {
	r.mu.Lock()
	defer r.mu.Unlock()
	entered := make(chan struct{})
	release := make(chan struct{})
	r.entered = entered
	r.release = release
	var once sync.Once
	return entered, func() { once.Do(func() { close(release) }) }
}

func (r *gatedReader) Read(p []byte) (int, error) {
	r.mu.Lock()
	entered, release := r.entered, r.release
	r.entered, r.release = nil, nil
	r.mu.Unlock()
	if entered != nil {
		close(entered)
		<-release
	}
	return r.Reader.Read(p)
}

func seated(t *testing.T, room *Room, token, name string) (*Conn, game.ParticipantID) {
	t.Helper()
	conn := attach(t, room, token)
	if !room.Seat(conn, name) {
		t.Fatal("seat request was not queued")
	}
	view := nextView(t, conn)
	return conn, view.Participants[len(view.Participants)-1].ID
}

func nextThrow(t *testing.T, conn *Conn) ThrowEvent {
	t.Helper()
	select {
	case event := <-conn.Throws():
		return event
	case <-time.After(2 * time.Second):
		t.Fatal("no throw event arrived")
		return ThrowEvent{}
	}
}

func assertNoThrow(t *testing.T, conn *Conn) {
	t.Helper()
	select {
	case event := <-conn.Throws():
		t.Errorf("unexpected throw event: %+v", event)
	case <-time.After(20 * time.Millisecond):
	}
}

func TestAThrowUsesTheSeatedSenderAndLeavesGameStateAlone(t *testing.T) {
	manager, _ := newTestManager(t)
	defer manager.Close()
	room, _ := manager.Create()

	actor, actorID := seated(t, room, "token-a", "Ada")
	target, targetID := seated(t, room, "token-b", "Grace")
	drain(actor)
	if !room.Throw(actor, targetID, ThrowPaperPlane) {
		t.Fatal("throw was not admitted")
	}
	fromActor := nextThrow(t, actor)
	fromTarget := nextThrow(t, target)
	if fromActor != fromTarget {
		t.Fatalf("clients received different events: %+v and %+v", fromActor, fromTarget)
	}
	if fromActor.Sender != actorID || fromActor.Target != targetID || fromActor.Object != ThrowPaperPlane {
		t.Errorf("event = %+v, want server-derived sender %q and target %q", fromActor, actorID, targetID)
	}
	assertNoUpdate(t, actor)
}

func TestThrowFanoutIncludesEveryCurrentTabOnlyInItsRoom(t *testing.T) {
	manager, _ := newTestManager(t)
	defer manager.Close()
	room, _ := manager.Create()
	otherRoom, _ := manager.Create()

	actor, _ := seated(t, room, "actor-token", "Ada")
	target, targetID := seated(t, room, "target-token", "Grace")
	actorTab := attach(t, room, "actor-token")
	observer, _ := seated(t, room, "observer-token", "Lin")
	outsider, _ := seated(t, otherRoom, "outsider-token", "Edsger")
	inRoom := []*Conn{actor, target, actorTab, observer}
	for _, conn := range inRoom {
		drain(conn)
	}
	drain(outsider)

	room.Throw(actorTab, targetID, ThrowFlowers)
	want := nextThrow(t, actor)
	for _, conn := range []*Conn{target, actorTab, observer} {
		if got := nextThrow(t, conn); got != want {
			t.Errorf("fanout event = %+v, want %+v", got, want)
		}
	}
	assertNoThrow(t, outsider)
}

func TestAllThrowObjectsWorkBeforeVotingAfterVotingAndAfterReveal(t *testing.T) {
	manager, clock := newTestManager(t)
	defer manager.Close()
	room, _ := manager.Create()

	actor, _ := seated(t, room, "token-a", "Ada")
	target, targetID := seated(t, room, "token-b", "Grace")
	drain(actor)
	drain(target)

	objects := []ThrowObject{ThrowPaperBall, ThrowPaperPlane, ThrowFlowers}
	assertStage := func(stage string) {
		t.Helper()
		for _, object := range objects {
			if !room.Throw(actor, targetID, object) {
				t.Fatalf("%s %q throw was not admitted", stage, object)
			}
			if got := nextThrow(t, actor).Object; got != object {
				t.Errorf("%s actor saw %q, want %q", stage, got, object)
			}
			if got := nextThrow(t, target).Object; got != object {
				t.Errorf("%s target saw %q, want %q", stage, got, object)
			}
			assertNoUpdate(t, actor)
			assertNoUpdate(t, target)
		}
		clock.Advance(time.Second)
	}

	assertStage("before voting")
	room.Vote(target, game.CardM)
	nextView(t, actor)
	nextView(t, target)
	assertStage("after voting")
	room.Reveal(actor)
	if view := nextView(t, actor); !view.Revealed {
		t.Fatal("reveal did not take effect")
	}
	nextView(t, target)
	assertStage("after reveal")
}

func TestInvalidThrowsAreRefusedPrivately(t *testing.T) {
	manager, _ := newTestManager(t)
	defer manager.Close()
	room, _ := manager.Create()

	unseated := attach(t, room, "unseated")
	actor, actorID := seated(t, room, "token-a", "Ada")
	target, targetID := seated(t, room, "token-b", "Grace")
	otherRoom, _ := manager.Create()
	other, otherID := seated(t, otherRoom, "token-other", "Lin")
	drain(unseated)
	drain(actor)
	drain(target)
	drain(other)

	tests := []struct {
		name   string
		conn   *Conn
		target game.ParticipantID
		object ThrowObject
		want   error
	}{
		{"unseated sender", unseated, targetID, ThrowPaperBall, game.ErrUnknownParticipant},
		{"unknown object", actor, targetID, "html-from-client", ErrUnknownThrowObject},
		{"self target", actor, actorID, ThrowFlowers, ErrThrowAtSelf},
		{"unknown target", actor, "elsewhere", ThrowFlowers, ErrThrowTargetAbsent},
		{"cross-room target", actor, otherID, ThrowFlowers, ErrThrowTargetAbsent},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !room.Throw(test.conn, test.target, test.object) {
				t.Fatal("request was not admitted for validation")
			}
			if err := nextRefusal(t, test.conn); !errors.Is(err, test.want) {
				t.Errorf("refusal = %v, want %v", err, test.want)
			}
			assertNoThrow(t, actor)
			assertNoThrow(t, target)
		})
	}

	room.Detach(target)
	settle(t, room)
	drain(actor)
	if !room.Throw(actor, targetID, ThrowPaperBall) {
		t.Fatal("away-target request was not admitted for validation")
	}
	if err := nextRefusal(t, actor); !errors.Is(err, ErrThrowTargetAbsent) {
		t.Errorf("away-target refusal = %v, want ErrThrowTargetAbsent", err)
	}
}

func TestParticipantThrowLimitSurvivesReconnectToTheSameSeat(t *testing.T) {
	manager, clock := newTestManager(t)
	defer manager.Close()
	room, _ := manager.Create()

	actor, actorID := seated(t, room, "shared-token", "Ada")
	target, targetID := seated(t, room, "target-token", "Grace")
	drain(actor)
	drain(target)
	for range ThrowsPerParticipant {
		room.Throw(actor, targetID, ThrowPaperBall)
		nextThrow(t, actor)
		nextThrow(t, target)
	}

	room.Detach(actor)
	settle(t, room)
	drain(target)
	reconnected := NewConn("shared-token")
	if room.Attach(reconnected) != Attached {
		t.Fatal("reconnect was refused")
	}
	if update := nextUpdate(t, reconnected); update.You != actorID {
		t.Fatalf("reconnected identity = %q, want %q", update.You, actorID)
	}
	drain(target)
	room.Throw(reconnected, targetID, ThrowFlowers)
	settle(t, room)
	assertNoThrow(t, target)

	clock.Advance(time.Second)
	room.Throw(reconnected, targetID, ThrowFlowers)
	nextThrow(t, target)
}

func TestParticipantThrowLimitSpansTabsAndUsesARollingWindow(t *testing.T) {
	manager, clock := newTestManager(t)
	defer manager.Close()
	room, _ := manager.Create()

	first, _ := seated(t, room, "shared-token", "Ada")
	target, targetID := seated(t, room, "target-token", "Grace")
	second := attach(t, room, "shared-token")
	drain(first)
	drain(target)

	for i, conn := range []*Conn{first, second, first} {
		if !room.Throw(conn, targetID, ThrowPaperBall) {
			t.Fatalf("throw %d was not admitted", i+1)
		}
		nextThrow(t, first)
		nextThrow(t, second)
		nextThrow(t, target)
	}
	if !room.Throw(second, targetID, ThrowFlowers) {
		t.Fatal("limited throw was not admitted for the room to discard")
	}
	settle(t, room)
	assertNoThrow(t, first)
	assertNoThrow(t, second)
	assertNoThrow(t, target)

	clock.Advance(time.Second - time.Nanosecond)
	room.Throw(first, targetID, ThrowFlowers)
	settle(t, room)
	assertNoThrow(t, target)

	clock.Advance(time.Nanosecond)
	room.Throw(first, targetID, ThrowFlowers)
	nextThrow(t, target)
}

func TestRoomThrowLimitCombinesParticipants(t *testing.T) {
	manager, _ := newTestManager(t)
	defer manager.Close()
	room, _ := manager.Create()

	target, targetID := seated(t, room, "target", "Target")
	actors := make([]*Conn, 0, 5)
	for i := 0; i < 5; i++ {
		actor, _ := seated(t, room, string(rune('a'+i)), "Actor")
		actors = append(actors, actor)
	}
	all := append([]*Conn{target}, actors...)
	for _, conn := range all {
		drain(conn)
	}

	for i := 0; i < ThrowsPerRoom; i++ {
		actor := actors[i%len(actors)]
		room.Throw(actor, targetID, ThrowPaperBall)
		for _, conn := range all {
			nextThrow(t, conn)
		}
	}
	room.Throw(actors[4], targetID, ThrowFlowers)
	settle(t, room)
	for _, conn := range all {
		assertNoThrow(t, conn)
	}

	// The room-level rejection above must not spend this participant's own
	// allowance. Once the room window expires, all three throws fit.
	managerClock := manager.clock.(*fakeClock)
	managerClock.Advance(time.Second)
	for range ThrowsPerParticipant {
		room.Throw(actors[4], targetID, ThrowFlowers)
		nextThrow(t, target)
	}
}

func TestCosmeticBacklogDoesNotCloseAConnectionOrConsumeReliableQueue(t *testing.T) {
	conn := NewConn("token")
	for i := 0; i < 10; i++ {
		conn.tryThrow(ThrowEvent{ID: string(rune('a' + i))})
	}

	select {
	case <-conn.Closed():
		t.Fatal("cosmetic overflow closed the connection")
	default:
	}
	if len(conn.throws) != throwDeliverySize {
		t.Errorf("pending throws = %d, want bounded size %d", len(conn.throws), throwDeliverySize)
	}
	if len(conn.updates) != 0 {
		t.Errorf("cosmetic traffic occupied %d reliable slots", len(conn.updates))
	}
}

func TestSaturatedThrowAdmissionStillProcessesGameCommands(t *testing.T) {
	clock := newFakeClock()
	random := &gatedReader{Reader: rand.Reader}
	manager := NewManager(clock, random, testGrace, testLimits)
	defer manager.Close()
	room, _ := manager.Create()

	actor, _ := seated(t, room, "actor", "Ada")
	target, targetID := seated(t, room, "target", "Grace")
	drain(actor)
	drain(target)

	entered, release := random.blockNext()
	t.Cleanup(release)
	if !room.Throw(actor, targetID, ThrowPaperBall) {
		t.Fatal("first throw was not admitted")
	}
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("room did not enter the controlled random read")
	}
	if !room.Throw(actor, targetID, ThrowPaperPlane) {
		t.Fatal("one pending cosmetic request should fit")
	}
	for i := 0; i < 20; i++ {
		if room.Throw(actor, targetID, ThrowFlowers) {
			t.Fatal("cosmetic inbox grew beyond its fixed slot")
		}
	}
	room.Vote(target, game.CardM)
	room.Reveal(actor)
	room.NewRound(actor)
	release()

	for i := 0; i < 3; i++ {
		nextView(t, actor)
	}
	if final := roomViewAfterBarrier(t, room, actor); final.Revealed || final.Participants[1].Voted {
		t.Fatalf("game commands did not complete through a cosmetic backlog: %+v", final)
	}
	select {
	case <-actor.Closed():
		t.Fatal("cosmetic saturation closed the healthy connection")
	default:
	}
}

func TestQueuedThrowRevalidatesTargetPresence(t *testing.T) {
	clock := newFakeClock()
	random := &gatedReader{Reader: rand.Reader}
	manager := NewManager(clock, random, testGrace, testLimits)
	defer manager.Close()
	room, _ := manager.Create()

	actor, _ := seated(t, room, "actor", "Ada")
	target, targetID := seated(t, room, "target", "Grace")
	drain(actor)
	drain(target)
	entered, release := random.blockNext()
	t.Cleanup(release)
	room.Throw(actor, targetID, ThrowPaperBall)
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("room did not enter the controlled random read")
	}
	if !room.Throw(actor, targetID, ThrowFlowers) {
		t.Fatal("queued throw was not admitted")
	}
	room.Detach(target)
	release()

	nextThrow(t, actor)
	nextView(t, actor)
	if err := nextRefusal(t, actor); !errors.Is(err, ErrThrowTargetAbsent) {
		t.Fatalf("queued throw refusal = %v, want ErrThrowTargetAbsent", err)
	}
	assertNoThrow(t, actor)
}

func roomViewAfterBarrier(t *testing.T, room *Room, conn *Conn) game.View {
	t.Helper()
	_, view := attachView(t, room, "probe")
	drain(conn)
	return view
}
