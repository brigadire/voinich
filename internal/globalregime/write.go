package globalregime

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
	a, err := analyze(c, p)
	if err != nil {
		return err
	}
	p.begin(7, "Writing results")
	if err = os.MkdirAll(filepath.Join(c.OutputDir, "plots"), 0o755); err != nil {
		return err
	}
	if err = writeYAML(filepath.Join(c.OutputDir, "global_distributional_regimes.yaml"), a.Out); err != nil {
		return err
	}
	if err = writeWindows(filepath.Join(c.OutputDir, "global_distributional_windows.tsv"), a.Scales); err != nil {
		return err
	}
	if err = writeChanges(filepath.Join(c.OutputDir, "global_distributional_change_points.tsv"), a.Scales); err != nil {
		return err
	}
	if err = writeStable(filepath.Join(c.OutputDir, "stable_distributional_boundaries.tsv"), a.Out.Boundaries, c.WindowSizes); err != nil {
		return err
	}
	if err = writeDiagnostics(filepath.Join(c.OutputDir, "global_distributional_clustering.tsv"), a.Out.Diagnostics); err != nil {
		return err
	}
	if err = writeAssignments(filepath.Join(c.OutputDir, "global_distributional_cluster_assignments.tsv"), a.Scales); err != nil {
		return err
	}
	if err = writeReport(filepath.Join(c.OutputDir, "global_distributional_report.md"), a); err != nil {
		return err
	}
	if err = plotChangeProfiles(filepath.Join(c.OutputDir, "plots", "global_distributional_change_profiles.svg"), a.Scales); err != nil {
		return err
	}
	if err = plotModelSelection(filepath.Join(c.OutputDir, "plots", "global_distributional_model_selection.svg"), a.Out.Diagnostics); err != nil {
		return err
	}
	p.update(1, 1, "Writing results")
	fmt.Printf("Global distributional regime analysis completed at %d scales; results written to %s\n", len(a.Scales), c.OutputDir)
	return nil
}

func writeYAML(path string, v any) error {
	b, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
func ff(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }
func createWriter(path string) (*os.File, *bufio.Writer, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, nil, err
	}
	return f, bufio.NewWriter(f), nil
}
func writeWindows(path string, scales []scaleAnalysis) error {
	f, w, err := createWriter(path)
	if err != nil {
		return err
	}
	defer f.Close()
	defer w.Flush()
	fmt.Fprintln(w, "window_size\twindow_index\tstart\tend\tcenter\tstep\tadjacent_js_distance\tweighted_overlap\tcosine_similarity\tvariation\tlocal_peak\tbroad_transition")
	for _, s := range scales {
		for _, x := range s.windows {
			fmt.Fprintf(w, "%d\t%d\t%d\t%d\t%d\t%d\t%s\t%s\t%s\t%s\t%t\t%t\n", x.WindowSize, x.Index, x.Start, x.End, x.Center, x.Step, ff(x.JSDistance), ff(x.WeightedOverlap), ff(x.Cosine), x.Variation, x.LocalPeak, x.BroadTransition)
		}
	}
	return nil
}
func writeChanges(path string, scales []scaleAnalysis) error {
	f, w, err := createWriter(path)
	if err != nil {
		return err
	}
	defer f.Close()
	defer w.Flush()
	fmt.Fprintln(w, "window_size\tposition\tmethod\tjump_strength\tthreshold")
	for _, s := range scales {
		for _, x := range s.changes {
			fmt.Fprintf(w, "%d\t%d\t%s\t%s\t%s\n", x.WindowSize, x.Position, x.Method, ff(x.Strength), ff(x.Threshold))
		}
	}
	return nil
}
func writeStable(path string, x []StableBoundary, sizes []int) error {
	f, w, err := createWriter(path)
	if err != nil {
		return err
	}
	defer f.Close()
	defer w.Flush()
	sizes = append([]int(nil), sizes...)
	sort.Ints(sizes)
	fmt.Fprint(w, "position")
	for _, s := range sizes {
		fmt.Fprintf(w, "\tsupport_%d", s)
	}
	fmt.Fprintln(w, "\tsupport_count\tsupport_fraction\tmean_position\tmean_jump_strength\tmax_jump_strength\tposition_uncertainty")
	for _, b := range x {
		fmt.Fprint(w, b.Position)
		for _, s := range sizes {
			v := 0
			if b.Support[s] {
				v = 1
			}
			fmt.Fprintf(w, "\t%d", v)
		}
		fmt.Fprintf(w, "\t%d\t%s\t%s\t%s\t%s\t%s\n", b.SupportCount, ff(b.SupportFraction), ff(b.MeanPosition), ff(b.MeanJumpStrength), ff(b.MaxJumpStrength), ff(b.PositionUncertainty))
	}
	return nil
}
func ints(x []int) string {
	v := make([]string, len(x))
	for i, n := range x {
		v[i] = strconv.Itoa(n)
	}
	return strings.Join(v, ",")
}
func writeDiagnostics(path string, x []ClusterDiagnostic) error {
	f, w, err := createWriter(path)
	if err != nil {
		return err
	}
	defer f.Close()
	defer w.Flush()
	fmt.Fprintln(w, "window_size\tmethod\tk\tsilhouette\twithin_cluster_dispersion\tbetween_cluster_distance\tcluster_sizes\ttransition_count\tsegment_fragmentation")
	for _, d := range x {
		fmt.Fprintf(w, "%d\t%s\t%d\t%s\t%s\t%s\t%s\t%d\t%s\n", d.WindowSize, d.Method, d.K, ff(d.Silhouette), ff(d.WithinDispersion), ff(d.BetweenDistance), ints(d.ClusterSizes), d.TransitionCount, ff(d.Fragmentation))
	}
	return nil
}
func writeAssignments(path string, scales []scaleAnalysis) error {
	f, w, err := createWriter(path)
	if err != nil {
		return err
	}
	defer f.Close()
	defer w.Flush()
	fmt.Fprintln(w, "window_size\tmethod\tk\twindow_index\tstart\tend\tcluster")
	for _, s := range scales {
		for _, d := range s.diagnostics {
			for i, label := range d.labels {
				x := s.windows[i]
				fmt.Fprintf(w, "%d\t%s\t%d\t%d\t%d\t%d\t%d\n", s.size, d.Method, d.K, i, x.Start, x.End, label)
			}
		}
	}
	return nil
}

