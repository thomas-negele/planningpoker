package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"de.thomasnegele.planningpoker/internal/hub"
	"github.com/coder/websocket"
)

// --- nothing is created before a connection exists -------------------------
//
// These are the regressions for the order of operations in Socket. Before it was
// reordered, the room was ensured on the handler's first line, so every request
// below created one.

func TestARequestThatNeverBecomesAConnectionCreatesNothing(t *testing.T) {
	srv := newTestServer(t)

	for _, tc := range []struct {
		name    string
		request func() (*http.Request, error)
	}{
		{
			"an ordinary GET with no upgrade headers at all",
			func() (*http.Request, error) {
				return http.NewRequest(http.MethodGet, srv.URL+"/ws/never-created-one", nil)
			},
		},
		{
			"a GET claiming to upgrade but carrying no key",
			func() (*http.Request, error) {
				r, err := http.NewRequest(http.MethodGet, srv.URL+"/ws/never-created-two", nil)
				if err != nil {
					return nil, err
				}
				r.Header.Set("Connection", "Upgrade")
				r.Header.Set("Upgrade", "websocket")
				return r, nil
			},
		},
		{
			"a complete upgrade from another website",
			func() (*http.Request, error) {
				r, err := http.NewRequest(http.MethodGet, srv.URL+"/ws/never-created-three", nil)
				if err != nil {
					return nil, err
				}
				r.Header.Set("Connection", "Upgrade")
				r.Header.Set("Upgrade", "websocket")
				r.Header.Set("Sec-WebSocket-Version", "13")
				r.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
				r.Header.Set("Origin", "https://evil.example.com")
				return r, nil
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := srv.manager.Len()

			req, err := tc.request()
			if err != nil {
				t.Fatalf("building the request: %v", err)
			}
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("making the request: %v", err)
			}
			defer func() { _ = res.Body.Close() }()

			if got := srv.manager.Len(); got != before {
				t.Errorf("%s created a room (%d rooms, was %d)", tc.name, got, before)
			}
			if cookies := res.Cookies(); len(cookies) != 0 {
				t.Errorf("%s was issued %d cookies: %v — a request that never becomes a "+
					"connection must be given no seat token", tc.name, len(cookies), cookies)
			}
			if res.StatusCode == http.StatusSwitchingProtocols {
				t.Errorf("%s was upgraded", tc.name)
			}
		})
	}
}

func TestAnUnusableNameStillCreatesNothingAndCostsNoToken(t *testing.T) {
	// An unusable identifier does get a socket, because the close code is the only
	// way to tell the page which rule the name broke. What it must not get is a room
	// or a seat token.
	srv := newTestServer(t)
	before := srv.manager.Len()

	c := srv.dial(t, "no")
	if len(c.cookies) != 0 {
		t.Errorf("an unusable room name was issued %d cookies, want 0", len(c.cookies))
	}
	if got := srv.manager.Len(); got != before {
		t.Errorf("an unusable room name created a room (%d, was %d)", got, before)
	}
}

// --- starting a game from another website ---------------------------------

func TestStartingAGameIsRefusedFromAnotherWebsite(t *testing.T) {
	srv := newTestServer(t)

	for _, tc := range []struct {
		name   string
		origin string
		want   int
	}{
		{"no origin at all, which is not a browser acting for a page", "", http.StatusOK},
		{"this site", "", http.StatusOK}, // filled in below, once the server URL is known
		{"another website", "https://evil.example.com", http.StatusForbidden},
	} {
		origin := tc.origin
		if tc.name == "this site" {
			origin = srv.URL
		}

		before := srv.manager.Len()

		req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/games", nil)
		if err != nil {
			t.Fatalf("building the request: %v", err)
		}
		if origin != "" {
			req.Header.Set("Origin", origin)
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("posting: %v", err)
		}
		_ = res.Body.Close()

		if res.StatusCode != tc.want {
			t.Errorf("starting a game with origin %q: status %d, want %d", tc.name, res.StatusCode, tc.want)
		}

		created := srv.manager.Len() - before
		if tc.want == http.StatusForbidden && created != 0 {
			t.Errorf("a refused cross-site request still created %d rooms", created)
		}
		if tc.want == http.StatusOK && created != 1 {
			t.Errorf("an allowed request created %d rooms, want 1", created)
		}
	}
}

