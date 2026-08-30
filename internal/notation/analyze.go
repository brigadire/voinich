package notation

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

var RarefactionCheckpoints = []int{5000, 10000, 20000, 39380}

// MetricRegistryVersion is the frozen version tag for the generic metric
// registry emitted by Analyze. Compare fails closed rather than joining two
// fingerprints produced under different registry versions (adversarial
// test A2).
const MetricRegistryVersion = "generic-metrics-1.0"

func Analyze(records []Record) (Fingerprint, error) {
	if err := Validate(records); err != nil {
		return Fingerprint{}, err
	}
	hash, err := CanonicalSHA256(records)
	if err != nil {
		return Fingerprint{}, err
	}
	fp := Fingerprint{SchemaVersion: "notation-fingerprint-1.0", CorpusID: records[0].CorpusID, Representation: records[0].Representation, InputSHA256: hash, RecordCount: len(records), Metadata: map[string]string{"analyzer": "generic/no-corpus-branches", "metric_registry_version": MetricRegistryVersion}}
	fp.Metrics = append(fp.Metrics, glyphMetrics(records)...)
	fp.Metrics = append(fp.Metrics, tokenMetrics(records)...)
	fp.Metrics = append(fp.Metrics, sequenceMetrics(records)...)
	fp.Metrics = append(fp.Metrics, lineMetrics(records)...)
	fp.Metrics = append(fp.Metrics, documentMetrics(records)...)
	fp.Curves = accumulation(records)
	return fp, nil
}

func metric(id, family string, v float64) Metric {
	return Metric{MetricID: id, Family: family, Value: v, Status: Comparable}
}
func missing(id, family, why string) Metric {
	return Metric{MetricID: id, Family: family, Status: NotComparable, Reason: why}
}

func glyphMetrics(rs []Record) []Metric {
	alphabet, initial, final, bigrams, trigrams := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}
	var totalSymbols int
	for _, r := range rs {
		for i, s := range r.Symbols {
			alphabet[s] = true
			totalSymbols++
			if i == 0 {
				initial[s] = true
			}
			if i == len(r.Symbols)-1 {
				final[s] = true
			}
			if i > 0 {
				bigrams[r.Symbols[i-1]+"\x1f"+s] = true
			}
			if i > 1 {
				trigrams[r.Symbols[i-2]+"\x1f"+r.Symbols[i-1]+"\x1f"+s] = true
			}
		}
	}
	a := float64(len(alphabet))
	bposs := a * a
	tposs := bposs * a
	return []Metric{
		metric("G01_ALPHABET_SIZE", "G", a),
		metric("G02_INITIAL_RESTRICTION_DENSITY", "G", safe(a-float64(len(initial)), a)),
		metric("G03_FINAL_RESTRICTION_DENSITY", "G", safe(a-float64(len(final)), a)),
		metric("G04_BIGRAM_OCCUPANCY", "G", safe(float64(len(bigrams)), bposs)),
		metric("G05_TRIGRAM_OCCUPANCY", "G", safe(float64(len(trigrams)), tposs)),
		metric("G06_SYMBOL_CONDITIONAL_ENTROPY_NORM", "G", conditionalEntropy(rs, 1)),
		metric("G07_HIGHER_ORDER_ENTROPY_REDUCTION", "G", conditionalEntropy(rs, 1)-conditionalEntropy(rs, 2)),
		metric("G08_MEAN_SYMBOLS_PER_TOKEN", "G", safe(float64(totalSymbols), float64(len(rs)))),
	}
}

