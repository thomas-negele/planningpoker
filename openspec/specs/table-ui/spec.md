# table-ui Specification

## Purpose

Defines what a person sees and does: starting a game, giving a name before sitting
down, the table and the people around it, playing and changing a card, revealing the
round, reading the results, starting a fresh one, renaming, and inviting others.

## Requirements

### Requirement: The entry screen offers one thing

The base URL SHALL show an entry screen dedicated to starting a new game. It SHALL offer two compact
deck options, labelled `T-shirt sizes` and `Fibonacci`, directly above the start control, with
`T-shirt sizes` selected by default. There is no list of games, no way to search for one and no way
to enter a room identifier by hand: a room is reached only by its invitation link, which is what
makes the link the thing that protects it.

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

#### Scenario: The entry screen offers nothing else

- **WHEN** a visitor looks at the entry screen
- **THEN** they can select a deck and start a game, but cannot browse, search for, or type in a room

### Requirement: Room deck settings stay unobtrusive and truthful

Every seated participant SHALL have a small icon-only room-settings control immediately beside the
invitation control, with the settings control to its left. Its accessible name SHALL identify it as
room settings, and its placement and size SHALL keep it visually secondary to voting, revealing,
starting a new round, and inviting participants.

Activating the control SHALL open a compact dialog offering exactly `T-shirt sizes` and `Fibonacci`.
The dialog SHALL identify the active deck. After a revealed round it SHALL also make clear that a
different selection is for the next round, and it SHALL show the latest pending selection to every
participant.

While the current round is hidden and contains at least one vote, the settings control SHALL remain
visible but disabled. A tooltip and equivalent accessible help text SHALL explain that the deck
cannot be changed while voting is in progress. The control SHALL be enabled while a hidden round has
no votes and after a round is revealed, subject only to connection availability.

The settings control and dialog SHALL work by keyboard and at narrow screen widths without obscuring
or displacing the invitation and round controls.

#### Scenario: Settings sit beside the invitation control

- **WHEN** a seated participant views the room on a wide or narrow screen
- **THEN** a small settings icon appears immediately left of the invitation control without
  competing visually with the primary game actions

#### Scenario: An empty round allows an immediate choice

- **WHEN** the current round is hidden and nobody has voted
- **THEN** the settings control is enabled, and choosing a deck makes it the active deck for the
  room

#### Scenario: Voting disables settings with a reason

- **WHEN** the current round is hidden and at least one participant has voted
- **THEN** the settings control remains visible but disabled, and pointer and assistive-technology
  users can discover that the deck cannot change while voting is in progress

#### Scenario: A revealed round offers the next deck

- **WHEN** the round has been revealed and a seated participant opens room settings
- **THEN** the control is enabled and the dialog distinguishes the revealed round's active deck from
  the deck selected for the next round

#### Scenario: Everyone sees the latest pending choice

- **WHEN** one seated participant changes the pending selection after a revealed round
- **THEN** every participant's settings dialog shows that latest selection for the next round

#### Scenario: Settings are keyboard operable

- **WHEN** a seated participant uses only the keyboard
- **THEN** they can reach the settings control, open and close the dialog, inspect both options, and
  select an allowed deck

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

### Requirement: The table shows everyone and what they have done

The table SHALL show every participant, each with their name. For each participant it SHALL show
whether they have played a card and whether they are away, and it SHALL be apparent which one is you.

A name of any length the rules permit SHALL be displayed in full, at every screen width. It MUST NOT
be clipped, shortened, replaced by an ellipsis or otherwise altered to fit the space available.
Truncation is not an acceptable outcome for a name the product itself accepted: it presents somebody
under a name they did not choose, with nothing on screen to explain why. Where a name and its
surroundings do not fit, it is the surroundings that give way.

The same rule applies to the field in which a participant edits their own name: it SHALL hold a name
of the full permitted length visibly, so that nobody edits a name they cannot entirely see.

On a wide screen the participants SHALL be arranged around the table. Each one SHALL sit clear of
the table's edge, measured from the middle of the whole seat — the card and the name together —
rather than from the card alone, so that nobody's card ends up lying on the table. This clearance
SHALL be derived from the seat's actual extent, which includes whatever width a full-length name
requires, so that a wider seat does not reintroduce the overlap.

