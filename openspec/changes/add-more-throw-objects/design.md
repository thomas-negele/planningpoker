## Context

See [proposal.md](proposal.md) for motivation and the delta specifications for the agreed
behaviour. The owner confirmed both objects by name, asked for the pile of poo to be drawn like the
existing icons and as close as possible to the familiar emoji, and asked for hearts in varying,
exclusively cheerful colours. The owner also asked for the pile of poo to be switchable in
`compose.yaml` or an `.env` file, off by default, and chose a minor version increase to 1.1.0.

The throw feature already has every mechanism these objects need; this change adds entries to it
rather than new machinery:

- The room hub accepts an object only if it appears in a fixed list (`validThrowObject` in
  `internal/hub/throws.go`); everything else is refused as an unknown object. Limits, delivery and
  refusals do not depend on the object.
- The browser keeps a second, independent list: it discards a `thrown` event whose object it does
  not know (`web/src/lib/connection.svelte.ts`), and the shared `ThrowObject` type in
  `web/src/lib/protocol.ts` names the valid values.
- Startup configuration lives in `cmd/planningpoker/config.go`: each setting is an environment
  variable with a named constant, a documented default and a refusal at startup for invalid values.
  `PLANNINGPOKER_LEGAL_DIR` is the existing precedent for a feature that is off unless the operator
  turns it on. `compose.yaml` passes each variable as `"${NAME:-default}"`, so an `.env` file beside
  it overrides the default without editing the Compose file.
- Before throwing is enabled, the first snapshot on a connection carries a throw policy
  (`throwPolicyMessage` in `internal/transport/protocol.go`) with the message allowances; the
  browser keeps throwing disabled until it has one.
- `ThrowIcon.svelte` draws each object as inline SVG in a 48 × 48 coordinate box. The flower
  already derives shape and colour from the event's 32-bit seed, so every client draws the same
  flower; colours are CSS custom properties per palette.
- `createFlight` in `web/src/lib/throw-effects.ts` turns seed, object and layout into a flight
  (start, impact, landing, arc height, turns, slide distance, resting rotation), and `poseAt` turns
  a flight and an age into a pose: position, rotation and opacity. A pose has no scale today, so a
  squash or a pulse cannot yet be expressed. `ThrowLayer.svelte` renders the pose as an SVG
  `transform`.
- The picker in `Seat.svelte` is a single-row pill of 2 rem buttons. With three buttons it is about
  7.2 rem wide; with five it becomes about 11.7 rem (roughly 190 CSS pixels). On the wide table it
  is centred above a trigger at the seat's right edge; in the narrow list it is anchored to the
  right and grows leftwards.

## Goals / Non-Goals

**Goals:**

- Add both objects with the least change to the existing structure: list entries, artwork, motion
  parameters, and one small extension of the pose.
- Keep the motion deterministic per seed, so the existing tests' approach (injected seeds and
  clocks, no browser) covers the new behaviour.

**Non-Goals:**

- A switch for any object other than the pile of poo, a switch per room, or a runtime setting in
  an administration area. The owner asked for exactly one operator switch in the deployment
  configuration.
- Splats, stains, particles, sound or any effect that outlives the object's fade.
- Changing the existing three objects, the timings or the limits.

## Decisions

### Wire identifiers `heart` and `poo`

The two objects travel as `heart` and `poo` in the existing `object` field, beside `paper-ball`,
`paper-plane` and `flowers`. Short, lower-case and unambiguous. `pile-of-poo` (the Unicode
character name) was considered; it adds length on every event without adding clarity. The
accessible names in the picker are "Throw a heart at ‹name›" and "Throw a pile of poo at ‹name›",
matching the wording of the existing labels.

### The pile of poo is an operator switch read at startup

`PLANNINGPOKER_POO_THROWS` is read by `loadConfig` beside the other settings and becomes a boolean
field in `config`, with a full-sentence comment like its neighbours. It accepts exactly `true` and
`false`; unset or empty means `false`. Go's `strconv.ParseBool` was rejected because it also
accepts `1`, `t`, `TRUE` and others: the specification would then have to list them, and an
operator reading `compose.yaml` would see two documented values while the program quietly accepted
more. Any other value stops the process with the variable and the value named, as every other
setting does.

The value travels from `main` into the room manager as the set of accepted objects, and
`validThrowObject` checks against that set instead of a fixed `switch`. The hub stays the one place
that decides; the transport layer copies the same set into the throw policy so the browser can
build its picker from it. The browser never decides on its own whether the pile of poo exists, so
there is no second default to keep in step.

The owner chose an environment variable over a managed setting in an administration area. The
consequence, stated so it is not discovered later: switching it requires a restart of the
container, and a restart ends every open room. `compose.yaml` defaults it to `false`, with an
explanation in full sentences, so the owner's server enables it with one line in its `.env` file.

