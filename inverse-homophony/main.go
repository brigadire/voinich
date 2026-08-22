// Command inverse-homophony implements task57's blind ciphertext-only
// homophone-class recovery method and its synthetic validation harness.
//
// "diagnose" is the section-19 early-stop diagnostic (requires an
// evaluator-only oracle mapping.tsv - never fed to the recovery path
// itself). "validate" runs the full development/validation protocol over
// the fixed synthetic corpus set and writes every required artifact under
// an output directory. "recover" runs the frozen method on ciphertext
// alone (no oracle) - the same code path "validate" uses internally, and
// the only mode Phase B's Voynich run is allowed to use. "voynich" runs
// Phase B and refuses to run without a METHOD_FROZEN marker.
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}
	switch args[0] {
	case "diagnose":
		return runDiagnose(args[1:])
	case "recover":
		return runRecover(args[1:])
	case "validate":
		return runValidate(args[1:])
	case "voynich":
		return runVoynich(args[1:])
	case "-h", "-help", "--help", "help":
		usage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n\n", args[0])
		usage()
		return 2
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `inverse-homophony <subcommand> [flags]

Subcommands:
  diagnose  -corpus PATH -mapping PATH        section-19 true/false pair AUC (evaluator-only)
  recover   -corpus PATH -output PATH         blind recovery (ciphertext only)
  validate  -out-dir DIR                       full synthetic validation protocol
  voynich   -corpus PATH -out-dir DIR          Phase B (requires METHOD_FROZEN)
`)
}

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	return fs
}
