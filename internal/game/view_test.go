package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// cardValuesReachableFrom walks a value and reports every non-empty Card it can
// reach, by the path it was reached along.
//
// Values of type Deck are skipped deliberately, and this is the subtlety that makes
// the guard test below work at all. The deck is part of the view on purpose — the
// client has to know which cards to offer — so every card value legitimately
// appears there. A naive search for card values in the rendered view would
// therefore always find them and could never fail. Skipping the deck leaves exactly
// the question that matters: has a card reached anywhere else?
func cardValuesReachableFrom(v reflect.Value, path string, found *[]string) {
	if !v.IsValid() {
		return
	}

	if v.Type() == reflect.TypeOf(Deck{}) {
		return
	}

	switch v.Kind() {
	case reflect.Struct:
		for i := range v.NumField() {
			cardValuesReachableFrom(v.Field(i), path+"."+v.Type().Field(i).Name, found)
		}
	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			cardValuesReachableFrom(v.Index(i), fmt.Sprintf("%s[%d]", path, i), found)
		}
	case reflect.Pointer, reflect.Interface:
		if !v.IsNil() {
			cardValuesReachableFrom(v.Elem(), path, found)
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			cardValuesReachableFrom(v.MapIndex(k), fmt.Sprintf("%s[%v]", path, k), found)
		}
	default:
		if v.Type() == reflect.TypeOf(Card("")) && v.String() != "" {
			*found = append(*found, fmt.Sprintf("%s = %q", path, v.String()))
		}
	}
}

// votedRoom builds a room in which every participant has played a different card,
// so that a leak of any one of them is visible.
func votedRoom(t *testing.T) (*Room, []ParticipantID) {
	t.Helper()
	room := newTestRoom(t)

	// Deliberately lowercase names: the search below looks for uppercase card
	// values, and a name containing one would be a false alarm rather than a leak.
	names := []string{"anna", "bert", "cora", "dora", "emil"}
	cards := []Card{CardXS, CardS, CardM, CardL, CardXL}

	ids := make([]ParticipantID, 0, len(names))
	for i, name := range names {
		id := join(t, room, name)
		if err := room.Vote(id, cards[i]); err != nil {
			t.Fatalf("Vote(%q): %v", cards[i], err)
		}
		ids = append(ids, id)
	}
	return room, ids
}

// TestHiddenRoundNeverDisclosesAVote is the most important test in this package.
//
// The rule it defends is the reason votes are hidden at all: transmitting a value
// and hiding it in the interface would put every vote one developer-tools panel away
// from anyone at the table, which defeats simultaneous estimation entirely.
//
// If this test ever fails, the fix is in the code that leaked the value. It must
// never be weakened to make a build pass.
func TestHiddenRoundNeverDisclosesAVote(t *testing.T) {
	room, _ := votedRoom(t)
	view := room.View()

	if view.Revealed {
		t.Fatal("the round is revealed; this test is checking the hidden case")
	}

	var found []string
	cardValuesReachableFrom(reflect.ValueOf(view), "View", &found)
	if len(found) > 0 {
		t.Errorf("a hidden round disclosed card values:\n  %s", strings.Join(found, "\n  "))
	}

	// Everyone must still be shown as having voted — hiding the value must not hide
	// the fact, or nobody would know when the table is ready.
	for _, p := range view.Participants {
		if !p.Voted {
			t.Errorf("participant %q voted but is not shown as having voted", p.Name)
		}
	}
}

func TestHiddenRoundLeaksNothingWhenSerialised(t *testing.T) {
	// The same rule approached from the other side: whatever the transport layer
	// does with this value, no card may come out of it. JSON stands in for
	// "serialised by any means", and is what the transport layer will actually use.
	room, _ := votedRoom(t)
	view := room.View()

	// The deck legitimately holds every card value, and the identifiers are rendered
	// in an uppercase alphabet in which a sequence like "XS" can occur by chance.
	// Blanking all three leaves only the fields where a card would be a genuine leak.
	view.Deck = Deck{}
	view.RoomID = ""
	for i := range view.Participants {
		view.Participants[i].ID = ""
	}

	encoded, err := json.Marshal(view)
	if err != nil {
		t.Fatalf("marshalling the view: %v", err)
	}

	// Searched over the values only, not over the raw text. A JSON object's keys are
	// field names chosen by this package, not data the view is disclosing, and some of
	// them contain a card by coincidence — "Scale" contains "S", which is a card of
	// this very deck. Matching against the raw bytes would report that as a leak and
	// leave the test failing for a reason that has nothing to do with anybody's vote.
	var decoded any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshalling the view: %v", err)
	}

	for _, card := range TShirtDeck().Cards {
		if where := findCardInValues(decoded, string(card), "view"); where != "" {
			t.Errorf("the serialised hidden view contains the card %q at %s:\n%s", card, where, encoded)
		}
	}
}

