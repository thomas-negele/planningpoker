// Command planningpoker serves the frontend and room WebSocket API.
package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"de.thomasnegele.planningpoker/internal/hub"
	"de.thomasnegele.planningpoker/internal/legal"
	"de.thomasnegele.planningpoker/internal/transport"
	"de.thomasnegele.planningpoker/internal/webassets"
)

func main() {
	if err := run(); err != nil {

		fmt.Fprintf(os.Stderr, "planningpoker: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := loadConfig(os.Getenv)
	if err != nil {
		return err
	}

	// Read the optional notices before anything else starts. An operator who
	// configured them but left them incomplete finds out here, in a process that
	// has not yet accepted a connection, rather than from a visitor following a
	// broken link.
	notices, err := legal.Load(cfg.LegalDir)
	if err != nil {
		return fmt.Errorf("%s=%q: %w", envLegalDir, cfg.LegalDir, err)
	}
	if notices != nil {
		log.Printf("serving legal notices from %s", cfg.LegalDir)
	}

	manager := hub.NewManager(hub.SystemClock{}, rand.Reader, cfg.RoomGracePeriod, hub.Limits{
		Rooms:               cfg.MaxRooms,
		ConnectionsPerRoom:  cfg.MaxConnectionsPerRoom,
		ParticipantsPerRoom: cfg.MaxParticipantsPerRoom,
	})
	rooms := transport.NewRoomHandlers(manager, transport.DefaultRandom, transport.RateLimit{
		PerSecond: cfg.MessageRate,
		Burst:     cfg.MessageBurst(),
	})

	sweeping, stopSweeping := context.WithCancel(context.Background())
	defer stopSweeping()
	go manager.Run(sweeping)

	router := transport.NewRouter(transport.Options{
		Assets:     webassets.New(),
		Socket:     http.HandlerFunc(rooms.Socket),
		CreateGame: http.HandlerFunc(rooms.CreateGame),

		// Registered whether or not notices exist: when they do not, the subtree
		// answers 404 instead of letting a notice URL reach the application document.
		LegalPages:  legal.Pages(notices),
		LegalStatus: legal.Status(notices),
	})

	server := &http.Server{
		Handler: router,
		// Bound incomplete HTTP headers. Long-lived WebSockets use transport deadlines.
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Bind before logging readiness so a port conflict cannot produce a false startup
	// message.
	listener, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		return fmt.Errorf("cannot listen on %s (from %s): %w", cfg.ListenAddr, envListenAddr, err)
	}

	// SIGTERM is what Docker sends when stopping a container; SIGINT is Ctrl-C.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(listener) }()
	log.Printf("listening on %s", listener.Addr())

	select {
	case err := <-serveErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server stopped: %w", err)
		}
		return nil
	case <-ctx.Done():
		log.Printf("shutdown requested, closing connections")
	}

	// Share one shutdown budget between HTTP requests and WebSockets.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	// Stop new socket registration first. HTTP Shutdown ignores hijacked WebSockets;
	// closing the rooms signals their handlers to exit.
	rooms.StopAccepting()
	stopSweeping()
	shutdownErr := server.Shutdown(shutdownCtx)
	manager.Close()

	// The library may block CloseNow behind an in-flight closing handshake.
	// Start closure without another wait; process exit releases remaining sockets.
	if !waitWithin(shutdownCtx, rooms.Wait) {
		log.Printf("shutdown budget of %s expired; leaving the rest to the exit", cfg.ShutdownTimeout)
		rooms.CloseNow()
		_ = server.Close()
	}

	if shutdownErr != nil {
		return fmt.Errorf("shutdown did not complete within %s (from %s): %w",
			cfg.ShutdownTimeout, envShutdownTimeout, shutdownErr)
	}

	log.Printf("stopped cleanly")
	return nil
}

// waitWithin reports whether wait finishes before the deadline. On timeout, the
// waiting goroutine may remain until cleanup completes or the process exits.
func waitWithin(ctx context.Context, wait func()) bool {
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
