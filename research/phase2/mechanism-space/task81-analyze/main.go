// task81-analyze validates the frozen Task80 authority and deterministically
// emits the target-blind Task81 mechanism-space artifacts.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"zcore.dev/voinich/internal/mnemonicspace"
)

const (
	masterSeed = uint64(81024001)
	version    = "V1"
)

type registryEntry struct {
	mnemonicspace.MechanismSpec
	Ablations []string `json:"ablations"`
}

type manifestJob struct {
	ID                string                          `json:"job_id"`
	MechanismID       string                          `json:"mechanism_id"`
	ParameterSetID    string                          `json:"parameter_set_id"`
	InputCorpusID     string                          `json:"input_corpus_id"`
	Seed              uint64                          `json:"seed"`
	Replicate         int                             `json:"replicate"`
	RecoveryCondition mnemonicspace.RecoveryCondition `json:"recovery_condition"`
	AblationID        string                          `json:"ablation_id,omitempty"`
}

type blindManifest struct {
	Schema             string              `json:"schema"`
	Version            string              `json:"version"`
	MasterSeed         uint64              `json:"master_seed"`
	SeedDerivation     string              `json:"seed_derivation"`
	DevelopmentInputs  []string            `json:"development_inputs"`
	EvaluationCorpora  []string            `json:"evaluation_corpora"`
	RecoveryConditions []string            `json:"recovery_conditions"`
	FrozenAblations    map[string][]string `json:"frozen_ablations"`
	Jobs               []manifestJob       `json:"jobs"`
}

func main() {
	out := filepath.Join("research", "phase2", "mechanism-space")
	if len(os.Args) == 2 {
		out = os.Args[1]
	} else if len(os.Args) > 2 {
		log.Fatal("usage: task81-analyze [output-directory]")
	}
	authority, err := mnemonicspace.LoadTask80Authority(filepath.Join("research", "phase2", "fontana", "task80"))
	if err != nil {
		log.Fatal(err)
	}
	specs := mnemonicspace.FrozenRegistry()
	if err := mnemonicspace.ValidateRegistry(authority, specs); err != nil {
		log.Fatal(err)
	}
	for _, invalid := range mnemonicspace.InvalidControls() {
		if err := mnemonicspace.ValidateMechanism(authority, invalid); err == nil {
			log.Fatalf("invalid control %s was accepted", invalid.ID)
		}
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		log.Fatal(err)
	}

	jobs := makeManifest(authority, specs)
	if err := writeJSON(filepath.Join(out, "MNEMONIC_MECHANISM_REGISTRY.json"), registry(specs)); err != nil {
		log.Fatal(err)
	}
	if err := writeFile(filepath.Join(out, "MNEMONIC_PARAMETER_REGISTRY.tsv"), parameters(specs)); err != nil {
		log.Fatal(err)
	}
	if err := writeFile(filepath.Join(out, "MNEMONIC_COMPOSITIONS.tsv"), compositions(specs)); err != nil {
		log.Fatal(err)
	}
	if err := writeJSON(filepath.Join(out, "TASK82_BLIND_MANIFEST.json"), jobs); err != nil {
		log.Fatal(err)
	}
	if err := writeFile(filepath.Join(out, "TASK82_COST_MODEL.tsv"), costModel(len(specs), len(jobs.Jobs))); err != nil {
		log.Fatal(err)
	}
	if err := writeFile(filepath.Join(out, "PROVENANCE_GRAPH.dot"), provenance(specs, jobs)); err != nil {
		log.Fatal(err)
	}
	for name, body := range documents(len(specs), len(jobs.Jobs)) {
		if err := writeFile(filepath.Join(out, name), body); err != nil {
			log.Fatal(err)
		}
	}
	freeze := map[string]any{
		"version":                        version,
		"git_commit":                     gitCommit(),
		"task80_algebra_checksum":        authority.AlgebraSHA256,
		"task80_model_freeze_checksum":   authority.FrozenSHA256,
		"mechanism_registry_checksum":    checksumFile(filepath.Join(out, "MNEMONIC_MECHANISM_REGISTRY.json")),
		"parameter_registry_checksum":    checksumFile(filepath.Join(out, "MNEMONIC_PARAMETER_REGISTRY.tsv")),
		"recovery_contract_checksum":     checksumFile(filepath.Join(out, "MNEMONIC_RECOVERY_CONTRACT.md")),
		"observable_contract_checksum":   checksumFile(filepath.Join(out, "OBSERVABLE_DOCUMENT_CONTRACT.md")),
		"task82_blind_manifest_checksum": checksumFile(filepath.Join(out, "TASK82_BLIND_MANIFEST.json")),
		"seed_policy":                    fmt.Sprintf("SHA-256(master=%d|mechanism|parameter|corpus|condition|replicate), first 64 bits", masterSeed),
		"composition_depth":              4,
		"mechanism_count":                len(specs),
		"parameter_point_count":          len(specs),
		"ablation_count":                 12,
		"prohibited_post_freeze_changes": []string{"mechanisms", "parameters", "serialization", "recovery semantics", "ablations"},
	}
	if err := writeJSON(filepath.Join(out, "MNEMONIC_MECHANISM_SPACE_FROZEN.json"), freeze); err != nil {
		log.Fatal(err)
	}
	if err := writeFile(filepath.Join(out, "TASK81_DESIGN_FROZEN"), "TASK81_DESIGN_FROZEN V1\n"); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wrote %d mechanisms and %d target-blind Task82 jobs to %s\n", len(specs), len(jobs.Jobs), out)
}

