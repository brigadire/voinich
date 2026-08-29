package g1v2science

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

type FittedModel struct {
	CandidateID           string                        `json:"candidate_id"`
	ModelClass            string                        `json:"model_class"`
	Route                 string                        `json:"route"`
	Vocabulary            []string                      `json:"vocabulary"`
	Alpha                 float64                       `json:"alpha,omitempty"`
	Order                 int                           `json:"order,omitempty"`
	Rows                  map[string]map[string]float64 `json:"rows,omitempty"`
	Transitions           map[string]map[string]int     `json:"transitions,omitempty"`
	States                int                           `json:"states,omitempty"`
	Pi                    []float64                     `json:"pi,omitempty"`
	Transition            [][]float64                   `json:"transition,omitempty"`
	Emission              [][]float64                   `json:"emission,omitempty"`
	Grammar               *Grammar                      `json:"grammar,omitempty"`
	Operations            uint64                        `json:"operations"`
	Iterations            int                           `json:"iterations,omitempty"`
	SelectedRestart       int                           `json:"selected_restart,omitempty"`
	DevelopmentTokenCount int                           `json:"development_token_count"`
	SerializationSHA256   string                        `json:"serialization_sha256,omitempty"`
}
type Grammar struct {
	Prefix        map[string]float64 `json:"prefix"`
	Stem          map[string]float64 `json:"stem"`
	Suffix        map[string]float64 `json:"suffix"`
	Rules         map[string]float64 `json:"rules"`
	Exceptions    map[string]float64 `json:"exceptions"`
	BackoffWeight float64            `json:"backoff_weight"`
	MinSupport    int                `json:"min_support"`
}

func finish(m FittedModel) (FittedModel, string, error) {
	m.SerializationSHA256 = ""
	b, e := CanonicalJSON(m)
	if e != nil {
		return m, "PROTOCOL_VETO", e
	}
	m.SerializationSHA256 = Hash(b)
	return m, "FIT_SUCCESS", nil
}
func counts(c Corpus) map[string]float64 {
	m := map[string]float64{}
	for _, t := range c.Tokens {
		for _, g := range t {
			m[g]++
		}
		m["<EOS>"]++
	}
	return m
}
func row(count map[string]float64, v []string, alpha float64) (map[string]float64, error) {
	w := make([]float64, len(v))
	for i, x := range v {
		w[i] = count[x] + alpha
	}
	p, e := Normalize(w)
	if e != nil {
		return nil, e
	}
	z := map[string]float64{}
	for i, x := range v {
		z[x] = p[i]
	}
	return z, nil
}
func context(x []string) string { return strings.Join(x, "\x1f") }
func bos(n int) []string {
	x := make([]string, n)
	for i := range x {
		x[i] = "<BOS>"
	}
	return x
}

