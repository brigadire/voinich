package higherorderseq

import (
	"math/rand"
	"sort"
)

// cmiWorkspace holds a dense, index-based compilation of one candidate's
// B-neighbor observation set, used by runCMI's within-block right-neighbor
// permutation null (permuteWithinBlocks + cmiBits's jointTable+sortedPairs
// rebuild).
//
// Both the left-neighbor and right-neighbor marginal counts are invariant
// across every permutation: permuteWithinBlocks only ever reshuffles which
// occurrence has which right-neighbor value *within* a block, never adding,
// removing, or moving a value across blocks, so the global multiset of
// right values (and, since left is never touched at all, of left values)
// is identical before and after any draw. Only the (left,right) *pairing*
// changes per replicate, so only the joint table needs rebuilding; the
// marginals are computed once here instead of once per replicate.
type cmiWorkspace struct {
	n int
	V int

	vocab    []string // alphabetically sorted union of all distinct left/right token values
	vocabIdx map[string]int32

	leftIdx, rightIdx []int32 // per-occurrence dense index, invariant
	mLeft, mRight     []int32 // per-token marginal counts over all n occurrences, invariant

	byBlock      [][]int32 // dense block-position index -> occurrence positions, in ascending blockIdx order (matching the reference's map+sort.Ints(blockIdxs) order)
	blockScratch [][]int32 // per dense block index, scratch buffer sized to that block's occurrence count
	permRightIdx []int32   // n, scratch reused across calls

	joint   []int32 // V*V flat table, scratch reused via dirty-list
	touched []int32
}

// newCMIWorkspace compiles the invariant, indexed view of one candidate's
// B-neighbor observations. It is built once per runCMI call and its buffers
// are reused for every replicate.
func newCMIWorkspace(obs []bNeighbor) *cmiWorkspace {
	n := len(obs)
	ws := &cmiWorkspace{n: n}

	vocabSet := map[string]bool{}
	for _, o := range obs {
		vocabSet[o.left] = true
		vocabSet[o.right] = true
	}
	ws.vocab = make([]string, 0, len(vocabSet))
	for t := range vocabSet {
		ws.vocab = append(ws.vocab, t)
	}
	sort.Strings(ws.vocab)
	ws.V = len(ws.vocab)
	ws.vocabIdx = make(map[string]int32, ws.V)
	for i, t := range ws.vocab {
		ws.vocabIdx[t] = int32(i)
	}

	ws.leftIdx = make([]int32, n)
	ws.rightIdx = make([]int32, n)
	ws.mLeft = make([]int32, ws.V)
	ws.mRight = make([]int32, ws.V)
	for i, o := range obs {
		li, ri := ws.vocabIdx[o.left], ws.vocabIdx[o.right]
		ws.leftIdx[i], ws.rightIdx[i] = li, ri
		ws.mLeft[li]++
		ws.mRight[ri]++
	}

	byBlockTmp := map[int][]int32{}
	for i, o := range obs {
		byBlockTmp[o.blockIdx] = append(byBlockTmp[o.blockIdx], int32(i))
	}
	blockIdxs := make([]int, 0, len(byBlockTmp))
	for bi := range byBlockTmp {
		blockIdxs = append(blockIdxs, bi)
	}
	sort.Ints(blockIdxs)
	ws.byBlock = make([][]int32, len(blockIdxs))
	ws.blockScratch = make([][]int32, len(blockIdxs))
	for i, bi := range blockIdxs {
		ws.byBlock[i] = byBlockTmp[bi]
		ws.blockScratch[i] = make([]int32, len(byBlockTmp[bi]))
	}

	ws.joint = make([]int32, ws.V*ws.V)
	ws.permRightIdx = make([]int32, n)
	return ws
}

// permute draws one within-block right-neighbor permutation into
// ws.permRightIdx and returns it. Every block is visited in the same fixed
// (ascending blockIdx) order every call, and each block's r.Shuffle is
// called with the same length every time (block membership never
// changes), so this draws exactly the same sequence of RNG values, in the
// same order, as the reference map+sort-based permuteWithinBlocks.
func (ws *cmiWorkspace) permute(r *rand.Rand) []int32 {
	copy(ws.permRightIdx, ws.rightIdx)
	for bi, idxs := range ws.byBlock {
		scratch := ws.blockScratch[bi]
		for j, oi := range idxs {
			scratch[j] = ws.permRightIdx[oi]
		}
		r.Shuffle(len(scratch), func(a, c int) { scratch[a], scratch[c] = scratch[c], scratch[a] })
		for j, oi := range idxs {
			ws.permRightIdx[oi] = scratch[j]
		}
	}
	return ws.permRightIdx
}

// cmiFor computes cmiBits' plug-in CMI estimate from ws.leftIdx (invariant)
// paired against rightIdx (either ws.rightIdx for the observed pass, or a
// permute() result), rebuilding ws.joint first. Cells are visited in
// (leftIdx ascending, rightIdx ascending) order - which, since vocab is
// alphabetically sorted, is exactly the reference's sorted-(left,right)-pair
// accumulation order - so the summation is bit-identical to sortedPairs'
// map+sort.Slice-based accumulation.
func (ws *cmiWorkspace) cmiFor(rightIdx []int32) float64 {
	if ws.n == 0 {
		return 0
	}
	ws.touched = ws.touched[:0]
	for i := 0; i < ws.n; i++ {
		cell := int(ws.leftIdx[i])*ws.V + int(rightIdx[i])
		if ws.joint[cell] == 0 {
			ws.touched = append(ws.touched, int32(cell))
		}
		ws.joint[cell]++
	}
	nf := float64(ws.n)
	cmi := 0.0
	for t := 0; t < ws.V; t++ {
		px := float64(ws.mLeft[t]) / nf
		if px <= 0 {
			continue
		}
		row := t * ws.V
		for c := 0; c < ws.V; c++ {
			cnt := ws.joint[row+c]
			if cnt == 0 {
				continue
			}
			py := float64(ws.mRight[c]) / nf
			if py <= 0 {
				continue
			}
			pxy := float64(cnt) / nf
			cmi += pxy * log2Ratio(pxy, px*py)
		}
	}
	for _, cell := range ws.touched {
		ws.joint[cell] = 0
	}
	return cmi
}

// observedContribution mirrors runCMI's frozen (A,C)-cell contribution
// computation, from the ORIGINAL (unpermuted) pairing only: a direct O(n)
// scan for how often (a,c) co-occurred, rather than reading it off a
// transient joint table, since a or c may not even be in vocab.
func (ws *cmiWorkspace) observedContribution(a, c string) float64 {
	if ws.n == 0 {
		return 0
	}
	ai, aok := ws.vocabIdx[a]
	ci, cok := ws.vocabIdx[c]
	if !aok || !cok {
		return 0
	}
	jn := 0
	for i := 0; i < ws.n; i++ {
		if ws.leftIdx[i] == ai && ws.rightIdx[i] == ci {
			jn++
		}
	}
	nf := float64(ws.n)
	pxy := float64(jn) / nf
	px := float64(ws.mLeft[ai]) / nf
	py := float64(ws.mRight[ci]) / nf
	if pxy > 0 && px > 0 && py > 0 {
		return pxy * log2Ratio(pxy, px*py)
	}
	return 0
}
