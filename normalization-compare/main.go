package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
	"zcore.dev/voinich/internal/normalization"
	"zcore.dev/voinich/internal/profiling"
	"zcore.dev/voinich/internal/sequenceanalyze"
	"zcore.dev/voinich/internal/workdir"
)

type SequenceSummary struct {
	N                 int `yaml:"n"`
	MultiLineRepeated int `yaml:"multi_line_repeated"`
}

type ContextOrder struct {
	ContextLength                     int     `yaml:"context_length"`
	ConditionalEntropy                float64 `yaml:"conditional_entropy"`
	RepeatedContextConditionalEntropy float64 `yaml:"repeated_context_conditional_entropy"`
	RepeatedContextCoverage           float64 `yaml:"repeated_context_coverage"`
}

type SequenceAnalysis struct {
	Meta struct {
		TokenOccurrences int `yaml:"token_occurrences"`
		Lines            int `yaml:"lines"`
		Transitions      int `yaml:"transitions"`
	} `yaml:"meta"`
	NGramSummary         []SequenceSummary `yaml:"ngram_summary"`
	ContextOrderAnalysis []ContextOrder    `yaml:"context_order_analysis"`
}

type RandomDistribution struct {
	Runs         int     `yaml:"runs"`
	Mean         float64 `yaml:"mean"`
	Stddev       float64 `yaml:"stddev"`
	Min          float64 `yaml:"min"`
	Max          float64 `yaml:"max"`
	Percentile05 float64 `yaml:"percentile_05"`
	Percentile50 float64 `yaml:"percentile_50"`
	Percentile95 float64 `yaml:"percentile_95"`
}

type Effect struct {
	RawValue        float64            `yaml:"raw_value"`
	StructuralValue float64            `yaml:"structural_value"`
	AbsoluteDelta   float64            `yaml:"absolute_delta"`
	Ratio           float64            `yaml:"ratio"`
	Direction       string             `yaml:"empirical_test_direction"`
	Random          RandomDistribution `yaml:"random"`
	ZScore          float64            `yaml:"z_score"`
	EmpiricalP      float64            `yaml:"empirical_p"`
}

type NGramComparison struct {
	N                 int    `yaml:"n"`
	CrossLineRepeated Effect `yaml:"cross_line_repeated"`
}

type ContextComparison struct {
	ContextLength              int    `yaml:"context_length"`
	ConditionalEntropy         Effect `yaml:"conditional_entropy"`
	RepeatedConditionalEntropy Effect `yaml:"repeated_context_conditional_entropy"`
	RepeatedContextCoverage    Effect `yaml:"repeated_context_coverage"`
}

type ModelComparison struct {
	Threshold                  float64                  `yaml:"threshold"`
	Label                      string                   `yaml:"label"`
	Normalization              normalization.ModelStats `yaml:"normalization"`
	MaxCrossLineSequenceLength Effect                   `yaml:"max_cross_line_sequence_length"`
	NGrams                     []NGramComparison        `yaml:"ngrams"`
	ContextOrder               []ContextComparison      `yaml:"context_order"`
}

type ComparisonOutput struct {
	Meta struct {
		RandomBaselines  int    `yaml:"random_baselines"`
		RandomSeed       int64  `yaml:"random_seed"`
		SequenceAnalyzer string `yaml:"sequence_analyzer"`
		RandomMatching   string `yaml:"random_matching"`
		EmpiricalTests   string `yaml:"empirical_tests"`
	} `yaml:"meta"`
	Models []ModelComparison `yaml:"models"`
}

type metrics struct {
	CrossLine map[int]float64
	MaxLength float64
	Contexts  map[int]ContextOrder
}

func main() {
	os.Exit(run())
}

// fatalError is panicked by fatal() and recovered in run(), so a deferred
// profiling session can still be stopped (and its CPU/heap profile written)
// on an error path, exactly as the other profiled CLIs do via an explicit
// `return 1`. fatal's call sites are unchanged; only its own body and the
// top-level control flow around it move to support this.
type fatalError struct{ message string }

