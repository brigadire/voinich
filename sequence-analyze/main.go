package main

import (
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
	input := flag.String("input", "data_work/ZL3b-x7.txt", "IVTT -x7 derived corpus")
	output := flag.String("output", workdir.Path("sequence_analysis.yaml"), "output YAML")
	minN := flag.Int("min-n", 2, "minimum n-gram length")
	maxN := flag.Int("max-n", 8, "maximum n-gram length")
	minCount := flag.Int("min-count", 2, "minimum count for repeated sequence sections")
	maxItems := flag.Int("max-items", 200, "maximum records per n; 0 means unlimited")
	contextLimit := flag.Int("context-limit", 10, "maximum displayed context tokens")
	maxContextLength := flag.Int("max-context-length", 7, "maximum left-context length for next-token analysis")
	contextMinObservations := flag.Int("context-min-observations", 10, "minimum long-context observations for context extensions")
	contextMaxItems := flag.Int("context-max-items", 200, "maximum context-extension records; 0 means unlimited")
	flag.Parse()

	parameters := Parameters{
		MinN: *minN, MaxN: *maxN, MinCount: *minCount, MaxItems: *maxItems, ContextLimit: *contextLimit,
		MaxContextLength: *maxContextLength, ContextMinObservations: *contextMinObservations, ContextMaxItems: *contextMaxItems,
	}
	result, err := analyzeFile(*input, parameters)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error analyzing corpus: %v\n", err)
		os.Exit(1)
	}
	data, err := yaml.Marshal(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding YAML: %v\n", err)
		os.Exit(1)
	}
	if err := workdir.EnsureParent(*output); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(*output, data, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("YAML file written to %s\n", *output)
}
