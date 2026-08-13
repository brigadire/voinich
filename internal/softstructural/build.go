package softstructural

import (
	"fmt"
	"math"
	"sort"

	"zcore.dev/voinich/internal/profilestability"
	"zcore.dev/voinich/internal/structuralreliability"
)

func ValidateConfig(c Config) error {
	if c.MinTokenCount < 1 {
		return fmt.Errorf("min-token-count must be at least 1")
	}
	if c.Neighbors < 1 {
		return fmt.Errorf("neighbors must be at least 1")
	}
	if c.MinEvidenceStrength < 0 || c.MinEvidenceStrength > 1 {
		return fmt.Errorf("min-evidence-strength must be in [0,1]")
	}
	if c.GraphMinSimilarity < 0 || c.GraphMinSimilarity > 1 {
		return fmt.Errorf("graph-min-similarity must be in [0,1]")
	}
	return nil
}

func PairReliabilities(curves structuralreliability.ReliabilityCurves, countA, countB int) (float64, float64, float64) {
	return structuralreliability.PairReliability(structuralreliability.NewReliabilityTable(curves.Position), countA, countB), structuralreliability.PairReliability(structuralreliability.NewReliabilityTable(curves.LeftContext), countA, countB), structuralreliability.PairReliability(structuralreliability.NewReliabilityTable(curves.RightContext), countA, countB)
}
func EvidenceStrength(position, left, right float64) float64 { return (position + left + right) / 3 }
func DiagnosticWeightedMean(similarities, reliabilities [3]float64) *float64 {
	denominator := reliabilities[0] + reliabilities[1] + reliabilities[2]
	if denominator == 0 {
		return nil
	}
	value := (similarities[0]*reliabilities[0] + similarities[1]*reliabilities[1] + similarities[2]*reliabilities[2]) / denominator
	return &value
}

func MakePair(a, b string, countA, countB int, left, right profilestability.Profile, curves structuralreliability.ReliabilityCurves) Pair {
	if b < a {
		a, b = b, a
		countA, countB = countB, countA
		left, right = right, left
	}
	c := profilestability.Compare(left, right)
	rp, rl, rr := PairReliabilities(curves, countA, countB)
	return Pair{TokenA: a, TokenB: b, CountA: countA, CountB: countB, PositionSimilarity: c.PositionSimilarity, LeftSimilarity: c.LeftSimilarity, RightSimilarity: c.RightSimilarity, RawSimilarity: c.Similarity, PositionReliability: rp, LeftReliability: rl, RightReliability: rr, TotalEvidenceWeight: rp + rl + rr, EvidenceStrength: EvidenceStrength(rp, rl, rr), DiagnosticWeightedSimilarity: DiagnosticWeightedMean([3]float64{c.PositionSimilarity, c.LeftSimilarity, c.RightSimilarity}, [3]float64{rp, rl, rr})}
}

// BuildAll is the file-oriented entry point used by the command.
func BuildAll(config Config) (Output, []Pair, error) {
	d, r, err := loadInputs(config)
	if err != nil {
		return Output{}, nil, err
	}
	eligible := make([]string, 0)
	for t, c := range d.counts {
		if c >= config.MinTokenCount {
			eligible = append(eligible, t)
		}
	}
	sort.Strings(eligible)
	bootstrap := map[[2]string]*float64{}
	refs := make([][2]string, 0)
	for _, x := range r.ReferencePairs {
		a, b := x.TokenA, x.TokenB
		if b < a {
			a, b = b, a
		}
		k := [2]string{a, b}
		if _, ok := bootstrap[k]; !ok {
			refs = append(refs, k)
		}
		bootstrap[k] = x.Bootstrap
	}
	sort.Slice(refs, func(i, j int) bool {
		if refs[i][0] == refs[j][0] {
			return refs[i][1] < refs[j][1]
		}
		return refs[i][0] < refs[j][0]
	})
	pairs := make([]Pair, 0, len(eligible)*(len(eligible)-1)/2)
	by := map[string][]Pair{}
	for i, a := range eligible {
		for _, b := range eligible[i+1:] {
			p := MakePair(a, b, d.counts[a], d.counts[b], d.profiles[a], d.profiles[b], r.Curves)
			p.BootstrapProbabilityAbove070 = bootstrap[[2]string{a, b}]
			pairs = append(pairs, p)
			by[a] = append(by[a], p)
			by[b] = append(by[b], p)
		}
	}
	return assembleOutput(config, d, r, eligible, pairs, by, refs), pairs, nil
}

