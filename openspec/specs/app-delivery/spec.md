# app-delivery Specification

## Purpose

Defines how the Planning Poker application reaches a visitor's browser and what that browser is
permitted to do once it has arrived: the single origin everything is served from, the fallback that
lets client-side routes survive a hard reload, the Content-Security-Policy the browser enforces,
the availability of the real-time connection, and how the running process is configured and
stopped.

## Requirements

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

The server SHALL accept a WebSocket upgrade request at a path naming a specific room, and hold the
connection open. Establishing the connection MUST succeed both when the server is reached directly
and when it is reached through a reverse proxy that terminates TLS.

The endpoint that echoed messages back is gone. It existed only to prove the transport path before
any game logic depended on it, and it has done that: what the connection now carries is defined by
the `live-updates` capability.

What this requirement continues to guarantee is the transport itself — that the upgrade succeeds in
both deployments, and that a closed connection releases everything holding it open. Every later
change rests on those two properties, and neither is obvious enough to leave untested.

#### Scenario: Upgrade succeeds directly

- **WHEN** a browser opens a WebSocket to a room on the application's own origin
- **THEN** the connection reaches the open state and the server begins sending that room's state

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

The values governed by this rule are the network address the server listens on, the shutdown
timeout described in the requirement below, the grace period after which an abandoned room is
discarded, and the four capacity limits: the number of rooms the process will hold, the number of
connections one room will hold, the number of participants one room will seat, and the rate at
which one connection may send messages.

An unparseable or invalid value SHALL cause the process to fail at startup with a message naming
the variable and the offending value, rather than silently falling back to the default — a
misconfiguration that is silently ignored is worse than one that stops the process.

A capacity limit SHALL be refused at startup if it is zero or negative. There is deliberately no
value meaning "no limit": the whole purpose of these settings is that some ceiling exists, and an
operator who wants effectively no ceiling can say so with a large number, which is honest about
what it costs.

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

#### Scenario: Capacity limits are configurable with documented defaults

- **WHEN** the process starts with none of the capacity-limit variables set
- **THEN** it applies the documented default for each, and each default is explained in
  `compose.yaml` in full sentences saying what it bounds and what happens when it is reached

#### Scenario: A capacity limit of zero is refused

- **WHEN** the process starts with any capacity-limit variable set to zero, to a negative number, or
  to something that is not a number
- **THEN** the process exits with a non-zero status and an error naming the variable and the value

### Requirement: The process holds a bounded number of rooms and connections

The process SHALL enforce a maximum number of rooms held at one time, and a maximum number of
connections attached to one room. Both SHALL be configurable as described above.

Reaching a limit SHALL refuse only the new work. Rooms already running, connections already
attached and games already in progress SHALL be unaffected, and no participant SHALL be removed from
a seat to make room for anybody else.

A refusal SHALL tell the page why it was refused, distinguishably from a room that does not exist
and from a temporary fault, so the interface can say what is full rather than showing an empty
table or retrying forever.

The two ceilings SHALL be distinguishable **from each other**, not only from those other cases.
Reaching the process's room ceiling and reaching one room's connection ceiling are different
facts about different things, and the useful advice differs: a full server frees up on its own,
while a room holding too many connections is usually one person's spare tabs. A single reason
covering both would tell somebody on an idle server that the server is busy.

Refused work SHALL NOT leave anything behind: no room, no goroutine, no seat token and no entry in
any collection that grows. A client that is refused repeatedly SHALL NOT cost the process more with
each attempt than the refusal itself.

These limits SHALL NOT be used to end a meeting. A connected participant is never disconnected
because a limit exists, however long they stay.

#### Scenario: A room beyond the ceiling is refused, not created

- **WHEN** the process already holds its maximum number of rooms and a connection arrives for a
  room identifier that does not exist yet
- **THEN** no room is created, the connection is told the server is full, and every existing room is
  untouched

#### Scenario: A connection beyond a room's ceiling is refused

- **WHEN** a room already holds its maximum number of connections and another connection arrives for
  it
- **THEN** that connection is refused with a reason the page can display, and the connections
  already attached continue without interruption

#### Scenario: Refused work accumulates nothing

- **WHEN** many connections are refused for capacity in succession
- **THEN** the number of rooms, goroutines and issued seat tokens is the same afterwards as before,
  and memory use does not grow with the number of refusals

#### Scenario: A long meeting is never ended by a limit

- **WHEN** participants stay connected to a room far longer than any timeout in the system, without
  voting or clicking anything
- **THEN** no capacity limit disconnects them, and their room is not discarded

