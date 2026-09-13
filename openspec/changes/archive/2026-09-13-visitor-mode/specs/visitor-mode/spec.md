## Purpose

Visitor mode lets a seated person participate in a planning-poker room without casting a vote while retaining the other actions available at the table.

## ADDED Requirements

### Requirement: A person can take a seat as a visitor

The room-join name form SHALL offer a "Visitor mode" checkbox, unchecked by default. An adjacent, accessible information control marked with an "i" SHALL explain briefly that visitors cannot vote. Confirming the form SHALL seat the person under the entered name with the selected mode. Visitor mode SHALL belong to the seat, not to the remembered-name choice.

#### Scenario: Joining as a visitor

- **WHEN** a person checks Visitor mode and confirms a valid name
- **THEN** they take a seat as a visitor and cannot cast a vote

#### Scenario: Joining as a voter

- **WHEN** a person leaves Visitor mode unchecked and confirms a valid name
- **THEN** they take a seat able to vote while the round is hidden

#### Scenario: The explanation is available without a pointer

- **WHEN** a keyboard user focuses the information control beside Visitor mode
- **THEN** they can read a short explanation that voting is unavailable in this mode

### Requirement: A seated person can change visitor mode at any time

The existing dialog used to edit one's own name SHALL also offer the Visitor mode checkbox. Its value SHALL reflect the seat's current mode. The new value SHALL take effect only when Save is used; Cancel SHALL leave the mode unchanged. This dialog and Save SHALL remain available during hidden and revealed rounds, including after the person has voted. The change SHALL be reflected for all participants and persist across reloads and additional connections to the same seat.

#### Scenario: Save changes the mode

- **WHEN** a seated person changes Visitor mode in their name dialog and selects Save
- **THEN** their seat changes mode and everyone receives the updated room state

#### Scenario: Cancel preserves the mode

- **WHEN** a seated person changes the checkbox but cancels the name dialog
- **THEN** their seat's mode and any vote remain unchanged

#### Scenario: Reload preserves the mode

- **WHEN** a visitor reloads and returns to their existing seat
- **THEN** they remain a visitor without joining as another participant

### Requirement: Visitor mode prevents voting but preserves other seated actions

A visitor SHALL NOT be able to cast or change a vote, even through a custom client. They SHALL retain every other action available to a seated participant, including Reveal, New round, changing the deck when otherwise allowed, changing their name, and throwing objects when otherwise allowed.

#### Scenario: A visitor attempts to vote

- **WHEN** a visitor sends a vote for a valid card while the round is hidden
- **THEN** the server refuses the vote with a specific visitor-mode reason and changes no vote

#### Scenario: A visitor controls the round

- **WHEN** a visitor selects Reveal or New round when that action is otherwise available
- **THEN** the action succeeds under the same rules as for another seated participant

### Requirement: Switching modes has round-specific effects on a vote

When a person saves a switch into visitor mode during a hidden round, their current vote, if any, SHALL be deleted permanently. Switching back during that hidden round SHALL NOT restore it, but they SHALL be allowed to cast a new vote. Once a round is revealed, changing visitor mode SHALL NOT alter any revealed card or tally for that round; the changed mode governs future voting. A new round SHALL start with no votes regardless of mode.

#### Scenario: Hidden vote is deleted

- **WHEN** a voter with a hidden vote saves a switch into visitor mode
- **THEN** their vote disappears from the round and all room snapshots, and cannot be recovered by switching back

#### Scenario: Switching back allows a fresh vote

- **WHEN** a visitor switches back to voter while the round remains hidden
- **THEN** they hold no vote until they play a new card

#### Scenario: Revealed result stays final

- **WHEN** a participant switches either mode after the round is revealed
- **THEN** their already revealed card and its contribution to the tally remain unchanged until New round
