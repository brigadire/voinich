package lineregime

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"

	"zcore.dev/voinich/internal/tokentransition"
)

// Acc accumulates token-pair distance statistics (task64 section 7/59;
// task65 section 33 reuses it verbatim to reproduce Task64's effect
// sizes byte-for-byte from the same corpus/split/seed).
type Acc struct {
	N                    int
	Exact, D1, DLE1      int
	SumDist, SumNormDist float64
}

func (a *Acc) Add(d, lenA, lenB int) {
	a.N++
	if d == 0 {
		a.Exact++
	}
	if d == 1 {
		a.D1++
	}
	if d <= 1 {
		a.DLE1++
	}
	a.SumDist += float64(d)
	m := lenA
	if lenB > m {
		m = lenB
	}
	if m > 0 {
		a.SumNormDist += float64(d) / float64(m)
	}
}

func (a Acc) Rate() float64 {
	if a.N == 0 {
		return 0
	}
	return float64(a.DLE1) / float64(a.N)
}

func (a Acc) Row(cond string) string {
	n := float64(max(1, a.N))
	return fmt.Sprintf("%s\t%d\t%.9f\t%.9f\t%.9f\t%.9f\t%.9f\n", cond, a.N,
		float64(a.Exact)/n, float64(a.D1)/n, float64(a.DLE1)/n, a.SumDist/n, a.SumNormDist/n)
}

// RateOnly computes the raw (unmatched, no page-pool sampling) d<=1 rate
// for either adjacent (sep1=true) or non-adjacent (sep1=false) within-group
// pairs (task64 section 28).
func RateOnly(lines []Line, minN int, sep1 bool) float64 {
	var a Acc
	for _, l := range Eligible(lines, minN) {
		for _, pr := range WithinLinePairs(l) {
			if (pr.Separation == 1) == sep1 {
				a.Add(pr.Distance, len(l.Tokens[pr.I]), len(l.Tokens[pr.J]))
			}
		}
	}
	return a.Rate()
}

// LenKey returns a symmetric bucket key for a pair of token lengths.
func LenKey(a, b int) string {
	if a > b {
		a, b = b, a
	}
	return strconv.Itoa(a) + "_" + strconv.Itoa(b)
}

// TokenRef locates one token occurrence inside a []Line slice by line
// index and in-line position.
type TokenRef struct {
	Line, Pos int
	Tok       []string
}

// Pool indexes token occurrences by length (globally and per page) for
// matched-control sampling (task64 sections 8-9).
type Pool struct {
	global map[int][]TokenRef
	page   map[string]map[int][]TokenRef
}

func BuildPool(lines []Line) *Pool {
	p := &Pool{global: map[int][]TokenRef{}, page: map[string]map[int][]TokenRef{}}
	for li, l := range lines {
		if p.page[l.Folio] == nil {
			p.page[l.Folio] = map[int][]TokenRef{}
		}
		for pi, t := range l.Tokens {
			ref := TokenRef{Line: li, Pos: pi, Tok: t}
			p.global[len(t)] = append(p.global[len(t)], ref)
			p.page[l.Folio][len(t)] = append(p.page[l.Folio][len(t)], ref)
		}
	}
	return p
}

// SampleControl draws one control pair of lengths (lenA,lenB), excluding
// excludeLine on both sides and requiring the two draws come from two
// different lines (a genuine between-line pair). requireFolio restricts
// the draw to one page (task64 section 8); excludeFolio additionally
// forbids drawing from the caller's own page when searching globally
// (task64 section 9). It returns ok=false after a bounded number of
// attempts rather than looping forever (task64 section 67).
func SampleControl(p *Pool, lines []Line, r *rand.Rand, excludeLine, lenA, lenB int, requireFolio, excludeFolio string) (TokenRef, TokenRef, bool) {
	var poolA, poolB []TokenRef
	if requireFolio != "" {
		m := p.page[requireFolio]
		if m == nil {
			return TokenRef{}, TokenRef{}, false
		}
		poolA, poolB = m[lenA], m[lenB]
	} else {
		poolA, poolB = p.global[lenA], p.global[lenB]
	}
	if len(poolA) == 0 || len(poolB) == 0 {
		return TokenRef{}, TokenRef{}, false
	}
	for attempt := 0; attempt < 8; attempt++ {
		a := poolA[r.Intn(len(poolA))]
		b := poolB[r.Intn(len(poolB))]
		if a.Line == excludeLine || b.Line == excludeLine || a.Line == b.Line {
			continue
		}
		if excludeFolio != "" && (lines[a.Line].Folio == excludeFolio || lines[b.Line].Folio == excludeFolio) {
			continue
		}
		return a, b, true
	}
	return TokenRef{}, TokenRef{}, false
}

