## Context

See proposal.md. What exists today: `hub.ErrAtCapacity` carries the sentence "the server is holding
as many rooms as it can" and is returned by `Manager.Create` and `Manager.EnsureRoom`. The transport
maps it to the code `at_capacity`, and — this is the mistake — also hands that same error to
`sendRefusal` when `Room.Attach` reports `AttachAtCapacity`, which is the per-room connection
ceiling and a different thing entirely.

The hub already tells the two apart internally: `AttachAtCapacity` is a distinct value returned by
`Attach`. Only the error handed onward from it was wrong.

## Goals / Non-Goals

**Goals:** Name the per-room connection ceiling as its own refusal, all the way to the sentence the
person reads.

**Non-Goals:** Changing any limit, changing when a refusal happens, or touching the seat ceiling —
`room_full` already exists for a full table and is a third, unrelated thing. No new configuration.

## Decisions

1. **A second sentinel, `hub.ErrRoomAtCapacity`**, rather than wrapping the existing one or passing
   a flag. The transport already maps sentinel errors to protocol codes through one table, and a
   test asserts every domain sentinel has its own code, so a second error is the shape this codebase
   already has for "a different refusal".

2. **The protocol code is `too_many_connections`, not `room_at_capacity`.** There are now three
   nearby refusals and the names have to be tellable apart at a glance in a log:
   `at_capacity` (no room left in the process), `room_full` (no seat left at the table),
   `too_many_connections` (no socket left in this room). Naming the third after the room would put
   two "room" codes next to each other meaning different limits, which is how somebody later
   collapses them by accident.

3. **The page keeps one screen for both, with different words.** `ServerFull.svelte` takes a
   `scope` of `server` or `room` and owns both wordings. The layout, the "your link is fine" framing
   and the retry button are right for both; only the heading, the explanation and the advice differ
   — waiting for the server, closing a spare tab for the room. Two components would duplicate the
   layout to vary three strings.

4. **Both still stop the automatic reconnecting.** That is unchanged and deliberate: retrying
   against a full server adds to the load that caused it, and retrying into a room that is out of
   sockets cannot succeed until somebody leaves either. The page says so and lets the person choose
   when to try again.

## Risks / Trade-offs

- **A third nearby code invites confusion between the three** → hence the naming in decision 2, and
  the specification now says in words why the two ceilings are separate entries, so a later reader
  finds the reasoning rather than guessing at it.
- **The frontend's mapping table must gain the new code or the sentence falls back to a loud
  placeholder** → that fallback is deliberate in this codebase and logs to the console; the check
  run against the real page will show it immediately if it is missed.

## Migration Plan

None. No stored data, no configuration, no compatibility surface: an older page receiving the new
code would show its generic "unexpected refusal" sentence, and there are no deployed instances.
