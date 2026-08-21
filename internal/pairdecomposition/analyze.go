package pairdecomposition

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
)

func ValidateConfig(c Config) error {
	if c.TopN < 0 {
		return fmt.Errorf("top must be non-negative")
	}
	if c.ContextLimit < 1 {
		return fmt.Errorf("context-limit must be positive")
	}
	if c.Controls < 0 {
		return fmt.Errorf("controls must be non-negative")
	}
	if c.FamilyID < 0 {
		return fmt.Errorf("family must be non-negative")
	}
	if c.Pair != "" {
		x := strings.Split(c.Pair, ",")
		if len(x) != 2 || strings.TrimSpace(x[0]) == "" || strings.TrimSpace(x[1]) == "" || x[0] == x[1] {
			return fmt.Errorf("pair must be tokenA,tokenB with distinct non-empty tokens")
		}
	}
	return nil
}

func key(a, b string) [2]string {
	if b < a {
		a, b = b, a
	}
	return [2]string{a, b}
}

func Run(c Config) (Output, []FamilyResult, error) {
	profiles, err := readDictionary(c.DictionaryPath)
	if err != nil {
		return Output{}, nil, err
	}
	pairs, err := readPairs(c.PairsPath)
	if err != nil {
		return Output{}, nil, fmt.Errorf("read pairs: %w", err)
	}
	families, err := readFamilies(c.FamiliesPath)
	if err != nil {
		return Output{}, nil, fmt.Errorf("read families: %w", err)
	}
	by := map[[2]string]PairSource{}
	for _, p := range pairs {
		by[key(p.TokenA, p.TokenB)] = p
	}
	selected := map[[2]string]bool{}
	var ordered [][2]string
	add := func(k [2]string) {
		if !selected[k] {
			selected[k] = true
			ordered = append(ordered, k)
		}
	}
	if c.Pair != "" {
		x := strings.Split(c.Pair, ",")
		add(key(strings.TrimSpace(x[0]), strings.TrimSpace(x[1])))
	} else if c.FamilyID > 0 {
		found := false
		for _, f := range families {
			if f.ID == c.FamilyID {
				found = true
				for _, e := range f.Edges {
					add(key(e.TokenA, e.TokenB))
				}
			}
		}
		if !found {
			return Output{}, nil, fmt.Errorf("family %d not found", c.FamilyID)
		}
	} else {
		distant, e := readDistant(c.DistantPath)
		if e != nil {
			return Output{}, nil, fmt.Errorf("read distant pairs: %w", e)
		}
		n := c.TopN
		if n == 0 || n > len(distant) {
			n = len(distant)
		}
		for _, k := range distant[:n] {
			add(key(k[0], k[1]))
		}
		for _, f := range families {
			for _, e := range f.Edges {
				add(key(e.TokenA, e.TokenB))
			}
		}
	}
	printCardinalityDiagnostics(len(ordered), families, c.Controls)
	globalLeft, globalRight := globalContext(profiles, "left"), globalContext(profiles, "right")
	out := Output{Methodology: map[string]string{
		"structural_similarity":       "copied unchanged from structural_graphemic_pairs.tsv; arithmetic mean of existing position, left-cosine, and right-cosine components",
		"context_probabilities":       "full observed predecessor/successor count distributions are stored and used for metrics; common/specific/differential display lists alone are truncated",
		"entropy":                     "Shannon entropy in natural-log units; effective vocabulary is exp(entropy)",
		"weighted_context_similarity": "1 - Jensen-Shannon divergence/base ln(2), computed over the full probability distributions",
		"negative_controls":           "minimum deterministic match cost over log-counts, graphemic distance, reliability, and distance from corpus-median structural similarity",
		"shared_absence":              "globally frequent context tokens absent from both distributions, reported only when each side has at least 30 context observations; absence means unobserved, not impossible",
	}}
	// A target can also be selected as a control for another target. Decomposition
	// is immutable for a PairSource, so reuse it without changing ordering or
	// serialization. This is particularly important when one large family adds
	// thousands of target edges.
	decomposed := make(map[[2]string]PairResult, len(ordered))
	decomposePair := func(p PairSource) (PairResult, error) {
		k := key(p.TokenA, p.TokenB)
		if r, ok := decomposed[k]; ok {
			return r, nil
		}
		r, e := decompose(p, profiles, globalLeft, globalRight, c.ContextLimit)
		if e != nil {
			return PairResult{}, e
		}
		decomposed[k] = r
		return r, nil
	}
	for _, k := range ordered {
		p, ok := by[k]
		if !ok {
			return Output{}, nil, fmt.Errorf("pair %s/%s not found in full pair file", k[0], k[1])
		}
		r, e := decomposePair(p)
		if e != nil {
			return Output{}, nil, e
		}
		out.Pairs = append(out.Pairs, r)
	}
	median := medianStructural(pairs)
	controlIndex := newControlIndex(pairs)
	for _, target := range out.Pairs {
		src := by[key(target.TokenA, target.TokenB)]
		controls := controlIndex.choose(src, median, c.Controls)
		for i, x := range controls {
			r, e := decomposePair(x.Pair)
			if e != nil {
				return Output{}, nil, e
			}
			out.Controls = append(out.Controls, Control{target.TokenA, target.TokenB, i + 1, x.Cost, r})
		}
	}
	var familyResults []FamilyResult
	for _, f := range families {
		if c.Pair != "" {
			break
		}
		if c.FamilyID > 0 && f.ID != c.FamilyID {
			continue
		}
		fr, e := decomposeFamily(f, by)
		if e != nil {
			return Output{}, nil, e
		}
		familyResults = append(familyResults, fr)
	}
	return out, familyResults, nil
}

