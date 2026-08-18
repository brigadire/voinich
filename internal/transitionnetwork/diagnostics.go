package transitionnetwork

import (
	"math"
	"sort"
)

func computePredictions(a *analysis, minBlock int) {
	backboneSource := map[string]bool{}
	for _, r := range a.Summaries {
		if r.Status == "BACKBONE_PREFERRED" || r.Status == "BACKBONE_DEPLETED" {
			backboneSource[r.Source] = true
		}
	}
	V := float64(len(a.Vocab))
	eligible := map[string]bool{}
	for _, t := range a.Vocab {
		eligible[t] = true
	}
	for hi, h := range a.Data {
		trainUni, trainEdges, trainOpp := map[string]int{}, map[EdgeKey]int{}, map[string]int{}
		contexts := map[[2]string]map[string]int{}
		contextN := map[[2]string]int{}
		for bi, d := range a.Data {
			if bi == hi {
				continue
			}
			for t, n := range d.Counts {
				trainUni[t] += n
			}
			for e, n := range d.Edges {
				trainEdges[e] += n
				trainOpp[e.Source] += n
			}
			bt := d.Block.Tokens
			for i := 2; i < len(bt); i++ {
				x, y, z := bt[i-2].Text, bt[i-1].Text, bt[i].Text
				if eligible[x] && eligible[y] && eligible[z] {
					k := [2]string{x, y}
					if contexts[k] == nil {
						contexts[k] = map[string]int{}
					}
					contexts[k][z]++
					contextN[k]++
				}
			}
		}
		heldUniN := 0
		for _, n := range h.Counts {
			heldUniN += n
		}
		type acc struct {
			n      int
			l0, l1 float64
		}
		all, bb := acc{}, acc{}
		m2n := 0
		m1loss, m2loss := 0., 0.
		for i := 0; i+1 < len(h.Block.Tokens); i++ {
			s, t := h.Block.Tokens[i].Text, h.Block.Tokens[i+1].Text
			if !eligible[s] || !eligible[t] || h.Opp[s] < minBlock {
				continue
			}
			p0 := (float64(h.Counts[t]) + alpha) / (float64(heldUniN) + alpha*V)
			p1 := (float64(trainEdges[EdgeKey{s, t}]) + alpha) / (float64(trainOpp[s]) + alpha*V)
			all.n++
			all.l0 -= math.Log(p0)
			all.l1 -= math.Log(p1)
			if backboneSource[s] {
				bb.n++
				bb.l0 -= math.Log(p0)
				bb.l1 -= math.Log(p1)
			}
		}
		addPred := func(scope string, x acc) {
			r := PredictionRow{Block: h.Block.ID, Scope: scope, N: x.n}
			if x.n > 0 {
				r.LossM0 = x.l0 / float64(x.n)
				r.LossM1 = x.l1 / float64(x.n)
				r.Delta = r.LossM0 - r.LossM1
			}
			a.Predictions = append(a.Predictions, r)
		}
		addPred("all_eligible", all)
		addPred("backbone_sources", bb)
		for i := 2; i < len(h.Block.Tokens); i++ {
			x, y, z := h.Block.Tokens[i-2].Text, h.Block.Tokens[i-1].Text, h.Block.Tokens[i].Text
			if !eligible[x] || !eligible[y] || !eligible[z] {
				continue
			}
			p1 := (float64(trainEdges[EdgeKey{y, z}]) + alpha) / (float64(trainOpp[y]) + alpha*V)
			k := [2]string{x, y}
			p2 := (float64(contexts[k][z]) + alpha) / (float64(contextN[k]) + alpha*V)
			m1loss -= math.Log(p1)
			m2loss -= math.Log(p2)
			m2n++
		}
		r := ModelOrderRow{Block: h.Block.ID, N: m2n}
		if m2n > 0 {
			r.LossM1 = m1loss / float64(m2n)
			r.LossM2 = m2loss / float64(m2n)
			r.Delta = r.LossM1 - r.LossM2
		}
		a.ModelOrder = append(a.ModelOrder, r)
	}
}

