package transport

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"de.thomasnegele.planningpoker/internal/game"
	"de.thomasnegele.planningpoker/internal/hub"
	"de.thomasnegele.planningpoker/internal/legal"
	"github.com/coder/websocket"
)

const testGrace = time.Hour

// testLimits and testRate are deliberately far above anything the ordinary tests
// reach, so that a test about seating or reconnecting fails for its own reason and
// never because it quietly ran into a ceiling. The tests that are about the ceilings
// build a server with their own.
var (
	testLimits = hub.Limits{Rooms: 1000, ConnectionsPerRoom: 1000, ParticipantsPerRoom: 1000}
	testRate   = RateLimit{PerSecond: 1000, Burst: 1000}
)

// testServer is a running application: a hub, the real handlers, and an HTTP server
// in front of them.
type testServer struct {
	*httptest.Server
	manager *hub.Manager
	rooms   *RoomHandlers
}

// roomOptions wires the application the way a default installation runs: these
// tests are about rooms, so legal notices are switched off, but their routes are
// still registered, exactly as they are in a process nobody configured them for.
func roomOptions(rooms *RoomHandlers) Options {
	return Options{
		Assets:      stub("assets"),
		Socket:      http.HandlerFunc(rooms.Socket),
		CreateGame:  http.HandlerFunc(rooms.CreateGame),
		LegalPages:  legal.Pages(nil),
		LegalStatus: legal.Status(nil),
	}
}

func newTestServer(t *testing.T) *testServer {
	t.Helper()
	return newTestServerWith(t, testLimits, testRate)
}

// newTestServerWith is the same application under chosen ceilings, for the tests
// that are about what happens when one is reached.
func newTestServerWith(t *testing.T, limits hub.Limits, rate RateLimit) *testServer {
	t.Helper()
	return newTestServerTimed(t, limits, rate, 0, 0)
}

// newTestServerTimed additionally shortens the heartbeat and the network deadline,
// for the tests that are about liveness. A zero for either keeps the production
// constant. Driving these in milliseconds is what makes those tests fast: proving a
// thirty-second heartbeat by waiting thirty seconds is a test nobody runs.
func newTestServerTimed(t *testing.T, limits hub.Limits, rate RateLimit, heartbeat, deadline time.Duration) *testServer {
	t.Helper()

	manager := hub.NewManager(hub.SystemClock{}, rand.Reader, testGrace, limits)
	rooms := NewRoomHandlers(manager, rand.Reader, rate)
	if heartbeat > 0 {
		rooms.heartbeat = heartbeat
	}
	if deadline > 0 {
		rooms.deadline = deadline
	}
	srv := httptest.NewServer(NewRouter(roomOptions(rooms)))

	t.Cleanup(func() {
		srv.Close()
		manager.Close()
		rooms.Wait()
	})

	return &testServer{Server: srv, manager: manager, rooms: rooms}
}

func (s *testServer) createGame(t *testing.T) string {
	t.Helper()
	res, err := http.Post(s.URL+"/api/games", "", nil)
	if err != nil {
		t.Fatalf("creating a game: %v", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("creating a game: status %d", res.StatusCode)
	}
	var body createGameResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decoding the create-game response: %v", err)
	}
	if body.RoomID == "" {
		t.Fatal("creating a game returned an empty room identifier")
	}
	return body.RoomID
}

func (s *testServer) createGameWithDeck(t *testing.T, deck string) string {
	t.Helper()
	body, err := json.Marshal(createGameRequest{Deck: deck})
	if err != nil {
		t.Fatalf("encoding game settings: %v", err)
	}
	res, err := http.Post(s.URL+"/api/games", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("creating a game: %v", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("creating a %q game: status %d", deck, res.StatusCode)
	}
	var response createGameResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		t.Fatalf("decoding the create-game response: %v", err)
	}
	return response.RoomID
}

// client is one browser: a socket plus the cookies it was given.
type client struct {
	conn    *websocket.Conn
	cookies []*http.Cookie
}

