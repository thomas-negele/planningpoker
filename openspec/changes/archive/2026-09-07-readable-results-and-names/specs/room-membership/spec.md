## MODIFIED Requirements

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