func writeReport(path string, a analysis) error {
	var b strings.Builder
	b.WriteString("# Global distributional regime discovery\n\nThe corpus is treated as one continuous token sequence. Discovery uses no folio, page, illustration, Currier, hand, or section metadata. Full window token distributions retain all probability mass; no RARE bucket is used.\n\n## Continuous change profile\n\nAdjacent-window Jensen–Shannon distance is the primary result. Weighted overlap and cosine similarity are retained as companion diagnostics. Local peaks, broad transitions, and low/high-variation intervals are descriptive labels rather than assumed true boundaries.\n\n## Stable boundaries\n\n| position | scale support | mean jump | uncertainty |\n|---:|---:|---:|---:|\n")
	limit := min(30, len(a.Out.Boundaries))
	for _, x := range a.Out.Boundaries[:limit] {
		fmt.Fprintf(&b, "| %d | %d/%d | %.5f | %.1f |\n", x.Position, x.SupportCount, len(a.Scales), x.MeanJumpStrength, x.PositionUncertainty)
	}
	fmt.Fprintf(&b, "\nCandidates combine threshold peaks, PELT on the distributional JS-jump series, and binary segmentation. Cross-scale matches use ±0.5 × the smaller window size. Multiple detector hits at one scale count as one scale vote.\n\n## Clustering diagnostics\n\nHierarchical single-link clustering and JS-distance k-medoids are unconstrained, so distant windows may share a regime. Contiguous binary segmentation is reported separately. K=2..15 is swept without selecting K by interpretability. To bound quadratic fitting, up to %d sequence-wide uniformly spaced windows are used for fit and silhouette diagnostics; every window is then assigned to a fitted distribution centroid. Cluster sizes, transitions, fragmentation, and the assignment table cover all windows. These clusterings are diagnostics and do not replace the continuous profile.\n", maxClusterFitWindows)
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func esc(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}
func plotChangeProfiles(path string, scales []scaleAnalysis) error {
	width, height := 1000, 620
	var b strings.Builder
	fmt.Fprintf(&b, "<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"%d\" height=\"%d\" viewBox=\"0 0 %d %d\"><rect width=\"100%%\" height=\"100%%\" fill=\"white\"/><text x=\"500\" y=\"24\" text-anchor=\"middle\" font-family=\"sans-serif\" font-size=\"16\">Multi-scale adjacent-window JS distance</text>", width, height, width, height)
	colors := []string{"#2563eb", "#059669", "#d97706", "#dc2626", "#7c3aed"}
	for si, s := range scales {
		y0 := 55 + si*105
		fmt.Fprintf(&b, "<path d=\"M60 %dH970\" stroke=\"#ddd\"/><text x=\"8\" y=\"%d\" font-family=\"sans-serif\" font-size=\"12\">w=%d</text><polyline fill=\"none\" stroke=\"%s\" stroke-width=\"1.5\" points=\"", y0+80, y0+15, s.size, colors[si%len(colors)])
		maxY := 0.
		for _, w := range s.windows {
			if w.JSDistance > maxY {
				maxY = w.JSDistance
			}
		}
		for i, w := range s.windows {
			x := 60.
			if len(s.windows) > 1 {
				x += 910 * float64(i) / float64(len(s.windows)-1)
			}
			y := float64(y0 + 80)
			if maxY > 0 {
				y -= 70 * w.JSDistance / maxY
			}
			fmt.Fprintf(&b, "%.1f,%.1f ", x, y)
		}
		b.WriteString("\"/>")
	}
	b.WriteString("</svg>")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
func plotModelSelection(path string, x []ClusterDiagnostic) error {
	var b strings.Builder
	b.WriteString("<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"900\" height=\"480\" viewBox=\"0 0 900 480\"><rect width=\"100%\" height=\"100%\" fill=\"white\"/><text x=\"450\" y=\"24\" text-anchor=\"middle\" font-family=\"sans-serif\" font-size=\"16\">Silhouette sweep (window 200)</text><path d=\"M60 40V430H870\" fill=\"none\" stroke=\"#333\"/>")
	methods := []string{"hierarchical", "k_medoids", "contiguous_segmentation"}
	colors := []string{"#2563eb", "#059669", "#dc2626"}
	for mi, m := range methods {
		var rows []ClusterDiagnostic
		for _, d := range x {
			if d.WindowSize == 200 && d.Method == m {
				rows = append(rows, d)
			}
		}
		if len(rows) == 0 {
			continue
		}
		fmt.Fprintf(&b, "<polyline fill=\"none\" stroke=\"%s\" stroke-width=\"2\" points=\"", colors[mi])
		for i, d := range rows {
			px := 60 + 810*float64(i)/float64(max(1, len(rows)-1))
			py := 430 - 360*(d.Silhouette+1)/2
			fmt.Fprintf(&b, "%.1f,%.1f ", px, py)
		}
		fmt.Fprintf(&b, "\"/><text x=\"%d\" y=\"%d\" font-family=\"sans-serif\" font-size=\"12\" fill=\"%s\">%s</text>", 70+mi*230, 60, colors[mi], esc(m))
	}
	b.WriteString("</svg>")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
