package localregime

import (
	"math/rand"
	"reflect"
	"testing"
)

// referenceDistanceDistributionMode is distanceDistributionMode exactly as
// it stood before the distanceDistributionAt hoist: it recomputes
// positions(c,t) on every call, even though it doesn't depend on d.
func referenceDistanceDistributionMode(c corpus, t string, d int, respect bool) profile {
	m := map[string]int{}
	for _, i := range positions(c, t) {
		j := i + d
		if j < len(c.Tokens) && (!respect || c.LineAt[j] == c.LineAt[i]) {
			m[c.Tokens[j]]++
		}
	}
	return normalizeCounts(m)
}

// fixtureCorpus builds a synthetic corpus with nLines lines of varying
// length drawn from a small shared vocabulary, so a given token occurs
// repeatedly with varied neighbors at multiple distances.
func fixtureCorpus(nLines int, seed int64) corpus {
	r := rand.New(rand.NewSource(seed))
	vocab := []string{"aiin", "chey", "shey", "ol", "or", "dy", "qokeey", "s"}
	var tokens []string
	var lineAt []int
	for li := 0; li < nLines; li++ {
		length := 3 + r.Intn(12)
		for k := 0; k < length; k++ {
			tokens = append(tokens, vocab[r.Intn(len(vocab))])
			lineAt = append(lineAt, li)
		}
	}
	return corpus{Tokens: tokens, LineAt: lineAt}
}

// TestDistanceDistributionAtHoistMatchesReference proves the
// distanceDistributionAt hoist produces byte-identical output to the
// pre-hoist reference (which recomputes positions(c,t) on every call),
// across several corpus sizes, tokens, distances, and both
// respect-line-boundaries settings.
func TestDistanceDistributionAtHoistMatchesReference(t *testing.T) {
	vocab := []string{"aiin", "chey", "shey", "ol", "or", "dy", "qokeey", "s"}
	for _, nLines := range []int{0, 3, 40} {
		c := fixtureCorpus(nLines, int64(nLines)*97+11)
		for _, tok := range vocab {
			pos := positions(c, tok)
			for d := 1; d <= 5; d++ {
				for _, respect := range []bool{false, true} {
					want := referenceDistanceDistributionMode(c, tok, d, respect)
					got := distanceDistributionAt(c, pos, d, respect)
					if !reflect.DeepEqual(got, want) {
						t.Fatalf("nLines=%d tok=%s d=%d respect=%v: diverged\ngot=%v\nwant=%v", nLines, tok, d, respect, got, want)
					}
				}
			}
		}
	}
}

func benchLocalregimeCorpus() corpus {
	return fixtureCorpus(2000, 42)
}

func BenchmarkDistanceSweepReference(b *testing.B) {
	c := benchLocalregimeCorpus()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for d := 1; d <= 20; d++ {
			referenceDistanceDistributionMode(c, "aiin", d, false)
		}
	}
}

func BenchmarkDistanceSweepHoisted(b *testing.B) {
	c := benchLocalregimeCorpus()
	pos := positions(c, "aiin")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for d := 1; d <= 20; d++ {
			distanceDistributionAt(c, pos, d, false)
		}
	}
}
