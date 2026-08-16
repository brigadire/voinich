package globalregime

import (
	"math"
	"math/rand"
	"sort"
)

func distanceMatrix(w []Window) [][]float64 {
	sorted := make([]sortedProfile, len(w))
	for i := range w {
		sorted[i] = sortProfile(w[i].distribution)
	}
	d := make([][]float64, len(w))
	for i := range d {
		d[i] = make([]float64, len(w))
		for j := 0; j < i; j++ {
			v := jsDistanceSorted(sorted[i], sorted[j])
			d[i][j], d[j][i] = v, v
		}
	}
	return d
}

const maxClusterFitWindows = 200

// clusteringSample spans the complete sequence. Only model fitting and
// quadratic diagnostics are sampled; every original window is assigned below.
func clusteringSample(w []Window) ([]Window, []int) {
	if len(w) <= maxClusterFitWindows {
		indices := make([]int, len(w))
		for i := range indices {
			indices[i] = i
		}
		return w, indices
	}
	indices := make([]int, maxClusterFitWindows)
	out := make([]Window, maxClusterFitWindows)
	for i := range indices {
		indices[i] = i * (len(w) - 1) / (maxClusterFitWindows - 1)
		out[i] = w[indices[i]]
	}
	return out, indices
}

func expandLabels(w, sample []Window, sampleLabels []int, k int) []int {
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
	sortedCentroids := make([]sortedProfile, k)
	for c := range centroids {
		if counts[c] == 0 {
			continue
		}
		sortedCentroids[c] = sortProfile(centroids[c])
	}
	labels := make([]int, len(w))
	for i := range w {
		sw := sortProfile(w[i].distribution)
		labels[i] = 0
		bestDistance := jsDistanceSorted(sw, sortedCentroids[0])
		for c := 1; c < k; c++ {
			if counts[c] == 0 {
				continue
			}
			distance := jsDistanceSorted(sw, sortedCentroids[c])
			if distance < bestDistance {
				labels[i] = c
				bestDistance = distance
			}
		}
	}
	return labels
}

func withFullAssignments(d ClusterDiagnostic, labels []int) ClusterDiagnostic {
	d.labels = labels
	d.ClusterSizes = make([]int, d.K)
	for _, c := range labels {
		if c >= 0 && c < d.K {
			d.ClusterSizes[c]++
		}
	}
	d.TransitionCount = 0
	for i := 1; i < len(labels); i++ {
		if labels[i] != labels[i-1] {
			d.TransitionCount++
		}
	}
	d.Fragmentation = float64(d.TransitionCount+1) / float64(max(1, d.K))
	return d
}

type edge struct {
	a, b int
	d    float64
}

func mstEdges(d [][]float64) []edge {
	n := len(d)
	if n == 0 {
		return nil
	}
	used := make([]bool, n)
	best := make([]float64, n)
	parent := make([]int, n)
	for i := range best {
		best[i] = math.Inf(1)
		parent[i] = -1
	}
	best[0] = 0
	out := make([]edge, 0, n-1)
	for z := 0; z < n; z++ {
		v := -1
		for i := 0; i < n; i++ {
			if !used[i] && (v < 0 || best[i] < best[v]) {
				v = i
			}
		}
		if v < 0 {
			break
		}
		used[v] = true
		if parent[v] >= 0 {
			out = append(out, edge{parent[v], v, best[v]})
		}
		for u := 0; u < n; u++ {
			if !used[u] && d[v][u] < best[u] {
				best[u], parent[u] = d[v][u], v
			}
		}
	}
	return out
}

type dsu struct{ p []int }

func newDSU(n int) *dsu {
	p := make([]int, n)
	for i := range p {
		p[i] = i
	}
	return &dsu{p}
}
func (q *dsu) find(a int) int {
	for q.p[a] != a {
		q.p[a] = q.p[q.p[a]]
		a = q.p[a]
	}
	return a
}
func (q *dsu) union(a, b int) {
	a, b = q.find(a), q.find(b)
	if a != b {
		q.p[b] = a
	}
}
func hierarchicalLabels(n, k int, edges []edge) []int {
	q := newDSU(n)
	x := append([]edge(nil), edges...)
	sort.Slice(x, func(i, j int) bool { return x[i].d < x[j].d })
	for i := 0; i < n-k && i < len(x); i++ {
		q.union(x[i].a, x[i].b)
	}
	ids := map[int]int{}
	labels := make([]int, n)
	for i := range labels {
		r := q.find(i)
		if _, ok := ids[r]; !ok {
			ids[r] = len(ids)
		}
		labels[i] = ids[r]
	}
	return labels
}

