package structuralprojection

import "io"

type Config struct {
	CorpusPath, StructuralPairsPath, DistancePairsPath, FamiliesPath, OutputDir  string
	MinStructuralSimilarity, MinReliability                                      float64
	ProjectionK, RandomProjections, MaxDistance, MinObservations, TopN, FamilyID int
	ProjectionMode, Pair                                                         string
	Seed                                                                         int64
	Quiet                                                                        bool
	ProgressWriter                                                               io.Writer
}

type Edge struct {
	A, B                                                                string
	CountA, CountB                                                      int
	Position, Left, Right, Similarity                                   float64
	PositionReliability, LeftReliability, RightReliability, Reliability float64
}

type Projection map[string]map[string]float64

type Metric struct {
	Distance                        int     `yaml:"distance"`
	TokenJS                         float64 `yaml:"token_js"`
	ProjectedJSFull                 float64 `yaml:"projected_js_full"`
	ProjectedJSAblated              float64 `yaml:"projected_js_ablated"`
	ProjectedJSFamily               float64 `yaml:"projected_js_family_control"`
	GainFull                        float64 `yaml:"gain_full"`
	GainAblated                     float64 `yaml:"gain_ablated"`
	GainFamily                      float64 `yaml:"gain_family_control"`
	RandomGainP95                   float64 `yaml:"random_projection_gain_p95"`
	TokenWeightedOverlap            float64 `yaml:"token_weighted_overlap"`
	ProjectedWeightedOverlapFull    float64 `yaml:"projected_weighted_overlap_full"`
	ProjectedWeightedOverlapAblated float64 `yaml:"projected_weighted_overlap_ablated"`
	TokenJaccard                    float64 `yaml:"token_support_jaccard"`
	ProjectedJaccardFull            float64 `yaml:"projected_support_jaccard_full"`
	ProjectedJaccardAblated         float64 `yaml:"projected_support_jaccard_ablated"`
	ObservationsA                   int     `yaml:"observations_a"`
	ObservationsB                   int     `yaml:"observations_b"`
	Reliability                     float64 `yaml:"reliability,omitempty"`
}

type GainControl struct {
	Observed            float64 `yaml:"observed_projection_gain"`
	RandomMean          float64 `yaml:"random_gain_mean"`
	RandomP95           float64 `yaml:"random_gain_p95"`
	RandomPercentile    float64 `yaml:"random_gain_percentile"`
	SmoothingMean       float64 `yaml:"smoothing_gain_mean"`
	SmoothingP95        float64 `yaml:"smoothing_gain_p95"`
	SmoothingPercentile float64 `yaml:"smoothing_gain_percentile"`
}

type PairSummary struct {
	MeanToken1To5   float64     `yaml:"mean_token_js_1_5"`
	MeanFull1To5    float64     `yaml:"mean_projected_js_1_5"`
	MeanAblated1To5 float64     `yaml:"mean_ablated_projected_js_1_5"`
	Gain1To5        float64     `yaml:"projection_gain_1_5"`
	Gain6To10       float64     `yaml:"gain_6_10"`
	Gain11To20      float64     `yaml:"gain_11_20"`
	Control         GainControl `yaml:"controls"`
	AblatedControl  GainControl `yaml:"ablated_controls"`
}

type PairResult struct {
	TokenA  string      `yaml:"token_a"`
	TokenB  string      `yaml:"token_b"`
	Right   []Metric    `yaml:"right_context"`
	Left    []Metric    `yaml:"left_context"`
	Summary PairSummary `yaml:"summary"`
}

type SequenceResult struct {
	TokenA                     string    `yaml:"token_a"`
	TokenB                     string    `yaml:"token_b"`
	Length                     int       `yaml:"sequence_length"`
	ExactSimilarity            float64   `yaml:"exact_suffix_similarity"`
	ProjectedSimilarityFull    float64   `yaml:"projected_suffix_similarity_full"`
	ProjectedSimilarityAblated float64   `yaml:"projected_suffix_similarity_ablated"`
	PositionFull               []float64 `yaml:"position_similarity_full"`
	PositionAblated            []float64 `yaml:"position_similarity_ablated"`
}

type FamilyDistance struct {
	Distance                   int     `yaml:"distance"`
	TokenCohesion              float64 `yaml:"token_family_cohesion"`
	ProjectedCohesionFull      float64 `yaml:"projected_family_cohesion_full"`
	ProjectedCohesionAblated   float64 `yaml:"projected_family_cohesion_ablated"`
	TokenDispersion            float64 `yaml:"token_dispersion"`
	ProjectedDispersionFull    float64 `yaml:"projected_dispersion_full"`
	ProjectedDispersionAblated float64 `yaml:"projected_dispersion_ablated"`
	TokenMedoid                string  `yaml:"token_medoid"`
	FullMedoid                 string  `yaml:"projected_medoid_full"`
	AblatedMedoid              string  `yaml:"projected_medoid_ablated"`
	MatchedPercentileToken     float64 `yaml:"matched_percentile_token"`
	MatchedPercentileFull      float64 `yaml:"matched_percentile_full"`
	MatchedPercentileAblated   float64 `yaml:"matched_percentile_ablated"`
}

type FamilyResult struct {
	ID        int              `yaml:"id"`
	Tokens    []string         `yaml:"tokens"`
	Distances []FamilyDistance `yaml:"right_context"`
}

type SweepRow struct {
	Method          string  `yaml:"method"`
	Parameter       float64 `yaml:"parameter"`
	MeanGainFull    float64 `yaml:"mean_gain_full"`
	MeanGainAblated float64 `yaml:"mean_gain_ablated"`
}
type ControlRow struct {
	TokenA, TokenB, Kind            string
	Observed, Mean, P95, Percentile float64
}
type ShuffleResult struct {
	Mode            string  `yaml:"mode"`
	MeanTokenJS     float64 `yaml:"mean_token_js_1_5"`
	MeanProjectedJS float64 `yaml:"mean_projected_js_1_5"`
	MeanGain        float64 `yaml:"mean_gain_1_5"`
}
type Transition struct {
	Source      string  `yaml:"source"`
	Destination string  `yaml:"destination"`
	Observed    float64 `yaml:"observed"`
	Baseline    float64 `yaml:"frequency_baseline"`
	Lift        float64 `yaml:"lift"`
}

type Output struct {
	Parameters  map[string]any    `yaml:"parameters"`
	Methodology map[string]string `yaml:"methodology"`
	Pairs       []PairResult      `yaml:"pairs"`
	Sweeps      []SweepRow        `yaml:"parameter_sweeps"`
	Shuffles    []ShuffleResult   `yaml:"shuffled_corpus_controls"`
	Transitions []Transition      `yaml:"strongest_structural_transitions"`
}

type corpus struct {
	Lines  [][]string
	Tokens []string
	Counts map[string]int
}
type pair struct{ A, B string }
type family struct {
	ID     int      `yaml:"id"`
	Tokens []string `yaml:"tokens"`
}
