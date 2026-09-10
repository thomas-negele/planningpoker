## 1. Domain Decks and Room Rules

- [x] 1.1 Add the fixed Fibonacci deck, stable supported-deck names, and domain lookup validation;
  verify deck tests assert the exact card and scale order for both decks and reject unknown names.
- [x] 1.2 Add selected-deck room construction while preserving t-shirt as the no-selection default;
  verify room tests cover explicit Fibonacci, omitted selection, fresh slices, and invalid selection
  without partial room creation.
- [x] 1.3 Implement the seated-participant deck-change action with immediate empty-round changes,
  voted-round refusal, revealed-round replacement of the pending choice, and activation on
  `NewRound`; verify domain tests cover every transition and prove refused changes preserve votes,
  active deck, pending deck, and reveal state.
- [x] 1.4 Expose active and optional pending decks in the room view while keeping revealed results on
  their original active deck; verify view tests cover pending-state visibility and rerun the hidden
  vote disclosure tests for both decks.

## 2. Creation and Live Protocols

- [x] 2.1 Extend manager creation to accept a supported deck while leaving `EnsureRoom` on t-shirt;
  verify manager tests cover selected issued rooms, default named rooms, capacity, and failed
  creation without admission.
- [x] 2.2 Extend `POST /api/games` with the optional deck selection and a client-error response for an
  unsupported non-empty name; verify HTTP tests cover Fibonacci, omitted-body compatibility,
  invalid input, same-origin checks, and no room allocation on refusal.
- [x] 2.3 Add the `setDeck` WebSocket intent, hub command, stable refusal mapping, and optional pending
  deck snapshot field; verify protocol and hub tests cover seated authorization, shared broadcasts,
  locked hidden rounds, latest pending choice, and activation on the next round.

## 3. Entry and Room Interface

- [x] 3.1 Update frontend protocol types, room connection actions, refusal text, and creation client
  to send and receive the new deck state; verify `npm run check` succeeds in `web/`.
- [x] 3.2 Add the two compact, keyboard-operable deck choices above the entry start button with
  t-shirt selected by default and no local persistence; verify both creation choices submit the
  expected deck and the entry screen still offers no room discovery.
- [x] 3.3 Add the icon-only settings control immediately left of `Invite players` and its focused
  two-choice dialog; verify it reflects active/pending state, is available to every seated
  participant, explains the voted-round disabled state to pointer and assistive-technology users,
  and labels post-reveal choices as applying to the next round.
- [x] 3.4 Make only the responsive adjustments required for eleven Fibonacci cards, nine result rows,
  and the adjacent header controls; verify t-shirt and Fibonacci rooms at wide and narrow viewport
  sizes have reachable cards, legible unclipped results, stable table/seats, and unobscured controls.

## 4. Documentation and End-to-End Verification

- [x] 4.1 Update the concise product documentation to name both fixed decks and the t-shirt default;
  verify it does not promise custom decks, roles, persistence, additional settings, or statistics.
- [x] 4.2 Run `go test ./...`, `npm run check`, and `npm run build`, then exercise creation, immediate
  switching, voted-round locking, post-reveal replacement of the pending choice, and next-round
  activation with two connected browsers; verify both clients stay synchronized and the revealed
  results never change decks underneath the completed round.
