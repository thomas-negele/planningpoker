# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Status of this document

**The application exists and works.** This file was written before any of it did, and much of it
therefore reads as a plan; treat it as the record of *why* things are the way they are, not as a
description of what is there. Where this document and the specifications disagree, **the
specifications win** — they are kept in step with the code, and this file is not.

### Where behaviour is written down

`openspec/specs/` holds the current contract, one file per capability: `app-delivery`,
`connection-resilience`, `estimation-rounds`, `game-sessions`, `live-updates`, `room-membership`,
`table-ui`. That is the authority on what the application does.

`openspec/changes/archive/` holds every change that produced it. Each one contains a proposal (why),
a delta specification (what), a design (how, and which alternatives were rejected), a task list, and
the verification results — including checks that were *not* run and why. Reading one end to end is
the fastest way to understand what is expected here.

### How a change flows

A behavioural change is proposed before it is built, and the four artifacts above are written
first. Then the code follows, the tasks are ticked only when the verification each one names has
actually been run, the delta is merged into the main specification, and the change is archived.
The commands are `/opsx:propose`, `/opsx:apply`, `/opsx:sync` and `/opsx:archive`.

Documentation that is not a behaviour contract — the README, the licence, the privacy notes — is
edited directly, without a change proposal.

## Scope discipline (most important rule here)

The owner has stated explicitly: **do not make decisions that were not asked for, and do not add
features that were not explicitly requested.** This outranks any instinct to be helpful. In
practice this means:

- No authentication, no accounts, no user profiles, no persistence of past sessions, no statistics,
  no timers, no spectator roles, no room settings, no "typical" planning-poker extras — unless the
  owner asks for them by name.
- When a requirement is ambiguous, ask rather than pick the "obvious" richer option. A missing
  feature is easy to add later; an unrequested one has to be discovered and removed.
- When a technical need genuinely forces a decision (for example: a reconnecting browser must be
  re-identifiable), name that need, propose the smallest solution, and get agreement before
  building it.

There is one standing tension to be aware of. The owner's global instructions say that values which
govern system behaviour (deadlines, limits, intervals, thresholds, switches) belong in an
administration area under "Settings", changeable at runtime, each with a full-sentence explanation
of what the value does. That instruction and the "no unrequested features" instruction can collide
in this project, because an administration area is itself a large unrequested feature. **Resolution:
when a behaviour-governing value first becomes necessary, do not silently hardcode it and do not
silently build an admin area — surface the value to the owner and let them decide** whether it
becomes a constant, an environment variable, or a managed setting.

## Hard rule: everything is self-contained, nothing is loaded from remote servers

The deployed application must never cause the visitor's browser to fetch anything from a host other
than the one serving the app. No fonts, no images, no icons, no stylesheets, no scripts, no
analytics, no error reporting, no third-party frames. If it travels over the network at runtime, it
comes from this container.

What this rules out in practice, because these are the ways it usually creeps in:

- **Web fonts from a font service** (Google Fonts and the like), whether linked via `<link>` or
  pulled in by an `@import` inside a stylesheet. Any font that is not a font already present on the
  user's system must be committed to this repository as a `.woff2` file, served by this app, and
  declared with a local `@font-face` rule.
- **Anything on a content delivery network** — a `<script src="https://cdn…">` or a stylesheet from
  a CDN. Dependencies are installed with npm and bundled by Vite into the build output.
- **Icons from an icon service or an icon font.** Use inline SVG in the components instead.
- **Remotely hosted images.** Images belong in the repository and are bundled; small ones can be
  inlined by the bundler as data URIs.

Two clarifications so the rule is applied to the right things. Emoji — the coffee cup on the
"pass" card, for example — are rendered from the operating system's own font and cause no network
request, so they are fine. And npm packages are a build-time concern, not a runtime one: installing
them during the Docker build is expected, what matters is that nothing in the *built output* points
at a foreign host.

The consequences worth knowing: this application works on a fully offline network, it leaks nothing
about its users to third parties, and it cannot break because someone else's server is down.

When touching the frontend build, verify the rule rather than assuming it: search the build output
for `http://` and `https://` references and confirm that the browser's network panel shows requests
to one origin only.

### Enforcement: Content-Security-Policy header (approved by the owner)

A Content-Security-Policy header — a response header that tells the browser which origins it is
allowed to load anything from at all — turns the rule above from a convention into something the
browser enforces, so a violation that slipped through review fails visibly instead of silently
phoning home. The owner has approved adding it.

It is set in the Go transport layer as middleware on every response that serves the application
(the HTML document and the static assets). The intended policy:

