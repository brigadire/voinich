package positionalcontinuation

import (
	"math"
	"math/rand"
	"sort"
)

// positionalWorkspace holds a dense, index-based compilation of one
// (xs, blockIDs, labels) occurrence set, used by runPositionalTests's
// within-block label-permutation null (permuteLabelsWithinBlocks +
// mutualInformationBits + the per-category countMap loop).
//
// Every quantity that does not depend on which permutation is being drawn -
// the continuation-token vocabulary and its per-token counts (mx, since xs
// itself never changes across replicates), block membership groupings, and
// the alphabetical vocab/category orderings needed to reproduce the
// reference implementation's map+sort.Strings accumulation order bit for
// bit - is computed once here instead of once per replicate. Token and
// label-category identity are reduced to dense indices so the replicate
// loop never allocates or hashes a string map; strings reappear only when
// building the final report rows.
type positionalWorkspace struct {
	n int
	C int // number of distinct label values actually observed

	vocab    []string // distinct xs tokens, alphabetically sorted
	vocabIdx map[string]int32
	xsIdx    []int32 // per-occurrence dense token index
	mx       []int32 // per-token count over all n occurrences, invariant
	fcIdx    int32   // vocab index of FrozenChey, or -1 if never observed

	catNames []string // distinct label values, alphabetically sorted (for the joint table's dense accumulation order)
	catIdx   map[string]int32

	labelIdx []int32   // per-occurrence dense category index of the ORIGINAL (unpermuted) labels, invariant
	byBlock  [][]int32 // dense block index -> occurrence positions; block indices assigned in sorted block-ID order (matching the reference's map+sort.Strings(keys) order)

	joint        []int32 // V*C flat table, scratch, reused across calls via dirty-list
	touched      []int32
	my           []int32   // C, scratch reused across calls (small, fully cleared)
	permLabelIdx []int32   // n, scratch reused across calls
	blockScratch [][]int32 // per dense block index, scratch buffer sized to that block's occurrence count
}

// newPositionalWorkspace compiles the invariant, indexed view of one
// (xs, blockIDs, labels) triple. It is built once per runPositionalTests
// call and its buffers are reused for every replicate.
func newPositionalWorkspace(xs, blockIDs, labels []string) *positionalWorkspace {
	n := len(xs)
	ws := &positionalWorkspace{n: n}

	vocabSet := map[string]bool{}
	for _, x := range xs {
		vocabSet[x] = true
	}
	ws.vocab = make([]string, 0, len(vocabSet))
	for t := range vocabSet {
		ws.vocab = append(ws.vocab, t)
	}
	sort.Strings(ws.vocab)
	ws.vocabIdx = make(map[string]int32, len(ws.vocab))
	for i, t := range ws.vocab {
		ws.vocabIdx[t] = int32(i)
	}
	ws.fcIdx = -1
	if i, ok := ws.vocabIdx[FrozenChey]; ok {
		ws.fcIdx = i
	}

	ws.xsIdx = make([]int32, n)
	ws.mx = make([]int32, len(ws.vocab))
	for i, x := range xs {
		idx := ws.vocabIdx[x]
		ws.xsIdx[i] = idx
		ws.mx[idx]++
	}

	catSet := map[string]bool{}
	for _, l := range labels {
		catSet[l] = true
	}
	ws.catNames = make([]string, 0, len(catSet))
	for c := range catSet {
		ws.catNames = append(ws.catNames, c)
	}
	sort.Strings(ws.catNames)
	ws.C = len(ws.catNames)
	ws.catIdx = make(map[string]int32, ws.C)
	for i, c := range ws.catNames {
		ws.catIdx[c] = int32(i)
	}

	ws.labelIdx = make([]int32, n)
	for i, l := range labels {
		ws.labelIdx[i] = ws.catIdx[l]
	}

	byBlockTmp := map[string][]int32{}
	for i, b := range blockIDs {
		byBlockTmp[b] = append(byBlockTmp[b], int32(i))
	}
	blockKeys := make([]string, 0, len(byBlockTmp))
	for b := range byBlockTmp {
		blockKeys = append(blockKeys, b)
	}
	sort.Strings(blockKeys)
	ws.byBlock = make([][]int32, len(blockKeys))
	ws.blockScratch = make([][]int32, len(blockKeys))
	for i, b := range blockKeys {
		ws.byBlock[i] = byBlockTmp[b]
		ws.blockScratch[i] = make([]int32, len(byBlockTmp[b]))
	}

	ws.joint = make([]int32, len(ws.vocab)*ws.C)
	ws.my = make([]int32, ws.C)
	ws.permLabelIdx = make([]int32, n)
	return ws
}

