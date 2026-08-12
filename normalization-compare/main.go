package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
	"zcore.dev/voinich/internal/normalization"
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
	classesPath := flag.String("classes", "structural_classes.yaml", "structural class-map YAML")
	inputPath := flag.String("input", "data_work/ivtt_output_1786282555007.txt", "raw corpus for random baselines")
	rawAnalysisPath := flag.String("raw-analysis", "sequence_analysis.yaml", "immutable raw sequence analysis")
	normalizedPattern := flag.String("normalized-pattern", "normalized_%s.txt", "normalized corpus path pattern")
	analysisPattern := flag.String("analysis-pattern", "sequence_analysis_%s.yaml", "structural sequence-analysis output pattern")
	sequenceAnalyzer := flag.String("sequence-analyzer", "bin/sequence-analyze", "compiled sequence-analyze executable")
	outputPath := flag.String("output", "normalization_comparison.yaml", "comparison YAML")
	randomRuns := flag.Int("random-baselines", 100, "matched random runs per threshold")
	randomSeed := flag.Int64("random-seed", 1, "base random seed")
	flag.Parse()
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

	for _, model := range classes.Models {
		normalizedPath := fmt.Sprintf(*normalizedPattern, model.Label)
		analysisPath := fmt.Sprintf(*analysisPattern, model.Label)
		if err := runSequenceAnalyzer(*sequenceAnalyzer, normalizedPath, analysisPath); err != nil {
			fatal(err.Error())
		}
		structural, err := loadSequence(analysisPath)
		if err != nil {
			fatal(fmt.Sprintf("read %s: %v", analysisPath, err))
		}
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
			randomAnalysis := filepath.Join(tempDir, "analysis.yaml")
			if err := normalization.WriteNormalized(randomCorpus, corpus, normalization.Mapping(randomModel, classes.Meta.SingletonMode)); err != nil {
				os.RemoveAll(tempDir)
				fatal(err.Error())
			}
			if err := runSequenceAnalyzer(*sequenceAnalyzer, randomCorpus, randomAnalysis); err != nil {
				os.RemoveAll(tempDir)
				fatal(err.Error())
			}
			analysis, err := loadSequence(randomAnalysis)
			os.RemoveAll(tempDir)
			if err != nil {
				fatal(err.Error())
			}
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
	if err := os.WriteFile(*outputPath, data, 0o644); err != nil {
		fatal(err.Error())
	}
	fmt.Printf("Comparison written to %s\n", *outputPath)
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

func runSequenceAnalyzer(binary, input, output string) error {
	command := exec.Command(binary, "-input", input, "-output", output)
	if data, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("sequence-analyze %s: %w: %s", input, err, data)
	}
	return nil
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
	fmt.Fprintln(os.Stderr, "Error:", message)
	os.Exit(1)
}
