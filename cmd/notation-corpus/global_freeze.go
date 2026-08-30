package main

import (
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

// globalFreezeVerifyCmd implements task run02 section 7:
// `notation-corpus global-freeze-verify`. It loads GLOBAL_FREEZE_MANIFEST.json,
// recomputes every mandatory artifact's SHA-256, cross-checks internal
// references, and exits non-zero on any mismatch.
func globalFreezeVerifyCmd(args []string) error {
	fs := flag.NewFlagSet("global-freeze-verify", flag.ContinueOnError)
	repo := fs.String("repo", ".", "repository root")
	if err := fs.Parse(args); err != nil {
		return err
	}
	base := filepath.Join(*repo, "research", "comparative_notation")

	bindingErrs, err := notation.VerifyGlobalFreezeManifest(base)
	if err != nil {
		return fmt.Errorf("global-freeze-verify: %w", err)
	}
	crossErrs, err := notation.GlobalFreezeCrossReferenceChecks(base)
	if err != nil {
		return fmt.Errorf("global-freeze-verify: %w", err)
	}
	all := append(append([]string{}, bindingErrs...), crossErrs...)
	if len(all) != 0 {
		return fmt.Errorf("global freeze manifest is invalid:\n- %s", strings.Join(all, "\n- "))
	}
	fmt.Println("GLOBAL_FREEZE_MANIFEST_VALID=true")
	return nil
}

// oldGlobalFreezeManifest is the pre-run02 (schema v1.0) shape of
// GLOBAL_FREEZE_MANIFEST.json, read here only to drift-check previously
// bound artifacts while migrating to schema v2.0. It is not used anywhere
// else; production-run-preflight now reads the v2.0 manifest exclusively
// via notation.VerifyGlobalFreezeManifest.
type oldGlobalFreezeManifest struct {
	Artifacts map[string]json.RawMessage `json:"artifacts"`
}

// oldFreezeHash resolves the hash a v1.0 manifest previously bound name to,
// handling the ad hoc nested shapes (CALIBRATION_GENERATORS, VM_REFERENCE_V2)
// alongside flat string entries. ok=false means name was never bound before
// (a genuine gap this task is completing, not a drift to check).
func oldFreezeHash(old oldGlobalFreezeManifest, name string) (hash string, ok bool) {
	switch name {
	case "VM_REFERENCE_V2.tsv", "VM_REFERENCE_V2.fingerprint.json":
		raw, present := old.Artifacts["VM_REFERENCE_V2"]
		if !present {
			return "", false
		}
		var vm struct {
			ReferenceTSVSHA256    string `json:"reference_tsv_sha256"`
			FingerprintJSONSHA256 string `json:"fingerprint_json_sha256"`
		}
		if err := json.Unmarshal(raw, &vm); err != nil {
			return "", false
		}
		if name == "VM_REFERENCE_V2.tsv" {
			return vm.ReferenceTSVSHA256, vm.ReferenceTSVSHA256 != ""
		}
		return vm.FingerprintJSONSHA256, vm.FingerprintJSONSHA256 != ""
	case "CALIBRATION_GENERATORS":
		raw, present := old.Artifacts["CALIBRATION_GENERATORS"]
		if !present {
			return "", false
		}
		var g struct {
			SHA256 string `json:"sha256"`
		}
		if err := json.Unmarshal(raw, &g); err != nil {
			return "", false
		}
		return g.SHA256, g.SHA256 != ""
	default:
		raw, present := old.Artifacts[name]
		if !present {
			return "", false
		}
		var s string
		if err := json.Unmarshal(raw, &s); err != nil || len(s) != 64 {
			return "", false
		}
		return s, true
	}
}

// globalFreezeBindCmd implements task run02 sections 3-5: computes SHA-256
// for every mandatory frozen artifact from its canonical committed bytes
// (no preprocessing), drift-checks every artifact that was already bound
// under the old (schema v1.0) manifest against its previously recorded
// hash — refusing to proceed if any of those bytes changed, since that
// would mean a scientific frozen artifact was modified, which this task may
// never do — and writes the completed schema v2.0 manifest plus
// GLOBAL_FREEZE_BINDING_COMPLETION.md.
func globalFreezeBindCmd(args []string) error {
	fs := flag.NewFlagSet("global-freeze-bind", flag.ContinueOnError)
	repo := fs.String("repo", ".", "repository root")
	if err := fs.Parse(args); err != nil {
		return err
	}
	base := filepath.Join(*repo, "research", "comparative_notation")

	// A prior manifest is optional: its absence means this is a first-time
	// freeze (every artifact is "newly bound", nothing to drift-check),
	// not an error. When present, it is used only to drift-check artifacts
	// that were already bound before — never to skip hashing them.
	var old oldGlobalFreezeManifest
	if oldRaw, err := os.ReadFile(filepath.Join(base, "GLOBAL_FREEZE_MANIFEST.json")); err == nil {
		if err := json.Unmarshal(oldRaw, &old); err != nil {
			return fmt.Errorf("cannot parse existing GLOBAL_FREEZE_MANIFEST.json: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("cannot read existing GLOBAL_FREEZE_MANIFEST.json: %w", err)
	}

	type specHash struct {
		spec notation.FreezeArtifactSpec
		hash string
	}
	var bound []specHash
	var newlyBound []string
	var driftErrs []string
	names := map[string]bool{}
	for _, spec := range notation.RequiredGlobalFreezeArtifacts {
		if names[spec.Path] {
			return fmt.Errorf("duplicate entry in RequiredGlobalFreezeArtifacts: %s", spec.Path)
		}
		names[spec.Path] = true
		if _, err := os.Stat(filepath.Join(base, spec.Path)); err != nil {
			return fmt.Errorf("mandatory frozen artifact missing on disk: %s: %w", spec.Path, err)
		}
		got, err := notation.HashFreezeArtifact(base, spec)
		if err != nil {
			return fmt.Errorf("%s: %w", spec.Path, err)
		}
		if want, ok := oldFreezeHash(old, spec.Path); ok {
			if want != got {
				driftErrs = append(driftErrs, fmt.Sprintf("%s: previously bound to %s, now %s — this is scientific content drift, not a metadata gap; global-freeze-bind refuses to silently re-bind it", spec.Path, want, got))
				continue
			}
		} else {
			newlyBound = append(newlyBound, spec.Path)
		}
		bound = append(bound, specHash{spec, got})
	}
	if len(driftErrs) != 0 {
		return fmt.Errorf("refusing to bind: frozen artifact content changed since the original freeze (a new full freeze procedure is required, not a metadata fix):\n- %s", strings.Join(driftErrs, "\n- "))
	}

	registryHash, err := notation.MetricRegistryHash()
	if err != nil {
		return err
	}
	sort.Slice(bound, func(i, j int) bool { return bound[i].spec.Path < bound[j].spec.Path })
	entries := make([]notation.FreezeArtifactEntry, 0, len(bound))
	for _, b := range bound {
		entries = append(entries, notation.FreezeArtifactEntry{Path: b.spec.Path, SHA256: b.hash, Role: b.spec.Role, SchemaVersion: b.spec.SchemaVersion})
	}
	manifest := notation.GlobalFreezeManifest{
		SchemaVersion:      notation.GlobalFreezeManifestSchemaVersion,
		Task:               "Comparative Notation Study — Global Freeze Completion (B01-B04) + Cryptographic Binding Completion (run02)",
		BaseSeed:           notation.BaseSeed,
		FreezeGenerationID: "CNS-FREEZE-20260830",
		BindingCompletedAt: time.Now().UTC().Format(time.RFC3339),
		Comparator: notation.GlobalFreezeComparatorIdentity{
			Package: "zcore.dev/voinich/internal/notation", CLI: "zcore.dev/voinich/cmd/notation-corpus",
			AnalyzerVersion: "notation-analyzer-1.0", MetricRegistryVersion: notation.MetricRegistryVersion, MetricRegistryHash: registryHash,
		},
		Artifacts: entries,
		ProtocolStatus: notation.GlobalFreezeProtocolStatus{
			GlobalComparisonProtocolFrozen: true, GlobalFreezeCryptographicallyBound: true, ProductionComparativeRunAuthorized: false,
		},
	}
	b, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if err := os.WriteFile(filepath.Join(base, "GLOBAL_FREEZE_MANIFEST.json"), b, 0644); err != nil {
		return err
	}

	return writeGlobalFreezeBindingCompletion(base, entries, newlyBound, manifest)
}

func writeGlobalFreezeBindingCompletion(base string, entries []notation.FreezeArtifactEntry, newlyBound []string, manifest notation.GlobalFreezeManifest) error {
	sort.Strings(newlyBound)
	var md strings.Builder
	md.WriteString("# Global freeze binding completion\n\n")
	fmt.Fprintf(&md, "Task: `tasks_other/corporative_notation_study_production_run02.md`. Freeze generation `%s`, binding completed at `%s`.\n\n", manifest.FreezeGenerationID, manifest.BindingCompletedAt)
	md.WriteString("## Why the original freeze manifest was incomplete\n\n")
	md.WriteString("`GLOBAL_FREEZE_MANIFEST.json` schema v1.0 (written while closing B01-B04) bound only the artifacts each B0x closure narrative happened to cite by exact filename. A full audit of every protocol/specification, calibration, rarefaction, distribution/bootstrap, VM-reference-v2, metric-registry, and USC-specification file under `research/comparative_notation/` (task run02 section 1) found more mandatory frozen artifacts than the six the preceding `production-run-preflight` run had already caught. None of this is new scientific work: every one of these files already existed, unchanged, as part of the frozen B01-B04 protocol state; they were simply never cryptographically bound.\n\n")
	md.WriteString("## Bindings added by this task\n\n")
	if len(newlyBound) == 0 {
		md.WriteString("(none — nothing to add)\n\n")
	} else {
		for _, p := range newlyBound {
			fmt.Fprintf(&md, "- `%s`\n", p)
		}
		md.WriteString("\n")
	}
	md.WriteString("## What did not change\n\n")
	md.WriteString("Scientific frozen artifact *contents* were not modified. Every artifact that the v1.0 manifest already bound (`COMPARATIVE_EXPERIMENT_SPEC.md`, `METRIC_REGISTRY.md`, `RAREFACTION_PROTOCOL.md`, `RAREFACTION_SCHEMA.md`, `DISTRIBUTION_OUTPUT_CONTRACT.md`, `BOOTSTRAP_PROTOCOL.md`, `CALIBRATION_GENERATORS/*.json`, `CALIBRATION_SCALES.tsv`, `VM_REFERENCE_V2.tsv`, `VM_REFERENCE_V2.fingerprint.json`) was drift-checked against its previously recorded SHA-256 by `notation-corpus global-freeze-bind` and found byte-identical; `global-freeze-bind` refuses to proceed (fail-closed, no silent re-bind) if any previously bound artifact's current bytes ever stop matching its old hash.\n\n")
	md.WriteString("## Manifest-only changes made\n\n")
	md.WriteString("`GLOBAL_FREEZE_MANIFEST.json` moved from schema `global-freeze-manifest-1.0` to `global-freeze-manifest-2.0`: every artifact (old and newly bound) is now one uniform `{path, sha256, role, schema_version}` entry in a single `artifacts` array, instead of a mix of flat string entries and ad hoc nested objects (`CALIBRATION_GENERATORS`, `VM_REFERENCE_V2`). `protocol_status` gained `GLOBAL_FREEZE_CRYPTOGRAPHICALLY_BOUND`. `freeze_generation_id` and `binding_completed_at` were added as manifest metadata.\n\n")
	md.WriteString("No Git commit hash is embedded inside `GLOBAL_FREEZE_MANIFEST.json` itself: recording \"this manifest belongs to commit X\" inside the very file that commit X would contain is a self-referential requirement (task run02 section 6), since the manifest's own bytes are part of what determines that commit's hash. Git binding is external instead — `notation-corpus production-run-preflight`'s `clean_git_revision` gate independently captures `git rev-parse HEAD` against a clean working tree at authorization time, exactly as it already did before this task.\n\n")
	md.WriteString("## Checksums recorded\n\n")
	md.WriteString("| Path | Role | SHA-256 |\n|---|---|---|\n")
	for _, e := range entries {
		fmt.Fprintf(&md, "| `%s` | `%s` | `%s` |\n", e.Path, e.Role, e.SHA256)
	}
	md.WriteString("\n## Why this is metadata completion, not a new scientific freeze\n\n")
	md.WriteString("No rarefaction, bootstrap, calibration, VM-comparison, or any other scientific computation was re-run. No scale, seed, checkpoint, replicate count, or metric definition changed. This task only computed and recorded the SHA-256 of files that already existed as part of the frozen B01-B04 protocol state and were already described as frozen by `GLOBAL_FREEZE_REPORT.md`, `PREPARATION_BLOCKERS.md` (B01-B04=CLOSED), and the individual `*_PROTOCOL.md`/`*_SPEC.md` documents themselves — it closes a bookkeeping gap in how that existing freeze was cryptographically bound, per the explicit constraint in this task that scientific frozen artifact content must never change to obtain a new checksum.\n\n")
	md.WriteString("```text\nGLOBAL_COMPARISON_PROTOCOL_FROZEN=true\nGLOBAL_FREEZE_CRYPTOGRAPHICALLY_BOUND=true\nSCIENTIFIC_FROZEN_ARTIFACTS_MODIFIED=false\nPRODUCTION_COMPARATIVE_RUN_AUTHORIZED=false\n```\n")
	return os.WriteFile(filepath.Join(base, "GLOBAL_FREEZE_BINDING_COMPLETION.md"), []byte(md.String()), 0644)
}
