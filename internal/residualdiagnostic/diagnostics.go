package residualdiagnostic

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"zcore.dev/voinich/internal/metadatavalidation"
)

func labelsOf(w []window, field string) []string {
	out := make([]string, len(w))
	for i, x := range w {
		switch field {
		case "currier":
			out[i] = x.Currier
		case "hand":
			out[i] = x.Hand
		case "joint":
			out[i] = x.Joint
		case "folio":
			out[i] = x.Folio
		case "block":
			out[i] = x.Block
		}
	}
	return out
}
func clustersOf(w []window) []int {
	a := make([]int, len(w))
	for i, x := range w {
		a[i] = x.ExistingCluster
	}
	return a
}
func vectorsOf(w []window, kind string) []sparse {
	a := make([]sparse, len(w))
	for i, x := range w {
		switch kind {
		case "original":
			a[i] = x.Raw
		case "whitened":
			a[i] = x.Whitened
		default:
			a[i] = x.Residual
		}
	}
	return a
}

type runInfo struct {
	Cluster, Count, Largest int
	Median, LargestFraction float64
}

func extractRuns(w []window, labels []int) []runInfo {
	type run struct{ c, n int }
	var rr []run
	lastC := -1
	lastBlock := ""
	lastStart := -1
	step := 0
	for i, x := range w {
		if i > 0 && x.Block == lastBlock {
			d := x.Start - lastStart
			if step == 0 {
				step = d
			}
		}
		cont := i > 0 && x.Block == lastBlock && labels[i] == lastC && (step == 0 || x.Start-lastStart == step)
		if cont {
			rr[len(rr)-1].n++
		} else {
			rr = append(rr, run{labels[i], 1})
		}
		lastC, lastBlock, lastStart = labels[i], x.Block, x.Start
	}
	by := map[int][]int{}
	for _, r := range rr {
		by[r.c] = append(by[r.c], r.n)
	}
	keys := make([]int, 0, len(by))
	for c := range by {
		keys = append(keys, c)
	}
	sort.Ints(keys)
	out := make([]runInfo, 0, len(keys))
	for _, c := range keys {
		v := by[c]
		sort.Ints(v)
		total := 0
		for _, n := range v {
			total += n
		}
		med := float64(v[len(v)/2])
		if len(v)%2 == 0 {
			med = float64(v[len(v)/2-1]+v[len(v)/2]) / 2
		}
		out = append(out, runInfo{c, len(v), v[len(v)-1], med, float64(v[len(v)-1]) / float64(total)})
	}
	return out
}

func representation(w []window, name string, v []sparse, labels []int) representationRow {
	cur, hand, joint, bl := labelsOf(w, "currier"), labelsOf(w, "hand"), labelsOf(w, "joint"), labelsOf(w, "block")
	runs := extractRuns(w, labels)
	rc, largest := 0, 0.
	for _, r := range runs {
		rc += r.Count
		if r.LargestFraction > largest {
			largest = r.LargestFraction
		}
	}
	return representationRow{name, sampledSilhouette(v, labels), assoc(labels, cur).NMI, assoc(labels, hand).NMI, assoc(labels, joint).NMI, assoc(labels, bl).NMI, rc, largest}
}

func quantileInt(x []int, q float64) float64 {
	if len(x) == 0 {
		return 0
	}
	sort.Ints(x)
	p := q * float64(len(x)-1)
	lo := int(math.Floor(p))
	hi := int(math.Ceil(p))
	if lo == hi {
		return float64(x[lo])
	}
	return float64(x[lo]) + (p-float64(lo))*float64(x[hi]-x[lo])
}
func f(x float64) string { return strconv.FormatFloat(x, 'g', -1, 64) }

