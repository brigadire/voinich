package conditionalregime

import (
	"fmt"
	"sort"
	"strings"
)

// outcomeThreshold is the conventional significance level used only to
// phrase the interpretation; every raw empirical_p is reported regardless.
const outcomeThreshold = 0.05

// classifyOutcome applies task19 Part F's three outcomes mechanically to the
// computed statistics: it never redefines them after seeing the numbers.
func classifyOutcome(r *runResult) map[string]any {
	sigWithin := 0
	for _, c := range r.Candidates {
		if c.Stats.EmpiricalP < outcomeThreshold {
			sigWithin++
		}
	}
	residSig := false
	for _, s := range r.ResidualCorrection {
		if s.EmpiricalP < outcomeThreshold {
			residSig = true
		}
	}
	crossBlockRecurring := 0
	for _, cb := range r.CrossBlock {
		if cb.BlocksContaining >= 2 {
			crossBlockRecurring++
		}
	}
	crossMetadataRecurring := 0
	for _, c := range r.ResidualCandRows {
		if c.JointClasses >= 2 {
			crossMetadataRecurring++
		}
	}
	lowResidualAssociation := len(r.ResidualAssoc) > 0
	for _, a := range r.ResidualAssoc {
		if a.InformationReduction < 0.3 {
			lowResidualAssociation = false
		}
	}
	label := "B"
	switch {
	case sigWithin == 0 && !residSig && crossBlockRecurring == 0:
		label = "A"
	case residSig && crossBlockRecurring > 0 && crossMetadataRecurring > 0 && lowResidualAssociation:
		label = "C"
	}
	return map[string]any{
		"label": label, "within_class_significant_candidates": sigWithin, "residual_global_significant": residSig,
		"clusters_recurring_across_blocks": crossBlockRecurring, "residual_clusters_recurring_across_joint_classes": crossMetadataRecurring,
		"low_residual_metadata_association": lowResidualAssociation,
	}
}

func outcomeNarrative(label string) string {
	switch label {
	case "A":
		return "**Outcome A — structure disappears.** Within-class clustering did not exceed the Null-A block-shuffle control, held-out stability was low, and no residual cluster recurred across multiple physical blocks. The previously observed blind distributional structure is largely explainable by Currier/hand heterogeneity."
	case "C":
		return "**Outcome C — reproducible residual structure.** After conditioning, residual clusters remain significant against Null-A/Null-B controls, recur across multiple physical blocks and multiple Currier/hand classes, and retain low residual association with the conditioning metadata. The corpus contains reproducible distributional organization not explained by tested Currier/hand metadata."
	default:
		return "**Outcome B — weak/inconclusive residual structure.** Some within-class or residual clusters reached significance, but recurrence was limited to one block, cross-metadata recurrence was weak, or significance was marginal. Evidence for additional organization beyond Currier/hand is weak or inconclusive."
	}
}

