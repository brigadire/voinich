// Package inversetransposition contains the deliberately small, controlled
// inverse search used by task54.  It has no language model and no reference
// corpus in its objective.
package inversetransposition

import (
	"fmt"
	"math"
	"sort"

	"zcore.dev/voinich/internal/corpustransform"
)

const ObjectiveVersion = "structural-v2"

type Candidate struct {
	Width  int    `json:"width"`
	Order  string `json:"order"`
	Rounds int    `json:"rounds"`
	Seed   int64  `json:"seed"`
}

func (c Candidate) ID() string { return fmt.Sprintf("w%03d-%s-r%02d", c.Width, c.Order, c.Rounds) }

func (c Candidate) Apply(tokens []string) ([]string, error) {
	return corpustransform.Untranspose(tokens, corpustransform.TranspositionParams{
		Width: c.Width, Order: c.Order, Round: c.Rounds, Seed: c.Seed,
	})
}

// Metrics are intentionally token-structural. Lexical counts are not part of
// this type because pure transposition leaves them invariant.
type Metrics struct {
	TransitionConcentration float64 `json:"transition_concentration"`
	RelationSignificance    float64 `json:"relation_significance"`
	SequenceRepetition      float64 `json:"sequence_repetition"`
	HigherOrderRepetition   float64 `json:"higher_order_repetition"`
}

func Measure(tokens []string) Metrics {
	if len(tokens) < 2 {
		return Metrics{}
	}
	counts := map[string]map[string]int{}
	for i := 0; i+1 < len(tokens); i++ {
		m := counts[tokens[i]]
		if m == nil {
			m = map[string]int{}
			counts[tokens[i]] = m
		}
		m[tokens[i+1]]++
	}
	var concentration, significance float64
	for _, next := range counts {
		total := 0
		for _, n := range next {
			total += n
		}
		if total == 0 {
			continue
		}
		var sq, entropy float64
		for _, n := range next {
			p := float64(n) / float64(total)
			sq += p * p
			entropy -= p * math.Log2(p)
		}
		concentration += sq
		if len(next) > 1 {
			significance += 1 - entropy/math.Log2(float64(len(next)))
		}
	}
	if len(counts) > 0 {
		concentration /= float64(len(counts))
		significance /= float64(len(counts))
	}
	return Metrics{TransitionConcentration: concentration, RelationSignificance: significance,
		SequenceRepetition: repeatRate(tokens, 2), HigherOrderRepetition: repeatRate(tokens, 3)}
}

func repeatRate(tokens []string, n int) float64 {
	if len(tokens) < n {
		return 0
	}
	counts := map[string]int{}
	for i := 0; i+n <= len(tokens); i++ {
		key := ""
		for _, t := range tokens[i : i+n] {
			key += "\x00" + t
		}
		counts[key]++
	}
	repeated := 0
	for _, n := range counts {
		if n > 1 {
			repeated += n
		}
	}
	return float64(repeated) / float64(len(tokens)-n+1)
}

type ScoredCandidate struct {
	Candidate `json:"candidate"`
	Metrics   Metrics `json:"metrics"`
	Score     float64 `json:"score"`
	Rank      int     `json:"rank"`
}

func Rank(tokens []string, candidates []Candidate) ([]ScoredCandidate, error) {
	result := make([]ScoredCandidate, 0, len(candidates))
	for _, c := range candidates {
		out, err := c.Apply(tokens)
		if err != nil {
			return nil, err
		}
		m := Measure(out)
		result = append(result, ScoredCandidate{Candidate: c, Metrics: m})
	}
	// structural-v2 is family-balanced: each metric is min-max normalized over
	// the pre-registered blind candidate set before taking the arithmetic mean.
	// This uses no oracle and prevents the raw scale of one family from
	// determining the objective.
	minv, maxv := [4]float64{math.Inf(1), math.Inf(1), math.Inf(1), math.Inf(1)}, [4]float64{math.Inf(-1), math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	for _, x := range result {
		v := [4]float64{x.Metrics.TransitionConcentration, x.Metrics.RelationSignificance, x.Metrics.SequenceRepetition, x.Metrics.HigherOrderRepetition}
		for i := range v {
			minv[i] = math.Min(minv[i], v[i])
			maxv[i] = math.Max(maxv[i], v[i])
		}
	}
	for i := range result {
		v := [4]float64{result[i].Metrics.TransitionConcentration, result[i].Metrics.RelationSignificance, result[i].Metrics.SequenceRepetition, result[i].Metrics.HigherOrderRepetition}
		var score float64
		for j := range v {
			if maxv[j] == minv[j] {
				score += 0.5
			} else {
				score += (v[j] - minv[j]) / (maxv[j] - minv[j])
			}
		}
		result[i].Score = score / 4
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Score == result[j].Score {
			return result[i].ID() < result[j].ID()
		}
		return result[i].Score > result[j].Score
	})
	for i := range result {
		result[i].Rank = i + 1
	}
	return result, nil
}
