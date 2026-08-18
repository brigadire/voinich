package transitionnetwork

import (
	"math"
	"math/rand"
	"sort"
)

// PermWorkspace holds an indexed, permutation-invariant compilation of the
// block data used by the within-block destination permutation null
// (permutedStatistics), plus scratch buffers reused across every replicate.
//
// Every quantity that does not depend on which permutation is being drawn
// (outgoing/incoming opportunity counts, baseline probabilities, baseline
// entropy, and which blocks ever qualify for a given token) is computed
// once here instead of once per replicate. Token identity is reduced to its
// position in analysis.Vocab (a dense "vocab index" in [0, V)) so the
// replicate loop never hashes a string: it only indexes []float64/[]int32
// slices. Strings reappear only when converting the final per-replicate
// results back into the EdgeKey/token-keyed maps the rest of the pipeline
// expects.
//
// The dense V*V transition-count matrix (permCounts) is sized for the
// eligible vocabulary only (bounded by -min-token-count), which in practice
// is in the hundreds; it is allocated once and cleared between blocks via a
// dirty list rather than reallocated or fully zeroed.
type PermWorkspace struct {
	vocab    []string
	V        int
	minBlock int
	blocks   []permBlock

	edgeSrc, edgeTgt []int32 // parallel to analysis.Edges, by vocab index

	permCounts []int32 // V*V flat matrix (row = source, col = target), shared scratch across blocks
	touched    []int32 // dirty list of permCounts indices set by the current block

	edgeVals [][]float64 // per analysis.Edges index, accumulator across blocks, reset each replicate

	outBlocksFor [][]int // per vocab index, block indices with Opp>=minBlock (invariant)
	inBlocksFor  [][]int // per vocab index, block indices with InOpp>=minBlock (invariant)

	outVectors [][][]float64 // [vocabIdx][occurrence][V], preallocated only for tokens with >=3 qualifying blocks
	inVectors  [][][]float64
	outEntropy [][]float64 // [vocabIdx][occurrence]
	inEntropy  [][]float64
	outCursor  []int32 // per vocab index, next write position within outVectors[idx] for the current replicate
	inCursor   []int32

	probScratch          []float64 // len V, reused for the conditional distribution before normalizeProb/entropy
	sumScratch           []float64 // len V, reused leave-one-out sum in lobo()
	refScratch           []float64 // len V, reused leave-one-out mean in lobo()
	corrScratch          []float64 // reused per-token correlation accumulator in lobo()
	signScratch          []float64 // reused per-token sign-agreement accumulator in lobo()
	entropyMedianScratch []float64 // reused per-token entropy-effect sort buffer in lobo()
}

type permBlock struct {
	Opp, InOpp     []int32   // len V
	BasePb         []float64 // len V; BasePb[i] = (Counts[vocab[i]]+alpha)/(n+alpha*V), invariant
	BaseProb       []float64 // len V; normalized copy of BasePb, invariant
	BaseEntropy    float64   // entropy(BaseProb), invariant
	OppSrc, OppDst []int32   // len = opportunities in block, in original (unpermuted) order; OppDst is -1 for an ineligible destination
	scratchDest    []int32   // len(OppDst), reused shuffle buffer
}

