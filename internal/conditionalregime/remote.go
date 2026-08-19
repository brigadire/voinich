package conditionalregime

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	mathrand "math/rand"
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

	"zcore.dev/voinich/internal/beginendanalyze"
	"zcore.dev/voinich/internal/higherorderseq"
	"zcore.dev/voinich/internal/normalization"
	"zcore.dev/voinich/internal/normalizationcompare"
	"zcore.dev/voinich/internal/pki"
	"zcore.dev/voinich/internal/positionalcontinuation"
	"zcore.dev/voinich/internal/replicatedlocalaudit"
	"zcore.dev/voinich/internal/structuralprojection"
	"zcore.dev/voinich/internal/tokenrelationvalidation"
	"zcore.dev/voinich/internal/transitionnetwork"
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
	// maxRemoteInputBytes bounds one staged input file a worker will fetch
	// via GET /v1/input/<hash>. Raised from 64 MiB after a real production
	// run (a larger generic corpus) produced a 67.5 MiB
	// soft_structural_pairs.tsv: every worker handshaked successfully but
	// silently failed stageInputs on every single generation ("input ...
	// exceeds limit"), forever, with no lease ever issued - the coordinator
	// itself has no matching cap when staging/serving these files, so only
	// the worker side needs raising. 256 MiB keeps a real, explicit bound
	// (never unbounded) while covering every input file size observed to
	// date with headroom for corpus growth.
	maxRemoteInputBytes = 256 << 20
	// maxRemoteMessageBytes bounds one lease/result/handshake JSON message
	// (not a staged input file - see maxRemoteInputBytes above). Every
	// pre-Task47 workload's result payload is a float64 or a small
	// edge-key/token map, comfortably under the original 1 MiB. Task47's
	// begin_end_candidate_batch is the first workload whose result payload
	// is verbatim per-candidate data (histograms, window tables) at
	// production batch sizes: a 2048-pair batch on the real Astafiev corpus
	// marshals to ~5.5 MB. At the original 1 MiB cap, decodeJSONBody
	// silently truncated the read, the coordinator answered "request
	// exceeds size limit" without draining the rest of the body, and the
	// worker's still-in-flight POST blocked on TCP backpressure until the
	// full remote-timeout elapsed - four times per job (one per retry),
	// with no error ever logged, since the failure was below the
	// application layer. Raised to 32 MiB: enough headroom for every batch
	// size the Task47 granularity study exercises (up to several thousand
	// pairs/batch) while remaining a real, explicit bound rather than
	// unbounded.
	maxRemoteMessageBytes = 32 << 20
	// remoteLeaseBackoff bounds how long a worker idles between "no work
	// available yet" polls, and the starting backoff after a transport
	// error. It is purely a polling cadence, not a scientific parameter.
	remoteLeaseBackoff = 200 * time.Millisecond
	remoteMaxBackoff   = 5 * time.Second

	// Task42 persistent-worker reconnect policy: bounded exponential
	// backoff with jitter between coordinator connection attempts, distinct
	// from remoteLeaseBackoff/remoteMaxBackoff above (which pace polling
	// *within* one already-established connection).
	reconnectMinBackoff = time.Second
	reconnectMaxBackoff = 60 * time.Second
)

// errStaleExperiment marks a /v1/lease rejection caused by the coordinator
// no longer recognizing this worker's experiment identity - either it
// restarted for a new experiment/run, or this worker is still holding an
// identity from before a coordinator restart. It is never a transport
// failure to retry indefinitely; a persistent worker treats it as "time to
// re-handshake."
var errStaleExperiment = errors.New("stale or inconsistent experiment identity")

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

// newNormalizationRemotePool is the Task42 coordinator for the
// normalization_compare_baseline job type: same lease queue/mTLS listener
// as newRemotePool and newStructuralRemotePool, staging the corpus and the
// complete structural_classes.yaml (every threshold, not just the one the
// coordinator happens to be comparing) once by content hash.
func newNormalizationRemotePool(c normalizationcompare.Config, classes normalization.ClassesOutput, fingerprint string) (*remotePool, error) {
	if c.RemoteListen == "" {
		return nil, fmt.Errorf("remote executor requires -remote-listen")
	}
	if c.TLSCert == "" || c.TLSKey == "" || c.ClientCA == "" {
		return nil, fmt.Errorf("remote executor requires -tls-cert, -tls-key and -client-ca")
	}
	paths := map[string]string{"corpus": c.InputPath, "classes": c.ClassesPath}
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
		pending: map[JobID]*pendingJob{}, leases: map[string]*activeLease{}, serveDone: make(chan error, 1), stopSweep: make(chan struct{}),
		cfgMsg: normalizationInit(c, classes.Meta.MinTokenCount, classes.Meta.SingletonMode, fingerprint)}
	p.listener = listener
	p.srv = &http.Server{Handler: p.routes(), ReadHeaderTimeout: 5 * time.Second}
	go func() { p.serveDone <- p.srv.Serve(listener) }()
	go p.sweepLoop()
	return p, nil
}

