## Why

Visitors cannot tell which application version is running. A visible version on the start page makes that clear, while a small, deliberate version step in the change process keeps the number meaningful.

## What Changes

- Show the built application's `major.minor.patch` version on the start page only, starting at `0.3.0`.
- Increase the version once for each functional change before it is merged into `main`. A rebuild of the same commit keeps the same version; documentation-only changes need no increase.
- Add a required version decision to the OpenSpec/contribution process: Codex proposes major, minor or patch, and the user decides before the version is changed.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `app-delivery`: A build carries the version recorded for its source revision, including on repeated builds.
- `table-ui`: The start page shows the version without adding it to room or other screens.

## Impact

The frontend version metadata, frontend build, start-page UI and `CONTRIBUTING.md` are affected. No API, runtime configuration or new dependency is required.
