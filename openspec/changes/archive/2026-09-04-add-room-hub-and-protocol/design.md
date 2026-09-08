## Context

See `proposal.md` — Why. The pieces this change joins already exist and constrain how it may be
done: `internal/game` is single-threaded by construction and carries no lock, the hidden card is
unrepresentable in its view rather than merely omitted, and `internal/transport` currently owns a
throwaway echo endpoint whose connection bookkeeping was deliberately built small enough to throw
away.

The concurrency model was settled before any of this was written and is not reopened here: a manager
holds the rooms behind a lock that protects the map and nothing else, and each room runs one
goroutine that owns its state. What this change decides is the shape of the messages, the ownership
of connections, and how time enters the system for the first time.

## Goals / Non-Goals

**Goals:**

- Make the two guarantees hold under real concurrency: exactly one owner per room, and no hidden
  card on the wire.
- Leave the frontend change with a protocol it can render directly — snapshots that stand alone, and
  refusals specific enough to produce a useful message.
- Introduce time in one place, so that "when is a room discarded" has a single answer that can be
  tested without waiting for real minutes to pass.

**Non-Goals:**

- Any user interface. The placeholder page is adjusted only so the deployed application does not
  display a failed connection; the table is the next change.
- Persistence, and therefore any form of surviving a restart.
- Reconnect logic in the client — retrying a dropped socket is a frontend concern and belongs with
  the interface that shows its state.
- Limits on how many rooms may exist or how fast they may be created. Nothing has asked for them,
  and adding them would be inventing a policy.

## Decisions

### Rooms are reached by message passing; nothing shares memory with a room goroutine

A room goroutine owns its `game.Room` outright. Every browser connection sends it commands over a
channel and receives snapshots over its own channel. Nothing else ever holds a pointer into that
room's state.

This is what makes the rules safe to use unchanged: `game.Room` has no lock and needs none, because
only one goroutine ever touches it. It also means the interesting logic stays testable as the pure
functions it already is — the concurrency lives entirely in this layer and can be reasoned about
separately from the rules.

The manager's lock protects the map of rooms and nothing else. It is held only while looking a room
up, inserting one or deleting one, and never while doing anything with a room's contents. A lock
held across a room operation would serialise every room in the process behind every other.

### Each connection has a buffered outbound channel and is dropped if it overflows

Writes to a connection go through a buffered channel owned by that connection. The room goroutine
sends a snapshot to each connection's channel and moves on; it never waits for a socket.

If a connection's buffer is full — meaning that browser has stopped reading and is far enough behind
that catching up is hopeless — that connection is dropped rather than allowed to apply back-pressure
to the room. One wedged tab must not be able to freeze a meeting for everyone else, and a snapshot is
a complete picture, so a dropped client that reconnects is immediately correct again.

The alternative, blocking until the slow client accepts the write, is the standard way this kind of
application deadlocks under exactly the conditions where it matters most.

### Snapshots are produced by the domain and serialised, never assembled by the transport

The transport layer sends what `game.Room.View()` gives it and adds only what is genuinely about the
connection rather than the room. It does not reach into the room for state, and it does not build a
message out of parts.

This is the whole reason the view type was designed the way it was. If the transport were free to
assemble its own message, the guarantee that a hidden card is unrepresentable would be worth nothing:
the leak would simply move one layer up. The test that proves no card crosses the network therefore
belongs here as well as in the domain, and it should be written against the actual bytes on the
socket.

### Time enters through an injected clock, so room expiry is testable without waiting

The manager takes its notion of "now" as a dependency rather than calling the system clock directly.
Production passes the real one; tests pass one they control.

Without this, testing that a room survives four minutes and is gone after six means a test that takes
six minutes, which means a test nobody runs. With it, the grace period rule is an ordinary unit test.

