package validation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"sort"

	"gopkg.in/yaml.v3"
	"zcore.dev/voinich/internal/normalization"
)

const leaveOneOutThreshold = .70

func Run(config Config) (Output, error) {
	if err := validateConfig(config); err != nil {
		return Output{}, err
	}
	corpus, corpusHash, err := LoadCorpus(config.InputPath)
	if err != nil {
		return Output{}, fmt.Errorf("load corpus: %w", err)
	}
	classes, classesHash, err := loadClasses(config.ClassesPath)
	if err != nil {
		return Output{}, fmt.Errorf("load full-corpus classes: %w", err)
	}
	fullModel, err := selectModel(classes.Models, leaveOneOutThreshold)
	if err != nil {
		return Output{}, err
	}

	output := Output{
		Meta: Meta{
			Input: config.InputPath, FullCorpusClasses: config.ClassesPath,
			PhysicalLines: len(corpus.Lines), NonEmptyLines: nonEmptyLines(corpus.Lines),
			TokenOccurrences: corpus.Occurrences, Transitions: corpus.Transitions,
			CorpusSHA256: corpusHash, FullCorpusClassesSHA256: classesHash,
		},
		Parameters: Parameters{
			Folds: config.Folds, FoldSeed: config.FoldSeed, Threshold: config.Threshold,
			MinTokenCount: config.MinTokenCount, RandomBaselines: config.RandomBaselines,
			RandomSeed: config.RandomSeed, MinN: config.MinN, MaxN: config.MaxN, MaxContext: config.MaxContext,
			LeaveOneOutThreshold: leaveOneOutThreshold,
		},
		Methodology: Methodology{
			SplitUnit:            "line",
			Split:                "line indexes are deterministically shuffled with math/rand using fold_seed, assigned round-robin to folds, then restored to source order; a line is never split",
			Training:             "each fold recomputes dictionary counts, all position observations, full predecessor/successor counts, pair similarities, and complete-link classes from TRAIN lines only; full-corpus structural_analysis.yaml is not read",
			Eligibility:          "TRAIN token count >= min_token_count; TEST counts never affect eligibility or candidate generation",
			Clustering:           "deterministic agglomerative complete-link using the arithmetic mean of TRAIN-only position similarity (1-JSD) and left/right cosine similarities; every pair must be present and >= threshold",
			TestApplication:      "fixed TRAIN multi-member mappings are applied to TEST; unknown and singleton tokens retain their surface form; TEST never changes a class",
			SequenceBoundary:     "each TEST physical line is independent; n-grams and contexts never cross line boundaries",
			CrossLineRepeated:    "a unique n-gram with at least two occurrences in at least two distinct TEST lines",
			ConditionalEntropy:   "H(next|context) is weighted by TEST context occurrences; entropy_delta is raw minus normalized; repeated contexts have at least two observations",
			RandomMatching:       "for each fold, TRAIN-derived class sizes are retained and TRAIN-eligible tokens are sampled without replacement from base-2 TRAIN-frequency bins with nearest-bin fallback; mappings are then frozen and applied to TEST",
			RandomSeedDerivation: "normalization.RandomModel receives random_seed + fold*10000000000; its deterministic derivation additionally includes threshold and run_number",
			EmpiricalTests:       "upper tail with +1 correction: random >= structural for counts, maximum length, coverage delta, and entropy reduction (raw entropy - normalized entropy)",
			ClassStability:       "reported pairs are the union of pairs co-clustered in at least one TRAIN fold; folds where either TRAIN token is ineligible are excluded from the denominator",
			Pooled:               "sum of independent fold-level TEST counts; TEST corpora are not recombined and no cross-fold sequences are searched",
			ThresholdSelection:   "0.70 is fixed before validation from the prior full-corpus experiment and is not tuned on folds",
		},
	}

	foldIndexes, err := SplitFolds(corpus.Lines, config.Folds, config.FoldSeed)
	if err != nil {
		return Output{}, err
	}
	models := make([]normalization.Model, 0, config.Folds)
	eligibility := make([]map[string]bool, 0, config.Folds)
	for foldIndex, indexes := range foldIndexes {
		train, test, err := Partition(corpus, indexes)
		if err != nil {
			return Output{}, err
		}
		model, eligible, err := BuildTrainModel(train, config)
		if err != nil {
			return Output{}, fmt.Errorf("fold %d training: %w", foldIndex+1, err)
		}
		models = append(models, model)
		eligibility = append(eligibility, eligible)
		mapping := normalization.Mapping(model, "preserve")
		normalizedTest := applyMapping(test, mapping)
		rawMetrics := AnalyzeSequences(test, config.MinN, config.MaxN, config.MaxContext)
		structuralMetrics := AnalyzeSequences(normalizedTest, config.MinN, config.MaxN, config.MaxContext)
		randomMetrics := make([]sequenceMetrics, 0, config.RandomBaselines)
		foldSeed := config.RandomSeed + int64(foldIndex+1)*10_000_000_000
		for run := 0; run < config.RandomBaselines; run++ {
			randomModel := normalization.RandomModel(model, normalizationCorpus(train), config.MinTokenCount, foldSeed, run)
			randomTest := applyMapping(test, normalization.Mapping(randomModel, "preserve"))
			randomMetrics = append(randomMetrics, AnalyzeSequences(randomTest, config.MinN, config.MaxN, config.MaxContext))
		}
		multi := multiMemberClasses(model)
		covered := coveredOccurrences(test, model)
		fold := FoldResult{
			Fold: foldIndex + 1, Train: partitionStats(train, false), Test: partitionStats(test, true),
			StructuralClasses: TrainClasses{
				EligibleTokens: len(eligible), MultiMemberClasses: len(multi),
				TokensInClasses: countMembers(multi), OccurrenceCoverage: model.Stats.TokenOccurrenceCoverage,
				Classes: multi,
			},
			TestNormalization: TestNormalization{
				OccurrencesCovered: covered, OccurrenceCoverage: ratioInt(covered, test.Occurrences),
			},
			SequenceComparison:    BuildSequenceComparison(rawMetrics, structuralMetrics, randomMetrics, config.MinN, config.MaxN, config.MaxContext),
			NewCrossLineSequences: NewCrossLineSequences(test, rawMetrics, structuralMetrics, config.MinN, config.MaxN),
		}
		output.Folds = append(output.Folds, fold)
		progress(config, fmt.Sprintf("validated fold %d/%d: %d TRAIN classes, %.2f%% TEST coverage", foldIndex+1, config.Folds, len(multi), fold.TestNormalization.OccurrenceCoverage*100))
	}
	output.ClassStability = BuildClassStability(models, eligibility)
	output.CrossValidationAggregate = AggregateFolds(output.Folds, config.MinN, config.MaxN)
	output.LeaveOneClassOut, output.MemberAblation = runAblations(corpus, fullModel, config)
	return output, nil
}

