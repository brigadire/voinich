package metadatavalidation

import (
	"fmt"
	"math"
	"testing"
)

// TestAssociationMetricsDeterministicAcrossCalls is a regression test for a
// same-seed nondeterminism bug found while validating the UniformBoundaries/
// clusterPermutationSummary optimizations (task27): AssociationMetrics
// summed mi, sumCell, sumA, and sumB — each a single running total fed by
// every entry of a map (tab, tab's rows, ra, cb respectively) — via
// `range`, so their results depended on Go's randomized map iteration
// order, not just the input. AssociationMetrics is exported and used by
// internal/conditionalregime, internal/residualdiagnostic, and
// internal/clustermetadataglobal (tests), so this affected MI/NMI/ARI
// computations package-wide, not just metadatavalidation. Confirmed
// against the pre-fix implementation: 24-28 distinct float64 bit patterns
// for MI across 500 calls on identical input. Fixed by summing over sorted
// (row, column) keys.
func TestAssociationMetricsDeterministicAcrossCalls(t *testing.T) {
	a := make([]string, 0, 400)
	b := make([]string, 0, 400)
	for i := 0; i < 400; i++ {
		a = append(a, fmt.Sprintf("L%d", i%17))
		b = append(b, fmt.Sprintf("C%d", (i*7+3)%13))
	}
	seenMI, seenNMI, seenARI, seenHom, seenComp := map[uint64]bool{}, map[uint64]bool{}, map[uint64]bool{}, map[uint64]bool{}, map[uint64]bool{}
	for i := 0; i < 500; i++ {
		m := AssociationMetrics(a, b)
		seenMI[math.Float64bits(m.MI)] = true
		seenNMI[math.Float64bits(m.NMI)] = true
		seenARI[math.Float64bits(m.ARI)] = true
		seenHom[math.Float64bits(m.Homogeneity)] = true
		seenComp[math.Float64bits(m.Completeness)] = true
	}
	if len(seenMI) != 1 || len(seenNMI) != 1 || len(seenARI) != 1 || len(seenHom) != 1 || len(seenComp) != 1 {
		t.Fatalf("AssociationMetrics produced distinct bit patterns across 500 calls on identical input: MI=%d NMI=%d ARI=%d Hom=%d Comp=%d", len(seenMI), len(seenNMI), len(seenARI), len(seenHom), len(seenComp))
	}
}

// TestEntropyCountsDeterministicAcrossCalls is a regression test for a
// same-seed nondeterminism bug: entropyCounts summed -p*log(p) over
// `range c` (a map[string]int), a single running total fed by every key.
// Confirmed nondeterministic before the fix (sortedIntMapSum).
func TestEntropyCountsDeterministicAcrossCalls(t *testing.T) {
	c := map[string]int{}
	for i := 0; i < 200; i++ {
		c[fmt.Sprintf("k%03d", i)] = 1 + i%11
	}
	seen := map[uint64]bool{}
	for i := 0; i < 500; i++ {
		seen[math.Float64bits(entropyCounts(c, 5000))] = true
	}
	if len(seen) != 1 {
		t.Fatalf("entropyCounts produced %d distinct bit patterns across 500 calls on identical input", len(seen))
	}
}

// TestConditionalEntropyDeterministicAcrossCalls is a regression test for a
// same-seed nondeterminism bug: conditionalEntropy summed
// sizes[k]/n*entropyCounts(...) over `range by` (a
// map[string]map[string]int), a single running total fed by every cluster
// key. Confirmed nondeterministic before the fix.
func TestConditionalEntropyDeterministicAcrossCalls(t *testing.T) {
	labels := make([]string, 0, 400)
	clusters := make([]string, 0, 400)
	for i := 0; i < 400; i++ {
		labels = append(labels, fmt.Sprintf("L%d", i%23))
		clusters = append(clusters, fmt.Sprintf("C%d", (i*7)%19))
	}
	seen := map[uint64]bool{}
	for i := 0; i < 500; i++ {
		seen[math.Float64bits(conditionalEntropy(labels, clusters))] = true
	}
	if len(seen) != 1 {
		t.Fatalf("conditionalEntropy produced %d distinct bit patterns across 500 calls on identical input", len(seen))
	}
}
