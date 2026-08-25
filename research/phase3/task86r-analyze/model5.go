package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// M5: frequent-substring deterministic slot grammar-v1.
// G1_EXECUTABLE_CONTRACT.json models.M5. Unit is COMPONENT: a within-TOKEN
// glyph substring of length 1..3. Every stage below is deterministic given
// DEVELOPMENT; the same frozen component inventory and segmentation
// procedure are reused, unchanged, to segment VALIDATION/HELDOUT/negative
// occurrences at scoring time (never re-fit).
type M5Model struct {
	candidateID string
	slotCount   int
	minSupport  int
	invCap      int

	componentSet map[string]bool // frozen inventory used by segmentation (glyph-joined strings)

	retainedRules map[string][]int // ruleKey -> length vector
	ruleOrder     []string         // deterministic enumeration order (support desc, key asc)
	ruleFinalCount map[string]int
	slotKeep      map[string]map[int]map[string]bool // rule -> slot -> kept component set
	slotCount2    map[string]map[int]map[string]int  // rule -> slot -> component -> count (all observed)

	exceptionCount map[string]int // glyph-joined string -> occurrence count

	// segCache memoizes segment() by glyph-joined key: the same TOKEN
	// type is frequently rescored many times (HELDOUT occurrences,
	// negative-pair positives/negatives, generation-convergence
	// replicates), and segmentation is a pure function of the glyphs
	// given this model's frozen componentSet/slotCount.
	segCache map[string]segCacheEntry
	exceptionOrder []string

	totalOccurrences  int
	totalRuleOccur    int
	totalExceptOccur  int
	nDev              int
	bitsReal          float64
}

const m5Alpha = 0.5

func joinGlyphs(g []string) string { return strings.Join(g, "\x00") }
func splitGlyphs(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "\x00")
}

func serializeRule(lengths []int) string {
	parts := make([]string, len(lengths))
	for i, l := range lengths {
		parts[i] = strconv.Itoa(l)
	}
	return fmt.Sprintf("%d:%s", len(lengths), strings.Join(parts, ","))
}

// buildComponentInventory implements the frozen "components" step.
func buildComponentInventory(occ []TokenOccurrence, minSupport, invCap int) map[string]bool {
	types := map[string][]string{} // distinct type (glyph-joined) -> its glyphs
	for _, o := range occ {
		types[joinGlyphs(o.Glyphs)] = o.Glyphs
	}
	support := map[string]int{} // substring (glyph-joined) -> distinct-type count
	for _, glyphs := range types {
		seen := map[string]bool{}
		for L := 1; L <= 3; L++ {
			for i := 0; i+L <= len(glyphs); i++ {
				sub := joinGlyphs(glyphs[i : i+L])
				seen[sub] = true
			}
		}
		for sub := range seen {
			support[sub]++
		}
	}
	type cand struct {
		sub string
		sup int
		ln  int
	}
	var cands []cand
	for sub, sup := range support {
		if sup >= minSupport {
			cands = append(cands, cand{sub: sub, sup: sup, ln: len(splitGlyphs(sub))})
		}
	}
	sort.Slice(cands, func(i, j int) bool {
		if cands[i].sup != cands[j].sup {
			return cands[i].sup > cands[j].sup
		}
		if cands[i].ln != cands[j].ln {
			return cands[i].ln > cands[j].ln
		}
		return cands[i].sub < cands[j].sub
	})
	if len(cands) > invCap {
		cands = cands[:invCap]
	}
	set := map[string]bool{}
	for _, c := range cands {
		set[c.sub] = true
	}
	// Append every single glyph not already retained, code-point order.
	alphabet := glyphAlphabet(occ)
	for _, g := range alphabet {
		set[g] = true // already glyph-joined-trivial for length 1
	}
	return set
}

