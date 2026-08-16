package globalregime

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
)

// referenceDistanceMatrix is distanceMatrix exactly as it stood before the
// sortedProfile/jsDistanceSorted rewrite: it calls jsDistance (which
// re-sorts the union of its two profiles' keys) on every pairwise call.
func referenceDistanceMatrix(w []Window) [][]float64 {
	d := make([][]float64, len(w))
	for i := range d {
		d[i] = make([]float64, len(w))
		for j := 0; j < i; j++ {
			v := jsDistance(w[i].distribution, w[j].distribution)
			d[i][j], d[j][i] = v, v
		}
	}
	return d
}

// referenceExpandLabels is expandLabels exactly as it stood before the
// sortedProfile/jsDistanceSorted rewrite.
func referenceExpandLabels(w, sample []Window, sampleLabels []int, k int) []int {
	centroids := make([]profile, k)
	counts := make([]int, k)
	for c := range centroids {
		centroids[c] = profile{}
	}
	for i, label := range sampleLabels {
		counts[label]++
		for token, value := range sample[i].distribution {
			centroids[label][token] += value
		}
	}
	for c := range centroids {
		if counts[c] == 0 {
			continue
		}
		for token := range centroids[c] {
			centroids[c][token] /= float64(counts[c])
		}
	}
	labels := make([]int, len(w))
	for i := range w {
		labels[i] = 0
		bestDistance := jsDistance(w[i].distribution, centroids[0])
		for c := 1; c < k; c++ {
			if counts[c] == 0 {
				continue
			}
			distance := jsDistance(w[i].distribution, centroids[c])
			if distance < bestDistance {
				labels[i] = c
				bestDistance = distance
			}
		}
	}
	return labels
}

// fixtureWindows builds n synthetic Windows with varied, overlapping but
// not identical sparse token-distribution profiles.
func fixtureWindows(n int, seed int64) []Window {
	r := rand.New(rand.NewSource(seed))
	out := make([]Window, n)
	for i := range out {
		p := profile{}
		for t := 0; t < 8; t++ {
			tok := fmt.Sprintf("tok%d", (i+t)%12)
			p[tok] += r.Float64()
		}
		total := 0.0
		for _, v := range p {
			total += v
		}
		for k := range p {
			p[k] /= total
		}
		out[i] = Window{WindowSize: 100, Index: i, distribution: p}
	}
	return out
}

func TestJSDistanceSortedMatchesReference(t *testing.T) {
	sizes := []int{1, 5, 50, 300}
	keeps := []float64{0.1, 0.5, 1.0}
	for _, n := range sizes {
		for _, keep := range keeps {
			r := rand.New(rand.NewSource(int64(n)*1000 + int64(keep*10)))
			a, b := profile{}, profile{}
			for i := 0; i < n; i++ {
				if r.Float64() < keep {
					a[fmt.Sprintf("t%03d", i)] = r.Float64()
				}
				if r.Float64() < keep {
					b[fmt.Sprintf("t%03d", i)] = r.Float64()
				}
			}
			want := jsDistance(a, b)
			got := jsDistanceSorted(sortProfile(a), sortProfile(b))
			if math.Float64bits(got) != math.Float64bits(want) {
				t.Fatalf("n=%d keep=%v: got %v want %v", n, keep, got, want)
			}
		}
	}
}

// TestDistanceMatrixHoistMatchesReference proves the sortedProfile rewrite
// produces byte-identical output to the pre-rewrite reference.
func TestDistanceMatrixHoistMatchesReference(t *testing.T) {
	for _, n := range []int{0, 1, 2, 20, 80} {
		w := fixtureWindows(n, int64(n)*97+11)
		want := referenceDistanceMatrix(w)
		got := distanceMatrix(w)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("n=%d: distanceMatrix diverged", n)
		}
	}
}

// TestExpandLabelsHoistMatchesReference proves the sortedProfile rewrite
// produces byte-identical output to the pre-rewrite reference, across
// several K values and sample/full window splits.
func TestExpandLabelsHoistMatchesReference(t *testing.T) {
	w := fixtureWindows(60, 7)
	sample, sampleIdx := clusteringSample(w)
	for k := 2; k <= 6; k++ {
		r := rand.New(rand.NewSource(int64(k)))
		sampleLabels := make([]int, len(sample))
		for i := range sampleLabels {
			sampleLabels[i] = r.Intn(k)
		}
		want := referenceExpandLabels(w, sample, sampleLabels, k)
		got := expandLabels(w, sample, sampleLabels, k)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("k=%d: expandLabels diverged\ngot=%v\nwant=%v", k, got, want)
		}
	}
	_ = sampleIdx
}

func benchWindowsFixture() []Window {
	return fixtureWindows(400, 42)
}

func BenchmarkDistanceMatrixReference(b *testing.B) {
	w := benchWindowsFixture()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		referenceDistanceMatrix(w)
	}
}

func BenchmarkDistanceMatrixHoisted(b *testing.B) {
	w := benchWindowsFixture()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		distanceMatrix(w)
	}
}

func BenchmarkExpandLabelsReference(b *testing.B) {
	w := benchWindowsFixture()
	sample, _ := clusteringSample(w)
	sampleLabels := make([]int, len(sample))
	r := rand.New(rand.NewSource(1))
	for i := range sampleLabels {
		sampleLabels[i] = r.Intn(5)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		referenceExpandLabels(w, sample, sampleLabels, 5)
	}
}

func BenchmarkExpandLabelsHoisted(b *testing.B) {
	w := benchWindowsFixture()
	sample, _ := clusteringSample(w)
	sampleLabels := make([]int, len(sample))
	r := rand.New(rand.NewSource(1))
	for i := range sampleLabels {
		sampleLabels[i] = r.Intn(5)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		expandLabels(w, sample, sampleLabels, 5)
	}
}
