package main

import (
	"math/rand"
	"testing"

	"zcore.dev/voinich/internal/evaglyph"
)

func TestPositions(t *testing.T) {
	c := count([][]string{{"a", "b", "c"}, {"x"}, {"d", "e"}})
	if c["a"].I != 1 || c["b"].M != 1 || c["c"].F != 1 || c["x"].S != 1 || c["d"].I != 1 || c["e"].F != 1 {
		t.Fatalf("unexpected positions: %#v", c)
	}
}
func TestEVAComposites(t *testing.T) {
	g := evaglyph.CollapseEVA("cthckhcphcfhiinainchsheein")
	want := []string{"C", "K", "P", "F", "N", "A", "H", "S", "E", "I"}
	if len(g) != len(want) {
		t.Fatalf("got %v", g)
	}
	for i := range want {
		if g[i] != want[i] {
			t.Fatalf("%v != %v", g, want)
		}
	}
}
func TestShufflesPreserveStructure(t *testing.T) {
	x := [][]string{{"a", "b", "a"}, {"c", "d"}}
	for _, global := range []bool{false, true} {
		y := shuffled(x, rand.New(rand.NewSource(1)), global)
		if len(y) != len(x) || len(y[0]) != 3 || len(y[1]) != 2 {
			t.Fatal("lengths changed")
		}
		cx, cy := count(x), count(y)
		if cx["a"].N != cy["a"].N || cx["d"].N != cy["d"].N {
			t.Fatal("frequencies changed")
		}
	}
}
func TestFDRMonotonic(t *testing.T) {
	r := []Row{{P: .01}, {P: .02}, {P: .5}}
	fdr(r)
	if r[0].Q > r[1].Q || r[1].Q > r[2].Q {
		t.Fatal("q values not monotone")
	}
}
func TestSpecialist(t *testing.T) {
	r, _, _ := analyze(Corpus{Tokens: [][]string{{"i", "m", "f"}, {"i", "m", "f"}}}, 10)
	for _, x := range r {
		if x.Glyph == "i" && (x.Dominant != "INITIAL" || x.Share != 1) {
			t.Fatalf("%+v", x)
		}
	}
}

// hom's position-independent branch (pos=false) must draw each
// occurrence's homophone index independently of the glyph's within-token
// position i. An earlier version derived the index from i (or from a
// running occurrence counter tied to i within each token), which
// deterministically re-created a position signal in what is supposed to
// be task59's negative control (sections 17-18). Build a corpus where
// glyph "e" always sits at the same token-internal index across many
// tokens of varying length, and require its homophone labels to spread
// across the H pool rather than collapse onto a single index-derived
// label.
func TestPositionIndependentHomophonySpreadsAcrossPool(t *testing.T) {
	c := Corpus{}
	for n := 0; n < 400; n++ {
		c.Tokens = append(c.Tokens, []string{"x", "e", "y", "z"}) // "e" always at i=1
	}
	h := hom(c, 4, false, rand.New(rand.NewSource(1)))
	seen := map[string]int{}
	for _, t := range h.Tokens {
		seen[t[1]]++
	}
	if len(seen) < 2 {
		t.Fatalf("position-independent homophony collapsed onto a single label per position: %v", seen)
	}
	for label, n := range seen {
		if n > 350 {
			t.Fatalf("label %q dominates (%d/400): selection looks tied to token position, not independent", label, n)
		}
	}
}

// hom's position-dependent branch (pos=true) is the positive control
// (task59 section 21): it must be detectable as producing strict
// per-position specialists, unlike the position-independent branch.
func TestPositionDependentHomophonyPositiveControlDetected(t *testing.T) {
	c := Corpus{}
	for n := 0; n < 200; n++ {
		c.Tokens = append(c.Tokens, []string{"e", "m", "e"}) // e: INITIAL and FINAL both
	}
	h := hom(c, 4, true, rand.New(rand.NewSource(1)))
	rows, _, _ := analyze(h, 5)
	found := false
	for _, r := range rows {
		if r.N < 10 {
			continue
		}
		found = true
		if r.Share != 1 {
			t.Fatalf("position-dependent control glyph %q is not a strict specialist: %+v", r.Glyph, r)
		}
	}
	if !found {
		t.Fatal("no sufficiently frequent synthetic glyph produced by the positive control")
	}
}
