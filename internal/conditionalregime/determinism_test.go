package conditionalregime

import (
	"fmt"
	"math"
	"testing"

	"zcore.dev/voinich/internal/globalregime"
)

// TestEuclideanDistanceDeterministicAcrossCalls is a regression test for a
// same-seed nondeterminism bug found while auditing conditionalregime for
// the residualFitPrep hoist (task27): euclideanDistance's sum was a single
// running total fed by ranging over two maps (a, then the not-yet-seen
// part of b), so the result depended on Go's randomized map iteration
// order, not just the input. euclideanDistance is the core distance
// function backing every residual clustering/silhouette computation in
// this package. Confirmed against the pre-fix implementation: 2-3 distinct
// float64 bit patterns across 500 calls on identical input.
func TestEuclideanDistanceDeterministicAcrossCalls(t *testing.T) {
	a, b := vector{}, vector{}
	for i := 0; i < 300; i++ {
		tok := fmt.Sprintf("t%03d", i)
		a[tok] = 1.0 / float64(1+i)
		if i%3 != 0 {
			b[tok] = 1.0 / float64(2+i)
		} else {
			b[fmt.Sprintf("u%03d", i)] = 2.0 / float64(3+i)
		}
	}
	sa, sb := sortVector(a), sortVector(b)
	seen := map[uint64]bool{}
	for i := 0; i < 500; i++ {
		seen[math.Float64bits(euclideanDistance(sa, sb))] = true
	}
	if len(seen) != 1 {
		t.Fatalf("euclideanDistance produced %d distinct float64 bit patterns across 500 calls on identical input", len(seen))
	}
}

// TestBoundarySignatureTieBreakDeterministicAcrossCalls is a regression test
// for a same-seed nondeterminism bug found while validating this session's
// conditionalregime work (task27, unrelated to the residual hoist itself):
// boundarySignature picked the token with the largest |delta| by ranging
// directly over the before/after profile maps, so when two or more tokens
// tied for the largest |delta| - as constructed here, "a" going from
// frequency 1 to 0 and "b" going from 0 to 1 across the boundary, both
// |delta|=1 - which one won the tie depended on Go's randomized map
// iteration order, not just the input.
func TestBoundarySignatureTieBreakDeterministicAcrossCalls(t *testing.T) {
	tokens := make([]string, 0, 20)
	for i := 0; i < 10; i++ {
		tokens = append(tokens, "a")
	}
	for i := 0; i < 10; i++ {
		tokens = append(tokens, "b")
	}
	windows := globalregime.BuildWindows(tokens, 10, 10)
	if len(windows) != 2 {
		t.Fatalf("expected 2 non-overlapping windows, got %d", len(windows))
	}
	type result struct {
		tok, dir string
		mag      float64
	}
	seen := map[result]bool{}
	for i := 0; i < 500; i++ {
		tok, dir, mag := boundarySignature(windows, windows[0].Center)
		seen[result{tok, dir, mag}] = true
	}
	if len(seen) != 1 {
		t.Fatalf("boundarySignature produced %d distinct (token,direction,magnitude) results across 500 calls on identical input: %v", len(seen), seen)
	}
	var got result
	for r := range seen {
		got = r
	}
	if got.tok != "a" || got.dir != "decrease" || got.mag != 1 {
		t.Fatalf("expected deterministic tie-break to pick the lexicographically smallest tied token (\"a\", decrease, 1), got %+v", got)
	}
}

// TestEntropyOfPairsDeterministicAcrossCalls is a regression test for the
// same class of bug in entropyOfPairs: h was a single running total fed by
// ranging over a map[ClassID]int.
func TestEntropyOfPairsDeterministicAcrossCalls(t *testing.T) {
	var pairs []ClassID
	for i := 0; i < 200; i++ {
		pairs = append(pairs, ClassID{Scheme: SchemeJoint, Currier: fmt.Sprintf("C%d", i%11), Hand: fmt.Sprintf("H%d", i%7)})
	}
	seen := map[uint64]bool{}
	for i := 0; i < 500; i++ {
		seen[math.Float64bits(entropyOfPairs(pairs))] = true
	}
	if len(seen) != 1 {
		t.Fatalf("entropyOfPairs produced %d distinct float64 bit patterns across 500 calls on identical input", len(seen))
	}
}
