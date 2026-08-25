package main

import (
	"fmt"
	"sort"
	"strings"
)

// M1: fixed-order glyph Markov additive-backoff-v1.
// G1_EXECUTABLE_CONTRACT.json models.M1. History length is order-1;
// "order-1 distinct BOS symbols" are prepended so every predicted
// position, including the first real glyph, has a full-length history.
type M1Model struct {
	candidateID string
	order       int
	alpha       float64
	alphabet    []string // sorted, DEVELOPMENT glyph inventory
	bos         []string // order-1 distinct positional BOS symbols

	// tables[depth][joined(context)] -> counts per next symbol; depth
	// ranges 1..order-1. depth0 is the flat marginal count table.
	tables map[int]map[string]map[string]int
	totals map[int]map[string]int
	count0 map[string]int
	c0     int
	bitsReal float64
}

func FitM1(occ []TokenOccurrence, cand Candidate, bitsReal float64) *M1Model {
	order := cand.Int("order", 1)
	alpha := cand.Float("alpha", 0.1)
	alphabet := glyphAlphabet(occ)
	histLen := order - 1
	bos := make([]string, histLen)
	for i := range bos {
		bos[i] = fmt.Sprintf("\x01BOS%d", i)
	}
	m := &M1Model{
		candidateID: cand.CandidateID, order: order, alpha: alpha, alphabet: alphabet, bos: bos,
		tables: map[int]map[string]map[string]int{}, totals: map[int]map[string]int{},
		count0: map[string]int{}, bitsReal: bitsReal,
	}
	for d := 1; d <= histLen; d++ {
		m.tables[d] = map[string]map[string]int{}
		m.totals[d] = map[string]int{}
	}
	for _, o := range occ {
		seq := make([]string, 0, histLen+len(o.Glyphs)+1)
		seq = append(seq, bos...)
		seq = append(seq, o.Glyphs...)
		seq = append(seq, eosSymbol)
		for i := histLen; i < len(seq); i++ {
			pred := seq[i]
			full := seq[i-histLen : i]
			m.count0[pred]++
			m.c0++
			for d := 1; d <= histLen; d++ {
				ctx := full[histLen-d:]
				key := strings.Join(ctx, "\x00")
				if m.tables[d][key] == nil {
					m.tables[d][key] = map[string]int{}
				}
				m.tables[d][key][pred]++
				m.totals[d][key]++
			}
		}
	}
	return m
}

func (m *M1Model) ModelClass() string  { return "M1" }
func (m *M1Model) CandidateID() string { return m.candidateID }
func (m *M1Model) Unit() string        { return "GLYPH" }

// A is alphabet size + EOS, matching the estimator's "(A+1)" base measure.
func (m *M1Model) baseA() int { return len(m.alphabet) + 1 }

func (m *M1Model) depth0Prob(x string) float64 {
	A := float64(m.baseA())
	denom := float64(m.c0) + m.alpha*(A+1)
	return (float64(m.count0[x]) + m.alpha) / denom
}

// contextKeys precomputes, once per query position, the joined map key at
// every depth 1..len(history) (keys[d-1] is the depth-d key, i.e. the
// joined last d symbols of history) -- avoiding O(alphabet size) redundant
// strings.Join calls per outcome in decisionAt/prob.
func (m *M1Model) contextKeys(history []string) []string {
	n := len(history)
	if n == 0 {
		return nil
	}
	keys := make([]string, n)
	// keys[n-1] = full history joined; keys[k] = history[n-1-k:] joined.
	keys[n-1] = strings.Join(history, "\x00")
	for d := n - 1; d >= 1; d-- {
		keys[d-1] = strings.Join(history[n-d:], "\x00")
	}
	return keys
}

// probWithKeys is prob() using precomputed per-depth keys (keys[d-1] is
// the depth-d key), avoiding repeated joins across the outcome loop.
func (m *M1Model) probWithKeys(keys []string, depth int, x string) float64 {
	if depth == 0 {
		return m.depth0Prob(x)
	}
	key := keys[depth-1]
	total := m.totals[depth][key]
	back := m.probWithKeys(keys, depth-1, x)
	if total == 0 {
		return back
	}
	counts := m.tables[depth][key]
	return (float64(counts[x]) + m.alpha*back) / (float64(total) + m.alpha)
}

