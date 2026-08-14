package tokenrelationvalidation

import (
	"bufio"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func RunAndWrite(c Config) error {
	if c.Permutations < 0 || c.RefinePermutations < 0 {
		return fmt.Errorf("permutations must be non-negative")
	}
	if c.ProgressWriter == nil && !c.Quiet {
		c.ProgressWriter = os.Stderr
	}
	p := newProgress(c.ProgressWriter)
	a, e := analyze(c, p)
	if e != nil {
		return e
	}
	p.begin(8, "Writing validation outputs")
	if e = writeAll(a); e != nil {
		return e
	}
	p.update(1, 1, "Writing validation outputs")
	fmt.Printf("Validated %d frozen relations across %d primary physical blocks; results written to %s\n", len(a.Candidates), len(a.Blocks), c.OutputDir)
	return nil
}
func f(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }
func b(v bool) string    { return strconv.FormatBool(v) }
func i(v int) string     { return strconv.Itoa(v) }
func writeTSV(path string, header []string, rows [][]string) error {
	file, e := os.Create(path)
	if e != nil {
		return e
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	defer w.Flush()
	fmt.Fprintln(w, strings.Join(header, "\t"))
	for _, r := range rows {
		fmt.Fprintln(w, strings.Join(r, "\t"))
	}
	return nil
}
func writeAll(a Analysis) error {
	if e := os.MkdirAll(filepath.Join(a.Parameters.OutputDir, "plots"), 0755); e != nil {
		return e
	}
	out := a.Parameters.OutputDir
	var rows [][]string
	for _, x := range a.Candidates {
		rows = append(rows, []string{x.ID, x.Family, x.A, x.B, x.Sequence, b(x.Directed), x.Sources, f(x.FrozenThreshold), i(x.StoredTokenCount), i(a.TokenCount)})
	}
	if e := writeTSV(filepath.Join(out, "frozen_candidate_inventory.tsv"), []string{"candidate_id", "family", "token_a", "token_b", "sequence", "directed", "frozen_sources", "frozen_threshold", "stored_token_count", "validation_token_count"}, rows); e != nil {
		return e
	}
	rows = nil
	for _, x := range a.DirectionBlocks {
		rows = append(rows, []string{x.CandidateID, x.A, x.B, x.BlockID, x.Currier, x.Hand, x.Joint, i(x.CountA), i(x.CountB), i(x.ABeforeB), i(x.BBeforeA), i(x.ImmediateAB), i(x.ImmediateBA), i(x.ExactAB[0]), i(x.ExactAB[1]), i(x.ExactAB[2]), i(x.ExactAB[3]), i(x.ExactAB[4]), i(x.Observations), f(x.Score), f(x.EnrichmentAB), f(x.EnrichmentBA), b(x.Eligible)})
	}
	if e := writeTSV(filepath.Join(out, "directional_block_validation.tsv"), []string{"candidate_id", "token_a", "token_b", "block", "currier", "hand", "joint", "count_a", "count_b", "a_before_b", "b_before_a", "immediate_ab", "immediate_ba", "ab_d1", "ab_d2", "ab_d3", "ab_d4", "ab_d5", "observations", "direction_score", "local_enrichment_ab", "local_enrichment_ba", "primary_eligible"}, rows); e != nil {
		return e
	}
	rows = nil
	for _, x := range a.Summaries {
		if x.Family != "directional" {
			continue
		}
		rows = append(rows, summaryRow(x))
	}
	if e := writeTSV(filepath.Join(out, "directional_relation_summary.tsv"), summaryHeader(), rows); e != nil {
		return e
	}
	rows = nil
	for _, x := range a.ProfileBlocks {
		rows = append(rows, []string{x.CandidateID, x.Family, x.A, x.B, x.BlockID, x.Currier, x.Hand, x.Joint, i(x.CountA), i(x.CountB), f(x.Position), f(x.Left), f(x.Right), f(x.Distance), f(x.Overlap), f(x.Similarity), f(x.GlobalSimilarity), f(x.TrainingReference), f(x.PooledSimilarity), b(x.EligiblePrimary), b(x.EligibleDescriptive)})
	}
	profileHead := []string{"candidate_id", "family", "token_a", "token_b", "block", "currier", "hand", "joint", "count_a", "count_b", "position_similarity", "left_similarity", "right_similarity", "distance_similarity", "weighted_overlap", "combined_similarity", "heldout_to_training_similarity", "training_reference_similarity", "block_to_canonical_pooled_similarity", "eligible_10", "eligible_5"}
	var drows, srows [][]string
	for n, r := range rows {
		if a.ProfileBlocks[n].Family == "distance-profile" {
			drows = append(drows, r)
		} else {
			srows = append(srows, r)
		}
	}
	if e := writeTSV(filepath.Join(out, "distance_profile_block_validation.tsv"), profileHead, drows); e != nil {
		return e
	}
	if e := writeTSV(filepath.Join(out, "structural_pair_block_validation.tsv"), profileHead, srows); e != nil {
		return e
	}
	drows, srows = nil, nil
	for _, x := range a.Summaries {
		if x.Family == "distance-profile" {
			drows = append(drows, summaryRow(x))
		}
		if x.Family == "structural" {
			srows = append(srows, summaryRow(x))
		}
	}
	if e := writeTSV(filepath.Join(out, "distance_profile_summary.tsv"), summaryHeader(), drows); e != nil {
		return e
	}
	if e := writeTSV(filepath.Join(out, "structural_pair_summary.tsv"), summaryHeader(), srows); e != nil {
		return e
	}
	rows = nil
	for _, x := range a.Sequences {
		rows = append(rows, []string{x.CandidateID, x.Sequence, i(x.Total), i(x.PhysicalBlocks), i(x.JointClasses), i(x.CurrierClasses), i(x.Hands), f(x.MaxBlockFraction), b(x.HighRecurrence), f(x.RawP), f(x.FDRQ)})
	}
	if e := writeTSV(filepath.Join(out, "sequence_block_recurrence.tsv"), []string{"candidate_id", "sequence", "total_occurrences", "block_recurrence_count", "joint_recurrence_count", "currier_classes", "hands", "max_block_fraction", "high_recurrence", "raw_empirical_p", "fdr_q"}, rows); e != nil {
		return e
	}
	rows = nil
	for _, x := range a.Transfers {
		rows = append(rows, []string{x.CandidateID, x.Family, x.HeldoutBlock, x.TrainMetadata, x.HeldoutMetadata, f(x.Expected), f(x.Observed), b(x.Success)})
	}
	if e := writeTSV(filepath.Join(out, "leave_one_block_out_transfer.tsv"), []string{"candidate_id", "family", "heldout_block", "training_metadata", "heldout_metadata", "training_reference", "heldout_value", "success"}, rows); e != nil {
		return e
	}
	rows = nil
	for _, x := range a.MetadataTransfers {
		rows = append(rows, []string{x.CandidateID, x.Family, x.Dimension, x.Training, x.Heldout, i(x.Tested), i(x.Successful), f(x.Fraction)})
	}
	if e := writeTSV(filepath.Join(out, "metadata_transfer_matrix.tsv"), []string{"candidate_id", "family", "dimension", "training_metadata", "heldout_metadata", "tested", "successful", "success_fraction"}, rows); e != nil {
		return e
	}
	rows = nil
	for _, x := range a.Controls {
		rows = append(rows, []string{x.CandidateID, x.Family, x.Kind, x.ControlA, x.ControlB, f(x.Observed), f(x.NullMean), f(x.RawP), f(x.Percentile), i(x.Permutations)})
	}
	if e := writeTSV(filepath.Join(out, "relation_controls.tsv"), []string{"candidate_id", "family", "control_kind", "control_a", "control_b", "observed", "null_mean", "raw_empirical_p", "control_percentile", "permutations"}, rows); e != nil {
		return e
	}
	rows = nil
	for _, x := range a.Summaries {
		rows = append(rows, []string{x.CandidateID, x.Family, x.A, x.B, x.Sequence, x.Classification, i(x.PhysicalBlocks), i(x.JointClasses), i(x.CurrierClasses), i(x.Hands), f(x.SignConsistency), f(x.MedianEnrichment), f(x.ProfileMedian), f(x.TransferSuccess), f(x.RawP), f(x.FDRQ)})
	}
	if e := writeTSV(filepath.Join(out, "relation_classification.tsv"), []string{"candidate_id", "family", "token_a", "token_b", "sequence", "classification", "physical_blocks", "joint_classes", "currier_classes", "hands", "direction_consistency", "median_enrichment", "profile_stability", "transfer_success", "raw_empirical_p", "fdr_q"}, rows); e != nil {
		return e
	}
	rows = nil
	for _, x := range a.Summaries {
		if RuleLike(x) {
			rows = append(rows, []string{x.CandidateID, x.Family, i(x.PhysicalBlocks), i(x.JointClasses), i(x.CurrierClasses), i(x.Hands), f(x.SignConsistency), f(x.MedianEnrichment), f(x.ProfileMedian), f(x.TransferSuccess), f(x.RawP), f(x.FDRQ)})
		}
	}
	if e := writeTSV(filepath.Join(out, "rule_like_relations.tsv"), []string{"relation", "type", "blocks", "joint_classes", "currier_classes", "hands", "direction_consistency", "median_enrichment", "profile_stability", "transfer_success", "empirical_p", "fdr_q"}, rows); e != nil {
		return e
	}
	y := map[string]any{"methodology": map[string]any{"candidate_policy": "identities, parameters, and thresholds are loaded only from pre-metadata discovery outputs; canonical counts are recomputed", "primary_unit": "contiguous known Currier x hand physical block", "unknown_metadata": "excluded from primary validation", "direction_enrichment": "P(B within frozen d | A, block) / P(B within frozen d | block)", "multiple_comparisons": "Benjamini-Hochberg separately within each relation family", "semantics": "formal transferable token relations only"}, "reproducibility": map[string]any{"corpus_sha256": a.CorpusSHA, "metadata_map_sha256": a.MetadataSHA, "token_count": a.TokenCount, "unknown_metadata_tokens_excluded": a.UnknownTokens, "seed": a.Parameters.Seed, "permutations": a.Parameters.Permutations, "refine_permutations": a.Parameters.RefinePermutations, "frozen_files": a.Files}, "frozen_candidate_count": len(a.Candidates), "physical_blocks": len(a.Blocks), "summaries": a.Summaries}
	yb, e := yaml.Marshal(y)
	if e != nil {
		return e
	}
	if e = os.WriteFile(filepath.Join(out, "token_relation_validation.yaml"), yb, 0644); e != nil {
		return e
	}
	if e = writeReport(filepath.Join(out, "token_relation_validation_report.md"), a); e != nil {
		return e
	}
	for _, name := range []string{"directional_cross_block_consistency.svg", "distance_profile_cross_block_stability.svg", "structural_pair_cross_block_stability.svg", "relation_block_coverage.svg", "metadata_transfer_heatmap.svg"} {
		if e = writePlot(filepath.Join(out, "plots", name), name, a); e != nil {
			return e
		}
	}
	return nil
}
func summaryHeader() []string {
	return []string{"candidate_id", "family", "token_a", "token_b", "eligible_blocks", "positive_blocks", "negative_blocks", "neutral_blocks", "joint_classes", "currier_classes", "hands", "sign_consistency", "weighted_direction_mean", "unweighted_direction_mean", "between_block_variance", "median_enrichment", "profile_mean", "profile_median", "profile_min", "profile_sd", "fraction_above_frozen_threshold", "mean_pairwise_weighted_overlap", "tested_heldout", "successful_heldout", "transfer_success", "raw_empirical_p", "fdr_q", "control_percentile", "classification"}
}
func summaryRow(x RelationSummary) []string {
	return []string{x.CandidateID, x.Family, x.A, x.B, i(x.EligibleBlocks), i(x.PositiveBlocks), i(x.NegativeBlocks), i(x.NeutralBlocks), i(x.JointClasses), i(x.CurrierClasses), i(x.Hands), f(x.SignConsistency), f(x.WeightedDirection), f(x.UnweightedDirection), f(x.BetweenBlockVariance), f(x.MedianEnrichment), f(x.ProfileMean), f(x.ProfileMedian), f(x.ProfileMin), f(x.ProfileSD), f(x.FractionAboveThreshold), f(x.ProfileOverlapMean), i(x.TestedHeldout), i(x.SuccessfulHeldout), f(x.TransferSuccess), f(x.RawP), f(x.FDRQ), f(x.ControlPercentile), x.Classification}
}

func writeReport(path string, a Analysis) error {
	counts := map[string]int{}
	testable, two, three, joint, cur, hand := 0, 0, 0, 0, 0, 0
	for _, x := range a.Summaries {
		counts[x.Classification]++
		if x.EligibleBlocks > 0 {
			testable++
		}
		if x.EligibleBlocks >= 2 {
			two++
		}
		if x.EligibleBlocks >= 3 {
			three++
		}
		if x.JointClasses >= 2 {
			joint++
		}
		if x.CurrierClasses >= 2 {
			cur++
		}
		if x.Hands >= 2 {
			hand++
		}
	}
	var z strings.Builder
	fmt.Fprintf(&z, "# Cross-metadata token relation validation\n\n## Global accounting\n\n- Frozen candidates: %d\n- Testable candidates: %d\n- Candidates with at least 2 physical blocks: %d\n- Candidates with at least 3 physical blocks: %d\n- Candidates crossing joint classes: %d\n- Candidates crossing Currier classes: %d\n- Candidates crossing hands: %d\n- Unknown-metadata tokens excluded from primary evidence: %d\n\n", len(a.Candidates), testable, two, three, joint, cur, hand, a.UnknownTokens)
	z.WriteString("## Classification counts\n\n| Class | Count |\n|---|---:|\n")
	for _, k := range []string{"UNIVERSAL", "CURRIER_SPECIFIC", "HAND_SPECIFIC", "BLOCK_SPECIFIC", "WEAK"} {
		fmt.Fprintf(&z, "| %s | %d |\n", k, counts[k])
	}
	z.WriteString("\n## Previously strongest relations\n\nOnly relations actually present in the frozen inventory are shown.\n\n| Relation | Family | Blocks | Joint classes | Transfer | q | Classification |\n|---|---|---:|---:|---:|---:|---|\n")
	wanted := map[string]bool{"chedy\x00shedy": true, "qokedy\x00qokeedy": true, "or\x00aiin": true, "chol\x00daiin": true, "sho\x00daiin": true, "qokeedy\x00chedy": true}
	for _, x := range a.Summaries {
		k := x.A + "\x00" + x.B
		if !wanted[k] {
			k = x.B + "\x00" + x.A
		}
		if wanted[k] {
			rel := x.Sequence
			if rel == "" {
				arrow := " / "
				if x.Family == "directional" {
					arrow = " -> "
				}
				rel = x.A + arrow + x.B
			}
			fmt.Fprintf(&z, "| `%s` | %s | %d | %d | %.3f | %.4g | %s |\n", rel, x.Family, x.EligibleBlocks, x.JointClasses, x.TransferSuccess, x.FDRQ, x.Classification)
		}
	}
	z.WriteString("\n## Interpretation\n\n`UNIVERSAL` means a formally stable, cross-metadata transferable local relation under the pre-specified criteria. It does not imply grammar, cipher, natural language, operators, operands, or an algorithmic language. Pooled frequency is not independent replication; the primary unit here is a physical block.\n\n## Reproducibility\n\n")
	fmt.Fprintf(&z, "- Corpus SHA256: `%s`\n- Canonical token count: %d\n- Metadata-map SHA256: `%s`\n- Seed: %d\n- Initial permutations: %d\n- Refinement permutations: %d\n\n", a.CorpusSHA, a.TokenCount, a.MetadataSHA, a.Parameters.Seed, a.Parameters.Permutations, a.Parameters.RefinePermutations)
	z.WriteString("| Frozen input | Stored token count | SHA256 |\n|---|---:|---|\n")
	for _, x := range a.Files {
		fmt.Fprintf(&z, "| `%s` | %d | `%s` |\n", filepath.Base(x.Path), x.StoredTokenCount, x.SHA256)
	}
	z.WriteString("\nPhysical blocks are zero-based contiguous runs of one known Currier×hand state, exactly matching residual diagnostics. Primary directional eligibility requires both token counts ≥5 and ≥5 directional observations; primary structural support requires both counts ≥10, with ≥5 retained separately as descriptive evidence. Refinement is restricted to raw p<0.01, at least 3 eligible blocks, and at least 2 joint classes. Inputs made on the 38,887-token corpus contribute candidate identities and frozen settings only; every validation statistic is recomputed on the canonical corpus.\n")
	return os.WriteFile(path, []byte(z.String()), 0644)
}
func writePlot(path, title string, a Analysis) error {
	if strings.Contains(title, "heatmap") {
		return writeHeatmap(path, title, a)
	}
	var vals []RelationSummary
	metric := func(x RelationSummary) float64 { return float64(x.EligibleBlocks) }
	label := "eligible physical blocks"
	switch {
	case strings.HasPrefix(title, "directional"):
		for _, x := range a.Summaries {
			if x.Family == "directional" {
				vals = append(vals, x)
			}
		}
		metric = func(x RelationSummary) float64 { return x.SignConsistency }
		label = "sign consistency"
	case strings.HasPrefix(title, "distance"):
		for _, x := range a.Summaries {
			if x.Family == "distance-profile" {
				vals = append(vals, x)
			}
		}
		metric = func(x RelationSummary) float64 { return x.ProfileMedian }
		label = "median pairwise block JS similarity"
	case strings.HasPrefix(title, "structural"):
		for _, x := range a.Summaries {
			if x.Family == "structural" {
				vals = append(vals, x)
			}
		}
		metric = func(x RelationSummary) float64 { return x.ProfileMedian }
		label = "median within-block structural similarity"
	default:
		vals = append(vals, a.Summaries...)
	}
	sort.Slice(vals, func(i, j int) bool { return metric(vals[i]) > metric(vals[j]) })
	if len(vals) > 30 {
		vals = vals[:30]
	}
	w, h := 900, 500
	var z strings.Builder
	fmt.Fprintf(&z, "<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"%d\" height=\"%d\" viewBox=\"0 0 %d %d\"><rect width=\"100%%\" height=\"100%%\" fill=\"white\"/><text x=\"30\" y=\"30\" font-family=\"sans-serif\" font-size=\"18\">%s</text><text x=\"30\" y=\"52\" font-family=\"sans-serif\" font-size=\"12\">%s; top 30 frozen relations</text><line x1=\"50\" y1=\"450\" x2=\"870\" y2=\"450\" stroke=\"#333\"/>", w, h, w, h, html.EscapeString(title), html.EscapeString(label))
	if len(vals) > 0 {
		bw := 800 / len(vals)
		if bw < 2 {
			bw = 2
		}
		maxB := 1.
		for _, x := range vals {
			if metric(x) > maxB {
				maxB = metric(x)
			}
		}
		for n, x := range vals {
			bh := int(380 * metric(x) / maxB)
			fmt.Fprintf(&z, "<rect x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" fill=\"#4378bf\"><title>%s: %.4f %s</title></rect>", 55+n*bw, 450-bh, max(1, bw-1), bh, html.EscapeString(x.CandidateID), metric(x), html.EscapeString(label))
		}
	}
	z.WriteString("</svg>\n")
	return os.WriteFile(path, []byte(z.String()), 0644)
}

func writeHeatmap(path, title string, a Analysis) error {
	type cell struct {
		dimension, from, to string
		sum                 float64
		n                   int
	}
	m := map[string]*cell{}
	for _, x := range a.MetadataTransfers {
		k := x.Dimension + "\x00" + x.Training + "\x00" + x.Heldout
		z := m[k]
		if z == nil {
			z = &cell{dimension: x.Dimension, from: x.Training, to: x.Heldout}
			m[k] = z
		}
		z.sum += x.Fraction
		z.n++
	}
	var cells []*cell
	for _, x := range m {
		cells = append(cells, x)
	}
	sort.Slice(cells, func(i, j int) bool {
		return cells[i].dimension+cells[i].from+cells[i].to < cells[j].dimension+cells[j].from+cells[j].to
	})
	var z strings.Builder
	z.WriteString("<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"1000\" height=\"600\"><rect width=\"100%\" height=\"100%\" fill=\"white\"/>")
	fmt.Fprintf(&z, "<text x=\"25\" y=\"30\" font-family=\"sans-serif\" font-size=\"18\">%s</text>", html.EscapeString(title))
	for n, x := range cells {
		if n >= 80 {
			break
		}
		v := x.sum / float64(x.n)
		red := int(230 * (1 - v))
		green := int(190 * v)
		col := n % 10
		row := n / 10
		fmt.Fprintf(&z, "<rect x=\"%d\" y=\"%d\" width=\"88\" height=\"58\" fill=\"rgb(%d,%d,90)\" stroke=\"white\"><title>%s %s -&gt; %s: %.4f</title></rect>", 25+col*95, 55+row*65, red, green, html.EscapeString(x.dimension), html.EscapeString(x.from), html.EscapeString(x.to), v)
	}
	z.WriteString("</svg>\n")
	return os.WriteFile(path, []byte(z.String()), 0644)
}
