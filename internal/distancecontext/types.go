package distancecontext

type Config struct {
	CorpusPath, DistantPath, FamiliesPath, ControlsPath, OutputDir string
	MaxDistance, MinObservations, TopN, FamilyID                   int
	Pair                                                           string
	RespectLineBoundaries                                          bool
}

type Metric struct {
	Distance           int     `yaml:"distance"`
	JSSimilarity       float64 `yaml:"js_similarity"`
	WeightedOverlap    float64 `yaml:"weighted_overlap"`
	Jaccard            float64 `yaml:"jaccard_support_overlap"`
	ObservationsA      int     `yaml:"observations_a"`
	ObservationsB      int     `yaml:"observations_b"`
	EffectiveSupportA  float64 `yaml:"effective_support_a"`
	EffectiveSupportB  float64 `yaml:"effective_support_b"`
	Reliability        float64 `yaml:"reliability"`
	Reliable           bool    `yaml:"reliable"`
	BaselinePercentile float64 `yaml:"baseline_percentile,omitempty"`
}

type BoundaryMetric struct {
	Distance    int     `yaml:"distance"`
	Continuous  float64 `yaml:"continuous_similarity"`
	LineBounded float64 `yaml:"line_bounded_similarity"`
	Difference  float64 `yaml:"difference"`
}

type Summary struct {
	At1             float64 `yaml:"similarity_at_1"`
	At2             float64 `yaml:"similarity_at_2"`
	At3             float64 `yaml:"similarity_at_3"`
	At5             float64 `yaml:"similarity_at_5"`
	At10            float64 `yaml:"similarity_at_10"`
	At20            float64 `yaml:"similarity_at_20"`
	Mean1To5        float64 `yaml:"mean_1_5"`
	Mean6To10       float64 `yaml:"mean_6_10"`
	Mean11To20      float64 `yaml:"mean_11_20"`
	Persistence1To5 float64 `yaml:"persistence_1_5_percentile"`
}

type PairResult struct {
	TokenA              string           `yaml:"token_a"`
	TokenB              string           `yaml:"token_b"`
	CountA              int              `yaml:"count_a,omitempty"`
	CountB              int              `yaml:"count_b,omitempty"`
	Right               []Metric         `yaml:"right_context"`
	Left                []Metric         `yaml:"left_context"`
	LineBoundedRight    []Metric         `yaml:"line_bounded_right_context"`
	LineBoundedLeft     []Metric         `yaml:"line_bounded_left_context"`
	BoundarySensitivity []BoundaryMetric `yaml:"boundary_sensitivity"`
	RightSummary        Summary          `yaml:"right_summary"`
	LeftSummary         Summary          `yaml:"left_summary"`
}

type BaselineRow struct {
	Distance int     `yaml:"distance"`
	Pairs    int     `yaml:"pair_count"`
	Median   float64 `yaml:"median"`
	P90      float64 `yaml:"p90"`
	P95      float64 `yaml:"p95"`
}

type SequenceMetric struct {
	Length            int     `yaml:"sequence_length"`
	JSSimilarity      float64 `yaml:"js_similarity"`
	WeightedOverlap   float64 `yaml:"weighted_overlap"`
	Jaccard           float64 `yaml:"jaccard_support_overlap"`
	ObservationsA     int     `yaml:"observations_a"`
	ObservationsB     int     `yaml:"observations_b"`
	EffectiveSupportA float64 `yaml:"effective_support_a"`
	EffectiveSupportB float64 `yaml:"effective_support_b"`
	Reliability       float64 `yaml:"reliability"`
	Reliable          bool    `yaml:"reliable"`
}

type SequencePair struct {
	TokenA      string           `yaml:"token_a"`
	TokenB      string           `yaml:"token_b"`
	Continuous  []SequenceMetric `yaml:"continuous_suffix_context"`
	LineBounded []SequenceMetric `yaml:"line_bounded_suffix_context"`
}

type ControlResult struct {
	TargetA  string   `yaml:"target_a"`
	TargetB  string   `yaml:"target_b"`
	Rank     int      `yaml:"control_rank"`
	ControlA string   `yaml:"control_a"`
	ControlB string   `yaml:"control_b"`
	Profile  []Metric `yaml:"right_context"`
}

type DistanceMatrix struct {
	Distance   int         `yaml:"distance"`
	Values     [][]float64 `yaml:"values"`
	Medoid     string      `yaml:"medoid"`
	Cohesion   float64     `yaml:"family_cohesion"`
	Percentile float64     `yaml:"random_matched_group_percentile"`
}

type FamilyResult struct {
	ID       int              `yaml:"id"`
	Tokens   []string         `yaml:"tokens"`
	Profiles []DistanceMatrix `yaml:"right_context_matrices"`
}

type Output struct {
	Parameters map[string]any `yaml:"parameters"`
	TokenCount int            `yaml:"token_count"`
	PairCount  int            `yaml:"pair_count"`
	Baseline   []BaselineRow  `yaml:"right_context_baseline"`
	Pairs      []PairResult   `yaml:"pairs"`
}
