package g1v2

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sort"
	"sync"
	"syscall"
	"time"
)

type Compatibility struct {
	ProtocolVersion string `json:"protocol_version"`
	CodeHash        string `json:"code_hash"`
	GOOS            string `json:"goos"`
	GOARCH          string `json:"goarch"`
	GoVersion       string `json:"go_version"`
	FreeStorage     int64  `json:"free_storage_bytes"`
	CacheAvailable  bool   `json:"cache_available"`
}

func LocalCompatibility(codeHash string, free int64) Compatibility {
	return Compatibility{ProtocolVersion, codeHash, runtime.GOOS, runtime.GOARCH, runtime.Version(), free, true}
}

func peakRSSBytes() int64 {
	var u syscall.Rusage
	if syscall.Getrusage(syscall.RUSAGE_SELF, &u) == nil {
		return u.Maxrss * 1024
	}
	return 0
}
func processCPUSeconds() float64 {
	var u syscall.Rusage
	if syscall.Getrusage(syscall.RUSAGE_SELF, &u) != nil {
		return 0
	}
	return float64(u.Utime.Sec+u.Stime.Sec) + float64(u.Utime.Usec+u.Stime.Usec)/1e6
}

func (c Compatibility) Compatible(w Compatibility) error {
	if w.ProtocolVersion != c.ProtocolVersion || w.CodeHash != c.CodeHash || w.GOOS != c.GOOS || w.GOARCH != c.GOARCH || w.GoVersion != c.GoVersion {
		return fmt.Errorf("worker executable/platform/runtime incompatible")
	}
	if !w.CacheAvailable || w.FreeStorage < c.FreeStorage {
		return fmt.Errorf("worker storage/cache incompatible")
	}
	return nil
}

type Lease struct {
	ID       string    `json:"lease_id"`
	Job      JobBundle `json:"job"`
	Attempt  int       `json:"attempt"`
	Deadline string    `json:"deadline"`
}
type leaseState struct {
	Lease
	Worker   string
	deadline time.Time
}

type Coordinator struct {
	mu       sync.Mutex
	manifest Manifest
	jobs     map[string]JobBundle
	store    Store
	expected Compatibility
	timeout  time.Duration
	leases   map[string]leaseState
	byJob    map[string]string
	attempts map[string]int
	history  map[string][]string
}

func NewCoordinator(m Manifest, store Store, expected Compatibility, timeout time.Duration) (*Coordinator, error) {
	if err := m.Validate(); err != nil {
		return nil, err
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("lease timeout must be positive")
	}
	c := &Coordinator{manifest: m, jobs: map[string]JobBundle{}, store: store, expected: expected, timeout: timeout, leases: map[string]leaseState{}, byJob: map[string]string{}, attempts: map[string]int{}, history: map[string][]string{}}
	for _, j := range m.Jobs {
		c.jobs[j.JobID] = j
	}
	if err := store.ensure(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Coordinator) sweep(now time.Time) {
	for id, l := range c.leases {
		if !now.Before(l.deadline) {
			delete(c.byJob, l.Job.JobID)
			delete(c.leases, id)
		}
	}
}

func (c *Coordinator) dependencyReady(j JobBundle) bool {
	for _, d := range j.DependsOn {
		x, err := c.store.ReadIndex(d)
		if err != nil || x.Status != "VERIFIED" {
			return false
		}
	}
	return true
}

func (c *Coordinator) Claim(worker string, info Compatibility, now time.Time) (*Lease, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if worker == "" {
		return nil, fmt.Errorf("authenticated worker identity required")
	}
	if err := c.expected.Compatible(info); err != nil {
		return nil, err
	}
	c.sweep(now)
	ids := make([]string, 0, len(c.jobs))
	for id := range c.jobs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		j := c.jobs[id]
		if c.store.Completed(id) || c.byJob[id] != "" || !c.dependencyReady(j) {
			continue
		}
		c.attempts[id]++
		lid := HashBytes([]byte(fmt.Sprintf("%s\x00%s\x00%d\x00%d", id, worker, c.attempts[id], now.UnixNano())))
		l := Lease{lid, j, c.attempts[id], now.Add(c.timeout).UTC().Format(time.RFC3339Nano)}
		c.leases[lid] = leaseState{l, worker, now.Add(c.timeout)}
		c.byJob[id] = lid
		c.history[id] = append(c.history[id], lid)
		return &l, nil
	}
	return nil, nil
}

