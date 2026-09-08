## 1. Compose deployment paths

- [x] 1.1 Add the app runtime restrictions and adjustable resource/log defaults to the common Compose service; verify them in the combined configuration/inspection batch in 3.1–3.2.
- [x] 1.2 Add explicit local and server overlays plus the small proxy configuration described in design.md; verify loopback-only local publishing, unpublished server app port and proxy routing in 3.1–3.2.
- [x] 1.3 Correct misleading Docker pinning comments and exclude non-build local metadata/secrets from the build context; verify the diff and successful fresh image build in 3.2 without altering tracked history.

## 2. Direct documentation alongside the change

- [x] 2.1 Write README local/server quickstarts, configuration and update notes directly (no new documentation spec); verify each command against the resolved configurations and smoke results from section 3. Document open room links and restart data loss without operator-specific details or a public-readiness claim.

## 3. One bundled verification

- [x] 3.1 Once all files are ready, validate base/local/server Compose configurations together, including one set of resource overrides; inspect bindings, privilege restrictions, limits and bounded logging. Do not add tests that merely mirror YAML.
- [x] 3.2 Build the image once and run a small local/proxy smoke covering frontend, room creation, WebSocket seat/vote/reveal, effective container restrictions and stop. Validate proxy syntax; use an isolated local TLS route where practical and record public DNS/certificate checks separately if unavailable. Do not rerun unrelated Go tests for configuration-only edits. If a required check cannot run, record it and leave its task open.
- [x] 3.3 Run strict OpenSpec validation and diff checks; record the exact results and any pending checks. After review, sync/archive this change through the corresponding OpenSpec workflow; only then mark the package complete.

## Verification results (2026-09-07)

All checks below were run against the working tree in this session. Public DNS and a
publicly trusted certificate were **not** checked and remain deployment-time work;
see the note at the end.

**3.1 — resolved Compose configurations.** `docker compose config` on base,
`compose.local.yaml` and `compose.server.yaml`:

- Base publishes no host port (`ports` absent, `expose: ["8080"]`).
- Local publishes `127.0.0.1:8080->8080/tcp` only — bound to loopback, not `0.0.0.0`.
- Server keeps the app unpublished on the `app_internal` network, which resolves with
  `internal: true`; only Caddy publishes 80/tcp, 443/tcp and 443/udp, on both
  `app_internal` and `proxy_public`.
- All three keep `read_only: true`, `cap_drop: [ALL]`,
  `security_opt: [no-new-privileges:true]`, `mem_limit` 268435456, `cpus` 1,
  `pids_limit` 128 and `json-file` logging with `max-size: 10m`, `max-file: 3`.
- Overrides applied together (`PLANNINGPOKER_MEMORY_LIMIT=384m`, `CPU_LIMIT=1.5`,
  `PID_LIMIT=200`, `LOG_MAX_SIZE=5m`, `LOG_MAX_FILES=2`, `HTTP_PORT=8081`) resolve to
  402653184 / 1.5 / 200 / 5m / 2 and `127.0.0.1:8081`, with every restriction intact.
- The server configuration fails closed without a domain: exit 1 with
  `required variable PLANNINGPOKER_DOMAIN is missing a value`.

**3.2 — build, running container and smoke.**

- Fresh image build with `--no-cache` succeeded; final image 15.6 MB.
- `docker inspect` on the running container: `ReadonlyRootfs: true`, `CapDrop: [ALL]`,
  `CapAdd: null`, `SecurityOpt: [no-new-privileges]`, `Memory: 268435456`,
  `NanoCpus: 1000000000`, `PidsLimit: 128`, `Privileged: false`, `Binds: null`,
  log config `max-size 10m` / `max-file 3`. `docker top` shows UID 65532 running
  `/planningpoker`.
- Local HTTP: `GET /` returns 200 with the application document, the expected
  `Content-Security-Policy` response header, and no `http(s)://` reference to any
  foreign host in the served document.
- Local WebSocket smoke (throwaway Go client, two participants, deleted afterwards):
  room creation over `POST /api/games` returns a 26-character identifier; the t-shirt
  deck arrives as `XS S M L XL ? ☕`; two distinct seats; after both votes the snapshot
  is still unrevealed and carries **no** results object while both participants show
  as having voted; reveal returns each card correctly plus a two-entry tally; a new
  round hides the results and clears both votes.
- Proxy route over an isolated local TLS route (`PLANNINGPOKER_DOMAIN=localhost`, so
  Caddy uses its internal CA, issuer `Caddy Local Authority - ECC Intermediate`):
  `https://localhost/` returns 200, `http://localhost/` returns a 308 redirect to
  HTTPS, the app container has empty `PortBindings` and is on `app_internal` only, and
  a direct request to `127.0.0.1:8080` is refused. The full seat/vote/reveal/new-round
  smoke passed unchanged over `wss://`.
- Seat cookie through the proxy:
  `Path=/ws/<room>; Max-Age=43200; HttpOnly; Secure; SameSite=Lax` — the `Secure`
  attribute confirms Caddy's forwarded-protocol information reaches the application.
- Reconnect through the proxy with a cookie jar: the same seat identifier is returned,
  no duplicate participant is created, and name, vote and non-away status survive.
- `caddy validate` on the supplied Caddyfile with a domain set: `Valid configuration`.
- Container logs contained only `listening on [::]:8080` — no write failures under the
  read-only root filesystem. `docker compose down` completed in well under a second.
- No Go or Svelte test suites were rerun: this package changed configuration and
  documentation only.

**Not checked here.** Public DNS resolution and a publicly trusted certificate need a
real domain and reachable ports 80/443, so they remain deployment-time checks. Nothing
in this package makes the application ready for unattended public exposure; the abuse,
capacity and shutdown findings from the review are still open.
