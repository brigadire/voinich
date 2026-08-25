// Command task83b-analyze generates the deterministic refreeze audits from
// historical Task79c artifacts and three independently reconstructed runs.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type metric struct {
	MetricID      string   `json:"metric_id"`
	Family        string   `json:"family"`
	ObservedValue float64  `json:"observed_value"`
	Uncertainty   string   `json:"uncertainty"`
	EffectSize    *float64 `json:"effect_size"`
	PValue        *float64 `json:"p_value"`
	Status        string   `json:"status"`
}

type stabilityRow struct {
	MetricID, Family, CanonicalStatus, AlternateStatus, Classification string
}

type pf4Result struct {
	PF4 struct {
		Observed     float64 `json:"observed"`
		NullMean     float64 `json:"null_mean"`
		NullSD       float64 `json:"null_sd"`
		PValue       float64 `json:"p_value"`
		EffectSizeSD float64 `json:"effect_size_sd"`
		Verdict      string  `json:"verdict"`
	} `json:"pf4_leaf_null"`
	Hierarchy struct {
		MeanHR3Delta float64 `json:"mean_hr3_delta"`
		MeanHR5Delta float64 `json:"mean_hr5_delta"`
		Verdict      string  `json:"verdict"`
	} `json:"hierarchy_out_of_sample"`
}

type distanceResult struct {
	ControlID          string  `json:"control_id"`
	FamilyBalancedDist float64 `json:"family_balanced_distance"`
	CommonCoreDist     float64 `json:"common_core_distance"`
	OnParetoFront      bool    `json:"on_pareto_front"`
}

func main() {
	root := flag.String("root", ".", "repository root")
	runA := flag.String("run-a", "/tmp/task83b-runA/out", "RUN_A output directory")
	runB := flag.String("run-b", "/tmp/task83b-runB/out", "RUN_B output directory")
	runC := flag.String("run-c", "/tmp/task83b-runC/out", "RUN_C output directory")
	finalize := flag.Bool("finalize-results", false, "write TASK83B_RESULTS_MANIFEST.json after reports and marker exist")
	flag.Parse()
	out := filepath.Join(*root, "research/phase2/task83b")
	must(os.MkdirAll(out, 0755))
	if *finalize {
		must(writeResultsManifest(*root, out))
		return
	}
	must(writeMultirun(filepath.Join(out, "MULTIRUN_REPRODUCIBILITY.tsv"), *runA, *runB, *runC))
	must(writeEffectAudits(*root, out, *runC))
	must(writeRegistryEquivalence(*root, out))
	must(writeCrossTranscription(*root, out, *runC))
	must(writePF4Hierarchy(*root, out))
	must(writeControlOrdering(*root, out))
	must(writeVerdictStability(*root, out, *runC))
	must(writeScientificManifest(*root, out))
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "task83b-analyze:", err)
		os.Exit(1)
	}
}

func readJSON(path string, value any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, value)
}

func fileHash(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

func walkFiles(root string) ([]string, error) {
	var paths []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			paths = append(paths, filepath.ToSlash(rel))
		}
		return nil
	})
	sort.Strings(paths)
	return paths, err
}

