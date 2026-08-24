package task82a

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"zcore.dev/voinich/internal/mnemonicspace"
)

// ManifestJob is one immutable Task82a job (task82a.txt sec.30).
type ManifestJob struct {
	JobID            string `json:"job_id"`
	MechanismID      string `json:"mechanism_id"`
	MechanismVersion string `json:"mechanism_version"`
	ParameterSetID   string `json:"parameter_set_id"`
	ScalingPolicyID  string `json:"scaling_policy_id"`
	InputCorpusID    string `json:"input_corpus_id"`
	CorpusScale      string `json:"corpus_scale"`
	Chunks           int    `json:"chunks"`
	Seed             uint64 `json:"seed"`
	Replicate        int    `json:"replicate"`
	ControlStatus    string `json:"ablation_control_status"`
}

func derivedSeed(j ManifestJob) uint64 {
	h := sha256.Sum256(fmt.Appendf(nil, "%d|%s|%s|%s|%s|%s|%d", MasterSeed, j.MechanismID, j.ScalingPolicyID, j.InputCorpusID, j.CorpusScale, j.MechanismVersion, j.Replicate))
	return binary.BigEndian.Uint64(h[:8])
}

func jobID(j ManifestJob) string {
	payload := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%d|%d", j.MechanismID, j.MechanismVersion, j.ParameterSetID, j.ScalingPolicyID, j.InputCorpusID, j.CorpusScale, j.Replicate, j.Seed)
	h := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(h[:16])
}

// Manifest is Task82a's frozen blind job list (task82a.txt sec.31).
type Manifest struct {
	Schema     string        `json:"schema"`
	Version    string        `json:"version"`
	MasterSeed uint64        `json:"master_seed"`
	Corpora    []string      `json:"evaluation_corpora"`
	Scales     []ScaleSpec   `json:"corpus_scales"`
	Replicates int           `json:"replicates"`
	Jobs       []ManifestJob `json:"jobs"`
}

func controlStatus(id string) string {
	switch {
	case strings.HasPrefix(id, "negative_"):
		return "NEGATIVE_CONTROL"
	case strings.HasPrefix(id, "synthetic_"):
		return "GENERIC_CONTROL"
	case strings.HasPrefix(id, "m_restricted_"):
		return "M_RESTRICTED_COUNTERFACTUAL"
	default:
		return "FULL_MECHANISM"
	}
}

// BuildManifest enumerates every frozen job cell: mechanism x its frozen
// scaling policies x corpus x corpus_scale x replicate. Ordering is
// deterministic (sorted mechanism/policy/corpus/scale) so the manifest and
// its checksum are reproducible from this source alone.
func BuildManifest() Manifest {
	reg := mnemonicspace.FrozenRegistry()
	sort.Slice(reg, func(i, j int) bool { return reg[i].ID < reg[j].ID })
	corpora := make([]string, 0, len(CorpusPaths))
	for id := range CorpusPaths {
		corpora = append(corpora, id)
	}
	sort.Strings(corpora)
	m := Manifest{Schema: "task82a-blind-manifest-v1", Version: FreezeVersion, MasterSeed: MasterSeed, Corpora: corpora, Scales: CorpusScales, Replicates: Replicates}
	for _, spec := range reg {
		scale, ok := scaleFor(spec.ID)
		if !ok {
			continue
		}
		for _, policy := range PoliciesFor(scale.Surface) {
			for _, corpusID := range corpora {
				for _, sc := range CorpusScales {
					for rep := 0; rep < Replicates; rep++ {
						j := ManifestJob{
							MechanismID: spec.ID, MechanismVersion: spec.Version, ParameterSetID: scale.ParameterSetID,
							ScalingPolicyID: policy, InputCorpusID: corpusID, CorpusScale: sc.ID, Chunks: sc.Chunks,
							Replicate: rep, ControlStatus: controlStatus(spec.ID),
						}
						j.Seed = derivedSeed(j)
						j.JobID = jobID(j)
						m.Jobs = append(m.Jobs, j)
					}
				}
			}
		}
	}
	sort.Slice(m.Jobs, func(i, j int) bool { return m.Jobs[i].JobID < m.Jobs[j].JobID })
	return m
}

// requiredLetters/requiredItems compute the matched input length every
// source corpus must provide: enough for the largest frozen corpus_scale,
// for every mechanism's capacity, with no per-corpus slack (task82a.txt
// sec.28: matched input size across corpora).
func requiredLengths() (letters, items int) {
	maxChunks := 0
	for _, sc := range CorpusScales {
		if sc.Chunks > maxChunks {
			maxChunks = sc.Chunks
		}
	}
	maxLiteralCap, maxCueCap := 0, 0
	for _, s := range MechanismScales {
		if s.Surface == SurfaceLiteral && s.Capacity > maxLiteralCap {
			maxLiteralCap = s.Capacity
		}
		if s.Surface == SurfaceCue && s.Capacity > maxCueCap {
			maxCueCap = s.Capacity
		}
	}
	return maxLiteralCap * maxChunks, maxCueCap * maxChunks
}

func outDir(root string) string { return filepath.Join(root, "research", "phase2", "task82a") }

