package clustermetadataglobal

import (
	"math"
	"math/rand"
)

// kindPrep holds everything derived once from real metadata for one kind
// (currier/hand): the label encoding, the block structure used to generate
// permutations, and the eligible-window index lists per scope. Eligibility
// is computed once from the real, unpermuted labels and reused unchanged for
// every permutation replicate, because permuteKnownBlocks never moves the
// known/unknown mask.
type kindPrep struct {
	numLabels  int
	valueIndex map[string]int8
	blocks     []labelBlock
	realCodes  []int8
	realCum    []int32
	scratch    []int32
	eligible   map[string]map[int][]int // scope name -> window_size -> eligible window indices
}

func prepareKind(fs *frozenSpace, labels []string) *kindPrep {
	codes, _, valueIndex := encodeLabels(labels)
	numLabels := len(valueIndex)
	cum := make([]int32, (len(codes)+1)*numLabels)
	buildCumulativeInto(codes, numLabels, cum)
	eligible := map[string]map[int][]int{}
	for _, sc := range Scopes {
		byWs := map[int][]int{}
		for _, ws := range WindowSizes {
			ranges := fs.Windows[ws]
			var idxs []int
			for i, r := range ranges {
				_, purity, known := windowMajority(cum, numLabels, r.Start, r.End)
				if !known {
					continue
				}
				if sc.Threshold > 0 && purity < sc.Threshold {
					continue
				}
				idxs = append(idxs, i)
			}
			byWs[ws] = idxs
		}
		eligible[sc.Name] = byWs
	}
	return &kindPrep{
		numLabels:  numLabels,
		valueIndex: valueIndex,
		blocks:     blocksOf(labels),
		realCodes:  codes,
		realCum:    cum,
		scratch:    make([]int32, (len(labels)+1)*numLabels),
		eligible:   eligible,
	}
}

func make3D(a, b, c int) [][][]float64 {
	out := make([][][]float64, a)
	for i := range out {
		out[i] = make([][]float64, b)
		for j := range out[i] {
			out[i][j] = make([]float64, c)
		}
	}
	return out
}

type grid struct{ NMI, ARI [][][]float64 }

// computeGrid evaluates every frozen (method, window_size, K) combination for
// one permutation replicate of one metadata kind and scope, sharing the same
// per-window majority-label codes across the whole method x K sweep at each
// window size.
func computeGrid(fs *frozenSpace, numLabels int, majority map[int][]int8, eligibleByWs map[int][]int) grid {
	nk := KMax - KMin + 1
	g := grid{NMI: make3D(len(Methods), len(WindowSizes), nk), ARI: make3D(len(Methods), len(WindowSizes), nk)}
	for mi, m := range Methods {
		for si, ws := range WindowSizes {
			labelCodes := majority[ws]
			eligible := eligibleByWs[ws]
			for ki, k := range ksRange() {
				combo := fs.Combos[comboKey{ws, m, k}]
				nmi, ari := fastMetrics(labelCodes, combo.Cluster, eligible, numLabels, combo.NumClusters)
				g.NMI[mi][si][ki] = nmi
				g.ARI[mi][si][ki] = ari
			}
		}
	}
	return g
}

// derived summarizes one metric's grid into the three primary statistics
// (A: per-method max over window x K, B: global max over window x method x
// K, C: per-method mean of scale-specific max-over-K) plus the secondary
// scale-minimum diagnostic.
type derived struct {
	PerMethodMax     []float64
	PerMethodArgSize []int
	PerMethodArgK    []int
	PersistenceVec   [][]float64
	PersistenceMean  []float64
	PersistenceMin   []float64
	GlobalMax        float64
	GlobalMethod     string
	GlobalSize       int
	GlobalK          int
}

func derive(values [][][]float64) derived {
	nm := len(Methods)
	d := derived{
		PerMethodMax: make([]float64, nm), PerMethodArgSize: make([]int, nm), PerMethodArgK: make([]int, nm),
		PersistenceVec: make([][]float64, nm), PersistenceMean: make([]float64, nm), PersistenceMin: make([]float64, nm),
	}
	globalMax := math.Inf(-1)
	gm, gs, gk := 0, 0, 0
	for mi := range Methods {
		best := math.Inf(-1)
		bestSi, bestKi := 0, 0
		pv := make([]float64, len(WindowSizes))
		for si := range WindowSizes {
			rowMax := math.Inf(-1)
			for ki, v := range values[mi][si] {
				if v > rowMax {
					rowMax = v
				}
				if v > best {
					best, bestSi, bestKi = v, si, ki
				}
			}
			pv[si] = rowMax
		}
		d.PerMethodMax[mi] = best
		d.PerMethodArgSize[mi] = bestSi
		d.PerMethodArgK[mi] = bestKi
		d.PersistenceVec[mi] = pv
		d.PersistenceMean[mi] = meanFloat(pv)
		d.PersistenceMin[mi] = minFloatSlice(pv)
		if best > globalMax {
			globalMax, gm, gs, gk = best, mi, bestSi, bestKi
		}
	}
	d.GlobalMax = globalMax
	d.GlobalMethod = Methods[gm]
	d.GlobalSize = WindowSizes[gs]
	d.GlobalK = KMin + gk
	return d
}

