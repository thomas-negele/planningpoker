## Why

The three throwable objects — paper ball, paper plane and flower — cover a nudge and a friendly
gesture, but not the two reactions a table most often wants to send each other: open affection and
good-natured disapproval. The owner asked for both by name: a heart, and a pile of poo drawn as
closely as possible to the familiar emoji.

## What Changes

- Add two objects to the throw picker: Heart and Pile of poo. The server accepts both under the
  same rules as the existing three — same eligible senders and targets, same refusals, same fixed
  limits of 3 throws per participant and 12 per room per rolling second.
- Make the pile of poo an operator switch, `PLANNINGPOKER_POO_THROWS`, set in `compose.yaml` or an
  `.env` file beside it. It is off by default; the owner will switch it on for their own server.
  While it is off, the picker offers four objects and the server refuses the fifth. The owner chose
  an environment variable over a managed setting, so a change takes effect on the next restart.
  The heart is always available.
- Draw both as local SVG illustrations in the style of the existing throw icons. The pile of poo
  resembles the familiar emoji — a brown, three-tiered swirl with a pointed tip, two eyes and a
  wide grin — but is drawn for this project rather than copied from any vendor's emoji artwork.
- Vary each heart's colour from the throw's shared seed, chosen only from cheerful colours, so every
  participant sees the same heart in the same colour.
- Give each new object its own motion: the pile of poo flies heavily, lands with a small plop,
  squashes briefly on impact, wobbles and barely slides; the heart flies lighter, lands softly and
  pulses once. Neither leaves a stain, trail or other residue after it fades. Reduced motion shows
  both directly at their resting position, without squash, wobble or pulse.
- The browser learns from the server which objects it may offer, as part of the throw policy it
  already receives before throwing is enabled.
- Keep the wider picker within the existing layout rules: it must not alter seat geometry, obscure
  names, cards or controls, or leave the viewport on the wide table and the narrow list.

No existing object, timing, limit or delivery rule changes, and no existing protocol field changes;
the throw policy only gains the list of accepted objects. A browser tab
that was opened before the update and is still running the old page never offers the new objects
and silently ignores hearts and piles of poo thrown by others, because it already discards objects
it does not know; reloading the page brings it up to date. An updated page talking to an old server
would receive the existing "unknown object" refusal, which does not arise in practice because both
come from the same container.

**Version:** this is a functional, backward-compatible addition. The owner chose a minor increase,
from 1.0.0 to 1.1.0.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `participant-throws`: The set of objects the server accepts grows from three to four, or five
  with the pile of poo enabled; the switch, its default and its invalid values are specified.
- `live-updates`: The throw policy sent to the browser lists the objects the server accepts.
- `app-delivery`: The switch joins the documented behaviour-governing environment variables.
- `table-ui`: The picker offers four or five named choices; the heart and pile of poo have specified
  artwork, seed-varied heart colours and object-specific motion; reduced motion excludes the new
  squash, wobble and pulse.

## Impact

- **Server:** two new object identifiers in the room hub's list of accepted objects
  (`internal/hub/throws.go`), the switch in the startup configuration (`cmd/planningpoker`) and its
  path into the hub, and a list of objects in the throw policy (`internal/transport`), with the
  matching Go tests. Existing messages keep their fields; the throw policy gains one.
- **Deployment:** `compose.yaml` gains the switch with its explanation; the README describes it.
- **Frontend:** the shared object type (`web/src/lib/protocol.ts`), two new drawings and a heart
  palette in `ThrowIcon.svelte`, two new choices in the picker (`Seat.svelte`), per-object motion
  parameters and a brief scale effect for squash and pulse in `web/src/lib/throw-effects.ts`, and
  the rendering layer that applies it (`ThrowLayer.svelte`).
- **Tests:** Go tests for acceptance of the new objects; frontend unit tests for the new motion;
  the Playwright smoke suite in `e2e/`, which currently expects three picker choices.
- **Version:** `web/package.json` and its lock file, raised to 1.1.0.
- No new dependencies, no external assets, no change to the Content-Security-Policy, and no
  settings beyond the one switch.
