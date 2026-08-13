package clustermetadataglobal

import (
	"fmt"
	"math"
	"sort"
)

// maxLabelsOrClusters bounds the dense fixed-size accumulators used by
// fastMetrics. Currier/hand have at most 8 known labels and the frozen K
// sweep tops out at 15, both comfortably under this bound.
const maxLabelsOrClusters = 16

// encodeLabels maps distinct non-empty labels to dense ascending integer
// codes (so majority-vote ties resolve to the lexicographically smallest
// label, matching metadatavalidation.MetadataComposition) and "" to -1.
func encodeLabels(labels []string) (codes []int8, values []string, valueIndex map[string]int8) {
	seen := map[string]bool{}
	for _, v := range labels {
		if v != "" {
			seen[v] = true
		}
	}
	values = make([]string, 0, len(seen))
	for v := range seen {
		values = append(values, v)
	}
	sort.Strings(values)
	valueIndex = make(map[string]int8, len(values))
	for i, v := range values {
		valueIndex[v] = int8(i)
	}
	codes = codesFromLabels(labels, valueIndex)
	return
}

func codesFromLabels(labels []string, valueIndex map[string]int8) []int8 {
	out := make([]int8, len(labels))
	for i, v := range labels {
		if v == "" {
			out[i] = -1
		} else {
			out[i] = valueIndex[v]
		}
	}
	return out
}

// buildCumulativeInto fills cum, sized (len(codes)+1)*numLabels, so that
// cum[p*numLabels+c] is the number of tokens with label code c among
// codes[0:p]. The caller-owned buffer is reused across permutations to avoid
// reallocating on every one of the (typically 10000) replicates.
func buildCumulativeInto(codes []int8, numLabels int, cum []int32) {
	for c := 0; c < numLabels; c++ {
		cum[c] = 0
	}
	for i, code := range codes {
		base := i * numLabels
		next := base + numLabels
		copy(cum[next:next+numLabels], cum[base:base+numLabels])
		if code >= 0 {
			cum[next+int(code)]++
		}
	}
}

// windowMajority returns the majority label code, its purity and whether the
// window has any known-label token, from a cumulative count table built by
// buildCumulativeInto. Ties resolve to the smallest code (lexicographically
// smallest label), matching metadatavalidation.MetadataComposition.
func windowMajority(cum []int32, numLabels, start, end int) (code int8, purity float64, known bool) {
	sBase, eBase := start*numLabels, end*numLabels
	best, bestCount, total := int8(-1), int32(0), int32(0)
	for c := 0; c < numLabels; c++ {
		v := cum[eBase+c] - cum[sBase+c]
		total += v
		if v > bestCount {
			bestCount, best = v, int8(c)
		}
	}
	if total == 0 {
		return -1, 0, false
	}
	return best, float64(bestCount) / float64(total), true
}

// majorityPerWindow computes the majority label code for every frozen window
// in ranges, sharing one cumulative table across the whole window_size x
// method x K search space for this permutation replicate.
func majorityPerWindow(cum []int32, numLabels int, ranges []WindowRange) []int8 {
	out := make([]int8, len(ranges))
	for i, r := range ranges {
		code, _, _ := windowMajority(cum, numLabels, r.Start, r.End)
		out[i] = code
	}
	return out
}

// fastMetrics computes NMI and ARI between per-window majority label codes
// and frozen per-window cluster ids, restricted to eligible window
// positions, using fixed-size dense accumulators instead of maps. It is a
// performance-optimized restatement of the exact same MI/NMI/ARI formulas as
// metadatavalidation.AssociationMetrics (verified equal by unit test), not a
// change to the similarity formulas themselves.
func fastMetrics(labelCodes []int8, clusterCodes []int, eligible []int, numLabels, numClusters int) (nmi, ari float64) {
	if numLabels > maxLabelsOrClusters || numClusters > maxLabelsOrClusters {
		panic(fmt.Sprintf("clustermetadataglobal: fastMetrics bound exceeded: numLabels=%d numClusters=%d", numLabels, numClusters))
	}
	var table [maxLabelsOrClusters * maxLabelsOrClusters]int32
	var rowSum, colSum [maxLabelsOrClusters]int32
	var n int32
	for _, i := range eligible {
		lc := labelCodes[i]
		if lc < 0 {
			continue
		}
		cc := clusterCodes[i]
		table[int(lc)*numClusters+cc]++
		rowSum[lc]++
		colSum[cc]++
		n++
	}
	if n == 0 {
		return 0, 0
	}
	mi := 0.0
	fn := float64(n)
	for l := 0; l < numLabels; l++ {
		if rowSum[l] == 0 {
			continue
		}
		for c := 0; c < numClusters; c++ {
			v := table[l*numClusters+c]
			if v == 0 {
				continue
			}
			p := float64(v) / fn
			mi += p * math.Log(p/(float64(rowSum[l])*float64(colSum[c])/(fn*fn)))
		}
	}
	ha := entropyFromCounts(rowSum[:numLabels], n)
	hb := entropyFromCounts(colSum[:numClusters], n)
	if ha+hb > 0 {
		nmi = 2 * mi / (ha + hb)
	}
	comb := func(x int32) float64 { return float64(x) * float64(x-1) / 2 }
	sumCell, sumA, sumB := 0.0, 0.0, 0.0
	for i := 0; i < numLabels*numClusters; i++ {
		sumCell += comb(table[i])
	}
	for l := 0; l < numLabels; l++ {
		sumA += comb(rowSum[l])
	}
	for c := 0; c < numClusters; c++ {
		sumB += comb(colSum[c])
	}
	den := comb(n)
	if den != 0 {
		expected := sumA * sumB / den
		d := .5*(sumA+sumB) - expected
		if d != 0 {
			ari = (sumCell - expected) / d
		}
	}
	return nmi, ari
}

func entropyFromCounts(counts []int32, n int32) float64 {
	h := 0.0
	fn := float64(n)
	for _, v := range counts {
		if v == 0 {
			continue
		}
		p := float64(v) / fn
		h -= p * math.Log(p)
	}
	return h
}
