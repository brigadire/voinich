package tokenrelationvalidation

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func analyze(c Config, pgr *progressReporter) (Analysis, error) {
	a := Analysis{Parameters: c}
	pgr.begin(1, "Loading canonical corpus and metadata")
	corpus, sha, e := loadCorpus(c.CorpusPath)
	if e != nil {
		return a, e
	}
	var tokens []Token
	var blocks []Block
	var unknown int
	var msha string
	if c.Generic {
		tokens, blocks, unknown, msha, e = loadGenericMetadata(c.CorpusPath, corpus)
	} else {
		tokens, blocks, unknown, msha, e = loadMetadata(c.MetadataPath, corpus)
	}
	_ = tokens
	if e != nil {
		return a, e
	}
	a.CorpusSHA, a.MetadataSHA, a.TokenCount, a.UnknownTokens, a.Blocks = sha, msha, len(corpus), unknown, blocks
	pgr.update(1, 1, "Loading canonical corpus and metadata")
	pgr.begin(2, "Freezing pre-metadata candidates")
	candidates, files, maxD, e := loadCandidates(c.DiscoveryDir)
	if e != nil {
		return a, e
	}
	a.Candidates, a.Files = candidates, files
	pgr.update(1, 1, "Freezing pre-metadata candidates")
	pgr.begin(3, "Validating directional relations by block")
	var directional []Candidate
	for _, x := range candidates {
		if x.Family == "directional" {
			directional = append(directional, x)
		}
	}
	for i, x := range directional {
		for _, b := range blocks {
			a.DirectionBlocks = append(a.DirectionBlocks, directionForBlock(x, b, maxD))
		}
		pgr.update(i+1, len(directional), "Validating directional relations by block")
	}
	pgr.begin(4, "Building exact-distance and structural profiles")
	profiles := make(map[string]localProfiles, len(blocks))
	for i, b := range blocks {
		profiles[b.ID] = buildLocalProfiles(b, maxD)
		pgr.update(i+1, len(blocks), "Building exact-distance and structural profiles")
	}
	var profileCandidates []Candidate
	for _, x := range candidates {
		if x.Family == "distance-profile" || x.Family == "structural" {
			profileCandidates = append(profileCandidates, x)
		}
	}
	for _, x := range profileCandidates {
		for _, b := range blocks {
			a.ProfileBlocks = append(a.ProfileBlocks, profileForBlock(x, b, profiles[b.ID], maxD))
		}
	}
	a.DistancePairwise = map[string][]float64{}
	a.DistancePairwiseOverlap = map[string][]float64{}
	pooledProfiles := mergeLocalProfiles(profiles, "", maxD)
	for _, cand := range profileCandidates {
		if cand.Family != "distance-profile" {
			continue
		}
		for n := range a.ProfileBlocks {
			x := &a.ProfileBlocks[n]
			if x.CandidateID != cand.ID || !x.EligiblePrimary {
				continue
			}
			ref := mergeLocalProfiles(profiles, x.BlockID, maxD)
			x.GlobalSimilarity, _ = compareDistanceProfiles(profiles[x.BlockID], ref, cand.A, cand.B, maxD)
			x.PooledSimilarity, _ = compareDistanceProfiles(profiles[x.BlockID], pooledProfiles, cand.A, cand.B, maxD)
			var train []float64
			for bi := 0; bi < len(blocks); bi++ {
				if blocks[bi].ID == x.BlockID {
					continue
				}
				pi := profileForBlock(cand, blocks[bi], profiles[blocks[bi].ID], maxD)
				if !pi.EligiblePrimary {
					continue
				}
				for bj := bi + 1; bj < len(blocks); bj++ {
					if blocks[bj].ID == x.BlockID {
						continue
					}
					pj := profileForBlock(cand, blocks[bj], profiles[blocks[bj].ID], maxD)
					if pj.EligiblePrimary {
						v, _ := compareDistanceProfiles(profiles[blocks[bi].ID], profiles[blocks[bj].ID], cand.A, cand.B, maxD)
						train = append(train, v)
					}
				}
			}
			x.TrainingReference = mean(train)
		}
		for i := 0; i < len(blocks); i++ {
			pi := profileForBlock(cand, blocks[i], profiles[blocks[i].ID], maxD)
			if !pi.EligiblePrimary {
				continue
			}
			for j := i + 1; j < len(blocks); j++ {
				pj := profileForBlock(cand, blocks[j], profiles[blocks[j].ID], maxD)
				if !pj.EligiblePrimary {
					continue
				}
				v, overlap := compareDistanceProfiles(profiles[blocks[i].ID], profiles[blocks[j].ID], cand.A, cand.B, maxD)
				a.DistancePairwise[cand.ID] = append(a.DistancePairwise[cand.ID], v)
				a.DistancePairwiseOverlap[cand.ID] = append(a.DistancePairwiseOverlap[cand.ID], overlap)
			}
		}
	}
	pgr.begin(5, "Validating frozen sequences")
	var sequenceCandidates []Candidate
	for _, x := range candidates {
		if x.Family == "sequence" {
			sequenceCandidates = append(sequenceCandidates, x)
		}
	}
	for i, x := range sequenceCandidates {
		a.Sequences = append(a.Sequences, sequenceRecurrence(x, blocks))
		pgr.update(i+1, len(sequenceCandidates), "Validating frozen sequences")
	}
	pgr.begin(6, "Leave-one-block-out and metadata transfer")
	a.Summaries, a.Transfers = summarize(a)
	a.MetadataTransfers = buildMetadataTransfers(a.Summaries, a.DirectionBlocks, a.ProfileBlocks, c.Generic)
	pgr.update(1, 1, "Leave-one-block-out and metadata transfer")
	pgr.begin(7, "Matched and within-block permutation controls")
	a.Controls = buildControls(&a, maxD, c.Seed)
	if c.Permutations > 0 {
		if e = directionPermutations(&a, maxD, c, pgr); e != nil {
			return a, e
		}
		if e = sequencePermutations(&a, maxD, c, pgr); e != nil {
			return a, e
		}
		if e = profilePermutations(&a, maxD, c, pgr); e != nil {
			return a, e
		}
		if c.RefinePermutations > c.Permutations {
			if e = refineDirectionalPermutations(&a, maxD, c, pgr); e != nil {
				return a, e
			}
			if e = refineSequencePermutations(&a, maxD, c, pgr); e != nil {
				return a, e
			}
			if e = refineProfilePermutations(&a, maxD, c, pgr); e != nil {
				return a, e
			}
		}
	}
	applyFDR(&a)
	classifyAll(&a, c.Generic)
	sort.Slice(a.Controls, func(i, j int) bool {
		left := a.Controls[i].CandidateID + "\x00" + a.Controls[i].Kind + "\x00" + a.Controls[i].ControlA + "\x00" + a.Controls[i].ControlB
		right := a.Controls[j].CandidateID + "\x00" + a.Controls[j].Kind + "\x00" + a.Controls[j].ControlA + "\x00" + a.Controls[j].ControlB
		return left < right
	})
	pgr.update(1, 1, "Matched and within-block permutation controls")
	return a, nil
}

