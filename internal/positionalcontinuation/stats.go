package positionalcontinuation

import (
	"math"
	"sort"
)

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

// quantile is a simple linear-interpolation quantile over a copy of x.
func quantile(x []float64, q float64) float64 {
	if len(x) == 0 {
		return 0
	}
	y := append([]float64(nil), x...)
	sort.Float64s(y)
	if len(y) == 1 {
		return y[0]
	}
	pos := q * float64(len(y)-1)
	lo := int(math.Floor(pos))
	hi := int(math.Ceil(pos))
	if lo == hi {
		return y[lo]
	}
	frac := pos - float64(lo)
	return y[lo]*(1-frac) + y[hi]*frac
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

// empiricalP is the standard exceedance-based permutation p-value formula:
// (exceedances + 1) / (N + 1).
func empiricalP(observed float64, null []float64) float64 {
	exceed := 0
	for _, v := range null {
		if v >= observed {
			exceed++
		}
	}
	return float64(exceed+1) / float64(len(null)+1)
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

// entropyBits is the Shannon entropy of a probability distribution, base 2.
// Keys are visited in sorted order rather than Go's randomized map iteration
// order so repeated runs sum the same float64 terms in the same sequence.
func entropyBits(p map[string]float64) float64 {
	h := 0.0
	for _, k := range stringKeysFloat(p) {
		if v := p[k]; v > 0 {
			h -= v * math.Log2(v)
		}
	}
	return h
}

func countEntropyBits(counts map[string]int) float64 {
	return entropyBits(toProbabilities(counts))
}

// smoothedProb applies fixed additive smoothing (alpha = 0.5, never
// optimized) over a vocabulary that is extended with x itself if x was never
// observed in counts/vocab, so held-out log loss is always finite.
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

// mutualInformationBits is the plug-in estimate I(X;Y) in bits from paired
// (x,y) observations.
func mutualInformationBits(xs, ys []string) float64 {
	n := len(xs)
	if n == 0 {
		return 0
	}
	joint := map[[2]string]int{}
	mx := map[string]int{}
	my := map[string]int{}
	for i := range xs {
		joint[[2]string{xs[i], ys[i]}]++
		mx[xs[i]]++
		my[ys[i]]++
	}
	pairs := make([][2]string, 0, len(joint))
	for k := range joint {
		pairs = append(pairs, k)
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i][0] != pairs[j][0] {
			return pairs[i][0] < pairs[j][0]
		}
		return pairs[i][1] < pairs[j][1]
	})
	nf := float64(n)
	mi := 0.0
	for _, pr := range pairs {
		pxy := float64(joint[pr]) / nf
		px := float64(mx[pr[0]]) / nf
		py := float64(my[pr[1]]) / nf
		if pxy > 0 && px > 0 && py > 0 {
			mi += pxy * math.Log2(pxy/(px*py))
		}
	}
	return mi
}

func normalizedPosition(index, blockLen int) float64 {
	if blockLen <= 1 {
		return 0
	}
	return float64(index) / float64(blockLen-1)
}

// blockBinFixed maps a normalized position in [0,1] to one of the ten fixed
// [0,0.1)...[0.9,1.0] bins (task23 section 13).
func blockBinFixed(p float64) string {
	buckets := []string{"B0", "B1", "B2", "B3", "B4", "B5", "B6", "B7", "B8", "B9"}
	idx := int(p * 10)
	if idx >= 10 {
		idx = 9
	}
	if idx < 0 {
		idx = 0
	}
	return buckets[idx]
}

// blockBinCoarse maps a normalized position to BLOCK_START/MIDDLE/END
// (task23 section 14).
func blockBinCoarse(p float64) string {
	switch {
	case p < 0.2:
		return "BLOCK_START"
	case p < 0.8:
		return "BLOCK_MIDDLE"
	default:
		return "BLOCK_END"
	}
}

// lineCategory implements task23 sections 8-11: LINE_START/LINE_END take
// priority over the continuous EARLY/MIDDLE/LATE thresholds.
func lineCategory(contextStartsLine, xEndsLine bool, normPos float64) string {
	switch {
	case contextStartsLine:
		return "LINE_START"
	case xEndsLine:
		return "LINE_END"
	case normPos < 0.25:
		return "LINE_EARLY"
	case normPos <= 0.75:
		return "LINE_MIDDLE"
	default:
		return "LINE_LATE"
	}
}

func mean(x []float64) float64 {
	m, _ := meanSD(x)
	return m
}
