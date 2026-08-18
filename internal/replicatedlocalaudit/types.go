package replicatedlocalaudit

import "io"

type Config struct {
	CorpusPath, MetadataPath, RelationDir, DiscoveryDir, OutputDir, CheckpointPath string
	Permutations                                                                   int
	Seed                                                                           int64
	Quiet                                                                          bool
	// Generic selects task43's corpus-only generic mode: tokens/blocks are
	// derived from internal/genericsegmentation instead of a real
	// IVTFF-sourced MetadataPath file, "replicate" status never claims a
	// Currier/hand-conditioned finding, and RelationDir is expected to hold
	// stage 23's own generic-mode output (see load.go/run.go/write.go).
	Generic        bool
	ProgressWriter io.Writer
}

type token struct {
	Text, Line, Currier, Hand, Joint, Block string
}

type block struct {
	ID, Currier, Hand, Joint string
	Tokens                   []token
}

type distanceCandidate struct {
	ID, A, B, Classification                                  string
	Eligible, Joint, Currier, Hands                           int
	Mean, Median, Min, Transfer, RawP, Q, Threshold, Fraction float64
}

type sequenceCandidate struct {
	ID, Sequence, Classification string
	Tokens                       []string
}

type seqObserved struct {
	Total, Eligible, Blocks, Joint, Currier, Hands    int
	MaxFraction, Entropy                              float64
	Validity                                          string
	TokenOccurrences                                  []int
	ContainsQuestion, ContainsAt, ContainsOtherMarker bool
	AbsentTokens                                      []string
}

type distanceEffect struct {
	CandidateID, Block string
	Observed           float64
	Nulls              []float64
}

type distanceRow struct {
	CandidateID, A, B, Block, Currier, Hand, Joint     string
	CountA, CountB, Observations, ComparedProfileCells int
	Similarity, LOBO, ShapeSimilarity                  float64
	Peak                                               int
	Center, Asymmetry                                  float64
	Transfer                                           bool
	NullMean, NullSD, P95, P99, Standardized, Effect   float64
}

type distanceResult struct {
	Candidate                                                       distanceCandidate
	Rows                                                            []distanceRow
	Positive, Negative                                              int
	MeanZ, WeightedZ, MaxObservationFraction, MaxEffectContribution float64
	FullEffect, MinJackknife, MaxJackknife, JackknifeSD             float64
	MaxJackknifeP                                                   float64
	JackknifeSurvives, CrossCurrier, CrossHand, CrossJoint          bool
	WithinCurrier, WithinHand, WithinJoint                          bool
	Status, FailedConditions                                        string
}

type sequenceResult struct {
	Candidate                                                              sequenceCandidate
	Observed                                                               seqObserved
	ShuffleP, ShuffleQ, ShuffleTotalP, ShuffleMeanBlocks, ShuffleMeanTotal float64
	MarkovP, MarkovTotalP, MarkovMeanBlocks, MarkovMeanTotal               float64
	MarkovAvailableBlocks                                                  int
	Status                                                                 string
}

type checkpoint struct {
	Version, Permutations int
	Fingerprint           string
	DistanceCompleted     int
	Distance              map[string][]float64
	ShuffleCompleted      int
	ShuffleExceedBlocks   map[string]int
	ShuffleExceedTotal    map[string]int
	ShuffleSumBlocks      map[string]float64
	ShuffleSumTotal       map[string]float64
	MarkovCompleted       int
	MarkovExceedBlocks    map[string]int
	MarkovExceedTotal     map[string]int
	MarkovSumBlocks       map[string]float64
	MarkovSumTotal        map[string]float64
}
