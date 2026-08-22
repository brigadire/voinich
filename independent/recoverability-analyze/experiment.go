package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"

	"zcore.dev/voinich/internal/mechanismspace"
	"zcore.dev/voinich/internal/tokenrepetition"
)

// Final measurements use the first 128 words of the block-wise TEST third.
// This is one of Task67's preregistered message sizes and keeps the complete
// 100-replicate corruption grid tractable without inspecting TEST content.
const testWords = 128

type encoded struct {
	plain  [][]string
	cipher [][]string
}

type decoder struct {
	candidate candidate
	counts    map[string]map[string]int
	best      map[string][]string
	forms     [][]string
	inverse   map[string]string
	nearest   map[string][]string
}

type recovery struct {
	glyph, token, seqEdit, normChar, wer, exact, entropy float64
}

type propagation struct {
	damaged, correctAfter int
	recoveryAfter         float64
	lSync                 int
	catastrophic          bool
}

func encodeAligned(c candidate, corpus mechanismspace.Corpus) encoded {
	o := mechanismspace.Transform(c.Config, corpus)
	plain := corpus.Glyphs()
	if c.Config.InputMode == mechanismspace.Stream {
		var glyphs []string
		for _, w := range plain {
			glyphs = append(glyphs, w...)
		}
		plain = make([][]string, 0, len(o.Tokens))
		pos := 0
		for i := range o.Tokens {
			length := c.Config.GroupLen
			if c.Config.Grouping == mechanismspace.StateGrouping {
				length = 1 + i%6
			}
			if length < 1 {
				length = 4
			}
			end := pos + length
			if end > len(glyphs) {
				end = len(glyphs)
			}
			plain = append(plain, cloneToken(glyphs[pos:end]))
			pos = end
		}
	}
	return encoded{plain: plain, cipher: cloneTokens(o.Tokens)}
}

func trainDecoder(c candidate, train, validation mechanismspace.Corpus) *decoder {
	d := &decoder{candidate: c, counts: map[string]map[string]int{}, best: map[string][]string{}, inverse: map[string]string{}, nearest: map[string][]string{}}
	for _, corpus := range []mechanismspace.Corpus{train, validation} {
		e := encodeAligned(c, corpus)
		for i := 0; i < min(len(e.plain), len(e.cipher)); i++ {
			ck, pk := tokenKey(e.cipher[i]), tokenKey(e.plain[i])
			if d.counts[ck] == nil {
				d.counts[ck] = map[string]int{}
			}
			d.counts[ck][pk]++
		}
	}
	keys := make([]string, 0, len(d.counts))
	for ck, pcs := range d.counts {
		keys = append(keys, ck)
		bestKey, bestN := "", -1
		for pk, n := range pcs {
			if n > bestN || n == bestN && pk < bestKey {
				bestKey, bestN = pk, n
			}
		}
		d.best[ck] = keyToken(bestKey)
	}
	sort.Strings(keys)
	for _, k := range keys {
		d.forms = append(d.forms, keyToken(k))
	}
	return d
}

func (d *decoder) decode(tokens [][]string) [][]string {
	out := make([][]string, len(tokens))
	if d.candidate.Name == "M1_MONOALPHABETIC" && len(d.inverse) == 0 {
		d.setCipherKey(tokens)
	}
	for i, tok := range tokens {
		switch d.candidate.Name {
		case "M0_IDENTITY":
			out[i] = cloneToken(tok)
		case "M1_MONOALPHABETIC":
			out[i] = make([]string, len(tok))
			for j, g := range tok {
				if p, ok := d.inverse[g]; ok {
					out[i][j] = p
				} else {
					out[i][j] = "?"
				}
			}
		case "M2_HOMOPHONY_H2":
			out[i] = make([]string, len(tok))
			for j, g := range tok {
				if k := strings.LastIndexByte(g, '_'); k > 0 {
					out[i][j] = g[:k]
				} else {
					out[i][j] = "?"
				}
			}
		default:
			key := tokenKey(tok)
			if p, ok := d.best[key]; ok {
				out[i] = cloneToken(p)
				continue
			}
			nearest := d.nearestForm(tok)
			if p, ok := d.best[tokenKey(nearest)]; ok {
				out[i] = cloneToken(p)
			} else {
				out[i] = []string{"?"}
			}
		}
	}
	return out
}

