package fingerprintv2

import (
	"encoding/json"
	"os"
)

// LoadLineProfiles reads a line_profiles.json file (as emitted by the
// Task79 canonical pipeline, e.g.
// experiments/fingerprint-v2-task79-v1/canonical-out/line_profiles.json)
// for reuse by Task79c's PF4 leaf-paired null and HR3/HR5 out-of-sample
// validation. It reuses the frozen LineProfile record unchanged.
func LoadLineProfiles(path string) ([]LineProfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var profiles []LineProfile
	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, err
	}
	return profiles, nil
}

// RunPF4LeafPairedNull exposes pf4LeafPairedNull (TASK79C_DESIGN.md §8) to
// callers outside this package.
func RunPF4LeafPairedNull(profiles []LineProfile, permutations int, seed int64) PF4LeafNullResult {
	return pf4LeafPairedNull(profiles, permutations, seed)
}

// RunHierarchyOutOfSample exposes hierarchyOutOfSample (TASK79C_DESIGN.md
// §9-10) to callers outside this package.
func RunHierarchyOutOfSample(profiles []LineProfile, folds int, seed int64, minGroupSize int) HierarchyValidation {
	return hierarchyOutOfSample(profiles, folds, seed, minGroupSize)
}
