## Context

See proposal.md. Two facts were established by probing the running server rather than by reading the
code, and they are what the documentation will assert:

- An upgrade carrying `Origin: https://poker.example.com` with `Host: planningpoker:8080` — a proxy
  that rewrote the host — is answered **426 Upgrade Required**. The same upgrade with the host
  preserved is answered **101 Switching Protocols**.
- Without `X-Forwarded-Proto: https` the seat cookie arrives as
  `pp_seat_…; Path=/ws/…; HttpOnly; SameSite=Lax`. With it, the same cookie carries `Secure`.

## Goals / Non-Goals

**Goals:** Say what any reverse proxy must do, show it once for a real one, and stop shipping a
proxy of our own.

**Non-Goals:** Certificate management, any second Compose file, and advice about which proxy to
choose. The base `compose.yaml` and `compose.local.yaml` are unchanged and remain the two things a
proxy attaches to.

## Decisions

1. **The worked example is Nginx Proxy Manager, in prose only.** It is what the author runs, so it
   is the one example that will be noticed when it goes wrong. Prose rather than a shipped file, for
   the reason this whole change exists: a configuration nobody exercises is a configuration nobody
   maintains.

2. **The example describes the interface, not the internals.** It names what to click — the proxy
   host, the forward target, the WebSockets toggle, the certificate tab — and then gives two checks
   the operator performs on their own installation. It deliberately does **not** claim which headers
   that product generates, because that is a claim about somebody else's software that this project
   cannot keep true.

3. **Two observable checks rather than a description.** Does the table connect at all — if not,
   suspect a rewritten `Host`. Does the `pp_seat_…` cookie carry `Secure` in the browser's developer
   tools — if not, `X-Forwarded-Proto` is not arriving. Both are things an operator can see in
   thirty seconds, and both point at the specific setting to change.

4. **The idle timeout is mentioned but needs no action.** nginx closes a connection with no traffic
   after sixty seconds by default; the application's heartbeat runs every thirty, so it holds the
   connection open. It is worth stating because the number matters in one direction: an operator who
   lowers that timeout below the heartbeat interval would break connections and have no idea why.

5. **The requirement about a supplied server configuration becomes a requirement about a documented
   one.** The scenario keeps its name and its point — the application works over HTTPS behind a
   proxy — while no longer promising that this repository ships the proxy.

## Risks / Trade-offs

- **Losing a tested end-to-end path.** The Caddy example was verified with real WSS and a `Secure`
  cookie. That evidence stays in the archived change that produced it; what is dropped is the
  ongoing promise to keep it working.
- **An operator with a proxy that rewrites `Host` and no way to stop it** would be stuck. Nothing in
  this change causes that — it is the application's existing origin check — but the documentation
  now names it instead of leaving them to guess.
- **Naming one product dates the documentation.** Accepted: an example that matches what somebody
  actually runs is worth more than a generic one that matches nobody.

## Migration Plan

Anyone running the deleted `compose.server.yaml` keeps running it — the file is gone from the
repository, not from their machine, and nothing in the application changed. The proxy requirements
that configuration satisfied are now written down, so it can be replaced with whatever they already
run. `compose.yaml` and `compose.local.yaml` are untouched.
