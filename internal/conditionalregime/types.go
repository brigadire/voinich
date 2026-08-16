// Package conditionalregime asks a stricter question than global-regime-analyze
// and metadata-validate: after conditioning on Currier and Davis hand, is there
// still reproducible distributional structure in the corpus? Discovery here
// never sees an "improved" corpus - it reuses the same token sequence, the
// same distributional representation (internal/globalregime) and the same
// change-point/clustering primitives, applied only within (Part A) or after
// removing (Part B/C) the Currier x hand signature.
//
// Every parameter in this package (window sizes, K ranges, eligibility
// thresholds, permutation counts) is fixed by task 19 before any result is
// examined and is not tuned afterward.
package conditionalregime

import (
	"context"
	"io"
	"time"
)

// Config describes one conditional-regime-analyze run.
type Config struct {
	CorpusPath          string
	TokenMetadataMap    string
	OutputDir           string
	WindowSizes         []int // within-class (Part A) frozen scales
	ResidualWindowSizes []int // pooled residual (Part B) frozen scales
	MinClassTokens      int
	MinBlockTokens      int
	KMin                int
	KMaxWithin          int
	KMaxResidual        int
	Permutations        int
	Seed                int64
	// Workers bounds the number of concurrently executing permutation jobs.
	// It is operational, not scientific, and is therefore intentionally not
	// part of the checkpoint fingerprint. Defaults to 1.
	Workers int
	// Executor selects the backend that runs permutation jobs: "goroutine"
	// (default) uses the bounded in-process pool from Task31; "process"
	// dispatches every job to one of Workers persistent subprocess workers
	// (Task32), which call the identical scientific implementation in a
	// separate OS process. Like Workers, this is operational, not
	// scientific: it is intentionally excluded from the checkpoint
	// fingerprint so a resumed run may switch backends.
	Executor string
	// RemoteListen is the coordinator's own mTLS listen address when
	// Executor is "remote" (e.g. "0.0.0.0:8443"). Workers dial in and lease
	// jobs from it; the coordinator never dials out (Task34: the
	// coordinator is the TLS server with the fixed, addressable identity,
	// every worker is a TLS client with its own individual identity).
	RemoteListen string
	// TLSCert/TLSKey are the coordinator's own certificate/key (EKU
	// serverAuth, signed by the project CA; internal/pki issues them).
	// ClientCA is the project CA bundle used to verify every connecting
	// worker's client certificate. All three are required when Executor is
	// "remote".
	TLSCert  string
	TLSKey   string
	ClientCA string
	// RemoteDenyList optionally names a JSON deny-list file
	// (internal/pki.DenyList) revoking specific certificate serials or
	// authenticated worker identities. Empty means nothing is revoked.
	RemoteDenyList string
	RemoteTimeout  time.Duration
	// RemoteRetries bounds how many times the coordinator reassigns a job
	// to another worker after a lease expires unanswered before failing it.
	RemoteRetries int
	Context       context.Context
	// CheckpointPath, if non-empty, is where progress is saved after every
	// completed unit of work (a class x window_size combo, or - for the
	// slowest loop, Part B's global permutation correction - every single
	// replicate). If a matching checkpoint (same corpus, metadata and every
	// parameter) exists at this path when the run starts, already-completed
	// work is loaded from it instead of being recomputed. Defaults to
	// <output-dir>/checkpoint.json; set to "-" to disable checkpointing.
	CheckpointPath string
	Quiet          bool
	ProgressWriter io.Writer
}

// refinementPermutations is the fixed 10000-permutation refinement budget
// for the strongest candidates (task19 section 22). It is not a CLI
// parameter: the candidate-selection rule and this count are frozen before
// any result is examined (see candidateForRefinement in withinclass.go).
const refinementPermutations = 10000

// Scheme is the conditioning variable used to build physical blocks: the
// joint Currier x hand class (primary) or one metadata dimension alone
// (secondary).
type Scheme string

const (
	SchemeJoint       Scheme = "joint"
	SchemeCurrierOnly Scheme = "currier_only"
	SchemeHandOnly    Scheme = "hand_only"
)

// ClassID identifies one conditioning class. Currier/Hand is empty when not
// part of the scheme (e.g. Hand is always "" for SchemeCurrierOnly).
type ClassID struct {
	Scheme  Scheme
	Currier string
	Hand    string
}

// Label is the display form used in output files, e.g. "2/2" for a joint
// class or "2" for a single-dimension class.
func (c ClassID) Label() string {
	switch c.Scheme {
	case SchemeCurrierOnly:
		return c.Currier
	case SchemeHandOnly:
		return c.Hand
	default:
		return c.Currier + "/" + c.Hand
	}
}

// Block is one maximal contiguous run of tokens belonging to the same class.
// Blocks never bridge an unknown-metadata gap or a change in class: task 19
// requires treating every contiguous run as its own physical block first,
// never silently joining non-adjacent runs into one artificial stream.
type Block struct {
	Class      ClassID
	Index      int // order of appearance among this class's blocks
	Start, End int // token positions [Start, End)
}

func (b Block) Len() int { return b.End - b.Start }

// ClassInfo is one row of the eligibility inventory (task19 section 9).
type ClassInfo struct {
	Class        ClassID
	TotalTokens  int
	BlockCount   int
	LargestBlock int
	MedianBlock  float64
	Eligible     bool
}

// EmpiricalStats is the full set of descriptive fields required for every
// permutation-based statistic (task19 section 43): never just a p-value.
type EmpiricalStats struct {
	Observed     float64
	NullMean     float64
	NullSD       float64
	NullP95      float64
	NullP99      float64
	NullMax      float64
	EffectSize   float64 // (observed - null mean) / null SD
	Exceedances  int
	EmpiricalP   float64
	Permutations int
}
