## MODIFIED Requirements

### Requirement: The server sends whole snapshots of the room, never partial updates

After any change to a room's game state, the server SHALL send every connection to that room a complete snapshot
of the room's state, including each participant's visitor mode, rather than a description of what changed.

A snapshot is small — a handful of participants and their votes — and sending the whole thing makes
a reconnecting client correct by construction, with no replay of missed events and no possibility of
two browsers disagreeing about what happened.

A connection SHALL receive a snapshot immediately on connecting, before anything else, so that a
browser has a complete picture from its first moment.

The client holds no authoritative state of its own: it renders what it is sent and sends intents
back. It MUST NOT be necessary for a client to have seen an earlier message to understand a later
one.

#### Scenario: Connecting yields the current state at once

- **WHEN** a browser opens a connection to a room in which people are already seated and voting
- **THEN** it receives a snapshot showing every participant and the state of the round before it
  receives anything else

#### Scenario: Every change reaches everyone

- **WHEN** any participant takes a seat, votes, reveals, renames themselves, changes visitor mode or starts a new round
- **THEN** every connection to that room receives a fresh snapshot reflecting it

#### Scenario: A snapshot stands on its own

- **WHEN** a browser has missed any number of earlier messages
- **THEN** the next snapshot it receives is sufficient to render the room correctly

Transient throws are separate cosmetic events, not partial updates to game state. They SHALL NOT
replace snapshots or be needed to interpret one. Even during throw traffic, the first server
message on a new connection SHALL be its complete snapshot; throw delivery is eligible only after
that snapshot has been written.

#### Scenario: Transient effects do not replace the initial snapshot

- **WHEN** a browser connects while other participants are throwing objects
- **THEN** it receives a complete snapshot before any throw event, and later snapshots remain
  sufficient to render the game without any earlier events

#### Scenario: Visitor mode is part of the snapshot

- **WHEN** a participant changes visitor mode
- **THEN** every connection receives a complete snapshot showing the new mode and any change to hidden vote status

### Requirement: A participant sends intents and the server decides

A connected browser SHALL be able to send exactly these intents: take a seat under a name and visitor-mode choice, play a
card, reveal the round, start a new round, select a supported deck, change its name and/or visitor mode, and throw one
of the supported objects at another present participant. Every rule about whether an intent is
allowed lives on the server; the client may hide an action it believes is unavailable, but hiding it
is never what enforces it.

An unrecognised or malformed message SHALL be refused without disturbing the room or the connection.
A client that sends nonsense affects nobody else.

#### Scenario: An intent that is allowed takes effect

- **WHEN** a seated participant plays a card while the round is hidden
- **THEN** the vote is recorded and every connection receives a snapshot showing that participant has
  voted

#### Scenario: A rule is enforced by the server, not the client

- **WHEN** a client sends a vote after the round has been revealed, whether by a stale interface or
  by a message constructed by hand
- **THEN** the server refuses it, no vote changes, and the client is told why

#### Scenario: A malformed message harms nobody

- **WHEN** a client sends a message that is not valid JSON, or names an intent that does not exist
- **THEN** the room is unchanged, other participants notice nothing, and the sender is told the
  message was not understood

#### Scenario: Throw eligibility is enforced on the server

- **WHEN** a client constructs a throw request that the interface would not offer
- **THEN** the server still validates the actual seated sender, object and target, applies the
  throw allowances, and never trusts client-supplied identity or animation coordinates

#### Scenario: A visitor vote is refused by the server

- **WHEN** a visitor sends a vote intent directly, bypassing the browser controls
- **THEN** the server refuses it with a visitor-mode reason and broadcasts no changed vote
