package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestVoynichExecutionPlanRegression(t *testing.T) {
	opt := orchestratorOptions{Executor: "process", LocalWorkers: 8}
	for _, st := range stages {
		got := stageArgsForInput(st, opt, "ivtff", "some-other-corpus")
		want := stageArgs(st, opt)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s: old plan changed: got %v want %v", st.Name, got, want)
		}
	}
	legacy := &Manifest{GitCommit: "c", IVTFFSHA256: "i", CorpusSHA256: "x", GoVersion: "g", GOOS: "o", GOARCH: "a"}
	explicit := *legacy
	explicit.InputMode = "ivtff"
	if computeExperimentID(legacy) != computeExperimentID(&explicit) {
		t.Fatal("legacy manifest and explicit IVTFF manifest have different experiment IDs")
	}
}

func TestGenericManifestDoesNotReadIVTFFAndHasCompleteInventory(t *testing.T) {
	repo, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	m, err := buildManifest(repo, "generic", "definitely/missing/ZL3b-n.txt", "data_test/pg2097-2.txt", orchestratorOptions{Executor: "process", LocalWorkers: 2}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if m.IVTFF != nil || m.IVTFFPath != "" || m.IVTFFSHA256 != "" {
		t.Fatalf("generic manifest contains IVTFF: %+v", m.IVTFF)
	}
	if m.Corpus == nil || m.Corpus.SHA256 == "" {
		t.Fatal("generic corpus hash missing")
	}
	if len(m.Stages) != 27 {
		t.Fatalf("got %d stages", len(m.Stages))
	}

	for i, sm := range m.Stages {
		st := stages[i]
		if st.RequiresMetadata && !st.Generic.Applicable {
			// task43 Class A stages: still NOT_APPLICABLE, but with a
			// specific scientific reason rather than the old blanket phrase.
			if sm.Status != "NOT_APPLICABLE" || sm.Reason == "" || sm.Reason == "requires IVTFF metadata" || len(sm.Args) != 0 {
				t.Fatalf("%s applicability: %+v", sm.Name, sm)
			}
			continue
		}
		if sm.Status != "PLANNED" {
			t.Fatalf("%s status %q", sm.Name, sm.Status)
		}
		joined := strings.Join(sm.Args, " ")
		if strings.Contains(joined, "ZL3b") {
			t.Fatalf("generic stage %s contains Voynich fallback: %v", sm.Name, sm.Args)
		}
		if st.Name == "dict-gen" || st.CorpusFlag != "" {
			if !strings.Contains(joined, "data_test/pg2097-2.txt") {
				t.Fatalf("generic stage %s lacks authoritative corpus: %v", sm.Name, sm.Args)
			}
		}
		// task43 Class B/C stages (23-27): must run generically via
		// -generic-corpus, never against a nonexistent generic
		// token_metadata_map.tsv.
		if st.RequiresMetadata && st.Generic.Applicable {
			if !strings.Contains(joined, "-generic-corpus") {
				t.Fatalf("generic-capable stage %s missing -generic-corpus: %v", sm.Name, sm.Args)
			}
			if strings.Contains(joined, "-token-metadata-map") {
				t.Fatalf("generic-capable stage %s still passes -token-metadata-map: %v", sm.Name, sm.Args)
			}
		}
	}
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "ZL3b-n.txt") || strings.Contains(string(b), "ZL3b-x7.txt") {
		t.Fatalf("generic manifest contaminated: %s", b)
	}
	if !strings.Contains(string(b), `"ivtff":null`) {
		t.Fatalf("generic manifest must explicitly record null IVTFF: %s", b)
	}
}

func TestGenericCLIRequiresExplicitCorpus(t *testing.T) {
	if got := run([]string{"manifest", "-generic-corpus", "-experiment-dir", filepath.Join(t.TempDir(), "experiment")}); got != 2 {
		t.Fatalf("exit code %d, want 2", got)
	}
}

