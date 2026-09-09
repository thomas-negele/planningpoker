## MODIFIED Requirements

### Requirement: Client-side routes survive a hard reload

The server SHALL return the application's `index.html` document, with HTTP status 200, for any
frontend request path that does not correspond to a file in the built frontend. API, WebSocket
and `/legal` routes are handled separately; an unavailable resource under `/legal` SHALL return
404 instead of this fallback. This is what makes a
client-side route such as `/g/<room-id>` work when it is opened directly or reloaded, rather than
returning "not found".

A request for a path that is clearly meant to be a static asset but does not exist SHALL return
HTTP status 404 rather than the HTML document, so that a mistyped or stale asset reference fails
visibly instead of silently delivering HTML where JavaScript or CSS was expected.

#### Scenario: Unknown application path returns the document

- **WHEN** a browser requests a frontend route that matches no file in the built frontend, for
  example `/g/abc123`
- **THEN** the server responds with status 200 and the content of `index.html`

#### Scenario: Existing asset is served as itself

- **WHEN** a browser requests a path that does match a file in the built frontend, for example the
  bundled JavaScript file
- **THEN** the server responds with that file's content and its correct content type, not with
  `index.html`

#### Scenario: Missing asset fails visibly

- **WHEN** a browser requests a non-existent file inside the built asset directory
- **THEN** the server responds with status 404 and does not return `index.html`