func buildReport(c Config, r *runResult) string {
	var b strings.Builder
	outcome := classifyOutcome(r)
	fmt.Fprintf(&b, "# Conditional residual distributional structure\n\n")
	fmt.Fprintf(&b, "Main question: after conditioning on Currier and Davis hand, does reproducible distributional structure remain in the corpus?\n\n")
	fmt.Fprintf(&b, "## Reproducibility\n\n- corpus: `%s`, SHA256 `%s`, %d tokens\n- metadata map: `%s`, SHA256 `%s`\n- excluded (unknown-metadata) tokens: %d\n- window sizes (within-class): %v; residual window sizes: %v\n- min class tokens: %d; min block tokens: %d\n- K: %d..%d (within-class), %d..%d (residual)\n- permutations: %d primary, %d refinement (top %d qualifying candidates, empirical_p<0.01 and effect_size>=2.0 at the primary pass)\n- seed: %d\n\n",
		c.CorpusPath, r.CorpusSHA256, r.TokenCount, c.TokenMetadataMap, r.MetadataSHA256, r.ExcludedTokens,
		c.WindowSizes, c.ResidualWindowSizes, c.MinClassTokens, c.MinBlockTokens, c.KMin, c.KMaxWithin, c.KMin, c.KMaxResidual,
		c.Permutations, refinementPermutations, refinementCandidateLimit, c.Seed)

	b.WriteString("## Eligible joint Currier x hand classes\n\n")
	b.WriteString("| Class | Total tokens | Blocks | Largest block | Eligible |\n|---|---:|---:|---:|---|\n")
	for _, ci := range r.Inventory {
		if ci.Class.Scheme != SchemeJoint {
			continue
		}
		fmt.Fprintf(&b, "| %s | %d | %d | %d | %t |\n", ci.Class.Label(), ci.TotalTokens, ci.BlockCount, ci.LargestBlock, ci.Eligible)
	}

	b.WriteString("\n## Part A — within-class discovery\n\nPer-class, per-window-size, per-method significance against Null A (within-block token shuffle), at each method's best-observed K (`within_class_permutations.yaml` has every candidate; `conditional_class_inventory.tsv`/`within_class_regimes.tsv`/`within_class_stability.tsv` have the full diagnostics):\n\n")
	b.WriteString("| Class | Scheme | Window | Method | K | Silhouette | Effect size | Empirical p | Refined |\n|---|---|---:|---|---:|---:|---:|---:|---|\n")
	sorted := append([]WithinClassCandidate(nil), r.Candidates...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Stats.EffectSize > sorted[j].Stats.EffectSize })
	for _, cand := range sorted {
		fmt.Fprintf(&b, "| %s | %s | %d | %s | %d | %.3f | %.2f | %s | %t |\n", cand.Class.Label(), cand.Class.Scheme, cand.WindowSize, cand.Method, cand.K, cand.Stats.Observed, cand.Stats.EffectSize, formatP(cand.Stats.EmpiricalP), cand.Refined)
	}

	b.WriteString("\n## Part B — metadata-residualized feature space\n\nPooled k_medoids clustering over the raw residual R_w = X_w - mu_(C,H) (training-fold-only centering; no held-out leakage), across the frozen scale x K residual search space:\n\n")
	for key, s := range r.ResidualCorrection {
		fmt.Fprintf(&b, "- %s: global max silhouette observed %.3f, null mean %.3f, P95 %.3f, P99 %.3f, effect size %.2f, empirical p %s\n", key, s.Observed, s.NullMean, s.NullP95, s.NullP99, s.EffectSize, formatP(s.EmpiricalP))
	}
	fmt.Fprintf(&b, "\nWinning combination (k_medoids, raw): window_size=%d, K=%d.\n\n", r.ResidualScale, r.ResidualK)
	b.WriteString("Metadata independence check (residual vs original global max NMI, task18's frozen `cluster_metadata_global_summary.tsv`):\n\n")
	b.WriteString("| Metadata | Original NMI | Residual NMI | Residual ARI | Information reduction |\n|---|---:|---:|---:|---:|\n")
	for _, a := range r.ResidualAssoc {
		fmt.Fprintf(&b, "| %s | %.3f | %.3f | %.3f | %.2f |\n", a.Metadata, a.OriginalNMI, a.ResidualNMI, a.ResidualARI, a.InformationReduction)
	}
	b.WriteString("\nCross-metadata/cross-block recurrence and the descriptive composite ranking are in `residual_regime_candidates.tsv`; the composite score is an unweighted sum of coverage fractions and is not an inferential statistic.\n")

	b.WriteString("\n## Part C — conditional boundaries and residual transitions\n\n")
	fmt.Fprintf(&b, "%d change points found within controlled physical blocks (never crossing a metadata transition); %d recurring boundary types (grouped by dominant-token signature rather than absolute position, since absolute positions of different blocks are not comparable). Full detail in `conditional_stable_boundaries.tsv`.\n\n", len(r.Boundaries), len(r.RecurringTypes))
	sigTrans := 0
	for _, t := range r.Transitions {
		if t.Stats.EmpiricalP < outcomeThreshold {
			sigTrans++
		}
	}
	fmt.Fprintf(&b, "%d of %d residual R_i -> R_j transition cells are enriched relative to the Null-B within-block window-order shuffle at p<%.2f (`residual_transition_matrix.tsv`).\n\n", sigTrans, len(r.Transitions), outcomeThreshold)

	b.WriteString("## Interpretation\n\n")
	b.WriteString(outcomeNarrative(outcome["label"].(string)))
	b.WriteString("\n\nThis outcome describes only whether additional reproducible distributional structure exists beyond tested Currier/hand metadata. It does not identify instructions, operators, operands, grammar, recipes, a cipher, or natural language, even under Outcome C.\n\n")
	b.WriteString("**Main limitation.** Currier and Davis hand are metadata models, not the complete cause of distributional heterogeneity. \"Residual after Currier x hand\" means unexplained by these annotations, not independent of scribal or material effects in general.\n")
	return b.String()
}

func formatP(p float64) string {
	if p < 0.0001 {
		return "<0.0001"
	}
	return fmt.Sprintf("%.4f", p)
}
