package legal

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// loaded builds what a running process holds, without a filesystem.
func loaded() *Documents {
	return &Documents{privacy: []byte("<p>privacy</p>"), imprint: []byte("<p>imprint</p>")}
}

func request(h http.Handler, method, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

func TestEnabledNoticesAreServed(t *testing.T) {
	pages := Pages(loaded())

	for _, tc := range []struct{ path, body, contentType string }{
		{PrivacyPath, "<p>privacy</p>", "text/html; charset=utf-8"},
		{ImprintPath, "<p>imprint</p>", "text/html; charset=utf-8"},
		{StylePath, "", "text/css; charset=utf-8"},
	} {
		rec := request(pages, http.MethodGet, tc.path)

		if rec.Code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", tc.path, rec.Code)
		}
		if got := rec.Header().Get("Content-Type"); got != tc.contentType {
			t.Errorf("GET %s content type = %q, want %q", tc.path, got, tc.contentType)
		}
		if tc.body != "" && rec.Body.String() != tc.body {
			t.Errorf("GET %s = %q, want %q", tc.path, rec.Body.String(), tc.body)
		}
		// A notice must be readable by somebody who is not playing, so serving
		// one allocates nothing and identifies nobody.
		if cookies := rec.Result().Cookies(); len(cookies) != 0 {
			t.Errorf("GET %s set %d cookie(s)", tc.path, len(cookies))
		}
		// A restart is what applies a change; a cached copy would outlive it.
		if got := rec.Header().Get("Cache-Control"); got != "no-store" {
			t.Errorf("GET %s Cache-Control = %q, want no-store", tc.path, got)
		}
	}
}

func TestHeadReportsTheSizeWithoutTheBody(t *testing.T) {
	rec := request(Pages(loaded()), http.MethodHead, PrivacyPath)

	if rec.Body.Len() != 0 {
		t.Errorf("HEAD returned %d bytes of body", rec.Body.Len())
	}
	if got, want := rec.Header().Get("Content-Length"), strconv.Itoa(len("<p>privacy</p>")); got != want {
		t.Errorf("HEAD Content-Length = %q, want %q", got, want)
	}
}

func TestNoticesAreNotWritable(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		rec := request(Pages(loaded()), method, PrivacyPath)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s %s = %d, want 405", method, PrivacyPath, rec.Code)
		}
		if got := rec.Header().Get("Allow"); got != "GET, HEAD" {
			t.Errorf("%s Allow = %q", method, got)
		}
	}
}

func TestDisabledNoticesReturnNotFound(t *testing.T) {
	// A link published while the feature was on has to fail visibly once it is off.
	for _, path := range []string{PrivacyPath, ImprintPath, StylePath, "/legal/anything"} {
		if code := request(Pages(nil), http.MethodGet, path).Code; code != http.StatusNotFound {
			t.Errorf("GET %s with notices off = %d, want 404", path, code)
		}
	}
}

func TestOnlyTheThreeNamedResourcesExist(t *testing.T) {
	// The handler compares the path against three fixed strings, so there is no
	// path-to-file mapping to attack. These record that.
	pages := Pages(loaded())

	for _, path := range []string{
		"/legal", "/legal/",
		"/legal/privacy.html",             // the file name rather than its address
		"/legal/notes.txt",                // another file the operator may keep there
		"/legal/PRIVACY",                  // the addresses are case-sensitive
		"/legal/../legal/privacy",         // traversal back into the subtree
		"/legal/%2e%2e/%2e%2e/etc/passwd", // an encoded escape attempt
	} {
		rec := request(pages, http.MethodGet, path)

		if rec.Code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", path, rec.Code)
		}
		if strings.Contains(rec.Body.String(), "<p>privacy</p>") {
			t.Errorf("GET %s disclosed a notice", path)
		}
	}
}

// wantPolicy is written out rather than referring to the constant, so an
// accidental edit fails this test instead of agreeing with itself.
const wantPolicy = "default-src 'none'; " +
	"style-src 'self'; " +
	"img-src 'self'; " +
	"script-src 'none'; " +
	"object-src 'none'; " +
	"base-uri 'self'; " +
	"form-action 'none'; " +
	"frame-ancestors 'none'"

func TestEveryLegalResponseCarriesTheStricterPolicy(t *testing.T) {
	// Operator HTML is the one place where the served markup was not written
	// here. This package has no build tags, so the policy applies in both builds.
	for _, tc := range []struct {
		name string
		docs *Documents
		path string
	}{
		{"a notice", loaded(), PrivacyPath},
		{"the stylesheet", loaded(), StylePath},
		{"an unknown resource", loaded(), "/legal/nothing"},
		{"a disabled notice", nil, PrivacyPath},
	} {
		got := request(Pages(tc.docs), http.MethodGet, tc.path).Header().Get("Content-Security-Policy")
		if got != wantPolicy {
			t.Errorf("%s: policy = %q, want %q", tc.name, got, wantPolicy)
		}
	}

	// Trusting operator HTML is not a reason to let it run code.
	for _, forbidden := range []string{"'unsafe-inline'", "'unsafe-eval'"} {
		if strings.Contains(ContentSecurityPolicy, forbidden) {
			t.Errorf("the legal policy contains %s", forbidden)
		}
	}
}

func TestAvailabilityReportsOnlyWhetherNoticesExist(t *testing.T) {
	// This is read by every visitor's browser. It says that notices exist, and
	// nothing about where they came from or what is in them.
	for _, tc := range []struct {
		docs *Documents
		want string
	}{
		{nil, `{"enabled":false}`},
		{loaded(), `{"enabled":true}`},
	} {
		rec := request(Status(tc.docs), http.MethodGet, StatusPath)

		if got := strings.TrimSpace(rec.Body.String()); got != tc.want {
			t.Errorf("body = %q, want %q", got, tc.want)
		}
		if got := rec.Header().Get("Cache-Control"); got != "no-store" {
			t.Errorf("Cache-Control = %q, want no-store", got)
		}
	}
}
