package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// StageRun is one stage's execution record - the unit run.yml phase 13's
// resume logic checks before deciding to skip or (re)run a stage.
type StageRun struct {
	Name             string    `json:"name"`
	Status           string    `json:"status"` // pending|running|completed|failed|NOT_APPLICABLE
	Reason           string    `json:"reason,omitempty"`
	StartedAt        time.Time `json:"started_at,omitempty"`
	FinishedAt       time.Time `json:"finished_at,omitempty"`
	DurationSeconds  float64   `json:"duration_seconds,omitempty"`
	ExitCode         int       `json:"exit_code"`
	UserCPUSeconds   float64   `json:"user_cpu_seconds"`
	SysCPUSeconds    float64   `json:"sys_cpu_seconds"`
	MaxRSSKB         int64     `json:"max_rss_kb"`
	LogPath          string    `json:"log_path"`
	Error            string    `json:"error,omitempty"`
	InvocationSHA256 string    `json:"invocation_sha256,omitempty"`
	Outputs          []string  `json:"outputs,omitempty"`
}

// RunState is Task36's checkpoint/resume record for the orchestrator
// itself: a crash of the orchestrating process (not just a single stage's
// own internal checkpoint) resumes from exactly the stage it stopped at,
// never from the beginning.
type RunState struct {
	ExperimentID string     `json:"experiment_id"`
	StartedAt    time.Time  `json:"started_at"`
	FinishedAt   time.Time  `json:"finished_at,omitempty"`
	Stages       []StageRun `json:"stages"`
}

func runStatePath(experimentDir string) string { return filepath.Join(experimentDir, "run-state.json") }

func newRunStateForManifest(m *Manifest) *RunState {
	experimentID := m.ExperimentID
	rs := &RunState{ExperimentID: experimentID, StartedAt: time.Now().UTC()}
	for _, s := range m.Stages {
		status := "pending"
		if s.Status == "NOT_APPLICABLE" {
			status = "NOT_APPLICABLE"
		}
		rs.Stages = append(rs.Stages, StageRun{Name: s.Name, Status: status, Reason: s.Reason})
	}
	return rs
}

// newRunState retains compatibility for old callers/tests.
func newRunState(experimentID string) *RunState {
	m := &Manifest{ExperimentID: experimentID}
	for i, s := range stages {
		m.Stages = append(m.Stages, StageManifest{Index: i + 1, Name: s.Name})
	}
	return newRunStateForManifest(m)
}

func loadRunState(experimentDir string) (*RunState, error) {
	b, err := os.ReadFile(runStatePath(experimentDir))
	if err != nil {
		return nil, err
	}
	var rs RunState
	if err := json.Unmarshal(b, &rs); err != nil {
		return nil, err
	}
	return &rs, nil
}

// saveRunState is called after every stage completes or fails, atomically,
// so a crash mid-run never leaves a corrupt or ambiguous status record.
func saveRunState(experimentDir string, rs *RunState) error {
	b, err := json.MarshalIndent(rs, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	tmp := runStatePath(experimentDir) + ".tmp"
	if err := os.WriteFile(tmp, b, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, runStatePath(experimentDir))
}

func (rs *RunState) stage(name string) *StageRun {
	for i := range rs.Stages {
		if rs.Stages[i].Name == name {
			return &rs.Stages[i]
		}
	}
	return nil
}

func (rs *RunState) allCompleted() bool {
	for _, s := range rs.Stages {
		if s.Status != "completed" && s.Status != "NOT_APPLICABLE" {
			return false
		}
	}
	return true
}
