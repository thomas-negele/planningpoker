package game

import (
	"errors"
	"testing"
)

// cardOf reveals the round and reads one participant's card out of the results.
// Reading a card any other way is impossible from outside this package, which is
// the point.
func cardOf(t *testing.T, room *Room, id ParticipantID) Card {
	t.Helper()
	view := room.View()
	if view.Results == nil {
		t.Fatal("the round is not revealed, so no card can be read")
	}
	for _, c := range view.Results.Cards {
		if c.ID == id {
			return c.Card
		}
	}
	t.Fatalf("participant %q is not in the results", id)
	return NoCard
}

func TestCastingAVoteRecordsIt(t *testing.T) {
	room := newTestRoom(t)
	id := join(t, room, "Thomas")

	if err := room.Vote(id, CardM); err != nil {
		t.Fatalf("Vote: %v", err)
	}
	if !room.View().Participants[0].Voted {
		t.Error("the participant is not shown as having voted")
	}

	if err := room.Reveal(id); err != nil {
		t.Fatalf("Reveal: %v", err)
	}
	if got := cardOf(t, room, id); got != CardM {
		t.Errorf("card = %q, want %q", got, CardM)
	}
}

func TestChangingAVoteReplacesItAndLeavesNoTrace(t *testing.T) {
	room := newTestRoom(t)
	id := join(t, room, "Thomas")

	for _, card := range []Card{CardM, CardL, CardS} {
		if err := room.Vote(id, card); err != nil {
			t.Fatalf("Vote(%q): %v", card, err)
		}
	}

	// Exactly one vote is held, and it is the last one.
	if got := len(room.round.votes); got != 1 {
		t.Errorf("the room holds %d votes for one participant, want 1", got)
	}
	if err := room.Reveal(id); err != nil {
		t.Fatalf("Reveal: %v", err)
	}
	if got := cardOf(t, room, id); got != CardS {
		t.Errorf("card = %q, want the last one played, %q", got, CardS)
	}

	// The discarded choices must be nowhere in the room's state. Keeping them would
	// leak a participant's hesitation, and nothing in this product has a use for
	// them.
	tally := room.View().Results.Tally
	for _, entry := range tally {
		if entry.Card == CardM || entry.Card == CardL {
			t.Errorf("a discarded earlier choice %q is still recorded in the tally", entry.Card)
		}
	}
	if len(tally) != 1 {
		t.Errorf("tally has %d entries, want 1 — only the final vote exists", len(tally))
	}
}

func TestCardOutsideTheDeckIsRefused(t *testing.T) {
	room := newTestRoom(t)
	id := join(t, room, "Thomas")

	for _, card := range []Card{"XXL", "13", "m", "", "☕☕"} {
		if err := room.Vote(id, card); !errors.Is(err, ErrCardNotInDeck) {
			t.Errorf("Vote(%q) error = %v, want ErrCardNotInDeck", card, err)
		}
	}

	if room.View().Participants[0].Voted {
		t.Error("a refused card was recorded as a vote")
	}
}

func TestNotVotedIsDistinctFromVotingUnknown(t *testing.T) {
	room := newTestRoom(t)
	silent := join(t, room, "Silent")
	unsure := join(t, room, "Unsure")

	if err := room.Vote(unsure, CardUnknown); err != nil {
		t.Fatalf("Vote: %v", err)
	}

	view := room.View()
	if view.Participants[0].Voted {
		t.Error("the participant who played nothing is shown as having voted")
	}
	if !view.Participants[1].Voted {
		t.Error("the participant who played \"?\" is not shown as having voted")
	}

	if err := room.Reveal(unsure); err != nil {
		t.Fatalf("Reveal: %v", err)
	}
	if got := cardOf(t, room, silent); got != NoCard {
		t.Errorf("the silent participant holds %q, want no card", got)
	}
	if got := cardOf(t, room, unsure); got != CardUnknown {
		t.Errorf("the unsure participant holds %q, want %q", got, CardUnknown)
	}
}

