## MODIFIED Requirements

### Requirement: A name is given before taking a seat

Opening a room's URL, by a visitor who does not already hold a seat in that room,
SHALL show a name prompt rather than seating them immediately. Only after they
confirm are they at the table and visible to everyone else.

A visitor who *does* already hold a seat in that room SHALL be returned to it
directly, without being asked to confirm a name they have already given. This is the
case a page reload falls into, and `connection-resilience` requires it — it is the
entire purpose of the seat cookie. The two rules do not conflict: the prompt exists
so that arriving somewhere new is a deliberate act, not so that reloading becomes a
small interruption every time.

The name SHALL NOT be stored on the visitor's device unless they have asked for it.
The prompt SHALL offer that choice, **not selected by default**, and SHALL say beside
it what would be stored and for how long. Only when it is selected is the name kept;
otherwise nothing is written and nothing is read back, and the visitor types their
name each time. Turning the choice off again SHALL delete what was stored, so that
the same control both grants and withdraws it.

When the choice has been made, the field SHALL arrive pre-filled for a returning
visitor and the choice SHALL still show as made, so that what is on screen matches
what is on the device. The visitor may change the name before confirming.

Wherever a name is entered, the interface SHALL say that a first name or nickname is
enough and that everyone with the link to the room can see it. It SHALL be worded
identically in every such place, so the same fact is not stated two ways.

This is a hint and SHALL NOT become a rule: no name is refused for being fuller than
suggested, nothing about a name is inspected, and the game behaves identically
whatever somebody enters. The point is only that the person deciding what to type
knows who will see it, at the moment they decide.

Choosing not to store a name SHALL NOT prevent anybody from playing, and nothing in
the game SHALL behave differently for somebody who declined.

Renaming at the table SHALL follow the same choice: it updates a stored name only if
the visitor asked for one, and never creates one on its own.

An empty name SHALL NOT be accepted, and a name too long for the rules SHALL be
refused with a message saying so rather than silently failing. Because these are the
rules' limits, the refusal SHALL come from the server's answer rather than being
guessed at by the page — the page may check first as a courtesy, but the server's
refusal is what is displayed.

The name field SHALL additionally stop accepting characters once the maximum
permitted length (see `room-membership`) has been typed, so that a name too long to
be accepted cannot be composed in the first place. This is the same courtesy check
by another means and it does not replace the rule above: the field's cap prevents the
common case, and a name that reaches the server too long anyway — pasted, autofilled,
or sent by something that is not this page — is still refused by the server, and that
refusal is still what gets displayed. The page MUST NOT decide for itself that a name
is acceptable and skip sending it.

While waiting at the name prompt the visitor SHALL be able to see who is already at
the table, so that opening a link tells them whether they are in the right meeting.
Names in that list SHALL be shown in full, under the same rule as at the table.

#### Scenario: Nothing is stored unless it was asked for

- **WHEN** a visitor types a name, leaves the offer to remember it untouched, and
  confirms
- **THEN** they are seated, and nothing about them is written to their device by the
  page

#### Scenario: A returning visitor confirms rather than retypes

- **WHEN** somebody who previously asked for their name to be remembered opens a room
  URL
- **THEN** the name prompt appears with that name already filled in, the choice shows
  as made, and confirming it seats them

#### Scenario: Declining is not a lesser experience

- **WHEN** a visitor who declined to have their name stored plays a full round
- **THEN** everything works exactly as it does for somebody who accepted, and they are
  not asked again during that session

#### Scenario: Withdrawing the choice deletes what was stored

- **WHEN** a visitor who had asked to be remembered turns that choice off
- **THEN** the stored name is deleted from their device, and the next visit shows an
  empty field

#### Scenario: The offer says what it means

- **WHEN** a visitor looks at the offer to remember their name
- **THEN** it states what is stored and how long it is kept, in plain words and
  without having to open another page

