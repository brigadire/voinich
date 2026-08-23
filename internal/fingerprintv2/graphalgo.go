package fingerprintv2

import (
	"math"
	"math/rand"
	"sort"
	"strconv"

	"zcore.dev/voinich/internal/tokenrepetition"
)

// bfsDistances returns shortest-path distances (in edges) from start to
// every node reachable from it within the graph restricted to inComponent.
// A depth cap <= 0 means unlimited.
func bfsDistances(g editGraph, start string, inComponent map[string]bool, maxDepth int) map[string]int {
	distance := map[string]int{start: 0}
	queue := []string{start}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		if maxDepth > 0 && distance[n] >= maxDepth {
			continue
		}
		next := make([]string, 0, len(g.adj[n]))
		for v := range g.adj[n] {
			if (inComponent == nil || inComponent[v]) && !containsDistance(distance, v) {
				next = append(next, v)
			}
		}
		sort.Strings(next)
		for _, v := range next {
			distance[v] = distance[n] + 1
			queue = append(queue, v)
		}
	}
	return distance
}

// averageShortestPathAndIndirectShare walks every pair once (undirected) and
// reports the mean shortest-path length and the share of pairs whose
// shortest path exceeds one hop (i.e. are only indirectly, transitively
// connected: a~b, b~c does not imply a~c in a paradigmatic sense).
func averageShortestPathAndIndirectShare(g editGraph, component []string) (avg, indirectShare float64) {
	if len(component) < 2 {
		return 0, 0
	}
	inComponent := map[string]bool{}
	for _, n := range component {
		inComponent[n] = true
	}
	sum, pairs, indirect := 0, 0, 0
	for _, start := range component {
		distances := bfsDistances(g, start, inComponent, 0)
		for _, other := range component {
			if other <= start {
				continue
			}
			d := distances[other]
			sum += d
			pairs++
			if d > 1 {
				indirect++
			}
		}
	}
	if pairs == 0 {
		return 0, 0
	}
	return float64(sum) / float64(pairs), float64(indirect) / float64(pairs)
}

// meanInternalEditDistance is the mean raw Levenshtein glyph distance
// between every member pair, independent of the (edit-distance-one) graph
// topology, so a family's true lexical spread can be compared against its
// graph diameter.
func meanInternalEditDistance(component []string, glyphs map[string][]string) float64 {
	if len(component) < 2 {
		return 0
	}
	sum, pairs := 0, 0
	for i := 0; i < len(component); i++ {
		for j := i + 1; j < len(component); j++ {
			sum += tokenrepetition.LevenshteinGlyphs(glyphs[component[i]], glyphs[component[j]])
			pairs++
		}
	}
	if pairs == 0 {
		return 0
	}
	return float64(sum) / float64(pairs)
}

// articulationPointsAndBridges implements the standard Tarjan low-link DFS
// on the (undirected, simple) subgraph induced by component.
func articulationPointsAndBridges(g editGraph, component []string) (articulation int, bridges int) {
	if len(component) < 2 {
		return 0, 0
	}
	disc, low := map[string]int{}, map[string]int{}
	isArticulation := map[string]bool{}
	timer := 0
	var dfs func(u, parent string)
	dfs = func(u, parent string) {
		timer++
		disc[u], low[u] = timer, timer
		children := 0
		neighbors := make([]string, 0, len(g.adj[u]))
		for v := range g.adj[u] {
			neighbors = append(neighbors, v)
		}
		sort.Strings(neighbors)
		for _, v := range neighbors {
			if v == parent {
				continue
			}
			if _, seen := disc[v]; !seen {
				children++
				dfs(v, u)
				if low[v] < low[u] {
					low[u] = low[v]
				}
				if low[v] > disc[u] {
					bridges++
				}
				if (parent == "" && children > 1) || (parent != "" && low[v] >= disc[u]) {
					isArticulation[u] = true
				}
			} else if low[u] > disc[v] {
				low[u] = disc[v]
			}
		}
	}
	if len(component) > 0 {
		dfs(component[0], "")
	}
	return len(isArticulation), bridges
}

