package inversehomophony

import (
	"math/rand/v2"
	"sort"
)

// PairDiscrimination is the task57 section 19 early-stop diagnostic: do the
// merge-evidence features actually separate true homophone pairs from
// random non-homophone pairs? oracle is evaluator-only and must never be
// passed into any recovery-path code.
type PairDiscrimination struct {
	TruePairs   int
	FalsePairs  int
	AUC         float64
	TrueScores  []float64
	FalseScores []float64
}

// DiscriminatePairs scores every true-homophone pair (same oracle class,
// distinct cipher token) that clears cfg.MinSupport, and an equal-sized
// sample of random non-homophone pairs drawn deterministically from seed,
// then reports the AUC of Score for separating the two (task57 section
// 19). AUC is computed by the standard rank-sum method (Mann-Whitney
// U / (n1*n2)), exact for these sample sizes.
func DiscriminatePairs(features map[string]*TokenFeatures, oracle Partition, cfg Config, seed int64) PairDiscrimination {
	classOf := oracle
	tokens := make([]string, 0, len(features))
	for t := range features {
		tokens = append(tokens, t)
	}
	sortStrings(tokens)

	byClass := make(map[string][]string)
	for _, t := range tokens {
		byClass[classOf[t]] = append(byClass[classOf[t]], t)
	}

	var truePairs []tokenPair
	trueSet := make(map[tokenPair]struct{})
	classNames := make([]string, 0, len(byClass))
	for c := range byClass {
		classNames = append(classNames, c)
	}
	sort.Strings(classNames)
	for _, c := range classNames {
		members := byClass[c]
		for i := 0; i < len(members); i++ {
			for j := i + 1; j < len(members); j++ {
				a, b := members[i], members[j]
				if a > b {
					a, b = b, a
				}
				pk := tokenPair{a, b}
				truePairs = append(truePairs, pk)
				trueSet[pk] = struct{}{}
			}
		}
	}

	var trueScores []float64
	for _, pk := range truePairs {
		fa, fb := features[pk.a], features[pk.b]
		score, _, _, _, _, support := CombinedScore(fa, fb)
		if support < cfg.MinSupport {
			continue
		}
		trueScores = append(trueScores, score)
	}

	// Matched-count random false-pair sample, deterministic given seed.
	r := rand.New(rand.NewPCG(uint64(seed), 0xD1A6))
	var falseScores []float64
	attempts, target := 0, len(trueScores)
	maxAttempts := target * 50
	if maxAttempts < 2000 {
		maxAttempts = 2000
	}
	for len(falseScores) < target && attempts < maxAttempts && len(tokens) >= 2 {
		attempts++
		a := tokens[r.IntN(len(tokens))]
		b := tokens[r.IntN(len(tokens))]
		if a == b {
			continue
		}
		if a > b {
			a, b = b, a
		}
		if _, ok := trueSet[tokenPair{a, b}]; ok {
			continue
		}
		fa, fb := features[a], features[b]
		score, _, _, _, _, support := CombinedScore(fa, fb)
		if support < cfg.MinSupport {
			continue
		}
		falseScores = append(falseScores, score)
	}

	return PairDiscrimination{
		TruePairs:   len(trueScores),
		FalsePairs:  len(falseScores),
		AUC:         auc(trueScores, falseScores),
		TrueScores:  trueScores,
		FalseScores: falseScores,
	}
}

// auc is the Mann-Whitney U statistic normalized to [0,1]: the probability
// that a random positive score exceeds a random negative score (ties count
// as one half).
func auc(pos, neg []float64) float64 {
	if len(pos) == 0 || len(neg) == 0 {
		return 0.5
	}
	sortedNeg := append([]float64(nil), neg...)
	sort.Float64s(sortedNeg)
	var wins float64
	for _, p := range pos {
		lo := sort.SearchFloat64s(sortedNeg, p)
		hi := sort.Search(len(sortedNeg), func(i int) bool { return sortedNeg[i] > p })
		less := lo
		equal := hi - lo
		wins += float64(less) + 0.5*float64(equal)
	}
	return wins / float64(len(pos)*len(neg))
}

// FreezeThreshold picks tau as the score maximizing Youden's J (TPR-FPR)
// over the development true/false score distributions (task57 section 6,
// "frozen on development corpora only"). Ties broken by the smaller
// (more conservative, fewer merges) candidate threshold.
func FreezeThreshold(d PairDiscrimination) float64 {
	candidates := append([]float64{}, d.TrueScores...)
	candidates = append(candidates, d.FalseScores...)
	sort.Float64s(candidates)
	bestJ := -1.0
	bestTau := 1.0
	for _, tau := range candidates {
		tp, fp := 0, 0
		for _, s := range d.TrueScores {
			if s > tau {
				tp++
			}
		}
		for _, s := range d.FalseScores {
			if s > tau {
				fp++
			}
		}
		tpr := ratio(tp, len(d.TrueScores))
		fpr := ratio(fp, len(d.FalseScores))
		j := tpr - fpr
		if j > bestJ || (j == bestJ && tau < bestTau) {
			bestJ = j
			bestTau = tau
		}
	}
	return bestTau
}

func ratio(n, d int) float64 {
	if d == 0 {
		return 0
	}
	return float64(n) / float64(d)
}
