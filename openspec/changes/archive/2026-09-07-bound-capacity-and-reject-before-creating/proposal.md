## Why

Nothing in the running process has an upper bound. Rooms, connections and participants are created
on demand with no ceiling, and a connection may send messages as fast as it can write them. A room
whose socket is held open never expires, so the five-minute grace period does not limit an
accumulating client at all. Separately, a plain HTTP request to a room's socket path creates a room
before the upgrade is attempted, so a request that never becomes a WebSocket connection — including
one refused because it came from another website — still leaves a room and a seat cookie behind.

## What Changes

- Add four capacity limits, each an environment variable with a documented default: rooms in the
  process, connections per room, participants per room, and message rate per connection.
- Refuse work above a limit in a way the page can explain, without disturbing rooms already running
  and without ending a meeting that is merely long.
- Create a room only once a connection has actually been accepted, and issue a seat cookie only
  then, so a refused upgrade or a foreign origin leaves nothing behind.
- Protect `POST /api/games` against creation triggered from another website.
- Set the incoming message size limit explicitly to a value this protocol needs, rather than relying
  on the WebSocket library's default.
- Handle invalid percent-encoding in the room route so a malformed link reaches the existing
  "not a game" screen instead of breaking the page before it renders.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `app-delivery`: The environment-variable requirement gains the new capacity values and stops
  claiming there are only two such values; a new requirement bounds rooms and connections in the
  process and defines what a refusal looks like.
- `game-sessions`: Opening a room's URL creates a room only once a connection is accepted, so a
  refused upgrade creates nothing; a malformed URL still reaches the "not a game" outcome.
- `room-membership`: A room holds a bounded number of participants, and a seat refused for capacity
  is refused with its own reason.
- `live-updates`: A connection has a bounded message rate and a bounded message size, and exceeding
  either is answered without affecting anyone else at the table.

## Impact

`internal/hub/manager.go` (room ceiling), `internal/hub/room.go` (connection and participant
ceilings), `internal/game/room.go` (participant ceiling in the rules), `internal/transport/rooms.go`
(order of room creation, cookie issuance, origin check, read limit, message rate),
`internal/transport/protocol.go` (new refusal reasons) and `web/src/lib/router.svelte.ts` (decoding).
`compose.yaml` and the README gain the four new settings with their explanations.

No new dependency, no protocol reshape, no change to how a game is played. Covers checklist items
S1, S2 and S5, and the message-size and rejection parts of S6. Deliberately out of scope: write
deadlines and heartbeats (S3) and the shutdown budget (B1), which are the next package, and any
form of login, role or persistence.
