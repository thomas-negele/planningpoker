## Why

An installation used by other people may need a privacy notice and an imprint. There
is nowhere to put them, and the operator's details do not belong in this repository.

## What Changes

- `PLANNINGPOKER_LEGAL_DIR` names a directory holding `privacy.html` and `imprint.html`.
  Unset — the default — and nothing is published.
- Both are read at startup and served at `/legal/privacy` and `/legal/imprint`, with a
  built-in `/legal/style.css` and a stricter policy of their own. The footer links them.
- `GET /api/legal` tells the frontend whether they exist.

## Capabilities

### New Capabilities

- `operator-legal-pages`: serving optional operator-supplied notices.

### Modified Capabilities

- `app-delivery`: `/legal` is reserved, so a missing notice is a 404 rather than the
  application's own document.

## Impact

Startup, the router, a new `internal/legal` package, the app footer, the Vite proxy
and a README section. No dependencies, no game-rule changes.
