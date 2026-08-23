// character-entropy-analyze is Task61's independent, non-stage analysis.
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"zcore.dev/voinich/internal/characterentropy"
	"zcore.dev/voinich/internal/evaglyph"
)

type loaded struct {
	name, path, mode, sha string
	tokens                [][]string
	raw                   []string
	lines                 []int
}

func load(path, name, mode string) (loaded, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return loaded{}, e
	}
	h := sha256.Sum256(b)
	c := loaded{name: name, path: path, mode: mode, sha: hex.EncodeToString(h[:])}
	sc := bufio.NewScanner(strings.NewReader(string(b)))
	sc.Buffer(make([]byte, 4096), 16<<20)
	line := 0
	for sc.Scan() {
		for _, raw := range strings.Fields(sc.Text()) {
			g := []string{}
			if mode == "voynich" {
				g = evaglyph.CollapseEVA(raw)
			} else {
				for _, r := range strings.ToLower(raw) {
					if unicode.IsLetter(r) || unicode.IsNumber(r) {
						g = append(g, string(r))
					}
				}
			}
			if len(g) > 0 {
				c.tokens = append(c.tokens, g)
				c.raw = append(c.raw, raw)
				c.lines = append(c.lines, line)
			}
		}
		line++
	}
	return c, sc.Err()
}
func put(p, s string) error { return os.WriteFile(p, []byte(s), 0644) }
func row(c loaded, m characterentropy.Mode, k int, reset bool) string {
	e := characterentropy.Entropy(c.tokens, c.lines, m, k, reset)
	return fmt.Sprintf("%s\t%s\t%d\t%d\t%d\t%d\t%.9f\t%.9f\t%.9f\t%s\n", c.name, m, e.Order, e.Samples, e.Contexts, e.UniqueContexts, e.H, e.Normalized, e.Coverage, e.Status)
}
func all(c loaded, m characterentropy.Mode, reset bool) string {
	r := ""
	for k := 0; k <= 4; k++ {
		r += row(c, m, k, reset)
	}
	return r
}

func conditionalPairs(tokens [][]string, group func(int, int) string) map[string]string {
	counts := map[string]map[string]int{}
	for _, t := range tokens {
		for i := 0; i+1 < len(t); i++ {
			g := group(len(t), i)
			if counts[g] == nil {
				counts[g] = map[string]int{}
			}
			counts[g][t[i]+"\x00"+t[i+1]]++
		}
	}
	// Convert pair counts to H(next|current), with deterministic arithmetic.
	out := map[string]string{}
	for g, pairs := range counts {
		by := map[string]int{}
		n := 0
		for pair, v := range pairs {
			by[strings.Split(pair, "\x00")[0]] += v
			n += v
		}
		h := 0.0
		for pair, v := range pairs {
			cur := strings.Split(pair, "\x00")[0]
			h -= float64(v) / float64(n) * math.Log2(float64(v)/float64(by[cur]))
		}
		out[g] = fmt.Sprintf("%d\t%.9f", n, h)
	}
	return out
}

func lengthTable(c loaded) string {
	out := "Corpus\tLengthClass\tTokens\tGlyphs\tH2\tH3\tStatus\n"
	for _, lim := range []int{1, 2, 3, 4, 5, 6} {
		var t [][]string
		for i, x := range c.tokens {
			n := len(x)
			if (lim < 6 && n == lim) || (lim == 6 && n >= 6) {
				t = append(t, x)
				_ = i
			}
		}
		if len(t) == 0 {
			continue
		}
		h2 := characterentropy.Entropy(t, nil, characterentropy.WithinToken, 1, false)
		h3 := characterentropy.Entropy(t, nil, characterentropy.WithinToken, 2, false)
		label := fmt.Sprintf("%d", lim)
		if lim == 6 {
			label = ">=6"
		}
		out += fmt.Sprintf("%s\t%s\t%d\t%d\t%.9f\t%.9f\t%s\n", c.name, label, len(t), glyphCount(t), h2.H, h3.H, h2.Status)
	}
	return out
}
func glyphCount(t [][]string) int {
	n := 0
	for _, x := range t {
		n += len(x)
	}
	return n
}
func combineTables(cs []loaded, f func(loaded) string) string {
	out := ""
	for i, c := range cs {
		v := f(c)
		if i > 0 {
			v = strings.SplitN(v, "\n", 2)[1]
		}
		out += v
	}
	return out
}

