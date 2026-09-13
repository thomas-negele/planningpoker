## Why

Participants need a playful way to nudge someone during estimation without interrupting the
meeting. Small, shared throws add that moment of humour while demonstrating polished animation,
accessible interaction, and bounded real-time processing in this demo project.

## What Changes

- Show a compact throw picker over another present participant on hover or touch. Retain standard
  keyboard access for accessibility, without custom shortcuts, keyboard hints or extra visible
  controls during pointer use. Offer exactly a crumpled paper ball, paper plane and single flower
  as local SVGs; flower colour and type may vary from the shared event seed.
- Let every seated participant target any other present participant, before or after voting
  and reveal. Share accepted throws with everyone connected to that room, including the sender
  and target.
- Animate objects entering from outside either horizontal viewport edge, chosen randomly, with
  varied flight paths, speeds and object-specific motion. Use visibly varied impact points near
  the target and let each object slide or roll a short, decelerating distance before resting just
  in front of/below the seat. Complete flight and settling after approximately 0.6–1.2 seconds,
  rest for 2 seconds, then fade over 0.6 seconds.
- Respect reduced motion by showing the object directly at its landing position.
- Limit accepted throws to 3 per second per participant across tabs and 12 per second per room.
  These are fixed constants, as explicitly chosen by the owner. Discard excess clicks without
  queueing retries; keep game actions ahead of cosmetic traffic and respect the existing
  configurable message allowance, including low settings.
- Deliver bounded, ephemeral events; do not persist throws or replay them after reconnection.

## Capabilities

### New Capabilities

- `participant-throws`: Target eligibility, the three objects, shared transient delivery, fixed
  throw limits, and bounded resource use.

### Modified Capabilities

- `table-ui`: Accessible participant picker, SVG flight/landing/fade presentation, reduced motion,
  and cleanup without obscuring the estimation controls.
- `live-updates`: Add the throw intent and transient event alongside authoritative snapshots,
  describe best-effort delivery and scoped refusals, and coordinate with message limits.
- `app-delivery`: Record the explicitly approved fixed throw limits as an exception to the
  environment-variable rule; existing deployment settings retain their current behaviour.

## Impact

The room hub, WebSocket protocol and writer, central frontend connection, and participant/table
components need changes. The implementation will add small SVG assets and client-side trajectory
code, without a physics engine, external assets, extra sockets, or per-frame server messages.
Existing Go tests will cover validation, admission, shared identity, backpressure and protocol
behaviour; focused frontend tests and production-browser checks will cover motion and interaction.
Game rules, votes and results do not change. Protocol additions retain complete snapshots and
must not disclose private seat credentials or hidden vote values.
