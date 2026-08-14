package positionalcontinuation

// findSAiinOccurrences implements task23 Part A: every exact in-block
// occurrence of "s aiin" together with the following token X (if any).
// Because a Block is already a maximal contiguous run of one joint metadata
// class, scanning consecutive indices within a single block's Tokens slice
// can never cross a physical-block boundary (section 5). Boundary
// occurrences (X missing) are recorded, never silently dropped (section 7).
func findSAiinOccurrences(blocks []Block, lineLength map[string]int, totalTokens int) []SAiinOccurrence {
	var out []SAiinOccurrence
	for _, blk := range blocks {
		n := len(blk.Tokens)
		for k := 0; k+1 < n; k++ {
			if blk.Tokens[k].Text != FrozenS || blk.Tokens[k+1].Text != FrozenAiin {
				continue
			}
			s, aiin := blk.Tokens[k], blk.Tokens[k+1]
			occ := SAiinOccurrence{
				PosS: s.Position, PosAiin: aiin.Position, PosX: -1,
				Block: blk.ID, Currier: blk.Currier, Hand: blk.Hand, Joint: blk.Joint,
				NormalizedBlockPosition: normalizedPosition(k+1, n),
			}
			occ.BlockBinFixed = blockBinFixed(occ.NormalizedBlockPosition)
			occ.BlockBinCoarse = blockBinCoarse(occ.NormalizedBlockPosition)
			occ.TokensFromBlockStart = k + 1
			occ.TokensToBlockEnd = n - 1 - (k + 1)

			occ.LineID = aiin.Line
			lineLen := lineLength[aiin.Line]
			occ.NormalizedLinePosition = normalizedPosition(aiin.TokenIndexLine, lineLen)
			occ.TokensFromLineStart = aiin.TokenIndexLine
			occ.TokensToLineEnd = lineLen - 1 - aiin.TokenIndexLine
			occ.SIsLineStart = s.TokenIndexLine == 0

			xEndsLine := false
			if k+2 < n {
				x := blk.Tokens[k+2]
				occ.PosX = x.Position
				occ.X = x.Text
				xEndsLine = x.TokenIndexLine == lineLength[x.Line]-1
				for j := 0; j < 3; j++ {
					if k+3+j < n {
						occ.TokensAfter[j] = blk.Tokens[k+3+j].Text
					}
				}
			} else {
				xEndsLine = aiin.TokenIndexLine == lineLen-1
				if aiin.Position+1 >= totalTokens {
					occ.XMissingCorpusEnd = true
				} else {
					occ.XMissingBlockEnd = true
				}
			}
			occ.XIsLineEnd = xEndsLine
			occ.LineCategory = lineCategory(occ.SIsLineStart, occ.XIsLineEnd, occ.NormalizedLinePosition)

			for j := 0; j < 3; j++ {
				idx := k - 1 - j
				if idx >= 0 {
					occ.TokensBefore[j] = blk.Tokens[idx].Text
				}
			}
			out = append(out, occ)
		}
	}
	return out
}

// findAiinOccurrences implements task23 Part H's control construction:
// every exact in-block occurrence of "aiin" alone, whatever precedes it
// (including nothing, at a block start).
func findAiinOccurrences(blocks []Block, lineLength map[string]int, totalTokens int) []AiinOccurrence {
	var out []AiinOccurrence
	for _, blk := range blocks {
		n := len(blk.Tokens)
		for k := 0; k < n; k++ {
			if blk.Tokens[k].Text != FrozenAiin {
				continue
			}
			aiin := blk.Tokens[k]
			occ := AiinOccurrence{
				PosAiin: aiin.Position, PosX: -1,
				Block: blk.ID, Currier: blk.Currier, Hand: blk.Hand, Joint: blk.Joint,
				NormalizedBlockPosition: normalizedPosition(k, n),
			}
			occ.BlockBinFixed = blockBinFixed(occ.NormalizedBlockPosition)
			occ.BlockBinCoarse = blockBinCoarse(occ.NormalizedBlockPosition)
			occ.LineID = aiin.Line
			lineLen := lineLength[aiin.Line]
			occ.NormalizedLinePosition = normalizedPosition(aiin.TokenIndexLine, lineLen)

			contextStartsLine := aiin.TokenIndexLine == 0
			if k > 0 {
				pred := blk.Tokens[k-1]
				occ.Predecessor = pred.Text
				occ.HasPredecessor = true
				occ.PredecessorIsS = pred.Text == FrozenS
				contextStartsLine = pred.TokenIndexLine == 0
			}

			xEndsLine := false
			if k+1 < n {
				x := blk.Tokens[k+1]
				occ.PosX = x.Position
				occ.X = x.Text
				xEndsLine = x.TokenIndexLine == lineLength[x.Line]-1
			} else {
				xEndsLine = aiin.TokenIndexLine == lineLen-1
				if aiin.Position+1 >= totalTokens {
					occ.XMissingCorpusEnd = true
				} else {
					occ.XMissingBlockEnd = true
				}
			}
			occ.LineCategory = lineCategory(contextStartsLine, xEndsLine, occ.NormalizedLinePosition)
			out = append(out, occ)
		}
	}
	return out
}
