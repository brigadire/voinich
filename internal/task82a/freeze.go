package task82a

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ResultsManifest is Task82a's own results-freeze manifest (task82a.txt
// sec.70), analogous to Task82's TASK82_BLIND_RESULTS_MANIFEST.json.
type ResultsManifest struct {
	Version                  string            `json:"version"`
	GitCommit                string            `json:"git_commit_at_execution"`
	CompletedJobCount        int               `json:"completed_job_count"`
	FailedJobCount           int               `json:"failed_job_count"`
	BlindManifestChecksum    string            `json:"blind_manifest_checksum"`
	Task81FreezeVersion      string            `json:"task81_freeze_version"`
	Task82ResultsChecksum    string            `json:"task82_results_manifest_checksum"`
	F2FreezeManifestChecksum string            `json:"f2_freeze_manifest_checksum"`
	OutputArtifactChecksums  map[string]string `json:"output_artifact_checksums"`
	FirewallAttestations     map[string]bool   `json:"firewall_attestations"`
}

var requiredOutputs = []string{
	"TASK82A_BLIND_MANIFEST.json", "TASK82A_COST_MODEL.tsv", "TASK82A_JOB_LEDGER.tsv",
	"SCALING_POLICIES.tsv", "BOUNDARY_PROVENANCE.tsv",
	"CORPUS_SCALE_TRANSFORMATION.tsv", "CORPUS_SCALE_RECOVERY.tsv", "KNOWLEDGE_DEPENDENCE_STABILITY.tsv",
	"COLLISION_SCALING.tsv", "AMBIGUITY_SCALING.tsv", "INPUT_DEPENDENCE.tsv",
	"F2_RAW_VECTORS.jsonl", "F2_COVERAGE.tsv", "F2_CROSS_CORPUS_STABILITY.tsv", "F2_CROSS_SEED_STABILITY.tsv", "F2_CROSS_SCALE_STABILITY.tsv",
	"MECHANISM_ELIGIBILITY.tsv", "TASK82A_REPORT.md",
}

