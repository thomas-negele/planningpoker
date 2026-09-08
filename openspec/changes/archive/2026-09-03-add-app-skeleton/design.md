## Context

The repository is empty; see `proposal.md` — Why for the motivation for starting with the delivery
machinery rather than the game rules. The constraints that shape this design are all already fixed
in `CLAUDE.md`: Go with `net/http` and no web framework, `github.com/coder/websocket` for the
real-time connection, Svelte 5 with TypeScript built by Vite without SvelteKit, the frontend
compiled into the binary with `go:embed`, a distroless container behind Caddy, and the rule that
nothing at runtime may be fetched from a foreign host.

Two properties of the toolchain drive most of the decisions below, and both are easy to miss:

- **`go:embed` refuses to compile when the directory it points at does not exist.** A fresh
  checkout has no `web/dist`, so a naive embed makes `go build ./...` and `go test ./...` fail
  until somebody has run `npm run build`. This is not a cosmetic annoyance; it breaks the ability
  to run Go tests at all.
- **`http.Server.Shutdown` does not wait for hijacked connections**, and a WebSocket connection is
  a hijacked connection. Graceful shutdown therefore has to close WebSocket connections itself;
  relying on `Shutdown` alone would return immediately while sockets are still open.

## Goals / Non-Goals

**Goals:**

- Prove, with a running container, that the three assembly-level risks are not risks: `go:embed`
  plus deep-link fallback, the Content-Security-Policy including the WebSocket, and the WebSocket
  upgrade through a TLS-terminating reverse proxy.
- Leave behind a shape that the next three changes can grow into without rearranging anything: a
  routing table with room for `/api/...`, a configuration path with room for more values, and a
  frontend project with room for real components.
- Keep `go build ./...` and `go test ./...` working on a checkout where `npm install` has never
  been run.

**Non-Goals:**

- Any game concept. There is no room, no participant, no deck, no vote in this change. Creating
  empty `internal/game` or `internal/hub` packages now would be speculative structure.
- Any visual design. The placeholder page exists to prove delivery, not to preview the table.
- The Caddy configuration on the VPS. It lives outside this repository. Verification that a
  TLS-terminating proxy forwards the upgrade correctly is in scope; the deployed proxy's config
  file is not.

## Decisions

### Development and production are separated by a build tag, not by configuration

`internal/webassets` gets two files with mutually exclusive build constraints. The production file
carries `//go:build embedassets` and contains the `go:embed` directive; the development file
carries `//go:build !embedassets` and reads the same files from disk at request time. The
`Dockerfile` compiles with `-tags embedassets`; a plain `go build` produces the development
variant.

The tag also decides whether the Content-Security-Policy middleware is installed.

Why a build tag rather than an environment variable, which would be the more usual choice: an
environment variable can be set wrongly in production, and the failure mode of "Content-Security-
Policy accidentally switched off on the public server" is silent and severe. A build tag cannot be
misconfigured after the fact — the production image either contains the policy or it is not the
production image. The cost is that toggling the behaviour requires a recompile, which is exactly
what we want here.

Alternative considered and rejected: making the embed unconditional and committing a placeholder
`web/dist/index.html` to the repository so the embed always has something to point at. It works,
but it puts build output under version control and creates a file that is confusingly stale most of
the time.

### The everyday development loop runs through Vite, not through Go

Two servers run side by side: the Go process on its port, and the Vite development server on
its own. Vite is configured to forward `/api` and `/ws` requests to the Go process. The developer
opens the Vite URL, gets hot module replacement for free, and the Go process is untouched by
frontend edits.

The disk-serving variant of `internal/webassets` is therefore not the primary development path —
it exists so that the Go binary is runnable and testable on its own, and so that a developer can
check the production-shaped serving behaviour (the deep-link fallback in particular) against a
real `npm run build` output without producing a container.

Alternative considered and rejected: having the Go process reverse-proxy unknown paths to the Vite
development server, so that there is only one URL to open in development. It is a nicer developer
experience, but it is Go code that exists only for development, has to be kept working, and can
itself be the thing that is broken. Vite's own proxy configuration achieves the same with three
lines of configuration and no Go code.

### The deep-link fallback distinguishes asset paths from application paths

The static file handler tries to serve the requested path. If no such file exists, the handler
returns `index.html` with status 200 — unless the path is inside the built asset directory, in
which case it returns 404.

The exception matters more than it looks. Vite emits hashed filenames into a single asset
directory. Without the exception, a stale or mistyped asset URL would receive an HTML document with
status 200, and the browser would report a syntax error deep inside what it thought was JavaScript
— one of the more time-consuming ways to lose an afternoon. With the exception, it reports a
missing file, which is what actually happened.

### The WebSocket endpoint keeps the library's default origin check

`github.com/coder/websocket` refuses an upgrade whose `Origin` header names a different host than
the request's `Host` header, unless that check is explicitly disabled. The default is kept.

The consequence to be aware of when deploying: this only works if the reverse proxy passes the
original `Host` through. Caddy's `reverse_proxy` does so by default, so nothing special is needed —
but if the upgrade is rejected in production while it works locally, this is the first thing to
look at, not the Content-Security-Policy.

### Graceful shutdown closes WebSocket connections explicitly

The process installs a signal handler for `SIGTERM` and `SIGINT`. On receiving one it calls
`http.Server.Shutdown` with a bounded timeout for ordinary HTTP requests, and separately closes
every open WebSocket connection, because `Shutdown` ignores hijacked connections entirely.