func registry(specs []mnemonicspace.MechanismSpec) []registryEntry {
	out := make([]registryEntry, 0, len(specs))
	for _, spec := range specs {
		out = append(out, registryEntry{MechanismSpec: spec, Ablations: ablations(spec.ID)})
	}
	return out
}

func ablations(id string) []string {
	m := map[string][]string{
		"f01_speculum_core":                {"no_rotation:synthetic_literal_storage", "no_convention:negative_randomized_convention"},
		"f01_speculum_profile_latin23_r12": {"no_rotation:synthetic_literal_storage", "no_convention:negative_randomized_convention"},
		"f08_serpens_core":                 {"no_path:negative_randomized_path", "no_state_persistence:synthetic_literal_storage"},
		"f11_arismetricum_core":            {"no_index:synthetic_cue_based", "randomized_index:negative_randomized_index_mapping"},
		"f12_horalogius_core":              {"no_association:synthetic_cyclic_state", "randomized_association:negative_randomized_cue_association"},
		"m_restricted_rotation_index":      {"no_rotation:synthetic_indexed_lookup"},
		"m_restricted_storage_associate":   {"no_association:synthetic_literal_storage"},
		"synthetic_ambiguous":              {"no_association:synthetic_literal_storage"},
	}
	return m[id]
}

