## Why

Two halves of the application exist and cannot reach each other. The delivery machinery serves a
placeholder page and holds a WebSocket open that echoes whatever it is sent. The rules of the game
are complete and tested but nothing calls them: no route creates a room, no message casts a vote,
and a room exists only for as long as one test function runs.

This change is the joint between them. It gives rooms somewhere to live and a way for browsers to
act on them, and it is where the two guarantees that have been designed for but not yet exercised
finally have to hold: that a hidden vote never crosses the network, and that room state has exactly
one owner even though many browsers are pushing at it simultaneously.

It also settles the question deferred since the first change — when an unused room is discarded —
because that is the first behaviour in this product that depends on the passing of time.

## What Changes

- **New `internal/hub` package** holding the rooms. A manager owns a map from room identifier to
  room, guarded by a lock that protects the map and nothing else. Each room runs its own goroutine
  which owns that room's entire state; no other goroutine reads or writes it. Browsers reach a room
  by sending it messages over a channel and receive updates the same way.
- **Creating a game** over HTTP. A request to create one returns the identifier of a fresh room,
  from which the client builds the invitation URL.
- **The real WebSocket protocol**, replacing the echo. A browser connects to a specific room and
  exchanges JSON messages: it sends intents — take a seat, play a card, reveal, start a new round,
  change my name — and receives whole snapshots of the room after every change, plus a reply when
  an intent is refused.
- **BREAKING**: **the temporary echo endpoint is deleted**, along with the placeholder page's probe
  of it. Both existed only to prove the transport path before anything depended on it, and both have
  done that job. The placeholder page is adjusted so the deployed application stays coherent rather
  than displaying a failed connection until the table is built in the next change.
- **A participant identity cookie per room**, opaque and scoped to that room's path, so a reloading
  browser is reseated in the same chair with the same vote and two rooms cannot tell they are
  looking at the same person.
- **Several connections per participant.** Two tabs of the same room are one seat at the table with
  two open sockets, both kept up to date. A participant is marked away only when the last of their
  connections drops.
- **Room lifetime**, the first behaviour here that depends on a clock: a room with nobody connected
  is discarded after a grace period, whether everyone has left or nobody ever arrived. The grace
  period is an environment variable with a documented default of five minutes, alongside the two
  values that already exist.
- **A visitor arriving at a room that no longer exists is told so**, and offered the one action that
  helps, rather than being left looking at a table that never loads.
- **Shutdown now closes rooms**, which is where the connection-tracking shape built for the echo
  probe gets replaced by the real thing.

## Capabilities

### New Capabilities

- `game-sessions`: a game as something that exists over time — creating one and getting an
  invitation URL, how long a room outlives the people in it, what a visitor sees when a room has
  gone, and the per-room cookie that lets a browser be recognised as the same participant.
- `live-updates`: how a browser and the server talk while a game is running — connecting to a
  specific room, the intents a client may send, the snapshots the server sends back after every
  change, what happens when an intent is refused, and how several connections belonging to one
  person behave.

### Modified Capabilities

- `app-delivery`: the requirement "A WebSocket connection can be established" currently describes an
  endpoint that echoes messages back and carries no application meaning. That endpoint is being
  deleted, so the requirement is rewritten to describe a connection into a specific room. Everything
  it guarantees about the transport — that the upgrade succeeds directly and through a
  TLS-terminating reverse proxy, and that connections close without leaking — is kept, because those
  are the properties the next change and every change after it depend on.

## Impact

- **Created code**: `internal/hub/` with its tests.
- **Modified code**: `internal/transport/` gains the room creation handler, the room WebSocket
  handler, the message encoding and the cookie handling, and loses `echo_probe.go` entirely.
  `cmd/planningpoker/` gains the grace period setting and closes the hub on shutdown instead of the
  probe. `web/` loses the connection probe from its placeholder page.
- **No new dependencies.** The WebSocket library and the standard library cover all of it;
  `encoding/json` is standard library and belongs to the transport layer, which is precisely why the
  domain was forbidden from importing it.
- **`internal/game` is not modified.** If this change finds itself wanting to alter a rule, that is
  a signal to stop and reconsider rather than to reach into the domain: the rules were agreed and
  specified separately on purpose.
- **After this change the product works without a user interface.** A room can be created, joined,
  voted in, revealed and revoted — over a WebSocket, by a script. What is still missing is the table
  itself, which is the next and last change of the MVP.
- **Two consequences worth restating**, because they become real here rather than theoretical:
  restarting the container destroys every open room, since nothing is persisted; and the application
  cannot run as more than one instance, because two containers would not share room state.
