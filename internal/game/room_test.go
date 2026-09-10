package game

import (
	"crypto/rand"
	"errors"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"
)

// testMaxParticipants is deliberately far above anything these tests seat, so that a
// test about voting or renaming fails for its own reason and never because it
// quietly ran into the seat limit. The tests that are about that limit set their own.
const testMaxParticipants = 1000

// newTestRoom creates a room, failing the test rather than returning an error, so
// that the tests below read as the rules they are checking.
func newTestRoom(t *testing.T) *Room {
	t.Helper()
	room, err := NewRoom(rand.Reader, testMaxParticipants)
	if err != nil {
		t.Fatalf("NewRoom: %v", err)
	}
	return room
}

// join seats somebody, failing the test if it does not work.
func join(t *testing.T, room *Room, name string) ParticipantID {
	t.Helper()
	id, err := room.Join(rand.Reader, name)
	if err != nil {
		t.Fatalf("Join(%q): %v", name, err)
	}
	return id
}

// nameOf reads a participant's displayed name out of the view, which is the only
// place names are visible from outside.
func nameOf(t *testing.T, room *Room, id ParticipantID) string {
	t.Helper()
	for _, p := range room.View().Participants {
		if p.ID == id {
			return p.Name
		}
	}
	t.Fatalf("participant %q is not at the table", id)
	return ""
}

func TestJoiningSeatsTheParticipant(t *testing.T) {
	room := newTestRoom(t)
	id := join(t, room, "Thomas")

	view := room.View()
	if len(view.Participants) != 1 {
		t.Fatalf("table has %d participants, want 1", len(view.Participants))
	}
	if view.Participants[0].ID != id {
		t.Errorf("seated participant is %q, want %q", view.Participants[0].ID, id)
	}
	if view.Participants[0].Name != "Thomas" {
		t.Errorf("name = %q, want %q", view.Participants[0].Name, "Thomas")
	}
}

func TestTwoPeopleMayShareAName(t *testing.T) {
	room := newTestRoom(t)
	first := join(t, room, "Thomas")
	second := join(t, room, "Thomas")

	if first == second {
		t.Fatal("both participants received the same identifier; they must be distinct")
	}
	if got := len(room.View().Participants); got != 2 {
		t.Fatalf("table has %d participants, want 2", got)
	}

	// Identity is carried by the identifier, so an operation on one must leave the
	// other untouched even though their names are identical.
	if err := room.Vote(first, CardM); err != nil {
		t.Fatalf("Vote: %v", err)
	}
	for _, p := range room.View().Participants {
		if p.ID == first && !p.Voted {
			t.Error("the participant who voted is not shown as having voted")
		}
		if p.ID == second && p.Voted {
			t.Error("voting for one participant also marked the other, who shares their name")
		}
	}
}

func TestEmptyNameIsRefused(t *testing.T) {
	room := newTestRoom(t)

	for _, name := range []string{"", " ", "\t", "\n", "   \t  \n "} {
		id, err := room.Join(rand.Reader, name)
		if !errors.Is(err, ErrEmptyName) {
			t.Errorf("Join(%q) error = %v, want ErrEmptyName", name, err)
		}
		if id != "" {
			t.Errorf("Join(%q) returned identifier %q alongside the error", name, id)
		}
	}

	if got := len(room.View().Participants); got != 0 {
		t.Errorf("table has %d participants after refused joins, want 0", got)
	}
}

func TestSurroundingWhitespaceIsTrimmed(t *testing.T) {
	room := newTestRoom(t)
	id := join(t, room, "  Thomas  ")

	if got := nameOf(t, room, id); got != "Thomas" {
		t.Errorf("name = %q, want %q", got, "Thomas")
	}
}

