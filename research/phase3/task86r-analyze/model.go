package main

// ComplexityBreakdown implements research/phase3/task85/GRAMMAR_COMPLEXITY_CONTRACT.md:
// Complexity(G) = StructureCost + LexiconCost + ExceptionCost, in bits.
type ComplexityBreakdown struct {
	StructureCost float64
	LexiconCost   float64
	ExceptionCost float64

	FreeParams int
	States     int
	Rules      int
	Components int
}

func (c ComplexityBreakdown) Total() float64 {
	return c.StructureCost + c.LexiconCost + c.ExceptionCost
}

// bitsPerRealParameter is the frozen BIC-style penalty
// ceil(log2(N_dev))/2 (GRAMMAR_COMPLEXITY_CONTRACT.md section 1),
// computed once per transcription from that transcription's own
// DEVELOPMENT TOKEN count.
func bitsPerRealParameter(nDev int) float64 {
	return ceilLog2(float64(nDev)) / 2
}

// GeneratedToken is one sampled TOKEN from a generative model.
type GeneratedToken struct {
	Raw           string   // M0's chosen type or "<UNK>"
	Glyphs        []string // M1-M5's glyph sequence (excludes EOS)
	NonGenerative bool     // did not terminate within maxTokenGlyphs
	Truncated     bool     // forced EOS at the cap but otherwise valid
}

// FittedModel is the shared interface every M0-M5 candidate fit satisfies.
type FittedModel interface {
	ModelClass() string
	CandidateID() string
	Unit() string // TOKEN, GLYPH, or COMPONENT

	// Events returns the ordered decision points used to score one
	// observed occurrence (raw string for M0, glyph sequence for
	// M1-M5), in generation order, including the terminal EOS/end
	// decision where applicable.
	Events(raw string, glyphs []string) []ScoreEvent

	// TokenNegLog2Prob is -log2 P(occurrence); for M1-M4 this equals
	// the sum of Events' NegLog2Prob by the chain rule, and is
	// computed independently for M0/M5 per their own estimators.
	TokenNegLog2Prob(raw string, glyphs []string) float64

	// ScoredUnits is the PM2 denominator contribution of one
	// occurrence: 1 for M0 (TOKEN-level), len(glyphs)+1 (one EOS) for
	// M1-M5.
	ScoredUnits(glyphs []string) int

	Generate(p *PRNG) GeneratedToken

	Complexity() ComplexityBreakdown

	// TrainingFailed / NonGenerativeClass report frozen failure
	// classes discovered during fitting itself (before any scoring).
	TrainingFailed() (bool, string)
}
