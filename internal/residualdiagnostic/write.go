package residualdiagnostic

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

func writeResults(c Config, r *results) error {
	if err := os.MkdirAll(filepath.Join(c.OutputDir, "plots"), 0755); err != nil {
		return err
	}
	for _, name := range sortedKeys(r.Files) {
		if err := writeTSV(filepath.Join(c.OutputDir, name), r.Files[name]); err != nil {
			return err
		}
	}
	b, err := yaml.Marshal(r.Summary)
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(c.OutputDir, "residual_diagnostic_summary.yaml"), b, 0644); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(c.OutputDir, "residual_diagnostic_report.md"), []byte(report(c, r)), 0644); err != nil {
		return err
	}
	return writePlots(c, r)
}
func writeTSV(path string, rows [][]string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	for _, r := range rows {
		if _, err := fmt.Fprintln(w, strings.Join(r, "\t")); err != nil {
			return err
		}
	}
	return nil
}

func report(c Config, r *results) string {
	cent := r.Summary["centering"].(map[string]float64)
	decision := r.Summary["decision"].(string)
	var b strings.Builder
	fmt.Fprintf(&b, "# Residual diagnostic report\n\n## Answer\n\n%s.\n\n", decision)
	fmt.Fprintf(&b, "The frozen diagnostic uses only the preselected `window=%d`, `K=%d`, k-medoids solution. No new scale, K, method, or token feature search was performed.\n\n", c.WindowSize, c.K)
	fmt.Fprintf(&b, "Training-fold residual mean is numerical zero (maximum L2 `%g`), while the largest held-out mean L2 is `%g`. This distinction is essential: each held-out physical block was transformed using estimates from training blocks only.\n\n", cent["maximum_training_residual_mean_l2"], cent["maximum_held_out_residual_mean_l2"])
	fmt.Fprintf(&b, "The headline Currier values compare the frozen global original baseline (`%.4f`) with the frozen residual winner (`%.4f`). NMI is not monotone under mean centering: centering changes the geometry and the selected residual partition isolates a metadata-pure physical block, so its label association can increase even though its silhouette does not beat the null.\n\n", r.OriginalCurrierNMI, r.Representations[1].CurrierNMI)
	clusters := map[int][]window{}
	for _, x := range r.Windows {
		clusters[x.ExistingCluster] = append(clusters[x.ExistingCluster], x)
	}
	clusterIDs := make([]int, 0, len(clusters))
	for cl := range clusters {
		clusterIDs = append(clusterIDs, cl)
	}
	sort.Ints(clusterIDs)
	for _, cl := range clusterIDs {
		ws := clusters[cl]
		blockSet, jointSet := map[string]bool{}, map[string]bool{}
		for _, x := range ws {
			blockSet[x.Block] = true
			jointSet[x.Joint] = true
		}
		if len(blockSet) == 1 && len(jointSet) == 1 {
			first := ws[0]
			minStart, maxEnd := first.Start, first.End
			for _, x := range ws {
				if x.Start < minStart {
					minStart = x.Start
				}
				if x.End > maxEnd {
					maxEnd = x.End
				}
			}
			fmt.Fprintf(&b, "The single-block cluster is exactly cluster `%d`: `%d` windows, Currier `%s`, hand `%s`, joint class `%s`, physical block `%s` (`[%d,%d)`), with window coverage `[%d,%d)`. It is one contiguous run.\n\n", cl, len(ws), first.Currier, first.Hand, first.Joint, first.Block, first.PhysicalStart, first.PhysicalEnd, minStart, maxEnd)
		}
	}
	b.WriteString("## Representation comparison\n\n| representation | silhouette | Currier NMI | hand NMI | joint NMI | block NMI |\n|---|---:|---:|---:|---:|---:|\n")
	for _, x := range r.Representations {
		fmt.Fprintf(&b, "| %s | %.4f | %.4f | %.4f | %.4f | %.4f |\n", x.Name, x.Silhouette, x.CurrierNMI, x.HandNMI, x.JointNMI, x.BlockNMI)
	}
	norm := r.Summary["norm_only"].(map[string]float64)
	fmt.Fprintf(&b, "\nNorm-only K=2 reproduces the frozen labels only partially (ARI `%.4f`, NMI `%.4f`), while frozen cluster–physical-block NMI is `%.4f`. Thus magnitude contributes, but the decisive geometry is also block-specific.\n", norm["ari_with_existing"], norm["nmi_with_existing"], r.Representations[1].BlockNMI)
	b.WriteString("\n## Interpretation guardrails\n\n- Mean centering removes training class means, not held-out drift, covariance, dispersion, sparsity, block identity, or position.\n- Whitening uses `0.9 Σ + 0.1 diag(Σ)` and a `1e-6 × largest-eigenvalue` floor, estimated within each training fold only.\n- A cluster confined to a contiguous physical block is not called a reproducible regime. Recurrence requires both clusters to pass the training within-cluster reference threshold in held-out blocks.\n- Block-aware permutations preserve physical grouping; no random window split is used.\n\n## Decision\n\n")
	b.WriteString(decision + ". Until the stated recurrence and metadata-removal criteria are met, residual discovery should not be expanded.\n")
	return b.String()
}

