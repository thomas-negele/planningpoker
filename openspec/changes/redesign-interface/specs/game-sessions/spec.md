## ADDED Requirements

### Requirement: The decks a game can be started with are readable

The server SHALL answer a read-only request for the decks a game can be started with. For each
supported deck the answer SHALL carry the stable name used when starting a game, the human-readable
label, and the deck's cards in the deck's own order.

The answer SHALL be derived from the same deck definitions the server uses when it creates a room,
so that what a visitor is shown before starting a game and what they are dealt after starting it
cannot disagree. A second list maintained anywhere else — in the page, in a configuration file —
would be a second truth, and this requirement exists precisely to prevent one.

The request SHALL create nothing, change nothing and require nothing: no room, no seat, no cookie.
It is safe to repeat and safe to make before a visitor has decided anything.

#### Scenario: Both supported decks are reported

- **WHEN** the decks are requested
- **THEN** the answer names the t-shirt deck and the Fibonacci deck, each with the identifier used
  to start a game with it, its label, and its cards

#### Scenario: The cards are in the deck's own order

- **WHEN** the decks are requested
- **THEN** each deck's cards appear in the order the deck defines, which is the order they are
  offered in at the table

#### Scenario: What is reported is what is dealt

- **WHEN** a game is started with a deck the answer described
- **THEN** the room's deck offers exactly the cards that answer listed, in that order

#### Scenario: Asking creates nothing

- **WHEN** the decks are requested any number of times
- **THEN** no room is created, no seat is issued, no cookie is set, and the number of rooms the
  process holds is unchanged
