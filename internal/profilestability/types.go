package profilestability

import "zcore.dev/voinich/internal/normalization"

type Config struct {
	InputPath       string
	ClassesPath     string
	Folds           int
	FoldSeed        int64
	MinTokenCount   int
	Neighbors       int
	BootstrapRuns   int
	BootstrapSeed   int64
	Threshold       float64
	ThresholdMargin float64
	Progress        func(string)
}

type Profile struct {
	Count     int
	Positions map[int]int
	Left      map[string]int
	Right     map[string]int
}

type Components struct {
	Similarity         float64 `yaml:"similarity"`
	PositionSimilarity float64 `yaml:"position_similarity"`
	LeftSimilarity     float64 `yaml:"left_context_similarity"`
	RightSimilarity    float64 `yaml:"right_context_similarity"`
}

type Distribution struct {
	Observations  int     `yaml:"observations"`
	Mean          float64 `yaml:"mean"`
	Stddev        float64 `yaml:"stddev"`
	Min           float64 `yaml:"min"`
	Max           float64 `yaml:"max"`
	Range         float64 `yaml:"range"`
	Percentile025 float64 `yaml:"percentile_025,omitempty"`
	Percentile50  float64 `yaml:"percentile_50,omitempty"`
	Percentile975 float64 `yaml:"percentile_975,omitempty"`
}

type Neighbor struct {
	Token      string `yaml:"token"`
	Components `yaml:",inline"`
	Rank       int `yaml:"rank"`
}

type TrainTestFold struct {
	Fold               int     `yaml:"fold"`
	PositionSimilarity float64 `yaml:"position_similarity"`
	LeftSimilarity     float64 `yaml:"left_context_similarity"`
	RightSimilarity    float64 `yaml:"right_context_similarity"`
}

type TokenProfileStability struct {
	Token                       string          `yaml:"token"`
	Count                       int             `yaml:"count"`
	EligibleFull                bool            `yaml:"eligible_full"`
	EligibleTrainFolds          []int           `yaml:"eligible_train_folds"`
	EligibleTestFolds           []int           `yaml:"eligible_test_folds"`
	FullCorpusNeighbors         []Neighbor      `yaml:"full_corpus_neighbors"`
	PositionStability           Distribution    `yaml:"position_stability"`
	LeftContextStability        Distribution    `yaml:"left_context_stability"`
	RightContextStability       Distribution    `yaml:"right_context_stability"`
	TrainTest                   []TrainTestFold `yaml:"train_test"`
	TrainTestPositionSimilarity Distribution    `yaml:"train_test_position_similarity"`
	TrainTestLeftSimilarity     Distribution    `yaml:"train_test_left_similarity"`
	TrainTestRightSimilarity    Distribution    `yaml:"train_test_right_similarity"`
}

type RankCorrelation struct {
	Comparisons     int     `yaml:"comparisons"`
	MeanCommonItems float64 `yaml:"mean_common_items"`
	MeanSpearmanRho float64 `yaml:"mean_spearman_rho"`
	MinSpearmanRho  float64 `yaml:"min_spearman_rho"`
	MaxSpearmanRho  float64 `yaml:"max_spearman_rho"`
}

type NeighborStability struct {
	Token                  string          `yaml:"token"`
	FoldPairComparisons    int             `yaml:"fold_pair_comparisons"`
	MeanJaccard            float64         `yaml:"mean_jaccard"`
	MinJaccard             float64         `yaml:"min_jaccard"`
	MaxJaccard             float64         `yaml:"max_jaccard"`
	Top1SameFraction       float64         `yaml:"top1_same_fraction"`
	Top3OverlapMean        float64         `yaml:"top3_overlap_mean"`
	Top5OverlapMean        float64         `yaml:"top5_overlap_mean"`
	Top10OverlapMean       float64         `yaml:"top10_overlap_mean"`
	FullTop1Neighbor       string          `yaml:"full_top1_neighbor,omitempty"`
	FoldsWhereTop1Eligible int             `yaml:"folds_where_top1_eligible"`
	FoldsWhereSameTop1     int             `yaml:"folds_where_same_top1"`
	Top1RecoveryFraction   float64         `yaml:"top1_recovery_fraction"`
	RankCorrelation        RankCorrelation `yaml:"rank_correlation"`
}

type FoldSimilarity struct {
	Fold       int `yaml:"fold"`
	Components `yaml:",inline"`
}

type ThresholdCrossing struct {
	Threshold              float64 `yaml:"threshold"`
	FoldsAboveThreshold    int     `yaml:"folds_above_threshold"`
	FoldsBelowThreshold    int     `yaml:"folds_below_threshold"`
	ThresholdCrossingCount int     `yaml:"threshold_crossing_count"`
}