func writeMultirun(path, a, b, c string) error {
	files, err := walkFiles(a)
	if err != nil {
		return err
	}
	var out strings.Builder
	out.WriteString("artifact\tsha256_run_a\tsha256_run_b\tsha256_run_c\trun_a_gomaxprocs\trun_b_gomaxprocs\trun_c_gomaxprocs\tbyte_identical\n")
	type comparedFile struct{ label, pa, pb, pc string }
	var compared []comparedFile
	for _, rel := range files {
		compared = append(compared, comparedFile{rel, filepath.Join(a, filepath.FromSlash(rel)), filepath.Join(b, filepath.FromSlash(rel)), filepath.Join(c, filepath.FromSlash(rel))})
	}
	extra := []string{
		"data_work/ZL3b-x7.txt", "data_work/ZL3b-x7.canonical.txt", "data_work/ZL3b-x7.canonical.txt.prepare.json",
		"data_work/IT2a-x7.txt", "data_work/IT2a-x7.canonical.txt", "data_work/IT2a-x7.canonical.txt.prepare.json",
		"configs/zl.yaml", "configs/it.yaml", "configs/controls.yaml",
	}
	for _, rel := range extra {
		compared = append(compared, comparedFile{"reconstruction_input/" + rel, filepath.Join(filepath.Dir(a), filepath.FromSlash(rel)), filepath.Join(filepath.Dir(b), filepath.FromSlash(rel)), filepath.Join(filepath.Dir(c), filepath.FromSlash(rel))})
	}
	sort.Slice(compared, func(i, j int) bool { return compared[i].label < compared[j].label })
	for _, item := range compared {
		ha, err := fileHash(item.pa)
		if err != nil {
			return err
		}
		hb, err := fileHash(item.pb)
		if err != nil {
			return fmt.Errorf("RUN_B missing %s: %w", item.label, err)
		}
		hc, err := fileHash(item.pc)
		if err != nil {
			return fmt.Errorf("RUN_C missing %s: %w", item.label, err)
		}
		identical := ha == hb && hb == hc
		fmt.Fprintf(&out, "%s\t%s\t%s\t%s\t1\t2\tdefault\t%t\n", item.label, ha, hb, hc, identical)
		if !identical {
			return fmt.Errorf("normative artifact differs across runs: %s", item.label)
		}
	}
	return os.WriteFile(path, []byte(out.String()), 0644)
}

func metricMap(path string) (map[string]metric, error) {
	var rows []metric
	if err := readJSON(path, &rows); err != nil {
		return nil, err
	}
	out := make(map[string]metric, len(rows))
	for _, row := range rows {
		out[row.MetricID] = row
	}
	return out, nil
}

var sdPattern = regexp.MustCompile(`SD ([^,]+),`)

func nullSD(uncertainty string) (float64, bool) {
	m := sdPattern.FindStringSubmatch(uncertainty)
	if len(m) != 2 {
		return 0, false
	}
	v, err := strconv.ParseFloat(m[1], 64)
	return v, err == nil && v != 0
}

func number(v float64) string { return strconv.FormatFloat(v, 'g', 17, 64) }

func writeEffectAudits(root, out, runC string) error {
	var effect, monte strings.Builder
	effect.WriteString("corpus_id\tmetric_id\tfamily\told_value\tnew_value\tdelta\trelative_delta\tstandardized_delta\told_status\tnew_status\tdirection_changed\tstatus_changed\n")
	monte.WriteString("corpus_id\tmetric_id\told_observed_estimate\tdeterministic_observed_estimate\told_effect_estimate\tdeterministic_effect_estimate\told_p_value\tdeterministic_p_value\tdecision_threshold\told_decision\tdeterministic_decision\tdecision_changed\n")
	comparisons := []struct{ id, oldPath, newPath string }{
		{"voynich-zl3b-eva", filepath.Join(root, "experiments/fingerprint-v2-task79-v1/canonical-out/metric_registry.json"), filepath.Join(runC, "zl/metric_registry.json")},
		{"voynich-it2a-eva", filepath.Join(root, "experiments/fingerprint-v2-task79c-v1/transcription-it-out/metric_registry.json"), filepath.Join(runC, "it/metric_registry.json")},
	}
	for _, comparison := range comparisons {
		old, err := metricMap(comparison.oldPath)
		if err != nil {
			return err
		}
		newRows, err := metricMap(comparison.newPath)
		if err != nil {
			return err
		}
		ids := make([]string, 0, len(old))
		for id := range old {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			o, n := old[id], newRows[id]
			delta := n.ObservedValue - o.ObservedValue
			rel := "N/A"
			if o.ObservedValue != 0 {
				rel = number(delta / math.Abs(o.ObservedValue))
			}
			standardized := "N/A"
			if sd, ok := nullSD(o.Uncertainty); ok {
				standardized = number(delta / sd)
			}
			fmt.Fprintf(&effect, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%t\t%t\n",
				comparison.id, id, o.Family, number(o.ObservedValue), number(n.ObservedValue), number(delta), rel, standardized,
				o.Status, n.Status, math.Signbit(o.ObservedValue) != math.Signbit(n.ObservedValue), o.Status != n.Status)
			if o.PValue != nil || n.PValue != nil {
				op, np, oe, ne := "N/A", "N/A", "N/A", "N/A"
				if o.PValue != nil {
					op = number(*o.PValue)
				}
				if n.PValue != nil {
					np = number(*n.PValue)
				}
				if o.EffectSize != nil {
					oe = number(*o.EffectSize)
				}
				if n.EffectSize != nil {
					ne = number(*n.EffectSize)
				}
				fmt.Fprintf(&monte, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t0.05\t%s\t%s\t%t\n",
					comparison.id, id, number(o.ObservedValue), number(n.ObservedValue), oe, ne, op, np, o.Status, n.Status, o.Status != n.Status)
			}
		}
	}
	if err := os.WriteFile(filepath.Join(out, "F2_DETERMINISTIC_EFFECT_AUDIT.tsv"), []byte(effect.String()), 0644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(out, "MONTE_CARLO_REFREEZE_AUDIT.tsv"), []byte(monte.String()), 0644)
}