func (c *Coordinator) Submit(worker, leaseID string, r ScientificResult, t Telemetry) (IndexRecord, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	l, ok := c.leases[leaseID]
	if !ok || l.Worker != worker || l.Job.JobID != r.ProducingJobID {
		return IndexRecord{}, fmt.Errorf("unknown or mismatched lease")
	}
	delete(c.leases, leaseID)
	delete(c.byJob, l.Job.JobID)
	t.Worker = worker
	t.LeaseHistory = append([]string(nil), c.history[l.Job.JobID]...)
	t.RetryCount = c.attempts[l.Job.JobID] - 1
	return c.store.Publish(l.Job, r, t)
}

func (c *Coordinator) Counts() (completed, total, leased int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for id := range c.jobs {
		if c.store.Completed(id) {
			completed++
		}
	}
	return completed, len(c.jobs), len(c.leases)
}

func ExecuteEngineering(ctx context.Context, j JobBundle) (ScientificResult, error) {
	if err := j.Validate(); err != nil {
		return ScientificResult{}, err
	}
	if j.Work.Kind != "sha256-chain-v1" {
		return ScientificResult{}, fmt.Errorf("unsupported engineering work kind %q", j.Work.Kind)
	}
	if j.Work.Iterations < 1 || j.Work.Iterations > 50_000_000 {
		return ScientificResult{}, fmt.Errorf("invalid engineering iterations")
	}
	h := sha256.Sum256([]byte(j.Work.Payload))
	b := h[:]
	for i := 0; i < j.Work.Iterations; i++ {
		if i%10000 == 0 {
			select {
			case <-ctx.Done():
				return ScientificResult{}, ctx.Err()
			default:
			}
		}
		x := sha256.Sum256(b)
		b = x[:]
	}
	payload, _ := canonicalJSON(struct {
		Digest     string `json:"digest"`
		Iterations int    `json:"iterations"`
		Model      string `json:"model"`
		Stage      string `json:"stage"`
	}{hex.EncodeToString(b), j.Work.Iterations, j.Model, j.Stage})
	a := Artifact{Name: "engineering_result", Hash: HashBytes(payload), Data: payload}
	return ScientificResult{SchemaVersion, j.JobID, append([]string(nil), j.InputHashes...), append([]string(nil), j.DependencyHashes...), j.CodeHash, j.ConfigHash, j.Seed, "EVIDENCE_COMPLETE", []Artifact{a}}, nil
}

func RunWorkers(ctx context.Context, c *Coordinator, workers int, info Compatibility) error {
	if workers < 1 {
		return fmt.Errorf("workers must be positive")
	}
	var wg sync.WaitGroup
	errCh := make(chan error, workers)
	for n := 0; n < workers; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			worker := fmt.Sprintf("local-worker-%d", n)
			host, _ := os.Hostname()
			for {
				select {
				case <-ctx.Done():
					errCh <- ctx.Err()
					return
				default:
				}
				start := time.Now()
				cpuStart := processCPUSeconds()
				l, err := c.Claim(worker, info, start)
				if err != nil {
					errCh <- err
					return
				}
				if l == nil {
					done, total, leased := c.Counts()
					if done == total {
						return
					}
					// Another worker may be between atomic publication and the next
					// readiness observation. A validated DAG cannot be intrinsically
					// blocked; keep polling just like a persistent Phase-I worker.
					_ = leased
					time.Sleep(time.Millisecond)
					continue
				}
				r, err := ExecuteEngineering(ctx, l.Job)
				if err != nil {
					errCh <- err
					return
				}
				end := time.Now()
				rb, _ := json.Marshal(r)
				t := Telemetry{Worker: worker, Host: host, StartUTC: start.UTC().Format(time.RFC3339Nano), EndUTC: end.UTC().Format(time.RFC3339Nano), WallSeconds: end.Sub(start).Seconds(), CPUSeconds: processCPUSeconds() - cpuStart, TransferBytes: int64(len(rb)), InfrastructureStatus: "SUCCESS"}
				t.PeakRSSBytes = peakRSSBytes()
				if _, err = c.Submit(worker, l.ID, r, t); err != nil {
					errCh <- err
					return
				}
			}
		}(n)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}
