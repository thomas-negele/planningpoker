# Implementation verification

Verified on 2026-10-02/03 from `add-more-throw-objects`.

## Automated checks

- `cd web && npm ci`: passed.
- `cd web && npm test`: passed, 24 tests. New tests cover the known-object check, the policy's
  object list (order, unknown entries, unusable lists), the seven heart colours and their
  determinism, normal size for the existing objects and for every resting, fading and
  reduced-motion pose, the pile of poo's squash, wobble and short slide, the heart's single pulse,
  and seed variation for both new objects. The pre-existing motion tests passed unchanged.
- `cd web && npm run check`: passed with 0 errors and 0 warnings.
- `cd web && npm run build`: passed.
- `go build ./...`, `go vet ./...`, `gofmt -l .` and `git diff --check`: passed with no output.
- `go test -race -count=3 ./...` and `go test -race -count=3 -tags embedassets ./...`: passed.
  New Go tests cover the switch (unset, empty, `true`, `false`; refusal naming variable and value
  for `yes`, `1`, `TRUE`, ` true`, `on`, `False`), the accepted-object list with the switch on,
  off and through `NewManager`'s clamping of invalid limits, both new objects before voting, after
  voting and after reveal, the unseated/self/unknown/away refusals for both, the shared
  per-participant allowance, the unknown-object refusal for `poo` while switched off, the throw
  policy listing four or five objects, and unchanged fan-out of both objects to every connection.
- `docker compose config`: resolves `PLANNINGPOKER_POO_THROWS` to `false` without an `.env` file
  and to `true` with one beside a copy of `compose.yaml`.
- `cd e2e && npm run check` and `cd e2e && npm test`: passed, 29 tests, now against two
  production-built servers, one per switch setting. The picker tests assert exactly four named
  choices with the switch off and five with it on, by hover and on a 320 px touch layout, a
  picker inside the 320 px viewport, and keyboard activation of the heart and the pile of poo with
  focus returning to the trigger and the object appearing and expiring.
- `openspec validate add-more-throw-objects --strict`: passed.

## Production inspection

- Artwork was rendered at picker size (24–25 px) and flight size (36–40 px) on the picker
  surface and the table surface. The pile of poo reads as the familiar emoji at both sizes. The
  interface has only a dark colour scheme; there is no light mode to check.
- Contrast, measured as WCAG ratios against the picker surface `#171a20` and the lightest table
  surface `#1a1e27`: heart fills 4.41–11.53, heart edges 3.20–3.34; the pile of poo's fill 3.95
  and outline 3.21. The first drawing failed this check — dark edges at 1.5–2.8 and the brown
  fill at 2.95 — so edges were lightened within their hue and the brown raised; dark seams inside
  the shape now carry the tiers.
- Picker bounds with eight participants and 15-character names: at 1280 px and on the 320 px list
  every picker lies inside the viewport. At 928 px, the narrowest table layout, the rightmost
  seat's picker reached x = 1026 and was clipped; pickers in the right third of the table now end
  at the seat's right edge and open towards the centre, after which all seats measured inside.
  The target seat's name and card and the estimation controls stay visible. On the narrow list
  the open picker overlaps the start of a long name in the row above while it is open; the
  three-choice picker already did so slightly.
- Per-frame transforms from a real throw: the pile of poo hit at scale 1.21 × 0.74, overshot to
  0.90 × 1.12, was back at 1 × 1 about 280 ms later and slid about 3 px; the heart swelled once
  to 1.25 and rested at 1. With reduced motion both appeared at rest with no rotation or scaling;
  switching to reduced motion mid-flight left each object with a single constant transform.
- The production build's `http://`/`https://` strings are identical before and after this change
  (XML namespaces, licence and Svelte diagnostic texts); no request to another host was added. No
  console errors or Content-Security-Policy violations appeared during any browser run.
