## MODIFIED Requirements

### Requirement: A dropped connection marks a participant away rather than removing them

When a participant's connection is lost, they SHALL be marked as away. They keep their seat, their
name and their vote, and they remain visible to everyone else at the table, distinguishably marked
as away. When they return, the away mark is cleared.

A connection that is lost **without being closed** counts as lost in the same way. A silently dead
connection — one whose network path has gone while the socket still appears open — SHALL be
recognised within a bounded time and its participant marked away like any other. Before, such a
participant stayed shown as present indefinitely, which is worse than showing them as away: the
table looked as though somebody was there to vote.

A participant is removed from a room only when the room itself ceases to exist. Nothing in these
rules removes a participant on a timer, and the recognition above is emphatically not such a rule —
it measures whether the network path is alive, never whether anybody has done anything.

#### Scenario: Going away keeps the seat and the vote

- **WHEN** a participant who has voted is marked away
- **THEN** they remain at the table, their vote remains recorded, and everyone else sees them
  marked as away

#### Scenario: Returning clears the away mark

- **WHEN** a participant who is marked away is marked present again
- **THEN** the away mark is cleared and nothing else about them changes

#### Scenario: Nobody is removed by the passage of time

- **WHEN** a participant has been away for any length of time
- **THEN** they are still seated at the table, because no rule removes a participant on a timer

#### Scenario: A connection that dies without closing is recognised

- **WHEN** a participant's connection stops answering without the socket being closed
- **THEN** within a bounded time they are marked away and the table stops showing them as present

#### Scenario: Thinking for a long time is not going away

- **WHEN** a participant stays connected but sends nothing at all for far longer than any interval
  in the system
- **THEN** they remain present and seated, because nothing measures how long it has been since
  somebody last did something
