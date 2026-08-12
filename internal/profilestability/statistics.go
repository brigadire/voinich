package profilestability

import (
	"math"
	"sort"
)

func Summarize(values []float64, percentiles bool) Distribution {
	result := Distribution{Observations: len(values)}
	if len(values) == 0 {
		return result
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	result.Min, result.Max = sorted[0], sorted[len(sorted)-1]
	result.Range = result.Max - result.Min
	for _, value := range sorted {
		result.Mean += value
	}
	result.Mean /= float64(len(sorted))
	for _, value := range sorted {
		difference := value - result.Mean
		result.Stddev += difference * difference
	}
	result.Stddev = math.Sqrt(result.Stddev / float64(len(sorted)))
	if result.Stddev < 1e-15 {
		result.Stddev = 0
	}
	if percentiles {
		result.Percentile025 = Percentile(sorted, .025)
		result.Percentile50 = Percentile(sorted, .5)
		result.Percentile975 = Percentile(sorted, .975)
	}
	return result
}

// Percentile returns the linearly-interpolated value at the given probability
// (0-1) within an already ascending-sorted slice. Exported for reuse by
// downstream analyses that need percentiles beyond the fixed 2.5/50/97.5 set.
func Percentile(sorted []float64, probability float64) float64 {
	if len(sorted) == 1 {
		return sorted[0]
	}
	position := probability * float64(len(sorted)-1)
	lower, upper := int(math.Floor(position)), int(math.Ceil(position))
	if lower == upper {
		return sorted[lower]
	}
	weight := position - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// Jaccard is the exported form of the neighbor-set overlap ratio, reused by
// downstream reliability analyses that recompute neighbor sets per threshold.
func Jaccard(left, right []Neighbor) float64 {
	leftSet, rightSet := make(map[string]bool), make(map[string]bool)
	for _, item := range left {
		leftSet[item.Token] = true
	}
	for _, item := range right {
		rightSet[item.Token] = true
	}
	intersection := 0
	for token := range leftSet {
		if rightSet[token] {
			intersection++
		}
	}
	union := len(leftSet) + len(rightSet) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

// OverlapAt is the exported top-K neighbor overlap ratio (intersection over
// min(k, available)), reused by downstream reliability analyses.
func OverlapAt(left, right []Neighbor, k int) float64 {
	if len(left) < k {
		k = len(left)
	}
	if len(right) < k {
		k = len(right)
	}
	if k == 0 {
		return 0
	}
	set := make(map[string]bool, k)
	for _, item := range left[:k] {
		set[item.Token] = true
	}
	intersection := 0
	for _, item := range right[:k] {
		if set[item.Token] {
			intersection++
		}
	}
	return float64(intersection) / float64(k)
}

func spearmanCommon(left, right []Neighbor) (float64, int, bool) {
	ranksLeft, ranksRight := make(map[string]int), make(map[string]int)
	for i, item := range left {
		ranksLeft[item.Token] = i + 1
	}
	for i, item := range right {
		ranksRight[item.Token] = i + 1
	}
	var tokens []string
	for token := range ranksLeft {
		if ranksRight[token] > 0 {
			tokens = append(tokens, token)
		}
	}
	if len(tokens) < 3 {
		return 0, len(tokens), false
	}
	sort.Strings(tokens)
	leftValues, rightValues := make([]float64, len(tokens)), make([]float64, len(tokens))
	for i, token := range tokens {
		leftValues[i] = float64(ranksLeft[token])
		rightValues[i] = float64(ranksRight[token])
	}
	return Pearson(leftValues, rightValues), len(tokens), true
}

// Pearson is the exported linear correlation coefficient, reused by
// downstream analyses to compute rank correlations on top of Pearson once
// values have already been converted to ranks.
func Pearson(left, right []float64) float64 {
	if len(left) != len(right) || len(left) == 0 {
		return 0
	}
	meanLeft, meanRight := 0.0, 0.0
	for i := range left {
		meanLeft += left[i]
		meanRight += right[i]
	}
	meanLeft /= float64(len(left))
	meanRight /= float64(len(right))
	numerator, leftSquare, rightSquare := 0.0, 0.0, 0.0
	for i := range left {
		a, b := left[i]-meanLeft, right[i]-meanRight
		numerator += a * b
		leftSquare += a * a
		rightSquare += b * b
	}
	if leftSquare == 0 || rightSquare == 0 {
		return 0
	}
	return numerator / math.Sqrt(leftSquare*rightSquare)
}

func frequencyBin(count int) string {
	switch {
	case count < 20:
		return "10-19"
	case count < 40:
		return "20-39"
	case count < 80:
		return "40-79"
	case count < 160:
		return "80-159"
	default:
		return "160+"
	}
}

var frequencyBins = []string{"10-19", "20-39", "40-79", "80-159", "160+"}
