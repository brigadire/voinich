package localregime

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

var sweepRadii = []int{50, 100, 200, 500}
var sweepGaps = []int{10, 20, 30, 50}
var windowSizes = []int{50, 100, 200, 500}
var blockSizes = []int{50, 100, 200}

func defaults(c Config) Config {
	if c.RegimeRadius <= 0 {
		c.RegimeRadius = 100
	}
	if c.RegimeGap < 0 {
		c.RegimeGap = 20
	}
	if c.RegimeControlsK <= 0 {
		c.RegimeControlsK = 5
	}
	if c.WindowStep <= 0 {
		c.WindowStep = 10
	}
	if c.MaxDistance <= 0 {
		c.MaxDistance = 20
	}
	if c.TopN <= 0 {
		c.TopN = 28
	}
	if c.Seed == 0 {
		c.Seed = 150015
	}
	return c
}

func analyze(cfg Config, pgr *progressReporter) (analysis, error) {
	if cfg.RegimeRadius <= 0 || cfg.RegimeGap < 0 || cfg.RegimeGap >= cfg.RegimeRadius {
		return analysis{}, fmt.Errorf("regime radius must be positive and greater than regime gap")
	}
	radii := uniqueInts(append(append([]int{}, sweepRadii...), cfg.RegimeRadius))
	gaps := uniqueInts(append(append([]int{}, sweepGaps...), cfg.RegimeGap))
	maxRadius := radii[len(radii)-1]
	pgr.begin(1, "Loading corpus and target pairs")
	c, e := readCorpus(cfg.CorpusPath)
	if e != nil {
		return analysis{}, e
	}
	pairs, e := readPairs(cfg.DistancePairsPath, cfg.TopN)
	if e != nil {
		return analysis{}, e
	}
	if cfg.Pair != "" {
		v := strings.Split(cfg.Pair, ",")
		if len(v) != 2 {
			return analysis{}, fmt.Errorf("-pair must be tokenA,tokenB")
		}
		pairs = []pair{{strings.TrimSpace(v[0]), strings.TrimSpace(v[1])}}
	}
	pgr.update(1, 1, "Loading corpus and target pairs")
	needed := map[string]bool{}
	for _, q := range pairs {
		needed[q.A] = true
		needed[q.B] = true
	}
	pgr.begin(2, "Building occurrence regime profiles")
	occ := map[string][]profile{}
	posByToken := map[string][]int{}
	offsets := map[string]offsetCounts{}
	for t := range needed {
		posByToken[t] = positions(c, t)
	}
	total := len(needed)
	done := 0
	// Primary occurrence profiles are retained; sweep profiles are aggregated on demand below.
	for t := range needed {
		for _, i := range posByToken[t] {
			occ[t] = append(occ[t], localProfile(c, i, cfg.RegimeRadius, cfg.RegimeGap, "symmetric", cfg.RespectLineBoundaries))
		}
		offsets[t] = buildOffsetCounts(c, posByToken[t], maxRadius, cfg.RespectLineBoundaries)
		done++
		pgr.update(done, total, "Building occurrence regime profiles")
	}
	pgr.begin(3, "Regime sweeps and pair metrics")
	results := make([]PairResult, 0, len(pairs))
	regByPair := map[pair]float64{}
	for pi, q := range pairs {
		r := PairResult{TokenA: q.A, TokenB: q.B, CountA: c.Counts[q.A], CountB: c.Counts[q.B]}
		for _, radius := range radii {
			for _, gap := range gaps {
				if gap >= radius {
					continue
				}
				for _, side := range []string{"symmetric", "left", "right"} {
					ac, bc := offsetProfile(offsets[q.A], radius, gap, side), offsetProfile(offsets[q.B], radius, gap, side)
					sac, sbc := sortProfile(ac), sortProfile(bc)
					m := RegimeMetric{Radius: radius, Gap: gap, Side: side, JSSimilarity: jsSimilaritySorted(sac, sbc), WeightedOverlap: weightedOverlapSorted(sac, sbc), Jaccard: jaccard(ac, bc), Cosine: cosineSorted(sac, sbc)}
					r.Regimes = append(r.Regimes, m)
					if radius == cfg.RegimeRadius && gap == cfg.RegimeGap && side == "symmetric" {
						r.Regimes[len(r.Regimes)-1].DispersionA = dispersion(occ[q.A], aggregate(occ[q.A]))
						r.Regimes[len(r.Regimes)-1].DispersionB = dispersion(occ[q.B], aggregate(occ[q.B]))
						r.Regimes[len(r.Regimes)-1].PairwiseJSA = pairwiseDispersion(occ[q.A])
						r.Regimes[len(r.Regimes)-1].PairwiseJSB = pairwiseDispersion(occ[q.B])
						regByPair[q] = m.JSSimilarity
						r.PrimaryRegime = m.JSSimilarity
						r.ConcentrationA = concentrationSorted(sac)
						r.ConcentrationB = concentrationSorted(sbc)
					}
				}
			}
		}
		results = append(results, r)
		pgr.update(pi+1, len(pairs), "Regime sweeps and pair metrics")
	}
	pgr.begin(4, "Regime-conditioned residuals")
	pool := buildControlPool(c, cfg)
	expectedCache := map[string]matchedFuture{}
	for pi, q := range pairs {
		if expectedCache[q.A] == nil {
			expectedCache[q.A] = matchedExpected(c, q.A, posByToken[q.A], occ[q.A], pool, cfg)
		}
		if expectedCache[q.B] == nil {
			expectedCache[q.B] = matchedExpected(c, q.B, posByToken[q.B], occ[q.B], pool, cfg)
		}
		expectedA, expectedB := expectedCache[q.A], expectedCache[q.B]
		for d := 1; d <= cfg.MaxDistance; d++ {
			obs := jsSimilarity(distanceDistributionAt(c, posByToken[q.A], d, cfg.RespectLineBoundaries), distanceDistributionAt(c, posByToken[q.B], d, cfg.RespectLineBoundaries))
			exp := jsSimilarity(expectedAt(expectedA, d), expectedAt(expectedB, d))
			results[pi].Distance = append(results[pi].Distance, DistanceMetric{Distance: d, Observed: obs, RegimeExpected: exp, ResidualExcess: residualDependency(obs, exp)})
		}
		summarizePair(&results[pi])
		pgr.update(pi+1, len(pairs), "Regime-conditioned residuals")
	}
	pgr.begin(5, "Global, line, and local-block shuffles")
	global := shuffleCorpus(c, "global", 0, cfg.Seed)
	line := shuffleCorpus(c, "line", 0, cfg.Seed+1)
	blocks := map[int]corpus{}
	for _, b := range blockSizes {
		blocks[b] = shuffleCorpus(c, "block", b, cfg.Seed+int64(b))
	}
	shuffles := []ShuffleResult{}
	for pi, q := range pairs {
		posGlobalA, posGlobalB := positions(global, q.A), positions(global, q.B)
		posLineA, posLineB := positions(line, q.A), positions(line, q.B)
		posBlockA, posBlockB := map[int][]int{}, map[int][]int{}
		for _, b := range blockSizes {
			posBlockA[b] = positions(blocks[b], q.A)
			posBlockB[b] = positions(blocks[b], q.B)
		}
		for d := 1; d <= cfg.MaxDistance; d++ {
			g := jsSimilarity(distanceDistributionAt(global, posGlobalA, d, cfg.RespectLineBoundaries), distanceDistributionAt(global, posGlobalB, d, cfg.RespectLineBoundaries))
			l := jsSimilarity(distanceDistributionAt(line, posLineA, d, cfg.RespectLineBoundaries), distanceDistributionAt(line, posLineB, d, cfg.RespectLineBoundaries))
			vals := []float64{}
			for _, b := range blockSizes {
				s := jsSimilarity(distanceDistributionAt(blocks[b], posBlockA[b], d, cfg.RespectLineBoundaries), distanceDistributionAt(blocks[b], posBlockB[b], d, cfg.RespectLineBoundaries))
				vals = append(vals, s)
				shuffles = append(shuffles, ShuffleResult{q.A, q.B, "local-block", b, d, s})
			}
			local := mean(vals)
			base := g
			dm := &results[pi].Distance[d-1]
			dm.GlobalShuffle = g
			dm.LineShuffle = l
			dm.LocalBlockShuffle = local
			if dm.Observed != 0 {
				dm.RetainedFraction = local / dm.Observed
			}
			dm.Baseline = base
			dm.RetainedEffect = retainedEffect(dm.Observed, local, base)
			shuffles = append(shuffles, ShuffleResult{q.A, q.B, "global", 0, d, g}, ShuffleResult{q.A, q.B, "line-preserving", 0, d, l})
		}
		pgr.update(pi+1, len(pairs), "Global, line, and local-block shuffles")
	}
	pgr.begin(6, "Sliding windows and change points")
	var windows []WindowRow
	var changes []ChangePoint
	var separations []SeparationRow
	windowProfiles := map[int][]profile{}
	for wi, size := range windowSizes {
		ps, rows := slidingProfiles(c, size, cfg.WindowStep)
		windowProfiles[size] = ps
		windows = append(windows, rows...)
		changes = append(changes, changePoints(rows)...)
		sortedPs := make([]sortedProfile, len(ps))
		for i, p := range ps {
			sortedPs[i] = sortProfile(p)
		}
		for _, sep := range []int{1, 2, 5, 10, 20} {
			var ds []float64
			for i := 0; i+sep < len(ps); i++ {
				ds = append(ds, 1-jsSimilaritySorted(sortedPs[i], sortedPs[i+sep]))
			}
			separations = append(separations, SeparationRow{size, sep, len(ds), mean(ds), 1 - mean(ds)})
		}
		pgr.update(wi+1, len(windowSizes), "Sliding windows and change points")
	}
	tokens := tokenProfiles(c, occ, windowProfiles[100], cfg.WindowStep)
	controls, controlErr := readControlPairs(cfg.ControlsPath)
	if controlErr != nil {
		controls = matchedPairControls(results, c)
	}
	controls = evaluateControls(controls, pairs, c, cfg)
	reg, dist := []float64{}, []float64{}
	for i, q := range pairs {
		reg = append(reg, regByPair[q])
		dist = append(dist, results[i].Observed1To5)
	}
	corr := []Correlation{{"regime_js_vs_distance_js_1_5", len(reg), pearson(reg, dist), spearman(reg, dist)}}
	out := Output{Parameters: map[string]any{"corpus": cfg.CorpusPath, "primary_radius": cfg.RegimeRadius, "primary_gap": cfg.RegimeGap, "radius_sweep": radii, "gap_sweep": gaps, "window_sizes": windowSizes, "window_step": cfg.WindowStep, "block_sizes": blockSizes, "regime_controls_k": cfg.RegimeControlsK, "seed": cfg.Seed, "respect_line_boundaries": cfg.RespectLineBoundaries}, Pairs: results, Correlations: corr, Separations: separations}
	return analysis{Out: out, Windows: windows, Changes: changes, Tokens: tokens, Shuffles: shuffles, Occurrence: occ, Controls: controls}, nil
}

