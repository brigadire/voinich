package conditionalregime

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Checkpoint captures enough progress to resume conditional-regime-analyze
// after an interruption (process kill, crash, power loss) without redoing
// already-completed work. It is only ever applied when its Fingerprint
// exactly matches the current corpus, metadata and parameters; any mismatch
// is treated as "no usable checkpoint" so a run never silently resumes into
// a different experiment.
type Checkpoint struct {
	Fingerprint string `json:"fingerprint"`

	WithinCombosDone map[string]bool        `json:"within_combos_done"`
	WithinRegimes    []WithinClassRegime    `json:"within_regimes"`
	WithinStability  []WithinClassStability `json:"within_stability"`
	CrossBlock       []CrossBlockRecurrence `json:"cross_block"`

	SignificanceCombosDone map[string]bool        `json:"significance_combos_done"`
	Candidates             []WithinClassCandidate `json:"candidates"`
	RefinementDone         bool                   `json:"refinement_done"`

	ResidualSweepCombosDone map[string]bool          `json:"residual_sweep_combos_done"`
	ResidualSummary         []ResidualClusterSummary `json:"residual_summary"`
	BestRaw                 residualSweepResult      `json:"best_raw"`
	BestRawHier             residualSweepResult      `json:"best_raw_hier"`

	// ResidualCorrectionNull holds the partial null distribution accumulated
	// so far for each method's global max-over-scale-x-K permutation
	// correction (task19 section 41) - the single most expensive loop in the
	// whole pipeline. It is saved after every replicate specifically so an
	// interruption here loses at most one replicate's work.
	ResidualCorrectionNull map[string][]float64      `json:"residual_correction_null"`
	ResidualCorrectionDone map[string]bool           `json:"residual_correction_done"`
	ResidualCorrection     map[string]EmpiricalStats `json:"residual_correction"`

	PartCDone      bool                        `json:"part_c_done"`
	Boundaries     []ConditionalStableBoundary `json:"boundaries"`
	RecurringTypes []RecurringBoundaryType     `json:"recurring_types"`
	Transitions    []ResidualTransitionCell    `json:"transitions"`
}

func newCheckpoint(fingerprint string) *Checkpoint {
	return &Checkpoint{
		Fingerprint:             fingerprint,
		WithinCombosDone:        map[string]bool{},
		SignificanceCombosDone:  map[string]bool{},
		ResidualSweepCombosDone: map[string]bool{},
		ResidualCorrectionNull:  map[string][]float64{},
		ResidualCorrectionDone:  map[string]bool{},
		ResidualCorrection:      map[string]EmpiricalStats{},
	}
}

// computeFingerprint binds a checkpoint to the exact inputs and parameters
// that produced it. Any change at all - a different corpus, metadata map, or
// any CLI parameter - must invalidate a prior checkpoint rather than risk
// silently resuming into a mismatched run.
func computeFingerprint(c Config, corpusHash, metaHash string) string {
	h := sha256.New()
	fmt.Fprintf(h, "corpus=%s\nmeta=%s\nwindow_sizes=%v\nresidual_window_sizes=%v\nmin_class_tokens=%d\nmin_block_tokens=%d\nk_min=%d\nk_max_within=%d\nk_max_residual=%d\npermutations=%d\nseed=%d\n",
		corpusHash, metaHash, c.WindowSizes, c.ResidualWindowSizes, c.MinClassTokens, c.MinBlockTokens, c.KMin, c.KMaxWithin, c.KMaxResidual, c.Permutations, c.Seed)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// saveCheckpoint writes atomically (write-then-rename) so a process killed
// mid-write can never leave a corrupt checkpoint file behind.
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

// loadCheckpoint returns (checkpoint, true, nil) only when a readable,
// well-formed checkpoint exists at path AND its fingerprint matches. Any
// other case - missing file, corrupt JSON, or a fingerprint mismatch -
// returns (nil, false, nil): the caller starts fresh rather than treat a
// read/parse error as fatal.
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
	if cp.WithinCombosDone == nil {
		cp.WithinCombosDone = map[string]bool{}
	}
	if cp.SignificanceCombosDone == nil {
		cp.SignificanceCombosDone = map[string]bool{}
	}
	if cp.ResidualSweepCombosDone == nil {
		cp.ResidualSweepCombosDone = map[string]bool{}
	}
	if cp.ResidualCorrectionNull == nil {
		cp.ResidualCorrectionNull = map[string][]float64{}
	}
	if cp.ResidualCorrectionDone == nil {
		cp.ResidualCorrectionDone = map[string]bool{}
	}
	if cp.ResidualCorrection == nil {
		cp.ResidualCorrection = map[string]EmpiricalStats{}
	}
	return &cp, true
}

func removeCheckpoint(path string) {
	if path == "" {
		return
	}
	_ = os.Remove(path)
}

func withinComboKey(class ClassID, windowSize int) string {
	return string(class.Scheme) + "|" + class.Label() + "|" + fmt.Sprint(windowSize)
}

func residualSweepComboKey(method string, standardized bool) string {
	return method + "|" + representationName(standardized)
}