type NonAdjObs struct{ Line, Dist int }
type AdjPairObs struct{ Line, LenA, LenB, Dist int }
type LineContribution struct{ NonAdjDLE1, NonAdjN, CtrlDLE1, CtrlN int }

// CoreStats is Task64's primary/matched-control statistics bundle
// (task64 sections 7-11, 41).
type CoreStats struct {
	Adj, NonAdj, Sep2, Sep3, DiffLineSamePage, DiffPage, AdjMatchedControl Acc
	PerLine                                                                []LineContribution
	NLines                                                                 int
}

// ComputeCoreStats is Task64's workhorse: within-line adjacent/non-adjacent
// distance accumulation, same-page and different-page length-matched
// controls, and a length-matched non-adjacent control for adjacent pairs.
// It is deterministic given (lines, minN, seed) and is exported
// specifically so Task65 can reproduce Task64's Delta_line effect sizes
// from the identical algorithm rather than a re-derived approximation
// (task65 section 33).
func ComputeCoreStats(lines []Line, minN int, seed int64) CoreStats {
	p := BuildPool(lines)
	r := rand.New(rand.NewSource(seed))
	var cs CoreStats
	elig := Eligible(lines, minN)
	cs.NLines = len(elig)
	nonAdjPool := map[string][]NonAdjObs{}
	var adjPairs []AdjPairObs
	for _, l := range elig {
		var lc LineContribution
		for _, pr := range WithinLinePairs(l) {
			lenA, lenB := len(l.Tokens[pr.I]), len(l.Tokens[pr.J])
			switch SeparationBucket(pr.Separation) {
			case "SEP1":
				cs.Adj.Add(pr.Distance, lenA, lenB)
				adjPairs = append(adjPairs, AdjPairObs{l.Index, lenA, lenB, pr.Distance})
			case "SEP2":
				cs.NonAdj.Add(pr.Distance, lenA, lenB)
				cs.Sep2.Add(pr.Distance, lenA, lenB)
				nonAdjPool[LenKey(lenA, lenB)] = append(nonAdjPool[LenKey(lenA, lenB)], NonAdjObs{l.Index, pr.Distance})
			default:
				cs.NonAdj.Add(pr.Distance, lenA, lenB)
				cs.Sep3.Add(pr.Distance, lenA, lenB)
				nonAdjPool[LenKey(lenA, lenB)] = append(nonAdjPool[LenKey(lenA, lenB)], NonAdjObs{l.Index, pr.Distance})
			}
			if pr.Separation > 1 {
				lc.NonAdjN++
				if pr.Distance <= 1 {
					lc.NonAdjDLE1++
				}
				if a, b, ok := SampleControl(p, lines, r, l.Index, lenA, lenB, l.Folio, ""); ok {
					d := tokentransition.EditDistance(a.Tok, b.Tok)
					cs.DiffLineSamePage.Add(d, lenA, lenB)
					lc.CtrlN++
					if d <= 1 {
						lc.CtrlDLE1++
					}
				}
				if a, b, ok := SampleControl(p, lines, r, l.Index, lenA, lenB, "", l.Folio); ok {
					cs.DiffPage.Add(tokentransition.EditDistance(a.Tok, b.Tok), lenA, lenB)
				}
			}
		}
		cs.PerLine = append(cs.PerLine, lc)
	}
	for _, ap := range adjPairs {
		cand := nonAdjPool[LenKey(ap.LenA, ap.LenB)]
		if len(cand) == 0 {
			continue
		}
		for attempt := 0; attempt < 10; attempt++ {
			c := cand[r.Intn(len(cand))]
			if c.Line == ap.Line {
				continue
			}
			cs.AdjMatchedControl.Add(c.Dist, ap.LenA, ap.LenB)
			break
		}
	}
	return cs
}

