# Verification — redesign-interface

What was run, what it said, and what was not run. Recorded at the end of
implementation, against the code as committed.

## The project's verification order

From CONTRIBUTING.md, in its order, all from the repository root.

| Check | Result |
| --- | --- |
| `cd web && npm ci && npm test` | 12 pass, 0 fail |
| `cd web && npm run check` | 184 files, 0 errors, 0 warnings |
| `cd web && npm run build` | built |
| `go vet ./...` | clean |
| `gofmt -l .` | no output |
| `go test -race -count=3 ./...` | ok in every package |
| `go test -race -count=3 -tags embedassets ./...` | ok in every package |
| `cd e2e && npm run check` | clean |
| `cd e2e && npm test` | 25 pass, 0 fail |

The browser suite grew from 3 tests to 25. This change contributes the entry
screen (7), the information controls (7) and reduced motion (3); the rest came
with main and still pass.

Re-run in full after merging main, not only before.

## The endpoint, the font and its licence, in the embedded build

Checked against a binary built with `-tags embedassets`, which is what the
container runs, rather than against the development server.

- `GET /api/decks` answers `200 application/json` with both decks and their
  cards in deck order.
- `/fonts/Geist-Variable.woff2` answers `200 font/woff2`, 69,760 bytes.
- `/fonts/Geist-OFL.txt` answers `200 text/plain`, 4,383 bytes, opening with the
  copyright line of the Geist authors. The complete licence: preamble,
  definitions, permission and conditions, termination, disclaimer.
- The browser contacts exactly one origin. The font request goes to that origin.
- No Content-Security-Policy violation is reported. `font-src 'self'` is
  unchanged and was not widened.
- The built output contains no reference to a foreign host other than strings
  inside Svelte's runtime — documentation links printed on a framework error and
  the XHTML namespace constant. Neither causes a request, and both predate this
  change.

## Glyph coverage

The bundled file carries 728 glyphs. Checked present: Polish `ł ą ż ś ć`, Czech
`č š ž`, Hungarian `ő ű`, Romanian `ș ț ă`, Turkish `ı İ ğ ş`, Baltic, Icelandic,
Maltese, Vietnamese, Cyrillic, and the `½` the Fibonacci deck needs. Participants
named Michał, Škoda and Gülşen were seated in the running application and every
letter rendered in Geist.

Not covered: Greek, Hebrew, Arabic, CJK. A name in one of those falls back to the
system font behind Geist — legible, in another typeface. Accepted; see design.md.

## Contrast

Measured with `docs/design/contrast.py`, which reads the values out of
`web/src/app.css` and composites translucent colours over the backdrop they
actually sit on.

```
TEXT — needs 4.5:1
--------------------------------------------------------------------------
  ok   18.08:1   --text on the page
  ok   17.01:1   --text on a panel
  ok   15.82:1   --text on a raised surface
  ok    6.28:1   --text-dim on the page
  ok    5.91:1   --text-dim on a panel
  ok    5.26:1   --text-dim on the table
  ok    6.05:1   --accent as the name of you
  ok    6.03:1   --accent-text on the primary button
  ok   19.07:1   --card-face-text on a revealed card
  ok    6.18:1   --bad on its own surface
  ok    7.18:1   --bad on the page

CONTROLS AND GRAPHICS — needs 3:1
--------------------------------------------------------------------------
  ok    3.28:1   --border as a control boundary on the page
  ok    3.08:1   --border as a control boundary on a panel
  ok    6.05:1   --accent as a focus ring on the page
  ok    5.69:1   --accent as a focus ring on a panel
  ok    3.28:1   the edge of a face-down card
  ok   19.92:1   a revealed card against the page
  ok   10.48:1   the connection dot
  ok    4.14:1   a tally bar against its track

DECORATION — measured, not required
--------------------------------------------------------------------------
        1.37:1   the fill of a face-down card
        1.22:1   --panel-border around a panel
        1.19:1   the table against the page
        1.22:1   a tally track against the table

Every measured pair meets the ratio the specification names.
```

The four pairs listed as decoration are measured but not required: the
specification asks for 3:1 from the boundary of a control, a focus indicator and
a graphic that carries meaning "rather than decoration". A panel is identified by
its fill and its content, the table is the surface a tally sits on rather than the
tally, the empty part of a bar says nothing the printed count does not, and the
fill of a face-down card is dark by design while its edge — measured as required,
at 3.28:1 — is what makes it perceivable as an object.

## Behaviours the specification fixes, checked in the running application

- Wide layout: seats sit on the ellipse around the table.
- Narrow layout at 390px: the table ends at y=442 and the seat list begins at
  y=462, below it.
- A name of exactly the maximum permitted length, `Maximiliane Bär`, is shown
  whole at both layouts.
- The deck is reachable along the bottom edge at both layouts.
- The tally shows per-card counts and a vote total, and nothing else: no average,
  no median.

## What was not run, and why

