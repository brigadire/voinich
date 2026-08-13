package globalregime

import "io"

type Config struct {
	CorpusPath     string
	OutputDir      string
	WindowSizes    []int
	Step           int
	Seed           int64
	Quiet          bool
	ProgressWriter io.Writer
}

type profile map[string]float64

type Window struct {
	WindowSize      int     `yaml:"window_size"`
	Index           int     `yaml:"window_index"`
	Start           int     `yaml:"start"`
	End             int     `yaml:"end"`
	Center          int     `yaml:"center"`
	Step            int     `yaml:"step"`
	JSDistance      float64 `yaml:"adjacent_js_distance"`
	WeightedOverlap float64 `yaml:"weighted_overlap"`
	Cosine          float64 `yaml:"cosine_similarity"`
	Variation       string  `yaml:"variation"`
	LocalPeak       bool    `yaml:"local_peak"`
	BroadTransition bool    `yaml:"broad_transition"`
	distribution    profile
}

type ChangePoint struct {
	WindowSize int     `yaml:"window_size"`
	Position   int     `yaml:"position"`
	Method     string  `yaml:"method"`
	Strength   float64 `yaml:"jump_strength"`
	Threshold  float64 `yaml:"threshold,omitempty"`
}

type StableBoundary struct {
	Position            int          `yaml:"position"`
	Support             map[int]bool `yaml:"scale_support"`
	SupportCount        int          `yaml:"scale_support_count"`
	SupportFraction     float64      `yaml:"scale_support_fraction"`
	MeanPosition        float64      `yaml:"mean_position"`
	MeanJumpStrength    float64      `yaml:"mean_jump_strength"`
	MaxJumpStrength     float64      `yaml:"max_jump_strength"`
	PositionUncertainty float64      `yaml:"position_uncertainty"`
}

type ClusterDiagnostic struct {
	WindowSize       int     `yaml:"window_size"`
	Method           string  `yaml:"method"`
	K                int     `yaml:"k"`
	Silhouette       float64 `yaml:"silhouette"`
	WithinDispersion float64 `yaml:"within_cluster_dispersion"`
	BetweenDistance  float64 `yaml:"between_cluster_distance"`
	ClusterSizes     []int   `yaml:"cluster_sizes"`
	TransitionCount  int     `yaml:"transition_count"`
	Fragmentation    float64 `yaml:"segment_fragmentation"`
	labels           []int
}

type Output struct {
	Parameters  map[string]any      `yaml:"parameters"`
	Boundaries  []StableBoundary    `yaml:"stable_boundaries"`
	Diagnostics []ClusterDiagnostic `yaml:"model_selection"`
}

type scaleAnalysis struct {
	size        int
	step        int
	windows     []Window
	changes     []ChangePoint
	diagnostics []ClusterDiagnostic
}

type analysis struct {
	Out    Output
	Scales []scaleAnalysis
}