func run() (code int) {
	start := time.Now()
	classesPath := flag.String("classes", workdir.Path("structural_classes.yaml"), "structural class-map YAML")
	inputPath := flag.String("input", "data_work/ZL3b-x7.txt", "IVTT -x7 corpus for random baselines")
	rawAnalysisPath := flag.String("raw-analysis", workdir.Path("sequence_analysis.yaml"), "immutable raw sequence analysis")
	normalizedPattern := flag.String("normalized-pattern", workdir.Path("normalized_%s.txt"), "normalized corpus path pattern")
	analysisPattern := flag.String("analysis-pattern", workdir.Path("sequence_analysis_%s.yaml"), "structural sequence-analysis output pattern")
	sequenceAnalyzer := flag.String("sequence-analyzer", workdir.Path("bin", "sequence-analyze"), "compiled sequence-analyze executable")
	outputPath := flag.String("output", workdir.Path("normalization_comparison.yaml"), "comparison YAML")
	randomRuns := flag.Int("random-baselines", 100, "matched random runs per threshold")
	randomSeed := flag.Int64("random-seed", 1, "base random seed")
	prof := profiling.RegisterFlags(flag.CommandLine)
	flag.Parse()

	defer profiling.PrintElapsed(os.Stderr, start)

	sess, err := profiling.Start(prof)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	defer func() {
		if err := sess.Stop(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			code = 1
		}
	}()

	defer func() {
		if r := recover(); r != nil {
			fe, ok := r.(fatalError)
			if !ok {
				panic(r)
			}
			fmt.Fprintln(os.Stderr, "Error:", fe.message)
			code = 1
		}
	}()

	if *randomRuns < 1 {
		fatal("random-baselines must be at least 1")
	}

	classes, err := loadClasses(*classesPath)
	if err != nil {
		fatal(fmt.Sprintf("read classes: %v", err))
	}
	corpus, err := normalization.LoadCorpus(*inputPath)
	if err != nil {
		fatal(fmt.Sprintf("read corpus: %v", err))
	}
	raw, err := loadSequence(*rawAnalysisPath)
	if err != nil {
		fatal(fmt.Sprintf("read raw analysis: %v", err))
	}
	output := ComparisonOutput{}
	output.Meta.RandomBaselines = *randomRuns
	output.Meta.RandomSeed = *randomSeed
	output.Meta.SequenceAnalyzer = *sequenceAnalyzer
	output.Meta.RandomMatching = classes.Meta.RandomMatching
	output.Meta.EmpiricalTests = "upper tail for repeat counts, maximum length, and coverage; lower tail for both entropy metrics; +1 correction"

	analyzerParameters := sequenceanalyze.DefaultParameters()
	for _, model := range classes.Models {
		normalizedPath := fmt.Sprintf(*normalizedPattern, model.Label)
		analysisPath := fmt.Sprintf(*analysisPattern, model.Label)
		structuralOutput, err := sequenceanalyze.AnalyzeFile(normalizedPath, analyzerParameters)
		if err != nil {
			fatal(fmt.Sprintf("analyze %s: %v", normalizedPath, err))
		}
		if err := writeAnalysisYAML(analysisPath, structuralOutput); err != nil {
			fatal(err.Error())
		}
		structural := fromAnalyzerOutput(structuralOutput)
		if structural.Meta != raw.Meta {
			fatal(fmt.Sprintf("corpus invariants changed for threshold %s", model.Label))
		}

		structuralMetrics := extractMetrics(structural)
		randomMetrics := make([]metrics, 0, *randomRuns)
		if model.Stats.MultiMemberClasses == 0 {
			// With no merges every matched model is exactly the raw model. Preserve
			// the requested run count without executing identical analyses.
			for run := 0; run < *randomRuns; run++ {
				randomMetrics = append(randomMetrics, structuralMetrics)
			}
		}
		for run := 0; run < *randomRuns; run++ {
			if model.Stats.MultiMemberClasses == 0 {
				break
			}
			randomModel := normalization.RandomModel(model, corpus, classes.Meta.MinTokenCount, *randomSeed, run)
			tempDir, err := os.MkdirTemp("", "voinich-normalization-")
			if err != nil {
				fatal(err.Error())
			}
			randomCorpus := filepath.Join(tempDir, "corpus.txt")
			if err := normalization.WriteNormalized(randomCorpus, corpus, normalization.Mapping(randomModel, classes.Meta.SingletonMode)); err != nil {
				os.RemoveAll(tempDir)
				fatal(err.Error())
			}
			randomOutput, err := sequenceanalyze.AnalyzeFile(randomCorpus, analyzerParameters)
			os.RemoveAll(tempDir)
			if err != nil {
				fatal(err.Error())
			}
			analysis := fromAnalyzerOutput(randomOutput)
			if analysis.Meta != raw.Meta {
				fatal(fmt.Sprintf("random corpus invariants changed for threshold %s run %d", model.Label, run))
			}
			randomMetrics = append(randomMetrics, extractMetrics(analysis))
		}
		output.Models = append(output.Models, compareModel(model, extractMetrics(raw), structuralMetrics, randomMetrics))
		fmt.Printf("Compared threshold %s (%d random baselines)\n", model.Label, *randomRuns)
	}
	data, err := yaml.Marshal(output)
	if err != nil {
		fatal(err.Error())
	}
	if err := workdir.EnsureParent(*outputPath); err != nil {
		fatal(err.Error())
	}
	if err := os.WriteFile(*outputPath, data, 0o644); err != nil {
		fatal(err.Error())
	}
	fmt.Printf("Comparison written to %s\n", *outputPath)
	return 0
}

