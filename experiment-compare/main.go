package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"zcore.dev/voinich/internal/experimentcompare"
	"zcore.dev/voinich/internal/workdir"
)

type experimentsFlag []string

func (f *experimentsFlag) String() string { return strings.Join(*f, ",") }
func (f *experimentsFlag) Set(v string) error {
	if strings.TrimSpace(v) == "" {
		return fmt.Errorf("empty -experiment")
	}
	*f = append(*f, v)
	return nil
}

func main() {
	fs := flag.NewFlagSet("experiment-compare", flag.ExitOnError)
	var exps experimentsFlag
	fs.Var(&exps, "experiment", "frozen experiment directory (repeatable)")
	out := fs.String("output-dir", workdir.Path("comparisons", "comparative-v1"), "comparison output directory")
	allow := fs.Bool("allow-unfrozen", false, "development-only: allow an explicit warning for non-frozen inputs")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: experiment-compare -experiment DIR -experiment DIR -output-dir DIR [-allow-unfrozen]")
		fs.PrintDefaults()
	}
	fs.Parse(os.Args[1:])
	commit := "unknown"
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" {
				commit = s.Value
			}
		}
	}
	if err := experimentcompare.Run(experimentcompare.Options{Experiments: exps, OutputDir: *out, AllowUnfrozen: *allow, Args: os.Args[1:], GitCommit: commit}); err != nil {
		fmt.Fprintln(os.Stderr, "experiment-compare:", err)
		os.Exit(1)
	}
	fmt.Println("comparison written to", *out)
}
