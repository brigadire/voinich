package structuralreliability

import "zcore.dev/voinich/internal/profilestability"

// tokenMetric wraps the public ContinuousTokenMetric with the top3/top5
// overlap means the cumulative-threshold and frequency-bin aggregates need
// (section 6) but which section 9 does not expose per token.
type tokenMetric struct {
	ContinuousTokenMetric
	Top3OverlapMean float64
	Top5OverlapMean float64
}

func toContinuousMetrics(items []tokenMetric) []ContinuousTokenMetric {
	result := make([]ContinuousTokenMetric, len(items))
	for i, item := range items {
		result[i] = item.ContinuousTokenMetric
	}
	return result
}

// buildTokenMetrics is the single reusable engine behind sections 5, 6 and 9:
// given a minimum-count threshold it recomputes eligibility and, crucially,
// nearest-neighbor geometry restricted to that threshold's own eligible set
// (task section 6) exactly as if structural-profile-stability had been
// invoked with -min-token-count=minCount. Every similarity number comes from
// profilestability.Compare/BuildProfiles/NearestNeighbors; nothing here
// reimplements a similarity formula.
func buildTokenMetrics(full map[string]profilestability.Profile, folds []foldProfiles, minCount, neighborsK int) ([]tokenMetric, []string, map[string][]profilestability.Neighbor) {
	fullEligible := profilestability.Eligible(full, minCount)
	fullNeighbors := make(map[string][]profilestability.Neighbor, len(fullEligible))
	for _, token := range fullEligible {
		fullNeighbors[token] = profilestability.NearestNeighbors(full, token, fullEligible, neighborsK)
	}

	trainEligible := make([]map[string]bool, len(folds))
	testEligible := make([]map[string]bool, len(folds))
	trainNeighbors := make([]map[string][]profilestability.Neighbor, len(folds))
	for f, fold := range folds {
		trainList := profilestability.Eligible(fold.trainProfiles, minCount)
		trainEligible[f] = toSet(trainList)
		testEligible[f] = toSet(profilestability.Eligible(fold.testProfiles, minCount))
		neighbors := make(map[string][]profilestability.Neighbor, len(trainList))
		for _, token := range trainList {
			neighbors[token] = profilestability.NearestNeighbors(fold.trainProfiles, token, trainList, neighborsK)
		}
		trainNeighbors[f] = neighbors
	}

	metrics := make([]tokenMetric, 0, len(fullEligible))
	for _, token := range fullEligible {
		var position, left, right, ttPosition, ttLeft, ttRight []float64
		for f := range folds {
			if trainEligible[f][token] && testEligible[f][token] {
				value := profilestability.Compare(folds[f].trainProfiles[token], folds[f].testProfiles[token])
				ttPosition = append(ttPosition, value.PositionSimilarity)
				ttLeft = append(ttLeft, value.LeftSimilarity)
				ttRight = append(ttRight, value.RightSimilarity)
			}
			for g := f + 1; g < len(folds); g++ {
				if !trainEligible[f][token] || !trainEligible[g][token] {
					continue
				}
				value := profilestability.Compare(folds[f].trainProfiles[token], folds[g].trainProfiles[token])
				position = append(position, value.PositionSimilarity)
				left = append(left, value.LeftSimilarity)
				right = append(right, value.RightSimilarity)
			}
		}

		var jaccards, top3, top5, top10 []float64
		var top1Candidate string
		if len(fullNeighbors[token]) > 0 {
			top1Candidate = fullNeighbors[token][0].Token
		}
		foldsWhereTop1Eligible, foldsWhereSameTop1 := 0, 0
		for f := range folds {
			if !trainEligible[f][token] {
				continue
			}
			if top1Candidate != "" && trainEligible[f][top1Candidate] {
				foldsWhereTop1Eligible++
				if items := trainNeighbors[f][token]; len(items) > 0 && items[0].Token == top1Candidate {
					foldsWhereSameTop1++
				}
			}
			for g := f + 1; g < len(folds); g++ {
				if !trainEligible[g][token] {
					continue
				}
				left, right := trainNeighbors[f][token], trainNeighbors[g][token]
				jaccards = append(jaccards, profilestability.Jaccard(left, right))
				top3 = append(top3, profilestability.OverlapAt(left, right, 3))
				top5 = append(top5, profilestability.OverlapAt(left, right, 5))
				top10 = append(top10, profilestability.OverlapAt(left, right, neighborsK))
			}
		}

		item := tokenMetric{ContinuousTokenMetric: ContinuousTokenMetric{Token: token, FullCount: full[token].Count}}
		item.PositionTrainTrainStability = profilestability.Summarize(position, false).Mean
		item.LeftTrainTrainStability = profilestability.Summarize(left, false).Mean
		item.RightTrainTrainStability = profilestability.Summarize(right, false).Mean
		item.TrainTrainObservations = len(position)
		item.PositionTrainTestStability = profilestability.Summarize(ttPosition, false).Mean
		item.LeftTrainTestStability = profilestability.Summarize(ttLeft, false).Mean
		item.RightTrainTestStability = profilestability.Summarize(ttRight, false).Mean
		item.TrainTestObservations = len(ttPosition)
		item.Top1Recovery = ratio(foldsWhereSameTop1, foldsWhereTop1Eligible)
		item.Top1RecoveryObservations = foldsWhereTop1Eligible
		item.Top10Jaccard = profilestability.Summarize(jaccards, false).Mean
		item.Top10JaccardObservations = len(jaccards)
		item.Top3OverlapMean = profilestability.Summarize(top3, false).Mean
		item.Top5OverlapMean = profilestability.Summarize(top5, false).Mean
		metrics = append(metrics, item)
	}
	return metrics, fullEligible, fullNeighbors
}

