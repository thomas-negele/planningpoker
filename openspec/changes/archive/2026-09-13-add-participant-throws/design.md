## Context

See [proposal.md](proposal.md) for motivation and the delta specifications for the agreed behaviour.
The owner confirmed all three objects, all present non-self targets in either round phase,
mouse/touch access with a discreet accessibility path via standard keyboard navigation, offscreen
left/right entry, the timing and reduced-motion behaviour, and fixed limits of 3 throws per
participant and 12 per room per rolling second. Custom shortcuts and visible keyboard instructions
were explicitly excluded in the owner's clarification.

The Go hub owns each room in one goroutine. Today all commands share a 32-entry channel, and each
connection has a 16-entry reliable update channel whose overflow drops the connection. Transport
checks a 1 KiB message ceiling and a per-connection token bucket before decoding; its configurable
default is 10 messages/second with a burst of 20. Persistent raw flooding closes that socket.
Simply sending throws through the snapshot broadcast path would therefore cause unnecessary
encoding and could displace useful updates or disconnect slow clients.

The Svelte 5 frontend uses one `RoomConnection` per room. `Seat` renders a card and full name;
`RoomView` positions seats around an ellipse or in a narrow-screen list. Own-name editing already
uses a button. There is currently no frontend test runner; Go tests use injected clocks, explicit
behavioural assertions, real WebSocket integration fixtures and the race detector. Production
enforces a single-origin CSP without unsafe inline script/style allowances.

## Goals / Non-Goals

**Goals:**

- Keep authority and quotas with the room owner, while animation stays entirely in the browser.
- Bound every new collection and prefer game work at both room admission and socket delivery.
- Make variation reproducible in tests and consistent across different clients' local layouts.
- Isolate trajectory calculations from components so motion can be tested without a browser.
- Preserve existing comments, public error mapping, test fixtures and production CSP conventions.

**Non-Goals:**

- A physics engine, object-to-object collisions, replay log, persistent reactions, sound, scores,
  moderation controls, additional settings or a general event platform.
- Frame-perfect cross-device synchronisation, guaranteed delivery of cosmetic events under load,
  or a promise to survive arbitrary traffic beyond the configured server capacity.
- Changes to voting semantics, existing reconnection policy or unrelated pending-vote behaviour.

## Decisions

### 1. Use a narrow transient protocol alongside complete game snapshots

Add `throw` with `target` and `object`, where the object is one of `paper-ball`, `paper-plane`,
`flowers`. Derive the sender from the connection's seated participant in the room goroutine.
Check that the target is different, present and part of this room; do not consult vote status or
reveal state. Keep this transient action out of the mutating `act` helper, which always broadcasts
a full game view. Use public participant information for eligibility and avoid changing the game
domain solely to host decorative state.

Emit `thrown` with an event ID, public sender/target IDs, object, a bounded variation seed, and
event age at serialization. Use a room-local sequence for IDs, encoded without JavaScript integer
precision ambiguity, and an injectable random source for visual seeds. No coordinates, velocity,
HTML, SVG markup, sender identity or client-supplied randomness are accepted as authoritative input.
One shared seed determines entry side, timing, arc, rotation and landing offset. Each viewport
derives its own pixel path from that data and the target element. Object and IDs identify the event;
it does not need an earlier event to be understood.

The transport adds a small optional `throwPolicy` to the state envelope: participant and room
ceilings plus the actual connection message rate and burst. This avoids a frontend assumption
about operator configuration or duplicated numeric policy. Older clients ignore additive fields
and unknown events; a new client connected to a server without policy leaves throws disabled.
The existing intent-list specification also names the already implemented `setDeck` intent so its
closed list stays consistent while adding `throw`.

Use specific refusal codes for unsupported objects, self-targets and unavailable targets, and the
existing unseated code for an unseated sender. These invalid requests follow existing private
refusal handling. Valid requests discarded for cosmetic capacity or quota are silent drops; they
do not overwrite game refusal text. Render only server-emitted events, including for the sender,
so a locally suppressed or remotely discarded request creates no phantom animation.