func (d *decoder) setCipherKey(tokens [][]string) {
	if d.candidate.Name == "M1_MONOALPHABETIC" {
		d.inverse = monoalphabeticInverse(cipherAlphabet(tokens), d.candidate.Config.Seed)
	}
}

func (d *decoder) nearestForm(tok []string) []string {
	key := tokenKey(tok)
	if cached, ok := d.nearest[key]; ok {
		return cached
	}
	bestD := math.MaxInt
	var best []string
	for _, f := range d.forms {
		dist := tokenrepetition.LevenshteinGlyphs(tok, f)
		if dist < bestD || dist == bestD && tokenKey(f) < tokenKey(best) {
			bestD, best = dist, f
		}
	}
	d.nearest[key] = best
	return best
}

func measureRecovery(want, got [][]string) recovery {
	wf, gf := flatten(want), flatten(got)
	charD := tokenrepetition.LevenshteinGlyphs(wf, gf)
	wk, gk := unitKeys(want), unitKeys(got)
	seqD := levenshteinStrings(wk, gk)
	correct := 0
	for i := 0; i < min(len(wk), len(gk)); i++ {
		if wk[i] == gk[i] {
			correct++
		}
	}
	denGlyph, denToken := max(1, max(len(wf), len(gf))), max(1, max(len(wk), len(gk)))
	return recovery{
		glyph:    1 - float64(charD)/float64(denGlyph),
		token:    float64(correct) / float64(denToken),
		seqEdit:  float64(seqD),
		normChar: float64(charD) / float64(denGlyph),
		wer:      float64(seqD) / float64(denToken),
		exact:    boolFloat(charD == 0 && seqD == 0),
		entropy:  entropyStrings(gf),
	}
}

func pluginInformation(e encoded) (hp, hc, mi, ratio float64) {
	n := min(len(e.plain), len(e.cipher))
	if n == 0 {
		return
	}
	pc, cc, joint := map[string]int{}, map[string]int{}, map[string]int{}
	for i := 0; i < n; i++ {
		p, c := tokenKey(e.plain[i]), tokenKey(e.cipher[i])
		pc[p]++
		cc[c]++
		joint[p+"\x00"+c]++
	}
	hp = entropyCounts(pc)
	hc = entropyCounts(joint) - entropyCounts(cc)
	if hc < 0 && hc > -1e-12 {
		hc = 0
	}
	mi = hp - hc
	if hp > 0 {
		ratio = mi / hp
	}
	return
}

func corruptOne(in [][]string, channel string, requested int, seed int64) ([][]string, int, string) {
	out := cloneTokens(in)
	if len(out) == 0 {
		return out, 0, "EMPTY"
	}
	r := rand.New(rand.NewSource(seed))
	pos := requested % len(out)
	alphabet := cipherAlphabet(out)
	alt := "§"
	if len(alphabet) > 0 {
		alt = alphabet[r.Intn(len(alphabet))]
	}
	switch channel {
	case "GLYPH_SUBSTITUTION":
		pos = seekNonEmpty(out, pos)
		j := len(out[pos]) / 2
		old := out[pos][j]
		for _, g := range alphabet {
			if g != old {
				alt = g
				break
			}
		}
		out[pos][j] = alt
		return out, pos, glyphLocation(j, len(out[pos]))
	case "GLYPH_DELETION":
		pos = seekNonEmpty(out, pos)
		j := len(out[pos]) / 2
		loc := glyphLocation(j, len(out[pos]))
		out[pos] = append(cloneToken(out[pos][:j]), out[pos][j+1:]...)
		return out, pos, loc
	case "GLYPH_INSERTION":
		pos = seekNonEmpty(out, pos)
		j := len(out[pos]) / 2
		out[pos] = insertGlyph(out[pos], j, alt)
		return out, pos, glyphLocation(j, len(out[pos]))
	case "TOKEN_BOUNDARY_INSERTION", "TOKEN_SPLIT":
		pos = seekSplittable(out, pos)
		j := len(out[pos]) / 2
		left, right := cloneToken(out[pos][:j]), cloneToken(out[pos][j:])
		out = append(out[:pos], append([][]string{left, right}, out[pos+1:]...)...)
		return out, pos, "TOKEN_MEDIAL"
	case "TOKEN_BOUNDARY_DELETION", "TOKEN_MERGE":
		if pos == len(out)-1 {
			pos--
		}
		if pos < 0 {
			pos = 0
		}
		merged := append(cloneToken(out[pos]), out[pos+1]...)
		out = append(out[:pos], append([][]string{merged}, out[pos+2:]...)...)
		return out, pos, "BOUNDARY_ADJACENT"
	}
	return out, pos, "UNKNOWN"
}

