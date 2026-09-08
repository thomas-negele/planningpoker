## MODIFIED Requirements

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
