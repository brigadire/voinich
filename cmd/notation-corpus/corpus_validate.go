package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"zcore.dev/voinich/internal/notation"
)

type frozenFileRef struct {
	Path      string `json:"path"`
	SHA256    string `json:"sha256"`
	Records   int    `json:"records,omitempty"`
	SizeBytes int64  `json:"size_bytes,omitempty"`
}

type corpusManifestCandidate struct {
	CandidateID       string          `json:"candidate_id"`
	SelectionStatus   string          `json:"selection_status"`
	CandidateManifest frozenFileRef   `json:"candidate_manifest"`
	Adapter           []frozenFileRef `json:"adapter,omitempty"`
	RawInputs         []frozenFileRef `json:"raw_inputs,omitempty"`
	ExtractionPolicy  frozenFileRef   `json:"extraction_policy,omitempty"`
	Normalization     frozenFileRef   `json:"normalization_profile,omitempty"`
	USC               []frozenFileRef `json:"usc,omitempty"`
	Provenance        frozenFileRef   `json:"provenance,omitempty"`
	DecisionRecord    frozenFileRef   `json:"decision_record,omitempty"`
	Validation        string          `json:"validation_status"`
	Reproducibility   string          `json:"reproducibility_status"`
	ProductionReady   bool            `json:"production_ready"`
	Reason            string          `json:"reason,omitempty"`
}

type corpusManifest struct {
	SchemaVersion  string                    `json:"schema_version"`
	Selection      frozenFileRef             `json:"selection"`
	FullPanelReady bool                      `json:"FULL_CANDIDATE_PANEL_READY"`
	SubsetFrozen   bool                      `json:"PRODUCTION_CORPUS_SUBSET_FROZEN"`
	IncludedCount  int                       `json:"PRODUCTION_CORPUS_INCLUDED_COUNT"`
	Authorized     bool                      `json:"PRODUCTION_COMPARATIVE_RUN_AUTHORIZED"`
	CandidateOrder []string                  `json:"candidate_order"`
	Candidates     []corpusManifestCandidate `json:"candidates"`
}

type corpusSelection struct {
	SchemaVersion      string                   `json:"schema_version"`
	SelectionVersion   string                   `json:"selection_version"`
	GitRevision        string                   `json:"git_revision"`
	FrozenAt           string                   `json:"frozen_at"`
	GlobalFrozen       bool                     `json:"GLOBAL_COMPARISON_PROTOCOL_FROZEN"`
	FullPanelReady     bool                     `json:"FULL_CANDIDATE_PANEL_READY"`
	SubsetFrozen       bool                     `json:"PRODUCTION_CORPUS_SUBSET_FROZEN"`
	IncludedCount      int                      `json:"PRODUCTION_CORPUS_INCLUDED_COUNT"`
	Authorized         bool                     `json:"PRODUCTION_COMPARATIVE_RUN_AUTHORIZED"`
	Included           []string                 `json:"included"`
	Excluded           map[string]selectionItem `json:"excluded"`
	Deferred           map[string]selectionItem `json:"deferred"`
	CandidateManifests map[string]frozenFileRef `json:"candidate_manifests"`
	PanelAssessment    struct {
		MultipleNotationFamilies bool `json:"multiple_notation_families"`
		NotSingleSourceFamily    bool `json:"not_single_source_family"`
		FrozenProtocolApplicable bool `json:"frozen_protocol_applicable"`
		LimitationsRecorded      bool `json:"limitations_recorded"`
	} `json:"panel_assessment"`
}

type selectionItem struct {
	Reason         string        `json:"reason"`
	DecisionRecord frozenFileRef `json:"decision_record"`
}

type candidateBundleIdentity struct {
	CandidateID     string `json:"candidate_id"`
	SelectionStatus string `json:"selection_status"`
	ProductionReady bool   `json:"production_ready"`
}

