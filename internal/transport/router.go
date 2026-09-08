// Package transport maps HTTP and WebSocket requests to room operations.
package transport

import "net/http"

// Options configures asset delivery, room creation and WebSocket handling.
type Options struct {
	// Assets serves the frontend and its static files.
	Assets http.Handler

	// Socket handles room WebSocket connections.
	Socket http.Handler

	// CreateGame creates an empty room and returns its URL identifier.
	CreateGame http.Handler
}

// NewRouter registers the application routes and applies build-specific security
// headers.
func NewRouter(opts Options) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /api/games", opts.CreateGame)

	mux.Handle("/ws/{roomID}", opts.Socket)

	// Serve frontend routes through the asset handler's SPA fallback.
	mux.Handle("/", opts.Assets)

	return withSecurityHeaders(mux)
}
