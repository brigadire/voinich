package vocabularygrowth

type Parameters struct {
	Checkpoints      []int
	WindowSizes      []int
	SegmentCounts    []int
	NullPermutations int
	Seed             int64
	FitMinN          int
	FitMaxN          int
}

func DefaultParameters() Parameters {
	return Parameters{WindowSizes: []int{500, 1000, 2000}, SegmentCounts: []int{4, 8}, NullPermutations: 100, Seed: 1}
}

type Point struct {
	N, Vocabulary, Hapax, Dis, Tri int
	TTR, BetaEffective             float64
}
type WindowPoint struct {
	Start, End, Tokens, NewTypes int
	NewTypeRate                  float64
}
type NullPoint struct {
	N                                              int
	Observed, NullMean, NullSD, Effect, EmpiricalP float64
}
type SegmentPoint struct {
	Segments, Segment, CheckpointN, Vocabulary int
	HeapsBeta, BetaEffective, NewTypeRate      float64
}
type Fit struct {
	K, Beta, R2, SSE, MaxAbsResidual float64
	NMin, NMax                       int
	Points                           int
}
type Result struct {
	TotalTokens int
	Final       Point
	Checkpoints []int
	Growth      []Point
	Windows     []WindowPoint
	Null        []NullPoint
	Segments    []SegmentPoint
	Fit         Fit
	Parameters  Parameters
}
