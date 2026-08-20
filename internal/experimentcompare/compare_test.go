package experimentcompare

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
func mathSqrt(v float64) float64 { return 1.4142135623730951 }
