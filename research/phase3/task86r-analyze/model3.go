package main

import (
	"math"
	"sort"
)

// M3: deterministic greedy JS state-merging DFA-v1.
// G1_EXECUTABLE_CONTRACT.json models.M3. Scoring/generation branch
// uniformly among sorted outgoing glyph edges plus EOS (when accepting);
// an unmatched glyph or a non-accepting end-of-token rejects the TOKEN
// (probability 0).
type M3Model struct {
	candidateID string
	fsa         *FSA
}

func FitM3(occ []TokenOccurrence, cand Candidate) *M3Model {
	threshold := cand.Float("merge_js_threshold", 0)
	maxStates := cand.Int("max_states", 64)
	return &M3Model{candidateID: cand.CandidateID, fsa: InduceFSA(occ, threshold, maxStates)}
}

func (m *M3Model) ModelClass() string  { return "M3" }
func (m *M3Model) CandidateID() string { return m.candidateID }
func (m *M3Model) Unit() string        { return "GLYPH" }

func (m *M3Model) TrainingFailed() (bool, string) { return m.fsa.Failed, m.fsa.FailWhy }

func (m *M3Model) alternatives(state int) []string {
	edges := m.fsa.Edges[state]
	labels := make([]string, 0, len(edges)+1)
	for g := range edges {
		labels = append(labels, g)
	}
	sort.Strings(labels)
	if m.fsa.Accept[state] > 0 {
		labels = append(labels, eosSymbol)
		sort.Strings(labels)
	}
	return labels
}

func (m *M3Model) decisionAt(state int, observed string) Decision {
	outs := m.alternatives(state)
	probs := make([]float64, len(outs))
	if len(outs) > 0 {
		p := 1.0 / float64(len(outs))
		for i := range probs {
			probs[i] = p
		}
	}
	obsIdx := -1
	for i, o := range outs {
		if o == observed {
			obsIdx = i
			break
		}
	}
	return Decision{Outcomes: outs, Probs: probs, ObservedIndex: obsIdx}
}

func (m *M3Model) Events(raw string, glyphs []string) []ScoreEvent {
	if m.fsa.Failed {
		return []ScoreEvent{{Confidence: 0, Correct: false, NegLog2Prob: posInf()}}
	}
	state := m.fsa.Root
	var events []ScoreEvent
	for _, g := range glyphs {
		d := m.decisionAt(state, g)
		ev := d.Event()
		events = append(events, ev)
		if d.ObservedIndex < 0 {
			return events // rejected: no further state to walk
		}
		state = m.fsa.Edges[state][g].target
	}
	d := m.decisionAt(state, eosSymbol)
	events = append(events, d.Event())
	return events
}

func (m *M3Model) TokenNegLog2Prob(raw string, glyphs []string) float64 {
	total := 0.0
	for _, e := range m.Events(raw, glyphs) {
		total += e.NegLog2Prob
	}
	return total
}

func (m *M3Model) ScoredUnits(glyphs []string) int { return len(glyphs) + 1 }

func (m *M3Model) Generate(p *PRNG) GeneratedToken {
	if m.fsa.Failed {
		return GeneratedToken{NonGenerative: true}
	}
	state := m.fsa.Root
	var glyphs []string
	for step := 0; step < maxTokenGlyphs; step++ {
		outs := m.alternatives(state)
		if len(outs) == 0 {
			return GeneratedToken{Glyphs: glyphs, NonGenerative: true}
		}
		probs := make([]float64, len(outs))
		p0 := 1.0 / float64(len(outs))
		for i := range probs {
			probs[i] = p0
		}
		idx := DrawIndex(p, Cumulative(probs))
		sym := outs[idx]
		if sym == eosSymbol {
			return GeneratedToken{Glyphs: glyphs}
		}
		glyphs = append(glyphs, sym)
		state = m.fsa.Edges[state][sym].target
	}
	if m.fsa.Accept[state] > 0 {
		return GeneratedToken{Glyphs: glyphs, Truncated: true}
	}
	return GeneratedToken{Glyphs: glyphs, NonGenerative: true}
}

func (m *M3Model) Complexity() ComplexityBreakdown {
	if m.fsa.Failed {
		return ComplexityBreakdown{}
	}
	structure := 0.0
	for _, s := range m.fsa.States {
		n := len(m.alternatives(s))
		if n > 0 {
			structure += log2(float64(n))
		}
	}
	return ComplexityBreakdown{StructureCost: structure, States: len(m.fsa.States)}
}

func posInf() float64 { return math.Inf(1) }