func compositionRows(w []window, k int) [][]string {
	var out [][]string
	for c := 0; c < k; c++ {
		var pos []int
		idx := []int{}
		for i, x := range w {
			if x.ExistingCluster == c {
				idx = append(idx, i)
				pos = append(pos, (x.Start+x.End)/2)
			}
		}
		if len(idx) == 0 {
			continue
		}
		minStart, maxEnd := w[idx[0]].Start, w[idx[0]].End
		physicalStart, physicalEnd := w[idx[0]].PhysicalStart, w[idx[0]].PhysicalEnd
		blockSet := map[string]bool{}
		for _, i := range idx {
			if w[i].Start < minStart {
				minStart = w[i].Start
			}
			if w[i].End > maxEnd {
				maxEnd = w[i].End
			}
			if w[i].PhysicalStart < physicalStart {
				physicalStart = w[i].PhysicalStart
			}
			if w[i].PhysicalEnd > physicalEnd {
				physicalEnd = w[i].PhysicalEnd
			}
			blockSet[w[i].Block] = true
		}
		out = append(out, []string{strconv.Itoa(c), "summary", "window_count", strconv.Itoa(len(idx)), "1", strconv.Itoa(int(quantileInt(pos, 0))), strconv.Itoa(int(quantileInt(pos, .25))), f(quantileInt(pos, .5)), strconv.Itoa(int(quantileInt(pos, .75))), strconv.Itoa(int(quantileInt(pos, 1))), strconv.Itoa(minStart), strconv.Itoa(maxEnd), strconv.Itoa(physicalStart), strconv.Itoa(physicalEnd)})
		out = append(out, []string{strconv.Itoa(c), "summary", "physical_block_count", strconv.Itoa(len(blockSet)), "1", "", "", "", "", "", strconv.Itoa(minStart), strconv.Itoa(maxEnd), strconv.Itoa(physicalStart), strconv.Itoa(physicalEnd)})
		for _, dim := range []string{"currier", "hand", "joint", "block", "folio"} {
			counts := map[string]int{}
			for _, i := range idx {
				counts[labelsOf(w, dim)[i]]++
			}
			keys := make([]string, 0, len(counts))
			for x := range counts {
				keys = append(keys, x)
			}
			sort.Strings(keys)
			for _, x := range keys {
				out = append(out, []string{strconv.Itoa(c), dim, x, strconv.Itoa(counts[x]), f(float64(counts[x]) / float64(len(idx))), "", "", "", "", "", "", "", "", ""})
			}
		}
	}
	return out
}

type dispersion struct {
	joint                                 string
	n                                     int
	total, mean, median, trace, effective float64
	variance                              sparse
	centered                              []sparse
	covFrob2                              float64
}

func dispersions(w []window) []dispersion {
	by := map[string][]sparse{}
	for _, x := range w {
		by[x.Joint] = append(by[x.Joint], x.Residual)
	}
	keys := make([]string, 0, len(by))
	for k := range by {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]dispersion, 0, len(keys))
	for _, k := range keys {
		vs := by[k]
		mu := meanSparse(vs)
		cen := make([]sparse, len(vs))
		for i, v := range vs {
			cen[i] = subtract(v, mu)
		}
		vari := varianceSparse(vs)
		vals := make([]float64, 0, len(vari))
		tr := 0.
		for _, v := range vari {
			vals = append(vals, v)
			tr += v
		}
		sort.Float64s(vals)
		fro := covCross(cen, cen)
		eff := 0.
		if fro > 0 {
			eff = tr * tr / fro
		}
		med := 0.
		if len(vals) > 0 {
			med = vals[len(vals)/2]
		}
		out = append(out, dispersion{k, len(vs), tr, tr / float64(max(1, len(vals))), med, tr, eff, vari, cen, fro})
	}
	return out
}
func covCross(a, b []sparse) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	s := 0.
	for _, x := range a {
		for _, y := range b {
			z := dot(x, y)
			s += z * z
		}
	}
	return s / float64(len(a)*len(b))
}
func covarianceDistance(a, b dispersion) (ratio, frob, diag float64) {
	if b.total > 0 {
		ratio = a.total / b.total
	}
	cross := covCross(a.centered, b.centered)
	d2 := a.covFrob2 + b.covFrob2 - 2*cross
	if d2 < 0 && d2 > -1e-12 {
		d2 = 0
	}
	frob = math.Sqrt(math.Max(0, d2))
	keys := map[string]bool{}
	for k := range a.variance {
		keys[k] = true
	}
	for k := range b.variance {
		keys[k] = true
	}
	for k := range keys {
		x := a.variance[k] - b.variance[k]
		diag += x * x
	}
	diag = math.Sqrt(diag)
	return
}

