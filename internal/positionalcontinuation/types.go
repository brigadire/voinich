// Package positionalcontinuation implements positional-continuation-validate
// (task23): a confirmatory deep-dive on one frozen higher-order finding from
// higher-order-sequence-validate, "s aiin -> chey", asking whether the
// positional concentration that finding showed (POSITION_DEPENDENT) reflects
// a position-conditioned continuation constraint, a general property of
// "aiin", a boundary formula, or an artifact of one physical block. The
// context "s aiin" and target continuation "chey" are frozen inputs - no new
// n-gram discovery ever happens here.
package positionalcontinuation

import (
	"context"
	"io"
	"time"
)

// Frozen inputs (task23 section 2): never re-derived, never re-chosen after
// looking at results, in IVTFF/Voynich mode. These are package-level vars
// rather than consts only so that task43's generic mode (see Config.Generic)
// can substitute a deterministically-selected target triple - the corpus's
// own top-ranked HIGHER_ORDER_REPLICATED candidate from
// higher-order-sequence-validate's generic output (load.go's
// resolveGenericTarget) - in place of this literal Voynichese finding,
// which a generic corpus will never contain. Every function below still
// only ever reads these three tokens by name; none of task23's math changes.
// Set once, before any other work, by RunAndWrite; never mutated again
// within a run.
var (
	FrozenS     = "s"
	FrozenAiin  = "aiin"
	FrozenChey  = "chey"
	FrozenSAiin = "s aiin"
)

// smoothingAlpha is the fixed additive smoothing task23 section 59 requires
// for the M1/M2/M3 model comparison - never optimized.
const smoothingAlpha = 0.5

