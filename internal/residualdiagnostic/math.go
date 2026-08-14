package residualdiagnostic

import (
	"math"
	"math/rand"
	"sort"

	"zcore.dev/voinich/internal/globalregime"
	"zcore.dev/voinich/internal/metadatavalidation"
)

func dot(a, b sparse) float64 {
	if len(a) > len(b) {
		a, b = b, a
	}
	s := 0.
	for x, v := range a {
		s += v * b[x]
	}
	return s
}
func normOf(v sparse) norms {
	n := norms{}
	for _, x := range v {
		a := math.Abs(x)
		n.L1 += a
		n.L2 += x * x
		if a > n.Linf {
			n.Linf = a
		}
	}
	n.L2 = math.Sqrt(n.L2)
	if len(v) > 0 {
		n.MeanAbs = n.L1 / float64(len(v))
	}
	return n
}
func meanSparse(vs []sparse) sparse {
	m := sparse{}
	if len(vs) == 0 {
		return m
	}
	for _, v := range vs {
		for x, z := range v {
			m[x] += z
		}
	}
	for x := range m {
		m[x] /= float64(len(vs))
	}
	return m
}
func varianceSparse(vs []sparse) sparse {
	out := sparse{}
	if len(vs) == 0 {
		return out
	}
	m := meanSparse(vs)
	for x, mu := range m {
		for _, v := range vs {
			d := v[x] - mu
			out[x] += d * d
		}
		out[x] /= float64(len(vs))
	}
	return out
}

// whiteningModel applies a deterministic regularized whitening transform.
// Sigma=.9*cov+.1*diag(cov).  Factoring Sigma=D^.5(I+B'B)D^.5 lets us
// compute a valid whitening without materializing an 8363x8363 matrix.
type whiteningModel struct {
	mean, diag sparse
	train      []sparse
	eigVal     []float64
	eigVec     [][]float64
}

func fitWhitening(raw []sparse) whiteningModel {
	m := whiteningModel{mean: meanSparse(raw)}
	centered := make([]sparse, len(raw))
	m.diag = sparse{}
	maxVar := 0.
	for i, x := range raw {
		centered[i] = subtract(x, m.mean)
		for f, v := range centered[i] {
			m.diag[f] += v * v
		}
	}
	for f := range m.diag {
		m.diag[f] /= float64(max(1, len(raw)))
		if m.diag[f] > maxVar {
			maxVar = m.diag[f]
		}
	}
	largest := largestCovarianceEigen(centered)
	if largest < maxVar {
		largest = maxVar
	}
	floor := 1e-6 * largest
	if floor == 0 {
		floor = 1e-12
	}
	for f := range m.mean {
		d := .1 * m.diag[f]
		if d < floor {
			d = floor
		}
		m.diag[f] = d
	}
	m.train = centered
	n := len(raw)
	g := make([][]float64, n)
	scale := math.Sqrt(.9 / float64(max(1, n)))
	for i := range g {
		g[i] = make([]float64, n)
		for j := 0; j <= i; j++ {
			s := 0.
			for f, v := range centered[i] {
				d := m.diag[f]
				if d == 0 {
					d = floor
				}
				s += v * centered[j][f] / d
			}
			s *= scale * scale
			g[i][j], g[j][i] = s, s
		}
	}
	m.eigVal, m.eigVec = jacobi(g)
	return m
}

func largestCovarianceEigen(centered []sparse) float64 {
	n := len(centered)
	if n == 0 {
		return 0
	}
	g := make([][]float64, n)
	for i := range g {
		g[i] = make([]float64, n)
		for j := 0; j <= i; j++ {
			z := dot(centered[i], centered[j]) / float64(n)
			g[i][j], g[j][i] = z, z
		}
	}
	v := make([]float64, n)
	for i := range v {
		v[i] = 1 / math.Sqrt(float64(n))
	}
	for it := 0; it < 30; it++ {
		next := make([]float64, n)
		norm := 0.
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				next[i] += g[i][j] * v[j]
			}
			norm += next[i] * next[i]
		}
		norm = math.Sqrt(norm)
		if norm == 0 {
			return 0
		}
		for i := range v {
			v[i] = next[i] / norm
		}
	}
	lambda := 0.
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			lambda += v[i] * g[i][j] * v[j]
		}
	}
	return lambda
}

func (m whiteningModel) apply(x sparse) sparse {
	r := subtract(x, m.mean)
	q := sparse{}
	for f, v := range r {
		d := m.diag[f]
		if d == 0 {
			d = 1e-12
		}
		q[f] = v / math.Sqrt(d)
	}
	n := len(m.train)
	if n == 0 {
		return q
	}
	scale := math.Sqrt(.9 / float64(n))
	bq := make([]float64, n)
	for i := range m.train {
		for f, v := range m.train[i] {
			d := m.diag[f]
			if d == 0 {
				d = 1e-12
			}
			bq[i] += scale * v / math.Sqrt(d) * q[f]
		}
	}
	coef := make([]float64, n)
	for e, lam := range m.eigVal {
		if lam < 1e-12 {
			continue
		}
		uDot := 0.
		for i := 0; i < n; i++ {
			uDot += m.eigVec[i][e] * bq[i]
		}
		fac := (1/math.Sqrt(1+lam) - 1) / lam * uDot
		for i := 0; i < n; i++ {
			coef[i] += m.eigVec[i][e] * fac
		}
	}
	for i, c := range coef {
		if c == 0 {
			continue
		}
		for f, v := range m.train[i] {
			d := m.diag[f]
			if d == 0 {
				d = 1e-12
			}
			q[f] += scale * v / math.Sqrt(d) * c
		}
	}
	return q
}

