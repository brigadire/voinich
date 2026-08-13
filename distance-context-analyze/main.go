package main

import (
	"flag"
	"fmt"
	"os"

	"zcore.dev/voinich/internal/distancecontext"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
	c := distancecontext.Config{}
	flag.StringVar(&c.CorpusPath, "corpus", "data_work/ZL3b-x7.txt", "IVTT -x7 corpus in original line order")
	flag.StringVar(&c.DistantPath, "distant-pairs", workdir.Path("structural_distant_top.tsv"), "ranked structural-distant pair TSV")
	flag.StringVar(&c.FamiliesPath, "families", workdir.Path("structural_distant_families.yaml"), "structural-distant family YAML")
	flag.StringVar(&c.ControlsPath, "controls", workdir.Path("pair_controls.tsv"), "matched negative-control TSV")
	flag.StringVar(&c.OutputDir, "output-dir", workdir.Dir, "result directory")
	flag.IntVar(&c.MaxDistance, "max-distance", 20, "largest exact context distance")
	flag.IntVar(&c.MinObservations, "min-observations", 30, "minimum observations per side for a reliable comparison")
	flag.IntVar(&c.TopN, "top", 50, "number of ranked structural-distant pairs; 0 disables ranked additions")
	flag.StringVar(&c.Pair, "pair", "", "analyze only tokenA,tokenB")
	flag.IntVar(&c.FamilyID, "family", 0, "analyze only one family ID")
	flag.BoolVar(&c.RespectLineBoundaries, "respect-line-boundaries", false, "make line-bounded mode primary (both modes are always computed)")
	flag.Parse()
	if err := distancecontext.RunAndWrite(c); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
