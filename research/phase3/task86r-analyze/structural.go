package main

var editFamilyMetrics = []string{"EF1_GIANT_COMPONENT_SHARE", "EF1_ISOLATE_SHARE", "EF2_GLOBAL_CLUSTERING", "EF3_DEGREE_FREQUENCY_SPEARMAN"}
var lexicalFamilyMetrics = []string{"LP1_RULE_SUPPORT_GINI", "LP4_PREFIX_ATTACHMENT_NMI", "LP4_SUFFIX_ATTACHMENT_NMI"}

// StructuralMetricPass evaluates one F2 metric's per-scale gate:
// abs(median_generated-heldout) <= max applicable MFC q0.95 structural
// distance threshold (equality passes); an unavailable/degenerate metric
// fails (G1_ADEQUACY_GATES.md).
func StructuralMetricPass(class, candID, metric string, medianGenerated float64, generatedValid bool, heldoutValue float64, heldoutValid bool, idx *ThresholdIndex) (pass bool, distance, threshold float64) {
	if !generatedValid || !heldoutValid || !isFinite(medianGenerated) || !isFinite(heldoutValue) {
		return false, negInf(), 0
	}
	distance = absf(medianGenerated - heldoutValue)
	t, ok := idx.Get("structural_distance", metric, class, candID)
	if !ok || !isFinite(t.Threshold) {
		return false, distance, 0
	}
	threshold = t.Threshold
	return distance <= threshold, distance, threshold
}

func absf(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// FamilyPass counts how many family members pass and applies the frozen
// "at least K of N" rule.
func FamilyPass(members []string, passing map[string]bool, minPass int) (pass bool, passCount int) {
	for _, m := range members {
		if passing[m] {
			passCount++
		}
	}
	return passCount >= minPass, passCount
}
