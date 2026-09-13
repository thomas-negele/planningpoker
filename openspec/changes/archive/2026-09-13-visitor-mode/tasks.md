## 1. Server behavior

- [x] 1.1 Add visitor mode to a seat and the room snapshot, defaulting new seats to voter and preserving a returning seat's mode; verify game tests for joining, rejoining, and snapshot state.
- [x] 1.2 Reject visitor votes, exclude present visitors from the everyone-voted calculation, and delete a hidden vote when Save switches a voter to visitor; verify game tests for refusal, zero eligible voters, switching back, and unchanged revealed cards/tallies.
- [x] 1.3 Keep Reveal, New round, deck selection, rename, and throws available to visitors under their existing rules; verify server tests exercise visitor actions.

## 2. Protocol and browser controls

- [x] 2.1 Extend seat and rename messages with optional visitor mode, apply name and mode changes atomically on rename, publish visitor state in snapshots, and add a specific vote-refusal code; verify transport tests cover omitted fields, false/true values, invalid names, and a direct visitor vote.
- [x] 2.2 Add the unchecked Visitor mode checkbox and keyboard-accessible short information hint to the join form; verify joining with and without the checkbox produces the selected seat mode.
- [x] 2.3 Add the mode checkbox to the existing own-name dialog and send it only on Save; verify Cancel leaves the mode and hidden vote intact, while Save changes the mode at any round stage.
- [x] 2.4 Show a Visitor text label at visitor seats without an empty card symbol, retain a card already revealed for that round, hide the visitor voting deck, and show "No voters this round" when appropriate; verify the table in wide and narrow layouts and confirm Reveal and New round remain usable.

## 3. Integration and release

- [x] 3.1 Run the Go and web checks and verify a two-browser round: hidden vote deletion on switching to visitor, no restoration on switching back, preserved revealed results, and role persistence after reload.
- [x] 3.2 Before merge into main, obtain the owner's major/minor/patch choice, apply that chosen version increase, and verify web/package.json and web/package-lock.json contain the same version.
