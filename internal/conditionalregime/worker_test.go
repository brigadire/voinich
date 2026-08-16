package conditionalregime

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

// workerHarness drives RunWorker in-process (no subprocess needed for pure
// protocol-framing tests): RunWorker runs in its own goroutine reading inR/
// writing outW, while every send/recv below happens sequentially in the
// calling test goroutine - the standard safe pattern for a synchronous
// ping-pong protocol over io.Pipe, and the only one that never touches the
// shared bufio.Writer from more than one goroutine.
type workerHarness struct {
	t       *testing.T
	inW     io.WriteCloser
	writer  *bufio.Writer
	scanner *bufio.Scanner
	errCh   chan error
}

func startWorkerHarness(t *testing.T) *workerHarness {
	t.Helper()
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()
	errCh := make(chan error, 1)
	go func() { errCh <- RunWorker(context.Background(), inR, outW) }()
	scanner := bufio.NewScanner(outR)
	scanner.Buffer(make([]byte, 4096), 1<<20)
	return &workerHarness{t: t, inW: inW, writer: bufio.NewWriter(inW), scanner: scanner, errCh: errCh}
}

func (h *workerHarness) send(m protocolMessage) {
	h.t.Helper()
	if err := writeMessage(h.writer, m); err != nil {
		h.t.Fatalf("send %+v: %v", m, err)
	}
}

func (h *workerHarness) recv() protocolMessage {
	h.t.Helper()
	if !h.scanner.Scan() {
		h.t.Fatalf("no message from worker: %v", h.scanner.Err())
	}
	var m protocolMessage
	if err := json.Unmarshal(h.scanner.Bytes(), &m); err != nil {
		h.t.Fatalf("decode message: %v", err)
	}
	return m
}

// closeAndWait closes stdin (a clean coordinator shutdown) and returns
// RunWorker's own exit error.
func (h *workerHarness) closeAndWait() error {
	_ = h.inW.Close()
	return <-h.errCh
}

func TestWorkerRejectsProtocolVersionMismatch(t *testing.T) {
	f := writeFixture(t)
	c := f.smallConfig()
	init := f.initMessage(t, c)
	init.Version = workerProtocolVersion + 1

	h := startWorkerHarness(t)
	h.send(init)
	reply := h.recv()
	if err := h.closeAndWait(); err == nil {
		t.Fatal("expected an error for a protocol version mismatch")
	}
	if reply.Kind != "ready" || reply.OK {
		t.Fatalf("expected an explicit ready/OK=false reply, got %+v", reply)
	}
	if !strings.Contains(reply.Error, "version") {
		t.Fatalf("expected the reply to explain the version mismatch, got %q", reply.Error)
	}
}

func TestWorkerRejectsFingerprintMismatch(t *testing.T) {
	f := writeFixture(t)
	c := f.smallConfig()
	init := f.initMessage(t, c)
	init.Fingerprint = "wrong"

	h := startWorkerHarness(t)
	h.send(init)
	reply := h.recv()
	if err := h.closeAndWait(); err == nil {
		t.Fatal("expected an error for a fingerprint mismatch")
	}
	if reply.OK {
		t.Fatalf("expected ready/OK=false, got %+v", reply)
	}
	if !strings.Contains(reply.Error, "fingerprint") {
		t.Fatalf("expected the reply to explain the fingerprint mismatch, got %q", reply.Error)
	}
}

func TestWorkerRejectsNonInitFirstMessage(t *testing.T) {
	f := writeFixture(t)
	c := f.smallConfig()
	init := f.initMessage(t, c)
	init.Kind = "job" // malformed handshake: anything other than "init" first

	h := startWorkerHarness(t)
	h.send(init)
	reply := h.recv()
	if err := h.closeAndWait(); err == nil {
		t.Fatal("expected an error when the first message is not an init")
	}
	if reply.OK {
		t.Fatalf("expected ready/OK=false, got %+v", reply)
	}
}

func TestWorkerServesJobsThenShutsDownOnClose(t *testing.T) {
	f := writeFixture(t)
	c := f.smallConfig()
	h := startWorkerHarness(t)
	h.send(f.initMessage(t, c))
	if ready := h.recv(); !ready.OK {
		t.Fatalf("expected a successful handshake, got %+v", ready)
	}

	id := JobID{Stage: "part_b_global_correction", Combination: "k_medoids|raw", ReplicateIndex: 0}
	h.send(protocolMessage{Kind: "job", JobID: &id})
	result := h.recv()
	if result.Kind != "result" || result.JobID == nil || *result.JobID != id || result.Error != "" {
		t.Fatalf("unexpected result: %+v", result)
	}

	if err := h.closeAndWait(); err != nil {
		t.Fatalf("RunWorker should exit cleanly when stdin closes: %v", err)
	}
}

func TestWorkerReportsUnknownStageAsJobErrorNotCrash(t *testing.T) {
	f := writeFixture(t)
	c := f.smallConfig()
	h := startWorkerHarness(t)
	h.send(f.initMessage(t, c))
	if ready := h.recv(); !ready.OK {
		t.Fatalf("expected a successful handshake, got %+v", ready)
	}

	id := JobID{Stage: "not_a_real_stage", Combination: "x", ReplicateIndex: 0}
	h.send(protocolMessage{Kind: "job", JobID: &id})
	result := h.recv()
	if result.Error == "" {
		t.Fatal("expected an explicit per-job error for an unknown stage, not a silent success")
	}

	_ = h.closeAndWait()
}