func loadClasses(path string) (normalization.ClassesOutput, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return normalization.ClassesOutput{}, err
	}
	var result normalization.ClassesOutput
	err = yaml.Unmarshal(data, &result)
	return result, err
}

func loadSequence(path string) (SequenceAnalysis, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SequenceAnalysis{}, err
	}
	var result SequenceAnalysis
	err = yaml.Unmarshal(data, &result)
	return result, err
}

// fromAnalyzerOutput extracts the subset of an in-process sequenceanalyze.Output
// this tool consumes, matching field-for-field the YAML tags SequenceAnalysis
// used to read back from a sequence-analyze subprocess's output file (see
// PERFORMANCE_REFACTOR_REPORT.md for the equivalence argument and test).
func fromAnalyzerOutput(o sequenceanalyze.Output) SequenceAnalysis {
	result := SequenceAnalysis{}
	result.Meta.TokenOccurrences = o.Meta.TokenOccurrences
	result.Meta.Lines = o.Meta.Lines
	result.Meta.Transitions = o.Meta.Transitions
	result.NGramSummary = make([]SequenceSummary, len(o.NGramSummary))
	for i, s := range o.NGramSummary {
		result.NGramSummary[i] = SequenceSummary{N: s.N, MultiLineRepeated: s.MultiLineRepeated}
	}
	result.ContextOrderAnalysis = make([]ContextOrder, len(o.ContextOrderAnalysis))
	for i, c := range o.ContextOrderAnalysis {
		result.ContextOrderAnalysis[i] = ContextOrder{
			ContextLength:                     c.ContextLength,
			ConditionalEntropy:                c.ConditionalEntropy,
			RepeatedContextConditionalEntropy: c.RepeatedContextConditionalEntropy,
			RepeatedContextCoverage:           c.RepeatedContextCoverage,
		}
	}
	return result
}

