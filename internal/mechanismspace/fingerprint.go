package mechanismspace

import (
	"math"
	"math/rand"
	"sort"

	"zcore.dev/voinich/internal/characterentropy"
	"zcore.dev/voinich/internal/evaglyph"
	"zcore.dev/voinich/internal/lineregime"
	"zcore.dev/voinich/internal/localregimetopology"
	"zcore.dev/voinich/internal/tokenformation"
	"zcore.dev/voinich/internal/tokenrepetition"
	"zcore.dev/voinich/internal/tokentransition"
)

// Fingerprint holds one value per task66 section 9 metric family (A-G),
// computed with the same authoritative Task58-65 primitives used to
// characterize Voynich, never a reinterpretation of them (section 8).
// Status records which fields could actually be computed (task66 section
// 10 / task52 semantics): VALUE, NOT_COMPUTED, NOT_APPLICABLE.
type Fingerprint struct {
	// A. TOKEN_ORDER (Task58)
	TokenOrderBits float64
	EdgeOrderBits  float64

	// B. POSITIONAL_STRUCTURE (Task59)
	PositionalWeightedEntropy float64
	HighFreqSpecialists       float64

	// C. REPETITION_EDIT_GEOMETRY (Task60)
	ExactAdjacentRepeatRate float64
	GiantComponentFraction  float64

	// D. CHARACTER_ENTROPY (Task61)
	H1, H2, H3, H4 float64

	// E. TOKEN_FORMATION (Task62)
	PositionGainBits float64
	OrderGainBits    float64

	// F. LOCAL_TRANSITION (Task63)
	AdjacentNearRate  float64
	ResidualAdjacency float64

	// G. LOCAL_REGIME_TOPOLOGY (Task64-65)
	CorrelationLengthTokens float64
	ClusterStability        float64
	WithinClusterDrift      float64
	LineBoundaryDelta       float64
	Topology                string

	Status map[string]string
}

const (
	StatusValue          = "VALUE"
	StatusCompletedEmpty = "COMPLETED_EMPTY"
	StatusNotApplicable  = "NOT_APPLICABLE"
	StatusNotComputed    = "NOT_COMPUTED"
	StatusMissing        = "MISSING_ARTIFACT"
	StatusFailed         = "FAILED_INVALID"
)

// FingerprintOptions controls the cost of extraction (task66 section 72's
// screening/development/final tiers): a screening pass uses a cheap
// subset (fewer shuffle replicates, no giant-component/topology pass); a
// development/final pass computes every family.
type FingerprintOptions struct {
	Full          bool // include C's giant component and G's topology (expensive)
	ShuffleReps   int  // Task58/59 null replicates
	VocabularyCap int  // Task58 top-N token cap
	Seed          int64
}

func DefaultScreeningOptions(seed int64) FingerprintOptions {
	return FingerprintOptions{Full: false, ShuffleReps: 5, VocabularyCap: 500, Seed: seed}
}
func DefaultFullOptions(seed int64) FingerprintOptions {
	return FingerprintOptions{Full: true, ShuffleReps: 30, VocabularyCap: 2000, Seed: seed}
}