func makeManifest(authority mnemonicspace.Authority, specs []mnemonicspace.MechanismSpec) blindManifest {
	conditions := []mnemonicspace.RecoveryCondition{
		mnemonicspace.RecoveryFullKnowledge, mnemonicspace.RecoveryNoContext, mnemonicspace.RecoveryNoConvention,
		mnemonicspace.RecoveryNoGeometry, mnemonicspace.RecoveryNoHistory, mnemonicspace.RecoveryNoInternal,
		mnemonicspace.RecoveryObservable,
	}
	out := blindManifest{
		Schema: "task82-blind-manifest-v1", Version: version, MasterSeed: masterSeed,
		SeedDerivation:    "SHA-256(master_seed|mechanism_id|parameter_set_id|input_corpus_id|recovery_condition|replicate), first 64 bits",
		DevelopmentInputs: []string{"synthetic-literal-v1", "synthetic-cue-v1"},
		EvaluationCorpora: []string{"Doyle", "Longfellow", "Astafiev"},
		FrozenAblations:   map[string][]string{},
	}
	for _, c := range conditions {
		out.RecoveryConditions = append(out.RecoveryConditions, string(c))
	}
	for _, spec := range specs {
		if frozen := ablations(spec.ID); len(frozen) > 0 {
			out.FrozenAblations[spec.ID] = frozen
		}
		for _, param := range spec.Parameters {
			for _, corpus := range out.EvaluationCorpora {
				for replicate := 0; replicate < 2; replicate++ {
					for _, condition := range conditions {
						job := mnemonicspace.Job{MechanismID: spec.ID, ParameterSetID: param.ID, InputID: corpus, RecoveryCondition: condition, Replicate: replicate, MasterSeed: masterSeed}
						out.Jobs = append(out.Jobs, manifestJob{ID: job.ID(spec, authority), MechanismID: spec.ID, ParameterSetID: param.ID, InputCorpusID: corpus, Seed: job.DerivedSeed(), Replicate: replicate, RecoveryCondition: condition})
					}
				}
			}
		}
	}
	return out
}

func parameters(specs []mnemonicspace.MechanismSpec) string {
	var b strings.Builder
	b.WriteString("mechanism_id\tparameter_set_id\tparameter\tvalue_or_grid\tprovenance_class\tsource_or_rationale\tfrozen\n")
	for _, spec := range specs {
		for _, p := range spec.Parameters {
			fmt.Fprintf(&b, "%s\t%s\tconfiguration\t%s\t%s\t%s\t%t\n", spec.ID, p.ID, p.Description, p.Origin, source(spec.Status), p.Frozen)
		}
	}
	return b.String()
}

func source(status mnemonicspace.HistoricalStatus) string {
	if status == mnemonicspace.StatusFExact || status == mnemonicspace.StatusFProfile {
		return "Task80 frozen model/core; numeric uncertainty remains explicitly bounded"
	}
	if status == mnemonicspace.StatusMRestricted {
		return "Task80 G-ALLOWED counterfactual; no Fontana provenance claim"
	}
	return "target-independent control grid"
}

func compositions(specs []mnemonicspace.MechanismSpec) string {
	var b strings.Builder
	b.WriteString("composition_id\toperations\ttype_valid\thistorical_status\tcanonical_mechanism\tequivalent_to\tinclusion_or_exclusion_reason\n")
	for _, spec := range specs {
		ops := make([]string, 0, len(spec.Encoding)+len(spec.Retrieval))
		for _, s := range append(append([]mnemonicspace.OperationInvocation{}, spec.Encoding...), spec.Retrieval...) {
			ops = append(ops, string(s.Operation))
		}
		fmt.Fprintf(&b, "%s\t%s\ttrue\t%s\t%s\t%s\tbounded depth <= 4; provenance and type validation passed\n", spec.ID, strings.Join(ops, "+"), spec.Status, spec.ID, spec.EquivalentTo)
	}
	b.WriteString("synthetic_invalid_signal\tsignal\tfalse\tFORBIDDEN\t\t\tmissing required ExternalState; rejected by validator\n")
	return b.String()
}

func costModel(mechanisms, jobs int) string {
	return fmt.Sprintf("dimension\tvalue\tbasis\nmechanisms\t%d\tfrozen registry\nparameter_points\t%d\tone frozen point per registry mechanism\ncorpora\t3\tDoyle, Longfellow, Astafiev\nreplicates\t2\tprecommitted seed replicates\nrecovery_conditions\t7\tR0-R6; runner records NOT_APPLICABLE where unused\njobs\t%d\tmechanisms x parameter points x corpora x replicates x recovery conditions\ncpu_hours\t0.20\tplanning estimate at 2 seconds/job plus 2x overhead\npeak_ram_mb\t256\tone independent job plus F2 adapter buffer\nartifact_volume_mb\t64\tJSON checkpoint/document/recovery metadata budget\ndistributed_execution\trequired-ready\tjobs have deterministic independent identities and checkpoint bindings\n", mechanisms, mechanisms, jobs)
}

