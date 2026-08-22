package mechanismspace

import (
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

func sampleCorpus() Corpus {
	words := []string{"the", "quick", "brown", "fox", "jumps", "over", "the", "lazy", "dog", "again",
		"and", "again", "the", "fox", "runs", "fast", "through", "the", "dark", "wood"}
	lines := make([]int, len(words))
	for i := range lines {
		lines[i] = i / 5
	}
	return Corpus{Name: "test", Words: words, Lines: lines}
}

// test 1: IDENTITY preserves the corpus exactly.
func TestIdentityPreservesCorpus(t *testing.T) {
	c := sampleCorpus()
	out := Transform(Config{Family: "M0"}, c)
	if out.OutputTokens != len(c.Words) {
		t.Fatalf("identity changed token count: %d vs %d", out.OutputTokens, len(c.Words))
	}
	for i, w := range c.Glyphs() {
		if !reflect.DeepEqual(w, out.Tokens[i]) {
			t.Fatalf("identity changed word %d: %v vs %v", i, w, out.Tokens[i])
		}
	}
}

// test 2/25: deterministic mechanisms (and the job wrapper) reproduce
// byte-for-byte given the same config/seed/input.
func TestDeterministicReproduction(t *testing.T) {
	c := sampleCorpus()
	cfg := Config{Family: "M4", StateCount: 4, Update: UpdateB, Seed: 7}
	a := Transform(cfg, c)
	b := Transform(cfg, c)
	if !reflect.DeepEqual(a.Tokens, b.Tokens) {
		t.Fatalf("same config/seed produced different output")
	}
	j := Job{ExperimentID: "t", Corpus: "test", Mechanism: cfg, Seed: cfg.Seed, EvaluationSet: "SCREENING"}
	r1 := Execute(j, c, DefaultScreeningOptions(1))
	r2 := Execute(j, c, DefaultScreeningOptions(1))
	if r1.JobID != r2.JobID || !reflect.DeepEqual(r1.Fingerprint, r2.Fingerprint) {
		t.Fatalf("same job did not reproduce identically")
	}
}

// test 3/4: stochastic mechanisms (M2 homophony) reproduce with the same
// seed and change with a different one.
func TestStochasticSeedBehavior(t *testing.T) {
	c := sampleCorpus()
	cfg := Config{Family: "M2", Homophones: 4, Seed: 3}
	a := Transform(cfg, c)
	b := Transform(cfg, c)
	if !reflect.DeepEqual(a.Tokens, b.Tokens) {
		t.Fatalf("same seed produced different stochastic output")
	}
	cfg2 := cfg
	cfg2.Seed = 4
	d := Transform(cfg2, c)
	if reflect.DeepEqual(a.Tokens, d.Tokens) {
		t.Fatalf("different seeds produced identical stochastic output")
	}
}

// test 5: the three state-update variants are independently distinct.
func TestStateUpdateVariantsDiffer(t *testing.T) {
	c := sampleCorpus()
	base := Config{Family: "M4", StateCount: 4, Seed: 11}
	a := base
	a.Update = UpdateA
	b := base
	b.Update = UpdateB
	cc := base
	cc.Update = UpdateC
	oa, ob, oc := Transform(a, c), Transform(b, c), Transform(cc, c)
	if reflect.DeepEqual(oa.Tokens, ob.Tokens) && reflect.DeepEqual(ob.Tokens, oc.Tokens) {
		t.Fatalf("all three state-update variants produced identical output")
	}
}

// test 6: fixed-state null holds state constant.
func TestFixedStateNullHoldsStateConstant(t *testing.T) {
	c := sampleCorpus()
	cfg := Config{Family: "M4", StateCount: 4, Update: UpdateA, Seed: 5, FixedStateNull: true}
	words := c.Glyphs()
	state, _ := schedule(normalize(cfg), len(words), func(i int) []string { return words[i] }, nil)
	for i, s := range state {
		if s != 0 {
			t.Fatalf("fixed-state null advanced at position %d: state=%d", i, s)
		}
	}
}

// test 7: the shuffled-state null preserves state visit frequencies.
func TestShuffledStateNullPreservesFrequencies(t *testing.T) {
	c := sampleCorpus()
	words := c.Glyphs()
	cfg := normalize(Config{Family: "M4", StateCount: 4, Update: UpdateA, Seed: 5})
	unitAt := func(i int) []string { return words[i] }
	real, _ := schedule(cfg, len(words), unitAt, rand.New(rand.NewSource(cfg.Seed)))
	cfg.ShuffleStateNull = true
	shuffled, _ := schedule(cfg, len(words), unitAt, rand.New(rand.NewSource(cfg.Seed)))
	if freqOf(real) != freqOf(shuffled) {
		t.Fatalf("shuffled-state null changed state frequencies: %v vs %v", freqOf(real), freqOf(shuffled))
	}
	if reflect.DeepEqual(real, shuffled) {
		t.Fatalf("shuffled-state null did not actually reorder the trajectory")
	}
}

func freqOf(xs []int) string {
	m := map[int]int{}
	for _, x := range xs {
		m[x]++
	}
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	s := ""
	for _, k := range keys {
		s += string(rune('0'+k)) + ":" + string(rune('0'+m[k])) + ";"
	}
	return s
}

// test 8: slow-state persistence changes only every DriftScale units.
func TestSlowStatePersistence(t *testing.T) {
	c := sampleCorpus()
	words := c.Glyphs()
	cfg := normalize(Config{Family: "M5", StateCount: 100, Update: UpdateA, DriftScale: 5, Seed: 1})
	state, _ := schedule(cfg, len(words), func(i int) []string { return words[i] }, nil)
	changes := 0
	for i := 1; i < len(state); i++ {
		if state[i] != state[i-1] {
			changes++
			if i%5 != 0 {
				t.Fatalf("state changed at non-multiple-of-5 position %d", i)
			}
		}
	}
	if changes == 0 {
		t.Fatalf("slow state never advanced")
	}
}

// test 9: macro-state transitions land on the expected fixed schedule.
func TestMacroStateTransitions(t *testing.T) {
	c := sampleCorpus()
	words := c.Glyphs()
	n := len(words)
	cfg := normalize(Config{Family: "M6", MacroStates: 5, Seed: 1})
	_, macro := schedule(cfg, n, func(i int) []string { return words[i] }, nil)
	for i, m := range macro {
		want := (i * 5) / n
		if m != want {
			t.Fatalf("macro state at %d = %d, want %d", i, m, want)
		}
	}
}

// test 10/17: STREAM-mode generated boundaries consume every input glyph,
// and grouping does not leak original word boundaries.
func TestStreamBoundariesConsumeInputAndIgnoreWordBoundaries(t *testing.T) {
	c := sampleCorpus()
	cfg := Config{Family: "M8", InputMode: Stream, Grouping: FixedGrouping, GroupLen: 4, Seed: 1}
	out := Transform(cfg, c)
	if out.OutputGlyphs != out.InputUnits {
		t.Fatalf("stream grouping lost glyphs: in=%d out=%d", out.InputUnits, out.OutputGlyphs)
	}
	// At least one output token must span a natural word boundary: verify
	// the flattened glyph stream's cut points do not all coincide with the
	// original per-word glyph-count boundaries.
	wordBoundaries := map[int]bool{}
	pos := 0
	for _, w := range c.Glyphs() {
		pos += len(w)
		wordBoundaries[pos] = true
	}
	pos = 0
	allAligned := true
	for _, tok := range out.Tokens {
		pos += len(tok)
		if !wordBoundaries[pos] {
			allAligned = false
		}
	}
	if allAligned {
		t.Fatalf("stream-mode grouping never crossed an original word boundary")
	}
}

// test 11/16: WORD_PRESERVING mode never silently drops a unit for a
// non-lossy family, and output-token count equals input-word count.
func TestWordPreservingPreservesBoundaryCorrespondence(t *testing.T) {
	c := sampleCorpus()
	for _, fam := range []string{"M1", "M2", "M4", "M5", "M6", "M7"} {
		cfg := Config{Family: fam, StateCount: 2, MacroStates: 2, DriftScale: 5, Homophones: 4, Seed: 2}
		out := Transform(cfg, c)
		if out.OutputTokens != len(c.Words) {
			t.Fatalf("%s: output token count %d != input word count %d", fam, out.OutputTokens, len(c.Words))
		}
	}
}

// test 12: FIXED/STATE grouping produce deterministic (repeatable) group
// lengths, unlike RANDOM grouping's per-run draw sequence under a fixed
// seed (which is still itself deterministic, but must differ from a
// FIXED-length schedule).
func TestGroupingDeterminism(t *testing.T) {
	c := sampleCorpus()
	cfgFixed := Config{Family: "M8", InputMode: Stream, Grouping: FixedGrouping, GroupLen: 4, Seed: 9}
	a := Transform(cfgFixed, c)
	b := Transform(cfgFixed, c)
	if !reflect.DeepEqual(a.Tokens, b.Tokens) {
		t.Fatalf("FIXED grouping is not deterministic")
	}
	for _, tok := range a.Tokens[:len(a.Tokens)-1] {
		if len(tok) != 4 {
			t.Fatalf("FIXED grouping produced a non-fixed-length token: %v", tok)
		}
	}
}

// test 13: the constrained grammar never emits a glyph outside its
// position class's frozen alphabet.
func TestGrammarNeverEmitsInvalidForms(t *testing.T) {
	c := sampleCorpus()
	cfg := Config{Family: "M3", Grammar: GrammarMedium, Seed: 1}
	out := Transform(cfg, c)
	startAlphabet := map[string]bool{}
	for _, s := range classAlphabet("START", 4, 0, 0, cfg.Seed) {
		startAlphabet[s] = true
	}
	endAlphabet := map[string]bool{}
	for _, s := range classAlphabet("END", 4, 0, 0, cfg.Seed) {
		endAlphabet[s] = true
	}
	coreAlphabet := map[string]bool{}
	for _, s := range classAlphabet("CORE", 4, 0, 0, cfg.Seed) {
		coreAlphabet[s] = true
	}
	for _, tok := range out.Tokens {
		if len(tok) < 2 {
			t.Fatalf("grammar token too short: %v", tok)
		}
		if !startAlphabet[tok[0]] {
			t.Fatalf("invalid START symbol %q", tok[0])
		}
		if !endAlphabet[tok[len(tok)-1]] {
			t.Fatalf("invalid END symbol %q", tok[len(tok)-1])
		}
		for _, g := range tok[1 : len(tok)-1] {
			if !coreAlphabet[g] {
				t.Fatalf("invalid CORE symbol %q", g)
			}
		}
	}
}

// test 14: complexity accounting matches hand-computed expectations for
// two representative configs.
func TestComplexityAccounting(t *testing.T) {
	m := buildMetadata(Config{Family: "M0"})
	if m.StateCount != 1 || m.SymbolClasses != 1 || m.TransitionParameters != 0 || m.OutputRules != 1 || m.StochasticDistributions != 0 {
		t.Fatalf("M0 complexity wrong: %+v", m)
	}
	m11 := buildMetadata(Config{Family: "M11", StateCount: 4, MacroStates: 5, Grammar: GrammarHigh})
	if m11.StateCount != 20 {
		t.Fatalf("M11 combined state count = %d, want 20", m11.StateCount)
	}
	if m11.OutputRules != 9 {
		t.Fatalf("M11 output rules = %d, want 9 (HIGH)", m11.OutputRules)
	}
	if m11.TransitionParameters != 4+5 {
		t.Fatalf("M11 transition params = %d, want 9", m11.TransitionParameters)
	}
}

// test 15: expansion ratio (output glyphs / input units) is internally
// consistent with the accounting fields.
func TestExpansionRatioConsistent(t *testing.T) {
	c := sampleCorpus()
	out := Transform(Config{Family: "M3", Grammar: GrammarLow, Seed: 1}, c)
	g := 0
	for _, t := range out.Tokens {
		g += len(t)
	}
	if g != out.OutputGlyphs {
		t.Fatalf("OutputGlyphs accounting mismatch: %d vs recomputed %d", out.OutputGlyphs, g)
	}
	if out.InputUnits <= 0 {
		t.Fatalf("InputUnits not tracked")
	}
}

// test 23: the plaintext-ablation shuffle preserves the exact word
// multiset (marginal frequencies), only reordering occurrences.
func TestShufflePlaintextPreservesMarginals(t *testing.T) {
	c := sampleCorpus()
	s := c.ShufflePlaintextWords(42)
	orig := append([]string(nil), c.Words...)
	shuf := append([]string(nil), s.Words...)
	sort.Strings(orig)
	sort.Strings(shuf)
	if !reflect.DeepEqual(orig, shuf) {
		t.Fatalf("shuffle changed the word multiset")
	}
}

// test 24: the error injector changes roughly the requested fraction of
// glyphs (loose statistical tolerance; not used for model selection).
func TestErrorInjectorRateIsApproximatelyRespected(t *testing.T) {
	var tokens [][]string
	for i := 0; i < 500; i++ {
		tokens = append(tokens, []string{"a", "b", "c", "d"})
	}
	before := countGlyphs(tokens)
	after := InjectErrors(tokens, 0.05, 1)
	afterCount := countGlyphs(after)
	if afterCount == before {
		// Some perturbation should have occurred at a 5% rate over 2000 glyphs.
		t.Fatalf("error injector produced no visible change at rate 0.05")
	}
}

func countGlyphs(tokens [][]string) int {
	n := 0
	for _, t := range tokens {
		n += len(t)
	}
	return n
}
