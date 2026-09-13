## 1. Room Authority and Throw Admission

- [x] 1.1 Add the three stable object identifiers and transient event data with public sender/target
  IDs and bounded variation input; verify focused Go tests reject unsupported objects and prevent
  client-supplied sender identity, coordinates or markup from becoming authoritative.
- [x] 1.2 Implement room-owned eligibility checks without calling the snapshot-broadcasting action
  helper; verify tests cover unseated senders, self/unknown/away/cross-room targets, present targets
  before and after voting/reveal, and no mutation or extra snapshot from either accepted or refused
  throws.
- [x] 1.3 Add fixed rolling-second quotas of 3 accepted throws per participant and 12 per room with
  bounded timestamp histories; verify injected-clock tests cover exact boundaries, mixed targets,
  combined tabs, reconnect to the same seat, independent participants/rooms, and rejected attempts
  consuming neither quota.
- [x] 1.4 Add nonblocking bounded cosmetic room admission with ordinary-command and shutdown
  priority; verify saturation drops excess requests without blocking socket readers, queued
  requests revalidate current presence/membership, and votes/reveal/new-round commands still run.

## 2. Transient Delivery and Wire Protocol

- [x] 2.1 Add separate bounded per-connection cosmetic delivery and immutable room-local fan-out;
  verify hub tests deliver to sender/target/other tabs only in the correct room, drop effects on
  cosmetic overflow without closing that connection, and keep reliable update capacity independent.
- [x] 2.2 Extend decoding/encoding with `throw`, `thrown`, event IDs/seeds/age, specific private
  validation refusals, and optional `throwPolicy` carrying actual message allowances; verify wire
  tests cover all fields, every refusal mapping, hidden-data exclusion, no snapshot per throw,
  malformed requests and the unchanged raw size/rate gate.
- [x] 2.3 Integrate cosmetic events into the existing single socket writer after initial snapshot
  delivery and behind reliable updates, due heartbeats and closure; verify transport tests cover
  snapshot-first ordering under concurrent throws, bounded queue age, slow writers, retained
  deadlines, persistent raw-flood handling and clean shutdown.
- [x] 2.4 Verify transient lifecycle with real WebSocket fixtures: current healthy clients see
  matching event identities/variation, joining/reconnecting clients receive no history, and
  accepted throws never expose hidden votes or private seat credentials on the network.

## 3. Frontend Connection and Pacing

- [x] 3.1 Extend frontend wire types and the central connection with throw policy, event delivery
  and refusal mappings; verify type checking and protocol consistency checks pass, missing policy
  disables only throwing, and game snapshots/refusals keep their existing behaviour.
- [x] 3.2 Implement local rolling throw admission plus conservative accounting of every sent
  intent, a reserved game-message allowance and outgoing-backlog suppression; add the small
  frontend test command and verify injected-clock tests at general rates 1 and 10 prove excess
  clicks are dropped, game intents are immediate, reconnect starts conservatively, and no retries
  or delayed throw backlog appear.
- [x] 3.3 Add disposable event subscriptions and bound client effect admission; verify lifecycle
  tests cover server-only animation, missing/stale events, the 48-object ceiling, fresh-seat
  gating, disconnect, room navigation and absence of replay after reconnect.

## 4. Picker and SVG Motion

- [x] 4.1 Refine the shared local SVG artwork into an irregular crumpled paper ball, the retained
  paper plane and a single seed-varied flower; verify all three remain distinct at their actual
  displayed size and fit the existing light/dark palettes without external assets or CSP changes.
- [x] 4.2 Refine the compact picker so fine-pointer hover exposes only the three choices and leaves
  no visible pin/toggle after departure, while touch activation remains available; verify pointer
  movement into and away from the overlay, all choices, outside dismissal, stale target/connection
  closure and unchanged seat geometry on wide/narrow screens with maximum-length names. Keep a
  browser interaction regression test for pointer departure and narrow touch activation.
- [x] 4.3 Preserve standard keyboard/button access as the discreet accessibility path; verify
  target/object accessible names, Tab and native activation, focus-visible styling, scoped Escape
  dismissal and focus return, with no custom shortcuts, global throw bindings, keyboard hints or
  extra visible controls during mouse/touch use.
- [x] 4.4 Extend the pure seeded trajectory and lifecycle functions with visibly varied impact and
  final-rest offsets plus object-specific decelerating roll/slide, retaining offscreen entry,
  0.6–1.2 second total flight/settling, 2 second rest and 0.6 second fade; verify deterministic tests
  cover both viewport edges, finite geometry, bounded distinct endpoints, momentum continuation,
  differing seeds, plane orientation and elapsed-time expiry.
- [x] 4.5 Update the single viewport-clipped SVG animation layer and target identification; verify
  objects never capture pointer input or create scrollbars, resting objects track the correct seat
  on scroll/reflow, obsolete flights are dropped, and expired/disconnected objects release their
  frame callbacks and subscriptions.
- [x] 4.6 Respect reduced motion and inactive-page timing; verify direct placement at the final rest
  position without travel, rotation, bounce or slide, correct rest/fade, preference changes during
  flight, and removal of expired objects on resuming a backgrounded page.

## 5. Documentation and Integrated Verification

- [x] 5.1 Document the three throws, mouse/touch interaction and fixed 3/12 ceilings alongside the
  general message limit; verify README and relevant configuration comments explain silent cosmetic
  drops, reduced effective rates under low message settings, and no new environment variables or
  settings. Document the frontend test command in contributing instructions.
- [x] 5.2 Run the repository's required checks: in `web`, `npm ci`, the frontend tests,
  `npm run check`, `npm run build` and the focused browser interaction test; at the root,
  `go build ./...`, `go vet ./...`,
  `gofmt -l .`, `go test -race -count=3 ./...` and
  `go test -race -count=3 -tags embedassets ./...`; record actual results and any unavailable check.
- [x] 5.3 Build/run the Docker deployment and inspect its production browser behaviour with
  multiple independent participants plus a same-seat extra tab; verify all objects from both
  edges, both layouts, long names, picker interaction, discreet keyboard accessibility, reduced
  motion, reflow, reconnect and ordinary estimation actions. Confirm no CSP violations and no
  foreign-origin requests or external runtime references in the built output.
- [x] 5.4 Exercise sustained multi-client throws alongside votes/reveal/new rounds, including
  room quota saturation and a slow recipient, at default and low message rates; verify bounded
  pending/live object counts, responsive game updates, quiet suppression of ordinary excess
  clicks and cleanup when activity stops. Record observed results without claiming unmeasured
  hardware throughput or general denial-of-service resistance.
- [x] 5.5 Validate the finished change against every delta scenario and run
  `openspec validate add-participant-throws --strict`; record implementation verification results
  and leave tasks incomplete for anything that has not actually been checked.
