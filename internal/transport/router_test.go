package transport

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// stub answers with a fixed marker so a test can tell which handler was reached.
func stub(marker string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(marker))
	})
}

func testRouter() http.Handler {
	return NewRouter(Options{
		Assets:     stub("assets"),
		Socket:     stub("socket"),
		CreateGame: stub("createGame"),
	})
}

func TestRouterSendsPathsToTheRightHandler(t *testing.T) {
	router := testRouter()

	for _, tc := range []struct{ method, path, want string }{
		{http.MethodPost, "/api/games", "createGame"},
		{http.MethodGet, "/ws/ABC123", "socket"},
		{http.MethodGet, "/", "assets"},
		{http.MethodGet, "/g/ABC123", "assets"},
		{http.MethodGet, "/assets/index-abc123.js", "assets"},
	} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))
		if got := rec.Body.String(); got != tc.want {
			t.Errorf("%s %s reached the %q handler, want %q", tc.method, tc.path, got, tc.want)
		}
	}
}

func TestCreatingAGameRequiresTheRightMethod(t *testing.T) {
	// The method is part of the routing pattern, so a stray link, a prefetch or a
	// crawler cannot create rooms simply by fetching a URL.
	router := testRouter()

	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(method, "/api/games", nil))
		if got := rec.Body.String(); got == "createGame" {
			t.Errorf("%s /api/games created a game; only POST may", method)
		}
	}
}
