package propertytrajectory

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

// referenceFallbackMatched is fallbackMatched exactly as it stood before
// the matchWorkspace rewrite: it rebuilds the full eligible-token candidate
// pool and recomputes every math.Log1p from scratch on every call, even
// though the pool and each token's log1p are invariant across every one of
// the ~40-80 fallbackMatched calls in one analyze() run.
func referenceFallbackMatched(target pair, c corpus, eligible []string, n int, r *rand.Rand) []pair {
	score := func(p pair) float64 {
		return math.Abs(math.Log1p(float64(c.Counts[p.A]))-math.Log1p(float64(c.Counts[target.A]))) + math.Abs(math.Log1p(float64(c.Counts[p.B]))-math.Log1p(float64(c.Counts[target.B])))
	}
	pool := make([]pair, 0)
	for i := 0; i < len(eligible); i++ {
		for j := i + 1; j < len(eligible); j++ {
			p := pair{eligible[i], eligible[j]}
			if p.A == target.A || p.A == target.B || p.B == target.A || p.B == target.B {
				continue
			}
			pool = append(pool, p)
		}
	}
	sort.Slice(pool, func(i, j int) bool {
		a, b := score(pool[i]), score(pool[j])
		if a == b {
			return pool[i].A+pool[i].B < pool[j].A+pool[j].B
		}
		return a < b
	})
	limit := min(len(pool), max(n*10, n))
	pool = pool[:limit]
	r.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	return pool[:min(n, len(pool))]
}

// fixtureEligibleCorpus builds a synthetic corpus with nTokens distinct
// eligible tokens, each with a distinct, varied count, so fallbackMatched's
// frequency-matching score is non-degenerate.
func fixtureEligibleCorpus(nTokens int, seed int64) (corpus, []string) {
	r := rand.New(rand.NewSource(seed))
	c := corpus{Counts: map[string]int{}}
	eligible := make([]string, nTokens)
	for i := 0; i < nTokens; i++ {
		tok := fmt.Sprintf("tok%03d", i)
		eligible[i] = tok
		count := 1 + r.Intn(500)
		c.Counts[tok] = count
		for k := 0; k < count; k++ {
			c.Tokens = append(c.Tokens, tok)
		}
	}
	sort.Strings(eligible)
	return c, eligible
}

func TestFallbackMatchedHoistMatchesReference(t *testing.T) {
	for _, n := range []int{2, 10, 60} {
		c, eligible := fixtureEligibleCorpus(n, int64(n)*97+11)
		ws := prepareMatchWorkspace(c, eligible)
		for _, target := range []pair{{eligible[0], eligible[1%len(eligible)]}, {eligible[len(eligible)-1], eligible[0]}} {
			for _, limitN := range []int{1, 5, 20} {
				want := referenceFallbackMatched(target, c, eligible, limitN, rand.New(rand.NewSource(7)))
				got := fallbackMatched(target, c, ws, limitN, rand.New(rand.NewSource(7)))
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("n=%d target=%v limitN=%d: diverged\ngot=%v\nwant=%v", n, target, limitN, got, want)
				}
			}
		}
	}
}

// TestFallbackMatchedHoistMatchesReferenceTargetBelowEligibility exercises
// a target token that is NOT itself in the eligible set (ws.logCount would
// not have an entry for it) - the scenario that requires targetLogA/
// targetLogB to be computed directly rather than read from the cache.
func TestFallbackMatchedHoistMatchesReferenceTargetBelowEligibility(t *testing.T) {
	c, eligible := fixtureEligibleCorpus(30, 55)
	c.Counts["rare"] = 2
	target := pair{"rare", eligible[3]}
	ws := prepareMatchWorkspace(c, eligible)
	want := referenceFallbackMatched(target, c, eligible, 15, rand.New(rand.NewSource(3)))
	got := fallbackMatched(target, c, ws, 15, rand.New(rand.NewSource(3)))
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("diverged\ngot=%v\nwant=%v", got, want)
	}
}

func benchEligibleCorpus() (corpus, []string) {
	return fixtureEligibleCorpus(539, 42)
}

func BenchmarkFallbackMatchedReference(b *testing.B) {
	c, eligible := benchEligibleCorpus()
	targets := []pair{{eligible[0], eligible[1]}, {eligible[10], eligible[20]}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, target := range targets {
			referenceFallbackMatched(target, c, eligible, 1000, rand.New(rand.NewSource(int64(i))))
		}
	}
}

func BenchmarkFallbackMatchedHoisted(b *testing.B) {
	c, eligible := benchEligibleCorpus()
	ws := prepareMatchWorkspace(c, eligible)
	targets := []pair{{eligible[0], eligible[1]}, {eligible[10], eligible[20]}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, target := range targets {
			fallbackMatched(target, c, ws, 1000, rand.New(rand.NewSource(int64(i))))
		}
	}
}
