## 1. Operator Switch

- [x] 1.1 Read `PLANNINGPOKER_POO_THROWS` in `cmd/planningpoker/config.go` as a boolean field with a
  full-sentence comment, accepting exactly `true` and `false` and treating unset or empty as
  `false`; verify config tests cover unset, empty, `true`, `false`, and refusal naming variable and
  value for `yes`, `1`, `TRUE` and ` true`.
- [x] 1.2 Add the switch to `compose.yaml` as `"${PLANNINGPOKER_POO_THROWS:-false}"` with an
  explanation in full sentences of what it does, that the other objects do not depend on it, that
  an `.env` file can enable it and that a change needs a restart; verify `docker compose config`
  resolves it to `false` without an `.env` file and to `true` with one.
- [x] 1.3 Describe the switch in `README.md` beside the other deployment settings and name all
  throwable objects in the throw sentence, noting that the pile of poo is off by default; verify the
  documented variable name and values match the code.

## 2. Server Accepts the New Objects

- [x] 2.1 Add `heart` and `poo` to the hub's throw objects in `internal/hub/throws.go` and replace
  the fixed check with the set of accepted objects passed in from `main`, which contains `poo` only
  when the switch is on; verify Go tests show the heart, and with the switch on the pile of poo, are
  accepted before voting, after the target votes and after reveal, are refused for self, away and
  unknown targets and unseated senders exactly like the existing objects, and share the existing
  3-per-participant and 12-per-room allowances; and that with the switch off `poo` receives the
  unknown-object refusal without mutation or broadcast.
- [x] 2.2 Add the list of accepted objects to the throw policy in `internal/transport` and extend
  the protocol tests: the first snapshot's policy lists four objects with the switch off and five
  with it on, a `thrown` event carries each new object unchanged to every connection in the room,
  and an unknown object still yields the existing refusal; verify `go test -race ./...`,
  `go vet ./...` and `gofmt -l .` pass with no findings.

## 3. Frontend Types, Connection and Picker

- [x] 3.1 Add `heart` and `poo` to `ThrowObject` and the policy's object list to `ThrowPolicy` in
  `web/src/lib/protocol.ts`, and to the list of known objects the connection accepts in
  `web/src/lib/connection.svelte.ts`; verify `npm run check` reports no errors, a `thrown` event
  with either object reaches the effect layer, an unknown object is still discarded, and a policy
  without a usable object list leaves throwing disabled.
- [x] 3.2 Build the picker in `Seat.svelte` from the policy's object list, in the order Paper ball,
  Paper plane, Flower, Heart, Pile of poo, with the accessible names "Throw a heart at ‹name›" and
  "Throw a pile of poo at ‹name›"; verify in the production build with the switch off that four
  choices appear with no gap, with it on that five appear, and that every offered choice can be
  selected by hover, touch and keyboard and appears at the target when thrown.
- [x] 3.3 Check the five-choice picker against the viewport for the outermost left and right seats
  of the wide table and for the narrow list at 320 pixels with a maximum-length name; if any picker
  crosses the viewport edge, open it towards the table's centre for those seats. Verify by
  screenshot that no picker is clipped and that names, cards and estimation controls stay readable
  and usable, and record the outcome in the verification notes.
- [x] 3.4 Update the Playwright suite in `e2e/tests/throw-picker.spec.ts` from three choices to
  four with the switch off and five with it on, including their accessible names; verify the tests
  pass against production-built servers started with each setting.

## 4. Artwork and Heart Colours

- [x] 4.1 Draw the pile of poo in `ThrowIcon.svelte` from the description in the design — three
  narrowing brown tiers, a tip curling to one side, darker edge, light highlights, two eyes and a
  wide grin — without tracing any vendor's emoji artwork; verify at the picker size and the flight
  size, in light and dark mode, that it is recognisable as the familiar emoji and fits the other
  icons.