// newTokenRelationRemotePool is the Task44 coordinator for the
// token_relation_permutation job type: same lease queue/mTLS listener as
// every other remote pool, staging the corpus, the token metadata map
// (skipped in generic mode), and every discovery-dir file
// tokenrelationvalidation.LoadForDistribution needs, each once by content
// hash. Discovery files are staged under keys prefixed "discovery:" so a
// connecting worker can tell them apart from "corpus"/"metadata" and
// reconstruct a local directory with the original filenames (see
// runWorkerGeneration's "token_relation_permutation" case) - loadCandidates
// itself is never touched, it just gets handed a directory that looks
// identical to the coordinator's own.
func newTokenRelationRemotePool(c tokenrelationvalidation.Config, fingerprint string) (*remotePool, error) {
	if c.RemoteListen == "" {
		return nil, fmt.Errorf("remote executor requires -remote-listen")
	}
	if c.TLSCert == "" || c.TLSKey == "" || c.ClientCA == "" {
		return nil, fmt.Errorf("remote executor requires -tls-cert, -tls-key and -client-ca")
	}
	paths := map[string]string{"corpus": c.CorpusPath}
	if !c.Generic {
		paths["metadata"] = c.MetadataPath
	}
	for _, name := range tokenrelationvalidation.RequiredDiscoveryFiles {
		paths["discovery:"+name] = filepath.Join(c.DiscoveryDir, name)
	}
	for _, name := range tokenrelationvalidation.OptionalDiscoveryFiles {
		p := filepath.Join(c.DiscoveryDir, name)
		if _, err := os.Stat(p); err == nil {
			paths["discovery:"+name] = p
		}
	}
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
		pending: map[JobID]*pendingJob{}, leases: map[string]*activeLease{}, serveDone: make(chan error, 1), stopSweep: make(chan struct{}),
		cfgMsg: tokenRelationInit(c, fingerprint)}
	p.listener = listener
	p.srv = &http.Server{Handler: p.routes(), ReadHeaderTimeout: 5 * time.Second}
	go func() { p.serveDone <- p.srv.Serve(listener) }()
	go p.sweepLoop()
	return p, nil
}

// reconstructTokenRelationDiscoveryDir rebuilds a directory holding every
// staged discovery file under its *original* filename (loadCandidates
// hardcodes those names, so a worker's cache - which stores every staged
// blob under its content hash - cannot be handed to it directly). It
// hardlinks rather than copies where possible, falling back to a copy
// across filesystems; every file is already staged under cacheDir before
// this runs (stageInputs), so this never itself fetches anything.
func reconstructTokenRelationDiscoveryDir(cacheDir string, inputs map[string]string) (string, error) {
	return reconstructNamedDir(cacheDir, inputs, "discovery:", "token-relation-discovery")
}