// --- the ceilings, seen from a browser ------------------------------------

func TestAFullServerSaysSoRatherThanFailingVaguely(t *testing.T) {
	srv := newTestServerWith(t, hub.Limits{Rooms: 1, ConnectionsPerRoom: 10, ParticipantsPerRoom: 10}, testRate)
	srv.dial(t, "the-only-room")

	// The socket opens — the handshake has already succeeded by the time the room is
	// asked for — and the refusal arrives on it, with a code the page can act on.
	// This is the server's own ceiling; a room running out of connections is a
	// different refusal, covered above.
	c := srv.dial(t, "one-room-too-many")
	if got := c.refusal(t); got.Code != codeAtCapacity {
		t.Errorf("a full server refused with code %q, want %q", got.Code, codeAtCapacity)
	}

	if got := srv.manager.Len(); got != 1 {
		t.Errorf("the server holds %d rooms after refusing one, want 1", got)
	}
}

func TestAFullRoomAndAFullServerAreDifferentRefusals(t *testing.T) {
	// The defect this replaces: both were reported as at_capacity, so somebody whose
	// room had too many tabs open was told the *server* was busy — on a server with
	// one room on it. The two need opposite advice, so they need different reasons.
	t.Run("the process is out of rooms", func(t *testing.T) {
		srv := newTestServerWith(t, hub.Limits{Rooms: 1, ConnectionsPerRoom: 10, ParticipantsPerRoom: 10}, testRate)
		srv.dial(t, "the-only-room")

		if got := srv.dial(t, "one-room-too-many").refusal(t); got.Code != codeAtCapacity {
			t.Errorf("a full server refused with %q, want %q", got.Code, codeAtCapacity)
		}
	})

	t.Run("one room is out of connections", func(t *testing.T) {
		srv := newTestServerWith(t, hub.Limits{Rooms: 10, ConnectionsPerRoom: 2, ParticipantsPerRoom: 10}, testRate)
		roomID := srv.createGame(t)

		first := srv.dial(t, roomID)
		first.state(t)
		second := srv.dial(t, roomID)
		second.state(t)

		got := srv.dial(t, roomID).refusal(t)
		if got.Code != codeTooManyConnections {
			t.Errorf("a full room refused with %q, want %q", got.Code, codeTooManyConnections)
		}
		if got.Code == codeAtCapacity {
			t.Error("a full room was reported as a full server, which is the defect this test exists for")
		}
	})
}

func TestAFullTableRefusesASeatWithoutDisturbingTheGame(t *testing.T) {
	srv := newTestServerWith(t, hub.Limits{Rooms: 10, ConnectionsPerRoom: 10, ParticipantsPerRoom: 1}, testRate)
	roomID := srv.createGame(t)

	seated := srv.dial(t, roomID)
	seated.state(t)
	seated.send(t, clientMessage{Type: intentSeat, Name: "Thomas"})
	seated.state(t)
	seated.send(t, clientMessage{Type: intentVote, Card: "M"})
	seated.state(t)

	late := srv.dial(t, roomID)
	late.state(t)
	// Attaching is itself a change everybody sees, so the seated player gets a
	// snapshot for it. That one is expected; what must not follow is a snapshot
	// caused by the refused seat.
	seated.state(t)

	late.send(t, clientMessage{Type: intentSeat, Name: "Anna"})
	refusal := late.refusal(t)
	if refusal.Code != codeRoomFull {
		t.Errorf("a full table refused with code %q, want %q", refusal.Code, codeRoomFull)
	}

	// The refused connection stays open, so somebody turned away still sees the table.
	late.send(t, clientMessage{Type: intentVote, Card: "L"})
	if got := late.refusal(t); got.Code != codeNotSeated {
		t.Errorf("the refused connection was not still usable: %+v", got)
	}

	// And nobody at the table learned any of this happened.
	assertNothingArrives(t, seated)
}

