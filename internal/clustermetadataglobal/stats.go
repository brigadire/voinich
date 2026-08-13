package clustermetadataglobal

import (
	"math"
	"sort"
)

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

func minFloatSlice(x []float64) float64 {
	m := math.Inf(1)
	for _, v := range x {
		if v < m {
			m = v
		}
	}
	return m
}

func medianFloat(x []float64) float64 {
	if len(x) == 0 {
		return 0
	}
	y := append([]float64(nil), x...)
	sort.Float64s(y)
	n := len(y)
	if n%2 == 1 {
		return y[n/2]
	}
	return (y[n/2-1] + y[n/2]) / 2
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

// percentileOf returns the value at the given percentile (0-100) of x.
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

func maxFloat(x []float64) float64 {
	m := math.Inf(-1)
	for _, v := range x {
		if v > m {
			m = v
		}
	}
	return m
}

// exceedances counts null values at least as extreme as observed (one-sided,
// high-tail test: is the observed statistic unusually large).
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
// is never reported.
func empiricalP(exceed, permutations int) float64 {
	return float64(exceed+1) / float64(permutations+1)
}

// EmpiricalSummary is the full set of descriptive fields required for every
// observed statistic (task18 section 7).
type EmpiricalSummary struct {
	Observed     float64
	NullMean     float64
	NullMedian   float64
	NullSD       float64
	NullP95      float64
	NullP99      float64
	NullMax      float64
	Exceedances  int
	EmpiricalP   float64
	Permutations int
}

func summarize(s *StatSeries) EmpiricalSummary {
	exceed := exceedances(s.Null, s.Observed)
	return EmpiricalSummary{
		Observed:     s.Observed,
		NullMean:     meanFloat(s.Null),
		NullMedian:   medianFloat(s.Null),
		NullSD:       sdFloat(s.Null),
		NullP95:      percentileOf(s.Null, 95),
		NullP99:      percentileOf(s.Null, 99),
		NullMax:      maxFloat(s.Null),
		Exceedances:  exceed,
		EmpiricalP:   empiricalP(exceed, len(s.Null)),
		Permutations: len(s.Null),
	}
}
