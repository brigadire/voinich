package higherorderseq

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func svgWrap(title, body string) []byte {
	return []byte(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="900" height="320" viewBox="0 0 900 320"><rect width="100%%" height="100%%" fill="white"/><style>text{font:12px sans-serif}.t{font:bold 15px sans-serif}</style><text class="t" x="20" y="25">%s</text>%s</svg>`, title, body))
}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func slug(sequence string) string {
	return strings.Trim(slugRe.ReplaceAllString(strings.ToLower(sequence), "_"), "_")
}

// writePlots implements task22 Part R (sections 82-83): every plot kind is
// rendered once per candidate rather than combined into one unreadable
// chart, so a reader can inspect each frozen sequence's evidence on its own.
func writePlots(dir string, results []*CandidateResult) error {
	for _, r := range results {
		s := slug(r.Candidate.Sequence)
		if err := plotEnrichmentByBlock(filepath.Join(dir, "plots", "conditional_enrichment_by_block_"+s+".svg"), r); err != nil {
			return err
		}
		if err := plotPCGivenBVsAB(filepath.Join(dir, "plots", "p_c_given_b_vs_ab_"+s+".svg"), r); err != nil {
			return err
		}
		if err := plotLogLoss(filepath.Join(dir, "plots", "first_vs_second_order_logloss_"+s+".svg"), r); err != nil {
			return err
		}
		if err := plotContinuationEntropy(filepath.Join(dir, "plots", "continuation_entropy_"+s+".svg"), r); err != nil {
			return err
		}
		if err := plotContextRank(filepath.Join(dir, "plots", "conditional_context_rank_"+s+".svg"), r); err != nil {
			return err
		}
		if err := plotBlockPosition(filepath.Join(dir, "plots", "sequence_block_position_"+s+".svg"), r); err != nil {
			return err
		}
	}
	return nil
}

func axis() string {
	return `<line x1="60" y1="260" x2="860" y2="260" stroke="#333"/><line x1="60" y1="40" x2="60" y2="260" stroke="#333"/>`
}

func plotEnrichmentByBlock(path string, r *CandidateResult) error {
	eligible := primaryEligible(r.ConditionalRows)
	sort.Slice(eligible, func(i, j int) bool { return eligible[i].Block < eligible[j].Block })
	var b strings.Builder
	b.WriteString(axis())
	maxE := 1.0
	for _, e := range eligible {
		if e.Enrichment > maxE {
			maxE = e.Enrichment
		}
	}
	n := len(eligible)
	for idx, e := range eligible {
		x := 70 + float64(idx)*(780/float64(max1(n)))
		h := 200 * e.Enrichment / (maxE * 1.1)
		color := "#2563eb"
		if e.Enrichment < 1 {
			color = "#dc2626"
		}
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`, x, 260-h, 780/float64(max1(n))*0.8, h, color)
	}
	fmt.Fprintf(&b, `<line x1="60" y1="%.1f" x2="860" y2="%.1f" stroke="#999" stroke-dasharray="4"/><text x="865" y="%.1f">1.0</text>`, 260-200/(maxE*1.1), 260-200/(maxE*1.1), 260-200/(maxE*1.1))
	return os.WriteFile(path, svgWrap("Conditional enrichment P(C|A,B)/P(C|B) by eligible block - "+r.Candidate.Sequence, b.String()), 0644)
}

func plotPCGivenBVsAB(path string, r *CandidateResult) error {
	eligible := primaryEligible(r.ConditionalRows)
	sort.Slice(eligible, func(i, j int) bool { return eligible[i].Block < eligible[j].Block })
	var b strings.Builder
	b.WriteString(axis())
	n := len(eligible)
	w := 780 / float64(max1(n))
	for idx, e := range eligible {
		x := 70 + float64(idx)*w
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="#94a3b8"/>`, x, 260-200*e.PCGivenB, w*0.35, 200*e.PCGivenB)
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="#2563eb"/>`, x+w*0.4, 260-200*e.PCGivenAB, w*0.35, 200*e.PCGivenAB)
	}
	b.WriteString(`<text x="60" y="300">gray = P(C|B), blue = P(C|A,B)</text>`)
	return os.WriteFile(path, svgWrap("P(C|B) vs P(C|A,B) by eligible block - "+r.Candidate.Sequence, b.String()), 0644)
}