// writeAnalysisYAML persists the full sequence-analyze output for a
// structural (non-random) model at analysisPath, exactly as the
// sequence-analyze binary itself would have when previously invoked as a
// subprocess — this file is a documented pipeline artifact (see
// run-normalization-analysis.sh's -analysis-pattern), not a transient
// temp file, so it is still written even though the analysis is now
// computed in-process.
func writeAnalysisYAML(path string, output sequenceanalyze.Output) error {
	data, err := yaml.Marshal(output)
	if err != nil {
		return err
	}
	if err := workdir.EnsureParent(path); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func extractMetrics(analysis SequenceAnalysis) metrics {
	result := metrics{CrossLine: make(map[int]float64), Contexts: make(map[int]ContextOrder)}
	for _, summary := range analysis.NGramSummary {
		result.CrossLine[summary.N] = float64(summary.MultiLineRepeated)
		if summary.MultiLineRepeated > 0 && float64(summary.N) > result.MaxLength {
			result.MaxLength = float64(summary.N)
		}
	}
	for _, context := range analysis.ContextOrderAnalysis {
		result.Contexts[context.ContextLength] = context
	}
	return result
}

func compareModel(model normalization.Model, raw, structural metrics, random []metrics) ModelComparison {
	result := ModelComparison{Threshold: model.Threshold, Label: model.Label, Normalization: model.Stats}
	var randomMax []float64
	for _, item := range random {
		randomMax = append(randomMax, item.MaxLength)
	}
	result.MaxCrossLineSequenceLength = makeEffect(raw.MaxLength, structural.MaxLength, randomMax, true)
	var lengths []int
	for n := range raw.CrossLine {
		lengths = append(lengths, n)
	}
	sort.Ints(lengths)
	for _, n := range lengths {
		var values []float64
		for _, item := range random {
			values = append(values, item.CrossLine[n])
		}
		result.NGrams = append(result.NGrams, NGramComparison{N: n, CrossLineRepeated: makeEffect(raw.CrossLine[n], structural.CrossLine[n], values, true)})
	}
	var contextLengths []int
	for length := range raw.Contexts {
		contextLengths = append(contextLengths, length)
	}
	sort.Ints(contextLengths)
	for _, length := range contextLengths {
		rawContext := raw.Contexts[length]
		structuralContext := structural.Contexts[length]
		var entropy, repeatedEntropy, coverage []float64
		for _, item := range random {
			context := item.Contexts[length]
			entropy = append(entropy, context.ConditionalEntropy)
			repeatedEntropy = append(repeatedEntropy, context.RepeatedContextConditionalEntropy)
			coverage = append(coverage, context.RepeatedContextCoverage)
		}
		result.ContextOrder = append(result.ContextOrder, ContextComparison{
			ContextLength:              length,
			ConditionalEntropy:         makeEffect(rawContext.ConditionalEntropy, structuralContext.ConditionalEntropy, entropy, false),
			RepeatedConditionalEntropy: makeEffect(rawContext.RepeatedContextConditionalEntropy, structuralContext.RepeatedContextConditionalEntropy, repeatedEntropy, false),
			RepeatedContextCoverage:    makeEffect(rawContext.RepeatedContextCoverage, structuralContext.RepeatedContextCoverage, coverage, true),
		})
	}
	return result
}

func makeEffect(raw, structural float64, random []float64, upperTail bool) Effect {
	distribution := summarize(random)
	ratio := 0.0
	if raw != 0 {
		ratio = structural / raw
	}
	zScore := 0.0
	if distribution.Stddev > 0 {
		zScore = (structural - distribution.Mean) / distribution.Stddev
	}
	extreme := 0
	for _, value := range random {
		if (upperTail && value >= structural) || (!upperTail && value <= structural) {
			extreme++
		}
	}
	direction := "random >= structural"
	if !upperTail {
		direction = "random <= structural"
	}
	return Effect{
		RawValue: raw, StructuralValue: structural, AbsoluteDelta: structural - raw, Ratio: ratio,
		Direction: direction, Random: distribution, ZScore: zScore, EmpiricalP: float64(extreme+1) / float64(len(random)+1),
	}
}

func summarize(values []float64) RandomDistribution {
	result := RandomDistribution{Runs: len(values)}
	if len(values) == 0 {
		return result
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	result.Min, result.Max = sorted[0], sorted[len(sorted)-1]
	for _, value := range sorted {
		result.Mean += value
	}
	result.Mean /= float64(len(sorted))
	for _, value := range sorted {
		difference := value - result.Mean
		result.Stddev += difference * difference
	}
	result.Stddev = math.Sqrt(result.Stddev / float64(len(sorted)))
	if result.Stddev < 1e-12 {
		result.Stddev = 0
	}
	result.Percentile05 = percentile(sorted, .05)
	result.Percentile50 = percentile(sorted, .50)
	result.Percentile95 = percentile(sorted, .95)
	return result
}

func percentile(sorted []float64, probability float64) float64 {
	if len(sorted) == 1 {
		return sorted[0]
	}
	position := probability * float64(len(sorted)-1)
	lower := int(math.Floor(position))
	upper := int(math.Ceil(position))
	if lower == upper {
		return sorted[lower]
	}
	weight := position - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

func fatal(message string) {
	panic(fatalError{message})
}
