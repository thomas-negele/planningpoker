## ADDED Requirements

### Requirement: The entry screen shows the application version

The entry screen at the base URL SHALL show the version carried by the running build as `Version major.minor.patch`. The version SHALL be visible only on the entry screen, without adding another action or interrupting the game-start flow.

#### Scenario: Visitor opens the entry screen

- **WHEN** a visitor opens the base URL
- **THEN** the entry screen shows the running build's version

#### Scenario: Visitor opens a room

- **WHEN** a visitor opens a room URL or starts a game from the entry screen
- **THEN** the room screen does not show the version
