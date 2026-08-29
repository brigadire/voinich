package g1v2science

import (
	"bytes"
	"fmt"
	"golang.org/x/text/unicode/norm"
	"math"
	"sort"
	"strconv"
	"strings"
)

type GenerationResult struct {
	Tokens       []string `json:"tokens"`
	CorpusBytes  []byte   `json:"corpus_bytes"`
	CorpusSHA256 string   `json:"corpus_sha256"`
	Draws        uint64   `json:"draws"`
	Status       string   `json:"status"`
}

func SerializeCorpus(tokens []string) ([]byte, error) {
	var b bytes.Buffer
	for i, t := range tokens {
		if t == "" || !norm.NFC.IsNormalString(t) || strings.ContainsAny(t, "\r\n") {
			return nil, fmt.Errorf("invalid token %d", i)
		}
		b.WriteString(t)
		b.WriteByte('\n')
	}
	return b.Bytes(), nil
}
func filtered(out []string, w []float64, allowed map[string]bool) ([]string, []float64, error) {
	o := []string{}
	z := []float64{}
	for i, x := range out {
		if allowed[x] && w[i] > 0 {
			o = append(o, x)
			z = append(z, w[i])
		}
	}
	p, e := Normalize(z)
	return o, p, e
}
func DirectCDF(out []string, w []float64, allowed map[string]bool, u float64) (string, error) {
	o, p, e := filtered(out, w, allowed)
	if e != nil {
		return "", e
	}
	s, c := 0., 0.
	for i, x := range p {
		t := s + x
		if math.Abs(s) >= math.Abs(x) {
			c += (s - t) + x
		} else {
			c += (x - t) + s
		}
		s = t
		if u < s+c {
			return o[i], nil
		}
	}
	return o[len(o)-1], nil
}
func ExponentialRace(out []string, w []float64, allowed map[string]bool, r *RNG) (string, error) {
	o, p, e := filtered(out, w, allowed)
	if e != nil {
		return "", e
	}
	best := ""
	score := math.Inf(1)
	for i, x := range o {
		u := r.U53()
		v := math.Inf(1)
		if u > 0 {
			v = -math.Log(u) / p[i]
		}
		if v < score {
			score, best = v, x
		}
	}
	return best, nil
}
func WalkerAlias(out []string, w []float64, allowed map[string]bool, r *RNG) (string, error) {
	o, p, e := filtered(out, w, allowed)
	if e != nil {
		return "", e
	}
	n := len(p)
	q := make([]float64, n)
	cut := make([]float64, n)
	alias := make([]int, n)
	small, large := []int{}, []int{}
	for i, x := range p {
		q[i] = x * float64(n)
		if q[i] < 1 {
			small = append(small, i)
		} else {
			large = append(large, i)
		}
	}
	for len(small) > 0 && len(large) > 0 {
		s, l := small[0], large[0]
		small, large = small[1:], large[1:]
		cut[s], alias[s] = q[s], l
		q[l] += q[s] - 1
		if q[l] < 1 {
			small = insert(small, l)
		} else {
			large = insert(large, l)
		}
	}
	for _, i := range append(small, large...) {
		cut[i] = 1
		alias[i] = i
	}
	col := int(math.Floor(r.U53() * float64(n)))
	if r.U53() < cut[col] {
		return o[col], nil
	}
	return o[alias[col]], nil
}
func insert(x []int, v int) []int {
	i := sort.SearchInts(x, v)
	x = append(x, 0)
	copy(x[i+1:], x[i:])
	x[i] = v
	return x
}
func cumulative(out []string, w []float64, allowed map[string]bool, r *RNG) (string, error) {
	return DirectCDF(out, w, allowed, r.U53())
}
func all(xs []string) map[string]bool {
	m := map[string]bool{}
	for _, x := range xs {
		m[x] = true
	}
	return m
}
func admissible(xs []string, before bool) map[string]bool {
	m := map[string]bool{}
	for _, x := range xs {
		if x != "<BOS>" && (!before || x != "<EOS>") {
			m[x] = true
		}
	}
	return m
}
func ordered(r map[string]float64, v []string) ([]string, []float64) {
	o := append([]string{}, v...)
	w := make([]float64, len(o))
	for i, x := range o {
		w[i] = r[x]
	}
	return o, w
}
func sample(author, primitive string, o []string, w []float64, a map[string]bool, r *RNG) (string, error) {
	if author == "A" || author == "FITTED" {
		return DirectCDF(o, w, a, r.U53())
	}
	switch primitive {
	case "EXPONENTIAL_RACE":
		return ExponentialRace(o, w, a, r)
	case "WALKER_ALIAS":
		return WalkerAlias(o, w, a, r)
	case "CUMULATIVE_ROW":
		return cumulative(o, w, a, r)
	}
	return "", fmt.Errorf("primitive")
}