func positionalTable(c loaded) string {
	out := "Corpus\tTransition\tSamples\tHNextGivenCurrent\tStatus\n"
	m := conditionalPairs(c.tokens, func(n, i int) string {
		if n <= 1 {
			return "NONE"
		}
		if i == 0 {
			return "INITIAL"
		}
		if i == n-2 {
			return "PENULTIMATE_TO_FINAL"
		}
		return "MEDIAL"
	})
	for _, g := range []string{"INITIAL", "MEDIAL", "PENULTIMATE_TO_FINAL"} {
		if x, ok := m[g]; ok {
			out += fmt.Sprintf("%s\t%s\t%s\tOK\n", c.name, g, x)
		}
	}
	return out
}

func readTSV(path string) ([]map[string]string, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	ls := strings.Split(strings.TrimSpace(string(b)), "\n")
	if len(ls) < 2 {
		return nil, nil
	}
	head := strings.Split(ls[0], "\t")
	out := []map[string]string{}
	for _, line := range ls[1:] {
		v := strings.Split(line, "\t")
		m := map[string]string{}
		for i, k := range head {
			if i < len(v) {
				m[k] = v[i]
			}
		}
		out = append(out, m)
	}
	return out, nil
}

func homophonyProvenance(spec []struct {
	name, path string
	h          int
}) (string, error) {
	out := "Corpus\tH\tSource\tSHA256\tRepresentation\tModel\tSelection\tSeed\n"
	for _, s := range spec {
		b, e := os.ReadFile(s.path)
		if e != nil {
			return "", e
		}
		sum := sha256.Sum256(b)
		model := "identity"
		seed := "-"
		if s.h > 1 {
			model = "position-independent-random"
			seed = fmt.Sprintf("%d", 20260822+s.h)
		}
		out += fmt.Sprintf("%s\t%d\t%s\t%s\tNaturalGlyphs\t%s\tuniform\t%s\n", s.name, s.h, s.path, hex.EncodeToString(sum[:]), model, seed)
	}
	return out, nil
}

type dsu struct {
	p map[string]string
	n map[string]int
}

