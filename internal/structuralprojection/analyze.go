package structuralprojection

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

var mandatory = []pair{{"chedy", "qokeey"}, {"chol", "daiin"}, {"ol", "y"}, {"chey", "ol"}, {"dar", "ol"}, {"ar", "ol"}, {"dal", "ol"}, {"qokain", "qol"}, {"aiin", "ar"}, {"or", "s"}, {"r", "s"}, {"or", "r"}}

type analysis struct {
	Out       Output
	Families  []FamilyResult
	Sequences []SequenceResult
	Controls  []ControlRow
}

func validate(c Config) error {
	if c.MinStructuralSimilarity < 0 || c.MinStructuralSimilarity > 1 {
		return fmt.Errorf("min-structural-similarity must be in [0,1]")
	}
	if c.MinReliability < 0 || c.MinReliability > 1 {
		return fmt.Errorf("min-reliability must be in [0,1]")
	}
	if c.ProjectionK < 0 || c.RandomProjections < 1 || c.MaxDistance < 1 || c.MinObservations < 1 || c.TopN < 0 {
		return fmt.Errorf("invalid non-positive count parameter")
	}
	if c.ProjectionMode != "full" && c.ProjectionMode != "ablated" && c.ProjectionMode != "both" {
		return fmt.Errorf("projection-mode must be full, ablated, or both")
	}
	if c.Pair != "" {
		x := strings.Split(c.Pair, ",")
		if len(x) != 2 || strings.TrimSpace(x[0]) == "" || strings.TrimSpace(x[1]) == "" {
			return fmt.Errorf("pair must be tokenA,tokenB")
		}
	}
	return nil
}

