package fingerprintv2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type rawDistribution struct {
	ID     string    `json:"id"`
	Values []float64 `json:"values"`
}

type rawCorpusResult struct {
	CorpusID      string            `json:"corpus_id"`
	Distributions []rawDistribution `json:"distributions"`
}

type rawResult struct {
	Version     string            `json:"version"`
	Fingerprint Fingerprint       `json:"fingerprint"`
	Raw         []rawCorpusResult `json:"raw"`
}

// RunFile is the CLI-facing reproducible entry point. It uses strict YAML
// decoding so an accidental misspelled parameter cannot silently change a
// research run.
func RunFile(path string) (Fingerprint, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Fingerprint{}, fmt.Errorf("read config: %w", err)
	}
	if err := validateDeclaredNumericConfig(b); err != nil {
		return Fingerprint{}, err
	}
	var c Config
	decoder := yaml.NewDecoder(bytes.NewReader(b))
	decoder.KnownFields(true)
	if err := decoder.Decode(&c); err != nil {
		return Fingerprint{}, fmt.Errorf("decode config: %w", err)
	}
	return runAndWrite(c, b)
}

func validateDeclaredNumericConfig(data []byte) error {
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decode config: %w", err)
	}
	for _, key := range []string{"repetitions", "min_rule_support", "graph_swaps"} {
		value, present := raw[key]
		if !present {
			continue
		}
		number, ok := yamlNumber(value)
		if !ok || math.IsNaN(number) || math.IsInf(number, 0) || number <= 0 || math.Trunc(number) != number {
			return fmt.Errorf("%s must be a positive integer when declared", key)
		}
	}
	for _, key := range []string{"alpha", "diagnostic_tolerance"} {
		value, present := raw[key]
		if !present {
			continue
		}
		number, ok := yamlNumber(value)
		if !ok || math.IsNaN(number) || math.IsInf(number, 0) || number <= 0 || (key == "alpha" && number >= 1) {
			return fmt.Errorf("%s must be a finite positive number%s when declared", key, map[bool]string{true: " below 1"}[key == "alpha"])
		}
	}
	return nil
}

func yamlNumber(value any) (float64, bool) {
	switch value := value.(type) {
	case int:
		return float64(value), true
	case int64:
		return float64(value), true
	case uint64:
		return float64(value), true
	case float64:
		return value, true
	default:
		return 0, false
	}
}

// Run performs calculation without writing artifacts. Tests and callers that
// need the required artifact set should use RunFile or RunAndWrite.
func Run(c Config) (Fingerprint, error) {
	c, err := c.normalized()
	if err != nil {
		return Fingerprint{}, err
	}
	return run(c)
}

// RunAndWrite exposes artifact writing to callers with programmatically
// constructed configs. The supplied configuration is copied verbatim when
// nonempty; otherwise normalized YAML is written.
func RunAndWrite(c Config, configCopy []byte) (Fingerprint, error) {
	return runAndWrite(c, configCopy)
}

func runAndWrite(c Config, configCopy []byte) (Fingerprint, error) {
	c, err := c.normalized()
	if err != nil {
		return Fingerprint{}, err
	}
	if len(configCopy) == 0 {
		configCopy, err = yaml.Marshal(c)
		if err != nil {
			return Fingerprint{}, fmt.Errorf("encode normalized config: %w", err)
		}
	}
	fingerprint, raw, err := runWithRaw(c)
	if err != nil {
		return Fingerprint{}, err
	}
	if err := writeArtifacts(c.OutputDir, configCopy, fingerprint, raw); err != nil {
		return Fingerprint{}, err
	}
	return fingerprint, nil
}

func run(c Config) (Fingerprint, error) {
	fingerprint, _, err := runWithRaw(c)
	return fingerprint, err
}

