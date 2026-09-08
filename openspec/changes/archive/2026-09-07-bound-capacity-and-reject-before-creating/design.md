## Context

See proposal.md for motivation. The constraints that shape the approach:

- `Manager` guards only its room map with a mutex; each room owns its own state in one goroutine and
  is spoken to over a command channel buffered at 32. A room's outbound channel per connection is
  buffered at 16. Nothing may reach into a room's state from outside that goroutine.
- `internal/game` is pure rules with no I/O and must stay that way, so anything counted in the rules
  (participants) is enforced there, while anything counted across rooms (rooms, connections) is
  enforced in `internal/hub` or the transport.
- A room's five-minute grace period only starts when its **last** connection closes. One held socket
  keeps a room alive indefinitely, which is why cleanup is not a defence against accumulation.
- Three settings already exist as environment variables with full-sentence explanations in
  `compose.yaml`. The owner has decided the four new limits follow that same pattern rather than
  becoming constants or an administration area.

## Goals / Non-Goals

**Goals:** Put a ceiling on everything that grows, refuse the excess in a way the page can explain,
and stop room creation from happening before a connection exists — with regression tests for the
behaviours that are easy to reintroduce.

**Non-Goals:** Write deadlines, heartbeats and liveness detection (S3) and enforcing the shutdown
budget (B1) — those are the next package and touch the same file, so this one must not pre-empt
them. Also excluded: per-IP accounting, any external rate-limit service, load-testing infrastructure,
and any limit that ends a meeting because nobody clicked.

## Decisions

1. **The four values and their defaults**, decided with the owner: 50 rooms in the process, 40
   connections per room, 20 participants per room, and 10 messages per second per connection with a
   burst allowance of 20. Environment variables named in the existing style
   (`PLANNINGPOKER_MAX_ROOMS`, `PLANNINGPOKER_MAX_CONNECTIONS_PER_ROOM`,
   `PLANNINGPOKER_MAX_PARTICIPANTS_PER_ROOM`, `PLANNINGPOKER_MESSAGE_RATE`), each with a
   full-sentence explanation in `compose.yaml` saying what it bounds, what a person hits first, and
   why there is no value meaning "unlimited". Connections are set to twice the participant limit on
   purpose: a second tab, a not-yet-seated visitor and the overlap during a reconnect must not
   consume the seat budget.

2. **The rate limit is per connection, enforced on the read side, as a token bucket.** A token bucket
   holds a number of tokens up to the burst size, refilled at the sustained rate; each message spends
   one. This is what lets a short cluster of intents through while still bounding the long-run
   average. The alternative, a fixed window counter, either refuses legitimate bursts or permits
   twice the rate across a window boundary. It lives in the transport's read loop, where the message
   already is, so a refused message never reaches a room's command channel.

3. **Order of operations at the socket endpoint is the whole of the S2 fix.** Today `EnsureRoom` and
   the seat cookie both happen before `websocket.Accept`. The new order is: validate the identifier,
   accept the upgrade (which is where the library's Origin check runs), and only then ensure the
   room and issue the cookie. One consequence must be handled rather than ignored: a cookie can only
   be set on an HTTP response, and after the upgrade there is no response left to set it on. The seat
   token therefore has to be **read** from the request before the upgrade, and a **newly minted** one
   delivered to the browser another way. The intended solution is to keep issuing the cookie on the
   handshake response but to mint it only after the identifier has been validated *and* the request
   has been confirmed to be a genuine same-origin upgrade — checking those two things before
   `Accept` costs nothing and does not require creating a room. The room itself is created after
   `Accept` succeeds. If that proves not to hold together in code, the fallback is a small
   `POST /api/rooms/{id}/seat` step, and that would be a decision to bring back to the owner rather
   than to take silently.

4. **The global room limit is checked inside the manager's write lock**, in the same critical section
   that inserts, because checking the count and then inserting under a second lock is exactly the
   race the existing code took the write lock for a read to avoid. `Create` and `EnsureRoom` both go
   through it.

5. **The per-room connection and participant limits are enforced inside the room's own goroutine**,
   in `attachCommand` and `seatCommand`, since those are the only places that may read that room's
   state. `Attach` already returns a boolean; it gains a distinguishable refusal instead. The
   participant ceiling is additionally enforced in `internal/game`, because "how many people may sit
   at a table" is a rule, not a transport concern, and it must hold for anything calling the rules
   directly in a test.

6. **Reseating never counts against the participant limit.** `Rejoin` finds an existing seat and adds
   nobody, so the check belongs on the path that creates a new participant, not on the seat intent as
   a whole. Getting this wrong would make reconnection fail exactly when a room is busy, which is the
   worst possible time.

7. **Message size** is set explicitly with the library's read limit to a value derived from the
   protocol's largest legitimate message — a rename or seat intent carrying a 15-character name —
   with generous headroom rather than a value tuned to the byte. The library's 32 KiB default is not
   wrong, but relying on a default for a security-relevant bound means a library upgrade can change
   it silently.

8. **The frontend decoding fix is a `try`/`catch` around `decodeURIComponent`** in
   `web/src/lib/router.svelte.ts`, returning a route that the existing "not a game" view already
   handles. Worth knowing: `parse()` runs at module initialisation, so today an invalid escape throws
   before anything renders and the page stays blank — this is not merely a wrong screen.

## Risks / Trade-offs

- **The cookie ordering in decision 3 may not survive contact with the code** → the fallback is named
  above and is an owner decision, not a silent redesign. Nothing else in the change depends on which
  way it goes.
- **A limit that is too low silently spoils a real meeting** → the defaults are far above any real
  team, refusals are explicit and distinguishable rather than silent, and all four are adjustable
  without a rebuild.
- **Refusing at capacity is itself a way to deny service**: whoever fills 50 rooms first keeps others
  out → accepted for a hobby deployment, and stated plainly rather than papered over. Per-IP
  accounting and proxy-level limits are deliberately out of scope; the ceiling protects the host,
  which is what this package is for.
- **The rate limiter could refuse a legitimate reconnect burst** → hence the burst allowance, and a
  test that a full room voting simultaneously passes untouched.
- **Touching the socket handler now and again for S3 next** → keep this change's edits to ordering
  and admission, and leave the read/write loops alone, so the next package does not have to unpick
  this one.

## Migration Plan

Deploy outside a meeting like any other update, since rooms live in memory. The defaults apply with
no configuration change, so an existing deployment needs no action. Rolling back is the ordinary
rebuild from the previous commit described in the README. No data, protocol or client compatibility
is affected: an older page talking to a newer server sees only new refusal reasons, which it displays
through the generic path it already has.
