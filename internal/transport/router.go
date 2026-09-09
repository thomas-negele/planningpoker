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

	// LegalPages serves the reserved /legal subtree. It is registered whether or
	// not an operator supplied notices, so that a request for one cannot fall
	// through to the frontend's catch-all document and be answered with the
	// application instead of a notice.
	LegalPages http.Handler

	// LegalStatus reports whether legal notices are available on this installation.
	LegalStatus http.Handler
}

// NewRouter registers the application routes and applies build-specific security
// headers.
func NewRouter(opts Options) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /api/games", opts.CreateGame)

	mux.Handle("GET /api/legal", opts.LegalStatus)

	// Both patterns, so that /legal answers for itself rather than being redirected
	// to /legal/ only to be refused there.
	mux.Handle("/legal", opts.LegalPages)
	mux.Handle("/legal/", opts.LegalPages)

	mux.Handle("/ws/{roomID}", opts.Socket)

	// Serve frontend routes through the asset handler's SPA fallback.
	mux.Handle("/", opts.Assets)

	return withSecurityHeaders(mux)
}
