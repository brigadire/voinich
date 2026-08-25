package mechanismspace

import "testing"

func TestParetoDominanceIgnoresMapInsertionOrder(t *testing.T) {
	a1 := map[string]float64{}
	a1["boundary"] = 0.8
	a1["line"] = 0.7
	a1["edit family"] = 0.9
	a2 := map[string]float64{}
	a2["edit family"] = 0.9
	a2["line"] = 0.7
	a2["boundary"] = 0.8
	b1 := map[string]float64{}
	b1["line"] = 0.6
	b1["edit family"] = 0.8
	b1["boundary"] = 0.8
	b2 := map[string]float64{}
	b2["boundary"] = 0.8
	b2["edit family"] = 0.8
	b2["line"] = 0.6

	if Dominates(a1, b1) != Dominates(a2, b2) || !Dominates(a1, b1) {
		t.Fatal("Pareto dominance changed under equivalent map insertion orders")
	}
}
