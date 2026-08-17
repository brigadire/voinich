package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"zcore.dev/voinich/internal/workdir"
)

func frozenMarkerPath(experimentDir string) string { return filepath.Join(experimentDir, "FROZEN") }

// isFrozen reports whether experimentDir already has a FROZEN marker -
// Task36's "no subsequent change may silently overwrite this baseline".
func isFrozen(experimentDir string) (bool, error) {
	_, err := os.Stat(frozenMarkerPath(experimentDir))
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, nil
}

// snapshotOutputs copies every file under workdir/ modified at or after
// runStartedAt (excluding workdir/bin, the build scratch area - never a
// scientific output) into experimentDir/outputs, preserving relative
// paths, and returns the (sorted) list of relative paths copied. The
// mtime cutoff is deliberate: workdir/ is a long-lived, shared scratch
// area other tasks/experiments have also written to, and a frozen
// baseline must contain only what *this* run actually produced - never
// unrelated stale content left over from earlier work.
func snapshotOutputs(experimentDir string, runStartedAt time.Time) ([]string, error) {
	src := workdir.Dir
	dst := filepath.Join(experimentDir, "outputs")
	if err := os.MkdirAll(dst, 0755); err != nil {
		return nil, err
	}
	var relPaths []string
	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if rel == "bin" || strings.HasPrefix(rel, "bin"+string(filepath.Separator)) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil // created lazily below, only for files actually copied
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.ModTime().Before(runStartedAt) {
			return nil // predates this run: stale content from earlier work, never this baseline's
		}
		dstPath := filepath.Join(dst, rel)
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}
		if err := copyFile(path, dstPath); err != nil {
			return fmt.Errorf("copy %s: %w", rel, err)
		}
		relPaths = append(relPaths, rel)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(relPaths)
	return relPaths, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// writeChecksums computes SHA256 for every file under outputs/ (by its