// reconstructNamedDir rebuilds a directory holding every staged input file
// whose key carries the given prefix, restored under its *original*
// filename (the part of the key after the prefix). Some loaders
// (loadCandidates, replicatedlocalaudit's loadInputs/
// loadFrozenDistanceDiagnostics) hardcode specific filenames within a
// directory, so a worker's cache - which stores every staged blob under
// its content hash - cannot be handed to them directly. Hardlinks where
// possible, falling back to a copy across filesystems; every file is
// already staged under cacheDir before this runs (stageInputs), so this
// never itself fetches anything.
func reconstructNamedDir(cacheDir string, inputs map[string]string, prefix, dirName string) (string, error) {
	dir := filepath.Join(cacheDir, dirName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	for key, hash := range inputs {
		name, ok := strings.CutPrefix(key, prefix)
		if !ok {
			continue
		}
		src := filepath.Join(cacheDir, hash)
		dst := filepath.Join(dir, name)
		_ = os.Remove(dst)
		if err := os.Link(src, dst); err != nil {
			b, err := os.ReadFile(src)
			if err != nil {
				return "", err
			}
			if err := os.WriteFile(dst, b, 0644); err != nil {
				return "", err
			}
		}
	}
	return dir, nil
}

// newTransitionNetworkRemotePool is the Task44 coordinator for the
// transition_network_permutation job type: same lease queue/mTLS listener
// as every other remote pool, staging the corpus and (unless Generic) the
// token metadata map once by content hash.
func newTransitionNetworkRemotePool(c transitionnetwork.Config, fingerprint string) (*remotePool, error) {
	if c.RemoteListen == "" {
		return nil, fmt.Errorf("remote executor requires -remote-listen")
	}
	if c.TLSCert == "" || c.TLSKey == "" || c.ClientCA == "" {
		return nil, fmt.Errorf("remote executor requires -tls-cert, -tls-key and -client-ca")
	}
	paths := map[string]string{"corpus": c.CorpusPath}
	if !c.Generic {
		paths["metadata"] = c.MetadataPath
	}
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
		pending: map[JobID]*pendingJob{}, leases: map[string]*activeLease{}, serveDone: make(chan error, 1), stopSweep: make(chan struct{}),
		cfgMsg: transitionNetworkInit(c, fingerprint)}
	p.listener = listener
	p.srv = &http.Server{Handler: p.routes(), ReadHeaderTimeout: 5 * time.Second}
	go func() { p.serveDone <- p.srv.Serve(listener) }()
	go p.sweepLoop()
	return p, nil
}

// newBeginEndRemotePool is the Task47 coordinator for the
// begin_end_candidate_batch job type: same lease queue/mTLS listener as
// every other remote pool, staging the corpus and the dictionary once by
// content hash. Unlike every other stage's remote pool, the batch space
// this pool dispatches over is entirely RNG-free - every worker
// independently recomputes the same permutation-null moments via
// beginendanalyze.LoadForDistribution's own frozen, sequential RNG stream
// before ever answering a lease.
func newBeginEndRemotePool(c beginendanalyze.Config, fingerprint string) (*remotePool, error) {
	if c.RemoteListen == "" {
		return nil, fmt.Errorf("remote executor requires -remote-listen")
	}
	if c.TLSCert == "" || c.TLSKey == "" || c.ClientCA == "" {
		return nil, fmt.Errorf("remote executor requires -tls-cert, -tls-key and -client-ca")
	}
	paths := map[string]string{"corpus": c.CorpusPath, "dictionary": c.DictionaryPath}
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
		pending: map[JobID]*pendingJob{}, leases: map[string]*activeLease{}, serveDone: make(chan error, 1), stopSweep: make(chan struct{}),
		cfgMsg: beginEndInit(c, fingerprint)}
	p.listener = listener
	p.srv = &http.Server{Handler: p.routes(), ReadHeaderTimeout: 5 * time.Second}
	go func() { p.serveDone <- p.srv.Serve(listener) }()
	go p.sweepLoop()
	return p, nil
}

// newReplicatedLocalAuditRemotePool is the Task44 coordinator for the
// replicated_local_null job type: same lease queue/mTLS listener as every
// other remote pool, staging the corpus, the token metadata map (skipped in
// generic mode), and every relation-dir/discovery-dir file
// replicatedlocalaudit.LoadForDistribution needs, each once by content
// hash. Relation/discovery files are staged under keys prefixed
// "relation:"/"discovery:" so a connecting worker can reconstruct both
// directories under their original filenames (see runWorkerGeneration's
// "replicated_local_null" case).
func newReplicatedLocalAuditRemotePool(c replicatedlocalaudit.Config, fingerprint string) (*remotePool, error) {
	if c.RemoteListen == "" {
		return nil, fmt.Errorf("remote executor requires -remote-listen")
	}
	if c.TLSCert == "" || c.TLSKey == "" || c.ClientCA == "" {
		return nil, fmt.Errorf("remote executor requires -tls-cert, -tls-key and -client-ca")
	}
	paths := map[string]string{"corpus": c.CorpusPath}
	if !c.Generic {
		paths["metadata"] = c.MetadataPath
	}
	for _, name := range replicatedlocalaudit.RelationDirFiles {
		paths["relation:"+name] = filepath.Join(c.RelationDir, name)
	}
	for _, name := range replicatedlocalaudit.DiscoveryDirFiles {
		paths["discovery:"+name] = filepath.Join(c.DiscoveryDir, name)
	}
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
		pending: map[JobID]*pendingJob{}, leases: map[string]*activeLease{}, serveDone: make(chan error, 1), stopSweep: make(chan struct{}),
		cfgMsg: replicatedLocalAuditInit(c, fingerprint)}
	p.listener = listener
	p.srv = &http.Server{Handler: p.routes(), ReadHeaderTimeout: 5 * time.Second}
	go func() { p.serveDone <- p.srv.Serve(listener) }()
	go p.sweepLoop()
	return p, nil
}

