## Why

The page stores the name somebody typed for a year, on every seating and every rename, without
asking and without saying so. That is a convenience, not something the application needs in order to
work, and it is stored for far longer than any planning session lasts.

The seat token is a different thing entirely and is being conflated with it: it is what makes a
reloaded tab the same participant rather than a second one, so it carries the function rather than
the convenience. Its twelve-hour lifetime is nonetheless longer than the meeting it exists to
survive, which leaves a credential on a shared machine for half a day — and a persistent cookie is
also the harder shape to justify as strictly necessary, which is exactly what carrying the function
is supposed to buy it.

## What Changes

- Remembering a name becomes an explicit choice, **off unless somebody turns it on**, offered where
  the name is typed and reversible from the same place. Nothing is stored otherwise.
- Say what the stored name is and how long it stays, in the place where the choice is made rather
  than only in a document nobody opens.
- Make the seat token a **session** cookie — gone when the browser closes — instead of one that
  persists for twelve hours, keeping everything else about it.
- Mark the remembered name `Secure` when the page was served over HTTPS.
- Write a short, truthful account of what this application stores and for how long, with the parts
  that depend on who operates it left explicitly blank rather than invented.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `table-ui`: The name prompt gains an explicit, off-by-default choice to remember the name, and
  pre-filling follows from that choice rather than happening regardless.
- `game-sessions`: The seat token becomes a session cookie, and what that costs when it goes is
  stated with it.

## Impact

`web/src/lib/name.ts` (opt-in storage, `Secure`, forgetting), `web/src/components/NamePrompt.svelte`
and `RoomView.svelte` (the choice, and rename no longer storing unasked), `internal/transport/rooms.go`
(the seat cookie loses its max age), and a short privacy document written directly.

Answers the review finding that both cookies needed a deliberate decision, the technical half of describing what is stored, and the cookie part of the input-hardening findings. There are no existing
`pp_name` cookies to migrate: the application has never run anywhere but locally.

**Deliberately not in scope, because it is not mine to write:** who the controller is, the legal
basis, the contact for data subjects and any operator's hosting arrangements. Those are D3 and need
the owner. The document produced here states the technical facts and marks those gaps as gaps rather
than filling them with plausible-sounding text.
