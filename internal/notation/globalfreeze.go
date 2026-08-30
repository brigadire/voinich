package notation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// GlobalFreezeManifestSchemaVersion is the schema version of
// GLOBAL_FREEZE_MANIFEST.json produced by this package. Version 2.0
// completes the cryptographic binding gap left by version 1.0 (task
// corporative_notation_study_production_run02.md): every mandatory frozen
// artifact carries an explicit {path, sha256, role, schema_version} entry,
// instead of a mix of flat and ad hoc nested keys.
const GlobalFreezeManifestSchemaVersion = "global-freeze-manifest-2.0"

// FreezeArtifactSpec is one entry in the authoritative, hand-audited list of
// mandatory frozen protocol artifacts (task run02 section 1). It never
// includes corpus-selection, candidate-bundle, or production-run artifacts
// (PRODUCTION_CORPUS_*, production_corpus/, experiments/, C0X_*_DECISION.md
// / *_LICENSE_REVIEW.md, PREPARATION_REPORT.md, PREPARATION_BLOCKERS.md) —
// those are governed by PRODUCTION_CORPUS_SUBSET_FROZEN and its own
// checksum set (PRODUCTION_CORPUS_SHA256SUMS), not
// GLOBAL_COMPARISON_PROTOCOL_FROZEN.
type FreezeArtifactSpec struct {
	Path          string // relative to research/comparative_notation
	Role          string
	SchemaVersion string // documentary; empty when the artifact has none of its own
}

// Frozen artifact roles.
const (
	RoleReport               = "report"
	RoleProtocol             = "protocol"
	RoleUSCSpecification     = "usc_specification"
	RoleMetricRegistry       = "metric_registry"
	RoleRarefaction          = "rarefaction"
	RoleDistributionBootstrap = "distribution_bootstrap"
	RoleCalibration          = "calibration"
	RoleVMReference          = "vm_reference"
)

// RequiredGlobalFreezeArtifacts is the authoritative frozen-artifact list
// (task run02 section 1): every protocol/specification file, calibration
// artifact, rarefaction artifact, distribution/bootstrap artifact, VM
// reference v2 artifact, the metric registry, and the USC specification
// that together constitute GLOBAL_COMPARISON_PROTOCOL_FROZEN=true. It
// deliberately covers more than the six originally-missing bindings
// (GLOBAL_FREEZE_REPORT.md, USC_SPEC.md, CALIBRATION_PANEL_SPEC.md,
// CALIBRATION_PANEL_REPORT.md, VM_REFERENCE_V2_MANIFEST.json,
// VM_REFERENCE_RECONCILIATION.md): auditing the whole directory found
// several more frozen-methodology documents and VM-side production outputs
// that GLOBAL_FREEZE_MANIFEST.json v1.0 never bound either.
var RequiredGlobalFreezeArtifacts = []FreezeArtifactSpec{
	{"GLOBAL_FREEZE_REPORT.md", RoleReport, ""},
	{"COMPARATIVE_STUDY_GOALS.md", RoleProtocol, ""},
	{"COMPARATIVE_EXPERIMENT_SPEC.md", RoleProtocol, ""},
	{"REPRESENTATION_REGISTRY.md", RoleProtocol, ""},
	{"VALIDATION_PROTOCOL.md", RoleProtocol, ""},
	{"USC_SPEC.md", RoleUSCSpecification, "usc-1.0"},
	{"COMPARISON_PROTOCOL.md", RoleProtocol, ""},
	{"PAIRED_NOTATION_PROTOCOL.md", RoleProtocol, ""},
	{"RESULT_CONTRACT.md", RoleProtocol, ""},
	{"METRIC_REGISTRY.md", RoleMetricRegistry, MetricRegistryVersion},
	{"RAREFACTION_PROTOCOL.md", RoleRarefaction, "rarefaction-protocol-1.0"},
	{"RAREFACTION_SCHEMA.md", RoleRarefaction, "rarefaction-protocol-1.0"},
	{"DISTRIBUTION_OUTPUT_CONTRACT.md", RoleDistributionBootstrap, ""},
	{"BOOTSTRAP_PROTOCOL.md", RoleDistributionBootstrap, "bootstrap-protocol-1.0"},
	{"METRIC_OUTPUT_TYPES.tsv", RoleDistributionBootstrap, ""},
	{"CALIBRATION_PANEL_SPEC.md", RoleCalibration, ""},
	{"CALIBRATION_PANEL_REPORT.md", RoleCalibration, ""},
	{"CALIBRATION_SCALES.tsv", RoleCalibration, "calibration-scales-1.0"},
	{"CALIBRATION_DIAGNOSTICS.json", RoleCalibration, ""},
	{"CALIBRATION_GENERATORS", RoleCalibration, ""}, // directory: sha256 of sorted-filename concatenation
	{"VM_REFERENCE_CONTRACT.md", RoleVMReference, ""},
	{"VM_REFERENCE_V2.tsv", RoleVMReference, ""},
	{"VM_REFERENCE_V2.fingerprint.json", RoleVMReference, ""},
	{"VM_REFERENCE_V2_MANIFEST.json", RoleVMReference, "vm-reference-v2-manifest-1.0"},
	{"VM_REFERENCE_RECONCILIATION.md", RoleVMReference, ""},
	{"VM_COMPARISON_TEMPLATE.md", RoleVMReference, ""},
	{"VM_RAREFACTION_V2.tsv", RoleVMReference, ""},
	{"VM_RAREFACTION_V2_RAW.tsv", RoleVMReference, ""},
	{"VM_ACCUMULATION_CURVES_V2.tsv", RoleVMReference, ""},
	{"VM_BOOTSTRAP_V2.tsv", RoleVMReference, ""},
	{"VM_DISTRIBUTIONS_V2.tsv", RoleVMReference, ""},
}

