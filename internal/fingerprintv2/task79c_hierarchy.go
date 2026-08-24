package fingerprintv2

import (
	"math"
	"math/rand"
)

// HierarchyFoldResult is one fold's held-out negative log-likelihood (NLL)
// for the flat baseline and the two hierarchical models specified in
// TASK79C_DESIGN.md section 9-10 (HR3: folio-level partial pooling; HR5:
// folio nested in section). Lower NLL is better.
type HierarchyFoldResult struct {
	Fold          int     `json:"fold"`
	HeldOutFolios int     `json:"held_out_folios"`
	HeldOutLines  int     `json:"held_out_lines"`
	FlatNLL       float64 `json:"flat_nll"`
	HR3NLL        float64 `json:"hr3_nll"`
	HR5NLL        float64 `json:"hr5_nll"`
	HR3Delta      float64 `json:"hr3_delta"` // HR3 - flat; negative = hierarchical better
	HR5Delta      float64 `json:"hr5_delta"`
	Testable      bool    `json:"testable"`
	Reason        string  `json:"reason,omitempty"`
}

// HierarchyValidation is the full out-of-sample HR3/HR5-vs-flat result.
type HierarchyValidation struct {
	TargetQuantity   string                `json:"target_quantity"`
	Folds            int                   `json:"folds"`
	Seed             int64                 `json:"seed"`
	MinGroupSize     int                   `json:"min_group_size"`
	FoldResults      []HierarchyFoldResult `json:"fold_results"`
	MeanHR3Delta     float64               `json:"mean_hr3_delta"`
	MeanHR5Delta     float64               `json:"mean_hr5_delta"`
	HR3BetterFolds   int                   `json:"hr3_better_folds"`
	HR5BetterFolds   int                   `json:"hr5_better_folds"`
	TestableFolds    int                   `json:"testable_folds"`
	Verdict          string                `json:"verdict"`
	VerdictRationale string                `json:"verdict_rationale"`
}

// foldFolios deterministically assigns every folio present in ls to one of
// k folds via a single seeded shuffle of the sorted folio list followed by
// contiguous splitting (TASK79C_DESIGN.md section 10), so every line from a
// given folio is always in the same fold.
func foldFolios(ls []LineProfile, k int, seed int64) map[string]int {
	folioSet := map[string]bool{}
	for _, l := range ls {
		folioSet[l.Folio] = true
	}
	folios := orderedKeys(folioSet)
	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(len(folios), func(i, j int) { folios[i], folios[j] = folios[j], folios[i] })
	assign := map[string]int{}
	n := len(folios)
	for i, f := range folios {
		fold := i * k / n
		if fold >= k {
			fold = k - 1
		}
		assign[f] = fold
	}
	return assign
}

// groupMoments returns, for a slice of (group,value) observations, each
// group's mean and count, and the grand mean/variance -- the same
// method-of-moments building block varianceShare already uses, generalized
// to two nesting levels.
func groupMoments(values []float64, group []string) (mean map[string]float64, n map[string]int, grand float64, variance float64) {
	mean, n = map[string]float64{}, map[string]int{}
	sum := map[string]float64{}
	total, count := 0.0, 0
	for i, v := range values {
		g := group[i]
		sum[g] += v
		n[g]++
		total += v
		count++
	}
	if count == 0 {
		return mean, n, 0, 0
	}
	grand = total / float64(count)
	for g, s := range sum {
		mean[g] = s / float64(n[g])
	}
	ss := 0.0
	for _, v := range values {
		d := v - grand
		ss += d * d
	}
	variance = ss / math.Max(1, float64(count-1))
	return mean, n, grand, variance
}

// shrinkageWeight is the classical empirical-Bayes/James-Stein-style
// partial-pooling weight n/(n+k), k = within/between variance ratio
// (method-of-moments, matching the style already used by varianceShare;
// not a fitted mixed-effects optimizer, per TASK79C_DESIGN.md section 9).
func shrinkageWeight(n int, within, between float64) float64 {
	if between <= 0 {
		return 0
	}
	k := within / between
	if k < 0 {
		k = 0
	}
	return float64(n) / (float64(n) + k)
}

// betweenWithinVariance decomposes a target quantity's total training-fold
// variance into between-group and (pooled) within-group components via the
// same additive method-of-moments decomposition varianceShare uses.
func betweenWithinVariance(values []float64, group []string) (between, within float64) {
	groupMean, n, grand, _ := groupMoments(values, group)
	betweenSS, withinSS := 0.0, 0.0
	for i, v := range values {
		gm := groupMean[group[i]]
		d := v - gm
		withinSS += d * d
	}
	for g, m := range groupMean {
		d := m - grand
		betweenSS += float64(n[g]) * d * d
	}
	count := float64(len(values))
	if count < 2 {
		return 0, 0
	}
	between = betweenSS / (count - 1)
	within = withinSS / math.Max(1, count-float64(len(groupMean)))
	return between, within
}

