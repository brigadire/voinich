package tokenrelationvalidation

import (
	"fmt"
	"math"
	"testing"
)

// TestJSOverlapDeterministicAcrossCalls is a regression test for a
// same-seed nondeterminism bug found while profiling tokenrelationvalidation
// for the profilePermutations hot path (task27): jsOverlap's div/o were
// single running sums fed by every key of the union of a and b, built via a
// map[string]bool and ranged over directly, so the result depended on Go's
// randomized map iteration order, not just the input.
func TestJSOverlapDeterministicAcrossCalls(t *testing.T) {
	a, b := map[string]int{}, map[string]int{}
	for i := 0; i < 200; i++ {
		tok := fmt.Sprintf("t%03d", i)
		a[tok] = 1 + i
		if i%3 != 0 {
			b[tok] = 1 + 2*i
		} else {
			b[fmt.Sprintf("u%03d", i)] = 2 + i
		}
	}
	seenJS, seenO := map[uint64]bool{}, map[uint64]bool{}
	for i := 0; i < 500; i++ {
		js, o := jsOverlap(a, b)
		seenJS[math.Float64bits(js)] = true
		seenO[math.Float64bits(o)] = true
	}
	if len(seenJS) != 1 {
		t.Fatalf("jsOverlap's JS divergence produced %d distinct float64 bit patterns across 500 calls on identical input", len(seenJS))
	}
	if len(seenO) != 1 {
		t.Fatalf("jsOverlap's overlap produced %d distinct float64 bit patterns across 500 calls on identical input", len(seenO))
	}
}
