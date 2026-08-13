package metadatavalidation

import "io"

// Config describes a blind validation run.  Frozen discovery files are read,
// never recomputed or modified.
type Config struct {
	IVTFFPath        string
	FrozenCorpusPath string
	DiscoveryDir     string
	OutputDir        string
	Permutations     int
	Seed             int64
	Tolerances       []int
	Quiet            bool
	ProgressWriter   io.Writer
}

type Locus struct {
	Folio, ID, Type, LineID string
	RawText, AlignmentText  string
	ParagraphID             int
	ParagraphStart          bool
	Variables               map[string]string
}

type Document struct {
	Loci          []Locus
	Pages         int
	SkippedLoci   int
	PageVariables map[string]map[string]string
}

type TokenMetadata struct {
	Position, IndexInLocus, IndexInLine, IndexInFolio int
	Token, Folio, LocusID, LocusType, LineID          string
	ParagraphID                                       int
	ParagraphStart                                    bool
	Currier, Hand, Quire                              string
}

type AlignmentResult struct {
	Tokens       []string
	Records      []TokenMetadata
	TotalLoci    int
	SkippedLoci  int
	CorpusSHA256 string
}

type AlignmentError struct {
	Position           int
	Locus              Locus
	Expected, Produced []string
	Before, After      []string
	Reason             string
}

func (e *AlignmentError) Error() string { return FormatAlignmentError(e) }

type MetadataBoundary struct {
	Position       int
	Kind, From, To string
}

type StableBoundary struct {
	Position, Support     int
	MeanJump, Uncertainty float64
}

type BoundaryValidation struct {
	Kind                                        string
	MinSupport, Tolerance, BlindCount, Matched  int
	MatchFraction, MeanDistance, MedianDistance float64
	UniformMean, UniformPercentile              float64
	CircularMean, CircularPercentile            float64
}

type Association struct {
	WindowSize                              int
	Method                                  string
	K                                       int
	Metadata, Subset                        string
	Windows                                 int
	MI, NMI, ARI, Homogeneity, Completeness float64
	ConditionalEntropy, EntropyReduction    float64
	Contingency                             string
}