// GenerateManifest writes TASK82A_BLIND_MANIFEST.json, SCALING_POLICIES.tsv,
// BOUNDARY_PROVENANCE.tsv, and TASK82A_COST_MODEL.tsv (task82a.txt
// sec.31-32, 69). It must run before any main-generation job.
func GenerateManifest(root string) error {
	m := BuildManifest()
	dir := outDir(root)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	manifestPath := filepath.Join(dir, "TASK82A_BLIND_MANIFEST.json")
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		return err
	}
	if err := writeScalingPolicies(dir); err != nil {
		return err
	}
	if err := writeBoundaryProvenance(dir); err != nil {
		return err
	}
	if err := writeCostModel(dir, m); err != nil {
		return err
	}
	return nil
}

func writeScalingPolicies(dir string) error {
	var b strings.Builder
	b.WriteString("policy_id\taxis\tvalue\tapplies_to_surface\tincluded\treason\n")
	rows := [][6]string{
		{PolicyLiteralReset, "state_policy", "RESET_EACH_CHUNK", "LITERAL", "true", "only state policy type-valid under frozen mnemonicspace.Runner.Prepare (no prior-state argument)"},
		{PolicyLiteralReset, "cue_namespace", "NOT_APPLICABLE", "LITERAL", "true", "literal mechanisms have no cue namespace axis"},
		{PolicyCueResetLocal, "state_policy", "RESET_EACH_CHUNK", "CUE", "true", "only state policy type-valid under frozen mnemonicspace.Runner.Prepare"},
		{PolicyCueResetLocal, "cue_namespace", "LOCAL_NAMESPACE", "CUE", "true", "cue labels C0..C(capacity-1) reused every chunk (matches Task82 convention)"},
		{PolicyCueResetGlobal, "state_policy", "RESET_EACH_CHUNK", "CUE", "true", "only state policy type-valid under frozen mnemonicspace.Runner.Prepare"},
		{PolicyCueResetGlobal, "cue_namespace", "GLOBAL_NAMESPACE", "CUE", "true", "cue labels are globally, deterministically indexed and never repeat across chunks"},
		{"ANY", "state_policy", "CONTINUE_STATE", "LITERAL,CUE", "false", "NOT_TYPE_VALID: mnemonicspace.Runner.Prepare takes no prior-state argument; continuation would require changing a frozen mechanism's Prepare contract (task82a.txt sec.6 forbids this) -- GENERIC_SCALING_POLICY finding, not a historical claim (sec.12)"},
		{"ANY", "convention_policy", "CONVENTION_GLOBAL", "LITERAL,CUE", "true", "single deterministic convention/association scheme reused for every sampled recovery test"},
		{"ANY", "convention_policy", "CONVENTION_PER_BLOCK", "LITERAL,CUE", "false", "preregistered but excluded from the main manifest for computational cost (TASK82A_COST_MODEL.tsv); GENERIC_SCALING_POLICY, not run"},
		{"ANY", "path_policy", "PATH_PER_CHUNK_RESTART", "LITERAL(geometry-carrier),CUE", "true", "geometry path re-derived fresh per chunk, consistent with RESET_EACH_CHUNK"},
		{"ANY", "path_policy", "PATH_REUSED_GLOBAL", "LITERAL(geometry-carrier)", "false", "would contradict RESET_EACH_CHUNK state policy; preregistered but excluded, not run"},
	}
	for _, r := range rows {
		b.WriteString(strings.Join(r[:], "\t") + "\n")
	}
	return os.WriteFile(filepath.Join(dir, "SCALING_POLICIES.tsv"), []byte(b.String()), 0o644)
}

func writeBoundaryProvenance(dir string) error {
	var b strings.Builder
	b.WriteString("boundary\tvalue\tprovenance\tnotes\n")
	rows := [][4]string{
		{"local_mechanism_boundary", "one local Runner.Prepare application per chunk", "GENERATED_BY_MECHANISM", "the frozen mechanism's own UnitBoundary is preserved unchanged from Task82"},
		{"token_boundary", "one token per chunk (chunk symbols joined, no separator)", "ASSEMBLER_DEFINED", "the frozen mechanism does not define a token concept; the assembler imposes exactly one token per local application"},
		{"line_boundary", "one line per chunk (== one token)", "ASSEMBLER_DEFINED", "generic layout variant fixed before generation, independent of chunk content or Voynich line statistics"},
		{"page_boundary", "not defined", "NOT_DEFINED", "no frozen mechanism defines pages/folios; no synthetic page structure is introduced (task82a.txt sec.21)"},
		{"assembly_boundary", "corpus_scale chunk count for the job", "ASSEMBLER_DEFINED", "distinct from local_mechanism_boundary and input_boundary; not treated as a historical page/line"},
		{"input_boundary", "matched literal-letter / word-token prefix length", "INHERITED_FROM_INPUT", "same source corpus files as Task82, truncated to a matched, scale-determined length"},
	}
	for _, r := range rows {
		b.WriteString(strings.Join(r[:], "\t") + "\n")
	}
	return os.WriteFile(filepath.Join(dir, "BOUNDARY_PROVENANCE.tsv"), []byte(b.String()), 0o644)
}
