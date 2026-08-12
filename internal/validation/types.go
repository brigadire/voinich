package validation

import "zcore.dev/voinich/internal/normalization"

type Config struct {
	InputPath       string
	ClassesPath     string
	Folds           int
	FoldSeed        int64
	Threshold       float64
	MinTokenCount   int
	RandomBaselines int
	RandomSeed      int64
	MinN            int
	MaxN            int
	MaxContext      int
	Progress        func(string)
}

type Line struct {
	ID     int
	Tokens []string
}

type Corpus struct {
	Lines       []Line
	Counts      map[string]int
	Occurrences int
	Transitions int
}

type Parameters struct {
	Folds                int     `yaml:"folds"`
	FoldSeed             int64   `yaml:"fold_seed"`
	Threshold            float64 `yaml:"threshold"`
	MinTokenCount        int     `yaml:"min_token_count"`
	RandomBaselines      int     `yaml:"random_baselines"`
	RandomSeed           int64   `yaml:"random_seed"`
	MinN                 int     `yaml:"min_n"`
	MaxN                 int     `yaml:"max_n"`
	MaxContext           int     `yaml:"max_context_length"`
	LeaveOneOutThreshold float64 `yaml:"leave_one_class_out_threshold"`
}

type Meta struct {
	Input                   string `yaml:"input"`
	FullCorpusClasses       string `yaml:"full_corpus_classes"`
	PhysicalLines           int    `yaml:"physical_lines"`
	NonEmptyLines           int    `yaml:"non_empty_lines"`
	TokenOccurrences        int    `yaml:"token_occurrences"`
	Transitions             int    `yaml:"transitions"`
	CorpusSHA256            string `yaml:"corpus_sha256"`
	FullCorpusClassesSHA256 string `yaml:"full_corpus_classes_sha256"`
}

type Methodology struct {
	SplitUnit            string `yaml:"split_unit"`
	Split                string `yaml:"split"`
	Training             string `yaml:"training"`
	Eligibility          string `yaml:"eligibility"`
	Clustering           string `yaml:"clustering"`
	TestApplication      string `yaml:"test_application"`
	SequenceBoundary     string `yaml:"sequence_boundary"`
	CrossLineRepeated    string `yaml:"cross_line_repeated"`
	ConditionalEntropy   string `yaml:"conditional_entropy"`
	RandomMatching       string `yaml:"random_matching"`
	RandomSeedDerivation string `yaml:"random_seed_derivation"`
	EmpiricalTests       string `yaml:"empirical_tests"`
	ClassStability       string `yaml:"class_stability"`
	Pooled               string `yaml:"pooled_test"`
	ThresholdSelection   string `yaml:"threshold_selection"`
}

type PartitionStats struct {
	PhysicalLines    int    `yaml:"physical_lines"`
	NonEmptyLines    int    `yaml:"non_empty_lines"`
	TokenOccurrences int    `yaml:"token_occurrences"`
	Transitions      int    `yaml:"transitions"`
	LineIDsSHA256    string `yaml:"line_ids_sha256"`
	LineIDs          []int  `yaml:"line_ids,omitempty"`
}

type TrainClasses struct {
	EligibleTokens     int                   `yaml:"eligible_tokens"`
	MultiMemberClasses int                   `yaml:"multi_member_classes"`
	TokensInClasses    int                   `yaml:"tokens_in_classes"`
	OccurrenceCoverage float64               `yaml:"occurrence_coverage"`
	Classes            []normalization.Class `yaml:"classes"`
}

