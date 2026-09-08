## 1. The name limit in the domain

- [x] 1.1 Change `MaxNameLength` in `internal/game/room.go` from `40` to `15`, and
      rewrite the comment above it: the present one argues that forty is "comfortably
      less than anything that would break a row of seats", which is the claim this
      change disproves. The new comment states that the limit and the geometry of the
      table are chosen together, that ten is what a seat can display in full without
      truncation, and that this is why it is a constant rather than an adjustable
      setting (see `design.md` — decision 5). Verify with `go build ./...`.
- [x] 1.2 Run `go test ./internal/game` and confirm the existing name tests still pass
      untouched. They build their fixtures from the constant
      (`strings.Repeat("a", MaxNameLength+1)` in `room_test.go` and `view_test.go`), so
      they should follow the new value down with no edit. If any test hardcodes `40`,
      that is a fixture to fix, not a reason to keep the old limit.
- [x] 1.3 Add a test in `internal/game/room_test.go` that joining with exactly fifteen
      characters succeeds and that sixteen is refused with `ErrNameTooLong`, naming the
      two boundary cases explicitly rather than relying on the constant-derived
      fixtures alone. Verify with `go test ./internal/game -run TestJoin`.
- [x] 1.4 Confirm nothing else in the repository hardcodes the old limit:
      `grep -rn "40" internal/ web/src/ --include=*.go --include=*.ts --include=*.svelte`
      and check each hit. The wire code and the refusal message
      (`web/src/lib/protocol.ts`, `name_too_long`) speak of the limit without naming a
      number and need no change; verify that is still true.

## 2. The deck's size scale on the wire

- [x] 2.1 Add `Scale []Card` to the `Deck` struct in `internal/game/deck.go` and
      populate it in `TShirtDeck()` with `XS, S, M, L, XL` — the cards that express a
      size, excluding `?` and `☕`. Return a fresh slice on every call, as `Cards`
      already does, so a caller cannot reorder the scale every other room is using.
      Document on the field that it exists so the client need not be taught which
      cards are sizes (see `design.md` — decision 1). Verify with `go build ./...`.
- [x] 2.2 Add a test in `internal/game/deck_test.go` asserting that `TShirtDeck().Scale`
      is exactly `XS, S, M, L, XL` in that order, that every entry of `Scale` is also
      in `Cards`, and that mutating a returned `Scale` does not affect the next call.
      Verify with `go test ./internal/game`.
- [x] 2.3 Extend `deckMessage` in `internal/transport/protocol.go` with
      `Scale []string \`json:"scale"\`` and populate it where `Cards` is populated
      (around `internal/transport/protocol.go:220`). Verify with
      `go test ./internal/transport`, extending the existing protocol encoding test so
      it asserts the field is present and correct rather than only that the message
      still encodes.
- [x] 2.4 Add `scale: Card[]` to the `Deck` interface in `web/src/lib/protocol.ts`,
      with a comment saying what it is for, keeping the file's stated discipline that
      every field mirrors the Go side by name and shape. Verify with
      `npm run check` from `web/`.

## 3. The results bar chart

- [x] 3.1 Rewrite `web/src/components/Results.svelte` to take the deck as well as the
      results, and to derive its rows by joining `deck.scale` against `results.tally`:
      one row per scale card, in the deck's order, with a count of zero where the tally
      has no entry. Do not sort by count and do not pad the tally on the server — the
      server keeps reporting only what was played (`design.md` — decision 2). Verify by
      revealing a round in which one size was unplayed and seeing its row present with
      an empty bar.
- [x] 3.2 Render each row as card label, a proportional horizontal bar, and the exact
      count in text. Scale bar length against the largest count in the round so the
      shape uses the available width. All bars share one colour: no highlight on the
      longest, no marking of a mode, majority or outlier (`design.md` — decision 3).
      Verify by revealing an uneven round and confirming that nothing distinguishes the
      tallest bar but its length.
- [x] 3.3 Render the cards that are in `deck.cards` but not in `deck.scale` — `?` and
      `☕` — beneath the scale on one shared row, side by side, and omit that row
      entirely when neither was played. Derive the set by difference against
      `deck.scale`; do not write `'?'` or `'☕'` into the page (`design.md` —
      decision 1). Verify by revealing one round with `?` played and one with neither,
      and confirming the row appears in the first and is absent in the second.