For this change the set of open connections is tracked by the small piece of code that owns the
echo endpoint. From the `add-room-hub-and-protocol` change onwards, connections are owned by their
room goroutine and shutdown becomes that layer's responsibility; the shape here is deliberately
small enough to be replaced rather than extended.

### Configuration is a struct built once at startup

`cmd/planningpoker` reads the environment into a small configuration struct and passes it into the
components that need it. Nothing reads the environment except that one function, and there are no
package-level configuration globals.

This is worth insisting on even though the struct has only two fields today. Package-level globals
read from the environment at import time are the reason configuration becomes untestable and
unchangeable later, and this project already knows it will grow more of these values — the room
grace period first among them.

The two values:

- **`PLANNINGPOKER_LISTEN_ADDR`**, the address the server listens on, defaulting to `:8080`.
- **`PLANNINGPOKER_SHUTDOWN_TIMEOUT`**, how long the process stays polite after being asked to
  stop, defaulting to `5s`. The default is deliberately below the ten seconds the container runtime
  waits after `SIGTERM` before sending `SIGKILL`: were the two equal, both clocks could expire
  together and the process would be killed in the middle of the cleanup this value exists to allow.
  A test asserts the relationship so it cannot be broken silently.

In both cases an unparseable value stops the process rather than falling back, so that a typo in
`compose.yaml` is discovered at deploy time and not months later when someone wonders why a setting
has no effect. Zero and negative shutdown timeouts are refused rather than given an invented
meaning, since zero could be read either as "drop everything at once" or as "wait forever"; someone
who wants the former can write `1ms`.

### The container publishes no host port

`compose.yaml` declares the port as `expose`, not `ports`. Caddy reaches the container by its
service name over the shared Docker network. A published host port would make the application
reachable over plain HTTP, bypassing TLS termination, which is both a security regression and an
easy thing to do by accident.

## Risks / Trade-offs

- **`connect-src 'self'` might not cover the WebSocket in some browser.** Content-Security-Policy
  Level 3 says `'self'` matches same-origin `ws:` and `wss:`, but older behaviour differed, and if
  this is wrong the entire application is dead rather than one image. → This is the single reason
  the WebSocket probe is in this change. It is verified in a real browser against a production
  build before the change is considered done. If it turns out to be wrong, the fix is an explicit
  `connect-src 'self' wss://<host>` and a note in `CLAUDE.md`; it is not a reason to weaken any
  other directive.
- **The `Origin` check can fail behind the reverse proxy** while working perfectly on localhost,
  producing an upgrade rejection that looks like a policy problem. → Verify through a proxy, not
  only directly, and record in the task list which of the two checks was performed.
- **The temporary echo endpoint is throwaway code.** It will be deleted in the change that
  introduces the real protocol. → Keep it in one small file with no dependencies on anything else,
  so that deleting it is a single-file operation rather than an untangling.
- **The build-tag split can rot**, because the development variant is exercised constantly and the
  production variant only when a container is built. A change that breaks embedding would not be
  noticed locally. → `docker compose up --build` is part of the acceptance of every change from
  here on, not only this one.
- **Unpinned base images drift.** A Node or Go minor version bump arriving silently in a rebuild
  can break the build months after this change. → Pin explicit versions in the `Dockerfile`.

## Migration Plan

There is nothing to migrate: the change creates a repository that had no code. Deployment is
`docker compose up --build` on the VPS, and rollback is not deploying it. From the next change
onwards, rollback means redeploying the previous image, with the standing consequence that any
deployment destroys the rooms currently in memory — as recorded in `CLAUDE.md`.

## Decisions taken with the owner during this proposal

These answer questions that `CLAUDE.md` listed as "ask the owner, do not decide alone". None of
them are implemented by this change; they are recorded here so that the later changes are written
against a settled understanding rather than re-opening the discussion.

1. **Reveal and revote may be triggered by any participant.** There is no host role, no creator
   identity in the state, and therefore no question of what happens when the creator leaves.
2. **A round may be revealed at any time**, including while some participants have not voted.
   Unvoted participants simply show no card. This means a single absent participant can never block
   a meeting, and it keeps the reveal rule trivial.
3. **After the reveal, each participant's card is shown, plus a count per size** — for example
   "2 × M, 1 × L, 1 × ☕". No average is computed: t-shirt sizes are an ordered scale without
   arithmetic, so any mean would be an invented number.
4. **A room outlives its last participant for a grace period** (suggested five minutes) before
   being discarded, so that a reload or a brief network drop does not destroy the session. The
   grace period is a behaviour-governing value in the sense of the owner's global rules.
5. **A second cookie holds an opaque participant identifier**, so a reloading browser is reseated
   in the same chair with the same vote instead of appearing as a duplicate. Two tabs of the same
   browser are consequently the same participant, and a private window is a different one.
6. **Duplicate names are allowed, and a seated participant may change their name.** There is no
   uniqueness rule; a rename is broadcast like any other state change.
7. **Behaviour-governing values are environment variables with documented defaults.** No
   administration area is built for the MVP. The consequence, which must be stated whenever such a
   value is added: changing one requires restarting the container, and because there is no
   persistence, a restart destroys every open room.

## Open Questions

- The container's internal port is assumed to be `8080` and the Compose service name
  `planningpoker`. Both are trivially changeable and neither affects the specs, the approach, or
  the task breakdown.
