package conditionalregime

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"zcore.dev/voinich/internal/pki"
	"zcore.dev/voinich/internal/structuralprojection"
)

// Task34 mTLS transport.
//
// Task33's push transport had the coordinator dial a fixed list of worker
// HTTP endpoints. Task34 authenticates every peer with an individual
// project-issued certificate (internal/pki), and Go's TLS stack ties a
// certificate's role (serverAuth vs clientAuth, DNS/IP SAN vs URI-identity
// SAN) to which side of a connection dials and which side listens. Making
// "the coordinator presents a server certificate with DNS/IP SANs and
// verifies each worker's individual client certificate" true - the actual
// Task34 requirement - means the coordinator is now the TLS/HTTP listener
// and every worker is the one that dials in, leases a job, computes it with
// the exact unchanged scientific implementation (workerState.compute), and
// posts the result back. This inverts *who initiates the TCP connection*;
// it does not change JobID, RNG, scheduling, checkpoints or any scientific
// output - the coordinator still owns the same pending-job set and the same
// bounded-concurrency/retry semantics as Task33, just realized as a lease
// queue instead of direct per-endpoint dispatch.
const (
	remoteProtocolVersion          = 3 // Task40 adds typed workload descriptors and blob trial results
	scientificCompatibilityVersion = "distributed-task40-v1"
	maxRemoteInputBytes            = 64 << 20
	maxRemoteMessageBytes          = 1 << 20
	// remoteLeaseBackoff bounds how long a worker idles between "no work
	// available yet" polls, and the starting backoff after a transport
	// error. It is purely a polling cadence, not a scientific parameter.
	remoteLeaseBackoff = 200 * time.Millisecond
	remoteMaxBackoff   = 5 * time.Second
)

type remoteInfo struct {
	Protocol      int    `json:"protocol"`
	Compatibility string `json:"compatibility"`
	GOOS          string `json:"goos"`
	GOARCH        string `json:"goarch"`
	GoVersion     string `json:"go_version"`
	Host          string `json:"host"`
}

func localRemoteInfo(host string) remoteInfo {
	return remoteInfo{Protocol: remoteProtocolVersion, Compatibility: scientificCompatibilityVersion, GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, GoVersion: runtime.Version(), Host: host}
}

func (a remoteInfo) compatibleWith(b remoteInfo) bool {
	return a.Protocol == b.Protocol && a.Compatibility == b.Compatibility && a.GOOS == b.GOOS && a.GOARCH == b.GOARCH && a.GoVersion == b.GoVersion
}

// remoteHandshakeResponse is fetched once by a worker at startup: it is the
// coordinator's compatibility identity plus everything workerState needs to
// build the exact same scientific state the coordinator itself would build
// (the same fields protocolMessage already carries for the Task32
// subprocess handshake).
type remoteHandshakeResponse struct {
	Info         remoteInfo        `json:"info"`
	ExperimentID string            `json:"experiment_id"`
	CorpusHash   string            `json:"corpus_hash"`
	MetadataHash string            `json:"metadata_hash"`
	Config       protocolMessage   `json:"config"`
	Inputs       map[string]string `json:"inputs,omitempty"`
}

type remoteLeaseRequest struct {
	Protocol      int    `json:"protocol"`
	Compatibility string `json:"compatibility"`
	GOOS          string `json:"goos"`
	GOARCH        string `json:"goarch"`
	GoVersion     string `json:"go_version"`
	ExperimentID  string `json:"experiment_id"`
}

type remoteLeaseResponse struct {
	NoWork  bool   `json:"no_work,omitempty"`
	LeaseID string `json:"lease_id,omitempty"`
	JobID   JobID  `json:"job_id,omitempty"`
}