// PageOrderOf returns the folios of lines in first-appearance order.
func PageOrderOf(lines []Line) []string {
	seen := map[string]bool{}
	var order []string
	for _, l := range lines {
		if !seen[l.Folio] {
			seen[l.Folio] = true
			order = append(order, l.Folio)
		}
	}
	return order
}

// Split holds Task64's contiguous-folio-block fold membership (task64
// section 34/45): Train+Validation = Discovery, Test = Replication.
type Split struct {
	Train, Validation, Test, Discovery, Replication map[string]bool
}

// SplitPages implements Task64's exact split rule: pages ordered by first
// appearance, train/validation/test take the first trainFrac/valFrac/rest
// fraction of pages, discovery/replication take the first
// discoveryFrac/rest; no page is ever split across folds.
func SplitPages(pageOrder []string, trainFrac, valFrac, discoveryFrac float64) Split {
	nPages := len(pageOrder)
	trainEnd := int(trainFrac * float64(nPages))
	valEnd := int((trainFrac + valFrac) * float64(nPages))
	discEnd := int(discoveryFrac * float64(nPages))
	if trainEnd < 1 {
		trainEnd = 1
	}
	if valEnd <= trainEnd {
		valEnd = trainEnd + 1
	}
	if valEnd >= nPages {
		valEnd = nPages - 1
	}
	if discEnd <= valEnd {
		discEnd = valEnd
	}
	if discEnd >= nPages {
		discEnd = nPages - 1
	}
	s := Split{Train: map[string]bool{}, Validation: map[string]bool{}, Test: map[string]bool{},
		Discovery: map[string]bool{}, Replication: map[string]bool{}}
	for i, pg := range pageOrder {
		switch {
		case i < trainEnd:
			s.Train[pg] = true
		case i < valEnd:
			s.Validation[pg] = true
		default:
			s.Test[pg] = true
		}
		if i < discEnd {
			s.Discovery[pg] = true
		} else {
			s.Replication[pg] = true
		}
	}
	return s
}

// FilterByPages returns the subset of lines whose Folio is in keep.
func FilterByPages(lines []Line, keep map[string]bool) []Line {
	out := make([]Line, 0, len(lines))
	for _, l := range lines {
		if keep[l.Folio] {
			out = append(out, l)
		}
	}
	return out
}

// -----------------------------------------------------------------------
// Giant d1-component membership and structural profile, generalized from
// Task64 (sections 13, 22) to operate on any flat token list so Task65's
// windows can reuse the identical, authoritative distance/profile
// definitions instead of redefining them (task64 section 4, task65
// section 11).
// -----------------------------------------------------------------------

type unionFind struct{ parent []int }

func newUnionFind(n int) *unionFind {
	p := make([]int, n)
	for i := range p {
		p[i] = i
	}
	return &unionFind{p}
}
func (u *unionFind) find(x int) int {
	for u.parent[x] != x {
		u.parent[x] = u.parent[u.parent[x]]
		x = u.parent[x]
	}
	return x
}
func (u *unionFind) union(a, b int) {
	ra, rb := u.find(a), u.find(b)
	if ra != rb {
		u.parent[ra] = rb
	}
}

// BuildGiantSet returns the set of token forms (glyphs joined) belonging
// to the largest d<=1 connected component of the vocabulary of tokens,
// using the authoritative Task60/63 edit distance (task64 section 13).
func BuildGiantSet(tokens [][]string) map[string]bool {
	seen := map[string][]string{}
	for _, t := range tokens {
		seen[joinGlyphs(t)] = t
	}
	forms := make([]string, 0, len(seen))
	for f := range seen {
		forms = append(forms, f)
	}
	sort.Strings(forms)
	byLen := map[int][]int{}
	for i, f := range forms {
		byLen[len(seen[f])] = append(byLen[len(seen[f])], i)
	}
	uf := newUnionFind(len(forms))
	for length, idxs := range byLen {
		for _, i := range idxs {
			for _, j := range idxs {
				if j <= i {
					continue
				}
				if tokentransition.EditDistance(seen[forms[i]], seen[forms[j]]) == 1 {
					uf.union(i, j)
				}
			}
			for _, j := range byLen[length+1] {
				if tokentransition.EditDistance(seen[forms[i]], seen[forms[j]]) == 1 {
					uf.union(i, j)
				}
			}
		}
	}
	compSize := map[int]int{}
	for i := range forms {
		compSize[uf.find(i)]++
	}
	giantRoot, giantSize := -1, 0
	for root, size := range compSize {
		if size > giantSize {
			giantRoot, giantSize = root, size
		}
	}
	giant := map[string]bool{}
	for i, f := range forms {
		if uf.find(i) == giantRoot {
			giant[f] = true
		}
	}
	return giant
}

