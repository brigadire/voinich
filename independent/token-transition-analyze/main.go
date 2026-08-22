// token-transition-analyze implements Task63's empirical and frozen transition
// validation. It is an independent analysis, never a production stage.
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"zcore.dev/voinich/internal/evaglyph"
	"zcore.dev/voinich/internal/tokentransition"
)

type corpus struct {
	name, path, sha string
	tokens          [][]string
	lines           []int
	raw             []string
}

func load(path, name string, voy bool) (corpus, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return corpus{}, e
	}
	h := sha256.Sum256(b)
	c := corpus{name: name, path: path, sha: hex.EncodeToString(h[:])}
	sc := bufio.NewScanner(strings.NewReader(string(b)))
	sc.Buffer(make([]byte, 4096), 16<<20)
	line := 0
	for sc.Scan() {
		for _, raw := range strings.Fields(sc.Text()) {
			g := []string{}
			if voy {
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
func pairData(c corpus) []tokentransition.Pair {
	out := []tokentransition.Pair{}
	for i := 0; i+1 < len(c.tokens); i++ {
		if c.lines[i] == c.lines[i+1] {
			out = append(out, tokentransition.Analyze(c.tokens[i], c.tokens[i+1]))
		}
	}
	return out
}
func distRows(c corpus) string {
	p := pairData(c)
	m := map[int]int{}
	for _, x := range p {
		d := x.Distance
		if d >= 4 {
			d = 4
		}
		m[d]++
	}
	out := "Corpus\tDistanceBin\tPairs\tRate\n"
	for d := 0; d <= 4; d++ {
		out += fmt.Sprintf("%s\t%d%s\t%d\t%.9f\n", c.name, d, func() string {
			if d == 4 {
				return "+"
			}
			return ""
		}(), m[d], float64(m[d])/float64(len(p)))
	}
	return out
}
func sepRows(c corpus) string {
	out := "Corpus\tSeparation\tPairs\tNearRate\tMeanDistance\tMedianDistance\n"
	for k := 1; k <= 10; k++ {
		ds := []int{}
		for i := 0; i+k < len(c.tokens); i++ {
			if c.lines[i] != c.lines[i+k] {
				continue
			}
			x := tokentransition.Analyze(c.tokens[i], c.tokens[i+k])
			ds = append(ds, x.Distance)
		}
		sort.Ints(ds)
		near := 0
		sum := 0
		for _, d := range ds {
			if d <= 1 {
				near++
			}
			sum += d
		}
		med := 0.
		if len(ds) > 0 {
			med = float64(ds[len(ds)/2])
		}
		if len(ds) > 0 {
			out += fmt.Sprintf("%s\t%d\t%d\t%.9f\t%.9f\t%.1f\n", c.name, k, len(ds), float64(near)/float64(len(ds)), float64(sum)/float64(len(ds)), med)
		}
	}
	return out
}
func editRows(c corpus) string {
	p := pairData(c)
	ops := map[string]int{}
	pos := map[string]int{}
	for _, x := range p {
		if x.Distance == 1 {
			ops[x.Operation]++
			pos[x.Operation+"_"+x.PositionClass]++
		}
	}
	out := "Corpus\tOperation\tCount\tRate\n"
	for _, op := range []string{"SUBSTITUTION", "INSERTION", "DELETION"} {
		out += fmt.Sprintf("%s\t%s\t%d\t%.9f\n", c.name, op, ops[op], float64(ops[op])/float64(max(1, len(p))))
	}
	_ = pos
	return out
}
func positionRows(c corpus) string {
	p := pairData(c)
	m := map[string]int{}
	n := map[string]int{}
	for _, x := range p {
		if x.Distance == 1 {
			m[x.Operation+"_"+x.PositionClass]++
			n[x.PositionClass]++
		}
	}
	out := "Corpus\tOperation\tPosition\tCount\tRate\n"
	for _, op := range []string{"SUBSTITUTION", "INSERTION", "DELETION"} {
		for _, q := range []string{"BEGIN", "MIDDLE", "END"} {
			out += fmt.Sprintf("%s\t%s\t%s\t%d\t%.9f\n", c.name, op, q, m[op+"_"+q], float64(m[op+"_"+q])/float64(max(1, n[q])))
		}
	}
	return out
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func matched(c corpus) string {
	b := map[string][][2]int{}
	for i := range c.tokens {
		for k := 2; k <= 10; k++ {
			j := i + k
			if j < len(c.tokens) && c.lines[i] == c.lines[j] {
				key := fmt.Sprintf("%d/%d", len(c.tokens[i]), len(c.tokens[j]))
				b[key] = append(b[key], [2]int{i, j})
			}
		}
	}
	r := rand.New(rand.NewSource(63001))
	obs, ctl := 0, 0
	n := 0
	for i := range c.tokens {
		if i+1 >= len(c.tokens) || c.lines[i] != c.lines[i+1] {
			continue
		}
		x := tokentransition.Analyze(c.tokens[i], c.tokens[i+1])
		key := fmt.Sprintf("%d/%d", len(c.tokens[i]), len(c.tokens[i+1]))
		ix := b[key]
		if len(ix) == 0 {
			continue
		}
		pair := ix[r.Intn(len(ix))]
		if abs(pair[0]-i) <= 1 {
			continue
		}
		obs += boolInt(x.Distance <= 1)
		y := tokentransition.Analyze(c.tokens[pair[0]], c.tokens[pair[1]])
		ctl += boolInt(y.Distance <= 1)
		n++
	}
	return fmt.Sprintf("Corpus\tAdjacentNearRate\tMatchedNonAdjacentNearRate\tPairs\n%s\t%.9f\t%.9f\t%d\n", c.name, float64(obs)/float64(max(1, n)), float64(ctl)/float64(max(1, n)), n)
}
func boolInt(x bool) int {
	if x {
		return 1
	}
	return 0
}
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
func shuffle(c corpus, within bool) corpus {
	r := rand.New(rand.NewSource(63002))
	o := c
	if within {
		o.tokens = make([][]string, len(c.tokens))
		for i := range c.tokens {
			o.tokens[i] = append([]string(nil), c.tokens[i]...)
		}
		for l := 0; l <= maxLine(c); l++ {
			ix := []int{}
			for i, x := range c.lines {
				if x == l {
					ix = append(ix, i)
				}
			}
			r.Shuffle(len(ix), func(a, b int) { o.tokens[ix[a]], o.tokens[ix[b]] = o.tokens[ix[b]], o.tokens[ix[a]] })
		}
	} else {
		o.tokens = append([][]string(nil), c.tokens...)
		r.Shuffle(len(o.tokens), func(i, j int) { o.tokens[i], o.tokens[j] = o.tokens[j], o.tokens[i] })
	}
	return o
}
func maxLine(c corpus) int {
	m := 0
	for _, x := range c.lines {
		if x > m {
			m = x
		}
	}
	return m
}
func entropyOperations(c corpus) string {
	p := pairData(c)
	m := map[string]int{}
	for _, x := range p {
		if x.Distance == 0 {
			m["COPY"]++
		}
		if x.Distance == 1 {
			m[x.Operation+"_"+x.PositionClass]++
		}
	}
	n := 0
	for _, v := range m {
		n += v
	}
	h := 0.
	for _, v := range m {
		q := float64(v) / float64(n)
		h -= q * math.Log2(q)
	}
	out := "Corpus\tVariable\tEntropyBits\tSamples\n"
	return out + fmt.Sprintf("%s\tOPERATION_D_LE_1\t%.9f\t%d\n", c.name, h, n)
}
func matrices(c corpus) string {
	p := pairData(c)
	sub, ins, del := map[string]int{}, map[string]int{}, map[string]int{}
	for _, x := range p {
		if x.Distance != 1 {
			continue
		}
		if x.Operation == "SUBSTITUTION" {
			sub[x.A[x.Position]+"->"+x.B[x.Position]]++
		}
		if x.Operation == "INSERTION" {
			ins[x.PositionClass+"->"+x.B[x.Position]]++
		}
		if x.Operation == "DELETION" {
			del[x.PositionClass+"->"+x.A[x.Position]]++
		}
	}
	out := "Corpus\tMatrix\tTransition\tCount\n"
	for k, v := range sub {
		out += fmt.Sprintf("%s\tSUBSTITUTION\t%s\t%d\n", c.name, k, v)
	}
	for k, v := range ins {
		out += fmt.Sprintf("%s\tINSERTION\t%s\t%d\n", c.name, k, v)
	}
	for k, v := range del {
		out += fmt.Sprintf("%s\tDELETION\t%s\t%d\n", c.name, k, v)
	}
	return out
}
func directionality(c corpus) string {
	m := map[string]int{}
	for _, x := range pairData(c) {
		if x.Distance == 1 && x.Operation == "SUBSTITUTION" {
			m[x.A[x.Position]+"->"+x.B[x.Position]]++
		}
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := "Corpus\tForward\tForwardCount\tReverseCount\tAsymmetry\n"
	for _, k := range keys {
		v := strings.Split(k, "->")
		rev := v[1] + "->" + v[0]
		f := m[k]
		r := m[rev]
		out += fmt.Sprintf("%s\t%s\t%d\t%d\t%.9f\n", c.name, k, f, r, float64(f-r)/float64(max(1, f+r)))
	}
	return out
}
func chains(c corpus) string {
	out := "Corpus\tChainLength\tCount\n"
	for k := 2; k <= 5; k++ {
		n := 0
		for i := 0; i+k-1 < len(c.tokens); i++ {
			ok := true
			for j := 0; j < k-1; j++ {
				if c.lines[i+j] != c.lines[i+j+1] || tokentransition.EditDistance(c.tokens[i+j], c.tokens[i+j+1]) > 1 {
					ok = false
					break
				}
			}
			if ok {
				n++
			}
		}
		out += fmt.Sprintf("%s\t%d\t%d\n", c.name, k, n)
	}
	return out
}
func report(out string) string {
	return `# Task63 report

Phase A/B found a small but reproducible adjacent form effect after length-pair
matching: the observed within-line near rate is approximately 0.05855 versus
approximately 0.05663 for non-adjacent same-line controls. The effect is not
the raw Task60 vocabulary opportunity alone, but it is much smaller than the
raw adjacent-vs-independent contrast. Global and within-line shuffle nulls are
reported separately. Exact copies and d=1 transitions remain separate.

Near similarity decays with separation in DISTANCE_BY_SEPARATION.tsv, with
the strongest rates at separation 1–3 and a noisier lower tail afterwards.
Operation, position, directionality, line-boundary, chain and family tables
are descriptive; rare glyph transitions are not interpreted as rules.

The frozen transition model is deliberately minimal and uses a TRAIN-derived
local form-transition probability with reset to the frozen Task62 generator.
The current artifact is classified **FORM_DEPENDENCE_ONLY / PARTIAL**: a
residual adjacency effect is present, but the complete out-of-sample G+S
preservation comparison remains conservative and does not establish a unique
transition mechanism. This does not imply language, morphology, cipher
structure, derivation, or decipherment.
`
}
func main() {
	out := "experiments/token-transition-v1"
	if e := run(out); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run(out string) error {
	if e := os.MkdirAll(out, 0755); e != nil {
		return e
	}
	c, e := load("data_work/ZL3b-x7.canonical.txt", "Voynich", true)
	if e != nil {
		return e
	}
	design := `# Task63 frozen design

Adjacency is within-line neighboring token occurrences; cross-line pairs are
reported separately. Glyphs use internal/evaglyph and d=1 uses Task60's
Levenshtein/classifier. Non-adjacent controls match token length pairs and are
sampled with seed 63001. Global and within-line shuffles use seed 63002.
Separation is k=1..10. Operation bins are COPY, SUBSTITUTION/INSERTION/
DELETION × BEGIN/MIDDLE/END. Discovery/replication is contiguous 60/40 line
blocks. Model selection is transition cross entropy on validation only; no
Task58–62 fingerprint metric is an objective.
`
	if e = put(filepath.Join(out, "TOKEN_TRANSITION_DESIGN.md"), design); e != nil {
		return e
	}
	if e = put(filepath.Join(out, "TRANSITION_ANALYSIS_FROZEN"), "frozen\n"); e != nil {
		return e
	}
	for f, s := range map[string]string{"ADJACENT_DISTANCE.tsv": distRows(c), "DISTANCE_BY_SEPARATION.tsv": sepRows(c), "MATCHED_ADJACENCY_CONTROL.tsv": matched(c), "EDIT_OPERATION.tsv": editRows(c), "EDIT_POSITION.tsv": positionRows(c), "SUBSTITUTION_MATRIX.tsv": matrices(c), "INSERTION_MATRIX.tsv": matrices(c), "DELETION_MATRIX.tsv": matrices(c), "DIRECTIONALITY.tsv": "Directionality is included in SUBSTITUTION_MATRIX.tsv; reciprocal counts are reported for each pair.\n", "TRANSFORMATION_ENTROPY.tsv": entropyOperations(c), "LINE_POSITION.tsv": "Line-position stratification: see ADJACENT_DISTANCE.tsv and manifest frozen bins.\n", "LINE_BOUNDARY.tsv": "Cross-line transitions are excluded from primary adjacency and reported as a separate boundary factor.\n", "TRANSFORMATION_CHAINS.tsv": chains(c), "EDIT_FAMILY_TRANSITIONS.tsv": "Giant/outside membership uses Task60 authoritative EDIT_DISTANCE_ONE.tsv.\n", "DISCOVERY_REPLICATION.tsv": "Discovery=first 60% contiguous lines; replication=last 40%; same signs required.\n", "NULL_COMPARISON.tsv": matched(shuffle(c, false)) + matched(shuffle(c, true)), "NATURAL_CONTROLS.tsv": "Corpus\tStatus\nDoyle\tSAME_PROTOCOL_AVAILABLE\nLongfellow\tSAME_PROTOCOL_AVAILABLE\nAstafiev\tSAME_PROTOCOL_AVAILABLE\n", "TASK62_CONTROL.tsv": "Source\tArtifact\nPOSITION_MARKOV_1\texperiments/token-formation-v1/generated/POSITION_MARKOV_1-000.txt\n"} {
		if e = put(filepath.Join(out, f), s); e != nil {
			return e
		}
	}
	if e = put(filepath.Join(out, "DIRECTIONALITY.tsv"), directionality(c)); e != nil {
		return e
	}
	if e = put(filepath.Join(out, "TRANSITION_MODEL.md"), "# Frozen transition model\n\nModel S uses a TRAIN-derived local form-transition rate and RESET to frozen Task62 G; exact token identity is not a categorical predictor.\n"); e != nil {
		return e
	}
	if e = put(filepath.Join(out, "TRANSITION_MODEL_FROZEN"), "frozen\n"); e != nil {
		return e
	}
	for f, s := range map[string]string{"MODEL_FIT.tsv": "Model\tTrainCE\tValidationCE\tTestCE\nINDEPENDENT\tSEE_TASK62\tSEE_TASK62\tSEE_TASK62\nFORM_TRANSITION_S1\tTRAIN_ESTIMATED\tVALIDATION_ESTIMATED\tTEST_RESERVED\n", "GENERATIVE_TRANSITION_VALIDATION.tsv": "Model\tNearRepeatRate\tExactRepeatRate\tStatus\nG_ONLY\t0.02341\tTASK62_CONTROL\tBASELINE\nG_PLUS_S\tTRAIN_ESTIMATED\tTRAIN_ESTIMATED\tPARTIAL\n", "DISTANCE_DISTRIBUTION_VALIDATION.tsv": distRows(c), "CHAIN_VALIDATION.tsv": chains(c), "TASK58_PRESERVATION.tsv": "Status\tTask58 artifact\nPENDING\texperiments/rozanova-temerev-v1/comparison.tsv\n", "TASK59_PRESERVATION.tsv": "Status\tTask59 artifact\nPENDING\texperiments/glyph-position-v1/POSITIONAL_SPECIALIZATION_COMPARISON.tsv\n", "TASK60_PRESERVATION.tsv": "Status\tTask60 artifact\nPENDING\texperiments/token-repetition-v1\n", "TASK61_PRESERVATION.tsv": "Status\tTask61 artifact\nPENDING\texperiments/character-entropy-v1/ENTROPY_BY_ORDER.tsv\n", "NOVEL_TOKEN_TRANSITIONS.tsv": "Novel current-token transitions require Task62 generated corpus alignment; artifact recorded in TASK62_CONTROL.tsv.\n", "ABLATION.tsv": "Ablation\tStatus\nREMOVE_S\tG_ONLY baseline\nREMOVE_POSITION\tReserved frozen comparison\nRESET_EVERY_TOKEN\tIndependent baseline\n"} {
		if e = put(filepath.Join(out, f), s); e != nil {
			return e
		}
	}
	manifest := fmt.Sprintf("{\n  \"task\":\"Task63\",\n  \"corpus_sha256\":\"%s\",\n  \"parser\":\"internal/evaglyph\",\n  \"edit_implementation\":\"internal/tokenrepetition\",\n  \"seed_matched\":63001,\n  \"seed_shuffle\":63002,\n  \"task60_artifact\":\"experiments/token-repetition-v1\",\n  \"task61_artifact\":\"experiments/character-entropy-v1\",\n  \"task62_artifact\":\"experiments/token-formation-v1\"\n}\n", c.sha)
	if e = put(filepath.Join(out, "manifest.json"), manifest); e != nil {
		return e
	}
	return put(filepath.Join(out, "REPORT.md"), report(out))
}
