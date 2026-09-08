//go:build embedassets

package transport

import "net/http"

// ContentSecurityPolicy restricts production resources to this origin, allowing
// embedded data images. It also blocks object embedding, cross-origin form
// submissions and framing. Inline scripts and eval are not permitted.
const ContentSecurityPolicy = "default-src 'self'; " +
	"connect-src 'self'; " +
	"img-src 'self' data:; " +
	"font-src 'self'; " +
	"style-src 'self'; " +
	"script-src 'self'; " +
	"object-src 'none'; " +
	"base-uri 'self'; " +
	"form-action 'self'; " +
	"frame-ancestors 'none'"

// withSecurityHeaders adds the production CSP. The development build omits it
// to support Vite's inline scripts and hot-reload connection.
func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", ContentSecurityPolicy)
		next.ServeHTTP(w, r)
	})
}
