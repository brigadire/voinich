package main

import (
	"math"
	"sort"
)

func buildOutput(dataset *Dataset, parameters Parameters) Output {
	output := Output{
		Meta:       dataset.Meta,
		Parameters: parameters,
		Methodology: Methodology{
			PositionalScore: "Jensen-Shannon divergence (base 2) between the token and corpus position distributions; range [0,1]",
			Predictability:  "1 - entropy/log2(unique observed neighbors); describes concentration among observed contexts, not proof of a structural restriction",
			Reliability:     "observations/(observations + reliability_prior_count), multiplied by available position coverage for position-based rankings; applied only to ranking scores",
			Asymmetry:       "(P(B|A)-P(A|B))/(P(B|A)+P(A|B)); range [-1,1]",
			PMI:             "log2(observed transitions / expected transitions under independent endpoints)",
			LogLikelihood:   "G-test statistic for the 2x2 transition contingency table",
			Equivalence:     "arithmetic mean of position, left-context and right-context similarities; context similarity is cosine, position similarity is 1-JSD",
		},
		PositionBaseline: makePositionBaseline(dataset.CorpusPositions),
	}

	output.PositionalSpecialization = positionalRanking(dataset, parameters)
	output.SuccessorPredictability = predictabilityRanking(dataset, parameters, true)
	output.PredecessorPredictability = predictabilityRanking(dataset, parameters, false)
	output.SignificantTransitions = transitionRanking(dataset, parameters)
	output.SelfTransitions = selfTransitionRanking(dataset, parameters)
	output.EquivalenceCandidates = equivalenceRanking(dataset, parameters)
	return output
}

func makePositionBaseline(counts map[int]int) []PositionBaseline {
	total := sumPositionCounts(counts)
	result := make([]PositionBaseline, 0, len(counts))
	for position, count := range counts {
		result = append(result, PositionBaseline{Position: position, Count: count, Probability: ratio(count, total)})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Position < result[j].Position })
	return result
}

func positionalRanking(dataset *Dataset, parameters Parameters) []PositionalResult {
	corpusTotal := sumPositionCounts(dataset.CorpusPositions)
	results := make([]PositionalResult, 0)
	for _, token := range dataset.Tokens {
		if token.Count < parameters.MinTokenCountForRanking {
			continue
		}
		positions := dataset.Positions[token.Token]
		positionTotal := sumPositionCounts(positions)
		positionCoverage := ratio(positionTotal, token.Count)
		score := positionJSD(positions, dataset.CorpusPositions)
		reliability := reliability(positionTotal, parameters.ReliabilityPriorCount) * positionCoverage
		result := PositionalResult{
			Token:                token.Token,
			Count:                token.Count,
			LineStartCount:       token.LineStartCount,
			LineEndCount:         token.LineEndCount,
			StartProbability:     ratio(token.LineStartCount, token.Count),
			EndProbability:       ratio(token.LineEndCount, token.Count),
			PositionObservations: positionTotal,
			PositionCoverage:     positionCoverage,
			Score:                score,
			Reliability:          reliability,
			RankingScore:         score * reliability,
		}
		for position, count := range positions {
			result.Positions = append(result.Positions, PositionValue{
				Position:          position,
				Count:             count,
				Probability:       ratio(count, positionTotal),
				CorpusProbability: ratio(dataset.CorpusPositions[position], corpusTotal),
			})
		}
		sort.Slice(result.Positions, func(i, j int) bool { return result.Positions[i].Position < result.Positions[j].Position })
		results = append(results, result)
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].RankingScore == results[j].RankingScore {
			return results[i].Token < results[j].Token
		}
		return results[i].RankingScore > results[j].RankingScore
	})
	return limitSlice(results, parameters.MaxItemsPerSection)
}

func predictabilityRanking(dataset *Dataset, parameters Parameters, successors bool) []ConstraintResult {
	results := make([]ConstraintResult, 0)
	for _, token := range dataset.Tokens {
		if token.Count < parameters.MinTokenCountForRanking {
			continue
		}
		contexts := dataset.Left[token.Token]
		if successors {
			contexts = dataset.Right[token.Token]
		}
		observedTransitions := sumStringCounts(contexts)
		if observedTransitions < parameters.MinContextObservations {
			continue
		}
		unique, entropyValue := entropy(contexts)
		maxEntropy := 0.0
		if unique > 1 {
			maxEntropy = math.Log2(float64(unique))
		}
		normalizedEntropy := 0.0
		if maxEntropy > 0 {
			normalizedEntropy = entropyValue / maxEntropy
		}
		predictability := 0.0
		if unique > 0 {
			predictability = 1 - normalizedEntropy
		}
		reliability := reliability(observedTransitions, parameters.ReliabilityPriorCount)
		result := ConstraintResult{
			Token:               token.Token,
			Count:               token.Count,
			ObservedTransitions: observedTransitions,
			Entropy:             entropyValue,
			MaxEntropy:          maxEntropy,
			NormalizedEntropy:   normalizedEntropy,
			Predictability:      predictability,
			Reliability:         reliability,
			RankingScore:        predictability * reliability,
		}
		if successors {
			result.UniqueSuccessors = unique
			result.DominantSuccessors = dominantNeighbors(contexts, parameters.DominantContextLimit)
		} else {
			result.UniquePredecessors = unique
			result.DominantPredecessors = dominantNeighbors(contexts, parameters.DominantContextLimit)
		}
		results = append(results, result)
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].RankingScore == results[j].RankingScore {
			return results[i].Token < results[j].Token
		}
		return results[i].RankingScore > results[j].RankingScore
	})
	return limitSlice(results, parameters.MaxItemsPerSection)
}

