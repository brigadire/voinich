// CONTRACT_REFERENCE_ONLY implements the repaired M0 construction independently.
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
)

type summary struct {
	Cases       int       `json:"cases"`
	Generation  int       `json:"generation_cases"`
	FixtureDen  float64   `json:"fixture_denominator"`
	FixtureProb []float64 `json:"fixture_probabilities"`
	Digest      string    `json:"digest"`
}

func fit(tokens []string, alpha float64) ([]rune, []float64, float64) {
	set := map[rune]bool{}
	for _, token := range tokens {
		for _, r := range token {
			set[r] = true
		}
	}
	ordinary := make([]rune, 0, len(set))
	for r := range set {
		ordinary = append(ordinary, r)
	}
	sort.Slice(ordinary, func(i, j int) bool { return string(ordinary[i]) < string(ordinary[j]) })
	counts := make([]float64, len(ordinary)+2)
	for _, token := range tokens {
		for _, r := range token {
			i := sort.Search(len(ordinary), func(i int) bool { return ordinary[i] >= r })
			counts[i]++
		}
		counts[len(counts)-1]++
	}
	n := 0.0
	for _, x := range counts {
		n += x
	}
	d := n + alpha*float64(len(counts))
	p := make([]float64, len(counts))
	for i := range p {
		p[i] = (counts[i] + alpha) / d
	}
	return ordinary, p, d
}

func cdf(p []float64, u float64) int {
	t := 0.0
	for i, x := range p {
		t += x
		if u < t {
			return i
		}
	}
	return len(p) - 1
}

func race(p, uniforms []float64) int {
	best, score := -1, math.Inf(1)
	for i, weight := range p {
		x := math.Inf(1)
		if uniforms[i] > 0 {
			x = -math.Log(uniforms[i]) / weight
		}
		if x < score {
			best, score = i, x
		}
	}
	return best
}

func main() {
	h := sha256.New()
	cases, gens := 32768, 8192
	for i := 0; i < cases; i++ {
		v := 1 + i%8
		n := 1 + (i/8)%12
		alpha := []float64{0, .1, .5, 1}[i%4]
		tokens := make([]string, n)
		for j := range tokens {
			b := make([]rune, 1+(i+j)%7)
			for k := range b {
				b[k] = rune('a' + (i+j+k)%v)
			}
			tokens[j] = string(b)
		}
		_, p, d := fit(tokens, alpha)
		if !(d > 0) {
			panic("denominator")
		}
		s := 0.0
		for _, x := range p {
			if x < 0 || math.IsNaN(x) {
				panic("probability")
			}
			s += x
		}
		if math.Abs(s-1) > 2e-15 {
			panic("normalization")
		}
		for _, x := range p {
			binary.Write(h, binary.BigEndian, math.Float64bits(x))
		}
	}
	for i := 0; i < gens; i++ {
		_, p, _ := fit([]string{"ab", "a"}, 1)
		if i < gens/2 {
			u := float64(i) / float64(gens/2)
			binary.Write(h, binary.BigEndian, uint64(cdf(p, u)))
		} else {
			u := make([]float64, len(p))
			for j := range u {
				u[j] = float64((i*1103515245+j*12345+1)%2147483647) / 2147483647
			}
			binary.Write(h, binary.BigEndian, uint64(race(p, u)))
		}
	}
	_, p, d := fit([]string{"ab", "a"}, 1)
	out := summary{cases, gens, d, p, hex.EncodeToString(h.Sum(nil))}
	b, _ := json.Marshal(out)
	fmt.Println(string(b))
}
