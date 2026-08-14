package positionalcontinuation

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// RunState accumulates every part's output for this single frozen-target
// analysis. Unlike higherorderseq's per-candidate map, there is only one
// target ("s aiin" -> "chey"), so the whole pipeline state is one struct
// checkpointed as a unit after each part completes.
type RunState struct {
	SAiinOccurrences []SAiinOccurrence
	AiinOccurrences  []AiinOccurrence

	ContinuationRows []ContinuationRow
	DistSummaryRows  []DistributionSummaryRow

	PositionDependence []PositionDependenceRow
	PositionalEntropy  []PositionalEntropyRow
	CheyEffect         []CheyEffectRow
	AiinControl        []AiinControlRow
	StratifiedPredecessor []StratifiedPredecessorRow

	ModelLOBO ModelLOBORow

	CrossBlock []CrossBlockPositionalRow

	Jackknife []PositionalJackknifeRow

	LineVsBlock            []LineVsBlockRow
	LineVsBlockCorrelation float64
	LineVsBlockSource      string

	BoundaryDistance []BoundaryDistanceRow

	SurroundingContext []SurroundingContextRow

	ReversePosition []ReversePositionRow

	Validation ValidationRow
}

// Checkpoint captures per-part progress so an interrupted run (crash, kill,
// power loss) can resume without redoing already-completed work. It is only
// ever applied when its Fingerprint exactly matches the current corpus,
// metadata, frozen higher-order-sequences inputs and CLI parameters.
type Checkpoint struct {
	Fingerprint string          `json:"fingerprint"`
	PartsDone   map[string]bool `json:"parts_done"`
	State       *RunState       `json:"state"`
}

func newCheckpoint(fingerprint string) *Checkpoint {
	return &Checkpoint{Fingerprint: fingerprint, PartsDone: map[string]bool{}, State: &RunState{}}
}

// computeFingerprint binds a checkpoint to the exact inputs and parameters
// that produced it. Any change - a different corpus, metadata map, frozen
// higher-order-sequences directory contents, or any CLI parameter - must
// invalidate a prior checkpoint rather than risk silently resuming into a
// mismatched run.
func computeFingerprint(c Config, corpusSHA, metaSHA, higherOrderSHA string) string {
	h := sha256.New()
	fmt.Fprintf(h, "v1\ncorpus=%s\nmeta=%s\nhigherorder=%s\npermutations=%d\nseed=%d\nalpha=%v\n",
		corpusSHA, metaSHA, higherOrderSHA, c.Permutations, c.Seed, smoothingAlpha)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func saveCheckpoint(path string, cp *Checkpoint) error {
	if path == "" {
		return nil
	}
	b, err := json.Marshal(cp)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func loadCheckpoint(path, fingerprint string) (*Checkpoint, bool) {
	if path == "" {
		return nil, false
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var cp Checkpoint
	if err := json.Unmarshal(b, &cp); err != nil {
		return nil, false
	}
	if cp.Fingerprint != fingerprint {
		return nil, false
	}
	if cp.PartsDone == nil {
		cp.PartsDone = map[string]bool{}
	}
	if cp.State == nil {
		cp.State = &RunState{}
	}
	return &cp, true
}

func removeCheckpoint(path string) {
	if path == "" {
		return
	}
	_ = os.Remove(path)
}