func corruptRate(in [][]string, channel string, rate float64, seed int64) ([][]string, int) {
	out := cloneTokens(in)
	if rate == 0 || len(out) == 0 {
		return out, 0
	}
	base := len(out)
	if strings.HasPrefix(channel, "GLYPH_") {
		base = len(flatten(out))
	}
	n := max(1, int(math.Round(rate*float64(base))))
	r := rand.New(rand.NewSource(seed))
	for i := 0; i < n; i++ {
		var pos int
		if len(out) > 0 {
			pos = r.Intn(len(out))
		}
		out, _, _ = corruptOne(out, channel, pos, seed+int64(i)*7919)
	}
	return out, n
}

func propagationMetrics(want, got [][]string, errorPos int) propagation {
	w, g := unitKeys(want), unitKeys(got)
	p := propagation{}
	start := min(errorPos, len(w))
	for i := start; i < len(w); i++ {
		if i < len(g) && w[i] == g[i] {
			p.correctAfter++
		} else {
			p.damaged++
		}
	}
	if len(w)-start > 0 {
		p.recoveryAfter = float64(p.correctAfter) / float64(len(w)-start)
	}
	p.lSync = -1
	for i := start + 1; i+2 < min(len(w), len(g)); i++ {
		if w[i] == g[i] && w[i+1] == g[i+1] && w[i+2] == g[i+2] {
			p.lSync = i - errorPos
			break
		}
	}
	if len(w)-start < 4 && p.lSync < 0 {
		p.lSync = -2 // right-censored by block end: three-unit run impossible
	}
	p.catastrophic = p.lSync == -1
	return p
}

func monoalphabeticInverse(alphabet []string, seed int64) map[string]string {
	plain := append([]string(nil), alphabet...)
	sort.Strings(plain)
	perm := append([]string(nil), plain...)
	r := rand.New(rand.NewSource(seed))
	r.Shuffle(len(perm), func(i, j int) { perm[i], perm[j] = perm[j], perm[i] })
	inv := make(map[string]string, len(plain))
	for i := range plain {
		inv[perm[i]] = plain[i]
	}
	return inv
}

func applyConflation(in [][]string, fraction float64, seed int64) ([][]string, int) {
	out := cloneTokens(in)
	a := cipherAlphabet(out)
	r := rand.New(rand.NewSource(seed))
	r.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i] })
	n := min(len(a)/2, max(1, int(math.Round(fraction*float64(len(a))))))
	m := map[string]string{}
	for i := 0; i < n; i++ {
		m[a[2*i+1]] = a[2*i]
	}
	for i := range out {
		for j, g := range out[i] {
			if x, ok := m[g]; ok {
				out[i][j] = x
			}
		}
	}
	return out, n
}