// segmentation is the frozen DP: exact partition into 1..slotCount
// components (each length 1..3, each a member of componentSet), primary
// key minimal component count, secondary key lexicographically maximal
// left-to-right length sequence, tertiary key lexicographically minimal
// left-to-right component-string sequence.
func segment(glyphs []string, componentSet map[string]bool, slotCount int) ([]string, bool) {
	n := len(glyphs)
	if n == 0 {
		return nil, false
	}
	// segIsBetter compares two full segmentations by the frozen 3-level
	// criterion: fewer components; then lexicographically maximal
	// left-to-right length sequence; then lexicographically minimal
	// left-to-right component-string sequence.
	segIsBetter := func(cur, prev []string) bool {
		if len(cur) != len(prev) {
			return len(cur) < len(prev)
		}
		for i := range cur {
			cl, pl := len(splitGlyphs(cur[i])), len(splitGlyphs(prev[i]))
			if cl != pl {
				return cl > pl
			}
		}
		for i := range cur {
			if cur[i] != prev[i] {
				return cur[i] < prev[i]
			}
		}
		return false
	}
	var best []string
	var comps []string
	var walk func(pos int)
	walk = func(pos int) {
		if pos == n {
			if len(comps) == 0 || len(comps) > slotCount {
				return
			}
			if best == nil || segIsBetter(comps, best) {
				best = append([]string(nil), comps...)
			}
			return
		}
		if len(comps) == slotCount {
			return
		}
		for L := 1; L <= 3 && pos+L <= n; L++ {
			sub := joinGlyphs(glyphs[pos : pos+L])
			if !componentSet[sub] {
				continue
			}
			comps = append(comps, sub)
			walk(pos + L)
			comps = comps[:len(comps)-1]
		}
	}
	walk(0)
	if best == nil {
		return nil, false
	}
	return best, true
}

func FitM5(occ []TokenOccurrence, cand Candidate, bitsReal float64) *M5Model {
	slotCount := cand.Int("slot_count", 2)
	minSupport := cand.Int("min_support", 2)
	invCap := cand.Int("inventory_cap", 64)

	componentSet := buildComponentInventory(occ, minSupport, invCap)

	m := &M5Model{
		candidateID: cand.CandidateID, slotCount: slotCount, minSupport: minSupport, invCap: invCap,
		componentSet:   componentSet,
		retainedRules:  map[string][]int{},
		ruleFinalCount: map[string]int{},
		slotKeep:       map[string]map[int]map[string]bool{},
		slotCount2:     map[string]map[int]map[string]int{},
		exceptionCount: map[string]int{},
		nDev:           len(occ),
		bitsReal:       bitsReal,
	}

	type segd struct {
		raw   string
		comps []string
	}
	ruleOccur := map[string]int{}
	ruleLengths := map[string][]int{}
	slotRaw := map[string]map[int]map[string]int{}
	var segmented []segd
	for _, o := range occ {
		key := joinGlyphs(o.Glyphs)
		comps, ok := segment(o.Glyphs, componentSet, slotCount)
		if !ok {
			m.exceptionCount[key]++
			continue
		}
		lengths := make([]int, len(comps))
		for i, c := range comps {
			lengths[i] = len(splitGlyphs(c))
		}
		rk := serializeRule(lengths)
		ruleOccur[rk]++
		ruleLengths[rk] = lengths
		if slotRaw[rk] == nil {
			slotRaw[rk] = map[int]map[string]int{}
		}
		for i, c := range comps {
			if slotRaw[rk][i] == nil {
				slotRaw[rk][i] = map[string]int{}
			}
			slotRaw[rk][i][c]++
		}
		segmented = append(segmented, segd{raw: key, comps: comps})
	}

	for rk, sup := range ruleOccur {
		if sup >= minSupport {
			m.retainedRules[rk] = ruleLengths[rk]
		}
	}
	m.ruleOrder = sortRuleKeysBySupport(m.retainedRules, ruleOccur)

	// Slot truncation per retained rule.
	for rk := range m.retainedRules {
		m.slotKeep[rk] = map[int]map[string]bool{}
		m.slotCount2[rk] = map[int]map[string]int{}
		for slot, counts := range slotRaw[rk] {
			type sc struct {
				s string
				c int
			}
			var list []sc
			for s, c := range counts {
				list = append(list, sc{s, c})
			}
			sort.Slice(list, func(i, j int) bool {
				if list[i].c != list[j].c {
					return list[i].c > list[j].c
				}
				return list[i].s < list[j].s
			})
			if len(list) > invCap {
				list = list[:invCap]
			}
			keep := map[string]bool{}
			for _, e := range list {
				keep[e.s] = true
			}
			m.slotKeep[rk][slot] = keep
			m.slotCount2[rk][slot] = counts
		}
	}

	// Reclassify occurrences of retained rules whose realized slot
	// component was truncated out of that slot's inventory.
	for _, s := range segmented {
		rk := serializeRuleFromComps(s.comps)
		if _, ok := m.retainedRules[rk]; !ok {
			m.exceptionCount[s.raw]++ // STRUCTURAL_EXCEPTION (rule below support)
			continue
		}
		allKept := true
		for i, c := range s.comps {
			if !m.slotKeep[rk][i][c] {
				allKept = false
				break
			}
		}
		if !allKept {
			m.exceptionCount[s.raw]++ // LEXICAL_EXCEPTION (slot truncated)
			continue
		}
		m.ruleFinalCount[rk]++
		m.totalRuleOccur++
	}
	for k := range m.exceptionCount {
		m.exceptionOrder = append(m.exceptionOrder, k)
	}
	sort.Strings(m.exceptionOrder)
	for _, c := range m.exceptionCount {
		m.totalExceptOccur += c
	}
	m.totalOccurrences = m.totalRuleOccur + m.totalExceptOccur
	return m
}