func dominantNeighbors(counts map[string]int, limit int) []DominantNeighbor {
	total := sumStringCounts(counts)
	result := make([]DominantNeighbor, 0, len(counts))
	for token, count := range counts {
		if count > 0 {
			result = append(result, DominantNeighbor{Token: token, Count: count, Probability: ratio(count, total)})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Count == result[j].Count {
			return result[i].Token < result[j].Token
		}
		return result[i].Count > result[j].Count
	})
	return limitSlice(result, limit)
}

func transitionRanking(dataset *Dataset, parameters Parameters) []SignificantTransition {
	results := make([]SignificantTransition, 0)
	total := dataset.Meta.Transitions
	for from, transitions := range dataset.Right {
		for to, count := range transitions {
			if count < parameters.MinTransitionCount {
				continue
			}
			fromTotal := dataset.Outgoing[from]
			toIncoming := dataset.Incoming[to]
			probability := ratio(count, fromTotal)
			reverseCount := dataset.Right[to][from]
			reverseProbability := ratio(reverseCount, dataset.Outgoing[to])
			asymmetry := 0.0
			if denominator := probability + reverseProbability; denominator > 0 {
				asymmetry = (probability - reverseProbability) / denominator
			}
			expected := float64(fromTotal) * float64(toIncoming) / float64(total)
			pmi := math.Log2(float64(count) / expected)
			results = append(results, SignificantTransition{
				From:               from,
				To:                 to,
				Count:              count,
				FromTransitions:    fromTotal,
				ToIncoming:         toIncoming,
				Probability:        probability,
				ReverseCount:       reverseCount,
				ReverseProbability: reverseProbability,
				Asymmetry:          asymmetry,
				Expected:           expected,
				PMI:                pmi,
				LogLikelihood:      logLikelihood2x2(count, fromTotal, toIncoming, total),
			})
		}
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].LogLikelihood == results[j].LogLikelihood {
			if results[i].From == results[j].From {
				return results[i].To < results[j].To
			}
			return results[i].From < results[j].From
		}
		return results[i].LogLikelihood > results[j].LogLikelihood
	})
	return limitSlice(results, parameters.MaxItemsPerSection)
}

func logLikelihood2x2(observed, rowTotal, columnTotal, total int) float64 {
	cells := [4]float64{
		float64(observed),
		float64(rowTotal - observed),
		float64(columnTotal - observed),
		float64(total - rowTotal - columnTotal + observed),
	}
	rows := [2]float64{float64(rowTotal), float64(total - rowTotal)}
	columns := [2]float64{float64(columnTotal), float64(total - columnTotal)}
	value := 0.0
	for row := 0; row < 2; row++ {
		for column := 0; column < 2; column++ {
			cell := cells[row*2+column]
			if cell == 0 {
				continue
			}
			expected := rows[row] * columns[column] / float64(total)
			value += cell * math.Log(cell/expected)
		}
	}
	return 2 * value
}

func selfTransitionRanking(dataset *Dataset, parameters Parameters) []SelfTransitionResult {
	results := make([]SelfTransitionResult, 0)
	for _, token := range dataset.Tokens {
		if token.Count < parameters.MinTokenCountForRanking {
			continue
		}
		count := dataset.Right[token.Token][token.Token]
		if count < parameters.MinSelfTransitionCount {
			continue
		}
		expected := float64(dataset.Outgoing[token.Token]) * float64(dataset.Incoming[token.Token]) / float64(dataset.Meta.Transitions)
		enrichment := 0.0
		if expected > 0 {
			enrichment = float64(count) / expected
		}
		results = append(results, SelfTransitionResult{
			Token:       token.Token,
			TokenCount:  token.Count,
			Count:       count,
			Outgoing:    dataset.Outgoing[token.Token],
			Incoming:    dataset.Incoming[token.Token],
			Probability: ratio(count, dataset.Outgoing[token.Token]),
			Expected:    expected,
			Enrichment:  enrichment,
		})
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Enrichment == results[j].Enrichment {
			return results[i].Token < results[j].Token
		}
		return results[i].Enrichment > results[j].Enrichment
	})
	return limitSlice(results, parameters.MaxItemsPerSection)
}

