package distancecontext

import (
	"math"
	"sort"
	"strings"
)

type tokenProfile struct {
	Right, Left []map[string]int
	Suffix      []map[string]int
}
type profiles map[string]*tokenProfile

func buildProfiles(c corpus, maxD int, bounded bool) profiles {
	p := profiles{}
	get := func(t string) *tokenProfile {
		x := p[t]
		if x == nil {
			x = &tokenProfile{Right: make([]map[string]int, maxD), Left: make([]map[string]int, maxD), Suffix: make([]map[string]int, 2)}
			for i := 0; i < maxD; i++ {
				x.Right[i] = map[string]int{}
				x.Left[i] = map[string]int{}
			}
			x.Suffix[0] = map[string]int{}
			x.Suffix[1] = map[string]int{}
			p[t] = x
		}
		return x
	}
	consume := func(seq []string) {
		for i, t := range seq {
			x := get(t)
			for d := 1; d <= maxD; d++ {
				if i+d < len(seq) {
					x.Right[d-1][seq[i+d]]++
				}
				if i-d >= 0 {
					x.Left[d-1][seq[i-d]]++
				}
			}
			for n := 2; n <= 3; n++ {
				if i+n < len(seq) {
					x.Suffix[n-2][strings.Join(seq[i+1:i+n+1], "\x1f")]++
				}
			}
		}
	}
	if bounded {
		for _, line := range c.Lines {
			consume(line)
		}
	} else {
		consume(c.Tokens)
	}
	return p
}

func total(m map[string]int) int {
	n := 0
	for _, x := range m {
		n += x
	}
	return n
}
func entropyEffective(m map[string]int) float64 {
	n := total(m)
	if n == 0 {
		return 0
	}
	h := 0.
	for _, x := range m {
		q := float64(x) / float64(n)
		h -= q * math.Log(q)
	}
	return math.Exp(h)
}
func rawMetrics(a, b map[string]int) (js, overlap, jac float64) {
	ta, tb := total(a), total(b)
	if ta == 0 || tb == 0 {
		return
	}
	keys := map[string]bool{}
	for k := range a {
		keys[k] = true
	}
	for k := range b {
		keys[k] = true
	}
	inter := 0
	div := 0.
	for k := range keys {
		pa, pb := float64(a[k])/float64(ta), float64(b[k])/float64(tb)
		if a[k] > 0 && b[k] > 0 {
			inter++
		}
		if pa < pb {
			overlap += pa
		} else {
			overlap += pb
		}
		m := (pa + pb) / 2
		if pa > 0 {
			div += .5 * pa * math.Log(pa/m)
		}
		if pb > 0 {
			div += .5 * pb * math.Log(pb/m)
		}
	}
	js = 1 - div/math.Ln2
	if js < 0 {
		js = 0
	} else if js > 1 {
		js = 1
	}
	if overlap < 0 {
		overlap = 0
	} else if overlap > 1 {
		overlap = 1
	}
	jac = float64(inter) / float64(len(keys))
	return
}
func metric(d int, a, b map[string]int, minObs int) Metric {
	js, o, j := rawMetrics(a, b)
	oa, ob := total(a), total(b)
	r := float64(min(oa, ob)) / float64(minObs)
	if r > 1 {
		r = 1
	}
	return Metric{d, js, o, j, oa, ob, entropyEffective(a), entropyEffective(b), r, oa >= minObs && ob >= minObs, 0}
}
func percentile(sorted []float64, v float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	n := sort.Search(len(sorted), func(i int) bool { return sorted[i] > v })
	return 100 * float64(n) / float64(len(sorted))
}
func quantile(x []float64, q float64) float64 {
	if len(x) == 0 {
		return 0
	}
	i := int(math.Ceil(q*float64(len(x)))) - 1
	if i < 0 {
		i = 0
	}
	if i >= len(x) {
		i = len(x) - 1
	}
	return x[i]
}
func avg(x []float64, lo, hi int) float64 {
	if lo < 1 {
		lo = 1
	}
	if hi > len(x) {
		hi = len(x)
	}
	if lo > hi {
		return 0
	}
	s := 0.
	for _, v := range x[lo-1 : hi] {
		s += v
	}
	return s / float64(hi-lo+1)
}
func summary(x []Metric) Summary {
	v := make([]float64, len(x))
	p := make([]float64, len(x))
	for i, m := range x {
		v[i] = m.JSSimilarity
		p[i] = m.BaselinePercentile
	}
	at := func(d int) float64 {
		if d <= len(v) {
			return v[d-1]
		}
		return 0
	}
	return Summary{At1: at(1), At2: at(2), At3: at(3), At5: at(5), At10: at(10), At20: at(20), Mean1To5: avg(v, 1, 5), Mean6To10: avg(v, 6, 10), Mean11To20: avg(v, 11, 20), Persistence1To5: avg(p, 1, 5)}
}
