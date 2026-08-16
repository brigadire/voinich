package structuralprojection

import (
	"fmt"
	"math"
	"testing"
)

// multiBinFixture builds a synthetic projection/counts pair spanning several
// distinct log2-frequency bins, so RandomizeProjection's per-bin shuffle
// actually exercises more than one bin. The existing
// TestRandomSpaceControlDeterministicAndPreservesRowMass fixture (3 tokens,
// counts 10/11/12) is a coverage gap: all three tokens land in the same
// log2 bin, so it can never observe bin-processing-order nondeterminism.
func multiBinFixture() (Projection, map[string]int, []string) {
	counts := map[string]int{}
	p := Projection{}
	tokens := make([]string, 0, 60)
	for i := 0; i < 60; i++ {
		tok := fmt.Sprintf("t%03d", i)
		tokens = append(tokens, tok)
		counts[tok] = 1 + (i*37)%500 // spans many distinct log2 bins
		row := map[string]float64{tok: 1}
		for j := 0; j < 5; j++ {
			row[fmt.Sprintf("t%03d", (i+j+1)%60)] = 0.1 * float64(j+1)
		}
		p[tok] = row
	}
	return p, counts, tokens
}

// TestRandomizeProjectionDeterministicAcrossManyBins is a regression test
// for a same-seed nondeterminism bug: RandomizeProjection shuffled each
// log2-frequency bin's destinations by ranging directly over
// `map[int][]string` bins with a single shared *rand.Rand. Which bin
// consumed which slice of the random stream therefore depended on Go's
// randomized map iteration order, not just the seed, so two calls with an
// identical seed (and more than one populated bin) could legitimately
// return different projections. Confirmed against the pre-fix
// implementation: 10 distinct results for token "t000" across 200 calls
// with the same seed. The fix visits bins in sorted key order instead.
func TestRandomizeProjectionDeterministicAcrossManyBins(t *testing.T) {
	p, counts, tokens := multiBinFixture()
	fb := buildFrequencyBins(tokens, counts)
	seen := map[string]bool{}
	for i := 0; i < 500; i++ {
		out := RandomizeProjection(p, fb, 12345)
		seen[fmt.Sprintf("%v", out["t000"])] = true
	}
	if len(seen) != 1 {
		t.Fatalf("RandomizeProjection produced %d distinct results for t000 across 500 calls with the same seed", len(seen))
	}
}

// TestNormalizeDeterministicAcrossCalls is a regression test for a
// same-seed nondeterminism bug: normalize summed a row's positive weights
// with `for _, v := range m`, a single running total fed by every key in
// one call, so the result depended on Go's randomized map iteration order,
// not just the input. Confirmed against the pre-fix implementation: 8-9
// distinct float64 bit patterns across 500 calls on identical input.
func TestNormalizeDeterministicAcrossCalls(t *testing.T) {
	m := map[string]float64{}
	for i := 0; i < 200; i++ {
		m[fmt.Sprintf("t%03d", i)] = 1.0 / float64(1+i)
	}
	seen := map[uint64]bool{}
	for i := 0; i < 500; i++ {
		seen[math.Float64bits(normalize(m)["t000"])] = true
	}
	if len(seen) != 1 {
		t.Fatalf("normalize produced %d distinct float64 bit patterns across 500 calls on identical input", len(seen))
	}
}

// TestProjectDistributionDeterministicAcrossCalls is a regression test for
// a same-seed nondeterminism bug: ProjectDistribution accumulated
// out[y] += mass*w for every observed token x whose projection row routes
// weight to y. Several distinct x's can target the same y within one call,
// so the order distinct x's were visited in (`range counts`) made the sum
// nondeterministic. Confirmed against the pre-fix implementation: 6-7
// distinct float64 bit patterns for out["y"] across 500 calls on identical
// input (100 observed tokens all routing partial weight to "y").
func TestProjectDistributionDeterministicAcrossCalls(t *testing.T) {
	counts := map[string]int{}
	p := Projection{}
	for i := 0; i < 100; i++ {
		obs := fmt.Sprintf("o%03d", i)
		counts[obs] = 1 + i%7
		p[obs] = map[string]float64{"y": 0.3, obs: 0.7}
	}
	seen := map[uint64]bool{}
	for i := 0; i < 500; i++ {
		seen[math.Float64bits(ProjectDistribution(counts, p)["y"])] = true
	}
	if len(seen) != 1 {
		t.Fatalf("ProjectDistribution produced %d distinct float64 bit patterns for out[\"y\"] across 500 calls on identical input", len(seen))
	}
}

