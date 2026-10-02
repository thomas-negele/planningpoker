# participant-throws Specification

## Purpose

Defines playful throws between participants as shared, short-lived effects, including target
eligibility and resource limits that keep estimation responsive.

## Requirements

### Requirement: A seated participant can throw a supported object at another present participant

The server SHALL accept a paper ball, paper plane, single flower or heart — and, when the operator
has enabled it, a pile of poo — from a seated participant targeting another present participant in
the same room. The sender SHALL be identified from their connection, never from a client-supplied
sender identity. Voting status and whether the round is revealed SHALL NOT restrict throwing.
Self-targets, away or unknown targets, unseated senders, and unsupported objects SHALL be rejected
without modifying the room or emitting a throw event.

#### Scenario: All three objects work throughout a round

- **WHEN** a seated participant throws a paper ball, a paper plane and a flower at another present
  participant before voting, after that participant votes, and after reveal
- **THEN** each otherwise admissible throw is accepted without changing any vote or round state

#### Scenario: The heart and the pile of poo work throughout a round

- **WHEN** the pile of poo is enabled and a seated participant throws a heart and a pile of poo at
  another present participant before voting, after that participant votes, and after reveal
- **THEN** each otherwise admissible throw is accepted without changing any vote or round state

#### Scenario: The new objects follow the existing rules and limits

- **WHEN** a seated participant throws hearts or piles of poo at themselves, at an away participant,
  or more often than the participant or room allowance permits
- **THEN** they are refused or silently discarded exactly as the existing three objects would be,
  sharing the same per-participant and per-room allowances rather than receiving their own

#### Scenario: A target must be someone else at this table

- **WHEN** a request targets the sender, an away participant, or an identifier absent from this room
- **THEN** the sender receives a specific refusal and nobody receives a throw event

#### Scenario: The server validates the sender and object

- **WHEN** an unseated connection requests a throw, or a seated connection supplies an unsupported
  object
- **THEN** the request is refused without mutation or broadcast

#### Scenario: Another tab cannot impersonate a different sender

- **WHEN** a request includes a forged sender identity
- **THEN** it cannot cause a throw attributed to that identity; authority comes from the connection's
  actual seat

### Requirement: Accepted throws are shared transient effects

An accepted throw SHALL be offered to every current connection in its room, including the sender,
target, and their other tabs, through a small event containing public sender and target identities,
object type, and shared variation data. Healthy recipients SHALL observe the same object, entry
side and variation, adapted to their own viewport and seat layout; exact frame timing across
different devices is not required.

Throws SHALL NOT become game state, contain private credentials or hidden vote values, or trigger
a room snapshot. Late connections and reconnecting clients SHALL receive no throw history.
Delivery SHALL be best effort: backlogged recipients can miss cosmetic events, with no retry,
replay or requirement for reliable game updates to wait for them.

#### Scenario: Everyone at this table sees the throw

- **WHEN** a throw is accepted with healthy connections in two rooms
- **THEN** only connections in the sender's room receive the event, including the sender's and
  target's tabs, with the same public identities, object and variation data

#### Scenario: A reconnect does not replay objects

- **WHEN** a participant reconnects or joins after earlier throws
- **THEN** the initial snapshot describes the game without old throw events or resting objects

#### Scenario: A throw does not disclose hidden information

- **WHEN** throws occur during a hidden round
- **THEN** their messages contain neither hidden vote values nor private seat credentials and no
  extra game snapshot is emitted because of a throw

### Requirement: Throw admission is bounded across tabs and across the room

The server SHALL accept at most 3 throws from one participant and at most 12 throws in one room
in any rolling one-second interval. These ceilings SHALL be fixed constants. Multiple connections
for one seat SHALL share the participant allowance, and reconnecting to the same retained seat
SHALL NOT reset it. Rejected attempts SHALL NOT consume another participant's allowance.

