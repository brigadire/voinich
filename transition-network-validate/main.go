package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"zcore.dev/voinich/internal/transitionnetwork"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
	c := transitionnetwork.Config{}
	flag.StringVar(&c.CorpusPath, "corpus", "data_work/ZL3b-x7.txt", "IVTT -x7 tokenized corpus")
	flag.StringVar(&c.MetadataPath, "token-metadata-map", workdir.Path("metadata-validation/token_metadata_map.tsv"), "per-token metadata TSV")
	flag.StringVar(&c.OutputDir, "output-dir", workdir.Path("transition-network"), "result directory")
	flag.IntVar(&c.MinTokenCount, "min-token-count", 10, "primary global token count")
	flag.IntVar(&c.MinBlockTokenCount, "min-block-token-count", 5, "minimum source opportunities in a block")
	flag.IntVar(&c.Permutations, "permutations", 1000, "primary within-block permutations")
	flag.IntVar(&c.RefinePermutations, "refine-permutations", 10000, "total permutations for pre-specified refinement")
	flag.Int64Var(&c.Seed, "seed", 1, "deterministic random seed")
	flag.StringVar(&c.CheckpointPath, "checkpoint-path", "", "checkpoint file (default <output-dir>/checkpoint.json; '-' disables)")
	flag.BoolVar(&c.Quiet, "quiet", false, "disable status bar")
	flag.Parse()
	if c.CheckpointPath == "" {
		c.CheckpointPath = filepath.Join(c.OutputDir, "checkpoint.json")
	}
	if err := transitionnetwork.RunAndWrite(c); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