// remoteResultRequest reports one leased job's outcome. WorkerID is
// deliberately absent: the coordinator only ever trusts the WorkerID it
// derived from this connection's verified client certificate (phase 5/8),
// never a value the request body could claim.
type remoteResultRequest struct {
	ExperimentID string          `json:"experiment_id"`
	LeaseID      string          `json:"lease_id"`
	JobID        JobID           `json:"job_id"`
	Value        float64         `json:"value"`
	Blob         json.RawMessage `json:"blob,omitempty"`
	Error        string          `json:"error,omitempty"`
}

type remoteResultResponse struct {
	Accepted bool   `json:"accepted"`
	Error    string `json:"error,omitempty"`
}

type remoteMetrics struct {
	Handshakes        int64 `json:"handshakes"`
	LeasesIssued      int64 `json:"leases_issued"`
	LeasesReclaimed   int64 `json:"leases_reclaimed"`
	ResultsAccepted   int64 `json:"results_accepted"`
	ResultsRejected   int64 `json:"results_rejected"`
	PendingJobs       int   `json:"pending_jobs"`
	OutstandingLeases int   `json:"outstanding_leases"`
}

func validSHA256(s string) bool {
	b, err := hex.DecodeString(s)
	return err == nil && len(b) == sha256.Size && s == strings.ToLower(s)
}

func writeRemoteJSON(w http.ResponseWriter, status int, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(b)
}

func decodeJSONBody(r *http.Request, v any) error {
	b, err := io.ReadAll(io.LimitReader(r.Body, maxRemoteMessageBytes+1))
	if err != nil {
		return err
	}
	if len(b) > maxRemoteMessageBytes {
		return fmt.Errorf("request exceeds size limit")
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("malformed request: %w", err)
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("malformed request: trailing JSON value")
	}
	return nil
}

// pendingJob is one JobID the coordinator currently wants computed: some
// goroutine inside executePermutationJobs is blocked in remotePool.Run
// waiting on result.
type pendingJob struct {
	id       JobID
	result   chan jobOutcome
	attempts int
}

type jobOutcome struct {
	value float64
	blob  []byte
	err   error
}

// activeLease is one JobID currently assigned to one authenticated worker,
// with a deadline after which the coordinator reclaims it for another
// worker (phase 8's LeaseID: "execution attempt identity", distinct from
// both JobID and WorkerID).
type activeLease struct {
	jobID    JobID
	workerID string
	deadline time.Time
	job      *pendingJob
}

// remotePool is the Task34 coordinator: it is the TLS/HTTP listener
// (jobExecutor for Config.Executor == "remote"), holding the same pending
// job set and bounded-concurrency semantics Task33's remotePool held, now
// realized as a lease queue that authenticated workers dial in and drain.
type remotePool struct {
	fingerprint, corpusHash, metaHash string
	corpus, metadata                  []byte
	inputs                            map[string][]byte
	inputNames                        map[string]string
	cfgMsg                            protocolMessage
	host                              string

	timeout time.Duration
	retries int

	listener  net.Listener
	srv       *http.Server
	serveDone chan error
	stopSweep chan struct{}

	mu      sync.Mutex
	queue   []JobID
	pending map[JobID]*pendingJob
	leases  map[string]*activeLease

	handshakes      atomic.Int64
	leasesIssued    atomic.Int64
	leasesReclaimed atomic.Int64
	resultsAccepted atomic.Int64
	resultsRejected atomic.Int64
}

