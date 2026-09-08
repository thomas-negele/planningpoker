## Why

The current table places Reveal above the status and results, keeping its space when hidden. The owner wants to retain that arrangement; the existing requirement for exactly equal space above and below the results unnecessarily constrains it.

## What Changes

- Allow the results and reserved action area to form one visually balanced group on the table.
- Keep horizontal centering, legibility, containment and a stable table and seat layout on reveal.
- Remove the requirement for exact vertical centering of the results independently of the action area.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `table-ui`: Relax only the vertical placement constraint in the results requirement.

## Impact

The main specification and the stale layout comment in `RoomView.svelte`. The existing UI arrangement is retained. No backend, protocol, dependency or deployment changes.
