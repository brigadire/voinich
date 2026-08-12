package structuralreliability

// Config holds every CLI-controlled parameter. All statistical formulas are
// reused from internal/profilestability and internal/validation; this
// package only measures how reproducible their outputs are as a function of
// how many observations back a token or pair.
type Config struct {
	InputPath             string
	ClassesPath           string
	Folds                 int
	FoldSeed              int64
	MinTokenCount         int
	Neighbors             int
	BootstrapRuns         int
	BootstrapSeed         int64
	Threshold             float64
	ThresholdMargin       float64
	CountThresholds       []int
	SubsampleMinFullCount int
	SubsampleRuns         int
	SubsampleSeed         int64
	Progress              func(string)
}

// Stat is a compact mean/median/stddev/observations summary, reused across
// every aggregate in this package instead of a bespoke type per section.
type Stat struct {
	Observations int     `yaml:"observations"`
	Mean         float64 `yaml:"mean"`
	Median       float64 `yaml:"median"`
	Stddev       float64 `yaml:"stddev"`
}

// PercentileStat additionally carries tail percentiles, used where the task
// asks for percentile_90/percentile_95 rather than just mean/median/stddev.
type PercentileStat struct {
	Observations int     `yaml:"observations"`
	Mean         float64 `yaml:"mean"`
	Median       float64 `yaml:"median"`
	Stddev       float64 `yaml:"stddev"`
	Percentile90 float64 `yaml:"percentile_90"`
	Percentile95 float64 `yaml:"percentile_95"`
}

type SelfProfileStability struct {
	Position     Stat `yaml:"position_similarity"`
	LeftContext  Stat `yaml:"left_context_similarity"`
	RightContext Stat `yaml:"right_context_similarity"`
}

type TrainTestStability struct {
	Position     Stat `yaml:"position_similarity"`
	LeftContext  Stat `yaml:"left_context_similarity"`
	RightContext Stat `yaml:"right_context_similarity"`
}

type NearestNeighborStability struct {
	MeanTop1Recovery   float64 `yaml:"mean_top1_recovery"`
	MedianTop1Recovery float64 `yaml:"median_top1_recovery"`
	MeanTop3Overlap    float64 `yaml:"mean_top3_overlap"`
	MeanTop5Overlap    float64 `yaml:"mean_top5_overlap"`
	MeanTop10Jaccard   float64 `yaml:"mean_top10_jaccard"`
}

type PairStabilitySummary struct {
	Pairs                                     int            `yaml:"pairs"`
	SimilarityStddev                          PercentileStat `yaml:"similarity_stddev"`
	ThresholdCrossing070Fraction              float64        `yaml:"threshold_crossing_070_fraction"`
	BootstrapProbabilityAbove070Ge095Fraction float64        `yaml:"bootstrap_probability_above_070_ge_095_fraction"`
	BootstrapPairsAvailable                   int            `yaml:"bootstrap_pairs_available"`
}

type CIWidthSummary struct {
	MeanCIWidth         float64 `yaml:"mean_ci_width"`
	MedianCIWidth       float64 `yaml:"median_ci_width"`
	Percentile90CIWidth float64 `yaml:"percentile_90_ci_width"`
	Observations        int     `yaml:"observations"`
}

// CumulativeThreshold is one row of section 3/5/6/7/8: every token/pair with
// count >= MinCount, using the same eligibility rule structural-profile-
// stability itself would use if invoked with -min-token-count=MinCount.
type CumulativeThreshold struct {
	MinCount              int                      `yaml:"min_count"`
	EligibleTokens        int                      `yaml:"eligible_tokens"`
	SelfProfileTrainTrain SelfProfileStability     `yaml:"self_profile_train_train"`
	TrainTest             TrainTestStability       `yaml:"train_test"`
	NearestNeighbors      NearestNeighborStability `yaml:"nearest_neighbor_stability"`
	Pairs                 PairStabilitySummary     `yaml:"pair_similarity_stability"`
	CIWidth               CIWidthSummary           `yaml:"bootstrap_ci_width"`
}

