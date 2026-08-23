package fingerprintv2

import (
	"fmt"
	"math/rand"
	"sort"

	"zcore.dev/voinich/internal/tokenrepetition"
)

type editGraph struct {
	nodes []string
	adj   map[string]map[string]bool
}

func buildGraph(c corpus) editGraph {
	nodes := vocabulary(c)
	glyphs := glyphByToken(c)
	base := tokenrepetition.BuildEditGraph(nodes, glyphs)
	out := editGraph{nodes: nodes, adj: map[string]map[string]bool{}}
	for _, n := range nodes {
		out.adj[n] = map[string]bool{}
		for _, v := range base.Adjacency[n] {
			out.adj[n][v] = true
		}
	}
	return out
}

func (g editGraph) edgeList() [][2]string {
	out := make([][2]string, 0)
	for _, a := range g.nodes {
		for b := range g.adj[a] {
			if a < b {
				out = append(out, [2]string{a, b})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i][0] != out[j][0] {
			return out[i][0] < out[j][0]
		}
		return out[i][1] < out[j][1]
	})
	return out
}

func (g editGraph) components() [][]string {
	seen := map[string]bool{}
	out := make([][]string, 0)
	for _, start := range g.nodes {
		if seen[start] {
			continue
		}
		stack := []string{start}
		seen[start] = true
		part := make([]string, 0)
		for len(stack) > 0 {
			n := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			part = append(part, n)
			next := make([]string, 0, len(g.adj[n]))
			for v := range g.adj[n] {
				if !seen[v] {
					next = append(next, v)
				}
			}
			sort.Sort(sort.Reverse(sort.StringSlice(next)))
			for _, v := range next {
				seen[v] = true
				stack = append(stack, v)
			}
		}
		sort.Strings(part)
		out = append(out, part)
	}
	sort.Slice(out, func(i, j int) bool {
		if len(out[i]) != len(out[j]) {
			return len(out[i]) > len(out[j])
		}
		return out[i][0] < out[j][0]
	})
	return out
}

func ruleFor(a, b string, glyphs map[string][]string) (string, bool) {
	op, pos, from, to, ok := tokenrepetition.ClassifyEditDistanceOne(glyphs[a], glyphs[b])
	if !ok {
		return "", false
	}
	zone := "INTERNAL"
	longer := len(glyphs[a])
	if len(glyphs[b]) > longer {
		longer = len(glyphs[b])
	}
	if pos == 0 {
		zone = "PREFIX"
	} else if pos == longer-1 {
		zone = "SUFFIX"
	}
	if from == "" {
		from = "∅"
	}
	if to == "" {
		to = "∅"
	}
	return fmt.Sprintf("%s|%s|%s|%s→%s", op, zone, tokenrepetition.PositionClass(pos, longer), from, to), true
}

func lp1(c corpus, g editGraph, threshold int) (LP1Result, map[string]bool) {
	glyphs := glyphByToken(c)
	counts := map[string]int{}
	for _, e := range g.edgeList() {
		for _, directed := range [][2]string{{e[0], e[1]}, {e[1], e[0]}} {
			rule, ok := ruleFor(directed[0], directed[1], glyphs)
			if ok {
				counts[rule]++
			}
		}
	}
	total := 0
	values := make([]int, 0, len(counts))
	for _, key := range orderedKeys(counts) {
		total += counts[key]
		values = append(values, counts[key])
	}
	rules := make([]RuleSupport, 0, len(counts))
	productive := map[string]bool{}
	for _, key := range orderedKeys(counts) {
		if counts[key] >= threshold {
			productive[key] = true
		}
		share := 0.0
		if total > 0 {
			share = float64(counts[key]) / float64(total)
		}
		rules = append(rules, RuleSupport{Rule: key, Support: counts[key], Share: share})
	}
	sort.Slice(rules, func(i, j int) bool {
		if rules[i].Support != rules[j].Support {
			return rules[i].Support > rules[j].Support
		}
		return rules[i].Rule < rules[j].Rule
	})
	top := 0.0
	if len(rules) > 0 {
		top = rules[0].Share
	}
	return LP1Result{
		DirectedPairCount: total, RuleCount: len(rules), SupportGini: gini(values), TopRuleShare: top,
		SupportThreshold: threshold, ProductiveSupport: len(productive), Rules: rules,
	}, productive
}