Alternatives: embedding throws in snapshots would retain decorative state and replay it on
reconnect; optimistic local animation would show rejected throws; an extra socket would duplicate
identity, admission, liveness and lifecycle handling. None is necessary here.

### 2. Enforce exact rolling ceilings in the room and conservative pacing in the client

Store a bounded timestamp history of accepted throws per participant (3 entries) and per room
(12 entries), owned by the room goroutine. Expire timestamps at least one second old, check both
histories, and append to both only after eligibility and both allowances succeed. This implements
the agreed rolling-second maximum without a fixed-window boundary burst or a token bucket's extra
initial burst. Use the existing injected clock; no limiter timers or goroutines are needed.

Key histories by public participant ID, retaining them through disconnects while the seat exists.
Capacity is bounded by the room's existing participant limit. Room destruction discards them.
A full room quota does not spend the sender's participant quota, and invalid requests spend neither.
These are named fixed constants, explicitly exempted from the environment-variable convention in
the delivery delta and documented alongside the general message limit.

In `RoomConnection`, use a local rolling history for sent throws and account for every sent intent
in a conservative message budget. Start that budget empty after the first fresh policy/snapshot,
refill at half the advertised general rate up to its advertised burst, and debit all sent intents.
Game intents bypass this advisory gate and can drive the local budget negative; new throws require
enough budget for themselves plus one reserved message. Also suppress throws while the WebSocket
has buffered outgoing bytes. Keep the local history through a same-seat reconnect; reset socket
budget conservatively and await fresh policy and identity. Never retain unsent clicks or retry them.

This makes the maximum effective throw rate lower than 3 when the configured message rate is low;
at the valid minimum of 1/second, cosmetic refill is only 0.5/second before accounting for game
traffic. The spare rate and reserved message accommodate ordinary voting/reveal actions. This is
advisory protection for the supplied UI, not a replacement for server authority or permission for
unlimited game clicks. Excessive game traffic still has the existing general-limit semantics.

Keep the transport's raw message size/rate check before decoding for every message, including
invalid throws. Multiple tabs are independently protected by the connection limit and jointly
bounded by room admission; their aggregate excess is silently dropped at the throw layer.
Do not exempt throws from the raw limiter, weaken it, or automatically raise configured limits.

Alternatives: client-only limits are bypassable; per-socket throw quotas multiply with tabs;
queued retries produce throws after the user stops; exempting cosmetic messages from raw limits
allows excessive decoding and room-command work.

### 3. Bound cosmetic admission and delivery separately from reliable game work

Add a single-slot, nonblocking throw inbox to each room, separate from its existing command
channel. If occupied, drop the new cosmetic request. In the room loop check shutdown and pending
ordinary commands before processing a cosmetic request; recheck ordinary work when both paths
are ready. Validation and quota decisions still happen only on the room goroutine, including
revalidation of connection membership and target presence after any intervening game command.
Flooded throw submissions must not block their socket reader waiting for the ordinary room queue.

Give each connection a separate single-slot pending throw channel. Fan out an immutable event
nonblockingly, dropping that recipient's new effect when its slot is full. Do not enqueue cosmetic
events into `updates`, and do not close a connection because its cosmetic slot is full. A single
writer continues to own the socket and its deadlines. It services reliable updates, closure and
due heartbeats before another cosmetic write; it must write the initial snapshot before selecting
any throw. At most one already-started cosmetic write can precede a subsequently arriving update;
the existing write deadline bounds that unavoidable shared-socket delay.

During implementation, raise the existing reliable per-connection update buffer from 16 to 32.
The default 20-seat table can produce a near-simultaneous vote burst, especially under race
instrumentation; the original buffer could disconnect a healthy reader before the writer drained
it. This bounded adjustment is independent of the single-slot cosmetic delivery path and does
not make throw traffic reliable.

Keep an acceptance time with pending events on the server. Discard a pending event once the maximum
flight interval of 1.2 seconds has elapsed rather than releasing an old burst. Compute its
nonnegative `ageMs` at serialization so the client can also discard an already stale event. An
eligible received event starts its full flight from offscreen, with lifecycle timing based on
local elapsed time from receipt; starting partway through the path would make objects pop into
view. No synchronised browser/server clocks are required. Network delay can still shift different
viewers' start times; this cosmetic effect does not promise cross-device frame synchronisation.