The client SHALL discard excess clicks locally without sending or deferring them. The server
SHALL independently discard requests exceeding the throw allowances without broadcasting,
queueing a future throw, presenting an error banner, or disconnecting anyone for exceeding those
throw allowances alone. The existing general message flood protection remains in force.

#### Scenario: Repeated clicks do not create a future burst

- **WHEN** a participant clicks rapidly on one or several targets
- **THEN** local excess clicks produce no network message, at most 3 throws per rolling second are
  accepted for the participant, and clicking stops without a queued sequence playing afterwards

#### Scenario: Tabs and reconnects share one allowance

- **WHEN** the same seat sends throws from several tabs and reconnects within the same second
- **THEN** the combined accepted count remains at most 3 within that rolling second

#### Scenario: The room ceiling applies to simultaneous senders

- **WHEN** several participants submit more than 12 eligible throws within a rolling second
- **THEN** the room accepts at most 12, silently discards excess requests, and does not defer them

#### Scenario: A window boundary grants no double allowance

- **WHEN** requests arrive immediately before and after a whole-second boundary
- **THEN** every rolling one-second interval still contains at most 3 accepted throws per
  participant and at most 12 for the room

### Requirement: Cosmetic work remains bounded and yields to game activity

Throw processing, pending delivery, live browser objects and their cleanup SHALL have bounded
resource use. The server SHALL perform no per-frame simulation or animation broadcast. Throw
traffic SHALL NOT fill the queue used for reliable room updates, delay them behind a cosmetic
backlog, or close a connection merely because its cosmetic delivery capacity is exhausted.

When resources are busy, the system SHALL discard cosmetic work and continue processing game
actions. Existing connection deadlines and flood protection SHALL still remove dead or abusive
connections. Client animation work SHALL end when all objects expire, when the connection is
lost, or when the user leaves the room.

#### Scenario: A slow cosmetic consumer leaves room updates usable

- **WHEN** one connection does not consume its pending throw events while others continue playing
- **THEN** cosmetic storage remains bounded, excess effects are dropped for that connection,
  and reliable updates remain independent of that backlog

#### Scenario: Voting continues during maximum throw traffic

- **WHEN** the room receives sustained permitted throws alongside votes, reveal and new-round
  actions
- **THEN** those game actions still take effect, their snapshots take priority over pending
  cosmetic work, and resource use does not grow with the session's duration

#### Scenario: Leaving removes animation work

- **WHEN** the browser disconnects, navigates away, or the last object finishes fading
- **THEN** the corresponding live objects and unnecessary animation callbacks are released

### Requirement: The pile of poo is available only when the operator enables it

The pile of poo SHALL be governed by the environment variable `PLANNINGPOKER_POO_THROWS`. It SHALL
be disabled when the variable is unset, empty or `false`, and enabled only when it is `true`. Any
other value SHALL stop the process at startup with an error naming the variable and the value. While
disabled, the server SHALL treat a pile of poo as an unsupported object. The setting takes effect on
the next start of the process.

#### Scenario: A default installation has no pile of poo

- **WHEN** the process starts without `PLANNINGPOKER_POO_THROWS`, or with it empty or `false`
- **THEN** no participant is offered the pile of poo, and a request to throw one, even from a
  custom client, receives the existing unknown-object refusal without mutation or broadcast

#### Scenario: An operator enables the pile of poo

- **WHEN** the process starts with `PLANNINGPOKER_POO_THROWS` set to `true`
- **THEN** every room on that server offers and accepts the pile of poo alongside the other four
  objects

#### Scenario: An unclear value stops the process

- **WHEN** the process starts with `PLANNINGPOKER_POO_THROWS` set to any value other than empty,
  `true` or `false`, such as `yes`, `1` or `TRUE`
- **THEN** the process exits with a non-zero status and an error naming the variable and the value

#### Scenario: The heart does not depend on the switch

- **WHEN** the process starts with the pile of poo disabled
- **THEN** the heart is still offered and accepted like the paper ball, paper plane and flower
