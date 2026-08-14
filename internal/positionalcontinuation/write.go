package positionalcontinuation

import (
	"encoding/csv"
	"math"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

func f(v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return "NA"
	}
	return strconv.FormatFloat(v, 'g', -1, 64)
}
func i(v int) string   { return strconv.Itoa(v) }
func bo(v bool) string { return strconv.FormatBool(v) }

func writeTSV(path string, head []string, rows [][]string) error {
	file, e := os.Create(path)
	if e != nil {
		return e
	}
	w := csv.NewWriter(file)
	w.Comma = '\t'
	if e = w.Write(head); e == nil {
		for _, r := range rows {
			if e = w.Write(r); e != nil {
				break
			}
		}
	}
	w.Flush()
	if e == nil {
		e = w.Error()
	}
	if ce := file.Close(); e == nil {
		e = ce
	}
	return e
}

func xOrMissing(x string) string {
	if x == "" {
		return ""
	}
	return x
}

func writeAll(c Config, meta runMeta, state *RunState) error {
	if err := os.MkdirAll(c.OutputDir, 0755); err != nil {
		return err
	}

	var rows [][]string
	for _, o := range state.SAiinOccurrences {
		missing := ""
		switch {
		case o.XMissingBlockEnd:
			missing = "block_end"
		case o.XMissingCorpusEnd:
			missing = "corpus_end"
		}
		rows = append(rows, []string{
			i(o.PosS), i(o.PosAiin), i(o.PosX), xOrMissing(o.X), missing,
			o.Block, o.Currier, o.Hand, o.Joint,
			f(o.NormalizedBlockPosition), o.BlockBinFixed, o.BlockBinCoarse,
			o.LineID, f(o.NormalizedLinePosition), o.LineCategory, bo(o.SIsLineStart), bo(o.XIsLineEnd),
			o.TokensBefore[2], o.TokensBefore[1], o.TokensBefore[0],
			o.TokensAfter[0], o.TokensAfter[1], o.TokensAfter[2],
			i(o.TokensFromLineStart), i(o.TokensToLineEnd), i(o.TokensFromBlockStart), i(o.TokensToBlockEnd),
		})
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "sai_in_context_occurrences.tsv"),
		[]string{"position_s", "position_aiin", "position_x", "x", "x_missing_reason",
			"physical_block", "currier", "hand", "joint_class",
			"normalized_block_position", "block_bin_fixed", "block_bin_coarse",
			"line_id", "normalized_line_position", "line_category", "s_is_line_start", "x_is_line_end",
			"token_before_3", "token_before_2", "token_before_1",
			"token_after_1", "token_after_2", "token_after_3",
			"tokens_from_line_start", "tokens_to_line_end", "tokens_from_block_start", "tokens_to_block_end"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, r := range state.ContinuationRows {
		rows = append(rows, []string{r.Context, r.Stratum, r.StratumType, r.Token, i(r.Count), f(r.Probability)})
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "sai_in_continuation_distributions.tsv"),
		[]string{"context", "stratum", "stratum_type", "token", "count", "probability"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, s := range state.DistSummaryRows {
		rows = append(rows, []string{s.Context, s.Stratum, s.StratumType, i(s.OccurrenceCount), i(s.UniqueContinuations),
			f(s.EntropyBits), f(s.NormalizedEntropy), s.TopContinuation, f(s.TopContinuationProb), f(s.CheyProbability)})
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "sai_in_position_distributions.tsv"),
		[]string{"context", "stratum", "stratum_type", "occurrence_count", "unique_continuations", "entropy_bits", "normalized_entropy",
			"top_continuation", "top_continuation_probability", "chey_probability"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, d := range state.PositionDependence {
		rows = append(rows, []string{d.PositionVariable, f(d.ObservedMIBits), f(d.NullMeanMIBits), f(d.NullSDMIBits), i(d.Permutations), f(d.EmpiricalP)})
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "positional_continuation_dependence.tsv"),
		[]string{"position_variable", "observed_mi_bits", "null_mean_mi_bits", "null_sd_mi_bits", "permutations", "empirical_p"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, e := range state.PositionalEntropy {
		rows = append(rows, []string{e.PositionVariable, e.Stratum, i(e.OccurrenceCount), f(e.EntropyBits), f(e.EntropyGlobalBits),
			f(e.EntropyDifference), f(e.EffectiveContinuationCount), i(e.UniqueContinuations), f(e.EmpiricalP), i(e.Permutations)})
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "positional_entropy.tsv"),
		[]string{"position_variable", "stratum", "occurrence_count", "entropy_bits", "entropy_global_bits", "entropy_difference",
			"effective_continuation_count", "unique_continuations", "empirical_p", "permutations"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, ce := range state.CheyEffect {
		rows = append(rows, []string{ce.PositionVariable, ce.Stratum, i(ce.OccurrenceCount), i(ce.CheyCount),
			f(ce.PCheyGivenPosition), f(ce.PCheyGlobal), f(ce.PositionalEnrichment), f(ce.EmpiricalP), i(ce.Permutations)})
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "positional_chey_effect.tsv"),
		[]string{"position_variable", "stratum", "occurrence_count", "chey_count", "p_chey_given_position", "p_chey_global",
			"positional_enrichment", "empirical_p", "permutations"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, ac := range state.AiinControl {
		rows = append(rows, []string{ac.PositionVariable, ac.Stratum, i(ac.AiinOccurrenceCount), f(ac.AiinEntropyBits), i(ac.AiinUniqueContinuations),
			i(ac.CheyCount), f(ac.PCheyGivenAiinPosition), f(ac.PCheyGivenSAiinPosition), f(ac.WithinPositionEnrichment)})
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "aiin_position_control.tsv"),
		[]string{"position_variable", "stratum", "aiin_occurrence_count", "aiin_entropy_bits", "aiin_unique_continuations",
			"chey_count", "p_chey_given_aiin_position", "p_chey_given_s_aiin_position", "within_position_enrichment"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, sp := range state.StratifiedPredecessor {
		rows = append(rows, []string{sp.PositionVariable, f(sp.ObservedStatistic), f(sp.NullMeanStatistic), f(sp.NullSDStatistic), i(sp.Permutations), f(sp.EmpiricalP)})
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "stratified_predecessor_effect.tsv"),
		[]string{"position_variable", "observed_pooled_chey_given_s_count", "null_mean", "null_sd", "permutations", "empirical_p"}, rows); err != nil {
		return err
	}

	m := state.ModelLOBO
	if err := writeTSV(filepath.Join(c.OutputDir, "positional_model_lobo.tsv"),
		[]string{"tested_blocks", "blocks_m2_better_m1", "blocks_m1_better_m2", "blocks_m3_better_m2", "blocks_m2_better_m3",
			"mean_delta_21_bits", "median_delta_21_bits", "mean_delta_32_bits", "median_delta_32_bits"},
		[][]string{{i(m.TestedBlocks), i(m.BlocksM2BetterM1), i(m.BlocksM1BetterM2), i(m.BlocksM3BetterM2), i(m.BlocksM2BetterM3),
			f(m.MeanDelta21), f(m.MedianDelta21), f(m.MeanDelta32), f(m.MedianDelta32)}}); err != nil {
		return err
	}

	rows = nil
	for _, cb := range state.CrossBlock {
		rows = append(rows, []string{cb.Block, cb.Currier, cb.Hand, cb.Joint, i(cb.AiinOccurrences), i(cb.SAiinOccurrences),
			f(cb.CheyGivenAiinPosition), f(cb.CheyGivenSAiinPosition), f(cb.Enrichment), cb.EffectSign})
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "positional_cross_block.tsv"),
		[]string{"block", "currier", "hand", "joint", "aiin_occurrences", "s_aiin_occurrences",
			"p_chey_given_aiin", "p_chey_given_s_aiin", "enrichment", "effect_sign"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, j := range state.Jackknife {
		rows = append(rows, []string{j.PositionVariable, i(j.Realizations),
			f(j.MIMin), f(j.MIMax), f(j.MIMedian), f(j.MISD),
			f(j.EntropyEffectMin), f(j.EntropyEffectMax), f(j.EntropyEffectMedian), f(j.EntropyEffectSD),
			f(j.CheyEnrichmentMin), f(j.CheyEnrichmentMax), f(j.CheyEnrichmentMedian), f(j.CheyEnrichmentSD),
			f(j.StratifiedSMin), f(j.StratifiedSMax), f(j.StratifiedSMedian), f(j.StratifiedSSD),
			bo(j.SingleBlockSensitive)})
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "positional_jackknife.tsv"),
		[]string{"position_variable", "realizations",
			"mi_min", "mi_max", "mi_median", "mi_sd",
			"entropy_effect_min", "entropy_effect_max", "entropy_effect_median", "entropy_effect_sd",
			"chey_enrichment_min", "chey_enrichment_max", "chey_enrichment_median", "chey_enrichment_sd",
			"stratified_s_min", "stratified_s_max", "stratified_s_median", "stratified_s_sd",
			"single_block_sensitive"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, lb := range state.LineVsBlock {
		rows = append(rows, []string{lb.Analysis, lb.LineCategory, lb.BlockCoarseBucket, i(lb.OccurrenceCount), f(lb.CheyProbability)})
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "line_vs_block_position.tsv"),
		[]string{"analysis", "line_category", "block_coarse_bucket", "occurrence_count", "chey_probability"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, bd := range state.BoundaryDistance {
		rows = append(rows, []string{bd.Group, bd.Metric, f(bd.Median), f(bd.Q25), f(bd.Q75), f(bd.PermutationP)})
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "boundary_distance.tsv"),
		[]string{"group", "metric", "median", "q25", "q75", "permutation_p"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, sc := range state.SurroundingContext {
		rows = append(rows, []string{sc.Group, i(sc.OccurrenceCount), f(sc.PrecedingEntropyBits), f(sc.FollowingEntropyBits), i(sc.UniqueSurroundingContexts)})
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "sai_in_surrounding_context.tsv"),
		[]string{"group", "occurrence_count", "preceding_entropy_bits", "following_entropy_bits", "unique_surrounding_contexts"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, rp := range state.ReversePosition {
		rows = append(rows, []string{rp.PositionVariable, rp.Stratum, f(rp.PGivenSAiin), f(rp.PGivenAiin), f(rp.TotalVariation)})
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "reverse_position.tsv"),
		[]string{"position_variable", "stratum", "p_stratum_given_s_aiin", "p_stratum_given_aiin", "total_variation"}, rows); err != nil {
		return err
	}

	v := state.Validation
	if err := writeTSV(filepath.Join(c.OutputDir, "positional_continuation_validation.tsv"),
		[]string{"final_status", "position_dependence_p", "position_dependence_significant", "stratified_predecessor_significant",
			"m3_better_than_m2_fraction", "cross_block_sign_consistency", "single_block_sensitive", "boundary_formula_supported", "eligible_blocks"},
		[][]string{{v.FinalStatus, f(v.PositionDependenceP), bo(v.PositionDependenceSig), bo(v.StratifiedPredecessorSig),
			f(v.M3BetterThanM2Fraction), f(v.CrossBlockSignConsistency), bo(v.SingleBlockSensitive), bo(v.BoundaryFormulaSupported), i(v.EligibleBlocks)}}); err != nil {
		return err
	}

	if err := writePlots(c.OutputDir, state); err != nil {
		return err
	}
	if err := writeYAML(c, meta, state); err != nil {
		return err
	}
	return writeReport(c, meta, state)
}

type yamlDoc struct {
	Meta            map[string]any `yaml:"meta"`
	Reproducibility map[string]any `yaml:"reproducibility"`
	Frozen          map[string]any `yaml:"frozen"`
	PositionalBins  map[string]any `yaml:"positional_bins"`
	FinalStatus     string         `yaml:"final_status"`
}

func writeYAML(c Config, meta runMeta, state *RunState) error {
	doc := yamlDoc{
		Meta: map[string]any{
			"confirmatory": true,
			"permutations": meta.Permutations,
			"seed":         meta.Seed,
			"alpha":        smoothingAlpha,
		},
		Reproducibility: map[string]any{
			"corpus_sha256":       meta.CorpusSHA256,
			"metadata_sha256":     meta.MetadataSHA256,
			"higher_order_sha256": meta.HigherOrderSHA256,
			"permutations":        meta.Permutations,
			"seed":                meta.Seed,
		},
		Frozen: map[string]any{
			"context":              FrozenSAiin,
			"target_continuation":  FrozenChey,
		},
		PositionalBins: map[string]any{
			"line_categories":        lineCategories,
			"block_coarse_categories": blockCoarseCategories,
			"block_fixed_categories":  blockFixedCategories,
		},
		FinalStatus: state.Validation.FinalStatus,
	}
	b, err := yaml.Marshal(doc)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(c.OutputDir, "positional_continuation_analysis.yaml"), b, 0644)
}
