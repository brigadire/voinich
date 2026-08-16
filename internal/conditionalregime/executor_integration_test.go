package conditionalregime

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// hashOutputDir returns a sorted relative-path -> sha256 map for every file
// RunAndWrite wrote, so two runs can be compared byte-for-byte without ever
// diffing raw bytes in a failure message.
func hashOutputDir(t *testing.T, dir string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		out[rel] = fmt.Sprintf("%x", sha256.Sum256(b))
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", dir, err)
	}
	if len(out) == 0 {
		t.Fatalf("no output files found under %s", dir)
	}
	return out
}

func assertIdenticalOutputs(t *testing.T, label string, a, b map[string]string) {
	t.Helper()
	if len(a) != len(b) {
		t.Fatalf("%s: output file count differs: %d vs %d", label, len(a), len(b))
	}
	names := make([]string, 0, len(a))
	for name := range a {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		hb, ok := b[name]
		if !ok {
			t.Fatalf("%s: %s present in first run, missing from second", label, name)
		}
		if a[name] != hb {
			t.Fatalf("%s: %s differs: %s vs %s", label, name, a[name], hb)
		}
	}
}

// runFixturePipeline runs the full RunAndWrite pipeline against the shared
// tiny fixture with checkpointing disabled and returns the hash of every
// written output file.
func runFixturePipeline(t *testing.T, f fixture, executor string, workers int) map[string]string {
	t.Helper()
	c := f.smallConfig()
	c.OutputDir = t.TempDir()
	c.CheckpointPath = "-"
	c.Quiet = true
	c.Executor = executor
	c.Workers = workers
	c.Context = context.Background()
	if err := RunAndWrite(c); err != nil {
		t.Fatalf("RunAndWrite(executor=%s, workers=%d): %v", executor, workers, err)
	}
	return hashOutputDir(t, c.OutputDir)
}

// TestProcessExecutorMatchesGoroutineExecutorByteForByte is Task32's primary
// invariant, automated: sequential/goroutine output and multi-process output
// must be bit-for-bit identical for the identical corpus, parameters and
// seed, at more than one worker count in each backend (phase 7, items 1-4).
func TestProcessExecutorMatchesGoroutineExecutorByteForByte(t *testing.T) {
	if testing.Short() {
		t.Skip("spawns real subprocess workers; skipped in -short")
	}
	f := writeFixture(t)

	oracle := runFixturePipeline(t, f, "goroutine", 1)
	goroutineFour := runFixturePipeline(t, f, "goroutine", 4)
	assertIdenticalOutputs(t, "goroutine workers=1 vs workers=4", oracle, goroutineFour)

	processOne := runFixturePipeline(t, f, "process", 1)
	assertIdenticalOutputs(t, "oracle vs process workers=1", oracle, processOne)

	processFour := runFixturePipeline(t, f, "process", 4)
	assertIdenticalOutputs(t, "oracle vs process workers=4", oracle, processFour)
}