- **No Docker build.** The licence requirement says the licence must be inside
  the container image. That was verified one level down, by checking it is inside
  the embedded asset tree the image is built from and is served by the embedded
  binary — the image copies that binary and nothing else. Building the image was
  not run here.
- **No real screen reader.** The association between a control and its
  explanation is asserted structurally, by checking `aria-describedby` resolves
  to the element holding the text, and by keeping that element out of any state
  that would remove it from the accessibility tree. No assistive technology was
  driven.
- **No test on a physical touch device.** Touch was emulated by Playwright with
  `hasTouch` and a phone viewport.
- **No check in a browser other than Chromium**, which is what the browser suite
  runs and what the headless captures used.

## Departures from the plan, and why

- **`--border` moved from `#23272f` to `#5c6274`.** Direction B's value reached
  1.33:1 against the page where the specification requires 3:1 from the visible
  boundary of a control. design.md had already named the resolution: where a pair
  fails, the value moves rather than the requirement. `--card-back-border`
  followed it for the same reason. `docs/design/visual-language.html` records both.
- **Task 5.1 asked for a component test.** This project has no component-test
  setup — `npm test` is the plain Node runner over pure TypeScript in `lib/` — and
  its way of testing interactive components is the browser suite, with
  `throw-picker.spec.ts` as the precedent. The instrument changed; the coverage
  did not. All three ways of opening the control are tested.
- **One scenario was corrected rather than implemented.** The reduced-motion
  scenario said a thrown object must still be seen "and by whom". The interface
  has never shown who threw one, and `participant-throws` names the sender only to
  settle who may throw. The requirement exists so that removing movement does not
  remove information, and there was no attribution to lose; adding one would have
  been a change to the game rather than an accommodation. The scenario now says
  the object still appears at the seat it was aimed at.
- **`Decks` is not an injected handler.** It was one first, which made
  `NewRouter` panic for every caller that did not set the field — `legal_test.go`
  did not, and said so. It depends on nothing, so it is registered directly and
  there is no field left to forget.

## Observations, not fixed

Both predate this change and neither is in its scope.

- `RoomView.svelte` renders `<main class="waiting">` containing `NamePrompt`'s own
  `<main>`. Nested `main` landmarks are invalid HTML and confuse assistive
  technology. Found because a test locator matched two elements.
- On a narrow screen with enough participants, the sticky deck overlays the last
  row of the seat list. The page scrolls, so nothing is unreachable, but the
  partially veiled row looks unfinished.

## After merging main

`main` had moved twice while this branch was being built — the application
version on the entry screen, and a visitor mode for non-voting participants.
Both were merged in before verifying, because five of the files this branch
rebuilt were touched there too, and reviewing code about to be rewritten by a
merge is wasted work.

Checked before merging: the two requirements this change replaces in full, "The
entry screen offers one thing" and "A name is given before taking a seat", are
unchanged in main. Archiving therefore cannot silently undo anything main wrote.

Three conflicts, all in imports or in markup this branch had restructured. Two
things the merge brought in did not know about the new design and now do: the
visitor-mode placeholder had a literal copy of the deck veil, which is a custom
property since this branch, and the visitor hint had a hardcoded corner radius.

## Found by verifying, and closed

- **Three scenarios had no evidence behind them.** The behaviour was right in
  every case; nothing would have noticed if it stopped being. The t-shirt deck
  being preselected is one of this change's own scenarios and now has a test for
  both paths, with the deck list and without it — this matters more than it did,
  because the option list is now built from what the server sends. The version
  line is asserted too: it belongs to another change, but this branch moved the
  element it sits below.

- **Escape dragged the focus onto the information control.** Escape is listened
  for on the window, and the bubble can be open purely because a pointer rests on
  the trigger, so pressing Escape while typing a name pulled the focus out of the
  field. The pattern came from the throw picker, where closing must hand the focus
  back because that picker contains buttons; this disclosure contains a sentence
  and nothing else. Fixed, with a test.

- **One browser test asserted a number rather than its claim.** Visitor mode
  changed the seat's geometry, so a settled throw lands 61px from a seat centre
  where the test allowed 52px, one seat width. The behaviour was unchanged. The
  test now says what the requirement says: within about a seat's reach of that
  seat, and nowhere near a viewport edge, which is where a flight would begin.

## Open, and deliberately not decided here

The interface now has **two different information controls**. This change built
one, and the visitor-mode change independently built another with the same job:
an "i" beside a checkbox, `aria-describedby`, a note that opens on hover, focus
and click. They appear three rows apart on the join screen and do not look alike
— one is a drawn circle at the right edge of its label, the other the letter "i"
in a bordered circle directly after its text.

The visitor one also keeps its note at `visibility: hidden` while closed, which
removes it from the accessibility tree, so the `aria-describedby` on that
checkbox resolves to nothing until the button is pressed. This change's own
control avoids that deliberately and says why in its source.

Consolidating them would mean rewriting part of a feature that was specified and
archived a day ago. That is a decision for the owner, not a thing to do quietly
inside a redesign.
