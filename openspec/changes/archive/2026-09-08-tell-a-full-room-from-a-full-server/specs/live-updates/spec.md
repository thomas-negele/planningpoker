## MODIFIED Requirements

### Requirement: A refused intent is answered with a reason the client can act on

When the server refuses an intent, it SHALL reply to the connection that sent it with a machine-
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
