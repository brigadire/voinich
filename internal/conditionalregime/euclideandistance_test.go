package conditionalregime

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"
)

// referenceEuclideanDistance is euclideanDistance exactly as it stood before
// the sortedVector merge-walk optimization (task27): it re-sorts the union
// of a and b's keys on every call. Profiling conditional-regime-analyze at
// production scale showed this sort was over 70% of the CLI's total CPU
// time, since it re-sorted the same vectors' keys on every one of the
// O(fitCap^2) distance-matrix pairs and every one of the O(n*k)
// centroid-assignment calls, across every K in the sweep and every
// permutation replicate.
func referenceEuclideanDistance(a, b vector) float64 {
	seen := make(map[string]bool, len(a)+len(b))
	keys := make([]string, 0, len(a)+len(b))
	for tok := range a {
		seen[tok] = true
		keys = append(keys, tok)
	}
	for tok := range b {
		if !seen[tok] {
			keys = append(keys, tok)
		}
	}
	sort.Strings(keys)
	sum := 0.0
	for _, tok := range keys {
		d := a[tok] - b[tok]
		sum += d * d
	}
	return math.Sqrt(sum)
}

// randomSparseVector builds a synthetic sparse vector over tokens named
// "t000".."t{n-1}", each included independently with probability keep, with
// values (including negatives, since residuals are signed) drawn from r.
func randomSparseVector(r *rand.Rand, n int, keep float64) vector {
	v := vector{}
	for i := 0; i < n; i++ {
		if r.Float64() < keep {
			v[fmt.Sprintf("t%03d", i)] = r.NormFloat64() * 10
		}
	}
	return v
}

// TestEuclideanDistanceMatchesReference proves the sortedVector merge-walk
// implementation is byte-identical to the pre-optimization
// sort-the-union-every-call reference, across disjoint, fully overlapping,
// partially overlapping, and empty-vector fixtures, plus randomized fixtures
// spanning small and large key sets.
func TestEuclideanDistanceMatchesReference(t *testing.T) {
	cases := []struct {
		name string
		a, b vector
	}{
		{"both empty", vector{}, vector{}},
		{"a empty", vector{}, vector{"x": 1, "y": -2}},
		{"b empty", vector{"x": 1, "y": -2}, vector{}},
		{"disjoint", vector{"a": 1, "b": 2}, vector{"c": 3, "d": 4}},
		{"identical keys", vector{"a": 1, "b": -2, "c": 3.5}, vector{"a": 1.5, "b": -2, "c": 0}},
		{"partial overlap", vector{"a": 1, "b": 2, "c": 3}, vector{"b": 2.5, "c": 3, "d": 4}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			want := referenceEuclideanDistance(c.a, c.b)
			got := euclideanDistance(sortVector(c.a), sortVector(c.b))
			if math.Float64bits(got) != math.Float64bits(want) {
				t.Fatalf("got %v, want %v (bits %x vs %x)", got, want, math.Float64bits(got), math.Float64bits(want))
			}
		})
	}

	sizes := []int{1, 5, 50, 400}
	keeps := []float64{0.1, 0.5, 0.9, 1.0}
	for _, n := range sizes {
		for _, keep := range keeps {
			for seed := int64(0); seed < 5; seed++ {
				r := rand.New(rand.NewSource(seed*1000 + int64(n)))
				a := randomSparseVector(r, n, keep)
				b := randomSparseVector(r, n, keep)
				want := referenceEuclideanDistance(a, b)
				got := euclideanDistance(sortVector(a), sortVector(b))
				if math.Float64bits(got) != math.Float64bits(want) {
					t.Fatalf("n=%d keep=%v seed=%d: got %v, want %v", n, keep, seed, got, want)
				}
			}
		}
	}
}

// benchmarkVectors returns pairsN sortedVectors and their raw-vector
// counterparts, sized and sparsified to mirror a realistic residual window
// (a few hundred distinct tokens out of a larger vocabulary).
func benchmarkVectors(pairsN, vocab int, keep float64) ([]vector, []sortedVector) {
	r := rand.New(rand.NewSource(7))
	raw := make([]vector, pairsN)
	sorted := make([]sortedVector, pairsN)
	for i := range raw {
		raw[i] = randomSparseVector(r, vocab, keep)
		sorted[i] = sortVector(raw[i])
	}
	return raw, sorted
}

func BenchmarkEuclideanDistanceReferenceMatrix(b *testing.B) {
	raw, _ := benchmarkVectors(200, 800, 0.3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for x := 0; x < len(raw); x++ {
			for y := 0; y < x; y++ {
				referenceEuclideanDistance(raw[x], raw[y])
			}
		}
	}
}

func BenchmarkEuclideanDistanceSortedMatrix(b *testing.B) {
	_, sorted := benchmarkVectors(200, 800, 0.3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for x := 0; x < len(sorted); x++ {
			for y := 0; y < x; y++ {
				euclideanDistance(sorted[x], sorted[y])
			}
		}
	}
}
