## Purpose

Defines what a person sees and does: starting a game, giving a name before sitting
down, the table and the people around it, playing and changing a card, revealing the
round, reading the results, starting a fresh one, renaming, and inviting others.

## ADDED Requirements

### Requirement: The entry screen offers one thing

The base URL SHALL show an entry screen whose only action is to start a new game.
There is no list of games, no way to search for one and no way to enter a room
identifier by hand: a room is reached only by its invitation link, which is what
makes the link the thing that protects it.

Starting a game SHALL create a room and take the visitor to it, at a URL they can
copy and send to others.

#### Scenario: Starting a game arrives at a table

- **WHEN** a visitor opens the base URL and starts a new game
- **THEN** a room is created and they arrive at that room's own URL, which is the
  address that invites everyone else

#### Scenario: The entry screen offers nothing else

- **WHEN** a visitor looks at the entry screen
- **THEN** starting a new game is the only action available, and there is no way to
  browse, search for or type in a room

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

The name SHALL be remembered in the browser and the field SHALL arrive pre-filled
for a returning visitor, so that nobody retypes their name every week. They may
change it before confirming.

An empty name SHALL NOT be accepted, and a name too long for the rules SHALL be
refused with a message saying so rather than silently failing. Because these are the
rules' limits, the refusal SHALL come from the server's answer rather than being
guessed at by the page — the page may check first as a courtesy, but the server's
refusal is what is displayed.

While waiting at the name prompt the visitor SHALL be able to see who is already at
the table, so that opening a link tells them whether they are in the right meeting.

#### Scenario: A returning visitor confirms rather than retypes

- **WHEN** somebody who has played before opens a room URL
- **THEN** the name prompt appears with their previous name already filled in, and
  confirming it seats them

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

#### Scenario: The table is visible before joining it

- **WHEN** a visitor is at the name prompt for a room that already has people in it
- **THEN** they can see who is there

### Requirement: The table shows everyone and what they have done

The table SHALL show every participant, arranged around it, each with their name.
For each participant it SHALL show whether they have played a card and whether they
are away, and it SHALL be apparent which one is you.

While the round is hidden, a participant who has voted SHALL be shown as holding a
face-down card, and one who has not SHALL be shown as holding nothing. The value
MUST NOT be shown, or discoverable, for anybody — including yourself, since anything
the page knows is in reach of everyone at the table.

The table SHALL remain readable as people arrive and leave, and on a narrow screen.

#### Scenario: A face-down card means a vote was cast

- **WHEN** somebody plays a card while the round is hidden
- **THEN** everyone sees that they are holding a card, and nobody sees which

#### Scenario: Away participants stay visible

- **WHEN** a participant's connection drops
- **THEN** they remain at the table, marked as away, and their face-down card stays
  where it was

#### Scenario: You can tell which one is you

- **WHEN** a participant looks at the table
- **THEN** their own seat is distinguishable from the others

### Requirement: The deck is always to hand

The cards the room's deck offers SHALL be shown along the bottom edge of the screen,
in the deck's own order, and SHALL stay reachable without scrolling the table away.

The deck SHALL be built from what the server sends rather than from a list written
into the page, so that a room offering a different deck later needs no change here.

Playing a card SHALL show it as played. Playing another SHALL replace it, as the
rules allow, with no separate step to withdraw the first.

There is one case where the played card cannot be shown, and it follows from the
hidden-vote guarantee rather than from any shortcoming here. A hidden round discloses
no card value to anybody — deliberately including the voter's own, because a message
sent to one person is still a message on the network. So a page that did not itself
cast the vote knows from the snapshot *that* this participant has voted, but cannot
know *which* card. That is the case after a page reload, and in a second tab opened
later.

In that case the deck SHALL show no card as played, while the table still shows the
participant holding a face-down card. Showing a guess would be worse than showing
nothing.

A dropped and re-established connection is **not** that case: the page is still the
one that cast the vote and has not forgotten it, so the card stays shown throughout.

Cards SHALL be operable by keyboard as well as by pointer.

#### Scenario: The deck comes from the server

- **WHEN** the table is shown
- **THEN** the cards offered are exactly those the server named, in the order it gave
  them

#### Scenario: Playing a card shows it as played

- **WHEN** a participant plays a card while the round is hidden
- **THEN** that card is visibly the one they hold

#### Scenario: Changing a card replaces it

- **WHEN** a participant who has played a card plays a different one
- **THEN** the new card is the one they hold, with no intermediate state in which
  they hold both or none

#### Scenario: A reloaded page does not guess which card was played

