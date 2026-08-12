package structuralreliability

import (
	"sort"

	"zcore.dev/voinich/internal/normalization"
)

// referenceClassPairs collects every member pair from every multi-member
// class of the reference model (threshold 0.70), plus the two named pairs
// task section 25 calls out explicitly even if they happened not to share a
// class (they do not need to: they are already forced into the master
// candidate set by specialReferencePairs).
func referenceClassPairs(model normalization.Model) []pairKey {
	set := make(map[pairKey]bool)
	for _, class := range model.Classes {
		if class.Size < 2 {
			continue
		}
		members := make([]string, 0, len(class.Members))
		for _, member := range class.Members {
			members = append(members, member.Token)
		}
		sort.Strings(members)
		for i, a := range members {
			for _, b := range members[i+1:] {
				set[makePair(a, b)] = true
			}
		}
	}
	for _, pair := range specialReferencePairs {
		set[makePair(pair[0], pair[1])] = true
	}
	return sortedPairs(set)
}

// buildReferencePairs joins the reference-class pair list against the
// already-computed continuous_pair_metrics and the raw bootstrap
// distributions, so section 25's report reuses every number (bootstrap CI,
// pair reliability, fold stddev) rather than recomputing anything.
func buildReferencePairs(pairs []pairKey, byPair map[pairKey]ContinuousPairMetric, bootstrap map[pairKey]bootstrapResult) []ReferencePairReliability {
	result := make([]ReferencePairReliability, 0, len(pairs))
	for _, pair := range pairs {
		metric, ok := byPair[pair]
		if !ok {
			continue
		}
		entry := ReferencePairReliability{
			TokenA: metric.TokenA, TokenB: metric.TokenB, CountA: metric.CountA, CountB: metric.CountB,
			FullSimilarity: metric.FullSimilarity, FullPositionSimilarity: metric.FullPositionSimilarity,
			FullLeftSimilarity: metric.FullLeftSimilarity, FullRightSimilarity: metric.FullRightSimilarity,
			FoldSimilarityStddev:    metric.FoldSimilarityStddev,
			PositionReliabilityPair: metric.PositionReliabilityPair, LeftReliabilityPair: metric.LeftReliabilityPair, RightReliabilityPair: metric.RightReliabilityPair,
		}
		if boot, ok := bootstrap[pair]; ok && boot.Observations > 0 {
			mean, probability := boot.Mean, boot.ProbabilityAboveThreshold
			ci := [2]float64{boot.Percentile025, boot.Percentile975}
			entry.BootstrapMean, entry.BootstrapProbabilityAbove070, entry.BootstrapCI95 = &mean, &probability, &ci
		}
		result = append(result, entry)
	}
	return result
}
