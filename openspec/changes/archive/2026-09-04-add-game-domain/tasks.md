## 1. Package skeleton and the constraint that protects it

- [x] 1.1 Create `internal/game` with a package comment stating what belongs here and what does not; verify `go build ./...` and `go vet ./...` are clean
- [x] 1.2 Write the test that walks this package's imports and fails if it ever imports `net/http` or the WebSocket library; verify it passes now and that it fails when such an import is added temporarily on purpose — a constraint test that has never been seen to fail is not known to work

## 2. Identifiers

- [x] 2.1 Implement identifier generation taking an `io.Reader` for randomness, consuming at least 128 bits and rendering it in a URL-safe alphabet that omits characters routinely misread for one another; verify with a test that the output contains only alphabet characters and that the entropy consumed meets the minimum
- [x] 2.2 Verify repeatability and independence: a fixed reader yields the same identifier twice, and many identifiers from a cryptographically secure reader are all distinct and show no sequential relationship
- [x] 2.3 Make a short read or an error from the reader fail room and participant creation outright; verify with a test using a reader that errors and one that returns too few bytes, asserting no room is created with a padded or partially random identifier

## 3. The deck

- [x] 3.1 Define the card type and the named deck, with the t-shirt deck listing `XS`, `S`, `M`, `L`, `XL`, `?`, `☕` in that order; verify with a test asserting the name, the exact membership and the order
- [x] 3.2 Implement deck membership lookup and verify that a value outside the deck, such as `XXL` or `13`, is not a valid card
- [x] 3.3 Verify a newly created room carries the t-shirt deck

## 4. Room membership

- [x] 4.1 Implement joining with a name: trim surrounding whitespace, refuse empty and whitespace-only names, refuse names beyond the maximum length, and allow duplicates; verify each case with a test, and record the chosen maximum in the code with the reasoning for it
- [x] 4.2 Implement participant identity: every participant carries an opaque identifier, and every operation addresses a participant by it; verify with a test where two participants share a display name and an operation on one leaves the other untouched
- [x] 4.3 Refuse any operation naming an identifier not seated in the room, with a sentinel error the caller can recognise; verify with a test that the error is identifiable and that no state changed
- [x] 4.4 Implement rejoining: presenting the identifier of an already seated participant reseats them with their seat, name and current vote intact, and never adds a second participant; verify with a test covering a participant who had already voted
- [x] 4.5 Make rejoining with a different name update the displayed name; verify with a test
- [x] 4.6 Implement the away mark: marking away keeps seat, name and vote; marking present clears it and changes nothing else; nothing removes a participant on a timer; verify all three with tests
- [x] 4.7 Implement renaming with the same name rules as joining, leaving identifier, seat and vote untouched, and permitted while the round is revealed; verify with tests including a rename to an invalid name leaving the old name in place
- [x] 4.8 Implement the report of whether anyone is present — true when at least one participant is not away, false when all are away or there are none; verify with tests for all three situations

## 5. Voting within a hidden round

- [x] 5.1 Implement casting a vote while the round is hidden, recording exactly one card per participant; verify with a test
- [x] 5.2 Implement changing a vote: repeated votes replace the previous card, leaving no trace of earlier choices; verify with a test that votes three times and asserts only the last card is held and the earlier two appear nowhere in the room's state
- [x] 5.3 Refuse a card the deck does not contain, with a recognisable sentinel error and no vote recorded; verify with a test
- [x] 5.4 Keep "has not voted" distinct from "voted for `?`"; verify with a test asserting the two states differ

## 6. Hiding, revealing and finality

- [x] 6.1 Implement the redacted view as a separate type that has nowhere to carry a card value while the round is hidden, reporting per participant only whether they have voted; verify with a test
- [x] 6.2 Write the guard test for the most important rule in the product: build a room in which every participant has voted a distinct card, produce the hidden view, render it to text, and assert that no card value from the deck appears anywhere in the output; verify it fails if a vote value is deliberately added to the view
- [x] 6.3 Verify the hidden view carries no tally either, since a count over a small table would let individual votes be inferred
- [x] 6.4 Implement revealing by any participant, permitted whether or not everyone has voted, with non-voters shown as holding no card; verify with tests for both the fully voted and the partly voted round
- [x] 6.5 Make revealing an already revealed round a harmless no-op rather than an error, so two people pressing at once is not a fault; verify with a test asserting the votes are unchanged and no error is returned
- [x] 6.6 Refuse casting or changing a vote once revealed, with a recognisable sentinel error and no change to any recorded vote; verify with tests for both a non-voter voting and a voter changing their card
- [x] 6.7 Verify the revealed view discloses every participant's card

## 7. Results

- [x] 7.1 Implement the per-card tally over a revealed round, ordered by the deck's own order and containing only cards actually played; verify with a test covering two `M`, one `L` and one `☕`
- [x] 7.2 Verify a card nobody played is absent from the tally entirely rather than present with a count of zero
- [x] 7.3 Verify non-voters are reported as holding no card and are counted in no tally, with the counts summing to the number who did vote
- [x] 7.4 Verify no average, median or other arithmetic over the cards is produced anywhere — search the package for such a computation as well as asserting on the results type

## 8. Starting a fresh round

- [x] 8.1 Implement starting a new round: hidden, with every participant holding no card; verify with a test after a fully voted revealed round
- [x] 8.2 Permit starting a new round from a hidden, partly voted round, discarding the votes cast so far; verify with a test
- [x] 8.3 Verify away participants remain seated, remain away, and hold no card in the new round
- [x] 8.4 Verify the previous round's votes are not retained anywhere, since there is no history in this product

## 9. Round completeness

- [x] 9.1 Implement the report of whether every participant who is not away has voted; verify with a test where an away participant has not voted and everyone present has, asserting the round reports complete
- [x] 9.2 Verify the round reports incomplete when at least one present participant has not voted
- [x] 9.3 Verify that marking an away participant present again, when they hold no card, returns the round to incomplete

## 10. Acceptance

- [x] 10.1 Review every error returned by the package and confirm each is a recognisable sentinel rather than free text, so the next change can turn a refusal into a message without matching on strings; verify with a test that each refusal path returns an identifiable error
- [x] 10.2 Confirm no mutex, channel or goroutine exists anywhere in the package, since each room is owned by a single goroutine in the next change and a lock here would signal that the ownership model had been abandoned
- [x] 10.3 Run the full command set from `CLAUDE.md` and confirm each is clean: `go build ./...`, `go test -race ./...`, `go vet ./...`, and `gofmt -l .` with empty output
- [x] 10.4 Confirm the deployed application is unchanged by this slice: `docker compose up --build` still serves the placeholder page, because nothing calls into this package yet
