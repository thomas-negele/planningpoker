## 1. The room goroutine

- [x] 1.1 Create `internal/hub` with a room that owns one `game.Room` and runs a single goroutine processing commands from a channel; verify with a test that a command sent to it changes the room and that nothing outside the goroutine holds a pointer into the room's state
- [x] 1.2 Give each connection a buffered outbound channel owned by that connection, and have the room send snapshots without ever waiting on a socket; verify with a test that the room continues serving other connections while one is not being read
- [x] 1.3 Drop a connection whose outbound buffer overflows rather than blocking the room; verify with a test that fills one connection's buffer, asserts the room still answers other connections promptly, and asserts the stuck connection is closed
- [x] 1.4 Have the room report whether it currently has open connections and when it last had one; verify with a test covering a room that never had one, one that has one, and one whose last connection has closed

## 2. The manager

- [x] 2.1 Implement the manager holding a map from room identifier to room behind a lock that protects the map and nothing else; verify by review that the lock is never held across a room operation, and with a test that operations on two different rooms proceed concurrently
- [x] 2.2 Implement creating a room, looking one up, and reporting that an identifier is unknown; verify with tests including a lookup of an identifier that never existed
- [x] 2.3 Make a room that has been removed refuse further commands rather than accepting them into a channel nobody reads; verify with a test that commands a room after its removal and asserts a clean refusal instead of a hang
- [x] 2.4 Take the clock as a dependency rather than calling the system clock, so expiry can be tested without waiting; verify with a test using a controlled clock

## 3. Room lifetime

- [x] 3.1 Implement the periodic sweep that discards rooms with no connection for longer than the grace period, covering both a room everyone left and a room nobody ever joined; record the sweep interval in the code with its reasoning
- [x] 3.2 Verify both directions of the rule with a controlled clock: a room just under the grace period still exists, and just over it is gone
- [x] 3.3 Verify a reload does not destroy a room — closing the only connection and opening a new one within the grace period leaves participants and votes intact
- [x] 3.4 Verify an occupied room is never discarded however long it runs
- [x] 3.5 Verify a discarded identifier reaches nothing afterwards, and that nothing recreates a room at a previously used identifier

## 4. Configuration

- [x] 4.1 Add `PLANNINGPOKER_ROOM_GRACE_PERIOD` to the configuration struct with a documented default of five minutes, rejecting zero, negative and unparseable values the way the existing settings do; verify with tests covering the default, valid overrides and every rejected form
- [x] 4.2 Document the value in `compose.yaml` in full sentences, saying what it does, what it does not do, and what happens at its boundary; verify the file still parses with `docker compose config`

## 5. Protocol encoding

- [x] 5.1 Define the message types the client may send — take a seat, play a card, reveal, start a new round, change name — and decode them, refusing anything unrecognised or malformed without disturbing the room; verify with tests including invalid JSON and an unknown intent name
- [x] 5.2 Define the snapshot message, built only from `game.Room.View()` plus whatever is genuinely about the connection; verify by review that the transport never reaches into the room for state and never assembles a message from parts
- [x] 5.3 Map every domain sentinel error to its own stable refusal code, exhaustively, so that an unmapped error is a test failure rather than a silent downgrade to a generic code; verify with a test that walks the package's sentinel errors and asserts each has a distinct code
- [x] 5.4 Send a refusal only to the connection that caused it, leaving the room unchanged and sending no snapshot on its account; verify with a test using two connections that asserts the second receives nothing

## 6. The WebSocket endpoint

- [x] 6.1 Replace the echo endpoint with a handler at a path naming a room, rejecting a connection to an unknown room in a way the client can distinguish from any other failure; verify with tests for a known and an unknown room
- [x] 6.2 Send a snapshot immediately on connecting, before anything else; verify with a test that connects to a room where people are already seated and voting and asserts the first message received is a complete snapshot
- [x] 6.3 Broadcast a fresh snapshot to every connection of a room after any change; verify with a test using three connections asserting all three receive it
- [x] 6.4 Delete `echo_probe.go` and its tests entirely; verify the package builds and no reference to the echo endpoint remains anywhere in the repository
- [x] 6.5 Keep the library's default `Origin` check and confirm it still refuses a foreign origin on the new path; verify with a test

