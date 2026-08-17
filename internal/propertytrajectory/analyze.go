package propertytrajectory

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
)

var requiredPairs = []pair{{"chedy", "qokeey"}, {"chol", "daiin"}, {"ol", "y"}, {"chey", "ol"}, {"dar", "ol"}, {"ar", "ol"}, {"dal", "ol"}, {"qokain", "qol"}, {"aiin", "ar"}, {"or", "s"}, {"r", "s"}, {"or", "r"}}
var modes = []string{"frequency-only", "graphemic-form-only", "position-only", "context-complexity-only", "structural-centrality-only", "all-properties", "all-minus-frequency", "all-minus-position", "all-minus-context", "all-minus-graphemic", "all-minus-structural"}

func defaults(c Config) Config {
	if c.MaxDistance <= 0 {
		c.MaxDistance = 20
	}
	if c.MinTokenFrequency <= 0 {
		c.MinTokenFrequency = 10
	}
	if c.TopN == 0 {
		c.TopN = 28
	}
	if c.RandomPairs <= 0 {
		c.RandomPairs = 1000
	}
	if c.Seed == 0 {
		c.Seed = 140014
	}
	return c
}
func selectPairs(previous []pair, single string, counts map[string]int) ([]pair, error) {
	if single != "" {
		x := strings.Split(single, ",")
		if len(x) != 2 || strings.TrimSpace(x[0]) == "" || strings.TrimSpace(x[1]) == "" {
			return nil, fmt.Errorf("--pair must be tokenA,tokenB")
		}
		return []pair{{strings.TrimSpace(x[0]), strings.TrimSpace(x[1])}}, nil
	}
	out := append([]pair(nil), previous...)
	for _, p := range requiredPairs {
		// The fixed list is supplemental evidence for the Voynich corpus.
		// Generic corpora should retain their upstream-discovered pairs
		// without failing on inapplicable Voynich-only token names.
		if counts[p.A] > 0 && counts[p.B] > 0 {
			out = append(out, p)
		}
	}
	seen := map[pair]bool{}
	var uniq []pair
	for _, p := range out {
		k := p
		if k.B < k.A {
			k.A, k.B = k.B, k.A
		}
		if !seen[k] {
			seen[k] = true
			uniq = append(uniq, p)
		}
	}
	return uniq, nil
}
func buildPairResult(cache trajectoryCache, p pair, c corpus, maxD int) PairResult {
	names := allPropertyNames()
	dp := profileCached(cache, p, maxD, names)
	r := PairResult{TokenA: p.A, TokenB: p.B, CountA: c.Counts[p.A], CountB: c.Counts[p.B], DistanceProfiles: dp}
	r.Summary.Cosine1To5 = rangeCos(dp, 1, 5)
	r.Summary.Cosine6To10 = rangeCos(dp, 6, 10)
	r.Summary.Cosine11To20 = rangeCos(dp, 11, 20)
	for _, mode := range modes {
		r.Summary.Modes = append(r.Summary.Modes, ModeScore{mode, scoreCached(cache, p, min(maxD, 5), modeNames(mode))})
	}
	var ranks []PropertyRanking
	for _, n := range names {
		var a, b []float64
		for _, d := range dp {
			a = append(a, d.Properties[n].MeanA)
			b = append(b, d.Properties[n].MeanB)
		}
		mad := 0.
		for i := range a {
			mad += math.Abs(a[i] - b[i])
		}
		if len(a) > 0 {
			mad /= float64(len(a))
		}
		ranks = append(ranks, PropertyRanking{n, pearson(a, b), mad})
	}
	sort.Slice(ranks, func(i, j int) bool {
		if ranks[i].MeanAbsoluteDifference == ranks[j].MeanAbsoluteDifference {
			return ranks[i].Property < ranks[j].Property
		}
		return ranks[i].MeanAbsoluteDifference < ranks[j].MeanAbsoluteDifference
	})
	r.Summary.StrongestMatching = append([]PropertyRanking(nil), ranks[:min(5, len(ranks))]...)
	sort.Slice(ranks, func(i, j int) bool {
		if ranks[i].MeanAbsoluteDifference == ranks[j].MeanAbsoluteDifference {
			return ranks[i].Property < ranks[j].Property
		}
		return ranks[i].MeanAbsoluteDifference > ranks[j].MeanAbsoluteDifference
	})
	r.Summary.StrongestDiffering = append([]PropertyRanking(nil), ranks[:min(5, len(ranks))]...)
	return r
}

