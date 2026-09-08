package transport

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"de.thomasnegele.planningpoker/internal/game"
)

// gameSentinels lists every refusal the rules can produce. The test below checks
// this list against the domain's own source, so adding a sentinel there without
// adding it here fails rather than silently going unmapped.
var gameSentinels = []error{
	game.ErrUnknownParticipant,
	game.ErrEmptyName,
	game.ErrNameTooLong,
	game.ErrCardNotInDeck,
	game.ErrRoundRevealed,
	game.ErrInvalidRoomID,
	game.ErrShortRandomRead,
	game.ErrRoomFull,
	game.ErrInvalidCapacity,
}

func TestEverySentinelHasItsOwnCode(t *testing.T) {
	seen := map[string]error{}

	for _, sentinel := range gameSentinels {
		code, known := refusalCode(sentinel)
		if !known {
			t.Errorf("%v has no protocol code; an unmapped refusal would reach the browser as "+
				"a generic failure and nobody would know why the message was unhelpful", sentinel)
			continue
		}
		if previous, clash := seen[code]; clash && code != codeServerError {
			t.Errorf("%v and %v share the code %q; the interface could not tell them apart",
				sentinel, previous, code)
		}
		seen[code] = sentinel
	}

	// The transport's own refusal, for a message the rules never see at all.
	if code, known := refusalCode(errMalformedMessage); !known || code != codeBadMessage {
		t.Errorf("a malformed message maps to (%q, %v), want (%q, true)", code, known, codeBadMessage)
	}
}

func TestAnUnknownErrorIsNotQuietlyDowngraded(t *testing.T) {
	// refusalCode must report that it does not recognise an error rather than
	// returning a plausible-looking code. The caller decides what to do about it; a
	// silent fallback here would turn a new sentinel into "something went wrong".
	if code, known := refusalCode(errUnrecognised{}); known {
		t.Errorf("an unrecognised error mapped to %q; it must be reported as unknown", code)
	}
}

type errUnrecognised struct{}

func (errUnrecognised) Error() string { return "an error this protocol has never heard of" }

func TestTheSentinelListMatchesTheDomainSource(t *testing.T) {
	// Parsing the domain's errors.go is what makes the list above self-maintaining:
	// a sentinel added there and forgotten here fails this test instead of reaching
	// production as an unmapped refusal.
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "../game/errors.go", nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parsing the domain's errors: %v", err)
	}

	var declared []string
	ast.Inspect(file, func(n ast.Node) bool {
		spec, ok := n.(*ast.ValueSpec)
		if !ok {
			return true
		}
		for _, name := range spec.Names {
			if strings.HasPrefix(name.Name, "Err") && name.IsExported() {
				declared = append(declared, name.Name)
			}
		}
		return true
	})

	if len(declared) != len(gameSentinels) {
		t.Errorf("the domain declares %d sentinel errors (%v) but this package maps %d; "+
			"every one needs a protocol code", len(declared), declared, len(gameSentinels))
	}
}

// --- the guarantee this change exists to protect --------------------------

// stringsIn collects every string value reachable in a decoded JSON message,
// skipping the deck.
//
// Skipping the deck is essential rather than convenient. The deck is part of every
// snapshot on purpose — the client has to know which cards to offer — so it
// legitimately contains every card value. A search that included it would match on
// every message and could never fail, which would make this test worthless while
// looking reassuring.
func stringsIn(v any, path string, skip string, out map[string]string) {
	switch value := v.(type) {
	case map[string]any:
		for key, child := range value {
			next := path + "." + key
			if next == skip {
				continue
			}
			stringsIn(child, next, skip, out)
		}
	case []any:
		for _, child := range value {
			stringsIn(child, path, skip, out)
		}
	case string:
		out[value] = path
	}
}

// cardsOnTheWire reports any card value found in a message outside the deck.
func cardsOnTheWire(t *testing.T, raw []byte) map[string]string {
	t.Helper()

	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decoding %s: %v", raw, err)
	}

	found := map[string]string{}
	stringsIn(decoded, "", ".room.deck", found)

	leaked := map[string]string{}
	for _, card := range game.TShirtDeck().Cards {
		if where, present := found[string(card)]; present {
			leaked[string(card)] = where
		}
	}
	return leaked
}

