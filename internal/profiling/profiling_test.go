package profiling

import (
	"os"
	"strings"
	"testing"
	"time"
)

// TestMemStatsIntervalDisabledByDefault confirms the sampler goroutine is
// never started when MemStatsInterval is left at its zero value (the
// default for every CLI that doesn't pass -memstats-interval) - Start/Stop
// must remain a complete no-op in that case, matching every other field in
// Config.
func TestMemStatsIntervalDisabledByDefault(t *testing.T) {
	sess, err := Start(&Config{})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if sess.memStatsStop != nil {
		t.Fatal("memStatsStop channel created despite MemStatsInterval being 0")
	}
	if err := sess.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
}

// TestMemStatsIntervalStopsCleanly confirms Stop() waits for the sampler
// goroutine to exit (no leaked goroutine, no race) rather than merely
// signaling it, and that it logs at least one snapshot to the given writer
// within a couple of intervals.
func TestMemStatsIntervalStopsCleanly(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe: %v", err)
	}
	origStderr := os.Stderr
	os.Stderr = w
	defer func() { os.Stderr = origStderr }()

	sess, err := Start(&Config{MemStatsInterval: 20 * time.Millisecond})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if sess.memStatsStop == nil {
		t.Fatal("memStatsStop channel not created despite MemStatsInterval being set")
	}
	time.Sleep(90 * time.Millisecond) // let a few ticks fire
	done := make(chan error, 1)
	go func() { done <- sess.Stop() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Stop: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Stop did not return - sampler goroutine likely leaked")
	}
	w.Close()
	os.Stderr = origStderr

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	out := string(buf[:n])
	if !strings.Contains(out, "memstats: HeapAlloc=") {
		t.Fatalf("expected at least one memstats snapshot on stderr, got: %q", out)
	}
}