#### Scenario: The two ceilings are told apart

- **WHEN** a connection is refused for the process's room ceiling, and another for a single room's
  connection ceiling
- **THEN** each refusal names its own reason, and neither is reported as the other

### Requirement: The process shuts down gracefully

On receiving `SIGTERM` or `SIGINT` — the signals sent when a container is stopped or a developer
interrupts the process — the server SHALL stop accepting new connections, close the connections it
holds, and exit with status 0 within a bounded time. If shutdown does not complete within that
time, the process SHALL exit anyway rather than hang.

The configured shutdown budget SHALL be **one deadline covering the whole shutdown**, not a budget
for one part of it. It starts when the stop signal arrives and applies to ordinary HTTP requests and
to WebSocket connections together. When it expires, whatever is still open SHALL be closed by force
rather than waited for, so that the process exits comfortably inside the grace period the container
runtime allows before it kills the process outright.

A client that has stopped reading SHALL NOT be able to delay shutdown. Neither shall one that never
answers the closing handshake: the budget is what bounds the wait, and it is enforced rather than
hoped for.

Once shutting down has begun, a new connection SHALL be refused rather than accepted, so that
cleanup does not race against arrivals.

#### Scenario: Container stop is clean

- **WHEN** the container is stopped and the process receives `SIGTERM`
- **THEN** the process stops accepting new connections, closes open WebSocket connections, and
  exits with status 0 without the container runtime having to kill it

#### Scenario: A client that has stopped reading does not delay shutdown

- **WHEN** the process is stopped while a connected client is no longer reading from its socket
- **THEN** that connection is closed by force once the budget expires, and the process still exits
  with status 0 within the budget rather than being killed by the container runtime

#### Scenario: The budget covers everything, not only HTTP

- **WHEN** the process is stopped while both an ordinary HTTP request and WebSocket connections are
  in flight
- **THEN** one deadline measured from the stop signal governs both, and the process does not spend
  the budget on one and then wait indefinitely on the other

#### Scenario: A connection arriving during shutdown is refused

- **WHEN** a new connection is attempted after the stop signal has been received
- **THEN** it is refused rather than accepted, and it neither creates a room nor delays the exit

### Requirement: A connection that has stopped answering is detected and released

The server SHALL send a periodic heartbeat on every open connection and SHALL close a connection
that does not answer within a bounded time. Every outgoing message SHALL have its own deadline, so
that a client which has stopped reading cannot hold a goroutine indefinitely.

This exists because a connection can die without saying so: a lost mobile signal, a suspended
laptop, or a proxy that drops a connection without sending a close frame. Nothing arrives to read
and nothing arrives to fail on, so without a heartbeat such a connection stays open in the server's
eyes forever — keeping a seat occupied and keeping the room from ever expiring.

**The heartbeat SHALL NOT be an inactivity timeout and SHALL NOT be turned into one.** It measures
whether the network path is alive, never whether anybody is doing anything. A browser answers it
automatically, with the page untouched and nobody at the keyboard, so a meeting in which people
think for an hour without voting is unaffected. No rule anywhere may end a connection, a seat or a
room on the grounds that no game action has arrived.

Closing a connection for either reason SHALL be indistinguishable, to everyone else at the table,
from any other dropped connection: the participant is marked away, keeps their seat, their name and
their vote, and is reseated on their return.

#### Scenario: A silently dead connection becomes an away participant

- **WHEN** a participant's connection stops answering entirely, without the socket being closed
- **THEN** the server closes it within a bounded time and that participant is shown as away, exactly
  as after any other dropped connection

#### Scenario: A quiet participant is never disturbed

- **WHEN** a participant stays connected far longer than the heartbeat interval without voting,
  revealing, renaming or sending anything at all
- **THEN** their connection is answered by their browser without their involvement, it stays open,
  and neither their seat nor their room is affected

#### Scenario: A client that stopped reading does not pin a goroutine

- **WHEN** the server writes to a connection whose client has stopped reading
- **THEN** the write fails once its deadline expires and the connection is released, rather than
  waiting indefinitely

#### Scenario: A released connection frees everything it held

- **WHEN** a connection is closed for failing to answer its heartbeat
- **THEN** the goroutines serving it finish, its entry in the room is removed, and nothing it held
  is left behind

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

The application SHALL be built into a single container image containing a statically linked
application binary, no shell and no package manager, and SHALL run as an unprivileged user. The
base deployment configuration SHALL NOT publish the application's port on the host, so that a
reverse proxy reaches it over the container network by service name. An explicit local
configuration SHALL publish the application only on the host's IPv4 loopback interface.

