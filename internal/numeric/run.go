package numeric

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var controlKinds = []string{"C1_WITHIN_TOKEN_GLYPH_SHUFFLE", "C2_TOKEN_SHUFFLE_WITHIN_LINE", "C3_GLYPH_BIGRAM_MARKOV"}

func Run(cfg Config) error {
	if cfg.Replicates <= 0 {
		cfg.Replicates = 40
	}
	if cfg.OptimizerSteps <= 0 {
		cfg.OptimizerSteps = 250
	}
	if cfg.Restarts <= 0 {
		cfg.Restarts = 2
	}
	if cfg.Seed == 0 {
		cfg.Seed = 20260829
	}
	vm, err := LoadVoynich("ZL3b", cfg.CorpusPath, cfg.IVTFFPath)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return err
	}
	baseMap := BaselineMapping(len(vm.Alphabet))
	base := Compute(vm, baseMap)
	bestMap, best := Optimize(vm, cfg.OptimizerSteps, cfg.Restarts, cfg.Seed)
	results := []MappingResult{{Corpus: "ZL3b", Control: "OBSERVED", Replicate: 0, Seed: cfg.Seed, Baseline: base, Best: best, Mapping: bestMap}}
	for ki, kind := range controlKinds {
		for rep := 0; rep < cfg.Replicates; rep++ {
			seed := cfg.Seed + int64((ki+1)*100000+rep)
			cc := Control(vm, kind, seed)
			bm, bt := Optimize(cc, cfg.OptimizerSteps, cfg.Restarts, seed)
			results = append(results, MappingResult{Corpus: "ZL3b", Control: kind, Replicate: rep + 1, Seed: seed, Baseline: Compute(cc, baseMap), Best: bt, Mapping: bm})
		}
	}
	replication := "NOT_COMPARABLE"
	var it *Corpus
	var itBase, itBest Metrics
	var itMap []int
	if cfg.IT2aPath != "" && cfg.IT2aIVTFFPath != "" {
		x, e := LoadVoynich("IT2a", cfg.IT2aPath, cfg.IT2aIVTFFPath)
		if e != nil {
			return e
		}
		it = &x
		itBase = Compute(x, BaselineMapping(len(x.Alphabet)))
		itMap, itBest = Optimize(x, cfg.OptimizerSteps, cfg.Restarts, cfg.Seed)
		if len(x.Alphabet) == len(vm.Alphabet) {
			if (best.Score-base.Score)*(itBest.Score-itBase.Score) > 0 {
				replication = "DIRECTION_REPLICATED"
			} else {
				replication = "NOT_REPLICATED"
			}
		}
		results = append(results, MappingResult{Corpus: "IT2a", Control: "OBSERVED", Seed: cfg.Seed, Baseline: itBase, Best: itBest, Mapping: itMap})
	}
	var natural *MappingResult
	if cfg.NaturalPath != "" {
		nc, e := LoadNatural(cfg.NaturalPath, vm.RawTokenCount)
		if e != nil {
			return e
		}
		nm, nb := Optimize(nc, cfg.OptimizerSteps, cfg.Restarts, cfg.Seed)
		nr := MappingResult{Corpus: nc.Name, Control: "NATURAL_TEXT", Seed: cfg.Seed, Baseline: Compute(nc, BaselineMapping(len(nc.Alphabet))), Best: nb, Mapping: nm}
		natural = &nr
		results = append(results, nr)
	}
	if err = writeInventory(cfg.OutputDir, cfg.CorpusPath, vm); err != nil {
		return err
	}
	if err = writeBaselineMapping(cfg.OutputDir, vm); err != nil {
		return err
	}
	if err = writeMappingResults(cfg.OutputDir, vm, results); err != nil {
		return err
	}
	if err = writeControlResults(cfg.OutputDir, results); err != nil {
		return err
	}
	stats := comparisons(best, results)
	if err = writePrimary(cfg.OutputDir, base, best, itBase, itBest, stats); err != nil {
		return err
	}
	if err = writeDocument(cfg.OutputDir, vm, baseMap); err != nil {
		return err
	}
	decision := decide(stats)
	return writeMarkdown(cfg, vm, base, best, bestMap, it, itBase, itBest, natural, replication, decision, stats)
}

