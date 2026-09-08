## Why

The name somebody types is shown to everyone who has the room's link, and that link is deliberately
the only thing guarding a room. Nothing at the moment of typing says so, so the obvious thing to
enter is a full name — which is then visible to anyone holding or guessing the link, including
before they give a name themselves.

`PRIVACY.md` already says to treat a room link like the meeting invitation, but nobody reads a
privacy document while filling in a form. The one place this can be said usefully is the field
itself.

## What Changes

- Add one line where a name is entered, saying that a first name or nickname is enough and why:
  everyone with the link can see it.
- Say it identically in both places a name is entered — the name prompt and the dialog for changing
  it at the table.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `table-ui`: The name prompt and the rename dialog state that a short name suffices and that
  everyone with the link sees it.

## Impact

`web/src/components/NamePrompt.svelte` and `web/src/components/NameDialog.svelte`. No server change,
no protocol change, no new storage.

This is a hint, **not** a rule: no name is refused for being too full, nothing is checked, and the
game behaves identically whatever somebody types. Deliberately out of scope: any warning dialog,
any attempt to detect a real name, and any change to the open-link decision itself, which stands.
