## MODIFIED Requirements

### Requirement: A participant may cast and freely change a vote while the round is hidden

While a round is hidden, a seated participant who is not a visitor SHALL be able to record a card, and SHALL be able to
replace it with a different card any number of times. The last card chosen is the participant's
vote.

Replacing a vote leaves no trace of what was chosen before: earlier choices within a round are not
recorded and are not recoverable, because nothing in this product has any use for them and keeping
them would leak a participant's hesitation.

Each participant holds at most one vote per round. A visitor holds no vote in a hidden round and any attempt to vote SHALL be refused.

#### Scenario: Casting a vote records it

- **WHEN** a non-visitor participant plays `M` while the round is hidden
- **THEN** their vote for the round is `M`

#### Scenario: Changing a vote replaces it

- **WHEN** a non-visitor participant who has played `M` then plays `L`, and then plays `S`
- **THEN** their vote for the round is `S`, they still hold exactly one vote, and neither `M` nor
  `L` is recorded anywhere

#### Scenario: A participant who has not voted holds no card

- **WHEN** a round is running and a participant has played nothing
- **THEN** that participant has no vote, which is distinct from having voted for `?`

#### Scenario: Visitor cannot vote

- **WHEN** a visitor attempts to play a card while the round is hidden
- **THEN** the attempt is refused and they hold no vote

### Requirement: The round reports whether everyone present has voted

A round SHALL report whether every participant who is neither marked away nor a visitor has voted. Participants
marked away and visitors are not counted, so that someone who has closed their laptop cannot leave the round
permanently incomplete.

When no present participant is eligible to vote, the report SHALL be false rather than suggesting that all votes are in.

This is a statement about the round, not a permission: revealing is allowed regardless of it, as
specified above.

#### Scenario: Everyone present has voted

- **WHEN** every participant who is neither away nor a visitor has voted, while an away participant has not
- **THEN** the round reports that everyone present has voted

#### Scenario: Someone present has not voted

- **WHEN** at least one participant who is neither away nor a visitor has not voted
- **THEN** the round reports that not everyone present has voted

#### Scenario: A returning participant without a vote makes the round incomplete again

- **WHEN** an away participant who holds no card is marked present, in a round that previously
  reported everyone present as having voted
- **THEN** the round now reports that not everyone present has voted

#### Scenario: Visitors do not hold up completion

- **WHEN** all present non-visitors have voted and a present visitor has not
- **THEN** the round reports that everyone eligible and present has voted

#### Scenario: Nobody present can vote

- **WHEN** every present participant is a visitor
- **THEN** the round does not report that everyone present has voted
