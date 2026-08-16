package conditionalregime

import (
	"context"
	"path/filepath"
	"testing"
)

func TestCheckpointSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "checkpoint.json")
	cp := newCheckpoint("fp-1")
	cp.WithinCombosDone["joint|A/1|50"] = true
	cp.Candidates = []WithinClassCandidate{{
		Class: ClassID{Scheme: SchemeJoint, Currier: "A", Hand: "1"}, WindowSize: 50, Method: "k_medoids", K: 3,
		Stats: EmpiricalStats{Observed: 0.5, Permutations: 10},
	}}
	cp.ResidualCorrectionNull["k_medoids|raw"] = []float64{0.1, 0.2, 0.3}
	jobID := JobID{Stage: "part_b_global_correction", Combination: "k_medoids|raw", ReplicateIndex: 4}
	cp.PermutationJobs[checkpointJobKey(jobID)] = 0.4

	if err := saveCheckpoint(path, cp); err != nil {
		t.Fatalf("saveCheckpoint: %v", err)
	}
	loaded, ok := loadCheckpoint(path, "fp-1")
	if !ok {
		t.Fatal("expected a matching checkpoint to load")
	}
	if !loaded.WithinCombosDone["joint|A/1|50"] {
		t.Fatal("expected within-class combo to be marked done after round trip")
	}
	if len(loaded.Candidates) != 1 || loaded.Candidates[0].K != 3 || loaded.Candidates[0].Stats.Observed != 0.5 {
		t.Fatalf("candidates did not survive round trip: %+v", loaded.Candidates)
	}
	if got := loaded.ResidualCorrectionNull["k_medoids|raw"]; len(got) != 3 || got[2] != 0.3 {
		t.Fatalf("partial null array did not survive round trip: %v", got)
	}
	if got := loaded.PermutationJobs[checkpointJobKey(jobID)]; got != 0.4 {
		t.Fatalf("JobID-keyed result did not survive checkpoint round trip: %v", got)
	}
}

func TestCheckpointFingerprintMismatchIsRejected(t *testing.T) {
	path := filepath.Join(t.TempDir(), "checkpoint.json")
	if err := saveCheckpoint(path, newCheckpoint("fp-1")); err != nil {
		t.Fatalf("saveCheckpoint: %v", err)
	}
	if _, ok := loadCheckpoint(path, "fp-2"); ok {
		t.Fatal("a checkpoint written for a different fingerprint must never be resumed into this run")
	}
}

func TestCheckpointMissingFileIsNotAnError(t *testing.T) {
	if _, ok := loadCheckpoint(filepath.Join(t.TempDir(), "missing.json"), "fp"); ok {
		t.Fatal("a missing checkpoint file must be treated as a fresh start, not an error")
	}
}

func TestFingerprintChangesWithAnyParameter(t *testing.T) {
	base := Config{
		WindowSizes: []int{50, 100}, ResidualWindowSizes: []int{50, 100, 1000},
		MinClassTokens: 1000, MinBlockTokens: 500, KMin: 2, KMaxWithin: 10, KMaxResidual: 15,
		Permutations: 1000, Seed: 1,
	}
	fp := computeFingerprint(base, "corpus-hash", "meta-hash")

	seedChanged := base
	seedChanged.Seed = 2
	if computeFingerprint(seedChanged, "corpus-hash", "meta-hash") == fp {
		t.Fatal("a different seed must change the fingerprint")
	}
	if computeFingerprint(base, "different-corpus-hash", "meta-hash") == fp {
		t.Fatal("a different corpus hash must change the fingerprint (never resume into a different corpus)")
	}
	if computeFingerprint(base, "corpus-hash", "different-meta-hash") == fp {
		t.Fatal("a different metadata hash must change the fingerprint")
	}
	permChanged := base
	permChanged.Permutations = 500
	if computeFingerprint(permChanged, "corpus-hash", "meta-hash") == fp {
		t.Fatal("a different permutation count must change the fingerprint")
	}
	workersChanged := base
	workersChanged.Workers = 12
	if computeFingerprint(workersChanged, "corpus-hash", "meta-hash") != fp {
		t.Fatal("worker count is operational and must not invalidate a scientific checkpoint")
	}
}

// TestResidualCorrectionResumesFromPartialNull is the crux of the
// checkpoint/resume mechanism: because every replicate is seeded from its
// own index (replicateSeed), resuming from a partial null distribution must
// compute the remaining replicates exactly as a from-scratch run would, and
// must not recompute the ones already supplied.
func TestResidualCorrectionResumesFromPartialNull(t *testing.T) {
	tokens, classes, blocks := syntheticResidualScenario()
	scales := []int{20, 40}

	var full []float64
	fresh := residualGlobalCorrection(tokens, classes, blocks, scales, 2, 3, "k_medoids", false, 0.5, 6, 55, nil, func(null []float64) {
		full = append([]float64(nil), null...)
	})
	if len(full) != 6 {
		t.Fatalf("expected 6 replicates from a fresh run, got %d", len(full))
	}

	partial := append([]float64(nil), full[:3]...)
	newReplicates := 0
	resumed := residualGlobalCorrection(tokens, classes, blocks, scales, 2, 3, "k_medoids", false, 0.5, 6, 55, partial, func(null []float64) {
		newReplicates++
	})
	if newReplicates != 3 {
		t.Fatalf("resuming from 3 already-computed replicates should compute exactly 3 more, got %d calls", newReplicates)
	}
	if resumed != fresh {
		t.Fatalf("resumed stats %+v must exactly match a from-scratch run %+v", resumed, fresh)
	}
}

// TestResidualCorrectionResumeSkipsWhenAlreadyComplete confirms that
// supplying a full-length resume slice recomputes nothing at all.
func TestResidualCorrectionResumeSkipsWhenAlreadyComplete(t *testing.T) {
	tokens, classes, blocks := syntheticResidualScenario()
	scales := []int{20, 40}
	complete := []float64{0.11, 0.22, 0.33, 0.44}
	calls := 0
	stats := residualGlobalCorrection(tokens, classes, blocks, scales, 2, 3, "k_medoids", false, 0.9, 4, 7, complete, func(null []float64) {
		calls++
	})
	if calls != 0 {
		t.Fatalf("a fully-supplied resume slice must trigger zero new replicates, got %d", calls)
	}
	if stats.Permutations != 4 || stats.Observed != 0.9 {
		t.Fatalf("unexpected stats from an already-complete resume: %+v", stats)
	}
}

func TestResidualCorrectionWorkerCountIndependence(t *testing.T) {
	tokens, classes, blocks := syntheticResidualScenario()
	scales := []int{20, 40}
	one, err := residualGlobalCorrectionParallel(context.Background(), 1, tokens, classes, blocks, scales, 2, 3, "k_medoids", false, 0.5, 8, 55, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	four, err := residualGlobalCorrectionParallel(context.Background(), 4, tokens, classes, blocks, scales, 2, 3, "k_medoids", false, 0.5, 8, 55, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if one != four {
		t.Fatalf("scientific reduction changed with worker count:\nworkers=1 %+v\nworkers=4 %+v", one, four)
	}
}
