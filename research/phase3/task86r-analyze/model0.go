package main

import (
	"sort"

	"zcore.dev/voinich/internal/evaglyph"
)

// M0: token categorical additive-v1. GRAMMAR_UNIT_REGISTRY TOKEN unit;
// G1_EXECUTABLE_CONTRACT.json models.M0.
type M0Model struct {
	candidateID string
	alpha       float64

	counts   map[string]int
	nDev     int
	vocab    []string // sorted distinct DEVELOPMENT types
	denom    float64
	decision Decision // shared context-free scoring decision (index into vocab+UNK)
	bitsReal float64
}

func FitM0(occ []TokenOccurrence, cand Candidate, bitsReal float64) *M0Model {
	alpha := cand.Float("alpha", 1.0)
	counts := map[string]int{}
	for _, o := range occ {
		counts[o.Raw]++
	}
	vocab := make([]string, 0, len(counts))
	for t := range counts {
		vocab = append(vocab, t)
	}
	sort.Strings(vocab)
	n := len(occ)
	v := len(vocab)
	denom := float64(n) + alpha*float64(v+1)

	outcomes := append(append([]string{}, vocab...), unkTokenLiteral)
	sort.Strings(outcomes)
	probs := make([]float64, len(outcomes))
	for i, t := range outcomes {
		if t == unkTokenLiteral {
			if _, isReal := counts[unkTokenLiteral]; !isReal {
				probs[i] = alpha / denom
				continue
			}
		}
		probs[i] = (float64(counts[t]) + alpha) / denom
	}
	return &M0Model{
		candidateID: cand.CandidateID, alpha: alpha, counts: counts, nDev: n,
		vocab: vocab, denom: denom,
		decision: Decision{Outcomes: outcomes, Probs: probs},
		bitsReal: bitsReal,
	}
}

func (m *M0Model) ModelClass() string  { return "M0" }
func (m *M0Model) CandidateID() string { return m.candidateID }
func (m *M0Model) Unit() string        { return "TOKEN" }

func (m *M0Model) probOf(raw string) float64 {
	for i, o := range m.decision.Outcomes {
		if o == raw {
			return m.decision.Probs[i]
		}
	}
	// raw not observed and not literally "<UNK>": falls into the UNK bucket.
	for i, o := range m.decision.Outcomes {
		if o == unkTokenLiteral {
			return m.decision.Probs[i]
		}
	}
	return 0
}

func (m *M0Model) indexOf(raw string) int {
	if _, ok := m.counts[raw]; !ok {
		raw = unkTokenLiteral
	}
	for i, o := range m.decision.Outcomes {
		if o == raw {
			return i
		}
	}
	return -1
}

func (m *M0Model) Events(raw string, glyphs []string) []ScoreEvent {
	d := m.decision
	d.ObservedIndex = m.indexOf(raw)
	return []ScoreEvent{d.Event()}
}

func (m *M0Model) TokenNegLog2Prob(raw string, glyphs []string) float64 {
	return m.Events(raw, glyphs)[0].NegLog2Prob
}

func (m *M0Model) ScoredUnits(glyphs []string) int { return 1 }

func (m *M0Model) Generate(p *PRNG) GeneratedToken {
	cum := Cumulative(m.decision.Probs)
	idx := DrawIndex(p, cum)
	raw := m.decision.Outcomes[idx]
	if raw == unkTokenLiteral {
		return GeneratedToken{Raw: unkTokenLiteral}
	}
	glyphs := evaglyph.CollapseEVA(raw)
	if len(glyphs) > maxTokenGlyphs {
		return GeneratedToken{Raw: raw, NonGenerative: true}
	}
	return GeneratedToken{Raw: raw}
}

// Complexity: M0's lexicon is its DEVELOPMENT frequency table (every
// distinct type stored); no structural choice points beyond the single
// smoothing parameter alpha; no exceptions.
func (m *M0Model) Complexity() ComplexityBreakdown {
	lex := 0.0
	for _, t := range m.vocab {
		p := float64(m.counts[t]) / float64(m.nDev)
		lex += -log2(p)
	}
	return ComplexityBreakdown{
		StructureCost: m.bitsReal, // one free real parameter: alpha's fitted smoothing role folds into the categorical table itself; alpha is the model's only free real parameter
		LexiconCost:   lex,
		ExceptionCost: 0,
		FreeParams:    1,
		States:        1,
		Rules:         0,
		Components:    len(m.vocab),
	}
}

func (m *M0Model) TrainingFailed() (bool, string) { return false, "" }