func analyze(c Config, progress *progressReporter) (analysis, error) {
	if e := validate(c); e != nil {
		return analysis{}, e
	}
	progress.begin(1, "Loading inputs")
	corp, e := readCorpus(c.CorpusPath)
	if e != nil {
		return analysis{}, fmt.Errorf("read corpus: %w", e)
	}
	edges, e := readEdges(c.StructuralPairsPath)
	if e != nil {
		return analysis{}, fmt.Errorf("read structural pairs: %w", e)
	}
	prev, e := readPrevious(c.DistancePairsPath, c.TopN)
	if e != nil {
		return analysis{}, fmt.Errorf("read distance pairs: %w", e)
	}
	fams, e := readFamilies(c.FamiliesPath)
	if e != nil {
		return analysis{}, fmt.Errorf("read families: %w", e)
	}
	allFams := append([]family(nil), fams...)
	progress.update(1, 1, "Loading inputs")
	progress.begin(2, "Building structural and distance profiles")
	tokens := uniqueTokens(corp)
	full := BuildProjection(tokens, edges, c.MinStructuralSimilarity, c.MinReliability, c.ProjectionK, "full")
	future := BuildProjection(tokens, edges, c.MinStructuralSimilarity, c.MinReliability, c.ProjectionK, "future-ablated")
	past := BuildProjection(tokens, edges, c.MinStructuralSimilarity, c.MinReliability, c.ProjectionK, "past-ablated")
	prof := buildProfiles(corp, c.MaxDistance, false)
	progress.update(1, 1, "Building structural and distance profiles")
	selectedPrev := map[pair]previousPair{}
	selected := []pair{}
	seen := map[pair]bool{}
	canon := func(a, b string) pair {
		if b < a {
			a, b = b, a
		}
		return pair{a, b}
	}
	add := func(x pair) {
		x = canon(strings.TrimSpace(x.A), strings.TrimSpace(x.B))
		if !seen[x] {
			seen[x] = true
			selected = append(selected, x)
		}
	}
	for _, x := range prev {
		p := canon(x.TokenA, x.TokenB)
		selectedPrev[p] = x
	}
	if c.Pair != "" {
		x := strings.Split(c.Pair, ",")
		add(pair{x[0], x[1]})
		fams = nil
	} else if c.FamilyID > 0 {
		var chosen []family
		for _, f := range fams {
			if f.ID == c.FamilyID {
				chosen = append(chosen, f)
				for i := range f.Tokens {
					for j := 0; j < i; j++ {
						add(pair{f.Tokens[i], f.Tokens[j]})
					}
				}
			}
		}
		if len(chosen) == 0 {
			return analysis{}, fmt.Errorf("family %d not found", c.FamilyID)
		}
		fams = chosen
	} else {
		for _, x := range mandatory {
			add(x)
		}
		for _, x := range prev {
			add(pair{x.TokenA, x.TokenB})
		}
	}
	for _, x := range selected {
		if corp.Counts[x.A] == 0 || corp.Counts[x.B] == 0 {
			return analysis{}, fmt.Errorf("selected token absent from corpus: %s/%s", x.A, x.B)
		}
	}
	a := analysis{Out: Output{Parameters: map[string]any{"corpus": c.CorpusPath, "structural_pairs": c.StructuralPairsPath, "token_level_source": c.DistancePairsPath, "min_structural_similarity": c.MinStructuralSimilarity, "min_reliability": c.MinReliability, "projection_k": c.ProjectionK, "projection_mode": c.ProjectionMode, "random_projections": c.RandomProjections, "seed": c.Seed, "max_distance": c.MaxDistance}, Methodology: map[string]string{"full_weight": "raw_similarity * evidence_strength; edges below either configured threshold are zero", "ablated_future_weight": "reliability-weighted mean(position_similarity, left_similarity), multiplied by mean component reliability; right-context component excluded", "ablated_past_weight": "reliability-weighted mean(position_similarity, right_similarity), multiplied by mean component reliability; left-context component excluded", "normalization": "W(X,X)=1; every retained row is divided by its row sum; projected(Y)=sum_X P(X)*normalized_W(X,Y)", "family_projection": "control only: family ID categories plus singleton categories for every token outside a family", "random_space": "destination permutation within log2 frequency bins preserves row degree and weights approximately frequency-matched", "generic_smoothing": "uniform smoothing to the same number of random tokens in the source token's log2 frequency bin", "sequence_kernel": "arithmetic mean of position-wise projected-distribution JS similarities"}}}
	progress.begin(3, "Analyzing target pairs")
	for i, x := range selected {
		pr := compare(x, prof, full, future, past, selectedPrev[canon(x.A, x.B)], c)
		a.Out.Pairs = append(a.Out.Pairs, pr)
		a.Sequences = append(a.Sequences, sequenceResults(x, prof, full, future, c.MinObservations)...)
		progress.update(i+1, len(selected), "Analyzing target pairs")
	}
	// Random-space and generic-smoothing null distributions. Controls are based
	// on mean right-context gain at exact distances 1..5.
	randomByPair := make([][]float64, len(selected))
	smoothByPair := make([][]float64, len(selected))
	randomAblatedByPair := make([][]float64, len(selected))
	smoothAblatedByPair := make([][]float64, len(selected))
	randomByDistance := make([][][]float64, len(selected))
	for i := range randomByDistance {
		randomByDistance[i] = make([][]float64, min(5, c.MaxDistance))
	}
	progress.begin(4, "Random and smoothing controls")
	for trial := 0; trial < c.RandomProjections; trial++ {
		rp := RandomizeProjection(full, corp.Counts, c.Seed+int64(trial)*7919)
		sp := GenericSmoothing(tokens, corp.Counts, full, c.Seed+int64(trial)*104729)
		rap := RandomizeProjection(future, corp.Counts, c.Seed+int64(trial)*15485863)
		sap := GenericSmoothing(tokens, corp.Counts, future, c.Seed+int64(trial)*32452843)
		rcache, scache, racache, sacache := map[string][]map[string]float64{}, map[string][]map[string]float64{}, map[string][]map[string]float64{}, map[string][]map[string]float64{}
		projected := func(t string, proj Projection, cache map[string][]map[string]float64) []map[string]float64 {
			if x := cache[t]; x != nil {
				return x
			}
			x := make([]map[string]float64, min(5, c.MaxDistance))
			for d := range x {
				x[d] = ProjectDistribution(prof[t].Right[d], proj)
			}
			cache[t] = x
			return x
		}
		for i, x := range selected {
			gs, ss, gas, sas := []float64{}, []float64{}, []float64{}, []float64{}
			ra, rb := projected(x.A, rp, rcache), projected(x.B, rp, rcache)
			sa, sb := projected(x.A, sp, scache), projected(x.B, sp, scache)
			raa, rab := projected(x.A, rap, racache), projected(x.B, rap, racache)
			saa, sab := projected(x.A, sap, sacache), projected(x.B, sap, sacache)
			for d := 0; d < min(5, c.MaxDistance); d++ {
				tj, _, _ := metricsFloat(countsFloat(prof[x.A].Right[d]), countsFloat(prof[x.B].Right[d]))
				rj, _, _ := metricsFloat(ra[d], rb[d])
				sj, _, _ := metricsFloat(sa[d], sb[d])
				raj, _, _ := metricsFloat(raa[d], rab[d])
				saj, _, _ := metricsFloat(saa[d], sab[d])
				g, sg := rj-tj, sj-tj
				ga, sga := raj-tj, saj-tj
				gs = append(gs, g)
				ss = append(ss, sg)
				gas = append(gas, ga)
				sas = append(sas, sga)
				randomByDistance[i][d] = append(randomByDistance[i][d], g)
			}
			randomByPair[i] = append(randomByPair[i], mean(gs))
			smoothByPair[i] = append(smoothByPair[i], mean(ss))
			randomAblatedByPair[i] = append(randomAblatedByPair[i], mean(gas))
			smoothAblatedByPair[i] = append(smoothAblatedByPair[i], mean(sas))
		}
		progress.update(trial+1, c.RandomProjections, "Random and smoothing controls")
	}
	for i := range a.Out.Pairs {
		sort.Float64s(randomByPair[i])
		sort.Float64s(smoothByPair[i])
		sort.Float64s(randomAblatedByPair[i])
		sort.Float64s(smoothAblatedByPair[i])
		obs := a.Out.Pairs[i].Summary.Gain1To5
		ctl := GainControl{Observed: obs, RandomMean: mean(randomByPair[i]), RandomP95: quantile(randomByPair[i], .95), RandomPercentile: percentile(randomByPair[i], obs), SmoothingMean: mean(smoothByPair[i]), SmoothingP95: quantile(smoothByPair[i], .95), SmoothingPercentile: percentile(smoothByPair[i], obs)}
		a.Out.Pairs[i].Summary.Control = ctl
		a.Controls = append(a.Controls, ControlRow{selected[i].A, selected[i].B, "random_space", obs, ctl.RandomMean, ctl.RandomP95, ctl.RandomPercentile}, ControlRow{selected[i].A, selected[i].B, "generic_smoothing", obs, ctl.SmoothingMean, ctl.SmoothingP95, ctl.SmoothingPercentile})
		obsA := a.Out.Pairs[i].Summary.MeanAblated1To5 - a.Out.Pairs[i].Summary.MeanToken1To5
		actl := GainControl{Observed: obsA, RandomMean: mean(randomAblatedByPair[i]), RandomP95: quantile(randomAblatedByPair[i], .95), RandomPercentile: percentile(randomAblatedByPair[i], obsA), SmoothingMean: mean(smoothAblatedByPair[i]), SmoothingP95: quantile(smoothAblatedByPair[i], .95), SmoothingPercentile: percentile(smoothAblatedByPair[i], obsA)}
		a.Out.Pairs[i].Summary.AblatedControl = actl
		a.Controls = append(a.Controls, ControlRow{selected[i].A, selected[i].B, "random_space_ablated", obsA, actl.RandomMean, actl.RandomP95, actl.RandomPercentile}, ControlRow{selected[i].A, selected[i].B, "generic_smoothing_ablated", obsA, actl.SmoothingMean, actl.SmoothingP95, actl.SmoothingPercentile})
		for d := range randomByDistance[i] {
			sort.Float64s(randomByDistance[i][d])
			a.Out.Pairs[i].Right[d].RandomGainP95 = quantile(randomByDistance[i][d], .95)
		}
	}
	// Prespecified threshold and K sweeps are all retained, never optimized.
	progress.begin(5, "Projection parameter sweeps")
	sweepDone := 0
	for _, th := range []float64{.50, .60, .65, .70, .75} {
		fp := BuildProjection(tokens, edges, th, c.MinReliability, 0, "full")
		ap := BuildProjection(tokens, edges, th, c.MinReliability, 0, "future-ablated")
		a.Out.Sweeps = append(a.Out.Sweeps, SweepRow{"threshold", th, meanGain(selected, prof, fp), meanGain(selected, prof, ap)})
		sweepDone++
		progress.update(sweepDone, 10, "Projection parameter sweeps")
	}
	for _, k := range []int{3, 5, 10, 20} {
		fp := BuildProjection(tokens, edges, 0, c.MinReliability, k, "full")
		ap := BuildProjection(tokens, edges, 0, c.MinReliability, k, "future-ablated")
		a.Out.Sweeps = append(a.Out.Sweeps, SweepRow{"knn", float64(k), meanGain(selected, prof, fp), meanGain(selected, prof, ap)})
		sweepDone++
		progress.update(sweepDone, 10, "Projection parameter sweeps")
	}
	familyProjection := buildFamilyProjection(tokens, allFams)
	for i, x := range selected {
		for d := 0; d < c.MaxDistance; d++ {
			j, _, _ := metricsFloat(ProjectDistribution(prof[x.A].Right[d], familyProjection), ProjectDistribution(prof[x.B].Right[d], familyProjection))
			a.Out.Pairs[i].Right[d].ProjectedJSFamily = j
			a.Out.Pairs[i].Right[d].GainFamily = j - a.Out.Pairs[i].Right[d].TokenJS
			j, _, _ = metricsFloat(ProjectDistribution(prof[x.A].Left[d], familyProjection), ProjectDistribution(prof[x.B].Left[d], familyProjection))
			a.Out.Pairs[i].Left[d].ProjectedJSFamily = j
			a.Out.Pairs[i].Left[d].GainFamily = j - a.Out.Pairs[i].Left[d].TokenJS
		}
	}
	a.Out.Sweeps = append(a.Out.Sweeps, SweepRow{"family_control", 0, meanGain(selected, prof, familyProjection), meanGain(selected, prof, familyProjection)})
	sweepDone++
	progress.update(sweepDone, 10, "Projection parameter sweeps")
	progress.begin(6, "Family, shuffle, and transition controls")
	controlTotal := len(fams) + 3
	controlDone := 0
	for _, f := range fams {
		a.Families = append(a.Families, familyAnalysis(f, prof, full, future, corp.Counts, c))
		controlDone++
		progress.update(controlDone, controlTotal, "Family, shuffle, and transition controls")
	}
	for _, mode := range []string{"global", "line-preserving"} {
		sc := shuffledCorpus(corp, mode, c.Seed+424242)
		sp := buildProfiles(sc, c.MaxDistance, mode == "line-preserving")
		tj, pj := 0.0, 0.0
		n := 0
		for _, x := range selected {
			for d := 0; d < min(5, c.MaxDistance); d++ {
				t, _, _ := metricsFloat(countsFloat(sp[x.A].Right[d]), countsFloat(sp[x.B].Right[d]))
				q, _, _ := metricsFloat(ProjectDistribution(sp[x.A].Right[d], full), ProjectDistribution(sp[x.B].Right[d], full))
				tj += t
				pj += q
				n++
			}
		}
		if n > 0 {
			tj /= float64(n)
			pj /= float64(n)
		}
		a.Out.Shuffles = append(a.Out.Shuffles, ShuffleResult{mode, tj, pj, pj - tj})
		controlDone++
		progress.update(controlDone, controlTotal, "Family, shuffle, and transition controls")
	}
	a.Out.Transitions = transitions(corp, full, 50)
	controlDone++
	progress.update(controlDone, controlTotal, "Family, shuffle, and transition controls")
	sort.SliceStable(a.Out.Pairs, func(i, j int) bool {
		value := func(p PairResult) float64 {
			if c.ProjectionMode == "ablated" {
				return p.Summary.MeanAblated1To5 - p.Summary.MeanToken1To5
			}
			return p.Summary.Gain1To5
		}
		return value(a.Out.Pairs[i]) > value(a.Out.Pairs[j])
	})
	return a, nil
}