func newRemotePool(c Config, fingerprint, corpusHash, metaHash string) (*remotePool, error) {
	if c.RemoteListen == "" {
		return nil, fmt.Errorf("remote executor requires -remote-listen (coordinator mTLS bind address)")
	}
	if c.TLSCert == "" || c.TLSKey == "" || c.ClientCA == "" {
		return nil, fmt.Errorf("remote executor requires -tls-cert, -tls-key and -client-ca")
	}
	corpus, err := os.ReadFile(c.CorpusPath)
	if err != nil {
		return nil, err
	}
	metadata, err := os.ReadFile(c.TokenMetadataMap)
	if err != nil {
		return nil, err
	}
	deny, err := pki.LoadDenyList(c.RemoteDenyList)
	if err != nil {
		return nil, fmt.Errorf("load deny list: %w", err)
	}
	tlsCfg, err := pki.CoordinatorServerTLSConfig(c.TLSCert, c.TLSKey, c.ClientCA, deny)
	if err != nil {
		return nil, fmt.Errorf("coordinator TLS config: %w", err)
	}
	listener, err := tls.Listen("tcp", c.RemoteListen, tlsCfg)
	if err != nil {
		return nil, fmt.Errorf("listen %s: %w", c.RemoteListen, err)
	}
	host, _ := os.Hostname()
	p := &remotePool{
		fingerprint: fingerprint, corpusHash: corpusHash, metaHash: metaHash,
		corpus: corpus, metadata: metadata, host: host,
		inputs:     map[string][]byte{corpusHash: corpus, metaHash: metadata},
		inputNames: map[string]string{"corpus": corpusHash, "metadata": metaHash},
		timeout:    c.RemoteTimeout, retries: c.RemoteRetries,
		pending: map[JobID]*pendingJob{}, leases: map[string]*activeLease{},
		serveDone: make(chan error, 1), stopSweep: make(chan struct{}),
		cfgMsg: protocolMessage{
			Fingerprint: fingerprint, WindowSizes: c.WindowSizes, ResidualWindowSizes: c.ResidualWindowSizes,
			MinClassTokens: c.MinClassTokens, MinBlockTokens: c.MinBlockTokens, KMin: c.KMin, KMaxWithin: c.KMaxWithin,
			KMaxResidual: c.KMaxResidual, Permutations: c.Permutations, Seed: c.Seed,
		},
	}
	p.listener = listener
	p.srv = &http.Server{Handler: p.routes(), ReadHeaderTimeout: 5 * time.Second}
	go func() { p.serveDone <- p.srv.Serve(listener) }()
	go p.sweepLoop()
	return p, nil
}

func newStructuralRemotePool(c structuralprojection.Config, fingerprint string) (*remotePool, error) {
	if c.RemoteListen == "" {
		return nil, fmt.Errorf("remote executor requires -remote-listen")
	}
	if c.TLSCert == "" || c.TLSKey == "" || c.ClientCA == "" {
		return nil, fmt.Errorf("remote executor requires -tls-cert, -tls-key and -client-ca")
	}
	paths := map[string]string{"corpus": c.CorpusPath, "structural_pairs": c.StructuralPairsPath, "distance_pairs": c.DistancePairsPath, "families": c.FamiliesPath}
	inputs := map[string][]byte{}
	names := map[string]string{}
	for name, path := range paths {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		sum := fmt.Sprintf("%x", sha256.Sum256(b))
		inputs[sum] = b
		names[name] = sum
	}
	deny, err := pki.LoadDenyList(c.RemoteDenyList)
	if err != nil {
		return nil, err
	}
	tlsCfg, err := pki.CoordinatorServerTLSConfig(c.TLSCert, c.TLSKey, c.ClientCA, deny)
	if err != nil {
		return nil, err
	}
	listener, err := tls.Listen("tcp", c.RemoteListen, tlsCfg)
	if err != nil {
		return nil, fmt.Errorf("listen %s: %w", c.RemoteListen, err)
	}
	host, _ := os.Hostname()
	p := &remotePool{fingerprint: fingerprint, corpusHash: names["corpus"], inputs: inputs, inputNames: names, host: host, timeout: c.RemoteTimeout, retries: c.RemoteRetries,
		pending: map[JobID]*pendingJob{}, leases: map[string]*activeLease{}, serveDone: make(chan error, 1), stopSweep: make(chan struct{}), cfgMsg: structuralInit(c, fingerprint)}
	p.listener = listener
	p.srv = &http.Server{Handler: p.routes(), ReadHeaderTimeout: 5 * time.Second}
	go func() { p.serveDone <- p.srv.Serve(listener) }()
	go p.sweepLoop()
	return p, nil
}

