package numeric

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

func BaselineMapping(n int) []int {
	m := make([]int, n)
	for i := range m {
		m[i] = i
	}
	return m
}

func Values(c Corpus, mapping []int) ([]float64, []int) {
	idx := map[byte]int{}
	for i, g := range c.Alphabet {
		idx[g] = i
	}
	v := make([]float64, len(c.Tokens))
	lengths := make([]int, len(c.Tokens))
	b := float64(len(c.Alphabet))
	for i, t := range c.Tokens {
		for _, g := range t.Glyphs {
			v[i] = v[i]*b + float64(mapping[idx[g]])
		}
		lengths[i] = len(t.Glyphs)
	}
	return v, lengths
}

func Compute(c Corpus, mapping []int) Metrics {
	v, lens := Values(c, mapping)
	var m Metrics
	if len(v) == 0 {
		return m
	}
	b := float64(len(c.Alphabet))
	logv := make([]float64, len(v))
	pos := make([]float64, len(v))
	for i := range v {
		m.LengthMean += float64(lens[i])
		logv[i] = math.Log(v[i]+1) / math.Log(b)
		m.LogMean += logv[i]
		pos[i] = float64(c.Tokens[i].IndexInLine)
	}
	m.LengthMean /= float64(len(v))
	m.LogMean /= float64(len(v))
	m.PositionLengthRho = spearman(pos, intsFloat(lens))
	m.PositionValueRho = spearman(pos, logv)
	var signed, absd, norm []float64
	deltaCounts := map[string]int{}
	ratios := map[string]int{}
	incLines, decLines, eligibleLines, pairs, triples := 0, 0, 0, 0, 0
	run, best, lastDirection := 1, 1, 0
	lagX := [5][]float64{}
	lagY := [5][]float64{}
	lineRanges := ranges(c.Tokens)
	for _, rr := range lineRanges {
		run, lastDirection = 1, 0
		lineInc, lineDec := true, true
		for i := rr[0]; i < rr[1]-1; i++ {
			d := v[i+1] - v[i]
			a := math.Abs(d)
			scale := math.Abs(v[i]) + math.Abs(v[i+1]) + 1
			signed = append(signed, d)
			absd = append(absd, a)
			norm = append(norm, a/scale)
			deltaCounts[fmt.Sprintf("%.3g", d/scale)]++
			pairs++
			m.AdjLengthDiffMean += math.Abs(float64(lens[i+1] - lens[i]))
			if d <= 0 {
				lineInc = false
			}
			if d >= 0 {
				lineDec = false
			}
			direction := 0
			if d > 0 {
				direction = 1
			} else if d < 0 {
				direction = -1
			}
			if direction != 0 && direction == lastDirection {
				run++
			} else if direction != 0 {
				run = 2
			} else {
				run = 1
			}
			lastDirection = direction
			if run > best {
				best = run
			}
			if v[i] > 0 && v[i+1] > 0 {
				ratios[fmt.Sprintf("%.3f", math.Log(v[i+1]/v[i]))]++
			}
		}
		if rr[1]-rr[0] >= 2 {
			eligibleLines++
			if lineInc {
				incLines++
			}
			if lineDec {
				decLines++
			}
		}
		for k := 1; k <= 5; k++ {
			for i := rr[0]; i+k < rr[1]; i++ {
				lagX[k-1] = append(lagX[k-1], v[i])
				lagY[k-1] = append(lagY[k-1], v[i+k])
			}
		}
		for i := rr[0]; i < rr[1]-2; i++ {
			d1 := v[i+1] - v[i]
			d2 := v[i+2] - v[i+1]
			m.APCloseness += math.Exp(-math.Abs(d2-d1) / (math.Abs(d1) + math.Abs(d2) + 1))
			triples++
		}
	}
	if pairs > 0 {
		m.AdjLengthDiffMean /= float64(pairs)
		m.SignedDeltaMean = mean(signed)
		m.AbsDeltaMean = mean(absd)
		m.NormalizedDeltaMean = mean(norm)
		m.RepeatedDeltaFraction = repeated(deltaCounts, pairs)
		m.DeltaEntropy = entropy(deltaCounts, pairs)
	}
	if eligibleLines > 0 {
		m.IncreasingFraction = float64(incLines) / float64(eligibleLines)
		m.DecreasingFraction = float64(decLines) / float64(eligibleLines)
	}
	if triples > 0 {
		m.APCloseness /= float64(triples)
	}
	m.RatioRepeat = repeated(ratios, sumCounts(ratios))
	m.LongestMonotonicRun = float64(best)
	for k := range m.LagRho {
		m.LagRho[k] = spearman(lagX[k], lagY[k])
	}
	zeroGlyph := -1
	for i, d := range mapping {
		if d == 0 {
			zeroGlyph = i
		}
	}
	idx := map[byte]int{}
	for i, g := range c.Alphabet {
		idx[g] = i
	}
	canon := map[string]map[string]bool{}
	admittedTypes := map[string]bool{}
	lead := 0
	for _, t := range c.Tokens {
		admittedTypes[t.Text] = true
		ds := make([]string, len(t.Glyphs))
		for i, g := range t.Glyphs {
			d := mapping[idx[g]]
			ds[i] = strconv.Itoa(d)
		}
		if idx[t.Glyphs[0]] == zeroGlyph {
			lead++
		}
		j := 0
		for j < len(ds)-1 && ds[j] == "0" {
			j++
		}
		key := strings.Join(ds[j:], ",")
		if canon[key] == nil {
			canon[key] = map[string]bool{}
		}
		canon[key][t.Text] = true
	}
	if len(c.Tokens) > 0 {
		m.LeadingZeroFraction = float64(lead) / float64(len(c.Tokens))
	}
	m.LeadingZeroTokenCount = float64(lead)
	colliding := 0
	collisionClasses := 0
	for _, ts := range canon {
		if len(ts) > 1 {
			colliding += len(ts)
			collisionClasses++
		}
	}
	m.CollidingTokenTypeCount = float64(colliding)
	m.CollisionClassCount = float64(collisionClasses)
	if len(admittedTypes) > 0 {
		m.CollisionFraction = float64(colliding) / float64(len(admittedTypes))
	}
	m.EditSubstitutionConsistency = editConsistency(c, mapping)
	eta := etaSquared(logv, c.Tokens, func(t Token) string { return t.Folio })
	m.SequentialComponent = math.Abs(m.LagRho[0])
	m.DifferenceComponent = (m.APCloseness + m.RepeatedDeltaFraction) / 2
	m.DocumentComponent = (math.Abs(m.PositionValueRho) + eta) / 2
	m.Score = (m.SequentialComponent + m.DifferenceComponent + m.DocumentComponent) / 3
	return m
}

