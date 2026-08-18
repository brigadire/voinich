// Package higherorderseq implements higher-order-sequence-validate (task22):
// it tests, for a small frozen set of n>=3 sequences ABC already established
// as replicated by the previous audit, whether the first token A carries
// additional information about the third token C once the second token B is
// already known - i.e. whether P(C|A,B) departs from P(C|B) in a way that
// reproduces across independent physical blocks and survives leave-one-block-
// out, jackknife, and a conditional-neighbor permutation null. No new
// sequence discovery happens here: every ABC tested is read programmatically
// from the frozen outputs of replicated-local-structure-audit.
package higherorderseq

import (
	"context"
	"io"
	"time"
)

// Config holds every CLI-controlled input to RunAndWrite.
type Config struct {
	CorpusPath       string
	TokenMetadataMap string
	AuditDir         string
	DiscoveryDir     string
	OutputDir        string
	CheckpointPath   string
	Permutations     int // primary-family permutation count (task22 section 67)
	Seed             int64
	Quiet            bool
	// Generic selects task43's corpus-only generic mode: Tokens/Blocks are
	// derived from internal/genericsegmentation instead of a real
	// IVTFF-sourced TokenMetadataMap file. AuditDir must then hold stage 24's
	// own generic-mode frozen candidates (see load.go/GENERIC_STAGE_
	// APPLICABILITY_AUDIT.md); DistinctJoint>=2 in that mode means "replicates
	// across >=2 independent generic resampling folds", not "independent
	// hands/Currier states".
	Generic        bool
	ProgressWriter io.Writer

	// Task44: where each frozen candidate's whole Part A-L computation
	// executes. Purely operational - excluded from every scientific
	// fingerprint/checkpoint key and never changes a single computed value;
	// see executor.go's runCandidateBattery for the one dispatch path every
	// backend (goroutine/process/remote) goes through.
	Executor                                                string
	Workers                                                 int
	RemoteListen, TLSCert, TLSKey, ClientCA, RemoteDenyList string
	RemoteTimeout                                           time.Duration
	RemoteRetries                                           int
	CandidateExecutor                                       CandidateExecutor
	Context                                                 context.Context
}

// secondaryPermutations returns the fixed 1/10th ratio task22 section 68
// specifies for the single secondary descriptive candidate.
func secondaryPermutations(primary int) int {
	if primary >= 10 {
		return primary / 10
	}
	return primary
}

// Token is one corpus position together with the metadata it needs for
// block/line boundary and position bookkeeping.
type Token struct {
	Position       int
	Text           string
	Line           string
	Currier        string
	Hand           string
	Joint          string
	TokenIndexLine int
}

// Block is a maximal contiguous run of tokens sharing one joint (Currier x
// hand) metadata class - the same "physical block" definition used by every
// earlier confirmatory stage in this pipeline. Tokens with unknown metadata
// are excluded entirely rather than stitched across.
type Block struct {
	ID      string
	Currier string
	Hand    string
	Joint   string
	Tokens  []Token
}

// Candidate is one frozen ABC sequence together with the previous audit's
// own p/q values, read verbatim rather than recomputed.
type Candidate struct {
	Sequence             string
	Tokens               []string // [A, B, C]
	Family               string   // "primary" or "secondary"
	CanonicalOccurrences int
	PhysicalBlocks       int
	JointClasses         int
	ShuffleFDRQ          float64
	MarkovBlockP         float64
}

func (c Candidate) A() string { return c.Tokens[0] }
func (c Candidate) B() string { return c.Tokens[1] }
func (c Candidate) C() string { return c.Tokens[2] }

// Occurrence is one exact in-block match of a candidate's A B C tokens.
type Occurrence struct {
	Sequence             string
	PosA, PosB, PosC     int
	Block                string
	Currier, Hand, Joint string
	NormalizedBlockPos   float64
	WithinSameLine       bool
	CrossesLineBoundary  bool
	LinePosition         string // "start", "middle", "end", or "" if unavailable
}

// BlockCounts holds the four raw counts task22 Part B needs for one
// candidate within one eligible physical block.
type BlockCounts struct {
	Sequence                           string
	Block, Currier, Hand, Joint        string
	CountB, CountAB, CountBC, CountABC int
}

type ConditionalRow struct {
	BlockCounts
	PCGivenB, PCGivenAB, Enrichment, DeltaProbability               float64
	EligiblePrimary, EligibleDescriptive                            bool
	PAGivenB, PAGivenBC, ReverseEnrichment, ReverseDeltaProbability float64
}

