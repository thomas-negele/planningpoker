## Context

See proposal.md for motivation. What shapes the approach:

- `web/src/lib/name.ts` writes `pp_name` from `document.cookie` with a one-year max age. It is called
  unconditionally from `NamePrompt.svelte` on confirming and from `RoomView.svelte` on renaming.
- The seat cookie is set by the server in `internal/transport/rooms.go`, with `seatCookieMaxAge` as
  a named constant beside the reasoning for its value.
- There are no deployed instances, so no `pp_name` cookie exists anywhere but on the owner's own
  machine. Nothing has to be migrated, which removes the awkward part of this kind of change.
- The owner decided both points directly: the name becomes an explicit choice that is off by
  default, and the seat token keeps its shape but becomes a session cookie instead of a
  twelve-hour one. The second decision was revised once, from four hours to a session, after the
  consent-exemption guidance was read properly; decision 6 records why.

## Goals / Non-Goals

**Goals:** Store the name only when somebody asked for it, say so where they are asked, let them
take it back in the same place, keep the seat token no longer than the browser session, and write
down truthfully what this application keeps.

**Non-Goals:** A consent framework or a banner — one checkbox in the place it applies is the whole
mechanism. Server-side storage of anything. Server-enforced expiry of the seat token, which is a
different piece of work and is not what was asked for. And every operator-specific fact: controller,
legal basis, contact, hosting. Those are D3.

## Decisions

1. **The choice lives in the name prompt, not in a banner.** It is where the name is typed, which is
   the only moment it means anything, and it is the moment somebody is already deciding what to be
   called. A banner over the page would ask about storage before there was anything to store, and
   would be the "large consent apparatus" the checklist explicitly rules out for a single optional
   convenience.

2. **Unchecked by default, and the checkbox is the whole state.** Checking it stores the name;
   unchecking it deletes what was stored, immediately, rather than only stopping future writes. One
   control that both grants and withdraws is easier to reason about than a preference plus a
   separate "forget me" button, and it makes withdrawal as easy as giving.

3. **`name.ts` gains `forgetName` and stops writing unconditionally.** `rememberName` is called only
   when the choice is on; `RoomView`'s rename path passes the same choice through rather than
   storing on its own. The module keeps owning the cookie's shape so there is still one place that
   knows its name, path and lifetime.

4. **`Secure` when the page was served over HTTPS**, decided by `location.protocol`. It cannot be
   set unconditionally, because on a plain-HTTP development server a `Secure` cookie is silently
   dropped and the feature would appear broken for no visible reason. This mirrors what the server
   already does for the seat cookie with `X-Forwarded-Proto`.

5. **Keep the stored lifetime at one year, now that it is chosen.** Shortening it was considered and
   rejected: the value of the feature is precisely that a fortnightly planning session still finds
   the name, and somebody who has asked for it and been told the duration is not better served by
   quietly forgetting sooner. The honest lever is the choice, not a shorter secret default.

6. **The seat token becomes a session cookie**: the max-age attribute goes away entirely, so the
   browser keeps it for the current session and discards it on closing. Nothing else about that
   cookie changes.

   **This replaces an earlier decision of four hours, and the reason is worth recording.** The
   Article 29 Working Party's Opinion 04/2012, the standard reading of the consent exemption that
   § 25(2) No. 2 TDDDG restates, names authentication cookies as a clear example of the exemption —
   but says an exempt cookie's lifespan must stand in direct relation to its purpose and expire once
   it is not needed, treating authentication as exempt *for the duration of a session*. A four-hour
   cookie survives closing the browser and is therefore the persistent shape that reading is
   sceptical of.

   The functional cost of the change is close to nothing, which is what settled it. Everything the
   token exists to survive — a reload, a closed tab, a sleeping laptop, a dropped connection, a lunch
   break — leaves the browser running, so a session cookie covers all of it. What is given up is only
   "quit the browser entirely, come back later", and even that is often preserved by browsers that
   restore the previous session.

   Note what this still does **not** do: the server does not check a token's age at all, so this is
   a browser-side lifetime only, and the specification says so rather than implying an expiry the
   server enforces. None of this is legal advice; it is the reasoning the owner decided on, with its
   sources named.

7. **The name field is `autocomplete="off"`, added after the owner tested it.** It said
   `nickname`, which made the browser keep its own copy of whatever had been typed and offer it
   back — so the name reappeared for somebody who had declined, and deleting our cookie did nothing
   about it. A field whose entire point is that storing is a choice cannot have the browser storing
   it regardless. This was not in the original plan; it was found by looking.

8. **Changing a name from the table opens a dialog carrying both controls**, replacing the inline
   edit at the seat. Also added after testing, and it fixes a defect against this change's own
   specification: the choice was offered at the prompt and then unreachable, so a seated person
   could not withdraw at all. Decision 2 says withdrawal must be as easy as giving; without a way
   back to the control, it was impossible. The seat had nowhere to put a second control, which is
   why the inline edit had to go rather than be extended.

9. **The privacy document is written directly, not as a spec**, and states only what can be read off
   the code: what is held in memory, for how long, what the two cookies are, and what the browser
   and any proxy will log. Every operator-dependent fact is left as a marked gap. Writing a
   plausible controller or legal basis would be inventing facts about somebody else's business,
   which is worse than an obvious blank.

## Risks / Trade-offs

- **The dialog replaces an inline edit, which is a small loss of immediacy** → accepted: renaming
  is rare, and the control it now carries has to live somewhere reachable.
- **A checkbox nobody reads is not meaningful consent** → it is unchecked, it sits next to the field
  it concerns, and it states the duration in words rather than linking to it. That is the most this
  design can do; whether it suffices legally is the owner's call and is recorded as such.
- **Somebody who declines retypes their name every session** → that is the accepted cost of the
  decision, and it is one short field.
- **Quitting the browser will cost a seat** → the consequence is written into the specification
  rather than left to be discovered, and it is the same outcome as any other lost cookie. A laptop
  closed over a long lunch is *not* this case: the browser keeps running and the cookie with it.
- **The privacy document will read as incomplete** → it is, deliberately and visibly. The gaps are
  D3's work and marking them is more useful than filling them.

## Migration Plan

Nothing to migrate: no instance has ever run outside the owner's machine, so no `pp_name` cookie
exists in the wild. A stale one on a development machine is read only when the choice is on, and is
deleted the first time it is turned off. The seat cookie's new shape applies to newly issued cookies;
any existing twelve-hour one keeps the lifetime it was given and expires on its own.