// objective computes only the three preregistered score components. Keeping
// diagnostics out of the annealing loop changes no objective and makes the
// identical optimization affordable for every null replicate.
func objective(c Corpus, mapping []int) Metrics {
	v, _ := Values(c, mapping)
	var m Metrics
	if len(v) == 0 {
		return m
	}
	pos, logv := make([]float64, len(v)), make([]float64, len(v))
	b := float64(len(c.Alphabet))
	for i := range v {
		pos[i] = float64(c.Tokens[i].IndexInLine)
		logv[i] = math.Log(v[i]+1) / math.Log(b)
	}
	counts := map[string]int{}
	pairs, triples := 0, 0
	var x, y []float64
	for _, rr := range ranges(c.Tokens) {
		for i := rr[0]; i < rr[1]-1; i++ {
			d := v[i+1] - v[i]
			scale := math.Abs(v[i]) + math.Abs(v[i+1]) + 1
			counts[fmt.Sprintf("%.3g", d/scale)]++
			pairs++
			x = append(x, v[i])
			y = append(y, v[i+1])
		}
		for i := rr[0]; i < rr[1]-2; i++ {
			d1, d2 := v[i+1]-v[i], v[i+2]-v[i+1]
			m.APCloseness += math.Exp(-math.Abs(d2-d1) / (math.Abs(d1) + math.Abs(d2) + 1))
			triples++
		}
	}
	if triples > 0 {
		m.APCloseness /= float64(triples)
	}
	m.RepeatedDeltaFraction = repeated(counts, pairs)
	m.LagRho[0] = spearman(x, y)
	m.PositionValueRho = spearman(pos, logv)
	m.SequentialComponent = math.Abs(m.LagRho[0])
	m.DifferenceComponent = (m.APCloseness + m.RepeatedDeltaFraction) / 2
	m.DocumentComponent = (math.Abs(m.PositionValueRho) + etaSquared(logv, c.Tokens, func(t Token) string { return t.Folio })) / 2
	m.Score = (m.SequentialComponent + m.DifferenceComponent + m.DocumentComponent) / 3
	return m
}