// matchWorkspace holds the parts of fallbackMatched's candidate search that
// don't depend on which target pair is being matched: the O(eligible^2)
// pool of every eligible-token pair, and each eligible token's log1p count
// (fallbackMatched's score is a function of log1p(count), and math.Log1p
// is a pure function of an invariant input, so caching it instead of
// recomputing it on every one of an O(pool log pool) sort's comparisons
// changes nothing about the result). Built once per analyze() call and
// reused across every one of the ~40-80 fallbackMatched calls that
// otherwise each rebuilt and re-sorted this same pool from scratch - the
// audit's own "dominant pipeline cost" finding.
type matchWorkspace struct {
	allPairs []pair
	logCount map[string]float64
}

func prepareMatchWorkspace(c corpus, eligible []string) matchWorkspace {
	logCount := make(map[string]float64, len(eligible))
	for _, t := range eligible {
		logCount[t] = math.Log1p(float64(c.Counts[t]))
	}
	allPairs := make([]pair, 0, len(eligible)*(len(eligible)-1)/2)
	for i := 0; i < len(eligible); i++ {
		for j := i + 1; j < len(eligible); j++ {
			allPairs = append(allPairs, pair{eligible[i], eligible[j]})
		}
	}
	return matchWorkspace{allPairs: allPairs, logCount: logCount}
}

