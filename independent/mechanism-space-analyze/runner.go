package main

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"zcore.dev/voinich/internal/mechanismspace"
)

// ReplicateResult is one (mechanism, corpus, seed) job's fingerprint,
// kept individually (not just averaged) so CI/seed-sensitivity can be
// reported (task66 section 40).
type ReplicateResult struct {
	Mechanism, Corpus string
	Seed              int64
	Fingerprint       mechanismspace.Fingerprint
	Output            mechanismspace.Output
}

// RunGrid executes every (grid entry x corpus x seed) job through the
// local worker pool (task66's immutable-job worker contract: Execute has
// no knowledge of any target or of other jobs) and returns one
// ReplicateResult per job.
func RunGrid(grid []GridEntry, corpora map[string]mechanismspace.Corpus, replicates int, opt mechanismspace.FingerprintOptions, baseSeed int64) []ReplicateResult {
	var jobs []Job2
	for _, e := range grid {
		for name := range corpora {
			for rep := 0; rep < replicates; rep++ {
				jobs = append(jobs, Job2{Mechanism: e, Corpus: name, Seed: baseSeed + int64(rep)})
			}
		}
	}
	return executeJobs2(jobs, corpora, opt)
}

// Job2 pairs a named grid entry with a corpus/seed - a thin wrapper over
// mechanismspace.Job that keeps the human-readable mechanism name
// alongside the immutable job for reporting.
type Job2 struct {
	Mechanism GridEntry
	Corpus    string
	Seed      int64
}

func executeJobs2(jobs []Job2, corpora map[string]mechanismspace.Corpus, opt mechanismspace.FingerprintOptions) []ReplicateResult {
	msJobs := make([]mechanismspace.Job, len(jobs))
	for i, j := range jobs {
		msJobs[i] = mechanismspace.Job{ExperimentID: "task66", Corpus: j.Corpus, Mechanism: j.Mechanism.Config, Seed: j.Seed, EvaluationSet: "GRID"}
	}
	results := mechanismspace.RunLocal(msJobs, corpora, opt, 11)
	out := make([]ReplicateResult, len(jobs))
	for i, r := range results {
		out[i] = ReplicateResult{Mechanism: jobs[i].Mechanism.Name, Corpus: jobs[i].Corpus, Seed: jobs[i].Seed, Fingerprint: r.Fingerprint, Output: r.Output}
	}
	return out
}

// GroupByMechanismCorpus buckets replicate results by (mechanism, corpus).
func GroupByMechanismCorpus(results []ReplicateResult) map[string][]ReplicateResult {
	out := map[string][]ReplicateResult{}
	for _, r := range results {
		key := r.Mechanism + "|" + r.Corpus
		out[key] = append(out[key], r)
	}
	return out
}

// MeanFingerprint averages every numeric field across replicates
// (task66 section 40's mean/CI reporting) and keeps the majority
// Topology label.
func MeanFingerprint(fps []mechanismspace.Fingerprint) mechanismspace.Fingerprint {
	n := float64(len(fps))
	if n == 0 {
		return mechanismspace.Fingerprint{}
	}
	var out mechanismspace.Fingerprint
	topo := map[string]int{}
	for _, f := range fps {
		out.TokenOrderBits += f.TokenOrderBits / n
		out.EdgeOrderBits += f.EdgeOrderBits / n
		out.PositionalWeightedEntropy += f.PositionalWeightedEntropy / n
		out.HighFreqSpecialists += f.HighFreqSpecialists / n
		out.ExactAdjacentRepeatRate += f.ExactAdjacentRepeatRate / n
		out.GiantComponentFraction += f.GiantComponentFraction / n
		if !math.IsNaN(f.H1) {
			out.H1 += f.H1 / n
		}
		if !math.IsNaN(f.H2) {
			out.H2 += f.H2 / n
		}
		if !math.IsNaN(f.H3) {
			out.H3 += f.H3 / n
		}
		if !math.IsNaN(f.H4) {
			out.H4 += f.H4 / n
		}
		out.PositionGainBits += f.PositionGainBits / n
		out.OrderGainBits += f.OrderGainBits / n
		out.AdjacentNearRate += f.AdjacentNearRate / n
		out.ResidualAdjacency += f.ResidualAdjacency / n
		out.CorrelationLengthTokens += f.CorrelationLengthTokens / n
		out.ClusterStability += f.ClusterStability / n
		out.WithinClusterDrift += f.WithinClusterDrift / n
		out.LineBoundaryDelta += f.LineBoundaryDelta / n
		topo[f.Topology]++
	}
	best, bestN := "", 0
	for k, v := range topo {
		if v > bestN || (v == bestN && k < best) {
			best, bestN = k, v
		}
	}
	out.Topology = best
	out.Status = fps[0].Status
	return out
}

// SeedCI returns (mean, lo95, hi95) for one representative metric across
// replicates (task66 section 40), using a simple normal approximation
// (n is always small; this is a descriptive CI, not a hypothesis test).
func SeedCI(values []float64) (mean, lo, hi float64) {
	var clean []float64
	for _, v := range values {
		if !math.IsNaN(v) {
			clean = append(clean, v)
		}
	}
	n := len(clean)
	if n == 0 {
		return 0, 0, 0
	}
	for _, v := range clean {
		mean += v
	}
	mean /= float64(n)
	if n < 2 {
		return mean, mean, mean
	}
	sd := 0.0
	for _, v := range clean {
		sd += (v - mean) * (v - mean)
	}
	sd = math.Sqrt(sd / float64(n-1))
	se := sd / math.Sqrt(float64(n))
	return mean, mean - 1.96*se, mean + 1.96*se
}

