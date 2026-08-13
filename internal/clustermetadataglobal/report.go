package clustermetadataglobal

import (
	"fmt"
	"os"
	"strings"
)

const (
	reportBeginMarker = "<!-- BEGIN GLOBAL MULTIPLE-COMPARISON CORRECTION (cluster-metadata-global) -->"
	reportEndMarker   = "<!-- END GLOBAL MULTIPLE-COMPARISON CORRECTION (cluster-metadata-global) -->"
	// significanceAlpha is the conventional threshold used only to phrase the
	// interpretation sentences; every raw empirical_p is reported regardless.
	significanceAlpha = 0.05
)

// updateValidationReport adds or replaces the "Global multiple-comparison
// correction" section of the existing blind metadata validation report
// (task18 section 20), without touching any other section. If the report
// does not exist yet, a new file containing only this section is created.
func updateValidationReport(path string, c Config, series map[string]*StatSeries) error {
	existing := ""
	if b, err := os.ReadFile(path); err == nil {
		existing = string(b)
	} else if !os.IsNotExist(err) {
		return err
	}
	section := reportBeginMarker + "\n" + buildReportSection(c, series) + "\n" + reportEndMarker + "\n"
	begin := strings.Index(existing, reportBeginMarker)
	var out string
	if begin < 0 {
		out = strings.TrimRight(existing, "\n")
		if out != "" {
			out += "\n\n"
		}
		out += section
	} else {
		end := strings.Index(existing, reportEndMarker)
		if end < 0 {
			return fmt.Errorf("report %s has a begin marker but no matching end marker; refusing to overwrite ambiguously", path)
		}
		end += len(reportEndMarker)
		for end < len(existing) && existing[end] == '\n' {
			end++
		}
		out = existing[:begin] + section + existing[end:]
	}
	return os.WriteFile(path, []byte(out), 0644)
}

func buildReportSection(c Config, series map[string]*StatSeries) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Global multiple-comparison correction\n\n")
	fmt.Fprintf(&b, "This section corrects the observed association between frozen blind distributional regimes and Currier/hand metadata for every choice that was already available before metadata was consulted: window size in {%s}, method in {%s} and K in %d..%d. Discovery (windows, clustering, boundaries) is unchanged from `global_distributional_*`; only metadata labels are permuted, using %d block-aware permutations (seed %d) that preserve contiguous metadata block lengths and the unknown-token mask exactly, so the same set of valid windows is compared for every observed and null statistic. Empirical p-values use `(exceedances + 1) / (permutations + 1)` and are never reported as exactly zero.\n\n",
		joinInts(WindowSizes), strings.Join(Methods, ", "), KMin, KMax, c.Permutations, c.Seed)

	for _, kind := range Kinds {
		title := "Currier"
		if kind == "hand" {
			title = "Hand"
		}
		fmt.Fprintf(&b, "### %s\n\n", title)
		fmt.Fprintf(&b, "Per-method correction (max NMI over window size x K, frozen K=%d..%d):\n\n", KMin, KMax)
		fmt.Fprintf(&b, "| Method | Observed NMI | window | K | null mean | null P95 | null P99 | empirical p |\n|---|---:|---:|---:|---:|---:|---:|---:|\n")
		for _, m := range Methods {
			s := series[seriesKey(kind, "primary", "NMI", m)]
			e := summarize(s)
			fmt.Fprintf(&b, "| %s | %.3f | %d | %d | %.3f | %.3f | %.3f | %s |\n", m, e.Observed, s.ObservedWindow, s.ObservedK, e.NullMean, e.NullP95, e.NullP99, formatP(e.EmpiricalP))
		}
		gs := series[seriesKey(kind, "primary", "NMI", "global")]
		ge := summarize(gs)
		fmt.Fprintf(&b, "\nGlobal correction (max NMI over window size x method x K, the complete frozen search space): observed %.3f at window=%d, method=%s, K=%d; null mean %.3f, P95 %.3f, P99 %.3f, empirical p %s.\n\n",
			ge.Observed, gs.ObservedWindow, gs.ObservedMethod, gs.ObservedK, ge.NullMean, ge.NullP95, ge.NullP99, formatP(ge.EmpiricalP))

		fmt.Fprintf(&b, "Scale-persistence (mean and minimum of the five prespecified scale-specific max-over-K values, no maximum selection across scales):\n\n")
		fmt.Fprintf(&b, "| Method | mean-across-scales | empirical p | min-across-scales | empirical p |\n|---|---:|---:|---:|---:|\n")
		for _, m := range Methods {
			meanS := series[seriesKey(kind, "primary", "NMI", m+"/persistence_mean")]
			minS := series[seriesKey(kind, "primary", "NMI", m+"/persistence_min")]
			me, mn := summarize(meanS), summarize(minS)
			fmt.Fprintf(&b, "| %s | %.3f | %s | %.3f | %s |\n", m, me.Observed, formatP(me.EmpiricalP), mn.Observed, formatP(mn.EmpiricalP))
		}
		b.WriteString("\n" + interpretation(kind, series) + "\n\n")
	}

	b.WriteString("### Purity sensitivity analysis\n\n")
	b.WriteString("The same global correction, repeated over purity >= 0.8 and purity >= 0.9 mixed-window subsets. These thresholds were fixed in advance and are reported as **sensitivity analysis**, not primary evidence; the primary test above always uses all windows with a known majority metadata label.\n\n")
	b.WriteString("| Metadata | Scope | Method | Observed max NMI | empirical p | Global max NMI | empirical p |\n|---|---|---|---:|---:|---:|---:|\n")
	for _, kind := range Kinds {
		for _, sc := range Scopes {
			if sc.Name == "primary" {
				continue
			}
			gs := series[seriesKey(kind, sc.Name, "NMI", "global")]
			ge := summarize(gs)
			for _, m := range Methods {
				s := series[seriesKey(kind, sc.Name, "NMI", m)]
				e := summarize(s)
				fmt.Fprintf(&b, "| %s | %s | %s | %.3f | %s | %.3f | %s |\n", kind, sc.Name, m, e.Observed, formatP(e.EmpiricalP), ge.Observed, formatP(ge.EmpiricalP))
			}
		}
	}
	b.WriteString("\nSecondary metric: the same three primary statistics (per-method max, global max, scale persistence) were also computed for Adjusted Rand Index; see `cluster_metadata_global_summary.tsv` and `cluster_metadata_scale_persistence.tsv` (metric=ARI) for full values, at every prespecified scope.\n")
	return b.String()
}

