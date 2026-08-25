package main

// ThresholdIndex is a fast lookup of materialized calibration thresholds.
type ThresholdIndex struct {
	m map[string]CalibThreshold
}

func NewThresholdIndex(all []CalibThreshold) *ThresholdIndex {
	idx := &ThresholdIndex{m: map[string]CalibThreshold{}}
	for _, t := range all {
		idx.m[calibKey(t.Quantity, t.Metric, t.ModelClass, t.CandidateID)] = t
	}
	return idx
}

func (t *ThresholdIndex) Get(quantity, metric, class, candID string) (CalibThreshold, bool) {
	v, ok := t.m[calibKey(quantity, metric, class, candID)]
	return v, ok
}