func TestAnyParticipantMayReveal(t *testing.T) {
	room := newTestRoom(t)
	creator := join(t, room, "Thomas")
	other := join(t, room, "Anna")
	if err := room.Vote(creator, CardM); err != nil {
		t.Fatalf("Vote: %v", err)
	}

	// There is no host and no creator role: whoever is at the table may reveal.
	if err := room.Reveal(other); err != nil {
		t.Fatalf("Reveal by a participant who did not create the room: %v", err)
	}
	if !room.Revealed() {
		t.Error("the round is not revealed")
	}
}

func TestRevealingBeforeEveryoneHasVotedIsAllowed(t *testing.T) {
	room := newTestRoom(t)
	voter := join(t, room, "Thomas")
	silent := join(t, room, "Anna")
	if err := room.Vote(voter, CardL); err != nil {
		t.Fatalf("Vote: %v", err)
	}

	// One person who is absent, distracted or on a dead network can never block a
	// meeting.
	if err := room.Reveal(voter); err != nil {
		t.Fatalf("Reveal: %v", err)
	}

	if got := cardOf(t, room, voter); got != CardL {
		t.Errorf("the voter holds %q, want %q", got, CardL)
	}
	if got := cardOf(t, room, silent); got != NoCard {
		t.Errorf("the participant who did not vote holds %q, want no card", got)
	}
}

func TestRevealingTwiceIsHarmless(t *testing.T) {
	room := newTestRoom(t)
	first := join(t, room, "Thomas")
	second := join(t, room, "Anna")
	if err := room.Vote(first, CardM); err != nil {
		t.Fatalf("Vote: %v", err)
	}
	if err := room.Vote(second, CardL); err != nil {
		t.Fatalf("Vote: %v", err)
	}

	if err := room.Reveal(first); err != nil {
		t.Fatalf("first Reveal: %v", err)
	}
	before := room.View()

	// Two people pressing the button at the same moment must not be a fault.
	if err := room.Reveal(second); err != nil {
		t.Fatalf("second Reveal: %v", err)
	}
	after := room.View()

	if !after.Revealed {
		t.Error("the round is no longer revealed after being revealed twice")
	}
	for i := range after.Results.Cards {
		if after.Results.Cards[i] != before.Results.Cards[i] {
			t.Errorf("card %d changed on the second reveal: %v, want %v",
				i, after.Results.Cards[i], before.Results.Cards[i])
		}
	}
}

func TestVotingAfterTheRevealIsRefused(t *testing.T) {
	room := newTestRoom(t)
	voter := join(t, room, "Thomas")
	silent := join(t, room, "Anna")
	if err := room.Vote(voter, CardM); err != nil {
		t.Fatalf("Vote: %v", err)
	}
	if err := room.Reveal(voter); err != nil {
		t.Fatalf("Reveal: %v", err)
	}

	// Somebody who did not vote cannot record one after the fact.
	if err := room.Vote(silent, CardL); !errors.Is(err, ErrRoundRevealed) {
		t.Errorf("Vote after reveal: error = %v, want ErrRoundRevealed", err)
	}
	if got := cardOf(t, room, silent); got != NoCard {
		t.Errorf("the silent participant now holds %q, want no card", got)
	}

	// And somebody who did vote cannot change it after seeing everyone else's.
	if err := room.Vote(voter, CardXL); !errors.Is(err, ErrRoundRevealed) {
		t.Errorf("changing a vote after reveal: error = %v, want ErrRoundRevealed", err)
	}
	if got := cardOf(t, room, voter); got != CardM {
		t.Errorf("card = %q, want it unchanged at %q", got, CardM)
	}
}

