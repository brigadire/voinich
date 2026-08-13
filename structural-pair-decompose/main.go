package main

import (
	"flag"
	"fmt"
	"os"

	"zcore.dev/voinich/internal/pairdecomposition"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
	c := pairdecomposition.Config{}
	flag.StringVar(&c.DictionaryPath, "dictionary", workdir.Path("dataset", "dictionary.yaml"), "path to the full dictionary YAML")
	flag.StringVar(&c.PairsPath, "pairs", workdir.Path("structural_graphemic_pairs.tsv"), "path to the full structural/graphemic pair TSV")
	flag.StringVar(&c.DistantPath, "distant", workdir.Path("structural_distant_top.tsv"), "ranked structural-distant pair TSV")
	flag.StringVar(&c.FamiliesPath, "families", workdir.Path("structural_distant_families.yaml"), "structural-distant family YAML")
	flag.StringVar(&c.OutputDir, "output-dir", workdir.Dir, "directory for reports and plots")
	flag.IntVar(&c.TopN, "top", 50, "number of ranked pairs; 0 means all")
	flag.StringVar(&c.Pair, "pair", "", "analyze one pair as tokenA,tokenB")
	flag.IntVar(&c.FamilyID, "family", 0, "analyze one family ID; 0 means all in default mode")
	flag.IntVar(&c.ContextLimit, "context-limit", 12, "rows per human-readable context list")
	flag.IntVar(&c.Controls, "controls", 3, "negative controls per target")
	flag.Parse()
	if err := pairdecomposition.ValidateConfig(c); err != nil {
		fmt.Fprintln(os.Stderr, "Invalid parameters:", err)
		os.Exit(2)
	}
	out, families, err := pairdecomposition.Run(c)
	if err == nil {
		err = pairdecomposition.WriteAll(c, out, families)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	fmt.Printf("Decomposed %d pairs and %d families into %s\n", len(out.Pairs), len(families), c.OutputDir)
}
