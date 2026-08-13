package clustermetadataglobal

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

func f(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }

func tsv(path, header string, rows []string) error {
	fh, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fh.Close()
	w := bufio.NewWriter(fh)
	defer w.Flush()
	fmt.Fprintln(w, header)
	for _, r := range rows {
		fmt.Fprintln(w, r)
	}
	return nil
}

func writeYAML(path string, v any) error {
	b, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

// maxMethodScopes lists, in output order, the method_scope values for the
// two primary/global "max" statistics (A and B).
func maxMethodScopes() []string {
	out := append([]string(nil), Methods...)
	return append(out, "global")
}

func writeSummary(path string, series map[string]*StatSeries) error {
	rows := []string{}
	for _, kind := range Kinds {
		for _, sc := range Scopes {
			for _, metric := range []string{"NMI", "ARI"} {
				for _, ms := range maxMethodScopes() {
					s := series[seriesKey(kind, sc.Name, metric, ms)]
					e := summarize(s)
					rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s\t%d\t%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s\t%d\t%s",
						kind, ms, metric, f(e.Observed), s.ObservedWindow, s.ObservedMethod, s.ObservedK,
						f(e.NullMean), f(e.NullMedian), f(e.NullSD), f(e.NullP95), f(e.NullP99), f(e.NullMax),
						e.Exceedances, f(e.EmpiricalP), e.Permutations, sc.Name))
				}
			}
		}
	}
	return tsv(path, "metadata\tmethod_scope\tmetric\tobserved\tobserved_window\tobserved_method\tobserved_k\tnull_mean\tnull_median\tnull_sd\tnull_p95\tnull_p99\tnull_max\texceedances\tempirical_p\tpermutations\tscope", rows)
}

func writeScalePersistence(path string, series map[string]*StatSeries, byWindowVector map[string][]float64) error {
	rows := []string{}
	for _, kind := range Kinds {
		for _, sc := range Scopes {
			for _, metric := range []string{"NMI", "ARI"} {
				for _, m := range Methods {
					vecKey := kind + "|" + sc.Name + "|" + metric + "|" + m
					vec := byWindowVector[vecKey]
					windowCols := make([]string, len(WindowSizes))
					for i := range WindowSizes {
						if i < len(vec) {
							windowCols[i] = f(vec[i])
						} else {
							windowCols[i] = f(0)
						}
					}
					for _, statistic := range []string{"mean_across_scales", "min_across_scales"} {
						ms := m + "/persistence_mean"
						if statistic == "min_across_scales" {
							ms = m + "/persistence_min"
						}
						s := series[seriesKey(kind, sc.Name, metric, ms)]
						e := summarize(s)
						rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s\t%d",
							kind, m, metric, sc.Name, statistic, strings.Join(windowCols, "\t"),
							f(e.Observed), f(e.NullMean), f(e.NullMedian), f(e.NullSD), f(e.NullP95), f(e.NullP99), f(e.NullMax),
							e.Exceedances, f(e.EmpiricalP), e.Permutations))
					}
				}
			}
		}
	}
	header := "metadata\tmethod\tmetric\tscope\tstatistic\twindow_50_max_nmi\twindow_100_max_nmi\twindow_200_max_nmi\twindow_500_max_nmi\twindow_1000_max_nmi\tobserved\tnull_mean\tnull_median\tnull_sd\tnull_p95\tnull_p99\tnull_max\texceedances\tempirical_p\tpermutations"
	return tsv(path, header, rows)
}

func writePermutationsYAML(path string, c Config, series map[string]*StatSeries) error {
	observed := map[string]any{}
	null := map[string]any{}
	for _, kind := range Kinds {
		obK, nullK := map[string]any{}, map[string]any{}
		for _, sc := range Scopes {
			obS, nullS := map[string]any{}, map[string]any{}
			for _, metric := range []string{"NMI", "ARI"} {
				obM, nullM := map[string]any{}, map[string]any{}
				for _, ms := range allMethodScopes() {
					s := series[seriesKey(kind, sc.Name, metric, ms)]
					obM[ms] = map[string]any{"value": s.Observed, "window_size": s.ObservedWindow, "method": s.ObservedMethod, "k": s.ObservedK}
					nullM[ms] = s.Null
				}
				obS[metric] = obM
				nullS[metric] = nullM
			}
			obK[sc.Name] = obS
			nullK[sc.Name] = nullS
		}
		observed[kind] = obK
		null[kind] = nullK
	}
	doc := map[string]any{
		"seed":         c.Seed,
		"permutations": c.Permutations,
		"search_space": map[string]any{"window_size": WindowSizes, "method": Methods, "k": []int{KMin, KMax}},
		"method": "one block-aware permuted metadata realization per replicate and per metadata kind, shared unchanged across the entire frozen window_size x method x K search space; unknown-token mask and block lengths preserved; only labels among known contiguous metadata blocks are reassigned",
		"observed":     observed,
		"null":         null,
	}
	return writeYAML(path, doc)
}

