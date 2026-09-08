## 1. Removing what the project should not own

- [x] 1.1 Delete `compose.server.yaml` and `Caddyfile`; verify that no file outside `openspec/changes/archive/` mentions either, nor `PLANNINGPOKER_DOMAIN`, nor Caddy.

## 2. Saying what a proxy must do

- [x] 2.1 Replace the README's HTTPS walkthrough with what any reverse proxy must do — forward the WebSocket upgrade, preserve the browser's `Host`, set `X-Forwarded-Proto: https` — each with the consequence of getting it wrong, and with how a proxy reaches the container (shared Docker network, or the loopback-published port).
- [x] 2.2 Add the Nginx Proxy Manager walkthrough as the worked example: the settings to make, and the two checks the operator runs afterwards. State no claim about which headers that product generates.
- [x] 2.3 Rewrite the README's update and rollback commands, which currently carry `PLANNINGPOKER_DOMAIN` and the deleted server file, and adjust the sentence in `PRIVACY.md` that refers to the Caddyfile.

## 3. Verification

- [x] 3.1 Re-establish the two facts the documentation asserts, against the running server: an upgrade with a rewritten `Host` is refused while the same upgrade with `Host` preserved succeeds, and `X-Forwarded-Proto: https` is what decides whether the seat cookie carries `Secure`. Record the observed responses.
- [x] 3.2 Run every command the README now contains, from a clean state, and confirm each does what it says.
- [x] 3.3 Run one batch: `go test -race -count=3 ./...`, `go vet ./...`, `gofmt -l .`, `npm run check`, `npm run build`, and resolve the base and local Compose configurations.
- [x] 3.4 Run strict OpenSpec validation and `git diff --check`; then sync and archive.

## Verification results (2026-09-08)

**3.1 — the two facts the README now asserts**, probed against the running server:

| Request | Response |
| --- | --- |
| `Origin: https://poker.example.com` + `Host: planningpoker:8080` (a proxy that rewrote the host) | `426 Upgrade Required` |
| the same with `Host: poker.example.com` (host preserved) | `101 Switching Protocols` |
| upgrade without `X-Forwarded-Proto` | seat cookie **without** `Secure` |
| upgrade with `X-Forwarded-Proto: https` | seat cookie **with** `Secure` |

Both claims therefore rest on observed responses rather than on reading the code.

**3.2 — every command the README contains was run.** The local Docker start builds
and answers 200 on `127.0.0.1:8080`; the logs command prints
`listening on [::]:8080`; the base configuration resolves with `ports: None`, so the
port really is unpublished; the stop command removes the container and its network.
The development path — two processes, port 5173 — was exercised separately and works,
including the WebSocket through the Vite proxy.

**3.3 — the bundled batch.** `go test -race -count=3 ./...` (three passes),
`go vet ./...`, `gofmt -l .` (empty), `npm run check` (175 files, no errors or
warnings), `npm run build`, and both Compose configurations resolve.

**Also removed, found while sweeping for leftovers.** `CLAUDE.md` still presented
Caddy as the architecture ("On the VPS, Caddy sits in front…") and its repository
layout still listed an HTTPS deployment example. Both now describe a proxy the
operator brings, and name the three requirements. A full-text search for `caddy`,
`compose.server` and `PLANNINGPOKER_DOMAIN` finds nothing outside this change's own
artifacts and the archive.

**Not checked here.** Nginx Proxy Manager itself. The walkthrough deliberately names
only its interface and gives two checks the operator performs on their own
installation, because what headers that product generates is a claim about somebody
else's software that this project cannot keep true. The nginx idle-timeout figure of
sixty seconds is quoted from its documented default, not measured here.
