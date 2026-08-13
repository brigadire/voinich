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
func selectPairs(previous []pair, single string) ([]pair, error) {
	if single != "" {
		x := strings.Split(single, ",")
		if len(x) != 2 || strings.TrimSpace(x[0]) == "" || strings.TrimSpace(x[1]) == "" {
			return nil, fmt.Errorf("--pair must be tokenA,tokenB")
		}
		return []pair{{strings.TrimSpace(x[0]), strings.TrimSpace(x[1])}}, nil
	}
	out := append([]pair(nil), previous...)
	out = append(out, requiredPairs...)
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
func fallbackMatched(target pair, c corpus, eligible []string, n int, r *rand.Rand) []pair {
	score := func(p pair) float64 {
		return math.Abs(math.Log1p(float64(c.Counts[p.A]))-math.Log1p(float64(c.Counts[target.A]))) + math.Abs(math.Log1p(float64(c.Counts[p.B]))-math.Log1p(float64(c.Counts[target.B])))
	}
	pool := make([]pair, 0)
	for i := 0; i < len(eligible); i++ {
		for j := i + 1; j < len(eligible); j++ {
			p := pair{eligible[i], eligible[j]}
			if p.A == target.A || p.A == target.B || p.B == target.A || p.B == target.B {
				continue
			}
			pool = append(pool, p)
		}
	}
	sort.Slice(pool, func(i, j int) bool {
		a, b := score(pool[i]), score(pool[j])
		if a == b {
			return pool[i].A+pool[i].B < pool[j].A+pool[j].B
		}
		return a < b
	})
	limit := min(len(pool), max(n*10, n))
	pool = pool[:limit]
	r.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	return pool[:min(n, len(pool))]
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
	selected, e := selectPairs(prev, c.Pair)
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
	matchedScores := map[pair][]float64{}
	randomScores := map[pair][]float64{}
	score := func(q pair, maxD int) float64 {
		return scoreCached(cache, q, maxD, allPropertyNames())
	}
	pgr.begin(4, "Frequency-matched controls")
	for i, p := range selected {
		x := controls[p]
		if len(x) == 0 {
			x = fallbackMatched(p, corp, eligible, 20, rng)
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
		qs := fallbackMatched(p, corp, eligible, c.RandomPairs, rng)
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
