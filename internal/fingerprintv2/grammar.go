package fingerprintv2

import (
	"fmt"
	"math/rand"
	"sort"
)

type grammarModel struct {
	alphabet    []string
	position    map[string]map[string]int // length, position -> glyph counts
	transitions map[string]map[string]int
	lengths     []int
}

func newGrammarModel(c corpus) grammarModel {
	alphabetSet := map[string]bool{}
	position := map[string]map[string]int{}
	transitions := map[string]map[string]int{}
	lengths := make([]int, len(c.records))
	for i, r := range c.records {
		lengths[i] = len(r.Glyph)
		for p, g := range r.Glyph {
			alphabetSet[g] = true
			key := positionKey(len(r.Glyph), p)
			if position[key] == nil {
				position[key] = map[string]int{}
			}
			position[key][g]++
			if p > 0 {
				if transitions[r.Glyph[p-1]] == nil {
					transitions[r.Glyph[p-1]] = map[string]int{}
				}
				transitions[r.Glyph[p-1]][g]++
			}
		}
	}
	alphabet := orderedKeys(alphabetSet)
	return grammarModel{alphabet: alphabet, position: position, transitions: transitions, lengths: lengths}
}

func positionKey(length, pos int) string { return fmt.Sprintf("%d:%d", length, pos) }

func (m grammarModel) generateOne(length int, rng *rand.Rand) ([]string, error) {
	if length <= 0 {
		return nil, fmt.Errorf("cannot generate zero-length token")
	}
	out := make([]string, length)
	for pos := range out {
		base := m.position[positionKey(length, pos)]
		if len(base) == 0 {
			return nil, fmt.Errorf("grammar has no position profile for length %d position %d", length, pos)
		}
		weights := map[string]int{}
		for _, glyph := range m.alphabet {
			w := base[glyph]
			if pos > 0 {
				// Position-constrained first-order transition weighting joins
				// the positional grammar to the observed local transition
				// profile without replaying any observed token strings.
				w *= m.transitions[out[pos-1]][glyph] + 1
			}
			if w > 0 {
				weights[glyph] = w
			}
		}
		out[pos] = weightedChoice(weights, rng)
	}
	return out, nil
}

func weightedChoice(weights map[string]int, rng *rand.Rand) string {
	keys := orderedKeys(weights)
	total := 0
	for _, k := range keys {
		total += weights[k]
	}
	if total == 0 {
		return ""
	}
	pick := rng.Intn(total)
	for _, k := range keys {
		pick -= weights[k]
		if pick < 0 {
			return k
		}
	}
	return keys[len(keys)-1]
}

func (m grammarModel) generate(source corpus, mode string, rng *rand.Rand) (corpus, error) {
	switch mode {
	case "structure-preserving":
		lengths := append([]int(nil), m.lengths...)
		rng.Shuffle(len(lengths), func(i, j int) { lengths[i], lengths[j] = lengths[j], lengths[i] })
		generated := make([][]string, len(lengths))
		for i, n := range lengths {
			g, err := m.generateOne(n, rng)
			if err != nil {
				return corpus{}, err
			}
			generated[i] = g
		}
		m.ensureAlphabet(generated)
		return generatedCorpus(source, generated), nil
	case "frequency-aware":
		return m.generateFrequencyAware(source, rng)
	default:
		return corpus{}, fmt.Errorf("unsupported grammar mode %q", mode)
	}
}

func (m grammarModel) generateFrequencyAware(source corpus, rng *rand.Rand) (corpus, error) {
	freq := frequencies(source)
	glyphs := glyphByToken(source)
	typesByLength := map[int][]string{}
	for token := range freq {
		typesByLength[len(glyphs[token])] = append(typesByLength[len(glyphs[token])], token)
	}
	lengths := make([]int, 0, len(typesByLength))
	for n := range typesByLength {
		lengths = append(lengths, n)
	}
	sort.Ints(lengths)
	generated := make([][]string, 0, len(source.records))
	for _, length := range lengths {
		originalTypes := typesByLength[length]
		sort.Strings(originalTypes)
		forms, err := m.uniqueForms(length, len(originalTypes), rng)
		if err != nil {
			return corpus{}, err
		}
		counts := make([]int, len(originalTypes))
		for i, token := range originalTypes {
			counts[i] = freq[token]
		}
		sort.Sort(sort.Reverse(sort.IntSlice(counts)))
		// Rank assignment is deliberately independent of generated form:
		// forms are deduplicated first, then a seeded permutation receives
		// each observed rank within its length class.
		assignment := rng.Perm(len(forms))
		for rank, formIndex := range assignment {
			for j := 0; j < counts[rank]; j++ {
				generated = append(generated, append([]string(nil), forms[formIndex]...))
			}
		}
	}
	if len(generated) != len(source.records) {
		return corpus{}, fmt.Errorf("frequency-aware generator emitted %d tokens for %d source tokens", len(generated), len(source.records))
	}
	rng.Shuffle(len(generated), func(i, j int) { generated[i], generated[j] = generated[j], generated[i] })
	m.ensureAlphabet(generated)
	return generatedCorpus(source, generated), nil
}

// ensureAlphabet repairs only an accidentally absent glyph after independent
// sampling. It replaces a duplicate glyph at a position where the missing
// glyph is attested, preserving token lengths and the declared alphabet
// without replaying source token forms.
func (m grammarModel) ensureAlphabet(tokens [][]string) {
	counts := map[string]int{}
	for _, token := range tokens {
		for _, glyph := range token {
			counts[glyph]++
		}
	}
	for _, missing := range m.alphabet {
		if counts[missing] > 0 {
			continue
		}
		repaired := false
		for tokenIndex, token := range tokens {
			for position, donor := range token {
				if m.position[positionKey(len(token), position)][missing] == 0 || counts[donor] <= 1 {
					continue
				}
				tokens[tokenIndex][position] = missing
				counts[donor]--
				counts[missing]++
				repaired = true
				break
			}
			if repaired {
				break
			}
		}
	}
}

