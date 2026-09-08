## 1. The hint

- [x] 1.1 Define the sentence once in `web/src/lib/name.ts` beside `REMEMBERED_FOR`, so the two places cannot drift apart, and show it in both `NamePrompt.svelte` and `NameDialog.svelte` beside the field and above the storage choice, styled as quiet explanatory text rather than a warning.

## 2. Verification

- [x] 2.1 Run `npm run check` and `npm run build`; no Go code changes, so do not rerun the Go suites.
- [x] 2.2 Confirm in a browser against the running container that the sentence appears at the name prompt and in the rename dialog with identical wording, and that a full name is still accepted and used unchanged.
- [x] 2.3 Run strict OpenSpec validation and `git diff --check`, then sync and archive.

## Verification results (2026-09-07)

`npm run check` reports 175 files with no errors or warnings; `npm run build`
succeeds. No Go code changed, so the Go suites were deliberately not rerun.

Checked in a browser against the running container:

- The name prompt shows: *"A first name or nickname is enough — everyone with the link
  to this room can see it."*
- The rename dialog shows the **same string**, compared programmatically rather than
  by eye — they are equal, which is what defining it once in `name.ts` buys.
- A deliberately full name, "Maria-Katharina", was accepted and appears at the table
  unchanged. Nothing warned, blocked or asked again. The hint is advice, not a rule,
  and that is now demonstrated rather than merely written down.
