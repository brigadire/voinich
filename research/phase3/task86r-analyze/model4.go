package main

import "sort"

// M4: shared-M3-topology probabilistic finite-state-v1.
// G1_EXECUTABLE_CONTRACT.json models.M4. Reuses M3's induced topology
// (same merge_js_threshold/max_states on the same DEVELOPMENT input) and
// the transition/accept counts already accumulated during that induction
// ("replay DEVELOPMENT paths" -- no EM, since the merge itself already
// aggregated occurrence-weighted counts per surviving state).
type M4Model struct {
	candidateID string
	alpha       float64
	fsa         *FSA
	bitsReal    float64
}

func FitM4(occ []TokenOccurrence, cand Candidate, bitsReal float64) *M4Model {
	threshold := cand.Float("merge_js_threshold", 0)
	maxStates := cand.Int("max_states", 64)
	alpha := cand.Float("alpha", 0.1)
	return &M4Model{
		candidateID: cand.CandidateID, alpha: alpha,
		fsa: InduceFSA(occ, threshold, maxStates), bitsReal: bitsReal,
	}
}

func (m *M4Model) ModelClass() string  { return "M4" }
func (m *M4Model) CandidateID() string { return m.candidateID }
func (m *M4Model) Unit() string        { return "GLYPH" }

func (m *M4Model) TrainingFailed() (bool, string) { return m.fsa.Failed, m.fsa.FailWhy }

// alternatives mirrors M3's enabled-outgoing set (glyph edges, plus EOS
// iff the state has a nonzero accept count).
func (m *M4Model) alternatives(state int) []string {
	edges := m.fsa.Edges[state]
	labels := make([]string, 0, len(edges)+1)
	for g := range edges {
		labels = append(labels, g)
	}
	if m.fsa.Accept[state] > 0 {
		labels = append(labels, eosSymbol)
	}
	sort.Strings(labels)
	return labels
}

func (m *M4Model) decisionAt(state int, observed string) Decision {
	outs := m.alternatives(state)
	K := float64(len(outs))
	total := float64(m.fsa.Accept[state])
	for _, e := range m.fsa.Edges[state] {
		total += float64(e.count)
	}
	denom := total + m.alpha*K
	probs := make([]float64, len(outs))
	for i, o := range outs {
		var c float64
		if o == eosSymbol {
			c = float64(m.fsa.Accept[state])
		} else {
			c = float64(m.fsa.Edges[state][o].count)
		}
		if denom > 0 {
			probs[i] = (c + m.alpha) / denom
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

func (m *M4Model) Events(raw string, glyphs []string) []ScoreEvent {
	if m.fsa.Failed {
		return []ScoreEvent{{Confidence: 0, Correct: false, NegLog2Prob: posInf()}}
	}
	state := m.fsa.Root
	var events []ScoreEvent
	for _, g := range glyphs {
		d := m.decisionAt(state, g)
		events = append(events, d.Event())
		if d.ObservedIndex < 0 {
			return events
		}
		state = m.fsa.Edges[state][g].target
	}
	d := m.decisionAt(state, eosSymbol)
	events = append(events, d.Event())
	return events
}

func (m *M4Model) TokenNegLog2Prob(raw string, glyphs []string) float64 {
	total := 0.0
	for _, e := range m.Events(raw, glyphs) {
		total += e.NegLog2Prob
	}
	return total
}

func (m *M4Model) ScoredUnits(glyphs []string) int { return len(glyphs) + 1 }

func (m *M4Model) Generate(p *PRNG) GeneratedToken {
	if m.fsa.Failed {
		return GeneratedToken{NonGenerative: true}
	}
	state := m.fsa.Root
	var glyphs []string
	for step := 0; step < maxTokenGlyphs; step++ {
		d := m.decisionAt(state, "")
		if len(d.Outcomes) == 0 {
			return GeneratedToken{Glyphs: glyphs, NonGenerative: true}
		}
		idx := DrawIndex(p, Cumulative(d.Probs))
		sym := d.Outcomes[idx]
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

// Complexity: same structural choice points as M3 (states x outgoing
// alternatives) plus one free real parameter (alpha) charged once via the
// frozen BIC-style constant, per state (each state's transition
// distribution is a separately fitted categorical).
func (m *M4Model) Complexity() ComplexityBreakdown {
	if m.fsa.Failed {
		return ComplexityBreakdown{}
	}
	structure := 0.0
	freeParams := 0
	for _, s := range m.fsa.States {
		n := len(m.alternatives(s))
		if n > 0 {
			structure += log2(float64(n))
			freeParams++
		}
	}
	structure += m.bitsReal * float64(freeParams)
	return ComplexityBreakdown{StructureCost: structure, FreeParams: freeParams, States: len(m.fsa.States)}
}
