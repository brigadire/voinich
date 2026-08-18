package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"zcore.dev/voinich/internal/profiling"
	"zcore.dev/voinich/internal/replicatedlocalaudit"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
	os.Exit(run())
}

func run() (code int) {
	start := time.Now()
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
	flag.BoolVar(&c.Generic, "generic-corpus", false, "task43 generic mode: derive blocks from -corpus alone (internal/genericsegmentation) instead of -token-metadata-map; -token-metadata-map is ignored, and -relation-dir must be token-relation-validate's own generic-mode output")
	prof := profiling.RegisterFlags(flag.CommandLine)
	flag.Parse()
	if c.CheckpointPath == "" {
		c.CheckpointPath = filepath.Join(c.OutputDir, "checkpoint.json")
	} else if c.CheckpointPath == "-" {
		c.CheckpointPath = ""
	}

	defer profiling.PrintElapsed(os.Stderr, start)

	sess, err := profiling.Start(prof)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	defer func() {
		if err := sess.Stop(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			code = 1
		}
	}()

	if err := replicatedlocalaudit.RunAndWrite(c); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	return 0
}
