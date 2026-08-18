package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"zcore.dev/voinich/internal/positionalcontinuation"
	"zcore.dev/voinich/internal/profiling"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
	os.Exit(run())
}

func run() (code int) {
	start := time.Now()
	c := positionalcontinuation.Config{}
	flag.StringVar(&c.CorpusPath, "corpus", "data_work/ZL3b-x7.txt", "IVTT -x7 tokenized corpus")
	flag.StringVar(&c.TokenMetadataMap, "token-metadata-map", workdir.Path("metadata-validation", "token_metadata_map.tsv"), "per-token metadata map")
	flag.StringVar(&c.HigherOrderDir, "higher-order-dir", workdir.Path("higher-order-sequences"), "previous higher-order-sequence-validate output directory")
	flag.StringVar(&c.OutputDir, "output-dir", workdir.Path("positional-continuation"), "result directory")
	flag.StringVar(&c.CheckpointPath, "checkpoint-path", "", "checkpoint file path (default <output-dir>/checkpoint.json; '-' disables)")
	flag.IntVar(&c.Permutations, "permutations", 10000, "permutations for every positional/stratified permutation test")
	flag.Int64Var(&c.Seed, "seed", 1, "deterministic random seed")
	flag.BoolVar(&c.Quiet, "quiet", false, "disable progress output")
	flag.BoolVar(&c.Generic, "generic-corpus", false, "task43 generic mode: derive blocks from -corpus alone (internal/genericsegmentation) instead of -token-metadata-map, and substitute the frozen s/aiin/chey target with the top-ranked candidate from -higher-order-dir's own generic-mode output; -token-metadata-map is ignored, and -higher-order-dir must be higher-order-sequence-validate's own generic-mode output")
	prof := profiling.RegisterFlags(flag.CommandLine)
	flag.Parse()

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

	if err := positionalcontinuation.RunAndWrite(c); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	return 0
}
