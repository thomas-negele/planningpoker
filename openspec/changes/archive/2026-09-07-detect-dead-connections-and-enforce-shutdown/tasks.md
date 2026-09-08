## 1. Keeping a connection honest

- [x] 1.1 Add the heartbeat interval and network deadline as named constants with their reasoning, and make them fields on `RoomHandlers` defaulting to those constants so tests can drive time in milliseconds; verify by construction that the production wiring uses the constants.
- [x] 1.2 Give every outgoing write its own deadline derived from the connection's context; verify with a test that a client which has stopped reading has its connection released instead of pinning a goroutine indefinitely.
- [x] 1.3 Send the heartbeat from the writer goroutine's existing select and end the loop when it goes unanswered, leaving the read loop and the rate limiting from the capacity change untouched; verify a connection that stops answering is closed, its participant marked away with seat and vote intact, and everything it held released.
- [x] 1.4 Verify with its own test that a participant who sends nothing at all across several heartbeat intervals stays present and seated — this is the behaviour the heartbeat must never break, and the reason it is not an inactivity timeout.

## 2. A shutdown budget that is actually the budget

- [x] 2.1 Add a mutex-guarded registry of live sockets and a stopping flag to `RoomHandlers`, entering a socket when accepted and removing it when its handler returns; verify a connection attempted after stopping has begun is refused and creates no room.
- [x] 2.2 Derive one deadline from the configured budget when the signal arrives and use it for the HTTP shutdown, the room closure and the socket wait, forcing what remains closed when it expires; verify with a test using an ordinary client and one with a client that never answers that both finish inside the budget and the process exits with status 0.
- [x] 2.3 Verify separately that a long connected session with no game actions does not expire, using driven rather than real time, so that nothing in this package can be mistaken for an inactivity rule.

## 3. One bundled verification

- [x] 3.1 Once every edit is in place, run one batch: `go test -race ./...` in both the development and `embedassets` modes, `go vet ./...`, `gofmt -l .`, `npm run check` and `npm run build`. Do not rerun suites between individual edits, and do not modify the existing rate-limit tests to accommodate this change — if they need changing, that is a finding to report.
- [x] 3.2 Build the image once and confirm in the real container that a normal game still works end to end and that `docker compose down` returns well inside the container's ten-second grace period, with the process exiting cleanly rather than being killed. Record the measured stop time.
- [x] 3.3 Run strict OpenSpec validation and `git diff --check`; record the exact results and any check left undone, ticking only what was verified. After review, sync and archive this change; only then mark the package complete.

## Verification results (2026-09-07)

**3.1 — the bundled batch.** `go test -race ./...` passed in both the development and
the `embedassets` production asset modes; `go vet ./...` and `gofmt -l .` are clean;
`npm run check` reports 174 files with no errors or warnings; `npm run build`
succeeds. The existing rate-limit and capacity tests were **not** modified and pass
unchanged, which was the point of leaving the read loop alone.

**3.2 — the real container.** One image built, then a full game played through it:
two participants seated, both voted, the round stayed hidden with no results object,
the reveal showed both cards and a new round hid them again. The container was then
stopped while two connections were still open, **one of them a client that had
stopped reading** — the case that used to cost five seconds per connection.

- `docker compose stop` returned in **0.17 s**, against a container grace period of
  ten seconds.
- `docker inspect` reports `ExitCode=0`, `OOMKilled=false`, no error: the process
  exited by itself rather than being killed.
- The logs end with `shutdown requested, closing connections` and `stopped cleanly`.
- A separate run measured `docker compose down` at 0.27 s under the same conditions.

**Tests added.** A connection that stops answering is closed and its participant
marked away with the seat kept; five such connections are released rather than
pinning goroutines; a shutdown against an ordinary client and a silent one finishes
inside its budget; a connection arriving after shutdown has begun is refused and
creates no room; and — the one that guards the owner's actual concern — a participant
who sends nothing at all across ten heartbeat intervals is still present and not
away, judged while their connection is still open. At the rules level, a room whose
participant says nothing survives twelve grace periods of driven time with the seat
still occupied and present.

**A design decision was corrected by measurement.** The original design said the
budget would end by forcing sockets closed and waiting again. `CloseNow` turns out
not to abort a polite `Close` that is already in flight — it queued behind it for 4.9
seconds against a client that never answered — so that second wait would have been
the unbounded wait this package removes. The code bounds its own waiting instead and
exits without a second wait; design.md decision 5 records what was believed, what was
measured, and what the code does.

**Not checked here.** Public DNS and a publicly trusted certificate, as in every
package so far. The load test S1 still wants was not done and S1 stays open. Whether
a reverse proxy's own idle timeout interacts badly with the thirty-second heartbeat
was reasoned about, not measured against a real proxy under load.
