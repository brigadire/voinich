package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"zcore.dev/voinich/internal/notation"
)

const productionRunID = "CNS-PROD01-20260830"

type preflightGate struct {
	ID     string   `json:"id"`
	Passed bool     `json:"passed"`
	Detail string   `json:"detail"`
	Errors []string `json:"errors,omitempty"`
}

type experimentPlan struct {
	ClassID         string   `json:"class_id"`
	Status          string   `json:"status"`
	Corpora         []string `json:"corpora"`
	Representations []string `json:"representations"`
	ProductionReady bool     `json:"production_ready"`
}

type candidateSnapshot struct {
	ClassID         string   `json:"class_id"`
	PlanPath        string   `json:"plan_path"`
	PlanSHA256      string   `json:"plan_sha256,omitempty"`
	PlanStatus      string   `json:"plan_status,omitempty"`
	ProductionReady bool     `json:"production_ready"`
	Corpora         []string `json:"corpora,omitempty"`
	Representations []string `json:"representations,omitempty"`
	InputFiles      []string `json:"input_files"`
	MissingFields   []string `json:"missing_fields,omitempty"`
}

type preflightSnapshot struct {
	RunID      string              `json:"run_id"`
	Valid      bool                `json:"valid"`
	Gates      []preflightGate     `json:"gates"`
	Candidates []candidateSnapshot `json:"candidates"`
	GitCommit  string              `json:"git_commit"`
	GitDirty   bool                `json:"git_dirty"`
	GoVersion  string              `json:"go_version"`
	GOOS       string              `json:"goos"`
	GOARCH     string              `json:"goarch"`
}

// requiredFrozenArtifactNames mirrors notation.RequiredGlobalFreezeArtifacts
// (the authoritative list, task run02 section 1), for callers that only
// need artifact names (e.g. test coverage assertions).
func requiredFrozenArtifactNames() []string {
	names := make([]string, len(notation.RequiredGlobalFreezeArtifacts))
	for i, s := range notation.RequiredGlobalFreezeArtifacts {
		names[i] = s.Path
	}
	return names
}

func productionPreflightCmd(args []string) error {
	fs := flag.NewFlagSet("production-preflight", flag.ContinueOnError)
	repo := fs.String("repo", ".", "repository root")
	runTests := fs.Bool("run-tests", true, "run go test ./... and go vet ./...")
	if err := fs.Parse(args); err != nil {
		return err
	}
	s, err := assessProductionPreflight(*repo, *runTests)
	if err != nil {
		return err
	}
	if err := writeBlockedProductionArtifacts(*repo, s); err != nil {
		return err
	}
	if !s.Valid {
		return fmt.Errorf("production preflight failed closed; diagnostics written, no production computation started")
	}
	return fmt.Errorf("production preflight passed but this preparation command intentionally cannot authorize or execute; use the separately reviewed production executor")
}