func ranges(t []Token) [][2]int {
	var r [][2]int
	for s := 0; s < len(t); {
		e := s + 1
		for e < len(t) && t[e].Line == t[s].Line {
			e++
		}
		r = append(r, [2]int{s, e})
		s = e
	}
	return r
}
func intsFloat(x []int) []float64 {
	r := make([]float64, len(x))
	for i := range x {
		r[i] = float64(x[i])
	}
	return r
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
func sumCounts(x map[string]int) int {
	s := 0
	for _, v := range x {
		s += v
	}
	return s
}
func repeated(x map[string]int, n int) float64 {
	if n == 0 {
		return 0
	}
	r := 0
	for _, v := range x {
		if v > 1 {
			r += v
		}
	}
	return float64(r) / float64(n)
}
func entropy(x map[string]int, n int) float64 {
	if n == 0 {
		return 0
	}
	h := 0.0
	keys := make([]string, 0, len(x))
	for k := range x {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		p := float64(x[k]) / float64(n)
		h -= p * math.Log2(p)
	}
	return h
}

func spearman(x, y []float64) float64 {
	if len(x) < 2 || len(x) != len(y) {
		return 0
	}
	return pearson(rank(x), rank(y))
}
func rank(x []float64) []float64 {
	idx := make([]int, len(x))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(i, j int) bool { return x[idx[i]] < x[idx[j]] })
	r := make([]float64, len(x))
	for i := 0; i < len(idx); {
		j := i + 1
		for j < len(idx) && x[idx[j]] == x[idx[i]] {
			j++
		}
		a := float64(i+j-1)/2 + 1
		for k := i; k < j; k++ {
			r[idx[k]] = a
		}
		i = j
	}
	return r
}
func pearson(x, y []float64) float64 {
	mx, my := mean(x), mean(y)
	a, b, c := 0.0, 0.0, 0.0
	for i := range x {
		dx, dy := x[i]-mx, y[i]-my
		a += dx * dy
		b += dx * dx
		c += dy * dy
	}
	if b == 0 || c == 0 {
		return 0
	}
	return a / math.Sqrt(b*c)
}

func etaSquared(v []float64, t []Token, key func(Token) string) float64 {
	if len(v) < 2 {
		return 0
	}
	mu := mean(v)
	total := 0.0
	groups := map[string][]float64{}
	for i, x := range v {
		d := x - mu
		total += d * d
		groups[key(t[i])] = append(groups[key(t[i])], x)
	}
	if total == 0 {
		return 0
	}
	between := 0.0
	for _, g := range groups {
		d := mean(g) - mu
		between += float64(len(g)) * d * d
	}
	return between / total
}

// The exact positional substitution identity is a representation-induced
// diagnostic, not independent evidence. We report its corpus-level rate.
func editConsistency(c Corpus, _ []int) float64 {
	types := map[string]bool{}
	for _, t := range c.Tokens {
		types[t.Text] = true
	}
	for a := range types {
		for p := range a {
			for _, g := range c.Alphabet {
				if g != a[p] && types[a[:p]+string(g)+a[p+1:]] {
					return 1
				}
			}
		}
	}
	return 0
}
