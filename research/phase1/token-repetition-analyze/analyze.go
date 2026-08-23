package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"

	"zcore.dev/voinich/internal/tokenrepetition"
)

const nearNullPermutations = 1000

// analyzeCorpus runs task60 sections 5-10 and 15-26 on one corpus,
// writing every row-per-corpus TSV and filling in the corpus's summary
// for REPORT.md / cross-family analysis.
func analyzeCorpus(c tokenrepetition.Corpus, mode tokenrepetition.GlyphMode, rng *rand.Rand, rep *report, w *writers) error {
	s := rep.summary(c.Name)
	glyphSeqs := tokenrepetition.GlyphSequences(c.Tokens, mode)

	// --- Exact adjacent repetition (section 5) ---
	adj := tokenrepetition.AdjacentRepetition(c.Tokens, c.LineOfToken, maxLociPerToken)
	s.R2 = adj.R2
	tokens := make([]string, 0, len(adj.Tokens))
	for t := range adj.Tokens {
		tokens = append(tokens, t)
	}
	sort.Strings(tokens)
	for _, t := range tokens {
		rt := adj.Tokens[t]
		loci := make([]string, len(rt.FirstLoci))
		for k, l := range rt.FirstLoci {
			loci[k] = strconv.Itoa(l)
		}
		w.exactAdjacent.row(c.Name, t, i(rt.Frequency), i(rt.AdjacentRepeats), i(rt.MaximumRun), strings.Join(loci, ","))
	}

	// --- Exact runs (section 6-7) ---
	runs := tokenrepetition.ExactRuns(c.Tokens, c.LineOfToken)
	for _, r := range runs {
		w.exactRuns.row(c.Name, r.Token, i(r.RunLength), i(r.StartPosition), i(r.GlobalFrequency))
	}
	maxK := tokenrepetition.MaxObservedRun(runs) + 3
	surv := tokenrepetition.RunLengthSurvival(runs, maxK)
	for k := 2; k <= maxK; k++ {
		frac := 0.0
		if len(c.Tokens) > 0 {
			frac = float64(surv[k]) / float64(len(c.Tokens))
		}
		w.runDistribution.row(c.Name, i(k), i(surv[k]), f8(frac))
	}
	s.MaxRun = tokenrepetition.MaxObservedRun(runs)
	s.Runs3, s.Runs4, s.Runs5 = countGE(runs, 3), countGE(runs, 4), countGE(runs, 5)

	// --- Near repetition (sections 15-22) ---
	if !c.Opaque {
		dists := tokenrepetition.AdjacentEditDistances(c.Tokens, c.LineOfToken, glyphSeqs)
		distSummary := tokenrepetition.SummarizeDistances(dists, glyphSeqs)
		s.PLe1 = distSummary.PLe1
		w.editDistDist.row(c.Name, f8(distSummary.PEq0), f8(distSummary.PEq1), f8(distSummary.PLe1), f8(distSummary.PLe2),
			f8(distSummary.PEq1GivenSameLength), f8(distSummary.SubstitutionOnlyRate), i(distSummary.Total))

		opPos := map[string]int{}
		directional := map[[2]string]int{}
		symmetric := map[[2]string]int{}
		for _, d := range dists {
			if d.Distance != 1 {
				continue
			}
			op, pos, src, tgt, ok := tokenrepetition.ClassifyDistanceOne(glyphSeqs[d.A], glyphSeqs[d.B])
			if !ok {
				continue
			}
			length := len(glyphSeqs[d.A])
			if len(glyphSeqs[d.B]) > length {
				length = len(glyphSeqs[d.B])
			}
			cls := tokenrepetition.PositionClass(pos, length)
			w.editDistOne.row(c.Name, i(d.Position), d.A, d.B, op, cls, src, tgt)
			opPos[op+"|"+cls]++
			if op == "SUBSTITUTION" {
				directional[[2]string{src, tgt}]++
				a, bb := src, tgt
				if a > bb {
					a, bb = bb, a
				}
				symmetric[[2]string{a, bb}]++
			}
		}
		for _, op := range []string{"SUBSTITUTION", "INSERTION", "DELETION"} {
			for _, cls := range []string{"BEGIN", "MIDDLE", "END"} {
				w.editOpPosition.row(c.Name, op, cls, i(opPos[op+"|"+cls]))
			}
		}
		writeMatrix(w.substMatrix, c.Name, "directional", directional)
		writeMatrix(w.substMatrix, c.Name, "symmetric", symmetric)

		// Frequency/length-matched null (section 20).
		matchedRate, _ := tokenrepetition.MatchedNullRate(dists, glyphSeqs, c.Tokens, matchedNullDraws, rankTolerance, rng)
		s.MatchedNullRate = matchedRate

		// Edit-family graph (sections 23-26).
		vocab := make([]string, 0, len(glyphSeqs))
		freq := map[string]int{}
		for _, t := range c.Tokens {
			freq[t]++
		}
		for t := range glyphSeqs {
			vocab = append(vocab, t)
		}
		g := tokenrepetition.BuildEditGraph(vocab, glyphSeqs)
		comps := g.ConnectedComponents()
		degree := g.DegreeOf()
		hist := map[int]int{}
		for _, comp := range comps {
			hist[len(comp)]++
		}
		var sizes []int
		for k := range hist {
			sizes = append(sizes, k)
		}
		sort.Ints(sizes)
		var histParts []string
		for _, k := range sizes {
			histParts = append(histParts, fmt.Sprintf("%d:%d", k, hist[k]))
		}
		largest := 0
		if len(comps) > 0 {
			largest = len(comps[0])
		}
		sumDeg, hubs := 0, make([]string, 0, len(degree))
		type hub struct {
			node string
			deg  int
		}
		var hs []hub
		for n, d := range degree {
			sumDeg += d
			hs = append(hs, hub{n, d})
		}
		sort.Slice(hs, func(i, j int) bool {
			if hs[i].deg != hs[j].deg {
				return hs[i].deg > hs[j].deg
			}
			return hs[i].node < hs[j].node
		})
		for k := 0; k < len(hs) && k < 10; k++ {
			hubs = append(hubs, fmt.Sprintf("%s:%d", hs[k].node, hs[k].deg))
		}
		meanDeg := 0.0
		if len(degree) > 0 {
			meanDeg = float64(sumDeg) / float64(len(degree))
		}
		expected, edgeCount := tokenrepetition.IndependenceExpectedAdjacency(g, freq, len(c.Tokens))
		observedD1 := 0
		for _, d := range dists {
			if d.Distance == 1 {
				observedD1++
			}
		}
		w.editFamilies.row(c.Name, strings.Join(histParts, ","), i(len(comps)), i(largest), f4(meanDeg), strings.Join(hubs, ","), f8(expected), i(observedD1), i(edgeCount))

		for _, ch := range g.GreedyChains(minChainLength) {
			w.nearRepeatChains.row(c.Name, i(len(ch.Tokens)), strings.Join(ch.Tokens, " -> "))
		}
	} else {
		rep.note("%s is opaque (Task46/55 ciphertext); near-repetition = NOT_APPLICABLE_OPAQUE_TOKENS (task60 section 27).", c.Name)
	}

	// --- Null models for exact and near repetition (sections 9-10, 19) ---
	runNullComparisons(c, glyphSeqs, s, rng, w)

	return nil
}

func countGE(runs []tokenrepetition.Run, k int) int {
	n := 0
	for _, r := range runs {
		if r.RunLength >= k {
			n++
		}
	}
	return n
}

func writeMatrix(w *tsvWriter, corpus, direction string, m map[[2]string]int) {
	type row struct {
		a, b string
		n    int
	}
	var rows []row
	for k, n := range m {
		rows = append(rows, row{k[0], k[1], n})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].a != rows[j].a {
			return rows[i].a < rows[j].a
		}
		return rows[i].b < rows[j].b
	})
	for _, r := range rows {
		w.row(corpus, direction, r.a, r.b, i(r.n))
	}
}

func mean(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	s := 0.0
	for _, x := range v {
		s += x
	}
	return s / float64(len(v))
}
func sd(v []float64) float64 {
	if len(v) < 2 {
		return 0
	}
	m := mean(v)
	s := 0.0
	for _, x := range v {
		s += (x - m) * (x - m)
	}
	return math.Sqrt(s / float64(len(v)-1))
}
func percentile(observed float64, null []float64) float64 {
	if len(null) == 0 {
		return 0.5
	}
	le := 0
	for _, v := range null {
		if v <= observed {
			le++
		}
	}
	return float64(le) / float64(len(null))
}