// Addr returns the coordinator's actual bound listen address (useful when
// Config.RemoteListen used port 0, e.g. in tests).
func (p *remotePool) Addr() string { return p.listener.Addr().String() }

func (p *remotePool) leaseTimeout() time.Duration {
	if p.timeout <= 0 {
		return 10 * time.Minute
	}
	return p.timeout
}

func (p *remotePool) sweepLoop() {
	interval := p.leaseTimeout() / 4
	if interval < 250*time.Millisecond {
		interval = 250 * time.Millisecond
	}
	if interval > 5*time.Second {
		interval = 5 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			p.mu.Lock()
			p.reclaimExpiredLocked()
			p.mu.Unlock()
		case <-p.stopSweep:
			return
		}
	}
}

// reclaimExpiredLocked returns any job whose lease deadline has passed to
// the queue for another worker, up to Config.RemoteRetries reassignments;
// beyond that it fails the job outright. Called with p.mu held.
func (p *remotePool) reclaimExpiredLocked() {
	now := time.Now()
	for id, lease := range p.leases {
		if now.Before(lease.deadline) {
			continue
		}
		delete(p.leases, id)
		p.leasesReclaimed.Add(1)
		job := lease.job
		job.attempts++
		if job.attempts > p.retries {
			job.result <- jobOutcome{err: fmt.Errorf("job %+v: no worker returned a result after %d attempt(s) (last leased to worker %q)", lease.jobID, job.attempts, lease.workerID)}
			delete(p.pending, lease.jobID)
			continue
		}
		p.queue = append(p.queue, lease.jobID)
	}
}

func (p *remotePool) removeFromQueueLocked(id JobID) {
	for i, qid := range p.queue {
		if qid == id {
			p.queue = append(p.queue[:i], p.queue[i+1:]...)
			return
		}
	}
}

func newLeaseID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// Run enqueues id and blocks until some authenticated worker leases and
// completes it (or ctx is cancelled). Duplicate/late deliveries for a job
// already removed from p.pending are harmlessly ignored by the /v1/result
// handler, matching Task33's "duplicate delivery cannot contribute twice".
func (p *remotePool) Run(ctx context.Context, id JobID) (float64, error) {
	outcome, err := p.runOutcome(ctx, id)
	return outcome.value, err
}

func (p *remotePool) RunBlob(ctx context.Context, id JobID) ([]byte, error) {
	outcome, err := p.runOutcome(ctx, id)
	return outcome.blob, err
}

func (p *remotePool) runOutcome(ctx context.Context, id JobID) (jobOutcome, error) {
	job := &pendingJob{id: id, result: make(chan jobOutcome, 1)}
	p.mu.Lock()
	if _, exists := p.pending[id]; exists {
		p.mu.Unlock()
		return jobOutcome{}, fmt.Errorf("job %+v is already in flight", id)
	}
	p.pending[id] = job
	p.queue = append(p.queue, id)
	p.mu.Unlock()

	select {
	case outcome := <-job.result:
		return outcome, outcome.err
	case <-ctx.Done():
		p.mu.Lock()
		delete(p.pending, id)
		p.removeFromQueueLocked(id)
		p.mu.Unlock()
		return jobOutcome{}, ctx.Err()
	}
}

func (p *remotePool) Close() error {
	close(p.stopSweep)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := p.srv.Shutdown(shutdownCtx)
	if serveErr := <-p.serveDone; err == nil && !errors.Is(serveErr, http.ErrServerClosed) {
		err = serveErr
	}
	return err
}

func (p *remotePool) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/handshake", p.withWorkerID(p.handshake))
	mux.HandleFunc("GET /v1/input/{hash}", p.withWorkerID(p.input))
	mux.HandleFunc("POST /v1/lease", p.withWorkerID(p.lease))
	mux.HandleFunc("POST /v1/result", p.withWorkerID(p.result))
	mux.HandleFunc("GET /v1/metrics", p.withWorkerID(p.metrics))
	return mux
}