func productionCorpusValidateCmd(args []string) error {
	fs := flag.NewFlagSet("production-corpus-validate", flag.ContinueOnError)
	repo := fs.String("repo", ".", "repository root")
	if err := fs.Parse(args); err != nil {
		return err
	}
	base := filepath.Join(*repo, "research", "comparative_notation")
	manifestPath := filepath.Join(base, "PRODUCTION_CORPUS_MANIFEST.json")
	b, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}
	var m corpusManifest
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	var errs []string
	if m.SchemaVersion != "production-corpus-manifest-2.0" || len(m.Candidates) != 9 || len(m.CandidateOrder) != 9 {
		errs = append(errs, "manifest must contain exactly C01-C09 under production-corpus-manifest-2.0")
	}
	selectionBytes, err := verifyAndReadRef(*repo, base, m.Selection)
	if err != nil {
		errs = append(errs, "selection: "+err.Error())
	}
	var selection corpusSelection
	if err == nil {
		if err := json.Unmarshal(selectionBytes, &selection); err != nil {
			errs = append(errs, "selection JSON: "+err.Error())
		}
	}
	if selection.SchemaVersion != "production-corpus-selection-1.0" || selection.SelectionVersion == "" || len(selection.GitRevision) != 40 || selection.FrozenAt == "" {
		errs = append(errs, "selection identity/version/revision/timestamp is incomplete")
	}
	if !selection.GlobalFrozen || selection.FullPanelReady || !selection.SubsetFrozen || selection.Authorized {
		errs = append(errs, "selection readiness booleans violate frozen-subset contract")
	}
	if !selection.PanelAssessment.MultipleNotationFamilies || !selection.PanelAssessment.NotSingleSourceFamily || !selection.PanelAssessment.FrozenProtocolApplicable || !selection.PanelAssessment.LimitationsRecorded {
		errs = append(errs, "scientific panel assessment is incomplete")
	}
	included := map[string]bool{}
	for _, id := range selection.Included {
		included[id] = true
	}
	if len(included) != len(selection.Included) || selection.IncludedCount != len(selection.Included) || m.IncludedCount != len(selection.Included) {
		errs = append(errs, "included count/list mismatch")
	}
	if m.FullPanelReady != selection.FullPanelReady || m.SubsetFrozen != selection.SubsetFrozen || m.Authorized != selection.Authorized {
		errs = append(errs, "manifest and selection readiness booleans disagree")
	}
	seen := map[string]bool{}
	ready := 0
	for i, c := range m.Candidates {
		want := fmt.Sprintf("C%02d", i+1)
		if c.CandidateID != want || m.CandidateOrder[i] != want || seen[want] {
			errs = append(errs, fmt.Sprintf("candidate ordering/identity error at %d", i))
		}
		seen[want] = true
		candidateBytes, candidateErr := verifyAndReadRef(*repo, base, c.CandidateManifest)
		if candidateErr != nil {
			errs = append(errs, want+" candidate manifest: "+candidateErr.Error())
		} else {
			var identity candidateBundleIdentity
			if err := json.Unmarshal(candidateBytes, &identity); err != nil || identity.CandidateID != want || identity.SelectionStatus != c.SelectionStatus || identity.ProductionReady != c.ProductionReady {
				errs = append(errs, want+": candidate manifest identity/status mismatch")
			}
		}
		selectedRef, ok := selection.CandidateManifests[want]
		if !ok || selectedRef.Path != c.CandidateManifest.Path || selectedRef.SHA256 != c.CandidateManifest.SHA256 {
			errs = append(errs, want+": selection does not bind candidate manifest")
		}
		if included[want] {
			if c.SelectionStatus != "INCLUDED" || !c.ProductionReady || c.Validation != "PASS" || c.Reproducibility != "BITWISE_PASS" || len(c.RawInputs) == 0 || len(c.USC) == 0 {
				errs = append(errs, want+": included readiness gates are incomplete")
			} else {
				ready++
			}
			expectedReps := map[string][]string{"C01": {"LATIN-EXPANDED"}, "C02": {"LATIN-DIPLOMATIC"}, "C06": {"MUSIC-R1", "MUSIC-R2", "MUSIC-R3"}}[want]
			if len(c.USC) != len(expectedReps) {
				errs = append(errs, want+": representation count mismatch")
			}
			seenReps := map[string]bool{}
			refs := append([]frozenFileRef(nil), c.RawInputs...)
			refs = append(refs, c.Adapter...)
			refs = append(refs, c.ExtractionPolicy, c.Normalization, c.Provenance)
			refs = append(refs, c.USC...)
			for _, ref := range refs {
				if _, err := verifyAndReadRef(*repo, base, ref); err != nil {
					errs = append(errs, want+": "+err.Error())
					continue
				}
				if ref.Records > 0 {
					f, openErr := os.Open(resolveCorpusPath(*repo, base, ref.Path))
					if openErr != nil {
						errs = append(errs, want+": cannot open USC: "+openErr.Error())
						continue
					}
					records, readErr := notation.ReadJSONL(f)
					f.Close()
					if readErr != nil || notation.Validate(records) != nil || len(records) != ref.Records {
						errs = append(errs, fmt.Sprintf("%s: USC schema/count validation failed for %s", want, ref.Path))
					} else {
						seenReps[records[0].Representation] = true
					}
				}
			}
			for _, rep := range expectedReps {
				if !seenReps[rep] {
					errs = append(errs, want+": missing representation "+rep)
				}
			}
			continue
		}
		item, excluded := selection.Excluded[want]
		deferred := false
		if !excluded {
			item, deferred = selection.Deferred[want]
		}
		if !excluded && !deferred {
			errs = append(errs, want+": neither included, excluded, nor deferred")
			continue
		}
		wantStatus := "DEFERRED"
		if excluded {
			wantStatus = strings.Split(c.SelectionStatus, ":")[0]
			if !strings.HasPrefix(c.SelectionStatus, "EXCLUDED_") {
				errs = append(errs, want+": invalid excluded status")
			}
		}
		if !excluded && c.SelectionStatus != wantStatus {
			errs = append(errs, want+": deferred status mismatch")
		}
		if c.ProductionReady || len(c.USC) != 0 || c.Reason == "" || item.Reason == "" {
			errs = append(errs, want+": excluded/deferred candidate contains USC/readiness or lacks reason")
		}
		if item.DecisionRecord.Path != c.DecisionRecord.Path || item.DecisionRecord.SHA256 != c.DecisionRecord.SHA256 {
			errs = append(errs, want+": decision record binding mismatch")
		}
		if _, err := verifyAndReadRef(*repo, base, c.DecisionRecord); err != nil {
			errs = append(errs, want+" decision: "+err.Error())
		}
	}
	if ready != len(selection.Included) {
		errs = append(errs, fmt.Sprintf("ready included count mismatch: computed=%d selection=%d", ready, len(selection.Included)))
	}
	if m.FullPanelReady != (ready == 9) {
		errs = append(errs, "FULL_CANDIDATE_PANEL_READY mismatch")
	}
	marker, markerErr := os.ReadFile(filepath.Join(base, "PRODUCTION_COMPARATIVE_RUN_AUTHORIZED"))
	if m.Authorized || markerErr != nil || strings.TrimSpace(string(marker)) != "false" {
		errs = append(errs, "comparative run authorization must remain false")
	}
	result := map[string]any{
		"schema_version": "production-corpus-bundle-validation-2.0", "manifest_sha256": notation.BytesSHA256(b),
		"selection_sha256": m.Selection.SHA256, "valid_frozen_subset": len(errs) == 0,
		"full_candidate_panel_ready": m.FullPanelReady, "production_corpus_subset_frozen": m.SubsetFrozen,
		"included_count": ready, "candidate_count": len(m.Candidates), "errors": errs, "comparative_metrics_executed": false,
	}
	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	if err := os.WriteFile(filepath.Join(base, "production_corpus", "validation.json"), out, 0644); err != nil {
		return err
	}
	if err := writeCorpusBundleChecksums(base); err != nil {
		return err
	}
	if len(errs) != 0 {
		return fmt.Errorf("production corpus bundle validation failed: %s", strings.Join(errs, "; "))
	}
	fmt.Printf("PRODUCTION-CORPUS-SUBSET-VALID included=%d/9 subset_frozen=true full_panel_ready=%t authorized=false\n", ready, m.FullPanelReady)
	return nil
}

