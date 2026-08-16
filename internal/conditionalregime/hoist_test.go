package conditionalregime

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"zcore.dev/voinich/internal/globalregime"
)

// referenceFitResidualClustering is fitResidualClustering exactly as it
// stood before the residualFitPrep hoist: it recomputes vecs, sampleIdx,
// sampleVecs, and the O(fitCap^2) sampleD distance matrix on every call,
// even across K values that share the same (rw, standardized, fitCap).
func referenceFitResidualClustering(rw []ResidualWindow, standardized bool, method string, k int, seed int64, fitCap int) (fitLabels, fullLabels []int, sampleD [][]float64) {
	vecs := denseResidualVectors(rw, standardized)
	sampleIdx := cappedSampleIndices(len(vecs), fitCap)
	sampleVecs := make([]denseVector, len(sampleIdx))
	for i, si := range sampleIdx {
		sampleVecs[i] = vecs[si]
	}
	sampleD = residualDistanceMatrix(sampleVecs)
	if method == "hierarchical" {
		fitLabels = globalregime.HierarchicalLabels(len(sampleVecs), k, sampleD)
	} else {
		fitLabels = globalregime.KMedoids(sampleD, k, seed)
	}
	fullLabels = expandResidualLabels(vecs, sampleIdx, fitLabels, k)
	return fitLabels, fullLabels, sampleD
}

// residualWindowFixture builds a synthetic set of ResidualWindows with
// varied, overlapping-but-not-identical sparse Residual/Standard vectors
// (so distances are non-degenerate), large enough to exercise both
// fitCap-limited sampling (fitCap < len(rw)) and the no-sampling-needed
// case (fitCap >= len(rw)).
func residualWindowFixture(n int) []ResidualWindow {
	r := rand.New(rand.NewSource(int64(n)*97 + 13))
	out := make([]ResidualWindow, n)
	for i := 0; i < n; i++ {
		res := vector{}
		std := vector{}
		for t := 0; t < 8; t++ {
			tok := fmt.Sprintf("tok%d", (i+t)%12)
			v := r.NormFloat64()
			res[tok] = v
			std[tok] = v * 1.5
		}
		out[i] = ResidualWindow{
			Class:      ClassID{Scheme: SchemeJoint, Currier: "A", Hand: "1"},
			BlockIndex: i % 4, WindowSize: 300,
			Residual: res, Standard: std,
			AbsStart: i * 300, AbsEnd: (i + 1) * 300,
		}
	}
	return out
}

// TestFitResidualClusteringHoistMatchesReference proves the
// residualFitPrep hoist produces byte-identical fitLabels/fullLabels/
// sampleD to the pre-hoist reference, across both representations
// (raw/standardized), both methods, several K values, several seeds, and
// both a fitCap smaller than and larger than the window count.
func TestFitResidualClusteringHoistMatchesReference(t *testing.T) {
	sizes := []int{20, 60}
	fitCaps := []int{10, 1000} // 1000 exceeds every fixture size: exercises the no-sampling-needed path
	for _, n := range sizes {
		rw := residualWindowFixture(n)
		for _, fitCap := range fitCaps {
			for _, standardized := range []bool{false, true} {
				prep := prepareResidualFit(rw, standardized, fitCap)
				for _, method := range []string{"hierarchical", "k_medoids"} {
					for k := 2; k <= 5; k++ {
						for seed := int64(0); seed < 5; seed++ {
							wantFit, wantFull, wantD := referenceFitResidualClustering(rw, standardized, method, k, seed, fitCap)
							gotFit, gotFull, gotD := fitResidualClustering(prep, method, k, seed)
							if !reflect.DeepEqual(gotFit, wantFit) {
								t.Fatalf("n=%d fitCap=%d std=%v method=%s k=%d seed=%d: fitLabels diverged\ngot=%v\nwant=%v", n, fitCap, standardized, method, k, seed, gotFit, wantFit)
							}
							if !reflect.DeepEqual(gotFull, wantFull) {
								t.Fatalf("n=%d fitCap=%d std=%v method=%s k=%d seed=%d: fullLabels diverged\ngot=%v\nwant=%v", n, fitCap, standardized, method, k, seed, gotFull, wantFull)
							}
							if !reflect.DeepEqual(gotD, wantD) {
								t.Fatalf("n=%d fitCap=%d std=%v method=%s k=%d seed=%d: sampleD diverged", n, fitCap, standardized, method, k, seed)
							}
						}
					}
				}
			}
		}
	}
}

// TestPrepareResidualFitIsInvariantAcrossK proves the specific claim the
// hoist depends on: prepareResidualFit's output for a fixed (rw,
// standardized, fitCap) does not vary — it's the same value whichever K a
// caller is about to fit, which is exactly why computing it once outside
// the K loop is safe.
func TestPrepareResidualFitIsInvariantAcrossK(t *testing.T) {
	rw := residualWindowFixture(40)
	first := prepareResidualFit(rw, false, 15)
	for k := 0; k < 5; k++ {
		again := prepareResidualFit(rw, false, 15)
		if !reflect.DeepEqual(first, again) {
			t.Fatalf("prepareResidualFit is not a pure function of (rw, standardized, fitCap): call %d diverged", k)
		}
	}
}

// benchResidualFixture mirrors a realistic residual-sweep shape: many
// pooled windows across a scale, sampled down via fitCap for clustering.
func benchResidualFixture() ([]ResidualWindow, int) {
	return residualWindowFixture(400), maxResidualFitWindows
}

func BenchmarkFitResidualClusteringReferenceKSweep(b *testing.B) {
	rw, fitCap := benchResidualFixture()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for k := 2; k <= 15; k++ {
			referenceFitResidualClustering(rw, false, "k_medoids", k, int64(i), fitCap)
		}
	}
}

func BenchmarkFitResidualClusteringHoistedKSweep(b *testing.B) {
	rw, fitCap := benchResidualFixture()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		prep := prepareResidualFit(rw, false, fitCap)
		for k := 2; k <= 15; k++ {
			fitResidualClustering(prep, "k_medoids", k, int64(i))
		}
	}
}
