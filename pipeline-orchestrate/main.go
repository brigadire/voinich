// Command pipeline-orchestrate is Task36's single orchestration CLI: it
// runs every current Voynich pipeline stage in dependency order against
// production parameters, using conditional-regime-analyze's distributed
// executor where that stage supports it, and freezes the result as an
// immutable, checksummed experiment directory.
//
// It never invents a parallel configuration system: every stage's own CLI
// defaults already are its frozen production parameters (see stages.go),
// so this tool adds only operational flags (quiet output, executor/worker
// concurrency, mTLS transport) - never a scientific one.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"zcore.dev/voinich/internal/workdir"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func repoRoot() string {
	// pipeline-orchestrate always ships and runs from the repo it
	// orchestrates; workdir.Dir is resolved relative to the current
	// working directory, which callers are expected to set to the repo
	// root (matching every other command in this repository).
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

func run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}
	switch args[0] {
	case "manifest":
		return runManifestCmd(args[1:])
	case "run":
		return runRunCmd(args[1:])
	case "freeze":
		return runFreezeCmd(args[1:])
	case "verify":
		return runVerifyCmd(args[1:])
	case "status":
		return runStatusCmd(args[1:])
	case "-h", "-help", "--help", "help":
		usage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand %q\n\n", args[0])
		usage()
		return 2
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `pipeline-orchestrate: Task36 single orchestration CLI

Usage:
  pipeline-orchestrate manifest -experiment-dir experiments/voynich-v1 [-ivtff FILE] [-corpus FILE] [-generic-corpus] [-executor process|goroutine|remote] [-workers N] [-remote-listen ADDR -tls-cert C -tls-key K -client-ca CA] [-force]
  pipeline-orchestrate run      -experiment-dir experiments/voynich-v1 [-only STAGE]
  pipeline-orchestrate freeze   -experiment-dir experiments/voynich-v1 [-force]
  pipeline-orchestrate verify   -experiment-dir experiments/voynich-v1
  pipeline-orchestrate status   -experiment-dir experiments/voynich-v1

Run from the repository root. "manifest" must be run first (immutable,
written once); "run" executes every stage in order and supports resume;
"freeze" snapshots workdir/'s outputs, checksums them, writes REPORT.md,
and marks the experiment directory FROZEN - refusing any further run/freeze
without -force. "verify" recomputes checksums against a frozen baseline.
`)
}

func runManifestCmd(args []string) int {
	fs := flag.NewFlagSet("manifest", flag.ContinueOnError)
	experimentDir := fs.String("experiment-dir", "experiments/voynich-v1", "experiment directory to create")
	ivtff := fs.String("ivtff", "data/ZL3b-n.txt", "frozen IVTFF source")
	corpus := fs.String("corpus", "data_work/ZL3b-x7.txt", "frozen IVTT -x7 corpus derivative")
	genericCorpus := fs.Bool("generic-corpus", false, "treat -corpus as an authoritative generic token corpus; do not read IVTFF metadata")
	executor := fs.String("executor", "process", "distributed stage executor: goroutine|process|remote")
	workers := fs.Int("workers", 8, "worker slots for distributed-capable stages")
	remoteListen := fs.String("remote-listen", "", "coordinator mTLS listen address (executor=remote)")
	tlsCert := fs.String("tls-cert", "", "coordinator certificate (executor=remote)")
	tlsKey := fs.String("tls-key", "", "coordinator private key (executor=remote)")
	clientCA := fs.String("client-ca", "", "project CA bundle (executor=remote)")
	remoteTimeout := fs.String("remote-timeout", "", "per-lease deadline (executor=remote)")
	remoteRetries := fs.Int("remote-retries", 0, "lease reassignment retries (executor=remote)")
	var remoteWorkers stringList
	fs.Var(&remoteWorkers, "remote-worker", "authenticated remote worker identity, for the manifest's worker list (repeatable, executor=remote)")
	force := fs.Bool("force", false, "overwrite an existing (non-frozen) manifest")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	corpusExplicit := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "corpus" {
			corpusExplicit = true
		}
	})
	if *genericCorpus && !corpusExplicit {
		fmt.Fprintln(os.Stderr, "Error: -generic-corpus requires an explicit -corpus (the Voynich default is never used as fallback)")
		return 2
	}
	if *executor == "remote" && (*remoteListen == "" || *tlsCert == "" || *tlsKey == "" || *clientCA == "") {
		fmt.Fprintln(os.Stderr, "Error: -executor remote requires -remote-listen, -tls-cert, -tls-key and -client-ca")
		return 2
	}
	if frozen, _ := isFrozen(*experimentDir); frozen {
		fmt.Fprintf(os.Stderr, "Error: %s is FROZEN; refusing to overwrite its manifest\n", *experimentDir)
		return 1
	}
	if _, err := os.Stat(manifestPath(*experimentDir)); err == nil && !*force {
		fmt.Fprintf(os.Stderr, "Error: %s already exists (use -force to overwrite; never do this after run has started)\n", manifestPath(*experimentDir))
		return 1
	}

	opt := orchestratorOptions{
		Executor:      *executor,
		LocalWorkers:  *workers,
		RemoteListen:  *remoteListen,
		TLSCert:       *tlsCert,
		TLSKey:        *tlsKey,
		ClientCA:      *clientCA,
		RemoteTimeout: *remoteTimeout,
		RemoteRetries: *remoteRetries,
	}
	repo := repoRoot()
	inputMode := "ivtff"
	if *genericCorpus {
		inputMode = "generic"
	}
	m, err := buildManifest(repo, inputMode, *ivtff, *corpus, opt, remoteWorkers)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	if err := saveManifest(*experimentDir, m); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	if err := saveRunState(*experimentDir, newRunStateForManifest(m)); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	fmt.Printf("Wrote %s\nExperimentID: %s\nGit commit: %s (dirty=%v)\nExecutor: %s\n", manifestPath(*experimentDir), m.ExperimentID, m.GitCommit, m.GitDirty, m.Executor)
	return 0
}

