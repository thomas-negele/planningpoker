package game

import (
	"io"
	"strings"
	"unicode/utf8"
)

// MaxNameLength limits trimmed names to 15 Unicode code points. Keep the frontend
// validation and seat layout consistent with this limit.
const MaxNameLength = 15

// RoomID is the public URL identifier. Both generated and user-chosen IDs are
// allowed; anyone with the URL can access the room.
type RoomID string

// ParticipantID identifies a seat in public snapshots; it is not a seat credential.
type ParticipantID string

// participant holds a seat's identity, display name and presence.
type participant struct {
	id   ParticipantID
	name string
	// Away participants retain their seat and vote.
	away bool
}

// round holds the current votes and reveal state.
type round struct {
	revealed bool
	// Each participant has at most one vote.
	votes map[ParticipantID]Card
}

func newRound() round {
	return round{votes: make(map[ParticipantID]Card)}
}

// Room holds domain state and is not safe for concurrent use. The hub serializes
// access.
type Room struct {
	id   RoomID
	deck Deck
	// Keep participants in join order for stable snapshots.
	participants []*participant
	round        round

	// maxParticipants counts all seats, including away participants; rejoining adds
	// none.
	maxParticipants int
}

// NewRoom creates an empty room with a random ID and the T-shirt deck.
func NewRoom(random io.Reader, maxParticipants int) (*Room, error) {
	id, err := newID(random)
	if err != nil {
		return nil, err
	}
	return NewRoomAt(RoomID(id), maxParticipants)
}

// NewRoomAt creates an empty room at a valid chosen or generated ID.
func NewRoomAt(id RoomID, maxParticipants int) (*Room, error) {
	if !ValidRoomID(id) {
		return nil, ErrInvalidRoomID
	}
	if maxParticipants <= 0 {
		return nil, ErrInvalidCapacity
	}
	return &Room{
		id:              id,
		deck:            TShirtDeck(),
		round:           newRound(),
		maxParticipants: maxParticipants,
	}, nil
}

// ID returns the room's identifier.
func (r *Room) ID() RoomID { return r.id }

// Deck returns the room's deck.
func (r *Room) Deck() Deck { return r.deck }

// Join validates a name and allocates a seat. Duplicate display names are allowed.
// Randomness is read before mutating the room so failures leave it unchanged.
func (r *Room) Join(random io.Reader, name string) (ParticipantID, error) {
	clean, err := normalizeName(name)
	if err != nil {
		return "", err
	}

	// The limit applies to new seats; rejoining uses the existing participant.
	if len(r.participants) >= r.maxParticipants {
		return "", ErrRoomFull
	}

	id, err := newID(random)
	if err != nil {
		return "", err
	}

	r.participants = append(r.participants, &participant{id: ParticipantID(id), name: clean})
	return ParticipantID(id), nil
}

// Rejoin updates the name and marks an existing seat present, preserving its vote.
func (r *Room) Rejoin(id ParticipantID, name string) error {
	p, err := r.find(id)
	if err != nil {
		return err
	}

	clean, err := normalizeName(name)
	if err != nil {
		return err
	}

	p.name = clean
	p.away = false
	return nil
}

// Rename changes the display name without changing the seat or vote.
func (r *Room) Rename(id ParticipantID, name string) error {
	p, err := r.find(id)
	if err != nil {
		return err
	}

	clean, err := normalizeName(name)
	if err != nil {
		return err
	}

	p.name = clean
	return nil
}

// MarkAway preserves the participant and vote; away seats do not expire individually.
func (r *Room) MarkAway(id ParticipantID) error {
	p, err := r.find(id)
	if err != nil {
		return err
	}
	p.away = true
	return nil
}

// MarkPresent marks an existing participant present.
func (r *Room) MarkPresent(id ParticipantID) error {
	p, err := r.find(id)
	if err != nil {
		return err
	}
	p.away = false
	return nil
}

// AnyonePresent reports whether any participant is present. Room expiry in the hub
// uses open connections, including connections without a seat.
func (r *Room) AnyonePresent() bool {
	for _, p := range r.participants {
		if !p.away {
			return true
		}
	}
	return false
}

// Vote accepts a deck card from a seated participant while the round is hidden.
func (r *Room) Vote(id ParticipantID, card Card) error {
	if _, err := r.find(id); err != nil {
		return err
	}
	if r.round.revealed {
		return ErrRoundRevealed
	}
	if !r.deck.Contains(card) {
		return ErrCardNotInDeck
	}

	r.round.votes[id] = card
	return nil
}

// Reveal exposes the round for any seated participant, even before everyone votes.
// Revealing an already revealed round is idempotent.
func (r *Room) Reveal(id ParticipantID) error {
	if _, err := r.find(id); err != nil {
		return err
	}
	r.round.revealed = true
	return nil
}

// NewRound clears all votes, including away participants' votes, and hides the round.
func (r *Room) NewRound(id ParticipantID) error {
	if _, err := r.find(id); err != nil {
		return err
	}
	r.round = newRound()
	return nil
}

// Revealed reports whether the current round has been revealed.
func (r *Room) Revealed() bool { return r.round.revealed }

// EveryonePresentHasVoted excludes away participants and is false if none are present.
func (r *Room) EveryonePresentHasVoted() bool {
	present := 0
	for _, p := range r.participants {
		if p.away {
			continue
		}
		present++
		if _, voted := r.round.votes[p.id]; !voted {
			return false
		}
	}
	return present > 0
}

func (r *Room) find(id ParticipantID) (*participant, error) {
	for _, p := range r.participants {
		if p.id == id {
			return p, nil
		}
	}
	return nil, ErrUnknownParticipant
}

// normalizeName trims surrounding whitespace and checks the Unicode code-point limit.
func normalizeName(name string) (string, error) {
	clean := strings.TrimSpace(name)
	if clean == "" {
		return "", ErrEmptyName
	}
	if utf8.RuneCountInString(clean) > MaxNameLength {
		return "", ErrNameTooLong
	}
	return clean, nil
}