type PairStability struct {
	TokenA                    string              `yaml:"token_a"`
	TokenB                    string              `yaml:"token_b"`
	TokenACount               int                 `yaml:"token_a_count"`
	TokenBCount               int                 `yaml:"token_b_count"`
	MinCount                  int                 `yaml:"min_count"`
	GeometricMeanCount        float64             `yaml:"geometric_mean_count"`
	Full                      Components          `yaml:"full"`
	Folds                     []FoldSimilarity    `yaml:"folds"`
	Summary                   Distribution        `yaml:"summary"`
	Thresholds                []ThresholdCrossing `yaml:"threshold_crossings"`
	MeanMargin                float64             `yaml:"mean_margin"`
	MinMargin                 float64             `yaml:"min_margin"`
	MaxMargin                 float64             `yaml:"max_margin"`
	NearThresholdFraction     float64             `yaml:"near_threshold_fraction"`
	MeanMemberNeighborJaccard float64             `yaml:"mean_member_neighbor_jaccard"`
}

type BootstrapPair struct {
	TokenA                    string       `yaml:"token_a"`
	TokenB                    string       `yaml:"token_b"`
	Similarity                Distribution `yaml:"similarity"`
	PositionSimilarity        Distribution `yaml:"position_similarity"`
	LeftContextSimilarity     Distribution `yaml:"left_context_similarity"`
	RightContextSimilarity    Distribution `yaml:"right_context_similarity"`
	ProbabilityAboveThreshold float64      `yaml:"probability_above_threshold"`
	MostVariableComponent     string       `yaml:"most_variable_component"`
}

type FrequencyBin struct {
	Bin                   string  `yaml:"bin"`
	Tokens                int     `yaml:"tokens"`
	MeanPositionStability float64 `yaml:"mean_position_stability"`
	MeanLeftStability     float64 `yaml:"mean_left_stability"`
	MeanRightStability    float64 `yaml:"mean_right_stability"`
	MeanNeighborJaccard   float64 `yaml:"mean_neighbor_jaccard"`
	MeanTop1Recovery      float64 `yaml:"mean_top1_recovery"`
}

type PairFrequencyBin struct {
	Bin                       string  `yaml:"bin"`
	Pairs                     int     `yaml:"pairs"`
	MeanSimilarityStddev      float64 `yaml:"mean_similarity_stddev"`
	MeanMemberNeighborJaccard float64 `yaml:"mean_member_neighbor_jaccard"`
	ThresholdCrossingFraction float64 `yaml:"threshold_crossing_fraction"`
}

type HardClassStability struct {
	ReferencePairs        int     `yaml:"reference_pairs"`
	MeanSameClassFraction float64 `yaml:"mean_same_class_fraction"`
	PairsAlwaysSameClass  int     `yaml:"pairs_always_same_class"`
	PairsNeverSameClass   int     `yaml:"pairs_never_same_class"`
}

type Summary struct {
	EligibleTokens     int                `yaml:"eligible_tokens"`
	CandidatePairs     int                `yaml:"candidate_pairs"`
	HardClassStability HardClassStability `yaml:"hard_class_stability"`
	SelfProfile        struct {
		MeanPositionStability float64 `yaml:"mean_position_stability"`
		MeanLeftStability     float64 `yaml:"mean_left_stability"`
		MeanRightStability    float64 `yaml:"mean_right_stability"`
	} `yaml:"self_profile"`
	NearestNeighbors struct {
		MeanTop1Recovery float64 `yaml:"mean_top1_recovery"`
		MeanTop5Overlap  float64 `yaml:"mean_top5_overlap"`
		MeanTop10Jaccard float64 `yaml:"mean_top10_jaccard"`
	} `yaml:"nearest_neighbors"`
	Pairs struct {
		MeanSimilarityStddev              float64 `yaml:"mean_similarity_stddev"`
		PairsNearThreshold                int     `yaml:"pairs_near_threshold"`
		PairsBootstrapProbabilityAbove095 int     `yaml:"pairs_with_bootstrap_p_above_070_ge_095"`
	} `yaml:"pairs"`
}

type ReferencePair struct {
	TokenA                       string     `yaml:"token_a"`
	TokenB                       string     `yaml:"token_b"`
	FullSimilarity               float64    `yaml:"full_similarity"`
	FoldMean                     float64    `yaml:"fold_mean"`
	FoldStddev                   float64    `yaml:"fold_stddev"`
	FoldMin                      float64    `yaml:"fold_min"`
	FoldMax                      float64    `yaml:"fold_max"`
	BootstrapMean                float64    `yaml:"bootstrap_mean"`
	BootstrapCI95                [2]float64 `yaml:"bootstrap_ci95"`
	BootstrapProbabilityAbove070 float64    `yaml:"bootstrap_probability_above_070"`
}

