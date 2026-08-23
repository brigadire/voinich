package fingerprintv2

import (
	"math"
	"math/rand"
	"sort"

	"zcore.dev/voinich/internal/tokenrepetition"
)

type bareMetrics struct {
	graph       editGraph
	lp1         LP1Result
	candidates  map[string]bool
	ef1         EF1Result
	clustering  float64
	triangles   int
	paths3      int
	cycles4     int
	spearman    float64
	prefixNMI   float64
	suffixNMI   float64
	attachmentN int
	excludedN   int
}

func analyzeBare(c corpus, cfg Config) bareMetrics {
	g := buildGraph(c)
	lp, candidates := lp1(c, g, cfg.MinRuleSupport)
	cluster, triangles, paths3, cycles4 := graphMotifs(g)
	return bareMetrics{
		graph: g, lp1: lp, candidates: candidates, ef1: ef1(g),
		clustering: cluster, triangles: triangles, paths3: paths3, cycles4: cycles4,
		spearman:  degreeFrequencySpearman(g, frequencies(c)),
		prefixNMI: attachmentMI(c, true), suffixNMI: attachmentMI(c, false),
		attachmentN: eligibleAttachmentCount(c), excludedN: len(c.records) - eligibleAttachmentCount(c),
	}
}

func attachmentValues(c corpus, prefix bool) (core, affix []string) {
	for _, token := range vocabulary(c) {
		glyph := glyphByToken(c)[token]
		if len(glyph) < 3 {
			continue
		}
		core = append(core, glyphKey(glyph[1:len(glyph)-1]))
		if prefix {
			affix = append(affix, glyph[0])
		} else {
			affix = append(affix, glyph[len(glyph)-1])
		}
	}
	return core, affix
}

func attachmentMI(c corpus, prefix bool) float64 {
	core, affix := attachmentValues(c, prefix)
	return normalizedMI(core, affix)
}

func eligibleAttachmentCount(c corpus) int {
	n := 0
	for _, token := range vocabulary(c) {
		if len(glyphByToken(c)[token]) >= 3 {
			n++
		}
	}
	return n
}

func degreeFrequencySpearman(g editGraph, freq map[string]int) float64 {
	x, y, names := make([]float64, 0, len(g.nodes)), make([]float64, 0, len(g.nodes)), append([]string(nil), g.nodes...)
	for _, n := range names {
		x = append(x, float64(len(g.adj[n])))
		y = append(y, math.Log1p(float64(freq[n])))
	}
	return spearman(x, y, names)
}

func randomPairGini(c corpus, g editGraph, frequencyMatched bool, rng *rand.Rand) float64 {
	glyphs, freq := glyphByToken(c), frequencies(c)
	byLength, byLengthFrequency := map[int][]string{}, map[string][]string{}
	for _, n := range g.nodes {
		length := len(glyphs[n])
		byLength[length] = append(byLength[length], n)
		key := pairBucket(length, frequencyBin(freq[n]))
		byLengthFrequency[key] = append(byLengthFrequency[key], n)
	}
	for _, v := range byLength {
		sort.Strings(v)
	}
	for _, v := range byLengthFrequency {
		sort.Strings(v)
	}
	counts := map[string]int{}
	for _, edge := range g.edgeList() {
		a, b := edge[0], edge[1]
		pick := func(token string) string {
			pool := byLength[len(glyphs[token])]
			if frequencyMatched {
				if matched := byLengthFrequency[pairBucket(len(glyphs[token]), frequencyBin(freq[token]))]; len(matched) > 0 {
					pool = matched
				}
			}
			if len(pool) == 0 {
				return ""
			}
			return pool[rng.Intn(len(pool))]
		}
		x, y := pick(a), pick(b)
		if x == "" || y == "" || x == y || tokenrepetition.LevenshteinGlyphs(glyphs[x], glyphs[y]) != 1 {
			continue
		}
		for _, pair := range [][2]string{{x, y}, {y, x}} {
			if rule, ok := ruleFor(pair[0], pair[1], glyphs); ok {
				counts[rule]++
			}
		}
	}
	values := make([]int, 0, len(counts))
	for _, key := range orderedKeys(counts) {
		values = append(values, counts[key])
	}
	return gini(values)
}

func frequencyBin(v int) int {
	b := 0
	for v > 1 {
		v >>= 1
		b++
	}
	return b
}
func pairBucket(length, bin int) string { return string(rune(length)) + "\x00" + string(rune(bin)) }