func tokenMetrics(rs []Record) []Metric {
	freq := map[string]int{}
	seqs := map[string][]string{}
	lengths := make([]float64, 0, len(rs))
	prefixes, suffixes, internals, alphabet := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, r := range rs {
		key := strings.Join(r.Symbols, "\x1f")
		freq[key]++
		seqs[key] = r.Symbols
		lengths = append(lengths, float64(len(r.Symbols)))
		prefixes[r.Symbols[0]] = true
		suffixes[r.Symbols[len(r.Symbols)-1]] = true
		for i, s := range r.Symbols {
			alphabet[s] = true
			if i > 0 && i < len(r.Symbols)-1 {
				internals[s] = true
			}
		}
	}
	hapax := 0
	vocab := sortedStringKeys(freq) // deterministic order: downstream float sums (clustering, Spearman) must not depend on map iteration order
	for _, t := range vocab {
		if freq[t] == 1 {
			hapax++
		}
	}
	adjacency, edges := buildEditAdjacency(vocab, seqs)
	possible := float64(len(vocab) * (len(vocab) - 1) / 2)
	return []Metric{
		metric("T01_MEAN_TOKEN_LENGTH", "T", mean(lengths)), metric("T02_TOKEN_LENGTH_SD", "T", stddev(lengths)),
		metric("T03_UNIQUE_TOKEN_RATIO", "T", safe(float64(len(freq)), float64(len(rs)))), metric("T04_HAPAX_RATIO", "T", safe(float64(hapax), float64(len(freq)))),
		metric("T05_PREFIX_DIVERSITY", "T", safe(float64(len(prefixes)), float64(len(freq)))), metric("T06_SUFFIX_DIVERSITY", "T", safe(float64(len(suffixes)), float64(len(freq)))),
		metric("T07_EDIT_GRAPH_DENSITY", "T", safe(float64(edges), possible)), metric("T08_EDIT_GRAPH_GIANT_SHARE", "T", giantShare(len(vocab), adjacency)),
		metric("T09_EDIT_GRAPH_CLUSTERING", "T", clustering(adjacency)),
		metric("T10_EDIT_DEGREE_FREQUENCY_SPEARMAN", "T", degreeFrequencySpearman(vocab, freq, adjacency)),
		metric("T11_POSITIONAL_RESTRICTION_DENSITY", "T", safe(float64(3*len(alphabet)-len(prefixes)-len(suffixes)-len(internals)), float64(3*len(alphabet)))),
	}
}

// supportRegime is one frozen sequence-support definition shared by every
// support-stratified metric family (COMPARATIVE_EXPERIMENT_SPEC "Frozen
// supports and sizes"): frequency>=5, frequency>=10, top100, top250, and the
// matched frequent-token vocabulary N=553.
type supportRegime struct {
	name     string
	min, top int
}

var supportRegimes = []supportRegime{{"FREQ_GE_5", 5, 0}, {"FREQ_GE_10", 10, 0}, {"TOP_100", 0, 100}, {"TOP_250", 0, 250}, {"MATCHED_VOCAB", 0, 553}}

func sequenceMetrics(rs []Record) []Metric {
	var out []Metric
	freq := map[string]int{}
	for _, r := range rs {
		freq[r.Token]++
	}
	for _, reg := range supportRegimes {
		keep := selectVocabulary(freq, reg.min, reg.top)
		transitions := map[string]int{}
		total := 0
		for i := 1; i < len(rs); i++ {
			if sameSequenceUnit(rs[i-1], rs[i]) && keep[rs[i-1].Token] && keep[rs[i].Token] {
				transitions[rs[i-1].Token+"\x1f"+rs[i].Token]++
				total++
			}
		}
		v := float64(len(keep))
		m := Metric{MetricID: "S01_TRANSITION_DENSITY", Family: "S", Regime: reg.name, Value: safe(float64(len(transitions)), v*v), Status: Comparable}
		if len(keep) < 2 {
			m.Status = NotComparable
			m.Reason = "fewer than two eligible token types"
		}
		out = append(out, m)
		out = append(out, Metric{MetricID: "S02_TRANSITION_ZERO_DENSITY", Family: "S", Regime: reg.name, Value: 1 - m.Value, Status: m.Status, Reason: m.Reason})
		out = append(out, Metric{MetricID: "S03_TRANSITION_ENTROPY_NORM", Family: "S", Regime: reg.name, Value: entropyCounts(transitions, total, int(v*v)), Status: m.Status, Reason: m.Reason})
		preferred, depleted := transitionPreference(rs, keep, transitions, total)
		out = append(out, Metric{MetricID: "S06_PREFERRED_TRANSITION_FRACTION", Family: "S", Regime: reg.name, Value: preferred, Status: m.Status, Reason: m.Reason}, Metric{MetricID: "S07_DEPLETED_TRANSITION_FRACTION", Family: "S", Regime: reg.name, Value: depleted, Status: m.Status, Reason: m.Reason}, Metric{MetricID: "S08_HIGHER_ORDER_PREDICTIVE_GAIN", Family: "S", Regime: reg.name, Value: tokenConditionalEntropy(rs, keep, 1) - tokenConditionalEntropy(rs, keep, 2), Status: m.Status, Reason: m.Reason})
	}
	repeated2, repeated3 := repeatedTokenNgrams(rs, 2), repeatedTokenNgrams(rs, 3)
	out = append(out, metric("S04_REPEATED_BIGRAM_TYPES", "S", float64(repeated2)), metric("S05_REPEATED_TRIGRAM_TYPES", "S", float64(repeated3)))
	return out
}