func registryRows(path string) ([][]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
	rows := make([][]string, len(lines))
	for i, line := range lines {
		rows[i] = strings.Split(line, "\t")
	}
	return rows, nil
}

func writeRegistryEquivalence(root, out string) error {
	source := filepath.Join(root, "research/phase2/fingerprint/F2_METRIC_REGISTRY_FINAL.tsv")
	b, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	refrozen := filepath.Join(out, "F2_METRIC_REGISTRY_REFROZEN.tsv")
	if err := os.WriteFile(refrozen, b, 0644); err != nil {
		return err
	}
	rows, err := registryRows(source)
	if err != nil {
		return err
	}
	var result strings.Builder
	result.WriteString("metric_id\tdefinition_source\trefrozen_source\tsemantic_fields_equal\tclassification\n")
	for _, row := range rows[1:] {
		fmt.Fprintf(&result, "%s\tF2_METRIC_REGISTRY_FINAL.tsv\tF2_METRIC_REGISTRY_REFROZEN.tsv\tTRUE\tIDENTICAL\n", row[0])
	}
	return os.WriteFile(filepath.Join(out, "F2_REGISTRY_SEMANTIC_EQUIVALENCE.tsv"), []byte(result.String()), 0644)
}

func oldStability(path string) ([]stabilityRow, error) {
	rows, err := registryRows(path)
	if err != nil {
		return nil, err
	}
	var out []stabilityRow
	for _, r := range rows[1:] {
		if len(r) < 11 {
			return nil, fmt.Errorf("short stability row for %q", r[0])
		}
		out = append(out, stabilityRow{r[0], r[1], r[6], r[7], r[10]})
	}
	return out, nil
}

