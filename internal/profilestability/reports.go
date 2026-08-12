package profilestability

import (
	"sort"

	"zcore.dev/voinich/internal/normalization"
)

func buildReferenceReports(model normalization.Model, pairs map[pairKey]PairStability, bootstrap map[pairKey]BootstrapPair, neighbors map[string]NeighborStability, profiles map[string]Profile) ([]ReferenceClass, []ComponentDiagnostic) {
	var classes []ReferenceClass
	diagnosticSet := make(map[pairKey]bool)
	for _, class := range model.Classes {
		if class.Size < 2 {
			continue
		}
		entry := ReferenceClass{ClassID: class.ID}
		for _, member := range class.Members {
			entry.Members = append(entry.Members, member.Token)
			neighbor := neighbors[member.Token]
			entry.MemberNeighborStability = append(entry.MemberNeighborStability, ReferenceMemberNeighbor{Token: member.Token, MeanJaccard: neighbor.MeanJaccard, Top1RecoveryFraction: neighbor.Top1RecoveryFraction})
		}
		sort.Strings(entry.Members)
		for i, a := range entry.Members {
			for _, b := range entry.Members[i+1:] {
				key := makePair(a, b)
				diagnosticSet[key] = true
				pair := pairs[key]
				boot := bootstrap[key]
				reference := ReferencePair{TokenA: a, TokenB: b, FullSimilarity: pair.Full.Similarity, FoldMean: pair.Summary.Mean, FoldStddev: pair.Summary.Stddev, FoldMin: pair.Summary.Min, FoldMax: pair.Summary.Max, BootstrapMean: boot.Similarity.Mean, BootstrapCI95: [2]float64{boot.Similarity.Percentile025, boot.Similarity.Percentile975}, BootstrapProbabilityAbove070: boot.ProbabilityAboveThreshold}
				entry.Pairwise = append(entry.Pairwise, reference)
			}
		}
		if len(entry.Pairwise) > 0 {
			weakest, strongest := entry.Pairwise[0], entry.Pairwise[0]
			for _, pair := range entry.Pairwise[1:] {
				if pair.FullSimilarity < weakest.FullSimilarity {
					weakest = pair
				}
				if pair.FullSimilarity > strongest.FullSimilarity {
					strongest = pair
				}
			}
			entry.WeakestPair = &weakest
			entry.StrongestPair = &strongest
		}
		classes = append(classes, entry)
	}
	var diagnostics []ComponentDiagnostic
	for _, key := range sortedPairs(diagnosticSet) {
		component := Compare(profiles[key.a], profiles[key.b])
		withoutPosition := (component.LeftSimilarity + component.RightSimilarity) / 2
		withoutLeft := (component.PositionSimilarity + component.RightSimilarity) / 2
		withoutRight := (component.PositionSimilarity + component.LeftSimilarity) / 2
		diagnostics = append(diagnostics, ComponentDiagnostic{TokenA: key.a, TokenB: key.b, FullSimilarity: component.Similarity, SimilarityWithoutPosition: withoutPosition, SimilarityWithoutLeft: withoutLeft, SimilarityWithoutRight: withoutRight, DeltaWithoutPosition: component.Similarity - withoutPosition, DeltaWithoutLeft: component.Similarity - withoutLeft, DeltaWithoutRight: component.Similarity - withoutRight})
	}
	return classes, diagnostics
}

func buildSummary(tokens []TokenProfileStability, neighbors []NeighborStability, pairs []PairStability, bootstrap []BootstrapPair, model normalization.Model, folds []foldData) Summary {
	result := Summary{EligibleTokens: len(tokens), CandidatePairs: len(pairs)}
	var position, left, right, top1, top5, jaccard, pairStddev []float64
	for _, item := range tokens {
		if item.PositionStability.Observations > 0 {
			position = append(position, item.PositionStability.Mean)
		}
		if item.LeftContextStability.Observations > 0 {
			left = append(left, item.LeftContextStability.Mean)
		}
		if item.RightContextStability.Observations > 0 {
			right = append(right, item.RightContextStability.Mean)
		}
	}
	for _, item := range neighbors {
		if item.FoldsWhereTop1Eligible > 0 {
			top1 = append(top1, item.Top1RecoveryFraction)
		}
		if item.FoldPairComparisons > 0 {
			top5 = append(top5, item.Top5OverlapMean)
			jaccard = append(jaccard, item.MeanJaccard)
		}
	}
	for _, item := range pairs {
		if item.Summary.Observations > 0 {
			pairStddev = append(pairStddev, item.Summary.Stddev)
		}
		if item.NearThresholdFraction > 0 {
			result.Pairs.PairsNearThreshold++
		}
	}
	for _, item := range bootstrap {
		if item.ProbabilityAboveThreshold >= .95 {
			result.Pairs.PairsBootstrapProbabilityAbove095++
		}
	}
	result.SelfProfile.MeanPositionStability = Summarize(position, false).Mean
	result.SelfProfile.MeanLeftStability = Summarize(left, false).Mean
	result.SelfProfile.MeanRightStability = Summarize(right, false).Mean
	result.NearestNeighbors.MeanTop1Recovery = Summarize(top1, false).Mean
	result.NearestNeighbors.MeanTop5Overlap = Summarize(top5, false).Mean
	result.NearestNeighbors.MeanTop10Jaccard = Summarize(jaccard, false).Mean
	result.Pairs.MeanSimilarityStddev = Summarize(pairStddev, false).Mean
	var fractions []float64
	for _, class := range model.Classes {
		if class.Size < 2 {
			continue
		}
		for i := 0; i < len(class.Members); i++ {
			for j := i + 1; j < len(class.Members); j++ {
				a, b := class.Members[i].Token, class.Members[j].Token
				eligible, same := 0, 0
				for _, fold := range folds {
					if !fold.trainEligible[a] || !fold.trainEligible[b] {
						continue
					}
					eligible++
					if sameClass(fold.model, a, b) {
						same++
					}
				}
				if eligible == 0 {
					continue
				}
				fraction := float64(same) / float64(eligible)
				fractions = append(fractions, fraction)
				if fraction == 1 {
					result.HardClassStability.PairsAlwaysSameClass++
				}
				if fraction == 0 {
					result.HardClassStability.PairsNeverSameClass++
				}
			}
		}
	}
	result.HardClassStability.ReferencePairs = len(fractions)
	result.HardClassStability.MeanSameClassFraction = Summarize(fractions, false).Mean
	return result
}

func sameClass(model normalization.Model, a, b string) bool {
	for _, class := range model.Classes {
		if class.Size < 2 {
			continue
		}
		foundA, foundB := false, false
		for _, member := range class.Members {
			if member.Token == a {
				foundA = true
			}
			if member.Token == b {
				foundB = true
			}
		}
		if foundA && foundB {
			return true
		}
	}
	return false
}
