package main

import "strconv"

// Stage describes one pipeline command in Task36's single orchestration
// CLI. Every stage's own CLI defaults already are the frozen/production
// parameters established by the task that introduced it (window sizes, K
// ranges, permutation counts, seeds - see each package's own doc comments
// for the freeze rationale); this table therefore adds only the small set
// of purely OPERATIONAL flags (quiet output, checkpoint path, executor/
// worker concurrency) that Task28-35 already made explicit as non-
// scientific. It never overrides a scientific parameter: doing so would
// defeat the entire point of a frozen baseline.
type Stage struct {
	// Name is the stable identifier used for the built binary
	// (workdir/bin/<Name>), log file name, and run-state key.
	Name string
	// SourceDir is the repo-relative package to `go build`. "." is the
	// root dictionary generator.
	SourceDir string
	// Positional carries required positional arguments in order (only the
	// two earliest, pre-flag-package commands use these).
	Positional []string
	// Quiet is true when the binary accepts -quiet (suppresses the
	// interactive progress bar, which would otherwise fill log files with
	// carriage-return spam).
	Quiet bool
	// Checkpoint is true when the binary accepts -checkpoint-path
	// (Task28-35's within-stage resume mechanism). The orchestrator lets
	// every checkpointed stage use its own default
	// <output-dir>/checkpoint.json rather than relocating it, since that
	// default already lives under the same workdir/ subtree this
	// orchestrator snapshots at freeze time.
	Checkpoint bool
	// Executor is true only for conditional-regime-analyze (Task31-34):
	// the sole stage with a distributed/local executor choice
	// (-executor, -workers).
	Executor bool
}

// stages is the complete, ordered (topologically sorted by dependency)
// Task36 production pipeline: 27 commands from the raw IVTFF corpus
// through every validation/discovery stage added by tasks 1-27. Order
// matters - every stage's default input path assumes every earlier stage
// in this list has already written its default output.
var stages = []Stage{
	{Name: "dict-gen", SourceDir: ".", Positional: []string{"data_work/ZL3b-x7.txt", "workdir/dataset/dictionary.yaml"}},
	{Name: "dict-analyze", SourceDir: "dict-analyze", Positional: []string{"workdir/dataset/dictionary.yaml", "workdir/dataset/tokens_analysis.yaml"}},
	{Name: "structural-analyze", SourceDir: "structural-analyze"},
	{Name: "sequence-analyze", SourceDir: "sequence-analyze"},
	{Name: "begin-end-analyze", SourceDir: "begin-end-analyze"},
	{Name: "structural-normalize", SourceDir: "structural-normalize"},
	{Name: "normalization-compare", SourceDir: "normalization-compare"},
	{Name: "structural-validate", SourceDir: "structural-validate"},
	{Name: "structural-profile-stability", SourceDir: "structural-profile-stability"},
	{Name: "structural-reliability", SourceDir: "structural-reliability"},
	{Name: "soft-structural-space", SourceDir: "soft-structural-space"},
	{Name: "structural-graphemic", SourceDir: "structural-graphemic"},
	{Name: "structural-pair-decompose", SourceDir: "structural-pair-decompose"},
	{Name: "distance-context-analyze", SourceDir: "distance-context-analyze"},
	{Name: "local-regime-analyze", SourceDir: "local-regime-analyze", Quiet: true},
	{Name: "property-trajectory-analyze", SourceDir: "property-trajectory-analyze", Quiet: true},
	{Name: "structural-projection-analyze", SourceDir: "structural-projection-analyze", Quiet: true},
	{Name: "global-regime-analyze", SourceDir: "global-regime-analyze", Quiet: true},
	{Name: "metadata-validate", SourceDir: "metadata-validate", Quiet: true},
	{Name: "cluster-metadata-global", SourceDir: "cluster-metadata-global", Quiet: true},
	{Name: "conditional-regime-analyze", SourceDir: "conditional-regime-analyze", Quiet: true, Checkpoint: true, Executor: true},
	{Name: "residual-diagnostic-analyze", SourceDir: "residual-diagnostic-analyze", Quiet: true},
	{Name: "token-relation-validate", SourceDir: "token-relation-validate", Quiet: true, Checkpoint: true},
	{Name: "replicated-local-structure-audit", SourceDir: "replicated-local-structure-audit", Quiet: true, Checkpoint: true},
	{Name: "higher-order-sequence-validate", SourceDir: "higher-order-sequence-validate", Quiet: true, Checkpoint: true},
	{Name: "positional-continuation-validate", SourceDir: "positional-continuation-validate", Quiet: true, Checkpoint: true},
	{Name: "transition-network-validate", SourceDir: "transition-network-validate", Quiet: true, Checkpoint: true},
}

// stageByName looks up a stage by its orchestrator Name, used by -only/
// -skip filtering and resume logic.
func stageByName(name string) (Stage, bool) {
	for _, s := range stages {
		if s.Name == name {
			return s, true
		}
	}
	return Stage{}, false
}

// orchestratorOptions carries the small set of purely operational choices
// that vary between a local production run and a run that spreads
// conditional-regime-analyze's permutation jobs across real mTLS-
// authenticated remote workers (Task33/34). None of these affect any
// scientific parameter.
type orchestratorOptions struct {
	Executor       string // goroutine|process|remote
	LocalWorkers   int
	RemoteListen   string
	TLSCert        string
	TLSKey         string
	ClientCA       string
	RemoteDenyList string
	RemoteTimeout  string
	RemoteRetries  int
}

// stageArgs builds the exact argument list one stage will be invoked with.
// Every argument here is operational (quiet output, executor/worker
// concurrency, mTLS transport); no scientific flag is ever set, so every
// stage's own frozen defaults remain in effect.
func stageArgs(s Stage, opt orchestratorOptions) []string {
	args := append([]string(nil), s.Positional...)
	if s.Quiet {
		args = append(args, "-quiet")
	}
	if s.Executor {
		executor := opt.Executor
		if executor == "" {
			executor = "goroutine"
		}
		args = append(args, "-executor", executor)
		if executor != "goroutine" {
			args = append(args, "-workers", strconv.Itoa(opt.LocalWorkers))
		}
		if executor == "remote" {
			args = append(args, "-remote-listen", opt.RemoteListen, "-tls-cert", opt.TLSCert, "-tls-key", opt.TLSKey, "-client-ca", opt.ClientCA)
			if opt.RemoteDenyList != "" {
				args = append(args, "-remote-deny-list", opt.RemoteDenyList)
			}
			if opt.RemoteTimeout != "" {
				args = append(args, "-remote-timeout", opt.RemoteTimeout)
			}
			if opt.RemoteRetries > 0 {
				args = append(args, "-remote-retries", strconv.Itoa(opt.RemoteRetries))
			}
		}
	}
	return args
}
