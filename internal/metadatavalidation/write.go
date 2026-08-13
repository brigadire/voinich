package metadatavalidation

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func coverage(x []TokenMetadata, get func(TokenMetadata) string) (int, int, float64) {
	known := 0
	for _, r := range x {
		if v := get(r); v != "" && v != "?" {
			known++
		}
	}
	return known, len(x) - known, ratio(known, len(x))
}
func writeAlignmentReport(path string, c Config, r runResult) error {
	var b strings.Builder
	first, last := tokenEdges(r.Alignment.Tokens)
	fmt.Fprintf(&b, "# IVTFF alignment report\n\nResult: **PASS**\n\n- IVTFF file: `%s`\n- frozen corpus file: `%s`\n- frozen corpus SHA256 (recorded now, not historical): `%s`\n- total pages: %d\n- total loci: %d\n- skipped/comment-only loci: %d\n- total frozen tokens: %d\n- aligned tokens: %d\n- mismatches: 0\n- discovery token count: %d\n- indexing: zero-based token positions, boundaries are positions between token `p-1` and token `p`\n\nThe concatenation of all aligned frozen ranges is token-identical to the complete frozen corpus. The parser only creates an alignment representation and does not replace IVTT.\n\n## Metadata coverage\n\n| Metadata | Known tokens | Unknown tokens | Coverage |\n|---|---:|---:|---:|\n", c.IVTFFPath, c.FrozenCorpusPath, r.Alignment.CorpusSHA256, r.Doc.Pages, r.Alignment.TotalLoci, r.Alignment.SkippedLoci, len(r.Alignment.Tokens), len(r.Alignment.Records), r.DiscoveryTokenCount)
	fields := []struct {
		name string
		get  func(TokenMetadata) string
	}{{"Currier ($C)", func(x TokenMetadata) string { return x.Currier }}, {"Hand ($H)", func(x TokenMetadata) string { return x.Hand }}, {"Folio", func(x TokenMetadata) string { return x.Folio }}, {"Paragraph", func(x TokenMetadata) string {
		if x.ParagraphID > 0 {
			return "known"
		}
		return ""
	}}, {"Quire ($Q; no folio heuristic)", func(x TokenMetadata) string { return x.Quire }}}
	for _, q := range fields {
		k, u, p := coverage(r.Alignment.Records, q.get)
		fmt.Fprintf(&b, "| %s | %d | %d | %.2f%% |\n", q.name, k, u, 100*p)
	}
	fmt.Fprintf(&b, "\nFrozen/aligned first tokens: `%s`; last tokens: `%s`. Their identity follows from the exact global invariant. Historical discovery metadata contains token count but no stored edge-token sample.\n", first, last)
	b.WriteString("\n## Alignment-only normalization\n\nIVTFF comments and `<%>`, `<$>` controls are omitted; `<->`, dots and commas become boundaries; the first explicit `[first:alternative]` reading is selected; braces retain their contents; `@NNN;`, apostrophes and `?` are preserved. Every rule is used solely to compare loci with the canonical frozen tokens.\n")
	return os.WriteFile(path, []byte(b.String()), 0644)
}

func writeFailedAlignmentReport(path string, c Config, d Document, a AlignmentResult, alignmentErr error) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# IVTFF alignment report\n\nResult: **FAIL**\n\n- IVTFF file: `%s`\n- frozen corpus file: `%s`\n- frozen corpus SHA256 (recorded now, not historical): `%s`\n- total pages: %d\n- total loci: %d\n- skipped/comment-only loci: %d\n- total frozen tokens: %d\n- aligned tokens before failure: %d\n- mismatches: 1\n- metadata validation executed: no\n- indexing: zero-based\n\n", c.IVTFFPath, c.FrozenCorpusPath, a.CorpusSHA256, d.Pages, a.TotalLoci, a.SkippedLoci, len(a.Tokens), len(a.Records))
	for _, q := range []struct {
		name string
		get  func(TokenMetadata) string
	}{{"Currier", func(x TokenMetadata) string { return x.Currier }}, {"Hand", func(x TokenMetadata) string { return x.Hand }}, {"Folio", func(x TokenMetadata) string { return x.Folio }}, {"Quire", func(x TokenMetadata) string { return x.Quire }}} {
		k, u, p := coverage(a.Records, q.get)
		fmt.Fprintf(&b, "- %s coverage before failure: %d known, %d unknown (%.2f%%)\n", q.name, k, u, 100*p)
	}
	fmt.Fprintf(&b, "\n%s\n", alignmentErr)
	return os.WriteFile(path, []byte(b.String()), 0644)
}