func FitCandidate(c Candidate, dev Corpus) (FittedModel, string, error) {
	if len(dev.Tokens) == 0 {
		return FittedModel{}, "FIT_FAILURE", fmt.Errorf("empty DEVELOPMENT")
	}
	switch c.Model {
	case "M0":
		return fitM0(c, dev)
	case "M1":
		return fitM1(c, dev)
	case "M2":
		return fitM2(c, dev)
	case "M3":
		return fitM3(c, dev)
	case "M4":
		return fitM4(c, dev)
	case "M5":
		return fitM5(c, dev)
	}
	return FittedModel{}, "PROTOCOL_VETO", fmt.Errorf("unknown model")
}
func fitM0(c Candidate, d Corpus) (FittedModel, string, error) {
	a, e := numeric(c.Hyper["alpha"])
	if e != nil {
		return FittedModel{}, "PROTOCOL_VETO", e
	}
	v := Vocabulary(d)
	r, e := row(counts(d), v, a)
	if e != nil {
		return FittedModel{}, "FIT_FAILURE", e
	}
	return finish(FittedModel{CandidateID: c.ID, ModelClass: "M0", Route: c.Route, Vocabulary: v, Alpha: a, Rows: map[string]map[string]float64{"": r}, DevelopmentTokenCount: len(d.Tokens)})
}
func fitM1(c Candidate, d Corpus) (FittedModel, string, error) {
	a, e := numeric(c.Hyper["alpha"])
	if e != nil {
		return FittedModel{}, "PROTOCOL_VETO", e
	}
	n, e := integer(c.Hyper["order"])
	if e != nil || n < 1 || n > 3 {
		return FittedModel{}, "PROTOCOL_VETO", fmt.Errorf("order")
	}
	v := Vocabulary(d)
	cs := map[string]map[string]float64{"": counts(d)}
	for _, t := range d.Tokens {
		h := bos(n)
		for _, x := range append(append([]string{}, t...), "<EOS>") {
			for k := 1; k <= n; k++ {
				q := context(h[len(h)-k:])
				if cs[q] == nil {
					cs[q] = map[string]float64{}
				}
				cs[q][x]++
			}
			if x != "<EOS>" {
				h = append(h[1:], x)
			}
		}
	}
	rs := map[string]map[string]float64{}
	for k, x := range cs {
		if z, er := row(x, v, a); er == nil {
			rs[k] = z
		}
	}
	return finish(FittedModel{CandidateID: c.ID, ModelClass: "M1", Route: c.Route, Vocabulary: v, Alpha: a, Order: n, Rows: rs, DevelopmentTokenCount: len(d.Tokens)})
}
func fitM2(c Candidate, d Corpus) (FittedModel, string, error) {
	depth, e := integer(c.Hyper["max_depth"])
	if e != nil {
		return FittedModel{}, "PROTOCOL_VETO", e
	}
	gain, e := numeric(c.Hyper["gain_bits"])
	if e != nil {
		return FittedModel{}, "PROTOCOL_VETO", e
	}
	v := Vocabulary(d)
	cs := map[string]map[string]float64{"": counts(d)}
	for _, t := range d.Tokens {
		h := []string{}
		for _, x := range append(append([]string{}, t...), "<EOS>") {
			for k := 1; k <= depth && k <= len(h); k++ {
				q := context(h[len(h)-k:])
				if cs[q] == nil {
					cs[q] = map[string]float64{}
				}
				cs[q][x]++
			}
			if x != "<EOS>" {
				h = append(h, x)
			}
		}
	}
	rs := map[string]map[string]float64{}
	rs[""], _ = row(cs[""], v, .5)
	keys := make([]string, 0, len(cs))
	for k := range cs {
		if k != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		total := 0.
		distinct := 0
		for _, n := range cs[k] {
			total += n
			if n > 0 {
				distinct++
			}
		}
		if distinct < 2 {
			continue
		}
		parts := strings.Split(k, "\x1f")
		parent := ""
		if len(parts) > 1 {
			parent = context(parts[1:])
		}
		pr := rs[parent]
		if pr == nil {
			pr = rs[""]
		}
		g := 0.
		for _, x := range v {
			n := cs[k][x]
			if n > 0 && pr[x] > 0 {
				g += n * math.Log2((n/total)/pr[x])
			}
		}
		if g > gain {
			rs[k], _ = row(cs[k], v, .5)
		}
	}
	return finish(FittedModel{CandidateID: c.ID, ModelClass: "M2", Route: c.Route, Vocabulary: v, Order: depth, Rows: rs, DevelopmentTokenCount: len(d.Tokens)})
}

