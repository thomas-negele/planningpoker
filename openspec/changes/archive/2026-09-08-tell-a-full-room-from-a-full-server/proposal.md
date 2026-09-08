## Why

Two different things are refused with the same word. When the process is holding as many rooms as
it may, and when a single room is holding as many connections as it may, the page is told
`at_capacity` and shows the same sentence: that the server is running as many games as it can.

On an otherwise empty server whose one room has too many tabs open, that sentence is simply untrue,
and it sends the person off to wait for a busy server that is not busy. The two situations also call
for opposite responses: waiting helps with a full server, while a full room is usually somebody's own
second and third tab, which they can close now.

## What Changes

- Give the per-room connection ceiling its own refusal reason, distinct from the process running out
  of rooms.
- Say the right thing for each: that the server is full and will free up, or that this room already
  has as many connections open as it allows and closing a spare tab will help.
- State in the specification that the two must be distinguishable, so this cannot quietly collapse
  back into one code.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `live-updates`: The list of refusal reasons that must be distinguishable gains the room's
  connection ceiling as its own entry.
- `app-delivery`: The requirement that bounds rooms and connections says that reaching one ceiling
  is distinguishable from reaching the other, not merely from a missing room or a fault.

## Impact

`internal/hub/manager.go` (a second sentinel error), `internal/transport/protocol.go` (a second
code), `internal/transport/rooms.go` (which error the attach refusal carries), and on the frontend
`protocol.ts`, `connection.svelte.ts` and `ServerFull.svelte`, which learns to say which of the two
happened.

No change to any limit, to when a refusal happens, or to what the game does. This is about naming a
refusal correctly, and it was found while reviewing the documentation rather than by a user.
