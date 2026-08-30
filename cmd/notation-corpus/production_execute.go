package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"zcore.dev/voinich/internal/notation"
)

// productionRunChecklist enumerates the frozen checkpoint set. Only 5000
// and 10000 are reachable by any candidate in the frozen C01,C02,C06
// subset (max observed size ~19132); 20000 and 39380 are preregistered
// NOT_COMPARABLE for this panel (PRODUCTION_CORPUS_SELECTION.json
// panel_assessment) and are still explicitly evaluated and recorded as
// such, never silently skipped.
var productionRunChecklist = notation.RarefactionCheckpoints

// ---- section 1: revalidation ----

type authorizationSnapshotFile struct {
	SchemaVersion string `json:"schema_version"`
	Authorized    bool   `json:"PRODUCTION_COMPARATIVE_RUN_AUTHORIZED"`
	GitCommit     string `json:"git_commit"`
	Bindings      struct {
		ProductionCorpusSelection frozenFileRef            `json:"production_corpus_selection"`
		ProductionCorpusManifest  frozenFileRef            `json:"production_corpus_manifest"`
		GlobalFreezeManifest      frozenFileRef            `json:"global_freeze_manifest"`
		ProductionRunManifest     frozenFileRef            `json:"production_run_manifest"`
		CandidateBundleSHA256     map[string]string        `json:"candidate_bundle_sha256"`
	} `json:"bindings"`
}

type revalidationResult struct {
	AuthorizationRevalidated bool
	InputBindingsValid       bool
	FrozenProtocolUnchanged  bool
	CorpusSelectionUnchanged bool
	WorktreeClean            bool
	GitCommit                string
	Errors                   []string
}

func (r revalidationResult) ok() bool {
	return r.AuthorizationRevalidated && r.InputBindingsValid && r.FrozenProtocolUnchanged && r.CorpusSelectionUnchanged && r.WorktreeClean && len(r.Errors) == 0
}

// revalidateAuthorization implements task run03 section 1. It only reads
// and cross-checks existing frozen/authorization state; it never rewrites
// PRODUCTION_RUN_AUTHORIZATION.json or performs a new authorization
// decision.
// runOutputDirPrefix is the one directory the production-run pipeline is
// itself allowed to have modified/created inside the repository at
// revalidation time: it holds nothing but generated run bundles, one per
// run id, and is exactly what this command is in the middle of writing —
// including, for the independent reproducibility second pass, the
// not-yet-committed primary pass's own bundle. Ignoring status lines under
// it is not a weaker clean-tree check on the actual frozen inputs (code,
// protocol, corpus): every one of those still fails revalidation on any
// real change, since none of them live under this path.
const runOutputDirPrefix = "research/comparative_notation/production_runs/"

