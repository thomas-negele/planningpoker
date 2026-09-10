package game

import (
	"errors"
	"reflect"
	"testing"
)

func TestEmptyHiddenRoundChangesDeckImmediately(t *testing.T) {
	room := newTestRoom(t)
	id := join(t, room, "Thomas")

	if err := room.SetDeck(id, FibonacciDeckName); err != nil {
		t.Fatalf("SetDeck: %v", err)
	}
	view := room.View()
	if view.Deck.Name != FibonacciDeckName {
		t.Errorf("active deck = %q, want %q", view.Deck.Name, FibonacciDeckName)
	}
	if view.PendingDeck != nil {
		t.Errorf("empty hidden round has pending deck %+v, want none", view.PendingDeck)
	}
}

func TestHiddenVoteLocksDeckWithoutChangingRoom(t *testing.T) {
	room := newTestRoom(t)
	id := join(t, room, "Thomas")
	if err := room.Vote(id, CardM); err != nil {
		t.Fatalf("Vote: %v", err)
	}
	before := room.View()

	if err := room.SetDeck(id, FibonacciDeckName); !errors.Is(err, ErrDeckLocked) {
		t.Fatalf("SetDeck after vote error = %v, want ErrDeckLocked", err)
	}
	after := room.View()
	if !reflect.DeepEqual(after, before) {
		t.Errorf("refused deck change mutated room:\nbefore: %+v\nafter:  %+v", before, after)
	}
}

func TestRevealedRoundStoresLatestDeckForNextRound(t *testing.T) {
	room := newTestRoom(t)
	id := join(t, room, "Thomas")
	if err := room.Vote(id, CardM); err != nil {
		t.Fatalf("Vote: %v", err)
	}
	if err := room.Reveal(id); err != nil {
		t.Fatalf("Reveal: %v", err)
	}

	if err := room.SetDeck(id, FibonacciDeckName); err != nil {
		t.Fatalf("SetDeck Fibonacci: %v", err)
	}
	view := room.View()
	if view.Deck.Name != TShirtDeckName {
		t.Errorf("revealed round changed active deck to %q", view.Deck.Name)
	}
	if view.PendingDeck == nil || view.PendingDeck.Name != FibonacciDeckName {
		t.Fatalf("pending deck = %+v, want Fibonacci", view.PendingDeck)
	}
	if got := view.Results.Cards[0].Card; got != CardM {
		t.Errorf("revealed vote = %q after pending change, want %q", got, CardM)
	}

	if err := room.SetDeck(id, TShirtDeckName); err != nil {
		t.Fatalf("SetDeck T-shirt: %v", err)
	}
	if got := room.View().PendingDeck; got == nil || got.Name != TShirtDeckName {
		t.Fatalf("latest pending deck = %+v, want T-shirt", got)
	}
	if err := room.SetDeck(id, FibonacciDeckName); err != nil {
		t.Fatalf("SetDeck Fibonacci again: %v", err)
	}

	if err := room.NewRound(id); err != nil {
		t.Fatalf("NewRound: %v", err)
	}
	view = room.View()
	if view.Deck.Name != FibonacciDeckName {
		t.Errorf("new round active deck = %q, want Fibonacci", view.Deck.Name)
	}
	if view.PendingDeck != nil {
		t.Errorf("new round kept pending deck %+v", view.PendingDeck)
	}
	if view.Revealed || view.Results != nil || view.Participants[0].Voted {
		t.Errorf("new round is not empty and hidden: %+v", view)
	}
}

func TestInvalidDeckChangesAreRecognisableAndMutationFree(t *testing.T) {
	room := newTestRoom(t)
	id := join(t, room, "Thomas")
	before := room.View()

	if err := room.SetDeck(id, "custom"); !errors.Is(err, ErrUnknownDeck) {
		t.Errorf("SetDeck custom error = %v, want ErrUnknownDeck", err)
	}
	if err := room.SetDeck("nobody", FibonacciDeckName); !errors.Is(err, ErrUnknownParticipant) {
		t.Errorf("SetDeck by stranger error = %v, want ErrUnknownParticipant", err)
	}
	if after := room.View(); !reflect.DeepEqual(after, before) {
		t.Errorf("refused changes mutated room:\nbefore: %+v\nafter:  %+v", before, after)
	}
}
