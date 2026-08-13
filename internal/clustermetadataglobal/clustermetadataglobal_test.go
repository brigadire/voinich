package clustermetadataglobal

import (
	"math"
	"math/rand"
	"strconv"
	"testing"

	"zcore.dev/voinich/internal/metadatavalidation"
)

// withSearchSpace temporarily overrides the frozen window_size/method sweep
// for a test and restores it afterwards. K is left at its real KMin..KMax
// bounds (they are package constants).
func withSearchSpace(t *testing.T, windowSizes []int, methods []string) {
	t.Helper()
	origWS, origM := WindowSizes, Methods
	WindowSizes, Methods = windowSizes, methods
	t.Cleanup(func() { WindowSizes, Methods = origWS, origM })
}

// buildFrozenSpace assembles a synthetic frozen search space covering
// exactly the current package WindowSizes x Methods x KMin..KMax, with
// windows and cluster ids produced by the supplied callbacks.
func buildFrozenSpace(t *testing.T, windows map[int][]WindowRange, clusterFn func(ws int, method string, k, windowIndex int) int) *frozenSpace {
	t.Helper()
	fs := &frozenSpace{Windows: windows, Combos: map[comboKey]comboData{}}
	for _, ranges := range windows {
		for _, r := range ranges {
			if r.End > fs.N {
				fs.N = r.End
			}
		}
	}
	for _, ws := range WindowSizes {
		ranges, ok := windows[ws]
		if !ok {
			t.Fatalf("test frozen space missing window_size=%d", ws)
		}
		for _, m := range Methods {
			for k := KMin; k <= KMax; k++ {
				cluster := make([]int, len(ranges))
				maxC := 0
				for i, r := range ranges {
					c := clusterFn(ws, m, k, r.Index)
					cluster[i] = c
					if c+1 > maxC {
						maxC = c + 1
					}
				}
				fs.Combos[comboKey{ws, m, k}] = comboData{Cluster: cluster, NumClusters: maxC}
			}
		}
	}
	return fs
}

func windowsOf(size, n int) []WindowRange {
	out := []WindowRange{}
	for i, start := 0, 0; start+size <= n; i, start = i+1, start+size {
		out = append(out, WindowRange{i, start, start + size})
	}
	return out
}

// alternatingLabels builds N tokens grouped into windows of `size` tokens,
// each window entirely "A" or entirely "B", alternating by window index.
func alternatingLabels(size, n int) []string {
	out := make([]string, n)
	for i := range out {
		if (i/size)%2 == 0 {
			out[i] = "A"
		} else {
			out[i] = "B"
		}
	}
	return out
}

func TestMaxOverK(t *testing.T) {
	withSearchSpace(t, []int{10}, []string{"m"})
	n := 100
	windows := map[int][]WindowRange{10: windowsOf(10, n)}
	// K=2 exactly reproduces the alternating A/B window partition; every
	// other K uses a mismatched grouping.
	fs := buildFrozenSpace(t, windows, func(ws int, method string, k, idx int) int {
		if k == 2 {
			return idx % 2
		}
		return idx % k
	})
	labels := alternatingLabels(10, n)
	series, _ := RunSearchSpace(fs, labels, labels, 0, 1, nil)
	s := series[seriesKey("currier", "primary", "NMI", "m")]
	if math.Abs(s.Observed-1.0) > 1e-9 {
		t.Fatalf("expected perfect max NMI 1.0 at K=2, got %v (window=%d k=%d)", s.Observed, s.ObservedWindow, s.ObservedK)
	}
	if s.ObservedK != 2 {
		t.Fatalf("expected max-over-K to select K=2, got K=%d", s.ObservedK)
	}
}

func TestMaxOverWindowTimesK(t *testing.T) {
	withSearchSpace(t, []int{10, 20}, []string{"m"})
	n := 100
	windows := map[int][]WindowRange{10: windowsOf(10, n), 20: windowsOf(20, n)}
	fs := buildFrozenSpace(t, windows, func(ws int, method string, k, idx int) int {
		if ws == 10 && k == 2 {
			return idx % 2 // perfect match against the size-10 alternation
		}
		// size-20 windows straddle two size-10 labels each (A,B) so no K at
		// window_size=20 can reach a perfect partition; use a fixed
		// deliberately weak grouping.
		return idx % k
	})
	labels := alternatingLabels(10, n)
	series, _ := RunSearchSpace(fs, labels, labels, 0, 1, nil)
	s := series[seriesKey("currier", "primary", "NMI", "m")]
	if math.Abs(s.Observed-1.0) > 1e-9 {
		t.Fatalf("expected per-method max over window x K to reach 1.0, got %v", s.Observed)
	}
	if s.ObservedWindow != 10 || s.ObservedK != 2 {
		t.Fatalf("expected argmax window=10 k=2, got window=%d k=%d", s.ObservedWindow, s.ObservedK)
	}
}

