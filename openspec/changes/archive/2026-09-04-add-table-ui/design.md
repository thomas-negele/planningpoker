## Context

See `proposal.md` — Why. Everything this change needs already exists on the server
and none of it should have to move: the protocol sends whole snapshots and typed
refusals, the seat cookie survives a reload, and the room outlives a brief
disconnection. What is missing is entirely inside `web/`.

Three constraints from earlier changes shape the work, and one of them is a trap.

The **Content-Security-Policy** forbids inline styles. That includes `style="..."`
attributes on elements, not only `<style>` blocks — and this is exactly the sort of
interface where somebody reaches for a computed inline style to place people around
a table. It would work perfectly in development, where no policy is applied, and be
blocked by the browser in production. Assume this will be attempted at least once.

The **client holds no authoritative state**: it renders the last snapshot and sends
intents. This was decided at the very start to avoid the classic bug where one
person's screen disagrees with the server about whether the round was revealed.

**Nothing may be fetched from another host.** No icon library, no web font, no CSS
framework, no clipboard polyfill from a CDN.

## Goals / Non-Goals

**Goals:**

- Make the product usable by a person, matching the interface described in the
  original specification and the reference screenshot: table in the centre, people
  around it, deck along the bottom.
- Exercise, for the first time with a real browser, the two mechanisms built for
  interruption — the seat cookie and the room grace period.
- Keep the interface honest about the rules: no control disabled for a reason the
  server would not enforce.

**Non-Goals:**

- Any server change. If one appears to be needed, that is a finding to report, not a
  thing to quietly do.
- A design system, a component library, or styling meant to be reused elsewhere.
- Animation beyond what makes a card change legible.
- Anything on the excluded list: other decks, deck selection, accounts, history,
  statistics, timers, spectators, room settings.

## Decisions

### Dynamic styles from components are safe; styles written into markup are not

This was measured rather than assumed, because the answer decides how the table may
be laid out. Chrome 152 was served a page under this project's exact policy, and
four ways of setting a style were tried:

| How the style is set | Under `style-src 'self'` |
|---|---|
| `style="..."` written in the HTML the browser parses | **blocked** |
| `element.setAttribute('style', …)` from JavaScript | **blocked** |
| `element.style.setProperty(…)` | allowed |
| `element.style.cssText = …` | allowed |

The two blocked cases log "Applying inline style violates the following Content
Security Policy directive" and the declaration is discarded. The two allowed cases
are the object model, which the policy does not govern at all.

Svelte then decides which of these we actually get, and it was checked against the
installed compiler. Both `style="left: {x}%"` and `style:left={x}` compile to
`set_style`, which goes through the object model. **Neither is blocked.** Svelte
never emits a markup style attribute and never calls `setAttribute('style', …)`.

So the table may be laid out whichever way reads best, including by computing seat
positions, and a stylesheet-only arrangement is a preference rather than a
requirement. Where a plain CSS arrangement is natural — a grid, a flex row above and
below the table — prefer it, because it is simpler and responds to a narrow screen
without help. But do not contort the design to avoid dynamic styles.

What remains genuinely blocked is worth knowing, because it is the part that will
bite:

- a `style="…"` attribute written by hand into `web/index.html`, which the browser
  parses from markup;
- anything injected with `{@html …}` that contains a style attribute;
- attribute spreads where a `style` key ends up going through `setAttribute` rather
  than `set_style`.

The acceptance check stays exactly as it was: load the finished table in a
production build with several participants and confirm the console reports no
violation. Measuring the mechanism does not remove the need to check the result.

And the rule that does not change: if a violation ever appears, fix the code that
caused it. **Never add `'unsafe-inline'` to the policy** — that is the value which
makes it worth having.

### One connection module, one store, dumb components

A single module owns the socket: connecting, reconnecting, sending intents, and
exposing two things — the last snapshot and the connection state. Components read
those and render; they never touch the socket.

This keeps the "client holds no authoritative state" rule enforceable by looking at
one file rather than by reviewing every component. It also means reconnection logic
exists once, instead of being spread across whichever component happened to notice.

### Reconnection backs off, with jitter, and stops when the game is gone

