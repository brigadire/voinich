package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"zcore.dev/voinich/internal/conditionalregime"
	"zcore.dev/voinich/internal/profiling"
	"zcore.dev/voinich/internal/workdir"
)

type intList []int

func (l *intList) String() string {
	s := make([]string, len(*l))
	for i, v := range *l {
		s[i] = strconv.Itoa(v)
	}
	return strings.Join(s, ",")
}
func (l *intList) Set(s string) error {
	*l = nil
	for _, part := range strings.Split(s, ",") {
		v, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return fmt.Errorf("invalid integer %q", part)
		}
		*l = append(*l, v)
	}
	return nil
}

func main() {
	os.Exit(run())
}

func run() (code int) {
	start := time.Now()
	c := conditionalregime.Config{}
	var windowSizes, residualWindowSizes intList
	var internalWorker bool
	var workerCA, workerCoordinator, remoteCache string
	var remoteConcurrency int
	flag.StringVar(&c.CorpusPath, "corpus", "data_work/ZL3b-x7.txt", "canonical IVTT -x7 corpus, unchanged")
	flag.StringVar(&c.TokenMetadataMap, "token-metadata-map", "workdir/metadata-validation/token_metadata_map.tsv", "frozen token_metadata_map.tsv from metadata-validate")
	flag.StringVar(&c.OutputDir, "output-dir", workdir.Path("conditional-regimes"), "result directory")
	flag.Var(&windowSizes, "window-sizes", "frozen within-class scales (Part A)")
	flag.Var(&residualWindowSizes, "residual-window-sizes", "frozen pooled residual scales (Part B)")
	flag.IntVar(&c.MinClassTokens, "min-class-tokens", 1000, "minimum total tokens for a class to be eligible")
	flag.IntVar(&c.MinBlockTokens, "min-block-tokens", 500, "minimum largest contiguous block for a class to be eligible")
	flag.IntVar(&c.KMin, "k-min", 2, "minimum K")
	flag.IntVar(&c.KMaxWithin, "k-max-within", 10, "maximum K for within-class discovery (Part A)")
	flag.IntVar(&c.KMaxResidual, "k-max-residual", 15, "maximum K for pooled residual clustering (Part B)")
	flag.IntVar(&c.Permutations, "permutations", 1000, "primary block-aware null permutations")
	flag.Int64Var(&c.Seed, "seed", 1, "deterministic random seed")
	flag.IntVar(&c.Workers, "workers", 1, "bounded local permutation workers")
	flag.StringVar(&c.Executor, "executor", "goroutine", "permutation job backend: goroutine|process|remote")
	flag.StringVar(&c.RemoteListen, "remote-listen", "", "coordinator mTLS bind address for -executor remote (for example 0.0.0.0:8443)")
	flag.StringVar(&c.TLSCert, "tls-cert", "", "this node's certificate: coordinator.crt (serverAuth) in coordinator mode, worker-N.crt (clientAuth) with -coordinator; see conditional-regime-pki")
	flag.StringVar(&c.TLSKey, "tls-key", "", "this node's private key, matching -tls-cert")
	flag.StringVar(&c.ClientCA, "client-ca", "", "project CA bundle used to verify connecting workers' client certificates (coordinator mode)")
	flag.StringVar(&c.RemoteDenyList, "remote-deny-list", "", "optional JSON deny-list revoking worker certificate serials/identities (coordinator mode)")
	flag.DurationVar(&c.RemoteTimeout, "remote-timeout", 10*time.Minute, "per-lease deadline before the coordinator reassigns a job to another worker")
	flag.IntVar(&c.RemoteRetries, "remote-retries", 2, "reassignments per job after a lease expires unanswered")
	flag.StringVar(&workerCoordinator, "coordinator", "", "run as an mTLS worker dialing this coordinator HTTPS URL instead of running the pipeline")
	flag.StringVar(&workerCA, "ca", "", "project CA bundle used to verify the coordinator's certificate (worker mode)")
	flag.StringVar(&remoteCache, "remote-cache-dir", "", "content-addressed remote worker input cache (worker mode)")
	flag.IntVar(&remoteConcurrency, "remote-concurrency", 1, "concurrent lease loops run by this worker process")
	flag.StringVar(&c.CheckpointPath, "checkpoint-path", "", "progress checkpoint file (default <output-dir>/checkpoint.json; \"-\" disables checkpointing)")
	flag.BoolVar(&c.Quiet, "quiet", false, "disable status bar")
	flag.BoolVar(&internalWorker, "internal-worker", false, "internal: speak the Task32 subprocess worker protocol on stdin/stdout instead of running the pipeline (do not use directly)")
	prof := profiling.RegisterFlags(flag.CommandLine)
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if internalWorker {
		if err := conditionalregime.RunWorker(ctx, os.Stdin, os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return 1
		}
		return 0
	}
	if workerCoordinator != "" {
		if workerCA == "" || c.TLSCert == "" || c.TLSKey == "" {
			fmt.Fprintln(os.Stderr, "Error: -ca, -tls-cert and -tls-key are required in worker mode")
			return 2
		}
		if remoteCache == "" {
			fmt.Fprintln(os.Stderr, "Error: -remote-cache-dir is required in worker mode")
			return 2
		}
		if err := conditionalregime.RunRemoteWorker(ctx, workerCoordinator, workerCA, c.TLSCert, c.TLSKey, remoteCache, remoteConcurrency); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return 1
		}
		return 0
	}

	c.WindowSizes = []int(windowSizes)
	c.ResidualWindowSizes = []int(residualWindowSizes)
	if c.Permutations < 1 {
		fmt.Fprintln(os.Stderr, "Error: permutations must be positive")
		return 2
	}
	if c.Workers < 1 {
		fmt.Fprintln(os.Stderr, "Error: workers must be positive")
		return 2
	}
	if c.Executor != "goroutine" && c.Executor != "process" && c.Executor != "remote" {
		fmt.Fprintln(os.Stderr, `Error: executor must be "goroutine", "process" or "remote"`)
		return 2
	}
	if c.Executor == "remote" && (c.RemoteListen == "" || c.TLSCert == "" || c.TLSKey == "" || c.ClientCA == "") {
		fmt.Fprintln(os.Stderr, "Error: -executor remote requires -remote-listen, -tls-cert, -tls-key and -client-ca")
		return 2
	}
	c.Context = ctx

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

	if err := conditionalregime.RunAndWrite(c); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	return 0
}
