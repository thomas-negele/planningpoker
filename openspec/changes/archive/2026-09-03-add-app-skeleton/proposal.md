## Why

The repository contains no code at all. Every later change — the game rules, the room hub, the
table user interface — has to be delivered by the same machinery: one Go binary that serves a
Svelte single-page application compiled into it, inside one Docker container, behind Caddy, with a
Content-Security-Policy header that the browser enforces.

That machinery carries the only real technical risk in this project. The planning-poker rules
themselves are a handful of pure functions over a small amount of data and are unlikely to
surprise anyone. The parts that can genuinely fail are the ones that only fail once everything is
assembled: whether `go:embed` plus the single-page-application fallback survives a hard reload of a
deep link, whether `connect-src 'self'` really permits the WebSocket connection in current
browsers (CLAUDE.md flags this explicitly as unverified), and whether Caddy forwards the WebSocket
upgrade request. Discovering any of these late means discovering them entangled with finished
feature code. This change proves all three while there is nothing else in the way.

## What Changes

- **New Go module** `de.thomasnegele.planningpoker` with the repository layout CLAUDE.md
  prescribes: `cmd/planningpoker/`, `internal/transport/`, `internal/webassets/`, `web/`.
- **New Svelte 5 + TypeScript + Vite frontend** in `web/`, containing a deliberate placeholder
  page only. It shows that the application is being served, and it opens a WebSocket to prove the
  connection works. No table, no cards, no name entry — those arrive in later changes.
- **Static asset serving with single-page-application fallback**: any request path that does not
  match a real file in the built frontend returns `index.html`, so that a future client-side route
  such as `/g/<room-id>` works on a hard reload rather than returning 404.
- **A build-tag split** between development and production asset serving: in development the files
  are read from disk so that a frontend edit does not require recompiling the Go binary; in
  production they come from the embedded filesystem baked in by `go:embed`.
- **Content-Security-Policy middleware** applying the exact policy agreed in CLAUDE.md to every
  application response, in the production path only. Verifying that policy in a real browser —
  including that the WebSocket is not blocked — is an acceptance criterion of this change, not a
  follow-up.
- **A temporary WebSocket endpoint** at `/ws` using `github.com/coder/websocket`, which accepts the
  upgrade and echoes back what it receives. Its entire purpose is to prove the upgrade path through
  Caddy and the `connect-src` directive. It carries no game meaning and **is replaced by the real
  protocol in the `add-room-hub-and-protocol` change**. It is named openly here so it can be
  dropped from scope if you would rather prove the WebSocket path later.
- **Configuration from environment variables** with documented defaults, read once at startup and
  passed explicitly into the components that need it. The skeleton needs two values — the address
  the server listens on, and how long shutdown stays polite before dropping connections; the
  mechanism is established now so that later values — the room grace period above all — have
  somewhere to go.
- **Graceful shutdown**: the process stops accepting new connections and closes open ones on
  `SIGTERM`, the signal Docker sends when stopping a container.
- **Multi-stage `Dockerfile` and `compose.yaml`** producing a single container from a distroless
  base image, running as an unprivileged user, publishing no host port.

## Capabilities

### New Capabilities
- `app-delivery`: How the application reaches the browser and what the browser is permitted to do
  once it arrives — serving the single-page application from a single origin, the fallback that
  makes deep links survive a reload, the Content-Security-Policy that forbids every foreign origin,
  the availability of a WebSocket endpoint, and how the running process is configured and stopped.

### Modified Capabilities

None. This is the first change in the repository; there are no existing specs.

## Impact

- **Created code**: `go.mod`, `cmd/planningpoker/`, `internal/transport/`,
  `internal/webassets/`, `web/` (with `package.json`, `vite.config.ts`, `svelte.config.js`,
  `tsconfig.json`, `src/`), `Dockerfile`, `compose.yaml`, `.dockerignore`, `.gitignore`.
- **New Go dependency**: `github.com/coder/websocket`. No other runtime dependency; routing is
  `net/http` from the standard library, as decided.
- **New build-time dependencies**: Node.js, Vite, Svelte 5, TypeScript, `svelte-check` — all
  confined to the `web/` directory and the Node stage of the Docker build. Nothing they install
  reaches the browser except as bundled output served from this origin.
- **Deployment**: after this change the application is deployable to the Hetzner VPS and answers on
  its port, showing a placeholder. Caddy configuration on the VPS is outside this repository and is
  not part of this change, but the WebSocket upgrade must be verified through a reverse proxy
  before this change is considered done.
- **Not affected**: no game logic exists yet, so `internal/game` and `internal/hub` are not created
  by this change. Creating empty packages ahead of their content would be speculative.
