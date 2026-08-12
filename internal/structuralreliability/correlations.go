package structuralreliability

// buildFrequencyVsTokenStability answers research question 1/3/4 (section
// 32): is the frequency/stability relationship systematic, and does it
// differ between position and left/right context? Spearman is invariant
// under the log2 transform, so this is equivalent to correlating raw count,
// but log2(count) is kept to match the task's own notation.
func buildFrequencyVsTokenStability(tokens []tokenMetric) []Correlation {
	type series struct {
		name string
		get  func(tokenMetric) (float64, bool)
	}
	all := []series{
		{"log2_count_vs_position_train_train_stability", func(t tokenMetric) (float64, bool) {
			return t.PositionTrainTrainStability, t.TrainTrainObservations > 0
		}},
		{"log2_count_vs_left_train_train_stability", func(t tokenMetric) (float64, bool) { return t.LeftTrainTrainStability, t.TrainTrainObservations > 0 }},
		{"log2_count_vs_right_train_train_stability", func(t tokenMetric) (float64, bool) { return t.RightTrainTrainStability, t.TrainTrainObservations > 0 }},
		{"log2_count_vs_position_train_test_stability", func(t tokenMetric) (float64, bool) { return t.PositionTrainTestStability, t.TrainTestObservations > 0 }},
		{"log2_count_vs_left_train_test_stability", func(t tokenMetric) (float64, bool) { return t.LeftTrainTestStability, t.TrainTestObservations > 0 }},
		{"log2_count_vs_right_train_test_stability", func(t tokenMetric) (float64, bool) { return t.RightTrainTestStability, t.TrainTestObservations > 0 }},
		{"log2_count_vs_top1_recovery", func(t tokenMetric) (float64, bool) { return t.Top1Recovery, t.Top1RecoveryObservations > 0 }},
		{"log2_count_vs_top10_jaccard", func(t tokenMetric) (float64, bool) { return t.Top10Jaccard, t.Top10JaccardObservations > 0 }},
	}
	result := make([]Correlation, 0, len(all))
	for _, entry := range all {
		var x, y []float64
		for _, token := range tokens {
			value, ok := entry.get(token)
			if !ok {
				continue
			}
			x = append(x, Log2(float64(token.FullCount)))
			y = append(y, value)
		}
		rho, n := Spearman(x, y)
		result = append(result, Correlation{Metric: entry.name, Rho: rho, Observations: n})
	}
	return result
}

// buildFrequencyVsPairStability answers section 11: does pair-level sample
// size (the weaker of the two members, or their geometric mean) explain
// pairwise instability the way single-token frequency explains self-profile
// instability?
func buildFrequencyVsPairStability(pairs []ContinuousPairMetric) []Correlation {
	var minCountVsStddevX, minCountVsStddevY []float64
	var minCountVsCIWidthX, minCountVsCIWidthY []float64
	var geoMeanVsStddevX, geoMeanVsStddevY []float64
	for _, pair := range pairs {
		if pair.FoldSimilarityStddev != nil {
			minCountVsStddevX = append(minCountVsStddevX, Log2(float64(pair.MinCount)))
			minCountVsStddevY = append(minCountVsStddevY, *pair.FoldSimilarityStddev)
			geoMeanVsStddevX = append(geoMeanVsStddevX, Log2(pair.GeometricMeanCount))
			geoMeanVsStddevY = append(geoMeanVsStddevY, *pair.FoldSimilarityStddev)
		}
		if pair.BootstrapCIWidth != nil {
			minCountVsCIWidthX = append(minCountVsCIWidthX, Log2(float64(pair.MinCount)))
			minCountVsCIWidthY = append(minCountVsCIWidthY, *pair.BootstrapCIWidth)
		}
	}
	rho1, n1 := Spearman(minCountVsStddevX, minCountVsStddevY)
	rho2, n2 := Spearman(minCountVsCIWidthX, minCountVsCIWidthY)
	rho3, n3 := Spearman(geoMeanVsStddevX, geoMeanVsStddevY)
	return []Correlation{
		{Metric: "log2_min_count_vs_similarity_stddev", Rho: rho1, Observations: n1},
		{Metric: "log2_min_count_vs_bootstrap_ci_width", Rho: rho2, Observations: n2},
		{Metric: "log2_geometric_mean_count_vs_similarity_stddev", Rho: rho3, Observations: n3},
	}
}

// buildContextDiversityVsStability answers section 23/24's research question
// 5: does context diversity (or count divided by it) explain residual
// instability better than raw token count?
func buildContextDiversityVsStability(diversity []TokenDiversity, byToken map[string]tokenMetric) []Correlation {
	type series struct {
		name string
		x    func(TokenDiversity) float64
	}
	entries := []series{
		{"unique_predecessors_vs_left_train_train_stability", func(d TokenDiversity) float64 { return float64(d.UniquePredecessors) }},
		{"unique_successors_vs_right_train_train_stability", func(d TokenDiversity) float64 { return float64(d.UniqueSuccessors) }},
		{"left_entropy_vs_left_train_train_stability", func(d TokenDiversity) float64 { return d.LeftEntropy }},
		{"right_entropy_vs_right_train_train_stability", func(d TokenDiversity) float64 { return d.RightEntropy }},
		{"effective_left_observations_vs_left_train_train_stability", func(d TokenDiversity) float64 { return d.EffectiveLeftObservations }},
		{"effective_right_observations_vs_right_train_train_stability", func(d TokenDiversity) float64 { return d.EffectiveRightObservations }},
	}
	useRight := map[string]bool{
		"unique_successors_vs_right_train_train_stability":            true,
		"right_entropy_vs_right_train_train_stability":                true,
		"effective_right_observations_vs_right_train_train_stability": true,
	}
	result := make([]Correlation, 0, len(entries))
	for _, entry := range entries {
		var x, y []float64
		for _, diversity := range diversity {
			token, ok := byToken[diversity.Token]
			if !ok || token.TrainTrainObservations == 0 {
				continue
			}
			x = append(x, entry.x(diversity))
			if useRight[entry.name] {
				y = append(y, token.RightTrainTrainStability)
			} else {
				y = append(y, token.LeftTrainTrainStability)
			}
		}
		rho, n := Spearman(x, y)
		result = append(result, Correlation{Metric: entry.name, Rho: rho, Observations: n})
	}
	return result
}