On a narrow screen they SHALL instead be a list below the table, one participant per row. A ring
collides with itself at that width, and a row of full-size cards costs more vertical space than a
phone has to give before the table has even begun. The table itself carries the result and the
action, so it comes first and the list of who is present follows it.

The card SHALL remain a card in both arrangements: face down for somebody who has played, face up
once revealed, and visibly empty for somebody who has not played.

While the round is hidden, a participant who has voted SHALL be shown as holding a face-down card,
and one who has not SHALL be shown as holding nothing. The value MUST NOT be shown, or discoverable,
for anybody — including yourself, since anything the page knows is in reach of everyone at the table.

The table SHALL remain readable as people arrive and leave, and on a narrow screen.

#### Scenario: A face-down card means a vote was cast

- **WHEN** somebody plays a card while the round is hidden
- **THEN** everyone sees that they are holding a card, and nobody sees which

#### Scenario: Away participants stay visible

- **WHEN** a participant's connection drops
- **THEN** they remain at the table, marked as away, and their face-down card stays where it was

#### Scenario: You can tell which one is you

- **WHEN** a participant looks at the table
- **THEN** their own seat is distinguishable from the others

#### Scenario: A name of the permitted length is shown whole

- **WHEN** a participant is seated under a name of the maximum permitted length, on a wide screen and
  again on a narrow one
- **THEN** every character of that name is visible at their seat, with no ellipsis and no clipping

#### Scenario: A full-length name can be edited in full

- **WHEN** a participant opens the field to change their own name and it holds a name of the maximum
  permitted length
- **THEN** the whole name is visible in the field without scrolling it sideways

#### Scenario: No card lies on the table

- **WHEN** participants are arranged around the table on a wide screen, at any position including
  the diagonals, each under a name of the maximum permitted length
- **THEN** no participant's card overlaps the table, and no seat overlaps another

#### Scenario: A narrow screen lists the participants below the table

- **WHEN** the table is shown on a screen too narrow for a ring
- **THEN** the participants appear as a list below the table, one per row, each still showing their
  card and whether they are away

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

The controls to reveal the round and to start a fresh one SHALL be available to every seated
participant. The interface MUST NOT hide or disable either according to who created the game or
whether everyone has voted, because the rules place no such condition on them and an interface that
pretends otherwise teaches people something untrue about the product.

Starting a new round SHALL be offered at all times, whether the round is hidden or revealed, and
SHALL sit outside the table so that it does not compete with the result for the same space.

Revealing SHALL be offered while the round is hidden, and SHALL NOT be offered once it has been
revealed — at that point it has nothing left to do. This is not a permission: anybody may reveal, at
any moment the round is still hidden. Removing it afterwards MUST NOT move anything else on the
screen, because a control that shifts position the instant somebody presses it is worse than a
little empty space.

The interface MAY show that everyone present has voted, as information. It MUST NOT turn that into a
precondition.

#### Scenario: Anybody may reveal

- **WHEN** any seated participant chooses to reveal
- **THEN** the round is revealed for everybody

#### Scenario: Revealing early is not prevented

- **WHEN** some participants have not yet voted
- **THEN** revealing is still offered, and using it reveals the round

#### Scenario: Revealing is no longer offered once done

- **WHEN** the round has been revealed
- **THEN** the reveal control is no longer shown, and nothing else on the screen has moved as a
  result

#### Scenario: A fresh round is always available

- **WHEN** the round is hidden or revealed
- **THEN** starting a new round is offered, away from the table, and it clears every card

### Requirement: Results show the cards and the count per card

Once revealed, the table SHALL show each participant's card in their own place, and
separately how many people played each card. A participant who did not vote SHALL be
shown as holding no card, which is distinct from having played the unknown card.

The count per card SHALL be presented as a bar chart on the table: one horizontal bar
per card, whose length is proportional to the number of participants who played it,
so that the spread of the round is legible as a shape rather than as a set of
numbers to be read and compared. Each bar SHALL also carry its card and its exact
count in text, because a bar alone cannot distinguish four votes from five and the
precise number is sometimes what the discussion turns on.

