package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// readSensitivityClasses tallies PLAINTEXT_SENSITIVITY.tsv's classes, for
// FINAL_ARCHITECTURE.tsv's PLAINTEXT_DEPENDENCE_PRESERVED verdict.
func readSensitivityClasses(path string) map[string]int {
	out := map[string]int{}
	f, err := os.Open(path)
	if err != nil {
		return out
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	first := true
	for sc.Scan() {
		if first {
			first = false
			continue
		}
		fields := strings.Split(sc.Text(), "\t")
		if len(fields) >= 3 {
			out[fields[2]]++
		}
	}
	return out
}

// WriteManifest writes manifest.json (task66 section 76).
func WriteManifest(dir string, frontier []string, overfit map[string]string) error {
	m := map[string]any{
		"task":                          "Task66",
		"version":                       "mechanism-space-v1",
		"design_frozen":                 true,
		"target_mode":                   "authoritative-artifact-only",
		"inverse_search":                false,
		"worker_contract":               "immutable experiment_id/corpus/mechanism/config_hash/seed/evaluation_set",
		"authoritative_target_manifest": "VOYNICH_TARGET_MANIFEST.tsv",
		"development":                   "DEVELOPMENT_RESULTS.tsv",
		"held_out":                      "HELDOUT_RESULTS.tsv",
		"held_out_access":               "coordinator-after-candidate-freeze",
		"pareto_frontier":               frontier,
		"overfit_classification":        overfit,
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "manifest.json"), b, 0644)
}

func topFamilies(rows []FamilyMetricsRow, mechanism string, n int) string {
	type fv struct {
		f string
		v float64
	}
	var all []fv
	seen := map[string]bool{}
	for _, r := range rows {
		if r.Mechanism != mechanism {
			continue
		}
		for f, v := range r.FamilyScores {
			if !seen[f] {
				all = append(all, fv{f, v})
				seen[f] = true
			}
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].v > all[j].v })
	if len(all) > n {
		all = all[:n]
	}
	var parts []string
	for _, x := range all {
		parts = append(parts, fmt.Sprintf("%s=%.2g", x.f, x.v))
	}
	return strings.Join(parts, ", ")
}

