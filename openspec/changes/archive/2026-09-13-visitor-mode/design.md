## Context

See proposal.md for the behavior change. Rooms already keep seat and vote state on the server and send full snapshots over WebSocket. The join form sends a seat intent; the name dialog sends a rename intent on Save. The room-settings gear only controls the deck and can be locked during a hidden vote, so it cannot host a mode switch that must always be available.

## Goals / Non-Goals

**Goals:**
- Keep visitor mode as one authoritative property of a seat across reconnections and tabs.
- Apply a name and mode edit together on Save, with no partial change if the name is invalid.
- Preserve a revealed round's card and tally when mode changes, while removing a hidden vote on entry to visitor mode.

**Non-Goals:**
- A separate observer connection outside the room's seat or capacity rules.
- A special host or facilitator permission model.
- Persistence of rooms beyond the existing room lifetime.

## Decisions

1. Add a boolean visitor flag to the room's participant, public participant snapshot, and corresponding wire and browser types. New seats default to voter when the seat intent omits the flag; an existing seat's mode remains unchanged on rejoin. This extends the current seat model instead of introducing a second participant type. The server refuses votes from visitors and excludes present visitors from the everyone-voted calculation; it does not restrict other seated actions.

2. Let the existing rename intent carry an optional visitor value. The name dialog sends the current name and chosen mode together on Save; the server validates the name before changing either field. Omission preserves the current mode so older rename clients cannot accidentally switch a visitor off. This avoids two separate intents whose updates could interleave or partially succeed. The initial seat intent carries the chosen visitor boolean for a new seat.

3. On a saved transition from voter to visitor, delete a hidden vote in the room's current vote map. Do not delete anything after reveal; the existing revealed results derive from that map and must remain final. New round continues to clear the map. The next full snapshot clears any locally remembered hidden card when the server reports no vote, so switching back cannot make an old selection appear restored.

4. Put the checkbox in the name prompt and existing own-name dialog. Show a small, keyboard-accessible information control with the concise text "Visitors cannot vote." The dialog checkbox is a draft until Save; Cancel discards it. Render a text "Visitor" label where an empty card slot would otherwise be, and retain a revealed card beside that label for the rest of its round. Hide the visitor's voting deck, but leave Reveal, New round, deck settings, and other eligible controls in place. The no-voters status is derived from the snapshot's present, non-visitor participants.

## Risks / Trade-offs

- A seat may be opened in more than one tab, so one tab can change its mode while another has a stale dialog open. The server's latest accepted Save wins and broadcasts a snapshot.
- Revealed results can show a card beside a current visitor label. This reflects a vote cast before reveal; retaining it protects the round's finality.
- Adding an optional boolean to intents requires distinguishing an omitted value from false. Parse it as an optional field so old clients preserve an existing seat's mode.

## Migration Plan

No persistent data migration is needed because rooms are in memory. Deploy the server and browser changes together. A seat or rename intent without the new optional field keeps the compatible default/preserved behavior; new snapshots add a boolean field. Rollback uses the previous build under the existing room-lifetime behavior.
