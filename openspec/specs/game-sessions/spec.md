# game-sessions Specification

## Purpose

Defines a game as something that exists over time: how one is created and shared, how long a room
outlives the people sitting in it, what a visitor sees when they arrive at a room that has gone, and
how a browser is recognised as the same participant when it comes back.

## Requirements

### Requirement: Starting a game creates a room and yields an invitation URL

The application SHALL provide a way to create a new game. Creating one produces a room with a fresh
unguessable identifier and returns that identifier, from which the invitation URL is formed.

Creating a game SHALL NOT seat anybody. The person who creates it takes a seat the same way everyone
else does, by opening the URL and giving a name. There is no host and no creator role anywhere in
this product, so there is nothing for the act of creation to confer.

Each creation SHALL produce a distinct room. Creating a game twice never returns the same room.

#### Scenario: Creating a game returns a usable room

- **WHEN** a visitor starts a new game
- **THEN** a room is created, its identifier is returned, and opening the corresponding URL reaches
  that room

#### Scenario: Creating a game seats nobody

- **WHEN** a game has just been created and nobody has opened its URL
- **THEN** the room has no participants, and the person who created it holds no privilege that
  anyone else lacks

#### Scenario: Two games are two rooms

- **WHEN** two games are created in succession
- **THEN** they have different identifiers, and a vote in one is not visible in the other

### Requirement: A room is discarded once nobody has been connected for the grace period

A room SHALL be discarded when no connection has been open to it for longer than a configured grace
period. The same rule covers both cases: a room everyone has left, and a room nobody ever joined.

The grace period exists so that ordinary interruptions do not destroy a session. A page reload, a
laptop lid, a dropped Wi-Fi connection or a few minutes between creating a link and the first person
opening it must all leave the room intact.

The grace period SHALL be readable from an environment variable with a documented default, and its
documentation MUST state what a boundary value means. It is a behaviour-governing value in the sense
of the project's rules, alongside the listen address and the shutdown timeout.

A discarded identifier is not retired. Opening its URL again creates a room there, which is what
makes a memorable name work as a standing address and what lets an interrupted group return. **A
consequence worth stating: a page left open somewhere keeps recreating its room after every restart
and holds it open indefinitely**, so a forgotten tab is what keeps a room alive rather than a
meeting.

#### Scenario: A reload does not destroy the room

- **WHEN** the only participant in a room reloads their page, so that for a moment no connection is
  open
- **THEN** the room still exists when the new connection arrives, with its participants and their
  votes intact

#### Scenario: A deserted room is discarded after the grace period

- **WHEN** every connection to a room has been closed for longer than the grace period
- **THEN** the room is discarded and its participants and votes are gone

#### Scenario: A room nobody ever joined is discarded too

- **WHEN** a game is created and nobody opens its URL within the grace period
- **THEN** that room is discarded on the same rule, without a separate timer or a separate setting

#### Scenario: An occupied room is never discarded

- **WHEN** at least one connection to a room has been open throughout, however long the game runs
- **THEN** the room is not discarded

#### Scenario: The grace period is configurable and documented

- **WHEN** the grace period environment variable is unset
- **THEN** the documented default applies, and both the code and the deployment configuration
  explain in full sentences what the value does and what happens at its boundary

### Requirement: A per-room cookie identifies a returning browser

The application SHALL give each browser an opaque **seat token** for each room it connects to,
stored in a cookie scoped to that room. On reconnecting, the browser presents that token and is
reseated as the same participant, keeping its seat, its name and its vote.

The seat token SHALL be a different value from the participant identifier that appears in snapshots,
and the server SHALL NOT send it to any client other than as that browser's own cookie. The two
values serve opposite purposes and must not be conflated: the participant identifier is public, since
every client needs it to render who is at the table, while the seat token is a credential that proves
a browser owns a seat. Were they the same value, every participant could read every other
participant's credential out of an ordinary snapshot and take their seat by setting one cookie.

The cookie SHALL be scoped per room rather than shared across rooms, so that two rooms cannot
recognise the same browser as the same person. Identity does not follow anyone around this product.

The token is opaque and carries no meaning: it is not derived from the name, from the room, or from
anything about the person. It SHALL NOT be readable by scripts running in the page, since nothing in
the page needs it and it is the one value that could be used to take over a seat.

