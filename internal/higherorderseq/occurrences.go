package higherorderseq

// findOccurrences implements task22 Part A: every exact in-block occurrence
// of a candidate's A B C tokens. Because a Block is already a maximal
// contiguous run of one joint metadata class, scanning consecutive indices
// within a single block's Tokens slice can never cross a physical-block
// boundary (section 6). Cross-line occurrences are recorded, not excluded
// (section 7), to stay compatible with the continuous-corpus sequence
// analysis the candidates were originally discovered under.
func findOccurrences(cand Candidate, blocks []Block, lineLength map[string]int) []Occurrence {
	a, b, c := cand.A(), cand.B(), cand.C()
	var out []Occurrence
	for _, blk := range blocks {
		n := len(blk.Tokens)
		for k := 0; k+2 < n; k++ {
			if blk.Tokens[k].Text != a || blk.Tokens[k+1].Text != b || blk.Tokens[k+2].Text != c {
				continue
			}
			t0, t1, t2 := blk.Tokens[k], blk.Tokens[k+1], blk.Tokens[k+2]
			sameLine := t0.Line == t1.Line && t1.Line == t2.Line
			occ := Occurrence{
				Sequence: cand.Sequence,
				PosA:     t0.Position, PosB: t1.Position, PosC: t2.Position,
				Block: blk.ID, Currier: blk.Currier, Hand: blk.Hand, Joint: blk.Joint,
				NormalizedBlockPos:  normalizedPosition(k, n),
				WithinSameLine:      sameLine,
				CrossesLineBoundary: !sameLine,
			}
			if sameLine {
				occ.LinePosition = linePosition(t0.TokenIndexLine, lineLength[t0.Line])
			}
			out = append(out, occ)
		}
	}
	return out
}

func normalizedPosition(index, blockLen int) float64 {
	if blockLen <= 1 {
		return 0
	}
	return float64(index) / float64(blockLen-1)
}

func linePosition(indexInLine, length int) string {
	if length <= 1 {
		return "start"
	}
	switch {
	case indexInLine <= 0:
		return "start"
	case indexInLine >= length-1:
		return "end"
	default:
		return "middle"
	}
}

// positionBucket maps a normalized position in [0,1] to one of the ten
// fixed [0,0.1)...[0.9,1.0] buckets task22 section 56 defines.
func positionBucket(p float64) string {
	buckets := []string{"[0,0.1)", "[0.1,0.2)", "[0.2,0.3)", "[0.3,0.4)", "[0.4,0.5)", "[0.5,0.6)", "[0.6,0.7)", "[0.7,0.8)", "[0.8,0.9)", "[0.9,1.0]"}
	idx := int(p * 10)
	if idx >= 10 {
		idx = 9
	}
	if idx < 0 {
		idx = 0
	}
	return buckets[idx]
}
