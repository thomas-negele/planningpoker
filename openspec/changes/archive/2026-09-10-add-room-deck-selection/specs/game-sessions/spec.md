## MODIFIED Requirements

### Requirement: Starting a game creates a room and yields an invitation URL

The application SHALL provide a way to create a new game using either the t-shirt or Fibonacci deck.
Creating one produces a room with a fresh unguessable identifier and returns that identifier, from
which the invitation URL is formed. If no deck is supplied, the room SHALL use the t-shirt deck so
existing clients and the default creation path retain their current behavior. An unsupported deck
selection SHALL be refused without creating a room.

Creating a game SHALL NOT seat anybody. The person who creates it takes a seat the same way everyone
else does, by opening the URL and giving a name. There is no host and no creator role anywhere in
this product, so there is nothing for the act of creation to confer.

Each creation SHALL produce a distinct room. Creating a game twice never returns the same room.

#### Scenario: Creating a Fibonacci game returns a usable room

- **WHEN** a visitor starts a new game and selects the Fibonacci deck
- **THEN** a room using Fibonacci is created, its identifier is returned, and opening the
  corresponding URL reaches that room

#### Scenario: Creating a game returns a usable room

- **WHEN** a visitor starts a new game with either supported deck
- **THEN** a room using that deck is created, its identifier is returned, and opening the
  corresponding URL reaches that room

#### Scenario: Omitting the selection keeps the default

- **WHEN** a game is created without a deck selection
- **THEN** the created room uses the t-shirt deck

#### Scenario: An unsupported selection creates nothing

- **WHEN** a game-creation request names a deck other than t-shirt or Fibonacci
- **THEN** the request is refused and no room is created

#### Scenario: Creating a game seats nobody

- **WHEN** a game has just been created and nobody has opened its URL
- **THEN** the room has no participants, and the person who created it holds no privilege that
  anyone else lacks

#### Scenario: Two games are two rooms

- **WHEN** two games are created in succession
- **THEN** they have different identifiers, and a vote or deck setting in one is not visible in the
  other

## ADDED Requirements

### Requirement: An implicitly created room uses the default deck

Opening a valid room URL whose room does not exist SHALL continue to create that room immediately
with the t-shirt deck. The visitor SHALL NOT be diverted to a deck-selection step; after taking a
seat they can change the deck under the same rules as any other room.

#### Scenario: An unused named URL starts with t-shirt sizes

- **WHEN** somebody opens a valid, unused URL such as `/g/team-alpha`
- **THEN** the room is created with the t-shirt deck and the visitor proceeds to its ordinary name
  prompt