func compare(x pair, p profiles, full, future, past Projection, old previousPair, c Config) PairResult {
	r := PairResult{TokenA: x.A, TokenB: x.B}
	one := func(right bool) []Metric {
		out := make([]Metric, c.MaxDistance)
		for d := 0; d < c.MaxDistance; d++ {
			a, b := p[x.A].Right[d], p[x.B].Right[d]
			abl := future
			if !right {
				a, b = p[x.A].Left[d], p[x.B].Left[d]
				abl = past
			}
			tj, to, tjacc := metricsFloat(countsFloat(a), countsFloat(b))
			obsA, obsB := total(a), total(b)
			rel := math.Min(1, float64(min(obsA, obsB))/float64(c.MinObservations))
			var om previousMetric
			if right && d < len(old.Right) {
				om = old.Right[d]
			}
			if !right && d < len(old.Left) {
				om = old.Left[d]
			}
			if om.Distance == d+1 {
				tj, to, tjacc, obsA, obsB, rel = om.JS, om.Overlap, om.Jaccard, om.ObservationsA, om.ObservationsB, om.Reliability
			}
			fj, fo, fjac := metricsFloat(ProjectDistribution(a, full), ProjectDistribution(b, full))
			aj, ao, ajac := metricsFloat(ProjectDistribution(a, abl), ProjectDistribution(b, abl))
			out[d] = Metric{Distance: d + 1, TokenJS: tj, ProjectedJSFull: fj, ProjectedJSAblated: aj, GainFull: fj - tj, GainAblated: aj - tj, TokenWeightedOverlap: to, ProjectedWeightedOverlapFull: fo, ProjectedWeightedOverlapAblated: ao, TokenJaccard: tjacc, ProjectedJaccardFull: fjac, ProjectedJaccardAblated: ajac, ObservationsA: obsA, ObservationsB: obsB, Reliability: rel}
		}
		return out
	}
	r.Right = one(true)
	r.Left = one(false)
	r.Summary = PairSummary{MeanToken1To5: metricMean(r.Right, 0, 5, "token"), MeanFull1To5: metricMean(r.Right, 0, 5, "full"), MeanAblated1To5: metricMean(r.Right, 0, 5, "ablated"), Gain1To5: metricMean(r.Right, 0, 5, "gain"), Gain6To10: metricMean(r.Right, 5, 10, "gain"), Gain11To20: metricMean(r.Right, 10, 20, "gain")}
	return r
}

