package distancecontext

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
	a, e := analyze(c)
	if e != nil {
		return e
	}
	if e = os.MkdirAll(filepath.Join(c.OutputDir, "plots"), 0o755); e != nil {
		return e
	}
	files := []struct {
		name string
		v    any
	}{{"distance_context_pairs.yaml", a.Out}, {"sequence_context_pairs.yaml", map[string]any{"methodology": "Exact 2-token and 3-token suffix distributions; the anchor token is not part of the suffix.", "pairs": a.Sequences}}, {"family_distance_profiles.yaml", map[string]any{"methodology": "Cohesion is mean pairwise JS similarity; percentile uses 200 deterministic frequency-matched token groups of equal size.", "families": a.Families}}}
	for _, x := range files {
		if e = writeYAML(filepath.Join(c.OutputDir, x.name), x.v); e != nil {
			return e
		}
	}
	if e = writeTop(filepath.Join(c.OutputDir, "distance_context_top.tsv"), a.Out.Pairs); e != nil {
		return e
	}
	if e = writeControls(filepath.Join(c.OutputDir, "distance_context_controls.tsv"), a.Controls); e != nil {
		return e
	}
	if e = writeReport(filepath.Join(c.OutputDir, "distance_context_report.md"), a); e != nil {
		return e
	}
	for _, p := range a.Out.Pairs {
		if e = writePairPlot(filepath.Join(c.OutputDir, "plots", safe(p.TokenA)+"_"+safe(p.TokenB)+"_context_decay.svg"), p, a.Out.Baseline, controlMedians(p, a.Controls)); e != nil {
			return e
		}
	}
	for _, f := range a.Families {
		if e = writeFamilyPlot(filepath.Join(c.OutputDir, "plots", fmt.Sprintf("family_%d_cohesion.svg", f.ID)), f); e != nil {
			return e
		}
	}
	fmt.Printf("Analyzed %d corpus tokens and %d target pairs; reports written to %s\n", a.Out.TokenCount, a.Out.PairCount, c.OutputDir)
	return nil
}
func writeYAML(path string, v any) error {
	b, e := yaml.Marshal(v)
	if e != nil {
		return e
	}
	return os.WriteFile(path, b, 0o644)
}
func ff(x float64) string { return strconv.FormatFloat(x, 'g', -1, 64) }
func writeTop(path string, p []PairResult) error {
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	fmt.Fprintln(w, "rank\ttoken_a\ttoken_b\tcount_a\tcount_b\tpersistence_1_5\tmean_js_1_5\tmean_js_6_10\tmean_js_11_20\tjs_d1\tpercentile_d1\tjs_d2\tpercentile_d2\tjs_d3\tpercentile_d3\tjs_d5\tpercentile_d5\tjs_d10\tpercentile_d10\tjs_d20\tpercentile_d20")
	for i, x := range p {
		cell := func(d int) (float64, float64) {
			if d <= len(x.Right) {
				return x.Right[d-1].JSSimilarity, x.Right[d-1].BaselinePercentile
			}
			return 0, 0
		}
		a1, p1 := cell(1)
		a2, p2 := cell(2)
		a3, p3 := cell(3)
		a5, p5 := cell(5)
		a10, p10 := cell(10)
		a20, p20 := cell(20)
		fmt.Fprintf(w, "%d\t%s\t%s\t%d\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", i+1, x.TokenA, x.TokenB, x.CountA, x.CountB, ff(x.RightSummary.Persistence1To5), ff(x.RightSummary.Mean1To5), ff(x.RightSummary.Mean6To10), ff(x.RightSummary.Mean11To20), ff(a1), ff(p1), ff(a2), ff(p2), ff(a3), ff(p3), ff(a5), ff(p5), ff(a10), ff(p10), ff(a20), ff(p20))
	}
	return nil
}
func writeControls(path string, c []ControlResult) error {
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	fmt.Fprintln(w, "target_a\ttarget_b\tcontrol_rank\tcontrol_a\tcontrol_b\tdistance\tjs_similarity\tweighted_overlap\tjaccard_support_overlap\tobservations_a\tobservations_b\treliability")
	for _, x := range c {
		for _, m := range x.Profile {
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%d\t%s\t%s\t%s\t%d\t%d\t%s\n", x.TargetA, x.TargetB, x.Rank, x.ControlA, x.ControlB, m.Distance, ff(m.JSSimilarity), ff(m.WeightedOverlap), ff(m.Jaccard), m.ObservationsA, m.ObservationsB, ff(m.Reliability))
		}
	}
	return nil
}

