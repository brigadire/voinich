package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math"
	"math/rand"
	"os"
	"sort"
	"strings"
	"unicode/utf8"

	"zcore.dev/voinich/internal/evaglyph"
)

const (
	defaultCap      = 2000
	defaultShuffles = 100
	defaultSeed     = int64(20260816)
)

type Corpus struct {
	Lines     [][]string
	SHA256    string
	Bytes     int64
	GlyphMode string
}
type Metric struct {
	Name                                                                        string
	Tokens, Lines, Pairs, Types                                                 int
	GlyphStatus                                                                 string
	TokenObserved, TokenShuffleMean, TokenShuffleSD, TokenCorrected, TokenShare float64
	EdgeObserved, EdgeShuffleMean, EdgeShuffleSD, EdgeCorrected, EdgeShare      float64
}

func loadCorpus(path string, mode string) (Corpus, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Corpus{}, err
	}
	s := sha256.Sum256(b)
	c := Corpus{SHA256: hex.EncodeToString(s[:]), Bytes: int64(len(b)), GlyphMode: mode}
	sc := bufio.NewScanner(strings.NewReader(string(b)))
	sc.Buffer(make([]byte, 1024), 16<<20)
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) > 0 {
			c.Lines = append(c.Lines, f)
		}
	}
	if err := sc.Err(); err != nil {
		return c, err
	}
	return c, nil
}

// glyphs delegates Voynich-mode tokenization to the shared evaglyph
// parser (also used by Task59/60) so all three interpret EVA composites
// identically; this is a pure refactor - evaglyph.CollapseEVA reproduces
// the exact lowercase+longest-first-composite-collapse+rune-split
// sequence this function used to do inline (verified byte-identical
// against the committed Task58 experiment artifacts).
func glyphs(token, mode string) []string {
	if mode == "voynich" {
		return evaglyph.CollapseEVA(token)
	}
	if mode == "opaque" {
		return nil
	}
	if !utf8.ValidString(token) {
		return nil
	}
	r := []rune(token)
	out := make([]string, len(r))
	for i, x := range r {
		out[i] = string(x)
	}
	return out
}
func feature(token, mode, which string) string {
	g := glyphs(token, mode)
	if len(g) == 0 {
		return ""
	}
	switch which {
	case "first":
		return g[0]
	case "last":
		return g[len(g)-1]
	case "first2":
		if len(g) > 1 {
			return g[0] + g[1]
		}
		return g[0]
	case "last2":
		if len(g) > 1 {
			return g[len(g)-2] + g[len(g)-1]
		}
		return g[len(g)-1]
	}
	return token
}
// entropy sums in sorted-key order so the float64 accumulation - and
// therefore the result's low-order bits - does not depend on Go's
// randomized map iteration order (a pre-existing non-determinism this
// fixes; see project memory on sorting map keys before float
// accumulation). This changes only the last few significant digits of
// every published Task58 MI value, never their qualitative conclusions;
// experiments/rozanova-temerev-v1 was regenerated after this fix.
func entropy(c map[string]int) float64 {
	n := 0
	for _, v := range c {
		n += v
	}
	if n == 0 {
		return 0
	}
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := 0.0
	for _, k := range keys {
		v := c[k]
		p := float64(v) / float64(n)
		h -= p * math.Log2(p)
	}
	return h
}
func mi(x, y []string) float64 {
	if len(x) == 0 {
		return 0
	}
	cx, cy, j := map[string]int{}, map[string]int{}, map[string]int{}
	for i := range x {
		cx[x[i]]++
		cy[y[i]]++
		j[x[i]+"\x00"+y[i]]++
	}
	return entropy(cx) + entropy(cy) - entropy(j)
}
func sd(v []float64) float64 {
	if len(v) < 2 {
		return 0
	}
	m := 0.0
	for _, x := range v {
		m += x
	}
	m /= float64(len(v))
	z := 0.0
	for _, x := range v {
		z += (x - m) * (x - m)
	}
	return math.Sqrt(z / float64(len(v)-1))
}

