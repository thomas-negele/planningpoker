package game

import (
	"crypto/rand"
	"slices"
	"testing"
)

func TestTShirtDeckHasExactlyTheAgreedCardsInOrder(t *testing.T) {
	deck := TShirtDeck()

	if deck.Name != TShirtDeckName {
		t.Errorf("deck name = %q, want %q", deck.Name, TShirtDeckName)
	}

	// The order is part of the deck, not incidental: it is the order the cards are
	// offered in and the order results are tallied in.
	want := []Card{CardXS, CardS, CardM, CardL, CardXL, CardUnknown, CardBreak}
	if !slices.Equal(deck.Cards, want) {
		t.Errorf("deck cards = %v, want %v", deck.Cards, want)
	}
}

func TestTShirtDeckScaleIsTheSizesOnly(t *testing.T) {
	deck := TShirtDeck()

	// The scale is what a display puts on an ordered axis, so its order matters for
	// the same reason the deck's does — and "?" and the coffee cup must not be on it.
	// They are not sizes, and a scale that included them would place "I cannot
	// estimate this" between two estimates.
	want := []Card{CardXS, CardS, CardM, CardL, CardXL}
	if !slices.Equal(deck.Scale, want) {
		t.Errorf("deck scale = %v, want %v", deck.Scale, want)
	}

	// The scale is a subset of the deck, not a second list that could drift from it.
	// A card on the scale that nobody can play would be a row nobody can fill.
	for _, card := range deck.Scale {
		if !slices.Contains(deck.Cards, card) {
			t.Errorf("scale contains %q, which is not a card of the deck", card)
		}
	}

	for _, card := range []Card{CardUnknown, CardBreak} {
		if slices.Contains(deck.Scale, card) {
			t.Errorf("scale contains %q, which expresses no size", card)
		}
	}
}

func TestDeckIsNotSharedBetweenCallers(t *testing.T) {
	// A caller who holds a Deck must not be able to reach back and reorder or
	// truncate the deck every other room is using.
	first := TShirtDeck()
	first.Cards[0] = "tampered"
	first.Scale[0] = "tampered"

	second := TShirtDeck()
	if second.Cards[0] != CardXS {
		t.Errorf("modifying one deck changed the next one: first card is %q, want %q",
			second.Cards[0], CardXS)
	}
	if second.Scale[0] != CardXS {
		t.Errorf("modifying one deck changed the next one: first scale card is %q, want %q",
			second.Scale[0], CardXS)
	}
}

func TestDeckMembership(t *testing.T) {
	deck := TShirtDeck()

	for _, card := range []Card{CardXS, CardS, CardM, CardL, CardXL, CardUnknown, CardBreak} {
		if !deck.Contains(card) {
			t.Errorf("deck does not contain %q, but it should", card)
		}
	}

	for _, card := range []Card{"XXL", "13", "m", "xs", " M", "M ", "☕☕"} {
		if deck.Contains(card) {
			t.Errorf("deck contains %q, but it should not", card)
		}
	}
}

func TestNoCardIsNotAPlayableCard(t *testing.T) {
	// NoCard is the absence of a vote, not something anyone can play.
	if TShirtDeck().Contains(NoCard) {
		t.Error("the deck reports NoCard as playable; it is the absence of a card")
	}
}

func TestNewRoomUsesTheTShirtDeck(t *testing.T) {
	room, err := NewRoom(rand.Reader, testMaxParticipants)
	if err != nil {
		t.Fatalf("NewRoom: %v", err)
	}

	deck := room.Deck()
	if deck.Name != TShirtDeckName {
		t.Errorf("new room's deck is %q, want %q", deck.Name, TShirtDeckName)
	}
	if !slices.Equal(deck.Cards, TShirtDeck().Cards) {
		t.Errorf("new room's cards = %v, want the t-shirt deck", deck.Cards)
	}
	if !slices.Equal(deck.Scale, TShirtDeck().Scale) {
		t.Errorf("new room's scale = %v, want the t-shirt deck's", deck.Scale)
	}
}
