package tokenrepetition

import (
	"math/rand"
	"sort"
)

// AdjacentStats holds the exact-adjacent-repetition summary (task60
// section 5): R2 and the per-repeated-token breakdown.
type AdjacentStats struct {
	ValidPairs int
	RepeatPairs int
	R2          float64
	Tokens      map[string]*RepeatedToken
}

// sameLine reports whether tokens at positions i and i+1 count as
// adjacent: lineOfToken == nil means "always" (e.g. an already-contiguous
// span with boundaries handled by the caller); otherwise the pair must
// share a natural line, matching Task58's own "adjacent" definition
// (its calc() skips any pair whose line IDs differ) - kept consistent
// here rather than defining "adjacent" a second, different way.
func sameLine(lineOfToken []int, i int) bool {
	return lineOfToken == nil || lineOfToken[i] == lineOfToken[i+1]
}

// AdjacentRepetition computes N(w_i = w_i+1) and R2 = that / valid pairs,
// plus per-token frequency/adjacent-repeat-count/maximum-run/loci
// (task60 section 5). A pair crossing a natural line boundary does not
// count (see sameLine); pass lineOfToken=nil to disable that check (e.g.
// for an already-contiguous single-line span). Loci are capped at
// maxLociPerToken corpus positions to keep the output bounded for very
// frequent tokens.
func AdjacentRepetition(tokens []string, lineOfToken []int, maxLociPerToken int) AdjacentStats {
	freq := map[string]int{}
	for _, t := range tokens {
		freq[t]++
	}
	st := AdjacentStats{Tokens: map[string]*RepeatedToken{}}
	if len(tokens) < 2 {
		return st
	}
	runs := ExactRuns(tokens, lineOfToken)
	maxRun := map[string]int{}
	for _, r := range runs {
		if r.RunLength > maxRun[r.Token] {
			maxRun[r.Token] = r.RunLength
		}
	}
	for i := 0; i+1 < len(tokens); i++ {
		if !sameLine(lineOfToken, i) {
			continue
		}
		st.ValidPairs++
		if tokens[i] == tokens[i+1] {
			st.RepeatPairs++
			rt := st.Tokens[tokens[i]]
			if rt == nil {
				rt = &RepeatedToken{Token: tokens[i], Frequency: freq[tokens[i]], MaximumRun: maxRun[tokens[i]]}
				st.Tokens[tokens[i]] = rt
			}
			rt.AdjacentRepeats++
			if len(rt.FirstLoci) < maxLociPerToken {
				rt.FirstLoci = append(rt.FirstLoci, i)
			}
		}
	}
	if st.ValidPairs > 0 {
		st.R2 = float64(st.RepeatPairs) / float64(st.ValidPairs)
	}
	return st
}

// ExactRuns finds every maximal run w^k (k>=2) in tokens (task60 section
// 6): O(N), no double-counting of overlapping runs (a run is reported
// once, at its start position, with its full maximal length). A run never
// crosses a natural line boundary (see sameLine); pass lineOfToken=nil to
// disable that check.
func ExactRuns(tokens []string, lineOfToken []int) []Run {
	freq := map[string]int{}
	for _, t := range tokens {
		freq[t]++
	}
	var runs []Run
	i := 0
	for i < len(tokens) {
		j := i + 1
		for j < len(tokens) && tokens[j] == tokens[i] && sameLine(lineOfToken, j-1) {
			j++
		}
		length := j - i
		if length >= 2 {
			runs = append(runs, Run{Token: tokens[i], RunLength: length, StartPosition: i, GlobalFrequency: freq[tokens[i]]})
		}
		i = j
	}
	return runs
}

// RunLengthSurvival returns, for k = 2..maxK, the count of runs with
// RunLength >= k (task60 section 7's P(run length >= k), as a raw count;
// the caller divides by whatever denominator - token count or run count -
// the comparison calls for).
func RunLengthSurvival(runs []Run, maxK int) []int {
	out := make([]int, maxK+1) // index by k, 0/1 unused
	for _, r := range runs {
		for k := 2; k <= maxK && k <= r.RunLength; k++ {
			out[k]++
		}
	}
	return out
}

// MaxObservedRun returns the longest run length in runs, or 1 if there are
// none (so callers can size RunLengthSurvival's k-range from real data,
// task60 section 7: "не ограничиваться maximum observed run" is honored by
// always extending a few steps past this value, done by the caller).
func MaxObservedRun(runs []Run) int {
	m := 1
	for _, r := range runs {
		if r.RunLength > m {
			m = r.RunLength
		}
	}
	return m
}

// GlobalShuffle is null model A (task60 section 9): permute the entire
// token sequence, preserving N, V, and exact per-type frequencies exactly,
// discarding line boundaries.
func GlobalShuffle(tokens []string, r *rand.Rand) []string {
	out := append([]string(nil), tokens...)
	r.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

// WithinLineShuffle is null model B (task60 section 10): permute tokens
// within each natural line independently, preserving line lengths and
// (per-line) composition exactly. lineOfToken must be non-decreasing (as
// produced by internal/genericsegmentation.ReadCorpus).
func WithinLineShuffle(tokens []string, lineOfToken []int, r *rand.Rand) []string {
	out := append([]string(nil), tokens...)
	start := 0
	for start < len(out) {
		end := start
		for end < len(out) && lineOfToken[end] == lineOfToken[start] {
			end++
		}
		seg := out[start:end]
		r.Shuffle(len(seg), func(i, j int) { seg[i], seg[j] = seg[j], seg[i] })
		start = end
	}
	return out
}

// sortedTokenKeys returns m's keys sorted, for deterministic float64
// accumulation order (see internal/evaglyph and project memory on sorting
// map keys before summing).
func sortedTokenKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
