package normalizationcompare

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
	"zcore.dev/voinich/internal/normalization"
	"zcore.dev/voinich/internal/sequenceanalyze"
	"zcore.dev/voinich/internal/workdir"
)

func LoadClasses(path string) (normalization.ClassesOutput, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return normalization.ClassesOutput{}, err
	}
	var result normalization.ClassesOutput
	err = yaml.Unmarshal(data, &result)
	return result, err
}

func LoadSequence(path string) (SequenceAnalysis, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SequenceAnalysis{}, err
	}
	var result SequenceAnalysis
	err = yaml.Unmarshal(data, &result)
	return result, err
}

// FromAnalyzerOutput extracts the subset of an in-process
// sequenceanalyze.Output this tool consumes, matching field-for-field the
// YAML tags SequenceAnalysis used to read back from a sequence-analyze
// subprocess's output file (see PERFORMANCE_REFACTOR_REPORT.md for the
// equivalence argument and test).
func FromAnalyzerOutput(o sequenceanalyze.Output) SequenceAnalysis {
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

// WriteAnalysisYAML persists the full sequence-analyze output for a
// structural (non-random) model at analysisPath, exactly as the
// sequence-analyze binary itself would have when previously invoked as a
// subprocess - this file is a documented pipeline artifact (see
// run-normalization-analysis.sh's -analysis-pattern), not a transient temp
// file, so it is still written even though the analysis is now computed
// in-process.
func WriteAnalysisYAML(path string, output sequenceanalyze.Output) error {
	data, err := yaml.Marshal(output)
	if err != nil {
		return err
	}
	if err := workdir.EnsureParent(path); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func ExtractMetrics(analysis SequenceAnalysis) Metrics {
	result := Metrics{CrossLine: make(map[int]float64), Contexts: make(map[int]ContextOrder)}
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

func CompareModel(model normalization.Model, raw, structural Metrics, random []Metrics) ModelComparison {
	result := ModelComparison{Threshold: model.Threshold, Label: model.Label, Normalization: model.Stats}
	var randomMax []float64
	for _, item := range random {
		randomMax = append(randomMax, item.MaxLength)
	}
	result.MaxCrossLineSequenceLength = MakeEffect(raw.MaxLength, structural.MaxLength, randomMax, true)
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
		result.NGrams = append(result.NGrams, NGramComparison{N: n, CrossLineRepeated: MakeEffect(raw.CrossLine[n], structural.CrossLine[n], values, true)})
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
			ConditionalEntropy:         MakeEffect(rawContext.ConditionalEntropy, structuralContext.ConditionalEntropy, entropy, false),
			RepeatedConditionalEntropy: MakeEffect(rawContext.RepeatedContextConditionalEntropy, structuralContext.RepeatedContextConditionalEntropy, repeatedEntropy, false),
			RepeatedContextCoverage:    MakeEffect(rawContext.RepeatedContextCoverage, structuralContext.RepeatedContextCoverage, coverage, true),
		})
	}
	return result
}

func MakeEffect(raw, structural float64, random []float64, upperTail bool) Effect {
	distribution := Summarize(random)
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

func Summarize(values []float64) RandomDistribution {
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
	result.Percentile05 = Percentile(sorted, .05)
	result.Percentile50 = Percentile(sorted, .50)
	result.Percentile95 = Percentile(sorted, .95)
	return result
}

func Percentile(sorted []float64, probability float64) float64 {
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

// RunRandomTrial is the pure, deterministic, order-independent computation
// for one (threshold, run) unit of work. Its only inputs are the frozen
// corpus, the resolved threshold Model, minTokenCount/singletonMode (both
// come from classes.yaml's Meta, never from a worker-local default), the
// base seed and the run index - normalization.RandomModel already derives
// its RNG source as seed + threshold*100*1e6 + run, so this has no
// dependency on execution order, worker identity, prior trials, or retry
// count. See NORMALIZATION_COMPARE_DISTRIBUTION_AUDIT.md section 2/3.
func RunRandomTrial(model normalization.Model, corpus normalization.Corpus, minTokenCount int, singletonMode string, seed int64, run int, params sequenceanalyze.Parameters) (BaselineResult, error) {
	randomModel := normalization.RandomModel(model, corpus, minTokenCount, seed, run)
	tempDir, err := os.MkdirTemp("", "voinich-normalization-")
	if err != nil {
		return BaselineResult{}, err
	}
	defer os.RemoveAll(tempDir)
	randomCorpus := filepath.Join(tempDir, "corpus.txt")
	if err := normalization.WriteNormalized(randomCorpus, corpus, normalization.Mapping(randomModel, singletonMode)); err != nil {
		return BaselineResult{}, err
	}
	randomOutput, err := sequenceanalyze.AnalyzeFile(randomCorpus, params)
	if err != nil {
		return BaselineResult{}, err
	}
	analysis := FromAnalyzerOutput(randomOutput)
	return BaselineResult{Metrics: ExtractMetrics(analysis), Meta: analysis.Meta}, nil
}

// Fingerprint is the scientific identity used by the remote executor's
// JobID/handshake compatibility check: the corpus content, the full
// classes.yaml content (every threshold's class assignment, not just the
// one this run happens to use), and every parameter that changes a random
// trial's result. It deliberately excludes operational fields (OutputPath,
// Executor, Workers, RemoteListen, ...), mirroring
// structuralprojection.Fingerprint and conditionalregime.computeFingerprint.
func Fingerprint(corpusPath, classesPath string, minTokenCount int, singletonMode string, seed int64, randomRuns int) (string, error) {
	h := sha256.New()
	for _, p := range []string{corpusPath, classesPath} {
		b, err := os.ReadFile(p)
		if err != nil {
			return "", err
		}
		s := sha256.Sum256(b)
		h.Write(s[:])
	}
	v := struct {
		MinTokenCount int
		SingletonMode string
		Seed          int64
		RandomRuns    int
	}{minTokenCount, singletonMode, seed, randomRuns}
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil)), nil
}