type comparison struct {
	Metric, Control                            string
	Observed, Mean, Median, Low, High, Z, P, Q float64
}

func metricValue(m Metrics, n string) float64 {
	switch n {
	case "SEQUENTIAL":
		return m.SequentialComponent
	case "DIFFERENCE":
		return m.DifferenceComponent
	case "DOCUMENT":
		return m.DocumentComponent
	case "EDIT":
		return m.EditSubstitutionConsistency
	}
	return m.Score
}
func comparisons(observed Metrics, all []MappingResult) []comparison {
	names := []string{"SEQUENTIAL", "DIFFERENCE", "DOCUMENT", "EDIT", "AGGREGATE"}
	var out []comparison
	for _, kind := range controlKinds {
		for _, name := range names {
			var x []float64
			for _, r := range all {
				if r.Corpus == "ZL3b" && r.Control == kind {
					x = append(x, metricValue(r.Best, name))
				}
			}
			sort.Float64s(x)
			o := metricValue(observed, name)
			mu := mean(x)
			sd := 0.0
			ge := 0
			for _, v := range x {
				sd += (v - mu) * (v - mu)
				if v >= o {
					ge++
				}
			}
			if len(x) > 1 {
				sd = math.Sqrt(sd / float64(len(x)-1))
			}
			z := 0.0
			if sd > 0 {
				z = (o - mu) / sd
			}
			out = append(out, comparison{name, kind, o, mu, quant(x, .5), quant(x, .025), quant(x, .975), z, float64(ge+1) / float64(len(x)+1), 0})
		}
	}
	bh(out)
	return out
}
func quant(x []float64, p float64) float64 {
	if len(x) == 0 {
		return 0
	}
	return x[int(math.Round(p*float64(len(x)-1)))]
}
func bh(x []comparison) {
	idx := make([]int, len(x))
	for i := range idx {
		idx[i] = i
	}
	sort.Slice(idx, func(i, j int) bool { return x[idx[i]].P < x[idx[j]].P })
	q := 1.0
	for rank := len(idx); rank >= 1; rank-- {
		i := idx[rank-1]
		v := x[i].P * float64(len(idx)) / float64(rank)
		if v < q {
			q = v
		}
		x[i].Q = math.Min(1, q)
	}
}
func decide(x []comparison) string {
	pass := map[string]bool{}
	for _, c := range x {
		if c.Q <= .05 && c.Z > 0 {
			pass[c.Metric] = true
		}
	}
	if pass["SEQUENTIAL"] && pass["EDIT"] && pass["DOCUMENT"] {
		return "GLOBAL_NUMERIC_SIGNAL"
	}
	if pass["SEQUENTIAL"] || pass["DIFFERENCE"] || pass["EDIT"] {
		return "LOCAL_NUMERIC_SIGNAL"
	}
	return "NO_NUMERIC_SIGNAL"
}

