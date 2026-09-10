## MODIFIED Requirements

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

## ADDED Requirements

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
