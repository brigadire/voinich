package mechanismspace

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"math/rand"
	"sort"

	"zcore.dev/voinich/internal/evaglyph"
)

// InputMode is task66 section 27's central WORD_PRESERVING/STREAM choice.
type InputMode string

const (
	WordPreserving InputMode = "WORD_PRESERVING"
	Stream         InputMode = "STREAM"
)

// StateUpdate is one of task66 section 17's three state-update variants.
// They are never mixed within one mechanism.
type StateUpdate string

const (
	UpdateA StateUpdate = "A" // S_{i+1} = F(S_i)
	UpdateB StateUpdate = "B" // S_{i+1} = F(S_i, P_i)
	UpdateC StateUpdate = "C" // S_{i+1} = F(S_i, C_i)
)

// Grouping is task66 section 22's boundary-generation rule.
type Grouping string

const (
	NoGrouping     Grouping = ""
	FixedGrouping  Grouping = "FIXED"
	RandomGrouping Grouping = "RANDOM"
	StateGrouping  Grouping = "STATE"
)

// GrammarLevel is task66 section 15's frozen complexity grid.
type GrammarLevel string

const (
	NoGrammar     GrammarLevel = ""
	GrammarLow    GrammarLevel = "LOW"
	GrammarMedium GrammarLevel = "MEDIUM"
	GrammarHigh   GrammarLevel = "HIGH"
)

// InformationStatus is task66 section 28's four-way classification.
type InformationStatus string

const (
	Reversible             InformationStatus = "REVERSIBLE"
	ReversibleWithKeyState InformationStatus = "REVERSIBLE_WITH_KEY_STATE"
	Lossy                  InformationStatus = "LOSSY"
	Ambiguous              InformationStatus = "AMBIGUOUS"
)

// Config is one immutable, self-contained mechanism job configuration
// (task66 section 33's frozen parameter grid plus the null-control flags
// of sections 55-58 and the plaintext ablation of section 65). Two Configs
// with the same field values always produce the same Hash and, given the
// same input corpus, the same output (task66 test 25).
type Config struct {
	Family      string // M0..M11
	InputMode   InputMode
	StateCount  int          // K for M4-M7, M10-M11 (1 = memoryless)
	MacroStates int          // K for M6/M7/M11 (1 = no macro state)
	DriftScale  int          // N: state advances every N units (M5/M7/M10/M11)
	Update      StateUpdate  // state-update variant for M4 (A/B/C)
	Grammar     GrammarLevel // M3/M9/M10/M11
	Grouping    Grouping     // M8/M9/M10/M11
	GroupLen    int          // mean/fixed group length for Grouping
	Homophones  int          // H for M2
	Seed        int64

	// Null controls (sections 55-58): applied as a post-processing
	// perturbation of the schedule, never combined with each other.
	ShuffleStateNull   bool
	FixedStateNull     bool
	FastStateNull      bool
	RandomBoundaryNull bool
	ShufflePlaintext   bool // section 65 ablation
}