func printCardinalityDiagnostics(targetPairs int, families []FamilyInput, controls int) {
	totalEdges, largestTokens, largestEdges := 0, 0, 0
	for _, f := range families {
		totalEdges += len(f.Edges)
		if len(f.Tokens) > largestTokens {
			largestTokens = len(f.Tokens)
		}
		if len(f.Edges) > largestEdges {
			largestEdges = len(f.Edges)
		}
	}
	estimated := targetPairs * (1 + controls)
	fmt.Fprintf(os.Stderr, "structural cardinality: target_pairs=%d estimated_decompositions=%d family_count=%d largest_family_tokens=%d largest_family_edges=%d total_family_edges=%d\n", targetPairs, estimated, len(families), largestTokens, largestEdges, totalEdges)
	if largestEdges > 1000 || (targetPairs > 0 && largestEdges > targetPairs*10) {
		fmt.Fprintf(os.Stderr, "STRUCTURAL_CARDINALITY_EXPLOSION target_pairs=%d estimated_decompositions=%d largest_family_edges=%d total_family_edges=%d\n", targetPairs, estimated, largestEdges, totalEdges)
	}
}

func decompose(s PairSource, profiles map[string]profile, globalLeft, globalRight map[string]int, limit int) (PairResult, error) {
	a, ok := profiles[s.TokenA]
	if !ok {
		return PairResult{}, fmt.Errorf("token %q missing from dictionary", s.TokenA)
	}
	b, ok := profiles[s.TokenB]
	if !ok {
		return PairResult{}, fmt.Errorf("token %q missing from dictionary", s.TokenB)
	}
	r := PairResult{TokenA: s.TokenA, TokenB: s.TokenB, CountA: s.CountA, CountB: s.CountB, StructuralSimilarity: s.Structural, Reliability: s.Reliability, GraphemicDistance: s.Graphemic, PositionSimilarity: s.Position, LeftSimilarity: s.Left, RightSimilarity: s.Right, PositionReliability: s.PositionReliability, LeftReliability: s.LeftReliability, RightReliability: s.RightReliability}
	r.PositionA = positionSummary(a)
	r.PositionB = positionSummary(b)
	r.PositionDistribution = positionRows(a.Positions, b.Positions)
	r.PositionJSSimilarity = jsInt(a.Positions, b.Positions)
	r.Left = contextProfile(a.Left, b.Left, s.Left, globalLeft, limit)
	r.Right = contextProfile(a.Right, b.Right, s.Right, globalRight, limit)
	r.SharedContextStrength = (r.Left.JensenShannonSimilarity + r.Right.JensenShannonSimilarity) / 2
	r.DifferentialContextStrength = (totalVariation(a.Left, b.Left) + totalVariation(a.Right, b.Right)) / 2
	r.PositionalAgreement = r.PositionJSSimilarity
	le := 1 - math.Abs(r.Left.EntropyA-r.Left.EntropyB)/math.Max(1, math.Max(r.Left.EntropyA, r.Left.EntropyB))
	re := 1 - math.Abs(r.Right.EntropyA-r.Right.EntropyB)/math.Max(1, math.Max(r.Right.EntropyA, r.Right.EntropyB))
	r.EntropyAgreement = (le + re) / 2
	r.Explanation = explain(r)
	return r, nil
}

