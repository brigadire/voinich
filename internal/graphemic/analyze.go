package graphemic

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

func Analyze(cfg Config) (Result, error) {
	pairs, err := readPairs(cfg.InputPath)
	if err != nil {
		return Result{}, err
	}
	result := Result{Pairs: pairs}
	tokens := map[string]bool{}
	x := make([]float64, 0, len(pairs))
	y := make([]float64, 0, len(pairs))
	for i := range pairs {
		p := &pairs[i]
		tokens[p.TokenA] = true
		tokens[p.TokenB] = true
		p.GraphemeDistance, p.NormalizedGraphemeDistance, p.GraphemeSimilarity, p.CommonPrefix, p.CommonSuffix, p.LengthDifference = GraphemeMetrics(p.TokenA, p.TokenB)
		p.DiscoveryScore = p.StructuralSimilarity * p.Reliability * p.NormalizedGraphemeDistance
		x = append(x, p.GraphemeSimilarity)
		y = append(y, p.StructuralSimilarity)
	}
	result.Pairs = pairs
	result.TokenCount = len(tokens)
	result.Pearson = pearson(x, y)
	result.Spearman = spearman(x, y)
	result.Bins = makeBins(pairs)
	for _, threshold := range []int{10, 20, 50, 100} {
		result.Frequency = append(result.Frequency, frequencyResult(pairs, threshold, cfg))
	}
	for _, p := range pairs {
		if p.StructuralSimilarity >= cfg.MinStructuralSimilarity && p.Reliability >= cfg.MinReliability && p.NormalizedGraphemeDistance >= cfg.MinGraphemicDistance {
			result.Distant = append(result.Distant, p)
		}
		if p.StructuralSimilarity >= cfg.MinStructuralSimilarity && p.Reliability >= cfg.MinReliability && p.GraphemeSimilarity >= cfg.MinCloseSimilarity {
			result.Close = append(result.Close, p)
		}
	}
	sortPairs(result.Distant, func(p Pair) float64 { return p.DiscoveryScore })
	sortPairs(result.Close, func(p Pair) float64 { return p.StructuralSimilarity * p.Reliability * p.GraphemeSimilarity })
	structVals, relVals, distVals := make([]float64, len(pairs)), make([]float64, len(pairs)), make([]float64, len(pairs))
	for i, p := range pairs {
		structVals[i] = p.StructuralSimilarity
		relVals[i] = p.Reliability
		distVals[i] = p.NormalizedGraphemeDistance
	}
	result.DistantPercentileCutoffs = map[string]float64{"structural_p95": percentile(structVals, .95), "reliability_p75": percentile(relVals, .75), "graphemic_distance_p75": percentile(distVals, .75)}
	for _, p := range pairs {
		if p.StructuralSimilarity >= result.DistantPercentileCutoffs["structural_p95"] && p.Reliability >= result.DistantPercentileCutoffs["reliability_p75"] && p.NormalizedGraphemeDistance >= result.DistantPercentileCutoffs["graphemic_distance_p75"] {
			result.PercentileDistant = append(result.PercentileDistant, p)
		}
	}
	sortPairs(result.PercentileDistant, func(p Pair) float64 { return p.DiscoveryScore })
	result.CloseFamilies = components(pairs, func(p Pair) bool {
		return p.StructuralSimilarity >= cfg.MinStructuralSimilarity && p.Reliability >= cfg.MinReliability && p.GraphemeSimilarity >= cfg.MinCloseSimilarity
	})
	result.DistantFamilies = components(pairs, func(p Pair) bool {
		return p.StructuralSimilarity >= cfg.MinStructuralSimilarity && p.Reliability >= cfg.MinReliability && p.NormalizedGraphemeDistance >= cfg.MinGraphemicDistance
	})
	return result, nil
}