The cookie SHALL last no longer than the browser session: it carries no expiry date of its own and
is gone when the browser is closed. It is a credential, and its purpose is to survive the
interruptions that happen inside a meeting — a reload, a closed tab, a sleeping laptop, a dropped
network — every one of which leaves the browser running. Outliving the browser buys only the case of
somebody quitting it entirely and returning later, which is not what the token is for and is the
harder thing to justify keeping on somebody's device.

What its absence costs SHALL be understood rather than discovered: a browser returning without it
takes a seat as a new participant, and the participant it used to be stays at the table marked away,
with whatever they had voted still counted in a revealed round. That is the same outcome as losing
the cookie by any other means and is why the token exists at all.

A token names nobody in a room that has just been created, so a browser returning to a recreated
room takes a seat as a new arrival.

#### Scenario: The seat token never reaches another participant

- **WHEN** several participants are seated and receiving snapshots
- **THEN** no message any of them receives contains any other participant's seat token, and no
  message contains their own either

#### Scenario: Reconnecting reseats the same participant

- **WHEN** a browser that has taken a seat closes its connection and opens a new one to the same
  room, presenting the seat token it was given
- **THEN** it is reseated as the same participant, with the same name and the same vote, and no
  second participant appears at the table

#### Scenario: A different room means a different identity

- **WHEN** the same browser takes a seat in a second room
- **THEN** it is given a separate seat token for that room, and neither room can tell that the two
  participants are the same browser

#### Scenario: A token from a room that is gone starts a new seat

- **WHEN** a browser presents a seat token that names no participant in the room it is connecting
  to
- **THEN** it is not reseated, and it takes a seat as a new participant by giving a name

#### Scenario: The token does not outlive the browser

- **WHEN** the seat cookie is issued
- **THEN** it carries no expiry date, so the browser keeps it only for the current session and
  discards it on closing

#### Scenario: Interruptions inside a meeting keep the seat

- **WHEN** a participant reloads the page, closes and reopens the tab, or loses the network for a
  while, without closing the browser
- **THEN** the token is still presented and they are reseated as the same participant

### Requirement: A room identifier is either issued or chosen

A room identifier is what appears after `/g/` in a URL. There are two kinds, and the difference
between them is the whole of what protects a room.

An **issued** identifier is what starting a new game produces: generated from a cryptographically
secure source, carrying at least 128 bits of entropy, rendered in a URL-safe alphabet that excludes
characters easily confused with one another. Nobody reaches such a room without being given the
link, because guessing one is not possible.

A **chosen** identifier is one somebody typed, such as `/g/team-alpha`. It SHALL be accepted if it
is between 5 and 64 characters long and contains only letters, digits, hyphens and underscores.
Anything else — a space, a slash, a punctuation mark, an empty name, something shorter or longer —
SHALL be refused.

The lower bound of five characters exists because very short names are the ones everybody reaches
for at once. `a`, `1`, `test` and `abc` would be typed by different teams within days of each other
and land them in the same room without either knowing why. Five characters is not privacy — a
chosen name never is — but it is enough that a name is something somebody meant rather than
something they happened to press.

A refusal SHALL say what the rule is, not merely that the link is wrong. Somebody who typed a
three-letter name needs to be told a name needs five characters; being told only that it is "not a
game" leaves them with nothing to do about it.

**A chosen identifier is not private.** Somebody who guesses `standup` or `planning` reaches that
room, sees the estimates and can reveal a round or start a new one. This is accepted deliberately: a
memorable address for a recurring meeting is worth more than secrecy over a handful of t-shirt
sizes, and there is nothing else in a room to find.

The interface SHALL NOT distinguish the two kinds. No badge, no warning, no separate wording — a
room is a room. Marking one as public would put a security caveat on a screen people look at all
day, about a product whose entire contents are t-shirt sizes, and would raise a question it cannot
usefully answer. The distinction is recorded here so that it stays a decision on record rather than
a discovery later; it is not something to explain to somebody who wants to estimate a ticket.

Identifiers SHALL be matched exactly, including case. `/g/Team-Alpha` and `/g/team-alpha` are two
different rooms. The address is taken literally and nothing is normalised behind the visitor's back:
what stands in the URL bar is the room they are in. The cost is that somebody typing a remembered
name with different capitalisation lands somewhere else, which is the same thing that happens if
they mistype any other character.

#### Scenario: An issued identifier cannot be guessed

- **WHEN** many games are started in succession
- **THEN** every issued identifier differs from every other, none can be derived from another, and
  each carries at least 128 bits of entropy