func metricMean(x []Metric, lo, hi int, kind string) float64 {
	if hi > len(x) {
		hi = len(x)
	}
	if lo >= hi {
		return 0
	}
	v := []float64{}
	for _, m := range x[lo:hi] {
		switch kind {
		case "token":
			v = append(v, m.TokenJS)
		case "full":
			v = append(v, m.ProjectedJSFull)
		case "ablated":
			v = append(v, m.ProjectedJSAblated)
		default:
			v = append(v, m.GainFull)
		}
	}
	return mean(v)
}
func meanGain(xs []pair, p profiles, proj Projection) float64 {
	v := []float64{}
	for _, x := range xs {
		for d := 0; d < min(5, len(p[x.A].Right)); d++ {
			v = append(v, gain(p[x.A].Right[d], p[x.B].Right[d], proj))
		}
	}
	return mean(v)
}
func mean(x []float64) float64 {
	if len(x) == 0 {
		return 0
	}
	s := 0.0
	for _, v := range x {
		s += v
	}
	return s / float64(len(x))
}
func quantile(x []float64, q float64) float64 {
	if len(x) == 0 {
		return 0
	}
	i := int(math.Ceil(q*float64(len(x)))) - 1
	if i < 0 {
		i = 0
	}
	if i >= len(x) {
		i = len(x) - 1
	}
	return x[i]
}
func percentile(sorted []float64, v float64) float64 {
	n := sort.Search(len(sorted), func(i int) bool { return sorted[i] > v })
	if len(sorted) == 0 {
		return 0
	}
	return 100 * float64(n) / float64(len(sorted))
}
