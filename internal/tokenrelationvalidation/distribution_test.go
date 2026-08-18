package tokenrelationvalidation

import (
	"os"
	"path/filepath"
	"testing"
)

// writeDiscoveryFixture writes the minimal valid set of discovery-dir
// files loadCandidates/Fingerprint need, so LoadForDistribution/Fingerprint
// can be exercised without a full pipeline run.
func writeDiscoveryFixture(t *testing.T, dir string) {
	t.Helper()
	files := map[string]string{
		"begin_end_candidates.yaml":   "meta:\n  token_occurrences: 100\nparameters:\n  max_window: 3\ncandidates: []\n",
		"distance_context_pairs.yaml": "token_count: 100\nparameters:\n  max_distance: 5\npairs: []\n",
		"sequence_analysis.yaml":      "meta:\n  token_occurrences: 100\nrepeated_ngrams: {}\n",
		"structural_reliability.yaml": "meta:\n  token_occurrences: 100\nparameters:\n  threshold: 0.7\nreference_pairs: []\n",
		"structural_classes.yaml":     "models: []\n",
		"soft_structural_space.yaml":  "parameters:\n  graph_min_similarity: 0.5\n  min_token_count: 3\n",
		"soft_structural_pairs.tsv":   "token_a\ttoken_b\traw_similarity\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func writeSmallCorpus(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "corpus.txt")
	if err := os.WriteFile(path, []byte("aiin chey aiin chey\nol or ol or\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestFingerprintAndLoadForDistributionRoundTripGeneric(t *testing.T) {
	dir := t.TempDir()
	corpusPath := writeSmallCorpus(t, dir)
	writeDiscoveryFixture(t, dir)
	c := Config{CorpusPath: corpusPath, DiscoveryDir: dir, Generic: true, Permutations: 10, RefinePermutations: 20, Seed: 1}

	blocks, candidates, maxD, err := LoadForDistribution(c)
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least one physical block")
	}
	if maxD <= 0 {
		t.Fatal("expected a positive maxD from distance_context_pairs.yaml")
	}
	_ = candidates // empty fixture has no candidates; loading without error is what's being proven here

	fp1, err := Fingerprint(c)
	if err != nil {
		t.Fatal(err)
	}
	fp2, err := Fingerprint(c)
	if err != nil {
		t.Fatal(err)
	}
	if fp1 != fp2 {
		t.Fatal("Fingerprint is not deterministic across identical calls")
	}

	// Changing a scientific parameter must change the fingerprint.
	c2 := c
	c2.Seed = 2
	fp3, err := Fingerprint(c2)
	if err != nil {
		t.Fatal(err)
	}
	if fp3 == fp1 {
		t.Fatal("Fingerprint did not change when Seed changed")
	}

	// Changing a discovery file's content must change the fingerprint.
	if err := os.WriteFile(filepath.Join(dir, "structural_classes.yaml"), []byte("models: []\n# touched\n"), 0644); err != nil {
		t.Fatal(err)
	}
	fp4, err := Fingerprint(c)
	if err != nil {
		t.Fatal(err)
	}
	if fp4 == fp1 {
		t.Fatal("Fingerprint did not change when a discovery file's bytes changed")
	}
}