## 7. Participant identity and connections

- [x] 7.1 Issue an opaque participant identifier in a cookie scoped to the room's path, not readable by scripts in the page; verify with a test asserting the cookie's path and that it is marked against script access
- [x] 7.2 Reseat a browser presenting an identifier already seated in that room, keeping seat, name and vote and adding no second participant; verify with a test
- [x] 7.3 Treat an identifier that names nobody in this room as a new arrival rather than an error; verify with a test presenting an identifier from a different room
- [x] 7.4 Verify the same browser taking a seat in two rooms receives two different identifiers, so neither room can tell they are the same browser
- [x] 7.5 Support several connections per participant: one seat, all connections receiving snapshots, an action on one appearing on the others; verify with a test using two connections for one identifier
- [x] 7.6 Mark a participant away only when their last connection closes, and clear the mark when a new one opens; verify with a test that closes one of two connections and asserts they are not away, then closes the second and asserts they are

## 8. Creating a game over HTTP

- [x] 8.1 Add the route that creates a game and returns the new room's identifier, seating nobody; verify with a test asserting the room exists, has no participants, and that the creator holds no privilege
- [x] 8.2 Verify two creations produce two distinct rooms, and that a vote in one is not visible in the other
- [x] 8.3 Verify the route rejects methods other than the one it defines, so that a stray link cannot create rooms

## 9. Shutdown

- [x] 9.1 Close every room goroutine and every connection on shutdown, within the existing shutdown budget, replacing the connection tracking that belonged to the echo probe; verify by sending `SIGTERM` to a running process with several rooms and connections open and confirming exit status 0 within the budget
- [x] 9.2 Verify no goroutine is leaked by shutdown — assert the room and connection goroutines have ended rather than assuming they have

## 10. The guarantee this change exists to protect

- [x] 10.1 Write the test that reads the actual bytes sent over a real WebSocket during a hidden round in which every participant has voted a different card, and asserts no card value appears in any message — this must inspect what crossed the socket, not a Go value, because the domain's own test already proves the struct is clean and this layer is where a leak would now occur
- [x] 10.2 Verify that test fails if the transport is deliberately made to include a vote value, so the guard is known to work rather than assumed to
- [x] 10.3 Verify the snapshot after a reveal does contain every card and the count per card, so the hiding is not simply breaking the reveal

## 11. Concurrency

- [x] 11.1 Write a concurrency test in which many connections take seats, vote, reveal and start new rounds against the same room simultaneously, asserting the room stays consistent and every snapshot is well-formed; verify it passes repeatedly under `go test -race`
- [x] 11.2 Verify the whole suite runs clean under `go test -race ./...` in both build variants, since the entire design rests on room state having a single owner

## 12. Keeping the deployed application coherent

- [x] 12.1 Remove the connection probe from the placeholder page, since the endpoint it probed no longer exists, and leave the page stating plainly that the table arrives in the next change; verify `npm run check` is clean and the page shows no failed connection
- [x] 12.2 Exercise the protocol end to end against the running container with a script: create a game, connect two clients, seat them, vote, reveal, revote — and record the transcript as evidence the product works without a user interface

## 13. Acceptance

- [x] 13.1 Run the full command set from `CLAUDE.md` and confirm each is clean: `go build ./...`, `go test -race ./...`, `go vet ./...`, `gofmt -l .` with empty output, and `npm run check` in `web/`
- [x] 13.2 Verify the container still builds and serves, the Content-Security-Policy is unchanged, and the WebSocket upgrade still succeeds through a TLS-terminating reverse proxy on the new path
- [x] 13.3 Confirm `internal/game` was not modified by this change; if it was, stop and say why rather than absorbing a rule change silently
