// Package normalizationcompare holds the normalization-compare CLI's
// scientific core, split out of the main package (Task42) so it can be
// shared between the in-process default execution path and the same
// mTLS distributed executor Task33-40 built for conditionalregime and
// structuralprojection. Every exported type/function here is byte-for-byte
// the code normalization-compare/main.go ran directly before Task42; moving
// it does not change any formula, threshold, or RNG semantics.
package normalizationcompare

import (
	"context"

	"zcore.dev/voinich/internal/normalization"
)

type SequenceSummary struct {
	N                 int `yaml:"n"`
	MultiLineRepeated int `yaml:"multi_line_repeated"`
}

type ContextOrder struct {
	ContextLength                     int     `yaml:"context_length"`
	ConditionalEntropy                float64 `yaml:"conditional_entropy"`
	RepeatedContextConditionalEntropy float64 `yaml:"repeated_context_conditional_entropy"`
	RepeatedContextCoverage           float64 `yaml:"repeated_context_coverage"`
}

// SequenceMeta is the corpus-invariant triple every model's (structural and
// random) sequence analysis must agree on with the immutable raw analysis -
// the existing "corpus invariants changed" guard, now also the value a
// distributed baseline trial reports back for the coordinator to check.
type SequenceMeta struct {
	TokenOccurrences int `yaml:"token_occurrences"`
	Lines            int `yaml:"lines"`
	Transitions      int `yaml:"transitions"`
}

type SequenceAnalysis struct {
	Meta                 SequenceMeta      `yaml:"meta"`
	NGramSummary         []SequenceSummary `yaml:"ngram_summary"`
	ContextOrderAnalysis []ContextOrder    `yaml:"context_order_analysis"`
}

type RandomDistribution struct {
	Runs         int     `yaml:"runs"`
	Mean         float64 `yaml:"mean"`
	Stddev       float64 `yaml:"stddev"`
	Min          float64 `yaml:"min"`
	Max          float64 `yaml:"max"`
	Percentile05 float64 `yaml:"percentile_05"`
	Percentile50 float64 `yaml:"percentile_50"`
	Percentile95 float64 `yaml:"percentile_95"`
}

type Effect struct {
	RawValue        float64            `yaml:"raw_value"`
	StructuralValue float64            `yaml:"structural_value"`
	AbsoluteDelta   float64            `yaml:"absolute_delta"`
	Ratio           float64            `yaml:"ratio"`
	Direction       string             `yaml:"empirical_test_direction"`
	Random          RandomDistribution `yaml:"random"`
	ZScore          float64            `yaml:"z_score"`
	EmpiricalP      float64            `yaml:"empirical_p"`
}

type NGramComparison struct {
	N                 int    `yaml:"n"`
	CrossLineRepeated Effect `yaml:"cross_line_repeated"`
}

type ContextComparison struct {
	ContextLength              int    `yaml:"context_length"`
	ConditionalEntropy         Effect `yaml:"conditional_entropy"`
	RepeatedConditionalEntropy Effect `yaml:"repeated_context_conditional_entropy"`
	RepeatedContextCoverage    Effect `yaml:"repeated_context_coverage"`
}

type ModelComparison struct {
	Threshold                  float64                  `yaml:"threshold"`
	Label                      string                   `yaml:"label"`
	Normalization              normalization.ModelStats `yaml:"normalization"`
	MaxCrossLineSequenceLength Effect                   `yaml:"max_cross_line_sequence_length"`
	NGrams                     []NGramComparison        `yaml:"ngrams"`
	ContextOrder               []ContextComparison      `yaml:"context_order"`
}

type ComparisonOutput struct {
	Meta struct {
		RandomBaselines  int    `yaml:"random_baselines"`
		RandomSeed       int64  `yaml:"random_seed"`
		SequenceAnalyzer string `yaml:"sequence_analyzer"`
		RandomMatching   string `yaml:"random_matching"`
		EmpiricalTests   string `yaml:"empirical_tests"`
	} `yaml:"meta"`
	Models []ModelComparison `yaml:"models"`
}

// Metrics is the subset of a SequenceAnalysis this tool compares across the
// raw, structural and random-baseline analyses.
type Metrics struct {
	CrossLine map[int]float64
	MaxLength float64
	Contexts  map[int]ContextOrder
}

// BaselineResult is one random-baseline trial's complete, order-independent
// contribution. Its scientific identity is (threshold label, run index)
// alone (see JobID in the distributed executor) - it carries no worker
// identity, timestamp, or arrival-order information.
type BaselineResult struct {
	Metrics Metrics
	Meta    SequenceMeta
}

// BaselineExecutor runs independent random-baseline trials. Task42 mirrors
// structuralprojection.TrialExecutor: only where a trial executes changes
// between goroutine/process/remote backends, never what it computes or in
// what order the coordinator reduces results.
type BaselineExecutor interface {
	Run(ctx context.Context, threshold string, run int) (BaselineResult, error)
	Close() error
}

// BaselineExecutorStats optionally reports live progress (active
// workers/leases, reclaimed leases) exactly as structuralprojection's
// TrialExecutorStats does.
type BaselineExecutorStats interface {
	BaselineStats() (active, retries int)
}
