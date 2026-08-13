package graphemic

import (
	"math"
	"sort"
	"unicode/utf8"
)

// TokenizeGraphemes preserves every sign, while treating IVTFF @NNN; codes as
// one grapheme. All other valid Unicode code points (including ?) are one item.
func TokenizeGraphemes(token string) []string {
	result := make([]string, 0, utf8.RuneCountInString(token))
	for i := 0; i < len(token); {
		if token[i] == '@' {
			j := i + 1
			for j < len(token) && token[j] >= '0' && token[j] <= '9' {
				j++
			}
			if j > i+1 && j < len(token) && token[j] == ';' {
				result = append(result, token[i:j+1])
				i = j + 1
				continue
			}
		}
		r, size := utf8.DecodeRuneInString(token[i:])
		if r == utf8.RuneError && size == 0 {
			break
		}
		result = append(result, token[i:i+size])
		i += size
	}
	return result
}

func Levenshtein(a, b []string) int {
	if len(a) > len(b) {
		a, b = b, a
	}
	previous := make([]int, len(a)+1)
	for i := range previous {
		previous[i] = i
	}
	for j, gb := range b {
		current := make([]int, len(a)+1)
		current[0] = j + 1
		for i, ga := range a {
			cost := 1
			if ga == gb {
				cost = 0
			}
			current[i+1] = min(current[i]+1, previous[i+1]+1, previous[i]+cost)
		}
		previous = current
	}
	return previous[len(a)]
}

func GraphemeMetrics(a, b string) (distance int, normalized float64, similarity float64, prefix int, suffix int, lengthDifference int) {
	ga, gb := TokenizeGraphemes(a), TokenizeGraphemes(b)
	distance = Levenshtein(ga, gb)
	denominator := max(len(ga), len(gb))
	if denominator > 0 {
		normalized = float64(distance) / float64(denominator)
	}
	similarity = 1 - normalized
	for prefix < min(len(ga), len(gb)) && ga[prefix] == gb[prefix] {
		prefix++
	}
	for suffix < min(len(ga), len(gb)) && ga[len(ga)-1-suffix] == gb[len(gb)-1-suffix] {
		suffix++
	}
	lengthDifference = len(ga) - len(gb)
	if lengthDifference < 0 {
		lengthDifference = -lengthDifference
	}
	return
}

func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	v := append([]float64(nil), values...)
	sort.Float64s(v)
	pos := p * float64(len(v)-1)
	lo, hi := int(math.Floor(pos)), int(math.Ceil(pos))
	if lo == hi {
		return v[lo]
	}
	return v[lo] + (v[hi]-v[lo])*(pos-float64(lo))
}

func pearson(x, y []float64) float64 {
	if len(x) != len(y) || len(x) == 0 {
		return 0
	}
	var sx, sy float64
	for i := range x {
		sx += x[i]
		sy += y[i]
	}
	mx, my := sx/float64(len(x)), sy/float64(len(x))
	var n, dx, dy float64
	for i := range x {
		a, b := x[i]-mx, y[i]-my
		n += a * b
		dx += a * a
		dy += b * b
	}
	if dx == 0 || dy == 0 {
		return 0
	}
	return n / math.Sqrt(dx*dy)
}

func ranks(v []float64) []float64 {
	idx := make([]int, len(v))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(i, j int) bool { return v[idx[i]] < v[idx[j]] })
	r := make([]float64, len(v))
	for i := 0; i < len(idx); {
		j := i + 1
		for j < len(idx) && v[idx[j]] == v[idx[i]] {
			j++
		}
		rank := (float64(i)+float64(j-1))/2 + 1
		for k := i; k < j; k++ {
			r[idx[k]] = rank
		}
		i = j
	}
	return r
}
func spearman(x, y []float64) float64 { return pearson(ranks(x), ranks(y)) }