```
default-src 'self';
connect-src 'self';
img-src 'self' data:;
font-src 'self';
style-src 'self';
script-src 'self';
object-src 'none';
base-uri 'self';
form-action 'self';
frame-ancestors 'none'
```

Why each part is there:

- `default-src 'self'` is the catch-all: any resource type not named below may only come from the
  origin serving the page. The directives after it are not redundant — they exist so that a later
  edit cannot widen one resource type by accident.
- `img-src 'self' data:` allows `data:` because Vite inlines small images directly into the bundle
  as data URIs. A data URI is embedded content, not a network request, so it does not violate the
  self-contained rule.
- `connect-src 'self'` covers the WebSocket. Per CSP Level 3, `'self'` matches same-origin `ws:`
  and `wss:` connections, so no separate entry is needed — but this is exactly the sort of thing to
  **verify in the browser once the socket is wired up**, because older browser behaviour differed
  and a broken policy here breaks the entire app rather than one image.
- `object-src 'none'`, `base-uri 'self'`, `form-action 'self'` and `frame-ancestors 'none'` are
  standard hardening rather than consequences of the self-contained rule: they block plugin
  embedding, stop an injected `<base>` tag from re-pointing every relative URL at a foreign host,
  keep form submissions on this origin, and forbid other sites from embedding this app in a frame.
  They are listed openly here so they can be dropped if the owner does not want them.

Note what is deliberately absent: **no `'unsafe-inline'` and no `'unsafe-eval'`.** Svelte compiles
component styles into real stylesheet files and does not need either. If a CSP violation does show
up in the browser console during development, **fix the code that caused it — never widen the
policy to make the message go away**, because those two values are precisely what makes a CSP worth
having.

One practical caveat for the development setup: the Vite dev server uses inline scripts and its own
WebSocket for hot reloading, which this policy would block. Apply the header in the production
build path, and either relax or omit it in the development path — the same build-tag split that
decides between disk-served and embedded assets.

## The product

A self-hosted Planning Poker application for estimating work items in a team, packaged as a
Docker container. The repository includes local and HTTPS reverse-proxy examples.

The flow, exactly as specified:

1. Calling the base URL shows an entry screen whose only option is **"Start a new game"**.
2. Starting a game generates a **unique URL** which is shared with others to invite them.
3. Before someone can take a seat at the table, they must **enter their name**. The name is stored
   **client-side in a cookie only on explicit opt-in**. A separate seat cookie supports reconnects.
4. Participants sit around the table on wide screens and in a list on narrow screens. The deck
   stays along the bottom edge, and "Invite players" is available regardless of participant count.
5. Voting uses **t-shirt sizes** for now: `XS, S, M, L, XL, ?, ☕`. Other decks (Fibonacci was named
   as the example) come **later** — so the deck must be a named, replaceable value in the domain
   model, but **only the t-shirt deck is implemented now**.
6. Votes stay hidden until any seated participant chooses **"Reveal"**, whether or not everyone
   has voted.
7. **"New round"** starts a fresh voting round and is available before and after reveal.

## Decisions already made

### Namespace and module path

The namespace is `de.thomasnegele.planningpoker`. In Go this becomes the module path in `go.mod`
verbatim:

```
module de.thomasnegele.planningpoker
```

This is a legal Go module path — Go only requires a dot in the first path element and URL-safe
characters — and it is deliberately **not** resolvable over the network. That is fine because this
module is never fetched as a dependency by anything else. The consequence to remember: **never try
to `go get` or `go install` this module path**; it is built from the working tree only.

### Backend: Go

Chosen by the owner. No web framework — the standard library's `net/http` with its pattern-based
routing (`http.ServeMux` with method and wildcard patterns, e.g. `POST /api/games`) is sufficient
for the handful of endpoints this app has, and avoids a dependency that would need maintaining.

### Real-time transport: WebSocket via `github.com/coder/websocket`

Planning poker needs the server to push state to every participant (someone joined, someone voted,
the round was revealed). WebSocket is a connection that stays open in both directions, unlike a
normal HTTP request that ends after one answer.

`github.com/coder/websocket` (previously `nhooyr/websocket`) offers context-aware reads, writes
and pings
and serialises concurrent writes internally. Its `Close` operation uses library deadlines rather
than a caller-supplied context; shutdown must account for that separately. The choice does not
depend on a claim that another WebSocket library is unmaintained.

### Concurrency model: one goroutine owns each room

The established pattern for this kind of application, and the one to follow:

- A **manager** holds a map from room identifier to room, guarded by a `sync.RWMutex` (a lock that
  allows many simultaneous readers but only one writer). The lock protects *the map only* —
  creating, finding and deleting rooms.
- Each **room runs its own single goroutine** which owns that room's entire state (participants,
  their votes, whether the round is revealed). Nothing outside that goroutine ever reads or writes
  room state directly.
- Each connected client has **two goroutines**, one reading from the socket and one writing to it,
  communicating with the room goroutine over **buffered channels**. Buffered means a slow or hung
  client cannot block the room or the other participants; if its buffer overflows, that client is
  dropped rather than allowed to stall everyone.

The payoff is that the interesting logic — who may vote, when a reveal is allowed — runs
single-threaded inside one goroutine and needs no locking at all, so it can be unit-tested as plain
functions on plain data.

### Persistence: none, deliberately

There is **no database**. State lives in memory in the Go process. A planning poker round exists for
the length of a meeting and is worthless afterwards, and nothing in the requirements asks for
history.

The two consequences must be stated to the owner whenever they become relevant rather than
discovered later:

- **Restarting the container destroys all open rooms.** A deployment during a session interrupts
  that session.
- **The app cannot run as more than one instance**, because two containers would not share room
  state. If horizontal scaling is ever wanted, the fix is a shared store (Redis is the natural fit)
  — but that is a change to be requested, not to be pre-built.

Keep this honest: write the room store behind a small interface *only if* it costs nothing, and do
not build a second implementation of it speculatively.

### Room identifiers

Starting a new game generates a room identifier with `crypto/rand`, carrying 128 bits of entropy
in a URL-safe alphabet. Custom identifiers are also intentionally accepted: 5–64 ASCII letters,
digits, hyphens or underscores. They can be easy to guess; there is no additional authentication.
See `game-sessions` for the access rules and reopening behaviour.

### Frontend: Svelte 5 + TypeScript + Vite

Chosen for this app because the whole user interface is one live view driven by a stream of state
from the server. Svelte compiles components to direct DOM updates instead of shipping a runtime
that diffs a virtual DOM, so a minimal app ships roughly 5–15 KB of gzipped JavaScript where an
equivalent React baseline starts around 40 KB and grows quickly once a router and state management
are added.

**No SvelteKit.** SvelteKit is Svelte's meta-framework and its main jobs are server-side rendering
and its own server-side routing — neither of which applies here, because the Go process is the
server and the app is a single client-rendered page. Plain Svelte built by Vite (the build tool
that compiles and bundles the frontend) into static files is the right size of tool.

The client keeps no authoritative state of its own: it renders what the server sends and sends user
intents back. This avoids the classic bug where a participant's screen disagrees with the server
about whether the round was revealed.

### Packaging: one binary, one container

The frontend is built to static files, and those files are compiled **into the Go binary** using
`go:embed` (a standard-library directive that bakes files into the executable at compile time). The
Go process serves them, falling back to `index.html` for any path that is not a real file, which is
what makes client-side routes like `/g/<room-id>` work on a hard page reload.

The Docker image is a multi-stage build — a Node stage builds the frontend, a Go stage compiles the
binary with `CGO_ENABLED=0` (no C dependencies, so the binary is fully static) and `-ldflags="-s -w"`
(strips debug symbols to shrink it) — and the final stage is `gcr.io/distroless/static-debian12:nonroot`,
an image containing no shell and no package manager, running as an unprivileged user. Expect a final
image in the ~15 MB range. The result is a single container exposing a single port.

A reverse proxy sits in front and terminates TLS. **Which one is not this project's business** — an
early version of this file named Caddy and the repository shipped a configuration for it; both were
removed, because whoever self-hosts this already runs a proxy that already manages their
certificates.

What the project does owe them is stated in the README: the proxy must forward the WebSocket
upgrade, must pass through the `Host` header the browser sent, and must set `X-Forwarded-Proto:
https`. The second is the one that costs an evening — the server compares `Origin` against `Host`
and refuses an upgrade whose two disagree, so a proxy that rewrites `Host` to the backend's name
breaks every connection while looking like a broken application.

The container publishes no host port; the proxy reaches it by Docker service name on a shared
network, or through the loopback port that `compose.local.yaml` publishes.

## Intended repository layout

```
cmd/planningpoker/        main package — process startup, configuration, wiring
internal/game/            domain: room, participant, round, deck, the rules. No I/O, no HTTP.
internal/hub/             room manager and the per-room goroutine; owns the channels
internal/transport/       HTTP handlers, routing, WebSocket upgrade, message encoding
internal/webassets/       go:embed of the built frontend + SPA fallback handler
web/                      Svelte frontend source (package.json, vite.config.ts, src/)
Dockerfile                multi-stage build described above
compose*.yaml             the base service, and explicit loopback-only local access
```