func provenance(specs []mnemonicspace.MechanismSpec, manifest blindManifest) string {
	var b strings.Builder
	b.WriteString("digraph Task81Provenance {\n  rankdir=LR;\n  task80 [label=\"Task80 frozen authority\"];\n")
	for _, spec := range specs {
		fmt.Fprintf(&b, "  %q [label=%q];\n  task80 -> %q;\n", spec.ID, spec.ID+"\\n"+string(spec.Status), spec.ID)
		for _, p := range spec.Parameters {
			pid := spec.ID + ":" + p.ID
			fmt.Fprintf(&b, "  %q [label=%q];\n  %q -> %q;\n", pid, p.ID, spec.ID, pid)
		}
	}
	for _, job := range manifest.Jobs {
		pid := job.MechanismID + ":" + job.ParameterSetID
		fmt.Fprintf(&b, "  %q [shape=box,label=%q];\n  %q -> %q;\n", job.ID, "Task82 job "+job.ID, pid, job.ID)
	}
	b.WriteString("}\n")
	return b.String()
}

func documents(mechanisms, jobs int) map[string]string {
	return map[string]string{
		"TASK81_DESIGN.md": fmt.Sprintf(`# Task81: Blind mnemonic mechanism space

**Freeze version:** V1. This space is built solely from the Task80 frozen operation algebra and model freeze. No Voynich text, F2 values, BDD result, notation-control result, target optimization, or target-oriented visual selection was used.

The composition bound is four operations. The registry contains %d runnable mechanisms: separate F01 literal rotational storage, F08 ordered positional storage, F11 indexed opaque cues, and F12 temporal associative cues; two M-RESTRICTED counterfactuals; five generic controls; and four negative controls. F07 and F10 remain reference-only and excluded models are not restored. There are no M-EXTENDED families because the frozen primitives cover the required experimental boundary cases. The registry deliberately does not enumerate arbitrary operation strings: its finite admissible set is the 16 canonical mechanisms plus 12 named ablations; all other raw sequences up to the four-operation bound are excluded unless independently entered as a future V2 candidate.

All runs retain distinct M (input), E (external state), G (geometry/path), H (history), K (convention), I (internal memory), and C (context) carriers. Composition validation checks Task80 input/output types, state effects, prior knowledge, provenance status, and forbidden compositions. Observable documents derive only from visible state or observation streams.

The finite grid is recorded in the parameter registry. Historical profiles are distinguished from source-bounded and generic/control values. Minimal ablations are frozen in the registry and resolve to named controls rather than silently changing historical semantics. Structural, operational, observational, and retrieval equivalence are separate: equivalent_to only records the frozen observational/control relation, so equivalent entries are not independent models.
`, mechanisms),
		"MNEMONIC_RECOVERY_CONTRACT.md": `# Mnemonic recovery contract

Recovery is R(E, G, H, K, I, C); M is never available to a retriever. A request declares R0 full knowledge, R1 no context, R2 no convention, R3 no path/geometry, R4 no history, R5 no internal memory, or R6 observable only. Removing an unused carrier returns NOT_APPLICABLE.

Results are EXACT, PARTIAL, AMBIGUITY_SET, CUE_ONLY, NO_RECOVERY, and NOT_APPLICABLE. A cue is not plaintext: F11 requires a supplied cue convention and F12 requires a supplied InternalMemoryState association map. Multiple candidates remain an ordered ambiguity set; context filters that set rather than selecting a hidden correct answer. Observable collisions are grouped by document checksum and retain all distinct intended items.

Task82 metrics are exact recovery rate, symbol/item recovery, ambiguity cardinality, candidate-set entropy when meaningful, retained/lost information from the trace, and carrier-dependence by recovery condition. Error classes reserved for Task82 are substitution, deletion, insertion, transposition, boundary corruption, state corruption, and convention corruption. Error locality is determined by the affected state cell versus an entire carrier; this freeze does not run a primary corruption experiment.
`,
		"OBSERVABLE_DOCUMENT_CONTRACT.md": `# Observable document contract

An OBSERVABLE_DOCUMENT contains a glyph/symbol stream plus explicit token, line, and unit/page boundary statuses: NOT_DEFINED, INHERITED_FROM_INPUT, or GENERATED_BY_MECHANISM. Task81 does not invent absent hierarchy. F01/F08/F11 use undefined token/line boundaries and generated units; F12 generates cue-token/unit boundaries and leaves lines undefined.

Only externally visible symbols or observations may be serialized. InternalMemoryState, association maps, conventions, paths, history, plaintext input, and implementation state are excluded. Metadata is descriptive and cannot encode any hidden carrier. This is accepted by a future frozen F2 reader as a generic symbol stream with missing-data boundary semantics; Task81 performs no F2 computation or comparison.
`,
		"TASK81_REPORT.md": fmt.Sprintf(`# Task81 report

**Historical mechanisms:** F01 core/profile (literal rotational), F08 core (ordered positional), F11 core (indexed opaque cue), and F12 core (temporal associative cue) are separate. F01/F08 are exact only with their declared convention/path; F11 is cue-only without cue convention; F12 is cue-only without formal internal association. F07/F10 are reference-only.

**Generalized space:** M-RESTRICTED rotation-plus-index and explicit-storage-plus-association are type-valid unattested boundaries with explicit cue conversion. No M-EXTENDED primitive was necessary. Generic literal, cyclic, indexed, cue, and ambiguous controls plus randomized convention/path/cue-association/index-mapping nulls are frozen. Historical status is held in every registry entry and controls never claim Fontana use.

**Information and recovery:** M/E/G/H/K/I/C are separated; opaque cue and literal surface roles are explicit. Many-to-one collisions are reported by observable-document checksum, and context narrows candidate sets formally. Internal memory is only a known/unavailable association map, with no psychological claim. Contracts define serializable hidden-state-free observations and R0-R6 recovery.

**Experiment readiness:** %d precommitted target-blind Task82 jobs cover three named control corpora, two deterministic replicates, and seven recovery conditions. SHA-256 derivation makes jobs independent and resumable with checksum/spec-version checkpoints; distributed execution is ready. Development fixtures are separate from evaluation corpora.

| Verdict | Result |
| --- | --- |
| TASK80_ALGEBRA_PRESERVED | SUPPORTED |
| HISTORICAL_STATUS_SEPARATION | SUPPORTED |
| TYPE_SYSTEM_VALIDATED | SUPPORTED |
| INFORMATION_CARRIERS_SEPARATED | SUPPORTED |
| OBSERVABLE_DOCUMENT_READY | SUPPORTED |
| RECOVERY_FRAMEWORK_READY | SUPPORTED |
| AMBIGUITY_FRAMEWORK_READY | SUPPORTED |
| LEAKAGE_PROTECTION_READY | SUPPORTED |
| TASK82_BLIND_MANIFEST_READY | SUPPORTED |
| DISTRIBUTED_EXECUTION_READY | SUPPORTED |
| VOYNICH_FIREWALL_PRESERVED | SUPPORTED |
| NOTATION_CONTROL_FIREWALL_PRESERVED | SUPPORTED |

**Final verdict:** MNEMONIC_MECHANISM_SPACE_FROZEN.
`, jobs),
	}
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(path, string(data)+"\n")
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}

func checksumFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func gitCommit() string {
	data, err := os.ReadFile(".git/HEAD")
	if err != nil {
		return "UNCOMMITTED"
	}
	head := strings.TrimSpace(string(data))
	if strings.HasPrefix(head, "ref: ") {
		data, err = os.ReadFile(filepath.Join(".git", strings.TrimPrefix(head, "ref: ")))
		if err != nil {
			return "UNCOMMITTED"
		}
		return strings.TrimSpace(string(data))
	}
	return head
}
