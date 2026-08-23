package fingerprintv2

import "fmt"

func findCSMetric(metrics []CrossScaleMetric, id string) *CrossScaleMetric {
	for i := range metrics {
		if metrics[i].MetricID == id {
			return &metrics[i]
		}
	}
	return nil
}

func csVerdictFromStatus(id string, m *CrossScaleMetric, hypothesisNote string) CrossScaleVerdict {
	if m == nil {
		return CrossScaleVerdict{
			ID: id, Value: "NOT_APPLICABLE", NullComparison: "n/a", HeldOutResult: "n/a", PartitionStability: "n/a", Sensitivity: "n/a",
			Limitations: hypothesisNote + " metric was not computed for this corpus.",
		}
	}
	value := m.Status
	if value == "" {
		value = "INCONCLUSIVE"
	}
	heldOut := "not attached to this metric"
	if m.HeldOutPerformance != nil {
		hp := m.HeldOutPerformance
		if hp.Folds > 0 {
			heldOut = fmt.Sprintf("%d-fold grouped-by-folio: baseline log loss %.4f, model log loss %.4f, improvement %.4f (SD %.4f)", hp.Folds, hp.BaselineLogLoss, hp.ModelLogLoss, hp.Improvement, hp.ImprovementSD)
		} else {
			heldOut = "held-out validation attempted but " + hp.Note
		}
	}
	stability := "not separately assessed for this metric"
	if len(m.PartitionStability) > 0 {
		stable, total := 0, 0
		for _, run := range m.PartitionStability {
			if run.Status == "NOT_TESTABLE" || run.Status == "INSUFFICIENT_DATA" {
				continue
			}
			total++
			if run.Status == "GLOBAL" || run.Status == "PARTITION_SPECIFIC" {
				stable++
			}
		}
		stability = fmt.Sprintf("%d/%d testable strata classified GLOBAL or PARTITION_SPECIFIC", stable, total)
	}
	return CrossScaleVerdict{
		ID: id, Value: value, PrimaryStatistic: m.Hypothesis, EffectSize: m.EffectSize, EffectDefined: m.EffectDefined,
		NullComparison: fmt.Sprintf("%s: observed=%.6g, null_mean=%.6g, null_sd=%.6g, p=%.4g, q=%.4g", m.NullModel, m.ObservedStatistic, m.NullMean, m.NullSD, m.EmpiricalP, m.MultipleTestingAdjustment),
		HeldOutResult:  heldOut, PartitionStability: stability, Sensitivity: m.Sensitivity, Limitations: m.Limitations,
	}
}

