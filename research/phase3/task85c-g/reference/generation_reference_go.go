// Command generation_reference_go is an independent Go implementation of
// the normative V1 generation primitives. It has no production-data access.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"strings"
	"unicode/utf8"
)

func add(s, c, x float64) (float64, float64) {
	t := s + x
	if math.Abs(s) >= math.Abs(x) {
		c += (s - t) + x
	} else {
		c += (x - t) + s
	}
	return t, c
}

func prepare(labels []string, weights []float64, allowed map[string]bool) (string, []string, []float64) {
	var out []string
	var p []float64
	s, c := 0.0, 0.0
	for i, label := range labels {
		if !allowed[label] {
			continue
		}
		w := weights[i]
		if w < 0 || math.IsNaN(w) || math.IsInf(w, 0) {
			return "NUMERICAL_FAILURE", nil, nil
		}
		out = append(out, label)
		p = append(p, w)
		s, c = add(s, c, w)
	}
	z := s + c
	if z <= 0 || math.IsNaN(z) || math.IsInf(z, 0) {
		return "GENERATION_FAILURE", nil, nil
	}
	for i := range p {
		p[i] /= z
	}
	return "OK", out, p
}

func categorical(labels []string, weights []float64, allowed map[string]bool, u float64) (string, string, int) {
	status, out, p := prepare(labels, weights, allowed)
	if status != "OK" {
		return status, "", 0
	}
	s, c, last := 0.0, 0.0, ""
	for i, x := range out {
		s, c = add(s, c, p[i])
		if p[i] > 0 {
			last = x
			if u < s+c {
				return "OK", x, 1
			}
		}
	}
	return "OK", last, 1
}

func race(labels []string, weights []float64, allowed map[string]bool, u []float64) (string, string, int) {
	status, out, p := prepare(labels, weights, allowed)
	if status != "OK" {
		return status, "", 0
	}
	best, answer, draws := math.Inf(1), "", 0
	for i, x := range out {
		if p[i] == 0 {
			continue
		}
		v := math.Inf(1)
		if u[draws] != 0 {
			v = -math.Log(u[draws]) / p[i]
		}
		if answer == "" || v < best {
			best, answer = v, x
		}
		draws++
	}
	return "OK", answer, draws
}

func aliasTable(labels []string, weights []float64, allowed map[string]bool) (string, []string, []float64, []int) {
	status, out, p := prepare(labels, weights, allowed)
	if status != "OK" {
		return status, nil, nil, nil
	}
	n := len(out)
	q := make([]float64, n)
	cutoff := make([]float64, n)
	alias := make([]int, n)
	var small, large []int
	for i := range p {
		q[i] = p[i] * float64(n)
		alias[i] = i
		if q[i] < 1 {
			small = append(small, i)
		} else {
			large = append(large, i)
		}
	}
	for len(small) > 0 && len(large) > 0 {
		s, l := small[0], large[0]
		small = small[1:]
		large = large[1:]
		cutoff[s] = q[s]
		alias[s] = l
		q[l] = (q[l] + q[s]) - 1
		if q[l] < 1 {
			small = insertSorted(small, l)
		} else {
			large = insertSorted(large, l)
		}
	}
	for _, i := range append(small, large...) {
		cutoff[i] = 1
		alias[i] = i
	}
	return "OK", out, cutoff, alias
}

func insertSorted(v []int, x int) []int {
	i := 0
	for i < len(v) && v[i] < x {
		i++
	}
	v = append(v, 0)
	copy(v[i+1:], v[i:])
	v[i] = x
	return v
}

func aliasSample(labels []string, weights []float64, allowed map[string]bool, u1, u2 float64) (string, string, int) {
	status, out, p, a := aliasTable(labels, weights, allowed)
	if status != "OK" {
		return status, "", 0
	}
	col := int(u1 * float64(len(out)))
	if u2 < p[col] {
		return "OK", out[col], 2
	}
	return "OK", out[a[col]], 2
}

func serialize(tokens []string) ([]byte, error) {
	if len(tokens) == 0 {
		return []byte{}, nil
	}
	for _, t := range tokens {
		if !utf8.ValidString(t) {
			return nil, fmt.Errorf("invalid UTF-8")
		}
		if t == "" || utf8.RuneCountInString(t) > 64 || strings.ContainsAny(t, "\r\n") || t == "<BOS>" || t == "<EOS>" || t == "<UNK>" {
			return nil, fmt.Errorf("invalid token")
		}
	}
	// The state machine's logical token domain is already NFC; serialization
	// performs no second normalization transform.
	return []byte(strings.Join(tokens, "\n") + "\n"), nil
}

func main() {
	labels := []string{"a", "b", "c", "d", "<EOS>"}
	weights := []float64{.28, .22, .18, .12, .20}
	allowed := map[string]bool{"a": true, "b": true, "c": true, "d": true}
	s, x, n := categorical(labels, weights, allowed, .92848667210989588)
	if s != "OK" || x != "d" || n != 1 {
		panic("PF-SC01")
	}
	s, x, n = race([]string{"a", "b", "c"}, []float64{.5, .3, .2}, map[string]bool{"a": true, "b": true, "c": true}, []float64{.2, .8, .5})
	if s != "OK" || x != "b" || n != 3 {
		panic("race")
	}
	s, x, n = aliasSample([]string{"a", "b", "c"}, []float64{.5, .3, .2}, map[string]bool{"a": true, "b": true, "c": true}, .8, .4)
	if s != "OK" || x != "c" || n != 2 {
		panic("alias")
	}
	b, err := serialize([]string{"a", "café"})
	if err != nil || hex.EncodeToString(b) != "610a636166c3a90a" {
		panic("serialization")
	}
	h := sha256.Sum256(b)
	fmt.Printf("GO_REFERENCE=PASS\nPF_SC01=d/1\nSERIAL_SHA256=%x\n", h)
	if len(os.Args) > 1 && os.Args[1] == "--quiet" {
		return
	}
}
