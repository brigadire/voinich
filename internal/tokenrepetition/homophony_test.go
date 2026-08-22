package tokenrepetition

import "testing"

func TestTheoreticalRunSurvivalUniform(t *testing.T) {
	// (1/H)^(k-1): H=4, k=3 -> (1/4)^2 = 0.0625
	got := TheoreticalRunSurvival(3, UniformWeights(4))
	if diff := got - 0.0625; diff > 1e-9 || diff < -1e-9 {
		t.Fatalf("got %v, want 0.0625", got)
	}
	// k=1 run always "survives" trivially
	if got := TheoreticalRunSurvival(1, UniformWeights(4)); got != 1 {
		t.Fatalf("k=1 should be 1, got %v", got)
	}
}

func TestTheoreticalRunSurvivalWeightedSumsToLessThanUniformForSkewedWeights(t *testing.T) {
	tri := TriangularWeights(4)
	sum := 0.0
	for _, w := range tri {
		sum += w
	}
	if diff := sum - 1.0; diff > 1e-9 || diff < -1e-9 {
		t.Fatalf("triangular weights must sum to 1, got %v", sum)
	}
	// skewed weights make same-homophone survival MORE likely than uniform
	// for k>=2 (Jensen's inequality on sum p_j^k, convex in p for k>=2).
	uni := TheoreticalRunSurvival(3, UniformWeights(4))
	skew := TheoreticalRunSurvival(3, tri)
	if skew <= uni {
		t.Fatalf("expected skewed weights to raise survival probability: skew=%v uni=%v", skew, uni)
	}
}

func TestRunSurvivalDoseResponseDetectsBrokenRun(t *testing.T) {
	plain := []string{"a", "a", "a", "b"}
	cipher := []string{"x1", "x1", "x2", "y1"} // run of 3 a's -> only 2 survive as x1,x1
	rows := RunSurvivalDoseResponse(plain, cipher, nil, func(string) int { return 2 }, UniformWeights)
	if len(rows) != 1 {
		t.Fatalf("expected 1 run, got %d", len(rows))
	}
	if rows[0].Survived {
		t.Fatal("expected the broken run to be marked as not survived")
	}
}

func TestRunSurvivalDoseResponseDetectsIntactRun(t *testing.T) {
	plain := []string{"a", "a", "a", "b"}
	cipher := []string{"x1", "x1", "x1", "y1"}
	rows := RunSurvivalDoseResponse(plain, cipher, nil, func(string) int { return 2 }, UniformWeights)
	if len(rows) != 1 || !rows[0].Survived {
		t.Fatalf("expected the intact run to be marked as survived: %+v", rows)
	}
}