// filterOutRunOutputStatusLines drops `git status --porcelain` lines whose
// path falls under runOutputDirPrefix, so a run bundle in progress never
// makes its own (or a nested reproducibility pass's) revalidation fail
// closed on itself.
func filterOutRunOutputStatusLines(status string) string {
	var kept []string
	for _, line := range strings.Split(status, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		fields := strings.Fields(trimmed)
		path := fields[len(fields)-1]
		if strings.HasPrefix(path, runOutputDirPrefix) {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

func revalidateAuthorization(repo string) revalidationResult {
	base := filepath.Join(repo, "research", "comparative_notation")
	var r revalidationResult

	authRaw, err := os.ReadFile(filepath.Join(base, "PRODUCTION_RUN_AUTHORIZATION.json"))
	if err != nil {
		r.Errors = append(r.Errors, "cannot read PRODUCTION_RUN_AUTHORIZATION.json: "+err.Error())
		return r
	}
	var auth authorizationSnapshotFile
	if err := json.Unmarshal(authRaw, &auth); err != nil {
		r.Errors = append(r.Errors, "cannot parse PRODUCTION_RUN_AUTHORIZATION.json: "+err.Error())
		return r
	}
	if !auth.Authorized {
		r.Errors = append(r.Errors, "PRODUCTION_RUN_AUTHORIZATION.json says PRODUCTION_COMPARATIVE_RUN_AUTHORIZED=false")
		return r
	}
	r.AuthorizationRevalidated = true

	checkHash := func(label, relPath, want string) {
		got, err := notation.FileSHA256(filepath.Join(repo, filepath.FromSlash(relPath)))
		if err != nil {
			r.Errors = append(r.Errors, label+": cannot hash "+relPath+": "+err.Error())
			return
		}
		if got != want {
			r.Errors = append(r.Errors, fmt.Sprintf("%s changed since authorization: %s now hashes to %s, authorization recorded %s", label, relPath, got, want))
		}
	}
	checkHash("global_freeze_manifest", auth.Bindings.GlobalFreezeManifest.Path, auth.Bindings.GlobalFreezeManifest.SHA256)
	checkHash("production_corpus_selection", auth.Bindings.ProductionCorpusSelection.Path, auth.Bindings.ProductionCorpusSelection.SHA256)
	checkHash("production_corpus_manifest", auth.Bindings.ProductionCorpusManifest.Path, auth.Bindings.ProductionCorpusManifest.SHA256)
	checkHash("production_run_manifest", auth.Bindings.ProductionRunManifest.Path, auth.Bindings.ProductionRunManifest.SHA256)
	for _, id := range productionSubsetCandidateOrder {
		want, ok := auth.Bindings.CandidateBundleSHA256[id]
		if !ok {
			r.Errors = append(r.Errors, "authorization has no recorded candidate_bundle_sha256 for "+id)
			continue
		}
		checkHash("candidate_bundle:"+id, filepath.ToSlash(filepath.Join("research", "comparative_notation", "production_corpus", id, "candidate_manifest.json")), want)
	}

	freezeErrs, ferr := notation.VerifyGlobalFreezeManifest(base)
	if ferr != nil {
		r.Errors = append(r.Errors, "global freeze re-verification failed: "+ferr.Error())
	} else if len(freezeErrs) != 0 {
		r.Errors = append(r.Errors, freezeErrs...)
	}
	crossErrs, cerr := notation.GlobalFreezeCrossReferenceChecks(base)
	if cerr != nil {
		r.Errors = append(r.Errors, "global freeze cross-reference re-verification failed: "+cerr.Error())
	} else if len(crossErrs) != 0 {
		r.Errors = append(r.Errors, crossErrs...)
	}
	r.FrozenProtocolUnchanged = ferr == nil && len(freezeErrs) == 0 && cerr == nil && len(crossErrs) == 0

	if err := productionCorpusValidateCmd([]string{"--repo", repo}); err != nil {
		r.Errors = append(r.Errors, "production corpus subset re-validation failed: "+err.Error())
	} else {
		r.CorpusSelectionUnchanged = true
	}

	commit, err := commandOutput(repo, "git", "rev-parse", "HEAD")
	if err != nil {
		r.Errors = append(r.Errors, "cannot resolve git HEAD: "+err.Error())
	} else {
		r.GitCommit = strings.TrimSpace(commit)
	}
	status, err := commandOutput(repo, "git", "status", "--porcelain")
	dirty := strings.TrimSpace(filterOutRunOutputStatusLines(status))
	if err != nil || dirty != "" {
		r.Errors = append(r.Errors, "working tree is not clean at execution time:\n"+dirty)
	} else {
		r.WorktreeClean = true
	}

	r.InputBindingsValid = len(r.Errors) == 0 || (r.AuthorizationRevalidated && r.FrozenProtocolUnchanged && r.CorpusSelectionUnchanged && r.WorktreeClean && len(r.Errors) == 0)
	return r
}

// ---- run identity (section 2) ----

type runIdentity struct {
	SchemaVersion              string            `json:"schema_version"`
	RunID                      string            `json:"run_id"`
	GeneratedAtUTC             string            `json:"generated_at_utc"`
	GitCommit                  string            `json:"git_commit"`
	GlobalFreezeManifestSHA256 string            `json:"global_freeze_manifest_sha256"`
	CorpusSelectionSHA256      string            `json:"corpus_selection_sha256"`
	RunManifestSHA256          string            `json:"production_run_manifest_sha256"`
	AuthorizationSHA256        string            `json:"production_run_authorization_sha256"`
	CandidateBundleSHA256      map[string]string `json:"candidate_bundle_sha256"`
	GoVersion                  string            `json:"go_version"`
	GOOS                       string            `json:"goos"`
	GOARCH                     string            `json:"goarch"`
	SeedScheduleID             string            `json:"seed_schedule_id"`
	BaseSeed                   int64             `json:"base_seed"`
	CandidateOrder             []string          `json:"candidate_order"`
	Representations            []subsetRepresentation `json:"representations"`
	Checkpoints                []int             `json:"checkpoints"`
	RarefactionReplicates      int               `json:"rarefaction_replicates"`
	BootstrapReplicates        int               `json:"bootstrap_replicates"`
}

func buildRunIdentity(base, runID string, reval revalidationResult) (runIdentity, error) {
	freezeSHA, err := notation.FileSHA256(filepath.Join(base, "GLOBAL_FREEZE_MANIFEST.json"))
	if err != nil {
		return runIdentity{}, err
	}
	selSHA, err := notation.FileSHA256(filepath.Join(base, "PRODUCTION_CORPUS_SELECTION.json"))
	if err != nil {
		return runIdentity{}, err
	}
	runManifestSHA, err := notation.FileSHA256(filepath.Join(base, "PRODUCTION_RUN_MANIFEST.json"))
	if err != nil {
		return runIdentity{}, err
	}
	authSHA, err := notation.FileSHA256(filepath.Join(base, "PRODUCTION_RUN_AUTHORIZATION.json"))
	if err != nil {
		return runIdentity{}, err
	}
	candidateSHA := map[string]string{}
	for _, id := range productionSubsetCandidateOrder {
		h, err := notation.FileSHA256(filepath.Join(base, "production_corpus", id, "candidate_manifest.json"))
		if err != nil {
			return runIdentity{}, err
		}
		candidateSHA[id] = h
	}
	goVersion, goos, goarch := goEnv()
	return runIdentity{
		SchemaVersion: "production-run-identity-1.0", RunID: runID, GeneratedAtUTC: time.Now().UTC().Format(time.RFC3339),
		GitCommit: reval.GitCommit, GlobalFreezeManifestSHA256: freezeSHA, CorpusSelectionSHA256: selSHA,
		RunManifestSHA256: runManifestSHA, AuthorizationSHA256: authSHA, CandidateBundleSHA256: candidateSHA,
		GoVersion: goVersion, GOOS: goos, GOARCH: goarch,
		SeedScheduleID: "SHA-256(base_seed,corpus_id,representation_id,family_group,checkpoint,replicate_index)", BaseSeed: notation.BaseSeed,
		CandidateOrder: productionSubsetCandidateOrder, Representations: productionSubsetRepresentations,
		Checkpoints: productionRunChecklist, RarefactionReplicates: notation.RarefactionReplicates, BootstrapReplicates: notation.BootstrapReplicates,
	}, nil
}

// ---- production-run-execute CLI ----

func productionRunExecuteCmd(args []string) error {
	fs := flag.NewFlagSet("production-run-execute", flag.ContinueOnError)
	repo := fs.String("repo", ".", "repository root")
	outDir := fs.String("out-dir", "", "override the result bundle directory (used for reproducibility second-pass runs)")
	runID := fs.String("run-id", "", "override the run id (used for reproducibility second-pass runs)")
	skipReproducibility := fs.Bool("skip-reproducibility", false, "skip the independent second-pass reproduction (used internally by the second pass itself)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	summary, err := executeProductionRun(productionRunOptions{Repo: *repo, OutDirOverride: *outDir, RunIDOverride: *runID, SkipReproducibility: *skipReproducibility})
	if err != nil {
		return err
	}
	fmt.Println(summary.finalStatusText())
	if !summary.Valid {
		return fmt.Errorf("production run did not reach VALID=true; see %s", summary.ReportPath)
	}
	return nil
}

type productionRunOptions struct {
	Repo                string
	OutDirOverride      string
	RunIDOverride       string
	SkipReproducibility bool
}

type productionRunSummary struct {
	RunID      string
	GitCommit  string
	Completed  bool
	Valid      bool
	BundlePath string
	ReportPath string
}

func (s productionRunSummary) finalStatusText() string {
	var b strings.Builder
	fmt.Fprintln(&b, "GLOBAL_COMPARISON_PROTOCOL_FROZEN=true")
	fmt.Fprintln(&b, "GLOBAL_FREEZE_CRYPTOGRAPHICALLY_BOUND=true")
	fmt.Fprintln(&b, "PRODUCTION_CORPUS_SUBSET_FROZEN=true")
	fmt.Fprintln(&b, "PRODUCTION_COMPARATIVE_RUN_AUTHORIZED=true")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "PRODUCTION_COMPARATIVE_RUN_COMPLETED=%t\n", s.Completed)
	fmt.Fprintf(&b, "PRODUCTION_COMPARATIVE_RUN_VALID=%t\n", s.Valid)
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "PRODUCTION_RUN_ID=%s\n", s.RunID)
	fmt.Fprintf(&b, "PRODUCTION_RUN_GIT_COMMIT=%s\n", s.GitCommit)
	fmt.Fprintf(&b, "PRODUCTION_RUN_RESULT_BUNDLE=%s\n", s.BundlePath)
	if s.Valid {
		fmt.Fprintln(&b, "PRODUCTION_COMPARATIVE_RUN_VALID=true")
	}
	return strings.TrimRight(b.String(), "\n")
}

// productionRunID03 derives a readable, collision-resistant run id from the
// current UTC time. It is generated once per top-level invocation and never
// changes after computation starts.
func newProductionRunID(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, time.Now().UTC().Format("20060102T150405Z"))
}

func writeJSONFile(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0644)
}

