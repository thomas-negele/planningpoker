//go:build !embedassets

package transport

import "net/http"

// Development omits the production CSP to support Vite hot reloading.
func withSecurityHeaders(next http.Handler) http.Handler {
	return next
}