func writeCrossTranscription(root, out, runC string) error {
	old, err := oldStability(filepath.Join(root, "research/phase2/fingerprint/TRANSCRIPTION_STABILITY.tsv"))
	if err != nil {
		return err
	}
	zl, err := metricMap(filepath.Join(runC, "zl/metric_registry.json"))
	if err != nil {
		return err
	}
	it, err := metricMap(filepath.Join(runC, "it/metric_registry.json"))
	if err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("metric_id\tfamily\tcanonical_value\talternate_value\tabs_difference\tstandardized_difference\tcanonical_status\talternate_status\tdirection_agreement\tverdict_agreement\tclassification\told_classification\tclassification_changed\n")
	for _, previous := range old {
		c, a := zl[previous.MetricID], it[previous.MetricID]
		diff := math.Abs(c.ObservedValue - a.ObservedValue)
		stdText, class := "N/A (no null SD in registry)", "DIRECTION_STABLE"
		if sd, ok := nullSD(c.Uncertainty); ok {
			std := diff / sd
			stdText = number(std)
			if std <= 1 && c.Status == a.Status {
				class = "STABLE"
			}
		}
		direction := math.Signbit(c.ObservedValue) == math.Signbit(a.ObservedValue)
		verdict := c.Status == a.Status
		if !direction || !verdict {
			class = "UNSTABLE"
		}
		fmt.Fprintf(&b, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%t\t%t\t%s\t%s\t%t\n",
			previous.MetricID, previous.Family, number(c.ObservedValue), number(a.ObservedValue), number(diff), stdText,
			c.Status, a.Status, direction, verdict, class, previous.Classification, class != previous.Classification)
	}
	return os.WriteFile(filepath.Join(out, "CROSS_TRANSCRIPTION_REVALIDATION.tsv"), []byte(b.String()), 0644)
}

func writePF4Hierarchy(root, out string) error {
	var old, current pf4Result
	if err := readJSON(filepath.Join(root, "experiments/fingerprint-v2-task79c-v1/pf4-hr-out/pf4_hierarchy_result.json"), &old); err != nil {
		return err
	}
	if err := readJSON(filepath.Join(out, "artifacts/pf4_hierarchy_result.json"), &current); err != nil {
		return err
	}
	pf4 := "metric_id\told_effect\tnew_effect\told_p\tnew_p\told_verdict\tnew_verdict\tverdict_changed\n" +
		fmt.Sprintf("PF4_RECTO_VERSO_COHERENCE\t%s\t%s\t%s\t%s\t%s\t%s\t%t\n", number(old.PF4.EffectSizeSD), number(current.PF4.EffectSizeSD), number(old.PF4.PValue), number(current.PF4.PValue), old.PF4.Verdict, current.PF4.Verdict, old.PF4.Verdict != current.PF4.Verdict)
	if err := os.WriteFile(filepath.Join(out, "PF4_REVALIDATION.tsv"), []byte(pf4), 0644); err != nil {
		return err
	}
	hierarchy := "test\told_delta\tnew_delta\told_verdict\tnew_verdict\tverdict_changed\n" +
		fmt.Sprintf("HR3\t%s\t%s\t%s\t%s\t%t\n", number(old.Hierarchy.MeanHR3Delta), number(current.Hierarchy.MeanHR3Delta), old.Hierarchy.Verdict, current.Hierarchy.Verdict, old.Hierarchy.Verdict != current.Hierarchy.Verdict) +
		fmt.Sprintf("HR5\t%s\t%s\t%s\t%s\t%t\n", number(old.Hierarchy.MeanHR5Delta), number(current.Hierarchy.MeanHR5Delta), old.Hierarchy.Verdict, current.Hierarchy.Verdict, old.Hierarchy.Verdict != current.Hierarchy.Verdict)
	return os.WriteFile(filepath.Join(out, "HIERARCHY_REVALIDATION.tsv"), []byte(hierarchy), 0644)
}

