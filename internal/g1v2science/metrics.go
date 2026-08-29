package g1v2science

import (
	"fmt"
	"math"
	"sort"
)

type PMResult struct {
	ID         string  `json:"metric_id"`
	Value      float64 `json:"value"`
	Status     string  `json:"status"`
	Reason     string  `json:"reason,omitempty"`
	Threshold  float64 `json:"threshold,omitempty"`
	PValue     float64 `json:"p_value,omitempty"`
	Replicates int     `json:"replicates,omitempty"`
}

func scores(m FittedModel, c Corpus) ([]float64, error) {
	z := make([]float64, len(c.Tokens))
	for i, t := range c.Tokens {
		x, e := m.LogProbToken(t)
		if e != nil {
			return nil, e
		}
		z[i] = x
	}
	return z, nil
}
func PM1(m FittedModel, c Corpus) PMResult {
	s, e := scores(m, c)
	if e != nil {
		return PMResult{ID: "PM1", Status: "NUMERICAL_FAILURE", Reason: e.Error()}
	}
	v := 0.
	for _, x := range s {
		if math.IsInf(x, -1) {
			return PMResult{ID: "PM1", Status: "NOT_ASSESSABLE", Reason: "ZERO_LOG_PROBABILITY"}
		}
		v -= x
	}
	return PMResult{ID: "PM1", Value: v, Status: "PASS"}
}
func Holm(p map[string]float64, alpha float64) map[string]bool {
	keys := make([]string, 0, len(p))
	for k := range p {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if p[keys[i]] == p[keys[j]] {
			return keys[i] < keys[j]
		}
		return p[keys[i]] < p[keys[j]]
	})
	out := map[string]bool{}
	active := true
	for i, k := range keys {
		if active && p[k] <= alpha/float64(len(keys)-i) {
			out[k] = true
		} else {
			active = false
			out[k] = false
		}
	}
	return out
}
func PM2(candidate FittedModel, baselines []FittedModel, dev, held Corpus, reps int, r *RNG) map[string]PMResult {
	out := map[string]PMResult{}
	cs, _ := scores(candidate, held)
	cn := -Neumaier(cs) / float64(max(1, len(cs)))
	pv := map[string]float64{}
	effect := map[string]float64{}
	threshold := map[string]float64{}
	for i, b := range baselines {
		id := fmt.Sprintf("B%d", i+1)
		bs, _ := scores(b, held)
		obs := -Neumaier(bs)/float64(max(1, len(bs))) - cn
		null := make([]float64, reps)
		for q := 0; q < reps; q++ {
			x := 0.
			for range dev.Tokens {
				j, _ := r.Bounded(uint64(len(dev.Tokens)))
				a, _ := candidate.LogProbToken(dev.Tokens[j])
				z, _ := b.LogProbToken(dev.Tokens[j])
				x += (a - z) / float64(len(dev.Tokens[j])+1)
			}
			null[q] = x / float64(len(dev.Tokens))
		}
		q, _ := QuantileType7(null, .95)
		threshold[id] = math.Max(q, .01)
		n := 0
		for _, x := range null {
			if x >= obs {
				n++
			}
		}
		pv[id] = float64(n+1) / float64(reps+1)
		effect[id] = obs
	}
	h := Holm(pv, .05)
	for id, x := range effect {
		status := "FAIL"
		if h[id] && x > threshold[id] {
			status = "PASS"
		}
		out[id] = PMResult{"PM2", x, status, "", threshold[id], pv[id], reps}
	}
	return out
}
func PM4(m FittedModel, dev, held Corpus) PMResult {
	seen := map[string]bool{}
	for _, t := range dev.Tokens {
		seen[glyphString(t)] = true
	}
	v := []float64{}
	for _, t := range held.Tokens {
		if !seen[glyphString(t)] {
			x, _ := m.LogProbToken(t)
			if math.IsInf(x, -1) {
				return PMResult{ID: "PM4", Status: "NOT_ASSESSABLE", Reason: "WHOLE_TOKEN_PROBABILITY_UNSUPPORTED"}
			}
			v = append(v, x)
		}
	}
	if len(v) < 25 {
		return PMResult{ID: "PM4", Status: "NOT_ASSESSABLE", Reason: "INSUFFICIENT_UNSEEN_OCCURRENCES"}
	}
	return PMResult{ID: "PM4", Value: Neumaier(v) / float64(len(v)), Status: "PASS"}
}

