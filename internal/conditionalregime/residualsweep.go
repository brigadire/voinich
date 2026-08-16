package conditionalregime

import (
	"context"
	"math"
	"math/rand"

	"zcore.dev/voinich/internal/globalregime"
)

// ResidualClusterSummary is one row of residual_cluster_summary.tsv: one
// scale x method x K x representation clustering over the pooled residual
// windows of every eligible class.
type ResidualClusterSummary struct {
	WindowSize          int
	Method              string
	K                   int
	Representation      string // "raw" (primary) or "standardized" (sensitivity)
	Windows             int
	Silhouette          float64
	WithinDispersion    float64
	BetweenDispersion   float64
	ClusterSizeEntropy  float64
	SmallestClusterFrac float64
}

func residualDiagnosticRow(windowSize int, method string, k int, representation string, n int, fitLabels []int, d [][]float64, fullLabels []int) ResidualClusterSummary {
	diag := globalregime.Diagnostics(windowSize, method, k, fitLabels, d)
	diag = globalregime.WithFullAssignments(diag, fullLabels)
	return ResidualClusterSummary{
		WindowSize: windowSize, Method: method, K: k, Representation: representation, Windows: n,
		Silhouette: diag.Silhouette, WithinDispersion: diag.WithinDispersion, BetweenDispersion: diag.BetweenDistance,
		ClusterSizeEntropy: clusterSizeEntropy(diag.ClusterSizes), SmallestClusterFrac: smallestClusterFraction(diag.ClusterSizes),
	}
}

// residualSweepResult is the outcome of clustering every scale x K
// combination in the new frozen residual search space for one method and
// representation.
type residualSweepResult struct {
	Rows           []ResidualClusterSummary
	BestWindowSize int
	BestK          int
	BestSilhouette float64
	BestFullLabels []int
	BestWindows    []ResidualWindow
}

// residualSweep clusters every (scale, K) in the frozen residual search
// space (task19 section 29) for one method/representation and returns the
// combination with the highest silhouette, exactly like the global max
// selection in the metadata multiple-comparison correction.
func residualSweep(tokens []string, classes []ClassID, blocksByClass map[ClassID][]Block, scales []int, kMin, kMax int, method string, standardized bool, seed int64) residualSweepResult {
	res := residualSweepResult{BestSilhouette: math.Inf(-1)}
	for _, scale := range scales {
		rw := buildResidualWindows(tokens, classes, blocksByClass, scale)
		if len(rw) < 2*kMin {
			continue
		}
		prep := prepareResidualFit(rw, standardized, maxResidualFitWindows)
		for k := kMin; k <= kMax; k++ {
			if 2*k > len(rw) {
				continue
			}
			fitLabels, fullLabels, d := fitResidualClustering(prep, method, k, seed+int64(scale))
			row := residualDiagnosticRow(scale, method, k, representationName(standardized), len(rw), fitLabels, d, fullLabels)
			res.Rows = append(res.Rows, row)
			if row.Silhouette > res.BestSilhouette {
				res.BestSilhouette, res.BestWindowSize, res.BestK = row.Silhouette, scale, k
				res.BestFullLabels, res.BestWindows = fullLabels, rw
			}
		}
	}
	return res
}

// residualSweepProgress is residualSweep with a callback invoked once per
// completed (scale, K) fit, for status-bar reporting.
func residualSweepProgress(tokens []string, classes []ClassID, blocksByClass map[ClassID][]Block, scales []int, kMin, kMax int, method string, standardized bool, seed int64, onStep func(n int)) residualSweepResult {
	res := residualSweepResult{BestSilhouette: math.Inf(-1)}
	for _, scale := range scales {
		rw := buildResidualWindows(tokens, classes, blocksByClass, scale)
		if len(rw) < 2*kMin {
			if onStep != nil {
				onStep(kMax - kMin + 1)
			}
			continue
		}
		prep := prepareResidualFit(rw, standardized, maxResidualFitWindows)
		for k := kMin; k <= kMax; k++ {
			if 2*k > len(rw) {
				if onStep != nil {
					onStep(1)
				}
				continue
			}
			fitLabels, fullLabels, d := fitResidualClustering(prep, method, k, seed+int64(scale))
			row := residualDiagnosticRow(scale, method, k, representationName(standardized), len(rw), fitLabels, d, fullLabels)
			res.Rows = append(res.Rows, row)
			if row.Silhouette > res.BestSilhouette {
				res.BestSilhouette, res.BestWindowSize, res.BestK = row.Silhouette, scale, k
				res.BestFullLabels, res.BestWindows = fullLabels, rw
			}
			if onStep != nil {
				onStep(1)
			}
		}
	}
	return res
}

func representationName(standardized bool) string {
	if standardized {
		return "standardized"
	}
	return "raw"
}