func kMedoids(d [][]float64, k int, seed int64) []int {
	n := len(d)
	med := make([]int, 0, k)
	r := rand.New(rand.NewSource(seed + int64(k)*7919))
	med = append(med, r.Intn(n))
	for len(med) < k {
		best, far := -1, -1.
		for i := 0; i < n; i++ {
			nearest := math.Inf(1)
			for _, m := range med {
				if d[i][m] < nearest {
					nearest = d[i][m]
				}
			}
			if nearest > far {
				best, far = i, nearest
			}
		}
		med = append(med, best)
	}
	labels := make([]int, n)
	for iteration := 0; iteration < 30; iteration++ {
		for i := 0; i < n; i++ {
			labels[i] = 0
			for c := 1; c < k; c++ {
				if d[i][med[c]] < d[i][med[labels[i]]] {
					labels[i] = c
				}
			}
		}
		changed := false
		for c := 0; c < k; c++ {
			best, bestSum := med[c], math.Inf(1)
			for candidate := 0; candidate < n; candidate++ {
				if labels[candidate] != c {
					continue
				}
				s := 0.
				for i := 0; i < n; i++ {
					if labels[i] == c {
						s += d[candidate][i]
					}
				}
				if s < bestSum {
					best, bestSum = candidate, s
				}
			}
			if best != med[c] {
				med[c] = best
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	for i := 0; i < n; i++ {
		labels[i] = 0
		for c := 1; c < k; c++ {
			if d[i][med[c]] < d[i][med[labels[i]]] {
				labels[i] = c
			}
		}
	}
	return labels
}

func diagnostics(size int, method string, k int, labels []int, d [][]float64) ClusterDiagnostic {
	n := len(labels)
	sizes := make([]int, k)
	for _, c := range labels {
		if c >= 0 && c < k {
			sizes[c]++
		}
	}
	sil := 0.
	within := 0.
	wn := 0
	between := 0.
	bn := 0
	for i := 0; i < n; i++ {
		a, ac := 0., 0
		bs := make([]float64, k)
		bc := make([]int, k)
		for j := 0; j < n; j++ {
			if i == j {
				continue
			}
			if labels[j] == labels[i] {
				a += d[i][j]
				ac++
			} else {
				bs[labels[j]] += d[i][j]
				bc[labels[j]]++
				between += d[i][j]
				bn++
			}
		}
		if ac > 0 {
			a /= float64(ac)
			within += a
			wn++
		}
		b := math.Inf(1)
		for c := 0; c < k; c++ {
			if c != labels[i] && bc[c] > 0 {
				v := bs[c] / float64(bc[c])
				if v < b {
					b = v
				}
			}
		}
		den := math.Max(a, b)
		if den > 0 && !math.IsInf(b, 1) {
			sil += (b - a) / den
		}
	}
	trans := 0
	for i := 1; i < n; i++ {
		if labels[i] != labels[i-1] {
			trans++
		}
	}
	frag := float64(trans+1) / float64(max(1, k))
	if wn > 0 {
		within /= float64(wn)
	}
	if bn > 0 {
		between /= float64(bn)
	}
	if n > 0 {
		sil /= float64(n)
	}
	return ClusterDiagnostic{size, method, k, sil, within, between, sizes, trans, frag, labels}
}

func clusterSweep(w []Window, seed int64, progress func(int, int)) []ClusterDiagnostic {
	if len(w) < 2 {
		return nil
	}
	d := distanceMatrix(w)
	edges := mstEdges(d)
	maxK := min(15, len(w))
	total := (maxK - 1) * 3
	done := 0
	var out []ClusterDiagnostic
	for k := 2; k <= maxK; k++ {
		labels := hierarchicalLabels(len(w), k, edges)
		out = append(out, diagnostics(w[0].WindowSize, "hierarchical", k, labels, d))
		done++
		if progress != nil {
			progress(done, total)
		}
		labels = kMedoids(d, k, seed)
		out = append(out, diagnostics(w[0].WindowSize, "k_medoids", k, labels, d))
		done++
		if progress != nil {
			progress(done, total)
		}
		labels = binarySegments(w, k)
		out = append(out, diagnostics(w[0].WindowSize, "contiguous_segmentation", k, labels, d))
		done++
		if progress != nil {
			progress(done, total)
		}
	}
	return out
}