// gaussianNLL is the held-out negative log-likelihood of observations under
// a Gaussian with the supplied per-point predicted mean and a single fixed
// residual variance estimated on the training fold only. Used for the flat
// baseline, whose variance genuinely does not vary by held-out point.
func gaussianNLL(observed, predicted []float64, variance float64) float64 {
	variances := make([]float64, len(observed))
	for i := range variances {
		variances[i] = variance
	}
	return gaussianNLLPerPoint(observed, predicted, variances)
}

// gaussianNLLPerPoint is gaussianNLL with a per-point predictive variance.
// HR3/HR5 need this because a held-out point's predictive variance must
// itself interpolate by the same shrinkage weight used for its predicted
// mean: a group with zero training coverage is scored at (near) the total
// training-fold variance (matching the flat baseline it necessarily
// collapses to), while a well-observed group is scored at (near) the
// within-group residual alone. Scoring every point at the same fixed
// within-group residual, regardless of how much training coverage its
// group actually had, understates predictive uncertainty for undercovered
// groups and blows up the NLL whenever such a point's true value is far
// from its point prediction.
func gaussianNLLPerPoint(observed, predicted, variance []float64) float64 {
	nll := 0.0
	for i := range observed {
		v := variance[i]
		if v <= 0 {
			v = 1e-6
		}
		d := observed[i] - predicted[i]
		nll += 0.5*math.Log(2*math.Pi*v) + (d*d)/(2*v)
	}
	return nll / math.Max(1, float64(len(observed)))
}