func equivalenceRanking(dataset *Dataset, parameters Parameters) []EquivalenceCandidate {
	eligible := make([]DictionaryToken, 0)
	for _, token := range dataset.Tokens {
		if token.Count >= parameters.MinTokenCountForRanking {
			eligible = append(eligible, token)
		}
	}
	results := make([]EquivalenceCandidate, 0)
	for i := 0; i < len(eligible); i++ {
		for j := i + 1; j < len(eligible); j++ {
			left := eligible[i]
			right := eligible[j]
			positionSimilarity := 1 - positionJSD(dataset.Positions[left.Token], dataset.Positions[right.Token])
			leftSimilarity := cosineSimilarity(dataset.Left[left.Token], dataset.Left[right.Token])
			rightSimilarity := cosineSimilarity(dataset.Right[left.Token], dataset.Right[right.Token])
			similarity := (positionSimilarity + leftSimilarity + rightSimilarity) / 3
			if similarity < parameters.MinEquivalenceSimilarity {
				continue
			}
			minimumCount := left.Count
			if right.Count < minimumCount {
				minimumCount = right.Count
			}
			leftCoverage := ratio(sumPositionCounts(dataset.Positions[left.Token]), left.Count)
			rightCoverage := ratio(sumPositionCounts(dataset.Positions[right.Token]), right.Count)
			positionCoverage := math.Sqrt(leftCoverage * rightCoverage)
			reliabilityValue := reliability(minimumCount, parameters.ReliabilityPriorCount) * positionCoverage
			results = append(results, EquivalenceCandidate{
				TokenA:                 left.Token,
				TokenB:                 right.Token,
				CountA:                 left.Count,
				CountB:                 right.Count,
				Similarity:             similarity,
				PositionSimilarity:     positionSimilarity,
				LeftContextSimilarity:  leftSimilarity,
				RightContextSimilarity: rightSimilarity,
				Reliability:            reliabilityValue,
				RankingScore:           similarity * reliabilityValue,
			})
		}
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].RankingScore == results[j].RankingScore {
			if results[i].TokenA == results[j].TokenA {
				return results[i].TokenB < results[j].TokenB
			}
			return results[i].TokenA < results[j].TokenA
		}
		return results[i].RankingScore > results[j].RankingScore
	})
	return limitSlice(results, parameters.MaxItemsPerSection)
}

func reliability(count int, prior float64) float64 {
	if count <= 0 {
		return 0
	}
	if prior <= 0 {
		return 1
	}
	return float64(count) / (float64(count) + prior)
}

func positionJSD(left, right map[int]int) float64 {
	leftTotal := sumPositionCounts(left)
	rightTotal := sumPositionCounts(right)
	if leftTotal == 0 || rightTotal == 0 {
		return 1
	}
	positions := make(map[int]struct{}, len(left)+len(right))
	for position := range left {
		positions[position] = struct{}{}
	}
	for position := range right {
		positions[position] = struct{}{}
	}
	value := 0.0
	for position := range positions {
		leftProbability := ratio(left[position], leftTotal)
		rightProbability := ratio(right[position], rightTotal)
		middle := (leftProbability + rightProbability) / 2
		if leftProbability > 0 {
			value += 0.5 * leftProbability * math.Log2(leftProbability/middle)
		}
		if rightProbability > 0 {
			value += 0.5 * rightProbability * math.Log2(rightProbability/middle)
		}
	}
	return clamp(value, 0, 1)
}

func cosineSimilarity(left, right map[string]int) float64 {
	if len(left) == 0 || len(right) == 0 {
		return 0
	}
	if len(left) > len(right) {
		left, right = right, left
	}
	dot := 0.0
	for token, count := range left {
		dot += float64(count * right[token])
	}
	leftNorm := 0.0
	for _, count := range left {
		leftNorm += float64(count * count)
	}
	rightNorm := 0.0
	for _, count := range right {
		rightNorm += float64(count * count)
	}
	if leftNorm == 0 || rightNorm == 0 {
		return 0
	}
	return clamp(dot/(math.Sqrt(leftNorm)*math.Sqrt(rightNorm)), 0, 1)
}

func clamp(value, minimum, maximum float64) float64 {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}

func limitSlice[T any](items []T, limit int) []T {
	if limit > 0 && len(items) > limit {
		return items[:limit]
	}
	return items
}
