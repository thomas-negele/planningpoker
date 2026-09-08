## MODIFIED Requirements

### Requirement: The process shuts down gracefully

On receiving `SIGTERM` or `SIGINT` — the signals sent when a container is stopped or a developer
interrupts the process — the server SHALL stop accepting new connections, close the connections it
holds, and exit with status 0 within a bounded time. If shutdown does not complete within that
time, the process SHALL exit anyway rather than hang.

The configured shutdown budget SHALL be **one deadline covering the whole shutdown**, not a budget
for one part of it. It starts when the stop signal arrives and applies to ordinary HTTP requests and
to WebSocket connections together. When it expires, whatever is still open SHALL be closed by force
rather than waited for, so that the process exits comfortably inside the grace period the container
runtime allows before it kills the process outright.

A client that has stopped reading SHALL NOT be able to delay shutdown. Neither shall one that never
answers the closing handshake: the budget is what bounds the wait, and it is enforced rather than
hoped for.

Once shutting down has begun, a new connection SHALL be refused rather than accepted, so that
cleanup does not race against arrivals.

#### Scenario: Container stop is clean

- **WHEN** the container is stopped and the process receives `SIGTERM`
- **THEN** the process stops accepting new connections, closes open WebSocket connections, and
  exits with status 0 without the container runtime having to kill it

#### Scenario: A client that has stopped reading does not delay shutdown

- **WHEN** the process is stopped while a connected client is no longer reading from its socket
- **THEN** that connection is closed by force once the budget expires, and the process still exits
  with status 0 within the budget rather than being killed by the container runtime

#### Scenario: The budget covers everything, not only HTTP

- **WHEN** the process is stopped while both an ordinary HTTP request and WebSocket connections are
  in flight
- **THEN** one deadline measured from the stop signal governs both, and the process does not spend
  the budget on one and then wait indefinitely on the other

#### Scenario: A connection arriving during shutdown is refused

- **WHEN** a new connection is attempted after the stop signal has been received
- **THEN** it is refused rather than accepted, and it neither creates a room nor delays the exit

## ADDED Requirements

### Requirement: A connection that has stopped answering is detected and released

The server SHALL send a periodic heartbeat on every open connection and SHALL close a connection
that does not answer within a bounded time. Every outgoing message SHALL have its own deadline, so
that a client which has stopped reading cannot hold a goroutine indefinitely.

This exists because a connection can die without saying so: a lost mobile signal, a suspended
laptop, or a proxy that drops a connection without sending a close frame. Nothing arrives to read
and nothing arrives to fail on, so without a heartbeat such a connection stays open in the server's
eyes forever — keeping a seat occupied and keeping the room from ever expiring.

**The heartbeat SHALL NOT be an inactivity timeout and SHALL NOT be turned into one.** It measures
whether the network path is alive, never whether anybody is doing anything. A browser answers it
automatically, with the page untouched and nobody at the keyboard, so a meeting in which people
think for an hour without voting is unaffected. No rule anywhere may end a connection, a seat or a
room on the grounds that no game action has arrived.

Closing a connection for either reason SHALL be indistinguishable, to everyone else at the table,
from any other dropped connection: the participant is marked away, keeps their seat, their name and
their vote, and is reseated on their return.

#### Scenario: A silently dead connection becomes an away participant

- **WHEN** a participant's connection stops answering entirely, without the socket being closed
- **THEN** the server closes it within a bounded time and that participant is shown as away, exactly
  as after any other dropped connection

#### Scenario: A quiet participant is never disturbed

- **WHEN** a participant stays connected far longer than the heartbeat interval without voting,
  revealing, renaming or sending anything at all
- **THEN** their connection is answered by their browser without their involvement, it stays open,
  and neither their seat nor their room is affected

#### Scenario: A client that stopped reading does not pin a goroutine

- **WHEN** the server writes to a connection whose client has stopped reading
- **THEN** the write fails once its deadline expires and the connection is released, rather than
  waiting indefinitely

#### Scenario: A released connection frees everything it held

- **WHEN** a connection is closed for failing to answer its heartbeat
- **THEN** the goroutines serving it finish, its entry in the room is removed, and nothing it held
  is left behind
