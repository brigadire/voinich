package normalizationcompare

import (
	"math"
	"testing"

	"zcore.dev/voinich/internal/normalization"
)

func TestSummarizeAndEmpiricalP(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5}
	distribution := Summarize(values)
	if distribution.Mean != 3 || distribution.Min != 1 || distribution.Max != 5 || distribution.Percentile50 != 3 {
		t.Fatalf("unexpected distribution: %+v", distribution)
	}
	if math.Abs(distribution.Stddev-math.Sqrt(2)) > 1e-12 {
		t.Fatalf("stddev = %v", distribution.Stddev)
	}
	upper := MakeEffect(1, 4, values, true)
	if math.Abs(upper.EmpiricalP-3.0/6.0) > 1e-12 {
		t.Fatalf("upper empirical p = %v", upper.EmpiricalP)
	}
	lower := MakeEffect(5, 2, values, false)
	if math.Abs(lower.EmpiricalP-3.0/6.0) > 1e-12 {
		t.Fatalf("lower empirical p = %v", lower.EmpiricalP)
	}
}

func TestCompareRawAndNormalizedMetrics(t *testing.T) {
	raw := Metrics{
		CrossLine: map[int]float64{2: 10, 3: 2}, MaxLength: 3,
		Contexts: map[int]ContextOrder{1: {ContextLength: 1, ConditionalEntropy: 4, RepeatedContextConditionalEntropy: 5, RepeatedContextCoverage: .5}},
	}
	structural := Metrics{
		CrossLine: map[int]float64{2: 20, 3: 4}, MaxLength: 4,
		Contexts: map[int]ContextOrder{1: {ContextLength: 1, ConditionalEntropy: 3, RepeatedContextConditionalEntropy: 4, RepeatedContextCoverage: .7}},
	}
	random := []Metrics{
		{CrossLine: map[int]float64{2: 12, 3: 3}, MaxLength: 3, Contexts: map[int]ContextOrder{1: {ConditionalEntropy: 3.5, RepeatedContextConditionalEntropy: 4.5, RepeatedContextCoverage: .6}}},
		{CrossLine: map[int]float64{2: 18, 3: 4}, MaxLength: 4, Contexts: map[int]ContextOrder{1: {ConditionalEntropy: 3.2, RepeatedContextConditionalEntropy: 4.2, RepeatedContextCoverage: .65}}},
	}
	result := CompareModel(normalization.Model{Threshold: .8, Label: "080"}, raw, structural, random)
	if len(result.NGrams) != 2 || result.NGrams[0].CrossLineRepeated.AbsoluteDelta != 10 || result.NGrams[0].CrossLineRepeated.Ratio != 2 {
		t.Fatalf("unexpected n-gram comparison: %+v", result.NGrams)
	}
	if result.MaxCrossLineSequenceLength.StructuralValue != 4 {
		t.Fatalf("unexpected max length effect: %+v", result.MaxCrossLineSequenceLength)
	}
	if len(result.ContextOrder) != 1 || result.ContextOrder[0].ConditionalEntropy.AbsoluteDelta != -1 || math.Abs(result.ContextOrder[0].RepeatedContextCoverage.AbsoluteDelta-.2) > 1e-12 {
		t.Fatalf("unexpected context comparison: %+v", result.ContextOrder)
	}
}
