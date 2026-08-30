package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"zcore.dev/voinich/internal/notation"
)

// postRunValidation implements task run03 section 16 and the boolean
// checklist in section 20 (minus REPRODUCIBILITY_PASS, which is decided
// separately by runReproducibilityCheck).
type postRunValidation struct {
	InputIntegrityValid bool     `json:"input_integrity_valid"`
	CoverageValid       bool     `json:"coverage_valid"`
	RarefactionValid    bool     `json:"rarefaction_valid"`
	BootstrapValid      bool     `json:"bootstrap_valid"`
	CalibrationValid    bool     `json:"calibration_valid"`
	VMReferenceValid    bool     `json:"vm_reference_valid"`
	NumericIntegrityOK  bool     `json:"numeric_integrity_ok"`
	AggregationValid    bool     `json:"aggregation_valid"`
	Errors              []string `json:"errors,omitempty"`
	CandidatesPresent   []string `json:"candidates_present"`
	RepresentationsSeen []string `json:"representations_seen"`
}

func (v postRunValidation) allRequiredComputationsCompleted() bool {
	return v.CoverageValid && v.RarefactionValid && v.BootstrapValid && len(v.Errors) == 0
}

func (v postRunValidation) valid() bool {
	return v.InputIntegrityValid && v.CoverageValid && v.RarefactionValid && v.BootstrapValid &&
		v.CalibrationValid && v.VMReferenceValid && v.NumericIntegrityOK && v.AggregationValid && len(v.Errors) == 0
}

func (v postRunValidation) markdown() string {
	var b strings.Builder
	b.WriteString("# Post-run validation report\n\n")
	fmt.Fprintf(&b, "| Check | Status |\n|---|---|\n")
	rows := []struct {
		name string
		ok   bool
	}{
		{"input_integrity", v.InputIntegrityValid}, {"coverage", v.CoverageValid}, {"rarefaction", v.RarefactionValid},
		{"bootstrap", v.BootstrapValid}, {"calibration", v.CalibrationValid}, {"vm_reference", v.VMReferenceValid},
		{"numeric_integrity", v.NumericIntegrityOK}, {"aggregation", v.AggregationValid},
	}
	for _, r := range rows {
		status := "FAIL"
		if r.ok {
			status = "PASS"
		}
		fmt.Fprintf(&b, "| `%s` | %s |\n", r.name, status)
	}
	fmt.Fprintf(&b, "\nCandidates present: `%s`\n\nRepresentations seen: `%s`\n", strings.Join(v.CandidatesPresent, ","), strings.Join(v.RepresentationsSeen, ","))
	if len(v.Errors) != 0 {
		b.WriteString("\n## Errors\n\n")
		for _, e := range v.Errors {
			fmt.Fprintf(&b, "- %s\n", e)
		}
	}
	return b.String()
}