func runWithRaw(c Config) (Fingerprint, []rawCorpusResult, error) {
	primary, err := loadCorpus(c.Primary)
	if err != nil {
		return Fingerprint{}, nil, err
	}
	commit, warning := implementationCommit()
	p, rawPrimary, err := analyzeOne(primary, c, c.Seed)
	if err != nil {
		return Fingerprint{}, nil, fmt.Errorf("analyze primary corpus: %w", err)
	}
	f := Fingerprint{
		Version: Version,
		Provenance: Provenance{
			ImplementationCommit: commit, MetricVersion: MetricVersion, NullModelVersion: NullModelVersion,
			Seed: c.Seed, Repetitions: c.Repetitions,
			PreprocessingProfile: "generic natural lines; optional strict IVTFF alignment; declared glyph mode",
			GeneratorSettings:    strings.Join(c.Grammar.Modes, ",") + "; position-constrained first-order generator; frequency ranks assigned after de-duplication within length",
		},
		Primary: p,
	}
	if warning != "" {
		f.Warnings = append(f.Warnings, warning)
	}
	if p.Grammar != nil && p.Grammar.Validation != "SUPPORTED" {
		f.Warnings = append(f.Warnings, "C-GRAMMAR diagnostics exceeded the configured tolerance for at least one replicate; inspect raw_results.json before inferential interpretation.")
	}
	raw := []rawCorpusResult{rawPrimary}
	for i, controlConfig := range c.Controls {
		control, e := loadCorpus(controlConfig.Corpus)
		if e != nil {
			return Fingerprint{}, nil, fmt.Errorf("load control %q: %w", controlConfig.Name, e)
		}
		result, controlRaw, e := analyzeOne(control, c, c.Seed+int64((i+1)*10000000))
		if e != nil {
			return Fingerprint{}, nil, fmt.Errorf("analyze control %q: %w", controlConfig.Name, e)
		}
		f.Controls = append(f.Controls, result)
		raw = append(raw, controlRaw)
	}
	f.Verdicts = verdicts(f)
	return f, raw, nil
}

func implementationCommit() (string, string) {
	command := exec.Command("git", "rev-parse", "HEAD")
	b, err := command.Output()
	if err != nil {
		return "UNAVAILABLE", "Git commit could not be recorded by git rev-parse HEAD."
	}
	return strings.TrimSpace(string(b)), ""
}

