package main

import (
	"math/rand"

	"zcore.dev/voinich/internal/tokenrepetition"
)

// runNullComparisons computes both null models (task60 sections 9-10),
// for both the exact-repetition rate R2 (nullPermutations draws) and the
// near-repetition rate P(d<=1) (nearNullPermutations draws, a smaller
// count fixed in METHOD.md given the added per-permutation edit-distance
// cost), and records z-scores/percentiles into the corpus summary plus
// the NULL_EXACT_REPETITION.tsv / NULL_NEAR_REPETITION.tsv rows.
func runNullComparisons(c tokenrepetition.Corpus, glyphSeqs map[string][]string, s *corpusSummary, rng *rand.Rand, w *writers) {
	// Exact repetition nulls.
	globalR2 := make([]float64, 0, nullPermutations)
	lineR2 := make([]float64, 0, nullPermutations)
	for k := 0; k < nullPermutations; k++ {
		shuffled := tokenrepetition.GlobalShuffle(c.Tokens, rng)
		globalR2 = append(globalR2, tokenrepetition.AdjacentRepetition(shuffled, nil, 0).R2)
	}
	for k := 0; k < nullPermutations; k++ {
		shuffled := tokenrepetition.WithinLineShuffle(c.Tokens, c.LineOfToken, rng)
		lineR2 = append(lineR2, tokenrepetition.AdjacentRepetition(shuffled, c.LineOfToken, 0).R2)
	}
	s.R2NullMeanGlobal, s.R2NullSDGlobal = mean(globalR2), sd(globalR2)
	if s.R2NullSDGlobal > 0 {
		s.R2ZGlobal = (s.R2 - s.R2NullMeanGlobal) / s.R2NullSDGlobal
	}
	s.R2PercentileGlobal = percentile(s.R2, globalR2)
	s.R2NullMeanLines, s.R2NullSDLines = mean(lineR2), sd(lineR2)
	if s.R2NullSDLines > 0 {
		s.R2ZLines = (s.R2 - s.R2NullMeanLines) / s.R2NullSDLines
	}
	s.R2PercentileLines = percentile(s.R2, lineR2)
	w.nullExact.row(c.Name, "global_shuffle", i(nullPermutations), f8(s.R2), f8(s.R2NullMeanGlobal), f8(s.R2NullSDGlobal), f4(s.R2ZGlobal), f4(s.R2PercentileGlobal))
	w.nullExact.row(c.Name, "within_line_shuffle", i(nullPermutations), f8(s.R2), f8(s.R2NullMeanLines), f8(s.R2NullSDLines), f4(s.R2ZLines), f4(s.R2PercentileLines))

	if c.Opaque {
		return
	}
	globalPLe1 := make([]float64, 0, nearNullPermutations)
	linePLe1 := make([]float64, 0, nearNullPermutations)
	for k := 0; k < nearNullPermutations; k++ {
		shuffled := tokenrepetition.GlobalShuffle(c.Tokens, rng)
		d := tokenrepetition.AdjacentEditDistances(shuffled, nil, glyphSeqs)
		globalPLe1 = append(globalPLe1, tokenrepetition.SummarizeDistances(d, glyphSeqs).PLe1)
	}
	for k := 0; k < nearNullPermutations; k++ {
		shuffled := tokenrepetition.WithinLineShuffle(c.Tokens, c.LineOfToken, rng)
		d := tokenrepetition.AdjacentEditDistances(shuffled, c.LineOfToken, glyphSeqs)
		linePLe1 = append(linePLe1, tokenrepetition.SummarizeDistances(d, glyphSeqs).PLe1)
	}
	s.PLe1NullMeanGlobal = mean(globalPLe1)
	if sdv := sd(globalPLe1); sdv > 0 {
		s.PLe1ZGlobal = (s.PLe1 - s.PLe1NullMeanGlobal) / sdv
	}
	s.PLe1PercentileGlobal = percentile(s.PLe1, globalPLe1)
	s.PLe1NullMeanLines = mean(linePLe1)
	if sdv := sd(linePLe1); sdv > 0 {
		s.PLe1ZLines = (s.PLe1 - s.PLe1NullMeanLines) / sdv
	}
	s.PLe1PercentileLines = percentile(s.PLe1, linePLe1)
	w.nullNear.row(c.Name, "global_shuffle", i(nearNullPermutations), f8(s.PLe1), f8(s.PLe1NullMeanGlobal), f8(sd(globalPLe1)), f4(s.PLe1ZGlobal), f4(s.PLe1PercentileGlobal))
	w.nullNear.row(c.Name, "within_line_shuffle", i(nearNullPermutations), f8(s.PLe1), f8(s.PLe1NullMeanLines), f8(sd(linePLe1)), f4(s.PLe1ZLines), f4(s.PLe1PercentileLines))
}