func sequenceRecurrence(c Candidate, blocks []Block) SequenceResult {
	x := SequenceResult{CandidateID: c.ID, Sequence: c.Sequence}
	needle := strings.Split(c.Sequence, " ")
	bs, js, cs, hs := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}
	maxBlock := 0
	for _, b := range blocks {
		n := 0
		for _, line := range splitLines(b) {
			for i := 0; i+len(needle) <= len(line); i++ {
				ok := true
				for j, w := range needle {
					if line[i+j].Text != w {
						ok = false
						break
					}
				}
				if ok {
					n++
				}
			}
		}
		if n > 0 {
			bs[b.ID] = true
			js[b.Joint] = true
			cs[b.Currier] = true
			hs[b.Hand] = true
			x.Total += n
			if n > maxBlock {
				maxBlock = n
			}
		}
	}
	x.PhysicalBlocks, x.JointClasses, x.CurrierClasses, x.Hands = len(bs), len(js), len(cs), len(hs)
	if x.Total > 0 {
		x.MaxBlockFraction = float64(maxBlock) / float64(x.Total)
	}
	x.HighRecurrence = x.PhysicalBlocks >= 3 && x.MaxBlockFraction <= .7
	return x
}

func summarize(a Analysis) ([]RelationSummary, []Transfer) {
	byDir := map[string][]DirectionBlock{}
	for _, x := range a.DirectionBlocks {
		if x.Eligible {
			byDir[x.CandidateID] = append(byDir[x.CandidateID], x)
		}
	}
	byProf := map[string][]ProfileBlock{}
	for _, x := range a.ProfileBlocks {
		if x.EligiblePrimary {
			byProf[x.CandidateID] = append(byProf[x.CandidateID], x)
		}
	}
	var out []RelationSummary
	var transfers []Transfer
	for _, c := range a.Candidates {
		if c.Family == "sequence" {
			continue
		}
		s := RelationSummary{CandidateID: c.ID, Family: c.Family, A: c.A, B: c.B, Sequence: c.Sequence, RawP: 1, FDRQ: 1}
		joint, cur, hand := map[string]bool{}, map[string]bool{}, map[string]bool{}
		if c.Family == "directional" {
			xs := byDir[c.ID]
			s.EligibleBlocks = len(xs)
			scores, enrich := []float64{}, []float64{}
			positive, negative := 0, 0
			weightedN := 0
			for _, x := range xs {
				joint[x.Joint] = true
				cur[x.Currier] = true
				hand[x.Hand] = true
				scores = append(scores, x.Score)
				enrich = append(enrich, x.EnrichmentAB)
				if x.Score > 0 {
					positive++
				} else if x.Score < 0 {
					negative++
				} else {
					s.NeutralBlocks++
				}
				s.WeightedDirection += x.Score * float64(x.Observations)
				weightedN += x.Observations
			}
			s.PositiveBlocks, s.NegativeBlocks = positive, negative
			if len(xs) > 0 {
				s.SignConsistency = float64(max(positive, negative)) / float64(len(xs))
				s.UnweightedDirection = mean(scores)
				if weightedN > 0 {
					s.WeightedDirection /= float64(weightedN)
				}
				s.BetweenBlockVariance = variance(scores)
				sort.Float64s(enrich)
				if len(enrich) > 0 {
					s.MedianEnrichment = enrich[len(enrich)/2]
				}
			}
			for i, h := range xs {
				train := append([]DirectionBlock(nil), xs[:i]...)
				train = append(train, xs[i+1:]...)
				if len(train) == 0 {
					continue
				}
				expected := meanDirection(train)
				success := sign(expected) != 0 && sign(expected) == sign(h.Score)
				transfers = append(transfers, Transfer{CandidateID: c.ID, Family: c.Family, HeldoutBlock: h.BlockID, TrainMetadata: "all other primary blocks", HeldoutMetadata: h.Joint, Expected: expected, Observed: h.Score, Success: success})
				s.TestedHeldout++
				if success {
					s.SuccessfulHeldout++
				}
			}
		} else {
			xs := byProf[c.ID]
			s.EligibleBlocks = len(xs)
			vals := []float64{}
			for _, x := range xs {
				joint[x.Joint] = true
				cur[x.Currier] = true
				hand[x.Hand] = true
				v := x.Similarity
				if c.Family == "distance-profile" {
					v = x.GlobalSimilarity
				}
				vals = append(vals, v)
			}
			s.ProfileMean, s.ProfileMedian, s.ProfileMin, s.ProfileSD = distribution(vals)
			if c.Family == "structural" {
				above := 0
				for _, v := range vals {
					if v >= c.FrozenThreshold {
						above++
					}
				}
				if len(vals) > 0 {
					s.FractionAboveThreshold = float64(above) / float64(len(vals))
				}
			}
			for i, h := range xs {
				if c.Family == "distance-profile" {
					if h.TrainingReference == 0 {
						continue
					}
					success := h.GlobalSimilarity >= h.TrainingReference
					transfers = append(transfers, Transfer{CandidateID: c.ID, Family: c.Family, HeldoutBlock: h.BlockID, TrainMetadata: "all other primary blocks", HeldoutMetadata: h.Joint, Expected: h.TrainingReference, Observed: h.GlobalSimilarity, Success: success})
					s.TestedHeldout++
					if success {
						s.SuccessfulHeldout++
					}
					continue
				}
				train := append([]ProfileBlock(nil), xs[:i]...)
				train = append(train, xs[i+1:]...)
				if len(train) == 0 {
					continue
				}
				expected := meanProfile(train, c.Family)
				observed := h.Similarity
				threshold := c.FrozenThreshold
				success := observed >= threshold
				transfers = append(transfers, Transfer{CandidateID: c.ID, Family: c.Family, HeldoutBlock: h.BlockID, TrainMetadata: "all other primary blocks", HeldoutMetadata: h.Joint, Expected: expected, Observed: observed, Success: success})
				s.TestedHeldout++
				if success {
					s.SuccessfulHeldout++
				}
			}
			if c.Family == "distance-profile" && len(a.DistancePairwise[c.ID]) > 0 {
				s.ProfileMean, s.ProfileMedian, s.ProfileMin, s.ProfileSD = distribution(a.DistancePairwise[c.ID])
				s.ProfileOverlapMean = mean(a.DistancePairwiseOverlap[c.ID])
			}
		}
		s.PhysicalBlocks, s.JointClasses, s.CurrierClasses, s.Hands = s.EligibleBlocks, len(joint), len(cur), len(hand)
		if s.TestedHeldout > 0 {
			s.TransferSuccess = float64(s.SuccessfulHeldout) / float64(s.TestedHeldout)
		}
		out = append(out, s)
	}
	for _, q := range a.Sequences {
		s := RelationSummary{CandidateID: q.CandidateID, Family: "sequence", Sequence: q.Sequence, EligibleBlocks: q.PhysicalBlocks, PhysicalBlocks: q.PhysicalBlocks, JointClasses: q.JointClasses, CurrierClasses: q.CurrierClasses, Hands: q.Hands, RawP: 1, FDRQ: 1, ProfileMedian: 1 - q.MaxBlockFraction}
		if q.HighRecurrence {
			s.TransferSuccess = 1
			s.ProfileMedian = 1
		}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CandidateID < out[j].CandidateID })
	return out, transfers
}
func meanDirection(x []DirectionBlock) float64 {
	v := make([]float64, len(x))
	for i, z := range x {
		v[i] = z.Score
	}
	return mean(v)
}
func meanProfile(x []ProfileBlock, f string) float64 {
	v := make([]float64, len(x))
	for i, z := range x {
		v[i] = z.Similarity
		if f == "distance-profile" {
			v[i] = z.GlobalSimilarity
		}
	}
	return mean(v)
}

