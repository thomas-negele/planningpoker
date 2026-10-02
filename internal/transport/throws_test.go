package transport

import (
	"encoding/json"
	"strings"
	"testing"

	"de.thomasnegele.planningpoker/internal/hub"
)

// sameThrow compares two deliveries of one event. AgeMS is left out because each
// connection's writer stamps it at its own moment, which may differ by a millisecond.
func sameThrow(a, b thrownMessage) bool {
	a.AgeMS, b.AgeMS = 0, 0
	return a == b
}

func TestThrowEventsAreTransientSharedAndSnapshotFirst(t *testing.T) {
	srv := newTestServer(t)
	roomID := srv.createGame(t)

	actor := srv.dial(t, roomID)
	initial := actor.state(t)
	if initial.ThrowPolicy == nil ||
		initial.ThrowPolicy.ParticipantPerSecond != hub.ThrowsPerParticipant ||
		initial.ThrowPolicy.RoomPerSecond != hub.ThrowsPerRoom ||
		initial.ThrowPolicy.MessagePerSecond != testRate.PerSecond ||
		initial.ThrowPolicy.MessageBurst != testRate.Burst ||
		strings.Join(initial.ThrowPolicy.Objects, ",") != "paper-ball,paper-plane,flowers,heart" {
		t.Fatalf("initial throw policy = %+v", initial.ThrowPolicy)
	}

	target := srv.dial(t, roomID)
	actor.state(t) // attaching the target broadcasts a fresh state
	target.state(t)

	actor.send(t, clientMessage{Type: intentSeat, Name: "Ada"})
	actorState := actor.state(t)
	target.state(t)
	actorID := actorState.You

	target.send(t, clientMessage{Type: intentSeat, Name: "Grace"})
	actor.state(t)
	targetState := target.state(t)
	targetID := targetState.You

	// Give the hidden round a private card value, then attach a second tab to the
	// target's existing seat. The tab deliberately does not read immediately.
	actor.send(t, clientMessage{Type: intentVote, Card: "M"})
	actor.state(t)
	target.state(t)
	targetTab := srv.dial(t, roomID, target.cookies...)
	actor.state(t)
	target.state(t)

	actor.send(t, clientMessage{Type: intentThrow, Target: targetID, Object: string(hub.ThrowPaperPlane)})
	fromActor := actor.thrown(t)
	fromTarget := target.thrown(t)
	initialTargetTab := targetTab.state(t)
	if initialTargetTab.You != targetID || initialTargetTab.Room.Results != nil {
		t.Fatalf("same-seat tab initial snapshot = %+v", initialTargetTab)
	}
	rawTargetTab := targetTab.raw(t)
	var fromTargetTab thrownMessage
	if err := json.Unmarshal(rawTargetTab, &fromTargetTab); err != nil {
		t.Fatalf("decoding target-tab throw %s: %v", rawTargetTab, err)
	}
	if !sameThrow(fromActor, fromTarget) {
		t.Fatalf("shared throw differs: %+v and %+v", fromActor, fromTarget)
	}
	if !sameThrow(fromTargetTab, fromActor) {
		t.Fatalf("same-seat tab throw differs: %+v and %+v", fromTargetTab, fromActor)
	}
	if fromActor.Sender != actorID || fromActor.Target != targetID ||
		fromActor.Object != string(hub.ThrowPaperPlane) || fromActor.ID == "" {
		t.Errorf("throw = %+v", fromActor)
	}
	if fromActor.AgeMS < 0 || fromActor.AgeMS >= 1200 {
		t.Errorf("throw age = %dms, want a fresh event", fromActor.AgeMS)
	}
	for _, private := range []string{actor.cookies[0].Value, target.cookies[0].Value, `"card":"M"`} {
		if strings.Contains(string(rawTargetTab), private) {
			t.Errorf("throw payload exposed private/hidden value %q: %s", private, rawTargetTab)
		}
	}

	late := srv.dial(t, roomID)
	// Existing clients see the attachment snapshot; the late client sees a snapshot
	// first and receives no replay of the throw sent above.
	actor.state(t)
	target.state(t)
	targetTab.state(t)
	if got := late.state(t); got.Type != messageState || len(got.Room.Participants) != 2 {
		t.Errorf("late initial message = %+v", got)
	}
}

