package conditionalregime

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func f(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }
func ints(x []int) string {
	s := make([]string, len(x))
	for i, v := range x {
		s[i] = strconv.Itoa(v)
	}
	return strings.Join(s, ",")
}

func tsv(path, header string, rows []string) error {
	fh, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fh.Close()
	w := bufio.NewWriter(fh)
	defer w.Flush()
	fmt.Fprintln(w, header)
	for _, r := range rows {
		fmt.Fprintln(w, r)
	}
	return nil
}

func writeYAML(path string, v any) error {
	b, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

func writeAll(c Config, r *runResult) error {
	if err := os.MkdirAll(filepath.Join(c.OutputDir, "plots"), 0755); err != nil {
		return err
	}
	writers := []func() error{
		func() error {
			return writeInventory(filepath.Join(c.OutputDir, "conditional_class_inventory.tsv"), r.Inventory)
		},
		func() error {
			return writeWithinRegimes(filepath.Join(c.OutputDir, "within_class_regimes.tsv"), r.WithinRegimes)
		},
		func() error {
			return writeWithinStability(filepath.Join(c.OutputDir, "within_class_stability.tsv"), r.WithinStability, r.CrossBlock)
		},
		func() error {
			return writeYAML(filepath.Join(c.OutputDir, "within_class_permutations.yaml"), withinPermutationsDoc(c, r))
		},
		func() error {
			return writeResidualAssignments(filepath.Join(c.OutputDir, "residual_cluster_assignments.tsv"), r.ResidualWindows, r.ResidualLabels, r.ResidualScale, r.ResidualK)
		},
		func() error {
			return writeResidualSummary(filepath.Join(c.OutputDir, "residual_cluster_summary.tsv"), r.ResidualSummary)
		},
		func() error {
			return writeResidualAssociation(filepath.Join(c.OutputDir, "residual_metadata_association.tsv"), r.ResidualAssoc)
		},
		func() error {
			return writeYAML(filepath.Join(c.OutputDir, "residual_permutations.yaml"), residualPermutationsDoc(c, r))
		},
		func() error {
			return writeConditionalBoundaries(filepath.Join(c.OutputDir, "conditional_stable_boundaries.tsv"), r.Boundaries)
		},
		func() error {
			return writeTransitionMatrix(filepath.Join(c.OutputDir, "residual_transition_matrix.tsv"), r.Transitions)
		},
		func() error {
			return writeResidualCandidates(filepath.Join(c.OutputDir, "residual_regime_candidates.tsv"), r.ResidualCandRows)
		},
		func() error {
			return writeYAML(filepath.Join(c.OutputDir, "conditional_regime_analysis.yaml"), analysisDoc(c, r))
		},
		func() error {
			return os.WriteFile(filepath.Join(c.OutputDir, "conditional_regime_report.md"), []byte(buildReport(c, r)), 0644)
		},
		func() error { return writePlots(c.OutputDir, r) },
	}
	for _, w := range writers {
		if err := w(); err != nil {
			return err
		}
	}
	return nil
}

func writeInventory(path string, inv []ClassInfo) error {
	rows := make([]string, len(inv))
	for i, ci := range inv {
		rows[i] = fmt.Sprintf("%s\t%s\t%s\t%s\t%d\t%d\t%d\t%s\t%t", ci.Class.Currier, ci.Class.Hand, string(ci.Class.Scheme), ci.Class.Label(), ci.TotalTokens, ci.BlockCount, ci.LargestBlock, f(ci.MedianBlock), ci.Eligible)
	}
	return tsv(path, "currier\thand\tscheme\tjoint_class\ttotal_tokens\tblock_count\tlargest_block\tmedian_block\teligible", rows)
}

func writeWithinRegimes(path string, rows []WithinClassRegime) error {
	out := make([]string, len(rows))
	for i, w := range rows {
		out[i] = fmt.Sprintf("%s\t%s\t%d\t%s\t%d\t%d\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%t",
			string(w.Scheme), w.Class, w.WindowSize, w.Method, w.K, w.AllowedKMax, w.Windows,
			f(w.Silhouette), f(w.MedoidSeparation), f(w.WithinDispersion), f(w.BetweenDispersion),
			f(w.ClusterSizeEntropy), f(w.SmallestClusterFrac), ints(w.ClusterSizes), w.Diagnostic)
	}
	return tsv(path, "scheme\tclass\twindow_size\tmethod\tk\tallowed_k_max\twindows\tsilhouette\tmedoid_separation\twithin_dispersion\tbetween_dispersion\tcluster_size_entropy\tsmallest_cluster_fraction\tcluster_sizes\tdiagnostic_only", out)
}

func writeWithinStability(path string, stability []WithinClassStability, cross []CrossBlockRecurrence) error {
	byKey := map[string]WithinClassStability{}
	for _, s := range stability {
		byKey[s.Class.Label()+"|"+string(s.Class.Scheme)+"|"+strconv.Itoa(s.WindowSize)+"|"+s.Method] = s
	}
	rows := []string{}
	seen := map[string]bool{}
	for _, cb := range cross {
		key := cb.Class + "|" + string(cb.Scheme) + "|" + strconv.Itoa(cb.WindowSize) + "|" + cb.Method
		seen[key] = true
		s := byKey[key]
		rows = append(rows, fmt.Sprintf("%s\t%s\t%d\t%s\t%d\t%d\t%s\t%d\t%d\t%d\t%s\t%d\t%s\t%s",
			string(cb.Scheme), cb.Class, cb.WindowSize, cb.Method, cb.K, cb.Cluster, f(s.Score), s.Folds,
			cb.BlocksContaining, cb.TotalBlocks, f(cb.BlockFraction), cb.EligibleBlocksForScale, f(cb.RecurrenceStrength), f(cb.CrossBlockSimilarity)))
	}
	for key, s := range byKey {
		if seen[key] {
			continue
		}
		rows = append(rows, fmt.Sprintf("%s\t%s\t%d\t%s\t%d\t-1\t%s\t%d\t0\t0\t0\t0\t0\t0", string(s.Class.Scheme), s.Class.Label(), s.WindowSize, s.Method, s.K, f(s.Score), s.Folds))
	}
	sort.Strings(rows)
	return tsv(path, "scheme\tclass\twindow_size\tmethod\tk\tcluster\theld_out_separation\tcv_folds\tblocks_containing\ttotal_blocks\tblock_fraction\teligible_blocks_for_scale\trecurrence_strength\tcross_block_similarity", rows)
}

func withinPermutationsDoc(c Config, r *runResult) map[string]any {
	byKey := map[string]any{}
	for _, cand := range r.Candidates {
		key := cand.Class.Label() + "/" + string(cand.Class.Scheme) + "/w" + strconv.Itoa(cand.WindowSize) + "/" + cand.Method
		byKey[key] = map[string]any{
			"k": cand.K, "refined": cand.Refined, "observed_silhouette": cand.Stats.Observed,
			"null_mean": cand.Stats.NullMean, "null_sd": cand.Stats.NullSD, "null_p95": cand.Stats.NullP95,
			"null_p99": cand.Stats.NullP99, "null_max": cand.Stats.NullMax, "effect_size": cand.Stats.EffectSize,
			"exceedances": cand.Stats.Exceedances, "empirical_p": cand.Stats.EmpiricalP, "permutations": cand.Stats.Permutations,
		}
	}
	return map[string]any{
		"seed": c.Seed, "primary_permutations": c.Permutations, "refinement_permutations": refinementPermutations,
		"refinement_rule": "empirical_p < 0.01 and effect_size >= 2.0 at the primary pass; top 5 qualifying combinations by effect size are refined",
		"null_model":      "Null A: tokens shuffled independently within each physical block, preserving unigram frequencies, block lengths and window lengths",
		"statistic":       "silhouette of the method's best-observed K, refit on shuffled data at that same K",
		"candidates":      byKey,
	}
}
