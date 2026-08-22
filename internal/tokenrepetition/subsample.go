package tokenrepetition

import "math/rand"

// SampleSummary is the set of size-sensitive statistics compared between
// the label corpus and matched-size running-text samples (task60 section
// 30-31): vocabulary size, hapax/type, exact-adjacent-repeat rate, and
// edit-distance<=1 adjacency rate.
type SampleSummary struct {
	V              int
	HapaxFraction  float64
	AdjacentRepeatRate float64
	EditLe1Rate    float64
}

// summarize computes SampleSummary for tokens, respecting lineOfToken's
// natural-line boundaries for the adjacency-sensitive statistics (nil
// means "treat as a single line", i.e. every position is adjacent to the
// next - appropriate only when the caller has already isolated one
// natural unit).
func summarize(tokens []string, lineOfToken []int, glyphMode GlyphMode) SampleSummary {
	freq := map[string]int{}
	for _, t := range tokens {
		freq[t]++
	}
	hapax := 0
	for _, n := range freq {
		if n == 1 {
			hapax++
		}
	}
	var s SampleSummary
	s.V = len(freq)
	if s.V > 0 {
		s.HapaxFraction = float64(hapax) / float64(s.V)
	}
	adj := AdjacentRepetition(tokens, lineOfToken, 0)
	s.AdjacentRepeatRate = adj.R2
	glyphs := GlyphSequences(tokens, glyphMode)
	dists := AdjacentEditDistances(tokens, lineOfToken, glyphs)
	sum := SummarizeDistances(dists, glyphs)
	s.EditLe1Rate = sum.PLe1
	return s
}

// RunningTextSubsampleNull draws `draws` random contiguous spans of
// `size` tokens from running (task60 section 31: matched-N comparison,
// preserving local adjacency structure rather than an unordered bag,
// since the compared statistics are themselves adjacency-sensitive), and
// summarizes each with the same statistics used for the label corpus,
// respecting running's own natural line boundaries within the sampled
// span (lineOfToken, parallel to running).
func RunningTextSubsampleNull(running []string, lineOfToken []int, size, draws int, glyphMode GlyphMode, r *rand.Rand) []SampleSummary {
	if size <= 0 || size > len(running) {
		return nil
	}
	out := make([]SampleSummary, 0, draws)
	maxStart := len(running) - size
	for i := 0; i < draws; i++ {
		start := 0
		if maxStart > 0 {
			start = r.Intn(maxStart + 1)
		}
		out = append(out, summarize(running[start:start+size], lineOfToken[start:start+size], glyphMode))
	}
	return out
}

// SummarizeLabels computes SampleSummary for the label corpus, respecting
// per-label boundaries (task60 section 4: two different labels are never
// treated as textually adjacent).
func SummarizeLabels(labelCorpus Corpus, glyphMode GlyphMode) SampleSummary {
	return summarize(labelCorpus.Tokens, labelCorpus.LineOfToken, glyphMode)
}

// Percentile returns the fraction of null values <= observed (task59's
// same empirical-percentile convention, reused here).
func Percentile(observed float64, null []float64) float64 {
	if len(null) == 0 {
		return 0.5
	}
	le := 0
	for _, v := range null {
		if v <= observed {
			le++
		}
	}
	return float64(le) / float64(len(null))
}
