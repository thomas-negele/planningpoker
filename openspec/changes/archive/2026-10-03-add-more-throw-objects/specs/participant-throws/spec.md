## RENAMED Requirements

- FROM: `### Requirement: A seated participant can throw one of three objects at another present participant`
- TO: `### Requirement: A seated participant can throw a supported object at another present participant`

## MODIFIED Requirements

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

## ADDED Requirements

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
