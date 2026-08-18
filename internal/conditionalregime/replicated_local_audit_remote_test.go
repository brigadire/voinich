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

	"zcore.dev/voinich/internal/replicatedlocalaudit"
)

// replicatedLocalAuditRemoteFixture writes a small generic corpus plus the
// minimal, self-consistent relation-dir/discovery-dir files loadInputs
// needs: an empty distance-candidate table (that phase's math tolerates
// zero eligible candidates without error) and exactly one GROUP_CONSISTENT
// sequence candidate (loadInputs hard-errors on zero sequence candidates),
// present in both relation_classification.tsv and
// sequence_block_recurrence.tsv as loadInputs' cross-file check requires.
// Every other RelationDirFiles/DiscoveryDirFiles entry only needs to exist
// (fingerprint hashes its bytes; loadInputs never parses it), so those get
// a header-only placeholder.
func replicatedLocalAuditRemoteFixture(t *testing.T) replicatedlocalaudit.Config {
	t.Helper()
	relationDir := t.TempDir()
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

	writeIn(relationDir, "frozen_candidate_inventory.tsv", "candidate_id\n")
	writeIn(relationDir, "distance_profile_summary.tsv", "family\n")
	writeIn(relationDir, "distance_profile_block_validation.tsv", "family\n")
	writeIn(relationDir, "relation_controls.tsv", "candidate_id\n")
	writeIn(relationDir, "leave_one_block_out_transfer.tsv", "candidate_id\n")
	writeIn(relationDir, "metadata_transfer_matrix.tsv", "candidate_id\n")
	writeIn(relationDir, "token_relation_validation.yaml", "meta: {}\n")
	writeIn(relationDir, "relation_classification.tsv", "candidate_id\tfamily\tsequence\tclassification\nseq1\tsequence\taiin chey\tGROUP_CONSISTENT\n")
	writeIn(relationDir, "sequence_block_recurrence.tsv", "candidate_id\nseq1\n")
	writeIn(discoveryDir, "distance_context_pairs.yaml", "pairs: []\n")
	writeIn(discoveryDir, "sequence_analysis.yaml", "repeated_ngrams: {}\n")

	return replicatedlocalaudit.Config{
		CorpusPath: corpusPath, RelationDir: relationDir, DiscoveryDir: discoveryDir, Generic: true,
		Permutations: 6, Seed: 1,
		Workers: 1, RemoteTimeout: 5 * time.Second, RemoteRetries: 2,
	}
}

