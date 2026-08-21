package pairdecomposition

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func WriteAll(c Config, out Output, families []FamilyResult) error {
	if err := os.MkdirAll(c.OutputDir, 0o755); err != nil {
		return err
	}
	plots := filepath.Join(c.OutputDir, "plots")
	if !c.SkipSVG {
		if err := os.MkdirAll(plots, 0o755); err != nil {
			return err
		}
	}
	if err := writeYAML(filepath.Join(c.OutputDir, "pair_decomposition.yaml"), out); err != nil {
		return err
	}
	if err := writeYAML(filepath.Join(c.OutputDir, "family_decomposition.yaml"), map[string]any{"methodology": "all matrices use every pair among family members; graph edges remain the input structural-distant edge list", "families": families}); err != nil {
		return err
	}
	if err := writePairTSV(filepath.Join(c.OutputDir, "pair_decomposition_top.tsv"), out.Pairs); err != nil {
		return err
	}
	if err := writeControlsTSV(filepath.Join(c.OutputDir, "pair_controls.tsv"), out.Controls); err != nil {
		return err
	}
	if err := writeReport(filepath.Join(c.OutputDir, "structural_pair_report.md"), out, families); err != nil {
		return err
	}
	if !c.SkipSVG {
		for _, p := range out.Pairs {
			if err := writePairSVG(filepath.Join(plots, "pair_"+safe(p.TokenA)+"_"+safe(p.TokenB)+".svg"), p); err != nil {
				return err
			}
		}
		for _, f := range families {
			if err := writeFamilySVG(filepath.Join(plots, fmt.Sprintf("family_%d.svg", f.ID)), f); err != nil {
				return err
			}
		}
	}
	return nil
}
func writeYAML(path string, v any) error {
	b, e := yaml.Marshal(v)
	if e != nil {
		return e
	}
	return os.WriteFile(path, b, 0o644)
}
func ff(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }
func writePairTSV(path string, p []PairResult) error {
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	fmt.Fprintln(w, "token_a\ttoken_b\tcount_a\tcount_b\tstructural_similarity\treliability\tnormalized_graphemic_distance\tposition_similarity\tleft_similarity\tright_similarity\tpredecessor_jaccard\tsuccessor_jaccard\tpredecessor_js_similarity\tsuccessor_js_similarity\tleft_entropy_a\tleft_entropy_b\tright_entropy_a\tright_entropy_b\tshared_context_strength\tdifferential_context_strength\tpositional_agreement\tentropy_agreement")
	for _, x := range p {
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", x.TokenA, x.TokenB, x.CountA, x.CountB, ff(x.StructuralSimilarity), ff(x.Reliability), ff(x.GraphemicDistance), ff(x.PositionSimilarity), ff(x.LeftSimilarity), ff(x.RightSimilarity), ff(x.Left.Jaccard), ff(x.Right.Jaccard), ff(x.Left.JensenShannonSimilarity), ff(x.Right.JensenShannonSimilarity), ff(x.Left.EntropyA), ff(x.Left.EntropyB), ff(x.Right.EntropyA), ff(x.Right.EntropyB), ff(x.SharedContextStrength), ff(x.DifferentialContextStrength), ff(x.PositionalAgreement), ff(x.EntropyAgreement))
	}
	return nil
}
func writeControlsTSV(path string, c []Control) error {
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	fmt.Fprintln(w, "target_a\ttarget_b\tcontrol_rank\tcontrol_a\tcontrol_b\tmatch_cost\tcount_a\tcount_b\tstructural_similarity\treliability\tnormalized_graphemic_distance\tposition_similarity\tleft_similarity\tright_similarity\tpredecessor_js_similarity\tsuccessor_js_similarity")
	for _, x := range c {
		p := x.Decomposition
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\t%d\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", x.TargetA, x.TargetB, x.Rank, p.TokenA, p.TokenB, ff(x.MatchCost), p.CountA, p.CountB, ff(p.StructuralSimilarity), ff(p.Reliability), ff(p.GraphemicDistance), ff(p.PositionSimilarity), ff(p.LeftSimilarity), ff(p.RightSimilarity), ff(p.Left.JensenShannonSimilarity), ff(p.Right.JensenShannonSimilarity))
	}
	return nil
}

