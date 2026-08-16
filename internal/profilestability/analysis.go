package profilestability

import (
	"math"
	"math/rand"

	"zcore.dev/voinich/internal/validation"
)

func buildTokenResults(full map[string]Profile, eligible []string, fullNeighbors map[string][]Neighbor, folds []foldData, config Config) ([]TokenProfileStability, []NeighborStability) {
	profiles := make([]TokenProfileStability, 0, len(eligible))
	neighbors := make([]NeighborStability, 0, len(eligible))
	for _, token := range eligible {
		item := TokenProfileStability{Token: token, Count: full[token].Count, EligibleFull: true, FullCorpusNeighbors: fullNeighbors[token]}
		var position, left, right, ttPosition, ttLeft, ttRight []float64
		for fold := range folds {
			if folds[fold].trainEligible[token] {
				item.EligibleTrainFolds = append(item.EligibleTrainFolds, fold+1)
			}
			if folds[fold].testEligible[token] {
				item.EligibleTestFolds = append(item.EligibleTestFolds, fold+1)
			}
			if folds[fold].trainEligible[token] && folds[fold].testEligible[token] {
				value := CompareSorted(folds[fold].trainWs[token], folds[fold].testWs[token])
				item.TrainTest = append(item.TrainTest, TrainTestFold{Fold: fold + 1, PositionSimilarity: value.PositionSimilarity, LeftSimilarity: value.LeftSimilarity, RightSimilarity: value.RightSimilarity})
				ttPosition = append(ttPosition, value.PositionSimilarity)
				ttLeft = append(ttLeft, value.LeftSimilarity)
				ttRight = append(ttRight, value.RightSimilarity)
			}
			for other := fold + 1; other < len(folds); other++ {
				if !folds[fold].trainEligible[token] || !folds[other].trainEligible[token] {
					continue
				}
				value := CompareSorted(folds[fold].trainWs[token], folds[other].trainWs[token])
				position = append(position, value.PositionSimilarity)
				left = append(left, value.LeftSimilarity)
				right = append(right, value.RightSimilarity)
			}
		}
		item.PositionStability = Summarize(position, false)
		item.LeftContextStability = Summarize(left, false)
		item.RightContextStability = Summarize(right, false)
		item.TrainTestPositionSimilarity = Summarize(ttPosition, false)
		item.TrainTestLeftSimilarity = Summarize(ttLeft, false)
		item.TrainTestRightSimilarity = Summarize(ttRight, false)
		profiles = append(profiles, item)
		neighbors = append(neighbors, buildNeighborStability(token, fullNeighbors[token], folds, config.Neighbors))
	}
	return profiles, neighbors
}

func buildNeighborStability(token string, full []Neighbor, folds []foldData, k int) NeighborStability {
	result := NeighborStability{Token: token}
	var jaccards, top3, top5, top10, rhos, common []float64
	top1Equal := 0
	for i := 0; i < len(folds); i++ {
		left, exists := folds[i].neighbors[token]
		if !exists {
			continue
		}
		for j := i + 1; j < len(folds); j++ {
			right, exists := folds[j].neighbors[token]
			if !exists {
				continue
			}
			result.FoldPairComparisons++
			jaccards = append(jaccards, Jaccard(left, right))
			top3 = append(top3, OverlapAt(left, right, 3))
			top5 = append(top5, OverlapAt(left, right, 5))
			top10 = append(top10, OverlapAt(left, right, k))
			if len(left) > 0 && len(right) > 0 && left[0].Token == right[0].Token {
				top1Equal++
			}
			if rho, n, ok := spearmanCommon(left, right); ok {
				rhos = append(rhos, rho)
				common = append(common, float64(n))
			}
		}
	}
	if len(jaccards) > 0 {
		summary := Summarize(jaccards, false)
		result.MeanJaccard, result.MinJaccard, result.MaxJaccard = summary.Mean, summary.Min, summary.Max
		result.Top1SameFraction = float64(top1Equal) / float64(len(jaccards))
		result.Top3OverlapMean = Summarize(top3, false).Mean
		result.Top5OverlapMean = Summarize(top5, false).Mean
		result.Top10OverlapMean = Summarize(top10, false).Mean
	}
	if len(full) > 0 {
		result.FullTop1Neighbor = full[0].Token
		for _, fold := range folds {
			if fold.trainEligible[token] && fold.trainEligible[result.FullTop1Neighbor] {
				result.FoldsWhereTop1Eligible++
				if items := fold.neighbors[token]; len(items) > 0 && items[0].Token == result.FullTop1Neighbor {
					result.FoldsWhereSameTop1++
				}
			}
		}
		if result.FoldsWhereTop1Eligible > 0 {
			result.Top1RecoveryFraction = float64(result.FoldsWhereSameTop1) / float64(result.FoldsWhereTop1Eligible)
		}
	}
	if len(rhos) > 0 {
		rho := Summarize(rhos, false)
		result.RankCorrelation = RankCorrelation{Comparisons: len(rhos), MeanCommonItems: Summarize(common, false).Mean, MeanSpearmanRho: rho.Mean, MinSpearmanRho: rho.Min, MaxSpearmanRho: rho.Max}
	}
	return result
}

