## Why

Teams estimate with different scales, but every room currently uses t-shirt sizes. A room needs a
small, shared choice between the existing t-shirt scale and a Fibonacci scale without turning the
product into a general settings system.

## What Changes

- Add a fixed Fibonacci deck containing `0`, `½`, `1`, `2`, `3`, `5`, `8`, `13`, `21`, `?`, `☕`
  in that order; t-shirt sizes remain the default.
- Let somebody choose either fixed deck while creating a room from the entry screen. Rooms created
  implicitly by opening a valid, unused room URL continue to start with the t-shirt deck.
- Let any seated participant change the room's deck before anybody votes in a hidden round.
- After a round is revealed, let any seated participant choose the deck for the next round. The
  current results keep their original deck, the latest pending choice wins, and the choice takes
  effect when somebody starts the new round.
- Add an unobtrusive settings icon beside the invitation control. Disable it while a hidden round
  contains at least one vote and explain the disabled state with accessible help text.
- Keep the scope to selecting between these two fixed decks. This change does not add custom decks,
  further room settings, host or creator privileges, round history, new statistics, or persistent
  user preferences.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `estimation-rounds`: Define the Fibonacci deck and the rules for changing a room's active or next
  deck without altering votes or revealed results.
- `game-sessions`: Allow an issued room to be created with a selected fixed deck while retaining
  t-shirt sizes as the default for omitted selections and implicitly created named rooms.
- `table-ui`: Add the compact creation choice and the in-room settings control, including its
  permission, disabled, pending, and responsive behavior.

## Impact

The room domain model and snapshots must represent the active deck and any pending next-round deck.
The room-creation HTTP contract gains a validated deck selection, and the WebSocket command set gains
a validated settings change broadcast to the room. The Svelte entry and room screens gain the two
small selection surfaces. Domain, transport, HTTP, frontend protocol, component, and end-to-end tests
cover both decks, timing rules, authorization, concurrent ordering, and unchanged hidden-vote
privacy. No new dependency, persistence layer, or role model is introduced.