func analyzeOne(corpus corpus, cfg Config, seed int64) (CorpusResult, rawCorpusResult, error) {
	observed := analyzeBare(corpus, cfg)
	if len(observed.graph.nodes) < 2 {
		reason := "INSUFFICIENT_SUPPORT: fewer than two vocabulary types are available for edit-graph or lexical-paradigm inference."
		return CorpusResult{
			Corpus: corpus.info,
			Metrics: Metrics{
				LP1: observed.lp1,
				LP2: LP2Result{Statistic: "LP1 rule-support Gini (type-level directed edit pairs)", ProductivityState: "INSUFFICIENT_SUPPORT"},
				EF1: observed.ef1,
				EF2: ef2Bare(observed),
				EF3: ef3Bare(observed),
				EF4: EF4Result{Verdict: "INSUFFICIENT_SUPPORT", Reason: reason},
			},
		}, rawCorpusResult{CorpusID: corpus.info.ID}, nil
	}
	model := newGrammarModel(corpus)
	runs := make([]GrammarRun, 0, len(cfg.Grammar.Modes)*cfg.Repetitions)
	byMode := map[string][]GrammarRun{}
	for modeIndex, mode := range cfg.Grammar.Modes {
		for replicate := 0; replicate < cfg.Repetitions; replicate++ {
			runSeed := seed + int64((modeIndex+1)*1000000+replicate)
			generated, err := model.generate(corpus, mode, rand.New(rand.NewSource(runSeed)))
			if err != nil {
				return CorpusResult{}, rawCorpusResult{}, fmt.Errorf("%s replicate %d: %w", mode, replicate, err)
			}
			bare := analyzeBare(generated, cfg)
			run := GrammarRun{
				Mode: mode, Replicate: replicate, Seed: runSeed, Diagnostic: grammarDiagnostic(corpus, generated),
				LP1Gini: bare.lp1.SupportGini, PrefixNMI: bare.prefixNMI, SuffixNMI: bare.suffixNMI,
				EF1: bare.ef1, EF2: ef2Bare(bare), EF3: ef3Bare(bare),
			}
			runs = append(runs, run)
			byMode[mode] = append(byMode[mode], run)
		}
	}
	sort.Slice(runs, func(i, j int) bool {
		if runs[i].Mode != runs[j].Mode {
			return runs[i].Mode < runs[j].Mode
		}
		return runs[i].Replicate < runs[j].Replicate
	})
	grammar := &GrammarSummary{Runs: runs}
	grammar.Validation, grammar.Reason = grammarValidation(runs, cfg.DiagnosticTolerance)
	validModes := validGrammarModes(byMode, cfg.DiagnosticTolerance)

	lexicalTests := make([]NullTest, 0, len(cfg.Grammar.Modes)+4)
	raw := rawCorpusResult{CorpusID: corpus.info.ID}
	for _, mode := range cfg.Grammar.Modes {
		null := grammarValues(byMode[mode], func(r GrammarRun) float64 { return r.LP1Gini })
		id := "lp2/c-grammar/" + mode
		if validModes[mode] {
			lexicalTests = append(lexicalTests, nullTest(id, "C-GRAMMAR "+mode, observed.lp1.SupportGini, null))
		}
		raw.Distributions = append(raw.Distributions, rawDistribution{ID: id, Values: null})
	}
	for _, pairing := range []struct {
		id   string
		freq bool
	}{{"lp2/c-len", false}, {"lp2/c-freq", true}} {
		null := make([]float64, cfg.Repetitions)
		for r := range null {
			null[r] = randomPairGini(corpus, observed.graph, pairing.freq, rand.New(rand.NewSource(seed+int64(2000000+r)+boolOffset(pairing.freq))))
		}
		modelName := "C-LEN length-matched random type pairing"
		if pairing.freq {
			modelName = "C-FREQ length-and-frequency-bin matched random type pairing"
		}
		lexicalTests = append(lexicalTests, nullTest(pairing.id, modelName, observed.lp1.SupportGini, null))
		raw.Distributions = append(raw.Distributions, rawDistribution{ID: pairing.id, Values: null})
	}
	prefixNull, suffixNull := attachmentPermutation(corpus, cfg.Repetitions, rand.New(rand.NewSource(seed+3000000)))
	lexicalTests = append(lexicalTests,
		nullTest("lp4/c-len-prefix", "C-LEN affix permutation within token length", observed.prefixNMI, prefixNull),
		nullTest("lp4/c-len-suffix", "C-LEN affix permutation within token length", observed.suffixNMI, suffixNull),
	)
	raw.Distributions = append(raw.Distributions,
		rawDistribution{ID: "lp4/c-len-prefix", Values: prefixNull},
		rawDistribution{ID: "lp4/c-len-suffix", Values: suffixNull},
	)
	for _, mode := range cfg.Grammar.Modes {
		prefix := grammarValues(byMode[mode], func(r GrammarRun) float64 { return r.PrefixNMI })
		suffix := grammarValues(byMode[mode], func(r GrammarRun) float64 { return r.SuffixNMI })
		if validModes[mode] {
			lexicalTests = append(lexicalTests,
				nullTest("lp4/c-grammar-prefix/"+mode, "C-GRAMMAR "+mode, observed.prefixNMI, prefix),
				nullTest("lp4/c-grammar-suffix/"+mode, "C-GRAMMAR "+mode, observed.suffixNMI, suffix),
			)
		}
		raw.Distributions = append(raw.Distributions,
			rawDistribution{ID: "lp4/c-grammar-prefix/" + mode, Values: prefix},
			rawDistribution{ID: "lp4/c-grammar-suffix/" + mode, Values: suffix},
		)
	}
	lexicalTests = fdr(lexicalTests)
	lp2Tests, prefixTest, suffixTest, prefixGrammarTests, suffixGrammarTests := splitLexicalTests(lexicalTests)
	productive := map[string]bool{}
	productivityState := "INCONCLUSIVE"
	if len(validModes) > 0 && grammarTestSignificant(lp2Tests, cfg.Alpha) {
		for rule := range observed.candidates {
			productive[rule] = true
		}
		if len(productive) > 0 {
			productivityState = "SUPPORTED"
		} else {
			productivityState = "INCONCLUSIVE"
		}
	} else if len(validModes) > 0 {
		productivityState = "NOT_SUPPORTED"
	}
	productiveNames := orderedKeys(productive)
	lp2Result := LP2Result{
		Statistic: "LP1 rule-support Gini (type-level directed edit pairs)", Tests: lp2Tests,
		ProductiveRules: productiveNames, ProductivityState: productivityState,
	}
	lp3Result := lp3(corpus, productive, observed.graph, cfg.Repetitions, rand.New(rand.NewSource(seed+4000000)))
	lp4Result := LP4Result{
		ZoneConvention: "exactly one glyph prefix and one glyph suffix; interior glyph sequence is core; lengths <3 excluded",
		Prefix:         AttachmentResult{NormalizedMI: observed.prefixNMI, Eligible: observed.attachmentN, Excluded: observed.excludedN, Permutation: prefixTest, GrammarTests: prefixGrammarTests},
		Suffix:         AttachmentResult{NormalizedMI: observed.suffixNMI, Eligible: observed.attachmentN, Excluded: observed.excludedN, Permutation: suffixTest, GrammarTests: suffixGrammarTests},
	}

	ef2Result := ef2WithControl(observed, cfg, rand.New(rand.NewSource(seed+5000000)))
	ef3Result := ef3WithControl(observed, corpus, cfg, rand.New(rand.NewSource(seed+6000000)))
	raw.Distributions = append(raw.Distributions,
		rawDistribution{ID: ef2Result.ConfigurationTest.ID, Values: configurationDistribution(observed, cfg, seed+5000000)},
		rawDistribution{ID: ef3Result.FrequencyControl.ID, Values: frequencyDistribution(observed, corpus, cfg, seed+6000000)},
	)
	ef4Result := ef4(observed, byMode, validModes)
	grammarTests := ef4Result.Tests
	ef4Result.Tests = nil
	efTests := append([]NullTest{ef2Result.ConfigurationTest, ef3Result.FrequencyControl}, grammarTests...)
	efTests = fdr(efTests)
	for _, test := range efTests {
		switch test.ID {
		case ef2Result.ConfigurationTest.ID:
			ef2Result.ConfigurationTest = test
		case ef3Result.FrequencyControl.ID:
			ef3Result.FrequencyControl = test
		default:
			ef4Result.Tests = append(ef4Result.Tests, test)
		}
	}
	ef4Result.Verdict, ef4Result.Reason = ef4Verdict(ef4Result.Tests, cfg, len(validModes) > 0)
	metrics := Metrics{LP1: observed.lp1, LP2: lp2Result, LP3: lp3Result, LP4: lp4Result, EF1: observed.ef1, EF2: ef2Result, EF3: ef3Result, EF4: ef4Result}
	return CorpusResult{Corpus: corpus.info, Metrics: metrics, Grammar: grammar}, raw, nil
}