func runRunCmd(args []string) int {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	experimentDir := fs.String("experiment-dir", "experiments/voynich-v1", "experiment directory (must already have a manifest)")
	only := fs.String("only", "", "run only this one stage (by name), ignoring resume state for it")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	m, err := loadManifest(*experimentDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: load manifest (run `manifest` first):", err)
		return 1
	}
	opt := orchestratorOptionsFromManifest(m)
	if err := runPipeline(repoRoot(), *experimentDir, opt, *only); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	return 0
}

// orchestratorOptionsFromManifest reconstructs the operational flags each
// stage was run with from the frozen manifest's conditional-regime-analyze
// entry, so `run` (including a resumed run) never depends on the caller
// re-specifying the executor choice made at `manifest` time.
func orchestratorOptionsFromManifest(m *Manifest) orchestratorOptions {
	opt := orchestratorOptions{Executor: m.Executor}
	for _, st := range m.Stages {
		if st.Name != "conditional-regime-analyze" {
			continue
		}
		for i := 0; i < len(st.Args); i++ {
			switch st.Args[i] {
			case "-workers":
				if i+1 < len(st.Args) {
					fmt.Sscanf(st.Args[i+1], "%d", &opt.LocalWorkers)
				}
			case "-remote-listen":
				opt.RemoteListen = valueAfter(st.Args, i)
			case "-tls-cert":
				opt.TLSCert = valueAfter(st.Args, i)
			case "-tls-key":
				opt.TLSKey = valueAfter(st.Args, i)
			case "-client-ca":
				opt.ClientCA = valueAfter(st.Args, i)
			case "-remote-deny-list":
				opt.RemoteDenyList = valueAfter(st.Args, i)
			case "-remote-timeout":
				opt.RemoteTimeout = valueAfter(st.Args, i)
			case "-remote-retries":
				fmt.Sscanf(valueAfter(st.Args, i), "%d", &opt.RemoteRetries)
			}
		}
	}
	return opt
}

func valueAfter(args []string, i int) string {
	if i+1 < len(args) {
		return args[i+1]
	}
	return ""
}

func runFreezeCmd(args []string) int {
	fs := flag.NewFlagSet("freeze", flag.ContinueOnError)
	experimentDir := fs.String("experiment-dir", "experiments/voynich-v1", "experiment directory")
	force := fs.Bool("force", false, "re-freeze an already-FROZEN experiment (overwrites the existing baseline)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if err := freezeExperiment(*experimentDir, *force); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	return 0
}

func runVerifyCmd(args []string) int {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	experimentDir := fs.String("experiment-dir", "experiments/voynich-v1", "experiment directory")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if err := verifyExperiment(*experimentDir); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	return 0
}

func runStatusCmd(args []string) int {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	experimentDir := fs.String("experiment-dir", "experiments/voynich-v1", "experiment directory")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	rs, err := loadRunState(*experimentDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	frozen, _ := isFrozen(*experimentDir)
	fmt.Printf("Experiment %s (frozen=%v)\n", rs.ExperimentID, frozen)
	for i, sr := range rs.Stages {
		fmt.Printf("[%2d] %-35s %-10s %.1fs\n", i+1, sr.Name, sr.Status, sr.DurationSeconds)
	}
	return 0
}

type stringList []string

func (l *stringList) String() string { return filepath.Join(*l...) }
func (l *stringList) Set(s string) error {
	*l = append(*l, s)
	return nil
}

var _ = workdir.Dir // referenced by run.go/freeze.go: workdir/ is this tool's INPUT snapshot source, never its output root (see experiments/ instead)
