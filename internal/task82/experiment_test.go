package task82

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"zcore.dev/voinich/internal/mnemonicspace"
)

func repositoryRoot() string { return filepath.Join("..", "..") }

func TestFrozenManifestAndJobIdentities(t *testing.T) {
	_, manifest, _, _, err := verifyFreeze(repositoryRoot())
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Jobs) != 672 {
		t.Fatalf("jobs=%d", len(manifest.Jobs))
	}
}

func TestDeterministicRegenerationAndLeakage(t *testing.T) {
	authority, manifest, specs, corpora, err := verifyFreeze(repositoryRoot())
	if err != nil {
		t.Fatal(err)
	}
	for _, index := range []int{0, 210, 671} {
		j := manifest.Jobs[index]
		s := specs[j.MechanismID]
		p, ok := s.ParameterSet(j.ParameterSetID)
		if !ok {
			t.Fatal(j.ParameterSetID)
		}
		a, err := runJob(authority, manifest, j, s, p, corpora[j.InputCorpusID])
		if err != nil {
			t.Fatal(err)
		}
		b, err := runJob(authority, manifest, j, s, p, corpora[j.InputCorpusID])
		if err != nil {
			t.Fatal(err)
		}
		aj, _ := json.Marshal(a)
		bj, _ := json.Marshal(b)
		if string(aj) != string(bj) {
			at, bt := string(aj), string(bj)
			for n := 0; n < len(at) && n < len(bt); n++ {
				if at[n] != bt[n] {
					lo, hi, hj := n-120, n+120, n+120
					if lo < 0 {
						lo = 0
					}
					if hi > len(at) {
						hi = len(at)
					}
					if hj > len(bt) {
						hj = len(bt)
					}
					t.Fatalf("job %s differs at byte %d: %q != %q", j.JobID, n, at[lo:hi], bt[lo:hj])
				}
			}
			t.Fatalf("job %s differs in encoded length %d != %d", j.JobID, len(at), len(bt))
		}
		if err := validateDocument(s, a.Observable); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRecoveryInvariantsAndNotApplicable(t *testing.T) {
	_, manifest, _, _, err := verifyFreeze(repositoryRoot())
	if err != nil {
		t.Fatal(err)
	}
	raw := filepath.Join(repositoryRoot(), "research", "phase2", "task82", "raw")
	if err := verifyRaw(manifest, raw, 0, 1); err != nil {
		t.Fatal(err)
	}
	exactRequired := map[string]bool{"f01_speculum_core": true, "f01_speculum_profile_latin23_r12": true, "f08_serpens_core": true, "synthetic_literal_storage": true}
	for _, j := range manifest.Jobs {
		if j.RecoveryCondition != string(mnemonicspace.RecoveryFullKnowledge) || !exactRequired[j.MechanismID] {
			continue
		}
		var a Artifact
		if err := readJSON(filepath.Join(raw, j.JobID+".json"), &a); err != nil {
			t.Fatal(err)
		}
		if !a.Metrics.ExactMatch {
			t.Fatalf("required R0 invariant failed: %s", j.JobID)
		}
	}
}