func validateConfig(config Config) error {
	if config.Folds < 2 {
		return fmt.Errorf("folds must be at least 2")
	}
	if config.Threshold < .70 || config.Threshold > 1 {
		return fmt.Errorf("threshold must be in [0.70,1]")
	}
	if config.MinTokenCount < 1 {
		return fmt.Errorf("min-token-count must be at least 1")
	}
	if config.RandomBaselines < 1 {
		return fmt.Errorf("random-baselines must be at least 1")
	}
	if config.MinN < 1 || config.MaxN < config.MinN {
		return fmt.Errorf("invalid n-gram range")
	}
	if config.MaxContext < 1 {
		return fmt.Errorf("max-context-length must be at least 1")
	}
	return nil
}

func loadClasses(path string) (normalization.ClassesOutput, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return normalization.ClassesOutput{}, "", err
	}
	var classes normalization.ClassesOutput
	if err := yaml.Unmarshal(data, &classes); err != nil {
		return normalization.ClassesOutput{}, "", err
	}
	sum := sha256.Sum256(data)
	return classes, hex.EncodeToString(sum[:]), nil
}

func selectModel(models []normalization.Model, threshold float64) (normalization.Model, error) {
	for _, model := range models {
		if math.Abs(model.Threshold-threshold) < 1e-12 {
			return model, nil
		}
	}
	return normalization.Model{}, fmt.Errorf("full-corpus class model for threshold %.2f is absent", threshold)
}

func multiMemberClasses(model normalization.Model) []normalization.Class {
	var result []normalization.Class
	for _, class := range model.Classes {
		if class.Size > 1 {
			result = append(result, class)
		}
	}
	return result
}

func countMembers(classes []normalization.Class) int {
	total := 0
	for _, class := range classes {
		total += class.Size
	}
	return total
}

func coveredOccurrences(corpus Corpus, model normalization.Model) int {
	members := make(map[string]bool)
	for _, class := range model.Classes {
		if class.Size < 2 {
			continue
		}
		for _, member := range class.Members {
			members[member.Token] = true
		}
	}
	total := 0
	for token, count := range corpus.Counts {
		if members[token] {
			total += count
		}
	}
	return total
}

