package main

import (
	"flag"
	"fmt"
	"os"

	"zcore.dev/voinich/internal/localregime"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
	c := localregime.Config{}
	flag.StringVar(&c.CorpusPath, "corpus", "data_work/ivtt_output_1786282555007.txt", "tokenized linear corpus")
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
	flag.Parse()
	if err := localregime.RunAndWrite(c); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
