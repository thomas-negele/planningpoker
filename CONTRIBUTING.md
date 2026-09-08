# Contributing

## Workflow

Use [OpenSpec](https://github.com/Fission-AI/OpenSpec) for behaviour changes:

1. Write a proposal, delta specifications and tasks. Add a design document when
   implementation decisions need explanation.
2. Implement the agreed change and run the relevant checks.
3. Record verification results, merge the delta into the main specs and archive the change.

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
(cd web && npm ci && npm run check && npm run build)
go build ./...
go vet ./...
gofmt -l .
go test -race -count=3 ./...
go test -race -count=3 -tags embedassets ./...
```

`gofmt -l .` should produce no output. The Go suites repeat three times to detect
intermittent failures. Bundle checks at the end of related changes.

For container, serving or CSP-related changes, also verify the application in Docker:

```sh
docker compose -f compose.yaml -f compose.local.yaml up --build -d
```

Development mode does not apply the production CSP.

## Reports

Open an issue for bugs. Report vulnerabilities through [SECURITY.md](SECURITY.md).
