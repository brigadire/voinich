package main

import "sort"

// FSA is the shared topology induced by M3's deterministic greedy
// Jensen-Shannon state-merging DFA algorithm
// (G1_EXECUTABLE_CONTRACT.json models.M3), reused unchanged by M4 for
// probability estimation on the same topology.
//
// States are addressed by a stable integer id (the surviving
// representative id from union-find merging). Edge counts are the
// occurrence counts accumulated through merging -- M3 discards them for
// its own (uniform) scoring/generation, M4 uses them directly.
type FSA struct {
	Root     int
	States   []int // sorted stable ids, deterministic iteration order
	Edges    map[int]map[string]fsaEdge
	Accept   map[int]int // state -> EOS/accept occurrence count
	Alphabet []string    // sorted DEVELOPMENT glyph alphabet
	Failed   bool
	FailWhy  string
}

type fsaEdge struct {
	target int
	count  int
}

type trieNode struct {
	access   string
	children map[string]int
	eos      int
}

const inductionOpCap = 100000

func InduceFSA(occ []TokenOccurrence, threshold float64, maxStates int) *FSA {
	// 1. Build the raw prefix trie over occurrence-weighted glyph
	// sequences (one edge per (state, glyph); EOS is an accept count on
	// the node, not a separate node).
	nodes := []*trieNode{{access: ""}}
	for _, o := range occ {
		cur := 0
		for _, g := range o.Glyphs {
			nxt, ok := nodes[cur].children[g]
			if !ok {
				nxt = len(nodes)
				nodes = append(nodes, &trieNode{access: nodes[cur].access + "\x00" + g})
				if nodes[cur].children == nil {
					nodes[cur].children = map[string]int{}
				}
				nodes[cur].children[g] = nxt
			}
			cur = nxt
		}
		nodes[cur].eos++
	}
	alphabet := glyphAlphabet(occ)
	A := float64(len(alphabet) + 1)

	parent := make([]int, len(nodes))
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(x int) int {
		for parent[x] != x {
			parent[x] = parent[parent[x]]
			x = parent[x]
		}
		return x
	}

	// Pass-through counts, bottom-up: children are always created after
	// their parent during trie insertion, so processing indices in
	// reverse guarantees every child is finalized before its parent.
	passThrough := make([]int, len(nodes))
	for i := len(nodes) - 1; i >= 0; i-- {
		total := nodes[i].eos
		for _, c := range nodes[i].children {
			total += passThrough[c]
		}
		passThrough[i] = total
	}

	liveEdges := make([]map[string]fsaEdge, len(nodes))
	liveEOS := make([]int, len(nodes))
	liveTotal := make([]int, len(nodes))
	for i, n := range nodes {
		liveEOS[i] = n.eos
		edges := map[string]fsaEdge{}
		for g, c := range n.children {
			edges[g] = fsaEdge{target: c, count: passThrough[c]}
		}
		liveEdges[i] = edges
		liveTotal[i] = passThrough[i]
	}

	// distCache memoizes each surviving representative's next-symbol
	// distribution, since it is compared against many blue candidates
	// before it next changes; union() invalidates the survivor's entry.
	// Without this cache, re-deriving every red state's distribution on
	// every blue-candidate comparison makes induction O(nodes x
	// max_states) map-heavy work, which is prohibitively slow on
	// high-type-diversity synthetic corpora (e.g. MFC0's near-unique
	// i.i.d. tokens).
	distCache := make([][]float64, len(nodes))
	dist := func(r int) []float64 {
		if d := distCache[r]; d != nil {
			return d
		}
		out := make([]float64, len(alphabet)+1)
		total := float64(liveTotal[r]) + 0.5*(A+1)
		for i, g := range alphabet {
			c := 0
			if e, ok := liveEdges[r][g]; ok {
				c = e.count
			}
			out[i] = (float64(c) + 0.5) / total
		}
		out[len(alphabet)] = (float64(liveEOS[r]) + 0.5) / total
		distCache[r] = out
		return out
	}

	opCount := 0
	failed := false
	failWhy := ""

	union := func(a, b int) {
		parent[b] = a
		liveEOS[a] += liveEOS[b]
		liveTotal[a] += liveTotal[b]
		distCache[a] = nil
		for g, e := range liveEdges[b] {
			if _, exists := liveEdges[a][g]; !exists {
				liveEdges[a][g] = e
			} else {
				ex := liveEdges[a][g]
				ex.count += e.count
				liveEdges[a][g] = ex
			}
		}
	}

	type pair struct{ a, b int }
	planClosure := func(x0, y0 int) ([]pair, bool) {
		x0, y0 = find(x0), find(y0)
		if x0 == y0 {
			return nil, true
		}
		visited := map[[2]int]bool{}
		queue := []pair{{x0, y0}}
		var pairs []pair
		for len(queue) > 0 {
			p := queue[0]
			queue = queue[1:]
			a, b := find(p.a), find(p.b)
			if a == b {
				continue
			}
			key := [2]int{a, b}
			if a > b {
				key = [2]int{b, a}
			}
			if visited[key] {
				continue
			}
			visited[key] = true
			opCount++
			if opCount > inductionOpCap {
				return nil, false
			}
			if (liveEOS[a] > 0) != (liveEOS[b] > 0) {
				return nil, false
			}
			pairs = append(pairs, pair{a, b})
			labels := map[string]bool{}
			for g := range liveEdges[a] {
				labels[g] = true
			}
			for g := range liveEdges[b] {
				labels[g] = true
			}
			for _, g := range sortedBoolKeys(labels) {
				ea, oka := liveEdges[a][g]
				eb, okb := liveEdges[b][g]
				if oka && okb {
					queue = append(queue, pair{ea.target, eb.target})
				}
			}
		}
		return pairs, true
	}

	// Processing order: root excluded, remaining nodes by (access
	// string length, access string) ascending -- shortest-access-string
	// order, the frozen tie-break.
	order := make([]int, 0, len(nodes)-1)
	for i := 1; i < len(nodes); i++ {
		order = append(order, i)
	}
	sort.Slice(order, func(i, j int) bool {
		ai, aj := nodes[order[i]].access, nodes[order[j]].access
		if len(ai) != len(aj) {
			return len(ai) < len(aj)
		}
		return ai < aj
	})

	red := []int{0}
	for _, n := range order {
		if failed {
			break
		}
		b := find(n)
		if b == 0 {
			continue // already folded into root via an earlier closure
		}
		already := false
		for _, r := range red {
			if find(r) == b {
				already = true
				break
			}
		}
		if already {
			continue
		}
		bestRed, bestJS := -1, 2.0
		bDist := dist(b)
		for _, r := range red {
			r = find(r)
			opCount++
			if opCount > inductionOpCap {
				failed, failWhy = true, "TRAINING_FAILED: induction operation cap exceeded"
				break
			}
			js := jsDivergence(bDist, dist(r))
			if js <= threshold && js < bestJS {
				bestJS, bestRed = js, r
			}
		}
		if failed {
			break
		}
		merged := false
		if bestRed >= 0 {
			pairs, ok := planClosure(b, bestRed)
			if pairs == nil && !ok && opCount > inductionOpCap {
				failed, failWhy = true, "TRAINING_FAILED: induction operation cap exceeded"
				break
			}
			if ok {
				for _, p := range pairs {
					a2, b2 := find(p.a), find(p.b)
					if a2 == b2 {
						continue
					}
					union(a2, b2)
				}
				merged = true
			}
		}
		if !merged {
			red = append(red, b)
		}
	}

	if !failed {
		reps := map[int]bool{}
		for i := range nodes {
			reps[find(i)] = true
		}
		if len(reps) > maxStates {
			failed, failWhy = true, "TRAINING_FAILED: fixed point exceeds max_states"
		}
	}

	fsa := &FSA{Root: find(0), Edges: map[int]map[string]fsaEdge{}, Accept: map[int]int{}, Alphabet: alphabet, Failed: failed, FailWhy: failWhy}
	if failed {
		return fsa
	}
	repSet := map[int]bool{}
	for i := range nodes {
		repSet[find(i)] = true
	}
	states := make([]int, 0, len(repSet))
	for r := range repSet {
		states = append(states, r)
	}
	sort.Ints(states)
	fsa.States = states
	for _, r := range states {
		edges := map[string]fsaEdge{}
		for g, e := range liveEdges[r] {
			edges[g] = fsaEdge{target: find(e.target), count: e.count}
		}
		fsa.Edges[r] = edges
		fsa.Accept[r] = liveEOS[r]
	}
	return fsa
}

func sortedBoolKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
