package structuralprojection

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// TrialWorker owns the invariant state shared by all projection trials in one
// process. Run is serialized because core.go intentionally reuses package-level
// scratch buffers; parallelism is across worker processes, never goroutines in
// one address space.
type TrialWorker struct {
	c        Config
	prof     profiles
	full     Projection
	future   Projection
	fb       frequencyBins
	selected []pair
}

// The scratch buffers in core.go belong to the package, not to an individual
// TrialWorker. This lock therefore also covers the unusual case of multiple
// worker states hosted in one process (tests/embedding); production remote
// parallelism remains process-isolated.
var trialScratchMu sync.Mutex

func selectedPairs(c Config, corp corpus, prev []previousPair, fams []family) ([]pair, error) {
	selected := []pair{}
	seen := map[pair]bool{}
	canon := func(a, b string) pair {
		if b < a {
			a, b = b, a
		}
		return pair{a, b}
	}
	add := func(x pair) {
		x = canon(strings.TrimSpace(x.A), strings.TrimSpace(x.B))
		if !seen[x] {
			seen[x] = true
			selected = append(selected, x)
		}
	}
	if c.Pair != "" {
		x := strings.Split(c.Pair, ",")
		add(pair{x[0], x[1]})
	} else if c.FamilyID > 0 {
		found := false
		for _, f := range fams {
			if f.ID != c.FamilyID {
				continue
			}
			found = true
			for i := range f.Tokens {
				for j := 0; j < i; j++ {
					add(pair{f.Tokens[i], f.Tokens[j]})
				}
			}
		}
		if !found {
			return nil, fmt.Errorf("family %d not found", c.FamilyID)
		}
	} else {
		for _, x := range applicableMandatory(corp.Counts) {
			add(x)
		}
		for _, x := range prev {
			add(pair{x.TokenA, x.TokenB})
		}
	}
	for _, x := range selected {
		if corp.Counts[x.A] == 0 || corp.Counts[x.B] == 0 {
			return nil, fmt.Errorf("selected token absent from corpus: %s/%s", x.A, x.B)
		}
	}
	return selected, nil
}

// NewTrialWorker stages large invariant inputs once per worker process.
func NewTrialWorker(c Config) (*TrialWorker, error) {
	trialScratchMu.Lock()
	defer trialScratchMu.Unlock()
	if err := validate(c); err != nil {
		return nil, err
	}
	corp, err := readCorpus(c.CorpusPath)
	if err != nil {
		return nil, err
	}
	edges, err := readEdges(c.StructuralPairsPath)
	if err != nil {
		return nil, err
	}
	prev, err := readPrevious(c.DistancePairsPath, c.TopN)
	if err != nil {
		return nil, err
	}
	fams, err := readFamilies(c.FamiliesPath)
	if err != nil {
		return nil, err
	}
	selected, err := selectedPairs(c, corp, prev, fams)
	if err != nil {
		return nil, err
	}
	tokens := uniqueTokens(corp)
	return &TrialWorker{c: c, prof: buildProfiles(corp, c.MaxDistance, false),
		full:   BuildProjection(tokens, edges, c.MinStructuralSimilarity, c.MinReliability, c.ProjectionK, "full"),
		future: BuildProjection(tokens, edges, c.MinStructuralSimilarity, c.MinReliability, c.ProjectionK, "future-ablated"),
		fb:     buildFrequencyBins(tokens, corp.Counts), selected: selected}, nil
}