// withWorkerID derives the caller's WorkerID from its verified TLS client
// certificate (phase 5: "derive WorkerID from the verified peer
// certificate") before any handler runs. tls.Config.ClientAuth is always
// RequireAndVerifyClientCert (internal/pki.CoordinatorServerTLSConfig), so
// r.TLS.VerifiedChains is already non-empty by the time net/http calls this
// handler; the error path below is an unreachable-in-practice backstop, not
// the primary enforcement point.
func (p *remotePool) withWorkerID(next func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.TLS == nil {
			http.Error(rw, "TLS client certificate required", http.StatusUpgradeRequired)
			return
		}
		workerID, err := pki.RequestWorkerID(r.TLS.VerifiedChains)
		if err != nil {
			http.Error(rw, "unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}
		next(rw, r, workerID)
	}
}

func (p *remotePool) handshake(rw http.ResponseWriter, _ *http.Request, _ string) {
	p.handshakes.Add(1)
	writeRemoteJSON(rw, http.StatusOK, remoteHandshakeResponse{
		Info: localRemoteInfo(p.host), ExperimentID: p.fingerprint,
		CorpusHash: p.corpusHash, MetadataHash: p.metaHash, Config: p.cfgMsg, Inputs: p.inputNames,
	})
}

func (p *remotePool) input(rw http.ResponseWriter, r *http.Request, _ string) {
	hash := r.PathValue("hash")
	data, ok := p.inputs[hash]
	if !ok {
		http.Error(rw, "unknown input hash", http.StatusNotFound)
		return
	}
	rw.Header().Set("Content-Type", "application/octet-stream")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(data)
}

func (p *remotePool) lease(rw http.ResponseWriter, r *http.Request, workerID string) {
	var req remoteLeaseRequest
	if err := decodeJSONBody(r, &req); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	peer := remoteInfo{Protocol: req.Protocol, Compatibility: req.Compatibility, GOOS: req.GOOS, GOARCH: req.GOARCH, GoVersion: req.GoVersion}
	if !peer.compatibleWith(localRemoteInfo("")) {
		http.Error(rw, fmt.Sprintf("protocol/code/runtime compatibility mismatch: worker=%+v coordinator=%+v", peer, localRemoteInfo("")), http.StatusConflict)
		return
	}
	if req.ExperimentID != p.fingerprint {
		http.Error(rw, "stale or inconsistent experiment identity", http.StatusConflict)
		return
	}
	writeRemoteJSON(rw, http.StatusOK, p.nextLease(workerID))
}

func (p *remotePool) nextLease(workerID string) remoteLeaseResponse {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.reclaimExpiredLocked()
	if len(p.queue) == 0 {
		return remoteLeaseResponse{NoWork: true}
	}
	id := p.queue[0]
	p.queue = p.queue[1:]
	job, ok := p.pending[id]
	if !ok { // raced with a ctx-cancelled Run() removing it; try the rest later
		return remoteLeaseResponse{NoWork: true}
	}
	leaseID := newLeaseID()
	p.leases[leaseID] = &activeLease{jobID: id, workerID: workerID, deadline: time.Now().Add(p.leaseTimeout()), job: job}
	p.leasesIssued.Add(1)
	return remoteLeaseResponse{LeaseID: leaseID, JobID: id}
}

func (p *remotePool) result(rw http.ResponseWriter, r *http.Request, workerID string) {
	var req remoteResultRequest
	if err := decodeJSONBody(r, &req); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	if req.ExperimentID != p.fingerprint {
		p.resultsRejected.Add(1)
		writeRemoteJSON(rw, http.StatusConflict, remoteResultResponse{Error: "stale or inconsistent experiment identity"})
		return
	}
	p.mu.Lock()
	lease, ok := p.leases[req.LeaseID]
	if !ok || lease.jobID != req.JobID {
		p.mu.Unlock()
		p.resultsRejected.Add(1)
		// Unknown, expired (already reassigned) or mismatched lease: ignored
		// rather than fatal. Run() either already has its answer from
		// whichever lease is now active, or is still waiting/retrying.
		writeRemoteJSON(rw, http.StatusConflict, remoteResultResponse{Error: "unknown or expired lease"})
		return
	}
	if lease.workerID != workerID {
		p.mu.Unlock()
		p.resultsRejected.Add(1)
		// A request cannot complete another worker's lease: WorkerID comes
		// only from this connection's verified certificate, never from the
		// request body (phase 5/9: "request cannot impersonate another
		// WorkerID").
		writeRemoteJSON(rw, http.StatusForbidden, remoteResultResponse{Error: "lease belongs to a different authenticated worker"})
		return
	}
	delete(p.leases, req.LeaseID)
	job := p.pending[req.JobID]
	delete(p.pending, req.JobID)
	p.mu.Unlock()
	if job == nil {
		writeRemoteJSON(rw, http.StatusOK, remoteResultResponse{Accepted: true})
		return
	}
	p.resultsAccepted.Add(1)
	outcome := jobOutcome{value: req.Value, blob: append([]byte(nil), req.Blob...)}
	if req.Error != "" {
		outcome.err = fmt.Errorf("worker %s: %s", workerID, req.Error)
	}
	job.result <- outcome
	writeRemoteJSON(rw, http.StatusOK, remoteResultResponse{Accepted: true})
}

func (p *remotePool) metrics(rw http.ResponseWriter, _ *http.Request, _ string) {
	p.mu.Lock()
	pending, leases := len(p.pending), len(p.leases)
	p.mu.Unlock()
	writeRemoteJSON(rw, http.StatusOK, remoteMetrics{
		Handshakes: p.handshakes.Load(), LeasesIssued: p.leasesIssued.Load(), LeasesReclaimed: p.leasesReclaimed.Load(),
		ResultsAccepted: p.resultsAccepted.Load(), ResultsRejected: p.resultsRejected.Load(),
		PendingJobs: pending, OutstandingLeases: leases,
	})
}

// ---- worker (TLS client) side ----

// RunRemoteWorker runs the Task34 worker client loop against a single
// coordinator until ctx is cancelled: it verifies the coordinator's TLS
// chain and server name, presents its own client certificate, fetches the
// two frozen inputs once by content hash into cacheDir, then repeatedly
// leases and computes jobs with concurrency independent lease loops. A
// worker that cannot reach the coordinator (not yet started, restarting,
// or finished) backs off and keeps retrying rather than exiting: its
// lifecycle is controlled by ctx/SIGINT, not by any one coordinator run.
func RunRemoteWorker(ctx context.Context, coordinatorURL, caFile, certFile, keyFile, cacheDir string, concurrency int) error {
	if concurrency < 1 {
		return fmt.Errorf("remote worker concurrency must be positive")
	}
	if cacheDir == "" {
		return fmt.Errorf("remote worker cache directory is required")
	}
	tlsCfg, err := pki.WorkerClientTLSConfig(certFile, keyFile, caFile)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return err
	}
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsCfg}}
	base := strings.TrimRight(strings.TrimSpace(coordinatorURL), "/")
	if base == "" {
		return fmt.Errorf("remote worker requires a coordinator URL")
	}
	u, err := url.Parse(base)
	if err != nil {
		return fmt.Errorf("invalid coordinator URL %q: %w", base, err)
	}
	// A worker may be started before, alongside, or after its coordinator
	// process; wait for the coordinator's TCP listener to accept
	// connections before attempting the authenticated handshake. This is
	// pure connectivity, decoupled from TLS/certificate validity: once
	// reachable, any handshake/auth failure below is real and fatal, never
	// silently retried.
	if err := awaitTCPReachable(ctx, u.Host); err != nil {
		return fmt.Errorf("coordinator %s unreachable: %w", u.Host, err)
	}

	hs, err := fetchHandshake(ctx, client, base)
	if err != nil {
		return fmt.Errorf("handshake with coordinator %s: %w", base, err)
	}
	if err := stageInputs(ctx, client, base, cacheDir, hs); err != nil {
		return fmt.Errorf("stage inputs from coordinator %s: %w", base, err)
	}
	if len(hs.Inputs) == 0 {
		hs.Inputs = map[string]string{"corpus": hs.CorpusHash, "metadata": hs.MetadataHash}
	}
	init := hs.Config
	init.CorpusPath = filepath.Join(cacheDir, hs.Inputs["corpus"])
	var computer protocolComputer
	if init.Workload == "structural_projection" {
		init.StructuralPairsPath = filepath.Join(cacheDir, hs.Inputs["structural_pairs"])
		init.DistancePairsPath = filepath.Join(cacheDir, hs.Inputs["distance_pairs"])
		init.FamiliesPath = filepath.Join(cacheDir, hs.Inputs["families"])
		cfg := structuralConfigFromMessage(init)
		computed, e := structuralprojection.Fingerprint(cfg)
		if e != nil || computed != hs.ExperimentID {
			return fmt.Errorf("input/config fingerprint does not match coordinator's declared experiment identity")
		}
		state, e := structuralprojection.NewTrialWorker(cfg)
		if e != nil {
			return fmt.Errorf("build structural worker state: %w", e)
		}
		computer = structuralComputer{state: state}
		concurrency = 1
	} else {
		init.TokenMetadataMap = filepath.Join(cacheDir, hs.Inputs["metadata"])
		if computed := computeFingerprint(init.scientificConfig(), hs.CorpusHash, hs.MetadataHash); computed != hs.ExperimentID {
			return fmt.Errorf("input/config fingerprint does not match coordinator's declared experiment identity")
		}
		state, e := newWorkerState(init)
		if e != nil {
			return fmt.Errorf("build worker state: %w", e)
		}
		computer = conditionalComputer{state: state}
	}

	var wg sync.WaitGroup
	errCh := make(chan error, concurrency)
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := workerLeaseLoop(ctx, client, base, hs.ExperimentID, computer); err != nil && ctx.Err() == nil {
				errCh <- err
			}
		}()
	}
	wg.Wait()
	close(errCh)
	return <-errCh
}