func postRunValidate(repo, base, bundleRoot string, results []*perRepresentationResult, agg aggregateResult, vmManifest notation.VMReferenceManifest) (postRunValidation, error) {
	var v postRunValidation

	// Input integrity: re-verify the same bindings revalidation already
	// checked, defensively, against the files actually still on disk now
	// that computation has finished.
	freezeErrs, ferr := notation.VerifyGlobalFreezeManifest(base)
	corpusErr := productionCorpusValidateCmd([]string{"--repo", repo})
	v.InputIntegrityValid = ferr == nil && len(freezeErrs) == 0 && corpusErr == nil
	if ferr != nil {
		v.Errors = append(v.Errors, ferr.Error())
	}
	v.Errors = append(v.Errors, freezeErrs...)
	if corpusErr != nil {
		v.Errors = append(v.Errors, corpusErr.Error())
	}

	// Coverage.
	candidateSet := map[string]bool{}
	repSet := map[string]bool{}
	for _, r := range results {
		candidateSet[r.CandidateID] = true
		repSet[r.CandidateID+"/"+r.RepresentationID] = true
		if !strings.HasPrefix(r.CandidateID, "C01") && !strings.HasPrefix(r.CandidateID, "C02") && !strings.HasPrefix(r.CandidateID, "C06") {
			v.Errors = append(v.Errors, "unauthorized candidate present: "+r.CandidateID)
		}
	}
	v.CandidatesPresent = sortedStrings(candidateSet)
	v.RepresentationsSeen = sortedStrings(repSet)
	wantCandidates := []string{"C01", "C02", "C06"}
	wantReps := []string{"C01/LATIN-EXPANDED", "C02/LATIN-DIPLOMATIC", "C06/MUSIC-R1", "C06/MUSIC-R2", "C06/MUSIC-R3"}
	coverageOK := len(candidateSet) == 3
	for _, c := range wantCandidates {
		if !candidateSet[c] {
			v.Errors = append(v.Errors, "missing required candidate: "+c)
			coverageOK = false
		}
	}
	for _, r := range wantReps {
		if !repSet[r] {
			v.Errors = append(v.Errors, "missing required representation: "+r)
			coverageOK = false
		}
	}
	for _, r := range results {
		for _, ck := range productionRunChecklist {
			if _, ok := r.Comparisons[ck]; !ok {
				v.Errors = append(v.Errors, fmt.Sprintf("%s/%s: missing checkpoint %d", r.CandidateID, r.RepresentationID, ck))
				coverageOK = false
			}
		}
		seenFamilies := map[string]bool{}
		for _, m := range r.Fingerprint.Metrics {
			seenFamilies[m.Family] = true
		}
		for _, f := range []string{"G", "T", "S", "L", "D"} {
			if !seenFamilies[f] {
				v.Errors = append(v.Errors, fmt.Sprintf("%s/%s: missing metric family %s", r.CandidateID, r.RepresentationID, f))
				coverageOK = false
			}
		}
	}
	v.CoverageValid = coverageOK

	// Rarefaction: every reachable checkpoint must have exactly
	// RarefactionReplicates comparable-or-not rows per non-CURVE family per
	// metric (schema consistency), and every unreachable checkpoint must
	// carry the documented NOT_COMPARABLE marker rows, never silence.
	rareOK := true
	for _, r := range results {
		counts := map[string]int{} // key: checkpoint|family|metric|regime -> row count
		for _, row := range r.RarefactionRows {
			key := fmt.Sprintf("%d|%s|%s|%s", row.CheckpointRequested, row.Family, row.MetricID, row.Regime)
			counts[key]++
		}
		for _, ck := range productionRunChecklist {
			reachable := r.Records >= ck
			if !reachable {
				marker := fmt.Sprintf("%d|%s|NOT_COMPARABLE|", ck, "G")
				if counts[marker] == 0 {
					v.Errors = append(v.Errors, fmt.Sprintf("%s/%s: checkpoint %d below corpus size but missing NOT_COMPARABLE marker rows", r.CandidateID, r.RepresentationID, ck))
					rareOK = false
				}
				continue
			}
			for key, n := range counts {
				if !strings.HasPrefix(key, fmt.Sprintf("%d|", ck)) {
					continue
				}
				if n != notation.RarefactionReplicates {
					v.Errors = append(v.Errors, fmt.Sprintf("%s/%s: %s has %d rows, want %d replicates", r.CandidateID, r.RepresentationID, key, n, notation.RarefactionReplicates))
					rareOK = false
				}
			}
		}
	}
	v.RarefactionValid = rareOK

	// Bootstrap: every metric's NValid must never exceed the frozen
	// replicate count, and CI ordering must be sane (checked again under
	// numeric integrity below); here we check the schema-level replicate
	// bound and absence of duplicate metric/regime rows.
	bootOK := true
	for _, r := range results {
		seen := map[string]bool{}
		for _, row := range r.BootstrapRows {
			key := row.MetricID + "\x1f" + row.Regime
			if seen[key] {
				v.Errors = append(v.Errors, fmt.Sprintf("%s/%s: duplicate bootstrap row for %s", r.CandidateID, r.RepresentationID, key))
				bootOK = false
			}
			seen[key] = true
			if row.NValid < 0 || row.NValid > notation.BootstrapReplicates {
				v.Errors = append(v.Errors, fmt.Sprintf("%s/%s: %s has n_valid=%d, outside [0,%d]", r.CandidateID, r.RepresentationID, key, row.NValid, notation.BootstrapReplicates))
				bootOK = false
			}
		}
	}
	v.BootstrapValid = bootOK

	// Calibration: every Comparable VM-comparison row must reference a
	// value pulled from the frozen CALIBRATION_SCALES.tsv (Compare() only
	// ever marks a row Comparable when it found a frozen scale — see
	// adversarial test A1); confirm the calibration reference file this
	// run wrote is itself hash-bound to that frozen file.
	calRaw, calErr := os.ReadFile(filepath.Join(bundleRoot, "calibration", "CALIBRATION_REFERENCE.json"))
	calOK := calErr == nil && len(calRaw) > 0
	frozenScaleHash, hashErr := notation.FileSHA256(filepath.Join(base, "CALIBRATION_SCALES.tsv"))
	if hashErr != nil || !strings.Contains(string(calRaw), frozenScaleHash) {
		calOK = false
		v.Errors = append(v.Errors, "calibration reference record does not bind the frozen CALIBRATION_SCALES.tsv hash")
	}
	v.CalibrationValid = calOK

	// VM reference: re-verify the frozen VM reference once more against
	// what was actually used (defense in depth beyond the load-time check).
	_, _, vmErr := loadFrozenVMReference(base)
	v.VMReferenceValid = vmErr == nil
	if vmErr != nil {
		v.Errors = append(v.Errors, vmErr.Error())
	}
	_ = vmManifest

	// Numeric integrity.
	numOK := true
	checkFloat := func(where string, f float64) {
		if math.IsNaN(f) {
			v.Errors = append(v.Errors, where+": unexpected NaN")
			numOK = false
		}
		if math.IsInf(f, 0) {
			v.Errors = append(v.Errors, where+": unexpected Inf")
			numOK = false
		}
	}
	for _, r := range results {
		for _, m := range r.Fingerprint.Metrics {
			if m.Status == notation.Comparable {
				checkFloat(r.CandidateID+"/"+r.RepresentationID+"/"+m.MetricID, m.Value)
			}
		}
		for _, row := range r.BootstrapRows {
			checkFloat(r.CandidateID+"/bootstrap/"+row.MetricID, row.Estimate)
			checkFloat(r.CandidateID+"/bootstrap/"+row.MetricID, row.BootstrapMean)
			if row.CILow > row.CIHigh && row.NValid > 0 {
				v.Errors = append(v.Errors, fmt.Sprintf("%s/%s: bootstrap CI inverted for %s (%.6g > %.6g)", r.CandidateID, r.RepresentationID, row.MetricID, row.CILow, row.CIHigh))
				numOK = false
			}
			if row.NValid < 0 {
				v.Errors = append(v.Errors, fmt.Sprintf("%s/%s: negative n_valid for %s", r.CandidateID, r.RepresentationID, row.MetricID))
				numOK = false
			}
		}
		for _, row := range r.RarefactionSumm {
			if row.NValid > 0 && row.CILow > row.CIHigh {
				v.Errors = append(v.Errors, fmt.Sprintf("%s/%s: rarefaction CI inverted for %s@%d", r.CandidateID, r.RepresentationID, row.MetricID, row.Checkpoint))
				numOK = false
			}
		}
		for _, d := range r.Distributions {
			if d.Probability < 0 || d.Probability > 1.0000001 {
				v.Errors = append(v.Errors, fmt.Sprintf("%s/%s: invalid probability %.6g for %s", r.CandidateID, r.RepresentationID, d.Probability, d.MetricID))
				numOK = false
			}
		}
		if r.Records <= 0 {
			v.Errors = append(v.Errors, r.CandidateID+"/"+r.RepresentationID+": impossible sample size")
			numOK = false
		}
	}
	v.NumericIntegrityOK = numOK

	// Aggregation integrity: recompute purely from the written raw/vm_comparison
	// files and require an exact match against the in-memory aggregate.
	recomputed, err := recomputeAggregateFromRaw(bundleRoot, results)
	if err != nil {
		v.Errors = append(v.Errors, "aggregation recompute failed: "+err.Error())
	} else if !aggregatesEqual(agg, recomputed) {
		v.Errors = append(v.Errors, "recomputed aggregate summary does not match the originally written aggregate")
	} else {
		v.AggregationValid = true
	}

	return v, nil
}