func (w *TrialWorker) Run(ctx context.Context, trial int) (TrialResult, error) {
	if trial < 0 || trial >= w.c.RandomProjections {
		return TrialResult{}, fmt.Errorf("trial index %d out of range", trial)
	}
	if err := ctx.Err(); err != nil {
		return TrialResult{}, err
	}
	trialScratchMu.Lock()
	defer trialScratchMu.Unlock()
	rp := RandomizeProjection(w.full, w.fb, w.c.Seed+int64(trial)*7919)
	sp := GenericSmoothing(w.fb, w.full, w.c.Seed+int64(trial)*104729)
	rap := RandomizeProjection(w.future, w.fb, w.c.Seed+int64(trial)*15485863)
	sap := GenericSmoothing(w.fb, w.future, w.c.Seed+int64(trial)*32452843)
	r := TrialResult{Random: make([]float64, len(w.selected)), Smoothing: make([]float64, len(w.selected)), RandomAblated: make([]float64, len(w.selected)), SmoothingAblated: make([]float64, len(w.selected)), RandomByDistance: make([][]float64, len(w.selected))}
	rcache, scache, racache, sacache := map[string][]map[string]float64{}, map[string][]map[string]float64{}, map[string][]map[string]float64{}, map[string][]map[string]float64{}
	projected := func(t string, proj Projection, cache map[string][]map[string]float64) []map[string]float64 {
		if x := cache[t]; x != nil {
			return x
		}
		x := make([]map[string]float64, min(5, w.c.MaxDistance))
		for d := range x {
			x[d] = ProjectDistribution(w.prof[t].Right[d], proj)
		}
		cache[t] = x
		return x
	}
	for i, x := range w.selected {
		gs, ss, gas, sas := []float64{}, []float64{}, []float64{}, []float64{}
		ra, rb := projected(x.A, rp, rcache), projected(x.B, rp, rcache)
		sa, sb := projected(x.A, sp, scache), projected(x.B, sp, scache)
		raa, rab := projected(x.A, rap, racache), projected(x.B, rap, racache)
		saa, sab := projected(x.A, sap, sacache), projected(x.B, sap, sacache)
		for d := 0; d < min(5, w.c.MaxDistance); d++ {
			tj, _, _ := metricsFloat(countsFloat(w.prof[x.A].Right[d]), countsFloat(w.prof[x.B].Right[d]))
			rj, _, _ := metricsFloat(ra[d], rb[d])
			sj, _, _ := metricsFloat(sa[d], sb[d])
			raj, _, _ := metricsFloat(raa[d], rab[d])
			saj, _, _ := metricsFloat(saa[d], sab[d])
			gs = append(gs, rj-tj)
			ss = append(ss, sj-tj)
			gas = append(gas, raj-tj)
			sas = append(sas, saj-tj)
			r.RandomByDistance[i] = append(r.RandomByDistance[i], rj-tj)
		}
		r.Random[i], r.Smoothing[i] = mean(gs), mean(ss)
		r.RandomAblated[i], r.SmoothingAblated[i] = mean(gas), mean(sas)
	}
	return r, nil
}

func (w *TrialWorker) Close() error { return nil }

// Fingerprint is the scientific identity used by JobID/checkpoint/remote staging.
func Fingerprint(c Config) (string, error) {
	h := sha256.New()
	for _, p := range []string{c.CorpusPath, c.StructuralPairsPath, c.DistancePairsPath, c.FamiliesPath} {
		b, err := os.ReadFile(p)
		if err != nil {
			return "", err
		}
		s := sha256.Sum256(b)
		h.Write(s[:])
	}
	v := struct {
		MinSim, MinRel                  float64
		K, N, MaxD, MinObs, Top, Family int
		Mode, Pair                      string
		Seed                            int64
	}{c.MinStructuralSimilarity, c.MinReliability, c.ProjectionK, c.RandomProjections, c.MaxDistance, c.MinObservations, c.TopN, c.FamilyID, c.ProjectionMode, c.Pair, c.Seed}
	b, _ := json.Marshal(v)
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil)), nil
}

// CanonicalTrialOrder validates and orders indexed results before reduction.
func CanonicalTrialOrder(results map[int]TrialResult, n int) ([]TrialResult, error) {
	out := make([]TrialResult, n)
	for i := range n {
		v, ok := results[i]
		if !ok {
			return nil, fmt.Errorf("missing projection trial %d", i)
		}
		out[i] = v
	}
	return out, nil
}

func validateTrialResult(r TrialResult, pairs, distances int) error {
	if len(r.Random) != pairs || len(r.Smoothing) != pairs || len(r.RandomAblated) != pairs || len(r.SmoothingAblated) != pairs || len(r.RandomByDistance) != pairs {
		return fmt.Errorf("invalid pair vector lengths: want %d", pairs)
	}
	for i := range r.RandomByDistance {
		if len(r.RandomByDistance[i]) != distances {
			return fmt.Errorf("invalid distance vector length for pair %d: got %d, want %d", i, len(r.RandomByDistance[i]), distances)
		}
	}
	return nil
}

