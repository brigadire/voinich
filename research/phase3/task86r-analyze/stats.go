package main

import (
	"math"
	"sort"
)

// sortedKeysStr returns sorted keys of a string-keyed map, used to force
// deterministic iteration order before any float64 accumulation per
// G1_SEED_CONTRACT.md.
func sortedKeysStr(m map[string]float64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sumSorted(m map[string]float64) float64 {
	total := 0.0
	for _, k := range sortedKeysStr(m) {
		total += m[k]
	}
	return total
}

// median is the sample median (mean of two central values for even n).
// values is sorted in place.
func median(values []float64) float64 {
	v := append([]float64(nil), values...)
	sort.Float64s(v)
	n := len(v)
	if n == 0 {
		return math.NaN()
	}
	if n%2 == 1 {
		return v[n/2]
	}
	return (v[n/2-1] + v[n/2]) / 2
}

// nearestRankQ95 is the nearest-rank 0.95 quantile at one-based index
// ceil(0.95*n), over a sorted ascending copy of values.
func nearestRankQ95(values []float64) float64 {
	v := append([]float64(nil), values...)
	sort.Float64s(v)
	n := len(v)
	if n == 0 {
		return math.NaN()
	}
	idx := int(math.Ceil(0.95 * float64(n)))
	if idx < 1 {
		idx = 1
	}
	if idx > n {
		idx = n
	}
	return v[idx-1]
}

// mfcDispersion computes the calibration dispersion for one MFC generator's
// 16-value sample: nearest-rank q0.95 of abs(x_i - median).
func mfcDispersion(values []float64) (float64, bool) {
	for _, x := range values {
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return math.NaN(), false
		}
	}
	center := median(values)
	dev := make([]float64, len(values))
	for i, x := range values {
		dev[i] = math.Abs(x - center)
	}
	return nearestRankQ95(dev), true
}

func mean(values []float64) float64 {
	if len(values) == 0 {
		return math.NaN()
	}
	total := 0.0
	for _, v := range values {
		total += v
	}
	return total / float64(len(values))
}

func sampleSD(values []float64) float64 {
	n := len(values)
	if n < 2 {
		return 0
	}
	m := mean(values)
	ss := 0.0
	for _, v := range values {
		d := v - m
		ss += d * d
	}
	return math.Sqrt(ss / float64(n-1))
}

func coefficientOfVariation(values []float64) float64 {
	m := mean(values)
	sd := sampleSD(values)
	if m != 0 {
		return sd / math.Abs(m)
	}
	return sd
}

// jsDivergence computes the Jensen-Shannon divergence (base 2) between two
// probability distributions given as parallel slices over the same
// outcome ordering; both must sum to (approximately) 1.
func jsDivergence(p, q []float64) float64 {
	kl := func(a, b []float64) float64 {
		s := 0.0
		for i := range a {
			if a[i] <= 0 {
				continue
			}
			m := b[i]
			if m <= 0 {
				m = 1e-300
			}
			s += a[i] * math.Log2(a[i]/m)
		}
		return s
	}
	m := make([]float64, len(p))
	for i := range p {
		m[i] = (p[i] + q[i]) / 2
	}
	return 0.5*kl(p, m) + 0.5*kl(q, m)
}

// theilSenSlope computes the median pairwise slope of y vs x (log-log
// caller responsibility) over parallel slices.
func theilSenSlope(x, y []float64) float64 {
	var slopes []float64
	for i := 0; i < len(x); i++ {
		for j := i + 1; j < len(x); j++ {
			if x[j] == x[i] {
				continue
			}
			slopes = append(slopes, (y[j]-y[i])/(x[j]-x[i]))
		}
	}
	if len(slopes) == 0 {
		return math.NaN()
	}
	return median(slopes)
}

// percentileNearestRank returns the nearest-rank percentile (0..1) of a
// sorted-ascending copy of values, one-based index ceil(p*n), clamped.
func percentileNearestRank(values []float64, p float64) float64 {
	v := append([]float64(nil), values...)
	sort.Float64s(v)
	n := len(v)
	if n == 0 {
		return math.NaN()
	}
	idx := int(math.Ceil(p * float64(n)))
	if idx < 1 {
		idx = 1
	}
	if idx > n {
		idx = n
	}
	return v[idx-1]
}

// giniCoefficient over non-negative support values (mean absolute
// difference form), deterministic given sorted input.
func giniCoefficient(values []float64) float64 {
	n := len(values)
	if n == 0 {
		return math.NaN()
	}
	v := append([]float64(nil), values...)
	sort.Float64s(v)
	var sumAbsDiff, total float64
	for i := 0; i < n; i++ {
		total += v[i]
		for j := 0; j < n; j++ {
			sumAbsDiff += math.Abs(v[i] - v[j])
		}
	}
	if total == 0 {
		return 0
	}
	return sumAbsDiff / (2 * float64(n) * total)
}
