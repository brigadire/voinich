// AX -- acrostic/extraction diagnostics (task82b.txt sec.43-50), an
// independent statistical family from Fingerprint V2 (sec.53: "SX/AX are
// not F2 v2.1"). Every function here is corpus-only and language-blind:
// no dictionary lookup, no anagram search, no plaintext decoding (sec.47).
//
// Of the seven candidate diagnostics sec.46 lists as a minimum to
// *consider*, this package implements four (AX3-AX6) as independent code.
// AX1 (information concentration at structural beginnings/endings), AX2
// (first-vs-internal distribution divergence) and AX7 (mutual information
// between structural position and token/glyph class) are, on inspection,
// exactly what F2's own BP1_BOUNDARY_TOKEN_NMI, LS2_POSITIONAL_LEXICON_NMI
// and 2DL1_LAYOUT_POSITION_MI already measure on the *carrier itself* --
// all three are already computed by ExtractF2 for every carrier, operator
// output and null in this study (F2_ASSEMBLER_PROJECTION, sec.3). Task82b
// therefore audits F2's own BP1/LS2/2DL1 values rather than reimplementing
// three redundant estimators (sec.61: do not double-count correlated
// metrics as independent evidence); see TASK82B_DESIGN.md's audit section
// and TASK82B_REPORT.md's answer to report question 17.
package task82b

import (
	"math"
	"sort"
)

// ShannonEntropyStrings returns the Shannon entropy, in bits, of the
// empirical distribution of a string sample. Keys are visited in sorted
// order before accumulation (project convention: deterministic float64
// accumulation order, see memory feedback_go_map_iteration_determinism).
func ShannonEntropyStrings(xs []string) float64 {
	counts := map[string]int{}
	for _, x := range xs {
		counts[x]++
	}
	return entropyOfCounts(counts, len(xs))
}

func entropyOfCounts(counts map[string]int, n int) float64 {
	if n == 0 {
		return 0
	}
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := 0.0
	for _, k := range keys {
		p := float64(counts[k]) / float64(n)
		if p > 0 {
			h -= p * math.Log2(p)
		}
	}
	return h
}

// normalizedMI computes NMI(xs;ys) = I(X;Y) / max(H(X),H(Y)) for two
// equal-length label sequences (0 if either marginal is degenerate). This
// is an independent AX implementation (AX5 only), not a reuse of
// fingerprintv2's unexported estimator.
func normalizedMI(xs, ys []string) float64 {
	n := len(xs)
	if n == 0 || n != len(ys) {
		return 0
	}
	joint := map[[2]string]int{}
	cx := map[string]int{}
	cy := map[string]int{}
	for i := range xs {
		joint[[2]string{xs[i], ys[i]}]++
		cx[xs[i]]++
		cy[ys[i]]++
	}
	hx := entropyOfCounts(cx, n)
	hy := entropyOfCounts(cy, n)
	if hx == 0 || hy == 0 {
		return 0
	}
	keys := make([][2]string, 0, len(joint))
	for k := range joint {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i][0] != keys[j][0] {
			return keys[i][0] < keys[j][0]
		}
		return keys[i][1] < keys[j][1]
	})
	mi := 0.0
	fn := float64(n)
	for _, k := range keys {
		pxy := float64(joint[k]) / fn
		px := float64(cx[k[0]]) / fn
		py := float64(cy[k[1]]) / fn
		if pxy > 0 && px > 0 && py > 0 {
			mi += pxy * math.Log2(pxy/(px*py))
		}
	}
	denom := max(hx, hy)
	return mi / denom
}

// AXResult holds the four independent AX diagnostics for one already
// -extracted token/glyph output stream (an operator's Render output, or a
// null's).
type AXResult struct {
	AX3StreamEntropy   float64 `json:"ax3_stream_entropy"`
	AX4TypeTokenRatio  float64 `json:"ax4_type_token_ratio"`
	AX5PeriodicNMIMax  float64 `json:"ax5_periodic_nmi_max"`
	AX5BestPeriod      int     `json:"ax5_best_period"`
	AX6LinePersistence float64 `json:"ax6_line_persistence"`
	N                  int     `json:"n"`
}

// ComputeAX runs the AX3-AX6 battery over one already-extracted output
// stream (the atoms an operator or null selected, rendered back to
// per-line groups by Render).
func ComputeAX(groups [][]string) AXResult {
	var flat []string
	firstOfLine := make([]string, 0, len(groups))
	for _, g := range groups {
		if len(g) == 0 {
			continue
		}
		flat = append(flat, g...)
		firstOfLine = append(firstOfLine, g[0])
	}

	ax3 := ShannonEntropyStrings(flat)

	types := map[string]bool{}
	for _, x := range flat {
		types[x] = true
	}
	ax4 := 0.0
	if len(flat) > 0 {
		ax4 = float64(len(types)) / float64(len(flat))
	}

	ax5Best, ax5Period := 0.0, 0
	for _, k := range []int{2, 3, 5, 7} {
		phase := make([]string, len(flat))
		for i := range flat {
			phase[i] = string(rune('0' + i%k))
		}
		nmi := normalizedMI(flat, phase)
		if nmi > ax5Best {
			ax5Best = nmi
			ax5Period = k
		}
	}

	matches, total := 0, 0
	for i := 0; i+1 < len(firstOfLine); i++ {
		total++
		if firstOfLine[i] == firstOfLine[i+1] {
			matches++
		}
	}
	observedRate := safeRatio(float64(matches), float64(total))
	expectedRate := 0.0
	if len(types) > 0 {
		expectedRate = 1.0 / float64(len(types))
	}
	ax6 := 0.0
	if expectedRate > 0 {
		ax6 = observedRate / expectedRate
	}

	return AXResult{
		AX3StreamEntropy:   ax3,
		AX4TypeTokenRatio:  ax4,
		AX5PeriodicNMIMax:  ax5Best,
		AX5BestPeriod:      ax5Period,
		AX6LinePersistence: ax6,
		N:                  len(flat),
	}
}
