package main

import (
	"math/rand"
	"strings"

	"zcore.dev/voinich/internal/tokenrepetition"
)

// runLabels implements task60 sections 3-4, 29-34: extract illustration
// labels from the raw IVTFF source, summarize their repetition, compare
// them against matched-size running-text samples, and compute vocabulary
// overlap.
func runLabels(voynich tokenrepetition.Corpus, outDir string, rng *rand.Rand, rep *report) error {
	labels, err := tokenrepetition.ExtractLabels(ivtffPath)
	if err != nil {
		return err
	}
	lc := newTSV(outDir, "LABEL_CORPUS.tsv", "Locus", "LabelID", "RawLabel", "ParsedTokens")
	for idx, l := range labels {
		lc.row(l.Locus, i(idx), l.RawLabel, strings.Join(l.Tokens, " "))
	}
	lc.close()

	labelCorpus := tokenrepetition.LabelCorpus(labels)
	stats := tokenrepetition.SummarizeLabelRepetition(labels)
	rep.labelValidPairs = tokenrepetition.AdjacentRepetition(labelCorpus.Tokens, labelCorpus.LineOfToken, 0).ValidPairs
	rr := newTSV(outDir, "LABEL_REPETITION.tsv",
		"Instances", "UniqueCompleteLabels", "RepeatedCompleteLabels", "HapaxLabelFraction",
		"TokenN", "TokenV", "TokenHapaxFraction", "RepeatedTokenTypes")
	rr.row(i(stats.Instances), i(stats.UniqueCompleteLabels), i(stats.RepeatedCompleteLabels), f8(stats.HapaxLabelFraction),
		i(stats.TokenN), i(stats.TokenV), f8(stats.TokenHapaxFraction), i(stats.RepeatedTokenTypes))
	rr.close()

	labelSummary := tokenrepetition.SummarizeLabels(labelCorpus, tokenrepetition.GlyphVoynich)
	null := tokenrepetition.RunningTextSubsampleNull(voynich.Tokens, voynich.LineOfToken, stats.TokenN, labelSubsamples, tokenrepetition.GlyphVoynich, rng)
	var vs, hs, rs, es []float64
	for _, n := range null {
		vs = append(vs, float64(n.V))
		hs = append(hs, n.HapaxFraction)
		rs = append(rs, n.AdjacentRepeatRate)
		es = append(es, n.EditLe1Rate)
	}
	lrc := newTSV(outDir, "LABEL_RUNNING_COMPARISON.tsv",
		"Statistic", "LabelValue", "RunningSampleMean", "RunningSampleSD", "Percentile", "Draws")
	lrc.row("V", f8(float64(labelSummary.V)), f8(mean(vs)), f8(sd(vs)), f4(percentile(float64(labelSummary.V), vs)), i(labelSubsamples))
	lrc.row("HapaxFraction", f8(labelSummary.HapaxFraction), f8(mean(hs)), f8(sd(hs)), f4(percentile(labelSummary.HapaxFraction, hs)), i(labelSubsamples))
	lrc.row("AdjacentRepeatRate", f8(labelSummary.AdjacentRepeatRate), f8(mean(rs)), f8(sd(rs)), f4(percentile(labelSummary.AdjacentRepeatRate, rs)), i(labelSubsamples))
	lrc.row("EditLe1Rate", f8(labelSummary.EditLe1Rate), f8(mean(es)), f8(sd(es)), f4(percentile(labelSummary.EditLe1Rate, es)), i(labelSubsamples))
	lrc.close()

	overlap := tokenrepetition.ComputeVocabularyOverlap(labelCorpus.Tokens, voynich.Tokens)
	lvo := newTSV(outDir, "LABEL_VOCABULARY_OVERLAP.tsv",
		"LabelTypes", "RunningTypes", "SharedTypes", "LabelTypeCoverage", "LabelOccurrenceCoverage", "Jaccard")
	lvo.row(i(overlap.LabelTypes), i(overlap.RunningTypes), i(overlap.SharedTypes), f8(overlap.LabelTypeCoverage), f8(overlap.LabelOccurrenceCoverage), f8(overlap.Jaccard))
	lvo.close()

	return nil
}

