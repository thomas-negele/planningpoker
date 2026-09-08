## 1. What an identifier may be

- [x] 1.1 Add a validator next to the identifier generator in `internal/game`, so the two cannot drift apart; verify a generated identifier always passes
- [x] 1.2 Widen it to admit chosen names: between 5 and 64 characters, letters, digits, hyphens and underscores only; verify with tests that `team-alpha`, `Standup_2`, a 5-character name and a 64-character name all pass, and that an empty name, `abc`, a 4-character name, 65 characters, a space, a slash, a dot and a percent sign all fail
- [x] 1.2b Record beside the lower bound why it exists — that `a`, `test` and `abc` are the names several teams reach for at once, so a floor keeps a name something somebody meant rather than something they pressed
- [x] 1.3 Match identifiers exactly, including case, and normalise nothing; verify with a test that `/g/Team-Alpha` and `/g/team-alpha` produce two separate rooms and that each reports the spelling it was asked for
- [x] 1.4 Verify an issued identifier still passes unchanged, so the widening did not break the kind that protects a room

## 2. Creating a room where it is looked for

- [x] 2.1 Add an operation to the manager that ensures a room exists at a given identifier, returning the existing one if there is any
- [x] 2.2 Refuse an identifier the validator does not accept, without creating anything
- [x] 2.3 Do the check and the insertion under one lock; verify with a test that many goroutines asking at once all end up with the same room, under `-race`
- [x] 2.4 Verify a created room is empty — no participants, no votes, a hidden round — and that its grace period starts fresh
- [x] 2.5 Verify asking for a room that is alive and occupied returns that room and disturbs nobody in it
- [x] 2.6 Verify the room map is keyed by the identifier exactly as written, so two spellings are two rooms

## 3. The socket creates the room

- [x] 3.1 Make the WebSocket handler create the room when it does not find one, instead of closing with "room not found"; verify with a test that connecting to an identifier nobody has used yields a working room
- [x] 3.2 Keep the refusal for an identifier the validator rejects, closing with a code the client can tell apart from any other failure; verify with a test
- [x] 3.3 Delete the reopen route and its handler; verify no reference to it remains anywhere in the repository
- [x] 3.4 Verify link-preview bots cannot create rooms: fetching the page over HTTP without opening a socket must leave the room count unchanged
- [x] 3.5 Verify the seat cookie is still issued for a room that has just been created

## 4. Losing a round, and saying so

- [x] 4.1 Have the connection notice when it held a seat, reconnected, and does not appear in the room it reached; verify this is derived from the snapshot rather than from a close code
- [x] 4.2 Show a brief dismissible announcement for exactly that case, saying the round was lost and a seat must be taken again; verify in a browser
- [x] 4.3 Verify an ordinary reconnection with the seat intact announces nothing
- [x] 4.4 Verify a first visit to a freshly created room announces nothing
- [x] 4.5 Verify the announcement blocks nothing: the table, the deck and every control stay usable while it is showing

## 5. The screens

- [x] 5.1 Remove the ended screen's role as the answer to a missing room; it is reached only by an identifier the server refuses, and says what a room name must look like — at least five characters, letters, digits, hyphens and underscores — rather than only that the link is wrong
- [x] 5.2 Remove the reopen button and the two paragraphs of explanation from that screen, leaving the heading and one action
- [x] 5.3 Verify the interface nowhere distinguishes a room reached by a chosen name from one reached by an issued identifier — no badge, no warning, no different wording
- [x] 5.4 Replace the entry screen's subtitle with a line that says what is about to happen

## 6. The case this change exists for

- [x] 6.1 With three browsers in one room, restart the server, and verify all three end up back in one room at the same URL with no clicks at all
- [x] 6.2 Verify each of the three is told the round was lost, and that the message can be dismissed
- [x] 6.3 Verify opening `/g/team-alpha`, a name nobody has used, produces a room there
- [x] 6.4 Verify two browsers opening the same unused name at the same moment end up in one room
- [x] 6.5 Verify a URL the server refuses says it is not a game and creates nothing
- [x] 6.6 Verify opening `/g/abc` loads the page, states the five-character rule, and creates no room

## 7. The interface adjustments made while testing

- [x] 7.1 Reveal is hidden once the round is revealed, and nothing else moves when it goes
- [x] 7.2 New round is offered at all times and sits outside the table
- [x] 7.3 No participant's card overlaps the table on a wide screen, at any position including the diagonals
- [x] 7.4 A narrow screen lists the participants below the table, one per row
- [x] 7.5 The sentence shown while somebody is alone is gone, along with the plumbing that told the invitation control how many people were present
- [x] 7.6 On a narrow screen the connection status and the invitation control sit side by side with New round centred beneath, and the header does not move when somebody joins

## 8. Acceptance

- [x] 8.1 Run the full command set from `CLAUDE.md` and confirm each is clean: `go build ./...`, `go test -race ./...`, `go vet ./...`, `gofmt -l .` with empty output, and `npm run check` in `web/`
- [x] 8.2 Confirm the container still builds and serves, with zero Content-Security-Policy violations in a production build
- [x] 8.3 Confirm what `internal/game` gained, and that no rule about seating, voting, revealing or results changed — it gained `ValidRoomID` with the two length bounds, the `ErrInvalidRoomID` sentinel, and `NewRoomAt`, which lets a room be created under an identifier the caller holds. `NewRoomAt` was not in the original plan and is recorded rather than absorbed: it is needed because a room's identifier travels in every snapshot, so a room has to actually be at the address it was reached by. No rule about seating, voting, revealing or results changed
- [x] 8.4 Rename `Manager.Reopen` to `EnsureRoom`, since nothing is reopened any more — the method makes sure a room exists where somebody looked for it; verify no reference to reopening survives anywhere in the repository
