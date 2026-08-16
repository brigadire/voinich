package conditionalregime

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	remoteProtocolVersion          = 1
	scientificCompatibilityVersion = "conditionalregime-task33-v1"
	maxRemoteInputBytes            = 64 << 20
	maxRemoteMessageBytes          = 1 << 20
)

type remoteInfo struct {
	Protocol      int    `json:"protocol"`
	Compatibility string `json:"compatibility"`
	GOOS          string `json:"goos"`
	GOARCH        string `json:"goarch"`
	GoVersion     string `json:"go_version"`
	CPUModel      string `json:"cpu_model"`
	Host          string `json:"host"`
}

type remoteJobRequest struct {
	Protocol      int             `json:"protocol"`
	Compatibility string          `json:"compatibility"`
	GOOS          string          `json:"goos"`
	GOARCH        string          `json:"goarch"`
	GoVersion     string          `json:"go_version"`
	ExperimentID  string          `json:"experiment_id"`
	CorpusHash    string          `json:"corpus_hash"`
	MetadataHash  string          `json:"metadata_hash"`
	Config        protocolMessage `json:"config"`
	JobID         JobID           `json:"job_id"`
}

type remoteJobResponse struct {
	OK           bool    `json:"ok"`
	Error        string  `json:"error,omitempty"`
	ExperimentID string  `json:"experiment_id"`
	JobID        JobID   `json:"job_id"`
	Value        float64 `json:"value"`
	Host         string  `json:"host"`
}

type remoteWorkerServer struct {
	cacheDir, token, host string
	sem                   chan struct{}
	mu                    sync.Mutex
	states                map[string]*workerState
	inputBytes            atomic.Int64
	stagingNanos          atomic.Int64
	cacheHits             atomic.Int64
	cacheMisses           atomic.Int64
	jobRequests           atomic.Int64
	jobRequestBytes       atomic.Int64
	jobResponseBytes      atomic.Int64
	jobFailures           atomic.Int64
}

type remoteMetrics struct {
	InputBytes       int64  `json:"input_bytes"`
	StagingNanos     int64  `json:"staging_nanos"`
	CacheHits        int64  `json:"cache_hits"`
	CacheMisses      int64  `json:"cache_misses"`
	JobRequests      int64  `json:"job_requests"`
	JobRequestBytes  int64  `json:"job_request_bytes"`
	JobResponseBytes int64  `json:"job_response_bytes"`
	JobFailures      int64  `json:"job_failures"`
	ProcessCPUTicks  uint64 `json:"process_cpu_ticks"`
	RSSBytes         uint64 `json:"rss_bytes"`
	PeakRSSBytes     uint64 `json:"peak_rss_bytes"`
	HeapAllocBytes   uint64 `json:"heap_alloc_bytes"`
	HeapSysBytes     uint64 `json:"heap_sys_bytes"`
}

// RunRemoteWorker serves the trusted-machine Task33 protocol until ctx is
// cancelled. Authentication is optional to support loopback tests, but
// operators must use a token and a protected network for non-loopback use.
func RunRemoteWorker(ctx context.Context, listen, cacheDir, token string, concurrency int) error {
	if concurrency < 1 {
		return fmt.Errorf("remote worker concurrency must be positive")
	}
	if cacheDir == "" {
		return fmt.Errorf("remote worker cache directory is required")
	}
	if token == "" && !loopbackListenAddress(listen) {
		return fmt.Errorf("refusing unauthenticated non-loopback listener %q", listen)
	}
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return err
	}
	host, _ := os.Hostname()
	w := &remoteWorkerServer{cacheDir: cacheDir, token: token, host: host, sem: make(chan struct{}, concurrency), states: map[string]*workerState{}}
	srv := &http.Server{Addr: listen, Handler: w.routes(), ReadHeaderTimeout: 5 * time.Second}
	done := make(chan error, 1)
	go func() { done <- srv.ListenAndServe() }()
	select {
	case err := <-done:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		return nil
	}
}