The alternatives were a generic list of enabled objects (more than was asked for, and it invites
configuring away the heart or the paper ball, which nobody requested) and a switch per room in the
interface (a room setting, which this project deliberately does not have).

### Own SVG drawings, not the system emoji

Both objects are drawn as inline SVG in `ThrowIcon.svelte`, like the other three. Rendering the
emoji characters ❤️ and 💩 from the operating system's font was rejected: they would look different
on every platform (and on some Linux systems not at all), would not match the drawn style of the
other objects, and their colour could not follow the seed. Copying or tracing a vendor's emoji
artwork was rejected because that artwork belongs to its vendor.

The pile of poo is built from the description alone: three stacked, rounded tiers that narrow
upwards in a warm mid-brown, a tip curling to one side, a darker brown edge and lighter highlights
in the existing icons' manner, two white eyes with dark pupils, and a wide, open grin. The heart is
one rounded heart with a darker edge of its own hue and a light highlight, so it reads as the same
family as the flower.

### Heart colour from the seed, out of seven cheerful palettes

The heart's palette is chosen as `seed % 7` from red, pink, orange, yellow, green, blue and purple.
Each palette, like the flower's, defines a fill, a lighter highlight and a darker edge of the same
hue. The edge keeps light colours such as yellow distinguishable against both the light and the
dark table surface. Black, white, grey and brown hearts — all part of common emoji sets — are
excluded because the owner asked for cheerful colours only.

The picker shows the heart in red, regardless of any seed, so the choice is always recognisable as
"heart". The picker icons for both objects fall under the existing contrast requirement for
meaningful graphics (3:1 against the picker surface in light and dark mode); their edges are
measured rather than assumed.

### A scale in the pose, and the squash anchored at the base

`Pose` gains `scaleX` and `scaleY`, both 1 for every existing object and for every resting, fading
and reduced-motion pose. `ThrowLayer` appends the scale to the existing transform. This is the
smallest change that can express a squash and a pulse; a separate animation mechanism, such as CSS
keyframes per object, was rejected because it would run outside `poseAt`, escape the injected-clock
tests and need its own handling for reduced motion and sleeping tabs.

The pile of poo squashes about its base, so it flattens onto the ground rather than shrinking
towards its centre and appearing to float. The heart pulses about its centre.

### Motion parameters

Both objects reuse the existing flight interval (0.6–1.2 seconds), the impact area and the
seed-derived direction of the slide. Only their parameters differ:

- **Pile of poo:** a flatter arc than the others, so it reads as heavy; a slight sway of a few
  degrees instead of a tumble; a short bounce at impact for the plop; a slide of only a few pixels.
  At impact `scaleY` drops to about 0.7 while `scaleX` widens to about 1.25, then recovers as a
  damped oscillation — the wobble — that has fully decayed by the end of the settling phase.
- **Heart:** a higher, lighter arc; a gentle rotation similar to the flower; a soft landing and a
  slide shorter than the paper plane's. During settling it grows to about 1.25 and returns to 1
  exactly once.

The exact values are tuned in the browser during implementation. The tests pin the properties the
specification names, not the numbers: squash and pulse happen during settling and are back at 1
before rest; the pile of poo's slide is shorter than the paper ball's; the heart's scale passes
through a single maximum.

### The wider picker stays one row

The picker keeps one row of 2 rem buttons. Two rows or smaller buttons were rejected: two rows
change the pill's shape and push it further over the seat, and smaller buttons make the touch
targets worse on exactly the narrow screens where touch is used.

Whether 11.7 rem fits is checked in the browser for the outermost seats of the wide table and for
the narrow list at 320 pixels. If a picker at an outer table seat would cross the viewport edge, it
opens towards the table's centre instead of being centred on its trigger — the narrow list already
does the equivalent by anchoring to the right. The outcome of this check is recorded in the
verification notes either way.

## Risks / Trade-offs

- [A pile of poo is not to every team's taste in a work meeting] → It is off unless the operator
  enables it for the whole server; there is no per-room choice.
- [Changing the switch needs a restart, which ends open rooms] → Accepted by the owner's choice of
  an environment variable; documented in `compose.yaml` and the README.
- [Seven heart colours plus flower colours make the table busier] → Hearts and flowers share the
  existing fixed throw limits, so the number of objects on screen at once does not grow.
- [A yellow heart on the light table could lack contrast] → The darker edge of each palette is
  measured against both surfaces.
- [An open tab from before the update ignores the new objects] → Intended: it already discards
  objects it does not know, and a reload fixes it. Restarting the container for the update ends
  every open room anyway.

## Migration Plan

Deploy as usual; the pile of poo stays off until `PLANNINGPOKER_POO_THROWS=true` is set and the
container is restarted. The server and the page ship together in one container. Rolling back removes
both objects; a leftover `PLANNINGPOKER_POO_THROWS` variable is ignored by the older version, which
does not read it.
