## 1. Configuration for the four limits

- [x] 1.1 Read the four limits from the environment with the defaults from design.md decision 1, refusing zero, negative and unparseable values at startup with an error naming the variable and the value; verify with unit tests over the parsing alone, covering each boundary once rather than once per variable.
- [x] 1.2 Add the four settings to `compose.yaml` with full-sentence explanations saying what each bounds, what a person hits first, and why no value means "unlimited"; verify by resolving the configuration and by overriding each once in the batch in 5.1.

## 2. Ceilings on rooms, connections and seats

- [x] 2.1 Enforce the room ceiling inside the manager's existing write lock, on both creation paths, so counting and inserting cannot race; verify with a test that fills the ceiling, is refused, and finds the room count unchanged.
- [x] 2.2 Enforce the connection ceiling in the room goroutine's attach command and give `Attach` a distinguishable refusal instead of a bare false; verify existing attach/detach tests still pass and a new test shows attached connections are undisturbed by a refusal.
- [x] 2.3 Enforce the participant ceiling in `internal/game` on the path that creates a new participant, leaving reseating unaffected; verify with rules-level tests that a full room refuses a new seat, counts away participants as occupying theirs, and still reseats a returning token.
- [x] 2.4 Carry both refusals through the transport as their own reasons; verify a refused connection or seat leaves the room's participants, votes and round state untouched and sends nothing to anyone else.

## 3. Nothing is created before a connection exists

- [x] 3.1 Reorder the socket handler so the identifier is validated and the request confirmed as a genuine same-origin upgrade before anything is created, the room is ensured only after the upgrade is accepted, and a seat token is minted only for a request that passes both; verify with regression tests that a plain HTTP request and a foreign-origin request each create no room and return no seat cookie. If the cookie cannot be issued this way, stop and bring design.md decision 3's fallback to the owner rather than choosing silently.
- [x] 3.2 Refuse `POST /api/games` when the request comes from another website, keeping ordinary same-origin use working; verify with tests for an absent, a matching and a foreign origin.
- [x] 3.3 Set the WebSocket read limit explicitly from the protocol's largest legitimate message with headroom; verify an oversized message is refused without the room changing.

## 4. Per-connection message rate

- [x] 4.1 Add a token-bucket rate limit per connection in the read loop, with the sustained rate and burst from design.md decision 1, refusing an over-rate message before it reaches the room and closing a connection that persists; verify a refused flood leaves the room unchanged, the participant keeps their seat and is marked away only on close, and the read and write loops are otherwise untouched so the next package can work on them.
- [x] 4.2 Verify with a test that a full room voting within the same second passes entirely untouched — this is the behaviour the limit must never break.

## 5. Frontend and one bundled verification

- [x] 5.1 Handle invalid percent-encoding in `web/src/lib/router.svelte.ts` so a malformed link reaches the existing "not a game" view, remembering that `parse()` runs at module initialisation and today leaves the page blank; verify `/g/%FF` shows that view and ordinary encoded paths still resolve.
- [x] 5.2 Once every edit above is in place, run one batch: `go test -race ./...`, `go vet ./...`, `gofmt -l .`, `npm run check` and `npm run build`; resolve the base and one overridden Compose configuration. Do not rerun suites between individual edits.
- [x] 5.3 Build the image once and smoke the real container: seat, vote, reveal and new round still work; a full table votes simultaneously without refusal; a refused capacity attempt shows its own message rather than an empty table. Record exactly what was checked and what was not.
- [x] 5.4 Run strict OpenSpec validation and `git diff --check`; record the exact results and any check left undone, claiming only what was verified. After review, sync and archive this change; only then mark the package complete.

## Verification results (2026-09-07)

**5.2 — the bundled batch.** `go test -race ./...` passed in both the development
and the `embedassets` production asset modes; `go vet ./...` and `gofmt -l .` are
clean; `npm run check` reports 174 files, 0 errors, 0 warnings; `npm run build`
succeeds (58.2 kB of JavaScript, 21.8 kB gzipped). The base, local and server
Compose configurations all resolve, and one overridden set
(`MAX_ROOMS=3`, `MAX_CONNECTIONS_PER_ROOM=6`, `MAX_PARTICIPANTS_PER_ROOM=2`,
`MESSAGE_RATE=4`) reaches the service with every container restriction intact and
the local port still bound to `127.0.0.1` only.

**5.3 — the real container**, built once and run with deliberately small ceilings
(2 rooms, 4 seats, 8 connections, 10 messages per second) so they could be reached:

- The game is unaffected: four participants seated, all four voted **within the same
  moment from their own connections and not one was refused**, the round stayed
  hidden with no results object, the reveal showed four cards, and a new round hid
  them again.
- A fifth person at the full table was refused with `room_full`, and that refused
  connection stayed open and usable — the next intent came back `not_seated`, not a
  closed socket.
- A third game beyond the ceiling of two was refused with HTTP 503, and a socket for
  a new room name was refused with `at_capacity` — its own code, distinct from a
  room that does not exist and from a temporary fault.
- An ordinary HTTP GET to a socket path returned 426 and **no** cookie; a POST to
  `/api/games` carrying a foreign `Origin` returned 403 and no cookie. Neither
  created a room.
- The frontend is still served whole with its Content-Security-Policy header.

**Unit and regression tests added:** the four settings and their boundaries; the
token bucket against a stepped clock (starts full, refills continuously, the burst is
spent once rather than earned every interval); the room ceiling under 50 simultaneous
creations, which is what proves counting and inserting share one critical section; a
full table refusing a seat while an away participant still occupies theirs; a
reconnecting browser never refused for capacity; refused work leaving no room,
goroutine or seat token behind; a message above the read limit ending only its own
connection.

**5.1 — confirmed, including the render.** The owner opened
`http://127.0.0.1:8080/g/%FF` in a browser and saw the "not a game" screen with its
explanation of what a room name must look like and the offer to start one — not the
blank page the unguarded call used to produce. Everything leading to it was verified
separately: `decodeURIComponent("%FF")` does throw a
`URIError` in Node, confirming the original fault; the guarded form returns the raw
segment while `"team%2Dalpha"` still decodes normally; the built bundle contains
`try{return decodeURIComponent(e)}catch{return e}`; `GET /g/%FF` is served as HTML
with status 200; and both `/ws/%FF` and `/ws/%25FF` are refused by the server with
`invalid_room_id`, which is precisely the code that puts the page into its `refused`
state and renders that screen.

Worth recording for whoever reads this later: the automated browser tooling could not
inject a script into the page at all — screenshots and text extraction both timed out,
on the plain entry page as well — so this one step was confirmed by hand instead. That
was a limitation of the tooling, not of the application.

**5.4 — validation, documentation and synchronization.**
`openspec validate bound-capacity-and-reject-before-creating --strict` passes,
`git diff --check` is clean, and all four delta specs were merged into their main
specs and checked back requirement by requirement and scenario by scenario.
The review findings this change answers — nothing created before a connection exists,
and the malformed-link handling — are settled; the capacity load test and the remaining
input hardening are recorded as still open, with precise notes on what is done and
what is not.

**Not checked here.** Public DNS and a publicly trusted certificate, as before. S1's
own completion criteria additionally name a controlled local load test showing that
an overloaded session does not drag down other rooms or the host; the ceilings are
verified but behaviour under real load is not, so S1 stays open. Write deadlines,
heartbeats and the shutdown budget are untouched and remain the next package.
