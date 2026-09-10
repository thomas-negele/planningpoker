## Context

See [proposal.md](proposal.md) for motivation. A room currently owns one t-shirt `Deck`; room
creation always constructs that deck, the room goroutine serializes every participant action, and a
snapshot sends the active deck to every client. The browser already renders cards and result scales
from that snapshot rather than from a hard-coded card list.

The change crosses the domain, room manager, HTTP creation endpoint, WebSocket protocol, snapshots,
and two Svelte screens. The existing no-host model, hidden-vote boundary, in-memory room lifetime,
and server-authoritative validation remain constraints.

## Goals / Non-Goals

**Goals:**

- Represent the two supported decks once in the domain and validate every deck name there.
- Make both immediate and next-round changes one serialized room action, so votes and settings
  cannot observe a partially applied state.
- Keep old creation requests and implicitly created named rooms on the t-shirt default.
- Expose enough shared state for every client to render the active and pending choices truthfully.
- Fit the longer Fibonacci deck and result scale at supported screen widths.

**Non-Goals:**

- A generic settings framework, custom deck editor, or operator-configured deck catalogue.
- Ownership, moderation, or any new distinction between seated participants.
- Remembering a deck on the browser or persisting room state across process restarts.
- Adding averages, consensus detection, round history, or any interpretation of results.

These boundaries are part of the implementation agreement: work that is not necessary for the two
fixed decks and their specified selection flow stays out of this change.

## Decisions

### 1. Keep a closed domain catalogue keyed by stable deck names

Add the Fibonacci card constants and a fresh-slice `FibonacciDeck` alongside `TShirtDeck`. Resolve
external names through a domain-owned lookup that accepts only `t-shirt` and `fibonacci`; callers do
not construct arbitrary `Deck` values. Constructors take a validated deck choice, while compatibility
entry points with no choice retain t-shirt.

This preserves one source of truth for cards, ordering, and sizing scales. A registry intended for
plugins or custom configuration was rejected because it would create extension behavior outside the
agreed two-deck scope.

### 2. Store active and optional pending decks on the room

The room owns its active deck and an optional pending deck. A `SetDeck` domain action receives the
acting participant and selected name:

- in a hidden round with no votes, it replaces the active deck immediately and clears any pending
  value;
- in a hidden round with one or more votes, it returns a dedicated refusal without mutation;
- in a revealed round, it replaces only the pending deck;
- `NewRound` activates a pending deck before exposing the new empty hidden round, then clears the
  pending value.

The same room command loop used by voting serializes this action. Therefore, if a first vote and a
deck change arrive close together, whichever command is accepted first establishes the state against
which the second is validated; no extra locking or client-side arbitration is needed.

Treating a revealed-round selection as an immediate deck replacement was rejected because the
displayed cards and result ordering would no longer describe the round that produced them. Starting
a new round automatically on selection was rejected because the agreed behavior retains the results
until somebody explicitly chooses `New round`.

### 3. Extend existing protocols narrowly and compatibly

The game-creation request accepts a JSON deck name. An omitted or empty choice means t-shirt for
compatibility; an unknown non-empty name receives a client error before room allocation. The manager
gains a creation path that accepts the validated choice, while `EnsureRoom` continues to create named
rooms with t-shirt.

Add one WebSocket intent, `setDeck`, carrying the selected deck name. It passes through the room's
ordinary seated-participant check and maps domain refusals to stable wire codes. Snapshots continue to
carry the full active deck and gain an optional full pending deck. Sending a deck value instead of a
bare flag lets all clients display exactly which choice is pending without reconstructing domain
state.

No separate settings endpoint or client-only optimistic authority is added: both alternatives would
duplicate the existing shared-room action and broadcast path.

### 4. Use small, explicit selection surfaces rather than a settings system

The entry screen gets a compact two-option field immediately above its existing start button. It
submits the selected stable name with creation; t-shirt is initially selected and nothing is stored
locally.

At the table, place an icon-only settings button immediately left of `Invite players` in the right
header group. It opens a focused dialog containing only the two deck choices. The dialog reads active
and pending state from the latest snapshot. After reveal, its wording marks the selection as applying
to the next round.

During a hidden round with votes, render the control disabled. Because disabled controls do not
reliably expose hover or focus help, wrap it with the tooltip target and include the reason in the
control's accessible name for assistive technology. Server validation remains authoritative if a
stale client sends a change anyway.

A permanently visible settings panel was rejected because it competes with the estimation controls;
hiding the icon while locked was rejected because it makes the setting appear to move or vanish.

### 5. Adapt existing data-driven deck and result layouts only where required

Keep `Deck` and `Results` data-driven. Verify and minimally adjust their sizing/wrapping so eleven
Fibonacci choices remain reachable and nine sizing rows plus the conditional non-size row remain
legible on wide and narrow screens. Do not redesign cards, results, the table, or navigation beyond
the changes needed to accommodate the larger fixed deck and the adjacent settings icon.

## Risks / Trade-offs

- [A late client renders settings as enabled after another participant has voted] → The server
  rejects the stale command without mutation and the next snapshot/refusal brings the UI current.
- [Eleven cards or nine result rows overflow the present layout] → Add component and browser-width
  coverage for both decks, then make only targeted responsive adjustments.
- [A pending deck is mistaken for the revealed round's deck] → Preserve the full active deck until
  `New round` and label pending state explicitly in the settings dialog.
- [New protocol fields break older clients] → Keep the pending field optional and preserve t-shirt
  when the creation deck is omitted; unknown new commands are already refused safely.

## Migration Plan

No stored data needs migration because rooms exist only in memory. Deploy server and bundled web
client together. Existing rooms at deployment do not survive a restart and newly created rooms
default to t-shirt unless Fibonacci is selected. Rollback is the previous binary and removes the new
choice without converting data.
