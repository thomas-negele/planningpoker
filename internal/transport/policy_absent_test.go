//go:build !embedassets

package transport

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// The counterpart to policy_test.go, which runs only in the production build.
//
// The specification requires that the development build applies no
// Content-Security-Policy, because the Vite development server relies on inline
// scripts and its own WebSocket for hot reloading and this policy forbids both.
// Without this test the requirement would hold only by inspection: the production
// tests cannot observe the development build, so a stray header added here would go
// unnoticed until someone wondered why hot reloading had stopped working.
func TestDevelopmentBuildAppliesNoPolicy(t *testing.T) {
	router := testRouter()

	for _, path := range []string{"/", "/g/abc123", "/assets/index-abc123.js"} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

		if got := rec.Header().Get("Content-Security-Policy"); got != "" {
			t.Errorf("%s: development build sent a Content-Security-Policy (%q); it must not, "+
				"because the policy blocks the Vite development server's hot reloading", path, got)
		}
	}
}
