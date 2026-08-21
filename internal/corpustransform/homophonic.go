package corpustransform

import (
	"fmt"
	"sort"
)

// HomophonicParams fully determines a deterministic token-level homophonic
// substitution (task46 sections 5-7).
type HomophonicParams struct {
	Model      string // fixed (global H) or frequency (frequency-v1, Hmax)
	Homophones int    // H or Hmax, depending on Model
	Selection  string // SelectionUniform or SelectionWeighted
	Seed       int64
}

// HomophoneEntry is one cipher token assigned to a plaintext token, with
// its selection probability.
type HomophoneEntry struct {
	CipherToken string
	Probability float64
}

// Mapping is the full plaintext -> homophone-list assignment, plus the
// opaque per-token PRNG substream state needed to reproduce occurrence
// selection. Entries is keyed by plaintext token; Vocabulary preserves the
// deterministic (sorted) assignment order used to derive opaque IDs and
// mapping.tsv row order.
type Mapping struct {
	Vocabulary []string
	Entries    map[string][]HomophoneEntry
	Allocation []AllocationRecord
}

// AllocationRecord is the frozen frequency-v1 allocation decision for one
// plaintext type. Rank is one-based, descending by frequency (lexical order
// breaks ties), and Quantile is in [0,1].
type AllocationRecord struct {
	PlaintextToken    string
	RawFrequency      int
	FrequencyRank     int
	FrequencyQuantile float64
	AllocatedH        int
}

const FrequencyModelVersion = "frequency-v1"
const GlobalModelVersion = "homophonic-global-v1"

// WeightScheme is the versioned, fixed formula used by
// -homophone-selection weighted (task46 section 6): "triangular-v1" gives
// homophone index k (0-based, in opaque-ID assignment order) weight
// (H-k) / (H*(H+1)/2), i.e. strictly decreasing, normalized to 1. It does
// not depend on the corpus, the plaintext token, or Voynich statistics -
// only on H.
const WeightScheme = "triangular-v1"

func triangularWeights(h int) []float64 {
	denom := float64(h*(h+1)) / 2
	weights := make([]float64, h)
	for k := range h {
		weights[k] = float64(h-k) / denom
	}
	return weights
}

func uniformWeights(h int) []float64 {
	weights := make([]float64, h)
	w := 1.0 / float64(h)
	for k := range weights {
		weights[k] = w
	}
	return weights
}

// BuildMapping assigns H opaque, non-plaintext-revealing cipher tokens to
// every distinct token in tokens (task46 section 5-6). Vocabulary order is
// the sorted token order, never map iteration order, so opaque IDs and
// mapping.tsv rows are fully deterministic (see the "Go map iteration
// determinism" project convention).
func BuildMapping(tokens []string, p HomophonicParams) (Mapping, error) {
	if p.Homophones < 1 {
		return Mapping{}, fmt.Errorf("-homophones must be >= 1, got %d", p.Homophones)
	}
	if p.Selection != SelectionUniform && p.Selection != SelectionWeighted {
		return Mapping{}, fmt.Errorf("unsupported homophone selection %q", p.Selection)
	}

	seen := make(map[string]struct{})
	for _, t := range tokens {
		seen[t] = struct{}{}
	}
	vocab := make([]string, 0, len(seen))
	for t := range seen {
		vocab = append(vocab, t)
	}
	sort.Strings(vocab)
	counts := TokenCounts(tokens)
	allocation := make([]AllocationRecord, 0, len(vocab))
	if p.Model == HomophoneModelFrequency {
		// Vocab is already lexical order. Sort a copy by raw frequency first;
		// lexical order is a deterministic tie-breaker.
		sort.SliceStable(vocab, func(i, j int) bool {
			if counts[vocab[i]] != counts[vocab[j]] {
				return counts[vocab[i]] > counts[vocab[j]]
			}
			return vocab[i] < vocab[j]
		})
	} else if p.Model != HomophoneModelFixed {
		return Mapping{}, fmt.Errorf("unsupported homophone model %q", p.Model)
	}

	entries := make(map[string][]HomophoneEntry, len(vocab))
	counter := 0
	for rank, t := range vocab {
		h := p.Homophones
		q := 0.0
		if p.Model == HomophoneModelFrequency {
			if len(vocab) > 1 {
				q = float64(len(vocab)-rank-1) / float64(len(vocab)-1)
			}
			h = 1 + int(float64(p.Homophones-1)*q)
		}
		allocation = append(allocation, AllocationRecord{PlaintextToken: t, RawFrequency: counts[t], FrequencyRank: rank + 1, FrequencyQuantile: q, AllocatedH: h})
		weights := uniformWeights(h)
		if p.Selection == SelectionWeighted {
			weights = triangularWeights(h)
		}
		list := make([]HomophoneEntry, h)
		for k := range h {
			list[k] = HomophoneEntry{CipherToken: opaqueID(counter), Probability: weights[k]}
			counter++
		}
		entries[t] = list
	}
	// Keep mapping.tsv order lexical for compatibility, while allocation rank
	// remains explicit in the diagnostic artifact.
	sort.Strings(vocab)
	return Mapping{Vocabulary: vocab, Entries: entries, Allocation: allocation}, nil
}

func opaqueID(counter int) string {
	return fmt.Sprintf("x%06d", counter)
}

// Encode substitutes every occurrence of a plaintext token with one of its
// assigned homophones, chosen by a single deterministic PRNG stream
// consumed in original left-to-right token order (task46 section 5, 8).
func Encode(tokens []string, mapping Mapping, seed int64) []string {
	r := subRand("homophonic-occurrence-selection", seed, 0)
	out := make([]string, len(tokens))
	for i, t := range tokens {
		list := mapping.Entries[t]
		out[i] = list[selectHomophone(list, r.Float64())].CipherToken
	}
	return out
}

func selectHomophone(list []HomophoneEntry, draw float64) int {
	cum := 0.0
	for k, e := range list {
		cum += e.Probability
		if draw < cum {
			return k
		}
	}
	return len(list) - 1
}

// Decode inverts Encode using mapping, used by round-trip tests (task46
// section 18) to prove reversibility.
func Decode(cipherTokens []string, mapping Mapping) ([]string, error) {
	reverse := make(map[string]string, len(mapping.Vocabulary)*4)
	for _, t := range mapping.Vocabulary {
		for _, e := range mapping.Entries[t] {
			reverse[e.CipherToken] = t
		}
	}
	out := make([]string, len(cipherTokens))
	for i, c := range cipherTokens {
		t, ok := reverse[c]
		if !ok {
			return nil, fmt.Errorf("corpustransform: cipher token %q has no plaintext mapping", c)
		}
		out[i] = t
	}
	return out, nil
}

// MappingCollisions counts cipher tokens that were assigned to more than
// one plaintext token. A correct mapping always has zero.
func MappingCollisions(mapping Mapping) int {
	owner := make(map[string]string, len(mapping.Vocabulary))
	collisions := 0
	for _, t := range mapping.Vocabulary {
		for _, e := range mapping.Entries[t] {
			if prev, ok := owner[e.CipherToken]; ok && prev != t {
				collisions++
				continue
			}
			owner[e.CipherToken] = t
		}
	}
	return collisions
}
