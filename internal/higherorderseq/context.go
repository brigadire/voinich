package higherorderseq

import "sort"

// contextAltMinCount is the frequency floor task22 sections 72 and 32 use
// for "sufficiently frequent" alternative left contexts: count(X,B) >= 3.
const contextAltMinCount = 3

// leftContextCounts pools, across every physical block (no eligibility
// filter - these are descriptive controls over all available data), how
// often each left token X immediately precedes the candidate's B, and how
// often X,B is itself immediately followed by C.
func leftContextCounts(cand Candidate, blocks []Block) (leftB, leftBC map[string]int, totalB, totalBC int) {
	b, c := cand.B(), cand.C()
	leftB, leftBC = map[string]int{}, map[string]int{}
	for _, blk := range blocks {
		n := len(blk.Tokens)
		for k := 0; k+1 < n; k++ {
			if blk.Tokens[k+1].Text != b {
				continue
			}
			x := blk.Tokens[k].Text
			leftB[x]++
			totalB++
			if k+2 < n && blk.Tokens[k+2].Text == c {
				leftBC[x]++
				totalBC++
			}
		}
	}
	return
}

// rightContextCounts pools how often each token X immediately follows B
// (context "B"), and how often X immediately follows the frozen A,B bigram
// specifically (context "A,B").
func rightContextCounts(cand Candidate, blocks []Block) (rightB, rightAB map[string]int, totalRightB, totalAB int) {
	a, b := cand.A(), cand.B()
	rightB, rightAB = map[string]int{}, map[string]int{}
	for _, blk := range blocks {
		n := len(blk.Tokens)
		for k := 0; k+1 < n; k++ {
			if blk.Tokens[k].Text == b {
				x := blk.Tokens[k+1].Text
				rightB[x]++
				totalRightB++
			}
		}
		for k := 0; k+2 < n; k++ {
			if blk.Tokens[k].Text == a && blk.Tokens[k+1].Text == b {
				x := blk.Tokens[k+2].Text
				rightAB[x]++
				totalAB++
			}
		}
	}
	return
}

// contextControlRows implements task22 Part F sections 30-34: for the
// candidate's B, the full left-token/right-token/count table, comparing the
// frozen A,B,C occurrence against every other sufficiently frequent
// alternative on both sides.
func contextControlRows(cand Candidate, blocks []Block) []ContextControlRow {
	leftB, leftBC, _, _ := leftContextCounts(cand, blocks)
	_, rightAB, _, totalAB := rightContextCounts(cand, blocks)
	var rows []ContextControlRow
	for _, x := range stringKeysInt(leftB) {
		if leftB[x] < contextAltMinCount {
			continue
		}
		rows = append(rows, ContextControlRow{
			Sequence: cand.Sequence, ContextType: "left_alt", AltToken: x, Count: leftB[x],
			Probability: float64(leftBC[x]) / float64(leftB[x]), IsFrozen: x == cand.A(),
		})
	}
	for _, x := range stringKeysInt(rightAB) {
		p := 0.0
		if totalAB > 0 {
			p = float64(rightAB[x]) / float64(totalAB)
		}
		rows = append(rows, ContextControlRow{
			Sequence: cand.Sequence, ContextType: "right_alt", AltToken: x, Count: rightAB[x],
			Probability: p, IsFrozen: x == cand.C(),
		})
	}
	return rows
}

// contextRankRow implements task22 Part O sections 72-77: the frozen AB
// context is ranked, purely descriptively, among every other sufficiently
// frequent X,B context for the same central B, by P(C|X,B) - answering
// whether "s aiin" is unusual among all "X aiin" contexts, not merely
// unusual relative to the whole corpus.
func contextRankRow(cand Candidate, blocks []Block) ContextRankRow {
	leftB, leftBC, totalB, totalBC := leftContextCounts(cand, blocks)
	var ps []float64
	frozenP := 0.0
	for _, x := range stringKeysInt(leftB) {
		if leftB[x] < contextAltMinCount {
			continue
		}
		p := float64(leftBC[x]) / float64(leftB[x])
		ps = append(ps, p)
		if x == cand.A() {
			frozenP = p
		}
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(ps)))
	rank := len(ps)
	for i, p := range ps {
		if p <= frozenP {
			rank = i + 1
			break
		}
	}
	row := ContextRankRow{Sequence: cand.Sequence, NumAlternatives: len(ps), FrozenP: frozenP, Rank: rank}
	if totalB > 0 {
		row.BaselineP = float64(totalBC) / float64(totalB)
	}
	if len(ps) > 0 {
		row.Percentile = 100 * float64(len(ps)-rank+1) / float64(len(ps))
		row.MinAltP, row.MaxAltP = ps[len(ps)-1], ps[0]
		row.MedianAltP = median(ps)
	}
	return row
}

// continuationDistributions and continuationEntropy implement task22 Part G
// sections 35-41: the full continuation distribution after B alone versus
// after the frozen A,B, and how much knowing A reduces the entropy of what
// comes next.
func continuationDistributions(cand Candidate, blocks []Block) []ContinuationRow {
	rightB, rightAB, totalRightB, totalAB := rightContextCounts(cand, blocks)
	var rows []ContinuationRow
	for _, x := range stringKeysInt(rightB) {
		rows = append(rows, ContinuationRow{Sequence: cand.Sequence, Context: "B", Token: x, Count: rightB[x], Probability: float64(rightB[x]) / float64(totalRightB)})
	}
	for _, x := range stringKeysInt(rightAB) {
		rows = append(rows, ContinuationRow{Sequence: cand.Sequence, Context: "AB", Token: x, Count: rightAB[x], Probability: float64(rightAB[x]) / float64(totalAB)})
	}
	return rows
}

func continuationEntropy(cand Candidate, blocks []Block) ContinuationEntropyRow {
	rightB, rightAB, _, _ := rightContextCounts(cand, blocks)
	pB := toProbabilities(rightB)
	pAB := toProbabilities(rightAB)
	hB := entropyBits(pB)
	hAB := entropyBits(pAB)
	return ContinuationEntropyRow{
		Sequence: cand.Sequence, HGivenB: hB, HGivenAB: hAB, EntropyReduction: hB - hAB,
		JSDivergence: jsDivergenceBits(pB, pAB), TotalVariation: totalVariation(pB, pAB),
	}
}

func stringKeysInt(m map[string]int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
