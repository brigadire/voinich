package distancecontext

import (
	"math"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestBuildProfilesContinuousCrossesLines(t *testing.T) {
	c := corpus{Lines: [][]string{{"a", "b"}, {"c", "a"}}, Tokens: []string{"a", "b", "c", "a"}}
	continuous := buildProfiles(c, 2, false)
	bounded := buildProfiles(c, 2, true)
	if got := continuous["b"].Right[0]["c"]; got != 1 {
		t.Fatalf("continuous b→c across line = %d, want 1", got)
	}
	if got := bounded["b"].Right[0]["c"]; got != 0 {
		t.Fatalf("line-bounded b→c = %d, want 0", got)
	}
	if got := continuous["a"].Right[1]["c"]; got != 1 {
		t.Fatalf("exact +2 a→c = %d, want 1", got)
	}
}

func TestDistributionMetrics(t *testing.T) {
	js, overlap, jaccard := rawMetrics(map[string]int{"x": 2, "y": 2}, map[string]int{"x": 1, "y": 1})
	if math.Abs(js-1) > 1e-12 || math.Abs(overlap-1) > 1e-12 || math.Abs(jaccard-1) > 1e-12 {
		t.Fatalf("identical normalized distributions: JS=%g overlap=%g jaccard=%g", js, overlap, jaccard)
	}
	js, overlap, jaccard = rawMetrics(map[string]int{"x": 1}, map[string]int{"y": 1})
	if math.Abs(js) > 1e-12 || overlap != 0 || jaccard != 0 {
		t.Fatalf("disjoint distributions: JS=%g overlap=%g jaccard=%g", js, overlap, jaccard)
	}
}

func TestOutputYAMLHasDistinctSummaryAndCountFields(t *testing.T) {
	v := Output{Pairs: []PairResult{{TokenA: "a", TokenB: "b", CountA: 2, CountB: 3, RightSummary: Summary{Mean1To5: .1, Mean6To10: .2, Mean11To20: .3}}}}
	if _, err := yaml.Marshal(v); err != nil {
		t.Fatal(err)
	}
}
