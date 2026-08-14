package conditionalregime

import (
	"math"

	"zcore.dev/voinich/internal/globalregime"
)

// foldSplit is one contiguous-block or contiguous-fold train/test split
// (task19 section 17: never a random window split).
type foldSplit struct {
	train, test []Block
}

// contiguousFolds builds leave-one-block-out folds when a class has
// multiple physical blocks, or three contiguous within-block folds when it
// has only one. Both are contiguous/physical-block based, never a random
// window split (task19 sections 17, 40).
func contiguousFolds(blocks []Block) []foldSplit {
	if len(blocks) >= 2 {
		folds := make([]foldSplit, len(blocks))
		for i := range blocks {
			var train []Block
			for j, b := range blocks {
				if j != i {
					train = append(train, b)
				}
			}
			folds[i] = foldSplit{train, []Block{blocks[i]}}
		}
		return folds
	}
	if len(blocks) == 1 {
		b := blocks[0]
		n := 3
		if b.Len() < 2*n {
			return nil
		}
		step := b.Len() / n
		folds := make([]foldSplit, 0, n)
		for i := 0; i < n; i++ {
			s, e := b.Start+i*step, b.Start+(i+1)*step
			if i == n-1 {
				e = b.End
			}
			test := Block{Class: b.Class, Index: b.Index, Start: s, End: e}
			var train []Block
			if s > b.Start {
				train = append(train, Block{Class: b.Class, Index: b.Index, Start: b.Start, End: s})
			}
			if e < b.End {
				train = append(train, Block{Class: b.Class, Index: b.Index, Start: e, End: b.End})
			}
			folds = append(folds, foldSplit{train, []Block{test}})
		}
		return folds
	}
	return nil
}

// heldOutSeparation is the transfer analogue of silhouette that task19
// section 17 asks for: for each held-out window, the normalized gap between
// its distance to the nearest and second-nearest medoid fitted on training
// blocks only. A model that generalizes to held-out material scores high;
// one that only fit training noise scores near zero.
func heldOutSeparation(heldOut []classWindow, trainMedoidProfiles []globalregime.Profile) float64 {
	if len(heldOut) == 0 || len(trainMedoidProfiles) < 2 {
		return 0
	}
	sum := 0.0
	for _, w := range heldOut {
		best, second := math.Inf(1), math.Inf(1)
		for _, p := range trainMedoidProfiles {
			d := globalregime.JSDistance(w.Distribution(), p)
			if d < best {
				second, best = best, d
			} else if d < second {
				second = d
			}
		}
		den := math.Max(best, second)
		if den > 0 && !math.IsInf(second, 1) {
			sum += (second - best) / den
		}
	}
	return sum / float64(len(heldOut))
}

// WithinClassStability is one row of within_class_stability.tsv: the
// cross-validated held-out separation for one class x window_size x method,
// at the K chosen from the full-data fit.
type WithinClassStability struct {
	Class      ClassID
	WindowSize int
	Method     string
	K          int
	Folds      int
	Score      float64
}

// stabilityForClass runs contiguous-fold cross-validation: fit on training
// blocks, transfer-assign held-out windows to the fitted medoids, and
// average heldOutSeparation over folds.
func stabilityForClass(tokens []string, class ClassID, blocks []Block, windowSize, k int, method string, seed int64) WithinClassStability {
	folds := contiguousFolds(blocks)
	var scores []float64
	for _, fold := range folds {
		trainW := buildClassWindows(tokens, fold.train, windowSize)
		testW := buildClassWindows(tokens, fold.test, windowSize)
		if len(trainW) < 2*k || len(testW) == 0 {
			continue
		}
		full := plainWindows(trainW)
		sample, _ := globalregime.ClusteringSample(full)
		d := globalregime.DistanceMatrix(sample)
		var labels []int
		if method == "hierarchical" {
			labels = globalregime.HierarchicalLabels(len(sample), k, d)
		} else {
			labels = globalregime.KMedoids(d, k, seed)
		}
		medoidProfiles := make([]globalregime.Profile, 0, k)
		for c := 0; c < k; c++ {
			if m := medoidOf(c, labels, d); m >= 0 {
				medoidProfiles = append(medoidProfiles, sample[m].Distribution())
			}
		}
		scores = append(scores, heldOutSeparation(testW, medoidProfiles))
	}
	return WithinClassStability{Class: class, WindowSize: windowSize, Method: method, K: k, Folds: len(scores), Score: meanFloat(scores)}
}
