package residualdiagnostic

import (
	"math"
	"reflect"
	"testing"
)

func TestTrainingResidualMeanApproximatelyZero(t *testing.T) {
	x := []sparse{{"a": 1}, {"a": 3, "b": 2}, {"b": 1}}
	mu := meanSparse(x)
	r := make([]sparse, len(x))
	for i := range x {
		r[i] = subtract(x[i], mu)
	}
	if got := normOf(meanSparse(r)).Linf; got > 1e-14 {
		t.Fatalf("training residual mean=%g", got)
	}
}

func TestHeldOutNeverUsedForMeanOrCovariance(t *testing.T) {
	train := []sparse{{"a": 1}, {"a": 3, "b": 1}, {"a": 2, "b": 2}}
	mu := meanSparse(train)
	model := fitWhitening(train)
	held := sparse{"a": 1000, "new": 50}
	_ = subtract(held, mu)
	_ = model.apply(held)
	changed := sparse{"a": -1000, "new": 900}
	_ = subtract(changed, mu)
	_ = model.apply(changed)
	if !reflect.DeepEqual(mu, meanSparse(train)) {
		t.Fatal("held-out observation changed training mean")
	}
	model2 := fitWhitening(train)
	if !reflect.DeepEqual(model.diag, model2.diag) || !reflect.DeepEqual(model.eigVal, model2.eigVal) {
		t.Fatal("held-out observation affected covariance fit")
	}
}

func TestCovarianceRegularizationAndDeterministicWhitening(t *testing.T) {
	x := []sparse{{"constant": 1, "a": 0}, {"constant": 1, "a": 1}, {"constant": 1, "a": 2}}
	a, b := fitWhitening(x), fitWhitening(x)
	z1, z2 := a.apply(sparse{"constant": 1, "a": 3}), b.apply(sparse{"constant": 1, "a": 3})
	if !reflect.DeepEqual(z1, z2) {
		t.Fatalf("non-deterministic whitening: %v %v", z1, z2)
	}
	for _, v := range z1 {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			t.Fatalf("regularization produced %g", v)
		}
	}
}

func TestPhysicalBlockSplitHasNoOverlap(t *testing.T) {
	bs := []block{{ID: "a", Start: 0, End: 10}, {ID: "b", Start: 10, End: 20}, {ID: "c", Start: 20, End: 30}}
	for _, f := range folds(bs) {
		seen := map[string]bool{}
		for _, b := range f.train {
			seen[b.ID] = true
		}
		for _, b := range f.test {
			if seen[b.ID] {
				t.Fatalf("block %s occurs in train and test", b.ID)
			}
		}
	}
}

func TestNormCalculation(t *testing.T) {
	n := normOf(sparse{"a": -3, "b": 4})
	if n.L1 != 7 || n.L2 != 5 || n.Linf != 4 {
		t.Fatalf("bad norms: %+v", n)
	}
}

func TestClusterRunExtraction(t *testing.T) {
	w := []window{{Block: "a", Start: 0}, {Block: "a", Start: 10}, {Block: "a", Start: 20}, {Block: "b", Start: 30}}
	r := extractRuns(w, []int{0, 0, 1, 1})
	if len(r) != 2 || r[0].Count != 1 || r[0].Largest != 2 || r[1].Count != 2 {
		t.Fatalf("bad runs: %+v", r)
	}
}

func TestBlockRecurrence(t *testing.T) {
	if bothRegimesRecur(map[int]bool{0: true}, 2) {
		t.Fatal("single regime declared recurrent")
	}
	if !bothRegimesRecur(map[int]bool{0: true, 1: true}, 2) {
		t.Fatal("two regimes not recognized")
	}
}

func TestNormalizedBlockPosition(t *testing.T) {
	b := block{Start: 100, End: 200}
	if got := normalizedBlockPosition(125, b); got != .25 {
		t.Fatalf("got %g", got)
	}
}

func TestBlockAwarePermutationAndSeed(t *testing.T) {
	labels := []string{"A", "A", "B", "B"}
	blocks := []string{"x", "x", "y", "y"}
	a := blockPermutation(labels, blocks, 7)
	b := blockPermutation(labels, blocks, 7)
	if !reflect.DeepEqual(a, b) {
		t.Fatal("permutation is not deterministic")
	}
	if a[0] != a[1] || a[2] != a[3] {
		t.Fatalf("block grouping broken: %v", a)
	}
	v := []sparse{{"x": 0}, {"x": .1}, {"x": 10}, {"x": 11}}
	if !reflect.DeepEqual(cluster(v, 2, 9), cluster(v, 2, 9)) {
		t.Fatal("clustering seed is not deterministic")
	}
}
