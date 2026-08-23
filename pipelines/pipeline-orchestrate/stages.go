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
	// Checkpoint is true when the binary has a within-stage resume mechanism.
	// Most checkpointed stages use their own default. Structural projection's
	// explicit path is removed after successful output, so it cannot enter a
	// frozen experiment.
	Checkpoint bool
	// Executor marks stages with a distributed/local executor choice
	// (-executor, -workers).
	Executor bool
	// CorpusFlag is the flag used to override the historical Voynich corpus
	// default.  An empty value means that the stage consumes only upstream
	// artifacts (dict-gen is handled through its positional argument).
	CorpusFlag string
	// RequiresMetadata marks stages whose actual inputs include IVTFF-derived
	// folio/hand/Currier metadata, directly or through token_metadata_map.tsv.
	// This alone no longer decides generic-mode applicability - see Generic
	// below, task43's per-stage audit (GENERIC_STAGE_APPLICABILITY_AUDIT.md).
	RequiresMetadata bool
	// Generic is task43's per-stage classification of RequiresMetadata
	// stages for generic (non-IVTFF) corpora. Stages with RequiresMetadata
	// false are unconditionally generic-applicable and Generic is the zero
	// value (unused).
	Generic GenericSupport
}

// GenericSupport records, for one RequiresMetadata stage, task43's audit
// verdict on whether it has a scientifically legitimate corpus-only mode.
// See GENERIC_STAGE_APPLICABILITY_AUDIT.md for the full per-stage evidence.
type GenericSupport struct {
	// Applicable is false for Class A stages (the IVTFF metadata identity
	// is itself the thing being tested, not a segmentation device) - they
	// stay NOT_APPLICABLE in generic mode, with Reason explaining exactly
	// why, never the generic "requires IVTFF metadata" phrase. It is true
	// for Class B/C stages, whose block-based statistics never actually
	// read metadata identity - only stageArgsForIsolatedInput appends
	// -generic-corpus for those.
	Applicable bool
	// Reason is the precise scientific reason a Class A stage cannot have
	// a generic mode. Required when Applicable is false; unused otherwise.
	Reason string
}

