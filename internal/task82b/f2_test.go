package task82b

import (
	"testing"
)

func TestExtractF2Smoke(t *testing.T) {
	groups := [][]string{
		{"the", "quick", "brown", "fox"},
		{"jumps", "over", "the", "lazy", "dog"},
		{"the", "fox", "runs"},
	}
	dir := t.TempDir()
	path := dir + "/corpus.txt"
	if err := WriteCorpusFile(path, groups); err != nil {
		t.Fatal(err)
	}
	v, err := ExtractF2(path, "job1", "smoke", 1, dir+"/out", groups)
	if err != nil {
		t.Fatal(err)
	}
	if len(v.Metrics) != len(AllMetricIDs()) {
		t.Fatalf("got %d metrics, want %d", len(v.Metrics), len(AllMetricIDs()))
	}
	for _, m := range v.Metrics {
		t.Logf("%-40s available=%v value=%v reason=%q", m.MetricID, m.Available, m.Value, m.MissingReason)
	}
}
