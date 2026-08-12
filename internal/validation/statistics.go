package validation

import (
	"math"
	"sort"

	"zcore.dev/voinich/internal/normalization"
)

func BuildSequenceComparison(raw, structural sequenceMetrics, random []sequenceMetrics, minN, maxN, maxContext int) SequenceComparison {
	result := SequenceComparison{
		RawMaxCrossLineLength:        raw.MaxLength,
		StructuralMaxCrossLineLength: structural.MaxLength,
	}
	maxValues := make([]float64, len(random))
	for i := range random {
		maxValues[i] = float64(random[i].MaxLength)
	}
	result.MaxLengthRandom = compareRandom(float64(structural.MaxLength), maxValues, true)
	for n := minN; n <= maxN; n++ {
		rawValue, structuralValue := raw.CrossLine[n], structural.CrossLine[n]
		var ratio *float64
		if rawValue > 0 {
			value := float64(structuralValue) / float64(rawValue)
			ratio = &value
		}
		values := make([]float64, len(random))
		for i := range random {
			values[i] = float64(random[i].CrossLine[n])
		}
		result.NGrams = append(result.NGrams, NGramComparison{
			N: n, RawCrossLineRepeated: rawValue, StructuralCrossLineRepeated: structuralValue,
			AbsoluteDelta: structuralValue - rawValue, Ratio: ratio,
			Random: compareRandom(float64(structuralValue), values, true),
		})
	}
	for length := 1; length <= maxContext; length++ {
		rawMetric := raw.Contexts[length]
		structuralMetric := structural.Contexts[length]
		coverageDelta := structuralMetric.RepeatedContextCoverage - rawMetric.RepeatedContextCoverage
		entropyReduction := rawMetric.ConditionalEntropy - structuralMetric.ConditionalEntropy
		repeatedEntropyReduction := rawMetric.RepeatedContextConditionalEntropy - structuralMetric.RepeatedContextConditionalEntropy
		coverageValues := make([]float64, len(random))
		entropyValues := make([]float64, len(random))
		repeatedEntropyValues := make([]float64, len(random))
		for i := range random {
			coverageValues[i] = random[i].Contexts[length].RepeatedContextCoverage - rawMetric.RepeatedContextCoverage
			entropyValues[i] = rawMetric.ConditionalEntropy - random[i].Contexts[length].ConditionalEntropy
			repeatedEntropyValues[i] = rawMetric.RepeatedContextConditionalEntropy - random[i].Contexts[length].RepeatedContextConditionalEntropy
		}
		result.ContextOrder = append(result.ContextOrder, ContextComparison{
			ContextLength:                        length,
			RawRepeatedContextCoverage:           rawMetric.RepeatedContextCoverage,
			StructuralRepeatedContextCoverage:    structuralMetric.RepeatedContextCoverage,
			CoverageDelta:                        coverageDelta,
			RawConditionalEntropy:                rawMetric.ConditionalEntropy,
			StructuralConditionalEntropy:         structuralMetric.ConditionalEntropy,
			EntropyDelta:                         entropyReduction,
			RawRepeatedConditionalEntropy:        rawMetric.RepeatedContextConditionalEntropy,
			StructuralRepeatedConditionalEntropy: structuralMetric.RepeatedContextConditionalEntropy,
			RepeatedEntropyDelta:                 repeatedEntropyReduction,
			CoverageRandom:                       compareRandom(coverageDelta, coverageValues, true),
			EntropyReductionRandom:               compareRandom(entropyReduction, entropyValues, true),
			RepeatedEntropyReductionRandom:       compareRandom(repeatedEntropyReduction, repeatedEntropyValues, true),
		})
	}
	return result
}

