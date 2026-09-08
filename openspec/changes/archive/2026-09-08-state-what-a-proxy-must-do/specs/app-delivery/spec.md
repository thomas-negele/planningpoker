## MODIFIED Requirements

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

## ADDED Requirements

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