// findCardInValues walks decoded JSON and returns the path at which some string
// value contains the card, or "" if none does. Object keys are walked into but never
// themselves matched, for the reason given at the call site.
func findCardInValues(value any, card, path string) string {
	switch v := value.(type) {
	case string:
		if strings.Contains(v, card) {
			return path
		}
	case []any:
		for i, item := range v {
			if where := findCardInValues(item, card, fmt.Sprintf("%s[%d]", path, i)); where != "" {
				return where
			}
		}
	case map[string]any:
		for key, item := range v {
			if where := findCardInValues(item, card, path+"."+key); where != "" {
				return where
			}
		}
	}
	return ""
}

func TestHiddenRoundCarriesNoTally(t *testing.T) {
	// A count over a table this small would let the individual votes be worked out,
	// so the tally is withheld exactly as the values are.
	room, _ := votedRoom(t)

	if results := room.View().Results; results != nil {
		t.Errorf("a hidden round carries results: %+v", results)
	}
}

func TestParticipantViewCannotCarryACard(t *testing.T) {
	// The guarantee is structural, not remembered: there is nowhere in a
	// ParticipantView to put a card, so a hidden round cannot leak one by accident.
	// If this fails, someone added a field and the guarantee now depends on their
	// remembering to leave it empty.
	pv := reflect.TypeOf(ParticipantView{})
	for i := range pv.NumField() {
		if pv.Field(i).Type == reflect.TypeOf(Card("")) {
			t.Errorf("ParticipantView has a card-typed field %q; card values belong in Results, "+
				"which is nil while the round is hidden", pv.Field(i).Name)
		}
	}
}

func TestRevealingDisclosesEveryVote(t *testing.T) {
	room, ids := votedRoom(t)
	silent := join(t, room, "frank")

	if err := room.Reveal(ids[0]); err != nil {
		t.Fatalf("Reveal: %v", err)
	}

	view := room.View()
	if view.Results == nil {
		t.Fatal("a revealed round carries no results")
	}

	want := map[ParticipantID]Card{
		ids[0]: CardXS, ids[1]: CardS, ids[2]: CardM, ids[3]: CardL, ids[4]: CardXL,
		silent: NoCard,
	}
	if len(view.Results.Cards) != len(want) {
		t.Fatalf("results hold %d cards, want %d", len(view.Results.Cards), len(want))
	}
	for _, c := range view.Results.Cards {
		if c.Card != want[c.ID] {
			t.Errorf("participant %q holds %q, want %q", c.ID, c.Card, want[c.ID])
		}
	}
}

