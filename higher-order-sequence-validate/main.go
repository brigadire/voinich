package main

import (
	"flag"
	"fmt"
	"os"

	"zcore.dev/voinich/internal/higherorderseq"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
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
	flag.Parse()
	if err := higherorderseq.RunAndWrite(c); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
