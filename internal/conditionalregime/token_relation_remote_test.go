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

	"zcore.dev/voinich/internal/tokenrelationvalidation"
)

// tokenRelationRemoteFixture writes a small generic corpus plus the
// minimal discovery-dir files loadCandidates needs (one eligible
// "directional" candidate, so the direction battery has real work to do)
// and returns the tokenrelationvalidation.Config every test in this file
// starts from.
func tokenRelationRemoteFixture(t *testing.T) tokenrelationvalidation.Config {
	t.Helper()
	d := t.TempDir()
	write := func(name, body string) string {
		p := filepath.Join(d, name)
		if err := os.WriteFile(p, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	corpus := strings.Repeat("aiin chey shey ol or dy qokeey s\n", 40)
	corpusPath := write("corpus.txt", corpus)
	write("begin_end_candidates.yaml", "meta:\n  token_occurrences: 320\nparameters:\n  max_window: 3\ncandidates:\n  - begin_candidate: aiin\n    end_candidate: chey\n")
	write("distance_context_pairs.yaml", "token_count: 320\nparameters:\n  max_distance: 3\npairs: []\n")
	write("sequence_analysis.yaml", "meta:\n  token_occurrences: 320\nrepeated_ngrams: {}\n")
	write("structural_reliability.yaml", "meta:\n  token_occurrences: 320\nparameters:\n  threshold: 0.7\nreference_pairs: []\n")
	write("structural_classes.yaml", "models: []\n")
	write("soft_structural_space.yaml", "parameters:\n  graph_min_similarity: 0.5\n  min_token_count: 3\n")
	write("soft_structural_pairs.tsv", "token_a\ttoken_b\traw_similarity\n")
	return tokenrelationvalidation.Config{
		CorpusPath: corpusPath, DiscoveryDir: d, Generic: true,
		Permutations: 6, RefinePermutations: 6, Seed: 1,
		Workers: 1, RemoteTimeout: 5 * time.Second, RemoteRetries: 2,
	}
}

func tokenRelationOracleReplicate(t *testing.T, c tokenrelationvalidation.Config, family string, run int) map[string]float64 {
	t.Helper()
	blocks, candidates, maxD, err := tokenrelationvalidation.LoadForDistribution(c)
	if err != nil {
		t.Fatal(err)
	}
	got, err := tokenrelationvalidation.ComputeReplicate(blocks, candidates, maxD, c.Seed, family, run)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestTokenRelationRemoteMatchesLocalOracle(t *testing.T) {
	c := tokenRelationRemoteFixture(t)
	want := tokenRelationOracleReplicate(t, c, "direction", 0)

	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	workerCrt, workerKey := pkiFix.issueWorker(t, "token-relation-worker")
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	ex, err := NewTokenRelationRemoteExecutor(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ex.(*tokenRelationExecutorAdapter).pool.Close()
	pool := ex.(*tokenRelationExecutorAdapter).pool.(*remotePool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startRemoteWorker(t, ctx, "https://localhost:"+strings.Split(pool.Addr(), ":")[1], pkiFix.caCrt, workerCrt, workerKey, 4)

	got, err := ex.Run(context.Background(), "direction", 0)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("remote replicate differs\ngot=%#v\nwant=%#v", got, want)
	}
}

func TestTokenRelationTwoRemoteWorkersMatchLocalInAnyCompletionOrder(t *testing.T) {
	c := tokenRelationRemoteFixture(t)
	n := c.Permutations
	want := make([]map[string]float64, n)
	for i := range want {
		want[i] = tokenRelationOracleReplicate(t, c, "direction", i)
	}

	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	ex, err := NewTokenRelationRemoteExecutor(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ex.(*tokenRelationExecutorAdapter).pool.Close()
	pool := ex.(*tokenRelationExecutorAdapter).pool.(*remotePool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	url := "https://localhost:" + strings.Split(pool.Addr(), ":")[1]
	for _, name := range []string{"token-relation-worker-a", "token-relation-worker-b"} {
		crt, key := pkiFix.issueWorker(t, name)
		startRemoteWorker(t, ctx, url, pkiFix.caCrt, crt, key, 4)
	}

	got := make([]map[string]float64, n)
	errCh := make(chan error, n)
	var wg sync.WaitGroup
	for i := range got {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var runErr error
			got[i], runErr = ex.Run(context.Background(), "direction", i)
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

// TestTokenRelationRemoteRetryAfterWorkerFailure mirrors
// TestNormalizationRemoteRetryAfterWorkerFailure: a worker leases the job
// over raw HTTP and then goes silent forever, forcing the coordinator to
// reclaim the lease after RemoteTimeout and complete the job with a
// second, distinct authenticated worker - byte-identical to the local
// oracle.
func TestTokenRelationRemoteRetryAfterWorkerFailure(t *testing.T) {
	c := tokenRelationRemoteFixture(t)
	c.RemoteTimeout = 150 * time.Millisecond
	want := tokenRelationOracleReplicate(t, c, "direction", 1)

	pkiFix := newRemotePKI(t, []string{"localhost"}, []string{"127.0.0.1"})
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	fp, err := tokenrelationvalidation.Fingerprint(c)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := newTokenRelationRemotePool(c, fp)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	base := "https://" + pool.Addr()

	id := JobID{Stage: "token_relation_permutation", Combination: "direction", ReplicateIndex: 1}
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

	crashCrt, crashKey := pkiFix.issueWorker(t, "token-relation-worker-crash")
	crashClient := workerHTTPClient(t, crashCrt, crashKey, pkiFix.caCrt)
	leased := leaseUntilAssigned(t, crashClient, base, fp)
	if leased.JobID != id {
		t.Fatalf("crashing worker leased %+v, want %+v", leased.JobID, id)
	}
	// The crashing worker now goes silent forever: no /v1/result ever
	// follows. The coordinator must reclaim the lease after RemoteTimeout.

	goodCrt, goodKey := pkiFix.issueWorker(t, "token-relation-worker-good")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startRemoteWorker(t, ctx, base, pkiFix.caCrt, goodCrt, goodKey, 1)

	select {
	case res := <-resultCh:
		if res.err != nil {
			t.Fatalf("reassigned job failed: %v", res.err)
		}
		var got map[string]float64
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

// TestTokenRelationRemoteSameSeedRepeated confirms two independent remote
// runs of the identical (corpus, discovery inputs, seed, family, run)
// produce byte-identical results.
func TestTokenRelationRemoteSameSeedRepeated(t *testing.T) {
	c := tokenRelationRemoteFixture(t)
	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	workerCrt, workerKey := pkiFix.issueWorker(t, "token-relation-worker")
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	ex, err := NewTokenRelationRemoteExecutor(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ex.(*tokenRelationExecutorAdapter).pool.Close()
	pool := ex.(*tokenRelationExecutorAdapter).pool.(*remotePool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startRemoteWorker(t, ctx, "https://localhost:"+strings.Split(pool.Addr(), ":")[1], pkiFix.caCrt, workerCrt, workerKey, 4)

	first, err := ex.Run(context.Background(), "direction", 2)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ex.Run(context.Background(), "direction", 2)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("repeated same-seed remote run diverged\nfirst=%#v\nsecond=%#v", first, second)
	}
}

// TestTokenRelationRemoteInterruptedThenResumedMatchesOracle simulates a
// coordinator crash-and-restart mid-battery: a fresh pool completes only
// runs [0,half), is closed (as if the process died), and a second fresh
// pool (a new coordinator generation, exactly what a real resumed
// pipeline-orchestrate run would start) completes the remaining
// [half,n) - proving a resumed remote run reproduces the same per-run
// values a single uninterrupted run would have, regardless of where the
// resume boundary falls.
func TestTokenRelationRemoteInterruptedThenResumedMatchesOracle(t *testing.T) {
	c := tokenRelationRemoteFixture(t)
	n := c.Permutations
	half := n / 2
	want := make([]map[string]float64, n)
	for i := range want {
		want[i] = tokenRelationOracleReplicate(t, c, "direction", i)
	}

	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt

	runRange := func(lo, hi int, workerID string) []map[string]float64 {
		ex, err := NewTokenRelationRemoteExecutor(c)
		if err != nil {
			t.Fatal(err)
		}
		pool := ex.(*tokenRelationExecutorAdapter).pool.(*remotePool)
		ctx, cancel := context.WithCancel(context.Background())
		crt, key := pkiFix.issueWorker(t, workerID)
		startRemoteWorker(t, ctx, "https://localhost:"+strings.Split(pool.Addr(), ":")[1], pkiFix.caCrt, crt, key, 2)
		out := make([]map[string]float64, hi-lo)
		for i := lo; i < hi; i++ {
			r, err := ex.Run(context.Background(), "direction", i)
			if err != nil {
				t.Fatal(err)
			}
			out[i-lo] = r
		}
		cancel()
		_ = ex.(*tokenRelationExecutorAdapter).pool.Close()
		return out
	}

	firstHalf := runRange(0, half, "token-relation-resume-worker-1")
	secondHalf := runRange(half, n, "token-relation-resume-worker-2")
	got := append(append([]map[string]float64{}, firstHalf...), secondHalf...)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("interrupted-then-resumed remote run differs from an uninterrupted oracle\ngot=%#v\nwant=%#v", got, want)
	}
}
