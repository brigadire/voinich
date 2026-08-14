package higherorderseq

import (
	"math"
	"sort"
)

// bh applies Benjamini-Hochberg FDR correction, matching the implementation
// already used by replicated-local-structure-audit and token-relation-validate.
func bh(p []float64) []float64 {
	idx := make([]int, len(p))
	for i := range idx {
		idx[i] = i
	}
	sort.Slice(idx, func(i, j int) bool { return p[idx[i]] < p[idx[j]] })
	q := make([]float64, len(p))
	next := 1.
	for rank := len(idx); rank >= 1; rank-- {
		i := idx[rank-1]
		v := p[i] * float64(len(idx)) / float64(rank)
		if v > next {
			v = next
		}
		if v > 1 {
			v = 1
		}
		q[i] = v
		next = v
	}
	return q
}

func meanSD(x []float64) (float64, float64) {
	if len(x) == 0 {
		return 0, 0
	}
	m := 0.
	for _, v := range x {
		m += v
	}
	m /= float64(len(x))
	s := 0.
	for _, v := range x {
		s += (v - m) * (v - m)
	}
	return m, math.Sqrt(s / float64(len(x)))
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

func minMax(x []float64) (float64, float64) {
	if len(x) == 0 {
		return 0, 0
	}
	lo, hi := x[0], x[0]
	for _, v := range x[1:] {
		if v < lo {
			lo = v
		}
		if v > hi {
			hi = v
		}
	}
	return lo, hi
}

// empiricalP is the standard exceedance-based permutation p-value formula
// task22 section 69 specifies: (exceedances + 1) / (N + 1).
func empiricalP(observed float64, null []float64) float64 {
	exceed := 0
	for _, v := range null {
		if v >= observed {
			exceed++
		}
	}
	return float64(exceed+1) / float64(len(null)+1)
}

// entropyBits is the Shannon entropy of a probability distribution, base 2.
// Keys are visited in sorted order rather than Go's randomized map iteration
// order so repeated runs sum the same float64 terms in the same sequence.
func entropyBits(p map[string]float64) float64 {
	h := 0.0
	for _, k := range stringFloatKeys(p) {
		if v := p[k]; v > 0 {
			h -= v * math.Log2(v)
		}
	}
	return h
}

func stringFloatKeys(m map[string]float64) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func unionKeysSorted(p, q map[string]float64) []string {
	keys := map[string]bool{}
	for k := range p {
		keys[k] = true
	}
	for k := range q {
		keys[k] = true
	}
	return stringKeys(keys)
}

// jsDivergenceBits is the Jensen-Shannon divergence in bits (0..1) between
// two distributions given as count maps over the same key space. Keys are
// visited in sorted order so repeated runs sum the same float64 terms in the
// same sequence rather than in Go's randomized map iteration order.
func jsDivergenceBits(p, q map[string]float64) float64 {
	keys := unionKeysSorted(p, q)
	m := map[string]float64{}
	for _, k := range keys {
		m[k] = (p[k] + q[k]) / 2
	}
	kl := func(a, b map[string]float64) float64 {
		s := 0.0
		for _, k := range keys {
			if a[k] > 0 && b[k] > 0 {
				s += a[k] * math.Log2(a[k]/b[k])
			}
		}
		return s
	}
	return kl(p, m)/2 + kl(q, m)/2
}

// totalVariation is the total variation distance between two distributions
// given as probability maps over the same key space.
func totalVariation(p, q map[string]float64) float64 {
	sum := 0.0
	for _, k := range unionKeysSorted(p, q) {
		d := p[k] - q[k]
		if d < 0 {
			d = -d
		}
		sum += d
	}
	return sum / 2
}

func toProbabilities(counts map[string]int) map[string]float64 {
	total := 0
	for _, n := range counts {
		total += n
	}
	p := make(map[string]float64, len(counts))
	if total == 0 {
		return p
	}
	for k, n := range counts {
		p[k] = float64(n) / float64(total)
	}
	return p
}

// smoothedProb applies fixed additive smoothing (task22 section 26: alpha =
// 0.5, never optimized) over a vocabulary that is extended with x itself if
// x was never observed in counts/vocab (a standard, deterministic OOV fix so
// the held-out log loss is always finite).
func smoothedProb(counts map[string]int, vocab map[string]bool, x string, total int, alpha float64) float64 {
	v := len(vocab)
	if !vocab[x] {
		v++
	}
	return (float64(counts[x]) + alpha) / (float64(total) + alpha*float64(v))
}

func log2Loss(p float64) float64 {
	if p <= 0 {
		return math.Inf(1)
	}
	return -math.Log2(p)
}

func stringKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
