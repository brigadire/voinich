package task82a

import (
	"fmt"
	"math"
	"sort"
)

// PilotRow is one measured point of the target-blind scale convergence
// pilot (task82a.txt sec.24-25). It never touches Voynich data: only the
// Doyle natural-language control's literal Latin23 stream, at increasing
// prefix lengths.
type PilotRow struct {
	Chunks            int     `json:"chunks"`
	Symbols           int     `json:"symbols"`
	SymbolEntropyBits float64 `json:"symbol_entropy_bits"`
	DeltaVsHalf       float64 `json:"delta_vs_half_chunks"`
	Converged         bool    `json:"converged"`
}

// PilotResult is the frozen-before-generation convergence finding used to
// pick SMALL/MEDIUM/LARGE. The convergence criterion (doubling the chunk
// count changes plug-in symbol entropy by <=0.01 bits) is fixed before any
// row is computed; it is never re-picked after seeing the numbers.
type PilotResult struct {
	Criterion    string     `json:"criterion"`
	Corpus       string     `json:"corpus"`
	Rows         []PilotRow `json:"rows"`
	ConvergedAt  int        `json:"converged_at_chunks"`
	SmallChunks  int        `json:"small_chunks"`
	MediumChunks int        `json:"medium_chunks"`
	LargeChunks  int        `json:"large_chunks"`
}

// RunScaleConvergencePilot measures plug-in character entropy of the
// literal Latin23 stream at doubling chunk-count checkpoints and reports
// the first checkpoint whose entropy is stable (<=0.01 bits) versus half
// that many chunks. SMALL/MEDIUM/LARGE are then set at a fixed 1x/4x/16x
// multiplicative ladder above the convergence point -- a generic scaling
// rule fixed before this function is ever run, not chosen from its output.
func RunScaleConvergencePilot(root string, capacity int, checkpoints []int) (PilotResult, error) {
	maxChunks := checkpoints[len(checkpoints)-1]
	corpora, err := loadSourceCorpora(root, capacity*maxChunks, 4)
	if err != nil {
		return PilotResult{}, err
	}
	doyle := corpora["Doyle"]
	res := PilotResult{
		Criterion: "plug-in symbol entropy of the literal Latin23 stream changes by <=0.01 bits when the chunk count doubles",
		Corpus:    "Doyle",
	}
	entropyAt := map[int]float64{}
	for _, n := range checkpoints {
		symbols := doyle.Letters[:n*capacity]
		h := plugInEntropy(symbols)
		entropyAt[n] = h
		row := PilotRow{Chunks: n, Symbols: len(symbols), SymbolEntropyBits: h}
		if half, ok := entropyAt[n/2]; ok && n/2 > 0 {
			row.DeltaVsHalf = math.Abs(h - half)
			row.Converged = row.DeltaVsHalf <= 0.01
		}
		res.Rows = append(res.Rows, row)
	}
	for _, row := range res.Rows {
		if row.Converged && res.ConvergedAt == 0 {
			res.ConvergedAt = row.Chunks
		}
	}
	if res.ConvergedAt == 0 {
		res.ConvergedAt = checkpoints[len(checkpoints)-1]
	}
	res.SmallChunks = res.ConvergedAt
	res.MediumChunks = res.ConvergedAt * 4
	res.LargeChunks = res.ConvergedAt * 16
	return res, nil
}

func plugInEntropy(symbols []string) float64 {
	counts := map[string]int{}
	for _, s := range symbols {
		counts[s]++
	}
	n := len(symbols)
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
		h -= p * math.Log2(p)
	}
	return h
}

func (r PilotResult) String() string {
	s := fmt.Sprintf("criterion: %s\ncorpus: %s\nconverged_at_chunks: %d\nSMALL=%d MEDIUM=%d LARGE=%d\n", r.Criterion, r.Corpus, r.ConvergedAt, r.SmallChunks, r.MediumChunks, r.LargeChunks)
	for _, row := range r.Rows {
		s += fmt.Sprintf("  chunks=%-6d symbols=%-7d entropy_bits=%.6f delta_vs_half=%.6f converged=%v\n", row.Chunks, row.Symbols, row.SymbolEntropyBits, row.DeltaVsHalf, row.Converged)
	}
	return s
}
