# live-updates Specification

## Purpose

Defines how a browser and the server talk while a game is running: connecting to a room, the intents
a participant may send, the snapshots the server sends back after every change, what happens when an
intent is refused, and how several connections belonging to one person behave.

## Requirements

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

### Requirement: A refused intent is answered with a reason the client can act on

When the server refuses an intent for invalid input or game rules, it SHALL reply to the connection that sent it with a machine-
readable reason, distinct per kind of refusal, so that the interface can say something specific
rather than "something went wrong".

The reasons that must be distinguishable are at least: the name was empty, the name was too long,
the card is not in this room's deck, the round has already been revealed, the sender is not seated
in this room, the room's seats are all taken, the connection is sending too fast, the process is
holding as many rooms as it may, and **this room is holding as many connections as it may**.

The last two are separate entries on purpose. They are different facts and they call for
opposite responses — a full server frees up by waiting, while a room full of connections is
usually somebody's own spare tabs, which they can close now. Telling somebody the server is
busy when it is idle is worse than saying nothing, because it sends them away to wait for
something that will not change.

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

#### Scenario: A full room and a full server are not the same refusal

- **WHEN** a connection is refused because its room already holds as many connections as it may,
  and another is refused because the process already holds as many rooms as it may
- **THEN** the two refusals carry different reasons, and the interface can say which happened

Invalid throw requests SHALL distinguish an unsupported object, self-target, unavailable target,
and an unseated sender. A valid throw discarded because its cosmetic allowance or delivery capacity
is exhausted is a best-effort drop, not a game refusal: it SHALL cause no error banner or retry.
General incoming-message flood refusals retain their existing behaviour.

#### Scenario: A harmless throw limit does not become a game error

- **WHEN** an otherwise valid throw exceeds the participant or room throw allowance
- **THEN** it is discarded without a game error, broadcast, delayed retry or connection closure

#### Scenario: Invalid throw input has a specific refusal

- **WHEN** a seated participant supplies an unknown object, their own target ID, or an unavailable
  target ID
- **THEN** only their connection receives a reason that distinguishes the invalid object,
  self-target and unavailable target

### Requirement: The browser coordinates throw traffic with the configured message allowance

The server SHALL make the connection's applicable message rate and burst allowance available to
the browser together with the fixed throw policy. The browser SHALL wait for this information and
a fresh confirmed seat before enabling throws. Missing policy information SHALL leave throws
unavailable without affecting game controls.

The central connection SHALL account for every locally sent intent when deciding whether another
throw fits. Excess clicks SHALL be dropped without buffering or automatic retry, and outgoing
socket backlog SHALL suppress new throws. At low configured message rates, the effective throw
rate SHALL decrease as necessary, preserving allowance for ordinary game actions. The browser
SHALL NOT assume the default 10 messages per second or treat a local limit as server authority.

#### Scenario: A low server rate does not turn clicks into disconnects

- **WHEN** the server's message rate is set to 1 per second and a user repeatedly clicks throws
  while making occasional ordinary game actions
- **THEN** the browser suppresses excess throws, leaves message capacity for game actions, and
  clicking alone causes neither repeated message-limit errors nor a disconnection

#### Scenario: A game action takes precedence over an unsent throw

- **WHEN** available message capacity is low and the user votes or reveals while clicking throws
- **THEN** the game intent is sent immediately, its cost is accounted for, and throws that do not
  fit the remaining allowance are discarded without delaying that game intent

#### Scenario: A custom client cannot bypass transport protection

- **WHEN** a client bypasses local controls and floods the socket with throw requests or oversized
  messages
- **THEN** the existing server-side message-rate, message-size and persistent-flood protections
  still apply before unbounded decoding or room work can occur

### Requirement: A connection's incoming messages are bounded in rate and size

A single connection SHALL be limited both in how fast it may send messages and in how large one
message may be. The rate SHALL be configurable as a sustained number of messages per second with a
short burst allowance above it, so that a handful of actions arriving together — a reconnecting page
catching up, or somebody voting and immediately revealing — passes through untouched. The message
size limit SHALL be set explicitly to a size this protocol needs, rather than left at whatever the
WebSocket library happens to default to.

Both limits apply **per connection**, not per room. Every participant voting in the same second is
one message on each of their own connections and SHALL never approach the limit; a table of any
permitted size can act simultaneously without being slowed.

Exceeding the rate SHALL be answered with a refusal naming that reason, and SHALL NOT disturb the
room or any other connection. A connection that continues to exceed it may be closed, and closing it
SHALL free everything it held, exactly as an ordinary disconnection does; the participant keeps their
seat and is marked away, as they would be after any dropped connection.

A message larger than the limit SHALL be refused without being processed and without the server
allocating space for its full contents.

Neither limit SHALL be reached by ordinary use. Thinking about an estimate for a long time sends no
messages at all and is therefore never affected.

#### Scenario: Everyone voting at once is unaffected

- **WHEN** every participant in a full room plays a card within the same second, each from their own
  connection
- **THEN** every vote is recorded and no connection is refused or slowed

#### Scenario: One connection flooding is refused, not the room

- **WHEN** a single connection sends messages far faster than the configured rate
- **THEN** that connection is refused with a reason naming the rate, the room is unchanged, and no
  other participant notices anything

#### Scenario: A persistent flood ends only that connection

- **WHEN** a connection keeps exceeding the rate after being refused
- **THEN** that connection may be closed, everything it held is released, and its participant is
  marked away while keeping their seat and their vote

#### Scenario: An oversized message is refused

- **WHEN** a client sends a message larger than the configured size limit
- **THEN** the message is not processed, the room is unchanged, and the server does not hold the
  full message in memory

#### Scenario: A long silence is not a violation

- **WHEN** a connection sends nothing at all for a long time while its participant thinks
- **THEN** neither limit is triggered and the connection is not affected by them

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