type ReferenceMemberNeighbor struct {
	Token                string  `yaml:"token"`
	MeanJaccard          float64 `yaml:"mean_jaccard"`
	Top1RecoveryFraction float64 `yaml:"top1_recovery_fraction"`
}

type ReferenceClass struct {
	ClassID                 string                    `yaml:"class_id"`
	Members                 []string                  `yaml:"members"`
	Pairwise                []ReferencePair           `yaml:"pairwise"`
	MemberNeighborStability []ReferenceMemberNeighbor `yaml:"member_neighbor_stability"`
	WeakestPair             *ReferencePair            `yaml:"weakest_pair,omitempty"`
	StrongestPair           *ReferencePair            `yaml:"strongest_pair,omitempty"`
}

type ComponentDiagnostic struct {
	TokenA                    string  `yaml:"token_a"`
	TokenB                    string  `yaml:"token_b"`
	FullSimilarity            float64 `yaml:"full_similarity"`
	SimilarityWithoutPosition float64 `yaml:"similarity_without_position"`
	SimilarityWithoutLeft     float64 `yaml:"similarity_without_left_context"`
	SimilarityWithoutRight    float64 `yaml:"similarity_without_right_context"`
	DeltaWithoutPosition      float64 `yaml:"delta_without_position"`
	DeltaWithoutLeft          float64 `yaml:"delta_without_left_context"`
	DeltaWithoutRight         float64 `yaml:"delta_without_right_context"`
}

type Meta struct {
	Input            string `yaml:"input"`
	Classes          string `yaml:"classes"`
	PhysicalLines    int    `yaml:"physical_lines"`
	TokenOccurrences int    `yaml:"token_occurrences"`
	UniqueTokens     int    `yaml:"unique_tokens"`
	InputSHA256      string `yaml:"input_sha256"`
	ClassesSHA256    string `yaml:"classes_sha256"`
}

type Parameters struct {
	Folds           int       `yaml:"folds"`
	FoldSeed        int64     `yaml:"fold_seed"`
	MinTokenCount   int       `yaml:"min_token_count"`
	Neighbors       int       `yaml:"neighbors"`
	BootstrapRuns   int       `yaml:"bootstrap_runs"`
	BootstrapSeed   int64     `yaml:"bootstrap_seed"`
	Threshold       float64   `yaml:"threshold"`
	ThresholdMargin float64   `yaml:"threshold_margin"`
	Thresholds      []float64 `yaml:"thresholds"`
}

type Methodology struct {
	Profiles          string `yaml:"profiles"`
	Similarity        string `yaml:"similarity"`
	Eligibility       string `yaml:"eligibility"`
	Split             string `yaml:"split"`
	TrainTest         string `yaml:"train_test"`
	Neighbors         string `yaml:"nearest_neighbors"`
	RankOverlap       string `yaml:"rank_overlap"`
	RankCorrelation   string `yaml:"rank_correlation"`
	Bootstrap         string `yaml:"bootstrap"`
	CandidatePairs    string `yaml:"candidate_pairs"`
	ThresholdCrossing string `yaml:"threshold_crossing"`
	ComponentAblation string `yaml:"component_ablation"`
	Interpretation    string `yaml:"interpretation"`
}

type Output struct {
	Meta                     Meta                    `yaml:"meta"`
	Parameters               Parameters              `yaml:"parameters"`
	Methodology              Methodology             `yaml:"methodology"`
	Summary                  Summary                 `yaml:"summary"`
	TokenProfileStability    []TokenProfileStability `yaml:"token_profile_stability"`
	NearestNeighborStability []NeighborStability     `yaml:"nearest_neighbor_stability"`
	PairSimilarityStability  []PairStability         `yaml:"pair_similarity_stability"`
	BootstrapPairUncertainty []BootstrapPair         `yaml:"bootstrap_pair_uncertainty"`
	FrequencyDependence      []FrequencyBin          `yaml:"frequency_dependence"`
	PairFrequencyDependence  []PairFrequencyBin      `yaml:"pair_frequency_dependence"`
	ReferenceClasses         []ReferenceClass        `yaml:"reference_classes"`
	ComponentDiagnostics     []ComponentDiagnostic   `yaml:"component_diagnostics"`
}

type classFile struct {
	Models []normalization.Model `yaml:"models"`
}
