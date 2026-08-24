// Statistical helpers shared by both branches (task82b.txt sec.58-61):
// paired bootstrap over block-level trajectory samples, a permutation-style
// null-comparison p-value/effect-size, Benjamini-Hochberg FDR for the
// preregistered multiple-testing policy, and Spearman correlation for the
// metric-redundancy audit. No new F2 null model is introduced here; these
// operate only on already-computed F2/AX/SX scalar outputs.
package task82b

import (
	"math"
	"math/rand"
	"sort"
)

// BootstrapCI returns a 95% percentile bootstrap confidence interval for
// the mean of samples, resampling with replacement reps times.
func BootstrapCI(samples []float64, seed int64, reps int) (mean, lo, hi float64) {
	n := len(samples)
	if n == 0 {
		return 0, 0, 0
	}
	mean = meanF(samples)
	if n == 1 {
		return mean, mean, mean
	}
	r := rand.New(rand.NewSource(seed))
	means := make([]float64, reps)
	for b := 0; b < reps; b++ {
		sum := 0.0
		for i := 0; i < n; i++ {
			sum += samples[r.Intn(n)]
		}
		means[b] = sum / float64(n)
	}
	sort.Float64s(means)
	lo = percentile(means, 0.025)
	hi = percentile(means, 0.975)
	return mean, lo, hi
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(p * float64(len(sorted)-1))
	return sorted[idx]
}

func meanF(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := 0.0
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

func sdF(xs []float64, mean float64) float64 {
	if len(xs) < 2 {
		return 0
	}
	s := 0.0
	for _, x := range xs {
		d := x - mean
		s += d * d
	}
	return math.Sqrt(s / float64(len(xs)-1))
}

// NullComparison summarizes an observed statistic against a distribution
// of null replicates of the same statistic (task82b.txt sec.58-59): a
// two-sided empirical p-value and a z-like effect size.
type NullComparison struct {
	Observed   float64 `json:"observed"`
	NullMean   float64 `json:"null_mean"`
	NullSD     float64 `json:"null_sd"`
	EffectSize float64 `json:"effect_size"` // (observed - null_mean) / null_sd, 0 if null_sd==0
	PValue     float64 `json:"p_value"`     // two-sided empirical, (#as-extreme + 1) / (n + 1)
	N          int     `json:"n_null_replicates"`
}

// CompareToNull computes NullComparison for one observed value against a
// slice of null replicate values of the same statistic.
func CompareToNull(observed float64, nullReplicates []float64) NullComparison {
	n := len(nullReplicates)
	nm := meanF(nullReplicates)
	nsd := sdF(nullReplicates, nm)
	effect := 0.0
	if nsd > 0 {
		effect = (observed - nm) / nsd
	}
	obsDev := math.Abs(observed - nm)
	extreme := 0
	for _, v := range nullReplicates {
		if math.Abs(v-nm) >= obsDev {
			extreme++
		}
	}
	p := float64(extreme+1) / float64(n+1)
	return NullComparison{Observed: observed, NullMean: nm, NullSD: nsd, EffectSize: effect, PValue: p, N: n}
}

// BenjaminiHochberg returns, for each input p-value, whether it survives
// FDR control at level alpha (task82b.txt sec.60's preregistered
// multiple-testing policy), preserving input order.
func BenjaminiHochberg(pvalues []float64, alpha float64) []bool {
	n := len(pvalues)
	reject := make([]bool, n)
	type idxP struct {
		i int
		p float64
	}
	sorted := make([]idxP, n)
	for i, p := range pvalues {
		sorted[i] = idxP{i, p}
	}
	sort.Slice(sorted, func(a, b int) bool { return sorted[a].p < sorted[b].p })
	maxK := -1
	for k, ip := range sorted {
		threshold := alpha * float64(k+1) / float64(n)
		if ip.p <= threshold {
			maxK = k
		}
	}
	for k := 0; k <= maxK; k++ {
		reject[sorted[k].i] = true
	}
	return reject
}

// SpearmanCorrelation returns the Spearman rank correlation of two
// equal-length samples (task82b.txt sec.61 metric-redundancy audit).
func SpearmanCorrelation(x, y []float64) float64 {
	n := len(x)
	if n == 0 || n != len(y) {
		return 0
	}
	rx := rankOf(x)
	ry := rankOf(y)
	mx, my := meanF(rx), meanF(ry)
	var num, dx2, dy2 float64
	for i := range rx {
		dx := rx[i] - mx
		dy := ry[i] - my
		num += dx * dy
		dx2 += dx * dx
		dy2 += dy * dy
	}
	if dx2 == 0 || dy2 == 0 {
		return 0
	}
	return num / math.Sqrt(dx2*dy2)
}

// rankOf returns the average rank (1-based, ties averaged) of each
// element of xs.
func rankOf(xs []float64) []float64 {
	n := len(xs)
	idx := make([]int, n)
	for i := range idx {
		idx[i] = i
	}
	sort.Slice(idx, func(a, b int) bool { return xs[idx[a]] < xs[idx[b]] })
	ranks := make([]float64, n)
	i := 0
	for i < n {
		j := i
		for j+1 < n && xs[idx[j+1]] == xs[idx[i]] {
			j++
		}
		avgRank := float64(i+j)/2 + 1
		for k := i; k <= j; k++ {
			ranks[idx[k]] = avgRank
		}
		i = j + 1
	}
	return ranks
}