func writeReport(path string, out Output, families []FamilyResult) error {
	var b strings.Builder
	b.WriteString("# Structural pair decomposition\n\nStructural similarity is reproduced unchanged from the existing pair dataset. All statements below are formal corpus descriptions; no token meaning is inferred. Context similarities and differences use full distributions, while tables are display-limited. Entropy uses natural logarithms and effective vocabulary is `exp(entropy)`.\n\n")
	for _, p := range out.Pairs {
		fmt.Fprintf(&b, "## `%s` / `%s`\n\nStructural similarity: %.4f; reliability: %.4f; normalized graphemic distance: %.4f; counts: %d/%d.\n\n| Component | Similarity | Reliability |\n|---|---:|---:|\n| Position | %.4f | %.4f |\n| Left context | %.4f | %.4f |\n| Right context | %.4f | %.4f |\n\n", p.TokenA, p.TokenB, p.StructuralSimilarity, p.Reliability, p.GraphemicDistance, p.CountA, p.CountB, p.PositionSimilarity, p.PositionReliability, p.LeftSimilarity, p.LeftReliability, p.RightSimilarity, p.RightReliability)
		for _, s := range p.Explanation {
			fmt.Fprintf(&b, "- %s\n", s)
		}
		fmt.Fprintf(&b, "\nPosition summaries (A/B): line-start %.4f/%.4f, line-end %.4f/%.4f, mean %.3f/%.3f, median %.3f/%.3f. Position JS similarity: %.4f.\n\n", p.PositionA.LineStartProbability, p.PositionB.LineStartProbability, p.PositionA.LineEndProbability, p.PositionB.LineEndProbability, p.PositionA.Mean, p.PositionB.Mean, p.PositionA.Median, p.PositionB.Median, p.PositionJSSimilarity)
		writeContextTable(&b, "Common predecessors", p.Left.Common)
		writeContextTable(&b, "Largest predecessor differences", p.Left.Differential)
		writeContextTable(&b, "Common successors", p.Right.Common)
		writeContextTable(&b, "Largest successor differences", p.Right.Differential)
		fmt.Fprintf(&b, "Context diagnostics: predecessor Jaccard %.4f, JS %.4f, entropy A/B %.3f/%.3f, effective vocabulary A/B %.2f/%.2f; successor Jaccard %.4f, JS %.4f, entropy A/B %.3f/%.3f, effective vocabulary A/B %.2f/%.2f.\n\n", p.Left.Jaccard, p.Left.JensenShannonSimilarity, p.Left.EntropyA, p.Left.EntropyB, p.Left.EffectiveVocabularyA, p.Left.EffectiveVocabularyB, p.Right.Jaccard, p.Right.JensenShannonSimilarity, p.Right.EntropyA, p.Right.EntropyB, p.Right.EffectiveVocabularyA, p.Right.EffectiveVocabularyB)
		if len(p.Left.SharedRare)+len(p.Right.SharedRare) > 0 {
			fmt.Fprintf(&b, "Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `%s`; right `%s`.\n\n", strings.Join(contextTokens(p.Left.SharedRare), "`, `"), strings.Join(contextTokens(p.Right.SharedRare), "`, `"))
		}
		if len(p.Left.SharedAbsent)+len(p.Right.SharedAbsent) > 0 {
			fmt.Fprintf(&b, "Shared unobserved high-frequency contexts (descriptive absence only): left `%s`; right `%s`.\n\n", strings.Join(p.Left.SharedAbsent, "`, `"), strings.Join(p.Right.SharedAbsent, "`, `"))
		}
	}
	b.WriteString("## Negative controls\n\nControls match unordered log-counts, normalized graphemic distance, and reliability, while favoring structural similarity near the full-corpus median. They are decomposed with exactly the target metrics.\n\n| Target | Control | Structural | Reliability | Distance | Match cost |\n|---|---|---:|---:|---:|---:|\n")
	for _, c := range out.Controls {
		p := c.Decomposition
		fmt.Fprintf(&b, "| %s/%s | %s/%s | %.4f | %.4f | %.4f | %.4f |\n", c.TargetA, c.TargetB, p.TokenA, p.TokenB, p.StructuralSimilarity, p.Reliability, p.GraphemicDistance, c.MatchCost)
	}
	b.WriteString("\n## Family decomposition\n\nA family is a connected component; only listed edges define direct structural-distant links. Complete matrices, including non-edge pairs, are in `family_decomposition.yaml`.\n\n")
	for _, f := range families {
		fmt.Fprintf(&b, "### Family %d\n\nTokens: `%s`. Structural medoid: `%s`. Peripheral token(s): `%s`.\n\nEdges:\n\n", f.ID, strings.Join(f.Tokens, "`, `"), f.Medoid, strings.Join(f.Peripheral, "`, `"))
		for _, e := range f.Edges {
			fmt.Fprintf(&b, "- `%s` / `%s`: similarity %.4f, reliability %.4f, distance %.4f\n", e.TokenA, e.TokenB, e.StructuralSimilarity, e.Reliability, e.GraphemicDistance)
		}
		b.WriteString("\n")
	}
	b.WriteString("## Limits\n\nObserved absence is not proof of a prohibition. Context observations at line boundaries have no neighbor and therefore context totals can be below token counts. Pair rows are statistically dependent because tokens recur across pairs. Control matching is descriptive and does not make pairs independent.\n")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
func contextTokens(rows []ContextRow) []string {
	out := make([]string, len(rows))
	for i, x := range rows {
		out[i] = x.Token
	}
	return out
}
func writeContextTable(b *strings.Builder, title string, rows []ContextRow) {
	fmt.Fprintf(b, "### %s\n\n| Token | P(A) | P(B) | A−B |\n|---|---:|---:|---:|\n", title)
	for _, x := range rows {
		fmt.Fprintf(b, "| %s | %.4f | %.4f | %+.4f |\n", x.Token, x.ProbabilityA, x.ProbabilityB, x.Difference)
	}
	if len(rows) == 0 {
		b.WriteString("| — | 0 | 0 | 0 |\n")
	}
	b.WriteString("\n")
}

func safe(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			fmt.Fprintf(&b, "_%x", r)
		}
	}
	if b.Len() == 0 {
		return "token"
	}
	return b.String()
}
func xe(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&apos;").Replace(s)
}
func writePairSVG(path string, p PairResult) error {
	const w = 900
	var b strings.Builder
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="900" height="540" viewBox="0 0 900 540"><rect width="100%" height="100%" fill="white"/><style>text{font:12px sans-serif}.title{font:bold 16px sans-serif}.a{fill:#377eb8}.bb{fill:#e41a1c}.grid{stroke:#ddd}</style>`)
	fmt.Fprintf(&b, "<text class=\"title\" x=\"20\" y=\"25\">%s / %s — structural %.3f</text>", xe(p.TokenA), xe(p.TokenB), p.StructuralSimilarity)
	drawBars := func(y int, title string, rows []ContextRow) {
		fmt.Fprintf(&b, "<text x=\"20\" y=\"%d\">%s</text>", y, xe(title))
		n := len(rows)
		if n > 8 {
			n = 8
		}
		for i := 0; i < n; i++ {
			r := rows[i]
			yy := y + 22 + i*18
			fmt.Fprintf(&b, "<text x=\"20\" y=\"%d\">%s</text><rect class=\"a\" x=\"130\" y=\"%d\" width=\"%.1f\" height=\"6\"/><rect class=\"bb\" x=\"130\" y=\"%d\" width=\"%.1f\" height=\"6\"/>", yy, xe(r.Token), yy-10, r.ProbabilityA*500, yy-3, r.ProbabilityB*500)
		}
	}
	fmt.Fprintf(&b, "<text x=\"20\" y=\"55\">POSITION: existing %.3f; JS %.3f; mean %.2f / %.2f</text>", p.PositionSimilarity, p.PositionJSSimilarity, p.PositionA.Mean, p.PositionB.Mean)
	drawBars(90, "LEFT CONTEXT (largest differences)", p.Left.Differential)
	drawBars(285, "RIGHT CONTEXT (largest differences)", p.Right.Differential)
	b.WriteString(`<rect class="a" x="690" y="40" width="12" height="8"/><text x="707" y="48">A</text><rect class="bb" x="745" y="40" width="12" height="8"/><text x="762" y="48">B</text></svg>`)
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
func writeFamilySVG(path string, f FamilyResult) error {
	n := len(f.Tokens)
	if n == 0 {
		return nil
	}
	const w, h = 800, 700
	cx, cy, r := 400., 350., 260.
	xs, ys := make([]float64, n), make([]float64, n)
	for i := range f.Tokens {
		angle := 2*3.141592653589793*float64(i)/float64(n) - 3.141592653589793/2
		xs[i] = cx + r*mathCos(angle)
		ys[i] = cy + r*mathSin(angle)
	}
	idx := map[string]int{}
	for i, t := range f.Tokens {
		idx[t] = i
	}
	var b strings.Builder
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="800" height="700" viewBox="0 0 800 700"><rect width="100%" height="100%" fill="white"/><style>text{font:13px sans-serif}.edge{stroke:#4878a8}.node{fill:#fff;stroke:#222}.medoid{fill:#ffe08a}</style>`)
	fmt.Fprintf(&b, "<text x=\"20\" y=\"25\">Family %d — layout is structural-edge topology only</text>", f.ID)
	for _, e := range f.Edges {
		i, j := idx[e.TokenA], idx[e.TokenB]
		fmt.Fprintf(&b, "<line class=\"edge\" x1=\"%.1f\" y1=\"%.1f\" x2=\"%.1f\" y2=\"%.1f\" stroke-width=\"%.2f\"/><text x=\"%.1f\" y=\"%.1f\">%.3f</text>", xs[i], ys[i], xs[j], ys[j], 1+7*e.StructuralSimilarity, (xs[i]+xs[j])/2, (ys[i]+ys[j])/2, e.StructuralSimilarity)
	}
	for i, t := range f.Tokens {
		class := "node"
		if t == f.Medoid {
			class = "node medoid"
		}
		fmt.Fprintf(&b, "<circle class=\"%s\" cx=\"%.1f\" cy=\"%.1f\" r=\"24\"/><text x=\"%.1f\" y=\"%.1f\" text-anchor=\"middle\">%s</text>", class, xs[i], ys[i], xs[i], ys[i]+4, xe(t))
	}
	b.WriteString("</svg>")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// Small wrappers keep SVG generation dependency-free.
func mathSin(x float64) float64 { // minimax is unnecessary; use standard-library indirection below
	return sin(x)
}
func mathCos(x float64) float64 { return cos(x) }
