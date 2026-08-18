package transitionnetwork

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

const alpha = 0.5

type blockData struct {
	Block        Block
	Counts       map[string]int
	Eligible     map[string]bool
	Edges        map[EdgeKey]int
	Opp          map[string]int
	Destinations []string
}

type analysis struct {
	Tokens                     []Token
	Blocks                     []Block
	Data                       []blockData
	Counts                     map[string]int
	Vocab                      []string
	Edges                      []EdgeKey
	BlockEffects               []BlockStats
	ByEdge                     map[EdgeKey][]BlockStats
	Summaries                  []*EdgeSummary
	Stability                  []ProfileStability
	Entropies                  []EntropyRow
	OutgoingRows, IncomingRows []profileRow
	Predictions                []PredictionRow
	ModelOrder                 []ModelOrderRow
	MetadataTransfer           []TransferRow
	GraphSimilarity            []GraphSimilarityRow
}
type profileRow struct {
	Token, Block, Direction, Destination string
	Count                                int
	Probability, Log2Enrichment          float64
}

func buildData(tokens []Token, blocks []Block, minCount int) (map[string]int, []string, []EdgeKey, []blockData) {
	counts := map[string]int{}
	for _, t := range tokens {
		counts[t.Text]++
	}
	var vocab []string
	eligible := map[string]bool{}
	for t, n := range counts {
		if n >= minCount {
			eligible[t] = true
			vocab = append(vocab, t)
		}
	}
	sort.Strings(vocab)
	globalEdges := map[EdgeKey]int{}
	data := make([]blockData, len(blocks))
	for z, b := range blocks {
		d := blockData{Block: b, Counts: map[string]int{}, Eligible: eligible, Edges: map[EdgeKey]int{}, Opp: map[string]int{}}
		for _, t := range b.Tokens {
			d.Counts[t.Text]++
		}
		for i := 0; i+1 < len(b.Tokens); i++ {
			a, c := b.Tokens[i].Text, b.Tokens[i+1].Text
			if eligible[a] {
				d.Opp[a]++
				d.Destinations = append(d.Destinations, c)
			}
			if eligible[a] && eligible[c] {
				e := EdgeKey{a, c}
				d.Edges[e]++
				globalEdges[e]++
			}
		}
		data[z] = d
	}
	edges := make([]EdgeKey, 0, len(globalEdges))
	for e := range globalEdges {
		edges = append(edges, e)
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].Source == edges[j].Source {
			return edges[i].Target < edges[j].Target
		}
		return edges[i].Source < edges[j].Source
	})
	return counts, vocab, edges, data
}

func effect(d blockData, e EdgeKey, V int) BlockStats {
	n := len(d.Block.Tokens)
	pc := (float64(d.Edges[e]) + alpha) / (float64(d.Opp[e.Source]) + alpha*float64(V))
	pb := (float64(d.Counts[e.Target]) + alpha) / (float64(n) + alpha*float64(V))
	en := pc / pb
	return BlockStats{d.Block.ID, d.Block.Currier, d.Block.Hand, d.Block.Joint, e.Source, e.Target, d.Counts[e.Source], d.Counts[e.Target], d.Edges[e], d.Opp[e.Source], n, pc, pb, en, math.Log2(en)}
}

