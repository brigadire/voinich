// Package tokenformation implements Task62's deliberately small local token
// formation models. It is independent of the production pipeline.
package tokenformation

import (
	"math"
	"math/rand"
	"sort"
)

type Kind string

const (
	IID        Kind = "IID"
	PosIID     Kind = "POSITION_IID"
	Markov1    Kind = "MARKOV_1"
	Markov2    Kind = "MARKOV_2"
	PosMarkov1 Kind = "POSITION_MARKOV_1"
)

type Corpus struct {
	Tokens [][]string
	Lines  []int
}
type Model struct {
	Kind       Kind
	Alpha      float64
	Alphabet   []string
	lengths    []int
	lengthProb []float64
	glyph      map[string]int
	pos        map[string]map[string]int
	one        map[string]map[string]int
	two        map[string]map[string]int
	posOne     map[string]map[string]map[string]int
	total      int
}

func Fit(c Corpus, k Kind, alpha float64) Model {
	m := Model{Kind: k, Alpha: alpha, glyph: map[string]int{}, pos: map[string]map[string]int{}, one: map[string]map[string]int{}, two: map[string]map[string]int{}, posOne: map[string]map[string]map[string]int{}}
	lm := map[int]int{}
	for _, t := range c.Tokens {
		if len(t) == 0 {
			continue
		}
		m.lengths = append(m.lengths, len(t))
		lm[len(t)]++
		for i, g := range t {
			m.glyph[g]++
			p := region(len(t), i)
			if m.pos[p] == nil {
				m.pos[p] = map[string]int{}
			}
			m.pos[p][g]++
			if i > 0 {
				prev := t[i-1]
				if m.one[prev] == nil {
					m.one[prev] = map[string]int{}
				}
				m.one[prev][g]++
				if m.posOne[p] == nil {
					m.posOne[p] = map[string]map[string]int{}
				}
				if m.posOne[p][prev] == nil {
					m.posOne[p][prev] = map[string]int{}
				}
				m.posOne[p][prev][g]++
			}
			if i > 1 {
				key := t[i-2] + "\x00" + t[i-1]
				if m.two[key] == nil {
					m.two[key] = map[string]int{}
				}
				m.two[key][g]++
			}
		}
	}
	for g := range m.glyph {
		m.Alphabet = append(m.Alphabet, g)
		m.total++
	}
	sort.Strings(m.Alphabet)
	for l, n := range lm {
		m.lengthProb = append(m.lengthProb, float64(n)/float64(len(m.lengths)))
		_ = l
	}
	sort.Slice(m.lengthProb, func(i, j int) bool { return m.lengthProb[i] < m.lengthProb[j] })
	return m
}
func region(n, i int) string {
	if n <= 1 {
		return "SINGLETON"
	}
	if i == 0 {
		return "INITIAL"
	}
	if i == n-1 {
		return "FINAL"
	}
	return "MEDIAL"
}
func (m Model) sampleLength(r *rand.Rand) int {
	if len(m.lengths) == 0 {
		return 1
	}
	return m.lengths[r.Intn(len(m.lengths))]
}
func (m Model) draw(count map[string]int, r *rand.Rand) string {
	z := 0.
	for _, g := range m.Alphabet {
		z += float64(count[g]) + m.Alpha
	}
	x := r.Float64() * z
	for _, g := range m.Alphabet {
		x -= float64(count[g]) + m.Alpha
		if x <= 0 {
			return g
		}
	}
	return m.Alphabet[len(m.Alphabet)-1]
}
func (m Model) Generate(n int, r *rand.Rand) [][]string {
	out := make([][]string, n)
	for q := range out {
		l := m.sampleLength(r)
		t := make([]string, l)
		for i := range t {
			p := region(l, i)
			switch m.Kind {
			case IID:
				t[i] = m.draw(m.glyph, r)
			case PosIID:
				t[i] = m.draw(m.pos[p], r)
			case Markov1:
				if i == 0 {
					t[i] = m.draw(m.glyph, r)
				} else {
					t[i] = m.draw(m.one[t[i-1]], r)
				}
			case Markov2:
				if i < 2 {
					if i == 0 {
						t[i] = m.draw(m.glyph, r)
					} else {
						t[i] = m.draw(m.one[t[i-1]], r)
					}
				} else {
					t[i] = m.draw(m.two[t[i-2]+"\x00"+t[i-1]], r)
				}
			case PosMarkov1:
				if i == 0 {
					t[i] = m.draw(m.pos[p], r)
				} else {
					t[i] = m.draw(m.posOne[p][t[i-1]], r)
				}
			}
		}
		out[q] = t
	}
	return out
}
func (m Model) CrossEntropy(tokens [][]string) float64 {
	if len(tokens) == 0 {
		return 0
	}
	s := 0.
	n := 0
	for _, t := range tokens {
		for i, g := range t {
			var c map[string]int
			switch m.Kind {
			case IID:
				c = m.glyph
			case PosIID:
				c = m.pos[region(len(t), i)]
			case Markov1:
				if i == 0 {
					c = m.glyph
				} else {
					c = m.one[t[i-1]]
				}
			case Markov2:
				if i == 0 {
					c = m.glyph
				} else if i == 1 {
					c = m.one[t[i-1]]
				} else {
					c = m.two[t[i-2]+"\x00"+t[i-1]]
				}
			case PosMarkov1:
				if i == 0 {
					c = m.pos[region(len(t), i)]
				} else {
					c = m.posOne[region(len(t), i)][t[i-1]]
				}
			}
			den := float64(len(m.Alphabet)) * m.Alpha
			for _, v := range c {
				den += float64(v)
			}
			p := (float64(c[g]) + m.Alpha) / den
			s -= math.Log2(p)
			n++
		}
	}
	return s / float64(n)
}
