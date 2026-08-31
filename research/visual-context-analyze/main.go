// The visual-context research aggregator reuses Task79's frozen line profiles and occurrence
// metadata. It does not rerun any fingerprint analyzer.
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"zcore.dev/voinich/internal/metadatavalidation"
)

const (
	seed         = int64(20260830)
	repetitions  = 1000
	minTokens    = 20
	minLines     = 2
	generatedTag = "<!-- GENERATED_PAGE_REGISTRY -->"
)

var sectionNames = map[string]string{"A": "Astronomical", "B": "Biological", "C": "Cosmological", "H": "Herbal", "P": "Pharmaceutical", "S": "Stars", "T": "Text", "Z": "Zodiac"}
var panelSuffix = regexp.MustCompile(`^(.+[rv])\d+$`)

type occurrence struct {
	Token, Folio, Section, CurrierLanguage, Scribe, Quire string
	AbsoluteTokenPosition                                 int
}

type lineProfile struct {
	LineID                 string  `json:"line_id"`
	Folio                  string  `json:"folio"`
	Section                string  `json:"section"`
	Currier                string  `json:"currier"`
	Scribe                 string  `json:"scribe"`
	TokenCount             int     `json:"token_count"`
	CharacterCount         int     `json:"character_count"`
	VocabularySize         int     `json:"vocabulary_size"`
	ExactRepetitionRate    float64 `json:"exact_repetition_rate"`
	NearEditRepetitionRate float64 `json:"near_edit_repetition_rate"`
	TransitionEntropy      float64 `json:"transition_entropy"`
	TokenEntropy           float64 `json:"token_entropy"`
	FirstToken             string  `json:"first_token"`
	FinalToken             string  `json:"final_token"`
}

type page struct {
	ID, Leaf, Code, Section, Currier, Scribe, Quire string
	Order, Tokens, Lines, Characters, Unique        int
	TokenCounts                                     map[string]int
	LineLengths                                     []float64
	ExactNum, NearNum, Transitions                  float64
	TransitionEntropyNum, TokenEntropyLineNum       float64
	InitialLenNum, FinalLenNum                      float64
	Values                                          map[string]float64
	ScribeCounts                                    map[string]int
}

type metric struct{ Family, Name string }

var metrics = []metric{
	{"token", "mean_token_length"},
	{"token", "type_token_ratio"},
	{"token", "token_entropy"},
	{"sequence", "exact_adjacent_repetition"},
	{"sequence", "near_edit_adjacent_repetition"},
	{"sequence", "mean_line_transition_entropy"},
	{"line_block", "mean_line_tokens"},
	{"line_block", "line_length_cv"},
	{"line_block", "boundary_length_asymmetry"},
	{"sequence", "mean_line_token_entropy"},
}

type comparison struct {
	Family, Metric                      string
	Pages, WithinPairs, BetweenPairs    int
	Within, Between, Delta, Ratio, P, Q float64
	Status                              string
}

type confoundResult struct {
	Family, Metric, Status                 string
	N, ReducedRank, FullRank, AddedRank    int
	ReducedR2, FullR2, IncrementalR2, P, Q float64
}