// kCoreDecomposition returns each node's coreness via the standard
// degeneracy-ordering peeling algorithm, restricted to component.
func kCoreDecomposition(g editGraph, component []string) map[string]int {
	inComponent := map[string]bool{}
	degree := map[string]int{}
	for _, n := range component {
		inComponent[n] = true
	}
	for _, n := range component {
		d := 0
		for v := range g.adj[n] {
			if inComponent[v] {
				d++
			}
		}
		degree[n] = d
	}
	coreness := map[string]int{}
	removed := map[string]bool{}
	running := 0
	for len(removed) < len(component) {
		minDeg, minNode := -1, ""
		for _, n := range component {
			if removed[n] {
				continue
			}
			if minDeg == -1 || degree[n] < minDeg || (degree[n] == minDeg && n < minNode) {
				minDeg, minNode = degree[n], n
			}
		}
		if minNode == "" {
			break
		}
		// The Batagelj-Zaversnik degeneracy ordering requires the running
		// max of removal-time degrees, not the raw removal-time degree
		// itself: once a low-degree vertex is peeled from what is really a
		// denser core (e.g. one member of a triangle removed only because
		// an unrelated pendant vertex was peeled first), the remaining
		// members of that core must not be under-reported.
		if minDeg > running {
			running = minDeg
		}
		coreness[minNode] = running
		removed[minNode] = true
		for v := range g.adj[minNode] {
			if inComponent[v] && !removed[v] {
				degree[v]--
			}
		}
	}
	return coreness
}

// hubRemovalGiantShare removes the top hubFraction share of nodes by degree
// (ties broken lexicographically) and reports the giant-component share
// before and after.
func hubRemovalGiantShare(g editGraph, hubFraction float64) HubDependence {
	if len(g.nodes) == 0 {
		return HubDependence{}
	}
	before := ef1(g).GiantComponentShare
	type degreeNode struct {
		node   string
		degree int
	}
	ranked := make([]degreeNode, len(g.nodes))
	for i, n := range g.nodes {
		ranked[i] = degreeNode{n, len(g.adj[n])}
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].degree != ranked[j].degree {
			return ranked[i].degree > ranked[j].degree
		}
		return ranked[i].node < ranked[j].node
	})
	remove := int(math.Ceil(hubFraction * float64(len(ranked))))
	if remove > len(ranked) {
		remove = len(ranked)
	}
	removedSet := map[string]bool{}
	for i := 0; i < remove; i++ {
		removedSet[ranked[i].node] = true
	}
	reduced := editGraph{adj: map[string]map[string]bool{}}
	for _, n := range g.nodes {
		if removedSet[n] {
			continue
		}
		reduced.nodes = append(reduced.nodes, n)
		reduced.adj[n] = map[string]bool{}
		for v := range g.adj[n] {
			if !removedSet[v] {
				reduced.adj[n][v] = true
			}
		}
	}
	after := 0.0
	if len(reduced.nodes) > 0 {
		after = ef1(reduced).GiantComponentShare
	}
	return HubDependence{
		HubFraction: hubFraction, RemovedNodes: remove,
		GiantShareBefore: before, GiantShareAfter: after, GiantShareDrop: before - after,
	}
}

// depthLimitedComponents recomputes connected components allowing only
// paths of at most maxHops, revealing how much of the giant component is
// only reachable through longer chains.
func depthLimitedComponents(g editGraph, maxHops int) PathRestriction {
	seen := map[string]bool{}
	sizes := make([]int, 0)
	for _, start := range g.nodes {
		if seen[start] {
			continue
		}
		distances := bfsDistances(g, start, nil, maxHops)
		size := 0
		for n := range distances {
			if !seen[n] {
				seen[n] = true
				size++
			}
		}
		sizes = append(sizes, size)
	}
	giant := 0
	for _, s := range sizes {
		if s > giant {
			giant = s
		}
	}
	share := 0.0
	if len(g.nodes) > 0 {
		share = float64(giant) / float64(len(g.nodes))
	}
	return PathRestriction{MaxHops: maxHops, ComponentCount: len(sizes), GiantShare: share}
}

