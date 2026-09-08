## ADDED Requirements

### Requirement: An issued room identifier is unguessable

Because there is no authentication anywhere in this product, an invitation link is the only thing
that can protect a room. An identifier **issued** by this package — the kind produced when somebody
starts a new game — SHALL therefore be generated from a cryptographically secure source of
randomness, carry at least 128 bits of entropy, and be rendered in a URL-safe alphabet that excludes
characters easily confused with one another when read aloud or copied by hand.

Issued identifiers MUST NOT be sequential, derived from a counter, or derived from the time of
creation, since any of those would let one room be found from another.

The source of randomness SHALL be supplied by the caller rather than read from a package-level
global, so that the rules can be tested with a fixed source and produce repeatable identifiers.

This requirement is now about one of two kinds of identifier rather than about all of them. A room
may also be reached by a name somebody typed, which is guessable by design and is described in
`game-sessions`. Nothing here weakens: an identifier this package issues is exactly as unguessable
as it ever was. What changed is that being unguessable is a property of how an identifier was made,
not of every identifier that may name a room — and saying so plainly is what stops somebody later
reading this requirement as a promise the product does not keep.

#### Scenario: Two rooms never share an identifier

- **WHEN** many rooms are created in succession from a cryptographically secure source
- **THEN** every identifier differs from every other, and none can be derived from another by
  incrementing, decrementing, or any other simple transformation

#### Scenario: Identifier is safe to put in a URL

- **WHEN** a room identifier is generated
- **THEN** it consists only of characters that need no escaping in a URL path, and contains none of
  the character pairs that are routinely misread for each other

#### Scenario: A fixed source produces a repeatable identifier

- **WHEN** a room is created twice from two sources of randomness yielding identical bytes
- **THEN** both rooms receive the same identifier, so that tests can assert on it

#### Scenario: A failing source of randomness is reported, never worked around

- **WHEN** the source of randomness returns an error or too few bytes
- **THEN** room creation fails with that error, and no room is created with a weak, padded or
  partially random identifier

#### Scenario: A typed name is not held to this rule

- **WHEN** a room is reached by a name somebody typed rather than by an issued identifier
- **THEN** it is a room like any other, and no claim is made that its address cannot be guessed

## REMOVED Requirements

### Requirement: A room is identified by an unguessable identifier

**Reason**: Replaced by "An issued room identifier is unguessable". Every rule it stated is carried
over word for word — the cryptographically secure source, the 128 bits, the alphabet without
confusable characters, no counters, no timestamps, and randomness supplied by the caller. Not one of
them is relaxed.

What changed is only its scope, and the old title was the problem: it said *a* room identifier is
unguessable, which stopped being true the moment a room could also be reached by a name somebody
typed. Left as it was, this requirement would have read as a promise the product no longer keeps —
and somebody would have found that out by trusting it.

**Migration**: None. No stored data and no protocol message is involved, and no identifier already
issued is affected.