func serializeRuleFromComps(comps []string) string {
	lengths := make([]int, len(comps))
	for i, c := range comps {
		lengths[i] = len(splitGlyphs(c))
	}
	return serializeRule(lengths)
}

// sortRuleKeysBySupport orders retained rule keys by the frozen
// enumeration rule: occurrence support descending, then serialized rule
// key ascending.
func sortRuleKeysBySupport(retained map[string][]int, support map[string]int) []string {
	keys := make([]string, 0, len(retained))
	for k := range retained {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if support[keys[i]] != support[keys[j]] {
			return support[keys[i]] > support[keys[j]]
		}
		return keys[i] < keys[j]
	})
	return keys
}

// ComponentCount is the frozen-segmentation component count of an
// arbitrary glyph sequence, used by NEGATIVE_TOKEN_PROTOCOL.md's M5-only
// extra source-matching criterion.
func (m *M5Model) ComponentCount(glyphs []string) (int, bool) {
	comps, ok := segment(glyphs, m.componentSet, m.slotCount)
	if !ok {
		return 0, false
	}
	return len(comps), true
}

func (m *M5Model) ModelClass() string             { return "M5" }
func (m *M5Model) CandidateID() string             { return m.candidateID }
func (m *M5Model) Unit() string                    { return "COMPONENT" }
func (m *M5Model) TrainingFailed() (bool, string)  { return false, "" }

// pExceptionBranch / pRuleBranch: additive-0.5 two-way split.
func (m *M5Model) branchProbs() (pRule, pException float64) {
	total := float64(m.totalOccurrences) + m.alpha2()
	pRule = (float64(m.totalRuleOccur) + m5Alpha) / total
	pException = (float64(m.totalExceptOccur) + m5Alpha) / total
	return
}
func (m *M5Model) alpha2() float64 { return m5Alpha * 2 }

func (m *M5Model) ruleProb(rk string) float64 {
	total := float64(m.totalRuleOccur) + m5Alpha*float64(len(m.ruleOrder))
	if total <= 0 {
		return 0
	}
	return (float64(m.ruleFinalCount[rk]) + m5Alpha) / total
}

func (m *M5Model) slotProb(rk string, slot int, comp string) float64 {
	keep := m.slotKeep[rk][slot]
	if !keep[comp] {
		return 0
	}
	counts := m.slotCount2[rk][slot]
	total := 0
	k := 0
	for s := range keep {
		total += counts[s]
		k++
	}
	return (float64(counts[comp]) + m5Alpha) / (float64(total) + m5Alpha*float64(k))
}

