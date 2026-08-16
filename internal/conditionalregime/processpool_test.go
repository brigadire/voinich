package conditionalregime

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestMain lets this test binary re-exec itself as a Task32 subprocess
// worker, the standard way to test real process-boundary behavior without a
// separate binary: a child process launched with CRA_TEST_WORKER_MODE set
// never runs any *testing.T - it runs the worker protocol (or, for
// CRA_TEST_WORKER_MODE=crash, a deliberately truncated handshake-then-exit
// stand-in) and calls os.Exit directly.
func TestMain(m *testing.M) {
	// Mirrors conditional-regime-analyze/main.go's own "-internal-worker"
	// branch: RunAndWrite's process executor re-execs os.Executable() (this
	// test binary, under `go test`) with exactly that flag, so intercepting
	// it here - before any -test.* flag parsing - lets the real RunAndWrite
	// code path spawn and talk to a genuine worker subprocess in tests
	// without needing the conditional-regime-analyze binary built.
	for _, arg := range os.Args[1:] {
		if arg == "-internal-worker" || arg == "--internal-worker" || arg == "-internal-worker=true" {
			if err := RunWorker(context.Background(), os.Stdin, os.Stdout); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}
	switch os.Getenv("CRA_TEST_WORKER_MODE") {
	case "worker":
		if err := RunWorker(context.Background(), os.Stdin, os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		os.Exit(0)
	case "crash":
		// Simulates a worker that completes the handshake and then dies
		// unexpectedly (killed, OOM, panic past recover) instead of ever
		// serving a job - the pool must diagnose this, not hang.
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Buffer(make([]byte, 4096), 1<<20)
		writer := bufio.NewWriter(os.Stdout)
		if !scanner.Scan() {
			os.Exit(1)
		}
		_ = writeMessage(writer, protocolMessage{Kind: "ready", OK: true})
		os.Exit(3)
	}
	os.Exit(m.Run())
}

func testWorkerCmd(mode string) func() *exec.Cmd {
	return func() *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=^$")
		cmd.Env = append(os.Environ(), "CRA_TEST_WORKER_MODE="+mode)
		cmd.Stderr = os.Stderr
		return cmd
	}
}

// fixtureCorpus writes a small, fast, deterministic corpus + token metadata
// map to disk (RunWorker only ever reads the corpus/metadata from files, per
// Task32 phase 5's option B) and returns everything a test needs to build a
// matching Config.
type fixture struct {
	corpusPath, metaPath string
	corpusHash, metaHash string
}

func writeFixture(t *testing.T) fixture {
	t.Helper()
	dir := t.TempDir()
	vocab := []string{"qokedy", "qokeedy", "chedy", "shedy", "daiin", "aiin", "otedy", "qokaiin"}
	var tokens []string
	var currier, hand []string
	appendBlock := func(c, h string, n, offset int) {
		for i := range n {
			tokens = append(tokens, vocab[(i+offset)%len(vocab)])
			currier, hand = append(currier, c), append(hand, h)
		}
	}
	appendBlock("A", "1", 300, 0)
	appendBlock("B", "2", 300, 3)

	corpusPath := dir + "/corpus.txt"
	if err := os.WriteFile(corpusPath, []byte(strings.Join(tokens, "\n")+"\n"), 0644); err != nil {
		t.Fatalf("write corpus fixture: %v", err)
	}
	var meta strings.Builder
	meta.WriteString("token_position\tcurrier\thand\n")
	for i := range tokens {
		fmt.Fprintf(&meta, "%d\t%s\t%s\n", i, currier[i], hand[i])
	}
	metaPath := dir + "/metadata.tsv"
	if err := os.WriteFile(metaPath, []byte(meta.String()), 0644); err != nil {
		t.Fatalf("write metadata fixture: %v", err)
	}

	gotTokens, corpusHash, err := readCorpus(corpusPath)
	if err != nil || len(gotTokens) != len(tokens) {
		t.Fatalf("sanity readCorpus: %v (%d tokens)", err, len(gotTokens))
	}
	_, _, metaHash, err := loadTokenLabels(metaPath)
	if err != nil {
		t.Fatalf("sanity loadTokenLabels: %v", err)
	}
	return fixture{corpusPath: corpusPath, metaPath: metaPath, corpusHash: corpusHash, metaHash: metaHash}
}

// smallConfig is the reduced scientific parameter set every test in this
// file uses: small enough to run in milliseconds, large enough to produce at
// least one eligible class in every scheme and a non-empty Part A/B job set.
func (f fixture) smallConfig() Config {
	return Config{
		CorpusPath: f.corpusPath, TokenMetadataMap: f.metaPath,
		WindowSizes: []int{50}, ResidualWindowSizes: []int{50},
		MinClassTokens: 100, MinBlockTokens: 50,
		KMin: 2, KMaxWithin: 3, KMaxResidual: 3,
		Permutations: 3, Seed: 7,
	}
}

func (f fixture) initMessage(t *testing.T, c Config) protocolMessage {
	t.Helper()
	return protocolMessage{
		Kind: "init", Version: workerProtocolVersion,
		Fingerprint:         computeFingerprint(c, f.corpusHash, f.metaHash),
		CorpusPath:          f.corpusPath,
		TokenMetadataMap:    f.metaPath,
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
}

// TestProcessPoolMatchesLocalComputation is the core Task32 claim at the
// subprocess-worker granularity: a job dispatched to a real child process
// must return the exact bit-for-bit value the same job would produce if
// computed locally by the goroutine backend's own closures, for every job
// shape (Part A significance/refinement, Part B correction).
func TestProcessPoolMatchesLocalComputation(t *testing.T) {
	f := writeFixture(t)
	c := f.smallConfig()
	tokens, _, err := readCorpus(f.corpusPath)
	if err != nil {
		t.Fatal(err)
	}
	currier, hand, _, err := loadTokenLabels(f.metaPath)
	if err != nil {
		t.Fatal(err)
	}
	allBlocks := buildAllBlocks(currier, hand)
	blocksByScheme := map[Scheme]map[ClassID][]Block{
		SchemeJoint: blocksByClass(allBlocks[SchemeJoint]),
	}
	classA := ClassID{Scheme: SchemeJoint, Currier: "A", Hand: "1"}
	blocksA := blocksByScheme[SchemeJoint][classA]
	rows, _ := withinClassSweep(tokens, classA, blocksA, 50, c.KMin, c.KMaxWithin, c.Seed)
	best := bestByMethod(rows)
	if _, ok := best["k_medoids"]; !ok {
		t.Fatal("fixture did not produce an observed k_medoids row - adjust fixture sizing")
	}

	pool, err := newProcessPool(2, testWorkerCmd("worker"), f.initMessage(t, c))
	if err != nil {
		t.Fatalf("newProcessPool: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()

	got, err := pool.Run(ctx, JobID{Stage: "part_a_significance", Combination: "joint|A/1|50|k_medoids", ReplicateIndex: 2})
	if err != nil {
		t.Fatalf("part_a_significance job: %v", err)
	}
	rng := rand.New(rand.NewSource(replicateSeed(c.Seed, methodSalt("k_medoids"), 2)))
	want := nullSilhouetteAtK(tokens, blocksA, 50, "k_medoids", best["k_medoids"].K, rng)
	if got != want {
		t.Fatalf("part_a_significance: subprocess=%v local=%v", got, want)
	}

	got, err = pool.Run(ctx, JobID{Stage: "part_a_refinement", Combination: "joint|A/1|50|k_medoids", ReplicateIndex: 1})
	if err != nil {
		t.Fatalf("part_a_refinement job: %v", err)
	}
	rng = rand.New(rand.NewSource(replicateSeed(c.Seed+999983, methodSalt("k_medoids")+100, 1)))
	want = nullSilhouetteAtK(tokens, blocksA, 50, "k_medoids", best["k_medoids"].K, rng)
	if got != want {
		t.Fatalf("part_a_refinement: subprocess=%v local=%v", got, want)
	}

	classB := ClassID{Scheme: SchemeJoint, Currier: "B", Hand: "2"}
	jointEligible := []ClassID{classA, classB}
	jointBlocks := map[ClassID][]Block{classA: blocksA, classB: blocksByClass(allBlocks[SchemeJoint])[classB]}
	got, err = pool.Run(ctx, JobID{Stage: "part_b_global_correction", Combination: "hierarchical|raw", ReplicateIndex: 1})
	if err != nil {
		t.Fatalf("part_b_global_correction job: %v", err)
	}
	rng = rand.New(rand.NewSource(replicateSeed(c.Seed+1, methodSalt("hierarchical"), 1)))
	want = residualNullMax(tokens, jointEligible, jointBlocks, c.ResidualWindowSizes, c.KMin, c.KMaxResidual, "hierarchical", false, rng)
	if got != want {
		t.Fatalf("part_b_global_correction: subprocess=%v local=%v", got, want)
	}
}

func TestProcessPoolWorkerCrashDiagnosesFailedJob(t *testing.T) {
	f := writeFixture(t)
	c := f.smallConfig()
	pool, err := newProcessPool(1, testWorkerCmd("crash"), f.initMessage(t, c))
	if err != nil {
		t.Fatalf("newProcessPool: %v", err)
	}
	defer pool.Close()

	_, err = pool.Run(context.Background(), JobID{Stage: "part_b_global_correction", Combination: "k_medoids|raw", ReplicateIndex: 0})
	if err == nil {
		t.Fatal("expected an error from a worker that exited without a result")
	}
	if !strings.Contains(err.Error(), "worker 0") {
		t.Fatalf("diagnostic error must identify the failed worker: %v", err)
	}
}

func TestProcessPoolCloseWaitsOutEveryWorker(t *testing.T) {
	f := writeFixture(t)
	c := f.smallConfig()
	pool, err := newProcessPool(3, testWorkerCmd("worker"), f.initMessage(t, c))
	if err != nil {
		t.Fatalf("newProcessPool: %v", err)
	}
	if err := pool.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	for i, w := range pool.workers {
		if w.cmd.ProcessState == nil {
			t.Fatalf("worker %d: Close returned without Wait completing (would leave a zombie)", i)
		}
	}
	if err := pool.Close(); err != nil {
		t.Fatalf("Close must be idempotent: %v", err)
	}
}

func TestNewProcessPoolStartupFailureLeavesNoRunningWorkers(t *testing.T) {
	f := writeFixture(t)
	c := f.smallConfig()
	badInit := f.initMessage(t, c)
	badInit.Fingerprint = "not-the-real-fingerprint"
	_, err := newProcessPool(3, testWorkerCmd("worker"), badInit)
	if err == nil {
		t.Fatal("expected handshake rejection to fail pool startup")
	}
	if !strings.Contains(err.Error(), "rejected handshake") {
		t.Fatalf("expected an explicit handshake rejection, got: %v", err)
	}
}
