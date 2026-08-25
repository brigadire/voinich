package main

import "math"

// MemorizationResult applies G1_EXECUTABLE_CONTRACT.json failure.MEMORIZATION_DOMINATED:
// DEVELOPMENT-to-HELDOUT PM2 gap exceeds the max MFC q0.95 gap AND
// Complexity/(N_dev*log2(alphabet_size+2)) >= 0.90.
type MemorizationResult struct {
	Gap          float64
	GapThreshold float64
	GapExceeds   bool
	Ratio        float64
	RatioExceeds bool
	Dominated    bool
}

const memorizationRatioThreshold = 0.90

func EvalMemorization(class, candID string, devPM2, heldPM2, complexity float64, nDev, alphabetSize int, idx *ThresholdIndex) MemorizationResult {
	gap := devPM2 - heldPM2
	t, ok := idx.Get("overfitting_gap", "PM2", class, candID)
	r := MemorizationResult{Gap: gap}
	if ok {
		r.GapThreshold = t.Threshold
		r.GapExceeds = isFinite(gap) && isFinite(t.Threshold) && gap > t.Threshold
	}
	denom := float64(nDev) * math.Log2(float64(alphabetSize)+2)
	if denom > 0 {
		r.Ratio = complexity / denom
	}
	r.RatioExceeds = r.Ratio >= memorizationRatioThreshold
	r.Dominated = r.GapExceeds && r.RatioExceeds
	return r
}
