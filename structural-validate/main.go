package main

import (
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
	"zcore.dev/voinich/internal/validation"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
	input := flag.String("input", "data_work/ZL3b-x7.txt", "IVTT -x7 derived corpus")
	classes := flag.String("classes", workdir.Path("structural_classes.yaml"), "full-corpus classes used only for leave-one-class-out")
	output := flag.String("output", workdir.Path("structural_validation.yaml"), "validation YAML")
	folds := flag.Int("folds", 5, "number of line-based cross-validation folds")
	foldSeed := flag.Int64("fold-seed", 1, "deterministic fold split seed")
	threshold := flag.Float64("threshold", .70, "fixed complete-link threshold")
	minTokenCount := flag.Int("min-token-count", 10, "minimum TRAIN token count")
	randomBaselines := flag.Int("random-baselines", 100, "TRAIN-derived matched random models per fold")
	randomSeed := flag.Int64("random-seed", 1, "base matched-random seed")
	minN := flag.Int("min-n", 2, "minimum n-gram length")
	maxN := flag.Int("max-n", 8, "maximum n-gram length")
	maxContext := flag.Int("max-context-length", 7, "maximum context length")
	flag.Parse()

	result, err := validation.Run(validation.Config{
		InputPath: *input, ClassesPath: *classes, Folds: *folds, FoldSeed: *foldSeed,
		Threshold: *threshold, MinTokenCount: *minTokenCount,
		RandomBaselines: *randomBaselines, RandomSeed: *randomSeed,
		MinN: *minN, MaxN: *maxN, MaxContext: *maxContext,
		Progress: func(message string) { fmt.Println(message) },
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	data, err := yaml.Marshal(result)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: encode output:", err)
		os.Exit(1)
	}
	if err := workdir.EnsureParent(*output); err != nil {
		fmt.Fprintln(os.Stderr, "Error: create output directory:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(*output, data, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "Error: write output:", err)
		os.Exit(1)
	}
	fmt.Printf("Validation written to %s\n", *output)
}
