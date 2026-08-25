package main

import (
	"fmt"
	"sort"
	"strings"
)

// M2: PPM-C variable-order Markov-v1. G1_EXECUTABLE_CONTRACT.json models.M2.
// Depths run max_depth downTo 1 (n-gram contexts), then depth 0 (empty
// context / order-0 marginal), then depth -1 (uniform floor over
// alphabet+EOS+UNK_GLYPH). Exclusion: a symbol assigned probability at a
// deeper level is removed from consideration at every shallower level.
type M2Model struct {
	candidateID     string
	maxDepth        int
	minContextCount int
	alphabet        []string
	bos             []string

	tables map[int]map[string]map[string]int // depth 1..maxDepth
	totals map[int]map[string]int
	count0 map[string]int
	c0     int
	bitsReal float64
	floorSet []string // sorted alphabet+EOS+UNK
}

func FitM2(occ []TokenOccurrence, cand Candidate, bitsReal float64) *M2Model {
	maxDepth := cand.Int("max_depth", 2)
	minCount := cand.Int("min_context_count", 1)
	alphabet := glyphAlphabet(occ)
	bos := make([]string, maxDepth)
	for i := range bos {
		bos[i] = fmt.Sprintf("\x01BOS%d", i)
	}
	m := &M2Model{
		candidateID: cand.CandidateID, maxDepth: maxDepth, minContextCount: minCount,
		alphabet: alphabet, bos: bos,
		tables: map[int]map[string]map[string]int{}, totals: map[int]map[string]int{},
		count0: map[string]int{}, bitsReal: bitsReal,
	}
	for d := 1; d <= maxDepth; d++ {
		m.tables[d] = map[string]map[string]int{}
		m.totals[d] = map[string]int{}
	}
	for _, o := range occ {
		seq := make([]string, 0, maxDepth+len(o.Glyphs)+1)
		seq = append(seq, bos...)
		seq = append(seq, o.Glyphs...)
		seq = append(seq, eosSymbol)
		for i := maxDepth; i < len(seq); i++ {
			pred := seq[i]
			full := seq[i-maxDepth : i]
			m.count0[pred]++
			m.c0++
			for d := 1; d <= maxDepth; d++ {
				ctx := full[maxDepth-d:]
				key := strings.Join(ctx, "\x00")
				if m.tables[d][key] == nil {
					m.tables[d][key] = map[string]int{}
				}
				m.tables[d][key][pred]++
				m.totals[d][key]++
			}
		}
	}
	floor := append([]string{}, alphabet...)
	floor = append(floor, eosSymbol, unkGlyphSymbol)
	sort.Strings(floor)
	m.floorSet = floor
	return m
}

func (m *M2Model) ModelClass() string  { return "M2" }
func (m *M2Model) CandidateID() string { return m.candidateID }
func (m *M2Model) Unit() string        { return "GLYPH" }

// retained reports whether context h at depth d survives pruning
// (min_context_count applies to depth>=1 only; depth 0 always retained).
func (m *M2Model) retained(d int, key string) (map[string]int, int, bool) {
	if d == 0 {
		return m.count0, m.c0, true
	}
	total, ok := m.totals[d][key]
	if !ok || total < m.minContextCount {
		return nil, 0, false
	}
	return m.tables[d][key], total, true
}

