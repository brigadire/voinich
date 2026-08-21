package voynichvalidation

import (
	"math"
	"reflect"
	"testing"

	"zcore.dev/voinich/internal/inversetransposition"
)

func TestOriginalEvaluatorDoesNotTransformInput(t *testing.T) {
	in := []string{"a", "b", "a", "c"}
	before := append([]string(nil), in...)
	_ = inversetransposition.Measure(in)
	if !reflect.DeepEqual(in, before) {
		t.Fatal("measure mutated original corpus")
	}
}
func TestFixedCandidateIsImmutableAndNoSearch(t *testing.T) {
	c := inversetransposition.Candidate{Width: 2, Order: "natural", Rounds: 1, Seed: 1}
	if c.Width != 2 || c.Order != "natural" {
		t.Fatal(c)
	}
	got, _ := c.Apply([]string{"a", "b", "c", "d"})
	if len(got) != 4 {
		t.Fatal(got)
	}
}
func TestRawEffectDoesNotUseCandidateLocalNormalisation(t *testing.T) {
	base := Metrics{TransitionConcentration: 1}
	got := Delta(Metrics{TransitionConcentration: 2}, base)
	if got.TransitionConcentration != 1 {
		t.Fatal(got)
	}
	if FixedCalibrationScore(got) == 0 || math.IsNaN(FixedCalibrationScore(got)) {
		t.Fatal("fixed scale missing")
	}
}
func TestDeterministicNull(t *testing.T) {
	a := NullDistribution([]string{"a", "b", "c", "a"}, 4, 9, Metrics{})
	b := NullDistribution([]string{"a", "b", "c", "a"}, 4, 9, Metrics{})
	if !reflect.DeepEqual(a, b) {
		t.Fatal("null distribution is not deterministic")
	}
}
func TestSplitPreservesLogicalBoundaries(t *testing.T) {
	a, b := SplitByLines([]string{"a", "b", "c", "d", "e"}, []int{1, 1, 1, 1, 1})
	if len(a) != 4 || len(b) != 1 {
		t.Fatalf("%v %v", a, b)
	}
}
