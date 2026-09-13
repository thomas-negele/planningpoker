## Context

See `proposal.md` for motivation. What shapes the approach is three existing constraints.

The visual design is scattered. `web/src/app.css` holds 28 custom properties, all of them colours,
while every radius, every shadow and the table's own surface sit hardcoded inside the components —
2,818 lines of component CSS across seventeen files. A colour change is one file; a change to the
shape of a card is nine.

Nothing may be fetched from a foreign host, and the browser enforces it: `app-delivery` requires a
Content-Security-Policy with `font-src 'self'`. A bundled typeface therefore needs no change to the
policy, and if it ever did, that would be the signal that it was being done wrong.

The server is the only place a deck is defined. `internal/game/deck.go` holds both decks, and
`table-ui` already requires the deck at the table to be built from what the server sends rather than
from a list in the page. The entry screen is the one place that has so far escaped this, because it
only needed labels — and it is exactly the place that now needs card values.

The chosen appearance and the alternatives that were rejected are recorded in
`docs/design/visual-language.html`, which stays in the repository as the reasoning behind the
palette. It is not authoritative: `app.css` is.

## Goals / Non-Goals

**Goals:**

- Put the whole visual design behind custom properties in one file, so that the next change to it is
  a change to values rather than a search across components.
- Ship the typeface as part of the application, identical on every operating system.
- Make both changed screens legible without hover and without a mouse.
- Keep the server the only definition of a deck while showing card values before a room exists.

**Non-Goals:**

- No light theme. The palette is dark, and a second one doubles the surface to verify for contrast.
- No design-token build step, no CSS framework, no utility classes. Custom properties in one
  stylesheet are enough at this size, and a build step would be a dependency to maintain.
- No change to the layout of the table, the ellipse of seats, the 58rem breakpoint, or anything the
  behaviour specifications fix. The empty table during a running round is a known weakness and
  stays out of this change; it is named in the design document and would need its own proposal.
- No subsetting pipeline for the font. The latin subset is downloaded once and committed.

## Decisions

### The visual design becomes custom properties, added rather than renamed

Radii, shadows, the table surface and the card back move out of the components into `app.css` as new
custom properties. The existing colour property names keep their names and change only their values,
so a component referring to `--surface` needs no edit.

Alternative considered: renaming the colour properties to something more systematic
(`--color-surface-raised` and so on). Rejected — it would touch every component for no behavioural
gain and would make the diff of this change unreadable, hiding the parts that matter inside a
rename.

Alternative considered: a design-token file compiled into CSS. Rejected as a dependency and a build
step for seventeen components.

### The typeface is a committed variable woff2, with the system stack behind it

Geist ships as one variable `.woff2` covering weights 400–700 in the latin subset, roughly 29 KB,
declared with a local `@font-face` and placed at the front of the existing `--font-stack`. The
system stack stays behind it as the fallback, so a failed font load degrades to exactly today's
appearance rather than to a serif default.

`font-display` is set to `swap`: the text is readable in the fallback face while the font loads, and
reflows once. The alternative, `block`, hides the text for up to three seconds, which on a
join screen means an invisible form.

Alternatives considered and rejected: Inter, which is the safer and more common choice but carries
no voice and is 47 KB; IBM Plex Sans, which has more character than a tool wants to look at all day.
Instrument Sans was ruled out on a fact rather than a preference — it has no `½` glyph, and the
Fibonacci deck contains that card, so exactly one card in the deck would render in a substituted
typeface. Each candidate's rendering is in `docs/design/visual-language.html`.

Geist's tighter default tracking is a known cost in the smallest uppercase labels, the `away` tag in
particular; those get explicit letter-spacing rather than a different font.

### `GET /api/decks` reads the decks the server already defines

A new read-only route in `internal/transport/`, answering from `game.TShirtDeck()` and
`game.FibonacciDeck()` — the same functions room creation calls. It creates nothing, sets no cookie
and takes no parameters.

Alternative considered: hardcoding the card lists in `web/src/lib/decks.ts` beside the labels that
already live there. Rejected by the owner, and rightly: the entry screen would then promise cards
that a room might not deal, and nothing would fail when they diverged. The cost of the endpoint is
one handler, one route and its tests.

Alternative considered: embedding the deck list into the served HTML at build time. Rejected — it
moves the duplication into the build instead of removing it, and the disk-served development path
and the embedded production path would then differ.

The entry screen fetches the list on load and renders the options without card values until it
arrives. If the request fails, the options keep their labels and starting a game still works; this
is required by the spec rather than left to chance, because a broken endpoint must not cost anybody
their meeting.

### One information control, used twice

A single small component renders the `i` control and its text, used at the name field and at the
storage choice. It discloses on hover, on focus and on click or tap, closes on `Escape` and on a
click outside, and is associated with the control it explains through `aria-describedby`, so the two
are announced together rather than as an unexplained icon.

Alternative considered: the native `title` attribute. Rejected — it does not appear on touch, it
does not appear on keyboard focus, its delay is not controllable, and screen readers treat it
inconsistently. It fails the requirement almost exactly.

Alternative considered: leaving the texts visible and merely making them smaller. Rejected because
the owner asked for the opposite, but the concern behind it is answered in the specification: the
storage choice's own label must still state that the name would be kept on this device, so the part
that carries the consent is never the part that is hidden.

### Contrast is measured, not asserted

The palette of direction B is checked pair by pair against the ratios the specification now names,
and the measured figures are recorded in the change's verification notes. Where a pair fails, the
value in `app.css` moves — not the requirement.

## Risks / Trade-offs

- **The palette was designed in a mirror of the application, not in the application.** The template
  reproduces the components' structure and measurements, but it is a reproduction. → Verify in the
  running application at both layouts, wide and narrow, before ticking the tasks; the mirror has
  already produced one such error, where the deck collapsed because a `<span>` cannot take a height.

- **A bundled font is a permanent 29 KB on every first load.** → Accepted deliberately; it is
  cached thereafter, it is smaller than most single images, and the alternative is an interface that
  looks different on every operating system.

- **`font-display: swap` reflows once when the font arrives.** → Accepted. The alternative hides
  text, and on the join screen that means a form nobody can read.

- **Moving consent-relevant text behind an icon reduces how many people read it.** → Mitigated by
  the specification: the storage choice's own label must say that the name is kept on this device,
  the information must be reachable by keyboard and touch as well as hover, and nothing is stored
  unless the choice is explicitly selected. The residual risk is real and was accepted knowingly.

- **A new endpoint is new surface.** → It is read-only, takes no parameters, creates nothing and
  reads constants; the bounded-resources requirement in `app-delivery` is unaffected because no room
  or connection is involved.

- **Reduced motion touches the throws feature**, which is entirely animation. → The requirement is
  written so that the information survives without the movement: the object still arrives and is
  still attributed. Verify with the system setting actually enabled, not by reading the code.

## Migration Plan

There is nothing to migrate. No stored value, protocol message, URL or configuration key changes,
and `pp_name` and the seat cookies keep their meaning. A browser that is connected while the new
version is deployed loses its room to the restart exactly as it does for any other deployment — a
consequence of holding state in memory, unchanged by this work.

Rollback is a redeploy of the previous image.
