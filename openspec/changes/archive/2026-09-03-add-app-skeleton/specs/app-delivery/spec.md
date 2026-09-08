## Purpose

Defines how the Planning Poker application reaches a visitor's browser and what that browser is
permitted to do once it has arrived: the single origin everything is served from, the fallback that
lets client-side routes survive a hard reload, the Content-Security-Policy the browser enforces,
the availability of the real-time connection, and how the running process is configured and
stopped.

## ADDED Requirements

### Requirement: Application is served from a single origin

The application SHALL serve the built frontend — the HTML document and every stylesheet, script,
font and image it references — from the same host that serves the application itself. The built
output MUST NOT contain any absolute reference to a foreign host, and loading the application MUST
NOT cause the browser to issue a request to any origin other than the one it was loaded from.

#### Scenario: Built output contains no foreign origin

- **WHEN** the built frontend output is searched for the strings `http://` and `https://`
- **THEN** no match refers to a host other than the one serving the application

#### Scenario: Browser contacts one origin only

- **WHEN** a visitor loads the application and the browser's network activity is inspected
- **THEN** every request, including the WebSocket connection, targets the origin the page was
  loaded from, and no request targets a font service, a content delivery network, an analytics
  endpoint, or any other external host

### Requirement: Client-side routes survive a hard reload

The server SHALL return the application's `index.html` document, with HTTP status 200, for any
request path that does not correspond to a file in the built frontend. This is what makes a
client-side route such as `/g/<room-id>` work when it is opened directly or reloaded, rather than
returning "not found".

A request for a path that is clearly meant to be a static asset but does not exist SHALL return
HTTP status 404 rather than the HTML document, so that a mistyped or stale asset reference fails
visibly instead of silently delivering HTML where JavaScript or CSS was expected.

#### Scenario: Unknown application path returns the document

- **WHEN** a browser requests a path that matches no file in the built frontend, for example
  `/g/abc123`
- **THEN** the server responds with status 200 and the content of `index.html`

#### Scenario: Existing asset is served as itself

- **WHEN** a browser requests a path that does match a file in the built frontend, for example the
  bundled JavaScript file
- **THEN** the server responds with that file's content and its correct content type, not with
  `index.html`

#### Scenario: Missing asset fails visibly

- **WHEN** a browser requests a non-existent file inside the built asset directory
- **THEN** the server responds with status 404 and does not return `index.html`

### Requirement: Browser enforces the self-contained rule via Content-Security-Policy

Every response that serves the application in a production build SHALL carry a
`Content-Security-Policy` header with the following directives:

```
default-src 'self';
connect-src 'self';
img-src 'self' data:;
font-src 'self';
style-src 'self';
script-src 'self';
object-src 'none';
base-uri 'self';
form-action 'self';
frame-ancestors 'none'
```

The policy MUST NOT contain `'unsafe-inline'` or `'unsafe-eval'`. If a violation is reported by the
browser, the code that caused it is to be corrected; the policy is not to be widened.

The header SHALL NOT be applied in the development build, because the Vite development server
relies on inline scripts and its own WebSocket for hot reloading, which this policy forbids.

#### Scenario: Production response carries the policy

- **WHEN** any application response is served by a production build — the HTML document or a
  static asset
- **THEN** the response includes the `Content-Security-Policy` header with exactly the directives
  listed above

#### Scenario: Application runs without a single violation

- **WHEN** a visitor loads the application in a production build and opens the browser console
- **THEN** no Content-Security-Policy violation is reported, including none caused by the
  WebSocket connection

#### Scenario: Development build omits the policy

- **WHEN** the application is run in the development build with the Vite development server
- **THEN** no `Content-Security-Policy` header is applied and hot reloading works

### Requirement: A WebSocket connection can be established

The server SHALL accept a WebSocket upgrade request at a fixed path and hold the connection open.
Establishing the connection MUST succeed both when the server is reached directly and when it is
reached through a reverse proxy that terminates TLS.

This requirement exists to prove the transport path end to end before any game logic depends on it.
Until the real protocol is specified, the endpoint echoes back any message it receives and carries
no application meaning.

