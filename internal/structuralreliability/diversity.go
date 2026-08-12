package structuralreliability

import "zcore.dev/voinich/internal/profilestability"

// buildContextDiversity computes the full-corpus context diversity of
// section 23 for each token: how many distinct predecessors/successors it
// has been observed with, and how concentrated (low entropy) or spread out
// (high entropy) those distributions are. Section 24's effective-
// observations diagnostic (count divided by unique-neighbor count) is
// derived from the same numbers.
func buildContextDiversity(full map[string]profilestability.Profile, tokens []string) []TokenDiversity {
	result := make([]TokenDiversity, 0, len(tokens))
	for _, token := range tokens {
		profile := full[token]
		uniqueLeft, uniqueRight := len(profile.Left), len(profile.Right)
		result = append(result, TokenDiversity{
			Token: token, FullCount: profile.Count,
			UniquePredecessors: uniqueLeft, UniqueSuccessors: uniqueRight,
			LeftEntropy: Entropy(profile.Left), RightEntropy: Entropy(profile.Right),
			EffectiveLeftObservations:  float64(profile.Count) / float64(maxInt(1, uniqueLeft)),
			EffectiveRightObservations: float64(profile.Count) / float64(maxInt(1, uniqueRight)),
		})
	}
	return result
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