// resultStats mirrors positionalTestResult's per-call statistics computed
// from one (permuted or observed) label assignment: the global I(X;label)
// and, per category (in the caller's given, not alphabetical, order), the
// occurrence count, entropy, chey count/probability and enrichment.
type resultStats struct {
	mi  float64
	cat []catStat
}

type catStat struct {
	n, cheyN, unique int
	h                float64
	cheyP            float64
}

// statsFor computes resultStats from labelIdx (either ws.labelIdx for the
// observed pass, or ws.permLabelIdx after a permutation), for the given
// categories in their caller-provided order. It rebuilds ws.joint/ws.my
// from labelIdx first.
func (ws *positionalWorkspace) statsFor(labelIdx []int32, categories []string) resultStats {
	ws.touched = ws.touched[:0]
	for c := range ws.my {
		ws.my[c] = 0
	}
	for i := 0; i < ws.n; i++ {
		t, c := ws.xsIdx[i], labelIdx[i]
		cell := int(t)*ws.C + int(c)
		if ws.joint[cell] == 0 {
			ws.touched = append(ws.touched, int32(cell))
		}
		ws.joint[cell]++
		ws.my[c]++
	}

	nf := float64(ws.n)
	mi := 0.0
	for t := 0; t < len(ws.vocab); t++ {
		px := float64(ws.mx[t]) / nf
		if px <= 0 {
			continue
		}
		row := t * ws.C
		for c := 0; c < ws.C; c++ {
			cnt := ws.joint[row+c]
			if cnt == 0 {
				continue
			}
			py := float64(ws.my[c]) / nf
			if py <= 0 {
				continue
			}
			pxy := float64(cnt) / nf
			mi += pxy * math.Log2(pxy/(px*py))
		}
	}

	cats := make([]catStat, len(categories))
	for ci, cat := range categories {
		idx, ok := ws.catIdx[cat]
		if !ok {
			continue
		}
		catN := int(ws.my[idx])
		h, unique := 0.0, 0
		for t := 0; t < len(ws.vocab); t++ {
			cnt := ws.joint[t*ws.C+int(idx)]
			if cnt == 0 {
				continue
			}
			unique++
			p := float64(cnt) / float64(catN)
			h -= p * math.Log2(p)
		}
		cheyN := 0
		if ws.fcIdx >= 0 {
			cheyN = int(ws.joint[int(ws.fcIdx)*ws.C+int(idx)])
		}
		cheyP := 0.0
		if catN > 0 {
			cheyP = float64(cheyN) / float64(catN)
		}
		cats[ci] = catStat{n: catN, cheyN: cheyN, unique: unique, h: h, cheyP: cheyP}
	}

	for _, cell := range ws.touched {
		ws.joint[cell] = 0
	}
	return resultStats{mi: mi, cat: cats}
}

// hGlobalAndCheyP computes the global (unconditioned) entropy of xs and the
// global P(chey), both invariant across replicates since they depend only
// on mx.
func (ws *positionalWorkspace) hGlobalAndCheyP() (hGlobal, globalCheyP float64) {
	n := float64(ws.n)
	for t := range ws.mx {
		if ws.mx[t] == 0 {
			continue
		}
		p := float64(ws.mx[t]) / n
		hGlobal -= p * math.Log2(p)
	}
	if ws.fcIdx >= 0 && ws.n > 0 {
		globalCheyP = float64(ws.mx[ws.fcIdx]) / n
	}
	return hGlobal, globalCheyP
}