- `docker compose -f compose.yaml -f compose.local.yaml up --build -d` built and started the
  image twice. Without `.env` the container ran with the switch off and offered four objects; with
  `PLANNINGPOKER_POO_THROWS=true` in a temporary `.env` it offered five. For every offered object
  two browsers in one room saw the same object, entry side and heart colour. Both runs returned
  the unchanged Content-Security-Policy and the entry screen showed `Version 1.1.0`. The
  temporary `.env` was removed afterwards; the container is still running with the switch on.
- `web/package.json` and `web/package-lock.json` both carry 1.1.0.

## Review and `openspec verify`

`/opsx:verify` found no critical issue: 20/20 tasks, all requirements implemented. It raised a
missing test for the shared room allowance, and three places where design.md described an
earlier plan rather than the code. A code review raised ten findings, each checked before acting:

- Not confirmed: that a page reads the object list only once. Every new socket resets the throw
  policy (`web/src/lib/connection.svelte.ts`), so a reconnect after a restart with a changed
  switch reads the new list. Also not confirmed: that pickers overflow at other table angles.
  Measured at 928 and 1280 px with 3, 5, 7, 10, 12, 15 and 19 participants, no picker crossed
  the viewport.
- Fixed: an empty picker bubble on hover before a usable object list arrives (the throw control
  now renders only with something to offer); the hub checking a second hard-coded list (rooms now
  check `Limits.ThrowObjects()`, the list the browser is offered); the new transport test
  comparing per-writer event ages; the object list being rebuilt per connection; the object type
  duplicated beside its array (now derived from it); per-object radius and arc as conditionals
  (now type-checked tables); and the squash anchor (0.44 of the icon size, matching the drawing).
- Added: `TestNewObjectsShareTheRoomAllowance`. design.md now describes `Limits.PooThrows`, the
  dark-only palette and the actual picker rule.
- Left as is: the browser suite compiling two servers in parallel, which shares Go's build cache
  and took 13.8–14.5 s against a 60 s limit.

## Final cleanup

A last pass over the whole diff, for a codebase that is meant to be shown:

- The switch is parsed like its neighbours (only a non-empty value is examined) and refuses with
  `errors.New`; `hub.Limits` documents its new field.
- The object-name conversion in the transport layer is a named helper; one `sameThrow` helper
  compares deliveries of an event without their per-writer age, now also in the pre-existing
  fan-out test that compared ages directly.
- Per-object turns joined the other motion tables, so every per-object motion value is a
  `Record<ThrowObject, …>` the type checker completes; long lines were wrapped.
- The browser tests take their two server addresses from `e2e/servers.ts` instead of importing
  the Playwright configuration.
- README wrapping and the proposal's impact list were brought in line with the result.

The cleanup changed no requirement. After it the whole verification order, the browser suite
(29 tests) and the Docker check with the switch on were run again and passed.

## Landing on the narrow list

The owner's manual test found that on the narrow list objects of every kind came to rest in the
row below their target. The table-ui delta, proposal, design and tasks were revised (group 8)
before the code changed.

- The new Playwright test failed first for the right reason (object centre at y = 437, target row
  ending at y = 417) and passes since the fix.
- `npm test`: 27 tests. New tests: every object rests on the row's middle inside the free space for
  200 seeds, a row narrower than the object centres it, and resting objects are refitted when the
  layout switches. The wide-table motion tests passed unchanged.
- At 320 px with the names `Al` and `Maximilianeeeee`, a heart, a pile of poo and a paper ball
  rested in the target's row between the end of the name (x = 48 or 150) and the controls
  (x = 237), covering neither. Switching to 1280 px moved them below the seat; switching back
  returned them to the row.
- Playwright's touch emulation placed a tap about 55 px below the picker entry it targeted, so
  the new test uses the narrow window with a mouse, as in the owner's report.
- The full verification order, the browser suite (30 tests) and the Docker check with the switch
  on passed again.

## Not run

- No check against a light colour scheme, because the interface has none.
- No check on a physical touch device; touch was emulated in Chromium.
