package numeric

import (
	"math"
	"math/rand"
)

func Control(c Corpus, kind string, seed int64) Corpus {
	r := rand.New(rand.NewSource(seed))
	out := c
	out.Name = c.Name + "_" + kind
	out.Tokens = make([]Token, len(c.Tokens))
	copy(out.Tokens, c.Tokens)
	switch kind {
	case "C1_WITHIN_TOKEN_GLYPH_SHUFFLE":
		for i := range out.Tokens {
			g := append([]byte(nil), out.Tokens[i].Glyphs...)
			r.Shuffle(len(g), func(a, b int) { g[a], g[b] = g[b], g[a] })
			out.Tokens[i].Glyphs = g
			out.Tokens[i].Text = string(g)
		}
	case "C2_TOKEN_SHUFFLE_WITHIN_LINE":
		for _, rr := range ranges(out.Tokens) {
			pool := make([]Token, rr[1]-rr[0])
			copy(pool, out.Tokens[rr[0]:rr[1]])
			r.Shuffle(len(pool), func(a, b int) { pool[a], pool[b] = pool[b], pool[a] })
			for j := range pool {
				dst := &out.Tokens[rr[0]+j]
				dst.Text = pool[j].Text
				dst.Glyphs = append([]byte(nil), pool[j].Glyphs...)
			}
		}
	case "C3_GLYPH_BIGRAM_MARKOV":
		start := map[byte]int{}
		trans := map[byte]map[byte]int{}
		for _, t := range c.Tokens {
			if len(t.Glyphs) > 0 {
				start[t.Glyphs[0]]++
			}
			for i := 1; i < len(t.Glyphs); i++ {
				if trans[t.Glyphs[i-1]] == nil {
					trans[t.Glyphs[i-1]] = map[byte]int{}
				}
				trans[t.Glyphs[i-1]][t.Glyphs[i]]++
			}
		}
		for i := range out.Tokens {
			n := len(out.Tokens[i].Glyphs)
			g := make([]byte, n)
			if n > 0 {
				g[0] = draw(start, c.Alphabet, r)
				for j := 1; j < n; j++ {
					g[j] = draw(trans[g[j-1]], c.Alphabet, r)
				}
			}
			out.Tokens[i].Glyphs = g
			out.Tokens[i].Text = string(g)
		}
	}
	uniq := map[string]bool{}
	for _, t := range out.Tokens {
		uniq[t.Text] = true
	}
	out.UniqueTokenCount = len(uniq)
	return out
}

func draw(counts map[byte]int, alphabet []byte, r *rand.Rand) byte {
	total := 0
	for _, g := range alphabet {
		total += counts[g]
	}
	if total == 0 {
		return alphabet[r.Intn(len(alphabet))]
	}
	x := r.Intn(total)
	for _, g := range alphabet {
		x -= counts[g]
		if x < 0 {
			return g
		}
	}
	return alphabet[len(alphabet)-1]
}

func Optimize(c Corpus, steps, restarts int, seed int64) ([]int, Metrics) {
	sample := sampleCorpus(c, 8)
	r := rand.New(rand.NewSource(seed))
	bestMap := BaselineMapping(len(c.Alphabet))
	best := objective(sample, bestMap)
	for restart := 0; restart < restarts; restart++ {
		m := BaselineMapping(len(c.Alphabet))
		r.Shuffle(len(m), func(i, j int) { m[i], m[j] = m[j], m[i] })
		cur := objective(sample, m)
		for step := 0; step < steps; step++ {
			i, j := r.Intn(len(m)), r.Intn(len(m))
			if i == j {
				continue
			}
			m[i], m[j] = m[j], m[i]
			next := objective(sample, m)
			temp := 0.02*(1-float64(step)/float64(max(1, steps))) + 0.0001
			if next.Score >= cur.Score || r.Float64() < math.Exp((next.Score-cur.Score)/temp) {
				cur = next
			} else {
				m[i], m[j] = m[j], m[i]
			}
			if cur.Score > best.Score {
				best = cur
				bestMap = append([]int(nil), m...)
			}
		}
	}
	candidate := Compute(c, bestMap)
	baselineMap := BaselineMapping(len(c.Alphabet))
	baseline := Compute(c, baselineMap)
	if baseline.Score >= candidate.Score {
		return baselineMap, baseline
	}
	return bestMap, candidate
}

func sampleCorpus(c Corpus, every int) Corpus {
	out := c
	out.Tokens = nil
	lineMap := map[int]int{}
	next := 0
	for _, t := range c.Tokens {
		if t.Line%every != 0 {
			continue
		}
		nt := t
		if _, ok := lineMap[t.Line]; !ok {
			lineMap[t.Line] = next
			next++
		}
		nt.Line = lineMap[t.Line]
		out.Tokens = append(out.Tokens, nt)
	}
	out.LineCount = next
	return out
}
