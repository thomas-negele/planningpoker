## Why

Two things about the table are harder to read than they need to be.

The first is the result of a revealed round. Today it is a row of small pills — a
card and a `×3` beside it — that wraps across the felt. That shape says *which*
cards were played but makes the spread hard to see: three votes and one vote look
almost identical, and the eye has to read numbers instead of shapes. A round of
planning poker is a conversation about disagreement, so the display should make the
disagreement visible at a glance.

The second is a plain bug. A participant's name is cut off at the seat — the
screenshot shows `Magdalena Schmi` where `Magdalena Schmidt` was typed. The cause is
not the rule but the layout: the server permits 40 characters, and the seat's name
box is fixed at `8rem` with `text-overflow: ellipsis`, so anything longer is silently
clipped. Somebody who typed a name they were allowed to type sees it mangled, with
nothing anywhere telling them why.

## What Changes

**Results as a bar chart on the table**

- After a reveal, the table SHALL show the tally as horizontal bars stacked
  vertically instead of a wrapped row of pills. Bar length is proportional to how
  many people played that card.
- The size cards run top to bottom along the deck's ordering of sizes, largest at the
  top and smallest at the bottom (`XL, L, M, S, XS`), so the scale reads with the big
  work above the small. All five SHALL be present in every revealed round, including
  those nobody played, so the chart is a fixed scale whose shape can be compared
  between rounds rather than a list that changes height.
- `?` and `☕` are not sizes and do not belong on that scale. They SHALL appear
  beneath it, side by side on a single shared row, and only when at least one of
  them was actually played. They are rarely chosen and should not cost two rows of
  felt on every round.
- Nothing on the chart SHALL be marked as the winner, the majority, the consensus or
  any other verdict. No highlighting of the tallest bar, no "most chose M" line, and
  as before no average or median. Reading the distribution is the human moderator's
  job, and an interface that names a winner quietly ends the discussion the round
  exists to start.
- The table SHALL be large enough that the fullest possible chart fits inside the
  felt, readable, without shrinking text and without overflowing onto the seats. If
  the current table is too small for that, the table gets bigger.
- The table SHALL NOT change size between the hidden state and the revealed state,
  because the seats are positioned relative to it and would all jump the moment
  somebody pressed Reveal.
- What the table displays SHALL sit at the table's own centre. The Reveal control
  keeps its reserved space so that revealing moves nothing, but that reservation must
  not push the result upwards off the middle of the felt.

**A name limit that the layout can actually honour**

- **BREAKING (for anyone mid-session at the moment of deployment):** the maximum
  display name shrinks from 40 characters to **15**. Names between 16 and 40
  characters, which are accepted today, will be refused. Nothing is stored anywhere,
  so there is no data to migrate; the practical effect is that somebody used to
  typing `Magdalena Schmidt` (17) must shorten it.
- The limit SHALL be enforced by the server, as it is now, and the name field SHALL
  also stop accepting further characters at 15 so that the limit is met while typing
  rather than announced after submitting. The page still displays the server's
  refusal when one arrives; the field's own cap is a courtesy, not a second rule.
- A name of the full 15 characters SHALL be shown complete at the seat, in the name
  prompt's list of who is already here, and in the rename field. No ellipsis, no
  clipping, at any screen width.
- The ring of seats SHALL still clear the felt once seats are wide enough to hold a
  full name, so that widening a seat does not put somebody's card back on the table.

## Capabilities

### New Capabilities

None. Both changes alter behaviour that existing capabilities already describe.

### Modified Capabilities

- `table-ui`: the requirement "Results show the cards and the count per card" gains
  the bar-chart form, the fixed size scale, the shared `?`/`☕` row and the explicit
  prohibition on declaring a winner; the table is required to be big enough for it
  and to keep one size across both states. The requirement "A name is given before
  taking a seat" gains the field's own 10-character cap. The requirement "The table
  shows everyone and what they have done" gains the guarantee that a name of the
  permitted length is displayed in full rather than truncated.
- `room-membership`: the requirement "A participant takes a seat under a name" names
  the maximum as 15 characters, where it previously spoke only of "the permitted
  maximum", and states the consequence that pins it to the table's geometry.

## Impact

- `internal/game/room.go` — `MaxNameLength` changes from `40` to `15`, and its
  comment (which currently argues for forty) must be rewritten to say why fifteen is
  the number and what it costs. `internal/game/room_test.go` and `internal/game/view_test.go` build names
  from that constant and continue to work unchanged.
- `internal/game/deck.go`, `internal/game/view.go` — the deck view needs to say which
  of its cards form the ordered size scale and which do not, because the client must
  not hardcode that `?` and `☕` are the odd ones out. `table-ui` already requires the
  deck to come from the server so that a second deck can be added later without
  touching the client, and a hardcoded exception list would break that promise.
- `internal/transport/protocol.go`, `web/src/lib/protocol.ts` — the deck's wire shape
  gains that information.
- `web/src/components/Results.svelte` — rewritten from pills to bars.
- `web/src/components/RoomView.svelte` — table dimensions, and the seat half-extents
  the ring geometry is computed from.
- `web/src/components/Seat.svelte` — the truncating name box, and the rename field's
  width and `maxlength`.
- `web/src/components/NamePrompt.svelte` — `maxlength` on the field, and the "already
  here" list.
- No change to the hidden-vote guarantee: the tally has always been part of `Results`,
  which is `nil` until the round is revealed, and nothing here moves a card value
  earlier.
