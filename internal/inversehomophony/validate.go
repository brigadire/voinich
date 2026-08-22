package inversehomophony

import (
	"fmt"

	"zcore.dev/voinich/internal/genericsegmentation"
)

// RandomSeeds is the fixed set of seeds used for every RANDOM_PARTITION
// draw in the validation harness (task57 section 33's null_distribution.tsv).
// Fixed here, not derived from any corpus.
var RandomSeeds = []int64{101, 102, 103, 104, 105}

// PairDiscriminationRow is one row of pair_discrimination.tsv.
type PairDiscriminationRow struct {
	Split       string
	Label       string
	CipherTypes int
	TruePairs   int
	FalsePairs  int
	AUC         float64
}

// ClassRecoveryRow is one row of class_recovery.tsv.
type ClassRecoveryRow struct {
	Split   string
	Label   string
	Method  string
	Metrics ClassRecoveryMetrics
}

// StructuralRecoveryRow is one row of structural_recovery.tsv.
type StructuralRecoveryRow struct {
	Split  string
	Label  string
	Method string
	StructuralComparison
}

// NullDistributionRow is one row of null_distribution.tsv: one
// RANDOM_PARTITION seed's outcome for one metric.
type NullDistributionRow struct {
	Split  string
	Label  string
	Seed   int64
	Metric string
	Value  float64
}

// corpusEval is everything computed for one synthetic corpus.
type corpusEval struct {
	spec           SyntheticCorpusSpec
	split          string
	cipherTypes    int
	pairDisc       PairDiscriminationRow
	classRows      []ClassRecoveryRow
	structRows     []StructuralRecoveryRow
	nullRows       []NullDistributionRow
	mergeEvents    []MergeEvent
	recoveredSizes []int
}

// evaluateCorpus runs the full frozen method plus every baseline over one
// synthetic corpus and returns every row needed for the validation
// artifacts. cfg.Threshold must already be frozen (see
// RunSyntheticValidation).
func evaluateCorpus(spec SyntheticCorpusSpec, split string, cfg Config) (*corpusEval, error) {
	loaded, err := LoadCorpus(spec.CipherPath)
	if err != nil {
		return nil, fmt.Errorf("load ciphertext %s: %w", spec.CipherPath, err)
	}
	oracleMapping, err := LoadOracleMapping(spec.MappingPath)
	if err != nil {
		return nil, fmt.Errorf("load oracle mapping %s: %w", spec.MappingPath, err)
	}
	plaintextPath, err := PlaintextPathFromManifest(spec.CipherPath)
	if err != nil {
		return nil, fmt.Errorf("locate plaintext for %s: %w", spec.CipherPath, err)
	}
	plainTokens, plainLineOf, _, err := genericsegmentation.ReadCorpus(plaintextPath)
	if err != nil {
		return nil, fmt.Errorf("load plaintext %s: %w", plaintextPath, err)
	}
	plainLines := groupLines(plainTokens, plainLineOf)

	oracle := oracleMapping.OraclePartitionForRelabel(loaded.Relabel)
	freq := make(map[string]int, len(loaded.Relabel.ToOpaque))
	for _, t := range loaded.Relabel.Tokens {
		freq[t]++
	}

	features := BuildFeatures(loaded.Relabel.Tokens, loaded.LineOfToken, cfg)
	disc := DiscriminatePairs(features, oracle, cfg, cfg.Seed)
	pairs := CandidatePairs(features, cfg)

	recovered, mergeEvents := Recover(freq, pairs, cfg)
	noCollapse := NoCollapsePartition(freq)
	freqOnly := FrequencyOnlyPartition(freq)
	oracleCollapse := oracle // oracle IS a partition already, in relabeled space

	recoveredSizes := PartitionClassSizes(recovered, freq)

	ev := &corpusEval{spec: spec, split: split, cipherTypes: len(features), mergeEvents: mergeEvents, recoveredSizes: recoveredSizes}
	ev.pairDisc = PairDiscriminationRow{Split: split, Label: spec.Label, CipherTypes: len(features), TruePairs: disc.TruePairs, FalsePairs: disc.FalsePairs, AUC: disc.AUC}

	addClassRow := func(method string, p Partition) {
		ev.classRows = append(ev.classRows, ClassRecoveryRow{Split: split, Label: spec.Label, Method: method, Metrics: EvaluateClassRecovery(p, oracle)})
	}
	addClassRow("recovered", recovered)
	addClassRow("no_collapse", noCollapse)
	addClassRow("frequency_only", freqOnly)
	addClassRow("oracle", oracleCollapse)

	plainStruct := ComputeStructural(plainTokens, plainLines)
	cipherStruct := ComputeStructural(loaded.Relabel.Tokens, loaded.Lines)

	addStructRows := func(method string, p Partition) {
		collapsedTokens := Collapse(loaded.Relabel.Tokens, p)
		collapsedLines := CollapseLines(loaded.Lines, p)
		recStruct := ComputeStructural(collapsedTokens, collapsedLines)
		for _, c := range CompareStructural(plainStruct, cipherStruct, recStruct) {
			ev.structRows = append(ev.structRows, StructuralRecoveryRow{Split: split, Label: spec.Label, Method: method, StructuralComparison: c})
		}
	}
	addStructRows("recovered", recovered)
	addStructRows("no_collapse", noCollapse)
	addStructRows("frequency_only", freqOnly)
	addStructRows("oracle", oracleCollapse)

	for _, seed := range RandomSeeds {
		rp := RandomPartition(freq, recoveredSizes, seed)
		addClassRow(fmt.Sprintf("random_seed%d", seed), rp)
		collapsedTokens := Collapse(loaded.Relabel.Tokens, rp)
		collapsedLines := CollapseLines(loaded.Lines, rp)
		recStruct := ComputeStructural(collapsedTokens, collapsedLines)
		cmp := CompareStructural(plainStruct, cipherStruct, recStruct)
		crm := EvaluateClassRecovery(rp, oracle)
		ev.nullRows = append(ev.nullRows, NullDistributionRow{Split: split, Label: spec.Label, Seed: seed, Metric: "pairwise_f1", Value: crm.PairwiseF1})
		ev.nullRows = append(ev.nullRows, NullDistributionRow{Split: split, Label: spec.Label, Seed: seed, Metric: "ari", Value: crm.ARI})
		for _, c := range cmp {
			ev.structRows = append(ev.structRows, StructuralRecoveryRow{Split: split, Label: spec.Label, Method: fmt.Sprintf("random_seed%d", seed), StructuralComparison: c})
			ev.nullRows = append(ev.nullRows, NullDistributionRow{Split: split, Label: spec.Label, Seed: seed, Metric: c.Metric, Value: c.Recovered})
		}
	}

	return ev, nil
}