// hierarchyOutOfSample implements TASK79C_DESIGN.md sections 9-10: folio-
// block k-fold cross-validation comparing a flat baseline against HR3
// (folio-level partial pooling) and HR5 (folio nested in section) on
// held-out line token counts.
func hierarchyOutOfSample(ls []LineProfile, folds int, seed int64, minGroupSize int) HierarchyValidation {
	assign := foldFolios(ls, folds, seed)
	var results []HierarchyFoldResult
	hr3Better, hr5Better, testable := 0, 0, 0
	var hr3Deltas, hr5Deltas []float64
	for fold := 0; fold < folds; fold++ {
		var trainVals []float64
		var trainFolioKeys, trainSectionKeys []string
		var heldOutVals []float64
		var heldOutFolioKeys, heldOutSectionKeys []string
		heldOutFolioSet := map[string]bool{}
		for _, l := range ls {
			v := float64(l.TokenCount)
			if assign[l.Folio] == fold {
				heldOutVals = append(heldOutVals, v)
				heldOutFolioKeys = append(heldOutFolioKeys, l.Folio)
				heldOutSectionKeys = append(heldOutSectionKeys, l.Section)
				heldOutFolioSet[l.Folio] = true
			} else {
				trainVals = append(trainVals, v)
				trainFolioKeys = append(trainFolioKeys, l.Folio)
				trainSectionKeys = append(trainSectionKeys, l.Section)
			}
		}
		res := HierarchyFoldResult{Fold: fold, HeldOutFolios: len(heldOutFolioSet), HeldOutLines: len(heldOutVals)}
		trainFolioSet := map[string]bool{}
		for _, f := range trainFolioKeys {
			trainFolioSet[f] = true
		}
		trainSectionSet := map[string]bool{}
		for _, s := range trainSectionKeys {
			trainSectionSet[s] = true
		}
		if len(trainVals) < 2*minGroupSize || len(heldOutVals) == 0 || len(trainFolioSet) < minGroupSize || len(trainSectionSet) < 2 {
			res.Reason = "insufficient training folios/sections for a stable variance-component estimate"
			results = append(results, res)
			continue
		}
		res.Testable = true
		testable++

		// Flat baseline.
		grandMean, grandVar := 0.0, 0.0
		for _, v := range trainVals {
			grandMean += v
		}
		grandMean /= float64(len(trainVals))
		for _, v := range trainVals {
			d := v - grandMean
			grandVar += d * d
		}
		grandVar /= math.Max(1, float64(len(trainVals)-1))
		flatPred := make([]float64, len(heldOutVals))
		for i := range flatPred {
			flatPred[i] = grandMean
		}
		res.FlatNLL = gaussianNLL(heldOutVals, flatPred, grandVar)

		// HR3: folio-level partial pooling. folioN[f]/folioMean[f] use Go's
		// map zero value (0, 0.0) for a folio absent from the training
		// fold, which is every held-out folio under strict folio-block CV
		// (section 10); shrinkageWeight(0, ...) is then exactly 0, so the
		// point prediction and its interpolated variance both correctly
		// collapse to the flat baseline's for such folios, per section 9's
		// addendum, rather than being scored at an inappropriately small
		// fixed residual variance.
		folioMean, folioN, _, _ := groupMoments(trainVals, trainFolioKeys)
		folioBetween, folioWithin := betweenWithinVariance(trainVals, trainFolioKeys)
		hr3Pred := make([]float64, len(heldOutVals))
		hr3Var := make([]float64, len(heldOutVals))
		for i, f := range heldOutFolioKeys {
			w := shrinkageWeight(folioN[f], folioWithin, folioBetween)
			hr3Pred[i] = grandMean + w*(folioMean[f]-grandMean)
			hr3Var[i] = folioWithin + (1-w)*folioBetween
		}
		res.HR3NLL = gaussianNLLPerPoint(heldOutVals, hr3Pred, hr3Var)

		// HR5: folio nested in section (two-level partial pooling). A
		// held-out folio's section is usually still represented by sibling
		// folios in training even though the folio itself never is, so the
		// section-level shrinkage weight carries genuine cross-fold
		// information; the folio-level term above it still vanishes
		// (folioN[f]=0), so the prediction is effectively the shrunk
		// section mean, but the formula is left in the general two-level
		// form for correctness if this code is ever reused with a
		// different (non-folio-block) split.
		sectionMean, sectionN, _, _ := groupMoments(trainVals, trainSectionKeys)
		sectionBetween, sectionWithin := betweenWithinVariance(trainVals, trainSectionKeys)
		hr5Pred := make([]float64, len(heldOutVals))
		hr5Var := make([]float64, len(heldOutVals))
		for i, f := range heldOutFolioKeys {
			sec := heldOutSectionKeys[i]
			ws := shrinkageWeight(sectionN[sec], sectionWithin, sectionBetween)
			shrunkSection := grandMean + ws*(sectionMean[sec]-grandMean)
			sectionVar := sectionWithin + (1-ws)*sectionBetween
			wf := shrinkageWeight(folioN[f], folioWithin, folioBetween)
			hr5Pred[i] = shrunkSection + wf*(folioMean[f]-grandMean)
			// folioWithin and sectionVar both already include the true
			// line-level residual (each is "a line's deviation from some
			// group mean"), so this deliberately double-counts it rather
			// than attempting an exact three-term decomposition; the bias
			// is conservative (overstates HR5's predictive uncertainty,
			// understating rather than overstating any hierarchical
			// benefit), consistent with not tuning this estimator to make
			// hierarchy look better.
			hr5Var[i] = folioWithin + (1-wf)*sectionVar
		}
		res.HR5NLL = gaussianNLLPerPoint(heldOutVals, hr5Pred, hr5Var)

		res.HR3Delta = res.HR3NLL - res.FlatNLL
		res.HR5Delta = res.HR5NLL - res.FlatNLL
		if res.HR3Delta < 0 {
			hr3Better++
		}
		if res.HR5Delta < 0 {
			hr5Better++
		}
		hr3Deltas = append(hr3Deltas, res.HR3Delta)
		hr5Deltas = append(hr5Deltas, res.HR5Delta)
		results = append(results, res)
	}
	v := HierarchyValidation{
		TargetQuantity: "line token count", Folds: folds, Seed: seed, MinGroupSize: minGroupSize,
		FoldResults: results, MeanHR3Delta: mean(hr3Deltas), MeanHR5Delta: mean(hr5Deltas),
		HR3BetterFolds: hr3Better, HR5BetterFolds: hr5Better, TestableFolds: testable,
	}
	v.Verdict, v.VerdictRationale = hierarchyVerdict(v, len(hr3Deltas))
	return v
}

func hierarchyVerdict(v HierarchyValidation, testableFolds int) (string, string) {
	if testableFolds == 0 {
		return "INCONCLUSIVE", "no fold had enough training folios/sections to estimate a stable variance component"
	}
	hr3Consistent := v.MeanHR3Delta < 0 && v.HR3BetterFolds >= 4
	hr5Consistent := v.MeanHR5Delta < 0 && v.HR5BetterFolds >= 4
	if hr3Consistent || hr5Consistent {
		return "SUPPORTED", "at least one of HR3/HR5 has a negative mean held-out NLL delta and beats the flat baseline in >=4/5 folds"
	}
	if v.MeanHR3Delta >= 0 && v.MeanHR5Delta >= 0 {
		return "NOT_SUPPORTED", "neither HR3 nor HR5 improves mean held-out NLL over the flat baseline"
	}
	return "INCONCLUSIVE", "fold-level sign is mixed for both HR3 and HR5 (no model beats flat in >=4/5 folds despite a negative mean)"
}
