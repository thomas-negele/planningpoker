## Why

Testing the finished product turned up a case the specification handled badly. A
server restart ended three participants' game at once, each of them saw "This game
has ended", each pressed the only button on offer, and the three of them ended up in
three separate rooms. Nothing was broken; that is exactly what was specified.

The specification forbade recreating a room at an identifier that had been used, on
the grounds that doing so would "destroy the property that makes an identifier
private". That reasoning does not hold up. An identifier carries 128 bits of
randomness, so nobody reaches a room without having been given the link — and
somebody who was given the link is precisely the person who should be able to get
back in.

A first attempt at this put a "reopen this game" button on the ended screen. Trying
it made the remaining question obvious: nobody would ever decline that offer. The
person had followed a link; what they wanted was to be in the room. A confirmation
step in front of an outcome nobody refuses is a step for its own sake, and it also
meant every participant had to press the same button separately after a restart.

So the button is gone before it was ever archived, and opening the URL is enough.

## What Changes

- **Opening a room's URL reaches a room.** If none exists there — expired, restarted,
  or never used — one is created and the visitor arrives at it. No screen, no
  confirmation.
- **BREAKING (to the specification, not to any client)**: the rule forbidding a room
  at a previously used identifier is replaced, and the screen that announced a game
  had ended is removed with it.
- **Room names may be chosen.** `/g/team-alpha` is a room called `team-alpha`, so a
  recurring meeting can have a standing address that nobody needs to send round.
  Anything up to 64 characters of letters, digits, hyphens and underscores is
  accepted; matching ignores capitalisation, so a remembered name typed differently
  still finds the same room.
- **A chosen name is not private, and the specification says so.** Somebody who
  guesses `standup` is in that room. This is accepted knowingly: there is nothing in
  a room but t-shirt sizes, and a memorable address is worth more than secrecy over
  them. An issued identifier is still unguessable, so the two kinds are genuinely
  different and the difference is written down.
- **Losing a round is announced; losing a connection is not.** A dismissible message
  appears only when somebody who was seated finds the room recreated underneath them
  — a restart, or an expiry while everyone was briefly away. An ordinary
  reconnection, which loses nothing, keeps saying nothing beyond the connection
  status it already showed.
- **The ended screen becomes the not-a-game screen**, reached only by a URL the
  server refuses outright.

### Also in this change: corrections for adjustments made while testing

Several interface adjustments were applied directly during testing, because each
repaired behaviour the specification already required. The text has not caught up
with them, and leaving it describing an interface that no longer exists is exactly
the drift this process exists to prevent:

- The **Reveal control is hidden once the round is revealed**, and **New round sits
  outside the table**.
- On a **narrow screen the participants are a list below the table**, because a ring
  collides with itself at that width; and no participant's card overlaps the table on
  a wide one.
- The **sentence shown while somebody was alone is gone** from beside the invitation
  control, along with the plumbing that told it how many people were present.

## Capabilities

### New Capabilities

None. Everything here changes behaviour that is already described.

### Modified Capabilities

- `game-sessions`: what an identifier may be, and what opening one does. The
  prohibition on creating a room at a used identifier is replaced by creating one;
  chosen names are admitted alongside issued ones; and the screen announcing a game
  had ended is removed.
- `connection-resilience`: reconnection no longer has an exception for a game that
  has ended, because that state no longer exists. What stops the page for good is now
  only a refused identifier. A new requirement covers what a participant is told when
  a round is genuinely lost.
- `table-ui`: three requirements corrected to describe the interface as it now is —
  the controls, the arrangement of participants on a narrow screen, and the
  invitation control standing on its own.

## Impact

- **Modified code**: `internal/game` (which identifiers are acceptable, and rooms
  under a chosen name), `internal/hub` (creating on demand, case-insensitive
  lookup), `internal/transport` (the socket creates a missing room; the reopen route
  and its button disappear), `web/` (no ended screen for a missing room, the
  announcement, the not-a-game screen).
- **No new dependencies and no new configuration.**
- **Two consequences worth stating plainly**, because both are permanent and neither
  is obvious:
  - An invitation link no longer expires in any practical sense. It stops leading to
    a live room, but it never stops being a way to make one.
  - A page left open somewhere keeps recreating its room after every restart and
    holds it open indefinitely. What keeps a room alive is a forgotten tab, not a
    meeting.