func summarizeEdges(a *analysis, minBlock int) {
	a.ByEdge = map[EdgeKey][]BlockStats{}
	for _, e := range a.Edges {
		for _, d := range a.Data {
			if d.Opp[e.Source] < minBlock {
				continue
			}
			x := effect(d, e, len(a.Vocab))
			a.BlockEffects = append(a.BlockEffects, x)
			a.ByEdge[e] = append(a.ByEdge[e], x)
		}
	}
	global := map[EdgeKey]int{}
	for _, d := range a.Data {
		for e, n := range d.Edges {
			global[e] += n
		}
	}
	for _, e := range a.Edges {
		xs := a.ByEdge[e]
		r := &EdgeSummary{EdgeKey: e, GlobalCount: global[e], EligibleBlocks: len(xs), EmpiricalP: 1, FDRQ: 1}
		var vals []float64
		joints, cs, hs := map[string]bool{}, map[string]bool{}, map[string]bool{}
		totalObs := 0
		sumAbs := 0.
		for _, x := range xs {
			vals = append(vals, x.Log2Enrichment)
			joints[x.Joint] = true
			cs[x.Currier] = true
			hs[x.Hand] = true
			totalObs += x.EdgeCount
			sumAbs += math.Abs(x.Log2Enrichment)
			if x.Log2Enrichment > 1e-12 {
				r.PositiveBlocks++
			} else if x.Log2Enrichment < -1e-12 {
				r.NegativeBlocks++
			} else {
				r.NeutralBlocks++
			}
		}
		r.JointClasses = len(joints)
		r.CurrierClasses = len(cs)
		r.Hands = len(hs)
		r.MedianLog2 = median(vals)
		r.MeanLog2 = mean(vals)
		r.BetweenBlockSD = sd(vals)
		r.ExpectedSign = "preferred"
		same := r.PositiveBlocks
		if r.MedianLog2 < 0 {
			r.ExpectedSign = "depleted"
			same = r.NegativeBlocks
		}
		if len(xs) > 0 {
			r.SignConsistency = float64(same) / float64(len(xs))
		}
		for _, x := range xs {
			if totalObs > 0 {
				r.MaxBlockObservationFraction = math.Max(r.MaxBlockObservationFraction, float64(x.EdgeCount)/float64(totalObs))
			}
			if sumAbs > 0 {
				r.MaxBlockEffectWeightFraction = math.Max(r.MaxBlockEffectWeightFraction, math.Abs(x.Log2Enrichment)/sumAbs)
			}
		}
		lobo(r, xs)
		a.Summaries = append(a.Summaries, r)
	}
}
func lobo(r *EdgeSummary, xs []BlockStats) {
	for i, x := range xs {
		var train []float64
		for j, y := range xs {
			if i != j {
				train = append(train, y.Log2Enrichment)
			}
		}
		if len(train) == 0 {
			continue
		}
		sgn := median(train)
		if sgn == 0 {
			continue
		}
		r.TestedBlocks++
		if (sgn > 0 && x.Log2Enrichment > 0) || (sgn < 0 && x.Log2Enrichment < 0) {
			r.SuccessfulSignPredictions++
		}
	}
	if r.TestedBlocks > 0 {
		r.TransferFraction = float64(r.SuccessfulSignPredictions) / float64(r.TestedBlocks)
	}
}

// permutedStatistics performs the specified conditional-on-opportunity null:
// every destination token is shuffled among transition opportunities in its
// own block, while source positions remain fixed.
type profileNullStat struct{ Correlation, SignAgreement, EntropyEffect float64 }

func permutedStatistics(a *analysis, rep int, seed int64, minBlock int) (map[EdgeKey]float64, map[string]profileNullStat, map[string]profileNullStat) {
	rng := rand.New(rand.NewSource(seed + int64(rep)*0x1f123bb5))
	vals := map[EdgeKey][]float64{}
	outBlocks := map[string][]map[string]float64{}
	inBlocks := map[string][]map[string]float64{}
	outEntropy, inEntropy := map[string][]float64{}, map[string][]float64{}
	for _, d := range a.Data {
		pe := permuteBlockEdges(d, rng)
		pd := d
		pd.Edges = pe
		for _, e := range a.Edges {
			if d.Opp[e.Source] >= minBlock {
				vals[e] = append(vals[e], effect(pd, e, len(a.Vocab)).Log2Enrichment)
			}
		}
		collectEffectVectors(pd, a.Vocab, minBlock, &outBlocks, &inBlocks)
		collectEntropyEffects(pd, a.Vocab, minBlock, outEntropy, inEntropy)
	}
	effectStats := map[EdgeKey]float64{}
	for e, x := range vals {
		effectStats[e] = median(x)
	}
	return effectStats, profileLOBOStats(outBlocks, outEntropy, a.Vocab), profileLOBOStats(inBlocks, inEntropy, a.Vocab)
}

func permuteBlockEdges(d blockData, rng *rand.Rand) map[EdgeKey]int {
	dest := append([]string(nil), d.Destinations...)
	rng.Shuffle(len(dest), func(i, j int) { dest[i], dest[j] = dest[j], dest[i] })
	pe := map[EdgeKey]int{}
	k := 0
	for i := 0; i+1 < len(d.Block.Tokens); i++ {
		s := d.Block.Tokens[i].Text
		if !d.Eligible[s] {
			continue
		}
		if d.Eligible[dest[k]] {
			pe[EdgeKey{s, dest[k]}]++
		}
		k++
	}
	return pe
}

