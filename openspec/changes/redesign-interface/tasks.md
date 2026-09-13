## 1. The deck endpoint

- [ ] 1.1 Add `GET /api/decks` to the transport layer, answering from the same `game.TShirtDeck()`
      and `game.FibonacciDeck()` that room creation uses, with each deck's stable name, label and
      cards in deck order. Verify with a handler test asserting the payload for both decks and the
      card order.
- [ ] 1.2 Add a test asserting the endpoint creates nothing — no room, no seat, no `Set-Cookie` —
      and that the manager's room count is unchanged after repeated requests. Verify with
      `go test ./internal/transport/...`.
- [ ] 1.3 Add a test asserting the endpoint's cards for a deck equal the cards a room created with
      that deck offers, so the two cannot drift. Verify with `go test ./internal/...`.
- [ ] 1.4 Confirm the route is reachable through the development proxy as well as the embedded
      build. Verify by requesting it against `go run ./cmd/planningpoker` and against a binary built
      with `-tags embedassets`.

## 2. The typeface

- [ ] 2.1 Commit the Geist variable `.woff2` (latin subset, weights 400–700) to the frontend asset
      path, together with its SIL Open Font License text stored beside it and named so the pairing
      is obvious. Verify both files are present and the licence names Geist.
- [ ] 2.2 Declare a local `@font-face` with `font-display: swap` and put the family at the front of
      `--font-stack`, keeping the existing system stack behind it. Verify the running page renders
      in Geist and that removing the file falls back to the system stack rather than to a serif.
- [ ] 2.3 Give the smallest uppercase labels — the `away` tag above all — explicit letter-spacing,
      because Geist sets tighter than the system stack there. Verify by eye at the table in both
      layouts.
- [ ] 2.4 Verify the self-contained rule still holds: search the built output for `http://` and
      `https://`, and confirm in the browser's network panel that every request, the font included,
      targets one origin. Verify no Content-Security-Policy violation appears in the console.

## 3. The visual design behind properties

- [ ] 3.1 Move the radii, shadows, table surface, card-back pattern, bar colours and deck veil out
      of the components into new custom properties in `web/src/app.css`, leaving the existing colour
      property names unchanged. Verify `npm run build` succeeds and the appearance is unchanged at
      this point, since only the location of the values has moved.
- [ ] 3.2 Apply the direction B values from `docs/design/visual-language.html`. Verify the entry
      screen, a running round and a revealed round against the template in the same three states.
- [ ] 3.3 Replace the striped card back in `Seat.svelte` with the calm surface and its centre mark.
      Verify a face-down card is distinguishable from the background at a ratio of at least 3:1.
- [ ] 3.4 Measure every text-on-surface and control-boundary pair in the new palette against the
      ratios the specification names, and record the measured figures in the change's verification
      notes. Verify by moving any failing value in `app.css` and measuring again.

## 4. The entry screen

- [ ] 4.1 Put the start control above the deck choice and render the decks as a list of options, one
      per deck, `T-shirt sizes` selected by default. Verify the existing behaviour still holds:
      selecting Fibonacci and starting creates a Fibonacci room.
- [ ] 4.2 Fetch the deck list from `GET /api/decks` on load and show each deck's cards
      comma-separated in the deck's own order. Verify the listed cards equal the cards the table
      then deals.
- [ ] 4.3 Handle a failed or slow fetch: both decks stay offered by name and starting a game still
      works. Verify with the request blocked in the browser's network panel, and with a Playwright
      test that intercepts and fails the route.
- [ ] 4.4 Keep the option list operable by keyboard, as the deck options are today. Verify by
      selecting each option with the keyboard alone.

## 5. The information controls

- [ ] 5.1 Build one small component for the `i` control and its text: discloses on hover, on
      keyboard focus and on tap, closes on `Escape` and on a click outside, and is tied to the
      control it explains through `aria-describedby`. Verify with a component test covering all
      three ways of opening it.
- [ ] 5.2 Use it in `NamePrompt.svelte` at the name field for the visibility hint, and at the
      storage choice for the storage text, removing the two visible paragraphs. Verify both texts
      are still the identical shared strings from `lib/name.ts`, not retyped copies.
- [ ] 5.3 Reword the storage choice's own label so that it states, without anything being opened,
      that the name would be kept on this device. Verify against the specification scenario "The
      offer is not misleading before it is opened".
- [ ] 5.4 Apply the same treatment where a name is entered at the table, in `NameDialog.svelte`, so
      the fact is worded identically in every place a name is entered. Verify both places show the
      same string.
- [ ] 5.5 Add Playwright coverage for reaching an information control by keyboard and by tap on a
      touch viewport. Verify with `cd e2e && npm test`.

## 6. Reduced motion

- [ ] 6.1 Honour `prefers-reduced-motion: reduce` for the card lift, the surface transitions and the
      thrown objects, so that movement is removed while what it conveyed remains visible — a thrown
      object still arrives and is still attributed. Verify with the system setting enabled, not by
      reading the code.
- [ ] 6.2 Add Playwright coverage under an emulated reduced-motion preference asserting a thrown
      object is still seen and attributed. Verify with `cd e2e && npm test`.

## 7. Verification and record

- [ ] 7.1 Run the project's verification order from CONTRIBUTING.md in full:
      `(cd web && npm ci && npm test && npm run check && npm run build)`, then `go vet ./...`,
      `gofmt -l .`, `go test -race -count=3 ./...` and
      `go test -race -count=3 -tags embedassets ./...`. Verify `gofmt -l .` prints nothing and every
      suite passes.
- [ ] 7.2 Run the browser suite: `cd e2e && npm ci && npx playwright install chromium && npm test`,
      and `cd e2e && npm run check`. Verify both pass.
- [ ] 7.3 Check the wide table layout and the narrow list layout across the 58rem breakpoint, and
      confirm the behaviours the specification fixes still hold: seats on the ellipse, a full-length
      name shown whole, the deck reachable along the bottom edge, no average computed.
- [ ] 7.4 Record the verification results in the change, naming explicitly any check that was not
      run and why.
- [ ] 7.5 Update `docs/design/visual-language.html` if the implementation departed from it anywhere,
      so that the document and `app.css` do not disagree. Verify by comparing the template's value
      list for direction B against the shipped `app.css`.
