package main

import (
	"flag"
	"fmt"
	"os"

	"zcore.dev/voinich/internal/positionalcontinuation"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
	c := positionalcontinuation.Config{}
	flag.StringVar(&c.CorpusPath, "corpus", "data_work/ZL3b-x7.txt", "IVTT -x7 tokenized corpus")
	flag.StringVar(&c.TokenMetadataMap, "token-metadata-map", workdir.Path("metadata-validation", "token_metadata_map.tsv"), "per-token metadata map")
	flag.StringVar(&c.HigherOrderDir, "higher-order-dir", workdir.Path("higher-order-sequences"), "previous higher-order-sequence-validate output directory")
	flag.StringVar(&c.OutputDir, "output-dir", workdir.Path("positional-continuation"), "result directory")
	flag.StringVar(&c.CheckpointPath, "checkpoint-path", "", "checkpoint file path (default <output-dir>/checkpoint.json; '-' disables)")
	flag.IntVar(&c.Permutations, "permutations", 10000, "permutations for every positional/stratified permutation test")
	flag.Int64Var(&c.Seed, "seed", 1, "deterministic random seed")
	flag.BoolVar(&c.Quiet, "quiet", false, "disable progress output")
	flag.Parse()
	if err := positionalcontinuation.RunAndWrite(c); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