// jacobi returns ascending eigenvalues and column eigenvectors for a real
// symmetric matrix. Fixed pivot order makes whitening bit-reproducible.
func jacobi(a [][]float64) ([]float64, [][]float64) {
	n := len(a)
	v := make([][]float64, n)
	for i := range v {
		v[i] = make([]float64, n)
		v[i][i] = 1
	}
	if n == 0 {
		return nil, v
	}
	for sweep := 0; sweep < 40; sweep++ {
		changed := false
		for p := 0; p < n; p++ {
			for q := p + 1; q < n; q++ {
				apq := a[p][q]
				if math.Abs(apq) < 1e-11 {
					continue
				}
				changed = true
				phi := .5 * math.Atan2(2*apq, a[q][q]-a[p][p])
				c, s := math.Cos(phi), math.Sin(phi)
				app, aqq := a[p][p], a[q][q]
				a[p][p] = c*c*app - 2*s*c*apq + s*s*aqq
				a[q][q] = s*s*app + 2*s*c*apq + c*c*aqq
				a[p][q], a[q][p] = 0, 0
				for k := 0; k < n; k++ {
					if k == p || k == q {
						continue
					}
					akp, akq := a[k][p], a[k][q]
					a[k][p], a[p][k] = c*akp-s*akq, c*akp-s*akq
					a[k][q], a[q][k] = s*akp+c*akq, s*akp+c*akq
				}
				for k := 0; k < n; k++ {
					vp, vq := v[k][p], v[k][q]
					v[k][p], v[k][q] = c*vp-s*vq, s*vp+c*vq
				}
			}
		}
		if !changed {
			break
		}
	}
	val := make([]float64, n)
	for i := range val {
		val[i] = a[i][i]
	}
	idx := make([]int, n)
	for i := range idx {
		idx[i] = i
	}
	sort.Slice(idx, func(i, j int) bool { return val[idx[i]] < val[idx[j]] })
	sv := make([]float64, n)
	vv := make([][]float64, n)
	for i := range vv {
		vv[i] = make([]float64, n)
	}
	for ne, oe := range idx {
		sv[ne] = val[oe]
		for i := 0; i < n; i++ {
			vv[i][ne] = v[i][oe]
		}
	}
	return sv, vv
}

func distances(v []sparse) [][]float64 {
	d := make([][]float64, len(v))
	for i := range d {
		d[i] = make([]float64, len(v))
		for j := 0; j < i; j++ {
			x := 0.
			seen := map[string]bool{}
			for f, z := range v[i] {
				u := z - v[j][f]
				x += u * u
				seen[f] = true
			}
			for f, z := range v[j] {
				if !seen[f] {
					x += z * z
				}
			}
			x = math.Sqrt(x)
			d[i][j], d[j][i] = x, x
		}
	}
	return d
}

func cluster(v []sparse, k int, seed int64) []int {
	return clusterMetric(v, k, seed, euclidean)
}

func clusterMetric(v []sparse, k int, seed int64, metric func(sparse, sparse) float64) []int {
	idx := sampleIndices(len(v), 200)
	sv := make([]sparse, len(idx))
	for i, j := range idx {
		sv[i] = v[j]
	}
	d := metricDistances(sv, metric)
	sl := globalregime.KMedoids(d, k, seed)
	cent := make([]sparse, k)
	counts := make([]int, k)
	for c := range cent {
		cent[c] = sparse{}
	}
	for i, j := range idx {
		c := sl[i]
		counts[c]++
		for f, z := range v[j] {
			cent[c][f] += z
		}
	}
	for c := range cent {
		for f := range cent[c] {
			cent[c][f] /= float64(max(1, counts[c]))
		}
	}
	out := make([]int, len(v))
	for i, x := range v {
		best, bd := 0, math.Inf(1)
		for c := 0; c < k; c++ {
			dd := metric(x, cent[c])
			if dd < bd {
				best, bd = c, dd
			}
		}
		out[i] = best
	}
	return out
}

func metricDistances(v []sparse, metric func(sparse, sparse) float64) [][]float64 {
	d := make([][]float64, len(v))
	for i := range d {
		d[i] = make([]float64, len(v))
		for j := 0; j < i; j++ {
			z := metric(v[i], v[j])
			d[i][j], d[j][i] = z, z
		}
	}
	return d
}