The bars SHALL be stacked vertically along the deck's own ordering of sizes, presented
with the largest at the top and the smallest at the bottom, so the scale reads with the
big work above the small.

Every card of the deck that expresses a size SHALL have a row in every revealed
round, including cards nobody played, which appear with an empty bar and a count of
zero. The chart is therefore a fixed scale rather than a list of what happened to be
chosen: its height and its rows are the same in every round, which is what makes two
consecutive rounds on the same item comparable at a glance and what stops the table
resizing between them.

Cards that do not express a size — in the t-shirt deck these are `?` and the coffee
cup — SHALL NOT appear on that scale, because placing "I cannot estimate this" and "I
need a break" between two sizes would imply an ordering that does not exist. They
SHALL instead be shown beneath the scale, side by side on one shared row, and only
when at least one of them was played. They are chosen rarely and a permanently
reserved pair of rows would cost more of the table than it earns.

Which cards form the size scale and which do not SHALL be determined by what the
server says about the room's deck, not by a list written into the page. A second deck
added later must not require the page to be taught its exceptions.

The interface MUST NOT declare an outcome for the round. It SHALL NOT mark, highlight,
label or otherwise single out a winning card, a majority, a consensus, a mode or an
outlier, and — as before — SHALL NOT show an average, median or any other number
derived by arithmetic over the cards. Two of these are wrong for different reasons and
both matter: arithmetic over t-shirt sizes invents a measurement that does not exist,
and naming a winner ends the conversation the round was held to start. Interpreting
the spread is the moderator's job, and the interface's job is to show the spread
plainly enough that they can.

The table SHALL be large enough to hold the fullest chart the room's deck can produce
— every size row, the shared row of non-size cards, and every bar's count — inside
the table itself, legibly and without reducing the text below the size used elsewhere
on the table, without any part of the chart scrolling, clipping or spilling over the
seats around it. Where the table is not large enough for that, the table SHALL be made
larger.

The displayed results SHALL be horizontally centred and placed within the table with
space above and below. The results and the space reserved for the reveal control may
be arranged as one group; the results need not have exactly equal space above and below
them independently of that control. Reserving the control's space remains required —
see the requirement on revealing — so hiding it does not shift the surrounding layout.

The table SHALL be the same size whether the round is hidden or revealed. The seats
are positioned relative to the table, so a table that grew on reveal would move every
participant at the moment somebody pressed the button.

#### Scenario: Every card is shown in its place

- **WHEN** the round is revealed
- **THEN** each participant's card appears at their seat

#### Scenario: The spread is visible at a glance

- **WHEN** the round is revealed and the votes are divided unevenly between cards
- **THEN** the counts appear as horizontal bars whose lengths are in proportion to
  those counts, so the division is apparent without reading the numbers

#### Scenario: The scale runs from large to small

- **WHEN** the round is revealed
- **THEN** the size cards appear as rows stacked from the largest at the top to the
  smallest at the bottom, following the deck's ordering of sizes

#### Scenario: The cards that are not sizes stay at the foot

- **WHEN** the round is revealed and `?` or the coffee cup was played
- **THEN** they appear below every size row, not above them and not between them,
  whichever way round the size scale is ordered

#### Scenario: Unchosen sizes keep their row

- **WHEN** the round is revealed and nobody played one of the sizes
- **THEN** that size still has its row, with an empty bar and a count of zero, and the
  chart has the same height it would have had if it were chosen

#### Scenario: The non-size cards share one row below the scale

- **WHEN** the round is revealed and at least one participant played `?` or the coffee
  cup
- **THEN** both appear together on a single row beneath the size scale, side by side,
  and neither appears among the sizes

#### Scenario: The shared row is absent when unused

- **WHEN** the round is revealed and nobody played `?` or the coffee cup
- **THEN** no row for them is shown at all

#### Scenario: The page is not told which cards are sizes

- **WHEN** the server describes a deck
- **THEN** which of its cards belong on the ordered scale is taken from that
  description, and the page contains no list of card values of its own

#### Scenario: No card is declared the winner

