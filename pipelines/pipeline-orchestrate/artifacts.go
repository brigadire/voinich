package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const artifactRegistryVersion = 1

type ArtifactRecord struct {
	Path             string `json:"path"`
	SHA256           string `json:"sha256"`
	Stage            string `json:"stage"`
	ExperimentID     string `json:"experiment_id"`
	CorpusSHA256     string `json:"corpus_sha256"`
	ParametersSHA256 string `json:"parameters_sha256"`
}

type ArtifactRegistry struct {
	Version      int              `json:"version"`
	ExperimentID string           `json:"experiment_id"`
	CorpusSHA256 string           `json:"corpus_sha256"`
	Artifacts    []ArtifactRecord `json:"artifacts"`
}

func workspaceDir(experimentDir string) string { return filepath.Join(experimentDir, "workspace") }
func workspaceWorkdir(experimentDir string) string {
	return filepath.Join(workspaceDir(experimentDir), "workdir")
}
func artifactRegistryPath(experimentDir string) string {
	return filepath.Join(experimentDir, "artifacts.json")
}

func newArtifactRegistry(m *Manifest) *ArtifactRegistry {
	return &ArtifactRegistry{Version: artifactRegistryVersion, ExperimentID: m.ExperimentID, CorpusSHA256: m.CorpusSHA256}
}

func loadArtifactRegistry(experimentDir string) (*ArtifactRegistry, error) {
	b, err := os.ReadFile(artifactRegistryPath(experimentDir))
	if err != nil {
		return nil, err
	}
	var r ArtifactRegistry
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func saveArtifactRegistry(experimentDir string, r *ArtifactRegistry) error {
	sort.Slice(r.Artifacts, func(i, j int) bool { return r.Artifacts[i].Path < r.Artifacts[j].Path })
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	tmp := artifactRegistryPath(experimentDir) + ".tmp"
	if err := os.WriteFile(tmp, b, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, artifactRegistryPath(experimentDir))
}

func invocationHash(m *Manifest, sm StageManifest) string {
	h := sha256.New()
	fmt.Fprintf(h, "experiment=%s\ncorpus=%s\nstage=%s\nargs=%q\n", m.ExperimentID, m.CorpusSHA256, sm.Name, sm.Args)
	return hex.EncodeToString(h.Sum(nil))
}

// scanScientificArtifacts hashes the experiment-local mutable output tree.
// Binaries are operational scratch and deliberately excluded.
func scanScientificArtifacts(experimentDir string) (map[string]string, error) {
	root := workspaceWorkdir(experimentDir)
	result := make(map[string]string)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		rel, err := filepath.Rel(root, path)
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
		// Checkpoints are stage-local operational state, not scientific
		// artifacts.  A stage can leave one behind when the process is
		// interrupted; it must remain available for the stage's own resume
		// logic without making the whole experiment workspace stale.
		name := filepath.Base(rel)
		if name == "checkpoint.json" || strings.HasSuffix(name, "-checkpoint.json") {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		sum, err := sha256File(path)
		if err != nil {
			return err
		}
		result[filepath.ToSlash(rel)] = sum
		return nil
	})
	if os.IsNotExist(err) {
		return result, nil
	}
	return result, err
}

func validateRegistry(experimentDir string, m *Manifest, rs *RunState, r *ArtifactRegistry) error {
	workspaceEntries, err := os.ReadDir(workspaceDir(experimentDir))
	if err != nil {
		return err
	}
	for _, entry := range workspaceEntries {
		if entry.Name() != "workdir" || !entry.IsDir() {
			return fmt.Errorf("unregistered/stale path in experiment workspace: %s", entry.Name())
		}
	}
	if r.Version != artifactRegistryVersion || r.ExperimentID != m.ExperimentID || r.CorpusSHA256 != m.CorpusSHA256 {
		return fmt.Errorf("artifact registry provenance does not match current experiment/corpus")
	}
	actual, err := scanScientificArtifacts(experimentDir)
	if err != nil {
		return err
	}
	registered := make(map[string]ArtifactRecord, len(r.Artifacts))
	stagePlan := make(map[string]StageManifest, len(m.Stages))
	for _, sm := range m.Stages {
		stagePlan[sm.Name] = sm
	}
	for _, a := range r.Artifacts {
		if a.ExperimentID != m.ExperimentID || a.CorpusSHA256 != m.CorpusSHA256 {
			return fmt.Errorf("artifact %s has foreign provenance", a.Path)
		}
		sr := rs.stage(a.Stage)
		if sr == nil || (sr.Status != "completed" && sr.Status != "running" && sr.Status != "failed") {
			return fmt.Errorf("artifact %s is not owned by an executed stage %s", a.Path, a.Stage)
		}
		sm, ok := stagePlan[a.Stage]
		if !ok || a.ParametersSHA256 != invocationHash(m, sm) || sr.InvocationSHA256 != invocationHash(m, sm) {
			return fmt.Errorf("artifact %s invocation provenance mismatch", a.Path)
		}
		if actual[a.Path] != a.SHA256 {
			return fmt.Errorf("artifact %s checksum mismatch or missing", a.Path)
		}
		registered[a.Path] = a
	}
	for path := range actual {
		if _, ok := registered[path]; !ok {
			return fmt.Errorf("unregistered/stale artifact in experiment workspace: %s", path)
		}
	}
	return nil
}

func validateStageDependencies(m *Manifest, sm StageManifest, r *ArtifactRegistry, rs *RunState) error {
	index := make(map[string]int, len(m.Stages))
	for _, stage := range m.Stages {
		index[stage.Name] = stage.Index
	}
	for _, a := range r.Artifacts {
		producerIndex, ok := index[a.Stage]
		if !ok || (producerIndex >= sm.Index && a.Stage != sm.Name) {
			return fmt.Errorf("stage %s cannot consume artifact %s owned by non-dependency stage %s", sm.Name, a.Path, a.Stage)
		}
		producer := rs.stage(a.Stage)
		if a.Stage != sm.Name && (producer == nil || producer.Status != "completed") {
			return fmt.Errorf("stage %s dependency artifact %s was not produced by a completed stage", sm.Name, a.Path)
		}
	}
	return nil
}

func registerStageChanges(r *ArtifactRegistry, before, after map[string]string, m *Manifest, sm StageManifest) []ArtifactRecord {
	byPath := make(map[string]ArtifactRecord, len(r.Artifacts))
	for _, a := range r.Artifacts {
		byPath[a.Path] = a
	}
	var produced []ArtifactRecord
	for path, sum := range after {
		if before[path] == sum {
			continue
		}
		a := ArtifactRecord{Path: path, SHA256: sum, Stage: sm.Name, ExperimentID: m.ExperimentID, CorpusSHA256: m.CorpusSHA256, ParametersSHA256: invocationHash(m, sm)}
		byPath[path] = a
		produced = append(produced, a)
	}
	for path := range before {
		if _, ok := after[path]; !ok {
			delete(byPath, path)
		}
	}
	r.Artifacts = r.Artifacts[:0]
	for _, a := range byPath {
		r.Artifacts = append(r.Artifacts, a)
	}
	sort.Slice(produced, func(i, j int) bool { return produced[i].Path < produced[j].Path })
	return produced
}