func loopbackListenAddress(listen string) bool {
	host, _, err := net.SplitHostPort(listen)
	if err != nil {
		return false
	}
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	return ip != nil && ip.IsLoopback()
}

func (w *remoteWorkerServer) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/info", w.auth(w.info))
	mux.HandleFunc("GET /v1/metrics", w.auth(w.metrics))
	mux.HandleFunc("HEAD /v1/input/{hash}", w.auth(w.inputHead))
	mux.HandleFunc("PUT /v1/input/{hash}", w.auth(w.inputPut))
	mux.HandleFunc("POST /v1/job", w.auth(w.job))
	return mux
}

func (w *remoteWorkerServer) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if w.token != "" {
			got := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if len(got) != len(w.token) || subtle.ConstantTimeCompare([]byte(got), []byte(w.token)) != 1 {
				http.Error(rw, "unauthorized", http.StatusUnauthorized)
				return
			}
		}
		next(rw, r)
	}
}

func writeRemoteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (w *remoteWorkerServer) writeJobJSON(rw http.ResponseWriter, status int, v remoteJobResponse) {
	b, err := json.Marshal(v)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	b = append(b, '\n')
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	n, _ := rw.Write(b)
	w.jobResponseBytes.Add(int64(n))
}

func localRemoteInfo(host string) remoteInfo {
	return remoteInfo{Protocol: remoteProtocolVersion, Compatibility: scientificCompatibilityVersion, GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, GoVersion: runtime.Version(), CPUModel: cpuModel(), Host: host}
}

func cpuModel() string {
	b, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return "unknown"
	}
	for _, line := range strings.Split(string(b), "\n") {
		if key, value, ok := strings.Cut(line, ":"); ok && strings.TrimSpace(key) == "model name" {
			return strings.TrimSpace(value)
		}
	}
	return "unknown"
}

func (w *remoteWorkerServer) info(rw http.ResponseWriter, _ *http.Request) {
	writeRemoteJSON(rw, 200, localRemoteInfo(w.host))
}

func (w *remoteWorkerServer) metrics(rw http.ResponseWriter, _ *http.Request) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	rss, peak := processRSS()
	writeRemoteJSON(rw, 200, remoteMetrics{
		InputBytes: w.inputBytes.Load(), StagingNanos: w.stagingNanos.Load(), CacheHits: w.cacheHits.Load(), CacheMisses: w.cacheMisses.Load(),
		JobRequests: w.jobRequests.Load(), JobRequestBytes: w.jobRequestBytes.Load(), JobResponseBytes: w.jobResponseBytes.Load(), JobFailures: w.jobFailures.Load(),
		ProcessCPUTicks: processCPUTicks(), RSSBytes: rss, PeakRSSBytes: peak, HeapAllocBytes: mem.HeapAlloc, HeapSysBytes: mem.HeapSys,
	})
}

func processCPUTicks() uint64 {
	b, err := os.ReadFile("/proc/self/stat")
	if err != nil {
		return 0
	}
	end := strings.LastIndex(string(b), ") ")
	if end < 0 {
		return 0
	}
	fields := strings.Fields(string(b)[end+2:])
	if len(fields) < 13 {
		return 0
	}
	user, _ := strconv.ParseUint(fields[11], 10, 64)
	system, _ := strconv.ParseUint(fields[12], 10, 64)
	return user + system
}

func processRSS() (rss, peak uint64) {
	b, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return 0, 0
	}
	for _, line := range strings.Split(string(b), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		value, _ := strconv.ParseUint(fields[1], 10, 64)
		switch strings.TrimSuffix(fields[0], ":") {
		case "VmRSS":
			rss = value * 1024
		case "VmHWM":
			peak = value * 1024
		}
	}
	return rss, peak
}

func validSHA256(s string) bool {
	b, err := hex.DecodeString(s)
	return err == nil && len(b) == sha256.Size && s == strings.ToLower(s)
}
func (w *remoteWorkerServer) inputPath(hash string) string { return filepath.Join(w.cacheDir, hash) }