func assessProductionPreflight(repo string, runTests bool) (preflightSnapshot, error) {
	s := preflightSnapshot{RunID: productionRunID, GoVersion: runtime.Version(), GOOS: runtime.GOOS, GOARCH: runtime.GOARCH}
	base := filepath.Join(repo, "research", "comparative_notation")

	commit, err := commandOutput(repo, "git", "rev-parse", "HEAD")
	if err != nil {
		s.Gates = append(s.Gates, preflightGate{ID: "software_revision", Passed: false, Detail: "Git revision unavailable", Errors: []string{err.Error()}})
	} else {
		s.GitCommit = strings.TrimSpace(commit)
		status, statusErr := commandOutput(repo, "git", "status", "--porcelain")
		s.GitDirty = statusErr != nil || strings.TrimSpace(status) != ""
		detail := "Git commit recorded and worktree clean"
		var errs []string
		if s.GitDirty {
			detail = "worktree is dirty; executable/frozen artifacts are not bound to the recorded commit"
			errs = append(errs, detail)
		}
		s.Gates = append(s.Gates, preflightGate{ID: "software_revision", Passed: !s.GitDirty, Detail: detail, Errors: errs})
	}

	freezeGate := verifyGlobalFreezeGate(base)
	s.Gates = append(s.Gates, freezeGate)

	candidateGate := preflightGate{ID: "production_candidates", Passed: true, Detail: "C01-C09 inputs and parameters are explicit and ready"}
	for i := 1; i <= 9; i++ {
		classID := fmt.Sprintf("C%02d", i)
		planRel := filepath.ToSlash(filepath.Join("research", "comparative_notation", "experiments", classID, "EXPERIMENT_PLAN.json"))
		planPath := filepath.Join(repo, filepath.FromSlash(planRel))
		c := candidateSnapshot{ClassID: classID, PlanPath: planRel, InputFiles: []string{}}
		b, readErr := os.ReadFile(planPath)
		if readErr != nil {
			c.MissingFields = append(c.MissingFields, "experiment_plan")
		} else {
			c.PlanSHA256 = notation.BytesSHA256(b)
			var p experimentPlan
			if err := json.Unmarshal(b, &p); err != nil {
				c.MissingFields = append(c.MissingFields, "valid_experiment_plan")
			} else {
				c.PlanStatus, c.ProductionReady, c.Corpora, c.Representations = p.Status, p.ProductionReady, p.Corpora, p.Representations
				if p.ClassID != classID {
					c.MissingFields = append(c.MissingFields, "matching_class_id")
				}
			}
		}
		if !c.ProductionReady {
			c.MissingFields = append(c.MissingFields, "production_ready=true")
		}
		corpusRoot := filepath.Join(base, "corpora")
		matches, _ := filepath.Glob(filepath.Join(corpusRoot, classID+"_*", "SOURCE_PROVENANCE.json"))
		for _, path := range matches {
			rel, _ := filepath.Rel(repo, path)
			c.InputFiles = append(c.InputFiles, filepath.ToSlash(rel))
		}
		sort.Strings(c.InputFiles)
		if len(c.InputFiles) == 0 {
			c.MissingFields = append(c.MissingFields, "source_provenance", "source_version", "input_checksums", "normalization_profile", "USC_input")
		}
		if len(c.MissingFields) != 0 {
			candidateGate.Errors = append(candidateGate.Errors, classID+": "+strings.Join(c.MissingFields, ", "))
		}
		s.Candidates = append(s.Candidates, c)
	}
	candidateGate.Passed = len(candidateGate.Errors) == 0
	if !candidateGate.Passed {
		candidateGate.Detail = "one or more C01-C09 candidates lack frozen production inputs or readiness"
	}
	s.Gates = append(s.Gates, candidateGate)

	// Reproduction and schema/output checks cannot legitimately run before all
	// inputs exist. Recording BLOCKED is preferable to executing a partial run.
	if !candidateGate.Passed || !freezeGate.Passed {
		s.Gates = append(s.Gates,
			preflightGate{ID: "vm_reference_reproduction", Passed: false, Detail: "blocked by earlier freeze/input gates", Errors: []string{"not executed"}},
			preflightGate{ID: "minimal_prerun_reproducibility", Passed: false, Detail: "blocked before candidate computation", Errors: []string{"not executed"}},
		)
	}

	if runTests {
		for _, tc := range []struct {
			id   string
			args []string
		}{
			{"unit_integration_and_A1_A10", []string{"test", "./..."}},
			{"go_vet", []string{"vet", "./..."}},
		} {
			out, err := commandOutput(repo, "go", tc.args...)
			g := preflightGate{ID: tc.id, Passed: err == nil, Detail: "passed"}
			if err != nil {
				g.Detail = "failed"
				g.Errors = []string{strings.TrimSpace(out + "\n" + err.Error())}
			}
			s.Gates = append(s.Gates, g)
		}
	} else {
		s.Gates = append(s.Gates, preflightGate{ID: "unit_integration_and_A1_A10", Passed: false, Detail: "not run", Errors: []string{"--run-tests=false is not sufficient for authorization"}})
	}

	s.Valid = true
	for _, g := range s.Gates {
		s.Valid = s.Valid && g.Passed
	}
	return s, nil
}