func TestExcessivelyLongNameIsRefused(t *testing.T) {
	room := newTestRoom(t)

	atLimit := strings.Repeat("a", MaxNameLength)
	if _, err := room.Join(rand.Reader, atLimit); err != nil {
		t.Errorf("a name of exactly the maximum length was refused: %v", err)
	}

	tooLong := strings.Repeat("a", MaxNameLength+1)
	if _, err := room.Join(rand.Reader, tooLong); !errors.Is(err, ErrNameTooLong) {
		t.Errorf("Join(name of %d characters) error = %v, want ErrNameTooLong", len(tooLong), err)
	}

	// The limit is counted in runes, so a name in any script gets the same
	// allowance rather than being cut short by its encoding.
	multibyte := strings.Repeat("ü", MaxNameLength)
	if _, err := room.Join(rand.Reader, multibyte); err != nil {
		t.Errorf("a name of %d multi-byte runes was refused: %v", MaxNameLength, err)
	}

	if got := len(room.View().Participants); got != 2 {
		t.Errorf("table has %d participants, want 2 (the two accepted names)", got)
	}
}

// The test above derives its fixtures from MaxNameLength, so it follows the constant
// wherever it goes and proves the boundary is enforced without saying where it is.
// This one says where it is. The limit is fifteen because that is what a seat can
// display in full, so it is worth pinning with names somebody might actually type: if
// a later change moves the constant, this test fails and the layout that was chosen
// alongside it gets looked at rather than silently outgrown.
func TestTheNameLimitIsFifteenCharacters(t *testing.T) {
	room := newTestRoom(t)

	const fits = "Maria-Katharina"     // fifteen
	const tooLong = "Johann Sebastian" // sixteen

	if got := utf8.RuneCountInString(fits); got != MaxNameLength {
		t.Fatalf("the fixture %q is %d characters, want %d — fix the fixture, not the limit", fits, got, MaxNameLength)
	}

	if _, err := room.Join(rand.Reader, fits); err != nil {
		t.Errorf("Join(%q) error = %v, want a seat: fifteen characters is the limit, not one past it", fits, err)
	}

	if _, err := room.Join(rand.Reader, tooLong); !errors.Is(err, ErrNameTooLong) {
		t.Errorf("Join(%q) error = %v, want ErrNameTooLong", tooLong, err)
	}

	// The refusal must leave nothing behind. A rejected name that still took a seat
	// would put somebody at the table under a name the rules just refused.
	if got := len(room.View().Participants); got != 1 {
		t.Errorf("table has %d participants, want 1 (only the accepted name)", got)
	}
}

func TestUnknownParticipantIsRefusedByEveryOperation(t *testing.T) {
	room := newTestRoom(t)
	seated := join(t, room, "Thomas")
	if err := room.Vote(seated, CardM); err != nil {
		t.Fatalf("Vote: %v", err)
	}

	const stranger ParticipantID = "NOTSEATEDHERE"

	operations := map[string]func() error{
		"Vote":        func() error { return room.Vote(stranger, CardL) },
		"SetDeck":     func() error { return room.SetDeck(stranger, FibonacciDeckName) },
		"Reveal":      func() error { return room.Reveal(stranger) },
		"NewRound":    func() error { return room.NewRound(stranger) },
		"Rename":      func() error { return room.Rename(stranger, "Mallory") },
		"Rejoin":      func() error { return room.Rejoin(stranger, "Mallory") },
		"MarkAway":    func() error { return room.MarkAway(stranger) },
		"MarkPresent": func() error { return room.MarkPresent(stranger) },
	}

	before := room.View()
	for name, op := range operations {
		if err := op(); !errors.Is(err, ErrUnknownParticipant) {
			t.Errorf("%s with an unseated identifier: error = %v, want ErrUnknownParticipant", name, err)
		}
	}

	after := room.View()
	if len(after.Participants) != len(before.Participants) {
		t.Errorf("the table changed size: %d participants, want %d",
			len(after.Participants), len(before.Participants))
	}
	if after.Revealed != before.Revealed {
		t.Error("a refused operation changed whether the round was revealed")
	}
	if !after.Participants[0].Voted {
		t.Error("a refused operation discarded the seated participant's vote")
	}
}