func confusionRows(target, model string, truth, pred []string) [][]string {
	tab := map[string]map[string]int{}
	for i, t := range truth {
		if tab[t] == nil {
			tab[t] = map[string]int{}
		}
		tab[t][pred[i]]++
	}
	var out [][]string
	for _, t := range uniqueStrings(truth) {
		ps := uniqueStrings(pred)
		for _, p := range ps {
			out = append(out, []string{target, model, t, p, strconv.Itoa(tab[t][p])})
		}
	}
	return out
}

func conditionalBlockAssociation(w []window, l []int) float64 {
	by := map[string][]int{}
	for i, x := range w {
		by[x.Joint] = append(by[x.Joint], i)
	}
	total, score := 0, 0.
	for _, idx := range by {
		blocks := make([]string, len(idx))
		cs := make([]int, len(idx))
		for j, i := range idx {
			blocks[j] = w[i].Block
			cs[j] = l[i]
		}
		m := assoc(cs, blocks)
		score += float64(len(idx)) * m.NMI
		total += len(idx)
	}
	if total == 0 {
		return 0
	}
	return score / float64(total)
}

func blockAssociationRows(w []window, l []int, perms int, seed int64) [][]string {
	blocks := labelsOf(w, "block")
	c := make([]string, len(l))
	for i, x := range l {
		c[i] = strconv.Itoa(x)
	}
	m := metadatavalidation.AssociationMetrics(c, blocks)
	cond := conditionalBlockAssociation(w, l)
	ex := 0
	for p := 0; p < perms; p++ {
		pl := blockPermutation(c, blocks, seed+int64(p))
		pm := metadatavalidation.AssociationMetrics(pl, blocks)
		if pm.NMI >= m.NMI {
			ex++
		}
	}
	return [][]string{{"overall", f(m.NMI), f(m.ARI), f(m.Homogeneity), f(m.Completeness), f(float64(ex+1) / float64(perms+1)), strconv.Itoa(perms)}, {"conditional_on_joint", f(cond), "", "", "", "", "0"}}
}

func positionRows(w []window, l []int, tokenCount int, bs []block) [][]string {
	bins := make([]string, len(w))
	byID := map[string]block{}
	for _, b := range bs {
		byID[b.ID] = b
	}
	var out [][]string
	for i, x := range w {
		mid := (x.Start + x.End) / 2
		bin := mid * 10 / max(1, tokenCount)
		if bin > 9 {
			bin = 9
		}
		bins[i] = strconv.Itoa(bin)
		b := byID[x.Block]
		p := normalizedBlockPosition(mid, b)
		out = append(out, []string{"window", strconv.Itoa(i), strconv.Itoa(l[i]), strconv.Itoa(bin), f(p), strconv.Itoa(mid)})
	}
	m := assoc(l, bins)
	out = append(out, []string{"summary", "", "", "", "", f(m.NMI)})
	return out
}

func normalizedBlockPosition(position int, b block) float64 {
	if b.len() <= 0 {
		return 0
	}
	p := float64(position-b.Start) / float64(b.len())
	if p < 0 {
		return 0
	}
	if p > 1 {
		return 1
	}
	return p
}

func bothRegimesRecur(seen map[int]bool, k int) bool {
	for c := 0; c < k; c++ {
		if !seen[c] {
			return false
		}
	}
	return true
}

