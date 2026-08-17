package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"zcore.dev/voinich/internal/workdir"
)

// runPipeline builds every stage's binary (once) and executes stages in
// order, skipping any already marked "completed" in run-state.json (Task36
// "checkpoint/resume": an orchestrator crash resumes from where it
// stopped, never from scratch). It stops at the first failing stage, since
// every later stage's default input path assumes the earlier one
// succeeded - Task36 explicitly forbids changing parameters mid-run, and
// silently skipping ahead past a failure would be worse: it would produce
// results computed against stale/partial upstream state.
func runPipeline(repoPath, experimentDir string, opt orchestratorOptions, only string) error {
	m, err := loadManifest(experimentDir)
	if err != nil {
		return fmt.Errorf("load manifest (run `manifest` first): %w", err)
	}
	if frozen, _ := isFrozen(experimentDir); frozen {
		return fmt.Errorf("experiment %s is FROZEN; refusing to run again (see FROZEN marker)", experimentDir)
	}

	rs, err := loadRunState(experimentDir)
	if err != nil {
		rs = newRunStateForManifest(m)
	}
	if rs.ExperimentID != m.ExperimentID {
		return fmt.Errorf("run-state.json belongs to experiment %s, manifest is %s - refusing to mix runs", rs.ExperimentID, m.ExperimentID)
	}

	logDir := filepath.Join(experimentDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}
	binDir := workdir.Path("bin")

	fmt.Printf("Task36 pipeline run: experiment %s\n", m.ExperimentID)
	for i, sm := range m.Stages {
		st, ok := stageByName(sm.Name)
		if !ok {
			return fmt.Errorf("manifest contains unknown stage %q", sm.Name)
		}
		if only != "" && st.Name != only {
			continue
		}
		sr := rs.stage(st.Name)
		if sr == nil {
			rs.Stages = append(rs.Stages, StageRun{Name: st.Name, Status: "pending"})
			sr = rs.stage(st.Name)
		}
		if sr.Status == "completed" || sr.Status == "NOT_APPLICABLE" {
			fmt.Printf("[%d/%d] %-35s SKIP (%s", i+1, len(m.Stages), st.Name, sr.Status)
			if sr.Reason != "" {
				fmt.Printf(": %s", sr.Reason)
			}
			fmt.Println(")")
			continue
		}

		binPath := filepath.Join(binDir, st.Name)
		fmt.Printf("[%d/%d] %-35s building...\n", i+1, len(m.Stages), st.Name)
		if err := buildBinary(repoPath, st.SourceDir, binPath); err != nil {
			sr.Status, sr.Error = "failed", err.Error()
			_ = saveRunState(experimentDir, rs)
			return fmt.Errorf("stage %s: %w", st.Name, err)
		}

		args := append([]string(nil), sm.Args...)
		logPath := filepath.Join(logDir, fmt.Sprintf("%02d-%s.log", i+1, st.Name))
		sr.Status = "running"
		sr.StartedAt = time.Now().UTC()
		sr.LogPath = logPath
		_ = saveRunState(experimentDir, rs)

		fmt.Printf("[%d/%d] %-35s running (log: %s)\n", i+1, len(m.Stages), st.Name, logPath)
		absBinPath, err := filepath.Abs(binPath)
		if err != nil {
			return err
		}
		result := runLogged(repoPath, absBinPath, args, logPath)

		sr.FinishedAt = time.Now().UTC()
		sr.DurationSeconds = result.Duration.Seconds()
		sr.ExitCode = result.ExitCode
		sr.UserCPUSeconds = result.UserCPUSeconds
		sr.SysCPUSeconds = result.SysCPUSeconds
		sr.MaxRSSKB = result.MaxRSSKB

		if result.Err != nil {
			sr.Status, sr.Error = "failed", result.Err.Error()
			_ = saveRunState(experimentDir, rs)
			return fmt.Errorf("stage %s failed after %s (see %s): %w", st.Name, result.Duration.Round(time.Second), logPath, result.Err)
		}
		sr.Status = "completed"
		_ = saveRunState(experimentDir, rs)
		fmt.Printf("[%d/%d] %-35s completed in %s\n", i+1, len(m.Stages), st.Name, result.Duration.Round(time.Second))
	}

	if rs.allCompleted() {
		rs.FinishedAt = time.Now().UTC()
		_ = saveRunState(experimentDir, rs)
		fmt.Println("All stages completed.")
	}
	return nil
}