- [x] 3.4 Keep the existing summary line ("Nobody voted in this round." / "Everyone who
      voted chose the same card." / "N votes.") unchanged. It counts voters and passes
      no judgement on cards, so it survives the no-winner rule. Verify by reading it in
      each of the three cases.
- [x] 3.5 Pass the deck through from `web/src/components/RoomView.svelte`, which already
      holds `room.deck`. Verify with `npm run check`.
- [x] 3.6 Order the size rows largest at the top and smallest at the bottom (`XL, L,
      M, S, XS`), reversing the deck's own small-to-large order for display only. The
      deck's order is unchanged, because it is also the order the cards are offered in
      along the bottom of the screen, where small-to-large is right. `?` and the coffee
      cup stay below every size row. Verified: rows render `XL, L, M, S, XS` then the
      shared row.

## 4. The table and the seats

- [x] 4.1 Enlarge `.arena` in `RoomView.svelte` at the same `16 / 9`, raise the table's
      vertical padding so the first and last chart rows are not pressed against the
      felt's curve, and raise `.info`'s `min-height` to the chart's full height. Verify
      that the table is the same size before and after pressing Reveal and that no seat
      moves. **Done:** `.arena` is a fixed `32rem` (`512 × 288px`) — pinned rather than
      `vw`-based, because a shrinking table is what the chart cannot survive — with
      `1.9rem` vertical padding, `.info` at `min-height: 9.5rem` and capped at `17rem`
      wide so no row overhangs the felt's curve. Measured: table `512 × 288` identical
      before and after Reveal, no seat moved, no row outside the felt, no clipping.
- [x] 4.2 Remove the truncation in `web/src/components/Seat.svelte`: drop
      `.who { max-width: 8rem }` and the `overflow: hidden` / `text-overflow: ellipsis`
      / `white-space: nowrap` rules on `.name`, so the name sizes itself. Verify by
      seating somebody under a ten-character name on a wide screen and again on a
      narrow one and reading every character.
- [x] 4.3 Widen `input.rename` in `Seat.svelte` from `7rem` to `12rem` so a
      fifteen-character name is wholly visible while being edited, and give it a
      `maxlength` from the shared `MAX_NAME_LENGTH`. Verified: no name clipped in the
      field at any tested width.
- [x] 4.4 Raise `SEAT_HALF_WIDTH` in `RoomView.svelte` from `2.1` to the worst-case
      half-extent of a seat at maximum name length, leave `SEAT_HALF_HEIGHT` alone since
      nothing about the seat got taller, and set `.table-area`'s side padding to match
      the new push. **Done: `5.9rem`, measured not estimated.** A seat holding fifteen
      capital Ms is `186px` wide; an earlier `5.1rem` budget taken from a typical name
      was `23px` short and survived only on slack. Padding `9rem` → `7.5rem`. Verified
      with seven participants at maximum name length, including the all-capital-M worst
      case: no card on the felt, no seat overlapping another, none off-screen.
- [x] 4.8 Raise the name limit from ten to fifteen at the owner's request, and carry the
      geometry with it: `MaxNameLength = 15`, `MAX_NAME_LENGTH = 15`,
      `SEAT_HALF_WIDTH = 5.9`, rename field `12rem`, side padding `7.5rem`, breakpoint
      `58rem`. The limit is not independently adjustable — this is the list of things
      that move with it. Verified at 1440, 940, 900 and 606 px with worst-case names.
- [x] 4.5 Check the ring layout at just above the breakpoint, which `design.md` —
      decision 4 names as the width most likely to break. Verify by resizing the
      browser through it and watching for overlap. **Done, and the predicted remedy was
      wrong.** Lowering the `vw` factor shrinks the table, which is precisely what makes
      the chart spill; measured, the ring needs `47.8rem` while the old breakpoint
      offered it at `45rem`. Fixed by pinning the table's width and moving the
      breakpoint (now `58rem`, after the limit rose to fifteen) in `RoomView.svelte`
      and `Seat.svelte`. Audited at 1440, 940, 900, 606 px: no overlap, nothing off-screen, no horizontal scroll.
- [x] 4.6 The narrow layout needed the same correction: its table was `16 / 9` and so
      about `12rem` tall against a chart needing `16rem`. Since nothing is positioned
      against the table when the seats are a list, `.table` there is now
      `aspect-ratio: auto` with a `min-height`, letting the felt be as tall as what it
      holds. Verified at 500px: nothing clipped, no row outside the felt.
- [x] 4.7 Centre what the table displays on the table's own middle. Measured before:
      `34px` of space above the chart and `87px` below, because the reveal control's
      reserved row sat beneath it and pushed it up. The control keeps its reserved
      height — revealing must still move nothing — but in the ring layout it is now
      pinned near the foot of the felt out of the flow. Measured after: `70px` above
      and `70px` below, offset from the table's centre exactly `0`, and the button does
      not overlap the chart.

## 5. The name field

- [x] 5.1 Cap the name input in `web/src/components/NamePrompt.svelte` at the limit.
      Done via a shared `MAX_NAME_LENGTH` in `web/src/lib/name.ts` mirroring the Go
      constant, used by this field and the rename field, so the two cannot drift apart.
- [x] 5.2 Confirm the submit path is otherwise untouched: `submit()` still rejects only
      an empty name and still sends anything else for the server to judge. Do **not**
      add a length check that returns early — that would replace a refusal explaining
      the problem with a button that silently does nothing (`design.md` — decision 6).
      Verify by pasting an over-long name past `maxlength` and confirming the server's
      refusal is displayed.
- [x] 5.3 Check the "already here" list in `NamePrompt.svelte` displays a full-length
      name in full, and fix it if the pill clips. Verified with six seated participants
      at fifteen characters: all shown complete, no clipping, nothing to fix.

## 6. Verification across the change

- [x] 6.1 Run `go build ./...`, `go vet ./...`, `gofmt -l .` (empty output means clean)
      and `go test -race ./...`. The race detector is mandatory here because
      `internal/game` was touched, per `CLAUDE.md`.
- [x] 6.2 Run `npm run check` and `npm run build` from `web/`, then search the build
      output for `http://` and `https://` and confirm no reference to a foreign host
      appeared — the self-contained rule is verified rather than assumed on every
      frontend change.
- [x] 6.3 Walk the spec's acceptance scenarios in a browser: a round in which every
      size was played *and* `?` and `☕` were played, on a wide screen and a narrow one,
      confirming the whole chart sits inside the felt with nothing clipped, scrolled or
      overlapping the seats, that no bar is marked as a winner, and that the table did
      not resize when Reveal was pressed.
- [x] 6.4 Confirm the hidden-round guarantee is intact: with a round unrevealed, check
      the browser's network panel and confirm no snapshot carries a `results` field or
      any card value, including your own. This change touches the results renderer, so
      the guarantee is re-checked rather than assumed.
- [x] 6.5 Run `openspec validate readable-results-and-names --strict` and confirm the
      delta specs parse and the change is complete.
