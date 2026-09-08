## MODIFIED Requirements

### Requirement: The table shows everyone and what they have done

The table SHALL show every participant, each with their name. For each participant it SHALL show
whether they have played a card and whether they are away, and it SHALL be apparent which one is you.

On a wide screen the participants SHALL be arranged around the table. Each one SHALL sit clear of
the table's edge, measured from the middle of the whole seat — the card and the name together —
rather than from the card alone, so that nobody's card ends up lying on the table.

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

#### Scenario: No card lies on the table

- **WHEN** participants are arranged around the table on a wide screen, at any position including
  the diagonals
- **THEN** no participant's card overlaps the table

#### Scenario: A narrow screen lists the participants below the table

- **WHEN** the table is shown on a screen too narrow for a ring
- **THEN** the participants appear as a list below the table, one per row, each still showing their
  card and whether they are away

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

## ADDED Requirements

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

## REMOVED Requirements

### Requirement: The invitation link is always available in one place

**Reason**: Replaced by "The invitation link stands on its own". Everything it required about the
control itself is carried over unchanged — fixed position, available to anybody, a working fallback
when the clipboard is refused. What is dropped is the sentence it required while a participant was
alone, saying so and pointing at the control.

That sentence was a reasonable idea that did not survive being looked at. It sat at the top of the
screen, wrapped onto two lines on a phone, and pushed everything below it down until somebody else
joined — all to explain a button that already says "Invite players".

**Migration**: None. No stored data and no protocol message is involved; the sentence is simply no
longer rendered.
