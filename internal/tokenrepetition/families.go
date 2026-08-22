package tokenrepetition

import (
	"sort"
	"strings"
)

// EditGraph is the edit-distance<=1 graph over distinct token types
// (task60 section 23): nodes are token types, edges connect types whose
// glyph-sequence Levenshtein distance is exactly 1.
type EditGraph struct {
	Adjacency map[string][]string // sorted, deduplicated neighbor lists
}

// BuildEditGraph finds every edit-distance-1 pair among vocab's types
// using deletion-signature and substitution-wildcard candidate indexing
// (task60 section 48: no O(V^2) all-pairs scan). Candidates are verified
// with the exact Levenshtein distance before being added as an edge.
func BuildEditGraph(vocab []string, glyphs map[string][]string) EditGraph {
	edgeSet := map[string]map[string]bool{}
	addEdge := func(a, b string) {
		if a == b {
			return
		}
		if edgeSet[a] == nil {
			edgeSet[a] = map[string]bool{}
		}
		if edgeSet[b] == nil {
			edgeSet[b] = map[string]bool{}
		}
		edgeSet[a][b] = true
		edgeSet[b][a] = true
	}

	// Substitution candidates: tokens of equal length sharing a
	// one-position wildcard signature differ (if at all) at exactly that
	// position.
	wildcard := map[string][]string{}
	for _, t := range vocab {
		g := glyphs[t]
		for i := range g {
			key := wildcardKey(g, i)
			wildcard[key] = append(wildcard[key], t)
		}
	}
	for _, group := range wildcard {
		if len(group) < 2 {
			continue
		}
		uniq := dedupeSorted(group)
		for i := 0; i < len(uniq); i++ {
			for j := i + 1; j < len(uniq); j++ {
				if LevenshteinGlyphs(glyphs[uniq[i]], glyphs[uniq[j]]) == 1 {
					addEdge(uniq[i], uniq[j])
				}
			}
		}
	}

	// Insertion/deletion candidates: a token's own string matches
	// another (one-glyph-longer) token's single-deletion variant.
	delMap := map[string][]string{} // deletion-variant string -> source tokens (one glyph longer)
	for _, t := range vocab {
		g := glyphs[t]
		for i := range g {
			delMap[joinGlyphs(deleteAt(g, i))] = append(delMap[joinGlyphs(deleteAt(g, i))], t)
		}
	}
	for _, t := range vocab {
		key := joinGlyphs(glyphs[t])
		for _, longer := range delMap[key] {
			if LevenshteinGlyphs(glyphs[t], glyphs[longer]) == 1 {
				addEdge(t, longer)
			}
		}
	}

	g := EditGraph{Adjacency: map[string][]string{}}
	for a, neigh := range edgeSet {
		keys := make([]string, 0, len(neigh))
		for b := range neigh {
			keys = append(keys, b)
		}
		sort.Strings(keys)
		g.Adjacency[a] = keys
	}
	return g
}

func wildcardKey(g []string, skip int) string {
	var b strings.Builder
	b.WriteString(joinGlyphs(g[:skip]))
	b.WriteByte(1)
	b.WriteString(joinGlyphs(g[skip+1:]))
	b.WriteByte(2)
	return b.String()
}

func deleteAt(g []string, i int) []string {
	out := make([]string, 0, len(g)-1)
	out = append(out, g[:i]...)
	out = append(out, g[i+1:]...)
	return out
}

func joinGlyphs(g []string) string { return strings.Join(g, "") }

func dedupeSorted(items []string) []string {
	m := map[string]bool{}
	for _, i := range items {
		m[i] = true
	}
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ConnectedComponents partitions the graph's nodes (task60 section 23).
func (g EditGraph) ConnectedComponents() [][]string {
	visited := map[string]bool{}
	var nodes []string
	for n := range g.Adjacency {
		nodes = append(nodes, n)
	}
	sort.Strings(nodes)
	var comps [][]string
	for _, n := range nodes {
		if visited[n] {
			continue
		}
		var comp []string
		stack := []string{n}
		visited[n] = true
		for len(stack) > 0 {
			cur := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			comp = append(comp, cur)
			for _, nb := range g.Adjacency[cur] {
				if !visited[nb] {
					visited[nb] = true
					stack = append(stack, nb)
				}
			}
		}
		sort.Strings(comp)
		comps = append(comps, comp)
	}
	sort.Slice(comps, func(i, j int) bool {
		if len(comps[i]) != len(comps[j]) {
			return len(comps[i]) > len(comps[j])
		}
		return comps[i][0] < comps[j][0]
	})
	return comps
}

// DegreeOf returns the number of edit-distance-1 neighbors each node has.
func (g EditGraph) DegreeOf() map[string]int {
	out := map[string]int{}
	for n, neigh := range g.Adjacency {
		out[n] = len(neigh)
	}
	return out
}

// GreedyChains performs one deterministic greedy walk per unvisited node
// (task60 section 26): starting from the lowest-sorted unvisited node,
// repeatedly extend to its lowest-sorted unvisited neighbor until stuck,
// then record the path if its length is >= minLength. This is a bounded,
// deterministic heuristic - not exhaustive path enumeration, which would
// be combinatorially unbounded in a dense component - documented as such
// in METHOD.md.
func (g EditGraph) GreedyChains(minLength int) []Chain {
	visited := map[string]bool{}
	var nodes []string
	for n := range g.Adjacency {
		nodes = append(nodes, n)
	}
	sort.Strings(nodes)
	var chains []Chain
	for _, start := range nodes {
		if visited[start] {
			continue
		}
		path := []string{start}
		visited[start] = true
		cur := start
		for {
			next := ""
			for _, nb := range g.Adjacency[cur] {
				if !visited[nb] {
					next = nb
					break
				}
			}
			if next == "" {
				break
			}
			visited[next] = true
			path = append(path, next)
			cur = next
		}
		if len(path) >= minLength {
			chains = append(chains, Chain{Tokens: path})
		}
	}
	return chains
}

// IndependenceExpectedAdjacency returns, for each edge (a,b) in g, the
// expected number of adjacent occurrences under the standard bigram-
// independence null freq(a)*freq(b)/(N-1) - the same independence
// formula task57's transition estimator uses - summed once per
// unordered edge (task60 section 24).
func IndependenceExpectedAdjacency(g EditGraph, freq map[string]int, n int) (expected float64, edgeCount int) {
	if n <= 1 {
		return 0, 0
	}
	seen := map[string]bool{}
	for a, neigh := range g.Adjacency {
		for _, b := range neigh {
			key := a + "\x00" + b
			rkey := b + "\x00" + a
			if seen[key] || seen[rkey] {
				continue
			}
			seen[key] = true
			expected += float64(freq[a]) * float64(freq[b]) / float64(n-1)
			edgeCount++
		}
	}
	return expected, edgeCount
}
