package transport

import (
	"encoding/json"
	"strings"
	"testing"

	"de.thomasnegele.planningpoker/internal/game"
)

func TestAnOmittedVisitorFieldIsNotTheSameAsFalse(t *testing.T) {
	tests := map[string]struct {
		raw  string
		want *bool
	}{
		"omitted": {`{"type":"seat","name":"Thomas"}`, nil},
		"false":   {`{"type":"seat","name":"Thomas","visitor":false}`, boolPointer(false)},
		"true":    {`{"type":"seat","name":"Thomas","visitor":true}`, boolPointer(true)},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			msg, err := decodeClientMessage([]byte(test.raw))
			if err != nil {
				t.Fatalf("decodeClientMessage: %v", err)
			}
			switch {
			case test.want == nil && msg.Visitor != nil:
				t.Errorf("visitor = %v, want it absent so a rename preserves the seat's mode", *msg.Visitor)
			case test.want != nil && msg.Visitor == nil:
				t.Errorf("visitor is absent, want %v", *test.want)
			case test.want != nil && *msg.Visitor != *test.want:
				t.Errorf("visitor = %v, want %v", *msg.Visitor, *test.want)
			}
			// A seat intent has no third state: an absent field seats a voter.
			if got, want := msg.visitorRequested(), test.want != nil && *test.want; got != want {
				t.Errorf("visitorRequested() = %v, want %v", got, want)
			}
		})
	}
}

func boolPointer(v bool) *bool { return &v }

func TestSnapshotsSayWhoIsAVisitor(t *testing.T) {
	srv := newTestServer(t)
	roomID := srv.createGame(t)

	visitor := srv.dial(t, roomID)
	visitor.state(t)
	state := seatAsVisitor(t, visitor, "Thomas")

	if len(state.Room.Participants) != 1 {
		t.Fatalf("table has %d participants, want 1", len(state.Room.Participants))
	}
	if !state.Room.Participants[0].Visitor {
		t.Error("somebody who joined with the visitor box ticked is not marked as a visitor")
	}

	// The field is not omitted for a voter, so a client can rely on reading it.
	voter := srv.dial(t, roomID)
	voter.state(t)
	body, err := json.Marshal(voter.seat(t, "Bert"))
	if err != nil {
		t.Fatalf("marshalling the snapshot: %v", err)
	}
	if !strings.Contains(string(body), `"visitor"`) {
		t.Errorf("a snapshot carries no visitor field: %s", body)
	}
}

// seatAsVisitor takes a seat with the visitor box ticked and returns the snapshot
// that follows, as client.seat does for a voter.
func seatAsVisitor(t *testing.T, c *client, name string) stateMessage {
	t.Helper()
	c.send(t, clientMessage{Type: intentSeat, Name: name, Visitor: boolPointer(true)})
	return c.state(t)
}

func TestSeatingWithoutTheFieldSeatsAVoter(t *testing.T) {
	srv := newTestServer(t)
	roomID := srv.createGame(t)

	player := srv.dial(t, roomID)
	player.state(t)
	state := player.seat(t, "Thomas")

	if state.Room.Participants[0].Visitor {
		t.Error("a seat intent without a visitor field produced a visitor")
	}
}

func TestAVisitorsVoteIsRefusedWithItsOwnCode(t *testing.T) {
	srv := newTestServer(t)
	roomID := srv.createGame(t)

	visitor := srv.dial(t, roomID)
	visitor.state(t)
	seatAsVisitor(t, visitor, "Thomas")

	// A hand-written message is exactly the case the server has to refuse: the
	// browser would not offer the deck at all.
	visitor.send(t, clientMessage{Type: intentVote, Card: string(game.CardM)})
	refusal := visitor.refusal(t)

	if refusal.Code != codeVisitorVote {
		t.Errorf("refusal code = %q, want %q", refusal.Code, codeVisitorVote)
	}
}

func TestRenamingWithoutTheFieldLeavesVisitorModeAlone(t *testing.T) {
	srv := newTestServer(t)
	roomID := srv.createGame(t)

	visitor := srv.dial(t, roomID)
	visitor.state(t)
	seatAsVisitor(t, visitor, "Thomas")

	visitor.send(t, clientMessage{Type: intentRename, Name: "Thomas N."})
	state := visitor.state(t)

	if !state.Room.Participants[0].Visitor {
		t.Error("a rename that named no mode switched visitor mode off")
	}
	if state.Room.Participants[0].Name != "Thomas N." {
		t.Errorf("name = %q, want %q", state.Room.Participants[0].Name, "Thomas N.")
	}
}

func TestRenamingAppliesTheNameAndTheModeTogether(t *testing.T) {
	srv := newTestServer(t)
	roomID := srv.createGame(t)

	player := srv.dial(t, roomID)
	player.state(t)
	player.seat(t, "Thomas")

	// A vote cast while the round is hidden disappears with the switch.
	player.send(t, clientMessage{Type: intentVote, Card: string(game.CardM)})
	if !player.state(t).Room.Participants[0].Voted {
		t.Fatal("the vote was not recorded")
	}

	player.send(t, clientMessage{Type: intentRename, Name: "Thomas", Visitor: boolPointer(true)})
	state := player.state(t)
	if !state.Room.Participants[0].Visitor {
		t.Error("saving visitor mode did not take effect")
	}
	if state.Room.Participants[0].Voted {
		t.Error("the hidden vote survived the switch into visitor mode")
	}

	// An invalid name must change neither field, which is why the two travel in one
	// message.
	player.send(t, clientMessage{Type: intentRename, Name: "   ", Visitor: boolPointer(false)})
	if code := player.refusal(t).Code; code != codeNameEmpty {
		t.Fatalf("refusal code = %q, want %q", code, codeNameEmpty)
	}

	player.send(t, clientMessage{Type: intentReveal})
	state = player.state(t)
	if !state.Room.Participants[0].Visitor {
		t.Error("a refused rename still switched visitor mode off")
	}
	if state.Room.Participants[0].Name != "Thomas" {
		t.Errorf("name = %q, want it unchanged as %q", state.Room.Participants[0].Name, "Thomas")
	}
}