func uniqueInts(x []int) []int {
	seen := map[int]bool{}
	var out []int
	for _, v := range x {
		if v >= 0 && !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	sort.Ints(out)
	return out
}

type matchCandidate struct {
	pos     int
	token   string
	count   int
	profile profile
	sorted  sortedProfile
	norm    float64
}

func buildControlPool(c corpus, cfg Config) []matchCandidate {
	stride := 1
	if len(c.Tokens) > 512 {
		stride = (len(c.Tokens) + 511) / 512
	}
	out := make([]matchCandidate, 0)
	for i := 0; i < len(c.Tokens); i += stride {
		p := localProfile(c, i, cfg.RegimeRadius, cfg.RegimeGap, "symmetric", cfg.RespectLineBoundaries)
		sp := sortProfile(p)
		out = append(out, matchCandidate{i, c.Tokens[i], c.Counts[c.Tokens[i]], p, sp, math.Sqrt(concentrationSorted(sp))})
	}
	return out
}

type matchedFuture map[int]profile

func matchedExpected(c corpus, token string, pos []int, ps []profile, pool []matchCandidate, cfg Config) matchedFuture {
	counts := map[int]map[string]int{}
	for d := 1; d <= cfg.MaxDistance; d++ {
		counts[d] = map[string]int{}
	}
	for oi, p := range ps {
		type scored struct {
			i int
			s float64
		}
		var best []scored
		targetCount := float64(max(1, c.Counts[token]))
		sp := sortProfile(p)
		pn := math.Sqrt(concentrationSorted(sp))
		for i, x := range pool {
			if x.token == token {
				continue
			}
			ratio := math.Abs(math.Log(float64(max(1, x.count)) / targetCount))
			sim := 0.0
			if pn > 0 && x.norm > 0 {
				sim = dotProductSorted(sp, x.sorted) / (pn * x.norm)
			}
			score := sim - .05*ratio
			if len(best) < cfg.RegimeControlsK {
				best = append(best, scored{i, score})
				sort.Slice(best, func(i, j int) bool { return best[i].s > best[j].s })
			} else if score > best[len(best)-1].s {
				best[len(best)-1] = scored{i, score}
				sort.Slice(best, func(i, j int) bool { return best[i].s > best[j].s })
			}
		}
		_ = oi
		_ = pos
		for _, b := range best {
			base := pool[b.i].pos
			for d := 1; d <= cfg.MaxDistance; d++ {
				if base+d < len(c.Tokens) && (!cfg.RespectLineBoundaries || c.LineAt[base+d] == c.LineAt[base]) {
					counts[d][c.Tokens[base+d]]++
				}
			}
		}
	}
	out := matchedFuture{}
	for d, m := range counts {
		out[d] = normalizeCounts(m)
	}
	return out
}
func expectedAt(m matchedFuture, d int) profile { return m[d] }
func rangeMean(x []DistanceMetric, lo, hi int, field func(DistanceMetric) float64) float64 {
	var v []float64
	for _, q := range x {
		if q.Distance >= lo && q.Distance <= hi {
			v = append(v, field(q))
		}
	}
	return mean(v)
}
func summarizePair(p *PairResult) {
	obs := func(x DistanceMetric) float64 { return x.Observed }
	exp := func(x DistanceMetric) float64 { return x.RegimeExpected }
	res := func(x DistanceMetric) float64 { return x.ResidualExcess }
	p.Observed1To5 = rangeMean(p.Distance, 1, 5, obs)
	p.Observed6To10 = rangeMean(p.Distance, 6, 10, obs)
	p.Observed11To20 = rangeMean(p.Distance, 11, 20, obs)
	p.Regime1To5 = rangeMean(p.Distance, 1, 5, exp)
	p.Regime6To10 = rangeMean(p.Distance, 6, 10, exp)
	p.Regime11To20 = rangeMean(p.Distance, 11, 20, exp)
	p.Residual1To5 = rangeMean(p.Distance, 1, 5, res)
	p.Residual6To10 = rangeMean(p.Distance, 6, 10, res)
	p.Residual11To20 = rangeMean(p.Distance, 11, 20, res)
}
func tokenProfiles(c corpus, occ map[string][]profile, windows []profile, step int) []TokenProfile {
	var out []TokenProfile
	for t, ps := range occ {
		cent := aggregate(ps)
		membership := profile{}
		for _, pos := range positions(c, t) {
			if len(windows) > 0 {
				idx := pos / step
				if idx >= len(windows) {
					idx = len(windows) - 1
				}
				membership[fmt.Sprint(idx)] += 1
			}
		}
		n := float64(len(ps))
		if n > 0 {
			for k := range membership {
				membership[k] /= n
			}
		}
		h, maxv := 0., 0.
		if n > 0 {
			for _, k := range sortedProfileKeys(membership) {
				q := membership[k]
				if q > 0 {
					h -= q * math.Log2(q)
				}
				if q > maxv {
					maxv = q
				}
			}
		}
		out = append(out, TokenProfile{t, c.Counts[t], concentration(membership), h, maxv, dispersion(ps, cent)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Token < out[j].Token })
	return out
}
func matchedPairControls(results []PairResult, c corpus) []controlRow {
	tokens := make([]string, 0, len(c.Counts))
	for t, n := range c.Counts {
		if n >= 10 {
			tokens = append(tokens, t)
		}
	}
	sort.Strings(tokens)
	var out []controlRow
	for _, p := range results {
		type srow struct {
			q pair
			s float64
		}
		var rows []srow
		for i := 0; i < len(tokens); i++ {
			for j := i + 1; j < len(tokens); j++ {
				if tokens[i] == p.TokenA || tokens[i] == p.TokenB || tokens[j] == p.TokenA || tokens[j] == p.TokenB {
					continue
				}
				s := math.Abs(math.Log(float64(c.Counts[tokens[i]])/float64(max(1, p.CountA)))) + math.Abs(math.Log(float64(c.Counts[tokens[j]])/float64(max(1, p.CountB))))
				rows = append(rows, srow{pair{tokens[i], tokens[j]}, s})
			}
		}
		sort.Slice(rows, func(i, j int) bool { return rows[i].s < rows[j].s })
		for i := 0; i < min(5, len(rows)); i++ {
			out = append(out, controlRow{Target: pair{p.TokenA, p.TokenB}, Control: rows[i].q, Rank: i + 1, Score: rows[i].s})
		}
	}
	return out
}

func evaluateControls(rows []controlRow, targets []pair, c corpus, cfg Config) []controlRow {
	wanted := map[pair]bool{}
	for _, q := range targets {
		wanted[q] = true
	}
	profileCache := map[string]profile{}
	positionsCache := map[string][]int{}
	for i := range rows {
		if !wanted[rows[i].Target] {
			continue
		}
		for _, t := range []string{rows[i].Control.A, rows[i].Control.B} {
			if profileCache[t] == nil {
				ps := positions(c, t)
				positionsCache[t] = ps
				x := buildOffsetCounts(c, ps, cfg.RegimeRadius, cfg.RespectLineBoundaries)
				profileCache[t] = offsetProfile(x, cfg.RegimeRadius, cfg.RegimeGap, "symmetric")
			}
		}
		a, b := profileCache[rows[i].Control.A], profileCache[rows[i].Control.B]
		rows[i].RegimeSimilarity = jsSimilarity(a, b)
		rows[i].ConcentrationA, rows[i].ConcentrationB = concentration(a), concentration(b)
		posA, posB := positionsCache[rows[i].Control.A], positionsCache[rows[i].Control.B]
		var ds []float64
		for d := 1; d <= min(5, cfg.MaxDistance); d++ {
			ds = append(ds, jsSimilarity(distanceDistributionAt(c, posA, d, cfg.RespectLineBoundaries), distanceDistributionAt(c, posB, d, cfg.RespectLineBoundaries)))
		}
		rows[i].DistanceSimilarity = mean(ds)
	}
	var out []controlRow
	for _, r := range rows {
		if wanted[r.Target] {
			out = append(out, r)
		}
	}
	return out
}