func workerLeaseLoop(ctx context.Context, client *http.Client, base, experimentID string, computer protocolComputer) error {
	backoff := remoteLeaseBackoff
	sleep := func() bool {
		select {
		case <-time.After(backoff):
			return true
		case <-ctx.Done():
			return false
		}
	}
	for {
		if ctx.Err() != nil {
			return nil
		}
		lease, err := requestLease(ctx, client, base, experimentID)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			if !sleep() {
				return nil
			}
			if backoff *= 2; backoff > remoteMaxBackoff {
				backoff = remoteMaxBackoff
			}
			continue
		}
		backoff = remoteLeaseBackoff
		if lease.NoWork {
			if !sleep() {
				return nil
			}
			continue
		}
		value, blob, computeErr := computer.compute(lease.JobID)
		resultReq := remoteResultRequest{ExperimentID: experimentID, LeaseID: lease.LeaseID, JobID: lease.JobID, Value: value, Blob: blob}
		if computeErr != nil {
			resultReq.Error = computeErr.Error()
		}
		// A rejected/expired-lease response means the coordinator already
		// reassigned this job elsewhere; this worker simply asks for its
		// next lease rather than treating that as fatal.
		_ = submitResult(ctx, client, base, resultReq)
	}
}

func awaitTCPReachable(ctx context.Context, addr string) error {
	backoff := remoteLeaseBackoff
	dialer := &net.Dialer{}
	for {
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err == nil {
			return conn.Close()
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return ctx.Err()
		}
		if backoff *= 2; backoff > remoteMaxBackoff {
			backoff = remoteMaxBackoff
		}
	}
}

