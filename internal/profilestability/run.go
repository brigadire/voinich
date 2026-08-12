package profilestability

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"sort"

	"gopkg.in/yaml.v3"
	"zcore.dev/voinich/internal/normalization"
	"zcore.dev/voinich/internal/validation"
)

type pairKey struct{ a, b string }

type foldData struct {
	trainProfiles map[string]Profile
	testProfiles  map[string]Profile
	trainEligible map[string]bool
	testEligible  map[string]bool
	neighbors     map[string][]Neighbor
	model         normalization.Model
}

type bootstrapValues struct{ combined, position, left, right []float64 }

func Run(config Config) (Output, error) {
	if err := validateConfig(config); err != nil {
		return Output{}, err
	}
	corpus, inputHash, err := validation.LoadCorpus(config.InputPath)
	if err != nil {
		return Output{}, err
	}
	classes, classesHash, err := readClasses(config.ClassesPath)
	if err != nil {
		return Output{}, err
	}
	referenceModel, err := selectModel(classes.Models, config.Threshold)
	if err != nil {
		return Output{}, err
	}
	fullProfiles := BuildProfiles(corpus)
	fullEligible := Eligible(fullProfiles, config.MinTokenCount)
	fullNeighbors := buildAllNeighbors(fullProfiles, fullEligible, config.Neighbors)
	candidates := make(map[pairKey]bool)
	addNeighborPairs(candidates, fullNeighbors)
	for i, tokenA := range fullEligible {
		for _, tokenB := range fullEligible[i+1:] {
			if Compare(fullProfiles[tokenA], fullProfiles[tokenB]).Similarity >= config.Threshold {
				candidates[makePair(tokenA, tokenB)] = true
			}
		}
	}
	addClassPairs(candidates, referenceModel)

	foldIndexes, err := validation.SplitFolds(corpus.Lines, config.Folds, config.FoldSeed)
	if err != nil {
		return Output{}, err
	}
	folds := make([]foldData, 0, config.Folds)
	for foldIndex, indexes := range foldIndexes {
		train, test, err := validation.Partition(corpus, indexes)
		if err != nil {
			return Output{}, err
		}
		trainProfiles, testProfiles := BuildProfiles(train), BuildProfiles(test)
		trainTokens, testTokens := Eligible(trainProfiles, config.MinTokenCount), Eligible(testProfiles, config.MinTokenCount)
		neighbors := buildAllNeighbors(trainProfiles, trainTokens, config.Neighbors)
		addNeighborPairs(candidates, neighbors)
		for i, tokenA := range trainTokens {
			for _, tokenB := range trainTokens[i+1:] {
				if Compare(trainProfiles[tokenA], trainProfiles[tokenB]).Similarity >= config.Threshold {
					candidates[makePair(tokenA, tokenB)] = true
				}
			}
		}
		model, _, err := validation.BuildTrainModel(train, validation.Config{Threshold: config.Threshold, MinTokenCount: config.MinTokenCount})
		if err != nil {
			return Output{}, fmt.Errorf("fold %d classes: %w", foldIndex+1, err)
		}
		addClassPairs(candidates, model)
		folds = append(folds, foldData{trainProfiles: trainProfiles, testProfiles: testProfiles, trainEligible: boolSet(trainTokens), testEligible: boolSet(testTokens), neighbors: neighbors, model: model})
		progress(config, fmt.Sprintf("profiles fold %d/%d: TRAIN eligible=%d TEST eligible=%d", foldIndex+1, config.Folds, len(trainTokens), len(testTokens)))
	}

	tokens, neighborResults := buildTokenResults(fullProfiles, fullEligible, fullNeighbors, folds, config)
	neighborByToken := make(map[string]NeighborStability, len(neighborResults))
	for _, item := range neighborResults {
		neighborByToken[item.Token] = item
	}
	pairResults := buildPairResults(candidates, fullProfiles, folds, neighborByToken, config)
	pairByKey := make(map[pairKey]PairStability, len(pairResults))
	for _, item := range pairResults {
		pairByKey[makePair(item.TokenA, item.TokenB)] = item
	}

	bootstrap := runBootstrap(corpus, candidates, config)
	bootstrapByKey := make(map[pairKey]BootstrapPair, len(bootstrap))
	for _, item := range bootstrap {
		bootstrapByKey[makePair(item.TokenA, item.TokenB)] = item
	}
	referenceClasses, diagnostics := buildReferenceReports(referenceModel, pairByKey, bootstrapByKey, neighborByToken, fullProfiles)
	frequency, pairFrequency := buildFrequencyReports(tokens, neighborByToken, pairResults)
	summary := buildSummary(tokens, neighborResults, pairResults, bootstrap, referenceModel, folds)

	return Output{
		Meta:       Meta{Input: config.InputPath, Classes: config.ClassesPath, PhysicalLines: len(corpus.Lines), TokenOccurrences: corpus.Occurrences, UniqueTokens: len(corpus.Counts), InputSHA256: inputHash, ClassesSHA256: classesHash},
		Parameters: Parameters{Folds: config.Folds, FoldSeed: config.FoldSeed, MinTokenCount: config.MinTokenCount, Neighbors: config.Neighbors, BootstrapRuns: config.BootstrapRuns, BootstrapSeed: config.BootstrapSeed, Threshold: config.Threshold, ThresholdMargin: config.ThresholdMargin, Thresholds: []float64{.70, .75, .80, .85, .90}},
		Methodology: Methodology{
			Profiles:          "token count plus complete absolute-position, immediate predecessor, and immediate successor distributions calculated independently from each corpus sample",
			Similarity:        "production structural formula unchanged: position=1-Jensen-Shannon divergence; left/right=cosine similarity; combined=(position+left+right)/3",
			Eligibility:       "count >= min_token_count independently in full, each TRAIN, each TEST, and each bootstrap sample; missing observations are not treated as instability",
			Split:             "the same deterministic line shuffle and round-robin fold assignment as structural-validate; TRAIN and TEST profiles are independent",
			TrainTest:         "same-token TRAIN-to-TEST similarities are reported only when the token is eligible on both sides",
			Neighbors:         "top-K among tokens eligible in the same sample, ordered by combined similarity descending then token lexicographically",
			RankOverlap:       "top-N overlap is intersection/min(N, available); top-K set stability additionally uses Jaccard intersection/union",
			RankCorrelation:   "Spearman correlation of actual top-K ranks on their common items; omitted from the mean when fewer than three common items exist",
			Bootstrap:         "line sampling with replacement, preserving the physical line count; each run recomputes profiles and includes a pair only when both sampled counts meet min_token_count",
			CandidatePairs:    "deduplicated union of full/fold similarity>=threshold pairs, full/fold top-K neighbor pairs, and threshold-0.70 reference/fold class pairs",
			ThresholdCrossing: "folds_above uses similarity>=threshold; crossing_count is the number of above/below state changes between consecutive observed folds; near threshold means abs(similarity-threshold)<=threshold_margin",
			ComponentAblation: "diagnostic two-component arithmetic means only for reference-class pairs; no classes or sequence analyses are built from ablated values",
			Interpretation:    "formal stability measurements only; the tool does not alter similarity weights, thresholds, clustering, or infer token meaning",
		},
		Summary: summary, TokenProfileStability: tokens, NearestNeighborStability: neighborResults,
		PairSimilarityStability: pairResults, BootstrapPairUncertainty: bootstrap,
		FrequencyDependence: frequency, PairFrequencyDependence: pairFrequency,
		ReferenceClasses: referenceClasses, ComponentDiagnostics: diagnostics,
	}, nil
}

