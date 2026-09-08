package transport

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"de.thomasnegele.planningpoker/internal/game"
	"de.thomasnegele.planningpoker/internal/hub"
	"github.com/coder/websocket"
)

// seatCookiePrefix namespaces seat cookies by room, independently of browser path
// scoping.
const seatCookiePrefix = "pp_seat_"

// seatCookieName is the cookie name for one particular room.
func seatCookieName(roomID game.RoomID) string { return seatCookiePrefix + string(roomID) }

// seatTokenBytes gives seat credentials 128 bits of randomness.
const seatTokenBytes = 16

// closeInvalidRoomID identifies an invalid room name to the browser after upgrade.
const closeInvalidRoomID = 4400

// RoomHandlers creates rooms over HTTP and serves their WebSocket connections.
type RoomHandlers struct {
	manager *hub.Manager
	random  io.Reader

	// Each connection receives its own rate allowance.
	rate RateLimit

	// Injectable clock for rate-limit tests.
	now func() time.Time

	// Tests can shorten the production heartbeat and I/O deadlines.
	heartbeat time.Duration
	deadline  time.Duration

	// sockets tracks handlers that reached a room. HTTP Shutdown ignores hijacked
	// sockets.
	sockets sync.WaitGroup

	// mu guards live and stopping, and nothing else.
	mu sync.Mutex

	// live tracks upgraded sockets so shutdown can request their closure.
	live map[*websocket.Conn]struct{}

	// stopping prevents newly upgraded sockets from entering the live registry.
	stopping bool
}

// Transport defaults are independent of deployment capacity.
const (
	// defaultHeartbeat pings each connection every thirty seconds. Browsers answer
	// automatically; no game action is required to keep a meeting alive.
	defaultHeartbeat = 30 * time.Second

	// defaultDeadline bounds a ping response or outgoing state write to ten seconds.
	// Unresponsive connections are released without imposing a user inactivity timeout.
	defaultDeadline = 10 * time.Second
)

// NewRoomHandlers wires the handlers to a manager. random supplies seat tokens, and
// rate bounds how fast any one connection may send.
func NewRoomHandlers(manager *hub.Manager, random io.Reader, rate RateLimit) *RoomHandlers {
	return &RoomHandlers{
		manager:   manager,
		random:    random,
		rate:      rate,
		now:       time.Now,
		heartbeat: defaultHeartbeat,
		deadline:  defaultDeadline,
		live:      make(map[*websocket.Conn]struct{}),
	}
}

// register records an open socket, reporting false if shutdown has already begun —
// in which case the caller must not serve it.
func (h *RoomHandlers) register(conn *websocket.Conn) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.stopping {
		return false
	}
	h.live[conn] = struct{}{}
	return true
}

// unregister forgets a socket whose handler has finished.
func (h *RoomHandlers) unregister(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.live, conn)
}

// StopAccepting prevents further socket registration during shutdown.
func (h *RoomHandlers) StopAccepting() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.stopping = true
}

// CloseNow starts closing all registered sockets without waiting. The library
// can block CloseNow behind an in-flight Close, so closures run concurrently;
// process exit releases any sockets still open after the shutdown budget.
func (h *RoomHandlers) CloseNow() {
	h.mu.Lock()
	conns := make([]*websocket.Conn, 0, len(h.live))
	for conn := range h.live {
		conns = append(conns, conn)
	}
	h.mu.Unlock()

	for _, conn := range conns {
		go func(conn *websocket.Conn) { _ = conn.CloseNow() }(conn)
	}
}

// closeWithin bounds the caller's wait for the library's context-free Close.
// On timeout, the closing goroutine remains until the library finishes or the
// process exits. See the archived shutdown design for the measured limitation.
func closeWithin(conn *websocket.Conn, code websocket.StatusCode, reason string, limit time.Duration) {
	done := make(chan struct{})
	go func() {
		_ = conn.Close(code, reason)
		close(done)
	}()

	timer := time.NewTimer(limit)
	defer timer.Stop()

	select {
	case <-done:
	case <-timer.C:
	}
}

// Wait blocks until handlers that reached a room have finished.
func (h *RoomHandlers) Wait() { h.sockets.Wait() }

// createGameResponse supplies the ID; the browser builds the public invitation URL.
type createGameResponse struct {
	RoomID string `json:"roomId"`
}