#### Scenario: Renaming does not store a name behind anyone's back

- **WHEN** a participant who declined storage changes their name at the table
- **THEN** the new name is used in the game and nothing is written to their device

#### Scenario: Opening a link does not seat anybody

- **WHEN** somebody who holds no seat in that room opens its URL and does not confirm
  the name prompt
- **THEN** they are not at the table, and nobody else sees them

#### Scenario: Reloading does not ask again

- **WHEN** a seated participant reloads the page
- **THEN** they are returned straight to the table in the same seat, with no name
  prompt to confirm

#### Scenario: An unusable name is refused with a reason

- **WHEN** a visitor confirms an empty name, or one longer than the rules allow
- **THEN** they are told which of the two is wrong and stay at the prompt

#### Scenario: Typing past the limit is not possible

- **WHEN** a visitor at the name prompt keeps typing after reaching the maximum
  permitted name length
- **THEN** the further characters do not appear in the field, and the name that would
  be submitted is exactly the permitted length

#### Scenario: An over-long name that reaches the server is still refused

- **WHEN** a name longer than the maximum arrives at the server despite the field's
  cap, for instance by being pasted or autofilled
- **THEN** the server refuses it and the visitor is shown that refusal, rather than
  the page silently accepting or silently shortening the name

#### Scenario: The table is visible before joining it

- **WHEN** a visitor is at the name prompt for a room that already has people in it
- **THEN** they can see who is there, each name shown in full

#### Scenario: The prompt says who will see the name

- **WHEN** a visitor is at the name prompt
- **THEN** it tells them a first name or nickname is enough and that everyone with the
  link can see it

#### Scenario: A full name is still accepted

- **WHEN** somebody enters a full name despite the hint
- **THEN** it is accepted and used exactly like any other name, and nothing warns,
  blocks or asks again

### Requirement: A participant may change their own name

A seated participant SHALL be able to change their own display name from the table,
and SHALL NOT be able to change anybody else's.

Changing it SHALL open the same two controls offered at the name prompt — the name
and the choice about storing it — rather than only the name. This is what makes the
choice reversible in practice: it is given at the prompt, and without a way back to
it from the table, somebody who agreed could never withdraw. Withdrawal must be as
easy as agreement, and a control that can only grant is not the control this design
claims to be.

Withdrawing SHALL delete the stored name immediately, at the moment the choice is
turned off, rather than when the change is saved. Somebody who withdraws and then
abandons the dialog must not find their name still on the device.

Agreeing from the table SHALL store the name in the same way as agreeing at the
prompt, so the two routes to the same decision behave identically.

The sentence about who can see the name SHALL appear here too, in the same words,
because changing a name is the same decision made again.

The same refusals apply as when joining, and are shown the same way.

#### Scenario: Renaming yourself works

- **WHEN** a participant changes their own name
- **THEN** everyone at the table sees the new name, and their card and seat are
  unaffected

#### Scenario: Nobody can rename anybody else

- **WHEN** a participant looks at another person's name
- **THEN** there is no way to edit it

#### Scenario: The choice is reachable from the table

- **WHEN** a seated participant opens their own name to change it
- **THEN** the choice about storing that name is shown alongside it, in the same words
  as at the name prompt, and set to whatever is currently true of their device

#### Scenario: Withdrawing from the table deletes the stored name

- **WHEN** a seated participant who had agreed turns the choice off
- **THEN** the stored name is deleted at once, and it stays deleted even if they then
  abandon the change without saving

#### Scenario: Agreeing from the table stores the name

- **WHEN** a seated participant who had declined at the prompt turns the choice on and
  saves
- **THEN** their name is stored exactly as if they had agreed at the prompt

#### Scenario: Changing a name says who will see it

- **WHEN** a seated participant opens their own name to change it
- **THEN** the same sentence shown at the name prompt tells them a first name or
  nickname is enough and that everyone with the link can see it