func fetchHandshake(ctx context.Context, client *http.Client, base string) (remoteHandshakeResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/v1/handshake", nil)
	if err != nil {
		return remoteHandshakeResponse{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return remoteHandshakeResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return remoteHandshakeResponse{}, fmt.Errorf("GET /v1/handshake: HTTP %s", resp.Status)
	}
	var hs remoteHandshakeResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxRemoteMessageBytes)).Decode(&hs); err != nil {
		return remoteHandshakeResponse{}, err
	}
	if !hs.Info.compatibleWith(localRemoteInfo("")) {
		return remoteHandshakeResponse{}, fmt.Errorf("coordinator (%s) incompatible: protocol/code/runtime %+v", hs.Info.Host, hs.Info)
	}
	if !validSHA256(hs.CorpusHash) || (hs.MetadataHash != "" && !validSHA256(hs.MetadataHash)) {
		return remoteHandshakeResponse{}, fmt.Errorf("invalid input hash from coordinator")
	}
	for _, hash := range hs.Inputs {
		if !validSHA256(hash) {
			return remoteHandshakeResponse{}, fmt.Errorf("invalid input hash from coordinator")
		}
	}
	return hs, nil
}

func stageInputs(ctx context.Context, client *http.Client, base, cacheDir string, hs remoteHandshakeResponse) error {
	hashes := make([]string, 0, len(hs.Inputs))
	for _, hash := range hs.Inputs {
		hashes = append(hashes, hash)
	}
	if len(hashes) == 0 {
		hashes = []string{hs.CorpusHash, hs.MetadataHash}
	}
	for _, hash := range hashes {
		path := filepath.Join(cacheDir, hash)
		if _, err := os.Stat(path); err == nil {
			continue
		}
		if err := fetchInput(ctx, client, base, hash, path); err != nil {
			return err
		}
	}
	return nil
}

