package tokenrelationvalidation

import (
	"math"
	"math/rand"
	"sort"
	"strings"

	"zcore.dev/voinich/internal/profilestability"
)

func splitLines(b Block) [][]Token {
	var out [][]Token
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

func directionForBlock(c Candidate, b Block, maxD int) DirectionBlock {
	x := DirectionBlock{CandidateID: c.ID, A: c.A, B: c.B, BlockID: b.ID, Currier: b.Currier, Hand: b.Hand, Joint: b.Joint}
	if c.FrozenThreshold > float64(maxD) && strings.Contains(c.Sources, "begin_end") {
		maxD = int(c.FrozenThreshold)
	}
	hitAB, hitBA, baselineA, baselineB, totalAnchors := 0, 0, 0, 0, 0
	for _, line := range splitLines(b) {
		for i, t := range line {
			totalAnchors++
			hasA, hasB := false, false
			if t.Text == c.A {
				x.CountA++
			}
			if t.Text == c.B {
				x.CountB++
			}
			for d := 1; d <= maxD && i+d < len(line); d++ {
				u := line[i+d].Text
				if u == c.A {
					hasA = true
				}
				if u == c.B {
					hasB = true
				}
				if t.Text == c.A && u == c.B {
					x.ABeforeB++
					if d == 1 {
						x.ImmediateAB++
					}
					if d <= 5 {
						x.ExactAB[d-1]++
					}
				}
				if t.Text == c.B && u == c.A {
					x.BBeforeA++
					if d == 1 {
						x.ImmediateBA++
					}
					if d <= 5 {
						x.ExactBA[d-1]++
					}
				}
			}
			if hasA {
				baselineA++
			}
			if hasB {
				baselineB++
			}
			if t.Text == c.A && hasB {
				hitAB++
			}
			if t.Text == c.B && hasA {
				hitBA++
			}
		}
	}
	x.Observations = x.ABeforeB + x.BBeforeA
	x.Eligible = x.CountA >= 5 && x.CountB >= 5 && x.Observations >= 5
	if x.Observations > 0 {
		x.Score = float64(x.ABeforeB-x.BBeforeA) / float64(x.Observations)
	}
	if totalAnchors > 0 {
		x.EnrichmentAB = localEnrichment(hitAB, x.CountA, baselineB, totalAnchors)
		x.EnrichmentBA = localEnrichment(hitBA, x.CountB, baselineA, totalAnchors)
	}
	return x
}

func localEnrichment(hits, anchors, baselineHits, positions int) float64 {
	if anchors == 0 || baselineHits == 0 || positions == 0 {
		return 0
	}
	return (float64(hits) / float64(anchors)) / (float64(baselineHits) / float64(positions))
}

type localProfiles struct {
	P map[string]profilestability.Profile
	D map[string][][]map[string]int
}

func mergeLocalProfiles(all map[string]localProfiles, exclude string, maxD int) localProfiles {
	out := localProfiles{P: map[string]profilestability.Profile{}, D: map[string][][]map[string]int{}}
	for id, x := range all {
		if id == exclude {
			continue
		}
		for token, p := range x.P {
			q := out.P[token]
			if q.Positions == nil {
				q.Positions = map[int]int{}
				q.Left = map[string]int{}
				q.Right = map[string]int{}
			}
			q.Count += p.Count
			for k, v := range p.Positions {
				q.Positions[k] += v
			}
			for k, v := range p.Left {
				q.Left[k] += v
			}
			for k, v := range p.Right {
				q.Right[k] += v
			}
			out.P[token] = q
		}
		for token, d := range x.D {
			z := out.D[token]
			if z == nil {
				z = make([][]map[string]int, 2)
				for side := 0; side < 2; side++ {
					z[side] = make([]map[string]int, maxD)
					for n := range z[side] {
						z[side][n] = map[string]int{}
					}
				}
				out.D[token] = z
			}
			for side := 0; side < 2; side++ {
				for n := 0; n < maxD; n++ {
					for k, v := range d[side][n] {
						z[side][n][k] += v
					}
				}
			}
		}
	}
	return out
}

func compareDistanceProfiles(left, right localProfiles, a, b string, maxD int) (float64, float64) {
	js, overlap, n := 0., 0., 0
	for _, token := range []string{a, b} {
		if left.D[token] == nil || right.D[token] == nil {
			continue
		}
		for side := 0; side < 2; side++ {
			for d := 0; d < maxD; d++ {
				if countMap(left.D[token][side][d]) == 0 || countMap(right.D[token][side][d]) == 0 {
					continue
				}
				x, o := jsOverlap(left.D[token][side][d], right.D[token][side][d])
				js += x
				overlap += o
				n++
			}
		}
	}
	if n == 0 {
		return 0, 0
	}
	return js / float64(n), overlap / float64(n)
}

func buildLocalProfiles(b Block, maxD int) localProfiles {
	x := localProfiles{P: map[string]profilestability.Profile{}, D: map[string][][]map[string]int{}}
	getD := func(t string) [][]map[string]int {
		z := x.D[t]
		if z == nil {
			z = make([][]map[string]int, 2)
			for side := range z {
				z[side] = make([]map[string]int, maxD)
				for d := range z[side] {
					z[side][d] = map[string]int{}
				}
			}
			x.D[t] = z
		}
		return z
	}
	for _, line := range splitLines(b) {
		for i, t := range line {
			p := x.P[t.Text]
			if p.Positions == nil {
				p.Positions = map[int]int{}
				p.Left = map[string]int{}
				p.Right = map[string]int{}
			}
			p.Count++
			p.Positions[t.LineIndex]++
			if i > 0 {
				p.Left[line[i-1].Text]++
			}
			if i+1 < len(line) {
				p.Right[line[i+1].Text]++
			}
			x.P[t.Text] = p
			dmap := getD(t.Text)
			for d := 1; d <= maxD; d++ {
				if i-d >= 0 {
					dmap[0][d-1][line[i-d].Text]++
				}
				if i+d < len(line) {
					dmap[1][d-1][line[i+d].Text]++
				}
			}
		}
	}
	return x
}

func countMap(m map[string]int) int {
	n := 0
	for _, v := range m {
		n += v
	}
	return n
}
func jsOverlap(a, b map[string]int) (float64, float64) {
	ta, tb := countMap(a), countMap(b)
	if ta == 0 || tb == 0 {
		return 0, 0
	}
	// div/o are single running sums fed by every key of the union of a and
	// b, so they are accumulated in sorted key order: summing in map
	// iteration order made this nondeterministic across otherwise
	// byte-identical calls (see determinism_test.go).
	keys := make([]string, 0, len(a)+len(b))
	seen := make(map[string]bool, len(a)+len(b))
	for k := range a {
		seen[k] = true
		keys = append(keys, k)
	}
	for k := range b {
		if !seen[k] {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	div, o := 0., 0.
	for _, k := range keys {
		pa, pb := float64(a[k])/float64(ta), float64(b[k])/float64(tb)
		m := (pa + pb) / 2
		if pa > 0 {
			div += .5 * pa * math.Log(pa/m)
		}
		if pb > 0 {
			div += .5 * pb * math.Log(pb/m)
		}
		o += math.Min(pa, pb)
	}
	return clamp(1 - div/math.Ln2), clamp(o)
}
func distanceSimilarity(p localProfiles, a, b string, maxD int) (float64, float64) {
	var js, o float64
	n := 0
	for side := 0; side < 2; side++ {
		for d := 0; d < maxD; d++ {
			x, y := jsOverlap(p.D[a][side][d], p.D[b][side][d])
			if countMap(p.D[a][side][d]) > 0 && countMap(p.D[b][side][d]) > 0 {
				js += x
				o += y
				n++
			}
		}
	}
	if n == 0 {
		return 0, 0
	}
	return js / float64(n), o / float64(n)
}
func clamp(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

func profileForBlock(c Candidate, b Block, p localProfiles, maxD int) ProfileBlock {
	x := ProfileBlock{CandidateID: c.ID, Family: c.Family, A: c.A, B: c.B, BlockID: b.ID, Currier: b.Currier, Hand: b.Hand, Joint: b.Joint}
	pa, pb := p.P[c.A], p.P[c.B]
	x.CountA, x.CountB = pa.Count, pb.Count
	x.EligiblePrimary = pa.Count >= 10 && pb.Count >= 10
	x.EligibleDescriptive = pa.Count >= 5 && pb.Count >= 5
	if pa.Count > 0 && pb.Count > 0 {
		v := profilestability.Compare(pa, pb)
		x.Position, x.Left, x.Right, x.Similarity = v.PositionSimilarity, v.LeftSimilarity, v.RightSimilarity, v.Similarity
		if p.D[c.A] != nil && p.D[c.B] != nil {
			x.Distance, x.Overlap = distanceSimilarity(p, c.A, c.B, maxD)
		}
	}
	return x
}

func mean(x []float64) float64 {
	if len(x) == 0 {
		return 0
	}
	s := 0.
	for _, v := range x {
		s += v
	}
	return s / float64(len(x))
}
func distribution(x []float64) (avg, median, minv, sd float64) {
	if len(x) == 0 {
		return
	}
	y := append([]float64(nil), x...)
	sort.Float64s(y)
	avg = mean(y)
	median = y[len(y)/2]
	if len(y)%2 == 0 {
		median = (y[len(y)/2-1] + y[len(y)/2]) / 2
	}
	minv = y[0]
	for _, v := range y {
		sd += (v - avg) * (v - avg)
	}
	sd = math.Sqrt(sd / float64(len(y)))
	return
}
func variance(x []float64) float64 { _, _, _, s := distribution(x); return s * s }
func sign(x float64) int {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}

func BH(p []float64) []float64 {
	n := len(p)
	idx := make([]int, n)
	for i := range idx {
		idx[i] = i
	}
	sort.Slice(idx, func(i, j int) bool { return p[idx[i]] < p[idx[j]] })
	q := make([]float64, n)
	next := 1.
	for rank := n; rank >= 1; rank-- {
		i := idx[rank-1]
		v := p[i] * float64(n) / float64(rank)
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

func PermuteWithinBlocks(blocks []Block, seed int64) []Block {
	r := rand.New(rand.NewSource(seed))
	out := make([]Block, len(blocks))
	for i, b := range blocks {
		out[i] = b
		out[i].Tokens = append([]Token(nil), b.Tokens...)
		texts := make([]string, len(b.Tokens))
		for j, t := range b.Tokens {
			texts[j] = t.Text
		}
		r.Shuffle(len(texts), func(a, z int) { texts[a], texts[z] = texts[z], texts[a] })
		for j := range out[i].Tokens {
			out[i].Tokens[j].Text = texts[j]
		}
	}
	return out
}

func Classify(s RelationSummary, crossCurrier, crossHand bool) string {
	return ClassifyDetailed(s, s.CurrierClasses == 1, crossCurrier, s.Hands == 1, crossHand)
}
func ClassifyDetailed(s RelationSummary, withinCurrier, crossCurrier, withinHand, crossHand bool) string {
	stable := s.SignConsistency >= .75 || s.ProfileMedian >= .7
	if s.EligibleBlocks >= 3 && s.JointClasses >= 2 && stable && s.TransferSuccess >= .67 && (crossCurrier || crossHand) {
		return "UNIVERSAL"
	}
	if s.EligibleBlocks >= 2 && stable && withinCurrier && !crossCurrier {
		return "CURRIER_SPECIFIC"
	}
	if s.EligibleBlocks >= 2 && stable && withinHand && !crossHand {
		return "HAND_SPECIFIC"
	}
	if s.EligibleBlocks > 0 && (s.TransferSuccess < .5 || s.PhysicalBlocks <= 1) {
		return "BLOCK_SPECIFIC"
	}
	return "WEAK"
}

// ClassifyGeneric is generic mode's replacement for ClassifyDetailed: it
// reasons about the single deterministic resampling Group dimension only,
// and its vocabulary deliberately never borrows "CURRIER_SPECIFIC"/
// "HAND_SPECIFIC"/"UNIVERSAL" - those claim a real manuscript covariate,
// which a generic corpus does not have.
func ClassifyGeneric(s RelationSummary, withinGroup, crossGroup bool) string {
	stable := s.SignConsistency >= .75 || s.ProfileMedian >= .7
	if s.EligibleBlocks >= 3 && s.JointClasses >= 2 && stable && s.TransferSuccess >= .67 && crossGroup {
		return "GROUP_CONSISTENT"
	}
	if s.EligibleBlocks >= 2 && stable && withinGroup && !crossGroup {
		return "GROUP_LIMITED"
	}
	if s.EligibleBlocks > 0 && (s.TransferSuccess < .5 || s.PhysicalBlocks <= 1) {
		return "BLOCK_SPECIFIC"
	}
	return "WEAK"
}

func RuleLike(s RelationSummary) bool {
	return s.Family == "directional" && s.EligibleBlocks >= 3 && s.JointClasses >= 2 && s.SignConsistency >= .75 && s.MedianEnrichment > 1 && s.TransferSuccess >= .67 && s.FDRQ <= .05
}
