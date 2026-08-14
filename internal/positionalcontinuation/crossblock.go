package positionalcontinuation

import "sort"

// crossBlockMinAiin is the minimum in-block aiin-with-continuation count
// task23 Part K's per-block eligibility uses. Full within-block, within-
// position stratification would leave 0-1 observations per cell given how
// rare "s aiin chey" is (canonical occurrences ~4), so each block's effect is
// evaluated pooled over position (a documented, pragmatic reduction of
// section 64's "P(chey|s,aiin,position) vs P(chey|aiin,position)" to the
// coarsest replication unit that still has any support: one physical block).
const crossBlockMinAiin = 1

// buildCrossBlockPositional implements task23 Part K sections 64-68: per
// eligible physical block, the sign and enrichment of "s" over plain "aiin"
// for predicting "chey", plus how many blocks replicate the sign.
func buildCrossBlockPositional(blocks []Block, aiinOccs []AiinOccurrence) []CrossBlockPositionalRow {
	byBlock := map[string][]AiinOccurrence{}
	for _, o := range aiinOccs {
		byBlock[o.Block] = append(byBlock[o.Block], o)
	}
	meta := map[string]Block{}
	for _, b := range blocks {
		meta[b.ID] = b
	}
	ids := make([]string, 0, len(byBlock))
	for b := range byBlock {
		ids = append(ids, b)
	}
	sort.Strings(ids)

	var rows []CrossBlockPositionalRow
	for _, id := range ids {
		occs := byBlock[id]
		var aiinN, aiinChey, sN, sChey int
		for _, o := range occs {
			if o.X == "" {
				continue
			}
			aiinN++
			if o.X == FrozenChey {
				aiinChey++
			}
			if o.PredecessorIsS {
				sN++
				if o.X == FrozenChey {
					sChey++
				}
			}
		}
		if aiinN < crossBlockMinAiin {
			continue
		}
		b := meta[id]
		row := CrossBlockPositionalRow{
			Block: id, Currier: b.Currier, Hand: b.Hand, Joint: b.Joint,
			AiinOccurrences: aiinN, SAiinOccurrences: sN,
		}
		row.CheyGivenAiinPosition = float64(aiinChey) / float64(aiinN)
		if sN > 0 {
			row.CheyGivenSAiinPosition = float64(sChey) / float64(sN)
		}
		if row.CheyGivenAiinPosition > 0 {
			row.Enrichment = row.CheyGivenSAiinPosition / row.CheyGivenAiinPosition
		}
		switch {
		case sN == 0:
			row.EffectSign = "neutral"
		case row.CheyGivenSAiinPosition > row.CheyGivenAiinPosition:
			row.EffectSign = "positive"
		case row.CheyGivenSAiinPosition < row.CheyGivenAiinPosition:
			row.EffectSign = "negative"
		default:
			row.EffectSign = "neutral"
		}
		rows = append(rows, row)
	}
	return rows
}

func crossBlockSignConsistency(rows []CrossBlockPositionalRow) (eligible, positive, negative, neutral int, consistency float64) {
	for _, r := range rows {
		if r.SAiinOccurrences == 0 {
			continue
		}
		eligible++
		switch r.EffectSign {
		case "positive":
			positive++
		case "negative":
			negative++
		default:
			neutral++
		}
	}
	if eligible > 0 {
		consistency = float64(positive) / float64(eligible)
	}
	return
}