func GenerateFitted(m FittedModel, author string, count int, r *RNG) (GenerationResult, error) {
	if count <= 0 {
		return GenerationResult{Status: "PROTOCOL_VETO"}, fmt.Errorf("count")
	}
	if m.ModelClass == "M5" {
		return generateM5(m, author, count, r)
	}
	tokens := make([]string, 0, count)
	for n := 0; n < count; n++ {
		h := bos(m.Order)
		state := 0
		glyphs := []string{}
		if m.ModelClass == "M4" {
			names := stateNames(m.States)
			x, e := sample(author, "CUMULATIVE_ROW", names, m.Pi, all(names), r)
			if e != nil {
				return GenerationResult{Status: "GENERATION_FAILURE"}, e
			}
			state, _ = strconv.Atoi(x)
		}
		for len(glyphs) < 64 {
			var o []string
			var w []float64
			primitive := "EXPONENTIAL_RACE"
			if m.ModelClass == "M4" {
				o = append([]string{}, m.Vocabulary...)
				w = append([]float64{}, m.Emission[state]...)
				primitive = "CUMULATIVE_ROW"
			} else {
				o, w = ordered(m.rowFor(h, state), m.Vocabulary)
				if m.ModelClass == "M1" {
					primitive = "WALKER_ALIAS"
				}
			}
			x, e := sample(author, primitive, o, w, admissible(o, len(glyphs) == 0), r)
			if e != nil {
				return GenerationResult{Status: "GENERATION_FAILURE"}, e
			}
			if x == "<EOS>" {
				break
			}
			if x == "<UNK>" {
				glyphs = append(glyphs, "�")
			} else {
				glyphs = append(glyphs, x)
			}
			switch m.ModelClass {
			case "M1", "M2":
				h = append(h, x)
				if len(h) > m.Order {
					h = h[len(h)-m.Order:]
				}
			case "M3":
				state = m.Transitions[strconv.Itoa(state)][x]
			case "M4":
				if len(glyphs) < 64 {
					names := stateNames(m.States)
					y, e := sample(author, "CUMULATIVE_ROW", names, m.Transition[state], all(names), r)
					if e != nil {
						return GenerationResult{Status: "GENERATION_FAILURE"}, e
					}
					state, _ = strconv.Atoi(y)
				}
			}
		}
		if len(glyphs) == 0 {
			return GenerationResult{Status: "GENERATION_FAILURE"}, fmt.Errorf("empty token")
		}
		tokens = append(tokens, strings.Join(glyphs, ""))
	}
	b, e := SerializeCorpus(tokens)
	if e != nil {
		return GenerationResult{Status: "PROTOCOL_VETO"}, e
	}
	return GenerationResult{tokens, b, Hash(b), r.Draw, "GENERATION_SUCCESS"}, nil
}
func stateNames(n int) []string {
	x := make([]string, n)
	for i := range x {
		x[i] = strconv.Itoa(i)
	}
	return x
}
func mapRow(m map[string]float64) ([]string, []float64) {
	o := make([]string, 0, len(m))
	for x := range m {
		o = append(o, x)
	}
	sort.Strings(o)
	w := make([]float64, len(o))
	for i, x := range o {
		w[i] = m[x]
	}
	return o, w
}
func generateM5(m FittedModel, author string, count int, r *RNG) (GenerationResult, error) {
	g := m.Grammar
	tokens := []string{}
	for len(tokens) < count {
		accepted := ""
		for attempt := 0; attempt < 1024; attempt++ {
			kind, e := sample(author, "EXPONENTIAL_RACE", []string{"DIRECT", "PRODUCTIVE"}, []float64{1 - g.BackoffWeight, g.BackoffWeight}, all([]string{"DIRECT", "PRODUCTIVE"}), r)
			if e != nil {
				return GenerationResult{Status: "GENERATION_FAILURE"}, e
			}
			proposal := ""
			if kind == "PRODUCTIVE" {
				po, pw := mapRow(g.Prefix)
				so, sw := mapRow(g.Stem)
				xo, xw := mapRow(g.Suffix)
				p, _ := sample(author, "EXPONENTIAL_RACE", po, pw, all(po), r)
				s, _ := sample(author, "EXPONENTIAL_RACE", so, sw, all(so), r)
				x, _ := sample(author, "EXPONENTIAL_RACE", xo, xw, all(xo), r)
				proposal = p + s + x
			} else {
				pool := g.Rules
				if len(pool) == 0 {
					pool = g.Exceptions
				}
				if len(pool) == 0 {
					continue
				}
				o, w := mapRow(pool)
				proposal, _ = sample(author, "EXPONENTIAL_RACE", o, w, all(o), r)
			}
			if proposal != "" && len([]rune(proposal)) <= 64 {
				accepted = proposal
				break
			}
		}
		if accepted == "" {
			return GenerationResult{Status: "GENERATION_FAILURE"}, fmt.Errorf("attempt cap")
		}
		tokens = append(tokens, accepted)
	}
	b, _ := SerializeCorpus(tokens)
	return GenerationResult{tokens, b, Hash(b), r.Draw, "GENERATION_SUCCESS"}, nil
}

func GenerateSynthetic(route GenerationRoute, count int, r *RNG) (GenerationResult, error) {
	if count <= 0 {
		return GenerationResult{Status: "PROTOCOL_VETO"}, fmt.Errorf("count")
	}
	tokens := []string{}
	for len(tokens) < count {
		glyphs := []string{}
		for len(glyphs) < 64 {
			out := []string{"a", "b", "c", "d", "<EOS>"}
			w := []float64{.28, .22, .18, .12, .20}
			if route.Model != "M0" {
				w = []float64{.25, .22, .2, .18, .15}
			}
			x, e := sample(route.Author, route.Primitive, out, w, admissible(out, len(glyphs) == 0), r)
			if e != nil {
				return GenerationResult{Status: "GENERATION_FAILURE"}, e
			}
			if x == "<EOS>" {
				break
			}
			glyphs = append(glyphs, x)
		}
		if len(glyphs) == 0 {
			continue
		}
		tokens = append(tokens, strings.Join(glyphs, ""))
	}
	b, _ := SerializeCorpus(tokens)
	return GenerationResult{tokens, b, Hash(b), r.Draw, "GENERATION_SUCCESS"}, nil
}
