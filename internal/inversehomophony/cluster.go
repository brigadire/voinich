package inversehomophony

import (
	"math"
	"sort"
)

// unionFind is a standard union-find with size and occurrence-weight
// tracking, so class-size-fraction and entropy constraints can be checked
// before committing a merge.
type unionFind struct {
	parent map[string]string
	weight map[string]int // occurrence-weighted size of each root's class
	total  int
}

func newUnionFind(freq map[string]int) *unionFind {
	uf := &unionFind{parent: make(map[string]string, len(freq)), weight: make(map[string]int, len(freq))}
	for t, f := range freq {
		uf.parent[t] = t
		uf.weight[t] = f
		uf.total += f
	}
	return uf
}

func (uf *unionFind) find(x string) string {
	for uf.parent[x] != x {
		uf.parent[x] = uf.parent[uf.parent[x]]
		x = uf.parent[x]
	}
	return x
}

func (uf *unionFind) union(a, b string) {
	ra, rb := uf.find(a), uf.find(b)
	if ra == rb {
		return
	}
	// Deterministic tie-break: smaller root string becomes the new root.
	if rb < ra {
		ra, rb = rb, ra
	}
	uf.weight[ra] += uf.weight[rb]
	delete(uf.weight, rb)
	uf.parent[rb] = ra
}

// classSizeFraction returns the occurrence-weighted fraction the class of
// x would have if a and b's classes were merged (without mutating uf).
func mergedFraction(uf *unionFind, ra, rb string) float64 {
	if uf.total == 0 {
		return 0
	}
	return float64(uf.weight[ra]+uf.weight[rb]) / float64(uf.total)
}

// classEntropy is the occurrence-weighted Shannon entropy (base 2, in
// bits) of the current class-size distribution. Roots are visited in
// sorted order, not map iteration order: floating-point addition is not
// associative, so summing in map order would make the accumulated result
// (and therefore every downstream merge/accept decision that compares it
// to minEntropy) depend on Go's randomized map iteration rather than only
// on the input (see the project's "Go map iteration determinism" convention).
func classEntropy(uf *unionFind) float64 {
	if uf.total == 0 {
		return 0
	}
	roots := make([]string, 0, len(uf.weight))
	for r := range uf.weight {
		roots = append(roots, r)
	}
	sort.Strings(roots)
	var h float64
	for _, r := range roots {
		w := uf.weight[r]
		if w == 0 {
			continue
		}
		p := float64(w) / float64(uf.total)
		h -= p * math.Log2(p)
	}
	return h
}

// Recover runs the frozen clustering/search method (task57 section 6/17)
// over pairs, starting from the NO_COLLAPSE partition implied by freq, and
// returns the resulting Partition plus the full merge audit trail
// (accepted and rejected candidates alike, in the order they were
// considered).
func Recover(freq map[string]int, pairs []PairScore, cfg Config) (Partition, []MergeEvent) {
	uf := newUnionFind(freq)
	noCollapseEntropy := classEntropy(uf)
	minEntropy := cfg.MinEntropyFraction * noCollapseEntropy

	events := make([]MergeEvent, 0, len(pairs))
	for _, p := range pairs {
		if p.Score <= cfg.Threshold {
			events = append(events, MergeEvent{A: p.A, B: p.B, Score: p.Score, Accepted: false, Reason: "below_threshold"})
			continue
		}
		ra, rb := uf.find(p.A), uf.find(p.B)
		if ra == rb {
			events = append(events, MergeEvent{A: p.A, B: p.B, Score: p.Score, Accepted: false, Reason: "already_same_class"})
			continue
		}
		frac := mergedFraction(uf, ra, rb)
		if frac > cfg.MaxClassFraction {
			events = append(events, MergeEvent{A: p.A, B: p.B, Score: p.Score, Accepted: false, Reason: "max_class_fraction"})
			continue
		}
		// Speculatively apply the merge, check the entropy floor, and
		// roll back if it would be violated - entropy of the *whole*
		// partition, not just the merged class, must be checked after
		// the fact.
		savedWeight := map[string]int{ra: uf.weight[ra], rb: uf.weight[rb]}
		savedParentB := uf.parent[rb]
		uf.union(p.A, p.B)
		if h := classEntropy(uf); h < minEntropy {
			// Roll back.
			uf.parent[rb] = savedParentB
			uf.weight[ra] = savedWeight[ra]
			uf.weight[rb] = savedWeight[rb]
			events = append(events, MergeEvent{A: p.A, B: p.B, Score: p.Score, Accepted: false, Reason: "min_entropy_fraction"})
			continue
		}
		newRoot := uf.find(p.A)
		events = append(events, MergeEvent{
			A: p.A, B: p.B, Score: p.Score, Accepted: true, Reason: "merged",
			ClassSizeAfter:     uf.weight[newRoot],
			ClassFractionAfter: float64(uf.weight[newRoot]) / float64(uf.total),
		})
	}

	partition := make(Partition, len(freq))
	for t := range freq {
		partition[t] = uf.find(t)
	}
	return partition, events
}

// PartitionClassSizes returns the occurrence-weighted size of every class
// in p, sorted descending (used for RANDOM_PARTITION matching and
// diagnostics).
func PartitionClassSizes(p Partition, freq map[string]int) []int {
	sizes := make(map[string]int)
	for t, c := range p {
		sizes[c] += freq[t]
	}
	out := make([]int, 0, len(sizes))
	for _, s := range sizes {
		out = append(out, s)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(out)))
	return out
}
