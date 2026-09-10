package game

// Card is a selectable deck value or the NoCard sentinel.
type Card string

// The supported decks contain sizing choices, an unknown estimate and a break
// request. Unknown and break cards are excluded from their sizing scales.
const (
	CardXS        Card = "XS"
	CardS         Card = "S"
	CardM         Card = "M"
	CardL         Card = "L"
	CardXL        Card = "XL"
	CardZero      Card = "0"
	CardHalf      Card = "½"
	CardOne       Card = "1"
	CardTwo       Card = "2"
	CardThree     Card = "3"
	CardFive      Card = "5"
	CardEight     Card = "8"
	CardThirteen  Card = "13"
	CardTwentyOne Card = "21"
	CardUnknown   Card = "?"
	CardBreak     Card = "☕"
)

// NoCard means no vote; it is not a selectable card.
const NoCard Card = ""

// Stable deck names cross the HTTP and WebSocket protocols.
const (
	TShirtDeckName    = "t-shirt"
	FibonacciDeckName = "fibonacci"
)

// Deck defines selectable cards and the ordered sizing scale.
type Deck struct {
	Name  string
	Cards []Card

	// Scale excludes non-sizing cards (? and coffee) and preserves size order.
	Scale []Card
}

// TShirtDeck returns fresh slices so callers cannot modify a shared deck.
func TShirtDeck() Deck {
	return Deck{
		Name:  TShirtDeckName,
		Cards: []Card{CardXS, CardS, CardM, CardL, CardXL, CardUnknown, CardBreak},
		Scale: []Card{CardXS, CardS, CardM, CardL, CardXL},
	}
}

// FibonacciDeck returns the agreed modified Fibonacci scale through 21.
func FibonacciDeck() Deck {
	return Deck{
		Name: FibonacciDeckName,
		Cards: []Card{
			CardZero, CardHalf, CardOne, CardTwo, CardThree, CardFive,
			CardEight, CardThirteen, CardTwentyOne, CardUnknown, CardBreak,
		},
		Scale: []Card{
			CardZero, CardHalf, CardOne, CardTwo, CardThree, CardFive,
			CardEight, CardThirteen, CardTwentyOne,
		},
	}
}

// DeckByName returns a fresh supported deck or rejects the name. The catalogue is
// deliberately closed; rooms cannot construct custom decks through public input.
func DeckByName(name string) (Deck, error) {
	switch name {
	case TShirtDeckName:
		return TShirtDeck(), nil
	case FibonacciDeckName:
		return FibonacciDeck(), nil
	default:
		return Deck{}, ErrUnknownDeck
	}
}

// Contains reports whether a card is selectable; NoCard is excluded.
func (d Deck) Contains(card Card) bool {
	if card == NoCard {
		return false
	}
	for _, c := range d.Cards {
		if c == card {
			return true
		}
	}
	return false
}
