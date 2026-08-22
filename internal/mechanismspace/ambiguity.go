package mechanismspace

import (
	"math/rand"
	"sort"

	"zcore.dev/voinich/internal/evaglyph"
)

// PreimageMultiplicity is task66 section 29's descriptive ambiguity
// measure for LOSSY/AMBIGUOUS mechanisms: for a sample of observed output
// tokens, how many distinct vocabulary words could have produced that
// exact output form under the same mechanism/state. It is purely
// descriptive (never used to attempt decryption).
func PreimageMultiplicity(cfg Config, vocabulary [][]string, outputs [][]string, sampleN int, seed int64) float64 {
	if len(outputs) == 0 || len(vocabulary) == 0 {
		return 0
	}
	r := rand.New(rand.NewSource(seed))
	idx := r.Perm(len(outputs))
	if sampleN > len(idx) {
		sampleN = len(idx)
	}
	total := 0
	for _, i := range idx[:sampleN] {
		target := joinTokens(outputs[i])
		count := 0
		for _, v := range vocabulary {
			candidate := form(v, 0, 0, cfg.Grammar, cfg.Seed)
			if cfg.Grammar == NoGrammar {
				candidate = mapGlyphs(v, 0, 0, cfg.Seed)
			}
			if joinTokens(candidate) == target {
				count++
			}
		}
		if count == 0 {
			count = 1
		}
		total += count
	}
	return float64(total) / float64(sampleN)
}

// InformationRetention is task66 section 64's coarse input/output mutual
// information test: it buckets each input unit's length and first glyph
// and each aligned output token's length and first glyph, then reports
// the mutual information between the two coarse feature streams
// (evaglyph.MI, the same estimator used by Task58/59).
func InformationRetention(inputWords [][]string, outputTokens [][]string) float64 {
	n := len(inputWords)
	if len(outputTokens) < n {
		n = len(outputTokens)
	}
	if n == 0 {
		return 0
	}
	a := make([]string, n)
	b := make([]string, n)
	for i := 0; i < n; i++ {
		a[i] = coarseFeature(inputWords[i])
		b[i] = coarseFeature(outputTokens[i])
	}
	return evaglyph.MI(a, b)
}

func coarseFeature(t []string) string {
	if len(t) == 0 {
		return "0:"
	}
	bucket := len(t)
	if bucket > 6 {
		bucket = 6
	}
	return string(rune('0'+bucket)) + ":" + t[0]
}

// SensitivityClass is task66 section 66's four-way classification.
type SensitivityClass string

const (
	StrongInputDependence  SensitivityClass = "STRONG_INPUT_DEPENDENCE"
	PartialInputDependence SensitivityClass = "PARTIAL_INPUT_DEPENDENCE"
	WeakInputDependence    SensitivityClass = "WEAK_INPUT_DEPENDENCE"
	InputIndependent       SensitivityClass = "INPUT_INDEPENDENT"
)

// ClassifySensitivity compares the fingerprint of a mechanism's real-input
// run against its shuffled-plaintext-ablation run (task66 sections 65-66):
// the larger the fraction of families that moved materially, the stronger
// the input dependence.
func ClassifySensitivity(real, shuffled Fingerprint) SensitivityClass {
	deltas := deltasBetween(real, shuffled)
	sort.Float64s(deltas)
	moved := 0
	for _, d := range deltas {
		if d > 0.05 {
			moved++
		}
	}
	frac := float64(moved) / float64(max1(len(deltas)))
	switch {
	case frac >= 0.75:
		return StrongInputDependence
	case frac >= 0.4:
		return PartialInputDependence
	case frac > 0:
		return WeakInputDependence
	default:
		return InputIndependent
	}
}