func fetchInput(ctx context.Context, client *http.Client, base, hash, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/v1/input/"+hash, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET input %s: HTTP %s", hash, resp.Status)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxRemoteInputBytes+1))
	if err != nil {
		return err
	}
	if len(b) > maxRemoteInputBytes {
		return fmt.Errorf("input %s exceeds limit", hash)
	}
	if actual := fmt.Sprintf("%x", sha256.Sum256(b)); actual != hash {
		return fmt.Errorf("input sha256 mismatch: coordinator served %s for requested %s", actual, hash)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".stage-")
	if err != nil {
		return err
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
	return err
}

func requestLease(ctx context.Context, client *http.Client, base, experimentID string) (remoteLeaseResponse, error) {
	body, err := json.Marshal(remoteLeaseRequest{
		Protocol: remoteProtocolVersion, Compatibility: scientificCompatibilityVersion,
		GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, GoVersion: runtime.Version(), ExperimentID: experimentID,
	})
	if err != nil {
		return remoteLeaseResponse{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/v1/lease", bytes.NewReader(body))
	if err != nil {
		return remoteLeaseResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return remoteLeaseResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return remoteLeaseResponse{}, fmt.Errorf("POST /v1/lease: HTTP %s: %s", resp.Status, string(b))
	}
	var out remoteLeaseResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxRemoteMessageBytes)).Decode(&out); err != nil {
		return remoteLeaseResponse{}, err
	}
	return out, nil
}

func submitResult(ctx context.Context, client *http.Client, base string, resultReq remoteResultRequest) error {
	body, err := json.Marshal(resultReq)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/v1/result", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxRemoteMessageBytes))
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("POST /v1/result: HTTP %s", resp.Status)
	}
	return nil
}