func validateConfig(config Config) error {
	if config.Folds < 2 {
		return fmt.Errorf("folds must be at least 2")
	}
	if config.MinTokenCount < 1 {
		return fmt.Errorf("min-token-count must be at least 1")
	}
	if config.Neighbors < 1 {
		return fmt.Errorf("neighbors must be at least 1")
	}
	if config.BootstrapRuns < 1 {
		return fmt.Errorf("bootstrap-runs must be at least 1")
	}
	if config.Threshold < 0 || config.Threshold > 1 {
		return fmt.Errorf("threshold must be in [0,1]")
	}
	if config.ThresholdMargin < 0 {
		return fmt.Errorf("threshold-margin cannot be negative")
	}
	return nil
}

func readClasses(path string) (classFile, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return classFile{}, "", err
	}
	var result classFile
	if err := yaml.Unmarshal(data, &result); err != nil {
		return classFile{}, "", err
	}
	sum := sha256.Sum256(data)
	return result, hex.EncodeToString(sum[:]), nil
}

func selectModel(models []normalization.Model, threshold float64) (normalization.Model, error) {
	for _, model := range models {
		if math.Abs(model.Threshold-threshold) < 1e-12 {
			return model, nil
		}
	}
	return normalization.Model{}, fmt.Errorf("class model threshold %.2f is absent", threshold)
}

func buildAllNeighbors(profiles map[string]Profile, eligible []string, k int) map[string][]Neighbor {
	result := make(map[string][]Neighbor, len(eligible))
	for _, token := range eligible {
		result[token] = NearestNeighbors(profiles, token, eligible, k)
	}
	return result
}

func boolSet(tokens []string) map[string]bool {
	result := make(map[string]bool, len(tokens))
	for _, token := range tokens {
		result[token] = true
	}
	return result
}

func makePair(a, b string) pairKey {
	if a > b {
		a, b = b, a
	}
	return pairKey{a, b}
}

func addNeighborPairs(target map[pairKey]bool, neighbors map[string][]Neighbor) {
	for token, items := range neighbors {
		for _, item := range items {
			target[makePair(token, item.Token)] = true
		}
	}
}

func addClassPairs(target map[pairKey]bool, model normalization.Model) {
	for _, class := range model.Classes {
		if class.Size < 2 {
			continue
		}
		for i := 0; i < len(class.Members); i++ {
			for j := i + 1; j < len(class.Members); j++ {
				target[makePair(class.Members[i].Token, class.Members[j].Token)] = true
			}
		}
	}
}

func progress(config Config, message string) {
	if config.Progress != nil {
		config.Progress(message)
	}
}

func sortedPairs(pairs map[pairKey]bool) []pairKey {
	result := make([]pairKey, 0, len(pairs))
	for pair := range pairs {
		result = append(result, pair)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].a != result[j].a {
			return result[i].a < result[j].a
		}
		return result[i].b < result[j].b
	})
	return result
}