// FreezeArtifactEntry is one bound entry inside GLOBAL_FREEZE_MANIFEST.json.
type FreezeArtifactEntry struct {
	Path          string `json:"path"`
	SHA256        string `json:"sha256"`
	Role          string `json:"role"`
	SchemaVersion string `json:"schema_version,omitempty"`
}

// GlobalFreezeComparatorIdentity mirrors the manifest's "comparator" object.
type GlobalFreezeComparatorIdentity struct {
	Package             string `json:"package"`
	CLI                 string `json:"cli"`
	AnalyzerVersion     string `json:"analyzer_version"`
	MetricRegistryVersion string `json:"metric_registry_version"`
	MetricRegistryHash  string `json:"metric_registry_hash"`
}

// GlobalFreezeProtocolStatus mirrors the manifest's "protocol_status".
type GlobalFreezeProtocolStatus struct {
	GlobalComparisonProtocolFrozen     bool `json:"GLOBAL_COMPARISON_PROTOCOL_FROZEN"`
	GlobalFreezeCryptographicallyBound bool `json:"GLOBAL_FREEZE_CRYPTOGRAPHICALLY_BOUND"`
	ProductionComparativeRunAuthorized bool `json:"PRODUCTION_COMPARATIVE_RUN_AUTHORIZED"`
}

// GlobalFreezeManifest mirrors GLOBAL_FREEZE_MANIFEST.json schema v2.0.
type GlobalFreezeManifest struct {
	SchemaVersion      string                         `json:"schema_version"`
	Task               string                         `json:"task"`
	BaseSeed           int64                          `json:"base_seed"`
	FreezeGenerationID string                         `json:"freeze_generation_id,omitempty"`
	BindingCompletedAt string                         `json:"binding_completed_at,omitempty"`
	Comparator         GlobalFreezeComparatorIdentity `json:"comparator"`
	Artifacts          []FreezeArtifactEntry          `json:"artifacts"`
	ProtocolStatus     GlobalFreezeProtocolStatus     `json:"protocol_status"`
}