func TestRejoiningKeepsSeatNameAndVote(t *testing.T) {
	room := newTestRoom(t)
	other := join(t, room, "Anna")
	id := join(t, room, "Thomas")

	if err := room.Vote(id, CardL); err != nil {
		t.Fatalf("Vote: %v", err)
	}
	if err := room.MarkAway(id); err != nil {
		t.Fatalf("MarkAway: %v", err)
	}

	// This is what a page reload looks like from here.
	if err := room.Rejoin(id, "Thomas"); err != nil {
		t.Fatalf("Rejoin: %v", err)
	}

	view := room.View()
	if len(view.Participants) != 2 {
		t.Fatalf("table has %d participants after a reload, want 2 — a reload must not duplicate anyone",
			len(view.Participants))
	}
	// Seating order is visible at the table and must be stable across a reload.
	if view.Participants[0].ID != other || view.Participants[1].ID != id {
		t.Error("the seating order changed when a participant rejoined")
	}

	rejoined := view.Participants[1]
	if rejoined.Name != "Thomas" {
		t.Errorf("name after rejoining = %q, want %q", rejoined.Name, "Thomas")
	}
	if !rejoined.Voted {
		t.Error("the vote was lost when the participant rejoined")
	}
	if rejoined.Away {
		t.Error("the participant is still marked away after rejoining")
	}

	if err := room.Reveal(id); err != nil {
		t.Fatalf("Reveal: %v", err)
	}
	for _, c := range room.View().Results.Cards {
		if c.ID == id && c.Card != CardL {
			t.Errorf("card after rejoining = %q, want %q", c.Card, CardL)
		}
	}
}

func TestRejoiningWithANewNameUpdatesTheName(t *testing.T) {
	room := newTestRoom(t)
	id := join(t, room, "Thomas")

	if err := room.Rejoin(id, "  Thomas N.  "); err != nil {
		t.Fatalf("Rejoin: %v", err)
	}

	if got := nameOf(t, room, id); got != "Thomas N." {
		t.Errorf("name = %q, want %q", got, "Thomas N.")
	}
	if got := len(room.View().Participants); got != 1 {
		t.Errorf("table has %d participants, want 1", got)
	}
}

func TestRejoiningWithAnInvalidNameIsRefused(t *testing.T) {
	room := newTestRoom(t)
	id := join(t, room, "Thomas")

	if err := room.Rejoin(id, "   "); !errors.Is(err, ErrEmptyName) {
		t.Errorf("Rejoin with an empty name: error = %v, want ErrEmptyName", err)
	}
	if got := nameOf(t, room, id); got != "Thomas" {
		t.Errorf("name after a refused rejoin = %q, want the previous name %q", got, "Thomas")
	}
}

func TestAwayKeepsSeatNameAndVote(t *testing.T) {
	room := newTestRoom(t)
	id := join(t, room, "Thomas")
	if err := room.Vote(id, CardM); err != nil {
		t.Fatalf("Vote: %v", err)
	}

	if err := room.MarkAway(id); err != nil {
		t.Fatalf("MarkAway: %v", err)
	}

	view := room.View()
	if len(view.Participants) != 1 {
		t.Fatalf("table has %d participants, want 1 — going away must not remove anyone",
			len(view.Participants))
	}
	p := view.Participants[0]
	if !p.Away {
		t.Error("the participant is not marked away")
	}
	if p.Name != "Thomas" {
		t.Errorf("name = %q, want %q", p.Name, "Thomas")
	}
	if !p.Voted {
		t.Error("the vote was discarded when the participant went away")
	}
}

