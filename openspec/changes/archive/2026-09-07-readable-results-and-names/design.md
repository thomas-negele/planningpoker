## Context

See `proposal.md` — Why. What follows is only the current state that constrains the
approach.

**The result display today.** `web/src/components/Results.svelte` renders
`results.tally` as a flex row of pills that wrap. The tally arrives in the deck's own
order: `internal/game/view.go` documents that `Results.Tally` counts each card "in the
deck's own order", and the t-shirt deck's order is `XS, S, M, L, XL, ?, ☕`. The chart
presents the sizes the other way up — largest at the top, smallest at the bottom — so
the client reverses the scale for display. That is a presentation choice and is made in
the client; the deck's own order stays as it is, because it is also the order the cards
are offered in along the bottom of the screen, where small-to-large is right. Cards
nobody played are absent from the tally rather than present with a count of zero.

**The felt.** In `RoomView.svelte` the table is `.arena { width: min(24rem, 52vw);
aspect-ratio: 16 / 9 }`, so 24 × 13.5rem at full size. Inside it: `padding: 1.5rem`, a
`.info` slot with `min-height: 3.6rem`, a `0.9rem` gap, and a `.action` slot of
`2.4rem` that holds Reveal and keeps its height after Reveal is hidden. That leaves
`13.5 − 3 − 0.9 − 2.4 = 7.2rem` of height for the chart, and the shape has a
`border-radius: 48% / 34%`, so the top and bottom of that box are inside a curve and
narrower than the box suggests.

**The seats.** `seatPosition()` places each seat on the table's edge and then pushes
it outwards by `SEAT_HALF_WIDTH = 2.1` or `SEAT_HALF_HEIGHT = 2.95` rem depending on
direction, plus `SEAT_GAP = 1.1`. Those two constants are a *promise about how big a
seat is*. `Seat.svelte` currently keeps that promise by force: `.who { max-width:
8rem }` and `.name { overflow: hidden; text-overflow: ellipsis; white-space: nowrap }`.
That is the bug — the geometry is honoured by mutilating the name.

