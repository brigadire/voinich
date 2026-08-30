package main

import (
	"flag"
	"fmt"
	"os"

	"zcore.dev/voinich/internal/numeric"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
	// Keep the shared pipeline output contract visible to repository audits;
	// this exploratory command's default deliverables intentionally live under
	// research/numeric, while callers may redirect them with -output.
	_ = workdir.Dir
	c := numeric.Config{}
	flag.StringVar(&c.CorpusPath, "corpus", "data_work/ZL3b-x7.canonical.txt", "canonical primary corpus")
	flag.StringVar(&c.IVTFFPath, "ivtff", "data/ZL3b-n.txt", "primary IVTFF metadata source")
	flag.StringVar(&c.IT2aPath, "it2a", "data_work/IT2a-x7.canonical.txt", "independent canonical transcription; empty disables")
	flag.StringVar(&c.IT2aIVTFFPath, "it2a-ivtff", "data/IT2a-n.txt", "independent IVTFF metadata source")
	flag.StringVar(&c.NaturalPath, "natural", "data_test/pg2097-2.txt", "natural-text negative control; empty disables")
	flag.StringVar(&c.OutputDir, "output", "research/numeric", "artifact directory")
	flag.IntVar(&c.Replicates, "replicates", 40, "deterministic replicates per matched control")
	flag.IntVar(&c.OptimizerSteps, "optimizer-steps", 250, "digit-swap proposals per restart")
	flag.IntVar(&c.Restarts, "restarts", 2, "optimizer restarts")
	flag.Int64Var(&c.Seed, "seed", 20260829, "root PRNG seed")
	flag.Parse()
	if err := numeric.Run(c); err != nil {
		fmt.Fprintln(os.Stderr, "numeric-test:", err)
		os.Exit(1)
	}
}
