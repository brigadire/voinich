// Package numeric implements the exploratory positional-number experiment.
package numeric

type Config struct {
	CorpusPath, IVTFFPath, IT2aPath, IT2aIVTFFPath, NaturalPath, OutputDir string
	Replicates, OptimizerSteps, Restarts                                   int
	Seed                                                                   int64
}

type Token struct {
	Text                      string
	Glyphs                    []byte
	Line, IndexInLine         int
	Folio, Section, LocusType string
}

type Corpus struct {
	Name, Path, SHA256, IVTFFPath, IVTFFSHA256                     string
	Tokens                                                         []Token
	RawTokenCount, UniqueTokenCount, ExcludedTokenCount, LineCount int
	Alphabet                                                       []byte
}

type Metrics struct {
	LengthMean, LogMean, AdjLengthDiffMean, PositionLengthRho                     float64
	SignedDeltaMean, AbsDeltaMean, NormalizedDeltaMean                            float64
	DeltaEntropy, RepeatedDeltaFraction                                           float64
	IncreasingFraction, DecreasingFraction, LongestMonotonicRun, PositionValueRho float64
	LagRho                                                                        [5]float64
	APCloseness, RatioRepeat, LeadingZeroFraction, CollisionFraction              float64
	LeadingZeroTokenCount, CollidingTokenTypeCount, CollisionClassCount           float64
	EditSubstitutionConsistency                                                   float64
	Score, SequentialComponent, DifferenceComponent, DocumentComponent            float64
}

type MappingResult struct {
	Corpus, Control string
	Replicate       int
	Seed            int64
	Baseline, Best  Metrics
	Mapping         []int
}
