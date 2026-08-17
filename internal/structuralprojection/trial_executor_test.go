package structuralprojection

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"
)

type indexedTrialExecutor struct {
	mu    sync.Mutex
	calls []int
	delay bool
}

func (e *indexedTrialExecutor) Run(_ context.Context, i int) (TrialResult, error) {
	if e.delay {
		time.Sleep(time.Duration(4-i) * time.Millisecond)
	}
	e.mu.Lock()
	e.calls = append(e.calls, i)
	e.mu.Unlock()
	v := float64(i) + .25
	return TrialResult{Random: []float64{v}, Smoothing: []float64{v + 1}, RandomAblated: []float64{v + 2}, SmoothingAblated: []float64{v + 3}, RandomByDistance: [][]float64{{v + 4}}}, nil
}
func (e *indexedTrialExecutor) Close() error { return nil }

func trialTestConfig(t *testing.T, n int) Config {
	t.Helper()
	d := t.TempDir()
	paths := []*string{}
	c := Config{RandomProjections: n, Workers: n, Seed: 7}
	paths = []*string{&c.CorpusPath, &c.StructuralPairsPath, &c.DistancePairsPath, &c.FamiliesPath}
	for i, p := range paths {
		name := filepath.Join(d, string(rune('a'+i)))
		if err := os.WriteFile(name, []byte{byte(i)}, 0600); err != nil {
			t.Fatal(err)
		}
		*p = name
	}
	return c
}

func TestRunTrialsCanonicalizesOutOfOrderCompletion(t *testing.T) {
	c := trialTestConfig(t, 4)
	e := &indexedTrialExecutor{delay: true}
	got, err := runTrials(c, e, newProgress(nil))
	if err != nil {
		t.Fatal(err)
	}
	for i, r := range got {
		if r.Random[0] != float64(i)+.25 {
			t.Fatalf("slot %d=%v", i, r)
		}
	}
}

func TestRunTrialsResumeSkipsCheckpointedTrials(t *testing.T) {
	c := trialTestConfig(t, 3)
	c.CheckpointPath = t.TempDir() + "/checkpoint.json"
	fp, err := Fingerprint(c)
	if err != nil {
		t.Fatal(err)
	}
	saved := TrialResult{Random: []float64{.25}, Smoothing: []float64{1.25}, RandomAblated: []float64{2.25}, SmoothingAblated: []float64{3.25}, RandomByDistance: [][]float64{{4.25}}}
	if err := saveTrialCheckpoint(c.CheckpointPath, fp, map[int]TrialResult{0: saved}); err != nil {
		t.Fatal(err)
	}
	e := &indexedTrialExecutor{}
	got, err := runTrials(c, e, newProgress(nil))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got[0], saved) {
		t.Fatalf("resumed=%v", got[0])
	}
	for _, i := range e.calls {
		if i == 0 {
			t.Fatal("checkpointed trial recomputed")
		}
	}
}

func TestCanonicalTrialOrderRejectsMissing(t *testing.T) {
	if _, err := CanonicalTrialOrder(map[int]TrialResult{1: {}}, 2); err == nil {
		t.Fatal("missing trial accepted")
	}
}
