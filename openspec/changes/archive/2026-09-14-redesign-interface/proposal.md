## Why

The interface is the first version anybody wrote and nobody chose it. Its palette is the blue every
framework ships with, its dividing lines are strong enough to speak rather than merely separate, and
its typography is whatever the visitor's operating system happens to provide — which means the
application looks different on every machine and has no voice of its own anywhere.

Three visual directions were built, rendered in the states that matter and compared side by side in
`docs/design/visual-language.html`. Direction B, "Clear room", was chosen: it takes the game
metaphor out, treats the table as a raised surface rather than green felt, and pushes contrast up.
Geist was chosen as the typeface over Inter and IBM Plex Sans.

Two smaller complaints are folded into the same change because they are about the same two screens
and would otherwise be restyled once and rebuilt again a week later: the join screen carries more
explanatory prose than a person reads before typing a name, and the entry screen puts the deck
choice above the start button while saying nothing about what is actually in either deck.

## What Changes

**Visual language.** The palette in `web/src/app.css` moves to direction B. Radii, shadows and the
table surface, which are hardcoded across the components today, become custom properties in
`app.css`, so that the visual design lives in one place and a later change is a value swap. Cards
get a 6px corner radius. The card back stops being a stripe pattern and becomes a calm surface with
a mark.

**Typeface.** Geist ships with the application as a variable `.woff2` taken from the upstream
release (68 KB, 728 glyphs) and served from its own origin, with its SIL Open Font License text
served beside it. It is placed in front of the existing system stack, which stays as the fallback.
Nothing is fetched from a foreign host, and the existing `font-src 'self'` directive covers it
without being widened.

The smaller latin subset that Google Fonts offers was rejected after being measured rather than
assumed: it lacks Polish, Czech, Hungarian, Romanian and Turkish letters, and the visible content of
this application is names that people type themselves.

**Entry screen.** The start control moves above the deck choice. The two decks become a bulleted
list of options in which each deck names its own cards, comma-separated, so that somebody who has
not seen a Fibonacci deck can tell what they are choosing. The card values come from a new
`GET /api/decks` endpoint rather than a second copy of the deck written into the page, so the server
stays the only place a deck is defined.

**Join screen.** The two explanatory paragraphs — the one saying a first name is enough and that
everyone with the link can see it, and the one saying what the storage choice stores and for how
long — move behind an information control marked `i`, one at the name field and one at the storage
choice. Each opens on hover, on keyboard focus and on tap, because hover alone would leave the text
unreachable on a phone and for anybody using a keyboard.

**BREAKING** — none. No protocol message, stored value, URL or configuration key changes, and a
browser connected during a deployment sees nothing different beyond the appearance.

## Capabilities

### New Capabilities

None. Everything here modifies behaviour that is already specified.

### Modified Capabilities

- `table-ui`: the entry screen's ordering changes and it must now name each deck's cards; the join
  screen's two advisory texts move behind an information control that must be reachable without a
  pointer; and two rules that hold for the whole interface are written down for the first time —
  minimum contrast, and respect for `prefers-reduced-motion`.
- `game-sessions`: a read-only endpoint that reports the decks a game can be started with, so the
  entry screen can name their cards without keeping its own copy.
- `app-delivery`: a third-party asset bundled into the application must ship with its licence text.
  The existing single-origin rule already covers the font itself; what it does not yet cover is the
  obligation that comes with the licence the font is offered under.

## Impact

- `web/src/app.css`: palette, and new custom properties for radii, shadows and the table surface.
- `web/src/components/`: `EntryScreen.svelte` (restructured), `NamePrompt.svelte` and
  `NameDialog.svelte` (information controls), `Seat.svelte` (card back), `Deck.svelte`,
  `Results.svelte`, `RoomView.svelte` (values move to properties). A new small component for the
  information control, used in both places that show it.
- `web/src/lib/decks.ts`: deck options gain their card lists, fetched rather than hardcoded.
- `web/public/` or an equivalent asset path: `Geist.woff2` and `Geist-LICENSE.txt`.
- `internal/transport/`: one new route and handler for `GET /api/decks`, plus its tests.
- `internal/game/`: no change. The decks are already defined there and are simply read.
- No change to the WebSocket protocol, the hub, the room lifetime, or any environment variable.
- `docs/design/visual-language.html` stays as the record of why the appearance is what it is.