// dial opens a connection to a room, optionally presenting cookies it already holds.
func (s *testServer) dial(t *testing.T, roomID string, cookies ...*http.Cookie) *client {
	t.Helper()

	header := http.Header{}
	for _, c := range cookies {
		header.Add("Cookie", c.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, res, err := websocket.Dial(ctx, wsURL(s.URL, roomID), &websocket.DialOptions{HTTPHeader: header})
	if err != nil {
		t.Fatalf("dialling room %q: %v", roomID, err)
	}

	got := cookies
	if res != nil {
		if issued := res.Cookies(); len(issued) > 0 {
			got = issued
		}
	}
	t.Cleanup(func() { _ = conn.CloseNow() })
	return &client{conn: conn, cookies: got}
}

func wsURL(base, roomID string) string {
	return "ws" + strings.TrimPrefix(base, "http") + "/ws/" + roomID
}

// send writes one intent.
func (c *client) send(t *testing.T, msg clientMessage) {
	t.Helper()
	body, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshalling %+v: %v", msg, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.conn.Write(ctx, websocket.MessageText, body); err != nil {
		t.Fatalf("writing %+v: %v", msg, err)
	}
}

// raw reads the next message as the bytes that actually crossed the socket.
func (c *client) raw(t *testing.T) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, body, err := c.conn.Read(ctx)
	if err != nil {
		t.Fatalf("reading: %v", err)
	}
	return body
}

// state reads the next message, requiring it to be a snapshot.
func (c *client) state(t *testing.T) stateMessage {
	t.Helper()
	body := c.raw(t)
	var msg stateMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		t.Fatalf("decoding %s: %v", body, err)
	}
	if msg.Type != messageState {
		t.Fatalf("expected a snapshot, got %s", body)
	}
	return msg
}

// refusal reads the next message, requiring it to be a refusal.
func (c *client) refusal(t *testing.T) errorMessage {
	t.Helper()
	body := c.raw(t)
	var msg errorMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		t.Fatalf("decoding %s: %v", body, err)
	}
	if msg.Type != messageError {
		t.Fatalf("expected a refusal, got %s", body)
	}
	return msg
}

// thrown reads the next message, requiring a transient throw event.
func (c *client) thrown(t *testing.T) thrownMessage {
	t.Helper()
	body := c.raw(t)
	var msg thrownMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		t.Fatalf("decoding %s: %v", body, err)
	}
	if msg.Type != messageThrown {
		t.Fatalf("expected a throw event, got %s", body)
	}
	return msg
}

// seat takes a seat and returns the snapshot that follows.
func (c *client) seat(t *testing.T, name string) stateMessage {
	t.Helper()
	c.send(t, clientMessage{Type: intentSeat, Name: name})
	return c.state(t)
}

func TestCreatingGamesProducesIndependentRooms(t *testing.T) {
	srv := newTestServer(t)

	first := srv.createGame(t)
	second := srv.createGame(t)
	if first == second {
		t.Fatal("two games produced the same room identifier")
	}

	// A game seats nobody: the person who created it holds no privilege and no chair.
	a := srv.dial(t, first)
	if got := len(a.state(t).Room.Participants); got != 0 {
		t.Errorf("a freshly created game shows %d participants, want 0", got)
	}
	a.seat(t, "Thomas")
	a.send(t, clientMessage{Type: intentVote, Card: string(game.CardM)})
	a.state(t)

	// A vote in one room is not visible in the other.
	b := srv.dial(t, second)
	if got := len(b.state(t).Room.Participants); got != 0 {
		t.Errorf("the second room shows %d participants, want 0", got)
	}
}

func TestCreatingAGameMaySelectFibonacci(t *testing.T) {
	srv := newTestServer(t)
	roomID := srv.createGameWithDeck(t, game.FibonacciDeckName)

	c := srv.dial(t, roomID)
	view := c.state(t)
	if view.Room.Deck.Name != game.FibonacciDeckName {
		t.Errorf("created room deck = %q, want Fibonacci", view.Room.Deck.Name)
	}
	wantCards := []string{"0", "½", "1", "2", "3", "5", "8", "13", "21", "?", "☕"}
	if !slices.Equal(view.Room.Deck.Cards, wantCards) {
		t.Errorf("Fibonacci cards = %v, want %v", view.Room.Deck.Cards, wantCards)
	}
}

func TestCreatingAGameWithoutABodyKeepsTShirtDefault(t *testing.T) {
	srv := newTestServer(t)
	roomID := srv.createGame(t)

	c := srv.dial(t, roomID)
	if got := c.state(t).Room.Deck.Name; got != game.TShirtDeckName {
		t.Errorf("default deck = %q, want T-shirt", got)
	}
}