func (m *M5Model) exceptionLiteralProb(key string) float64 {
	c, ok := m.exceptionCount[key]
	if !ok {
		return 0
	}
	numEntries := len(m.exceptionOrder)
	return (float64(c) + m5Alpha) / (float64(m.totalExceptOccur) + m5Alpha*float64(numEntries))
}

type segCacheEntry struct {
	rk    string
	comps []string
	ok    bool
}

// ruleForToken deterministically re-segments an arbitrary glyph sequence
// under the frozen inventory, for scoring occurrences never seen at fit
// time (VALIDATION/HELDOUT/negative controls). Memoized: segmentation is a
// pure function of glyphs given this model's frozen state.
func (m *M5Model) ruleForToken(glyphs []string) (rk string, comps []string, ok bool) {
	key := joinGlyphs(glyphs)
	if e, hit := m.segCache[key]; hit {
		return e.rk, e.comps, e.ok
	}
	comps, ok = segment(glyphs, m.componentSet, m.slotCount)
	if !ok {
		if m.segCache == nil {
			m.segCache = map[string]segCacheEntry{}
		}
		m.segCache[key] = segCacheEntry{ok: false}
		return "", nil, false
	}
	rk = serializeRuleFromComps(comps)
	_, retained := m.retainedRules[rk]
	if m.segCache == nil {
		m.segCache = map[string]segCacheEntry{}
	}
	if !retained {
		m.segCache[key] = segCacheEntry{ok: false}
		return "", nil, false
	}
	m.segCache[key] = segCacheEntry{rk: rk, comps: comps, ok: true}
	return rk, comps, true
}

// dominantPath picks, deterministically, which generative path this exact
// occurrence is explained by: an exact exception-table literal match takes
// priority (matches training's own classification for DEVELOPMENT
// occurrences and is the simplest sufficient rule for any occurrence),
// else a valid frozen-rule segmentation, else neither.
func (m *M5Model) dominantPath(glyphs []string) (isException bool, rk string, comps []string, valid bool) {
	key := joinGlyphs(glyphs)
	if _, ok := m.exceptionCount[key]; ok {
		return true, "", nil, true
	}
	if rk, comps, ok := m.ruleForToken(glyphs); ok {
		allKept := true
		for i, c := range comps {
			if !m.slotKeep[rk][i][c] {
				allKept = false
				break
			}
		}
		if allKept {
			return false, rk, comps, true
		}
	}
	return false, "", nil, false
}

func (m *M5Model) Events(raw string, glyphs []string) []ScoreEvent {
	pRule, pException := m.branchProbs()
	isException, rk, comps, valid := m.dominantPath(glyphs)
	if !valid {
		branchDecision := Decision{Outcomes: []string{"EXCEPTION", "RULE"}, Probs: []float64{pException, pRule}, ObservedIndex: -1}
		return []ScoreEvent{branchDecision.Event()}
	}
	branchDecision := Decision{Outcomes: []string{"EXCEPTION", "RULE"}, Probs: []float64{pException, pRule}}
	events := make([]ScoreEvent, 0, 2+m.slotCount)
	if isException {
		branchDecision.ObservedIndex = 0
		events = append(events, branchDecision.Event())
		key := joinGlyphs(glyphs)
		outs := m.exceptionOrder
		probs := make([]float64, len(outs))
		obs := -1
		for i, o := range outs {
			probs[i] = m.exceptionLiteralProb(o)
			if o == key {
				obs = i
			}
		}
		d := Decision{Outcomes: outs, Probs: probs, ObservedIndex: obs}
		events = append(events, d.Event())
		return events
	}
	branchDecision.ObservedIndex = 1
	events = append(events, branchDecision.Event())
	ruleOuts := m.ruleOrder
	ruleProbs := make([]float64, len(ruleOuts))
	obsRule := -1
	for i, r := range ruleOuts {
		ruleProbs[i] = m.ruleProb(r)
		if r == rk {
			obsRule = i
		}
	}
	rd := Decision{Outcomes: ruleOuts, Probs: ruleProbs, ObservedIndex: obsRule}
	events = append(events, rd.Event())
	for i, c := range comps {
		keep := m.slotKeep[rk][i]
		var outs []string
		for s := range keep {
			outs = append(outs, s)
		}
		sort.Strings(outs)
		probs := make([]float64, len(outs))
		obs := -1
		for j, s := range outs {
			probs[j] = m.slotProb(rk, i, s)
			if s == c {
				obs = j
			}
		}
		sd := Decision{Outcomes: outs, Probs: probs, ObservedIndex: obs}
		events = append(events, sd.Event())
	}
	return events
}

