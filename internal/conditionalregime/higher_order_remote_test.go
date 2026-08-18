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

	"zcore.dev/voinich/internal/higherorderseq"
)

// higherOrderRemoteFixture writes a small generic corpus plus the minimal,
// self-consistent audit-dir/discovery-dir files loadFrozenCandidates/
// structuralRelatives need: two frozen n=3 primary candidates
// ("aiin chey shey" and "chey shey ol") passing both the n>=3 and
// shuffle_block_fdr_q<=0.05 gates, each with a matching markov_block_p row.
// Two distinct candidates (rather than one repeated) matter here because
// runOutcome rejects a JobID that is already in flight - concurrent
// dispatch tests need concurrently-distinct JobIDs, exactly as a real
// multi-candidate battery would produce. Every other
// AuditDirFiles/DiscoveryDirFiles entry only needs to exist (Fingerprint
// hashes its bytes; loadFrozenCandidates/structuralRelatives never parse
// it), so those get a header-only/empty placeholder.
func higherOrderRemoteFixture(t *testing.T) higherorderseq.Config {
	t.Helper()
	auditDir := t.TempDir()
	discoveryDir := t.TempDir()
	writeIn := func(dir, name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	corpus := strings.Repeat("aiin chey shey ol or dy qokeey s\n", 40)
	corpusPath := filepath.Join(t.TempDir(), "corpus.txt")
	if err := os.WriteFile(corpusPath, []byte(corpus), 0600); err != nil {
		t.Fatal(err)
	}

	writeIn(auditDir, "universal_sequence_inventory.tsv", "sequence\n")
	writeIn(auditDir, "sequence_replication_status.tsv", "sequence\n")
	writeIn(auditDir, "replicated_local_structure.yaml", "meta: {}\n")
	writeIn(auditDir, "strict_replicated_sequences.tsv",
		"sequence\tn\tcanonical_occurrences\tphysical_blocks\tjoint_classes\tmax_block_fraction\tshuffle_block_fdr_q\n"+
			"aiin chey shey\t3\t4\t4\t3\t0.25\t0.01\n"+
			"chey shey ol\t3\t4\t4\t3\t0.25\t0.01\n")
	writeIn(auditDir, "sequence_null_validation.tsv",
		"sequence\tmarkov_block_p\n"+
			"aiin chey shey\t0.01\n"+
			"chey shey ol\t0.01\n")
	writeIn(discoveryDir, "structural_classes.yaml", "models: []\n")

	return higherorderseq.Config{
		CorpusPath: corpusPath, AuditDir: auditDir, DiscoveryDir: discoveryDir, Generic: true,
		Permutations: 50, Seed: 1,
		Workers: 1, RemoteTimeout: 5 * time.Second, RemoteRetries: 2,
	}
}

func higherOrderOracleReplicate(t *testing.T, c higherorderseq.Config, sequence string) higherorderseq.CandidateResult {
	t.Helper()
	candidates, blocks, lineLength, relatives, err := higherorderseq.LoadForDistribution(c)
	if err != nil {
		t.Fatal(err)
	}
	got, err := higherorderseq.ComputeCandidate(candidates, blocks, lineLength, relatives, c.Permutations, c.Seed, sequence)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestHigherOrderRemoteMatchesLocalOracle(t *testing.T) {
	c := higherOrderRemoteFixture(t)
	want := higherOrderOracleReplicate(t, c, "aiin chey shey")

	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	workerCrt, workerKey := pkiFix.issueWorker(t, "higher-order-worker")
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	ex, err := NewHigherOrderRemoteExecutor(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ex.(*higherOrderExecutorAdapter).pool.Close()
	pool := ex.(*higherOrderExecutorAdapter).pool.(*remotePool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startRemoteWorker(t, ctx, "https://localhost:"+strings.Split(pool.Addr(), ":")[1], pkiFix.caCrt, workerCrt, workerKey, 4)

	got, err := ex.Run(context.Background(), "aiin chey shey")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("remote candidate result differs\ngot=%#v\nwant=%#v", got, want)
	}
}

// TestHigherOrderTwoRemoteWorkersMatchLocalInAnyCompletionOrder dispatches
// two distinct frozen candidates concurrently across two workers: since
// higher-order-sequence-validate's distributable unit is a whole candidate
// (never a permutation - see CandidateExecutor's doc comment), and the
// coordinator rejects a JobID that is already in flight, the concurrency
// this stage can exercise is across distinct candidates, not repeats of
// one candidate's JobID.
func TestHigherOrderTwoRemoteWorkersMatchLocalInAnyCompletionOrder(t *testing.T) {
	c := higherOrderRemoteFixture(t)
	sequences := []string{"aiin chey shey", "chey shey ol"}
	want := make(map[string]higherorderseq.CandidateResult, len(sequences))
	for _, seq := range sequences {
		want[seq] = higherOrderOracleReplicate(t, c, seq)
	}

	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	ex, err := NewHigherOrderRemoteExecutor(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ex.(*higherOrderExecutorAdapter).pool.Close()
	pool := ex.(*higherOrderExecutorAdapter).pool.(*remotePool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	url := "https://localhost:" + strings.Split(pool.Addr(), ":")[1]
	for _, name := range []string{"higher-order-worker-a", "higher-order-worker-b"} {
		crt, key := pkiFix.issueWorker(t, name)
		startRemoteWorker(t, ctx, url, pkiFix.caCrt, crt, key, 4)
	}

	got := make(map[string]higherorderseq.CandidateResult, len(sequences))
	var mu sync.Mutex
	errCh := make(chan error, len(sequences))
	var wg sync.WaitGroup
	for _, seq := range sequences {
		wg.Add(1)
		go func(seq string) {
			defer wg.Done()
			r, runErr := ex.Run(context.Background(), seq)
			mu.Lock()
			got[seq] = r
			mu.Unlock()
			errCh <- runErr
		}(seq)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("multi-worker remote candidate results differ from local oracle\ngot=%#v\nwant=%#v", got, want)
	}
}

// TestHigherOrderRemoteRetryAfterWorkerFailure mirrors
// TestTokenRelationRemoteRetryAfterWorkerFailure: a worker leases the job
// over raw HTTP and then goes silent forever, forcing the coordinator to
// reclaim the lease after RemoteTimeout and complete the job with a
// second, distinct authenticated worker - byte-identical to the local
// oracle.
func TestHigherOrderRemoteRetryAfterWorkerFailure(t *testing.T) {
	c := higherOrderRemoteFixture(t)
	c.RemoteTimeout = 150 * time.Millisecond
	want := higherOrderOracleReplicate(t, c, "aiin chey shey")

	pkiFix := newRemotePKI(t, []string{"localhost"}, []string{"127.0.0.1"})
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	fp, err := higherorderseq.Fingerprint(c)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := newHigherOrderRemotePool(c, fp)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	base := "https://" + pool.Addr()

	id := JobID{Stage: "higher_order_candidate", Combination: "aiin chey shey", ReplicateIndex: 0}
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

	crashCrt, crashKey := pkiFix.issueWorker(t, "higher-order-worker-crash")
	crashClient := workerHTTPClient(t, crashCrt, crashKey, pkiFix.caCrt)
	leased := leaseUntilAssigned(t, crashClient, base, fp)
	if leased.JobID != id {
		t.Fatalf("crashing worker leased %+v, want %+v", leased.JobID, id)
	}
	// The crashing worker now goes silent forever: no /v1/result ever
	// follows. The coordinator must reclaim the lease after RemoteTimeout.

	goodCrt, goodKey := pkiFix.issueWorker(t, "higher-order-worker-good")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startRemoteWorker(t, ctx, base, pkiFix.caCrt, goodCrt, goodKey, 1)

	select {
	case res := <-resultCh:
		if res.err != nil {
			t.Fatalf("reassigned job failed: %v", res.err)
		}
		var got higherorderseq.CandidateResult
		if err := json.Unmarshal(res.blob, &got); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("retried remote candidate result differs from local oracle\ngot=%#v\nwant=%#v", got, want)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for the job to be reassigned and completed")
	}
	if pool.leasesReclaimed.Load() == 0 {
		t.Fatal("expected at least one reclaimed lease from the crashing worker")
	}
}

// TestHigherOrderRemoteSameSeedRepeated confirms two independent remote
// runs of the identical (corpus, audit/discovery inputs, seed, candidate)
// produce byte-identical results.
func TestHigherOrderRemoteSameSeedRepeated(t *testing.T) {
	c := higherOrderRemoteFixture(t)
	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	workerCrt, workerKey := pkiFix.issueWorker(t, "higher-order-worker")
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	ex, err := NewHigherOrderRemoteExecutor(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ex.(*higherOrderExecutorAdapter).pool.Close()
	pool := ex.(*higherOrderExecutorAdapter).pool.(*remotePool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startRemoteWorker(t, ctx, "https://localhost:"+strings.Split(pool.Addr(), ":")[1], pkiFix.caCrt, workerCrt, workerKey, 4)

	first, err := ex.Run(context.Background(), "aiin chey shey")
	if err != nil {
		t.Fatal(err)
	}
	second, err := ex.Run(context.Background(), "aiin chey shey")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("repeated same-seed remote run diverged\nfirst=%#v\nsecond=%#v", first, second)
	}
}

// TestHigherOrderRemoteInterruptedThenResumedMatchesOracle simulates a
// coordinator crash-and-restart mid-battery: a fresh pool completes only
// the first frozen candidate, is closed (as if the process died), and a
// second fresh pool (a new coordinator generation - exactly what a real
// resumed pipeline-orchestrate run would start) completes the remaining
// candidate - proving a resumed remote run reproduces the same per-
// candidate values an uninterrupted run would have. Candidates (not
// permutation runs) are the resume boundary here because a whole candidate
// is this stage's atomic unit of distributed work.
func TestHigherOrderRemoteInterruptedThenResumedMatchesOracle(t *testing.T) {
	c := higherOrderRemoteFixture(t)
	sequences := []string{"aiin chey shey", "chey shey ol"}
	want := make(map[string]higherorderseq.CandidateResult, len(sequences))
	for _, seq := range sequences {
		want[seq] = higherOrderOracleReplicate(t, c, seq)
	}

	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt

	runOne := func(seq, workerID string) higherorderseq.CandidateResult {
		ex, err := NewHigherOrderRemoteExecutor(c)
		if err != nil {
			t.Fatal(err)
		}
		pool := ex.(*higherOrderExecutorAdapter).pool.(*remotePool)
		ctx, cancel := context.WithCancel(context.Background())
		crt, key := pkiFix.issueWorker(t, workerID)
		startRemoteWorker(t, ctx, "https://localhost:"+strings.Split(pool.Addr(), ":")[1], pkiFix.caCrt, crt, key, 2)
		r, err := ex.Run(context.Background(), seq)
		if err != nil {
			t.Fatal(err)
		}
		cancel()
		_ = ex.(*higherOrderExecutorAdapter).pool.Close()
		return r
	}

	got := map[string]higherorderseq.CandidateResult{
		sequences[0]: runOne(sequences[0], "higher-order-resume-worker-1"),
		sequences[1]: runOne(sequences[1], "higher-order-resume-worker-2"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("interrupted-then-resumed remote run differs from an uninterrupted oracle\ngot=%#v\nwant=%#v", got, want)
	}
}