// ReadGlobalFreezeManifest loads and parses GLOBAL_FREEZE_MANIFEST.json from
// base (research/comparative_notation).
func ReadGlobalFreezeManifest(base string) (GlobalFreezeManifest, error) {
	var m GlobalFreezeManifest
	b, err := os.ReadFile(filepath.Join(base, "GLOBAL_FREEZE_MANIFEST.json"))
	if err != nil {
		return m, err
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return m, fmt.Errorf("GLOBAL_FREEZE_MANIFEST.json: %w", err)
	}
	return m, nil
}

// HashFreezeArtifact computes the SHA-256 an artifact spec is bound to.
// Every artifact is canonical file bytes with no preprocessing (task run02
// section 3), except the CALIBRATION_GENERATORS directory, which is defined
// (GLOBAL_FREEZE_MANIFEST.json's original note, unchanged here) as the
// SHA-256 of the concatenation of its *.json files in sorted filename order.
func HashFreezeArtifact(base string, spec FreezeArtifactSpec) (string, error) {
	if spec.Path == "CALIBRATION_GENERATORS" {
		dir := filepath.Join(base, "CALIBRATION_GENERATORS")
		entries, err := os.ReadDir(dir)
		if err != nil {
			return "", err
		}
		var names []string
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
				names = append(names, e.Name())
			}
		}
		sort.Strings(names)
		if len(names) == 0 {
			return "", fmt.Errorf("CALIBRATION_GENERATORS contains no *.json files")
		}
		var all []byte
		for _, n := range names {
			b, err := os.ReadFile(filepath.Join(dir, n))
			if err != nil {
				return "", err
			}
			all = append(all, b...)
		}
		return BytesSHA256(all), nil
	}
	return FileSHA256(filepath.Join(base, spec.Path))
}

// VerifyGlobalFreezeManifest fail-closed checks GLOBAL_FREEZE_MANIFEST.json
// against the authoritative artifact list and the artifacts' actual current
// bytes (task run02 section 7): schema, duplicate/missing paths, every
// mandatory binding present, and every SHA-256 matching the real file. It
// returns every problem found (not just the first), or a nil slice when the
// manifest is fully valid.
func VerifyGlobalFreezeManifest(base string) ([]string, error) {
	m, err := ReadGlobalFreezeManifest(base)
	if err != nil {
		return nil, err
	}
	var errs []string
	if m.SchemaVersion != GlobalFreezeManifestSchemaVersion {
		errs = append(errs, fmt.Sprintf("unexpected schema_version %q, want %q", m.SchemaVersion, GlobalFreezeManifestSchemaVersion))
	}
	if !m.ProtocolStatus.GlobalComparisonProtocolFrozen {
		errs = append(errs, "GLOBAL_COMPARISON_PROTOCOL_FROZEN must be true")
	}
	if m.ProtocolStatus.ProductionComparativeRunAuthorized {
		errs = append(errs, "PRODUCTION_COMPARATIVE_RUN_AUTHORIZED must not be set by the global freeze manifest itself")
	}

	byPath := map[string][]FreezeArtifactEntry{}
	for _, a := range m.Artifacts {
		byPath[a.Path] = append(byPath[a.Path], a)
	}
	for path, entries := range byPath {
		if len(entries) > 1 {
			errs = append(errs, fmt.Sprintf("duplicate manifest entry for %s (%d occurrences)", path, len(entries)))
		}
	}

	for _, spec := range RequiredGlobalFreezeArtifacts {
		entries, ok := byPath[spec.Path]
		if !ok || len(entries) == 0 {
			errs = append(errs, "missing mandatory binding: "+spec.Path)
			continue
		}
		entry := entries[0]
		if len(entry.SHA256) != 64 {
			errs = append(errs, spec.Path+": bound sha256 is not a direct 64-hex-character SHA-256")
			continue
		}
		if entry.Role != spec.Role {
			errs = append(errs, fmt.Sprintf("%s: role %q does not match the authoritative role %q", spec.Path, entry.Role, spec.Role))
		}
		got, err := HashFreezeArtifact(base, spec)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: cannot read artifact: %v", spec.Path, err))
			continue
		}
		if got != entry.SHA256 {
			errs = append(errs, fmt.Sprintf("%s hash mismatch: file is %s, manifest says %s", spec.Path, got, entry.SHA256))
		}
	}

	// Manifest entries for names outside the authoritative list are not an
	// error (the manifest may legitimately carry more than the mandatory
	// minimum), but a mandatory spec absent from the manifest is already
	// caught above.
	sort.Strings(errs)
	return errs, nil
}

