package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"zcore.dev/voinich/internal/profiling"
	"zcore.dev/voinich/internal/structuralreliability"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
	os.Exit(run())
}

func run() (code int) {
	start := time.Now()
	input := flag.String("input", "data_work/ZL3b-x7.txt", "IVTT -x7 derived corpus")
	classes := flag.String("classes", workdir.Path("structural_classes.yaml"), "full-corpus structural classes")
	output := flag.String("output", workdir.Path("structural_reliability.yaml"), "output YAML")
	folds := flag.Int("folds", 5, "number of deterministic line folds")
	foldSeed := flag.Int64("fold-seed", 1, "fold assignment seed")
	minCount := flag.Int("min-token-count", 10, "base minimum count independently in each sample")
	neighbors := flag.Int("neighbors", 10, "nearest structural neighbors per token")
	bootstrapRuns := flag.Int("bootstrap-runs", 200, "line bootstrap runs")
	bootstrapSeed := flag.Int64("bootstrap-seed", 1, "line bootstrap seed")
	threshold := flag.Float64("threshold", .70, "fixed reference threshold")
	thresholdMargin := flag.Float64("threshold-margin", .05, "absolute near-threshold margin")
	countThresholds := flag.String("count-thresholds", "10,20,40,80,160,320", "comma-separated cumulative frequency thresholds")
	subsampleMinFullCount := flag.Int("subsample-min-full-count", 160, "minimum full-corpus count for the subsampling experiment")
	subsampleRuns := flag.Int("subsample-runs", 100, "repetitions per token per sample size")
	subsampleSeed := flag.Int64("subsample-seed", 1, "subsampling seed")
	prof := profiling.RegisterFlags(flag.CommandLine)
	flag.Parse()

	defer profiling.PrintElapsed(os.Stderr, start)

	sess, err := profiling.Start(prof)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	defer func() {
		if err := sess.Stop(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			code = 1
		}
	}()

	thresholds, err := parseIntList(*countThresholds)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: -count-thresholds:", err)
		return 1
	}

	result, err := structuralreliability.Run(structuralreliability.Config{
		InputPath: *input, ClassesPath: *classes, Folds: *folds, FoldSeed: *foldSeed,
		MinTokenCount: *minCount, Neighbors: *neighbors, BootstrapRuns: *bootstrapRuns, BootstrapSeed: *bootstrapSeed,
		Threshold: *threshold, ThresholdMargin: *thresholdMargin, CountThresholds: thresholds,
		SubsampleMinFullCount: *subsampleMinFullCount, SubsampleRuns: *subsampleRuns, SubsampleSeed: *subsampleSeed,
		Progress: func(message string) { fmt.Println(message) },
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	data, err := yaml.Marshal(result)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: encode output:", err)
		return 1
	}
	if err := workdir.EnsureParent(*output); err != nil {
		fmt.Fprintln(os.Stderr, "Error: create output directory:", err)
		return 1
	}
	if err := os.WriteFile(*output, data, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "Error: write output:", err)
		return 1
	}
	fmt.Printf("Structural reliability written to %s\n", *output)
	return 0
}

func parseIntList(raw string) ([]int, error) {
	var result []int
	for _, field := range strings.Split(raw, ",") {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		value, err := strconv.Atoi(field)
		if err != nil {
			return nil, fmt.Errorf("invalid integer %q: %w", field, err)
		}
		result = append(result, value)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("must list at least one threshold")
	}
	return result, nil
}
