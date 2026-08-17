package structuralprojection

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func RunAndWrite(c Config) error {
	if c.ProgressWriter == nil && !c.Quiet {
		c.ProgressWriter = os.Stderr
	}
	progress := newProgress(c.ProgressWriter)
	a, e := analyze(c, progress)
	if e != nil {
		return e
	}
	if e = os.MkdirAll(filepath.Join(c.OutputDir, "plots"), 0o755); e != nil {
		return e
	}
	progress.begin(7, "Writing results")
	files := []struct {
		name string
		v    any
	}{{"structural_projection_pairs.yaml", a.Out}, {"structural_projection_families.yaml", map[string]any{"methodology": "token and projected mean pairwise JS cohesion; dispersion is population SD; percentiles use 200 deterministic frequency-matched groups", "families": a.Families}}, {"projected_sequence_context.yaml", map[string]any{"methodology": "exact suffix JS plus a sequence kernel equal to the mean position-wise JS after soft projection", "pairs": a.Sequences}}}
	for _, f := range files {
		if e = writeYAML(filepath.Join(c.OutputDir, f.name), f.v); e != nil {
			return e
		}
	}
	if e = writeTop(filepath.Join(c.OutputDir, "structural_projection_top.tsv"), a.Out.Pairs); e != nil {
		return e
	}
	if e = writeControls(filepath.Join(c.OutputDir, "structural_projection_controls.tsv"), a); e != nil {
		return e
	}
	if e = writeReport(filepath.Join(c.OutputDir, "structural_projection_report.md"), a); e != nil {
		return e
	}
	for _, p := range a.Out.Pairs {
		if isMainPlot(p.TokenA, p.TokenB) {
			if e = writePairPlot(filepath.Join(c.OutputDir, "plots", "projection_"+safe(p.TokenA)+"_"+safe(p.TokenB)+".svg"), p); e != nil {
				return e
			}
		}
	}
	for _, f := range a.Families {
		if f.ID == 2 {
			if e = writeFamilyPlot(filepath.Join(c.OutputDir, "plots", "projection_family_2.svg"), f); e != nil {
				return e
			}
		}
	}
	progress.update(1, 1, "Writing results")
	// A checkpoint is operational recovery state, not a scientific output.
	// Remove it only after every final output has been written successfully.
	if c.CheckpointPath != "" && c.CheckpointPath != "-" {
		if e = os.Remove(c.CheckpointPath); e != nil && !os.IsNotExist(e) {
			return e
		}
	}
	fmt.Printf("Structural projection analyzed %d pairs; results written to %s\n", len(a.Out.Pairs), c.OutputDir)
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
func writeTop(path string, p []PairResult) error {
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	fmt.Fprintln(w, "rank\ttoken_a\ttoken_b\tmean_token_js_1_5\tmean_projected_js_1_5\tmean_ablated_projected_js_1_5\tprojection_gain_1_5\tprojection_gain_percentile\tgain_6_10\tgain_11_20")
	for i, x := range p {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", i+1, x.TokenA, x.TokenB, ff(x.Summary.MeanToken1To5), ff(x.Summary.MeanFull1To5), ff(x.Summary.MeanAblated1To5), ff(x.Summary.Gain1To5), ff(x.Summary.Control.RandomPercentile), ff(x.Summary.Gain6To10), ff(x.Summary.Gain11To20))
	}
	return nil
}
func writeControls(path string, a analysis) error {
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	fmt.Fprintln(w, "record_type\ttoken_a\ttoken_b\tcontrol\tparameter\tobserved_gain\tcontrol_mean\tcontrol_p95\tpercentile\tmean_gain_full\tmean_gain_ablated")
	for _, x := range a.Controls {
		fmt.Fprintf(w, "pair\t%s\t%s\t%s\t\t%s\t%s\t%s\t%s\t\t\n", x.TokenA, x.TokenB, x.Kind, ff(x.Observed), ff(x.Mean), ff(x.P95), ff(x.Percentile))
	}
	for _, x := range a.Out.Sweeps {
		fmt.Fprintf(w, "sweep\t\t\t%s\t%s\t\t\t\t\t%s\t%s\n", x.Method, ff(x.Parameter), ff(x.MeanGainFull), ff(x.MeanGainAblated))
	}
	for _, x := range a.Out.Shuffles {
		fmt.Fprintf(w, "shuffle\t\t\t%s\t\t%s\t\t\t\t%s\t\n", x.Mode, ff(x.MeanGain), ff(x.MeanProjectedJS))
	}
	return nil
}

func writeReport(path string, a analysis) error {
	var b strings.Builder
	b.WriteString("# Soft structural projection analysis\n\nThis is a formal structural-neighbourhood experiment. It makes no semantic, grammatical, morphological, or syntactic claims. Token-level results remain the reference and are never replaced by projected metrics.\n\n## Projection and anti-circularity\n\nFor an observed token `X`, `W(X,X)=1`. Each eligible neighbour receives `raw structural similarity × evidence reliability`; a row is normalized to sum to one. Thus `P_projected(Y|A,d) = Σ_X P(X|A,d) W_normalized(X,Y)`. The future-context ablation reconstructs weights from position and left-context components only; the past-context ablation analogously excludes left context.\n\nFamily projection is reported only as a coarse control and keeps all non-family tokens as singletons. No graphemic quantity is loaded or used.\n\n## Main ranking\n\n| Pair | token JS 1–5 | full projected | ablated | full gain | random percentile | random P95 | smoothing P95 |\n|---|---:|---:|---:|---:|---:|---:|---:|\n")
	for _, x := range a.Out.Pairs {
		c := x.Summary.Control
		fmt.Fprintf(&b, "| `%s` / `%s` | %.4f | %.4f | %.4f | %+.4f | P%.1f | %+.4f | %+.4f |\n", x.TokenA, x.TokenB, x.Summary.MeanToken1To5, x.Summary.MeanFull1To5, x.Summary.MeanAblated1To5, x.Summary.Gain1To5, c.RandomPercentile, c.RandomP95, c.SmoothingP95)
	}
	b.WriteString("\nAblated random-space and smoothing percentiles are stored alongside every pair and in the controls TSV.\n\n## Prespecified sensitivity sweep\n\nNo parameter is selected from these results.\n\n| Method | Parameter | Full mean gain | Ablated mean gain |\n|---|---:|---:|---:|\n")
	for _, x := range a.Out.Sweeps {
		fmt.Fprintf(&b, "| %s | %.2f | %+.4f | %+.4f |\n", x.Method, x.Parameter, x.MeanGainFull, x.MeanGainAblated)
	}
	b.WriteString("\n## Shuffled-corpus controls\n\n| Shuffle | token JS | projected JS | gain |\n|---|---:|---:|---:|\n")
	for _, x := range a.Out.Shuffles {
		fmt.Fprintf(&b, "| %s | %.4f | %.4f | %+.4f |\n", x.Mode, x.MeanTokenJS, x.MeanProjectedJS, x.MeanGain)
	}
	b.WriteString("\n## Suffix sequences\n\n`projected_sequence_context.yaml` reports exact suffix JS, each position's projected JS, and their transparent arithmetic-mean sequence kernel for lengths 2 and 3. This tests structural resemblance without constructing a dense Cartesian product.\n\n## Families and transitions\n\n`structural_projection_families.yaml` reports cohesion, within-family dispersion, medoids, and matched percentiles for token/full/ablated spaces. `strongest_structural_transitions` in the pair YAML contains directed expected soft transitions ranked by lift over the product-of-marginals frequency baseline. Family 2 is plotted separately.\n\n## Controls and limitations\n\nRandom spaces preserve each row's degree and weights while permuting destinations within log2-frequency bins. Generic smoothing uses the same degree but uniform random neighbours in the same bins. Global and line-preserving corpus shuffles retain token frequencies. Smoothing necessarily tends to increase distributional similarity; only gain beyond these nulls is structurally specific. Structural components were learned elsewhere on the full corpus, so the ablation removes the direct context component but is not a full train/test reconstruction. Cross-validation is intentionally not claimed by this run. Pair observations and distances are dependent; percentiles are deterministic empirical diagnostics, not classical independent-sample p-values.\n")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func isMainPlot(a, b string) bool {
	if b < a {
		a, b = b, a
	}
	return a == "chedy" && b == "qokeey" || a == "chol" && b == "daiin" || a == "or" && b == "s"
}
func safe(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-' {
			b.WriteRune(r)
		} else {
			fmt.Fprintf(&b, "_%x", r)
		}
	}
	return b.String()
}
func esc(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
}
func writePairPlot(path string, p PairResult) error {
	vals := [][]float64{make([]float64, len(p.Right)), make([]float64, len(p.Right)), make([]float64, len(p.Right)), make([]float64, len(p.Right))}
	for i, m := range p.Right {
		vals[0][i] = m.TokenJS
		vals[1][i] = m.ProjectedJSFull
		vals[2][i] = m.ProjectedJSAblated
		vals[3][i] = m.TokenJS + m.RandomGainP95
	}
	return svgLines(path, p.TokenA+" / "+p.TokenB+" — structural projection", vals, []string{"token-level", "full projection", "ablated projection", "random-projection P95"})
}
func writeFamilyPlot(path string, f FamilyResult) error {
	vals := [][]float64{make([]float64, len(f.Distances)), make([]float64, len(f.Distances)), make([]float64, len(f.Distances))}
	for i, m := range f.Distances {
		vals[0][i] = m.TokenCohesion
		vals[1][i] = m.ProjectedCohesionFull
		vals[2][i] = m.ProjectedCohesionAblated
	}
	return svgLines(path, "family 2 cohesion", vals, []string{"token-level", "full projection", "ablated projection"})
}
func svgLines(path, title string, vals [][]float64, labels []string) error {
	colors := []string{"#1769aa", "#d95f02", "#2ca02c", "#7b1fa2"}
	left, top, w, h := 65., 45., 730., 390.
	n := 0
	if len(vals) > 0 {
		n = len(vals[0])
	}
	x := func(i int) float64 {
		if n < 2 {
			return left
		}
		return left + float64(i)*w/float64(n-1)
	}
	y := func(v float64) float64 { return top + (1-v)*h }
	var b strings.Builder
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="840" height="510" viewBox="0 0 840 510"><rect width="100%" height="100%" fill="white"/><style>text{font:12px sans-serif}.title{font:bold 16px sans-serif}.grid{stroke:#ddd}.axis{stroke:#333}</style>`)
	fmt.Fprintf(&b, "<text class=\"title\" x=\"20\" y=\"25\">%s</text>", esc(title))
	for i := 0; i <= 5; i++ {
		v := float64(i) / 5
		fmt.Fprintf(&b, "<line class=\"grid\" x1=\"%.0f\" y1=\"%.1f\" x2=\"%.0f\" y2=\"%.1f\"/><text x=\"25\" y=\"%.1f\">%.1f</text>", left, y(v), left+w, y(v), y(v)+4, v)
	}
	for k, v := range vals {
		fmt.Fprintf(&b, "<polyline fill=\"none\" stroke=\"%s\" stroke-width=\"%d\" points=\"", colors[k%len(colors)], 3-k/3)
		for i, q := range v {
			fmt.Fprintf(&b, "%.1f,%.1f ", x(i), y(q))
		}
		b.WriteString("\"/>")
		fmt.Fprintf(&b, "<line x1=\"%d\" y1=\"%d\" x2=\"%d\" y2=\"%d\" stroke=\"%s\" stroke-width=\"3\"/><text x=\"%d\" y=\"%d\">%s</text>", 80+k*180, 480, 105+k*180, 480, colors[k], 110+k*180, 484, labels[k])
	}
	b.WriteString("</svg>")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// Stable ordering helper for callers serializing transitions independently.
func sortedTransitions(x []Transition) []Transition {
	out := append([]Transition(nil), x...)
	sort.Slice(out, func(i, j int) bool { return out[i].Lift > out[j].Lift })
	return out
}