// shuffleAllClassesCompact applies Null A across every eligible class's
// physical blocks, copying only their own tokens (not the whole corpus) into
// a small shared buffer, using one deterministic random stream so a single
// call produces one consistent shuffled-corpus realization that is then
// reused, unchanged, across the entire scale x K residual search space for
// that replicate (the same "one permutation, whole search space" principle
// used by the frozen Currier/hand multiple-comparison correction). classes
// must be in a fixed deterministic order.
func shuffleAllClassesCompact(tokens []string, classes []ClassID, blocksByClass map[ClassID][]Block, rng *rand.Rand) (buf []string, rebased map[ClassID][]Block) {
	rebased = make(map[ClassID][]Block, len(classes))
	for _, class := range classes {
		for _, b := range blocksByClass[class] {
			start := len(buf)
			buf = append(buf, tokens[b.Start:b.End]...)
			seg := buf[start:]
			rng.Shuffle(len(seg), func(i, j int) { seg[i], seg[j] = seg[j], seg[i] })
			rebased[class] = append(rebased[class], Block{Class: b.Class, Index: b.Index, Start: start, End: len(buf)})
		}
	}
	return buf, rebased
}

// residualNullMax draws one Null-A realization and returns the maximum
// silhouette over the complete frozen scale x K residual search space for
// one method/representation (task19 section 41).
func residualNullMax(tokens []string, classes []ClassID, blocksByClass map[ClassID][]Block, scales []int, kMin, kMax int, method string, standardized bool, rng *rand.Rand) float64 {
	buf, rebased := shuffleAllClassesCompact(tokens, classes, blocksByClass, rng)
	best := math.Inf(-1)
	for _, scale := range scales {
		rw := buildResidualWindows(buf, classes, rebased, scale)
		if len(rw) < 2*kMin {
			continue
		}
		prep := prepareResidualFit(rw, standardized, nullFitCap)
		for k := kMin; k <= kMax; k++ {
			if 2*k > len(rw) {
				continue
			}
			fitLabels, _, d := fitResidualClustering(prep, method, k, rng.Int63())
			sil := globalregime.Diagnostics(scale, method, k, fitLabels, d).Silhouette
			if sil > best {
				best = sil
			}
		}
	}
	if math.IsInf(best, -1) {
		return 0
	}
	return best
}

// residualGlobalCorrection runs the primary permutation loop backing
// residual_permutations.yaml: the global max-over-scale x K statistic for
// one method/representation, against `permutations` Null-A replicates. This
// is the single most expensive loop in the pipeline, so every replicate uses
// its own independently seeded rng (replicateSeed) and resume lets a caller
// supply already-computed replicates (e.g. reloaded from a checkpoint) to
// continue from, rather than recomputing them. onSave, if non-nil, is
// invoked after every new replicate with the null slice accumulated so far,
// so a caller can checkpoint after each one.
func residualGlobalCorrection(tokens []string, classes []ClassID, blocksByClass map[ClassID][]Block, scales []int, kMin, kMax int, method string, standardized bool, observed float64, permutations int, seed int64, resume []float64, onSave func(null []float64)) EmpiricalStats {
	stats, _ := residualGlobalCorrectionParallel(context.Background(), 1, nil, tokens, classes, blocksByClass, scales, kMin, kMax, method, standardized, observed, permutations, seed, resume, onSave)
	return stats
}

func residualGlobalCorrectionParallel(ctx context.Context, workers int, pool *processPool, tokens []string, classes []ClassID, blocksByClass map[ClassID][]Block, scales []int, kMin, kMax int, method string, standardized bool, observed float64, permutations int, seed int64, resume []float64, onSave func(null []float64)) (EmpiricalStats, error) {
	return residualGlobalCorrectionParallelState(ctx, workers, pool, tokens, classes, blocksByClass, scales, kMin, kMax, method, standardized, observed, permutations, seed, resume, nil, onSave, nil)
}

func residualGlobalCorrectionParallelState(ctx context.Context, workers int, pool jobExecutor, tokens []string, classes []ClassID, blocksByClass map[ClassID][]Block, scales []int, kMin, kMax int, method string, standardized bool, observed float64, permutations int, seed int64, resume []float64, completed map[int]float64, onSave func(null []float64), onComplete func(JobResult)) (EmpiricalStats, error) {
	salt := methodSalt(method)
	null, err := runIndexedReplicatesState(ctx, workers, pool, "part_b_global_correction", method+"|"+representationName(standardized), permutations, resume, completed, func(ctx context.Context, i int) (float64, error) {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		rng := rand.New(rand.NewSource(replicateSeed(seed, salt, i)))
		return residualNullMax(tokens, classes, blocksByClass, scales, kMin, kMax, method, standardized, rng), nil
	}, onSave, onComplete)
	if err != nil {
		return EmpiricalStats{}, err
	}
	return buildEmpiricalStats(observed, null), nil
}