func sum[K comparable](m map[K]int) int {
	n := 0
	for _, v := range m {
		n += v
	}
	return n
}
func positionSummary(p profile) PositionSummary {
	r := PositionSummary{}
	if p.Count > 0 {
		r.LineStartProbability = float64(p.Start) / float64(p.Count)
		r.LineEndProbability = float64(p.End) / float64(p.Count)
	}
	total := sum(p.Positions)
	if total == 0 {
		return r
	}
	var vals []int
	for pos, n := range p.Positions {
		r.Mean += float64(pos * n)
		for i := 0; i < n; i++ {
			vals = append(vals, pos)
		}
	}
	r.Mean /= float64(total)
	sort.Ints(vals)
	if len(vals)%2 == 1 {
		r.Median = float64(vals[len(vals)/2])
	} else {
		r.Median = float64(vals[len(vals)/2-1]+vals[len(vals)/2]) / 2
	}
	return r
}
func positionRows(a, b map[int]int) []PositionRow {
	keys := map[int]bool{}
	for x := range a {
		keys[x] = true
	}
	for x := range b {
		keys[x] = true
	}
	var ks []int
	for x := range keys {
		ks = append(ks, x)
	}
	sort.Ints(ks)
	ta, tb := sum(a), sum(b)
	out := make([]PositionRow, 0, len(ks))
	for _, x := range ks {
		out = append(out, PositionRow{x, prob(a[x], ta), prob(b[x], tb)})
	}
	return out
}
func prob(n, d int) float64 {
	if d == 0 {
		return 0
	}
	return float64(n) / float64(d)
}

func entropy(m map[string]int) float64 {
	t := sum(m)
	if t == 0 {
		return 0
	}
	h := 0.
	for _, n := range m {
		p := prob(n, t)
		h -= p * math.Log(p)
	}
	return h
}
func jsInt[K comparable](a, b map[K]int) float64 {
	ta, tb := sum(a), sum(b)
	if ta == 0 || tb == 0 {
		return 0
	}
	keys := map[K]bool{}
	for x := range a {
		keys[x] = true
	}
	for x := range b {
		keys[x] = true
	}
	d := 0.
	for x := range keys {
		p, q := prob(a[x], ta), prob(b[x], tb)
		m := (p + q) / 2
		if p > 0 {
			d += .5 * p * math.Log(p/m)
		}
		if q > 0 {
			d += .5 * q * math.Log(q/m)
		}
	}
	v := 1 - d/math.Ln2
	if v < 0 {
		return 0
	}
	return v
}
func jaccard(a, b map[string]int) float64 {
	intersection := 0
	for k := range a {
		if b[k] > 0 {
			intersection++
		}
	}
	u := len(a) + len(b) - intersection
	if u == 0 {
		return 0
	}
	return float64(intersection) / float64(u)
}
func totalVariation(a, b map[string]int) float64 {
	ta, tb := sum(a), sum(b)
	keys := map[string]bool{}
	for x := range a {
		keys[x] = true
	}
	for x := range b {
		keys[x] = true
	}
	v := 0.
	for x := range keys {
		v += math.Abs(prob(a[x], ta) - prob(b[x], tb))
	}
	return v / 2
}