func tokenEdges(x []string) (string, string) {
	n := min(10, len(x))
	return strings.Join(x[:n], " "), strings.Join(x[len(x)-n:], " ")
}

func nearestFor(p int, refs map[string][]MetadataBoundary, kind string) (int, int) {
	return NearestBoundary(p, positions(refs[kind]))
}
func strongestRows(stable []StableBoundary, refs map[string][]MetadataBoundary, onlyResidual bool) []string {
	rows := []string{}
	highJump := 0.0
	if onlyResidual {
		jumps := []float64{}
		for _, b := range stable {
			if b.Support >= 4 {
				jumps = append(jumps, b.MeanJump)
			}
		}
		sort.Float64s(jumps)
		if len(jumps) > 0 {
			highJump = jumps[(len(jumps)*3)/4]
		}
	}
	limit := len(stable)
	if !onlyResidual && limit > 100 {
		limit = 100
	}
	for _, b := range stable[:limit] {
		vals := []int{}
		nearest := []int{}
		for _, kind := range []string{"folio", "paragraph", "currier", "hand", "quire"} {
			n, d := nearestFor(b.Position, refs, kind)
			nearest = append(nearest, n)
			vals = append(vals, d)
		}
		minD := -1
		for _, d := range vals {
			if d >= 0 && (minD < 0 || d < minD) {
				minD = d
			}
		}
		if onlyResidual && (b.Support < 4 || b.MeanJump < highJump || minD <= 200) {
			continue
		}
		rows = append(rows, fmt.Sprintf("%d\t%d\t%s\t%s\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d", b.Position, b.Support, f(b.MeanJump), f(b.Uncertainty), nearest[0], vals[0], nearest[1], vals[1], nearest[2], vals[2], nearest[3], vals[3], nearest[4], vals[4], minD))
	}
	return rows
}
func writeStrongest(path string, stable []StableBoundary, refs map[string][]MetadataBoundary, residual bool) error {
	return tsv(path, "position\tscale_support\tmean_jump\tuncertainty\tnearest_folio\tdistance_folio\tnearest_paragraph\tdistance_paragraph\tnearest_currier_transition\tdistance_currier\tnearest_hand_transition\tdistance_hand\tnearest_quire\tdistance_quire\tnearest_known_metadata_distance", strongestRows(stable, refs, residual))
}

func writeMetadataYAML(path string, c Config, r runResult) error {
	first, last := tokenEdges(r.Alignment.Tokens)
	cov := map[string]any{}
	for _, q := range []struct {
		name string
		get  func(TokenMetadata) string
	}{{"currier", func(x TokenMetadata) string { return x.Currier }}, {"hand", func(x TokenMetadata) string { return x.Hand }}, {"folio", func(x TokenMetadata) string { return x.Folio }}, {"quire", func(x TokenMetadata) string { return x.Quire }}} {
		k, u, p := coverage(r.Alignment.Records, q.get)
		cov[q.name] = map[string]any{"known_tokens": k, "unknown_tokens": u, "coverage": p}
	}
	return writeYAML(path, map[string]any{"result": "PASS", "category": validationCategory(r.BoundaryRows), "canonical_corpus": map[string]any{"path": c.FrozenCorpusPath, "token_count": len(r.Alignment.Tokens), "first_tokens": first, "last_tokens": last, "sha256_recorded_now": r.Alignment.CorpusSHA256, "historical_hash_available": false, "historical_edge_tokens_available": false, "indexing": "zero-based"}, "alignment": map[string]any{"ivtff": c.IVTFFPath, "loci": r.Alignment.TotalLoci, "aligned_tokens": len(r.Alignment.Records), "mismatches": 0}, "metadata_coverage": cov, "discovery_frozen": true, "permutations": c.Permutations, "seed": c.Seed, "tolerances": c.Tolerances})
}

