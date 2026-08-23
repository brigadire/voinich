package main

import (
	"fmt"
	"sort"
	"strings"

	"zcore.dev/voinich/internal/mechanismspace"
)

// RunAblation executes task66 section 26's compositional ablation matrix
// (G only / S only / M only / G+S / M+S / M+G / M+S+G) on every corpus and
// writes ARCHITECTURE_ABLATION.tsv, using the same targets/baselines as
// the main grid so each row's family scores are directly comparable.
func RunAblation(corpora map[string]mechanismspace.Corpus, baselines map[string]mechanismspace.Fingerprint, targets []Target, replicates int, opt mechanismspace.FingerprintOptions, path string) []FamilyMetricsRow {
	grid := AblationEntries()
	results := RunGrid(grid, corpora, replicates, opt, 500000)
	grouped := GroupByMechanismCorpus(results)
	rows := ComputeFamilyMetrics(grouped, baselines, targets, "DEVELOPMENT_ABLATION", StageDevelopment)
	writeAblationTSV(path, rows)
	return rows
}

func writeAblationTSV(path string, rows []FamilyMetricsRow) {
	var b strings.Builder
	b.WriteString("operation_combination\tcorpus\tfamily\tprogress\toverall_status\n")
	for _, r := range rows {
		for _, f := range sortedFamilyNames(r.FamilyScores) {
			b.WriteString(fmt.Sprintf("%s\t%s\t%s\t%.9g\t%s\n", r.Mechanism, r.Corpus, f, r.FamilyScores[f], r.OverallStatus))
		}
	}
	writeFile(path, b.String())
}

func sortedFamilyNames(m map[string]float64) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