func buildPairResults(candidates map[pairKey]bool, full map[string]SortedProfile, folds []foldData, neighbors map[string]NeighborStability, config Config) []PairStability {
	thresholds := []float64{.70, .75, .80, .85, .90}
	result := make([]PairStability, 0, len(candidates))
	for _, pair := range sortedPairs(candidates) {
		countA, countB := full[pair.a].profile.Count, full[pair.b].profile.Count
		item := PairStability{TokenA: pair.a, TokenB: pair.b, TokenACount: countA, TokenBCount: countB, MinCount: min(countA, countB), GeometricMeanCount: math.Sqrt(float64(countA) * float64(countB)), Full: CompareSorted(full[pair.a], full[pair.b])}
		var values, margins []float64
		for foldIndex, fold := range folds {
			if fold.trainEligible[pair.a] && fold.trainEligible[pair.b] {
				value := CompareSorted(fold.trainWs[pair.a], fold.trainWs[pair.b])
				item.Folds = append(item.Folds, FoldSimilarity{Fold: foldIndex + 1, Components: value})
				values = append(values, value.Similarity)
				margins = append(margins, value.Similarity-config.Threshold)
			}
		}
		item.Summary = Summarize(values, false)
		item.Thresholds = thresholdCrossings(item.Folds, thresholds)
		if len(margins) > 0 {
			summary := Summarize(margins, false)
			item.MeanMargin, item.MinMargin, item.MaxMargin = summary.Mean, summary.Min, summary.Max
			near := 0
			for _, margin := range margins {
				if math.Abs(margin) <= config.ThresholdMargin {
					near++
				}
			}
			item.NearThresholdFraction = float64(near) / float64(len(margins))
		}
		item.MeanMemberNeighborJaccard = (neighbors[pair.a].MeanJaccard + neighbors[pair.b].MeanJaccard) / 2
		result = append(result, item)
	}
	return result
}

func thresholdCrossings(folds []FoldSimilarity, thresholds []float64) []ThresholdCrossing {
	result := make([]ThresholdCrossing, 0, len(thresholds))
	for _, threshold := range thresholds {
		cross := ThresholdCrossing{Threshold: threshold}
		previous, hasPrevious := false, false
		for _, fold := range folds {
			above := fold.Similarity >= threshold
			if above {
				cross.FoldsAboveThreshold++
			} else {
				cross.FoldsBelowThreshold++
			}
			if hasPrevious && above != previous {
				cross.ThresholdCrossingCount++
			}
			previous, hasPrevious = above, true
		}
		result = append(result, cross)
	}
	return result
}

func runBootstrap(corpus validation.Corpus, candidates map[pairKey]bool, config Config) []BootstrapPair {
	pairs := sortedPairs(candidates)
	values := make(map[pairKey]*bootstrapValues, len(pairs))
	for _, pair := range pairs {
		values[pair] = &bootstrapValues{}
	}
	rng := rand.New(rand.NewSource(config.BootstrapSeed))
	for run := 0; run < config.BootstrapRuns; run++ {
		sample := validation.Corpus{Counts: make(map[string]int)}
		for index := 0; index < len(corpus.Lines); index++ {
			source := corpus.Lines[rng.Intn(len(corpus.Lines))]
			tokens := append([]string(nil), source.Tokens...)
			sample.Lines = append(sample.Lines, validation.Line{ID: index + 1, Tokens: tokens})
		}
		profiles := BuildProfiles(sample)
		ws := PrecomputeAll(profiles)
		for _, pair := range pairs {
			if profiles[pair.a].Count < config.MinTokenCount || profiles[pair.b].Count < config.MinTokenCount {
				continue
			}
			component := CompareSorted(ws[pair.a], ws[pair.b])
			item := values[pair]
			item.combined = append(item.combined, component.Similarity)
			item.position = append(item.position, component.PositionSimilarity)
			item.left = append(item.left, component.LeftSimilarity)
			item.right = append(item.right, component.RightSimilarity)
		}
		if (run+1)%25 == 0 || run+1 == config.BootstrapRuns {
			progress(config, "bootstrap "+fmtInt(run+1)+"/"+fmtInt(config.BootstrapRuns))
		}
	}
	result := make([]BootstrapPair, 0, len(pairs))
	for _, pair := range pairs {
		item := values[pair]
		entry := BootstrapPair{TokenA: pair.a, TokenB: pair.b, Similarity: Summarize(item.combined, true), PositionSimilarity: Summarize(item.position, true), LeftContextSimilarity: Summarize(item.left, true), RightContextSimilarity: Summarize(item.right, true)}
		if len(item.combined) > 0 {
			above := 0
			for _, value := range item.combined {
				if value >= config.Threshold {
					above++
				}
			}
			entry.ProbabilityAboveThreshold = float64(above) / float64(len(item.combined))
		}
		entry.MostVariableComponent = mostVariable(entry.PositionSimilarity.Stddev, entry.LeftContextSimilarity.Stddev, entry.RightContextSimilarity.Stddev)
		result = append(result, entry)
	}
	return result
}