The expiry check itself is periodic rather than a timer per room: a single sweep asks each room
whether it has been unoccupied for longer than the grace period. One goroutine and one interval is
easier to reason about than a timer per room that has to be cancelled and rescheduled on every
connect and disconnect, and the imprecision it introduces — a room may outlive its grace period by up
to one sweep interval — does not matter for a value measured in minutes.

### The room decides its own emptiness; the manager decides its fate

A room reports whether it currently has open connections and when it last had one. The manager reads
those and deletes the room. The room does not delete itself.

Keeping the decision in the manager means there is one place where a room can be removed from the map
and one place where the map is locked, which is what makes it possible to state confidently that a
connection can never arrive at a room that is being deleted underneath it.

### Refusals are typed codes derived from the domain's sentinel errors

The domain returns sentinel errors, which is what makes this possible: the transport maps each one to
a short stable code in the protocol, and the frontend maps that code to a sentence. Nothing matches
on error text at any layer.

Mapping is explicit rather than automatic — a lookup that turns an unmapped error into a generic code
would silently degrade a specific refusal into "something went wrong" the moment someone adds a new
sentinel. An unmapped error should be conspicuous, so the mapping is exhaustive and a new sentinel
without a code is a test failure.

### The participant cookie is scoped to the room's path and is not readable by scripts

The cookie is set with the room's path, so the browser only ever sends it to the room it belongs to,
and two rooms receive different identifiers for the same browser. It is marked so that scripts in the
page cannot read it: nothing in the page needs it, and it is the one value that would let somebody
take over another person's seat.

The name cookie is a different matter and stays where it is — the page itself needs to read the name
to pre-fill the entry field, so it is not the server's cookie to hide. It carries nothing sensitive.

### The echo endpoint is deleted rather than kept alongside

`echo_probe.go` goes, and so does the placeholder page's probe of it. It was written in one
self-contained file precisely so that this would be a single deletion.

Keeping it "just in case" would leave an endpoint that accepts connections and echoes arbitrary
content, which is a small but real thing to have on a public server for no reason.

## Risks / Trade-offs

- **A hidden card could leak here even though the domain makes it unrepresentable**, if the transport
  assembles its own message or logs a view. → The test that proves it must read the actual bytes sent
  over a real WebSocket, not a Go value. A test that inspects a struct proves the domain's guarantee
  again rather than this layer's.
- **The classic deadlock in this design is a room goroutine waiting on a connection** that is waiting
  on the room. → Buffered outbound channels, and a drop rather than a block when the buffer is full;
  plus a test that deliberately stops reading from one connection and asserts the others continue.
- **A room could be deleted between being looked up and being used.** → The lookup and the decision
  to use a room happen under the same lock, and a room that has been removed refuses further commands
  rather than accepting them into a channel nobody is reading.
- **The race detector only finds races that are actually exercised.** A test with two goroutines
  proves less than it appears to. → The concurrency test should run many connections doing conflicting
  things at once, and it must run under `-race` in the ordinary test command rather than in some
  separate suite that is easy to skip.
- **The grace period is the first thing here that can be got wrong in a way nobody notices**, because
  a room expiring too eagerly looks like a bug in something else entirely. → Test both directions
  explicitly: survives just under the period, gone just over it.

## Migration Plan

Additive apart from one deletion. The echo endpoint disappears, which nothing depends on except the
placeholder page's probe, and that is removed in the same change. Deployment is unchanged:
`docker compose up --build`, with one new environment variable that has a working default.

Rollback is redeploying the previous image, with the standing consequence that a deployment destroys
every open room. That consequence becomes real with this change rather than theoretical, and it is
worth saying plainly to whoever runs the deployment: do not deploy during a planning session.

## Open Questions

- The sweep interval for expiring rooms is not fixed here. It must be well below the grace period so
  that a room does not outlive it noticeably, and it is otherwise uninteresting. It is settled during
  implementation and recorded in the code with its reasoning; it becomes a setting only if it ever
  turns out to need adjusting, which would be surprising for a value that only trades a little
  idle work against a little imprecision.
