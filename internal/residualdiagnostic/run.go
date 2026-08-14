package residualdiagnostic

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func RunAndWrite(c Config) error {
	if c.ProgressWriter == nil && !c.Quiet {
		c.ProgressWriter = os.Stderr
	}
	p := newProgress(c.ProgressWriter)
	p.begin(1, "Validating frozen inputs")
	for _, name := range []string{"residual_cluster_assignments.tsv", "residual_cluster_summary.tsv", "residual_metadata_association.tsv", "residual_regime_candidates.tsv", "residual_permutations.yaml", "conditional_regime_analysis.yaml"} {
		if _, err := os.Stat(filepath.Join(c.ConditionalDir, name)); err != nil {
			return fmt.Errorf("required frozen result %s: %w", name, err)
		}
	}
	tokens, corpusSHA, err := readCorpus(c.CorpusPath)
	if err != nil {
		return err
	}
	m, metadataSHA, err := readMetadata(c.MetadataPath)
	if err != nil {
		return err
	}
	if len(tokens) != len(m.Currier) {
		return fmt.Errorf("corpus/metadata token count mismatch: %d != %d", len(tokens), len(m.Currier))
	}
	assign, eligible, err := readAssignments(filepath.Join(c.ConditionalDir, "residual_cluster_assignments.tsv"), c.WindowSize, c.K)
	if err != nil {
		return err
	}
	originalCurrier, originalHand, err := readFrozenOriginalNMI(filepath.Join(c.ConditionalDir, "residual_metadata_association.tsv"), c.WindowSize, c.K)
	if err != nil {
		return err
	}
	bs := buildBlocks(m)
	p.update(1, 1, "Validating frozen inputs")
	p.begin(2, "Reconstructing leakage-safe folds and whitening")
	w, foldRows, err := loadWindows(tokens, m, bs, assign, eligible, c.WindowSize)
	if err != nil {
		return err
	}
	p.update(1, 1, "Reconstructing leakage-safe folds and whitening")
	r := &results{CorpusSHA: corpusSHA, MetadataSHA: metadataSHA, TokenCount: len(tokens), OriginalCurrierNMI: originalCurrier, OriginalHandNMI: originalHand, Windows: w, Folds: foldRows, Files: map[string][][]string{}, Summary: map[string]any{}}
	p.begin(3, "Verifying residual centering")
	r.Files["residual_centering_diagnostics.tsv"] = withHeader([]string{"window_size", "joint_class", "split", "l1_norm", "l2_norm", "max_absolute_component", "mean_absolute_component"}, centeringRows(w, foldRows))
	var fr [][]string
	for _, d := range foldRows {
		fr = append(fr, []string{strconv.Itoa(d.WindowSize), strconv.Itoa(d.Fold), d.Joint, strconv.Itoa(d.TrainWindows), strconv.Itoa(d.TestWindows), f(d.TrainMean.L2), f(d.TestMean.L2), f(d.TrainMean.L1), f(d.TestMean.L1)})
	}
	r.Files["residual_fold_diagnostics.tsv"] = withHeader([]string{"window_size", "fold", "joint_class", "training_windows", "test_windows", "mean_train_residual_l2", "mean_test_residual_l2", "mean_train_residual_l1", "mean_test_residual_l1"}, fr)
	p.update(1, 1, "Verifying residual centering")
	p.begin(4, "Identifying the frozen K=2 clusters")
	existing := clustersOf(w)
	r.Files["winning_cluster_composition.tsv"] = withHeader([]string{"cluster", "dimension", "value", "count", "fraction", "position_min", "position_q25", "position_median", "position_q75", "position_max", "window_coverage_start", "window_coverage_end", "physical_block_interval_start", "physical_block_interval_end"}, compositionRows(w, c.K))
	runs := extractRuns(w, existing)
	var runRows [][]string
	for _, x := range runs {
		runRows = append(runRows, []string{strconv.Itoa(x.Cluster), strconv.Itoa(x.Count), strconv.Itoa(x.Largest), f(x.Median), f(x.LargestFraction), strconv.FormatBool(x.LargestFraction >= .8)})
	}
	r.Files["winning_cluster_runs.tsv"] = withHeader([]string{"cluster", "run_count", "largest_run_windows", "median_run_windows", "largest_run_fraction", "nearly_single_interval"}, runRows)
	p.update(1, 1, "Identifying the frozen K=2 clusters")
	p.begin(5, "Measuring residual dispersion and covariance")
	disp := dispersions(w)
	var dr, cd [][]string
	for _, d := range disp {
		dr = append(dr, []string{strconv.Itoa(c.WindowSize), d.joint, strconv.Itoa(d.n), f(d.total), f(d.mean), f(d.median), f(d.trace), f(d.effective)})
	}
	for i := 0; i < len(disp); i++ {
		for j := i + 1; j < len(disp); j++ {
			ratio, frob, diag := covarianceDistance(disp[i], disp[j])
			cd = append(cd, []string{strconv.Itoa(c.WindowSize), disp[i].joint, disp[j].joint, f(ratio), f(frob), f(diag)})
		}
	}
	r.Files["residual_dispersion_by_metadata.tsv"] = withHeader([]string{"window_size", "joint_class", "windows", "total_variance", "mean_feature_variance", "median_feature_variance", "covariance_trace", "effective_dimensionality"}, dr)
	r.Files["residual_covariance_distances.tsv"] = withHeader([]string{"window_size", "joint_class_a", "joint_class_b", "variance_ratio_a_over_b", "covariance_frobenius_distance", "diagonal_covariance_distance"}, cd)
	p.update(1, 1, "Measuring residual dispersion and covariance")
	p.begin(6, "Cross-validating metadata classifiers")
	residual := vectorsOf(w, "residual")
	blocks := labelsOf(w, "block")
	secondOrder := secondOrderFeatures(residual)
	residualGram, secondOrderGram := gramMatrix(residual), gramMatrix(secondOrder)
	var classRows, confRows [][]string
	classSummary := map[string]any{}
	targets := []string{"currier", "hand", "joint"}
	for ti, target := range targets {
		truth := labelsOf(w, target)
		for _, model := range []string{"multinomial_logistic", "second_order_logistic"} {
			features := residual
			gram := residualGram
			if model == "second_order_logistic" {
				features = secondOrder
				gram = secondOrderGram
			}
			pred, ce := cvLogistic(features, gram, truth, blocks, c.Seed+int64(ti))
			bal, f1 := scoreClassification(truth, pred)
			ex := 0
			for pi := 0; pi < c.Permutations; pi++ {
				yp := blockPermutation(truth, blocks, c.Seed+10_000+int64(ti*c.Permutations+pi))
				pb, _ := scoreClassification(yp, pred)
				if pb >= bal {
					ex++
				}
			}
			pp := float64(ex+1) / float64(c.Permutations+1)
			classRows = append(classRows, []string{target, model, "leave_physical_block_out", f(bal), f(f1), f(ce), f(pp), strconv.Itoa(c.Permutations)})
			confRows = append(confRows, confusionRows(target, model, truth, pred)...)
			if model == "multinomial_logistic" {
				classSummary[target+"_linear_balanced_accuracy"] = bal
			}
			p.update(ti*2+map[bool]int{true: 2, false: 1}[model == "second_order_logistic"], 6, "Cross-validating metadata classifiers")
		}
		maj, mce := majorityBaseline(truth)
		b, f1 := scoreClassification(truth, maj)
		classRows = append(classRows, []string{target, "majority_baseline", "leave_physical_block_out", f(b), f(f1), f(mce), "", "0"})
		freq, fce := frequencyBaseline(truth, c.Seed+int64(ti))
		b, f1 = scoreClassification(truth, freq)
		classRows = append(classRows, []string{target, "class_frequency_random", "leave_physical_block_out", f(b), f(f1), f(fce), "", "0"})
	}
	r.Files["residual_metadata_classification.tsv"] = withHeader([]string{"target", "model", "split", "balanced_accuracy", "macro_f1", "cross_entropy", "block_permutation_p", "permutations"}, classRows)
	r.Files["residual_metadata_confusion.tsv"] = withHeader([]string{"target", "model", "truth", "prediction", "count"}, confRows)
	p.begin(7, "Running norm-only diagnostic")
	normVec := make([]sparse, len(w))
	for i, x := range w {
		n := normOf(x.Residual)
		normVec[i] = sparse{"l1": n.L1, "l2": n.L2, "linf": n.Linf}
	}
	normLabels := cluster(normVec, c.K, c.Seed)
	nr := normRows(w)
	nr = append(nr, []string{"cluster_comparison", "ARI", strconv.Itoa(len(w)), f(adjustedRand(existing, normLabels)), "", ""}, []string{"cluster_comparison", "NMI", strconv.Itoa(len(w)), f(nmiInt(existing, normLabels)), "", ""})
	r.Files["residual_norm_diagnostics.tsv"] = withHeader([]string{"dimension", "value", "windows", "mean_l1", "mean_l2", "mean_linf"}, nr)
	p.update(1, 1, "Running norm-only diagnostic")
	p.begin(8, "Comparing original, centered and whitened representations")
	original := vectorsOf(w, "original")
	white := vectorsOf(w, "whitened")
	originalLabels := clusterRaw(original, c.K, c.Seed)
	whiteLabels := cluster(white, c.K, c.Seed)
	r.Representations = []representationRow{representation(w, "original_features", original, originalLabels), representation(w, "mean_residual", residual, existing), representation(w, "whitened_residual", white, whiteLabels)}
	r.Representations[0].Silhouette = sampledSilhouetteMetric(original, originalLabels, jsDistance)
	var repr [][]string
	for _, x := range r.Representations {
		repr = append(repr, []string{x.Name, f(x.Silhouette), f(x.CurrierNMI), f(x.HandNMI), f(x.JointNMI), f(x.BlockNMI), strconv.Itoa(x.ClusterRunCount), f(x.LargestRunFraction)})
	}
	r.Files["representation_comparison.tsv"] = withHeader([]string{"representation", "silhouette", "currier_nmi", "hand_nmi", "joint_nmi", "physical_block_nmi", "cluster_run_count", "largest_run_fraction"}, repr)
	p.update(1, 1, "Comparing original, centered and whitened representations")
	p.begin(9, "Testing block recurrence and position effects")
	r.Files["residual_block_association.tsv"] = withHeader([]string{"scope", "nmi", "ari", "homogeneity", "completeness", "block_permutation_p", "permutations"}, blockAssociationRows(w, existing, c.Permutations, c.Seed))
	r.Files["residual_block_recurrence.tsv"] = withHeader([]string{"held_out_block", "windows", "mean_assignment_confidence", "mean_medoid_distance", "clusters_above_reference_threshold", "both_regimes_recur"}, recurrenceRows(w, c.K, c.Seed))
	r.Files["residual_position_diagnostics.tsv"] = withHeader([]string{"row_type", "window_index", "cluster", "absolute_position_bin", "normalized_block_position", "absolute_midpoint_or_position_bin_nmi"}, positionRows(w, existing, len(tokens), bs))
	p.update(1, 1, "Testing block recurrence and position effects")
	r.Summary = buildSummary(c, r, classSummary, normLabels, existing, runs)
	p.begin(10, "Writing diagnostic outputs and plots")
	if err := writeResults(c, r); err != nil {
		return err
	}
	p.update(1, 1, "Writing diagnostic outputs and plots")
	return nil
}

