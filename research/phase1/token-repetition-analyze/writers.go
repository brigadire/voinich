package main

// writers bundles every output TSV so one row per (corpus, ...) can be
// appended as each corpus is analyzed, rather than writing N separate
// per-corpus files for what task60 section 44 names as a single artifact.
// Every TSV below therefore starts with a "Corpus" column.
type writers struct {
	exactAdjacent   *tsvWriter
	exactRuns       *tsvWriter
	runDistribution *tsvWriter
	editDistDist    *tsvWriter
	editDistOne     *tsvWriter
	editOpPosition  *tsvWriter
	substMatrix     *tsvWriter
	editFamilies    *tsvWriter
	nearRepeatChains *tsvWriter
	nullExact       *tsvWriter
	nullNear        *tsvWriter
	homDoseResponse *tsvWriter
	homTheoretical  *tsvWriter
}

func openWriters(dir string) *writers {
	return &writers{
		exactAdjacent: newTSV(dir, "EXACT_ADJACENT_REPETITION.tsv",
			"Corpus", "Token", "Frequency", "AdjacentRepeatCount", "MaximumRun", "Loci"),
		exactRuns: newTSV(dir, "EXACT_RUNS.tsv",
			"Corpus", "Token", "RunLength", "StartPosition", "GlobalFrequency"),
		runDistribution: newTSV(dir, "RUN_DISTRIBUTION.tsv",
			"Corpus", "K", "CountRunsGEK", "FractionOfTokensGEK"),
		editDistDist: newTSV(dir, "EDIT_DISTANCE_DISTRIBUTION.tsv",
			"Corpus", "PEq0", "PEq1", "PLe1", "PLe2", "PEq1GivenSameLength", "SubstitutionOnlyRate", "TotalPairs"),
		editDistOne: newTSV(dir, "EDIT_DISTANCE_ONE.tsv",
			"Corpus", "Position", "TokenA", "TokenB", "Operation", "PositionClass", "SourceGlyph", "TargetGlyph"),
		editOpPosition: newTSV(dir, "EDIT_OPERATION_POSITION.tsv",
			"Corpus", "Operation", "PositionClass", "Count"),
		substMatrix: newTSV(dir, "SUBSTITUTION_MATRIX.tsv",
			"Corpus", "Direction", "SourceGlyph", "TargetGlyph", "Count"),
		editFamilies: newTSV(dir, "EDIT_FAMILIES.tsv",
			"Corpus", "ComponentSizeHistogram", "ComponentCount", "LargestComponent", "MeanDegree", "TopHubs", "ExpectedIndependentAdjacency", "ObservedD1Adjacency", "EdgeCount"),
		nearRepeatChains: newTSV(dir, "NEAR_REPEAT_CHAINS.tsv",
			"Corpus", "ChainLength", "Chain"),
		nullExact: newTSV(dir, "NULL_EXACT_REPETITION.tsv",
			"Corpus", "NullModel", "Permutations", "ObservedR2", "NullMeanR2", "NullSDR2", "Z", "Percentile"),
		nullNear: newTSV(dir, "NULL_NEAR_REPETITION.tsv",
			"Corpus", "NullModel", "Permutations", "ObservedPLe1", "NullMeanPLe1", "NullSDPLe1", "Z", "Percentile"),
		homDoseResponse: newTSV(dir, "HOMOPHONY_RUN_DOSE_RESPONSE.tsv",
			"Series", "H", "Model", "R2", "Runs3", "Runs4", "Runs5", "MaxRun"),
		homTheoretical: newTSV(dir, "HOMOPHONY_THEORETICAL_VS_OBSERVED.tsv",
			"Series", "H", "Model", "RunLength", "Count", "MeanPredictedSurvival", "ObservedSurvivalFraction"),
	}
}

func (w *writers) closeAll() {
	for _, x := range []*tsvWriter{w.exactAdjacent, w.exactRuns, w.runDistribution, w.editDistDist, w.editDistOne,
		w.editOpPosition, w.substMatrix, w.editFamilies, w.nearRepeatChains, w.nullExact, w.nullNear,
		w.homDoseResponse, w.homTheoretical} {
		x.close()
	}
}
