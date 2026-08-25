package main

import "math"

// MetricGateResult is one predictive metric's per-transcription gate
// evaluation.
type MetricGateResult struct {
	Metric        string
	ImprovementB1 float64
	ThresholdB1   float64
	PassB1        bool
	B2Applicable  bool
	ImprovementB2 float64
	ThresholdB2   float64
	PassB2        bool
	Pass          bool
}

var higherBetter = map[string]bool{"PM4": true, "PM6": true}

func metricImprovement(metric string, cand, base float64) float64 {
	if higherBetter[metric] {
		return cand - base
	}
	return base - cand
}

// EvalPredictiveMetric evaluates one metric's gate on one transcription.
func EvalPredictiveMetric(metric, class, candID string, candValue, b1Value float64, b2Applicable bool, b2Value float64, idx *ThresholdIndex) MetricGateResult {
	r := MetricGateResult{Metric: metric, B2Applicable: b2Applicable}
	if t, ok := idx.Get("predictive_gain_vs_b1", metric, class, candID); ok {
		r.ThresholdB1 = t.Threshold
		r.ImprovementB1 = metricImprovement(metric, candValue, b1Value)
		r.PassB1 = isFinite(r.ImprovementB1) && isFinite(r.ThresholdB1) && r.ImprovementB1 > r.ThresholdB1
	}
	if b2Applicable {
		if t, ok := idx.Get("predictive_gain_vs_b2", metric, class, candID); ok {
			r.ThresholdB2 = t.Threshold
			r.ImprovementB2 = metricImprovement(metric, candValue, b2Value)
			r.PassB2 = isFinite(r.ImprovementB2) && isFinite(r.ThresholdB2) && r.ImprovementB2 > r.ThresholdB2
		}
		r.Pass = r.PassB1 && r.PassB2
	} else {
		r.Pass = r.PassB1
	}
	return r
}

func isFinite(f float64) bool { return !math.IsNaN(f) && !math.IsInf(f, 0) }
