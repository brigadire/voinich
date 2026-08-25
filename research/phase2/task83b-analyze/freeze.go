package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	localTask83bProducerCommit  = "c7aca5c6b554de6771b5e10a55da5f0e4a56c5e4"
	remoteTask83bProducerCommit = "16b0b1a8160159e3310bfadb6f43aa2dba0af7bb"
)

type parent struct {
	ID     string `json:"id"`
	SHA256 string `json:"sha256"`
}

type binding struct {
	JSONPath string `json:"json_path"`
	ParentID string `json:"parent_id"`
}

type provenanceArtifact struct {
	ID             string    `json:"id"`
	Path           string    `json:"path"`
	SHA256         string    `json:"sha256"`
	Kind           string    `json:"kind"`
	Producer       string    `json:"producer"`
	ProducerCommit string    `json:"producer_git_commit"`
	Arguments      []string  `json:"producer_arguments,omitempty"`
	SeedContract   string    `json:"seed_contract_identifier,omitempty"`
	RegistryHash   string    `json:"metric_registry_checksum,omitempty"`
	Parents        []parent  `json:"parents,omitempty"`
	Bindings       []binding `json:"embedded_bindings,omitempty"`
}

type scientificManifest struct {
	SchemaVersion                int                  `json:"schema_version"`
	Version                      string               `json:"version"`
	Status                       string               `json:"status"`
	SeedContract                 string               `json:"seed_contract_identifier"`
	MetricRegistryChecksum       string               `json:"metric_registry_checksum"`
	HistoricalUnresolvedMetadata map[string]string    `json:"historical_unresolved_metadata"`
	Artifacts                    []provenanceArtifact `json:"artifacts"`
}

