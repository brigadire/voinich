package main

import (
	"fmt"
	"strings"

	"zcore.dev/voinich/internal/mechanismspace"
)

// stateNullTargets are the representative stateful mechanisms task66
// sections 55-57's null controls are run against: one plain stateful
// mechanism (M4), one slow-drift mechanism (M5) and one mixed mechanism
// (M7), each a real grid entry so the null and its real counterpart are
// directly comparable.
func stateNullTargets() []GridEntry {
	return []GridEntry{
		{"M4_STATE_K4_A", mechanismspace.Config{Family: "M4", StateCount: 4, Update: mechanismspace.UpdateA}},
		{"M5_DRIFT_N20", mechanismspace.Config{Family: "M5", StateCount: 1000, DriftScale: 20}},
		{"M7_MIXED_K5_N20", mechanismspace.Config{Family: "M7", MacroStates: 5, StateCount: 1000, DriftScale: 20}},
	}
}

// RunStateNulls runs each state-null target's real config plus its
// shuffled-state (55), fixed-state (57) and, for the drift mechanisms,
// fast-state (56) variants, and writes STATE_NULLS.tsv.
func RunStateNulls(corpora map[string]mechanismspace.Corpus, replicates int, opt mechanismspace.FingerprintOptions, path string) {
	type variant struct {
		suffix string
		apply  func(mechanismspace.Config) mechanismspace.Config
	}
	variants := []variant{
		{"REAL", func(c mechanismspace.Config) mechanismspace.Config { return c }},
		{"SHUFFLED_STATE", func(c mechanismspace.Config) mechanismspace.Config { c.ShuffleStateNull = true; return c }},
		{"FIXED_STATE", func(c mechanismspace.Config) mechanismspace.Config { c.FixedStateNull = true; return c }},
		{"FAST_STATE", func(c mechanismspace.Config) mechanismspace.Config { c.FastStateNull = true; return c }},
	}
	var b strings.Builder
	b.WriteString("mechanism\tnull\tcorpus\tcorrelation_length_tokens\tcluster_stability\tweighted_entropy\th2\n")
	for _, target := range stateNullTargets() {
		for _, v := range variants {
			cfg := v.apply(target.Config)
			entry := GridEntry{Name: target.Name + "_" + v.suffix, Config: cfg}
			results := RunGrid([]GridEntry{entry}, corpora, replicates, opt, 600000)
			for _, k := range sortedKeys(GroupByMechanismCorpus(results)) {
				rs := GroupByMechanismCorpus(results)[k]
				mean := MeanFingerprint(fingerprintsOf(rs))
				parts := strings.SplitN(k, "|", 2)
				b.WriteString(fmt.Sprintf("%s\t%s\t%s\t%.9g\t%.9g\t%.9g\t%.9g\n", target.Name, v.suffix, parts[1], mean.CorrelationLengthTokens, mean.ClusterStability, mean.PositionalWeightedEntropy, mean.H2))
			}
		}
	}
	writeFile(path, b.String())
}

// boundaryNullTarget is the STREAM-mode generated-boundary mechanism
// task66 section 58's null is compared against.
func boundaryNullTarget() GridEntry {
	return GridEntry{"M8_BOUNDARY_STATE", mechanismspace.Config{Family: "M8", InputMode: mechanismspace.Stream, Grouping: mechanismspace.StateGrouping, GroupLen: 4, StateCount: 4}}
}

// RunBoundaryNulls runs the real state-dependent boundary rule and its
// randomized-boundary null (same output-length distribution, task66
// section 58) and writes BOUNDARY_NULLS.tsv.
func RunBoundaryNulls(corpora map[string]mechanismspace.Corpus, replicates int, opt mechanismspace.FingerprintOptions, path string) {
	target := boundaryNullTarget()
	real := GridEntry{target.Name + "_REAL", target.Config}
	nullCfg := target.Config
	nullCfg.RandomBoundaryNull = true
	null := GridEntry{target.Name + "_RANDOM_BOUNDARY_NULL", nullCfg}
	results := RunGrid([]GridEntry{real, null}, corpora, replicates, opt, 700000)
	grouped := GroupByMechanismCorpus(results)
	var b strings.Builder
	b.WriteString("mechanism\tcorpus\treplicates\tweighted_entropy\tgiant_component_fraction\th2\ttoken_order_bits\n")
	for _, k := range sortedKeys(grouped) {
		rs := grouped[k]
		mean := MeanFingerprint(fingerprintsOf(rs))
		parts := strings.SplitN(k, "|", 2)
		b.WriteString(fmt.Sprintf("%s\t%s\t%d\t%.9g\t%.9g\t%.9g\t%.9g\n", parts[0], parts[1], len(rs), mean.PositionalWeightedEntropy, mean.GiantComponentFraction, mean.H2, mean.TokenOrderBits))
	}
	writeFile(path, b.String())
}
