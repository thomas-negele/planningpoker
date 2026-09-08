## 1. Align the contract with the approved layout

- [x] 1.1 Merge the `table-ui` delta into the main specification; inspect the diff to confirm that only vertical centering and its scenario change, with the containment and stable-layout requirements retained.
- [x] 1.2 Replace the obsolete comment claiming the action is out of flow at the foot of the table; verify against the current `.table`, `.info` and `.action` CSS, without changing markup or styles.
- [x] 1.3 Run `openspec validate --all --strict` and `git diff --check`; record that no runtime code changes or visual browser acceptance are part of this specification correction.

A separate design document is unnecessary: this change records the owner's placement decision and introduces no implementation choice. Existing layout behaviour is retained.

## Verification results (2026-09-08)

Merged and checked. The diff against the main specification touches only the vertical
placement sentence and its scenario: horizontal centring, containment within the
table, and the requirement to reserve the reveal control's space are all retained
word for word.

The obsolete comment above `.action` was the larger half of this. It still described
the arrangement that was replaced — the action "lifted out of the flow and pinned near
the foot of the felt" so the chart alone landed on the table's exact middle — and gave
that as the reason for the `.table` sizing. It now describes what the code does and
says why the earlier arrangement was dropped, so the next reader does not restore it.

The current layout satisfies the loosened requirement, measured rather than assumed:
the chart is horizontally centred on the table, contained within it with 94 px above
and 42 px below, and the reveal control's slot stays at the same position before and
after revealing, so nothing moves when it is hidden.

`openspec validate --all --strict` passes for all eight items, `git diff --check` is
clean, and `npm run check` reports no errors or warnings. No runtime code changed —
the only source edit is a comment — so no browser acceptance was repeated for this
correction.
