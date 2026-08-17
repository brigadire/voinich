package conditionalregime

import "encoding/json"

// workerProtocolVersion identifies the wire format spoken between the
// coordinator and a local subprocess worker (Task32). It has no relation to
// PID, wall clock, hostname or completion order: a worker that speaks a
// different version rejects the handshake explicitly instead of guessing.
const workerProtocolVersion = 2

// protocolMessage is the single line-delimited-JSON message shape used for
// every message in both directions. Kind selects which fields are
// meaningful; every other field is the zero value and omitted from the wire
// via `omitempty`. This is deliberately the smallest standard-library-only
// representation (encoding/json over a newline-delimited stream) rather than
// a binary or generated-code protocol - Task32 explicitly prefers this
// unless measurement disproves it, and job payloads here are a handful of
// bytes each way.
type protocolMessage struct {
	Kind     string `json:"kind"` // "init", "ready", "job", "result", "shutdown"
	Workload string `json:"workload,omitempty"`

	// Init fields: the coordinator's explicit input/config identity. A
	// worker must reject the handshake unless its own corpus/metadata load
	// and its own copy of every scientific parameter produce the identical
	// Fingerprint (computeFingerprint, the same function that guards
	// checkpoint resume) - this is what makes "malformed/incompatible
	// messages fail explicitly" concrete rather than aspirational.
	Version                 int     `json:"version,omitempty"`
	Fingerprint             string  `json:"fingerprint,omitempty"`
	CorpusPath              string  `json:"corpus_path,omitempty"`
	TokenMetadataMap        string  `json:"token_metadata_map,omitempty"`
	WindowSizes             []int   `json:"window_sizes,omitempty"`
	ResidualWindowSizes     []int   `json:"residual_window_sizes,omitempty"`
	MinClassTokens          int     `json:"min_class_tokens,omitempty"`
	MinBlockTokens          int     `json:"min_block_tokens,omitempty"`
	KMin                    int     `json:"k_min,omitempty"`
	KMaxWithin              int     `json:"k_max_within,omitempty"`
	KMaxResidual            int     `json:"k_max_residual,omitempty"`
	Permutations            int     `json:"permutations,omitempty"`
	Seed                    int64   `json:"seed,omitempty"`
	StructuralPairsPath     string  `json:"structural_pairs_path,omitempty"`
	DistancePairsPath       string  `json:"distance_pairs_path,omitempty"`
	FamiliesPath            string  `json:"families_path,omitempty"`
	MinStructuralSimilarity float64 `json:"min_structural_similarity,omitempty"`
	MinReliability          float64 `json:"min_reliability,omitempty"`
	ProjectionK             int     `json:"projection_k,omitempty"`
	RandomProjections       int     `json:"random_projections,omitempty"`
	MaxDistance             int     `json:"max_distance,omitempty"`
	MinObservations         int     `json:"min_observations,omitempty"`
	TopN                    int     `json:"top_n,omitempty"`
	FamilyID                int     `json:"family_id,omitempty"`
	ProjectionMode          string  `json:"projection_mode,omitempty"`
	Pair                    string  `json:"pair,omitempty"`

	// normalization_compare workload fields (Task42). ClassesPath names the
	// staged structural_classes.yaml input; MinTokenCount/SingletonMode come
	// from that file's own Meta (never a worker-local default) and Seed
	// above is reused as the random-baseline base seed.
	ClassesPath   string `json:"classes_path,omitempty"`
	MinTokenCount int    `json:"min_token_count,omitempty"`
	SingletonMode string `json:"singleton_mode,omitempty"`
	RandomRuns    int    `json:"random_runs,omitempty"`

	// Ready fields.
	OK    bool   `json:"ok,omitempty"`
	Error string `json:"error,omitempty"`

	// Job/Result fields.
	JobID *JobID          `json:"job_id,omitempty"`
	Value float64         `json:"value,omitempty"`
	Blob  json.RawMessage `json:"blob,omitempty"`
}

// scientificConfig extracts exactly the Config fields computeFingerprint
// hashes, so a worker can reconstruct the identical fingerprint from an
// Init message without needing any operational field (OutputDir, Workers,
// Executor, CheckpointPath, ...).
func (m protocolMessage) scientificConfig() Config {
	return Config{
		WindowSizes:         m.WindowSizes,
		ResidualWindowSizes: m.ResidualWindowSizes,
		MinClassTokens:      m.MinClassTokens,
		MinBlockTokens:      m.MinBlockTokens,
		KMin:                m.KMin,
		KMaxWithin:          m.KMaxWithin,
		KMaxResidual:        m.KMaxResidual,
		Permutations:        m.Permutations,
		Seed:                m.Seed,
	}
}