func TestInvalidThrowRequestsReceiveSpecificPrivateRefusals(t *testing.T) {
	srv := newTestServer(t)
	roomID := srv.createGame(t)

	actor := srv.dial(t, roomID)
	actor.state(t)
	target := srv.dial(t, roomID)
	actor.state(t)
	target.state(t)

	actorState := actor.seat(t, "Ada")
	target.state(t)
	targetState := target.seat(t, "Grace")
	actor.state(t)

	tests := []struct {
		target string
		object string
		code   string
	}{
		{targetState.You, "client-html", codeUnknownThrow},
		{actorState.You, string(hub.ThrowFlowers), codeThrowAtSelf},
		{"participant-from-another-room", string(hub.ThrowPaperBall), codeThrowTarget},
	}
	for _, test := range tests {
		actor.send(t, clientMessage{Type: intentThrow, Target: test.target, Object: test.object})
		if got := actor.refusal(t).Code; got != test.code {
			t.Errorf("throw (%q, %q) refusal = %q, want %q", test.target, test.object, got, test.code)
		}
	}
}

func TestThePolicyListsThePileOfPooOnlyWhenSwitchedOn(t *testing.T) {
	limits := testLimits
	limits.PooThrows = true
	srv := newTestServerWith(t, limits, testRate)
	roomID := srv.createGame(t)

	actor := srv.dial(t, roomID)
	initial := actor.state(t)
	if initial.ThrowPolicy == nil ||
		strings.Join(initial.ThrowPolicy.Objects, ",") != "paper-ball,paper-plane,flowers,heart,poo" {
		t.Fatalf("initial throw policy = %+v", initial.ThrowPolicy)
	}
}

func TestNewObjectsReachEveryConnectionUnchanged(t *testing.T) {
	limits := testLimits
	limits.PooThrows = true
	srv := newTestServerWith(t, limits, testRate)
	roomID := srv.createGame(t)

	actor := srv.dial(t, roomID)
	actor.state(t)
	target := srv.dial(t, roomID)
	actor.state(t)
	target.state(t)
	observer := srv.dial(t, roomID)
	actor.state(t)
	target.state(t)
	observer.state(t)

	actorState := actor.seat(t, "Ada")
	target.state(t)
	observer.state(t)
	targetState := target.seat(t, "Grace")
	actor.state(t)
	observer.state(t)

	for _, object := range []hub.ThrowObject{hub.ThrowHeart, hub.ThrowPoo} {
		actor.send(t, clientMessage{Type: intentThrow, Target: targetState.You, Object: string(object)})
		want := actor.thrown(t)
		if want.Object != string(object) || want.Sender != actorState.You || want.Target != targetState.You {
			t.Errorf("throw = %+v, want %q from %q at %q", want, object, actorState.You, targetState.You)
		}
		for name, client := range map[string]*client{"target": target, "observer": observer} {
			if got := client.thrown(t); !sameThrow(got, want) {
				t.Errorf("%s saw %+v, want %+v", name, got, want)
			}
		}
	}
}

func TestThePileOfPooIsRefusedWhileSwitchedOff(t *testing.T) {
	srv := newTestServer(t)
	roomID := srv.createGame(t)

	actor := srv.dial(t, roomID)
	actor.state(t)
	target := srv.dial(t, roomID)
	actor.state(t)
	target.state(t)

	actor.seat(t, "Ada")
	target.state(t)
	targetState := target.seat(t, "Grace")
	actor.state(t)

	actor.send(t, clientMessage{Type: intentThrow, Target: targetState.You, Object: string(hub.ThrowPoo)})
	if got := actor.refusal(t).Code; got != codeUnknownThrow {
		t.Errorf("refusal = %q, want %q", got, codeUnknownThrow)
	}

	// The heart is unaffected, and its event is the next thing the target receives:
	// the refused request produced nothing for anyone else.
	actor.send(t, clientMessage{Type: intentThrow, Target: targetState.You, Object: string(hub.ThrowHeart)})
	if got := target.thrown(t); got.Object != string(hub.ThrowHeart) {
		t.Errorf("target's next event = %+v, want the heart", got)
	}
}
