package structuralprojection

import "strings"

type profile struct {
	Right, Left []map[string]int
	Suffix      []map[string]int
}
type profiles map[string]*profile

func buildProfiles(c corpus, maxD int, bounded bool) profiles {
	p := profiles{}
	get := func(t string) *profile {
		x := p[t]
		if x == nil {
			x = &profile{Right: make([]map[string]int, maxD), Left: make([]map[string]int, maxD), Suffix: make([]map[string]int, 2)}
			for i := 0; i < maxD; i++ {
				x.Right[i] = map[string]int{}
				x.Left[i] = map[string]int{}
			}
			x.Suffix[0] = map[string]int{}
			x.Suffix[1] = map[string]int{}
			p[t] = x
		}
		return x
	}
	consume := func(seq []string) {
		for i, t := range seq {
			x := get(t)
			for d := 1; d <= maxD; d++ {
				if i+d < len(seq) {
					x.Right[d-1][seq[i+d]]++
				}
				if i-d >= 0 {
					x.Left[d-1][seq[i-d]]++
				}
			}
			for n := 2; n <= 3; n++ {
				if i+n < len(seq) {
					x.Suffix[n-2][strings.Join(seq[i+1:i+n+1], "\x1f")]++
				}
			}
		}
	}
	if bounded {
		for _, line := range c.Lines {
			consume(line)
		}
	} else {
		consume(c.Tokens)
	}
	return p
}

func shuffledCorpus(c corpus, mode string, seed int64) corpus {
	// Implemented without changing frequencies or line lengths.
	r := newRand(seed)
	out := corpus{Counts: c.Counts}
	if mode == "global" {
		out.Tokens = append([]string(nil), c.Tokens...)
		r.shuffle(out.Tokens)
		pos := 0
		for _, line := range c.Lines {
			n := len(line)
			out.Lines = append(out.Lines, append([]string(nil), out.Tokens[pos:pos+n]...))
			pos += n
		}
		return out
	}
	for _, line := range c.Lines {
		x := append([]string(nil), line...)
		r.shuffle(x)
		out.Lines = append(out.Lines, x)
		out.Tokens = append(out.Tokens, x...)
	}
	return out
}

type deterministicRand struct{ x uint64 }

func newRand(seed int64) *deterministicRand {
	return &deterministicRand{x: uint64(seed) + 0x9e3779b97f4a7c15}
}
func (r *deterministicRand) next() uint64 {
	r.x ^= r.x << 13
	r.x ^= r.x >> 7
	r.x ^= r.x << 17
	return r.x
}
func (r *deterministicRand) shuffle(x []string) {
	for i := len(x) - 1; i > 0; i-- {
		j := int(r.next() % uint64(i+1))
		x[i], x[j] = x[j], x[i]
	}
}