func validationCategory(x []BoundaryValidation) string {
	strong, total := 0, 0
	for _, r := range x {
		if r.MinSupport == 4 && r.Tolerance == 50 && (r.Kind == "folio" || r.Kind == "paragraph" || r.Kind == "currier" || r.Kind == "hand") {
			total++
			if r.UniformPercentile >= 95 && r.CircularPercentile >= 95 {
				strong++
			}
		}
	}
	if total > 0 && strong*2 >= total {
		return "Strong alignment"
	}
	if strong > 0 {
		return "Partial alignment"
	}
	return "Weak alignment"
}
func writeValidationReport(path string, c Config, r runResult) error {
	cat := validationCategory(r.BoundaryRows)
	var b strings.Builder
	fmt.Fprintf(&b, "# Blind metadata validation\n\nAlignment result: **PASS**. Validation category: **%s**. This category describes association with reference metadata, not accuracy against ground truth.\n\nThe frozen IVTT `-x7 ASCII Full` sequence (%d tokens) and all frozen distributional boundaries and cluster assignments were used unchanged. The IVTFF parser supplies metadata only. Corpus SHA256 recorded during this run: `%s`; it was not historically stored by discovery.\n\n## Boundary association\n\n| Metadata | support | tolerance | matched / blind | uniform percentile | circular percentile |\n|---|---:|---:|---:|---:|---:|\n", cat, len(r.Alignment.Tokens), r.Alignment.CorpusSHA256)
	for _, v := range r.BoundaryRows {
		if v.MinSupport == 4 && v.Tolerance == 50 {
			fmt.Fprintf(&b, "| %s | ≥%d/5 | ±%d | %d/%d | %.1f | %.1f |\n", v.Kind, v.MinSupport, v.Tolerance, v.Matched, v.BlindCount, v.UniformPercentile, v.CircularPercentile)
		}
	}
	b.WriteString("\nLine transitions are a dense sanity control and are not interpreted from raw overlap alone. Known→unknown and unknown→known changes are excluded from Currier, hand and quire transitions. Fixed tolerances were prespecified and are all retained.\n\n## Frozen cluster association\n\n`cluster_metadata_association.tsv` reports MI, NMI, ARI, homogeneity, completeness and conditional entropy for every frozen window scale, method and K, both for all known-majority windows and purity ≥0.8/≥0.9 subsets. Metadata never changes clusters. The max-over-K permutation control shuffles labels among contiguous blocks while retaining block lengths.\n\n## Residual structure\n\n`unexplained_distributional_structure.tsv` contains support ≥4/5 boundaries farther than 200 tokens from every available tested transition. These are neutral unexplained distributional structures, not proposed sections or languages.\n")
	return os.WriteFile(path, []byte(b.String()), 0644)
}