Delays grow after each failed attempt up to a cap, and each delay is randomised
slightly. The growth stops a browser left open overnight from retrying every second;
the cap stops it from waiting ten minutes after the server comes back; the jitter
stops every participant of a meeting from reconnecting in the same instant when a
deploy ends, which is the moment they would all be retrying together.

Two signals end retrying permanently rather than triggering it: the server's
`room_not_found` refusal, and its close code for the same. Both mean the room is
gone, and since nothing ever recreates a room at a used identifier, retrying is
guaranteed to fail forever.

A server shutdown closes with "going away" instead, which *is* retried — and the
retry then discovers the room is gone and stops. That is the correct sequence rather
than a special case: the page does not need to know why the server went away, only
what it finds when it comes back.

### The name lives in a cookie the page writes; the seat token does not

The name cookie is written and read by the page, because the page is what needs it —
to pre-fill the field. It is an ordinary readable cookie holding something the
person typed about themselves, and it carries nothing that could be used against
them.

The seat token remains what it already is: set by the server, hidden from scripts,
scoped to the room. The page never sees it and never needs to; the browser attaches
it to the socket automatically. These two cookies are deliberately unlike each other
and it is worth not blurring them while working on the same screen.

### Routing is two routes written by hand

`/` is the entry screen and `/g/{roomID}` is a table. That is the whole routing
requirement, and it is met with the History API and a listener for the back button
in a few dozen lines.

A router library would be a dependency to maintain for two routes, and this project
already decided against SvelteKit for the same kind of reason. The server's
single-page fallback already returns the document for any unmatched path, so a deep
link and a hard reload both work.

### Refusals are translated from codes, exhaustively

Each refusal code from the protocol maps to a sentence here. The mapping is written
out rather than defaulting, so a code with no sentence is visible during development
instead of silently rendering as a generic apology. This mirrors how the server maps
its sentinel errors to codes, for the same reason.

### The interface never enforces a rule the server does not

Reveal and revote are always available, because the rules place no condition on
them. The interface may report that everyone present has voted; it must not turn
that into a precondition.

The temptation here is real — disabling a button feels tidier than letting somebody
reveal a round early. But a control disabled for a reason the server would not
enforce teaches people something untrue about the product, and the rules were
written the way they were on purpose: one absent colleague must never be able to
block a meeting.

## Risks / Trade-offs

- **The Content-Security-Policy breaks the table in production only.** → Layout by
  stylesheet, plus an acceptance check in a real browser against a production build,
  with the console inspected for violations. This is the single most likely way this
  change ships broken.
- **Reconnection is easy to get subtly wrong** — retrying something unrecoverable
  forever, or giving up on something temporary. → Test both directions explicitly:
  a dropped socket to a room that still exists must recover unaided, and a room that
  is gone must stop and say so.
- **A stale table that looks live is worse than an obviously broken one.** → The
  connection state is part of the interface rather than a detail, and controls that
  send intents are unavailable while there is nothing to send them over.
- **The temptation to keep client-side state** will appear the moment something
  needs to feel instant — showing a card as played before the server confirms it. →
  The snapshot is the truth. If responsiveness ever genuinely requires anticipating
  the server, that is a change to propose, not one to slip in.
- **Emoji rendering varies by platform**, so the coffee cup will not look identical
  everywhere. → Acceptable and expected: it comes from the operating system's own
  font, which is exactly why it costs no network request. Shipping an image instead
  would violate the self-contained rule.

## Migration Plan

Additive from the user's point of view and destructive only of the placeholder,
which is replaced wholesale. No data exists to migrate, no protocol changes, no new
configuration. Deployment is unchanged, with the standing warning that a deployment
ends every game in progress.

Rollback is redeploying the previous image, which returns to a placeholder page and
a working but unusable server.

## Open Questions

- The exact backoff numbers — first delay, growth factor, cap — are not fixed here.
  They need to be short enough that a blip is invisible and long enough that a
  hundred tabs do not stampede a returning server; anything in the region of a
  second growing to under a minute satisfies both. They are settled during
  implementation and recorded in the code with their reasoning, and become settings
  only if they ever turn out to need adjusting.
- How many participants the table arrangement should stay comfortable with before it
  simply scrolls. Planning poker is played by a team, so the honest answer is
  somewhere around a dozen; it is settled by looking at it rather than by deciding
  now.