**The name limit.** `game.MaxNameLength = 40`, enforced in `normalizeName`, surfaced
as `ErrNameTooLong` → `name_too_long` → "That name is too long. Please use a shorter
one." The tests build their fixtures from the constant (`strings.Repeat("a",
MaxNameLength+1)`), so they follow it down without editing.

## Goals / Non-Goals

**Goals**

- The revealed round is legible as a shape, at a glance, from across a meeting room.
- The page learns which cards form the ordered scale from the server, so the
  `table-ui` promise that "the deck comes from the server" survives this change.
- One name limit, defined once, enforced by the server, met by the field while typing,
  and honoured by the layout without truncation.
- The seat geometry stays *derived* — the push constants continue to describe the real
  extent of a seat rather than a size the seat is then clipped into.

**Non-Goals**

- No second deck, and no mechanism for choosing between decks. The scale information
  is added because this feature needs it, not to prepare for Fibonacci.
- No configurable name limit and no administration screen for it. Reasoning below.
- No change to what the server sends about a *hidden* round. The tally has always
  lived in `Results`, which is `nil` until reveal, and nothing here moves it earlier.
- No change to how each participant's own card is shown at their seat.

## Decisions

### 1. The server says which cards are on the scale; the page does not guess

**Decision.** `game.Deck` gains a second ordered list, `Scale []Card`, holding exactly
the cards that express a size. For the t-shirt deck that is `XS, S, M, L, XL`. It goes
on the wire as `deck.scale` alongside the existing `deck.cards`, and
`web/src/lib/protocol.ts` gains the matching field. `Results.svelte` renders one row
per entry of `deck.scale`, then puts everything in `deck.cards` that is *not* in
`deck.scale` on the shared row below.

**Why not hardcode `?` and `☕` in the page.** It is two lines and it is wrong. The
`table-ui` requirement "The deck is always to hand" says the deck is built "from what
the server sends rather than from a list written into the page, so that a room
offering a different deck later needs no change here". A hardcoded exception list
would make that false the day a second deck arrives, and it would fail quietly — a
Fibonacci deck would render `?` on the scale between two numbers, which is exactly the
false ordering this feature exists to avoid.

**Why a separate list rather than a flag per card.** `deck.cards` is `Card[]` — a flat
array of strings — and `Deck.svelte` iterates it directly. Turning it into an array of
objects would touch the deck strip, the vote message and every test that constructs a
deck, for no gain. A second list is additive: nothing that reads `cards` today changes,
and `Deck.Contains` is untouched.

**Why the field is empty-safe.** A deck with no scale at all (every card unordered) is
representable and renders as no size rows and one shared row. Nothing in the code needs
to special-case it; it simply falls out.

### 2. Zero rows are shown; the scale is fixed

**Decision.** Every card in `deck.scale` gets a row in every revealed round, whether or
not anybody played it. The client fills gaps against the tally rather than the server
padding the tally with zeroes.

**Why the client and not the server.** `view.go` states deliberately that "cards nobody
played are absent rather than present with a count of zero". That is a statement about
what the round *was*, and it is correct. Padding it with zeroes would make the data
answer a presentation question. The client has `deck.scale` and the tally and can join
them; the server keeps saying what happened.

**Why zero rows at all.** Three reasons, in order of weight. It fixes the height of the
chart, which is what keeps the felt from resizing between rounds and between the hidden
and revealed states. It makes two consecutive rounds on the same story comparable —
"we moved from S/M to M/L" is visible as the shape shifting down the scale, which is
invisible if the rows themselves come and go. And it is neutral: a chart that lists
only the chosen cards implicitly presents them as the candidates, where a full scale
presents the whole range and lets the gaps speak.

**Why the `?`/`☕` row is *not* always present.** It is the owner's stated reason —
those are chosen rarely and a permanently reserved row costs felt on every round for a
case that is usually empty. It is one row, so its coming and going costs one row of
height; the table is sized for the case where it is present, so the chart never
outgrows the felt, it just sits slightly higher when the row is absent.

### 3. Nothing is declared the winner

**Decision.** No highlight on the longest bar, no "most chose M", no consensus badge,
no outlier marking. Bars are all one colour. The existing summary line
("Everyone who voted chose the same card." / "3 votes.") stays as a plain statement of
fact — it counts voters, it does not judge cards.

**Why.** The owner's instruction: interpretation belongs to the human moderator. It is
also the same principle that already forbids an average here — the interface's job is
to present the round faithfully, and a round whose disagreement has been resolved by
the software before anybody spoke has had its purpose removed. A highlight is not a
neutral visual aid; it is a verdict rendered in colour.

**Consequence for implementation.** Do not reach for a "primary/highlight" colour token
when styling the bars, and do not sort by count. Both would smuggle the verdict back in.

### 4. The table grows, once, for the fullest chart

**Decision.** The chart's worst case is five size rows plus one shared row. At a row
height of roughly `1.15rem` with `0.25rem` between rows, that is about `6.75rem` for the
scale, plus a gap and the shared row: call it `8.3rem`, plus the summary line. Against
today's `7.2rem` of usable height — and less than that in practice, because the top and
bottom of that box lie inside the felt's curve — it does not fit.

The table therefore grows. **Measured outcome:** `.arena` is `32rem` at `16 / 9`
(`512 × 288px`), vertical padding `1.9rem`, and the chart is capped at `17rem` wide.
The width cap is not decoration — the felt is a rounded shape, so near the top edge it
is much narrower than its bounding box, and a row spanning the full inner width had its
ends hanging over the curve.

**The width is pinned, not a share of the viewport, and that reverses this document's
first instinct.** A `vw`-based width shrinks as the window narrows, and a *smaller*
table is exactly what the chart cannot survive: at `56vw` the table fell to `414px`
near the breakpoint and the chart spilled out again. Since the felt is `16 / 9`, the
chart's `~16.6rem` of height forces a minimum width of about `29.5rem`, so the size is
fixed and the breakpoint below is what handles a screen too small to hold it.

**One size, both states.** The larger table applies whether the round is hidden or
revealed. `.info` keeps a `min-height` sized for the chart, exactly as it already does
at `3.6rem` for the pills today. The seats are positioned against `.arena`, so a table
that grew on reveal would slide every participant outwards at the moment somebody
pressed the button.

**The reveal control comes out of the flow, so the result sits at the table's centre.**
With the button in the normal column the chart measured `34px` of space above it and
`87px` below — it was `26px` above the table's middle and read as clinging to the top
edge. The button still needs its reserved height so that revealing moves nothing, so
the reservation is made by pinning it near the foot of the felt (`position: absolute`)
rather than by having it occupy a row that displaces the chart. The result then centres
on the table's own middle, measured at `70px` above and `70px` below. This applies in
the ring layout only; in the list layout the table hugs its content, so the ordinary
flow already produces the right thing and the button stays in it.

**These numbers are a starting point, not a result.** Row height depends on the font,
and the felt's curve makes the usable area smaller than the box. The acceptance test is
the spec scenario — a round in which every size and both non-size cards were played,
rendered with nothing clipped or overlapping — checked in a browser at a few widths,
not arithmetic on this page.

**The breakpoint moves from `45rem` to `58rem`, and it had to.** Measured in the
browser, the ring needs `table (32rem) + 2 × (push 7rem + half a seat 5.9rem)`
≈ `57.8rem`. At the old `45rem` breakpoint the ring was therefore switched on at a
width where it could not fit — the table and the seats both grew, for the chart and for
the names respectively, and the old threshold predates both. `58rem` (`928px`) leaves a
little slack, verified at `940px` with worst-case names.

Note that the remedy this section originally proposed — "lower the `vw` factor rather
than shrinking the seats" — is wrong, and it is left recorded here rather than quietly
deleted. It treats the table as the flexible part, when the table is the part with a
hard minimum. The seats are not the problem either. The only thing that can actually
give is *which layout is used at that width*, which is what moving the breakpoint does:
below `58rem` the list layout applies, and it has no ring geometry to violate.

**The narrow layout needed the same correction.** Its table was `16 / 9` at
`min(22rem, 90vw)`, about `12rem` tall against a chart needing `16rem`. Since nothing
is positioned against the table when the seats are a list, the fix is to drop the fixed
proportion there (`aspect-ratio: auto`, with a `min-height`) and let the felt be as tall
as what it holds.

### 5. Fifteen characters, and the layout is what pays for it

**Decision.** `game.MaxNameLength` becomes `15`. The comment above it, which currently
argues that forty "is comfortably more than any real name needs and comfortably less
than anything that would break a row of seats", is now false in its second half and
must be rewritten to say what the new reasoning is: the limit and the ring geometry are
chosen together, and fifteen is what a seat can display in full — at the cost of a ring
that needs a wider screen before it is offered at all.

In `Seat.svelte`, `.who { max-width: 8rem }` and the ellipsis rules go. The name sizes
itself; `.who` gets no maximum. The rename `input` widens from `7rem` to `12rem` to hold
fifteen characters comfortably. Both fields take their `maxlength` from a single
`MAX_NAME_LENGTH` in `web/src/lib/name.ts`, mirroring the Go constant, so the two cannot
drift apart from each other.

`SEAT_HALF_WIDTH` rises from `2.1` to `5.9rem`, and that figure is measured rather than
estimated: a seat holding fifteen capital Ms — the worst case the rules permit — comes
to `186px`, so half of it is `5.9rem`. An earlier pass budgeted `5.1rem` from a typical
name and was wrong by `23px`; it did not collide at `1440px` only because there was
slack, which is exactly the kind of latent break this change exists to remove. It is
deliberately not measured at runtime — feeding a resize observer back into the position
would couple layout to a value that varies by a few tenths of a rem.
`SEAT_HALF_HEIGHT` is unchanged; nothing about the seat got taller. `.table-area`'s side
padding goes from `9rem` to `7.5rem` to match the new push.

**Why the limit is a constant and not a setting.** The owner's global rule puts
behaviour-governing values in an administration area. `CLAUDE.md` resolves the tension
by requiring the value to be surfaced rather than silently decided — which is what
happened: ten was chosen explicitly. It stays a constant because it is not independently
adjustable. Raising it to twenty would not produce a working product with longer names;
it would produce truncated names or a broken ring, because `SEAT_HALF_WIDTH`, the
table's width and the breakpoint were all chosen against it. A setting that cannot be
changed without changing three other things is a trap, not a setting. If longer names
are ever wanted, that is a change to the table's layout, and the constant moves with it.

**Why not wrap the name onto two lines instead.** Wrapping would keep seats narrow and
honour "displayed in full". But a ten-character name is usually one word, so wrapping
breaks it mid-word, and a seat that is sometimes one line tall and sometimes two would
make `SEAT_HALF_HEIGHT` wrong for half the table. Widening is the honest fix.

### 6. The field's cap does not become the page's own rule

**Decision.** `maxlength` (from `MAX_NAME_LENGTH`) prevents the common case. The submit path is otherwise
unchanged: the page still sends the name and still displays the server's refusal. The
existing courtesy check in `NamePrompt.submit` — reject empty — stays exactly as
generous as it is; nothing is added that decides a name is *acceptable*.

**Why this matters.** `table-ui` requires that "the page may check first as a courtesy,
but the server's refusal is what is displayed". `maxlength` does not stop a paste, an
autofill or a browser that ignores it, so the refusal path stays live and must keep
working. The temptation to add `if (trimmed.length > 10) return;` should be resisted:
it would replace a message that explains the problem with a button that silently does
nothing. Verified in the browser: a name pasted past the field's cap is sent, refused by
the server, and the refusal is what the visitor sees.

## Risks / Trade-offs

- **Fifteen characters will still be too short for somebody**, and it now costs screen
  width at the other end: the ring needs about `58rem` (`928px`), so a window narrower
  than that gets the list layout rather than the table with people around it. → It is
  the owner's explicit choice, made with the trade-off in front of them. The refusal message already names
  the limit, so the failure is legible rather than mysterious. Raising it later means
  revisiting decision 5's geometry, which is why the reasoning is written down there
  rather than left implicit in a constant.

- **A session in progress when this deploys will refuse renames that used to work.**
  → There is no persistence, so nothing is stored to migrate, and an already-seated
  participant keeps the name they have; only a *new* join or rename is checked. A
  deployment during a meeting already destroys every room (`CLAUDE.md`, "Persistence:
  none"), so this adds nothing to a risk that is already understood.

- **The bigger table may crowd the ring on a laptop, particularly just above the
  `45rem` breakpoint.** → Named in decision 4 as the first thing to check in a browser,
  with the stated remedy (lower the `vw` factor, do not shrink the seats).

- **The row-height arithmetic in decision 4 is an estimate.** → The spec scenario is the
  test, not the arithmetic. If the chart does not fit, the table grows further; the
  requirement is written as "where the table is not large enough, it SHALL be made
  larger" precisely so that the number is not the contract.

- **`deck.scale` is a field two of the three consumers do not use, which invites the
  question of why it exists.** → Answered in the field's own comment on both sides of
  the wire. Without it the page must hardcode `?` and `☕`, which breaks a requirement
  the project already made.

- **Full-length names could still collide at the diagonals of the ring.** The push is
  computed from a single worst-case half-width in each axis, and the diagonals are where
  seats are closest to one another. → Covered by the spec scenario that requires no seat
  to overlap another with every participant at maximum name length; check it with a
  table of six or seven people, which is where the ring is densest.

## Migration Plan

No data to migrate and no staged rollout: one deployment replaces the whole application,
and rooms do not survive a restart in any case. Rollback is redeploying the previous
image. The only externally visible discontinuity is the name limit, and it applies from
the first request after the restart.

## Open Questions

None. The decisions that were genuinely the owner's were all put to them and answered:
the limit of fifteen characters, the absence of a declared winner, the scale running
largest-to-smallest, and the result sitting at the table's centre. The last three came
during implementation rather than planning, and this document was revised to match
rather than left describing a version that no longer exists.
