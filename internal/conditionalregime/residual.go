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

// sortedVector pairs a residual feature vector with its own keys sorted
// once. euclideanDistance must visit the union of two vectors' keys in a
// fixed order (summing in Go's randomized map iteration order was the
// same-seed nondeterminism bug fixed in determinism_test.go), and
// merge-walking two already-sorted key lists visits exactly that sorted
// union without re-sorting it on every pairwise call - each vector's own
// keys are sorted exactly once however many times it's compared against
// others (profiling showed the former sort-the-union-per-call approach was
// >70% of this CLI's total CPU time; see euclideandistance_test.go for the
// equivalence proof against that reference implementation).
type sortedVector struct {
	v    vector
	keys []string
}

func sortVector(v vector) sortedVector {
	keys := make([]string, 0, len(v))
	for tok := range v {
		keys = append(keys, tok)
	}
	sort.Strings(keys)
	return sortedVector{v: v, keys: keys}
}

// euclideanDistance sums (a[tok]-b[tok])^2 over the union of a and b's keys
// (a Go map defaults an absent key to its zero value, so this is exactly
// equivalent to treating a missing entry as 0 on either side), visiting the
// union in sorted key order by merge-walking a's and b's already-sorted key
// lists.
func euclideanDistance(a, b sortedVector) float64 {
	ak, bk := a.keys, b.keys
	i, j := 0, 0
	sum := 0.0
	for i < len(ak) && j < len(bk) {
		switch {
		case ak[i] < bk[j]:
			d := a.v[ak[i]]
			sum += d * d
			i++
		case ak[i] > bk[j]:
			d := b.v[bk[j]]
			sum += d * d
			j++
		default:
			d := a.v[ak[i]] - b.v[bk[j]]
			sum += d * d
			i++
			j++
		}
	}
	for ; i < len(ak); i++ {
		d := a.v[ak[i]]
		sum += d * d
	}
	for ; j < len(bk); j++ {
		d := b.v[bk[j]]
		sum += d * d
	}
	return math.Sqrt(sum)
}

func residualDistanceMatrix(vecs []sortedVector) [][]float64 {
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

func residualCentroids(vecs []sortedVector, sampleIdx, sampleLabels []int, k int) []vector {
	centroids := make([]vector, k)
	counts := make([]int, k)
	for c := range centroids {
		centroids[c] = vector{}
	}
	for i, si := range sampleIdx {
		c := sampleLabels[i]
		counts[c]++
		for tok, v := range vecs[si].v {
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

func expandResidualLabels(vecs []sortedVector, sampleIdx, sampleLabels []int, k int) []int {
	centroids := residualCentroids(vecs, sampleIdx, sampleLabels, k)
	sortedCentroids := make([]sortedVector, k)
	for c := range centroids {
		sortedCentroids[c] = sortVector(centroids[c])
	}
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
			d := euclideanDistance(v, sortedCentroids[c])
			if d < bd {
				best, bd = c, d
			}
		}
		labels[i] = best
	}
	return labels
}

// residualVectors selects the raw or standardized representation for a set
// of residual windows, pre-sorting each vector's keys once (see
// sortedVector).
func residualVectors(rw []ResidualWindow, standardized bool) []sortedVector {
	out := make([]sortedVector, len(rw))
	for i, w := range rw {
		if standardized {
			out[i] = sortVector(w.Standard)
		} else {
			out[i] = sortVector(w.Residual)
		}
	}
	return out
}

// residualFitPrep holds the (rw, standardized, fitCap)-derived sample and
// its distance matrix: everything fitResidualClustering used to recompute
// on every call, even though none of it depends on k or the clustering
// method. cappedSampleIndices is a deterministic (non-random) even-spacing
// selection and residualDistanceMatrix/residualVectors do no RNG draws, so
// prepareResidualFit's result is exactly invariant across every K value in
// a kMin..kMax sweep for a fixed (scale, replicate) — see
// determinism_test.go for the equivalence proof. globalregime's
// HierarchicalLabels/KMedoids/Diagnostics only ever read their d
// parameter, never write it, so the same sampleD can be safely shared
// across every K's clustering call without one K's fit corrupting
// another's.
type residualFitPrep struct {
	vecs       []sortedVector
	sampleIdx  []int
	sampleVecs []sortedVector
	sampleD    [][]float64
}

// prepareResidualFit computes the part of fitResidualClustering that does
// not depend on k: call this once per (scale, replicate) before the K
// loop, not once per K.
func prepareResidualFit(rw []ResidualWindow, standardized bool, fitCap int) residualFitPrep {
	vecs := residualVectors(rw, standardized)
	sampleIdx := cappedSampleIndices(len(vecs), fitCap)
	sampleVecs := make([]sortedVector, len(sampleIdx))
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