func joinGlyphs(t []string) string {
	s := ""
	for _, g := range t {
		s += g
	}
	return s
}

// Profile is a compact structural descriptor of a token window (task64
// section 22, generalized beyond one physical line).
type Profile struct{ MeanLen, GiantFrac, TopInit, TopFinal, TypeEnt float64 }

func topShare(counts map[string]int, n int) float64 {
	best := 0
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if counts[k] > best {
			best = counts[k]
		}
	}
	if n == 0 {
		return 0
	}
	return float64(best) / float64(n)
}

func shannonH(counts map[string]int) float64 {
	n := 0
	for _, v := range counts {
		n += v
	}
	if n == 0 {
		return 0
	}
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := 0.0
	for _, k := range keys {
		p := float64(counts[k]) / float64(n)
		if p > 0 {
			h -= p * math.Log2(p)
		}
	}
	return h
}

// ComputeProfile builds a Profile over an arbitrary set of tokens (a
// physical line in Task64, or a fixed/overlapping window in Task65).
func ComputeProfile(tokens [][]string, giant map[string]bool) Profile {
	if len(tokens) == 0 {
		return Profile{}
	}
	var p Profile
	initCount, finalCount, typeCount := map[string]int{}, map[string]int{}, map[string]int{}
	giantHits := 0
	for _, t := range tokens {
		if len(t) == 0 {
			continue
		}
		p.MeanLen += float64(len(t))
		initCount[t[0]]++
		finalCount[t[len(t)-1]]++
		typeCount[joinGlyphs(t)]++
		if giant[joinGlyphs(t)] {
			giantHits++
		}
	}
	n := float64(len(tokens))
	p.MeanLen /= n
	p.GiantFrac = float64(giantHits) / n
	p.TopInit = topShare(initCount, len(tokens))
	p.TopFinal = topShare(finalCount, len(tokens))
	p.TypeEnt = shannonH(typeCount)
	return p
}

// ProfileDistance is Task64's authoritative primary profile distance: an
// unweighted Euclidean distance over (MeanLen, GiantFrac, TopInit,
// TopFinal, TypeEnt) (task64 section 23, task65 section 11). It is
// symmetric and zero for identical profiles by construction.
func ProfileDistance(a, b Profile) float64 {
	d1, d2, d3, d4, d5 := a.MeanLen-b.MeanLen, a.GiantFrac-b.GiantFrac, a.TopInit-b.TopInit, a.TopFinal-b.TopFinal, a.TypeEnt-b.TypeEnt
	return math.Sqrt(d1*d1 + d2*d2 + d3*d3 + d4*d4 + d5*d5)
}

// ComputeD1Degrees returns, for every distinct token form in tokens, the
// number of other distinct forms within edit distance 1 (the same
// authoritative distance as BuildGiantSet/ProfileDistance), for use as a
// "mean d1 neighborhood degree" window feature (task65 section 8).
func ComputeD1Degrees(tokens [][]string) map[string]int {
	seen := map[string][]string{}
	for _, t := range tokens {
		seen[joinGlyphs(t)] = t
	}
	forms := make([]string, 0, len(seen))
	for f := range seen {
		forms = append(forms, f)
	}
	sort.Strings(forms)
	byLen := map[int][]int{}
	for i, f := range forms {
		byLen[len(seen[f])] = append(byLen[len(seen[f])], i)
	}
	degree := make([]int, len(forms))
	for length, idxs := range byLen {
		for _, i := range idxs {
			for _, j := range idxs {
				if j <= i {
					continue
				}
				if tokentransition.EditDistance(seen[forms[i]], seen[forms[j]]) == 1 {
					degree[i]++
					degree[j]++
				}
			}
			for _, j := range byLen[length+1] {
				if tokentransition.EditDistance(seen[forms[i]], seen[forms[j]]) == 1 {
					degree[i]++
					degree[j]++
				}
			}
		}
	}
	out := make(map[string]int, len(forms))
	for i, f := range forms {
		out[f] = degree[i]
	}
	return out
}
