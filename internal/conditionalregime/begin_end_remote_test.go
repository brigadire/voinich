package conditionalregime

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
	"zcore.dev/voinich/internal/beginendanalyze"
)

// beginEndRemoteFixture writes a small corpus/dictionary pair (4 eligible
// tokens across several pages, so the flat candidate-pair space has enough
// batches at CandidateBatchSize 1 for the multi-worker/retry/resume tests
// below) and returns the beginendanalyze.Config every test in this file
// starts from.
func beginEndRemoteFixture(t *testing.T) beginendanalyze.Config {
	t.Helper()
	d := t.TempDir()
	corpusText := strings.Repeat("a b c d\nb c d a\nc d a b\nd a b c\n\n", 6)
	corpusPath := filepath.Join(d, "corpus.txt")
	if err := os.WriteFile(corpusPath, []byte(corpusText), 0600); err != nil {
		t.Fatal(err)
	}
	dictionaryPath := filepath.Join(d, "dictionary.yaml")
	if err := os.WriteFile(dictionaryPath, buildBeginEndDictionary(t, corpusText), 0600); err != nil {
		t.Fatal(err)
	}
	return beginendanalyze.Config{
		DictionaryPath: dictionaryPath, CorpusPath: corpusPath,
		MaxWindow: 3, Permutations: 5, MinTokenCount: 1, RandomSeed: 1,
		PermutationMode: "page", MaxCandidates: 20, CandidateBatchSize: 1,
		Workers: 1, RemoteTimeout: 5 * time.Second, RemoteRetries: 2,
	}
}