func (w *remoteWorkerServer) inputHead(rw http.ResponseWriter, r *http.Request) {
	h := r.PathValue("hash")
	if !validSHA256(h) {
		http.Error(rw, "invalid sha256", 400)
		return
	}
	if _, err := os.Stat(w.inputPath(h)); err != nil {
		w.cacheMisses.Add(1)
		http.Error(rw, "not found", 404)
		return
	}
	w.cacheHits.Add(1)
	rw.WriteHeader(204)
}

func (w *remoteWorkerServer) inputPut(rw http.ResponseWriter, r *http.Request) {
	started := time.Now()
	h := r.PathValue("hash")
	if !validSHA256(h) {
		http.Error(rw, "invalid sha256", 400)
		return
	}
	b, err := io.ReadAll(io.LimitReader(r.Body, maxRemoteInputBytes+1))
	if err != nil || len(b) > maxRemoteInputBytes {
		http.Error(rw, "input exceeds limit", 413)
		return
	}
	actual := fmt.Sprintf("%x", sha256.Sum256(b))
	if actual != h {
		http.Error(rw, "input sha256 mismatch", 409)
		return
	}
	path := w.inputPath(h)
	if _, err := os.Stat(path); err == nil {
		rw.WriteHeader(204)
		return
	}
	tmp, err := os.CreateTemp(w.cacheDir, ".stage-")
	if err != nil {
		http.Error(rw, err.Error(), 500)
		return
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err = tmp.Chmod(0600); err == nil {
		_, err = tmp.Write(b)
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(tmpName, path)
	}
	if err != nil {
		http.Error(rw, err.Error(), 500)
		return
	}
	w.inputBytes.Add(int64(len(b)))
	w.stagingNanos.Add(time.Since(started).Nanoseconds())
	rw.WriteHeader(201)
}

func (w *remoteWorkerServer) job(rw http.ResponseWriter, r *http.Request) {
	select {
	case w.sem <- struct{}{}:
		defer func() { <-w.sem }()
	case <-r.Context().Done():
		return
	}
	b, readErr := io.ReadAll(io.LimitReader(r.Body, maxRemoteMessageBytes+1))
	if readErr != nil || len(b) > maxRemoteMessageBytes {
		http.Error(rw, "job message exceeds limit", http.StatusRequestEntityTooLarge)
		return
	}
	w.jobRequests.Add(1)
	w.jobRequestBytes.Add(int64(len(b)))
	var req remoteJobRequest
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(rw, "malformed job: "+err.Error(), 400)
		return
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		http.Error(rw, "malformed job: trailing JSON value", 400)
		return
	}
	resp := remoteJobResponse{ExperimentID: req.ExperimentID, JobID: req.JobID, Host: w.host}
	fail := func(status int, err error) {
		w.jobFailures.Add(1)
		resp.Error = err.Error()
		w.writeJobJSON(rw, status, resp)
	}
	if req.Protocol != remoteProtocolVersion || req.Compatibility != scientificCompatibilityVersion {
		fail(409, fmt.Errorf("protocol/code compatibility mismatch"))
		return
	}
	if req.GOOS != runtime.GOOS || req.GOARCH != runtime.GOARCH || req.GoVersion != runtime.Version() {
		fail(409, fmt.Errorf("runtime compatibility mismatch: worker=%s/%s/%s coordinator=%s/%s/%s", runtime.GOOS, runtime.GOARCH, runtime.Version(), req.GOOS, req.GOARCH, req.GoVersion))
		return
	}
	if req.ExperimentID == "" || req.ExperimentID != req.Config.Fingerprint {
		fail(409, fmt.Errorf("stale or inconsistent experiment identity"))
		return
	}
	if !validSHA256(req.CorpusHash) || !validSHA256(req.MetadataHash) {
		fail(400, fmt.Errorf("invalid input hash"))
		return
	}
	if computed := computeFingerprint(req.Config.scientificConfig(), req.CorpusHash, req.MetadataHash); computed != req.ExperimentID {
		fail(409, fmt.Errorf("input/config fingerprint does not match experiment identity"))
		return
	}
	if req.JobID.Stage == "" || len(req.JobID.Combination) > 4096 || req.JobID.ReplicateIndex < 0 {
		fail(400, fmt.Errorf("invalid JobID"))
		return
	}
	limit := req.Config.Permutations
	if req.JobID.Stage == "part_a_refinement" {
		limit = refinementPermutations
	}
	if limit < 1 || req.JobID.ReplicateIndex >= limit {
		fail(400, fmt.Errorf("replicate index outside configured range"))
		return
	}
	w.mu.Lock()
	state := w.states[req.ExperimentID]
	if state == nil {
		init := req.Config
		init.CorpusPath, init.TokenMetadataMap = w.inputPath(req.CorpusHash), w.inputPath(req.MetadataHash)
		var err error
		state, err = newWorkerState(init)
		if err != nil {
			w.mu.Unlock()
			fail(409, err)
			return
		}
		w.states[req.ExperimentID] = state
	}
	w.mu.Unlock()
	value, err := state.compute(req.JobID)
	if err != nil {
		fail(422, err)
		return
	}
	resp.OK, resp.Value = true, value
	w.writeJobJSON(rw, 200, resp)
}

