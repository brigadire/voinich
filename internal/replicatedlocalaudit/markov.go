package replicatedlocalaudit

import (
	"math/rand"
	"sort"
)

type transitionModel struct {
	vocab   []string
	unigram []int
	next    map[string][]string
	weights map[string][]int
}

func buildMarkov(training []block) *transitionModel {
	u := map[string]int{}
	trans := map[string]map[string]int{}
	for _, b := range training {
		for _, line := range splitBlockLines(b) {
			for i, t := range line {
				u[t.Text]++
				if i+1 < len(line) {
					if trans[t.Text] == nil {
						trans[t.Text] = map[string]int{}
					}
					trans[t.Text][line[i+1].Text]++
				}
			}
		}
	}
	if len(u) == 0 {
		return nil
	}
	m := &transitionModel{next: map[string][]string{}, weights: map[string][]int{}}
	for x := range u {
		m.vocab = append(m.vocab, x)
	}
	sort.Strings(m.vocab)
	for _, x := range m.vocab {
		m.unigram = append(m.unigram, u[x])
	}
	for from, z := range trans {
		var xs []string
		for x := range z {
			xs = append(xs, x)
		}
		sort.Strings(xs)
		for _, x := range xs {
			m.next[from] = append(m.next[from], x)
			m.weights[from] = append(m.weights[from], z[x])
		}
	}
	return m
}
func weightedChoice(r *rand.Rand, x []string, w []int) string {
	total := 0
	for _, n := range w {
		total += n
	}
	if total == 0 {
		return ""
	}
	k := r.Intn(total)
	for i, n := range w {
		if k < n {
			return x[i]
		}
		k -= n
	}
	return x[len(x)-1]
}

// markovHeldOut pairs a held-out block with the leave-one-block-out
// training model built from every other block sharing its Joint metadata.
type markovHeldOut struct {
	held  block
	model *transitionModel
}

// buildMarkovTraining precomputes the leave-one-block-out Markov training
// model for every held-out block that has a non-empty training partition.
// The training partition for a given held-out block (every other block with
// the same Joint) and the model built from it do not depend on any
// permutation replicate's seed, only on the fixed block set — so this needs
// to run once per RunAndWrite call rather than once per replicate. Held-out
// blocks with no training data (m == nil) are omitted, exactly as the
// former inline per-replicate loop skipped them via `continue`.
func buildMarkovTraining(blocks []block) []markovHeldOut {
	out := make([]markovHeldOut, 0, len(blocks))
	for _, held := range blocks {
		var train []block
		for _, b := range blocks {
			if b.ID != held.ID && b.Joint == held.Joint {
				train = append(train, b)
			}
		}
		m := buildMarkov(train)
		if m == nil {
			continue
		}
		out = append(out, markovHeldOut{held: held, model: m})
	}
	return out
}

// markovBlocks draws one leakage-free first-order Markov replicate from the
// precomputed training models, in the same held-out-block order used to
// build them, so the sequence of rand draws (and thus the output) is
// identical to the former implementation that rebuilt each model inline.
func markovBlocks(training []markovHeldOut, seed int64) ([]block, int) {
	r := rand.New(rand.NewSource(seed))
	out := make([]block, 0, len(training))
	for _, ho := range training {
		held, m := ho.held, ho.model
		z := held
		z.Tokens = append([]token(nil), held.Tokens...)
		for i := 0; i < len(z.Tokens); {
			j := i + 1
			for j < len(z.Tokens) && z.Tokens[j].Line == z.Tokens[i].Line {
				j++
			}
			prev := weightedChoice(r, m.vocab, m.unigram)
			for k := i; k < j; k++ {
				if k > i {
					n := weightedChoice(r, m.next[prev], m.weights[prev])
					if n == "" {
						n = weightedChoice(r, m.vocab, m.unigram)
					}
					prev = n
				}
				z.Tokens[k].Text = prev
			}
			i = j
		}
		out = append(out, z)
	}
	return out, len(training)
}
