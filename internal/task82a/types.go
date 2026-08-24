// Package task82a implements Task82a: a corpus-scale CorpusScaleAssembler
// layer over the frozen Task81 V1.1 mnemonic mechanisms (internal/task82,
// internal/mnemonicspace), followed by frozen Fingerprint V2 feature
// extraction (internal/fingerprintv2). It never reads Voynich data, a
// Voynich reference vector, or any notation-control (BDD) artifact.
package task82a

// Version/FreezeVersion identify this package's own frozen scientific
// design, independent of the Task81/Task82/F2 freezes it binds to.
const (
	Version       = "V1.0"
	FreezeVersion = "V1.0"
)

// MasterSeed is Task82a's own master seed, following the same convention
// as Task81/Task82 (a task-numbered constant folded into a SHA-256-derived
// per-job seed; see (Job).DerivedSeed).
const MasterSeed uint64 = 82024001

// CueCapacity is the frozen local-unit capacity shared by every
// cue-addressable mechanism that has no explicit Capacity/Period field of
// its own (F11, rotation-index). It is not derived from Voynich data: it
// is the same value already frozen for the sibling F12/storage-associate
// mechanisms in MNEMONIC_MECHANISM_REGISTRY.json, reused here as a
// preregistered generic scaling rule (task82a.txt sec.9-10).
const CueCapacity = 4

// PilotCheckpoints are the doubling chunk-count checkpoints probed by the
// target-blind scale convergence pilot (task82a.txt sec.24-25).
var PilotCheckpoints = []int{16, 32, 64, 128, 256, 512, 1024, 2048, 4096}

// SurfaceRole classifies a mechanism's local-unit capacity domain.
type SurfaceRole string

const (
	SurfaceLiteral SurfaceRole = "LITERAL"
	SurfaceCue     SurfaceRole = "CUE"
)

// MechanismScale is the frozen, target-independent local-unit capacity and
// surface role for one Task81 V1.1 mechanism, keyed by mechanism_id. Every
// value here traces to a field already present in
// MNEMONIC_MECHANISM_REGISTRY.json / the frozen ParameterSet, or is marked
// GENERIC_SCALING_POLICY when the mechanism has no explicit capacity field.
type MechanismScale struct {
	MechanismID    string
	ParameterSetID string
	Surface        SurfaceRole
	Capacity       int
	CapacitySource string
}

// MechanismScales is the frozen capacity table for all 16 Task81 V1.1
// mechanisms (task82a.txt sec.8-10).
var MechanismScales = []MechanismScale{
	{"f01_speculum_core", "f01-core-bounded", SurfaceLiteral, 12, "F01Parameters.NumRings"},
	{"f01_speculum_profile_latin23_r12", "f01-profile-latin23-r12", SurfaceLiteral, 12, "F01Parameters.NumRings"},
	{"f08_serpens_core", "f08-centre-edge-small", SurfaceLiteral, 10, "F08Parameters.Capacity"},
	{"synthetic_literal_storage", "f08-centre-edge-small", SurfaceLiteral, 10, "F08Parameters.Capacity (shared parameter set)"},
	{"negative_randomized_convention", "f01-core-bounded", SurfaceLiteral, 12, "F01Parameters.NumRings (shared parameter set)"},
	{"negative_randomized_path", "f08-centre-edge-small", SurfaceLiteral, 10, "F08Parameters.Capacity (shared parameter set)"},

	{"f11_arismetricum_core", "f11-core-default", SurfaceCue, CueCapacity, "GENERIC_SCALING_POLICY: no explicit capacity field; matches sibling F12/storage-associate capacity"},
	{"f12_horalogius_core", "f12-cycle-default", SurfaceCue, CueCapacity, "F12Parameters.Period"},
	{"m_restricted_rotation_index", "rotation-index-small", SurfaceCue, CueCapacity, "GENERIC_SCALING_POLICY: no explicit capacity field; matches sibling F12/storage-associate capacity"},
	{"m_restricted_storage_associate", "storage-associate-small", SurfaceCue, CueCapacity, "StorageAssociateParameters.Capacity"},
	{"synthetic_cyclic_state", "f12-cycle-default", SurfaceCue, CueCapacity, "F12Parameters.Period (shared parameter set)"},
	{"synthetic_indexed_lookup", "f11-core-default", SurfaceCue, CueCapacity, "GENERIC_SCALING_POLICY (shared parameter set)"},
	{"synthetic_cue_based", "f12-cycle-default", SurfaceCue, CueCapacity, "F12Parameters.Period (shared parameter set)"},
	{"synthetic_ambiguous", "storage-associate-small", SurfaceCue, CueCapacity, "StorageAssociateParameters.Capacity (shared parameter set)"},
	{"negative_randomized_cue_association", "f12-cycle-default", SurfaceCue, CueCapacity, "F12Parameters.Period (shared parameter set)"},
	{"negative_randomized_index_mapping", "f11-core-default", SurfaceCue, CueCapacity, "GENERIC_SCALING_POLICY (shared parameter set)"},
}

