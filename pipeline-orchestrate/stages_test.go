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
	repoRoot := "../"
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

	entries, err := os.ReadDir(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	knownDirs := map[string]bool{".": true}
	for _, s := range stages {
		knownDirs[s.SourceDir] = true
	}
	// Tools that are not pipeline stages: administrative/orchestration
	// commands this very task introduces or that Task34 introduced.
	exempt := map[string]bool{
		"conditional-regime-pki":       true,
		"codex_prepare":                true,
		"codex_orientation":            true,
		"corpus-transform":             true,
		"inverse-homophony":            true,
		"inverse-transposition-search": true,
		"voynich-validation":           true,
		"pipeline-orchestrate":         true,
		"experiment-compare":           true,
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(repoRoot, e.Name(), "main.go")); err != nil {
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