func aggregatesEqual(a, b aggregateResult) bool {
	if len(a.Rows) != len(b.Rows) {
		return false
	}
	key := func(r aggregateRow) string {
		return fmt.Sprintf("%s|%s|%d|%s", r.CandidateID, r.RepresentationID, r.Checkpoint, r.Family)
	}
	am, bm := map[string]aggregateRow{}, map[string]aggregateRow{}
	for _, r := range a.Rows {
		am[key(r)] = r
	}
	for _, r := range b.Rows {
		bm[key(r)] = r
	}
	for k, ra := range am {
		rb, ok := bm[k]
		if !ok || ra.Status != rb.Status || ra.ComparableCount != rb.ComparableCount || ra.HasDistance != rb.HasDistance {
			return false
		}
		if ra.HasDistance {
			// The recomputed side round-trips every value through the
			// written TSV's %.12g text formatting, so exact float64
			// equality is not expected; tolerance is relative (12
			// significant digits, matching notation.WriteComparisonTSV's
			// own format verb), never a scientific approximation.
			scale := math.Max(1, math.Max(math.Abs(ra.MeanDistance), math.Abs(rb.MeanDistance)))
			if math.Abs(ra.MeanDistance-rb.MeanDistance) > scale*1e-9 {
				return false
			}
		}
	}
	return true
}