// newHigherOrderRemotePool is the Task44 coordinator for the
// higher_order_candidate job type: same lease queue/mTLS listener as every
// other remote pool, staging the corpus, the token metadata map (skipped in
// generic mode), and every audit-dir/discovery-dir file
// higherorderseq.LoadForDistribution needs, each once by content hash.
// Audit/discovery files are staged under keys prefixed "audit:"/
// "discovery:" so a connecting worker can reconstruct both directories
// under their original filenames (see runWorkerGeneration's
// "higher_order_candidate" case).
func newHigherOrderRemotePool(c higherorderseq.Config, fingerprint string) (*remotePool, error) {
	if c.RemoteListen == "" {
		return nil, fmt.Errorf("remote executor requires -remote-listen")
	}
	if c.TLSCert == "" || c.TLSKey == "" || c.ClientCA == "" {
		return nil, fmt.Errorf("remote executor requires -tls-cert, -tls-key and -client-ca")
	}
	paths := map[string]string{"corpus": c.CorpusPath}
	if !c.Generic {
		paths["metadata"] = c.TokenMetadataMap
	}
	for _, name := range higherorderseq.AuditDirFiles {
		paths["audit:"+name] = filepath.Join(c.AuditDir, name)
	}
	for _, name := range higherorderseq.DiscoveryDirFiles {
		paths["discovery:"+name] = filepath.Join(c.DiscoveryDir, name)
	}
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
		pending: map[JobID]*pendingJob{}, leases: map[string]*activeLease{}, serveDone: make(chan error, 1), stopSweep: make(chan struct{}),
		cfgMsg: higherOrderInit(c, fingerprint)}
	p.listener = listener
	p.srv = &http.Server{Handler: p.routes(), ReadHeaderTimeout: 5 * time.Second}
	go func() { p.serveDone <- p.srv.Serve(listener) }()
	go p.sweepLoop()
	return p, nil
}

// newPositionalContinuationRemotePool is the Task44 coordinator for the
// positional_continuation_battery job type: same lease queue/mTLS listener
// as every other remote pool, staging the corpus, the token metadata map
// (skipped in generic mode), and every higher-order-dir file
// positionalcontinuation.LoadForDistribution needs, each once by content
// hash. Higher-order files are staged under keys prefixed "higherorder:" so
// a connecting worker can reconstruct that directory under its original
// filenames (see runWorkerGeneration's "positional_continuation_battery"
// case).
func newPositionalContinuationRemotePool(c positionalcontinuation.Config, fingerprint string) (*remotePool, error) {
	if c.RemoteListen == "" {
		return nil, fmt.Errorf("remote executor requires -remote-listen")
	}
	if c.TLSCert == "" || c.TLSKey == "" || c.ClientCA == "" {
		return nil, fmt.Errorf("remote executor requires -tls-cert, -tls-key and -client-ca")
	}
	paths := map[string]string{"corpus": c.CorpusPath}
	if !c.Generic {
		paths["metadata"] = c.TokenMetadataMap
	}
	for _, name := range positionalcontinuation.HigherOrderDirFiles {
		paths["higherorder:"+name] = filepath.Join(c.HigherOrderDir, name)
	}
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
		pending: map[JobID]*pendingJob{}, leases: map[string]*activeLease{}, serveDone: make(chan error, 1), stopSweep: make(chan struct{}),
		cfgMsg: positionalContinuationInit(c, fingerprint)}
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

