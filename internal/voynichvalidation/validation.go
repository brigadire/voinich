// Package voynichvalidation contains the post-search, non-optimising Task54b
// measurements. It deliberately keeps raw metrics separate from search score.
package voynichvalidation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strings"

	"zcore.dev/voinich/internal/inversetransposition"
)

type Metrics = inversetransposition.Metrics

func Read(path string) ([]string, []int, []byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, nil, err
	}
	lines := strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n")
	var tokens []string
	lengths := make([]int, 0, len(lines))
	for _, line := range lines {
		f := strings.Fields(line)
		if len(f) > 0 {
			tokens = append(tokens, f...)
			lengths = append(lengths, len(f))
		}
	}
	return tokens, lengths, b, nil
}

func SHA256(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

func Delta(a, b Metrics) Metrics {
	return Metrics{TransitionConcentration: a.TransitionConcentration - b.TransitionConcentration, RelationSignificance: a.RelationSignificance - b.RelationSignificance, SequenceRepetition: a.SequenceRepetition - b.SequenceRepetition, HigherOrderRepetition: a.HigherOrderRepetition - b.HigherOrderRepetition}
}
func Relative(d, base Metrics) Metrics {
	return Metrics{TransitionConcentration: relative(d.TransitionConcentration, base.TransitionConcentration), RelationSignificance: relative(d.RelationSignificance, base.RelationSignificance), SequenceRepetition: relative(d.SequenceRepetition, base.SequenceRepetition), HigherOrderRepetition: relative(d.HigherOrderRepetition, base.HigherOrderRepetition)}
}
func relative(x, d float64) float64 {
	if d == 0 || math.IsNaN(d) {
		return math.NaN()
	}
	return x / d
}

// SplitByLines is deterministic and preserves every existing logical line.
// Task54 did not preregister a split, so callers must document this limitation.
func SplitByLines(tokens []string, lengths []int) ([]string, []string) {
	cut := (len(lengths)*4 + 2) / 5
	if cut < 1 {
		cut = 1
	}
	if cut >= len(lengths) {
		cut = len(lengths) - 1
	}
	flatten := func(from, to int) []string {
		var out []string
		pos := 0
		for i, n := range lengths {
			if i >= from && i < to {
				out = append(out, tokens[pos:pos+n]...)
			}
			pos += n
		}
		return out
	}
	return flatten(0, cut), flatten(cut, len(lengths))
}

type CandidateRow struct {
	ID                   string
	Width                int
	Order                string
	Score                float64
	Raw, Delta, Relative Metrics
}

func CandidateRows(original Metrics, rows []CandidateRow) {
	for i := range rows {
		rows[i].Delta = Delta(rows[i].Raw, original)
		rows[i].Relative = Relative(rows[i].Delta, original)
	}
}

// FixedCalibrationScore uses only the frozen Doyle/T2/T4/T8 ranges from
// INVERSE_TRANSPOSITION_TASK54_1_REPORT.md. It is a post-hoc effect scale, not
// the search objective. Values outside the control range are intentionally not
// clipped.
func FixedCalibrationScore(d Metrics) float64 {
	lo := [4]float64{0.672405, 0.002440, 0.357911, 0.023976}
	hi := [4]float64{0.700458, 0.013249, 0.520635, 0.139530}
	v := [4]float64{d.TransitionConcentration, d.RelationSignificance, d.SequenceRepetition, d.HigherOrderRepetition}
	var s float64
	for i := range v {
		s += (v[i] - lo[i]) / (hi[i] - lo[i])
	}
	return s / 4
}

type NullRow struct {
	Replicate int
	Metrics   Metrics
	Delta     Metrics
	Score     float64
}

func NullDistribution(tokens []string, n int, seed int64, original Metrics) []NullRow {
	r := rand.New(rand.NewSource(seed))
	out := make([]NullRow, 0, n)
	for i := 0; i < n; i++ {
		p := r.Perm(len(tokens))
		shuffled := make([]string, len(tokens))
		for j, k := range p {
			shuffled[j] = tokens[k]
		}
		m := inversetransposition.Measure(shuffled)
		d := Delta(m, original)
		out = append(out, NullRow{i + 1, m, d, FixedCalibrationScore(d)})
	}
	return out
}

func Quantile(xs []float64, q float64) float64 {
	if len(xs) == 0 {
		return math.NaN()
	}
	sort.Float64s(xs)
	if q <= 0 {
		return xs[0]
	}
	if q >= 1 {
		return xs[len(xs)-1]
	}
	x := q * float64(len(xs)-1)
	lo := int(x)
	hi := lo + 1
	if hi >= len(xs) {
		return xs[lo]
	}
	return xs[lo] + (xs[hi]-xs[lo])*(x-float64(lo))
}
func Percentile(nulls []NullRow, d Metrics) Metrics {
	vals := [4][]float64{{}, {}, {}, {}}
	for _, n := range nulls {
		x := []float64{n.Delta.TransitionConcentration, n.Delta.RelationSignificance, n.Delta.SequenceRepetition, n.Delta.HigherOrderRepetition}
		for i := range x {
			vals[i] = append(vals[i], x[i])
		}
	}
	v := []float64{d.TransitionConcentration, d.RelationSignificance, d.SequenceRepetition, d.HigherOrderRepetition}
	var p Metrics
	for i := range v {
		count := 0
		for _, x := range vals[i] {
			if x <= v[i] {
				count++
			}
		}
		z := float64(count) / float64(len(vals[i]))
		switch i {
		case 0:
			p.TransitionConcentration = z
		case 1:
			p.RelationSignificance = z
		case 2:
			p.SequenceRepetition = z
		case 3:
			p.HigherOrderRepetition = z
		}
	}
	return p
}

func FormatMetric(m Metrics) string {
	return fmt.Sprintf("%.12g\t%.12g\t%.12g\t%.12g", m.TransitionConcentration, m.RelationSignificance, m.SequenceRepetition, m.HigherOrderRepetition)
}
