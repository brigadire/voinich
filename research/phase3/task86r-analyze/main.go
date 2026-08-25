package main

import (
	"fmt"
	"os"
)

const namespace = "voinich.phase3.task86r.g1.v1.1"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		writeBlockedMarker(err)
		os.Exit(1)
	}
}

// writeBlockedMarker issues TASK86R_EXPERIMENT_BLOCKED when run() fails
// before a scientific verdict could be produced (preflight/hash/contract
// failures), per task86r.txt section 4/59: a failed methodological
// preflight is never reported as a scientific Voynichese finding.
func writeBlockedMarker(cause error) {
	_ = os.MkdirAll(outDir, 0o755)
	content := fmt.Sprintf("TASK86R_EXPERIMENT_BLOCKED\nversion=task86r-v1\ncode_revision=%s\nreason=%s\n", gitHeadShort(), cause.Error())
	_ = os.WriteFile(outDir+"/TASK86R_EXPERIMENT_BLOCKED", []byte(content), 0o644)
}