// ComputeFingerprint extracts the task66 section 9 fingerprint from an
// arbitrary token corpus (Voynich or a transformed plaintext candidate):
// every metric is computed by the same generic primitive Task58-65 uses
// for Voynich, never a bespoke recomputation (task66 section 8).
func ComputeFingerprint(tokens [][]string, lines []int, opt FingerprintOptions) Fingerprint {
	fp := Fingerprint{Status: map[string]string{}}
	if len(tokens) < 4 {
		fp.Status["ALL"] = StatusCompletedEmpty
		return fp
	}

	tokenOrder, edgeOrder := taskOrderMI(tokens, lines, opt)
	fp.TokenOrderBits, fp.EdgeOrderBits = tokenOrder, edgeOrder
	fp.Status["TOKEN_ORDER"] = StatusValue

	fp.PositionalWeightedEntropy, fp.HighFreqSpecialists = positionalSpecialization(tokens, opt)
	fp.Status["POSITIONAL_STRUCTURE"] = StatusValue

	flat := joinedForms(tokens)
	adj := tokenrepetition.AdjacentRepetition(flat, lines, 0)
	fp.ExactAdjacentRepeatRate = adj.R2
	fp.Status["REPETITION_EXACT"] = StatusValue
	var giant map[string]bool
	if opt.Full {
		var vocabSize int
		giant, vocabSize = boundedGiantSet(tokens, giantSetVocabCap, opt.Seed+2)
		fp.GiantComponentFraction = giantComponentFraction(giant, vocabSize)
		fp.Status["REPETITION_EDIT"] = StatusValue
	} else {
		fp.Status["REPETITION_EDIT"] = StatusNotComputed
	}

	fp.H1 = orderEntropy(tokens, lines, 0)
	fp.H2 = orderEntropy(tokens, lines, 1)
	fp.H3 = orderEntropy(tokens, lines, 2)
	fp.H4 = orderEntropy(tokens, lines, 3)
	fp.Status["CHARACTER_ENTROPY"] = StatusValue

	fp.PositionGainBits, fp.OrderGainBits = tokenFormationGains(tokens, lines)
	fp.Status["TOKEN_FORMATION"] = StatusValue

	fp.AdjacentNearRate, fp.ResidualAdjacency = localTransition(tokens)
	fp.Status["LOCAL_TRANSITION"] = StatusValue

	if opt.Full {
		topo := computeTopologyWithGiant(tokens, lines, giant)
		fp.CorrelationLengthTokens = topo.CorrelationLengthTokens
		fp.ClusterStability = topo.ClusterStability
		fp.WithinClusterDrift = topo.WithinClusterDrift
		fp.LineBoundaryDelta = topo.LineBoundaryDelta
		fp.Topology = topo.Topology
		fp.Status["LOCAL_TOPOLOGY"] = StatusValue
	} else {
		fp.Topology = "NOT_COMPUTED"
		fp.Status["LOCAL_TOPOLOGY"] = StatusNotComputed
	}
	return fp
}

// deltasBetween returns one normalized |a-b| value per fingerprint family
// that both sides actually computed, for use by sensitivity/robustness
// comparisons (task66 sections 65-67).
func deltasBetween(a, b Fingerprint) []float64 {
	rel := func(x, y float64) float64 {
		d := math.Abs(x - y)
		scale := math.Abs(y)
		if scale < 1e-9 {
			scale = 1e-9
		}
		return d / scale
	}
	out := []float64{
		rel(a.TokenOrderBits, b.TokenOrderBits),
		rel(a.PositionalWeightedEntropy, b.PositionalWeightedEntropy),
		rel(a.ExactAdjacentRepeatRate, b.ExactAdjacentRepeatRate),
		rel(a.H2, b.H2),
		rel(a.OrderGainBits, b.OrderGainBits),
		rel(a.AdjacentNearRate, b.AdjacentNearRate),
	}
	if a.Status["LOCAL_TOPOLOGY"] == StatusValue && b.Status["LOCAL_TOPOLOGY"] == StatusValue {
		out = append(out, rel(a.CorrelationLengthTokens, b.CorrelationLengthTokens))
	}
	return out
}

func joinedForms(tokens [][]string) []string {
	out := make([]string, len(tokens))
	for i, t := range tokens {
		out[i] = joinTokens(t)
	}
	return out
}