func boolOffset(v bool) int64 {
	if v {
		return 1
	}
	return 0
}

func grammarValues(runs []GrammarRun, get func(GrammarRun) float64) []float64 {
	out := make([]float64, len(runs))
	for i, r := range runs {
		out[i] = get(r)
	}
	return out
}

func splitLexicalTests(all []NullTest) ([]NullTest, NullTest, NullTest, []NullTest, []NullTest) {
	lp2 := make([]NullTest, 0)
	prefixGrammar, suffixGrammar := make([]NullTest, 0), make([]NullTest, 0)
	var prefix, suffix NullTest
	for _, test := range all {
		switch test.ID {
		case "lp4/c-len-prefix":
			prefix = test
		case "lp4/c-len-suffix":
			suffix = test
		default:
			if strings.HasPrefix(test.ID, "lp4/c-grammar-prefix/") {
				prefixGrammar = append(prefixGrammar, test)
				continue
			}
			if strings.HasPrefix(test.ID, "lp4/c-grammar-suffix/") {
				suffixGrammar = append(suffixGrammar, test)
				continue
			}
			lp2 = append(lp2, test)
		}
	}
	stableTests(lp2)
	return lp2, prefix, suffix, prefixGrammar, suffixGrammar
}

func validGrammarModes(byMode map[string][]GrammarRun, tolerance float64) map[string]bool {
	out := map[string]bool{}
	for mode, runs := range byMode {
		validation, _ := grammarValidation(runs, tolerance)
		if validation == "SUPPORTED" {
			out[mode] = true
		}
	}
	return out
}