func assembleOutput(c Config, d dataset, r reliabilityInput, eligible []string, pairs []Pair, by map[string][]Pair, refs [][2]string) Output {
	var out Output
	out.Parameters.MinTokenCount = c.MinTokenCount
	out.Parameters.Neighbors = c.Neighbors
	out.Parameters.MinEvidenceStrength = c.MinEvidenceStrength
	out.Parameters.GraphMinSimilarity = c.GraphMinSimilarity
	out.Parameters.PairsFile = c.PairsPath
	out.Methodology.Similarity = "unchanged profilestability.Compare: position=1-JSD, left/right=cosine, raw_combined=arithmetic mean of the three components"
	out.Methodology.Reliability = "component pair reliability is sqrt(R_component(count_a)*R_component(count_b)); lookup uses existing log2 interpolation and endpoint clamping"
	out.Methodology.EvidenceStrength = "(position_reliability+left_reliability+right_reliability)/3; independent of similarity"
	out.Methodology.Diagnostic = "sum(component_similarity*component_reliability)/sum(component_reliability), null when the sum is zero; diagnostic only, not a replacement similarity or clustering score"
	out.Methodology.Graph = "edges filtered by graph_min_similarity only to limit presentation size; no inference or clustering"
	out.EligibleTokens = len(eligible)
	out.PairCount = len(pairs)
	raw := make([]float64, 0, len(pairs))
	evidence := make([]float64, 0, len(pairs))
	for _, p := range pairs {
		raw = append(raw, p.RawSimilarity)
		evidence = append(evidence, p.EvidenceStrength)
		if p.RawSimilarity >= .7 {
			out.JointDiagnostics.RawGE070++
			if p.EvidenceStrength >= .5 {
				out.JointDiagnostics.RawGE070EvidenceGE050++
			}
			if p.EvidenceStrength >= .7 {
				out.JointDiagnostics.RawGE070EvidenceGE070++
			}
			if p.EvidenceStrength >= .9 {
				out.JointDiagnostics.RawGE070EvidenceGE090++
			}
		}
		if p.RawSimilarity >= c.GraphMinSimilarity {
			out.GraphEdges = append(out.GraphEdges, GraphEdge{p.TokenA, p.TokenB, p.RawSimilarity, p.EvidenceStrength})
		}
	}
	out.RawSimilarityDistribution = distribution(raw, true)
	out.EvidenceStrengthDistribution = distribution(evidence, false)
	out.DiagnosticBuckets = buckets(pairs)
	rawTop := map[string]map[string]bool{}
	supportedTop := map[string]map[string]bool{}
	for _, token := range eligible {
		items := by[token]
		rawRank := rank(items, token, "raw", 0)
		supportedRank := rank(items, token, "supported", 0)
		highRank := rank(items, token, "raw", c.MinEvidenceStrength)
		n := TokenNeighborhood{Token: token, Count: d.counts[token], TopRawNeighbors: limit(rawRank, c.Neighbors), TopSupportedNeighbors: limit(supportedRank, c.Neighbors), TopHighEvidenceNeighbors: limit(highRank, c.Neighbors)}
		out.Neighborhoods = append(out.Neighborhoods, n)
		rawTop[token] = neighborSet(n.TopRawNeighbors)
		supportedTop[token] = neighborSet(n.TopSupportedNeighbors)
	}
	out.MutualRaw = mutual(pairs, rawTop)
	out.MutualSupported = mutual(pairs, supportedTop)
	for _, key := range refs {
		a, b := key[0], key[1]
		pa, oka := d.profiles[a]
		pb, okb := d.profiles[b]
		if !oka || !okb {
			continue
		}
		p := MakePair(a, b, d.counts[a], d.counts[b], pa, pb, r.Curves)
		p.BootstrapProbabilityAbove070 = findBootstrap(r, key)
		out.ReferencePairs = append(out.ReferencePairs, reference(p))
	}
	return out
}