// taskOrderMI replicates the authoritative Task58 metric exactly: token-
// identity MI (rare types capped to "<other>") and glyph-edge MI, each
// corrected by a within-line shuffle null (independent/rozanova-temerev's
// analyse(), reused rather than reinterpreted per task66 section 8/44).
func taskOrderMI(tokens [][]string, lines []int, opt FingerprintOptions) (tokenBits, edgeBits float64) {
	flat := joinedForms(tokens)
	lineIDs := lines
	if len(lineIDs) != len(flat) {
		lineIDs = make([]int, len(flat)) // no line structure: treat as one line
	}
	freq := map[string]int{}
	for _, t := range flat {
		freq[t]++
	}
	keys := make([]string, 0, len(freq))
	for k := range freq {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if freq[keys[i]] != freq[keys[j]] {
			return freq[keys[i]] > freq[keys[j]]
		}
		return keys[i] < keys[j]
	})
	cap := opt.VocabularyCap
	if cap <= 0 || cap > len(keys) {
		cap = len(keys)
	}
	keep := map[string]bool{}
	for i := 0; i < cap; i++ {
		keep[keys[i]] = true
	}
	calc := func(order []int) (float64, float64) {
		var tokL, tokR, edgeL, edgeR []string
		for k := 0; k < len(order)-1; k++ {
			a, b := order[k], order[k+1]
			if lineIDs[a] != lineIDs[b] {
				continue
			}
			la, lb := flat[a], flat[b]
			if !keep[la] {
				la = "<other>"
			}
			if !keep[lb] {
				lb = "<other>"
			}
			tokL = append(tokL, la)
			tokR = append(tokR, lb)
			if len(tokens[a]) > 0 && len(tokens[b]) > 0 {
				edgeL = append(edgeL, tokens[a][len(tokens[a])-1])
				edgeR = append(edgeR, tokens[b][0])
			}
		}
		e := 0.0
		if len(edgeL) > 0 {
			e = evaglyph.MI(edgeL, edgeR)
		}
		return evaglyph.MI(tokL, tokR), e
	}
	order := make([]int, len(flat))
	for i := range order {
		order[i] = i
	}
	to, eo := calc(order)
	r := rand.New(rand.NewSource(opt.Seed + 1))
	reps := opt.ShuffleReps
	if reps < 1 {
		reps = 1
	}
	var tn, en []float64
	for s := 0; s < reps; s++ {
		p := withinLineShuffle(order, lineIDs, r)
		a, b := calc(p)
		tn = append(tn, a)
		en = append(en, b)
	}
	return to - meanF(tn), eo - meanF(en)
}

func withinLineShuffle(order, lineIDs []int, r *rand.Rand) []int {
	p := append([]int(nil), order...)
	start := 0
	for start < len(p) {
		end := start
		for end < len(p) && lineIDs[end] == lineIDs[start] {
			end++
		}
		for i := end - 1; i > start; i-- {
			j := start + r.Intn(i-start+1)
			p[i], p[j] = p[j], p[i]
		}
		start = end
	}
	return p
}

func meanF(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	s := 0.0
	for _, x := range v {
		s += x
	}
	return s / float64(len(v))
}

type posCount struct{ i, m, f, s int }

// positionalSpecialization replicates independent/glyph-position-analyze's
// weighted-entropy statistic (task59): the average, occurrence-weighted
// Shannon entropy (natural log) of each glyph's INITIAL/MEDIAL/FINAL(/
// SINGLETON) distribution. Low values mean glyphs specialize strongly by
// within-token position.
func positionalSpecialization(tokens [][]string, _ FingerprintOptions) (weightedEntropy float64, highFreqSpecialists float64) {
	counts := map[string]*posCount{}
	for _, t := range tokens {
		for i, g := range t {
			c := counts[g]
			if c == nil {
				c = &posCount{}
				counts[g] = c
			}
			c.i2(evaglyph.Classify(len(t), i))
		}
	}
	totalOcc, sumH := 0.0, 0.0
	nHigh := 0
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, g := range keys {
		c := counts[g]
		n := float64(c.i + c.m + c.f + c.s)
		if n == 0 {
			continue
		}
		vals := []float64{float64(c.i), float64(c.m), float64(c.f)}
		if c.s > 0 {
			vals = append(vals, float64(c.s))
		}
		h, share := 0.0, 0.0
		for _, v := range vals {
			if v > share {
				share = v
			}
			if v > 0 {
				p := v / n
				h -= p * math.Log(p)
			}
		}
		sumH += h * n
		totalOcc += n
		if n >= 100 && share/n >= 0.95 {
			nHigh++
		}
	}
	if totalOcc == 0 {
		return 0, 0
	}
	return sumH / totalOcc, float64(nHigh)
}

func (c *posCount) i2(class string) {
	switch class {
	case "INITIAL":
		c.i++
	case "MEDIAL":
		c.m++
	case "FINAL":
		c.f++
	case "SINGLETON":
		c.s++
	}
}

