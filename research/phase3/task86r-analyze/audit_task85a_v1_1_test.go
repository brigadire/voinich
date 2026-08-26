package main

import (
	"fmt"
	"math/bits"
	"sort"
	"strings"
	"testing"
	"unicode/utf8"
)

// auditFSAResult is intentionally test-only: it implements the all-pairs
// procedure frozen in G1_EXECUTABLE_CONTRACT.json without changing Task86R.
type auditFSAResult struct {
	fsa        *FSA
	operations int
}

func induceFSAAllPairsAudit(occ []TokenOccurrence, threshold float64, maxStates int) auditFSAResult {
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
	find := func(x int) int { return x }
	find = func(x int) int {
		for parent[x] != x {
			parent[x] = parent[parent[x]]
			x = parent[x]
		}
		return x
	}
	pass := make([]int, len(nodes))
	for i := len(nodes) - 1; i >= 0; i-- {
		pass[i] = nodes[i].eos
		for _, c := range nodes[i].children {
			pass[i] += pass[c]
		}
	}
	edges := make([]map[string]fsaEdge, len(nodes))
	eos := make([]int, len(nodes))
	total := make([]int, len(nodes))
	for i, n := range nodes {
		edges[i] = map[string]fsaEdge{}
		eos[i], total[i] = n.eos, pass[i]
		for g, c := range n.children {
			edges[i][g] = fsaEdge{target: c, count: pass[c]}
		}
	}
	dist := func(r int) []float64 {
		out := make([]float64, len(alphabet)+1)
		// alphabet glyphs plus EOS gives A outcomes. Unlike Task86R's
		// historical A+1 denominator, this is a normalized additive-0.5
		// distribution as required by the frozen contract.
		denom := float64(total[r]) + 0.5*A
		for i, g := range alphabet {
			c := 0
			if e, ok := edges[r][g]; ok {
				c = e.count
			}
			out[i] = (float64(c) + 0.5) / denom
		}
		out[len(alphabet)] = (float64(eos[r]) + 0.5) / denom
		return out
	}
	type pair struct{ a, b int }
	ops := 0
	plan := func(x0, y0 int) ([]pair, bool) { return nil, false }
	plan = func(x0, y0 int) ([]pair, bool) {
		queue := []pair{{find(x0), find(y0)}}
		seen := map[[2]int]bool{}
		var out []pair
		for len(queue) > 0 {
			p := queue[0]
			queue = queue[1:]
			a, b := find(p.a), find(p.b)
			if a == b {
				continue
			}
			key := [2]int{a, b}
			if key[0] > key[1] {
				key[0], key[1] = key[1], key[0]
			}
			if seen[key] {
				continue
			}
			seen[key] = true
			ops++
			if ops > inductionOpCap || (eos[a] > 0) != (eos[b] > 0) {
				return nil, false
			}
			out = append(out, pair{a, b})
			labels := map[string]bool{}
			for g := range edges[a] {
				labels[g] = true
			}
			for g := range edges[b] {
				labels[g] = true
			}
			for _, g := range sortedBoolKeys(labels) {
				ea, oka := edges[a][g]
				eb, okb := edges[b][g]
				if oka && okb {
					queue = append(queue, pair{ea.target, eb.target})
				}
			}
		}
		return out, true
	}
	union := func(a, b int) {}
	union = func(a, b int) {
		a, b = find(a), find(b)
		if a == b {
			return
		}
		// The lexically earlier access representative survives.
		if len(nodes[b].access) < len(nodes[a].access) || (len(nodes[b].access) == len(nodes[a].access) && nodes[b].access < nodes[a].access) {
			a, b = b, a
		}
		parent[b] = a
		eos[a] += eos[b]
		total[a] += total[b]
		for g, e := range edges[b] {
			if ex, ok := edges[a][g]; ok {
				ex.count += e.count
				edges[a][g] = ex
			} else {
				edges[a][g] = e
			}
		}
	}

	failed, failWhy := false, ""
	for {
		repsSet := map[int]bool{}
		for i := range nodes {
			repsSet[find(i)] = true
		}
		reps := make([]int, 0, len(repsSet))
		for r := range repsSet {
			reps = append(reps, r)
		}
		sort.Slice(reps, func(i, j int) bool {
			ai, aj := nodes[reps[i]].access, nodes[reps[j]].access
			if len(ai) != len(aj) {
				return len(ai) < len(aj)
			}
			return ai < aj
		})
		merged := false
		for i := 0; i < len(reps) && !merged; i++ {
			for j := i + 1; j < len(reps); j++ {
				ops++
				if ops > inductionOpCap {
					failed, failWhy = true, "TRAINING_FAILED: induction operation cap exceeded"
					break
				}
				if jsDivergence(dist(reps[i]), dist(reps[j])) > threshold {
					continue
				}
				closure, ok := plan(reps[i], reps[j])
				if ops > inductionOpCap {
					failed, failWhy = true, "TRAINING_FAILED: induction operation cap exceeded"
					break
				}
				if !ok {
					continue
				}
				for _, p := range closure {
					union(p.a, p.b)
				}
				merged = true
				break
			}
			if failed {
				break
			}
		}
		if failed || !merged {
			break
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
		return auditFSAResult{fsa: fsa, operations: ops}
	}
	repSet := map[int]bool{}
	for i := range nodes {
		repSet[find(i)] = true
	}
	for r := range repSet {
		fsa.States = append(fsa.States, r)
	}
	sort.Ints(fsa.States)
	for _, r := range fsa.States {
		fsa.Accept[r] = eos[r]
		fsa.Edges[r] = map[string]fsaEdge{}
		for g, e := range edges[r] {
			fsa.Edges[r][g] = fsaEdge{target: find(e.target), count: e.count}
		}
	}
	return auditFSAResult{fsa: fsa, operations: ops}
}

func auditAccepts(f *FSA, token []string) bool {
	if f.Failed {
		return false
	}
	s := f.Root
	for _, g := range token {
		e, ok := f.Edges[s][g]
		if !ok {
			return false
		}
		s = e.target
	}
	return f.Accept[s] > 0
}

func auditLanguage(f *FSA, alphabet []string, maxLen int) string {
	var accepted []string
	var walk func([]string, int)
	walk = func(prefix []string, left int) {
		if len(prefix) > 0 && auditAccepts(f, prefix) {
			accepted = append(accepted, strings.Join(prefix, ""))
		}
		if left == 0 {
			return
		}
		for _, g := range alphabet {
			walk(append(append([]string(nil), prefix...), g), left-1)
		}
	}
	walk(nil, maxLen)
	return strings.Join(accepted, ",")
}

func syntheticOccurrences(tokens []string) []TokenOccurrence {
	out := make([]TokenOccurrence, len(tokens))
	for i, token := range tokens {
		glyphs := make([]string, 0, len(token))
		for _, r := range token {
			glyphs = append(glyphs, string(r))
		}
		out[i] = TokenOccurrence{Raw: token, Glyphs: glyphs, Partition: "DEVELOPMENT"}
	}
	return out
}

func TestTask85aV11StateMergingSyntheticCounterexample(t *testing.T) {
	universe := []string{"a", "b", "aa", "ab", "ba", "bb", "aaa", "aab", "aba", "abb", "baa", "bab", "bba", "bbb"}
	thresholds := []float64{0, 0.05, 0.1, 0.5, 1}
	for mask := 1; mask < 1<<len(universe); mask++ {
		var corpus []string
		for i, token := range universe {
			if mask&(1<<i) != 0 {
				corpus = append(corpus, token)
			}
		}
		occ := syntheticOccurrences(corpus)
		for _, threshold := range thresholds {
			blue := InduceFSA(occ, threshold, 256)
			exhaustive := induceFSAAllPairsAudit(occ, threshold, 256)
			blueLang := auditLanguage(blue, []string{"a", "b"}, 4)
			exhaustiveLang := auditLanguage(exhaustive.fsa, []string{"a", "b"}, 4)
			if blue.Failed != exhaustive.fsa.Failed || len(blue.States) != len(exhaustive.fsa.States) || blueLang != exhaustiveLang {
				t.Logf("COUNTEREXAMPLE corpus=%s threshold=%g blue_failed=%t exhaustive_failed=%t blue_states=%d exhaustive_states=%d exhaustive_operations=%d blue_language=%q exhaustive_language=%q", strings.Join(corpus, ","), threshold, blue.Failed, exhaustive.fsa.Failed, len(blue.States), len(exhaustive.fsa.States), exhaustive.operations, blueLang, exhaustiveLang)
				return
			}
		}
	}
	t.Fatal("no counterexample found in deterministic bounded suite")
}

func TestTask85aV11StateMergingDevelopmentSensitivity(t *testing.T) {
	split, err := loadSplitPartitions("../task85/GRAMMAR_CORPUS_SPLIT.tsv")
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range corpusSources {
		// Tests run with package directory as cwd; adjust repository-relative paths.
		source.canonical = "../../../" + source.canonical
		source.ivtff = "../../../" + source.ivtff
		corpus, err := loadTranscription(source, split)
		if err != nil {
			t.Fatal(err)
		}
		dev := corpus.Partition("DEVELOPMENT")
		for _, threshold := range []float64{0, 0.05, 0.1} {
			r := induceFSAAllPairsAudit(dev, threshold, 256)
			t.Logf("DEVELOPMENT transcription=%s threshold=%g max_states=256 failed=%t reason=%q operations=%d", corpus.Name, threshold, r.fsa.Failed, r.fsa.FailWhy, r.operations)
			if !r.fsa.Failed || !strings.Contains(r.fsa.FailWhy, "operation cap") {
				t.Errorf("expected exhaustive audit to hit frozen cap for %s threshold %g", corpus.Name, threshold)
			}
		}
	}
}

func TestTask85aV11RNGVectors(t *testing.T) {
	for _, seed := range []uint64{0, 1, 42, ^uint64(0)} {
		p := NewPRNG(seed)
		values := make([]string, 8)
		for i := range values {
			values[i] = fmt.Sprintf("%016x", p.Uint64())
		}
		t.Logf("RNG seed=%d outputs=%s", seed, strings.Join(values, ","))
	}
	// A no-warm-up stream is a contract-consistent counterfactual because the
	// frozen text specifies only two SplitMix outputs as state.
	seed := uint64(42)
	sm := &splitMix64{state: seed}
	noWarm := &PRNG{state: u128{hi: sm.next(), lo: sm.next()}}
	actual := NewPRNG(seed)
	if actual.Uint64() == noWarm.Uint64() {
		t.Fatal("warm-up counterfactual unexpectedly identical")
	}
}

func TestTask85aV11GlyphAliasRoundTrip(t *testing.T) {
	split, err := loadSplitPartitions("../task85/GRAMMAR_CORPUS_SPLIT.tsv")
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range corpusSources {
		source.canonical = "../../../" + source.canonical
		source.ivtff = "../../../" + source.ivtff
		corpus, err := loadTranscription(source, split)
		if err != nil {
			t.Fatal(err)
		}
		alphabet := glyphAlphabet(corpus.Occurrences)
		alias := NewGlyphAlias(alphabet)
		reverse := map[rune]string{}
		for glyph, r := range alias.toRune {
			if prior, exists := reverse[r]; exists {
				t.Fatalf("alias collision: %q and %q", prior, glyph)
			}
			reverse[r] = glyph
		}
		for _, glyph := range alphabet {
			encoded := alias.Encode([]string{glyph})
			r, n := utf8.DecodeRuneInString(encoded)
			if n != len(encoded) || reverse[r] != glyph {
				t.Fatalf("round trip failed for %s %q", corpus.Name, glyph)
			}
		}
		t.Logf("GLYPH_ROUNDTRIP transcription=%s alphabet=%d collisions=0 failures=0", corpus.Name, len(alphabet))
	}
}

func TestTask85aV11GlyphAliasF2Regression(t *testing.T) {
	split, err := loadSplitPartitions("../task85/GRAMMAR_CORPUS_SPLIT.tsv")
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range corpusSources {
		source.canonical = "../../../" + source.canonical
		source.ivtff = "../../../" + source.ivtff
		corpus, err := loadTranscription(source, split)
		if err != nil {
			t.Fatal(err)
		}
		held := corpus.Partition("HELDOUT")
		var representative [][]string
		for _, o := range held {
			representative = append(representative, o.Glyphs)
			if len(representative) == 512 {
				break
			}
		}
		alphabet := glyphAlphabet(held)
		primary := NewGlyphAlias(alphabet)
		pool := aliasPool()
		alternate := &GlyphAlias{toRune: map[string]rune{}}
		keys := make([]string, 0, len(primary.toRune))
		for g := range primary.toRune {
			keys = append(keys, g)
		}
		sort.Strings(keys)
		for i, g := range keys {
			alternate.toRune[g] = pool[len(pool)-1-i]
		}
		seed := int64(8675309)
		gotPrimary, okPrimary, err := StructuralMetrics(primary, representative, seed, "/tmp")
		if err != nil || !okPrimary {
			t.Fatalf("primary F2 %s: valid=%t err=%v", corpus.Name, okPrimary, err)
		}
		gotAlternate, okAlternate, err := StructuralMetrics(alternate, representative, seed, "/tmp")
		if err != nil || !okAlternate {
			t.Fatalf("alternate F2 %s: valid=%t err=%v", corpus.Name, okAlternate, err)
		}
		for _, metric := range StructuralMetricIDs {
			delta := gotPrimary[metric] - gotAlternate[metric]
			if delta < 0 {
				delta = -delta
			}
			t.Logf("F2_ALIAS transcription=%s metric=%s primary=%.17g alternate=%.17g abs_delta=%.17g", corpus.Name, metric, gotPrimary[metric], gotAlternate[metric], delta)
			if delta > 1e-12 {
				t.Errorf("label-renaming changed %s/%s by %g", corpus.Name, metric, delta)
			}
		}

		// Direct natural-mode regression on the frozen HELDOUT subset for which
		// every frozen glyph is already exactly one rune.
		var directPopulation [][]string
		directAlphabet := map[string]bool{}
		for _, o := range held {
			safe := true
			for _, g := range o.Glyphs {
				if utf8.RuneCountInString(g) != 1 || strings.ToLower(g) != g {
					safe = false
					break
				}
			}
			if safe {
				directPopulation = append(directPopulation, o.Glyphs)
				for _, g := range o.Glyphs {
					directAlphabet[g] = true
				}
			}
			if len(directPopulation) == 512 {
				break
			}
		}
		direct := &GlyphAlias{toRune: map[string]rune{}}
		var directKeys []string
		for g := range directAlphabet {
			directKeys = append(directKeys, g)
		}
		sort.Strings(directKeys)
		for _, g := range directKeys {
			r, _ := utf8.DecodeRuneInString(g)
			direct.toRune[g] = r
		}
		encoded := NewGlyphAlias(directKeys)
		gotDirect, directOK, err := StructuralMetrics(direct, directPopulation, seed, "/tmp")
		if err != nil || !directOK {
			t.Fatalf("direct F2 %s: valid=%t err=%v", corpus.Name, directOK, err)
		}
		gotEncoded, encodedOK, err := StructuralMetrics(encoded, directPopulation, seed, "/tmp")
		if err != nil || !encodedOK {
			t.Fatalf("encoded F2 %s: valid=%t err=%v", corpus.Name, encodedOK, err)
		}
		for _, metric := range StructuralMetricIDs {
			delta := gotDirect[metric] - gotEncoded[metric]
			if delta < 0 {
				delta = -delta
			}
			t.Logf("F2_DIRECT transcription=%s metric=%s direct=%.17g encoded=%.17g abs_delta=%.17g tokens=%d", corpus.Name, metric, gotDirect[metric], gotEncoded[metric], delta, len(directPopulation))
			if delta > 1e-12 {
				t.Errorf("direct representation changed %s/%s by %g", corpus.Name, metric, delta)
			}
		}
	}
}

func TestTask85aV11PM6SingletonExhaustion(t *testing.T) {
	split, err := loadSplitPartitions("../task85/GRAMMAR_CORPUS_SPLIT.tsv")
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range corpusSources {
		source.canonical = "../../../" + source.canonical
		source.ivtff = "../../../" + source.ivtff
		corpus, err := loadTranscription(source, split)
		if err != nil {
			t.Fatal(err)
		}
		observedSingleton := map[string]bool{}
		heldSingleton := 0
		for _, o := range corpus.Occurrences {
			if len(o.Glyphs) == 1 {
				observedSingleton[o.Glyphs[0]] = true
				if o.Partition == "HELDOUT" {
					heldSingleton++
				}
			}
		}
		devAlphabet := glyphAlphabet(corpus.Partition("DEVELOPMENT"))
		missing := []string{}
		for _, g := range devAlphabet {
			if !observedSingleton[g] {
				missing = append(missing, g)
			}
		}
		t.Logf("PM6_SINGLETON transcription=%s heldout_singletons=%d development_alphabet=%d observed_singleton_types=%d missing_development_glyph_singletons=%q", corpus.Name, heldSingleton, len(devAlphabet), len(observedSingleton), strings.Join(missing, ","))
		// Every negative must be unique, and a one-glyph negative can only be
		// one of these not-yet-observed DEVELOPMENT-alphabet glyphs. The class
		// restriction can only reduce this upper bound.
		if heldSingleton == 0 || heldSingleton <= len(missing) {
			t.Errorf("singleton impossibility proof failed for %s", corpus.Name)
		}
	}
}

func TestMul128AuditSanity(t *testing.T) {
	// Guard the copied audit environment against an accidental architecture
	// assumption while retaining a reference to math/bits in this test file.
	hi, lo := bits.Mul64(2, 3)
	if hi != 0 || lo != 6 {
		t.Fatal("unexpected uint64 multiplication")
	}
}