func allMethodScopes() []string {
	out := append([]string(nil), maxMethodScopes()...)
	for _, m := range Methods {
		out = append(out, m+"/persistence_mean", m+"/persistence_min")
	}
	return out
}

func writePlots(dir string, series map[string]*StatSeries, byWindowVector map[string][]float64) error {
	for _, kind := range Kinds {
		s := series[seriesKey(kind, "primary", "NMI", "global")]
		if err := plotNullHistogram(filepath.Join(dir, "plots", kind+"_global_nmi_null.svg"), "Global max NMI null ("+kind+", window x method x K)", s); err != nil {
			return err
		}
		if err := plotScalePersistence(filepath.Join(dir, "plots", kind+"_nmi_scale_persistence.svg"), kind, byWindowVector); err != nil {
			return err
		}
	}
	return nil
}

func svgWrap(title, body string) []byte {
	return []byte(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="1000" height="320" viewBox="0 0 1000 320"><rect width="100%%" height="100%%" fill="white"/><style>text{font:12px sans-serif}.t{font:bold 16px sans-serif}</style><text class="t" x="20" y="25">%s</text>%s</svg>`, title, body))
}

func plotNullHistogram(path, title string, s *StatSeries) error {
	var b strings.Builder
	null := append([]float64(nil), s.Null...)
	sort.Float64s(null)
	bins := 30
	lo, hi := 0.0, 0.0
	if len(null) > 0 {
		lo, hi = null[0], null[len(null)-1]
	}
	if s.Observed > hi {
		hi = s.Observed
	}
	if hi <= lo {
		hi = lo + 1
	}
	counts := make([]int, bins)
	for _, v := range null {
		bin := int((v - lo) / (hi - lo) * float64(bins))
		if bin >= bins {
			bin = bins - 1
		}
		if bin < 0 {
			bin = 0
		}
		counts[bin]++
	}
	maxCount := 1
	for _, c := range counts {
		if c > maxCount {
			maxCount = c
		}
	}
	plotW, plotH, x0, y0 := 920.0, 220.0, 40.0, 260.0
	for i, c := range counts {
		bw := plotW / float64(bins)
		x := x0 + float64(i)*bw
		h := plotH * float64(c) / float64(maxCount)
		fmt.Fprintf(&b, `<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="#2563eb" opacity=".75"/>`, x, y0-h, bw-1, h)
	}
	ox := x0 + plotW*(s.Observed-lo)/(hi-lo)
	fmt.Fprintf(&b, `<line x1="%.2f" y1="%.2f" x2="%.2f" y2="40" stroke="#dc2626" stroke-width="2"/><text x="%.2f" y="55" fill="#dc2626">observed %.3f</text>`, ox, y0, ox, ox+4, s.Observed)
	fmt.Fprintf(&b, `<line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="#333"/>`, x0, y0, x0+plotW, y0)
	return os.WriteFile(path, svgWrap(title, b.String()), 0644)
}

func plotScalePersistence(path, kind string, byWindowVector map[string][]float64) error {
	var b strings.Builder
	colors := map[string]string{"contiguous_segmentation": "#dc2626", "hierarchical": "#2563eb", "k_medoids": "#059669"}
	b.WriteString(`<line x1="60" y1="260" x2="960" y2="260" stroke="#333"/><line x1="60" y1="40" x2="60" y2="260" stroke="#333"/>`)
	for i, ws := range WindowSizes {
		x := 60 + 900*float64(i)/float64(len(WindowSizes)-1)
		fmt.Fprintf(&b, `<text x="%.2f" y="275">%d</text>`, x-10, ws)
	}
	for _, m := range Methods {
		vec := byWindowVector[kind+"|primary|NMI|"+m]
		if len(vec) == 0 {
			continue
		}
		fmt.Fprintf(&b, `<polyline fill="none" stroke="%s" stroke-width="2" points="`, colors[m])
		for i, v := range vec {
			x := 60 + 900*float64(i)/float64(len(WindowSizes)-1)
			y := 260 - 200*v
			fmt.Fprintf(&b, "%.1f,%.1f ", x, y)
		}
		b.WriteString(`"/>`)
	}
	i := 0
	for _, m := range Methods {
		fmt.Fprintf(&b, `<text x="%d" y="300" fill="%s">%s</text>`, 60+i*300, colors[m], m)
		i++
	}
	return os.WriteFile(path, svgWrap("Scale persistence (max NMI over K, primary scope) — "+kind, b.String()), 0644)
}
