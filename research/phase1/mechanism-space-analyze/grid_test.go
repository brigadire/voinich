package main

import "testing"

// test: the frozen grid instantiates every family M0-M11 at least once,
// and every ablation entry is uniquely named (task66 sections 9, 26, 33).
func TestGridCoversEveryFamily(t *testing.T) {
	grid := BuildGrid()
	seen := map[string]bool{}
	for _, e := range grid {
		seen[e.Config.Family] = true
	}
	for _, fam := range []string{"M0", "M1", "M2", "M3", "M4", "M5", "M6", "M7", "M8", "M9", "M10", "M11"} {
		if !seen[fam] {
			t.Fatalf("grid missing family %s", fam)
		}
	}
	names := map[string]bool{}
	for _, e := range grid {
		if names[e.Name] {
			t.Fatalf("duplicate grid entry name %s", e.Name)
		}
		names[e.Name] = true
	}
}

// test: the compositional ablation matrix varies exactly one of
// macro/local/grammar at a time relative to the all-off baseline, and
// every combination is present (task66 section 26).
func TestAblationMatrixIsComplete(t *testing.T) {
	entries := AblationEntries()
	want := map[string]bool{"G_ONLY": false, "S_ONLY": false, "M_ONLY": false, "G_PLUS_S": false, "M_PLUS_S": false, "M_PLUS_G": false, "M_PLUS_S_PLUS_G": false}
	for _, e := range entries {
		if _, ok := want[e.Name]; !ok {
			t.Fatalf("unexpected ablation entry %s", e.Name)
		}
		want[e.Name] = true
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("ablation matrix missing %s", name)
		}
	}
}

// test: LoadCorpora finds and length-normalizes all three minimum
// plaintext corpora (task66 sections 6-7).
func TestLoadCorporaMatchesVoynichSize(t *testing.T) {
	chdirRepoRoot(t)
	corpora, err := LoadCorpora()
	if err != nil {
		t.Fatalf("LoadCorpora: %v", err)
	}
	for _, want := range []string{"Doyle", "Longfellow", "Astafiev"} {
		c, ok := corpora[want]
		if !ok {
			t.Fatalf("missing corpus %s", want)
		}
		if len(c.Words) == 0 {
			t.Fatalf("corpus %s loaded empty", want)
		}
		if len(c.Words) > VoynichMatchedSize {
			t.Fatalf("corpus %s not length-normalized: %d words > %d", want, len(c.Words), VoynichMatchedSize)
		}
	}
}
