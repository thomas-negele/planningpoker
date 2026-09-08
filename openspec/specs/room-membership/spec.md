# room-membership Specification

## Purpose

Defines what a planning poker room is and who is sitting at it: how a room is created and
identified, how someone takes a seat under a name, how a returning browser is recognised as the
same person rather than a duplicate, what happens to a seat when a connection drops, and when a
room counts as empty.

## Requirements

### Requirement: A participant takes a seat under a name

A participant SHALL join a room by supplying a display name. The name is what other participants
see at the table.

A name that is empty, or consists only of whitespace, SHALL be rejected. Leading and trailing
whitespace SHALL be trimmed before the name is stored, so that two names differing only in stray
spaces do not appear different at the table.

The maximum length SHALL be **15 characters**, counted in Unicode code points rather than bytes so
that a name in any script gets the same allowance. Trimming happens before the length is measured,
so surrounding spaces never count against the limit.

The limit and the table have to agree, and that is the whole of the reasoning. A name is displayed
at a seat in a ring around the table, and `table-ui` requires that a name the rules accept is
displayed in full — never clipped, never abbreviated. So the limit cannot be chosen on its own:
every extra character widens a seat, and every widened seat pushes the ring outwards until it no
longer fits the screen it is drawn on. The alternative to sizing the table for the limit is
truncating names at the seat, which is what used to happen and is what this limit exists to prevent.

Fifteen characters takes a name like `Maria-Katharina`. The price is paid at the other end: the ring
needs roughly 58rem of width to hold the table plus two seats of that size, so a window narrower
than that is given the list layout instead of the ring. That is the trade this number makes, and it
is a deliberate one.

The limit is a fixed rule of the product rather than something an operator tunes: it is chosen
together with the geometry of the table and could not be raised without changing that geometry too,
so exposing it as an adjustable value would offer a freedom the interface cannot honour.

The limit SHALL be enforced by the server. A page may prevent a longer name from being composed as a
courtesy, but that never substitutes for the server's check, since the server is spoken to by more
than one page and by things that are not this page at all.

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

#### Scenario: A name of exactly fifteen characters is accepted

- **WHEN** someone joins with a name of exactly 15 characters, such as `Maria-Katharina`
- **THEN** the join succeeds and the participant is seated under that name unchanged

#### Scenario: An excessively long name is refused

- **WHEN** someone tries to join with a name of 16 characters or more, such as `Johann Sebastian`
- **THEN** the join is refused with an error naming the limit, and no seat is taken, and the name is
  not shortened to fit

#### Scenario: The limit counts characters, not bytes

- **WHEN** someone joins with a name of 15 characters whose characters are multi-byte, such as a
  name written in a non-Latin script
- **THEN** the join succeeds, because the allowance is the same in every script

#### Scenario: Trimming happens before the length is measured

- **WHEN** someone joins with a name of 15 characters surrounded by spaces
- **THEN** the surrounding spaces are removed and the name is accepted, rather than counted towards
  the limit and refused

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

A connection that is lost **without being closed** counts as lost in the same way. A silently dead
connection — one whose network path has gone while the socket still appears open — SHALL be
recognised within a bounded time and its participant marked away like any other. Before, such a
participant stayed shown as present indefinitely, which is worse than showing them as away: the
table looked as though somebody was there to vote.

A participant is removed from a room only when the room itself ceases to exist. Nothing in these
rules removes a participant on a timer, and the recognition above is emphatically not such a rule —
it measures whether the network path is alive, never whether anybody has done anything.

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

#### Scenario: A connection that dies without closing is recognised

- **WHEN** a participant's connection stops answering without the socket being closed
- **THEN** within a bounded time they are marked away and the table stops showing them as present

#### Scenario: Thinking for a long time is not going away

- **WHEN** a participant stays connected but sends nothing at all for far longer than any interval
  in the system
- **THEN** they remain present and seated, because nothing measures how long it has been since
  somebody last did something

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

### Requirement: An issued room identifier is unguessable

Because there is no authentication anywhere in this product, an invitation link is the only thing
that can protect a room. An identifier **issued** by this package — the kind produced when somebody
starts a new game — SHALL therefore be generated from a cryptographically secure source of
randomness, carry at least 128 bits of entropy, and be rendered in a URL-safe alphabet that excludes
characters easily confused with one another when read aloud or copied by hand.

Issued identifiers MUST NOT be sequential, derived from a counter, or derived from the time of
creation, since any of those would let one room be found from another.

The source of randomness SHALL be supplied by the caller rather than read from a package-level
global, so that the rules can be tested with a fixed source and produce repeatable identifiers.

This requirement is now about one of two kinds of identifier rather than about all of them. A room
may also be reached by a name somebody typed, which is guessable by design and is described in
`game-sessions`. Nothing here weakens: an identifier this package issues is exactly as unguessable
as it ever was. What changed is that being unguessable is a property of how an identifier was made,
not of every identifier that may name a room — and saying so plainly is what stops somebody later
reading this requirement as a promise the product does not keep.

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

#### Scenario: A typed name is not held to this rule

- **WHEN** a room is reached by a name somebody typed rather than by an issued identifier
- **THEN** it is a room like any other, and no claim is made that its address cannot be guessed

### Requirement: A room seats a bounded number of participants

A room SHALL seat no more than a configured maximum number of participants. The limit counts every
participant holding a seat, including those currently marked away, because an away participant keeps
their seat and their vote and is still shown at the table.

An attempt to take a seat in a full room SHALL be refused with its own distinguishable reason, so
the page can say the table is full rather than showing a generic failure. The refusal SHALL leave
the room exactly as it was: nobody is unseated, no vote is lost, and no snapshot is sent on account
of it. The connection SHALL remain open, so that somebody who is refused can watch the table and
take a seat later if one is freed by a new round of the room's lifetime.

A browser that is being reseated into a seat it already holds SHALL NOT be refused by this limit,
because it occupies no additional seat. Reconnecting must not fail merely because the room filled up
while the connection was away.

#### Scenario: A seat beyond the limit is refused

- **WHEN** a room already seats its maximum number of participants and another connection asks to
  take a seat
- **THEN** the request is refused with a reason naming the full table, no participant is added, and
  everybody already seated is unaffected

#### Scenario: An away participant still occupies their seat

- **WHEN** a room is at its participant limit and one of those participants is marked away after a
  dropped connection
- **THEN** the seat is still counted, and a different browser asking for a seat is still refused

#### Scenario: Reconnecting into an existing seat is never refused for capacity

- **WHEN** a browser holding a seat token for a full room reconnects
- **THEN** it is reseated into the seat it already held, without being refused for capacity

#### Scenario: A refused seat leaves the room untouched

- **WHEN** a seat request is refused because the room is full
- **THEN** the room's participants, votes and round state are exactly as they were, and no other
  connection receives anything as a result