func withHeader(header []string, rows [][]string) [][]string {
	return append([][]string{header}, rows...)
}

func buildSummary(c Config, r *results, classification map[string]any, norm, existing []int, runs []runInfo) map[string]any {
	maxTrain, maxTest := 0., 0.
	for _, d := range r.Folds {
		maxTrain = math.Max(maxTrain, d.TrainMean.L2)
		maxTest = math.Max(maxTest, d.TestMean.L2)
	}
	recur := 0
	rows := r.Files["residual_block_recurrence.tsv"]
	for _, row := range rows[1:] {
		if row[5] == "true" {
			recur++
		}
	}
	white := r.Representations[2]
	outcomes := []string{}
	if maxTrain > 1e-9 {
		outcomes = append(outcomes, "centering implementation error: training residual means are not numerical zero")
	} else {
		if maxTest > 10*math.Max(maxTrain, 1e-12) {
			outcomes = append(outcomes, "held-out residual means drift across physical blocks")
		}
		if recur == 0 {
			outcomes = append(outcomes, "the apparent residual regime is block-specific and not cross-block recurrent")
		}
		if white.CurrierNMI < .25 && white.HandNMI < .25 {
			outcomes = append(outcomes, "whitening removes metadata association, consistent with a second-order effect")
		} else {
			outcomes = append(outcomes, "metadata survives leakage-safe whitening, so the current conditioning model remains insufficient")
		}
	}
	outcome := strings.Join(outcomes, "; ")
	return map[string]any{"corpus": map[string]any{"path": c.CorpusPath, "sha256": r.CorpusSHA, "token_count": r.TokenCount, "canonical_sha256_expected": "360d99583145ec549b80edfafdc3f93534f3a11b85a0d52997ba8425e92b87c2", "canonical_match": r.CorpusSHA == "360d99583145ec549b80edfafdc3f93534f3a11b85a0d52997ba8425e92b87c2"}, "metadata": map[string]any{"path": c.MetadataPath, "sha256": r.MetadataSHA}, "parameters": map[string]any{"window_size": c.WindowSize, "k": c.K, "method": "k_medoids", "permutations": c.Permutations, "seed": c.Seed, "covariance_shrinkage": map[string]float64{"full_covariance": .9, "diagonal": .1}, "eigenvalue_floor_relative": 1e-6, "split": "leave-physical-block-out; three contiguous folds only for single-block classes"}, "frozen_global_baseline": map[string]float64{"currier_nmi": r.OriginalCurrierNMI, "hand_nmi": r.OriginalHandNMI}, "centering": map[string]float64{"maximum_training_residual_mean_l2": maxTrain, "maximum_held_out_residual_mean_l2": maxTest}, "classification": classification, "norm_only": map[string]float64{"ari_with_existing": adjustedRand(existing, norm), "nmi_with_existing": nmiInt(existing, norm)}, "cross_block": map[string]any{"held_out_blocks_with_both_regimes": recur}, "decision": outcome}
}

func sortedKeys(m map[string][][]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
