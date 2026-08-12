package main

import (
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
	"zcore.dev/voinich/internal/profilestability"
)

func main() {
	input := flag.String("input", "data_work/ivtt_output_1786282555007.txt", "raw corpus")
	classes := flag.String("classes", "structural_classes.yaml", "full-corpus structural classes")
	output := flag.String("output", "structural_profile_stability.yaml", "output YAML")
	folds := flag.Int("folds", 5, "number of deterministic line folds")
	foldSeed := flag.Int64("fold-seed", 1, "fold assignment seed")
	minCount := flag.Int("min-token-count", 10, "minimum count independently in each sample")
	neighbors := flag.Int("neighbors", 10, "nearest structural neighbors per token")
	bootstrapRuns := flag.Int("bootstrap-runs", 200, "line bootstrap runs")
	bootstrapSeed := flag.Int64("bootstrap-seed", 1, "line bootstrap seed")
	threshold := flag.Float64("threshold", .70, "fixed reference threshold")
	thresholdMargin := flag.Float64("threshold-margin", .05, "absolute near-threshold margin")
	flag.Parse()
	result, err := profilestability.Run(profilestability.Config{InputPath: *input, ClassesPath: *classes, Folds: *folds, FoldSeed: *foldSeed, MinTokenCount: *minCount, Neighbors: *neighbors, BootstrapRuns: *bootstrapRuns, BootstrapSeed: *bootstrapSeed, Threshold: *threshold, ThresholdMargin: *thresholdMargin, Progress: func(message string) { fmt.Println(message) }})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	data, err := yaml.Marshal(result)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: encode output:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(*output, data, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "Error: write output:", err)
		os.Exit(1)
	}
	fmt.Printf("Profile stability written to %s\n", *output)
}