The room cap bounds successful fan-out at 12 events/second times its configured connection ceiling:
with defaults, at most 480 delivery attempts per room/second and 24,000 across 50 full rooms.
These are bounds, not a throughput benchmark. Raw attempts are additionally bounded by connection
counts and the configured message limiter; pending memory never grows with activity duration.
Encode the recipient-independent cosmetic payload once for its actual serialization where practical;
do not build a domain snapshot for each throw. Keep shutdown ownership and existing dead-client
handling intact and exercise all new paths under the race detector.

Alternatives: using reliable queues risks dropping clients solely due to animations; unbounded
channels or goroutines defer the overload instead of bounding it. A shared writer preserves socket
ordering and avoids concurrent writer/heartbeat ownership.

### 4. Add the picker without changing the seat layout

Keep `Seat` responsible for exposing three small buttons. Fine-pointer hover reveals the choices
directly and leaves no separate trigger, pin or toggle visible below them. Retain a real
target-specific activation button for touch and standard keyboard navigation, but expose it only
in the modality that needs it: a coarse-pointer affordance for touch, or `:focus-visible` for
keyboard use. Own-name editing remains its existing button; other present seats get the throw
affordance. Keep the panel outside normal flow and within the viewport, with a continuous hover
region between seat and choices.

Use normal button/tab semantics rather than introducing a custom menu navigation model. Support
Escape, outside activation, focus leaving the picker, and closing when eligibility disappears.
Touch activation opens the same picker without relying on hover. Keep keyboard focus stable after
throws and return it to the trigger on explicit dismissal. Controls use the project's existing
English UI vocabulary; the choices are Paper ball, Paper plane and Flower. Use `:focus-visible`
for keyboard focus styling. Do not add global throw key bindings, letter/number shortcuts, shortcut
badges, keyboard instruction text or extra keyboard-only controls in the normal pointer interface.
Escape dismissal applies only to the open picker and is standard overlay behaviour, not a throw
shortcut; Tab and native button activation provide the accessibility path.

Alternatives: hover-only access excludes touch and keyboard users; changing the entire list item
into a button produces awkward nested interactive elements; adding the overlay to normal layout
would move participants as the pointer travels.

### 5. Render small local SVGs through one browser animation layer

Draw the paper ball with an irregular rounded silhouette, overlapping paper edges, creases and
small light/shadow regions so it reads as a sheet crumpled into a ball. Keep the plane's distinct
wings and clear nose. Draw one loose flower, not a bouquet, and derive one of several flower shapes
and colour palettes deterministically from the event seed. Share each illustration between picker
and flying object. Use the existing palette, compact proportions and modest shading; do not load
images or fonts.
This work needs vector assets, so no raster image generation or asset service is involved.

Mount one fixed, viewport-clipped SVG effect layer in `RoomView`, pointer-transparent and below
dialogs. Animate SVG transform/opacity attributes from one `requestAnimationFrame` loop; style
static appearance in compiled component styles. Do not generate inline style sheets or relax CSP.
Identify target DOM anchors by participant ID. Read necessary target rectangles together before
updating SVG attributes, and share each measured target between its active objects.

For each seeded event, choose a start x beyond the relevant viewport edge by the object's rotated
bounding radius, a bounded launch height, a visibly varied target-relative impact point and a duration in the
agreed interval. Use a time-based ballistic arc with exact endpoint constraints; parameterising
height above the straight start/end line gives gravity-like acceleration and avoids random misses.
The paper ball tumbles, bounces and rolls on; the plane follows its path tangent nose-first, glides
and skids; the flower rotates more slowly and slides a shorter distance after a soft impact. Derive
a bounded continuation vector from the incoming direction plus seed variation, then ease its
velocity to zero at a distinct final resting point. Include impact, slide and settling in the
0.6–1.2 second flight interval. Keep formulae and seeded parameter generation in pure TypeScript
functions with injected inputs.