// WriteReport writes REPORT.md, answering task66 section 77's twenty
// required questions directly from the computed tables - no claim here
// goes beyond "a mechanism with these operations moved several
// independent metric families toward Voynich's fingerprint" (section 71).
func WriteReport(dir string, targets []Target, devRows, ablationRows []FamilyMetricsRow, verdicts []FinalVerdict, frontier []string, overfit map[string]string, sensClasses map[string]int) error {
	var b strings.Builder
	b.WriteString("# Task66 report: mechanism-space search\n\n")
	b.WriteString("This is a statistical-compatibility study, not a decryption attempt and not a\n")
	b.WriteString("claim about the Voynich cipher. No inverse transformation was applied to\n")
	b.WriteString("Voynich; the Voynich sequence was only ever compared against, never mined for\n")
	b.WriteString("a plaintext (task66 sections 3-4, 71).\n\n")

	missing := 0
	for _, t := range targets {
		if t.Status != "VALUE" {
			missing++
		}
	}
	b.WriteString(fmt.Sprintf("Authoritative target manifest: %d metrics loaded from frozen Task58-65\nartifacts, %d MISSING_ARTIFACT.\n\n", len(targets), missing))

	b.WriteString("## 1. What do memoryless transformations explain?\n\n")
	b.WriteString(fmt.Sprintf("M1 (monoalphabetic): %s. M2 (homophony, H=4): %s.\n\n", topFamilies(devRows, "M1_MONOALPHABETIC", 7), topFamilies(devRows, "M2_HOMOPHONY_H4", 7)))

	b.WriteString("## 2-3. What does statefulness / slow persistence add?\n\n")
	b.WriteString(fmt.Sprintf("M4 (per-unit state, K=4, update A): %s.\nM5 (drift scale 20): %s.\n\n", topFamilies(devRows, "M4_STATE_K4_A", 7), topFamilies(devRows, "M5_DRIFT_N20", 7)))

	b.WriteString("## 4. Can slow state reproduce Task65 distance decay?\n\n")
	b.WriteString(fmt.Sprintf("See SLOW_STATE_REQUIRED in FINAL_ARCHITECTURE.tsv and TOPOLOGY_RESULTS.tsv's\ncorrelation_length_tokens column for M5 against the Voynich target row.\n\n"))

	b.WriteString("## 5-6. Are explicit macro-states needed? What creates MIXED_DRIFT_AND_STATES?\n\n")
	b.WriteString(fmt.Sprintf("M6 (macro only, K=5): %s.\nM7 (mixed, K=5, drift 20): %s.\nSee MACRO_STATE_REQUIRED in FINAL_ARCHITECTURE.tsv.\n\n", topFamilies(devRows, "M6_MACRO_K5", 7), topFamilies(devRows, "M7_MIXED_K5_N20", 7)))

	b.WriteString("## 7-8. Is constrained formation needed for Tasks59-62? Can state alone get token-internal fingerprint?\n\n")
	b.WriteString(fmt.Sprintf("Ablation G_ONLY: %s.\nAblation S_ONLY: %s.\nSee CONSTRAINED_FORMATION_REQUIRED and MEMORY_REQUIRED in FINAL_ARCHITECTURE.tsv.\n\n", topFamilies(ablationRows, "G_ONLY", 7), topFamilies(ablationRows, "S_ONLY", 7)))

	b.WriteString("## 9-10. Can form grammar alone get local topology? Are state+form compatible with Task58/63?\n\n")
	b.WriteString(fmt.Sprintf("Ablation G_PLUS_S: %s.\nAblation M_PLUS_S_PLUS_G: %s.\n\n", topFamilies(ablationRows, "G_PLUS_S", 7), topFamilies(ablationRows, "M_PLUS_S_PLUS_G", 7)))

	b.WriteString("## 11-12. What changes without word boundaries? Does generated grouping help?\n\n")
	b.WriteString(fmt.Sprintf("M9 (STREAM + generated boundaries + form): %s.\nSee GENERATED_BOUNDARIES_REQUIRED in FINAL_ARCHITECTURE.tsv (compares against\nM3's WORD_PRESERVING form-only result).\n\n", topFamilies(devRows, "M9_GROUP_FORM_STATE", 7)))

	b.WriteString("## 13-14. Does the mechanism retain real plaintext dependence, or does it ignore the input?\n\n")
	b.WriteString(fmt.Sprintf("Plaintext-sensitivity classes across representative mechanisms: %v.\nSee INFORMATION_RETENTION.tsv for coarse input/output mutual information and\nPLAINTEXT_DEPENDENCE_PRESERVED in FINAL_ARCHITECTURE.tsv.\n\n", sensClasses))

	b.WriteString("## 15-16. Does the result transfer Doyle -> Longfellow -> Astafiev? What survives held-out?\n\n")
	b.WriteString(fmt.Sprintf("Pareto frontier frozen before held-out was opened: %v.\nOverfit classification per candidate: %v.\nSee CORPUS_TRANSFER.tsv and HELDOUT_RESULTS.tsv.\n\n", frontier, overfit))

	b.WriteString("## 17. Which operations are required vs redundant?\n\n")
	for _, v := range verdicts {
		b.WriteString(fmt.Sprintf("- %s: **%s** (%s)\n", v.Operation, v.Verdict, v.Evidence))
	}
	b.WriteString("\n")

	b.WriteString("## 18. Error robustness\n\n")
	b.WriteString("See ERROR_ROBUSTNESS.tsv for the frontier's fingerprint degradation under\n0.1%-5% scribal-like error rates; robustness was not used to select the\nfrontier (task66 sections 67-68).\n\n")

	b.WriteString("## 19. Minimal architecture found\n\n")
	req := []string{}
	for _, v := range verdicts {
		if v.Verdict == "REQUIRED" || v.Verdict == "SUPPORTED" {
			req = append(req, v.Operation)
		}
	}
	if len(req) == 0 {
		b.WriteString("No operation class reached REQUIRED/SUPPORTED at this grid/replicate scale;\nsee FINAL_ARCHITECTURE.tsv for the full UNRESOLVED/NOT_REQUIRED breakdown.\n\n")
	} else {
		b.WriteString(fmt.Sprintf("Operations classed REQUIRED or SUPPORTED: %v. A transformation architecture\ncombining these is what this study found statistically compatible with several\nindependent Voynich fingerprint families - this is not a claim that Voynich\nused any such mechanism (task66 section 71).\n\n", req))
	}

	b.WriteString("## 20. What remains unexplained\n\n")
	b.WriteString("Any metric family whose best frontier candidate's progress stays below the\n0.15 movement threshold in FAMILY_METRICS.tsv/HELDOUT_RESULTS.tsv, and any\nfamily whose authoritative target artifact was MISSING_ARTIFACT in\nVOYNICH_TARGET_MANIFEST.tsv, is left unexplained by this study.\n")

	return os.WriteFile(filepath.Join(dir, "REPORT.md"), []byte(b.String()), 0644)
}
