package inversehomophony

import (
	"math"
	"sort"
)

// ClassRecoveryMetrics is task57 section 10's permutation-invariant
// clustering comparison between a predicted Partition and the (evaluator
// -only) oracle Partition. Computed entirely from the contingency table -
// never by comparing class-ID strings directly, which is meaningless
// across two independently-labeled partitions.
type ClassRecoveryMetrics struct {
	PairwisePrecision float64
	PairwiseRecall    float64
	PairwiseF1        float64
	ARI               float64
	NMI               float64
	PredictedClasses  int
	OracleClasses     int
}

// EvaluateClassRecovery compares predicted against oracle over the same
// token set (both Partitions must cover the same domain, e.g. both in
// relabeled ID space).
func EvaluateClassRecovery(predicted, oracle Partition) ClassRecoveryMetrics {
	contingency := make(map[string]map[string]int)
	rowTotal := make(map[string]int)
	colTotal := make(map[string]int)
	n := 0
	for t, pc := range predicted {
		oc, ok := oracle[t]
		if !ok {
			continue
		}
		if contingency[pc] == nil {
			contingency[pc] = make(map[string]int)
		}
		contingency[pc][oc]++
		rowTotal[pc]++
		colTotal[oc]++
		n++
	}
	if n == 0 {
		return ClassRecoveryMetrics{}
	}

	// Every accumulation below sums float64 values (choose2, log2 terms)
	// derived from map contents. float64 addition is not associative, so
	// every such sum is done over sorted keys, never map iteration order -
	// see the project's "Go map iteration determinism" convention; this
	// matters here because gate decisions later compare these values with
	// > and !=.
	predClasses := sortedKeys(rowTotal)
	oracleClasses := sortedKeys(colTotal)

	var sumNijC2, sumAiC2, sumBjC2 float64
	for _, pc := range predClasses {
		row := contingency[pc]
		ocs := sortedKeys(row)
		for _, oc := range ocs {
			sumNijC2 += choose2(row[oc])
		}
	}
	for _, pc := range predClasses {
		sumAiC2 += choose2(rowTotal[pc])
	}
	for _, oc := range oracleClasses {
		sumBjC2 += choose2(colTotal[oc])
	}
	nC2 := choose2(n)

	precision := safeDiv(sumNijC2, sumAiC2)
	recall := safeDiv(sumNijC2, sumBjC2)
	f1 := 0.0
	if precision+recall > 0 {
		f1 = 2 * precision * recall / (precision + recall)
	}

	expectedIndex := safeDiv(sumAiC2*sumBjC2, nC2)
	maxIndex := 0.5 * (sumAiC2 + sumBjC2)
	ari := 0.0
	if maxIndex != expectedIndex {
		ari = (sumNijC2 - expectedIndex) / (maxIndex - expectedIndex)
	} else if sumNijC2 == maxIndex {
		ari = 1
	}

	var mi, hPred, hOracle float64
	nf := float64(n)
	for _, pc := range predClasses {
		row := contingency[pc]
		ai := float64(rowTotal[pc])
		for _, oc := range sortedKeys(row) {
			nij := row[oc]
			if nij == 0 {
				continue
			}
			bj := float64(colTotal[oc])
			pij := float64(nij) / nf
			mi += pij * math.Log2(pij/((ai/nf)*(bj/nf)))
		}
	}
	for _, pc := range predClasses {
		p := float64(rowTotal[pc]) / nf
		if p > 0 {
			hPred -= p * math.Log2(p)
		}
	}
	for _, oc := range oracleClasses {
		p := float64(colTotal[oc]) / nf
		if p > 0 {
			hOracle -= p * math.Log2(p)
		}
	}
	nmi := 0.0
	if hPred+hOracle > 0 {
		nmi = 2 * mi / (hPred + hOracle)
	} else {
		nmi = 1 // both partitions are a single class each - perfectly (trivially) agreeing
	}

	return ClassRecoveryMetrics{
		PairwisePrecision: precision,
		PairwiseRecall:    recall,
		PairwiseF1:        f1,
		ARI:               ari,
		NMI:               nmi,
		PredictedClasses:  len(rowTotal),
		OracleClasses:     len(colTotal),
	}
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func choose2(n int) float64 {
	if n < 2 {
		return 0
	}
	f := float64(n)
	return f * (f - 1) / 2
}

func safeDiv(num, den float64) float64 {
	if den == 0 {
		return 0
	}
	return num / den
}