func main() {
	linesPath := flag.String("line-profiles", "experiments/fingerprint-v2-task79-v1/canonical-out/line_profiles.json", "frozen Task79 line profiles")
	occPath := flag.String("occurrences", "experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl", "frozen Task79 occurrence metadata")
	ivtffPath := flag.String("ivtff", "data/ZL3b-n.txt", "IVTFF page-header metadata source")
	outDir := flag.String("out", ".", "output directory")
	flag.Parse()

	pages, err := loadPages(*occPath, *linesPath, *ivtffPath)
	must(err)
	must(writeTaxonomy(*outDir, pages))
	must(writeFingerprints(*outDir, pages))
	must(writeContingencies(*outDir, pages))
	comparisons := compareSections(inferential(pages))
	confounders := analyzeConfounders(inferential(pages))
	must(writeComparisons(*outDir, comparisons))
	must(writeSectionSummary(*outDir, inferential(pages)))
	must(writeConfounders(*outDir, confounders))
	must(writeClassification(*outDir, inferential(pages)))
	must(writeManifest(*outDir, []string{
		"VISUAL_CONTEXT_EXPERIMENT_REVIEW.md", "VISUAL_CONTEXT_EXISTING_RESEARCH_AUDIT.md",
		"VISUAL_CONTEXT_EXISTING_EVIDENCE.tsv", "VISUAL_CONTEXT_TAXONOMY.md",
		"VISUAL_CONTEXT_TAXONOMY.tsv", "VISUAL_CONTEXT_PAGE_FINGERPRINTS.tsv",
		"VISUAL_CONTEXT_CONTINGENCY.tsv", "VISUAL_CONTEXT_SECTION_COMPARISONS.tsv",
		"VISUAL_CONTEXT_SECTION_SUMMARY.tsv",
		"VISUAL_CONTEXT_CONFOUNDER_ANALYSIS.tsv", "VISUAL_CONTEXT_CLASSIFICATION.tsv",
		"VISUAL_CONTEXT_EXPERIMENT_REPORT.md",
	}))
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func loadPages(occPath, linePath, ivtffPath string) ([]*page, error) {
	doc, err := metadatavalidation.ParseIVTFF(ivtffPath)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(occPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	byID := map[string]*page{}
	var order []string
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for s.Scan() {
		var o occurrence
		if err := json.Unmarshal(s.Bytes(), &o); err != nil {
			return nil, err
		}
		p := byID[o.Folio]
		vars := doc.PageVariables[o.Folio]
		code := vars["I"]
		if p == nil {
			name, ok := sectionNames[code]
			if !ok {
				name = "Unknown"
			}
			p = &page{ID: o.Folio, Leaf: leafID(o.Folio), Code: code, Section: name, Currier: vars["L"], Quire: vars["Q"], Order: len(order), TokenCounts: map[string]int{}, Values: map[string]float64{}, ScribeCounts: map[string]int{}}
			byID[o.Folio] = p
			order = append(order, o.Folio)
		} else if p.Code != code || p.Currier != vars["L"] || p.Quire != vars["Q"] {
			return nil, fmt.Errorf("conflicting metadata within page %s", o.Folio)
		}
		p.Tokens++
		p.TokenCounts[o.Token]++
		p.ScribeCounts[blank(o.Scribe)]++
		p.Characters += glyphLen(o.Token)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	b, err := os.ReadFile(linePath)
	if err != nil {
		return nil, err
	}
	var lines []lineProfile
	if err := json.Unmarshal(b, &lines); err != nil {
		return nil, err
	}
	for _, l := range lines {
		p := byID[l.Folio]
		if p == nil {
			return nil, fmt.Errorf("line profile references unknown page %s", l.Folio)
		}
		p.Lines++
		p.LineLengths = append(p.LineLengths, float64(l.TokenCount))
		tr := float64(max(0, l.TokenCount-1))
		p.Transitions += tr
		p.ExactNum += l.ExactRepetitionRate * tr
		p.NearNum += l.NearEditRepetitionRate * tr
		p.TransitionEntropyNum += l.TransitionEntropy * tr
		p.TokenEntropyLineNum += l.TokenEntropy * float64(l.TokenCount)
		p.InitialLenNum += float64(glyphLen(l.FirstToken))
		p.FinalLenNum += float64(glyphLen(l.FinalToken))
	}
	pages := make([]*page, 0, len(order))
	for _, id := range order {
		p := byID[id]
		finalize(p)
		pages = append(pages, p)
	}
	return pages, nil
}

func leafID(s string) string {
	if m := panelSuffix.FindStringSubmatch(s); m != nil {
		return m[1]
	}
	return s
}
func glyphLen(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\x1f") + 1
}

func finalize(p *page) {
	p.Unique = len(p.TokenCounts)
	p.Scribe = dominant(p.ScribeCounts)
	p.Values["mean_token_length"] = div(float64(p.Characters), float64(p.Tokens))
	p.Values["type_token_ratio"] = div(float64(p.Unique), float64(p.Tokens))
	for _, n := range p.TokenCounts {
		q := float64(n) / float64(p.Tokens)
		p.Values["token_entropy"] -= q * math.Log2(q)
	}
	p.Values["exact_adjacent_repetition"] = div(p.ExactNum, p.Transitions)
	p.Values["near_edit_adjacent_repetition"] = div(p.NearNum, p.Transitions)
	p.Values["mean_line_transition_entropy"] = div(p.TransitionEntropyNum, p.Transitions)
	p.Values["mean_line_tokens"] = div(float64(p.Tokens), float64(p.Lines))
	p.Values["line_length_cv"] = cv(p.LineLengths)
	p.Values["boundary_length_asymmetry"] = math.Abs(div(p.InitialLenNum, float64(p.Lines)) - div(p.FinalLenNum, float64(p.Lines)))
	p.Values["mean_line_token_entropy"] = div(p.TokenEntropyLineNum, float64(p.Tokens))
}

func dominant(counts map[string]int) string {
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	best := ""
	for _, k := range keys {
		if counts[k] > counts[best] {
			best = k
		}
	}
	if len(keys) > 1 {
		return best + "_dominant_mixed_" + strings.Join(keys, "_")
	}
	return best
}
func div(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}
func cv(x []float64) float64 {
	if len(x) < 2 {
		return 0
	}
	m := mean(x)
	var s float64
	for _, v := range x {
		s += (v - m) * (v - m)
	}
	return math.Sqrt(s/float64(len(x)-1)) / m
}
func mean(x []float64) float64 {
	var s float64
	for _, v := range x {
		s += v
	}
	return div(s, float64(len(x)))
}
func inferential(ps []*page) []*page {
	var out []*page
	for _, p := range ps {
		if p.Tokens >= minTokens && p.Lines >= minLines && p.Code != "" {
			out = append(out, p)
		}
	}
	return out
}

func writeTaxonomy(out string, ps []*page) error {
	var t strings.Builder
	t.WriteString("page_id\tphysical_leaf\tvisual_code\tvisual_class\tsource\tambiguity\tinclusion_status\tcurrier\tscribe\tquire\ttoken_count\tline_count\n")
	for _, p := range ps {
		status := "INCLUDED"
		if p.Code == "" {
			status = "MISSING_VISUAL_CLASS"
		}
		fmt.Fprintf(&t, "%s\t%s\t%s\t%s\tTask79 strict IVTFF $I alignment\tnone_recorded\t%s\t%s\t%s\t%s\t%d\t%d\n", p.ID, p.Leaf, p.Code, p.Section, status, blank(p.Currier), blank(p.Scribe), blank(p.Quire), p.Tokens, p.Lines)
	}
	if err := os.WriteFile(filepath.Join(out, "VISUAL_CONTEXT_TAXONOMY.tsv"), []byte(t.String()), 0644); err != nil {
		return err
	}
	path := filepath.Join(out, "VISUAL_CONTEXT_TAXONOMY.md")
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	base := strings.Split(string(b), generatedTag)[0] + generatedTag + "\n\n"
	var md strings.Builder
	md.WriteString(base)
	md.WriteString("| Page ID | Physical leaf | Code | Class | Source | Ambiguity | Status | Currier | Scribe | Quire | Tokens | Lines |\n|---|---|---|---|---|---|---|---|---|---|---:|---:|\n")
	for _, p := range ps {
		status := "INCLUDED"
		if p.Code == "" {
			status = "MISSING_VISUAL_CLASS"
		}
		fmt.Fprintf(&md, "| %s | %s | %s | %s | Task79/IVTFF `$I` | none recorded | %s | %s | %s | %s | %d | %d |\n", p.ID, p.Leaf, p.Code, p.Section, status, blank(p.Currier), blank(p.Scribe), blank(p.Quire), p.Tokens, p.Lines)
	}
	return os.WriteFile(path, []byte(md.String()), 0644)
}

func blank(s string) string {
	if s == "" {
		return "NA"
	}
	return s
}

func writeFingerprints(out string, ps []*page) error {
	var b strings.Builder
	b.WriteString("page_id\tphysical_leaf\tpage_order\tvisual_code\tvisual_class\tcurrier\tscribe\tquire\ttoken_count\tline_count\tunique_tokens")
	for _, m := range metrics {
		b.WriteByte('\t')
		b.WriteString(m.Name)
	}
	b.WriteByte('\n')
	for _, p := range ps {
		fmt.Fprintf(&b, "%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\t%d\t%d\t%d", p.ID, p.Leaf, p.Order, p.Code, p.Section, blank(p.Currier), blank(p.Scribe), blank(p.Quire), p.Tokens, p.Lines, p.Unique)
		for _, m := range metrics {
			fmt.Fprintf(&b, "\t%.9g", p.Values[m.Name])
		}
		b.WriteByte('\n')
	}
	return os.WriteFile(filepath.Join(out, "VISUAL_CONTEXT_PAGE_FINGERPRINTS.tsv"), []byte(b.String()), 0644)
}

func pairStats(ps []*page, m string, labels []string) (within, between float64, nw, nb int) {
	for i := 0; i < len(ps); i++ {
		for j := i + 1; j < len(ps); j++ {
			if ps[i].Leaf == ps[j].Leaf {
				continue
			}
			d := math.Abs(ps[i].Values[m] - ps[j].Values[m])
			if labels[i] == labels[j] {
				within += d
				nw++
			} else {
				between += d
				nb++
			}
		}
	}
	return div(within, float64(nw)), div(between, float64(nb)), nw, nb
}

func compareSections(ps []*page) []comparison {
	labels := make([]string, len(ps))
	for i, p := range ps {
		labels[i] = p.Code
	}
	rng := rand.New(rand.NewSource(seed))
	out := make([]comparison, 0, len(metrics))
	groups := leafGroups(ps)
	for _, m := range metrics {
		w, b, nw, nb := pairStats(ps, m.Name, labels)
		obs := b - w
		ex := 0
		for r := 0; r < repetitions; r++ {
			perm := permuteByLeafShape(labels, groups, rng)
			pw, pb, _, _ := pairStats(ps, m.Name, perm)
			if pb-pw >= obs-1e-15 {
				ex++
			}
		}
		status := "OK"
		if nw == 0 || nb == 0 {
			status = "INSUFFICIENT_PAIRS"
		}
		out = append(out, comparison{m.Family, m.Name, len(ps), nw, nb, w, b, obs, div(b, w), float64(ex+1) / float64(repetitions+1), 0, status})
	}
	qs := bh(extractPComparisons(out))
	for i := range out {
		out[i].Q = qs[i]
	}
	return out
}

func leafGroups(ps []*page) [][]int {
	by := map[string][]int{}
	var keys []string
	for i, p := range ps {
		if _, ok := by[p.Leaf]; !ok {
			keys = append(keys, p.Leaf)
		}
		by[p.Leaf] = append(by[p.Leaf], i)
	}
	var out [][]int
	for _, k := range keys {
		out = append(out, by[k])
	}
	return out
}
func permuteByLeafShape(labels []string, groups [][]int, rng *rand.Rand) []string {
	out := append([]string(nil), labels...)
	bySize := map[int][][]int{}
	for _, g := range groups {
		bySize[len(g)] = append(bySize[len(g)], g)
	}
	sizes := sortedIntKeys(bySize)
	for _, size := range sizes {
		gs := bySize[size]
		perm := rng.Perm(len(gs))
		for i, dst := range gs {
			src := gs[perm[i]]
			for k := range dst {
				out[dst[k]] = labels[src[k]]
			}
		}
	}
	return out
}

func sortedIntKeys[T any](m map[int]T) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}
func extractPComparisons(x []comparison) []float64 {
	r := make([]float64, len(x))
	for i := range x {
		r[i] = x[i].P
	}
	return r
}

func writeComparisons(out string, x []comparison) error {
	var b strings.Builder
	b.WriteString("family\tmetric\tpages\twithin_pairs\tbetween_pairs\twithin_mean_abs_diff\tbetween_mean_abs_diff\tbetween_minus_within\tbetween_within_ratio\tpermutation_p\tbh_q\trepetitions\tpermutation_unit\tstatus\n")
	for _, r := range x {
		fmt.Fprintf(&b, "%s\t%s\t%d\t%d\t%d\t%.9g\t%.9g\t%.9g\t%.9g\t%.9g\t%.9g\t%d\tphysical_leaf_shape_block\t%s\n", r.Family, r.Metric, r.Pages, r.WithinPairs, r.BetweenPairs, r.Within, r.Between, r.Delta, r.Ratio, r.P, r.Q, repetitions, r.Status)
	}
	return os.WriteFile(filepath.Join(out, "VISUAL_CONTEXT_SECTION_COMPARISONS.tsv"), []byte(b.String()), 0644)
}

func writeSectionSummary(out string, ps []*page) error {
	var b strings.Builder
	b.WriteString("family\tmetric\tvisual_code\tvisual_class\tpages\tmean\tsd\tmedian\tmin\tmax\n")
	for _, m := range metrics {
		by := map[string][]float64{}
		for _, p := range ps {
			by[p.Code] = append(by[p.Code], p.Values[m.Name])
		}
		codes := make([]string, 0, len(by))
		for code := range by {
			codes = append(codes, code)
		}
		sort.Strings(codes)
		for _, code := range codes {
			x := by[code]
			sort.Float64s(x)
			mu := mean(x)
			var ss float64
			for _, v := range x {
				ss += (v - mu) * (v - mu)
			}
			sd := math.Sqrt(div(ss, float64(max(1, len(x)-1))))
			med := x[len(x)/2]
			if len(x)%2 == 0 {
				med = (x[len(x)/2-1] + x[len(x)/2]) / 2
			}
			fmt.Fprintf(&b, "%s\t%s\t%s\t%s\t%d\t%.9g\t%.9g\t%.9g\t%.9g\t%.9g\n", m.Family, m.Name, code, sectionNames[code], len(x), mu, sd, med, x[0], x[len(x)-1])
		}
	}
	return os.WriteFile(filepath.Join(out, "VISUAL_CONTEXT_SECTION_SUMMARY.tsv"), []byte(b.String()), 0644)
}

func writeContingencies(out string, ps []*page) error {
	var b strings.Builder
	b.WriteString("factor\tvisual_class\tfactor_level\tpages\ttokens\n")
	for _, factor := range []string{"currier", "scribe", "quire"} {
		counts := map[string][2]int{}
		for _, p := range ps {
			level := p.Currier
			if factor == "scribe" {
				level = p.Scribe
			}
			if factor == "quire" {
				level = p.Quire
			}
			k := p.Section + "\t" + blank(level)
			v := counts[k]
			v[0]++
			v[1] += p.Tokens
			counts[k] = v
		}
		keys := sortedKeys2(counts)
		for _, k := range keys {
			v := counts[k]
			fmt.Fprintf(&b, "%s\t%s\t%d\t%d\n", factor, k, v[0], v[1])
		}
	}
	return os.WriteFile(filepath.Join(out, "VISUAL_CONTEXT_CONTINGENCY.tsv"), []byte(b.String()), 0644)
}
func sortedKeys2(m map[string][2]int) []string {
	r := make([]string, 0, len(m))
	for k := range m {
		r = append(r, k)
	}
	sort.Strings(r)
	return r
}

func design(ps []*page, full bool) [][]float64 {
	cur, hand, quire, sec := levels(ps, func(p *page) string { return blank(p.Currier) }), levels(ps, func(p *page) string { return blank(p.Scribe) }), levels(ps, func(p *page) string { return blank(p.Quire) }), levels(ps, func(p *page) string { return p.Code })
	var x [][]float64
	for _, p := range ps {
		row := []float64{1}
		row = appendDummies(row, blank(p.Currier), cur)
		row = appendDummies(row, blank(p.Scribe), hand)
		row = appendDummies(row, blank(p.Quire), quire)
		row = append(row, math.Log1p(float64(p.Tokens)), float64(p.Lines), float64(p.Order)/float64(max(1, len(ps)-1)))
		if full {
			row = appendDummies(row, p.Code, sec)
		}
		x = append(x, row)
	}
	return x
}
func levels(ps []*page, f func(*page) string) []string {
	set := map[string]bool{}
	for _, p := range ps {
		set[f(p)] = true
	}
	var r []string
	for s := range set {
		r = append(r, s)
	}
	sort.Strings(r)
	return r
}
func appendDummies(row []float64, v string, levels []string) []float64 {
	for _, l := range levels[1:] {
		if v == l {
			row = append(row, 1)
		} else {
			row = append(row, 0)
		}
	}
	return row
}

// project returns fitted values and numerical rank using modified Gram-Schmidt.
func project(x [][]float64, y []float64) ([]float64, int) {
	n := len(x)
	if n == 0 {
		return nil, 0
	}
	p := len(x[0])
	q := make([][]float64, 0, p)
	for j := 0; j < p; j++ {
		v := make([]float64, n)
		for i := range x {
			v[i] = x[i][j]
		}
		for _, u := range q {
			dot := dot(v, u)
			for i := range v {
				v[i] -= dot * u[i]
			}
		}
		norm := math.Sqrt(dot(v, v))
		if norm > 1e-9 {
			for i := range v {
				v[i] /= norm
			}
			q = append(q, v)
		}
	}
	fit := make([]float64, n)
	for _, u := range q {
		c := dot(y, u)
		for i := range fit {
			fit[i] += c * u[i]
		}
	}
	return fit, len(q)
}
func dot(a, b []float64) float64 {
	var s float64
	for i := range a {
		s += a[i] * b[i]
	}
	return s
}
func sse(y, fit []float64) float64 {
	var s float64
	for i := range y {
		d := y[i] - fit[i]
		s += d * d
	}
	return s
}
func r2(y, fit []float64) float64 {
	m := mean(y)
	var tot float64
	for _, v := range y {
		d := v - m
		tot += d * d
	}
	return 1 - div(sse(y, fit), tot)
}

func analyzeConfounders(ps []*page) []confoundResult {
	xr, xf := design(ps, false), design(ps, true)
	_, rr := project(xr, make([]float64, len(ps)))
	_, fr := project(xf, make([]float64, len(ps)))
	groups := leafGroups(ps)
	rng := rand.New(rand.NewSource(seed + 1))
	var out []confoundResult
	for _, m := range metrics {
		y := make([]float64, len(ps))
		for i, p := range ps {
			y[i] = p.Values[m.Name]
		}
		fitR, _ := project(xr, y)
		fitF, _ := project(xf, y)
		sr, sf := sse(y, fitR), sse(y, fitF)
		inc := div(sr-sf, sr)
		status := "ESTIMABLE"
		pv := 1.0
		if fr == rr {
			status = "NOT_IDENTIFIABLE_RANK_ALIASED"
			inc = 0
		} else {
			res := make([]float64, len(y))
			for i := range y {
				res[i] = y[i] - fitR[i]
			}
			obs := sr - sf
			ex := 0
			for r := 0; r < repetitions; r++ {
				idx := permuteIndicesWithinQuireByLeafShape(ps, groups, rng)
				yp := make([]float64, len(y))
				for i := range y {
					yp[i] = fitR[i] + res[idx[i]]
				}
				pr, _ := project(xr, yp)
				pf, _ := project(xf, yp)
				if sse(yp, pr)-sse(yp, pf) >= obs-1e-15 {
					ex++
				}
			}
			pv = float64(ex+1) / float64(repetitions+1)
		}
		out = append(out, confoundResult{m.Family, m.Name, status, len(ps), rr, fr, fr - rr, r2(y, fitR), r2(y, fitF), inc, pv, 0})
	}
	qs := make([]float64, len(out))
	var ix []int
	for i, r := range out {
		if r.Status == "ESTIMABLE" {
			ix = append(ix, i)
			qs[i] = r.P
		}
	}
	var pp []float64
	for _, i := range ix {
		pp = append(pp, qs[i])
	}
	adj := bh(pp)
	for j, i := range ix {
		out[i].Q = adj[j]
	}
	return out
}
func permuteIndicesByLeafShape(groups [][]int, rng *rand.Rand) []int {
	n := 0
	for _, g := range groups {
		for _, i := range g {
			if i+1 > n {
				n = i + 1
			}
		}
	}
	out := make([]int, n)
	by := map[int][][]int{}
	for _, g := range groups {
		by[len(g)] = append(by[len(g)], g)
	}
	for _, size := range sortedIntKeys(by) {
		gs := by[size]
		perm := rng.Perm(len(gs))
		for i, d := range gs {
			s := gs[perm[i]]
			for k := range d {
				out[d[k]] = s[k]
			}
		}
	}
	return out
}

func permuteIndicesWithinQuireByLeafShape(ps []*page, groups [][]int, rng *rand.Rand) []int {
	out := make([]int, len(ps))
	buckets := map[string][][]int{}
	for _, g := range groups {
		key := ps[g[0]].Quire + "\x00" + strconv.Itoa(len(g))
		buckets[key] = append(buckets[key], g)
	}
	keys := make([]string, 0, len(buckets))
	for key := range buckets {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		gs := buckets[key]
		perm := rng.Perm(len(gs))
		for i, dst := range gs {
			src := gs[perm[i]]
			for k := range dst {
				out[dst[k]] = src[k]
			}
		}
	}
	return out
}

func permuteLabelsWithinQuireByLeafShape(labels []string, ps []*page, groups [][]int, rng *rand.Rand) []string {
	indices := permuteIndicesWithinQuireByLeafShape(ps, groups, rng)
	out := make([]string, len(labels))
	for i := range out {
		out[i] = labels[indices[i]]
	}
	return out
}

func writeConfounders(out string, x []confoundResult) error {
	var b strings.Builder
	b.WriteString("family\tmetric\tpages\treduced_rank\tfull_rank\tsection_added_rank\treduced_r2\tfull_r2\tincremental_section_r2\tpermutation_p\tbh_q\treduced_model\tfull_model\tpermutation_unit\tstatus\n")
	for _, r := range x {
		fmt.Fprintf(&b, "%s\t%s\t%d\t%d\t%d\t%d\t%.9g\t%.9g\t%.9g\t%.9g\t%.9g\tcurrier+scribe+quire+log_tokens+lines+position\treduced+visual_section\tphysical_leaf_shape_block_within_quire\t%s\n", r.Family, r.Metric, r.N, r.ReducedRank, r.FullRank, r.AddedRank, r.ReducedR2, r.FullR2, r.IncrementalR2, r.P, r.Q, r.Status)
	}
	return os.WriteFile(filepath.Join(out, "VISUAL_CONTEXT_CONFOUNDER_ANALYSIS.tsv"), []byte(b.String()), 0644)
}

func writeClassification(out string, ps []*page) error {
	labels := make([]string, len(ps))
	for i, p := range ps {
		labels[i] = p.Code
	}
	accuracy, baseline := classify(ps, labels)
	rng := rand.New(rand.NewSource(seed + 2))
	groups := leafGroups(ps)
	exceed := 0
	for r := 0; r < repetitions; r++ {
		permuted := permuteLabelsWithinQuireByLeafShape(labels, ps, groups, rng)
		a, _ := classify(ps, permuted)
		if a >= accuracy-1e-15 {
			exceed++
		}
	}
	p := float64(exceed+1) / float64(repetitions+1)
	var b strings.Builder
	b.WriteString("model\tsplit\tpages_scored\taccuracy\tmajority_baseline_accuracy\tlabel_permutation_p\trepetitions\tpermutation_unit\tstatus\tnotes\n")
	fmt.Fprintf(&b, "nearest_centroid\tleave_one_quire_out\t%d\t%.9g\t%.9g\t%.9g\t%d\tphysical_leaf_shape_block_within_quire\tDIAGNOSTIC_ONLY\tall frozen scalar families; unseen test classes count as errors; no causal interpretation\n", len(ps), accuracy, baseline, p, repetitions)
	return os.WriteFile(filepath.Join(out, "VISUAL_CONTEXT_CLASSIFICATION.tsv"), []byte(b.String()), 0644)
}

func classify(ps []*page, labels []string) (float64, float64) {
	correct, base, n := 0, 0, 0
	byQ := map[string][]int{}
	for i, p := range ps {
		byQ[p.Quire] = append(byQ[p.Quire], i)
	}
	quires := make([]string, 0, len(byQ))
	for q := range byQ {
		quires = append(quires, q)
	}
	sort.Strings(quires)
	for _, q := range quires {
		test := byQ[q]
		var train []int
		for i, p := range ps {
			if p.Quire != q {
				train = append(train, i)
			}
		}
		counts := map[string]int{}
		for _, i := range train {
			counts[labels[i]]++
		}
		major := ""
		for c, k := range counts {
			if k > counts[major] {
				major = c
			}
		}
		means, sd := trainScale(ps, train)
		cent := map[string][]float64{}
		cn := map[string]int{}
		for _, i := range train {
			c := labels[i]
			if cent[c] == nil {
				cent[c] = make([]float64, len(metrics))
			}
			cn[c]++
			for j, m := range metrics {
				cent[c][j] += (ps[i].Values[m.Name] - means[j]) / sd[j]
			}
		}
		for c := range cent {
			for j := range cent[c] {
				cent[c][j] /= float64(cn[c])
			}
		}
		for _, i := range test {
			truth := labels[i]
			best := ""
			bd := math.Inf(1)
			for c, v := range cent {
				var d float64
				for j, m := range metrics {
					z := (ps[i].Values[m.Name] - means[j]) / sd[j]
					d += (z - v[j]) * (z - v[j])
				}
				if d < bd {
					bd = d
					best = c
				}
			}
			if best == truth && cn[truth] > 0 {
				correct++
			}
			if major == truth {
				base++
			}
			n++
		}
	}
	return div(float64(correct), float64(n)), div(float64(base), float64(n))
}
func trainScale(ps []*page, ix []int) ([]float64, []float64) {
	mu := make([]float64, len(metrics))
	sd := make([]float64, len(metrics))
	for _, i := range ix {
		for j, m := range metrics {
			mu[j] += ps[i].Values[m.Name]
		}
	}
	for j := range mu {
		mu[j] /= float64(len(ix))
	}
	for _, i := range ix {
		for j, m := range metrics {
			d := ps[i].Values[m.Name] - mu[j]
			sd[j] += d * d
		}
	}
	for j := range sd {
		sd[j] = math.Sqrt(sd[j] / float64(max(1, len(ix)-1)))
		if sd[j] == 0 {
			sd[j] = 1
		}
	}
	return mu, sd
}

func bh(p []float64) []float64 {
	type pair struct {
		i int
		p float64
	}
	x := make([]pair, len(p))
	for i, v := range p {
		x[i] = pair{i, v}
	}
	sort.Slice(x, func(i, j int) bool { return x[i].p < x[j].p })
	out := make([]float64, len(p))
	next := 1.0
	for j := len(x) - 1; j >= 0; j-- {
		v := x[j].p * float64(len(x)) / float64(j+1)
		if v > next {
			v = next
		}
		if v > 1 {
			v = 1
		}
		next = v
		out[x[j].i] = v
	}
	return out
}

func writeManifest(out string, files []string) error {
	type entry struct {
		Path, SHA256 string
		Bytes        int
	}
	type manifest struct {
		Analysis, Version                string
		Seed                             int64
		Repetitions, MinTokens, MinLines int
		Inputs                           []entry
		Outputs                          []entry
	}
	m := manifest{Analysis: "visual-context", Version: "1.0.0", Seed: seed, Repetitions: repetitions, MinTokens: minTokens, MinLines: minLines}
	for _, name := range []string{"experiments/fingerprint-v2-task79-v1/canonical-out/line_profiles.json", "experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl", "data/ZL3b-n.txt", "research/visual-context-analyze/main.go", "research/visual-context-analyze/main_test.go"} {
		b, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		m.Inputs = append(m.Inputs, entry{name, fmt.Sprintf("%x", sha256.Sum256(b)), len(b)})
	}
	for _, name := range files {
		b, err := os.ReadFile(filepath.Join(out, name))
		if err != nil {
			return err
		}
		m.Outputs = append(m.Outputs, entry{name, fmt.Sprintf("%x", sha256.Sum256(b)), len(b)})
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(filepath.Join(out, "VISUAL_CONTEXT_RESULTS_MANIFEST.json"), b, 0644)
}