type TestNormalization struct {
	OccurrencesCovered int     `yaml:"occurrences_covered"`
	OccurrenceCoverage float64 `yaml:"occurrence_coverage"`
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

type RandomComparison struct {
	StructuralValue float64            `yaml:"structural_value"`
	Direction       string             `yaml:"empirical_test_direction"`
	Random          RandomDistribution `yaml:"random"`
	EmpiricalP      float64            `yaml:"empirical_p"`
}

type NGramComparison struct {
	N                           int              `yaml:"n"`
	RawCrossLineRepeated        int              `yaml:"raw_cross_line_repeated"`
	StructuralCrossLineRepeated int              `yaml:"structural_cross_line_repeated"`
	AbsoluteDelta               int              `yaml:"absolute_delta"`
	Ratio                       *float64         `yaml:"ratio"`
	Random                      RandomComparison `yaml:"random_baseline"`
}

type ContextMetrics struct {
	ContextLength                     int     `yaml:"context_length"`
	ConditionalEntropy                float64 `yaml:"conditional_entropy"`
	RepeatedContextConditionalEntropy float64 `yaml:"repeated_context_conditional_entropy"`
	RepeatedContextCoverage           float64 `yaml:"repeated_context_coverage"`
}

type ContextComparison struct {
	ContextLength                        int              `yaml:"context_length"`
	RawRepeatedContextCoverage           float64          `yaml:"raw_repeated_context_coverage"`
	StructuralRepeatedContextCoverage    float64          `yaml:"structural_repeated_context_coverage"`
	CoverageDelta                        float64          `yaml:"coverage_delta"`
	RawConditionalEntropy                float64          `yaml:"raw_conditional_entropy"`
	StructuralConditionalEntropy         float64          `yaml:"structural_conditional_entropy"`
	EntropyDelta                         float64          `yaml:"entropy_delta"`
	RawRepeatedConditionalEntropy        float64          `yaml:"raw_repeated_context_conditional_entropy"`
	StructuralRepeatedConditionalEntropy float64          `yaml:"structural_repeated_context_conditional_entropy"`
	RepeatedEntropyDelta                 float64          `yaml:"repeated_entropy_delta"`
	CoverageRandom                       RandomComparison `yaml:"coverage_random_baseline"`
	EntropyReductionRandom               RandomComparison `yaml:"entropy_reduction_random_baseline"`
	RepeatedEntropyReductionRandom       RandomComparison `yaml:"repeated_entropy_reduction_random_baseline"`
}

type SequenceComparison struct {
	NGrams                       []NGramComparison   `yaml:"cross_line_ngrams"`
	RawMaxCrossLineLength        int                 `yaml:"raw_max_cross_line_sequence_length"`
	StructuralMaxCrossLineLength int                 `yaml:"structural_max_cross_line_sequence_length"`
	MaxLengthRandom              RandomComparison    `yaml:"max_cross_line_sequence_length_random_baseline"`
	ContextOrder                 []ContextComparison `yaml:"context_order"`
}

type Coordinate struct {
	LineID      int `yaml:"line_id"`
	TokenOffset int `yaml:"token_offset"`
}

type SurfaceRealization struct {
	LineID      int      `yaml:"line_id"`
	TokenOffset int      `yaml:"token_offset"`
	RawTokens   []string `yaml:"raw_tokens"`
}

type NewCrossLineSequence struct {
	NormalizedTokens []string             `yaml:"normalized_tokens"`
	N                int                  `yaml:"n"`
	Count            int                  `yaml:"count"`
	LineCount        int                  `yaml:"line_count"`
	Occurrences      []SurfaceRealization `yaml:"occurrences"`
}

type FoldResult struct {
	Fold                  int                    `yaml:"fold"`
	Train                 PartitionStats         `yaml:"train"`
	Test                  PartitionStats         `yaml:"test"`
	StructuralClasses     TrainClasses           `yaml:"structural_classes"`
	TestNormalization     TestNormalization      `yaml:"test_normalization"`
	SequenceComparison    SequenceComparison     `yaml:"sequence_comparison"`
	NewCrossLineSequences []NewCrossLineSequence `yaml:"new_cross_line_sequences"`
}

type StabilityPair struct {
	TokenA            string  `yaml:"token_a"`
	TokenB            string  `yaml:"token_b"`
	FoldsBothEligible int     `yaml:"folds_both_eligible"`
	FoldsSameClass    int     `yaml:"folds_same_class"`
	Stability         float64 `yaml:"stability"`
}

type ClassStability struct {
	ReportedPairRule      string          `yaml:"reported_pair_rule"`
	StablePairs100Percent int             `yaml:"stable_pairs_100_percent"`
	StablePairs80Percent  int             `yaml:"stable_pairs_80_percent"`
	UnstablePairs         int             `yaml:"unstable_pairs"`
	Pairs                 []StabilityPair `yaml:"pairs"`
}

type AggregateNGram struct {
	N              int     `yaml:"n"`
	FoldsPositive  int     `yaml:"folds_positive"`
	FoldsZero      int     `yaml:"folds_zero"`
	FoldsNegative  int     `yaml:"folds_negative"`
	MeanRaw        float64 `yaml:"mean_raw"`
	MeanStructural float64 `yaml:"mean_structural"`
	MeanDelta      float64 `yaml:"mean_delta"`
	MedianDelta    float64 `yaml:"median_delta"`
}

type PooledNGram struct {
	N                           int `yaml:"n"`
	RawCrossLineRepeated        int `yaml:"raw_cross_line_repeated"`
	StructuralCrossLineRepeated int `yaml:"structural_cross_line_repeated"`
	AbsoluteDelta               int `yaml:"absolute_delta"`
}

type CrossValidationAggregate struct {
	CrossLineNGrams []AggregateNGram `yaml:"cross_line_ngrams"`
	PooledTest      []PooledNGram    `yaml:"pooled_test"`
}

type SimpleNGram struct {
	N                 int `yaml:"n"`
	CrossLineRepeated int `yaml:"cross_line_repeated"`
}

type LeaveOneOutVariant struct {
	ClassRemoved            string           `yaml:"class_removed"`
	ClassMembers            []string         `yaml:"class_members"`
	ClassOccurrenceCoverage float64          `yaml:"class_occurrence_coverage"`
	CrossLineNGrams         []SimpleNGram    `yaml:"cross_line_ngrams"`
	MaxCrossLineLength      int              `yaml:"max_cross_line_sequence_length"`
	RepeatedContextCoverage []ContextMetrics `yaml:"context_order"`
	ContributionN3          int              `yaml:"contribution_n3"`
	ContributionFractionN3  *float64         `yaml:"contribution_fraction_n3"`
	ContributionN4          int              `yaml:"contribution_n4"`
	ContributionFractionN4  *float64         `yaml:"contribution_fraction_n4"`
}

type LeaveOneClassOut struct {
	Raw                     []SimpleNGram        `yaml:"raw_cross_line_ngrams"`
	AllClasses              []SimpleNGram        `yaml:"all_classes_cross_line_ngrams"`
	AllClassesMaxLength     int                  `yaml:"all_classes_max_cross_line_sequence_length"`
	Variants                []LeaveOneOutVariant `yaml:"variants"`
	DominantClassFractionN3 *float64             `yaml:"dominant_class_fraction_n3"`
	DominantClassFractionN4 *float64             `yaml:"dominant_class_fraction_n4"`
}

type MemberAblation struct {
	ClassID                    string        `yaml:"class_id"`
	OriginalMembers            []string      `yaml:"original_members"`
	MemberRestored             string        `yaml:"member_restored_to_surface"`
	RemainingNormalizedMembers []string      `yaml:"remaining_normalized_members"`
	CrossLineNGrams            []SimpleNGram `yaml:"cross_line_ngrams"`
}

type Output struct {
	Meta                     Meta                     `yaml:"meta"`
	Parameters               Parameters               `yaml:"parameters"`
	Methodology              Methodology              `yaml:"methodology"`
	Folds                    []FoldResult             `yaml:"folds"`
	ClassStability           ClassStability           `yaml:"class_stability"`
	CrossValidationAggregate CrossValidationAggregate `yaml:"cross_validation_aggregate"`
	LeaveOneClassOut         LeaveOneClassOut         `yaml:"leave_one_class_out"`
	MemberAblation           []MemberAblation         `yaml:"member_ablation"`
}
