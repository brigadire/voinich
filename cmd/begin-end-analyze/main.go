package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"zcore.dev/voinich/internal/beginendanalyze"
	"zcore.dev/voinich/internal/conditionalregime"
	"zcore.dev/voinich/internal/profiling"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
	os.Exit(run())
}

func run() (code int) {
	start := time.Now()
	c := beginendanalyze.Config{}
	flag.StringVar(&c.DictionaryPath, "dictionary", workdir.Path("dataset", "dictionary.yaml"), "input YAML dictionary")
	flag.StringVar(&c.CorpusPath, "corpus", "data_work/ZL3b-x7.txt", "IVTT -x7 linear corpus")
	flag.IntVar(&c.MaxWindow, "max-window", 55, "maximum token-distance window")
	flag.IntVar(&c.Permutations, "permutations", 100, "number of boundary-preserving permutations")
	flag.IntVar(&c.MinTokenCount, "min-frequency", 10, "minimum token frequency")
	flag.Int64Var(&c.RandomSeed, "random-seed", 1, "random seed")
	flag.StringVar(&c.OutputDir, "output-dir", workdir.Dir, "result directory")
	flag.StringVar(&c.PermutationMode, "permutation-mode", "page", "page or line")
	flag.BoolVar(&c.IncludeUnclear, "include-unclear", false, "include tokens containing ? in main ranking")
	flag.IntVar(&c.MaxCandidates, "max-candidates", 1000, "maximum non-local candidates in YAML; 0 means unlimited")
	flag.IntVar(&c.CandidateBatchSize, "candidate-batch-size", beginendanalyze.DefaultCandidateBatchSize, "candidate-pair work-unit size for -executor process/remote (Task47)")
	flag.StringVar(&c.Executor, "executor", "goroutine", "candidate-pair backend: goroutine, process, or remote")
	flag.IntVar(&c.Workers, "workers", 1, "maximum concurrent local workers/processes")
	flag.StringVar(&c.RemoteListen, "remote-listen", "", "coordinator mTLS listen address")
	flag.StringVar(&c.TLSCert, "tls-cert", "", "coordinator/worker TLS certificate")
	flag.StringVar(&c.TLSKey, "tls-key", "", "coordinator/worker TLS private key")
	flag.StringVar(&c.ClientCA, "client-ca", "", "CA used to authenticate workers")
	flag.StringVar(&c.RemoteDenyList, "remote-deny-list", "", "optional worker certificate deny list")
	flag.DurationVar(&c.RemoteTimeout, "remote-timeout", 10*time.Minute, "remote batch lease timeout")
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
		ex, err := conditionalregime.NewBeginEndProcessExecutor(c)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return 1
		}
		c.BatchExecutor = ex
	case "remote":
		ex, err := conditionalregime.NewBeginEndRemoteExecutor(c)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return 1
		}
		c.BatchExecutor = ex
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

	result, err := beginendanalyze.RunAndWrite(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	if err := os.MkdirAll(c.OutputDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		return 1
	}
	if err := beginendanalyze.WriteReports(c.OutputDir, result); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing reports: %v\n", err)
		return 1
	}
	fmt.Printf("Reports written to %s, %s and %s\n", filepath.Join(c.OutputDir, "begin_end_candidates.yaml"), filepath.Join(c.OutputDir, "begin_end_top.tsv"), filepath.Join(c.OutputDir, "begin_end_report.md"))
	return 0
}