func applySplitting(in [][]string, fraction float64, seed int64) ([][]string, int) {
	out := cloneTokens(in)
	a := cipherAlphabet(out)
	r := rand.New(rand.NewSource(seed))
	r.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i] })
	n := min(len(a), max(1, int(math.Round(fraction*float64(len(a))))))
	selected := map[string]bool{}
	for _, g := range a[:n] {
		selected[g] = true
	}
	for i := range out {
		for j, g := range out[i] {
			if selected[g] {
				out[i][j] = fmt.Sprintf("%s#%d", g, r.Intn(2)+1)
			}
		}
	}
	return out, n
}

func removeSplittingMarks(in [][]string) [][]string {
	out := cloneTokens(in)
	for i := range out {
		for j, g := range out[i] {
			if k := strings.LastIndexByte(g, '#'); k > 0 {
				out[i][j] = g[:k]
			}
		}
	}
	return out
}

func reconstructBoundaries(flat []string, d *decoder) [][]string {
	if len(flat) == 0 {
		return nil
	}
	maxLen := 1
	for _, f := range d.forms {
		maxLen = max(maxLen, len(f))
	}
	if d.candidate.Name == "M0_IDENTITY" || d.candidate.Name == "M1_MONOALPHABETIC" || d.candidate.Name == "M2_HOMOPHONY_H2" {
		maxLen = 16
	}
	cost := make([]float64, len(flat)+1)
	prev := make([]int, len(flat)+1)
	for i := 1; i <= len(flat); i++ {
		cost[i] = math.Inf(1)
	}
	for end := 1; end <= len(flat); end++ {
		for l := 1; l <= maxLen && l <= end; l++ {
			start := end - l
			seg := flat[start:end]
			dist := 1
			if _, exact := d.best[tokenKey(seg)]; exact {
				dist = 0
			} else if len(d.forms) > 0 && d.candidate.Config.Grammar != mechanismspace.NoGrammar {
				dist = tokenrepetition.LevenshteinGlyphs(seg, d.nearestForm(seg))
			}
			candidateCost := cost[start] + float64(dist) + .02
			if candidateCost < cost[end] {
				cost[end], prev[end] = candidateCost, start
			}
		}
	}
	var reversed [][]string
	for end := len(flat); end > 0; {
		start := prev[end]
		reversed = append(reversed, cloneToken(flat[start:end]))
		end = start
	}
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	return reversed
}

func boundaryScores(want, got [][]string) (precision, recall, f1 float64) {
	w, g := boundarySet(want), boundarySet(got)
	tp := 0
	for p := range g {
		if w[p] {
			tp++
		}
	}
	if len(g) > 0 {
		precision = float64(tp) / float64(len(g))
	}
	if len(w) > 0 {
		recall = float64(tp) / float64(len(w))
	}
	if precision+recall > 0 {
		f1 = 2 * precision * recall / (precision + recall)
	}
	return
}

func resetDecode(d *decoder, clean, damaged [][]string, errorPos, interval int) [][]string {
	if interval <= 0 {
		return d.decode(damaged)
	}
	var out [][]string
	for start := 0; start < len(clean); start += interval {
		end := min(len(clean), start+interval)
		segment := cloneTokens(clean[start:end])
		if errorPos >= start && errorPos < end {
			local := errorPos - start
			// The same deleted boundary merges adjacent tokens unless the
			// boundary is protected by an explicit surviving reset marker.
			if local+1 < len(segment) {
				merged := append(cloneToken(segment[local]), segment[local+1]...)
				segment = append(segment[:local], append([][]string{merged}, segment[local+2:]...)...)
			}
		}
		decoded := d.decode(segment)
		for len(decoded) < end-start {
			decoded = append(decoded, []string{"<ERASURE>"})
		}
		out = append(out, decoded[:end-start]...)
	}
	return out
}

func limitCorpus(c mechanismspace.Corpus, n int) mechanismspace.Corpus {
	if len(c.Words) <= n {
		return c
	}
	return mechanismspace.Corpus{Name: c.Name, Words: cloneToken(c.Words[:n]), Lines: append([]int(nil), c.Lines[:n]...)}
}

