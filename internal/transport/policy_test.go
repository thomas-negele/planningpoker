//go:build embedassets

package transport

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// wantPolicy is written out here rather than referring to the constant, so that an
// accidental edit to the policy fails this test instead of silently agreeing with
// itself.
const wantPolicy = "default-src 'self'; " +
	"connect-src 'self'; " +
	"img-src 'self' data:; " +
	"font-src 'self'; " +
	"style-src 'self'; " +
	"script-src 'self'; " +
	"object-src 'none'; " +
	"base-uri 'self'; " +
	"form-action 'self'; " +
	"frame-ancestors 'none'"

func TestEveryResponseCarriesThePolicy(t *testing.T) {
	router := testRouter()

	// The document and a static asset are both application responses and both must
	// carry the header.
	for _, path := range []string{"/", "/g/abc123", "/assets/index-abc123.js"} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

		got := rec.Header().Get("Content-Security-Policy")
		if got != wantPolicy {
			t.Errorf("%s: policy mismatch\n got: %q\nwant: %q", path, got, wantPolicy)
		}
	}
}

func TestPolicyForbidsTheTwoValuesThatWouldDefeatIt(t *testing.T) {
	// These are precisely what makes a Content-Security-Policy worth having. If a
	// violation ever needs fixing, the code that caused it gets fixed, not this.
	for _, forbidden := range []string{"'unsafe-inline'", "'unsafe-eval'"} {
		if strings.Contains(ContentSecurityPolicy, forbidden) {
			t.Errorf("the policy contains %s", forbidden)
		}
	}
}

func TestPolicyNamesEveryDirectiveExplicitly(t *testing.T) {
	// default-src is a catch-all, but the explicit directives exist so that a later
	// edit cannot widen one resource type by accident. Losing one would be silent.
	for _, directive := range []string{
		"default-src", "connect-src", "img-src", "font-src", "style-src",
		"script-src", "object-src", "base-uri", "form-action", "frame-ancestors",
	} {
		if !strings.Contains(ContentSecurityPolicy, directive+" ") {
			t.Errorf("the policy does not name %s", directive)
		}
	}
}