- [x] 4.2 Draw the heart with seven seed-selected palettes (red, pink, orange, yellow, green, blue,
  purple; `seed % 7`), each with fill, highlight and darker edge, and a fixed red heart in the
  picker; verify a unit test or rendered check that seeds map to all seven palettes and only to
  them, and that the same seed always yields the same colour.
- [x] 4.3 Measure the contrast of the picker icons' edges against the picker surface in light and
  dark mode, and of the yellow and other light hearts' edges against the table surface; verify each
  meaningful edge reaches at least 3:1 and record the measured ratios.
- [x] 4.4 Set the in-flight display sizes of both objects in `ThrowLayer.svelte`; verify the
  production build contains no `http://` or `https://` reference introduced by this change and the
  browser shows no Content-Security-Policy violation while both objects fly.

## 5. Motion

- [x] 5.1 Extend `Pose` in `web/src/lib/throw-effects.ts` with `scaleX` and `scaleY` and apply them
  in `ThrowLayer.svelte`, anchored at the base for the pile of poo and at the centre for the heart;
  verify the existing unit tests still pass unchanged and every existing object, and every resting,
  fading and reduced-motion pose, has a scale of exactly 1.
- [x] 5.2 Add the pile of poo's motion parameters — flatter arc, slight sway, short plop, slide of
  a few pixels — and its squash-and-wobble during settling; verify unit tests show the squash occurs
  after impact, has fully decayed before rest, and that its slide is shorter than the paper ball's
  for the same seeds.
- [x] 5.3 Add the heart's motion parameters — lighter arc, gentle rotation, soft landing, short
  slide — and its single pulse during settling; verify unit tests show the scale passes through one
  maximum and is back at 1 before rest, and that repeated seeds still vary path, speed and resting
  point within 0.6–1.2 seconds.
- [x] 5.4 Tune the final values in the browser and confirm the feel described in the specification
  on both the wide table and the narrow list; verify with reduced motion enabled that both objects
  appear at rest at normal size without squash, wobble or pulse, and that switching the preference
  mid-flight stops their motion.

## 6. Version

- [x] 6.1 Apply the minor version increase the owner chose, 1.0.0 → 1.1.0, before merge into main;
  verify `web/package.json` and `web/package-lock.json` carry the same version and the entry screen
  shows it.

## 7. Integration

- [x] 7.1 Run the full verification order from CONTRIBUTING.md — frontend tests, `npm run check`,
  production build, Go tests with the race detector against embedded assets, `go vet`, `gofmt -l .`
  and `openspec validate add-more-throw-objects --strict`; verify every step passes and write the
  results, including any check not run and why, to `verification.md`.
- [x] 7.2 Rebuild the Docker image and start it once without and once with
  `PLANNINGPOKER_POO_THROWS=true` in an `.env` file; throw every offered object between two
  browsers in one room and verify both see the same object, side and heart colour for each throw,
  that the pile of poo appears only in the second run, and that the response headers still carry
  the unchanged Content-Security-Policy.

## 8. Landing on the narrow list

- [x] 8.1 Replace the seat's bottom edge in `createFlight` with a landing zone and keep impact,
  slide and rest inside it; verify unit tests show every object lands within the zone for many
  seeds, a zone narrower than the object centres it, and the existing wide-table tests pass
  unchanged.
- [x] 8.2 Measure the zone in `ThrowLayer.svelte` per layout (the band below the seat on the wide
  table, the free space between name and controls in a narrow-list row); verify by screenshot at
  320 px with short and 15-character names that objects rest in the target's row without covering
  name, throw control or card, and that switching layout keeps resting objects at their seat.
- [x] 8.3 Add a Playwright test at 320 px in which a throw at the row above comes to rest inside
  that row; verify it fails against the previous landing and passes now.
- [x] 8.4 Run the full verification order, the browser suite and the Docker check again; verify
  every step passes and add the results to `verification.md`, correcting its statement that no
  requirement changed.
