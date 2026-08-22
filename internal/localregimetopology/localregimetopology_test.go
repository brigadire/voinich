package localregimetopology

import (
	"strings"
	"testing"
)

func tok(s string) []string { return strings.Split(s, "") }

func recs(n int, folioOf func(i int) string, lineOf func(i int) int) []TokenRecord {
	out := make([]TokenRecord, n)
	for i := 0; i < n; i++ {
		out[i] = TokenRecord{GlobalIndex: i, Line: lineOf(i), Folio: folioOf(i), Glyphs: tok("a")}
	}
	return out
}

func TestBuildWindowsSizeAndStep(t *testing.T) {
	r := recs(10, func(i int) string { return "f1r" }, func(i int) int { return 0 })
	ws := BuildWindows(r, 4, 2)
	// positions: 0,2,4 -> starts 0,2,4,6 (6+4=10 fits); next 8+4=12 doesn't fit.
	if len(ws) != 4 {
		t.Fatalf("want 4 windows, got %d", len(ws))
	}
	for _, w := range ws {
		if len(w.Tokens) != 4 {
			t.Fatalf("window size wrong: %d", len(w.Tokens))
		}
	}
}

func TestBuildWindowsDetectsPageAndLineCrossing(t *testing.T) {
	r := recs(6, func(i int) string {
		if i < 3 {
			return "f1r"
		}
		return "f1v"
	}, func(i int) int { return i / 2 })
	ws := BuildWindows(r, 4, 4)
	if len(ws) != 1 {
		t.Fatalf("want 1 window, got %d", len(ws))
	}
	if !ws[0].CrossesPage {
		t.Fatal("window spanning f1r/f1v should be marked CrossesPage")
	}
	if !ws[0].CrossesLine {
		t.Fatal("window spanning multiple lines should be marked CrossesLine")
	}
}

func TestBuildWindowsStepGuard(t *testing.T) {
	r := recs(5, func(i int) string { return "f1r" }, func(i int) int { return 0 })
	ws := BuildWindows(r, 3, 0)
	if len(ws) != 3 {
		t.Fatalf("step<1 should behave like step=1: want 3 windows, got %d", len(ws))
	}
}

func TestScanChangePointsDeterministic(t *testing.T) {
	r := recs(20, func(i int) string { return "f1r" }, func(i int) int { return 0 })
	// vary token length (not just which glyph dominates), since Profile is
	// deliberately identity-agnostic: a uniform run of any single glyph
	// looks the same as a uniform run of any other single glyph.
	for i := 10; i < 20; i++ {
		r[i].Glyphs = tok("zzz")
	}
	giant := map[string]bool{}
	a := ScanChangePoints(r, 5, giant)
	b := ScanChangePoints(r, 5, giant)
	if len(a) != len(b) {
		t.Fatal("non-deterministic change-point count")
	}
	maxScore := 0.0
	for i := range a {
		if a[i].Score != b[i].Score {
			t.Fatal("non-deterministic change-point scores")
		}
		if a[i].Score > maxScore {
			maxScore = a[i].Score
		}
	}
	if maxScore <= 0 {
		t.Fatal("expected a nonzero change-point score across the a/z transition")
	}
}

func TestCUSUMMaxFindsPeak(t *testing.T) {
	// a step change (0 -> 5 at index 3): the cumulative deviation from the
	// series mean grows more negative up to the last point still below the
	// mean, then climbs back toward 0 - its extreme sits at the last
	// pre-change point (index 2), which is what marks the changepoint.
	scores := []float64{0, 0, 0, 5, 5, 5, 5}
	peakIdx, peakVal := CUSUMMax(scores)
	if peakIdx != 2 {
		t.Fatalf("want peak at index 2 (last pre-change point), got %d", peakIdx)
	}
	if peakVal <= 0 {
		t.Fatal("want positive peak value")
	}
	if pi, _ := CUSUMMax(nil); pi != -1 {
		t.Fatal("empty series should return -1, not crash")
	}
}

func TestKMedoidsDeterministicAndSeparatesClusters(t *testing.T) {
	vectors := [][]float64{{0, 0}, {0.1, 0}, {10, 10}, {10.1, 10}}
	a1, m1 := KMedoids(vectors, 2, 10)
	a2, m2 := KMedoids(vectors, 2, 10)
	for i := range a1 {
		if a1[i] != a2[i] {
			t.Fatal("k-medoids assignment not deterministic")
		}
	}
	if len(m1) != len(m2) {
		t.Fatal("medoid count mismatch")
	}
	if a1[0] != a1[1] || a1[2] != a1[3] || a1[0] == a1[2] {
		t.Fatalf("expected {0,1} and {2,3} to form separate clusters, got %v", a1)
	}
}

func TestStandardizeColumnsUsesTrainOnly(t *testing.T) {
	vectors := [][]float64{{0}, {2}, {100}}
	out := StandardizeColumns(vectors, []int{0, 1})
	// mean/sd computed from rows 0,1 only (mean=1, sd=1): row0 -> -1, row1 -> +1.
	if out[0][0] != -1 || out[1][0] != 1 {
		t.Fatalf("unexpected standardized values: %v", out)
	}
	zero := StandardizeColumns([][]float64{{5}, {5}}, []int{0, 1})
	if zero[0][0] != 0 || zero[1][0] != 0 {
		t.Fatalf("zero-variance column should standardize to 0, got %v", zero)
	}
}