func (m grammarModel) uniqueForms(length, want int, rng *rand.Rand) ([][]string, error) {
	forms := make([][]string, 0, want)
	seen := map[string]bool{}
	maxAttempts := max(1000, want*200)
	for tries := 0; len(forms) < want && tries < maxAttempts; tries++ {
		g, err := m.generateOne(length, rng)
		if err != nil {
			return nil, err
		}
		key := glyphKey(g)
		if !seen[key] {
			seen[key] = true
			forms = append(forms, g)
		}
	}
	if len(forms) != want {
		return nil, fmt.Errorf("frequency-aware C-GRAMMAR could generate %d unique length-%d forms after %d attempts (need %d)", len(forms), length, maxAttempts, want)
	}
	return forms, nil
}

func grammarDiagnostic(observed, generated corpus) GrammarDiagnostic {
	obsLengths, genLengths := lengthCounts(observed), lengthCounts(generated)
	obsAlphabet, genAlphabet := alphabetCounts(observed), alphabetCounts(generated)
	obsPosition, genPosition := positionCounts(observed), positionCounts(generated)
	obsInitial, genInitial := endpointCounts(observed, true), endpointCounts(generated, true)
	obsFinal, genFinal := endpointCounts(observed, false), endpointCounts(generated, false)
	obsBigram, genBigram := bigramCounts(observed), bigramCounts(generated)
	obsFreq, genFreq := frequencies(observed), frequencies(generated)
	obsV, genV := len(obsFreq), len(genFreq)
	obsSingleton, genSingleton := frequencyShare(obsFreq, 1), frequencyShare(genFreq, 1)
	obsRare, genRare := frequencyShareAtMost(obsFreq, 2), frequencyShareAtMost(genFreq, 2)
	return GrammarDiagnostic{
		TokenCountExact:         len(observed.records) == len(generated.records),
		LengthDistributionExact: equalCounts(obsLengths, genLengths),
		AlphabetExact:           equalKeySets(obsAlphabet, genAlphabet),
		PositionalGlyphTV:       totalVariation(obsPosition, genPosition),
		InitialGlyphTV:          totalVariation(obsInitial, genInitial),
		FinalGlyphTV:            totalVariation(obsFinal, genFinal),
		BigramTV:                totalVariation(obsBigram, genBigram),
		VocabularySize:          DistributionDiagnostic{Name: "vocabulary_size", Observed: float64(obsV), Generated: float64(genV), Distance: absFloat(float64(obsV - genV))},
		SingletonShare:          DistributionDiagnostic{Name: "singleton_share", Observed: obsSingleton, Generated: genSingleton, Distance: absFloat(obsSingleton - genSingleton)},
		RareShare:               DistributionDiagnostic{Name: "rare_share", Observed: obsRare, Generated: genRare, Distance: absFloat(obsRare - genRare)},
		TokenFrequencyTV:        totalVariation(frequencyHistogram(obsFreq), frequencyHistogram(genFreq)),
	}
}

func lengthCounts(c corpus) map[string]int {
	out := map[string]int{}
	for _, r := range c.records {
		out[fmt.Sprint(len(r.Glyph))]++
	}
	return out
}
func alphabetCounts(c corpus) map[string]int {
	out := map[string]int{}
	for _, r := range c.records {
		for _, g := range r.Glyph {
			out[g]++
		}
	}
	return out
}
func positionCounts(c corpus) map[string]int {
	out := map[string]int{}
	for _, r := range c.records {
		for i, g := range r.Glyph {
			out[positionKey(len(r.Glyph), i)+"\x00"+g]++
		}
	}
	return out
}
func endpointCounts(c corpus, initial bool) map[string]int {
	out := map[string]int{}
	for _, r := range c.records {
		i := len(r.Glyph) - 1
		if initial {
			i = 0
		}
		out[r.Glyph[i]]++
	}
	return out
}
func bigramCounts(c corpus) map[string]int {
	out := map[string]int{}
	for _, r := range c.records {
		for i := 1; i < len(r.Glyph); i++ {
			out[r.Glyph[i-1]+"\x00"+r.Glyph[i]]++
		}
	}
	return out
}
func frequencyHistogram(freq map[string]int) map[string]int {
	out := map[string]int{}
	for _, v := range freq {
		out[fmt.Sprint(v)]++
	}
	return out
}
func frequencyShare(freq map[string]int, n int) float64 {
	if len(freq) == 0 {
		return 0
	}
	count := 0
	for _, v := range freq {
		if v == n {
			count++
		}
	}
	return float64(count) / float64(len(freq))
}
func frequencyShareAtMost(freq map[string]int, n int) float64 {
	if len(freq) == 0 {
		return 0
	}
	count := 0
	for _, v := range freq {
		if v <= n {
			count++
		}
	}
	return float64(count) / float64(len(freq))
}
func equalCounts(a, b map[string]int) bool {
	if len(a) != len(b) {
		return false
	}
	for _, k := range orderedKeys(a) {
		if a[k] != b[k] {
			return false
		}
	}
	return true
}
func equalKeySets(a, b map[string]int) bool {
	if len(a) != len(b) {
		return false
	}
	for _, k := range orderedKeys(a) {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}
func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