func writeScientificManifest(root, out string) error {
	sources := []string{
		"cmd/ivtff-x7-extract/main.go",
		"cmd/codex_prepare/main.go",
		"internal/corpusprep/corpusprep.go",
		"internal/metadatavalidation/alignment.go",
		"internal/metadatavalidation/parser.go",
		"internal/metadatavalidation/run.go",
		"internal/metadatavalidation/types.go",
		"internal/metadatavalidation/validation.go",
		"internal/metadatavalidation/write.go",
		"internal/fingerprintv2/crossscale.go",
		"internal/fingerprintv2/graphalgo.go",
		"internal/fingerprintv2/stats.go",
		"internal/fingerprintv2/task79.go",
		"internal/fingerprintv2/task79c_hierarchy.go",
		"internal/fingerprintv2/task79c_pf4.go",
		"internal/fingerprintv2/task79c_run.go",
		"internal/fingerprintv2/determinism_test.go",
		"internal/mechanismspace/progress.go",
		"internal/mechanismspace/progress_test.go",
		"cmd/fingerprint-v2-analyze/main.go",
		"cmd/task79c-pf4-hr/main.go",
		"cmd/task79c-distance-pareto/main.go",
		"research/phase2/task83b-analyze/main.go",
		"research/phase2/task83b-analyze/freeze.go",
	}
	var sourceTSV strings.Builder
	sourceTSV.WriteString("path\tsha256\tgit_head_at_build\tworktree_role\n")
	for _, path := range sources {
		h, err := fileHash(filepath.Join(root, path))
		if err != nil {
			return err
		}
		fmt.Fprintf(&sourceTSV, "%s\t%s\t%s\tAUTHORITATIVE_PRODUCER_SOURCE\n", path, h, localTask83bProducerCommit)
	}
	if err := os.WriteFile(filepath.Join(out, "SOURCE_PROVENANCE.tsv"), []byte(sourceTSV.String()), 0644); err != nil {
		return err
	}

	var artifacts []provenanceArtifact
	byID := map[string]provenanceArtifact{}
	add := func(id, rel, kind, producer, commit string, parentIDs []string, bindings []binding) error {
		h, err := fileHash(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			return fmt.Errorf("hash %s: %w", rel, err)
		}
		a := provenanceArtifact{ID: id, Path: rel, SHA256: h, Kind: kind, Producer: producer, ProducerCommit: commit, Bindings: bindings}
		if kind != "raw" && kind != "historical" {
			a.SeedContract = "FINGERPRINT_V2.1_PRNG_CONTRACT"
			if registry, ok := byID["old_registry"]; ok {
				a.RegistryHash = registry.SHA256
			}
		}
		switch producer {
		case "fingerprint-v2-analyze":
			a.Arguments = []string{"-config", "bound configuration parent"}
		case "task79c-pf4-hr":
			a.Arguments = []string{"-pf4-permutations", "1000", "-pf4-seed", "20260824", "-hr-folds", "5", "-hr-seed", "40260824", "-hr-min-group-size", "5"}
		case "task79c-distance-pareto":
			a.Arguments = []string{"-discriminative", "combined_discriminative", "-registry", "artifact_zl_metric_registry_json"}
		case "ivtff-x7-extract":
			a.Arguments = []string{"-input", "bound raw parent", "-output", "bound extracted artifact"}
		case "codex_prepare":
			a.Arguments = []string{"prepare", "-encoding", "utf-8", "-case", "preserve", "-line-policy", "preserve"}
		}
		if kind != "raw" && kind != "historical" && len(a.Arguments) == 0 {
			a.Arguments = []string{"deterministic generation from bound parent artifacts"}
		}
		for _, pid := range parentIDs {
			p, ok := byID[pid]
			if !ok {
				return fmt.Errorf("unknown parent %s for %s", pid, id)
			}
			a.Parents = append(a.Parents, parent{ID: pid, SHA256: p.SHA256})
		}
		artifacts = append(artifacts, a)
		byID[id] = a
		return nil
	}

	raw := []struct{ id, path, producer, commit string }{
		{"zl_raw", "data/ZL3b-n.txt", "external IVTFF source", ""},
		{"it_raw", "data/IT2a-n.txt", "external IVTFF source", ""},
		{"doyle", "data_test/pg2097-2.txt", "frozen control corpus", localTask83bProducerCommit},
		{"bdd", "data_test/matched-replicates/filtered/bdd-koeln-edd-c-119-full.txt", "Task79c frozen control preparation", localTask83bProducerCommit},
		{"msdos_full", "data_test/matched-replicates/filtered/msdos2.0-full.txt", "Task79c frozen control preparation", localTask83bProducerCommit},
		{"msdos_rep0", "data_test/matched-replicates/filtered/msdos2.0-rep0.txt", "Task79c frozen control preparation", localTask83bProducerCommit},
		{"msdos_rep1", "data_test/matched-replicates/filtered/msdos2.0-rep1.txt", "Task79c frozen control preparation", localTask83bProducerCommit},
		{"msdos_rep2", "data_test/matched-replicates/filtered/msdos2.0-rep2.txt", "Task79c frozen control preparation", localTask83bProducerCommit},
		{"msdos_rep3", "data_test/matched-replicates/filtered/msdos2.0-rep3.txt", "Task79c frozen control preparation", localTask83bProducerCommit},
		{"msdos_rep4", "data_test/matched-replicates/filtered/msdos2.0-rep4.txt", "Task79c frozen control preparation", localTask83bProducerCommit},
		{"old_registry", "research/phase2/fingerprint/F2_METRIC_REGISTRY_FINAL.tsv", "Task79c scientific definition freeze", localTask83bProducerCommit},
		{"old_zl_metrics", "experiments/fingerprint-v2-task79-v1/canonical-out/metric_registry.json", "Task79 historical baseline", localTask83bProducerCommit},
		{"old_it_metrics", "experiments/fingerprint-v2-task79c-v1/transcription-it-out/metric_registry.json", "Task79c historical baseline", localTask83bProducerCommit},
		{"old_pf4_hr", "experiments/fingerprint-v2-task79c-v1/pf4-hr-out/pf4_hierarchy_result.json", "Task79c historical baseline", localTask83bProducerCommit},
		{"old_distance", "experiments/fingerprint-v2-task79c-v1/distance-pareto-out/full_portfolio_distance_pareto.json", "Task79c historical baseline", localTask83bProducerCommit},
		{"old_stability", "research/phase2/fingerprint/TRANSCRIPTION_STABILITY.tsv", "Task79c historical baseline", localTask83bProducerCommit},
		{"canonical_config_source", "experiments/fingerprint-v2-task79-v1/canonical.yaml", "Task79 frozen config", localTask83bProducerCommit},
		{"it_config_source", "experiments/fingerprint-v2-task79c-v1/transcription-it.yaml", "Task79c frozen config", localTask83bProducerCommit},
		{"controls_config_source", "experiments/fingerprint-v2-task79c-v1/controls-portfolio.yaml", "Task79c frozen config", localTask83bProducerCommit},
		{"task83b_specification", "tasks_ph2/task83b.txt", "Task83b task specification", localTask83bProducerCommit},
		{"task81_freeze", "research/phase2/mechanism-space/TASK81_DESIGN_FROZEN", "Task81 blind freeze", localTask83bProducerCommit},
		{"task82_manifest", "research/phase2/task82/TASK82_BLIND_RESULTS_MANIFEST.json", "Task82 blind freeze", localTask83bProducerCommit},
		{"task82a_manifest", "research/phase2/task82a/TASK82A_BLIND_MANIFEST.json", "Task82a blind freeze", localTask83bProducerCommit},
		{"task82a1_contract", "research/phase2/task82a1/TASK83_COMPARISON_CONTRACT.md", "Task82a.1 target-blind comparison contract", localTask83bProducerCommit},
		{"task82b_manifest", "research/phase2/task82b/TASK82B_MANIFEST.json", "Task82b blind freeze", localTask83bProducerCommit},
	}
	for _, r := range raw {
		if err := add(r.id, r.path, "raw", r.producer, r.commit, nil, nil); err != nil {
			return err
		}
	}
	for i, path := range sources {
		if err := add(fmt.Sprintf("source_%02d", i+1), path, "raw", "Task83b producer source", localTask83bProducerCommit, nil, nil); err != nil {
			return err
		}
	}
	var sourceParents []string
	for i := range sources {
		sourceParents = append(sourceParents, fmt.Sprintf("source_%02d", i+1))
	}
	if err := add("source_bundle", "research/phase2/task83b/SOURCE_PROVENANCE.tsv", "generated", "task83b-analyze", localTask83bProducerCommit, sourceParents, nil); err != nil {
		return err
	}
	if err := add("zl_extracted", "research/phase2/task83b/artifacts/prepared/ZL3b-x7.txt", "generated", "ivtff-x7-extract", localTask83bProducerCommit, []string{"zl_raw", "source_bundle"}, nil); err != nil {
		return err
	}
	if err := add("zl_prepared", "research/phase2/task83b/artifacts/prepared/ZL3b-x7.canonical.txt", "generated", "codex_prepare", localTask83bProducerCommit, []string{"zl_extracted", "source_bundle"}, nil); err != nil {
		return err
	}
	if err := add("zl_prepare_metadata", "research/phase2/task83b/artifacts/prepared/ZL3b-x7.canonical.txt.prepare.json", "generated", "codex_prepare", localTask83bProducerCommit, []string{"zl_extracted", "zl_prepared", "source_bundle"}, []binding{{"input_sha256", "zl_extracted"}, {"output_sha256", "zl_prepared"}}); err != nil {
		return err
	}
	if err := add("it_extracted", "research/phase2/task83b/artifacts/prepared/IT2a-x7.txt", "generated", "ivtff-x7-extract", remoteTask83bProducerCommit, []string{"it_raw", "source_bundle"}, nil); err != nil {
		return err
	}
	if err := add("it_prepared", "research/phase2/task83b/artifacts/prepared/IT2a-x7.canonical.txt", "generated", "codex_prepare", remoteTask83bProducerCommit, []string{"it_extracted", "source_bundle"}, nil); err != nil {
		return err
	}
	if err := add("it_prepare_metadata", "research/phase2/task83b/artifacts/prepared/IT2a-x7.canonical.txt.prepare.json", "generated", "codex_prepare", remoteTask83bProducerCommit, []string{"it_extracted", "it_prepared", "source_bundle"}, []binding{{"input_sha256", "it_extracted"}, {"output_sha256", "it_prepared"}}); err != nil {
		return err
	}
	configs := []struct{ id, path, source string }{
		{"zl_config", "research/phase2/task83b/artifacts/configs/zl.yaml", "canonical_config_source"},
		{"it_config", "research/phase2/task83b/artifacts/configs/it.yaml", "it_config_source"},
		{"controls_config", "research/phase2/task83b/artifacts/configs/controls.yaml", "controls_config_source"},
	}
	for _, c := range configs {
		if err := add(c.id, c.path, "generated", "Task83b output-path canonicalization", remoteTask83bProducerCommit, []string{c.source}, nil); err != nil {
			return err
		}
	}

	inputParents := map[string][]string{
		"zl":       {"zl_prepared", "zl_config", "doyle", "source_bundle"},
		"it":       {"it_prepared", "it_config", "doyle", "source_bundle"},
		"controls": {"zl_prepared", "controls_config", "bdd", "msdos_full", "msdos_rep0", "msdos_rep1", "msdos_rep2", "msdos_rep3", "msdos_rep4", "source_bundle"},
	}
	for _, branch := range []string{"zl", "it", "controls"} {
		producerCommit := remoteTask83bProducerCommit
		if branch == "zl" {
			producerCommit = localTask83bProducerCommit
		}
		dir := filepath.Join(out, "artifacts", branch)
		files, err := walkFiles(dir)
		if err != nil {
			return err
		}
		for _, rel := range files {
			id := "artifact_" + branch + "_" + sanitizeID(rel)
			var bindings []binding
			parentID := "zl_prepared"
			if branch == "it" {
				parentID = "it_prepared"
			}
			switch rel {
			case "fingerprint.json":
				bindings = []binding{{"primary.corpus.sha256", parentID}}
			case "raw_results.json":
				bindings = []binding{{"fingerprint.primary.corpus.sha256", parentID}}
			case "freeze_manifest.json":
				bindings = []binding{{"corpus_sha256", parentID}}
			}
			if branch == "controls" && rel == "fingerprint.json" {
				bindings = append(bindings,
					binding{"controls.0.corpus.sha256", "msdos_full"}, binding{"controls.1.corpus.sha256", "msdos_rep0"},
					binding{"controls.2.corpus.sha256", "msdos_rep1"}, binding{"controls.3.corpus.sha256", "msdos_rep2"},
					binding{"controls.4.corpus.sha256", "msdos_rep3"}, binding{"controls.5.corpus.sha256", "msdos_rep4"},
					binding{"controls.6.corpus.sha256", "bdd"})
			}
			if (branch == "zl" || branch == "it") && rel == "fingerprint.json" {
				bindings = append(bindings, binding{"controls.0.corpus.sha256", "doyle"})
			}
			if (branch == "zl" || branch == "it") && rel == "raw_results.json" {
				bindings = append(bindings, binding{"fingerprint.controls.0.corpus.sha256", "doyle"})
			}
			if branch == "controls" && rel == "raw_results.json" {
				bindings = append(bindings,
					binding{"fingerprint.controls.0.corpus.sha256", "msdos_full"}, binding{"fingerprint.controls.1.corpus.sha256", "msdos_rep0"},
					binding{"fingerprint.controls.2.corpus.sha256", "msdos_rep1"}, binding{"fingerprint.controls.3.corpus.sha256", "msdos_rep2"},
					binding{"fingerprint.controls.4.corpus.sha256", "msdos_rep3"}, binding{"fingerprint.controls.5.corpus.sha256", "msdos_rep4"},
					binding{"fingerprint.controls.6.corpus.sha256", "bdd"})
			}
			path := filepath.ToSlash(filepath.Join("research/phase2/task83b/artifacts", branch, filepath.FromSlash(rel)))
			if err := add(id, path, "normative", "fingerprint-v2-analyze", producerCommit, inputParents[branch], bindings); err != nil {
				return err
			}
		}
	}
	if err := add("pf4_hierarchy", "research/phase2/task83b/artifacts/pf4_hierarchy_result.json", "normative", "task79c-pf4-hr", remoteTask83bProducerCommit, []string{"artifact_zl_line_profiles_json", "source_bundle"}, nil); err != nil {
		return err
	}
	if err := add("combined_discriminative", "research/phase2/task83b/artifacts/combined_discriminative_validation.json", "normative", "Task83b deterministic concatenation", remoteTask83bProducerCommit, []string{"artifact_zl_discriminative_validation_json", "artifact_controls_discriminative_validation_json"}, nil); err != nil {
		return err
	}
	if err := add("distance_pareto", "research/phase2/task83b/artifacts/full_portfolio_distance_pareto.json", "normative", "task79c-distance-pareto", remoteTask83bProducerCommit, []string{"combined_discriminative", "artifact_zl_metric_registry_json", "source_bundle"}, nil); err != nil {
		return err
	}
	if err := add("refrozen_registry", "research/phase2/task83b/F2_METRIC_REGISTRY_REFROZEN.tsv", "normative", "task83b-analyze byte-preserving refreeze", localTask83bProducerCommit, []string{"old_registry"}, nil); err != nil {
		return err
	}
	audits := []struct {
		id, file string
		parents  []string
	}{
		{"multirun_audit", "MULTIRUN_REPRODUCIBILITY.tsv", []string{"source_bundle"}},
		{"effect_audit", "F2_DETERMINISTIC_EFFECT_AUDIT.tsv", []string{"old_zl_metrics", "old_it_metrics", "artifact_zl_metric_registry_json", "artifact_it_metric_registry_json"}},
		{"monte_carlo_audit", "MONTE_CARLO_REFREEZE_AUDIT.tsv", []string{"old_zl_metrics", "old_it_metrics", "artifact_zl_metric_registry_json", "artifact_it_metric_registry_json"}},
		{"registry_equivalence", "F2_REGISTRY_SEMANTIC_EQUIVALENCE.tsv", []string{"old_registry", "refrozen_registry"}},
		{"transcription_revalidation", "CROSS_TRANSCRIPTION_REVALIDATION.tsv", []string{"old_stability", "artifact_zl_metric_registry_json", "artifact_it_metric_registry_json"}},
		{"pf4_revalidation", "PF4_REVALIDATION.tsv", []string{"old_pf4_hr", "pf4_hierarchy"}},
		{"hierarchy_revalidation", "HIERARCHY_REVALIDATION.tsv", []string{"old_pf4_hr", "pf4_hierarchy"}},
		{"control_revalidation", "CONTROL_ORDERING_REVALIDATION.tsv", []string{"old_distance", "distance_pareto"}},
		{"verdict_stability", "TASK79C_DETERMINISTIC_VERDICT_STABILITY.tsv", []string{"transcription_revalidation", "pf4_revalidation", "hierarchy_revalidation", "control_revalidation", "effect_audit"}},
		{"defect_registry", "DETERMINISM_DEFECT_REGISTRY.tsv", []string{"source_bundle"}},
		{"semantics_audit", "SCIENTIFIC_SEMANTICS_AUDIT.tsv", []string{"defect_registry", "old_registry"}},
		{"prng_contract", "PRNG_DETERMINISM_CONTRACT.md", []string{"source_bundle", "old_registry"}},
		{"task83b_design", "TASK83B_DESIGN.md", []string{"task83b_specification", "old_registry"}},
		{"downstream_impact", "DOWNSTREAM_FREEZE_IMPACT.tsv", []string{"task81_freeze", "task82_manifest", "task82a_manifest", "task82a1_contract", "task82b_manifest"}},
	}
	for _, audit := range audits {
		if err := add(audit.id, "research/phase2/task83b/"+audit.file, "normative", "task83b-analyze", localTask83bProducerCommit, audit.parents, nil); err != nil {
			return err
		}
	}
	registryHash := byID["refrozen_registry"].SHA256
	manifest := scientificManifest{
		SchemaVersion:          1,
		Version:                "FINGERPRINT_V2.1_DETERMINISTIC_SCIENTIFIC_REFREEZE",
		Status:                 "AUTHORITATIVE",
		SeedContract:           "FINGERPRINT_V2.1_PRNG_CONTRACT",
		MetricRegistryChecksum: registryHash,
		HistoricalUnresolvedMetadata: map[string]string{
			"sha256":    "3fb9531a11d896b5227e54c8d119cc13986eb69e48e1a5ab72b1a1ba64b5b4c0",
			"status":    "NON_AUTHORITATIVE",
			"authority": "HISTORICAL_UNRESOLVED_METADATA",
		},
		Artifacts: artifacts,
	}
	b, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if err := os.WriteFile(filepath.Join(out, "FINGERPRINT_V2_DETERMINISTIC_MANIFEST.json"), b, 0644); err != nil {
		return err
	}
	return writeProvenanceGraph(filepath.Join(out, "PROVENANCE_GRAPH.tsv"), artifacts)
}