func (m *M5Model) TokenNegLog2Prob(raw string, glyphs []string) float64 {
	total := 0.0
	for _, e := range m.Events(raw, glyphs) {
		total += e.NegLog2Prob
	}
	return total
}

func (m *M5Model) ScoredUnits(glyphs []string) int { return len(glyphs) + 1 }

func (m *M5Model) Generate(p *PRNG) GeneratedToken {
	pRule, pException := m.branchProbs()
	branchIdx := DrawIndex(p, Cumulative([]float64{pException, pRule}))
	if branchIdx == 0 {
		outs := m.exceptionOrder
		if len(outs) == 0 {
			return GeneratedToken{NonGenerative: true}
		}
		probs := make([]float64, len(outs))
		for i, o := range outs {
			probs[i] = m.exceptionLiteralProb(o)
		}
		idx := DrawIndex(p, Cumulative(probs))
		g := splitGlyphs(outs[idx])
		if len(g) == 0 {
			return GeneratedToken{NonGenerative: true}
		}
		return GeneratedToken{Glyphs: g}
	}
	if len(m.ruleOrder) == 0 {
		return GeneratedToken{NonGenerative: true}
	}
	ruleProbs := make([]float64, len(m.ruleOrder))
	for i, r := range m.ruleOrder {
		ruleProbs[i] = m.ruleProb(r)
	}
	rk := m.ruleOrder[DrawIndex(p, Cumulative(ruleProbs))]
	lengths := m.retainedRules[rk]
	var glyphs []string
	for slot := range lengths {
		keep := m.slotKeep[rk][slot]
		var outs []string
		for s := range keep {
			outs = append(outs, s)
		}
		sort.Strings(outs)
		if len(outs) == 0 {
			return GeneratedToken{NonGenerative: true}
		}
		probs := make([]float64, len(outs))
		for i, s := range outs {
			probs[i] = m.slotProb(rk, slot, s)
		}
		comp := outs[DrawIndex(p, Cumulative(probs))]
		glyphs = append(glyphs, splitGlyphs(comp)...)
	}
	if len(glyphs) == 0 {
		return GeneratedToken{NonGenerative: true}
	}
	return GeneratedToken{Glyphs: glyphs}
}

// Complexity: per-slot COMPONENT inventories are LexiconCost (explicit
// GRAMMAR_COMPLEXITY_CONTRACT.md worked example); the exception table is
// ExceptionCost; the exception/rule split, rule choice, and per-slot
// choices are StructureCost choice points. Alpha=0.5 is a fixed algorithm
// constant, not a per-candidate fitted hyperparameter, so it is not
// charged as a free real parameter.
func (m *M5Model) Complexity() ComplexityBreakdown {
	structure := log2(2) // exception-vs-rule
	if len(m.ruleOrder) > 0 {
		structure += log2(float64(len(m.ruleOrder)))
	}
	lex := 0.0
	components := 0
	for _, rk := range m.ruleOrder {
		for slot, keep := range m.slotKeep[rk] {
			k := len(keep)
			if k > 0 {
				structure += log2(float64(k))
			}
			for s := range keep {
				c := m.slotCount2[rk][slot][s]
				if c > 0 {
					lex += -log2(float64(c) / float64(m.nDev))
					components++
				}
			}
		}
	}
	exc := 0.0
	for _, k := range m.exceptionOrder {
		c := m.exceptionCount[k]
		exc += -log2(float64(c)/float64(m.nDev)) + 1
	}
	return ComplexityBreakdown{
		StructureCost: structure, LexiconCost: lex, ExceptionCost: exc,
		FreeParams: 0, Rules: len(m.ruleOrder), Components: components,
	}
}
