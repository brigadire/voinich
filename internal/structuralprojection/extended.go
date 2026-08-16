package structuralprojection

import (
	"math"
	"sort"
	"strings"
)

func sequenceResults(x pair, p profiles, full, abl Projection, minObs int) []SequenceResult {
	out := []SequenceResult{}
	for idx, n := range []int{2, 3} {
		a, b := p[x.A].Suffix[idx], p[x.B].Suffix[idx]
		exact, _, _ := metricsFloat(countsFloat(a), countsFloat(b))
		pf, pa := []float64{}, []float64{}
		for pos := 0; pos < n; pos++ {
			ma := suffixMarginal(a, pos)
			mb := suffixMarginal(b, pos)
			j, _, _ := metricsFloat(ProjectDistribution(ma, full), ProjectDistribution(mb, full))
			q, _, _ := metricsFloat(ProjectDistribution(ma, abl), ProjectDistribution(mb, abl))
			pf = append(pf, j)
			pa = append(pa, q)
		}
		out = append(out, SequenceResult{TokenA: x.A, TokenB: x.B, Length: n, ExactSimilarity: exact, ProjectedSimilarityFull: mean(pf), ProjectedSimilarityAblated: mean(pa), PositionFull: pf, PositionAblated: pa})
	}
	return out
}
func suffixMarginal(x map[string]int, pos int) map[string]int {
	out := map[string]int{}
	for s, n := range x {
		z := strings.Split(s, "\x1f")
		if pos < len(z) {
			out[z[pos]] += n
		}
	}
	return out
}

func buildFamilyProjection(tokens []string, fams []family) Projection {
	member := map[string]string{}
	for _, f := range fams {
		for _, t := range f.Tokens {
			member[t] = "family_" + itoa(f.ID)
		}
	}
	out := Projection{}
	for _, t := range tokens {
		label := member[t]
		if label == "" {
			label = "singleton:" + t
		}
		out[t] = map[string]float64{label: 1}
	}
	return out
}
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	b := make([]byte, 0, 10)
	for n > 0 {
		b = append(b, byte('0'+n%10))
		n /= 10
	}
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return string(b)
}

