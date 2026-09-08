## MODIFIED Requirements

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