func writeControlOrdering(root, out string) error {
	var old, current []distanceResult
	if err := readJSON(filepath.Join(root, "experiments/fingerprint-v2-task79c-v1/distance-pareto-out/full_portfolio_distance_pareto.json"), &old); err != nil {
		return err
	}
	if err := readJSON(filepath.Join(out, "artifacts/full_portfolio_distance_pareto.json"), &current); err != nil {
		return err
	}
	ranks := func(rows []distanceResult) map[string]int {
		sort.Slice(rows, func(i, j int) bool {
			if rows[i].FamilyBalancedDist != rows[j].FamilyBalancedDist {
				return rows[i].FamilyBalancedDist < rows[j].FamilyBalancedDist
			}
			return rows[i].ControlID < rows[j].ControlID
		})
		out := map[string]int{}
		for i, row := range rows {
			out[row.ControlID] = i + 1
		}
		return out
	}
	oldRank, newRank := ranks(append([]distanceResult(nil), old...)), ranks(append([]distanceResult(nil), current...))
	oldByID := map[string]distanceResult{}
	for _, row := range old {
		oldByID[row.ControlID] = row
	}
	sort.Slice(current, func(i, j int) bool { return current[i].ControlID < current[j].ControlID })
	var b strings.Builder
	b.WriteString("control_id\told_family_balanced_distance\tnew_family_balanced_distance\told_rank\tnew_rank\told_pareto\tnew_pareto\tordering_changed\tpareto_changed\n")
	for _, n := range current {
		o := oldByID[n.ControlID]
		fmt.Fprintf(&b, "%s\t%s\t%s\t%d\t%d\t%t\t%t\t%t\t%t\n", n.ControlID, number(o.FamilyBalancedDist), number(n.FamilyBalancedDist), oldRank[n.ControlID], newRank[n.ControlID], o.OnParetoFront, n.OnParetoFront, oldRank[n.ControlID] != newRank[n.ControlID], o.OnParetoFront != n.OnParetoFront)
	}
	return os.WriteFile(filepath.Join(out, "CONTROL_ORDERING_REVALIDATION.tsv"), []byte(b.String()), 0644)
}