func grammarTestSignificant(tests []NullTest, alpha float64) bool {
	for _, test := range tests {
		if strings.HasPrefix(test.ID, "lp2/c-grammar/") && test.QValue <= alpha {
			return true
		}
	}
	return false
}

func grammarValidation(runs []GrammarRun, tolerance float64) (string, string) {
	if len(runs) == 0 {
		return "INCONCLUSIVE", "No C-GRAMMAR replicates were produced."
	}
	for _, run := range runs {
		d := run.Diagnostic
		if !d.TokenCountExact || !d.LengthDistributionExact || !d.AlphabetExact {
			return "NOT_SUPPORTED", "A required exact marginal (token count, length distribution, or alphabet) was not preserved."
		}
		if d.PositionalGlyphTV > tolerance || d.InitialGlyphTV > tolerance || d.FinalGlyphTV > tolerance || d.BigramTV > tolerance {
			return "PARTIALLY_SUPPORTED", "At least one positional or bigram total-variation diagnostic exceeded diagnostic_tolerance."
		}
	}
	return "SUPPORTED", "All required exact marginals held and all positional/start/end/bigram diagnostic distances were within diagnostic_tolerance."
}

func configurationDistribution(b bareMetrics, cfg Config, seed int64) []float64 {
	rng := rand.New(rand.NewSource(seed))
	out := make([]float64, cfg.Repetitions)
	for i := range out {
		swapped := degreePreservingSwap(b.graph, cfg.GraphSwaps*max(1, len(b.graph.edgeList())), rng)
		out[i], _, _, _ = graphMotifs(swapped)
	}
	return out
}

func frequencyDistribution(b bareMetrics, c corpus, cfg Config, seed int64) []float64 {
	freq := frequencies(c)
	values := make([]int, 0, len(b.graph.nodes))
	for _, n := range b.graph.nodes {
		values = append(values, freq[n])
	}
	rng := rand.New(rand.NewSource(seed))
	out := make([]float64, cfg.Repetitions)
	for r := range out {
		p := rng.Perm(len(values))
		shuffled := map[string]int{}
		for i, n := range b.graph.nodes {
			shuffled[n] = values[p[i]]
		}
		out[r] = math.Abs(degreeFrequencySpearman(b.graph, shuffled))
	}
	return out
}

func ef4(observed bareMetrics, byMode map[string][]GrammarRun, validModes map[string]bool) EF4Result {
	tests := make([]NullTest, 0, len(byMode)*3)
	for _, mode := range orderedKeys(byMode) {
		if !validModes[mode] {
			continue
		}
		runs := byMode[mode]
		tests = append(tests,
			nullTest("ef1/c-grammar/"+mode, "C-GRAMMAR "+mode, observed.ef1.GiantComponentShare, grammarValues(runs, func(r GrammarRun) float64 { return r.EF1.GiantComponentShare })),
			nullTest("ef2/c-grammar/"+mode, "C-GRAMMAR "+mode, observed.clustering, grammarValues(runs, func(r GrammarRun) float64 { return r.EF2.GlobalClustering })),
		)
		null := grammarValues(runs, func(r GrammarRun) float64 { return math.Abs(r.EF3.SpearmanDegreeLogFrequency) })
		tests = append(tests, nullTest("ef3/c-grammar/"+mode, "C-GRAMMAR "+mode, math.Abs(observed.spearman), null))
	}
	stableTests(tests)
	return EF4Result{Tests: tests}
}

