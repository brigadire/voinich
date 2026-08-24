package task82a

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"zcore.dev/voinich/internal/fingerprintv2"
	"zcore.dev/voinich/internal/mnemonicspace"
)

// task81Bindings are the same three Task81 V1.1 authoritative-input
// checksums internal/task82.bindings already verifies; Task82a re-verifies
// them independently rather than trusting Task82's own check.
var task81Bindings = map[string]string{
	"MNEMONIC_MECHANISM_REGISTRY.json": "2b1d53038a790863d356892a9a147acf8e551fd6e721632111c11276a1d9016e",
	"MNEMONIC_PARAMETER_REGISTRY.tsv":  "fb88dcf9a5ca9c1dc52dd8c3fd45be2d0cc2e3da3245127fc58d53947dc3fcdb",
	"MNEMONIC_RECOVERY_CONTRACT.md":    "d0e6fe3756dd5e65c8c108db794fbb6404a9a3aee739d4fe76fdf77b2a333fcb",
}

var resultsShaRe = regexp.MustCompile(`results_manifest_sha256=([0-9a-f]{64})`)
var f2ManifestShaRe = regexp.MustCompile(`freeze_manifest_sha256:\s*([0-9a-f]{64})`)

// Freeze holds everything runManifest/Aggregate need after verifying every
// upstream freeze binding (task82a.txt sec.1, "AUTHORITATIVE INPUTS").
type Freeze struct {
	Authority      mnemonicspace.Authority
	Registry       map[string]mnemonicspace.MechanismSpec
	Task82Commit   string
	F2FreezeLine   string
	ManifestOnDisk Manifest
}

func verifyFreeze(root string) (Freeze, error) {
	mechDir := filepath.Join(root, "research", "phase2", "mechanism-space")
	for name, want := range task81Bindings {
		got, err := fileHashPath(filepath.Join(mechDir, name))
		if err != nil {
			return Freeze{}, err
		}
		if got != want {
			return Freeze{}, fmt.Errorf("Task81 binding %s checksum %s, want %s", name, got, want)
		}
	}
	authority, err := mnemonicspace.LoadTask80Authority(filepath.Join(root, "research", "phase2", "fontana", "task80"))
	if err != nil {
		return Freeze{}, err
	}
	reg := mnemonicspace.FrozenRegistry()
	if err := mnemonicspace.ValidateRegistry(authority, reg); err != nil {
		return Freeze{}, err
	}
	specs := map[string]mnemonicspace.MechanismSpec{}
	for _, s := range reg {
		specs[s.ID] = s
	}

	task82Frozen, err := os.ReadFile(filepath.Join(root, "research", "phase2", "task82", "TASK82_BLIND_RESULTS_FROZEN"))
	if err != nil {
		return Freeze{}, fmt.Errorf("Task82 results are not frozen: %w", err)
	}
	m := resultsShaRe.FindStringSubmatch(string(task82Frozen))
	if m == nil {
		return Freeze{}, fmt.Errorf("TASK82_BLIND_RESULTS_FROZEN missing results_manifest_sha256")
	}
	gotTask82, err := fileHashPath(filepath.Join(root, "research", "phase2", "task82", "TASK82_BLIND_RESULTS_MANIFEST.json"))
	if err != nil {
		return Freeze{}, err
	}
	if gotTask82 != m[1] {
		return Freeze{}, fmt.Errorf("TASK82_BLIND_RESULTS_MANIFEST.json checksum mismatch: results are not the frozen ones")
	}

	f2Frozen, err := os.ReadFile(filepath.Join(root, "research", "phase2", "fingerprint", "FINGERPRINT_V2_FROZEN"))
	if err != nil {
		return Freeze{}, fmt.Errorf("Fingerprint V2 is not frozen: %w", err)
	}
	fm := f2ManifestShaRe.FindStringSubmatch(string(f2Frozen))
	if fm == nil {
		return Freeze{}, fmt.Errorf("FINGERPRINT_V2_FROZEN missing freeze_manifest_sha256")
	}
	gotF2, err := fileHashPath(filepath.Join(root, "research", "phase2", "fingerprint", "FINGERPRINT_V2_FREEZE_MANIFEST.json"))
	if err != nil {
		return Freeze{}, err
	}
	if gotF2 != fm[1] {
		return Freeze{}, fmt.Errorf("FINGERPRINT_V2_FREEZE_MANIFEST.json checksum mismatch: F2 is not the frozen extractor")
	}

	var onDisk Manifest
	manifestPath := filepath.Join(outDir(root), "TASK82A_BLIND_MANIFEST.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return Freeze{}, fmt.Errorf("TASK82A_BLIND_MANIFEST.json missing; run -gen-manifest first: %w", err)
	}
	if err := json.Unmarshal(data, &onDisk); err != nil {
		return Freeze{}, err
	}
	fresh := BuildManifest()
	if len(fresh.Jobs) != len(onDisk.Jobs) {
		return Freeze{}, fmt.Errorf("TASK82A_BLIND_MANIFEST.json job count %d does not match a fresh regeneration (%d): manifest was edited after freeze", len(onDisk.Jobs), len(fresh.Jobs))
	}
	sort.Slice(fresh.Jobs, func(i, j int) bool { return fresh.Jobs[i].JobID < fresh.Jobs[j].JobID })
	sort.Slice(onDisk.Jobs, func(i, j int) bool { return onDisk.Jobs[i].JobID < onDisk.Jobs[j].JobID })
	for i := range fresh.Jobs {
		if fresh.Jobs[i] != onDisk.Jobs[i] {
			return Freeze{}, fmt.Errorf("TASK82A_BLIND_MANIFEST.json job %d does not match a fresh regeneration: manifest was edited after freeze", i)
		}
	}
	return Freeze{Authority: authority, Registry: specs, F2FreezeLine: strings.TrimSpace(string(f2Frozen)), ManifestOnDisk: onDisk}, nil
}

