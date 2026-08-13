// Package clustermetadataglobal performs the confirmatory, whole-search-space
// multiple-comparison correction for the association between frozen blind
// distributional regimes and Currier/hand metadata. It never recomputes
// discovery: windows, clustering and cluster assignments are loaded exactly
// as produced by global-regime-analyze, and only the metadata labels are
// permuted.
package clustermetadataglobal

import "io"

// Config describes one confirmatory run. The frozen discovery search space
// (window_size in {50,100,200,500,1000}, method in {contiguous_segmentation,
// hierarchical, k_medoids}, K in 2..15) is fixed by the package and is never
// derived from Config.
type Config struct {
	DiscoveryDir       string
	TokenMetadataMap   string
	MetadataReportPath string
	OutputDir          string
	Permutations       int
	Seed               int64
	Quiet              bool
	ProgressWriter     io.Writer
}

// WindowSizes is the exact, pre-existing frozen sweep. No new window size may
// be added at this stage.
var WindowSizes = []int{50, 100, 200, 500, 1000}

// Methods is the exact frozen clustering method sweep, named as they appear
// in global_distributional_cluster_assignments.tsv.
var Methods = []string{"contiguous_segmentation", "hierarchical", "k_medoids"}

// KMin and KMax bound the frozen K sweep (inclusive).
const KMin, KMax = 2, 15

// Kinds are the two metadata dimensions tested against the frozen regimes.
var Kinds = []string{"currier", "hand"}

func ksRange() []int {
	r := make([]int, 0, KMax-KMin+1)
	for k := KMin; k <= KMax; k++ {
		r = append(r, k)
	}
	return r
}

// Scope is one analysis lens over the frozen search space: the primary
// all-windows test, or a purity-restricted sensitivity variant. Thresholds
// are fixed in advance and are never chosen post hoc.
type Scope struct {
	Name      string
	Threshold float64
}

// Scopes lists the primary analysis and the two prespecified purity
// sensitivity experiments, in that order. Only Scopes[0] is primary evidence.
var Scopes = []Scope{{"primary", 0}, {"purity_0.8", 0.8}, {"purity_0.9", 0.9}}

// WindowRange is one frozen window's token span, shared by every method/K at
// its window size.
type WindowRange struct{ Index, Start, End int }

type comboKey struct {
	WindowSize int
	Method     string
	K          int
}

// comboData holds one frozen (window_size, method, K) cluster assignment,
// aligned index-for-index with the window order for that window size.
type comboData struct {
	Cluster     []int
	NumClusters int
}

// frozenSpace is the fully loaded, validated frozen discovery search space.
type frozenSpace struct {
	Windows map[int][]WindowRange
	Combos  map[comboKey]comboData
	N       int
}

// StatSeries is one primary or secondary global-correction statistic: an
// observed value with recorded coordinates, plus its null distribution.
type StatSeries struct {
	Metadata       string
	Scope          string
	Metric         string
	MethodScope    string
	Observed       float64
	ObservedWindow int
	ObservedMethod string
	ObservedK      int
	Null           []float64
}

// ScalePersistence records the five prespecified scale-specific max-over-K
// values (no maximum selection across scales) for one method/metadata/metric.
type ScalePersistence struct {
	Metadata string
	Scope    string
	Metric   string
	Method   string
	PerWindow map[int]float64 // window_size -> max NMI/ARI over frozen K
	Mean      float64
	Min       float64
}
