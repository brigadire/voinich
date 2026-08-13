package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"zcore.dev/voinich/internal/metadatavalidation"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
	c := metadatavalidation.Config{}
	var tolerances string
	flag.StringVar(&c.IVTFFPath, "ivtff", "data/ZL3b-n.txt", "original IVTFF source used only for metadata")
	flag.StringVar(&c.FrozenCorpusPath, "frozen-corpus", "data_work/ZL3b-x7.txt", "canonical IVTT -x7 ASCII Full corpus derived from the IVTFF source")
	flag.StringVar(&c.DiscoveryDir, "discovery-dir", workdir.Dir, "directory containing frozen discovery results")
	flag.StringVar(&c.OutputDir, "output-dir", workdir.Dir, "result directory")
	flag.IntVar(&c.Permutations, "permutations", 10000, "deterministic null permutations")
	flag.Int64Var(&c.Seed, "seed", 1, "deterministic random seed")
	flag.StringVar(&tolerances, "boundary-tolerances", "10,25,50,100,200", "fixed token tolerances")
	flag.BoolVar(&c.Quiet, "quiet", false, "disable status bar")
	flag.Parse()
	for _, s := range strings.Split(tolerances, ",") {
		v, e := strconv.Atoi(strings.TrimSpace(s))
		if e != nil || v < 0 {
			fmt.Fprintln(os.Stderr, "Error: boundary-tolerances must be non-negative comma-separated integers")
			os.Exit(2)
		}
		c.Tolerances = append(c.Tolerances, v)
	}
	if c.Permutations < 1 {
		fmt.Fprintln(os.Stderr, "Error: permutations must be positive")
		os.Exit(2)
	}
	if e := metadatavalidation.RunAndWrite(c); e != nil {
		fmt.Fprintln(os.Stderr, "Error:", e)
		os.Exit(1)
	}
}
