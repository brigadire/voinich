package notation

import (
	"bytes"
	"math"
	"testing"
)

func distributionFixture() []Record {
	return syntheticLinedCorpus(60, 8)
}

func TestD1ProbabilityNormalization(t *testing.T) {
	r := distributionFixture()
	for _, d := range [][]DistributionPoint{
		TokenLengthDistribution(r, "C10", "R1"),
		PositionalRestrictionDistribution(r, "C10", "R1"),
	} {
		var sum float64
		for _, p := range d {
			if !p.Comparable {
				continue
			}
			sum += p.Probability
		}
		if math.Abs(sum-1) > 1e-9 {
			t.Fatalf("distribution does not sum to 1: %v", sum)
		}
	}
}

func TestD2SerializationRoundTrip(t *testing.T) {
	r := baseFixture(t)
	want := BuildDistributions(r, "C10", "R1")
	var buf bytes.Buffer
	if err := WriteDistributionsTSV(&buf, want); err != nil {
		t.Fatal(err)
	}
	got, err := ReadDistributionsTSV(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("round-trip row count %d != %d", len(got), len(want))
	}
	for i := range want {
		if got[i].BinOrCategory != want[i].BinOrCategory || math.Abs(got[i].Probability-want[i].Probability) > 1e-9 {
			t.Fatalf("row %d mismatch: %+v vs %+v", i, got[i], want[i])
		}
	}
}

func TestD3LabelInvariance(t *testing.T) {
	r := baseFixture(t)
	renamed, err := RenameSymbols(r, map[string]string{"a": "#", "b": "7", "c": "X", "d": "Q"})
	if err != nil {
		t.Fatal(err)
	}
	a := TokenLengthDistribution(r, "C10", "R1")
	b := TokenLengthDistribution(renamed, "C10", "R1")
	if len(a) != len(b) {
		t.Fatalf("length distribution changed under rename: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if math.Abs(a[i].Probability-b[i].Probability) > 1e-12 {
			t.Fatalf("token-length distribution changed under rename at %d", i)
		}
	}
	pa := PositionalRestrictionDistribution(r, "C10", "R1")
	pb := PositionalRestrictionDistribution(renamed, "C10", "R1")
	for i := range pa {
		if math.Abs(pa[i].Probability-pb[i].Probability) > 1e-12 {
			t.Fatalf("positional restriction distribution changed under rename at %d", i)
		}
	}
}

func TestD4BootstrapDeterminism(t *testing.T) {
	rs := syntheticLinedCorpus(60, 8)
	a, err := RunBootstrap(rs, "SYN-RAREFY", "SYN-R1", 20, BaseSeed)
	if err != nil {
		t.Fatal(err)
	}
	b, err := RunBootstrap(rs, "SYN-RAREFY", "SYN-R1", 20, BaseSeed)
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != len(b) {
		t.Fatalf("row counts differ: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("bootstrap not deterministic at row %d: %+v vs %+v", i, a[i], b[i])
		}
	}
}

func TestD5DegenerateDistributionNoNaN(t *testing.T) {
	// A single repeated-symbol token "aaa" makes "a" appear initial,
	// internal, and final at once, so no symbol is ever restricted from any
	// position: total==0 in PositionalRestrictionDistribution must not
	// divide by zero into NaN.
	rs := []Record{{
		SchemaVersion: SchemaVersion, CorpusID: "DEG", Representation: "R1",
		Document: ObservedLevel{Value: "d", Observed: true}, PhysicalLine: ObservedLevel{Value: "1", Observed: true},
		TokenID: "t1", TokenIndex: 0, Token: "aaa", Symbols: []string{"a", "a", "a"},
	}}
	for _, p := range PositionalRestrictionDistribution(rs, "DEG", "R1") {
		if math.IsNaN(p.Probability) || math.IsInf(p.Probability, 0) {
			t.Fatalf("NaN/Inf emitted: %+v", p)
		}
		if p.Comparable {
			t.Fatalf("degenerate distribution should not be marked comparable: %+v", p)
		}
	}
	if s := EstimateScale(nil); s.Status != ScaleStatusDegenerate {
		t.Fatalf("empty scale sample must be DEGENERATE, got %s", s.Status)
	}
	if s := EstimateScale([]float64{5, 5, 5, 5}); s.Status != ScaleStatusDegenerate {
		t.Fatalf("zero-MAD zero-IQR sample must be DEGENERATE, got %s", s.Status)
	}
}

func TestD6CommonSupportEnforcement(t *testing.T) {
	support := []string{"INITIAL_RESTRICTED", "INTERNAL_RESTRICTED", "FINAL_RESTRICTED"}
	p := []DistributionPoint{{BinOrCategory: "INITIAL_RESTRICTED", Probability: 0.5}, {BinOrCategory: "INTERNAL_RESTRICTED", Probability: 0.5}}
	q := []DistributionPoint{{BinOrCategory: "OUTSIDE_SUPPORT", Probability: 1}}
	if _, err := CategoricalJensenShannon(support, p, q); err == nil {
		t.Fatal("expected rejection of a category outside the frozen common support")
	}
	q2 := []DistributionPoint{{BinOrCategory: "FINAL_RESTRICTED", Probability: 1}}
	js, err := CategoricalJensenShannon(support, p, q2)
	if err != nil {
		t.Fatal(err)
	}
	if js <= 0 {
		t.Fatal("expected a positive JS divergence for disjoint mass")
	}
}

func TestMetricOutputTypesCoverEveryEmittedMetric(t *testing.T) {
	types := map[string]bool{}
	for _, r := range MetricOutputTypes() {
		types[r.MetricID] = true
	}
	fp := mustAnalyze(t, baseFixture(t))
	for _, m := range fp.Metrics {
		if !types[m.MetricID] {
			t.Errorf("metric %s has no frozen output type", m.MetricID)
		}
	}
}