func TestMaxOverWindowMethodK(t *testing.T) {
	withSearchSpace(t, []int{10}, []string{"weak", "strong"})
	n := 100
	windows := map[int][]WindowRange{10: windowsOf(10, n)}
	fs := buildFrozenSpace(t, windows, func(ws int, method string, k, idx int) int {
		if method == "strong" && k == 2 {
			return idx % 2
		}
		if method == "weak" {
			return 0 // degenerate single cluster: NMI/ARI stay at 0 for every K
		}
		return idx % k
	})
	labels := alternatingLabels(10, n)
	series, _ := RunSearchSpace(fs, labels, labels, 0, 1, nil)
	g := series[seriesKey("currier", "primary", "NMI", "global")]
	if math.Abs(g.Observed-1.0) > 1e-9 {
		t.Fatalf("expected global max over window x method x K to reach 1.0, got %v", g.Observed)
	}
	if g.ObservedMethod != "strong" || g.ObservedK != 2 {
		t.Fatalf("expected global argmax method=strong k=2, got method=%s k=%d", g.ObservedMethod, g.ObservedK)
	}
	weak := series[seriesKey("currier", "primary", "NMI", "weak")]
	if weak.Observed >= g.Observed {
		t.Fatalf("expected the weak method's own max (%v) to stay below the global max (%v)", weak.Observed, g.Observed)
	}
}

func TestScaleMeanAndMinStatistics(t *testing.T) {
	withSearchSpace(t, []int{10, 20, 30}, []string{"only"})
	nk := KMax - KMin + 1
	values := make3D(1, 3, nk)
	rowMaxes := []float64{0.9, 0.2, 0.5}
	for si, rowMax := range rowMaxes {
		for ki := range values[0][si] {
			values[0][si][ki] = rowMax - float64(ki)*0.001 // rowMax is the max of the row
		}
	}
	d := derive(values)
	wantMean := (0.9 + 0.2 + 0.5) / 3
	if math.Abs(d.PersistenceMean[0]-wantMean) > 1e-9 {
		t.Fatalf("persistence mean = %v, want %v", d.PersistenceMean[0], wantMean)
	}
	if math.Abs(d.PersistenceMin[0]-0.2) > 1e-9 {
		t.Fatalf("persistence min = %v, want 0.2", d.PersistenceMin[0])
	}
}

func TestEmpiricalPPlusOneCorrection(t *testing.T) {
	null := make([]float64, 999)
	for i := range null {
		null[i] = 0.1
	}
	// No null replicate reaches the observed value: exceedances must be 0,
	// but empirical_p must never be exactly zero.
	p := empiricalP(exceedances(null, 0.9), len(null))
	if p <= 0 {
		t.Fatalf("empirical p must never be exactly zero, got %v", p)
	}
	wantP := 1.0 / 1000.0
	if math.Abs(p-wantP) > 1e-12 {
		t.Fatalf("empirical p = %v, want %v", p, wantP)
	}
	// Every null replicate exceeds observed: p should approach 1, still using +1.
	all := make([]float64, 9)
	for i := range all {
		all[i] = 1.0
	}
	p2 := empiricalP(exceedances(all, 0.0), len(all))
	if math.Abs(p2-1.0) > 1e-12 {
		t.Fatalf("empirical p = %v, want 1.0 when every replicate exceeds observed", p2)
	}
}

func TestDeterministicSeed(t *testing.T) {
	withSearchSpace(t, []int{10, 20}, []string{"m1", "m2"})
	n := 60
	windows := map[int][]WindowRange{10: windowsOf(10, n), 20: windowsOf(20, n)}
	fs := buildFrozenSpace(t, windows, func(ws int, method string, k, idx int) int {
		return (idx + k + len(method)) % k
	})
	currier := alternatingLabels(10, n)
	hand := alternatingLabels(5, n)
	s1, _ := RunSearchSpace(fs, currier, hand, 25, 7, nil)
	s2, _ := RunSearchSpace(fs, currier, hand, 25, 7, nil)
	key := seriesKey("hand", "primary", "NMI", "global")
	for i := range s1[key].Null {
		if s1[key].Null[i] != s2[key].Null[i] {
			t.Fatalf("same seed produced different null replicate %d: %v vs %v", i, s1[key].Null[i], s2[key].Null[i])
		}
	}
	s3, _ := RunSearchSpace(fs, currier, hand, 25, 8, nil)
	same := true
	for i := range s1[key].Null {
		if s1[key].Null[i] != s3[key].Null[i] {
			same = false
			break
		}
	}
	if same {
		t.Fatalf("different seeds produced an identical null distribution; the permutation stream is not seed-sensitive")
	}
}

