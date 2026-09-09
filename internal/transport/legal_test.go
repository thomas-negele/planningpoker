package transport

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"de.thomasnegele.planningpoker/internal/legal"
	"de.thomasnegele.planningpoker/internal/webassets"
)

// routerWithNotices wires the real legal handlers and a real asset handler,
// which is the point of these tests: the two must not answer for each other.
func routerWithNotices(t *testing.T, dir string) http.Handler {
	t.Helper()

	notices, err := legal.Load(dir)
	if err != nil {
		t.Fatalf("loading notices from %q: %v", dir, err)
	}

	return NewRouter(Options{
		Assets: webassets.NewFromFS(fstest.MapFS{
			"index.html": &fstest.MapFile{Data: []byte("<!doctype html><title>Planning Poker</title>")},
		}),
		Socket:      stub("socket"),
		CreateGame:  stub("createGame"),
		LegalPages:  legal.Pages(notices),
		LegalStatus: legal.Status(notices),
	})
}

func noticesDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	for name, content := range map[string]string{
		"privacy.html": "<title>Privacy</title>the privacy notice",
		"imprint.html": "<title>Legal notice</title>the imprint",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("writing %s: %v", name, err)
		}
	}
	return dir
}

func TestNoticesAreServedThroughTheRouter(t *testing.T) {
	router := routerWithNotices(t, noticesDir(t))

	for _, tc := range []struct{ path, want string }{
		{"/legal/privacy", "the privacy notice"},
		{"/legal/imprint", "the imprint"},
		{"/api/legal", `{"enabled":true}`},
	} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))

		if rec.Code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", tc.path, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), tc.want) {
			t.Errorf("GET %s = %q, want it to contain %q", tc.path, rec.Body.String(), tc.want)
		}
	}
}

func TestAMissingNoticeDoesNotFallBackToTheApplication(t *testing.T) {
	// Every other unknown path is answered with index.html so a client-side
	// route survives a reload. Under /legal that would be wrong: somebody asking
	// for a notice would be shown the game with no way to tell it is absent.
	for _, tc := range []struct {
		name  string
		dir   string
		paths []string
	}{
		{"switched off", "", []string{"/legal", "/legal/privacy", "/legal/imprint", "/api/legal"}},
		{"on, but this resource does not exist", noticesDir(t), []string{"/legal", "/legal/", "/legal/other"}},
	} {
		router := routerWithNotices(t, tc.dir)

		for _, path := range tc.paths {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

			if strings.Contains(rec.Body.String(), "Planning Poker") {
				t.Errorf("%s: GET %s returned the application document", tc.name, path)
			}
			if path == "/api/legal" {
				continue // answers 200 with {"enabled":false}, not 404
			}
			if rec.Code != http.StatusNotFound {
				t.Errorf("%s: GET %s = %d, want 404", tc.name, path, rec.Code)
			}
		}
	}
}

func TestNoticesCarryTheirOwnPolicyInEveryBuild(t *testing.T) {
	// No build tag on this file, so it runs in the development build, which
	// applies no application policy, and in production, which applies one to
	// every response. In both the notice's own stricter policy is what arrives.
	router := routerWithNotices(t, noticesDir(t))

	for _, path := range []string{"/legal/privacy", "/legal/style.css", "/legal/absent"} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

		if got := rec.Header().Get("Content-Security-Policy"); got != legal.ContentSecurityPolicy {
			t.Errorf("GET %s policy = %q, want the legal policy", path, got)
		}
	}
}
