package positionalcontinuation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func svgWrap(title, body string) []byte {
	return []byte(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="900" height="320" viewBox="0 0 900 320"><rect width="100%%" height="100%%" fill="white"/><style>text{font:12px sans-serif}.t{font:bold 15px sans-serif}</style><text class="t" x="20" y="25">%s</text>%s</svg>`, title, body))
}

func axis() string {
	return `<line x1="60" y1="260" x2="860" y2="260" stroke="#333"/><line x1="60" y1="40" x2="60" y2="260" stroke="#333"/>`
}

func max1(n int) int {
	if n < 1 {
		return 1
	}
	return n
}

// writePlots implements task23 Part T (section 99): the eight required
// plots, one purpose each rather than combined into one unreadable chart.
func writePlots(dir string, state *RunState) error {
	if err := os.MkdirAll(filepath.Join(dir, "plots"), 0755); err != nil {
		return err
	}
	plots := []struct {
		name string
		fn   func(string) error
	}{
		{"sai_in_continuation_by_line_position.svg", func(p string) error { return plotContinuationByStratum(p, state, "line_position", lineCategories) }},
		{"sai_in_continuation_by_block_position.svg", func(p string) error { return plotContinuationByStratum(p, state, "block_position_coarse", blockCoarseCategories) }},
		{"chey_probability_by_position.svg", func(p string) error { return plotCheyProbability(p, state) }},
		{"sai_in_vs_aiin_chey_probability.svg", func(p string) error { return plotSAiinVsAiin(p, state) }},
		{"continuation_entropy_by_position.svg", func(p string) error { return plotEntropyByPosition(p, state) }},
		{"model_lobo_comparison.svg", func(p string) error { return plotModelLOBO(p, state) }},
		{"sai_in_block_positions.svg", func(p string) error { return plotOccurrenceCounts(p, state, "block_position_coarse", blockCoarseCategories) }},
		{"sai_in_line_positions.svg", func(p string) error { return plotOccurrenceCounts(p, state, "line_position", lineCategories) }},
	}
	for _, pl := range plots {
		if err := pl.fn(filepath.Join(dir, "plots", pl.name)); err != nil {
			return err
		}
	}
	return nil
}

func barChart(title, note string, labels []string, values []float64, color string) []byte {
	var b strings.Builder
	b.WriteString(axis())
	maxV := 0.0
	for i, v := range values {
		if i == 0 || v > maxV {
			maxV = v
		}
	}
	if maxV <= 0 {
		maxV = 1
	}
	w := 780 / float64(max1(len(labels)))
	for idx, v := range values {
		x := 70 + float64(idx)*w
		h := 200 * v / (maxV * 1.1)
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`, x, 260-h, w*0.8, h, color)
		fmt.Fprintf(&b, `<text x="%.1f" y="275" font-size="9">%s</text>`, x, labels[idx])
	}
	if note != "" {
		b.WriteString(`<text x="60" y="300">` + note + `</text>`)
	}
	return svgWrap(title, b.String())
}

func plotContinuationByStratum(path string, state *RunState, stratumType string, cats []string) error {
	var vals []float64
	for _, c := range cats {
		v := 0.0
		for _, s := range state.DistSummaryRows {
			if s.StratumType == stratumType && s.Stratum == c {
				v = s.TopContinuationProb
			}
		}
		vals = append(vals, v)
	}
	return os.WriteFile(path, barChart("Top continuation probability by "+stratumType, "bar height = P(top continuation | s aiin, position)", cats, vals, "#2563eb"), 0644)
}

func plotCheyProbability(path string, state *RunState) error {
	var labels []string
	var vals []float64
	for _, ce := range state.CheyEffect {
		if ce.PositionVariable != "line_position" {
			continue
		}
		labels = append(labels, ce.Stratum)
		vals = append(vals, ce.PCheyGivenPosition)
	}
	return os.WriteFile(path, barChart("P(chey | s,aiin,position) by line position", "", labels, vals, "#dc2626"), 0644)
}

func plotSAiinVsAiin(path string, state *RunState) error {
	var labels []string
	var sVals, aVals []float64
	for _, ac := range state.AiinControl {
		if ac.PositionVariable != "line_position" {
			continue
		}
		labels = append(labels, ac.Stratum)
		sVals = append(sVals, ac.PCheyGivenSAiinPosition)
		aVals = append(aVals, ac.PCheyGivenAiinPosition)
	}
	var b strings.Builder
	b.WriteString(axis())
	maxV := 0.01
	for i := range sVals {
		if sVals[i] > maxV {
			maxV = sVals[i]
		}
		if aVals[i] > maxV {
			maxV = aVals[i]
		}
	}
	w := 780 / float64(max1(len(labels)))
	for idx := range labels {
		x := 70 + float64(idx)*w
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="#94a3b8"/>`, x, 260-200*aVals[idx]/maxV, w*0.35, 200*aVals[idx]/maxV)
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="#2563eb"/>`, x+w*0.4, 260-200*sVals[idx]/maxV, w*0.35, 200*sVals[idx]/maxV)
		fmt.Fprintf(&b, `<text x="%.1f" y="275" font-size="9">%s</text>`, x, labels[idx])
	}
	b.WriteString(`<text x="60" y="300">gray = P(chey|aiin,position), blue = P(chey|s,aiin,position)</text>`)
	return os.WriteFile(path, svgWrap("s aiin vs aiin: chey probability by position", b.String()), 0644)
}

func plotEntropyByPosition(path string, state *RunState) error {
	var labels []string
	var vals []float64
	for _, pe := range state.PositionalEntropy {
		if pe.PositionVariable != "line_position" {
			continue
		}
		labels = append(labels, pe.Stratum)
		vals = append(vals, pe.EntropyBits)
	}
	return os.WriteFile(path, barChart("H(X | s,aiin,position) by line position", "lower = more constrained continuation set", labels, vals, "#059669"), 0644)
}

func plotModelLOBO(path string, state *RunState) error {
	labels := []string{"M2>M1", "M1>M2", "M3>M2", "M2>M3"}
	m := state.ModelLOBO
	vals := []float64{float64(m.BlocksM2BetterM1), float64(m.BlocksM1BetterM2), float64(m.BlocksM3BetterM2), float64(m.BlocksM2BetterM3)}
	note := fmt.Sprintf("mean delta_21=%.3f bits, mean delta_32=%.3f bits (tested blocks=%d)", m.MeanDelta21, m.MeanDelta32, m.TestedBlocks)
	return os.WriteFile(path, barChart("M1/M2/M3 held-out log-loss comparison", note, labels, vals, "#7c3aed"), 0644)
}

func plotOccurrenceCounts(path string, state *RunState, stratumType string, cats []string) error {
	var vals []float64
	for _, c := range cats {
		v := 0.0
		for _, s := range state.DistSummaryRows {
			if s.StratumType == stratumType && s.Stratum == c {
				v = float64(s.OccurrenceCount)
			}
		}
		vals = append(vals, v)
	}
	return os.WriteFile(path, barChart("s aiin occurrence count by "+stratumType, "", cats, vals, "#f59e0b"), 0644)
}
