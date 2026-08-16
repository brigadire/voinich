package conditionalregime

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"zcore.dev/voinich/internal/globalregime"
)

// shuffleWithinBlocks is Null A (task19 section 20): every physical block's
// own tokens are independently permuted. Unigram frequencies, block lengths
// and window lengths are exactly preserved; sequential structure is
// destroyed. Positions outside the given blocks are left untouched.
func shuffleWithinBlocks(tokens []string, blocks []Block, rng *rand.Rand) []string {
	out := append([]string(nil), tokens...)
	for _, b := range blocks {
		seg := out[b.Start:b.End]
		rng.Shuffle(len(seg), func(i, j int) { seg[i], seg[j] = seg[j], seg[i] })
	}
	return out
}

// refinementCandidateLimit and qualifiesForRefinement together are the
// formal, pre-registered rule (task19 section 22) deciding which
// (class, window_size, method) combinations are re-run at 10000 permutations
// after the primary 1000-permutation pass. This rule is fixed in code before
// any real result is examined and is never adjusted afterward.
const refinementCandidateLimit = 5

func qualifiesForRefinement(s EmpiricalStats) bool {
	return s.EmpiricalP < 0.01 && s.EffectSize >= 2.0
}

// WithinClassCandidate is one class x window_size x method significance
// test: the primary statistic is silhouette at the single K that had the
// highest observed silhouette for that method (avoiding a fresh K sweep
// inside the null loop, which would multiply the permutation cost by the K
// range for no benefit: the K was already fixed by the observed data).
type WithinClassCandidate struct {
	Class      ClassID
	WindowSize int
	Method     string
	K          int
	Stats      EmpiricalStats
	Refined    bool
}

// nullFitCap bounds the quadratic clustering cost of the permutation loops
// specifically (not the one-shot observed fit, which keeps the full
// maxClusterFitWindows-equivalent resolution). A null distribution's shape
// does not need the same fitting-sample size as the single observed value it
// is compared against, and this cap is what keeps a 1000-replicate primary
// pass tractable.
const nullFitCap = 60

// shuffleBlocksCompact copies only the given blocks' own tokens (not the
// whole corpus) into a small buffer, shuffles each block's segment
// independently (Null A), and returns blocks re-based to offsets inside that
// buffer. This avoids reallocating and copying the entire corpus on every
// one of the (typically thousands of) permutation replicates.
func shuffleBlocksCompact(tokens []string, blocks []Block, rng *rand.Rand) ([]string, []Block) {
	total := 0
	for _, b := range blocks {
		total += b.Len()
	}
	buf := make([]string, 0, total)
	rebased := make([]Block, len(blocks))
	for i, b := range blocks {
		start := len(buf)
		buf = append(buf, tokens[b.Start:b.End]...)
		seg := buf[start:]
		rng.Shuffle(len(seg), func(x, y int) { seg[x], seg[y] = seg[y], seg[x] })
		rebased[i] = Block{Class: b.Class, Index: b.Index, Start: start, End: len(buf)}
	}
	return buf, rebased
}

// nullSilhouetteAtK refits the given method/K on Null-A-shuffled tokens and
// returns the resulting silhouette, reusing the exact same clustering and
// diagnostic code path as the observed fit.
func nullSilhouetteAtK(tokens []string, blocks []Block, windowSize int, method string, k int, rng *rand.Rand) float64 {
	buf, rebased := shuffleBlocksCompact(tokens, blocks, rng)
	cw := buildClassWindows(buf, rebased, windowSize)
	if len(cw) < 2*k {
		return 0
	}
	idx := cappedSampleIndices(len(cw), nullFitCap)
	sample := make([]globalregime.Window, len(idx))
	for i, si := range idx {
		sample[i] = cw[si].Window
	}
	d := globalregime.DistanceMatrix(sample)
	var labels []int
	if method == "hierarchical" {
		labels = globalregime.HierarchicalLabels(len(sample), k, d)
	} else {
		labels = globalregime.KMedoids(d, k, rng.Int63())
	}
	return globalregime.Diagnostics(windowSize, method, k, labels, d).Silhouette
}

