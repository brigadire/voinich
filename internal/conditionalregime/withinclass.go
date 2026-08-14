package conditionalregime

import (
	"math"
	"sort"

	"zcore.dev/voinich/internal/globalregime"
)

// classWindow is one within-class window: the same distributional
// representation global-regime-analyze uses, tagged with the physical block
// it came from and its absolute corpus position. Windows are built
// per-block (task19 section 11: a window never crosses a class/metadata
// boundary) and only concatenated afterward for distance-based clustering,
// which does not care about window order.
type classWindow struct {
	globalregime.Window
	BlockIdx         int
	AbsStart, AbsEnd int
}

// buildClassWindows builds every window at one scale for one class, drawing
// tokens only from that class's own physical blocks. tokens is the full
// canonical corpus; it is never modified or reordered.
func buildClassWindows(tokens []string, blocks []Block, windowSize int) []classWindow {
	var out []classWindow
	for bi, b := range blocks {
		if b.Len() < windowSize {
			continue
		}
		ws := globalregime.BuildWindows(tokens[b.Start:b.End], windowSize, 0)
		for _, w := range ws {
			out = append(out, classWindow{Window: w, BlockIdx: bi, AbsStart: b.Start + w.Start, AbsEnd: b.Start + w.End})
		}
	}
	return out
}

func plainWindows(cw []classWindow) []globalregime.Window {
	out := make([]globalregime.Window, len(cw))
	for i, w := range cw {
		out[i] = w.Window
	}
	return out
}

// allowedKMax is the fixed small-class cap: K may never exceed
// floor(numWindows/20) even though kMaxWithin is otherwise available
// (task19 section 14). It returns 0 if no K in [kMin,cap] is usable.
func allowedKMax(numWindows, kMin, kMax int) int {
	capped := numWindows / 20
	if capped > kMax {
		capped = kMax
	}
	if capped < kMin {
		return 0
	}
	return capped
}

func clusterSizeEntropy(sizes []int) float64 {
	total := 0
	for _, s := range sizes {
		total += s
	}
	if total == 0 {
		return 0
	}
	h := 0.0
	for _, s := range sizes {
		if s == 0 {
			continue
		}
		p := float64(s) / float64(total)
		h -= p * math.Log(p)
	}
	return h
}

func smallestClusterFraction(sizes []int) float64 {
	total, smallest := 0, -1
	for _, s := range sizes {
		total += s
		if s > 0 && (smallest < 0 || s < smallest) {
			smallest = s
		}
	}
	if total == 0 || smallest < 0 {
		return 0
	}
	return float64(smallest) / float64(total)
}

func medoidOf(cluster int, labels []int, d [][]float64) int {
	best, bestSum := -1, math.Inf(1)
	for i, li := range labels {
		if li != cluster {
			continue
		}
		s := 0.0
		for j, lj := range labels {
			if lj == cluster {
				s += d[i][j]
			}
		}
		if s < bestSum {
			best, bestSum = i, s
		}
	}
	return best
}