func contextProfile(a, b map[string]int, existing float64, global map[string]int, limit int) ContextProfile {
	ta, tb := sum(a), sum(b)
	r := ContextProfile{ObservationsA: ta, ObservationsB: tb, EntropyA: entropy(a), EntropyB: entropy(b), Jaccard: jaccard(a, b), JensenShannonSimilarity: jsInt(a, b), ExistingCosineSimilarity: existing}
	r.EffectiveVocabularyA = math.Exp(r.EntropyA)
	r.EffectiveVocabularyB = math.Exp(r.EntropyB)
	r.EntropyDifference = r.EntropyA - r.EntropyB
	r.DistributionA = probabilityRows(a)
	r.DistributionB = probabilityRows(b)
	keys := map[string]bool{}
	for x := range a {
		keys[x] = true
	}
	for x := range b {
		keys[x] = true
	}
	var all []ContextRow
	globalTotal := sum(global)
	for x := range keys {
		pa, pb, pg := prob(a[x], ta), prob(b[x], tb), prob(global[x], globalTotal)
		row := ContextRow{Token: x, ProbabilityA: pa, ProbabilityB: pb, Difference: pa - pb}
		if pg > 0 {
			row.AssociationA, row.AssociationB = pa/pg, pb/pg
		}
		all = append(all, row)
		if a[x] > 0 && b[x] > 0 {
			r.Common = append(r.Common, row)
			if a[x] >= 2 && b[x] >= 2 {
				r.AssociatedBoth = append(r.AssociatedBoth, row)
				if ta >= 30 && tb >= 30 && pa <= .02 && pb <= .02 {
					r.SharedRare = append(r.SharedRare, row)
				}
			}
		}
		if row.Difference > 0 {
			r.SpecificA = append(r.SpecificA, row)
		} else if row.Difference < 0 {
			r.SpecificB = append(r.SpecificB, row)
		}
	}
	sort.Slice(r.Common, func(i, j int) bool {
		x, y := math.Min(r.Common[i].ProbabilityA, r.Common[i].ProbabilityB), math.Min(r.Common[j].ProbabilityA, r.Common[j].ProbabilityB)
		if x != y {
			return x > y
		}
		return r.Common[i].Token < r.Common[j].Token
	})
	sort.Slice(r.AssociatedBoth, func(i, j int) bool {
		x, y := math.Min(r.AssociatedBoth[i].AssociationA, r.AssociatedBoth[i].AssociationB), math.Min(r.AssociatedBoth[j].AssociationA, r.AssociatedBoth[j].AssociationB)
		if x != y {
			return x > y
		}
		return r.AssociatedBoth[i].Token < r.AssociatedBoth[j].Token
	})
	sort.Slice(r.SharedRare, func(i, j int) bool {
		x, y := math.Max(r.SharedRare[i].ProbabilityA, r.SharedRare[i].ProbabilityB), math.Max(r.SharedRare[j].ProbabilityA, r.SharedRare[j].ProbabilityB)
		if x != y {
			return x < y
		}
		return r.SharedRare[i].Token < r.SharedRare[j].Token
	})
	sort.Slice(r.SpecificA, func(i, j int) bool {
		if r.SpecificA[i].Difference != r.SpecificA[j].Difference {
			return r.SpecificA[i].Difference > r.SpecificA[j].Difference
		}
		return r.SpecificA[i].Token < r.SpecificA[j].Token
	})
	sort.Slice(r.SpecificB, func(i, j int) bool {
		if r.SpecificB[i].Difference != r.SpecificB[j].Difference {
			return r.SpecificB[i].Difference < r.SpecificB[j].Difference
		}
		return r.SpecificB[i].Token < r.SpecificB[j].Token
	})
	r.Differential = append([]ContextRow(nil), all...)
	sort.Slice(r.Differential, func(i, j int) bool {
		x, y := math.Abs(r.Differential[i].Difference), math.Abs(r.Differential[j].Difference)
		if x != y {
			return x > y
		}
		return r.Differential[i].Token < r.Differential[j].Token
	})
	if ta >= 30 && tb >= 30 {
		type kv struct {
			s string
			n int
		}
		var absent []kv
		for x, n := range global {
			if a[x] == 0 && b[x] == 0 {
				absent = append(absent, kv{x, n})
			}
		}
		sort.Slice(absent, func(i, j int) bool {
			if absent[i].n != absent[j].n {
				return absent[i].n > absent[j].n
			}
			return absent[i].s < absent[j].s
		})
		for i := 0; i < len(absent) && i < limit; i++ {
			r.SharedAbsent = append(r.SharedAbsent, absent[i].s)
		}
	}
	r.Common = limitRows(r.Common, limit)
	r.AssociatedBoth = limitRows(r.AssociatedBoth, limit)
	r.SpecificA = limitRows(r.SpecificA, limit)
	r.SpecificB = limitRows(r.SpecificB, limit)
	r.Differential = limitRows(r.Differential, limit)
	r.SharedRare = limitRows(r.SharedRare, limit)
	return r
}
func probabilityRows(m map[string]int) []ProbabilityRow {
	total := sum(m)
	out := make([]ProbabilityRow, 0, len(m))
	for token, count := range m {
		out = append(out, ProbabilityRow{token, count, prob(count, total)})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Probability != out[j].Probability {
			return out[i].Probability > out[j].Probability
		}
		return out[i].Token < out[j].Token
	})
	return out
}
func limitRows(x []ContextRow, n int) []ContextRow {
	if len(x) > n {
		return x[:n]
	}
	return x
}
func globalContext(p map[string]profile, side string) map[string]int {
	r := map[string]int{}
	for _, x := range p {
		m := x.Left
		if side == "right" {
			m = x.Right
		}
		for k, n := range m {
			r[k] += n
		}
	}
	return r
}

