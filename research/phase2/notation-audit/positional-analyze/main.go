// Command notation-positional-analyze evaluates a fixed, preregistered set
// of positional channels without searching for a channel that looks meaningful.
package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"zcore.dev/voinich/internal/evaglyph"
)

type channel struct {
	ID        string
	Singleton bool
	Lines     [][]string
	Strata    []int
}

type result struct {
	ID           string
	Observations int
	Pairs        int
	Entropy      float64
	NMI          float64
	NullMean     float64
	NullSD       float64
	PValue       float64
	QValue       float64
}

func main() {
	input := flag.String("input", "", "whitespace-tokenized corpus preserving input lines")
	output := flag.String("output", "", "output TSV path")
	seed := flag.Int64("seed", 20260824, "PRNG seed")
	repetitions := flag.Int("repetitions", 1000, "permutation repetitions")
	flag.Parse()
	if *input == "" || *output == "" {
		flag.Usage()
		os.Exit(2)
	}
	if *repetitions <= 0 {
		fatalf("repetitions must be positive")
	}
	data, err := os.ReadFile(*input)
	if err != nil {
		fatalf("read input: %v", err)
	}
	channels := buildChannels(strings.Split(strings.TrimSpace(string(data)), "\n"))
	results := make([]result, 0, len(channels))
	for i, c := range channels {
		results = append(results, analyze(c, *repetitions, rand.New(rand.NewSource(*seed+int64(i)*1000003))))
	}
	adjustBH(results)
	if err := os.MkdirAll(filepath.Dir(*output), 0755); err != nil {
		fatalf("create output directory: %v", err)
	}
	if err := os.WriteFile(*output, render(results, *input, data, *seed, *repetitions), 0644); err != nil {
		fatalf("write output: %v", err)
	}
}

func buildChannels(rawLines []string) []channel {
	channels := []channel{
		{ID: "line_first_glyph", Singleton: true},
		{ID: "line_last_glyph", Singleton: true},
		{ID: "token_first_glyph"},
		{ID: "token_last_glyph"},
		{ID: "token_glyph_position_2"},
		{ID: "token_glyph_position_3"},
		{ID: "token_ordinal_2_first_glyph", Singleton: true},
		{ID: "token_ordinal_3_first_glyph", Singleton: true},
		{ID: "even_token_ordinal_first_glyph"},
		{ID: "line_first_token", Singleton: true},
		{ID: "line_last_token", Singleton: true},
	}
	for _, raw := range rawLines {
		words := strings.Fields(raw)
		tokens := make([][]string, 0, len(words))
		for _, word := range words {
			if glyphs := evaglyph.CollapseEVA(word); len(glyphs) > 0 {
				tokens = append(tokens, glyphs)
			}
		}
		values := make([][]string, len(channels))
		if len(tokens) > 0 {
			values[0] = []string{tokens[0][0]}
			values[1] = []string{tokens[len(tokens)-1][len(tokens[len(tokens)-1])-1]}
			values[9] = []string{strings.Join(tokens[0], "")}
			values[10] = []string{strings.Join(tokens[len(tokens)-1], "")}
		}
		for i, glyphs := range tokens {
			values[2] = append(values[2], glyphs[0])
			values[3] = append(values[3], glyphs[len(glyphs)-1])
			if len(glyphs) >= 2 {
				values[4] = append(values[4], glyphs[1])
			}
			if len(glyphs) >= 3 {
				values[5] = append(values[5], glyphs[2])
			}
			switch i {
			case 1:
				values[6] = append(values[6], glyphs[0])
			case 2:
				values[7] = append(values[7], glyphs[0])
			}
			if i%2 == 1 {
				values[8] = append(values[8], glyphs[0])
			}
		}
		for i := range channels {
			channels[i].Lines = append(channels[i].Lines, values[i])
			channels[i].Strata = append(channels[i].Strata, len(tokens))
		}
	}
	return channels
}

func analyze(c channel, repetitions int, rng *rand.Rand) result {
	observations, pairs, entropy, nmi := statistics(c.Lines, c.Singleton)
	null := make([]float64, repetitions)
	for i := range null {
		nullLines := cloneLines(c.Lines)
		if c.Singleton {
			permuteSingletons(nullLines, c.Strata, rng)
		} else {
			for _, line := range nullLines {
				rng.Shuffle(len(line), func(a, b int) { line[a], line[b] = line[b], line[a] })
			}
		}
		_, _, _, null[i] = statistics(nullLines, c.Singleton)
	}
	mean, sd := meanSD(null)
	exceed := 0
	for _, value := range null {
		if value >= nmi {
			exceed++
		}
	}
	return result{
		ID: c.ID, Observations: observations, Pairs: pairs, Entropy: entropy, NMI: nmi,
		NullMean: mean, NullSD: sd, PValue: float64(exceed+1) / float64(repetitions+1),
	}
}

