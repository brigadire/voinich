package task82b

import "testing"

func TestOperatorRegistrySmoke(t *testing.T) {
	groups := [][]string{
		{"the", "quick", "brown", "fox"},
		{"jumps", "over", "the", "lazy", "dog"},
		{"the", "fox", "runs", "away", "fast"},
		{"a"},
		{},
	}
	tokenAtoms, glyphAtoms := BuildAtoms(groups)
	if len(tokenAtoms) != 15 {
		t.Fatalf("tokenAtoms = %d, want 15", len(tokenAtoms))
	}
	ops := Registry()
	if len(ops) != 20 {
		t.Fatalf("registry size = %d, want 20", len(ops))
	}
	seen := map[string]bool{}
	for _, op := range ops {
		if seen[op.ID] {
			t.Fatalf("duplicate operator id %s", op.ID)
		}
		seen[op.ID] = true
		sel := Apply(op, tokenAtoms, glyphAtoms, len(groups))
		out := Render(sel, tokenAtoms, glyphAtoms, len(groups))
		total := 0
		for _, g := range out {
			total += len(g)
		}
		if total != len(sel.Chosen) {
			t.Fatalf("%s: rendered %d tokens, selection chose %d", op.ID, total, len(sel.Chosen))
		}
		t.Logf("%-30s kind=%-6s chosen=%-4d pool=%-4d skipped=%d class=%s", op.ID, sel.Kind, len(sel.Chosen), sel.CandidatePool, sel.SkippedGroups, op.ExtractionClass)
	}
}

func TestFirstGlyphOfTokenValues(t *testing.T) {
	groups := [][]string{{"cat", "dog"}, {"emu"}}
	tokenAtoms, glyphAtoms := BuildAtoms(groups)
	op := Operator{StructuralClass: "FIRST_GLYPH_OF_TOKEN", NullClass: "PER_GROUP"}
	sel := Apply(op, tokenAtoms, glyphAtoms, len(groups))
	out := Render(sel, tokenAtoms, glyphAtoms, len(groups))
	if got := out[0]; len(got) != 2 || got[0] != "c" || got[1] != "d" {
		t.Fatalf("line0 = %v, want [c d]", got)
	}
	if got := out[1]; len(got) != 1 || got[0] != "e" {
		t.Fatalf("line1 = %v, want [e]", got)
	}
}

func TestPeriodicTokenPhaseZero(t *testing.T) {
	groups := [][]string{{"a", "b", "c", "d", "e", "f", "g"}}
	tokenAtoms, glyphAtoms := BuildAtoms(groups)
	op := Operator{StructuralClass: "PERIODIC_TOKEN", Param: 3, NullClass: "PERIODIC"}
	sel := Apply(op, tokenAtoms, glyphAtoms, len(groups))
	out := Render(sel, tokenAtoms, glyphAtoms, len(groups))
	want := []string{"a", "d", "g"}
	if len(out[0]) != len(want) {
		t.Fatalf("got %v, want %v", out[0], want)
	}
	for i, w := range want {
		if out[0][i] != w {
			t.Fatalf("got %v, want %v", out[0], want)
		}
	}
}
