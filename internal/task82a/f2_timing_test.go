package task82a

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestF2Timing(t *testing.T) {
	if os.Getenv("TASK82A_TIMING") == "" {
		t.Skip("set TASK82A_TIMING=1 to run manual timing pilot")
	}
	tmp := t.TempDir()
	for _, n := range []int{64, 256, 1024, 4096} {
		tokens := make([]string, n)
		for i := 0; i < n; i++ {
			tokens[i] = fmt.Sprintf("C%d", i)
		}
		path := filepath.Join(tmp, fmt.Sprintf("global_%d.txt", n))
		lines := make([][]string, n)
		for i, tok := range tokens {
			lines[i] = []string{tok}
		}
		if err := writeAssembledCorpusFile(path, lines); err != nil {
			t.Fatal(err)
		}
		start := time.Now()
		metrics, warnings, err := extractF2(path, "timing-test", 1, filepath.Join(tmp, "out"))
		if err != nil {
			t.Fatalf("n=%d: %v", n, err)
		}
		t.Logf("GLOBAL n=%d elapsed=%s warnings=%v", n, time.Since(start), warnings)
		for _, m := range metrics {
			t.Logf("  %-45s available=%v value=%v reason=%q", m.MetricID, m.Available, m.Value, m.MissingReason)
		}
	}
	for _, n := range []int{64, 256, 1024, 4096} {
		tokens := make([]string, n)
		for i := 0; i < n; i++ {
			tokens[i] = fmt.Sprintf("C%d", i%4)
		}
		path := filepath.Join(tmp, fmt.Sprintf("local_%d.txt", n))
		lines := make([][]string, n)
		for i, tok := range tokens {
			lines[i] = []string{tok}
		}
		if err := writeAssembledCorpusFile(path, lines); err != nil {
			t.Fatal(err)
		}
		start := time.Now()
		_, _, err := extractF2(path, "timing-test", 1, filepath.Join(tmp, "out2"))
		if err != nil {
			t.Fatalf("n=%d: %v", n, err)
		}
		t.Logf("LOCAL n=%d elapsed=%s", n, time.Since(start))
	}
}