func TestMarkingPresentClearsAwayAndChangesNothingElse(t *testing.T) {
	room := newTestRoom(t)
	id := join(t, room, "Thomas")
	if err := room.Vote(id, CardM); err != nil {
		t.Fatalf("Vote: %v", err)
	}
	if err := room.MarkAway(id); err != nil {
		t.Fatalf("MarkAway: %v", err)
	}

	before := room.View()
	if err := room.MarkPresent(id); err != nil {
		t.Fatalf("MarkPresent: %v", err)
	}
	after := room.View()

	if after.Participants[0].Away {
		t.Error("the away mark was not cleared")
	}
	if after.Participants[0].Name != before.Participants[0].Name {
		t.Error("the name changed when the participant returned")
	}
	if !after.Participants[0].Voted {
		t.Error("the vote was lost when the participant returned")
	}
}

func TestNobodyIsRemovedByTheirAbsence(t *testing.T) {
	// There is no timer anywhere in this package, so however long a participant is
	// away, they are still at the table. This test states that as a rule rather than
	// letting it be an accident of there being no code to remove them.
	room := newTestRoom(t)
	id := join(t, room, "Thomas")
	if err := room.MarkAway(id); err != nil {
		t.Fatalf("MarkAway: %v", err)
	}

	// Any number of unrelated operations, standing in for the passage of time.
	for range 100 {
		_ = room.View()
		_ = room.AnyonePresent()
	}

	if got := len(room.View().Participants); got != 1 {
		t.Errorf("table has %d participants, want 1 — nothing may remove a participant on a timer", got)
	}
}

func TestRenamingChangesOnlyTheName(t *testing.T) {
	room := newTestRoom(t)
	id := join(t, room, "Thomas")
	if err := room.Vote(id, CardXL); err != nil {
		t.Fatalf("Vote: %v", err)
	}

	if err := room.Rename(id, "Thomas N."); err != nil {
		t.Fatalf("Rename: %v", err)
	}

	view := room.View()
	if view.Participants[0].ID != id {
		t.Error("the identifier changed when the participant was renamed")
	}
	if view.Participants[0].Name != "Thomas N." {
		t.Errorf("name = %q, want %q", view.Participants[0].Name, "Thomas N.")
	}
	if !view.Participants[0].Voted {
		t.Error("the vote was lost when the participant was renamed")
	}
}

func TestRenamingToAnInvalidNameIsRefused(t *testing.T) {
	room := newTestRoom(t)
	id := join(t, room, "Thomas")

	if err := room.Rename(id, ""); !errors.Is(err, ErrEmptyName) {
		t.Errorf("Rename to empty: error = %v, want ErrEmptyName", err)
	}
	if err := room.Rename(id, strings.Repeat("a", MaxNameLength+1)); !errors.Is(err, ErrNameTooLong) {
		t.Errorf("Rename to an over-long name: error = %v, want ErrNameTooLong", err)
	}
	if got := nameOf(t, room, id); got != "Thomas" {
		t.Errorf("name after refused renames = %q, want the previous name %q", got, "Thomas")
	}
}

func TestRenamingDuringARevealedRoundIsAllowed(t *testing.T) {
	room := newTestRoom(t)
	id := join(t, room, "Thomas")
	if err := room.Vote(id, CardS); err != nil {
		t.Fatalf("Vote: %v", err)
	}
	if err := room.Reveal(id); err != nil {
		t.Fatalf("Reveal: %v", err)
	}

	if err := room.Rename(id, "Thomas N."); err != nil {
		t.Fatalf("Rename during a revealed round: %v", err)
	}

	view := room.View()
	if view.Participants[0].Name != "Thomas N." {
		t.Errorf("name = %q, want %q", view.Participants[0].Name, "Thomas N.")
	}
	if view.Results.Cards[0].Card != CardS {
		t.Errorf("the revealed card changed on rename: %q, want %q", view.Results.Cards[0].Card, CardS)
	}
}

