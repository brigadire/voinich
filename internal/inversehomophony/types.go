// Package inversehomophony implements task57's blind, ciphertext-only
// homophone-class recovery method and its synthetic validation harness.
//
// The recovery engine (features.go, similarity.go, cluster.go) never reads
// plaintext, a Task46/55 mapping, or any filename/metadata that reveals
// plaintext identity - see corpus.go's Relabel and INVERSE_HOMOPHONY_DESIGN.md
// section 5/2. Oracle-dependent code (diagnostic.go's true/false pair
// labels, baselines.go's OraclePartition, classrecovery.go) is evaluator-only
// and is never given to Recover.
package inversehomophony

// MethodVersion identifies the frozen recovery method (features,
// normalization, similarity formula, clustering rule, thresholds). Bump
// only for a deliberate, documented method change - never to chase a
// Voynich result (task57 section 22).
const MethodVersion = "inverse-homophony-v1"

// Config is the complete, frozen hyperparameter set for the recovery
// engine. Every field here is fixed on DEVELOPMENT corpora only (see
// INVERSE_HOMOPHONY_DESIGN.md sections 4, 6, 16) before any VALIDATION or
// Voynich corpus is scored.
type Config struct {
	// Distances are the extra (beyond the immediate predecessor/successor)
	// context offsets folded into the distance-context profile.
	Distances []int
	// PositionalBuckets is the fixed number of index-in-line/line-length
	// histogram buckets.
	PositionalBuckets int
	// MinSupport is the minimum combined predecessor+successor observation
	// count a pair must have on both members before it is even scored.
	MinSupport int
	// Threshold (tau) is the minimum combined similarity score required to
	// merge two classes. Fit on development corpora via the pair
	// discrimination diagnostic (Youden's J), then frozen.
	Threshold float64
	// MaxClassFraction caps a merged class's occurrence-weighted fraction
	// of the whole corpus (anti-trivial-collapse, task57 section 15).
	MaxClassFraction float64
	// MinEntropyFraction floors the occurrence-weighted Shannon entropy of
	// the class-size distribution, as a fraction of the NO_COLLAPSE
	// partition's entropy (anti-trivial-collapse, task57 section 15).
	MinEntropyFraction float64
	// Seed drives every deterministic pseudo-random draw (RANDOM_PARTITION
	// baseline sampling, false-pair sampling in the diagnostic).
	Seed int64
}

// FrozenConfig is the one method configuration used for every VALIDATION
// and Voynich run. Its Threshold is filled in by FreezeThreshold from the
// development pair-discrimination diagnostic - see validate.go - and must
// not change afterward.
func FrozenConfig() Config {
	return Config{
		Distances:          []int{2, 3, 4, 5},
		PositionalBuckets:  5,
		MinSupport:         5,
		MaxClassFraction:   0.15,
		MinEntropyFraction: 0.5,
		Seed:               1,
		// Threshold is set by FreezeThreshold once development
		// discrimination has been measured; 0 here is not a valid frozen
		// value and RunSyntheticValidation fails if it is left unset.
	}
}

// Partition maps a relabeled (opaque) cipher token to a class label. Labels
// are opaque strings (never numeric IDs meant to be compared across
// partitions - task57 section 10 forbids comparing class IDs directly;
// always compare partitions via ClassRecoveryMetrics's contingency table).
type Partition map[string]string

// TokenFeatures holds every ciphertext-only feature for one relabeled
// token (task57 section 7). All distributions are raw counts; callers
// normalize before comparing (similarity.go).
type TokenFeatures struct {
	Freq    int
	Pred    map[string]int // distance -1
	Succ    map[string]int // distance +1
	DistCtx map[string]int // union of distances in Config.Distances, both sides
	PosHist []int          // Config.PositionalBuckets buckets
}

// PairScore is one scored candidate edge in the merge-evidence graph
// (task57 section 18), kept for the audit trail (section 17).
type PairScore struct {
	A, B      string
	Support   int
	PredScore float64
	SuccScore float64
	DistScore float64
	PosScore  float64
	Score     float64
}

// MergeEvent is one accepted or rejected merge decision, the audit trail
// required by task57 section 17/18.
type MergeEvent struct {
	A, B               string
	Score              float64
	Accepted           bool
	Reason             string
	ClassSizeAfter     int
	ClassFractionAfter float64
}