func lineMetrics(rs []Record) []Metric {
	for _, r := range rs {
		if !r.PhysicalLine.Observed {
			return []Metric{missing("L01_LINE_TOKEN_COUNT_MEAN", "L", "physical lines not observed"), missing("L02_LINE_SYMBOL_COUNT_MEAN", "L", "physical lines not observed"), missing("L03_BOUNDARY_SPECIALIZATION", "L", "physical lines not observed"), missing("L04_POSITION_PROGRESSION", "L", "physical lines not observed"), missing("L05_LINE_ASYMMETRY", "L", "physical lines not observed"), missing("L06_SAME_LINE_COOCCURRENCE_DENSITY", "L", "physical lines not observed"), missing("L07_SAME_LINE_NONCOOCCURRENCE_DENSITY", "L", "physical lines not observed")}
		}
	}
	lineTokens, lineSymbols := map[string]int{}, map[string]int{}
	first, last := map[string]int{}, map[string]int{}
	freq := map[string]int{}
	for _, r := range rs {
		k := lineKey(r)
		lineTokens[k]++
		lineSymbols[k] += len(r.Symbols)
		if r.TokenIndex == 0 {
			first[r.Token]++
		}
		last[r.Token]++
		freq[r.Token]++
	}
	var tc, sc []float64
	for k := range lineTokens {
		tc = append(tc, float64(lineTokens[k]))
		sc = append(sc, float64(lineSymbols[k]))
	}
	out := []Metric{metric("L01_LINE_TOKEN_COUNT_MEAN", "L", mean(tc)), metric("L02_LINE_SYMBOL_COUNT_MEAN", "L", mean(sc)), metric("L03_BOUNDARY_SPECIALIZATION", "L", jaccardDistance(first, last)), metric("L04_POSITION_PROGRESSION", "L", positionProgression(rs)), metric("L05_LINE_ASYMMETRY", "L", safe(stddev(tc), mean(tc)))}
	// L06/L07 are stratified over the same frozen sequence-support regimes
	// as S01/S02 (VM_REFERENCE_CONTRACT anchor "same-line zero density" is
	// defined on the frequency>=10 support, not an arbitrary top-100 cutoff).
	for _, reg := range supportRegimes {
		keep := selectVocabulary(freq, reg.min, reg.top)
		co, status, reason := sameLineDensity(rs, keep)
		out = append(out, Metric{MetricID: "L06_SAME_LINE_COOCCURRENCE_DENSITY", Family: "L", Regime: reg.name, Value: co, Status: status, Reason: reason})
		nco := 1 - co
		if status != Comparable {
			nco = 0
		}
		out = append(out, Metric{MetricID: "L07_SAME_LINE_NONCOOCCURRENCE_DENSITY", Family: "L", Regime: reg.name, Value: nco, Status: status, Reason: reason})
	}
	return out
}

func documentMetrics(rs []Record) []Metric {
	levels := []struct {
		id  string
		get func(Record) ObservedLevel
	}{{"LINE", func(r Record) ObservedLevel { return r.PhysicalLine }}, {"SECTION", func(r Record) ObservedLevel { return r.Section }}, {"PAGE", func(r Record) ObservedLevel { return r.Page }}, {"LOCUS", func(r Record) ObservedLevel { return r.Locus }}}
	var out []Metric
	for _, lv := range levels {
		observed := true
		groups := map[string]map[string]bool{}
		for _, r := range rs {
			x := lv.get(r)
			if !x.Observed {
				observed = false
				break
			}
			if groups[x.Value] == nil {
				groups[x.Value] = map[string]bool{}
			}
			groups[x.Value][r.Token] = true
		}
		if !observed {
			out = append(out, missing("D_"+lv.id+"_COHERENCE", "D", strings.ToLower(lv.id)+" not observed"), missing("D_"+lv.id+"_EXCLUSIVITY", "D", strings.ToLower(lv.id)+" not observed"), missing("D_"+lv.id+"_VARIANCE_SHARE", "D", strings.ToLower(lv.id)+" not observed"), missing("D_"+lv.id+"_PROGRESSION", "D", strings.ToLower(lv.id)+" not observed"))
			continue
		}
		out = append(out, metric("D_"+lv.id+"_COHERENCE", "D", groupCoherence(groups)), metric("D_"+lv.id+"_EXCLUSIVITY", "D", groupExclusivity(groups)), metric("D_"+lv.id+"_VARIANCE_SHARE", "D", groupVarianceShare(rs, lv.get)), metric("D_"+lv.id+"_PROGRESSION", "D", groupProgression(rs, lv.get)))
	}
	return out
}