// TestNoCardCrossesTheNetworkWhileHidden is the most important test in this
// package.
//
// The domain has its own version of this, and it proves something different: that
// the view struct cannot hold a hidden card. This one reads the bytes that actually
// crossed a real socket, because this is the layer where a leak would now happen —
// a field added to a message, a log line, an extra value assembled by hand.
//
// If it fails, the fix is in whatever put the value on the wire. It must never be
// weakened to make a build pass.
func TestNoCardCrossesTheNetworkWhileHidden(t *testing.T) {
	srv := newTestServer(t)
	roomID := srv.createGame(t)

	// Five participants, five different cards, so a leak of any one of them shows.
	names := []string{"anna", "bert", "cora", "dora", "emil"}
	cards := []game.Card{game.CardXS, game.CardS, game.CardM, game.CardL, game.CardXL}

	clients := make([]*client, 0, len(names))
	for i, name := range names {
		c := srv.dial(t, roomID)
		clients = append(clients, c)

		// Attaching broadcasts to every connection, so each of them has exactly one
		// message waiting. Reading them keeps the counting below exact.
		for _, other := range clients {
			other.state(t)
		}

		c.send(t, clientMessage{Type: intentSeat, Name: name})
		for _, other := range clients {
			other.state(t)
		}

		c.send(t, clientMessage{Type: intentVote, Card: string(cards[i])})
		for _, other := range clients {
			raw := other.raw(t)
			if leaked := cardsOnTheWire(t, raw); len(leaked) > 0 {
				t.Fatalf("a hidden round put card values on the wire: %v\nmessage: %s", leaked, raw)
			}
		}
	}

	// Everyone has now voted, which is the state in which a leak would matter most.
	last := clients[len(clients)-1]
	last.send(t, clientMessage{Type: intentRename, Name: "renamed"})
	for _, other := range clients {
		raw := other.raw(t)
		if leaked := cardsOnTheWire(t, raw); len(leaked) > 0 {
			t.Fatalf("a fully voted hidden round put card values on the wire: %v\nmessage: %s", leaked, raw)
		}
		var msg stateMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			t.Fatalf("decoding: %v", err)
		}
		if msg.Room.Results != nil {
			t.Error("a hidden round carried results, which would include the tally")
		}
		for _, p := range msg.Room.Participants {
			if !p.Voted {
				t.Errorf("participant %q voted but is not shown as having voted", p.Name)
			}
		}
	}
}

func TestRevealingPutsEveryCardOnTheWire(t *testing.T) {
	// The counterpart: hiding the votes must not be achieved by simply breaking the
	// reveal.
	srv := newTestServer(t)
	roomID := srv.createGame(t)

	first := srv.dial(t, roomID)
	first.state(t)
	first.seat(t, "anna")
	first.send(t, clientMessage{Type: intentVote, Card: string(game.CardM)})
	first.state(t)

	second := srv.dial(t, roomID, first.cookies...)
	second.state(t)
	first.state(t)

	first.send(t, clientMessage{Type: intentReveal})
	raw := first.raw(t)

	var msg stateMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if !msg.Room.Revealed {
		t.Fatal("the round is not revealed")
	}
	if msg.Room.Results == nil {
		t.Fatal("a revealed round carries no results")
	}
	if len(msg.Room.Results.Cards) != 1 || msg.Room.Results.Cards[0].Card != string(game.CardM) {
		t.Errorf("results cards = %+v, want one M", msg.Room.Results.Cards)
	}
	if len(msg.Room.Results.Tally) != 1 || msg.Room.Results.Tally[0].Count != 1 {
		t.Errorf("tally = %+v, want M counted once", msg.Room.Results.Tally)
	}
	if leaked := cardsOnTheWire(t, raw); len(leaked) == 0 {
		t.Error("a revealed round put no card values on the wire; the reveal shows nothing")
	}
}
