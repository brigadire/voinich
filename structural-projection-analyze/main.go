package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"zcore.dev/voinich/internal/profiling"
	"zcore.dev/voinich/internal/structuralprojection"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
	os.Exit(run())
}

func run() (code int) {
	start := time.Now()
	c := structuralprojection.Config{}
	flag.StringVar(&c.CorpusPath, "corpus", "data_work/ZL3b-x7.txt", "IVTT -x7 tokenized corpus")
	flag.StringVar(&c.StructuralPairsPath, "structural-pairs", workdir.Path("soft_structural_pairs.tsv"), "soft structural pair TSV")
	flag.StringVar(&c.DistancePairsPath, "distance-pairs", workdir.Path("distance_context_pairs.yaml"), "existing token-level distance analysis")
	flag.StringVar(&c.FamiliesPath, "families", workdir.Path("structural_distant_families.yaml"), "structural-distant families")
	flag.StringVar(&c.OutputDir, "output-dir", workdir.Dir, "result directory")
	flag.Float64Var(&c.MinStructuralSimilarity, "min-structural-similarity", .65, "minimum structural similarity")
	flag.Float64Var(&c.MinReliability, "min-reliability", .70, "minimum evidence reliability")
	flag.IntVar(&c.ProjectionK, "projection-k", 0, "use K nearest structural neighbours instead of thresholding (0 disables)")
	flag.StringVar(&c.ProjectionMode, "projection-mode", "both", "primary ranking mode: full, ablated, or both (both metrics are retained)")
	flag.IntVar(&c.RandomProjections, "random-projections", 200, "random structural-space controls")
	flag.Int64Var(&c.Seed, "seed", 130013, "deterministic random seed")
	flag.IntVar(&c.MaxDistance, "max-distance", 20, "largest exact distance")
	flag.IntVar(&c.MinObservations, "min-observations", 30, "reliability observation threshold")
	flag.IntVar(&c.TopN, "top", 28, "number of pairs from previous analysis")
	flag.StringVar(&c.Pair, "pair", "", "analyze only tokenA,tokenB")
	flag.IntVar(&c.FamilyID, "family", 0, "analyze only one family ID")
	flag.BoolVar(&c.Quiet, "quiet", false, "disable progress output")
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

	if err := structuralprojection.RunAndWrite(c); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	return 0
}
