package graphemic

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func WriteAll(cfg Config, r Result) error {
	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		return err
	}
	if err := writePairs(filepath.Join(cfg.OutputDir, "structural_graphemic_pairs.tsv"), r.Pairs, 0); err != nil {
		return err
	}
	if err := writePairs(filepath.Join(cfg.OutputDir, "structural_distant_top.tsv"), r.Distant, cfg.TopN); err != nil {
		return err
	}
	if err := writePairs(filepath.Join(cfg.OutputDir, "structural_graphemic_close_top.tsv"), r.Close, cfg.TopN); err != nil {
		return err
	}
	closeOut := FamiliesOutput{"graphemic-structural family", map[string]float64{"min_structural_similarity": cfg.MinStructuralSimilarity, "min_reliability": cfg.MinReliability, "min_graphemic_similarity": cfg.MinCloseSimilarity}, r.CloseFamilies}
	distantOut := FamiliesOutput{"structural-distant family", map[string]float64{"min_structural_similarity": cfg.MinStructuralSimilarity, "min_reliability": cfg.MinReliability, "min_normalized_graphemic_distance": cfg.MinGraphemicDistance}, r.DistantFamilies}
	if err := writeYAML(filepath.Join(cfg.OutputDir, "graphemic_structural_families.yaml"), closeOut); err != nil {
		return err
	}
	if err := writeYAML(filepath.Join(cfg.OutputDir, "structural_distant_families.yaml"), distantOut); err != nil {
		return err
	}
	if err := writeReport(filepath.Join(cfg.OutputDir, "structural_graphemic_report.md"), cfg, r); err != nil {
		return err
	}
	return writeSVG(filepath.Join(cfg.OutputDir, "structural_vs_graphemic.svg"), r)
}

func writeYAML(path string, v any) error {
	b, e := yaml.Marshal(v)
	if e != nil {
		return e
	}
	return os.WriteFile(path, b, 0o644)
}
func f64(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }
func writePairs(path string, pairs []Pair, limit int) error {
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	fmt.Fprintln(w, "token_a\ttoken_b\tcount_a\tcount_b\tstructural_similarity\treliability\tposition_similarity\tleft_similarity\tright_similarity\tposition_reliability\tleft_reliability\tright_reliability\ttotal_evidence_weight\tdiagnostic_weighted_similarity\tgrapheme_distance\tnormalized_grapheme_distance\tgrapheme_similarity\tcommon_prefix\tcommon_suffix\tlength_difference\tdiscovery_score")
	if limit > 0 && len(pairs) > limit {
		pairs = pairs[:limit]
	}
	for _, p := range pairs {
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s\t%s\t%d\t%d\t%d\t%s\n", p.TokenA, p.TokenB, p.CountA, p.CountB, f64(p.StructuralSimilarity), f64(p.Reliability), f64(p.PositionSimilarity), f64(p.LeftSimilarity), f64(p.RightSimilarity), f64(p.PositionReliability), f64(p.LeftReliability), f64(p.RightReliability), f64(p.TotalEvidenceWeight), p.DiagnosticWeightedSimilarity, p.GraphemeDistance, f64(p.NormalizedGraphemeDistance), f64(p.GraphemeSimilarity), p.CommonPrefix, p.CommonSuffix, p.LengthDifference, f64(p.DiscoveryScore))
	}
	return nil
}

func writeReport(path string, cfg Config, r Result) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# Structural–graphemic analysis\n\nThis analysis compares two independent coordinates. The existing structural similarity was copied unchanged from `raw_similarity`; spelling was used only for the graphemic metrics.\n\n## Scope\n\n- Tokens: %d\n- Pairs: %d\n- Pearson correlation (graphemic similarity vs structural similarity): %.6f\n- Spearman correlation: %.6f\n\nPair observations share tokens and are not independent; the correlations are descriptive and ordinary pairwise p-values are intentionally not reported.\n\n", r.TokenCount, len(r.Pairs), r.Pearson, r.Spearman)
	b.WriteString("## Structural similarity by normalized grapheme-distance bin\n\n| Distance | Pairs | Mean | Median | P90 | P95 |\n|---|---:|---:|---:|---:|---:|\n")
	for _, x := range r.Bins {
		fmt.Fprintf(&b, "| %s | %d | %.4f | %.4f | %.4f | %.4f |\n", x.Range, x.PairCount, x.Mean, x.Median, x.P90, x.P95)
	}
	fmt.Fprintf(&b, "\n## Selection and percentile view\n\nConfigurable distant selection uses structural similarity ≥ %.3f, reliability ≥ %.3f, and normalized grapheme distance ≥ %.3f. It yields %d pairs. Ranking is `structural_similarity × reliability × normalized_grapheme_distance`; this score is neither a probability nor statistical significance.\n\nAs a threshold-free companion view, corpus cutoffs are structural P95 = %.4f, reliability P75 = %.4f, and grapheme-distance P75 = %.4f. Their intersection yields %d pairs, ranked by the same transparent score. These percentiles describe ranking coordinates and do not define the class.\n\n", cfg.MinStructuralSimilarity, cfg.MinReliability, cfg.MinGraphemicDistance, len(r.Distant), r.DistantPercentileCutoffs["structural_p95"], r.DistantPercentileCutoffs["reliability_p75"], r.DistantPercentileCutoffs["graphemic_distance_p75"], len(r.PercentileDistant))
	writeTopTable(&b, "Structurally close / graphically distant", r.Distant, 20)
	writeTopTable(&b, "Percentile-ranked distant view", r.PercentileDistant, 20)
	writeTopTable(&b, "Structurally close / graphically close (control)", r.Close, 20)
	b.WriteString("## Frequency control\n\n| Minimum count for both tokens | Pairs | Pearson | Spearman | Distant candidates |\n|---:|---:|---:|---:|---:|\n")
	for _, x := range r.Frequency {
		fmt.Fprintf(&b, "| %d | %d | %.5f | %.5f | %d |\n", x.MinimumCount, x.PairCount, x.Pearson, x.Spearman, x.DistantCandidates)
	}
	fmt.Fprintf(&b, "\n## Families\n\nThe same explicit edge criteria form connected components. There are %d graphemic-structural families and %d structural-distant families. Components can contain token pairs that are connected only through intermediate tokens; inspect the saved edge list before interpreting a whole component. The neutral term “family” denotes a graph component only.\n\n", len(r.CloseFamilies), len(r.DistantFamilies))
	writeFamilySummary(&b, "Graphemic-structural families", r.CloseFamilies)
	writeFamilySummary(&b, "Structural-distant families", r.DistantFamilies)
	b.WriteString("## Graphemic-distance distribution\n\nThe bin counts above are the full empirical distribution of normalized edit distance. Edit operations are performed on grapheme sequences: `@NNN;` is one grapheme, `?` is one unknown grapheme, and no signs are deleted or normalized.\n\n## Limitations\n\n- Pair rows are dependent because each token occurs in many pairs.\n- Reliability and frequency reduce, but cannot eliminate, instability of sparse profiles.\n- Levenshtein distance assigns equal cost to every insertion, deletion, and substitution and contains no palaeographic model.\n- Connected components are threshold-sensitive descriptive groups, not linguistic categories.\n- Correlation does not establish that one coordinate causes the other.\n- The analysis makes no claim about language, morphology, commands, operators, or cipher mechanisms.\n")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
