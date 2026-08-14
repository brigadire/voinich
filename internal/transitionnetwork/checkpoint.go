package transitionnetwork

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type checkpoint struct {
	Version          int            `json:"version"`
	Fingerprint      string         `json:"fingerprint"`
	Phase            string         `json:"phase"`
	Completed        int            `json:"completed"`
	EdgeExceed       map[string]int `json:"edge_exceed"`
	OutExceed        map[string]int `json:"out_exceed"`
	InExceed         map[string]int `json:"in_exceed"`
	OutSignExceed    map[string]int `json:"out_sign_exceed"`
	InSignExceed     map[string]int `json:"in_sign_exceed"`
	OutEntropyExceed map[string]int `json:"out_entropy_exceed"`
	InEntropyExceed  map[string]int `json:"in_entropy_exceed"`
	RefineCandidates []string       `json:"refine_candidates,omitempty"`
	RefineCompleted  int            `json:"refine_completed,omitempty"`
	RefineExceed     map[string]int `json:"refine_exceed,omitempty"`
}

func fingerprint(c Config, corpusSHA, metaSHA string) string {
	s := fmt.Sprintf("v2|%s|%s|%d|%d|%d|%d|%d", corpusSHA, metaSHA, c.MinTokenCount, c.MinBlockTokenCount, c.Permutations, c.RefinePermutations, c.Seed)
	x := sha256.Sum256([]byte(s))
	return hex.EncodeToString(x[:])
}
func freshCheckpoint(fp string) checkpoint {
	return checkpoint{Version: 2, Fingerprint: fp, Phase: "primary", EdgeExceed: map[string]int{}, OutExceed: map[string]int{}, InExceed: map[string]int{}, OutSignExceed: map[string]int{}, InSignExceed: map[string]int{}, OutEntropyExceed: map[string]int{}, InEntropyExceed: map[string]int{}, RefineExceed: map[string]int{}}
}
func loadCheckpoint(path, fp string) (checkpoint, bool, error) {
	if path == "-" {
		return freshCheckpoint(fp), false, nil
	}
	raw, e := os.ReadFile(path)
	if os.IsNotExist(e) {
		return freshCheckpoint(fp), false, nil
	}
	if e != nil {
		return checkpoint{}, false, e
	}
	var cp checkpoint
	if json.Unmarshal(raw, &cp) != nil || cp.Version != 2 || cp.Fingerprint != fp {
		return freshCheckpoint(fp), false, nil
	}
	if cp.EdgeExceed == nil {
		cp.EdgeExceed = map[string]int{}
	}
	if cp.OutExceed == nil {
		cp.OutExceed = map[string]int{}
	}
	if cp.InExceed == nil {
		cp.InExceed = map[string]int{}
	}
	if cp.OutSignExceed == nil {
		cp.OutSignExceed = map[string]int{}
	}
	if cp.InSignExceed == nil {
		cp.InSignExceed = map[string]int{}
	}
	if cp.OutEntropyExceed == nil {
		cp.OutEntropyExceed = map[string]int{}
	}
	if cp.InEntropyExceed == nil {
		cp.InEntropyExceed = map[string]int{}
	}
	if cp.RefineExceed == nil {
		cp.RefineExceed = map[string]int{}
	}
	return cp, true, nil
}
func saveCheckpoint(path string, cp checkpoint) error {
	if path == "-" {
		return nil
	}
	if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
		return e
	}
	raw, e := json.Marshal(cp)
	if e != nil {
		return e
	}
	tmp := path + ".tmp"
	if e = os.WriteFile(tmp, raw, 0644); e != nil {
		return e
	}
	return os.Rename(tmp, path)
}