func seriesKey(kind, scope, metric, methodScope string) string {
	return kind + "|" + scope + "|" + metric + "|" + methodScope
}

func newSeriesSet(permutations int) map[string]*StatSeries {
	out := map[string]*StatSeries{}
	for _, kind := range Kinds {
		for _, sc := range Scopes {
			for _, metric := range []string{"NMI", "ARI"} {
				add := func(methodScope string) {
					k := seriesKey(kind, sc.Name, metric, methodScope)
					out[k] = &StatSeries{Metadata: kind, Scope: sc.Name, Metric: metric, MethodScope: methodScope, Null: make([]float64, permutations)}
				}
				for _, m := range Methods {
					add(m)
					add(m + "/persistence_mean")
					add(m + "/persistence_min")
				}
				add("global")
			}
		}
	}
	return out
}

// RunSearchSpace evaluates the entire frozen window_size x method x K search
// space, for every metadata kind and scope, on the real (observed) metadata
// and on `permutations` block-aware null realizations that reuse one
// permuted metadata sequence across the complete search space per replicate.
// onProgress, if non-nil, is called after every completed null replicate.
func RunSearchSpace(fs *frozenSpace, currier, hand []string, permutations int, seed int64, onProgress func(done, total int)) (map[string]*StatSeries, map[string][]float64) {
	preps := map[string]*kindPrep{
		"currier": prepareKind(fs, currier),
		"hand":    prepareKind(fs, hand),
	}
	rngs := map[string]*rand.Rand{}
	for i, kind := range Kinds {
		rngs[kind] = rand.New(rand.NewSource(seed + int64(i)))
	}
	series := newSeriesSet(permutations)
	byWindowVector := map[string][]float64{}
	store := func(r int, kind, scope, metric string, d derived) {
		set := func(methodScope string, observed float64, ow, ok int, om string) {
			s := series[seriesKey(kind, scope, metric, methodScope)]
			if r == 0 {
				s.Observed, s.ObservedWindow, s.ObservedK, s.ObservedMethod = observed, ow, ok, om
			} else {
				s.Null[r-1] = observed
			}
		}
		for mi, m := range Methods {
			set(m, d.PerMethodMax[mi], WindowSizes[d.PerMethodArgSize[mi]], KMin+d.PerMethodArgK[mi], m)
			set(m+"/persistence_mean", d.PersistenceMean[mi], 0, 0, m)
			set(m+"/persistence_min", d.PersistenceMin[mi], 0, 0, m)
			if r == 0 {
				byWindowVector[kind+"|"+scope+"|"+metric+"|"+m] = d.PersistenceVec[mi]
			}
		}
		set("global", d.GlobalMax, d.GlobalSize, d.GlobalK, d.GlobalMethod)
	}
	for r := 0; r <= permutations; r++ {
		for _, kind := range Kinds {
			prep := preps[kind]
			var codes []int8
			var cum []int32
			if r == 0 {
				codes, cum = prep.realCodes, prep.realCum
			} else {
				permuted := permuteKnownBlocks(prep.blocks, rngs[kind])
				codes = codesFromLabels(permuted, prep.valueIndex)
				cum = prep.scratch
				buildCumulativeInto(codes, prep.numLabels, cum)
			}
			majority := map[int][]int8{}
			for _, ws := range WindowSizes {
				majority[ws] = majorityPerWindow(cum, prep.numLabels, fs.Windows[ws])
			}
			for _, sc := range Scopes {
				g := computeGrid(fs, prep.numLabels, majority, prep.eligible[sc.Name])
				store(r, kind, sc.Name, "NMI", derive(g.NMI))
				store(r, kind, sc.Name, "ARI", derive(g.ARI))
			}
		}
		if r > 0 && onProgress != nil {
			onProgress(r, permutations)
		}
	}
	return series, byWindowVector
}
