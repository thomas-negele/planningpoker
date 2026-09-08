## Context

See proposal.md for motivation. The constraints that shape the approach:

- Each connection is served by two goroutines: `readIntents` on the caller's goroutine and
  `writeUpdates` on its own. Both take the request's context, which for a hijacked WebSocket is
  never cancelled by `http.Server.Shutdown` — `Shutdown` ignores hijacked connections entirely.
- `RoomHandlers.sockets` is a `sync.WaitGroup` that shutdown already waits on. It can say how many
  connections exist but cannot reach them, which is exactly what forced closure needs.
- `websocket.Conn.Ping` sends a ping and waits for the matching pong, returning an error if its
  context expires. It serves as heartbeat and liveness check in one call, so no pong handler or
  bookkeeping is needed.
- Concurrent writes are already serialised by the library, so a heartbeat sent from a third
  goroutine does not race the writer.
- The capacity change, `bound-capacity-and-reject-before-creating`, left the read and write loops deliberately untouched so this one could work on
  them. Its rate limiting lives at the top of the read loop and must keep working unchanged.

## Goals / Non-Goals

**Goals:** Notice a connection that has died silently, stop one slow client from pinning a
goroutine, and make the configured shutdown budget the budget that is actually enforced.

**Non-Goals:** Any inactivity rule measured in game actions; changes to the room grace period or to
capacity limits; a shutdown framework; per-room timers. The load test S1 still wants and the
remaining S6 input hardening are separate.

## Decisions

1. **Two constants, not settings**, decided with the owner: a heartbeat every **30 seconds** and a
   **10-second** deadline used both for waiting on the pong and for completing one write. Unlike the
   capacity ceilings, these are not values an operator sensibly tunes to their machine — they follow
   from how browsers and proxies behave, and setting them wrongly breaks connections rather than
   saving resources. Thirty seconds also sits comfortably under the 60-second idle timeout common in
   reverse proxies, so the heartbeat keeps the connection alive as a side effect. Worst-case
   detection is therefore 40 seconds. They are written as named constants with the reasoning beside
   them, and the standing rule from CLAUDE.md is satisfied by having asked rather than by adding a
   knob nobody wants.

2. **The heartbeat runs in the writer goroutine, not a third one.** `writeUpdates` already selects
   over the update channel and the context; adding a `time.Ticker` case makes it the one place that
   writes to the socket. A separate pinger would work — the library serialises writes — but it would
   be a third goroutine per connection whose lifetime has to be tied to the other two, and this
   avoids that entirely. A failing ping ends the loop exactly as a failing write does, and the
   existing cleanup then detaches the connection and marks the participant away.

3. **Every write gets its own deadline** from a context derived from the connection's, rather than
   the connection context itself. That is the whole of the fix for a client that has stopped
   reading: today such a write blocks forever and the goroutine cannot even reach its select to see
   the stop signal.

4. **Shutdown gets a registry of live sockets** on `RoomHandlers`: a mutex-guarded set of
   `*websocket.Conn`, entered when a socket is accepted and removed when its handler returns, plus a
   `stopping` flag. The flag is checked before accepting, so cleanup does not race arrivals. This is
   the smallest thing that can turn "how many are open" into "close them", and the checklist's own
   suggestion.

5. **One deadline for the whole shutdown.** `main.go` derives a single context from the configured
   budget at the moment the signal arrives and uses it for `server.Shutdown`, for closing rooms, and
   for waiting on sockets. Ordering matters and is worth stating — refuse new sockets first, then ask
   rooms to close so pages learn why, then wait. Forcing first would deny the clean goodbye the
   budget exists to allow.

   **Corrected during implementation, because the first version of this decision was wrong.** It
   said that when the budget expired we would `CloseNow()` every remaining socket and wait again.
   Measured, that does not work: `CloseNow` on a connection whose polite `Close` is already in
   flight does not abort it but queues behind it, returning after 4.9 seconds against a client that
   never answers — the same five seconds the `Close` itself costs. A second wait would therefore
   have been precisely the unbounded wait this package exists to remove.

   What the code does instead is bound its own waiting rather than pretend it can interrupt the
   library. The polite close in the writer goroutine is wrapped so it cannot block for longer than
   the connection deadline; `CloseNow` nudges every socket without waiting for any of them; and when
   the budget expires `main.go` logs, nudges, closes the HTTP side and **exits without waiting
   again**. Sockets still open are closed by the operating system as the process dies, which is what
   the specification's "exit anyway rather than hang" means. The goroutine left inside the library's
   closing handshake is accepted deliberately: it cannot be interrupted, it costs a stack, and the
   process is on its way out.

6. **Tests drive time rather than sleeping through it.** The heartbeat interval and deadline become
   fields on `RoomHandlers` defaulting to the constants, so a test can set them to milliseconds. A
   test that proves a 30-second heartbeat by waiting 30 seconds is a test nobody runs. The
   deliberately quiet participant is checked the same way: with a short interval, several heartbeat
   rounds pass while a connection sends nothing, and it must still be present.

## Risks / Trade-offs

- **A heartbeat mistaken for an activity timeout by a later change** → the spec says twice that it is
  not one, there is a test that a silent participant survives many intervals, and the constants
  carry the warning where somebody editing them will read it.
- **10 seconds is tight for a very slow mobile connection** → it bounds one write and one pong, not a
  whole session, and a dropped connection reconnects on its own. The alternative, an unbounded write,
  is what this package exists to remove.
- **Forced closure loses a clean goodbye for whoever is still open at the deadline** → that is the
  intended trade and the reason the budget is refused at zero; the orderly path is tried first and
  the budget only ends the waiting.
- **Touching both loops right after the capacity change did** → `bound-capacity-and-reject-before-creating` stayed out of them on purpose. Keep
  the rate limiting exactly as it is and re-run its tests unchanged rather than adjusting them.

## Migration Plan

No configuration changes: the two new values are constants and the existing shutdown budget keeps
its meaning, now enforced. Deploy outside a meeting as always, since rooms live in memory. Rolling
back is the ordinary rebuild from the previous commit. Clients need no change — the browser answers
pings itself, and a connection closed for liveness is reconnected by the page's existing logic.