func create(path string) (*bufio.Writer, *os.File, error) {
	f, e := os.Create(path)
	if e != nil {
		return nil, nil, e
	}
	return bufio.NewWriter(f), f, nil
}
func closeWriter(w *bufio.Writer, f *os.File) error {
	if e := w.Flush(); e != nil {
		f.Close()
		return e
	}
	return f.Close()
}
func writeInventory(dir, path string, c Corpus) error {
	b, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	type row struct {
		occ, tok, ini, med, fin, single int
		examples                        map[string]bool
	}
	rows := map[byte]*row{}
	for _, s := range strings.Fields(string(b)) {
		seen := map[byte]bool{}
		for i, g := range []byte(s) {
			r := rows[g]
			if r == nil {
				r = &row{examples: map[string]bool{}}
				rows[g] = r
			}
			r.occ++
			seen[g] = true
			if len(s) == 1 {
				r.single++
			} else if i == 0 {
				r.ini++
			} else if i == len(s)-1 {
				r.fin++
			} else {
				r.med++
			}
			if len(r.examples) < 5 {
				r.examples[s] = true
			}
		}
		for g := range seen {
			rows[g].tok++
		}
	}
	keys := make([]int, 0, len(rows))
	for g := range rows {
		keys = append(keys, int(g))
	}
	sort.Ints(keys)
	w, f, e := create(filepath.Join(dir, "GLYPH_INVENTORY.tsv"))
	if e != nil {
		return e
	}
	fmt.Fprintln(w, "representation\tcodepoint\tadmitted\tfrequency\ttoken_frequency\tinitial\tmedial\tfinal\tsingleton\texample_tokens")
	for _, k := range keys {
		g := byte(k)
		r := rows[g]
		ex := make([]string, 0, len(r.examples))
		for s := range r.examples {
			ex = append(ex, s)
		}
		sort.Strings(ex)
		fmt.Fprintf(w, "%s\tU+%04X\t%t\t%d\t%d\t%d\t%d\t%d\t%d\t%s\n", string(g), g, g >= 'a' && g <= 'z', r.occ, r.tok, r.ini, r.med, r.fin, r.single, strings.Join(ex, ";"))
	}
	return closeWriter(w, f)
}
func writeBaselineMapping(dir string, c Corpus) error {
	w, f, e := create(filepath.Join(dir, "BASELINE_DIGIT_MAPPING.tsv"))
	if e != nil {
		return e
	}
	fmt.Fprintln(w, "glyph\tdigit\tbase\tpolicy")
	for i, g := range c.Alphabet {
		fmt.Fprintf(w, "%s\t%d\t%d\tASCII byte ascending over admitted lowercase a-z symbols actually observed\n", string(g), i, len(c.Alphabet))
	}
	return closeWriter(w, f)
}

