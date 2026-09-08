## Purpose

Defines how a browser and the server talk while a game is running: connecting to a room, the intents
a participant may send, the snapshots the server sends back after every change, what happens when an
intent is refused, and how several connections belonging to one person behave.

## ADDED Requirements

### Requirement: The server sends whole snapshots of the room, never partial updates

After any change to a room, the server SHALL send every connection to that room a complete snapshot
of the room's state, rather than a description of what changed.

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

- **WHEN** any participant takes a seat, votes, reveals, renames themselves or starts a new round
- **THEN** every connection to that room receives a fresh snapshot reflecting it

#### Scenario: A snapshot stands on its own

- **WHEN** a browser has missed any number of earlier messages
- **THEN** the next snapshot it receives is sufficient to render the room correctly

### Requirement: A hidden vote never crosses the network

While a round is hidden, a snapshot SHALL report only *whether* each participant has voted. It MUST
NOT contain any vote's value, nor a count of votes per card, nor anything else from which a value
could be recovered.

This is the guarantee the whole product rests on, and it is the reason the rules were written to make
a hidden card unrepresentable rather than merely omitted. The transport layer SHALL send what the
rules give it and MUST NOT reach past them for state to include.

#### Scenario: A hidden round transmits no card

- **WHEN** every participant has voted and the round has not been revealed
- **THEN** no message any client receives contains any card value, as observed on the network rather
  than in the interface

#### Scenario: Revealing transmits the cards

- **WHEN** the round is revealed
- **THEN** the next snapshot contains every participant's card and the count per card

### Requirement: A participant sends intents and the server decides

A connected browser SHALL be able to send exactly these intents: take a seat under a name, play a
card, reveal the round, start a new round, and change its name. Every rule about whether an intent is
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

### Requirement: A refused intent is answered with a reason the client can act on

When the server refuses an intent, it SHALL reply to the connection that sent it with a machine-
readable reason, distinct per kind of refusal, so that the interface can say something specific
rather than "something went wrong".

The reasons that must be distinguishable are at least: the name was empty, the name was too long,
the card is not in this room's deck, the round has already been revealed, and the sender is not
seated in this room.

A refusal SHALL be sent only to the connection that caused it. Nobody else at the table learns that
someone tried something that was not allowed.

#### Scenario: Refusals are distinguishable

- **WHEN** intents are refused for different reasons
- **THEN** each refusal carries its own identifiable reason rather than a single generic failure

#### Scenario: A refusal is private

- **WHEN** one participant's intent is refused
- **THEN** only that participant's connection receives the refusal, and no other connection receives
  anything as a result

#### Scenario: A refusal changes nothing

- **WHEN** an intent is refused
- **THEN** the room is exactly as it was, and no snapshot is sent on account of it

### Requirement: One participant may hold several connections

A participant SHALL be able to hold more than one open connection to the same room, which is what
happens when they open a second tab. All of their connections receive the same snapshots.

Several connections are one seat at the table: the participant appears once, and an action taken on
one connection is reflected on the others.

A participant SHALL be marked away only when the last of their connections closes, and the away mark
SHALL be cleared when a new one opens.

#### Scenario: A second tab is not a second participant

- **WHEN** a browser opens the same room in a second tab, presenting the identifier it already holds
- **THEN** the table still shows one participant for that person, and both connections receive
  snapshots

#### Scenario: An action on one connection appears on the other

- **WHEN** a participant votes in one of their tabs
- **THEN** their other tab receives a snapshot showing that they have voted

#### Scenario: Away means the last connection closed

- **WHEN** a participant with two open connections closes one of them
- **THEN** they are not marked away; and when they close the second, they are

### Requirement: One goroutine owns each room

Each room's state SHALL be owned by exactly one goroutine. No other goroutine reads or writes it
directly; everything reaches a room by sending it a message.

The interesting logic therefore runs single-threaded and needs no locking, and the shared structure
that remains — the collection of rooms — is guarded separately and protects nothing but itself.

A connection that stops reading MUST NOT be able to stall the room or the other participants. If a
slow or stuck connection cannot keep up, that connection is dropped rather than allowed to hold up
everyone else.

The whole design SHALL be exercised under Go's race detector, since a data race here would be the
kind of fault that appears only under load and only in production.

#### Scenario: Concurrent activity produces a consistent room

- **WHEN** many browsers connect, take seats, vote, reveal and start new rounds against the same room
  at the same time
- **THEN** the room's state stays consistent, every connection receives well-formed snapshots, and
  the race detector reports nothing

#### Scenario: A stuck connection does not stall the room

- **WHEN** one connection stops reading what the server sends it
- **THEN** the other participants continue unaffected, and the stuck connection is eventually dropped
  rather than blocking the room

#### Scenario: Shutdown closes rooms and connections

- **WHEN** the process is asked to stop
- **THEN** every room goroutine ends and every connection is closed, within the shutdown budget and
  without leaking a goroutine