func fallbackMatched(target pair, c corpus, ws matchWorkspace, n int, r *rand.Rand) []pair {
	// target may be below the eligibility threshold (ws.logCount only
	// covers eligible tokens), so its own log1p is computed directly here
	// rather than risking a missing-key zero-value default from the cache.
	targetLogA := math.Log1p(float64(c.Counts[target.A]))
	targetLogB := math.Log1p(float64(c.Counts[target.B]))
	type scored struct {
		p pair
		s float64
	}
	pool := make([]scored, 0, len(ws.allPairs))
	for _, p := range ws.allPairs {
		if p.A == target.A || p.A == target.B || p.B == target.A || p.B == target.B {
			continue
		}
		s := math.Abs(ws.logCount[p.A]-targetLogA) + math.Abs(ws.logCount[p.B]-targetLogB)
		pool = append(pool, scored{p, s})
	}
	sort.Slice(pool, func(i, j int) bool {
		if pool[i].s == pool[j].s {
			return pool[i].p.A+pool[i].p.B < pool[j].p.A+pool[j].p.B
		}
		return pool[i].s < pool[j].s
	})
	limit := min(len(pool), max(n*10, n))
	pool = pool[:limit]
	out := make([]pair, len(pool))
	for i, x := range pool {
		out[i] = x.p
	}
	r.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out[:min(n, len(out))]
}
func analyze(cfg Config, pgr *progressReporter) (analysis, error) {
	c := defaults(cfg)
	pgr.begin(1, "Loading inputs")
	corp, e := readCorpus(c.CorpusPath)
	if e != nil {
		return analysis{}, e
	}
	prev, e := readPrevious(c.DistancePairsPath, c.TopN)
	if e != nil {
		return analysis{}, e
	}
	controls, e := readControls(c.ControlsPath)
	if e != nil {
		return analysis{}, e
	}
	edges, e := readStructural(c.StructuralPairsPath)
	if e != nil {
		return analysis{}, e
	}
	selected, e := selectPairs(prev, c.Pair, corp.Counts)
	if e != nil {
		return analysis{}, e
	}
	pgr.update(1, 1, "Loading inputs")
	pgr.begin(2, "Building and globally normalizing token properties")
	props, stats := buildProperties(corp, edges, c.MinTokenFrequency)
	eligible := make([]string, 0, len(props))
	var tokenRows []TokenProperties
	for t, p := range props {
		eligible = append(eligible, t)
		tokenRows = append(tokenRows, p)
	}
	sort.Strings(eligible)
	sort.Slice(tokenRows, func(i, j int) bool { return tokenRows[i].Token < tokenRows[j].Token })
	pgr.update(1, 1, "Building and globally normalizing token properties")
	for _, p := range selected {
		if _, ok := corp.Counts[p.A]; !ok {
			return analysis{}, fmt.Errorf("target token %q absent from corpus", p.A)
		}
		if _, ok := corp.Counts[p.B]; !ok {
			return analysis{}, fmt.Errorf("target token %q absent from corpus", p.B)
		}
	}
	out := Output{Parameters: map[string]any{"max_distance": c.MaxDistance, "min_token_frequency": c.MinTokenFrequency, "top": c.TopN, "random_pairs": c.RandomPairs, "seed": c.Seed, "rare_token_policy": "exclude from exact-distance aggregation and report excluded counts"}, Methodology: map[string]string{"identity": "subsequent token identities are retained for lookup only; trajectories aggregate intrinsic properties without projection or smoothing", "normalization": "heavy-tailed count/degree/unique/effective-count properties use log1p before one global eligible-token z-score; raw and normalized values are both retained", "distance": "rightward exact distance in the linear corpus; shuffled controls preserve global counts", "interpretation": "formal diagnostic only; no semantic, grammatical, morphological, page, or latent-state claims"}, PropertyGroups: propertyGroups, Normalization: stats}
	cache := buildTrajectoryCache(corp.Tokens, c.MaxDistance, props)
	pgr.begin(3, "Analyzing target property trajectories")
	for i, p := range selected {
		out.Pairs = append(out.Pairs, buildPairResult(cache, p, corp, c.MaxDistance))
		pgr.update(i+1, len(selected), "Analyzing target property trajectories")
	}
	rng := rand.New(rand.NewSource(c.Seed))
	matchWs := prepareMatchWorkspace(corp, eligible)
	matchedScores := map[pair][]float64{}
	randomScores := map[pair][]float64{}
	score := func(q pair, maxD int) float64 {
		return scoreCached(cache, q, maxD, allPropertyNames())
	}
	pgr.begin(4, "Frequency-matched controls")
	for i, p := range selected {
		x := controls[p]
		if len(x) == 0 {
			x = fallbackMatched(p, corp, matchWs, 20, rng)
		}
		for _, q := range x {
			matchedScores[p] = append(matchedScores[p], score(q, min(5, c.MaxDistance)))
		}
		if len(out.Pairs) > i {
			out.Pairs[i].Summary.MatchedPercentile = percentileRank(matchedScores[p], out.Pairs[i].Summary.Cosine1To5)
		}
		pgr.update(i+1, len(selected), "Frequency-matched controls")
	}
	pgr.begin(5, "Random frequency-matched baseline")
	distanceVals := make([][]float64, c.MaxDistance)
	rangeVals := []float64{}
	total := len(selected) * c.RandomPairs
	done := 0
	for i, p := range selected {
		qs := fallbackMatched(p, corp, matchWs, c.RandomPairs, rng)
		for _, q := range qs {
			cv := cosinesCached(cache, q, c.MaxDistance, allPropertyNames())
			s := mean(cv[:min(5, len(cv))])
			randomScores[p] = append(randomScores[p], s)
			rangeVals = append(rangeVals, s)
			for d, v := range cv {
				distanceVals[d] = append(distanceVals[d], v)
			}
			done++
			pgr.update(done, total, "Random frequency-matched baseline")
		}
		out.Pairs[i].Summary.RandomPercentile = percentileRank(randomScores[p], out.Pairs[i].Summary.Cosine1To5)
	}
	mkbase := func(scope string, d int, v []float64) Baseline {
		return Baseline{scope, d, quantile(v, .5), quantile(v, .9), quantile(v, .95), quantile(v, .99)}
	}
	out.RandomBaselines = append(out.RandomBaselines, mkbase("mean_1_5", 0, rangeVals))
	for d, v := range distanceVals {
		out.RandomBaselines = append(out.RandomBaselines, mkbase("exact_distance", d+1, v))
	}
	pgr.begin(6, "Global and line-preserving shuffled controls")
	for mi, mode := range []string{"global", "line-preserving"} {
		tokens := shuffleCorpus(corp, mode, rand.New(rand.NewSource(c.Seed+int64(mi)+1)))
		shuffledCache := buildTrajectoryCache(tokens, min(5, c.MaxDistance), props)
		for i, p := range selected {
			out.Shuffles = append(out.Shuffles, ShuffleResult{mode, p.A, p.B, scoreCached(shuffledCache, p, min(5, c.MaxDistance), allPropertyNames())})
			pgr.update(mi*len(selected)+i+1, 2*len(selected), "Global and line-preserving shuffled controls")
		}
	}
	return analysis{Out: out, Tokens: tokenRows, Matched: matchedScores, Random: randomScores}, nil
}
