package propertytrajectory

import "io"

type Config struct {
	CorpusPath, StructuralPairsPath, DistancePairsPath, ControlsPath, OutputDir string
	Pair                                                                        string
	MaxDistance, MinTokenFrequency, TopN, RandomPairs                           int
	Seed                                                                        int64
	Quiet                                                                       bool
	ProgressWriter                                                              io.Writer
}

type PropertyValue struct {
	Raw        float64 `yaml:"raw"`
	Normalized float64 `yaml:"normalized"`
}

type TokenProperties struct {
	Token      string                   `yaml:"token"`
	Count      int                      `yaml:"count"`
	Properties map[string]PropertyValue `yaml:"properties"`
}

type PropertySummary struct {
	MeanA, MeanB, Delta                float64
	RawMeanA, RawMeanB                 float64
	MedianA, MedianB, StddevA, StddevB float64
	P25A, P25B, P75A, P75B             float64
}

func (p PropertySummary) MarshalYAML() (any, error) {
	return struct {
		MeanA    float64 `yaml:"mean_a"`
		MeanB    float64 `yaml:"mean_b"`
		Delta    float64 `yaml:"delta"`
		RawMeanA float64 `yaml:"raw_mean_a"`
		RawMeanB float64 `yaml:"raw_mean_b"`
		MedianA  float64 `yaml:"median_a"`
		MedianB  float64 `yaml:"median_b"`
		StddevA  float64 `yaml:"stddev_a"`
		StddevB  float64 `yaml:"stddev_b"`
		P25A     float64 `yaml:"p25_a"`
		P25B     float64 `yaml:"p25_b"`
		P75A     float64 `yaml:"p75_a"`
		P75B     float64 `yaml:"p75_b"`
	}{p.MeanA, p.MeanB, p.Delta, p.RawMeanA, p.RawMeanB, p.MedianA, p.MedianB, p.StddevA, p.StddevB, p.P25A, p.P25B, p.P75A, p.P75B}, nil
}

type DistanceProfile struct {
	Distance              int                        `yaml:"distance"`
	ObservationsA         int                        `yaml:"observations_a,omitempty"`
	ObservationsB         int                        `yaml:"observations_b,omitempty"`
	ExcludedLowFrequencyA int                        `yaml:"excluded_low_frequency_a,omitempty"`
	ExcludedLowFrequencyB int                        `yaml:"excluded_low_frequency_b,omitempty"`
	CosineSimilarity      float64                    `yaml:"cosine_similarity"`
	EuclideanDistance     float64                    `yaml:"euclidean_distance"`
	ManhattanDistance     float64                    `yaml:"manhattan_distance"`
	Correlation           float64                    `yaml:"correlation"`
	Properties            map[string]PropertySummary `yaml:"properties"`
}

type PropertyRanking struct {
	Property               string  `yaml:"property"`
	TrajectoryCorrelation  float64 `yaml:"trajectory_correlation"`
	MeanAbsoluteDifference float64 `yaml:"mean_absolute_difference"`
}

type ModeScore struct {
	Mode           string  `yaml:"mode"`
	MeanCosine1To5 float64 `yaml:"mean_cosine_1_5"`
}
type PairSummary struct {
	Cosine1To5, Cosine6To10, Cosine11To20 float64
	MatchedPercentile, RandomPercentile   float64
	Modes                                 []ModeScore       `yaml:"property_group_ablation"`
	StrongestMatching, StrongestDiffering []PropertyRanking `yaml:"strongest_matching_properties"`
}

func (s PairSummary) MarshalYAML() (any, error) {
	return struct {
		Cosine1To5         float64           `yaml:"cosine_1_5"`
		Cosine6To10        float64           `yaml:"cosine_6_10"`
		Cosine11To20       float64           `yaml:"cosine_11_20"`
		MatchedPercentile  float64           `yaml:"matched_percentile"`
		RandomPercentile   float64           `yaml:"random_percentile"`
		Modes              []ModeScore       `yaml:"property_group_ablation"`
		StrongestMatching  []PropertyRanking `yaml:"strongest_matching_properties"`
		StrongestDiffering []PropertyRanking `yaml:"strongest_differing_properties"`
	}{s.Cosine1To5, s.Cosine6To10, s.Cosine11To20, s.MatchedPercentile, s.RandomPercentile, s.Modes, s.StrongestMatching, s.StrongestDiffering}, nil
}

type PairResult struct {
	TokenA, TokenB   string
	CountA, CountB   int
	DistanceProfiles []DistanceProfile `yaml:"distance_profiles"`
	Summary          PairSummary       `yaml:"summary"`
}

func (p PairResult) MarshalYAML() (any, error) {
	return struct {
		Pair struct {
			A string `yaml:"a"`
			B string `yaml:"b"`
		} `yaml:"pair"`
		Counts struct {
			A int `yaml:"a"`
			B int `yaml:"b"`
		} `yaml:"counts"`
		DistanceProfiles []DistanceProfile `yaml:"distance_profiles"`
		Summary          PairSummary       `yaml:"summary"`
	}{struct {
		A string `yaml:"a"`
		B string `yaml:"b"`
	}{p.TokenA, p.TokenB}, struct {
		A int `yaml:"a"`
		B int `yaml:"b"`
	}{p.CountA, p.CountB}, p.DistanceProfiles, p.Summary}, nil
}

type Baseline struct {
	Scope    string  `yaml:"scope"`
	Distance int     `yaml:"distance,omitempty"`
	Median   float64 `yaml:"median"`
	P90      float64 `yaml:"p90"`
	P95      float64 `yaml:"p95"`
	P99      float64 `yaml:"p99"`
}
type ShuffleResult struct {
	Mode           string  `yaml:"mode"`
	PairA          string  `yaml:"token_a"`
	PairB          string  `yaml:"token_b"`
	MeanCosine1To5 float64 `yaml:"mean_cosine_1_5"`
}
type Output struct {
	Parameters      map[string]any                `yaml:"parameters"`
	Methodology     map[string]string             `yaml:"methodology"`
	PropertyGroups  map[string][]string           `yaml:"property_groups"`
	Normalization   map[string]map[string]float64 `yaml:"global_normalization"`
	Pairs           []PairResult                  `yaml:"pairs"`
	RandomBaselines []Baseline                    `yaml:"random_frequency_matched_baseline"`
	Shuffles        []ShuffleResult               `yaml:"shuffled_corpus_controls"`
}

type corpus struct {
	Lines  [][]string
	Tokens []string
	Counts map[string]int
}
type pair struct{ A, B string }
type analysis struct {
	Out     Output
	Tokens  []TokenProperties
	Matched map[pair][]float64
	Random  map[pair][]float64
}