func sanitizeID(path string) string {
	var b strings.Builder
	for _, r := range path {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	return strings.Trim(b.String(), "_")
}

func writeProvenanceGraph(path string, artifacts []provenanceArtifact) error {
	rows := []string{"artifact_id\tartifact_path\tartifact_sha256\tparent_id\tparent_sha256\tproducer\tproducer_git_commit"}
	for _, a := range artifacts {
		producerCommit := a.ProducerCommit
		if producerCommit == "" {
			producerCommit = "N/A"
		}
		if len(a.Parents) == 0 {
			rows = append(rows, fmt.Sprintf("%s\t%s\t%s\tROOT\tN/A\t%s\t%s", a.ID, a.Path, a.SHA256, a.Producer, producerCommit))
			continue
		}
		for _, p := range a.Parents {
			rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s", a.ID, a.Path, a.SHA256, p.ID, p.SHA256, a.Producer, producerCommit))
		}
	}
	sort.Strings(rows[1:])
	return os.WriteFile(path, []byte(strings.Join(rows, "\n")+"\n"), 0644)
}

func writeResultsManifest(root, out string) error {
	files, err := walkFiles(out)
	if err != nil {
		return err
	}
	checksums := map[string]string{}
	for _, rel := range files {
		if strings.HasPrefix(rel, "artifacts/") || rel == "TASK83B_RESULTS_MANIFEST.json" {
			continue
		}
		h, err := fileHash(filepath.Join(out, filepath.FromSlash(rel)))
		if err != nil {
			return err
		}
		checksums[rel] = h
	}
	type resultsManifest struct {
		Schema                  string            `json:"schema"`
		Version                 string            `json:"version"`
		Status                  string            `json:"status"`
		Task83RReady            bool              `json:"task83r_ready"`
		FinalMarker             string            `json:"final_marker"`
		ScientificManifestHash  string            `json:"scientific_manifest_sha256"`
		MetricRegistryHash      string            `json:"metric_registry_sha256"`
		OutputArtifactChecksums map[string]string `json:"output_artifact_checksums"`
		SelfCheck               string            `json:"self_check"`
	}
	scientificHash, err := fileHash(filepath.Join(out, "FINGERPRINT_V2_DETERMINISTIC_MANIFEST.json"))
	if err != nil {
		return err
	}
	registryHash, err := fileHash(filepath.Join(out, "F2_METRIC_REGISTRY_REFROZEN.tsv"))
	if err != nil {
		return err
	}
	m := resultsManifest{
		Schema:                  "TASK83B_RESULTS_MANIFEST_V1",
		Version:                 "FINGERPRINT_V2.1_DETERMINISTIC_SCIENTIFIC_REFREEZE",
		Status:                  "AUTHORITATIVE",
		Task83RReady:            true,
		FinalMarker:             "FINGERPRINT_V2_DETERMINISTIC_SCIENTIFIC_REFROZEN",
		ScientificManifestHash:  scientificHash,
		MetricRegistryHash:      registryHash,
		OutputArtifactChecksums: checksums,
		SelfCheck:               "SHA-256 every top-level Task83b output listed in output_artifact_checksums; artifacts are transitively bound by the scientific manifest; this results manifest excludes itself",
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(out, "TASK83B_RESULTS_MANIFEST.json"), append(b, '\n'), 0644)
}
