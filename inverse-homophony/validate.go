package main

import (
	"fmt"
	"os"

	"zcore.dev/voinich/internal/inversehomophony"
)

func runValidate(args []string) int {
	fs := newFlagSet("validate")
	outDir := fs.String("out-dir", "experiments/inverse-homophony/synthetic-validation", "output directory (task57 section 33)")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	rc := inversehomophony.ValidationRunConfig{OutDir: *outDir, GitCommit: gitCommit(), GitDirty: gitDirty()}
	report, err := inversehomophony.RunSyntheticValidation(rc)
	if err != nil {
		fmt.Fprintln(os.Stderr, "validate:", err)
		return 1
	}

	fmt.Printf("development AUC: %.4f (tau=%.4f)\n", report.DevelopmentAUC, report.Config.Threshold)
	fmt.Printf("validation gate: %s\n", verdictWord(report.Gate.Pass))
	for _, c := range report.Gate.Criteria {
		fmt.Printf("  [%v] %s: %s\n", c.Pass, c.Name, c.Detail)
	}
	fmt.Printf("artifacts written to %s\n", *outDir)

	if report.Gate.Pass {
		if err := writeFreeze(*outDir, report); err != nil {
			fmt.Fprintln(os.Stderr, "freeze:", err)
			return 1
		}
		fmt.Println("METHOD_FROZEN written - gate passed, method is now immutable.")
		return 0
	}
	fmt.Println("Gate FAILED (task57 section 21): current ciphertext-only features are insufficient for blind inverse-homophony recovery. Voynich will not be analyzed.")
	return 3
}

func verdictWord(pass bool) string {
	if pass {
		return "PASS"
	}
	return "FAIL"
}
