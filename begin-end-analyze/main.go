package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"zcore.dev/voinich/internal/workdir"
)

func main() {
	dictionary := flag.String("dictionary", workdir.Path("dataset", "dictionary.yaml"), "input YAML dictionary")
	corpusPath := flag.String("corpus", "data_work/ZL3b-x7.txt", "IVTT -x7 linear corpus")
	maxWindow := flag.Int("max-window", 55, "maximum token-distance window")
	permutations := flag.Int("permutations", 100, "number of boundary-preserving permutations")
	minFrequency := flag.Int("min-frequency", 10, "minimum token frequency")
	randomSeed := flag.Int64("random-seed", 1, "random seed")
	outputDir := flag.String("output-dir", workdir.Dir, "result directory")
	permutationMode := flag.String("permutation-mode", "page", "page or line")
	includeUnclear := flag.Bool("include-unclear", false, "include tokens containing ? in main ranking")
	maxCandidates := flag.Int("max-candidates", 1000, "maximum non-local candidates in YAML; 0 means unlimited")
	flag.Parse()

	parameters := Parameters{MaxWindow: *maxWindow, Permutations: *permutations, MinTokenCount: *minFrequency, RandomSeed: *randomSeed, PermutationMode: *permutationMode, IncludeUnclear: *includeUnclear, MaxCandidates: *maxCandidates}
	result, err := runAnalysis(*dictionary, *corpusPath, parameters)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}
	if err := writeReports(*outputDir, result); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing reports: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Reports written to %s, %s and %s\n", filepath.Join(*outputDir, "begin_end_candidates.yaml"), filepath.Join(*outputDir, "begin_end_top.tsv"), filepath.Join(*outputDir, "begin_end_report.md"))
}
