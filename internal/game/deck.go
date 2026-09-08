package game

// Card is a selectable deck value or the NoCard sentinel.
type Card string

// The T-shirt deck contains sizes XS–XL, an unknown estimate and a break request.
// Unknown and break cards are excluded from the sizing scale.
const (
	CardXS      Card = "XS"
	CardS       Card = "S"
	CardM       Card = "M"
	CardL       Card = "L"
	CardXL      Card = "XL"
	CardUnknown Card = "?"
	CardBreak   Card = "☕"
)

// NoCard means no vote; it is not a selectable card.
const NoCard Card = ""

// TShirtDeckName identifies the currently supported deck.
const TShirtDeckName = "t-shirt"

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
