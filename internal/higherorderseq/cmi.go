package higherorderseq

import (
	"math"
	"math/rand"
	"sort"
)

// bNeighbor is one occurrence of the candidate's central token B that has
// both a left and a right neighbor inside the same physical block.
type bNeighbor struct {
	blockIdx    int
	left, right string
}

// collectBNeighbors implements task22 Part D section 17: for the candidate's
// central token B, the left/right neighbor pair at every occurrence that has
// both neighbors within one physical block (an occurrence at a block edge
// has no defined neighbor on that side and is excluded, which is also what
// keeps block boundaries from ever being crossed).
func collectBNeighbors(centerToken string, blocks []Block) []bNeighbor {
	var out []bNeighbor
	for bi, blk := range blocks {
		n := len(blk.Tokens)
		for k := 1; k+1 < n; k++ {
			if blk.Tokens[k].Text == centerToken {
				out = append(out, bNeighbor{blockIdx: bi, left: blk.Tokens[k-1].Text, right: blk.Tokens[k+1].Text})
			}
		}
	}
	return out
}

// jointTable builds the (left,right) contingency table plus left/right
// marginal counts from a set of B-neighbor observations.
func jointTable(obs []bNeighbor) (joint map[[2]string]int, left map[string]int, right map[string]int) {
	joint = map[[2]string]int{}
	left = map[string]int{}
	right = map[string]int{}
	for _, o := range obs {
		joint[[2]string{o.left, o.right}]++
		left[o.left]++
		right[o.right]++
	}
	return
}

// cmiBits is the plug-in conditional mutual information estimate I(X;Y|B=b)
// in bits from a fixed set of B-neighbor observations, task22 section 17.
func cmiBits(obs []bNeighbor) float64 {
	if len(obs) == 0 {
		return 0
	}
	joint, left, right := jointTable(obs)
	n := float64(len(obs))
	cmi := 0.0
	// Iterate pairs in a fixed sorted order rather than Go's randomized map
	// iteration order: summing float64 terms in a different order changes the
	// result in the low bits, which would silently break the byte-identical
	// reproducibility task22 section 90 requires.
	for _, pair := range sortedPairs(joint) {
		jn := joint[pair]
		pxy := float64(jn) / n
		px := float64(left[pair[0]]) / n
		py := float64(right[pair[1]]) / n
		if pxy > 0 && px > 0 && py > 0 {
			cmi += pxy * log2Ratio(pxy, px*py)
		}
	}
	return cmi
}

func sortedPairs(m map[[2]string]int) [][2]string {
	out := make([][2]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i][0] != out[j][0] {
			return out[i][0] < out[j][0]
		}
		return out[i][1] < out[j][1]
	})
	return out
}

func log2Ratio(a, b float64) float64 {
	return math.Log2(a / b)
}

// permuteWithinBlocks shuffles the right-neighbor values among B occurrences
// independently within each physical block (task22 sections 20-21): it
// preserves count(B), the left-neighbor distribution, the right-neighbor
// distribution, and block composition, while destroying the left<->right
// pairing conditional on B. Nothing is ever permuted across blocks.
func permuteWithinBlocks(obs []bNeighbor, r *rand.Rand) []bNeighbor {
	out := make([]bNeighbor, len(obs))
	copy(out, obs)
	byBlock := map[int][]int{}
	for i, o := range out {
		byBlock[o.blockIdx] = append(byBlock[o.blockIdx], i)
	}
	// Blocks must be visited in a fixed order (not Go's randomized map
	// iteration order): the shuffles below consume r's random stream in
	// sequence, so a different visitation order would consume that stream
	// differently and break reproducibility for a fixed seed.
	blockIdxs := make([]int, 0, len(byBlock))
	for bi := range byBlock {
		blockIdxs = append(blockIdxs, bi)
	}
	sort.Ints(blockIdxs)
	for _, bi := range blockIdxs {
		idxs := byBlock[bi]
		rights := make([]string, len(idxs))
		for j, idx := range idxs {
			rights[j] = out[idx].right
		}
		r.Shuffle(len(rights), func(a, b int) { rights[a], rights[b] = rights[b], rights[a] })
		for j, idx := range idxs {
			out[idx].right = rights[j]
		}
	}
	return out
}

// runCMI implements the whole of Part D for one candidate: the observed
// CMI around its central token B, a conditional-neighbor permutation null
// (never relying on the plug-in estimator alone - section 19), and the
// specific frozen (A,C) cell's pointwise contribution to that CMI.
func runCMI(cand Candidate, blocks []Block, permutations int, seed int64) CMIResult {
	obs := collectBNeighbors(cand.B(), blocks)
	observed := cmiBits(obs)
	r := rand.New(rand.NewSource(seed))
	null := make([]float64, permutations)
	for i := 0; i < permutations; i++ {
		null[i] = cmiBits(permuteWithinBlocks(obs, r))
	}
	mean, sd := meanSD(null)
	joint, left, right := jointTable(obs)
	n := float64(len(obs))
	contribution := 0.0
	if n > 0 {
		jn := float64(joint[[2]string{cand.A(), cand.C()}])
		pxy := jn / n
		px := float64(left[cand.A()]) / n
		py := float64(right[cand.C()]) / n
		if pxy > 0 && px > 0 && py > 0 {
			contribution = pxy * log2Ratio(pxy, px*py)
		}
	}
	return CMIResult{
		Sequence: cand.Sequence, CenterToken: cand.B(), Occurrences: len(obs),
		ObservedCMIBits: observed, NullMeanCMIBits: mean, NullSDCMIBits: sd,
		Permutations: permutations, EmpiricalP: empiricalP(observed, null),
		ContributionBits: contribution,
	}
}
