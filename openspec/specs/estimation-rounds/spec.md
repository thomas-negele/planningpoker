# estimation-rounds Specification

## Purpose

Defines how a single estimate is taken at the table: which cards may be played, how a vote is cast
and changed, what stays hidden while the round runs, who may reveal it and when, what the results
show once revealed, and how a fresh round is started.

## Requirements

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

### Requirement: A participant may cast and freely change a vote while the round is hidden

While a round is hidden, a seated participant SHALL be able to record a card, and SHALL be able to
replace it with a different card any number of times. The last card chosen is the participant's
vote.

Replacing a vote leaves no trace of what was chosen before: earlier choices within a round are not
recorded and are not recoverable, because nothing in this product has any use for them and keeping
them would leak a participant's hesitation.

Each participant holds at most one vote per round.

#### Scenario: Casting a vote records it

- **WHEN** a participant plays `M` while the round is hidden
- **THEN** their vote for the round is `M`

#### Scenario: Changing a vote replaces it

- **WHEN** a participant who has played `M` then plays `L`, and then plays `S`
- **THEN** their vote for the round is `S`, they still hold exactly one vote, and neither `M` nor
  `L` is recorded anywhere

#### Scenario: A participant who has not voted holds no card

- **WHEN** a round is running and a participant has played nothing
- **THEN** that participant has no vote, which is distinct from having voted for `?`

### Requirement: A hidden round never discloses what anyone voted

While a round is hidden, the view of the room that leaves these rules SHALL report, for each
participant, only *whether* they have voted. It MUST NOT contain the value of any vote, in any
form, for any participant — not the voter's own, and not in any field, count, summary or ordering
from which a value could be recovered.

This is the single most important rule in the product. Transmitting a value and hiding it in the
interface would put every vote one developer-tools panel away from anyone at the table, which
defeats the entire purpose of estimating simultaneously.

Once the round is revealed, the same view SHALL report every vote in full.

#### Scenario: A hidden round exposes only that someone voted

- **WHEN** several participants have voted and the round has not been revealed
- **THEN** the view shows for each participant whether they have voted, and contains no card value
  anywhere

#### Scenario: The tally is withheld while hidden

- **WHEN** a round is hidden
- **THEN** the view contains no count per card either, since a tally of a small table would let the
  individual votes be worked out

#### Scenario: Revealing discloses every vote

- **WHEN** the round is revealed
- **THEN** the view shows each participant's card, and participants who did not vote are shown as
  holding no card

### Requirement: Any participant may reveal the round at any time

Any seated participant SHALL be able to reveal the round. There is no host, no creator role and no
permission to check: the table is small and the people at it are colleagues.

Revealing SHALL be permitted whether or not everyone has voted. Participants who have not voted are
shown as holding no card. This is deliberate: it means one person who is absent, distracted or on a
dead network can never block a meeting.

Revealing an already revealed round SHALL leave it unchanged rather than fail, so that two people
pressing the button at the same moment is not an error.

#### Scenario: Any participant can reveal

- **WHEN** a participant who did not create the room reveals the round
- **THEN** the round becomes revealed

#### Scenario: Revealing before everyone has voted is allowed

- **WHEN** the round is revealed while some participants have not voted
- **THEN** the round becomes revealed, showing the cards of those who voted and no card for those
  who did not

#### Scenario: Revealing twice is harmless

- **WHEN** a round that is already revealed is revealed again
- **THEN** the round remains revealed with exactly the same votes, and no error is reported

### Requirement: A revealed round is final until a new one is started

Once a round is revealed, no vote may be cast or changed within it. An attempt to do so SHALL be
refused with an error and SHALL NOT alter any recorded vote.

Without this rule the reveal would decide nothing: anyone could adjust their estimate after seeing
everyone else's, which is exactly the anchoring that voting simultaneously exists to prevent, and
the others would have no way of knowing it had happened.

#### Scenario: Voting after the reveal is refused

- **WHEN** a participant who did not vote tries to play a card after the round has been revealed
- **THEN** the attempt is refused with an error and they still hold no card

#### Scenario: Changing a vote after the reveal is refused

- **WHEN** a participant who voted `M` tries to change it to `L` after the reveal
- **THEN** the attempt is refused with an error and their vote remains `M`

### Requirement: Results show every card and a count per card

A revealed round SHALL report each participant's card, and additionally a count of how many
participants played each card. The counts SHALL be ordered by the deck's own order, and SHALL
include only cards that were actually played.

No average, median or other arithmetic SHALL be computed. T-shirt sizes are an ordered scale
without arithmetic; `?` and `☕` are not sizes at all. Any mean would be an invented number wearing
the appearance of a measurement.

Participants who did not vote are reported as holding no card and are counted in no tally.

#### Scenario: The tally counts each played card

- **WHEN** a round is revealed in which two participants played `M`, one played `L` and one played
  `☕`
- **THEN** the results report each of those four participants' cards, and a tally of `M` twice, `L`
  once and `☕` once, in the deck's order

#### Scenario: Cards nobody played are absent from the tally

- **WHEN** a round is revealed in which nobody played `XS`
- **THEN** `XS` does not appear in the tally at all, rather than appearing with a count of zero

#### Scenario: Non-voters are excluded from the tally

- **WHEN** a round is revealed in which one participant did not vote
- **THEN** that participant is reported as holding no card, and the counts add up to the number of
  participants who did vote

#### Scenario: No average is reported

- **WHEN** any round is revealed
- **THEN** the results contain no average, median, or other number derived by arithmetic over the
  cards

### Requirement: Any participant may start a fresh round

Any seated participant SHALL be able to start a new round. The new round is hidden and holds no
votes: every participant starts again with no card, including participants who are marked away.

Starting a new round SHALL be permitted whether the current round is revealed or still hidden.
Starting one from a hidden round discards the votes cast so far, which is the only way to abandon a
round that was begun by mistake.

The previous round's votes are not retained. There is no history, and nothing in this product reads
a round after the one that replaced it.

#### Scenario: A new round clears every vote

- **WHEN** a new round is started after a round in which everyone voted and which was revealed
- **THEN** the round is hidden, no participant holds a card, and the previous votes are gone

#### Scenario: A new round may be started from a hidden round

- **WHEN** a new round is started while the current round is still hidden and partly voted
- **THEN** the votes cast so far are discarded and a fresh hidden round begins

#### Scenario: Away participants also start with no card

- **WHEN** a new round is started while some participants are marked away
- **THEN** those participants remain seated and marked away, and hold no card in the new round

### Requirement: The round reports whether everyone present has voted

A round SHALL report whether every participant who is not marked away has voted. Participants
marked away are not counted, so that someone who has closed their laptop cannot leave the round
permanently incomplete.

This is a statement about the round, not a permission: revealing is allowed regardless of it, as
specified above.

#### Scenario: Everyone present has voted

- **WHEN** every participant who is not away has voted, while an away participant has not
- **THEN** the round reports that everyone present has voted

#### Scenario: Someone present has not voted

- **WHEN** at least one participant who is not away has not voted
- **THEN** the round reports that not everyone present has voted

#### Scenario: A returning participant without a vote makes the round incomplete again

- **WHEN** an away participant who holds no card is marked present, in a round that previously
  reported everyone present as having voted
- **THEN** the round now reports that not everyone present has voted
