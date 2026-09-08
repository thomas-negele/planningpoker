## MODIFIED Requirements

### Requirement: The process holds a bounded number of rooms and connections

The process SHALL enforce a maximum number of rooms held at one time, and a maximum number of
connections attached to one room. Both SHALL be configurable as described above.

Reaching a limit SHALL refuse only the new work. Rooms already running, connections already
attached and games already in progress SHALL be unaffected, and no participant SHALL be removed from
a seat to make room for anybody else.

A refusal SHALL tell the page why it was refused, distinguishably from a room that does not exist
and from a temporary fault, so the interface can say what is full rather than showing an empty
table or retrying forever.

The two ceilings SHALL be distinguishable **from each other**, not only from those other cases.
Reaching the process's room ceiling and reaching one room's connection ceiling are different
facts about different things, and the useful advice differs: a full server frees up on its own,
while a room holding too many connections is usually one person's spare tabs. A single reason
covering both would tell somebody on an idle server that the server is busy.

Refused work SHALL NOT leave anything behind: no room, no goroutine, no seat token and no entry in
any collection that grows. A client that is refused repeatedly SHALL NOT cost the process more with
each attempt than the refusal itself.

These limits SHALL NOT be used to end a meeting. A connected participant is never disconnected
because a limit exists, however long they stay.

#### Scenario: A room beyond the ceiling is refused, not created

- **WHEN** the process already holds its maximum number of rooms and a connection arrives for a
  room identifier that does not exist yet
- **THEN** no room is created, the connection is told the server is full, and every existing room is
  untouched

#### Scenario: A connection beyond a room's ceiling is refused

- **WHEN** a room already holds its maximum number of connections and another connection arrives for
  it
- **THEN** that connection is refused with a reason the page can display, and the connections
  already attached continue without interruption

#### Scenario: Refused work accumulates nothing

- **WHEN** many connections are refused for capacity in succession
- **THEN** the number of rooms, goroutines and issued seat tokens is the same afterwards as before,
  and memory use does not grow with the number of refusals

#### Scenario: A long meeting is never ended by a limit

- **WHEN** participants stay connected to a room far longer than any timeout in the system, without
  voting or clicking anything
- **THEN** no capacity limit disconnects them, and their room is not discarded

#### Scenario: The two ceilings are told apart

- **WHEN** a connection is refused for the process's room ceiling, and another for a single room's
  connection ceiling
- **THEN** each refusal names its own reason, and neither is reported as the other
