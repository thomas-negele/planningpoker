## Purpose

Defines a game as something that exists over time: how one is created and shared, how long a room
outlives the people sitting in it, what a visitor sees when they arrive at a room that has gone, and
how a browser is recognised as the same participant when it comes back.

## ADDED Requirements

### Requirement: Starting a game creates a room and yields an invitation URL

The application SHALL provide a way to create a new game. Creating one produces a room with a fresh
unguessable identifier and returns that identifier, from which the invitation URL is formed.

Creating a game SHALL NOT seat anybody. The person who creates it takes a seat the same way everyone
else does, by opening the URL and giving a name. There is no host and no creator role anywhere in
this product, so there is nothing for the act of creation to confer.

Each creation SHALL produce a distinct room. Creating a game twice never returns the same room.

#### Scenario: Creating a game returns a usable room

- **WHEN** a visitor starts a new game
- **THEN** a room is created, its identifier is returned, and opening the corresponding URL reaches
  that room

#### Scenario: Creating a game seats nobody

- **WHEN** a game has just been created and nobody has opened its URL
- **THEN** the room has no participants, and the person who created it holds no privilege that
  anyone else lacks

#### Scenario: Two games are two rooms

- **WHEN** two games are created in succession
- **THEN** they have different identifiers, and a vote in one is not visible in the other

### Requirement: A room is discarded once nobody has been connected for the grace period

A room SHALL be discarded when no connection has been open to it for longer than a configured grace
period. The same rule covers both cases: a room everyone has left, and a room nobody ever joined.

The grace period exists so that ordinary interruptions do not destroy a session. A page reload, a
laptop lid, a dropped Wi-Fi connection or a few minutes between creating a link and the first person
opening it must all leave the room intact.

The grace period SHALL be readable from an environment variable with a documented default, and its
documentation MUST state what a boundary value means. It is a behaviour-governing value in the sense
of the project's rules, alongside the listen address and the shutdown timeout.

Once a room has been discarded, its identifier is gone. Nothing recreates a room at an identifier
that was previously in use, because doing so would hand a working room to anyone who still had an
old link.

#### Scenario: A reload does not destroy the room

- **WHEN** the only participant in a room reloads their page, so that for a moment no connection is
  open
- **THEN** the room still exists when the new connection arrives, with its participants and their
  votes intact

#### Scenario: A deserted room is discarded after the grace period

- **WHEN** every connection to a room has been closed for longer than the grace period
- **THEN** the room is discarded and its identifier no longer reaches anything

#### Scenario: A room nobody ever joined is discarded too

- **WHEN** a game is created and nobody opens its URL within the grace period
- **THEN** that room is discarded on the same rule, without a separate timer or a separate setting

#### Scenario: An occupied room is never discarded

- **WHEN** at least one connection to a room has been open throughout, however long the game runs
- **THEN** the room is not discarded

#### Scenario: The grace period is configurable and documented

- **WHEN** the grace period environment variable is unset
- **THEN** the documented default applies, and both the code and the deployment configuration
  explain in full sentences what the value does and what happens at its boundary

### Requirement: A visitor arriving at a room that no longer exists is told so

When a visitor opens the URL of a room that does not exist — because it was discarded, because the
server was restarted, or because the identifier was mistyped — the application SHALL tell them the
game is no longer available and offer to start a new one.

It MUST NOT leave them looking at a table that never loads, and it MUST NOT create a room at the
requested identifier. Creating one would destroy the property that makes an identifier private: that
only someone who was given the link can reach the room behind it.

#### Scenario: An expired invitation explains itself

- **WHEN** a visitor opens the URL of a room that has been discarded
- **THEN** they are told the game is no longer available and are offered a way to start a new one

#### Scenario: A mistyped identifier is treated the same way

- **WHEN** a visitor opens a URL whose room identifier never existed
- **THEN** they see the same message, and no room is created at that identifier

#### Scenario: Restarting the server ends open games

- **WHEN** the container is restarted while games are in progress
- **THEN** every room is gone, and everyone who reconnects is told their game is no longer available
  rather than being shown an empty table

### Requirement: A per-room cookie identifies a returning browser

The application SHALL give each browser an opaque **seat token** for each room it connects to,
stored in a cookie scoped to that room. On reconnecting, the browser presents that token and is
reseated as the same participant, keeping its seat, its name and its vote.

The seat token SHALL be a different value from the participant identifier that appears in snapshots,
and the server SHALL NOT send it to any client other than as that browser's own cookie. The two
values serve opposite purposes and must not be conflated: the participant identifier is public, since
every client needs it to render who is at the table, while the seat token is a credential that proves
a browser owns a seat. Were they the same value, every participant could read every other
participant's credential out of an ordinary snapshot and take their seat by setting one cookie.

The cookie SHALL be scoped per room rather than shared across rooms, so that two rooms cannot
recognise the same browser as the same person. Identity does not follow anyone around this product.

The token is opaque and carries no meaning: it is not derived from the name, from the room, or from
anything about the person. It SHALL NOT be readable by scripts running in the page, since nothing in
the page needs it and it is the one value that could be used to take over a seat.

#### Scenario: The seat token never reaches another participant

- **WHEN** several participants are seated and receiving snapshots
- **THEN** no message any of them receives contains any other participant's seat token, and no
  message contains their own either

#### Scenario: Reconnecting reseats the same participant

- **WHEN** a browser that has taken a seat closes its connection and opens a new one to the same
  room, presenting the seat token it was given
- **THEN** it is reseated as the same participant, with the same name and the same vote, and no
  second participant appears at the table

#### Scenario: A different room means a different identity

- **WHEN** the same browser takes a seat in a second room
- **THEN** it is given a separate seat token for that room, and neither room can tell that the two
  participants are the same browser

#### Scenario: A token from a room that is gone starts a new seat

- **WHEN** a browser presents a seat token that names no participant in the room it is connecting
  to
- **THEN** it is not reseated, and it takes a seat as a new participant by giving a name