// Execute runs the frozen Task82a manifest: for each job it assembles the
// corpus-scale OBSERVABLE_DOCUMENT, samples recovery, extracts frozen
// Fingerprint V2 features, and writes the raw job artifact. When run
// unsharded it also regenerates the aggregate portfolio from the raw
// artifacts on disk.
func Execute(o Options) error {
	if o.Root == "" {
		o.Root = "."
	}
	if o.ShardCount <= 0 {
		o.ShardCount = 1
	}
	if o.ShardIndex < 0 || o.ShardIndex >= o.ShardCount {
		return fmt.Errorf("invalid shard %d/%d", o.ShardIndex, o.ShardCount)
	}
	fz, err := verifyFreeze(o.Root)
	if err != nil {
		return fmt.Errorf("FREEZE_MISMATCH: %w", err)
	}
	manifest := fz.ManifestOnDisk
	rawDir := filepath.Join(outDir(o.Root), "raw")
	f2Dir := filepath.Join(outDir(o.Root), "raw", "f2corpus")
	if err := os.MkdirAll(rawDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(f2Dir, 0o755); err != nil {
		return err
	}
	if o.VerifyOnly {
		return verifyRawArtifacts(manifest, rawDir, o.ShardIndex, o.ShardCount)
	}

	letters, items := requiredLengths()
	corpora, err := loadSourceCorpora(o.Root, letters, items)
	if err != nil {
		return err
	}

	for n, job := range manifest.Jobs {
		if n%o.ShardCount != o.ShardIndex {
			continue
		}
		path := filepath.Join(rawDir, job.JobID+".json")
		if o.Resume {
			if data, e := os.ReadFile(path); e == nil {
				var a Artifact
				if json.Unmarshal(data, &a) == nil && a.Job.JobID == job.JobID && a.FreezeVersion == FreezeVersion {
					continue
				}
			}
		}
		spec, ok := fz.Registry[job.MechanismID]
		if !ok {
			return fmt.Errorf("manifest mechanism %s is not frozen", job.MechanismID)
		}
		param, ok := spec.ParameterSet(job.ParameterSetID)
		if !ok {
			return fmt.Errorf("manifest parameter %s is not frozen for %s", job.ParameterSetID, spec.ID)
		}
		scale, ok := scaleFor(job.MechanismID)
		if !ok {
			return fmt.Errorf("no frozen capacity table entry for %s", job.MechanismID)
		}
		corpus, ok := corpora[job.InputCorpusID]
		if !ok {
			return fmt.Errorf("manifest corpus %s is unavailable", job.InputCorpusID)
		}
		a, err := runOneJob(spec, param, scale, job, corpus, f2Dir)
		if err != nil {
			return fmt.Errorf("job %s: %w", job.JobID, err)
		}
		data, err := json.MarshalIndent(a, "", "  ")
		if err != nil {
			return err
		}
		data = append(data, '\n')
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return err
		}
	}
	if o.ShardCount == 1 {
		return Aggregate(o.Root, manifest)
	}
	return nil
}

