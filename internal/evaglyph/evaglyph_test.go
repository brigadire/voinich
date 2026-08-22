package evaglyph

import (
	"math/rand"
	"testing"
)

func TestCollapseEVA(t *testing.T) {
	got := CollapseEVA("CTHCKHCPHCFHIINAINCHSHEEIN")
	want := []string{"C", "K", "P", "F", "N", "A", "H", "S", "E", "I"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestCollapseEVAPlainGlyphsUnaffected(t *testing.T) {
	got := CollapseEVA("qokedy")
	want := []string{"q", "o", "k", "e", "d", "y"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestNaturalGlyphsStripsPunctuation(t *testing.T) {
	got := NaturalGlyphs("Doyle's, 2nd!")
	want := []string{"d", "o", "y", "l", "e", "s", "2", "n", "d"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		n, i int
		want string
	}{
		{1, 0, "SINGLETON"},
		{2, 0, "INITIAL"},
		{2, 1, "FINAL"},
		{3, 0, "INITIAL"},
		{3, 1, "MEDIAL"},
		{3, 2, "FINAL"},
	}
	for _, c := range cases {
		if got := Classify(c.n, c.i); got != c.want {
			t.Fatalf("Classify(%d,%d) = %s, want %s", c.n, c.i, got, c.want)
		}
	}
}

func TestRandomHomophonyPreservesShapeAndIsPositionIndependent(t *testing.T) {
	tokens := [][]string{{"a", "b"}, {"c"}, {"a", "b", "a"}}
	out := RandomHomophony(tokens, 4, rand.New(rand.NewSource(1)))
	if len(out) != len(tokens) {
		t.Fatalf("token count changed: %d != %d", len(out), len(tokens))
	}
	for i := range tokens {
		if len(out[i]) != len(tokens[i]) {
			t.Fatalf("token %d length changed: %d != %d", i, len(out[i]), len(tokens[i]))
		}
	}
	// Same fixed within-token index (i=0), many occurrences of glyph "a" at
	// varying global positions: labels must not collapse onto one value.
	var many [][]string
	for n := 0; n < 200; n++ {
		many = append(many, []string{"a", "x"})
	}
	out2 := RandomHomophony(many, 4, rand.New(rand.NewSource(2)))
	seen := map[string]bool{}
	for _, t := range out2 {
		seen[t[0]] = true
	}
	if len(seen) < 2 {
		t.Fatalf("homophone labels collapsed onto a single value: %v", seen)
	}
}
func TestMIKnownValues(t *testing.T) {
	// independent (x always matches its own index parity, y is
	// independently alternating) -> a b a b a b paired with a b a b a b:
	// perfectly dependent, MI == H(x) == 1 bit.
	x := []string{"a", "b", "a", "b", "a", "b"}
	if mi := MI(x, x); mi < 0.99 {
		t.Fatalf("expected ~1 bit for identical alternating sequences, got %v", mi)
	}
	// constant y carries no information: MI == 0.
	y := []string{"z", "z", "z", "z", "z", "z"}
	if mi := MI(x, y); mi != 0 {
		t.Fatalf("expected 0 MI against a constant sequence, got %v", mi)
	}
}
