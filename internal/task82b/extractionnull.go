package task82b

import (
	"math/rand"
	"sort"
)

// RandomSubsequenceMatched is ACROSTIC NULL 1 (task82b.txt sec.34): choose
// as many atoms of the operator's own Kind as the operator chose, sampled
// uniformly at random without replacement from the whole carrier, kept in
// ascending (corpus) order -- a *subsequence*, not a shuffle.
func RandomSubsequenceMatched(sel Selection, seed int64) Selection {
	r := rand.New(rand.NewSource(seed))
	k := len(sel.Chosen)
	pool := sel.CandidatePool
	chosen := sampleWithoutReplacement(r, pool, k)
	return Selection{Kind: sel.Kind, Chosen: chosen, NullClass: "RANDOM_SUBSEQUENCE_MATCHED", CandidatePool: pool}
}

// StratifiedRandom is ACROSTIC NULL 2 (task82b.txt sec.35): for every
// group the operator actually picked from, choose one candidate from that
// same group uniformly at random instead of the operator's fixed rule.
// Only meaningful for operators with NullClass=="PER_GROUP".
func StratifiedRandom(sel Selection, seed int64) Selection {
	r := rand.New(rand.NewSource(seed))
	out := Selection{Kind: sel.Kind, NullClass: "POSITION_STRATIFIED_RANDOM"}
	for _, gid := range sel.GroupOf {
		cand := sel.GroupCandidates[gid]
		if len(cand) == 0 {
			continue
		}
		pick := cand[r.Intn(len(cand))]
		out.Chosen = append(out.Chosen, pick)
	}
	sortInts(out.Chosen)
	return out
}

// PeriodicPhases is ACROSTIC NULL 3 (task82b.txt sec.36): every other
// phase of the same period, computed deterministically (no seed).
func PeriodicPhases(sel Selection) map[int]Selection {
	out := map[int]Selection{}
	for phase := 0; phase < sel.Period; phase++ {
		if phase == sel.Phase {
			continue
		}
		var chosen []int
		for i := 0; i < sel.CandidatePool; i++ {
			if i%sel.Period == phase {
				chosen = append(chosen, i)
			}
		}
		out[phase] = Selection{Kind: sel.Kind, Chosen: chosen, NullClass: "PERIODIC_PHASE", Period: sel.Period, Phase: phase, CandidatePool: sel.CandidatePool}
	}
	return out
}

func sampleWithoutReplacement(r *rand.Rand, pool, k int) []int {
	if k > pool {
		k = pool
	}
	if k <= 0 {
		return nil
	}
	// Reservoir-free approach: shuffle index list is O(pool); fine at this
	// grid's scale (pool bounded by carrier token/glyph counts, <10^6).
	perm := r.Perm(pool)
	chosen := append([]int{}, perm[:k]...)
	sortInts(chosen)
	return chosen
}

func sortInts(xs []int) { sort.Ints(xs) }
