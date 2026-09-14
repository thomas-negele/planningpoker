package transport

import (
	"encoding/json"
	"net/http"
	"slices"
	"testing"

	"de.thomasnegele.planningpoker/internal/game"
)

// fetchDecks asks the running application for the startable decks.
func fetchDecks(t *testing.T, srv *testServer) []deckOption {
	t.Helper()

	res, err := http.Get(srv.URL + "/api/decks")
	if err != nil {
		t.Fatalf("requesting the decks: %v", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusOK)
	}

	var decks []deckOption
	if err := json.NewDecoder(res.Body).Decode(&decks); err != nil {
		t.Fatalf("decoding the decks: %v", err)
	}
	return decks
}

func TestBothStartableDecksAreReportedWithTheirCards(t *testing.T) {
	srv := newTestServer(t)

	decks := fetchDecks(t, srv)
	if len(decks) != 2 {
		t.Fatalf("got %d decks, want 2", len(decks))
	}

	// The order is the order the entry screen offers them in, so it is part of the
	// answer rather than incidental.
	want := []deckOption{
		{
			Name:  game.TShirtDeckName,
			Label: "T-shirt sizes",
			Cards: []string{"XS", "S", "M", "L", "XL", "?", "☕"},
		},
		{
			Name:  game.FibonacciDeckName,
			Label: "Fibonacci",
			Cards: []string{"0", "½", "1", "2", "3", "5", "8", "13", "21", "?", "☕"},
		},
	}

	for i, wanted := range want {
		if decks[i].Name != wanted.Name {
			t.Errorf("deck %d name = %q, want %q", i, decks[i].Name, wanted.Name)
		}
		if decks[i].Label != wanted.Label {
			t.Errorf("deck %d label = %q, want %q", i, decks[i].Label, wanted.Label)
		}
		if !slices.Equal(decks[i].Cards, wanted.Cards) {
			t.Errorf("deck %d cards = %v, want %v", i, decks[i].Cards, wanted.Cards)
		}
	}
}

func TestListingDecksCreatesNothing(t *testing.T) {
	// The entry screen asks for this before the visitor has decided anything, and a
	// link-preview bot may ask for it too. Neither may cost a room or a seat.
	srv := newTestServer(t)
	before := srv.manager.Len()

	for range 3 {
		res, err := http.Get(srv.URL + "/api/decks")
		if err != nil {
			t.Fatalf("requesting the decks: %v", err)
		}
		if cookies := res.Cookies(); len(cookies) != 0 {
			t.Errorf("listing decks set %d cookies, want none: %v", len(cookies), cookies)
		}
		_ = res.Body.Close()
	}

	if got := srv.manager.Len(); got != before {
		t.Errorf("manager grew from %d rooms to %d by listing decks", before, got)
	}
}

func TestReportedCardsAreTheCardsDealt(t *testing.T) {
	// The whole reason this endpoint exists rather than a list in the page: what the
	// entry screen promises and what the table deals cannot be allowed to disagree.
	srv := newTestServer(t)

	for _, reported := range fetchDecks(t, srv) {
		roomID := srv.createGameWithDeck(t, reported.Name)
		dealt := srv.dial(t, roomID).state(t).Room.Deck

		if dealt.Name != reported.Name {
			t.Errorf("room deck = %q, want %q", dealt.Name, reported.Name)
		}
		if !slices.Equal(dealt.Cards, reported.Cards) {
			t.Errorf("deck %q: dealt %v, but the entry screen was told %v",
				reported.Name, dealt.Cards, reported.Cards)
		}
	}
}

func TestEveryStartableDeckCanActuallyStartAGame(t *testing.T) {
	// startableDecks is a second list beside the game's own catalogue, kept for the
	// labels. This is what stops the two drifting apart unnoticed.
	for _, entry := range startableDecks {
		if _, err := game.DeckByName(entry.name); err != nil {
			t.Errorf("deck %q is offered on the entry screen but unknown to the game: %v",
				entry.name, err)
		}
		if entry.label == "" {
			t.Errorf("deck %q is offered without a label", entry.name)
		}
	}
}
