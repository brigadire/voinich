package conditionalregime

import (
	"math/rand"
	"sort"
)

// residualTransitionCounts counts adjacent-window label transitions R_i ->
// R_j within the pooled residual windows, never bridging across a
// class/block boundary (rw is grouped contiguously by (class, block) in
// position order by construction of buildResidualWindows).
func residualTransitionCounts(rw []ResidualWindow, labels []int, k int) [][]int {
	counts := make([][]int, k)
	for i := range counts {
		counts[i] = make([]int, k)
	}
	for i := 1; i < len(rw); i++ {
		if rw[i].Class != rw[i-1].Class || rw[i].BlockIndex != rw[i-1].BlockIndex {
			continue
		}
		counts[labels[i-1]][labels[i]]++
	}
	return counts
}

// shuffleLabelsWithinBlocks is Null B (task19 section 21) restated directly
// on the label sequence: within each physical block, the assigned residual
// labels are randomly reordered among that block's own window positions.
// This destroys trajectory/transition order while leaving every window's
// own content, and the overall label composition of the block, unchanged -
// exactly the same statistical effect as shuffling window order.
func shuffleLabelsWithinBlocks(rw []ResidualWindow, labels []int, rng *rand.Rand) []int {
	out := append([]int(nil), labels...)
	type key struct {
		c ClassID
		b int
	}
	groups := map[key][]int{}
	for i, w := range rw {
		k := key{w.Class, w.BlockIndex}
		groups[k] = append(groups[k], i)
	}
	keys := make([]key, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].c != keys[j].c {
			return keys[i].c.Label() < keys[j].c.Label()
		}
		return keys[i].b < keys[j].b
	})
	for _, k := range keys {
		idxs := groups[k]
		rng.Shuffle(len(idxs), func(i, j int) { out[idxs[i]], out[idxs[j]] = out[idxs[j]], out[idxs[i]] })
	}
	return out
}

// ResidualTransitionCell is one R_i -> R_j entry of residual_transition_matrix.tsv,
// with its enrichment relative to the Null-B within-block order shuffle
// (task19 section 38: reproducibility, directionality, enrichment).
type ResidualTransitionCell struct {
	From, To int
	Stats    EmpiricalStats
}

// residualTransitionMatrix tests every R_i -> R_j transition against Null B.
func residualTransitionMatrix(rw []ResidualWindow, labels []int, k, permutations int, seed int64) []ResidualTransitionCell {
	observed := residualTransitionCounts(rw, labels, k)
	nullCounts := make([][][]float64, k)
	for i := range nullCounts {
		nullCounts[i] = make([][]float64, k)
		for j := range nullCounts[i] {
			nullCounts[i][j] = make([]float64, permutations)
		}
	}
	rng := rand.New(rand.NewSource(seed))
	for p := 0; p < permutations; p++ {
		shuffled := shuffleLabelsWithinBlocks(rw, labels, rng)
		counts := residualTransitionCounts(rw, shuffled, k)
		for i := 0; i < k; i++ {
			for j := 0; j < k; j++ {
				nullCounts[i][j][p] = float64(counts[i][j])
			}
		}
	}
	out := make([]ResidualTransitionCell, 0, k*k)
	for i := 0; i < k; i++ {
		for j := 0; j < k; j++ {
			out = append(out, ResidualTransitionCell{i, j, buildEmpiricalStats(float64(observed[i][j]), nullCounts[i][j])})
		}
	}
	return out
}
