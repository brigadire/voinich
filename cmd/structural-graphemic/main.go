package main

import (
	"flag"
	"fmt"
	"os"

	"zcore.dev/voinich/internal/graphemic"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
	input := flag.String("input", workdir.Path("soft_structural_pairs.tsv"), "existing structural pair TSV")
	output := flag.String("output-dir", workdir.Dir, "result directory")
	minStructural := flag.Float64("min-structural-similarity", .65, "minimum unchanged structural similarity")
	minReliability := flag.Float64("min-reliability", .7, "minimum evidence reliability")
	minDistance := flag.Float64("min-graphemic-distance", .6, "minimum normalized graphemic distance")
	minClose := flag.Float64("min-graphemic-similarity", .75, "minimum graphemic similarity for control and families")
	top := flag.Int("top", 200, "maximum rows in each ranked TSV; 0 means unlimited")
	flag.Parse()
	cfg := graphemic.Config{InputPath: *input, OutputDir: *output, MinStructuralSimilarity: *minStructural, MinReliability: *minReliability, MinGraphemicDistance: *minDistance, MinCloseSimilarity: *minClose, TopN: *top}
	r, err := graphemic.Analyze(cfg)
	if err == nil {
		err = graphemic.WriteAll(cfg, r)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Analyzed %d tokens and %d pairs; reports written to %s\n", r.TokenCount, len(r.Pairs), *output)
}
