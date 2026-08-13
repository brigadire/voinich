package propertytrajectory

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
	if e = writeYAML(filepath.Join(c.OutputDir, "property_trajectory_pairs.yaml"), a.Out); e != nil {
		return e
	}
	if e = writeTop(filepath.Join(c.OutputDir, "property_trajectory_top.tsv"), a.Out.Pairs); e != nil {
		return e
	}
	if e = writeProperties(filepath.Join(c.OutputDir, "property_trajectory_properties.tsv"), a.Tokens, a.Out.Pairs); e != nil {
		return e
	}
	if e = writeControls(filepath.Join(c.OutputDir, "property_trajectory_controls.tsv"), a); e != nil {
		return e
	}
	if e = writeReport(filepath.Join(c.OutputDir, "property_trajectory_report.md"), a.Out); e != nil {
		return e
	}
	for _, x := range a.Out.Pairs {
		if plotPair(x.TokenA, x.TokenB) {
			if e = writePlot(filepath.Join(c.OutputDir, "plots", "property_trajectory_"+safe(x.TokenA)+"_"+safe(x.TokenB)+".svg"), x); e != nil {
				return e
			}
		}
	}
	p.update(1, 1, "Writing results")
	fmt.Printf("Property trajectories analyzed for %d pairs; results written to %s\n", len(a.Out.Pairs), c.OutputDir)
	return nil
}
func writeYAML(path string, v any) error {
	b, e := yaml.Marshal(v)
	if e != nil {
		return e
	}
	return os.WriteFile(path, b, 0o644)
}
func f(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }
func writeTop(path string, x []PairResult) error {
	w, e := os.Create(path)
	if e != nil {
		return e
	}
	defer w.Close()
	b := bufio.NewWriter(w)
	defer b.Flush()
	fmt.Fprintln(b, "rank\ttoken_a\ttoken_b\tcount_a\tcount_b\tcosine_1_5\tcosine_6_10\tcosine_11_20\tmatched_percentile\trandom_percentile")
	y := append([]PairResult(nil), x...)
	sort.SliceStable(y, func(i, j int) bool { return y[i].Summary.Cosine1To5 > y[j].Summary.Cosine1To5 })
	for i, p := range y {
		fmt.Fprintf(b, "%d\t%s\t%s\t%d\t%d\t%s\t%s\t%s\t%s\t%s\n", i+1, p.TokenA, p.TokenB, p.CountA, p.CountB, f(p.Summary.Cosine1To5), f(p.Summary.Cosine6To10), f(p.Summary.Cosine11To20), f(p.Summary.MatchedPercentile), f(p.Summary.RandomPercentile))
	}
	return nil
}
func writeProperties(path string, x []TokenProperties, pairs []PairResult) error {
	w, e := os.Create(path)
	if e != nil {
		return e
	}
	defer w.Close()
	b := bufio.NewWriter(w)
	defer b.Flush()
	fmt.Fprintln(b, "record_type\ttoken\tcount\ttoken_a\ttoken_b\tdistance\tproperty\traw_value\tnormalized_value\tmean_a\tmean_b\tdelta\traw_mean_a\traw_mean_b\ttrajectory_correlation\tmean_absolute_difference")
	for _, t := range x {
		names := make([]string, 0, len(t.Properties))
		for n := range t.Properties {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			v := t.Properties[n]
			fmt.Fprintf(b, "token_property\t%s\t%d\t\t\t\t%s\t%s\t%s\t\t\t\t\t\t\t\n", t.Token, t.Count, n, f(v.Raw), f(v.Normalized))
		}
	}
	for _, p := range pairs {
		ranking := map[string]PropertyRanking{}
		for _, r := range append(append([]PropertyRanking(nil), p.Summary.StrongestMatching...), p.Summary.StrongestDiffering...) {
			ranking[r.Property] = r
		}
		// Compute rankings for properties outside the displayed top/bottom sets as well.
		for name := range p.DistanceProfiles[0].Properties {
			var aVals, bVals []float64
			mad := 0.
			for _, d := range p.DistanceProfiles {
				q := d.Properties[name]
				aVals = append(aVals, q.MeanA)
				bVals = append(bVals, q.MeanB)
				mad += abs(q.Delta)
			}
			ranking[name] = PropertyRanking{name, pearson(aVals, bVals), mad / float64(len(p.DistanceProfiles))}
		}
		for _, d := range p.DistanceProfiles {
			names := make([]string, 0, len(d.Properties))
			for n := range d.Properties {
				names = append(names, n)
			}
			sort.Strings(names)
			for _, n := range names {
				q, r := d.Properties[n], ranking[n]
				fmt.Fprintf(b, "distance_property\t\t\t%s\t%s\t%d\t%s\t\t\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", p.TokenA, p.TokenB, d.Distance, n, f(q.MeanA), f(q.MeanB), f(q.Delta), f(q.RawMeanA), f(q.RawMeanB), f(r.TrajectoryCorrelation), f(r.MeanAbsoluteDifference))
			}
		}
	}
	return nil
}
func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
func writeControls(path string, a analysis) error {
	w, e := os.Create(path)
	if e != nil {
		return e
	}
	defer w.Close()
	b := bufio.NewWriter(w)
	defer b.Flush()
	fmt.Fprintln(b, "record_type\ttoken_a\ttoken_b\tcontrol\tdistance\tvalue\tmedian\tp90\tp95\tp99\tpercentile")
	for _, p := range a.Out.Pairs {
		for _, v := range a.Matched[pair{p.TokenA, p.TokenB}] {
			fmt.Fprintf(b, "matched\t%s\t%s\tfrequency_matched\t\t%s\t\t\t\t\t%s\n", p.TokenA, p.TokenB, f(v), f(p.Summary.MatchedPercentile))
		}
		for _, v := range a.Random[pair{p.TokenA, p.TokenB}] {
			fmt.Fprintf(b, "random_pair\t%s\t%s\tfrequency_matched_random\t\t%s\t\t\t\t\t%s\n", p.TokenA, p.TokenB, f(v), f(p.Summary.RandomPercentile))
		}
	}
	for _, x := range a.Out.RandomBaselines {
		fmt.Fprintf(b, "random_summary\t\t\t%s\t%d\t\t%s\t%s\t%s\t%s\t\n", x.Scope, x.Distance, f(x.Median), f(x.P90), f(x.P95), f(x.P99))
	}
	for _, x := range a.Out.Shuffles {
		fmt.Fprintf(b, "shuffle\t%s\t%s\t%s\t\t%s\t\t\t\t\t\n", x.PairA, x.PairB, x.Mode, f(x.MeanCosine1To5))
	}
	return nil
}
func writeReport(path string, o Output) error {
	var b strings.Builder
	b.WriteString("# Property-trajectory analysis\n\nThis analysis compares exact-distance trajectories of formal, intrinsic properties of subsequent tokens. It uses neither token classes, smoothing, nor structural projection. Rare subsequent tokens below the configured threshold are excluded and counted in every distance profile.\n\n## Main results\n\n| Pair | cosine 1–5 | 6–10 | 11–20 | matched percentile | random percentile |\n|---|---:|---:|---:|---:|---:|\n")
	for _, p := range o.Pairs {
		fmt.Fprintf(&b, "| `%s` / `%s` | %.4f | %.4f | %.4f | P%.1f | P%.1f |\n", p.TokenA, p.TokenB, p.Summary.Cosine1To5, p.Summary.Cosine6To10, p.Summary.Cosine11To20, p.Summary.MatchedPercentile, p.Summary.RandomPercentile)
	}
	target := meanPairCos(o.Pairs)
	matched, random := 0., 0.
	for _, p := range o.Pairs {
		matched += p.Summary.MatchedPercentile
		random += p.Summary.RandomPercentile
	}
	if len(o.Pairs) > 0 {
		matched /= float64(len(o.Pairs))
		random /= float64(len(o.Pairs))
	}
	shuffleValues := map[string][]float64{}
	for _, x := range o.Shuffles {
		shuffleValues[x.Mode] = append(shuffleValues[x.Mode], x.MeanCosine1To5)
	}
	globalShuffle, lineShuffle := mean(shuffleValues["global"]), mean(shuffleValues["line-preserving"])
	b.WriteString("\n## Critical null comparison\n\n")
	if matched >= 95 && random >= 95 && target > globalShuffle && target > lineShuffle {
		b.WriteString("The target set clears the joint diagnostic criterion: mean matched and random percentiles are at least 95, and observed similarity exceeds both shuffle controls. This is a diagnostic positive result, not evidence for a state machine.\n")
	} else {
		b.WriteString("The target set does not clear the joint diagnostic criterion (P95 against both matched and random controls, plus a decrease under both shuffles). The property-trajectory hypothesis is therefore not supported as a general explanation by this run.\n")
	}
	fmt.Fprintf(&b, " Mean target cosine 1–5 is %.4f; mean matched percentile is %.1f, mean random percentile is %.1f, global-shuffle cosine is %.4f, and line-preserving-shuffle cosine is %.4f. Pair-level values remain in the controls TSV and YAML.\n", target, matched, random, globalShuffle, lineShuffle)
	b.WriteString("\n## What drives the score\n\nEach pair stores frequency-only, graphemic-form-only, position-only, context-complexity-only, structural-centrality-only, all-properties, and five leave-one-group-out scores. Per-property normalized deltas, trajectory correlations, and the strongest matching/differing rankings remain inspectable rather than being replaced by one score.\n\n## Limits\n\nCosine in a globally z-scored property space can be negative. Empirical percentiles are deterministic diagnostics, not independent-sample p-values. Structural inputs contribute only each subsequent token's centrality statistics; no pair projection or family membership is used. No semantic labels or latent states are inferred.\n")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
func meanPairCos(x []PairResult) float64 {
	v := make([]float64, len(x))
	for i, p := range x {
		v[i] = p.Summary.Cosine1To5
	}
	return mean(v)
}
func plotPair(a, b string) bool {
	k := a + "/" + b
	if b < a {
		k = b + "/" + a
	}
	return k == "chedy/qokeey" || k == "chol/daiin" || k == "or/s" || k == "r/s" || k == "ol/y"
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
func writePlot(path string, p PairResult) error {
	left, top, w, h := 60., 45., 720., 330.
	n := len(p.DistanceProfiles)
	x := func(i int) float64 {
		if n < 2 {
			return left
		}
		return left + float64(i)*w/float64(n-1)
	}
	y := func(v float64) float64 { return top + (1-(v+1)/2)*h }
	var b strings.Builder
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="840" height="460" viewBox="0 0 840 460"><rect width="100%" height="100%" fill="white"/><style>text{font:12px sans-serif}.title{font:bold 16px sans-serif}.grid{stroke:#ddd}</style>`)
	fmt.Fprintf(&b, "<text class=\"title\" x=\"20\" y=\"25\">%s / %s — property trajectory</text>", esc(p.TokenA), esc(p.TokenB))
	for i := -1; i <= 1; i++ {
		yy := y(float64(i))
		fmt.Fprintf(&b, "<line class=\"grid\" x1=\"%.0f\" y1=\"%.1f\" x2=\"%.0f\" y2=\"%.1f\"/><text x=\"25\" y=\"%.1f\">%d</text>", left, yy, left+w, yy, yy+4, i)
	}
	b.WriteString(`<polyline fill="none" stroke="#1769aa" stroke-width="3" points="`)
	for i, d := range p.DistanceProfiles {
		fmt.Fprintf(&b, "%.1f,%.1f ", x(i), y(d.CosineSimilarity))
	}
	b.WriteString(`"/>`)
	for i, d := range p.DistanceProfiles {
		if d.Distance == 1 || d.Distance%5 == 0 {
			fmt.Fprintf(&b, "<text x=\"%.1f\" y=\"400\">+%d</text>", x(i)-7, d.Distance)
		}
	}
	b.WriteString(`<text x="310" y="435">cosine similarity in normalized property space</text></svg>`)
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
