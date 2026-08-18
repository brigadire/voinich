// Package transitionnetwork implements the frozen adjacent-transition network
// validation specified by task24.
package transitionnetwork

import (
	"context"
	"io"
	"time"
)

type Config struct {
	CorpusPath, MetadataPath, OutputDir, CheckpointPath                 string
	MinTokenCount, MinBlockTokenCount, Permutations, RefinePermutations int
	Seed                                                                int64
	Quiet                                                               bool
	// Generic selects task43's corpus-only generic mode: Tokens/Blocks are
	// derived from internal/genericsegmentation instead of a real
	// IVTFF-sourced MetadataPath file, and computeGraphDiagnostics never
	// runs its Currier/hand metadata-transfer comparison in that mode (see
	// load.go/analyze.go and GENERIC_STAGE_APPLICABILITY_AUDIT.md).
	Generic bool

	// Task44: where each of the primary/refine permutation-null batteries'
	// replicates execute. Purely operational - excluded from every
	// scientific fingerprint/checkpoint key - and never changes a single
	// computed value; see executor.go's runBattery for the one reduction
	// path every backend (goroutine/process/remote) goes through.
	Executor                                                string
	Workers                                                 int
	RemoteListen, TLSCert, TLSKey, ClientCA, RemoteDenyList string
	RemoteTimeout                                           time.Duration
	RemoteRetries                                           int
	PermutationExecutor                                     PermutationExecutor
	Context                                                 context.Context

	ProgressWriter io.Writer
}

type Token struct {
	Position                   int
	Text, Currier, Hand, Joint string
}
type Block struct {
	ID, Currier, Hand, Joint string
	Tokens                   []Token
}
type EdgeKey struct{ Source, Target string }

func (e EdgeKey) String() string { return e.Source + "\x00" + e.Target }

type BlockStats struct {
	Block, Currier, Hand, Joint, Source, Target                     string
	SourceCount, TargetCount, EdgeCount, Opportunities, BlockTokens int
	PConditional, PBaseline, Enrichment, Log2Enrichment             float64
}

type EdgeSummary struct {
	EdgeKey
	GlobalCount, EligibleBlocks, PositiveBlocks, NegativeBlocks, NeutralBlocks int
	JointClasses, CurrierClasses, Hands                                        int
	MedianLog2, MeanLog2, BetweenBlockSD, SignConsistency                      float64
	ExpectedSign                                                               string
	TestedBlocks, SuccessfulSignPredictions                                    int
	TransferFraction, EmpiricalP, FDRQ                                         float64
	Permutations                                                               int
	MaxBlockObservationFraction, MaxBlockEffectWeightFraction                  float64
	Status                                                                     string
}

type ProfileStability struct {
	Token, Direction                                                              string
	GlobalCount, EligibleBlocks, JointClasses                                     int
	PairwiseJSMean, PairwiseJSMedian, PairwiseJSMin, PairwiseJSSD                 float64
	LOBOMedianCorrelation, LOBOMeanCorrelation, LOBOMedianSpearman, SignAgreement float64
	PermutationP, SignPermutationP                                                float64
	EntropyEffect, EntropySignConsistency, EntropyPermutationP                    float64
	EntropyStatus                                                                 string
	Replicated                                                                    bool
}

type EntropyRow struct {
	Token, Block, Direction                                            string
	ConditionalEntropy, EffectiveCount, BaselineEntropy, EntropyEffect float64
	Eligible                                                           bool
}

type PredictionRow struct {
	Block, Scope          string
	N                     int
	LossM0, LossM1, Delta float64
}
type ModelOrderRow struct {
	Block                 string
	N                     int
	LossM1, LossM2, Delta float64
}
type TransferRow struct {
	Dimension, GroupA, GroupB                           string
	CommonEdges                                         int
	SignAgreement, EffectCorrelation, ProfileSimilarity float64
}
type GraphSimilarityRow struct {
	BlockA, BlockB                                 string
	EdgesA, EdgesB, Intersection                   int
	EdgeJaccard, DegreeRankCorrelation, SCCOverlap float64
}
