package game

import (
	"crypto/rand"
	"errors"
	"testing"
)

// joinVisitor seats somebody who may not vote, failing the test if it does not work.
func joinVisitor(t *testing.T, room *Room, name string) ParticipantID {
	t.Helper()
	id, err := room.Join(rand.Reader, name, true)
	if err != nil {
		t.Fatalf("Join(%q) as a visitor: %v", name, err)
	}
	return id
}

// visitorOf reads a seat's mode out of the view, which is the only place it is
// visible from outside the package.
func visitorOf(t *testing.T, room *Room, id ParticipantID) bool {
	t.Helper()
	for _, p := range room.View().Participants {
		if p.ID == id {
			return p.Visitor
		}
	}
	t.Fatalf("participant %q is not at the table", id)
	return false
}

// votedOf reports whether the view says this seat holds a card for the round.
func votedOf(t *testing.T, room *Room, id ParticipantID) bool {
	t.Helper()
	for _, p := range room.View().Participants {
		if p.ID == id {
			return p.Voted
		}
	}
	t.Fatalf("participant %q is not at the table", id)
	return false
}

// visitorMode is the address-of helper the optional rename argument needs; a nil
// argument means "leave the mode as it is", which is a distinct case and is tested
// separately.
func visitorMode(v bool) *bool { return &v }

func TestJoiningAsAVisitorTakesANonVotingSeat(t *testing.T) {
	room := newTestRoom(t)
	id := joinVisitor(t, room, "Thomas")

	if !visitorOf(t, room, id) {
		t.Error("somebody who joined as a visitor is not shown as one")
	}
	if err := room.Vote(id, CardM); !errors.Is(err, ErrVisitorCannotVote) {
		t.Errorf("Vote by a visitor: error = %v, want ErrVisitorCannotVote", err)
	}
	if votedOf(t, room, id) {
		t.Error("a refused visitor vote was recorded anyway")
	}
}

func TestJoiningWithoutAskingForVisitorModeTakesAVotingSeat(t *testing.T) {
	room := newTestRoom(t)
	id := join(t, room, "Thomas")

	if visitorOf(t, room, id) {
		t.Error("a seat taken without asking for visitor mode became a visitor")
	}
	if err := room.Vote(id, CardM); err != nil {
		t.Errorf("Vote by a voter: %v", err)
	}
}

func TestRejoiningKeepsTheSeatsVisitorMode(t *testing.T) {
	room := newTestRoom(t)

	visitor := joinVisitor(t, room, "Thomas")
	if err := room.Rejoin(visitor, "Thomas N."); err != nil {
		t.Fatalf("Rejoin: %v", err)
	}
	if !visitorOf(t, room, visitor) {
		t.Error("a visitor who returned to their seat lost visitor mode")
	}

	voter := join(t, room, "Bert")
	if err := room.Rejoin(voter, "Bert"); err != nil {
		t.Fatalf("Rejoin: %v", err)
	}
	if visitorOf(t, room, voter) {
		t.Error("a voter who returned to their seat became a visitor")
	}
}

func TestRenameSwitchesVisitorModeAndLeavingItNilPreservesIt(t *testing.T) {
	room := newTestRoom(t)
	id := join(t, room, "Thomas")

	if err := room.Rename(id, "Thomas", visitorMode(true)); err != nil {
		t.Fatalf("Rename into visitor mode: %v", err)
	}
	if !visitorOf(t, room, id) {
		t.Fatal("saving visitor mode did not change the seat")
	}

	// A client that does not know about visitor mode must not switch it off by
	// renaming.
	if err := room.Rename(id, "Thomas N.", nil); err != nil {
		t.Fatalf("Rename without a mode: %v", err)
	}
	if !visitorOf(t, room, id) {
		t.Error("a rename that named no mode switched visitor mode off")
	}
	if got := nameOf(t, room, id); got != "Thomas N." {
		t.Errorf("name = %q, want %q", got, "Thomas N.")
	}

	if err := room.Rename(id, "Thomas N.", visitorMode(false)); err != nil {
		t.Fatalf("Rename out of visitor mode: %v", err)
	}
	if visitorOf(t, room, id) {
		t.Error("saving with visitor mode unticked did not return the seat to voting")
	}
}

func TestARefusedNameChangesNeitherNameNorMode(t *testing.T) {
	room := newTestRoom(t)
	id := join(t, room, "Thomas")

	if err := room.Rename(id, "   ", visitorMode(true)); !errors.Is(err, ErrEmptyName) {
		t.Fatalf("Rename to an empty name: error = %v, want ErrEmptyName", err)
	}
	if visitorOf(t, room, id) {
		t.Error("a rename the rules refused still switched visitor mode on")
	}
	if got := nameOf(t, room, id); got != "Thomas" {
		t.Errorf("name = %q, want it unchanged as %q", got, "Thomas")
	}
}