- **WHEN** the round is revealed and one card has more votes than every other
- **THEN** that card's bar is longer, and nothing marks, highlights or labels it as
  the winner, the majority or the result

#### Scenario: Nothing is averaged

- **WHEN** any round is revealed
- **THEN** no average, median or other computed number over the cards appears anywhere

#### Scenario: The chart fits on the table

- **WHEN** a round is revealed in which every size was played and `?` and the coffee
  cup were played as well
- **THEN** the whole chart is inside the table, fully legible, with nothing scrolled,
  clipped or overlapping the seats

#### Scenario: The result sits at the centre of the table

- **WHEN** the round is revealed
- **THEN** the chart is horizontally centred and contained within the table with space
  above and below; the reserved reveal-control space may be part of the same group,
  without requiring equal vertical margins around the chart itself

#### Scenario: Revealing does not resize the table

- **WHEN** the round is revealed
- **THEN** the table has exactly the size it had while the round was hidden, and no
  participant's seat has moved

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

### Requirement: The invitation link stands on its own

The table SHALL offer a way to copy the room's invitation link, in a fixed position that does not
move as people arrive or leave. Anybody at the table may use it, not only whoever created the game.

Using it SHALL copy the link and confirm that it did. If the browser refuses clipboard access, the
link SHALL be shown in a form that can be selected and copied by hand, so that the action never
dead-ends.

The control SHALL carry no accompanying explanatory text, and SHALL NOT change according to how many
people are at the table. A button labelled "Invite players" already says what it does; a line of
prose that appears and disappears with the number of participants is noise at the top of every
screen, bought for a moment that lasts seconds.

#### Scenario: Anyone can invite a latecomer

- **WHEN** a meeting is under way and somebody needs the link
- **THEN** any participant can copy it from the same place it has always been

#### Scenario: The control is the same whether alone or not

- **WHEN** a participant is the only one at the table, and again once others have joined
- **THEN** the invitation control looks the same and sits in the same place both times, with no
  explanatory text beside it in either case

#### Scenario: A refused clipboard still yields the link

- **WHEN** the browser does not permit writing to the clipboard
- **THEN** the link is displayed so it can be selected and copied by hand

### Requirement: Other present participants expose an accessible throw picker

Hovering another present participant's seat SHALL reveal a compact overlay offering exactly
Paper ball, Paper plane and Flower. Moving the pointer away from both seat and picker SHALL hide
the overlay again. The mouse interface SHALL NOT show a separate persistent trigger, pin or toggle
button beneath the three choices. Touch activation of that seat's throw control SHALL open the
same picker. Standard keyboard navigation and button activation SHALL remain available as an
accessibility path, without custom shortcuts or global throw key bindings. Each choice SHALL have
an accessible name; the trigger SHALL identify its target and expose its expanded state. Keyboard
focus SHALL be visibly indicated when navigating by keyboard, but keyboard hints, shortcut labels,
extra keyboard controls and keyboard-only focus decoration SHALL NOT appear during normal
mouse/touch use.

Moving from the seat to its picker SHALL keep the picker usable. Escape and activating outside
the picker SHALL close it; closing by keyboard SHALL return focus to its trigger. The picker
SHALL NOT alter seat geometry, obscure essential controls, or make full participant names
unreadable. The viewer's own seat and away seats SHALL not offer throws. Throw controls SHALL
be unavailable while disconnected or before a fresh connection has confirmed the viewer's seat.

#### Scenario: A pointer can reach every object

- **WHEN** a participant hovers another present seat and moves into its picker
- **THEN** the overlay remains open and each of the three objects can be selected

#### Scenario: Pointer departure leaves no persistent control

- **WHEN** a mouse user moves away from another participant and its picker
- **THEN** the picker and its activation control disappear without leaving a pin, toggle or fixed
  throw button visible

#### Scenario: Keyboard and touch provide the same choices

- **WHEN** a participant uses the keyboard or taps the throw trigger on a narrow touchscreen
- **THEN** the same three named choices are available and can be activated without hover

#### Scenario: Dismissal preserves keyboard orientation

- **WHEN** a keyboard user dismisses an open picker with Escape
- **THEN** the picker closes and focus returns to the trigger for that participant

