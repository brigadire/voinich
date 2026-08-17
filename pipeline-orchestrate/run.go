package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
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
	if m.IsolationVersion > 0 && computeExperimentID(m) != m.ExperimentID {
		return fmt.Errorf("manifest ExperimentID does not match its isolated execution plan")
	}
	currentCorpusHash, err := sha256File(m.CorpusPath)
	if err != nil || currentCorpusHash != m.CorpusSHA256 {
		return fmt.Errorf("corpus identity changed since manifest creation (want %s, got %s): %v", m.CorpusSHA256, currentCorpusHash, err)
	}
	if m.IsolationVersion == 0 {
		return runPipelineLegacy(repoPath, experimentDir, opt, only, m, rs)
	}
	r, err := loadArtifactRegistry(experimentDir)
	if err != nil {
		return fmt.Errorf("load artifacts.json: %w", err)
	}
	if err := validateRegistry(experimentDir, m, rs, r); err != nil {
		return err
	}

	logDir := filepath.Join(experimentDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}
	binDir := filepath.Join(workspaceWorkdir(experimentDir), "bin")
	runDir := workspaceDir(experimentDir)

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
		if sr.Status == "NOT_APPLICABLE" || (sr.Status == "completed" && only == "") {
			if sr.Status == "completed" && sr.InvocationSHA256 != invocationHash(m, sm) {
				return fmt.Errorf("stage %s resume provenance mismatch", st.Name)
			}
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
		if err := validateRegistry(experimentDir, m, rs, r); err != nil {
			return err
		}
		if err := validateStageDependencies(m, sm, r, rs); err != nil {
			return err
		}
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
		before, err := scanScientificArtifacts(experimentDir)
		if err != nil {
			return err
		}
		var inputs []string
		for _, a := range r.Artifacts {
			inputs = append(inputs, a.Path)
		}
		header := orchestrationHeader(m, sm, runDir, inputs)
		result := runLogged(runDir, absBinPath, args, logPath, header)
		after, scanErr := scanScientificArtifacts(experimentDir)
		if scanErr != nil {
			return scanErr
		}
		produced := registerStageChanges(r, before, after, m, sm)
		if len(produced) > 0 {
			sr.Outputs = sr.Outputs[:0]
			for _, a := range produced {
				sr.Outputs = append(sr.Outputs, a.Path)
			}
		}
		sr.InvocationSHA256 = invocationHash(m, sm)
		if err := saveArtifactRegistry(experimentDir, r); err != nil {
			return err
		}

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

func orchestrationHeader(m *Manifest, sm StageManifest, dir string, inputs []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "ExperimentID: %s\nInput mode: %s\nCorpus: %s\nCorpus SHA256: %s\nStage: %s\nWorking directory: %s\nInputs:\n", m.ExperimentID, m.effectiveInputMode(), m.CorpusPath, m.CorpusSHA256, sm.Name, dir)
	if len(inputs) == 0 {
		b.WriteString("  (corpus only)\n")
	}
	for _, in := range inputs {
		fmt.Fprintf(&b, "  workdir/%s\n", in)
	}
	b.WriteString("Outputs:\n  registered in artifacts.json after completion\n\n")
	return b.String()
}

// Unsafe legacy shared-workdir manifests remain readable/verifiable but may
// not be resumed. New manifests never enter this path.
func runPipelineLegacy(repoPath, experimentDir string, opt orchestratorOptions, only string, m *Manifest, rs *RunState) error {
	return fmt.Errorf("legacy non-isolated experiments are read-only; create a new experiment manifest")
}