// relative path) in the standard `sha256sum`-compatible format, so
// `sha256sum -c checksums.sha256` (run from experimentDir/outputs)
// verifies the frozen baseline at any point in the future.
func writeChecksums(experimentDir string, relPaths []string) (string, error) {
	outputsDir := filepath.Join(experimentDir, "outputs")
	path := filepath.Join(experimentDir, "checksums.sha256")
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	for _, rel := range relPaths {
		sum, err := sha256File(filepath.Join(outputsDir, rel))
		if err != nil {
			return "", err
		}
		line := fmt.Sprintf("%s  %s\n", sum, rel)
		if _, err := f.WriteString(line); err != nil {
			return "", err
		}
		io.WriteString(h, line)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// writeReport writes the Task36 final report: per-stage wall-time/status/
// resource usage, total wall-time, and pointers to the manifest and
// checksums.
func writeReport(experimentDir string, m *Manifest, rs *RunState) error {
	var b strings.Builder
	title := "Voynich Baseline"
	if m.effectiveInputMode() == "generic" {
		title = "Generic Corpus Pipeline"
	}
	fmt.Fprintf(&b, "# %s - Experiment Report\n\n", title)
	fmt.Fprintf(&b, "ExperimentID: `%s`\n\n", m.ExperimentID)
	fmt.Fprintf(&b, "Git commit: `%s`%s\n\n", m.GitCommit, dirtyNote(m.GitDirty))
	fmt.Fprintf(&b, "Created: %s\n\n", m.CreatedAt.Format(time.RFC3339))
	fmt.Fprintf(&b, "Platform: %s/%s, Go %s, host `%s`, %d CPUs\n\n", m.GOOS, m.GOARCH, m.GoVersion, m.Hostname, m.NumCPU)
	if m.effectiveInputMode() == "generic" {
		fmt.Fprintf(&b, "Input mode: generic corpus\n\nCorpus: `%s`\n\nCorpus SHA256: `%s`\n\n", m.CorpusPath, m.CorpusSHA256)
		n, err := tokenCount(m.CorpusPath)
		if err != nil {
			return fmt.Errorf("count generic corpus tokens: %w", err)
		}
		fmt.Fprintf(&b, "Token count: %d\n\n", n)
	} else {
		fmt.Fprintf(&b, "Input mode: IVTFF\n\nIVTFF source: `%s` (sha256 `%s`)\n\n", m.IVTFFPath, m.IVTFFSHA256)
		fmt.Fprintf(&b, "Frozen corpus: `%s` (sha256 `%s`)\n\n", m.CorpusPath, m.CorpusSHA256)
	}
	fmt.Fprintf(&b, "Executor: `%s` - %s\n\n", m.Executor, m.ExecutorNote)
	fmt.Fprintf(&b, "Workers:\n\n")
	for _, w := range m.Workers {
		fmt.Fprintf(&b, "- %s: `%s`\n", w.Kind, w.Name)
	}
	fmt.Fprintf(&b, "\n## Stage results\n\n")
	fmt.Fprintf(&b, "| # | Stage | Status | Wall time | User CPU | Sys CPU | Max RSS |\n")
	fmt.Fprintf(&b, "|---|---|---|---|---|---|---|\n")
	var total time.Duration
	for i, sr := range rs.Stages {
		total += time.Duration(sr.DurationSeconds * float64(time.Second))
		fmt.Fprintf(&b, "| %d | %s | %s | %.1fs | %.1fs | %.1fs | %d KB |\n",
			i+1, sr.Name, sr.Status, sr.DurationSeconds, sr.UserCPUSeconds, sr.SysCPUSeconds, sr.MaxRSSKB)
	}
	if m.effectiveInputMode() == "generic" {
		fmt.Fprintf(&b, "\n## Not applicable stages\n\n")
		for _, sr := range rs.Stages {
			if sr.Status == "NOT_APPLICABLE" {
				fmt.Fprintf(&b, "- %s — reason: %s\n", sr.Name, sr.Reason)
			}
		}
	}
	fmt.Fprintf(&b, "\n**Total wall time (sum of stages): %s**\n\n", total.Round(time.Second))
	fmt.Fprintf(&b, "Full manifest: `manifest.json`. Per-file checksums: `checksums.sha256`. Per-stage logs: `logs/`.\n")
	return os.WriteFile(filepath.Join(experimentDir, "REPORT.md"), []byte(b.String()), 0644)
}

func tokenCount(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	s.Split(bufio.ScanWords)
	n := 0
	for s.Scan() {
		n++
	}
	return n, s.Err()
}

func dirtyNote(dirty bool) string {
	if dirty {
		return " (**dirty working tree at manifest time**)"
	}
	return ""
}

// freezeExperiment finalizes a fully-completed run: snapshot workdir/'s
// outputs, checksum them, write the report, and write the FROZEN marker
// that refuses any further run/freeze against this directory.
func freezeExperiment(experimentDir string, force bool) error {
	frozen, err := isFrozen(experimentDir)
	if err != nil {
		return err
	}
	if frozen && !force {
		return fmt.Errorf("experiment %s is already FROZEN; pass -force to refreeze (this OVERWRITES the existing baseline - only do this deliberately)", experimentDir)
	}
	m, err := loadManifest(experimentDir)
	if err != nil {
		return fmt.Errorf("load manifest: %w", err)
	}
	rs, err := loadRunState(experimentDir)
	if err != nil {
		return fmt.Errorf("load run-state: %w", err)
	}
	if !rs.allCompleted() {
		return fmt.Errorf("not all stages completed; refusing to freeze an incomplete run")
	}
	if m.IsolationVersion > 0 {
		return freezeIsolatedExperiment(experimentDir, force, m, rs)
	}

	relPaths, err := snapshotOutputs(experimentDir, rs.StartedAt)
	if err != nil {
		return fmt.Errorf("snapshot outputs: %w", err)
	}
	checksumsDigest, err := writeChecksums(experimentDir, relPaths)
	if err != nil {
		return fmt.Errorf("write checksums: %w", err)
	}
	if err := writeReport(experimentDir, m, rs); err != nil {
		return fmt.Errorf("write report: %w", err)
	}

	label := "Voynich Baseline"
	if m.effectiveInputMode() == "generic" {
		label = "Generic Corpus Pipeline"
	}
	marker := fmt.Sprintf("%s frozen at %s\nExperimentID: %s\nchecksums.sha256 sha256: %s\nFiles: %d\n", label,
		time.Now().UTC().Format(time.RFC3339), m.ExperimentID, checksumsDigest, len(relPaths))
	if err := os.WriteFile(frozenMarkerPath(experimentDir), []byte(marker), 0444); err != nil {
		return fmt.Errorf("write FROZEN marker: %w", err)
	}
	fmt.Printf("Frozen %s: %d output files, checksums.sha256 sha256=%s\n", experimentDir, len(relPaths), checksumsDigest)
	return nil
}

// verifyExperiment recomputes every checksum in checksums.sha256 and
// reports any mismatch or missing file - the drift-detection mechanism
// backing "no change may silently overwrite the baseline".
func verifyExperiment(experimentDir string) error {
	m, err := loadManifest(experimentDir)
	if err != nil {
		return fmt.Errorf("load manifest: %w", err)
	}
	if m.IsolationVersion > 0 {
		if err := verifyIsolatedProvenance(experimentDir, m); err != nil {
			return err
		}
	}
	outputsDir := filepath.Join(experimentDir, "outputs")
	b, err := os.ReadFile(filepath.Join(experimentDir, "checksums.sha256"))
	if err != nil {
		return err
	}
	lines := strings.Split(strings.TrimRight(string(b), "\n"), "\n")
	mismatches := 0
	for _, line := range lines {
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) != 2 {
			continue
		}
		want, rel := parts[0], parts[1]
		got, err := sha256File(filepath.Join(outputsDir, rel))
		if err != nil {
			fmt.Printf("MISSING  %s: %v\n", rel, err)
			mismatches++
			continue
		}
		if got != want {
			fmt.Printf("MISMATCH %s: want %s got %s\n", rel, want, got)
			mismatches++
		}
	}
	if mismatches > 0 {
		return fmt.Errorf("%d file(s) failed verification", mismatches)
	}
	fmt.Printf("Verified %d files: all match checksums.sha256\n", len(lines))
	return nil
}

func freezeIsolatedExperiment(experimentDir string, force bool, m *Manifest, rs *RunState) error {
	r, err := loadArtifactRegistry(experimentDir)
	if err != nil {
		return fmt.Errorf("load artifacts.json: %w", err)
	}
	if err := validateRegistry(experimentDir, m, rs, r); err != nil {
		return fmt.Errorf("artifact provenance: %w", err)
	}
	for _, a := range r.Artifacts {
		if sr := rs.stage(a.Stage); sr == nil || sr.Status != "completed" {
			return fmt.Errorf("artifact %s producer %s is not completed", a.Path, a.Stage)
		}
		for _, sm := range m.Stages {
			if sm.Name == a.Stage && sm.Status == "NOT_APPLICABLE" {
				return fmt.Errorf("NOT_APPLICABLE stage %s owns artifact %s", a.Stage, a.Path)
			}
		}
	}
	outputsDir := filepath.Join(experimentDir, "outputs")
	if entries, err := os.ReadDir(outputsDir); err == nil && len(entries) > 0 {
		return fmt.Errorf("outputs directory is not empty; refusing to mix snapshots")
	}
	if err := os.MkdirAll(outputsDir, 0755); err != nil {
		return err
	}
	var relPaths []string
	for _, a := range r.Artifacts {
		if err := copyFile(filepath.Join(workspaceWorkdir(experimentDir), filepath.FromSlash(a.Path)), filepath.Join(outputsDir, filepath.FromSlash(a.Path))); err != nil {
			return fmt.Errorf("copy registered artifact %s: %w", a.Path, err)
		}
		relPaths = append(relPaths, a.Path)
	}
	sort.Strings(relPaths)
	checksumsDigest, err := writeChecksums(experimentDir, relPaths)
	if err != nil {
		return err
	}
	if err := writeReport(experimentDir, m, rs); err != nil {
		return err
	}
	marker := fmt.Sprintf("Isolated experiment frozen at %s\nExperimentID: %s\nCorpus SHA256: %s\nchecksums.sha256 sha256: %s\nFiles: %d\n", time.Now().UTC().Format(time.RFC3339), m.ExperimentID, m.CorpusSHA256, checksumsDigest, len(relPaths))
	if err := os.WriteFile(frozenMarkerPath(experimentDir), []byte(marker), 0444); err != nil {
		return err
	}
	fmt.Printf("Frozen %s: %d registered output files\n", experimentDir, len(relPaths))
	return nil
}

func verifyIsolatedProvenance(experimentDir string, m *Manifest) error {
	if computeExperimentID(m) != m.ExperimentID {
		return fmt.Errorf("manifest ExperimentID does not match its isolated execution plan")
	}
	rs, err := loadRunState(experimentDir)
	if err != nil {
		return fmt.Errorf("load run-state: %w", err)
	}
	r, err := loadArtifactRegistry(experimentDir)
	if err != nil {
		return fmt.Errorf("load artifacts.json: %w", err)
	}
	if rs.ExperimentID != m.ExperimentID || r.ExperimentID != m.ExperimentID || r.CorpusSHA256 != m.CorpusSHA256 {
		return fmt.Errorf("manifest/run-state/artifact registry provenance mismatch")
	}
	if !rs.allCompleted() {
		return fmt.Errorf("frozen experiment has incomplete stages")
	}
	stagePlan := make(map[string]StageManifest, len(m.Stages))
	for _, sm := range m.Stages {
		stagePlan[sm.Name] = sm
		sr := rs.stage(sm.Name)
		if sr == nil {
			return fmt.Errorf("run-state missing stage %s", sm.Name)
		}
		if sm.Status == "NOT_APPLICABLE" && sr.Status != "NOT_APPLICABLE" {
			return fmt.Errorf("stage %s applicability mismatch", sm.Name)
		}
		if sm.Status != "NOT_APPLICABLE" && sr.Status != "completed" {
			return fmt.Errorf("stage %s is not completed", sm.Name)
		}
	}
	seen := make(map[string]bool)
	for _, a := range r.Artifacts {
		if seen[a.Path] {
			return fmt.Errorf("duplicate artifact registry path %s", a.Path)
		}
		seen[a.Path] = true
		if a.ExperimentID != m.ExperimentID || a.CorpusSHA256 != m.CorpusSHA256 {
			return fmt.Errorf("foreign provenance for artifact %s", a.Path)
		}
		sr := rs.stage(a.Stage)
		if sr == nil || sr.Status != "completed" {
			return fmt.Errorf("invalid producing stage for artifact %s", a.Path)
		}
		sm, ok := stagePlan[a.Stage]
		if !ok || a.ParametersSHA256 != invocationHash(m, sm) || sr.InvocationSHA256 != invocationHash(m, sm) {
			return fmt.Errorf("stage invocation provenance mismatch for artifact %s", a.Path)
		}
		got, err := sha256File(filepath.Join(experimentDir, "outputs", filepath.FromSlash(a.Path)))
		if err != nil || got != a.SHA256 {
			return fmt.Errorf("artifact registry checksum mismatch for %s", a.Path)
		}
	}
	outputsRoot := filepath.Join(experimentDir, "outputs")
	if err := filepath.WalkDir(outputsRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(outputsRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if !seen[rel] {
			return fmt.Errorf("unregistered file in frozen outputs: %s", rel)
		}
		return nil
	}); err != nil {
		return err
	}
	for _, sm := range m.Stages {
		if sm.Status != "NOT_APPLICABLE" {
			continue
		}
		for _, a := range r.Artifacts {
			if a.Stage == sm.Name {
				return fmt.Errorf("NOT_APPLICABLE stage %s has artifact %s", sm.Name, a.Path)
			}
		}
	}
	return nil
}
