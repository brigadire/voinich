package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"zcore.dev/voinich/internal/notation"
)

// perRepresentationResult holds everything computed for one
// (candidate_id, representation_id) pair, kept in memory only long enough
// to write it out and feed the aggregate/validation passes — it is never
// itself the only stored form of a measurement (task section 12).
type perRepresentationResult struct {
	CandidateID      string
	RepresentationID string
	Dir              string // relative to the run bundle root, e.g. "C01" or "C06/MUSIC-R1"
	Records          int
	Fingerprint      notation.Fingerprint
	RarefactionRows  []notation.RarefactionRow
	RarefactionSumm  []notation.RarefactionSummaryRow
	BootstrapRows    []notation.BootstrapRow
	Distributions    []notation.DistributionPoint
	Comparisons      map[int]vmComparisonAtCheckpoint // by checkpoint
}

type vmComparisonAtCheckpoint struct {
	Checkpoint int
	Reachable  bool // candidate's actual size >= checkpoint
	Rows       []notation.ComparisonRow
	Families   []notation.FamilyDistance
}

func representationDir(candidateID, representationID string) string {
	if candidateID == "C06" {
		return filepath.Join("C06", representationID)
	}
	return candidateID
}

// executeProductionRun implements task run03 end to end for one pass. When
// opts.OutDirOverride/RunIDOverride are set, this is the independent
// second pass used for reproducibility verification (section 17); its
// scientific outputs must be byte-identical to the primary pass, and it
// never mutates PRODUCTION_RUN_AUTHORIZATION.json or PRODUCTION_RUN_MANIFEST.json.
func executeProductionRun(opts productionRunOptions) (productionRunSummary, error) {
	repo := opts.Repo
	base := filepath.Join(repo, "research", "comparative_notation")

	reval := revalidateAuthorization(repo)
	if !reval.ok() {
		return productionRunSummary{Completed: false, Valid: false}, fmt.Errorf("revalidation failed, no computation started: %s", strings.Join(reval.Errors, "; "))
	}

	runID := opts.RunIDOverride
	if runID == "" {
		runID = newProductionRunID("CNS-PROD01-RUN")
	}
	bundleRoot := opts.OutDirOverride
	if bundleRoot == "" {
		bundleRoot = filepath.Join(base, "production_runs", runID)
	}
	if _, err := os.Stat(bundleRoot); err == nil {
		return productionRunSummary{}, fmt.Errorf("refusing to overwrite existing run bundle directory %s", bundleRoot)
	}
	if err := os.MkdirAll(bundleRoot, 0755); err != nil {
		return productionRunSummary{}, err
	}

	identity, err := buildRunIdentity(base, runID, reval)
	if err != nil {
		return productionRunSummary{}, err
	}
	if err := writeJSONFile(filepath.Join(bundleRoot, "RUN_MANIFEST.json"), identity); err != nil {
		return productionRunSummary{}, err
	}
	authRaw, err := os.ReadFile(filepath.Join(base, "PRODUCTION_RUN_AUTHORIZATION.json"))
	if err != nil {
		return productionRunSummary{}, err
	}
	if err := os.WriteFile(filepath.Join(bundleRoot, "AUTHORIZATION_SNAPSHOT.json"), authRaw, 0644); err != nil {
		return productionRunSummary{}, err
	}
	inputBindings := map[string]any{
		"schema_version": "production-run-input-bindings-1.0", "run_id": runID,
		"global_freeze_manifest_sha256": identity.GlobalFreezeManifestSHA256, "corpus_selection_sha256": identity.CorpusSelectionSHA256,
		"production_run_manifest_sha256": identity.RunManifestSHA256, "candidate_bundle_sha256": identity.CandidateBundleSHA256,
		"git_commit": identity.GitCommit,
	}
	if err := writeJSONFile(filepath.Join(bundleRoot, "INPUT_BINDINGS.json"), inputBindings); err != nil {
		return productionRunSummary{}, err
	}

	var cm corpusManifest
	cmRaw, err := os.ReadFile(filepath.Join(base, "PRODUCTION_CORPUS_MANIFEST.json"))
	if err != nil {
		return productionRunSummary{}, err
	}
	if err := json.Unmarshal(cmRaw, &cm); err != nil {
		return productionRunSummary{}, err
	}
	vmFP, vmManifest, err := loadFrozenVMReference(base)
	if err != nil {
		return productionRunSummary{}, err
	}
	calScales, err := loadFrozenCalibrationScales(base)
	if err != nil {
		return productionRunSummary{}, err
	}

	var results []*perRepresentationResult
	for _, id := range productionSubsetCandidateOrder {
		byRep, err := loadCandidateUSC(repo, base, cm, id)
		if err != nil {
			return productionRunSummary{}, err
		}
		var repIDs []string
		for rep := range byRep {
			repIDs = append(repIDs, rep)
		}
		sort.Strings(repIDs)
		for _, repID := range repIDs {
			res, err := computeOneRepresentation(id, repID, byRep[repID], vmFP, calScales)
			if err != nil {
				return productionRunSummary{}, fmt.Errorf("%s/%s: %w", id, repID, err)
			}
			if err := writeRepresentationOutputs(bundleRoot, res); err != nil {
				return productionRunSummary{}, fmt.Errorf("%s/%s: %w", id, repID, err)
			}
			results = append(results, res)
		}
	}

	if err := writePairedNotationDelta(bundleRoot, results); err != nil {
		return productionRunSummary{}, err
	}
	agg, err := computeAggregate(results)
	if err != nil {
		return productionRunSummary{}, err
	}
	if err := writeAggregateOutputs(bundleRoot, agg); err != nil {
		return productionRunSummary{}, err
	}
	if err := writeCalibrationReferenceRecord(base, bundleRoot, calScales); err != nil {
		return productionRunSummary{}, err
	}

	validation, err := postRunValidate(repo, base, bundleRoot, results, agg, vmManifest)
	if err != nil {
		return productionRunSummary{}, err
	}
	if err := writeJSONFile(filepath.Join(bundleRoot, "validation", "POST_RUN_VALIDATION.json"), validation); err != nil {
		return productionRunSummary{}, err
	}
	if err := os.WriteFile(filepath.Join(bundleRoot, "validation", "POST_RUN_VALIDATION.md"), []byte(validation.markdown()), 0644); err != nil {
		return productionRunSummary{}, err
	}

	completed := validation.allRequiredComputationsCompleted()
	summary := productionRunSummary{RunID: runID, GitCommit: identity.GitCommit, BundlePath: mustRel(repo, bundleRoot), Completed: completed}

	var repro reproducibilityResult
	if !opts.SkipReproducibility && completed {
		repro, err = runReproducibilityCheck(opts, bundleRoot, runID)
		if err != nil {
			return productionRunSummary{}, err
		}
	} else if opts.SkipReproducibility {
		repro = reproducibilityResult{Skipped: true}
	}
	if err := writeJSONFile(filepath.Join(bundleRoot, "validation", "REPRODUCIBILITY_REPORT.json"), repro); err != nil {
		return productionRunSummary{}, err
	}
	if err := os.WriteFile(filepath.Join(bundleRoot, "validation", "REPRODUCIBILITY_REPORT.md"), []byte(repro.markdown()), 0644); err != nil {
		return productionRunSummary{}, err
	}

	valid := completed && validation.valid() && (opts.SkipReproducibility || repro.Pass)

	if err := writeProductionResultChecksums(bundleRoot); err != nil {
		return productionRunSummary{}, err
	}

	report := buildRunReport(identity, results, validation, repro, completed, valid, mustRel(repo, bundleRoot))
	reportPath := filepath.Join(base, "PRODUCTION_COMPARATIVE_RUN_REPORT.md")
	if err := os.WriteFile(reportPath, []byte(report), 0644); err != nil {
		return productionRunSummary{}, err
	}
	if err := updateComparativeRunManifest(repo, base, runID, identity.GitCommit, completed, valid, bundleRoot); err != nil {
		return productionRunSummary{}, err
	}

	summary.Valid = valid
	summary.ReportPath = mustRel(repo, reportPath)
	return summary, nil
}

