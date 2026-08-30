package notation

import (
	"fmt"
	"math"
	"sort"
)

type DeltaRow struct {
	MetricID, Family, Regime string
	Left, Right, Delta       float64
	Status                   Status
	Reason                   string
}

func NotationDelta(left, right Fingerprint) []DeltaRow {
	rm := map[string]Metric{}
	for _, m := range right.Metrics {
		rm[m.MetricID+"\x1f"+m.Regime] = m
	}
	var out []DeltaRow
	for _, l := range left.Metrics {
		k := l.MetricID + "\x1f" + l.Regime
		r, ok := rm[k]
		d := DeltaRow{MetricID: l.MetricID, Family: l.Family, Regime: l.Regime, Left: l.Value, Status: NotComparable}
		if !ok {
			d.Reason = "metric absent from paired representation"
		} else if l.Status != Comparable || r.Status != Comparable {
			d.Reason = "one side is not comparable"
		} else {
			d.Right = r.Value
			d.Delta = r.Value - l.Value
			d.Status = Comparable
		}
		out = append(out, d)
	}
	return out
}

func JensenShannon(p, q []float64) (float64, error) {
	if len(p) != len(q) || len(p) == 0 {
		return 0, fmt.Errorf("distributions must have equal non-zero length")
	}
	pn, qn := normalizeDistribution(p), normalizeDistribution(q)
	var js float64
	for i := range pn {
		m := (pn[i] + qn[i]) / 2
		if pn[i] > 0 {
			js += .5 * pn[i] * math.Log2(pn[i]/m)
		}
		if qn[i] > 0 {
			js += .5 * qn[i] * math.Log2(qn[i]/m)
		}
	}
	return js, nil
}
func Wasserstein1(xp, p, xq, q []float64) (float64, error) {
	if len(xp) != len(p) || len(xq) != len(q) || len(p) == 0 || len(q) == 0 {
		return 0, fmt.Errorf("support/value length mismatch")
	}
	type point struct {
		x, w float64
		side int
	}
	var pts []point
	pn, qn := normalizeDistribution(p), normalizeDistribution(q)
	for i, x := range xp {
		pts = append(pts, point{x, pn[i], 0})
	}
	for i, x := range xq {
		pts = append(pts, point{x, qn[i], 1})
	}
	sort.Slice(pts, func(i, j int) bool { return pts[i].x < pts[j].x })
	var cp, cq, area float64
	for i, z := range pts {
		if i > 0 {
			area += math.Abs(cp-cq) * (z.x - pts[i-1].x)
		}
		if z.side == 0 {
			cp += z.w
		} else {
			cq += z.w
		}
	}
	return area, nil
}
func NormalizedCurveArea(x, a, b []float64) (float64, error) {
	if len(x) != len(a) || len(x) != len(b) || len(x) < 2 {
		return 0, fmt.Errorf("curves require matching lengths >=2")
	}
	var area, scale float64
	for i := 1; i < len(x); i++ {
		dx := x[i] - x[i-1]
		if dx <= 0 {
			return 0, fmt.Errorf("curve support must increase")
		}
		area += dx * (math.Abs(a[i-1]-b[i-1]) + math.Abs(a[i]-b[i])) / 2
		scale += dx * (math.Max(math.Abs(a[i-1]), math.Abs(b[i-1])) + math.Max(math.Abs(a[i]), math.Abs(b[i]))) / 2
	}
	return safe(area, scale), nil
}
func normalizeDistribution(x []float64) []float64 {
	out := make([]float64, len(x))
	var s float64
	for _, v := range x {
		if v > 0 {
			s += v
		}
	}
	if s == 0 {
		return out
	}
	for i, v := range x {
		if v > 0 {
			out[i] = v / s
		}
	}
	return out
}

type Scale struct {
	MetricID, Regime string
	Center, Spread   float64
}
type ComparisonRow struct {
	MetricID, Family, Regime       string
	Candidate, Reference, Distance float64
	Status                         Status
	Reason                         string
}
type FamilyDistance struct {
	Family            string
	Distance          float64
	ComparableMetrics int
	Status            Status
	Reason            string
}

// Compare uses a pre-frozen scale. It never estimates a scale from either
// input and never emits a cross-family total score.
func Compare(candidate, reference Fingerprint, scales []Scale) ([]ComparisonRow, []FamilyDistance, error) {
	cv, rv := candidate.Metadata["metric_registry_version"], reference.Metadata["metric_registry_version"]
	if cv == "" || rv == "" || cv != rv {
		return nil, nil, fmt.Errorf("metric registry version mismatch: candidate=%q reference=%q", cv, rv)
	}
	ref := map[string]Metric{}
	for _, m := range reference.Metrics {
		ref[m.MetricID+"\x1f"+m.Regime] = m
	}
	sm := map[string]Scale{}
	for _, s := range scales {
		if s.Spread <= 0 {
			return nil, nil, fmt.Errorf("scale %s/%s has non-positive spread", s.MetricID, s.Regime)
		}
		sm[s.MetricID+"\x1f"+s.Regime] = s
	}
	var rows []ComparisonRow
	for _, c := range candidate.Metrics {
		k := c.MetricID + "\x1f" + c.Regime
		r, ok := ref[k]
		row := ComparisonRow{MetricID: c.MetricID, Family: c.Family, Regime: c.Regime, Candidate: c.Value, Status: NotComparable}
		if !ok {
			row.Reason = "metric absent from frozen VM reference"
		} else if c.Status != Comparable || r.Status != Comparable {
			row.Reason = "candidate or reference metric is not comparable"
		} else if s, ok := sm[k]; !ok {
			row.Reason = "metric absent from pre-frozen scale"
		} else {
			row.Reference = r.Value
			row.Distance = math.Abs((c.Value-s.Center)/s.Spread - (r.Value-s.Center)/s.Spread)
			row.Status = Comparable
		}
		rows = append(rows, row)
	}
	by := map[string][]float64{}
	for _, r := range rows {
		if r.Status == Comparable {
			by[r.Family] = append(by[r.Family], r.Distance)
		}
	}
	var fam []FamilyDistance
	for _, f := range []string{"G", "T", "S", "L", "D"} {
		x := by[f]
		d := FamilyDistance{Family: f, Status: NotComparable, Reason: "no mutually comparable scaled metrics"}
		if len(x) > 0 {
			d.Status = Comparable
			d.Reason = ""
			d.Distance = mean(x)
			d.ComparableMetrics = len(x)
		}
		fam = append(fam, d)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Family != rows[j].Family {
			return rows[i].Family < rows[j].Family
		}
		if rows[i].MetricID != rows[j].MetricID {
			return rows[i].MetricID < rows[j].MetricID
		}
		return rows[i].Regime < rows[j].Regime
	})
	return rows, fam, nil
}