// stages is the complete, ordered (topologically sorted by dependency)
// Task49 production pipeline: 28 commands from the raw IVTFF corpus
// through every validation/discovery stage added by tasks 1-27. Order
// matters - every stage's default input path assumes every earlier stage
// in this list has already written its default output.
var stages = []Stage{
	{Name: "dict-gen", SourceDir: "cmd/voinich", Positional: []string{"data_work/ZL3b-x7.txt", "workdir/dataset/dictionary.yaml"}},
	{Name: "dict-analyze", SourceDir: "cmd/dict-analyze", Positional: []string{"workdir/dataset/dictionary.yaml", "workdir/dataset/tokens_analysis.yaml"}},
	{Name: "structural-analyze", SourceDir: "cmd/structural-analyze"},
	{Name: "sequence-analyze", SourceDir: "cmd/sequence-analyze", CorpusFlag: "-input"},
	{Name: "begin-end-analyze", SourceDir: "cmd/begin-end-analyze", Executor: true, CorpusFlag: "-corpus"},
	{Name: "structural-normalize", SourceDir: "cmd/structural-normalize", CorpusFlag: "-input"},
	{Name: "normalization-compare", SourceDir: "cmd/normalization-compare", Executor: true, CorpusFlag: "-input"},
	{Name: "structural-validate", SourceDir: "cmd/structural-validate", CorpusFlag: "-input"},
	{Name: "structural-profile-stability", SourceDir: "cmd/structural-profile-stability", CorpusFlag: "-input"},
	{Name: "structural-reliability", SourceDir: "cmd/structural-reliability", CorpusFlag: "-input"},
	{Name: "soft-structural-space", SourceDir: "cmd/soft-structural-space"},
	{Name: "structural-graphemic", SourceDir: "cmd/structural-graphemic"},
	{Name: "structural-pair-decompose", SourceDir: "cmd/structural-pair-decompose"},
	{Name: "distance-context-analyze", SourceDir: "cmd/distance-context-analyze", CorpusFlag: "-corpus"},
	{Name: "local-regime-analyze", SourceDir: "cmd/local-regime-analyze", Quiet: true, CorpusFlag: "-corpus"},
	{Name: "property-trajectory-analyze", SourceDir: "cmd/property-trajectory-analyze", Quiet: true, CorpusFlag: "-corpus"},
	{Name: "structural-projection-analyze", SourceDir: "cmd/structural-projection-analyze", Quiet: true, Checkpoint: true, Executor: true, CorpusFlag: "-corpus"},
	{Name: "global-regime-analyze", SourceDir: "cmd/global-regime-analyze", Quiet: true, CorpusFlag: "-corpus"},
	{Name: "metadata-validate", SourceDir: "cmd/metadata-validate", Quiet: true, CorpusFlag: "-frozen-corpus", RequiresMetadata: true, Generic: GenericSupport{Reason: "validates blind boundary/cluster discovery against real external manuscript ground truth (line, paragraph, folio, Currier-language, scribal-hand, quire boundaries parsed from IVTFF); a generic corpus has no independent external segmentation record to validate against, so the analysis has no referent"}},
	{Name: "cluster-metadata-global", SourceDir: "cmd/cluster-metadata-global", Quiet: true, RequiresMetadata: true, Generic: GenericSupport{Reason: "performs the whole-search-space multiple-comparison-corrected test of association between real Currier-language/scribal-hand identity and blind distributional clustering; Currier and hand are Voynich-manuscript-specific external labels with no generic-corpus analogue"}},
	{Name: "conditional-regime-analyze", SourceDir: "cmd/conditional-regime-analyze", Quiet: true, Checkpoint: true, Executor: true, CorpusFlag: "-corpus", RequiresMetadata: true, Generic: GenericSupport{Reason: "tests whether distributional/regime structure remains after conditioning on real Currier/hand scribal metadata; the class identity is the experimental variable under test, not a partitioning device, so there is no generic corpus-only analogue of the same hypothesis"}},
	{Name: "residual-diagnostic-analyze", SourceDir: "cmd/residual-diagnostic-analyze", Quiet: true, CorpusFlag: "-corpus", RequiresMetadata: true, Generic: GenericSupport{Reason: "diagnoses metadata association (Currier/hand/joint NMI) in conditional-regime-analyze's frozen residual clusters; without real Currier/hand truth labels there is nothing to diagnose"}},
	{Name: "token-relation-validate", SourceDir: "cmd/token-relation-validate", Quiet: true, Checkpoint: true, Executor: true, CorpusFlag: "-corpus", RequiresMetadata: true, Generic: GenericSupport{Applicable: true}},
	{Name: "replicated-local-structure-audit", SourceDir: "cmd/replicated-local-structure-audit", Quiet: true, Checkpoint: true, Executor: true, CorpusFlag: "-corpus", RequiresMetadata: true, Generic: GenericSupport{Applicable: true}},
	{Name: "higher-order-sequence-validate", SourceDir: "cmd/higher-order-sequence-validate", Quiet: true, Checkpoint: true, Executor: true, CorpusFlag: "-corpus", RequiresMetadata: true, Generic: GenericSupport{Applicable: true}},
	{Name: "positional-continuation-validate", SourceDir: "cmd/positional-continuation-validate", Quiet: true, Checkpoint: true, Executor: true, CorpusFlag: "-corpus", RequiresMetadata: true, Generic: GenericSupport{Applicable: true}},
	{Name: "transition-network-validate", SourceDir: "cmd/transition-network-validate", Quiet: true, Checkpoint: true, Executor: true, CorpusFlag: "-corpus", RequiresMetadata: true, Generic: GenericSupport{Applicable: true}},
	{Name: "vocabulary-growth-analyze", SourceDir: "cmd/vocabulary-growth-analyze", CorpusFlag: "-input"},
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
// distributed-capable stages' jobs across real mTLS-authenticated remote
// workers (Task33/34/40). None of these affect any scientific parameter.
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
	if s.Name == "structural-projection-analyze" {
		args = append(args, "-checkpoint", "workdir/structural-projection-checkpoint.json")
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

// stageArgsForInput preserves the byte-for-byte historical Voynich command
// lines, while making the authoritative corpus explicit in generic mode.
func stageArgsForInput(s Stage, opt orchestratorOptions, inputMode, corpusPath string) []string {
	args := stageArgs(s, opt)
	if inputMode != "generic" {
		return args
	}
	return appendExplicitCorpus(args, s, corpusPath)
}

// genericCapableStages are the task43 Class B/C stages whose CLI accepts
// -generic-corpus in place of -token-metadata-map (see each package's
// main.go). Kept as a set here rather than re-deriving it from
// Stage.Generic.Applicable at every call site, since stageArgsForIsolatedInput
// needs a simple membership check per case below.
var genericCapableStages = map[string]bool{
	"token-relation-validate":          true,
	"replicated-local-structure-audit": true,
	"higher-order-sequence-validate":   true,
	"positional-continuation-validate": true,
	"transition-network-validate":      true,
}

func stageArgsForIsolatedInput(s Stage, opt orchestratorOptions, corpusPath, inputMode string) []string {
	args := appendExplicitCorpus(stageArgs(s, opt), s, corpusPath)
	generic := inputMode == "generic" && genericCapableStages[s.Name]
	if generic {
		args = append(args, "-generic-corpus")
	}
	switch s.Name {
	case "structural-analyze":
		args = append(args, "-dictionary", "workdir/dataset/dictionary.yaml", "-analysis", "workdir/dataset/tokens_analysis.yaml", "-output", "workdir/dataset/structural_analysis.yaml")
	case "sequence-analyze":
		args = append(args, "-output", "workdir/sequence_analysis.yaml")
	case "begin-end-analyze":
		args = append(args, "-dictionary", "workdir/dataset/dictionary.yaml", "-output-dir", "workdir")
	case "structural-normalize":
		args = append(args, "-structural", "workdir/dataset/structural_analysis.yaml", "-output", "workdir/normalized.txt", "-classes", "workdir/structural_classes.yaml")
	case "normalization-compare":
		args = append(args, "-classes", "workdir/structural_classes.yaml", "-raw-analysis", "workdir/sequence_analysis.yaml", "-normalized-pattern", "workdir/normalized_%s.txt", "-analysis-pattern", "workdir/sequence_analysis_%s.yaml", "-sequence-analyzer", "workdir/bin/sequence-analyze", "-output", "workdir/normalization_comparison.yaml")
	case "structural-validate":
		args = append(args, "-classes", "workdir/structural_classes.yaml", "-output", "workdir/structural_validation.yaml")
	case "structural-profile-stability":
		args = append(args, "-classes", "workdir/structural_classes.yaml", "-output", "workdir/structural_profile_stability.yaml")
	case "structural-reliability":
		args = append(args, "-classes", "workdir/structural_classes.yaml", "-output", "workdir/structural_reliability.yaml")
	case "soft-structural-space":
		args = append(args, "-dictionary", "workdir/dataset/dictionary.yaml", "-analysis", "workdir/dataset/tokens_analysis.yaml", "-reliability", "workdir/structural_reliability.yaml", "-output", "workdir/soft_structural_space.yaml", "-pairs-output", "workdir/soft_structural_pairs.tsv")
	case "structural-graphemic":
		args = append(args, "-input", "workdir/soft_structural_pairs.tsv", "-output-dir", "workdir")
	case "structural-pair-decompose":
		args = append(args, "-dictionary", "workdir/dataset/dictionary.yaml", "-pairs", "workdir/structural_graphemic_pairs.tsv", "-distant", "workdir/structural_distant_top.tsv", "-families", "workdir/structural_distant_families.yaml")
	case "distance-context-analyze":
		args = append(args, "-distant-pairs", "workdir/structural_distant_top.tsv", "-families", "workdir/structural_distant_families.yaml", "-controls", "workdir/pair_controls.tsv", "-output-dir", "workdir")
	case "local-regime-analyze":
		args = append(args, "-distance-pairs", "workdir/distance_context_pairs.yaml", "-controls", "workdir/distance_context_controls.tsv", "-output-dir", "workdir")
	case "property-trajectory-analyze":
		args = append(args, "-structural-pairs", "workdir/soft_structural_pairs.tsv", "-distance-pairs", "workdir/distance_context_pairs.yaml", "-controls", "workdir/distance_context_controls.tsv", "-output-dir", "workdir")
	case "structural-projection-analyze":
		args = append(args, "-structural-pairs", "workdir/soft_structural_pairs.tsv", "-distance-pairs", "workdir/distance_context_pairs.yaml", "-families", "workdir/structural_distant_families.yaml", "-output-dir", "workdir")
	case "global-regime-analyze":
		args = append(args, "-output-dir", "workdir")
	case "metadata-validate":
		args = append(args, "-discovery-dir", "workdir", "-output-dir", "workdir/metadata-validation")
	case "cluster-metadata-global":
		args = append(args,
			"-discovery-dir", "workdir",
			"-token-metadata-map", "workdir/metadata-validation/token_metadata_map.tsv",
			"-metadata-report", "workdir/metadata-validation/metadata_validation_report.md",
			"-output-dir", "workdir/metadata-validation")
	case "conditional-regime-analyze":
		args = append(args, "-token-metadata-map", "workdir/metadata-validation/token_metadata_map.tsv", "-output-dir", "workdir/conditional-regimes")
	case "residual-diagnostic-analyze":
		args = append(args, "-conditional-dir", "workdir/conditional-regimes", "-token-metadata-map", "workdir/metadata-validation/token_metadata_map.tsv", "-output-dir", "workdir/residual-diagnostics")
	case "token-relation-validate":
		if !generic {
			args = append(args, "-token-metadata-map", "workdir/metadata-validation/token_metadata_map.tsv")
		}
		args = append(args, "-discovery-dir", "workdir", "-output-dir", "workdir/token-relation-validation")
	case "replicated-local-structure-audit":
		if !generic {
			args = append(args, "-token-metadata-map", "workdir/metadata-validation/token_metadata_map.tsv")
		}
		args = append(args, "-relation-dir", "workdir/token-relation-validation", "-discovery-dir", "workdir", "-output-dir", "workdir/replicated-local-structure")
	case "higher-order-sequence-validate":
		if !generic {
			args = append(args, "-token-metadata-map", "workdir/metadata-validation/token_metadata_map.tsv")
		}
		args = append(args, "-audit-dir", "workdir/replicated-local-structure", "-discovery-dir", "workdir", "-output-dir", "workdir/higher-order-sequences")
	case "positional-continuation-validate":
		if !generic {
			args = append(args, "-token-metadata-map", "workdir/metadata-validation/token_metadata_map.tsv")
		}
		args = append(args, "-higher-order-dir", "workdir/higher-order-sequences", "-output-dir", "workdir/positional-continuation")
	case "transition-network-validate":
		if !generic {
			args = append(args, "-token-metadata-map", "workdir/metadata-validation/token_metadata_map.tsv")
		}
		args = append(args, "-output-dir", "workdir/transition-network")
	case "vocabulary-growth-analyze":
		args = append(args, "-output-dir", "workdir/vocabulary-growth")
	}
	return args
}

func appendExplicitCorpus(args []string, s Stage, corpusPath string) []string {
	if s.Name == "dict-gen" {
		args[0] = corpusPath
	} else if s.CorpusFlag != "" {
		args = append(args, s.CorpusFlag, corpusPath)
	}
	return args
}