func plotLogLoss(path string, r *CandidateResult) error {
	var b strings.Builder
	b.WriteString(axis())
	l := r.LOBO
	labels := []string{"M2 better", "M1 better", "ties"}
	vals := []int{l.M2BetterBlocks, l.M1BetterBlocks, l.Ties}
	maxV := 1
	for _, v := range vals {
		if v > maxV {
			maxV = v
		}
	}
	for idx, v := range vals {
		x := 100 + float64(idx)*250
		h := 200 * float64(v) / float64(maxV)
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="120" height="%.1f" fill="#059669"/><text x="%.1f" y="275">%s (%d)</text>`, x, 260-h, h, x, labels[idx], v)
	}
	fmt.Fprintf(&b, `<text x="60" y="300">mean delta log loss = %s bits (positive favors M2)</text>`, f(l.MeanDeltaLogLoss))
	return os.WriteFile(path, svgWrap("First- vs second-order held-out log loss - "+r.Candidate.Sequence, b.String()), 0644)
}

func plotContinuationEntropy(path string, r *CandidateResult) error {
	var b strings.Builder
	b.WriteString(axis())
	e := r.ContinuationEnt
	maxH := e.HGivenB
	if e.HGivenAB > maxH {
		maxH = e.HGivenAB
	}
	if maxH == 0 {
		maxH = 1
	}
	fmt.Fprintf(&b, `<rect x="200" y="%.1f" width="120" height="%.1f" fill="#94a3b8"/><text x="200" y="275">H(X|B)=%.3f</text>`, 260-200*e.HGivenB/maxH, 200*e.HGivenB/maxH, e.HGivenB)
	fmt.Fprintf(&b, `<rect x="450" y="%.1f" width="120" height="%.1f" fill="#2563eb"/><text x="450" y="275">H(X|A,B)=%.3f</text>`, 260-200*e.HGivenAB/maxH, 200*e.HGivenAB/maxH, e.HGivenAB)
	fmt.Fprintf(&b, `<text x="60" y="300">entropy reduction = %.4f bits, JS divergence = %.4f, TVD = %.4f</text>`, e.EntropyReduction, e.JSDivergence, e.TotalVariation)
	return os.WriteFile(path, svgWrap("Continuation entropy H(X|B) vs H(X|A,B) - "+r.Candidate.Sequence, b.String()), 0644)
}

func plotContextRank(path string, r *CandidateResult) error {
	var b strings.Builder
	b.WriteString(axis())
	cr := r.ContextRank
	if cr.NumAlternatives == 0 {
		b.WriteString(`<text x="60" y="150">no sufficiently frequent alternative left contexts</text>`)
		return os.WriteFile(path, svgWrap("Conditional context rank - "+r.Candidate.Sequence, b.String()), 0644)
	}
	maxP := cr.MaxAltP
	if maxP <= 0 {
		maxP = 1
	}
	fmt.Fprintf(&b, `<rect x="150" y="%.1f" width="120" height="%.1f" fill="#94a3b8"/><text x="150" y="275">baseline</text>`, 260-200*cr.BaselineP/maxP, 200*cr.BaselineP/maxP)
	fmt.Fprintf(&b, `<rect x="350" y="%.1f" width="120" height="%.1f" fill="#2563eb"/><text x="350" y="275">frozen (rank %d/%d)</text>`, 260-200*cr.FrozenP/maxP, 200*cr.FrozenP/maxP, cr.Rank, cr.NumAlternatives)
	fmt.Fprintf(&b, `<rect x="550" y="%.1f" width="120" height="%.1f" fill="#dc2626"/><text x="550" y="275">max alt</text>`, 260-200*cr.MaxAltP/maxP, 200*cr.MaxAltP/maxP)
	return os.WriteFile(path, svgWrap("Frozen context rank among alternative left contexts - "+r.Candidate.Sequence, b.String()), 0644)
}

func plotBlockPosition(path string, r *CandidateResult) error {
	var b strings.Builder
	b.WriteString(axis())
	buckets := []string{"[0,0.1)", "[0.1,0.2)", "[0.2,0.3)", "[0.3,0.4)", "[0.4,0.5)", "[0.5,0.6)", "[0.6,0.7)", "[0.7,0.8)", "[0.8,0.9)", "[0.9,1.0]"}
	byBucket := map[string]PositionRow{}
	for _, pr := range r.Position {
		if pr.Metric == "block_position_bin" {
			byBucket[pr.Bucket] = pr
		}
	}
	w := 780 / float64(len(buckets))
	for idx, bucket := range buckets {
		pr := byBucket[bucket]
		x := 70 + float64(idx)*w
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="#94a3b8"/>`, x, 260-200*pr.ABFraction, w*0.35, 200*pr.ABFraction)
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="#2563eb"/>`, x+w*0.4, 260-200*pr.ABCFraction, w*0.35, 200*pr.ABCFraction)
	}
	b.WriteString(`<text x="60" y="300">gray = AB baseline fraction, blue = ABC fraction, by normalized block position decile</text>`)
	return os.WriteFile(path, svgWrap("ABC vs AB normalized block position - "+r.Candidate.Sequence, b.String()), 0644)
}

func max1(n int) int {
	if n < 1 {
		return 1
	}
	return n
}