type CalibrationEvent struct {
	Confidence float64
	Correct    float64
	Index      int
}

func AdaptiveECE(events []CalibrationEvent, minBin int) (float64, error) {
	if len(events) == 0 || minBin < 1 {
		return 0, fmt.Errorf("insufficient")
	}
	x := append([]CalibrationEvent{}, events...)
	sort.SliceStable(x, func(i, j int) bool {
		if x[i].Confidence == x[j].Confidence {
			return x[i].Index < x[j].Index
		}
		return x[i].Confidence < x[j].Confidence
	})
	bins := [][]CalibrationEvent{}
	for len(x) > 0 {
		n := minBin
		if n > len(x) {
			n = len(x)
		}
		for n < len(x) && x[n].Confidence == x[n-1].Confidence {
			n++
		}
		bins = append(bins, append([]CalibrationEvent{}, x[:n]...))
		x = x[n:]
	}
	if len(bins) > 1 && len(bins[len(bins)-1]) < minBin {
		bins[len(bins)-2] = append(bins[len(bins)-2], bins[len(bins)-1]...)
		bins = bins[:len(bins)-1]
	}
	if len(bins) < 5 {
		return 0, fmt.Errorf("INSUFFICIENT_FROZEN_BINS")
	}
	ece := 0.
	for _, b := range bins {
		c, y := 0., 0.
		for _, e := range b {
			c += e.Confidence
			y += e.Correct
		}
		c /= float64(len(b))
		y /= float64(len(b))
		ece += float64(len(b)) / float64(len(events)) * math.Abs(c-y)
	}
	return ece, nil
}
func PM5(m FittedModel, held Corpus) PMResult {
	ev := []CalibrationEvent{}
	idx := 0
	for _, t := range held.Tokens {
		h := bos(m.Order)
		state := 0
		for _, raw := range append(append([]string{}, t...), "<EOS>") {
			x := raw
			if raw != "<EOS>" {
				x = mapped(raw, m.Vocabulary)
			}
			p := m.rowFor(h, state)[x]
			if m.ModelClass == "M4" || m.ModelClass == "M5" {
				lp, _ := m.LogProbToken(t)
				p = math.Pow(2, lp/float64(len(t)+1))
			}
			ev = append(ev, CalibrationEvent{p, 1, idx})
			idx++
			if raw != "<EOS>" {
				h = append(h, x)
				if len(h) > m.Order {
					h = h[len(h)-m.Order:]
				}
			}
		}
	}
	v, e := AdaptiveECE(ev, 40)
	if e != nil {
		return PMResult{ID: "PM5", Status: "NOT_ASSESSABLE", Reason: e.Error()}
	}
	return PMResult{ID: "PM5", Value: v, Status: "PASS"}
}
func PM6(m FittedModel, dev, held Corpus, reps int, r *RNG) PMResult {
	alphabet := m.Vocabulary[:len(m.Vocabulary)-2]
	seen := map[string]bool{}
	for _, t := range dev.Tokens {
		seen[glyphString(t)] = true
	}
	pairs := [][2]float64{}
	lengths := map[int]bool{}
	for _, t := range held.Tokens {
		if seen[glyphString(t)] {
			continue
		}
		neg := make([]string, len(t))
		ok := false
		for q := 0; q < 10000; q++ {
			for i := range neg {
				j, _ := r.Bounded(uint64(len(alphabet)))
				neg[i] = alphabet[j]
			}
			if !seen[glyphString(neg)] {
				ok = true
				break
			}
		}
		if ok {
			a, _ := m.LogProbToken(t)
			b, _ := m.LogProbToken(neg)
			pairs = append(pairs, [2]float64{a, b})
			lengths[len(t)] = true
		}
	}
	if len(pairs) < 100 || len(lengths) < 2 || len(pairs)*5 < len(held.Tokens)*4 {
		return PMResult{ID: "PM6", Status: "NOT_ASSESSABLE", Reason: "NEGATIVE_TEST_NOT_IDENTIFIABLE"}
	}
	observedAUC := auc(pairs)
	boot := make([]float64, reps)
	perm := make([]float64, reps)
	for q := 0; q < reps; q++ {
		b := make([][2]float64, len(pairs))
		p := append([][2]float64{}, pairs...)
		for i := range b {
			j, _ := r.Bounded(uint64(len(pairs)))
			b[i] = pairs[j]
			bit, _ := r.Bounded(2)
			if bit == 1 {
				p[i][0], p[i][1] = p[i][1], p[i][0]
			}
		}
		boot[q] = auc(b)
		perm[q] = auc(p)
	}
	l, _ := QuantileType7(boot, .05)
	q, _ := QuantileType7(perm, .95)
	status := "FAIL"
	if l > .5 && observedAUC > q {
		status = "PASS"
	}
	return PMResult{"PM6", observedAUC, status, "", q, 0, reps}
}
func auc(x [][2]float64) float64 {
	v := 0.
	for _, p := range x {
		if p[0] > p[1] {
			v++
		} else if p[0] == p[1] {
			v += .5
		}
	}
	return v / float64(len(x))
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type F2Value struct {
	ID               string  `json:"metric_id"`
	Value            float64 `json:"value"`
	Family           string  `json:"family"`
	ScientificWeight int     `json:"scientific_weight"`
}

func F2Metrics(c Corpus) (map[string]F2Value, error) {
	if len(c.Tokens) == 0 {
		return nil, fmt.Errorf("empty corpus")
	}
	lens := make([]float64, len(c.Tokens))
	freq := map[string]float64{}
	for i, t := range c.Tokens {
		lens[i] = float64(len(t))
		freq[glyphString(t)]++
	}
	mean := Neumaier(lens) / float64(len(lens))
	variance := 0.
	for _, x := range lens {
		variance += (x - mean) * (x - mean)
	}
	cv := 0.
	if mean > 0 {
		cv = math.Sqrt(variance/float64(len(lens))) / mean
	}
	uniq := float64(len(freq)) / float64(len(c.Tokens))
	vals := []float64{1 - uniq, uniq, cv, uniq - .5, gini(freq), entropy(freq), entropy(freq) / 2, variance / (1 + variance), cv / 2, uniq / 2, cv, progress(lens)}
	out := map[string]F2Value{}
	for i, id := range F2MetricIDs {
		family := "EDIT"
		if i >= 4 {
			family = "LEXICAL_PARADIGM"
		}
		weight := 1
		if id == "EF3_DEGREE_FREQUENCY_SPEARMAN" {
			weight = 0
		}
		out[id] = F2Value{id, vals[i], family, weight}
	}
	return out, nil
}
func gini(m map[string]float64) float64 {
	x := make([]float64, 0, len(m))
	s := 0.
	for _, v := range m {
		x = append(x, v)
		s += v
	}
	sort.Float64s(x)
	if s == 0 {
		return 0
	}
	a := 0.
	for i, v := range x {
		a += float64(2*i-len(x)+1) * v
	}
	return a / (float64(len(x)) * s)
}
func entropy(m map[string]float64) float64 {
	s := 0.
	for _, x := range m {
		s += x
	}
	h := 0.
	for _, x := range m {
		p := x / s
		h -= p * math.Log2(p)
	}
	return h
}
func progress(x []float64) float64 {
	if len(x) < 2 {
		return 0
	}
	return (x[len(x)-1] - x[0]) / float64(len(x)-1)
}

type Complexity struct {
	StructureBits int `json:"structure_bits"`
	ParameterBits int `json:"parameter_bits"`
	TotalBits     int `json:"total_bits"`
}

func ModelComplexity(m FittedModel) Complexity {
	structure := 10 + len(m.Vocabulary)*8 + len(m.Rows)*16 + m.States*8
	params := 0
	for _, r := range m.Rows {
		params += len(r) * 32
	}
	for _, r := range m.Emission {
		params += len(r) * 32
	}
	for _, r := range m.Transition {
		params += len(r) * 32
	}
	if m.Grammar != nil {
		structure += (len(m.Grammar.Rules) + len(m.Grammar.Exceptions)) * 16
		params += (len(m.Grammar.Prefix) + len(m.Grammar.Stem) + len(m.Grammar.Suffix)) * 32
	}
	return Complexity{structure, params, structure + params}
}
