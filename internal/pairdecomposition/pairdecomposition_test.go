package pairdecomposition

import (
	"math"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDistributionMetricsUseFullSupport(t *testing.T) {
	a := map[string]int{"shared": 2, "a": 1}
	b := map[string]int{"shared": 2, "b": 1}
	p := contextProfile(a, b, .8, map[string]int{}, 1)
	if p.Jaccard != 1.0/3.0 {
		t.Fatalf("Jaccard = %v, want 1/3", p.Jaccard)
	}
	if len(p.Common) != 1 || p.Common[0].Token != "shared" {
		t.Fatalf("common = %+v", p.Common)
	}
	if len(p.DistributionA) != 2 || len(p.DistributionB) != 2 {
		t.Fatalf("full distributions were truncated: %d/%d", len(p.DistributionA), len(p.DistributionB))
	}
	// Display truncation must not alter the full-distribution metric.
	want := 2.0 / 3.0
	if math.Abs(p.JensenShannonSimilarity-want) > 1e-12 {
		t.Fatalf("JS similarity = %.15g, want %.15g", p.JensenShannonSimilarity, want)
	}
	if math.Abs(p.EffectiveVocabularyA-3/math.Pow(2, 2.0/3.0)) > 1e-12 {
		t.Fatalf("effective vocabulary = %.15g", p.EffectiveVocabularyA)
	}
}

func TestPositionSummaryAndSimilarity(t *testing.T) {
	p := profile{Count: 4, Start: 1, End: 2, Positions: map[int]int{0: 1, 1: 2, 3: 1}}
	s := positionSummary(p)
	if s.LineStartProbability != .25 || s.LineEndProbability != .5 || s.Mean != 1.25 || s.Median != 1 {
		t.Fatalf("summary = %+v", s)
	}
	if got := jsInt(p.Positions, p.Positions); math.Abs(got-1) > 1e-12 {
		t.Fatalf("self JS similarity = %v", got)
	}
}

func TestControlsFavorMedianAndExcludeTargetTokens(t *testing.T) {
	target := PairSource{TokenA: "a", TokenB: "b", CountA: 100, CountB: 120, Structural: .8, Reliability: .9, Graphemic: .8}
	all := []PairSource{
		target,
		{TokenA: "a", TokenB: "x", CountA: 100, CountB: 120, Structural: .25, Reliability: .9, Graphemic: .8},
		{TokenA: "c", TokenB: "d", CountA: 100, CountB: 120, Structural: .26, Reliability: .9, Graphemic: .8},
		{TokenA: "e", TokenB: "f", CountA: 100, CountB: 120, Structural: .6, Reliability: .9, Graphemic: .8},
	}
	got := chooseControls(target, all, .25, 1)
	if len(got) != 1 || got[0].Pair.TokenA != "c" || got[0].Pair.TokenB != "d" {
		t.Fatalf("control = %+v", got)
	}
}

func TestOutputYAMLHasSeparateFields(t *testing.T) {
	b, err := yaml.Marshal(Output{Pairs: []PairResult{{TokenA: "a", TokenB: "b", Left: ContextProfile{ObservationsA: 1, ObservationsB: 2}}}})
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, want := range []string{"token_a: a", "token_b: b", "observations_a: 1", "observations_b: 2"} {
		if !contains(s, want) {
			t.Fatalf("YAML lacks %q:\n%s", want, s)
		}
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