func TestCurrierUnknownMaskPreservation(t *testing.T) {
	labels := []string{"X", "X", "", "", "", "Y", "Y", "Y", "", "X", "X", "X", "X", "", "Y"}
	blocks := blocksOf(labels)
	rng := rand.New(rand.NewSource(42))
	for iter := 0; iter < 200; iter++ {
		out := permuteKnownBlocks(blocks, rng)
		if len(out) != len(labels) {
			t.Fatalf("permuted length %d != original %d", len(out), len(labels))
		}
		for i, v := range out {
			if (labels[i] == "") != (v == "") {
				t.Fatalf("iteration %d: unknown mask changed at position %d: real=%q permuted=%q", iter, i, labels[i], v)
			}
		}
	}
}

// TestSamePermutationReusedAcrossSearchSpace verifies that one permutation
// replicate builds exactly one cumulative label table that is then shared,
// unchanged, by every frozen window size. If a bug drew an independent
// permutation per window size, per-label counts over a coarse window would
// generically stop matching the sum of its constituent fine windows' counts
// under the SAME replicate.
func TestSamePermutationReusedAcrossSearchSpace(t *testing.T) {
	n := 100
	labels := make([]string, n)
	for i := range labels {
		if i%3 == 0 {
			labels[i] = ""
		} else if (i/7)%2 == 0 {
			labels[i] = "X"
		} else {
			labels[i] = "Y"
		}
	}
	fs := &frozenSpace{N: n}
	prep := prepareKind(fs, labels)
	rng := rand.New(rand.NewSource(99))
	permuted := permuteKnownBlocks(prep.blocks, rng)
	codes := codesFromLabels(permuted, prep.valueIndex)
	cum := make([]int32, (n+1)*prep.numLabels)
	buildCumulativeInto(codes, prep.numLabels, cum)

	fine := windowsOf(50, n)   // two size-50 windows
	coarse := windowsOf(100, n) // one size-100 window spanning both

	for c := 0; c < prep.numLabels; c++ {
		fineSum := int32(0)
		for _, r := range fine {
			fineSum += cum[r.End*prep.numLabels+c] - cum[r.Start*prep.numLabels+c]
		}
		coarseSum := cum[coarse[0].End*prep.numLabels+c] - cum[coarse[0].Start*prep.numLabels+c]
		if fineSum != coarseSum {
			t.Fatalf("label code %d: size-50 windows summed to %d but the size-100 window (same permutation replicate) shows %d", c, fineSum, coarseSum)
		}
	}
}

// TestFastMetricsMatchesReferenceFormula checks that the array-based fast
// path used inside the 10000-replicate permutation loop computes exactly the
// same NMI/ARI as metadatavalidation.AssociationMetrics's map-based
// reference implementation, on the same categorical data.
func TestFastMetricsMatchesReferenceFormula(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	labelStrings := []string{"A", "B", "C"}
	n := 500
	labels := make([]string, n)
	clusters := make([]string, n)
	labelCodes := make([]int8, n)
	clusterCodes := make([]int, n)
	for i := 0; i < n; i++ {
		li := rng.Intn(3)
		ci := rng.Intn(5)
		labels[i] = labelStrings[li]
		labelCodes[i] = int8(li)
		clusters[i] = strconv.Itoa(ci)
		clusterCodes[i] = ci
	}
	eligible := make([]int, n)
	for i := range eligible {
		eligible[i] = i
	}
	wantNMI := metadatavalidation.AssociationMetrics(labels, clusters).NMI
	wantARI := metadatavalidation.AssociationMetrics(labels, clusters).ARI
	gotNMI, gotARI := fastMetrics(labelCodes, clusterCodes, eligible, 3, 5)
	if math.Abs(gotNMI-wantNMI) > 1e-9 {
		t.Fatalf("fast NMI %v != reference NMI %v", gotNMI, wantNMI)
	}
	if math.Abs(gotARI-wantARI) > 1e-9 {
		t.Fatalf("fast ARI %v != reference ARI %v", gotARI, wantARI)
	}
}