func mustRel(base, target string) string {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	return filepath.ToSlash(rel)
}

// computeOneRepresentation performs the frozen raw/rarefaction/bootstrap/
// distribution/VM-comparison computation for one (candidate,representation)
// pair (task run03 sections 3, 5, 6, 7, 8, 9, 10, 11). It never uses the
// technical pre-run (production_run/prerun/*) as a substitute for these
// replicates — every replicate here is computed fresh.
func computeOneRepresentation(candidateID, representationID string, records []notation.Record, vmFP notation.Fingerprint, calScales []notation.CalibrationScale) (*perRepresentationResult, error) {
	fp, err := notation.Analyze(records)
	if err != nil {
		return nil, err
	}
	rareRows, rareSumm, err := notation.RunRarefaction(records, candidateID, representationID, productionRunChecklist, notation.RarefactionReplicates, notation.BaseSeed)
	if err != nil {
		return nil, err
	}
	bootRows, err := notation.RunBootstrap(records, candidateID, representationID, notation.BootstrapReplicates, notation.BaseSeed)
	if err != nil {
		return nil, err
	}
	dist := notation.BuildDistributions(records, candidateID, representationID)

	comparisons := map[int]vmComparisonAtCheckpoint{}
	for _, ck := range productionRunChecklist {
		reachable := len(records) >= ck
		c := vmComparisonAtCheckpoint{Checkpoint: ck, Reachable: reachable}
		if reachable {
			scales := notation.ScalesFromCalibration(calScales, ck)
			rows, fam, err := notation.Compare(fp, vmFP, scales)
			if err != nil {
				return nil, fmt.Errorf("checkpoint %d: %w", ck, err)
			}
			c.Rows, c.Families = rows, fam
		} else {
			c.Rows, c.Families = notComparableRowsForRegistry(fmt.Sprintf("corpus size %d below checkpoint %d", len(records), ck))
		}
		comparisons[ck] = c
	}

	return &perRepresentationResult{
		CandidateID: candidateID, RepresentationID: representationID, Dir: representationDir(candidateID, representationID),
		Records: len(records), Fingerprint: fp, RarefactionRows: rareRows, RarefactionSumm: rareSumm,
		BootstrapRows: bootRows, Distributions: dist, Comparisons: comparisons,
	}, nil
}