func statistics(lines [][]string, acrossLines bool) (int, int, float64, float64) {
	counts := map[string]int{}
	pairs := map[string]int{}
	left, right := map[string]int{}, map[string]int{}
	observations, pairCount := 0, 0
	var previous string
	hasPrevious := false
	for _, line := range lines {
		for _, value := range line {
			counts[value]++
			observations++
		}
		if acrossLines && len(line) == 1 {
			if hasPrevious {
				key := previous + "\x00" + line[0]
				pairs[key]++
				left[previous]++
				right[line[0]]++
				pairCount++
			}
			previous, hasPrevious = line[0], true
			continue
		}
		for i := 1; i < len(line); i++ {
			key := line[i-1] + "\x00" + line[i]
			pairs[key]++
			left[line[i-1]]++
			right[line[i]]++
			pairCount++
		}
	}
	if observations == 0 || pairCount == 0 {
		return observations, pairCount, 0, 0
	}
	entropy := entropyOf(counts, observations)
	mi := 0.0
	for key, n := range pairs {
		parts := strings.SplitN(key, "\x00", 2)
		p := float64(n) / float64(pairCount)
		mi += p * math.Log2(float64(n*pairCount)/float64(left[parts[0]]*right[parts[1]]))
	}
	leftEntropy, rightEntropy := entropyOf(left, pairCount), entropyOf(right, pairCount)
	if leftEntropy+rightEntropy == 0 {
		return observations, pairCount, entropy, 0
	}
	return observations, pairCount, entropy, 2 * mi / (leftEntropy + rightEntropy)
}

func entropyOf(counts map[string]int, total int) float64 {
	value := 0.0
	for _, n := range counts {
		p := float64(n) / float64(total)
		value -= p * math.Log2(p)
	}
	return value
}

func permuteSingletons(lines [][]string, strata []int, rng *rand.Rand) {
	buckets := map[int][]int{}
	values := map[int][]string{}
	for i, line := range lines {
		if len(line) == 1 {
			bucket := strata[i]
			buckets[bucket] = append(buckets[bucket], i)
			values[bucket] = append(values[bucket], line[0])
		}
	}
	for bucket, positions := range buckets {
		rng.Shuffle(len(values[bucket]), func(a, b int) {
			values[bucket][a], values[bucket][b] = values[bucket][b], values[bucket][a]
		})
		for i, position := range positions {
			lines[position][0] = values[bucket][i]
		}
	}
}

func cloneLines(lines [][]string) [][]string {
	clone := make([][]string, len(lines))
	for i, line := range lines {
		clone[i] = append([]string(nil), line...)
	}
	return clone
}

func meanSD(values []float64) (float64, float64) {
	total := 0.0
	for _, value := range values {
		total += value
	}
	mean := total / float64(len(values))
	if len(values) < 2 {
		return mean, 0
	}
	sumSquares := 0.0
	for _, value := range values {
		sumSquares += (value - mean) * (value - mean)
	}
	return mean, math.Sqrt(sumSquares / float64(len(values)-1))
}

func adjustBH(results []result) {
	order := make([]int, len(results))
	for i := range results {
		order[i] = i
	}
	sort.Slice(order, func(i, j int) bool { return results[order[i]].PValue < results[order[j]].PValue })
	previous := 1.0
	for i := len(order) - 1; i >= 0; i-- {
		q := results[order[i]].PValue * float64(len(results)) / float64(i+1)
		if q > previous {
			q = previous
		}
		previous = q
		results[order[i]].QValue = q
	}
}

func render(results []result, input string, data []byte, seed int64, repetitions int) []byte {
	hash := sha256.Sum256(data)
	var b strings.Builder
	fmt.Fprintf(&b, "# notation-positional-analyze-v1\tinput=%s\tsha256=%x\tseed=%d\trepetitions=%d\talternative=greater\tfdr=Benjamini-Hochberg\n", input, hash, seed, repetitions)
	b.WriteString("channel_id\tobservations\tadjacent_pairs\tentropy_bits\tadjacent_nmi\tnull_mean\tnull_sd\tempirical_p\tbh_q\n")
	for _, r := range results {
		fmt.Fprintf(&b, "%s\t%d\t%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
			r.ID, r.Observations, r.Pairs, number(r.Entropy), number(r.NMI), number(r.NullMean),
			number(r.NullSD), number(r.PValue), number(r.QValue))
	}
	return []byte(b.String())
}

func number(v float64) string {
	return strconv.FormatFloat(v, 'f', 9, 64)
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