func distribution(values []float64, raw bool) Distribution {
	if len(values) == 0 {
		return Distribution{}
	}
	sort.Float64s(values)
	d := Distribution{Mean: mean(values), Median: percentile(values, .5), P90: percentile(values, .9)}
	if raw {
		d.P95 = percentile(values, .95)
		d.P99 = percentile(values, .99)
		d.Max = values[len(values)-1]
	} else {
		d.P10 = percentile(values, .1)
		d.P25 = percentile(values, .25)
		d.P50 = percentile(values, .5)
		d.P75 = percentile(values, .75)
	}
	return d
}
func mean(v []float64) float64 {
	s := 0.
	for _, x := range v {
		s += x
	}
	if len(v) == 0 {
		return 0
	}
	return s / float64(len(v))
}
func percentile(v []float64, p float64) float64 {
	if len(v) == 0 {
		return 0
	}
	x := p * float64(len(v)-1)
	lo := int(math.Floor(x))
	hi := int(math.Ceil(x))
	if lo == hi {
		return v[lo]
	}
	return v[lo] + (v[hi]-v[lo])*(x-float64(lo))
}
func buckets(pairs []Pair) []BucketRow {
	labels := []string{"<0.5", "0.5-0.6", "0.6-0.7", "0.7-0.8", "0.8-0.9", ">=0.9"}
	rows := make([]BucketRow, 6)
	for i := range rows {
		rows[i].RawSimilarityBin = labels[i]
	}
	for _, p := range pairs {
		i := 5
		if p.RawSimilarity < .5 {
			i = 0
		} else if p.RawSimilarity < .6 {
			i = 1
		} else if p.RawSimilarity < .7 {
			i = 2
		} else if p.RawSimilarity < .8 {
			i = 3
		} else if p.RawSimilarity < .9 {
			i = 4
		}
		switch {
		case p.EvidenceStrength < .3:
			rows[i].EvidenceLT030++
		case p.EvidenceStrength < .5:
			rows[i].Evidence030To050++
		case p.EvidenceStrength < .7:
			rows[i].Evidence050To070++
		case p.EvidenceStrength < .9:
			rows[i].Evidence070To090++
		default:
			rows[i].EvidenceGE090++
		}
	}
	return rows
}
func other(p Pair, t string) string {
	if p.TokenA == t {
		return p.TokenB
	}
	return p.TokenA
}
func rank(items []Pair, token, kind string, minEvidence float64) []Neighbor {
	copyItems := append([]Pair(nil), items...)
	filtered := copyItems[:0]
	for _, p := range copyItems {
		if p.EvidenceStrength >= minEvidence {
			filtered = append(filtered, p)
		}
	}
	sort.Slice(filtered, func(i, j int) bool {
		var a, b float64
		if kind == "supported" {
			if filtered[i].DiagnosticWeightedSimilarity != nil {
				a = *filtered[i].DiagnosticWeightedSimilarity
			}
			if filtered[j].DiagnosticWeightedSimilarity != nil {
				b = *filtered[j].DiagnosticWeightedSimilarity
			}
		} else {
			a, b = filtered[i].RawSimilarity, filtered[j].RawSimilarity
		}
		if a != b {
			return a > b
		}
		return other(filtered[i], token) < other(filtered[j], token)
	})
	result := make([]Neighbor, 0, len(filtered))
	for _, p := range filtered {
		result = append(result, Neighbor{other(p, token), p.RawSimilarity, p.EvidenceStrength, p.DiagnosticWeightedSimilarity, p.BootstrapProbabilityAbove070})
	}
	return result
}
func limit[T any](x []T, n int) []T {
	if len(x) > n {
		return x[:n]
	}
	return x
}
func neighborSet(x []Neighbor) map[string]bool {
	m := map[string]bool{}
	for _, n := range x {
		m[n.Token] = true
	}
	return m
}
func mutual(pairs []Pair, sets map[string]map[string]bool) []MutualPair {
	r := []MutualPair{}
	for _, p := range pairs {
		if sets[p.TokenA][p.TokenB] && sets[p.TokenB][p.TokenA] {
			r = append(r, MutualPair{p.TokenA, p.TokenB, p.RawSimilarity, p.EvidenceStrength, p.DiagnosticWeightedSimilarity})
		}
	}
	return r
}
func reference(p Pair) ReferencePair {
	return ReferencePair{TokenA: p.TokenA, TokenB: p.TokenB, CountA: p.CountA, CountB: p.CountB, Position: ComponentEvidence{p.PositionSimilarity, p.PositionReliability}, LeftContext: ComponentEvidence{p.LeftSimilarity, p.LeftReliability}, RightContext: ComponentEvidence{p.RightSimilarity, p.RightReliability}, RawStructuralSimilarity: p.RawSimilarity, TotalEvidenceWeight: p.TotalEvidenceWeight, EvidenceStrength: p.EvidenceStrength, DiagnosticWeightedSimilarity: p.DiagnosticWeightedSimilarity, BootstrapProbabilityAbove070: p.BootstrapProbabilityAbove070}
}
func findBootstrap(r reliabilityInput, key [2]string) *float64 {
	for _, x := range r.ReferencePairs {
		a, b := x.TokenA, x.TokenB
		if b < a {
			a, b = b, a
		}
		if a == key[0] && b == key[1] {
			return x.Bootstrap
		}
	}
	return nil
}
