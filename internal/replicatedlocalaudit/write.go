package replicatedlocalaudit

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func f(v float64) string {
	if math.IsNaN(v) {
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

func writeOutputs(c Config, sha string, ds []distanceResult, ss []sequenceResult) error {
	if e := os.MkdirAll(c.OutputDir, 0755); e != nil {
		return e
	}
	var rows [][]string
	for _, x := range ds {
		d := x.Candidate
		rows = append(rows, []string{d.A + " / " + d.B, i(d.Eligible), i(d.Joint), i(d.Currier), i(d.Hands), f(d.Mean), f(d.Median), f(d.Min), f(d.Transfer), f(d.RawP), f(d.Q), f(d.Threshold), f(d.Fraction), x.FailedConditions})
	}
	if e := writeTSV(filepath.Join(c.OutputDir, "distance_profile_classification_audit.tsv"), []string{"relation", "eligible_blocks", "joint_classes", "currier_classes", "hands", "mean_similarity", "median_similarity", "min_similarity", "transfer_success", "raw_p", "fdr_q", "frozen_threshold", "fraction_above_threshold", "failed_classification_conditions"}, rows); e != nil {
		return e
	}
	rows = nil
	for _, x := range ds {
		for _, z := range x.Rows {
			rows = append(rows, []string{z.CandidateID, z.A + " / " + z.B, z.Block, z.Currier, z.Hand, z.Joint, i(z.CountA), i(z.CountB), i(z.Observations), i(z.ComparedProfileCells), f(z.Similarity), f(z.LOBO), bo(z.Transfer), f(z.ShapeSimilarity), i(z.Peak), f(z.Center), f(z.Asymmetry)})
		}
	}
	if e := writeTSV(filepath.Join(c.OutputDir, "distance_profile_lobo.tsv"), []string{"candidate_id", "relation", "block", "currier", "hand", "joint_class", "token_count_a", "token_count_b", "observations", "compared_profile_cells", "similarity", "similarity_to_LOBO_reference", "transfer_success", "unit_normalized_shape_similarity", "peak_distance", "center_of_mass", "left_right_asymmetry"}, rows); e != nil {
		return e
	}
	rows = nil
	for _, x := range ds {
		for _, z := range x.Rows {
			rows = append(rows, []string{z.CandidateID, z.Block, f(z.Observed()), f(z.NullMean), f(z.NullSD), f(z.P95), f(z.P99), f(z.Standardized), f(z.Effect)})
		}
	}
	if e := writeTSV(filepath.Join(c.OutputDir, "distance_profile_null_effects.tsv"), []string{"candidate_id", "block", "observed_similarity", "null_mean", "null_sd", "null_p95", "null_p99", "standardized_similarity", "effect"}, rows); e != nil {
		return e
	}
	rows = nil
	for _, x := range ds {
		rows = append(rows, []string{x.Candidate.ID, f(x.FullEffect), f(x.MinJackknife), f(x.MaxJackknife), f(x.JackknifeSD), f(x.MaxJackknifeP), bo(x.JackknifeSurvives), f(x.MaxObservationFraction), f(x.MaxEffectContribution)})
	}
	if e := writeTSV(filepath.Join(c.OutputDir, "distance_profile_jackknife.tsv"), []string{"candidate_id", "full_effect", "min_jackknife_effect", "max_jackknife_effect", "jackknife_sd", "max_jackknife_raw_p", "significance_survives_each_block_removal", "max_block_observation_fraction", "max_block_effect_contribution"}, rows); e != nil {
		return e
	}
	rows = nil
	statusHead := []string{"candidate_id", "replication_status", "positive_effect_blocks", "negative_effect_blocks", "effect_sign_consistency", "unweighted_mean_standardized_effect", "weighted_mean_standardized_effect"}
	if c.Generic {
		// No real hand dimension exists in generic mode (see Config.Generic);
		// within/cross-group transfer (WithinJoint/CrossJoint) is reported
		// instead of the Currier/hand-labeled columns, which would otherwise
		// show a structurally-constant, misleading "hand" signal.
		statusHead = append(statusHead, "within_group_transfer", "cross_group_transfer")
		for _, x := range ds {
			rows = append(rows, []string{x.Candidate.ID, x.Status, i(x.Positive), i(x.Negative), f(float64(x.Positive) / float64(max(1, len(x.Rows)))), f(x.MeanZ), f(x.WeightedZ), bo(x.WithinJoint), bo(x.CrossJoint)})
		}
	} else {
		statusHead = append(statusHead, "within_currier_transfer", "cross_currier_transfer", "within_hand_transfer", "cross_hand_transfer", "within_joint_transfer", "cross_joint_transfer")
		for _, x := range ds {
			rows = append(rows, []string{x.Candidate.ID, x.Status, i(x.Positive), i(x.Negative), f(float64(x.Positive) / float64(max(1, len(x.Rows)))), f(x.MeanZ), f(x.WeightedZ), bo(x.WithinCurrier), bo(x.CrossCurrier), bo(x.WithinHand), bo(x.CrossHand), bo(x.WithinJoint), bo(x.CrossJoint)})
		}
	}
	if e := writeTSV(filepath.Join(c.OutputDir, "distance_profile_replication_status.tsv"), statusHead, rows); e != nil {
		return e
	}
	rows = nil

	for _, x := range ss {
		o := x.Observed
		occ := make([]string, len(o.TokenOccurrences))
		for j, n := range o.TokenOccurrences {
			occ[j] = i(n)
		}
		rows = append(rows, []string{x.Candidate.Sequence, i(len(x.Candidate.Tokens)), i(o.Total), i(o.Eligible), i(o.Blocks), i(o.Joint), i(o.Currier), i(o.Hands), f(o.MaxFraction), f(o.Entropy), x.Candidate.Classification, o.Validity, strings.Join(occ, ","), bo(o.ContainsQuestion), bo(o.ContainsAt), bo(o.ContainsOtherMarker), strings.Join(o.AbsentTokens, ",")})
	}
	if e := writeTSV(filepath.Join(c.OutputDir, "universal_sequence_inventory.tsv"), []string{"sequence", "n", "canonical_occurrences", "eligible_occurrences", "physical_blocks", "joint_classes", "currier_classes", "hands", "max_block_fraction", "block_entropy_bits", "classification", "token_validity", "canonical_token_occurrences", "contains_question", "contains_at", "contains_other_nonstandard_marker", "absent_tokens"}, rows); e != nil {
		return e
	}
	rows = nil
	for _, x := range ss {
		rows = append(rows, []string{x.Candidate.Sequence, i(len(x.Candidate.Tokens)), i(x.Observed.Total), i(x.Observed.Blocks), f(x.ShuffleMeanTotal), f(x.ShuffleMeanBlocks), f(x.ShuffleTotalP), f(x.ShuffleP), f(x.ShuffleQ), i(x.MarkovAvailableBlocks), f(x.MarkovMeanTotal), f(x.MarkovMeanBlocks), f(x.MarkovTotalP), f(x.MarkovP)})
	}
	if e := writeTSV(filepath.Join(c.OutputDir, "sequence_null_validation.tsv"), []string{"sequence", "n", "observed_total_occurrences", "observed_block_recurrence", "shuffle_null_mean_total", "shuffle_null_mean_blocks", "shuffle_total_p", "shuffle_block_p", "shuffle_block_fdr_q", "markov_available_blocks", "markov_null_mean_total", "markov_null_mean_blocks", "markov_total_p", "markov_block_p"}, rows); e != nil {
		return e
	}
	rows = nil
	for _, x := range ss {
		o := x.Observed
		if o.Total >= 3 && o.Blocks >= 3 && o.Joint >= 2 && o.MaxFraction <= .7 && o.Validity == "canonical-clean" {
			rows = append(rows, []string{x.Candidate.Sequence, i(len(x.Candidate.Tokens)), i(o.Total), i(o.Blocks), i(o.Joint), f(o.MaxFraction), f(x.ShuffleQ)})
		}
	}
	if e := writeTSV(filepath.Join(c.OutputDir, "strict_replicated_sequences.tsv"), []string{"sequence", "n", "canonical_occurrences", "physical_blocks", "joint_classes", "max_block_fraction", "shuffle_block_fdr_q"}, rows); e != nil {
		return e
	}
	rows = nil
	for _, x := range ss {
		rows = append(rows, []string{x.Candidate.Sequence, i(len(x.Candidate.Tokens)), x.Observed.Validity, x.Status, f(x.ShuffleP), f(x.ShuffleQ), f(x.MarkovP)})
	}
	if e := writeTSV(filepath.Join(c.OutputDir, "sequence_replication_status.tsv"), []string{"sequence", "n", "token_validity", "sequence_status", "shuffle_block_p", "shuffle_block_fdr_q", "markov_block_p"}, rows); e != nil {
		return e
	}
	if e := writeSummary(c, sha, ds, ss); e != nil {
		return e
	}
	return writeReport(c, sha, ds, ss)
}
func (z distanceRow) Observed() float64 { return z.LOBO }

type familySummary struct {
	Family            string `yaml:"family"`
	FrozenCandidates  int    `yaml:"frozen_candidates"`
	Testable          int    `yaml:"testable"`
	PreviousUniversal int    `yaml:"previous_universal"`
	FDRSignificant    int    `yaml:"fdr_significant"`
	RobustCrossBlock  int    `yaml:"robust_cross_block"`
	CrossCurrier      int    `yaml:"cross_currier"`
	CrossHand         int    `yaml:"cross_hand"`
	ArtifactOrInvalid int    `yaml:"artifact_or_invalid"`
}

func summaries(ds []distanceResult, ss []sequenceResult, generic bool) []familySummary {
	topClass, topStatus := "UNIVERSAL", "ROBUST_RELATIVE_REPLICATION"
	if generic {
		topClass, topStatus = "GROUP_CONSISTENT", "GROUP_REPLICATED"
	}
	d := familySummary{Family: "distance-profile", FrozenCandidates: len(ds)}
	for _, x := range ds {
		if len(x.Rows) > 0 {
			d.Testable++
		}
		if x.Candidate.Classification == topClass {
			d.PreviousUniversal++
		}
		if x.Candidate.Q <= .05 {
			d.FDRSignificant++
		}
		if x.Status == topStatus {
			d.RobustCrossBlock++
		}
		if x.CrossCurrier {
			d.CrossCurrier++
		}
		if x.CrossHand {
			d.CrossHand++
		}
	}
	s := familySummary{Family: "sequence", FrozenCandidates: len(ss), Testable: len(ss), PreviousUniversal: len(ss)}
	for _, x := range ss {
		if x.ShuffleQ <= .05 {
			s.FDRSignificant++
		}
		if x.Status == "REPLICATED_ABOVE_FREQUENCY_NULL" {
			s.RobustCrossBlock++
		}
		if x.Observed.Currier >= 2 {
			s.CrossCurrier++
		}
		if x.Observed.Hands >= 2 {
			s.CrossHand++
		}
		if x.Status == "TRANSCRIPTION_AMBIGUOUS" || x.Status == "INVALID_CANONICAL_SUPPORT" {
			s.ArtifactOrInvalid++
		}
	}
	return []familySummary{d, s}
}
func writeSummary(c Config, sha string, ds []distanceResult, ss []sequenceResult) error {
	sum := summaries(ds, ss, c.Generic)
	var rows [][]string
	for _, x := range sum {
		rows = append(rows, []string{x.Family, i(x.FrozenCandidates), i(x.Testable), i(x.PreviousUniversal), i(x.FDRSignificant), i(x.RobustCrossBlock), i(x.CrossCurrier), i(x.CrossHand), i(x.ArtifactOrInvalid)})
	}
	if e := writeTSV(filepath.Join(c.OutputDir, "replicated_local_structure_summary.tsv"), []string{"family", "frozen_candidates", "testable", "previous_universal", "fdr_significant", "robust_cross_block", "cross_currier", "cross_hand", "artifact_or_invalid"}, rows); e != nil {
		return e
	}
	doc := struct {
		Meta        map[string]any  `yaml:"meta"`
		Methodology map[string]any  `yaml:"methodology"`
		Summary     []familySummary `yaml:"summary"`
		MainOutcome string          `yaml:"main_outcome"`
	}{Meta: map[string]any{"corpus_sha256": sha, "permutations": c.Permutations, "seed": c.Seed, "confirmatory": true}, Methodology: map[string]any{"profile_similarity": "mean 1 - Jensen-Shannon divergence / ln(2) over non-empty token context distributions at frozen signed distances 1..20", "fraction_above_frozen_threshold": "legacy validation compares [0,1] similarity to discovery max_distance=20; retained unchanged and therefore zero", "sequence_primary_statistic": "physical block recurrence count", "shuffle_null": "token order shuffled within physical blocks", "markov_null": "first-order model trained on other blocks of the held-out joint metadata class; unavailable blocks are NA"}, Summary: sum}
	doc.MainOutcome = mainOutcome(ds, ss)
	b, e := yaml.Marshal(doc)
	if e != nil {
		return e
	}
	return os.WriteFile(filepath.Join(c.OutputDir, "replicated_local_structure.yaml"), b, 0644)
}

func mainOutcome(ds []distanceResult, ss []sequenceResult) string {
	distance, higher := false, false
	for _, x := range ds {
		if x.Status == "ROBUST_RELATIVE_REPLICATION" {
			distance = true
		}
	}
	for _, x := range ss {
		if len(x.Candidate.Tokens) >= 3 && x.Status == "REPLICATED_ABOVE_FREQUENCY_NULL" {
			higher = true
		}
	}
	if distance && higher {
		return "BOTH_REMAIN"
	}
	if distance {
		return "DISTANCE_STRUCTURE_REMAINS"
	}
	if higher {
		return "HIGHER_ORDER_SEQUENCES_REMAIN"
	}
	return "NOTHING_REMAINS"
}
func writeReport(c Config, sha string, ds []distanceResult, ss []sequenceResult) error {
	sum := summaries(ds, ss, c.Generic)
	ng := map[int]int{}
	freq := []float64{}
	statuses := map[string]int{}
	higherMarkov := 0
	higherShuffle := 0
	var significantDistance, descriptiveDistance []float64
	distanceRemains := false
	robustDistanceTokens := map[string]bool{}
	for _, x := range ds {
		if x.Candidate.Q <= .05 {
			significantDistance = append(significantDistance, x.Candidate.Median)
		} else {
			descriptiveDistance = append(descriptiveDistance, x.Candidate.Median)
		}
		if x.Status == "ROBUST_RELATIVE_REPLICATION" {
			distanceRemains = true
			robustDistanceTokens[x.Candidate.A] = true
			robustDistanceTokens[x.Candidate.B] = true
		}
	}
	overlapSequences := 0
	for _, x := range ss {
		ng[len(x.Candidate.Tokens)]++
		freq = append(freq, float64(x.Observed.Total))
		statuses[x.Status]++
		if len(x.Candidate.Tokens) >= 3 && x.Status == "REPLICATED_ABOVE_FREQUENCY_NULL" && !math.IsNaN(x.MarkovP) && x.MarkovP <= .05 {
			higherMarkov++
		}
		if len(x.Candidate.Tokens) >= 3 && x.Status == "REPLICATED_ABOVE_FREQUENCY_NULL" {
			higherShuffle++
		}
		if len(x.Candidate.Tokens) >= 3 && x.Status == "REPLICATED_ABOVE_FREQUENCY_NULL" {
			for _, tok := range x.Candidate.Tokens {
				if robustDistanceTokens[tok] {
					overlapSequences++
					break
				}
			}
		}
	}
	sort.Float64s(freq)
	mean, _ := meanSD(freq)
	median, p90, minv, maxv := quantile(freq, .5), quantile(freq, .9), quantile(freq, 0), quantile(freq, 1)
	var b strings.Builder
	fmt.Fprintf(&b, "# Replicated local structure audit\n\nConfirmatory audit of frozen candidates only. Corpus SHA256: `%s`. No token pairs, n-grams, distances, thresholds, or classification rules were added or changed.\n\n## Metric semantics\n\nDistance profile similarity is the mean Jensen–Shannon similarity (`1 - JSD/ln(2)`) across non-empty left/right context distributions at frozen distances ±1…±20. The legacy `fraction_above_frozen_threshold` compares this [0,1] similarity with the discovery field `max_distance=20`; its zeros are retained and documented, not repaired retrospectively. Absolute LOBO similarity and null-relative standardized effect are reported separately.\n\n## Summary\n\n", sha)
	for _, x := range sum {
		fmt.Fprintf(&b, "- %s: %d frozen, %d FDR-significant, %d robust cross-block, %d artifact/invalid.\n", x.Family, x.FrozenCandidates, x.FDRSignificant, x.RobustCrossBlock, x.ArtifactOrInvalid)
	}
	conclusion := "A. NOTHING REMAINS — no convincing transferable local structure remains under current tests."
	if distanceRemains && higherShuffle > 0 {
		conclusion = "D. BOTH REMAIN — distance profiles and higher-order sequences retain confirmatory support; their frozen token-family overlap is a descriptive follow-up only."
	} else if distanceRemains {
		conclusion = "B. DISTANCE STRUCTURE REMAINS — weak but reproducible relative-distance organization exists."
	} else if higherShuffle > 0 {
		conclusion = "C. HIGHER-ORDER SEQUENCES REMAIN — reproducible higher-order sequential organization exists."
	}
	if distanceRemains && higherShuffle > 0 {
		conclusion += fmt.Sprintf(" %d/%d supported higher-order sequences share at least one token with a robust distance pair.", overlapSequences, higherShuffle)
	}
	fmt.Fprintf(&b, "\n## Distance-profile distributions\n\nFrozen median similarity: q<=0.05 primary set median %.6g (n=%d); q>0.05 descriptive set median %.6g (n=%d). This comparison performs no new selection or significance test.\n\n## Main outcome\n\n%s\n", quantile(significantDistance, .5), len(significantDistance), quantile(descriptiveDistance, .5), len(descriptiveDistance), conclusion)
	fmt.Fprintf(&b, "\n## UNIVERSAL sequence inventory\n\nLength distribution:")
	keys := make([]int, 0, len(ng))
	for k := range ng {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		fmt.Fprintf(&b, " n=%d: %d;", k, ng[k])
	}
	fmt.Fprintf(&b, "\n\nCanonical occurrence distribution: min %.0f, median %.1f, mean %.2f, P90 %.1f, max %.0f.\n\nDiagnostic statuses:\n", minv, median, mean, p90, maxv)
	keysS := make([]string, 0, len(statuses))
	for k := range statuses {
		keysS = append(keysS, k)
	}
	sort.Strings(keysS)
	for _, k := range keysS {
		fmt.Fprintf(&b, "- %s: %d\n", k, statuses[k])
	}
	fmt.Fprintf(&b, "\nHigher-order sequences (n>=3) exceeding both the shuffle FDR criterion and nominal Markov p<=0.05: %d. Markov values are secondary and are NA where leakage-free same-class training is unavailable.\n\n## Interpretation guardrails\n\nDistance and sequence p-values are separate families. New diagnostic statuses do not replace the frozen UNIVERSAL/WEAK/BLOCK_SPECIFIC classifications. The audit establishes formal reproducibility only; it does not establish language, semantics, grammar, or decipherment.\n", higherMarkov)
	return os.WriteFile(filepath.Join(c.OutputDir, "replicated_local_structure_report.md"), []byte(b.String()), 0644)
}
