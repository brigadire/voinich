package main

import (
	"testing"
	"zcore.dev/voinich/internal/mechanismspace"
)

func TestFrozenCandidatesAreStable(t *testing.T) {
	cs := candidates()
	if len(cs) != 7 {
		t.Fatalf("candidate count = %d, want 7", len(cs))
	}
	if cs[0].Name != "M0_IDENTITY" || cs[1].Name != "M1_MONOALPHABETIC" {
		t.Fatalf("controls must be first: %#v", cs[:2])
	}
	for _, c := range cs {
		if c.Config.Hash() == "" || c.Class == "" {
			t.Fatalf("incomplete frozen candidate: %#v", c)
		}
	}
}

func TestCleanControlsAndZeroDamage(t *testing.T) {
	plain := mechanismspace.Corpus{Name: "test", Words: []string{"abc", "cab"}, Lines: []int{0, 0}}
	for _, c := range candidates()[:2] {
		o := mechanismspace.Transform(c.Config, plain)
		if len(o.Tokens) != len(plain.Words) || o.InputUnits != 6 {
			t.Fatalf("%s accounting mismatch: %#v", c.Name, o)
		}
		if got := corruptionRecovery(c, "GLYPH_SUBSTITUTION", 0); got != 1 {
			t.Fatalf("%s zero-damage recovery = %v, want 1", c.Name, got)
		}
	}
}

func TestPreimageEstimatorMonotonicity(t *testing.T) {
	m := metric{ri: .25}
	if got, want := m.preimages(16), m.preimages(8); got <= want {
		t.Fatalf("preimage estimate must grow with block length: %v <= %v", got, want)
	}
}
