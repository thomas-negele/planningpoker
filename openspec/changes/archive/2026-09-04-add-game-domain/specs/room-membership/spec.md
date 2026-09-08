## Purpose

Defines what a planning poker room is and who is sitting at it: how a room is created and
identified, how someone takes a seat under a name, how a returning browser is recognised as the
same person rather than a duplicate, what happens to a seat when a connection drops, and when a
room counts as empty.

## ADDED Requirements

### Requirement: A room is identified by an unguessable identifier

Because there is no authentication anywhere in this product, the invitation link is the only thing
protecting a room. A room identifier SHALL therefore be generated from a cryptographically secure
source of randomness, carry at least 128 bits of entropy, and be rendered in a URL-safe alphabet
that excludes characters easily confused with one another when read aloud or copied by hand.

Identifiers MUST NOT be sequential, derived from a counter, or derived from the time of creation,
since any of those would let one room be found from another.

The source of randomness SHALL be supplied by the caller rather than read from a package-level
global, so that the rules can be tested with a fixed source and produce repeatable identifiers.

#### Scenario: Two rooms never share an identifier

- **WHEN** many rooms are created in succession from a cryptographically secure source
- **THEN** every identifier differs from every other, and none can be derived from another by
  incrementing, decrementing, or any other simple transformation

#### Scenario: Identifier is safe to put in a URL

- **WHEN** a room identifier is generated
- **THEN** it consists only of characters that need no escaping in a URL path, and contains none of
  the character pairs that are routinely misread for each other

#### Scenario: A fixed source produces a repeatable identifier

- **WHEN** a room is created twice from two sources of randomness yielding identical bytes
- **THEN** both rooms receive the same identifier, so that tests can assert on it

#### Scenario: A failing source of randomness is reported, never worked around

- **WHEN** the source of randomness returns an error or too few bytes
- **THEN** room creation fails with that error, and no room is created with a weak, padded or
  partially random identifier

### Requirement: A participant takes a seat under a name

A participant SHALL join a room by supplying a display name. The name is what other participants
see at the table.

A name that is empty, or consists only of whitespace, SHALL be rejected. Leading and trailing
whitespace SHALL be trimmed before the name is stored, so that two names differing only in stray
spaces do not appear different at the table. A maximum length SHALL be enforced so that one
participant cannot disrupt the table for everyone else.

Two participants in the same room MAY hold the same name. Duplicate names are explicitly allowed
and are not an error: the table is small enough that people resolve a collision between themselves,
and identity is carried by the participant identifier rather than by the name.

#### Scenario: Joining with a name seats the participant

- **WHEN** someone joins a room with the name `Thomas`
- **THEN** the room contains a participant named `Thomas`, and every other participant in that room
  can see them at the table

#### Scenario: Two people may share a name

- **WHEN** a second person joins the same room, also with the name `Thomas`
- **THEN** the join succeeds, the room contains two distinct participants both named `Thomas`, and
  each remains individually addressable

#### Scenario: An empty name is refused

- **WHEN** someone tries to join with an empty name, or with a name consisting only of spaces
- **THEN** the join is refused with an error saying the name is required, and no seat is taken

#### Scenario: Surrounding whitespace is trimmed

- **WHEN** someone joins with the name `"  Thomas  "`
- **THEN** the participant is seated under the name `Thomas`

#### Scenario: An excessively long name is refused

- **WHEN** someone tries to join with a name longer than the permitted maximum
- **THEN** the join is refused with an error naming the limit, and no seat is taken

### Requirement: A participant is identified by an opaque identifier, not by name

Every participant SHALL carry an opaque identifier, generated the same way a room identifier is and
from the same caller-supplied source of randomness. That identifier, and never the display name, is
what identifies a participant in every operation: casting a vote, renaming, going away, returning.

This is what makes duplicate names harmless and renaming safe.

#### Scenario: Operations address a participant by identifier

- **WHEN** a room contains two participants who share the display name `Thomas`, and one of them
  votes
- **THEN** the vote is recorded against exactly that participant, and the other `Thomas` is
  unaffected

#### Scenario: An unknown identifier is refused

- **WHEN** an operation names a participant identifier that is not seated in that room
- **THEN** the operation is refused with an error, and no state changes

### Requirement: A returning browser is reseated, not duplicated

A participant who rejoins a room presenting the identifier of a participant already seated there
SHALL be reseated as that same participant, keeping their seat, their name and their vote in the
current round. They MUST NOT appear as a second participant.

This is what makes an ordinary page reload invisible to everyone else.

#### Scenario: A reload keeps seat, name and vote

- **WHEN** a participant who has already voted disconnects and rejoins presenting the same
  participant identifier
- **THEN** they occupy the same seat, under the same name, with their vote in the current round
  still recorded, and the number of participants at the table is unchanged

#### Scenario: Rejoining with a new name updates the name

- **WHEN** a participant rejoins with the same identifier but a different display name
- **THEN** they are reseated as the same participant and their displayed name becomes the new one

### Requirement: A dropped connection marks a participant away rather than removing them

When a participant's connection is lost, they SHALL be marked as away. They keep their seat, their
name and their vote, and they remain visible to everyone else at the table, distinguishably marked
as away. When they return, the away mark is cleared.

A participant is removed from a room only when the room itself ceases to exist. Nothing in these
rules removes a participant on a timer.

#### Scenario: Going away keeps the seat and the vote

- **WHEN** a participant who has voted is marked away
- **THEN** they remain at the table, their vote remains recorded, and everyone else sees them
  marked as away

#### Scenario: Returning clears the away mark

- **WHEN** a participant who is marked away is marked present again
- **THEN** the away mark is cleared and nothing else about them changes

#### Scenario: Nobody is removed by the passage of time

- **WHEN** a participant has been away for any length of time
- **THEN** they are still seated at the table, because no rule removes a participant on a timer

### Requirement: A participant may change their own name

A seated participant SHALL be able to change their display name at any time, whether a round is
hidden or revealed. The new name is subject to the same rules as a name given when joining: it is
trimmed, it may not be empty, it has a maximum length, and it may duplicate another participant's
name.

Renaming SHALL NOT affect the participant's identifier, their seat, or their vote.

#### Scenario: Renaming changes only the name

- **WHEN** a participant who has voted renames themselves from `Thomas` to `Thomas N.`
- **THEN** their displayed name changes, and their identifier, their seat and their recorded vote
  are unchanged

#### Scenario: Renaming to an invalid name is refused

- **WHEN** a participant tries to rename themselves to an empty name, or to one exceeding the
  maximum length
- **THEN** the rename is refused with an error and the previous name is kept

#### Scenario: Renaming during a revealed round is allowed

- **WHEN** a participant renames themselves while the round is revealed
- **THEN** the rename succeeds and the revealed results continue to show that participant's card,
  now under the new name

### Requirement: A room reports whether anyone is still present

A room SHALL report whether it currently has any participant who is not marked away. This is the
question the layer above needs in order to decide a room's fate; these rules answer the question
but take no action on it, because discarding a room is a matter of elapsed time rather than a rule.

#### Scenario: A room with a present participant is not deserted

- **WHEN** at least one participant in the room is not marked away
- **THEN** the room reports that someone is present

#### Scenario: A room whose participants are all away is deserted

- **WHEN** every participant in the room is marked away, or the room has no participants at all
- **THEN** the room reports that nobody is present