// FrequencyBin is one row of section 4/5/6/7/8: an independent, non-
// cumulative frequency interval, reusing the per-threshold computation whose
// MinCount equals the bin's lower bound.
type FrequencyBin struct {
	Bin                   string                   `yaml:"bin"`
	LowerBound            int                      `yaml:"lower_bound"`
	UpperBound            *int                     `yaml:"upper_bound"`
	Tokens                int                      `yaml:"tokens"`
	SelfProfileTrainTrain SelfProfileStability     `yaml:"self_profile_train_train"`
	TrainTest             TrainTestStability       `yaml:"train_test"`
	NearestNeighbors      NearestNeighborStability `yaml:"nearest_neighbor_stability"`
	Pairs                 PairStabilitySummary     `yaml:"pair_similarity_stability"`
	CIWidth               CIWidthSummary           `yaml:"bootstrap_ci_width"`
}

// ContinuousTokenMetric is one row per eligible (count>=min-token-count)
// token, section 9: the un-binned raw material for the correlations in
// section 10.
type ContinuousTokenMetric struct {
	Token                       string  `yaml:"token"`
	FullCount                   int     `yaml:"full_count"`
	PositionTrainTrainStability float64 `yaml:"position_train_train_stability"`
	LeftTrainTrainStability     float64 `yaml:"left_train_train_stability"`
	RightTrainTrainStability    float64 `yaml:"right_train_train_stability"`
	TrainTrainObservations      int     `yaml:"train_train_observations"`
	PositionTrainTestStability  float64 `yaml:"position_train_test_stability"`
	LeftTrainTestStability      float64 `yaml:"left_train_test_stability"`
	RightTrainTestStability     float64 `yaml:"right_train_test_stability"`
	TrainTestObservations       int     `yaml:"train_test_observations"`
	Top1Recovery                float64 `yaml:"top1_recovery"`
	Top1RecoveryObservations    int     `yaml:"top1_recovery_observations"`
	Top10Jaccard                float64 `yaml:"top10_jaccard"`
	Top10JaccardObservations    int     `yaml:"top10_jaccard_observations"`
}

// ContinuousPairMetric is one row per master candidate pair, sections 7/8/11
// plus the (diagnostic-only) pair reliability of sections 19/20.
type ContinuousPairMetric struct {
	TokenA                       string   `yaml:"token_a"`
	TokenB                       string   `yaml:"token_b"`
	CountA                       int      `yaml:"count_a"`
	CountB                       int      `yaml:"count_b"`
	MinCount                     int      `yaml:"min_count"`
	GeometricMeanCount           float64  `yaml:"geometric_mean_count"`
	FullSimilarity               float64  `yaml:"full_similarity"`
	FullPositionSimilarity       float64  `yaml:"full_position_similarity"`
	FullLeftSimilarity           float64  `yaml:"full_left_similarity"`
	FullRightSimilarity          float64  `yaml:"full_right_similarity"`
	FoldSimilarityStddev         *float64 `yaml:"fold_similarity_stddev"`
	FoldObservations             int      `yaml:"fold_observations"`
	BootstrapObservations        int      `yaml:"bootstrap_observations"`
	BootstrapMean                *float64 `yaml:"bootstrap_mean"`
	BootstrapCIWidth             *float64 `yaml:"bootstrap_ci_width"`
	BootstrapProbabilityAbove070 *float64 `yaml:"bootstrap_probability_above_070"`
	PositionReliabilityPair      float64  `yaml:"position_reliability_pair"`
	LeftReliabilityPair          float64  `yaml:"left_reliability_pair"`
	RightReliabilityPair         float64  `yaml:"right_reliability_pair"`
	PositionSupport              float64  `yaml:"position_support"`
	LeftSupport                  float64  `yaml:"left_support"`
	RightSupport                 float64  `yaml:"right_support"`
}

type Correlation struct {
	Metric       string  `yaml:"metric"`
	Rho          float64 `yaml:"rho"`
	Observations int     `yaml:"observations"`
}

type Correlations struct {
	FrequencyVsTokenStability   []Correlation `yaml:"frequency_vs_token_stability"`
	FrequencyVsPairStability    []Correlation `yaml:"frequency_vs_pair_stability"`
	ContextDiversityVsStability []Correlation `yaml:"context_diversity_vs_stability"`
}

type ComponentSizeStat struct {
	Size             int     `yaml:"n"`
	Tokens           int     `yaml:"tokens"`
	Runs             int     `yaml:"runs"`
	MeanSimilarity   float64 `yaml:"mean_similarity"`
	MedianSimilarity float64 `yaml:"median_similarity"`
	Percentile05     float64 `yaml:"percentile_05"`
	Percentile95     float64 `yaml:"percentile_95"`
}

