## MODIFIED Requirements

### Requirement: A refused intent is answered with a reason the client can act on

When the server refuses an intent, it SHALL reply to the connection that sent it with a machine-
readable reason, distinct per kind of refusal, so that the interface can say something specific
rather than "something went wrong".

The reasons that must be distinguishable are at least: the name was empty, the name was too long,
the card is not in this room's deck, the round has already been revealed, the sender is not seated
in this room, the room's seats are all taken, and the connection is sending too fast.

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

## ADDED Requirements

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