// notComparableRowsForRegistry emits one NOT_COMPARABLE row per registered
// metric with an explicit reason (task section 11: never substitute a
// missing result with zero/blank without a schema-level reason).
func notComparableRowsForRegistry(reason string) ([]notation.ComparisonRow, []notation.FamilyDistance) {
	var rows []notation.ComparisonRow
	for _, m := range notation.MetricRegistry() {
		rows = append(rows, notation.ComparisonRow{MetricID: m.ID, Family: m.Family, Status: notation.NotComparable, Reason: reason})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Family != rows[j].Family {
			return rows[i].Family < rows[j].Family
		}
		return rows[i].MetricID < rows[j].MetricID
	})
	var fam []notation.FamilyDistance
	for _, f := range []string{"G", "T", "S", "L", "D"} {
		fam = append(fam, notation.FamilyDistance{Family: f, Status: notation.NotComparable, Reason: reason})
	}
	return rows, fam
}

func writeRepresentationOutputs(bundleRoot string, r *perRepresentationResult) error {
	fpOut, err := createOut(filepath.Join(bundleRoot, "candidates", r.Dir, "RAW_FINGERPRINT.json"))
	if err != nil {
		return err
	}
	if err := notation.WriteFingerprintJSON(fpOut, r.Fingerprint); err != nil {
		fpOut.Close()
		return err
	}
	fpOut.Close()

	if err := writeTSV(filepath.Join(bundleRoot, "candidates", r.Dir, "RAW_METRICS.tsv"), func(w *os.File) error { return notation.WriteMetricsTSV(w, r.Fingerprint) }); err != nil {
		return err
	}
	if err := writeTSV(filepath.Join(bundleRoot, "candidates", r.Dir, "CURVES.tsv"), func(w *os.File) error { return notation.WriteCurvesTSV(w, r.Fingerprint.Curves) }); err != nil {
		return err
	}
	if err := writeTSV(filepath.Join(bundleRoot, "distributions", r.Dir, "DISTRIBUTIONS.tsv"), func(w *os.File) error { return notation.WriteDistributionsTSV(w, r.Distributions) }); err != nil {
		return err
	}
	if err := writeTSV(filepath.Join(bundleRoot, "rarefaction", r.Dir, "RAREFACTION.tsv"), func(w *os.File) error { return notation.WriteRarefactionTSV(w, r.RarefactionRows) }); err != nil {
		return err
	}
	if err := writeTSV(filepath.Join(bundleRoot, "rarefaction", r.Dir, "RAREFACTION_SUMMARY.tsv"), func(w *os.File) error { return notation.WriteRarefactionSummaryTSV(w, r.RarefactionSumm) }); err != nil {
		return err
	}
	if err := writeTSV(filepath.Join(bundleRoot, "bootstrap", r.Dir, "BOOTSTRAP_RESULTS.tsv"), func(w *os.File) error { return notation.WriteBootstrapTSV(w, r.BootstrapRows) }); err != nil {
		return err
	}

	var checkpoints []int
	for ck := range r.Comparisons {
		checkpoints = append(checkpoints, ck)
	}
	sort.Ints(checkpoints)
	for _, ck := range checkpoints {
		c := r.Comparisons[ck]
		dir := filepath.Join(bundleRoot, "vm_comparison", r.Dir, fmt.Sprintf("checkpoint_%d", ck))
		if err := writeTSV(filepath.Join(dir, "VM_COMPARISON.tsv"), func(w *os.File) error { return notation.WriteComparisonTSV(w, c.Rows, c.Families) }); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "VM_COMPARISON.md"), []byte(vmComparisonMarkdown(r, c)), 0644); err != nil {
			return err
		}
	}
	if err := writeTSV(filepath.Join(bundleRoot, "vm_comparison", r.Dir, "VM_COMPARISON_UNCERTAINTY.tsv"), func(w *os.File) error { return writeUncertaintyTSV(w, r) }); err != nil {
		return err
	}
	return nil
}

