package inversehomophony

import (
	"math"
	"sort"

	"zcore.dev/voinich/internal/sequenceanalyze"
	"zcore.dev/voinich/internal/vocabularygrowth"
)

// StructuralMetrics is one corpus state's scalar values for the task57
// section 12 primary structural families. Vocabulary and Sequence reuse
// the production vocabularygrowth/sequenceanalyze packages directly (pure
// functions over token slices); Transition is a self-contained lightweight
// analogue of Stage 27's backbone/significant-relation estimand -
// INVERSE_HOMOPHONY_DESIGN.md section 0/7.3 explains why the full Stage 27
// package is reserved for the one-time Voynich comparison instead of this
// repeated-many-times validation loop.
type StructuralMetrics struct {
	VocabSize                 int
	HapaxFraction             float64
	DisFraction               float64
	HeapsBeta                 float64
	RepeatedBigramFraction    float64
	SignificantBigramFraction float64
}

// ComputeStructural computes every family for one corpus state.
func ComputeStructural(tokens []string, lines [][]string) StructuralMetrics {
	vg, _ := vocabularygrowth.Analyze(tokens, vocabularygrowth.Parameters{NullPermutations: 0, Seed: 1})
	seq, _ := sequenceanalyze.AnalyzeLines(lines, sequenceanalyze.DefaultParameters())

	m := StructuralMetrics{
		VocabSize: vg.Final.Vocabulary,
		HeapsBeta: vg.Fit.Beta,
	}
	if vg.Final.Vocabulary > 0 {
		m.HapaxFraction = float64(vg.Final.Hapax) / float64(vg.Final.Vocabulary)
		m.DisFraction = float64(vg.Final.Dis) / float64(vg.Final.Vocabulary)
	}
	for _, s := range seq.NGramSummary {
		if s.N == 2 {
			if s.Unique > 0 {
				m.RepeatedBigramFraction = float64(s.Repeated) / float64(s.Unique)
			}
			break
		}
	}
	m.SignificantBigramFraction = significantBigramFraction(lines)
	return m
}

// significantBigramFraction is the self-contained transition-family
// estimand (see type doc): among distinct within-line bigrams observed at
// least twice, the fraction whose co-occurrence significantly exceeds the
// independence expectation freq(a)*freq(b)/N, by a G-test against a fixed
// chi-square critical value at alpha=0.01 (df=1, critical ~6.635). No
// multiple-comparison correction beyond this one fixed alpha - deliberately
// simple and frozen before any corpus is scored.
func significantBigramFraction(lines [][]string) float64 {
	const chiSqCritical = 6.635
	const minCount = 2

	freq := make(map[string]int)
	bigram := make(map[[2]string]int)
	total := 0
	for _, line := range lines {
		for i, t := range line {
			freq[t]++
			total++
			if i+1 < len(line) {
				bigram[[2]string{t, line[i+1]}]++
			}
		}
	}
	if total == 0 || len(bigram) == 0 {
		return 0
	}
	n := float64(total)
	keys := make([][2]string, 0, len(bigram))
	for k := range bigram {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i][0] != keys[j][0] {
			return keys[i][0] < keys[j][0]
		}
		return keys[i][1] < keys[j][1]
	})

	considered, significant := 0, 0
	for _, k := range keys {
		o := float64(bigram[k])
		if o < minCount {
			continue
		}
		e := float64(freq[k[0]]) * float64(freq[k[1]]) / n
		if e <= 0 {
			continue
		}
		g := 2 * o * math.Log(o/e)
		considered++
		if g > chiSqCritical {
			significant++
		}
	}
	if considered == 0 {
		return 0
	}
	return float64(significant) / float64(considered)
}

// StructuralComparison is task57 section 11's before/after/recovered
// triple for one metric.
type StructuralComparison struct {
	Metric           string
	Plaintext        float64
	Ciphertext       float64
	Recovered        float64
	DeltaCipher      float64
	DeltaRecover     float64
	RecoveryFraction float64 // NaN when |DeltaCipher| ~ 0 (formula not meaningful, task57 section 11)
}

// CompareStructural builds every StructuralComparison for the six scalar
// metrics tracked in StructuralMetrics.
func CompareStructural(p, h, r StructuralMetrics) []StructuralComparison {
	metrics := []struct {
		name       string
		pv, hv, rv float64
	}{
		{"vocab_size", float64(p.VocabSize), float64(h.VocabSize), float64(r.VocabSize)},
		{"hapax_fraction", p.HapaxFraction, h.HapaxFraction, r.HapaxFraction},
		{"dis_fraction", p.DisFraction, h.DisFraction, r.DisFraction},
		{"heaps_beta", p.HeapsBeta, h.HeapsBeta, r.HeapsBeta},
		{"repeated_bigram_fraction", p.RepeatedBigramFraction, h.RepeatedBigramFraction, r.RepeatedBigramFraction},
		{"significant_bigram_fraction", p.SignificantBigramFraction, h.SignificantBigramFraction, r.SignificantBigramFraction},
	}
	out := make([]StructuralComparison, 0, len(metrics))
	for _, m := range metrics {
		dc := m.hv - m.pv
		dr := m.rv - m.pv
		frac := math.NaN()
		if math.Abs(dc) > 1e-9 {
			frac = 1 - math.Abs(dr)/math.Abs(dc)
		}
		out = append(out, StructuralComparison{
			Metric: m.name, Plaintext: m.pv, Ciphertext: m.hv, Recovered: m.rv,
			DeltaCipher: dc, DeltaRecover: dr, RecoveryFraction: frac,
		})
	}
	return out
}
