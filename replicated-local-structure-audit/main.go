package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"zcore.dev/voinich/internal/replicatedlocalaudit"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
	c := replicatedlocalaudit.Config{}
	flag.StringVar(&c.CorpusPath, "corpus", "data_work/ZL3b-x7.txt", "canonical tokenized corpus")
	flag.StringVar(&c.MetadataPath, "token-metadata-map", workdir.Path("metadata-validation/token_metadata_map.tsv"), "token metadata map TSV")
	flag.StringVar(&c.RelationDir, "relation-dir", workdir.Path("token-relation-validation"), "frozen token relation validation directory")
	flag.StringVar(&c.DiscoveryDir, "discovery-dir", workdir.Dir, "frozen discovery directory")
	flag.StringVar(&c.OutputDir, "output-dir", workdir.Path("replicated-local-structure"), "result directory")
	flag.IntVar(&c.Permutations, "permutations", 1000, "confirmatory null replicates")
	flag.Int64Var(&c.Seed, "seed", 1, "deterministic random seed")
	flag.StringVar(&c.CheckpointPath, "checkpoint-path", "", "progress checkpoint (default <output-dir>/checkpoint.json; '-' disables)")
	flag.BoolVar(&c.Quiet, "quiet", false, "disable status bar")
	flag.Parse()
	if c.CheckpointPath == "" {
		c.CheckpointPath = filepath.Join(c.OutputDir, "checkpoint.json")
	} else if c.CheckpointPath == "-" {
		c.CheckpointPath = ""
	}
	if err := replicatedlocalaudit.RunAndWrite(c); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