func TestUnknownCreationDeckIsAClientErrorAndCreatesNothing(t *testing.T) {
	srv := newTestServer(t)
	before := srv.manager.Len()
	body := bytes.NewBufferString(`{"deck":"custom"}`)

	res, err := http.Post(srv.URL+"/api/games", "application/json", body)
	if err != nil {
		t.Fatalf("creating a game: %v", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", res.StatusCode, http.StatusBadRequest)
	}
	if got := srv.manager.Len(); got != before {
		t.Errorf("manager grew from %d rooms to %d after invalid deck", before, got)
	}
}

func TestConnectingToAnUnusedNameCreatesTheRoom(t *testing.T) {
	// The whole of what makes an old link work and what puts an interrupted group
	// back together: the connection creates the room it was looking for.
	srv := newTestServer(t)

	before := srv.manager.Len()
	c := srv.dial(t, "team-alpha")
	view := c.state(t)

	if view.Room.ID != "team-alpha" {
		t.Errorf("room id = %q, want the name that was asked for", view.Room.ID)
	}
	if len(view.Room.Participants) != 0 {
		t.Errorf("a freshly created room has %d participants, want 0", len(view.Room.Participants))
	}
	if view.Room.Deck.Name != game.TShirtDeckName {
		t.Errorf("implicitly created room deck = %q, want T-shirt", view.Room.Deck.Name)
	}
	if srv.manager.Len() != before+1 {
		t.Errorf("rooms went from %d to %d, want one more", before, srv.manager.Len())
	}
}

func TestConnectingWithAnUnusableNameIsRefused(t *testing.T) {
	srv := newTestServer(t)

	for name, id := range map[string]string{
		"too short":           "abc",
		"a dot":               "team.alpha",
		"an exclamation mark": "team%21",
		"far too long":        strings.Repeat("a", 65),
	} {
		before := srv.manager.Len()
		c := srv.dial(t, id)

		msg := c.refusal(t)
		if msg.Code != codeInvalidRoomID {
			t.Errorf("%s: code = %q, want %q", name, msg.Code, codeInvalidRoomID)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, _, err := c.conn.Read(ctx)
		cancel()
		if got := websocket.CloseStatus(err); got != closeInvalidRoomID {
			t.Errorf("%s: close status = %d, want %d", name, got, closeInvalidRoomID)
		}
		if srv.manager.Len() != before {
			t.Errorf("%s: a room was created for an unusable name", name)
		}
	}
}

func TestTwoSpellingsAreTwoRooms(t *testing.T) {
	// Nothing is folded: the address bar is the address.
	srv := newTestServer(t)

	upper := srv.dial(t, "Team-Alpha")
	upper.state(t)
	upper.seat(t, "Anna")

	lower := srv.dial(t, "team-alpha")
	view := lower.state(t)
	if len(view.Room.Participants) != 0 {
		t.Errorf("the differently spelled room already has %d participants, want 0",
			len(view.Room.Participants))
	}
}

func TestSimultaneousArrivalsConverge(t *testing.T) {
	// The ordinary case after a restart, not an edge case: several people opening the
	// same URL in the same instant must end up in one room.
	srv := newTestServer(t)

	const arrivals = 8
	var wg sync.WaitGroup
	for range arrivals {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = srv.dial(t, "standup-room")
		}()
	}
	wg.Wait()

	if got := srv.manager.Len(); got != 1 {
		t.Errorf("%d rooms exist at that name, want 1", got)
	}
}

func TestFetchingThePageCreatesNoRoom(t *testing.T) {
	// Link-preview bots in chat clients fetch the HTML of any URL that is pasted.
	// They do not run scripts and do not open sockets, so unfurling an old link must
	// not bring a room into existence.
	srv := newTestServer(t)

	before := srv.manager.Len()
	res, err := http.Get(srv.URL + "/g/some-room-name")
	if err != nil {
		t.Fatalf("fetching the page: %v", err)
	}
	_ = res.Body.Close()

	if srv.manager.Len() != before {
		t.Errorf("fetching a room URL created a room; a preview bot must not be able to")
	}
}

func TestASeatCookieIsIssuedForAFreshlyCreatedRoom(t *testing.T) {
	srv := newTestServer(t)
	c := srv.dial(t, "brand-new-room")

	if len(c.cookies) != 1 {
		t.Fatalf("the handshake issued %d cookies, want 1", len(c.cookies))
	}
	if want := seatCookieName("brand-new-room"); c.cookies[0].Name != want {
		t.Errorf("cookie name = %q, want %q", c.cookies[0].Name, want)
	}
}

func TestSnapshotArrivesFirstAndCarriesTheWholeRoom(t *testing.T) {
	srv := newTestServer(t)
	roomID := srv.createGame(t)

	seated := srv.dial(t, roomID)
	seated.state(t)
	seated.seat(t, "Thomas")
	seated.send(t, clientMessage{Type: intentVote, Card: string(game.CardM)})
	seated.state(t)

	late := srv.dial(t, roomID)
	first := late.state(t)

	if first.Room.ID != roomID {
		t.Errorf("snapshot names room %q, want %q", first.Room.ID, roomID)
	}
	if len(first.Room.Participants) != 1 {
		t.Fatalf("the first message shows %d participants, want 1", len(first.Room.Participants))
	}
	if !first.Room.Participants[0].Voted {
		t.Error("the first message does not show that the seated participant has voted")
	}
	if first.Room.Deck.Name != game.TShirtDeckName || len(first.Room.Deck.Cards) != 7 {
		t.Errorf("deck = %+v, want the seven-card t-shirt deck", first.Room.Deck)
	}
	// The scale must survive the crossing intact. It is what tells the client which
	// cards belong on an ordered axis, and a client that received it empty would have
	// to guess — which is the guessing this field exists to remove.
	wantScale := []string{"XS", "S", "M", "L", "XL"}
	if !slices.Equal(first.Room.Deck.Scale, wantScale) {
		t.Errorf("deck scale = %v, want %v", first.Room.Deck.Scale, wantScale)
	}
	for _, card := range first.Room.Deck.Scale {
		if !slices.Contains(first.Room.Deck.Cards, card) {
			t.Errorf("scale names %q, which the deck does not offer", card)
		}
	}
	if first.You != "" {
		t.Errorf("you = %q, want empty for a connection that has not taken a seat", first.You)
	}
}

func TestEveryChangeReachesEveryConnection(t *testing.T) {
	srv := newTestServer(t)
	roomID := srv.createGame(t)

	clients := []*client{srv.dial(t, roomID), srv.dial(t, roomID), srv.dial(t, roomID)}

	// Every attach broadcasts to everyone already connected, so the connections that
	// arrived earlier have several snapshots waiting. Read exactly those, rather than
	// draining by timeout: a read whose context expires closes the connection.
	for i, c := range clients {
		for range len(clients) - i {
			c.state(t)
		}
	}

	clients[0].send(t, clientMessage{Type: intentSeat, Name: "Thomas"})

	for i, c := range clients {
		if got := len(c.state(t).Room.Participants); got != 1 {
			t.Errorf("connection %d sees %d participants, want 1", i, got)
		}
	}
}

func TestDeckChangesFollowRoundTimingAndReachEveryParticipant(t *testing.T) {
	srv := newTestServer(t)
	roomID := srv.createGame(t)

	actor := srv.dial(t, roomID)
	actor.state(t)
	actor.seat(t, "Thomas")
	observer := srv.dial(t, roomID)
	observer.state(t)
	actor.state(t)
	observer.seat(t, "Anna")
	actor.state(t)

	actor.send(t, clientMessage{Type: intentVote, Card: string(game.CardM)})
	actor.state(t)
	observer.state(t)
	actor.send(t, clientMessage{Type: intentSetDeck, Deck: game.FibonacciDeckName})
	if got := actor.refusal(t).Code; got != codeDeckLocked {
		t.Errorf("deck change during voting code = %q, want %q", got, codeDeckLocked)
	}

	actor.send(t, clientMessage{Type: intentReveal})
	actor.state(t)
	observer.state(t)
	actor.send(t, clientMessage{Type: intentSetDeck, Deck: game.FibonacciDeckName})
	for i, c := range []*client{actor, observer} {
		view := c.state(t).Room
		if view.Deck.Name != game.TShirtDeckName || view.PendingDeck == nil ||
			view.PendingDeck.Name != game.FibonacciDeckName {
			t.Errorf("client %d active/pending = %+v/%+v", i, view.Deck, view.PendingDeck)
		}
		if view.Results == nil || view.Results.Cards[0].Card != string(game.CardM) {
			t.Errorf("client %d lost revealed T-shirt results: %+v", i, view.Results)
		}
	}

	observer.send(t, clientMessage{Type: intentSetDeck, Deck: game.TShirtDeckName})
	for i, c := range []*client{actor, observer} {
		view := c.state(t).Room
		if view.PendingDeck == nil || view.PendingDeck.Name != game.TShirtDeckName {
			t.Errorf("client %d latest pending deck = %+v, want T-shirt", i, view.PendingDeck)
		}
	}
	actor.send(t, clientMessage{Type: intentSetDeck, Deck: game.FibonacciDeckName})
	actor.state(t)
	observer.state(t)

	actor.send(t, clientMessage{Type: intentNewRound})
	for i, c := range []*client{actor, observer} {
		view := c.state(t).Room
		if view.Deck.Name != game.FibonacciDeckName || view.PendingDeck != nil || view.Revealed {
			t.Errorf("client %d new-round deck state = %+v", i, view)
		}
		if view.Participants[0].Voted || view.Participants[1].Voted {
			t.Errorf("client %d kept votes in the new round: %+v", i, view.Participants)
		}
	}
}

func TestSeatCookieIsScopedAndHidden(t *testing.T) {
	srv := newTestServer(t)
	roomID := srv.createGame(t)

	c := srv.dial(t, roomID)
	if len(c.cookies) != 1 {
		t.Fatalf("the handshake issued %d cookies, want 1", len(c.cookies))
	}
	cookie := c.cookies[0]

	if want := seatCookieName(game.RoomID(roomID)); cookie.Name != want {
		t.Errorf("cookie name = %q, want %q", cookie.Name, want)
	}
	if want := "/ws/" + roomID; cookie.Path != want {
		t.Errorf("cookie path = %q, want %q — the token must not be sent to other rooms", cookie.Path, want)
	}
	if !cookie.HttpOnly {
		t.Error("the seat cookie is readable by scripts; it is the one value that could take over a seat")
	}
	if cookie.Value == "" {
		t.Error("the seat cookie is empty")
	}
}

func TestTheSeatCookieDoesNotOutliveTheBrowser(t *testing.T) {
	// A session cookie: no Max-Age and no Expires, so the browser discards it when it
	// closes. Everything this token exists to survive — a reload, a closed tab, a
	// sleeping laptop, a dropped connection — happens with the browser still running.
	//
	// Asserted on the header the server actually sent rather than on the parsed
	// cookie, because Go's parser reports a missing Max-Age and a Max-Age of zero
	// identically, and those two mean opposite things on the wire: absent is "for this
	// session", while `Max-Age=0` tells the browser to delete the cookie at once.
	srv := newTestServer(t)
	roomID := srv.createGame(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, res, err := websocket.Dial(ctx, wsURL(srv.URL, roomID), nil)
	if err != nil {
		t.Fatalf("dialling: %v", err)
	}
	defer func() { _ = conn.CloseNow() }()

	header := res.Header.Get("Set-Cookie")
	if header == "" {
		t.Fatal("the handshake set no cookie")
	}
	lowered := strings.ToLower(header)
	if strings.Contains(lowered, "max-age") {
		t.Errorf("the seat cookie carries a max age: %s", header)
	}
	if strings.Contains(lowered, "expires") {
		t.Errorf("the seat cookie carries an expiry date: %s", header)
	}
}

func TestTheSeatTokenNeverReachesAnyClient(t *testing.T) {
	// The token is a credential and the participant identifier is public. Were they
	// the same value, anyone at the table could read somebody else's credential out
	// of an ordinary snapshot and take their seat.
	srv := newTestServer(t)
	roomID := srv.createGame(t)

	first := srv.dial(t, roomID)
	first.state(t)
	seated := first.seat(t, "Thomas")

	second := srv.dial(t, roomID)
	view := second.state(t)

	token := first.cookies[0].Value
	if seated.You == token {
		t.Error("the participant identifier is the seat token; anyone at the table could take this seat")
	}

	for _, body := range [][]byte{mustJSON(t, seated), mustJSON(t, view)} {
		if strings.Contains(string(body), token) {
			t.Errorf("a message contains the seat token:\n%s", body)
		}
	}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	body, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshalling: %v", err)
	}
	return body
}

func TestReconnectingReseatsTheSameParticipant(t *testing.T) {
	srv := newTestServer(t)
	roomID := srv.createGame(t)

	first := srv.dial(t, roomID)
	first.state(t)
	seated := first.seat(t, "Thomas")
	participant := seated.You

	first.send(t, clientMessage{Type: intentVote, Card: string(game.CardL)})
	first.state(t)
	_ = first.conn.Close(websocket.StatusNormalClosure, "reload")

	// The same browser comes back with the cookie it was given.
	again := srv.dial(t, roomID, first.cookies...)
	view := again.state(t)

	if len(view.Room.Participants) != 1 {
		t.Fatalf("the table shows %d participants after a reload, want 1", len(view.Room.Participants))
	}
	if view.You != participant {
		t.Errorf("you = %q after reconnecting, want the original %q", view.You, participant)
	}
	if p := view.Room.Participants[0]; p.Name != "Thomas" || !p.Voted || p.Away {
		t.Errorf("participant after reconnecting = %+v, want Thomas, voted, not away", p)
	}
}

func TestATokenFromAnotherRoomIsANewArrival(t *testing.T) {
	srv := newTestServer(t)
	first := srv.createGame(t)
	second := srv.createGame(t)

	a := srv.dial(t, first)
	a.state(t)
	a.seat(t, "Thomas")

	// The same browser in a different room is given its own token and starts fresh.
	b := srv.dial(t, second, a.cookies...)
	view := b.state(t)
	if len(view.Room.Participants) != 0 {
		t.Errorf("the second room shows %d participants, want 0", len(view.Room.Participants))
	}
	if b.cookies[0].Value == a.cookies[0].Value {
		t.Error("both rooms issued the same seat token; neither room may recognise the same browser")
	}
	if b.cookies[0].Name == a.cookies[0].Name || b.cookies[0].Path == a.cookies[0].Path {
		t.Errorf("the two rooms' cookies are not distinct: %q at %q and %q at %q",
			a.cookies[0].Name, a.cookies[0].Path, b.cookies[0].Name, b.cookies[0].Path)
	}
}

func TestTwoTabsAreOneSeat(t *testing.T) {
	srv := newTestServer(t)
	roomID := srv.createGame(t)

	first := srv.dial(t, roomID)
	first.state(t)
	first.seat(t, "Thomas")

	// A second tab presents the same cookie.
	second := srv.dial(t, roomID, first.cookies...)
	view := second.state(t)
	if len(view.Room.Participants) != 1 {
		t.Fatalf("a second tab produced %d participants, want 1", len(view.Room.Participants))
	}
	first.state(t) // the attach broadcast

	// An action in one tab appears in the other.
	second.send(t, clientMessage{Type: intentVote, Card: string(game.CardM)})
	if !first.state(t).Room.Participants[0].Voted {
		t.Error("a vote in the second tab did not appear in the first")
	}
	second.state(t)

	// Closing one connection must not mark the participant away.
	_ = first.conn.Close(websocket.StatusNormalClosure, "closed one tab")
	view = second.state(t)
	if view.Room.Participants[0].Away {
		t.Error("closing one of two connections marked the participant away")
	}
}

func TestRefusalsAreSpecificPrivateAndHarmless(t *testing.T) {
	srv := newTestServer(t)
	roomID := srv.createGame(t)

	actor := srv.dial(t, roomID)
	actor.state(t)
	actor.seat(t, "Thomas")

	observer := srv.dial(t, roomID)
	observer.state(t)
	actor.state(t) // the observer's attach broadcast

	for _, tc := range []struct {
		name string
		msg  clientMessage
		want string
	}{
		{"a card outside the deck", clientMessage{Type: intentVote, Card: "XXL"}, codeCardNotInDeck},
		{"an unknown deck", clientMessage{Type: intentSetDeck, Deck: "custom"}, codeUnknownDeck},
		{"an empty name", clientMessage{Type: intentRename, Name: "  "}, codeNameEmpty},
		{"an over-long name", clientMessage{Type: intentRename, Name: strings.Repeat("a", game.MaxNameLength+1)}, codeNameTooLong},
	} {
		actor.send(t, tc.msg)
		if got := actor.refusal(t).Code; got != tc.want {
			t.Errorf("%s: code = %q, want %q", tc.name, got, tc.want)
		}
	}

	// Voting after the reveal is refused by the server, whatever the interface thinks.
	actor.send(t, clientMessage{Type: intentReveal})
	actor.state(t)
	observer.state(t)
	actor.send(t, clientMessage{Type: intentVote, Card: string(game.CardM)})
	if got := actor.refusal(t).Code; got != codeRoundRevealed {
		t.Errorf("voting after the reveal: code = %q, want %q", got, codeRoundRevealed)
	}

	// None of that reached anybody else, and nothing changed on account of it.
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	if _, body, err := observer.conn.Read(ctx); err == nil {
		t.Errorf("the observer received %s; a refusal must reach only the connection that caused it", body)
	}
}

func TestAnUnseatedConnectionMayNotAct(t *testing.T) {
	srv := newTestServer(t)
	roomID := srv.createGame(t)

	c := srv.dial(t, roomID)
	c.state(t)

	c.send(t, clientMessage{Type: intentVote, Card: string(game.CardM)})
	if got := c.refusal(t).Code; got != codeNotSeated {
		t.Errorf("code = %q, want %q", got, codeNotSeated)
	}

	c.send(t, clientMessage{Type: intentSetDeck, Deck: game.FibonacciDeckName})
	if got := c.refusal(t).Code; got != codeNotSeated {
		t.Errorf("set-deck code = %q, want %q", got, codeNotSeated)
	}
}

func TestAMalformedMessageHarmsNobody(t *testing.T) {
	srv := newTestServer(t)
	roomID := srv.createGame(t)

	actor := srv.dial(t, roomID)
	actor.state(t)
	actor.seat(t, "Thomas")

	observer := srv.dial(t, roomID)
	observer.state(t)
	actor.state(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, nonsense := range []string{`not json at all`, `{"type":"selfDestruct"}`, `{}`, `[1,2,3]`} {
		if err := actor.conn.Write(ctx, websocket.MessageText, []byte(nonsense)); err != nil {
			t.Fatalf("writing %q: %v", nonsense, err)
		}
		if got := actor.refusal(t).Code; got != codeBadMessage {
			t.Errorf("%q: code = %q, want %q", nonsense, got, codeBadMessage)
		}
	}

	// The room is unchanged and nobody else noticed.
	quiet, cancelQuiet := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancelQuiet()
	if _, body, err := observer.conn.Read(quiet); err == nil {
		t.Errorf("the observer received %s after somebody sent nonsense", body)
	}
}

func TestForeignOriginIsStillRefused(t *testing.T) {
	srv := newTestServer(t)
	roomID := srv.createGame(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	header := http.Header{}
	header.Set("Origin", "https://evil.example.com")
	conn, _, err := websocket.Dial(ctx, wsURL(srv.URL, roomID), &websocket.DialOptions{HTTPHeader: header})
	if err == nil {
		_ = conn.CloseNow()
		t.Fatal("a connection from a foreign origin was accepted")
	}
}

func TestShutdownLeavesNoGoroutineBehind(t *testing.T) {
	// The process exiting at all is already evidence — Wait blocks on every socket
	// handler and manager.Close on every room goroutine, so a leak would hang it.
	// This asserts it directly rather than inferring it.
	before := runtime.NumGoroutine()

	manager := hub.NewManager(hub.SystemClock{}, rand.Reader, testGrace, testLimits)
	rooms := NewRoomHandlers(manager, rand.Reader, testRate)
	srv := httptest.NewServer(NewRouter(roomOptions(rooms)))

	inner := &testServer{Server: srv, manager: manager, rooms: rooms}
	for range 3 {
		roomID := inner.createGame(t)
		for range 3 {
			c := inner.dial(t, roomID)
			c.state(t)
		}
	}

	srv.Close()
	manager.Close()
	rooms.Wait()

	// Goroutines unwind asynchronously, so allow a moment before concluding.
	deadline := time.Now().Add(5 * time.Second)
	for runtime.NumGoroutine() > before+2 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if after := runtime.NumGoroutine(); after > before+2 {
		t.Errorf("goroutines went from %d to %d across nine connections in three rooms; "+
			"something did not shut down", before, after)
	}
}
