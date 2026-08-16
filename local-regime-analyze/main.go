package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"zcore.dev/voinich/internal/localregime"
	"zcore.dev/voinich/internal/profiling"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
	os.Exit(run())
}

func run() (code int) {
	start := time.Now()
	c := localregime.Config{}
	flag.StringVar(&c.CorpusPath, "corpus", "data_work/ZL3b-x7.txt", "IVTT -x7 linear corpus")
	flag.StringVar(&c.DistancePairsPath, "distance-pairs", workdir.Path("distance_context_pairs.yaml"), "previous distance-context pair YAML")
	flag.StringVar(&c.ControlsPath, "controls", workdir.Path("distance_context_controls.tsv"), "existing matched controls (optional)")
	flag.StringVar(&c.OutputDir, "output-dir", workdir.Dir, "result directory")
	flag.IntVar(&c.RegimeRadius, "regime-radius", 100, "primary outer local-regime radius")
	flag.IntVar(&c.RegimeGap, "regime-gap", 20, "primary exclusion gap")
	flag.IntVar(&c.RegimeControlsK, "regime-controls-k", 5, "matched occurrences per occurrence")
	flag.IntVar(&c.WindowStep, "window-step", 10, "sliding-window step")
	flag.IntVar(&c.MaxDistance, "max-distance", 20, "largest exact rightward distance")
	flag.IntVar(&c.TopN, "top", 28, "number of pairs from previous analysis")
	flag.StringVar(&c.Pair, "pair", "", "analyze only tokenA,tokenB")
	flag.Int64Var(&c.Seed, "seed", 150015, "deterministic shuffle seed")
	flag.BoolVar(&c.RespectLineBoundaries, "respect-line-boundaries", false, "use line-bounded profiles as the primary diagnostic")
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

	if err := localregime.RunAndWrite(c); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	return 0
}
