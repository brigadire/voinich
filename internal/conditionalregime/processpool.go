package conditionalregime

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

// newExecutorPool builds the process pool for Config.Executor == "process",
// or returns (nil, nil) for the "goroutine" (default) backend. The pool
// re-execs the current binary with -internal-worker: the coordinator and
// every worker are always the exact same build (Task32 phase 3 - "one
// scientific implementation"), so there is never a second copy of any
// formula to keep in sync. fingerprint must be the same value
// computeFingerprint produced for the coordinator's own corpus/metadata/
// parameters; every worker independently recomputes it from the paths it is
// given and refuses to run on a mismatch.
type jobExecutor interface {
	Run(context.Context, JobID) (float64, error)
	Close() error
}

func newExecutorPool(c Config, fingerprint, corpusHash, metaHash string) (jobExecutor, error) {
	switch c.Executor {
	case "", "goroutine":
		return nil, nil
	case "process":
		selfPath, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("resolve executable path for process workers: %w", err)
		}
		init := protocolMessage{
			Fingerprint:         fingerprint,
			CorpusPath:          c.CorpusPath,
			TokenMetadataMap:    c.TokenMetadataMap,
			WindowSizes:         c.WindowSizes,
			ResidualWindowSizes: c.ResidualWindowSizes,
			MinClassTokens:      c.MinClassTokens,
			MinBlockTokens:      c.MinBlockTokens,
			KMin:                c.KMin,
			KMaxWithin:          c.KMaxWithin,
			KMaxResidual:        c.KMaxResidual,
			Permutations:        c.Permutations,
			Seed:                c.Seed,
		}
		ctx := c.Context
		if ctx == nil {
			ctx = context.Background()
		}
		newCmd := func() *exec.Cmd {
			cmd := exec.CommandContext(ctx, selfPath, "-internal-worker")
			cmd.Stderr = os.Stderr
			return cmd
		}
		return newProcessPool(c.Workers, newCmd, init)
	case "remote":
		return newRemotePool(c, fingerprint, corpusHash, metaHash)
	default:
		return nil, fmt.Errorf("unknown executor %q (want \"goroutine\", \"process\" or \"remote\")", c.Executor)
	}
}

// processWorkerShutdownTimeout bounds how long Close waits for a worker to
// exit after asking it to shut down cleanly before it is killed outright.
// This is what makes "no orphan/zombie workers" true even if a worker is
// stuck: every code path that starts a process also waits on it exactly
// once, with a bounded fallback to Kill.
const processWorkerShutdownTimeout = 5 * time.Second

// processWorker is one persistent subprocess worker: one OS process handling
// jobs one at a time over its own stdin/stdout pipes, started once and
// reused for the pool's whole lifetime. Task32 measured that
// process-per-replicate would repeat the corpus/metadata parse on every
// single job - milliseconds against jobs that can be tens of milliseconds
// (Part A) or seconds (Part B), and strictly wasted against a job whose
// scientific result never changes - so the pool always uses persistent
// workers; see DISTRIBUTED_EXECUTION_IMPLEMENTATION.md for the measurement.
type processWorker struct {
	index   int
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	scanner *bufio.Scanner
	writer  *bufio.Writer
}

