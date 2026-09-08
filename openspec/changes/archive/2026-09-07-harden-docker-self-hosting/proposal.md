## Why

The current container is a suitable small deployment unit, but a fresh checkout has no browser-accessible Docker quickstart or complete HTTPS example. The supplied service also lacks explicit resource and runtime restrictions needed for a public hobby deployment.

## What Changes

- Keep the base Compose service unpublished and add explicit local-loopback and HTTPS-proxy deployment examples.
- Apply nonroot, read-only filesystem, dropped capabilities, no-new-privileges, resource limits and bounded logs to the app service.
- Preserve one in-memory app instance, open room links and unlimited meeting duration while connected.
- Write the general README outside the specification workflow, using the resulting commands; exclude operator-specific infrastructure.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `app-delivery`: Distinguish unpublished server deployment from explicit loopback-only local publishing, and require bounded, restricted app-container execution.

## Impact

Compose files, a small Caddy example, Docker-related comments and ignore rules. No Go/Svelte behavior or protocol changes, no new application dependencies. Answers the review findings about a missing Docker quickstart, a missing HTTPS example and absent container restrictions, and supports the reproducible-build and documentation work that follows. Resource defaults are proposed in design.md and remain operator-adjustable. README prose is a direct documentation task, not a new product capability.