type SubsamplingResult struct {
	SampleSize   int               `yaml:"sample_size"`
	Position     ComponentSizeStat `yaml:"position"`
	LeftContext  ComponentSizeStat `yaml:"left_context"`
	RightContext ComponentSizeStat `yaml:"right_context"`
}

type PerTokenSampleSize struct {
	N            int     `yaml:"n"`
	Runs         int     `yaml:"runs"`
	PositionMean float64 `yaml:"position_mean"`
	LeftMean     float64 `yaml:"left_mean"`
	RightMean    float64 `yaml:"right_mean"`
}

type PerTokenSubsampling struct {
	Token       string               `yaml:"token"`
	FullCount   int                  `yaml:"full_count"`
	SampleSizes []PerTokenSampleSize `yaml:"sample_sizes"`
}

type HeterogeneityComponent struct {
	Percentile10 float64 `yaml:"percentile_10"`
	Percentile25 float64 `yaml:"percentile_25"`
	Median       float64 `yaml:"median"`
	Percentile75 float64 `yaml:"percentile_75"`
	Percentile90 float64 `yaml:"percentile_90"`
}

type Heterogeneity struct {
	SampleSize   int                    `yaml:"sample_size"`
	Position     HeterogeneityComponent `yaml:"position"`
	LeftContext  HeterogeneityComponent `yaml:"left_context"`
	RightContext HeterogeneityComponent `yaml:"right_context"`
}

type Subsampling struct {
	MinFullCount  int                   `yaml:"min_full_count"`
	SampleSizes   []int                 `yaml:"sample_sizes"`
	Runs          int                   `yaml:"runs"`
	Tokens        int                   `yaml:"tokens"`
	Results       []SubsamplingResult   `yaml:"results"`
	PerToken      []PerTokenSubsampling `yaml:"per_token"`
	Heterogeneity []Heterogeneity       `yaml:"heterogeneity"`
}

// ReliabilityCurves is the lookup table of section 17: for each component,
// tested sample size -> mean subsampled-vs-reference similarity.
type ReliabilityCurves struct {
	Position     map[int]float64 `yaml:"position"`
	LeftContext  map[int]float64 `yaml:"left_context"`
	RightContext map[int]float64 `yaml:"right_context"`
}

type ComponentReliabilityThresholds struct {
	R80 *int `yaml:"r80"`
	R90 *int `yaml:"r90"`
	R95 *int `yaml:"r95"`
}

type ReliabilityThresholds struct {
	Position     ComponentReliabilityThresholds `yaml:"position"`
	LeftContext  ComponentReliabilityThresholds `yaml:"left_context"`
	RightContext ComponentReliabilityThresholds `yaml:"right_context"`
}

type TokenDiversity struct {
	Token                      string  `yaml:"token"`
	FullCount                  int     `yaml:"full_count"`
	UniquePredecessors         int     `yaml:"unique_predecessors"`
	UniqueSuccessors           int     `yaml:"unique_successors"`
	LeftEntropy                float64 `yaml:"left_entropy"`
	RightEntropy               float64 `yaml:"right_entropy"`
	EffectiveLeftObservations  float64 `yaml:"effective_left_observations"`
	EffectiveRightObservations float64 `yaml:"effective_right_observations"`
}

type ContextDiversity struct {
	Tokens []TokenDiversity `yaml:"tokens"`
}

type ReferencePairReliability struct {
	TokenA                       string      `yaml:"token_a"`
	TokenB                       string      `yaml:"token_b"`
	CountA                       int         `yaml:"count_a"`
	CountB                       int         `yaml:"count_b"`
	FullSimilarity               float64     `yaml:"full_similarity"`
	FullPositionSimilarity       float64     `yaml:"full_position_similarity"`
	FullLeftSimilarity           float64     `yaml:"full_left_similarity"`
	FullRightSimilarity          float64     `yaml:"full_right_similarity"`
	FoldSimilarityStddev         *float64    `yaml:"fold_similarity_stddev"`
	PositionReliabilityPair      float64     `yaml:"position_reliability_pair"`
	LeftReliabilityPair          float64     `yaml:"left_reliability_pair"`
	RightReliabilityPair         float64     `yaml:"right_reliability_pair"`
	BootstrapMean                *float64    `yaml:"bootstrap_mean"`
	BootstrapCI95                *[2]float64 `yaml:"bootstrap_ci95"`
	BootstrapProbabilityAbove070 *float64    `yaml:"bootstrap_probability_above_070"`
}

