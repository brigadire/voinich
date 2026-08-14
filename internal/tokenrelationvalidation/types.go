package tokenrelationvalidation

import "io"

type Config struct {
	CorpusPath, MetadataPath, DiscoveryDir, OutputDir, CheckpointPath string
	Permutations, RefinePermutations                                  int
	Seed                                                              int64
	Quiet                                                             bool
	ProgressWriter                                                    io.Writer
}

type Token struct {
	Position                                  int
	Text, Line, Currier, Hand, Joint, BlockID string
	LineIndex                                 int
}

type Block struct {
	ID, Currier, Hand, Joint string
	Start, End               int
	Tokens                   []Token
}

type Pair struct{ A, B string }

type Candidate struct {
	ID, Family, A, B, Sequence, Sources string
	Directed                            bool
	FrozenThreshold                     float64
	StoredTokenCount                    int
}

type InventoryFile struct {
	Path, SHA256     string
	StoredTokenCount int
}

type DirectionBlock struct {
	CandidateID, A, B, BlockID, Currier, Hand, Joint                           string
	CountA, CountB, ABeforeB, BBeforeA, ImmediateAB, ImmediateBA, Observations int
	ExactAB, ExactBA                                                           [5]int
	Score, EnrichmentAB, EnrichmentBA                                          float64
	Eligible                                                                   bool
}

type RelationSummary struct {
	CandidateID, Family, A, B, Sequence, Classification                           string
	EligibleBlocks, PositiveBlocks, NegativeBlocks, NeutralBlocks                 int
	PhysicalBlocks, JointClasses, CurrierClasses, Hands                           int
	SignConsistency, WeightedDirection, UnweightedDirection, BetweenBlockVariance float64
	MedianEnrichment, ProfileMean, ProfileMedian, ProfileMin, ProfileSD           float64
	FractionAboveThreshold, ProfileOverlapMean                                    float64
	TransferSuccess                                                               float64
	TestedHeldout, SuccessfulHeldout                                              int
	RawP, FDRQ, ControlPercentile                                                 float64
}

type ProfileBlock struct {
	CandidateID, Family, A, B, BlockID, Currier, Hand, Joint               string
	CountA, CountB                                                         int
	Position, Left, Right, Distance, Similarity, Overlap, GlobalSimilarity float64
	TrainingReference                                                      float64
	PooledSimilarity                                                       float64
	EligiblePrimary, EligibleDescriptive                                   bool
}

type SequenceResult struct {
	CandidateID, Sequence                                      string
	Total, PhysicalBlocks, JointClasses, CurrierClasses, Hands int
	MaxBlockFraction                                           float64
	HighRecurrence                                             bool
	RawP, FDRQ                                                 float64
}

type Transfer struct {
	CandidateID, Family, HeldoutBlock, TrainMetadata, HeldoutMetadata string
	Expected, Observed                                                float64
	Success                                                           bool
}

type MetadataTransfer struct {
	CandidateID, Family, Dimension, Training, Heldout string
	Tested, Successful                                int
	Fraction                                          float64
}

type Control struct {
	CandidateID, Family, Kind            string
	ControlA, ControlB                   string
	Observed, NullMean, RawP, Percentile float64
	Permutations                         int
}

type Analysis struct {
	CorpusSHA, MetadataSHA    string
	TokenCount, UnknownTokens int
	Files                     []InventoryFile
	Candidates                []Candidate
	Blocks                    []Block
	DirectionBlocks           []DirectionBlock
	ProfileBlocks             []ProfileBlock
	Sequences                 []SequenceResult
	Summaries                 []RelationSummary
	Transfers                 []Transfer
	MetadataTransfers         []MetadataTransfer
	Controls                  []Control
	DistancePairwise          map[string][]float64
	DistancePairwiseOverlap   map[string][]float64
	Parameters                Config
}
