package transitionnetwork

import (
	"math"
	"sort"
)

func mean(x []float64) float64 {
	if len(x) == 0 {
		return 0
	}
	s := 0.
	for _, v := range x {
		s += v
	}
	return s / float64(len(x))
}
func median(x []float64) float64 {
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
func sd(x []float64) float64 {
	if len(x) < 2 {
		return 0
	}
	m := mean(x)
	s := 0.
	for _, v := range x {
		d := v - m
		s += d * d
	}
	return math.Sqrt(s / float64(len(x)-1))
}
func pearson(a, b []float64) float64 {
	if len(a) != len(b) || len(a) < 2 {
		return 0
	}
	ma, mb := mean(a), mean(b)
	var n, da, db float64
	for i := range a {
		x, y := a[i]-ma, b[i]-mb
		n += x * y
		da += x * x
		db += y * y
	}
	if da == 0 || db == 0 {
		return 0
	}
	return n / math.Sqrt(da*db)
}
func ranks(x []float64) []float64 {
	idx := make([]int, len(x))
	for i := range idx {
		idx[i] = i
	}
	sort.Slice(idx, func(i, j int) bool { return x[idx[i]] < x[idx[j]] })
	r := make([]float64, len(x))
	for i := 0; i < len(idx); {
		j := i + 1
		for j < len(idx) && x[idx[j]] == x[idx[i]] {
			j++
		}
		v := float64(i+j-1)/2 + 1
		for k := i; k < j; k++ {
			r[idx[k]] = v
		}
		i = j
	}
	return r
}
func spearman(a, b []float64) float64 { return pearson(ranks(a), ranks(b)) }
func entropy(p []float64) float64 {
	s := 0.
	for _, v := range p {
		if v > 0 {
			s -= v * math.Log(v)
		}
	}
	return s
}
func jsSimilarity(a, b []float64) float64 {
	m := make([]float64, len(a))
	for i := range a {
		m[i] = (a[i] + b[i]) / 2
	}
	kl := func(x, y []float64) float64 {
		s := 0.
		for i := range x {
			if x[i] > 0 {
				s += x[i] * math.Log2(x[i]/y[i])
			}
		}
		return s
	}
	return 1 - (kl(a, m)+kl(b, m))/2
}
func bh(rows []*EdgeSummary, positive bool) {
	var v []*EdgeSummary
	for _, r := range rows {
		if (r.ExpectedSign == "preferred") == positive && r.EligibleBlocks >= 3 {
			v = append(v, r)
		}
	}
	sort.Slice(v, func(i, j int) bool { return v[i].EmpiricalP < v[j].EmpiricalP })
	q := 1.
	for i := len(v) - 1; i >= 0; i-- {
		x := v[i].EmpiricalP * float64(len(v)) / float64(i+1)
		if x > 1 {
			x = 1
		}
		if x < q {
			q = x
		}
		v[i].FDRQ = q
	}
}
