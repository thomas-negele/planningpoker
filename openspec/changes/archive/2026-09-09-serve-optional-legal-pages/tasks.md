## 1. Implementation

- [x] 1.1 Add `PLANNINGPOKER_LEGAL_DIR` and an `internal/legal` package that loads both documents at startup, refusing anything unusable.
- [x] 1.2 Serve the two notices, the stylesheet and `GET /api/legal`; reserve `/legal` in the router so a missing notice is a 404 and not the app document.
- [x] 1.3 Add the footer, the availability fetch and the Vite `/legal` proxy.

## 2. Verification

- [x] 2.1 Run the CONTRIBUTING.md checks and exercise the feature on and off in Docker.

## Verification results (2026-09-09)

`go build`, `go vet`, `gofmt -l .` clean; `npm run check` and `npm run build` pass;
`go test -race -count=3 ./...` passes in both the development and the `embedassets`
build.

`TestAFullTableVotingAtOnceIsNeverSlowed` failed intermittently under `-race`. It was
reproduced on unmodified `HEAD` in a separate worktree, so it is pre-existing and was
left alone.

In Docker, with no legal directory: `/api/legal` reports disabled and `/legal/*` returns
404 rather than the application document. With one mounted read-only: both notices are
served with `no-store` and the stricter policy, HEAD returns the headers without a body,
unknown paths under `/legal` stay 404, and removing a document stops startup with an
error naming the setting and the file. Editing a document and restarting published the
new text with no rebuild; recreating the container without the setting returned the
notices to 404. Through Vite the same routes worked with no frontend rebuild.

The browser check — footer on entry and room screens, keyboard focus, narrow layout, a
notice opening without dropping the room connection, and no request leaving the origin —
was confirmed by the owner, not by the assistant.

**Not checked.** No completed operator documents ship here, and nothing about whether a
given notice satisfies a legal obligation is established by any check in this change.