// RunRemoteWorker runs one Task34 handshake-through-disconnect worker
// generation against a single coordinator: it verifies the coordinator's
// TLS chain and server name, presents its own client certificate, fetches
// this experiment's inputs once by content hash into cacheDir, then
// repeatedly leases and computes jobs with concurrency independent lease
// loops until ctx is cancelled or the coordinator stops recognizing this
// worker's experiment identity (errStaleExperiment). It never rebuilds its
// computer state mid-call - that only happens across generations, which is
// what RunPersistentRemoteWorker (Task42) is for. Most callers should use
// RunPersistentRemoteWorker; this lower-level entry point remains exported
// for tests and any caller that deliberately wants exactly one experiment's
// worth of connection.
func RunRemoteWorker(ctx context.Context, coordinatorURL, caFile, certFile, keyFile, cacheDir string, concurrency int) error {
	return runWorkerGeneration(ctx, coordinatorURL, caFile, certFile, keyFile, cacheDir, concurrency, noopWorkerNotify)
}

// RunPersistentRemoteWorker is the Task42 long-lived worker entry point: a
// worker deployed once keeps calling runWorkerGeneration for as long as ctx
// is alive, so it survives a coordinator that has not started yet, has
// stopped, has restarted, or has moved on to a new experiment - none of
// which require redeploying or restarting this process. Between
// generations it never carries over the previous generation's computer
// state (corpus, classes/metadata, fingerprint): every generation starts
// with a fresh handshake, so no experiment's state can contaminate another
// (Task42 phase 12/13). Reconnect attempts use bounded exponential backoff
// with jitter (reconnectMinBackoff..reconnectMaxBackoff), never a tight
// loop, and lifecycle state transitions are logged once per change, never
// once per attempt.
func RunPersistentRemoteWorker(ctx context.Context, coordinatorURL, caFile, certFile, keyFile, cacheDir string, concurrency int) error {
	logger := newWorkerStateLogger(os.Stderr)
	backoff := reconnectMinBackoff
	firstAttempt := true
	for {
		if ctx.Err() != nil {
			return nil
		}
		if !firstAttempt {
			logger.notify(stateReconnecting, "")
		}
		firstAttempt = false

		generationStart := time.Now()
		err := runWorkerGeneration(ctx, coordinatorURL, caFile, certFile, keyFile, cacheDir, concurrency, logger.notify)
		if ctx.Err() != nil {
			return nil
		}
		if isPermanentWorkerError(err) {
			logger.notify(stateUnavailable, fmt.Sprintf("permanent failure, not retrying: %v", err))
			return err
		}
		if err == nil {
			err = fmt.Errorf("coordinator closed the connection")
		}
		logger.notify(stateDisconnected, err.Error())
		// A generation that stayed connected for a while (successfully
		// leasing/computing, not just failing the handshake) resets the
		// backoff: a long-lived worker that loses an otherwise-healthy
		// connection once should reconnect quickly, not inherit whatever
		// backoff an earlier, unrelated outage had grown to.
		if time.Since(generationStart) >= reconnectMinBackoff {
			backoff = reconnectMinBackoff
		}
		if !sleepWithJitter(ctx, backoff) {
			return nil
		}
		backoff = nextReconnectBackoff(backoff)
	}
}

// workerLifecycleState is the small state machine a worker's connection to
// its coordinator moves through, logged on every transition (Task42
// section 10) and never spammed once per second while stable/idle.
type workerLifecycleState int

const (
	stateUnavailable workerLifecycleState = iota
	stateReconnecting
	stateConnected
	stateAuthenticated
	stateRegistered
	stateDisconnected
)

func (s workerLifecycleState) String() string {
	switch s {
	case stateUnavailable:
		return "coordinator unavailable"
	case stateReconnecting:
		return "reconnecting"
	case stateConnected:
		return "connected"
	case stateAuthenticated:
		return "authenticated"
	case stateRegistered:
		return "registered"
	case stateDisconnected:
		return "disconnected"
	default:
		return "unknown"
	}
}

type workerNotifyFunc func(state workerLifecycleState, detail string)

func noopWorkerNotify(workerLifecycleState, string) {}

// workerStateLogger prints a line only when the lifecycle state actually
// changes, so a worker idling in one state (e.g. polling a coordinator with
// no pending work) never repeats the same log line every second.
type workerStateLogger struct {
	out  io.Writer
	mu   sync.Mutex
	last workerLifecycleState
	seen bool
}

func newWorkerStateLogger(out io.Writer) *workerStateLogger { return &workerStateLogger{out: out} }