- **WHEN** a participant who has voted reloads the page while the round is still
  hidden
- **THEN** the table shows them holding a face-down card, and the deck shows no card
  as played, because the value is genuinely unknown to that page

#### Scenario: A repaired connection keeps showing the played card

- **WHEN** a participant's connection drops and is re-established while the round is
  still hidden
- **THEN** the deck still shows the card they played, because the page never lost
  its own note of it

#### Scenario: The deck is reachable by keyboard

- **WHEN** a participant navigates with the keyboard alone
- **THEN** every card can be reached and played

### Requirement: Revealing and starting a new round are available to everyone

The controls to reveal the round and to start a fresh one SHALL be available to
every seated participant, at any time. The interface MUST NOT hide or disable them
according to who created the game or whether everyone has voted, because the rules
place no such condition on them and an interface that pretends otherwise teaches
people something untrue about the product.

The interface MAY show that everyone present has voted, as information. It MUST NOT
turn that into a precondition.

#### Scenario: Anybody may reveal

- **WHEN** any seated participant chooses to reveal
- **THEN** the round is revealed for everybody

#### Scenario: Revealing early is not prevented

- **WHEN** some participants have not yet voted
- **THEN** revealing is still offered, and using it reveals the round

#### Scenario: A fresh round is always available

- **WHEN** the round is hidden or revealed
- **THEN** starting a new round is offered, and it clears every card

### Requirement: Results show the cards and the count per card

Once revealed, the table SHALL show each participant's card in their own place, and
separately a count of how many people played each card. A participant who did not
vote SHALL be shown as holding no card, which is distinct from having played the
unknown card.

No average, median or other number derived by arithmetic over the cards SHALL be
shown, because t-shirt sizes are an ordered scale without arithmetic and any such
number would be invented.

#### Scenario: Every card is shown in its place

- **WHEN** the round is revealed
- **THEN** each participant's card appears at their seat

#### Scenario: The spread is visible at a glance

- **WHEN** the round is revealed and several people played the same card
- **THEN** a count per card is shown, in the deck's order

#### Scenario: Nothing is averaged

- **WHEN** any round is revealed
- **THEN** no average or other computed number over the cards appears anywhere

### Requirement: A participant may change their own name

A seated participant SHALL be able to change their own display name from the table,
and SHALL NOT be able to change anybody else's.

The same refusals apply as when joining, and are shown the same way.

#### Scenario: Renaming yourself works

- **WHEN** a participant changes their own name
- **THEN** everyone at the table sees the new name, and their card and seat are
  unaffected

#### Scenario: Nobody can rename anybody else

- **WHEN** a participant looks at another person's name
- **THEN** there is no way to edit it

### Requirement: The invitation link is always available in one place

The table SHALL offer a way to copy the room's invitation link, in a fixed position
that does not move as people arrive or leave. Anybody at the table may use it, not
only whoever created the game.

Using it SHALL copy the link and confirm that it did. If the browser refuses
clipboard access, the link SHALL be shown in a form that can be selected and copied
by hand, so that the action never dead-ends.

While a participant is the only person in the room, the interface SHALL additionally
say so and point at that control. That prompt SHALL disappear once somebody else is
seated; the control itself SHALL NOT move.

#### Scenario: Anyone can invite a latecomer

- **WHEN** a meeting is under way and somebody needs the link
- **THEN** any participant can copy it from the same place it has always been

#### Scenario: Being alone is pointed out once

- **WHEN** a participant is the only one at the table
- **THEN** the interface says so and directs them to the invitation control

#### Scenario: The prompt goes but the control stays

- **WHEN** a second participant sits down
- **THEN** the prompt about being alone disappears and the invitation control remains
  exactly where it was

#### Scenario: A refused clipboard still yields the link

- **WHEN** the browser does not permit writing to the clipboard
- **THEN** the link is displayed so it can be selected and copied by hand

### Requirement: The interface renders what the server sent

The page SHALL render the most recent snapshot it received and SHALL NOT keep its
own authoritative copy of the room. An action taken by a participant is sent as an
intent; what appears on the table is what comes back.

A refusal from the server SHALL be shown to the participant who caused it, in words
that match the reason given, rather than as a generic failure.

#### Scenario: The table follows the server

- **WHEN** the server sends a snapshot that differs from what is displayed
- **THEN** the display changes to match it

#### Scenario: A refusal says what was wrong

- **WHEN** an intent is refused
- **THEN** the participant is told specifically what was refused, distinguishing at
  least an empty name, a name that is too long, a card the deck does not hold, and a
  round that has already been revealed
