package hub

import (
	"errors"
	"testing"

	"de.thomasnegele.planningpoker/internal/game"
)

// seatedVisitor seats somebody who may not vote and returns their connection and
// public identifier.
func seatedVisitor(t *testing.T, room *Room, token, name string) (*Conn, game.ParticipantID) {
	t.Helper()
	conn := attach(t, room, token)
	if !room.Seat(conn, name, true) {
		t.Fatal("seat request was not queued")
	}
	view := nextView(t, conn)
	return conn, view.Participants[len(view.Participants)-1].ID
}

// visitorIn reports what a snapshot says about one seat's mode.
func visitorIn(t *testing.T, view game.View, id game.ParticipantID) bool {
	t.Helper()
	for _, p := range view.Participants {
		if p.ID == id {
			return p.Visitor
		}
	}
	t.Fatalf("participant %q is not in the snapshot", id)
	return false
}

func TestASeatKeepsItsVisitorModeAcrossConnections(t *testing.T) {
	manager, _ := newTestManager(t)
	defer manager.Close()
	room, _ := manager.Create()

	first, id := seatedVisitor(t, room, "token-a", "Thomas")

	// A second tab for the same seat must see the same mode, and reconnecting with
	// the same credential must not turn a visitor back into a voter.
	second, view := attachView(t, room, "token-a")
	if !visitorIn(t, view, id) {
		t.Error("a second connection to a visitor's seat did not see visitor mode")
	}

	drain(first)
	drain(second)
	if !room.Seat(second, "Thomas", false) {
		t.Fatal("Seat was refused")
	}
	if !visitorIn(t, nextView(t, second), id) {
		t.Error("returning to a seat with the visitor box unticked switched the mode off")
	}
}

func TestTheRoomRefusesAVisitorsVoteAndTellsOnlyThem(t *testing.T) {
	manager, _ := newTestManager(t)
	defer manager.Close()
	room, _ := manager.Create()

	visitor, id := seatedVisitor(t, room, "token-a", "Thomas")
	other, _ := seated(t, room, "token-b", "Bert")

	drain(visitor)
	drain(other)
	if !room.Vote(visitor, game.CardM) {
		t.Fatal("Vote was not queued")
	}

	if err := nextRefusal(t, visitor); !errors.Is(err, game.ErrVisitorCannotVote) {
		t.Errorf("refusal = %v, want ErrVisitorCannotVote", err)
	}
	assertNoUpdate(t, other)

	view := currentView(t, room, "token-c")
	for _, p := range view.Participants {
		if p.ID == id && p.Voted {
			t.Error("the refused vote was recorded anyway")
		}
	}
}

func TestAVisitorMayRunTheRoundAndChangeTheDeck(t *testing.T) {
	manager, _ := newTestManager(t)
	defer manager.Close()
	room, _ := manager.Create()

	visitor, _ := seatedVisitor(t, room, "token-a", "Thomas")

	drain(visitor)
	if !room.SetDeck(visitor, game.FibonacciDeckName) {
		t.Fatal("SetDeck was not queued")
	}
	if got := nextView(t, visitor).Deck.Name; got != game.FibonacciDeckName {
		t.Errorf("deck = %q after a visitor changed it, want %q", got, game.FibonacciDeckName)
	}

	if !room.Reveal(visitor) {
		t.Fatal("Reveal was not queued")
	}
	if !nextView(t, visitor).Revealed {
		t.Error("a round a visitor revealed is still hidden")
	}

	if !room.NewRound(visitor) {
		t.Fatal("NewRound was not queued")
	}
	if nextView(t, visitor).Revealed {
		t.Error("a round a visitor started fresh is still revealed")
	}
}

func TestAVisitorMayRenameThemselvesAndThrow(t *testing.T) {
	manager, _ := newTestManager(t)
	defer manager.Close()
	room, _ := manager.Create()

	visitor, id := seatedVisitor(t, room, "token-a", "Thomas")
	target, targetID := seated(t, room, "token-b", "Bert")

	drain(visitor)
	if !room.Rename(visitor, "Thomas N.", nil) {
		t.Fatal("Rename was not queued")
	}
	view := nextView(t, visitor)
	if !visitorIn(t, view, id) {
		t.Error("renaming without naming a mode switched visitor mode off")
	}

	drain(visitor)
	drain(target)
	if !room.Throw(visitor, targetID, ThrowPaperBall) {
		t.Fatal("Throw was not accepted")
	}
	if event := nextThrow(t, target); event.Sender != id {
		t.Errorf("throw sender = %q, want the visitor %q", event.Sender, id)
	}
}

func TestSavingVisitorModeDeletesTheHiddenVoteForEveryone(t *testing.T) {
	manager, _ := newTestManager(t)
	defer manager.Close()
	room, _ := manager.Create()

	voter, id := seated(t, room, "token-a", "Thomas")
	watcher, _ := seated(t, room, "token-b", "Bert")

	drain(voter)
	drain(watcher)
	if !room.Vote(voter, game.CardM) {
		t.Fatal("Vote was not queued")
	}
	nextView(t, voter)
	nextView(t, watcher)

	switchOn := true
	if !room.Rename(voter, "Thomas", &switchOn) {
		t.Fatal("Rename was not queued")
	}

	// Everyone's snapshot, not only the voter's own, must show the vote gone.
	for name, conn := range map[string]*Conn{"the voter": voter, "the other participant": watcher} {
		view := nextView(t, conn)
		for _, p := range view.Participants {
			if p.ID != id {
				continue
			}
			if p.Voted {
				t.Errorf("%s still sees a vote for the new visitor", name)
			}
			if !p.Visitor {
				t.Errorf("%s does not see the new visitor mode", name)
			}
		}
	}
}

// currentView attaches a fresh connection purely to read the room's current state.
func currentView(t *testing.T, room *Room, token string) game.View {
	t.Helper()
	_, view := attachView(t, room, token)
	return view
}
