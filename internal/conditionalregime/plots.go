package conditionalregime

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func svgWrap(title, body string) []byte {
	return []byte(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="1000" height="320" viewBox="0 0 1000 320"><rect width="100%%" height="100%%" fill="white"/><style>text{font:12px sans-serif}.t{font:bold 16px sans-serif}</style><text class="t" x="20" y="25">%s</text>%s</svg>`, title, body))
}

func writePlots(dir string, r *runResult) error {
	if err := plotWithinStabilityByScale(filepath.Join(dir, "plots", "within_class_stability_by_scale.svg"), r.WithinStability); err != nil {
		return err
	}
	if err := plotResidualClusterStability(filepath.Join(dir, "plots", "residual_cluster_stability.svg"), r.ResidualSummary); err != nil {
		return err
	}
	if err := plotOriginalVsResidualNMI(filepath.Join(dir, "plots", "original_vs_residual_currier_nmi.svg"), "currier", r.ResidualAssoc); err != nil {
		return err
	}
	if err := plotOriginalVsResidualNMI(filepath.Join(dir, "plots", "original_vs_residual_hand_nmi.svg"), "hand", r.ResidualAssoc); err != nil {
		return err
	}
	if err := plotMetadataEntropy(filepath.Join(dir, "plots", "residual_regime_metadata_entropy.svg"), r.ResidualCandRows); err != nil {
		return err
	}
	if err := plotTransitionEnrichment(filepath.Join(dir, "plots", "residual_transition_enrichment.svg"), r.Transitions); err != nil {
		return err
	}
	return nil
}

func plotWithinStabilityByScale(path string, stability []WithinClassStability) error {
	sizes := map[int]bool{}
	byMethodScale := map[string]map[int][]float64{"k_medoids": {}, "hierarchical": {}}
	for _, s := range stability {
		sizes[s.WindowSize] = true
		byMethodScale[s.Method][s.WindowSize] = append(byMethodScale[s.Method][s.WindowSize], s.Score)
	}
	var xs []int
	for s := range sizes {
		xs = append(xs, s)
	}
	sort.Ints(xs)
	var b strings.Builder
	b.WriteString(`<line x1="60" y1="260" x2="960" y2="260" stroke="#333"/><line x1="60" y1="40" x2="60" y2="260" stroke="#333"/>`)
	for i, x := range xs {
		px := 60 + 900*float64(i)/float64(max(1, len(xs)-1))
		fmt.Fprintf(&b, `<text x="%.1f" y="275">%d</text>`, px-10, x)
	}
	colors := map[string]string{"k_medoids": "#059669", "hierarchical": "#2563eb"}
	for _, method := range []string{"k_medoids", "hierarchical"} {
		fmt.Fprintf(&b, `<polyline fill="none" stroke="%s" stroke-width="2" points="`, colors[method])
		for i, x := range xs {
			vals := byMethodScale[method][x]
			px := 60 + 900*float64(i)/float64(max(1, len(xs)-1))
			py := 260 - 200*meanFloat(vals)
			fmt.Fprintf(&b, "%.1f,%.1f ", px, py)
		}
		b.WriteString(`"/>`)
	}
	fmt.Fprintf(&b, `<text x="60" y="300" fill="%s">k_medoids</text><text x="220" y="300" fill="%s">hierarchical</text>`, colors["k_medoids"], colors["hierarchical"])
	return os.WriteFile(path, svgWrap("Within-class held-out separation by scale (mean across classes)", b.String()), 0644)
}