func sortedCompleted(m map[int]TrialResult) []int {
	out := make([]int, 0, len(m))
	for i := range m {
		out = append(out, i)
	}
	sort.Ints(out)
	return out
}

func runTrials(c Config, executor TrialExecutor, progress *progressReporter) ([]TrialResult, error) {
	ctx := c.Context
	if ctx == nil {
		ctx = context.Background()
	}
	workers := c.Workers
	if workers < 1 {
		workers = 1
	}
	// The in-process worker deliberately stays serial because package scratch
	// is shared. Process/remote executors provide safe address-space isolation.
	if _, ok := executor.(*TrialWorker); ok {
		workers = 1
	}
	fingerprint, err := Fingerprint(c)
	if err != nil {
		return nil, err
	}
	byIndex := map[int]TrialResult{}
	if c.CheckpointPath != "" && c.CheckpointPath != "-" {
		byIndex, err = loadTrialCheckpoint(c.CheckpointPath, fingerprint)
		if err != nil {
			return nil, err
		}
	}
	ready := make([]bool, c.RandomProjections)
	for i := range byIndex {
		if i < 0 || i >= len(ready) {
			return nil, fmt.Errorf("structural checkpoint contains out-of-range trial %d", i)
		}
		ready[i] = true
	}
	type done struct {
		index  int
		result TrialResult
		err    error
	}
	jobs := make(chan int)
	results := make(chan done, workers)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				r, e := executor.Run(ctx, i)
				results <- done{i, r, e}
				if e != nil {
					return
				}
			}
		}()
	}
	go func() {
		defer close(jobs)
		for i := range c.RandomProjections {
			if ready[i] {
				continue
			}
			select {
			case jobs <- i:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() { wg.Wait(); close(results) }()
	completed := len(byIndex)
	for d := range results {
		if d.err != nil {
			cancel()
			return nil, fmt.Errorf("projection trial %d: %w", d.index, d.err)
		}
		if _, ok := byIndex[d.index]; ok {
			continue
		}
		if d.index < 0 || d.index >= c.RandomProjections {
			cancel()
			return nil, fmt.Errorf("executor returned out-of-range projection trial %d", d.index)
		}
		byIndex[d.index] = d.result
		completed++
		if c.CheckpointPath != "" && c.CheckpointPath != "-" {
			if err := saveTrialCheckpoint(c.CheckpointPath, fingerprint, byIndex); err != nil {
				cancel()
				return nil, err
			}
		}
		active, retries := min(workers, c.RandomProjections-completed), 0
		if s, ok := executor.(TrialExecutorStats); ok {
			active, retries = s.TrialStats()
		}
		progress.trials(completed, c.RandomProjections, active, retries)
	}
	return CanonicalTrialOrder(byIndex, c.RandomProjections)
}

type trialCheckpoint struct {
	Fingerprint string              `json:"fingerprint"`
	Results     map[int]TrialResult `json:"results"`
}

func loadTrialCheckpoint(path, fingerprint string) (map[int]TrialResult, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return map[int]TrialResult{}, nil
	}
	if err != nil {
		return nil, err
	}
	var cp trialCheckpoint
	if err := json.Unmarshal(b, &cp); err != nil {
		return nil, fmt.Errorf("read structural checkpoint: %w", err)
	}
	if cp.Fingerprint != fingerprint {
		return nil, fmt.Errorf("structural checkpoint fingerprint mismatch")
	}
	if cp.Results == nil {
		cp.Results = map[int]TrialResult{}
	}
	return cp.Results, nil
}

func saveTrialCheckpoint(path, fingerprint string, results map[int]TrialResult) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(trialCheckpoint{Fingerprint: fingerprint, Results: results}, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	tmp, err := os.CreateTemp(filepath.Dir(path), ".structural-checkpoint-")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err = tmp.Chmod(0600); err == nil {
		_, err = tmp.Write(b)
	}
	if err == nil {
		err = tmp.Sync()
	}
	if e := tmp.Close(); err == nil {
		err = e
	}
	if err == nil {
		err = os.Rename(name, path)
	}
	if err == nil {
		dir, openErr := os.Open(filepath.Dir(path))
		if openErr != nil {
			err = openErr
		} else {
			err = dir.Sync()
			if closeErr := dir.Close(); err == nil {
				err = closeErr
			}
		}
	}
	return err
}
