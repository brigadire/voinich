package conditionalregime

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func newRemoteTestServer(t *testing.T, token string, concurrency int) (*httptest.Server, string) {
	t.Helper()
	cache := t.TempDir()
	w := &remoteWorkerServer{cacheDir: cache, token: token, host: "test-host", sem: make(chan struct{}, concurrency), states: map[string]*workerState{}}
	server := newHTTPTestServer(t, w.routes())
	t.Cleanup(server.Close)
	return server, cache
}

func newHTTPTestServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Skipf("local TCP unavailable: %v", err)
	}
	server := httptest.NewUnstartedServer(handler)
	server.Listener = listener
	server.Start()
	return server
}

func remoteConfig(f fixture, endpoints ...string) Config {
	c := f.smallConfig()
	c.Executor = "remote"
	c.Workers = 4
	c.RemoteWorkers = endpoints
	c.RemoteToken = "test-secret"
	c.RemoteTimeout = 30 * time.Second
	c.RemoteRetries = 2
	return c
}

func TestRemoteExecutorMatchesOracleWithTwoWorkersAndCachedInput(t *testing.T) {
	if testing.Short() {
		t.Skip("full remote pipeline integration")
	}
	f := writeFixture(t)
	oracle := runFixturePipeline(t, f, "goroutine", 1)
	s1, cache1 := newRemoteTestServer(t, "test-secret", 2)
	s2, cache2 := newRemoteTestServer(t, "test-secret", 3)

	run := func() map[string]string {
		c := remoteConfig(f, s1.URL, s2.URL)
		c.OutputDir, c.CheckpointPath, c.Quiet = t.TempDir(), "-", true
		if err := RunAndWrite(c); err != nil {
			t.Fatalf("remote RunAndWrite: %v", err)
		}
		return hashOutputDir(t, c.OutputDir)
	}
	assertIdenticalOutputs(t, "oracle vs two remote workers", oracle, run())
	for _, cache := range []string{cache1, cache2} {
		entries, err := os.ReadDir(cache)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 2 {
			t.Fatalf("cold cache %s contains %d objects, want corpus+metadata", cache, len(entries))
		}
	}
	assertIdenticalOutputs(t, "oracle vs cached remote run", oracle, run())
	for _, cache := range []string{cache1, cache2} {
		entries, _ := os.ReadDir(cache)
		if len(entries) != 2 {
			t.Fatalf("cached run restaged mutable objects: %d", len(entries))
		}
	}
}

func TestRemoteTransportRetryAndExactJobIdentity(t *testing.T) {
	f := writeFixture(t)
	var posts atomic.Int32
	w := &remoteWorkerServer{cacheDir: t.TempDir(), token: "test-secret", host: "retry-host", sem: make(chan struct{}, 1), states: map[string]*workerState{}}
	real := w.routes()
	retryServer := newHTTPTestServer(t, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/v1/job" && posts.Add(1) == 1 {
			http.Error(rw, "temporary", 503)
			return
		}
		real.ServeHTTP(rw, r)
	}))
	defer retryServer.Close()
	c := remoteConfig(f, retryServer.URL)
	fingerprint := computeFingerprint(c, f.corpusHash, f.metaHash)
	p, err := newRemotePool(c, fingerprint, f.corpusHash, f.metaHash)
	if err != nil {
		t.Fatal(err)
	}
	id := JobID{Stage: "part_b_global_correction", Combination: "k_medoids|raw", ReplicateIndex: 1}
	got, err := p.Run(context.Background(), id)
	if err != nil {
		t.Fatal(err)
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
	if got != want {
		t.Fatalf("retried remote result=%v local=%v", got, want)
	}
	duplicate, err := p.Run(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if duplicate != got {
		t.Fatalf("duplicate delivery changed result: first=%v duplicate=%v", got, duplicate)
	}
}

func TestRemoteWorkerRejectsStaleExperimentAndRuntime(t *testing.T) {
	f := writeFixture(t)
	s, _ := newRemoteTestServer(t, "test-secret", 1)
	c := remoteConfig(f, s.URL)
	fp := computeFingerprint(c, f.corpusHash, f.metaHash)
	p, err := newRemotePool(c, fp, f.corpusHash, f.metaHash)
	if err != nil {
		t.Fatal(err)
	}
	req := p.req
	req.ExperimentID = "stale"
	req.JobID = JobID{Stage: "part_b_global_correction", Combination: "k_medoids|raw", ReplicateIndex: 0}
	b, _ := json.Marshal(req)
	resp, err := p.request(context.Background(), "POST", s.URL+"/v1/job", b)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("stale experiment HTTP %d, want 409", resp.StatusCode)
	}
	req.ExperimentID, req.GoVersion = fp, runtime.Version()+"-different"
	b, _ = json.Marshal(req)
	resp, err = p.request(context.Background(), "POST", s.URL+"/v1/job", b)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("runtime mismatch HTTP %d, want 409", resp.StatusCode)
	}
}

func TestRemoteWorkerMetricsAndNonLoopbackAuthentication(t *testing.T) {
	if err := RunRemoteWorker(context.Background(), "0.0.0.0:0", t.TempDir(), "", 1); err == nil {
		t.Fatal("unauthenticated non-loopback worker must be rejected")
	}
	s, _ := newRemoteTestServer(t, "test-secret", 1)
	req, err := http.NewRequest("GET", s.URL+"/v1/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer test-secret")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("metrics HTTP %d", resp.StatusCode)
	}
	var metrics remoteMetrics
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "linux" && (metrics.RSSBytes == 0 || metrics.PeakRSSBytes == 0) {
		t.Fatalf("metrics omitted process RSS: %+v", metrics)
	}
}

func TestRemoteCheckpointResumeSkipsCompletedJob(t *testing.T) {
	f := writeFixture(t)
	s, _ := newRemoteTestServer(t, "test-secret", 2)
	c := remoteConfig(f, s.URL)
	fp := computeFingerprint(c, f.corpusHash, f.metaHash)
	p, err := newRemotePool(c, fp, f.corpusHash, f.metaHash)
	if err != nil {
		t.Fatal(err)
	}
	completedID := JobID{Stage: "part_b_global_correction", Combination: "k_medoids|raw", ReplicateIndex: 1}
	completedValue, err := p.Run(context.Background(), completedID)
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
	values, err := runIndexedReplicatesState(context.Background(), 3, p, completedID.Stage, completedID.Combination, 3, nil, checkpointJobsFor(loaded.PermutationJobs, completedID.Stage, completedID.Combination, 3), func(context.Context, int) (float64, error) { panic("remote executor expected") }, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if values[1] != completedValue {
		t.Fatalf("resumed value changed: %v vs %v", values[1], completedValue)
	}
}