// medoidSeparation is the mean pairwise distance between the K cluster
// medoids: how far apart the discovered regimes' representative windows are.
func medoidSeparation(labels []int, d [][]float64, k int) float64 {
	medoids := make([]int, 0, k)
	for c := 0; c < k; c++ {
		if m := medoidOf(c, labels, d); m >= 0 {
			medoids = append(medoids, m)
		}
	}
	sum, n := 0.0, 0
	for i := 0; i < len(medoids); i++ {
		for j := i + 1; j < len(medoids); j++ {
			sum += d[medoids[i]][medoids[j]]
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return sum / float64(n)
}

// fitClustering fits one (method, K) clustering on a deterministically
// sampled subset (mirroring global-regime-analyze's fitting strategy for
// long sequences), then expands the assignment to every window.
func fitClustering(all []classWindow, method string, k int, seed int64) (fitLabels []int, fullLabels []int, sampleD [][]float64) {
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

// WithinClassRegime is one row of within_class_regimes.tsv: one
// class x window_size x method x K clustering, with the diagnostics required
// by task19 section 16.
type WithinClassRegime struct {
	Scheme              Scheme
	Class               string
	WindowSize          int
	Method              string
	K                   int
	AllowedKMax         int
	Windows             int
	Silhouette          float64
	MedoidSeparation    float64
	WithinDispersion    float64
	BetweenDispersion   float64
	ClusterSizeEntropy  float64
	SmallestClusterFrac float64
	Diagnostic          bool // window_size=1000: secondary diagnostic only (task19 section 10)
	ClusterSizes        []int
	fullLabels          []int
	windowsRef          []classWindow
}

// withinClassSweep clusters one class's windows at one window size, for K in
// [kMin, min(kMaxWithin, floor(numWindows/20))], for both the primary
// (k_medoids) and secondary (hierarchical) methods.
func withinClassSweep(tokens []string, class ClassID, blocks []Block, windowSize, kMin, kMaxWithin int, seed int64) ([]WithinClassRegime, []classWindow) {
	cw := buildClassWindows(tokens, blocks, windowSize)
	if len(cw) < 2*kMin {
		return nil, cw
	}
	kmax := allowedKMax(len(cw), kMin, kMaxWithin)
	var out []WithinClassRegime
	if kmax == 0 {
		return out, cw
	}
	for _, method := range []string{"k_medoids", "hierarchical"} {
		for k := kMin; k <= kmax; k++ {
			fitLabels, fullLabels, sampleD := fitClustering(cw, method, k, seed)
			d := globalregime.Diagnostics(windowSize, method, k, fitLabels, sampleD)
			sep := medoidSeparation(fitLabels, sampleD, k)
			d = globalregime.WithFullAssignments(d, fullLabels)
			out = append(out, WithinClassRegime{
				Scheme: class.Scheme, Class: class.Label(), WindowSize: windowSize, Method: method, K: k,
				AllowedKMax: kmax, Windows: len(cw), Silhouette: d.Silhouette, MedoidSeparation: sep,
				WithinDispersion: d.WithinDispersion, BetweenDispersion: d.BetweenDistance,
				ClusterSizeEntropy: clusterSizeEntropy(d.ClusterSizes), SmallestClusterFrac: smallestClusterFraction(d.ClusterSizes),
				Diagnostic: windowSize >= 1000, ClusterSizes: d.ClusterSizes, fullLabels: fullLabels, windowsRef: cw,
			})
		}
	}
	return out, cw
}

// bestByMethod returns, for each method present, the regime row with the
// highest silhouette (the max-over-K selection used for reporting and for
// the permutation candidate rule). Ties resolve to the smallest K.
func bestByMethod(rows []WithinClassRegime) map[string]WithinClassRegime {
	out := map[string]WithinClassRegime{}
	for _, r := range rows {
		cur, ok := out[r.Method]
		if !ok || r.Silhouette > cur.Silhouette || (r.Silhouette == cur.Silhouette && r.K < cur.K) {
			out[r.Method] = r
		}
	}
	return out
}

// CrossBlockRecurrence is one cluster's recurrence across a class's physical
// blocks (task19 section 18).
type CrossBlockRecurrence struct {
	Scheme                 Scheme
	Class                  string
	WindowSize             int
	Method                 string
	K                      int
	Cluster                int
	BlocksContaining       int
	TotalBlocks            int
	BlockFraction          float64
	EligibleBlocksForScale int
	RecurrenceStrength     float64
	CrossBlockSimilarity   float64 // 1 - mean pairwise JS distance between per-block mean profiles; NaN-safe 0 if <2 blocks
}

func meanProfile(ws []classWindow) globalregime.Profile {
	p := globalregime.Profile{}
	if len(ws) == 0 {
		return p
	}
	for _, w := range ws {
		for tok, v := range w.Distribution() {
			p[tok] += v
		}
	}
	for tok := range p {
		p[tok] /= float64(len(ws))
	}
	return p
}

func crossBlockRecurrence(class ClassID, windowSize int, method string, k int, cw []classWindow, fullLabels []int, blocks []Block) []CrossBlockRecurrence {
	eligibleBlocks := 0
	for _, b := range blocks {
		if b.Len() >= windowSize {
			eligibleBlocks++
		}
	}
	out := make([]CrossBlockRecurrence, 0, k)
	for c := 0; c < k; c++ {
		byBlock := map[int][]classWindow{}
		for i, label := range fullLabels {
			if label == c {
				byBlock[cw[i].BlockIdx] = append(byBlock[cw[i].BlockIdx], cw[i])
			}
		}
		blockIdxs := make([]int, 0, len(byBlock))
		for bi := range byBlock {
			blockIdxs = append(blockIdxs, bi)
		}
		sort.Ints(blockIdxs)
		similarity := 0.0
		if len(blockIdxs) >= 2 {
			profiles := make([]globalregime.Profile, len(blockIdxs))
			for i, bi := range blockIdxs {
				profiles[i] = meanProfile(byBlock[bi])
			}
			sum, n := 0.0, 0
			for i := 0; i < len(profiles); i++ {
				for j := i + 1; j < len(profiles); j++ {
					sum += globalregime.JSDistance(profiles[i], profiles[j])
					n++
				}
			}
			if n > 0 {
				similarity = 1 - sum/float64(n)
			}
		}
		frac := 0.0
		if len(blocks) > 0 {
			frac = float64(len(blockIdxs)) / float64(len(blocks))
		}
		strength := 0.0
		if eligibleBlocks > 0 {
			strength = float64(len(blockIdxs)) / float64(eligibleBlocks)
		}
		out = append(out, CrossBlockRecurrence{
			Scheme: class.Scheme, Class: class.Label(), WindowSize: windowSize, Method: method, K: k, Cluster: c,
			BlocksContaining: len(blockIdxs), TotalBlocks: len(blocks), BlockFraction: frac,
			EligibleBlocksForScale: eligibleBlocks, RecurrenceStrength: strength, CrossBlockSimilarity: similarity,
		})
	}
	return out
}
