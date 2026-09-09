## Context

The Go process serves the embedded frontend and falls back to `index.html` for unknown
paths. There is no place for operator-supplied content.

## Goals / Non-Goals

**Goals:** let an operator supply two documents at run time without rebuilding the image,
and leave a default installation untouched.

**Non-Goals:** Markdown rendering, uploads, an editor, live file watching, or any opinion
about what the documents must say.

## Decisions

### Read both documents once, at startup

Reading at startup rather than per request keeps the filesystem out of the request path
and makes it impossible for the two pages to come from different edits. The cost is that
a text change needs a restart, which ends running games; the README says so. Each file
must be a regular non-symlink file, nonempty, valid UTF-8 and at most 1 MiB — a bound so
a mistyped path fails loudly instead of loading something huge. Either file being
unusable stops the process, because half a set of notices means advertising a link that
fails. Errors name the setting and the file, never the contents, since they go to a log.

### A separate package, and a reserved subtree

`internal/legal` holds the loading and the serving; the router only wires it. Its routes
are three fixed strings compared literally, so there is no path-to-file mapping to
attack and nothing else in the operator's directory is reachable.

`/legal` is registered whether or not notices exist. Without that, a request for a notice
would reach the SPA fallback and be answered with the game — a visitor could not tell an
absent notice from a working one.

Legal responses carry their own stricter policy in **both** build modes, applied in the
handler rather than the build-tagged middleware. This is the only markup in the app that
was not written here, and the development build applies no policy at all. The application
policy is untouched.

### The frontend asks the backend

The same built frontend ships to every installation, so activation cannot be a build-time
value. `GET /api/legal` answers `{"enabled":bool}` and nothing else. The footer opens
notices with `target="_blank"`, so following one does not tear down the room's socket.

## Risks / Trade-offs

- Operator HTML could contain scripts or remote references: the policy blocks both.
- A restart to change text ends games: documented, done between meetings.