func TestAnyonePresent(t *testing.T) {
	room := newTestRoom(t)

	if room.AnyonePresent() {
		t.Error("an empty room reports somebody present")
	}

	first := join(t, room, "Thomas")
	second := join(t, room, "Anna")
	if !room.AnyonePresent() {
		t.Error("a room with two seated participants reports nobody present")
	}

	if err := room.MarkAway(first); err != nil {
		t.Fatalf("MarkAway: %v", err)
	}
	if !room.AnyonePresent() {
		t.Error("a room with one away and one present participant reports nobody present")
	}

	if err := room.MarkAway(second); err != nil {
		t.Fatalf("MarkAway: %v", err)
	}
	if room.AnyonePresent() {
		t.Error("a room whose participants are all away reports somebody present")
	}
}

// --- the seat limit -------------------------------------------------------

// newTestRoomSeating creates a room with a chosen number of seats, for the tests
// that are about the limit rather than about the game.
func newTestRoomSeating(t *testing.T, seats int) *Room {
	t.Helper()
	room, err := NewRoom(rand.Reader, seats)
	if err != nil {
		t.Fatalf("NewRoom with %d seats: %v", seats, err)
	}
	return room
}

func TestAFullTableRefusesANewSeat(t *testing.T) {
	room := newTestRoomSeating(t, 2)
	join(t, room, "Thomas")
	join(t, room, "Anna")

	_, err := room.Join(rand.Reader, "Late")
	if !errors.Is(err, ErrRoomFull) {
		t.Fatalf("joining a full table returned %v, want ErrRoomFull", err)
	}

	if got := len(room.View().Participants); got != 2 {
		t.Errorf("the table holds %d participants after a refused join, want 2 — "+
			"nobody may be unseated to make space", got)
	}
}

func TestAnAwayParticipantStillOccupiesTheirSeat(t *testing.T) {
	// An away participant keeps their seat, their name and their vote, and is still
	// shown at the table. Freeing their seat on a dropped connection would let
	// somebody else take it while they were reconnecting.
	room := newTestRoomSeating(t, 1)
	only := join(t, room, "Thomas")

	if err := room.MarkAway(only); err != nil {
		t.Fatalf("MarkAway: %v", err)
	}

	if _, err := room.Join(rand.Reader, "Opportunist"); !errors.Is(err, ErrRoomFull) {
		t.Errorf("joining while the only seat was held by an away participant returned %v, "+
			"want ErrRoomFull", err)
	}
}

func TestRejoiningIsNeverRefusedForCapacity(t *testing.T) {
	// This is the case that would hurt most if it were wrong: reconnection failing
	// exactly when a room is busiest. A returning browser occupies no new seat.
	room := newTestRoomSeating(t, 1)
	only := join(t, room, "Thomas")

	if err := room.MarkAway(only); err != nil {
		t.Fatalf("MarkAway: %v", err)
	}
	if err := room.Rejoin(only, "Thomas"); err != nil {
		t.Fatalf("rejoining a full room was refused with %v; it takes no new seat", err)
	}

	if got := len(room.View().Participants); got != 1 {
		t.Errorf("rejoining produced %d participants, want 1", got)
	}
}

func TestARefusedSeatLeavesTheRoomExactlyAsItWas(t *testing.T) {
	room := newTestRoomSeating(t, 1)
	seated := join(t, room, "Thomas")
	if err := room.Vote(seated, CardM); err != nil {
		t.Fatalf("Vote: %v", err)
	}

	before := room.View()
	if _, err := room.Join(rand.Reader, "Late"); !errors.Is(err, ErrRoomFull) {
		t.Fatalf("Join returned %v, want ErrRoomFull", err)
	}
	after := room.View()

	if !reflect.DeepEqual(before, after) {
		t.Errorf("a refused seat changed the room:\nbefore %+v\nafter  %+v", before, after)
	}
}

func TestARoomNeedsAPositiveSeatLimit(t *testing.T) {
	// There is deliberately no value meaning "no limit": an unbounded table is the
	// condition the limit exists to prevent.
	for _, seats := range []int{0, -1} {
		if _, err := NewRoom(rand.Reader, seats); !errors.Is(err, ErrInvalidCapacity) {
			t.Errorf("NewRoom with %d seats returned %v, want ErrInvalidCapacity", seats, err)
		}
	}
}
