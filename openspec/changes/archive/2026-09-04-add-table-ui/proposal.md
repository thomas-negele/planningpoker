## Why

Everything the product does works, and nobody can use it. A room can be created,
joined, voted in, revealed and revoted — but only by a script holding a WebSocket
open. The placeholder page says so in as many words.

This change builds the part a person actually touches: the entry screen, the name
prompt, the table with everyone sitting around it, the deck along the bottom, and
the two buttons that end and restart a round. It is the last change of the MVP as
specified, and after it the product is the thing that was described in the first
place.

It also has to make good on two promises that have been built for and never once
exercised by a real user. The seat cookie and the room grace period exist so that
closing a laptop and opening it again puts you back in your chair with your vote
intact — but nothing has ever reconnected a dropped socket, so that path has never
run outside a test. And the rules have allowed renaming since they were written,
with no control anywhere that can invoke it.

## What Changes

- **An entry screen** at the base URL whose only option is to start a new game, as
  specified. Starting one navigates to the new room.
- **A name screen** before taking a seat. The name is remembered in a cookie and the
  field arrives pre-filled, so a returning visitor confirms rather than retypes —
  but still chooses the moment they join, because opening a link somebody sent is
  not the same as being ready to sit down.
- **The table**: participants arranged around it, each showing their name, whether
  they have voted, and whether they are away. Cards stay face down until the round
  is revealed.
- **The deck** fixed along the bottom edge, offering the cards the server says the
  room holds — `XS`, `S`, `M`, `L`, `XL`, `?` and `☕` — with the played card
  visibly raised. Playing a different card replaces it, as the rules allow.
- **Reveal and revote**, available to anyone at the table at any time, because that
  is what the rules say and the interface must not pretend otherwise.
- **Results after the reveal**: every participant's card, plus the count per card.
  No average, because t-shirt sizes have none.
- **An "Invite players" control** in a fixed place that never moves, copying the room
  URL to the clipboard. While you are the only person there, a quiet line beneath it
  says so; that line goes when somebody arrives, and the control does not.
- **Renaming** by clicking your own name, which is what makes the rule reachable.
- **A visible connection state**, and automatic reconnection with a growing delay
  when the socket drops. A game that has ended says so and offers a new one, rather
  than retrying something that will never succeed.
- **The end-of-game screen** for an invitation that no longer works, which the
  server has been able to report since the protocol was built and nothing has ever
  displayed.
- The placeholder page is **replaced**, not extended.

## Capabilities

### New Capabilities

- `table-ui`: what a person sees and does — starting a game, giving a name, the
  table and the people around it, playing and changing a card, revealing, reading
  the results, starting a fresh round, renaming, and inviting others.
- `connection-resilience`: how the page stays connected and what it tells you when
  it is not — reconnecting after a drop, distinguishing a temporary fault from a
  game that has ended, and never leaving a stale table on screen pretending to be
  live.

### Modified Capabilities

None. `app-delivery` already requires that a client-side route such as
`/g/<room-id>` survives a hard reload, which is the route this change introduces;
`game-sessions` and `live-updates` already describe everything the server does here.
This change is the first one that only consumes existing behaviour rather than
adding any.

## Impact

- **Modified code**: `web/` almost entirely — the placeholder page is replaced by
  the real interface, with components, a small client-side router, a connection
  layer and a store holding the last snapshot.
- **No Go changes are expected.** The protocol, the rules and the delivery machinery
  are all in place. If this change finds itself needing a server change, that is a
  signal to stop and say so rather than to reach across the boundary.
- **No new dependencies**, and nothing loaded from another host: no icon library, no
  web font, no CSS framework. Icons are inline SVG, the coffee cup is an emoji from
  the operating system's own font, and the styling is written here.
- **A constraint worth knowing precisely**: the Content-Security-Policy forbids
  inline styles. Measured against this project's own policy, that blocks a
  `style="..."` attribute written into markup and blocks `setAttribute('style', …)`,
  but does not block the object model — and Svelte compiles its dynamic styles
  through the object model, so components are free to compute a seat position. What
  it does block is a style attribute hand-written into `web/index.html`, or one
  arriving through `{@html …}`. See `design.md` for the measurements.
- **After this change the MVP is complete** as originally described. What remains
  unbuilt is everything that was deliberately excluded: other decks, deck selection,
  accounts, history, statistics, timers, spectators and room settings.
