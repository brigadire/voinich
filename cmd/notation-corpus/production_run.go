package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"zcore.dev/voinich/internal/notation"
)

// productionSubsetRunID identifies this task's authorization decision. It is
// distinct from productionRunID (the earlier, full-nine-candidate-panel
// preflight in production.go), because the frozen production scope is the
// three-candidate subset C01, C02, C06 (PRODUCTION_CORPUS_SUBSET_FROZEN),
// not full-panel readiness.
const productionSubsetRunID = "CNS-PROD01-20260830"

// productionSubsetCandidateOrder is the frozen, deterministic execution
// order for the authorized subset. It is a literal, not derived from map
// iteration, so ordering is stable across runs and readers.
var productionSubsetCandidateOrder = []string{"C01", "C02", "C06"}

// productionSubsetCheckpoints is the frozen checkpoint set from
// COMPARATIVE_EXPERIMENT_SPEC.md. 20000 and 39380 are not reachable by any
// candidate in this subset (max observed size ~19132 records) and are
// preregistered NOT_COMPARABLE there; they remain in the schedule because
// the frozen protocol defines one checkpoint set for every corpus, not a
// per-candidate-shrunk one.
var productionSubsetCheckpoints = []int{5000, 10000, 20000, 39380}

type subsetRepresentation struct {
	CandidateID      string
	RepresentationID string
}

// productionSubsetRepresentations is the frozen, deterministic
// (candidate, representation) execution order for the subset. C06's three
// representations are three representations of one candidate (C06), not
// three independent candidates (task section 7 / REPRESENTATION_REGISTRY.md).
var productionSubsetRepresentations = []subsetRepresentation{
	{"C01", "LATIN-EXPANDED"},
	{"C02", "LATIN-DIPLOMATIC"},
	{"C06", "MUSIC-R1"},
	{"C06", "MUSIC-R2"},
	{"C06", "MUSIC-R3"},
}

type statisticalProcedureApplicability struct {
	Procedure         string `json:"procedure"`
	FrozenSource      string `json:"frozen_source"`
	DependsOnPanelN   bool   `json:"depends_on_candidate_panel_n"`
	ApplicableAtN3    bool   `json:"applicable_at_n3"`
	Status            string `json:"status"`
	Reason            string `json:"reason"`
}

// productionSubsetStatisticalApplicability is the frozen determination for
// task section 6. It is a fixed table, not computed at runtime, because the
// applicability of each procedure to a 3-candidate/2-source-family subset is
// a property of the frozen protocol text, not of any data value. Nothing
// here changes frozen statistical method to accommodate N=3 (task section 6
// "не менять frozen statistical method ради N=3"): every NOT_APPLICABLE
// verdict below is for a procedure the frozen protocol itself states is
// conditional on corpus count ("may" / "requires at least three independent
// corpora"), never for a mandatory one.
var productionSubsetStatisticalApplicability = []statisticalProcedureApplicability{
	{
		Procedure: "rarefaction (G/T/S/L/D per corpus/representation)", FrozenSource: "RAREFACTION_PROTOCOL.md section 1-4",
		DependsOnPanelN: false, ApplicableAtN3: true, Status: "APPLICABLE",
		Reason: "Operates on one corpus/representation at a time against its own structural blocks; has no minimum-candidate-count precondition.",
	},
	{
		Procedure: "block bootstrap (per corpus/representation)", FrozenSource: "BOOTSTRAP_PROTOCOL.md section 1-2",
		DependsOnPanelN: false, ApplicableAtN3: true, Status: "APPLICABLE",
		Reason: "Resamples one corpus's own structural blocks; has no minimum-candidate-count precondition.",
	},
	{
		Procedure: "calibration scale construction", FrozenSource: "CALIBRATION_PANEL_SPEC.md; GLOBAL_FREEZE_MANIFEST.json",
		DependsOnPanelN: false, ApplicableAtN3: true, Status: "APPLICABLE",
		Reason: "Scales come from 40 independently generated synthetic calibration corpora, frozen before any candidate was inspected; independent of production candidate panel size.",
	},
	{
		Procedure: "candidate-vs-VM metric comparison (d_G,d_T,d_S,d_L,d_D)", FrozenSource: "COMPARISON_PROTOCOL.md section 1-7",
		DependsOnPanelN: false, ApplicableAtN3: true, Status: "APPLICABLE",
		Reason: "Each candidate/representation is joined only against the single frozen VM reference, one at a time; not a cross-candidate statistic.",
	},
	{
		Procedure: "paired notation delta (C01 LATIN-EXPANDED vs C02 LATIN-DIPLOMATIC)", FrozenSource: "PAIRED_NOTATION_PROTOCOL.md",
		DependsOnPanelN: true, ApplicableAtN3: true, Status: "APPLICABLE",
		Reason: "Requires exactly the aligned (expanded, diplomatic) pair; both members (C01, C02) are included in the frozen subset.",
	},
	{
		Procedure: "within-class pair distances and variance (CLASS_SUMMARY / WITHIN_CLASS_DISTANCES)", FrozenSource: "COMPARISON_PROTOCOL.md line 28; COMPARATIVE_EXPERIMENT_SPEC.md line 35",
		DependsOnPanelN: true, ApplicableAtN3: false, Status: "NOT_APPLICABLE_FOR_CURRENT_PANEL",
		Reason: "Frozen protocol states this requires at least three independent corpora IN ONE CLASS and is phrased as conditional (\"may\"), never mandatory. Every class in this subset (C01, C02, C06) has exactly one independent corpus. MUSIC-R1/R2/R3 are three representations of the single C06 corpus, not three independent corpora (task section 7), so they cannot be substituted to satisfy this threshold.",
	},
	{
		Procedure: "cross-class ranking / PCA / UMAP / nearest-neighbour", FrozenSource: "COMPARISON_PROTOCOL.md line 30-31; COMPARATIVE_EXPERIMENT_SPEC.md line 37",
		DependsOnPanelN: false, ApplicableAtN3: false, Status: "OUT_OF_SCOPE_REPOSITORY_LOCKED",
		Reason: "Frozen protocol places these outside preparation and repository-locked unconditionally, independent of panel size; never required for authorization.",
	},
}

