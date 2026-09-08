## MODIFIED Requirements

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
- **THEN** the page loads and tells them the link is not a game, exactly as it does for an
  identifier that is merely refused, rather than failing to render
