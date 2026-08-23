package main

import (
	"fmt"
	"strings"
)

// WriteTopologyResultsTSV extracts task66 section 52's topology
// classification (plus its supporting fields) from the already-computed
// development-grade results, and appends the authoritative Voynich
// target row for direct visual comparison (task66 section 51).
func WriteTopologyResultsTSV(path string, grouped map[string][]ReplicateResult, targets []Target) {
	var b strings.Builder
	b.WriteString("mechanism\tcorpus\tcorrelation_length_tokens\tcluster_stability\twithin_cluster_drift\tline_boundary_delta\ttopology\n")
	for _, k := range sortedKeys(grouped) {
		rs := grouped[k]
		mean := MeanFingerprint(fingerprintsOf(rs))
		parts := strings.SplitN(k, "|", 2)
		b.WriteString(fmt.Sprintf("%s\t%s\t%.9g\t%.9g\t%.9g\t%.9g\t%s\n", parts[0], parts[1], mean.CorrelationLengthTokens, mean.ClusterStability, mean.WithinClusterDrift, mean.LineBoundaryDelta, mean.Topology))
	}
	var corrLen, lineDelta float64
	for _, t := range targets {
		if t.Metric == "correlation_length_tokens" && t.Status == "VALUE" {
			corrLen = t.Voynich
		}
		if t.Metric == "line_boundary_delta" && t.Status == "VALUE" {
			lineDelta = t.Voynich
		}
	}
	b.WriteString(fmt.Sprintf("VOYNICH_TARGET\tVoynich\t%.9g\t\t\t%.9g\tMIXED_DRIFT_AND_STATES\n", corrLen, lineDelta))
	writeFile(path, b.String())
}