// verifyGlobalFreezeGate checks that every mandatory frozen protocol
// artifact under base is manifest-bound with a direct, matching SHA-256, and
// that the freeze/authorization booleans are in their required initial
// state. It is shared by the full-panel preflight and the frozen-subset
// production run preflight, since the global protocol freeze is identical
// for both.
// verifyGlobalFreezeGate delegates to notation.VerifyGlobalFreezeManifest
// (schema, duplicate/missing paths, every mandatory SHA-256 binding) and
// notation.GlobalFreezeCrossReferenceChecks (internal consistency between
// frozen artifacts, task run02 section 2) — the single source of truth also
// used directly by `notation-corpus global-freeze-verify`.
func verifyGlobalFreezeGate(base string) preflightGate {
	freezeGate := preflightGate{ID: "global_freeze", Passed: true, Detail: "all mandatory frozen artifacts are manifest-bound, unchanged, and internally consistent"}
	bindingErrs, err := notation.VerifyGlobalFreezeManifest(base)
	if err != nil {
		freezeGate.Errors = append(freezeGate.Errors, err.Error())
	} else {
		freezeGate.Errors = append(freezeGate.Errors, bindingErrs...)
	}
	crossErrs, err := notation.GlobalFreezeCrossReferenceChecks(base)
	if err != nil {
		freezeGate.Errors = append(freezeGate.Errors, err.Error())
	} else {
		freezeGate.Errors = append(freezeGate.Errors, crossErrs...)
	}
	freezeGate.Passed = len(freezeGate.Errors) == 0
	if !freezeGate.Passed {
		freezeGate.Detail = "global freeze contract is incomplete or inconsistent"
	}
	return freezeGate
}

func commandOutput(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var b bytes.Buffer
	cmd.Stdout, cmd.Stderr = &b, &b
	err := cmd.Run()
	return b.String(), err
}

