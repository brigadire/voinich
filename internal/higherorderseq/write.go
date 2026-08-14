package higherorderseq

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

func writeAll(c Config, meta runMeta, candidates []Candidate, results []*CandidateResult, dependence []DependenceRow, validation []ValidationRow) error {
	if err := os.MkdirAll(filepath.Join(c.OutputDir, "plots"), 0755); err != nil {
		return err
	}

	var rows [][]string
	for _, cand := range candidates {
		rows = append(rows, []string{cand.Sequence, i(len(cand.Tokens)), cand.Family, i(cand.CanonicalOccurrences), i(cand.PhysicalBlocks), i(cand.JointClasses), f(cand.ShuffleFDRQ), f(cand.MarkovBlockP)})
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "higher_order_candidate_inventory.tsv"), []string{"sequence", "n", "family", "canonical_occurrences", "physical_blocks", "joint_classes", "source_shuffle_block_fdr_q", "source_markov_block_p"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, r := range results {
		for _, o := range r.Occurrences {
			rows = append(rows, []string{o.Sequence, i(o.PosA), i(o.PosB), i(o.PosC), o.Block, o.Currier, o.Hand, o.Joint, f(o.NormalizedBlockPos), bo(o.WithinSameLine), bo(o.CrossesLineBoundary), o.LinePosition})
		}
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "higher_order_occurrences.tsv"), []string{"sequence", "token_position_a", "token_position_b", "token_position_c", "physical_block", "currier", "hand", "joint_class", "normalized_block_position", "within_same_line", "crosses_line_boundary", "line_position"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, r := range results {
		for _, cr := range r.ConditionalRows {
			rows = append(rows, []string{cr.Sequence, cr.Block, cr.Currier, cr.Hand, cr.Joint,
				i(cr.CountB), i(cr.CountAB), i(cr.CountBC), i(cr.CountABC),
				f(cr.PCGivenB), f(cr.PCGivenAB), f(cr.Enrichment), f(cr.DeltaProbability),
				bo(cr.EligiblePrimary), bo(cr.EligibleDescriptive),
				f(cr.PAGivenB), f(cr.PAGivenBC), f(cr.ReverseEnrichment), f(cr.ReverseDeltaProbability)})
		}
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "conditional_probability_by_block.tsv"),
		[]string{"sequence", "block", "currier", "hand", "joint", "count_b", "count_ab", "count_bc", "count_abc",
			"p_c_given_b", "p_c_given_ab", "enrichment", "delta_probability", "eligible_primary", "eligible_descriptive",
			"p_a_given_b", "p_a_given_bc", "reverse_enrichment", "reverse_delta_probability"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, d := range dependence {
		rows = append(rows, []string{d.Sequence, d.Family, i(d.Permutations), f(d.EmpiricalP), f(d.FDRQ), bo(d.Significant)})
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "conditional_dependence.tsv"), []string{"sequence", "family", "permutations", "empirical_p", "fdr_q_or_raw_p_for_secondary", "significant"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, r := range results {
		m := r.CMI
		rows = append(rows, []string{m.Sequence, m.CenterToken, i(m.Occurrences), f(m.ObservedCMIBits), f(m.NullMeanCMIBits), f(m.NullSDCMIBits), i(m.Permutations), f(m.EmpiricalP), f(m.ContributionBits)})
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "conditional_cmi.tsv"), []string{"sequence", "center_token_b", "b_occurrences_with_both_neighbors", "observed_cmi_bits", "null_mean_cmi_bits", "null_sd_cmi_bits", "permutations", "empirical_p", "frozen_ac_pointwise_contribution_bits"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, r := range results {
		for _, cont := range r.Continuations {
			rows = append(rows, []string{cont.Sequence, cont.Context, cont.Token, i(cont.Count), f(cont.Probability)})
		}
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "continuation_distributions.tsv"), []string{"sequence", "context", "token", "count", "probability"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, r := range results {
		e := r.ContinuationEnt
		rows = append(rows, []string{e.Sequence, f(e.HGivenB), f(e.HGivenAB), f(e.EntropyReduction), f(e.JSDivergence), f(e.TotalVariation)})
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "continuation_entropy.tsv"), []string{"sequence", "h_x_given_b_bits", "h_x_given_ab_bits", "entropy_reduction_bits", "js_divergence_bits", "total_variation_distance"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, r := range results {
		l := r.LOBO
		rows = append(rows, []string{l.Sequence, i(l.TestedBlocks), i(l.M2BetterBlocks), i(l.M1BetterBlocks), i(l.Ties), f(l.MeanDeltaLogLoss), f(l.MedianDeltaLogLoss), f(l.HeldoutLogLikelihoodRatio)})
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "first_vs_second_order_lobo.tsv"), []string{"sequence", "tested_blocks", "m2_better_blocks", "m1_better_blocks", "ties", "mean_delta_log_loss_bits", "median_delta_log_loss_bits", "heldout_log_likelihood_ratio_bits"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, r := range results {
		for _, cc := range r.ContextControls {
			rows = append(rows, []string{cc.Sequence, cc.ContextType, cc.AltToken, i(cc.Count), f(cc.Probability), bo(cc.IsFrozen)})
		}
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "conditional_context_controls.tsv"), []string{"sequence", "context_type", "alt_token", "count", "probability", "is_frozen"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, r := range results {
		cr := r.ContextRank
		rows = append(rows, []string{cr.Sequence, i(cr.NumAlternatives), f(cr.FrozenP), f(cr.BaselineP), i(cr.Rank), f(cr.Percentile), f(cr.MinAltP), f(cr.MedianAltP), f(cr.MaxAltP)})
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "conditional_context_rank.tsv"), []string{"sequence", "num_alternatives", "frozen_p_c_given_ab", "baseline_p_c_given_b", "rank_descending", "percentile", "min_alt_p", "median_alt_p", "max_alt_p"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, r := range results {
		cb, m := r.CrossBlock, r.Meta
		rows = append(rows, []string{cb.Sequence, i(cb.EligibleBlocks), i(cb.PositiveEnrichmentBlocks), i(cb.NegativeEnrichmentBlocks), f(cb.SignConsistency), i(cb.DistinctCurrier), i(cb.DistinctHand), i(cb.DistinctJoint), bo(cb.CrossCurrier), bo(cb.CrossHand), bo(cb.CrossJoint),
			f(m.UnweightedMeanLogEnrichment), f(m.WeightedMeanLogEnrichment), f(m.MedianLogEnrichment), f(m.BetweenBlockVariance), f(m.CochranQ), f(m.I2), f(m.MaxBlockWeightFraction)})
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "higher_order_cross_block.tsv"), []string{"sequence", "eligible_blocks", "positive_enrichment_blocks", "negative_enrichment_blocks", "sign_consistency", "distinct_currier", "distinct_hand", "distinct_joint", "cross_currier", "cross_hand", "cross_joint",
		"log_enrichment_unweighted_mean", "log_enrichment_weighted_mean", "log_enrichment_median", "log_enrichment_between_block_variance", "cochran_q", "i_squared_pct", "max_block_weight_fraction"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, r := range results {
		j := r.Jackknife
		rows = append(rows, []string{j.Sequence, i(j.Realizations), f(j.EnrichmentMin), f(j.EnrichmentMax), f(j.EnrichmentMedian), f(j.EnrichmentSD), f(j.CMIMin), f(j.CMIMax), f(j.CMIMedian), f(j.CMISD), f(j.DeltaLogLossMin), f(j.DeltaLogLossMax), f(j.DeltaLogLossMedian), f(j.DeltaLogLossSD), bo(j.SingleBlockSensitive)})
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "higher_order_jackknife.tsv"), []string{"sequence", "realizations", "enrichment_min", "enrichment_max", "enrichment_median", "enrichment_sd", "cmi_min", "cmi_max", "cmi_median", "cmi_sd", "delta_log_loss_min", "delta_log_loss_max", "delta_log_loss_median", "delta_log_loss_sd", "single_block_sensitive"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, r := range results {
		for _, pr := range r.Position {
			rows = append(rows, []string{pr.Sequence, pr.Metric, pr.Bucket, i(pr.ABCCount), i(pr.ABCount), f(pr.ABCFraction), f(pr.ABFraction)})
		}
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "higher_order_position.tsv"), []string{"sequence", "metric", "bucket", "abc_count", "ab_count", "abc_fraction", "ab_fraction"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, r := range results {
		for _, sf := range r.StructuralFamily {
			rows = append(rows, []string{sf.Sequence, sf.TokenRole, sf.Token, sf.Relative, bo(sf.Sufficient), f(sf.FrozenP), f(sf.RelativeP), bo(sf.SignHolds)})
		}
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "higher_order_structural_family.tsv"), []string{"sequence", "token_role", "token", "relative", "sufficient_data", "baseline_p_x_given_b", "relative_p_x_given_relative_b", "sign_holds"}, rows); err != nil {
		return err
	}

	rows = nil
	for _, v := range validation {
		rows = append(rows, []string{v.Sequence, v.Family, v.FinalStatus, f(v.ConditionalFDRQ), i(v.EligibleBlocks), f(v.SignConsistency), f(v.LOBOAdvantageFraction), bo(v.SingleBlockSensitive), i(v.DistinctJointClasses), bo(v.PositionDependent), bo(v.MetadataLimited)})
	}
	if err := writeTSV(filepath.Join(c.OutputDir, "higher_order_validation.tsv"), []string{"sequence", "family", "final_status", "conditional_fdr_q", "eligible_blocks", "sign_consistency", "lobo_advantage_fraction", "single_block_sensitive", "distinct_joint_classes", "position_dependent", "metadata_limited"}, rows); err != nil {
		return err
	}

	if err := writePlots(c.OutputDir, results); err != nil {
		return err
	}
	if err := writeYAML(c, meta, candidates, validation); err != nil {
		return err
	}
	return writeReport(c, meta, candidates, results, dependence, validation)
}

type yamlDoc struct {
	Meta         map[string]any   `yaml:"meta"`
	Reproducibility map[string]any `yaml:"reproducibility"`
	FrozenCandidates []map[string]any `yaml:"frozen_candidates"`
	Outcomes     []map[string]any `yaml:"outcomes"`
	MainOutcome  string           `yaml:"main_outcome"`
}

func writeYAML(c Config, meta runMeta, candidates []Candidate, validation []ValidationRow) error {
	doc := yamlDoc{
		Meta: map[string]any{
			"confirmatory": true,
			"token_count":  meta.TokenCount,
			"permutations": meta.Permutations,
			"seed":         meta.Seed,
		},
		Reproducibility: map[string]any{
			"corpus_sha256":            meta.CorpusSHA256,
			"metadata_sha256":          meta.MetadataSHA256,
			"previous_audit_sha256":    meta.AuditSHA256,
			"permutations_primary":     meta.Permutations,
			"permutations_secondary":  secondaryPermutations(meta.Permutations),
			"seed":                     meta.Seed,
		},
	}
	for _, cand := range candidates {
		doc.FrozenCandidates = append(doc.FrozenCandidates, map[string]any{
			"sequence": cand.Sequence, "family": cand.Family,
			"shuffle_block_fdr_q": cand.ShuffleFDRQ, "markov_block_p": cand.MarkovBlockP,
		})
	}
	replicated := false
	for _, v := range validation {
		doc.Outcomes = append(doc.Outcomes, map[string]any{
			"sequence": v.Sequence, "family": v.Family, "final_status": v.FinalStatus,
		})
		if v.FinalStatus == "HIGHER_ORDER_REPLICATED" {
			replicated = true
		}
	}
	if replicated {
		doc.MainOutcome = "HIGHER_ORDER_CONDITIONAL_DEPENDENCE_REPLICATED"
	} else {
		doc.MainOutcome = "NO_REPLICATED_HIGHER_ORDER_CONDITIONAL_DEPENDENCE"
	}
	b, err := yaml.Marshal(doc)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(c.OutputDir, "higher_order_sequence_analysis.yaml"), b, 0644)
}