func attachmentPermutation(c corpus, repetitions int, rng *rand.Rand) (prefix, suffix []float64) {
	core, p, s := attachmentTriples(c)
	if len(core) == 0 {
		return make([]float64, repetitions), make([]float64, repetitions)
	}
	byLength := map[int][]int{}
	for i, token := range vocabulary(c) {
		if glyph := glyphByToken(c)[token]; len(glyph) >= 3 {
			byLength[len(glyph)] = append(byLength[len(glyph)], i)
		}
	}
	// triples are occurrence-ordered only among eligible records; construct
	// matching buckets in that index space to preserve the length marginal.
	eligibleIndex := map[int]int{}
	j := 0
	for i, token := range vocabulary(c) {
		if len(glyphByToken(c)[token]) >= 3 {
			eligibleIndex[i] = j
			j++
		}
	}
	buckets := make([][]int, 0, len(byLength))
	for _, source := range byLength {
		b := make([]int, len(source))
		for i, pos := range source {
			b[i] = eligibleIndex[pos]
		}
		buckets = append(buckets, b)
	}
	sort.Slice(buckets, func(i, j int) bool { return buckets[i][0] < buckets[j][0] })
	prefix, suffix = make([]float64, repetitions), make([]float64, repetitions)
	for r := 0; r < repetitions; r++ {
		pp, ss := append([]string(nil), p...), append([]string(nil), s...)
		for _, bucket := range buckets {
			perm := rng.Perm(len(bucket))
			old := make([]string, len(bucket))
			for i, idx := range bucket {
				old[i] = pp[idx]
			}
			for i, idx := range bucket {
				pp[idx] = old[perm[i]]
			}
			perm = rng.Perm(len(bucket))
			for i, idx := range bucket {
				old[i] = ss[idx]
			}
			for i, idx := range bucket {
				ss[idx] = old[perm[i]]
			}
		}
		prefix[r], suffix[r] = normalizedMI(core, pp), normalizedMI(core, ss)
	}
	return prefix, suffix
}

func attachmentTriples(c corpus) (core, prefix, suffix []string) {
	for _, token := range vocabulary(c) {
		glyph := glyphByToken(c)[token]
		if len(glyph) < 3 {
			continue
		}
		core = append(core, glyphKey(glyph[1:len(glyph)-1]))
		prefix = append(prefix, glyph[0])
		suffix = append(suffix, glyph[len(glyph)-1])
	}
	return
}

func ef2WithControl(b bareMetrics, cfg Config, rng *rand.Rand) EF2Result {
	null := make([]float64, cfg.Repetitions)
	for i := range null {
		swapped := degreePreservingSwap(b.graph, cfg.GraphSwaps*max(1, len(b.graph.edgeList())), rng)
		null[i], _, _, _ = graphMotifs(swapped)
	}
	return EF2Result{
		GlobalClustering: b.clustering, Triangles: b.triangles, Paths3: b.paths3, Cycles4: b.cycles4,
		ConfigurationTest:  nullTest("ef2/degree-preserving-edge-swaps", "degree-preserving edge swaps", b.clustering, null),
		ControlDescription: "Simple-graph double-edge swaps preserve the exact observed degree sequence; each replicate attempts graph_swaps × edge_count seeded swaps.",
	}
}

func ef2Bare(b bareMetrics) EF2Result {
	return EF2Result{
		GlobalClustering: b.clustering, Triangles: b.triangles, Paths3: b.paths3, Cycles4: b.cycles4,
		ControlDescription: "Configuration control is computed for the observed corpus only.",
	}
}

func ef3WithControl(b bareMetrics, c corpus, cfg Config, rng *rand.Rand) EF3Result {
	freq := frequencies(c)
	values := make([]int, 0, len(b.graph.nodes))
	for _, n := range b.graph.nodes {
		values = append(values, freq[n])
	}
	null := make([]float64, cfg.Repetitions)
	for r := range null {
		p := rng.Perm(len(values))
		shuffled := map[string]int{}
		for i, n := range b.graph.nodes {
			shuffled[n] = values[p[i]]
		}
		null[r] = math.Abs(degreeFrequencySpearman(b.graph, shuffled))
	}
	return EF3Result{
		SpearmanDegreeLogFrequency: b.spearman,
		FrequencyControl:           nullTest("ef3/c-freq", "C-FREQ frequency-label permutation", math.Abs(b.spearman), null),
	}
}

func ef3Bare(b bareMetrics) EF3Result {
	return EF3Result{SpearmanDegreeLogFrequency: b.spearman}
}
