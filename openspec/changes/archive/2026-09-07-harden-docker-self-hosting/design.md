## Context

See proposal.md for motivation. The existing distroless nonroot image embeds its frontend and needs no writable application data. The base Compose service exposes port 8080 only inside its network. A single process owns all room state. Keep the existing five-second shutdown setting and five-minute abandoned-room grace period unchanged in this package.

## Goals / Non-Goals

**Goals:** Reuse one app service for local and HTTPS deployments, add low-cost runtime restrictions and document exact working commands.

**Non-Goals:** Host firewall automation, any operator's private infrastructure, authentication, persistence, a monitoring stack, changes to WebSocket shutdown or abuse controls. Those last two remain separate checklist packages.

## Decisions

1. Keep `compose.yaml` as the common unpublished app service. Add `compose.local.yaml` with a loopback-only port mapping (default host port 8080), and `compose.server.yaml` with a Caddy service, public ports 80/443 and persistent certificate storage. Use explicit `-f` combinations; do not auto-load an override that could accidentally publish a server port. A proxy-free base remains usable with an existing reverse proxy.
2. Restrict the app with `read_only: true`, `cap_drop: [ALL]` and `security_opt: [no-new-privileges:true]`; retain the image's nonroot user and default seccomp profile. No tmpfs unless the actual runtime needs one. The proxy's certificate storage is separate from the app's read-only filesystem.
3. Proposed adjustable app defaults: memory `256m`, CPU `1.0`, PID limit `128`, and JSON logs capped at `10m` per file with `3` files. Explain each value and adjustment through Compose environment interpolation. These are conservative starting budgets, not benchmarked promises. Prefer these few knobs over an administration interface. Never enforce meeting inactivity through them.
4. Put app and proxy on a dedicated bridge network in the server example; evaluate an internal app network with a separate external proxy network during implementation. No host networking, Docker socket or unrelated services. Document generic installation responsibilities rather than writing host firewall rules. The app has no runtime outbound API dependency; builds require package registries.
5. A small example Caddyfile takes an operator-supplied domain, forwards to `planningpoker:8080` and preserves the HTTPS origin/forwarding information needed for WSS and Secure cookies. Resolve a maintained image tag against official documentation during implementation and describe tags honestly; do not claim immutable pinning without digests.
6. README is written directly alongside implementation, outside the OpenSpec artifact contract: local/server commands, three existing app settings, resource overrides, logs, stop/update, loss of rooms on restart, open room links, and spec navigation. Exclude local editor/agent state and environment secrets from the Docker build context. Do not remove already tracked files or change licensing in this package.

## Risks / Trade-offs

- Small memory/PID budgets may not fit larger meetings → keep them adjustable and validate a representative small session once; no heavy load benchmark here.
- Compose merges can accidentally retain published ports → inspect base, local and server resolved configurations in one batch; local and server overlays are alternatives.
- Public TLS needs a real domain and reachable ports → validate the generic configuration and proxy route locally, and explicitly record any public certificate check left for deployment. Do not mark an unperformed HTTPS check as passed.
- Known Go abuse/shutdown issues remain → README status must not imply that this package alone grants public-service readiness.

## Migration Plan

Existing base usage remains unpublished. Select either local or server overlay when desired. Deploy outside a meeting; restart loses room state. Roll back to the prior tested image/configuration if necessary. After implementation and its bundled checks, sync/archive this change before starting another change to app-delivery.
