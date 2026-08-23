package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"zcore.dev/voinich/internal/sequenceanalyze"
	"zcore.dev/voinich/internal/vocabularygrowth"
	"zcore.dev/voinich/internal/workdir"
)

type ints []int

func (x *ints) String() string {
	a := make([]string, len(*x))
	for i, v := range *x {
		a[i] = strconv.Itoa(v)
	}
	return strings.Join(a, ",")
}
func (x *ints) Set(s string) error {
	for _, p := range strings.Split(s, ",") {
		v, e := strconv.Atoi(strings.TrimSpace(p))
		if e != nil || v <= 0 {
			return fmt.Errorf("invalid positive integer %q", p)
		}
		*x = append(*x, v)
	}
	return nil
}

func main() {
	fs := flag.NewFlagSet("vocabulary-growth-analyze", flag.ExitOnError)
	input := fs.String("input", "data_work/ZL3b-x7.txt", "canonical corpus")
	out := fs.String("output-dir", workdir.Path("vocabulary-growth"), "output directory")
	var cps, windows, segments ints
	fs.Var(&cps, "checkpoint", "checkpoint (repeatable or comma-separated; default adaptive grid)")
	fs.Var(&windows, "window-size", "new-type window size (repeatable; default 500,1000,2000)")
	fs.Var(&segments, "segments", "positional segment count (repeatable; default 4,8)")
	nulls := fs.Int("null-permutations", 100, "deterministic shuffled-token null replicates")
	seed := fs.Int64("seed", 1, "base RNG seed")
	executor := fs.String("executor", "goroutine", "executor; null ensemble is intentionally single-machine")
	fitMin := fs.Int("fit-min-n", 0, "minimum checkpoint used for Heaps fit; 0 means all")
	fitMax := fs.Int("fit-max-n", 0, "maximum checkpoint used for Heaps fit; 0 means all")
	fs.Parse(os.Args[1:])
	if *executor != "goroutine" {
		fmt.Fprintln(os.Stderr, "vocabulary-growth-analyze: remote/process execution is not enabled; profiling audit found O(N) core and bounded null ensemble unsuitable for distribution by default")
		os.Exit(2)
	}
	tokens, err := sequenceanalyze.ReadTokens(*input)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read corpus:", err)
		os.Exit(1)
	}
	p := vocabularygrowth.DefaultParameters()
	if len(cps) > 0 {
		p.Checkpoints = cps
	}
	if len(windows) > 0 {
		p.WindowSizes = windows
	}
	if len(segments) > 0 {
		p.SegmentCounts = segments
	}
	p.NullPermutations = *nulls
	p.Seed = *seed
	p.FitMinN = *fitMin
	p.FitMaxN = *fitMax
	r, err := vocabularygrowth.Analyze(tokens, p)
	if err != nil {
		fmt.Fprintln(os.Stderr, "analyze:", err)
		os.Exit(1)
	}
	if err = vocabularygrowth.Write(r, *out); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		os.Exit(1)
	}
	fmt.Printf("Vocabulary growth analysis written to %s (%d tokens)\n", *out, len(tokens))
}