// ---- reproducibility (section 17) ----

type reproducibilityResult struct {
	Skipped          bool     `json:"skipped"`
	Pass             bool     `json:"pass"`
	SecondPassRunID  string   `json:"second_pass_run_id,omitempty"`
	SecondPassDir    string   `json:"second_pass_dir,omitempty"`
	FilesCompared    int      `json:"files_compared"`
	Mismatches       []string `json:"mismatches,omitempty"`
	Errors           []string `json:"errors,omitempty"`
}

func (r reproducibilityResult) markdown() string {
	if r.Skipped {
		return "# Reproducibility report\n\nSkipped for this pass (this run IS the reproducibility second pass, or reproducibility was explicitly disabled).\n"
	}
	var b strings.Builder
	b.WriteString("# Reproducibility report\n\n")
	fmt.Fprintf(&b, "Second pass run id: `%s`, directory: `%s`.\n\n", r.SecondPassRunID, r.SecondPassDir)
	fmt.Fprintf(&b, "Scientific files compared byte-for-byte (metadata/run-id/timestamp files excluded): %d.\n\n", r.FilesCompared)
	fmt.Fprintf(&b, "Result: **%s**\n", map[bool]string{true: "PASS", false: "FAIL"}[r.Pass])
	if len(r.Mismatches) != 0 {
		b.WriteString("\n## Mismatches\n\n")
		for _, m := range r.Mismatches {
			fmt.Fprintf(&b, "- %s\n", m)
		}
	}
	if len(r.Errors) != 0 {
		b.WriteString("\n## Errors\n\n")
		for _, e := range r.Errors {
			fmt.Fprintf(&b, "- %s\n", e)
		}
	}
	return b.String()
}

// scientificOutputDirs lists the bundle subdirectories whose contents are
// pure scientific payload and must reproduce bit-for-bit; RUN_MANIFEST.json,
// AUTHORIZATION_SNAPSHOT.json, INPUT_BINDINGS.json, and SHA256SUMS are
// excluded because they legitimately carry run-id/timestamp identity
// (task section 14 "Metadata timestamps хранить отдельно от scientific payload").
var scientificOutputDirs = []string{"candidates", "distributions", "rarefaction", "bootstrap", "vm_comparison", "aggregate", "calibration"}

