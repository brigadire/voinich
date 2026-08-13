package main

import (
	"flag"
	"fmt"
	"os"
	"zcore.dev/voinich/internal/propertytrajectory"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
	c := propertytrajectory.Config{}
	flag.StringVar(&c.CorpusPath, "corpus", "data_work/ivtt_output_1786282555007.txt", "tokenized linear corpus")
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
	flag.Parse()
	if e := propertytrajectory.RunAndWrite(c); e != nil {
		fmt.Fprintln(os.Stderr, "Error:", e)
		os.Exit(1)
	}
}