func accumulation(rs []Record) []CurvePoint {
	var out []CurvePoint
	for _, n := range RarefactionCheckpoints {
		if len(rs) < n {
			for _, id := range []string{"A2_BIGRAM_TYPES", "A3_TRIGRAM_TYPES", "AT_TRANSITION_TYPES"} {
				out = append(out, CurvePoint{CurveID: id, Checkpoint: n, Status: NotComparable, Reason: "corpus smaller than checkpoint"})
			}
			continue
		}
		a2, a3, at := AccumulationCounts(rs[:n])
		out = append(out, CurvePoint{"A2_BIGRAM_TYPES", n, float64(a2), Comparable, ""}, CurvePoint{"A3_TRIGRAM_TYPES", n, float64(a3), Comparable, ""}, CurvePoint{"AT_TRANSITION_TYPES", n, float64(at), Comparable, ""})
	}
	return out
}

// AccumulationCounts returns the observed symbol-bigram types, symbol-trigram
// types, and token-transition types over exactly the given record set with
// no internal slicing. It is shared by the raw prefix curve above and by the
// boundary-preserving rarefaction draws in rarefaction.go, so both the raw
// and the rarefied accumulation estimate use identical counting semantics.
func AccumulationCounts(rs []Record) (a2, a3, at int) {
	bg, tg, transitions := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for i, r := range rs {
		for j := 1; j < len(r.Symbols); j++ {
			bg[r.Symbols[j-1]+"\x1f"+r.Symbols[j]] = true
		}
		for j := 2; j < len(r.Symbols); j++ {
			tg[r.Symbols[j-2]+"\x1f"+r.Symbols[j-1]+"\x1f"+r.Symbols[j]] = true
		}
		if i > 0 && sameSequenceUnit(rs[i-1], r) {
			transitions[rs[i-1].Token+"\x1f"+r.Token] = true
		}
	}
	return len(bg), len(tg), len(transitions)
}