func writeTopTable(b *strings.Builder, title string, p []Pair, n int) {
	fmt.Fprintf(b, "## %s\n\n| A | B | Counts | Structural | Reliability | Norm. distance | Discovery score |\n|---|---|---:|---:|---:|---:|---:|\n", title)
	if len(p) < n {
		n = len(p)
	}
	for _, x := range p[:n] {
		fmt.Fprintf(b, "| %s | %s | %d/%d | %.4f | %.4f | %.4f | %.4f |\n", x.TokenA, x.TokenB, x.CountA, x.CountB, x.StructuralSimilarity, x.Reliability, x.NormalizedGraphemeDistance, x.DiscoveryScore)
	}
	b.WriteString("\n")
}
func writeFamilySummary(b *strings.Builder, title string, f []Family) {
	fmt.Fprintf(b, "### %s\n\n", title)
	n := len(f)
	if n > 20 {
		n = 20
	}
	for _, x := range f[:n] {
		tokens := x.Tokens
		if len(tokens) > 15 {
			tokens = tokens[:15]
		}
		fmt.Fprintf(b, "- Family %d (%d tokens): %s", x.ID, len(x.Tokens), strings.Join(tokens, ", "))
		if len(x.Tokens) > len(tokens) {
			b.WriteString(", …")
		}
		b.WriteString("\n")
	}
	if n == 0 {
		b.WriteString("No components met the configured criteria.\n")
	}
	b.WriteString("\n")
}

func writeSVG(path string, r Result) error {
	const w, h = 1000, 650
	var b strings.Builder
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="1000" height="650" viewBox="0 0 1000 650"><rect width="100%" height="100%" fill="white"/><style>text{font:12px sans-serif}.axis{stroke:#333}.point{fill:#376a9f;fill-opacity:.12}.out{fill:#b21f2d;fill-opacity:.8}</style>`)
	fmt.Fprintf(&b, "<text x=\"500\" y=\"22\" text-anchor=\"middle\">Structural similarity vs normalized grapheme distance</text><line class=\"axis\" x1=\"70\" y1=\"590\" x2=\"970\" y2=\"590\"/><line class=\"axis\" x1=\"70\" y1=\"40\" x2=\"70\" y2=\"590\"/><text x=\"500\" y=\"635\" text-anchor=\"middle\">normalized grapheme distance</text><text transform=\"translate(18 330) rotate(-90)\" text-anchor=\"middle\">structural similarity</text>")
	for i := 0; i <= 10; i++ {
		x := 70 + 90*i
		y := 590 - 55*i
		fmt.Fprintf(&b, "<text x=\"%d\" y=\"607\" text-anchor=\"middle\">%.1f</text><text x=\"62\" y=\"%d\" text-anchor=\"end\">%.1f</text>", x, float64(i)/10, y+4, float64(i)/10)
	}
	for _, p := range r.Pairs {
		x := 70 + p.NormalizedGraphemeDistance*900
		y := 590 - p.StructuralSimilarity*550
		fmt.Fprintf(&b, "<circle class=\"point\" cx=\"%.1f\" cy=\"%.1f\" r=\"1.2\"/>", x, y)
	}
	outs := append([]Pair(nil), r.Distant...)
	sortPairs(outs, func(p Pair) float64 { return p.DiscoveryScore })
	if len(outs) > 10 {
		outs = outs[:10]
	}
	for i, p := range outs {
		x := 70 + p.NormalizedGraphemeDistance*900
		y := 590 - p.StructuralSimilarity*550
		label := xmlEscape(p.TokenA + "/" + p.TokenB)
		dy := -5
		if i%2 == 1 {
			dy = 14
		}
		fmt.Fprintf(&b, "<circle class=\"out\" cx=\"%.1f\" cy=\"%.1f\" r=\"3\"/><text x=\"%.1f\" y=\"%.1f\">%s</text>", x, y, x+4, y+float64(dy), label)
	}
	b.WriteString("</svg>")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func xmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&apos;")
	return r.Replace(s)
}
