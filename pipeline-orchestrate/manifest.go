package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// StageManifest freezes the exact operational command line one stage is
// invoked with. It intentionally does not re-list every scientific default
// value textually: those already live in, and are frozen by, the pinned
// GitCommit above (checking out that commit reproduces the identical
// defaults byte-for-byte). Reproducing a run therefore means "checkout
// GitCommit, run these Args" - not "read numeric values out of this JSON".
type StageManifest struct {
	Index              int      `json:"index"`
	Name               string   `json:"name"`
	Dir                string   `json:"source_dir"`
	Args               []string `json:"args"`
	Status             string   `json:"status,omitempty"`
	Reason             string   `json:"reason,omitempty"`
	WorkingDirectory   string   `json:"working_directory,omitempty"`
	ArtifactInputRoot  string   `json:"artifact_input_root,omitempty"`
	ArtifactOutputRoot string   `json:"artifact_output_root,omitempty"`
}

type InputFileManifest struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

// WorkerManifest documents one execution resource actually used by the
// distributed-capable stages (conditional regime and structural projection).
// Task36 requires "a list of workers" in the
// manifest; this is that list, honestly describing whatever was actually
// used for that run - local process-pool slots, or specific named remote
// hosts - never a fabricated fleet.
type WorkerManifest struct {
	Kind string `json:"kind"` // "local" or "remote"
	Name string `json:"name"` // hostname/identity
}

