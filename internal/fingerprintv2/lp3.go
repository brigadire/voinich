package fingerprintv2

import (
	"math/rand"
	"sort"
)

func productiveGraph(c corpus, base editGraph, productive map[string]bool) editGraph {
	glyphs := glyphByToken(c)
	out := editGraph{adj: map[string]map[string]bool{}}
	for _, edge := range base.edgeList() {
		a, b := edge[0], edge[1]
		ab, abOK := ruleFor(a, b, glyphs)
		ba, baOK := ruleFor(b, a, glyphs)
		if (abOK && productive[ab]) || (baOK && productive[ba]) {
			if out.adj[a] == nil {
				out.adj[a] = map[string]bool{}
			}
			if out.adj[b] == nil {
				out.adj[b] = map[string]bool{}
			}
			out.adj[a][b], out.adj[b][a] = true, true
		}
	}
	for n := range out.adj {
		out.nodes = append(out.nodes, n)
	}
	sort.Strings(out.nodes)
	return out
}

func componentDiameter(g editGraph, component []string) int {
	if len(component) < 2 {
		return 0
	}
	inComponent := map[string]bool{}
	for _, n := range component {
		inComponent[n] = true
	}
	best := 0
	for _, start := range component {
		distance := map[string]int{start: 0}
		queue := []string{start}
		for len(queue) > 0 {
			n := queue[0]
			queue = queue[1:]
			next := make([]string, 0, len(g.adj[n]))
			for v := range g.adj[n] {
				if inComponent[v] && !containsDistance(distance, v) {
					next = append(next, v)
				}
			}
			sort.Strings(next)
			for _, v := range next {
				distance[v] = distance[n] + 1
				if distance[v] > best {
					best = distance[v]
				}
				queue = append(queue, v)
			}
		}
	}
	return best
}

func containsDistance(m map[string]int, key string) bool {
	_, ok := m[key]
	return ok
}

func lp3(c corpus, productive map[string]bool, base editGraph, repetitions int, rng *rand.Rand) LP3Result {
	g := productiveGraph(c, base, productive)
	if len(g.nodes) == 0 {
		return LP3Result{ProductiveRuleCount: len(productive), Locality: LocalityResult{
			Available: true, FamilyCount: 0,
		}}
	}
	comps := g.components()
	summaries := make([]FamilySummary, 0)
	small := 0
	large := make([][]string, 0)
	for _, component := range comps {
		if len(component) < 3 {
			small++
			continue
		}
		degreeSum, overlapping := 0, 0
		for _, n := range component {
			degreeSum += len(g.adj[n])
			// Different productive rules incident on one token are the
			// declared graph-theoretic approximation to rule overlap.
			rules := map[string]bool{}
			for v := range g.adj[n] {
				if r, ok := ruleFor(n, v, glyphByToken(c)); ok && productive[r] {
					rules[r] = true
				}
				if r, ok := ruleFor(v, n, glyphByToken(c)); ok && productive[r] {
					rules[r] = true
				}
			}
			if len(rules) > 1 {
				overlapping++
			}
		}
		summaries = append(summaries, FamilySummary{
			Size: len(component), Branching: float64(degreeSum) / float64(len(component)),
			Depth: componentDiameter(g, component), Overlap: float64(overlapping) / float64(len(component)),
			DepthMethod: "exact all-source shortest-path diameter",
		})
		large = append(large, component)
	}
	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].Size != summaries[j].Size {
			return summaries[i].Size > summaries[j].Size
		}
		return summaries[i].Depth > summaries[j].Depth
	})
	locality := locality(c, large, repetitions, rng)
	return LP3Result{ProductiveRuleCount: len(productive), Families: summaries, SmallFamilyCount: small, Locality: locality}
}

func locality(c corpus, families [][]string, repetitions int, rng *rand.Rand) LocalityResult {
	if len(families) == 0 {
		return LocalityResult{Available: true}
	}
	member := map[string]int{}
	for i, family := range families {
		for _, token := range family {
			member[token] = i
		}
	}
	counts := make([]int, len(families))
	lineGroups, pageGroups := make([]map[int]int, len(families)), make([]map[string]int, len(families))
	hasPage := false
	for i := range families {
		lineGroups[i], pageGroups[i] = map[int]int{}, map[string]int{}
	}
	for _, record := range c.records {
		family, ok := member[record.Token]
		if !ok {
			continue
		}
		counts[family]++
		lineGroups[family][record.Line]++
		if record.Page != "" {
			hasPage = true
			pageGroups[family][record.Page]++
		}
	}
	rate := func(groups []map[int]int) float64 {
		num, den := 0, 0
		for i, total := range counts {
			den += choose2(total)
			for _, group := range groups[i] {
				num += choose2(group)
			}
		}
		if den == 0 {
			return 0
		}
		return float64(num) / float64(den)
	}
	observedLine := rate(lineGroups)
	null := make([]float64, repetitions)
	for r := range null {
		order := rng.Perm(len(c.records))
		groups := make([]map[int]int, len(families))
		for i := range groups {
			groups[i] = map[int]int{}
		}
		offset := 0
		for i, n := range counts {
			for _, index := range order[offset : offset+n] {
				groups[i][c.records[index].Line]++
			}
			offset += n
		}
		null[r] = rate(groups)
	}
	result := LocalityResult{
		Available: true, SameLineRate: observedLine, FamilyCount: len(families),
	}
	test := nullTest("lp3/c-global-same-line", "C-GLOBAL family-occurrence placement", observedLine, null)
	result.GlobalNull = &test
	if hasPage {
		num, den := 0, 0
		for i, total := range counts {
			den += choose2(total)
			for _, group := range pageGroups[i] {
				num += choose2(group)
			}
		}
		v := 0.0
		if den > 0 {
			v = float64(num) / float64(den)
		}
		result.SamePageRate = &v
	}
	return result
}

func choose2(v int) int { return v * (v - 1) / 2 }
