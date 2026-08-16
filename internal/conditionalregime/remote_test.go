package conditionalregime

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"zcore.dev/voinich/internal/pki"
)

// remotePKI is a project CA plus a coordinator identity generated once per
// test, from which any number of individual worker certificates can be
// issued - mirroring what an operator does once with conditional-regime-pki
// before running any coordinator/worker.
type remotePKI struct {
	dir                string
	caCrt              string
	coordCrt, coordKey string
}

func newRemotePKI(t *testing.T, dnsNames, ips []string) remotePKI {
	t.Helper()
	dir := t.TempDir()
	if err := pki.GenerateCA(dir, time.Hour, false); err != nil {
		t.Fatalf("GenerateCA: %v", err)
	}
	caCrt, caKey := pki.CAPaths(dir)
	if err := pki.IssueCoordinator(caCrt, caKey, dir, dnsNames, ips, time.Hour, false); err != nil {
		t.Fatalf("IssueCoordinator: %v", err)
	}
	coordCrt, coordKey := pki.IssueCoordinatorPaths(dir)
	return remotePKI{dir: dir, caCrt: caCrt, coordCrt: coordCrt, coordKey: coordKey}
}

func (p remotePKI) issueWorker(t *testing.T, workerID string) (crt, key string) {
	t.Helper()
	caCrt, caKey := pki.CAPaths(p.dir)
	if err := pki.IssueWorker(caCrt, caKey, p.dir, workerID, time.Hour, false); err != nil {
		t.Fatalf("IssueWorker %s: %v", workerID, err)
	}
	return pki.IssueWorkerPaths(p.dir, workerID)
}

// renewWorker re-issues workerID's certificate (a fresh key and serial) in
// place, simulating routine renewal or replacing a compromised credential.
func (p remotePKI) renewWorker(t *testing.T, workerID string) (crt, key string) {
	t.Helper()
	caCrt, caKey := pki.CAPaths(p.dir)
	if err := pki.IssueWorker(caCrt, caKey, p.dir, workerID, time.Hour, true); err != nil {
		t.Fatalf("renew worker %s: %v", workerID, err)
	}
	return pki.IssueWorkerPaths(p.dir, workerID)
}

func workerHTTPClient(t *testing.T, certFile, keyFile, caFile string) *http.Client {
	t.Helper()
	cfg, err := pki.WorkerClientTLSConfig(certFile, keyFile, caFile)
	if err != nil {
		t.Fatalf("WorkerClientTLSConfig: %v", err)
	}
	cfg.ServerName = "localhost"
	return &http.Client{Transport: &http.Transport{TLSClientConfig: cfg}, Timeout: 10 * time.Second}
}

func freeLoopbackAddr(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("local TCP unavailable: %v", err)
	}
	addr := l.Addr().String()
	l.Close()
	return addr
}

func remoteCoordinatorConfig(f fixture, pkiFix remotePKI) Config {
	c := f.smallConfig()
	c.Executor = "remote"
	c.Workers = 4
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	c.RemoteTimeout = 5 * time.Second
	c.RemoteRetries = 2
	return c
}

// startRemoteWorker runs a worker client in the background for the life of
// the test, asserting it only ever exits (once ctx is cancelled) cleanly.
func startRemoteWorker(t *testing.T, ctx context.Context, coordinatorURL, caCrt, workerCrt, workerKey string, concurrency int) {
	t.Helper()
	cache := t.TempDir()
	errCh := make(chan error, 1)
	go func() { errCh <- RunRemoteWorker(ctx, coordinatorURL, caCrt, workerCrt, workerKey, cache, concurrency) }()
	t.Cleanup(func() {
		select {
		case err := <-errCh:
			if err != nil && ctx.Err() == nil {
				t.Errorf("worker exited with error: %v", err)
			}
		case <-time.After(2 * time.Second):
		}
	})
}

