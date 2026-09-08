## Why

The application can now be delivered to a browser, but it has no idea what a planning poker game
is. There is no room, no participant, no deck, no vote and no round. Everything the product
actually does still has to be written.

Writing those rules first, on their own, is worth doing deliberately. The rules are where the
product's meaning lives — who may vote, what a hidden round hides, what a reveal is allowed to
show — and they are the part that must be provably correct. Expressed as plain functions over plain
data, with no network and no concurrency anywhere near them, they can be tested exhaustively and
read by someone who knows nothing about Go's HTTP handling. Written the other way round, as
behaviour that emerges from a WebSocket handler, the same rules become reachable only by opening a
socket and are far easier to get subtly wrong.

The most important thing this change protects is the rule that a hidden vote must never leave the
server. That guarantee is worth exactly as much as the one place it is enforced, and this change
makes that place a pure function with a test rather than a habit of remembering to redact.

## What Changes

- **New `internal/game` package** holding rooms, participants, the deck, rounds and the rules that
  govern them. It imports nothing from `net/http` and nothing from the WebSocket library — a
  constraint that is checked by a test, not merely intended.
- **Rooms** with an identifier long enough not to be guessable, since the invitation link is the
  only thing protecting a room. The identifier is generated from a caller-supplied source of
  randomness, so that production uses a cryptographically secure source while tests can supply a
  fixed one and get repeatable identifiers.
- **Participants** who join a room by name, are identified by an opaque identifier rather than by
  their name, may rename themselves, and may be marked away when their connection drops without
  losing their seat or their vote.
- **The t-shirt deck** — `XS`, `S`, `M`, `L`, `XL`, `?`, `☕` — as a *named* deck, so that a second
  deck can be added later without restructuring anything. **Only the t-shirt deck is implemented in
  this change.** No Fibonacci deck, no deck selection, no room settings.
- **A round** that is hidden while voting and revealed afterwards, with the rules agreed with the
  owner: a participant may change their card freely while the round is hidden; nobody may vote or
  change a vote once it is revealed; any participant may reveal, at any time, whether or not
  everyone has voted; and any participant may start a fresh round.
- **A redacted view of the room** which is the only shape in which room state is allowed to leave
  the domain. While a round is hidden it reports *that* a participant has voted and never *what*.
- **Results after the reveal**: every participant's card, plus a count per card, as decided. No
  average is computed — t-shirt sizes are an ordered scale without arithmetic, and any mean would
  be an invented number.

## Capabilities

### New Capabilities

- `room-membership`: what a room is and who is sitting at it — creating a room with an unguessable
  identifier, joining it by name, being recognised again on reconnect, being marked away rather
  than removed when a connection drops, renaming, and when a room counts as empty.
- `estimation-rounds`: how a single estimate is taken — the deck of cards, casting and changing a
  vote, what stays hidden while the round runs, revealing, the results and their tally, and
  starting a fresh round.

### Modified Capabilities

None. `app-delivery` describes how the application reaches the browser and is untouched by this
change: nothing here is served, routed or transmitted.

## Impact

- **Created code**: `internal/game/` with its tests. Nothing else in the repository changes.
- **No new dependencies.** The package uses only the standard library, and only the parts of it
  that have nothing to do with input or output.
- **Nothing is visible to a user after this change.** No route serves these rules and no message
  carries them; the placeholder page still shows what it shows today. That is the intended shape of
  this slice — the rules become reachable in the next change, which adds the room manager and the
  WebSocket protocol.
- **Deliberately not included**, because each belongs to a later change or was never requested: the
  per-room goroutine and the room manager, the grace period before an empty room is discarded (a
  timer, and therefore not a pure rule), the JSON wire format, cookies, and any second deck.
