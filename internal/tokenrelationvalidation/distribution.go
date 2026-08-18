package tokenrelationvalidation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
)

// LoadForDistribution reconstructs exactly the Blocks/Candidates/MaxD a
// local analyze() run would compute from the same corpus/metadata/
// discovery-dir inputs, calling the very same loadCorpus +
// loadMetadata/loadGenericMetadata + loadCandidates path analyze() itself
// uses. It is exported so internal/conditionalregime's distributed worker
// (Task44) can reconstruct identical state from staged copies of those
// same input files without duplicating this package's loading logic.
func LoadForDistribution(c Config) (blocks []Block, candidates []Candidate, maxD int, err error) {
	corpus, _, err := loadCorpus(c.CorpusPath)
	if err != nil {
		return nil, nil, 0, err
	}
	if c.Generic {
		_, blocks, _, _, err = loadGenericMetadata(c.CorpusPath, corpus)
	} else {
		_, blocks, _, _, err = loadMetadata(c.MetadataPath, corpus)
	}
	if err != nil {
		return nil, nil, 0, err
	}
	candidates, _, maxD, err = loadCandidates(c.DiscoveryDir)
	if err != nil {
		return nil, nil, 0, err
	}
	return blocks, candidates, maxD, nil
}

// discoveryFiles are the exact named files loadCandidates reads from a
// discovery directory - the required ones first, then the optional ones
// it only includes when present. Exported so a distributed coordinator
// (internal/conditionalregime) can stage precisely these by content hash,
// mirroring how structural-projection-analyze/normalization-compare stage
// their own small, fixed input sets.
var (
	RequiredDiscoveryFiles = []string{
		"begin_end_candidates.yaml",
		"distance_context_pairs.yaml",
		"sequence_analysis.yaml",
		"structural_reliability.yaml",
		"structural_classes.yaml",
		"soft_structural_space.yaml",
		"soft_structural_pairs.tsv",
	}
	OptionalDiscoveryFiles = []string{
		"begin_end_top.tsv",
		"distance_context_top.tsv",
		"structural_validation.yaml",
		"structural_profile_stability.yaml",
	}
)

// Fingerprint hashes every byte-identity-relevant input this stage reads
// (corpus, metadata map or generic-mode marker, every discovery file that
// exists) plus every scientific parameter, mirroring
// structuralprojection.Fingerprint/normalizationcompare.Fingerprint. A
// distributed worker must reproduce this exact value from its own staged
// copies before it is trusted to compute a single replicate.
func Fingerprint(c Config) (string, error) {
	h := sha256.New()
	paths := []string{c.CorpusPath}
	if !c.Generic {
		paths = append(paths, c.MetadataPath)
	}
	for _, name := range RequiredDiscoveryFiles {
		paths = append(paths, filepath.Join(c.DiscoveryDir, name))
	}
	for _, name := range OptionalDiscoveryFiles {
		p := filepath.Join(c.DiscoveryDir, name)
		if _, err := os.Stat(p); err == nil {
			paths = append(paths, p)
		}
	}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			return "", err
		}
		sum := sha256.Sum256(b)
		h.Write(sum[:])
	}
	v := struct {
		Generic                          bool
		Permutations, RefinePermutations int
		Seed                             int64
	}{c.Generic, c.Permutations, c.RefinePermutations, c.Seed}
	b, _ := json.Marshal(v)
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil)), nil
}
