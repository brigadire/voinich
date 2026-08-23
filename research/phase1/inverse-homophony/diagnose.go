package main

import (
	"fmt"
	"os"

	"zcore.dev/voinich/internal/inversehomophony"
)

func runDiagnose(args []string) int {
	fs := newFlagSet("diagnose")
	corpus := fs.String("corpus", "", "ciphertext corpus path (required)")
	mapping := fs.String("mapping", "", "corpus-transform mapping.tsv path (evaluator-only, required)")
	seed := fs.Int64("seed", 1, "deterministic seed for the matched false-pair sample")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *corpus == "" || *mapping == "" {
		fmt.Fprintln(os.Stderr, "diagnose: -corpus and -mapping are required")
		return 2
	}

	loaded, err := inversehomophony.LoadCorpus(*corpus)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load corpus:", err)
		return 1
	}
	oracleMapping, err := inversehomophony.LoadOracleMapping(*mapping)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load mapping:", err)
		return 1
	}
	oracle := oracleMapping.OraclePartitionForRelabel(loaded.Relabel)

	cfg := inversehomophony.FrozenConfig()
	features := inversehomophony.BuildFeatures(loaded.Relabel.Tokens, loaded.LineOfToken, cfg)
	d := inversehomophony.DiscriminatePairs(features, oracle, cfg, *seed)

	fmt.Printf("corpus: %s\n", *corpus)
	fmt.Printf("cipher types: %d\n", len(features))
	fmt.Printf("true pairs scored: %d\n", d.TruePairs)
	fmt.Printf("false pairs scored: %d\n", d.FalsePairs)
	fmt.Printf("AUC: %.4f\n", d.AUC)
	tau := inversehomophony.FreezeThreshold(d)
	fmt.Printf("Youden-J-optimal tau (development-fit): %.4f\n", tau)
	return 0
}
