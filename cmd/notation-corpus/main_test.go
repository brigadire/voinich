package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestPreparationLocks(t *testing.T) {
	cases := [][]string{{"analyze", "--input", "x", "--output", "y", "--manifest", "z"}, {"compare-vm"}, {"compare-classes", "--authorize-production"}}
	for _, args := range cases {
		err := run(args)
		if err == nil {
			t.Fatalf("%v unexpectedly authorized", args)
		}
		msg := err.Error()
		if !strings.Contains(msg, "lock") && !strings.Contains(msg, "not authorized") && !strings.Contains(msg, "authorization") {
			t.Fatalf("%v: %v", args, err)
		}
	}
}

// TestA10CandidateGuardRejectsC01ThroughC09 checks the B03-B04-B01
// preparation commands refuse to run against anything shaped like a
// C01-C09 candidate corpus id, independent of the separate
// production-authorization lock exercised above (adversarial test A10:
// C01-C09 full runs must remain blocked during this task).
func TestA10CandidateGuardRejectsC01ThroughC09(t *testing.T) {
	for i := 1; i <= 9; i++ {
		id := fmt.Sprintf("C%02d-SOMECORPUS", i)
		if err := guardNotUnauthorizedCandidate(id, false); err == nil {
			t.Fatalf("corpus_id %q should have been rejected without --fixture", id)
		}
		if err := guardNotUnauthorizedCandidate(id, true); err != nil {
			t.Fatalf("corpus_id %q with --fixture should be allowed: %v", id, err)
		}
	}
	for _, id := range []string{"VM-ZL3b-x7", "CAL-IID", "SYN-RAREFY", "C10-SYNTHETIC"} {
		if err := guardNotUnauthorizedCandidate(id, false); err != nil {
			t.Fatalf("corpus_id %q should not require --fixture: %v", id, err)
		}
	}
}