func task66Compatibility(path string) (map[string]map[string]map[string]float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	if !s.Scan() {
		return nil, fmt.Errorf("empty %s", path)
	}
	header := strings.Split(s.Text(), "\t")
	index := map[string]int{}
	for i, name := range header {
		index[name] = i
	}
	mi, mok := index["mechanism"]
	ci, cok := index["corpus"]
	fi, fok := index["family"]
	pi, pok := index["progress"]
	if !mok || !cok || !fok || !pok {
		return nil, fmt.Errorf("%s lacks mechanism/corpus/family/progress columns", path)
	}
	out := map[string]map[string]map[string]float64{}
	for s.Scan() {
		p := strings.Split(s.Text(), "\t")
		if len(p) <= max(mi, ci, fi, pi) {
			continue
		}
		v, err := strconv.ParseFloat(p[pi], 64)
		if err != nil {
			return nil, err
		}
		mechanism, corpus, family := p[mi], p[ci], p[fi]
		if out[mechanism] == nil {
			out[mechanism] = map[string]map[string]float64{}
		}
		if out[mechanism][corpus] == nil {
			out[mechanism][corpus] = map[string]float64{}
		}
		out[mechanism][corpus][family] = v
	}
	return out, s.Err()
}

func candidateTask66Name(name string) string {
	if name == "G_FORM_MEDIUM" {
		return "M3_FORM_MEDIUM"
	}
	return name
}

func tokenKey(t []string) string { return strings.Join(t, "\x1f") }
func keyToken(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "\x1f")
}
func cloneToken(t []string) []string { return append([]string(nil), t...) }
func cloneTokens(t [][]string) [][]string {
	o := make([][]string, len(t))
	for i := range t {
		o[i] = cloneToken(t[i])
	}
	return o
}
func flatten(t [][]string) []string {
	var o []string
	for _, x := range t {
		o = append(o, x...)
	}
	return o
}
func unitKeys(t [][]string) []string {
	o := make([]string, len(t))
	for i := range t {
		o[i] = tokenKey(t[i])
	}
	return o
}
func cipherAlphabet(t [][]string) []string {
	m := map[string]bool{}
	for _, x := range t {
		for _, g := range x {
			m[g] = true
		}
	}
	o := make([]string, 0, len(m))
	for g := range m {
		o = append(o, g)
	}
	sort.Strings(o)
	return o
}
func seekNonEmpty(t [][]string, p int) int {
	for i := 0; i < len(t); i++ {
		j := (p + i) % len(t)
		if len(t[j]) > 0 {
			return j
		}
	}
	return 0
}
func seekSplittable(t [][]string, p int) int {
	for i := 0; i < len(t); i++ {
		j := (p + i) % len(t)
		if len(t[j]) >= 2 {
			return j
		}
	}
	return 0
}
func insertGlyph(t []string, p int, g string) []string {
	o := make([]string, 0, len(t)+1)
	o = append(o, t[:p]...)
	o = append(o, g)
	o = append(o, t[p:]...)
	return o
}
func glyphLocation(j, n int) string {
	if j == 0 {
		return "TOKEN_INITIAL"
	}
	if j >= n-1 {
		return "TOKEN_FINAL"
	}
	return "TOKEN_MEDIAL"
}
func boundarySet(t [][]string) map[int]bool {
	o := map[int]bool{}
	p := 0
	for i, x := range t {
		p += len(x)
		if i+1 < len(t) {
			o[p] = true
		}
	}
	return o
}
func entropyCounts(c map[string]int) float64 {
	n := 0
	for _, x := range c {
		n += x
	}
	if n == 0 {
		return 0
	}
	h := 0.0
	for _, x := range c {
		p := float64(x) / float64(n)
		h -= p * math.Log2(p)
	}
	return h
}
func entropyStrings(x []string) float64 {
	c := map[string]int{}
	for _, s := range x {
		c[s]++
	}
	return entropyCounts(c)
}
func levenshteinStrings(a, b []string) int {
	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur := make([]int, len(b)+1)
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(b)]
}
func boolFloat(v bool) float64 {
	if v {
		return 1
	}
	return 0
}
func boolText(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
