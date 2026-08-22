package tokenrepetition

import "testing"

func TestRunningTextSubsampleNullDeterministic(t *testing.T) {
	running := make([]string, 0, 200)
	lineOf := make([]int, 0, 200)
	words := []string{"a", "b", "c", "d", "e"}
	for i := 0; i < 200; i++ {
		running = append(running, words[i%len(words)])
		lineOf = append(lineOf, i/20)
	}
	a := RunningTextSubsampleNull(running, lineOf, 30, 10, GlyphNatural, newTestRand(42))
	b := RunningTextSubsampleNull(running, lineOf, 30, 10, GlyphNatural, newTestRand(42))
	if len(a) != len(b) {
		t.Fatalf("length mismatch: %d != %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("non-deterministic at %d: %+v != %+v", i, a[i], b[i])
		}
	}
}