func readPairs(path string) ([]Pair, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	if !scanner.Scan() {
		if scanner.Err() != nil {
			return nil, scanner.Err()
		}
		return nil, fmt.Errorf("empty pair dataset")
	}
	h := strings.Split(scanner.Text(), "\t")
	cols := map[string]int{}
	for i, s := range h {
		cols[s] = i
	}
	required := []string{"token_a", "token_b", "count_a", "count_b", "position_similarity", "left_similarity", "right_similarity", "raw_similarity", "position_reliability", "left_reliability", "right_reliability", "total_evidence_weight", "evidence_strength", "diagnostic_weighted_similarity"}
	for _, s := range required {
		if _, ok := cols[s]; !ok {
			return nil, fmt.Errorf("missing column %q", s)
		}
	}
	var out []Pair
	line := 1
	for scanner.Scan() {
		line++
		row := strings.Split(scanner.Text(), "\t")
		if len(row) != len(h) {
			return nil, fmt.Errorf("line %d: got %d columns, want %d", line, len(row), len(h))
		}
		get := func(s string) string { return row[cols[s]] }
		atoi := func(s string) (int, error) { return strconv.Atoi(get(s)) }
		atof := func(s string) (float64, error) { return strconv.ParseFloat(get(s), 64) }
		p := Pair{TokenA: get("token_a"), TokenB: get("token_b")}
		if p.CountA, e = atoi("count_a"); e != nil {
			return nil, fmt.Errorf("line %d count_a: %w", line, e)
		}
		if p.CountB, e = atoi("count_b"); e != nil {
			return nil, e
		}
		vals := []*float64{&p.PositionSimilarity, &p.LeftSimilarity, &p.RightSimilarity, &p.StructuralSimilarity, &p.PositionReliability, &p.LeftReliability, &p.RightReliability, &p.TotalEvidenceWeight, &p.Reliability}
		names := []string{"position_similarity", "left_similarity", "right_similarity", "raw_similarity", "position_reliability", "left_reliability", "right_reliability", "total_evidence_weight", "evidence_strength"}
		for i := range vals {
			if *vals[i], e = atof(names[i]); e != nil {
				return nil, fmt.Errorf("line %d %s: %w", line, names[i], e)
			}
		}
		p.DiagnosticWeightedSimilarity = get("diagnostic_weighted_similarity")
		out = append(out, p)
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	return out, nil
}

func makeBins(pairs []Pair) []Bin {
	out := make([]Bin, 10)
	vals := make([][]float64, 10)
	for i := range out {
		out[i].Range = fmt.Sprintf("%.1f–%.1f", float64(i)/10, float64(i+1)/10)
	}
	for _, p := range pairs {
		i := int(p.NormalizedGraphemeDistance * 10)
		if i > 9 {
			i = 9
		}
		vals[i] = append(vals[i], p.StructuralSimilarity)
	}
	for i, v := range vals {
		out[i].PairCount = len(v)
		if len(v) > 0 {
			var sum float64
			for _, n := range v {
				sum += n
			}
			out[i].Mean = sum / float64(len(v))
			out[i].Median = percentile(v, .5)
			out[i].P90 = percentile(v, .9)
			out[i].P95 = percentile(v, .95)
		}
	}
	return out
}

func frequencyResult(pairs []Pair, t int, cfg Config) FrequencyResult {
	var x, y []float64
	n := 0
	for _, p := range pairs {
		if p.CountA >= t && p.CountB >= t {
			x = append(x, p.GraphemeSimilarity)
			y = append(y, p.StructuralSimilarity)
			if p.StructuralSimilarity >= cfg.MinStructuralSimilarity && p.Reliability >= cfg.MinReliability && p.NormalizedGraphemeDistance >= cfg.MinGraphemicDistance {
				n++
			}
		}
	}
	return FrequencyResult{t, len(x), pearson(x, y), spearman(x, y), n}
}
func sortPairs(p []Pair, score func(Pair) float64) {
	sort.SliceStable(p, func(i, j int) bool {
		a, b := score(p[i]), score(p[j])
		if a != b {
			return a > b
		}
		if p[i].TokenA != p[j].TokenA {
			return p[i].TokenA < p[j].TokenA
		}
		return p[i].TokenB < p[j].TokenB
	})
}

func components(pairs []Pair, accept func(Pair) bool) []Family {
	adj := map[string][]string{}
	edges := map[string][]FamilyEdge{}
	for _, p := range pairs {
		if !accept(p) {
			continue
		}
		adj[p.TokenA] = append(adj[p.TokenA], p.TokenB)
		adj[p.TokenB] = append(adj[p.TokenB], p.TokenA)
		e := FamilyEdge{p.TokenA, p.TokenB, p.StructuralSimilarity, p.Reliability, p.NormalizedGraphemeDistance}
		edges[p.TokenA] = append(edges[p.TokenA], e)
		edges[p.TokenB] = append(edges[p.TokenB], e)
	}
	keys := make([]string, 0, len(adj))
	for k := range adj {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	seen := map[string]bool{}
	var out []Family
	for _, start := range keys {
		if seen[start] {
			continue
		}
		queue := []string{start}
		seen[start] = true
		var nodes []string
		edgeSet := map[string]FamilyEdge{}
		for len(queue) > 0 {
			n := queue[0]
			queue = queue[1:]
			nodes = append(nodes, n)
			for _, e := range edges[n] {
				key := e.TokenA + "\x00" + e.TokenB
				edgeSet[key] = e
			}
			for _, next := range adj[n] {
				if !seen[next] {
					seen[next] = true
					queue = append(queue, next)
				}
			}
		}
		if len(nodes) < 2 {
			continue
		}
		sort.Strings(nodes)
		es := make([]FamilyEdge, 0, len(edgeSet))
		for _, e := range edgeSet {
			es = append(es, e)
		}
		sort.Slice(es, func(i, j int) bool {
			if es[i].TokenA != es[j].TokenA {
				return es[i].TokenA < es[j].TokenA
			}
			return es[i].TokenB < es[j].TokenB
		})
		out = append(out, Family{ID: len(out) + 1, Tokens: nodes, Edges: es})
	}
	return out
}