#### Scenario: Upgrade succeeds directly

- **WHEN** the placeholder page opens a WebSocket to the application's own origin
- **THEN** the connection reaches the open state, and a message sent by the client comes back
  unchanged

#### Scenario: Upgrade succeeds through a TLS-terminating reverse proxy

- **WHEN** the application is reached through the reverse proxy over `https`, so that the browser
  opens a `wss` connection
- **THEN** the connection reaches the open state and is not blocked by the Content-Security-Policy

#### Scenario: Connection closes cleanly

- **WHEN** the client closes the WebSocket, or the browser tab is closed
- **THEN** the server releases the connection and the goroutines serving it, without leaking them

### Requirement: Behaviour-governing values are set by environment variable

Every value that governs how the running process behaves SHALL be readable from an environment
variable and SHALL have a documented default that applies when the variable is unset or empty. Each
such value MUST be accompanied, in the code and in `compose.yaml`, by a full-sentence explanation
of what it does and what a boundary value means.

For this change there are two such values: the network address the server listens on, and the
shutdown timeout described in the requirement below.

An unparseable or invalid value SHALL cause the process to fail at startup with a message naming
the variable and the offending value, rather than silently falling back to the default — a
misconfiguration that is silently ignored is worse than one that stops the process.

#### Scenario: Default applies when unset

- **WHEN** the process starts with no listen-address variable set
- **THEN** it listens on the documented default address and logs which address it is listening on

#### Scenario: Environment variable overrides the default

- **WHEN** the process starts with the listen-address variable set to a valid address
- **THEN** it listens on that address

#### Scenario: Invalid value stops the process

- **WHEN** the process starts with the listen-address variable set to a value that cannot be used
- **THEN** the process exits with a non-zero status and an error message naming the variable and
  the value

#### Scenario: Shutdown timeout is configurable with a documented default

- **WHEN** the process starts with no shutdown-timeout variable set
- **THEN** it uses the documented default, and that default is shorter than the grace period the
  container runtime allows before it kills the process outright

#### Scenario: Ambiguous shutdown timeout is refused

- **WHEN** the process starts with the shutdown-timeout variable set to zero, to a negative
  duration, or to something that is not a duration at all
- **THEN** the process exits with a non-zero status and an error naming the variable and the value,
  because a boundary value with no agreed meaning is worse than no setting at all

### Requirement: The process shuts down gracefully

On receiving `SIGTERM` or `SIGINT` — the signals sent when a container is stopped or a developer
interrupts the process — the server SHALL stop accepting new connections, close the connections it
holds, and exit with status 0 within a bounded time. If shutdown does not complete within that
time, the process SHALL exit anyway rather than hang.

#### Scenario: Container stop is clean

- **WHEN** the container is stopped and the process receives `SIGTERM`
- **THEN** the process stops accepting new connections, closes open WebSocket connections, and
  exits with status 0 without the container runtime having to kill it

### Requirement: Frontend changes are visible without recompiling the binary in development

In the development build the frontend SHALL be served such that a change to a frontend source file
becomes visible in the browser without rebuilding the Go binary. In the production build the
frontend SHALL be served from files compiled into the binary, so that the binary is self-sufficient
and needs no accompanying directory of assets.

#### Scenario: Development serves current frontend sources

- **WHEN** a developer edits a frontend source file while the development setup is running
- **THEN** the change is visible in the browser without the Go binary being rebuilt

#### Scenario: Production binary is self-sufficient

- **WHEN** the production binary is run in a location containing no frontend files at all
- **THEN** it still serves the complete application

### Requirement: The application ships as one container

The application SHALL be built into a single container image containing a statically linked binary
and nothing else — no shell and no package manager — running as an unprivileged user. The image
SHALL NOT publish a port on the host; it is reached over the container network by service name, by
the reverse proxy in front of it.

#### Scenario: One command builds and runs the application

- **WHEN** `docker compose up --build` is run in a clean checkout
- **THEN** the frontend is built, the binary is compiled, and the container serves the application
  on its container port

#### Scenario: Container runs unprivileged

- **WHEN** the running container's user is inspected
- **THEN** it is not `root`
