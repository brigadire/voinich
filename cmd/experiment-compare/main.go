package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
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
	dirty := false
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" {
				commit = s.Value
			}
			if s.Key == "vcs.modified" {
				dirty = s.Value == "true"
			}
		}
	}
	if commit == "unknown" {
		if out, err := exec.Command("git", "rev-parse", "HEAD").Output(); err == nil {
			commit = strings.TrimSpace(string(out))
		}
	}
	if out, err := exec.Command("git", "status", "--porcelain").Output(); err == nil {
		dirty = len(strings.TrimSpace(string(out))) > 0
	}
	if err := experimentcompare.Run(experimentcompare.Options{Experiments: exps, OutputDir: *out, AllowUnfrozen: *allow, Args: os.Args[1:], GitCommit: commit, GitDirty: dirty}); err != nil {
		fmt.Fprintln(os.Stderr, "experiment-compare:", err)
		os.Exit(1)
	}
	fmt.Println("comparison written to", *out)
}
