## 1. Routing and the entry screen

- [x] 1.1 Add a hand-written client-side router for the two routes — the entry screen and `/g/{roomID}` — using the History API and handling the back button; verify by navigating between them and by pressing back
- [x] 1.2 Build the entry screen offering only "Start a new game", with no way to browse, search for or type in a room; verify by inspection that no other action exists
- [x] 1.3 Make starting a game call the server, then navigate to the new room's URL; verify in a browser that a game is created and the address bar shows the room URL
- [x] 1.4 Verify a hard reload of `/g/{roomID}` still loads the application, which is the single-page fallback the first change built

## 2. The connection layer

- [x] 2.1 Write the module that owns the socket and exposes exactly two things: the last snapshot and the connection state; verify by review that no component opens a socket or keeps its own copy of the room
- [x] 2.2 Send the five intents — take a seat, play a card, reveal, start a new round, rename — as the protocol defines them; verify each against a running server
- [x] 2.3 Translate every refusal code to a sentence, written out exhaustively rather than defaulting, so a code with no sentence is visible during development; verify with a check that every code the server can send has a translation
- [x] 2.4 Reconnect automatically after a drop, with a delay that grows to a cap and carries jitter; record the chosen numbers in the code with their reasoning
- [x] 2.5 Stop retrying permanently on the room-not-found refusal and on its close code, and only on those; verify both that a recoverable drop recovers and that a vanished room stops

## 3. Name entry

- [x] 3.1 Show the name prompt on opening a room URL, seating nobody until it is confirmed; verify from a second browser that an unconfirmed visitor is invisible at the table
- [x] 3.2 Remember the name in a cookie the page writes and reads, pre-filling the field for a returning visitor while still letting them change it; verify by reloading with a stored name
- [x] 3.3 Show the server's refusal for an empty or over-long name, distinguishing the two, and stay at the prompt; verify by submitting both
- [x] 3.4 Show who is already at the table while the visitor is still at the prompt; verify with a room that already has people in it

## 4. The table

- [x] 4.1 Lay out participants around a central table, preferring stylesheet rules where a plain CSS arrangement is natural since it responds to a narrow screen without help; dynamic styles from components are permitted because Svelte compiles them through the object model, which the policy does not govern — what must not appear is a style attribute hand-written into `web/index.html` or arriving through `{@html …}`; verify by searching the HTML shell and the components for both
- [x] 4.2 Show each participant's name, whether they hold a card, and whether they are away; verify with a room containing a voter, a non-voter and an away participant
- [x] 4.3 Make a participant's own seat distinguishable from the others; verify from two browsers that each sees itself marked
- [x] 4.4 Show a face-down card for somebody who has voted while the round is hidden, and nothing for somebody who has not; verify that no card value is present anywhere in the page for a hidden round, including in the DOM
- [x] 4.5 Keep the table readable as people arrive and leave and on a narrow screen; verify at a phone-sized viewport with several participants

## 5. The deck

- [x] 5.1 Render the deck along the bottom edge from what the server sent, in the server's order, never from a list written into the page; verify by inspection that no card is hardcoded in a component
- [x] 5.2 Keep the deck reachable without scrolling the table away; verify at a short viewport
- [x] 5.3 Show the played card as played, and make playing another replace it with no state in which the participant holds both or none; verify by playing three cards in succession
- [x] 5.4 Make every card operable by keyboard, as a real button rather than a clickable element; verify by playing a card using only the keyboard

## 6. Revealing, results and a fresh round

- [x] 6.1 Offer reveal and new-round to every seated participant at all times, never disabled by who created the game or by whether everyone has voted; verify that revealing works from a participant who did not create the room and while somebody has not voted
- [x] 6.2 Show that everyone present has voted, as information only, with no control conditioned on it; verify by review that nothing is disabled on account of it
- [x] 6.3 Show each participant's card at their seat once revealed, with no card for anybody who did not vote, distinct from the unknown card; verify with a round in which somebody abstained and somebody played the unknown card
- [x] 6.4 Show the count per card in the deck's order; verify with a round in which two people played the same card
- [x] 6.5 Verify no average, median or other computed number over the cards appears anywhere in the interface
- [x] 6.6 Make starting a new round clear every card and return the table to the hidden state; verify from two browsers

## 7. Renaming

- [x] 7.1 Let a participant edit their own name from the table, and nobody else's; verify from two browsers that neither can edit the other
- [x] 7.2 Show the same refusals as at the name prompt, in the same words; verify with an empty and an over-long name
- [x] 7.3 Verify renaming leaves the participant's seat and card untouched, including during a revealed round

## 8. The invitation link

- [x] 8.1 Add the invitation control in a fixed position that does not move as people arrive or leave, available to every participant; verify by watching it while a second participant joins
- [x] 8.2 Copy the room URL on use and confirm that it was copied; verify in a browser
- [x] 8.3 Fall back to showing the link in a selectable form when the browser refuses clipboard access, so the action never dead-ends; verify by denying or stubbing clipboard permission
- [x] 8.4 Show the "you are the only one here" prompt while alone and remove it when somebody sits down, without moving the control; verify by joining from a second browser

## 9. Connection state in the interface

- [x] 9.1 Make the connection state visible at all times — established, re-establishing, ended — conveyed by text and not by colour alone, and exposed to assistive technology; verify with a screen reader or by inspecting the accessible name
- [x] 9.2 Stop presenting the table as current while the connection is not established; verify by killing the server with the table open
- [x] 9.3 Make the intent-sending controls unavailable while disconnected, rather than accepting a click that goes nowhere; verify by clicking a card while disconnected
- [x] 9.4 Show the game-has-ended screen with an offer to start a new one when the room is gone; verify by opening a stale link and by restarting the server with the table open

## 10. The interruption promises, exercised for the first time

- [x] 10.1 Verify a dropped connection to a room that still exists recovers with no action from the participant, putting them back in the same seat with the same name and the same vote — this is what the seat cookie and the grace period were built for and no real browser has ever done it
- [x] 10.2 Verify a laptop that sleeps and wakes within the grace period comes back to a current table rather than the one it left
- [x] 10.3 Verify a restarted server ends the game rather than reconnecting forever, and says so
- [x] 10.4 Verify two tabs of the same room are one seat: a vote in one appears in the other, and closing one does not mark the participant away

## 11. Acceptance

- [x] 11.1 Load the finished table in a **production** build with several participants and confirm the browser console reports zero Content-Security-Policy violations — the mechanism has been measured, but measuring it does not remove the need to check the actual result, and a violation here would be invisible in development where no policy is applied
- [x] 11.2 Confirm the built output contains no reference to a foreign host and that the browser contacts one origin only, as `CLAUDE.md` requires be verified rather than assumed
- [x] 11.3 Play a complete game through the container with two browsers: create, invite, both join, vote, reveal, read the results, revote — and record what was seen
- [x] 11.4 Run the full command set from `CLAUDE.md` and confirm each is clean: `go build ./...`, `go test -race ./...`, `go vet ./...`, `gofmt -l .` with empty output, and `npm run check` in `web/`
- [x] 11.5 Confirm no Go package was modified by this change; if one was, stop and say why rather than absorbing a server change silently