// crossScaleVerdicts implements task77's twelve required final verdicts.
// Each is derived directly from already-computed metric records rather
// than recomputed, so the verdict can never disagree with the metric it
// summarizes.
func crossScaleVerdicts(m Metrics, eg EditFamilyValidation, cs CrossScaleResult) []CrossScaleVerdict {
	out := make([]CrossScaleVerdict, 0, 12)

	// 1. TASK75_RESULTS_REPRODUCED: a documentation-level verdict about
	// this implementation, not a per-corpus statistic; see TASK75_AUDIT.md.
	out = append(out, CrossScaleVerdict{
		ID: "TASK75_RESULTS_REPRODUCED", Value: "PARTIALLY_SUPPORTED",
		PrimaryStatistic: "code re-derivation + determinism re-run (TestSeededPipelineIsDeterministic) + first real-corpus execution",
		NullComparison:   "n/a (an infrastructure audit, not a statistical test)",
		HeldOutResult:    "n/a", PartitionStability: "n/a", Sensitivity: "n/a",
		Limitations: "Every LP1-LP4/EF1-EF4 formula and null construction was confirmed correct; four infrastructure deviations (IVTFF normalization, two O(vocabulary x corpus)/O(attempts x edges) performance defects, one C-GRAMMAR generator attempt-budget defect) were found and fixed because this is the first real-corpus run this pipeline has ever had. See TASK75_AUDIT.md for the itemized table.",
	})

	// 2. EDIT_FAMILIES_STRUCTURALLY_STABLE
	{
		testable, stable := 0, 0
		for _, run := range eg.StabilityRuns {
			if run.Status == "NOT_TESTABLE" || run.Status == "INSUFFICIENT_DATA" {
				continue
			}
			testable++
			if run.Status == "GLOBAL" {
				stable++
			} else if run.Status == "PARTITION_SPECIFIC" {
				stable++ // counted, but see limitations: not full GLOBAL stability
			}
		}
		value := "INCONCLUSIVE"
		if testable > 0 {
			share := float64(stable) / float64(testable)
			switch {
			case share >= 0.8:
				value = "SUPPORTED"
			case share >= 0.4:
				value = "PARTIALLY_SUPPORTED"
			default:
				value = "NOT_SUPPORTED"
			}
		}
		if eg.ConsensusStatus != "CONSENSUS_FAMILIES" && value == "SUPPORTED" {
			value = "PARTIALLY_SUPPORTED"
		}
		out = append(out, CrossScaleVerdict{
			ID: "EDIT_FAMILIES_STRUCTURALLY_STABLE", Value: value,
			PrimaryStatistic:   fmt.Sprintf("consensus_status=%s across %d testable perturbations", eg.ConsensusStatus, testable),
			NullComparison:     "n/a (a stability-across-perturbations claim, not a null-model comparison)",
			HeldOutResult:      "n/a (see EDIT_CROSS_SCALE_BLOCK_READY for held-out CS1 prediction)",
			PartitionStability: fmt.Sprintf("%d/%d testable perturbations GLOBAL/PARTITION_SPECIFIC (see edit_graph_validation.stability_runs)", stable, testable),
			Sensitivity:        "min_rule_support +/-1, rare-token cutoff, folio-half partition, community-detection seed (see stability_runs)",
			Limitations:        "PARTITION_SPECIFIC families are not distinguished from artifacts by this aggregate alone; task77 §2.3 warns a partition-specific effect should not be automatically called an artifact, but this verdict does not itself adjudicate that per family.",
		})
	}

	// 3. EDIT_FAMILIES_EXCEED_C_GRAMMAR_NULL (reuses EF4's own verdict)
	{
		value := map[string]string{
			"EXCEEDS_GRAMMAR_BOUND":         "SUPPORTED",
			"CONSISTENT_WITH_GRAMMAR_BOUND": "NOT_SUPPORTED",
			"MIXED":                         "PARTIALLY_SUPPORTED",
			"INCONCLUSIVE":                  "INCONCLUSIVE",
			"INSUFFICIENT_SUPPORT":          "INCONCLUSIVE",
		}[m.EF4.Verdict]
		if value == "" {
			value = "INCONCLUSIVE"
		}
		tests := ""
		for i, t := range m.EF4.Tests {
			if i > 0 {
				tests += "; "
			}
			tests += fmt.Sprintf("%s p=%.4g q=%.4g", t.ID, t.PValue, t.QValue)
		}
		out = append(out, CrossScaleVerdict{
			ID: "EDIT_FAMILIES_EXCEED_C_GRAMMAR_NULL", Value: value,
			PrimaryStatistic: "EF1 giant-component share, EF2 clustering, EF3 |Spearman(degree, log-frequency)| vs C-GRAMMAR",
			NullComparison:   "C-GRAMMAR (" + tests + ")", HeldOutResult: "n/a", PartitionStability: "inherited from EF4's own grammar-mode validity check",
			Sensitivity: "depends on which C-GRAMMAR mode(s) validated (grammar.validation)", Limitations: m.EF4.Reason,
		})
	}

	// 4. TRANSFORMATION_MOTIFS_STABLE
	{
		cc := eg.TransitiveMerge.CommunityVsComponents
		hub := eg.TransitiveMerge.HubDependence
		value := "INCONCLUSIVE"
		if len(eg.TransitiveMerge.Families) > 0 {
			switch {
			case cc.ARI >= 0.7 && hub.GiantShareDrop < 0.2:
				value = "SUPPORTED"
			case cc.ARI >= 0.4 || hub.GiantShareDrop < 0.4:
				value = "PARTIALLY_SUPPORTED"
			default:
				value = "NOT_SUPPORTED"
			}
		}
		out = append(out, CrossScaleVerdict{
			ID: "TRANSFORMATION_MOTIFS_STABLE", Value: value,
			PrimaryStatistic:   fmt.Sprintf("components-vs-label-propagation ARI=%.4g, NMI=%.4g", cc.ARI, cc.NMI),
			NullComparison:     "n/a (a definition-robustness comparison, not a null-model test)",
			HeldOutResult:      "n/a",
			PartitionStability: fmt.Sprintf("hub-removal (top %.0f%%) giant-share drop=%.4g (before=%.4g, after=%.4g)", hub.HubFraction*100, hub.GiantShareDrop, hub.GiantShareBefore, hub.GiantShareAfter),
			Sensitivity:        "depth-limited (1/2/3-hop) component counts in edit_graph_validation.transitive_merge.path_restrictions",
			Limitations:        "\"Motif\" here means the family/component definition's robustness to an alternative community-detection method and to hub removal, not a claim about specific recurring transformation sequences.",
		})
	}

	// 5-7, 9: direct from CS metrics
	simple := []struct {
		id     string
		metric string
		note   string
	}{
		{"FAMILY_LINE_POSITION_DEPENDENCE", "cs1/family-line-position", "CS1"},
		{"TRANSFORMATION_CONTEXT_DEPENDENCE", "cs2/prev-family-current-family", "CS2"},
		{"FAMILY_LOCUS_DEPENDENCE", "cs3/family-locus-type", "CS3"},
		{"STRUCTURAL_DISTANCE_EDIT_DISTANCE_DEPENDENCE", "cs7/edit-distance-x-structural-distance", "CS7"},
	}
	for _, s := range simple {
		out = append(out, csVerdictFromStatus(s.id, findCSMetric(cs.Metrics, s.metric), s.note))
	}

	// 8. FAMILY_FOLIO_REGIME_DEPENDENCE combines both CS4 sub-tests
	// (Currier and Section), since either alone would understate the
	// evidence; EF5's C-REGIME comparison is cross-referenced, not
	// re-tested, since it is the identical computation.
	{
		currier := findCSMetric(cs.Metrics, "cs4/family-currier")
		section := findCSMetric(cs.Metrics, "cs4/family-section")
		statuses := map[string]bool{}
		for _, mm := range []*CrossScaleMetric{currier, section} {
			if mm != nil {
				statuses[mm.Status] = true
			}
		}
		value := "NOT_APPLICABLE"
		switch {
		case statuses["SUPPORTED"]:
			value = "SUPPORTED"
		case statuses["PARTIALLY_SUPPORTED"]:
			value = "PARTIALLY_SUPPORTED"
		case statuses["NOT_SUPPORTED"]:
			value = "NOT_SUPPORTED"
		case statuses["INCONCLUSIVE"]:
			value = "INCONCLUSIVE"
		}
		describe := func(mm *CrossScaleMetric, label string) string {
			if mm == nil {
				return label + "=not computed"
			}
			return fmt.Sprintf("%s: status=%s p=%.4g q=%.4g", label, mm.Status, mm.EmpiricalP, mm.MultipleTestingAdjustment)
		}
		var regimeNullDesc string
		regimeRate := m.EF5.SameRegimeRate
		if regimeRate != nil && m.EF5.RegimeNull != nil {
			regimeNullDesc = fmt.Sprintf("; EF5 same-regime rate=%.6g (C-REGIME p=%.4g)", *regimeRate, m.EF5.RegimeNull.PValue)
		}
		out = append(out, CrossScaleVerdict{
			ID: "FAMILY_FOLIO_REGIME_DEPENDENCE", Value: value,
			PrimaryStatistic: describe(currier, "cs4/family-currier") + "; " + describe(section, "cs4/family-section") + regimeNullDesc,
			NullComparison:   "folio-level Currier/Section-label permutation (see cs4/family-currier and cs4/family-section null_model)",
			HeldOutResult:    "n/a", PartitionStability: "see EDIT_FAMILIES_STRUCTURALLY_STABLE",
			Sensitivity: "combines two sub-tests (Currier language, Section/$I code); either alone significant is reported as at least PARTIALLY_SUPPORTED",
			Limitations: "Restricted to family-bearing occurrences; Currier B and rarer sections have less power (task65 precedent).",
		})
	}

	// 10. CROSS_SCALE_EFFECTS_SURVIVE_CONDITIONING (from CS8's per-stratum
	// partition_stability attached to cs1)
	{
		cs1 := findCSMetric(cs.Metrics, "cs1/family-line-position")
		value := "INCONCLUSIVE"
		note := "cs1/family-line-position had no partition_stability rows to condition on"
		if cs1 != nil && len(cs1.PartitionStability) > 0 {
			survive, testable := 0, 0
			for _, run := range cs1.PartitionStability {
				if run.Status == "INSUFFICIENT_DATA" {
					continue
				}
				testable++
				if run.Status == "PARTITION_SPECIFIC" {
					survive++
				}
			}
			note = fmt.Sprintf("%d/%d strata (Currier A, Currier B, TEXT-locus) retained a significant within-stratum effect", survive, testable)
			switch {
			case testable == 0:
				value = "INCONCLUSIVE"
			case survive == testable:
				value = "SUPPORTED"
			case survive > 0:
				value = "PARTIALLY_SUPPORTED"
			default:
				value = "NOT_SUPPORTED"
			}
		}
		out = append(out, CrossScaleVerdict{
			ID: "CROSS_SCALE_EFFECTS_SURVIVE_CONDITIONING", Value: value,
			PrimaryStatistic: "CS1 family/line-position effect re-tested within Currier A, Currier B and TEXT-locus strata (each against its own within-line null)",
			NullComparison:   note, HeldOutResult: "n/a", PartitionStability: note,
			Sensitivity: "see TestCS1ConfoundedByRegime for the synthetic case this verdict is designed to catch",
			Limitations: "Only conditions on Currier/locus-type, the two axes task77 names most directly for this check; does not condition on every possible combination of scale variables jointly.",
		})
	}

	// 11. CROSS_SCALE_EFFECTS_GENERALIZE (from CS1's held-out performance)
	{
		cs1 := findCSMetric(cs.Metrics, "cs1/family-line-position")
		value := "INCONCLUSIVE"
		note := "no held-out result attached"
		if cs1 != nil && cs1.HeldOutPerformance != nil {
			hp := cs1.HeldOutPerformance
			if hp.Folds >= 2 {
				note = fmt.Sprintf("%d-fold grouped-by-folio improvement=%.4f (SD %.4f)", hp.Folds, hp.Improvement, hp.ImprovementSD)
				switch {
				case hp.Improvement > hp.ImprovementSD && hp.Improvement > 0:
					value = "SUPPORTED"
				case hp.Improvement > 0:
					value = "PARTIALLY_SUPPORTED"
				default:
					value = "NOT_SUPPORTED"
				}
			} else {
				note = "held-out validation could not run: " + hp.Note
			}
		}
		out = append(out, CrossScaleVerdict{
			ID: "CROSS_SCALE_EFFECTS_GENERALIZE", Value: value,
			PrimaryStatistic: "out-of-fold log-loss improvement of family-label features over a folio-marginal baseline for predicting line-position class",
			NullComparison:   "M0 (folio-marginal position-class distribution) vs M1 (M0 + family label)", HeldOutResult: note,
			PartitionStability: "grouped by folio, so no occurrence's fold membership leaks information about a neighboring occurrence in the same folio",
			Sensitivity:        "sensitive to fold count (cross_scale.folds) and to how many folios exist in the corpus",
			Limitations:        "Only CS1 (the headline family x line-position claim) is held-out validated in this block; CS2-CS7 rely on their permutation nulls alone (task77 §8's held-out schemes are applied to the metric task77 §8 names explicitly, 'предсказывает ли edit family line position').",
		})
	}

	// 12. EDIT_CROSS_SCALE_BLOCK_READY
	{
		computed, total := 0, 0
		for _, mm := range cs.Metrics {
			total++
			if mm.Status != "NOT_APPLICABLE" {
				computed++
			}
		}
		// Never reported SUPPORTED outright: this is infrastructural
		// readiness, not a content conclusion (see Limitations below).
		value := "PARTIALLY_SUPPORTED"
		if computed == 0 {
			value = "INCONCLUSIVE"
		}
		out = append(out, CrossScaleVerdict{
			ID: "EDIT_CROSS_SCALE_BLOCK_READY", Value: value,
			PrimaryStatistic: fmt.Sprintf("%d/%d cross-scale metrics computed (not NOT_APPLICABLE); consensus_status=%s", computed, total, eg.ConsensusStatus),
			NullComparison:   "inherited from each contributing metric", HeldOutResult: "see CROSS_SCALE_EFFECTS_GENERALIZE",
			PartitionStability: "see EDIT_FAMILIES_STRUCTURALLY_STABLE", Sensitivity: "see redundancy_matrix/metric_classifications for which metrics carry independent information",
			Limitations: "This is an infrastructural readiness label (task75's own LEXICAL_PARADIGM_BLOCK_READY precedent), not a content conclusion: it reports that the block ran end-to-end with real data and declared nulls, not that every dependency it found should be frozen into Fingerprint v2 as-is. See TASK77_REPORT.md's recommendations for what is and is not ready to freeze.",
		})
	}

	return out
}