type Meta struct {
	Input            string `yaml:"input"`
	Classes          string `yaml:"classes"`
	PhysicalLines    int    `yaml:"physical_lines"`
	TokenOccurrences int    `yaml:"token_occurrences"`
	Transitions      int    `yaml:"transitions"`
	UniqueTokens     int    `yaml:"unique_tokens"`
	InputSHA256      string `yaml:"input_sha256"`
	ClassesSHA256    string `yaml:"classes_sha256"`
}

type ReferenceModel struct {
	Threshold          float64 `yaml:"threshold"`
	Classes            int     `yaml:"classes"`
	TokensInClasses    int     `yaml:"tokens_in_classes"`
	OccurrenceCoverage float64 `yaml:"occurrence_coverage"`
}

type Parameters struct {
	Folds                 int     `yaml:"folds"`
	FoldSeed              int64   `yaml:"fold_seed"`
	MinTokenCount         int     `yaml:"min_token_count"`
	Neighbors             int     `yaml:"neighbors"`
	BootstrapRuns         int     `yaml:"bootstrap_runs"`
	BootstrapSeed         int64   `yaml:"bootstrap_seed"`
	Threshold             float64 `yaml:"threshold"`
	ThresholdMargin       float64 `yaml:"threshold_margin"`
	CountThresholds       []int   `yaml:"count_thresholds"`
	SubsampleMinFullCount int     `yaml:"subsample_min_full_count"`
	SubsampleSizes        []int   `yaml:"subsample_sizes"`
	SubsampleRuns         int     `yaml:"subsample_runs"`
	SubsampleSeed         int64   `yaml:"subsample_seed"`
}

type Methodology struct {
	Scope                string `yaml:"scope"`
	Similarity           string `yaml:"similarity"`
	CumulativeThresholds string `yaml:"cumulative_thresholds"`
	FrequencyBins        string `yaml:"frequency_bins"`
	NearestNeighbors     string `yaml:"nearest_neighbors"`
	Bootstrap            string `yaml:"bootstrap"`
	Subsampling          string `yaml:"subsampling"`
	ReliabilityCurve     string `yaml:"reliability_curve"`
	PairReliability      string `yaml:"pair_reliability"`
	Interpretation       string `yaml:"interpretation"`
}

type Summary struct {
	TokenOccurrences               int            `yaml:"token_occurrences"`
	PhysicalLines                  int            `yaml:"physical_lines"`
	Transitions                    int            `yaml:"transitions"`
	ReferenceModel                 ReferenceModel `yaml:"reference_model"`
	BaseEligibleTokens             int            `yaml:"base_eligible_tokens"`
	CandidatePairs                 int            `yaml:"candidate_pairs"`
	LargestCumulativeThreshold     int            `yaml:"largest_cumulative_threshold"`
	LargestThresholdEligibleTokens int            `yaml:"largest_threshold_eligible_tokens"`
}

type Output struct {
	Meta                   Meta                       `yaml:"meta"`
	Parameters             Parameters                 `yaml:"parameters"`
	Methodology            Methodology                `yaml:"methodology"`
	Summary                Summary                    `yaml:"summary"`
	CumulativeThresholds   []CumulativeThreshold      `yaml:"cumulative_thresholds"`
	FrequencyBins          []FrequencyBin             `yaml:"frequency_bins"`
	ContinuousTokenMetrics []ContinuousTokenMetric    `yaml:"continuous_token_metrics"`
	ContinuousPairMetrics  []ContinuousPairMetric     `yaml:"continuous_pair_metrics"`
	Correlations           Correlations               `yaml:"correlations"`
	Subsampling            Subsampling                `yaml:"subsampling"`
	ReliabilityCurves      ReliabilityCurves          `yaml:"reliability_curves"`
	ReliabilityThresholds  ReliabilityThresholds      `yaml:"reliability_thresholds"`
	ContextDiversity       ContextDiversity           `yaml:"context_diversity"`
	ReferencePairs         []ReferencePairReliability `yaml:"reference_pairs"`
}