func createOut(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	return os.Create(path)
}

// loadCandidateUSC resolves and hash-verifies the USC file(s) for one
// included candidate from PRODUCTION_CORPUS_MANIFEST.json, returning
// records keyed by representation_id.
func loadCandidateUSC(repo, base string, cm corpusManifest, candidateID string) (map[string][]notation.Record, error) {
	out := map[string][]notation.Record{}
	for _, c := range cm.Candidates {
		if c.CandidateID != candidateID {
			continue
		}
		for _, ref := range c.USC {
			b, err := verifyAndReadRef(repo, base, ref)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", candidateID, err)
			}
			records, err := notation.ReadJSONL(bytes.NewReader(b))
			if err != nil {
				return nil, fmt.Errorf("%s: %w", candidateID, err)
			}
			if err := notation.Validate(records); err != nil {
				return nil, fmt.Errorf("%s: USC validation failed: %w", candidateID, err)
			}
			out[repKey(ref)] = records
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no USC found for included candidate %s", candidateID)
	}
	return out, nil
}

func loadFrozenVMReference(base string) (notation.Fingerprint, notation.VMReferenceManifest, error) {
	manifestRaw, err := os.ReadFile(filepath.Join(base, "VM_REFERENCE_V2_MANIFEST.json"))
	if err != nil {
		return notation.Fingerprint{}, notation.VMReferenceManifest{}, err
	}
	var manifest notation.VMReferenceManifest
	if err := json.Unmarshal(manifestRaw, &manifest); err != nil {
		return notation.Fingerprint{}, notation.VMReferenceManifest{}, err
	}
	fpRaw, err := os.ReadFile(filepath.Join(base, "VM_REFERENCE_V2.fingerprint.json"))
	if err != nil {
		return notation.Fingerprint{}, notation.VMReferenceManifest{}, err
	}
	if err := notation.VerifyFrozenVMReference(fpRaw, manifest); err != nil {
		return notation.Fingerprint{}, notation.VMReferenceManifest{}, fmt.Errorf("frozen VM reference failed verification: %w", err)
	}
	fp, err := notation.ReadFingerprintJSON(bytes.NewReader(fpRaw))
	if err != nil {
		return notation.Fingerprint{}, notation.VMReferenceManifest{}, err
	}
	return fp, manifest, nil
}

func loadFrozenCalibrationScales(base string) ([]notation.CalibrationScale, error) {
	f, err := os.Open(filepath.Join(base, "CALIBRATION_SCALES.tsv"))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return notation.ReadCalibrationScalesTSV(f)
}

func sortedStrings(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
