## MODIFIED Requirements

### Requirement: Cards come from a named deck

A room SHALL hold one active deck, identified by a name, listing the cards that may be played. Two
decks are defined:

- The **t-shirt** deck contains `XS`, `S`, `M`, `L`, `XL`, `?`, `☕`, in that order. Its ordered
  sizing scale is `XS`, `S`, `M`, `L`, `XL`.
- The **Fibonacci** deck contains `0`, `½`, `1`, `2`, `3`, `5`, `8`, `13`, `21`, `?`, `☕`, in that
  order. Its ordered sizing scale is `0`, `½`, `1`, `2`, `3`, `5`, `8`, `13`, `21`.

The order is part of each deck because it determines both the order of selectable cards and the
order of results. In both decks `?` means "I cannot estimate this" and `☕` means "I need a break";
neither belongs to the sizing scale.

No custom deck and no third deck SHALL be offered. A participant's vote SHALL be accepted only when
its card belongs to the room's active deck.

#### Scenario: The t-shirt deck has its fixed cards

- **WHEN** a room uses the t-shirt deck
- **THEN** its cards are exactly `XS`, `S`, `M`, `L`, `XL`, `?`, `☕` in that order, and its sizing
  scale excludes `?` and `☕`

#### Scenario: A new room uses the t-shirt deck

- **WHEN** a room is created without an explicit supported deck selection
- **THEN** its deck is the one named for t-shirt sizes, listing exactly `XS`, `S`, `M`, `L`, `XL`,
  `?`, `☕` in that order

#### Scenario: The Fibonacci deck has its fixed cards

- **WHEN** a room uses the Fibonacci deck
- **THEN** its cards are exactly `0`, `½`, `1`, `2`, `3`, `5`, `8`, `13`, `21`, `?`, `☕` in that
  order, and its sizing scale excludes `?` and `☕`

#### Scenario: A card outside the deck is refused

- **WHEN** a participant tries to play a card that the room's active deck does not contain
- **THEN** the vote is refused with an error, and no vote is recorded for that participant

## ADDED Requirements

### Requirement: A room's deck changes only outside an active vote

Any seated participant SHALL be able to select either supported deck for the shared room; there is
no host or creator privilege. A visitor who is not seated SHALL NOT be able to change it.

While a round is hidden and contains no votes, a valid selection SHALL become the active deck
immediately. Once any participant has voted in a hidden round, a deck selection SHALL be refused and
MUST leave the active deck and every vote unchanged.

After a round is revealed, a valid selection SHALL be stored for the next round. It MUST NOT change
the active deck, cards, ordering, or results of the revealed round. Another valid selection before
the next round starts SHALL replace the pending selection, so the latest selection is the one that
takes effect. Starting the next round SHALL activate that pending deck while clearing the previous
round's votes as usual.

The shared room state SHALL identify the active deck and, when one exists, the pending next-round
deck so that all seated participants see the same setting and timing.

#### Scenario: An empty hidden round changes immediately

- **WHEN** a seated participant selects Fibonacci while the current round is hidden and nobody has
  voted
- **THEN** Fibonacci becomes the active deck immediately for everyone in the room

#### Scenario: A hidden vote locks the deck

- **WHEN** any participant has voted in a hidden round and a seated participant tries to select a
  different deck
- **THEN** the selection is refused, the active deck does not change, and every recorded vote is
  preserved

#### Scenario: A revealed round keeps its deck and results

- **WHEN** a t-shirt round is revealed and a seated participant selects Fibonacci
- **THEN** the revealed t-shirt cards and results remain unchanged, and Fibonacci is shown as the
  pending deck for the next round

#### Scenario: The latest pending selection wins

- **WHEN** participants select Fibonacci and then t-shirt after a round is revealed but before the
  next round starts
- **THEN** t-shirt is the pending deck and Fibonacci is no longer pending

#### Scenario: Starting a round activates the pending deck

- **WHEN** Fibonacci is pending and any seated participant starts a new round
- **THEN** the new round is hidden and empty, Fibonacci is active, and no deck remains pending

#### Scenario: A visitor cannot change the deck

- **WHEN** somebody who has not taken a seat tries to change a room's deck
- **THEN** the change is refused and neither the active nor pending deck changes