func replicatedLocalAuditOracleReplicate(t *testing.T, c replicatedlocalaudit.Config, phase string, run int) replicatedlocalaudit.ReplicateResult {
	t.Helper()
	state, _, err := replicatedlocalaudit.LoadForDistribution(c)
	if err != nil {
		t.Fatal(err)
	}
	got, err := replicatedlocalaudit.ComputeReplicate(state, c.Seed, phase, run)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestReplicatedLocalAuditRemoteMatchesLocalOracle(t *testing.T) {
	c := replicatedLocalAuditRemoteFixture(t)
	want := replicatedLocalAuditOracleReplicate(t, c, "shuffle", 0)

	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	workerCrt, workerKey := pkiFix.issueWorker(t, "replicated-local-audit-worker")
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	ex, err := NewReplicatedLocalAuditRemoteExecutor(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ex.(*replicatedLocalAuditExecutorAdapter).pool.Close()
	pool := ex.(*replicatedLocalAuditExecutorAdapter).pool.(*remotePool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startRemoteWorker(t, ctx, "https://localhost:"+strings.Split(pool.Addr(), ":")[1], pkiFix.caCrt, workerCrt, workerKey, 4)

	got, err := ex.Run(context.Background(), "shuffle", 0)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("remote replicate differs\ngot=%#v\nwant=%#v", got, want)
	}
}

func TestReplicatedLocalAuditTwoRemoteWorkersMatchLocalInAnyCompletionOrder(t *testing.T) {
	c := replicatedLocalAuditRemoteFixture(t)
	n := c.Permutations
	want := make([]replicatedlocalaudit.ReplicateResult, n)
	for i := range want {
		want[i] = replicatedLocalAuditOracleReplicate(t, c, "shuffle", i)
	}

	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	ex, err := NewReplicatedLocalAuditRemoteExecutor(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ex.(*replicatedLocalAuditExecutorAdapter).pool.Close()
	pool := ex.(*replicatedLocalAuditExecutorAdapter).pool.(*remotePool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	url := "https://localhost:" + strings.Split(pool.Addr(), ":")[1]
	for _, name := range []string{"replicated-local-audit-worker-a", "replicated-local-audit-worker-b"} {
		crt, key := pkiFix.issueWorker(t, name)
		startRemoteWorker(t, ctx, url, pkiFix.caCrt, crt, key, 4)
	}

	got := make([]replicatedlocalaudit.ReplicateResult, n)
	errCh := make(chan error, n)
	var wg sync.WaitGroup
	for i := range got {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var runErr error
			got[i], runErr = ex.Run(context.Background(), "shuffle", i)
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
		t.Fatalf("multi-worker remote replicates differ from local oracle in canonical order\ngot=%#v\nwant=%#v", got, want)
	}
}

// TestReplicatedLocalAuditRemoteRetryAfterWorkerFailure mirrors
// TestTokenRelationRemoteRetryAfterWorkerFailure: a worker leases the job
// over raw HTTP and then goes silent forever, forcing the coordinator to
// reclaim the lease after RemoteTimeout and complete the job with a
// second, distinct authenticated worker - byte-identical to the local
// oracle.
func TestReplicatedLocalAuditRemoteRetryAfterWorkerFailure(t *testing.T) {
	c := replicatedLocalAuditRemoteFixture(t)
	c.RemoteTimeout = 150 * time.Millisecond
	want := replicatedLocalAuditOracleReplicate(t, c, "shuffle", 1)

	pkiFix := newRemotePKI(t, []string{"localhost"}, []string{"127.0.0.1"})
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	fp, err := replicatedlocalaudit.Fingerprint(c)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := newReplicatedLocalAuditRemotePool(c, fp)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	base := "https://" + pool.Addr()

	id := JobID{Stage: "replicated_local_null", Combination: "shuffle", ReplicateIndex: 1}
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

	crashCrt, crashKey := pkiFix.issueWorker(t, "replicated-local-audit-worker-crash")
	crashClient := workerHTTPClient(t, crashCrt, crashKey, pkiFix.caCrt)
	leased := leaseUntilAssigned(t, crashClient, base, fp)
	if leased.JobID != id {
		t.Fatalf("crashing worker leased %+v, want %+v", leased.JobID, id)
	}
	// The crashing worker now goes silent forever: no /v1/result ever
	// follows. The coordinator must reclaim the lease after RemoteTimeout.

	goodCrt, goodKey := pkiFix.issueWorker(t, "replicated-local-audit-worker-good")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startRemoteWorker(t, ctx, base, pkiFix.caCrt, goodCrt, goodKey, 1)

	select {
	case res := <-resultCh:
		if res.err != nil {
			t.Fatalf("reassigned job failed: %v", res.err)
		}
		var got replicatedlocalaudit.ReplicateResult
		if err := json.Unmarshal(res.blob, &got); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("retried remote replicate differs from local oracle\ngot=%#v\nwant=%#v", got, want)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for the job to be reassigned and completed")
	}
	if pool.leasesReclaimed.Load() == 0 {
		t.Fatal("expected at least one reclaimed lease from the crashing worker")
	}
}

// TestReplicatedLocalAuditRemoteSameSeedRepeated confirms two independent
// remote runs of the identical (corpus, relation/discovery inputs, seed,
// phase, run) produce byte-identical results.
func TestReplicatedLocalAuditRemoteSameSeedRepeated(t *testing.T) {
	c := replicatedLocalAuditRemoteFixture(t)
	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	workerCrt, workerKey := pkiFix.issueWorker(t, "replicated-local-audit-worker")
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	ex, err := NewReplicatedLocalAuditRemoteExecutor(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ex.(*replicatedLocalAuditExecutorAdapter).pool.Close()
	pool := ex.(*replicatedLocalAuditExecutorAdapter).pool.(*remotePool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startRemoteWorker(t, ctx, "https://localhost:"+strings.Split(pool.Addr(), ":")[1], pkiFix.caCrt, workerCrt, workerKey, 4)

	first, err := ex.Run(context.Background(), "shuffle", 2)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ex.Run(context.Background(), "shuffle", 2)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("repeated same-seed remote run diverged\nfirst=%#v\nsecond=%#v", first, second)
	}
}

// TestReplicatedLocalAuditRemoteInterruptedThenResumedMatchesOracle
// simulates a coordinator crash-and-restart mid-battery: a fresh pool
// completes only runs [0,half), is closed (as if the process died), and a
// second fresh pool (a new coordinator generation, exactly what a real
// resumed pipeline-orchestrate run would start) completes the remaining
// [half,n) - proving a resumed remote run reproduces the same per-run
// values an uninterrupted run would have, regardless of where the resume
// boundary falls.
func TestReplicatedLocalAuditRemoteInterruptedThenResumedMatchesOracle(t *testing.T) {
	c := replicatedLocalAuditRemoteFixture(t)
	n := c.Permutations
	half := n / 2
	want := make([]replicatedlocalaudit.ReplicateResult, n)
	for i := range want {
		want[i] = replicatedLocalAuditOracleReplicate(t, c, "shuffle", i)
	}

	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt

	runRange := func(lo, hi int, workerID string) []replicatedlocalaudit.ReplicateResult {
		ex, err := NewReplicatedLocalAuditRemoteExecutor(c)
		if err != nil {
			t.Fatal(err)
		}
		pool := ex.(*replicatedLocalAuditExecutorAdapter).pool.(*remotePool)
		ctx, cancel := context.WithCancel(context.Background())
		crt, key := pkiFix.issueWorker(t, workerID)
		startRemoteWorker(t, ctx, "https://localhost:"+strings.Split(pool.Addr(), ":")[1], pkiFix.caCrt, crt, key, 2)
		out := make([]replicatedlocalaudit.ReplicateResult, hi-lo)
		for i := lo; i < hi; i++ {
			r, err := ex.Run(context.Background(), "shuffle", i)
			if err != nil {
				t.Fatal(err)
			}
			out[i-lo] = r
		}
		cancel()
		_ = ex.(*replicatedLocalAuditExecutorAdapter).pool.Close()
		return out
	}

	firstHalf := runRange(0, half, "replicated-local-audit-resume-worker-1")
	secondHalf := runRange(half, n, "replicated-local-audit-resume-worker-2")
	got := append(append([]replicatedlocalaudit.ReplicateResult{}, firstHalf...), secondHalf...)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("interrupted-then-resumed remote run differs from an uninterrupted oracle\ngot=%#v\nwant=%#v", got, want)
	}
}
