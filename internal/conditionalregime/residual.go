package conditionalregime

import (
	"math"
	"sort"

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

// denseVector stores residual features in one deterministic, lexicographically
// ordered feature space shared by every vector in a residualFitPrep. Residual
// maps are converted once per (scale, replicate), before either the pairwise
// distance matrix or the K sweep. The ordering is deliberately the same as
// the former sorted-key union traversal, preserving floating-point summation
// order while removing string comparisons and map lookups from the hot loop.
type denseVector []float64

// euclideanDistance sums (a[i]-b[i])^2 in deterministic feature-index order.
// All callers pass vectors from the same denseResidualVectors conversion, so
// their lengths and feature meanings are identical. It performs no allocation.
func euclideanDistance(a, b denseVector) float64 {
	sum := 0.0
	for i, av := range a {
		d := av - b[i]
		sum += d * d
	}
	return math.Sqrt(sum)
}

func residualDistanceMatrix(vecs []denseVector) [][]float64 {
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

func residualCentroids(vecs []denseVector, sampleIdx, sampleLabels []int, k int) []denseVector {
	centroids := make([]denseVector, k)
	counts := make([]int, k)
	dimensions := 0
	if len(vecs) > 0 {
		dimensions = len(vecs[0])
	}
	for c := range centroids {
		centroids[c] = make(denseVector, dimensions)
	}
	for i, si := range sampleIdx {
		c := sampleLabels[i]
		counts[c]++
		for feature, v := range vecs[si] {
			centroids[c][feature] += v
		}
	}
	for c := range centroids {
		if counts[c] == 0 {
			continue
		}
		for feature := range centroids[c] {
			centroids[c][feature] /= float64(counts[c])
		}
	}
	return centroids
}

func expandResidualLabels(vecs []denseVector, sampleIdx, sampleLabels []int, k int) []int {
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

// denseResidualVectors establishes one deterministic feature index for a set
// of residual windows, then converts every selected sparse map exactly once.
// The returned vectors all share the same feature ordering and can be reused
// throughout distance-matrix construction and the complete K sweep.
func denseResidualVectors(rw []ResidualWindow, standardized bool) []denseVector {
	features := make(map[string]struct{})
	for _, w := range rw {
		v := w.Residual
		if standardized {
			v = w.Standard
		}
		for tok := range v {
			features[tok] = struct{}{}
		}
	}
	keys := make([]string, 0, len(features))
	for tok := range features {
		keys = append(keys, tok)
	}
	sort.Strings(keys)
	featureIndex := make(map[string]int, len(keys))
	for i, tok := range keys {
		featureIndex[tok] = i
	}

	out := make([]denseVector, len(rw))
	for i, w := range rw {
		out[i] = make(denseVector, len(keys))
		v := w.Residual
		if standardized {
			v = w.Standard
		}
		for tok, value := range v {
			out[i][featureIndex[tok]] = value
		}
	}
	return out
}

// residualFitPrep holds the (rw, standardized, fitCap)-derived sample and
// its distance matrix: everything fitResidualClustering used to recompute
// on every call, even though none of it depends on k or the clustering
// method. cappedSampleIndices is a deterministic (non-random) even-spacing
// selection and residualDistanceMatrix/denseResidualVectors do no RNG draws, so
// prepareResidualFit's result is exactly invariant across every K value in
// a kMin..kMax sweep for a fixed (scale, replicate) — see
// determinism_test.go for the equivalence proof. globalregime's
// HierarchicalLabels/KMedoids/Diagnostics only ever read their d
// parameter, never write it, so the same sampleD can be safely shared
// across every K's clustering call without one K's fit corrupting
// another's.
type residualFitPrep struct {
	vecs       []denseVector
	sampleIdx  []int
	sampleVecs []denseVector
	sampleD    [][]float64
}

// prepareResidualFit computes the part of fitResidualClustering that does
// not depend on k: call this once per (scale, replicate) before the K
// loop, not once per K.
func prepareResidualFit(rw []ResidualWindow, standardized bool, fitCap int) residualFitPrep {
	vecs := denseResidualVectors(rw, standardized)
	sampleIdx := cappedSampleIndices(len(vecs), fitCap)
	sampleVecs := make([]denseVector, len(sampleIdx))
	for i, si := range sampleIdx {
		sampleVecs[i] = vecs[si]
	}
	return residualFitPrep{vecs: vecs, sampleIdx: sampleIdx, sampleVecs: sampleVecs, sampleD: residualDistanceMatrix(sampleVecs)}
}

// fitResidualClustering fits one (method, K) clustering on prep's
// precomputed sample and distance matrix, and expands it to every window.
func fitResidualClustering(prep residualFitPrep, method string, k int, seed int64) (fitLabels, fullLabels []int, sampleD [][]float64) {
	if method == "hierarchical" {
		fitLabels = globalregime.HierarchicalLabels(len(prep.sampleVecs), k, prep.sampleD)
	} else {
		fitLabels = globalregime.KMedoids(prep.sampleD, k, seed)
	}
	fullLabels = expandResidualLabels(prep.vecs, prep.sampleIdx, fitLabels, k)
	return fitLabels, fullLabels, prep.sampleD
}