func writePlots(c Config, r *results) error {
	plots := map[string]string{}
	plots["residual_norm_by_currier.svg"] = barPlot("Residual norm by Currier", groupNorm(r.Windows, "currier"))
	plots["residual_norm_by_hand.svg"] = barPlot("Residual norm by hand", groupNorm(r.Windows, "hand"))
	plots["residual_cluster_position.svg"] = positionPlot(r.Windows)
	plots["residual_cluster_block_composition.svg"] = stackedBlocks(r.Windows)
	vals := map[string]float64{}
	for _, x := range r.Representations {
		vals[x.Name+"/Currier"] = x.CurrierNMI
		vals[x.Name+"/hand"] = x.HandNMI
	}
	plots["original_vs_centered_vs_whitened_metadata_nmi.svg"] = barPlot("Currier NMI by representation", vals)
	drift := map[string]float64{}
	for _, d := range r.Folds {
		key := d.Joint + "/f" + strconv.Itoa(d.Fold)
		drift[key] = d.TestMean.L2
	}
	plots["heldout_residual_drift.svg"] = barPlot("Held-out residual mean L2", drift)
	for name, s := range plots {
		if err := os.WriteFile(filepath.Join(c.OutputDir, "plots", name), []byte(s), 0644); err != nil {
			return err
		}
	}
	return nil
}
func groupNorm(w []window, dim string) map[string]float64 {
	sum := map[string]float64{}
	n := map[string]int{}
	labels := labelsOf(w, dim)
	for i, x := range w {
		sum[labels[i]] += normOf(x.Residual).L2
		n[labels[i]]++
	}
	for k := range sum {
		sum[k] /= float64(n[k])
	}
	return sum
}
func svgStart(title string) string {
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="960" height="520" viewBox="0 0 960 520"><rect width="100%%" height="100%%" fill="white"/><text x="480" y="28" text-anchor="middle" font-family="sans-serif" font-size="20">%s</text>`, xmlEscape(title))
}
func barPlot(title string, v map[string]float64) string {
	keys := make([]string, 0, len(v))
	mx := 0.
	for k, x := range v {
		keys = append(keys, k)
		if x > mx {
			mx = x
		}
	}
	sort.Strings(keys)
	if mx == 0 {
		mx = 1
	}
	var b strings.Builder
	b.WriteString(svgStart(title))
	width := 800. / float64(max(1, len(keys)))
	for i, k := range keys {
		x := 80 + float64(i)*width
		h := 400 * v[k] / mx
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="#4c78a8"/><text x="%.1f" y="485" text-anchor="middle" font-family="sans-serif" font-size="11">%s</text><text x="%.1f" y="%.1f" text-anchor="middle" font-family="sans-serif" font-size="11">%.3g</text>`, x, 450-h, width*.72, h, x+width*.36, xmlEscape(k), x+width*.36, 445-h, v[k])
	}
	b.WriteString(`</svg>`)
	return b.String()
}
func positionPlot(w []window) string {
	var b strings.Builder
	b.WriteString(svgStart("Frozen residual clusters across corpus position"))
	mx := 1
	for _, x := range w {
		if x.End > mx {
			mx = x.End
		}
	}
	colors := []string{"#e45756", "#4c78a8", "#72b7b2", "#f58518"}
	for _, x := range w {
		cx := 60 + 850*float64((x.Start+x.End)/2)/float64(mx)
		cy := 120 + float64(x.ExistingCluster)*180
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.5" fill="%s" opacity=".75"/>`, cx, cy, colors[x.ExistingCluster%len(colors)])
	}
	b.WriteString(`</svg>`)
	return b.String()
}
func stackedBlocks(w []window) string {
	blocks := uniqueStrings(labelsOf(w, "block"))
	counts := map[string]map[int]int{}
	for _, x := range w {
		if counts[x.Block] == nil {
			counts[x.Block] = map[int]int{}
		}
		counts[x.Block][x.ExistingCluster]++
	}
	var b strings.Builder
	b.WriteString(svgStart("Residual cluster composition by physical block"))
	barH := 430. / float64(max(1, len(blocks)))
	colors := []string{"#e45756", "#4c78a8", "#72b7b2"}
	for i, bl := range blocks {
		total := 0
		for _, n := range counts[bl] {
			total += n
		}
		x := 160.
		y := 55 + float64(i)*barH
		for c := 0; c < 8; c++ {
			n := counts[bl][c]
			if n == 0 {
				continue
			}
			ww := 740 * float64(n) / float64(total)
			fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`, x, y, ww, barH*.72, colors[c%len(colors)])
			x += ww
		}
		fmt.Fprintf(&b, `<text x="150" y="%.1f" text-anchor="end" font-family="sans-serif" font-size="11">%s</text>`, y+barH*.5, xmlEscape(bl))
	}
	b.WriteString(`</svg>`)
	return b.String()
}

func xmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&apos;")
	return r.Replace(s)
}