func TestGenericResumeTreatsNotApplicableAsComplete(t *testing.T) {
	m := &Manifest{ExperimentID: "generic", InputMode: "generic"}
	for i, st := range stages {
		sm := StageManifest{Index: i + 1, Name: st.Name, Status: "PLANNED"}
		if st.RequiresMetadata {
			sm.Status, sm.Reason = "NOT_APPLICABLE", "requires IVTFF metadata"
		}
		m.Stages = append(m.Stages, sm)
	}
	rs := newRunStateForManifest(m)
	for _, sr := range rs.Stages {
		if sr.Status == "pending" {
			rs.stage(sr.Name).Status = "completed"
		}
	}
	if !rs.allCompleted() {
		t.Fatal("NOT_APPLICABLE stages prevent completion/resume")
	}
	for _, st := range stages {
		if st.RequiresMetadata && rs.stage(st.Name).Status != "NOT_APPLICABLE" {
			t.Fatalf("%s was not preserved", st.Name)
		}
	}
}

func TestGenericFreezeReportAndVerify(t *testing.T) {
	tmp := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWD) })
	if err := os.Mkdir("workdir", 0755); err != nil {
		t.Fatal(err)
	}
	corpus := filepath.Join(tmp, "corpus.txt")
	if err := os.WriteFile(corpus, []byte("one two\nthree\n"), 0644); err != nil {
		t.Fatal(err)
	}
	experiment := filepath.Join(tmp, "experiment")
	m := &Manifest{ExperimentID: "generic", InputMode: "generic", CorpusPath: corpus, CorpusSHA256: "hash", Corpus: &InputFileManifest{Path: corpus, SHA256: "hash"}, CreatedAt: time.Now()}
	for i, st := range stages {
		status, reason := "PLANNED", ""
		if st.RequiresMetadata {
			status, reason = "NOT_APPLICABLE", "requires IVTFF metadata"
		}
		m.Stages = append(m.Stages, StageManifest{Index: i + 1, Name: st.Name, Status: status, Reason: reason})
	}
	if err := saveManifest(experiment, m); err != nil {
		t.Fatal(err)
	}
	rs := newRunStateForManifest(m)
	for i := range rs.Stages {
		if rs.Stages[i].Status == "pending" {
			rs.Stages[i].Status = "completed"
		}
	}
	// Exclude any repository workdir files from this isolated freeze.
	rs.StartedAt = time.Now().Add(time.Hour)
	if err := saveRunState(experiment, rs); err != nil {
		t.Fatal(err)
	}
	if err := freezeExperiment(experiment, false); err != nil {
		t.Fatal(err)
	}
	if err := verifyExperiment(experiment); err != nil {
		t.Fatal(err)
	}
	report, err := os.ReadFile(filepath.Join(experiment, "REPORT.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(report)
	for _, want := range []string{"Input mode: generic corpus", "Corpus SHA256: `hash`", "Token count: 3", "Not applicable stages", "requires IVTFF metadata"} {
		if !strings.Contains(text, want) {
			t.Errorf("report lacks %q", want)
		}
	}
	if strings.Contains(text, "IVTFF source:") || strings.Contains(text, "Voynich Baseline") {
		t.Fatalf("generic report mentions Voynich input/baseline:\n%s", text)
	}
}

func TestIsolatedWorkspaceRejectsStaleAndFreezeUsesRegistryOnly(t *testing.T) {
	tmp := t.TempDir()
	experiment := filepath.Join(tmp, "experiment-b")
	corpus := filepath.Join(tmp, "corpus-b.txt")
	if err := os.WriteFile(corpus, []byte("beta corpus\n"), 0644); err != nil {
		t.Fatal(err)
	}
	sum, err := sha256File(corpus)
	if err != nil {
		t.Fatal(err)
	}
	m := &Manifest{InputMode: "generic", CorpusPath: corpus, CorpusSHA256: sum, IsolationVersion: 1, Workspace: "workspace", CreatedAt: time.Now()}
	m.Stages = []StageManifest{{Index: 1, Name: stages[0].Name, Status: "PLANNED", Args: []string{corpus}, WorkingDirectory: "workspace"}}
	m.ExperimentID = computeExperimentID(m)
	if err := saveManifest(experiment, m); err != nil {
		t.Fatal(err)
	}
	rs := newRunStateForManifest(m)
	if err := os.MkdirAll(workspaceWorkdir(experiment), 0755); err != nil {
		t.Fatal(err)
	}
	r := newArtifactRegistry(m)
	if err := saveArtifactRegistry(experiment, r); err != nil {
		t.Fatal(err)
	}

	// A foreign experiment may physically coexist elsewhere; it is never
	// scanned because B's workspace is its complete mutable namespace.
	foreign := filepath.Join(tmp, "experiment-a", "workspace", "workdir")
	if err := os.MkdirAll(foreign, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(foreign, "SENTINEL_READ_IS_FAILURE"), []byte("corpus-a"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := validateRegistry(experiment, m, rs, r); err != nil {
		t.Fatalf("foreign workspace contaminated B: %v", err)
	}

	// The same sentinel inside B is a hard error unless a completed stage
	// registered it with matching experiment/corpus provenance.
	stale := filepath.Join(workspaceWorkdir(experiment), "stale-a.yaml")
	if err := os.WriteFile(stale, []byte("corpus-a"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := validateRegistry(experiment, m, rs, r); err == nil || !strings.Contains(err.Error(), "unregistered/stale") {
		t.Fatalf("stale artifact was accepted: %v", err)
	}
	if err := os.Remove(stale); err != nil {
		t.Fatal(err)
	}

	valid := filepath.Join(workspaceWorkdir(experiment), "result.yaml")
	if err := os.WriteFile(valid, []byte("corpus: beta\n"), 0644); err != nil {
		t.Fatal(err)
	}
	after, err := scanScientificArtifacts(experiment)
	if err != nil {
		t.Fatal(err)
	}
	produced := registerStageChanges(r, map[string]string{}, after, m, m.Stages[0])
	if len(produced) != 1 {
		t.Fatalf("produced %d artifacts", len(produced))
	}
	rs.Stages[0].Status = "completed"
	rs.Stages[0].InvocationSHA256 = invocationHash(m, m.Stages[0])
	if err := saveRunState(experiment, rs); err != nil {
		t.Fatal(err)
	}
	if err := saveArtifactRegistry(experiment, r); err != nil {
		t.Fatal(err)
	}
	if err := freezeExperiment(experiment, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(experiment, "outputs", "result.yaml")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(experiment, "outputs", "SENTINEL_READ_IS_FAILURE")); !os.IsNotExist(err) {
		t.Fatal("foreign sentinel entered freeze")
	}
	if err := verifyExperiment(experiment); err != nil {
		t.Fatal(err)
	}
}

func TestResumeRejectsChangedCorpus(t *testing.T) {
	tmp := t.TempDir()
	corpus := filepath.Join(tmp, "corpus.txt")
	if err := os.WriteFile(corpus, []byte("corpus A\n"), 0644); err != nil {
		t.Fatal(err)
	}
	sum, _ := sha256File(corpus)
	experiment := filepath.Join(tmp, "experiment")
	m := &Manifest{InputMode: "generic", CorpusPath: corpus, CorpusSHA256: sum, IsolationVersion: 1, Workspace: "workspace"}
	m.ExperimentID = computeExperimentID(m)
	if err := saveManifest(experiment, m); err != nil {
		t.Fatal(err)
	}
	if err := saveRunState(experiment, newRunStateForManifest(m)); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(workspaceWorkdir(experiment), 0755); err != nil {
		t.Fatal(err)
	}
	if err := saveArtifactRegistry(experiment, newArtifactRegistry(m)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(corpus, []byte("corpus B\n"), 0644); err != nil {
		t.Fatal(err)
	}
	err := runPipeline("..", experiment, orchestratorOptions{}, "")
	if err == nil || !strings.Contains(err.Error(), "corpus identity changed") {
		t.Fatalf("changed corpus resume result: %v", err)
	}
}