// newPermWorkspace compiles the invariant, indexed view of a and its
// per-block data used by the permutation null. It is built once per
// RunAndWrite call and its buffers are reused for every replicate.
func newPermWorkspace(a *analysis, minBlock int) *PermWorkspace {
	V := len(a.Vocab)
	idx := make(map[string]int, V)
	for i, t := range a.Vocab {
		idx[t] = i
	}
	ws := &PermWorkspace{vocab: a.Vocab, V: V, minBlock: minBlock}
	ws.blocks = make([]permBlock, len(a.Data))
	ws.outBlocksFor = make([][]int, V)
	ws.inBlocksFor = make([][]int, V)
	for bi, d := range a.Data {
		n := len(d.Block.Tokens)
		pb := make([]float64, V)
		for i, t := range a.Vocab {
			pb[i] = (float64(d.Counts[t]) + alpha) / (float64(n) + alpha*float64(V))
		}
		baseProb := append([]float64(nil), pb...)
		normalizeProb(baseProb)
		opp := make([]int32, V)
		inOpp := make([]int32, V)
		for i, t := range a.Vocab {
			opp[i] = int32(d.Opp[t])
			inOpp[i] = int32(incomingOpp(d, t))
		}
		var oppSrc, oppDst []int32
		for i := 0; i+1 < len(d.Block.Tokens); i++ {
			s, c := d.Block.Tokens[i].Text, d.Block.Tokens[i+1].Text
			if d.Eligible[s] {
				oppSrc = append(oppSrc, int32(idx[s]))
				if d.Eligible[c] {
					oppDst = append(oppDst, int32(idx[c]))
				} else {
					oppDst = append(oppDst, -1)
				}
			}
		}
		ws.blocks[bi] = permBlock{
			Opp: opp, InOpp: inOpp, BasePb: pb, BaseProb: baseProb,
			BaseEntropy: entropy(baseProb),
			OppSrc:      oppSrc, OppDst: oppDst,
			scratchDest: make([]int32, len(oppDst)),
		}
		for i := 0; i < V; i++ {
			if opp[i] >= int32(minBlock) {
				ws.outBlocksFor[i] = append(ws.outBlocksFor[i], bi)
			}
			if inOpp[i] >= int32(minBlock) {
				ws.inBlocksFor[i] = append(ws.inBlocksFor[i], bi)
			}
		}
	}
	ws.edgeSrc = make([]int32, len(a.Edges))
	ws.edgeTgt = make([]int32, len(a.Edges))
	for i, e := range a.Edges {
		ws.edgeSrc[i] = int32(idx[e.Source])
		ws.edgeTgt[i] = int32(idx[e.Target])
	}
	ws.permCounts = make([]int32, V*V)
	ws.edgeVals = make([][]float64, len(a.Edges))
	ws.outVectors = make([][][]float64, V)
	ws.inVectors = make([][][]float64, V)
	ws.outEntropy = make([][]float64, V)
	ws.inEntropy = make([][]float64, V)
	for i := 0; i < V; i++ {
		if len(ws.outBlocksFor[i]) >= 3 {
			ws.outVectors[i] = make([][]float64, len(ws.outBlocksFor[i]))
			for j := range ws.outVectors[i] {
				ws.outVectors[i][j] = make([]float64, V)
			}
			ws.outEntropy[i] = make([]float64, len(ws.outBlocksFor[i]))
		}
		if len(ws.inBlocksFor[i]) >= 3 {
			ws.inVectors[i] = make([][]float64, len(ws.inBlocksFor[i]))
			for j := range ws.inVectors[i] {
				ws.inVectors[i][j] = make([]float64, V)
			}
			ws.inEntropy[i] = make([]float64, len(ws.inBlocksFor[i]))
		}
	}
	ws.outCursor = make([]int32, V)
	ws.inCursor = make([]int32, V)
	ws.probScratch = make([]float64, V)
	ws.sumScratch = make([]float64, V)
	ws.refScratch = make([]float64, V)
	return ws
}

// run draws replicate rep of the within-block destination permutation null
// and returns the same shapes permutedStatistics does: median permuted
// log2-enrichment per edge, and (when computeProfiles is true) the LOBO
// null statistics per token for the outgoing and incoming direction. When
// computeProfiles is false (used during pre-specified refinement, where
// only edge exceedance is consulted) the profile-vector and entropy work is
// skipped entirely, since those results would otherwise be discarded.
//
// Every arithmetic expression below mirrors, operation for operation, the
// corresponding map-based expression in permuteBlockEdges/effect/
// collectEffectVectors/collectEntropyEffects/averageVectors/
// profileLOBOStats, so results are bit-identical wherever floating-point
// operation order is unchanged, and within documented roundoff (<1e-12)
// where a leave-one-out mean is derived from a cached total instead of a
// fresh per-exclusion sum (see lobo below).
func (ws *PermWorkspace) run(seed int64, rep int, computeProfiles bool) (map[EdgeKey]float64, map[string]profileNullStat, map[string]profileNullStat) {
	rng := rand.New(rand.NewSource(seed + int64(rep)*0x1f123bb5))
	V := ws.V
	for i := range ws.edgeVals {
		ws.edgeVals[i] = ws.edgeVals[i][:0]
	}
	if computeProfiles {
		for i := range ws.outCursor {
			ws.outCursor[i] = 0
			ws.inCursor[i] = 0
		}
	}
	for bi := range ws.blocks {
		blk := &ws.blocks[bi]
		copy(blk.scratchDest, blk.OppDst)
		dest := blk.scratchDest
		rng.Shuffle(len(dest), func(i, j int) { dest[i], dest[j] = dest[j], dest[i] })
		ws.touched = ws.touched[:0]
		for k, s := range blk.OppSrc {
			t := dest[k]
			if t < 0 {
				continue
			}
			cell := int(s)*V + int(t)
			if ws.permCounts[cell] == 0 {
				ws.touched = append(ws.touched, int32(cell))
			}
			ws.permCounts[cell]++
		}
		for ei, s := range ws.edgeSrc {
			if blk.Opp[s] < int32(ws.minBlock) {
				continue
			}
			t := ws.edgeTgt[ei]
			cnt := ws.permCounts[int(s)*V+int(t)]
			pc := (float64(cnt) + alpha) / (float64(blk.Opp[s]) + alpha*float64(V))
			en := pc / blk.BasePb[t]
			ws.edgeVals[ei] = append(ws.edgeVals[ei], math.Log2(en))
		}
		if computeProfiles {
			for s := 0; s < V; s++ {
				if blk.Opp[s] < int32(ws.minBlock) || ws.outVectors[s] == nil {
					continue
				}
				pos := ws.outCursor[s]
				ws.outCursor[s]++
				row := ws.outVectors[s][pos]
				rowBase := s * V
				oppS := blk.Opp[s]
				for t := 0; t < V; t++ {
					cnt := ws.permCounts[rowBase+t]
					pc := (float64(cnt) + alpha) / (float64(oppS) + alpha*float64(V))
					row[t] = math.Log2(pc / blk.BasePb[t])
					ws.probScratch[t] = pc
				}
				normalizeProb(ws.probScratch)
				ws.outEntropy[s][pos] = entropy(ws.probScratch) - blk.BaseEntropy
			}
			for t := 0; t < V; t++ {
				if blk.InOpp[t] < int32(ws.minBlock) || ws.inVectors[t] == nil {
					continue
				}
				pos := ws.inCursor[t]
				ws.inCursor[t]++
				row := ws.inVectors[t][pos]
				oppT := blk.InOpp[t]
				for s := 0; s < V; s++ {
					cnt := ws.permCounts[s*V+t]
					pc := (float64(cnt) + alpha) / (float64(oppT) + alpha*float64(V))
					row[s] = math.Log2(pc / blk.BasePb[s])
					ws.probScratch[s] = pc
				}
				normalizeProb(ws.probScratch)
				ws.inEntropy[t][pos] = entropy(ws.probScratch) - blk.BaseEntropy
			}
		}
		for _, cell := range ws.touched {
			ws.permCounts[cell] = 0
		}
	}
	effectStats := make(map[EdgeKey]float64, len(ws.edgeSrc))
	for ei, vals := range ws.edgeVals {
		if len(vals) == 0 {
			// No block ever has Opp[source]>=minBlock for this edge (an
			// invariant property); the reference map-based implementation
			// never creates a map entry in this case, so this edge must
			// stay absent here too rather than default to a spurious 0.
			continue
		}
		sort.Float64s(vals)
		effectStats[EdgeKey{ws.vocab[ws.edgeSrc[ei]], ws.vocab[ws.edgeTgt[ei]]}] = medianOfSorted(vals)
	}
	if !computeProfiles {
		return effectStats, nil, nil
	}
	return effectStats, ws.lobo(ws.outVectors, ws.outEntropy), ws.lobo(ws.inVectors, ws.inEntropy)
}

