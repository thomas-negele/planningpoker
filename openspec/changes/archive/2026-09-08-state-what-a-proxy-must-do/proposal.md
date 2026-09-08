## Why

The repository ships a working Caddy configuration that terminates TLS and obtains certificates.
That claims a job this project has no business doing: whoever self-hosts this already has a reverse
proxy, and it already manages their certificates. The example is not what the author runs either, so
it is a second deployment path that nobody exercises and nobody will maintain.

What is missing matters more than what is shipped. **Two requirements the application places on any
reverse proxy are documented nowhere**, and both fail in ways that look like a broken application
rather than a proxy setting:

- A proxy that rewrites the `Host` header breaks every WebSocket connection. The server compares the
  browser's `Origin` against `Host` and refuses the upgrade when they disagree — observed as
  `426 Upgrade Required`, with nothing on screen to suggest the proxy is at fault.
- A proxy that does not set `X-Forwarded-Proto: https` causes the seat cookie to be issued without
  `Secure`, even though the connection is encrypted.

## What Changes

- Remove the supplied Caddy configuration: `compose.server.yaml` and `Caddyfile`, and the README
  section that walks through certificates and DNS.
- Replace it with what any proxy must do, and how to connect one to the container — a shared Docker
  network, or the loopback-published port.
- Stop promising a supplied server configuration in the specification, and state instead what the
  application requires of a proxy, as behaviour that can be observed.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `app-delivery`: The container requirement no longer promises a supplied server configuration; a
  requirement is added for what the application needs from a reverse proxy and how it behaves when
  those needs are not met.

## Impact

Deletes `compose.server.yaml` and `Caddyfile`. `README.md` loses the HTTPS walkthrough and gains a
proxy-requirements section; its update and rollback commands lose `PLANNINGPOKER_DOMAIN`.
`PRIVACY.md` mentions the Caddyfile in its logging note and needs adjusting.

No Go or Svelte code changes: the behaviour described is what the application already does, verified
by probing the running server. `compose.yaml` keeps its unpublished port and `compose.local.yaml`
keeps its loopback publishing — both are still correct and both are still what a proxy attaches to.