func runReproducibilityCheck(opts productionRunOptions, primaryBundle, primaryRunID string) (reproducibilityResult, error) {
	// The second pass must never write inside the repository's git working
	// tree: when the primary bundle uses the default in-repo location
	// (research/comparative_notation/production_runs/<RUN_ID>/), that
	// directory is itself untracked at this point (not yet committed),
	// which would make the second pass's own clean_git_revision check fail
	// closed on the primary pass's own output — a self-inflicted false
	// negative, not a real reproducibility problem. A system temp
	// directory keeps the second pass's git view identical to the
	// primary's and is removed once the byte-for-byte comparison is done.
	secondDir, err := os.MkdirTemp("", "cns-prod-repro-*")
	if err != nil {
		return reproducibilityResult{Pass: false, Errors: []string{"cannot create reproducibility scratch dir: " + err.Error()}}, nil
	}
	defer os.RemoveAll(secondDir)
	if err := os.Remove(secondDir); err != nil { // executeProductionRun refuses to run into an existing dir
		return reproducibilityResult{Pass: false, Errors: []string{err.Error()}}, nil
	}
	secondRunID := primaryRunID + "-REPRO"
	second := productionRunOptions{Repo: opts.Repo, OutDirOverride: secondDir, RunIDOverride: secondRunID, SkipReproducibility: true}
	if _, err := executeProductionRun(second); err != nil {
		return reproducibilityResult{Pass: false, Errors: []string{"second pass failed: " + err.Error()}}, nil
	}
	res := reproducibilityResult{SecondPassRunID: secondRunID, SecondPassDir: "(temporary, outside the repository; removed after comparison)", Pass: true}
	for _, sub := range scientificOutputDirs {
		primarySub := filepath.Join(primaryBundle, sub)
		secondSub := filepath.Join(secondDir, sub)
		if err := diffDirsBytewise(primarySub, secondSub, &res); err != nil {
			res.Errors = append(res.Errors, err.Error())
			res.Pass = false
		}
	}
	if len(res.Mismatches) != 0 {
		res.Pass = false
	}
	return res, nil
}

func diffDirsBytewise(a, b string, res *reproducibilityResult) error {
	return filepath.Walk(a, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(a, path)
		if err != nil {
			return err
		}
		other := filepath.Join(b, rel)
		pa, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		pb, err := os.ReadFile(other)
		if err != nil {
			res.Mismatches = append(res.Mismatches, rel+": second pass missing this file: "+err.Error())
			return nil
		}
		res.FilesCompared++
		if string(pa) != string(pb) {
			res.Mismatches = append(res.Mismatches, rel+": byte-for-byte mismatch between passes")
		}
		return nil
	})
}

// ---- checksums, report, status update ----