// CreateGame creates a room without seating anyone or assigning a host role.
func (h *RoomHandlers) CreateGame(w http.ResponseWriter, r *http.Request) {
	// Reject cross-origin browser requests before allocating a room.
	if !isSameOrigin(r) {
		http.Error(w, "games may only be started from this site", http.StatusForbidden)
		return
	}

	room, err := h.manager.Create()
	if err != nil {
		if errors.Is(err, hub.ErrAtCapacity) {
			// A distinct status lets the frontend report capacity separately from
			// server errors.
			http.Error(w, "the server is holding as many games as it can", http.StatusServiceUnavailable)
			return
		}
		log.Printf("creating a room: %v", err)
		http.Error(w, "could not create a game", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(createGameResponse{RoomID: string(room.ID())}); err != nil {
		log.Printf("writing the create-game response: %v", err)
	}
}

// Socket upgrades a request before ensuring its room exists. Seat cookies must
// be sent in the HTTP handshake; they are currently issued before full upgrade
// validation and capacity checks.
func (h *RoomHandlers) Socket(w http.ResponseWriter, r *http.Request) {
	roomID := game.RoomID(r.PathValue("roomID"))

	// These preliminary checks do not access the manager or issue cookies.
	if !isWebSocketUpgrade(r) || !isSameOrigin(r) {

		http.Error(w, "this endpoint serves WebSocket connections from this site only",
			http.StatusUpgradeRequired)
		return
	}

	// Invalid IDs still receive a socket so the browser can read a specific close code.
	validID := game.ValidRoomID(roomID)

	var token string
	if validID {
		existing, err := r.Cookie(seatCookieName(roomID))
		if err == nil && existing.Value != "" {
			token = existing.Value
		} else {
			token, err = newSeatToken(h.random)
			if err != nil {
				log.Printf("generating a seat token: %v", err)
				http.Error(w, "could not start a session", http.StatusInternalServerError)
				return
			}
			http.SetCookie(w, seatCookie(roomID, token, isSecureRequest(r)))
		}
	}

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		// Accept performs the complete upgrade validation and writes any refusal
		// response.
		log.Printf("websocket upgrade refused: %v", err)
		return
	}

	// Set the protocol size bound before reading any messages.
	conn.SetReadLimit(maxIncomingMessageBytes)

	// Reject registration during shutdown before creating a room.
	if !h.register(conn) {
		_ = conn.Close(websocket.StatusGoingAway, "the server is shutting down")
		return
	}
	defer h.unregister(conn)

	if !validID {
		h.sendInvalidRoomID(r.Context(), conn)
		return
	}

	// Opening a valid link recreates an expired room at the same address.
	room, err := h.manager.EnsureRoom(roomID)
	if err != nil {
		if !errors.Is(err, hub.ErrAtCapacity) {
			log.Printf("reaching room %q: %v", roomID, err)
		}
		h.sendRefusal(r.Context(), conn, err)
		return
	}

	h.sockets.Add(1)
	defer h.sockets.Done()

	h.serve(r.Context(), conn, room, hub.NewConn(token))
}

// maxIncomingMessageBytes bounds JSON intents to one kilobyte, with room for
// a name and its JSON escaping. Oversized messages are not decoded.
const maxIncomingMessageBytes = 1024

// isWebSocketUpgrade performs preliminary checks before cookie issuance.
// websocket.Accept remains responsible for complete handshake validation.
func isWebSocketUpgrade(r *http.Request) bool {
	return r.Method == http.MethodGet &&
		headerHasToken(r.Header, "Connection", "upgrade") &&
		strings.EqualFold(r.Header.Get("Upgrade"), "websocket") &&
		r.Header.Get("Sec-WebSocket-Key") != ""
}

// headerHasToken matches tokens in comma-separated, possibly repeated headers.
func headerHasToken(header http.Header, name, token string) bool {
	for _, value := range header.Values(name) {
		for _, part := range strings.Split(value, ",") {
			if strings.EqualFold(strings.TrimSpace(part), token) {
				return true
			}
		}
	}
	return false
}

// isSameOrigin compares Origin and Host when Origin is present. Requests without
// Origin are allowed, matching websocket.Accept; this is not authentication.
func isSameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}

	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return strings.EqualFold(parsed.Host, r.Host)
}

// sendInvalidRoomID sends the refusal and a browser-readable application close code.
func (h *RoomHandlers) sendInvalidRoomID(ctx context.Context, conn *websocket.Conn) {
	body, err := json.Marshal(errorMessage{
		Type:    messageError,
		Code:    codeInvalidRoomID,
		Message: "not a usable room name",
	})
	if err == nil {
		_ = conn.Write(ctx, websocket.MessageText, body)
	}
	_ = conn.Close(closeInvalidRoomID, "invalid room identifier")
}

// sendRefusal reports a post-upgrade failure and closes the socket.
func (h *RoomHandlers) sendRefusal(ctx context.Context, conn *websocket.Conn, cause error) {
	code, known := refusalCode(cause)
	if !known {
		code = codeServerError
	}

	body, err := json.Marshal(errorMessage{Type: messageError, Code: code, Message: cause.Error()})
	if err == nil {
		_ = conn.Write(ctx, websocket.MessageText, body)
	}
	_ = conn.Close(websocket.StatusTryAgainLater, code)
}