// giantSetVocabCap bounds the d<=1 giant-component computation's cost.
// internal/lineregime.BuildGiantSet is quadratic within same-length
// vocabulary buckets; that is fine at Voynich/natural-language scale
// (types spread over many token lengths), but a few grid mechanisms
// (uniform-length STREAM grouping) concentrate almost the entire
// vocabulary into one or two length buckets, which would otherwise make
// a single fingerprint pass take tens of seconds. Capping is the same
// performance escape hatch Task58 already uses for its own MI
// computation (vocabulary-cap, default 2000): a bounded, seeded random
// sample of types stands in for the full vocabulary.
const giantSetVocabCap = 3000

// boundedGiantSet computes BuildGiantSet's authoritative giant-component
// membership on at most cap distinct token types, chosen by a seeded
// random sample when the corpus vocabulary is larger. sampleVocabSize is
// the number of types the giant set was actually computed over (either
// the full vocabulary, or cap), so giantComponentFraction can divide by
// the same population the sample was drawn from rather than silently
// biasing the fraction downward.
func boundedGiantSet(tokens [][]string, cap int, seed int64) (giant map[string]bool, sampleVocabSize int) {
	seen := map[string][]string{}
	for _, t := range tokens {
		seen[joinTokens(t)] = t
	}
	if len(seen) <= cap {
		return lineregime.BuildGiantSet(tokens), len(seen)
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	r := rand.New(rand.NewSource(seed))
	r.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })
	sample := make([][]string, cap)
	for i := 0; i < cap; i++ {
		sample[i] = seen[keys[i]]
	}
	return lineregime.BuildGiantSet(sample), cap
}

// giantComponentFraction is the task60/64 d<=1 giant-component metric
// (reused via internal/lineregime.BuildGiantSet, the authoritative
// implementation Task64/65 already share), expressed as a fraction of
// the vocabulary the giant set was actually computed over (see
// boundedGiantSet).
func giantComponentFraction(giant map[string]bool, vocabSize int) float64 {
	if vocabSize == 0 {
		return 0
	}
	return float64(len(giant)) / float64(vocabSize)
}

// orderEntropy is Task61's h_{order+1}: the plug-in conditional character
// entropy of order `order` in TOKEN_BOUNDARY mode with line resets, using
// the authoritative internal/characterentropy estimator directly.
func orderEntropy(tokens [][]string, lines []int, order int) float64 {
	est := characterentropy.Entropy(tokens, lines, characterentropy.TokenBoundary, order, true)
	if est.Status != "OK" {
		return math.NaN()
	}
	return est.H
}

// tokenFormationGains characterizes Task62's token-formation fingerprint
// as the held-out cross-entropy gain from adding position-dependence and
// from adding first-order sequential dependence, comparable to the
// authoritative MODEL_HELDOUT_FIT.tsv IID/POSITION_IID/MARKOV_1 rows. It
// never uses a Voynich-trained model as a transformation (task66 sections
// 14, 48): here tokenformation.Fit only characterizes whatever corpus it
// is given.
func tokenFormationGains(tokens [][]string, lines []int) (positionGain, orderGain float64) {
	n := len(tokens)
	if n < 20 {
		return 0, 0
	}
	split := n * 7 / 10
	train := tokenformation.Corpus{Tokens: tokens[:split], Lines: subLines(lines, 0, split)}
	test := tokenformation.Corpus{Tokens: tokens[split:], Lines: subLines(lines, split, n)}
	iid := tokenformation.Fit(train, tokenformation.IID, 0.1)
	pos := tokenformation.Fit(train, tokenformation.PosIID, 0.1)
	mk1 := tokenformation.Fit(train, tokenformation.Markov1, 0.1)
	ceIID := iid.CrossEntropy(test.Tokens)
	cePos := pos.CrossEntropy(test.Tokens)
	ceMk1 := mk1.CrossEntropy(test.Tokens)
	return ceIID - cePos, ceIID - ceMk1
}

func subLines(lines []int, a, b int) []int {
	if len(lines) < b {
		return nil
	}
	return lines[a:b]
}