func runOneJob(spec mnemonicspace.MechanismSpec, param mnemonicspace.ParameterSet, scale MechanismScale, job ManifestJob, corpus SourceCorpus, f2Dir string) (Artifact, error) {
	assembled, err := assembleJob(spec, param, scale, job.ScalingPolicyID, corpus, job.Chunks, job.Seed)
	if err != nil {
		return Artifact{}, err
	}
	f2Path := filepath.Join(f2Dir, job.JobID+".txt")
	if err := writeAssembledCorpusFile(f2Path, assembled.AssembledLines); err != nil {
		return Artifact{}, err
	}
	f2CorpusChecksum, err := fileHashPath(f2Path)
	if err != nil {
		return Artifact{}, err
	}
	f2Metrics, f2Warnings, err := extractF2(f2Path, job.JobID, int64(job.Seed), filepath.Join(f2Dir, "out", job.JobID))
	if err != nil {
		return Artifact{}, fmt.Errorf("F2 extraction: %w", err)
	}
	localCollisions := detectLocalCollisions(assembled.Chunks)
	var flatTokens []string
	for _, line := range assembled.AssembledLines {
		flatTokens = append(flatTokens, line...)
	}
	metrics := computeCorpusMetrics(assembled.Document.Symbols, flatTokens)
	warnings := append([]string(nil), assembled.Warnings...)
	warnings = append(warnings, f2Warnings...)
	a := Artifact{
		Schema: "TASK82A_RAW_JOB_V1", Implementation: Version, FreezeVersion: FreezeVersion,
		Job: job, Family: string(spec.Family), HistoricalStatus: string(spec.Status),
		Corpus:             CorpusInfo{ID: corpus.ID, Path: corpus.Path, SHA256: corpus.SHA256},
		LocalRunCount:      job.Chunks,
		StatePolicy:        "RESET_EACH_CHUNK",
		ScalingPolicyID:    job.ScalingPolicyID,
		BoundaryProvenance: BoundaryProvenance{Token: "ASSEMBLER_DEFINED", Line: "ASSEMBLER_DEFINED", Page: "NOT_DEFINED"},
		Chunks:             assembled.Chunks,
		Document:           assembled.Document,
		LocalRecoveries:    assembled.Recoveries,
		LocalCollisions:    localCollisions,
		F2:                 F2Result{CorpusFileChecksum: f2CorpusChecksum, CorpusFilePath: relPath(f2Path), Metrics: f2Metrics},
		Metrics:            metrics,
		DocumentSHA256:     assembled.Document.Checksum(),
		Warnings:           warnings,
		SoftwareVersion:    "task82a-analyze/" + Version,
	}
	return a, nil
}

func relPath(p string) string {
	i := strings.Index(p, "research/phase2/task82a/")
	if i < 0 {
		return p
	}
	return p[i:]
}

func detectLocalCollisions(chunks []ChunkSummary) []Collision {
	groups := map[string]*Collision{}
	for _, c := range chunks {
		g, ok := groups[c.Checksum]
		if !ok {
			g = &Collision{Checksum: c.Checksum}
			groups[c.Checksum] = g
		}
		g.ChunkIndices = append(g.ChunkIndices, c.Index)
		g.IntendedIDs = append(g.IntendedIDs, c.IntendedID)
	}
	var out []Collision
	for _, g := range groups {
		if len(g.ChunkIndices) > 1 {
			out = append(out, *g)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Checksum < out[j].Checksum })
	return out
}

func verifyRawArtifacts(m Manifest, raw string, si, sc int) error {
	for n, j := range m.Jobs {
		if n%sc != si {
			continue
		}
		b, err := os.ReadFile(filepath.Join(raw, j.JobID+".json"))
		if err != nil {
			return err
		}
		var a Artifact
		if err = json.Unmarshal(b, &a); err != nil {
			return err
		}
		if a.Job.JobID != j.JobID || a.DocumentSHA256 != a.Document.Checksum() {
			return fmt.Errorf("artifact mismatch %s", j.JobID)
		}
	}
	return nil
}

func fileHashPath(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return sum(b), nil
}

// assertF2NoConfigLeak is a build-time reachability check used by
// TestF2FeatureExtractionOnlyGuard: it must be possible to construct a
// fingerprintv2.Config from Task82a without ever setting IVTFFPath.
var _ = fingerprintv2.CorpusConfig{}
