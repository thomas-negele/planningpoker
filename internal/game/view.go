package game

// View is a snapshot of public room state. Individual card values and tallies are
// absent until reveal.
type View struct {
	RoomID RoomID

	// Deck lists all selectable cards in display order.
	Deck Deck

	Revealed bool

	// Participants are ordered by joining time.
	Participants []ParticipantView

	// EveryonePresentHasVoted is informational; it does not restrict reveal.
	EveryonePresentHasVoted bool

	// Results is nil while the round is hidden.
	Results *Results
}

// ParticipantView exposes identity, presence and whether a vote exists, but no card
// value.
type ParticipantView struct {
	ID   ParticipantID
	Name string

	// Away participants keep their seat and vote.
	Away bool

	// Voted reports whether a card has been selected.
	Voted bool
}

// Results contains revealed cards and their tally.
type Results struct {
	// Cards includes every participant; NoCard marks those who did not vote.
	Cards []ParticipantCard

	// Tally follows deck order and includes only selected cards. No averages are
	// computed.
	Tally []CardCount
}

// ParticipantCard pairs a participant with their revealed card or NoCard.
type ParticipantCard struct {
	ID   ParticipantID
	Card Card
}

// CardCount holds a selected card and its vote count.
type CardCount struct {
	Card  Card
	Count int
}

// View builds a public snapshot. Treat its deck slices as read-only; they share
// the room's deck storage.
func (r *Room) View() View {
	v := View{
		RoomID:                  r.id,
		Deck:                    r.deck,
		Revealed:                r.round.revealed,
		Participants:            make([]ParticipantView, 0, len(r.participants)),
		EveryonePresentHasVoted: r.EveryonePresentHasVoted(),
	}

	for _, p := range r.participants {
		_, voted := r.round.votes[p.id]
		v.Participants = append(v.Participants, ParticipantView{
			ID:    p.id,
			Name:  p.name,
			Away:  p.away,
			Voted: voted,
		})
	}

	if r.round.revealed {
		v.Results = r.results()
	}

	return v
}

// results builds the revealed cards and tally.
func (r *Room) results() *Results {
	res := &Results{Cards: make([]ParticipantCard, 0, len(r.participants))}

	for _, p := range r.participants {
		// A missing vote uses the map's zero value, NoCard.
		res.Cards = append(res.Cards, ParticipantCard{ID: p.id, Card: r.round.votes[p.id]})
	}

	// Iterate the deck to keep tally ordering stable.
	for _, card := range r.deck.Cards {
		count := 0
		for _, played := range r.round.votes {
			if played == card {
				count++
			}
		}
		if count > 0 {
			res.Tally = append(res.Tally, CardCount{Card: card, Count: count})
		}
	}

	return res
}
