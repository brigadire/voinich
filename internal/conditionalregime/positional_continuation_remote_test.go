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

	"zcore.dev/voinich/internal/positionalcontinuation"
)

// positionalContinuationRemoteFixture writes a small generic corpus plus
// the minimal, self-consistent higher-order-dir files loadCorpusAndBlocks/
// resolveGenericTarget need: a frozen HIGHER_ORDER_REPLICATED "s aiin chey"
// target triple in higher_order_validation.tsv. Every other
// HigherOrderDirFiles entry only needs to exist (Fingerprint hashes its
// bytes; resolveGenericTarget never parses it), so those get a header-only
// placeholder.
func positionalContinuationRemoteFixture(t *testing.T) positionalcontinuation.Config {
	t.Helper()
	higherOrderDir := t.TempDir()
	writeIn := func(name, body string) {
		if err := os.WriteFile(filepath.Join(higherOrderDir, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	corpus := strings.Repeat("aiin chey shey ol or dy qokeey s\n", 40)
	corpusPath := filepath.Join(t.TempDir(), "corpus.txt")
	if err := os.WriteFile(corpusPath, []byte(corpus), 0600); err != nil {
		t.Fatal(err)
	}

	for _, name := range positionalcontinuation.HigherOrderDirFiles {
		if name == "higher_order_validation.tsv" || strings.HasSuffix(name, ".yaml") {
			continue
		}
		writeIn(name, "sequence\n")
	}
	writeIn("higher_order_sequence_analysis.yaml", "meta: {}\n")
	writeIn("higher_order_validation.tsv",
		"sequence\tfinal_status\tconditional_fdr_q\n"+
			"s aiin chey\tHIGHER_ORDER_REPLICATED\t0.01\n")

	return positionalcontinuation.Config{
		CorpusPath: corpusPath, HigherOrderDir: higherOrderDir, Generic: true,
		Permutations: 50, Seed: 1,
		Workers: 1, RemoteTimeout: 5 * time.Second, RemoteRetries: 2,
	}
}

func positionalContinuationOracleReplicate(t *testing.T, c positionalcontinuation.Config, battery string) positionalcontinuation.BatteryResult {
	t.Helper()
	sAiinOccs, aiinOccs, err := positionalcontinuation.LoadForDistribution(c)
	if err != nil {
		t.Fatal(err)
	}
	got, err := positionalcontinuation.ComputeBattery(sAiinOccs, aiinOccs, battery, c.Permutations, c.Seed)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestPositionalContinuationRemoteMatchesLocalOracle(t *testing.T) {
	c := positionalContinuationRemoteFixture(t)
	want := positionalContinuationOracleReplicate(t, c, "postest_line")

	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	workerCrt, workerKey := pkiFix.issueWorker(t, "positional-continuation-worker")
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	ex, err := NewPositionalContinuationRemoteExecutor(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ex.(*positionalContinuationExecutorAdapter).pool.Close()
	pool := ex.(*positionalContinuationExecutorAdapter).pool.(*remotePool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startRemoteWorker(t, ctx, "https://localhost:"+strings.Split(pool.Addr(), ":")[1], pkiFix.caCrt, workerCrt, workerKey, 4)

	got, err := ex.Run(context.Background(), "postest_line")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("remote battery result differs\ngot=%#v\nwant=%#v", got, want)
	}
}

// TestPositionalContinuationTwoRemoteWorkersMatchLocalInAnyCompletionOrder
// dispatches all 5 distinct distributable batteries concurrently across two
// workers: since a JobID already in flight is rejected by the coordinator
// (remotePool.runOutcome), and this stage's job identity is the battery
// name with a fixed ReplicateIndex of 0, concurrency here is necessarily
// across distinct batteries, never repeats of one battery's JobID.
func TestPositionalContinuationTwoRemoteWorkersMatchLocalInAnyCompletionOrder(t *testing.T) {
	c := positionalContinuationRemoteFixture(t)
	batteries := []string{"postest_line", "postest_block", "stratified_line", "stratified_block", "boundary"}
	want := make(map[string]positionalcontinuation.BatteryResult, len(batteries))
	for _, b := range batteries {
		want[b] = positionalContinuationOracleReplicate(t, c, b)
	}

	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	ex, err := NewPositionalContinuationRemoteExecutor(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ex.(*positionalContinuationExecutorAdapter).pool.Close()
	pool := ex.(*positionalContinuationExecutorAdapter).pool.(*remotePool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	url := "https://localhost:" + strings.Split(pool.Addr(), ":")[1]
	for _, name := range []string{"positional-continuation-worker-a", "positional-continuation-worker-b"} {
		crt, key := pkiFix.issueWorker(t, name)
		startRemoteWorker(t, ctx, url, pkiFix.caCrt, crt, key, 4)
	}

	got := make(map[string]positionalcontinuation.BatteryResult, len(batteries))
	var mu sync.Mutex
	errCh := make(chan error, len(batteries))
	var wg sync.WaitGroup
	for _, b := range batteries {
		wg.Add(1)
		go func(b string) {
			defer wg.Done()
			r, runErr := ex.Run(context.Background(), b)
			mu.Lock()
			got[b] = r
			mu.Unlock()
			errCh <- runErr
		}(b)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("multi-worker remote battery results differ from local oracle\ngot=%#v\nwant=%#v", got, want)
	}
}

// TestPositionalContinuationRemoteRetryAfterWorkerFailure mirrors
// TestTokenRelationRemoteRetryAfterWorkerFailure: a worker leases the job
// over raw HTTP and then goes silent forever, forcing the coordinator to
// reclaim the lease after RemoteTimeout and complete the job with a
// second, distinct authenticated worker - byte-identical to the local
// oracle.
func TestPositionalContinuationRemoteRetryAfterWorkerFailure(t *testing.T) {
	c := positionalContinuationRemoteFixture(t)
	c.RemoteTimeout = 150 * time.Millisecond
	want := positionalContinuationOracleReplicate(t, c, "boundary")

	pkiFix := newRemotePKI(t, []string{"localhost"}, []string{"127.0.0.1"})
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	fp, err := positionalcontinuation.Fingerprint(c)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := newPositionalContinuationRemotePool(c, fp)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	base := "https://" + pool.Addr()

	id := JobID{Stage: "positional_continuation_battery", Combination: "boundary", ReplicateIndex: 0}
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

	crashCrt, crashKey := pkiFix.issueWorker(t, "positional-continuation-worker-crash")
	crashClient := workerHTTPClient(t, crashCrt, crashKey, pkiFix.caCrt)
	leased := leaseUntilAssigned(t, crashClient, base, fp)
	if leased.JobID != id {
		t.Fatalf("crashing worker leased %+v, want %+v", leased.JobID, id)
	}
	// The crashing worker now goes silent forever: no /v1/result ever
	// follows. The coordinator must reclaim the lease after RemoteTimeout.

	goodCrt, goodKey := pkiFix.issueWorker(t, "positional-continuation-worker-good")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startRemoteWorker(t, ctx, base, pkiFix.caCrt, goodCrt, goodKey, 1)

	select {
	case res := <-resultCh:
		if res.err != nil {
			t.Fatalf("reassigned job failed: %v", res.err)
		}
		var got positionalcontinuation.BatteryResult
		if err := json.Unmarshal(res.blob, &got); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("retried remote battery result differs from local oracle\ngot=%#v\nwant=%#v", got, want)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for the job to be reassigned and completed")
	}
	if pool.leasesReclaimed.Load() == 0 {
		t.Fatal("expected at least one reclaimed lease from the crashing worker")
	}
}

// TestPositionalContinuationRemoteSameSeedRepeated confirms two independent
// remote runs of the identical (corpus, higher-order inputs, seed, battery)
// produce byte-identical results.
func TestPositionalContinuationRemoteSameSeedRepeated(t *testing.T) {
	c := positionalContinuationRemoteFixture(t)
	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	workerCrt, workerKey := pkiFix.issueWorker(t, "positional-continuation-worker")
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	ex, err := NewPositionalContinuationRemoteExecutor(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ex.(*positionalContinuationExecutorAdapter).pool.Close()
	pool := ex.(*positionalContinuationExecutorAdapter).pool.(*remotePool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startRemoteWorker(t, ctx, "https://localhost:"+strings.Split(pool.Addr(), ":")[1], pkiFix.caCrt, workerCrt, workerKey, 4)

	first, err := ex.Run(context.Background(), "postest_line")
	if err != nil {
		t.Fatal(err)
	}
	second, err := ex.Run(context.Background(), "postest_line")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("repeated same-seed remote run diverged\nfirst=%#v\nsecond=%#v", first, second)
	}
}

// TestPositionalContinuationRemoteInterruptedThenResumedMatchesOracle
// simulates a coordinator crash-and-restart mid-battery-set: a fresh pool
// completes only the first battery, is closed (as if the process died),
// and a second fresh pool (a new coordinator generation - exactly what a
// real resumed pipeline-orchestrate run would start) completes the rest -
// proving a resumed remote run reproduces the same per-battery values an
// uninterrupted run would have. Batteries (not permutation runs) are the
// resume boundary here because a whole battery is this stage's atomic
// unit of distributed work.
func TestPositionalContinuationRemoteInterruptedThenResumedMatchesOracle(t *testing.T) {
	c := positionalContinuationRemoteFixture(t)
	batteries := []string{"postest_line", "postest_block", "stratified_line", "stratified_block", "boundary"}
	want := make(map[string]positionalcontinuation.BatteryResult, len(batteries))
	for _, b := range batteries {
		want[b] = positionalContinuationOracleReplicate(t, c, b)
	}

	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt

	runRange := func(bs []string, workerID string) map[string]positionalcontinuation.BatteryResult {
		ex, err := NewPositionalContinuationRemoteExecutor(c)
		if err != nil {
			t.Fatal(err)
		}
		pool := ex.(*positionalContinuationExecutorAdapter).pool.(*remotePool)
		ctx, cancel := context.WithCancel(context.Background())
		crt, key := pkiFix.issueWorker(t, workerID)
		startRemoteWorker(t, ctx, "https://localhost:"+strings.Split(pool.Addr(), ":")[1], pkiFix.caCrt, crt, key, 2)
		out := make(map[string]positionalcontinuation.BatteryResult, len(bs))
		for _, b := range bs {
			r, err := ex.Run(context.Background(), b)
			if err != nil {
				t.Fatal(err)
			}
			out[b] = r
		}
		cancel()
		_ = ex.(*positionalContinuationExecutorAdapter).pool.Close()
		return out
	}

	firstHalf := runRange(batteries[:2], "positional-continuation-resume-worker-1")
	secondHalf := runRange(batteries[2:], "positional-continuation-resume-worker-2")
	got := map[string]positionalcontinuation.BatteryResult{}
	for k, v := range firstHalf {
		got[k] = v
	}
	for k, v := range secondHalf {
		got[k] = v
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("interrupted-then-resumed remote run differs from an uninterrupted oracle\ngot=%#v\nwant=%#v", got, want)
	}
}
