package structuralreliability

import "sort"

// ReliabilityTable is the reusable lookup + interpolation this task exists
// to produce (section 18): given a component's empirical subsampling curve,
// it answers "how reproducible is a profile built from n observations?" for
// any n, not just the tested sample sizes. It is exported so the future
// soft structural analyzer can import internal/structuralreliability and
// call Reliability directly, without redoing the subsampling experiment.
type ReliabilityTable struct {
	points []reliabilityPoint
}

type reliabilityPoint struct {
	n     int
	value float64
}

// NewReliabilityTable builds a table from a lookup of tested sample size ->
// mean similarity. The input need not be sorted.
func NewReliabilityTable(curve map[int]float64) ReliabilityTable {
	points := make([]reliabilityPoint, 0, len(curve))
	for n, value := range curve {
		points = append(points, reliabilityPoint{n: n, value: value})
	}
	sort.Slice(points, func(i, j int) bool { return points[i].n < points[j].n })
	return ReliabilityTable{points: points}
}

// Reliability implements the rules of section 18 exactly:
//   - an exact tested sample size returns its lookup value;
//   - between two tested points, linear interpolation happens in log2(n)
//     space;
//   - below the smallest tested point, the smallest point's value is
//     returned (never extrapolated downward);
//   - above the largest tested point, the largest point's value is returned
//     (never extrapolated above the observed maximum).
func (table ReliabilityTable) Reliability(n int) float64 {
	if len(table.points) == 0 {
		return 0
	}
	if n <= table.points[0].n {
		return table.points[0].value
	}
	last := table.points[len(table.points)-1]
	if n >= last.n {
		return last.value
	}
	for i := 1; i < len(table.points); i++ {
		if float64(n) > float64(table.points[i].n) {
			continue
		}
		lower, upper := table.points[i-1], table.points[i]
		if lower.n == upper.n {
			return lower.value
		}
		weight := (Log2(float64(n)) - Log2(float64(lower.n))) / (Log2(float64(upper.n)) - Log2(float64(lower.n)))
		return lower.value + (upper.value-lower.value)*weight
	}
	return last.value
}

// PairReliability computes the diagnostic-only, geometric-mean pair
// reliability of section 19: R_component_pair = geometric_mean(R(countA),
// R(countB)). It is never folded back into the similarity formula.
func PairReliability(table ReliabilityTable, countA, countB int) float64 {
	return GeometricMean(table.Reliability(countA), table.Reliability(countB))
}

// ComponentSupport is the diagnostic-only value of section 20:
// component_similarity * component_reliability.
func ComponentSupport(similarity, reliability float64) float64 { return similarity * reliability }

// ReliabilityThresholdsFor finds, for one component's tested curve, the
// smallest tested n at which the curve's value first reaches each of
// 0.80/0.90/0.95 (section 21). It never interpolates or extrapolates beyond
// tested points: a level absent from the tested points is reported as null.
func ReliabilityThresholdsFor(sortedSizes []int, curve map[int]float64) ComponentReliabilityThresholds {
	return ComponentReliabilityThresholds{
		R80: firstAtOrAbove(sortedSizes, curve, .80),
		R90: firstAtOrAbove(sortedSizes, curve, .90),
		R95: firstAtOrAbove(sortedSizes, curve, .95),
	}
}

func firstAtOrAbove(sortedSizes []int, curve map[int]float64, level float64) *int {
	for _, size := range sortedSizes {
		if curve[size] >= level {
			found := size
			return &found
		}
	}
	return nil
}