The project SHALL NOT supply a reverse proxy or manage certificates. Whoever self-hosts this
already has a proxy and it already manages their certificates; a second one shipped here would be a
deployment path nobody exercises. What the project owes them instead is a plain statement of what
the application requires of a proxy — see the requirement below.

#### Scenario: One command builds and runs the application

- **WHEN** `docker compose up --build` is run in a clean checkout
- **THEN** the frontend is built, the binary is compiled, and the container serves the application
  on its container port without publishing it on the host

#### Scenario: Container runs unprivileged

- **WHEN** the running container's user is inspected
- **THEN** it is not `root`

#### Scenario: Server example reaches the application through HTTPS

- **WHEN** the operator puts the application behind their own reverse proxy, configured as the
  documentation describes, with a valid domain and the documented DNS and public port prerequisites
- **THEN** the proxy serves the application over HTTPS, its WebSocket connection works over WSS,
  and the application's own port remains unpublished

#### Scenario: Local Docker access is explicit and restricted

- **WHEN** the operator starts the supplied local configuration
- **THEN** the application is reachable through the documented loopback HTTP address and its
  published port is bound only to `127.0.0.1`

### Requirement: The supplied application service has bounded resources and restricted privileges

The supplied application service SHALL operate with a read-only root filesystem, no Linux
capabilities and no ability to gain privileges. It SHALL NOT require host-directory or
Docker-socket mounts. The supplied local configuration SHALL retain these restrictions.

The service SHALL have positive, operator-adjustable memory, CPU and process limits and bounded
container log retention. Defaults SHALL be documented as small-deployment starting points rather
than capacity guarantees. These deployment limits SHALL NOT introduce an application inactivity
timeout or alter room access rules.

#### Scenario: Restricted service supports a planning session

- **WHEN** the supplied application service runs with its default restrictions
- **THEN** it serves the frontend and allows participants to join, vote and reveal without
  filesystem writes or elevated privileges

#### Scenario: Resource and log bounds are effective

- **WHEN** the resolved configuration and running container are inspected
- **THEN** positive memory, CPU and process limits are present and log files have configured size
  and file-count bounds

#### Scenario: Operators can adjust resource budgets

- **WHEN** an operator supplies documented resource-budget overrides
- **THEN** the resolved application service uses those budgets while retaining its privilege
  restrictions

### Requirement: The application states what it requires of a reverse proxy

The application SHALL be usable behind any reverse proxy that terminates TLS, and the project SHALL
document what such a proxy must do. Three things are required, and each SHALL be stated where an
operator will read it rather than left to be discovered:

1. **The WebSocket upgrade must be forwarded.** Without it there is no connection at all.
2. **The `Host` header the browser sent must be preserved.** The server compares the browser's
   `Origin` against `Host` and refuses an upgrade whose two disagree. A proxy that rewrites `Host`
   to the backend's name therefore breaks every connection.
3. **`X-Forwarded-Proto: https` must be set** when the proxy terminates TLS, because the server
   cannot otherwise know that the visitor's connection is encrypted and will issue the seat cookie
   without `Secure`.

The second is the one that costs an evening if it is not written down: the failure looks like a
broken application rather than a proxy setting, and nothing on screen suggests otherwise.

Each requirement SHALL be observable, so that an operator can check their own proxy rather than
trust a description.

#### Scenario: A proxy that rewrites the Host header is refused

- **WHEN** an upgrade arrives carrying a browser's `Origin` for the public name while `Host` has
  been rewritten to the backend's name
- **THEN** the upgrade is refused rather than silently accepted, so the misconfiguration surfaces
  immediately

#### Scenario: A proxy that preserves the Host header works

- **WHEN** the same upgrade arrives with `Host` as the browser sent it
- **THEN** the connection is established

#### Scenario: The forwarded protocol decides the cookie

- **WHEN** an upgrade arrives with `X-Forwarded-Proto: https`, and again without it
- **THEN** the seat cookie carries `Secure` in the first case and not in the second

#### Scenario: A request that is not from a browser is not refused for this

- **WHEN** an upgrade arrives with no `Origin` header at all
- **THEN** it is accepted, because the comparison guards browsers acting on some page's behalf and
  is not what stops a program written by whoever ran it

### Requirement: Application provides a browser tab icon

The application SHALL reference a card-themed favicon from its HTML document and serve
the icon from the application origin, in development and production.

#### Scenario: Entry and room pages reference the icon

- **WHEN** a visitor opens the entry page or a room URL
- **THEN** the HTML document references the same application favicon
- **AND** requesting that icon returns image content from the application origin