func explain(r PairResult) []string {
	type c struct {
		name string
		v    float64
	}
	xs := []c{{"positional agreement", r.PositionSimilarity}, {"predecessor-distribution overlap", r.LeftSimilarity}, {"successor-distribution overlap", r.RightSimilarity}}
	sort.Slice(xs, func(i, j int) bool { return xs[i].v > xs[j].v })
	out := []string{fmt.Sprintf("Primary component: %s (%.3f).", xs[0].name, xs[0].v)}
	if xs[1].v >= .6 {
		out = append(out, fmt.Sprintf("Similarity is multidimensional: %s also contributes %.3f.", xs[1].name, xs[1].v))
	} else {
		out = append(out, fmt.Sprintf("Similarity is concentrated: the next component, %s, is %.3f.", xs[1].name, xs[1].v))
	}
	side, row := "left", r.Left.Differential
	if len(r.Right.Differential) > 0 && (len(row) == 0 || math.Abs(r.Right.Differential[0].Difference) > math.Abs(row[0].Difference)) {
		side, row = "right", r.Right.Differential
	}
	if len(row) > 0 {
		direction := r.TokenA
		if row[0].Difference < 0 {
			direction = r.TokenB
		}
		out = append(out, fmt.Sprintf("Largest %s-context difference: %s is more frequent for %s (absolute probability difference %.3f).", side, row[0].Token, direction, math.Abs(row[0].Difference)))
	}
	return out
}

type controlCandidate struct {
	Pair PairSource
	Cost float64
}

type indexedControlCandidate struct {
	pair PairSource
	a, b float64
}
type controlIndex struct{ candidates []indexedControlCandidate }

func newControlIndex(all []PairSource) controlIndex {
	x := controlIndex{candidates: make([]indexedControlCandidate, 0, len(all))}
	for _, p := range all {
		counts := []float64{math.Log1p(float64(p.CountA)), math.Log1p(float64(p.CountB))}
		sort.Float64s(counts)
		x.candidates = append(x.candidates, indexedControlCandidate{p, counts[0], counts[1]})
	}
	return x
}
func (x controlIndex) choose(t PairSource, median float64, n int) []controlCandidate {
	var near, fallback []controlCandidate
	tc := []float64{math.Log1p(float64(t.CountA)), math.Log1p(float64(t.CountB))}
	sort.Float64s(tc)
	for _, indexed := range x.candidates {
		p := indexed.pair
		if key(p.TokenA, p.TokenB) == key(t.TokenA, t.TokenB) || p.TokenA == t.TokenA || p.TokenA == t.TokenB || p.TokenB == t.TokenA || p.TokenB == t.TokenB {
			continue
		}
		cost := math.Abs(indexed.a-tc[0]) + math.Abs(indexed.b-tc[1]) + 2*math.Abs(p.Graphemic-t.Graphemic) + 2*math.Abs(p.Reliability-t.Reliability) + 10*math.Abs(p.Structural-median)
		candidate := controlCandidate{p, cost}
		fallback = append(fallback, candidate)
		if math.Abs(p.Structural-median) <= .05 {
			near = append(near, candidate)
		}
	}
	out := near
	if len(out) < n {
		out = fallback
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Cost != out[j].Cost {
			return out[i].Cost < out[j].Cost
		}
		return out[i].Pair.TokenA+"\x00"+out[i].Pair.TokenB < out[j].Pair.TokenA+"\x00"+out[j].Pair.TokenB
	})
	if len(out) > n {
		out = out[:n]
	}
	return out
}

