package transport

import (
	"context"
	"crypto/rand"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"de.thomasnegele.planningpoker/internal/hub"
	"github.com/coder/websocket"
)

// fastTiming is short enough that several heartbeat rounds pass in well under a
// second, and still long enough not to be flaky on a loaded machine.
const (
	fastHeartbeat = 50 * time.Millisecond
	fastDeadline  = 150 * time.Millisecond
)

// dialSilent opens a connection and never reads from it.
//
// That is what a client whose network has gone looks like from the server's side: the
// socket is still there, nothing arrives, and nothing fails. Because this library
// processes control frames while reading, a client that never reads also never
// answers a ping — so one connection stands in for both the client that stopped
// reading and the one that died silently.
func dialSilent(t *testing.T, srv *testServer, roomID string) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL(srv.URL, roomID), nil)
	if err != nil {
		t.Fatalf("dialling: %v", err)
	}
	t.Cleanup(func() { _ = conn.CloseNow() })
	return conn
}

func TestAConnectionThatStopsAnsweringIsReleased(t *testing.T) {
	srv := newTestServerTimed(t, testLimits, testRate, fastHeartbeat, fastDeadline)
	roomID := srv.createGame(t)

	// Somebody who stays, so the room survives to be observed.
	watcher := srv.dial(t, roomID)
	watcher.state(t)
	watcher.seat(t, "Anna")

	// And somebody who takes a seat and then goes silent.
	silent := dialSilent(t, srv, roomID)
	seatSilently(t, silent)

	// The watcher sees them arrive and sit down.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if len(watcher.state(t).Room.Participants) == 2 {
			break
		}
	}

	// Now nothing is sent to them and nothing is read by them. Without a heartbeat
	// this connection would stay open in the server's eyes for as long as the process
	// lived, holding a seat and keeping the room from ever expiring.
	for time.Now().Before(deadline) {
		msg := watcher.state(t)
		away := 0
		for _, p := range msg.Room.Participants {
			if p.Away {
				away++
			}
		}
		if away == 1 {
			// Marked away, not removed: the seat, the name and the vote are all still there.
			if len(msg.Room.Participants) != 2 {
				t.Errorf("the silent participant was removed rather than marked away: %+v",
					msg.Room.Participants)
			}
			return
		}
	}
	t.Error("a connection that stopped answering was never noticed; the participant stayed present")
}

func TestAQuietParticipantIsNeverDisturbed(t *testing.T) {
	// The behaviour the heartbeat must never break, and the reason it is not an
	// inactivity timeout. A browser answers a ping by itself, with nobody at the
	// keyboard — so a table where people argue for an hour without voting is
	// untouched. Here that hour is many heartbeat intervals, driven fast.
	srv := newTestServerTimed(t, testLimits, testRate, fastHeartbeat, fastDeadline)
	roomID := srv.createGame(t)

	quiet := srv.dial(t, roomID)
	quiet.state(t)
	quiet.seat(t, "Thomas")

	// Keep reading, exactly as a browser does, and send nothing at all. The library
	// answers the pings while reading; the person does nothing.
	//
	// The read runs on one long-lived context rather than a short one per attempt.
	// Cancelling a read in this library does not merely abandon that read, it closes
	// the connection — so a loop of short reads would tear down the very connection
	// this test asserts survives, and would blame the server for it.
	const rounds = 10
	readCtx, stopReading := context.WithCancel(context.Background())
	defer stopReading()

	failed := make(chan error, 1)
	go func() {
		for {
			if _, _, err := quiet.conn.Read(readCtx); err != nil {
				select {
				case failed <- err:
				default:
				}
				return
			}
		}
	}()

	select {
	case err := <-failed:
		t.Fatalf("a quiet connection was closed after doing nothing wrong: %v", err)
	case <-time.After(rounds * fastHeartbeat):
	}

	// Judged while that connection is still open. A newly arriving connection is sent
	// the room as it stands, so this snapshot is the state after the silence — and it
	// must show somebody present, not somebody who was quietly given up on.
	observer := srv.dial(t, roomID)
	msg := observer.state(t)

	if len(msg.Room.Participants) != 1 {
		t.Fatalf("after %d heartbeat intervals of silence the table holds %d participants, want 1",
			rounds, len(msg.Room.Participants))
	}
	if msg.Room.Participants[0].Away {
		t.Error("a participant who simply said nothing was marked away; nothing may measure " +
			"how long it has been since somebody last did something")
	}
	if msg.Room.Participants[0].Name != "Thomas" {
		t.Errorf("the quiet participant is shown as %q", msg.Room.Participants[0].Name)
	}
}