func writeReport(path string, a analysis) error {
	var b strings.Builder
	b.WriteString("# Distance-specific context analysis\n\nThis report describes formal token-sequence distributions and does not assign meanings to tokens. The main analysis treats all non-empty lines as one continuous sequence; pages are not used. Physical lines are retained only for the line-bounded control.\n\n## Parameters and corpus\n\n")
	fmt.Fprintf(&b, "Corpus tokens: %d; target pairs: %d; maximum exact distance: %v; minimum observations: %v; requested primary mode: `%v`. Similarity at every distance is computed separately, never from a pooled window. Frequency-matched baseline pairs have unordered counts within a factor of 2. Effective support is `exp(Shannon entropy)` using natural-log units.\n\n", a.Out.TokenCount, a.Out.PairCount, a.Out.Parameters["max_distance"], a.Out.Parameters["min_observations"], a.Out.Parameters["primary_mode"])
	b.WriteString("## Right-context baseline\n\n| Distance | Pairs | Median JS | P90 | P95 |\n|---:|---:|---:|---:|---:|\n")
	for _, x := range a.Out.Baseline {
		fmt.Fprintf(&b, "| +%d | %d | %.4f | %.4f | %.4f |\n", x.Distance, x.Pairs, x.Median, x.P90, x.P95)
	}
	b.WriteString("\n## Long-range structural persistence ranking\n\n`persistence_1_5` is the transparent mean of baseline percentile ranks at exact distances +1 through +5. It is a ranking aid, not a replacement for the profiles.\n\n| Pair | Persistence 1–5 | Mean JS 1–5 | 6–10 | 11–20 |\n|---|---:|---:|---:|---:|\n")
	for _, x := range a.Out.Pairs {
		fmt.Fprintf(&b, "| `%s` / `%s` | %.2f | %.4f | %.4f | %.4f |\n", x.TokenA, x.TokenB, x.RightSummary.Persistence1To5, x.RightSummary.Mean1To5, x.RightSummary.Mean6To10, x.RightSummary.Mean11To20)
	}
	b.WriteString("\n## Target profiles, directionality, and boundary sensitivity\n\n")
	for _, x := range a.Out.Pairs {
		fmt.Fprintf(&b, "### `%s` / `%s`\n\nCounts: %d/%d. Right mean JS 1–5: %.4f; left mean JS 1–5: %.4f. Sequence suffix similarities n=2/n=3: %.4f/%.4f.\n\n| d | Right JS | Percentile | Left JS | Continuous | Line-bounded | Difference | Continuous obs A/B | Line obs A/B | Reliable C/L |\n|---:|---:|---:|---:|---:|---:|---:|---:|---:|:---:|\n", x.TokenA, x.TokenB, x.CountA, x.CountB, x.RightSummary.Mean1To5, x.LeftSummary.Mean1To5, seqJS(a.Sequences, x.TokenA, x.TokenB, 0), seqJS(a.Sequences, x.TokenA, x.TokenB, 1))
		for i, m := range x.Right {
			lm := x.LineBoundedRight[i]
			fmt.Fprintf(&b, "| %d | %.4f | P%.1f | %.4f | %.4f | %.4f | %+.4f | %d/%d | %d/%d | %t/%t |\n", m.Distance, m.JSSimilarity, m.BaselinePercentile, x.Left[i].JSSimilarity, m.JSSimilarity, lm.JSSimilarity, x.BoundarySensitivity[i].Difference, m.ObservationsA, m.ObservationsB, lm.ObservationsA, lm.ObservationsB, m.Reliable, lm.Reliable)
		}
		b.WriteString("\n")
	}
	b.WriteString("## Matched negative controls\n\n| Target | Controls | Target mean JS 1–5 | Control median mean JS 1–5 |\n|---|---:|---:|---:|\n")
	for _, x := range a.Out.Pairs {
		v, n := controlMean(x, a.Controls)
		fmt.Fprintf(&b, "| `%s` / `%s` | %d | %.4f | %.4f |\n", x.TokenA, x.TokenB, n, x.RightSummary.Mean1To5, v)
	}
	b.WriteString("\n## Family cohesion\n\nMatrices and distance-specific medoids are in `family_distance_profiles.yaml`. Random controls are deterministic groups of equal size whose member frequencies match the corresponding family member within ×2.\n\n| Family | d | Cohesion | Matched percentile | Medoid |\n|---:|---:|---:|---:|---|\n")
	for _, f := range a.Families {
		for _, x := range f.Profiles {
			fmt.Fprintf(&b, "| %d | %d | %.4f | P%.1f | `%s` |\n", f.ID, x.Distance, x.Cohesion, x.Percentile, x.Medoid)
		}
	}
	b.WriteString("\n## Sequence-context signatures\n\nThe 2- and 3-token suffix distributions are compared separately with JS similarity, weighted overlap, support Jaccard, observation counts, effective support, and reliability in `sequence_context_pairs.yaml`.\n\n## Limitations\n\nSimilarity at greater distance is descriptive context-similarity persistence, not physical decay. Corpus positions and pair comparisons are dependent. Frequency matching controls only a major sampling confound and is not a causal null model. Support overlap is sensitive to rare observations. A low-reliability value is retained but must not be treated as equivalent to a frequent comparison. Line-bounded results remove cross-line observations and therefore can change both distributions and sample sizes. Sequence suffixes become sparse quickly; only lengths 2 and 3 are used. No semantic interpretation is made.\n")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
func seqJS(s []SequencePair, a, b string, i int) float64 {
	for _, x := range s {
		if x.TokenA == a && x.TokenB == b {
			return x.Continuous[i].JSSimilarity
		}
	}
	return 0
}
func controlMean(p PairResult, c []ControlResult) (float64, int) {
	var x []float64
	for _, v := range c {
		if canon(v.TargetA, v.TargetB) == canon(p.TokenA, p.TokenB) {
			q := make([]float64, len(v.Profile))
			for i, m := range v.Profile {
				q[i] = m.JSSimilarity
			}
			x = append(x, avg(q, 1, 5))
		}
	}
	sort.Float64s(x)
	return quantile(x, .5), len(x)
}
func controlMedians(p PairResult, c []ControlResult) []float64 {
	out := make([]float64, len(p.Right))
	for d := range out {
		var x []float64
		for _, v := range c {
			if canon(v.TargetA, v.TargetB) == canon(p.TokenA, p.TokenB) && d < len(v.Profile) {
				x = append(x, v.Profile[d].JSSimilarity)
			}
		}
		sort.Float64s(x)
		out[d] = quantile(x, .5)
	}
	return out
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
func writePairPlot(path string, p PairResult, base []BaselineRow, controls []float64) error {
	left, top, pw, ph := 65., 45., 730., 390.
	x := func(d int) float64 {
		if len(p.Right) == 1 {
			return left
		}
		return left + float64(d-1)*pw/float64(len(p.Right)-1)
	}
	y := func(v float64) float64 { return top + (1-v)*ph }
	var b strings.Builder
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="840" height="500" viewBox="0 0 840 500"><rect width="100%" height="100%" fill="white"/><style>text{font:12px sans-serif}.title{font:bold 16px sans-serif}.grid{stroke:#ddd}.target{stroke:#1769aa;fill:none;stroke-width:3}.median{stroke:#555;fill:none;stroke-dasharray:5 4}.p95{stroke:#d95f02;fill:none;stroke-dasharray:7 4}.control{stroke:#2ca02c;fill:none;stroke-dasharray:2 4}</style>`)
	fmt.Fprintf(&b, "<text class=\"title\" x=\"20\" y=\"25\">%s / %s — right context-similarity profile</text>", esc(p.TokenA), esc(p.TokenB))
	for i := 0; i <= 5; i++ {
		v := float64(i) / 5
		fmt.Fprintf(&b, "<line class=\"grid\" x1=\"%.0f\" y1=\"%.1f\" x2=\"%.0f\" y2=\"%.1f\"/><text x=\"25\" y=\"%.1f\">%.1f</text>", left, y(v), left+pw, y(v), y(v)+4, v)
	}
	poly := func(class string, vals []float64) {
		b.WriteString("<polyline class=\"" + class + "\" points=\"")
		for i, v := range vals {
			fmt.Fprintf(&b, "%.1f,%.1f ", x(i+1), y(v))
		}
		b.WriteString("\"/>")
	}
	target, med, p95 := make([]float64, len(p.Right)), make([]float64, len(p.Right)), make([]float64, len(p.Right))
	for i, m := range p.Right {
		target[i] = m.JSSimilarity
		med[i] = base[i].Median
		p95[i] = base[i].P95
	}
	poly("target", target)
	poly("median", med)
	poly("p95", p95)
	poly("control", controls)
	for _, d := range []int{1, 2, 3, 5, 10, 20} {
		if d <= len(p.Right) {
			fmt.Fprintf(&b, "<text x=\"%.1f\" y=\"455\" text-anchor=\"middle\">%d</text>", x(d), d)
		}
	}
	b.WriteString(`<text x="680" y="58" fill="#1769aa">target</text><text x="680" y="75">baseline median</text><text x="680" y="92" fill="#d95f02">baseline P95</text><text x="680" y="109" fill="#2ca02c">control median</text><text x="400" y="482">exact token distance</text></svg>`)
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
func writeFamilyPlot(path string, f FamilyResult) error {
	if len(f.Profiles) == 0 {
		return nil
	}
	const left, top, pw, ph = 65., 45., 730., 390.
	x := func(d int) float64 {
		if len(f.Profiles) == 1 {
			return left
		}
		return left + float64(d-1)*pw/float64(len(f.Profiles)-1)
	}
	y := func(v float64) float64 { return top + (1-v)*ph }
	var b strings.Builder
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="840" height="500"><rect width="100%" height="100%" fill="white"/><style>text{font:12px sans-serif}.line{stroke:#6a3d9a;fill:none;stroke-width:3}.grid{stroke:#ddd}</style>`)
	fmt.Fprintf(&b, "<text x=\"20\" y=\"25\">Family %d — right-context cohesion</text>", f.ID)
	for i := 0; i <= 5; i++ {
		v := float64(i) / 5
		fmt.Fprintf(&b, "<line class=\"grid\" x1=\"%.0f\" y1=\"%.1f\" x2=\"%.0f\" y2=\"%.1f\"/>", left, y(v), left+pw, y(v))
	}
	b.WriteString(`<polyline class="line" points="`)
	for _, v := range f.Profiles {
		fmt.Fprintf(&b, "%.1f,%.1f ", x(v.Distance), y(v.Cohesion))
	}
	b.WriteString(`"/><text x="400" y="482">exact token distance</text></svg>`)
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