// buildBeginEndDictionary derives a valid dictionary.yaml from corpusText,
// mirroring dict-gen's own per-token bookkeeping closely enough to satisfy
// loadCorpus/validateCorpusDictionary's consistency checks (exact counts,
// position totals, line-start/end counts) - the scientific content of
// these auxiliary fields does not matter for the local/remote/retry/resume
// identity this file tests, only that the fixture is internally
// consistent.
func buildBeginEndDictionary(t *testing.T, corpusText string) []byte {
	t.Helper()
	type neighbor struct {
		Token string `yaml:"token"`
		Count int    `yaml:"count"`
	}
	type position struct {
		Position int `yaml:"position"`
		Count    int `yaml:"count"`
	}
	type token struct {
		Token            string     `yaml:"token"`
		Count            int        `yaml:"count"`
		PositionInString []position `yaml:"position_in_string"`
		WordBefore       []neighbor `yaml:"word_before"`
		WordAfter        []neighbor `yaml:"word_after"`
		LineStartCount   int        `yaml:"line_start_count"`
		LineEndCount     int        `yaml:"line_end_count"`
	}
	counts := map[string]*token{}
	addNeighbor := func(items *[]neighbor, name string) {
		for i := range *items {
			if (*items)[i].Token == name {
				(*items)[i].Count++
				return
			}
		}
		*items = append(*items, neighbor{Token: name, Count: 1})
	}
	for _, line := range strings.Split(corpusText, "\n") {
		fields := strings.Fields(line)
		for i, tok := range fields {
			item := counts[tok]
			if item == nil {
				item = &token{Token: tok}
				counts[tok] = item
			}
			item.Count++
			found := false
			for j := range item.PositionInString {
				if item.PositionInString[j].Position == i {
					item.PositionInString[j].Count++
					found = true
				}
			}
			if !found {
				item.PositionInString = append(item.PositionInString, position{Position: i, Count: 1})
			}
			if i == 0 {
				item.LineStartCount++
			}
			if i == len(fields)-1 {
				item.LineEndCount++
			}
			if i > 0 {
				addNeighbor(&item.WordBefore, fields[i-1])
			}
			if i+1 < len(fields) {
				addNeighbor(&item.WordAfter, fields[i+1])
			}
		}
	}
	dictionary := make([]token, 0, len(counts))
	for _, item := range counts {
		dictionary = append(dictionary, *item)
	}
	b, err := yaml.Marshal(dictionary)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// beginEndLargePayloadFixture builds a corpus with enough distinct tokens
// that computing every candidate pair in a single batch marshals to well
// over the pre-Task47 1 MiB maxRemoteMessageBytes cap (see remote.go's
// comment on that constant) - a regression fixture for the real bug the
// Astafiev-corpus granularity study caught: a large single JSON message
// silently truncated by decodeJSONBody, which never produced an
// application-level error and instead hung every attempt until
// remote-timeout, four times over, with no result ever returned.
func beginEndLargePayloadFixture(t *testing.T) beginendanalyze.Config {
	t.Helper()
	d := t.TempDir()
	tokens := make([]string, 60)
	for i := range tokens {
		tokens[i] = fmt.Sprintf("tok%02d", i)
	}
	var b strings.Builder
	for page := 0; page < 20; page++ {
		for line := 0; line < 5; line++ {
			for i, tok := range tokens {
				if (i+line+page)%3 != 0 {
					continue
				}
				b.WriteString(tok)
				b.WriteByte(' ')
			}
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}
	corpusText := b.String()
	corpusPath := filepath.Join(d, "corpus.txt")
	if err := os.WriteFile(corpusPath, []byte(corpusText), 0600); err != nil {
		t.Fatal(err)
	}
	dictionaryPath := filepath.Join(d, "dictionary.yaml")
	if err := os.WriteFile(dictionaryPath, buildBeginEndDictionary(t, corpusText), 0600); err != nil {
		t.Fatal(err)
	}
	ws, err := beginendanalyze.LoadForDistribution(beginendanalyze.Config{
		DictionaryPath: dictionaryPath, CorpusPath: corpusPath,
		MaxWindow: 55, Permutations: 5, MinTokenCount: 1, RandomSeed: 1,
		PermutationMode: "page", MaxCandidates: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	total := ws.TotalPairs()
	whole := beginendanalyze.ComputeBatch(ws, 0, total)
	blob, err := json.Marshal(whole)
	if err != nil {
		t.Fatal(err)
	}
	if len(blob) <= 1<<20 {
		t.Fatalf("fixture produces only %d bytes for the whole pair space (%d pairs) - not large enough to exercise the >1 MiB regression; widen the token/line counts", len(blob), total)
	}
	t.Logf("large-payload fixture: %d pairs, whole-batch JSON = %d bytes", total, len(blob))
	return beginendanalyze.Config{
		DictionaryPath: dictionaryPath, CorpusPath: corpusPath,
		MaxWindow: 55, Permutations: 5, MinTokenCount: 1, RandomSeed: 1,
		PermutationMode: "page", MaxCandidates: 20, CandidateBatchSize: total,
		Workers: 1, RemoteTimeout: 30 * time.Second, RemoteRetries: 1,
	}
}

func TestBeginEndRemoteHandlesPayloadOverOldOneMiBCap(t *testing.T) {
	c := beginEndLargePayloadFixture(t)
	want := beginEndOracleBatch(t, c, 0)

	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	workerCrt, workerKey := pkiFix.issueWorker(t, "begin-end-worker-large")
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	ex, err := NewBeginEndRemoteExecutor(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ex.(*beginEndExecutorAdapter).pool.Close()
	pool := ex.(*beginEndExecutorAdapter).pool.(*remotePool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startRemoteWorker(t, ctx, "https://localhost:"+strings.Split(pool.Addr(), ":")[1], pkiFix.caCrt, workerCrt, workerKey, 1)

	got, err := ex.Run(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatal("large remote batch (>1 MiB payload) differs from local oracle")
	}
}

func beginEndOracleBatch(t *testing.T, c beginendanalyze.Config, batchIndex int) beginendanalyze.BatchResult {
	t.Helper()
	ws, err := beginendanalyze.LoadForDistribution(c)
	if err != nil {
		t.Fatal(err)
	}
	return beginendanalyze.ComputeBatch(ws, batchIndex, c.CandidateBatchSize)
}

func beginEndBatchCount(t *testing.T, c beginendanalyze.Config) int {
	t.Helper()
	ws, err := beginendanalyze.LoadForDistribution(c)
	if err != nil {
		t.Fatal(err)
	}
	total := ws.TotalPairs()
	return (total + c.CandidateBatchSize - 1) / c.CandidateBatchSize
}

func TestBeginEndRemoteMatchesLocalOracle(t *testing.T) {
	c := beginEndRemoteFixture(t)
	want := beginEndOracleBatch(t, c, 0)

	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	workerCrt, workerKey := pkiFix.issueWorker(t, "begin-end-worker")
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	ex, err := NewBeginEndRemoteExecutor(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ex.(*beginEndExecutorAdapter).pool.Close()
	pool := ex.(*beginEndExecutorAdapter).pool.(*remotePool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startRemoteWorker(t, ctx, "https://localhost:"+strings.Split(pool.Addr(), ":")[1], pkiFix.caCrt, workerCrt, workerKey, 4)

	got, err := ex.Run(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("remote batch differs\ngot=%#v\nwant=%#v", got, want)
	}
}

func TestBeginEndTwoRemoteWorkersMatchLocalInAnyCompletionOrder(t *testing.T) {
	c := beginEndRemoteFixture(t)
	n := beginEndBatchCount(t, c)
	want := make([]beginendanalyze.BatchResult, n)
	for i := range want {
		want[i] = beginEndOracleBatch(t, c, i)
	}

	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	ex, err := NewBeginEndRemoteExecutor(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ex.(*beginEndExecutorAdapter).pool.Close()
	pool := ex.(*beginEndExecutorAdapter).pool.(*remotePool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	url := "https://localhost:" + strings.Split(pool.Addr(), ":")[1]
	for _, name := range []string{"begin-end-worker-a", "begin-end-worker-b"} {
		crt, key := pkiFix.issueWorker(t, name)
		startRemoteWorker(t, ctx, url, pkiFix.caCrt, crt, key, 4)
	}

	got := make([]beginendanalyze.BatchResult, n)
	errCh := make(chan error, n)
	var wg sync.WaitGroup
	for i := range got {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var runErr error
			got[i], runErr = ex.Run(context.Background(), i)
			errCh <- runErr
		}(i)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("multi-worker remote batches differ from local oracle in canonical order\ngot=%#v\nwant=%#v", got, want)
	}
}

func TestBeginEndRemoteRetryAfterWorkerFailure(t *testing.T) {
	c := beginEndRemoteFixture(t)
	c.RemoteTimeout = 150 * time.Millisecond
	want := beginEndOracleBatch(t, c, 1)

	pkiFix := newRemotePKI(t, []string{"localhost"}, []string{"127.0.0.1"})
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	fp, err := beginendanalyze.Fingerprint(c)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := newBeginEndRemotePool(c, fp)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	base := "https://" + pool.Addr()

	id := JobID{Stage: "begin_end_candidate_batch", Combination: "candidates", ReplicateIndex: 1}
	resultCh := make(chan struct {
		blob []byte
		err  error
	}, 1)
	go func() {
		b, runErr := pool.RunBlob(context.Background(), id)
		resultCh <- struct {
			blob []byte
			err  error
		}{b, runErr}
	}()

	crashCrt, crashKey := pkiFix.issueWorker(t, "begin-end-worker-crash")
	crashClient := workerHTTPClient(t, crashCrt, crashKey, pkiFix.caCrt)
	leased := leaseUntilAssigned(t, crashClient, base, fp)
	if leased.JobID != id {
		t.Fatalf("crashing worker leased %+v, want %+v", leased.JobID, id)
	}

	goodCrt, goodKey := pkiFix.issueWorker(t, "begin-end-worker-good")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startRemoteWorker(t, ctx, base, pkiFix.caCrt, goodCrt, goodKey, 1)

	select {
	case res := <-resultCh:
		if res.err != nil {
			t.Fatalf("reassigned job failed: %v", res.err)
		}
		var wire wireBeginEndBatchResult
		if err := json.Unmarshal(res.blob, &wire); err != nil {
			t.Fatal(err)
		}
		got := wire.decode()
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("retried remote batch differs from local oracle\ngot=%#v\nwant=%#v", got, want)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for the job to be reassigned and completed")
	}
	if pool.leasesReclaimed.Load() == 0 {
		t.Fatal("expected at least one reclaimed lease from the crashing worker")
	}
}

func TestBeginEndRemoteSameSeedRepeated(t *testing.T) {
	c := beginEndRemoteFixture(t)
	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	workerCrt, workerKey := pkiFix.issueWorker(t, "begin-end-worker")
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	ex, err := NewBeginEndRemoteExecutor(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ex.(*beginEndExecutorAdapter).pool.Close()
	pool := ex.(*beginEndExecutorAdapter).pool.(*remotePool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startRemoteWorker(t, ctx, "https://localhost:"+strings.Split(pool.Addr(), ":")[1], pkiFix.caCrt, workerCrt, workerKey, 4)

	first, err := ex.Run(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ex.Run(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("repeated same-seed remote run diverged\nfirst=%#v\nsecond=%#v", first, second)
	}
}

func TestBeginEndRemoteInterruptedThenResumedMatchesOracle(t *testing.T) {
	c := beginEndRemoteFixture(t)
	n := beginEndBatchCount(t, c)
	if n < 2 {
		t.Fatalf("fixture too small for an interrupted/resumed test: %d batches", n)
	}
	half := n / 2
	want := make([]beginendanalyze.BatchResult, n)
	for i := range want {
		want[i] = beginEndOracleBatch(t, c, i)
	}

	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt

	runRange := func(lo, hi int, workerID string) []beginendanalyze.BatchResult {
		ex, err := NewBeginEndRemoteExecutor(c)
		if err != nil {
			t.Fatal(err)
		}
		pool := ex.(*beginEndExecutorAdapter).pool.(*remotePool)
		ctx, cancel := context.WithCancel(context.Background())
		crt, key := pkiFix.issueWorker(t, workerID)
		startRemoteWorker(t, ctx, "https://localhost:"+strings.Split(pool.Addr(), ":")[1], pkiFix.caCrt, crt, key, 2)
		out := make([]beginendanalyze.BatchResult, hi-lo)
		for i := lo; i < hi; i++ {
			r, err := ex.Run(context.Background(), i)
			if err != nil {
				t.Fatal(err)
			}
			out[i-lo] = r
		}
		cancel()
		_ = ex.(*beginEndExecutorAdapter).pool.Close()
		return out
	}

	firstHalf := runRange(0, half, "begin-end-resume-worker-1")
	secondHalf := runRange(half, n, "begin-end-resume-worker-2")
	got := append(append([]beginendanalyze.BatchResult{}, firstHalf...), secondHalf...)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("interrupted-then-resumed remote run differs from an uninterrupted oracle\ngot=%#v\nwant=%#v", got, want)
	}
}
