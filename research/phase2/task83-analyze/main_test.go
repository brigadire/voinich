package main

import (
	"math"
	"testing"
)

func TestCompareCoveragePenalty(t *testing.T) {
	target := map[string]float64{"EF1_GIANT_COMPONENT_SHARE": 1, "EF2_GLOBAL_CLUSTERING": 2, "EF3_DEGREE_FREQUENCY_SPEARMAN": 3}
	e := endpoint{Values: map[string]float64{"EF1_GIANT_COMPONENT_SHARE": 0}, Available: map[string]bool{"EF1_GIANT_COMPONENT_SHARE": true}}
	d, a, n := compareOne(target, e, map[string]float64{"EF1_GIANT_COMPONENT_SHARE": 1})
	if d != 1 || a != 3 || n != 1 {
		t.Fatalf("got %v %v %v", d, a, n)
	}
}

func TestCompareNoCoverageIsUnavailable(t *testing.T) {
	d, a, n := compareOne(map[string]float64{"EF1_GIANT_COMPONENT_SHARE": 1}, endpoint{Values: map[string]float64{}, Available: map[string]bool{}}, map[string]float64{"EF1_GIANT_COMPONENT_SHARE": 1})
	if !math.IsNaN(d) || !math.IsNaN(a) || n != 0 {
		t.Fatalf("missing coverage must stay unavailable: %v %v %d", d, a, n)
	}
}
func TestVectorStats(t *testing.T) {
	x := map[string]float64{"EF1_GIANT_COMPONENT_SHARE": 2, "EF2_GLOBAL_CLUSTERING": 4, "EF3_DEGREE_FREQUENCY_SPEARMAN": 6}
	n := map[string]float64{"EF1_GIANT_COMPONENT_SHARE": 1, "EF2_GLOBAL_CLUSTERING": 2, "EF3_DEGREE_FREQUENCY_SPEARMAN": 3}
	tr := traj{Values: map[string]float64{"EF1_GIANT_COMPONENT_SHARE": 1, "EF2_GLOBAL_CLUSTERING": 2, "EF3_DEGREE_FREQUENCY_SPEARMAN": 3}, Available: map[string]bool{"EF1_GIANT_COMPONENT_SHARE": true, "EF2_GLOBAL_CLUSTERING": true, "EF3_DEGREE_FREQUENCY_SPEARMAN": true}}
	sc := map[string]float64{"EF1_GIANT_COMPONENT_SHARE": 1, "EF2_GLOBAL_CLUSTERING": 1, "EF3_DEGREE_FREQUENCY_SPEARMAN": 1}
	c, s, m, k := vectorStats(x, n, tr, sc)
	if math.Abs(c-1) > 1e-12 || s != 1 || m != 1 || k != 3 {
		t.Fatalf("got %v %v %v %v", c, s, m, k)
	}
}
