package game

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"
	"strings"
	"testing"
)

// fixedRandom returns a reader yielding the same byte over and over, so that a test
// can assert on the identifier that comes out rather than merely that two differ.
func fixedRandom(b byte) io.Reader {
	return bytes.NewReader(bytes.Repeat([]byte{b}, 1024))
}

// shortRandom yields fewer bytes than an identifier needs.
type shortRandom struct{ remaining int }

func (s *shortRandom) Read(p []byte) (int, error) {
	if s.remaining <= 0 {
		return 0, io.EOF
	}
	n := min(len(p), s.remaining)
	s.remaining -= n
	return n, nil
}

// failingRandom fails immediately, as a source of entropy can when the system is
// under unusual conditions.
type failingRandom struct{}

func (failingRandom) Read([]byte) (int, error) { return 0, errors.New("entropy source unavailable") }

func TestIdentifierUsesOnlyTheSafeAlphabet(t *testing.T) {
	id, err := newID(rand.Reader)
	if err != nil {
		t.Fatalf("newID: %v", err)
	}

	for _, r := range id {
		if !strings.ContainsRune(idAlphabet, r) {
			t.Errorf("identifier %q contains %q, which is not in the alphabet", id, r)
		}
	}

	// The characters people confuse when reading a link aloud or retyping it from a
	// screenshot must not be in there at all.
	for _, confusable := range []string{"I", "L", "O", "U"} {
		if strings.Contains(idAlphabet, confusable) {
			t.Errorf("the alphabet contains %q, which is routinely misread", confusable)
		}
	}
}

func TestIdentifierCarriesTheRequiredEntropy(t *testing.T) {
	if idEntropyBytes*8 < 128 {
		t.Fatalf("identifiers consume %d bits of randomness, want at least 128",
			idEntropyBytes*8)
	}

	// 32 characters in the alphabet is 5 bits each, so the rendered length must be
	// enough to carry every bit that was consumed.
	id, err := newID(rand.Reader)
	if err != nil {
		t.Fatalf("newID: %v", err)
	}
	if got, want := len(id)*5, idEntropyBytes*8; got < want {
		t.Errorf("identifier %q carries %d bits, want at least %d", id, got, want)
	}
}

func TestFixedSourceProducesRepeatableIdentifier(t *testing.T) {
	first, err := newID(fixedRandom(0x42))
	if err != nil {
		t.Fatalf("newID: %v", err)
	}
	second, err := newID(fixedRandom(0x42))
	if err != nil {
		t.Fatalf("newID: %v", err)
	}

	if first != second {
		t.Errorf("two identical sources produced %q and %q; tests could not assert on identifiers",
			first, second)
	}

	other, err := newID(fixedRandom(0x43))
	if err != nil {
		t.Fatalf("newID: %v", err)
	}
	if other == first {
		t.Errorf("different sources produced the same identifier %q", first)
	}
}

func TestIdentifiersAreDistinctAndNotSequential(t *testing.T) {
	const count = 1000
	seen := make(map[string]bool, count)
	ids := make([]string, 0, count)

	for range count {
		id, err := newID(rand.Reader)
		if err != nil {
			t.Fatalf("newID: %v", err)
		}
		if seen[id] {
			t.Fatalf("identifier %q was produced twice in %d draws", id, count)
		}
		seen[id] = true
		ids = append(ids, id)
	}

	// A counter rendered in this alphabet would leave consecutive identifiers
	// sharing a long common prefix. Genuine randomness does not.
	for i := 1; i < len(ids); i++ {
		if shared := commonPrefix(ids[i-1], ids[i]); shared > 4 {
			t.Errorf("consecutive identifiers %q and %q share a %d-character prefix, "+
				"which suggests they are derived from a counter rather than from randomness",
				ids[i-1], ids[i], shared)
		}
	}
}

