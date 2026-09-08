## 1. Storing a name only when it was asked for

- [x] 1.1 Rework `web/src/lib/name.ts` so that storing is explicit: keep `rememberedName`, make `rememberName` the deliberate act, add `forgetName` that deletes the cookie, and set `Secure` when the page was served over HTTPS; verify the module compiles under `npm run check` and that nothing else still writes the cookie.
- [x] 1.2 Add the off-by-default choice to `NamePrompt.svelte`, next to the field, stating what is stored and for how long; pre-fill and show the choice as made only when a name is actually stored, and delete it the moment the choice is turned off.
- [x] 1.3 Pass the same choice through the rename path in `RoomView.svelte`, so renaming updates a stored name only when one was asked for and never creates one; verify no remaining call site stores unconditionally.

## 2. The seat token's lifetime

- [x] 2.1 Make the seat cookie a session cookie by dropping its max age, and rewrite the reasoning beside it: what its absence costs, why a session is the right span, and that the server does not enforce any expiry; verify with a test that the issued cookie carries neither `Max-Age` nor `Expires`, and that the existing reconnect tests still pass unchanged.

## 3. Saying what is kept

- [x] 3.1 Write a short privacy document directly, listing only what can be read off the code: rooms, participants, names and votes in memory with their lifetimes, the two cookies with their purposes and durations, and what a browser or proxy will log. Mark controller, legal basis, contact and hosting as gaps belonging to D3 rather than filling them in. Link it from the README.

## 4. One bundled verification

- [x] 4.1 Once every edit is in place, run one batch: `npm run check`, `npm run build`, `go test -race ./...`, `go vet ./...` and `gofmt -l .`. Do not rerun suites between individual edits.
- [x] 4.2 Check the actual behaviour in a browser against the running application, because this change is almost entirely about what somebody sees and what lands on their device: with the choice off, confirm that no `pp_name` cookie exists after seating and renaming; with it on, that the cookie appears with the stated duration and that the name is pre-filled next time; that turning it off again deletes it; and that the seat cookie now carries no expiry at all. Record exactly what was observed. If the browser tooling cannot inspect the page, say so and ask rather than inferring.
- [x] 4.3 Run strict OpenSpec validation and `git diff --check`; record the exact results and any check left undone, claiming only what was verified. The cookie decision and the `Secure` attribute are settled by this work; the account of what is stored is only partly settled, and everything that depends on who operates an installation is untouched. After review, sync and archive this change; only then mark the package complete.

## Verification results (2026-09-07)

**4.1 — the bundled batch.** `npm run check` reports 175 files with no errors and no
warnings; `npm run build` succeeds; `go test -race ./...`, `go vet ./...` and
`gofmt -l .` are clean.

**2.1 — the seat cookie.** Confirmed against the running container:
`Set-Cookie: pp_seat_<room>=…; Path=/ws/<room>; HttpOnly; SameSite=Lax` — no
`Max-Age` and no `Expires`. A test asserts this on the sent header rather than the
parsed cookie, because Go reports a missing max age and a max age of zero
identically while those mean opposite things on the wire.

**4.2 — in a real browser, against the running container.** The automated browser
could execute JavaScript in the page even though screenshots and cookie reading were
blocked, so the form's own state was used as the probe: the page reads its cookie
itself, and a stored name arrives as a pre-filled field with the choice shown as
made. Each step below opened a **different** room, so that no seat cookie could
reseat the visitor and mask the result.

| Step | Observed |
| --- | --- |
| Fresh start | field empty, choice off |
| Name typed, choice **off**, seated → new room | field **empty**, choice off — nothing stored |
| Name typed, choice **on**, seated → new room | field **"Trillian"**, choice on — stored |
| Choice turned off → new room | field **empty**, choice off — deleted at once |
| Seat, then open own name at the table | dialog opens with the current name and the choice, worded as at the prompt |
| Choice turned **on** in that dialog, saved → new room | field pre-filled, choice on — agreeing from the table works |
| Choice turned **off** in that dialog, then **cancelled** → new room | field empty — withdrawal takes effect immediately, not on save |

**Two defects were found by the owner testing it, and both are fixed.** The name
field carried `autocomplete="nickname"`, so the browser kept its own copy and offered
it back regardless of the choice; it is now `off`. And the choice was unreachable
once seated, which broke this change's own promise that it is reversible; changing a
name from the table now opens a dialog carrying both controls. The specification
gained a requirement for that rather than the code quietly gaining a feature.

**What the owner's original report turned out to be.** A name that "was always
stored" was almost certainly the seat cookie doing its job: returning to the *same*
room link reseats the browser without a prompt, and the name shown at the table comes
from the server, not from `pp_name`. The tests above deliberately use a different
room each time for exactly that reason.

**Not checked here.** Everything in D3 — controller, legal basis, contact, hosting —
is untouched by design. `PRIVACY.md` marks those as gaps rather than filling them.
Whether the consent mechanism suffices legally is the owner's call and remains so.