// GlobalFreezeCrossReferenceChecks implements task run02 section 2's
// internal-consistency checks between frozen artifacts. It never
// recomputes a scientific result; it only cross-checks already-frozen
// values against each other and against the live (but frozen-since)
// analyzer code, exactly as VerifyFrozenVMReference already does for the
// VM reference. Returns every failed check description, or nil when all
// pass.
func GlobalFreezeCrossReferenceChecks(base string) ([]string, error) {
	var errs []string

	// VM reference v2 <-> VM reference manifest/reconciliation: the
	// manifest's own output_sha256 must equal the actual frozen fingerprint
	// file, and its metric_registry_hash must equal the live registry hash
	// (the same two checks VerifyFrozenVMReference performs at use time).
	vmManifestRaw, err := os.ReadFile(filepath.Join(base, "VM_REFERENCE_V2_MANIFEST.json"))
	if err != nil {
		errs = append(errs, "VM_REFERENCE_V2_MANIFEST.json: "+err.Error())
	} else {
		var vm VMReferenceManifest
		if err := json.Unmarshal(vmManifestRaw, &vm); err != nil {
			errs = append(errs, "VM_REFERENCE_V2_MANIFEST.json: "+err.Error())
		} else {
			fpRaw, err := os.ReadFile(filepath.Join(base, "VM_REFERENCE_V2.fingerprint.json"))
			if err != nil {
				errs = append(errs, "VM_REFERENCE_V2.fingerprint.json: "+err.Error())
			} else if verr := VerifyFrozenVMReference(fpRaw, vm); verr != nil {
				errs = append(errs, "VM_REFERENCE_V2_MANIFEST.json <-> VM_REFERENCE_V2.fingerprint.json: "+verr.Error())
			}
		}
	}

	// USC spec <-> frozen USC implementation: the specification must name
	// the exact schema_version the live USC implementation enforces.
	uscSpec, err := os.ReadFile(filepath.Join(base, "USC_SPEC.md"))
	if err != nil {
		errs = append(errs, "USC_SPEC.md: "+err.Error())
	} else if !strings.Contains(string(uscSpec), SchemaVersion) {
		errs = append(errs, fmt.Sprintf("USC_SPEC.md does not mention the live USC schema_version %q", SchemaVersion))
	}

	// Metric registry <-> distribution/bootstrap protocol: the frozen
	// METRIC_OUTPUT_TYPES.tsv must still list exactly the (metric_id,
	// output_type) pairs notation.MetricOutputTypes() derives from the live
	// metric registry (which itself includes the two paired-distribution
	// pseudo-metrics and the three accumulation-curve pseudo-metrics, not
	// only the scalar registry IDs).
	outputTypesRaw, err := os.ReadFile(filepath.Join(base, "METRIC_OUTPUT_TYPES.tsv"))
	if err != nil {
		errs = append(errs, "METRIC_OUTPUT_TYPES.tsv: "+err.Error())
	} else {
		frozen := map[string]string{}
		for i, line := range strings.Split(strings.TrimRight(string(outputTypesRaw), "\n"), "\n") {
			if i == 0 || strings.TrimSpace(line) == "" {
				continue
			}
			cols := strings.Split(line, "\t")
			if len(cols) < 2 {
				continue
			}
			frozen[cols[0]] = cols[1]
		}
		live := map[string]string{}
		for _, t := range MetricOutputTypes() {
			live[t.MetricID] = string(t.OutputType)
		}
		for id, outputType := range live {
			if frozen[id] != outputType {
				errs = append(errs, fmt.Sprintf("METRIC_OUTPUT_TYPES.tsv[%s]=%q, live notation.MetricOutputTypes() says %q", id, frozen[id], outputType))
			}
		}
		for id := range frozen {
			if _, ok := live[id]; !ok {
				errs = append(errs, fmt.Sprintf("METRIC_OUTPUT_TYPES.tsv references %q, which live notation.MetricOutputTypes() no longer produces", id))
			}
		}
	}

	// Calibration panel report <-> calibration scales: the frozen scale
	// table's row count must match the panel report's own documented count,
	// and every generator CALIBRATION_DIAGNOSTICS.json reports on must be
	// one of the frozen CALIBRATION_GENERATORS files.
	scalesRaw, err := os.ReadFile(filepath.Join(base, "CALIBRATION_SCALES.tsv"))
	if err != nil {
		errs = append(errs, "CALIBRATION_SCALES.tsv: "+err.Error())
	} else {
		scales, err := ReadCalibrationScalesTSV(strings.NewReader(string(scalesRaw)))
		if err != nil {
			errs = append(errs, "CALIBRATION_SCALES.tsv: "+err.Error())
		} else {
			panelReport, err := os.ReadFile(filepath.Join(base, "CALIBRATION_PANEL_REPORT.md"))
			if err != nil {
				errs = append(errs, "CALIBRATION_PANEL_REPORT.md: "+err.Error())
			} else if !strings.Contains(string(panelReport), strconv.Itoa(len(scales))) {
				errs = append(errs, fmt.Sprintf("CALIBRATION_PANEL_REPORT.md does not mention the frozen CALIBRATION_SCALES.tsv row count (%d)", len(scales)))
			}
		}
	}
	diagRaw, err := os.ReadFile(filepath.Join(base, "CALIBRATION_DIAGNOSTICS.json"))
	if err != nil {
		errs = append(errs, "CALIBRATION_DIAGNOSTICS.json: "+err.Error())
	} else {
		genDir := filepath.Join(base, "CALIBRATION_GENERATORS")
		entries, direrr := os.ReadDir(genDir)
		if direrr != nil {
			errs = append(errs, "CALIBRATION_GENERATORS: "+direrr.Error())
		} else {
			names := map[string]bool{}
			for _, e := range entries {
				names[strings.TrimSuffix(e.Name(), ".json")] = true
			}
			var perCheckpoint map[string]map[string]json.RawMessage
			if err := json.Unmarshal(diagRaw, &perCheckpoint); err != nil {
				errs = append(errs, "CALIBRATION_DIAGNOSTICS.json: "+err.Error())
			} else {
				for _, checkpointDiag := range perCheckpoint {
					var counts map[string]json.RawMessage
					if err := json.Unmarshal(checkpointDiag["generator_observation_counts"], &counts); err != nil {
						continue
					}
					for gen := range counts {
						if !names[gen] {
							errs = append(errs, fmt.Sprintf("CALIBRATION_DIAGNOSTICS.json references generator %q with no matching CALIBRATION_GENERATORS/%s.json", gen, gen))
						}
					}
				}
			}
		}
	}

	// Rarefaction protocol <-> schema: the schema document must describe
	// the exact frozen TSV header the live writer emits.
	schemaRaw, err := os.ReadFile(filepath.Join(base, "RAREFACTION_SCHEMA.md"))
	if err != nil {
		errs = append(errs, "RAREFACTION_SCHEMA.md: "+err.Error())
	} else {
		var header strings.Builder
		_ = WriteRarefactionTSV(&header, nil)
		cols := strings.Split(strings.TrimSpace(strings.SplitN(header.String(), "\n", 2)[0]), "\t")
		for _, c := range cols {
			if !strings.Contains(string(schemaRaw), c) {
				errs = append(errs, fmt.Sprintf("RAREFACTION_SCHEMA.md does not document rarefaction column %q", c))
			}
		}
	}

	sort.Strings(errs)
	return errs, nil
}
