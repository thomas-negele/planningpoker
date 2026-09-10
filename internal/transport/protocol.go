package transport

import (
	"encoding/json"
	"errors"
	"fmt"

	"de.thomasnegele.planningpoker/internal/game"
	"de.thomasnegele.planningpoker/internal/hub"
)

// Wire types and conversions between JSON messages and domain snapshots.

// Client messages.

// intent names the things a participant can ask for. Anything else is refused.
const (
	intentSeat     = "seat"
	intentVote     = "vote"
	intentReveal   = "reveal"
	intentNewRound = "newRound"
	intentSetDeck  = "setDeck"
	intentRename   = "rename"
)

// clientMessage carries a client intent and its optional payload.
type clientMessage struct {
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
	Card string `json:"card,omitempty"`
	Deck string `json:"deck,omitempty"`
}

// decodeClientMessage parses JSON and rejects missing or unknown intent names.
// Domain operations validate the name and card values.
func decodeClientMessage(raw []byte) (clientMessage, error) {
	var msg clientMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		return clientMessage{}, fmt.Errorf("%w: %v", errMalformedMessage, err)
	}

	switch msg.Type {
	case intentSeat, intentVote, intentReveal, intentNewRound, intentSetDeck, intentRename:
		return msg, nil
	case "":
		return clientMessage{}, fmt.Errorf("%w: no intent named", errMalformedMessage)
	default:
		return clientMessage{}, fmt.Errorf("%w: unknown intent %q", errMalformedMessage, msg.Type)
	}
}

// errMalformedMessage identifies JSON or intent decoding failures.
var errMalformedMessage = errors.New("message could not be understood")

// errTooFast refuses a message before it reaches the room.
var errTooFast = errors.New("messages are arriving faster than this connection may send them")

// Server messages.

const (
	messageState = "state"
	messageError = "error"
)

// stateMessage is the public room snapshot sent to a connection.
type stateMessage struct {
	Type string `json:"type"`

	// You is a public participant ID; private seat credentials are not included in
	// snapshots.
	You string `json:"you"`

	Room roomMessage `json:"room"`
}

type roomMessage struct {
	ID                      string             `json:"id"`
	Deck                    deckMessage        `json:"deck"`
	PendingDeck             *deckMessage       `json:"pendingDeck,omitempty"`
	Revealed                bool               `json:"revealed"`
	Participants            []participantEntry `json:"participants"`
	EveryonePresentHasVoted bool               `json:"everyonePresentHasVoted"`

	// Results is omitted while the round is hidden.
	Results *resultsMessage `json:"results,omitempty"`
}

type deckMessage struct {
	Name  string   `json:"name"`
	Cards []string `json:"cards"`

	// Scale lists sizing cards in display order, excluding non-sizing choices.
	Scale []string `json:"scale"`
}

// participantEntry exposes presence and vote status without hidden card values.
type participantEntry struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Away  bool   `json:"away"`
	Voted bool   `json:"voted"`
}

type resultsMessage struct {
	Cards []participantCard `json:"cards"`
	Tally []cardCount       `json:"tally"`
}

type participantCard struct {
	ID string `json:"id"`
	// Card is empty for somebody who did not vote, which is distinct from having
	// voted for the unknown card.
	Card string `json:"card"`
}

type cardCount struct {
	Card  string `json:"card"`
	Count int    `json:"count"`
}

// errorMessage carries a machine-readable refusal and fallback text.
type errorMessage struct {
	Type string `json:"type"`
	// Code is stable and machine-readable, so the interface can say something
	// specific instead of "something went wrong".
	Code string `json:"code"`
	// Message describes the refusal; the frontend maps known codes to its own text.
	Message string `json:"message"`
}

// These codes are part of the client/server protocol; keep the frontend mapping in
// sync.
const (
	codeNotSeated     = "not_seated"
	codeNameEmpty     = "name_empty"
	codeNameTooLong   = "name_too_long"
	codeCardNotInDeck = "card_not_in_deck"
	codeUnknownDeck   = "unknown_deck"
	codeDeckLocked    = "deck_locked"
	codeRoundRevealed = "round_revealed"
	codeBadMessage    = "bad_message"
	codeInvalidRoomID = "invalid_room_id"
	codeServerError   = "server_error"
	codeRoomFull      = "room_full"
	codeAtCapacity    = "at_capacity"
	codeTooFast       = "too_fast"

	// Connection capacity is distinct from server room capacity and participant seat
	// capacity.
	codeTooManyConnections = "too_many_connections"
)

