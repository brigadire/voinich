package conditionalregime

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"
)

// referenceEuclideanDistance is the original deterministic sparse-map
// calculation. It is retained as a bit-exact oracle for the dense rewrite.
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

// densePair converts two vectors using the same lexicographically ordered
// feature mapping denseResidualVectors establishes for a whole prep.
func densePair(a, b vector) (denseVector, denseVector) {
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
	da, db := make(denseVector, len(keys)), make(denseVector, len(keys))
	for i, tok := range keys {
		da[i], db[i] = a[tok], b[tok]
	}
	return da, db
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

// TestEuclideanDistanceMatchesReference proves the dense implementation is
// byte-identical to the sorted sparse-map reference, across disjoint, fully
// overlapping, partially overlapping, and empty-vector fixtures, plus
// randomized fixtures spanning small and large key sets.
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
			da, db := densePair(c.a, c.b)
			got := euclideanDistance(da, db)
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
				da, db := densePair(a, b)
				got := euclideanDistance(da, db)
				if math.Float64bits(got) != math.Float64bits(want) {
					t.Fatalf("n=%d keep=%v seed=%d: got %v, want %v", n, keep, seed, got, want)
				}
			}
		}
	}
}

// benchmarkVectors returns sparse reference vectors and equivalent dense
// vectors, sized to mirror residual windows.
func benchmarkVectors(pairsN, vocab int, keep float64) ([]vector, []denseVector) {
	r := rand.New(rand.NewSource(7))
	raw := make([]vector, pairsN)
	dense := make([]denseVector, pairsN)
	for i := range raw {
		raw[i] = randomSparseVector(r, vocab, keep)
		dense[i] = make(denseVector, vocab)
		for feature := 0; feature < vocab; feature++ {
			dense[i][feature] = raw[i][fmt.Sprintf("t%03d", feature)]
		}
	}
	return raw, dense
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

func BenchmarkEuclideanDistanceDenseMatrix(b *testing.B) {
	_, dense := benchmarkVectors(200, 800, 0.3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for x := 0; x < len(dense); x++ {
			for y := 0; y < x; y++ {
				euclideanDistance(dense[x], dense[y])
			}
		}
	}
}