#### Scenario: Accessible keyboard use adds no pointer-mode interface

- **WHEN** a participant operates the picker with a mouse or touchscreen
- **THEN** no keyboard instructions, shortcut labels or extra keyboard controls are displayed
- **AND** switching to standard keyboard navigation reveals focus and permits the same actions
  without a custom throw shortcut

#### Scenario: A changed target or connection is no longer actionable

- **WHEN** an open picker's target becomes away, or the viewer loses their connection or seat
- **THEN** the picker closes and cannot submit a throw until the relevant conditions are valid

### Requirement: Thrown objects fly in from offscreen and settle at the target seat

The three objects SHALL be recognisable, locally served SVG illustrations that fit the existing
interface. The paper ball SHALL read as irregular crumpled paper rather than a faceted polygon.
The flower SHALL be one loose flower rather than a bouquet, with seed-selected variation in flower
shape and colour. Every received throw SHALL enter fully from outside the visible left or right edge,
chosen randomly per event, and fly toward the target's actual displayed seat. It SHALL finish
just in front of/below the seat without hiding its name, card or controls, remain there for
2 seconds, and fade out over 0.6 seconds before removal.

Flight SHALL take approximately 0.6–1.2 seconds with bounded variation in duration, trajectory,
rotation and impact offset. Impact positions SHALL vary visibly across a bounded area in front
of/below the target instead of converging on one apparent magnetic point. Motion SHALL convey
gravity, momentum and a soft landing: the paper ball tumbles, bounces and rolls onward briefly;
the paper plane glides nose-first and skids; and the single flower rotates gently, lands softly and
slides a shorter distance. Each object SHALL decelerate along a seed-derived continuation of its
incoming direction before reaching a distinct final resting position. The flight interval includes
this settling and slide. Repeated throws SHALL not all follow the same path, speed, impact point or
resting point. Objects SHALL not collide with other objects or move participants, cards or the table.

Animation SHALL use the local layout on both the wide table and narrow participant list.
Scrolling, resizing and participants changing position SHALL not leave resting objects attached
to the wrong seat. Objects SHALL not intercept input or introduce page scrollbars. A target
outside the viewport SHALL not cause automatic scrolling or a misplaced visible effect.

#### Scenario: A throw begins outside the visible page

- **WHEN** a throw starts from either selected side
- **THEN** the entire object initially lies beyond that viewport edge before flying into view

#### Scenario: Repeated throws vary without missing their target

- **WHEN** the same object is thrown repeatedly at one participant
- **THEN** flights show differing paths and speeds within the stated interval, hit visibly varied
  points, and slide or roll a varied short distance before resting in front of/below that seat

#### Scenario: Landed objects disappear on schedule

- **WHEN** an object has settled
- **THEN** it rests for 2 seconds, fades over 0.6 seconds, and is removed without moving the layout

#### Scenario: A changing layout keeps the target meaningful

- **WHEN** the viewport changes between the wide table and narrow list, scrolls, or seats move
- **THEN** retained resting objects follow the correct target, an obsolete flight is safely
  retargeted or discarded, and no effect appears at a stale seat position

#### Scenario: Game controls remain usable under repeated throws

- **WHEN** several objects are flying or resting near one participant
- **THEN** the name and card remain readable, estimation controls still receive input, and the
  page gains no scrollbars from objects entering offscreen

### Requirement: Reduced motion and inactive pages do not replay flights

When reduced motion is requested, a throw SHALL appear directly at its final resting position,
without travel, bounce, slide or rotation, and then use the same rest and fade durations. Switching the preference
while objects are moving SHALL stop their motion. Object age SHALL follow elapsed time rather
than frame count, so returning to a backgrounded tab SHALL not replay expired animations.

#### Scenario: Reduced motion keeps the reaction visible

- **WHEN** a throw arrives with reduced motion enabled
- **THEN** its object appears directly below the target, rests for 2 seconds and fades over
  0.6 seconds without a flight

#### Scenario: A sleeping tab does not resume old throws

- **WHEN** a page resumes after its objects' lifetimes have elapsed
- **THEN** those objects are removed instead of completing old flights in front of the user
