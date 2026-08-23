package main

import (
	"os"
	"path/filepath"
	"testing"
)

// chdirRepoRoot walks up from the package directory to the module root
// (marked by go.mod), matching the cwd every independent/*-analyze binary
// assumes when it reads repo-relative paths like data_test/... or
// experiments/....
func chdirRepoRoot(t *testing.T) {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root (go.mod)")
		}
		dir = parent
	}
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(old) })
}

// test: authoritative target values are actually read from the frozen
// Task58-65 artifacts (not left as MISSING_ARTIFACT), when run from the
// repository root - and no metric is silently defaulted to zero if its
// artifact is absent (task66 section 10).
func TestLoadVoynichTargetsReadsAuthoritativeArtifacts(t *testing.T) {
	chdirRepoRoot(t)
	targets, err := LoadVoynichTargets()
	if err != nil {
		t.Fatalf("LoadVoynichTargets: %v", err)
	}
	if len(targets) == 0 {
		t.Fatalf("expected at least one target")
	}
	families := map[string]bool{}
	haveValue := false
	for _, tg := range targets {
		families[tg.Family] = true
		if tg.Status == "VALUE" {
			haveValue = true
		}
		if tg.Status != "VALUE" && tg.Voynich != 0 {
			t.Fatalf("metric %s/%s has status %s but a nonzero value %v (should be blank, not a silent default)", tg.Family, tg.Metric, tg.Status, tg.Voynich)
		}
	}
	if !haveValue {
		t.Fatalf("expected at least one VALUE-status target when run from the repo root")
	}
	for _, want := range []string{"TOKEN_ORDER", "POSITIONAL_STRUCTURE", "REPETITION_EDIT_GEOMETRY", "CHARACTER_ENTROPY", "TOKEN_FORMATION", "LOCAL_TRANSITION", "LOCAL_REGIME_TOPOLOGY"} {
		if !families[want] {
			t.Fatalf("missing family %s from target manifest", want)
		}
	}
}

// test: the development/held-out metric split is disjoint and every
// family has at least one metric on each side (task66 section 5).
func TestMetricRegistryDevelopmentHeldoutSplitIsDisjoint(t *testing.T) {
	byFamily := map[string]map[string]bool{}
	for _, m := range MetricRegistry {
		if byFamily[m.Family] == nil {
			byFamily[m.Family] = map[string]bool{}
		}
		byFamily[m.Family][m.Stage] = true
	}
	for fam, stages := range byFamily {
		if !stages[StageDevelopment] {
			t.Fatalf("family %s has no DEVELOPMENT metric", fam)
		}
		if !stages[StageHeldout] {
			t.Fatalf("family %s has no HELDOUT metric", fam)
		}
	}
}