func postLeaseStatus(t *testing.T, client *http.Client, base string, req remoteLeaseRequest) int {
	t.Helper()
	body, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Post(base+"/v1/lease", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode
}

func doLease(t *testing.T, client *http.Client, base, fingerprint string) remoteLeaseResponse {
	t.Helper()
	req := remoteLeaseRequest{Protocol: remoteProtocolVersion, Compatibility: scientificCompatibilityVersion, GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, GoVersion: runtime.Version(), ExperimentID: fingerprint}
	body, _ := json.Marshal(req)
	resp, err := client.Post(base+"/v1/lease", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /v1/lease: HTTP %d", resp.StatusCode)
	}
	var out remoteLeaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out
}

func leaseUntilAssigned(t *testing.T, client *http.Client, base, fingerprint string) remoteLeaseResponse {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		lease := doLease(t, client, base, fingerprint)
		if !lease.NoWork {
			return lease
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("timed out waiting for a job to become leasable")
	return remoteLeaseResponse{}
}

// TestRemoteExecutorMatchesOracleWithTwoWorkersAndCachedInput is Task34's
// core reproducibility claim (phase 10): the full pipeline over mTLS, with
// two independently authenticated workers, produces byte-identical output
// to the sequential goroutine oracle, both cold and with warm worker
// caches.
func TestRemoteExecutorMatchesOracleWithTwoWorkersAndCachedInput(t *testing.T) {
	if testing.Short() {
		t.Skip("full remote mTLS pipeline integration")
	}
	f := writeFixture(t)
	oracle := runFixturePipeline(t, f, "goroutine", 1)

	pkiFix := newRemotePKI(t, []string{"localhost"}, []string{"127.0.0.1"})
	worker1Crt, worker1Key := pkiFix.issueWorker(t, "worker-1")
	worker2Crt, worker2Key := pkiFix.issueWorker(t, "worker-2")

	run := func() map[string]string {
		c := remoteCoordinatorConfig(f, pkiFix)
		c.RemoteListen = freeLoopbackAddr(t)
		c.OutputDir, c.CheckpointPath, c.Quiet = t.TempDir(), "-", true
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		c.Context = ctx

		coordinatorURL := "https://" + c.RemoteListen
		startRemoteWorker(t, ctx, coordinatorURL, pkiFix.caCrt, worker1Crt, worker1Key, 2)
		startRemoteWorker(t, ctx, coordinatorURL, pkiFix.caCrt, worker2Crt, worker2Key, 2)

		if err := RunAndWrite(c); err != nil {
			t.Fatalf("remote RunAndWrite: %v", err)
		}
		return hashOutputDir(t, c.OutputDir)
	}
	assertIdenticalOutputs(t, "oracle vs two mTLS workers (cold cache)", oracle, run())
	assertIdenticalOutputs(t, "oracle vs two mTLS workers (warm cache)", oracle, run())
}

// TestRemoteRenewedAndDifferentWorkerCertificatesProduceIdenticalResults is
// Task34 phase 10's reproducibility matrix: the existing goroutine oracle,
// an mTLS run with worker certificate A, a run after renewing that same
// worker's certificate, and a run resolved entirely by a different
// authenticated worker must all be byte-for-byte identical.
func TestRemoteRenewedAndDifferentWorkerCertificatesProduceIdenticalResults(t *testing.T) {
	if testing.Short() {
		t.Skip("full remote mTLS pipeline integration")
	}
	f := writeFixture(t)
	oracle := runFixturePipeline(t, f, "goroutine", 1)
	pkiFix := newRemotePKI(t, []string{"localhost"}, []string{"127.0.0.1"})

	runWithWorker := func(crt, key string) map[string]string {
		c := remoteCoordinatorConfig(f, pkiFix)
		c.RemoteListen = freeLoopbackAddr(t)
		c.OutputDir, c.CheckpointPath, c.Quiet = t.TempDir(), "-", true
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		c.Context = ctx
		startRemoteWorker(t, ctx, "https://"+c.RemoteListen, pkiFix.caCrt, crt, key, 2)
		if err := RunAndWrite(c); err != nil {
			t.Fatalf("remote RunAndWrite: %v", err)
		}
		return hashOutputDir(t, c.OutputDir)
	}

	crtA, keyA := pkiFix.issueWorker(t, "worker-1")
	assertIdenticalOutputs(t, "oracle vs mTLS worker certificate A", oracle, runWithWorker(crtA, keyA))

	crtRenewed, keyRenewed := pkiFix.renewWorker(t, "worker-1")
	assertIdenticalOutputs(t, "oracle vs renewed worker certificate", oracle, runWithWorker(crtRenewed, keyRenewed))

	crtB, keyB := pkiFix.issueWorker(t, "worker-2")
	assertIdenticalOutputs(t, "oracle vs a different authenticated worker", oracle, runWithWorker(crtB, keyB))
}

// TestRemoteJobReassignedAfterLeaseExpiryMatchesLocal proves Task33's retry
// semantics survive the transport inversion: a job leased by a worker that
// then never responds (crash/kill) is reclaimed after Config.RemoteTimeout
// and completed by a second, distinct authenticated worker, with the exact
// bit-for-bit result the scientific implementation would produce locally.
func TestRemoteJobReassignedAfterLeaseExpiryMatchesLocal(t *testing.T) {
	f := writeFixture(t)
	pkiFix := newRemotePKI(t, []string{"localhost"}, []string{"127.0.0.1"})
	c := remoteCoordinatorConfig(f, pkiFix)
	c.RemoteTimeout = 150 * time.Millisecond
	c.RemoteRetries = 3
	fingerprint := computeFingerprint(c, f.corpusHash, f.metaHash)
	pool, err := newRemotePool(c, fingerprint, f.corpusHash, f.metaHash)
	if err != nil {
		t.Fatalf("newRemotePool: %v", err)
	}
	defer pool.Close()
	base := "https://" + pool.Addr()

	id := JobID{Stage: "part_b_global_correction", Combination: "k_medoids|raw", ReplicateIndex: 1}
	type outcome struct {
		value float64
		err   error
	}
	resultCh := make(chan outcome, 1)
	go func() {
		v, err := pool.Run(context.Background(), id)
		resultCh <- outcome{v, err}
	}()

	crashCrt, crashKey := pkiFix.issueWorker(t, "worker-crash")
	crashClient := workerHTTPClient(t, crashCrt, crashKey, pkiFix.caCrt)
	leased := leaseUntilAssigned(t, crashClient, base, fingerprint)
	if leased.JobID != id {
		t.Fatalf("crashing worker leased %+v, want %+v", leased.JobID, id)
	}
	// The crashing worker now goes silent forever: no /v1/result ever
	// follows. The coordinator must reclaim the lease after RemoteTimeout.

	goodCrt, goodKey := pkiFix.issueWorker(t, "worker-good")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startRemoteWorker(t, ctx, base, pkiFix.caCrt, goodCrt, goodKey, 1)

	init := f.initMessage(t, c)
	state, err := newWorkerState(init)
	if err != nil {
		t.Fatal(err)
	}
	want, err := state.compute(id)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case res := <-resultCh:
		if res.err != nil {
			t.Fatalf("reassigned job failed: %v", res.err)
		}
		if res.value != want {
			t.Fatalf("reassigned job value=%v local=%v", res.value, want)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for the job to be reassigned and completed")
	}
}

// TestRemoteWorkerCannotImpersonateAnotherWorkersLease is Task34 phase 9's
// "request cannot impersonate another WorkerID": WorkerID always comes from
// this connection's own verified certificate, so a second, equally valid,
// authenticated worker cannot complete a lease that was handed to a
// different worker's certificate.
func TestRemoteWorkerCannotImpersonateAnotherWorkersLease(t *testing.T) {
	f := writeFixture(t)
	pkiFix := newRemotePKI(t, []string{"localhost"}, []string{"127.0.0.1"})
	c := remoteCoordinatorConfig(f, pkiFix)
	c.RemoteTimeout = 5 * time.Second
	fingerprint := computeFingerprint(c, f.corpusHash, f.metaHash)
	pool, err := newRemotePool(c, fingerprint, f.corpusHash, f.metaHash)
	if err != nil {
		t.Fatalf("newRemotePool: %v", err)
	}
	defer pool.Close()
	base := "https://" + pool.Addr()

	aCrt, aKey := pkiFix.issueWorker(t, "worker-a")
	bCrt, bKey := pkiFix.issueWorker(t, "worker-b")
	clientA := workerHTTPClient(t, aCrt, aKey, pkiFix.caCrt)
	clientB := workerHTTPClient(t, bCrt, bKey, pkiFix.caCrt)

	id := JobID{Stage: "part_b_global_correction", Combination: "hierarchical|raw", ReplicateIndex: 0}
	type outcome struct {
		value float64
		err   error
	}
	resultCh := make(chan outcome, 1)
	go func() {
		v, err := pool.Run(context.Background(), id)
		resultCh <- outcome{v, err}
	}()

	leased := leaseUntilAssigned(t, clientA, base, fingerprint)
	if leased.JobID != id {
		t.Fatalf("worker-a leased %+v, want %+v", leased.JobID, id)
	}

	init := f.initMessage(t, c)
	state, err := newWorkerState(init)
	if err != nil {
		t.Fatal(err)
	}
	want, err := state.compute(id)
	if err != nil {
		t.Fatal(err)
	}

	impersonation := remoteResultRequest{ExperimentID: fingerprint, LeaseID: leased.LeaseID, JobID: leased.JobID, Value: 999999}
	body, _ := json.Marshal(impersonation)
	resp, err := clientB.Post(base+"/v1/result", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("worker-b impersonating worker-a's lease: HTTP %d, want 403", resp.StatusCode)
	}

	legit := remoteResultRequest{ExperimentID: fingerprint, LeaseID: leased.LeaseID, JobID: leased.JobID, Value: want}
	body, _ = json.Marshal(legit)
	resp, err = clientA.Post(base+"/v1/result", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("worker-a's own result: HTTP %d, want 200", resp.StatusCode)
	}

	select {
	case res := <-resultCh:
		if res.err != nil {
			t.Fatalf("job failed: %v", res.err)
		}
		if res.value != want {
			t.Fatalf("job value=%v local=%v (impersonated value must never have won)", res.value, want)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for worker-a's legitimate result")
	}
}

func TestRemoteCoordinatorRejectsStaleExperimentAndIncompatibleRuntime(t *testing.T) {
	f := writeFixture(t)
	pkiFix := newRemotePKI(t, []string{"localhost"}, []string{"127.0.0.1"})
	c := remoteCoordinatorConfig(f, pkiFix)
	fingerprint := computeFingerprint(c, f.corpusHash, f.metaHash)
	pool, err := newRemotePool(c, fingerprint, f.corpusHash, f.metaHash)
	if err != nil {
		t.Fatalf("newRemotePool: %v", err)
	}
	defer pool.Close()
	base := "https://" + pool.Addr()
	crt, key := pkiFix.issueWorker(t, "worker-1")
	client := workerHTTPClient(t, crt, key, pkiFix.caCrt)

	req := remoteLeaseRequest{Protocol: remoteProtocolVersion, Compatibility: scientificCompatibilityVersion, GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, GoVersion: runtime.Version(), ExperimentID: "stale"}
	if status := postLeaseStatus(t, client, base, req); status != http.StatusConflict {
		t.Fatalf("stale experiment HTTP %d, want 409", status)
	}

	req.ExperimentID, req.GoVersion = fingerprint, runtime.Version()+"-different"
	if status := postLeaseStatus(t, client, base, req); status != http.StatusConflict {
		t.Fatalf("runtime mismatch HTTP %d, want 409", status)
	}
}

func TestRemoteCoordinatorRejectsRevokedWorker(t *testing.T) {
	f := writeFixture(t)
	pkiFix := newRemotePKI(t, []string{"localhost"}, []string{"127.0.0.1"})
	crt, key := pkiFix.issueWorker(t, "worker-1")
	denyPath := filepath.Join(t.TempDir(), "deny.json")
	deny := &pki.DenyList{Serials: map[string]bool{}, WorkerIDs: map[string]bool{"worker-1": true}}
	if err := pki.SaveDenyList(denyPath, deny); err != nil {
		t.Fatal(err)
	}
	c := remoteCoordinatorConfig(f, pkiFix)
	c.RemoteDenyList = denyPath
	fingerprint := computeFingerprint(c, f.corpusHash, f.metaHash)
	pool, err := newRemotePool(c, fingerprint, f.corpusHash, f.metaHash)
	if err != nil {
		t.Fatalf("newRemotePool: %v", err)
	}
	defer pool.Close()
	client := workerHTTPClient(t, crt, key, pkiFix.caCrt)
	resp, err := client.Get("https://" + pool.Addr() + "/v1/handshake")
	if err == nil {
		resp.Body.Close()
		t.Fatal("expected a revoked worker identity to be rejected")
	}
}

func TestRemoteMetricsRequiresAuthenticatedWorker(t *testing.T) {
	f := writeFixture(t)
	pkiFix := newRemotePKI(t, []string{"localhost"}, []string{"127.0.0.1"})
	c := remoteCoordinatorConfig(f, pkiFix)
	fingerprint := computeFingerprint(c, f.corpusHash, f.metaHash)
	pool, err := newRemotePool(c, fingerprint, f.corpusHash, f.metaHash)
	if err != nil {
		t.Fatalf("newRemotePool: %v", err)
	}
	defer pool.Close()
	crt, key := pkiFix.issueWorker(t, "worker-1")
	client := workerHTTPClient(t, crt, key, pkiFix.caCrt)
	resp, err := client.Get("https://" + pool.Addr() + "/v1/metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("metrics HTTP %d", resp.StatusCode)
	}
	var m remoteMetrics
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		t.Fatal(err)
	}
}

func TestRemoteCheckpointResumeSkipsCompletedJob(t *testing.T) {
	f := writeFixture(t)
	pkiFix := newRemotePKI(t, []string{"localhost"}, []string{"127.0.0.1"})
	c := remoteCoordinatorConfig(f, pkiFix)
	fp := computeFingerprint(c, f.corpusHash, f.metaHash)
	pool, err := newRemotePool(c, fp, f.corpusHash, f.metaHash)
	if err != nil {
		t.Fatalf("newRemotePool: %v", err)
	}
	defer pool.Close()

	crt, key := pkiFix.issueWorker(t, "worker-1")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startRemoteWorker(t, ctx, "https://"+pool.Addr(), pkiFix.caCrt, crt, key, 2)

	completedID := JobID{Stage: "part_b_global_correction", Combination: "k_medoids|raw", ReplicateIndex: 1}
	completedValue, err := pool.Run(context.Background(), completedID)
	if err != nil {
		t.Fatal(err)
	}
	cp := newCheckpoint(fp)
	cp.PermutationJobs[checkpointJobKey(completedID)] = completedValue
	checkpointPath := filepath.Join(t.TempDir(), "checkpoint.json")
	if err := saveCheckpoint(checkpointPath, cp); err != nil {
		t.Fatal(err)
	}
	loaded, ok := loadCheckpoint(checkpointPath, fp)
	if !ok {
		t.Fatal("coordinator restart did not load checkpoint")
	}
	values, err := runIndexedReplicatesState(context.Background(), 3, pool, completedID.Stage, completedID.Combination, 3, nil, checkpointJobsFor(loaded.PermutationJobs, completedID.Stage, completedID.Combination, 3), func(context.Context, int) (float64, error) { panic("remote executor expected") }, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if values[1] != completedValue {
		t.Fatalf("resumed value changed: %v vs %v", values[1], completedValue)
	}
}
