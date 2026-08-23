package main

import (
	"fmt"
	"sort"
	"strings"

	"zcore.dev/voinich/internal/mechanismspace"
)

// WriteFamilyMetricsTSV writes FAMILY_METRICS.tsv (task66 section 37).
// "progress" is the family's robust (median) normalized progress across
// every DEVELOPMENT-stage metric registered for it (mechanismspace.
// FamilyScore); a family with more than one DEVELOPMENT metric (e.g.
// CHARACTER_ENTROPY's h1/h2) has no single (baseline, voynich, candidate)
// triple to report per row, so those per-metric values live in
// DEVELOPMENT_RESULTS.tsv/VOYNICH_TARGET_MANIFEST.tsv instead of being
// left blank here.
func WriteFamilyMetricsTSV(path string, rows []FamilyMetricsRow) error {
	var b strings.Builder
	b.WriteString("mechanism\tcorpus\tevaluation_set\tfamily\tprogress\toverall_status\n")
	for _, r := range rows {
		fams := make([]string, 0, len(r.FamilyScores))
		for f := range r.FamilyScores {
			fams = append(fams, f)
		}
		sort.Strings(fams)
		for _, f := range fams {
			b.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%.9g\t%s\n", r.Mechanism, r.Corpus, r.EvaluationSet, f, r.FamilyScores[f], r.OverallStatus))
		}
	}
	return writeFile(path, b.String())
}

// MeanFamilyScoresAcrossCorpora averages each mechanism's per-family
// score over every corpus it was evaluated on (task66 section 38's
// Pareto dominance is computed mechanism-vs-mechanism, not per corpus).
func MeanFamilyScoresAcrossCorpora(rows []FamilyMetricsRow) (mechanisms []string, scores []map[string]float64) {
	byMech := map[string][]map[string]float64{}
	for _, r := range rows {
		byMech[r.Mechanism] = append(byMech[r.Mechanism], r.FamilyScores)
	}
	names := make([]string, 0, len(byMech))
	for m := range byMech {
		names = append(names, m)
	}
	sort.Strings(names)
	for _, m := range names {
		sums := map[string]float64{}
		counts := map[string]int{}
		for _, fs := range byMech[m] {
			for f, v := range fs {
				sums[f] += v
				counts[f]++
			}
		}
		avg := map[string]float64{}
		for f, s := range sums {
			avg[f] = s / float64(counts[f])
		}
		mechanisms = append(mechanisms, m)
		scores = append(scores, avg)
	}
	return mechanisms, scores
}

// WriteParetoTSV writes MECHANISM_PARETO.tsv (task66 section 38).
func WriteParetoTSV(path string, mechanisms []string, scores []map[string]float64) ([]string, error) {
	front := mechanismspace.ParetoFront(scores)
	onFront := map[int]bool{}
	for _, i := range front {
		onFront[i] = true
	}
	var b strings.Builder
	b.WriteString("mechanism\ton_pareto_front\tmean_family_scores\n")
	var frontier []string
	for i, m := range mechanisms {
		fams := make([]string, 0, len(scores[i]))
		for f := range scores[i] {
			fams = append(fams, f)
		}
		sort.Strings(fams)
		var parts []string
		for _, f := range fams {
			parts = append(parts, fmt.Sprintf("%s=%.4g", f, scores[i][f]))
		}
		b.WriteString(fmt.Sprintf("%s\t%v\t%s\n", m, onFront[i], strings.Join(parts, ";")))
		if onFront[i] {
			frontier = append(frontier, m)
		}
	}
	return frontier, writeFile(path, b.String())
}

// WriteCorpusTransferTSV writes CORPUS_TRANSFER.tsv (task66 section 43):
// each frozen candidate's per-corpus family scores under the identical
// parameter set.
func WriteCorpusTransferTSV(path string, candidates []string, rows []FamilyMetricsRow) error {
	want := map[string]bool{}
	for _, c := range candidates {
		want[c] = true
	}
	var b strings.Builder
	b.WriteString("mechanism\tcorpus\tfamily\tprogress\n")
	for _, r := range rows {
		if !want[r.Mechanism] {
			continue
		}
		fams := make([]string, 0, len(r.FamilyScores))
		for f := range r.FamilyScores {
			fams = append(fams, f)
		}
		sort.Strings(fams)
		for _, f := range fams {
			b.WriteString(fmt.Sprintf("%s\t%s\t%s\t%.9g\n", r.Mechanism, r.Corpus, f, r.FamilyScores[f]))
		}
	}
	return writeFile(path, b.String())
}
