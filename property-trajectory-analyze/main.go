package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"zcore.dev/voinich/internal/profiling"
	"zcore.dev/voinich/internal/propertytrajectory"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
	os.Exit(run())
}

func run() (code int) {
	start := time.Now()
	c := propertytrajectory.Config{}
	flag.StringVar(&c.CorpusPath, "corpus", "data_work/ZL3b-x7.txt", "IVTT -x7 linear corpus")
	flag.StringVar(&c.StructuralPairsPath, "structural-pairs", workdir.Path("soft_structural_pairs.tsv"), "soft structural pair TSV (centrality properties only)")
	flag.StringVar(&c.DistancePairsPath, "distance-pairs", workdir.Path("distance_context_pairs.yaml"), "previous distance-context pair YAML")
	flag.StringVar(&c.ControlsPath, "controls", workdir.Path("distance_context_controls.tsv"), "existing frequency/reliability-matched controls")
	flag.StringVar(&c.OutputDir, "output-dir", workdir.Dir, "result directory")
	flag.IntVar(&c.MinTokenFrequency, "min-token-frequency", 10, "minimum frequency for an eligible subsequent token")
	flag.IntVar(&c.MaxDistance, "max-distance", 20, "largest exact rightward distance")
	flag.IntVar(&c.TopN, "top", 28, "number of previous distance-context pairs (required pairs are also retained)")
	flag.StringVar(&c.Pair, "pair", "", "analyze only tokenA,tokenB")
	flag.IntVar(&c.RandomPairs, "random-pairs", 1000, "frequency-matched random pairs per target")
	flag.Int64Var(&c.Seed, "seed", 140014, "deterministic random seed")
	flag.BoolVar(&c.Quiet, "quiet", false, "disable status bar")
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

	if e := propertytrajectory.RunAndWrite(c); e != nil {
		fmt.Fprintln(os.Stderr, "Error:", e)
		return 1
	}
	return 0
}
