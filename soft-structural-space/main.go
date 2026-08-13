package main

import (
	"flag"
	"fmt"
	"os"

	"zcore.dev/voinich/internal/softstructural"
)

func main() {
	c := softstructural.Config{}
	flag.StringVar(&c.DictionaryPath, "dictionary", "dataset/dictionary.yaml", "path to dictionary YAML")
	flag.StringVar(&c.AnalysisPath, "analysis", "dataset/tokens_analysis.yaml", "path to token analysis YAML")
	flag.StringVar(&c.ReliabilityPath, "reliability", "structural_reliability.yaml", "path to structural reliability YAML")
	flag.StringVar(&c.OutputPath, "output", "soft_structural_space.yaml", "path to summary YAML")
	flag.StringVar(&c.PairsPath, "pairs-output", "soft_structural_pairs.tsv", "path to machine-oriented pair TSV")
	flag.IntVar(&c.MinTokenCount, "min-token-count", 10, "minimum token occurrence count")
	flag.IntVar(&c.Neighbors, "neighbors", 5, "number of neighbors and mutual-neighbor K")
	flag.Float64Var(&c.MinEvidenceStrength, "min-evidence-strength", .7, "presentation filter for high-evidence raw neighbors")
	flag.Float64Var(&c.GraphMinSimilarity, "graph-min-similarity", .6, "presentation filter for graph edges")
	flag.Parse()
	if err := softstructural.ValidateConfig(c); err != nil {
		fmt.Fprintln(os.Stderr, "Invalid parameters:", err)
		os.Exit(2)
	}
	out, pairs, err := softstructural.BuildAll(c)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	if err = softstructural.WriteTSV(c.PairsPath, pairs); err == nil {
		err = softstructural.WriteYAML(c.OutputPath, out)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error writing output:", err)
		os.Exit(1)
	}
	fmt.Printf("YAML written to %s; %d pairs written to %s\n", c.OutputPath, len(pairs), c.PairsPath)
}
