package localregime

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
	c = defaults(c)
	if c.ProgressWriter == nil && !c.Quiet {
		c.ProgressWriter = os.Stderr
	}
	p := newProgress(c.ProgressWriter)
	a, e := analyze(c, p)
	if e != nil {
		return e
	}
	p.begin(7, "Writing results")
	if e = os.MkdirAll(filepath.Join(c.OutputDir, "plots"), 0o755); e != nil {
		return e
	}
	writes := []struct {
		name string
		v    any
	}{{"local_regime_pairs.yaml", a.Out}, {"local_regime_changes.yaml", struct {
		Changes []ChangePoint `yaml:"change_points"`
	}{a.Changes}}, {"token_regime_profiles.yaml", struct {
		Tokens []TokenProfile `yaml:"tokens"`
	}{a.Tokens}}, {"occurrence_regime_profiles.yaml", struct {
		Profiles map[string][]profile `yaml:"profiles"`
	}{a.Occurrence}}, {"block_shuffle_results.yaml", struct {
		Results []ShuffleResult `yaml:"results"`
	}{a.Shuffles}}}
	for _, x := range writes {
		if e = writeYAML(filepath.Join(c.OutputDir, x.name), x.v); e != nil {
			return e
		}
	}
	if e = writeTop(filepath.Join(c.OutputDir, "local_regime_top.tsv"), a.Out.Pairs); e != nil {
		return e
	}
	if e = writeControls(filepath.Join(c.OutputDir, "local_regime_controls.tsv"), a.Controls); e != nil {
		return e
	}
	if e = writeWindows(filepath.Join(c.OutputDir, "local_regime_windows.tsv"), a.Windows, a.Out.Separations); e != nil {
		return e
	}
	if e = writeReport(filepath.Join(c.OutputDir, "local_regime_report.md"), a); e != nil {
		return e
	}
	if e = plotNonstationarity(filepath.Join(c.OutputDir, "plots", "nonstationarity_window_js.svg"), a.Windows); e != nil {
		return e
	}
	if e = plotSeparations(filepath.Join(c.OutputDir, "plots", "window_similarity_by_distance.svg"), a.Out.Separations); e != nil {
		return e
	}
	for _, q := range a.Out.Pairs {
		if plotPair(q.TokenA, q.TokenB) {
			if e = plotRegime(filepath.Join(c.OutputDir, "plots", "regime_"+safe(q.TokenA)+"_"+safe(q.TokenB)+".svg"), q); e != nil {
				return e
			}
		}
	}
	p.update(1, 1, "Writing results")
	fmt.Printf("Local-regime analysis completed for %d pairs; results written to %s\n", len(a.Out.Pairs), c.OutputDir)
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
func writeTop(path string, x []PairResult) error {
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	fmt.Fprintln(w, "rank\ttoken_a\ttoken_b\tcount_a\tcount_b\tregime_js\tdistance_js_1_5\tdistance_js_6_10\tdistance_js_11_20\tregime_expected_1_5\tresidual_1_5\tconcentration_a\tconcentration_b\tlocal_shuffle_retained_fraction\tretained_effect")
	y := append([]PairResult(nil), x...)
	sort.SliceStable(y, func(i, j int) bool { return primaryRegime(y[i]) > primaryRegime(y[j]) })
	for i, p := range y {
		rf, re := 0., 0.
		if len(p.Distance) > 0 {
			var a, b []float64
			for _, d := range p.Distance {
				if d.Distance <= 5 {
					a = append(a, d.RetainedFraction)
					b = append(b, d.RetainedEffect)
				}
			}
			rf, re = mean(a), mean(b)
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%d\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", i+1, p.TokenA, p.TokenB, p.CountA, p.CountB, ff(primaryRegime(p)), ff(p.Observed1To5), ff(p.Observed6To10), ff(p.Observed11To20), ff(p.Regime1To5), ff(p.Residual1To5), ff(p.ConcentrationA), ff(p.ConcentrationB), ff(rf), ff(re))
	}
	return nil
}
func primaryRegime(p PairResult) float64 {
	return p.PrimaryRegime
}
func writeControls(path string, x []controlRow) error {
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	fmt.Fprintln(w, "target_a\ttarget_b\tcontrol_rank\tcontrol_a\tcontrol_b\tfrequency_match_score\tregime_js\tdistance_js_1_5\tregime_concentration_a\tregime_concentration_b")
	for _, r := range x {
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", r.Target.A, r.Target.B, r.Rank, r.Control.A, r.Control.B, ff(r.Score), ff(r.RegimeSimilarity), ff(r.DistanceSimilarity), ff(r.ConcentrationA), ff(r.ConcentrationB))
	}
	return nil
}
func writeWindows(path string, x []WindowRow, seps []SeparationRow) error {
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	fmt.Fprintln(w, "record_type\twindow_size\twindow_index\tstart\tend\tadjacent_js_distance\tconcentration\tseparation\tcomparisons\tmean_js_distance\tmean_similarity")
	for _, r := range x {
		fmt.Fprintf(w, "window\t%d\t%d\t%d\t%d\t%s\t%s\t\t\t\t\n", r.Size, r.Index, r.Start, r.End, ff(r.AdjacentJSDistance), ff(r.Concentration))
	}
	for _, r := range seps {
		fmt.Fprintf(w, "separation\t%d\t\t\t\t\t\t%d\t%d\t%s\t%s\n", r.Size, r.Separation, r.Comparisons, ff(r.MeanJSDistance), ff(r.MeanSimilarity))
	}
	return nil
}
func writeReport(path string, a analysis) error {
	var b strings.Builder
	b.WriteString("# Local-regime / non-stationarity analysis\n\nThe corpus is treated as a continuous token sequence. Profiles exclude the central gap; no semantic labels, manuscript metadata, projection, or property trajectories are used.\n\n## Pair decomposition\n\n| pair | regime JS | distance JS 1–5 | expected | residual |\n|---|---:|---:|---:|---:|\n")
	for _, p := range a.Out.Pairs {
		fmt.Fprintf(&b, "| `%s` / `%s` | %.4f | %.4f | %.4f | %.4f |\n", p.TokenA, p.TokenB, primaryRegime(p), p.Observed1To5, p.Regime1To5, p.Residual1To5)
	}
	b.WriteString("\n## Correlation diagnostic\n\n")
	for _, c := range a.Out.Correlations {
		fmt.Fprintf(&b, "%s: Pearson %.4f, Spearman %.4f (N=%d). These are descriptive diagnostics; ordinary independent-pair p-values are not used.\n\n", c.Metric, c.Pearson, c.Spearman, c.N)
	}
	fmt.Fprintf(&b, "Detected %d neutral distributional change points by a mean-plus-one-standard-deviation JS-jump threshold. Parameter sweeps are retained in YAML and no radius, gap, block size, or number of regimes was selected post hoc.\n\n## Interpretation\n\nCompare original, regime-expected, and local-block-shuffle series. A small residual and high retained effect support shared local composition; a large positive residual and a decrease after block shuffle support additional sequential structure. Retained fractions are ratios, not probabilities.\n", len(a.Changes))
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
func plotPair(a, b string) bool {
	k := a + "/" + b
	if b < a {
		k = b + "/" + a
	}
	return k == "chedy/qokeey" || k == "chol/daiin" || k == "or/s"
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
func svgStart(title string) string {
	title = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(title)
	return fmt.Sprintf("<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"900\" height=\"480\" viewBox=\"0 0 900 480\"><rect width=\"100%%\" height=\"100%%\" fill=\"white\"/><text x=\"450\" y=\"24\" text-anchor=\"middle\" font-family=\"sans-serif\" font-size=\"16\">%s</text><path d=\"M60 40V430H870\" fill=\"none\" stroke=\"#333\"/>", title)
}
func polyline(v []float64, color string, maxY float64) string {
	if len(v) == 0 {
		return ""
	}
	var p strings.Builder
	for i, y := range v {
		x := 60.
		if len(v) > 1 {
			x += 810 * float64(i) / float64(len(v)-1)
		}
		yy := 430.
		if maxY > 0 {
			yy -= 380 * y / maxY
		}
		fmt.Fprintf(&p, "%.1f,%.1f ", x, yy)
	}
	return fmt.Sprintf("<polyline points=\"%s\" fill=\"none\" stroke=\"%s\" stroke-width=\"2\"/>", p.String(), color)
}
func plotNonstationarity(path string, rows []WindowRow) error {
	var v []float64
	for _, r := range rows {
		if r.Size == 100 {
			v = append(v, r.AdjacentJSDistance)
		}
	}
	s := svgStart("Adjacent-window JS distance (window 100)") + polyline(v, "#2b6cb0", 1) + "</svg>"
	return os.WriteFile(path, []byte(s), 0o644)
}
func plotSeparations(path string, x []SeparationRow) error {
	var v []float64
	for _, r := range x {
		if r.Size == 100 {
			v = append(v, r.MeanSimilarity)
		}
	}
	s := svgStart("Window similarity by separation (window 100)") + polyline(v, "#805ad5", 1) + "</svg>"
	return os.WriteFile(path, []byte(s), 0o644)
}
func plotRegime(path string, p PairResult) error {
	var o, e, l, g []float64
	for _, d := range p.Distance {
		o = append(o, d.Observed)
		e = append(e, d.RegimeExpected)
		l = append(l, d.LocalBlockShuffle)
		g = append(g, d.GlobalShuffle)
	}
	s := svgStart(p.TokenA+" / "+p.TokenB) + polyline(o, "#1a202c", 1) + polyline(e, "#2b6cb0", 1) + polyline(l, "#38a169", 1) + polyline(g, "#e53e3e", 1) + "<text x=\"70\" y=\"60\" font-family=\"sans-serif\" font-size=\"12\">original (black), regime expected (blue), local shuffle (green), global shuffle (red)</text></svg>"
	return os.WriteFile(path, []byte(s), 0o644)
}
