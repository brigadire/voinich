package conditionalregime

import (
	"math"
	"sort"
)

// cappedSampleIndices deterministically subsamples n items down to at most
// cap, using the same uniform-stride scheme global-regime-analyze's own
// clusteringSample uses. It is reused, with a smaller cap, inside the
// permutation loops: quadratic clustering cost is what makes a 1000-replicate
// null tractable, and the null distribution's own resolution does not
// require the same fitting-sample size as the one-shot observed fit.
func cappedSampleIndices(n, cap int) []int {
	if n <= cap {
		idx := make([]int, n)
		for i := range idx {
			idx[i] = i
		}
		return idx
	}
	idx := make([]int, cap)
	for i := range idx {
		idx[i] = i * (n - 1) / (cap - 1)
	}
	return idx
}

func meanFloat(x []float64) float64 {
	if len(x) == 0 {
		return 0
	}
	s := 0.0
	for _, v := range x {
		s += v
	}
	return s / float64(len(x))
}

func sdFloat(x []float64) float64 {
	if len(x) < 2 {
		return 0
	}
	m := meanFloat(x)
	s := 0.0
	for _, v := range x {
		d := v - m
		s += d * d
	}
	return math.Sqrt(s / float64(len(x)-1))
}

func maxFloat(x []float64) float64 {
	m := math.Inf(-1)
	for _, v := range x {
		if v > m {
			m = v
		}
	}
	if math.IsInf(m, -1) {
		return 0
	}
	return m
}

func percentileOf(x []float64, p float64) float64 {
	if len(x) == 0 {
		return 0
	}
	y := append([]float64(nil), x...)
	sort.Float64s(y)
	idx := int(math.Ceil(p/100*float64(len(y)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(y) {
		idx = len(y) - 1
	}
	return y[idx]
}

// exceedances counts null values at least as extreme as observed (one-sided,
// high-tail: is the observed statistic unusually large).
func exceedances(null []float64, observed float64) int {
	n := 0
	for _, v := range null {
		if v >= observed {
			n++
		}
	}
	return n
}

// empiricalP applies the required +1 correction so a p-value of exactly zero
// is never reported (task19 section 43).
func empiricalP(exceed, permutations int) float64 {
	return float64(exceed+1) / float64(permutations+1)
}

// buildEmpiricalStats assembles the full descriptive set task19 section 43
// requires for every permutation-based statistic: never just a p-value.
func buildEmpiricalStats(observed float64, null []float64) EmpiricalStats {
	sd := sdFloat(null)
	effect := 0.0
	if sd > 0 {
		effect = (observed - meanFloat(null)) / sd
	}
	exceed := exceedances(null, observed)
	return EmpiricalStats{
		Observed: observed, NullMean: meanFloat(null), NullSD: sd,
		NullP95: percentileOf(null, 95), NullP99: percentileOf(null, 99), NullMax: maxFloat(null),
		EffectSize: effect, Exceedances: exceed, EmpiricalP: empiricalP(exceed, len(null)), Permutations: len(null),
	}
}
