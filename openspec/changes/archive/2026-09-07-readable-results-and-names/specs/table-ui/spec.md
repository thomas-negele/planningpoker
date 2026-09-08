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

The name SHALL be remembered in the browser and the field SHALL arrive pre-filled
for a returning visitor, so that nobody retypes their name every week. They may
change it before confirming.

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

What the table displays SHALL be centred on the table's own middle, vertically as well
as horizontally. Space reserved for a control that is not currently visible MUST NOT
push the displayed content off that centre: a result sitting against the top edge with
a third of the felt above it and two thirds below reads as something that has slipped
rather than something that was placed. Reserving the space is still required — see the
requirement on revealing — so the reservation has to be made in a way that does not
displace what is being shown.

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
- **THEN** the chart is centred on the table's middle, with the space above it equal to
  the space below it, and the reserved room for the reveal control does not shift it
  upwards

#### Scenario: Revealing does not resize the table

- **WHEN** the round is revealed
- **THEN** the table has exactly the size it had while the round was hidden, and no
  participant's seat has moved
