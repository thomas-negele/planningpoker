## Why

Some people need to follow and facilitate an estimation round without casting a vote. Today every seated person is treated as a voter, so the table and the round-completion indicator cannot represent that role clearly.

## What Changes

- Add an unchecked "Visitor mode" checkbox to the room-join name form, with a short, accessible information hint that visitors cannot vote.
- Let a seated person change this setting at any time in the existing name dialog. The change takes effect on Save. Visitors keep the same seat and all non-voting actions, including Reveal and New round.
- Show visitors at the table with a compact "Visitor" label instead of an empty card slot. Hide their voting deck. Keep a card and tally already revealed for the current round visible even if its owner switches modes afterwards.
- When someone enables visitor mode during a hidden round, permanently delete their current vote. Disabling it does not restore that vote; they may vote again while the round remains hidden. Exclude present visitors from the everyone-voted indicator and show a clear no-voters message when appropriate.
- Make visitor state and vote restrictions server-authoritative and preserve the setting across reconnection.
- Propose a **minor** version increase from `0.3.0` to `0.4.0` for this new feature. The owner chooses major, minor, or patch before the version is changed.

## Capabilities

### New Capabilities

- `visitor-mode`: Joining, switching, and retaining a non-voting seat; visitor permissions; effects on hidden and revealed votes.

### Modified Capabilities

- `estimation-rounds`: Voting eligibility and the everyone-voted indicator exclude visitors.
- `table-ui`: Join and name dialog controls, visitor seat/deck presentation, and no-voters status.
- `live-updates`: The visitor setting becomes an intent and part of the authoritative room snapshot.

## Impact

The room model and view, WebSocket protocol and validation, and Svelte join form, name dialog, table, and voting controls will change. No new dependency is expected.
