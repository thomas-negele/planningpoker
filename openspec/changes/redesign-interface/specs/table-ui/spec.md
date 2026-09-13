## ADDED Requirements

### Requirement: Text and controls are legible against their background

The interface SHALL meet the contrast ratios of WCAG 2.1 level AA against the background each
element actually sits on: at least 4.5:1 for body text, at least 3:1 for large text (24px, or
18.66px when bold, and above), and at least 3:1 for the visible boundary of an interactive control,
for a focus indicator, and for any graphic that carries meaning rather than decoration.

A colour SHALL NOT be used for text on a surface where it fails that ratio, however well it works
elsewhere. An accent that is legible as a filled button is not thereby legible as coloured text on a
light surface, and the two uses are judged separately.

This is a property of the palette that is checked rather than assumed. Whoever changes a colour
SHALL measure the pairs that colour takes part in, not reason about them.

#### Scenario: Body text meets the ratio

- **WHEN** the contrast ratio between any body text and the surface behind it is measured
- **THEN** it is at least 4.5:1

#### Scenario: A control can be seen and its focus can be seen

- **WHEN** the contrast ratio between an interactive control's visible boundary, or its focus
  indicator, and the surface behind it is measured
- **THEN** it is at least 3:1

#### Scenario: A face-down card is distinguishable from the table

- **WHEN** a participant has played a card in a hidden round
- **THEN** the face-down card is distinguishable from the surface behind it at a ratio of at least
  3:1, so that "has voted" is legible without relying on colour perception alone

### Requirement: Movement respects a stated preference against it

Where the interface moves — objects thrown across the table, cards lifting, surfaces fading — it
SHALL honour the `prefers-reduced-motion: reduce` setting the visitor's system reports, by removing
the movement or reducing it to a minimum.

Removing the movement SHALL NOT remove the information it carried. A thrown object still arrives and
is still seen; it simply does not fly. A played card is still shown as played; it simply does not
rise. Anything a participant could learn from a moving interface SHALL remain learnable from a still
one.

#### Scenario: Thrown objects arrive without flying

- **WHEN** a participant whose system asks for reduced motion is thrown an object
- **THEN** they see that the object was thrown and by whom, without it being animated across the
  screen

#### Scenario: The interface still says what it said

- **WHEN** a participant whose system asks for reduced motion plays a card, reveals a round and
  starts a new one
- **THEN** every state the animated interface would have shown is visible in the still one

## MODIFIED Requirements

### Requirement: The entry screen offers one thing

The base URL SHALL show an entry screen dedicated to starting a new game. The start control SHALL
come first, and the deck choice SHALL be presented below it as a list of options — one per supported
deck — with `T-shirt sizes` selected by default. There is no list of games, no way to search for one
and no way to enter a room identifier by hand: a room is reached only by its invitation link, which
is what makes the link the thing that protects it.

Each option SHALL name its deck and SHALL show that deck's cards, in the deck's own order, separated
by commas, so that somebody who has never seen the deck can tell what they are choosing before they
choose it. Those card values SHALL come from the server rather than from a list written into the
page, under the same rule that already governs the deck at the table: the server is the only place a
deck is defined.

If the card values cannot be obtained, the entry screen SHALL still offer both decks by name and
SHALL still start a game. Not knowing what is in a deck is a smaller failure than not being able to
start one.

Starting a game SHALL create a room with the selected deck and take the visitor to it, at a URL they
can copy and send to others. Choosing a deck does not seat the visitor or remember a preference for
later rooms.

#### Scenario: Starting a game uses the visible selection

- **WHEN** a visitor selects Fibonacci on the entry screen and starts a new game
- **THEN** a Fibonacci room is created and they arrive at that room's own invitation URL

#### Scenario: Starting a game arrives at a table

- **WHEN** a visitor opens the base URL, keeps either supported deck selected, and starts a new game
- **THEN** a room using that deck is created and they arrive at that room's own URL, which is the
  address that invites everyone else

#### Scenario: T-shirt sizes are preselected

- **WHEN** a visitor opens the entry screen and starts a game without changing the deck choice
- **THEN** the new room uses the t-shirt deck

#### Scenario: Each option names its cards

- **WHEN** a visitor looks at the deck options
- **THEN** each option shows that deck's cards in the deck's own order, separated by commas

#### Scenario: The cards shown are the cards dealt

- **WHEN** a visitor starts a game with the deck whose cards the entry screen listed
- **THEN** the deck at the table offers exactly those cards, in that order

#### Scenario: A game can still be started when the cards are unknown

- **WHEN** the entry screen cannot obtain the card values
- **THEN** both decks are still offered by name and starting a game still works

#### Scenario: The entry screen offers nothing else

- **WHEN** a visitor looks at the entry screen
- **THEN** they can select a deck and start a game, but cannot browse, search for, or type in a room

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
The prompt SHALL offer that choice, **not selected by default**. The choice itself
SHALL say, in its own label and without anything being opened, that the name would be
kept on this device; what exactly is stored and for how long SHALL be available from
an information control beside it. Only when the choice is selected is the name kept;
otherwise nothing is written and nothing is read back, and the visitor types their
name each time. Turning the choice off again SHALL delete what was stored, so that
the same control both grants and withdraws it.

When the choice has been made, the field SHALL arrive pre-filled for a returning
visitor and the choice SHALL still show as made, so that what is on screen matches
what is on the device. The visitor may change the name before confirming.

Wherever a name is entered, the interface SHALL make available, from an information
control beside the name field, that a first name or nickname is enough and that
everyone with the link to the room can see it. It SHALL be worded identically in
every such place, so the same fact is not stated two ways.

An information control SHALL be reachable by every means of operating the page, not
only by a pointer: it SHALL disclose its text on hover, on keyboard focus and on tap,
and its text SHALL be associated with the control it explains so that a screen reader
announces the two together. Hover alone would put the text out of reach on a
touchscreen and for anybody working without a mouse, which for the storage choice
would mean deciding about what is kept on their device without being able to read
what that is.

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

- **WHEN** a visitor opens the information control beside the offer to remember their
  name
- **THEN** it states what is stored and how long it is kept, in plain words and
  without leaving the page

#### Scenario: The offer is not misleading before it is opened

- **WHEN** a visitor looks at the offer to remember their name without opening its
  information
- **THEN** the choice itself already says that the name would be kept on this device,
  so nothing is stored on the strength of a label that did not say so

#### Scenario: The information is reachable without a pointer

- **WHEN** a visitor reaches an information control by keyboard, or taps it on a
  touchscreen
- **THEN** its text is disclosed, exactly as it is when a pointer hovers it

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

- **WHEN** a visitor opens the information control beside the name field
- **THEN** it tells them a first name or nickname is enough and that everyone with the
  link can see it

#### Scenario: A full name is still accepted

- **WHEN** somebody enters a full name despite the hint
- **THEN** it is accepted and used exactly like any other name, and nothing warns,
  blocks or asks again
