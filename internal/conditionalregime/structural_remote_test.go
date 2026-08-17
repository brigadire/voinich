package conditionalregime

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"zcore.dev/voinich/internal/structuralprojection"
)

func structuralRemoteFixture(t *testing.T) structuralprojection.Config {
	t.Helper()
	d := t.TempDir()
	write := func(name, body string) string {
		p := filepath.Join(d, name)
		if err := os.WriteFile(p, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	corpus := strings.Repeat("x a y b x b y a\n", 20)
	header := "token_a\ttoken_b\tcount_a\tcount_b\tposition_similarity\tleft_similarity\tright_similarity\traw_similarity\tposition_reliability\tleft_reliability\tright_reliability\tevidence_strength\n"
	edges := header + "x\ty\t40\t40\t0.9\t0.8\t0.7\t0.8\t1\t1\t1\t1\n"
	return structuralprojection.Config{CorpusPath: write("corpus.txt", corpus), StructuralPairsPath: write("edges.tsv", edges), DistancePairsPath: write("pairs.yaml", "pairs: []\n"), FamiliesPath: write("families.yaml", "families: []\n"), MinStructuralSimilarity: .65, MinReliability: .7, ProjectionMode: "both", RandomProjections: 1, MaxDistance: 2, MinObservations: 1, Pair: "x,y", Seed: 130013, Workers: 1, RemoteTimeout: 2 * time.Second, RemoteRetries: 2}
}

func TestStructuralTrialsTwoRemoteWorkersMatchLocalInAnyCompletionOrder(t *testing.T) {
	c := structuralRemoteFixture(t)
	c.RandomProjections = 3
	local, err := structuralprojection.NewTrialWorker(c)
	if err != nil {
		t.Fatal(err)
	}
	want := make([]structuralprojection.TrialResult, c.RandomProjections)
	for i := range want {
		want[i], err = local.Run(context.Background(), i)
		if err != nil {
			t.Fatal(err)
		}
	}
	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	ex, err := NewStructuralRemoteExecutor(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ex.Close()
	pool := ex.(*structuralExecutorAdapter).pool.(*remotePool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	url := "https://localhost:" + strings.Split(pool.Addr(), ":")[1]
	for _, name := range []string{"structural-worker-a", "structural-worker-b"} {
		crt, key := pkiFix.issueWorker(t, name)
		startRemoteWorker(t, ctx, url, pkiFix.caCrt, crt, key, 4)
	}
	got := make([]structuralprojection.TrialResult, len(want))
	errCh := make(chan error, len(want))
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
		t.Fatalf("multi-worker remote results differ\ngot=%#v\nwant=%#v", got, want)
	}
}

func TestStructuralTrialRemoteMatchesLocalOracle(t *testing.T) {
	c := structuralRemoteFixture(t)
	local, err := structuralprojection.NewTrialWorker(c)
	if err != nil {
		t.Fatal(err)
	}
	want, err := local.Run(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	pkiFix := newRemotePKI(t, []string{"localhost"}, nil)
	workerCrt, workerKey := pkiFix.issueWorker(t, "structural-worker")
	c.RemoteListen = "127.0.0.1:0"
	c.TLSCert, c.TLSKey, c.ClientCA = pkiFix.coordCrt, pkiFix.coordKey, pkiFix.caCrt
	ex, err := NewStructuralRemoteExecutor(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ex.Close()
	adapter := ex.(*structuralExecutorAdapter)
	pool := adapter.pool.(*remotePool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startRemoteWorker(t, ctx, "https://localhost:"+strings.Split(pool.Addr(), ":")[1], pkiFix.caCrt, workerCrt, workerKey, 4)
	got, err := ex.Run(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("remote result differs\ngot=%#v\nwant=%#v", got, want)
	}
}
