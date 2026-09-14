package transport

import (
	"encoding/json"
	"log"
	"net/http"

	"de.thomasnegele.planningpoker/internal/game"
)

// The decks a game can be started with, reported so that the entry screen can
// name their cards before any room exists. The alternative — a second list of
// cards written into the page — would be a second truth, and the two would
// disagree the first time a deck changed here.

// deckOption describes one startable deck to the entry screen.
type deckOption struct {
	// Name is the identifier POST /api/games expects.
	Name string `json:"name"`

	// Label is the wording the interface shows.
	Label string `json:"label"`

	// Cards are the deck's selectable values in the deck's own order, which is
	// the order they are offered in at the table.
	Cards []string `json:"cards"`
}

// startableDecks fixes the order the options are offered in. It is a list rather
// than a map so that the answer is stable between requests.
var startableDecks = []struct {
	name string

	// label is presentation rather than a rule of the game, which is why it lives
	// in the transport layer and not in internal/game.
	label string
}{
	{name: game.TShirtDeckName, label: "T-shirt sizes"},
	{name: game.FibonacciDeckName, label: "Fibonacci"},
}

// Decks reports the decks a game can be started with. It reads the same deck
// definitions room creation reads, so what a visitor is shown before starting a
// game cannot disagree with what they are dealt after starting it.
//
// The request creates nothing: no room, no seat, no cookie. It is safe to repeat.
func Decks(w http.ResponseWriter, _ *http.Request) {
	options := make([]deckOption, 0, len(startableDecks))
	for _, entry := range startableDecks {
		deck, err := game.DeckByName(entry.name)
		if err != nil {
			// Only reachable if this list and the game's catalogue drift apart, which
			// is a programming error rather than a bad request.
			log.Printf("listing decks: %v", err)
			http.Error(w, "could not list the decks", http.StatusInternalServerError)
			return
		}

		cards := make([]string, 0, len(deck.Cards))
		for _, card := range deck.Cards {
			cards = append(cards, string(card))
		}
		options = append(options, deckOption{Name: deck.Name, Label: entry.label, Cards: cards})
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// The catalogue is compiled in and changes only with a deployment, but a stale
	// cached copy would show cards a room does not deal. Revalidation is cheap.
	w.Header().Set("Cache-Control", "no-cache")

	if err := json.NewEncoder(w).Encode(options); err != nil {
		log.Printf("writing the decks response: %v", err)
	}
}