func TestTallyCountsEachPlayedCardInDeckOrder(t *testing.T) {
	room := newTestRoom(t)
	first := join(t, room, "anna")
	second := join(t, room, "bert")
	third := join(t, room, "cora")
	fourth := join(t, room, "dora")

	for id, card := range map[ParticipantID]Card{
		first: CardM, second: CardM, third: CardL, fourth: CardBreak,
	} {
		if err := room.Vote(id, card); err != nil {
			t.Fatalf("Vote: %v", err)
		}
	}
	if err := room.Reveal(first); err != nil {
		t.Fatalf("Reveal: %v", err)
	}

	want := []CardCount{
		{Card: CardM, Count: 2},
		{Card: CardL, Count: 1},
		{Card: CardBreak, Count: 1},
	}
	got := room.View().Results.Tally
	if len(got) != len(want) {
		t.Fatalf("tally = %+v, want %+v", got, want)
	}
	// The order is the deck's own, which is why M comes before L and the coffee cup
	// comes last, rather than any order the votes happened to arrive in.
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("tally[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestCardsNobodyPlayedAreAbsentFromTheTally(t *testing.T) {
	room := newTestRoom(t)
	id := join(t, room, "anna")
	if err := room.Vote(id, CardM); err != nil {
		t.Fatalf("Vote: %v", err)
	}
	if err := room.Reveal(id); err != nil {
		t.Fatalf("Reveal: %v", err)
	}

	tally := room.View().Results.Tally
	if len(tally) != 1 {
		t.Fatalf("tally has %d entries, want 1 — unplayed cards must be absent entirely", len(tally))
	}
	for _, entry := range tally {
		if entry.Count == 0 {
			t.Errorf("card %q appears in the tally with a count of zero; it should be absent", entry.Card)
		}
	}
}

func TestNonVotersAreExcludedFromTheTally(t *testing.T) {
	room := newTestRoom(t)
	voter := join(t, room, "anna")
	silent := join(t, room, "bert")
	if err := room.Vote(voter, CardM); err != nil {
		t.Fatalf("Vote: %v", err)
	}
	if err := room.Reveal(voter); err != nil {
		t.Fatalf("Reveal: %v", err)
	}

	results := room.View().Results
	if got := cardOf(t, room, silent); got != NoCard {
		t.Errorf("the non-voter holds %q, want no card", got)
	}

	total := 0
	for _, entry := range results.Tally {
		total += entry.Count
	}
	if total != 1 {
		t.Errorf("the tally counts %d votes, want 1 — non-voters must be counted in no entry", total)
	}
}

func TestResultsContainNoArithmeticOverTheCards(t *testing.T) {
	// T-shirt sizes are an ordered scale without arithmetic, and "?" and the coffee
	// cup are not sizes at all. Any average would be an invented number wearing the
	// appearance of a measurement, so there must be nowhere to put one.
	results := reflect.TypeOf(Results{})
	allowed := map[string]bool{"Cards": true, "Tally": true}

	for i := range results.NumField() {
		name := results.Field(i).Name
		if !allowed[name] {
			t.Errorf("Results has an unexpected field %q; if this is a derived number "+
				"over the cards, it does not belong here", name)
		}
	}

	for _, forbidden := range []string{"Average", "Mean", "Median", "Sum", "Consensus", "Estimate"} {
		if _, present := results.FieldByName(forbidden); present {
			t.Errorf("Results has a %q field; no arithmetic over the cards may be reported", forbidden)
		}
	}
}

func TestEveryRefusalReturnsARecognisableError(t *testing.T) {
	// The transport layer has to turn each refusal into its own message for the
	// browser. Matching on error text would break the first time one of these
	// sentences is reworded, so every refusal must be identifiable with errors.Is.
	room := newTestRoom(t)
	id := join(t, room, "anna")

	sentinels := []error{
		ErrUnknownParticipant, ErrEmptyName, ErrNameTooLong,
		ErrCardNotInDeck, ErrRoundRevealed, ErrShortRandomRead,
	}

	refusals := map[string]func() error{
		"vote by a stranger":     func() error { return room.Vote("nobody", CardM) },
		"empty name on join":     func() error { _, err := room.Join(fixedRandom(1), " "); return err },
		"over-long name on join": func() error { _, err := room.Join(fixedRandom(1), strings.Repeat("a", MaxNameLength+1)); return err },
		"card outside the deck":  func() error { return room.Vote(id, "XXL") },
		"broken randomness":      func() error { _, err := room.Join(failingRandom{}, "bert"); return err },
		"empty name on rename":   func() error { return room.Rename(id, "") },
		"rejoin by a stranger":   func() error { return room.Rejoin("nobody", "mallory") },
	}

	for name, refuse := range refusals {
		err := refuse()
		if err == nil {
			t.Errorf("%s: no error was returned", name)
			continue
		}
		if !matchesAny(err, sentinels) {
			t.Errorf("%s: error %v matches none of the package's sentinel errors", name, err)
		}
	}

	// And the refusal that only exists once a round is revealed.
	if err := room.Reveal(id); err != nil {
		t.Fatalf("Reveal: %v", err)
	}
	if err := room.Vote(id, CardM); !matchesAny(err, sentinels) {
		t.Errorf("vote after reveal: error %v matches none of the package's sentinel errors", err)
	}
}

func matchesAny(err error, sentinels []error) bool {
	for _, s := range sentinels {
		if errors.Is(err, s) {
			return true
		}
	}
	return false
}