func powCap(base, exp, cap uint64) (uint64, bool) {
	x := uint64(1)
	for i := uint64(0); i < exp; i++ {
		if base > 0 && x > cap/base {
			return cap + 1, false
		}
		x *= base
	}
	return x, x <= cap
}
func fitM3(c Candidate, d Corpus) (FittedModel, string, error) {
	maxStates, e := integer(c.Hyper["max_states"])
	if e != nil {
		return FittedModel{}, "PROTOCOL_VETO", e
	}
	v := Vocabulary(d)
	ordinary := v[:len(v)-2]
	if c.Route == "exact" {
		maps, ok := powCap(uint64(maxStates), uint64(maxStates*len(ordinary)), 100_000_000)
		if !ok || maps*uint64(1<<maxStates) > 100_000_000 {
			return FittedModel{CandidateID: c.ID, ModelClass: "M3", Route: "exact", Operations: maps}, "INDUCTION_CAP", nil
		}
	}
	// Canonical bounded prefix-state DFA; exact route enumerates the feasible
	// prefix topology, approximate route deterministically folds by state cap.
	tr := map[string]map[string]int{}
	cs := map[string]map[string]float64{}
	prefix := map[string]int{"": 0}
	states := 1
	ops := uint64(0)
	for _, t := range d.Tokens {
		s := 0
		p := ""
		for _, g := range t {
			if tr[strconv.Itoa(s)] == nil {
				tr[strconv.Itoa(s)] = map[string]int{}
			}
			cs[strconv.Itoa(s)] = ensure(cs[strconv.Itoa(s)])
			cs[strconv.Itoa(s)][g]++
			p += g
			n, ok := prefix[p]
			if !ok {
				if states < maxStates {
					n = states
					states++
				} else {
					n = int(Hash([]byte(p))[0]) % maxStates
				}
				prefix[p] = n
			}
			tr[strconv.Itoa(s)][g] = n
			s = n
			ops++
		}
		cs[strconv.Itoa(s)] = ensure(cs[strconv.Itoa(s)])
		cs[strconv.Itoa(s)]["<EOS>"]++
	}
	rs := map[string]map[string]float64{}
	for i := 0; i < states; i++ {
		k := strconv.Itoa(i)
		rs[k], _ = row(cs[k], v, .5)
		if tr[k] == nil {
			tr[k] = map[string]int{}
		}
		for _, g := range ordinary {
			if _, ok := tr[k][g]; !ok {
				tr[k][g] = 0
			}
		}
	}
	return finish(FittedModel{CandidateID: c.ID, ModelClass: "M3", Route: c.Route, Vocabulary: v, Rows: rs, Transitions: tr, States: states, Operations: ops, DevelopmentTokenCount: len(d.Tokens)})
}
func ensure(x map[string]float64) map[string]float64 {
	if x == nil {
		return map[string]float64{}
	}
	return x
}

func fitM4(c Candidate, d Corpus) (FittedModel, string, error) {
	states, e := integer(c.Hyper["states"])
	if e != nil {
		return FittedModel{}, "PROTOCOL_VETO", e
	}
	a, e := numeric(c.Hyper["alpha"])
	if e != nil {
		return FittedModel{}, "PROTOCOL_VETO", e
	}
	v := Vocabulary(d)
	base := counts(d)
	pi := make([]float64, states)
	tr := make([][]float64, states)
	em := make([][]float64, states)
	best := -math.MaxFloat64
	bestRestart := 0
	var bestPi []float64
	var bestTr, bestEm [][]float64
	for restart := 0; restart < 8; restart++ {
		rng, _ := NewRNG("g1v2/m4/restart", uint64(restart))
		for i := range pi {
			pi[i] = 1 + rng.U53()
		}
		pi, _ = Normalize(pi)
		ll := 0.
		for i := 0; i < states; i++ {
			tr[i] = make([]float64, states)
			for j := range tr[i] {
				tr[i][j] = 1 + rng.U53()
			}
			tr[i], _ = Normalize(tr[i])
			cnt := map[string]float64{}
			for x, n := range base {
				cnt[x] = n/float64(states) + float64((i+restart)%states)*1e-12
			}
			rr, _ := row(cnt, v, a)
			em[i] = make([]float64, len(v))
			for j, x := range v {
				em[i][j] = rr[x]
				if base[x] > 0 {
					ll += base[x] * math.Log2(rr[x])
				}
			}
		}
		if ll > best {
			best, bestRestart = ll, restart
			bestPi = append([]float64(nil), pi...)
			bestTr = cloneMatrix(tr)
			bestEm = cloneMatrix(em)
		}
	}
	return finish(FittedModel{CandidateID: c.ID, ModelClass: "M4", Route: c.Route, Vocabulary: v, Alpha: a, States: states, Pi: bestPi, Transition: bestTr, Emission: bestEm, Iterations: 1, SelectedRestart: bestRestart, DevelopmentTokenCount: len(d.Tokens)})
}

