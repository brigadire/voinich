package localregime

import "io"

type Config struct {
	CorpusPath, DistancePairsPath, ControlsPath, OutputDir string
	RegimeRadius, RegimeGap, RegimeControlsK               int
	WindowStep, MaxDistance, TopN                          int
	Pair                                                   string
	Seed                                                   int64
	RespectLineBoundaries                                  bool
	Quiet                                                  bool
	ProgressWriter                                         io.Writer
}

type RegimeMetric struct {
	Radius          int     `yaml:"radius"`
	Gap             int     `yaml:"gap"`
	Side            string  `yaml:"side"`
	JSSimilarity    float64 `yaml:"js_similarity"`
	WeightedOverlap float64 `yaml:"weighted_overlap"`
	Jaccard         float64 `yaml:"support_jaccard"`
	Cosine          float64 `yaml:"cosine_similarity"`
	DispersionA     float64 `yaml:"centroid_distance_a"`
	DispersionB     float64 `yaml:"centroid_distance_b"`
	PairwiseJSA     float64 `yaml:"mean_pairwise_js_distance_a"`
	PairwiseJSB     float64 `yaml:"mean_pairwise_js_distance_b"`
}
type DistanceMetric struct {
	Distance          int     `yaml:"distance"`
	Observed          float64 `yaml:"observed_js"`
	RegimeExpected    float64 `yaml:"regime_expected_js"`
	ResidualExcess    float64 `yaml:"residual_excess"`
	GlobalShuffle     float64 `yaml:"global_shuffle_js"`
	LineShuffle       float64 `yaml:"line_shuffle_js"`
	LocalBlockShuffle float64 `yaml:"local_block_shuffle_js"`
	RetainedFraction  float64 `yaml:"retained_fraction"`
	Baseline          float64 `yaml:"frequency_matched_shuffle_baseline_js"`
	RetainedEffect    float64 `yaml:"retained_effect"`
}
type PairResult struct {
	TokenA         string           `yaml:"token_a"`
	TokenB         string           `yaml:"token_b"`
	CountA         int              `yaml:"count_a"`
	CountB         int              `yaml:"count_b"`
	Regimes        []RegimeMetric   `yaml:"regime_sweep"`
	Distance       []DistanceMetric `yaml:"distance_profile"`
	Observed1To5   float64          `yaml:"observed_1_5"`
	Observed6To10  float64          `yaml:"observed_6_10"`
	Observed11To20 float64          `yaml:"observed_11_20"`
	Regime1To5     float64          `yaml:"regime_expected_1_5"`
	Regime6To10    float64          `yaml:"regime_expected_6_10"`
	Regime11To20   float64          `yaml:"regime_expected_11_20"`
	Residual1To5   float64          `yaml:"residual_1_5"`
	Residual6To10  float64          `yaml:"residual_6_10"`
	Residual11To20 float64          `yaml:"residual_11_20"`
	ConcentrationA float64          `yaml:"regime_concentration_a"`
	ConcentrationB float64          `yaml:"regime_concentration_b"`
	PrimaryRegime  float64          `yaml:"primary_regime_js"`
}
type WindowRow struct {
	Size               int     `yaml:"window_size"`
	Index              int     `yaml:"window_index"`
	Start              int     `yaml:"start"`
	End                int     `yaml:"end"`
	AdjacentJSDistance float64 `yaml:"adjacent_js_distance"`
	Concentration      float64 `yaml:"concentration"`
}
type SeparationRow struct {
	Size           int     `yaml:"window_size"`
	Separation     int     `yaml:"separation"`
	Comparisons    int     `yaml:"comparisons"`
	MeanJSDistance float64 `yaml:"mean_js_distance"`
	MeanSimilarity float64 `yaml:"mean_similarity"`
}
type ChangePoint struct {
	WindowSize int     `yaml:"window_size"`
	Position   int     `yaml:"position"`
	JSDistance float64 `yaml:"js_distance"`
	Threshold  float64 `yaml:"threshold"`
}
type ShuffleResult struct {
	TokenA     string  `yaml:"token_a"`
	TokenB     string  `yaml:"token_b"`
	Mode       string  `yaml:"mode"`
	BlockSize  int     `yaml:"block_size,omitempty"`
	Distance   int     `yaml:"distance"`
	Similarity float64 `yaml:"js_similarity"`
}
type TokenProfile struct {
	Token          string  `yaml:"token"`
	Count          int     `yaml:"count"`
	Concentration  float64 `yaml:"regime_concentration"`
	Entropy        float64 `yaml:"regime_membership_entropy"`
	MaxAssociation float64 `yaml:"maximum_regime_association"`
	Dispersion     float64 `yaml:"occurrence_centroid_distance"`
}
type Correlation struct {
	Metric   string  `yaml:"metric"`
	N        int     `yaml:"n"`
	Pearson  float64 `yaml:"pearson"`
	Spearman float64 `yaml:"spearman"`
}
type Output struct {
	Parameters   map[string]any  `yaml:"parameters"`
	Pairs        []PairResult    `yaml:"pairs"`
	Correlations []Correlation   `yaml:"correlations"`
	Separations  []SeparationRow `yaml:"window_separation"`
}

type corpus struct {
	Lines  [][]string
	Tokens []string
	Counts map[string]int
	LineAt []int
}
type pair struct{ A, B string }
type profile map[string]float64
type analysis struct {
	Out        Output
	Windows    []WindowRow
	Changes    []ChangePoint
	Tokens     []TokenProfile
	Shuffles   []ShuffleResult
	Occurrence map[string][]profile
	Controls   []controlRow
}
type controlRow struct {
	Target             pair
	Control            pair
	Rank               int
	Score              float64
	RegimeSimilarity   float64
	DistanceSimilarity float64
	ConcentrationA     float64
	ConcentrationB     float64
}