// Hash is the config-hash used as the immutable job's identity (task66
// section 74/final paragraph: experiment_id/corpus/mechanism/config_hash/
// seed/evaluation_set).
func (c Config) Hash() string {
	s := fmt.Sprintf("%+v", c)
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// Metadata is task66 section 75's required Transformer.Metadata() output.
type Metadata struct {
	Family                  string
	Parameters              Config
	Seed                    int64
	StateCount              int
	Information             InformationStatus
	SymbolClasses           int
	TransitionParameters    int
	OutputRules             int
	StochasticDistributions int
}

// Output is one mechanism's transformed corpus plus its accounting
// (task66 section 30: input/output symbol and word/token counts).
type Output struct {
	Tokens                                             [][]string
	Lines                                              []int
	InputUnits, OutputGlyphs, InputWords, OutputTokens int
	Metadata                                           Metadata
}

// Transform runs one immutable mechanism job against one plaintext corpus.
// It never reads a Voynich target value: it is a pure function of (Config,
// Corpus) (task66 sections 3-4).
func Transform(cfg Config, c Corpus) Output {
	cfg = normalize(cfg)
	meta := buildMetadata(cfg)

	if cfg.ShufflePlaintext {
		if cfg.InputMode == WordPreserving {
			c = c.ShufflePlaintextWords(cfg.Seed + 900001)
		}
	}

	inputWords := c.Glyphs()
	inputUnits := 0
	for _, w := range inputWords {
		inputUnits += len(w)
	}

	if cfg.Family == "M0" {
		return finish(inputWords, c.Lines, inputUnits, len(inputWords), meta)
	}

	r := rand.New(rand.NewSource(cfg.Seed))

	switch cfg.InputMode {
	case Stream:
		var glyphs []string
		if cfg.ShufflePlaintext {
			glyphs = c.ShufflePlaintextGlyphs(cfg.Seed + 900002)
		} else {
			for _, w := range c.Words {
				glyphs = append(glyphs, evaglyph.NaturalGlyphs(w)...)
			}
		}
		toks := transformStream(cfg, glyphs, r)
		return finish(toks, nil, len(glyphs), len(inputWords), meta)
	default:
		var perm map[string]string
		if cfg.Family == "M1" {
			perm = monoalphabeticPermutation(distinctGlyphs(inputWords), cfg.Seed)
		}
		toks := transformWords(cfg, inputWords, perm, r)
		return finish(toks, c.Lines, inputUnits, len(inputWords), meta)
	}
}

// distinctGlyphs returns the sorted distinct glyph alphabet observed in
// words, for building M1's monoalphabetic permutation (task66 section
// 12): a genuine bijection needs a table sized to the actual alphabet,
// not the hash-based approximation used elsewhere, or a small alphabet
// can suffer hash collisions and stop being a true relabeling.
func distinctGlyphs(words [][]string) []string {
	seen := map[string]bool{}
	for _, w := range words {
		for _, g := range w {
			seen[g] = true
		}
	}
	out := make([]string, 0, len(seen))
	for g := range seen {
		out = append(out, g)
	}
	sort.Strings(out)
	return out
}

func monoalphabeticPermutation(alphabet []string, seed int64) map[string]string {
	perm := append([]string(nil), alphabet...)
	r := rand.New(rand.NewSource(seed))
	r.Shuffle(len(perm), func(i, j int) { perm[i], perm[j] = perm[j], perm[i] })
	m := make(map[string]string, len(alphabet))
	for i, g := range alphabet {
		m[g] = perm[i]
	}
	return m
}

func normalize(c Config) Config {
	if c.InputMode == "" {
		c.InputMode = WordPreserving
	}
	if c.StateCount < 1 {
		c.StateCount = 1
	}
	if c.MacroStates < 1 {
		c.MacroStates = 1
	}
	if c.DriftScale < 1 {
		c.DriftScale = 1
	}
	if c.Update == "" {
		c.Update = UpdateA
	}
	if c.GroupLen < 1 {
		c.GroupLen = 4
	}
	return c
}

func finish(toks [][]string, lines []int, inputUnits, inputWords int, meta Metadata) Output {
	g, ot := 0, len(toks)
	for _, t := range toks {
		g += len(t)
	}
	return Output{Tokens: toks, Lines: lines, InputUnits: inputUnits, OutputGlyphs: g, InputWords: inputWords, OutputTokens: ot, Metadata: meta}
}

// schedule produces the per-unit (state, macro) pair for n units under
// cfg, honoring the null-control overrides (task66 sections 55-57).
// Macro state follows a fixed, content-independent block schedule
// (section 19: it changes parameters, it does not itself evolve
// token-by-token). Local state follows cfg.Update's causal rule
// (section 17): UpdateC's C_i is computable inline as soon as state[i]
// is known (C_i = G(P_i, state[i], macro[i])), so no separate output
// pre-pass is needed even for that variant.
func schedule(cfg Config, n int, unitAt func(i int) []string, r *rand.Rand) (state []int, macro []int) {
	state = make([]int, n)
	macro = make([]int, n)
	effScale := max1(cfg.DriftScale)
	if cfg.FastStateNull {
		effScale = 1
	}
	s := 0
	for i := 0; i < n; i++ {
		if cfg.MacroStates > 1 {
			macro[i] = (i * cfg.MacroStates) / n
		}
		state[i] = s
		if cfg.FixedStateNull {
			continue
		}
		// The periodicity gate (every effScale units, task66 section 18's
		// slow-state-drift scale) is common to all three variants; what
		// differs between A/B/C (section 17) is the increment: A's is
		// fixed (content-independent), B's depends on the input unit
		// P_i, C's depends on the emitted output C_i. Using content only
		// to size the increment - rather than to gate whether an advance
		// happens at all - keeps the three variants distinguishable even
		// when effScale==1 (M4's default, no drift).
		if (i+1)%effScale == 0 && cfg.StateCount > 1 {
			inc := 1
			switch cfg.Update {
			case UpdateB:
				inc = 1 + int(hashUnit(unitAt(i), cfg.Seed)%uint32(cfg.StateCount-1))
			case UpdateC:
				probe := mapGlyphs(unitAt(i), state[i], macro[i], cfg.Seed)
				inc = 1 + int(hashUnit(probe, cfg.Seed)%uint32(cfg.StateCount-1))
			}
			s = (s + inc) % cfg.StateCount
		}
	}
	if cfg.ShuffleStateNull {
		state = shuffleInts(state, r)
		macro = shuffleInts(macro, r)
	}
	return state, macro
}

func max1(x int) int {
	if x < 1 {
		return 1
	}
	return x
}

func shuffleInts(x []int, r *rand.Rand) []int {
	y := append([]int(nil), x...)
	r.Shuffle(len(y), func(i, j int) { y[i], y[j] = y[j], y[i] })
	return y
}

func joinTokens(t []string) string {
	s := ""
	for _, g := range t {
		s += g
	}
	return s
}

func hashUnit(u []string, seed int64) uint32 { return hashString(joinTokens(u), seed) }

func hashString(s string, seed int64) uint32 {
	h := fnv.New32a()
	fmt.Fprintf(h, "%d:%s", seed, s)
	return h.Sum32()
}

// substituteGlyph is the compact, content-dependent, seeded substitution
// used by every memoryless/stateful family (M1/M4/M5/M6/M7 and the glyph
// layer beneath M3/M8/M9/M10/M11). It is a formula, not a per-token
// lookup table, so it carries no free per-Voynich-token parameters
// (task66 section 32).
func substituteGlyph(glyph string, state, macro int, seed int64) string {
	h := hashString(glyph, seed+int64(state)*1000003+int64(macro)*7000033)
	return string(rune('a' + int(h%26)))
}

func transformWords(cfg Config, words [][]string, perm map[string]string, r *rand.Rand) [][]string {
	n := len(words)
	state, macro := schedule(cfg, n, func(i int) []string { return words[i] }, r)
	out := make([][]string, n)
	for i, w := range words {
		switch cfg.Family {
		case "M1":
			out[i] = applyPermutation(w, perm)
		case "M2":
			toks2 := evaglyph.RandomHomophony([][]string{w}, max1(cfg.Homophones), r)
			out[i] = toks2[0]
		case "M3":
			out[i] = form(w, 0, 0, cfg.Grammar, cfg.Seed)
		default: // M4, M5, M6, M7 and their nulls
			out[i] = mapGlyphs(w, state[i], macro[i], cfg.Seed)
		}
	}
	return out
}

func applyPermutation(w []string, perm map[string]string) []string {
	out := make([]string, len(w))
	for j, g := range w {
		out[j] = perm[g]
	}
	return out
}

func mapGlyphs(w []string, state, macro int, seed int64) []string {
	out := make([]string, len(w))
	for j, g := range w {
		out[j] = substituteGlyph(g, state, macro, seed)
	}
	return out
}

func transformStream(cfg Config, glyphs []string, r *rand.Rand) [][]string {
	n := len(glyphs)
	lengths := groupLengths(cfg, n, r)
	if cfg.RandomBoundaryNull {
		lengths = shuffleInts(lengths, r)
	}
	var out [][]string
	pos := 0
	groupIdx := 0
	state, macro := schedule(cfg, n, func(i int) []string { return glyphs[i : i+1] }, r)
	for pos < n {
		l := 1
		if groupIdx < len(lengths) {
			l = lengths[groupIdx]
		}
		if l < 1 {
			l = 1
		}
		end := pos + l
		if end > n {
			end = n
		}
		raw := glyphs[pos:end]
		s, m := 0, 0
		if pos < len(state) {
			s, m = state[pos], macro[pos]
		}
		var tok []string
		if cfg.Grammar != NoGrammar {
			tok = form(raw, s, m, cfg.Grammar, cfg.Seed)
		} else {
			tok = mapGlyphs(raw, s, m, cfg.Seed)
		}
		out = append(out, tok)
		pos = end
		groupIdx++
	}
	return out
}

// groupLengths produces the output-group length schedule for STREAM-mode
// boundary generation (task66 section 22). It never reads the empirical
// Voynich token-length sequence (section 22's prohibition); FIXED/RANDOM
// are content-independent, STATE derives length from the same state
// schedule used elsewhere.
func groupLengths(cfg Config, n int, r *rand.Rand) []int {
	var out []int
	switch cfg.Grouping {
	case RandomGrouping:
		for total := 0; total < n; {
			l := 1 + r.Intn(2*cfg.GroupLen-1)
			out = append(out, l)
			total += l
		}
	case StateGrouping:
		s := 0
		for total := 0; total < n; {
			l := 1 + (s % 6)
			out = append(out, l)
			total += l
			s++
		}
	default: // FixedGrouping and the boundary-generation-free families
		for total := 0; total < n; total += cfg.GroupLen {
			out = append(out, cfg.GroupLen)
		}
	}
	return out
}

// form is task66 section 15's minimal constrained-output grammar:
// START -> CORE* -> END, with position-dependent symbol classes and a
// small frozen output alphabet. It is deliberately content-dependent (the
// CORE symbols are a function of the actual input glyphs) so the
// mechanism remains a genuine transformation of the plaintext rather than
// an input-independent generator (task66 sections 64-66); it is also
// deliberately lossy: many distinct inputs of the same length class emit
// the same output form.
func form(in []string, state, macro int, level GrammarLevel, seed int64) []string {
	if len(in) == 0 {
		return nil
	}
	classSize := map[GrammarLevel]int{GrammarLow: 2, GrammarMedium: 4, GrammarHigh: 6}[level]
	if classSize == 0 {
		classSize = 3
	}
	maxCore := map[GrammarLevel]int{GrammarLow: 2, GrammarMedium: 4, GrammarHigh: 6}[level]
	if maxCore == 0 {
		maxCore = 3
	}
	startAlphabet := classAlphabet("START", classSize, state, macro, seed)
	coreAlphabet := classAlphabet("CORE", classSize, state, macro, seed)
	endAlphabet := classAlphabet("END", classSize, state, macro, seed)

	contentHash := hashString(joinTokens(in), seed)
	nCore := 1 + int(contentHash%uint32(maxCore))
	out := make([]string, 0, nCore+2)
	out = append(out, startAlphabet[int(hashString(in[0], seed))%len(startAlphabet)])
	for i := 0; i < nCore; i++ {
		g := in[i%len(in)]
		out = append(out, coreAlphabet[int(hashString(g, seed+int64(i)))%len(coreAlphabet)])
	}
	out = append(out, endAlphabet[int(hashString(in[len(in)-1], seed+1))%len(endAlphabet)])
	return out
}

func classAlphabet(class string, size, state, macro int, seed int64) []string {
	out := make([]string, size)
	for i := range out {
		h := hashString(fmt.Sprintf("%s:%d:%d:%d", class, i, state, macro), seed)
		out[i] = string(rune('a' + int(h%26)))
	}
	return out
}

// BuildMetadata exposes task66 section 31's structural complexity
// accounting for a Config without running Transform, e.g. for
// MECHANISM_COMPLEXITY.tsv.
func BuildMetadata(c Config) Metadata { return buildMetadata(c) }

// buildMetadata computes task66 section 31's structural complexity
// accounting from the config alone (never from the observed output).
func buildMetadata(c Config) Metadata {
	c = normalize(c)
	states := c.StateCount * c.MacroStates
	symbolClasses := 1
	if c.Grammar != NoGrammar {
		symbolClasses = 3 * grammarClassSize(c.Grammar)
	}
	transitions := 0
	if c.StateCount > 1 {
		transitions += c.StateCount
	}
	if c.MacroStates > 1 {
		transitions += c.MacroStates
	}
	outputRules := 1
	if c.Grammar != NoGrammar {
		outputRules = map[GrammarLevel]int{GrammarLow: 3, GrammarMedium: 6, GrammarHigh: 9}[c.Grammar]
	}
	stochastic := 0
	if c.Homophones > 1 || c.Grouping == RandomGrouping {
		stochastic = 1
	}
	info := classifyInformation(c)
	return Metadata{
		Family: c.Family, Parameters: c, Seed: c.Seed, StateCount: states,
		Information: info, SymbolClasses: symbolClasses, TransitionParameters: transitions,
		OutputRules: outputRules, StochasticDistributions: stochastic,
	}
}

func grammarClassSize(g GrammarLevel) int {
	return map[GrammarLevel]int{GrammarLow: 2, GrammarMedium: 4, GrammarHigh: 6}[g]
}

func classifyInformation(c Config) InformationStatus {
	switch c.Family {
	case "M0", "M1":
		return Reversible
	case "M2":
		return Ambiguous
	case "M3", "M9", "M10", "M11":
		return Lossy
	case "M4", "M5", "M6", "M7":
		return ReversibleWithKeyState
	case "M8":
		if c.Grouping == RandomGrouping {
			return Ambiguous
		}
		return ReversibleWithKeyState
	}
	return Ambiguous
}
