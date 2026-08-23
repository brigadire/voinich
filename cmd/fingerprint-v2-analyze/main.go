// fingerprint-v2-analyze computes the deterministic LP1-LP4 / EF1-EF4
// Fingerprint v2 lexical-paradigm block from a declarative YAML config.
package main

import (
	"flag"
	"fmt"
	"os"

	"zcore.dev/voinich/internal/fingerprintv2"
	"zcore.dev/voinich/internal/workdir"
)

func main() {
	config := flag.String("config", "", fmt.Sprintf("YAML analysis configuration (required; output_dir is explicit, not implicit %s)", workdir.Dir))
	flag.Parse()
	if *config == "" {
		fmt.Fprintln(os.Stderr, "usage: fingerprint-v2-analyze -config analysis.yaml")
		os.Exit(2)
	}
	fingerprint, err := fingerprintv2.RunFile(*config)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fingerprint-v2-analyze:", err)
		os.Exit(1)
	}
	fmt.Printf("Fingerprint v2 lexical-paradigm analysis completed: %s (%d tokens); outputs: %s\n",
		fingerprint.Primary.Corpus.ID, fingerprint.Primary.Corpus.TokenCount, fingerprint.Primary.Corpus.Preprocessing)
}