// localTransition is Task63's near-rate metric (evaluation delegated to
// internal/tokentransition.Analyze, which itself delegates edit semantics
// to Task60): adjacent (separation 1) near-rate and its residual over the
// matched separation-10 baseline.
func localTransition(tokens [][]string) (adjacentNearRate, residual float64) {
	near1, n1 := nearRateAtSeparation(tokens, 1)
	near10, n10 := nearRateAtSeparation(tokens, 10)
	if n1 == 0 {
		return 0, 0
	}
	adjacentNearRate = near1
	if n10 > 0 {
		residual = near1 - near10
	}
	return
}

func nearRateAtSeparation(tokens [][]string, sep int) (rate float64, pairs int) {
	near := 0
	for i := 0; i+sep < len(tokens); i++ {
		p := tokentransition.Analyze(tokens[i], tokens[i+sep])
		pairs++
		if tokentransition.IsNear(p) {
			near++
		}
	}
	if pairs == 0 {
		return 0, 0
	}
	return float64(near) / float64(pairs), pairs
}

// Topology is task66 section 52's simplified but data-driven analogue of
// Task65's classifier, built from the same shared primitives
// (localregimetopology.BuildWindows/ScanChangePoints/KMedoids and
// lineregime.ComputeProfile/ProfileDistance): a full replication of
// Task65's bootstrap/permutation apparatus is not repeated per grid point
// (task66 section 72 explicitly allows computing the needed fields
// directly instead).
type Topology struct {
	CorrelationLengthTokens float64
	ClusterStability        float64
	WithinClusterDrift      float64
	LineBoundaryDelta       float64
	Topology                string
}

func ComputeTopology(tokens [][]string, lines []int) Topology {
	return computeTopologyWithGiant(tokens, lines, lineregime.BuildGiantSet(tokens))
}

func computeTopologyWithGiant(tokens [][]string, lines []int, giant map[string]bool) Topology {
	const w, step = 20, 5
	records := make([]localregimetopology.TokenRecord, len(tokens))
	for i, t := range tokens {
		line := i
		if i < len(lines) {
			line = lines[i]
		}
		records[i] = localregimetopology.TokenRecord{GlobalIndex: i, Line: line, Glyphs: t}
	}
	windows := localregimetopology.BuildWindows(records, w, step)
	if len(windows) < 8 {
		return Topology{Topology: "UNRESOLVED"}
	}
	profiles := make([]lineregime.Profile, len(windows))
	for i, win := range windows {
		profiles[i] = lineregime.ComputeProfile(win.Tokens, giant)
	}
	nullMean := meanPairwiseDistance(profiles, 200)
	decay1 := meanLagDistance(profiles, 1)
	localStructure := decay1 < 0.9*nullMean

	corrLen := math.NaN()
	for lag := 1; lag < len(profiles)/2 && lag < 40; lag++ {
		if meanLagDistance(profiles, lag) >= 0.5*nullMean {
			corrLen = float64(lag * step)
			break
		}
	}

	vecs := make([][]float64, len(profiles))
	for i, p := range profiles {
		vecs[i] = []float64{p.MeanLen, p.GiantFrac, p.TopInit, p.TopFinal, p.TypeEnt}
	}
	all := make([]int, len(vecs))
	for i := range all {
		all[i] = i
	}
	std := localregimetopology.StandardizeColumns(vecs, all)
	assign, _ := localregimetopology.KMedoids(std, 5, 30)
	stability := clusterStabilityScore(std, assign, 5)
	clusteringSupported := stability > 0.3

	withinDrift := withinClusterDriftScore(profiles, assign)

	lineDelta := 0.0
	if len(lines) == len(tokens) {
		lineDelta = boundaryDiscontinuity(tokens, lines, giant, decay1)
	}

	topo := "UNRESOLVED"
	switch {
	case clusteringSupported && withinDrift > 0.05:
		topo = "MIXED_DRIFT_AND_STATES"
	case clusteringSupported:
		topo = "DISCRETE_REGIMES"
	case localStructure:
		topo = "CONTINUOUS_DRIFT"
	case !math.IsNaN(decay1):
		topo = "STATIONARY"
	}
	return Topology{CorrelationLengthTokens: corrLen, ClusterStability: stability, WithinClusterDrift: withinDrift, LineBoundaryDelta: lineDelta, Topology: topo}
}