type remotePool struct {
	client    *http.Client
	endpoints []string
	token     string
	retries   int
	req       remoteJobRequest
	next      atomic.Uint64
	stageMu   sync.Mutex
	staged    map[string]bool
	corpus    []byte
	metadata  []byte
}

func newRemotePool(c Config, fingerprint, corpusHash, metaHash string) (*remotePool, error) {
	if len(c.RemoteWorkers) == 0 {
		return nil, fmt.Errorf("remote executor requires at least one worker endpoint")
	}
	corpus, err := os.ReadFile(c.CorpusPath)
	if err != nil {
		return nil, err
	}
	metadata, err := os.ReadFile(c.TokenMetadataMap)
	if err != nil {
		return nil, err
	}
	p := &remotePool{client: &http.Client{Timeout: c.RemoteTimeout}, token: c.RemoteToken, retries: c.RemoteRetries, staged: map[string]bool{}, corpus: corpus, metadata: metadata}
	init := protocolMessage{Fingerprint: fingerprint, WindowSizes: c.WindowSizes, ResidualWindowSizes: c.ResidualWindowSizes, MinClassTokens: c.MinClassTokens, MinBlockTokens: c.MinBlockTokens, KMin: c.KMin, KMaxWithin: c.KMaxWithin, KMaxResidual: c.KMaxResidual, Permutations: c.Permutations, Seed: c.Seed}
	p.req = remoteJobRequest{Protocol: remoteProtocolVersion, Compatibility: scientificCompatibilityVersion, GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, GoVersion: runtime.Version(), ExperimentID: fingerprint, CorpusHash: corpusHash, MetadataHash: metaHash, Config: init}
	ctx := c.Context
	if ctx == nil {
		ctx = context.Background()
	}
	available := 0
	var lastErr error
	for _, raw := range c.RemoteWorkers {
		ep := strings.TrimRight(strings.TrimSpace(raw), "/")
		if ep == "" {
			continue
		}
		p.endpoints = append(p.endpoints, ep)
		if err := p.ensureEndpoint(ctx, ep); err != nil {
			lastErr = err
			continue
		}
		available++
	}
	if len(p.endpoints) == 0 {
		return nil, fmt.Errorf("remote executor has no valid endpoints")
	}
	if available == 0 {
		return nil, fmt.Errorf("no remote worker available: %w", lastErr)
	}
	return p, nil
}

func (p *remotePool) ensureEndpoint(ctx context.Context, ep string) error {
	p.stageMu.Lock()
	if p.staged[ep] {
		p.stageMu.Unlock()
		return nil
	}
	p.stageMu.Unlock()
	if err := p.checkAndStage(ctx, ep, p.req.CorpusHash, p.corpus, p.req.MetadataHash, p.metadata); err != nil {
		return err
	}
	p.stageMu.Lock()
	p.staged[ep] = true
	p.stageMu.Unlock()
	return nil
}

