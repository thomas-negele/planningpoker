# Contributing

## Workflow

Use [OpenSpec](https://github.com/Fission-AI/OpenSpec) for behaviour changes:

1. Write a proposal, delta specifications and tasks. Add a design document when
   implementation decisions need explanation.
2. For each functional change, propose a major, minor or patch increase. The user
   chooses the step. Before merging into `main`, apply that choice in `web/` with
   `npm version <major|minor|patch> --no-git-tag-version` and check that both
   `package.json` and `package-lock.json` contain the new version. Rebuilding the
   same source does not increase it again. Documentation-only changes need no bump.
3. Implement the agreed change and run the relevant checks.
4. Record verification results, merge the delta into the main specs and archive the change.

Documentation and comment-only corrections do not require a behaviour change.

- [Current specifications](openspec/specs/)
- [Archived changes](openspec/changes/archive/)
- [Example: heartbeat and shutdown](openspec/changes/archive/2026-09-07-detect-dead-connections-and-enforce-shutdown/)

## Conventions

- Keep game rules independent of HTTP and WebSocket code.
- Each room's goroutine owns its state; components use the central frontend connection.
- Write concise comments for constraints and non-obvious decisions. Keep detailed
  reasoning in design documents.
- Load runtime resources from the application origin. Do not weaken the production CSP.
- Agree on changes to scope, limits and deadlines before implementing them.
- Record only checks that were actually run.

## Development

See [README.md](README.md) for Docker and development startup.

## Verification

Run from the repository root. Build frontend assets before testing the embedded variant:

```sh
(cd web && npm ci && npm test && npm run check && npm run build)
go build ./...
go vet ./...
gofmt -l .
go test -race -count=3 ./...
go test -race -count=3 -tags embedassets ./...
```

`gofmt -l .` should produce no output. The Go suites repeat three times to detect
intermittent failures. Bundle checks at the end of related changes.

For the throw picker browser interaction checks, install the separate test dependencies with
`cd e2e && npm ci`, then install Playwright's Chromium once with
`cd e2e && npx playwright install chromium`. Run `cd e2e && npm test` after the web dependencies
are installed. That command builds the frontend and starts its own Go server on
`127.0.0.1:4324`; it does not use the Docker demo server on port 8080. Run
`cd e2e && npm run check` to type-check the browser tests.

For container, serving or CSP-related changes, also verify the application in Docker:

```sh
docker compose -f compose.yaml -f compose.local.yaml up --build -d
```

Development mode does not apply the production CSP.

## Reports

Open an issue for bugs. Report vulnerabilities through [SECURITY.md](SECURITY.md).
