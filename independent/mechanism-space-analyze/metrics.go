package main

import "zcore.dev/voinich/internal/mechanismspace"

// MetricSpec binds one task66 section 9 fingerprint field to its
// VOYNICH_TARGET_MANIFEST metric key, so progress/Pareto computation
// (sections 35-38) can walk every metric generically. Stage marks
// whether the metric is on the DEVELOPMENT side (visible during
// screening/development/Pareto selection) or held back per section 5
// until candidates are frozen.
type MetricSpec struct {
	Family, Metric, Stage string
	Extract               func(mechanismspace.Fingerprint) float64
}

const (
	StageDevelopment = "DEVELOPMENT"
	StageHeldout     = "HELDOUT"
)

// MetricRegistry is task66 section 5's frozen development/held-out split,
// fixed before any mechanism search: within every family, one primary
// metric is DEVELOPMENT (used for screening, Pareto selection and gating)
// and at least one companion metric is HELDOUT (opened only once, after
// Pareto candidates are frozen - task66 sections 39-42).
var MetricRegistry = []MetricSpec{
	{"TOKEN_ORDER", "token_order_bits", StageDevelopment, func(f mechanismspace.Fingerprint) float64 { return f.TokenOrderBits }},
	{"TOKEN_ORDER", "edge_order_bits", StageHeldout, func(f mechanismspace.Fingerprint) float64 { return f.EdgeOrderBits }},
	{"POSITIONAL_STRUCTURE", "weighted_entropy", StageDevelopment, func(f mechanismspace.Fingerprint) float64 { return f.PositionalWeightedEntropy }},
	{"POSITIONAL_STRUCTURE", "high_freq_specialists", StageHeldout, func(f mechanismspace.Fingerprint) float64 { return f.HighFreqSpecialists }},
	{"REPETITION_EDIT_GEOMETRY", "giant_component_fraction", StageDevelopment, func(f mechanismspace.Fingerprint) float64 { return f.GiantComponentFraction }},
	{"REPETITION_EDIT_GEOMETRY", "exact_adjacent_repeat_rate", StageHeldout, func(f mechanismspace.Fingerprint) float64 { return f.ExactAdjacentRepeatRate }},
	{"CHARACTER_ENTROPY", "h1", StageDevelopment, func(f mechanismspace.Fingerprint) float64 { return f.H1 }},
	{"CHARACTER_ENTROPY", "h2", StageDevelopment, func(f mechanismspace.Fingerprint) float64 { return f.H2 }},
	{"CHARACTER_ENTROPY", "h3", StageHeldout, func(f mechanismspace.Fingerprint) float64 { return f.H3 }},
	{"CHARACTER_ENTROPY", "h4", StageHeldout, func(f mechanismspace.Fingerprint) float64 { return f.H4 }},
	{"TOKEN_FORMATION", "position_gain_bits", StageDevelopment, func(f mechanismspace.Fingerprint) float64 { return f.PositionGainBits }},
	{"TOKEN_FORMATION", "order_gain_bits", StageHeldout, func(f mechanismspace.Fingerprint) float64 { return f.OrderGainBits }},
	{"LOCAL_TRANSITION", "adjacent_near_rate", StageDevelopment, func(f mechanismspace.Fingerprint) float64 { return f.AdjacentNearRate }},
	{"LOCAL_TRANSITION", "residual_adjacency", StageHeldout, func(f mechanismspace.Fingerprint) float64 { return f.ResidualAdjacency }},
	{"LOCAL_REGIME_TOPOLOGY", "correlation_length_tokens", StageDevelopment, func(f mechanismspace.Fingerprint) float64 { return f.CorrelationLengthTokens }},
	{"LOCAL_REGIME_TOPOLOGY", "line_boundary_delta", StageHeldout, func(f mechanismspace.Fingerprint) float64 { return f.LineBoundaryDelta }},
}

// BuildMetricTargets pairs every MetricRegistry entry of the given stage
// (DEVELOPMENT or HELDOUT) that has a real (VALUE-status) Voynich target
// with its corpus-specific M0 baseline, producing the preregistered
// (baseline, target, direction) triples task66 section 35 requires.
// Metrics whose authoritative target is missing are skipped rather than
// defaulted to zero (task66 section 10); HELDOUT metrics are never
// included unless the caller explicitly asks for them (task66 section 5).
func BuildMetricTargets(targets []Target, baseline mechanismspace.Fingerprint, stage string) []mechanismspace.MetricTarget {
	byKey := map[string]Target{}
	for _, t := range targets {
		byKey[t.Family+"|"+t.Metric] = t
	}
	var out []mechanismspace.MetricTarget
	for _, m := range MetricRegistry {
		if m.Stage != stage {
			continue
		}
		t, ok := byKey[m.Family+"|"+m.Metric]
		if !ok || t.Status != "VALUE" {
			continue
		}
		b := m.Extract(baseline)
		dir := mechanismspace.Higher
		if t.Voynich < b {
			dir = mechanismspace.Lower
		}
		out = append(out, mechanismspace.MetricTarget{Family: m.Family, Metric: m.Metric, Baseline: b, Voynich: t.Voynich, Direction: dir})
	}
	return out
}

// ProgressFor computes every metric's Progress (task66 section 35) for
// one candidate fingerprint against the corpus-specific targets.
func ProgressFor(mts []mechanismspace.MetricTarget, candidate mechanismspace.Fingerprint) []mechanismspace.Progress {
	byKey := map[string]func(mechanismspace.Fingerprint) float64{}
	for _, m := range MetricRegistry {
		byKey[m.Family+"|"+m.Metric] = m.Extract
	}
	var out []mechanismspace.Progress
	for _, mt := range mts {
		extract := byKey[mt.Family+"|"+mt.Metric]
		out = append(out, mechanismspace.ComputeProgress(mt, extract(candidate)))
	}
	return out
}
