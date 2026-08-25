package main

import (
	"math"
	"sort"
)

// PredictiveMetrics holds one candidate's PM1-PM6 on one partition/corpus.
type PredictiveMetrics struct {
	PM1, PM2, PM3, PM4, PM5 float64
	PM6                     float64
	PM6Valid                bool
	PM6CILow, PM6CIHigh     float64
	UndefinedFraction       float64
	NormError               float64
}

func devVocabulary(dev []TokenOccurrence) map[string]bool {
	v := map[string]bool{}
	for _, o := range dev {
		v[o.Raw] = true
	}
	return v
}

// ComputePM1PM2PM3PM5 scores every occurrence in target under model.
func ComputePM1PM2PM3PM5(model FittedModel, target []TokenOccurrence) PredictiveMetrics {
	var pm1, units float64
	var events []ScoreEvent
	undefined := 0
	for _, o := range target {
		evs := model.Events(o.Raw, o.Glyphs)
		for _, e := range evs {
			pm1 += e.NegLog2Prob
			if math.IsNaN(e.Confidence) || math.IsNaN(e.NegLog2Prob) {
				undefined++
			}
		}
		units += float64(model.ScoredUnits(o.Glyphs))
		events = append(events, evs...)
	}
	pm2 := pm1 / units
	pm3 := math.Pow(2, pm2)
	pm5 := computePM5(events)
	uf := 0.0
	if len(events) > 0 {
		uf = float64(undefined) / float64(len(events))
	}
	return PredictiveMetrics{PM1: pm1, PM2: pm2, PM3: pm3, PM5: pm5, UndefinedFraction: uf}
}

// ComputePM2Only computes just PM2 (cross entropy per scored unit), for
// callers that only need the DEVELOPMENT-side overfitting-gap term and
// would otherwise pay for PM5's event-slice accumulation over a large
// partition for no benefit.
func ComputePM2Only(model FittedModel, target []TokenOccurrence) float64 {
	var pm1, units float64
	for _, o := range target {
		for _, e := range model.Events(o.Raw, o.Glyphs) {
			pm1 += e.NegLog2Prob
		}
		units += float64(model.ScoredUnits(o.Glyphs))
	}
	return pm1 / units
}

func computePM5(events []ScoreEvent) float64 {
	type bin struct {
		sumConf, sumCorrect float64
		n                   int
	}
	bins := make([]bin, 10)
	for _, e := range events {
		c := e.Confidence
		idx := int(c * 10)
		if idx > 9 {
			idx = 9
		}
		if idx < 0 {
			idx = 0
		}
		bins[idx].sumConf += c
		if e.Correct {
			bins[idx].sumCorrect++
		}
		bins[idx].n++
	}
	total := len(events)
	if total == 0 {
		return math.NaN()
	}
	pm5 := 0.0
	for _, b := range bins {
		if b.n == 0 {
			continue
		}
		meanConf := b.sumConf / float64(b.n)
		meanCorrect := b.sumCorrect / float64(b.n)
		pm5 += (float64(b.n) / float64(total)) * math.Abs(meanConf-meanCorrect)
	}
	return pm5
}

// ComputePM4 is the mean whole-TOKEN probability over target occurrences
// whose raw TOKEN string is absent from devVocab (type-unseen HELDOUT
// occurrences).
func ComputePM4(model FittedModel, target []TokenOccurrence, devVocab map[string]bool) float64 {
	sum := 0.0
	n := 0
	for _, o := range target {
		if devVocab[o.Raw] {
			continue
		}
		nll := model.TokenNegLog2Prob(o.Raw, o.Glyphs)
		sum += math.Exp2(-nll)
		n++
	}
	if n == 0 {
		return math.NaN()
	}
	return sum / float64(n)
}

// aucFromScores computes ROC AUC via midranks: pos and neg are natural-log
// probability scores (higher = stronger positive prediction).
func aucFromScores(pos, neg []float64) float64 {
	type item struct {
		score float64
		label int // 1 = positive, 0 = negative
	}
	items := make([]item, 0, len(pos)+len(neg))
	for _, s := range pos {
		items = append(items, item{s, 1})
	}
	for _, s := range neg {
		items = append(items, item{s, 0})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].score != items[j].score {
			return items[i].score < items[j].score
		}
		return items[i].label < items[j].label
	})
	n := len(items)
	ranks := make([]float64, n)
	i := 0
	for i < n {
		j := i
		for j < n && items[j].score == items[i].score {
			j++
		}
		avgRank := float64(i+1+j) / 2.0
		for k := i; k < j; k++ {
			ranks[k] = avgRank
		}
		i = j
	}
	sumPosRank := 0.0
	nPos, nNeg := 0, 0
	for idx, it := range items {
		if it.label == 1 {
			sumPosRank += ranks[idx]
			nPos++
		} else {
			nNeg++
		}
	}
	if nPos == 0 || nNeg == 0 {
		return math.NaN()
	}
	return (sumPosRank - float64(nPos)*(float64(nPos)+1)/2) / (float64(nPos) * float64(nNeg))
}

const pm6BootstrapResamples = 2000

// ComputePM6 scores matched positive/negative pairs and returns the AUC
// plus a deterministic percentile bootstrap CI.
func ComputePM6(model FittedModel, pairs []NegativePair, namespace, transcription, candidateID, modelClass string) (auc float64, ok bool, ciLow, ciHigh float64) {
	if len(pairs) == 0 {
		return math.NaN(), false, math.NaN(), math.NaN()
	}
	posScores := make([]float64, len(pairs))
	negScores := make([]float64, len(pairs))
	for i, pr := range pairs {
		posScores[i] = -model.TokenNegLog2Prob(pr.PositiveRaw, pr.PositiveGlyphs) * math.Ln2
		negScores[i] = -model.TokenNegLog2Prob(pr.NegativeRaw, pr.NegativeGlyphs) * math.Ln2
	}
	auc = aucFromScores(posScores, negScores)
	seed := SeedFields{Namespace: namespace, ModelClass: modelClass, CandidateID: candidateID, CorpusID: "PM6_BOOTSTRAP", Transcription: transcription, Partition: "HELDOUT", Scale: 1.0, Replicate: 0}
	prng := NewSeededPRNG(seed)
	boot := make([]float64, pm6BootstrapResamples)
	n := len(pairs)
	for r := 0; r < pm6BootstrapResamples; r++ {
		rp := make([]float64, n)
		rn := make([]float64, n)
		for i := 0; i < n; i++ {
			idx := int(prng.Float64() * float64(n))
			if idx >= n {
				idx = n - 1
			}
			rp[i] = posScores[idx]
			rn[i] = negScores[idx]
		}
		boot[r] = aucFromScores(rp, rn)
	}
	ciLow = percentileNearestRank(boot, 0.025)
	ciHigh = percentileNearestRank(boot, 0.975)
	return auc, true, ciLow, ciHigh
}