Rest for 2 seconds, fade for 0.6 seconds, and remove expired objects by elapsed time, even after
a background-tab pause. Derive the live-object bound from the rate and maximum lifetime:
`12 * ceil(1.2 + 2 + 0.6) = 48` per page. If delayed delivery fills it, discard incoming effects
instead of growing it. Do not simulate collisions or piles; bounded landing offsets keep objects
visually near the target and below its name/card. Stop requesting frames when no effects remain.

Use participant-relative resting positions so scrolling and reflow keep objects at the correct
seat. On layout discontinuity discard in-flight effects whose path would be misleading; never
scroll the page to reveal a target. Skip new effects for offscreen/missing targets and remove
effects whose target element no longer exists. Clear the layer and subscriptions on disconnect
or room teardown. Ordinary vote/reveal/new-round snapshots do not restart effects.

Under reduced motion, skip the flight and start the rest phase at the final resting point;
switching preference during flight moves it directly into a full rest phase. An object already
resting or fading retains its elapsed rest/fade age while dropping any static rotation. Keep the
same cleanup path and no unnecessary motion callbacks once all objects have expired.

Alternatives: independent CSS randomisation disagrees across viewers; per-frame server physics
magnifies traffic; generic emoji rendering varies by platform; a physics engine adds size and
complexity without useful interaction between bodies.

### 6. Verify behaviour at the layer that owns it

Extend Go fixtures and write behaviour-named tests using injected clocks, controlled randomness,
explicit errors and bounded waits. Cover identity/target validation, rolling-window boundaries,
same-seat tabs/reconnects, room isolation, snapshot-first delivery, no hidden data, cosmetic queue
saturation, reliable update/heartbeat priority, and shutdown. Do not prove rate limits by sleeping.

Add a small frontend test command using the development Node runtime's built-in test facilities
for the pure trajectory/lifecycle and local pacing modules. Keep test helpers small and injected
time/seed/geometry explicit; check endpoints, offscreen bounds, varied paths, expiry, reduced motion,
the 48-object ceiling and mixed intents at message rates 1 and 10. Avoid component snapshots or
tests that merely restate implementation constants. The separate Go protocol-consistency tests
must cover new event fields and refusal codes as appropriate.

After review, add a separate, focused Playwright browser smoke test for the picker. Run it against
an isolated local Go server using the production-built frontend, so real pointer departure and
touch activation are checked without coupling the pure module tests to browser setup. Keep its
development dependencies in a separate package excluded from the production Docker context.

Run the repository's required build, type, vet, formatting and repeated race checks in apply.
Also inspect the built application in Docker with its production CSP: multiple independent
participants plus a second tab sharing a seat, wide and narrow layouts, long names, all three
objects from both edges, keyboard/touch, reduced motion, scroll/reflow, rapid clicking, and game
actions during sustained throws. Record actual results and any unavailable checks in tasks.

## Risks / Trade-offs

- Cosmetic events can be missed under load → intentional best-effort delivery; never show an
  optimistic throw or replay it later. Reliable game updates retain their independent path.
- A socket write already in progress cannot be preempted → keep events small, prioritise game
  updates before each write and retain network deadlines; do not claim zero possible added latency.
- Bounded work is not proof of hardware throughput → verify sustained multi-client traffic and
  memory cleanup; avoid claims that this provides general denial-of-service protection.
- Shared seeds do not give identical pixels across layouts → assert shared choices and correct
  local target anchoring, with browser inspection for visual quality.
- Narrow layouts have limited space beneath a name → size the artwork and landing offsets within
  real seat geometry, checking long names and dense participant lists without hiding controls.
- Conservative pacing can suppress a harmless click at low settings → silent drops are the agreed
  behaviour and leave capacity for estimation; no scheduled retry accumulates.

## Migration Plan

No stored data or new environment variables need migration. Deploy the frontend and backend
together in the existing single binary/container; the deployment still loses in-memory rooms as
documented by this project. Additive policy/event handling keeps ordinary game use compatible
with an older open frontend. Roll back by redeploying the previous bundle; reconnecting newer
pages without throw policy keep game controls usable and leave throwing disabled. No throw state
needs recovery or removal from persistent storage.