// aggregateTokenMetrics reduces a group of tokens (a cumulative threshold or
// a frequency bin) to the mean/median/stddev-of-per-token-means aggregates
// used throughout sections 5, 6 and 7, following the same convention
// structural-profile-stability's own FrequencyDependence report uses.
func aggregateTokenMetrics(items []tokenMetric) (SelfProfileStability, TrainTestStability, NearestNeighborStability) {
	var position, left, right, ttPosition, ttLeft, ttRight []float64
	var top1, top3, top5, top10 []float64
	for _, item := range items {
		if item.TrainTrainObservations > 0 {
			position = append(position, item.PositionTrainTrainStability)
			left = append(left, item.LeftTrainTrainStability)
			right = append(right, item.RightTrainTrainStability)
		}
		if item.TrainTestObservations > 0 {
			ttPosition = append(ttPosition, item.PositionTrainTestStability)
			ttLeft = append(ttLeft, item.LeftTrainTestStability)
			ttRight = append(ttRight, item.RightTrainTestStability)
		}
		if item.Top1RecoveryObservations > 0 {
			top1 = append(top1, item.Top1Recovery)
		}
		if item.Top10JaccardObservations > 0 {
			top3 = append(top3, item.Top3OverlapMean)
			top5 = append(top5, item.Top5OverlapMean)
			top10 = append(top10, item.Top10Jaccard)
		}
	}
	self := SelfProfileStability{Position: SummarizeStat(position), LeftContext: SummarizeStat(left), RightContext: SummarizeStat(right)}
	trainTest := TrainTestStability{Position: SummarizeStat(ttPosition), LeftContext: SummarizeStat(ttLeft), RightContext: SummarizeStat(ttRight)}
	top1Stat := SummarizeStat(top1)
	neighbors := NearestNeighborStability{
		MeanTop1Recovery: top1Stat.Mean, MedianTop1Recovery: top1Stat.Median,
		MeanTop3Overlap: profilestability.Summarize(top3, false).Mean, MeanTop5Overlap: profilestability.Summarize(top5, false).Mean,
		MeanTop10Jaccard: profilestability.Summarize(top10, false).Mean,
	}
	return self, trainTest, neighbors
}

func toSet(tokens []string) map[string]bool {
	result := make(map[string]bool, len(tokens))
	for _, token := range tokens {
		result[token] = true
	}
	return result
}

func ratio(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}