func writeTSV(path string, fn func(*os.File) error) error {
	f, err := createOut(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return fn(f)
}

func vmComparisonMarkdown(r *perRepresentationResult, c vmComparisonAtCheckpoint) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# CORPUS VS VM STRUCTURAL REPORT\n\ncandidate_id=`%s` representation_id=`%s` checkpoint=`%d` reachable=`%t`\n\n", r.CandidateID, r.RepresentationID, c.Checkpoint, c.Reachable)
	b.WriteString("| Family | Comparable | Mean distance | Comparable metrics |\n|---|---|---:|---:|\n")
	for _, f := range c.Families {
		fmt.Fprintf(&b, "| %s | %s | %s | %d |\n", f.Family, f.Status, formatOptionalFloat(f.Status == notation.Comparable, f.Distance), f.ComparableMetrics)
	}
	var comparable []notation.ComparisonRow
	for _, row := range c.Rows {
		if row.Status == notation.Comparable {
			comparable = append(comparable, row)
		}
	}
	sort.Slice(comparable, func(i, j int) bool { return comparable[i].Distance < comparable[j].Distance })
	b.WriteString("\n## Strongest similarities (smallest calibrated distance)\n\n")
	writeTopRows(&b, comparable, 10)
	b.WriteString("\n## Strongest differences (largest calibrated distance)\n\n")
	reversed := make([]notation.ComparisonRow, len(comparable))
	copy(reversed, comparable)
	sort.Slice(reversed, func(i, j int) bool { return reversed[i].Distance > reversed[j].Distance })
	writeTopRows(&b, reversed, 10)
	b.WriteString("\n## Shared corpus rules / VM rules not reproduced / Candidate rules absent in VM\n\n")
	b.WriteString("NOT_COMPUTED: the generic USC/Fingerprint schema retains only aggregate metric values, not symbol- or rule-level detail, so per-rule reproduction cannot be derived without new, out-of-scope code.\n")
	b.WriteString("\n## Corpus-size sensitivity\n\n")
	fmt.Fprintf(&b, "See `rarefaction/%s/RAREFACTION_SUMMARY.tsv` for this candidate/representation's own metric estimates at every frozen checkpoint.\n", r.Dir)
	b.WriteString("\n## Comparability limitations\n\n")
	if !c.Reachable {
		fmt.Fprintf(&b, "Corpus size %d is below checkpoint %d; every metric is NOT_COMPARABLE at this checkpoint by frozen protocol rule.\n", r.Records, c.Checkpoint)
	} else {
		fmt.Fprintf(&b, "Corpus size %d reaches checkpoint %d.\n", r.Records, c.Checkpoint)
	}
	b.WriteString("\n## Result\n\n")
	b.WriteString("`STRUCTURALLY_CLOSE_ON: PENDING`  \n`STRUCTURALLY_DISTANT_ON: PENDING`  \n`NOT_COMPARABLE_ON: PENDING`\n\nNo frozen numeric threshold for STRUCTURALLY_CLOSE_ON/STRUCTURALLY_DISTANT_ON exists anywhere in the frozen protocol (VM_COMPARISON_TEMPLATE.md itself marks this PENDING); inventing one here would be an undocumented post-hoc parameter, so it stays PENDING for a separate task. This report contains no interpretation of what these distances mean about the Voynich Manuscript.\n")
	return b.String()
}

