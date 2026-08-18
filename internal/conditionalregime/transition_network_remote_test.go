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

	"zcore.dev/voinich/internal/transitionnetwork"
)

// transitionNetworkRemoteFixture writes a small generic corpus and returns
// the transitionnetwork.Config every test in this file starts from.
func transitionNetworkRemoteFixture(t *testing.T) transitionnetwork.Config {
	t.Helper()
	d := t.TempDir()
	corpus := strings.Repeat("aiin chey shey ol or dy qokeey s\n", 40)
	corpusPath := filepath.Join(d, "corpus.txt")
	if err := os.WriteFile(corpusPath, []byte(corpus), 0600); err != nil {
		t.Fatal(err)
	}
	return transitionnetwork.Config{
		CorpusPath: corpusPath, Generic: true,
		MinTokenCount: 1, MinBlockTokenCount: 1,
		Permutations: 6, RefinePermutations: 6, Seed: 1,
		Workers: 1, RemoteTimeout: 5 * time.Second, RemoteRetries: 2,
	}
}

func transitionNetworkOracleReplicate(t *testing.T, c transitionnetwork.Config, phase string, rep int) transitionnetwork.ReplicateResult {
	t.Helper()
	ws, _, _, err := transitionnetwork.LoadForDistribution(c)
	if err != nil {
		t.Fatal(err)
	}
	got, err := transitionnetwork.ComputeReplicate(ws, c.Seed, phase, rep)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestTransitionNetworkRemoteMatchesLocalOracle(t *testing.T) {
	c := transitionNetworkRemoteFixture(t)
	want := transitionNetworkOracleReplicate(t, c, "primary", 0)

	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	workerCrt, workerKey := pkiFix.issueWorker(t, "transition-network-worker")
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	ex, err := NewTransitionNetworkRemoteExecutor(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ex.(*transitionNetworkExecutorAdapter).pool.Close()
	pool := ex.(*transitionNetworkExecutorAdapter).pool.(*remotePool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startRemoteWorker(t, ctx, "https://localhost:"+strings.Split(pool.Addr(), ":")[1], pkiFix.caCrt, workerCrt, workerKey, 4)

	got, err := ex.Run(context.Background(), "primary", 0)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("remote replicate differs\ngot=%#v\nwant=%#v", got, want)
	}
}

func TestTransitionNetworkTwoRemoteWorkersMatchLocalInAnyCompletionOrder(t *testing.T) {
	c := transitionNetworkRemoteFixture(t)
	n := c.Permutations
	want := make([]transitionnetwork.ReplicateResult, n)
	for i := range want {
		want[i] = transitionNetworkOracleReplicate(t, c, "primary", i)
	}

	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	ex, err := NewTransitionNetworkRemoteExecutor(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ex.(*transitionNetworkExecutorAdapter).pool.Close()
	pool := ex.(*transitionNetworkExecutorAdapter).pool.(*remotePool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	url := "https://localhost:" + strings.Split(pool.Addr(), ":")[1]
	for _, name := range []string{"transition-network-worker-a", "transition-network-worker-b"} {
		crt, key := pkiFix.issueWorker(t, name)
		startRemoteWorker(t, ctx, url, pkiFix.caCrt, crt, key, 4)
	}

	got := make([]transitionnetwork.ReplicateResult, n)
	errCh := make(chan error, n)
	var wg sync.WaitGroup
	for i := range got {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var runErr error
			got[i], runErr = ex.Run(context.Background(), "primary", i)
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

func TestTransitionNetworkRemoteRetryAfterWorkerFailure(t *testing.T) {
	c := transitionNetworkRemoteFixture(t)
	c.RemoteTimeout = 150 * time.Millisecond
	want := transitionNetworkOracleReplicate(t, c, "primary", 1)

	pkiFix := newRemotePKI(t, []string{"localhost"}, []string{"127.0.0.1"})
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	fp, err := transitionnetwork.Fingerprint(c)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := newTransitionNetworkRemotePool(c, fp)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	base := "https://" + pool.Addr()

	id := JobID{Stage: "transition_network_permutation", Combination: "primary", ReplicateIndex: 1}
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

	crashCrt, crashKey := pkiFix.issueWorker(t, "transition-network-worker-crash")
	crashClient := workerHTTPClient(t, crashCrt, crashKey, pkiFix.caCrt)
	leased := leaseUntilAssigned(t, crashClient, base, fp)
	if leased.JobID != id {
		t.Fatalf("crashing worker leased %+v, want %+v", leased.JobID, id)
	}

	goodCrt, goodKey := pkiFix.issueWorker(t, "transition-network-worker-good")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startRemoteWorker(t, ctx, base, pkiFix.caCrt, goodCrt, goodKey, 1)

	select {
	case res := <-resultCh:
		if res.err != nil {
			t.Fatalf("reassigned job failed: %v", res.err)
		}
		var got transitionnetwork.ReplicateResult
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

func TestTransitionNetworkRemoteSameSeedRepeated(t *testing.T) {
	c := transitionNetworkRemoteFixture(t)
	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	workerCrt, workerKey := pkiFix.issueWorker(t, "transition-network-worker")
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	ex, err := NewTransitionNetworkRemoteExecutor(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ex.(*transitionNetworkExecutorAdapter).pool.Close()
	pool := ex.(*transitionNetworkExecutorAdapter).pool.(*remotePool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startRemoteWorker(t, ctx, "https://localhost:"+strings.Split(pool.Addr(), ":")[1], pkiFix.caCrt, workerCrt, workerKey, 4)

	first, err := ex.Run(context.Background(), "primary", 2)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ex.Run(context.Background(), "primary", 2)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("repeated same-seed remote run diverged\nfirst=%#v\nsecond=%#v", first, second)
	}
}

func TestTransitionNetworkRemoteInterruptedThenResumedMatchesOracle(t *testing.T) {
	c := transitionNetworkRemoteFixture(t)
	n := c.Permutations
	half := n / 2
	want := make([]transitionnetwork.ReplicateResult, n)
	for i := range want {
		want[i] = transitionNetworkOracleReplicate(t, c, "primary", i)
	}

	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt

	runRange := func(lo, hi int, workerID string) []transitionnetwork.ReplicateResult {
		ex, err := NewTransitionNetworkRemoteExecutor(c)
		if err != nil {
			t.Fatal(err)
		}
		pool := ex.(*transitionNetworkExecutorAdapter).pool.(*remotePool)
		ctx, cancel := context.WithCancel(context.Background())
		crt, key := pkiFix.issueWorker(t, workerID)
		startRemoteWorker(t, ctx, "https://localhost:"+strings.Split(pool.Addr(), ":")[1], pkiFix.caCrt, crt, key, 2)
		out := make([]transitionnetwork.ReplicateResult, hi-lo)
		for i := lo; i < hi; i++ {
			r, err := ex.Run(context.Background(), "primary", i)
			if err != nil {
				t.Fatal(err)
			}
			out[i-lo] = r
		}
		cancel()
		_ = ex.(*transitionNetworkExecutorAdapter).pool.Close()
		return out
	}

	firstHalf := runRange(0, half, "transition-network-resume-worker-1")
	secondHalf := runRange(half, n, "transition-network-resume-worker-2")
	got := append(append([]transitionnetwork.ReplicateResult{}, firstHalf...), secondHalf...)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("interrupted-then-resumed remote run differs from an uninterrupted oracle\ngot=%#v\nwant=%#v", got, want)
	}
}
