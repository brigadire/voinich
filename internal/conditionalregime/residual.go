package conditionalregime

import (
	"math"

	"zcore.dev/voinich/internal/globalregime"
)

// vector is a sparse residual feature vector: a token-frequency window
// profile minus its class's expected profile. Unlike a probability profile
// it may be negative, so it is compared with Euclidean distance rather than
// the Jensen-Shannon distance global-regime-analyze uses for raw profiles.
// This is the one deliberate, minimal representation change task19 Part B
// requires (removing the class signature); no new token features are added.
type vector map[string]float64

func subtractProfile(x, mean globalregime.Profile) vector {
	r := make(vector, len(x)+len(mean))
	for tok, v := range x {
		r[tok] = v - mean[tok]
	}
	for tok, m := range mean {
		if _, ok := x[tok]; !ok {
			r[tok] = -m
		}
	}
	return r
}

// standardize divides a residual by its class's per-token standard
// deviation, regularized so near-zero-variance tokens do not explode the
// scale (task19 section 27).
func standardize(r vector, variance globalregime.Profile, floor float64) vector {
	z := make(vector, len(r))
	for tok, v := range r {
		sd := math.Sqrt(variance[tok])
		if sd < floor {
			sd = floor
		}
		z[tok] = v / sd
	}
	return z
}

// meanAndVarianceProfiles computes the training-fold mean and variance
// feature vector for a set of windows. Tokens absent from a given window
// contribute zero, exactly as in the underlying probability profile.
func meanAndVarianceProfiles(ws []classWindow) (mean, variance globalregime.Profile) {
	mean, variance = globalregime.Profile{}, globalregime.Profile{}
	n := len(ws)
	if n == 0 {
		return
	}
	for _, w := range ws {
		for tok, v := range w.Distribution() {
			mean[tok] += v
		}
	}
	for tok := range mean {
		mean[tok] /= float64(n)
	}
	for tok, m := range mean {
		s := 0.0
		for _, w := range ws {
			d := w.Distribution()[tok] - m
			s += d * d
		}
		variance[tok] = s / float64(n)
	}
	return mean, variance
}

// ResidualWindow is one out-of-fold residualized window: X_w with its
// class's fold-training-only expected feature vector removed. Every
// ResidualWindow was held out of the fold that estimated its own centering
// statistics, so there is no metadata leakage (task19 sections 25-28).
type ResidualWindow struct {
	Class      ClassID
	BlockIndex int // index of the class's physical block this window came from
	WindowSize int
	Residual   vector // R_w, primary representation
	Standard   vector // Z_w, sensitivity representation
	AbsStart   int
	AbsEnd     int
}

// varianceFloor regularizes standardization against near-zero-variance
// tokens (task19 section 27). Fixed in advance, not tuned on results.
const varianceFloor = 1e-6

// buildResidualWindows residualizes every eligible class's windows at one
// scale using leave-one-block-out (or, for single-block classes,
// three-way contiguous) folds, so every window's centering statistics come
// only from blocks that do not contain it.
func buildResidualWindows(tokens []string, classes []ClassID, blocksByClass map[ClassID][]Block, windowSize int) []ResidualWindow {
	var out []ResidualWindow
	for _, class := range classes {
		blocks := blocksByClass[class]
		for _, fold := range contiguousFolds(blocks) {
			trainW := buildClassWindows(tokens, fold.train, windowSize)
			testW := buildClassWindows(tokens, fold.test, windowSize)
			if len(trainW) == 0 || len(testW) == 0 {
				continue
			}
			mean, variance := meanAndVarianceProfiles(trainW)
			for _, w := range testW {
				r := subtractProfile(w.Distribution(), mean)
				out = append(out, ResidualWindow{
					Class: class, BlockIndex: fold.test[0].Index, WindowSize: windowSize,
					Residual: r, Standard: standardize(r, variance, varianceFloor),
					AbsStart: w.AbsStart, AbsEnd: w.AbsEnd,
				})
			}
		}
	}
	return out
}

func euclideanDistance(a, b vector) float64 {
	sum := 0.0
	seen := make(map[string]bool, len(a))
	for tok, v := range a {
		d := v - b[tok]
		sum += d * d
		seen[tok] = true
	}
	for tok, v := range b {
		if seen[tok] {
			continue
		}
		sum += v * v
	}
	return math.Sqrt(sum)
}

func residualDistanceMatrix(vecs []vector) [][]float64 {
	n := len(vecs)
	d := make([][]float64, n)
	for i := range d {
		d[i] = make([]float64, n)
	}
	for i := 0; i < n; i++ {
		for j := 0; j < i; j++ {
			v := euclideanDistance(vecs[i], vecs[j])
			d[i][j], d[j][i] = v, v
		}
	}
	return d
}

// maxResidualFitWindows bounds quadratic clustering cost, mirroring
// global-regime-analyze's sampling strategy for long sequences.
const maxResidualFitWindows = 200

func residualCentroids(vecs []vector, sampleIdx, sampleLabels []int, k int) []vector {
	centroids := make([]vector, k)
	counts := make([]int, k)
	for c := range centroids {
		centroids[c] = vector{}
	}
	for i, si := range sampleIdx {
		c := sampleLabels[i]
		counts[c]++
		for tok, v := range vecs[si] {
			centroids[c][tok] += v
		}
	}
	for c := range centroids {
		if counts[c] == 0 {
			continue
		}
		for tok := range centroids[c] {
			centroids[c][tok] /= float64(counts[c])
		}
	}
	return centroids
}

func expandResidualLabels(vecs []vector, sampleIdx, sampleLabels []int, k int) []int {
	centroids := residualCentroids(vecs, sampleIdx, sampleLabels, k)
	haveCount := make([]bool, k)
	for _, c := range sampleLabels {
		haveCount[c] = true
	}
	labels := make([]int, len(vecs))
	for i, v := range vecs {
		best, bd := -1, math.Inf(1)
		for c := 0; c < k; c++ {
			if !haveCount[c] {
				continue
			}
			d := euclideanDistance(v, centroids[c])
			if d < bd {
				best, bd = c, d
			}
		}
		labels[i] = best
	}
	return labels
}

// residualVectors selects the raw or standardized representation for a set
// of residual windows.
func residualVectors(rw []ResidualWindow, standardized bool) []vector {
	out := make([]vector, len(rw))
	for i, w := range rw {
		if standardized {
			out[i] = w.Standard
		} else {
			out[i] = w.Residual
		}
	}
	return out
}

// fitResidualClustering fits one (method, K) clustering on a deterministic
// sample of the pooled residual windows (at most fitCap of them) and expands
// it to every window. The observed sweep uses maxResidualFitWindows; the
// permutation loop uses a smaller cap to stay tractable across thousands of
// replicates (task19 section 41's own null distribution does not need the
// same fitting-sample resolution as the single observed statistic).
func fitResidualClustering(rw []ResidualWindow, standardized bool, method string, k int, seed int64, fitCap int) (fitLabels, fullLabels []int, sampleD [][]float64) {
	vecs := residualVectors(rw, standardized)
	sampleIdx := cappedSampleIndices(len(vecs), fitCap)
	sampleVecs := make([]vector, len(sampleIdx))
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