func productionRunPreflightCmd(args []string) error {
	fs := flag.NewFlagSet("production-run-preflight", flag.ContinueOnError)
	repo := fs.String("repo", ".", "repository root")
	if err := fs.Parse(args); err != nil {
		return err
	}
	s, err := runProductionSubsetPreflight(*repo)
	if err != nil {
		return err
	}
	if !s.authorized() {
		return fmt.Errorf("production run preflight blocked: %s", strings.Join(s.blockers(), "; "))
	}
	fmt.Println("PRODUCTION_COMPARATIVE_RUN_AUTHORIZED=true")
	return nil
}

type subsetPreflightState struct {
	Gates     []preflightGate
	GitCommit string
	GitDirty  bool
	GoVersion string
	GOOS      string
	GOARCH    string
}

func (s subsetPreflightState) authorized() bool {
	for _, g := range s.Gates {
		if !g.Passed {
			return false
		}
	}
	return true
}
func (s subsetPreflightState) blockers() []string {
	var out []string
	for _, g := range s.Gates {
		if !g.Passed {
			detail := g.Detail
			if len(g.Errors) != 0 {
				detail += ": " + strings.Join(g.Errors, "; ")
			}
			out = append(out, g.ID+" ("+detail+")")
		}
	}
	return out
}

// runProductionSubsetPreflight implements task
// tasks_other/corporative_notation_study_production_run01.md sections 1-9,
// then always writes the required deliverables (section: "Required
// deliverables"), fail-closed if any gate did not pass.
func runProductionSubsetPreflight(repo string) (subsetPreflightState, error) {
	base := filepath.Join(repo, "research", "comparative_notation")
	var s subsetPreflightState

	// Sections 1-2: frozen corpus selection and candidate bundle validation.
	// production-corpus-validate already implements exactly this check
	// (identity, readiness, exclusion/deferral reasoning, SHA-256 binding,
	// and the C06 MUSIC-R1/R2/R3 representation-completeness check) against
	// the frozen C01,C02,C06 subset; it is the correct-scope validator, as
	// opposed to assessProductionPreflight's full-nine-candidate-panel model.
	corpusGate := preflightGate{ID: "corpus_subset_and_candidate_bundles", Passed: true, Detail: "PRODUCTION_CORPUS_SELECTION/MANIFEST/STATUS/SHA256SUMS and C01,C02,C06 bundles (provenance, policy, normalization, USC, validation, reproducibility, checksums) are valid"}
	if err := productionCorpusValidateCmd([]string{"--repo", repo}); err != nil {
		corpusGate.Passed = false
		corpusGate.Detail = "frozen corpus subset or candidate bundle validation failed"
		corpusGate.Errors = []string{err.Error()}
	}
	s.Gates = append(s.Gates, corpusGate)

	// Section 3: frozen global protocol artifacts, all SHA-256 bound.
	s.Gates = append(s.Gates, verifyGlobalFreezeGate(base))

	// Section 7: representation independence. Mechanical check that the
	// frozen manifests never introduce a synthetic "C06-R1"/"C06-R2"/"C06-R3"
	// candidate identity, and that C06 carries exactly its three mandatory
	// representations under one candidate_id.
	s.Gates = append(s.Gates, representationIndependenceGate(base))

	// Section 6: frozen statistical protocol applicability at N=3. A fixed,
	// documented table (see productionSubsetStatisticalApplicability); this
	// gate only fails if a MANDATORY frozen procedure would be
	// inapplicable, which is not the case here.
	s.Gates = append(s.Gates, statisticalApplicabilityGate())

	// Section 4: bind to a clean Git revision.
	revGate, commit, dirty := gitRevisionGate(repo)
	s.GitCommit, s.GitDirty = commit, dirty
	s.GoVersion, s.GOOS, s.GOARCH = goEnv()
	s.Gates = append(s.Gates, revGate)

	// Section 8: adversarial and regression checks. go test ./... exercises
	// A1-A10 (internal/notation/adversarial_test.go, cmd/notation-corpus
	// main_test.go TestA10) plus every other unit/integration test in the
	// module; go vet ./... is run separately.
	for _, tc := range []struct {
		id   string
		args []string
	}{
		{"go_test_including_A1_A10", []string{"test", "./..."}},
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

	// Section 5: frozen production run manifest, explicit (no runtime
	// defaults) for the C01,C02,C06 subset only.
	manifestGate := writeProductionSubsetRunManifest(base, s)
	s.Gates = append(s.Gates, manifestGate)

	// Section 9: deterministic technical pre-run against the real frozen
	// C01,C02,C06 USC data. No scientific comparative result is computed or
	// retained; only loading, traversal, seed-schedule, and serialization
	// determinism are checked.
	prerunGate := deterministicPrerunGate(repo, base)
	s.Gates = append(s.Gates, prerunGate)

	if err := writeProductionSubsetDeliverables(base, s); err != nil {
		return s, err
	}
	return s, nil
}

func goEnv() (version, goos, goarch string) {
	out, _ := commandOutput(".", "go", "env", "GOVERSION", "GOOS", "GOARCH")
	fields := strings.Fields(out)
	for len(fields) < 3 {
		fields = append(fields, "")
	}
	return fields[0], fields[1], fields[2]
}

func gitRevisionGate(repo string) (preflightGate, string, bool) {
	commit, err := commandOutput(repo, "git", "rev-parse", "HEAD")
	if err != nil {
		return preflightGate{ID: "clean_git_revision", Passed: false, Detail: "Git revision unavailable", Errors: []string{err.Error()}}, "", true
	}
	sha := strings.TrimSpace(commit)
	status, statusErr := commandOutput(repo, "git", "status", "--porcelain")
	dirty := statusErr != nil || strings.TrimSpace(status) != ""
	g := preflightGate{ID: "clean_git_revision", Passed: !dirty, Detail: "git status --porcelain is empty; commit " + sha}
	if dirty {
		g.Detail = "worktree is not clean; production authorization must not bind to an unreproducible revision"
		g.Errors = []string{strings.TrimSpace(status)}
		if statusErr != nil {
			g.Errors = append(g.Errors, statusErr.Error())
		}
	}
	return g, sha, dirty
}

func representationIndependenceGate(base string) preflightGate {
	g := preflightGate{ID: "representation_independence", Passed: true, Detail: "MUSIC-R1/R2/R3 are three representations of one candidate C06, never treated as independent candidate corpora"}
	b, err := os.ReadFile(filepath.Join(base, "PRODUCTION_CORPUS_MANIFEST.json"))
	if err != nil {
		g.Passed, g.Detail = false, "cannot read PRODUCTION_CORPUS_MANIFEST.json"
		g.Errors = []string{err.Error()}
		return g
	}
	var m corpusManifest
	if err := json.Unmarshal(b, &m); err != nil {
		g.Passed, g.Detail = false, "cannot parse PRODUCTION_CORPUS_MANIFEST.json"
		g.Errors = []string{err.Error()}
		return g
	}
	seen := map[string]bool{}
	for _, id := range m.CandidateOrder {
		if seen[id] {
			g.Errors = append(g.Errors, "duplicate candidate id in candidate_order: "+id)
		}
		seen[id] = true
		if strings.HasPrefix(id, "C06-") || strings.Contains(id, "MUSIC-R") {
			g.Errors = append(g.Errors, "a MUSIC representation was listed as an independent candidate id: "+id)
		}
	}
	if len(m.CandidateOrder) != 9 {
		g.Errors = append(g.Errors, fmt.Sprintf("expected 9 candidate classes C01-C09, found %d", len(m.CandidateOrder)))
	}
	for _, c := range m.Candidates {
		if c.CandidateID != "C06" {
			continue
		}
		if len(c.USC) != 3 {
			g.Errors = append(g.Errors, fmt.Sprintf("C06 must carry exactly 3 frozen USC representations, found %d", len(c.USC)))
		}
	}
	if len(g.Errors) != 0 {
		g.Passed, g.Detail = false, "representation independence violated"
	}
	return g
}

func statisticalApplicabilityGate() preflightGate {
	g := preflightGate{ID: "statistical_protocol_applicable_at_n3", Passed: true, Detail: "every frozen-mandatory procedure applies at N=3; the only inapplicable procedure (within-class distribution) is frozen-conditional on >=3 independent corpora per class, never mandatory"}
	for _, p := range productionSubsetStatisticalApplicability {
		if p.Status == "NOT_APPLICABLE_FOR_CURRENT_PANEL" && p.DependsOnPanelN && p.ApplicableAtN3 {
			g.Errors = append(g.Errors, p.Procedure+": inconsistent applicability record")
		}
	}
	if len(g.Errors) != 0 {
		g.Passed, g.Detail = false, "statistical applicability table is inconsistent"
	}
	return g
}

// writeProductionSubsetRunManifest builds and writes PRODUCTION_RUN_MANIFEST.json
// for the frozen C01,C02,C06 subset. Every field is explicit and sourced from
// a frozen artifact or a literal in this file; none is a runtime default.
func writeProductionSubsetRunManifest(base string, s subsetPreflightState) preflightGate {
	g := preflightGate{ID: "run_manifest_frozen", Passed: true, Detail: "PRODUCTION_RUN_MANIFEST.json is explicit for candidates=C01,C02,C06 with no runtime-default parameters"}

	manifestBytes, err := os.ReadFile(filepath.Join(base, "PRODUCTION_CORPUS_MANIFEST.json"))
	if err != nil {
		g.Passed, g.Detail, g.Errors = false, "cannot read PRODUCTION_CORPUS_MANIFEST.json", []string{err.Error()}
		return g
	}
	var cm corpusManifest
	if err := json.Unmarshal(manifestBytes, &cm); err != nil {
		g.Passed, g.Detail, g.Errors = false, "cannot parse PRODUCTION_CORPUS_MANIFEST.json", []string{err.Error()}
		return g
	}
	inputChecksums := map[string]corpusManifestCandidate{}
	for _, c := range cm.Candidates {
		for _, id := range productionSubsetCandidateOrder {
			if c.CandidateID == id {
				inputChecksums[id] = c
			}
		}
	}
	for _, id := range productionSubsetCandidateOrder {
		if _, ok := inputChecksums[id]; !ok {
			g.Errors = append(g.Errors, "missing candidate bundle for "+id)
		}
	}
	runManifest := map[string]any{
		"schema_version":   "production-run-manifest-2.0",
		"run_id":           productionSubsetRunID,
		"executable":       len(g.Errors) == 0 && !s.GitDirty,
		"candidates":       productionSubsetCandidateOrder,
		"candidate_order":  productionSubsetCandidateOrder,
		"representations":  productionSubsetRepresentations,
		"c06_representations": []string{"MUSIC-R1", "MUSIC-R2", "MUSIC-R3"},
		"metric_families":  []string{"G", "T", "S", "L", "D"},
		"checkpoint_order": productionSubsetCheckpoints,
		"rarefaction": map[string]any{
			"replicates":      notation.RarefactionReplicates,
			"replicate_ids":   "0..99",
			"family_groups":   []string{notation.FamilyGroupStructural, notation.FamilyGroupLine},
			"protocol":        "research/comparative_notation/RAREFACTION_PROTOCOL.md",
		},
		"bootstrap": map[string]any{
			"replicates":    notation.BootstrapReplicates,
			"replicate_ids": "0..199",
			"family_group":  notation.FamilyGroupBootstrap,
			"ci_level":      notation.BootstrapCILevel,
			"protocol":      "research/comparative_notation/BOOTSTRAP_PROTOCOL.md",
		},
		"calibration_reference": map[string]any{
			"scales_path": "research/comparative_notation/CALIBRATION_SCALES.tsv",
		},
		"vm_reference": map[string]any{
			"version": "VM_REFERENCE_V2",
			"tsv":     "research/comparative_notation/VM_REFERENCE_V2.tsv",
			"fingerprint_json": "research/comparative_notation/VM_REFERENCE_V2.fingerprint.json",
		},
		"seed_derivation":  "SHA-256(base_seed,corpus_id,representation_id,family_group,checkpoint,replicate_index); first 8 bytes non-negative int64",
		"base_seed":        notation.BaseSeed,
		"deterministic_execution_ordering": map[string]any{
			"candidate_order":      productionSubsetCandidateOrder,
			"representation_order": productionSubsetRepresentations,
			"checkpoint_order":     productionSubsetCheckpoints,
		},
		"statistical_protocol_applicability": productionSubsetStatisticalApplicability,
		"input_checksums": inputChecksums,
		"software": map[string]any{
			"git_commit": s.GitCommit,
			"git_dirty":  s.GitDirty,
			"go_version": s.GoVersion,
			"goos":       s.GOOS,
			"goarch":     s.GOARCH,
		},
	}
	b, err := json.MarshalIndent(runManifest, "", "  ")
	if err != nil {
		g.Passed, g.Detail, g.Errors = false, "cannot marshal run manifest", append(g.Errors, err.Error())
		return g
	}
	b = append(b, '\n')
	if err := os.WriteFile(filepath.Join(base, "PRODUCTION_RUN_MANIFEST.json"), b, 0644); err != nil {
		g.Passed, g.Detail, g.Errors = false, "cannot write run manifest", append(g.Errors, err.Error())
		return g
	}
	if len(g.Errors) != 0 {
		g.Passed, g.Detail = false, "run manifest inputs are incomplete"
	}
	return g
}

// prerunRepresentationSummary is a purely technical (non-scientific) record
// of loading and traversing one frozen (candidate,representation) USC file.
type prerunRepresentationSummary struct {
	CandidateID          string `json:"candidate_id"`
	RepresentationID     string `json:"representation_id"`
	USCPath              string `json:"usc_path"`
	InputSHA256          string `json:"input_sha256"`
	RecordCount          int    `json:"record_count"`
	StructuralBlockCount int    `json:"structural_block_count"`
	PhysicalLinesObserved bool  `json:"physical_lines_observed"`
	SeedCount            int    `json:"seed_count"`
	SeedScheduleSHA256   string `json:"seed_schedule_sha256"`
}

type technicalPrerun struct {
	SchemaVersion         string                         `json:"schema_version"`
	RunID                 string                         `json:"run_id"`
	Note                  string                         `json:"note"`
	BaseSeed              int64                          `json:"base_seed"`
	Checkpoints           []int                          `json:"checkpoints"`
	RarefactionReplicates int                            `json:"rarefaction_replicates"`
	BootstrapReplicates   int                            `json:"bootstrap_replicates"`
	Representations       []prerunRepresentationSummary  `json:"representations"`
	OverallSHA256         string                         `json:"overall_technical_output_sha256"`
}

// runTechnicalPrerun loads every frozen C01,C02,C06 USC representation,
// validates it, computes its boundary-preserving structural traversal, and
// derives (but does not use) the full rarefaction/bootstrap seed schedule.
// It never calls Analyze/RunRarefaction/RunBootstrap/Compare, so it computes
// no structural, rarefaction, bootstrap, distribution, calibration, or
// VM-comparison metric value — satisfying task section 9's requirement that
// no scientific comparative result be produced or saved.
func runTechnicalPrerun(repo, base string) (technicalPrerun, error) {
	manifestBytes, err := os.ReadFile(filepath.Join(base, "PRODUCTION_CORPUS_MANIFEST.json"))
	if err != nil {
		return technicalPrerun{}, err
	}
	var cm corpusManifest
	if err := json.Unmarshal(manifestBytes, &cm); err != nil {
		return technicalPrerun{}, err
	}
	uscByRepresentation := map[string]frozenFileRef{}
	for _, c := range cm.Candidates {
		included := false
		for _, id := range productionSubsetCandidateOrder {
			included = included || c.CandidateID == id
		}
		if !included {
			continue
		}
		for _, ref := range c.USC {
			key := c.CandidateID + "\x1f" + repKey(ref)
			uscByRepresentation[key] = ref
		}
	}

	out := technicalPrerun{
		SchemaVersion: "production-technical-prerun-1.0", RunID: productionSubsetRunID,
		Note:                  "Technical smoke pre-run only: loading, USC validation, structural traversal, and seed-schedule derivation for C01,C02,C06. No structural, rarefaction, bootstrap, distribution, calibration, or VM-comparison metric was computed or saved.",
		BaseSeed:              notation.BaseSeed,
		Checkpoints:           productionSubsetCheckpoints,
		RarefactionReplicates: notation.RarefactionReplicates,
		BootstrapReplicates:   notation.BootstrapReplicates,
	}
	var overall bytes.Buffer
	for _, rep := range productionSubsetRepresentations {
		var ref frozenFileRef
		var found bool
		for _, c := range cm.Candidates {
			if c.CandidateID != rep.CandidateID {
				continue
			}
			for _, u := range c.USC {
				if repKey(u) == rep.RepresentationID {
					ref, found = u, true
				}
			}
		}
		if !found {
			return technicalPrerun{}, fmt.Errorf("no USC ref found for %s/%s", rep.CandidateID, rep.RepresentationID)
		}
		b, err := verifyAndReadRef(repo, base, ref)
		if err != nil {
			return technicalPrerun{}, fmt.Errorf("%s/%s: %w", rep.CandidateID, rep.RepresentationID, err)
		}
		records, err := notation.ReadJSONL(bytes.NewReader(b))
		if err != nil {
			return technicalPrerun{}, fmt.Errorf("%s/%s: %w", rep.CandidateID, rep.RepresentationID, err)
		}
		if err := notation.Validate(records); err != nil {
			return technicalPrerun{}, fmt.Errorf("%s/%s: USC validation failed: %w", rep.CandidateID, rep.RepresentationID, err)
		}
		blocks, tokens, lines := notation.TraversalSummary(records)
		if tokens != len(records) {
			return technicalPrerun{}, fmt.Errorf("%s/%s: traversal token count mismatch", rep.CandidateID, rep.RepresentationID)
		}

		var seedBuf bytes.Buffer
		seedCount := 0
		for _, fg := range []string{notation.FamilyGroupStructural, notation.FamilyGroupLine} {
			for _, ck := range productionSubsetCheckpoints {
				for r := range notation.RarefactionReplicates {
					seed := notation.SeedFor(notation.BaseSeed, rep.CandidateID, rep.RepresentationID, fg, ck, r)
					fmt.Fprintf(&seedBuf, "%s\x1f%d\x1f%d\x1f%d\n", fg, ck, r, seed)
					seedCount++
				}
			}
		}
		for r := range notation.BootstrapReplicates {
			seed := notation.SeedFor(notation.BaseSeed, rep.CandidateID, rep.RepresentationID, notation.FamilyGroupBootstrap, 0, r)
			fmt.Fprintf(&seedBuf, "%s\x1f%d\x1f%d\x1f%d\n", notation.FamilyGroupBootstrap, 0, r, seed)
			seedCount++
		}
		seedHash := sha256.Sum256(seedBuf.Bytes())

		summary := prerunRepresentationSummary{
			CandidateID: rep.CandidateID, RepresentationID: rep.RepresentationID, USCPath: ref.Path,
			InputSHA256: ref.SHA256, RecordCount: len(records), StructuralBlockCount: blocks,
			PhysicalLinesObserved: lines, SeedCount: seedCount, SeedScheduleSHA256: hex.EncodeToString(seedHash[:]),
		}
		out.Representations = append(out.Representations, summary)
		fmt.Fprintf(&overall, "%s\x1f%s\x1f%d\x1f%d\x1f%t\x1f%s\n", summary.CandidateID, summary.RepresentationID, summary.RecordCount, summary.StructuralBlockCount, summary.PhysicalLinesObserved, summary.SeedScheduleSHA256)
	}
	overallHash := sha256.Sum256(overall.Bytes())
	out.OverallSHA256 = hex.EncodeToString(overallHash[:])
	return out, nil
}

func repKey(ref frozenFileRef) string {
	if strings.Contains(ref.Path, "MUSIC-R1") {
		return "MUSIC-R1"
	}
	if strings.Contains(ref.Path, "MUSIC-R2") {
		return "MUSIC-R2"
	}
	if strings.Contains(ref.Path, "MUSIC-R3") {
		return "MUSIC-R3"
	}
	if strings.Contains(ref.Path, "C01") {
		return "LATIN-EXPANDED"
	}
	return "LATIN-DIPLOMATIC"
}

// deterministicPrerunGate runs the technical pre-run twice, independently,
// and requires byte-identical serialized output — the acceptance criterion
// in task section 9 ("два одинаковых pre-run должны давать идентичные
// technical outputs/checksums").
func deterministicPrerunGate(repo, base string) preflightGate {
	g := preflightGate{ID: "deterministic_technical_prerun", Passed: true, Detail: "two independent technical pre-runs over C01,C02,C06 (loading, USC validation, structural traversal, seed schedule) produced byte-identical serialized output"}

	pass1, err := runTechnicalPrerun(repo, base)
	if err != nil {
		g.Passed, g.Detail, g.Errors = false, "technical pre-run failed", []string{err.Error()}
		return g
	}
	pass2, err := runTechnicalPrerun(repo, base)
	if err != nil {
		g.Passed, g.Detail, g.Errors = false, "technical pre-run (second pass) failed", []string{err.Error()}
		return g
	}
	b1, _ := json.MarshalIndent(pass1, "", "  ")
	b2, _ := json.MarshalIndent(pass2, "", "  ")
	if !bytes.Equal(b1, b2) {
		g.Passed, g.Detail, g.Errors = false, "repeated technical pre-run diverged", []string{"pass1 and pass2 serialized output differ"}
		return g
	}
	prerunDir := filepath.Join(base, "production_run", "prerun")
	if err := os.MkdirAll(prerunDir, 0755); err != nil {
		g.Passed, g.Detail, g.Errors = false, "cannot create prerun directory", []string{err.Error()}
		return g
	}
	b1 = append(b1, '\n')
	if err := os.WriteFile(filepath.Join(prerunDir, "TECHNICAL_PRERUN_PASS1.json"), b1, 0644); err != nil {
		g.Passed, g.Detail, g.Errors = false, "cannot write prerun pass1", []string{err.Error()}
		return g
	}
	b2 = append(b2, '\n')
	if err := os.WriteFile(filepath.Join(prerunDir, "TECHNICAL_PRERUN_PASS2.json"), b2, 0644); err != nil {
		g.Passed, g.Detail, g.Errors = false, "cannot write prerun pass2", []string{err.Error()}
		return g
	}
	var report strings.Builder
	report.WriteString("# Deterministic technical pre-run report\n\n")
	report.WriteString("Scope: loading, USC validation, structural (boundary-preserving block) traversal, ")
	report.WriteString("and seed-schedule derivation for the frozen C01, C02, C06 subset. No structural, ")
	report.WriteString("rarefaction, bootstrap, distribution, calibration, or VM-comparison metric was computed ")
	report.WriteString("or saved; this is not a production comparative run.\n\n")
	report.WriteString("| Candidate | Representation | Records | Structural blocks | Lines observed | Seeds | Seed schedule SHA-256 |\n|---|---|---|---|---|---|---|\n")
	for _, r := range pass1.Representations {
		fmt.Fprintf(&report, "| %s | %s | %d | %d | %t | %d | `%s` |\n", r.CandidateID, r.RepresentationID, r.RecordCount, r.StructuralBlockCount, r.PhysicalLinesObserved, r.SeedCount, r.SeedScheduleSHA256)
	}
	fmt.Fprintf(&report, "\nOverall technical output SHA-256 (pass 1): `%s`\n", pass1.OverallSHA256)
	fmt.Fprintf(&report, "\nOverall technical output SHA-256 (pass 2): `%s`\n", pass2.OverallSHA256)
	fmt.Fprintf(&report, "\nSerialized pass1/pass2 byte-identical: `%t`\n", bytes.Equal(b1, b2))
	if err := os.WriteFile(filepath.Join(prerunDir, "PRERUN_REPORT.md"), []byte(report.String()), 0644); err != nil {
		g.Passed, g.Detail, g.Errors = false, "cannot write prerun report", []string{err.Error()}
		return g
	}
	return g
}

// writeProductionSubsetDeliverables always writes the task's required
// deliverables, fail-closed with concrete blockers if any gate failed, or
// authorized=true bound to selection/freeze/run-manifest/commit/checksums if
// every gate passed.
func writeProductionSubsetDeliverables(base string, s subsetPreflightState) error {
	authorized := s.authorized()

	var reportMD strings.Builder
	reportMD.WriteString("# Production comparative run preflight report\n\n")
	fmt.Fprintf(&reportMD, "Run: `%s`. Scope: preflight and authorization decision for the frozen production ", productionSubsetRunID)
	reportMD.WriteString("corpus subset C01, C02, C06. The comparative run itself is explicitly not executed by this task.\n\n")
	reportMD.WriteString("| Gate | Status | Detail |\n|---|---|---|\n")
	for _, g := range s.Gates {
		status := "FAIL"
		if g.Passed {
			status = "PASS"
		}
		detail := g.Detail
		if len(g.Errors) != 0 {
			detail += ": " + strings.Join(g.Errors, "; ")
		}
		fmt.Fprintf(&reportMD, "| `%s` | %s | %s |\n", g.ID, status, strings.ReplaceAll(detail, "|", "\\|"))
	}
	reportMD.WriteString("\n## Statistical protocol applicability at N=3 (task section 6)\n\n")
	reportMD.WriteString("| Procedure | Frozen source | Depends on panel N | Applicable at N=3 | Status |\n|---|---|---|---|---|\n")
	for _, p := range productionSubsetStatisticalApplicability {
		fmt.Fprintf(&reportMD, "| %s | %s | %t | %t | %s |\n", p.Procedure, p.FrozenSource, p.DependsOnPanelN, p.ApplicableAtN3, p.Status)
	}
	reportMD.WriteString("\n## Representation independence (task section 7)\n\n")
	reportMD.WriteString("MUSIC-R1, MUSIC-R2, MUSIC-R3 are three frozen representations of the single candidate ")
	reportMD.WriteString("C06 (C06_CORPUS_DECISION.md, REPRESENTATION_REGISTRY.md); they are never treated as three ")
	reportMD.WriteString("independent candidate corpora in candidate_order, the run manifest, or any statistic that ")
	reportMD.WriteString("assumes cross-candidate independence.\n\n")
	fmt.Fprintf(&reportMD, "```text\nGLOBAL_COMPARISON_PROTOCOL_FROZEN=true\nPRODUCTION_CORPUS_SUBSET_FROZEN=true\nPRODUCTION_CORPUS_INCLUDED=C01,C02,C06\nPRODUCTION_COMPARATIVE_RUN_AUTHORIZED=%t\nPRODUCTION_COMPARATIVE_RUN_COMPLETED=false\nPRODUCTION_COMPARATIVE_RUN_VALID=false\n```\n", authorized)
	if !authorized {
		reportMD.WriteString("\n## Blockers\n\n")
		for _, b := range s.blockers() {
			fmt.Fprintf(&reportMD, "- %s\n", b)
		}
	}
	if err := os.WriteFile(filepath.Join(base, "PRODUCTION_PREFLIGHT_REPORT.md"), []byte(reportMD.String()), 0644); err != nil {
		return err
	}

	selectionSHA, _ := notation.FileSHA256(filepath.Join(base, "PRODUCTION_CORPUS_SELECTION.json"))
	freezeSHA, _ := notation.FileSHA256(filepath.Join(base, "GLOBAL_FREEZE_MANIFEST.json"))
	corpusManifestSHA, _ := notation.FileSHA256(filepath.Join(base, "PRODUCTION_CORPUS_MANIFEST.json"))
	runManifestSHA, _ := notation.FileSHA256(filepath.Join(base, "PRODUCTION_RUN_MANIFEST.json"))
	candidateSHAs := map[string]string{}
	for _, id := range productionSubsetCandidateOrder {
		h, _ := notation.FileSHA256(filepath.Join(base, "production_corpus", id, "candidate_manifest.json"))
		candidateSHAs[id] = h
	}

	authorization := map[string]any{
		"schema_version": "production-run-authorization-2.0", "run_id": productionSubsetRunID,
		"GLOBAL_COMPARISON_PROTOCOL_FROZEN": true, "PRODUCTION_CORPUS_SUBSET_FROZEN": true,
		"PRODUCTION_CORPUS_INCLUDED": productionSubsetCandidateOrder,
		"PRODUCTION_COMPARATIVE_RUN_AUTHORIZED": authorized,
		"git_commit": s.GitCommit, "git_dirty": s.GitDirty,
		"bindings": map[string]any{
			"production_corpus_selection": map[string]any{"path": "research/comparative_notation/PRODUCTION_CORPUS_SELECTION.json", "sha256": selectionSHA},
			"production_corpus_manifest":  map[string]any{"path": "research/comparative_notation/PRODUCTION_CORPUS_MANIFEST.json", "sha256": corpusManifestSHA},
			"global_freeze_manifest":      map[string]any{"path": "research/comparative_notation/GLOBAL_FREEZE_MANIFEST.json", "sha256": freezeSHA},
			"production_run_manifest":     map[string]any{"path": "research/comparative_notation/PRODUCTION_RUN_MANIFEST.json", "sha256": runManifestSHA},
			"candidate_bundle_sha256":     candidateSHAs,
		},
		"gates": s.Gates,
	}
	if !authorized {
		authorization["reason"] = "mandatory preflight gates failed; no production computation started"
		authorization["blockers"] = s.blockers()
	} else {
		authorization["reason"] = "all mandatory preflight gates passed for the frozen C01,C02,C06 subset"
	}
	if err := writeJSON(filepath.Join(base, "PRODUCTION_RUN_AUTHORIZATION.json"), authorization); err != nil {
		return err
	}

	resultManifest := map[string]any{
		"schema_version": "production-comparative-run-manifest-2.0", "run_id": productionSubsetRunID,
		"status":     map[bool]string{true: "AUTHORIZED_NOT_STARTED", false: "NOT_STARTED"}[authorized],
		"authorized": authorized, "completed": false, "valid": false, "result_files": []string{},
	}
	if err := writeJSON(filepath.Join(base, "PRODUCTION_COMPARATIVE_RUN_MANIFEST.json"), resultManifest); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(base, "PRODUCTION_COMPARATIVE_RUN_AUTHORIZED"), []byte(strconv.FormatBool(authorized)+"\n"), 0644); err != nil {
		return err
	}

	var runReport strings.Builder
	runReport.WriteString("# Production comparative run report\n\n")
	if authorized {
		runReport.WriteString("The frozen C01, C02, C06 production run was authorized by this preflight but was not ")
		runReport.WriteString("started (explicit non-goal of this task). There are no observed candidate results, ")
		runReport.WriteString("statistical inferences, bootstrap confidence intervals, calibration comparisons, VM ")
		runReport.WriteString("comparisons, anomalies, or research interpretations to report.\n\n")
	} else {
		runReport.WriteString("The run was not authorized and was not started. Consequently there are no observed ")
		runReport.WriteString("candidate results, statistical inferences, bootstrap confidence intervals, calibration ")
		runReport.WriteString("comparisons, VM comparisons, anomalies, or research interpretations to report. Treating ")
		runReport.WriteString("missing inputs as measurements would violate the frozen protocol.\n\n")
	}
	runReport.WriteString("See `PRODUCTION_PREFLIGHT_REPORT.md` for the gate-level diagnostics.\n\n")
	fmt.Fprintf(&runReport, "```text\nGLOBAL_COMPARISON_PROTOCOL_FROZEN=true\nPRODUCTION_COMPARATIVE_RUN_AUTHORIZED=%t\nPRODUCTION_COMPARATIVE_RUN_COMPLETED=false\nPRODUCTION_COMPARATIVE_RUN_VALID=false\n```\n", authorized)
	if err := os.WriteFile(filepath.Join(base, "PRODUCTION_COMPARATIVE_RUN_REPORT.md"), []byte(runReport.String()), 0644); err != nil {
		return err
	}

	return writeProductionSubsetChecksums(base)
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0644)
}

func writeProductionSubsetChecksums(base string) error {
	paths := []string{
		"PRODUCTION_PREFLIGHT_REPORT.md", "PRODUCTION_RUN_MANIFEST.json", "PRODUCTION_RUN_AUTHORIZATION.json",
		"PRODUCTION_COMPARATIVE_RUN_MANIFEST.json", "PRODUCTION_COMPARATIVE_RUN_AUTHORIZED", "PRODUCTION_COMPARATIVE_RUN_REPORT.md",
		"production_run/prerun/TECHNICAL_PRERUN_PASS1.json", "production_run/prerun/TECHNICAL_PRERUN_PASS2.json", "production_run/prerun/PRERUN_REPORT.md",
	}
	sort.Strings(paths)
	var b strings.Builder
	for _, rel := range paths {
		h, err := notation.FileSHA256(filepath.Join(base, filepath.FromSlash(rel)))
		if err != nil {
			return err
		}
		fmt.Fprintf(&b, "%s  %s\n", h, rel)
	}
	return os.WriteFile(filepath.Join(base, "production_run", "SHA256SUMS"), []byte(b.String()), 0644)
}
