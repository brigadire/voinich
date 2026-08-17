package conditionalregime

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"zcore.dev/voinich/internal/normalizationcompare"
)

// namedNormalizationFixture is normalizationRemoteFixture but with caller-
// controlled corpus/threshold content, so two fixtures can represent two
// distinguishable experiments (different corpus -> different Fingerprint
// *and* different scientific output, not just a relabeled isomorphic copy -
// otherwise a contamination bug that served experiment A's corpus for an
// experiment B job would go undetected by comparing output alone).
func namedNormalizationFixture(t *testing.T, line string, reps int, tokenA, tokenB, label string) normalizationcompare.Config {
	t.Helper()
	d := t.TempDir()
	write := func(name, body string) string {
		p := filepath.Join(d, name)
		if err := os.WriteFile(p, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	corpus := strings.Repeat(line, reps)
	classes := `
meta:
  singleton_mode: token
  min_token_count: 1
  random_matching: "matched by frequency bin"
models:
  - threshold: 0.8
    label: "` + label + `"
    stats:
      multi_member_classes: 1
      classes: 2
    classes:
      - id: C0001
        size: 2
        members:
          - token: ` + tokenA + `
            count: 60
          - token: ` + tokenB + `
            count: 60
`
	return normalizationcompare.Config{
		InputPath: write("corpus.txt", corpus), ClassesPath: write("structural_classes.yaml", classes),
		RandomRuns: 3, RandomSeed: 1, Workers: 1, RemoteTimeout: 5 * time.Second, RemoteRetries: 2,
	}
}

// bindNormalizationCoordinator starts a normalization_compare_baseline
// coordinator bound to a specific, caller-chosen address (rather than :0),
// so a second coordinator can later bind the exact same address once the
// first is closed - simulating "the coordinator restarted" or "the
// coordinator moved on to a different experiment" from a persistent
// worker's point of view.
func bindNormalizationCoordinator(t *testing.T, addr string, c normalizationcompare.Config) *remotePool {
	t.Helper()
	classes, err := normalizationcompare.LoadClasses(c.ClassesPath)
	if err != nil {
		t.Fatal(err)
	}
	fp, err := normalizationcompare.Fingerprint(c.InputPath, c.ClassesPath, classes.Meta.MinTokenCount, classes.Meta.SingletonMode, c.RandomSeed, c.RandomRuns)
	if err != nil {
		t.Fatal(err)
	}
	c.RemoteListen = addr
	var pool *remotePool
	deadline := time.Now().Add(5 * time.Second)
	for {
		pool, err = newNormalizationRemotePool(c, classes, fp)
		if err == nil {
			return pool
		}
		if time.Now().After(deadline) {
			t.Fatalf("bind %s: %v", addr, err)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func normalizationOracle(t *testing.T, c normalizationcompare.Config, run int) normalizationcompare.BaselineResult {
	t.Helper()
	classes, err := normalizationcompare.LoadClasses(c.ClassesPath)
	if err != nil {
		t.Fatal(err)
	}
	corpus, err := loadNormalizationCorpusForTest(t, c.InputPath)
	if err != nil {
		t.Fatal(err)
	}
	want, err := normalizationcompare.RunRandomTrial(classes.Models[0], corpus, classes.Meta.MinTokenCount, classes.Meta.SingletonMode, c.RandomSeed, run, defaultSequenceParams())
	if err != nil {
		t.Fatal(err)
	}
	return want
}

func mustUnmarshal(t *testing.T, b []byte, v any) {
	t.Helper()
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatal(err)
	}
}

// TestPersistentWorkerReconnectsAfterCoordinatorRestartSameExperiment covers
// Task42 test 14: a coordinator that stops and restarts for the *same*
// experiment (same corpus/classes/fingerprint) never requires the worker to
// restart or rebuild its computer state - it just reconnects.
func TestPersistentWorkerReconnectsAfterCoordinatorRestartSameExperiment(t *testing.T) {
	c := normalizationRemoteFixture(t)
	addr := freeLoopbackAddr(t)
	pkiFix := newRemotePKI(t, []string{"localhost"}, []string{"127.0.0.1"})
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	workerCrt, workerKey := pkiFix.issueWorker(t, "persistent-worker")
	cacheDir := t.TempDir()

	pool1 := bindNormalizationCoordinator(t, addr, c)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	workerDone := make(chan error, 1)
	go func() {
		workerDone <- RunPersistentRemoteWorker(ctx, "https://"+addr, pkiFix.caCrt, workerCrt, workerKey, cacheDir, 1)
	}()

	id0 := JobID{Stage: "normalization_compare_baseline", Combination: "080", ReplicateIndex: 0}
	b, err := pool1.RunBlob(context.Background(), id0)
	if err != nil {
		t.Fatal(err)
	}
	var got0 normalizationcompare.BaselineResult
	mustUnmarshal(t, b, &got0)
	if want0 := normalizationOracle(t, c, 0); !reflect.DeepEqual(got0, want0) {
		t.Fatalf("pre-restart job differs from oracle\ngot=%#v\nwant=%#v", got0, want0)
	}
	if err := pool1.Close(); err != nil {
		t.Fatal(err)
	}

	pool2 := bindNormalizationCoordinator(t, addr, c)
	defer pool2.Close()
	id1 := JobID{Stage: "normalization_compare_baseline", Combination: "080", ReplicateIndex: 1}
	b, err = pool2.RunBlob(context.Background(), id1)
	if err != nil {
		t.Fatal(err)
	}
	var got1 normalizationcompare.BaselineResult
	mustUnmarshal(t, b, &got1)
	if want1 := normalizationOracle(t, c, 1); !reflect.DeepEqual(got1, want1) {
		t.Fatalf("post-restart job differs from oracle (worker did not reconnect correctly)\ngot=%#v\nwant=%#v", got1, want1)
	}

	cancel()
	select {
	case err := <-workerDone:
		if err != nil {
			t.Fatalf("worker did not shut down cleanly: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for persistent worker to stop after ctx cancellation")
	}
}

// TestPersistentWorkerHandlesSequentialExperimentsWithoutContamination
// covers Task42 tests 16/12/13: one worker deployment, started once, must
// serve two sequential experiments (a different coordinator process, a
// different corpus/classes.yaml, therefore a different Fingerprint) sharing
// the same on-disk cache directory - without ever answering an experiment B
// job using experiment A's corpus/classes.
func TestPersistentWorkerHandlesSequentialExperimentsWithoutContamination(t *testing.T) {
	experimentA := namedNormalizationFixture(t, "alpha beta c alpha c beta\n", 30, "alpha", "beta", "080")
	experimentB := namedNormalizationFixture(t, "gamma gamma delta c delta c gamma\n", 45, "gamma", "delta", "090")

	addr := freeLoopbackAddr(t)
	pkiFix := newRemotePKI(t, []string{"localhost"}, []string{"127.0.0.1"})
	experimentA.TLSCert, experimentA.TLSKey, experimentA.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	experimentB.TLSCert, experimentB.TLSKey, experimentB.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	workerCrt, workerKey := pkiFix.issueWorker(t, "persistent-worker")
	cacheDir := t.TempDir() // deliberately reused across both experiments

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	workerDone := make(chan error, 1)
	go func() {
		workerDone <- RunPersistentRemoteWorker(ctx, "https://"+addr, pkiFix.caCrt, workerCrt, workerKey, cacheDir, 1)
	}()

	poolA := bindNormalizationCoordinator(t, addr, experimentA)
	idA := JobID{Stage: "normalization_compare_baseline", Combination: "080", ReplicateIndex: 0}
	b, err := poolA.RunBlob(context.Background(), idA)
	if err != nil {
		t.Fatal(err)
	}
	var gotA normalizationcompare.BaselineResult
	mustUnmarshal(t, b, &gotA)
	if wantA := normalizationOracle(t, experimentA, 0); !reflect.DeepEqual(gotA, wantA) {
		t.Fatalf("experiment A job differs from oracle\ngot=%#v\nwant=%#v", gotA, wantA)
	}
	if err := poolA.Close(); err != nil {
		t.Fatal(err)
	}

	// Experiment B reuses the same address and the same worker cache
	// directory, but is a completely different corpus/classes.yaml -
	// different Fingerprint. The worker is still holding experiment A's
	// ExperimentID in its lease loop; the coordinator must reject that
	// (409) and force a fresh handshake before any experiment B job can be
	// leased at all.
	poolB := bindNormalizationCoordinator(t, addr, experimentB)
	defer poolB.Close()
	idB := JobID{Stage: "normalization_compare_baseline", Combination: "090", ReplicateIndex: 0}
	b, err = poolB.RunBlob(context.Background(), idB)
	if err != nil {
		t.Fatal(err)
	}
	var gotB normalizationcompare.BaselineResult
	mustUnmarshal(t, b, &gotB)
	wantB := normalizationOracle(t, experimentB, 0)
	if !reflect.DeepEqual(gotB, wantB) {
		t.Fatalf("experiment B job differs from its own oracle - possible cross-experiment contamination\ngot=%#v\nwant=%#v", gotB, wantB)
	}
	// The strongest possible contamination check: experiment B's result
	// must not equal what experiment A would have produced, proving the
	// worker actually rebuilt its computer state from B's corpus/classes
	// rather than continuing to serve A's.
	wantAFor0 := normalizationOracle(t, experimentA, 0)
	if reflect.DeepEqual(gotB.Metrics, wantAFor0.Metrics) {
		t.Fatal("experiment B result matches experiment A's oracle - cross-experiment contamination")
	}

	cancel()
	select {
	case err := <-workerDone:
		if err != nil {
			t.Fatalf("worker did not shut down cleanly: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for persistent worker to stop after ctx cancellation")
	}
}

// TestPersistentWorkerConnectsWhenCoordinatorStartsLate covers Task42 test
// 15: the worker is started first, with no coordinator listening yet, and
// must connect automatically once one appears - no restart required.
func TestPersistentWorkerConnectsWhenCoordinatorStartsLate(t *testing.T) {
	c := normalizationRemoteFixture(t)
	addr := freeLoopbackAddr(t)
	pkiFix := newRemotePKI(t, []string{"localhost"}, []string{"127.0.0.1"})
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	workerCrt, workerKey := pkiFix.issueWorker(t, "late-coordinator-worker")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	workerDone := make(chan error, 1)
	go func() {
		workerDone <- RunPersistentRemoteWorker(ctx, "https://"+addr, pkiFix.caCrt, workerCrt, workerKey, t.TempDir(), 1)
	}()

	// The worker has nothing to connect to yet; give it a moment to prove
	// it is blocked retrying rather than exiting.
	time.Sleep(300 * time.Millisecond)
	select {
	case err := <-workerDone:
		t.Fatalf("worker exited before any coordinator existed: %v", err)
	default:
	}

	pool := bindNormalizationCoordinator(t, addr, c)
	defer pool.Close()
	id := JobID{Stage: "normalization_compare_baseline", Combination: "080", ReplicateIndex: 0}
	b, err := pool.RunBlob(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	var got normalizationcompare.BaselineResult
	mustUnmarshal(t, b, &got)
	if want := normalizationOracle(t, c, 0); !reflect.DeepEqual(got, want) {
		t.Fatalf("late-starting coordinator job differs from oracle\ngot=%#v\nwant=%#v", got, want)
	}

	cancel()
	select {
	case err := <-workerDone:
		if err != nil {
			t.Fatalf("worker did not shut down cleanly: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for persistent worker to stop after ctx cancellation")
	}
}

// TestPersistentWorkerReconnectUsesBoundedBackoffNotTightLoop covers Task42
// section 10: with no coordinator ever available, the worker must not spin
// in a tight reconnect loop. A bare TCP listener that accepts and instantly
// closes every connection stands in for "something is at that address but
// it isn't a working coordinator" (a superset of "nothing is listening at
// all", which awaitTCPReachable already bounds identically) - counting its
// accepted connections is a black-box measure of how many times the worker
// actually dialed, with no test-only hook needed in production code.
func TestPersistentWorkerReconnectUsesBoundedBackoffNotTightLoop(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("local TCP unavailable: %v", err)
	}
	defer l.Close()
	var attempts int64
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			atomic.AddInt64(&attempts, 1)
			conn.Close()
		}
	}()

	pkiFix := newRemotePKI(t, []string{"localhost"}, []string{"127.0.0.1"})
	workerCrt, workerKey := pkiFix.issueWorker(t, "tightloop-worker")

	ctx, cancel := context.WithTimeout(context.Background(), 3500*time.Millisecond)
	defer cancel()
	_ = RunPersistentRemoteWorker(ctx, "https://"+l.Addr().String(), pkiFix.caCrt, workerCrt, workerKey, t.TempDir(), 1)

	got := atomic.LoadInt64(&attempts)
	if got == 0 {
		t.Fatal("test setup issue: worker never attempted to connect at all")
	}
	if got > 15 {
		t.Fatalf("reconnect loop made %d connection attempts in 3.5s; expected bounded backoff to keep this far below a tight loop's thousands", got)
	}
}

// TestPersistentWorkerPermanentAuthFailureStopsRetrying covers Task42
// section 17: an untrusted CA is a permanent identity failure, not a
// transient outage, and must not retry forever.
func TestPersistentWorkerPermanentAuthFailureStopsRetrying(t *testing.T) {
	c := normalizationRemoteFixture(t)
	addr := freeLoopbackAddr(t)
	pkiFix := newRemotePKI(t, []string{"localhost"}, []string{"127.0.0.1"})
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	pool := bindNormalizationCoordinator(t, addr, c)
	defer pool.Close()

	wrongCA := newRemotePKI(t, []string{"localhost"}, []string{"127.0.0.1"})
	workerCrt, workerKey := wrongCA.issueWorker(t, "impostor-worker")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	start := time.Now()
	err := RunPersistentRemoteWorker(ctx, "https://"+addr, wrongCA.caCrt, workerCrt, workerKey, t.TempDir(), 1)
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("expected a permanent TLS/identity error, got nil")
	}
	if ctx.Err() != nil {
		t.Fatalf("worker retried an untrusted-CA failure for the full 5s timeout instead of failing fast: %v", ctx.Err())
	}
	if elapsed > 2*time.Second {
		t.Fatalf("permanent auth failure took %s to surface; should fail fast without retrying", elapsed)
	}
}
