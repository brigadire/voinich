package conditionalregime

import (
	"fmt"
)

func writeResidualAssignments(path string, rw []ResidualWindow, labels []int, windowSize, k int) error {
	rows := make([]string, len(rw))
	for i, w := range rw {
		label := -1
		if i < len(labels) {
			label = labels[i]
		}
		rows[i] = fmt.Sprintf("%s\t%s\t%s\t%d\t%d\t%d\t%d\t%d", w.Class.Currier, w.Class.Hand, w.Class.Label(), w.BlockIndex, w.AbsStart, w.AbsEnd, windowSize, label)
	}
	_ = k
	return tsv(path, "currier\thand\tjoint_class\tblock_index\tabs_start\tabs_end\twindow_size\tresidual_cluster", rows)
}

func writeResidualSummary(path string, rows []ResidualClusterSummary) error {
	out := make([]string, len(rows))
	for i, s := range rows {
		out[i] = fmt.Sprintf("%d\t%s\t%d\t%s\t%d\t%s\t%s\t%s\t%s\t%s", s.WindowSize, s.Method, s.K, s.Representation, s.Windows,
			f(s.Silhouette), f(s.WithinDispersion), f(s.BetweenDispersion), f(s.ClusterSizeEntropy), f(s.SmallestClusterFrac))
	}
	return tsv(path, "window_size\tmethod\tk\trepresentation\twindows\tsilhouette\twithin_dispersion\tbetween_dispersion\tcluster_size_entropy\tsmallest_cluster_fraction", out)
}

func writeResidualAssociation(path string, rows []ResidualMetadataAssociation) error {
	out := make([]string, len(rows))
	for i, a := range rows {
		out[i] = fmt.Sprintf("%d\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s", a.WindowSize, a.K, a.Method, a.Representation, a.Metadata,
			f(a.ResidualNMI), f(a.ResidualARI), f(a.OriginalNMI), f(a.InformationReduction), a.OriginalSource)
	}
	return tsv(path, "window_size\tk\tmethod\trepresentation\tmetadata\tresidual_nmi\tresidual_ari\toriginal_nmi\tinformation_reduction\toriginal_source", out)
}

func writeResidualCandidates(path string, rows []ResidualCandidate) error {
	out := make([]string, len(rows))
	for i, c := range rows {
		out[i] = fmt.Sprintf("%d\t%s\t%d\t%s\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%s\t%s", c.WindowSize, c.Method, c.K, c.Representation, c.Cluster, c.Size,
			c.CurrierClasses, c.Hands, c.JointClasses, c.PhysicalBlocks, c.TotalPhysicalBlocks, f(c.MetadataEntropy), f(c.CompositeScore))
	}
	return tsv(path, "window_size\tmethod\tk\trepresentation\tcluster\tsize\tcurrier_classes\thands\tjoint_classes\tphysical_blocks\ttotal_physical_blocks\tmetadata_entropy\tcomposite_score", out)
}

func writeConditionalBoundaries(path string, rows []ConditionalStableBoundary) error {
	out := make([]string, len(rows))
	for i, b := range rows {
		out[i] = fmt.Sprintf("%s\t%d\t%d\t%d\t%d\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s", b.Class.Label(), b.BlockIndex, b.BlockLen, b.Position, b.AbsolutePosition,
			b.SupportCount, f(b.SupportFraction), f(b.MeanJumpStrength), f(b.MaxJumpStrength), f(b.PositionUncertainty), b.SignatureToken, b.SignatureDirection, f(b.SignatureMagnitude))
	}
	return tsv(path, "joint_class\tblock_index\tblock_len\tposition_in_block\tabsolute_position\tscale_support_count\tscale_support_fraction\tmean_jump_strength\tmax_jump_strength\tposition_uncertainty\tsignature_token\tsignature_direction\tsignature_magnitude", out)
}

func writeTransitionMatrix(path string, cells []ResidualTransitionCell) error {
	out := make([]string, len(cells))
	for i, c := range cells {
		e := c.Stats
		out[i] = fmt.Sprintf("%d\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s\t%d", c.From, c.To, f(e.Observed), f(e.NullMean), f(e.NullSD), f(e.NullP95), f(e.NullP99), f(e.NullMax), e.Exceedances, f(e.EmpiricalP), e.Permutations)
	}
	return tsv(path, "from_cluster\tto_cluster\tobserved_count\tnull_mean\tnull_sd\tnull_p95\tnull_p99\tnull_max\texceedances\tempirical_p\tpermutations", out)
}

func residualPermutationsDoc(c Config, r *runResult) map[string]any {
	corr := map[string]any{}
	for key, s := range r.ResidualCorrection {
		corr[key] = map[string]any{
			"observed": s.Observed, "null_mean": s.NullMean, "null_sd": s.NullSD, "null_p95": s.NullP95,
			"null_p99": s.NullP99, "null_max": s.NullMax, "effect_size": s.EffectSize,
			"exceedances": s.Exceedances, "empirical_p": s.EmpiricalP, "permutations": s.Permutations,
		}
	}
	return map[string]any{
		"seed": c.Seed, "permutations": c.Permutations,
		"search_space":          map[string]any{"window_size": c.ResidualWindowSizes, "k": []int{c.KMin, c.KMaxResidual}, "primary_method": "k_medoids", "secondary_method": "hierarchical (corrected separately, task19 section 42)"},
		"statistic":             "max silhouette over the complete scale x K residual search space, raw residual representation",
		"null_model":            "Null A applied to every eligible joint class's physical blocks with one shared shuffled-corpus realization per replicate, reused unchanged across the whole scale x K sweep",
		"global_max_correction": corr,
	}
}

func analysisDoc(c Config, r *runResult) map[string]any {
	return map[string]any{
		"corpus":   map[string]any{"path": c.CorpusPath, "sha256": r.CorpusSHA256, "token_count": r.TokenCount},
		"metadata": map[string]any{"path": c.TokenMetadataMap, "sha256": r.MetadataSHA256, "excluded_unknown_tokens": r.ExcludedTokens},
		"parameters": map[string]any{
			"window_sizes": c.WindowSizes, "residual_window_sizes": c.ResidualWindowSizes,
			"min_class_tokens": c.MinClassTokens, "min_block_tokens": c.MinBlockTokens,
			"k_min": c.KMin, "k_max_within": c.KMaxWithin, "k_max_residual": c.KMaxResidual,
			"permutations": c.Permutations, "refinement_permutations": refinementPermutations, "seed": c.Seed,
		},
		"eligible_classes": eligibleClassLabels(r.Inventory),
		"outcome":          classifyOutcome(r),
	}
}

func eligibleClassLabels(inv []ClassInfo) map[string][]string {
	out := map[string][]string{"joint": {}, "currier_only": {}, "hand_only": {}}
	for _, ci := range inv {
		if !ci.Eligible {
			continue
		}
		out[string(ci.Class.Scheme)] = append(out[string(ci.Class.Scheme)], ci.Class.Label())
	}
	return out
}
