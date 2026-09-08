## MODIFIED Requirements

### Requirement: The application ships as one container

The application SHALL be built into a single container image containing a statically linked application binary, no shell and no package manager, and SHALL run as an unprivileged user. The base and server deployment configurations SHALL NOT publish the application's port on the host; a reverse proxy reaches it over the container network by service name. An explicit local configuration SHALL publish the application only on the host's IPv4 loopback interface.

#### Scenario: One command builds and runs the application

- **WHEN** `docker compose up --build` is run in a clean checkout
- **THEN** the frontend is built, the binary is compiled, and the container serves the application on its container port without publishing it on the host

#### Scenario: Container runs unprivileged

- **WHEN** the running container's user is inspected
- **THEN** it is not `root`

#### Scenario: Local Docker access is explicit and restricted

- **WHEN** the operator starts the supplied local configuration
- **THEN** the application is reachable through the documented loopback HTTP address and its published port is bound only to `127.0.0.1`

#### Scenario: Server example reaches the application through HTTPS

- **WHEN** the operator starts the supplied server configuration with a valid domain and the documented DNS and public port prerequisites
- **THEN** the proxy serves the application over HTTPS, its WebSocket connection works over WSS, and the application's own port remains unpublished

## ADDED Requirements

### Requirement: The supplied application service has bounded resources and restricted privileges

The supplied application service SHALL operate with a read-only root filesystem, no Linux capabilities and no ability to gain privileges. It SHALL NOT require host-directory or Docker-socket mounts. The supplied local and server configurations SHALL retain these restrictions.

The service SHALL have positive, operator-adjustable memory, CPU and process limits and bounded container log retention. Defaults SHALL be documented as small-deployment starting points rather than capacity guarantees. These deployment limits SHALL NOT introduce an application inactivity timeout or alter room access rules.

#### Scenario: Restricted service supports a planning session

- **WHEN** the supplied application service runs with its default restrictions
- **THEN** it serves the frontend and allows participants to join, vote and reveal without filesystem writes or elevated privileges

#### Scenario: Resource and log bounds are effective

- **WHEN** the resolved configuration and running container are inspected
- **THEN** positive memory, CPU and process limits are present and log files have configured size and file-count bounds

#### Scenario: Operators can adjust resource budgets

- **WHEN** an operator supplies documented resource-budget overrides
- **THEN** the resolved application service uses those budgets while retaining its privilege restrictions