// serve runs one connection: a goroutine writing updates to the socket, and this
// goroutine reading intents from it.
func (h *RoomHandlers) serve(ctx context.Context, conn *websocket.Conn, room *hub.Room, hubConn *hub.Conn) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	switch room.Attach(hubConn) {
	case hub.AttachRoomStopped:
		// The room may expire between lookup and attachment. A retry can recreate it.
		_ = conn.Close(websocket.StatusTryAgainLater, "the room was closing; reconnect")
		return
	case hub.AttachAtCapacity:
		// Keep the room connection ceiling distinct from the process room ceiling.
		h.sendRefusal(ctx, conn, hub.ErrRoomAtCapacity)
		return
	}

	var writing sync.WaitGroup
	writing.Add(1)
	go func() {
		defer writing.Done()
		defer cancel()
		writeUpdates(ctx, conn, hubConn, h.heartbeat, h.deadline)
	}()

	// A per-connection allowance lets participants act independently.
	readIntents(ctx, conn, room, hubConn, newBucket(h.rate, h.now))

	cancel()
	room.Detach(hubConn)
	hubConn.Close()
	writing.Wait()
	_ = conn.CloseNow()
}

// writeUpdates sends snapshots and heartbeats from one goroutine. Writes and
// pings have individual deadlines; room closure starts a bounded closing handshake.
func writeUpdates(
	ctx context.Context,
	conn *websocket.Conn,
	hubConn *hub.Conn,
	heartbeat, deadline time.Duration,
) {
	ticker := time.NewTicker(heartbeat)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-hubConn.Closed():
			// Bound the wait for clients that do not answer the closing handshake.
			closeWithin(conn, websocket.StatusGoingAway, "the room closed this connection", deadline)
			return
		case <-ticker.C:
			// Ping detects a dead network path even when no game actions arrive.
			pingCtx, done := context.WithTimeout(ctx, deadline)
			err := conn.Ping(pingCtx)
			done()
			if err != nil {
				return
			}
		case update := <-hubConn.Updates():
			body, err := encodeUpdate(update)
			if err != nil {
				log.Printf("encoding an update: %v", err)
				return
			}

			// Bound each write so a non-reading client cannot hold this loop
			// indefinitely.
			writeCtx, done := context.WithTimeout(ctx, deadline)
			err = conn.Write(writeCtx, websocket.MessageText, body)
			done()
			if err != nil {
				return
			}
		}
	}
}

// readIntents reads what the browser sends and turns it into commands for the room.
func readIntents(ctx context.Context, conn *websocket.Conn, room *hub.Room, hubConn *hub.Conn, allowance *bucket) {
	// Count consecutive refusals; any accepted message resets the run.
	refusedInARow := 0

	for {
		_, raw, err := conn.Read(ctx)
		if err != nil {
			// Closed sockets and messages exceeding the read limit end this connection.
			return
		}

		if !allowance.allow() {
			// Rate-limit before decoding or dispatching to the room.
			refusedInARow++
			if refusedInARow > allowance.burstSize() {
				// Continued flooding closes the socket; normal detach handling marks
				// its seat away.
				_ = conn.Close(websocket.StatusPolicyViolation, codeTooFast)
				return
			}
			if !hubConn.Refuse(errTooFast) {
				return
			}
			continue
		}
		refusedInARow = 0

		msg, err := decodeClientMessage(raw)
		if err != nil {
			// Malformed intents are refused without reaching the room.
			if !hubConn.Refuse(err) {
				return
			}
			continue
		}

		if !dispatch(room, hubConn, msg) {
			return
		}
	}
}

// dispatch sends one intent to the room. It reports false when the room has gone,
// which ends the connection.
func dispatch(room *hub.Room, hubConn *hub.Conn, msg clientMessage) bool {
	switch msg.Type {
	case intentSeat:
		return room.Seat(hubConn, msg.Name)
	case intentVote:
		return room.Vote(hubConn, game.Card(msg.Card))
	case intentReveal:
		return room.Reveal(hubConn)
	case intentNewRound:
		return room.NewRound(hubConn)
	case intentRename:
		return room.Rename(hubConn, msg.Name)
	default:

		return true
	}
}

// newSeatToken generates a private seat credential, distinct from the public
// participant ID. The configured entropy source must be secure in production.
func newSeatToken(random io.Reader) (string, error) {
	buf := make([]byte, seatTokenBytes)
	if _, err := io.ReadFull(random, buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// seatCookie builds the room's session credential cookie.
func seatCookie(roomID game.RoomID, token string, secure bool) *http.Cookie {
	return &http.Cookie{
		Name:  seatCookieName(roomID),
		Value: token,

		// Scope the credential to this room's WebSocket endpoint.
		Path: "/ws/" + string(roomID),

		// Keep the credential inaccessible to JavaScript.
		HttpOnly: true,

		// Restrict cross-site cookie use; WebSocket origin validation is also required.
		SameSite: http.SameSiteLaxMode,

		Secure: secure,

		// No Expires or MaxAge: session lifetime is browser-controlled and may survive
		// a browser restart through session restoration. The server does not check
		// token age.
	}
}

// isSecureRequest accepts direct TLS or X-Forwarded-Proto: https. A terminating
// proxy must overwrite this header to reflect the client connection.
func isSecureRequest(r *http.Request) bool {
	return r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
}

// DefaultRandom is the source of seat tokens in production.
var DefaultRandom = rand.Reader