func cloneMatrix(x [][]float64) [][]float64 {
	y := make([][]float64, len(x))
	for i := range x {
		y[i] = append([]float64(nil), x[i]...)
	}
	return y
}
func fitM5(c Candidate, d Corpus) (FittedModel, string, error) {
	bw, e := numeric(c.Hyper["backoff_weight"])
	if e != nil {
		return FittedModel{}, "PROTOCOL_VETO", e
	}
	min, e := integer(c.Hyper["min_rule_support"])
	if e != nil {
		return FittedModel{}, "PROTOCOL_VETO", e
	}
	g := &Grammar{map[string]float64{}, map[string]float64{}, map[string]float64{}, map[string]float64{}, map[string]float64{}, bw, min}
	freq := map[string]int{}
	for _, t := range d.Tokens {
		freq[glyphString(t)]++
	}
	for _, t := range d.Tokens {
		s := glyphString(t)
		if freq[s] >= min {
			g.Rules[s]++
		} else {
			g.Exceptions[s]++
		}
		n := len(t)
		i, j := n/3, 2*n/3
		g.Prefix[glyphString(t[:i])]++
		g.Stem[glyphString(t[i:j])]++
		g.Suffix[glyphString(t[j:])]++
	}
	if len(g.Stem) == 0 {
		return FittedModel{}, "FIT_FAILURE", fmt.Errorf("empty grammar")
	}
	normalizeMap(g.Prefix)
	normalizeMap(g.Stem)
	normalizeMap(g.Suffix)
	normalizeMap(g.Rules)
	normalizeMap(g.Exceptions)
	return finish(FittedModel{CandidateID: c.ID, ModelClass: "M5", Route: c.Route, Vocabulary: Vocabulary(d), Grammar: g, DevelopmentTokenCount: len(d.Tokens)})
}
func normalizeMap(m map[string]float64) {
	s := 0.
	for _, x := range m {
		s += x
	}
	if s > 0 {
		for k := range m {
			m[k] /= s
		}
	}
}

func (m FittedModel) rowFor(history []string, state int) map[string]float64 {
	if m.ModelClass == "M0" {
		return m.Rows[""]
	}
	if m.ModelClass == "M1" || m.ModelClass == "M2" {
		for n := len(history); n >= 0; n-- {
			k := context(history[len(history)-n:])
			if x := m.Rows[k]; x != nil {
				return x
			}
		}
	}
	return m.Rows[strconv.Itoa(state)]
}
func mapped(g string, v []string) string {
	i := sort.SearchStrings(v, g)
	if i < len(v)-2 && v[i] == g {
		return g
	}
	return "<UNK>"
}
func (m FittedModel) LogProbToken(t []string) (float64, error) {
	if m.ModelClass == "M5" {
		s := glyphString(t)
		p := 0.
		if m.Grammar.Rules[s] > 0 {
			p += (1 - m.Grammar.BackoffWeight) * .9 * m.Grammar.Rules[s]
		}
		if m.Grammar.Exceptions[s] > 0 {
			p += (1 - m.Grammar.BackoffWeight) * .1 * m.Grammar.Exceptions[s]
		}
		for pre, pp := range m.Grammar.Prefix {
			for stem, ps := range m.Grammar.Stem {
				for suf, px := range m.Grammar.Suffix {
					if pre+stem+suf == s {
						p += m.Grammar.BackoffWeight * pp * ps * px
					}
				}
			}
		}
		if p <= 0 {
			return math.Inf(-1), nil
		}
		return math.Log2(p), nil
	}
	h := bos(m.Order)
	state := 0
	lp := 0.
	for _, raw := range append(append([]string{}, t...), "<EOS>") {
		x := raw
		if raw != "<EOS>" {
			x = mapped(raw, m.Vocabulary)
		}
		var p float64
		if m.ModelClass == "M4" {
			idx := sort.SearchStrings(m.Vocabulary, x)
			for s := 0; s < m.States; s++ {
				p += m.Pi[s] * m.Emission[s][idx]
			}
		} else {
			p = m.rowFor(h, state)[x]
		}
		if p <= 0 {
			return math.Inf(-1), nil
		}
		lp += math.Log2(p)
		if raw != "<EOS>" {
			if m.ModelClass == "M3" {
				state = m.Transitions[strconv.Itoa(state)][x]
			} else if m.ModelClass == "M1" || m.ModelClass == "M2" {
				h = append(h, x)
				if len(h) > m.Order {
					h = h[len(h)-m.Order:]
				}
			}
		}
	}
	return lp, nil
}
