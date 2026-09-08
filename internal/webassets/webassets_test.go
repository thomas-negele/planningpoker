package webassets

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

// build stands in for the output of `npm run build`: an application document and
// one hashed asset below the asset directory.
func build() fstest.MapFS {
	return fstest.MapFS{
		"index.html":              {Data: []byte("<!doctype html><title>Planning Poker</title>")},
		"assets/index-abc123.js":  {Data: []byte("console.log('bundle')")},
		"assets/index-abc123.css": {Data: []byte("body{margin:0}")},
		"favicon-placeholder.txt": {Data: []byte("a file outside the asset directory")},
	}
}

func get(t *testing.T, h http.Handler, target string) *http.Response {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec.Result()
}

func body(t *testing.T, res *http.Response) string {
	t.Helper()
	defer func() { _ = res.Body.Close() }()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("reading body: %v", err)
	}
	return string(b)
}

func TestExistingAssetIsServedAsItself(t *testing.T) {
	h := NewFromFS(build())

	for _, tc := range []struct {
		path        string
		wantType    string
		wantContent string
	}{
		{"/assets/index-abc123.js", "text/javascript", "console.log('bundle')"},
		{"/assets/index-abc123.css", "text/css", "body{margin:0}"},
	} {
		res := get(t, h, tc.path)
		if res.StatusCode != http.StatusOK {
			t.Errorf("%s: status = %d, want 200", tc.path, res.StatusCode)
		}
		if got := res.Header.Get("Content-Type"); !contains(got, tc.wantType) {
			t.Errorf("%s: Content-Type = %q, want it to mention %q", tc.path, got, tc.wantType)
		}
		if got := body(t, res); got != tc.wantContent {
			t.Errorf("%s: body = %q, want %q", tc.path, got, tc.wantContent)
		}
	}
}

func TestUnknownPathReturnsTheDocument(t *testing.T) {
	h := NewFromFS(build())

	// The middle one is the case that matters in practice: a client-side room URL
	// opened directly or hard-reloaded must produce the application, not a 404.
	for _, path := range []string{"/", "/g/abc123", "/some/deep/unknown/path"} {
		res := get(t, h, path)
		if res.StatusCode != http.StatusOK {
			t.Errorf("%s: status = %d, want 200", path, res.StatusCode)
		}
		if got := res.Header.Get("Content-Type"); !contains(got, "text/html") {
			t.Errorf("%s: Content-Type = %q, want it to mention text/html", path, got)
		}
		if got := body(t, res); !contains(got, "Planning Poker") {
			t.Errorf("%s: body = %q, want the application document", path, got)
		}
	}
}

func TestMissingAssetFailsVisibly(t *testing.T) {
	h := NewFromFS(build())

	for _, path := range []string{"/assets/index-deleted.js", "/assets/nothing-here.css", "/assets/"} {
		res := get(t, h, path)
		if res.StatusCode != http.StatusNotFound {
			t.Errorf("%s: status = %d, want 404 — a stale asset reference must not be answered with HTML",
				path, res.StatusCode)
		}
		if got := body(t, res); contains(got, "Planning Poker") {
			t.Errorf("%s: body contained the application document, want a not-found response", path)
		}
	}
}

func TestDocumentIsNotCached(t *testing.T) {
	// After a deployment the hashed asset names have changed, so a cached document
	// would ask the browser to fetch files that no longer exist.
	res := get(t, NewFromFS(build()), "/g/abc123")
	if got := res.Header.Get("Cache-Control"); got != "no-cache" {
		t.Errorf("Cache-Control = %q, want %q", got, "no-cache")
	}
}

func TestMissingBuildExplainsItself(t *testing.T) {
	h := unavailableHandler{err: errNoBuild}
	res := get(t, h, "/")
	if res.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", res.StatusCode)
	}
	if got := body(t, res); !contains(got, "npm run build") {
		t.Errorf("body = %q, want it to say how to produce a build", got)
	}
}

func contains(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}