// Config holds every CLI-controlled input to RunAndWrite.
type Config struct {
	CorpusPath       string
	TokenMetadataMap string
	HigherOrderDir   string
	OutputDir        string
	CheckpointPath   string
	Permutations     int
	Seed             int64
	Quiet            bool
	// Generic selects task43's corpus-only generic mode: Tokens/Blocks come
	// from internal/genericsegmentation instead of a real IVTFF-sourced
	// TokenMetadataMap file, and the frozen target triple (FrozenS/
	// FrozenAiin/FrozenChey, see types.go) is read from HigherOrderDir's own
	// generic-mode top-ranked candidate instead of the literal Voynich
	// finding.
	Generic        bool
	ProgressWriter io.Writer

	// Task44: where each of the 5 distributable batteries executes. Purely
	// operational - excluded from every scientific fingerprint/checkpoint
	// key and never changes a single computed value; see executor.go's
	// runBatteryDispatch for the one dispatch path every backend
	// (goroutine/process/remote) goes through.
	Executor                                                string
	Workers                                                 int
	RemoteListen, TLSCert, TLSKey, ClientCA, RemoteDenyList string
	RemoteTimeout                                           time.Duration
	RemoteRetries                                           int
	BatteryExecutor                                         BatteryExecutor
	Context                                                 context.Context
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
// hand) metadata class - the "physical block" definition used by every
// earlier confirmatory stage in this pipeline.
type Block struct {
	ID      string
	Currier string
	Hand    string
	Joint   string
	Tokens  []Token
}

// SAiinOccurrence is one exact in-block occurrence of "s aiin" together with
// every positional, boundary and surrounding-context fact task23 Parts A-C, N
// and O need.
type SAiinOccurrence struct {
	PosS, PosAiin, PosX int // PosX = -1 if X is missing
	X                   string
	XMissingBlockEnd    bool
	XMissingCorpusEnd   bool

	Block, Currier, Hand, Joint string
	NormalizedBlockPosition     float64 // of aiin within its block
	BlockBinFixed               string  // "B0".."B9"
	BlockBinCoarse              string  // BLOCK_START/MIDDLE/END

	LineID                 string // line of aiin
	NormalizedLinePosition float64
	LineCategory           string // LINE_START/EARLY/MIDDLE/LATE/END
	SIsLineStart           bool   // s is the first token of its line
	XIsLineEnd             bool   // X is the last token of its line (or aiin is last, if X missing)

	TokensBefore [3]string // s-3, s-2, s-1 (relative to s); "" if unavailable
	TokensAfter  [3]string // X+1, X+2, X+3 (relative to X); "" if unavailable

	TokensFromLineStart  int
	TokensToLineEnd      int
	TokensFromBlockStart int
	TokensToBlockEnd     int
}

// AiinOccurrence is the Part H control construction: every in-block
// occurrence of "aiin" alone, whatever precedes it (including nothing, at a
// block start).
type AiinOccurrence struct {
	Predecessor       string // "" if aiin is the first token of its block
	HasPredecessor    bool
	PredecessorIsS    bool
	PosAiin, PosX     int
	X                 string
	XMissingBlockEnd  bool
	XMissingCorpusEnd bool

	Block, Currier, Hand, Joint string
	NormalizedBlockPosition     float64
	BlockBinFixed               string
	BlockBinCoarse              string

	LineID                 string
	NormalizedLinePosition float64
	LineCategory           string
}

// ContinuationRow is one (context, positional stratum, token) cell of a
// continuation distribution (task23 Part D and Part H).
type ContinuationRow struct {
	Context     string // "s_aiin" or "aiin"
	Stratum     string // "global", or a line/block-position category
	StratumType string // "" for global, "line_position" or "block_position"
	Token       string
	Count       int
	Probability float64
}

// DistributionSummaryRow is the one-row-per-stratum summary task23 section 20
// asks for.
type DistributionSummaryRow struct {
	Context             string
	Stratum             string
	StratumType         string
	OccurrenceCount     int
	UniqueContinuations int
	EntropyBits         float64
	NormalizedEntropy   float64
	TopContinuation     string
	TopContinuationProb float64
	CheyProbability     float64
}

// PositionDependenceRow implements task23 Part E's primary test.
type PositionDependenceRow struct {
	PositionVariable string // "line_position" or "block_position_coarse"
	ObservedMIBits   float64
	NullMeanMIBits   float64
	NullSDMIBits     float64
	Permutations     int
	EmpiricalP       float64
}

// PositionalEntropyRow implements task23 Part F.
type PositionalEntropyRow struct {
	PositionVariable           string
	Stratum                    string
	OccurrenceCount            int
	EntropyBits                float64
	EntropyGlobalBits          float64
	EntropyDifference          float64
	EffectiveContinuationCount float64
	UniqueContinuations        int
	EmpiricalP                 float64
	Permutations               int
}

// CheyEffectRow implements task23 Part G.
type CheyEffectRow struct {
	PositionVariable     string
	Stratum              string
	OccurrenceCount      int
	CheyCount            int
	PCheyGivenPosition   float64
	PCheyGlobal          float64
	PositionalEnrichment float64
	EmpiricalP           float64
	Permutations         int
}

// AiinControlRow implements task23 Part H sections 42-47.
type AiinControlRow struct {
	PositionVariable         string
	Stratum                  string
	AiinOccurrenceCount      int
	AiinEntropyBits          float64
	AiinUniqueContinuations  int
	CheyCount                int
	PCheyGivenAiinPosition   float64
	PCheyGivenSAiinPosition  float64
	WithinPositionEnrichment float64
}

// StratifiedPredecessorRow implements task23 Part I sections 48-53.
type StratifiedPredecessorRow struct {
	PositionVariable  string
	ObservedStatistic float64 // pooled count of chey among predecessor==s occurrences, stratified
	NullMeanStatistic float64
	NullSDStatistic   float64
	Permutations      int
	EmpiricalP        float64
}

// ModelLOBORow implements task23 Part J.
type ModelLOBORow struct {
	TestedBlocks     int
	BlocksM2BetterM1 int
	BlocksM1BetterM2 int
	BlocksM3BetterM2 int
	BlocksM2BetterM3 int
	MeanDelta21      float64
	MedianDelta21    float64
	MeanDelta32      float64
	MedianDelta32    float64
}

// CrossBlockPositionalRow implements task23 Part K.
type CrossBlockPositionalRow struct {
	Block                  string
	Currier, Hand, Joint   string
	AiinOccurrences        int
	SAiinOccurrences       int
	CheyGivenAiinPosition  float64
	CheyGivenSAiinPosition float64
	Enrichment             float64
	EffectSign             string // "positive", "negative", "neutral"
}

// PositionalJackknifeRow implements task23 Part L.
type PositionalJackknifeRow struct {
	PositionVariable                                                             string
	Realizations                                                                 int
	MIMin, MIMax, MIMedian, MISD                                                 float64
	EntropyEffectMin, EntropyEffectMax, EntropyEffectMedian, EntropyEffectSD     float64
	CheyEnrichmentMin, CheyEnrichmentMax, CheyEnrichmentMedian, CheyEnrichmentSD float64
	StratifiedSMin, StratifiedSMax, StratifiedSMedian, StratifiedSSD             float64
	SingleBlockSensitive                                                         bool
}

// LineVsBlockRow implements task23 Part M.
type LineVsBlockRow struct {
	Analysis          string // "association", "line_controlling_block", "block_controlling_line"
	LineCategory      string
	BlockCoarseBucket string
	OccurrenceCount   int
	CheyProbability   float64
}

// BoundaryDistanceRow implements task23 Part N.
type BoundaryDistanceRow struct {
	Group            string // "s_aiin_chey" or "s_aiin_all"
	Metric           string // "tokens_from_line_start", etc.
	Median, Q25, Q75 float64
	PermutationP     float64
}

// SurroundingContextRow implements task23 Part O.
type SurroundingContextRow struct {
	Group                     string // "chey" or "not_chey"
	OccurrenceCount           int
	PrecedingEntropyBits      float64
	FollowingEntropyBits      float64
	UniqueSurroundingContexts int
}

// ReversePositionRow implements task23 Part P.
type ReversePositionRow struct {
	PositionVariable string
	Stratum          string
	PGivenSAiin      float64
	PGivenAiin       float64
	TotalVariation   float64
}

// ValidationRow implements task23 Part Q's final diagnostic classification.
type ValidationRow struct {
	FinalStatus               string
	PositionDependenceP       float64
	PositionDependenceSig     bool
	CheyEnrichmentSig         bool
	StratifiedPredecessorSig  bool
	M3BetterThanM2Fraction    float64
	CrossBlockSignConsistency float64
	SingleBlockSensitive      bool
	BoundaryFormulaSupported  bool
	EligibleBlocks            int
}
