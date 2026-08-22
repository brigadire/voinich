package inversehomophony

import "fmt"

// FitThresholdFromDevelopment computes the pair-discrimination diagnostic
// (task57 section 19) on every DEVELOPMENT corpus, pools the true/false
// score distributions across them, and returns base with Threshold set to
// the pooled Youden-J-optimal value (task57 section 6/16: fixed on
// development controls, never touched again). The pooled diagnostic is
// returned too, for the manifest/report.
func FitThresholdFromDevelopment(specs []SyntheticCorpusSpec, base Config) (Config, PairDiscrimination, error) {
	pooled := PairDiscrimination{}
	for _, spec := range specs {
		loaded, err := LoadCorpus(spec.CipherPath)
		if err != nil {
			return base, PairDiscrimination{}, fmt.Errorf("development corpus %s: %w", spec.Label, err)
		}
		oracleMapping, err := LoadOracleMapping(spec.MappingPath)
		if err != nil {
			return base, PairDiscrimination{}, fmt.Errorf("development mapping %s: %w", spec.Label, err)
		}
		oracle := oracleMapping.OraclePartitionForRelabel(loaded.Relabel)
		features := BuildFeatures(loaded.Relabel.Tokens, loaded.LineOfToken, base)
		d := DiscriminatePairs(features, oracle, base, base.Seed)
		pooled.TrueScores = append(pooled.TrueScores, d.TrueScores...)
		pooled.FalseScores = append(pooled.FalseScores, d.FalseScores...)
		pooled.TruePairs += d.TruePairs
		pooled.FalsePairs += d.FalsePairs
	}
	pooled.AUC = auc(pooled.TrueScores, pooled.FalseScores)
	base.Threshold = FreezeThreshold(pooled)
	return base, pooled, nil
}
