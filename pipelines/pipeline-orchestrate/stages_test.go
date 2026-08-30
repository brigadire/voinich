package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestStagesMatchRepository is a guard against the exact class of bug this
// table is most at risk of: a stage renamed/removed/added in the repo
// without this orchestrator being updated to match. It does not run any
// stage - only checks that every SourceDir really exists and really has a
// main.go, and that every top-level pipeline command directory in the repo
// (per internal/workdir's own contract test enumeration) is accounted for
// exactly once.
func TestStagesMatchRepository(t *testing.T) {
	// Task70 moved this orchestrator into pipelines/pipeline-orchestrate/
	// (one level deeper than its pre-Task70 root location) and every stage's
	// SourceDir now points into cmd/, so the repository-root offset and the
	// "every command directory is accounted for" scan both target cmd/.
	repoRoot := "../../"
	cmdDir := filepath.Join(repoRoot, "cmd")
	seen := map[string]bool{}
	for _, s := range stages {
		if seen[s.Name] {
			t.Errorf("duplicate stage name %q", s.Name)
		}
		seen[s.Name] = true
		mainGo := filepath.Join(repoRoot, s.SourceDir, "main.go")
		if _, err := os.Stat(mainGo); err != nil {
			t.Errorf("stage %q: %s: %v", s.Name, mainGo, err)
		}
	}

	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		t.Fatal(err)
	}
	knownDirs := map[string]bool{}
	for _, s := range stages {
		knownDirs[filepath.Base(s.SourceDir)] = true
	}
	// Tools under cmd/ that are not pipeline stages: administrative/
	// orchestration commands Task34/45/46/48 introduced. research/phase1/*
	// and pipelines/pipeline-orchestrate itself live outside cmd/ entirely
	// after Task70 and are out of scope for this scan, same as before the
	// move. fingerprint-v2-analyze is a separately configured Phase II
	// research entry point: it has no frozen Task49 stage contract and
	// requires an explicit YAML output_dir.
	exempt := map[string]bool{
		"conditional-regime-pki": true,
		"codex_prepare":          true,
		"codex_orientation":      true,
		"experiment-compare":     true,
		"fingerprint-v2-analyze": true,
		"fingerprint-v2-verify":  true,
		// Task86C-v2 distributed manifest coordinator/worker and evidence
		// verifier. It executes a separate frozen DAG, not a Task49 stage.
		"g1v2-executor": true,
		// Task79c closure tools: separately configured Phase II research
		// entry points with explicit, caller-chosen output paths, same as
		// fingerprint-v2-analyze; none has a frozen Task49 stage contract.
		"tei-abbr-extract":        true,
		"generic-glyph-filter":    true,
		"ivtff-x7-extract":        true,
		"task79c-pf4-hr":          true,
		"task79c-distance-pareto": true,
		// Task82b closure tools: same class again, independent research
		// entry points with an explicit -out flag and no frozen Task49
		// stage contract.
		"task82b-run":       true,
		"task82b-aggregate": true,
		// Exploratory positional-numeral experiment; outputs are frozen under
		// research/numeric and are not part of the Task49 production DAG.
		"numeric-test": true,
		// Descriptive structural catalog and frozen query tool; outputs are
		// frozen under research/structure_catalog, outside the Task49 DAG.
		"vm-structure": true,
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(cmdDir, e.Name(), "main.go")); err != nil {
			continue // not a command directory
		}
		if !knownDirs[e.Name()] && !exempt[e.Name()] {
			t.Errorf("repository has command directory %q with no corresponding stage in stages.go (or exempt list)", e.Name())
		}
	}
}

func TestStageArgsNeverSetsAScientificFlag(t *testing.T) {
	// A defensive allowlist: every flag this orchestrator is permitted to
	// pass. Anything else appearing in stageArgs' output would mean a
	// scientific parameter leaked in - exactly what a frozen baseline must
	// never allow.
	allowed := map[string]bool{
		"-quiet": true, "-checkpoint": true, "-executor": true, "-workers": true,
		"-remote-listen": true, "-tls-cert": true, "-tls-key": true, "-client-ca": true,
		"-remote-deny-list": true, "-remote-timeout": true, "-remote-retries": true,
	}
	opt := orchestratorOptions{Executor: "remote", LocalWorkers: 8, RemoteListen: "x", TLSCert: "x", TLSKey: "x", ClientCA: "x", RemoteDenyList: "x", RemoteTimeout: "1m", RemoteRetries: 2}
	for _, s := range stages {
		args := stageArgs(s, opt)
		for _, a := range args {
			if len(a) > 0 && a[0] == '-' && !allowed[a] {
				t.Errorf("stage %q: unexpected flag %q in generated args %v", s.Name, a, args)
			}
		}
	}
}