func medianStructural(p []PairSource) float64 {
	v := make([]float64, len(p))
	for i, x := range p {
		v[i] = x.Structural
	}
	sort.Float64s(v)
	if len(v) == 0 {
		return 0
	}
	if len(v)%2 == 1 {
		return v[len(v)/2]
	}
	return (v[len(v)/2-1] + v[len(v)/2]) / 2
}
func chooseControls(t PairSource, all []PairSource, median float64, n int) []controlCandidate {
	return newControlIndex(all).choose(t, median, n)
}

func decomposeFamily(f FamilyInput, by map[[2]string]PairSource) (FamilyResult, error) {
	r := FamilyResult{ID: f.ID, Tokens: append([]string(nil), f.Tokens...), MeanDistance: map[string]float64{}}
	sort.Strings(r.Tokens)
	for _, e := range f.Edges {
		r.Edges = append(r.Edges, FamilyEdge{e.TokenA, e.TokenB, e.StructuralSimilarity, e.Reliability, e.GraphemicDistance})
	}
	r.Structural = matrix(r.Tokens, by, func(p PairSource) float64 { return p.Structural }, 1)
	r.Reliability = matrix(r.Tokens, by, func(p PairSource) float64 { return p.Reliability }, 1)
	r.Graphemic = matrix(r.Tokens, by, func(p PairSource) float64 { return p.Graphemic }, 0)
	r.Position = matrix(r.Tokens, by, func(p PairSource) float64 { return p.Position }, 1)
	r.Left = matrix(r.Tokens, by, func(p PairSource) float64 { return p.Left }, 1)
	r.Right = matrix(r.Tokens, by, func(p PairSource) float64 { return p.Right }, 1)
	for i, t := range r.Tokens {
		d := 0.
		for j := range r.Tokens {
			if i != j {
				d += 1 - r.Structural.Values[i][j]
			}
		}
		if len(r.Tokens) > 1 {
			d /= float64(len(r.Tokens) - 1)
		}
		r.MeanDistance[t] = d
	}
	ordered := append([]string(nil), r.Tokens...)
	sort.Slice(ordered, func(i, j int) bool {
		if r.MeanDistance[ordered[i]] != r.MeanDistance[ordered[j]] {
			return r.MeanDistance[ordered[i]] < r.MeanDistance[ordered[j]]
		}
		return ordered[i] < ordered[j]
	})
	if len(ordered) > 0 {
		r.Medoid = ordered[0]
		max := r.MeanDistance[ordered[len(ordered)-1]]
		for i := len(ordered) - 1; i >= 0 && math.Abs(r.MeanDistance[ordered[i]]-max) < 1e-12; i-- {
			r.Peripheral = append(r.Peripheral, ordered[i])
		}
		sort.Strings(r.Peripheral)
	}
	return r, nil
}
func matrix(tokens []string, by map[[2]string]PairSource, get func(PairSource) float64, diag float64) Matrix {
	m := Matrix{Tokens: append([]string(nil), tokens...), Values: make([][]float64, len(tokens))}
	for i := range tokens {
		m.Values[i] = make([]float64, len(tokens))
		for j := range tokens {
			if i == j {
				m.Values[i][j] = diag
			} else if p, ok := by[key(tokens[i], tokens[j])]; ok {
				m.Values[i][j] = get(p)
			}
		}
	}
	return m
}
