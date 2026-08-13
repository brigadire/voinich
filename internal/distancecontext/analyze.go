package distancecontext

import (
	"fmt"
	"sort"
	"strings"
)

var requiredPairs = []pair{{"or", "s"}, {"r", "s"}, {"or", "r"}, {"chol", "daiin"}, {"chor", "daiin"}, {"daiin", "dol"}, {"chol", "cthy"}, {"qokain", "qol"}, {"qokaiin", "qol"}, {"qokedy", "qol"}, {"lchedy", "qokain"}, {"lchedy", "qokeey"}}

type analysis struct {
	Out       Output
	Controls  []ControlResult
	Sequences []SequencePair
	Families  []FamilyResult
}

func validate(c Config) error {
	if c.MaxDistance < 1 {
		return fmt.Errorf("max-distance must be positive")
	}
	if c.MinObservations < 1 {
		return fmt.Errorf("min-observations must be positive")
	}
	if c.TopN < 0 {
		return fmt.Errorf("top must be non-negative")
	}
	if c.FamilyID < 0 {
		return fmt.Errorf("family must be non-negative")
	}
	if c.Pair != "" {
		x := strings.Split(c.Pair, ",")
		if len(x) != 2 || strings.TrimSpace(x[0]) == "" || strings.TrimSpace(x[1]) == "" || strings.TrimSpace(x[0]) == strings.TrimSpace(x[1]) {
			return fmt.Errorf("pair must be distinct tokenA,tokenB")
		}
	}
	return nil
}

func analyze(c Config) (analysis, error) {
	if e := validate(c); e != nil {
		return analysis{}, e
	}
	corp, e := readCorpus(c.CorpusPath)
	if e != nil {
		return analysis{}, e
	}
	fam, e := readFamilies(c.FamiliesPath)
	if e != nil {
		return analysis{}, fmt.Errorf("read families: %w", e)
	}
	var ranked []pair
	if c.TopN > 0 {
		ranked, e = readPairTSV(c.DistantPath, c.TopN)
		if e != nil {
			return analysis{}, fmt.Errorf("read distant pairs: %w", e)
		}
	}
	ctrl, e := readControls(c.ControlsPath)
	if e != nil {
		return analysis{}, fmt.Errorf("read controls: %w", e)
	}
	continuous, bounded := buildProfiles(corp, c.MaxDistance, false), buildProfiles(corp, c.MaxDistance, true)
	base, baseVals := baselines(continuous, corp.Counts, c.MaxDistance, c.MinObservations)
	selected := []pair{}
	seen := map[pair]bool{}
	add := func(x pair) {
		x = canon(strings.TrimSpace(x.A), strings.TrimSpace(x.B))
		if !seen[x] {
			seen[x] = true
			selected = append(selected, x)
		}
	}
	selectedFamilies := fam
	if c.Pair != "" {
		x := strings.Split(c.Pair, ",")
		add(pair{x[0], x[1]})
		selectedFamilies = nil
	} else if c.FamilyID > 0 {
		selectedFamilies = nil
		for _, f := range fam {
			if f.ID == c.FamilyID {
				selectedFamilies = append(selectedFamilies, f)
				for _, x := range f.Edges {
					add(pair{x.A, x.B})
				}
			}
		}
		if len(selectedFamilies) == 0 {
			return analysis{}, fmt.Errorf("family %d not found", c.FamilyID)
		}
	} else {
		for _, x := range requiredPairs {
			add(x)
		}
		for _, x := range ranked {
			add(x)
		}
		for _, f := range fam {
			for _, x := range f.Edges {
				add(pair{x.A, x.B})
			}
		}
	}
	for _, x := range selected {
		if corp.Counts[x.A] == 0 || corp.Counts[x.B] == 0 {
			return analysis{}, fmt.Errorf("selected pair %s/%s contains token absent from corpus", x.A, x.B)
		}
	}
	a := analysis{Out: Output{Parameters: map[string]any{"corpus": c.CorpusPath, "max_distance": c.MaxDistance, "min_observations": c.MinObservations, "primary_mode": map[bool]string{false: "continuous", true: "line_bounded"}[c.RespectLineBoundaries], "frequency_matching": "unordered token counts within a factor of 2", "js_similarity": "1 - Jensen-Shannon divergence / ln(2)", "weighted_overlap": "sum_x min(P_A(x),P_B(x))", "effective_support": "exp(Shannon entropy in natural-log units)", "reliability": "min(1, min(observations_A, observations_B) / min_observations)"}, TokenCount: len(corp.Tokens), PairCount: len(selected), Baseline: base}}
	for _, x := range selected {
		pr := comparePair(x, corp.Counts, continuous, bounded, baseVals, c.MinObservations, c.MaxDistance)
		a.Out.Pairs = append(a.Out.Pairs, pr)
		a.Sequences = append(a.Sequences, sequencePair(x, continuous, bounded, c.MinObservations))
	}
	for _, x := range ctrl {
		if !seen[canon(x.TargetA, x.TargetB)] {
			continue
		}
		if continuous[x.A] == nil || continuous[x.B] == nil {
			continue
		}
		a.Controls = append(a.Controls, ControlResult{x.TargetA, x.TargetB, x.Rank, x.A, x.B, profileDirection(continuous[x.A].Right, continuous[x.B].Right, nil, c.MinObservations)})
	}
	for _, f := range selectedFamilies {
		a.Families = append(a.Families, familyResult(f, continuous, corp.Counts, c.MinObservations, c.MaxDistance))
	}
	sort.SliceStable(a.Out.Pairs, func(i, j int) bool {
		return a.Out.Pairs[i].RightSummary.Persistence1To5 > a.Out.Pairs[j].RightSummary.Persistence1To5
	})
	return a, nil
}