func TestASilentConnectionDoesNotPinAGoroutine(t *testing.T) {
	srv := newTestServerTimed(t, testLimits, testRate, fastHeartbeat, fastDeadline)
	roomID := srv.createGame(t)

	before := runtime.NumGoroutine()
	for range 5 {
		conn := dialSilent(t, srv, roomID)
		seatSilently(t, conn)
	}

	// Each of those connections is served by two goroutines that would otherwise
	// never return: one blocked in a read that will never complete, one blocked in a
	// write to a client that is not taking anything.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if runtime.NumGoroutine() <= before+4 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Errorf("goroutines went from %d to %d and stayed there; five silent connections were "+
		"never released", before, runtime.NumGoroutine())
}

// --- shutdown --------------------------------------------------------------

func TestShutdownFinishesWithinItsBudgetDespiteASilentClient(t *testing.T) {
	// The failure this replaces: the budget went only to server.Shutdown, and the
	// wait on sockets afterwards had no deadline at all. A client that had stopped
	// reading held the process open until the container runtime killed it — while
	// compose.yaml promised orderly cleanup.
	manager := hub.NewManager(hub.SystemClock{}, rand.Reader, testGrace, testLimits)
	rooms := NewRoomHandlers(manager, rand.Reader, testRate)
	// Long enough that the heartbeat cannot be what rescues this test: the shutdown
	// path itself has to be what bounds the wait.
	rooms.heartbeat = time.Hour
	rooms.deadline = time.Hour

	srv := httptest.NewServer(NewRouter(roomOptions(rooms)))
	inner := &testServer{Server: srv, manager: manager, rooms: rooms}

	roomID := inner.createGame(t)
	ordinary := inner.dial(t, roomID)
	ordinary.state(t)
	silent := dialSilent(t, inner, roomID)
	seatSilently(t, silent)

	const budget = 300 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), budget)
	defer cancel()

	started := time.Now()
	rooms.StopAccepting()
	_ = srv.Config.Shutdown(ctx)
	manager.Close()
	if !waitWithinTest(ctx, rooms.Wait) {
		rooms.CloseNow()
		_ = srv.Config.Close()
	}
	took := time.Since(started)

	// Generously bounded, but far below the five seconds the WebSocket library spends
	// on a closing handshake nobody answers — which is what used to be paid here.
	if took > 10*budget {
		t.Errorf("shutdown took %s against a budget of %s; a silent client delayed it", took, budget)
	}
	srv.Close()
}

func TestAConnectionArrivingDuringShutdownIsRefused(t *testing.T) {
	srv := newTestServer(t)
	before := srv.manager.Len()

	srv.rooms.StopAccepting()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL(srv.URL, "arrived-too-late"), nil)
	if err == nil {
		// The handshake may still succeed; what must not happen is being served.
		defer func() { _ = conn.CloseNow() }()
		readCtx, done := context.WithTimeout(context.Background(), 2*time.Second)
		_, _, readErr := conn.Read(readCtx)
		done()
		if readErr == nil {
			t.Error("a connection arriving after shutdown had begun was served a snapshot")
		}
	}

	if got := srv.manager.Len(); got != before {
		t.Errorf("a connection arriving during shutdown created a room (%d, was %d)", got, before)
	}
}

// --- helpers ---------------------------------------------------------------

// seatSilently takes a seat without reading anything back, which is what a client
// that has gone quiet would have done before it went quiet.
func seatSilently(t *testing.T, conn *websocket.Conn) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := conn.Write(ctx, websocket.MessageText, []byte(`{"type":"seat","name":"Silent"}`)); err != nil {
		t.Fatalf("seating silently: %v", err)
	}
}

// waitWithinTest mirrors the production helper, so the test exercises the same shape
// of shutdown that main.go performs.
func waitWithinTest(ctx context.Context, wait func()) bool {
	done := make(chan struct{})
	go func() {
		wait()
		close(done)
	}()
	select {
	case <-done:
		return true
	case <-ctx.Done():
		return false
	}
}