func computeGraphDiagnostics(a *analysis, minBlock int, generic bool) {
	// Metadata transfer compares aggregated normalized edge effects within every
	// pair of groups of each pre-specified metadata dimension. In generic
	// mode there is no real Currier-language or scribal-hand covariate to
	// test transfer across (see Config.Generic) - that comparison is a
	// genuine Class A hypothesis test (GENERIC_STAGE_APPLICABILITY_AUDIT.md)
	// and stays NOT_APPLICABLE; only the "joint" dimension, which is already
	// the block-group partition itself, is computed.
	dims := []string{"currier", "hand", "joint"}
	if generic {
		dims = []string{"joint"}
	}
	for _, dim := range dims {
		groups := map[string]map[EdgeKey][]float64{}
		for e, xs := range a.ByEdge {
			for _, x := range xs {
				g := x.Joint
				if dim == "currier" {
					g = x.Currier
				} else if dim == "hand" {
					g = x.Hand
				}
				if groups[g] == nil {
					groups[g] = map[EdgeKey][]float64{}
				}
				groups[g][e] = append(groups[g][e], x.Log2Enrichment)
			}
		}
		var names []string
		for g := range groups {
			names = append(names, g)
		}
		sort.Strings(names)
		for i := range names {
			for j := i + 1; j < len(names); j++ {
				ga, gb := groups[names[i]], groups[names[j]]
				var x, y []float64
				sign := 0
				for e, xa := range ga {
					xb, ok := gb[e]
					if !ok {
						continue
					}
					u, v := median(xa), median(xb)
					x = append(x, u)
					y = append(y, v)
					if (u > 0) == (v > 0) {
						sign++
					}
				}
				r := TransferRow{Dimension: dim, GroupA: names[i], GroupB: names[j], CommonEdges: len(x), EffectCorrelation: pearson(x, y), ProfileSimilarity: (pearson(x, y) + 1) / 2}
				if len(x) > 0 {
					r.SignAgreement = float64(sign) / float64(len(x))
				}
				a.MetadataTransfer = append(a.MetadataTransfer, r)
			}
		}
	}
	graphs := make([]map[EdgeKey]bool, len(a.Data))
	degrees := make([]map[string]float64, len(a.Data))
	for i, d := range a.Data {
		graphs[i] = map[EdgeKey]bool{}
		degrees[i] = map[string]float64{}
		for _, r := range a.Summaries {
			if r.FDRQ > .05 || d.Opp[r.Source] < minBlock {
				continue
			}
			x := effect(d, r.EdgeKey, len(a.Vocab))
			if (r.ExpectedSign == "preferred" && x.Log2Enrichment > 0) || (r.ExpectedSign == "depleted" && x.Log2Enrichment < 0) {
				graphs[i][r.EdgeKey] = true
				degrees[i][r.Source]++
				degrees[i][r.Target]++
			}
		}
	}
	for i := range a.Data {
		for j := i + 1; j < len(a.Data); j++ {
			inter := 0
			union := map[EdgeKey]bool{}
			for e := range graphs[i] {
				union[e] = true
				if graphs[j][e] {
					inter++
				}
			}
			for e := range graphs[j] {
				union[e] = true
			}
			r := GraphSimilarityRow{BlockA: a.Data[i].Block.ID, BlockB: a.Data[j].Block.ID, EdgesA: len(graphs[i]), EdgesB: len(graphs[j]), Intersection: inter}
			if len(union) > 0 {
				r.EdgeJaccard = float64(inter) / float64(len(union))
			}
			var x, y []float64
			for _, t := range a.Vocab {
				x = append(x, degrees[i][t])
				y = append(y, degrees[j][t])
			}
			r.DegreeRankCorrelation = spearman(x, y)
			r.SCCOverlap = sccOverlap(graphs[i], graphs[j])
			a.GraphSimilarity = append(a.GraphSimilarity, r)
		}
	}
}

func sccNodes(g map[EdgeKey]bool) map[string]bool {
	adj, rev := map[string][]string{}, map[string][]string{}
	nodes := map[string]bool{}
	for e := range g {
		adj[e.Source] = append(adj[e.Source], e.Target)
		rev[e.Target] = append(rev[e.Target], e.Source)
		nodes[e.Source] = true
		nodes[e.Target] = true
	}
	seen := map[string]bool{}
	var order []string
	var dfs func(string)
	dfs = func(v string) {
		if seen[v] {
			return
		}
		seen[v] = true
		for _, w := range adj[v] {
			dfs(w)
		}
		order = append(order, v)
	}
	for v := range nodes {
		dfs(v)
	}
	seen = map[string]bool{}
	best := map[string]bool{}
	var collect func(string, map[string]bool)
	collect = func(v string, c map[string]bool) {
		if seen[v] {
			return
		}
		seen[v] = true
		c[v] = true
		for _, w := range rev[v] {
			collect(w, c)
		}
	}
	for i := len(order) - 1; i >= 0; i-- {
		if seen[order[i]] {
			continue
		}
		c := map[string]bool{}
		collect(order[i], c)
		if len(c) > len(best) {
			best = c
		}
	}
	return best
}
func sccOverlap(a, b map[EdgeKey]bool) float64 {
	x, y := sccNodes(a), sccNodes(b)
	u, n := map[string]bool{}, 0
	for t := range x {
		u[t] = true
		if y[t] {
			n++
		}
	}
	for t := range y {
		u[t] = true
	}
	if len(u) == 0 {
		return 0
	}
	return float64(n) / float64(len(u))
}
