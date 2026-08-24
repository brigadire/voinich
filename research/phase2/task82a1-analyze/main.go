// Command task82a1-analyze extends F2 coverage from frozen Task82a raw
// documents. It never calls a mnemonic mechanism or reads target data.
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"zcore.dev/voinich/internal/fingerprintv2"
	"zcore.dev/voinich/internal/task82a"
)

const (
	outRel = "research/phase2/task82a1"
)

type metricSpec struct {
	ID, Class, Family, Structures, Status, Reason string
}

// Every row is taken from the frozen registry. The only computation added
// here is the generic ordered-group projection explicitly labelled
// ASSEMBLER_APPLICABLE; it never claims physical-line semantics.
var metrics = []metricSpec{
	{"2DL1_LAYOUT_POSITION_MI", "CORE", "2D-LITE", "TOKEN_SEQUENCE,ASSEMBLER_LINES", "ASSEMBLER_APPLICABLE", "Frozen order/category statistic is computable on assembler-defined lines only."},
	{"BP1_BOUNDARY_TOKEN_NMI", "CORE", "boundary", "TOKEN_SEQUENCE,TOKEN_BOUNDARIES,ASSEMBLER_LINES", "ASSEMBLER_APPLICABLE", "Frozen boundary statistic is computable on assembler-defined lines only."},
	{"EF1_GIANT_COMPONENT_SHARE", "CORE", "edit family", "GLYPH_SEQUENCE,TOKEN_SEQUENCE", "DIRECTLY_APPLICABLE", "Vocabulary edit graph requires no layout or manuscript metadata."},
	{"EF1_ISOLATE_SHARE", "SUPPORTING", "edit family", "GLYPH_SEQUENCE,TOKEN_SEQUENCE", "DIRECTLY_APPLICABLE", "Vocabulary edit graph requires no layout or manuscript metadata."},
	{"EF2_GLOBAL_CLUSTERING", "CORE", "edit family", "GLYPH_SEQUENCE,TOKEN_SEQUENCE", "DIRECTLY_APPLICABLE", "Vocabulary edit graph requires no layout or manuscript metadata."},
	{"EF3_DEGREE_FREQUENCY_SPEARMAN", "CORE", "edit family", "GLYPH_SEQUENCE,TOKEN_SEQUENCE", "DIRECTLY_APPLICABLE", "Vocabulary edit graph requires no layout or manuscript metadata."},
	{"HR1_FOLIO_VARIANCE_SHARE", "CORE", "hierarchy", "FOLIOS,PHYSICAL_LINES", "NOT_APPLICABLE_STRUCTURE", "No folio or physical-line hierarchy exists."},
	{"HR1_LOCUS_VARIANCE_SHARE", "SUPPORTING", "hierarchy", "LOCUS_COORDINATES", "NOT_APPLICABLE_METADATA", "No locus metadata exists."},
	{"HR1_SECTION_VARIANCE_SHARE", "CORE", "hierarchy", "SECTIONS", "NOT_APPLICABLE_METADATA", "No section metadata exists."},
	{"HR6_CURRIER_SECTION_NMI", "SUPPORTING", "hierarchy", "CURRIER_OR_EQUIVALENT_METADATA,SECTIONS", "NOT_APPLICABLE_METADATA", "No Currier-equivalent or section metadata exists."},
	{"LC1_LOCUS_TYPE_NMI", "CORE", "locus", "LOCUS_COORDINATES", "NOT_APPLICABLE_METADATA", "No documented locus type exists."},
	{"LC2_LABEL_TEXT_NMI", "CORE", "locus", "LOCUS_COORDINATES,FOLIOS", "NOT_APPLICABLE_METADATA", "No locus label/text or folio metadata exists."},
	{"LC5_IVTFF_I_NMI", "SUPPORTING", "locus", "CURRIER_OR_EQUIVALENT_METADATA", "NOT_APPLICABLE_METADATA", "IVTFF variable metadata is unavailable."},
	{"LC5_IVTFF_X_NMI", "SUPPORTING", "locus", "CURRIER_OR_EQUIVALENT_METADATA", "NOT_APPLICABLE_METADATA", "IVTFF variable metadata is unavailable."},
	{"LP1_RULE_SUPPORT_GINI", "SUPPORTING", "lexical paradigm", "GLYPH_SEQUENCE,TOKEN_SEQUENCE", "DIRECTLY_APPLICABLE", "Rule support is computed from the observable vocabulary."},
	{"LP4_PREFIX_ATTACHMENT_NMI", "SUPPORTING", "lexical paradigm", "GLYPH_SEQUENCE,TOKEN_SEQUENCE", "DIRECTLY_APPLICABLE", "Attachment statistic is computed from observable tokens."},
	{"LP4_SUFFIX_ATTACHMENT_NMI", "SUPPORTING", "lexical paradigm", "GLYPH_SEQUENCE,TOKEN_SEQUENCE", "DIRECTLY_APPLICABLE", "Attachment statistic is computed from observable tokens."},
	{"LS1_LINE_LENGTH_CV", "SUPPORTING", "line", "ASSEMBLER_LINES", "ASSEMBLER_APPLICABLE", "Token counts are defined only for assembler lines."},
	{"LS2_POSITIONAL_LEXICON_NMI", "CORE", "line", "TOKEN_SEQUENCE,ASSEMBLER_LINES", "ASSEMBLER_APPLICABLE", "Position classes are defined only for assembler lines."},
	{"LS3_BOUNDARY_LENGTH_ASYMMETRY", "CORE", "line", "TOKEN_BOUNDARIES,ASSEMBLER_LINES", "ASSEMBLER_APPLICABLE", "Boundary comparison is defined only for assembler lines."},
	{"LS4_WITHIN_LINE_EXACT_REPETITION", "SUPPORTING", "line", "TOKEN_SEQUENCE,ASSEMBLER_LINES", "ASSEMBLER_APPLICABLE", "Within-line recurrence is defined only for assembler lines."},
	{"PF2_FOLIO_COHERENCE", "CORE", "folio", "FOLIOS,PHYSICAL_LINES", "NOT_APPLICABLE_STRUCTURE", "No folios exist."},
	{"PF3_ADJACENT_FOLIO_CONTINUITY", "SUPPORTING", "folio", "FOLIOS", "NOT_APPLICABLE_STRUCTURE", "No ordered folios exist."},
	{"PF4_RECTO_VERSO_COHERENCE", "SUPPORTING", "folio", "RECTO_VERSO", "NOT_APPLICABLE_STRUCTURE", "No recto/verso pairs exist."},
	{"PF5_WITHIN_FOLIO_PROGRESSION", "CORE", "folio", "FOLIOS,PHYSICAL_LINES", "NOT_APPLICABLE_STRUCTURE", "No folios exist."},
	{"cs1/family-line-position", "SUPPORTING", "cross-scale", "ASSEMBLER_LINES,TOKEN_SEQUENCE", "ASSEMBLER_APPLICABLE", "Line positions are assembler-defined."},
	{"cs2/prev-family-current-family", "SUPPORTING", "cross-scale", "TOKEN_SEQUENCE", "DIRECTLY_APPLICABLE", "Token adjacency needs no manuscript structure."},
	{"cs3/family-locus-type", "SUPPORTING", "cross-scale", "LOCUS_COORDINATES", "NOT_APPLICABLE_METADATA", "No locus metadata exists."},
	{"cs4/family-currier", "SUPPORTING", "cross-scale", "CURRIER_OR_EQUIVALENT_METADATA", "NOT_APPLICABLE_METADATA", "No Currier-equivalent metadata exists."},
	{"cs4/family-section", "SUPPORTING", "cross-scale", "SECTIONS", "NOT_APPLICABLE_METADATA", "No section metadata exists."},
	{"cs5/local-adjacency-x-regime", "SUPPORTING", "cross-scale", "CURRIER_OR_EQUIVALENT_METADATA,SECTIONS", "NOT_APPLICABLE_METADATA", "No regime metadata exists."},
	{"cs6/family-diversity-x-line-length", "SUPPORTING", "cross-scale", "ASSEMBLER_LINES,TOKEN_SEQUENCE", "ASSEMBLER_APPLICABLE", "Line lengths are assembler-defined."},
	{"cs7/edit-distance-x-structural-distance", "SUPPORTING", "cross-scale", "GLYPH_SEQUENCE,ASSEMBLER_LINES", "ASSEMBLER_APPLICABLE", "Structural distances are assembler-line distances."},
}