func familyAnalysis(f family, p profiles, full, abl Projection, counts map[string]int, c Config) FamilyResult {
	r := FamilyResult{ID: f.ID, Tokens: f.Tokens}
	candidates := make([]string, 0, len(p))
	for t := range p {
		if counts[t] >= c.MinObservations {
			candidates = append(candidates, t)
		}
	}
	sort.Strings(candidates)
	// matchedGroup(f.Tokens, candidates, counts, trial) depends only on the
	// family's own tokens, the (fixed, already-sorted) candidate pool, and
	// the trial index — never on distance d — so precompute all 200 trial
	// groups once instead of recomputing them on every one of the
	// c.MaxDistance distances (a c.MaxDistance-fold redundant recompute).
	matchedGroups := make([][]string, 200)
	for trial := range matchedGroups {
		matchedGroups[trial] = matchedGroup(f.Tokens, candidates, counts, trial)
	}
	for d := 0; d < c.MaxDistance; d++ {
		// A matched token is reused across many trials. Cache its projected
		// distribution for this distance, while discarding the cache before the
		// next distance to keep memory bounded.
		fullCache, ablCache := map[string]map[string]float64{}, map[string]map[string]float64{}
		get := func(t string, proj Projection, cache map[string]map[string]float64) map[string]float64 {
			if x := cache[t]; x != nil {
				return x
			}
			x := ProjectDistribution(p[t].Right[d], proj)
			cache[t] = x
			return x
		}
		coh := func(group []string, mode string) (meanV, disp float64, medoid string) {
			if len(group) == 0 {
				return
			}
			sums := make([]float64, len(group))
			vals := []float64{}
			for i := range group {
				for j := 0; j < i; j++ {
					var a, b map[string]float64
					switch mode {
					case "token":
						a, b = countsFloat(p[group[i]].Right[d]), countsFloat(p[group[j]].Right[d])
					case "full":
						a, b = get(group[i], full, fullCache), get(group[j], full, fullCache)
					default:
						a, b = get(group[i], abl, ablCache), get(group[j], abl, ablCache)
					}
					v, _, _ := metricsFloat(a, b)
					vals = append(vals, v)
					sums[i] += v
					sums[j] += v
				}
			}
			meanV = mean(vals)
			for _, v := range vals {
				disp += (v - meanV) * (v - meanV)
			}
			if len(vals) > 0 {
				disp = math.Sqrt(disp / float64(len(vals)))
			}
			best := 0
			for i := 1; i < len(sums); i++ {
				if sums[i] > sums[best] {
					best = i
				}
			}
			medoid = group[best]
			return
		}
		tc, td, tm := coh(f.Tokens, "token")
		fc, fd, fm := coh(f.Tokens, "full")
		ac, ad, am := coh(f.Tokens, "ablated")
		tv, fv, av := []float64{}, []float64{}, []float64{}
		for trial := 0; trial < 200; trial++ {
			g := matchedGroups[trial]
			if len(g) != len(f.Tokens) {
				continue
			}
			x, _, _ := coh(g, "token")
			y, _, _ := coh(g, "full")
			z, _, _ := coh(g, "ablated")
			tv = append(tv, x)
			fv = append(fv, y)
			av = append(av, z)
		}
		sort.Float64s(tv)
		sort.Float64s(fv)
		sort.Float64s(av)
		r.Distances = append(r.Distances, FamilyDistance{Distance: d + 1, TokenCohesion: tc, ProjectedCohesionFull: fc, ProjectedCohesionAblated: ac, TokenDispersion: td, ProjectedDispersionFull: fd, ProjectedDispersionAblated: ad, TokenMedoid: tm, FullMedoid: fm, AblatedMedoid: am, MatchedPercentileToken: percentile(tv, tc), MatchedPercentileFull: percentile(fv, fc), MatchedPercentileAblated: percentile(av, ac)})
	}
	return r
}
func cohesion(group []string, p profiles, d int, proj Projection) (meanV, disp float64, medoid string) {
	if len(group) == 0 {
		return
	}
	sums := make([]float64, len(group))
	vals := []float64{}
	for i := range group {
		for j := 0; j < i; j++ {
			var v float64
			if proj == nil {
				v, _, _ = metricsFloat(countsFloat(p[group[i]].Right[d]), countsFloat(p[group[j]].Right[d]))
			} else {
				v, _, _ = metricsFloat(ProjectDistribution(p[group[i]].Right[d], proj), ProjectDistribution(p[group[j]].Right[d], proj))
			}
			vals = append(vals, v)
			sums[i] += v
			sums[j] += v
		}
	}
	meanV = mean(vals)
	for _, v := range vals {
		disp += (v - meanV) * (v - meanV)
	}
	if len(vals) > 0 {
		disp = math.Sqrt(disp / float64(len(vals)))
	}
	best := 0
	for i := 1; i < len(sums); i++ {
		if sums[i] > sums[best] {
			best = i
		}
	}
	medoid = group[best]
	return
}
func matchedGroup(target, candidates []string, counts map[string]int, trial int) []string {
	out := []string{}
	used := map[string]bool{}
	for i, t := range target {
		eligible := []string{}
		for _, x := range candidates {
			if !used[x] && max(counts[x], counts[t]) <= 2*min(counts[x], counts[t]) {
				eligible = append(eligible, x)
			}
		}
		if len(eligible) == 0 {
			return nil
		}
		x := eligible[(trial*37+i*17)%len(eligible)]
		used[x] = true
		out = append(out, x)
	}
	return out
}

func transitions(c corpus, p Projection, limit int) []Transition {
	joint := map[string]float64{}
	source, dest := map[string]float64{}, map[string]float64{}
	total := 0.0
	for i := 0; i+1 < len(c.Tokens); i++ {
		a, b := p[c.Tokens[i]], p[c.Tokens[i+1]]
		if len(a) == 0 {
			a = map[string]float64{c.Tokens[i]: 1}
		}
		if len(b) == 0 {
			b = map[string]float64{c.Tokens[i+1]: 1}
		}
		for x, wx := range a {
			source[x] += wx
			for y, wy := range b {
				joint[x+"\x1f"+y] += wx * wy
			}
		}
		for y, wy := range b {
			dest[y] += wy
		}
		total++
	}
	out := []Transition{}
	if total == 0 {
		return out
	}
	for key, n := range joint {
		x := strings.SplitN(key, "\x1f", 2)
		obs := n / total
		base := (source[x[0]] / total) * (dest[x[1]] / total)
		lift := 0.0
		if base > 0 {
			lift = obs / base
		}
		if n >= 5 {
			out = append(out, Transition{x[0], x[1], obs, base, lift})
		}
	}
	// out is built by ranging over the joint map (extended.go above), so its
	// initial order is randomized independently of the seed/input; when two
	// transitions have exactly equal Lift and Observed, the unstable sort
	// below previously had no further tie-breaker, so which one sorted
	// first (and thus which survived a boundary at limit) depended on that
	// randomized initial order rather than being deterministic (see
	// determinism_test.go). Source/Destination lexicographic order gives a
	// fully-specified, deterministic resolution for genuine ties.
	sort.Slice(out, func(i, j int) bool {
		if out[i].Lift != out[j].Lift {
			return out[i].Lift > out[j].Lift
		}
		if out[i].Observed != out[j].Observed {
			return out[i].Observed > out[j].Observed
		}
		if out[i].Source != out[j].Source {
			return out[i].Source < out[j].Source
		}
		return out[i].Destination < out[j].Destination
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}