// withinClassSignificance runs the primary permutation Null-A test for every
// eligible class x window_size x method's best-observed K. Each replicate
// uses its own independently seeded rng (replicateSeed), so this combo is
// reproducible regardless of whether it runs in one pass or is checkpointed
// combo-by-combo.
func withinClassSignificance(tokens []string, class ClassID, blocks []Block, windowSize int, best map[string]WithinClassRegime, permutations int, seed int64) []WithinClassCandidate {
	out, _ := withinClassSignificanceParallel(context.Background(), 1, nil, tokens, class, blocks, windowSize, best, permutations, seed, nil, nil)
	return out
}

func withinClassSignificanceParallel(ctx context.Context, workers int, pool jobExecutor, tokens []string, class ClassID, blocks []Block, windowSize int, best map[string]WithinClassRegime, permutations int, seed int64, saved map[string]float64, onComplete func(JobResult)) ([]WithinClassCandidate, error) {
	var out []WithinClassCandidate
	for _, method := range []string{"k_medoids", "hierarchical"} {
		row, ok := best[method]
		if !ok {
			continue
		}
		salt := methodSalt(method)
		combination := string(class.Scheme) + "|" + class.Label() + "|" + fmt.Sprint(windowSize) + "|" + method
		completed := checkpointJobsFor(saved, "part_a_significance", combination, permutations)
		null, err := runIndexedReplicatesState(ctx, workers, pool, "part_a_significance", combination, permutations, nil, completed, func(ctx context.Context, i int) (float64, error) {
			if err := ctx.Err(); err != nil {
				return 0, err
			}
			rng := rand.New(rand.NewSource(replicateSeed(seed, salt, i)))
			return nullSilhouetteAtK(tokens, blocks, windowSize, method, row.K, rng), nil
		}, nil, onComplete)
		if err != nil {
			return nil, err
		}
		out = append(out, WithinClassCandidate{
			Class: class, WindowSize: windowSize, Method: method, K: row.K,
			Stats: buildEmpiricalStats(row.Silhouette, null),
		})
	}
	return out, nil
}

// refineTopCandidates re-runs the top refinementCandidateLimit qualifying
// candidates at refinementPermutations, using a distinct deterministic seed
// offset so the refinement pass is reproducible but independent of the
// primary pass's draws. blocksByClass must map every candidate's ClassID to
// its physical blocks.
func refineTopCandidates(tokens []string, blocksByClass map[ClassID][]Block, candidates []WithinClassCandidate, seed int64) []WithinClassCandidate {
	out, _ := refineTopCandidatesParallel(context.Background(), 1, nil, tokens, blocksByClass, candidates, seed, nil, nil)
	return out
}

func refineTopCandidatesParallel(ctx context.Context, workers int, pool jobExecutor, tokens []string, blocksByClass map[ClassID][]Block, candidates []WithinClassCandidate, seed int64, saved map[string]float64, onComplete func(JobResult)) ([]WithinClassCandidate, error) {
	type qualifying struct {
		idx int
		eff float64
	}
	var q []qualifying
	for i, c := range candidates {
		if qualifiesForRefinement(c.Stats) {
			q = append(q, qualifying{i, c.Stats.EffectSize})
		}
	}
	sort.SliceStable(q, func(i, j int) bool { return q[i].eff > q[j].eff })
	if len(q) > refinementCandidateLimit {
		q = q[:refinementCandidateLimit]
	}
	for _, entry := range q {
		c := candidates[entry.idx]
		blocks := blocksByClass[c.Class]
		salt := methodSalt(c.Method) + 100 // distinct stream from the primary pass
		combination := string(c.Class.Scheme) + "|" + c.Class.Label() + "|" + fmt.Sprint(c.WindowSize) + "|" + c.Method
		completed := checkpointJobsFor(saved, "part_a_refinement", combination, refinementPermutations)
		null, err := runIndexedReplicatesState(ctx, workers, pool, "part_a_refinement", combination, refinementPermutations, nil, completed, func(ctx context.Context, i int) (float64, error) {
			if err := ctx.Err(); err != nil {
				return 0, err
			}
			rng := rand.New(rand.NewSource(replicateSeed(seed+999983, salt, i)))
			return nullSilhouetteAtK(tokens, blocks, c.WindowSize, c.Method, c.K, rng), nil
		}, nil, onComplete)
		if err != nil {
			return nil, err
		}
		candidates[entry.idx].Stats = buildEmpiricalStats(c.Stats.Observed, null)
		candidates[entry.idx].Refined = true
	}
	return candidates, nil
}