func TestNewRoundClearsEveryVote(t *testing.T) {
	room := newTestRoom(t)
	first := join(t, room, "Thomas")
	second := join(t, room, "Anna")
	if err := room.Vote(first, CardM); err != nil {
		t.Fatalf("Vote: %v", err)
	}
	if err := room.Vote(second, CardL); err != nil {
		t.Fatalf("Vote: %v", err)
	}
	if err := room.Reveal(first); err != nil {
		t.Fatalf("Reveal: %v", err)
	}

	if err := room.NewRound(second); err != nil {
		t.Fatalf("NewRound: %v", err)
	}

	view := room.View()
	if view.Revealed {
		t.Error("the new round is revealed; it must start hidden")
	}
	if view.Results != nil {
		t.Error("the new round already carries results")
	}
	for _, p := range view.Participants {
		if p.Voted {
			t.Errorf("participant %q still holds a card in the new round", p.Name)
		}
	}
	// There is no history in this product: the previous round is simply gone.
	if got := len(room.round.votes); got != 0 {
		t.Errorf("the new round holds %d votes, want 0", got)
	}
}

func TestNewRoundMayBeStartedFromAHiddenRound(t *testing.T) {
	room := newTestRoom(t)
	first := join(t, room, "Thomas")
	second := join(t, room, "Anna")
	if err := room.Vote(first, CardM); err != nil {
		t.Fatalf("Vote: %v", err)
	}

	// Abandoning a round begun by mistake is the only way out of it.
	if err := room.NewRound(second); err != nil {
		t.Fatalf("NewRound from a hidden round: %v", err)
	}

	if room.Revealed() {
		t.Error("the new round is revealed")
	}
	for _, p := range room.View().Participants {
		if p.Voted {
			t.Errorf("participant %q kept a card across the new round", p.Name)
		}
	}
}

func TestNewRoundLeavesAwayParticipantsSeatedAndCardless(t *testing.T) {
	room := newTestRoom(t)
	present := join(t, room, "Thomas")
	absent := join(t, room, "Anna")
	if err := room.Vote(absent, CardM); err != nil {
		t.Fatalf("Vote: %v", err)
	}
	if err := room.MarkAway(absent); err != nil {
		t.Fatalf("MarkAway: %v", err)
	}

	if err := room.NewRound(present); err != nil {
		t.Fatalf("NewRound: %v", err)
	}

	view := room.View()
	if len(view.Participants) != 2 {
		t.Fatalf("table has %d participants, want 2 — a new round removes nobody", len(view.Participants))
	}
	for _, p := range view.Participants {
		if p.ID == absent {
			if !p.Away {
				t.Error("the away participant lost their away mark when a new round started")
			}
			if p.Voted {
				t.Error("the away participant kept a card in the new round")
			}
		}
	}
}

func TestEveryonePresentHasVoted(t *testing.T) {
	room := newTestRoom(t)

	// An empty table reports false: the condition is vacuously true, but its only
	// use is telling people that everyone has voted, which would be false here in
	// every sense that matters.
	if room.EveryonePresentHasVoted() {
		t.Error("an empty room reports that everyone present has voted")
	}

	present := join(t, room, "Thomas")
	absent := join(t, room, "Anna")

	if room.EveryonePresentHasVoted() {
		t.Error("nobody has voted, yet the round reports everyone present has")
	}

	if err := room.Vote(present, CardM); err != nil {
		t.Fatalf("Vote: %v", err)
	}
	if room.EveryonePresentHasVoted() {
		t.Error("one of two present participants has voted, yet the round reports everyone has")
	}

	// The away participant must not be counted: someone who has closed their laptop
	// cannot leave the round permanently incomplete.
	if err := room.MarkAway(absent); err != nil {
		t.Fatalf("MarkAway: %v", err)
	}
	if !room.EveryonePresentHasVoted() {
		t.Error("everyone present has voted, but the round reports otherwise")
	}
	if !room.View().EveryonePresentHasVoted {
		t.Error("the view disagrees with the room about whether everyone present has voted")
	}

	// Coming back without a vote makes the round incomplete again, which is correct:
	// they are here and they have not voted.
	if err := room.MarkPresent(absent); err != nil {
		t.Fatalf("MarkPresent: %v", err)
	}
	if room.EveryonePresentHasVoted() {
		t.Error("a returning participant without a vote did not make the round incomplete again")
	}
}