// buildMetadataTransfers computes leave-one-block-out transfer success
// grouped by metadata dimension. In generic mode there is no real hand
// dimension (see Config.Generic), so only the single Currier-carried
// dimension (the deterministic resampling Group, in that mode) is
// computed, and it is labeled "group" rather than "Currier" so a reader
// never mistakes it for a real manuscript covariate.
func buildMetadataTransfers(summaries []RelationSummary, dirs []DirectionBlock, profs []ProfileBlock, generic bool) []MetadataTransfer {
	var out []MetadataTransfer
	dims := []string{"Currier", "hand"}
	if generic {
		dims = []string{"group"}
	}
	for _, s := range summaries {
		type obs struct {
			block, c, h string
			v           float64
		}
		var xs []obs
		if s.Family == "directional" {
			for _, x := range dirs {
				if x.CandidateID == s.CandidateID && x.Eligible {
					xs = append(xs, obs{x.BlockID, x.Currier, x.Hand, x.Score})
				}
			}
		} else if s.Family == "structural" || s.Family == "distance-profile" {
			for _, x := range profs {
				if x.CandidateID == s.CandidateID && x.EligiblePrimary {
					v := x.Similarity
					if s.Family == "distance-profile" {
						v = x.GlobalSimilarity
					}
					xs = append(xs, obs{x.BlockID, x.Currier, x.Hand, v})
				}
			}
		}
		for _, dim := range dims {
			type key struct{ a, b string }
			m := map[key]*MetadataTransfer{}
			for i, tr := range xs {
				for j, ho := range xs {
					if i == j || tr.block == ho.block {
						continue
					}
					a, b := tr.c, ho.c
					if dim == "hand" {
						a, b = tr.h, ho.h
					}
					k := key{a, b}
					z := m[k]
					if z == nil {
						z = &MetadataTransfer{CandidateID: s.CandidateID, Family: s.Family, Dimension: dim, Training: a, Heldout: b}
						m[k] = z
					}
					z.Tested++
					success := false
					if s.Family == "directional" {
						success = sign(tr.v) != 0 && sign(tr.v) == sign(ho.v)
					} else {
						success = ho.v >= tr.v
					}
					if success {
						z.Successful++
					}
				}
			}
			for _, z := range m {
				z.Fraction = float64(z.Successful) / float64(z.Tested)
				out = append(out, *z)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CandidateID+out[i].Dimension+out[i].Training+out[i].Heldout < out[j].CandidateID+out[j].Dimension+out[j].Training+out[j].Heldout
	})
	return out
}

func buildControls(a *Analysis, maxD int, seed int64) []Control {
	freq := map[string]int{}
	for _, b := range a.Blocks {
		for _, t := range b.Tokens {
			freq[t.Text]++
		}
	}
	var vocab []string
	for t := range freq {
		vocab = append(vocab, t)
	}
	sort.Strings(vocab)
	frozen := map[Pair]bool{}
	for _, c := range a.Candidates {
		x := Pair{c.A, c.B}
		if x.B < x.A {
			x.A, x.B = x.B, x.A
		}
		frozen[x] = true
	}
	r := rand.New(rand.NewSource(seed))
	var out []Control
	blockProfiles := map[string]localProfiles{}
	for _, b := range a.Blocks {
		blockProfiles[b.ID] = buildLocalProfiles(b, maxD)
	}
	for _, s := range a.Summaries {
		if s.Family == "sequence" {
			continue
		}
		var choices []Pair
		for _, x := range vocab {
			if !matched(freq[x], freq[s.A]) {
				continue
			}
			for _, y := range vocab {
				if x == y || !matched(freq[y], freq[s.B]) {
					continue
				}
				k := Pair{x, y}
				u := k
				if u.B < u.A {
					u.A, u.B = u.B, u.A
				}
				if !frozen[u] {
					choices = append(choices, k)
				}
				if len(choices) >= 256 {
					break
				}
			}
			if len(choices) >= 256 {
				break
			}
		}
		if len(choices) == 0 {
			continue
		}
		observed := s.SignConsistency
		if s.Family != "directional" {
			observed = s.ProfileMedian
		}
		r.Shuffle(len(choices), func(i, j int) { choices[i], choices[j] = choices[j], choices[i] })
		if len(choices) > 20 {
			choices = choices[:20]
		}
		nulls := make([]float64, 0, len(choices))
		exceed := 0
		for _, k := range choices {
			null := controlMetric(s.Family, k, a.Blocks, blockProfiles, maxD)
			nulls = append(nulls, null)
			if null >= observed {
				exceed++
			}
			out = append(out, Control{CandidateID: s.CandidateID, Family: s.Family, Kind: "frequency-matched", ControlA: k.A, ControlB: k.B, Observed: observed, NullMean: null, Permutations: len(choices)})
		}
		raw := float64(1+exceed) / float64(1+len(nulls))
		pct := 100 * float64(len(nulls)-exceed) / float64(len(nulls))
		for n := range out {
			if out[n].CandidateID == s.CandidateID && out[n].Kind == "frequency-matched" {
				out[n].RawP = raw
				out[n].Percentile = pct
			}
		}
		for n := range a.Summaries {
			if a.Summaries[n].CandidateID == s.CandidateID {
				a.Summaries[n].ControlPercentile = pct
				if s.Family != "directional" {
					a.Summaries[n].RawP = raw
				}
			}
		}
	}
	return out
}

func controlMetric(family string, k Pair, blocks []Block, profiles map[string]localProfiles, maxD int) float64 {
	if family == "directional" {
		pos, neg, n := 0, 0, 0
		for _, b := range blocks {
			x := directionForBlock(Candidate{A: k.A, B: k.B}, b, maxD)
			if !x.Eligible {
				continue
			}
			n++
			if x.Score > 0 {
				pos++
			} else if x.Score < 0 {
				neg++
			}
		}
		if n > 0 {
			return float64(max(pos, neg)) / float64(n)
		}
		return 0
	}
	var vals []float64
	for bi, b := range blocks {
		x := profileForBlock(Candidate{Family: family, A: k.A, B: k.B}, b, profiles[b.ID], maxD)
		if !x.EligiblePrimary {
			continue
		}
		if family == "distance-profile" {
			for bj := bi + 1; bj < len(blocks); bj++ {
				other := profileForBlock(Candidate{A: k.A, B: k.B}, blocks[bj], profiles[blocks[bj].ID], maxD)
				if other.EligiblePrimary {
					v, _ := compareDistanceProfiles(profiles[b.ID], profiles[blocks[bj].ID], k.A, k.B, maxD)
					vals = append(vals, v)
				}
			}
		} else {
			vals = append(vals, x.Similarity)
		}
	}
	_, median, _, _ := distribution(vals)
	return median
}
func matched(a, b int) bool { return a > 0 && b > 0 && float64(max(a, b))/float64(min(a, b)) <= 2 }

type checkpoint struct {
	Version                 int                `json:"version"`
	CorpusSHA               string             `json:"corpus_sha256"`
	Seed                    int64              `json:"seed"`
	Candidates              int                `json:"frozen_candidates"`
	Completed               int                `json:"completed_permutations"`
	Exceed                  map[string]int     `json:"directional_exceed"`
	DirectionSum            map[string]float64 `json:"directional_null_sum"`
	RefineCompleted         int                `json:"completed_refine_permutations"`
	RefineExceed            map[string]int     `json:"directional_refine_exceed"`
	RefineSum               map[string]float64 `json:"directional_refine_null_sum"`
	SequenceCompleted       int                `json:"completed_sequence_permutations"`
	SequenceExceed          map[string]int     `json:"sequence_exceed"`
	SequenceSum             map[string]float64 `json:"sequence_null_sum"`
	SequenceRefineCompleted int                `json:"completed_sequence_refine_permutations"`
	SequenceRefineExceed    map[string]int     `json:"sequence_refine_exceed"`
	SequenceRefineSum       map[string]float64 `json:"sequence_refine_null_sum"`
	ProfileCompleted        int                `json:"completed_profile_permutations"`
	ProfileExceed           map[string]int     `json:"profile_exceed"`
	ProfileSum              map[string]float64 `json:"profile_null_sum"`
	ProfileRefineCompleted  int                `json:"completed_profile_refine_permutations"`
	ProfileRefineExceed     map[string]int     `json:"profile_refine_exceed"`
	ProfileRefineSum        map[string]float64 `json:"profile_refine_null_sum"`
}

type directedRef struct {
	id      string
	forward bool
	maxD    int
}

// directionEdges is directionScoresAll's per-replicate-invariant index: for
// every (from,to) token pair, which candidate IDs it references and in
// which direction, plus the widest maxD in play. candidates and defaultMax
// never change across a permutation replicate loop, so buildDirectionEdges
// computes this once for the caller to reuse across every replicate,
// instead of directionScoresAll rebuilding it from scratch on every call.
type directionEdges struct {
	edges     map[Pair][]directedRef
	globalMax int
}

func buildDirectionEdges(candidates map[string]Candidate, defaultMax int) directionEdges {
	edges := map[Pair][]directedRef{}
	globalMax := defaultMax
	for id, c := range candidates {
		d := defaultMax
		if strings.Contains(c.Sources, "begin_end") && int(c.FrozenThreshold) > d {
			d = int(c.FrozenThreshold)
		}
		if d > globalMax {
			globalMax = d
		}
		edges[Pair{c.A, c.B}] = append(edges[Pair{c.A, c.B}], directedRef{id, true, d})
		edges[Pair{c.B, c.A}] = append(edges[Pair{c.B, c.A}], directedRef{id, false, d})
	}
	return directionEdges{edges: edges, globalMax: globalMax}
}

func directionScoresAll(blocks []Block, candidates map[string]Candidate, de directionEdges) map[string]float64 {
	pos, neg, eligible := map[string]int{}, map[string]int{}, map[string]int{}
	for _, block := range blocks {
		freq := map[string]int{}
		for _, t := range block.Tokens {
			freq[t.Text]++
		}
		ab, ba := map[string]int{}, map[string]int{}
		for _, line := range splitLines(block) {
			for i, t := range line {
				for d := 1; d <= de.globalMax && i+d < len(line); d++ {
					refs := de.edges[Pair{t.Text, line[i+d].Text}]
					for _, ref := range refs {
						if d > ref.maxD {
							continue
						}
						if ref.forward {
							ab[ref.id]++
						} else {
							ba[ref.id]++
						}
					}
				}
			}
		}
		for id, c := range candidates {
			n := ab[id] + ba[id]
			if freq[c.A] < 5 || freq[c.B] < 5 || n < 5 {
				continue
			}
			eligible[id]++
			score := float64(ab[id]-ba[id]) / float64(n)
			if score > 0 {
				pos[id]++
			} else if score < 0 {
				neg[id]++
			}
		}
	}
	out := map[string]float64{}
	for id, n := range eligible {
		out[id] = float64(max(pos[id], neg[id])) / float64(n)
	}
	return out
}

func directionPermutations(a *Analysis, maxD int, c Config, pgr *progressReporter) error {
	obs := map[string]float64{}
	for _, s := range a.Summaries {
		if s.Family == "directional" && s.EligibleBlocks >= 1 {
			obs[s.CandidateID] = s.SignConsistency
		}
	}
	cp := checkpoint{Exceed: map[string]int{}}
	if c.CheckpointPath != "" {
		if b, e := os.ReadFile(c.CheckpointPath); e == nil {
			_ = json.Unmarshal(b, &cp)
			if cp.Exceed == nil {
				cp.Exceed = map[string]int{}
			}
		}
	}
	if cp.Version != 1 || cp.CorpusSHA != a.CorpusSHA || cp.Seed != c.Seed || cp.Candidates != len(a.Candidates) {
		cp = checkpoint{Exceed: map[string]int{}}
	}
	if cp.DirectionSum == nil {
		cp.DirectionSum = map[string]float64{}
	}
	cp.Version = 1
	cp.CorpusSHA, cp.Seed, cp.Candidates = a.CorpusSHA, c.Seed, len(a.Candidates)
	executor := permutationExecutorFor(c, a, maxD)
	err := runBattery(executorContext(c), executor, "direction", cp.Completed, c.Permutations, executorWorkers(c), func(run int, scores map[string]float64) error {
		for id := range obs {
			score := scores[id]
			cp.DirectionSum[id] += score
			if score >= obs[id] {
				cp.Exceed[id]++
			}
		}
		cp.Completed = run + 1
		if c.CheckpointPath != "" && (cp.Completed%25 == 0 || cp.Completed == c.Permutations) {
			if e := os.MkdirAll(filepath.Dir(c.CheckpointPath), 0755); e != nil {
				return e
			}
			b, _ := json.MarshalIndent(cp, "", "  ")
			if e := os.WriteFile(c.CheckpointPath, b, 0644); e != nil {
				return e
			}
		}
		pgr.update(cp.Completed, c.Permutations, "Within-block permutation controls")
		return nil
	})
	if err != nil {
		return err
	}
	for i := range a.Summaries {
		if a.Summaries[i].Family == "directional" {
			a.Summaries[i].RawP = float64(1+cp.Exceed[a.Summaries[i].CandidateID]) / float64(c.Permutations+1)
		}
	}
	for id, v := range obs {
		a.Controls = append(a.Controls, Control{CandidateID: id, Family: "directional", Kind: "within-block-permutation", Observed: v, NullMean: cp.DirectionSum[id] / float64(c.Permutations), RawP: float64(1+cp.Exceed[id]) / float64(c.Permutations+1), Permutations: c.Permutations})
	}
	return nil
}

func refineDirectionalPermutations(a *Analysis, maxD int, c Config, pgr *progressReporter) error {
	obs := map[string]float64{}
	for _, s := range a.Summaries {
		if s.Family == "directional" && s.RawP < .01 && s.EligibleBlocks >= 3 && s.JointClasses >= 2 {
			obs[s.CandidateID] = s.SignConsistency
		}
	}
	if len(obs) == 0 {
		return nil
	}
	cp := checkpoint{}
	if c.CheckpointPath != "" {
		if data, e := os.ReadFile(c.CheckpointPath); e == nil {
			_ = json.Unmarshal(data, &cp)
		}
	}
	if cp.Version != 1 || cp.CorpusSHA != a.CorpusSHA || cp.Seed != c.Seed || cp.Candidates != len(a.Candidates) {
		cp = checkpoint{}
	}
	cp.Version = 1
	cp.CorpusSHA, cp.Seed, cp.Candidates = a.CorpusSHA, c.Seed, len(a.Candidates)
	if cp.RefineExceed == nil {
		cp.RefineExceed = map[string]int{}
	}
	if cp.RefineSum == nil {
		cp.RefineSum = map[string]float64{}
	}
	executor := permutationExecutorFor(c, a, maxD)
	err := runBattery(executorContext(c), executor, "refine_direction", cp.RefineCompleted, c.RefinePermutations, executorWorkers(c), func(run int, scores map[string]float64) error {
		for id := range obs {
			score := scores[id]
			cp.RefineSum[id] += score
			if score >= obs[id] {
				cp.RefineExceed[id]++
			}
		}
		cp.RefineCompleted = run + 1
		if c.CheckpointPath != "" && (cp.RefineCompleted%25 == 0 || cp.RefineCompleted == c.RefinePermutations) {
			data, _ := json.MarshalIndent(cp, "", "  ")
			if e := os.WriteFile(c.CheckpointPath, data, 0644); e != nil {
				return e
			}
		}
		pgr.update(cp.RefineCompleted, c.RefinePermutations, "Refining pre-specified candidates")
		return nil
	})
	if err != nil {
		return err
	}
	for i := range a.Summaries {
		if _, ok := obs[a.Summaries[i].CandidateID]; ok {
			a.Summaries[i].RawP = float64(1+cp.RefineExceed[a.Summaries[i].CandidateID]) / float64(c.RefinePermutations+1)
		}
	}
	for id, v := range obs {
		a.Controls = append(a.Controls, Control{CandidateID: id, Family: "directional", Kind: "within-block-permutation-refined", Observed: v, NullMean: cp.RefineSum[id] / float64(c.RefinePermutations), RawP: float64(1+cp.RefineExceed[id]) / float64(c.RefinePermutations+1), Permutations: c.RefinePermutations})
	}
	return nil
}

type sequenceTrie struct {
	Next map[string]*sequenceTrie
	IDs  []string
}

func makeSequenceTrie(candidates []Candidate) *sequenceTrie {
	root := &sequenceTrie{Next: map[string]*sequenceTrie{}}
	for _, c := range candidates {
		if c.Family != "sequence" {
			continue
		}
		node := root
		for _, token := range strings.Split(c.Sequence, " ") {
			if node.Next[token] == nil {
				node.Next[token] = &sequenceTrie{Next: map[string]*sequenceTrie{}}
			}
			node = node.Next[token]
		}
		node.IDs = append(node.IDs, c.ID)
	}
	return root
}
func sequenceScores(blocks []Block, trie *sequenceTrie) map[string]float64 {
	totals, maxima := map[string]int{}, map[string]int{}
	blockCounts := map[string]int{}
	for _, b := range blocks {
		local := map[string]int{}
		for _, line := range splitLines(b) {
			for start := range line {
				node := trie
				for j := start; j < len(line); j++ {
					node = node.Next[line[j].Text]
					if node == nil {
						break
					}
					for _, id := range node.IDs {
						local[id]++
					}
				}
			}
		}
		for id, n := range local {
			totals[id] += n
			blockCounts[id]++
			if n > maxima[id] {
				maxima[id] = n
			}
		}
	}
	out := map[string]float64{}
	for id, total := range totals {
		out[id] = float64(blockCounts[id]) * (1 - float64(maxima[id])/float64(total))
	}
	return out
}
func sequencePermutations(a *Analysis, maxD int, c Config, pgr *progressReporter) error {
	observed := map[string]float64{}
	for _, x := range a.Sequences {
		observed[x.CandidateID] = float64(x.PhysicalBlocks) * (1 - x.MaxBlockFraction)
	}
	cp := checkpoint{}
	if c.CheckpointPath != "" {
		if data, e := os.ReadFile(c.CheckpointPath); e == nil {
			_ = json.Unmarshal(data, &cp)
		}
	}
	if cp.Version != 1 || cp.CorpusSHA != a.CorpusSHA || cp.Seed != c.Seed || cp.Candidates != len(a.Candidates) {
		cp = checkpoint{}
	}
	cp.Version = 1
	cp.CorpusSHA, cp.Seed, cp.Candidates = a.CorpusSHA, c.Seed, len(a.Candidates)
	if cp.SequenceExceed == nil {
		cp.SequenceExceed = map[string]int{}
	}
	if cp.SequenceSum == nil {
		cp.SequenceSum = map[string]float64{}
	}
	executor := permutationExecutorFor(c, a, maxD)
	err := runBattery(executorContext(c), executor, "sequence", cp.SequenceCompleted, c.Permutations, executorWorkers(c), func(run int, scores map[string]float64) error {
		for id, v := range observed {
			cp.SequenceSum[id] += scores[id]
			if scores[id] >= v {
				cp.SequenceExceed[id]++
			}
		}
		cp.SequenceCompleted = run + 1
		if c.CheckpointPath != "" && (cp.SequenceCompleted%25 == 0 || cp.SequenceCompleted == c.Permutations) {
			data, _ := json.MarshalIndent(cp, "", "  ")
			if e := os.WriteFile(c.CheckpointPath, data, 0644); e != nil {
				return e
			}
		}
		pgr.update(cp.SequenceCompleted, c.Permutations, "Sequence permutation controls")
		return nil
	})
	if err != nil {
		return err
	}
	for i := range a.Summaries {
		if a.Summaries[i].Family == "sequence" {
			a.Summaries[i].RawP = float64(1+cp.SequenceExceed[a.Summaries[i].CandidateID]) / float64(c.Permutations+1)
		}
	}
	for id, v := range observed {
		a.Controls = append(a.Controls, Control{CandidateID: id, Family: "sequence", Kind: "within-block-permutation", Observed: v, NullMean: cp.SequenceSum[id] / float64(c.Permutations), RawP: float64(1+cp.SequenceExceed[id]) / float64(c.Permutations+1), Permutations: c.Permutations})
	}
	return nil
}

func refineSequencePermutations(a *Analysis, maxD int, c Config, pgr *progressReporter) error {
	selected := map[string]float64{}
	for _, s := range a.Summaries {
		if s.Family == "sequence" && s.RawP < .01 && s.EligibleBlocks >= 3 && s.JointClasses >= 2 {
			selected[s.CandidateID] = float64(s.PhysicalBlocks) * (1 - sequenceConcentration(a.Sequences, s.CandidateID))
		}
	}
	if len(selected) == 0 {
		return nil
	}
	cp := checkpoint{}
	if c.CheckpointPath != "" {
		if data, e := os.ReadFile(c.CheckpointPath); e == nil {
			_ = json.Unmarshal(data, &cp)
		}
	}
	if cp.Version != 1 || cp.CorpusSHA != a.CorpusSHA || cp.Seed != c.Seed || cp.Candidates != len(a.Candidates) {
		cp = checkpoint{}
	}
	cp.Version = 1
	cp.CorpusSHA, cp.Seed, cp.Candidates = a.CorpusSHA, c.Seed, len(a.Candidates)
	if cp.SequenceRefineExceed == nil {
		cp.SequenceRefineExceed = map[string]int{}
	}
	if cp.SequenceRefineSum == nil {
		cp.SequenceRefineSum = map[string]float64{}
	}
	executor := permutationExecutorFor(c, a, maxD)
	err := runBattery(executorContext(c), executor, "refine_sequence", cp.SequenceRefineCompleted, c.RefinePermutations, executorWorkers(c), func(run int, scores map[string]float64) error {
		for id, v := range selected {
			cp.SequenceRefineSum[id] += scores[id]
			if scores[id] >= v {
				cp.SequenceRefineExceed[id]++
			}
		}
		cp.SequenceRefineCompleted = run + 1
		if c.CheckpointPath != "" && (cp.SequenceRefineCompleted%25 == 0 || cp.SequenceRefineCompleted == c.RefinePermutations) {
			data, _ := json.MarshalIndent(cp, "", "  ")
			if e := os.WriteFile(c.CheckpointPath, data, 0644); e != nil {
				return e
			}
		}
		pgr.update(cp.SequenceRefineCompleted, c.RefinePermutations, "Refining sequence candidates")
		return nil
	})
	if err != nil {
		return err
	}
	for i := range a.Summaries {
		if _, ok := selected[a.Summaries[i].CandidateID]; ok {
			a.Summaries[i].RawP = float64(1+cp.SequenceRefineExceed[a.Summaries[i].CandidateID]) / float64(c.RefinePermutations+1)
		}
	}
	for id, v := range selected {
		a.Controls = append(a.Controls, Control{CandidateID: id, Family: "sequence", Kind: "within-block-permutation-refined", Observed: v, NullMean: cp.SequenceRefineSum[id] / float64(c.RefinePermutations), RawP: float64(1+cp.SequenceRefineExceed[id]) / float64(c.RefinePermutations+1), Permutations: c.RefinePermutations})
	}
	return nil
}
func sequenceConcentration(x []SequenceResult, id string) float64 {
	for _, v := range x {
		if v.CandidateID == id {
			return v.MaxBlockFraction
		}
	}
	return 1
}

func profilePermutationScores(blocks []Block, candidates map[string]Candidate, maxD int, ws *profileWorkspace) map[string]float64 {
	profiles := map[string]localProfiles{}
	for _, b := range blocks {
		profiles[b.ID] = ws.buildLocalProfiles(b, maxD)
	}
	values := map[string][]float64{}
	for id, c := range candidates {
		if c.Family == "structural" {
			for _, b := range blocks {
				x := profileForBlock(c, b, profiles[b.ID], maxD)
				if x.EligiblePrimary {
					values[id] = append(values[id], x.Similarity)
				}
			}
		} else {
			for i := 0; i < len(blocks); i++ {
				pi := profileForBlock(c, blocks[i], profiles[blocks[i].ID], maxD)
				if !pi.EligiblePrimary {
					continue
				}
				for j := i + 1; j < len(blocks); j++ {
					pj := profileForBlock(c, blocks[j], profiles[blocks[j].ID], maxD)
					if pj.EligiblePrimary {
						v, _ := compareDistanceProfiles(profiles[blocks[i].ID], profiles[blocks[j].ID], c.A, c.B, maxD)
						values[id] = append(values[id], v)
					}
				}
			}
		}
	}
	out := map[string]float64{}
	for id, v := range values {
		_, out[id], _, _ = distribution(v)
	}
	return out
}

func profilePermutations(a *Analysis, maxD int, c Config, pgr *progressReporter) error {
	obs := map[string]float64{}
	lookup := map[string]Candidate{}
	for _, s := range a.Summaries {
		if (s.Family == "structural" || s.Family == "distance-profile") && s.EligibleBlocks > 0 {
			obs[s.CandidateID] = s.ProfileMedian
		}
	}
	for _, x := range a.Candidates {
		if _, ok := obs[x.ID]; ok {
			lookup[x.ID] = x
		}
	}
	cp := checkpoint{}
	if c.CheckpointPath != "" {
		if data, e := os.ReadFile(c.CheckpointPath); e == nil {
			_ = json.Unmarshal(data, &cp)
		}
	}
	if cp.Version != 1 || cp.CorpusSHA != a.CorpusSHA || cp.Seed != c.Seed || cp.Candidates != len(a.Candidates) {
		cp = checkpoint{}
	}
	cp.Version = 1
	cp.CorpusSHA, cp.Seed, cp.Candidates = a.CorpusSHA, c.Seed, len(a.Candidates)
	if cp.ProfileExceed == nil {
		cp.ProfileExceed = map[string]int{}
	}
	if cp.ProfileSum == nil {
		cp.ProfileSum = map[string]float64{}
	}
	executor := permutationExecutorFor(c, a, maxD)
	err := runBattery(executorContext(c), executor, "profile", cp.ProfileCompleted, c.Permutations, executorWorkers(c), func(run int, scores map[string]float64) error {
		for id, v := range obs {
			cp.ProfileSum[id] += scores[id]
			if scores[id] >= v {
				cp.ProfileExceed[id]++
			}
		}
		cp.ProfileCompleted = run + 1
		if c.CheckpointPath != "" && (cp.ProfileCompleted%25 == 0 || cp.ProfileCompleted == c.Permutations) {
			data, _ := json.MarshalIndent(cp, "", "  ")
			if e := os.WriteFile(c.CheckpointPath, data, 0644); e != nil {
				return e
			}
		}
		pgr.update(cp.ProfileCompleted, c.Permutations, "Profile permutation controls")
		return nil
	})
	if err != nil {
		return err
	}
	for i := range a.Summaries {
		if _, ok := obs[a.Summaries[i].CandidateID]; ok {
			a.Summaries[i].RawP = float64(1+cp.ProfileExceed[a.Summaries[i].CandidateID]) / float64(c.Permutations+1)
		}
	}
	for id, v := range obs {
		a.Controls = append(a.Controls, Control{CandidateID: id, Family: lookup[id].Family, Kind: "within-block-permutation", Observed: v, NullMean: cp.ProfileSum[id] / float64(c.Permutations), RawP: float64(1+cp.ProfileExceed[id]) / float64(c.Permutations+1), Permutations: c.Permutations})
	}
	return nil
}

func refineProfilePermutations(a *Analysis, maxD int, c Config, pgr *progressReporter) error {
	obs := map[string]float64{}
	lookup := map[string]Candidate{}
	for _, s := range a.Summaries {
		if (s.Family == "structural" || s.Family == "distance-profile") && s.RawP < .01 && s.EligibleBlocks >= 3 && s.JointClasses >= 2 {
			obs[s.CandidateID] = s.ProfileMedian
		}
	}
	if len(obs) == 0 {
		return nil
	}
	for _, x := range a.Candidates {
		if _, ok := obs[x.ID]; ok {
			lookup[x.ID] = x
		}
	}
	cp := checkpoint{}
	if c.CheckpointPath != "" {
		if data, e := os.ReadFile(c.CheckpointPath); e == nil {
			_ = json.Unmarshal(data, &cp)
		}
	}
	if cp.Version != 1 || cp.CorpusSHA != a.CorpusSHA || cp.Seed != c.Seed || cp.Candidates != len(a.Candidates) {
		cp = checkpoint{}
	}
	cp.Version = 1
	cp.CorpusSHA, cp.Seed, cp.Candidates = a.CorpusSHA, c.Seed, len(a.Candidates)
	if cp.ProfileRefineExceed == nil {
		cp.ProfileRefineExceed = map[string]int{}
	}
	if cp.ProfileRefineSum == nil {
		cp.ProfileRefineSum = map[string]float64{}
	}
	executor := permutationExecutorFor(c, a, maxD)
	err := runBattery(executorContext(c), executor, "refine_profile", cp.ProfileRefineCompleted, c.RefinePermutations, executorWorkers(c), func(run int, scores map[string]float64) error {
		for id, v := range obs {
			cp.ProfileRefineSum[id] += scores[id]
			if scores[id] >= v {
				cp.ProfileRefineExceed[id]++
			}
		}
		cp.ProfileRefineCompleted = run + 1
		if c.CheckpointPath != "" && (cp.ProfileRefineCompleted%25 == 0 || cp.ProfileRefineCompleted == c.RefinePermutations) {
			data, _ := json.MarshalIndent(cp, "", "  ")
			if e := os.WriteFile(c.CheckpointPath, data, 0644); e != nil {
				return e
			}
		}
		pgr.update(cp.ProfileRefineCompleted, c.RefinePermutations, "Refining profile candidates")
		return nil
	})
	if err != nil {
		return err
	}
	for i := range a.Summaries {
		if _, ok := obs[a.Summaries[i].CandidateID]; ok {
			a.Summaries[i].RawP = float64(1+cp.ProfileRefineExceed[a.Summaries[i].CandidateID]) / float64(c.RefinePermutations+1)
		}
	}
	for id, v := range obs {
		a.Controls = append(a.Controls, Control{CandidateID: id, Family: lookup[id].Family, Kind: "within-block-permutation-refined", Observed: v, NullMean: cp.ProfileRefineSum[id] / float64(c.RefinePermutations), RawP: float64(1+cp.ProfileRefineExceed[id]) / float64(c.RefinePermutations+1), Permutations: c.RefinePermutations})
	}
	return nil
}

func applyFDR(a *Analysis) {
	families := map[string][]int{}
	for i, s := range a.Summaries {
		families[s.Family] = append(families[s.Family], i)
	}
	for _, idx := range families {
		p := make([]float64, len(idx))
		for j, i := range idx {
			p[j] = a.Summaries[i].RawP
		}
		q := BH(p)
		for j, i := range idx {
			a.Summaries[i].FDRQ = q[j]
		}
	}
	for i := range a.Sequences {
		for _, s := range a.Summaries {
			if s.CandidateID == a.Sequences[i].CandidateID {
				a.Sequences[i].RawP, a.Sequences[i].FDRQ = s.RawP, s.FDRQ
			}
		}
	}
}

// classifyAll assigns each summary's Classification. In generic mode this
// only ever reasons about the single deterministic resampling Group
// dimension (via ClassifyGeneric), and never emits the Currier/hand-
// conditioned vocabulary ClassifyDetailed produces - see
// GENERIC_STAGE_APPLICABILITY_AUDIT.md.
func classifyAll(a *Analysis, generic bool) {
	for i := range a.Summaries {
		s := &a.Summaries[i]
		if generic {
			withinG, crossG := false, false
			for _, m := range a.MetadataTransfers {
				if m.CandidateID != s.CandidateID || m.Fraction < .67 {
					continue
				}
				if m.Dimension == "group" {
					if m.Training == m.Heldout {
						withinG = true
					} else {
						crossG = true
					}
				}
			}
			if s.Family == "sequence" {
				withinG = s.EligibleBlocks >= 2
				crossG = s.CurrierClasses >= 2
			}
			s.Classification = ClassifyGeneric(*s, withinG, crossG)
			continue
		}
		withinC, crossC, withinH, crossH := false, false, false, false
		for _, m := range a.MetadataTransfers {
			if m.CandidateID != s.CandidateID || m.Fraction < .67 {
				continue
			}
			if m.Dimension == "Currier" {
				if m.Training == m.Heldout {
					withinC = true
				} else {
					crossC = true
				}
			}
			if m.Dimension == "hand" {
				if m.Training == m.Heldout {
					withinH = true
				} else {
					crossH = true
				}
			}
		}
		if s.Family == "sequence" {
			withinC = s.EligibleBlocks >= 2
			withinH = s.EligibleBlocks >= 2
			crossC = s.CurrierClasses >= 2
			crossH = s.Hands >= 2
		}
		s.Classification = ClassifyDetailed(*s, withinC, crossC, withinH, crossH)
	}
}