// labelPropagation is a deterministic-seed, asynchronous label-propagation
// community detector used only as an alternative partition to compare
// against connected components; it is not asserted as ground truth.
func labelPropagation(g editGraph, rng *rand.Rand, maxIterations int) map[string]string {
	labels := map[string]string{}
	for _, n := range g.nodes {
		labels[n] = n
	}
	if len(g.nodes) == 0 {
		return labels
	}
	order := append([]string(nil), g.nodes...)
	for iter := 0; iter < maxIterations; iter++ {
		rng.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
		changed := false
		for _, n := range order {
			counts := map[string]int{}
			for v := range g.adj[n] {
				counts[labels[v]]++
			}
			if len(counts) == 0 {
				continue
			}
			best, bestCount := labels[n], -1
			for _, label := range orderedKeys(counts) {
				if counts[label] > bestCount {
					best, bestCount = label, counts[label]
				}
			}
			if best != labels[n] {
				labels[n] = best
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	return labels
}

// clusterAgreement computes the Adjusted Rand Index, Normalized Mutual
// Information and Variation of Information between two labelings of the
// SAME ordered node set.
func clusterAgreement(a, b []string) (ari, nmi, vi float64) {
	n := len(a)
	if n == 0 || len(b) != n {
		return 0, 0, 0
	}
	contingency := map[string]map[string]int{}
	rowTotal, colTotal := map[string]int{}, map[string]int{}
	for i := 0; i < n; i++ {
		if contingency[a[i]] == nil {
			contingency[a[i]] = map[string]int{}
		}
		contingency[a[i]][b[i]]++
		rowTotal[a[i]]++
		colTotal[b[i]]++
	}
	sumComb := func(v int) float64 { return float64(v*(v-1)) / 2 }
	sumIJ, sumI, sumJ := 0.0, 0.0, 0.0
	for _, row := range contingency {
		for _, v := range row {
			sumIJ += sumComb(v)
		}
	}
	for _, v := range rowTotal {
		sumI += sumComb(v)
	}
	for _, v := range colTotal {
		sumJ += sumComb(v)
	}
	total := sumComb(n)
	expected := sumI * sumJ / maxFloat(total, 1)
	maxIndex := (sumI + sumJ) / 2
	denom := maxIndex - expected
	if denom == 0 {
		ari = 0
	} else {
		ari = (sumIJ - expected) / denom
	}
	ac, bc, joint := map[string]int{}, map[string]int{}, map[string]int{}
	for i := 0; i < n; i++ {
		ac[a[i]]++
		bc[b[i]]++
		joint[a[i]+"\x00"+b[i]]++
	}
	ha, hb := entropy(ac), entropy(bc)
	hJoint := entropy(joint)
	mi := ha + hb - hJoint
	if ha+hb > 0 {
		nmi = 2 * mi / (ha + hb)
	}
	vi = 2*hJoint - ha - hb
	if vi < 0 {
		vi = 0
	}
	return ari, nmi, vi
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// pairwiseCoMembershipStability takes one node->family-label map per
// stability-battery replicate (empty label = node absent from that
// replicate's comparable vocabulary) and reports, for every node in
// universe, the bootstrap inclusion probability: the modal-label share
// across replicates where the node was present at all. A node present in
// every replicate under the same family gets a score of 1; one that
// switches family or drops out gets a lower score.
func pairwiseCoMembershipStability(replicates []map[string]string, universe []string) map[string]float64 {
	scores := map[string]float64{}
	for _, n := range universe {
		counts := map[string]int{}
		present := 0
		for _, labels := range replicates {
			label, ok := labels[n]
			if !ok || label == "" {
				continue
			}
			counts[label]++
			present++
		}
		if present == 0 {
			scores[n] = 0
			continue
		}
		best := 0
		for _, c := range counts {
			if c > best {
				best = c
			}
		}
		scores[n] = float64(best) / float64(present)
	}
	return scores
}

func componentLabel(i int) string { return "C" + strconv.Itoa(i) }
