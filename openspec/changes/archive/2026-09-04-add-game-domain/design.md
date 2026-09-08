## Context

See `proposal.md` — Why for the motivation. The constraints that shape this design are fixed in
`CLAUDE.md` and in the previous change: the domain lives in `internal/game`, it must not import
`net/http` or the WebSocket library, room identifiers come from `crypto/rand` with roughly 128 bits
of entropy, hidden votes must never leave the server, and there is no persistence of any kind.

One constraint is worth stating plainly because it shapes almost every decision below: **this
package is single-threaded by construction.** In the next change each room is owned by exactly one
goroutine, and nothing outside that goroutine touches the room's state. So nothing here needs a
mutex, and nothing here should grow one — a lock inside these types would be a sign that the
ownership model above had been abandoned.

## Goals / Non-Goals

**Goals:**

- Make the rules readable as rules. Someone should be able to answer "may I vote after the reveal?"
  by reading one function, not by tracing a message through a handler.
- Make the hidden-vote guarantee structural rather than remembered: it should be difficult to leak
  a vote by accident and impossible to do so without deleting a test.
- Leave the next change — the room manager and the WebSocket protocol — with an interface it can
  call directly, without needing to reinterpret any rule.

**Non-Goals:**

- Concurrency of any kind. No mutexes, no channels, no goroutines in this package.
- Serialisation. No JSON tags, no encoding, no wire format. The transport layer decides how the
  view is written to a socket; the domain decides what may be in it.
- A second deck, or any mechanism for choosing between decks.
- Room lifetime. Deciding when an unused room is discarded needs a clock and a timer, which makes
  it a matter for the layer that owns time, not for a rule.

## Decisions

### The redacted view is a separate type, not a flag on the room

The room does not hand out itself. It hands out a distinct view type, built by a function that takes
the round's revealed state into account, and that view is the only thing the transport layer ever
sees. While the round is hidden, the view's vote fields do not merely hold a blank — the type has
nowhere to put a card value at all for an unrevealed round.

The alternative is the obvious one: expose the room and have the transport layer remember to omit
the votes. It is rejected because the guarantee then depends on every future author remembering,
and because "we removed the field before sending" cannot be tested — you can only test that this
particular code path removed it.

With a separate type, the test is direct: build a room where everyone has voted, produce the hidden
view, serialise it to text by any means, and assert that no card value appears anywhere in the
output. That test fails the moment someone adds a field that carries one.

### Vote values are a distinct type, not strings

A card is its own type rather than a bare `string`. The deck is the authoritative list, and a value
is only a card if the deck contains it. This makes "a card outside the deck is refused" a check in
one place rather than a validation repeated at every call site, and it makes it impossible to pass
a participant's name where a card was expected.

The alternative — plain strings with validation at the boundary — is less code today and reliably
grows a second, subtly different validation later.

### Randomness is a parameter, never a package-level global

Creating a room and creating a participant both need random bytes. The source is passed in as an
`io.Reader`. Production passes `crypto/rand.Reader`; tests pass a reader over fixed bytes and get
repeatable identifiers.

This costs one parameter and buys two things: the tests can assert on identifiers instead of merely
asserting that two of them differ, and there is no hidden global that a future change could swap for
something weaker without anyone noticing. `io` is not input or output in the sense the "no I/O"
rule means — the rule is about network and files, and specifically about `net/http` and the
WebSocket library.

A failure from the reader is returned, never worked around. A room whose identifier fell back to
something weak because entropy was briefly unavailable would be a room anyone could find, and the
invitation link is the only thing protecting it.

### The identifier alphabet excludes confusable characters

Identifiers are rendered in an alphabet that omits characters which are routinely misread for one
another. People do read these links aloud and retype them, and a room that cannot be found because
someone heard `l` and typed `1` is a room that failed for a reason that was entirely avoidable.

Entropy is counted in bits of randomness consumed, not in characters produced, so shrinking the
alphabet lengthens the identifier slightly rather than weakening it.

### Away is a property of the participant, not a separate collection

A participant marked away stays in the same collection as everyone else, carrying a flag. The
alternative — moving them to a second list of absent participants — would mean every rule that
walks the participants has to remember to consult both lists, and the first rule that forgets
produces a bug that only appears when someone's network drops.

The consequence to keep in mind: any code asking "how many participants are there?" must be explicit
about whether it means everyone seated or only those present. The two rules that care —
"everyone present has voted" and "is anyone still here" — say so in their names.

### Errors are sentinel values the caller can recognise

Refusals return errors the transport layer can identify rather than free text: an unknown
participant, a card not in the deck, a vote after the reveal, a name that is empty or too long. The
next change has to turn a refusal into a message for the browser, and matching on error strings is
how that turns into a bug the first time an error message is reworded.

### Rules are methods on the room, and every one of them is total

Every operation either changes the room and reports success, or changes nothing and reports why.
There is no partial application: a vote is recorded or it is not, a rename takes effect or the old
name stands. Nothing is left half-applied for the caller to clean up, which matters because the
caller is a goroutine that will simply carry on with the next message.

## Risks / Trade-offs

- **The hidden-vote guarantee could be undone by a well-meaning addition** — someone adds a field to
  the view for a good reason and puts a card in it. → The test that serialises the hidden view and
  searches it for card values catches exactly this, and it fails loudly rather than subtly. It is
  the one test in this change that must never be weakened to make a build pass.
- **The "no `net/http`" rule is easy to state and easy to violate** by an import added for one
  convenient helper. → An automated test walks this package's imports and fails on `net/http` or the
  WebSocket library, so the constraint is enforced rather than remembered.
- **These rules cannot be exercised by a person after this change.** Nothing serves them, so an
  error in them is invisible until the next change wires them up. → Mitigated by testing them
  exhaustively here, which is the entire reason for isolating them; but it does mean the next change
  should expect to find at least one thing the tests did not think of.
- **Away participants accumulate over a very long session**, since nothing removes them. → Bounded
  in practice: a reconnecting browser is reseated rather than added, so growth requires genuinely
  different people joining and leaving, and the room itself is discarded once deserted. If it ever
  becomes a problem, the fix is a rule about removal, which is a change to be requested.

## Migration Plan

Nothing to migrate: the package does not exist yet and nothing depends on it. It is additive, and
the deployed application behaves identically before and after, since nothing calls into it. Rollback
is deleting the package.

## Open Questions

- The exact maximum length for a display name is not fixed here. It exists to stop one participant
  disrupting the table for everyone, and any value in the region of a few dozen characters serves
  that; it is not a value whose precise number changes anyone's behaviour, so it is settled during
  implementation and recorded in the code with its reasoning. If it later needs to be operator-
  adjustable, it becomes an environment variable like the others.