func ef4Verdict(tests []NullTest, cfg Config, grammarValid bool) (string, string) {
	if !grammarValid || len(tests) == 0 {
		return "INCONCLUSIVE", "No C-GRAMMAR mode passed validation, so grammar-boundedness inference is unavailable."
	}
	significant := 0
	for _, test := range tests {
		if test.QValue <= cfg.Alpha {
			significant++
		}
	}
	verdict, reason := "CONSISTENT_WITH_GRAMMAR_BOUND", "No EF graph statistic exceeds the C-GRAMMAR distribution after the declared EF FDR correction."
	if significant == len(tests) && significant > 0 {
		verdict, reason = "EXCEEDS_GRAMMAR_BOUND", "Every reported EF statistic exceeds its C-GRAMMAR distribution after the declared EF FDR correction."
	} else if significant > 0 {
		verdict, reason = "MIXED", "Some but not all EF graph statistics exceed C-GRAMMAR after the declared EF FDR correction."
	}
	return verdict, reason
}

func verdicts(f Fingerprint) []Verdict {
	p := f.Primary
	grammarValue, grammarBasis := "INCONCLUSIVE", "No C-GRAMMAR summary was produced."
	if p.Grammar != nil {
		grammarValue, grammarBasis = p.Grammar.Validation, p.Grammar.Reason
	}
	directional := p.Metrics.LP2.ProductivityState
	if directional == "SUPPORTED" {
		directional = "SUPPORTED"
	} else if directional == "INCONCLUSIVE" || directional == "INSUFFICIENT_SUPPORT" {
		directional = "INCONCLUSIVE"
	} else {
		directional = "NOT_SUPPORTED"
	}
	edit := "INCONCLUSIVE"
	if p.Metrics.EF4.Verdict == "EXCEEDS_GRAMMAR_BOUND" && directional == "SUPPORTED" {
		edit = "SUPPORTED"
	} else if p.Metrics.EF4.Verdict == "CONSISTENT_WITH_GRAMMAR_BOUND" {
		edit = "NOT_SUPPORTED"
	}
	context := "INCONCLUSIVE"
	if p.Metrics.LP3.Locality.GlobalNull != nil {
		if p.Metrics.LP3.Locality.GlobalNull.PValue <= 0.05 {
			context = "PARTIALLY_SUPPORTED"
		} else if p.Metrics.LP3.ProductiveRuleCount > 0 {
			context = "NOT_SUPPORTED"
		}
	}
	productivity := directional
	ready := "PARTIALLY_SUPPORTED"
	if grammarValue != "SUPPORTED" {
		ready = "INCONCLUSIVE"
	}
	return []Verdict{
		{ID: "C_GRAMMAR_VALIDATION", Value: grammarValue, Basis: grammarBasis, Limitations: "Diagnostics establish marginal adequacy, not absence of all latent lexical structure."},
		{ID: "EDIT_NEIGHBORHOODS_EXCEED_GRAMMAR_NULL", Value: edit, Basis: p.Metrics.EF4.Reason, Limitations: "LP and EF comparisons share the C-GRAMMAR null and are not independent evidence."},
		{ID: "DIRECTIONAL_TRANSFORMATIONS_SUPPORTED", Value: directional, Basis: "LP2 C-GRAMMAR/C-LEN/C-FREQ support-concentration tests with lexical-family FDR.", Limitations: "Global concentration does not identify individual linguistic or cryptographic operations."},
		{ID: "CONTEXT_CONDITIONING_SUPPORTED", Value: context, Basis: "LP3 same-line family-occurrence locality under C-GLOBAL.", Limitations: "This first block reports line/page locality only; richer metadata conditioning is deferred."},
		{ID: "PARADIGM_PRODUCTIVITY_SUPPORTED", Value: productivity, Basis: "LP2 concentration plus declared support threshold, with LP3 restricted to selected rules.", Limitations: "No held-out-folio generalization test is implemented in this first block."},
		{ID: "LEXICAL_PARADIGM_BLOCK_READY", Value: ready, Basis: "LP1-LP4 and EF1-EF4 are computed with deterministic diagnostics and declared nulls.", Limitations: "Canonical corpus availability, metadata folds, and task77 stability extensions remain separate requirements."},
	}
}