`internal/` is a Go convention with teeth: the compiler refuses to let any other module import
packages under it, which keeps the domain from leaking into a public surface by accident.

The rule that matters most for testability: **`internal/game` must not import `net/http` or the
WebSocket library.** The rules of the game are pure functions over data; the transport layer
translates messages into calls on them.

## Commands

Run commands from the repository root unless a different directory is stated. See
CONTRIBUTING.md for the complete verification order, including building assets before embedded tests.

To run the application while working on it, start two processes and open
<http://localhost:5173> — Vite serves the frontend with hot reloading and forwards `/api` and `/ws`
to the Go process on 8080:

```bash
go run ./cmd/planningpoker        # terminal 1
cd web && npm run dev             # terminal 2
```

Note that development mode applies **no** Content-Security-Policy, because Vite's hot reloading
needs inline scripts and a WebSocket of its own. A policy violation therefore only appears once the
production build runs; see the README for the rest of the differences.

```bash
# Backend
go build ./...
go test ./...                                  # all tests
go test ./internal/game -run TestReveal        # one test by name (-run takes a regex)
go test -race ./...                            # race detector — mandatory before any commit
                                               # touching concurrency, since the whole hub design
                                               # rests on state having a single owner
go vet ./...
gofmt -l .                                     # lists misformatted files; empty output means clean

# Frontend (from web/)
npm ci
npm run dev                                    # Vite dev server with hot reload
npm run build                                  # static output that gets embedded into the binary
npm run check                                  # svelte-check: TypeScript + template diagnostics

# Container
docker compose up --build
```

The build-tag split exists and is what keeps a frontend change from requiring a Go rebuild: a
development build reads the frontend from disk, and `-tags embedassets` reads it from a copy
compiled into the binary. See `internal/webassets/source_disk.go` and `source_embed.go`.

## Protocol notes

The client and server exchange JSON messages over one WebSocket per participant. Two principles
worth fixing early:

- **The server sends whole state snapshots** of the room rather than incremental patches, at least
  initially. Rooms hold a handful of participants; a snapshot is tiny, and it makes a reconnecting
  client correct by construction with no replay logic.
- **Hidden votes must never leave the server.** While a round is unrevealed, the snapshot may say
  *that* a participant has voted, but must not contain *what* they voted. Sending the value and
  hiding it in the user interface would let anyone read it in the browser's developer tools, which
  defeats the entire point of simultaneous estimation.

## The questions this file once left open — all decided

They are recorded here because the reasoning is useful, but the **specifications are where each
now lives**, and that is what to read before changing any of them.

1. **Room lifetime.** A grace period, not immediate deletion: a room survives five minutes after
   its last connection closes, adjustable through `PLANNINGPOKER_ROOM_GRACE_PERIOD`. While anybody
   is connected it never expires. See `game-sessions`.
2. **Reconnect identity.** Confirmed and built: a per-room seat token in a cookie, separate from
   the public participant identifier, never sent to any client but its own. It is a session cookie
   without an explicit expiry. Browser session restoration can preserve it across restarts. See
   `game-sessions` for the intended session scope.
3. **Reveal permissions.** Any participant may reveal and any participant may start a new round.
   There is no host and no creator role anywhere in this product. See `estimation-rounds`.
4. **Reveal before everyone has voted.** Allowed at any time. The interface reports whether
   everyone present has voted, but that is a statement about the round, not a permission.
5. **Result presentation.** Each participant's card, plus a count per card. **No average, median or
   any other arithmetic** — t-shirt sizes are an ordered scale without arithmetic, and `?` and the
   coffee cup are not sizes at all. See `estimation-rounds`.
6. **Name changes and duplicates.** A seated participant may change their own name and nobody
   else's. Duplicate names are explicitly allowed, because identity is carried by an opaque
   identifier and a small table resolves a collision socially. See `room-membership` and
   `table-ui`.

**When a new question of this kind appears, the rule from "Scope discipline" still applies: name
it, propose the smallest answer, and get agreement before building it.** That is especially true of
values that govern behaviour — deadlines, limits, intervals, thresholds. Several exist now
(capacity ceilings, the heartbeat, the shutdown budget), and each was put to the owner rather than
chosen quietly. Two of them deliberately became constants rather than settings, because they follow
from how browsers and proxies behave rather than from the size of the machine; that reasoning is
written beside them in the code.