func plotResidualClusterStability(path string, rows []ResidualClusterSummary) error {
	bestByScale := map[int]float64{}
	for _, s := range rows {
		if s.Method != "k_medoids" || s.Representation != "raw" {
			continue
		}
		if s.Silhouette > bestByScale[s.WindowSize] {
			bestByScale[s.WindowSize] = s.Silhouette
		}
	}
	var xs []int
	for x := range bestByScale {
		xs = append(xs, x)
	}
	sort.Ints(xs)
	var b strings.Builder
	b.WriteString(`<line x1="60" y1="260" x2="960" y2="260" stroke="#333"/><line x1="60" y1="40" x2="60" y2="260" stroke="#333"/>`)
	fmt.Fprintf(&b, `<polyline fill="none" stroke="#059669" stroke-width="2" points="`)
	for i, x := range xs {
		px := 60 + 900*float64(i)/float64(max(1, len(xs)-1))
		py := 260 - 200*bestByScale[x]
		fmt.Fprintf(&b, "%.1f,%.1f ", px, py)
		fmt.Fprintf(&b, `" /><text x="%.1f" y="275">%d</text><polyline fill="none" stroke="#059669" stroke-width="2" points="`, px-10, x)
	}
	b.WriteString(`"/>`)
	return os.WriteFile(path, svgWrap("Best pooled residual silhouette by scale (k_medoids, raw)", b.String()), 0644)
}

func plotOriginalVsResidualNMI(path, kind string, rows []ResidualMetadataAssociation) error {
	var orig, resid float64
	for _, a := range rows {
		if a.Metadata == kind {
			orig, resid = a.OriginalNMI, a.ResidualNMI
		}
	}
	var b strings.Builder
	b.WriteString(`<line x1="60" y1="260" x2="960" y2="260" stroke="#333"/>`)
	fmt.Fprintf(&b, `<rect x="150" y="%.1f" width="120" height="%.1f" fill="#2563eb"/><text x="150" y="275">original</text>`, 260-200*orig, 200*orig)
	fmt.Fprintf(&b, `<rect x="400" y="%.1f" width="120" height="%.1f" fill="#dc2626"/><text x="400" y="275">residual</text>`, 260-200*resid, 200*resid)
	fmt.Fprintf(&b, `<text x="150" y="%.1f">%.3f</text><text x="400" y="%.1f">%.3f</text>`, 260-200*orig-6, orig, 260-200*resid-6, resid)
	return os.WriteFile(path, svgWrap("Original vs residual global max NMI — "+kind, b.String()), 0644)
}

func plotMetadataEntropy(path string, rows []ResidualCandidate) error {
	var b strings.Builder
	b.WriteString(`<line x1="60" y1="260" x2="960" y2="260" stroke="#333"/>`)
	n := len(rows)
	for i, c := range rows {
		x := 70 + float64(i)*(880/float64(max(1, n)))
		h := 200 * c.MetadataEntropy / 2.0
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="24" height="%.1f" fill="#7c3aed"/><text x="%.1f" y="275">c%d</text>`, x, 260-h, h, x, c.Cluster)
	}
	return os.WriteFile(path, svgWrap("Residual regime metadata (Currier x hand) entropy", b.String()), 0644)
}

func plotTransitionEnrichment(path string, cells []ResidualTransitionCell) error {
	k := 0
	for _, c := range cells {
		if c.From+1 > k {
			k = c.From + 1
		}
		if c.To+1 > k {
			k = c.To + 1
		}
	}
	var b strings.Builder
	if k == 0 {
		return os.WriteFile(path, svgWrap("Residual transition enrichment (no transitions)", b.String()), 0644)
	}
	cell := 900.0 / float64(k)
	for _, c := range cells {
		shade := int(255 * clamp01(0.5+c.Stats.EffectSize/6))
		x := 60 + float64(c.To)*cell
		y := 40 + float64(c.From)*cell
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="rgb(%d,%d,255)"/>`, x, y, cell-1, cell-1, 255-shade, 255-shade)
	}
	b.WriteString(`<text x="60" y="30">to cluster (columns) / from cluster (rows); redder = more enriched vs Null B</text>`)
	return os.WriteFile(path, svgWrap("Residual regime transition enrichment", b.String()), 0644)
}

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}
