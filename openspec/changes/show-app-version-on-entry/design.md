## Context

`web/package.json` and `web/package-lock.json` currently record `0.0.0`. Vite builds the frontend, and Docker embeds that output in the Go binary. There is no CI or release automation. The entry screen is a separate Svelte component; the shared footer also appears in rooms.

## Goals / Non-Goals

**Goals:** Use one checked-in version value, keep repeat builds stable, and make the version decision part of future OpenSpec changes.

**Non-Goals:** Build counters, Git tags, automatic publication, or a new endpoint.

## Decisions

1. Use `web/package.json` as the version source. Set the first visible value to the user-selected `0.3.0`. Use `npm version <major|minor|patch> --no-git-tag-version` for later user-approved steps so `package.json` and `package-lock.json` stay aligned. A separate version file would duplicate existing metadata; a build-time counter would change the version on identical rebuilds.
2. Read that version into the entry-screen component through Vite's existing JSON import support. The production bundle then contains the source version without server changes. Render it in the entry screen rather than the shared footer, which also appears in rooms.
3. Add a required version-decision step to `CONTRIBUTING.md` and the OpenSpec artifact rules in `openspec/config.yaml`. For each functional change, Codex proposes major, minor or patch, the user chooses, and the implementation tasks include the selected increase before merge. Documentation-only changes are exempt. This is a human process requirement; the user chose not to add a technical gate.

## Risks / Trade-offs

- A manual step can be missed → Make it explicit in both the contribution workflow and OpenSpec planning instructions; check the recorded version during change verification.
- Package and lockfile versions can diverge → Use `npm version ... --no-git-tag-version`, which updates both.

## Migration Plan

Set the checked-in version to `0.3.0` when implementing this change and deploy the normal rebuilt image. Existing deployments retain their old code until rebuilt and restarted. Rolling back to an earlier commit and rebuilding restores that commit's version.
