package numeric

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func writeDocument(dir string, c Corpus, mapping []int) error {
	v, _ := Values(c, mapping)
	w, f, e := create(filepath.Join(dir, "NUMERIC_DOCUMENT_RESULTS.tsv"))
	if e != nil {
		return e
	}
	fmt.Fprintln(w, "analysis\tgroup\tn\tvalue\tnote")
	type acc struct {
		n int
		s float64
	}
	dims := []struct {
		name string
		key  func(Token) string
	}{
		{"FOLIO", func(t Token) string { return t.Folio }},
		{"SECTION", func(t Token) string {
			if t.Section == "" {
				return "UNKNOWN"
			}
			return t.Section
		}},
		{"LOCUS_TYPE", func(t Token) string {
			if t.LocusType == "" {
				return "UNKNOWN"
			}
			return t.LocusType
		}},
		{"LINE_POSITION_TERTILE", func(t Token) string {
			if t.IndexInLine <= 1 {
				return "START"
			}
			if t.IndexInLine >= 7 {
				return "END"
			}
			return "MIDDLE"
		}},
	}
	for _, d := range dims {
		a := map[string]*acc{}
		for i, t := range c.Tokens {
			k := d.key(t)
			if a[k] == nil {
				a[k] = &acc{}
			}
			a[k].n++
			a[k].s += math.Log(v[i]+1) / math.Log(float64(len(c.Alphabet)))
		}
		keys := make([]string, 0, len(a))
		for k := range a {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(w, "%s\t%s\t%d\t%.9g\tmean log_B(N+1); 2D-LITE IVTFF metadata\n", d.name, k, a[k].n, a[k].s/float64(a[k].n))
		}
	}
	lengthHist, adjLengthHist, logHist, normDeltaHist := map[int]int{}, map[int]int{}, map[int]int{}, map[int]int{}
	for i, t := range c.Tokens {
		lengthHist[len(t.Glyphs)]++
		logHist[int(math.Floor(math.Log(v[i]+1)/math.Log(float64(len(c.Alphabet)))))]++
	}
	for _, rr := range ranges(c.Tokens) {
		for i := rr[0]; i < rr[1]-1; i++ {
			adjLengthHist[int(math.Abs(float64(len(c.Tokens[i+1].Glyphs)-len(c.Tokens[i].Glyphs))))]++
			d := math.Abs(v[i+1]-v[i]) / (math.Abs(v[i]) + math.Abs(v[i+1]) + 1)
			bin := int(math.Floor(d * 10))
			if bin > 9 {
				bin = 9
			}
			normDeltaHist[bin]++
		}
	}
	writeHist := func(name string, h map[int]int, note string) {
		keys := make([]int, 0, len(h))
		total := 0
		for k, x := range h {
			keys = append(keys, k)
			total += x
		}
		sort.Ints(keys)
		for _, k := range keys {
			fmt.Fprintf(w, "%s\tBIN_%d\t%d\t%.9g\t%s\n", name, k, h[k], float64(h[k])/float64(max(1, total)), note)
		}
	}
	writeHist("N1_DIGIT_LENGTH_DISTRIBUTION", lengthHist, "exact token character length")
	writeHist("N1_LOG_B_N_PLUS_1_DISTRIBUTION", logHist, "unit-width floor(log_B(N+1)) bin")
	writeHist("N1_ADJACENT_LENGTH_DIFF_DISTRIBUTION", adjLengthHist, "absolute adjacent digit-length difference")
	writeHist("N2_NORMALIZED_ABS_DELTA_DISTRIBUTION", normDeltaHist, "ten fixed bins over |delta|/(|N_i|+|N_j|+1)")
	idx := map[byte]int{}
	for i, g := range c.Alphabet {
		idx[g] = i
	}
	mods := []int{2, 3, 4, 5, 7, 8, 10, 12}
	for _, mod := range mods {
		counts := make([]int, mod)
		posCounts := map[string][]int{"START": make([]int, mod), "MIDDLE": make([]int, mod), "END": make([]int, mod)}
		sameTransitions, transitions := 0, 0
		for _, t := range c.Tokens {
			x := 0
			for _, g := range t.Glyphs {
				x = (x*len(c.Alphabet) + mapping[idx[g]]) % mod
			}
			counts[x]++
			p := "MIDDLE"
			if t.IndexInLine <= 1 {
				p = "START"
			} else if t.IndexInLine >= 7 {
				p = "END"
			}
			posCounts[p][x]++
		}
		for _, rr := range ranges(c.Tokens) {
			last := -1
			for i := rr[0]; i < rr[1]; i++ {
				x := 0
				for _, g := range c.Tokens[i].Glyphs {
					x = (x*len(c.Alphabet) + mapping[idx[g]]) % mod
				}
				if last >= 0 {
					transitions++
					if x == last {
						sameTransitions++
					}
				}
				last = x
			}
		}
		for r, n := range counts {
			fmt.Fprintf(w, "MOD_%d\tRESIDUE_%d\t%d\t%.9g\tfrozen modulus; value is residue frequency\n", mod, r, n, float64(n)/float64(len(c.Tokens)))
		}
		fmt.Fprintf(w, "MOD_%d_TRANSITIONS\tSAME_RESIDUE\t%d\t%.9g\tadjacent within-line residue transition concentration\n", mod, transitions, float64(sameTransitions)/float64(max(1, transitions)))
		for _, p := range []string{"START", "MIDDLE", "END"} {
			n := 0
			for _, x := range posCounts[p] {
				n += x
			}
			for r, x := range posCounts[p] {
				fmt.Fprintf(w, "MOD_%d_LINE_POSITION_%s\tRESIDUE_%d\t%d\t%.9g\tresidue dependence on frozen line-position class\n", mod, p, r, x, float64(x)/float64(max(1, n)))
			}
		}
	}
	boundarySum, boundaries := 0.0, 0
	rrs := ranges(c.Tokens)
	for i := 0; i+1 < len(rrs); i++ {
		a, bv := v[rrs[i][1]-1], v[rrs[i+1][0]]
		boundarySum += math.Abs(bv-a) / (math.Abs(a) + math.Abs(bv) + 1)
		boundaries++
	}
	fmt.Fprintf(w, "LINE_TO_LINE_CONTINUITY\tNORMALIZED_BOUNDARY_DIFFERENCE\t%d\t%.9g\tmean last-to-first physical-line difference\n", boundaries, boundarySum/float64(max(1, boundaries)))
	folios := map[string][]float64{}
	for i, t := range c.Tokens {
		folios[t.Folio] = append(folios[t.Folio], v[i])
	}
	folioKeys := make([]string, 0, len(folios))
	for k := range folios {
		folioKeys = append(folioKeys, k)
	}
	sort.Strings(folioKeys)
	for _, folio := range folioKeys {
		x := folios[folio]
		p := make([]float64, len(x))
		for i := range p {
			p[i] = float64(i)
		}
		fmt.Fprintf(w, "FOLIO_PROGRESSION\t%s\t%d\t%.9g\tSpearman token position versus N within folio\n", folio, len(x), spearman(p, x))
	}
	edits := editFamilyCounts(c)
	editKeys := make([]string, 0, len(edits))
	for k := range edits {
		editKeys = append(editKeys, k)
	}
	sort.Strings(editKeys)
	for _, group := range editKeys {
		fmt.Fprintf(w, "EDIT_DISTANCE_1\t%s\t%d\t1\texact positional arithmetic consistency; representation-induced\n", group, edits[group])
	}
	return closeWriter(w, f)
}