func runAblations(corpus Corpus, model normalization.Model, config Config) (LeaveOneClassOut, []MemberAblation) {
	raw := AnalyzeSequences(corpus, config.MinN, config.MaxN, config.MaxContext)
	all := AnalyzeSequences(applyMapping(corpus, normalization.Mapping(model, "preserve")), config.MinN, config.MaxN, config.MaxContext)
	result := LeaveOneClassOut{
		Raw: simpleNGrams(raw, config.MinN, config.MaxN), AllClasses: simpleNGrams(all, config.MinN, config.MaxN),
		AllClassesMaxLength: all.MaxLength,
	}
	classes := multiMemberClasses(model)
	var memberAblations []MemberAblation
	for _, removed := range classes {
		mapping := mappingExcept(model, removed.ID, "")
		metrics := AnalyzeSequences(applyMapping(corpus, mapping), config.MinN, config.MaxN, config.MaxContext)
		members := memberNames(removed)
		variant := LeaveOneOutVariant{
			ClassRemoved: removed.ID, ClassMembers: members,
			ClassOccurrenceCoverage: ratioInt(classOccurrenceCount(removed), corpus.Occurrences),
			CrossLineNGrams:         simpleNGrams(metrics, config.MinN, config.MaxN),
			MaxCrossLineLength:      metrics.MaxLength,
		}
		for length := 1; length <= config.MaxContext; length++ {
			variant.RepeatedContextCoverage = append(variant.RepeatedContextCoverage, metrics.Contexts[length])
		}
		variant.ContributionN3 = all.CrossLine[3] - metrics.CrossLine[3]
		variant.ContributionN4 = all.CrossLine[4] - metrics.CrossLine[4]
		variant.ContributionFractionN3 = contributionFraction(variant.ContributionN3, all.CrossLine[3]-raw.CrossLine[3])
		variant.ContributionFractionN4 = contributionFraction(variant.ContributionN4, all.CrossLine[4]-raw.CrossLine[4])
		result.Variants = append(result.Variants, variant)

		if removed.Size > 2 {
			for _, restored := range removed.Members {
				mapping := mappingExcept(model, "", restored.Token)
				ablationMetrics := AnalyzeSequences(applyMapping(corpus, mapping), 3, 4, min(config.MaxContext, 3))
				remaining := make([]string, 0, removed.Size-1)
				for _, member := range removed.Members {
					if member.Token != restored.Token {
						remaining = append(remaining, member.Token)
					}
				}
				memberAblations = append(memberAblations, MemberAblation{
					ClassID: removed.ID, OriginalMembers: members, MemberRestored: restored.Token,
					RemainingNormalizedMembers: remaining, CrossLineNGrams: simpleNGrams(ablationMetrics, 3, 4),
				})
			}
		}
	}
	result.DominantClassFractionN3 = dominantFraction(result.Variants, true)
	result.DominantClassFractionN4 = dominantFraction(result.Variants, false)
	return result, memberAblations
}

func mappingExcept(model normalization.Model, removedClass, restoredMember string) map[string]string {
	mapping := make(map[string]string)
	for _, class := range model.Classes {
		if class.Size < 2 || class.ID == removedClass {
			continue
		}
		for _, member := range class.Members {
			if member.Token != restoredMember {
				mapping[member.Token] = class.ID
			}
		}
	}
	return mapping
}

func simpleNGrams(metrics sequenceMetrics, minN, maxN int) []SimpleNGram {
	result := make([]SimpleNGram, 0, maxN-minN+1)
	for n := minN; n <= maxN; n++ {
		result = append(result, SimpleNGram{N: n, CrossLineRepeated: metrics.CrossLine[n]})
	}
	return result
}

func memberNames(class normalization.Class) []string {
	result := make([]string, len(class.Members))
	for i, member := range class.Members {
		result[i] = member.Token
	}
	sort.Strings(result)
	return result
}

func classOccurrenceCount(class normalization.Class) int {
	total := 0
	for _, member := range class.Members {
		total += member.Count
	}
	return total
}

func contributionFraction(contribution, denominator int) *float64 {
	if denominator <= 0 {
		return nil
	}
	value := float64(contribution) / float64(denominator)
	return &value
}

func dominantFraction(variants []LeaveOneOutVariant, n3 bool) *float64 {
	var best *float64
	for _, variant := range variants {
		value := variant.ContributionFractionN4
		if n3 {
			value = variant.ContributionFractionN3
		}
		if value != nil && (best == nil || *value > *best) {
			copy := *value
			best = &copy
		}
	}
	return best
}

func progress(config Config, message string) {
	if config.Progress != nil {
		config.Progress(message)
	}
}
