package positionalcontinuation

func presentXAiin(occs []AiinOccurrence) []string {
	var xs []string
	for _, o := range occs {
		if o.X != "" {
			xs = append(xs, o.X)
		}
	}
	return xs
}

func filterAiinByLineCategory(occs []AiinOccurrence, cat string) []AiinOccurrence {
	var out []AiinOccurrence
	for _, o := range occs {
		if o.LineCategory == cat {
			out = append(out, o)
		}
	}
	return out
}

func filterAiinByBlockCoarse(occs []AiinOccurrence, cat string) []AiinOccurrence {
	var out []AiinOccurrence
	for _, o := range occs {
		if o.BlockBinCoarse == cat {
			out = append(out, o)
		}
	}
	return out
}

// buildAiinControlRows implements task23 Part H sections 42-47: the same
// positional stratification repeated for "aiin" alone (no predecessor
// requirement), joined against the "s aiin" chey-effect rows already computed
// by Parts E-G so the within-position enrichment E(position) =
// P(chey|s,aiin,position) / P(chey|aiin,position) can be reported directly.
func buildAiinControlRows(aiinOccs []AiinOccurrence, sAiinCheyByVariable map[string][]CheyEffectRow) []AiinControlRow {
	var rows []AiinControlRow
	variables := []struct {
		name       string
		categories []string
		filter     func([]AiinOccurrence, string) []AiinOccurrence
	}{
		{"line_position", lineCategories, filterAiinByLineCategory},
		{"block_position_coarse", blockCoarseCategories, filterAiinByBlockCoarse},
	}
	for _, v := range variables {
		sAiinByStratum := map[string]CheyEffectRow{}
		for _, ce := range sAiinCheyByVariable[v.name] {
			sAiinByStratum[ce.Stratum] = ce
		}
		for _, cat := range v.categories {
			xs := presentXAiin(v.filter(aiinOccs, cat))
			counts := countMap(xs)
			row := AiinControlRow{
				PositionVariable: v.name, Stratum: cat, AiinOccurrenceCount: len(xs),
				AiinEntropyBits: countEntropyBits(counts), AiinUniqueContinuations: len(counts),
				CheyCount: counts[FrozenChey],
			}
			if len(xs) > 0 {
				row.PCheyGivenAiinPosition = float64(counts[FrozenChey]) / float64(len(xs))
			}
			row.PCheyGivenSAiinPosition = sAiinByStratum[cat].PCheyGivenPosition
			if row.PCheyGivenAiinPosition > 0 {
				row.WithinPositionEnrichment = row.PCheyGivenSAiinPosition / row.PCheyGivenAiinPosition
			}
			rows = append(rows, row)
		}
	}
	return rows
}