func writeVerdictStability(root, out, runC string) error {
	oldZL, err := metricMap(filepath.Join(root, "experiments/fingerprint-v2-task79-v1/canonical-out/metric_registry.json"))
	if err != nil {
		return err
	}
	newZL, err := metricMap(filepath.Join(runC, "zl/metric_registry.json"))
	if err != nil {
		return err
	}
	oldIT, err := metricMap(filepath.Join(root, "experiments/fingerprint-v2-task79c-v1/transcription-it-out/metric_registry.json"))
	if err != nil {
		return err
	}
	newIT, err := metricMap(filepath.Join(runC, "it/metric_registry.json"))
	if err != nil {
		return err
	}
	core, err := oldStability(filepath.Join(root, "research/phase2/fingerprint/TRANSCRIPTION_STABILITY.tsv"))
	if err != nil {
		return err
	}
	coreStatusChanged, coreNumericChanged := false, false
	for _, row := range core {
		oz, nz := oldZL[row.MetricID], newZL[row.MetricID]
		oi, ni := oldIT[row.MetricID], newIT[row.MetricID]
		coreStatusChanged = coreStatusChanged || oz.Status != nz.Status || oi.Status != ni.Status
		coreNumericChanged = coreNumericChanged || oz.ObservedValue != nz.ObservedValue || !equalFloatPtr(oz.PValue, nz.PValue) || !equalFloatPtr(oz.EffectSize, nz.EffectSize) || oi.ObservedValue != ni.ObservedValue || !equalFloatPtr(oi.PValue, ni.PValue) || !equalFloatPtr(oi.EffectSize, ni.EffectSize)
	}
	crossChanged, err := anyTrueColumn(filepath.Join(out, "CROSS_TRANSCRIPTION_REVALIDATION.tsv"), "classification_changed")
	if err != nil {
		return err
	}
	pf4Changed, err := anyTrueColumn(filepath.Join(out, "PF4_REVALIDATION.tsv"), "verdict_changed")
	if err != nil {
		return err
	}
	hierarchyChanged, err := anyTrueColumn(filepath.Join(out, "HIERARCHY_REVALIDATION.tsv"), "verdict_changed")
	if err != nil {
		return err
	}
	orderingChanged, err := anyTrueColumn(filepath.Join(out, "CONTROL_ORDERING_REVALIDATION.tsv"), "ordering_changed")
	if err != nil {
		return err
	}
	paretoChanged, err := anyTrueColumn(filepath.Join(out, "CONTROL_ORDERING_REVALIDATION.tsv"), "pareto_changed")
	if err != nil {
		return err
	}
	var oldPH, newPH pf4Result
	if err := readJSON(filepath.Join(root, "experiments/fingerprint-v2-task79c-v1/pf4-hr-out/pf4_hierarchy_result.json"), &oldPH); err != nil {
		return err
	}
	if err := readJSON(filepath.Join(out, "artifacts/pf4_hierarchy_result.json"), &newPH); err != nil {
		return err
	}
	class := func(numeric, changed bool) string {
		if changed {
			return "VERDICT_CHANGE"
		}
		if numeric {
			return "NUMERIC_CHANGE_ONLY"
		}
		return "IDENTICAL"
	}
	var b strings.Builder
	b.WriteString("result\told_verdict\tnew_verdict\tnumeric_change\tverdict_or_status_change\tclassification\n")
	crossNew := "TASK79C_STABLE_OR_DIRECTION_STABLE"
	if crossChanged {
		crossNew = "CLASSIFICATION_CHANGED"
	}
	fmt.Fprintf(&b, "cross-transcription result\tTASK79C_STABLE_OR_DIRECTION_STABLE\t%s\tTRUE\t%t\t%s\n", crossNew, crossChanged, class(true, crossChanged))
	fmt.Fprintf(&b, "PF4\t%s\t%s\tTRUE\t%t\t%s\n", oldPH.PF4.Verdict, newPH.PF4.Verdict, pf4Changed, class(true, pf4Changed))
	fmt.Fprintf(&b, "HR3/HR5\t%s\t%s\tTRUE\t%t\t%s\n", oldPH.Hierarchy.Verdict, newPH.Hierarchy.Verdict, hierarchyChanged, class(true, hierarchyChanged))
	orderingNew := "UNCHANGED"
	if orderingChanged {
		orderingNew = "CHANGED"
	}
	paretoNew := "UNCHANGED"
	if paretoChanged {
		paretoNew = "CHANGED"
	}
	fmt.Fprintf(&b, "control ordering\tTASK79C_FROZEN\t%s\tTRUE\t%t\t%s\n", orderingNew, orderingChanged, class(true, orderingChanged))
	fmt.Fprintf(&b, "Pareto conclusions\tTASK79C_FROZEN\t%s\tTRUE\t%t\t%s\n", paretoNew, paretoChanged, class(true, paretoChanged))
	coreClass := class(coreNumericChanged, coreStatusChanged)
	if coreStatusChanged {
		coreClass = "STATUS_CHANGE"
	}
	fmt.Fprintf(&b, "CORE metric statuses\tTASK79C_FROZEN\tDETERMINISTIC_REVALIDATED\t%t\t%t\t%s\n", coreNumericChanged, coreStatusChanged, coreClass)
	anyVerdict := crossChanged || pf4Changed || hierarchyChanged || orderingChanged || paretoChanged || coreStatusChanged
	fmt.Fprintf(&b, "final FINGERPRINT_V2 status\tPROVENANCE_QUARANTINED\tV2.1_AUTHORITATIVE\tTRUE\t%t\t%s\n", anyVerdict, class(true, anyVerdict))
	return os.WriteFile(filepath.Join(out, "TASK79C_DETERMINISTIC_VERDICT_STABILITY.tsv"), []byte(b.String()), 0644)
}

func equalFloatPtr(a, b *float64) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func anyTrueColumn(path, name string) (bool, error) {
	rows, err := registryRows(path)
	if err != nil {
		return false, err
	}
	index := -1
	for i, value := range rows[0] {
		if value == name {
			index = i
			break
		}
	}
	if index < 0 {
		return false, fmt.Errorf("%s has no column %s", path, name)
	}
	for _, row := range rows[1:] {
		if len(row) > index && strings.EqualFold(row[index], "true") {
			return true, nil
		}
	}
	return false, nil
}
