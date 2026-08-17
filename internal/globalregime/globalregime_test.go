package globalregime

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMetricsAndSlidingWindows(t *testing.T) {
	a := profile{"a": 1}
	b := profile{"b": 1}
	if got := jsDistance(a, a); got != 0 {
		t.Fatalf("identical JS distance=%g", got)
	}
	if got := jsDistance(a, b); got < .999999 {
		t.Fatalf("disjoint JS distance=%g", got)
	}
	w := slidingWindows([]string{"a", "a", "a", "a", "b", "b", "b", "b"}, 4, 2)
	if len(w) != 3 {
		t.Fatalf("windows=%d", len(w))
	}
	if w[0].Start != 0 || w[2].End != 8 {
		t.Fatalf("window bounds: %#v", w)
	}
	if w[1].WeightedOverlap != .5 {
		t.Fatalf("overlap=%g", w[1].WeightedOverlap)
	}
	// A user-supplied step may leave gaps between windows.
	w = slidingWindows([]string{"a", "a", "x", "x", "b", "b"}, 2, 4)
	if len(w) != 2 || w[1].distribution["b"] != 1 {
		t.Fatalf("gapped windows: %#v", w)
	}
}

func TestSlidingWindowsUsesActualCorpusSize(t *testing.T) {
	const (
		windowSize = 1000
		step       = 100
	)
	for _, n := range []int{1000, 8000, 39026, 60017} {
		t.Run(fmt.Sprintf("N=%d", n), func(t *testing.T) {
			tokens := make([]string, n)
			for i := range tokens {
				tokens[i] = fmt.Sprintf("token-%d", i%17)
			}

			windows := slidingWindows(tokens, windowSize, step)
			want := 1 + (n-windowSize)/step
			if len(windows) != want {
				t.Fatalf("windows=%d, want %d for corpus size %d", len(windows), want, n)
			}
			for i, w := range windows {
				if w.Start != i*step || w.End != i*step+windowSize || w.End > n {
					t.Fatalf("window %d has invalid bounds [%d,%d) for corpus size %d", i, w.Start, w.End, n)
				}
				total := 0.0
				for _, probability := range w.distribution {
					total += probability
				}
				if math.Abs(total-1) > 1e-12 {
					t.Fatalf("window %d distribution sums to %.17g, want 1", i, total)
				}
			}
		})
	}
}

func TestShortCorpusReportsFixedWindowAsNotApplicable(t *testing.T) {
	d := t.TempDir()
	corpus := filepath.Join(d, "corpus.txt")
	if err := os.WriteFile(corpus, []byte(strings.Repeat("token ", 999)), 0o644); err != nil {
		t.Fatal(err)
	}
	err := RunAndWrite(Config{CorpusPath: corpus, OutputDir: d, Quiet: true})
	if err == nil {
		t.Fatal("short corpus unexpectedly accepted all fixed window sizes")
	}
	if want := "window size 1000 exceeds corpus length 999"; !strings.Contains(err.Error(), want) {
		t.Fatalf("error=%q, want diagnostic containing %q", err, want)
	}
}

func TestStableBoundariesCountEachScaleOnce(t *testing.T) {
	x := []ChangePoint{{50, 100, "threshold", .4, 0}, {50, 102, "pelt", .3, 0}, {100, 110, "threshold", .2, 0}, {500, 900, "threshold", .8, 0}}
	b := stableBoundaries(x, []int{50, 100, 500})
	if len(b) < 2 {
		t.Fatalf("boundaries=%#v", b)
	}
	if b[0].SupportCount != 2 {
		t.Fatalf("support counted detector duplicates: %#v", b[0])
	}
	if b[0].SupportFraction != 2.0/3.0 {
		t.Fatalf("fraction=%g", b[0].SupportFraction)
	}
}

func TestRunWritesContractAndProgress(t *testing.T) {
	d := t.TempDir()
	corpus := filepath.Join(d, "corpus.txt")
	var tokens []string
	for i := 0; i < 40; i++ {
		tokens = append(tokens, "a")
	}
	for i := 0; i < 40; i++ {
		tokens = append(tokens, "b")
	}
	if err := os.WriteFile(corpus, []byte(strings.Join(tokens, " ")), 0o644); err != nil {
		t.Fatal(err)
	}
	var progress bytes.Buffer
	err := RunAndWrite(Config{CorpusPath: corpus, OutputDir: d, WindowSizes: []int{10, 20}, Step: 5, Seed: 1, ProgressWriter: &progress})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"global_distributional_regimes.yaml", "global_distributional_windows.tsv", "global_distributional_change_points.tsv", "stable_distributional_boundaries.tsv", "global_distributional_clustering.tsv", "global_distributional_cluster_assignments.tsv", "global_distributional_report.md", "plots/global_distributional_change_profiles.svg"} {
		if _, err := os.Stat(filepath.Join(d, name)); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	for _, want := range []string{"[1/7]", "[7/7]", "elapsed", "Building multi-scale"} {
		if !strings.Contains(progress.String(), want) {
			t.Errorf("progress lacks %q: %s", want, progress.String())
		}
	}
}

func TestValidation(t *testing.T) {
	if _, err := analyze(Config{CorpusPath: "x", OutputDir: "x", WindowSizes: []int{1}}, nil); err == nil {
		t.Fatal("accepted size one")
	}
}