func fmtInt(value int) string {
	if value == 0 {
		return "0"
	}
	digits := make([]byte, 0, 10)
	for value > 0 {
		digits = append(digits, byte('0'+value%10))
		value /= 10
	}
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	return string(digits)
}

func mostVariable(position, left, right float64) string {
	if position >= left && position >= right {
		return "position"
	}
	if left >= right {
		return "left_context"
	}
	return "right_context"
}

func buildFrequencyReports(tokens []TokenProfileStability, neighborMap map[string]NeighborStability, pairs []PairStability) ([]FrequencyBin, []PairFrequencyBin) {
	tokenGroups := make(map[string][]TokenProfileStability)
	for _, item := range tokens {
		tokenGroups[frequencyBin(item.Count)] = append(tokenGroups[frequencyBin(item.Count)], item)
	}
	var tokenResult []FrequencyBin
	for _, bin := range frequencyBins {
		items := tokenGroups[bin]
		entry := FrequencyBin{Bin: bin, Tokens: len(items)}
		var position, left, right, jaccard, recovery []float64
		for _, item := range items {
			if item.PositionStability.Observations > 0 {
				position = append(position, item.PositionStability.Mean)
			}
			if item.LeftContextStability.Observations > 0 {
				left = append(left, item.LeftContextStability.Mean)
			}
			if item.RightContextStability.Observations > 0 {
				right = append(right, item.RightContextStability.Mean)
			}
			neighbor := neighborMap[item.Token]
			if neighbor.FoldPairComparisons > 0 {
				jaccard = append(jaccard, neighbor.MeanJaccard)
			}
			if neighbor.FoldsWhereTop1Eligible > 0 {
				recovery = append(recovery, neighbor.Top1RecoveryFraction)
			}
		}
		entry.MeanPositionStability = Summarize(position, false).Mean
		entry.MeanLeftStability = Summarize(left, false).Mean
		entry.MeanRightStability = Summarize(right, false).Mean
		entry.MeanNeighborJaccard = Summarize(jaccard, false).Mean
		entry.MeanTop1Recovery = Summarize(recovery, false).Mean
		tokenResult = append(tokenResult, entry)
	}
	pairGroups := make(map[string][]PairStability)
	for _, item := range pairs {
		pairGroups[frequencyBin(item.MinCount)] = append(pairGroups[frequencyBin(item.MinCount)], item)
	}
	var pairResult []PairFrequencyBin
	for _, bin := range frequencyBins {
		items := pairGroups[bin]
		entry := PairFrequencyBin{Bin: bin, Pairs: len(items)}
		var stddev, jaccard []float64
		crossing := 0
		for _, item := range items {
			if item.Summary.Observations > 0 {
				stddev = append(stddev, item.Summary.Stddev)
			}
			jaccard = append(jaccard, item.MeanMemberNeighborJaccard)
			if len(item.Thresholds) > 0 && item.Thresholds[0].ThresholdCrossingCount > 0 {
				crossing++
			}
		}
		entry.MeanSimilarityStddev = Summarize(stddev, false).Mean
		entry.MeanMemberNeighborJaccard = Summarize(jaccard, false).Mean
		if len(items) > 0 {
			entry.ThresholdCrossingFraction = float64(crossing) / float64(len(items))
		}
		pairResult = append(pairResult, entry)
	}
	return tokenResult, pairResult
}
