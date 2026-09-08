## 1. Go module and repository skeleton

- [x] 1.1 Create `go.mod` with module path `de.thomasnegele.planningpoker` and the Go version pinned; verify `go build ./...` succeeds on the empty module
- [x] 1.2 Add `.gitignore` covering `web/node_modules/`, the frontend build output and the compiled binary, and `.dockerignore` covering the same plus `.git`; verify `git status` shows no build output as untracked after a frontend build
- [x] 1.3 Add `cmd/planningpoker/main.go` that starts an HTTP server on a hardcoded address and answers one route, so there is something runnable to grow from; verify `go run ./cmd/planningpoker` serves that route

## 2. Frontend project and placeholder page

- [x] 2.1 Create the Vite + Svelte 5 + TypeScript project in `web/` (`package.json`, `vite.config.ts`, `svelte.config.js`, `tsconfig.json`, `index.html`, `src/main.ts`, `src/App.svelte`); verify `npm install && npm run build` produces `index.html` and a hashed asset directory in the build output directory — which is `internal/webassets/dist/`, not `web/dist/`, because `go:embed` can only name files inside its own package directory
- [x] 2.2 Write the placeholder `App.svelte` — a heading naming the application and a visible connection-status line — using no external font, no icon library, no remote image; verify `npm run check` reports no errors
- [x] 2.3 Configure the Vite development server to forward `/api` and `/ws` to the Go process's port; verify that with both processes running, a request to the Vite URL under `/ws` reaches the Go process
- [x] 2.4 Search the built output for `http://` and `https://` and confirm every match is a comment, a source-map path, or otherwise not a runtime fetch from a foreign host; record the result — this is the check `CLAUDE.md` requires rather than assumes

## 3. Asset serving and the deep-link fallback

- [x] 3.1 Create `internal/webassets` with the development variant (build constraint `!embedassets`) that serves the built frontend from disk; verify `go build ./...` succeeds on a checkout where the build output does not exist
- [x] 3.2 Add the production variant (build constraint `embedassets`) with the `go:embed` directive over the built frontend; verify `go build -tags embedassets ./...` succeeds after `npm run build` and fails clearly without it
- [x] 3.3 Implement the fallback handler: an existing file is served with its correct content type, an unknown path returns `index.html` with status 200, and an unknown path inside the asset directory returns 404; verify with Go tests covering all three cases against a small in-memory `fs.FS`
- [x] 3.4 Verify the fallback end to end in a browser: load a path such as `/g/abc123` directly and reload it hard, and confirm the application document is returned rather than a 404

## 4. HTTP transport, configuration and shutdown

- [x] 4.1 Create `internal/transport` with the router built on `http.ServeMux`, taking the asset handler and configuration as parameters rather than reading globals; verify a Go test can construct the router and exercise a route with `httptest`
- [x] 4.2 Implement the Content-Security-Policy middleware with exactly the directives listed in the spec, installed only in the `embedassets` build; verify with a Go test that the header is present and byte-for-byte correct, and that it contains neither `'unsafe-inline'` nor `'unsafe-eval'`
- [x] 4.3 Implement configuration: read `PLANNINGPOKER_LISTEN_ADDR` into a configuration struct in `cmd/planningpoker`, default `:8080`, with a full-sentence comment explaining the value; verify with Go tests that unset yields the default, a valid value overrides it, and an invalid value returns an error naming the variable and the value
- [x] 4.4 Make an invalid configuration stop the process with a non-zero exit status and the error message, rather than falling back to the default; verify by running the binary with a deliberately broken value and checking the exit status
- [x] 4.6 Make the shutdown budget configurable as `PLANNINGPOKER_SHUTDOWN_TIMEOUT`, defaulting to five seconds so it stays below the ten-second grace period the container runtime allows before sending `SIGKILL`; reject zero, negative and unparseable values rather than giving an ambiguous boundary value a meaning; verify with Go tests covering the default, valid overrides, every rejected form, and that the default is below the grace period
- [x] 4.5 Implement graceful shutdown on `SIGTERM` and `SIGINT`: stop accepting new connections, close open WebSocket connections explicitly because `http.Server.Shutdown` ignores hijacked connections, and exit within a bounded timeout; verify by sending `SIGTERM` to a running process with a connected client and confirming exit status 0 without the timeout elapsing

## 5. WebSocket connection probe

- [x] 5.1 Add `github.com/coder/websocket` to `go.mod` and implement the temporary echo endpoint at `/ws` in a single self-contained file, keeping the library's default `Origin` check; verify with a Go test that a client connects, sends a message and receives it back
- [x] 5.2 Wire the placeholder page to open the WebSocket, send one message, display the round trip in the connection-status line, and close cleanly on page unload; verify in the browser that the status line reports success
- [x] 5.3 Verify no goroutine or connection is leaked when the client disconnects — run the WebSocket tests under `go test -race ./...` and confirm the handler returns after the client closes

## 6. Container build and deployment shape

- [x] 6.1 Write the multi-stage `Dockerfile`: a pinned Node stage running `npm ci && npm run build`, a pinned Go stage compiling with `CGO_ENABLED=0`, `-tags embedassets` and `-ldflags="-s -w"`, and a final `gcr.io/distroless/static-debian12:nonroot` stage; verify the image builds and `docker image ls` shows a size in the expected range of roughly 15 MB
- [x] 6.2 Write `compose.yaml` exposing the container port without publishing it on the host, setting `PLANNINGPOKER_LISTEN_ADDR` with an explanatory comment; verify `docker compose up --build` from a clean checkout serves the application and that the port is not reachable from the host
- [x] 6.3 Confirm the container runs as an unprivileged user; verify by inspecting the image's configured user, since the distroless image has no shell to ask from inside

## 7. Acceptance in a real browser

- [x] 7.1 Against a production build, load the application and inspect the browser's network panel; verify every request including the WebSocket targets one origin and no request goes to any external host — verified from a Chrome network log: the document, the bundled script, the stylesheet, `favicon.ico` and `ws://127.0.0.1:8080/ws` were the complete set, all on the one origin
- [x] 7.2 Against a production build, confirm the browser console reports zero Content-Security-Policy violations, and specifically that `connect-src 'self'` does not block the WebSocket — this is the assumption `CLAUDE.md` flags as unverified, so record the browser and version the check was performed with — **verified in Google Chrome 152.0.7977.75**: the rendered page reported `data-state="ok"`, "connected, echo received", under the full production policy, and the console contained no violation of any kind
- [x] 7.3 Put a TLS-terminating reverse proxy in front of the container and confirm the WebSocket upgrade succeeds over `wss`, so that the `Origin`-versus-`Host` behaviour is proven through a proxy and not only against localhost; record which proxy was used — **Caddy v2.11.4** with its internal certificate authority: `https://localhost/` and `/g/abc123` both returned 200, the policy header survived the proxy, the `wss` echo round trip succeeded, and an upgrade carrying `Origin: https://evil.example.com` was correctly refused
- [x] 7.4 Run the full command set from `CLAUDE.md` and confirm each is clean: `go build ./...`, `go test -race ./...`, `go vet ./...`, `gofmt -l .` with empty output, and `npm run check` in `web/` — all clean, and the Go commands were run in both build variants