func recurrenceRows(w []window, k int, seed int64) [][]string {
	blocks := uniqueStrings(labelsOf(w, "block"))
	var out [][]string
	all := vectorsOf(w, "residual")
	for bi, held := range blocks {
		var train, test []int
		for i, x := range w {
			if x.Block == held {
				test = append(test, i)
			} else {
				train = append(train, i)
			}
		}
		tv := make([]sparse, len(train))
		for i, j := range train {
			tv[i] = all[j]
		}
		lab := cluster(tv, k, seed+int64(bi))
		med := make([]sparse, k)
		count := make([]int, k)
		for c := range med {
			med[c] = sparse{}
		}
		for i, c := range lab {
			count[c]++
			for f, z := range tv[i] {
				med[c][f] += z
			}
		}
		for c := range med {
			for f := range med[c] {
				med[c][f] /= float64(max(1, count[c]))
			}
		}
		within := make([][]float64, k)
		for i, c := range lab {
			within[c] = append(within[c], euclidean(tv[i], med[c]))
		}
		threshold := make([]float64, k)
		for c := range within {
			sort.Float64s(within[c])
			if len(within[c]) > 0 {
				threshold[c] = within[c][int(.95*float64(len(within[c])-1))]
			}
		}
		seen := map[int]bool{}
		sumD, sumConf := 0., 0.
		for _, i := range test {
			best, second, bc := math.Inf(1), math.Inf(1), 0
			for c := 0; c < k; c++ {
				d := euclidean(all[i], med[c])
				if d < best {
					second, best, bc = best, d, c
				} else if d < second {
					second = d
				}
			}
			conf := 0.
			if second > 0 && !math.IsInf(second, 1) {
				conf = (second - best) / second
			}
			sumD += best
			sumConf += conf
			if best <= threshold[bc] {
				seen[bc] = true
			}
		}
		out = append(out, []string{held, strconv.Itoa(len(test)), f(sumConf / float64(max(1, len(test)))), f(sumD / float64(max(1, len(test)))), strconv.Itoa(len(seen)), strconv.FormatBool(bothRegimesRecur(seen, k))})
	}
	return out
}

func centeringRows(w []window, folds []foldDiagnostic) [][]string {
	by := map[string][]sparse{}
	for _, x := range w {
		by[x.Joint] = append(by[x.Joint], x.Residual)
	}
	var out [][]string
	classes := make([]string, 0, len(by))
	for c := range by {
		classes = append(classes, c)
	}
	sort.Strings(classes)
	for _, c := range classes {
		test := normOf(meanSparse(by[c]))
		train := norms{}
		n := 0
		for _, d := range folds {
			if d.Joint == c {
				train.L1 += d.TrainMean.L1
				train.L2 += d.TrainMean.L2
				train.Linf = math.Max(train.Linf, d.TrainMean.Linf)
				train.MeanAbs += d.TrainMean.MeanAbs
				n++
			}
		}
		if n > 0 {
			train.L1 /= float64(n)
			train.L2 /= float64(n)
			train.MeanAbs /= float64(n)
		}
		out = append(out, []string{strconv.Itoa(folds[0].WindowSize), c, "training", f(train.L1), f(train.L2), f(train.Linf), f(train.MeanAbs)})
		out = append(out, []string{strconv.Itoa(folds[0].WindowSize), c, "held_out", f(test.L1), f(test.L2), f(test.Linf), f(test.MeanAbs)})
	}
	return out
}

func normRows(w []window) [][]string {
	var out [][]string
	for _, dim := range []string{"currier", "hand", "joint", "block"} {
		labels := labelsOf(w, dim)
		for _, v := range uniqueStrings(labels) {
			n := 0
			sum := norms{}
			for i, x := range w {
				if labels[i] != v {
					continue
				}
				z := normOf(x.Residual)
				sum.L1 += z.L1
				sum.L2 += z.L2
				sum.Linf += z.Linf
				n++
			}
			out = append(out, []string{dim, v, strconv.Itoa(n), f(sum.L1 / float64(n)), f(sum.L2 / float64(n)), f(sum.Linf / float64(n))})
		}
	}
	return out
}

func joinDistribution(x []string) string { return strings.Join(x, "|") }
func fmtBool(x bool) string              { return fmt.Sprintf("%t", x) }
