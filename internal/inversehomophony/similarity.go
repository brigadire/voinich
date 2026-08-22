package inversehomophony

import "math"

// jsDivergence is the Jensen-Shannon divergence between two count
// distributions (over possibly-disjoint supports), base-2, in [0,1]. It is
// symmetric and well-defined for disjoint support (unlike KL divergence),
// which matters here since two rare cipher types can have entirely
// non-overlapping observed contexts.
func jsDivergence(a, b map[string]int) float64 {
	ta, tb := sumCounts(a), sumCounts(b)
	if ta == 0 && tb == 0 {
		return 0
	}
	if ta == 0 || tb == 0 {
		return 1
	}
	keySet := make(map[string]struct{}, len(a)+len(b))
	for k := range a {
		keySet[k] = struct{}{}
	}
	for k := range b {
		keySet[k] = struct{}{}
	}
	// Sorted, not map-iteration order: float64 addition is not
	// associative, so summing klAM/klBM in map order would make the
	// result depend on Go's randomized map iteration rather than only on
	// (a,b) - see the project's "Go map iteration determinism" convention.
	keys := make([]string, 0, len(keySet))
	for k := range keySet {
		keys = append(keys, k)
	}
	sortStrings(keys)
	var klAM, klBM float64
	for _, k := range keys {
		pa := float64(a[k]) / float64(ta)
		pb := float64(b[k]) / float64(tb)
		m := 0.5 * (pa + pb)
		if pa > 0 {
			klAM += pa * math.Log2(pa/m)
		}
		if pb > 0 {
			klBM += pb * math.Log2(pb/m)
		}
	}
	js := 0.5*klAM + 0.5*klBM
	if js < 0 {
		js = 0
	}
	if js > 1 {
		js = 1
	}
	return js
}

func sumCounts(m map[string]int) int {
	s := 0
	for _, v := range m {
		s += v
	}
	return s
}

// similarity turns a divergence in [0,1] into a similarity in [0,1].
func similarity(divergence float64) float64 { return 1 - divergence }

// CombinedScore computes the four component similarities (predecessor,
// successor, distance-context, positional) and their unweighted mean
// (task57 section 18, frozen equal weighting - INVERSE_HOMOPHONY_DESIGN.md
// section 4). support is the combined predecessor+successor observation
// count on both tokens, used by CandidatePairs as an evidence floor.
func CombinedScore(fa, fb *TokenFeatures) (score float64, predS, succS, distS, posS float64, support int) {
	predS = similarity(jsDivergence(fa.Pred, fb.Pred))
	succS = similarity(jsDivergence(fa.Succ, fb.Succ))
	distS = similarity(jsDivergence(fa.DistCtx, fb.DistCtx))
	posS = similarity(jsDivergence(histToMap(fa.PosHist), histToMap(fb.PosHist)))
	score = (predS + succS + distS + posS) / 4
	support = min(sumCounts(fa.Pred)+sumCounts(fa.Succ), sumCounts(fb.Pred)+sumCounts(fb.Succ))
	return
}

func histToMap(h []int) map[string]int {
	m := make(map[string]int, len(h))
	for i, v := range h {
		if v > 0 {
			m[itoa(i)] = v
		}
	}
	return m
}

func itoa(i int) string {
	// Tiny fixed-width helper: PositionalBuckets is always small (<=5 in
	// the frozen config), so a byte-based conversion is fine and avoids an
	// extra import for such a small hot loop.
	if i < 10 {
		return string([]byte{byte('0' + i)})
	}
	return string([]byte{byte('0' + i/10), byte('0' + i%10)})
}

// CandidatePairs enumerates candidate pairs and scores them (task57
// section 18). A full O(n^2) enumeration over all distinct cipher types is
// not tractable at the vocabulary sizes homophonic expansion produces
// (tens of thousands of types), so candidates are first narrowed by a
// standard blocking pass: two tokens are only scored if they share at
// least one predecessor or successor context token. Two tokens with zero
// predecessor/successor overlap always score similarity 0 on those two
// components (disjoint-support Jensen-Shannon divergence is exactly 1 -
// see jsDivergence), so blocking on "shares no local context at all"
// discards only pairs that were already very unlikely to clear any
// reasonable threshold - this is a scalability step, applied identically
// to every corpus, fixed before any corpus is scored, not a
// corpus-specific tuning. Pairs are returned sorted by descending Score,
// ties broken lexicographically on (A,B) for determinism.
func CandidatePairs(features map[string]*TokenFeatures, cfg Config) []PairScore {
	tokens := make([]string, 0, len(features))
	for t := range features {
		tokens = append(tokens, t)
	}
	sortStrings(tokens)

	candidates := blockCandidates(features, tokens)

	pairs := make([]PairScore, 0, len(candidates))
	for _, pk := range candidates {
		fa, fb := features[pk.a], features[pk.b]
		score, predS, succS, distS, posS, support := CombinedScore(fa, fb)
		if support < cfg.MinSupport {
			continue
		}
		pairs = append(pairs, PairScore{
			A: pk.a, B: pk.b, Support: support,
			PredScore: predS, SuccScore: succS, DistScore: distS, PosScore: posS,
			Score: score,
		})
	}
	sortPairScores(pairs)
	return pairs
}

type tokenPair struct{ a, b string }

// blockCandidates returns every unordered pair of distinct tokens sharing
// at least one predecessor or successor context token, each pair
// canonicalized (a<b) and deduplicated.
func blockCandidates(features map[string]*TokenFeatures, tokens []string) []tokenPair {
	buckets := make(map[string][]string)
	addToBucket := func(ctx, owner string) {
		buckets[ctx] = append(buckets[ctx], owner)
	}
	for _, t := range tokens {
		f := features[t]
		for ctx := range f.Pred {
			addToBucket("p:"+ctx, t)
		}
		for ctx := range f.Succ {
			addToBucket("s:"+ctx, t)
		}
	}
	seen := make(map[tokenPair]struct{})
	var out []tokenPair
	for _, owners := range buckets {
		if len(owners) < 2 {
			continue
		}
		for i := 0; i < len(owners); i++ {
			for j := i + 1; j < len(owners); j++ {
				a, b := owners[i], owners[j]
				if a == b {
					continue
				}
				if a > b {
					a, b = b, a
				}
				pk := tokenPair{a, b}
				if _, ok := seen[pk]; ok {
					continue
				}
				seen[pk] = struct{}{}
				out = append(out, pk)
			}
		}
	}
	return out
}