func TestFailingRandomSourceIsReportedNeverWorkedAround(t *testing.T) {
	for name, source := range map[string]io.Reader{
		"a source that errors":           failingRandom{},
		"a source with too few bytes":    &shortRandom{remaining: idEntropyBytes - 1},
		"a source with no bytes at all":  &shortRandom{remaining: 0},
		"no source of randomness at all": nil,
	} {
		t.Run(name, func(t *testing.T) {
			id, err := newID(source)
			if err == nil {
				t.Fatalf("newID returned %q and no error; a weak identifier must never be produced", id)
			}
			if !errors.Is(err, ErrShortRandomRead) {
				t.Errorf("error is %v, want it to match ErrShortRandomRead", err)
			}
			if id != "" {
				t.Errorf("newID returned the identifier %q alongside an error", id)
			}

			// The same must hold one level up: no room and no participant may come
			// into existence with a padded or partially random identifier.
			if room, err := NewRoom(source, testMaxParticipants); err == nil {
				t.Errorf("NewRoom succeeded with a broken source, producing room %q", room.ID())
			}
		})
	}
}

func commonPrefix(a, b string) int {
	n := 0
	for n < len(a) && n < len(b) && a[n] == b[n] {
		n++
	}
	return n
}

func TestValidRoomIDAcceptsWhatTheGeneratorProduces(t *testing.T) {
	// The two must never drift apart, so this checks the generator's own output
	// rather than a hand-written example that could go stale.
	for range 200 {
		id, err := newID(rand.Reader)
		if err != nil {
			t.Fatalf("newID: %v", err)
		}
		if !ValidRoomID(RoomID(id)) {
			t.Errorf("the generator produced %q, which its own validator rejects", id)
		}
	}
}

func TestValidRoomIDAcceptsChosenNames(t *testing.T) {
	// A name somebody typed is as valid as one the generator issued. It is not as
	// private — nothing about the identifier says which kind it is — but it names a
	// room just as well.
	for _, id := range []RoomID{
		"team-alpha",
		"Standup_2",
		"abcde", // exactly the minimum
		RoomID(strings.Repeat("a", MaxRoomIDLength)), // exactly the maximum
		"UPPER-and-lower_123",
	} {
		if !ValidRoomID(id) {
			t.Errorf("%q was refused, but it is a usable name", id)
		}
	}
}

func TestValidRoomIDRejectsAnythingElse(t *testing.T) {
	for name, id := range map[string]RoomID{
		"empty":               "",
		"one character":       "a",
		"four characters":     "abcd",
		"one over the limit":  RoomID(strings.Repeat("a", MaxRoomIDLength+1)),
		"a space":             "team alpha",
		"a slash":             "team/alpha",
		"a dot":               "team.alpha",
		"a percent sign":      "team%20alpha",
		"an exclamation mark": "team!",
		"an umlaut":           "gruppe-fünf",
	} {
		if ValidRoomID(id) {
			t.Errorf("%s: %q was accepted", name, id)
		}
	}
}

func TestValidRoomIDMatchesExactly(t *testing.T) {
	// Nothing is folded. Two spellings are two names, which is what keeps the address
	// bar honest: what stands there is where you are.
	if !ValidRoomID("Team-Alpha") || !ValidRoomID("team-alpha") {
		t.Fatal("both spellings should be usable names")
	}
	if RoomID("Team-Alpha") == RoomID("team-alpha") {
		t.Error("the two spellings compare equal; they must not")
	}
}

func TestAnIssuedIdentifierIsStillAcceptedAfterWidening(t *testing.T) {
	// The widening admitted chosen names. It must not have broken the kind that
	// actually protects a room.
	for range 200 {
		id, err := newID(rand.Reader)
		if err != nil {
			t.Fatalf("newID: %v", err)
		}
		if !ValidRoomID(RoomID(id)) {
			t.Errorf("the generator produced %q, which its own validator rejects", id)
		}
		if len(id) < MinRoomIDLength || len(id) > MaxRoomIDLength {
			t.Errorf("issued identifier %q is %d characters, outside the accepted bounds", id, len(id))
		}
	}
}
