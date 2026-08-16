package replicatedlocalaudit

import (
	"math"
	"math/rand"
	"sort"
	"strings"
)

const maxDistance = 20

type profile map[string][2][maxDistance]map[string]int

func splitBlockLines(b block) [][]token {
	var out [][]token
	for i := 0; i < len(b.Tokens); {
		j := i + 1
		for j < len(b.Tokens) && b.Tokens[j].Line == b.Tokens[i].Line {
			j++
		}
		out = append(out, b.Tokens[i:j])
		i = j
	}
	return out
}
func buildProfile(b block) profile {
	p := profile{}
	for _, line := range splitBlockLines(b) {
		for i, t := range line {
			z, ok := p[t.Text]
			if !ok {
				for side := 0; side < 2; side++ {
					for d := 0; d < maxDistance; d++ {
						z[side][d] = map[string]int{}
					}
				}
			}
			for d := 1; d <= maxDistance; d++ {
				if i-d >= 0 {
					z[0][d-1][line[i-d].Text]++
				}
				if i+d < len(line) {
					z[1][d-1][line[i+d].Text]++
				}
			}
			p[t.Text] = z
		}
	}
	return p
}
func mergeProfiles(all map[string]profile, exclude string) profile {
	o := profile{}
	for id, p := range all {
		if id == exclude {
			continue
		}
		for tok, z := range p {
			q, ok := o[tok]
			if !ok {
				for s := 0; s < 2; s++ {
					for d := 0; d < maxDistance; d++ {
						q[s][d] = map[string]int{}
					}
				}
			}
			for s := 0; s < 2; s++ {
				for d := 0; d < maxDistance; d++ {
					for k, v := range z[s][d] {
						q[s][d][k] += v
					}
				}
			}
			o[tok] = q
		}
	}
	return o
}
func countMap(m map[string]int) int {
	n := 0
	for _, v := range m {
		n += v
	}
	return n
}
func jsSimilarity(a, b map[string]int) float64 {
	ta, tb := countMap(a), countMap(b)
	if ta == 0 || tb == 0 {
		return 0
	}
	keySet := map[string]bool{}
	for k := range a {
		keySet[k] = true
	}
	for k := range b {
		keySet[k] = true
	}
	// Accumulate in sorted key order: map iteration order is randomized
	// independently per range statement execution, so summing in `range
	// keySet` order made this float64 accumulation nondeterministic across
	// otherwise byte-identical calls (see determinism_test.go).
	keys := make([]string, 0, len(keySet))
	for k := range keySet {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	div := 0.
	for _, k := range keys {
		pa, pb := float64(a[k])/float64(ta), float64(b[k])/float64(tb)
		m := (pa + pb) / 2
		if pa > 0 {
			div += .5 * pa * math.Log(pa/m)
		}
		if pb > 0 {
			div += .5 * pb * math.Log(pb/m)
		}
	}
	return math.Max(0, math.Min(1, 1-div/math.Ln2))
}
func compareProfiles(left, right profile, a, b string) (float64, int) {
	sum := 0.
	n := 0
	for _, tok := range []string{a, b} {
		lz, lok := left[tok]
		rz, rok := right[tok]
		if !lok || !rok {
			continue
		}
		for s := 0; s < 2; s++ {
			for d := 0; d < maxDistance; d++ {
				if countMap(lz[s][d]) > 0 && countMap(rz[s][d]) > 0 {
					sum += jsSimilarity(lz[s][d], rz[s][d])
					n++
				}
			}
		}
	}
	if n == 0 {
		return 0, 0
	}
	return sum / float64(n), n
}
func tokenCount(p profile, t string) int {
	z, ok := p[t]
	if !ok {
		return 0
	}
	n := 0
	for _, m := range z[0] {
		n += countMap(m)
	}
	return int(math.Round(float64(n) / maxDistance))
}
func pairShape(p profile, a, b string) []float64 {
	v := make([]float64, 2*maxDistance)
	za, oka := p[a]
	zb, okb := p[b]
	if !oka || !okb {
		return v
	}
	for d := 0; d < maxDistance; d++ {
		v[d] = float64(za[1][d][b] + zb[0][d][a])
		v[maxDistance+d] = float64(za[0][d][b] + zb[1][d][a])
	}
	return v
}
func cosine(a, b []float64) float64 {
	dot, aa, bb := 0., 0., 0.
	for i := range a {
		dot += a[i] * b[i]
		aa += a[i] * a[i]
		bb += b[i] * b[i]
	}
	if aa == 0 || bb == 0 {
		return 0
	}
	return dot / math.Sqrt(aa*bb)
}
func shapeStats(v []float64) (peak int, center, asym float64) {
	sum, left, right := 0., 0., 0.
	best := -1.
	for i, x := range v {
		if x > best {
			best = x
			if i < maxDistance {
				peak = i + 1
			} else {
				peak = -(i - maxDistance + 1)
			}
		}
		sum += x
		if i < maxDistance {
			right += x
			center += x * float64(i+1)
		} else {
			left += x
			center += x * float64(i-maxDistance+1)
		}
	}
	if sum > 0 {
		center /= sum
		asym = (right - left) / sum
	} else {
		peak = 0
	}
	return
}

func sequenceObserved(c sequenceCandidate, tokens []token, blocks []block) seqObserved {
	vocab := map[string]int{}
	for _, t := range tokens {
		vocab[t.Text]++
	}
	o := seqObserved{Validity: "canonical-clean"}
	for _, x := range c.Tokens {
		o.TokenOccurrences = append(o.TokenOccurrences, vocab[x])
		if vocab[x] == 0 {
			o.Validity = "absent-from-canonical"
			o.AbsentTokens = append(o.AbsentTokens, x)
		}
		if strings.Contains(x, "?") {
			o.ContainsQuestion = true
		}
		if strings.Contains(x, "@") {
			o.ContainsAt = true
		}
		if !standardToken(x) && !strings.ContainsAny(x, "?@") {
			o.ContainsOtherMarker = true
		}
		if o.Validity != "absent-from-canonical" && (o.ContainsQuestion || o.ContainsAt || o.ContainsOtherMarker) {
			o.Validity = "ambiguous-transcription"
		}
	}
	o.Total = countSequence(tokens, c.Tokens)
	counts := map[string]int{}
	joints := map[string]bool{}
	curs := map[string]bool{}
	hands := map[string]bool{}
	for _, b := range blocks {
		n := countSequence(b.Tokens, c.Tokens)
		o.Eligible += n
		if n > 0 {
			counts[b.ID] = n
			joints[b.Joint] = true
			curs[b.Currier] = true
			hands[b.Hand] = true
		}
	}
	o.Blocks, o.Joint, o.Currier, o.Hands = len(counts), len(joints), len(curs), len(hands)
	if o.Total == 0 {
		o.Validity = "absent-from-canonical"
	}
	maxN := 0
	for _, n := range counts {
		if n > maxN {
			maxN = n
		}
	}
	if o.Eligible > 0 {
		o.MaxFraction = float64(maxN) / float64(o.Eligible)
	}
	// Accumulate in sorted block-ID order: map iteration order is
	// randomized independently per range statement execution, so summing
	// in `range counts` order made this float64 accumulation
	// nondeterministic across otherwise byte-identical calls (see
	// determinism_test.go).
	blockIDs := make([]string, 0, len(counts))
	for id := range counts {
		blockIDs = append(blockIDs, id)
	}
	sort.Strings(blockIDs)
	total := float64(max(1, sumCounts(counts)))
	for _, id := range blockIDs {
		p := float64(counts[id]) / total
		o.Entropy -= p * math.Log2(p)
	}
	return o
}
func standardToken(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
			return false
		}
	}
	return true
}
func sumCounts(m map[string]int) int {
	n := 0
	for _, v := range m {
		n += v
	}
	return n
}
func countSequence(tokens []token, seq []string) int {
	n := 0
	for i := 0; i+len(seq) <= len(tokens); i++ {
		ok := true
		line := tokens[i].Line
		for j, x := range seq {
			if tokens[i+j].Line != line || tokens[i+j].Text != x {
				ok = false
				break
			}
		}
		if ok {
			n++
		}
	}
	return n
}
func sequenceStats(cs []sequenceCandidate, blocks []block) (map[string]int, map[string]int) {
	tot, bc := map[string]int{}, map[string]int{}
	for _, c := range cs {
		seen := 0
		for _, b := range blocks {
			n := countSequence(b.Tokens, c.Tokens)
			tot[c.ID] += n
			if n > 0 {
				seen++
			}
		}
		bc[c.ID] = seen
	}
	return tot, bc
}
func shuffledBlocks(blocks []block, seed int64) []block {
	r := rand.New(rand.NewSource(seed))
	out := make([]block, len(blocks))
	for i, b := range blocks {
		out[i] = b
		out[i].Tokens = append([]token(nil), b.Tokens...)
		texts := make([]string, len(b.Tokens))
		for j := range b.Tokens {
			texts[j] = b.Tokens[j].Text
		}
		r.Shuffle(len(texts), func(a, z int) { texts[a], texts[z] = texts[z], texts[a] })
		for j := range out[i].Tokens {
			out[i].Tokens[j].Text = texts[j]
		}
	}
	return out
}
func bh(p []float64) []float64 {
	idx := make([]int, len(p))
	for i := range idx {
		idx[i] = i
	}
	sort.Slice(idx, func(i, j int) bool { return p[idx[i]] < p[idx[j]] })
	q := make([]float64, len(p))
	next := 1.
	for rank := len(idx); rank >= 1; rank-- {
		i := idx[rank-1]
		v := p[i] * float64(len(idx)) / float64(rank)
		if v > next {
			v = next
		}
		if v > 1 {
			v = 1
		}
		q[i] = v
		next = v
	}
	return q
}
func quantile(x []float64, p float64) float64 {
	if len(x) == 0 {
		return 0
	}
	y := append([]float64(nil), x...)
	sort.Float64s(y)
	i := int(math.Ceil(p*float64(len(y)))) - 1
	if i < 0 {
		i = 0
	}
	if i >= len(y) {
		i = len(y) - 1
	}
	return y[i]
}
func meanSD(x []float64) (float64, float64) {
	if len(x) == 0 {
		return 0, 0
	}
	m := 0.
	for _, v := range x {
		m += v
	}
	m /= float64(len(x))
	s := 0.
	for _, v := range x {
		s += (v - m) * (v - m)
	}
	return m, math.Sqrt(s / float64(len(x)))
}
