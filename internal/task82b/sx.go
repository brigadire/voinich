package task82b

import (
	"fmt"
	"sort"
	"strings"
)

// SXMetric is one row of SX_RESULTS.tsv (task82b.txt sec.51/52): a
// diagnostic of the shorthand *process* itself, computed from the
// abbreviated/expanded alignment, not from Fingerprint V2. SX exists only
// because Task82b's own audit (sec.51) found F2 has no access to the
// abbreviated<->expanded alignment at all (F2 sees one corpus at a time),
// so none of these seven properties are measurable by F2 regardless of
// corpus size -- a structural gap, not a sensitivity shortfall.
type SXMetric struct {
	ID    string  `json:"id"`
	Value float64 `json:"value"`
	Note  string  `json:"note"`
}

// ComputeSX computes the seven minimum SX metrics (task82b.txt sec.52)
// from one witness's PairUnit list.
func ComputeSX(pairs []PairUnit) []SXMetric {
	type key struct{ a, e string }
	abbrTypes := map[string]bool{}
	expanTypes := map[string]bool{}
	pairCounts := map[key]int{}
	expansOfAbbr := map[string]map[string]bool{}
	abbrsOfExpan := map[string]map[string]bool{}

	totalDeleted, totalExpanLen := 0, 0
	initDeleted, initLen, initN := 0, 0, 0
	nonInitDeleted, nonInitLen, nonInitN := 0, 0, 0
	firstOrderOnLine := map[int]int{}
	for _, p := range pairs {
		if cur, ok := firstOrderOnLine[p.Line]; !ok || p.Order < cur {
			firstOrderOnLine[p.Line] = p.Order
		}
	}

	for _, p := range pairs {
		expan := strings.ToLower(strings.TrimSpace(p.ExpanText))
		if expan == "" {
			continue
		}
		abbrTypes[p.AbbrText] = true
		expanTypes[expan] = true
		pairCounts[key{p.AbbrText, expan}]++
		if expansOfAbbr[p.AbbrText] == nil {
			expansOfAbbr[p.AbbrText] = map[string]bool{}
		}
		expansOfAbbr[p.AbbrText][expan] = true
		if abbrsOfExpan[expan] == nil {
			abbrsOfExpan[expan] = map[string]bool{}
		}
		abbrsOfExpan[expan][p.AbbrText] = true

		d := DeletionCount(p)
		totalDeleted += d
		totalExpanLen += len([]rune(expan))
		if p.Order == firstOrderOnLine[p.Line] {
			initDeleted += d
			initLen += len([]rune(expan))
			initN++
		} else {
			nonInitDeleted += d
			nonInitLen += len([]rune(expan))
			nonInitN++
		}
	}

	sx1 := safeRatio(float64(totalDeleted), float64(totalExpanLen))

	nAmbiguousTypes, sumExpansionsPerAbbr := 0, 0
	for _, exps := range expansOfAbbr {
		sumExpansionsPerAbbr += len(exps)
		if len(exps) > 1 {
			nAmbiguousTypes++
		}
	}
	sx2 := safeRatio(float64(sumExpansionsPerAbbr), float64(len(expansOfAbbr)))
	sx5 := safeRatio(float64(nAmbiguousTypes), float64(len(expansOfAbbr)))

	counts := make([]float64, 0, len(pairCounts))
	for _, c := range pairCounts {
		counts = append(counts, float64(c))
	}
	sx3 := giniCoefficient(counts)

	initRate := safeRatio(float64(initDeleted), float64(initLen))
	nonInitRate := safeRatio(float64(nonInitDeleted), float64(nonInitLen))
	sx4 := initRate - nonInitRate

	sx6 := safeRatio(float64(len(abbrTypes)), float64(len(expanTypes)))

	edges := len(pairCounts)
	possible := float64(len(abbrTypes)) * float64(len(expanTypes))
	sx7 := safeRatio(float64(edges), possible)

	return []SXMetric{
		{ID: "SX1_CONTRACTION_RATE", Value: sx1, Note: "mean fraction of expanded letters removed by the real abbreviation, alignment heuristic in shorthandnull.go"},
		{ID: "SX2_EXPANSION_AMBIGUITY", Value: sx2, Note: "mean number of distinct expansions per distinct abbreviated surface form"},
		{ID: "SX3_ABBREVIATION_FAMILY_REUSE", Value: sx3, Note: "Gini coefficient of (abbr,expan) pair occurrence counts; higher = a few forms dominate reuse"},
		{ID: "SX4_POSITIONAL_ABBREVIATION_PREFERENCE", Value: sx4, Note: sxPositionalNote(initN, nonInitN)},
		{ID: "SX5_CONTEXT_DEPENDENCE", Value: sx5, Note: "fraction of abbreviated-form types with >=2 observed distinct expansions"},
		{ID: "SX6_MANY_TO_ONE_MAPPING", Value: sx6, Note: "distinct abbreviated-form types / distinct expansion types"},
		{ID: "SX7_ABBREVIATION_GRAPH_DENSITY", Value: sx7, Note: "bipartite (abbr type, expan type) edge density"},
	}
}

func sxPositionalNote(initN, nonInitN int) string {
	return fmt.Sprintf("line-initial minus non-initial mean contraction rate (n_initial=%d, n_non_initial=%d)", initN, nonInitN)
}

func safeRatio(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}

// giniCoefficient computes the Gini coefficient of a non-negative sample
// (mean absolute difference / (2*mean*n)), used here purely descriptively
// (reuse concentration), not as a fairness/inequality claim.
func giniCoefficient(xs []float64) float64 {
	n := len(xs)
	if n == 0 {
		return 0
	}
	sorted := append([]float64{}, xs...)
	sort.Float64s(sorted)
	var sumAbsDiff, sum float64
	for i, xi := range sorted {
		sum += xi
		for _, xj := range sorted[i+1:] {
			sumAbsDiff += xj - xi
		}
	}
	if sum == 0 {
		return 0
	}
	// sumAbsDiff sums (xj-xi) over unordered pairs i<j (sorted ascending),
	// i.e. half of sum_i sum_j |xi-xj|; the standard Gini denominator is
	// n*sum(x), not 2*n*sum(x), so no extra factor of 2 is applied here.
	return sumAbsDiff / (float64(n) * sum)
}