func analyse(c Corpus, cap, shuffles int, seed int64) (Metric, error) {
	if cap < 1 || shuffles < 1 {
		return Metric{}, errors.New("cap and shuffles must be positive")
	}
	lines := c.Lines
	flat := []string{}
	lineIDs := []int{}
	for li, l := range lines {
		for _, t := range l {
			flat = append(flat, t)
			lineIDs = append(lineIDs, li)
		}
	}
	if len(flat) < 2 {
		return Metric{Tokens: len(flat), Lines: len(lines), Types: len(flat), GlyphStatus: "NOT_APPLICABLE"}, nil
	}
	freq := map[string]int{}
	for _, t := range flat {
		freq[t]++
	}
	types := len(freq)
	keys := make([]string, 0, len(freq))
	for t := range freq {
		keys = append(keys, t)
	}
	sort.Slice(keys, func(i, j int) bool {
		if freq[keys[i]] != freq[keys[j]] {
			return freq[keys[i]] > freq[keys[j]]
		}
		return keys[i] < keys[j]
	})
	ret := map[string]bool{}
	for i := 0; i < len(keys) && i < cap; i++ {
		ret[keys[i]] = true
	}
	pairsL, pairsR := []int{}, []int{}
	for i := 0; i < len(flat)-1; i++ {
		if lineIDs[i] == lineIDs[i+1] {
			pairsL = append(pairsL, i)
			pairsR = append(pairsR, i+1)
		}
	}
	calc := func(order []int) (float64, float64) {
		tokL, tokR, edgeL, edgeR := []string{}, []string{}, []string{}, []string{}
		for k := 0; k < len(order)-1; k++ {
			a, b := order[k], order[k+1]
			if lineIDs[a] != lineIDs[b] {
				continue
			}
			ta, tb := flat[a], flat[b]
			la := ta
			if !ret[ta] {
				la = "<other>"
			}
			lb := tb
			if !ret[tb] {
				lb = "<other>"
			}
			tokL = append(tokL, la)
			tokR = append(tokR, lb)
			if c.GlyphMode != "opaque" {
				edgeL = append(edgeL, feature(ta, c.GlyphMode, "last"))
				edgeR = append(edgeR, feature(tb, c.GlyphMode, "first"))
			}
		}
		e := 0.0
		if len(edgeL) > 0 {
			e = mi(edgeL, edgeR)
		}
		return mi(tokL, tokR), e
	}
	order := make([]int, len(flat))
	for i := range order {
		order[i] = i
	}
	to, eo := calc(order)
	tn, en := []float64{}, []float64{}
	rng := rand.New(rand.NewSource(seed))
	for s := 0; s < shuffles; s++ {
		p := make([]int, len(order))
		copy(p, order)
		for li := range lines {
			start := 0
			for start < len(lineIDs) && lineIDs[start] < li {
				start++
			}
			end := start
			for end < len(lineIDs) && lineIDs[end] == li {
				end++
			}
			for i := end - 1; i > start; i-- {
				j := start + rng.Intn(i-start+1)
				p[i], p[j] = p[j], p[i]
			}
		}
		a, b := calc(p)
		tn = append(tn, a)
		en = append(en, b)
	}
	m := Metric{Tokens: len(flat), Lines: len(lines), Pairs: len(pairsL), Types: types, TokenObserved: to, TokenShuffleMean: mean(tn), TokenShuffleSD: sd(tn)}
	m.TokenCorrected = m.TokenObserved - m.TokenShuffleMean // normalized by entropy of capped token identities
	den := map[string]int{}
	for _, t := range flat {
		if ret[t] {
			den[t]++
		} else {
			den["<other>"]++
		}
	}
	m.TokenShare = m.TokenCorrected / entropy(den)
	if c.GlyphMode == "opaque" {
		m.GlyphStatus = "NOT_APPLICABLE_OPAQUE_TOKENS"
	} else {
		m.GlyphStatus = "APPLICABLE"
		m.EdgeObserved = eo
		m.EdgeShuffleMean = mean(en)
		m.EdgeShuffleSD = sd(en)
		m.EdgeCorrected = eo - m.EdgeShuffleMean
		first := map[string]int{}
		for _, i := range pairsR {
			first[feature(flat[i], c.GlyphMode, "first")]++
		}
		if h := entropy(first); h > 0 {
			m.EdgeShare = m.EdgeCorrected / h
		}
	}
	return m, nil
}
func mean(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	x := 0.0
	for _, z := range v {
		x += z
	}
	return x / float64(len(v))
}
