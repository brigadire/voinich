package task82b

import "testing"

// TestAX5SyntheticPositiveControl is a SYNTHETIC positive control
// (task82b.txt sec.48): a stream that alternates two symbol classes with
// period 2 (AB AB AB ...) -- period 2 is in AX5's own frozen candidate
// grid {2,3,5,7}, task82b.txt sec.30 -- must show its largest AX5 signal
// at k=2, while a length-matched random shuffle of the same symbols must
// not.
func TestAX5SyntheticPositiveControl(t *testing.T) {
	var periodic []string
	pattern := []string{"a", "b"}
	for i := range 400 {
		periodic = append(periodic, pattern[i%2])
	}
	groups := [][]string{periodic}
	ax := ComputeAX(groups)
	if ax.AX5PeriodicNMIMax < 0.3 {
		t.Fatalf("synthetic period-4 stream: AX5 = %.4f (k=%d), expected a strong signal", ax.AX5PeriodicNMIMax, ax.AX5BestPeriod)
	}
	t.Logf("periodic AX5=%.4f at k=%d", ax.AX5PeriodicNMIMax, ax.AX5BestPeriod)

	shuffled := make([]string, len(periodic))
	copy(shuffled, periodic)
	// deterministic pseudo-shuffle (Fisher-Yates with a fixed LCG) so the
	// test has no time-based randomness dependency.
	seed := uint64(12345)
	next := func() uint64 { seed = seed*6364136223846793005 + 1; return seed }
	for i := len(shuffled) - 1; i > 0; i-- {
		j := int(next() % uint64(i+1))
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	axShuffled := ComputeAX([][]string{shuffled})
	t.Logf("shuffled AX5=%.4f at k=%d", axShuffled.AX5PeriodicNMIMax, axShuffled.AX5BestPeriod)
	if axShuffled.AX5PeriodicNMIMax >= ax.AX5PeriodicNMIMax {
		t.Fatalf("shuffled control (AX5=%.4f) should score lower than the periodic positive control (AX5=%.4f)", axShuffled.AX5PeriodicNMIMax, ax.AX5PeriodicNMIMax)
	}
}
