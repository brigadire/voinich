package propertytrajectory

import (
	"fmt"
	"math"
	"testing"
)

// TestEntropyDeterministicAcrossCalls is a regression test for a same-seed
// nondeterminism bug found while validating item 9 (task27):
// entropy's h was a single running sum fed by ranging directly over its
// input map, so the result depended on Go's randomized map iteration
// order, not just the input. entropy feeds predecessor_entropy,
// successor_entropy, positional_entropy, positional_specialization, and
// (via math.Pow(2, entropy(...))) effective_predecessor_count/
// effective_successor_count - core token properties used throughout the
// analyzer, which is why this one bug caused ULP-level differences across
// nearly every output file when the same binary was run twice.
func TestEntropyDeterministicAcrossCalls(t *testing.T) {
	m := map[string]int{}
	for i := 0; i < 300; i++ {
		m[fmt.Sprintf("tok%03d", i)] = 1 + i%17
	}
	seen := map[uint64]bool{}
	for i := 0; i < 500; i++ {
		seen[math.Float64bits(entropy(m))] = true
	}
	if len(seen) != 1 {
		t.Fatalf("entropy produced %d distinct float64 bit patterns across 500 calls on identical input", len(seen))
	}
}