// --- the message rate ------------------------------------------------------

func TestAFullTableVotingAtOnceIsNeverSlowed(t *testing.T) {
	// This is the behaviour the rate limit must never break, and the reason it is
	// per connection rather than per room. Twenty people playing a card in the same
	// second is one message each, not twenty on anybody's allowance.
	const seats = 20
	srv := newTestServerWith(t,
		hub.Limits{Rooms: 10, ConnectionsPerRoom: 40, ParticipantsPerRoom: seats},
		RateLimit{PerSecond: 10, Burst: 20})
	roomID := srv.createGame(t)

	clients := make([]*client, seats)
	stop := make(chan struct{})
	defer close(stop)
	for i := range clients {
		clients[i] = srv.dial(t, roomID)
		clients[i].state(t)
		clients[i].send(t, clientMessage{Type: intentSeat, Name: "Player"})
		clients[i].state(t)

		// Start draining as soon as this participant sits. Under race
		// instrumentation, waiting until all twenty have joined can otherwise fill
		// an early participant's reliable snapshot queue during test setup.
		go func(c *client) {
			for {
				select {
				case <-stop:
					return
				default:
				}
				if _, err := c.readWithin(5 * time.Second); err != nil {
					return
				}
			}
		}(clients[i])
	}

	// Every one of these connections must keep reading for the whole test. A table
	// of twenty produces a snapshot per person per action, and a connection that
	// stops reading is dropped by the hub for being slow — which would look exactly
	// like the rate limit refusing it, and prove nothing.
	// The drainers above already cover both setup and the simultaneous vote.

	// Everybody votes at once, from their own connection.
	var wg sync.WaitGroup
	for _, c := range clients {
		wg.Add(1)
		go func(c *client) {
			defer wg.Done()
			c.send(t, clientMessage{Type: intentVote, Card: "M"})
		}(c)
	}
	wg.Wait()

	// A connection arriving now is sent the room as it stands. If any of those
	// simultaneous votes had been refused for rate, this snapshot would never show a
	// full table with everybody voted.
	//
	// Each attempt closes its connection again before the next one. Leaving them open
	// would walk the room into its own connection ceiling after enough retries — and
	// the refusal for that arrives here as a capacity error, which would look exactly
	// like the failure this test is meant to detect while having nothing to do with it.
	deadline := time.Now().Add(10 * time.Second)
	for {
		if time.Now().After(deadline) {
			t.Fatal("not every simultaneous vote was recorded; the rate limit refused a real table")
		}

		observer := srv.dial(t, roomID)
		msg := observer.state(t)
		_ = observer.conn.CloseNow()

		if len(msg.Room.Participants) == seats && msg.Room.EveryonePresentHasVoted {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func TestGoingOverTheRateIsAnsweredWithItsOwnReason(t *testing.T) {
	// Deliberately small and deliberate: just past the burst, so what is observed is
	// the limit answering rather than a race between a flood and a buffer.
	srv := newTestServerWith(t,
		hub.Limits{Rooms: 10, ConnectionsPerRoom: 10, ParticipantsPerRoom: 10},
		RateLimit{PerSecond: 1, Burst: 2})
	roomID := srv.createGame(t)

	c := srv.dial(t, roomID)
	c.state(t)

	// Two messages fit the burst; the third does not, and none of them may be
	// answered with anything but their own reason.
	c.send(t, clientMessage{Type: intentSeat, Name: "Thomas"})
	c.state(t)
	c.send(t, clientMessage{Type: intentRename, Name: "Thomas II"})
	c.state(t)
	c.send(t, clientMessage{Type: intentRename, Name: "Thomas III"})

	if got := c.refusal(t); got.Code != codeTooFast {
		t.Errorf("the message past the burst was refused with %q, want %q", got.Code, codeTooFast)
	}
}

func TestAFloodNeverReachesTheRoomAndEndsItsOwnConnection(t *testing.T) {
	srv := newTestServerWith(t,
		hub.Limits{Rooms: 10, ConnectionsPerRoom: 10, ParticipantsPerRoom: 10},
		RateLimit{PerSecond: 1, Burst: 2})
	roomID := srv.createGame(t)

	observer := srv.dial(t, roomID)
	observer.state(t)
	observer.send(t, clientMessage{Type: intentSeat, Name: "Anna"})
	observer.state(t)

	flooder := srv.dial(t, roomID)
	flooder.state(t)

	// Well past the burst, as fast as the socket takes them. Each one would rename
	// somebody if it reached the room.
	for i := 0; i < 50; i++ {
		body, err := json.Marshal(clientMessage{Type: intentRename, Name: "Flood"})
		if err != nil {
			t.Fatalf("marshalling: %v", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		err = flooder.conn.Write(ctx, websocket.MessageText, body)
		cancel()
		if err != nil {
			break // the connection has already been ended, which is a permitted outcome
		}
	}

	// The connection ends, one way or another: refused until it is closed, or dropped
	// for not keeping up with its own refusals. Either way it does not go on being
	// served.
	ended := false
	for i := 0; i < 100; i++ {
		if _, err := flooder.readWithin(2 * time.Second); err != nil {
			ended = true
			break
		}
	}
	if !ended {
		t.Error("a connection sending far above its rate was served indefinitely")
	}

	// The point of all of it: not one of those renames reached the room. The observer
	// sees snapshots — connections coming and going are real changes — but never a
	// table with anybody called Flood at it.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		body, err := observer.readWithin(300 * time.Millisecond)
		if err != nil {
			break
		}
		var msg stateMessage
		if json.Unmarshal(body, &msg) != nil {
			continue
		}
		for _, p := range msg.Room.Participants {
			if p.Name == "Flood" {
				t.Fatalf("a rename from a flooding connection reached the room: %+v", msg.Room.Participants)
			}
		}
	}
}

// --- the message size limit ------------------------------------------------

func TestAnOversizedMessageIsRefusedAndChangesNothing(t *testing.T) {
	srv := newTestServer(t)
	roomID := srv.createGame(t)

	observer := srv.dial(t, roomID)
	observer.state(t)
	observer.send(t, clientMessage{Type: intentSeat, Name: "Anna"})
	observer.state(t)

	fat := srv.dial(t, roomID)
	fat.state(t)
	observer.state(t)

	// Comfortably above the limit, and still a well-formed intent, so that what is
	// being tested is the size and nothing else.
	huge := clientMessage{Type: intentSeat, Name: strings.Repeat("a", maxIncomingMessageBytes*2)}
	body, err := json.Marshal(huge)
	if err != nil {
		t.Fatalf("marshalling: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = fat.conn.Write(ctx, websocket.MessageText, body)

	// A message above the read limit ends that connection, and a connection leaving
	// is a change everybody sees — so the observer does get a snapshot. What matters
	// is what it contains: the message was never decoded, so nobody was seated under
	// that thousand-character name and the room is exactly as it was.
	// The oversized message ends the connection that sent it, and a connection
	// leaving is a real change, so the observer does get a snapshot for it. What
	// matters is what that snapshot contains: the message was never decoded, so
	// nobody was seated under a thousand-character name.
	if _, err := fat.readWithin(3 * time.Second); err == nil {
		t.Error("a message above the read limit did not end the connection that sent it")
	}

	msg := observer.state(t)
	if len(msg.Room.Participants) != 1 || msg.Room.Participants[0].Name != "Anna" {
		t.Errorf("an oversized message changed the room: %+v", msg.Room.Participants)
	}
}

// --- helpers ---------------------------------------------------------------

// readWithin reads one message, giving up after a bounded wait.
func (c *client) readWithin(d time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	_, body, err := c.conn.Read(ctx)
	return body, err
}

// assertNothingArrives fails if a connection receives anything in a short window,
// which is how "nobody else at the table noticed" is checked.
func assertNothingArrives(t *testing.T, c *client) {
	t.Helper()
	body, err := c.readWithin(300 * time.Millisecond)
	if err == nil {
		t.Errorf("a connection received %s when nothing should have reached it", body)
		return
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("reading returned %v, want a timeout meaning nothing arrived", err)
	}
}