func writeArtifacts(dir string, configCopy []byte, fingerprint Fingerprint, raw []rawCorpusResult) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	writeJSON := func(name string, value any) error {
		b, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal %s: %w", name, err)
		}
		b = append(b, '\n')
		if err := os.WriteFile(filepath.Join(dir, name), b, 0644); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
		return nil
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), configCopy, 0644); err != nil {
		return fmt.Errorf("write config copy: %w", err)
	}
	if err := writeJSON("fingerprint.json", fingerprint); err != nil {
		return err
	}
	if err := writeJSON("raw_results.json", rawResult{Version: Version, Fingerprint: fingerprint, Raw: raw}); err != nil {
		return err
	}
	if err := writeJSON("warnings.json", fingerprint.Warnings); err != nil {
		return err
	}
	if err := writeJSON("errors.json", fingerprint.Errors); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "report.md"), []byte(compactReport(fingerprint)), 0644); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	return nil
}

func compactReport(f Fingerprint) string {
	p := f.Primary
	var b strings.Builder
	fmt.Fprintf(&b, "# Fingerprint v2 lexical-paradigm report\n\n")
	fmt.Fprintf(&b, "- Version: `%s`\n- Commit: `%s`\n- Seed/repetitions: `%d` / `%d`\n", f.Version, f.Provenance.ImplementationCommit, f.Provenance.Seed, f.Provenance.Repetitions)
	fmt.Fprintf(&b, "- Primary corpus: `%s`; tokens `%d`; types `%d`; SHA-256 `%s`\n", p.Corpus.ID, p.Corpus.TokenCount, p.Corpus.VocabularySize, p.Corpus.SHA256)
	fmt.Fprintf(&b, "- LP1 support Gini: `%.6g`; top-rule share: `%.6g`\n", p.Metrics.LP1.SupportGini, p.Metrics.LP1.TopRuleShare)
	fmt.Fprintf(&b, "- EF1 edges/isolates/giant share: `%d` / `%d` / `%.6g`\n", p.Metrics.EF1.EdgeCount, p.Metrics.EF1.IsolateCount, p.Metrics.EF1.GiantComponentShare)
	fmt.Fprintf(&b, "- EF2 clustering/triangles/3-paths/4-cycles: `%.6g` / `%d` / `%d` / `%d`\n", p.Metrics.EF2.GlobalClustering, p.Metrics.EF2.Triangles, p.Metrics.EF2.Paths3, p.Metrics.EF2.Cycles4)
	fmt.Fprintf(&b, "- EF3 Spearman(degree, log frequency): `%.6g`\n\n", p.Metrics.EF3.SpearmanDegreeLogFrequency)
	fmt.Fprintf(&b, "## Verdicts\n\n")
	for _, v := range f.Verdicts {
		fmt.Fprintf(&b, "- **%s:** `%s` — %s\n", v.ID, v.Value, v.Basis)
	}
	fmt.Fprintf(&b, "\nRaw null distributions, per-replicate grammar diagnostics, and input checksums are in `raw_results.json`. These values describe only the configured input corpus; this report does not identify it as a canonical Voynich run.\n")
	return b.String()
}