func jsDistance(a, b sparse) float64 {
	klA, klB := 0., 0.
	seen := map[string]bool{}
	for f, x := range a {
		m := .5 * (x + b[f])
		if x > 0 && m > 0 {
			klA += x * math.Log(x/m)
		}
		if y := b[f]; y > 0 && m > 0 {
			klB += y * math.Log(y/m)
		}
		seen[f] = true
	}
	for f, y := range b {
		if seen[f] {
			continue
		}
		if y > 0 {
			klB += y * math.Log(2)
		}
	}
	return math.Sqrt(math.Max(0, .5*(klA+klB)))
}
func clusterRaw(v []sparse, k int, seed int64) []int { return clusterMetric(v, k, seed, jsDistance) }
func sampleIndices(n, cap int) []int {
	if n <= cap {
		a := make([]int, n)
		for i := range a {
			a[i] = i
		}
		return a
	}
	a := make([]int, cap)
	for i := range a {
		a[i] = i * (n - 1) / (cap - 1)
	}
	return a
}
func euclidean(a, b sparse) float64 { return math.Sqrt(dot(subtract(a, b), subtract(a, b))) }
func silhouette(v []sparse, l []int) float64 {
	if len(v) < 2 {
		return 0
	}
	d := distances(v)
	sum := 0.
	for i := range v {
		same := 0.
		a := 0.
		other := map[int]struct {
			s float64
			n int
		}{}
		for j := range v {
			if i == j {
				continue
			}
			if l[i] == l[j] {
				a += d[i][j]
				same++
			} else {
				x := other[l[j]]
				x.s += d[i][j]
				x.n++
				other[l[j]] = x
			}
		}
		if same == 0 {
			continue
		}
		a /= float64(same)
		b := math.Inf(1)
		for _, x := range other {
			if x.n > 0 && x.s/float64(x.n) < b {
				b = x.s / float64(x.n)
			}
		}
		if den := math.Max(a, b); den > 0 && !math.IsInf(b, 1) {
			sum += (b - a) / den
		}
	}
	return sum / float64(len(v))
}

func sampledSilhouette(v []sparse, l []int) float64 {
	idx := sampleIndices(len(v), 200)
	sv := make([]sparse, len(idx))
	sl := make([]int, len(idx))
	for i, j := range idx {
		sv[i], sl[i] = v[j], l[j]
	}
	return silhouette(sv, sl)
}

func sampledSilhouetteMetric(v []sparse, l []int, metric func(sparse, sparse) float64) float64 {
	idx := sampleIndices(len(v), 200)
	sv := make([]sparse, len(idx))
	sl := make([]int, len(idx))
	for i, j := range idx {
		sv[i], sl[i] = v[j], l[j]
	}
	d := metricDistances(sv, metric)
	sum := 0.
	for i := range sv {
		same := 0.
		a := 0.
		other := map[int]struct {
			s float64
			n int
		}{}
		for j := range sv {
			if i == j {
				continue
			}
			if sl[i] == sl[j] {
				a += d[i][j]
				same++
			} else {
				x := other[sl[j]]
				x.s += d[i][j]
				x.n++
				other[sl[j]] = x
			}
		}
		if same == 0 {
			continue
		}
		a /= float64(same)
		b := math.Inf(1)
		for _, x := range other {
			if x.n > 0 && x.s/float64(x.n) < b {
				b = x.s / float64(x.n)
			}
		}
		if den := math.Max(a, b); den > 0 && !math.IsInf(b, 1) {
			sum += (b - a) / den
		}
	}
	return sum / float64(len(sv))
}
func assoc(labels []int, values []string) metadatavalidation.Metrics {
	c := make([]string, len(labels))
	for i, x := range labels {
		c[i] = strconvI(x)
	}
	return metadatavalidation.AssociationMetrics(values, c)
}
func strconvI(x int) string {
	if x == 0 {
		return "0"
	}
	return fmtInt(x)
}
func fmtInt(x int) string {
	digits := ""
	for x > 0 {
		digits = string(rune('0'+x%10)) + digits
		x /= 10
	}
	return digits
}

func adjustedRand(a, b []int) float64 {
	x := make([]string, len(a))
	y := make([]string, len(b))
	for i := range a {
		x[i] = strconvI(a[i])
		y[i] = strconvI(b[i])
	}
	return metadatavalidation.AssociationMetrics(x, y).ARI
}
func nmiInt(a, b []int) float64 {
	x := make([]string, len(a))
	y := make([]string, len(b))
	for i := range a {
		x[i] = strconvI(a[i])
		y[i] = strconvI(b[i])
	}
	return metadatavalidation.AssociationMetrics(x, y).NMI
}

func blockPermutation(labels []string, blocks []string, seed int64) []string {
	rng := rand.New(rand.NewSource(seed))
	by := map[string][]int{}
	for i, b := range blocks {
		by[b] = append(by[b], i)
	}
	keys := make([]string, 0, len(by))
	for k := range by {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	perm := rng.Perm(len(keys))
	out := make([]string, len(labels))
	for i, k := range keys {
		src := by[keys[perm[i]]]
		dst := by[k]
		for j, p := range dst {
			out[p] = labels[src[j%len(src)]]
		}
	}
	return out
}
