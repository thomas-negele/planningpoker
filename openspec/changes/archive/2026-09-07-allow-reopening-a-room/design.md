## Context

See `proposal.md` — Why, including why the button this change first built was removed
again before it was ever archived.

Two constraints from earlier changes still bind. The manager's lock protects the map
of rooms and nothing else, and is never held while doing anything with a room's
contents. And issued identifiers come from a cryptographically secure source with 128
bits of entropy — that property survives this change untouched, and is now one of two
kinds of identifier rather than the only one.

## Goals / Non-Goals

**Goals:**

- An old link works. Opening it puts you in a room, whether or not one was there.
- A group interrupted by a restart ends up together again, without anybody pressing
  anything or sending anything.
- A memorable address for a recurring meeting.
- Say something when a round is actually lost, and nothing when it is not.

**Non-Goals:**

- Persistence. A room that comes back is empty; restoring participants or votes would
  mean storing them, which this product has deliberately never done.
- Any privacy claim about a chosen name. It is guessable and the specification says
  so.
- Any change to the rules of the game. Seating, voting, revealing and results are
  untouched.

## Decisions

### There are two kinds of identifier, and they are not the same kind of thing

An **issued** identifier is generated and unguessable. A **chosen** one is typed and
is not. Both reach a room; only the first protects it.

The temptation is to blur them — to treat "what may appear after `/g/`" as one
question with one answer. Keeping them apart is what makes it possible to say
truthfully that starting a new game gives you a private room, while `/g/standup` is a
public one. If they were one concept, either the specification would have to stop
claiming privacy for issued links, or a chosen name would have to pretend to
something it does not have.

### Identifiers are matched exactly, and the interface never mentions the difference

Nothing is folded or normalised. `/g/Team-Alpha` and `/g/team-alpha` are two rooms.
The address bar is the address: what stands there is where you are, with no rule
running behind it that a visitor cannot see.

The alternative — folding case so a remembered name always finds its room — trades
that for a hidden rule, and buys less than it looks. It rescues one kind of typo and
none of the others, so somebody who mistypes a letter is still alone in a room of
their own, and now cannot tell why that one failed when the capital letter did not.

For the same reason the interface says nothing about a chosen room being public. A
badge would put a security caveat permanently on screen, about a product containing
nothing but t-shirt sizes, and would prompt a question it has no useful answer to.
The property is written into the specification, which is where a decision belongs;
it is not something to explain to somebody who wants to estimate a ticket.

### A room is created where it is looked for, not by a separate request

The socket handler creates the room when it does not find one. There is no reopen
route, no separate call, and nothing for the page to decide.

The first version had all three, and the resulting flow was: connect, discover the
room is gone, close the connection, show a screen, wait for a click, make an HTTP
request, reconnect. Six steps for something nobody would ever answer "no" to. Now the
connection that was going to happen anyway is enough.

That the creation hangs off the socket rather than the page request matters for one
practical reason: link-preview bots in chat clients fetch the HTML of any URL that is
pasted. They do not run scripts and do not open sockets, so they cannot bring rooms
into existence by unfurling an old link.

### Creation is "ensure it exists", under one lock

Several people opening the same URL in the same instant is the ordinary case after a
restart, not an edge case. The operation is therefore "make sure a room is here", and
it returns the room whether it made it or found it.

The check and the insertion happen under the same lock, or two simultaneous arrivals
could each find nothing, each create a room, and one would silently replace the other
— taking whoever had already joined it along.

### The announcement is triggered by observation, not by a close code

A page knows a round was lost when it held a seat, reconnected, and does not appear in
the room it reached. That is the fact, and it is true regardless of which of the many
ways the socket happened to close.

Deriving it from the close code instead would mean guessing: a restart, an expiry and
a network fault can all close a socket the same way, and only some of them cost you
anything. The observable difference is whether the room still knows you.

### Nothing is announced for an ordinary reconnection

A dropped connection loses nothing — the room outlives it and the seat comes back —
so the connection status line saying "reconnecting" is the whole of what it deserves.
A message for that would train people to dismiss messages, and the one message that
matters would go with the rest.

## Risks / Trade-offs

- **An invitation link stops expiring in any practical sense**, and **a chosen name is
  public**. → Both are in the specification as consequences rather than footnotes,
  both were decided with the consequence stated, and the second is the reason issued
  and chosen identifiers are described as different kinds.
- **A forgotten tab keeps a room alive forever**, recreating it after every restart. →
  Accepted knowingly. A room is a goroutine and a small struct; a handful of forgotten
  tabs costs nothing worth defending against, and the alternative is refusing to
  reconnect, which is the behaviour being removed.
- **Anybody can now put a room in the map by opening a URL.** → Bounded by what counts
  as an acceptable identifier, and a room is cheap. This is worth watching if the
  application is ever exposed somewhere hostile, where it becomes a way to occupy
  memory; a limit on the number of rooms would be the answer, and it is not built
  speculatively.
- **The announcement could fire when nothing was lost**, if the test for "the room
  does not know me" is wrong — for instance during the first snapshot of a normal
  first visit, when nobody is seated yet. → It fires only for a page that *held* a
  seat and then found itself absent, which a fresh visit never satisfies.

## Migration Plan

Additive for anyone using it, destructive only of a screen. Deployment is unchanged.

This is the change that makes a deployment during a session recoverable: everybody's
page reconnects into a room at the same address. It does not make it invisible — the
round in progress is still lost, and now says so.

## Open Questions

None. The two that would have changed the shape of this were both settled with their
consequences on the table: whether creation happens on arrival or on request, and
whether a chosen name may be guessable.