func scaleFor(mechanismID string) (MechanismScale, bool) {
	for _, s := range MechanismScales {
		if s.MechanismID == mechanismID {
			return s, true
		}
	}
	return MechanismScale{}, false
}

// Scaling policy IDs (task82a.txt sec.11, 15, 30, 56).
const (
	PolicyLiteralReset   = "LITERAL_RESET_V1"
	PolicyCueResetLocal  = "CUE_RESET_LOCAL_V1"
	PolicyCueResetGlobal = "CUE_RESET_GLOBAL_V1"
)

// PoliciesFor lists the frozen, included scaling policies for a mechanism's
// surface role. Every mechanism gets RESET_EACH_CHUNK state policy (the
// only state policy that is type-valid under the frozen
// mnemonicspace.Runner.Prepare contract, which takes no prior-state
// argument); cue mechanisms additionally vary cue namespace.
func PoliciesFor(surface SurfaceRole) []string {
	if surface == SurfaceLiteral {
		return []string{PolicyLiteralReset}
	}
	return []string{PolicyCueResetLocal, PolicyCueResetGlobal}
}

// Corpus scale IDs and chunk counts (task82a.txt sec.23-25). Values are
// filled in by design freeze from the RunScaleConvergencePilot result, not
// chosen from Voynich statistics.
type ScaleSpec struct {
	ID     string
	Chunks int
}

// CorpusScales are frozen from RunScaleConvergencePilot's real output (see
// TASK82A_DESIGN.md "Scale convergence pilot"): the pilot's own
// entropy-convergence criterion (<=0.01 bit delta on doubling) is met at
// 256 chunks, so LARGE is fixed at the convergence point itself and
// SMALL/MEDIUM are the two preceding pre-convergence doublings -- this
// shows the approach to stability without scaling past it. A second,
// independently target-blind ceiling applies on top of the entropy
// criterion: GLOBAL-cue-namespace and literal vocabularies grow with chunk
// count, and a real timing pilot (f2_timing_test.go) measured frozen F2
// extraction cost at that vocabulary size directly (0.44s/3.3s/29s at
// 64/256/1024 GLOBAL-namespace tokens); scaling past the entropy
// convergence point would multiply job cost without adding scientific
// signal the pilot didn't already show converging, so 256 chunks is also
// used as the computational-cost ceiling (task82a.txt sec.23: estimator
// convergence and computational cost, not Voynich statistics).
var CorpusScales = []ScaleSpec{
	{"SMALL", 64},
	{"MEDIUM", 128},
	{"LARGE", 256},
}

// Replicates is the number of deterministic replicate seeds per job cell,
// matching Task82's own frozen replicate count (task82a.txt sec.29).
const Replicates = 2

type Options struct {
	Root       string
	ShardIndex int
	ShardCount int
	Resume     bool
	VerifyOnly bool
}
