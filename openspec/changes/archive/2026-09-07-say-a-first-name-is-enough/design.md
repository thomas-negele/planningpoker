## Context

See proposal.md. The only thing shaping this is that a name is entered in two places since package
D: `NamePrompt.svelte` on arriving, and `NameDialog.svelte` when changing it at the table. Both
already carry the sentence explaining the storage choice, so there is a precedent for how such a
line looks and where it sits.

## Goals / Non-Goals

**Goals:** Say, at the moment somebody chooses what to be called, that a short name is enough and
who will see it.

**Non-Goals:** Any check on what is typed, any warning, any change to the deliberately open room
links, and any attempt to detect whether a name is real.

## Decisions

1. **One sentence, identical in both places.** "A first name or nickname is enough — everyone with
   the link to this room can see it." Stating the reason rather than only the advice is what makes
   it act on: "use a short name" invites the question why, and the answer is the whole point.

2. **It lives beside the field, above the storage choice.** Reading order matters: what will be
   visible to others is the question somebody answers by typing, while whether it is kept on the
   device is a separate question they answer afterwards.

3. **Defined once in `name.ts` and imported**, next to `REMEMBERED_FOR`. The specification requires
   the same wording in both places, and two string literals in two components are two things that
   drift apart. This makes "identically" true by construction rather than by discipline.

4. **Deliberately not a warning.** No colour, no icon, no role="alert" — it is the same quiet grey
   as the storage explanation. A name is not an error, and treating an ordinary act as a hazard
   teaches people to dismiss the notice rather than read it.

## Risks / Trade-offs

- **One more line at a prompt that already has two** → it is one short sentence, and the prompt is
  the only screen in this application that anybody reads carefully.
- **A hint nobody reads changes nothing** → true, and accepted: the alternative is a rule about
  what may be typed, which this project has no business imposing on a display name.

## Migration Plan

None. No stored data, no protocol, no configuration; a rebuild is the whole of it.
