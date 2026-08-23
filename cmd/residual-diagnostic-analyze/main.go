package main

import (
	"flag"
	"fmt"
	"os"

	"zcore.dev/voinich/internal/residualdiagnostic"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
	c := residualdiagnostic.Config{}
	flag.StringVar(&c.ConditionalDir, "conditional-dir", workdir.Path("conditional-regimes"), "existing conditional-regime-analyze result directory")
	flag.StringVar(&c.CorpusPath, "corpus", "data_work/ZL3b-x7.txt", "canonical unchanged corpus")
	flag.StringVar(&c.MetadataPath, "token-metadata-map", workdir.Path("metadata-validation", "token_metadata_map.tsv"), "frozen token metadata map")
	flag.StringVar(&c.OutputDir, "output-dir", workdir.Path("residual-diagnostics"), "diagnostic output directory")
	flag.IntVar(&c.WindowSize, "window-size", 500, "frozen winning residual window size")
	flag.IntVar(&c.K, "k", 2, "frozen winning residual K")
	flag.IntVar(&c.Permutations, "permutations", 1000, "block-aware diagnostic permutations")
	flag.Int64Var(&c.Seed, "seed", 1, "deterministic seed")
	flag.BoolVar(&c.Quiet, "quiet", false, "disable status bar")
	flag.Parse()
	if c.WindowSize < 1 || c.K < 2 || c.Permutations < 1 {
		fmt.Fprintln(os.Stderr, "Error: window-size and permutations must be positive, k must be at least 2")
		os.Exit(2)
	}
	if err := residualdiagnostic.RunAndWrite(c); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
