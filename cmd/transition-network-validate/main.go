package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"zcore.dev/voinich/internal/conditionalregime"
	"zcore.dev/voinich/internal/profiling"
	"zcore.dev/voinich/internal/transitionnetwork"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
	os.Exit(run())
}

func run() (code int) {
	start := time.Now()
	c := transitionnetwork.Config{}
	flag.StringVar(&c.CorpusPath, "corpus", "data_work/ZL3b-x7.txt", "IVTT -x7 tokenized corpus")
	flag.StringVar(&c.MetadataPath, "token-metadata-map", workdir.Path("metadata-validation/token_metadata_map.tsv"), "per-token metadata TSV")
	flag.StringVar(&c.OutputDir, "output-dir", workdir.Path("transition-network"), "result directory")
	flag.IntVar(&c.MinTokenCount, "min-token-count", 10, "primary global token count")
	flag.IntVar(&c.MinBlockTokenCount, "min-block-token-count", 5, "minimum source opportunities in a block")
	flag.IntVar(&c.Permutations, "permutations", 1000, "primary within-block permutations")
	flag.IntVar(&c.RefinePermutations, "refine-permutations", 10000, "total permutations for pre-specified refinement")
	flag.Int64Var(&c.Seed, "seed", 1, "deterministic random seed")
	flag.StringVar(&c.CheckpointPath, "checkpoint-path", "", "checkpoint file (default <output-dir>/checkpoint.json; '-' disables)")
	flag.BoolVar(&c.Quiet, "quiet", false, "disable status bar")
	flag.BoolVar(&c.Generic, "generic-corpus", false, "task43 generic mode: derive blocks from -corpus alone (internal/genericsegmentation) instead of -token-metadata-map; -token-metadata-map is ignored, and the Currier/hand metadata-transfer diagnostic is skipped (no generic analogue)")
	flag.StringVar(&c.Executor, "executor", "goroutine", "permutation-null backend: goroutine, process, or remote")
	flag.IntVar(&c.Workers, "workers", 1, "maximum concurrent local workers/processes")
	flag.StringVar(&c.RemoteListen, "remote-listen", "", "coordinator mTLS listen address")
	flag.StringVar(&c.TLSCert, "tls-cert", "", "coordinator/worker TLS certificate")
	flag.StringVar(&c.TLSKey, "tls-key", "", "coordinator/worker TLS private key")
	flag.StringVar(&c.ClientCA, "client-ca", "", "CA used to authenticate workers")
	flag.StringVar(&c.RemoteDenyList, "remote-deny-list", "", "optional worker certificate deny list")
	flag.DurationVar(&c.RemoteTimeout, "remote-timeout", 10*time.Minute, "remote replicate lease timeout")
	flag.IntVar(&c.RemoteRetries, "remote-retries", 3, "remote lease retries")
	var internalWorker bool
	flag.BoolVar(&internalWorker, "internal-worker", false, "internal persistent worker protocol")
	prof := profiling.RegisterFlags(flag.CommandLine)
	flag.Parse()
	if c.CheckpointPath == "" {
		c.CheckpointPath = filepath.Join(c.OutputDir, "checkpoint.json")
	}

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
		ex, err := conditionalregime.NewTransitionNetworkProcessExecutor(c)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return 1
		}
		c.PermutationExecutor = ex
	case "remote":
		ex, err := conditionalregime.NewTransitionNetworkRemoteExecutor(c)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return 1
		}
		c.PermutationExecutor = ex
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

	if err := transitionnetwork.RunAndWrite(c); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	return 0
}