// permute draws one within-block label permutation into ws.permLabelIdx and
// returns it. Every block is visited in the same fixed (sorted-block-ID)
// order every call, and each block's r.Shuffle is called with the same
// length every time (block membership never changes), so this draws
// exactly the same sequence of RNG values, in the same order, as the
// reference map+sort-based permuteLabelsWithinBlocks.
func (ws *positionalWorkspace) permute(r *rand.Rand) []int32 {
	copy(ws.permLabelIdx, ws.labelIdx)
	for bi, idxs := range ws.byBlock {
		scratch := ws.blockScratch[bi]
		for j, oi := range idxs {
			scratch[j] = ws.permLabelIdx[oi]
		}
		r.Shuffle(len(scratch), func(a, c int) { scratch[a], scratch[c] = scratch[c], scratch[a] })
		for j, oi := range idxs {
			ws.permLabelIdx[oi] = scratch[j]
		}
	}
	return ws.permLabelIdx
}

// stratifiedWorkspace holds a dense, index-based compilation of one
// occurrence set used by runStratifiedPredecessorTest's within-stratum
// isS-permutation null: stratum (block|position-category) membership never
// changes across replicates, so it is grouped once here instead of once
// per replicate via permuteIsSWithinStrata's map+sort rebuild.
type stratifiedWorkspace struct {
	n           int
	baselineIsS []bool
	isChey      []bool
	byStratum   [][]int32 // dense stratum index -> occurrence positions, in sorted-stratum-string order (matching the reference's map+sort.Strings(keys) order)
	scratch     [][]bool  // per stratum, reused shuffle buffer sized to that stratum's occurrence count
	permIsS     []bool    // n, reused output buffer
}

func newStratifiedWorkspace(obs []stratifiedObs) *stratifiedWorkspace {
	n := len(obs)
	ws := &stratifiedWorkspace{n: n, baselineIsS: make([]bool, n), isChey: make([]bool, n), permIsS: make([]bool, n)}
	byStratumTmp := map[string][]int32{}
	for i, o := range obs {
		ws.baselineIsS[i] = o.isS
		ws.isChey[i] = o.isChey
		byStratumTmp[o.stratum] = append(byStratumTmp[o.stratum], int32(i))
	}
	keys := make([]string, 0, len(byStratumTmp))
	for k := range byStratumTmp {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	ws.byStratum = make([][]int32, len(keys))
	ws.scratch = make([][]bool, len(keys))
	for i, k := range keys {
		ws.byStratum[i] = byStratumTmp[k]
		ws.scratch[i] = make([]bool, len(byStratumTmp[k]))
	}
	return ws
}

// observedStatistic returns the pooled chey-among-predecessor-is-s count
// for the original (unpermuted) isS assignment.
func (ws *stratifiedWorkspace) observedStatistic() float64 {
	n := 0.0
	for i, v := range ws.baselineIsS {
		if v && ws.isChey[i] {
			n++
		}
	}
	return n
}

// permuteAndStatistic draws one within-stratum isS permutation and returns
// the pooled chey-among-predecessor-is-s count for it. As in
// positionalWorkspace.permute, every stratum is visited in the same fixed
// (sorted-stratum-string) order every call with the same length every time,
// reproducing the reference implementation's RNG draw sequence exactly.
func (ws *stratifiedWorkspace) permuteAndStatistic(r *rand.Rand) float64 {
	copy(ws.permIsS, ws.baselineIsS)
	for si, idxs := range ws.byStratum {
		scratch := ws.scratch[si]
		for j, oi := range idxs {
			scratch[j] = ws.permIsS[oi]
		}
		r.Shuffle(len(scratch), func(a, c int) { scratch[a], scratch[c] = scratch[c], scratch[a] })
		for j, oi := range idxs {
			ws.permIsS[oi] = scratch[j]
		}
	}
	n := 0.0
	for i, v := range ws.permIsS {
		if v && ws.isChey[i] {
			n++
		}
	}
	return n
}
