package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"zcore.dev/voinich/internal/higherorderseq"
	"zcore.dev/voinich/internal/profiling"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
	os.Exit(run())
}

func run() (code int) {
	start := time.Now()
	c := higherorderseq.Config{}
	flag.StringVar(&c.CorpusPath, "corpus", "data_work/ZL3b-x7.txt", "IVTT -x7 tokenized corpus")
	flag.StringVar(&c.TokenMetadataMap, "token-metadata-map", workdir.Path("metadata-validation", "token_metadata_map.tsv"), "per-token metadata map")
	flag.StringVar(&c.AuditDir, "audit-dir", workdir.Path("replicated-local-structure"), "previous replicated-local-structure-audit output directory")
	flag.StringVar(&c.DiscoveryDir, "discovery-dir", workdir.Dir, "root pipeline output directory (for structural_classes.yaml)")
	flag.StringVar(&c.OutputDir, "output-dir", workdir.Path("higher-order-sequences"), "result directory")
	flag.StringVar(&c.CheckpointPath, "checkpoint-path", "", "checkpoint file path (default <output-dir>/checkpoint.json; '-' disables)")
	flag.IntVar(&c.Permutations, "permutations", 10000, "primary-family conditional-neighbor permutations (secondary descriptive candidate uses 1/10th)")
	flag.Int64Var(&c.Seed, "seed", 1, "deterministic random seed")
	flag.BoolVar(&c.Quiet, "quiet", false, "disable progress output")
	flag.BoolVar(&c.Generic, "generic-corpus", false, "task43 generic mode: derive blocks from -corpus alone (internal/genericsegmentation) instead of -token-metadata-map; -token-metadata-map is ignored, and -audit-dir must be replicated-local-structure-audit's own generic-mode output")
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

	if err := higherorderseq.RunAndWrite(c); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	return 0
}