// lobo mirrors profileLOBOStats over dense per-token vectors: for every
// token with at least 3 qualifying block occurrences this replicate, it
// computes the leave-one-out correlation and sign agreement of each
// occurrence against the mean of the others, and the median entropy
// effect.
//
// The leave-one-out mean is derived from a single cached sum of all
// occurrences (sum-minus-self)/(n-1) rather than a fresh sum over the other
// n-1 occurrences per exclusion; both compute the same quantity, but
// floating-point addition is not associative, so results can differ from
// the reference (map-based) implementation by up to a few ULPs, far below
// the documented 1e-12 scientific-equivalence tolerance.
func (ws *PermWorkspace) lobo(vectors [][][]float64, entropyVals [][]float64) map[string]profileNullStat {
	out := map[string]profileNullStat{}
	for idx := 0; idx < ws.V; idx++ {
		xs := vectors[idx]
		n := len(xs)
		if n < 3 {
			continue
		}
		sum := ws.sumScratch
		for t := range sum {
			sum[t] = 0
		}
		for _, x := range xs {
			for t, v := range x {
				sum[t] += v
			}
		}
		ref := ws.refScratch
		nf := float64(n - 1)
		cs := ws.corrScratch[:0]
		ss := ws.signScratch[:0]
		for _, x := range xs {
			for t := range ref {
				ref[t] = (sum[t] - x[t]) / nf
			}
			cs = append(cs, pearson(x, ref))
			ss = append(ss, vectorSignAgreementDense(x, ref))
		}
		ws.corrScratch = cs
		ws.signScratch = ss
		sort.Float64s(cs)
		ent := append(ws.entropyMedianScratch[:0], entropyVals[idx]...)
		ws.entropyMedianScratch = ent
		sort.Float64s(ent)
		out[ws.vocab[idx]] = profileNullStat{Correlation: medianOfSorted(cs), SignAgreement: mean(ss), EntropyEffect: medianOfSorted(ent)}
	}
	return out
}

func vectorSignAgreementDense(x, ref []float64) float64 {
	if len(x) == 0 {
		return 0
	}
	n := 0
	for i, v := range x {
		if (v > 0) == (ref[i] > 0) {
			n++
		}
	}
	return float64(n) / float64(len(x))
}

func medianOfSorted(y []float64) float64 {
	n := len(y)
	if n == 0 {
		return 0
	}
	if n%2 == 1 {
		return y[n/2]
	}
	return (y[n/2-1] + y[n/2]) / 2
}
