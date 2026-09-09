package legal

import (
	_ "embed"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// The fixed addresses under the reserved /legal subtree. There is no mapping from
// a request path to a file name anywhere in this package: these three strings are
// compared literally, so no request can select another file, list the directory or
// escape it by any spelling of "..".
const (
	PrivacyPath = "/legal/privacy"
	ImprintPath = "/legal/imprint"
	StylePath   = "/legal/style.css"
)

// StatusPath reports whether notices are available, so the frontend can decide
// whether to show its links without a build-time setting.
const StatusPath = "/api/legal"

// styleSheet is compiled into the binary rather than shipped with the frontend
// build, so a notice is styled even when it is opened directly and regardless of
// how the frontend was built.
//
//go:embed style.css
var styleSheet []byte

// ContentSecurityPolicy applies to every response under /legal, in the
// development build as well as the production one. The notices are operator-
// supplied HTML, and this is what keeps a stray script tag or an image pasted in
// from a foreign host from being loaded: scripts and plugins cannot run at all,
// styles and images may only come from this origin, the page cannot be framed and
// a form in it cannot post anywhere.
//
// It deliberately differs from the application's own policy — it is stricter, and
// applying it here does not widen or weaken that one.
const ContentSecurityPolicy = "default-src 'none'; " +
	"style-src 'self'; " +
	"img-src 'self'; " +
	"script-src 'none'; " +
	"object-src 'none'; " +
	"base-uri 'self'; " +
	"form-action 'none'; " +
	"frame-ancestors 'none'"

const htmlContentType = "text/html; charset=utf-8"

// Pages serves the reserved /legal subtree. A nil docs means the feature is
// switched off, in which case every path below /legal answers 404 — the subtree is
// still reserved, because falling through to the frontend's catch-all document
// would answer a request for a notice with the application itself.
func Pages(docs *Documents) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", ContentSecurityPolicy)

		// The notices are documents to read. Anything else is refused before the
		// path is examined.
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "legal notices can only be read", http.StatusMethodNotAllowed)
			return
		}

		if docs == nil {
			http.NotFound(w, r)
			return
		}

		switch r.URL.Path {
		case PrivacyPath:
			serve(w, r, htmlContentType, docs.privacy)
		case ImprintPath:
			serve(w, r, htmlContentType, docs.imprint)
		case StylePath:
			serve(w, r, "text/css; charset=utf-8", styleSheet)
		default:
			http.NotFound(w, r)
		}
	})
}

// availability is the whole of the metadata response. It says whether notices
// exist and nothing else: not where they came from, not what they contain.
type availability struct {
	Enabled bool `json:"enabled"`
}

// Status answers whether legal notices are available on this installation.
func Status(docs *Documents) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		if err := json.NewEncoder(w).Encode(availability{Enabled: docs != nil}); err != nil {
			log.Printf("writing the legal availability response: %v", err)
		}
	})
}

// serve writes one in-memory document. Nothing here opens a file, allocates a
// room or sets a cookie: a notice must be readable by someone who is not playing.
func serve(w http.ResponseWriter, r *http.Request, contentType string, body []byte) {
	w.Header().Set("Content-Type", contentType)

	// Do not cache. Enabling, disabling or editing notices takes effect at a
	// restart, and a cached copy in a browser or proxy would outlive that.
	w.Header().Set("Cache-Control", "no-store")

	// Set the length explicitly so a HEAD request, which writes no body, still
	// reports the size the corresponding GET would return.
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(http.StatusOK)

	if r.Method != http.MethodHead {
		_, _ = w.Write(body)
	}
}
