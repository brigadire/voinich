package structuralreliability

import (
	"fmt"
	"sort"

	"zcore.dev/voinich/internal/profilestability"
	"zcore.dev/voinich/internal/validation"
)

type thresholdComputation struct {
	items         []tokenMetric
	fullEligible  []string
	fullNeighbors map[string][]profilestability.Neighbor
}

// Run measures the statistical reproducibility of the existing structural
// profile similarity (position/left/right, unchanged) as a function of how
// many observations back each token or pair. It builds no classes, changes
// no formula, weights or threshold, and does not touch sequence
// normalization - see task07 sections 2, 20 and 31.
func Run(config Config) (Output, error) {
	if err := validateConfig(config); err != nil {
		return Output{}, err
	}
	corpus, inputHash, err := validation.LoadCorpus(config.InputPath)
	if err != nil {
		return Output{}, err
	}
	classes, classesHash, err := ReadClasses(config.ClassesPath)
	if err != nil {
		return Output{}, err
	}
	referenceModel, err := SelectModel(classes.Models, config.Threshold)
	if err != nil {
		return Output{}, err
	}

	fullProfiles := profilestability.BuildProfiles(corpus)
	fullWs := profilestability.PrecomputeAll(fullProfiles)
	folds, err := BuildFolds(corpus, config.Folds, config.FoldSeed)
	if err != nil {
		return Output{}, err
	}
	occurrences := CollectOccurrences(corpus)
	progress(config, fmt.Sprintf("loaded corpus: lines=%d occurrences=%d unique_tokens=%d", len(corpus.Lines), corpus.Occurrences, len(corpus.Counts)))

	thresholds := uniqueSortedInts(append([]int{config.MinTokenCount}, config.CountThresholds...))
	cache := make(map[int]thresholdComputation, len(thresholds))
	for _, t := range thresholds {
		items, fullEligible, fullNeighbors := buildTokenMetrics(fullProfiles, fullWs, folds, t, config.Neighbors)
		cache[t] = thresholdComputation{items: items, fullEligible: fullEligible, fullNeighbors: fullNeighbors}
		progress(config, fmt.Sprintf("token stability min_count=%d eligible=%d", t, len(fullEligible)))
	}
	base := cache[config.MinTokenCount]
	baseByToken := make(map[string]tokenMetric, len(base.items))
	for _, item := range base.items {
		baseByToken[item.Token] = item
	}

	masterPairSet := buildMasterPairs(fullWs, base.fullEligible, base.fullNeighbors, referenceModel, config.Threshold)
	masterPairs := sortedPairs(masterPairSet)
	baseFoldValues := pairFoldSimilarities(masterPairs, folds, config.MinTokenCount)
	bootstrap := runBootstrap(corpus, masterPairs, config.BootstrapRuns, config.BootstrapSeed, config.MinTokenCount, config.Threshold, config.Progress)
	progress(config, fmt.Sprintf("master candidate pairs=%d bootstrap runs=%d", len(masterPairs), config.BootstrapRuns))

	subsampling := runSubsampling(occurrences, fullProfiles, fullWs, config)
	reliabilityCurves := extractReliabilityCurves(subsampling)
	positionTable := NewReliabilityTable(reliabilityCurves.Position)
	leftTable := NewReliabilityTable(reliabilityCurves.LeftContext)
	rightTable := NewReliabilityTable(reliabilityCurves.RightContext)

	pairMetrics := make([]ContinuousPairMetric, 0, len(masterPairs))
	byPair := make(map[pairKey]ContinuousPairMetric, len(masterPairs))
	for _, pair := range masterPairs {
		countA, countB := fullProfiles[pair.a].Count, fullProfiles[pair.b].Count
		full := profilestability.CompareSorted(fullWs[pair.a], fullWs[pair.b])
		metric := ContinuousPairMetric{
			TokenA: pair.a, TokenB: pair.b, CountA: countA, CountB: countB,
			MinCount: min(countA, countB), GeometricMeanCount: GeometricMean(float64(countA), float64(countB)),
			FullSimilarity: full.Similarity, FullPositionSimilarity: full.PositionSimilarity,
			FullLeftSimilarity: full.LeftSimilarity, FullRightSimilarity: full.RightSimilarity,
		}
		if values := baseFoldValues[pair]; len(values) > 0 {
			stddev := profilestability.Summarize(values, false).Stddev
			metric.FoldSimilarityStddev, metric.FoldObservations = &stddev, len(values)
		}
		if boot := bootstrap[pair]; boot.Observations > 0 {
			mean, ciWidth, probability := boot.Mean, boot.CIWidth, boot.ProbabilityAboveThreshold
			metric.BootstrapObservations = boot.Observations
			metric.BootstrapMean, metric.BootstrapCIWidth, metric.BootstrapProbabilityAbove070 = &mean, &ciWidth, &probability
		}
		metric.PositionReliabilityPair = PairReliability(positionTable, countA, countB)
		metric.LeftReliabilityPair = PairReliability(leftTable, countA, countB)
		metric.RightReliabilityPair = PairReliability(rightTable, countA, countB)
		metric.PositionSupport = ComponentSupport(full.PositionSimilarity, metric.PositionReliabilityPair)
		metric.LeftSupport = ComponentSupport(full.LeftSimilarity, metric.LeftReliabilityPair)
		metric.RightSupport = ComponentSupport(full.RightSimilarity, metric.RightReliabilityPair)
		pairMetrics = append(pairMetrics, metric)
		byPair[pair] = metric
	}

	cumulative := make([]CumulativeThreshold, 0, len(config.CountThresholds))
	for _, t := range sortedUnique(config.CountThresholds) {
		comp := cache[t]
		self, trainTest, neighbors := aggregateTokenMetrics(comp.items)
		pairsAtT := filterPairs(masterPairs, fullProfiles, func(count int) bool { return count >= t })
		foldValuesAtT := pairFoldSimilarities(pairsAtT, folds, t)
		cumulative = append(cumulative, CumulativeThreshold{
			MinCount: t, EligibleTokens: len(comp.fullEligible),
			SelfProfileTrainTrain: self, TrainTest: trainTest, NearestNeighbors: neighbors,
			Pairs:   buildPairStabilitySummary(pairsAtT, foldValuesAtT, bootstrap, config.Threshold),
			CIWidth: buildCIWidthSummary(pairsAtT, bootstrap),
		})
	}

	bins := make([]FrequencyBin, 0, len(config.CountThresholds))
	sortedThresholds := sortedUnique(config.CountThresholds)
	for i, lower := range sortedThresholds {
		var upper *int
		if i+1 < len(sortedThresholds) {
			next := sortedThresholds[i+1]
			upper = &next
		}
		comp := cache[lower]
		var binItems []tokenMetric
		for _, item := range comp.items {
			if item.FullCount >= lower && (upper == nil || item.FullCount < *upper) {
				binItems = append(binItems, item)
			}
		}
		self, trainTest, neighbors := aggregateTokenMetrics(binItems)
		pairsInBin := filterPairs(masterPairs, fullProfiles, func(count int) bool { return count >= lower && (upper == nil || count < *upper) })
		foldValuesInBin := pairFoldSimilarities(pairsInBin, folds, lower)
		bins = append(bins, FrequencyBin{
			Bin: binLabel(lower, upper), LowerBound: lower, UpperBound: upper, Tokens: len(binItems),
			SelfProfileTrainTrain: self, TrainTest: trainTest, NearestNeighbors: neighbors,
			Pairs:   buildPairStabilitySummary(pairsInBin, foldValuesInBin, bootstrap, config.Threshold),
			CIWidth: buildCIWidthSummary(pairsInBin, bootstrap),
		})
	}

	diversity := buildContextDiversity(fullProfiles, base.fullEligible)
	correlations := Correlations{
		FrequencyVsTokenStability:   buildFrequencyVsTokenStability(base.items),
		FrequencyVsPairStability:    buildFrequencyVsPairStability(pairMetrics),
		ContextDiversityVsStability: buildContextDiversityVsStability(diversity, baseByToken),
	}

	referencePairs := buildReferencePairs(referenceClassPairs(referenceModel), byPair, bootstrap)

	classCount, tokensInClasses := 0, 0
	for _, class := range referenceModel.Classes {
		if class.Size < 2 {
			continue
		}
		classCount++
		tokensInClasses += class.Size
	}

	largestThreshold := sortedThresholds[len(sortedThresholds)-1]

	return Output{
		Meta: Meta{
			Input: config.InputPath, Classes: config.ClassesPath, PhysicalLines: len(corpus.Lines),
			TokenOccurrences: corpus.Occurrences, Transitions: corpus.Transitions, UniqueTokens: len(corpus.Counts),
			InputSHA256: inputHash, ClassesSHA256: classesHash,
		},
		Parameters: Parameters{
			Folds: config.Folds, FoldSeed: config.FoldSeed, MinTokenCount: config.MinTokenCount, Neighbors: config.Neighbors,
			BootstrapRuns: config.BootstrapRuns, BootstrapSeed: config.BootstrapSeed, Threshold: config.Threshold, ThresholdMargin: config.ThresholdMargin,
			CountThresholds: sortedThresholds, SubsampleMinFullCount: config.SubsampleMinFullCount, SubsampleSizes: subsampling.SampleSizes,
			SubsampleRuns: config.SubsampleRuns, SubsampleSeed: config.SubsampleSeed,
		},
		Methodology: Methodology{
			Scope:                "measures statistical reproducibility of the existing position/left/right structural similarity as a function of observation count; the similarity formula, weights, threshold and complete-link clustering are unchanged and reused verbatim from internal/profilestability and internal/normalization",
			Similarity:           "position=1-Jensen-Shannon divergence; left/right=cosine similarity; combined=(position+left+right)/3, computed by profilestability.Compare without modification",
			CumulativeThresholds: "for each min_count, eligibility, TRAIN/TEST/full neighbor geometry and folds are recomputed exactly as structural-profile-stability would with -min-token-count=min_count; tokens/pairs below the threshold are excluded, not treated as unstable",
			FrequencyBins:        "independent, non-overlapping intervals reusing the cumulative-threshold computation whose min_count equals the bin's lower bound, filtered to full_count within the bin",
			NearestNeighbors:     "top-K neighbor geometry is recomputed within each threshold's own eligible set, so low-frequency tokens never distort a high-frequency subspace (task section 6)",
			Bootstrap:            "line sampling with replacement, preserving physical line count, at the base min_token_count; computed once and reused by filtering pairs into every downstream frequency grouping",
			Subsampling:          "occurrence-level subsampling (position, predecessor, successor) of real occurrences for tokens with full_count>=subsample_min_full_count, compared against the full-corpus reference profile; no synthetic text is created",
			ReliabilityCurve:     "reliability(component,n): exact lookup at tested sample sizes, linear interpolation in log2(n) between them, clamped to the smallest/largest tested value outside that range - never extrapolated beyond the observed maximum",
			PairReliability:      "diagnostic only: geometric mean of the two members' component reliability, and component_similarity*component_reliability; neither is folded back into the similarity formula, no classes are built from them",
			Interpretation:       "reliability here means only the statistical reproducibility of a measured structural profile, not semantic equivalence; this tool does not normalize, build classes, or run sequence matching",
		},
		Summary: Summary{
			TokenOccurrences: corpus.Occurrences, PhysicalLines: len(corpus.Lines), Transitions: corpus.Transitions,
			ReferenceModel: ReferenceModel{
				Threshold: referenceModel.Threshold, Classes: classCount, TokensInClasses: tokensInClasses,
				OccurrenceCoverage: referenceModel.Stats.TokenOccurrenceCoverage,
			},
			BaseEligibleTokens: len(base.fullEligible), CandidatePairs: len(masterPairs),
			LargestCumulativeThreshold: largestThreshold, LargestThresholdEligibleTokens: len(cache[largestThreshold].fullEligible),
		},
		CumulativeThresholds:   cumulative,
		FrequencyBins:          bins,
		ContinuousTokenMetrics: toContinuousMetrics(base.items),
		ContinuousPairMetrics:  pairMetrics,
		Correlations:           correlations,
		Subsampling:            subsampling,
		ReliabilityCurves:      reliabilityCurves,
		ReliabilityThresholds: ReliabilityThresholds{
			Position:     ReliabilityThresholdsFor(subsampling.SampleSizes, reliabilityCurves.Position),
			LeftContext:  ReliabilityThresholdsFor(subsampling.SampleSizes, reliabilityCurves.LeftContext),
			RightContext: ReliabilityThresholdsFor(subsampling.SampleSizes, reliabilityCurves.RightContext),
		},
		ContextDiversity: ContextDiversity{Tokens: diversity},
		ReferencePairs:   referencePairs,
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
	if len(config.CountThresholds) == 0 {
		return fmt.Errorf("count-thresholds must not be empty")
	}
	for _, t := range config.CountThresholds {
		if t < 1 {
			return fmt.Errorf("count-thresholds must be positive, got %d", t)
		}
	}
	if config.SubsampleMinFullCount < 1 {
		return fmt.Errorf("subsample-min-full-count must be at least 1")
	}
	if config.SubsampleRuns < 1 {
		return fmt.Errorf("subsample-runs must be at least 1")
	}
	return nil
}

func progress(config Config, message string) {
	if config.Progress != nil {
		config.Progress(message)
	}
}

func uniqueSortedInts(values []int) []int {
	set := make(map[int]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	result := make([]int, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Ints(result)
	return result
}

func sortedUnique(values []int) []int { return uniqueSortedInts(values) }

func filterPairs(pairs []pairKey, full map[string]profilestability.Profile, predicate func(int) bool) []pairKey {
	var result []pairKey
	for _, pair := range pairs {
		if predicate(full[pair.a].Count) && predicate(full[pair.b].Count) {
			result = append(result, pair)
		}
	}
	return result
}

func binLabel(lower int, upper *int) string {
	if upper == nil {
		return fmt.Sprintf("%d+", lower)
	}
	return fmt.Sprintf("%d-%d", lower, *upper-1)
}

func buildPairStabilitySummary(pairs []pairKey, foldValues map[pairKey][]float64, bootstrap map[pairKey]bootstrapResult, threshold float64) PairStabilitySummary {
	var stddevs []float64
	crossingPairs, comparablePairs := 0, 0
	for _, pair := range pairs {
		values := foldValues[pair]
		if len(values) == 0 {
			continue
		}
		stddevs = append(stddevs, profilestability.Summarize(values, false).Stddev)
		if len(values) < 2 {
			continue
		}
		comparablePairs++
		for i := 1; i < len(values); i++ {
			if (values[i-1] >= threshold) != (values[i] >= threshold) {
				crossingPairs++
				break
			}
		}
	}
	available, above095 := 0, 0
	for _, pair := range pairs {
		boot, ok := bootstrap[pair]
		if !ok || boot.Observations == 0 {
			continue
		}
		available++
		if boot.ProbabilityAboveThreshold >= .95 {
			above095++
		}
	}
	summary := PairStabilitySummary{Pairs: len(pairs), SimilarityStddev: SummarizePercentileStat(stddevs), BootstrapPairsAvailable: available}
	if comparablePairs > 0 {
		summary.ThresholdCrossing070Fraction = float64(crossingPairs) / float64(comparablePairs)
	}
	if available > 0 {
		summary.BootstrapProbabilityAbove070Ge095Fraction = float64(above095) / float64(available)
	}
	return summary
}

func buildCIWidthSummary(pairs []pairKey, bootstrap map[pairKey]bootstrapResult) CIWidthSummary {
	var widths []float64
	for _, pair := range pairs {
		boot, ok := bootstrap[pair]
		if !ok || boot.Observations == 0 {
			continue
		}
		widths = append(widths, boot.CIWidth)
	}
	if len(widths) == 0 {
		return CIWidthSummary{}
	}
	stat := SummarizeStat(widths)
	return CIWidthSummary{MeanCIWidth: stat.Mean, MedianCIWidth: stat.Median, Percentile90CIWidth: PercentileOf(widths, .90), Observations: len(widths)}
}

func extractReliabilityCurves(subsampling Subsampling) ReliabilityCurves {
	curves := ReliabilityCurves{Position: map[int]float64{}, LeftContext: map[int]float64{}, RightContext: map[int]float64{}}
	for _, result := range subsampling.Results {
		curves.Position[result.SampleSize] = result.Position.MeanSimilarity
		curves.LeftContext[result.SampleSize] = result.LeftContext.MeanSimilarity
		curves.RightContext[result.SampleSize] = result.RightContext.MeanSimilarity
	}
	return curves
}