func safe(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}
func mean(x []float64) float64 {
	var s float64
	for _, v := range x {
		s += v
	}
	return safe(s, float64(len(x)))
}
func stddev(x []float64) float64 {
	m := mean(x)
	var s float64
	for _, v := range x {
		d := v - m
		s += d * d
	}
	return math.Sqrt(safe(s, float64(len(x))))
}
func conditionalEntropy(rs []Record, order int) float64 {
	contexts := map[string]map[string]int{}
	total := map[string]int{}
	alphabet := map[string]bool{}
	for _, r := range rs {
		for _, s := range r.Symbols {
			alphabet[s] = true
		}
		for i := order; i < len(r.Symbols); i++ {
			c := strings.Join(r.Symbols[i-order:i], "\x1f")
			if contexts[c] == nil {
				contexts[c] = map[string]int{}
			}
			contexts[c][r.Symbols[i]]++
			total[c]++
		}
	}
	var h float64
	var n int
	for _, c := range sortedStringKeys(total) {
		counts := contexts[c]
		for _, s := range sortedStringKeys(counts) {
			v := counts[s]
			p := float64(v) / float64(total[c])
			h -= float64(total[c]) * p * math.Log2(p)
		}
		n += total[c]
	}
	if len(alphabet) < 2 || n == 0 {
		return 0
	}
	return h / float64(n) / math.Log2(float64(len(alphabet)))
}
func selectVocabulary(freq map[string]int, min, top int) map[string]bool {
	type kv struct {
		k string
		v int
	}
	a := make([]kv, 0, len(freq))
	for k, v := range freq {
		if min == 0 || v >= min {
			a = append(a, kv{k, v})
		}
	}
	sort.Slice(a, func(i, j int) bool {
		if a[i].v != a[j].v {
			return a[i].v > a[j].v
		}
		return a[i].k < a[j].k
	})
	if top > 0 && len(a) > top {
		a = a[:top]
	}
	out := map[string]bool{}
	for _, x := range a {
		out[x.k] = true
	}
	return out
}
func sameSequenceUnit(a, b Record) bool {
	return a.Document == b.Document && a.Section == b.Section && a.Page == b.Page && a.Locus == b.Locus && a.PhysicalLine == b.PhysicalLine
}
func lineKey(r Record) string {
	return r.Document.Value + "\x1f" + r.Section.Value + "\x1f" + r.Page.Value + "\x1f" + r.Locus.Value + "\x1f" + r.PhysicalLine.Value
}
func repeatedTokenNgrams(rs []Record, n int) int {
	c := map[string]int{}
	for i := n - 1; i < len(rs); i++ {
		ok := true
		parts := make([]string, n)
		for j := 0; j < n; j++ {
			parts[j] = rs[i-n+1+j].Token
			if j > 0 && !sameSequenceUnit(rs[i-n+j], rs[i-n+1+j]) {
				ok = false
			}
		}
		if ok {
			c[strings.Join(parts, "\x1f")]++
		}
	}
	r := 0
	for _, v := range c {
		if v > 1 {
			r++
		}
	}
	return r
}
func entropyCounts(c map[string]int, total, possible int) float64 {
	if total == 0 || possible < 2 {
		return 0
	}
	var h float64
	for _, k := range sortedStringKeys(c) {
		p := float64(c[k]) / float64(total)
		h -= p * math.Log2(p)
	}
	return h / math.Log2(float64(possible))
}
func sequenceEditOne(ar, br []string) bool {
	if d := len(ar) - len(br); d > 1 || d < -1 {
		return false
	}
	i, j, diffs := 0, 0, 0
	for i < len(ar) && j < len(br) {
		if ar[i] == br[j] {
			i++
			j++
			continue
		}
		diffs++
		if diffs > 1 {
			return false
		}
		if len(ar) > len(br) {
			i++
		} else if len(br) > len(ar) {
			j++
		} else {
			i++
			j++
		}
	}
	diffs += (len(ar) - i) + (len(br) - j)
	return diffs == 1
}
func buildEditAdjacency(vocab []string, seqs map[string][]string) (map[int]map[int]bool, int) {
	adj := map[int]map[int]bool{}
	edges := map[[2]int]bool{}
	add := func(i, j int) {
		if i == j {
			return
		}
		if i > j {
			i, j = j, i
		}
		e := [2]int{i, j}
		if edges[e] {
			return
		}
		edges[e] = true
		if adj[i] == nil {
			adj[i] = map[int]bool{}
		}
		if adj[j] == nil {
			adj[j] = map[int]bool{}
		}
		adj[i][j] = true
		adj[j][i] = true
	}
	exact := map[string]int{}
	for i, k := range vocab {
		exact[strings.Join(seqs[k], "\x1f")] = i
	}
	subs := map[string][]int{}
	for i, k := range vocab {
		s := seqs[k]
		for p := range s {
			d := append([]string(nil), s[:p]...)
			d = append(d, s[p+1:]...)
			sig := fmt.Sprintf("%d:%d:%s", len(s), p, strings.Join(d, "\x1f"))
			subs[sig] = append(subs[sig], i)
			if j, ok := exact[strings.Join(d, "\x1f")]; ok {
				add(i, j)
			}
		}
	}
	for _, group := range subs {
		for i := 0; i < len(group); i++ {
			for j := i + 1; j < len(group); j++ {
				if sequenceEditOne(seqs[vocab[group[i]]], seqs[vocab[group[j]]]) {
					add(group[i], group[j])
				}
			}
		}
	}
	return adj, len(edges)
}
func giantShare(n int, a map[int]map[int]bool) float64 {
	seen := map[int]bool{}
	max := 0
	for i := 0; i < n; i++ {
		if seen[i] {
			continue
		}
		q := []int{i}
		seen[i] = true
		size := 0
		for len(q) > 0 {
			x := q[0]
			q = q[1:]
			size++
			for y := range a[x] {
				if !seen[y] {
					seen[y] = true
					q = append(q, y)
				}
			}
		}
		if size > max {
			max = size
		}
	}
	return safe(float64(max), float64(n))
}
func clustering(a map[int]map[int]bool) float64 {
	var sum float64
	var nodes int
	for _, i := range sortedIntKeys(a) {
		ns := a[i]
		k := len(ns)
		if k < 2 {
			continue
		}
		e := 0
		for x := range ns {
			for y := range ns {
				if x < y && a[x][y] {
					e++
				}
			}
		}
		sum += safe(float64(e), float64(k*(k-1)/2))
		nodes++
	}
	return safe(sum, float64(nodes))
}

