package globalregime

import (
	"fmt"
	"math"
	"testing"
)

// TestJSDistanceDeterministicAcrossCalls is a regression test for a
// same-seed nondeterminism bug found while validating conditionalregime's
// residual-distance-matrix hoist (task27): jsDistance's d was a single
// running sum fed by every key of the union of a and b, summed via two
// separate `range` loops over maps, so the result depended on Go's
// randomized map iteration order, not just the input. jsDistance is
// exported (as JSDistance) and used by internal/conditionalregime and
// internal/residualdiagnostic in addition to this package's own CLI, so
// this affected every JS-distance-based computation across three
// packages. Confirmed against the pre-fix implementation: 14-22 distinct
// float64 bit patterns across 500 calls on identical input.
func TestJSDistanceDeterministicAcrossCalls(t *testing.T) {
	a, b := profile{}, profile{}
	for i := 0; i < 300; i++ {
		tok := fmt.Sprintf("t%03d", i)
		a[tok] = 1.0 / float64(1+i)
	}
	for k, v := range a {
		b[k] = v * 0.7
	}
	seen := map[uint64]bool{}
	for i := 0; i < 500; i++ {
		seen[math.Float64bits(jsDistance(a, b))] = true
	}
	if len(seen) != 1 {
		t.Fatalf("jsDistance produced %d distinct float64 bit patterns across 500 calls on identical input", len(seen))
	}
}

// TestOverlapDeterministicAcrossCalls is a regression test for the same
// class of bug in overlap (used for Window.WeightedOverlap): s was a
// single running sum fed by every key of a.
func TestOverlapDeterministicAcrossCalls(t *testing.T) {
	a, b := profile{}, profile{}
	for i := 0; i < 300; i++ {
		tok := fmt.Sprintf("t%03d", i)
		a[tok] = 1.0 / float64(1+i)
		b[tok] = 1.0 / float64(2+i)
	}
	seen := map[uint64]bool{}
	for i := 0; i < 500; i++ {
		seen[math.Float64bits(overlap(a, b))] = true
	}
	if len(seen) != 1 {
		t.Fatalf("overlap produced %d distinct float64 bit patterns across 500 calls on identical input", len(seen))
	}
}

// TestCosineDeterministicAcrossCalls is a regression test for the same
// class of bug in cosine (used for Window.Cosine): dot/aa were single
// running sums fed by every key of a, bb by every key of b.
func TestCosineDeterministicAcrossCalls(t *testing.T) {
	a, b := profile{}, profile{}
	for i := 0; i < 300; i++ {
		tok := fmt.Sprintf("t%03d", i)
		a[tok] = 1.0 / float64(1+i)
		b[tok] = 1.0 / float64(2+i)
	}
	seen := map[uint64]bool{}
	for i := 0; i < 500; i++ {
		seen[math.Float64bits(cosine(a, b))] = true
	}
	if len(seen) != 1 {
		t.Fatalf("cosine produced %d distinct float64 bit patterns across 500 calls on identical input", len(seen))
	}
}
