package structuralreliability

import (
	"math"
	"sort"

	"zcore.dev/voinich/internal/profilestability"
)

// SummarizeStat reduces a slice of observations to Stat using the shared
// profilestability.Summarize implementation, so mean/stddev/percentile math
// is computed in exactly one place across the whole project.
func SummarizeStat(values []float64) Stat {
	d := profilestability.Summarize(values, true)
	return Stat{Observations: d.Observations, Mean: d.Mean, Median: d.Percentile50, Stddev: d.Stddev}
}

// SummarizePercentileStat is SummarizeStat plus the 90th/95th percentiles
// needed by the pair-stability and CI-width sections.
func SummarizePercentileStat(values []float64) PercentileStat {
	stat := SummarizeStat(values)
	result := PercentileStat{Observations: stat.Observations, Mean: stat.Mean, Median: stat.Median, Stddev: stat.Stddev}
	if len(values) == 0 {
		return result
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	result.Percentile90 = profilestability.Percentile(sorted, .90)
	result.Percentile95 = profilestability.Percentile(sorted, .95)
	return result
}

// PercentileOf sorts a copy of values and returns the linearly-interpolated
// percentile at the given probability (0-1).
func PercentileOf(values []float64, probability float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	return profilestability.Percentile(sorted, probability)
}

// Log2 is a tiny named wrapper so call sites read as "log2(count)" rather
// than a bare math.Log2, matching the task's own notation.
func Log2(value float64) float64 { return math.Log2(value) }

// GeometricMean of two positive counts, used for pair-level sample size and
// for pair reliability (geometric mean of the two per-token reliabilities).
func GeometricMean(a, b float64) float64 {
	if a <= 0 || b <= 0 {
		return 0
	}
	return math.Sqrt(a * b)
}

// Entropy is the Shannon entropy in bits of a frequency distribution, used
// for context diversity (section 23). It is a generic formula independent
// of the structural similarity model. Keys are visited in a fixed sorted
// order so floating-point summation order - and therefore the last-bit
// result - never depends on Go's randomized map iteration order (required
// for the byte-identical YAML of task section 29).
func Entropy(counts map[string]int) float64 {
	total := 0
	for _, count := range counts {
		total += count
	}
	if total == 0 {
		return 0
	}
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	value := 0.0
	for _, key := range keys {
		count := counts[key]
		if count == 0 {
			continue
		}
		p := float64(count) / float64(total)
		value -= p * math.Log2(p)
	}
	return value
}

// Spearman computes the Spearman rank correlation between two equal-length
// series, averaging ranks on ties, then reusing profilestability.Pearson on
// the resulting ranks. Spearman correlation is invariant under any strictly
// monotonic transform of either series, so it is equally valid whether the
// caller passes raw counts or log2(counts).
func Spearman(x, y []float64) (rho float64, n int) {
	if len(x) != len(y) || len(x) < 3 {
		return 0, len(x)
	}
	return profilestability.Pearson(rank(x), rank(y)), len(x)
}

// rank assigns 1-based ranks to values, using the average rank for ties.
func rank(values []float64) []float64 {
	type indexed struct {
		value float64
		index int
	}
	items := make([]indexed, len(values))
	for i, value := range values {
		items[i] = indexed{value: value, index: i}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].value < items[j].value })
	ranks := make([]float64, len(values))
	i := 0
	for i < len(items) {
		j := i
		for j+1 < len(items) && items[j+1].value == items[i].value {
			j++
		}
		averageRank := float64(i+j)/2 + 1
		for k := i; k <= j; k++ {
			ranks[items[k].index] = averageRank
		}
		i = j + 1
	}
	return ranks
}
