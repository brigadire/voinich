package positionalcontinuation

import (
	"math"
	"math/rand"
	"sort"
)

type boundaryMetric struct {
	name string
	get  func(SAiinOccurrence) int
}

var boundaryMetrics = []boundaryMetric{
	{"tokens_from_line_start", func(o SAiinOccurrence) int { return o.TokensFromLineStart }},
	{"tokens_to_line_end", func(o SAiinOccurrence) int { return o.TokensToLineEnd }},
	{"tokens_from_block_start", func(o SAiinOccurrence) int { return o.TokensFromBlockStart }},
	{"tokens_to_block_end", func(o SAiinOccurrence) int { return o.TokensToBlockEnd }},
}

// buildBoundaryDistanceRows implements task23 Part N sections 78-82: exact
// boundary distances for s-aiin-chey occurrences compared against all s-aiin
// occurrences, descriptively (median/Q25/Q75) and via a permutation
// comparison of the median difference (chey-labeled vs not, shuffled within
// physical blocks - never across blocks).
func buildBoundaryDistanceRows(occs []SAiinOccurrence, permutations int, seed int64) []BoundaryDistanceRow {
	withX := make([]SAiinOccurrence, 0, len(occs))
	for _, o := range occs {
		if o.X != "" {
			withX = append(withX, o)
		}
	}

	var rows []BoundaryDistanceRow
	r := rand.New(rand.NewSource(seed))
	for _, m := range boundaryMetrics {
		var allVals, cheyVals []float64
		var blockIDs []string
		var isChey []bool
		for _, o := range withX {
			v := float64(m.get(o))
			allVals = append(allVals, v)
			blockIDs = append(blockIDs, o.Block)
			isChey = append(isChey, o.X == FrozenChey)
			if o.X == FrozenChey {
				cheyVals = append(cheyVals, v)
			}
		}

		observedDiff := medianDiff(allVals, isChey)
		var null []float64
		if permutations > 0 {
			null = make([]float64, permutations)
			for p := 0; p < permutations; p++ {
				perm := permuteIsCheyWithinBlocks(blockIDs, isChey, r)
				null[p] = medianDiffAbs(allVals, perm)
			}
		}
		permP := 1.0
		if permutations > 0 {
			permP = empiricalP(math.Abs(observedDiff), null)
		}

		med, q25, q75 := 0.0, 0.0, 0.0
		if len(cheyVals) > 0 {
			med, q25, q75 = median(cheyVals), quantile(cheyVals, 0.25), quantile(cheyVals, 0.75)
		}
		rows = append(rows, BoundaryDistanceRow{Group: "s_aiin_chey", Metric: m.name, Median: med, Q25: q25, Q75: q75, PermutationP: permP})
		rows = append(rows, BoundaryDistanceRow{
			Group: "s_aiin_all", Metric: m.name,
			Median: median(allVals), Q25: quantile(allVals, 0.25), Q75: quantile(allVals, 0.75),
			PermutationP: permP,
		})
	}
	return rows
}

func medianDiff(vals []float64, isChey []bool) float64 {
	var a, b []float64
	for i, v := range vals {
		if isChey[i] {
			a = append(a, v)
		} else {
			b = append(b, v)
		}
	}
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	return median(a) - median(b)
}

func medianDiffAbs(vals []float64, isChey []bool) float64 {
	d := medianDiff(vals, isChey)
	if d < 0 {
		return -d
	}
	return d
}

func permuteIsCheyWithinBlocks(blockIDs []string, isChey []bool, r *rand.Rand) []bool {
	out := make([]bool, len(isChey))
	copy(out, isChey)
	byBlock := map[string][]int{}
	for i, b := range blockIDs {
		byBlock[b] = append(byBlock[b], i)
	}
	keys := make([]string, 0, len(byBlock))
	for k := range byBlock {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		idxs := byBlock[k]
		vals := make([]bool, len(idxs))
		for j, idx := range idxs {
			vals[j] = out[idx]
		}
		r.Shuffle(len(vals), func(a, c int) { vals[a], vals[c] = vals[c], vals[a] })
		for j, idx := range idxs {
			out[idx] = vals[j]
		}
	}
	return out
}