func (p *remotePool) request(ctx context.Context, method, url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if p.token != "" {
		req.Header.Set("Authorization", "Bearer "+p.token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return p.client.Do(req)
}

func (p *remotePool) checkAndStage(ctx context.Context, ep string, corpusHash string, corpus []byte, metaHash string, metadata []byte) error {
	resp, err := p.request(ctx, "GET", ep+"/v1/info", nil)
	if err != nil {
		return fmt.Errorf("remote worker %s compatibility check: %w", ep, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("remote worker %s compatibility check: HTTP %s", ep, resp.Status)
	}
	var info remoteInfo
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxRemoteMessageBytes)).Decode(&info); err != nil {
		return err
	}
	local := localRemoteInfo("")
	if info.Protocol != local.Protocol || info.Compatibility != local.Compatibility || info.GOOS != local.GOOS || info.GOARCH != local.GOARCH || info.GoVersion != local.GoVersion {
		return fmt.Errorf("remote worker %s (%s) incompatible: protocol/code/runtime %d/%s/%s/%s/%s", ep, info.Host, info.Protocol, info.Compatibility, info.GOOS, info.GOARCH, info.GoVersion)
	}
	for _, input := range []struct {
		hash string
		data []byte
	}{{corpusHash, corpus}, {metaHash, metadata}} {
		head, e := p.request(ctx, "HEAD", ep+"/v1/input/"+input.hash, nil)
		if e == nil {
			head.Body.Close()
		}
		if e == nil && head.StatusCode == 204 {
			continue
		}
		put, e := p.request(ctx, "PUT", ep+"/v1/input/"+input.hash, input.data)
		if e != nil {
			return fmt.Errorf("stage input on %s: %w", ep, e)
		}
		io.Copy(io.Discard, io.LimitReader(put.Body, 4096))
		put.Body.Close()
		if put.StatusCode != 201 && put.StatusCode != 204 {
			return fmt.Errorf("stage input on %s: HTTP %s", ep, put.Status)
		}
	}
	return nil
}

func (p *remotePool) Run(ctx context.Context, id JobID) (float64, error) {
	start := int(p.next.Add(1)-1) % len(p.endpoints)
	var last error
	attempts := p.retries + 1
	if attempts < len(p.endpoints) {
		attempts = len(p.endpoints)
	}
	for attempt := 0; attempt < attempts; attempt++ {
		ep := p.endpoints[(start+attempt)%len(p.endpoints)]
		if err := p.ensureEndpoint(ctx, ep); err != nil {
			last = fmt.Errorf("remote worker %s job %+v: prepare endpoint: %w", ep, id, err)
			continue
		}
		req := p.req
		req.JobID = id
		b, _ := json.Marshal(req)
		resp, err := p.request(ctx, "POST", ep+"/v1/job", b)
		if err != nil {
			last = fmt.Errorf("remote worker %s job %+v: %w", ep, id, err)
			continue
		}
		var out remoteJobResponse
		decodeErr := json.NewDecoder(io.LimitReader(resp.Body, maxRemoteMessageBytes)).Decode(&out)
		resp.Body.Close()
		if decodeErr != nil {
			last = fmt.Errorf("remote worker %s job %+v: decode response: %w", ep, id, decodeErr)
			if resp.StatusCode >= 500 || resp.StatusCode == 200 {
				continue
			}
			return 0, last
		}
		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			last = fmt.Errorf("remote worker %s (%s) job %+v: HTTP %s: %s", ep, out.Host, id, resp.Status, out.Error)
			continue
		}
		if resp.StatusCode != 200 || !out.OK {
			return 0, fmt.Errorf("remote worker %s (%s) job %+v rejected: %s", ep, out.Host, id, out.Error)
		}
		if out.ExperimentID != p.req.ExperimentID || out.JobID != id {
			return 0, fmt.Errorf("remote worker %s returned stale/mismatched result for job %+v", ep, id)
		}
		return out.Value, nil
	}
	return 0, fmt.Errorf("job %+v failed after %d attempt(s): %w", id, attempts, last)
}

func (p *remotePool) Close() error { return nil }