func writeTopRows(b *strings.Builder, rows []notation.ComparisonRow, n int) {
	if len(rows) == 0 {
		b.WriteString("(no comparable metrics at this checkpoint)\n")
		return
	}
	b.WriteString("| metric_id | family | distance |\n|---|---|---:|\n")
	for i, row := range rows {
		if i >= n {
			break
		}
		fmt.Fprintf(b, "| %s | %s | %.6g |\n", row.MetricID, row.Family, row.Distance)
	}
}

func formatOptionalFloat(ok bool, v float64) string {
	if !ok {
		return ""
	}
	return fmt.Sprintf("%.6g", v)
}

// writeUncertaintyTSV joins each reachable checkpoint's VM comparison rows
// with this representation's own bootstrap CI (computed at the
// representation's actual observed size) as the candidate value's
// uncertainty (task section 10 "uncertainty / CI").
func writeUncertaintyTSV(w *os.File, r *perRepresentationResult) error {
	bootByKey := map[string]notation.BootstrapRow{}
	for _, row := range r.BootstrapRows {
		bootByKey[row.MetricID+"\x1f"+row.Regime] = row
	}
	fmt.Fprintln(w, "candidate_id\trepresentation_id\tcheckpoint\treachable\tmetric_id\tfamily\tregime\tcandidate_value\tcandidate_ci_low\tcandidate_ci_high\tvm_value\tcalibrated_distance\tstatus\treason")
	var checkpoints []int
	for ck := range r.Comparisons {
		checkpoints = append(checkpoints, ck)
	}
	sort.Ints(checkpoints)
	for _, ck := range checkpoints {
		c := r.Comparisons[ck]
		for _, row := range c.Rows {
			ci, ok := bootByKey[row.MetricID+"\x1f"+row.Regime]
			ciLow, ciHigh := "", ""
			if ok && row.Status == notation.Comparable {
				ciLow, ciHigh = fmt.Sprintf("%.12g", ci.CILow), fmt.Sprintf("%.12g", ci.CIHigh)
			}
			candVal, vmVal, dist := "", "", ""
			if row.Status == notation.Comparable {
				candVal, vmVal, dist = fmt.Sprintf("%.12g", row.Candidate), fmt.Sprintf("%.12g", row.Reference), fmt.Sprintf("%.12g", row.Distance)
			}
			fmt.Fprintf(w, "%s\t%s\t%d\t%t\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				r.CandidateID, r.RepresentationID, ck, c.Reachable, row.MetricID, row.Family, row.Regime, candVal, ciLow, ciHigh, vmVal, dist, row.Status, row.Reason)
		}
	}
	return nil
}

// writePairedNotationDelta implements PAIRED_NOTATION_PROTOCOL.md for the
// only aligned pair in the frozen subset: C01 LATIN-EXPANDED and C02
// LATIN-DIPLOMATIC.
func writePairedNotationDelta(bundleRoot string, results []*perRepresentationResult) error {
	var c01, c02 *perRepresentationResult
	for _, r := range results {
		if r.CandidateID == "C01" {
			c01 = r
		}
		if r.CandidateID == "C02" {
			c02 = r
		}
	}
	if c01 == nil || c02 == nil {
		return fmt.Errorf("paired notation delta requires both C01 and C02 results")
	}
	delta := notation.NotationDelta(c01.Fingerprint, c02.Fingerprint)
	return writeTSV(filepath.Join(bundleRoot, "vm_comparison", "PAIRED", "C01_C02_NOTATION_DELTA.tsv"), func(w *os.File) error {
		fmt.Fprintln(w, "metric_id\tfamily\tregime\tleft_c01_expanded\tright_c02_diplomatic\tdelta\tstatus\treason")
		for _, d := range delta {
			left, right, dv := "", "", ""
			if d.Status == notation.Comparable {
				left, right, dv = fmt.Sprintf("%.12g", d.Left), fmt.Sprintf("%.12g", d.Right), fmt.Sprintf("%.12g", d.Delta)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", d.MetricID, d.Family, d.Regime, left, right, dv, d.Status, d.Reason)
		}
		return nil
	})
}

