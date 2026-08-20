package experimentcompare

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func addTask49(t *testing.T, dir string) {
	t.Helper()
	root := filepath.Join(dir, "outputs", "vocabulary-growth")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "vocabulary_growth.tsv"), []byte("checkpoint_n\tvocabulary_size\ttype_token_ratio\thapax_count\tdis_count\ttri_count\tbeta_effective\n100\t80\t0.8\t70\t5\t2\t0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "summary.yaml"), []byte("final_profile:\n  total_tokens: 100\n  unique_tokens: 80\n  heaps_K: 2\n  heaps_beta: 0.7\n  heaps_R2: 0.9\n  hapax: 70\n  hapax_fraction_of_types: 0.875\n  hapax_fraction_of_tokens: 0.7\n  dis_legomena: 5\n  singleton_to_doubleton_ratio: 14\n"), 0644); err != nil {
		t.Fatal(err)
	}
}

func fixture(t *testing.T, root, id string, tokens, eligible int) string {
	t.Helper()
	dir := filepath.Join(root, id)
	out := filepath.Join(dir, "outputs")
	if err := os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	manifest := fmt.Sprintf(`{"experiment_id":%q,"git_commit":"fixture","input_mode":"generic","corpus_path":"corpus.txt","corpus_sha256":"fixture","stages":[{"name":"x","status":"completed"}]}`, id)
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(manifest), 0644); err != nil {
		t.Fatal(err)
	}
	transition := fmt.Sprintf("token_count: %d\nunique_tokens: 10\neligible_tokens: %d\nfdr_significant_preferred: 4\nfdr_significant_depleted: 2\nbackbone_preferred: 2\nbackbone_depleted: 1\nreplicated_outgoing_profiles: 1\nreplicated_incoming_profiles: 1\n", tokens, eligible)
	name := "transition-network/transition_network_summary.yaml"
	p := filepath.Join(out, name)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(transition), 0644); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(b)
	if err := os.WriteFile(filepath.Join(dir, "checksums.sha256"), []byte(fmt.Sprintf("%x  %s\n", sum, name)), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "FROZEN"), []byte("frozen\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestFrozenVerificationRejectsCorruption(t *testing.T) {
	root := t.TempDir()
	dir := fixture(t, root, "a", 100, 10)
	p := filepath.Join(dir, "outputs", "transition-network", "transition_network_summary.yaml")
	if err := os.WriteFile(p, []byte("corrupt\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadExperiment(dir, false); err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("expected checksum refusal, got %v", err)
	}
}
func TestMissingStageIsNotZero(t *testing.T) {
	root := t.TempDir()
	dir := fixture(t, root, "a", 100, 10)
	e, err := loadExperiment(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := e.Normalized["higher_order.replication_rate"]; ok {
		t.Fatal("missing higher-order artifact became a metric")
	}
}
func TestDistanceOracle(t *testing.T) {
	if got := distance("manhattan", []float64{1, 2}, []float64{4, 6}, []float64{1, 1}); got != 7 {
		t.Fatalf("manhattan=%v", got)
	}
	if got := distance("standardized_euclidean", []float64{1, 2}, []float64{4, 6}, []float64{3, 4}); got != mathSqrt(2) {
		t.Fatalf("standardized euclidean=%v", got)
	}
}

func TestTransitionRetentionSemantics(t *testing.T) {
	root := t.TempDir()
	dir := fixture(t, root, "transition", 100, 10)
	e, err := loadExperiment(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if got := e.Normalized["transition.preferred_backbone_retention"]; got != 0.5 {
		t.Fatalf("preferred retention=%v", got)
	}
	if _, ok := e.Normalized["transition.preferred_rate"]; ok {
		t.Fatal("prototype preferred_rate leaked into v2")
	}
}

func TestLegacyMissingTask49AndSchemaFamilies(t *testing.T) {
	root := t.TempDir()
	dir := fixture(t, root, "legacy", 100, 10)
	e, err := loadExperiment(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if value, status := val(e, "vocabulary_growth.heaps_beta"); value != "" || status != "MISSING_ARTIFACT" {
		t.Fatalf("legacy Task49 status=%q/%q", value, status)
	}
	if len(fingerprintFeatures()) == 0 {
		t.Fatal("fingerprint feature registry is empty")
	}
}

func TestDistanceKindPrefixes(t *testing.T) {
	a, b := Experiment{ID: "a", Normalized: map[string]float64{"x": 1}}, Experiment{ID: "b", Normalized: map[string]float64{"x": 4}}
	row := distanceRow(a, b, "pairwise_available_manhattan", []string{"x"}, map[string]float64{"x": 1})
	if got := parse(row[3]); got != 3 {
		t.Fatalf("manhattan distance=%v", got)
	}
}

func TestTask49OptionalArtifact(t *testing.T) {
	root := t.TempDir()
	dir := fixture(t, root, "task49", 100, 10)
	addTask49(t, dir)
	e, err := loadExperiment(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if e.Raw["vocabulary_growth.heaps_beta"] != 0.7 {
		t.Fatalf("heaps beta=%v", e.Raw["vocabulary_growth.heaps_beta"])
	}
	if _, status := val(e, "vocabulary_growth.V_1000_per_token"); status != "NOT_APPLICABLE" {
		t.Fatalf("checkpoint status=%s", status)
	}
}
func mathSqrt(v float64) float64 { return 1.4142135623730951 }
