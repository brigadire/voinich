package main

import (
	"flag"
	"fmt"
	"os"

	"zcore.dev/voinich/internal/clustermetadataglobal"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
	c := clustermetadataglobal.Config{}
	flag.StringVar(&c.DiscoveryDir, "discovery-dir", workdir.Dir, "directory containing frozen global-regime-analyze discovery results")
	flag.StringVar(&c.TokenMetadataMap, "token-metadata-map", workdir.Path("token_metadata_map.tsv"), "token_metadata_map.tsv produced by metadata-validate")
	flag.StringVar(&c.MetadataReportPath, "metadata-report", workdir.Path("metadata_validation_report.md"), "metadata_validation_report.md to add the global-correction section to")
	flag.StringVar(&c.OutputDir, "output-dir", workdir.Dir, "result directory")
	flag.IntVar(&c.Permutations, "permutations", 10000, "deterministic block-aware null permutations, shared across the entire frozen search space per replicate")
	flag.Int64Var(&c.Seed, "seed", 1, "deterministic random seed")
	flag.BoolVar(&c.Quiet, "quiet", false, "disable status bar")
	flag.Parse()
	if c.Permutations < 1 {
		fmt.Fprintln(os.Stderr, "Error: permutations must be positive")
		os.Exit(2)
	}
	if err := clustermetadataglobal.RunAndWrite(c); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
