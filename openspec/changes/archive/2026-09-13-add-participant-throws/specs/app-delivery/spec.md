## MODIFIED Requirements

### Requirement: Behaviour-governing values are set by environment variable

Except for the explicitly fixed participant-throw limits described below, every value that governs
how the running process behaves SHALL be readable from an environment
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

Participant throws are an explicit exception chosen for this feature: at most 3 accepted throws
per participant and 12 per room in a rolling second SHALL be fixed, named constants rather than
new environment variables or runtime settings. Documentation SHALL explain that throw quotas are
additional to the existing configurable per-connection message limit, and that reaching a throw
quota drops cosmetic work without closing the connection.

#### Scenario: Throw ceilings are fixed and documented

- **WHEN** the application starts with any valid existing deployment configuration
- **THEN** the throw ceilings remain 3 per participant and 12 per room per rolling second,
  documented as fixed constants, while the configured general message limit still applies
