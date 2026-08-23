package fingerprintv2

import (
	"math"
	"sort"
)

func orderedKeys[T any](m map[string]T) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func mean(x []float64) float64 {
	if len(x) == 0 {
		return 0
	}
	s := 0.0
	for _, v := range x {
		s += v
	}
	return s / float64(len(x))
}

func sd(x []float64, m float64) float64 {
	if len(x) < 2 {
		return 0
	}
	s := 0.0
	for _, v := range x {
		d := v - m
		s += d * d
	}
	return math.Sqrt(s / float64(len(x)-1))
}

func nullTest(id, model string, observed float64, null []float64) NullTest {
	m := mean(null)
	s := sd(null, m)
	ge := 0
	for _, v := range null {
		if v >= observed {
			ge++
		}
	}
	effect := 0.0
	effectDefined := false
	if s > 0 {
		effect = (observed - m) / s
		effectDefined = true
	}
	return NullTest{
		ID: id, NullModel: model, Observed: observed, NullMean: m, NullSD: s,
		EffectSize: effect, EffectDefined: effectDefined, PValue: float64(ge+1) / float64(len(null)+1),
		Replicates: len(null), Alternative: "greater",
	}
}

// fdr applies Benjamini-Hochberg correction to the supplied independent
// reporting family. The tests are sorted afterwards for stable output.
func fdr(tests []NullTest) []NullTest {
	if len(tests) == 0 {
		return tests
	}
	order := make([]int, len(tests))
	for i := range tests {
		order[i] = i
	}
	sort.Slice(order, func(i, j int) bool {
		a, b := tests[order[i]], tests[order[j]]
		if a.PValue != b.PValue {
			return a.PValue < b.PValue
		}
		return a.ID < b.ID
	})
	q := make([]float64, len(tests))
	last := 1.0
	for i := len(order) - 1; i >= 0; i-- {
		v := tests[order[i]].PValue * float64(len(tests)) / float64(i+1)
		if v > last {
			v = last
		}
		if v > 1 {
			v = 1
		}
		last = v
		q[order[i]] = v
	}
	for i := range tests {
		tests[i].QValue = q[i]
	}
	stableTests(tests)
	return tests
}

func gini(values []int) float64 {
	if len(values) == 0 {
		return 0
	}
	v := append([]int(nil), values...)
	sort.Ints(v)
	total := 0
	weighted := 0
	for i, x := range v {
		total += x
		weighted += (i + 1) * x
	}
	if total == 0 {
		return 0
	}
	n := len(v)
	return (2*float64(weighted))/float64(n*total) - float64(n+1)/float64(n)
}

func totalVariation(a, b map[string]int) float64 {
	na, nb := 0, 0
	for _, k := range orderedKeys(a) {
		na += a[k]
	}
	for _, k := range orderedKeys(b) {
		nb += b[k]
	}
	if na == 0 && nb == 0 {
		return 0
	}
	keys := map[string]bool{}
	for k := range a {
		keys[k] = true
	}
	for k := range b {
		keys[k] = true
	}
	s := 0.0
	for _, k := range orderedKeys(keys) {
		pa, pb := 0.0, 0.0
		if na > 0 {
			pa = float64(a[k]) / float64(na)
		}
		if nb > 0 {
			pb = float64(b[k]) / float64(nb)
		}
		s += math.Abs(pa - pb)
	}
	return s / 2
}

func entropy(counts map[string]int) float64 {
	total := 0
	for _, k := range orderedKeys(counts) {
		total += counts[k]
	}
	if total == 0 {
		return 0
	}
	h := 0.0
	for _, k := range orderedKeys(counts) {
		if counts[k] == 0 {
			continue
		}
		p := float64(counts[k]) / float64(total)
		h -= p * math.Log2(p)
	}
	return h
}

func normalizedMI(a, b []string) float64 {
	if len(a) == 0 || len(a) != len(b) {
		return 0
	}
	ac, bc, joint := map[string]int{}, map[string]int{}, map[string]int{}
	for i := range a {
		ac[a[i]]++
		bc[b[i]]++
		joint[a[i]+"\x00"+b[i]]++
	}
	ha, hb := entropy(ac), entropy(bc)
	if ha+hb == 0 {
		return 0
	}
	mi := ha + hb - entropy(joint)
	return 2 * mi / (ha + hb)
}

func ranks(values []float64, names []string) []float64 {
	idx := make([]int, len(values))
	for i := range idx {
		idx[i] = i
	}
	sort.Slice(idx, func(i, j int) bool {
		a, b := values[idx[i]], values[idx[j]]
		if a != b {
			return a < b
		}
		return names[idx[i]] < names[idx[j]]
	})
	out := make([]float64, len(values))
	for start := 0; start < len(idx); {
		end := start + 1
		for end < len(idx) && values[idx[end]] == values[idx[start]] {
			end++
		}
		rank := float64(start+1+end) / 2
		for _, i := range idx[start:end] {
			out[i] = rank
		}
		start = end
	}
	return out
}

func spearman(x, y []float64, names []string) float64 {
	if len(x) < 2 || len(x) != len(y) {
		return 0
	}
	rx, ry := ranks(x, names), ranks(y, names)
	mx, my := mean(rx), mean(ry)
	num, dx, dy := 0.0, 0.0, 0.0
	for i := range rx {
		a, b := rx[i]-mx, ry[i]-my
		num += a * b
		dx += a * a
		dy += b * b
	}
	if dx == 0 || dy == 0 {
		return 0
	}
	return num / math.Sqrt(dx*dy)
}
