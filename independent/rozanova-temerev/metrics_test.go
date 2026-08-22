package main

import "testing"

func toy(lines ...[]string) Corpus { return Corpus{Lines: lines, GlyphMode: "text"} }
func TestKnownDependentAndIndependentMI(t *testing.T) {
	a := toy([]string{"a", "b", "a", "b", "a", "b"})
	x, _ := analyse(a, 2000, 8, 1)
	if x.TokenObserved < 0.9 {
		t.Fatalf("dependent MI=%v", x.TokenObserved)
	}
}
func TestDeterministicAndOpaqueEdge(t *testing.T) {
	c := toy([]string{"alpha", "beta", "alpha", "beta"})
	a, _ := analyse(c, 2, 10, 42)
	b, _ := analyse(c, 2, 10, 42)
	if a != b {
		t.Fatal("analysis is not deterministic")
	}
	c.GlyphMode = "opaque"
	z, _ := analyse(c, 2, 2, 42)
	if z.GlyphStatus != "NOT_APPLICABLE_OPAQUE_TOKENS" || z.EdgeCorrected != 0 {
		t.Fatalf("opaque edge was calculated: %+v", z)
	}
}
func TestLineBoundariesAndCap(t *testing.T) {
	c := toy([]string{"a", "b"}, []string{"a", "b"})
	m, _ := analyse(c, 10, 3, 1)
	if m.Pairs != 2 {
		t.Fatalf("cross-line pair included: %d", m.Pairs)
	}
	c = toy([]string{"rare", "common", "common"})
	m, _ = analyse(c, 1, 2, 1)
	if m.TokenCorrected < 0 {
		t.Fatal("cap produced invalid corrected MI")
	}
}
func TestVoynichCompositeGlyphs(t *testing.T) {
	if got := feature("cthaiin", "voynich", "first"); got != "C" {
		t.Fatalf("first collapsed glyph=%q", got)
	}
	if got := feature("cthaiin", "voynich", "last"); got != "N" {
		t.Fatalf("last collapsed glyph=%q", got)
	}
}