// fullDistribution computes the complete PPM-C escape-blended
// distribution over floorSet for the given (padded, length maxDepth)
// history.
func (m *M2Model) fullDistribution(history []string) map[string]float64 {
	probs := map[string]float64{}
	excluded := map[string]bool{}
	remaining := 1.0
	for d := m.maxDepth; d >= 0; d-- {
		var key string
		if d > 0 {
			key = strings.Join(history[m.maxDepth-d:], "\x00")
		}
		counts, _, ok := m.retained(d, key)
		if !ok {
			continue
		}
		syms := make([]string, 0, len(counts))
		total := 0
		// Integer count accumulation is exact regardless of map
		// iteration order; sorting is unnecessary here (the exported
		// Decision.Outcomes ordering is sorted separately).
		for s, c := range counts {
			if excluded[s] {
				continue
			}
			syms = append(syms, s)
			total += c
		}
		t := len(syms)
		if t == 0 {
			continue
		}
		denom := float64(total + t)
		for _, s := range syms {
			probs[s] = remaining * float64(counts[s]) / denom
			excluded[s] = true
		}
		remaining = remaining * float64(t) / denom
	}
	var floor []string
	for _, s := range m.floorSet {
		if !excluded[s] {
			floor = append(floor, s)
		}
	}
	k := len(floor)
	for _, s := range floor {
		probs[s] = remaining / float64(k)
	}
	return probs
}

func (m *M2Model) remapTarget(observed string) string {
	if observed == eosSymbol {
		return eosSymbol
	}
	for _, a := range m.alphabet {
		if a == observed {
			return observed
		}
	}
	return unkGlyphSymbol
}

func (m *M2Model) decisionAt(history []string, observed string, forScoring bool) Decision {
	dist := m.fullDistribution(history)
	var outs []string
	if forScoring {
		outs = m.floorSet
	} else {
		for _, s := range m.floorSet {
			if s != unkGlyphSymbol {
				outs = append(outs, s)
			}
		}
	}
	probs := make([]float64, len(outs))
	sum := 0.0
	for i, s := range outs {
		probs[i] = dist[s]
		sum += probs[i]
	}
	if !forScoring && sum > 0 {
		for i := range probs {
			probs[i] /= sum
		}
	}
	obsIdx := -1
	target := m.remapTarget(observed)
	for i, s := range outs {
		if s == target {
			obsIdx = i
			break
		}
	}
	return Decision{Outcomes: outs, Probs: probs, ObservedIndex: obsIdx}
}

func (m *M2Model) Events(raw string, glyphs []string) []ScoreEvent {
	seq := make([]string, 0, m.maxDepth+len(glyphs)+1)
	seq = append(seq, m.bos...)
	seq = append(seq, glyphs...)
	seq = append(seq, eosSymbol)
	events := make([]ScoreEvent, 0, len(seq)-m.maxDepth)
	for i := m.maxDepth; i < len(seq); i++ {
		history := seq[i-m.maxDepth : i]
		d := m.decisionAt(history, seq[i], true)
		events = append(events, d.Event())
	}
	return events
}

func (m *M2Model) TokenNegLog2Prob(raw string, glyphs []string) float64 {
	total := 0.0
	for _, e := range m.Events(raw, glyphs) {
		total += e.NegLog2Prob
	}
	return total
}

func (m *M2Model) ScoredUnits(glyphs []string) int { return len(glyphs) + 1 }

func (m *M2Model) Generate(p *PRNG) GeneratedToken {
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
		history = append(history[1:], sym)
	}
	return GeneratedToken{Glyphs: glyphs, Truncated: true}
}

func (m *M2Model) Complexity() ComplexityBreakdown {
	A := float64(len(m.alphabet) + 1)
	structure := 0.0
	freeParams := 0
	contexts := 0
	for d := 1; d <= m.maxDepth; d++ {
		for _, key := range sortedIntKeys(m.totals[d]) {
			if m.totals[d][key] < m.minContextCount {
				continue
			}
			contexts++
			structure += log2(A + 1)
			freeParams++
		}
	}
	structure += m.bitsReal * float64(freeParams+1)
	lex := 0.0
	for _, x := range sortedIntKeys(m.count0) {
		p := float64(m.count0[x]) / float64(m.c0)
		lex += -log2(p)
	}
	return ComplexityBreakdown{
		StructureCost: structure, LexiconCost: lex, ExceptionCost: 0,
		FreeParams: freeParams + 1, States: contexts + 1,
	}
}

func (m *M2Model) TrainingFailed() (bool, string) { return false, "" }
