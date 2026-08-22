// Package characterentropy contains the independent Task61 estimator.  It is
// intentionally separate from the production pipeline and uses evaglyph for
// Voynich glyph tokenisation.
package characterentropy

import (
	"math"
	"math/rand"
	"sort"
)

type Mode string

const (
	Continuous    Mode = "CONTINUOUS"
	TokenBoundary Mode = "TOKEN_BOUNDARY"
	WithinToken   Mode = "WITHIN_TOKEN_ONLY"
)

type Corpus struct {
	Name   string
	Tokens [][]string
	Lines  []int
}
type Estimate struct {
	Order, Samples, Contexts, UniqueContexts int
	H, Normalized, Coverage                  float64
	Status                                   string
}

func H(counts map[string]int) float64 {
	n := 0
	for _, v := range counts {
		n += v
	}
	if n == 0 {
		return 0
	}
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	x := 0.0
	for _, k := range keys {
		p := float64(counts[k]) / float64(n)
		x -= p * math.Log2(p)
	}
	return x
}

// Entropy computes the plug-in Shannon estimate H(X_i | X_{i-1}...X_{i-k}).
// Each context is counted once per observed continuation.  A line reset is
// represented by a new segment in streams; no newline is silently a glyph.
func Entropy(tokens [][]string, lines []int, mode Mode, order int, resetLines bool) Estimate {
	if order < 0 {
		return Estimate{Order: order, Status: "INVALID_ORDER"}
	}
	seqs := Streams(tokens, lines, mode, resetLines)
	counts := map[string]map[string]int{}
	samples := 0
	for _, s := range seqs {
		if len(s) <= order {
			continue
		}
		for i := order; i < len(s); i++ {
			key := ""
			for j := i - order; j < i; j++ {
				key += s[j] + "\x00"
			}
			if counts[key] == nil {
				counts[key] = map[string]int{}
			}
			counts[key][s[i]]++
			samples++
		}
	}
	if samples == 0 {
		return Estimate{Order: order, Status: "INSUFFICIENT_DATA"}
	}
	ctxKeys := make([]string, 0, len(counts))
	for k := range counts {
		ctxKeys = append(ctxKeys, k)
	}
	sort.Strings(ctxKeys)
	total := 0.0
	contexts := 0
	unique := 0
	for _, ck := range ctxKeys {
		next := counts[ck]
		contexts++
		n := 0
		for _, w := range next {
			n += w
		}
		contKeys := make([]string, 0, len(next))
		for k := range next {
			contKeys = append(contKeys, k)
		}
		sort.Strings(contKeys)
		for _, nk := range contKeys {
			v := next[nk]
			unique++
			p := float64(v) / float64(n)
			total -= float64(v) / float64(samples) * math.Log2(p)
		}
	}
	inv := 1.0
	alphabet := map[string]bool{}
	for _, s := range seqs {
		for _, g := range s {
			alphabet[g] = true
		}
	}
	if len(alphabet) > 1 {
		inv = 1 / math.Log2(float64(len(alphabet)))
	}
	return Estimate{Order: order, Samples: samples, Contexts: contexts, UniqueContexts: unique, H: total, Normalized: total * inv, Coverage: float64(contexts) / float64(samples), Status: "OK"}
}

func Streams(tokens [][]string, lines []int, mode Mode, resetLines bool) [][]string {
	out := [][]string{}
	cur := []string{}
	last := -1
	flush := func() {
		if len(cur) > 0 {
			out = append(out, cur)
			cur = nil
		}
	}
	for i, t := range tokens {
		line := i
		if i < len(lines) {
			line = lines[i]
		}
		if resetLines && last >= 0 && line != last {
			flush()
		}
		last = line
		if mode == WithinToken {
			out = append(out, append([]string(nil), t...))
			continue
		}
		if mode == TokenBoundary && len(cur) > 0 {
			cur = append(cur, "<WB>")
		}
		cur = append(cur, t...)
	}
	flush()
	return out
}

func GlyphShuffle(tokens [][]string, r *rand.Rand) [][]string {
	out := clone(tokens)
	all := []string{}
	for _, t := range out {
		all = append(all, t...)
	}
	r.Shuffle(len(all), func(i, j int) { all[i], all[j] = all[j], all[i] })
	q := 0
	for i, t := range out {
		for j := range t {
			out[i][j] = all[q]
			q++
		}
	}
	return out
}
func WithinTokenShuffle(tokens [][]string, r *rand.Rand) [][]string {
	out := clone(tokens)
	for _, t := range out {
		r.Shuffle(len(t), func(i, j int) { t[i], t[j] = t[j], t[i] })
	}
	return out
}
func TokenShuffle(tokens [][]string, r *rand.Rand) [][]string {
	out := clone(tokens)
	r.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}
func WithinLineTokenShuffle(tokens [][]string, lines []int, r *rand.Rand) [][]string {
	out := clone(tokens)
	groups := map[int][]int{}
	for i := range out {
		l := i
		if i < len(lines) {
			l = lines[i]
		}
		groups[l] = append(groups[l], i)
	}
	keys := make([]int, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		ix := groups[k]
		r.Shuffle(len(ix), func(i, j int) { out[ix[i]], out[ix[j]] = out[ix[j]], out[ix[i]] })
	}
	return out
}
func clone(x [][]string) [][]string {
	y := make([][]string, len(x))
	for i, t := range x {
		y[i] = append([]string(nil), t...)
	}
	return y
}
