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
// entropy()/mi() used to sum over an unsorted Go map, so with enough
// distinct token types float64 accumulation order (and thus the result's
// low-order bits) varied from run to run despite an identical seed and
// corpus - Task58's own determinism requirement (section 34 test 1). A
// 2-type toy corpus can't detect this (two-term float addition is
// commutative regardless of order), so this uses a wider vocabulary.
func TestEntropyDeterministicAcrossManyRuns(t *testing.T) {
	var line []string
	for i := 0; i < 40; i++ {
		line = append(line, string(rune('a'+i%23)))
	}
	c := toy(line)
	first, _ := analyse(c, 100, 5, 7)
	for i := 0; i < 20; i++ {
		got, _ := analyse(c, 100, 5, 7)
		if got != first {
			t.Fatalf("run %d: analysis not deterministic across repeated calls:\n%+v\n%+v", i, first, got)
		}
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
