package fingerprintv2

import (
	"math"
	"sort"
)

// pearson is a small local helper; the rest of the package standardizes on
// Spearman, but a redundancy scan over already-computed scalar summary
// values benefits from the simpler, well-understood linear correlation.
func pearson(x, y []float64) float64 {
	if len(x) < 3 || len(x) != len(y) {
		return 0
	}
	mx, my := mean(x), mean(y)
	num, dx, dy := 0.0, 0.0, 0.0
	for i := range x {
		a, b := x[i]-mx, y[i]-my
		num += a * b
		dx += a * a
		dy += b * b
	}
	if dx == 0 || dy == 0 {
		return 0
	}
	return num / math.Sqrt(dx*dy)
}

// summaryMetricSeries collects one scalar per grammar replicate (across
// every validated mode) for a declared set of LP/EF/CS-adjacent scalar
// summaries, giving a same-corpus, same-generative-process sample large
// enough for a redundancy correlation scan without needing multiple
// independent corpora.
func summaryMetricSeries(runs []GrammarRun) map[string][]float64 {
	series := map[string][]float64{
		"lp1_support_gini":             {},
		"lp4_prefix_nmi":               {},
		"lp4_suffix_nmi":               {},
		"ef1_giant_component_share":    {},
		"ef1_isolate_share":            {},
		"ef2_global_clustering":        {},
		"ef3_spearman_degree_log_freq": {},
	}
	for _, r := range runs {
		series["lp1_support_gini"] = append(series["lp1_support_gini"], r.LP1Gini)
		series["lp4_prefix_nmi"] = append(series["lp4_prefix_nmi"], r.PrefixNMI)
		series["lp4_suffix_nmi"] = append(series["lp4_suffix_nmi"], r.SuffixNMI)
		series["ef1_giant_component_share"] = append(series["ef1_giant_component_share"], r.EF1.GiantComponentShare)
		series["ef1_isolate_share"] = append(series["ef1_isolate_share"], r.EF1.IsolateShare)
		series["ef2_global_clustering"] = append(series["ef2_global_clustering"], r.EF2.GlobalClustering)
		series["ef3_spearman_degree_log_freq"] = append(series["ef3_spearman_degree_log_freq"], r.EF3.SpearmanDegreeLogFrequency)
	}
	return series
}

func redundancyAnalysis(runs []GrammarRun) ([]RedundancyRow, []MetricClassification) {
	series := summaryMetricSeries(runs)
	names := orderedKeys(series)
	var rows []RedundancyRow
	n := 0
	if len(names) > 0 {
		n = len(series[names[0]])
	}
	adjacency := map[string]map[string]float64{}
	for _, name := range names {
		adjacency[name] = map[string]float64{}
	}
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			r := pearson(series[names[i]], series[names[j]])
			rows = append(rows, RedundancyRow{MetricA: names[i], MetricB: names[j], Correlation: r, N: n})
			adjacency[names[i]][names[j]] = r
			adjacency[names[j]][names[i]] = r
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].MetricA != rows[j].MetricA {
			return rows[i].MetricA < rows[j].MetricA
		}
		return rows[i].MetricB < rows[j].MetricB
	})
	// Group metrics whose pairwise |r| >= 0.9 into a redundancy cluster and
	// keep one canonical (CORE) member per cluster by declared priority;
	// low-N corpora (n<2) cannot certify any correlation, so every metric
	// is reported SUPPORTING/DIAGNOSTIC by declared role instead.
	classifications := make([]MetricClassification, 0, len(names))
	assigned := map[string]bool{}
	priority := []string{"lp1_support_gini", "ef1_giant_component_share", "ef2_global_clustering", "ef3_spearman_degree_log_freq", "lp4_prefix_nmi", "lp4_suffix_nmi", "ef1_isolate_share"}
	for _, canonical := range priority {
		if assigned[canonical] {
			continue
		}
		if n < 4 {
			classifications = append(classifications, MetricClassification{MetricID: canonical, Class: "SUPPORTING", Reason: "fewer than 4 grammar replicates were available to certify redundancy; classified by declared analytical role instead"})
			assigned[canonical] = true
			continue
		}
		cluster := []string{canonical}
		for _, other := range names {
			if other == canonical || assigned[other] {
				continue
			}
			if math.Abs(adjacency[canonical][other]) >= 0.9 {
				cluster = append(cluster, other)
			}
		}
		classifications = append(classifications, MetricClassification{MetricID: canonical, Class: "CORE", Reason: "declared canonical summary for its family; retained regardless of correlation because it has the clearest standalone interpretation"})
		assigned[canonical] = true
		for _, other := range cluster[1:] {
			classifications = append(classifications, MetricClassification{MetricID: other, Class: "REDUNDANT", Reason: "|Pearson r| >= 0.9 with " + canonical + " across grammar replicates of this corpus; retained as a diagnostic, not dropped, since null behavior can still differ"})
			assigned[other] = true
		}
	}
	for _, name := range names {
		if !assigned[name] {
			classifications = append(classifications, MetricClassification{MetricID: name, Class: "DIAGNOSTIC", Reason: "not correlated >=0.9 with any canonical summary; carries independent information but is not itself a primary reported statistic"})
		}
	}
	sort.Slice(classifications, func(i, j int) bool { return classifications[i].MetricID < classifications[j].MetricID })
	return rows, classifications
}