func collectEffectVectors(d blockData, vocab []string, minBlock int, out, in *map[string][]map[string]float64) {
	for _, s := range vocab {
		if d.Opp[s] < minBlock {
			continue
		}
		v := map[string]float64{}
		for _, t := range vocab {
			v[t] = effect(d, EdgeKey{s, t}, len(vocab)).Log2Enrichment
		}
		(*out)[s] = append((*out)[s], v)
	}
	// incoming opportunities are target occurrences with an eligible predecessor.
	inOpp := map[string]int{}
	for e, n := range d.Edges {
		inOpp[e.Target] += n
	}
	n := len(d.Block.Tokens)
	for _, t := range vocab {
		if inOpp[t] < minBlock {
			continue
		}
		v := map[string]float64{}
		for _, s := range vocab {
			pc := (float64(d.Edges[EdgeKey{s, t}]) + alpha) / (float64(inOpp[t]) + alpha*float64(len(vocab)))
			pb := (float64(d.Counts[s]) + alpha) / (float64(n) + alpha*float64(len(vocab)))
			v[s] = math.Log2(pc / pb)
		}
		(*in)[t] = append((*in)[t], v)
	}
}

func vectorCorrelation(x, y map[string]float64, vocab []string) float64 {
	a, b := make([]float64, len(vocab)), make([]float64, len(vocab))
	for i, t := range vocab {
		a[i], b[i] = x[t], y[t]
	}
	return pearson(a, b)
}
func vectorSignAgreement(x, y map[string]float64, vocab []string) float64 {
	if len(vocab) == 0 {
		return 0
	}
	n := 0
	for _, t := range vocab {
		if (x[t] > 0) == (y[t] > 0) {
			n++
		}
	}
	return float64(n) / float64(len(vocab))
}
func averageVectors(xs []map[string]float64, skip int, vocab []string) map[string]float64 {
	v := map[string]float64{}
	n := 0
	for i, x := range xs {
		if i == skip {
			continue
		}
		n++
		for _, t := range vocab {
			v[t] += x[t]
		}
	}
	if n > 0 {
		for _, t := range vocab {
			v[t] /= float64(n)
		}
	}
	return v
}
func profileLOBOStats(blocks map[string][]map[string]float64, entropyEffects map[string][]float64, vocab []string) map[string]profileNullStat {
	out := map[string]profileNullStat{}
	for tok, xs := range blocks {
		if len(xs) < 3 {
			continue
		}
		var c, s []float64
		for i, x := range xs {
			ref := averageVectors(xs, i, vocab)
			c = append(c, vectorCorrelation(x, ref, vocab))
			s = append(s, vectorSignAgreement(x, ref, vocab))
		}
		out[tok] = profileNullStat{Correlation: median(c), SignAgreement: mean(s), EntropyEffect: median(entropyEffects[tok])}
	}
	return out
}

func collectEntropyEffects(d blockData, vocab []string, minBlock int, out, in map[string][]float64) {
	V := len(vocab)
	n := len(d.Block.Tokens)
	base := make([]float64, V)
	for i, t := range vocab {
		base[i] = (float64(d.Counts[t]) + alpha) / (float64(n) + alpha*float64(V))
	}
	normalizeProb(base)
	bh := entropy(base)
	for _, s := range vocab {
		if d.Opp[s] < minBlock {
			continue
		}
		p := make([]float64, V)
		for i, t := range vocab {
			p[i] = (float64(d.Edges[EdgeKey{s, t}]) + alpha) / (float64(d.Opp[s]) + alpha*float64(V))
		}
		normalizeProb(p)
		out[s] = append(out[s], entropy(p)-bh)
	}
	for _, t := range vocab {
		opp := incomingOpp(d, t)
		if opp < minBlock {
			continue
		}
		p := make([]float64, V)
		for i, s := range vocab {
			p[i] = (float64(d.Edges[EdgeKey{s, t}]) + alpha) / (float64(opp) + alpha*float64(V))
		}
		normalizeProb(p)
		in[t] = append(in[t], entropy(p)-bh)
	}
}

