package main

import (
	"os"
	"path/filepath"
	"testing"
)

// test 21: held-out evaluation rejects unfrozen candidates.
func TestHeldoutRejectsUnfrozenCandidates(t *testing.T) {
	dir := t.TempDir()
	if _, err := ReadFrozenCandidates(dir); err == nil {
		t.Fatalf("expected an error reading frozen candidates before any freeze")
	}
}

// test 20 (development side of the guarantee): freezing then reading
// back returns exactly the frozen frontier, and re-freezing the same
// frontier is a no-op, but a different frontier is refused (section 42:
// no reworking a candidate after held-out access begins).
func TestFreezeCandidatesIsWriteOnceForAGivenFrontier(t *testing.T) {
	dir := t.TempDir()
	if err := FreezeCandidates(dir, []string{"M7_MIXED_K5_N20", "M4_STATE_K4_A"}); err != nil {
		t.Fatalf("freeze failed: %v", err)
	}
	got, err := ReadFrozenCandidates(dir)
	if err != nil {
		t.Fatalf("read after freeze failed: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 frozen candidates, got %v", got)
	}
	if err := FreezeCandidates(dir, []string{"M4_STATE_K4_A", "M7_MIXED_K5_N20"}); err != nil {
		t.Fatalf("re-freezing the identical frontier (different order) should be a no-op: %v", err)
	}
	if err := FreezeCandidates(dir, []string{"M4_STATE_K4_A"}); err == nil {
		t.Fatalf("expected re-freezing a DIFFERENT frontier to be refused")
	}
	if b, _ := os.ReadFile(filepath.Join(dir, FrozenMarker)); len(b) == 0 {
		t.Fatalf("frozen marker file missing")
	}
}