func (l *workerStateLogger) notify(state workerLifecycleState, detail string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.seen && state == l.last {
		return
	}
	l.seen, l.last = true, state
	if detail == "" {
		fmt.Fprintf(l.out, "worker: %s\n", state)
	} else {
		fmt.Fprintf(l.out, "worker: %s: %s\n", state, detail)
	}
}

// isPermanentWorkerError reports whether err reflects an mTLS identity
// problem (unknown/untrusted CA, wrong coordinator hostname, invalid
// certificate) rather than a transient transport failure. Task42 section
// 17 requires these to fail with clear diagnostics instead of retrying
// forever: they will not resolve themselves the way a coordinator restart
// or a brief network blip does.
func isPermanentWorkerError(err error) bool {
	if err == nil {
		return false
	}
	var unknownAuthority x509.UnknownAuthorityError
	var hostnameErr x509.HostnameError
	var certInvalid x509.CertificateInvalidError
	var certVerify *tls.CertificateVerificationError
	return errors.As(err, &unknownAuthority) || errors.As(err, &hostnameErr) || errors.As(err, &certInvalid) || errors.As(err, &certVerify)
}

// sleepWithJitter waits a random duration in [0, base) (full jitter, per
// Task42 section 10's "bounded exponential backoff with jitter"), returning
// false if ctx is cancelled first so a caller can shut down immediately
// instead of sleeping through the rest of a long backoff.
func sleepWithJitter(ctx context.Context, base time.Duration) bool {
	delay := base
	if base > 0 {
		delay = time.Duration(mathrand.Int63n(int64(base)))
	}
	select {
	case <-time.After(delay):
		return true
	case <-ctx.Done():
		return false
	}
}

func nextReconnectBackoff(cur time.Duration) time.Duration {
	next := cur * 2
	if next > reconnectMaxBackoff {
		next = reconnectMaxBackoff
	}
	return next
}

