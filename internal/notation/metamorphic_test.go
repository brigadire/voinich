package notation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func baseFixture(t *testing.T) []Record {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(fixtureRoot(t), "C10_SYNTHETIC_TOKEN", "source.json"))
	if err != nil {
		t.Fatal(err)
	}
	var src SourceDocument
	if err := json.Unmarshal(b, &src); err != nil {
		t.Fatal(err)
	}
	r, err := NormalizeFixture(src)
	if err != nil {
		t.Fatal(err)
	}
	return r
}
func mustAnalyze(t *testing.T, r []Record) Fingerprint {
	t.Helper()
	fp, err := Analyze(r)
	if err != nil {
		t.Fatal(err)
	}
	return fp
}

func TestM1RenameInvariant(t *testing.T) {
	r := baseFixture(t)
	renamed, err := RenameSymbols(r, map[string]string{"a": "#", "b": "7", "c": "X", "d": "Q"})
	if err != nil {
		t.Fatal(err)
	}
	assertFamilyEqual(t, mustAnalyze(t, r), mustAnalyze(t, renamed), "GTSLD")
}
func TestM2DuplicationPreservesRelativeCore(t *testing.T) {
	a, b := mustAnalyze(t, baseFixture(t)), mustAnalyze(t, Duplicate(baseFixture(t)))
	am, bm := fingerprintMetricMap(a), fingerprintMetricMap(b)
	for _, id := range []string{"G02_INITIAL_RESTRICTION_DENSITY/", "G03_FINAL_RESTRICTION_DENSITY/", "G04_BIGRAM_OCCUPANCY/", "G05_TRIGRAM_OCCUPANCY/", "T03_UNIQUE_TOKEN_RATIO/", "T04_HAPAX_RATIO/"} {
		if id == "T03_UNIQUE_TOKEN_RATIO/" || id == "T04_HAPAX_RATIO/" {
			continue
		}
		if abs(am[id].Value-bm[id].Value) > 1e-12 {
			t.Errorf("%s changed", id)
		}
	}
}
func TestM3TokenShuffleTargetsSequence(t *testing.T) {
	a := mustAnalyze(t, baseFixture(t))
	b := mustAnalyze(t, ShuffleTokenOrder(baseFixture(t), 20260830))
	assertFamilyEqual(t, a, b, "GT")
	if abs(fingerprintMetricMap(a)["S04_REPEATED_BIGRAM_TYPES/"].Value-fingerprintMetricMap(b)["S04_REPEATED_BIGRAM_TYPES/"].Value) < 1e-12 {
		t.Fatal("sequence fixture did not react to token shuffle")
	}
}
func TestM4WithinTokenShuffleTargetsAdjacency(t *testing.T) {
	a := mustAnalyze(t, baseFixture(t))
	b := mustAnalyze(t, ShuffleWithinTokens(baseFixture(t), 20260830))
	if abs(fingerprintMetricMap(a)["T01_MEAN_TOKEN_LENGTH/"].Value-fingerprintMetricMap(b)["T01_MEAN_TOKEN_LENGTH/"].Value) > 1e-12 {
		t.Fatal("token lengths changed")
	}
	if abs(fingerprintMetricMap(a)["G04_BIGRAM_OCCUPANCY/"].Value-fingerprintMetricMap(b)["G04_BIGRAM_OCCUPANCY/"].Value) < 1e-12 {
		t.Fatal("glyph adjacency did not react")
	}
}
func TestM5M6LowerGrammarStable(t *testing.T) {
	a := mustAnalyze(t, baseFixture(t))
	line := mustAnalyze(t, ShuffleLines(baseFixture(t), 20260830))
	page := mustAnalyze(t, ShufflePages(baseFixture(t), 20260830))
	assertFamilyEqual(t, a, line, "GTSL")
	assertFamilyEqual(t, a, page, "GTSL")
	am, lm, pm := fingerprintMetricMap(a), fingerprintMetricMap(line), fingerprintMetricMap(page)
	if abs(am["D_LINE_PROGRESSION/"].Value-lm["D_LINE_PROGRESSION/"].Value) < 1e-12 {
		t.Fatal("line shuffle did not change document progression")
	}
	if abs(am["D_PAGE_PROGRESSION/"].Value-pm["D_PAGE_PROGRESSION/"].Value) < 1e-12 {
		t.Fatal("page shuffle did not change document progression")
	}
}
