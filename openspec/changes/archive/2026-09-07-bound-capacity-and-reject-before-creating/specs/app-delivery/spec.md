## MODIFIED Requirements

### Requirement: Behaviour-governing values are set by environment variable

Every value that governs how the running process behaves SHALL be readable from an environment
variable and SHALL have a documented default that applies when the variable is unset or empty. Each
such value MUST be accompanied, in the code and in `compose.yaml`, by a full-sentence explanation
of what it does and what a boundary value means.

The values governed by this rule are the network address the server listens on, the shutdown
timeout described in the requirement below, the grace period after which an abandoned room is
discarded, and the four capacity limits: the number of rooms the process will hold, the number of
connections one room will hold, the number of participants one room will seat, and the rate at
which one connection may send messages.

An unparseable or invalid value SHALL cause the process to fail at startup with a message naming
the variable and the offending value, rather than silently falling back to the default — a
misconfiguration that is silently ignored is worse than one that stops the process.

A capacity limit SHALL be refused at startup if it is zero or negative. There is deliberately no
value meaning "no limit": the whole purpose of these settings is that some ceiling exists, and an
operator who wants effectively no ceiling can say so with a large number, which is honest about
what it costs.

#### Scenario: Default applies when unset

- **WHEN** the process starts with no listen-address variable set
- **THEN** it listens on the documented default address and logs which address it is listening on

#### Scenario: Environment variable overrides the default

- **WHEN** the process starts with the listen-address variable set to a valid address
- **THEN** it listens on that address

#### Scenario: Invalid value stops the process

- **WHEN** the process starts with the listen-address variable set to a value that cannot be used
- **THEN** the process exits with a non-zero status and an error message naming the variable and
  the value

#### Scenario: Shutdown timeout is configurable with a documented default

- **WHEN** the process starts with no shutdown-timeout variable set
- **THEN** it uses the documented default, and that default is shorter than the grace period the
  container runtime allows before it kills the process outright

#### Scenario: Ambiguous shutdown timeout is refused

- **WHEN** the process starts with the shutdown-timeout variable set to zero, to a negative
  duration, or to something that is not a duration at all
- **THEN** the process exits with a non-zero status and an error naming the variable and the value,
  because a boundary value with no agreed meaning is worse than no setting at all

#### Scenario: Capacity limits are configurable with documented defaults

- **WHEN** the process starts with none of the capacity-limit variables set
- **THEN** it applies the documented default for each, and each default is explained in
  `compose.yaml` in full sentences saying what it bounds and what happens when it is reached

#### Scenario: A capacity limit of zero is refused

- **WHEN** the process starts with any capacity-limit variable set to zero, to a negative number, or
  to something that is not a number
- **THEN** the process exits with a non-zero status and an error naming the variable and the value

## ADDED Requirements

### Requirement: The process holds a bounded number of rooms and connections

The process SHALL enforce a maximum number of rooms held at one time, and a maximum number of
connections attached to one room. Both SHALL be configurable as described above.

Reaching a limit SHALL refuse only the new work. Rooms already running, connections already
attached and games already in progress SHALL be unaffected, and no participant SHALL be removed from
a seat to make room for anybody else.

A refusal SHALL tell the page why it was refused, distinguishably from a room that does not exist
and from a temporary fault, so the interface can say that the server is full rather than showing an
empty table or retrying forever.

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