func observedProfiles(a *analysis, minBlock int) {
	out, in := map[string][]map[string]float64{}, map[string][]map[string]float64{}
	rawOut, rawIn := map[string][][]float64{}, map[string][][]float64{}
	for _, d := range a.Data {
		collectEffectVectors(d, a.Vocab, minBlock, &out, &in)
		collectRawProfiles(d, a.Vocab, minBlock, rawOut, rawIn)
		profileRowsAndEntropy(a, d, minBlock)
	}
	appendStability := func(direction string, all map[string][]map[string]float64, raw map[string][][]float64) {
		for _, tok := range a.Vocab {
			xs := all[tok]
			if len(xs) == 0 {
				continue
			}
			r := ProfileStability{Token: tok, Direction: direction, GlobalCount: a.Counts[tok], EligibleBlocks: len(xs), PermutationP: 1, SignPermutationP: 1, EntropyPermutationP: 1}
			joints := map[string]bool{}
			for _, d := range a.Data {
				if (direction == "outgoing" && d.Opp[tok] >= minBlock) || (direction == "incoming" && incomingOpp(d, tok) >= minBlock) {
					joints[d.Block.Joint] = true
				}
			}
			r.JointClasses = len(joints)
			var pair []float64
			for i := range raw[tok] {
				for j := i + 1; j < len(raw[tok]); j++ {
					pair = append(pair, jsSimilarity(raw[tok][i], raw[tok][j]))
				}
			}
			r.PairwiseJSMean = mean(pair)
			r.PairwiseJSMedian = median(pair)
			r.PairwiseJSMin = 1
			if len(pair) > 0 {
				sort.Float64s(pair)
				r.PairwiseJSMin = pair[0]
			}
			r.PairwiseJSSD = sd(pair)
			var cs, rs, ss []float64
			for i, x := range xs {
				ref := averageVectors(xs, i, a.Vocab)
				cs = append(cs, vectorCorrelation(x, ref, a.Vocab))
				xv, yv := make([]float64, len(a.Vocab)), make([]float64, len(a.Vocab))
				for j, t := range a.Vocab {
					xv[j], yv[j] = x[t], ref[t]
				}
				rs = append(rs, spearman(xv, yv))
				ss = append(ss, vectorSignAgreement(x, ref, a.Vocab))
			}
			r.LOBOMedianCorrelation = median(cs)
			r.LOBOMeanCorrelation = mean(cs)
			r.LOBOMedianSpearman = median(rs)
			r.SignAgreement = mean(ss)
			var ee []float64
			pos, neg := 0, 0
			for _, e := range a.Entropies {
				if e.Token == tok && e.Direction == direction && e.Eligible {
					ee = append(ee, e.EntropyEffect)
					if e.EntropyEffect > 0 {
						pos++
					} else if e.EntropyEffect < 0 {
						neg++
					}
				}
			}
			r.EntropyEffect = median(ee)
			if len(ee) > 0 {
				same := pos
				if r.EntropyEffect < 0 {
					same = neg
				}
				r.EntropySignConsistency = float64(same) / float64(len(ee))
			}
			a.Stability = append(a.Stability, r)
		}
	}
	appendStability("outgoing", out, rawOut)
	appendStability("incoming", in, rawIn)
}

