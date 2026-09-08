## Purpose

Defines how the page stays connected to its room and what it tells the person in
front of it when it is not: reconnecting after a drop, telling a temporary fault
apart from a game that has ended, and never leaving a stale table on screen
pretending to be live.

## ADDED Requirements

### Requirement: The page reconnects on its own after a dropped connection

When the connection to a room is lost for any reason other than the game having
ended, the page SHALL try to re-establish it without anybody having to press
anything.

Attempts SHALL be spaced by a delay that grows after each failure rather than
repeating immediately, so that a server that is briefly unavailable is not made
worse by every open browser retrying at once. The delay SHALL be capped, so that a
page left open through a long outage still recovers promptly when the server
returns rather than waiting out an ever-longer interval.

On reconnecting, the page SHALL present the seat it already holds, so that the
participant is put back in the same chair with the same name and the same vote.

This is the requirement that makes the seat cookie and the room grace period do
what they were built for. Without it neither is ever exercised by a real person.

#### Scenario: A brief interruption repairs itself

- **WHEN** a participant's connection drops and comes back within the room's grace
  period
- **THEN** the page reconnects without any action from them, and the table shows
  them in their original seat with their vote intact

#### Scenario: A sleeping laptop comes back to a live table

- **WHEN** a participant closes their laptop, opens it later within the grace
  period, and the socket had closed in the meantime
- **THEN** the page reconnects and the table is current, not as it was when the lid
  closed

#### Scenario: Retries slow down rather than hammering

- **WHEN** several attempts to reconnect fail in succession
- **THEN** each attempt waits longer than the one before, up to a fixed maximum

### Requirement: A game that has ended is not retried

When the server reports that the room no longer exists, the page SHALL stop trying
and SHALL tell the visitor that the game is no longer available, offering to start a
new one.

It MUST NOT keep reconnecting to something that cannot come back. A room is
discarded once nobody has been connected for the grace period, and nothing ever
recreates one at the same identifier, so retrying is guaranteed to fail forever.

This SHALL be distinguishable from a temporary fault: the visitor is told the game
has ended, not that something went wrong.

#### Scenario: An expired invitation is explained, once

- **WHEN** somebody opens the link to a room that has been discarded
- **THEN** they are told the game is no longer available and offered a new one, and
  the page does not retry

#### Scenario: A restarted server ends the games it was holding

- **WHEN** the server is restarted while a participant has the table open
- **THEN** their page discovers the room is gone and says the game has ended, rather
  than retrying indefinitely or showing an empty table

#### Scenario: A fault is not reported as an ending

- **WHEN** the connection drops for a reason other than the room being gone
- **THEN** the visitor is told the connection was lost and that it is being
  re-established, not that their game has ended

### Requirement: The connection state is always visible and never misleading

The page SHALL make the state of its connection apparent at all times:
established, re-establishing, or ended.

While the connection is not established, the interface MUST NOT present the table as
current. A table that has stopped receiving updates but still looks live is the
worst of the possible failures, because somebody can sit reading stale information
with no reason to doubt it.

Controls that send an intent SHALL be unavailable while there is no connection to
send it over, rather than accepting a click that goes nowhere.

The connection state SHALL be conveyed to assistive technology as well as visually,
and MUST NOT rely on colour alone.

#### Scenario: A lost connection is admitted immediately

- **WHEN** the connection drops
- **THEN** the page says so at once, and the table is no longer presented as current

#### Scenario: Actions are not accepted into nowhere

- **WHEN** the connection is not established
- **THEN** playing a card, revealing, starting a new round and renaming are all
  unavailable rather than silently doing nothing

#### Scenario: Recovery is announced as clearly as failure

- **WHEN** the connection is re-established
- **THEN** the page says so and the table is current again

#### Scenario: The state does not depend on seeing colour

- **WHEN** the connection state changes
- **THEN** it is conveyed by text as well as by any colour, and is available to a
  screen reader
