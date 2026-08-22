package main

import (
	"os"
	"path/filepath"
	"reflect"
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

func TestTask66CompatibilityUsesNamedColumns(t *testing.T) {
	path := filepath.Join(t.TempDir(), "family.tsv")
	data := "mechanism\tcorpus\tevaluation_set\tfamily\tprogress\toverall_status\nM1\tDoyle\tDEVELOPMENT\tTOKEN_ORDER\t0.25\tOK\n"
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := task66Compatibility(path)
	if err != nil {
		t.Fatal(err)
	}
	if v := got["M1"]["Doyle"]["TOKEN_ORDER"]; v != .25 {
		t.Fatalf("header-driven progress = %v, want .25", v)
	}
}

func TestCleanControlsAndZeroDamage(t *testing.T) {
	plain := mechanismspace.Corpus{Name: "test", Words: []string{"abc", "cab"}, Lines: []int{0, 0}}
	for _, c := range candidates()[:2] {
		d := trainDecoder(c, plain, plain)
		e := encodeAligned(c, plain)
		got := d.decode(e.cipher)
		if r := measureRecovery(e.plain, got); r.glyph != 1 || r.token != 1 || r.exact != 1 {
			t.Fatalf("%s clean recovery = %#v, want exact", c.Name, r)
		}
		short := d.decode(e.cipher[:1])
		if r := measureRecovery(e.plain[:1], short); r.exact != 1 {
			t.Fatalf("%s short-block recovery lost the full-message key: %#v", c.Name, r)
		}
	}
}

func TestCorruptionDeterministicAndSingle(t *testing.T) {
	in := [][]string{{"a", "b"}, {"c", "d"}, {"e", "f"}}
	a, apos, _ := corruptOne(in, "GLYPH_SUBSTITUTION", 1, 67)
	b, bpos, _ := corruptOne(in, "GLYPH_SUBSTITUTION", 1, 67)
	if apos != bpos || !reflect.DeepEqual(a, b) {
		t.Fatal("single-error corruption is not deterministic")
	}
	if n := changedGlyphs(in, a); n != 1 {
		t.Fatalf("changed glyphs = %d, want 1", n)
	}
}

func TestConflationManyToOneAndSplittingCollapses(t *testing.T) {
	in := [][]string{{"a", "b", "c", "d"}, {"a", "b", "c", "d"}}
	conflated, pairs := applyConflation(in, .5, 67)
	if pairs < 1 || reflect.DeepEqual(in, conflated) {
		t.Fatal("conflation did not create a many-to-one representation")
	}
	split, classes := applySplitting(in, .5, 67)
	if classes < 1 || !reflect.DeepEqual(in, removeSplittingMarks(split)) {
		t.Fatal("splitting collapse oracle does not recover physical glyph classes")
	}
}

func TestPropagationAndResetLocalizeBoundaryDeletion(t *testing.T) {
	c := candidates()[0]
	plain := mechanismspace.Corpus{Name: "test", Words: []string{"aa", "bb", "cc", "dd", "ee", "ff"}, Lines: []int{0, 0, 0, 1, 1, 1}}
	d := trainDecoder(c, plain, plain)
	e := encodeAligned(c, plain)
	damaged, pos, _ := corruptOne(e.cipher, "TOKEN_BOUNDARY_DELETION", 2, 67)
	noReset := propagationMetrics(e.plain, d.decode(damaged), pos)
	withReset := propagationMetrics(e.plain, resetDecode(d, e.cipher, damaged, pos, 3), pos)
	if withReset.damaged >= noReset.damaged {
		t.Fatalf("reset did not localize damage: reset=%#v no-reset=%#v", withReset, noReset)
	}
}
