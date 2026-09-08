// Package webassets serves the frontend from embedded production assets or an
// on-disk build in development. Missing builds produce a 503 response.
package webassets

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// assetDir is the subdirectory the bundler writes hashed files into. A request
// for a missing file below it is answered with 404 rather than with the HTML
// document; see the handler for why that distinction matters.
const assetDir = "assets"

// indexFile is the document every unrecognised path falls back to, which is what
// makes a client-side route survive being opened directly or reloaded.
const indexFile = "index.html"

// New opens the build-selected asset source, falling back to a 503 handler if
// unavailable.
func New() http.Handler {
	fsys, err := open()
	if err != nil {
		return unavailableHandler{err: err}
	}
	return NewFromFS(fsys)
}

// NewFromFS serves assets and uses index.html as the fallback for frontend routes.
func NewFromFS(fsys fs.FS) http.Handler {
	return &handler{fsys: fsys, files: http.FileServerFS(fsys)}
}

type handler struct {
	fsys  fs.FS
	files http.Handler
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")

	if name != "" && exists(h.fsys, name) {
		h.files.ServeHTTP(w, r)
		return
	}

	// Missing static assets must return 404, not the SPA HTML document.
	if name == assetDir || strings.HasPrefix(name, assetDir+"/") {
		http.NotFound(w, r)
		return
	}

	h.serveIndex(w, r)
}

func (h *handler) serveIndex(w http.ResponseWriter, r *http.Request) {
	f, err := h.fsys.Open(indexFile)
	if err != nil {
		http.Error(w, "application document is missing from the build", http.StatusInternalServerError)
		return
	}
	defer func() { _ = f.Close() }()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Do not cache index.html; a deployment may reference new hashed assets.
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_, _ = io.Copy(w, f)
	}
}

func exists(fsys fs.FS, name string) bool {
	info, err := fs.Stat(fsys, name)
	return err == nil && !info.IsDir()
}

// unavailableHandler reports that no frontend build is available.
type unavailableHandler struct{ err error }

func (h unavailableHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = io.WriteString(w, "No built frontend is available.\n\n"+
		"Run `npm run build` in web/ to produce one, or use the Vite development\n"+
		"server (`npm run dev` in web/), which serves the frontend itself and\n"+
		"forwards /api and /ws to this process.\n\n"+
		"Underlying error: "+h.err.Error()+"\n")
}

// errNoBuild identifies a missing frontend build.
var errNoBuild = errors.New("no built frontend found")