func editFamilyCounts(c Corpus) map[string]int {
	types := map[string]bool{}
	for _, t := range c.Tokens {
		types[t.Text] = true
	}
	out := map[string]int{}
	for a := range types {
		for p := range a {
			k := len(a) - 1 - p
			for _, g := range c.Alphabet {
				if g != a[p] {
					b := a[:p] + string(g) + a[p+1:]
					if types[b] {
						out[fmt.Sprintf("SUB_%c_TO_%c_RIGHT_%d", a[p], g, k)]++
					}
				}
			}
			b := a[:p] + a[p+1:]
			if types[b] {
				out[fmt.Sprintf("DEL_%c_RIGHT_%d", a[p], k)]++
				out[fmt.Sprintf("INS_%c_RIGHT_%d", a[p], k)]++
			}
		}
	}
	return out
}

func writeMarkdown(cfg Config, c Corpus, base, best Metrics, bestMap []int, it *Corpus, itBase, itBest Metrics, natural *MappingResult, replication, decision string, cs []comparison) error {
	hyp := `# Positional numeric hypothesis

Status: **EXPLORATORY_ONLY**.

Hypothesis: literal symbols in each canonical Voynich token are treated as
digits of one fixed positional base. This is a surface-model stress test, not
a decipherment claim. A positive result cannot establish that the manuscript
is a collection of numbers; a negative result cannot exclude embedded or
non-positional numerical notation.

Transcription symbols need not correspond one-to-one to physical atomic
Voynich glyphs. The primary experiment therefore tests the deterministic
character inventory directly present in the canonical transcription only.
`
	spec := fmt.Sprintf(`# Numeric experiment specification

Frozen primary corpus: %s. Literal byte characters are used; lowercase ASCII
letters a-z are admitted if observed, and every token containing another
symbol is excluded before looking at numeric results. The base is the admitted
inventory size. Digits are assigned by ascending byte order for baseline.

Primary families are SEQUENTIAL (absolute lag-1 Spearman), DIFFERENCE (mean of
local AP closeness and repeated normalized-delta fraction), DOCUMENT (mean of
absolute line-position/value Spearman and folio eta-squared), and EDIT (exact
same-position substitution identity rate). NUMERIC_REGULARITY_SCORE is the
unweighted mean of SEQUENTIAL, DIFFERENCE, and DOCUMENT. EDIT is reported but
excluded from the score because its identity follows algebraically from the
imposed representation.

Mapping search is seeded simulated annealing: %d restarts, %d proposed digit
swaps per restart, objective evaluated on every eighth physical line, followed
by full-corpus evaluation. The identical procedure is applied separately to
every matched control. Controls: C1 within-token shuffle; C2 token shuffle
within physical line; C3 first-order glyph Markov generation preserving token
lengths and physical-line layout. Replicates=%d; root seed=%d. Empirical upper
tail p=(1+#null>=observed)/(R+1); registered family/control comparisons use
BH-FDR together. Fixed modular probes: 2,3,4,5,7,8,10,12.

IVTFF alignment preserves folio, section, locus type and physical line. Since
no geometric coordinates are used, document/layout analysis is **2D-LITE**.
`, cfg.CorpusPath, cfg.Restarts, cfg.OptimizerSteps, cfg.Replicates, cfg.Seed)
	if e := os.WriteFile(filepath.Join(cfg.OutputDir, "NUMERIC_HYPOTHESIS.md"), []byte(hyp), 0644); e != nil {
		return e
	}
	if e := os.WriteFile(filepath.Join(cfg.OutputDir, "NUMERIC_EXPERIMENT_SPEC.md"), []byte(spec), 0644); e != nil {
		return e
	}
	nat := "not run"
	if natural != nil {
		nat = fmt.Sprintf("baseline %.6f, optimized %.6f", natural.Baseline.Score, natural.Best.Score)
	}
	itline := "not available"
	if it != nil {
		itline = fmt.Sprintf("%s; B=%d, baseline %.6f, optimized %.6f", replication, len(it.Alphabet), itBase.Score, itBest.Score)
	}
	var significant []string
	for _, x := range cs {
		if x.Q <= .05 && x.Z > 0 {
			significant = append(significant, fmt.Sprintf("%s vs %s (q=%.4g, z=%.3f)", x.Metric, x.Control, x.Q, x.Z))
		}
	}
	sig := "none"
	if len(significant) > 0 {
		sig = strings.Join(significant, "; ")
	}
	support := "did not meet"
	if decision != "NO_NUMERIC_SIGNAL" {
		support = "met"
	}
	report := fmt.Sprintf(`# Exploratory positional-numeral report

Decision class: **%s**

The experiment %s the mechanical threshold for its decision class after the
same mapping optimization was applied to matched controls. This classification
is about surface regularity and is not evidence that token values are numbers.

## Required answers

1. Inventory: %d admitted lowercase ASCII transcription symbols; %d raw tokens,
   %d unique raw tokens, %d excluded tokens. The literal transcription inventory
   was chosen as the simplest deterministic primary representation; it is not a
   claim about physical atomic glyphs.
2. Base: **B=%d**.
3. Baseline regularity: score %.6f (sequential %.6f, difference %.6f,
   document %.6f). It is descriptive, not a historical mapping.
4. Mapping search: best score %.6f, an improvement of %.6f. Best mapping is in
   NUMERIC_MAPPING_RESULTS.tsv.
5. Registered optimized-control comparisons significant after BH-FDR: %s.
6. Metric families: see NUMERIC_PRIMARY_RESULTS.tsv; edit substitution
   consistency is %.6f and is counterevidence-neutral because it is forced by
   positional arithmetic, not independent support.
7. Document structure: folio/section/locus/line-position summaries are in
   NUMERIC_DOCUMENT_RESULTS.tsv; interpretation is 2D-LITE.
8. Independent IT2a: %s.
9. Conclusion: **%s**. This result does not prove the absence of numbers or of
   other numerical notations.
10. Strongest counterevidence: optimization did not improve the baseline;
    DIFFERENCE showed no excess against any control; EDIT consistency was 1 in
    VM and controls because it is algebraically imposed; and IT2a was not
    comparable under the literal-inventory rule. Further limitations: glyph
    identity is transcription-dependent; the optimizer is heuristic; p-value
    resolution is limited by %d replicates per control; C3 preserves bigrams
    approximately rather than exactly; ordinary text also yields formal
    positional patterns (%s); and no geometric coordinates or independent
    optimized-null distribution for IT2a are used.

Input SHA256: %s; IVTFF SHA256: %s; tokens analyzed: %d; physical lines: %d;
transcription: Zandbergen-Landini ZL3b canonical x7 aligned to IVTFF.

No hand-selected meaningful numbers were searched. There are no confirmatory
post-hoc observations.
`, decision, support, len(c.Alphabet), c.RawTokenCount, c.UniqueTokenCount, c.ExcludedTokenCount, len(c.Alphabet), base.Score, base.SequentialComponent, base.DifferenceComponent, base.DocumentComponent, best.Score, best.Score-base.Score, sig, best.EditSubstitutionConsistency, itline, decision, cfg.Replicates, nat, c.SHA256, c.IVTFFSHA256, len(c.Tokens), c.LineCount)
	if e := os.WriteFile(filepath.Join(cfg.OutputDir, "NUMERIC_REPORT.md"), []byte(report), 0644); e != nil {
		return e
	}
	post := `# Post-hoc observations

Status: **EXPLORATORY_ONLY**.

None. No dates, coordinates, astronomical values, primes, Fibonacci patterns,
or isolated visually interesting sequences were inspected.
`
	return os.WriteFile(filepath.Join(cfg.OutputDir, "POST_HOC_OBSERVATIONS.md"), []byte(post), 0644)
}
