package mechanismspace

import "sort"

// Direction is task66 section 35's preregistered "which way is toward
// Voynich" for one metric.
type Direction int

const (
	Higher Direction = iota // Voynich value is higher than baseline
	Lower                   // Voynich value is lower than baseline
)

// MetricTarget is one preregistered (baseline, target, direction) triple
// (task66 section 35): baseline is the untransformed plaintext's value
// (M0), target is the authoritative Voynich value.
type MetricTarget struct {
	Family    string
	Metric    string
	Baseline  float64
	Voynich   float64
	Direction Direction
}

// OvershootClass is task66 section 36's three-way bucket.
type OvershootClass string

const (
	Undershoot  OvershootClass = "UNDERSHOOT"
	TargetRange OvershootClass = "TARGET_RANGE"
	Overshoot   OvershootClass = "OVERSHOOT"
)

// Progress is task66 section 35's normalized-progress statistic for one
// metric, plus its overshoot classification (section 36).
type Progress struct {
	Family, Metric                      string
	Baseline, Voynich, Candidate, Value float64
	Clipped                             bool
	Overshoot                           OvershootClass
}

// ComputeProgress computes Progress = (candidate-baseline)/(voynich-
// baseline) with a documented clipping diagnostic (section 35's "не
// использовать без clipping diagnostics"): progress is reported
// unclipped, but Clipped flags |Progress|>1.5 so a report can call out an
// extreme value rather than silently averaging it away.
func ComputeProgress(t MetricTarget, candidate float64) Progress {
	denom := t.Voynich - t.Baseline
	p := Progress{Family: t.Family, Metric: t.Metric, Baseline: t.Baseline, Voynich: t.Voynich, Candidate: candidate}
	if denom == 0 {
		p.Value = 0
		p.Overshoot = TargetRange
		return p
	}
	p.Value = (candidate - t.Baseline) / denom
	if p.Value < -1.5 || p.Value > 1.5 {
		p.Clipped = true
	}
	switch {
	case p.Value < 0.85:
		p.Overshoot = Undershoot
	case p.Value <= 1.15:
		p.Overshoot = TargetRange
	default:
		p.Overshoot = Overshoot
	}
	return p
}

// FamilyScore is task66 section 37's robust per-family aggregate: the
// median normalized progress across every metric preregistered for that
// family, so a family with many correlated dimensions cannot dominate.
func FamilyScore(progresses []Progress) map[string]float64 {
	byFamily := map[string][]float64{}
	for _, p := range progresses {
		byFamily[p.Family] = append(byFamily[p.Family], p.Value)
	}
	out := map[string]float64{}
	for f, vs := range byFamily {
		out[f] = medianF(vs)
	}
	return out
}

func medianF(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	s := append([]float64(nil), v...)
	sort.Float64s(s)
	n := len(s)
	if n%2 == 1 {
		return s[n/2]
	}
	return (s[n/2-1] + s[n/2]) / 2
}

// Dominates is task66 section 38's Pareto dominance rule: a dominates b
// if a is not worse on every family and strictly better on at least one.
// "Not worse" is judged on |1-progress| (closeness to Voynich, not
// literal magnitude), so overshoot is not rewarded as if it were extra
// progress (section 36).
func Dominates(a, b map[string]float64) bool {
	strictlyBetter := false
	families := make([]string, 0, len(a))
	for family := range a {
		families = append(families, family)
	}
	sort.Strings(families)
	for _, family := range families {
		av := a[family]
		bv, ok := b[family]
		if !ok {
			continue
		}
		ad, bd := closeness(av), closeness(bv)
		if ad < bd { // a is worse (farther from target) on this family
			return false
		}
		if ad > bd {
			strictlyBetter = true
		}
	}
	return strictlyBetter
}

func closeness(progress float64) float64 {
	d := progress - 1
	if d < 0 {
		d = -d
	}
	return -d // higher is closer to Voynich (progress==1)
}

// ParetoFront returns the indices of candidates not dominated by any
// other candidate (task66 section 38).
func ParetoFront(scores []map[string]float64) []int {
	var front []int
	for i := range scores {
		dominated := false
		for j := range scores {
			if i == j {
				continue
			}
			if Dominates(scores[j], scores[i]) {
				dominated = true
				break
			}
		}
		if !dominated {
			front = append(front, i)
		}
	}
	return front
}