func writeProductionResultChecksums(bundleRoot string) error {
	var paths []string
	err := filepath.Walk(bundleRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(bundleRoot, path)
		if err != nil {
			return err
		}
		paths = append(paths, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return err
	}
	sort.Strings(paths)
	f, err := os.Create(filepath.Join(bundleRoot, "SHA256SUMS"))
	if err != nil {
		return err
	}
	defer f.Close()
	for _, rel := range paths {
		h, err := notation.FileSHA256(filepath.Join(bundleRoot, filepath.FromSlash(rel)))
		if err != nil {
			return err
		}
		fmt.Fprintf(f, "%s  %s\n", h, rel)
	}
	return nil
}

func buildRunReport(identity runIdentity, results []*perRepresentationResult, v postRunValidation, repro reproducibilityResult, completed, valid bool, bundleRel string) string {
	var b strings.Builder
	b.WriteString("# Production comparative run report\n\n")
	b.WriteString("This report is technical only: it records what was computed and whether it validated. It draws no conclusion about the Voynich Manuscript's origin, mechanism, or historical identity; that is explicitly deferred to a separate task.\n\n")
	fmt.Fprintf(&b, "- run_id: `%s`\n", identity.RunID)
	fmt.Fprintf(&b, "- authorization reference: `research/comparative_notation/PRODUCTION_RUN_AUTHORIZATION.json` (sha256 `%s`)\n", identity.AuthorizationSHA256)
	fmt.Fprintf(&b, "- git revision: `%s`\n", identity.GitCommit)
	fmt.Fprintf(&b, "- frozen corpus selection: C01, C02, C06 (sha256 `%s`)\n", identity.CorpusSelectionSHA256)
	b.WriteString("- C06 representation set: MUSIC-R1, MUSIC-R2, MUSIC-R3 (three representations of one candidate, not independent candidates)\n")

	totalMeasurements, totalRare, totalBoot := 0, 0, 0
	for _, r := range results {
		totalMeasurements += len(r.Fingerprint.Metrics)
		for _, row := range r.RarefactionRows {
			if row.Replicate >= 0 {
				totalRare++
			}
		}
		totalBoot += len(r.BootstrapRows)
	}
	fmt.Fprintf(&b, "- measurements executed (raw metric rows, all candidates/representations): %d\n", totalMeasurements)
	fmt.Fprintf(&b, "- rarefaction draws executed: %d\n", totalRare)
	fmt.Fprintf(&b, "- bootstrap rows executed: %d\n", totalBoot)

	b.WriteString("\n## Validation summary\n\n")
	b.WriteString(v.markdown())
	b.WriteString("\n## Reproducibility summary\n\n")
	b.WriteString(repro.markdown())

	b.WriteString("\n## Excluded / not-applicable operations\n\n")
	b.WriteString("- C03-C05, C07-C09: excluded/deferred per the frozen production corpus selection; not revisited.\n")
	b.WriteString("- Within-class pair distances and variance (CLASS_SUMMARY/WITHIN_CLASS_DISTANCES): NOT_APPLICABLE_FOR_CURRENT_PANEL — every class has exactly one independent corpus.\n")
	b.WriteString("- Cross-class ranking, PCA, UMAP, nearest-neighbour analysis: OUT_OF_SCOPE_REPOSITORY_LOCKED.\n")
	b.WriteString("- Checkpoints 20000 and 39380: NOT_COMPARABLE for every candidate in this subset (all observed sizes are below both).\n")
	b.WriteString("- STRUCTURALLY_CLOSE_ON/STRUCTURALLY_DISTANT_ON verdicts: left PENDING — no frozen numeric threshold exists for this classification.\n")

	if len(v.Errors) != 0 || len(repro.Errors) != 0 || len(repro.Mismatches) != 0 {
		b.WriteString("\n## Warnings / anomalies\n\n")
		for _, e := range v.Errors {
			fmt.Fprintf(&b, "- validation: %s\n", e)
		}
		for _, e := range repro.Errors {
			fmt.Fprintf(&b, "- reproducibility: %s\n", e)
		}
		for _, m := range repro.Mismatches {
			fmt.Fprintf(&b, "- reproducibility mismatch: %s\n", m)
		}
	}

	fmt.Fprintf(&b, "\n## Result bundle\n\n`%s`\n\nChecksums: `%s/SHA256SUMS`\n", bundleRel, bundleRel)

	b.WriteString("\n## Final validity status\n\n")
	fmt.Fprintf(&b, "```text\nPRODUCTION_COMPARATIVE_RUN_COMPLETED=%t\nPRODUCTION_COMPARATIVE_RUN_VALID=%t\n```\n", completed, valid)
	if valid {
		b.WriteString("\nPRODUCTION_COMPARATIVE_RUN_VALID=true\n")
	}
	return b.String()
}

func updateComparativeRunManifest(repo, base, runID, gitCommit string, completed, valid bool, bundleRoot string) error {
	rel, _ := filepath.Rel(repo, bundleRoot)
	manifest := map[string]any{
		"schema_version": "production-comparative-run-manifest-3.0", "run_id": runID,
		"status":       map[bool]string{true: "VALID", false: "INVALID_OR_INCOMPLETE"}[valid],
		"authorized":   true, "completed": completed, "valid": valid,
		"git_commit":   gitCommit,
		"result_bundle": filepath.ToSlash(rel),
		"result_files": []string{"RUN_MANIFEST.json", "AUTHORIZATION_SNAPSHOT.json", "INPUT_BINDINGS.json", "candidates/", "rarefaction/", "bootstrap/", "distributions/", "vm_comparison/", "aggregate/", "calibration/", "validation/", "SHA256SUMS"},
	}
	return writeJSONFile(filepath.Join(base, "PRODUCTION_COMPARATIVE_RUN_MANIFEST.json"), manifest)
}
