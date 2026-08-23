package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"zcore.dev/voinich/internal/conditionalregime"
	"zcore.dev/voinich/internal/normalizationcompare"
	"zcore.dev/voinich/internal/profiling"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
	os.Exit(run())
}

func run() (code int) {
	start := time.Now()
	c := normalizationcompare.Config{}
	flag.StringVar(&c.ClassesPath, "classes", workdir.Path("structural_classes.yaml"), "structural class-map YAML")
	flag.StringVar(&c.InputPath, "input", "data_work/ZL3b-x7.txt", "IVTT -x7 corpus for random baselines")
	flag.StringVar(&c.RawAnalysisPath, "raw-analysis", workdir.Path("sequence_analysis.yaml"), "immutable raw sequence analysis")
	flag.StringVar(&c.NormalizedPattern, "normalized-pattern", workdir.Path("normalized_%s.txt"), "normalized corpus path pattern")
	flag.StringVar(&c.AnalysisPattern, "analysis-pattern", workdir.Path("sequence_analysis_%s.yaml"), "structural sequence-analysis output pattern")
	flag.StringVar(&c.SequenceAnalyzerPath, "sequence-analyzer", workdir.Path("bin", "sequence-analyze"), "compiled sequence-analyze executable (recorded in output only; analysis always runs in-process)")
	flag.StringVar(&c.OutputPath, "output", workdir.Path("normalization_comparison.yaml"), "comparison YAML")
	flag.IntVar(&c.RandomRuns, "random-baselines", 100, "matched random runs per threshold")
	flag.Int64Var(&c.RandomSeed, "random-seed", 1, "base random seed")
	flag.StringVar(&c.Executor, "executor", "goroutine", "random-baseline trial backend: goroutine, process, or remote")
	flag.IntVar(&c.Workers, "workers", 1, "maximum concurrent local workers/processes")
	flag.StringVar(&c.RemoteListen, "remote-listen", "", "coordinator mTLS listen address")
	flag.StringVar(&c.TLSCert, "tls-cert", "", "coordinator/worker TLS certificate")
	flag.StringVar(&c.TLSKey, "tls-key", "", "coordinator/worker TLS private key")
	flag.StringVar(&c.ClientCA, "client-ca", "", "CA used to authenticate workers")
	flag.StringVar(&c.RemoteDenyList, "remote-deny-list", "", "optional worker certificate deny list")
	flag.DurationVar(&c.RemoteTimeout, "remote-timeout", 10*time.Minute, "remote baseline lease timeout")
	flag.IntVar(&c.RemoteRetries, "remote-retries", 3, "remote lease retries")
	var internalWorker bool
	flag.BoolVar(&internalWorker, "internal-worker", false, "internal persistent worker protocol")
	prof := profiling.RegisterFlags(flag.CommandLine)
	flag.Parse()

	if internalWorker {
		if err := conditionalregime.RunWorker(context.Background(), os.Stdin, os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return 1
		}
		return 0
	}

	if c.RandomRuns < 1 {
		fmt.Fprintln(os.Stderr, "Error: random-baselines must be at least 1")
		return 2
	}
	if c.Workers < 1 {
		fmt.Fprintln(os.Stderr, "Error: workers must be positive")
		return 2
	}
	if c.Executor != "" && c.Executor != "goroutine" && c.Executor != "process" && c.Executor != "remote" {
		fmt.Fprintln(os.Stderr, `Error: executor must be "goroutine", "process" or "remote"`)
		return 2
	}
	if c.Executor == "remote" && (c.RemoteListen == "" || c.TLSCert == "" || c.TLSKey == "" || c.ClientCA == "") {
		fmt.Fprintln(os.Stderr, "Error: -executor remote requires -remote-listen, -tls-cert, -tls-key and -client-ca")
		return 2
	}
	c.Context = context.Background()

	switch c.Executor {
	case "process":
		ex, err := conditionalregime.NewNormalizationProcessExecutor(c)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return 1
		}
		c.BaselineExecutor = ex
	case "remote":
		ex, err := conditionalregime.NewNormalizationRemoteExecutor(c)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return 1
		}
		c.BaselineExecutor = ex
	}

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

	if err := normalizationcompare.RunAndWrite(c); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	return 0
}