func meanLagDistance(profiles []lineregime.Profile, lag int) float64 {
	if lag >= len(profiles) {
		return math.NaN()
	}
	sum, n := 0.0, 0
	for i := 0; i+lag < len(profiles); i++ {
		sum += lineregime.ProfileDistance(profiles[i], profiles[i+lag])
		n++
	}
	if n == 0 {
		return math.NaN()
	}
	return sum / float64(n)
}

func meanPairwiseDistance(profiles []lineregime.Profile, samplePairs int) float64 {
	r := rand.New(rand.NewSource(42))
	n := len(profiles)
	if n < 2 {
		return 0
	}
	sum, k := 0.0, 0
	for k < samplePairs {
		i, j := r.Intn(n), r.Intn(n)
		if i == j {
			continue
		}
		sum += lineregime.ProfileDistance(profiles[i], profiles[j])
		k++
	}
	return sum / float64(samplePairs)
}

func clusterStabilityScore(vecs [][]float64, assign []int, k int) float64 {
	n := len(vecs)
	if n < 2*k {
		return 0
	}
	half := n / 2
	first := localregimetopology.StandardizeColumns(vecs[:half], allIdx(half))
	second := localregimetopology.StandardizeColumns(vecs[half:], allIdx(n-half))
	a1, _ := localregimetopology.KMedoids(first, k, 30)
	a2, _ := localregimetopology.KMedoids(second, k, 30)
	agree1 := coClusterAgreement(assign[:half], a1)
	agree2 := coClusterAgreement(assign[half:], a2)
	return (agree1 + agree2) / 2
}

func allIdx(n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = i
	}
	return out
}

// coClusterAgreement is the fraction of index pairs that are co-clustered
// (or co-separated) identically in both assignments - a simple Rand-index
// style stability proxy, sampled to stay cheap on large windows.
func coClusterAgreement(a, b []int) float64 {
	n := len(a)
	if n != len(b) || n < 2 {
		return 0
	}
	r := rand.New(rand.NewSource(7))
	samples := 400
	agree := 0
	for s := 0; s < samples; s++ {
		i, j := r.Intn(n), r.Intn(n)
		if i == j {
			continue
		}
		sameA := a[i] == a[j]
		sameB := b[i] == b[j]
		if sameA == sameB {
			agree++
		}
	}
	return float64(agree) / float64(samples)
}

func withinClusterDriftScore(profiles []lineregime.Profile, assign []int) float64 {
	groups := map[int][]int{}
	for i, a := range assign {
		groups[a] = append(groups[a], i)
	}
	var scores []float64
	for _, idxs := range groups {
		if len(idxs) < 6 {
			continue
		}
		sort.Ints(idxs)
		near, far := 0.0, 0.0
		nn, nf := 0, 0
		for a := 0; a < len(idxs); a++ {
			for b := a + 1; b < len(idxs); b++ {
				d := lineregime.ProfileDistance(profiles[idxs[a]], profiles[idxs[b]])
				if b-a <= 2 {
					near += d
					nn++
				} else {
					far += d
					nf++
				}
			}
		}
		if nn > 0 && nf > 0 {
			scores = append(scores, far/float64(nf)-near/float64(nn))
		}
	}
	return meanF(scores)
}

// boundaryDiscontinuity is task66 section 50/53's line-boundary special-
// ness check: observed profile distance across a natural line boundary
// minus what the decay curve alone would predict at that separation
// (task65's BOUNDARY_DISCONTINUITY.tsv, computed directly rather than via
// the full pipeline per section 72).
func boundaryDiscontinuity(tokens [][]string, lines []int, giant map[string]bool, expected float64) float64 {
	var obs []float64
	const half = 5
	for i := half; i+half < len(tokens); i++ {
		if lines[i-1] == lines[i] {
			continue // only evaluate at an actual line change
		}
		left := tokens[i-half : i]
		right := tokens[i : i+half]
		pl := lineregime.ComputeProfile(left, giant)
		pr := lineregime.ComputeProfile(right, giant)
		obs = append(obs, lineregime.ProfileDistance(pl, pr))
	}
	if len(obs) == 0 {
		return 0
	}
	return meanF(obs) - expected
}
