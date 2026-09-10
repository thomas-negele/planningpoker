package game

import "errors"

// Domain errors are sentinels; callers can classify them with errors.Is.
var (
	// ErrUnknownParticipant means the seat does not exist.
	ErrUnknownParticipant = errors.New("no such participant in this room")

	// ErrEmptyName means the trimmed name is empty.
	ErrEmptyName = errors.New("a name is required")

	// ErrNameTooLong means the trimmed name exceeds MaxNameLength.
	ErrNameTooLong = errors.New("name is too long")

	// ErrCardNotInDeck rejects cards outside the room's deck.
	ErrCardNotInDeck = errors.New("card is not in this room's deck")

	// ErrUnknownDeck rejects names outside the two supported room decks.
	ErrUnknownDeck = errors.New("unknown room deck")

	// ErrDeckLocked rejects deck changes after a hidden round has received a vote.
	ErrDeckLocked = errors.New("the deck cannot change while voting is in progress")

	// ErrRoundRevealed rejects voting after reveal.
	ErrRoundRevealed = errors.New("the round has been revealed and its votes are final")

	// ErrInvalidRoomID rejects identifiers outside the allowed length and alphabet.
	ErrInvalidRoomID = errors.New("not a room identifier")

	// ErrRoomFull means no new participant seat is available.
	ErrRoomFull = errors.New("this room's seats are all taken")

	// ErrInvalidCapacity rejects nonpositive seat limits.
	ErrInvalidCapacity = errors.New("a room's seat limit must be greater than zero")

	// ErrShortRandomRead means the entropy source did not supply enough bytes.
	ErrShortRandomRead = errors.New("could not read enough random bytes for an identifier")
)
