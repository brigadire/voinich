// Command task79c-pf4-hr runs Task79c's PF4 leaf-paired null (Gate E) and
// HR3/HR5 out-of-sample predictive-hierarchy validation (Gate F) against a
// real Task79 line_profiles.json, per
// research/phase2/fingerprint/TASK79C_DESIGN.md sections 8-10. It computes
// no other Fingerprint v2 metric and does not touch the frozen Task79
// metric registry.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"zcore.dev/voinich/internal/fingerprintv2"
	"zcore.dev/voinich/internal/workdir"
)

type output struct {
	LineProfilesPath string                            `json:"line_profiles_path"`
	LineProfileCount int                               `json:"line_profile_count"`
	PF4              fingerprintv2.PF4LeafNullResult   `json:"pf4_leaf_null"`
	Hierarchy        fingerprintv2.HierarchyValidation `json:"hierarchy_out_of_sample"`
}

func main() {
	os.Exit(run())
}

func run() int {
	lineProfiles := flag.String("line-profiles", "", "path to a Task79 line_profiles.json (required)")
	out := flag.String("output", "", fmt.Sprintf("path for the combined result JSON (required; output_dir is explicit, not implicit %s)", workdir.Dir))
	pf4Permutations := flag.Int("pf4-permutations", 1000, "PF4 leaf-paired null permutation count (TASK79C_DESIGN.md section 8)")
	pf4Seed := flag.Int64("pf4-seed", 20260824, "PF4 leaf-paired null seed (TASK79C_DESIGN.md section 8)")
	hrFolds := flag.Int("hr-folds", 5, "HR3/HR5 folio-block CV fold count (TASK79C_DESIGN.md section 10)")
	hrSeed := flag.Int64("hr-seed", 40260824, "HR3/HR5 folio-fold assignment seed (TASK79C_DESIGN.md section 10)")
	hrMinGroupSize := flag.Int("hr-min-group-size", 5, "HR3/HR5 minimum training folios/sections per fold (TASK79C_DESIGN.md section 10)")
	flag.Parse()
	if *lineProfiles == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "usage: task79c-pf4-hr -line-profiles PATH -output PATH [flags]")
		return 2
	}

	profiles, err := fingerprintv2.LoadLineProfiles(*lineProfiles)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}

	pf4 := fingerprintv2.RunPF4LeafPairedNull(profiles, *pf4Permutations, *pf4Seed)
	hr := fingerprintv2.RunHierarchyOutOfSample(profiles, *hrFolds, *hrSeed, *hrMinGroupSize)

	result := output{
		LineProfilesPath: *lineProfiles,
		LineProfileCount: len(profiles),
		PF4:              pf4,
		Hierarchy:        hr,
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	data = append(data, '\n')
	if err := os.WriteFile(*out, data, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	fmt.Printf("Wrote %s\nPF4 verdict: %s (p=%g, paired leaves=%d)\nHierarchy verdict: %s (mean HR3 delta=%g, mean HR5 delta=%g)\n",
		*out, pf4.Verdict, pf4.PValue, pf4.PairedLeafCount, hr.Verdict, hr.MeanHR3Delta, hr.MeanHR5Delta)
	return 0
}
