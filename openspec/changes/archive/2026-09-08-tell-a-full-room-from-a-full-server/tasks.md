## 1. Naming the refusal

- [x] 1.1 Add `hub.ErrRoomAtCapacity` with the reasoning beside it, and hand it to the refusal path for `AttachAtCapacity` instead of the room-ceiling error; verify the existing hub tests still pass unchanged.
- [x] 1.2 Add the `too_many_connections` code and map the new sentinel to it; verify the test asserting that every sentinel has its own distinct code still passes.

## 2. Saying the right thing

- [x] 2.1 Give `ServerFull.svelte` a scope of `server` or `room` and both wordings, and route the new code to it through `connection.svelte.ts`, keeping the automatic reconnecting stopped in both cases; verify `npm run check` is clean.

## 3. Verification

- [x] 3.1 Add a transport test that a connection refused for a room's connection ceiling carries the new code, and that one refused for the process's room ceiling still carries `at_capacity` — the point being that the two differ.
- [x] 3.2 Run one batch: `go test -race -count=3 ./...` (three passes, not one — a single pass has hidden flaky failures in this project before), `go vet ./...`, `gofmt -l .`, `npm run check`, `npm run build`.
- [x] 3.3 Confirm in a browser against the running container that filling a room's connections shows the room wording and not the server wording.
- [x] 3.4 Run strict OpenSpec validation and `git diff --check`; record the result; then sync and archive.

## Verification results (2026-09-08)

**3.2 — the bundled batch.** `go test -race -count=3 ./...` passed — three full passes,
not one — plus `go vet ./...`, `gofmt -l .` (empty), `npm run check` (175 files, no
errors or warnings) and `npm run build`.

**3.1 — the distinguishing test.** `TestAFullRoomAndAFullServerAreDifferentRefusals`
covers both halves: a process out of rooms still answers `at_capacity`, and a room out
of connections answers `too_many_connections`. It asserts the second is *not* the
first, which is the defect it exists for.

**3.3 — both wordings, seen in a browser** against the real container:

| Situation | Heading | What it says |
| --- | --- | --- |
| Room limited to two connections, both held, a third arrives | **This room is full** | "…the server is not busy — this one room simply has too many connections open. If you have it open in another tab or window, closing that one frees a place immediately." |
| Process limited to one room, one held, a second name opened | **This server is full** | "…running as many games as it can right now. Please try again in a few minutes." |

Worth recording, because it briefly looked like a failure: the first attempt at the
server case *joined* the room instead of being refused. The cause was the test's
timing, not the code — several minutes passed between creating the room over HTTP and
opening the browser, and an empty room is discarded five minutes after it is created.
Repeating it without the gap gave the refusal above, and a `POST /api/games` answering
503 confirmed the ceiling was in force throughout.

**Not checked here.** Nothing new: the container image scan and the load test that S1
wants remain open, unchanged by this change.
