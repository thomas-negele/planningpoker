# operator-legal-pages Specification

## Purpose

Lets an operator publish this installation's own privacy notice and imprint from two
local HTML files, without their details entering the repository or the image.

## Requirements

### Requirement: Legal notices are opt-in and loaded at startup

The application SHALL leave notices disabled when `PLANNINGPOKER_LEGAL_DIR` is unset or
empty. When it is set, the directory SHALL supply both `privacy.html` and `imprint.html`
as nonempty, regular, UTF-8 files of at most 1 MiB. Both SHALL be read before the server
accepts requests and retained for the process lifetime.

#### Scenario: A default installation needs no documents

- **WHEN** the application starts with no legal directory configured
- **THEN** it starts without reading any operator document
- **AND** no privacy or legal-notice link appears in the app

#### Scenario: An unusable configuration stops the process

- **WHEN** the directory or either document is missing, unreadable, empty, a symlink,
  not a regular file, not valid UTF-8, or larger than 1 MiB
- **THEN** startup fails with an error naming the setting and the affected file
- **AND** the error does not include the document's contents

#### Scenario: Editing the text needs a restart, not a rebuild

- **WHEN** an operator edits a document while the process is running
- **THEN** the running process keeps serving its loaded copy
- **AND** a restart serves the edited text without the image being rebuilt

### Requirement: Only the three named legal resources are served

When enabled, the application SHALL serve the documents at `/legal/privacy` and
`/legal/imprint` as UTF-8 HTML, plus a built-in stylesheet at `/legal/style.css`.
Notices SHALL be readable without JavaScript, a seat or a room, and SHALL set no cookie.
Notice and availability responses SHALL prevent caching. Legal responses SHALL prohibit
scripts and resources from other hosts, without weakening the application's own policy.

#### Scenario: A visitor reads a notice without playing

- **WHEN** a visitor requests either notice
- **THEN** it is returned with status 200, without creating a room or setting a cookie
- **AND** rendering it requires no connection to another host

#### Scenario: An unavailable legal resource returns 404

- **WHEN** a notice is requested while the feature is disabled, or any other path under
  `/legal/` is requested
- **THEN** the server returns 404 rather than the application's HTML document

#### Scenario: Nothing else in the directory is reachable

- **WHEN** a request names another file, the directory itself, or a path traversing out of it
- **THEN** nothing but the two notices and the built-in stylesheet is disclosed

### Requirement: Notices are reachable from every screen

When enabled, every screen SHALL offer keyboard-accessible links labelled `Privacy` and
`Legal notice`, which open without disturbing the current room or its connection. The
frontend SHALL learn whether notices exist from the running backend rather than from a
build-time setting.

#### Scenario: A seated participant opens a notice

- **WHEN** a participant follows either link
- **THEN** it opens in a separate browsing context
- **AND** their room connection and played card are unaffected

#### Scenario: Availability cannot be read

- **WHEN** the frontend cannot get a valid availability response
- **THEN** it shows a short unavailable message and the game is otherwise unaffected