func TestSwitchingIntoVisitorModeDeletesAHiddenVoteForGood(t *testing.T) {
	room := newTestRoom(t)
	id := join(t, room, "Thomas")

	if err := room.Vote(id, CardM); err != nil {
		t.Fatalf("Vote: %v", err)
	}
	if err := room.Rename(id, "Thomas", visitorMode(true)); err != nil {
		t.Fatalf("Rename into visitor mode: %v", err)
	}
	if votedOf(t, room, id) {
		t.Error("a hidden vote survived the switch into visitor mode")
	}

	// Switching back must not bring the old card back, but a new one may be played.
	if err := room.Rename(id, "Thomas", visitorMode(false)); err != nil {
		t.Fatalf("Rename out of visitor mode: %v", err)
	}
	if votedOf(t, room, id) {
		t.Error("switching back restored the deleted vote")
	}

	if err := room.Vote(id, CardL); err != nil {
		t.Fatalf("Vote after returning to voting: %v", err)
	}
	if err := room.Reveal(id); err != nil {
		t.Fatalf("Reveal: %v", err)
	}
	if got := cardOf(t, room, id); got != CardL {
		t.Errorf("revealed card = %q, want the newly played %q", got, CardL)
	}
}

func TestSwitchingModeAfterRevealLeavesTheResultAlone(t *testing.T) {
	room := newTestRoom(t)
	thomas := join(t, room, "Thomas")
	bert := join(t, room, "Bert")

	for _, id := range []ParticipantID{thomas, bert} {
		if err := room.Vote(id, CardM); err != nil {
			t.Fatalf("Vote: %v", err)
		}
	}
	if err := room.Reveal(thomas); err != nil {
		t.Fatalf("Reveal: %v", err)
	}

	if err := room.Rename(thomas, "Thomas", visitorMode(true)); err != nil {
		t.Fatalf("Rename into visitor mode after reveal: %v", err)
	}

	if got := cardOf(t, room, thomas); got != CardM {
		t.Errorf("revealed card = %q after switching to visitor mode, want %q", got, CardM)
	}

	tally := room.View().Results.Tally
	if len(tally) != 1 || tally[0].Card != CardM || tally[0].Count != 2 {
		t.Errorf("tally = %v after a mode change, want two votes for %q", tally, CardM)
	}
}

func TestANewRoundStartsEmptyForVisitorsAndVotersAlike(t *testing.T) {
	room := newTestRoom(t)
	voter := join(t, room, "Thomas")
	visitor := joinVisitor(t, room, "Bert")

	if err := room.Vote(voter, CardM); err != nil {
		t.Fatalf("Vote: %v", err)
	}
	if err := room.NewRound(visitor); err != nil {
		t.Fatalf("NewRound started by a visitor: %v", err)
	}

	if votedOf(t, room, voter) || votedOf(t, room, visitor) {
		t.Error("a new round did not start with an empty table of votes")
	}
	if !visitorOf(t, room, visitor) {
		t.Error("a new round changed somebody's visitor mode")
	}
}

func TestVisitorsDoNotHoldUpTheEveryoneHasVotedReport(t *testing.T) {
	room := newTestRoom(t)
	voter := join(t, room, "Thomas")
	joinVisitor(t, room, "Bert")

	if err := room.Vote(voter, CardM); err != nil {
		t.Fatalf("Vote: %v", err)
	}
	if !room.EveryonePresentHasVoted() {
		t.Error("a present visitor who cannot vote kept the round from being complete")
	}
}

func TestATableOfVisitorsIsNotAFinishedRound(t *testing.T) {
	room := newTestRoom(t)
	first := joinVisitor(t, room, "Thomas")
	joinVisitor(t, room, "Bert")

	if room.EveryonePresentHasVoted() {
		t.Error("a round nobody present may vote in reported that all votes were in")
	}

	// The report is a statement about the round, not a permission: revealing such a
	// round must still work.
	if err := room.Reveal(first); err != nil {
		t.Errorf("Reveal by a visitor: %v", err)
	}
}

func TestAVisitorKeepsEverySeatedActionApartFromVoting(t *testing.T) {
	room := newTestRoom(t)
	visitor := joinVisitor(t, room, "Thomas")

	if err := room.Rename(visitor, "Thomas N.", nil); err != nil {
		t.Errorf("Rename by a visitor: %v", err)
	}
	if err := room.SetDeck(visitor, FibonacciDeckName); err != nil {
		t.Errorf("SetDeck by a visitor: %v", err)
	}
	if err := room.Reveal(visitor); err != nil {
		t.Errorf("Reveal by a visitor: %v", err)
	}
	if err := room.NewRound(visitor); err != nil {
		t.Errorf("NewRound by a visitor: %v", err)
	}
}

func TestAVisitorIsRefusedBeforeTheDeckIsEvenConsulted(t *testing.T) {
	room := newTestRoom(t)
	visitor := joinVisitor(t, room, "Thomas")

	// The refusal must name visitor mode rather than the card or the round, so the
	// interface can say why.
	if err := room.Vote(visitor, "not-a-card"); !errors.Is(err, ErrVisitorCannotVote) {
		t.Errorf("Vote with an unknown card by a visitor: error = %v, want ErrVisitorCannotVote", err)
	}
	if err := room.Reveal(visitor); err != nil {
		t.Fatalf("Reveal: %v", err)
	}
	if err := room.Vote(visitor, CardM); !errors.Is(err, ErrVisitorCannotVote) {
		t.Errorf("Vote after reveal by a visitor: error = %v, want ErrVisitorCannotVote", err)
	}
}
