## Why

A connection that dies silently is never noticed. `conn.Read` blocks with no heartbeat behind it, so
a browser whose network vanished without closing the socket — a laptop lid, a lost mobile signal, a
proxy that drops a connection without a close frame — leaves a participant shown as present forever,
holding a seat and keeping the room from ever expiring.

The configured shutdown budget is also not the budget it claims to be. It is passed only to
`server.Shutdown`; `manager.Close()` and `rooms.Wait()` afterwards have no deadline at all, and a
write blocked against a client that stopped reading cannot even observe the stop signal, because the
goroutine is inside a write rather than at its select. The container runtime then kills the process
after ten seconds while `compose.yaml` promises five seconds of orderly cleanup.

## What Changes

- Send a heartbeat on every connection and close one that does not answer, so a silently dead
  connection becomes an away participant within a bounded time instead of never.
- Give every outgoing write its own deadline, so one client that has stopped reading cannot pin a
  goroutine indefinitely.
- Make the shutdown budget a single deadline that covers HTTP and WebSocket together, forcing what
  remains closed when it expires, so the process exits well inside the container's ten seconds.
- Track open sockets so shutdown can reach them, and refuse new ones once stopping has begun.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `app-delivery`: Shutting down gains a real shared deadline and forced closure; a new requirement
  covers the heartbeat and the write deadline that keep a connection honest.
- `room-membership`: A connection that dies without saying so marks its participant away within a
  bounded time, like any other dropped connection.

## Impact

`internal/transport/rooms.go` (heartbeat, write deadlines, socket tracking, refusing sockets while
stopping) and `cmd/planningpoker/main.go` (one deadline covering the whole shutdown). No new
dependency, no protocol change, no change to what a game does.

Answers the review findings about missing write deadlines and liveness detection, and about the shutdown budget not being enforced. **Nothing here is an inactivity timeout, and nothing here may
become one.** A heartbeat is answered by the browser itself with nobody touching the page, so a
meeting where people think for an hour without clicking is untouched; the existing room grace period
and its five-minute default are not altered. Deliberately out of scope: the load test S1 still
wants, and every remaining part of S6.