// FreezeResults verifies every mandatory integrity gate (task82a.txt
// sec.70, 76) and, only if they all pass, writes
// TASK82A_RESULTS_MANIFEST.json and the TASK82A_CORPUS_SCALE_PORTFOLIO_FROZEN
// marker. On any gate failure it returns an error and writes nothing.
func FreezeResults(root string) error {
	fz, err := verifyFreeze(root)
	if err != nil {
		return fmt.Errorf("FREEZE_MISMATCH: %w", err)
	}
	manifest := fz.ManifestOnDisk
	dir := outDir(root)
	arts, err := loadArtifacts(root, manifest)
	if err != nil {
		return fmt.Errorf("INCOMPLETE_LEDGER: %w", err)
	}
	if err := Aggregate(root, manifest); err != nil {
		return fmt.Errorf("aggregate-from-raw regeneration failed: %w", err)
	}
	for _, j := range manifest.Jobs {
		a := arts[j.JobID]
		if a.DocumentSHA256 != a.Document.Checksum() {
			return fmt.Errorf("CHECKSUM_FAILURE: job %s observable checksum mismatch", j.JobID)
		}
		if a.FreezeVersion != FreezeVersion {
			return fmt.Errorf("FREEZE_MISMATCH: job %s freeze_version %s != %s", j.JobID, a.FreezeVersion, FreezeVersion)
		}
		if err := verifyNoVoynichOrNotationLeak(a); err != nil {
			return err
		}
	}
	checksums := map[string]string{}
	for _, name := range requiredOutputs {
		h, err := fileHashPath(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("required output missing: %s: %w", name, err)
		}
		checksums[name] = h
	}
	manifestChecksum := checksums["TASK82A_BLIND_MANIFEST.json"]

	task82ResultsChecksum, err := fileHashPath(filepath.Join(root, "research", "phase2", "task82", "TASK82_BLIND_RESULTS_MANIFEST.json"))
	if err != nil {
		return err
	}
	f2Checksum, err := fileHashPath(filepath.Join(root, "research", "phase2", "fingerprint", "FINGERPRINT_V2_FREEZE_MANIFEST.json"))
	if err != nil {
		return err
	}

	rm := ResultsManifest{
		Version: Version, GitCommit: gitHead(root), CompletedJobCount: len(arts), FailedJobCount: 0,
		BlindManifestChecksum: manifestChecksum, Task81FreezeVersion: "V1.1",
		Task82ResultsChecksum: task82ResultsChecksum, F2FreezeManifestChecksum: f2Checksum,
		OutputArtifactChecksums: checksums,
		FirewallAttestations:    map[string]bool{"voynich": true, "notation_control": true},
	}
	data, err := json.MarshalIndent(rm, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	resultsManifestPath := filepath.Join(dir, "TASK82A_RESULTS_MANIFEST.json")
	if err := os.WriteFile(resultsManifestPath, data, 0o644); err != nil {
		return err
	}
	resultsManifestChecksum, err := fileHashPath(resultsManifestPath)
	if err != nil {
		return err
	}

	reportPath := filepath.Join(dir, "TASK82A_REPORT.md")
	reportBytes, err := os.ReadFile(reportPath)
	if err != nil {
		return err
	}
	if !strings.Contains(string(reportBytes), "TASK82A_CORPUS_SCALE_PORTFOLIO_FROZEN") {
		return fmt.Errorf("TASK82A_SCALING_NOT_READY: report's own verdict did not reach TASK82A_CORPUS_SCALE_PORTFOLIO_FROZEN")
	}

	designChecksum, err := fileHashPath(filepath.Join(dir, "TASK82A_DESIGN.md"))
	if err != nil {
		return err
	}

	marker := fmt.Sprintf(`TASK82A_CORPUS_SCALE_PORTFOLIO_FROZEN
version=%s
git_commit=%s
task81_v1_1_mechanism_registry_checksum=%s
task82_results_manifest_checksum=%s
task82a_design_checksum=%s
blind_manifest_checksum=%s
f2_freeze_manifest_checksum=%s
results_manifest_checksum=%s
`, Version, gitHead(root), task81Bindings["MNEMONIC_MECHANISM_REGISTRY.json"], task82ResultsChecksum, designChecksum, manifestChecksum, f2Checksum, resultsManifestChecksum)
	if err := os.WriteFile(filepath.Join(dir, "TASK82A_CORPUS_SCALE_PORTFOLIO_FROZEN"), []byte(marker), 0o644); err != nil {
		return err
	}
	return writeHandoff(root, dir, resultsManifestChecksum)
}

var voynichLeakRe = regexp.MustCompile(`(?i)voynich|zl3b|it2a|bdd-koeln`)

func verifyNoVoynichOrNotationLeak(a Artifact) error {
	blob, err := json.Marshal(a)
	if err != nil {
		return err
	}
	if voynichLeakRe.Match(blob) {
		return fmt.Errorf("VOYNICH_FIREWALL: job %s artifact mentions a Voynich/notation-control identifier", a.Job.JobID)
	}
	return nil
}

func writeHandoff(root, dir, resultsManifestChecksum string) error {
	b, err := os.ReadFile(filepath.Join(dir, "MECHANISM_ELIGIBILITY.tsv"))
	if err != nil {
		return err
	}
	var out strings.Builder
	out.WriteString("# Task83 handoff\n\n")
	out.WriteString("This file contains only frozen Task82a portfolio paths, checksums, mechanism/scaling-policy IDs, F2 coverage, convergence status, and comparison eligibility. It does not include, reference, or imply any Voynich comparison result.\n\n")
	fmt.Fprintf(&out, "Frozen portfolio root: `research/phase2/task82a/`\n\n")
	fmt.Fprintf(&out, "Results manifest checksum: `%s`\n\n", resultsManifestChecksum)
	out.WriteString("Frozen artifact paths (see TASK82A_RESULTS_MANIFEST.json for individual checksums):\n\n")
	for _, name := range requiredOutputs {
		fmt.Fprintf(&out, "- `research/phase2/task82a/%s`\n", name)
	}
	out.WriteString("\n## Mechanism eligibility (technical only, no Voynich similarity)\n\n```\n")
	out.Write(b)
	out.WriteString("```\n\n## Known limitations\n\n")
	out.WriteString("- Only the edit-family (EF)/lexical-paradigm (LP) F2 families and task77's cross-scale (cs1-cs7) family were attempted; hierarchy-dependent families (2DL, BP, HR, LC, LS, PF) require fingerprintv2's Task79Config pipeline and were out of scope on cost grounds (see TASK82A_DESIGN.md).\n")
	out.WriteString("- cs3/cs4/cs5 cross-scale metrics are structurally NOT_APPLICABLE for every job: Task82a's assembled documents never carry real IVTFF locus/Currier/section metadata.\n")
	out.WriteString("- CONTINUE_STATE and CONVENTION_PER_BLOCK/PATH_REUSED_GLOBAL scaling policies were preregistered but not run; see SCALING_POLICIES.tsv for why.\n")
	return os.WriteFile(filepath.Join(root, "research", "phase2", "task82a", "TASK83_HANDOFF.md"), []byte(out.String()), 0o644)
}
