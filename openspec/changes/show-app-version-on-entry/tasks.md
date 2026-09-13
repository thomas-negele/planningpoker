## 1. Version and start page

- [ ] 1.1 Set the initial version to `0.3.0` with `npm version 0.3.0 --no-git-tag-version` in `web/`; verify `package.json` and `package-lock.json` both contain `0.3.0`.
- [ ] 1.2 Render the version from package metadata in the entry-screen component only; verify `npm run check` and `npm run build` pass and the built application shows `Version 0.3.0` at `/` but not on a room URL.

## 2. Future change process

- [ ] 2.1 Update `CONTRIBUTING.md` and the proposal/task rules in `openspec/config.yaml` so each functional change includes a proposed major/minor/patch step, the user's choice, and a version increase before merge; verify the documents also exempt documentation-only changes and `openspec instructions` exposes the new rules.
- [ ] 2.2 Rebuild the same source revision and verify the recorded version remains `0.3.0` in both builds and the working-tree version files do not change during either build.
