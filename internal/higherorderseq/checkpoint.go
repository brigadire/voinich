package higherorderseq

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Checkpoint captures per-candidate, per-part progress so an interrupted run
// (crash, kill, power loss) can resume without redoing already-completed
// work. Every candidate's analysis is broken into six checkpointed parts
// (occurrences+conditional A-C, CMI permutation D, LOBO E, context/
// continuation/cross-block/meta F-I, jackknife J, position/structural-family
// K-L); Part D's permutation loop is by far the most expensive step for a
// frequent central token, so it is its own checkpoint unit. It is only ever
// applied when its Fingerprint exactly matches the current corpus, metadata,
// frozen audit inputs and CLI parameters.
type Checkpoint struct {
	Fingerprint string                       `json:"fingerprint"`
	PartsDone   map[string]bool              `json:"parts_done"`
	Results     map[string]*CandidateResult  `json:"results"`
}

func newCheckpoint(fingerprint string) *Checkpoint {
	return &Checkpoint{Fingerprint: fingerprint, PartsDone: map[string]bool{}, Results: map[string]*CandidateResult{}}
}

func partKey(sequence, part string) string { return sequence + "|" + part }

func (cp *Checkpoint) resultFor(sequence string) *CandidateResult {
	r, ok := cp.Results[sequence]
	if !ok {
		r = &CandidateResult{}
		cp.Results[sequence] = r
	}
	return r
}

// computeFingerprint binds a checkpoint to the exact inputs and parameters
// that produced it. Any change - a different corpus, metadata map, frozen
// audit directory contents, or any CLI parameter - must invalidate a prior
// checkpoint rather than risk silently resuming into a mismatched run.
func computeFingerprint(c Config, corpusSHA, metaSHA, auditSHA string) string {
	h := sha256.New()
	fmt.Fprintf(h, "v1\ncorpus=%s\nmeta=%s\naudit=%s\npermutations=%d\nseed=%d\n", corpusSHA, metaSHA, auditSHA, c.Permutations, c.Seed)
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
	if cp.Results == nil {
		cp.Results = map[string]*CandidateResult{}
	}
	return &cp, true
}

func removeCheckpoint(path string) {
	if path == "" {
		return
	}
	_ = os.Remove(path)
}
