# connection-resilience Specification

## Purpose

Defines how the page stays connected to its room and what it tells the person in
front of it when it is not: reconnecting after a drop, telling a temporary fault
apart from a game that has ended, and never leaving a stale table on screen
pretending to be live.

## Requirements

### Requirement: The page reconnects on its own after a dropped connection

When the connection to a room is lost, the page SHALL try to re-establish it without anybody having
to press anything. There is no longer an exception for a game that has ended: a room that has gone
comes back on the next attempt, which is what puts an interrupted group together again.

Attempts SHALL be spaced by a delay that grows after each failure rather than repeating immediately,
so that a server that is briefly unavailable is not made worse by every open browser retrying at
once. The delay SHALL be capped, so that a page left open through a long outage still recovers
promptly when the server returns rather than waiting out an ever-longer interval. The delay SHALL
carry some randomness, so that every participant of a meeting whose server was just redeployed does
not reconnect in the same instant.

On reconnecting, the page SHALL present the seat it already holds, so that the participant is put
back in the same chair with the same name and the same vote — whenever that seat still exists.

This is the requirement that makes the seat cookie and the room grace period do what they were built
for. Without it neither is ever exercised by a real person.

#### Scenario: A brief interruption repairs itself

- **WHEN** a participant's connection drops and comes back within the room's grace period
- **THEN** the page reconnects without any action from them, and the table shows them in their
  original seat with their vote intact

#### Scenario: A sleeping laptop comes back to a live table

- **WHEN** a participant closes their laptop, opens it later within the grace period, and the socket
  had closed in the meantime
- **THEN** the page reconnects and the table is current, not as it was when the lid closed

#### Scenario: Retries slow down rather than hammering

- **WHEN** several attempts to reconnect fail in succession
- **THEN** each attempt waits longer than the one before, up to a fixed maximum

#### Scenario: A room that has gone is reconnected to, not given up on

- **WHEN** the room a page is connected to ceases to exist and the page reconnects
- **THEN** it reaches a room at that URL rather than stopping

### Requirement: The connection state is always visible and never misleading

The page SHALL make the state of its connection apparent at all times: established, re-establishing,
or stopped for good.

While the connection is not established, the interface MUST NOT present the table as current. A
table that has stopped receiving updates but still looks live is the worst of the possible failures,
because somebody can sit reading stale information with no reason to doubt it.

Controls that send an intent SHALL be unavailable while there is no connection to send it over,
rather than accepting a click that goes nowhere.

The connection state SHALL be conveyed to assistive technology as well as visually, and MUST NOT
rely on colour alone.

#### Scenario: A lost connection is admitted immediately

- **WHEN** the connection drops
- **THEN** the page says so at once, and the table is no longer presented as current

#### Scenario: Actions are not accepted into nowhere

- **WHEN** the connection is not established
- **THEN** playing a card, revealing, starting a new round and renaming are all unavailable rather
  than silently doing nothing

#### Scenario: Recovery is announced as clearly as failure

- **WHEN** the connection is re-established
- **THEN** the page says so and the table is current again

#### Scenario: The state does not depend on seeing colour

- **WHEN** the connection state changes
- **THEN** it is conveyed by text as well as by any colour, and is available to a screen reader

### Requirement: A link that is not a game is not retried

When the server refuses an identifier outright, the page SHALL stop trying and SHALL say so,
offering to start a new game. Retrying would fail for as long as the page was left open, because
nothing about a refused identifier changes with time.

This is now the only case that ends a connection for good. A room that has merely gone is not one:
the page reconnects and a room comes into existence there, which is what puts an interrupted group
back together without anybody pressing anything.

It SHALL be distinguishable from a temporary fault: the visitor is told the link is not a game, not
that something went wrong.

#### Scenario: A refused link is explained, once

- **WHEN** somebody opens a URL whose identifier the server refuses
- **THEN** they are told it is not a game and offered to start one, and the page does not retry

#### Scenario: A restarted server does not end the connection

- **WHEN** the server is restarted while a participant has the table open
- **THEN** their page reconnects and finds a room at the same URL, rather than stopping

#### Scenario: A fault is not reported as a refusal

- **WHEN** the connection drops for a reason other than the identifier being refused
- **THEN** the visitor is told the connection was lost and that it is being re-established

### Requirement: Losing a round is announced, losing a connection is not

The page SHALL tell a participant when the round they were in has been lost, and SHALL NOT make a
fuss about anything less than that.

A dropped connection is not a loss. The room outlives it, the seat and the vote are still there, and
the participant returns to exactly what they left. The connection state already says "reconnecting",
which is as much as that deserves.

What is a loss is the room having gone and been created again underneath somebody who was sitting in
it: their seat, their vote and the whole round are gone, and everybody must sit down again. A server
restart is the usual cause; a room that expired while everyone was briefly disconnected is the
other.

The page SHALL recognise that case by what it can observe — it held a seat, it reconnected, and the
room it reached does not contain it — rather than by guessing from the reason a socket closed.

The announcement SHALL be brief, SHALL say what was lost rather than naming a technical cause, and
SHALL be dismissible. It MUST NOT block the table or require an answer: by the time it appears the
participant can already sit down again, and the message is there to explain why they must.

#### Scenario: A recreated room is announced

- **WHEN** a participant who was seated finds, on reconnecting, that the room no longer knows them
- **THEN** they are told the round was lost and that they need to take a seat again, in a message
  they can dismiss

#### Scenario: An ordinary reconnection says nothing extra

- **WHEN** a participant's connection drops and is re-established with their seat and vote intact
- **THEN** no announcement appears; the connection state alone reported it

#### Scenario: Somebody arriving fresh is told nothing

- **WHEN** somebody opens a room URL for the first time and a room is created for them
- **THEN** nothing is announced, because they have lost nothing

#### Scenario: The announcement does not block the table

- **WHEN** the announcement is showing
- **THEN** the participant can still take a seat, play a card and use every control, and dismissing
  it requires nothing of them