func verifyAndReadRef(repo, base string, ref frozenFileRef) ([]byte, error) {
	if ref.Path == "" || len(ref.SHA256) != 64 {
		return nil, fmt.Errorf("incomplete frozen file reference")
	}
	path := resolveCorpusPath(repo, base, ref.Path)
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ref.Path, err)
	}
	if got := notation.BytesSHA256(b); got != ref.SHA256 {
		return nil, fmt.Errorf("hash mismatch for %s got=%s want=%s", ref.Path, got, ref.SHA256)
	}
	if ref.SizeBytes != 0 && int64(len(b)) != ref.SizeBytes {
		return nil, fmt.Errorf("size mismatch for %s", ref.Path)
	}
	return b, nil
}

func resolveCorpusPath(repo, base, path string) string {
	if strings.HasPrefix(path, "production_corpus/") {
		return filepath.Join(base, filepath.FromSlash(path))
	}
	if !strings.Contains(path, "/") {
		return filepath.Join(base, path)
	}
	return filepath.Join(repo, filepath.FromSlash(path))
}

func writeCorpusBundleChecksums(base string) error {
	var paths []string
	root := filepath.Join(base, "production_corpus")
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(base, path)
		if err != nil {
			return err
		}
		paths = append(paths, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return err
	}
	for _, name := range []string{
		"PRODUCTION_CORPUS_SELECTION.json", "PRODUCTION_CORPUS_MANIFEST.json", "PRODUCTION_CORPUS_STATUS.tsv", "PRODUCTION_CORPUS_PREPARATION_REPORT.md",
		"C03_SOURCE_DECISION.md", "C04_LICENSE_REVIEW.md", "C05_LICENSE_REVIEW.md", "C06_CORPUS_DECISION.md", "C07_LICENSE_REVIEW.md", "C08_SOURCE_DECISION.md", "C09_SOURCE_DECISION.md",
		"experiments/C01/EXPERIMENT_PLAN.json", "experiments/C02/EXPERIMENT_PLAN.json", "experiments/C06/EXPERIMENT_PLAN.json",
	} {
		if _, err := os.Stat(filepath.Join(base, name)); err == nil {
			paths = append(paths, name)
		}
	}
	sort.Strings(paths)
	var out strings.Builder
	for _, rel := range paths {
		h, err := notation.FileSHA256(filepath.Join(base, filepath.FromSlash(rel)))
		if err != nil {
			return err
		}
		fmt.Fprintf(&out, "%s  %s\n", h, rel)
	}
	return os.WriteFile(filepath.Join(base, "PRODUCTION_CORPUS_SHA256SUMS"), []byte(out.String()), 0644)
}
