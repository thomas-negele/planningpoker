## MODIFIED Requirements

### Requirement: The table shows everyone and what they have done

The table SHALL show every participant, each with their name. For each participant it SHALL show
whether they have played a card, whether they are a visitor, and whether they are away, and it SHALL be apparent which one is you.

A name of any length the rules permit SHALL be displayed in full, at every screen width. It MUST NOT
be clipped, shortened, replaced by an ellipsis or otherwise altered to fit the space available.
Truncation is not an acceptable outcome for a name the product itself accepted: it presents somebody
under a name they did not choose, with nothing on screen to explain why. Where a name and its
surroundings do not fit, it is the surroundings that give way.

The same rule applies to the field in which a participant edits their own name: it SHALL hold a name
of the full permitted length visibly, so that nobody edits a name they cannot entirely see.

On a wide screen the participants SHALL be arranged around the table. Each one SHALL sit clear of
the table's edge, measured from the middle of the whole seat — the card or visitor marker and the name together —
rather than from the card alone, so that no card or visitor marker ends up lying on the table. This clearance
SHALL be derived from the seat's actual extent, which includes whatever width a full-length name
requires, so that a wider seat does not reintroduce the overlap.

On a narrow screen they SHALL instead be a list below the table, one participant per row. A ring
collides with itself at that width, and a row of full-size cards costs more vertical space than a
phone has to give before the table has even begun. The table itself carries the result and the
action, so it comes first and the list of who is present follows it.

For a non-visitor, the card SHALL remain a card in both arrangements: face down for somebody who has played, face up once revealed, and visibly empty for somebody who has not played. A visitor who holds no revealed card for the current round SHALL instead have a compact text label "Visitor" in place of the card slot. If a card was revealed before that person switched to visitor mode, that revealed card SHALL remain visible for this round alongside the visitor label.

While the round is hidden, a non-visitor who has voted SHALL be shown as holding a face-down card,
and one who has not SHALL be shown as holding nothing. A visitor SHALL show no card symbol. The value MUST NOT be shown, or discoverable,
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
  card or visitor marker and whether they are away

#### Scenario: A visitor keeps a seat without a card symbol

- **WHEN** a visitor has no revealed card for the current round on a wide or narrow screen
- **THEN** their name and a "Visitor" text label remain visible, with no card symbol

#### Scenario: A revealed card remains after a mode change

- **WHEN** a participant with a revealed card switches to visitor mode
- **THEN** their revealed card and visitor label both remain visible for that round

### Requirement: The deck is always to hand

For a non-visitor, the cards the room's deck offers SHALL be shown along the bottom edge of the screen,
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

Cards SHALL be operable by keyboard as well as by pointer. For a visitor, the voting deck SHALL not be shown as an available action; its place SHALL carry a short "Visitor mode" label, so that the space the deck occupies does not simply fall empty.

#### Scenario: The deck comes from the server

- **WHEN** the table is shown to a non-visitor
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

#### Scenario: Visitor has no voting deck

- **WHEN** a visitor views the table during a hidden round
- **THEN** the voting cards are not offered to them, and a short "Visitor mode" label stands where the deck would be

## ADDED Requirements

### Requirement: The table identifies a round without present voters

When no present participant is eligible to vote, the round status SHALL say "No voters this round" instead of waiting for votes. Reveal and New round SHALL remain available under their usual rules.

#### Scenario: Only visitors are present

- **WHEN** every present participant is a visitor and the round is hidden
- **THEN** the status says "No voters this round" and Reveal remains available