// refusal pairs one sentinel error with the code that stands for it on the wire.
type refusal struct {
	err  error
	code string
}

// refusals maps known errors to protocol codes and messages.
var refusals = []refusal{
	{game.ErrUnknownParticipant, codeNotSeated},
	{game.ErrEmptyName, codeNameEmpty},
	{game.ErrNameTooLong, codeNameTooLong},
	{game.ErrCardNotInDeck, codeCardNotInDeck},
	{game.ErrUnknownDeck, codeUnknownDeck},
	{game.ErrDeckLocked, codeDeckLocked},
	{game.ErrRoundRevealed, codeRoundRevealed},
	{game.ErrInvalidRoomID, codeInvalidRoomID},
	{game.ErrShortRandomRead, codeServerError},
	{game.ErrRoomFull, codeRoomFull},
	// An invalid configured capacity is an internal error, not a full table.
	{game.ErrInvalidCapacity, codeServerError},
	{hub.ErrAtCapacity, codeAtCapacity},
	{hub.ErrRoomAtCapacity, codeTooManyConnections},
	{errTooFast, codeTooFast},
	{errMalformedMessage, codeBadMessage},
}

// refusalCode returns the code for an error, and false if it is not one this
// protocol knows how to describe.
func refusalCode(err error) (string, bool) {
	for _, r := range refusals {
		if errors.Is(err, r.err) {
			return r.code, true
		}
	}
	return "", false
}

// Snapshot encoding.

// encodeUpdate turns what a room produced into the bytes for one connection.
func encodeUpdate(u hub.Update) ([]byte, error) {
	if u.Err != nil {
		code, known := refusalCode(u.Err)
		if !known {
			// Keep unexpected errors from exposing internal details.
			code = codeServerError
		}
		return json.Marshal(errorMessage{Type: messageError, Code: code, Message: u.Err.Error()})
	}

	if u.View == nil {
		return nil, errors.New("update carried neither a view nor an error")
	}
	return json.Marshal(stateMessage{
		Type: messageState,
		You:  string(u.You),
		Room: roomFromView(*u.View),
	})
}

// roomFromView converts domain snapshots to their wire representation.
func roomFromView(v game.View) roomMessage {
	deck := deckFromView(v.Deck)

	participants := make([]participantEntry, 0, len(v.Participants))
	for _, p := range v.Participants {
		participants = append(participants, participantEntry{
			ID:    string(p.ID),
			Name:  p.Name,
			Away:  p.Away,
			Voted: p.Voted,
		})
	}

	msg := roomMessage{
		ID:                      string(v.RoomID),
		Deck:                    deck,
		Revealed:                v.Revealed,
		Participants:            participants,
		EveryonePresentHasVoted: v.EveryonePresentHasVoted,
	}
	if v.PendingDeck != nil {
		pending := deckFromView(*v.PendingDeck)
		msg.PendingDeck = &pending
	}

	// Leave Results nil until the domain supplies revealed results.
	if v.Results != nil {
		msg.Results = resultsFromView(*v.Results)
	}

	return msg
}

func deckFromView(d game.Deck) deckMessage {
	deck := deckMessage{
		Name:  d.Name,
		Cards: make([]string, 0, len(d.Cards)),
		Scale: make([]string, 0, len(d.Scale)),
	}
	for _, c := range d.Cards {
		deck.Cards = append(deck.Cards, string(c))
	}
	for _, c := range d.Scale {
		deck.Scale = append(deck.Scale, string(c))
	}
	return deck
}

func resultsFromView(r game.Results) *resultsMessage {
	out := &resultsMessage{
		Cards: make([]participantCard, 0, len(r.Cards)),
		Tally: make([]cardCount, 0, len(r.Tally)),
	}
	for _, c := range r.Cards {
		out.Cards = append(out.Cards, participantCard{ID: string(c.ID), Card: string(c.Card)})
	}
	for _, t := range r.Tally {
		out.Tally = append(out.Tally, cardCount{Card: string(t.Card), Count: t.Count})
	}
	return out
}
