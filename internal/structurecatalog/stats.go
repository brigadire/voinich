package structurecatalog

import (
	"math"
	"sort"
)

func ratio(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) / float64(b)
}
func effect(obs, exp float64) float64 {
	if exp == 0 {
		if obs > 0 {
			return math.Inf(1)
		}
		return 0
	}
	return obs / exp
}

// poissonTail returns a directional one-sided tail. It is deterministic and
// is used as an analytic approximation to the frozen, marginal-preserving null.
func poissonTail(k int, lambda float64, upper bool) float64 {
	if lambda <= 0 {
		if k == 0 {
			return 1
		}
		return 0
	}
	if lambda > 100 {
		z := (float64(k) + .5 - lambda) / math.Sqrt(lambda)
		cdf := .5 * (1 + math.Erf(z/math.Sqrt2))
		if upper {
			return clamp(1 - cdf)
		}
		return clamp(cdf)
	}
	p := math.Exp(-lambda)
	cdf := p
	for i := 1; i <= k; i++ {
		p *= lambda / float64(i)
		cdf += p
	}
	if upper {
		if k == 0 {
			return 1
		}
		return clamp(1 - (cdf - p))
	}
	return clamp(cdf)
}
func clamp(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

func applyBH(rules []Rule) {
	type pv struct {
		i int
		p float64
	}
	a := make([]pv, 0, len(rules))
	for i := range rules {
		if !math.IsNaN(rules[i].PRaw) {
			a = append(a, pv{i, rules[i].PRaw})
		}
	}
	sort.Slice(a, func(i, j int) bool { return a[i].p < a[j].p })
	q := 1.0
	for j := len(a) - 1; j >= 0; j-- {
		v := a[j].p * float64(len(a)) / float64(j+1)
		if v < q {
			q = v
		}
		rules[a[j].i].QValue = clamp(q)
	}
	for i := range rules {
		classify(&rules[i])
	}
}

func classify(r *Rule) {
	if r.OpportunityCount == 0 {
		r.InferredStatus = "INSUFFICIENT_SUPPORT"
		return
	}
	if math.IsNaN(r.PRaw) {
		r.InferredStatus = "NOT_TESTED"
		return
	}
	if r.QValue > .05 {
		r.InferredStatus = "EXPECTED"
		return
	}
	preferred := float64(r.ObservedCount) > r.ExpectedCount
	if preferred {
		if r.QValue <= .01 {
			r.InferredStatus = "STRONGLY_PREFERRED"
		} else {
			r.InferredStatus = "PREFERRED"
		}
		return
	}
	if r.ObservedCount == 0 {
		if r.QValue <= .01 {
			r.InferredStatus = "STRONGLY_AVOIDED"
		} else {
			r.InferredStatus = "AVOIDED"
		}
		return
	}
	r.InferredStatus = "DEPLETED"
}

func median(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	y := append([]float64(nil), xs...)
	sort.Float64s(y)
	m := len(y) / 2
	if len(y)%2 == 1 {
		return y[m]
	}
	return (y[m-1] + y[m]) / 2
}
func entropy(counts []int) float64 {
	n := 0
	for _, x := range counts {
		n += x
	}
	if n == 0 {
		return 0
	}
	h := 0.0
	for _, x := range counts {
		if x > 0 {
			p := float64(x) / float64(n)
			h -= p * math.Log2(p)
		}
	}
	return h
}