#### Scenario: A chosen identifier is accepted

- **WHEN** somebody opens `/g/team-alpha`
- **THEN** that is a valid identifier and reaches a room by that name

#### Scenario: An unusable identifier is refused

- **WHEN** an identifier contains a space or a punctuation mark, is empty, or is longer than 64
  characters
- **THEN** it is refused, and no room comes into existence at it

#### Scenario: A name shorter than five characters is refused, and the rule is given

- **WHEN** somebody opens `/g/abc`
- **THEN** no room is created, the page loads and tells them a room name needs at least five
  characters, and what a name may contain

#### Scenario: Capitalisation is part of the address

- **WHEN** one person opens `/g/Team-Alpha` and another opens `/g/team-alpha`
- **THEN** they are in two different rooms, because the identifier is taken exactly as written

#### Scenario: Neither kind of room is marked as such

- **WHEN** somebody is in a room reached by a chosen name, and again in one reached by an issued
  identifier
- **THEN** the interface looks the same in both and says nothing about the difference

### Requirement: Opening a room's URL reaches a room

Opening the URL of a room SHALL reach that room. If no room exists there yet — because it expired,
because the server was restarted, or because nobody has ever used that name — one SHALL be created,
and the visitor arrives at it.

This replaces the earlier behaviour, in which a URL whose room had gone showed a screen explaining
that the game had ended and offered to bring it back. That screen was a step in front of an outcome
nobody would decline: the person had followed a link, and what they wanted was to be in the room.
Making them confirm it also meant that after a restart every participant had to press the same
button separately, when what they needed was to end up in the same place.

A room created this way is **empty**. No participant, no vote and no round survives, because nothing
was kept. What comes back is the address, not the game that was at it.

Several people opening the same URL at once SHALL all reach one room. Whichever request arrives
first brings it into existence and the rest find it; none replaces another or strands anybody who
had already arrived.

An identifier that is refused by the rules above SHALL NOT produce a room. The page SHALL still
load, and SHALL tell the visitor which rule the name broke, so that a name they can fix is a name
they can fix.

A room SHALL come into existence only for a request that has actually become a live connection.
A request that reaches the socket path but never completes the upgrade — an ordinary HTTP request,
a request refused because it came from another website, a request refused because the process is at
capacity — SHALL create no room and SHALL be issued no seat token. This is what keeps the act of
creating a room behind something a browser can only do deliberately, rather than behind any request
that happens to carry a plausible name in its path.

Starting a game over HTTP SHALL likewise be refused when the request comes from another website, so
that a page elsewhere cannot create rooms in a visitor's name.

#### Scenario: An old link simply works again

- **WHEN** somebody opens the URL of a room that has expired
- **THEN** they arrive at an empty room at that URL, with nothing to confirm first

#### Scenario: A restart puts everybody back together

- **WHEN** the server is restarted while several people have the same room open
- **THEN** their pages reconnect and all of them end up in one room again at the same URL

#### Scenario: A name nobody has used yet becomes a room

- **WHEN** somebody opens `/g/team-alpha` and no room of that name exists
- **THEN** a room is created there and they arrive at it

#### Scenario: Simultaneous arrivals converge

- **WHEN** several people open the same URL at the same moment and no room exists there
- **THEN** exactly one room comes into existence and all of them are in it

#### Scenario: A link that is not a game says so

- **WHEN** somebody opens a URL whose identifier is refused — too short, containing a space, or far
  too long
- **THEN** the page loads and tells them what a room name must look like, offers to start a game,
  and nothing is created

#### Scenario: A request that never becomes a connection creates nothing

- **WHEN** an ordinary HTTP request is made to the socket path of a room name that does not exist,
  without the headers that would upgrade it to a WebSocket
- **THEN** the request is refused, no room comes into existence at that name, and no seat cookie is
  returned

#### Scenario: A foreign origin creates nothing

- **WHEN** a page on another website tries to open a connection to a room name that does not exist,
  or to start a game
- **THEN** the attempt is refused, no room comes into existence, and no seat cookie is returned

#### Scenario: A malformed link reaches the same explanation

- **WHEN** somebody opens a room URL whose identifier cannot be decoded at all, such as one
  containing an invalid percent-escape
- **THEN** the page loads and tells the visitor the link is not a game, exactly as it does for an
  identifier that is merely refused, rather than failing to render