func mappingString(a []byte, m []int) string {
	p := make([]string, len(a))
	for i, g := range a {
		p[i] = fmt.Sprintf("%c=%d", g, m[i])
	}
	return strings.Join(p, ";")
}
func writeMappingResults(dir string, c Corpus, rs []MappingResult) error {
	w, f, e := create(filepath.Join(dir, "NUMERIC_MAPPING_RESULTS.tsv"))
	if e != nil {
		return e
	}
	fmt.Fprintln(w, "corpus\tcontrol\treplicate\tseed\tmapping_kind\tscore\tzero_glyph\tleading_zero_tokens\tleading_zero_fraction\tcolliding_token_types\tcollision_classes\tcollision_fraction\tmapping")
	for _, r := range rs {
		z := ""
		for i, d := range r.Mapping {
			if d == 0 && i < len(c.Alphabet) {
				z = string(c.Alphabet[i])
			}
		}
		alphabet := c.Alphabet
		if len(r.Mapping) != len(alphabet) {
			alphabet = make([]byte, len(r.Mapping))
			for i := range alphabet {
				alphabet[i] = byte('a' + i)
			}
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\tOPTIMIZED\t%.9g\t%s\t%.0f\t%.9g\t%.0f\t%.0f\t%.9g\t%s\n", r.Corpus, r.Control, r.Replicate, r.Seed, r.Best.Score, z, r.Best.LeadingZeroTokenCount, r.Best.LeadingZeroFraction, r.Best.CollidingTokenTypeCount, r.Best.CollisionClassCount, r.Best.CollisionFraction, mappingString(alphabet, r.Mapping))
	}
	for zi := range c.Alphabet {
		m := make([]int, len(c.Alphabet))
		next := 1
		for i := range m {
			if i == zi {
				m[i] = 0
			} else {
				m[i] = next
				next++
			}
		}
		x := Compute(c, m)
		fmt.Fprintf(w, "ZL3b\tZERO_SENSITIVITY\t%d\t0\tFIXED_ZERO_ROLE\t%.9g\t%c\t%.0f\t%.9g\t%.0f\t%.0f\t%.9g\t%s\n", zi+1, x.Score, c.Alphabet[zi], x.LeadingZeroTokenCount, x.LeadingZeroFraction, x.CollidingTokenTypeCount, x.CollisionClassCount, x.CollisionFraction, mappingString(c.Alphabet, m))
	}
	return closeWriter(w, f)
}
func writeControlResults(dir string, rs []MappingResult) error {
	w, f, e := create(filepath.Join(dir, "NUMERIC_CONTROL_RESULTS.tsv"))
	if e != nil {
		return e
	}
	fmt.Fprintln(w, "corpus\tcontrol\treplicate\tseed\tbaseline_score\toptimized_score\tsequential\tdifference\tdocument\tedit_consistency")
	for _, r := range rs {
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%.9g\t%.9g\t%.9g\t%.9g\t%.9g\t%.9g\n", r.Corpus, r.Control, r.Replicate, r.Seed, r.Baseline.Score, r.Best.Score, r.Best.SequentialComponent, r.Best.DifferenceComponent, r.Best.DocumentComponent, r.Best.EditSubstitutionConsistency)
	}
	return closeWriter(w, f)
}

func metricRows(m Metrics) map[string]float64 {
	return map[string]float64{"N1_LENGTH_MEAN": m.LengthMean, "N1_ADJACENT_LENGTH_DIFF_MEAN": m.AdjLengthDiffMean, "N1_LOG_B_N_PLUS_1_MEAN": m.LogMean, "N1_POSITION_LENGTH_RHO": m.PositionLengthRho, "N2_SIGNED_DELTA_MEAN": m.SignedDeltaMean, "N2_ABS_DELTA_MEAN": m.AbsDeltaMean, "N2_NORMALIZED_DELTA_MEAN": m.NormalizedDeltaMean, "N2_DELTA_ENTROPY": m.DeltaEntropy, "N2_REPEATED_DELTA_FRACTION": m.RepeatedDeltaFraction, "N3_INCREASING_FRACTION": m.IncreasingFraction, "N3_DECREASING_FRACTION": m.DecreasingFraction, "N3_LONGEST_MONOTONIC_RUN": m.LongestMonotonicRun, "N3_POSITION_VALUE_RHO": m.PositionValueRho, "N4_LAG1_RHO": m.LagRho[0], "N4_LAG2_RHO": m.LagRho[1], "N4_LAG3_RHO": m.LagRho[2], "N4_LAG4_RHO": m.LagRho[3], "N4_LAG5_RHO": m.LagRho[4], "N5_AP_CLOSENESS": m.APCloseness, "N6_RATIO_REPEAT": m.RatioRepeat, "LEADING_ZERO_FRACTION": m.LeadingZeroFraction, "COLLISION_FRACTION": m.CollisionFraction, "EDIT_SUBSTITUTION_CONSISTENCY": m.EditSubstitutionConsistency, "NUMERIC_REGULARITY_SCORE": m.Score}
}
func writePrimary(dir string, base, best, itBase, itBest Metrics, cs []comparison) error {
	w, f, e := create(filepath.Join(dir, "NUMERIC_PRIMARY_RESULTS.tsv"))
	if e != nil {
		return e
	}
	fmt.Fprintln(w, "corpus\tmapping\tmetric\tobserved\tnull_mean\tnull_median\tnull_ci_2_5\tnull_ci_97_5\tstandardized_effect\tempirical_p\tbh_fdr_q\tstatus")
	sets := []struct {
		c, k string
		m    Metrics
	}{{"ZL3b", "BASELINE", base}, {"ZL3b", "OPTIMIZED", best}, {"IT2a", "BASELINE", itBase}, {"IT2a", "OPTIMIZED", itBest}}
	for _, s := range sets {
		rows := metricRows(s.m)
		names := make([]string, 0, len(rows))
		for n := range rows {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			v := rows[n]
			fmt.Fprintf(w, "%s\t%s\t%s\t%.9g\tNA\tNA\tNA\tNA\tNA\tNA\tNA\tDESCRIPTIVE\n", s.c, s.k, n, v)
		}
	}
	for _, x := range cs {
		fmt.Fprintf(w, "ZL3b\tOPTIMIZED_VS_%s\t%s\t%.9g\t%.9g\t%.9g\t%.9g\t%.9g\t%.9g\t%.9g\t%.9g\tPRIMARY_BH_FDR\n", x.Control, x.Metric, x.Observed, x.Mean, x.Median, x.Low, x.High, x.Z, x.P, x.Q)
	}
	return closeWriter(w, f)
}