func startProcessWorker(index int, newCmd func() *exec.Cmd, init protocolMessage) (*processWorker, error) {
	cmd := newCmd()
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("worker %d: stdin pipe: %w", index, err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("worker %d: stdout pipe: %w", index, err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("worker %d: start: %w", index, err)
	}
	w := &processWorker{
		index:   index,
		cmd:     cmd,
		stdin:   stdin,
		scanner: bufio.NewScanner(stdout),
		writer:  bufio.NewWriter(stdin),
	}
	w.scanner.Buffer(make([]byte, 4096), 1<<20)

	init.Kind = "init"
	init.Version = workerProtocolVersion
	if err := writeMessage(w.writer, init); err != nil {
		_ = w.terminate()
		return nil, fmt.Errorf("worker %d: sending init: %w", index, err)
	}
	ready, ok, err := readMessage(w.scanner)
	if err != nil {
		_ = w.terminate()
		return nil, fmt.Errorf("worker %d: reading ready: %w", index, err)
	}
	if !ok {
		_ = w.terminate()
		return nil, fmt.Errorf("worker %d: exited before completing the handshake", index)
	}
	if ready.Kind != "ready" || !ready.OK {
		_ = w.terminate()
		return nil, fmt.Errorf("worker %d: rejected handshake: %s", index, ready.Error)
	}
	return w, nil
}

// run sends exactly one job and waits for its matching result. A protocol
// violation (wrong kind, wrong JobID, worker exit mid-job) is reported with
// both the worker index and the JobID, per Task32's diagnostics requirement.
func (w *processWorker) run(id JobID) (float64, error) {
	if err := writeMessage(w.writer, protocolMessage{Kind: "job", JobID: &id}); err != nil {
		return 0, fmt.Errorf("worker %d: job %+v: sending job: %w", w.index, id, err)
	}
	msg, ok, err := readMessage(w.scanner)
	if err != nil {
		return 0, fmt.Errorf("worker %d: job %+v: reading result: %w", w.index, id, err)
	}
	if !ok {
		return 0, fmt.Errorf("worker %d: job %+v: worker exited without returning a result", w.index, id)
	}
	if msg.Kind != "result" || msg.JobID == nil || *msg.JobID != id {
		return 0, fmt.Errorf("worker %d: job %+v: unexpected reply %+v", w.index, id, msg)
	}
	if msg.Error != "" {
		return 0, fmt.Errorf("worker %d: job %+v: %s", w.index, id, msg.Error)
	}
	return msg.Value, nil
}

// shutdown asks the worker to exit cleanly and always waits for it, falling
// back to Kill if it does not exit within timeout. Wait is called exactly
// once per process on every path through this function.
func (w *processWorker) shutdown(timeout time.Duration) error {
	_ = writeMessage(w.writer, protocolMessage{Kind: "shutdown"})
	_ = w.stdin.Close()

	done := make(chan error, 1)
	go func() { done <- w.cmd.Wait() }()
	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		_ = w.cmd.Process.Kill()
		return <-done
	}
}

// terminate is the startup-failure path: a worker that never completed its
// handshake is killed and waited on immediately rather than left running.
func (w *processWorker) terminate() error {
	_ = w.stdin.Close()
	if w.cmd.Process != nil {
		_ = w.cmd.Process.Kill()
	}
	return w.cmd.Wait()
}

// processPool is a bounded set of persistent worker processes: "maximum N
// active worker processes" (Task32 phase 4). Run dispatches one job to
// whichever worker is currently free; since every worker handles at most one
// job at a time, at most len(workers) processes are ever doing work
// concurrently, matching the goroutine executor's own bound. Close performs
// a clean shutdown of every worker exactly once, regardless of whether every
// job succeeded, so a fatal job error can never leave a process behind.
type processPool struct {
	workers []*processWorker
	free    chan *processWorker

	mu     sync.Mutex
	closed bool
}

// newProcessPool starts n persistent workers, each built by newCmd and
// handshaked with init (Version/Kind are set by startProcessWorker; every
// other field must already describe the coordinator's explicit input/config
// identity). If any worker fails to start or handshake, every already-
// started worker is cleanly shut down before the error is returned - a
// partially-failed pool never leaks processes.
func newProcessPool(n int, newCmd func() *exec.Cmd, init protocolMessage) (*processPool, error) {
	if n < 1 {
		return nil, fmt.Errorf("process pool requires at least 1 worker")
	}
	p := &processPool{free: make(chan *processWorker, n)}
	for i := range n {
		w, err := startProcessWorker(i, newCmd, init)
		if err != nil {
			_ = p.Close()
			return nil, err
		}
		p.workers = append(p.workers, w)
		p.free <- w
	}
	return p, nil
}

// Run executes one job on the next free worker process, blocking until one
// is free or ctx is done. A worker whose job fails is not returned to the
// free rotation - Task31/32 semantics require one fatal job error to
// terminate the whole batch, and Close still shuts down every worker
// (including that one) exactly once regardless of this.
func (p *processPool) Run(ctx context.Context, id JobID) (float64, error) {
	select {
	case w, ok := <-p.free:
		if !ok {
			return 0, fmt.Errorf("job %+v: process pool is closed", id)
		}
		value, err := w.run(id)
		if err != nil {
			return 0, err
		}
		p.free <- w
		return value, nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

// Close shuts down every worker exactly once. It is idempotent and safe to
// call from a defer even after a fatal error aborted a job batch early.
func (p *processPool) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	p.mu.Unlock()

	var firstErr error
	for _, w := range p.workers {
		if err := w.shutdown(processWorkerShutdownTimeout); err != nil {
			if _, isExit := err.(*exec.ExitError); !isExit && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}