// TestMetricsFloatDeterministicAcrossCalls is a regression test for a
// same-seed nondeterminism bug: metricsFloat summed js's divergence term
// and overlap over the union of a and b's keys with `for k := range keys`
// (a map[string]bool), a single running total per call fed by every key, so
// both results depended on Go's randomized map iteration order. Confirmed
// against the pre-fix implementation: up to 14 distinct float64 bit
// patterns across 500 calls on identical input. jaccard is an exact integer
// ratio and was already deterministic.
func TestMetricsFloatDeterministicAcrossCalls(t *testing.T) {
	a, b := map[string]float64{}, map[string]float64{}
	for i := 0; i < 200; i++ {
		tok := fmt.Sprintf("t%03d", i)
		a[tok] = 1.0 / float64(1+i)
		b[tok] = 1.0 / float64(2+i)
	}
	seenJS, seenOverlap := map[uint64]bool{}, map[uint64]bool{}
	for i := 0; i < 500; i++ {
		js, overlap, _ := metricsFloat(a, b)
		seenJS[math.Float64bits(js)] = true
		seenOverlap[math.Float64bits(overlap)] = true
	}
	if len(seenJS) != 1 || len(seenOverlap) != 1 {
		t.Fatalf("metricsFloat produced %d distinct js and %d distinct overlap bit patterns across 500 calls on identical input", len(seenJS), len(seenOverlap))
	}
}

// TestTransitionsTieOrderDeterministicAcrossCalls is a regression test for
// a same-seed nondeterminism bug found while validating the GenericSmoothing
// buffer-reuse optimization (task27): transitions() built its output slice
// by ranging over a map[string]float64 (joint), then sorted with a
// comparator that had no tie-breaker beyond (Lift, Observed). When two
// transitions have exactly equal Lift and Observed (a genuine tie — this
// happens in the real corpus for tokens "d"/"o"), the unstable sort's
// result depended on that map's randomized initial iteration order, so
// which of the tied transitions appeared first (and thus which survived a
// boundary at -top/limit) varied run to run even with byte-identical
// input. This is unrelated to GenericSmoothing itself — transitions() only
// depends on the corpus and the (fixed, pre-trial-loop) structural
// projection — but was only surfaced by the required old-vs-new output
// comparison. Fixed by adding a deterministic Source/Destination
// lexicographic tie-breaker to the sort comparator.
func TestTransitionsTieOrderDeterministicAcrossCalls(t *testing.T) {
	var tokens []string
	for i := 0; i < 5; i++ {
		tokens = append(tokens, "x", "y", "y", "x")
	}
	c := corpus{Tokens: tokens}
	out := transitions(c, Projection{}, 50)
	var xy, yx *Transition
	for i := range out {
		if out[i].Source == "x" && out[i].Destination == "y" {
			xy = &out[i]
		}
		if out[i].Source == "y" && out[i].Destination == "x" {
			yx = &out[i]
		}
	}
	if xy == nil || yx == nil {
		t.Fatalf("fixture did not produce both x->y and y->x: %+v", out)
	}
	if xy.Lift != yx.Lift || xy.Observed != yx.Observed {
		t.Fatalf("fixture is not a true tie: xy=%+v yx=%+v", xy, yx)
	}
	seen := map[string]bool{}
	for i := 0; i < 500; i++ {
		out := transitions(c, Projection{}, 50)
		key := fmt.Sprintf("%s->%s,%s->%s", out[0].Source, out[0].Destination, out[1].Source, out[1].Destination)
		seen[key] = true
	}
	if len(seen) != 1 {
		t.Fatalf("transitions() tie order produced %d distinct orderings across 500 calls on identical input: %v", len(seen), seen)
	}
}