func writeBlockedProductionArtifacts(repo string, s preflightSnapshot) error {
	base := filepath.Join(repo, "research", "comparative_notation")
	prod := filepath.Join(base, "production_run")
	if err := os.MkdirAll(filepath.Join(prod, "validation"), 0755); err != nil {
		return err
	}
	if s.Valid {
		return fmt.Errorf("refusing to emit a blocked artifact set for a passing preflight")
	}

	candidateManifest := map[string]any{
		"schema_version": "production-candidate-manifest-1.0", "run_id": s.RunID,
		"frozen": false, "valid": false, "candidate_order": []string{"C01", "C02", "C03", "C04", "C05", "C06", "C07", "C08", "C09"},
		"candidates": s.Candidates, "metric_families": []string{"G", "T", "S", "L", "D"},
		"checkpoints": []int{5000, 10000, 20000, 39380}, "rarefaction_replicates": notation.RarefactionReplicates,
		"bootstrap_replicates": notation.BootstrapReplicates, "base_seed": notation.BaseSeed,
		"calibration_scales": "research/comparative_notation/CALIBRATION_SCALES.tsv", "vm_reference_version": "VM_REFERENCE_V2",
	}
	runManifest := map[string]any{
		"schema_version": "production-run-manifest-1.0", "run_id": s.RunID, "executable": false,
		"candidate_order":  []string{"C01", "C02", "C03", "C04", "C05", "C06", "C07", "C08", "C09"},
		"checkpoint_order": []int{5000, 10000, 20000, 39380}, "metric_order": []string{"G", "T", "S", "L", "D"},
		"seed_derivation": "SHA-256(base_seed,corpus_id,representation_id,family_group,checkpoint,replicate_index); first 8 bytes non-negative int64",
		"base_seed":       notation.BaseSeed, "rarefaction_replicate_ids": "0..99", "bootstrap_replicate_ids": "0..199",
		"software": map[string]any{"git_commit": s.GitCommit, "git_dirty": s.GitDirty, "go_version": s.GoVersion, "goos": s.GOOS, "goarch": s.GOARCH},
	}
	authorization := map[string]any{
		"schema_version": "production-run-authorization-1.0", "run_id": s.RunID,
		"GLOBAL_COMPARISON_PROTOCOL_FROZEN": true, "PRODUCTION_COMPARATIVE_RUN_AUTHORIZED": false,
		"reason":             "mandatory preflight gates failed; no production computation started",
		"freeze_manifest":    "research/comparative_notation/GLOBAL_FREEZE_MANIFEST.json",
		"candidate_manifest": "research/comparative_notation/PRODUCTION_CANDIDATE_MANIFEST.json",
		"run_manifest":       "research/comparative_notation/PRODUCTION_RUN_MANIFEST.json", "git_commit": s.GitCommit,
	}
	resultManifest := map[string]any{
		"schema_version": "production-comparative-run-manifest-1.0", "run_id": s.RunID, "status": "NOT_STARTED",
		"authorized": false, "completed": false, "valid": false, "result_files": []string{},
	}

	jsonFiles := map[string]any{
		filepath.Join(base, "PRODUCTION_CANDIDATE_MANIFEST.json"):       candidateManifest,
		filepath.Join(base, "PRODUCTION_RUN_MANIFEST.json"):             runManifest,
		filepath.Join(base, "PRODUCTION_RUN_AUTHORIZATION.json"):        authorization,
		filepath.Join(base, "PRODUCTION_COMPARATIVE_RUN_MANIFEST.json"): resultManifest,
		filepath.Join(prod, "manifest.json"):                            resultManifest,
		filepath.Join(prod, "authorization.json"):                       authorization,
		filepath.Join(prod, "validation", "preflight.json"):             s,
	}
	for path, value := range jsonFiles {
		b, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			return err
		}
		b = append(b, '\n')
		if err := os.WriteFile(path, b, 0644); err != nil {
			return err
		}
	}

	var md strings.Builder
	md.WriteString("# Production preflight report\n\n")
	fmt.Fprintf(&md, "Run: `%s`\n\nResult: **FAIL-CLOSED**. No C01-C09 computation was started.\n\n", s.RunID)
	md.WriteString("| Gate | Status | Detail |\n|---|---|---|\n")
	for _, g := range s.Gates {
		status := "FAIL"
		if g.Passed {
			status = "PASS"
		}
		detail := g.Detail
		if len(g.Errors) != 0 {
			detail += ": " + strings.Join(g.Errors, "; ")
		}
		fmt.Fprintf(&md, "| `%s` | %s | %s |\n", g.ID, status, strings.ReplaceAll(detail, "|", "\\|"))
	}
	md.WriteString("\nFinal status:\n\n```text\nGLOBAL_COMPARISON_PROTOCOL_FROZEN=true\nPRODUCTION_COMPARATIVE_RUN_AUTHORIZED=false\nPRODUCTION_COMPARATIVE_RUN_COMPLETED=false\nPRODUCTION_COMPARATIVE_RUN_VALID=false\n```\n")
	if err := os.WriteFile(filepath.Join(base, "PRODUCTION_PREFLIGHT_REPORT.md"), []byte(md.String()), 0644); err != nil {
		return err
	}
	report := "# Production comparative run report\n\n" +
		"The run was not authorized and was not started. Consequently there are no observed candidate results, statistical inferences, bootstrap confidence intervals, calibration comparisons, VM comparisons, anomalies, or research interpretations to report. Treating missing inputs as measurements would violate the frozen protocol.\n\n" +
		"See `PRODUCTION_PREFLIGHT_REPORT.md` for the gate-level diagnostics.\n\n" +
		"```text\nGLOBAL_COMPARISON_PROTOCOL_FROZEN=true\nPRODUCTION_COMPARATIVE_RUN_AUTHORIZED=false\nPRODUCTION_COMPARATIVE_RUN_COMPLETED=false\nPRODUCTION_COMPARATIVE_RUN_VALID=false\n```\n"
	if err := os.WriteFile(filepath.Join(base, "PRODUCTION_COMPARATIVE_RUN_REPORT.md"), []byte(report), 0644); err != nil {
		return err
	}
	validation := "# Production validation report\n\nStatus: `NOT_RUN`. Output schema, aggregate consistency, CI bounds, rarefaction counts, bootstrap counts, and NaN/Inf checks are inapplicable because authorization failed before computation. The absence of candidate result files is intentional and fail-closed.\n"
	if err := os.WriteFile(filepath.Join(prod, "validation", "VALIDATION_REPORT.md"), []byte(validation), 0644); err != nil {
		return err
	}
	return writeProductionChecksums(base, prod)
}

func writeProductionChecksums(base, prod string) error {
	paths := []string{
		"PRODUCTION_CANDIDATE_MANIFEST.json", "PRODUCTION_RUN_MANIFEST.json", "PRODUCTION_PREFLIGHT_REPORT.md",
		"PRODUCTION_RUN_AUTHORIZATION.json", "PRODUCTION_COMPARATIVE_RUN_REPORT.md", "PRODUCTION_COMPARATIVE_RUN_MANIFEST.json",
		"production_run/manifest.json", "production_run/authorization.json", "production_run/validation/preflight.json", "production_run/validation/VALIDATION_REPORT.md",
	}
	var b strings.Builder
	for _, rel := range paths {
		h, err := notation.FileSHA256(filepath.Join(base, filepath.FromSlash(rel)))
		if err != nil {
			return err
		}
		fmt.Fprintf(&b, "%s  %s\n", h, rel)
	}
	return os.WriteFile(filepath.Join(prod, "SHA256SUMS"), []byte(b.String()), 0644)
}
