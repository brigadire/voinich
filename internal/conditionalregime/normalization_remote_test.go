package conditionalregime

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"zcore.dev/voinich/internal/normalization"
	"zcore.dev/voinich/internal/normalizationcompare"
	"zcore.dev/voinich/internal/sequenceanalyze"
)

func loadNormalizationCorpusForTest(t *testing.T, path string) (normalization.Corpus, error) {
	t.Helper()
	return normalization.LoadCorpus(path)
}

func defaultSequenceParams() sequenceanalyze.Parameters { return sequenceanalyze.DefaultParameters() }

// normalizationRemoteFixture writes a small corpus and structural_classes.yaml
// with one multi-member class (so RandomModel/RunRandomTrial has real merge
// work to do) and returns the normalizationcompare.Config every test in this
// file starts from.
func normalizationRemoteFixture(t *testing.T) normalizationcompare.Config {
	t.Helper()
	d := t.TempDir()
	write := func(name, body string) string {
		p := filepath.Join(d, name)
		if err := os.WriteFile(p, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	corpus := strings.Repeat("a b c a c b\n", 30)
	classes := `
meta:
  input_corpus: corpus.txt
  singleton_mode: token
  min_token_count: 1
  random_matching: "matched by frequency bin"
models:
  - threshold: 0.8
    label: "080"
    stats:
      multi_member_classes: 1
      classes: 2
    classes:
      - id: C0001
        size: 2
        members:
          - token: a
            count: 60
          - token: b
            count: 60
`
	return normalizationcompare.Config{
		InputPath: write("corpus.txt", corpus), ClassesPath: write("structural_classes.yaml", classes),
		RandomRuns: 3, RandomSeed: 1, Workers: 1, RemoteTimeout: 5 * time.Second, RemoteRetries: 2,
	}
}

func TestNormalizationRemoteMatchesLocalOracle(t *testing.T) {
	c := normalizationRemoteFixture(t)
	classes, err := normalizationcompare.LoadClasses(c.ClassesPath)
	if err != nil {
		t.Fatal(err)
	}
	corpus, err := loadNormalizationCorpusForTest(t, c.InputPath)
	if err != nil {
		t.Fatal(err)
	}
	want, err := normalizationcompare.RunRandomTrial(classes.Models[0], corpus, classes.Meta.MinTokenCount, classes.Meta.SingletonMode, c.RandomSeed, 0, defaultSequenceParams())
	if err != nil {
		t.Fatal(err)
	}

	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	workerCrt, workerKey := pkiFix.issueWorker(t, "normalization-worker")
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	ex, err := NewNormalizationRemoteExecutor(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ex.Close()
	pool := ex.(*normalizationExecutorAdapter).pool.(*remotePool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startRemoteWorker(t, ctx, "https://localhost:"+strings.Split(pool.Addr(), ":")[1], pkiFix.caCrt, workerCrt, workerKey, 4)

	got, err := ex.Run(context.Background(), "080", 0)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("remote baseline differs\ngot=%#v\nwant=%#v", got, want)
	}
}

func TestNormalizationTwoRemoteWorkersMatchLocalInAnyCompletionOrder(t *testing.T) {
	c := normalizationRemoteFixture(t)
	classes, err := normalizationcompare.LoadClasses(c.ClassesPath)
	if err != nil {
		t.Fatal(err)
	}
	corpus, err := loadNormalizationCorpusForTest(t, c.InputPath)
	if err != nil {
		t.Fatal(err)
	}
	want := make([]normalizationcompare.BaselineResult, c.RandomRuns)
	for i := range want {
		want[i], err = normalizationcompare.RunRandomTrial(classes.Models[0], corpus, classes.Meta.MinTokenCount, classes.Meta.SingletonMode, c.RandomSeed, i, defaultSequenceParams())
		if err != nil {
			t.Fatal(err)
		}
	}

	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	ex, err := NewNormalizationRemoteExecutor(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ex.Close()
	pool := ex.(*normalizationExecutorAdapter).pool.(*remotePool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	url := "https://localhost:" + strings.Split(pool.Addr(), ":")[1]
	for _, name := range []string{"normalization-worker-a", "normalization-worker-b"} {
		crt, key := pkiFix.issueWorker(t, name)
		startRemoteWorker(t, ctx, url, pkiFix.caCrt, crt, key, 4)
	}

	got := make([]normalizationcompare.BaselineResult, len(want))
	errCh := make(chan error, len(want))
	var wg sync.WaitGroup
	for i := range got {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var runErr error
			got[i], runErr = ex.Run(context.Background(), "080", i)
			errCh <- runErr
		}(i)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("multi-worker remote baselines differ from local oracle in canonical order\ngot=%#v\nwant=%#v", got, want)
	}
}

// TestNormalizationRemoteRetryAfterWorkerFailure mirrors
// TestRemoteJobReassignedAfterLeaseExpiryMatchesLocal: a worker leases the
// job over raw HTTP and then goes silent forever (no /v1/result), forcing
// the coordinator to reclaim the lease after RemoteTimeout and complete the
// job with a second, distinct authenticated worker - byte-identical to the
// local oracle. A deterministic crash-via-raw-lease beats racing a real
// worker process against a computation that may finish in microseconds.
func TestNormalizationRemoteRetryAfterWorkerFailure(t *testing.T) {
	c := normalizationRemoteFixture(t)
	c.RemoteTimeout = 150 * time.Millisecond
	classes, err := normalizationcompare.LoadClasses(c.ClassesPath)
	if err != nil {
		t.Fatal(err)
	}
	corpus, err := loadNormalizationCorpusForTest(t, c.InputPath)
	if err != nil {
		t.Fatal(err)
	}
	want, err := normalizationcompare.RunRandomTrial(classes.Models[0], corpus, classes.Meta.MinTokenCount, classes.Meta.SingletonMode, c.RandomSeed, 1, defaultSequenceParams())
	if err != nil {
		t.Fatal(err)
	}

	pkiFix := newRemotePKI(t, []string{"localhost"}, []string{"127.0.0.1"})
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	fp, err := normalizationcompare.Fingerprint(c.InputPath, c.ClassesPath, classes.Meta.MinTokenCount, classes.Meta.SingletonMode, c.RandomSeed, c.RandomRuns)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := newNormalizationRemotePool(c, classes, fp)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	base := "https://" + pool.Addr()

	id := JobID{Stage: "normalization_compare_baseline", Combination: "080", ReplicateIndex: 1}
	resultCh := make(chan struct {
		blob []byte
		err  error
	}, 1)
	go func() {
		b, runErr := pool.RunBlob(context.Background(), id)
		resultCh <- struct {
			blob []byte
			err  error
		}{b, runErr}
	}()

	crashCrt, crashKey := pkiFix.issueWorker(t, "normalization-worker-crash")
	crashClient := workerHTTPClient(t, crashCrt, crashKey, pkiFix.caCrt)
	leased := leaseUntilAssigned(t, crashClient, base, fp)
	if leased.JobID != id {
		t.Fatalf("crashing worker leased %+v, want %+v", leased.JobID, id)
	}
	// The crashing worker now goes silent forever: no /v1/result ever
	// follows. The coordinator must reclaim the lease after RemoteTimeout.

	goodCrt, goodKey := pkiFix.issueWorker(t, "normalization-worker-good")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startRemoteWorker(t, ctx, base, pkiFix.caCrt, goodCrt, goodKey, 1)

	select {
	case res := <-resultCh:
		if res.err != nil {
			t.Fatalf("reassigned job failed: %v", res.err)
		}
		var got normalizationcompare.BaselineResult
		if err := json.Unmarshal(res.blob, &got); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("retried remote baseline differs from local oracle\ngot=%#v\nwant=%#v", got, want)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for the job to be reassigned and completed")
	}
	if pool.leasesReclaimed.Load() == 0 {
		t.Fatal("expected at least one reclaimed lease from the crashing worker")
	}
}

// TestNormalizationRemoteSameSeedRepeated confirms two independent remote
// runs of the identical (corpus, classes, seed, threshold, run) produce
// byte-identical results - repetition is not a source of nondeterminism.
func TestNormalizationRemoteSameSeedRepeated(t *testing.T) {
	c := normalizationRemoteFixture(t)
	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	workerCrt, workerKey := pkiFix.issueWorker(t, "normalization-worker")
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	ex, err := NewNormalizationRemoteExecutor(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ex.Close()
	pool := ex.(*normalizationExecutorAdapter).pool.(*remotePool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startRemoteWorker(t, ctx, "https://localhost:"+strings.Split(pool.Addr(), ":")[1], pkiFix.caCrt, workerCrt, workerKey, 4)

	first, err := ex.Run(context.Background(), "080", 2)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ex.Run(context.Background(), "080", 2)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("repeated same-seed remote run diverged\nfirst=%#v\nsecond=%#v", first, second)
	}
}
