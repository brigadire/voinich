package residualdiagnostic

import (
	"math"
	"math/rand"
	"sort"
)

type classification struct {
	Target, Model                           string
	BalancedAccuracy, MacroF1, CrossEntropy float64
	PermutationP                            float64
	Truth, Pred                             []string
}

func uniqueStrings(x []string) []string {
	m := map[string]bool{}
	for _, v := range x {
		if v != "" {
			m[v] = true
		}
	}
	out := make([]string, 0, len(m))
	for v := range m {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

// cvLogistic performs strict leave-physical-block-out evaluation. Feature
// scaling and model fitting are repeated inside every training fold.
func gramMatrix(x []sparse) [][]float64 {
	g := make([][]float64, len(x))
	for i := range g {
		g[i] = make([]float64, len(x))
		for j := 0; j <= i; j++ {
			z := dot(x[i], x[j])
			g[i][j], g[j][i] = z, z
		}
	}
	return g
}

func cvLogistic(x []sparse, gram [][]float64, y, blocks []string, seed int64) (pred []string, ce float64) {
	classes := uniqueStrings(y)
	ci := map[string]int{}
	for i, c := range classes {
		ci[c] = i
	}
	pred = make([]string, len(y))
	probs := make([][]float64, len(y))
	ub := uniqueStrings(blocks)
	for fi, held := range ub {
		var train, test []int
		for i, b := range blocks {
			if b == held {
				test = append(test, i)
			} else {
				train = append(train, i)
			}
		}
		if len(train) == 0 {
			continue
		}
		alpha, bias := trainLogisticGram(gram, y, train, classes, ci, seed+int64(fi))
		for _, i := range test {
			p := predictLogisticGram(gram, i, train, alpha, bias)
			probs[i] = p
			best := 0
			for c := 1; c < len(p); c++ {
				if p[c] > p[best] {
					best = c
				}
			}
			pred[i] = classes[best]
		}
	}
	for i, t := range y {
		if probs[i] == nil {
			continue
		}
		p := probs[i][ci[t]]
		if p < 1e-15 {
			p = 1e-15
		}
		ce -= math.Log(p)
	}
	if len(y) > 0 {
		ce /= float64(len(y))
	}
	return pred, ce
}

func trainLogisticGram(gram [][]float64, y []string, idx []int, classes []string, ci map[string]int, seed int64) ([][]float64, []float64) {
	alpha := make([][]float64, len(idx))
	for i := range alpha {
		alpha[i] = make([]float64, len(classes))
	}
	bias := make([]float64, len(classes))
	counts := make([]int, len(classes))
	for _, i := range idx {
		counts[ci[y[i]]]++
	}
	for c := range bias {
		bias[c] = math.Log(float64(counts[c]+1) / float64(len(idx)+len(classes)))
	}
	rng := rand.New(rand.NewSource(seed))
	for epoch := 0; epoch < 35; epoch++ {
		lr := 1.2 / math.Sqrt(1+float64(epoch))
		grad := make([][]float64, len(idx))
		gb := make([]float64, len(classes))
		for ii, i := range idx {
			p := predictLogisticGram(gram, i, idx, alpha, bias)
			grad[ii] = make([]float64, len(classes))
			for c := range classes {
				z := p[c]
				if c == ci[y[i]] {
					z--
				}
				z /= float64(len(idx))
				grad[ii][c] = z
				gb[c] += z
			}
		}
		shrink := 1 - lr*1e-4
		for ii := range alpha {
			for c := range classes {
				alpha[ii][c] = alpha[ii][c]*shrink - lr*grad[ii][c]
			}
		}
		for c := range bias {
			bias[c] -= lr * gb[c]
		}
		_ = rng
	}
	return alpha, bias
}

func predictLogisticGram(gram [][]float64, j int, train []int, alpha [][]float64, bias []float64) []float64 {
	z := append([]float64(nil), bias...)
	mx := -math.MaxFloat64
	for ii, i := range train {
		g := gram[i][j]
		for c := range z {
			z[c] += alpha[ii][c] * g
		}
	}
	for _, v := range z {
		if v > mx {
			mx = v
		}
	}
	sum := 0.
	for c := range z {
		z[c] = math.Exp(z[c] - mx)
		sum += z[c]
	}
	for c := range z {
		z[c] /= sum
	}
	return z
}

func trainLogistic(x []sparse, y []string, idx []int, classes []string, ci map[string]int, seed int64) ([]sparse, []float64, sparse) {
	scale := sparse{}
	mean := sparse{}
	for _, i := range idx {
		for f, v := range x[i] {
			mean[f] += v
			scale[f] += v * v
		}
	}
	for f := range scale {
		mu := mean[f] / float64(len(idx))
		v := scale[f]/float64(len(idx)) - mu*mu
		if v < 1e-12 {
			v = 1
		}
		scale[f] = 1 / math.Sqrt(v)
	}
	w := make([]sparse, len(classes))
	bias := make([]float64, len(classes))
	rng := rand.New(rand.NewSource(seed))
	for c := range w {
		w[c] = sparse{}
		bias[c] = (rng.Float64() - .5) * 1e-9
	}
	for epoch := 0; epoch < 180; epoch++ {
		lr := .35 / math.Sqrt(1+float64(epoch)/20)
		order := rng.Perm(len(idx))
		for _, oi := range order {
			i := idx[oi]
			p := predictLogistic(x[i], w, bias, scale)
			truth := ci[y[i]]
			for c := range classes {
				g := p[c]
				if c == truth {
					g--
				}
				bias[c] -= lr * g
				for f, v := range x[i] {
					w[c][f] -= lr * (g*v*scale[f] + 1e-5*w[c][f])
				}
			}
		}
	}
	return w, bias, scale
}

func predictLogistic(x sparse, w []sparse, bias []float64, scale sparse) []float64 {
	z := make([]float64, len(w))
	mx := -math.MaxFloat64
	for c := range w {
		z[c] = bias[c]
		for f, v := range x {
			z[c] += w[c][f] * v * scale[f]
		}
		if z[c] > mx {
			mx = z[c]
		}
	}
	sum := 0.
	for c := range z {
		z[c] = math.Exp(z[c] - mx)
		sum += z[c]
	}
	for c := range z {
		z[c] /= sum
	}
	return z
}

func scoreClassification(truth, pred []string) (bal, f1 float64) {
	classes := uniqueStrings(truth)
	for _, c := range classes {
		tp, fn, fp := 0., 0., 0.
		for i, t := range truth {
			if t == c && pred[i] == c {
				tp++
			}
			if t == c && pred[i] != c {
				fn++
			}
			if t != c && pred[i] == c {
				fp++
			}
		}
		rec := 0.
		if tp+fn > 0 {
			rec = tp / (tp + fn)
		}
		prec := 0.
		if tp+fp > 0 {
			prec = tp / (tp + fp)
		}
		bal += rec
		if prec+rec > 0 {
			f1 += 2 * prec * rec / (prec + rec)
		}
	}
	if len(classes) > 0 {
		bal /= float64(len(classes))
		f1 /= float64(len(classes))
	}
	return
}

func majorityBaseline(y []string) ([]string, float64) {
	counts := map[string]int{}
	for _, v := range y {
		counts[v]++
	}
	best := ""
	for c, n := range counts {
		if n > counts[best] || (n == counts[best] && c < best) {
			best = c
		}
	}
	p := make([]string, len(y))
	for i := range p {
		p[i] = best
	}
	ce := 0.
	if len(y) > 0 {
		freq := float64(counts[best]) / float64(len(y))
		for _, v := range y {
			q := (1 - freq) / float64(max(1, len(counts)-1))
			if v == best {
				q = freq
			}
			if q <= 0 {
				q = 1e-15
			}
			ce -= math.Log(q)
		}
		ce /= float64(len(y))
	}
	return p, ce
}

func frequencyBaseline(y []string, seed int64) ([]string, float64) {
	classes := uniqueStrings(y)
	counts := map[string]int{}
	for _, v := range y {
		counts[v]++
	}
	rng := rand.New(rand.NewSource(seed))
	p := make([]string, len(y))
	for i := range p {
		z := rng.Intn(len(y))
		acc := 0
		for _, c := range classes {
			acc += counts[c]
			if z < acc {
				p[i] = c
				break
			}
		}
	}
	ce := 0.
	for _, v := range y {
		q := float64(counts[v]) / float64(len(y))
		ce -= math.Log(q)
	}
	return p, ce / float64(len(y))
}

func secondOrderFeatures(x []sparse) []sparse {
	out := make([]sparse, len(x))
	for i, v := range x {
		n := normOf(v)
		p, neg, zero := 0., 0., 0.
		for _, z := range v {
			if z > 0 {
				p++
			} else if z < 0 {
				neg++
			} else {
				zero++
			}
		}
		out[i] = sparse{"l1": n.L1, "l2": n.L2, "l2_sq": n.L2 * n.L2, "linf": n.Linf, "positive": p, "negative": neg, "sparsity": zero}
	}
	return out
}