// WriteResultsTSV writes one SCREENING_RESULTS.tsv/DEVELOPMENT_RESULTS.tsv
// row per (mechanism, corpus): mean fingerprint fields plus the seed CI of
// the token-order metric as a representative stochasticity indicator.
func WriteResultsTSV(path string, grouped map[string][]ReplicateResult) error {
	var b strings.Builder
	b.WriteString("mechanism\tcorpus\treplicates\ttoken_order_bits\ttoken_order_ci_lo\ttoken_order_ci_hi\tedge_order_bits\tweighted_entropy\tgiant_component_fraction\texact_adjacent_repeat_rate\th1\th2\th3\th4\tposition_gain_bits\torder_gain_bits\tadjacent_near_rate\tresidual_adjacency\tcorrelation_length_tokens\tcluster_stability\twithin_cluster_drift\tline_boundary_delta\ttopology\n")
	keys := sortedKeys(grouped)
	for _, k := range keys {
		rs := grouped[k]
		parts := strings.SplitN(k, "|", 2)
		mean := MeanFingerprint(fingerprintsOf(rs))
		var tob []float64
		for _, r := range rs {
			tob = append(tob, r.Fingerprint.TokenOrderBits)
		}
		_, lo, hi := SeedCI(tob)
		b.WriteString(fmt.Sprintf("%s\t%s\t%d\t%.9g\t%.9g\t%.9g\t%.9g\t%.9g\t%.9g\t%.9g\t%.9g\t%.9g\t%.9g\t%.9g\t%.9g\t%.9g\t%.9g\t%.9g\t%.9g\t%.9g\t%.9g\t%.9g\t%s\n",
			parts[0], parts[1], len(rs), mean.TokenOrderBits, lo, hi, mean.EdgeOrderBits, mean.PositionalWeightedEntropy,
			mean.GiantComponentFraction, mean.ExactAdjacentRepeatRate, mean.H1, mean.H2, mean.H3, mean.H4,
			mean.PositionGainBits, mean.OrderGainBits, mean.AdjacentNearRate, mean.ResidualAdjacency,
			mean.CorrelationLengthTokens, mean.ClusterStability, mean.WithinClusterDrift, mean.LineBoundaryDelta, mean.Topology))
	}
	return writeFile(path, b.String())
}

func fingerprintsOf(rs []ReplicateResult) []mechanismspace.Fingerprint {
	out := make([]mechanismspace.Fingerprint, len(rs))
	for i, r := range rs {
		out[i] = r.Fingerprint
	}
	return out
}

func sortedKeys(m map[string][]ReplicateResult) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// FamilyMetricsRow is one FAMILY_METRICS.tsv row (task66 section 37):
// robust (median) per-family normalized progress for one mechanism on one
// corpus.
type FamilyMetricsRow struct {
	Mechanism, Corpus, EvaluationSet string
	FamilyScores                     map[string]float64
	OverallStatus                    string
}

// ComputeFamilyMetrics computes every mechanism's per-corpus family score
// against that corpus's own M0 baseline and the shared Voynich targets,
// using only the given stage's metrics (task66 sections 5, 35-37): pass
// StageDevelopment for every selection/gating use, and StageHeldout only
// after candidates are frozen (section 41).
func ComputeFamilyMetrics(grouped map[string][]ReplicateResult, baselines map[string]mechanismspace.Fingerprint, targets []Target, evalSet, stage string) []FamilyMetricsRow {
	var out []FamilyMetricsRow
	for _, k := range sortedKeys(grouped) {
		rs := grouped[k]
		parts := strings.SplitN(k, "|", 2)
		mech, corpus := parts[0], parts[1]
		mts := BuildMetricTargets(targets, baselines[corpus], stage)
		mean := MeanFingerprint(fingerprintsOf(rs))
		progresses := ProgressFor(mts, mean)
		scores := mechanismspace.FamilyScore(progresses)
		out = append(out, FamilyMetricsRow{Mechanism: mech, Corpus: corpus, EvaluationSet: evalSet, FamilyScores: scores, OverallStatus: ClassifyMechanismResult(scores)})
	}
	return out
}

// ClassifyMechanismResult is task66 section 69's per-(mechanism,corpus)
// classification, based on how many independent families show real
// (non-overshoot, non-trivial) progress toward Voynich.
func ClassifyMechanismResult(scores map[string]float64) string {
	moved, overshoot := 0, 0
	for _, v := range scores {
		if v > 1.15 {
			overshoot++
		} else if v > 0.15 {
			moved++
		}
	}
	switch {
	case overshoot >= 2 && moved+overshoot >= len(scores)-1:
		return "PATHOLOGICAL"
	case moved >= 4:
		return "BROAD_COMPATIBILITY"
	case moved >= 2:
		return "PARTIAL_MULTI_FAMILY"
	case moved == 1:
		return "SINGLE_FAMILY_EFFECT"
	default:
		return "NO_EFFECT"
	}
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
