## ADDED Requirements

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
