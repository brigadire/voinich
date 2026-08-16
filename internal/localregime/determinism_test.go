package localregime

import (
	"fmt"
	"math"
	"testing"
)

// fixtureOverlappingProfiles builds two profiles with partial key overlap
// (some keys only in a, some only in b, some in both), non-trivial enough
// to exercise every branch of jsSimilarity/weightedOverlap/cosine.
func fixtureOverlappingProfiles() (profile, profile) {
	a, b := profile{}, profile{}
	for i := 0; i < 300; i++ {
		tok := fmt.Sprintf("t%03d", i)
		a[tok] = 1.0 / float64(1+i)
		switch {
		case i%3 == 0:
			b[fmt.Sprintf("u%03d", i)] = 2.0 / float64(3+i)
		default:
			b[tok] = 1.0 / float64(2+i)
		}
	}
	return a, b
}

// TestJSSimilarityDeterministicAcrossCalls is a regression test for a
// same-seed nondeterminism bug found while auditing localregime for item 7
// (task27): jsSimilarity's d was a single running sum fed by ranging over a
// map[string]bool built from the union of a and b's keys, so the result
// depended on Go's randomized map iteration order, not just the input.
func TestJSSimilarityDeterministicAcrossCalls(t *testing.T) {
	a, b := fixtureOverlappingProfiles()
	seen := map[uint64]bool{}
	for i := 0; i < 500; i++ {
		seen[math.Float64bits(jsSimilarity(a, b))] = true
	}
	if len(seen) != 1 {
		t.Fatalf("jsSimilarity produced %d distinct float64 bit patterns across 500 calls on identical input", len(seen))
	}
}

// TestWeightedOverlapDeterministicAcrossCalls is the same regression for
// weightedOverlap, whose s was a single running sum fed by ranging directly
// over a.
func TestWeightedOverlapDeterministicAcrossCalls(t *testing.T) {
	a, b := fixtureOverlappingProfiles()
	seen := map[uint64]bool{}
	for i := 0; i < 500; i++ {
		seen[math.Float64bits(weightedOverlap(a, b))] = true
	}
	if len(seen) != 1 {
		t.Fatalf("weightedOverlap produced %d distinct float64 bit patterns across 500 calls on identical input", len(seen))
	}
}

// TestCosineDeterministicAcrossCalls is the same regression for cosine,
// whose n/aa were single running sums fed by ranging directly over a, bb by
// ranging directly over b.
func TestCosineDeterministicAcrossCalls(t *testing.T) {
	a, b := fixtureOverlappingProfiles()
	seen := map[uint64]bool{}
	for i := 0; i < 500; i++ {
		seen[math.Float64bits(cosine(a, b))] = true
	}
	if len(seen) != 1 {
		t.Fatalf("cosine produced %d distinct float64 bit patterns across 500 calls on identical input", len(seen))
	}
}

// TestConcentrationDeterministicAcrossCalls is the same regression for
// concentration, whose s was a single running sum fed by ranging directly
// over p (a single map, not even a union of two - the same bug class still
// applies since summation order still depends on map iteration order).
func TestConcentrationDeterministicAcrossCalls(t *testing.T) {
	a, _ := fixtureOverlappingProfiles()
	seen := map[uint64]bool{}
	for i := 0; i < 500; i++ {
		seen[math.Float64bits(concentration(a))] = true
	}
	if len(seen) != 1 {
		t.Fatalf("concentration produced %d distinct float64 bit patterns across 500 calls on identical input", len(seen))
	}
}