func (m *M1Model) outcomesForScoring() []string {
	out := append([]string{}, m.alphabet...)
	out = append(out, eosSymbol, unkGlyphSymbol)
	sort.Strings(out)
	return out
}

func (m *M1Model) decisionAt(history []string, observed string, forScoring bool) Decision {
	var outs []string
	if forScoring {
		outs = m.outcomesForScoring()
	} else {
		outs = append([]string{}, m.alphabet...)
		outs = append(outs, eosSymbol)
		sort.Strings(outs)
	}
	keys := m.contextKeys(history)
	probs := make([]float64, len(outs))
	sum := 0.0
	for i, o := range outs {
		probs[i] = m.probWithKeys(keys, len(history), o)
		sum += probs[i]
	}
	if !forScoring && sum > 0 {
		for i := range probs {
			probs[i] /= sum
		}
	}
	obsIdx := -1
	target := observed
	if forScoring {
		found := false
		for _, a := range m.alphabet {
			if a == observed {
				found = true
				break
			}
		}
		if observed != eosSymbol && !found {
			target = unkGlyphSymbol
		}
	}
	for i, o := range outs {
		if o == target {
			obsIdx = i
			break
		}
	}
	return Decision{Outcomes: outs, Probs: probs, ObservedIndex: obsIdx}
}

func (m *M1Model) histLen() int { return m.order - 1 }

func (m *M1Model) Events(raw string, glyphs []string) []ScoreEvent {
	seq := make([]string, 0, m.histLen()+len(glyphs)+1)
	seq = append(seq, m.bos...)
	seq = append(seq, glyphs...)
	seq = append(seq, eosSymbol)
	events := make([]ScoreEvent, 0, len(seq)-m.histLen())
	for i := m.histLen(); i < len(seq); i++ {
		history := seq[i-m.histLen() : i]
		d := m.decisionAt(history, seq[i], true)
		events = append(events, d.Event())
	}
	return events
}

func (m *M1Model) TokenNegLog2Prob(raw string, glyphs []string) float64 {
	total := 0.0
	for _, e := range m.Events(raw, glyphs) {
		total += e.NegLog2Prob
	}
	return total
}

func (m *M1Model) ScoredUnits(glyphs []string) int { return len(glyphs) + 1 }

func (m *M1Model) Generate(p *PRNG) GeneratedToken {
	history := append([]string{}, m.bos...)
	var glyphs []string
	for step := 0; step < maxTokenGlyphs; step++ {
		d := m.decisionAt(history, "", false)
		idx := DrawIndex(p, Cumulative(d.Probs))
		sym := d.Outcomes[idx]
		if sym == eosSymbol {
			return GeneratedToken{Glyphs: glyphs}
		}
		glyphs = append(glyphs, sym)
		if m.histLen() > 0 {
			history = append(history[1:], sym)
		}
	}
	return GeneratedToken{Glyphs: glyphs, Truncated: true}
}

// Complexity: structure cost is context-count x alphabet-size choice
// points across all realized depths, plus one free real parameter (alpha)
// per depth; lexicon cost is the depth-0 marginal table (the only
// explicitly stored frequency table -- deeper context tables are the
// model's own productive structure, not a separately charged lexicon).
func (m *M1Model) Complexity() ComplexityBreakdown {
	A := float64(m.baseA())
	structure := 0.0
	freeParams := 0
	contexts := 0
	for d := 1; d <= m.histLen(); d++ {
		n := len(m.tables[d])
		contexts += n
		structure += float64(n) * log2(A+1)
		freeParams += n
	}
	structure += m.bitsReal * float64(freeParams+1) // +1 for alpha itself
	lex := 0.0
	for _, x := range sortedIntKeys(m.count0) {
		p := float64(m.count0[x]) / float64(m.c0)
		lex += -log2(p)
	}
	return ComplexityBreakdown{
		StructureCost: structure, LexiconCost: lex, ExceptionCost: 0,
		FreeParams: freeParams + 1, States: contexts + 1, Rules: 0, Components: 0,
	}
}

func (m *M1Model) TrainingFailed() (bool, string) { return false, "" }