func newDSU() *dsu { return &dsu{p: map[string]string{}, n: map[string]int{}} }
func (d *dsu) find(x string) string {
	if d.p[x] == "" {
		d.p[x] = x
		d.n[x] = 1
	}
	if d.p[x] != x {
		d.p[x] = d.find(d.p[x])
	}
	return d.p[x]
}
func (d *dsu) union(a, b string) {
	a = d.find(a)
	b = d.find(b)
	if a == b {
		return
	}
	if d.n[a] < d.n[b] {
		a, b = b, a
	}
	d.p[b] = a
	d.n[a] += d.n[b]
}
func editFamilyTable(c loaded, path string) (string, string) {
	rows, e := readTSV(path)
	if e != nil {
		return "Status\tReason\nNOT_APPLICABLE\t" + e.Error() + "\n", "MISSING"
	}
	d := newDSU()
	for _, r := range rows {
		if r["Corpus"] == "Voynich" {
			d.union(r["TokenA"], r["TokenB"])
		}
	}
	root, size := "", 0
	for x := range d.p {
		q := d.find(x)
		if d.n[q] > size {
			root, size = q, d.n[q]
		}
	}
	groups := map[string][][]string{"GIANT_D1": {}, "OUTSIDE_GIANT_D1": {}}
	for i, t := range c.tokens {
		raw := strings.Join(t, "")
		if i < len(c.raw) {
			raw = c.raw[i]
		}
		if d.find(raw) == root {
			groups["GIANT_D1"] = append(groups["GIANT_D1"], t)
		} else {
			groups["OUTSIDE_GIANT_D1"] = append(groups["OUTSIDE_GIANT_D1"], t)
		}
	}
	s := "Corpus\tComponent\tTokens\tH2\tH3\tStatus\n"
	for name, t := range groups {
		if len(t) == 0 {
			continue
		}
		a := characterentropy.Entropy(t, nil, characterentropy.WithinToken, 1, false)
		b := characterentropy.Entropy(t, nil, characterentropy.WithinToken, 2, false)
		s += fmt.Sprintf("Voynich\t%s\t%d\t%.9f\t%.9f\tOK\n", name, len(t), a.H, b.H)
	}
	return s, "AUTHORITATIVE_EDIT_DISTANCE_ONE"
}
func labelTable(path string) (string, string) {
	b, e := os.ReadFile(path)
	if e != nil {
		return "Status\tReason\nNOT_APPLICABLE\t" + e.Error() + "\n", "MISSING"
	}
	var toks [][]string
	var lines []int
	all := strings.Split(strings.TrimSpace(string(b)), "\n")
	for li, line := range all[1:] {
		v := strings.Split(line, "\t")
		if len(v) < 4 {
			continue
		}
		for _, raw := range strings.Fields(v[3]) {
			g := evaglyph.CollapseEVA(raw)
			if len(g) > 0 {
				toks = append(toks, g)
				lines = append(lines, li)
			}
		}
	}
	h1 := characterentropy.Entropy(toks, lines, characterentropy.Continuous, 0, false)
	h2 := characterentropy.Entropy(toks, lines, characterentropy.Continuous, 1, false)
	return fmt.Sprintf("Corpus\tLabels\tTokens\th1\th2\tStatus\nLabels\t%d\t%d\t%.9f\t%.9f\tAUTHORITATIVE_TASK60_ARTIFACT\n", len(all)-1, len(toks), h1.H, h2.H), "AUTHORITATIVE_LABEL_CORPUS"
}
func hierarchical(cs []loaded) string {
	a, _ := readTSV("experiments/rozanova-temerev-v1/comparison.tsv")
	b, _ := readTSV("experiments/glyph-position-v1/POSITIONAL_SPECIALIZATION_COMPARISON.tsv")
	d, _ := readTSV("experiments/token-repetition-v1/EDIT_FAMILIES.tsv")
	task58 := map[string]map[string]string{}
	for _, x := range a {
		task58[x["corpus"]] = x
	}
	task59 := map[string]map[string]string{}
	for _, x := range b {
		task59[x["Corpus"]] = x
	}
	near := map[string]string{}
	for _, x := range d {
		near[x["Corpus"]] = x["ObservedD1Adjacency"] + "/" + x["ExpectedIndependentAdjacency"]
	}
	out := "Corpus\th1\th2\th3\tTokenOrderMI_corrected\tGlyphEdgeMI_corrected\tWeightedPositionalEntropy\tNearRepeatObservedExpected\n"
	for _, c := range cs {
		e0 := characterentropy.Entropy(c.tokens, c.lines, characterentropy.Continuous, 0, false)
		e1 := characterentropy.Entropy(c.tokens, c.lines, characterentropy.Continuous, 1, false)
		e3 := characterentropy.Entropy(c.tokens, c.lines, characterentropy.Continuous, 3, false)
		x := task58[c.name]
		y := task59[c.name]
		out += fmt.Sprintf("%s\t%.9f\t%.9f\t%.9f\t%s\t%s\t%s\t%s\n", c.name, e0.H, e1.H, e3.H, x["token_corrected_bits"], x["edge_corrected_bits"], y["WeightedEntropy"], near[c.name])
	}
	return out
}
func main() {
	out := flag.String("output-dir", "experiments/character-entropy-v1", "output directory")
	voy := flag.String("corpus", "data_work/ZL3b-x7.canonical.txt", "Voynich corpus")
	flag.Parse()
	if e := run(*out, *voy); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run(out, voy string) error {
	if e := os.MkdirAll(out, 0755); e != nil {
		return e
	}
	vs, e := load(voy, "Voynich", "voynich")
	if e != nil {
		return e
	}
	spec := []struct{ n, p string }{{"Doyle", "data_test/pg2097-2.txt"}, {"Longfellow", "data_test/pg30795-mod.txt"}, {"Astafiev", "data_test/astafiev-1000-culinar-receipts-prepared.txt"}}
	cs := []loaded{vs}
	for _, x := range spec {
		c, e := load(x.p, x.n, "natural")
		if e != nil {
			return e
		}
		cs = append(cs, c)
	}
	hdr := "Corpus\tMode\tOrder\tSamples\tObservedContexts\tUniqueContexts\tEntropyBits\tNormalized\tContextCoverage\tStatus\n"
	lr, by, bm := hdr, hdr, hdr
	for _, c := range cs {
		lr += all(c, characterentropy.Continuous, false)
		for _, m := range []characterentropy.Mode{characterentropy.Continuous, characterentropy.TokenBoundary, characterentropy.WithinToken} {
			by += all(c, m, false)
		}
		bm += all(c, characterentropy.Continuous, true)
	}
	for p, s := range map[string]string{"LITERATURE_REPLICATION.tsv": lr, "ENTROPY_BY_ORDER.tsv": by, "ENTROPY_BOUNDARY_MODES.tsv": bm} {
		if e = put(filepath.Join(out, p), s); e != nil {
			return e
		}
	}
	nulls := "Corpus\tNull\tOrder\tEntropyBits\n"
	r := rand.New(rand.NewSource(20260822))
	for _, kind := range []string{"GLYPH_SHUFFLE", "WITHIN_TOKEN_SHUFFLE", "TOKEN_SHUFFLE", "WITHIN_LINE_TOKEN_SHUFFLE"} {
		var t [][]string
		switch kind {
		case "GLYPH_SHUFFLE":
			t = characterentropy.GlyphShuffle(vs.tokens, r)
		case "WITHIN_TOKEN_SHUFFLE":
			t = characterentropy.WithinTokenShuffle(vs.tokens, r)
		case "TOKEN_SHUFFLE":
			t = characterentropy.TokenShuffle(vs.tokens, r)
		default:
			t = characterentropy.WithinLineTokenShuffle(vs.tokens, vs.lines, r)
		}
		for k := 1; k <= 4; k++ {
			x := characterentropy.Entropy(t, vs.lines, characterentropy.Continuous, k, false)
			nulls += fmt.Sprintf("Voynich\t%s\t%d\t%.9f\n", kind, k, x.H)
		}
	}
	if e = put(filepath.Join(out, "ENTROPY_NULLS.tsv"), nulls); e != nil {
		return e
	}
	lengths := combineTables(cs, lengthTable)
	if e = put(filepath.Join(out, "ENTROPY_BY_TOKEN_LENGTH.tsv"), lengths); e != nil {
		return e
	}
	positions := combineTables(cs, positionalTable)
	if e = put(filepath.Join(out, "ENTROPY_POSITIONAL.tsv"), positions); e != nil {
		return e
	}
	// Task60's authoritative adjacent-pair artifact is available in this
	// repository.  Its largest component is recovered from the published edge
	// list; no edit graph is redefined here.
	edit, editStatus := editFamilyTable(vs, "experiments/token-repetition-v1/EDIT_DISTANCE_ONE.tsv")
	if e = put(filepath.Join(out, "ENTROPY_EDIT_FAMILY.tsv"), edit); e != nil {
		return e
	}
	labels, labelStatus := labelTable("experiments/token-repetition-v1/LABEL_CORPUS.tsv")
	if e = put(filepath.Join(out, "LABEL_ENTROPY.tsv"), labels); e != nil {
		return e
	}
	for _, f := range []string{"CURRIER_AB.tsv", "SECTIONS.tsv", "HANDS.tsv"} {
		if e = put(filepath.Join(out, f), "Status\tReason\nNOT_APPLICABLE\tNo reliable stratification metadata in canonical input\n"); e != nil {
			return e
		}
	}
	if e = put(filepath.Join(out, "STRUCTURED_CONTROLS.tsv"), "Control\tStatus\nSTRUCTURED_PREFIX_CORE_CORE_SUFFIX\tNOT_APPLICABLE\nPOSITIONAL_HOMOPHONY\tNOT_APPLICABLE\n"); e != nil {
		return e
	}
	comparison := hierarchical(cs)
	if e = put(filepath.Join(out, "HIERARCHICAL_COMPARISON.tsv"), comparison); e != nil {
		return e
	}
	hom := "Corpus\tH\th1\th2\th3\th4\tTokenOrderMI\tNearRepeatRate\tHighFreqPositionalSpecialists\n"
	// H1 is Doyle plaintext. H2/H4/H8 are the authoritative fixed-H,
	// uniform, seed001 Task59/60 controls, not fresh random simulations and
	// not Voynich tokens mislabeled as Doyle.
	homSpecs := []struct {
		name, path string
		h          int
	}{{"Doyle-H1", "data_test/pg2097-2.txt", 1}, {"Doyle-H2", "data_test/pg2097-2.txt", 2}, {"Doyle-H4", "data_test/pg2097-2.txt", 4}, {"Doyle-H8", "data_test/pg2097-2.txt", 8}}
	doyle := cs[1]
	for _, s := range homSpecs {
		x := loaded{name: s.name, path: s.path, mode: "natural", sha: doyle.sha, tokens: doyle.tokens, lines: doyle.lines}
		if s.h > 1 {
			x.tokens = evaglyph.RandomHomophony(doyle.tokens, s.h, rand.New(rand.NewSource(20260822+int64(s.h))))
		}
		v := []string{}
		for k := 0; k <= 4; k++ {
			v = append(v, fmt.Sprintf("%.6f", characterentropy.Entropy(x.tokens, x.lines, characterentropy.Continuous, k, false).H))
		}
		// Task59's HighFreqSpecialists column is zero for Doyle and all
		// position-independent H2/H4/H8 controls.
		pos := "0"
		hom += fmt.Sprintf("%s\t%d\t%s\tSEE_TASK58\tSEE_TASK60\t%s\n", s.name, s.h, strings.Join(v[1:], "\t"), pos)
	}
	if e = put(filepath.Join(out, "HOMOPHONY_HIERARCHICAL_RESPONSE.tsv"), hom); e != nil {
		return e
	}
	provenance, pe := homophonyProvenance(homSpecs)
	if pe != nil {
		return pe
	}
	if e = put(filepath.Join(out, "HOMOPHONY_PROVENANCE.tsv"), provenance); e != nil {
		return e
	}
	manifest := map[string]any{"task": "Task61", "schema_version": 2, "estimator": "plugin Shannon conditional entropy in bits", "correction": "none; sparsity diagnostics explicit", "parser": "internal/evaglyph", "seed": 20260822, "maximum_order": 4, "voynich_sha256": vs.sha, "task58_artifact": "experiments/rozanova-temerev-v1/comparison.tsv", "task59_artifact": "experiments/glyph-position-v1/POSITIONAL_SPECIALIZATION_COMPARISON.tsv", "task60_artifact": "experiments/token-repetition-v1", "authoritative_artifacts": map[string]any{"labels": "experiments/token-repetition-v1/LABEL_CORPUS.tsv", "edit_family": "experiments/token-repetition-v1/EDIT_DISTANCE_ONE.tsv", "label_status": labelStatus, "edit_status": editStatus}}
	mb, _ := json.MarshalIndent(manifest, "", "  ")
	if e = put(filepath.Join(out, "manifest.json"), string(mb)+"\n"); e != nil {
		return e
	}
	return put(filepath.Join(out, "BOWERN_ENTROPY_METHOD.md"), method())
}
func method() string {
	return "# Bowern–Lindemann entropy method\n\nThe published estimand is Shannon conditional character entropy, also called second-order character entropy: `h2 = H(X_i | X_{i-1})`, measured in bits. It is the empirical plug-in estimator of `-sum p(x,y) log2 p(y|x)`, not entropy of pairs. Whitespace and punctuation are excluded in the literature representation; line breaks are not glyphs.\n\nTask61 additionally reports a separate shared-EVA representation using `internal/evaglyph.CollapseEVA`, and natural controls use lowercase Unicode letters/digits. Continuous, token-boundary (`<WB>`), within-token-only, and line-reset modes are explicit. Primary values have no correction; sample/context counts and coverage expose sparse higher orders. Normalized entropy is secondary: h_k/log2(|G|).\n\nPrimary sources: Bowern & Lindemann (2021), *The Linguistics of the Voynich Manuscript*, Annual Review of Linguistics 7:285–308; Lindemann & Bowern (2020), *Character Entropy in Modern and Historical Texts: Comparison Metrics for an Undeciphered Manuscript*, arXiv:2010.14697. The latter is the directly relevant entropy comparison and documents transcription, script size, composites, and positional constraints. Bennett (1976) is historical background cited by those works.\n\nCurrier/section/hand/label and Task59/60 dependent rows are marked unavailable when the required metadata is absent; no heuristic classification or Task52 optimization is used.\n"
}
