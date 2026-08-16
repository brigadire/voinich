package replicatedlocalaudit

import (
	"fmt"
	"math"
	"testing"
)

// buildSkewedFrequencyMaps returns two frequency maps with enough distinct
// keys and skewed magnitudes that summing their per-key JS-divergence terms
// in a different order changes the accumulated float64's bit pattern purely
// from floating-point rounding. Go's map iteration order is randomized
// independently for every `range` statement execution (not just once per
// process), so repeated calls to a function that sums over `range m` on the
// SAME map value can legitimately return different results run to run.
func buildSkewedFrequencyMaps() (map[string]int, map[string]int) {
	a, b := map[string]int{}, map[string]int{}
	for i := 0; i < 60; i++ {
		k := fmt.Sprintf("tok%03d", i)
		a[k] = 1 + (i*i)%37
		b[k] = 1 + (i*7+3)%41
	}
	return a, b
}

// TestJSSimilarityDeterministicAcrossCalls is a regression test for the
// same-seed nondeterminism discovered in the distance-profile stage: with
// the pre-fix implementation, jsSimilarity summed its per-key divergence
// contributions in the iteration order of a map[string]bool union-key set,
// so calling it repeatedly on byte-identical input could yield different
// float64 results depending on that random iteration order.
func TestJSSimilarityDeterministicAcrossCalls(t *testing.T) {
	a, b := buildSkewedFrequencyMaps()
	bits := map[uint64]bool{}
	for i := 0; i < 2000; i++ {
		v := jsSimilarity(a, b)
		bits[math.Float64bits(v)] = true
	}
	if len(bits) != 1 {
		t.Fatalf("jsSimilarity produced %d distinct float64 bit patterns across 2000 calls on identical input (map-iteration-order-dependent summation)", len(bits))
	}
}

// buildManyBlocksForEntropy builds a fixed (order-deterministic, since it's
// a slice) set of blocks with varying per-block occurrence counts of a
// single-token sequence, so sequenceObserved's block-count-keyed entropy
// accumulation has many distinct, differently-scaled terms to sum.
func buildManyBlocksForEntropy() []block {
	var blocks []block
	for i := 0; i < 50; i++ {
		joint := fmt.Sprintf("J%d/H%d", i%3, i%2)
		id := fmt.Sprintf("%s#%d", joint, i)
		n := 1 + (i*13)%29
		var toks []token
		for j := 0; j < n; j++ {
			toks = append(toks, token{Text: "seqtok", Line: fmt.Sprintf("%d", j), Joint: joint})
		}
		blocks = append(blocks, block{ID: id, Joint: joint, Tokens: toks})
	}
	return blocks
}

// TestSequenceObservedEntropyDeterministicAcrossCalls is a regression test
// for the same-seed nondeterminism observed in universal_sequence_inventory.tsv:
// with the pre-fix implementation, the per-block occurrence-count entropy
// term was summed in the iteration order of a map[string]int keyed by block
// ID, so repeated calls with identical blocks/candidate input could yield
// different float64 Entropy values.
func TestSequenceObservedEntropyDeterministicAcrossCalls(t *testing.T) {
	blocks := buildManyBlocksForEntropy()
	var tokens []token
	for _, b := range blocks {
		tokens = append(tokens, b.Tokens...)
	}
	cand := sequenceCandidate{ID: "seq", Sequence: "seqtok", Tokens: []string{"seqtok"}}
	bits := map[uint64]bool{}
	for i := 0; i < 2000; i++ {
		o := sequenceObserved(cand, tokens, blocks)
		bits[math.Float64bits(o.Entropy)] = true
	}
	if len(bits) != 1 {
		t.Fatalf("sequenceObserved.Entropy produced %d distinct float64 bit patterns across 2000 calls on identical input", len(bits))
	}
}
