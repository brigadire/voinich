package tokenrepetition

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// TriangularWeights reproduces task46's "triangular-v1" selection-weight
// formula (weight(k) = (H-k)/(H*(H+1)/2) for k=0..H-1, strictly
// decreasing, normalized to 1). It is unexported in
// internal/corpustransform, so this is a documented duplicate of a fixed
// two-line constant formula, not a second implementation of an estimand.
func TriangularWeights(h int) []float64 {
	denom := float64(h*(h+1)) / 2
	w := make([]float64, h)
	for k := 0; k < h; k++ {
		w[k] = float64(h-k) / denom
	}
	return w
}

// UniformWeights returns H equal selection probabilities.
func UniformWeights(h int) []float64 {
	w := make([]float64, h)
	for k := range w {
		w[k] = 1.0 / float64(h)
	}
	return w
}

// TheoreticalRunSurvival is task60 section 11/14's prediction: the
// probability that a plaintext exact run of length k survives as an
// exact ciphertext run, given independent per-occurrence homophone
// selection with probabilities weights (sum_j p_j^k; equals (1/H)^(k-1)
// when weights is uniform).
func TheoreticalRunSurvival(k int, weights []float64) float64 {
	if k < 1 {
		return 1
	}
	sum := 0.0
	for _, p := range weights {
		sum += pow(p, k)
	}
	return sum
}

func pow(p float64, k int) float64 {
	r := 1.0
	for i := 0; i < k; i++ {
		r *= p
	}
	return r
}

// AllocationEntry is one row of a Task55 homophone_allocation.tsv
// sidecar.
type AllocationEntry struct {
	PlaintextToken string
	RawFrequency   int
	AllocatedH     int
}

// LoadAllocation reads a Task55 <cipher>.homophone_allocation.tsv file
// (task60 section 14: frequency-v1's per-token H, never a substituted
// Hmax).
func LoadAllocation(path string) (map[string]AllocationEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string]AllocationEntry{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	header := true
	for sc.Scan() {
		if header {
			header = false
			continue
		}
		cols := strings.Split(sc.Text(), "\t")
		if len(cols) < 5 {
			continue
		}
		freq, _ := strconv.Atoi(cols[1])
		h, _ := strconv.Atoi(cols[4])
		out[cols[0]] = AllocationEntry{PlaintextToken: cols[0], RawFrequency: freq, AllocatedH: h}
	}
	return out, sc.Err()
}

// RunSurvivalRow is one plaintext run's fate under a homophonic
// transformation: theoretical predicted survival probability vs the
// directly observed outcome (did the aligned ciphertext span collapse
// into one exact run of the same length?).
type RunSurvivalRow struct {
	Token          string
	RunLength      int
	H              int
	Predicted      float64
	Survived       bool
}

// RunSurvivalDoseResponse walks every exact run in plainTokens and checks
// the aligned span in cipherTokens (task46/55 preserve N and occurrence
// order 1:1, so position i in plaintext and ciphertext refer to the same
// occurrence) for exact survival, alongside the theoretical prediction
// from hOf/weightsOf (task60 sections 11/13/14). plainTokens and
// cipherTokens must have equal length.
func RunSurvivalDoseResponse(plainTokens, cipherTokens []string, plainLineOfToken []int, hOf func(token string) int, weightsOf func(h int) []float64) []RunSurvivalRow {
	if len(plainTokens) != len(cipherTokens) {
		return nil
	}
	runs := ExactRuns(plainTokens, plainLineOfToken)
	rows := make([]RunSurvivalRow, 0, len(runs))
	for _, r := range runs {
		h := hOf(r.Token)
		predicted := TheoreticalRunSurvival(r.RunLength, weightsOf(h))
		survived := true
		first := cipherTokens[r.StartPosition]
		for p := r.StartPosition; p < r.StartPosition+r.RunLength; p++ {
			if cipherTokens[p] != first {
				survived = false
				break
			}
		}
		rows = append(rows, RunSurvivalRow{Token: r.Token, RunLength: r.RunLength, H: h, Predicted: predicted, Survived: survived})
	}
	return rows
}

// AggregateSurvivalByLength groups RunSurvivalDoseResponse rows by run
// length k: mean predicted probability vs observed survival fraction.
type SurvivalAggregate struct {
	RunLength        int
	Count            int
	MeanPredicted    float64
	ObservedFraction float64
}

func AggregateSurvivalByLength(rows []RunSurvivalRow) []SurvivalAggregate {
	byK := map[int][]RunSurvivalRow{}
	for _, r := range rows {
		byK[r.RunLength] = append(byK[r.RunLength], r)
	}
	var ks []int
	for k := range byK {
		ks = append(ks, k)
	}
	sortInts(ks)
	out := make([]SurvivalAggregate, 0, len(ks))
	for _, k := range ks {
		grp := byK[k]
		sumP, survived := 0.0, 0
		for _, r := range grp {
			sumP += r.Predicted
			if r.Survived {
				survived++
			}
		}
		out = append(out, SurvivalAggregate{RunLength: k, Count: len(grp), MeanPredicted: sumP / float64(len(grp)), ObservedFraction: float64(survived) / float64(len(grp))})
	}
	return out
}

func sortInts(a []int) {
	for i := 1; i < len(a); i++ {
		for j := i; j > 0 && a[j-1] > a[j]; j-- {
			a[j-1], a[j] = a[j], a[j-1]
		}
	}
}
