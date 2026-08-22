// Package tokenrepetition implements task60's independent analysis of
// exact adjacent repetition, exact runs, near-repetition (edit-distance-1
// adjacency), and illustration-label repetition in the canonical Voynich
// corpus, plus the natural-language and existing homophonic controls
// needed to judge compatibility with simple homophonic substitution.
//
// This is deliberately not a pipeline stage (task60's own instruction);
// it is driven by independent/token-repetition-analyze.
package tokenrepetition

// Corpus is a loaded token stream with natural line boundaries.
type Corpus struct {
	Name        string
	Path        string
	SHA256      string
	Tokens      []string
	LineOfToken []int
	Opaque      bool // true for Task46/55 xNNNNNN ciphertext: glyph-level analysis is NOT_APPLICABLE
}

// RepeatedToken summarizes one token type's exact-adjacent-repeat behavior.
type RepeatedToken struct {
	Token            string
	Frequency        int
	AdjacentRepeats  int
	MaximumRun       int
	FirstLoci        []int // corpus positions (token index) of run starts, capped
}

// Run is one maximal exact run w^k, k>=2.
type Run struct {
	Token          string
	RunLength      int
	StartPosition  int // token index of the run's first occurrence
	GlobalFrequency int
}

// EditEvent is one classified adjacent pair with edit distance 1.
type EditEvent struct {
	Position   int // token index of the left member of the pair
	A, B       string
	Operation  string // SUBSTITUTION, INSERTION, DELETION
	PositionClass string // BEGIN, MIDDLE, END
	SourceGlyph, TargetGlyph string // for SUBSTITUTION; for INSERTION/DELETION, the inserted/deleted glyph
}

// Chain is a maximal near-repeat mutation chain (edit distance <=1 between
// consecutive members) of length >= 3, found in the edit-family graph.
type Chain struct {
	Tokens []string
}
