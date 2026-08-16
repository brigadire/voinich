package conditionalregime

import (
	"math/rand"
	"reflect"
	"testing"

	"zcore.dev/voinich/internal/globalregime"
)

// referenceFitClustering is fitClustering exactly as it stood before the
// withinFitPrep hoist: it recomputes full, sample, and the distance matrix
// on every call, even across K/method values that share the same window
// set.
func referenceFitClustering(all []classWindow, method string, k int, seed int64) (fitLabels []int, fullLabels []int, sampleD [][]float64) {
	full := plainWindows(all)
	sample, _ := globalregime.ClusteringSample(full)
	sampleD = globalregime.DistanceMatrix(sample)
	switch method {
	case "hierarchical":
		fitLabels = globalregime.HierarchicalLabels(len(sample), k, sampleD)
	default:
		fitLabels = globalregime.KMedoids(sampleD, k, seed)
	}
	fullLabels = globalregime.ExpandLabels(full, sample, fitLabels, k)
	return fitLabels, fullLabels, sampleD
}

// classWindowFixture builds a synthetic class window set spanning nBlocks
// blocks of blockLen tokens each, drawn from a small shared vocabulary so
// windows are non-degenerate but not identical, large enough at blockLen to
// exercise both globalregime's clusteringSample no-op path (<=200 windows)
// and its sampled path (>200 windows) depending on the caller's choice of
// blockLen/windowSize.
func classWindowFixture(nBlocks, blockLen, windowSize int, seed int64) []classWindow {
	r := rand.New(rand.NewSource(seed))
	vocab := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	tokens := make([]string, 0, nBlocks*blockLen)
	blocks := make([]Block, nBlocks)
	for bi := 0; bi < nBlocks; bi++ {
		start := len(tokens)
		for i := 0; i < blockLen; i++ {
			tokens = append(tokens, vocab[r.Intn(len(vocab))])
		}
		blocks[bi] = Block{Class: ClassID{Scheme: SchemeJoint, Currier: "A", Hand: "1"}, Index: bi, Start: start, End: len(tokens)}
	}
	return buildClassWindows(tokens, blocks, windowSize)
}

// TestFitClusteringHoistMatchesReference proves the withinFitPrep hoist
// produces byte-identical fitLabels/fullLabels/sampleD to the pre-hoist
// reference, across a window count below globalregime's 200-window
// clustering-sample cap and one above it, both methods, several K values,
// and several seeds.
func TestFitClusteringHoistMatchesReference(t *testing.T) {
	fixtures := map[string][]classWindow{
		"small (<=200 windows, no sampling)": classWindowFixture(3, 50, 10, 1),
		"large (>200 windows, sampled)":      classWindowFixture(3, 300, 10, 2),
	}
	for name, cw := range fixtures {
		t.Run(name, func(t *testing.T) {
			if len(cw) == 0 {
				t.Fatalf("fixture %q produced no windows", name)
			}
			prep := prepareWithinFit(cw)
			for _, method := range []string{"hierarchical", "k_medoids"} {
				for k := 2; k <= 5; k++ {
					for seed := int64(0); seed < 5; seed++ {
						wantFit, wantFull, wantD := referenceFitClustering(cw, method, k, seed)
						gotFit, gotFull, gotD := fitClustering(prep, method, k, seed)
						if !reflect.DeepEqual(gotFit, wantFit) {
							t.Fatalf("%s method=%s k=%d seed=%d: fitLabels diverged\ngot=%v\nwant=%v", name, method, k, seed, gotFit, wantFit)
						}
						if !reflect.DeepEqual(gotFull, wantFull) {
							t.Fatalf("%s method=%s k=%d seed=%d: fullLabels diverged\ngot=%v\nwant=%v", name, method, k, seed, gotFull, wantFull)
						}
						if !reflect.DeepEqual(gotD, wantD) {
							t.Fatalf("%s method=%s k=%d seed=%d: sampleD diverged", name, method, k, seed)
						}
					}
				}
			}
		})
	}
}

// TestPrepareWithinFitIsInvariantAcrossCalls proves the specific claim the
// hoist depends on: prepareWithinFit's output for a fixed window set does
// not vary from call to call, which is exactly why computing it once
// outside the (method, K) loop is safe.
func TestPrepareWithinFitIsInvariantAcrossCalls(t *testing.T) {
	cw := classWindowFixture(3, 300, 10, 3)
	first := prepareWithinFit(cw)
	for i := 0; i < 5; i++ {
		again := prepareWithinFit(cw)
		if !reflect.DeepEqual(first, again) {
			t.Fatalf("prepareWithinFit is not a pure function of the window set: call %d diverged", i)
		}
	}
}

// benchWithinFixture mirrors a realistic within-class sweep shape: enough
// windows to exercise globalregime's sampling path.
func benchWithinFixture() []classWindow {
	return classWindowFixture(4, 300, 10, 42)
}

func BenchmarkFitClusteringReferenceSweep(b *testing.B) {
	cw := benchWithinFixture()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, method := range []string{"k_medoids", "hierarchical"} {
			for k := 2; k <= 10; k++ {
				referenceFitClustering(cw, method, k, int64(i))
			}
		}
	}
}

func BenchmarkFitClusteringHoistedSweep(b *testing.B) {
	cw := benchWithinFixture()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		prep := prepareWithinFit(cw)
		for _, method := range []string{"k_medoids", "hierarchical"} {
			for k := 2; k <= 10; k++ {
				fitClustering(prep, method, k, int64(i))
			}
		}
	}
}