// sortedStringKeys returns the keys of a string-keyed map in stable sorted
// order, so that summing float64 values in map iteration order (which Go
// randomizes per process) never affects a reproducible result.
func sortedStringKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// sortedIntKeys is the int-keyed analogue of sortedStringKeys.
func sortedIntKeys[V any](m map[int]V) []int {
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}
func degreeFrequencySpearman(vocab []string, freq map[string]int, a map[int]map[int]bool) float64 {
	if len(vocab) < 2 {
		return 0
	}
	x, y := make([]float64, len(vocab)), make([]float64, len(vocab))
	for i, k := range vocab {
		x[i] = float64(len(a[i]))
		y[i] = float64(freq[k])
	}
	return pearson(rank(x), rank(y))
}
func rank(x []float64) []float64 {
	type pair struct {
		i int
		v float64
	}
	p := make([]pair, len(x))
	for i, v := range x {
		p[i] = pair{i, v}
	}
	sort.Slice(p, func(i, j int) bool { return p[i].v < p[j].v })
	r := make([]float64, len(x))
	for i := 0; i < len(p); {
		j := i + 1
		for j < len(p) && p[j].v == p[i].v {
			j++
		}
		avg := float64(i+j-1)/2 + 1
		for k := i; k < j; k++ {
			r[p[k].i] = avg
		}
		i = j
	}
	return r
}
func pearson(x, y []float64) float64 {
	xm, ym := mean(x), mean(y)
	var num, xx, yy float64
	for i := range x {
		a, b := x[i]-xm, y[i]-ym
		num += a * b
		xx += a * a
		yy += b * b
	}
	return safe(num, math.Sqrt(xx*yy))
}
func transitionPreference(rs []Record, keep map[string]bool, observed map[string]int, total int) (float64, float64) {
	if total == 0 {
		return 0, 0
	}
	left, right := map[string]int{}, map[string]int{}
	for i := 1; i < len(rs); i++ {
		if sameSequenceUnit(rs[i-1], rs[i]) && keep[rs[i-1].Token] && keep[rs[i].Token] {
			left[rs[i-1].Token]++
			right[rs[i].Token]++
		}
	}
	pref, dep, eligible := 0, 0, 0
	for a := range keep {
		for b := range keep {
			expected := float64(left[a]*right[b]) / float64(total)
			if expected < 5 {
				continue
			}
			eligible++
			ratio := safe(float64(observed[a+"\x1f"+b]), expected)
			if ratio >= 2 {
				pref++
			}
			if ratio <= .5 {
				dep++
			}
		}
	}
	return safe(float64(pref), float64(eligible)), safe(float64(dep), float64(eligible))
}
func tokenConditionalEntropy(rs []Record, keep map[string]bool, order int) float64 {
	contexts := map[string]map[string]int{}
	totals := map[string]int{}
	eligible := 0
	for _, v := range keep {
		_ = v
		eligible++
	}
	if eligible < 2 {
		return 0
	}
	for i := order; i < len(rs); i++ {
		parts := make([]string, order)
		ok := keep[rs[i].Token]
		for j := 0; j < order; j++ {
			p := rs[i-order+j]
			parts[j] = p.Token
			ok = ok && keep[p.Token]
			if j > 0 {
				ok = ok && sameSequenceUnit(rs[i-order+j-1], p)
			}
		}
		ok = ok && sameSequenceUnit(rs[i-1], rs[i])
		if !ok {
			continue
		}
		c := strings.Join(parts, "\x1f")
		if contexts[c] == nil {
			contexts[c] = map[string]int{}
		}
		contexts[c][rs[i].Token]++
		totals[c]++
	}
	var h float64
	var n int
	for _, c := range sortedStringKeys(totals) {
		counts := contexts[c]
		for _, s := range sortedStringKeys(counts) {
			p := float64(counts[s]) / float64(totals[c])
			h -= float64(totals[c]) * p * math.Log2(p)
		}
		n += totals[c]
	}
	return safe(h, float64(n)*math.Log2(float64(eligible)))
}
func jaccardDistance(a, b map[string]int) float64 {
	inter, union := 0, 0
	seen := map[string]bool{}
	for k := range a {
		seen[k] = true
	}
	for k := range b {
		seen[k] = true
	}
	for k := range seen {
		union++
		if a[k] > 0 && b[k] > 0 {
			inter++
		}
	}
	return 1 - safe(float64(inter), float64(union))
}
func positionProgression(rs []Record) float64 {
	var pos, lens []float64
	counts := map[string]int{}
	for _, r := range rs {
		counts[lineKey(r)]++
	}
	for _, r := range rs {
		c := counts[lineKey(r)]
		if c > 1 {
			pos = append(pos, float64(r.TokenIndex)/float64(c-1))
			lens = append(lens, float64(len(r.Symbols)))
		}
	}
	return pearson(pos, lens)
}
func sameLineDensity(rs []Record, keep map[string]bool) (float64, Status, string) {
	lines := map[string]map[string]bool{}
	for _, r := range rs {
		k := lineKey(r)
		if lines[k] == nil {
			lines[k] = map[string]bool{}
		}
		lines[k][r.Token] = true
	}
	if len(keep) < 2 {
		return 0, NotComparable, "fewer than two eligible token types"
	}
	pairs := map[string]bool{}
	for _, set := range lines {
		var a []string
		for t := range set {
			if keep[t] {
				a = append(a, t)
			}
		}
		sort.Strings(a)
		for i := 0; i < len(a); i++ {
			for j := i + 1; j < len(a); j++ {
				pairs[a[i]+"\x1f"+a[j]] = true
			}
		}
	}
	return safe(float64(len(pairs)), float64(len(keep)*(len(keep)-1)/2)), Comparable, ""
}
func groupCoherence(g map[string]map[string]bool) float64 {
	if len(g) < 2 {
		return 1
	}
	keys := sortedStringKeys(g)
	var s float64
	var n int
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			s += 1 - setJaccardDistance(g[keys[i]], g[keys[j]])
			n++
		}
	}
	return safe(s, float64(n))
}
func groupExclusivity(g map[string]map[string]bool) float64 {
	owners := map[string]int{}
	for _, set := range g {
		for t := range set {
			owners[t]++
		}
	}
	ex, total := 0, 0
	for _, n := range owners {
		total++
		if n == 1 {
			ex++
		}
	}
	return safe(float64(ex), float64(total))
}
func groupProgression(rs []Record, get func(Record) ObservedLevel) float64 {
	type acc struct{ symbols, tokens int }
	groups := map[string]*acc{}
	var order []string
	for _, r := range rs {
		k := r.Document.Value + "\x1f" + get(r).Value
		if groups[k] == nil {
			groups[k] = &acc{}
			order = append(order, k)
		}
		groups[k].symbols += len(r.Symbols)
		groups[k].tokens++
	}
	if len(order) < 2 {
		return 0
	}
	xm := float64(len(order)-1) / 2
	ys := make([]float64, len(order))
	var ym float64
	for i, k := range order {
		ys[i] = safe(float64(groups[k].symbols), float64(groups[k].tokens))
		ym += ys[i]
	}
	ym /= float64(len(ys))
	var cov, vx float64
	for i, y := range ys {
		x := float64(i)
		cov += (x - xm) * (y - ym)
		vx += (x - xm) * (x - xm)
	}
	return safe(cov, vx*math.Max(ym, 1))
}
func groupVarianceShare(rs []Record, get func(Record) ObservedLevel) float64 {
	overall := make([]float64, len(rs))
	groups := map[string][]float64{}
	for i, r := range rs {
		v := float64(len(r.Symbols))
		overall[i] = v
		k := r.Document.Value + "\x1f" + get(r).Value
		groups[k] = append(groups[k], v)
	}
	total := stddev(overall)
	if total == 0 {
		return 0
	}
	var between float64
	om := mean(overall)
	for _, k := range sortedStringKeys(groups) {
		x := groups[k]
		d := mean(x) - om
		between += float64(len(x)) * d * d
	}
	return safe(between, float64(len(rs))*total*total)
}
func setJaccardDistance(a, b map[string]bool) float64 {
	inter, union := 0, len(a)
	for k := range b {
		if a[k] {
			inter++
		} else {
			union++
		}
	}
	return 1 - safe(float64(inter), float64(union))
}