func compareRandom(structural float64, values []float64, upperTail bool) RandomComparison {
	direction := "random >= structural"
	if !upperTail {
		direction = "random <= structural"
	}
	return RandomComparison{
		StructuralValue: structural, Direction: direction,
		Random: Summarize(values), EmpiricalP: EmpiricalP(values, structural, upperTail),
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
	result.Percentile05 = percentile(sorted, .05)
	result.Percentile50 = percentile(sorted, .50)
	result.Percentile95 = percentile(sorted, .95)
	return result
}

func EmpiricalP(values []float64, structural float64, upperTail bool) float64 {
	extreme := 0
	for _, value := range values {
		if (upperTail && value >= structural) || (!upperTail && value <= structural) {
			extreme++
		}
	}
	return float64(extreme+1) / float64(len(values)+1)
}

func percentile(sorted []float64, probability float64) float64 {
	if len(sorted) == 1 {
		return sorted[0]
	}
	position := probability * float64(len(sorted)-1)
	lower, upper := int(math.Floor(position)), int(math.Ceil(position))
	if lower == upper {
		return sorted[lower]
	}
	weight := position - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

func AggregateFolds(folds []FoldResult, minN, maxN int) CrossValidationAggregate {
	result := CrossValidationAggregate{}
	for n := minN; n <= maxN; n++ {
		item := AggregateNGram{N: n}
		pooled := PooledNGram{N: n}
		var deltas []float64
		for _, fold := range folds {
			metric := fold.SequenceComparison.NGrams[n-minN]
			item.MeanRaw += float64(metric.RawCrossLineRepeated)
			item.MeanStructural += float64(metric.StructuralCrossLineRepeated)
			delta := metric.AbsoluteDelta
			deltas = append(deltas, float64(delta))
			switch {
			case delta > 0:
				item.FoldsPositive++
			case delta < 0:
				item.FoldsNegative++
			default:
				item.FoldsZero++
			}
			pooled.RawCrossLineRepeated += metric.RawCrossLineRepeated
			pooled.StructuralCrossLineRepeated += metric.StructuralCrossLineRepeated
		}
		if len(folds) > 0 {
			item.MeanRaw /= float64(len(folds))
			item.MeanStructural /= float64(len(folds))
			item.MeanDelta = item.MeanStructural - item.MeanRaw
			sort.Float64s(deltas)
			item.MedianDelta = percentile(deltas, .5)
		}
		pooled.AbsoluteDelta = pooled.StructuralCrossLineRepeated - pooled.RawCrossLineRepeated
		result.CrossLineNGrams = append(result.CrossLineNGrams, item)
		result.PooledTest = append(result.PooledTest, pooled)
	}
	return result
}

type tokenPair struct{ left, right string }

func BuildClassStability(models []normalization.Model, eligible []map[string]bool) ClassStability {
	pairs := make(map[tokenPair]bool)
	for _, model := range models {
		for _, class := range model.Classes {
			if class.Size < 2 {
				continue
			}
			for i := 0; i < len(class.Members); i++ {
				for j := i + 1; j < len(class.Members); j++ {
					left, right := class.Members[i].Token, class.Members[j].Token
					if left > right {
						left, right = right, left
					}
					pairs[tokenPair{left, right}] = true
				}
			}
		}
	}
	result := ClassStability{ReportedPairRule: "union of token pairs assigned to the same multi-member class in at least one TRAIN fold; eligibility denominator includes every fold where both TRAIN counts meet min_token_count"}
	for pair := range pairs {
		entry := StabilityPair{TokenA: pair.left, TokenB: pair.right}
		for fold, model := range models {
			if !eligible[fold][pair.left] || !eligible[fold][pair.right] {
				continue
			}
			entry.FoldsBothEligible++
			if sameClass(model, pair.left, pair.right) {
				entry.FoldsSameClass++
			}
		}
		if entry.FoldsBothEligible > 0 {
			entry.Stability = float64(entry.FoldsSameClass) / float64(entry.FoldsBothEligible)
		}
		switch {
		case entry.Stability == 1:
			result.StablePairs100Percent++
		case entry.Stability >= .8:
			result.StablePairs80Percent++
		default:
			result.UnstablePairs++
		}
		result.Pairs = append(result.Pairs, entry)
	}
	sort.Slice(result.Pairs, func(i, j int) bool {
		if result.Pairs[i].TokenA != result.Pairs[j].TokenA {
			return result.Pairs[i].TokenA < result.Pairs[j].TokenA
		}
		return result.Pairs[i].TokenB < result.Pairs[j].TokenB
	})
	return result
}

func sameClass(model normalization.Model, left, right string) bool {
	for _, class := range model.Classes {
		if class.Size < 2 {
			continue
		}
		foundLeft, foundRight := false, false
		for _, member := range class.Members {
			if member.Token == left {
				foundLeft = true
			}
			if member.Token == right {
				foundRight = true
			}
		}
		if foundLeft && foundRight {
			return true
		}
	}
	return false
}