type rawMetric struct {
	MetricID, MetricVersion, Family, Classification, Verdict, MissingReason, BoundaryProvenance string
	Value                                                                                       float64
	Available                                                                                   bool
}

type vector struct {
	JobID, MechanismID, Policy, Corpus, Scale string
	Replicate                                 int
	Metrics                                   []rawMetric
	Diagnostics                               groupDiagnostics
}

type groupDiagnostics struct {
	NonEmptyGroups, MultiTokenGroups, TokenCount, VocabularySize, WithinGroupTransitions int
}

type stabilityObservation struct {
	mech, policy, corpus, scale, metric string
	rep                                 int
	value                               float64
}

func main() {
	root := flag.String("root", ".", "repository root")
	verifyOnly := flag.Bool("verify-only", false, "verify frozen inputs and Task82a.1 outputs without rewriting them")
	flag.Parse()
	if *verifyOnly {
		if err := verifyInputs(*root); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := verifyFrozenOutputs(*root); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if err := run(*root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(root string) error {
	if err := verifyInputs(root); err != nil {
		return err
	}
	out := filepath.Join(root, outRel)
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	if err := writeDesignAndPlan(out); err != nil {
		return err
	}
	if err := writeExtensionManifest(out, root); err != nil {
		return err
	}
	vecs, err := loadVectors(root)
	if err != nil {
		return err
	}
	if err := writeVectors(out, vecs); err != nil {
		return err
	}
	if err := writeTabularOutputs(out, vecs); err != nil {
		return err
	}
	if err := writeNarrativeOutputs(out, vecs); err != nil {
		return err
	}
	return freeze(out, root, vecs)
}

func verifyInputs(root string) error {
	marker, err := os.ReadFile(filepath.Join(root, "research/phase2/task82a/TASK82A_CORPUS_SCALE_PORTFOLIO_FROZEN"))
	if err != nil || !strings.Contains(string(marker), "TASK82A_CORPUS_SCALE_PORTFOLIO_FROZEN") {
		return fmt.Errorf("Task82a portfolio is not frozen")
	}
	f2marker, err := os.ReadFile(filepath.Join(root, "research/phase2/fingerprint/FINGERPRINT_V2_FROZEN"))
	if err != nil || !strings.Contains(string(f2marker), "FINGERPRINT_V2_FROZEN") {
		return fmt.Errorf("Fingerprint V2 is not frozen")
	}
	var rm struct {
		CompletedJobCount       int               `json:"completed_job_count"`
		OutputArtifactChecksums map[string]string `json:"output_artifact_checksums"`
	}
	b, err := os.ReadFile(filepath.Join(root, "research/phase2/task82a/TASK82A_RESULTS_MANIFEST.json"))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &rm); err != nil {
		return err
	}
	if rm.CompletedJobCount != 468 {
		return fmt.Errorf("Task82a manifest has %d completed jobs, want 468", rm.CompletedJobCount)
	}
	base := filepath.Join(root, "research/phase2/task82a")
	for name, want := range rm.OutputArtifactChecksums {
		if got, err := hashFile(filepath.Join(base, name)); err != nil || got != want {
			return fmt.Errorf("Task82a frozen output checksum mismatch: %s", name)
		}
	}
	// Verify the Fingerprint V2 binding through the checksum recorded by the
	// frozen Task82a results manifest. This reads only freeze metadata, never
	// the target corpus or target metric values.
	var fullRM struct {
		F2FreezeManifestChecksum string `json:"f2_freeze_manifest_checksum"`
	}
	if err := json.Unmarshal(b, &fullRM); err != nil {
		return err
	}
	f2Manifest := filepath.Join(root, "research/phase2/fingerprint/FINGERPRINT_V2_FREEZE_MANIFEST.json")
	if got, err := hashFile(f2Manifest); err != nil || got != fullRM.F2FreezeManifestChecksum {
		return fmt.Errorf("Fingerprint V2 freeze manifest checksum mismatch")
	}
	registryPath := filepath.Join(root, "research/phase2/fingerprint/F2_METRIC_REGISTRY_FINAL.tsv")
	registryIDs, err := registryMetricIDs(registryPath)
	if err != nil {
		return err
	}
	if len(registryIDs) != len(metrics) {
		return fmt.Errorf("frozen registry has %d metrics, audit has %d", len(registryIDs), len(metrics))
	}
	for _, m := range metrics {
		if !registryIDs[m.ID] {
			return fmt.Errorf("audit metric %s is absent from frozen registry", m.ID)
		}
	}
	return nil
}

func registryMetricIDs(path string) (map[string]bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string]bool{}
	s := bufio.NewScanner(f)
	first := true
	for s.Scan() {
		if first {
			first = false
			continue
		}
		id := strings.SplitN(s.Text(), "\t", 2)[0]
		if id != "" {
			out[id] = true
		}
	}
	return out, s.Err()
}

func verifyFrozenOutputs(root string) error {
	out := filepath.Join(root, outRel)
	b, err := os.ReadFile(filepath.Join(out, "TASK82A1_RESULTS_MANIFEST.json"))
	if err != nil {
		return err
	}
	var m struct {
		CompletedJobCount       int               `json:"completed_job_count"`
		OutputArtifactChecksums map[string]string `json:"output_artifact_checksums"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	if m.CompletedJobCount != 468 {
		return fmt.Errorf("Task82a.1 manifest has %d jobs, want 468", m.CompletedJobCount)
	}
	for name, want := range m.OutputArtifactChecksums {
		if got, err := hashFile(filepath.Join(out, name)); err != nil || got != want {
			return fmt.Errorf("Task82a.1 frozen output checksum mismatch: %s", name)
		}
	}
	marker := string(mustRead(filepath.Join(out, "TASK82A_F2_COVERAGE_EXTENDED_FROZEN")))
	if !strings.Contains(marker, "results_manifest_sha256="+mustHash(filepath.Join(out, "TASK82A1_RESULTS_MANIFEST.json"))) {
		return fmt.Errorf("Task82a.1 freeze marker does not bind the results manifest")
	}
	return nil
}

func writeDesignAndPlan(out string) error {
	var audit strings.Builder
	audit.WriteString("metric_id\tfamily\tclassification\trequired_structures\tstructures_available\testimator_availability\tcomputational_requirements\ttask82a_original_status\tapplicability_status\treason\n")
	for _, m := range metrics {
		available := "GLYPH_SEQUENCE,TOKEN_SEQUENCE,TOKEN_BOUNDARIES,LOCAL_MECHANISM_BOUNDARIES,ASSEMBLER_LINES"
		if m.Status == "NOT_APPLICABLE_METADATA" || m.Status == "NOT_APPLICABLE_STRUCTURE" {
			available = "GLYPH_SEQUENCE,TOKEN_SEQUENCE,TOKEN_BOUNDARIES,LOCAL_MECHANISM_BOUNDARIES,ASSEMBLER_LINES; missing required manuscript structure"
		}
		estimator, cost, old := "FROZEN_EXISTING", "IMPORTED_FROM_TASK82A", "COMPUTED_OR_ATTEMPTED"
		if isNewGenericMetric(m.ID) {
			estimator, cost, old = "GENERIC_EXTRACTION_EQUAL_TO_TASK79_V1", "O(tokens); zero repetitions; full 468-job portfolio", "PREVIOUSLY_COST_BOUNDED"
		} else if m.Status == "ASSEMBLER_APPLICABLE" {
			estimator, cost = "FROZEN_TASK82A_EXISTING", "IMPORTED_FROM_TASK82A"
		}
		if m.Status == "NOT_APPLICABLE_METADATA" || m.Status == "NOT_APPLICABLE_STRUCTURE" {
			old = m.Status
		}
		audit.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", m.ID, m.Family, m.Class, m.Structures, available, estimator, cost, old, m.Status, m.Reason))
	}
	if err := write(out, "F2_APPLICABILITY_AUDIT.tsv", audit.String()); err != nil {
		return err
	}
	design := "# Task82a.1 frozen F2 coverage extension\n\nThis post-generation measurement extension consumes only immutable Task82a raw observable documents and frozen F2 definitions. It creates no pages, folios, sections, hands, loci, recto/verso pairs, or Currier-like metadata. Metrics computed on `ASSEMBLER_LINE` groups remain segregated from direct metrics and are never called physical-line measurements.\n\n## Frozen computation plan\n\nThe plan is written before extended-vector generation. Existing Task82a EF/LP/cross-scale rows are imported verbatim. Six observed Task79 statistics are extracted through a generic ordered-group API proven bit-equivalent to the frozen implementation on regression fixtures. They need no permutation/bootstrap repetitions because this extension reports observed values, not new significance verdicts. In particular, `2DL1` preserves the frozen task79-v1 implementation's three-class boundary behavior; the registry prose/implementation mismatch is documented and is not corrected because doing so would create a new metric version.\n\n## Preregistered stability rules\n\nFor cross-corpus and cross-seed ranges: `STABLE <= 0.01`, `PARTIALLY_STABLE <= 0.10`, otherwise `UNSTABLE`; fewer than two available observations is `INCONCLUSIVE`. Scale convergence applies to paired MEDIUM/LARGE observations with the same thresholds and names `CONVERGED`, `PARTIALLY_CONVERGED`, `NOT_CONVERGED`. Cue-policy effects use the same absolute-range bands and never aggregate LOCAL with GLOBAL. These descriptive thresholds were fixed before regenerated extended values were written.\n"
	if err := write(out, "TASK82A1_DESIGN.md", design); err != nil {
		return err
	}
	if err := write(out, "TASK82A1_DESIGN_FROZEN", "TASK82A1_DESIGN_FROZEN\nversion=V1.0\n"); err != nil {
		return err
	}
	var plan strings.Builder
	plan.WriteString("metric_id\tfamily\tclassification\taction\treason\n")
	for _, m := range metrics {
		action := "DO_NOT_RUN_NOT_APPLICABLE"
		reason := m.Reason
		if m.Status == "DIRECTLY_APPLICABLE" {
			action, reason = "IMPORT_EXISTING", "Already measured in frozen Task82a raw vector."
		}
		if m.Status == "ASSEMBLER_APPLICABLE" {
			action, reason = "IMPORT_EXISTING", "Already attempted by frozen Task82a with assembler provenance."
		}
		if isNewGenericMetric(m.ID) {
			action, reason = "RUN", "Generic task79-v1-equivalent observed statistic on immutable assembler-line corpus."
		}
		plan.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\n", m.ID, m.Family, m.Class, action, reason))
	}
	if err := write(out, "F2_EXTENSION_PLAN.tsv", plan.String()); err != nil {
		return err
	}
	if err := write(out, "F2_EXTENSION_PLAN_FROZEN", "F2_EXTENSION_PLAN_FROZEN\nscope=all 468 immutable Task82a jobs\n"); err != nil {
		return err
	}
	return write(out, "F2_EXTENSION_COST_MODEL.tsv", "family\testimator\trepresentative_input_size\twall_time\tcpu_time\tmemory\trepetitions_bootstrap_permutations\tprojected_full_cost\n2D-LITE\tOrderedGroupMetrics/2DL1\t256 lines x 4 tokens\t0.335 ms (three-run median pilot)\t0.335 ms\t285025 B shared six-metric allocation\t0\t<0.16 CPU-second shared full portfolio\nboundary\tOrderedGroupMetrics/BP1\t256 lines x 4 tokens\t0.335 ms (shared pilot)\t0.335 ms\t285025 B shared six-metric allocation\t0\t<0.16 CPU-second shared full portfolio\nline\tOrderedGroupMetrics/LS1-LS4\t256 lines x 4 tokens\t0.335 ms (shared pilot)\t0.335 ms\t285025 B shared six-metric allocation\t0\t<0.16 CPU-second shared full portfolio\n")
}

func isNewGenericMetric(id string) bool {
	switch id {
	case "2DL1_LAYOUT_POSITION_MI", "BP1_BOUNDARY_TOKEN_NMI", "LS1_LINE_LENGTH_CV", "LS2_POSITIONAL_LEXICON_NMI", "LS3_BOUNDARY_LENGTH_ASYMMETRY", "LS4_WITHIN_LINE_EXACT_REPETITION":
		return true
	default:
		return false
	}
}

func writeExtensionManifest(out, root string) error {
	rawDir := filepath.Join(root, "research/phase2/task82a/raw")
	names, err := filepath.Glob(filepath.Join(rawDir, "*.json"))
	if err != nil {
		return err
	}
	sort.Strings(names)
	jobs := make([]map[string]any, 0, len(names)*6)
	for _, name := range names {
		b, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		var a task82a.Artifact
		if err := json.Unmarshal(b, &a); err != nil {
			return err
		}
		for _, m := range metrics {
			if !isNewGenericMetric(m.ID) {
				continue
			}
			jobs = append(jobs, map[string]any{
				"task82a_job_id": a.Job.JobID, "mechanism_id": a.Job.MechanismID,
				"scaling_policy_id": a.Job.ScalingPolicyID, "corpus_id": a.Job.InputCorpusID,
				"scale": a.Job.CorpusScale, "replicate": a.Job.Replicate,
				"metric_family": m.Family, "metric_id": m.ID, "estimator_version": "task79-v1-generic-assembler-v1",
			})
		}
	}
	b, err := json.MarshalIndent(map[string]any{
		"schema": "task82a1-extension-manifest-v1", "input": "TASK82A_CORPUS_SCALE_PORTFOLIO_FROZEN",
		"jobs": jobs, "run_count": len(jobs), "expected_task82a_jobs": len(names),
		"metric_count_per_job": 6, "generation_prohibited": true,
	}, "", "  ")
	if err != nil {
		return err
	}
	return write(out, "TASK82A1_MANIFEST.json", string(b)+"\n")
}

func loadVectors(root string) ([]vector, error) {
	rawDir := filepath.Join(root, "research/phase2/task82a/raw")
	names, err := filepath.Glob(filepath.Join(rawDir, "*.json"))
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	if len(names) != 468 {
		return nil, fmt.Errorf("found %d raw jobs, want 468", len(names))
	}
	out := make([]vector, 0, len(names))
	for _, name := range names {
		b, err := os.ReadFile(name)
		if err != nil {
			return nil, err
		}
		var a task82a.Artifact
		if err := json.Unmarshal(b, &a); err != nil {
			return nil, err
		}
		if a.DocumentSHA256 != a.Document.Checksum() {
			return nil, fmt.Errorf("raw checksum mismatch for %s", a.Job.JobID)
		}
		if a.BoundaryProvenance.Token != "ASSEMBLER_DEFINED" || a.BoundaryProvenance.Line != "ASSEMBLER_DEFINED" || a.BoundaryProvenance.Page != "NOT_DEFINED" {
			return nil, fmt.Errorf("boundary provenance drift for %s", a.Job.JobID)
		}
		corpusPath := filepath.Join(root, a.F2.CorpusFilePath)
		allowedDir := filepath.Clean(filepath.Join(root, "research/phase2/task82a/raw/f2corpus")) + string(os.PathSeparator)
		if !strings.HasPrefix(filepath.Clean(corpusPath), allowedDir) {
			return nil, fmt.Errorf("input path escapes frozen Task82a f2 corpus directory: %s", a.F2.CorpusFilePath)
		}
		if got, err := hashFile(corpusPath); err != nil || got != a.F2.CorpusFileChecksum {
			return nil, fmt.Errorf("assembler corpus checksum mismatch for %s", a.Job.JobID)
		}
		ms := make([]rawMetric, 0, len(a.F2.Metrics)+6)
		for _, m := range a.F2.Metrics {
			ms = append(ms, rawMetric{m.MetricID, m.MetricVersion, m.Family, m.Classification, m.Verdict, m.MissingReason, m.BoundaryProvenance, m.Value, m.Available})
		}
		lines, err := readLines(corpusPath)
		if err != nil {
			return nil, err
		}
		for id, val := range genericMetrics(lines) {
			class, family := metricClassFamily(id)
			ms = append(ms, rawMetric{id, "task79-v1", family, class, "DESCRIPTIVE", "", "ASSEMBLER_LINE", val, true})
		}
		seen := map[string]bool{}
		for _, m := range ms {
			seen[m.MetricID] = true
		}
		for _, spec := range metrics {
			if !seen[spec.ID] {
				ms = append(ms, rawMetric{MetricID: spec.ID, MetricVersion: "task79-v1", Family: spec.Family, Classification: spec.Class, Verdict: spec.Status, MissingReason: spec.Reason, BoundaryProvenance: "NOT_AVAILABLE", Available: false})
			}
		}
		if len(ms) != len(metrics) {
			return nil, fmt.Errorf("job %s has %d extended metrics, want %d", a.Job.JobID, len(ms), len(metrics))
		}
		sort.Slice(ms, func(i, j int) bool { return ms[i].MetricID < ms[j].MetricID })
		out = append(out, vector{a.Job.JobID, a.Job.MechanismID, a.Job.ScalingPolicyID, a.Job.InputCorpusID, a.Job.CorpusScale, a.Job.Replicate, ms, diagnoseGroups(lines)})
	}
	return out, nil
}

func diagnoseGroups(lines [][]string) groupDiagnostics {
	var d groupDiagnostics
	vocab := map[string]bool{}
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		d.NonEmptyGroups++
		if len(line) >= 2 {
			d.MultiTokenGroups++
		}
		d.TokenCount += len(line)
		d.WithinGroupTransitions += len(line) - 1
		for _, tok := range line {
			vocab[tok] = true
		}
	}
	d.VocabularySize = len(vocab)
	return d
}

func readLines(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines [][]string
	s := bufio.NewScanner(f)
	for s.Scan() {
		lines = append(lines, strings.Fields(s.Text()))
	}
	return lines, s.Err()
}

func genericMetrics(lines [][]string) map[string]float64 {
	return fingerprintv2.OrderedGroupMetrics(lines)
}

func writeVectors(out string, vecs []vector) error {
	f, err := os.Create(filepath.Join(out, "F2_RAW_VECTORS_EXTENDED.jsonl"))
	if err != nil {
		return err
	}
	defer f.Close()
	e := json.NewEncoder(f)
	for _, v := range vecs {
		if err := e.Encode(v); err != nil {
			return err
		}
	}
	return nil
}

func writeTabularOutputs(out string, vecs []vector) error {
	var sample, coverage strings.Builder
	sample.WriteString("job_id\tmetric_id\testimator_unit\tcontributing_units\tvocabulary_support\tpair_count\tnull_repetitions\tstatus\treason\n")
	coverage.WriteString("job_id\tmechanism_id\tscaling_policy_id\tcorpus_id\tcorpus_scale\treplicate\tdirect_core_available\tassembler_core_available\ttotal_mathematically_available_core\tdirect_supporting_available\tassembler_supporting_available\n")
	for _, v := range vecs {
		dc, ac, ds, as := 0, 0, 0, 0
		for _, m := range v.Metrics {
			spec, ok := findMetric(m.MetricID)
			if !ok {
				continue
			}
			unit, n, pairs, nulls := diagnosticCounts(m.MetricID, v.Diagnostics)
			status, reason := "AVAILABLE", "NONE"
			if !m.Available {
				status, reason = "NOT_APPLICABLE", m.MissingReason
				if strings.Contains(m.MissingReason, "INCONCLUSIVE") || strings.Contains(m.MissingReason, "INSUFFICIENT") {
					status = "NOT_APPLICABLE_DATA_DEGENERACY"
				}
			}
			if isNewGenericMetric(m.MetricID) {
				nulls = 0
			}
			sample.WriteString(fmt.Sprintf("%s\t%s\t%s\t%d\t%d\t%d\t%d\t%s\t%s\n", v.JobID, m.MetricID, unit, n, v.Diagnostics.VocabularySize, pairs, nulls, status, strings.ReplaceAll(reason, "\t", " ")))
			if m.Available && spec.Class == "CORE" {
				if spec.Status == "DIRECTLY_APPLICABLE" {
					dc++
				} else if spec.Status == "ASSEMBLER_APPLICABLE" {
					ac++
				}
			}
			if m.Available && spec.Class == "SUPPORTING" {
				if spec.Status == "DIRECTLY_APPLICABLE" {
					ds++
				} else if spec.Status == "ASSEMBLER_APPLICABLE" {
					as++
				}
			}
		}
		coverage.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%d\t%d\t%d\t%d\t%d\t%d\n", v.JobID, v.MechanismID, v.Policy, v.Corpus, v.Scale, v.Replicate, dc, ac, dc+ac, ds, as))
	}
	if err := write(out, "F2_SAMPLE_SIZE_DIAGNOSTICS.tsv", sample.String()); err != nil {
		return err
	}
	if err := write(out, "F2_COVERAGE_EXTENDED.tsv", coverage.String()); err != nil {
		return err
	}
	return writeMeasuredStability(out, vecs)
}

func diagnosticCounts(id string, d groupDiagnostics) (unit string, contributing, pairs, nulls int) {
	nulls = 30 // imported Task82a null precision where the frozen metric uses a null
	switch id {
	case "EF1_GIANT_COMPONENT_SHARE", "EF1_ISOLATE_SHARE", "EF2_GLOBAL_CLUSTERING", "EF3_DEGREE_FREQUENCY_SPEARMAN", "LP1_RULE_SUPPORT_GINI", "LP4_PREFIX_ATTACHMENT_NMI", "LP4_SUFFIX_ATTACHMENT_NMI":
		return "VOCABULARY_TYPE", d.VocabularySize, d.VocabularySize * (d.VocabularySize - 1) / 2, nulls
	case "LS1_LINE_LENGTH_CV":
		return "ASSEMBLER_LINE", d.NonEmptyGroups, 0, 0
	case "LS3_BOUNDARY_LENGTH_ASYMMETRY":
		return "ASSEMBLER_LINE_WITH_2PLUS_TOKENS", d.MultiTokenGroups, d.MultiTokenGroups, 0
	case "LS4_WITHIN_LINE_EXACT_REPETITION", "cs2/prev-family-current-family":
		return "TOKEN_TRANSITION", d.WithinGroupTransitions, d.WithinGroupTransitions, nulls
	case "cs7/edit-distance-x-structural-distance":
		return "VOCABULARY_PAIR", d.VocabularySize * (d.VocabularySize - 1) / 2, d.VocabularySize * (d.VocabularySize - 1) / 2, nulls
	case "2DL1_LAYOUT_POSITION_MI", "BP1_BOUNDARY_TOKEN_NMI", "LS2_POSITIONAL_LEXICON_NMI", "cs1/family-line-position":
		return "TOKEN_IN_ASSEMBLER_LINE", d.TokenCount, 0, nulls
	case "cs6/family-diversity-x-line-length":
		return "ASSEMBLER_LINE", d.NonEmptyGroups, 0, nulls
	default:
		return "REQUIRED_STRUCTURE_UNAVAILABLE", 0, 0, 0
	}
}

func writeNarrativeOutputs(out string, vecs []vector) error {
	var direct, assembler strings.Builder
	direct.WriteString("metric_id\tfamily\tclassification\tcomparison_namespace\n")
	assembler.WriteString("metric_id\tfamily\tclassification\tcomparison_namespace\n")
	for _, m := range metrics {
		if m.Status == "DIRECTLY_APPLICABLE" {
			direct.WriteString(fmt.Sprintf("%s\t%s\t%s\tF2_COMMON_DIRECT\n", m.ID, m.Family, m.Class))
		}
		if m.Status == "ASSEMBLER_APPLICABLE" {
			assembler.WriteString(fmt.Sprintf("%s\t%s\t%s\tASSEMBLER_LINE_ONLY\n", m.ID, m.Family, m.Class))
		}
	}
	if err := write(out, "F2_COMMON_DIRECT.tsv", direct.String()); err != nil {
		return err
	}
	if err := write(out, "F2_ASSEMBLER_PROJECTION.tsv", assembler.String()); err != nil {
		return err
	}
	var eligibility strings.Builder
	eligibility.WriteString("mechanism_id\tscaling_policy_id\tcomparison_eligibility\treason\n")
	type availability struct{ have, total int }
	byCondition := map[string]availability{}
	for _, v := range vecs {
		k := v.MechanismID + "\t" + v.Policy
		a := byCondition[k]
		for _, m := range v.Metrics {
			spec, ok := findMetric(m.MetricID)
			if ok && spec.Class == "CORE" && spec.Status == "DIRECTLY_APPLICABLE" {
				a.total++
				if m.Available {
					a.have++
				}
			}
		}
		byCondition[k] = a
	}
	keys := make([]string, 0, len(byCondition))
	for k := range byCondition {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		a := byCondition[k]
		status, reason := "NOT_COMPARABLE", "No direct CORE observation survives data degeneracy in this condition."
		if a.have == a.total && a.total > 0 {
			status, reason = "PARTIALLY_COMPARABLE", "All direct CORE metrics are available, but manuscript-specific CORE families are structurally absent."
		} else if a.have > 0 {
			status, reason = "PROJECTION_COMPARABLE", "Only a subset of direct CORE observations is available because of data degeneracy; coverage penalty is mandatory."
		}
		eligibility.WriteString(fmt.Sprintf("%s\t%s\t%s (%d/%d direct CORE observations)\n", k, status, reason, a.have, a.total))
	}
	if err := write(out, "MECHANISM_COMPARISON_ELIGIBILITY_V2.tsv", eligibility.String()); err != nil {
		return err
	}
	contract := "# Task83 comparison contract (frozen target-blind)\n\n## Allowed spaces\n\nThe metric IDs in `F2_COMMON_DIRECT.tsv` are the only permitted direct synthetic/target intersection. `F2_ASSEMBLER_PROJECTION.tsv` is an `ASSEMBLER_LINE_ONLY` diagnostic space and must never be compared numerically with target physical-line metrics. Hierarchy, locus, folio, recto/verso, section, hand, and Currier-dependent metrics are excluded. The intersection is fixed by semantics, not values.\n\n## Missingness and eligibility\n\nMissing values remain missing: no zero, mean, target, control, or model imputation. `PARTIALLY_COMPARABLE` requires all three direct CORE edit-family metrics in every compared cell; partial availability is `PROJECTION_COMPARABLE`; no direct CORE value is `NOT_COMPARABLE`. Supporting metrics cannot rescue missing CORE eligibility.\n\n## Normalization and distance\n\nFor each available direct metric, use the frozen F2 standardized per-metric difference and its pre-target natural-language-control scale. A metric lacking its frozen normalization scale is excluded with an explicit reason, never rescaled from Task82a or target values. Average metric distances within family, then average families with equal family weight. Preserve replicate uncertainty and report cell-level distance intervals; do not pool LOCAL and GLOBAL cue policies.\n\nLet `C` be available eligible direct-family weight divided by total eligible direct-family weight in `F2_COMMON_DIRECT.tsv`. Report both raw pairwise-available distance `D` and coverage-adjusted distance `D_adjusted = D / C`; if `C=0`, distance is unavailable. This preregistered penalty prevents missing difficult families from improving a score without imputing their values. Report metric, family, and CORE coverage beside every distance.\n\n## Aggregation constraints\n\nMechanism-level summaries keep corpus, replicate, scale, and scaling-policy strata visible. Family aggregation is family-balanced. Instability is propagated as `UNSTABLE_FOR_CONFIRMATORY_COMPARISON`; it is not a basis for deleting a mechanism. No metric, threshold, normalization, null, family weight, coverage rule, or uncertainty rule may change after target access. No ranking or selection decision is authorized by this contract.\n"
	if err := write(out, "TASK83_COMPARISON_CONTRACT.md", contract); err != nil {
		return err
	}
	handoff := "# Task83 handoff V2\n\nThe frozen portfolio is ready for a partial, target-blind direct comparison under `TASK83_COMPARISON_CONTRACT.md`. Direct edit-family CORE metrics and direct lexical/cross-scale SUPPORTING projections are retained subject to per-cell availability. Assembler-only 2D-lite, boundary, line, and cross-scale metrics are segregated and are not target-comparison inputs. Folio, hierarchy, locus, and manuscript-metadata families remain explicitly unavailable. All mechanisms and policies remain in the handoff, including degenerate conditions; `MECHANISM_COMPARISON_ELIGIBILITY_V2.tsv` supplies technical eligibility only.\n"
	if err := write(out, "TASK83_HANDOFF_V2.md", handoff); err != nil {
		return err
	}
	corpusTally := tallyLastColumn(filepath.Join(out, "F2_CROSS_CORPUS_STABILITY_EXTENDED.tsv"))
	seedTally := tallyLastColumn(filepath.Join(out, "F2_CROSS_SEED_STABILITY_EXTENDED.tsv"))
	scaleTally := tallyLastColumn(filepath.Join(out, "F2_CROSS_SCALE_STABILITY_EXTENDED.tsv"))
	report := fmt.Sprintf("# Task82a.1 report\n\n## Audit and extension\n\nAll 33 frozen metrics were audited: 8 `DIRECTLY_APPLICABLE`, 9 `ASSEMBLER_APPLICABLE`, 5 `NOT_APPLICABLE_STRUCTURE`, 11 `NOT_APPLICABLE_METADATA`, and 0 `ESTIMATOR_INCOMPATIBLE`. Six metrics (the observed 2D-lite, boundary, and line estimators) were absent only because Task82a bounded the full Task79 pipeline by cost; all six were recovered across 468 immutable jobs. Existing EF/LP/cross-scale results were imported rather than recalculated.\n\nThe generic extraction is regression-tested against task79-v1. No mathematical definition, null, normalization, bootstrap unit, threshold, direction, or classification changed. `2DL1` retains the frozen implementation behavior despite a prose mismatch in the registry; correcting that mismatch would require a new fingerprint version. No synthetic hierarchy was introduced.\n\n## Coverage and stability\n\nMaximum direct CORE metric coverage is 3/13 (one direct CORE family); assembler projection adds 4/13 CORE metrics (three families), for 7/13 mathematically available CORE metrics. Hierarchy, locus, and folio CORE families remain unavailable because folios, physical lines, sections, loci, recto/verso, and manuscript metadata do not exist. Direct and assembler coverage are never combined for target distance.\n\nCross-corpus tally: %s. Cross-seed tally: %s. MEDIUM/LARGE scale tally: %s. Exact rows and cue LOCAL/GLOBAL effects are in the corresponding TSV files. Inconclusive imported metrics remain distinct from structural non-applicability; vocabulary-collapse cases are marked `NOT_APPLICABLE_DATA_DEGENERACY` in sample diagnostics and are not removed.\n\n## Required answers\n\n1. Audited: 33. 2. Direct: 8. 3. Assembler: 9. 4. Structure-unavailable: 5. 5. Metadata-unavailable: 11. 6. Estimator-incompatible: 0. 7. Previously cost-bounded only: 6. 8. Recovered: all 6. 9. Genericized dependency: task79-v1 ordered-group observed estimators. 10. Mathematical definition changed: no. 11. Direct CORE: 3/13. 12. Assembler CORE: 4/13. 13-14. Hierarchy/locus/folio remain unavailable for absent manuscript structure/metadata. 15-18. Corpus, seed, scale, and cue-policy results are frozen in the stability tables. 19. Data-degenerate cells are explicit. 20. `F2_COMMON_DIRECT.tsv` is frozen. 21. `F2_ASSEMBLER_PROJECTION.tsv` is separate. 22. Eligibility is per mechanism/policy. 23. Task83 contract is frozen. 24-25. Both firewalls were preserved. 26. Portfolio readiness is partial because full-manuscript comparison is structurally impossible.\n\n## Final verdicts\n\n| Verdict | Result |\n| --- | --- |\n| TASK82A_INPUT_FREEZE_INTEGRITY | SUPPORTED |\n| F2_APPLICABILITY_AUDIT_COMPLETE | SUPPORTED |\n| F2_COST_VS_STRUCTURE_SEPARATED | SUPPORTED |\n| NO_SYNTHETIC_HIERARCHY_INTRODUCED | SUPPORTED |\n| F2_GENERIC_EXTRACTION_VALID | SUPPORTED |\n| F2_DIRECT_CORE_COVERAGE | 3/13 PARTIAL |\n| F2_PROJECTION_CORE_COVERAGE | 4/13 PARTIAL |\n| F2_CROSS_CORPUS_STABILITY | PARTIAL |\n| F2_CROSS_SEED_STABILITY | PARTIAL |\n| F2_SCALE_STABILITY | PARTIAL |\n| TASK83_COMPARISON_CONTRACT_FROZEN | SUPPORTED |\n| TASK83_PORTFOLIO_READY | PARTIAL |\n| VOYNICH_FIREWALL_PRESERVED | SUPPORTED |\n| NOTATION_CONTROL_FIREWALL_PRESERVED | SUPPORTED |\n\n**TASK82A_F2_COVERAGE_EXTENDED_FROZEN**\n", formatTally(corpusTally), formatTally(seedTally), formatTally(scaleTally))
	return write(out, "TASK82A1_REPORT.md", report)
}

func tallyLastColumn(path string) map[string]int {
	out := map[string]int{}
	s := bufio.NewScanner(strings.NewReader(string(mustRead(path))))
	first := true
	for s.Scan() {
		if first {
			first = false
			continue
		}
		parts := strings.Split(s.Text(), "\t")
		if len(parts) > 0 {
			out[parts[len(parts)-1]]++
		}
	}
	return out
}

func formatTally(m map[string]int) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", k, m[k]))
	}
	return strings.Join(parts, ", ")
}

func writeMeasuredStability(out string, vecs []vector) error {
	var rows []stabilityObservation
	for _, v := range vecs {
		for _, m := range v.Metrics {
			if m.Available {
				rows = append(rows, stabilityObservation{v.MechanismID, v.Policy, v.Corpus, v.Scale, m.MetricID, v.Replicate, m.Value})
			}
		}
	}
	class := func(values []float64) string {
		if len(values) < 2 {
			return "INCONCLUSIVE"
		}
		lo, hi := values[0], values[0]
		for _, v := range values[1:] {
			if v < lo {
				lo = v
			}
			if v > hi {
				hi = v
			}
		}
		switch d := hi - lo; {
		case d <= 0.01:
			return "STABLE"
		case d <= 0.10:
			return "PARTIALLY_STABLE"
		default:
			return "UNSTABLE"
		}
	}
	rangeOf := func(values []float64) float64 {
		if len(values) == 0 {
			return 0
		}
		lo, hi := values[0], values[0]
		for _, v := range values[1:] {
			lo, hi = math.Min(lo, v), math.Max(hi, v)
		}
		return hi - lo
	}
	aggregate := func(key func(stabilityObservation) string, filter func(stabilityObservation) bool) map[string][]float64 {
		m := map[string][]float64{}
		for _, r := range rows {
			if filter(r) {
				m[key(r)] = append(m[key(r)], r.value)
			}
		}
		return m
	}
	writeGroups := func(name, header string, groups map[string][]float64, verdict func([]float64) string) error {
		keys := make([]string, 0, len(groups))
		for k := range groups {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var b strings.Builder
		b.WriteString(header)
		for _, k := range keys {
			v := groups[k]
			b.WriteString(fmt.Sprintf("%s\t%d\t%.6f\t%s\n", strings.ReplaceAll(k, "|", "\t"), len(v), rangeOf(v), verdict(v)))
		}
		return write(out, name, b.String())
	}
	corpusGroups := aggregate(func(r stabilityObservation) string {
		return r.mech + "|" + r.policy + "|" + r.scale + "|" + strconv.Itoa(r.rep) + "|" + r.metric
	}, func(stabilityObservation) bool { return true })
	if err := writeGroups("F2_CROSS_CORPUS_STABILITY_EXTENDED.tsv", "mechanism_id\tscaling_policy_id\tcorpus_scale\treplicate\tmetric_id\tn\trange\tstability\n", corpusGroups, class); err != nil {
		return err
	}
	seedGroups := aggregate(func(r stabilityObservation) string {
		return r.mech + "|" + r.policy + "|" + r.scale + "|" + r.corpus + "|" + r.metric
	}, func(stabilityObservation) bool { return true })
	if err := writeGroups("F2_CROSS_SEED_STABILITY_EXTENDED.tsv", "mechanism_id\tscaling_policy_id\tcorpus_scale\tcorpus_id\tmetric_id\tn\trange\tstability\n", seedGroups, class); err != nil {
		return err
	}
	byScale := map[string]map[string]float64{}
	for _, r := range rows {
		k := r.mech + "|" + r.policy + "|" + r.corpus + "|" + strconv.Itoa(r.rep) + "|" + r.metric
		if byScale[k] == nil {
			byScale[k] = map[string]float64{}
		}
		byScale[k][r.scale] = r.value
	}
	var scale strings.Builder
	scale.WriteString("mechanism_id\tscaling_policy_id\tcorpus_id\treplicate\tmetric_id\tn\trange\tconvergence\n")
	keys := make([]string, 0, len(byScale))
	for k := range byScale {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := byScale[k]
		pair := []float64{}
		if med, ok := v["MEDIUM"]; ok {
			if large, ok := v["LARGE"]; ok {
				pair = []float64{med, large}
			}
		}
		convergence := "NOT_APPLICABLE"
		if len(pair) == 2 {
			switch d := rangeOf(pair); {
			case d <= 0.01:
				convergence = "CONVERGED"
			case d <= 0.10:
				convergence = "PARTIALLY_CONVERGED"
			default:
				convergence = "NOT_CONVERGED"
			}
		}
		scale.WriteString(fmt.Sprintf("%s\t%d\t%.6f\t%s\n", strings.ReplaceAll(k, "|", "\t"), len(pair), rangeOf(pair), convergence))
	}
	if err := write(out, "F2_CROSS_SCALE_STABILITY_EXTENDED.tsv", scale.String()); err != nil {
		return err
	}
	var policy strings.Builder
	policy.WriteString("metric_id\tmechanism_id\tcorpus_id\tcorpus_scale\treplicate\tlocal_global_range\teffect\n")
	byPolicy := map[string]map[string]float64{}
	for _, r := range rows {
		if !strings.HasPrefix(r.policy, "CUE_") {
			continue
		}
		k := r.metric + "|" + r.mech + "|" + r.corpus + "|" + r.scale + "|" + strconv.Itoa(r.rep)
		if byPolicy[k] == nil {
			byPolicy[k] = map[string]float64{}
		}
		byPolicy[k][r.policy] = r.value
	}
	keys = keys[:0]
	for k := range byPolicy {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := byPolicy[k]
		local, lok := v["CUE_RESET_LOCAL_V1"]
		global, gok := v["CUE_RESET_GLOBAL_V1"]
		effect, spread := "NOT_APPLICABLE", 0.0
		if lok && gok {
			spread = math.Abs(local - global)
			effect = class([]float64{local, global})
		}
		policy.WriteString(fmt.Sprintf("%s\t%.6f\t%s\n", strings.ReplaceAll(k, "|", "\t"), spread, effect))
	}
	if err := write(out, "F2_SCALING_POLICY_EFFECT_EXTENDED.tsv", policy.String()); err != nil {
		return err
	}
	return writeFamilyStability(out, vecs, corpusGroups, seedGroups, byScale, class)
}

func writeFamilyStability(out string, vecs []vector, corpusGroups, seedGroups map[string][]float64, byScale map[string]map[string]float64, classify func([]float64) string) error {
	type agg struct {
		available, total, corpusUnstable, seedUnstable, scaleUnstable int
		statuses                                                      map[string]bool
	}
	groups := map[string]*agg{}
	for _, v := range vecs {
		for _, m := range v.Metrics {
			k := v.MechanismID + "|" + v.Policy + "|" + m.Family
			if groups[k] == nil {
				groups[k] = &agg{statuses: map[string]bool{}}
			}
			groups[k].total++
			if m.Available {
				groups[k].available++
			}
			if spec, ok := findMetric(m.MetricID); ok {
				groups[k].statuses[spec.Status] = true
			}
		}
	}
	for key, vals := range corpusGroups {
		parts := strings.Split(key, "|")
		spec, ok := findMetric(parts[len(parts)-1])
		if ok && classify(vals) == "UNSTABLE" {
			groups[parts[0]+"|"+parts[1]+"|"+spec.Family].corpusUnstable++
		}
	}
	for key, vals := range seedGroups {
		parts := strings.Split(key, "|")
		spec, ok := findMetric(parts[len(parts)-1])
		if ok && classify(vals) == "UNSTABLE" {
			groups[parts[0]+"|"+parts[1]+"|"+spec.Family].seedUnstable++
		}
	}
	for key, vals := range byScale {
		parts := strings.Split(key, "|")
		spec, ok := findMetric(parts[len(parts)-1])
		med, mok := vals["MEDIUM"]
		large, lok := vals["LARGE"]
		if ok && mok && lok && math.Abs(med-large) > 0.10 {
			groups[parts[0]+"|"+parts[1]+"|"+spec.Family].scaleUnstable++
		}
	}
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("mechanism_id\tscaling_policy_id\tfamily\tavailable_observations\ttotal_observations\tcorpus_unstable_groups\tseed_unstable_groups\tscale_not_converged_groups\tapplicability\tcoverage\n")
	for _, k := range keys {
		a := groups[k]
		status := "MIXED"
		if len(a.statuses) == 1 {
			for s := range a.statuses {
				status = s
			}
		}
		b.WriteString(fmt.Sprintf("%s\t%d\t%d\t%d\t%d\t%d\t%s\t%.6f\n", strings.ReplaceAll(k, "|", "\t"), a.available, a.total, a.corpusUnstable, a.seedUnstable, a.scaleUnstable, status, safeRatio(a.available, a.total)))
	}
	return write(out, "F2_FAMILY_STABILITY.tsv", b.String())
}

func safeRatio(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) / float64(b)
}

func freeze(out, root string, vecs []vector) error {
	entries, err := os.ReadDir(out)
	if err != nil {
		return err
	}
	checks := map[string]string{}
	for _, e := range entries {
		if e.IsDir() || e.Name() == "TASK82A1_RESULTS_MANIFEST.json" || e.Name() == "TASK82A_F2_COVERAGE_EXTENDED_FROZEN" {
			continue
		}
		h, err := hashFile(filepath.Join(out, e.Name()))
		if err != nil {
			return err
		}
		checks[e.Name()] = h
	}
	m := map[string]any{
		"version": "V1.0", "completed_job_count": len(vecs), "completed_extension_metric_jobs": len(vecs) * 6,
		"audited_metric_count": len(metrics), "metrics_per_extended_vector": len(metrics),
		"task82a_results_manifest_sha256": mustHash(filepath.Join(root, "research/phase2/task82a/TASK82A_RESULTS_MANIFEST.json")),
		"f2_freeze_manifest_sha256":       mustHash(filepath.Join(root, "research/phase2/fingerprint/FINGERPRINT_V2_FREEZE_MANIFEST.json")),
		"f2_metric_registry_sha256":       mustHash(filepath.Join(root, "research/phase2/fingerprint/F2_METRIC_REGISTRY_FINAL.tsv")),
		"raw_document_integrity":          "468/468 embedded observable and assembler-corpus checksums verified",
		"output_artifact_checksums":       checks,
		"firewall_attestations":           map[string]bool{"voynich": true, "notation_control": true},
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	if err := write(out, "TASK82A1_RESULTS_MANIFEST.json", string(b)+"\n"); err != nil {
		return err
	}
	return write(out, "TASK82A_F2_COVERAGE_EXTENDED_FROZEN", "TASK82A_F2_COVERAGE_EXTENDED_FROZEN\nversion=V1.0\nresults_manifest_sha256="+mustHash(filepath.Join(out, "TASK82A1_RESULTS_MANIFEST.json"))+"\n")
}

func metricClassFamily(id string) (string, string) {
	m, _ := findMetric(id)
	return m.Class, m.Family
}
func findMetric(id string) (metricSpec, bool) {
	for _, m := range metrics {
		if m.ID == id {
			return m, true
		}
	}
	return metricSpec{}, false
}
func write(dir, name, data string) error {
	return os.WriteFile(filepath.Join(dir, name), []byte(data), 0o644)
}
func hashFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}
func mustHash(path string) string {
	h, err := hashFile(path)
	if err != nil {
		panic(err)
	}
	return h
}
func mustRead(path string) []byte {
	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return b
}