func interpretation(kind string, series map[string]*StatSeries) string {
	kmedoids := series[seriesKey(kind, "primary", "NMI", "k_medoids")]
	global := series[seriesKey(kind, "primary", "NMI", "global")]
	persistMean := series[seriesKey(kind, "primary", "NMI", "k_medoids/persistence_mean")]
	kmedoidsP := summarize(kmedoids).EmpiricalP
	globalP := summarize(global).EmpiricalP
	persistP := summarize(persistMean).EmpiricalP
	var lines []string
	if kmedoidsP < significanceAlpha {
		lines = append(lines, fmt.Sprintf("k-medoids remains significant after correcting for window size and K within the frozen k-medoids search space (empirical p=%s): association is not explained by post-hoc selection of window size or K within that method.", formatP(kmedoidsP)))
	} else {
		lines = append(lines, fmt.Sprintf("k-medoids does not remain significant after correcting for window size and K within the frozen k-medoids search space (empirical p=%s).", formatP(kmedoidsP)))
	}
	if globalP < significanceAlpha {
		lines = append(lines, fmt.Sprintf("The global maximum remains significant after correcting across the complete frozen window size x method x K search space (empirical p=%s): the association survives correction across every clustering choice considered before metadata was consulted.", formatP(globalP)))
	} else {
		lines = append(lines, fmt.Sprintf("The global maximum does not remain significant after correcting across the complete frozen search space (empirical p=%s).", formatP(globalP)))
	}
	if (kmedoidsP < significanceAlpha || globalP < significanceAlpha) && persistP >= significanceAlpha {
		lines = append(lines, "The max statistic is significant while the scale-persistence statistic is not: the association may be scale-specific rather than reproduced across the five prespecified window sizes.")
	} else if persistP < significanceAlpha {
		lines = append(lines, fmt.Sprintf("The scale-persistence statistic is also significant (empirical p=%s): the association is reproduced across multiple prespecified scales, not concentrated in a single window size.", formatP(persistP)))
	}
	return strings.Join(lines, " ")
}

func formatP(p float64) string {
	if p < 0.0001 {
		return "<0.0001"
	}
	return fmt.Sprintf("%.4f", p)
}

func joinInts(x []int) string {
	parts := make([]string, len(x))
	for i, v := range x {
		parts[i] = fmt.Sprint(v)
	}
	return strings.Join(parts, ", ")
}
