# Implementation verification

Verified on 2026-09-11 from `participant-throws`.

## Automated checks

- `cd web && npm ci`: passed. npm printed only its existing optional `fsevents` install-script
  policy warning.
- `cd web && npm test`: passed, 10 tests. The tests cover seeded geometry from both viewport
  edges, all three object lifecycles, reduced motion, stale/missing/full effect admission and local
  pacing at message rates 1 and 10.
- `cd web && npm run check`: passed with 0 errors and 0 warnings.
- `cd web && npm run build`: passed; the production bundle was written to
  `internal/webassets/dist`.
- `go build ./...`, `go vet ./...`, `gofmt -l .` and `git diff --check`: passed; both formatting
  commands produced no output.
- `go test -race -count=3 ./...`: passed.
- `go test -race -count=3 -tags embedassets ./...`: passed.
- `go test -race -count=20 ./internal/transport -run TestAFullTableVotingAtOnceIsNeverSlowed`:
  passed while investigating the reliable-queue burst capacity under race instrumentation.
- Focused hub and real-WebSocket fixtures passed for eligibility, exact rolling quota boundaries,
  same-seat tabs/reconnect, room isolation, cosmetic saturation, initial-snapshot ordering, slow
  recipients, no history and hidden-data exclusion. The saturated-room fixture gates cosmetic
  work while vote, reveal and new-round commands continue.
- `openspec validate add-participant-throws --strict`: passed.

## Production inspection

- `docker compose -f compose.yaml -f compose.local.yaml up --build -d`: built and started
  successfully at `127.0.0.1:8080`.
- HTTP returned 200 with the unchanged single-origin CSP. Firefox Network showed only the document,
  compiled JS/CSS, WebSocket, legal endpoint and favicon from the local origin; there were no CSP
  errors, foreign requests or external runtime assets.
- Two independent participants and an extra tab restoring the first participant's seat joined the
  same room. A sent `throw` and matching server `thrown` frame carried the same target, object,
  event ID and server seed, and contained neither credentials nor hidden card values.
- Wide and 320x480 layouts were inspected. A 15-character name remained readable, seat geometry
  stayed stable, the three distinct local SVG choices fit the picker, and touch simulation opened
  the same picker only for the other present participant. No shortcut or keyboard-help UI appeared;
  the accessibility tree exposed target-specific native button names.
- Throw selection and ordinary room controls remained interactive. Repeated-click suppression,
  both seeded entry edges, object-specific motion, reduced-motion direct landing, elapsed-time
  cleanup, the 48-object bound and low-rate behaviour were verified by deterministic frontend
  tests rather than inferred from a short visual sample.
- Docker logs contained only the normal listening message. No hardware-throughput or general
  denial-of-service claim is made; the verified guarantees are the fixed quotas and bounded,
  best-effort queues described by the change.

## Feedback refinement

- The production Docker image was rebuilt after the visual and motion revision and returned HTTP
  200. The server was intentionally left running for owner feedback.
- Firefox with two independent participants showed only the three throw choices on desktop; the
  former round paper-ball trigger was no longer visible below the picker. The native trigger
  remains available for coarse pointers and as a focus-visible keyboard accessibility path.
- The revised paper ball was inspected at picker and flight sizes: it has a rounded irregular
  silhouette, layered paper regions and crease lines instead of the former faceted outline.
- The flower is now a single stem and bloom. Its three shapes and four palettes are selected from
  the shared event seed; a single-flower flight was inspected in the production build.
- Pure trajectory tests now assert broad seed variation in both impact and final-rest offsets,
  bounded placement, momentum-direction continuation and the ordered ball/plane/flower slide
  distances. Reduced motion still begins directly at the final resting point.
- After the refinement, `npm ci`, all 10 frontend tests, Svelte checking with 0 warnings, the
  production build, Go build/vet/format checks, and both three-run race suites passed again.

## Review remediation (2026-09-12)

- A reduced-motion preference change now gives an in-flight effect its full 2 second rest from
  the change, while already resting or fading effects retain their elapsed age and lose rotation.
  A new deterministic test covers both transitions and the unchanged expiry boundary.
- The SVG layer now computes each pose once per frame and caches each resting target's DOM
  rectangle once per frame. A new test checks shared-target measurement, missing targets and
  expired effects after a sleeping-tab interval.
- The duplicate `drain(actor)` in a Go test was removed. Design now records the reliable update
  buffer's 16-to-32 adjustment and the actual participant-ID anchor lookup.
- A separate Playwright smoke suite passed 2/2 tests against an isolated production-built Go
  server: desktop hover/picker departure leaves no visible trigger, and narrow touch activation
  offers the same three choices. Its locked dependencies live in `e2e/`, outside the Docker
  build context; `npm ci` and the E2E type check also passed. It does not replace the pure
  lifecycle or Go race tests.
- The final web suite passed 12/12 tests, Svelte checking found 0 errors and 0 warnings, and the
  production build passed. Both Go suites passed with `-race -count=3`, as did build and vet.
  `gofmt -l .`, `git diff --check` and strict OpenSpec validation produced no findings.
- The current production Docker image rebuilt successfully after isolating the browser-test
  dependencies. The replaced container returned HTTP 200 with the unchanged single-origin CSP;
  its logs showed only the normal listening message. The image remains running on
  `127.0.0.1:8080` for owner feedback.