func ExplainMetric(id string) (string, error) {
	for _, m := range MetricRegistry() {
		if m.ID == id {
			return m.Definition, nil
		}
	}
	return "", fmt.Errorf("unknown metric %q", id)
}

type MetricDefinition struct{ ID, Family, Definition string }

func MetricRegistry() []MetricDefinition {
	return []MetricDefinition{
		{"G01_ALPHABET_SIZE", "G", "number of distinct observed symbols"}, {"G02_INITIAL_RESTRICTION_DENSITY", "G", "share of alphabet never token-initial"}, {"G03_FINAL_RESTRICTION_DENSITY", "G", "share of alphabet never token-final"}, {"G04_BIGRAM_OCCUPANCY", "G", "observed symbol bigrams divided by |A|^2"}, {"G05_TRIGRAM_OCCUPANCY", "G", "observed symbol trigrams divided by |A|^3"}, {"G06_SYMBOL_CONDITIONAL_ENTROPY_NORM", "G", "H(symbol_i|symbol_i-1)/log2(|A|)"}, {"G07_HIGHER_ORDER_ENTROPY_REDUCTION", "G", "normalized H1 minus normalized H2"}, {"G08_MEAN_SYMBOLS_PER_TOKEN", "G", "mean observed symbols per token"},
		{"T01_MEAN_TOKEN_LENGTH", "T", "mean symbols per token"}, {"T02_TOKEN_LENGTH_SD", "T", "population SD of token symbol counts"}, {"T03_UNIQUE_TOKEN_RATIO", "T", "distinct tokens divided by tokens"}, {"T04_HAPAX_RATIO", "T", "once-observed types divided by types"}, {"T05_PREFIX_DIVERSITY", "T", "distinct initial symbols divided by types"}, {"T06_SUFFIX_DIVERSITY", "T", "distinct final symbols divided by types"}, {"T07_EDIT_GRAPH_DENSITY", "T", "symbol-sequence edit-one edges divided by possible undirected edges"}, {"T08_EDIT_GRAPH_GIANT_SHARE", "T", "largest edit component divided by types"}, {"T09_EDIT_GRAPH_CLUSTERING", "T", "mean local clustering among degree>=2 nodes"}, {"T10_EDIT_DEGREE_FREQUENCY_SPEARMAN", "T", "Spearman correlation of type frequency and edit degree"}, {"T11_POSITIONAL_RESTRICTION_DENSITY", "T", "mean alphabet restriction at initial, internal, and final positions"},
		{"S01_TRANSITION_DENSITY", "S", "observed eligible token transitions divided by V^2"}, {"S02_TRANSITION_ZERO_DENSITY", "S", "one minus transition density"}, {"S03_TRANSITION_ENTROPY_NORM", "S", "transition entropy divided by log2(V^2)"}, {"S04_REPEATED_BIGRAM_TYPES", "S", "token bigram types occurring at least twice"}, {"S05_REPEATED_TRIGRAM_TYPES", "S", "token trigram types occurring at least twice"}, {"S06_PREFERRED_TRANSITION_FRACTION", "S", "eligible pairs with observed/expected >=2"}, {"S07_DEPLETED_TRANSITION_FRACTION", "S", "eligible pairs with observed/expected <=0.5"}, {"S08_HIGHER_ORDER_PREDICTIVE_GAIN", "S", "normalized first-order minus second-order conditional token entropy"},
		{"L01_LINE_TOKEN_COUNT_MEAN", "L", "mean observed tokens per physical line"}, {"L02_LINE_SYMBOL_COUNT_MEAN", "L", "mean observed symbols per physical line"}, {"L03_BOUNDARY_SPECIALIZATION", "L", "Jaccard distance of line-initial and line-final token-type sets"}, {"L04_POSITION_PROGRESSION", "L", "correlation of normalized line position and token length"}, {"L05_LINE_ASYMMETRY", "L", "coefficient of variation of line token counts"}, {"L06_SAME_LINE_COOCCURRENCE_DENSITY", "L", "observed undirected same-line pairs among the frozen support regime, stratified by support_regime"}, {"L07_SAME_LINE_NONCOOCCURRENCE_DENSITY", "L", "one minus the same-line co-occurrence density for the same support regime"},
		{"D_LINE_COHERENCE", "D", "mean pairwise vocabulary coherence among lines"}, {"D_LINE_EXCLUSIVITY", "D", "share of types exclusive to one line"}, {"D_LINE_VARIANCE_SHARE", "D", "between-line token-length variance share"}, {"D_LINE_PROGRESSION", "D", "normalized token-length trend across line order"},
		{"D_SECTION_COHERENCE", "D", "mean pairwise vocabulary coherence among sections"}, {"D_SECTION_EXCLUSIVITY", "D", "share of types exclusive to one section"}, {"D_SECTION_VARIANCE_SHARE", "D", "between-section token-length variance share"}, {"D_SECTION_PROGRESSION", "D", "normalized token-length trend across section order"},
		{"D_PAGE_COHERENCE", "D", "mean pairwise vocabulary coherence among pages"}, {"D_PAGE_EXCLUSIVITY", "D", "share of types exclusive to one page"}, {"D_PAGE_VARIANCE_SHARE", "D", "between-page token-length variance share"}, {"D_PAGE_PROGRESSION", "D", "normalized token-length trend across page order"},
		{"D_LOCUS_COHERENCE", "D", "mean pairwise vocabulary coherence among loci"}, {"D_LOCUS_EXCLUSIVITY", "D", "share of types exclusive to one locus"}, {"D_LOCUS_VARIANCE_SHARE", "D", "between-locus token-length variance share"}, {"D_LOCUS_PROGRESSION", "D", "normalized token-length trend across locus order"},
	}
}
