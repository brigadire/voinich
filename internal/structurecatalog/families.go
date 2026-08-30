package structurecatalog

import (
	"fmt"
	"sort"
	"strings"
)

type editEdge struct {
	A, B, Kind, From, To string
	Position             int
}
type dsu struct{ p, s []int }

func newDSU(n int) *dsu {
	d := &dsu{p: make([]int, n), s: make([]int, n)}
	for i := range d.p {
		d.p[i] = i
		d.s[i] = 1
	}
	return d
}
func (d *dsu) find(x int) int {
	for d.p[x] != x {
		d.p[x] = d.p[d.p[x]]
		x = d.p[x]
	}
	return x
}
func (d *dsu) union(a, b int) {
	a = d.find(a)
	b = d.find(b)
	if a == b {
		return
	}
	if d.s[a] < d.s[b] {
		a, b = b, a
	}
	d.p[b] = a
	d.s[a] += d.s[b]
}

func buildFamilies(counts map[string]int) ([]Family, map[string]int, []editEdge, map[string]int) {
	tokens := make([]string, 0, len(counts))
	for t := range counts {
		tokens = append(tokens, t)
	}
	sort.Strings(tokens)
	idx := map[string]int{}
	for i, t := range tokens {
		idx[t] = i
	}
	d := newDSU(len(tokens))
	edgeMap := map[string]editEdge{}
	add := func(e editEdge) {
		if e.A > e.B {
			e.A, e.B = e.B, e.A
		}
		k := e.A + "\x00" + e.B
		if _, ok := edgeMap[k]; !ok {
			edgeMap[k] = e
			d.union(idx[e.A], idx[e.B])
		}
	}
	// Insert/delete edges are found by deleting one rune from every token.
	for _, long := range tokens {
		r := []rune(long)
		for i := range r {
			short := string(append(append([]rune{}, r[:i]...), r[i+1:]...))
			if _, ok := idx[short]; ok {
				add(editEdge{A: short, B: long, Kind: "INSERTION", From: "", To: string(r[i]), Position: i})
			}
		}
	}
	// Substitutions share an exact wildcard signature.
	buckets := map[string][]struct {
		t string
		g rune
		p int
	}{}
	for _, t := range tokens {
		r := []rune(t)
		for i, g := range r {
			k := fmt.Sprintf("%d:%s\x00%s", i, string(r[:i]), string(r[i+1:]))
			buckets[k] = append(buckets[k], struct {
				t string
				g rune
				p int
			}{t, g, i})
		}
	}
	for _, b := range buckets {
		for i := 0; i < len(b); i++ {
			for j := i + 1; j < len(b); j++ {
				if b[i].g != b[j].g {
					add(editEdge{A: b[i].t, B: b[j].t, Kind: "SUBSTITUTION", From: string(b[i].g), To: string(b[j].g), Position: b[i].p})
				}
			}
		}
	}
	groups := map[int][]string{}
	for i, t := range tokens {
		groups[d.find(i)] = append(groups[d.find(i)], t)
	}
	roots := make([]int, 0, len(groups))
	for r := range groups {
		roots = append(roots, r)
	}
	sort.Slice(roots, func(i, j int) bool {
		a, b := groups[roots[i]], groups[roots[j]]
		if len(a) != len(b) {
			return len(a) > len(b)
		}
		return strings.Join(a, "\x00") < strings.Join(b, "\x00")
	})
	fams := make([]Family, 0, len(roots))
	tf := map[string]int{}
	for i, r := range roots {
		id := i + 1
		sort.Strings(groups[r])
		fams = append(fams, Family{ID: id, Tokens: groups[r]})
		for _, t := range groups[r] {
			tf[t] = id
		}
	}
	edges := make([]editEdge, 0, len(edgeMap))
	degree := map[string]int{}
	for _, e := range edgeMap {
		edges = append(edges, e)
		degree[e.A]++
		degree[e.B]++
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].A != edges[j].A {
			return edges[i].A < edges[j].A
		}
		return edges[i].B < edges[j].B
	})
	return fams, tf, edges, degree
}