func writePlots(dir string, r runResult) error {
	for _, kind := range []string{"folio", "paragraph", "currier", "hand"} {
		if e := plotBoundaries(filepath.Join(dir, "plots", "blind_boundaries_vs_"+kind+".svg"), kind, r.Stable, positions(r.References[kind]), len(r.Alignment.Tokens)); e != nil {
			return e
		}
	}
	if e := plotCDF(filepath.Join(dir, "plots", "boundary_distance_cdf.svg"), r); e != nil {
		return e
	}
	for _, kind := range []string{"currier", "hand"} {
		if e := plotNMI(filepath.Join(dir, "plots", "nmi_by_k_"+kind+".svg"), kind, r.Associations); e != nil {
			return e
		}
		if e := plotHeatmap(filepath.Join(dir, "plots", "regime_"+kind+"_heatmap.svg"), kind, r.Associations); e != nil {
			return e
		}
	}
	return nil
}
func svgWrap(title, body string) []byte {
	return []byte(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="1000" height="300" viewBox="0 0 1000 300"><rect width="100%%" height="100%%" fill="white"/><style>text{font:12px sans-serif}.t{font:bold 16px sans-serif}</style><text class="t" x="20" y="25">%s</text>%s</svg>`, title, body))
}
func plotBoundaries(path, kind string, blind []StableBoundary, refs []int, n int) error {
	var b strings.Builder
	b.WriteString(`<line x1="40" y1="150" x2="960" y2="150" stroke="#aaa"/>`)
	for _, p := range refs {
		x := 40 + 920*float64(p)/float64(n)
		fmt.Fprintf(&b, `<line x1="%.2f" y1="100" x2="%.2f" y2="150" stroke="#2563eb" opacity=".45"/>`, x, x)
	}
	for _, p := range blind {
		if p.Support < 3 {
			continue
		}
		x := 40 + 920*float64(p.Position)/float64(n)
		fmt.Fprintf(&b, `<line x1="%.2f" y1="150" x2="%.2f" y2="220" stroke="#dc2626" stroke-width="2"/>`, x, x)
	}
	b.WriteString(`<text x="40" y="90">metadata transitions</text><text x="40" y="245">blind boundaries (support ≥3/5)</text>`)
	return os.WriteFile(path, svgWrap("Blind boundaries vs "+kind, b.String()), 0644)
}
func plotCDF(path string, r runResult) error {
	colors := map[string]string{"folio": "#2563eb", "paragraph": "#059669", "currier": "#dc2626", "hand": "#7c3aed"}
	var b strings.Builder
	b.WriteString(`<path d="M50 40V250H960" fill="none" stroke="#333"/>`)
	blind := []int{}
	for _, v := range r.Stable {
		if v.Support >= 4 {
			blind = append(blind, v.Position)
		}
	}
	i := 0
	for _, kind := range []string{"folio", "paragraph", "currier", "hand"} {
		ds := distances(blind, positions(r.References[kind]))
		sort.Ints(ds)
		if len(ds) == 0 {
			continue
		}
		fmt.Fprintf(&b, `<polyline fill="none" stroke="%s" stroke-width="2" points="`, colors[kind])
		for j, d := range ds {
			x := 50 + 910*float64(min(d, 500))/500
			y := 250 - 200*float64(j+1)/float64(len(ds))
			fmt.Fprintf(&b, "%.1f,%.1f ", x, y)
		}
		b.WriteString(`"/>`)
		fmt.Fprintf(&b, `<text x="%d" y="285" fill="%s">%s</text>`, 60+i*150, colors[kind], kind)
		i++
	}
	return os.WriteFile(path, svgWrap("Nearest-boundary distance CDF (support ≥4/5)", b.String()), 0644)
}
func plotNMI(path, kind string, a []Association) error {
	var b strings.Builder
	b.WriteString(`<path d="M50 40V250H960" fill="none" stroke="#333"/>`)
	colors := map[string]string{"hierarchical": "#2563eb", "k_medoids": "#059669", "contiguous_segmentation": "#dc2626"}
	for method, color := range colors {
		fmt.Fprintf(&b, `<polyline fill="none" stroke="%s" stroke-width="2" points="`, color)
		for _, x := range a {
			if x.Metadata == kind && x.Subset == "all" && x.WindowSize == 200 && x.Method == method {
				px := 50 + float64(x.K-2)*70
				py := 250 - 200*x.NMI
				fmt.Fprintf(&b, "%.1f,%.1f ", px, py)
			}
		}
		b.WriteString(`"/>`)
	}
	return os.WriteFile(path, svgWrap("Frozen K sweep NMI — "+kind, b.String()), 0644)
}
func plotHeatmap(path, kind string, a []Association) error {
	var b strings.Builder
	i := 0
	for _, x := range a {
		if x.Metadata != kind || x.Subset != "all" || x.WindowSize != 200 {
			continue
		}
		shade := int(255 * (1 - minFloat(1, x.NMI)))
		fmt.Fprintf(&b, `<rect x="%d" y="70" width="28" height="120" fill="rgb(255,%d,%d)"/><text x="%d" y="215" transform="rotate(60 %d 215)">%s K%d</text>`, 30+i*30, shade, shade, 30+i*30, 30+i*30, x.Method, x.K)
		i++
	}
	return os.WriteFile(path, svgWrap("Regime/"+kind+" association heatmap (NMI)", b.String()), 0644)
}
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
