package fontanafamily

import (
	"encoding/json"
	"testing"
)

var latin23 = []rune("ABCDEFGHIKLMNOPQRSTVXYZ")

func TestSerpensRoundTrip(t *testing.T) {
	c := SerpensConfig{Capacity: 12, Alphabet: latin23, Start: 0, Direction: Forward, EmptyMarker: '?'}
	s, err := c.Encode("MEMORIA")
	if err != nil {
		t.Fatal(err)
	}
	got, err := c.Decode(s, 7)
	if err != nil {
		t.Fatal(err)
	}
	if got != "MEMORIA" {
		t.Fatalf("got %q", got)
	}
}

func TestSerpensStateJSONUsesVisibleSymbols(t *testing.T) {
	c := SerpensConfig{Capacity: 3, Alphabet: latin23, Start: 0, Direction: Forward}
	s, _ := c.Encode("AB")
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(b); got != `{"holes":["A","B",null]}` {
		t.Fatalf("state JSON=%s", got)
	}
}

func TestSerpensAblationsAndDamage(t *testing.T) {
	c := SerpensConfig{Capacity: 8, Alphabet: latin23, Start: 0, Direction: Forward, EmptyMarker: '?'}
	s, _ := c.Encode("FONTANA")
	starts, _ := c.CompatibleTraversals(s, 7, false, true)
	if len(starts) != 2 {
		t.Fatalf("unknown start produced %d distinct candidates, want 2", len(starts))
	}
	dirs, _ := c.CompatibleTraversals(s, 7, true, false)
	if len(dirs) != 2 {
		t.Fatalf("unknown direction produced %d candidates, want 2", len(dirs))
	}
	removed := SerpensRemove(s, 2)
	got, _ := c.Decode(removed, 7)
	if got != "FO?TANA" {
		t.Fatalf("local removal got %q", got)
	}
	collapsed := SerpensCollapse(s, 2)
	got, _ = c.Decode(collapsed, 7)
	if got != "FOTANA?" {
		t.Fatalf("collapse got %q", got)
	}
}

func TestSerpensLinearTraversalDoesNotWrap(t *testing.T) {
	c := SerpensConfig{Capacity: 8, Alphabet: latin23, Start: 7, Direction: Forward}
	if _, err := c.Encode("AB"); err == nil {
		t.Fatal("linear spiral must not wrap edge to centre")
	}
	c.Direction = Reverse
	s, err := c.Encode("AB")
	if err != nil {
		t.Fatal(err)
	}
	got, err := c.Decode(s, 2)
	if err != nil || got != "AB" {
		t.Fatalf("reverse linear traversal got %q, %v", got, err)
	}
}

func TestNormalizedEditDistance(t *testing.T) {
	if got := NormalizedEditDistance("FONTANA", "FOTANA"); got != 1.0/7.0 {
		t.Fatalf("normalized edit distance=%v", got)
	}
}

func TestRotaCycle(t *testing.T) {
	r := Rota{Alphabet: latin23}
	if r.Period(1) != 23 {
		t.Fatalf("period=%d", r.Period(1))
	}
	for i := 0; i < 23; i++ {
		r = r.Rotate(1)
	}
	if r.Offset != 0 {
		t.Fatalf("offset=%d", r.Offset)
	}
}

func TestCylinderRoundTripAndIndependentTransition(t *testing.T) {
	c := Cylinder{Alphabet: latin23, Offsets: make([]int, 7)}
	s, err := c.Encode("MEMORIA", Forward)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := s.Read(Forward)
	if got != "MEMORIA" {
		t.Fatalf("got %q", got)
	}
	changed, _ := s.RotateBand(2, 1)
	if changed.Offsets[2] == s.Offsets[2] {
		t.Fatal("selected band did not change")
	}
	for i := range s.Offsets {
		if i != 2 && changed.Offsets[i] != s.Offsets[i] {
			t.Fatal("rotation leaked to another band")
		}
	}
	if s.StateSpaceSize() != 3404825447 {
		t.Fatalf("state space=%d", s.StateSpaceSize())
	}
}

func TestArismetricumLookupAndCorruption(t *testing.T) {
	a := Arismetricum{Slots: map[int]string{1: "LABOR", 2: "MEMORIA"}}
	if got, ok := a.Lookup(2); !ok || got != "MEMORIA" {
		t.Fatalf("lookup=%q,%v", got, ok)
	}
	if _, ok := a.Remove(2).Lookup(2); ok {
		t.Fatal("removed slot still present")
	}
	if got, _ := a.Swap(1, 2).Lookup(1); got != "MEMORIA" {
		t.Fatalf("swap got %q", got)
	}
}

func TestHoralogiusTransitionAndLearnedRecall(t *testing.T) {
	h := Horalogius{Period: 12, Tick: 10, Cues: map[int]string{0: "BELL"}}
	h, cues, err := h.Advance(2)
	if err != nil {
		t.Fatal(err)
	}
	if h.Tick != 0 || len(cues) != 1 || cues[0] != "BELL" {
		t.Fatalf("state=%+v cues=%v", h, cues)
	}
	if _, ok := Recall("BELL", nil); ok {
		t.Fatal("untrained decoder recovered meaning")
	}
	if got, ok := Recall("BELL", map[string]string{"BELL": "LABOR"}); !ok || got != "LABOR" {
		t.Fatalf("recall=%q,%v", got, ok)
	}
}