type CMIResult struct {
	Sequence         string
	CenterToken      string
	Occurrences      int
	ObservedCMIBits  float64
	NullMeanCMIBits  float64
	NullSDCMIBits    float64
	Permutations     int
	EmpiricalP       float64
	ContributionBits float64 // pointwise contribution of the frozen (A,C) cell
}

type DependenceRow struct {
	Sequence     string
	Family       string
	Permutations int
	EmpiricalP   float64
	FDRQ         float64
	Significant  bool
}

type LOBORow struct {
	Sequence                  string
	TestedBlocks              int
	M2BetterBlocks            int
	M1BetterBlocks            int
	Ties                      int
	MeanDeltaLogLoss          float64
	MedianDeltaLogLoss        float64
	HeldoutLogLikelihoodRatio float64
}

type ContextControlRow struct {
	Sequence    string
	ContextType string // "left_alt" or "right_alt"
	AltToken    string
	Count       int
	Probability float64
	IsFrozen    bool
}

type ContextRankRow struct {
	Sequence                     string
	NumAlternatives              int
	FrozenP                      float64
	BaselineP                    float64
	Rank                         int
	Percentile                   float64
	MinAltP, MedianAltP, MaxAltP float64
}

type ContinuationRow struct {
	Sequence    string
	Context     string // "B" or "AB"
	Token       string
	Count       int
	Probability float64
}

type ContinuationEntropyRow struct {
	Sequence         string
	HGivenB          float64
	HGivenAB         float64
	EntropyReduction float64
	JSDivergence     float64
	TotalVariation   float64
}

type CrossBlockRow struct {
	Sequence                 string
	EligibleBlocks           int
	PositiveEnrichmentBlocks int
	NegativeEnrichmentBlocks int
	SignConsistency          float64
	DistinctCurrier          int
	DistinctHand             int
	DistinctJoint            int
	CrossCurrier             bool
	CrossHand                bool
	CrossJoint               bool
}

type MetaAnalysisRow struct {
	Sequence                    string
	Blocks                      int
	UnweightedMeanLogEnrichment float64
	WeightedMeanLogEnrichment   float64
	MedianLogEnrichment         float64
	BetweenBlockVariance        float64
	CochranQ                    float64
	I2                          float64
	MaxBlockWeightFraction      float64
}

type JackknifeRow struct {
	Sequence                                                             string
	Realizations                                                         int
	EnrichmentMin, EnrichmentMax, EnrichmentMedian, EnrichmentSD         float64
	CMIMin, CMIMax, CMIMedian, CMISD                                     float64
	DeltaLogLossMin, DeltaLogLossMax, DeltaLogLossMedian, DeltaLogLossSD float64
	SingleBlockSensitive                                                 bool
}

type PositionRow struct {
	Sequence    string
	Metric      string // "block_position_bin" or "line_position"
	Bucket      string
	ABCCount    int
	ABCount     int
	ABCFraction float64
	ABFraction  float64
}

type StructuralFamilyRow struct {
	Sequence   string
	TokenRole  string // "A" or "C"
	Token      string
	Relative   string
	Sufficient bool
	FrozenP    float64
	RelativeP  float64
	SignHolds  bool
}

type ValidationRow struct {
	Sequence              string
	Family                string
	FinalStatus           string
	ConditionalFDRQ       float64
	EligibleBlocks        int
	SignConsistency       float64
	LOBOAdvantageFraction float64
	SingleBlockSensitive  bool
	DistinctJointClasses  int
	PositionDependent     bool
	MetadataLimited       bool
}

// CandidateResult bundles every Part A-P output computed for one frozen
// candidate. It is the unit both the checkpoint and the final writers work
// from.
type CandidateResult struct {
	Candidate        Candidate
	Occurrences      []Occurrence
	ConditionalRows  []ConditionalRow
	CMI              CMIResult
	LOBO             LOBORow
	ContextControls  []ContextControlRow
	ContextRank      ContextRankRow
	Continuations    []ContinuationRow
	ContinuationEnt  ContinuationEntropyRow
	CrossBlock       CrossBlockRow
	Meta             MetaAnalysisRow
	Jackknife        JackknifeRow
	Position         []PositionRow
	StructuralFamily []StructuralFamilyRow
	BlockPosTVD      float64
	LinePosTVD       float64
}