func ef1(g editGraph) EF1Result {
	degrees, sizes := map[int]int{}, map[int]int{}
	isolate, giant := 0, 0
	for _, n := range g.nodes {
		d := len(g.adj[n])
		degrees[d]++
		if d == 0 {
			isolate++
		}
	}
	comps := g.components()
	for _, c := range comps {
		sizes[len(c)]++
		if len(c) > giant {
			giant = len(c)
		}
	}
	toCounts := func(m map[int]int) []Count {
		keys := make([]int, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		out := make([]Count, len(keys))
		for i, k := range keys {
			out[i] = Count{Value: k, Count: m[k]}
		}
		return out
	}
	share := func(v int) float64 {
		if len(g.nodes) == 0 {
			return 0
		}
		return float64(v) / float64(len(g.nodes))
	}
	return EF1Result{
		VertexCount: len(g.nodes), EdgeCount: len(g.edgeList()), IsolateCount: isolate, IsolateShare: share(isolate),
		ComponentCount: len(comps), GiantComponent: giant, GiantComponentShare: share(giant),
		DegreeDistribution: toCounts(degrees), ComponentSizes: toCounts(sizes),
	}
}

func graphMotifs(g editGraph) (clustering float64, triangles, paths3, cycles4 int) {
	wedges := 0
	common := map[string]int{}
	for _, n := range g.nodes {
		neighbors := make([]string, 0, len(g.adj[n]))
		for v := range g.adj[n] {
			neighbors = append(neighbors, v)
		}
		sort.Strings(neighbors)
		wedges += len(neighbors) * (len(neighbors) - 1) / 2
		for i := 0; i < len(neighbors); i++ {
			for j := i + 1; j < len(neighbors); j++ {
				key := neighbors[i] + "\x00" + neighbors[j]
				common[key]++
			}
		}
	}
	for _, e := range g.edgeList() {
		a, b := e[0], e[1]
		for v := range g.adj[a] {
			if v > b && g.adj[b][v] {
				triangles++
			}
		}
	}
	for _, k := range orderedKeys(common) {
		cycles4 += common[k] * (common[k] - 1) / 2
	}
	cycles4 /= 2
	if wedges > 0 {
		clustering = float64(3*triangles) / float64(wedges)
	}
	return clustering, triangles, wedges, cycles4
}

// degreePreservingSwap is a deterministic-seed configuration-like control:
// it rewires two disjoint edges only when it keeps a simple graph and every
// vertex degree unchanged.
func degreePreservingSwap(g editGraph, attempts int, rng *rand.Rand) editGraph {
	out := editGraph{nodes: append([]string(nil), g.nodes...), adj: map[string]map[string]bool{}}
	for _, n := range g.nodes {
		out.adj[n] = map[string]bool{}
		for v := range g.adj[n] {
			out.adj[n][v] = true
		}
	}
	edges := out.edgeList()
	for i := 0; i < attempts && len(edges) > 1; i++ {
		x, y := edges[rng.Intn(len(edges))], edges[rng.Intn(len(edges))]
		a, b, c, d := x[0], x[1], y[0], y[1]
		if a == c || a == d || b == c || b == d {
			continue
		}
		if rng.Intn(2) == 1 {
			c, d = d, c
		}
		if a == d || c == b || out.adj[a][d] || out.adj[c][b] {
			continue
		}
		delete(out.adj[a], b)
		delete(out.adj[b], a)
		delete(out.adj[c], d)
		delete(out.adj[d], c)
		out.adj[a][d], out.adj[d][a] = true, true
		out.adj[c][b], out.adj[b][c] = true, true
		edges = out.edgeList()
	}
	return out
}