func collectRawProfiles(d blockData, vocab []string, minBlock int, out, in map[string][][]float64) {
	V := len(vocab)
	for _, s := range vocab {
		if d.Opp[s] < minBlock {
			continue
		}
		p := make([]float64, V)
		for i, t := range vocab {
			p[i] = (float64(d.Edges[EdgeKey{s, t}]) + alpha) / (float64(d.Opp[s]) + alpha*float64(V))
		}
		normalizeProb(p)
		out[s] = append(out[s], p)
	}
	for _, t := range vocab {
		opp := incomingOpp(d, t)
		if opp < minBlock {
			continue
		}
		p := make([]float64, V)
		for i, s := range vocab {
			p[i] = (float64(d.Edges[EdgeKey{s, t}]) + alpha) / (float64(opp) + alpha*float64(V))
		}
		normalizeProb(p)
		in[t] = append(in[t], p)
	}
}
func normalizeProb(p []float64) {
	s := 0.
	for _, v := range p {
		s += v
	}
	if s > 0 {
		for i := range p {
			p[i] /= s
		}
	}
}
func effectToProb(x map[string]float64, vocab []string) []float64 {
	p := make([]float64, len(vocab))
	s := 0.
	for i, t := range vocab {
		p[i] = math.Pow(2, x[t])
		s += p[i]
	}
	if s > 0 {
		for i := range p {
			p[i] /= s
		}
	}
	return p
}
func incomingOpp(d blockData, t string) int {
	n := 0
	for e, x := range d.Edges {
		if e.Target == t {
			n += x
		}
	}
	return n
}
func profileRowsAndEntropy(a *analysis, d blockData, minBlock int) {
	V := len(a.Vocab)
	n := len(d.Block.Tokens)
	baseP := make([]float64, V)
	for i, t := range a.Vocab {
		baseP[i] = (float64(d.Counts[t]) + alpha) / (float64(n) + alpha*float64(V))
	}
	normalizeProb(baseP)
	baseH := entropy(baseP)
	for _, s := range a.Vocab {
		if d.Opp[s] < minBlock {
			continue
		}
		p := make([]float64, V)
		for i, t := range a.Vocab {
			x := effect(d, EdgeKey{s, t}, V)
			p[i] = x.PConditional
			a.OutgoingRows = append(a.OutgoingRows, profileRow{s, d.Block.ID, "outgoing", t, d.Edges[EdgeKey{s, t}], x.PConditional, x.Log2Enrichment})
		}
		normalizeProb(p)
		h := entropy(p)
		a.Entropies = append(a.Entropies, EntropyRow{s, d.Block.ID, "outgoing", h, math.Exp(h), baseH, h - baseH, true})
	}
	for _, t := range a.Vocab {
		opp := incomingOpp(d, t)
		if opp < minBlock {
			continue
		}
		p := make([]float64, V)
		for i, s := range a.Vocab {
			pc := (float64(d.Edges[EdgeKey{s, t}]) + alpha) / (float64(opp) + alpha*float64(V))
			p[i] = pc
			le := math.Log2(pc / baseP[i])
			a.IncomingRows = append(a.IncomingRows, profileRow{t, d.Block.ID, "incoming", s, d.Edges[EdgeKey{s, t}], pc, le})
		}
		normalizeProb(p)
		h := entropy(p)
		a.Entropies = append(a.Entropies, EntropyRow{t, d.Block.ID, "incoming", h, math.Exp(h), baseH, h - baseH, true})
	}
}

// classify assigns r.Status. In generic mode (see Config.Generic) there is
// no real Currier/hand covariate, so the JointClasses<2 branch - "this
// effect appears in only one metadata class" - is relabeled GROUP_SPECIFIC
// ("...only one generic resampling group") rather than METADATA_SPECIFIC;
// the JointClasses gate itself is unchanged, since it was already reasoning
// about the group/joint partition alone.
func classify(r *EdgeSummary, generic bool) {
	switch {
	case r.EligibleBlocks < 3:
		r.Status = "INSUFFICIENT_SUPPORT"
	case r.FDRQ > 0.05:
		r.Status = "EXPECTED_FROM_NULL"
	case r.MaxBlockObservationFraction > 0.7 || r.MaxBlockEffectWeightFraction > 0.7:
		r.Status = "BLOCK_CONCENTRATED"
	case r.JointClasses < 2:
		if generic {
			r.Status = "GROUP_SPECIFIC"
		} else {
			r.Status = "METADATA_SPECIFIC"
		}
	case r.SignConsistency < .75 || r.TransferFraction < .67:
		r.Status = "SIGNIFICANT_UNSTABLE"
	case r.ExpectedSign == "preferred":
		r.Status = "BACKBONE_PREFERRED"
	default:
		r.Status = "BACKBONE_DEPLETED"
	}
}

func validateConfig(c Config) error {
	if c.MinTokenCount < 1 || c.MinBlockTokenCount < 1 {
		return fmt.Errorf("count thresholds must be positive")
	}
	if c.Permutations < 1 {
		return fmt.Errorf("permutations must be positive")
	}
	if c.RefinePermutations < c.Permutations {
		return fmt.Errorf("refine-permutations must be >= permutations")
	}
	return nil
}
