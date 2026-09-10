# Planning Poker

Self-hosted Planning Poker with Go, Svelte and WebSockets. One container, in-memory
state, no database. The repository also demonstrates specification-driven development
with [OpenSpec](openspec/specs/).

## Features and limits

- Choose the default t-shirt deck (`XS S M L XL ? ☕`) or Fibonacci
  (`0 ½ 1 2 3 5 8 13 21 ? ☕`), then share the room URL and vote together.
- Votes stay hidden until a participant reveals them. Results show counts per card.
- Any seated participant can reveal or start a new round.
- No accounts or passwords: anyone who knows or guesses a room URL can join.
- Custom room IDs: 5–64 ASCII letters, digits, hyphens or underscores, e.g. `/g/team-alpha`.
- Names and room IDs should contain no confidential information.
- One instance only. Restarting it deletes every room and vote.
- Empty rooms expire after five minutes by default. Connected rooms have no inactivity timeout.
- Opening an expired room URL creates a new empty room at the same address.

## Run with Docker

Requires Docker with Compose. Run from the repository root:

```sh
docker compose -f compose.yaml -f compose.local.yaml up --build -d
```

Open <http://127.0.0.1:8080>. The port is published only on IPv4 loopback.
Set `PLANNINGPOKER_HTTP_PORT` to use another host port.

```sh
# Logs
docker compose -f compose.yaml -f compose.local.yaml logs --tail=100

# Stop
docker compose -f compose.yaml -f compose.local.yaml down
```

The image runs as nonroot with a read-only filesystem, dropped capabilities and
resource limits. No prebuilt image is supplied; Compose builds it from source.

## Reverse proxy

Bring your own reverse proxy and TLS certificates. Configure DNS for its public
hostname and allow HTTPS traffic to the proxy. The proxy must:

- Forward HTTP and WebSocket upgrades, including `/ws/`.
- Preserve the browser-facing `Host` header so it matches the request's `Origin`.
- Set `X-Forwarded-Proto: https` for HTTPS visitors so seat cookies receive `Secure`.

A containerised proxy can join the application's Docker network and forward to
`planningpoker:8080`; use `compose.yaml` alone to avoid publishing a host port.
A proxy running on the host can use the local configuration and forward to
`127.0.0.1:8080`. A proxy container's loopback address refers to that container.

Verify room connectivity and the cookie's `Secure` flag after deployment. Host
security, proxy configuration and certificate management belong to the installation.

## Configuration

[compose.yaml](compose.yaml) documents all settings and how to change them.
Application defaults:

| Setting | Default |
| --- | --- |
| Listen address | `:8080` |
| Shutdown budget | `5s` |
| Empty-room grace period | `5m` |
| Rooms per process | `50` |
| Connections per room | `40` |
| Participants per room, including away seats | `20` |
| Messages per second per connection | `10`, burst `20` |

Container defaults: 256 MiB memory, 1 CPU, 128 PIDs and three 10 MiB log files.
These are starting budgets; capacity increases may require more resources.

### Optional legal pages

Off unless configured. Put your own `privacy.html` and `imprint.html` in a directory,
point `PLANNINGPOKER_LEGAL_DIR` at it, and they are served at `/legal/privacy` and
`/legal/imprint` and linked from the footer. Use `/legal/style.css` for styling; the
pages may load nothing from another host. In Compose:

```yaml
environment:
  PLANNINGPOKER_LEGAL_DIR: "/legal"
volumes:
  - ./legal:/legal:ro
```

Both files are read at startup, so editing them needs a restart — which ends running
games — but no rebuild. A missing or unreadable file stops the process with a message
naming it. Keep completed documents out of Git; `legal/` is already ignored.

## Development

Requires Go matching [go.mod](go.mod) and Node.js with npm. The Docker build uses
Go 1.27 and Node 26. From the repository root:

```sh
(cd web && npm ci)

# Terminal 1
go run ./cmd/planningpoker

# Terminal 2
cd web && npm run dev
```

Open <http://localhost:5173>. Vite provides hot reloading and proxies `/api`, `/ws` and `/legal`
to Go on port 8080. No `npm run build` is needed. Development omits the production CSP;
the container embeds the built frontend and applies it.

## Updates

Update outside meetings; rebuilding and restarting ends active games:

```sh
git pull
docker compose -f compose.yaml -f compose.local.yaml build --pull
docker compose -f compose.yaml -f compose.local.yaml up -d
```

Use the same Compose files as at startup. To roll back, check out the previous commit
and rebuild. Build the image for the architecture of the target host.

## Code and specifications

| Path | Responsibility |
| --- | --- |
| `cmd/planningpoker/` | Startup and configuration |
| `internal/game/` | Game rules, no I/O |
| `internal/hub/` | Room lifecycle and concurrency |
| `internal/transport/` | HTTP and WebSocket protocol |
| `internal/legal/` | Optional operator legal pages |
| `internal/webassets/` | Frontend assets and routing fallback |
| `web/` | Svelte frontend |
| [openspec/specs/](openspec/specs/) | Behaviour contracts |
| [openspec/changes/](openspec/changes/) | Active and archived changes |

See [CONTRIBUTING.md](CONTRIBUTING.md) for the workflow and checks.

## Privacy, security and licence

- [PRIVACY.md](PRIVACY.md): stored data, cookies and logging.
- [SECURITY.md](SECURITY.md): private vulnerability reports. Use issues for ordinary bugs.
- [LICENSE](LICENSE): MIT. OpenSpec-generated files under `.agents/skills/` and
  `.claude/commands/` retain their licence metadata and attribution.