// Manifest is the Task36 immutable experiment manifest: everything needed
// to know, byte-for-byte, what a run means and to attempt reproducing it.
// It is written once, before the run starts, and never edited afterward -
// experiments/<name>/FROZEN (see freeze.go) is what makes that immutability
// enforceable rather than just a convention.
type Manifest struct {
	ExperimentID     string             `json:"experiment_id"`
	CreatedAt        time.Time          `json:"created_at"`
	GitCommit        string             `json:"git_commit"`
	GitDirty         bool               `json:"git_dirty"`
	IVTFFPath        string             `json:"ivtff_path,omitempty"`
	IVTFFSHA256      string             `json:"ivtff_sha256,omitempty"`
	CorpusPath       string             `json:"corpus_path"`
	CorpusSHA256     string             `json:"corpus_sha256"`
	InputMode        string             `json:"input_mode,omitempty"`
	Corpus           *InputFileManifest `json:"corpus,omitempty"`
	IVTFF            *InputFileManifest `json:"ivtff"`
	GoVersion        string             `json:"go_version"`
	GOOS             string             `json:"goos"`
	GOARCH           string             `json:"goarch"`
	Hostname         string             `json:"hostname"`
	NumCPU           int                `json:"num_cpu"`
	Executor         string             `json:"executor"`      // executor for every distributed-capable stage
	ExecutorNote     string             `json:"executor_note"` // honest description of what "distributed" meant for this run
	Workers          []WorkerManifest   `json:"workers"`
	Stages           []StageManifest    `json:"stages"`
	IsolationVersion int                `json:"isolation_version,omitempty"`
	Workspace        string             `json:"workspace,omitempty"`
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func gitCommit(repoPath string) (commit string, dirty bool, err error) {
	out, err := exec.Command("git", "-C", repoPath, "rev-parse", "HEAD").Output()
	if err != nil {
		return "", false, fmt.Errorf("git rev-parse HEAD: %w", err)
	}
	commit = strings.TrimSpace(string(out))
	// --untracked-files=no deliberately excludes untracked files: this
	// experiment directory itself (created moments before this check) and
	// scratch/profiling output are untracked by design and never affect
	// what the pinned commit's compiled binaries do. Only modifications to
	// already-tracked files threaten reproducibility.
	statusOut, err := exec.Command("git", "-C", repoPath, "status", "--porcelain", "--untracked-files=no").Output()
	if err != nil {
		return commit, false, fmt.Errorf("git status --porcelain: %w", err)
	}
	dirty = len(strings.TrimSpace(string(statusOut))) > 0
	return commit, dirty, nil
}

// buildManifest assembles the immutable manifest for one production run.
// executor/workerConcurrency describe every distributed-capable stage;
// remoteWorkers
// is the honest list of remote worker identities actually configured for
// this run's -executor remote coordinator (empty for local execution). opt
// is passed straight through to stageArgs for every stage, so the manifest
// records the exact command line each stage will actually be run with.
func buildManifest(repoPath, inputMode, ivtffPath, corpusPath string, opt orchestratorOptions, remoteWorkers []string) (*Manifest, error) {
	absCorpusPath, err := filepath.Abs(filepath.Join(repoPath, corpusPath))
	if filepath.IsAbs(corpusPath) {
		absCorpusPath = filepath.Clean(corpusPath)
	}
	if err != nil {
		return nil, fmt.Errorf("resolve corpus path: %w", err)
	}
	commit, dirty, err := gitCommit(repoPath)
	if err != nil {
		return nil, err
	}
	var ivtffHash string
	absIVTFFPath := ""
	if inputMode != "generic" {
		absIVTFFPath, err = filepath.Abs(filepath.Join(repoPath, ivtffPath))
		if filepath.IsAbs(ivtffPath) {
			absIVTFFPath = filepath.Clean(ivtffPath)
		}
		if err != nil {
			return nil, fmt.Errorf("resolve IVTFF source: %w", err)
		}
		ivtffHash, err = sha256File(absIVTFFPath)
		if err != nil {
			return nil, fmt.Errorf("hash IVTFF source: %w", err)
		}
	}
	corpusHash, err := sha256File(absCorpusPath)
	if err != nil {
		return nil, fmt.Errorf("hash frozen corpus: %w", err)
	}
	hostname, _ := os.Hostname()

	var workers []WorkerManifest
	executorNote := ""
	switch opt.Executor {
	case "remote":
		for _, w := range remoteWorkers {
			workers = append(workers, WorkerManifest{Kind: "remote", Name: w})
		}
		executorNote = fmt.Sprintf("distributed-capable stages used -executor remote with %d authenticated remote worker(s) over the shared mTLS service", len(remoteWorkers))
	case "process", "goroutine":
		workers = append(workers, WorkerManifest{Kind: "local", Name: hostname})
		executorNote = fmt.Sprintf("distributed-capable stages used -executor %s with %d local worker slot(s) on %s; no remote machines were part of this run", opt.Executor, opt.LocalWorkers, hostname)
	default:
		workers = append(workers, WorkerManifest{Kind: "local", Name: hostname})
		executorNote = "distributed-capable stages used the default in-process goroutine executor"
	}
	if inputMode == "generic" {
		executorNote = "structural-projection-analyze uses the recorded executor configuration; conditional-regime-analyze is NOT_APPLICABLE because generic corpora lack IVTFF metadata"
	}

	m := &Manifest{
		CreatedAt:        time.Now().UTC(),
		GitCommit:        commit,
		GitDirty:         dirty,
		CorpusPath:       absCorpusPath,
		CorpusSHA256:     corpusHash,
		InputMode:        inputMode,
		Corpus:           &InputFileManifest{Path: absCorpusPath, SHA256: corpusHash},
		GoVersion:        runtime.Version(),
		GOOS:             runtime.GOOS,
		GOARCH:           runtime.GOARCH,
		Hostname:         hostname,
		NumCPU:           runtime.NumCPU(),
		Executor:         opt.Executor,
		ExecutorNote:     executorNote,
		Workers:          workers,
		IsolationVersion: 1,
		Workspace:        "workspace",
	}
	if inputMode != "generic" {
		m.IVTFFPath, m.IVTFFSHA256 = absIVTFFPath, ivtffHash
		m.IVTFF = &InputFileManifest{Path: absIVTFFPath, SHA256: ivtffHash}
	}
	for i, st := range stages {
		args := stageArgsForIsolatedInput(st, opt, absCorpusPath)
		if st.Name == "metadata-validate" && inputMode != "generic" {
			args = append(args, "-ivtff", absIVTFFPath)
		}
		sm := StageManifest{Index: i + 1, Name: st.Name, Dir: st.SourceDir, Args: args, Status: "PLANNED", WorkingDirectory: "workspace", ArtifactInputRoot: "workspace/workdir", ArtifactOutputRoot: "workspace/workdir"}
		if inputMode == "generic" && st.RequiresMetadata {
			sm.Status, sm.Reason = "NOT_APPLICABLE", "requires IVTFF metadata"
		}
		m.Stages = append(m.Stages, sm)
	}
	m.ExperimentID = computeExperimentID(m)
	return m, nil
}

// computeExperimentID hashes everything that determines reproducibility:
// commit, both frozen input hashes, every stage's exact argument list, and
// the toolchain/platform. Two runs with the same ExperimentID were given
// byte-identical instructions.
func computeExperimentID(m *Manifest) string {
	h := sha256.New()
	if m.effectiveInputMode() == "ivtff" && m.IsolationVersion == 0 {
		// Preserve Task36's ID calculation byte-for-byte for old invocations.
		fmt.Fprintf(h, "commit=%s\nivtff=%s\ncorpus=%s\ngo=%s\ngoos=%s\ngoarch=%s\n", m.GitCommit, m.IVTFFSHA256, m.CorpusSHA256, m.GoVersion, m.GOOS, m.GOARCH)
	} else {
		fmt.Fprintf(h, "commit=%s\nmode=%s\nivtff=%s\ncorpus=%s\ngo=%s\ngoos=%s\ngoarch=%s\nisolation=%d\nworkspace=%s\n", m.GitCommit, m.effectiveInputMode(), m.IVTFFSHA256, m.CorpusSHA256, m.GoVersion, m.GOOS, m.GOARCH, m.IsolationVersion, m.Workspace)
	}
	for _, st := range m.Stages {
		fmt.Fprintf(h, "stage=%s cwd=%s inputs=%s outputs=%s args=%v\n", st.Name, st.WorkingDirectory, st.ArtifactInputRoot, st.ArtifactOutputRoot, st.Args)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (m *Manifest) effectiveInputMode() string {
	if m.InputMode == "" {
		return "ivtff"
	}
	return m.InputMode
}

func manifestPath(experimentDir string) string { return filepath.Join(experimentDir, "manifest.json") }

func saveManifest(experimentDir string, m *Manifest) error {
	if err := os.MkdirAll(experimentDir, 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	tmp := manifestPath(experimentDir) + ".tmp"
	if err := os.WriteFile(tmp, b, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, manifestPath(experimentDir))
}

func loadManifest(experimentDir string) (*Manifest, error) {
	b, err := os.ReadFile(manifestPath(experimentDir))
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