func baselines(p profiles, counts map[string]int, maxD, minObs int) ([]BaselineRow, [][]float64) {
	tokens := make([]string, 0, len(p))
	for t := range p {
		if counts[t] >= minObs {
			tokens = append(tokens, t)
		}
	}
	sort.Strings(tokens)
	vals := make([][]float64, maxD)
	for i, a := range tokens {
		for _, b := range tokens[i+1:] {
			ca, cb := counts[a], counts[b]
			if max(ca, cb) > 2*min(ca, cb) {
				continue
			}
			for d := 0; d < maxD; d++ {
				if total(p[a].Right[d]) < minObs || total(p[b].Right[d]) < minObs {
					continue
				}
				v, _, _ := rawMetrics(p[a].Right[d], p[b].Right[d])
				vals[d] = append(vals[d], v)
			}
		}
	}
	out := make([]BaselineRow, maxD)
	for d := range vals {
		sort.Float64s(vals[d])
		out[d] = BaselineRow{d + 1, len(vals[d]), quantile(vals[d], .5), quantile(vals[d], .9), quantile(vals[d], .95)}
	}
	return out, vals
}
func profileDirection(a, b []map[string]int, baseline [][]float64, minObs int) []Metric {
	n := min(len(a), len(b))
	out := make([]Metric, n)
	for d := 0; d < n; d++ {
		out[d] = metric(d+1, a[d], b[d], minObs)
		if baseline != nil {
			out[d].BaselinePercentile = percentile(baseline[d], out[d].JSSimilarity)
		}
	}
	return out
}
func comparePair(x pair, counts map[string]int, p, b profiles, base [][]float64, minObs, maxD int) PairResult {
	r := PairResult{TokenA: x.A, TokenB: x.B, CountA: counts[x.A], CountB: counts[x.B]}
	r.Right = profileDirection(p[x.A].Right, p[x.B].Right, base, minObs)
	r.Left = profileDirection(p[x.A].Left, p[x.B].Left, nil, minObs)
	r.LineBoundedRight = profileDirection(b[x.A].Right, b[x.B].Right, nil, minObs)
	r.LineBoundedLeft = profileDirection(b[x.A].Left, b[x.B].Left, nil, minObs)
	for d := 0; d < maxD; d++ {
		r.BoundarySensitivity = append(r.BoundarySensitivity, BoundaryMetric{d + 1, r.Right[d].JSSimilarity, r.LineBoundedRight[d].JSSimilarity, r.Right[d].JSSimilarity - r.LineBoundedRight[d].JSSimilarity})
	}
	r.RightSummary = summary(r.Right)
	r.LeftSummary = summary(r.Left)
	return r
}
func sequencePair(x pair, p, b profiles, minObs int) SequencePair {
	one := func(q profiles) []SequenceMetric {
		out := make([]SequenceMetric, 2)
		for i := 0; i < 2; i++ {
			m := metric(i+2, q[x.A].Suffix[i], q[x.B].Suffix[i], minObs)
			out[i] = SequenceMetric{m.Distance, m.JSSimilarity, m.WeightedOverlap, m.Jaccard, m.ObservationsA, m.ObservationsB, m.EffectiveSupportA, m.EffectiveSupportB, m.Reliability, m.Reliable}
		}
		return out
	}
	return SequencePair{x.A, x.B, one(p), one(b)}
}

func familyResult(f familyInput, p profiles, counts map[string]int, minObs, maxD int) FamilyResult {
	r := FamilyResult{ID: f.ID, Tokens: f.Tokens}
	all := make([]string, 0, len(p))
	for t := range p {
		if counts[t] >= minObs {
			all = append(all, t)
		}
	}
	sort.Strings(all)
	for d := 0; d < maxD; d++ {
		n := len(f.Tokens)
		mat := make([][]float64, n)
		sums := make([]float64, n)
		sumv, np := 0., 0
		for i := range mat {
			mat[i] = make([]float64, n)
			mat[i][i] = 1
			for j := 0; j < i; j++ {
				v, _, _ := rawMetrics(p[f.Tokens[i]].Right[d], p[f.Tokens[j]].Right[d])
				mat[i][j], mat[j][i] = v, v
				sums[i] += v
				sums[j] += v
				sumv += v
				np++
			}
		}
		med := 0
		for i := 1; i < n; i++ {
			if sums[i] > sums[med] {
				med = i
			}
		}
		coh := 0.
		if np > 0 {
			coh = sumv / float64(np)
		}
		random := matchedCohesions(f.Tokens, all, p, counts, d, 200)
		sort.Float64s(random)
		r.Profiles = append(r.Profiles, DistanceMatrix{d + 1, mat, f.Tokens[med], coh, percentile(random, coh)})
	}
	return r
}
func matchedCohesions(target, candidates []string, p profiles, counts map[string]int, d, trials int) []float64 {
	out := make([]float64, 0, trials)
	for trial := 0; trial < trials; trial++ {
		group := []string{}
		used := map[string]bool{}
		ok := true
		for i, t := range target {
			eligible := []string{}
			for _, x := range candidates {
				if !used[x] && max(counts[x], counts[t]) <= 2*min(counts[x], counts[t]) {
					eligible = append(eligible, x)
				}
			}
			if len(eligible) == 0 {
				ok = false
				break
			}
			x := eligible[(trial*37+i*17)%len(eligible)]
			used[x] = true
			group = append(group, x)
		}
		if !ok {
			continue
		}
		s, n := 0., 0
		for i := range group {
			for j := 0; j < i; j++ {
				v, _, _ := rawMetrics(p[group[i]].Right[d], p[group[j]].Right[d])
				s += v
				n++
			}
		}
		if n > 0 {
			out = append(out, s/float64(n))
		}
	}
	return out
}