// runWorkerGeneration is exactly one handshake-through-disconnect
// connection to a coordinator: RunRemoteWorker calls it once,
// RunPersistentRemoteWorker calls it repeatedly across generations. Every
// call starts from zero - a fresh handshake and a freshly built computer
// state - so nothing from a previous generation (a different experiment's
// corpus, classes, or fingerprint) can leak into this one.
func runWorkerGeneration(ctx context.Context, coordinatorURL, caFile, certFile, keyFile, cacheDir string, concurrency int, notify workerNotifyFunc) error {
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
	// silently retried within this generation (RunPersistentRemoteWorker
	// decides whether a failure here starts a new generation or gives up
	// for good).
	unavailableLogged := false
	onUnreachable := func() {
		if !unavailableLogged {
			notify(stateUnavailable, base)
			unavailableLogged = true
		}
	}
	if err := awaitTCPReachable(ctx, u.Host, onUnreachable); err != nil {
		return fmt.Errorf("coordinator %s unreachable: %w", u.Host, err)
	}
	notify(stateConnected, base)

	hs, err := fetchHandshake(ctx, client, base)
	if err != nil {
		return fmt.Errorf("handshake with coordinator %s: %w", base, err)
	}
	notify(stateAuthenticated, "")
	if err := stageInputs(ctx, client, base, cacheDir, hs); err != nil {
		return fmt.Errorf("stage inputs from coordinator %s: %w", base, err)
	}
	if len(hs.Inputs) == 0 {
		hs.Inputs = map[string]string{"corpus": hs.CorpusHash, "metadata": hs.MetadataHash}
	}
	init := hs.Config
	init.CorpusPath = filepath.Join(cacheDir, hs.Inputs["corpus"])
	var computer protocolComputer
	switch init.Workload {
	case "structural_projection":
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
	case "normalization_compare":
		init.ClassesPath = filepath.Join(cacheDir, hs.Inputs["classes"])
		state, e := newNormalizationComputer(init)
		if e != nil {
			return fmt.Errorf("input/config fingerprint does not match coordinator's declared experiment identity: %w", e)
		}
		computer = state
	case "token_relation_permutation":
		if !init.Generic {
			init.TokenMetadataMap = filepath.Join(cacheDir, hs.Inputs["metadata"])
		}
		discoveryDir, e := reconstructTokenRelationDiscoveryDir(cacheDir, hs.Inputs)
		if e != nil {
			return fmt.Errorf("reconstruct discovery directory: %w", e)
		}
		init.DiscoveryDir = discoveryDir
		state, e := newTokenRelationComputer(init)
		if e != nil {
			return fmt.Errorf("input/config fingerprint does not match coordinator's declared experiment identity: %w", e)
		}
		computer = state
	case "transition_network_permutation":
		if !init.Generic {
			init.TokenMetadataMap = filepath.Join(cacheDir, hs.Inputs["metadata"])
		}
		state, e := newTransitionNetworkComputer(init)
		if e != nil {
			return fmt.Errorf("input/config fingerprint does not match coordinator's declared experiment identity: %w", e)
		}
		computer = state
	case "begin_end_candidate_batch":
		init.DictionaryPath = filepath.Join(cacheDir, hs.Inputs["dictionary"])
		state, e := newBeginEndComputer(init)
		if e != nil {
			return fmt.Errorf("input/config fingerprint does not match coordinator's declared experiment identity: %w", e)
		}
		computer = state
	case "replicated_local_null":
		if !init.Generic {
			init.TokenMetadataMap = filepath.Join(cacheDir, hs.Inputs["metadata"])
		}
		relationDir, e := reconstructNamedDir(cacheDir, hs.Inputs, "relation:", "replicated-local-relation")
		if e != nil {
			return fmt.Errorf("reconstruct relation directory: %w", e)
		}
		discoveryDir, e := reconstructNamedDir(cacheDir, hs.Inputs, "discovery:", "replicated-local-discovery")
		if e != nil {
			return fmt.Errorf("reconstruct discovery directory: %w", e)
		}
		init.RelationDir = relationDir
		init.DiscoveryDir = discoveryDir
		state, e := newReplicatedLocalAuditComputer(init)
		if e != nil {
			return fmt.Errorf("input/config fingerprint does not match coordinator's declared experiment identity: %w", e)
		}
		computer = state
	case "higher_order_candidate":
		if !init.Generic {
			init.TokenMetadataMap = filepath.Join(cacheDir, hs.Inputs["metadata"])
		}
		auditDir, e := reconstructNamedDir(cacheDir, hs.Inputs, "audit:", "higher-order-audit")
		if e != nil {
			return fmt.Errorf("reconstruct audit directory: %w", e)
		}
		discoveryDir, e := reconstructNamedDir(cacheDir, hs.Inputs, "discovery:", "higher-order-discovery")
		if e != nil {
			return fmt.Errorf("reconstruct discovery directory: %w", e)
		}
		init.AuditDir = auditDir
		init.DiscoveryDir = discoveryDir
		state, e := newHigherOrderComputer(init)
		if e != nil {
			return fmt.Errorf("input/config fingerprint does not match coordinator's declared experiment identity: %w", e)
		}
		computer = state
	case "positional_continuation_battery":
		if !init.Generic {
			init.TokenMetadataMap = filepath.Join(cacheDir, hs.Inputs["metadata"])
		}
		higherOrderDir, e := reconstructNamedDir(cacheDir, hs.Inputs, "higherorder:", "positional-continuation-higher-order")
		if e != nil {
			return fmt.Errorf("reconstruct higher-order directory: %w", e)
		}
		init.HigherOrderDir = higherOrderDir
		state, e := newPositionalContinuationComputer(init)
		if e != nil {
			return fmt.Errorf("input/config fingerprint does not match coordinator's declared experiment identity: %w", e)
		}
		computer = state
	default:
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
	notify(stateRegistered, init.Workload)

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
			if errors.Is(err, errStaleExperiment) {
				// Don't keep polling an experiment identity the coordinator
				// has already rejected: surface this immediately so the
				// caller can re-handshake instead of retrying it forever.
				return err
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

func awaitTCPReachable(ctx context.Context, addr string, onUnreachable func()) error {
	backoff := remoteLeaseBackoff
	dialer := &net.Dialer{}
	for {
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err == nil {
			return conn.Close()
		}
		if onUnreachable != nil {
			onUnreachable()
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
	if resp.StatusCode == http.StatusConflict {
		// The coordinator has moved on to a different experiment/run (or
		// this worker is stating a stale identity from before a coordinator
		// restart): this is not a transport error to retry forever against
		// - it is the signal a persistent worker uses to re-handshake and
		// rebuild its computer state for whatever the coordinator is
		// running now (Task42 phase 12: no experiment/job is bound to the
		// worker process itself).
		return remoteLeaseResponse{}, fmt.Errorf("%w: lease request rejected by coordinator", errStaleExperiment)
	}
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
