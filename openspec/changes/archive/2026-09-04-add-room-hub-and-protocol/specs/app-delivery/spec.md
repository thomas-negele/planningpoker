## MODIFIED Requirements

### Requirement: A WebSocket connection can be established

The server SHALL accept a WebSocket upgrade request at a path naming a specific room, and hold the
connection open. Establishing the connection MUST succeed both when the server is reached directly
and when it is reached through a reverse proxy that terminates TLS.

The endpoint that echoed messages back is gone. It existed only to prove the transport path before
any game logic depended on it, and it has done that: what the connection now carries is defined by
the `live-updates` capability.

What this requirement continues to guarantee is the transport itself — that the upgrade succeeds in
both deployments, and that a closed connection releases everything holding it open. Every later
change rests on those two properties, and neither is obvious enough to leave untested.

#### Scenario: Upgrade succeeds directly

- **WHEN** a browser opens a WebSocket to a room on the application's own origin
- **THEN** the connection reaches the open state and the server begins sending that room's state

#### Scenario: Upgrade succeeds through a TLS-terminating reverse proxy

- **WHEN** the application is reached through the reverse proxy over `https`, so that the browser
  opens a `wss` connection
- **THEN** the connection reaches the open state and is not blocked by the Content-Security-Policy

#### Scenario: Connection closes cleanly

- **WHEN** the client closes the WebSocket, or the browser tab is closed
- **THEN** the server releases the connection and the goroutines serving it, without leaking them
