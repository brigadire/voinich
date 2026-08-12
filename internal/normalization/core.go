package normalization

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func LoadStructural(path string) (StructuralInput, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return StructuralInput{}, err
	}
	var input StructuralInput
	if err := yaml.Unmarshal(data, &input); err != nil {
		return StructuralInput{}, err
	}
	return input, nil
}

func LoadCorpus(path string) (Corpus, error) {
	file, err := os.Open(path)
	if err != nil {
		return Corpus{}, err
	}
	defer file.Close()
	corpus := Corpus{Counts: make(map[string]int)}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	for scanner.Scan() {
		tokens := strings.Fields(scanner.Text())
		corpus.Lines = append(corpus.Lines, tokens)
		if len(tokens) == 0 {
			continue
		}
		corpus.NonEmpty++
		corpus.Occurrences += len(tokens)
		corpus.Transitions += len(tokens) - 1
		for _, token := range tokens {
			corpus.Counts[token]++
		}
	}
	if err := scanner.Err(); err != nil {
		return Corpus{}, err
	}
	if corpus.Occurrences-corpus.NonEmpty != corpus.Transitions {
		return Corpus{}, fmt.Errorf("corpus invariant failed")
	}
	return corpus, nil
}

func ParseThresholds(value string) ([]float64, error) {
	seen := make(map[float64]struct{})
	var thresholds []float64
	for _, raw := range strings.Split(value, ",") {
		threshold, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
		if err != nil || threshold < 0 || threshold > 1 {
			return nil, fmt.Errorf("invalid threshold %q", raw)
		}
		if _, exists := seen[threshold]; !exists {
			thresholds = append(thresholds, threshold)
			seen[threshold] = struct{}{}
		}
	}
	if len(thresholds) == 0 {
		return nil, fmt.Errorf("at least one threshold is required")
	}
	sort.Float64s(thresholds)
	return thresholds, nil
}

func ThresholdLabel(threshold float64) string {
	return fmt.Sprintf("%03d", int(math.Round(threshold*100)))
}

func BuildModels(corpus Corpus, structural StructuralInput, config Config) ([]Model, map[string]PairMetrics, error) {
	pairs := make(map[string]PairMetrics)
	for _, candidate := range structural.EquivalenceCandidates {
		if _, ok := corpus.Counts[candidate.TokenA]; !ok {
			return nil, nil, fmt.Errorf("candidate token %q is absent from corpus", candidate.TokenA)
		}
		if _, ok := corpus.Counts[candidate.TokenB]; !ok {
			return nil, nil, fmt.Errorf("candidate token %q is absent from corpus", candidate.TokenB)
		}
		pairs[pairKey(candidate.TokenA, candidate.TokenB)] = PairMetrics{
			Similarity: candidate.Similarity, PositionSimilarity: candidate.PositionSimilarity,
			LeftContextSimilarity: candidate.LeftContextSimilarity, RightContextSimilarity: candidate.RightContextSimilarity,
		}
	}
	models := make([]Model, 0, len(config.Thresholds))
	for _, threshold := range config.Thresholds {
		model, err := buildModel(corpus, pairs, threshold, config)
		if err != nil {
			return nil, nil, err
		}
		models = append(models, model)
	}
	return models, pairs, nil
}

func buildModel(corpus Corpus, pairs map[string]PairMetrics, threshold float64, config Config) (Model, error) {
	eligible := make(map[string]struct{})
	for key, metrics := range pairs {
		if compatible(metrics, threshold, config) {
			left, right := splitPairKey(key)
			if corpus.Counts[left] >= config.MinTokenCount && corpus.Counts[right] >= config.MinTokenCount {
				eligible[left] = struct{}{}
				eligible[right] = struct{}{}
			}
		}
	}
	clusters := make([][]string, 0, len(eligible))
	for token := range eligible {
		clusters = append(clusters, []string{token})
	}
	sortClusters(clusters)

	for {
		bestI, bestJ := -1, -1
		bestScore := -1.0
		for i := 0; i < len(clusters); i++ {
			for j := i + 1; j < len(clusters); j++ {
				score, ok := completeLinkScore(clusters[i], clusters[j], pairs, threshold, config)
				if !ok {
					continue
				}
				merged := mergeMembers(clusters[i], clusters[j])
				bestMerged := []string(nil)
				if bestI >= 0 {
					bestMerged = mergeMembers(clusters[bestI], clusters[bestJ])
				}
				if score > bestScore || (score == bestScore && compareMembers(merged, bestMerged) < 0) {
					bestI, bestJ, bestScore = i, j, score
				}
			}
		}
		if bestI < 0 {
			break
		}
		clusters[bestI] = mergeMembers(clusters[bestI], clusters[bestJ])
		clusters = append(clusters[:bestJ], clusters[bestJ+1:]...)
		sortClusters(clusters)
	}

	used := make(map[string]struct{})
	for _, cluster := range clusters {
		for _, token := range cluster {
			used[token] = struct{}{}
		}
	}
	for token := range corpus.Counts {
		if _, exists := used[token]; !exists {
			clusters = append(clusters, []string{token})
		}
	}
	sortClusters(clusters)

	model := Model{Threshold: threshold, Label: ThresholdLabel(threshold)}
	classifiedOccurrences := 0
	nextID := 1
	for _, members := range clusters {
		id := ""
		for id == "" {
			candidate := fmt.Sprintf("C%04d", nextID)
			nextID++
			if _, collides := corpus.Counts[candidate]; !collides {
				id = candidate
			}
		}
		class := makeClass(id, members, corpus.Counts, pairs)
		model.Classes = append(model.Classes, class)
		if class.Size > model.Stats.LargestClass {
			model.Stats.LargestClass = class.Size
		}
		if class.Size > 1 {
			model.Stats.MultiMemberClasses++
			model.Stats.TokensInMultiClasses += class.Size
			model.Stats.ClassifiedTokens += class.Size
			for _, member := range class.Members {
				classifiedOccurrences += member.Count
			}
		} else {
			model.Stats.SingletonTokens++
		}
	}
	model.Stats.RawUniqueTokens = len(corpus.Counts)
	model.Stats.Classes = len(model.Classes)
	model.Stats.NormalizedUniqueSymbols = len(corpus.Counts) - model.Stats.TokensInMultiClasses + model.Stats.MultiMemberClasses
	model.Stats.TokenOccurrenceCoverage = float64(classifiedOccurrences) / float64(corpus.Occurrences)
	model.Stats.CompressionRatio = float64(model.Stats.NormalizedUniqueSymbols) / float64(model.Stats.RawUniqueTokens)
	if err := ValidateModel(model, corpus.Counts, pairs, config); err != nil {
		return Model{}, err
	}
	return model, nil
}

func compatible(metrics PairMetrics, threshold float64, config Config) bool {
	return metrics.Similarity >= threshold && metrics.PositionSimilarity >= config.MinPositionSimilarity &&
		metrics.LeftContextSimilarity >= config.MinLeftContextSimilarity && metrics.RightContextSimilarity >= config.MinRightContextSimilarity
}

func completeLinkScore(left, right []string, pairs map[string]PairMetrics, threshold float64, config Config) (float64, bool) {
	minimum := math.Inf(1)
	for _, a := range left {
		for _, b := range right {
			metrics, exists := pairs[pairKey(a, b)]
			if !exists || !compatible(metrics, threshold, config) {
				return 0, false
			}
			if metrics.Similarity < minimum {
				minimum = metrics.Similarity
			}
		}
	}
	return minimum, true
}

func makeClass(id string, members []string, counts map[string]int, pairs map[string]PairMetrics) Class {
	class := Class{ID: id, Size: len(members)}
	for _, token := range members {
		class.Members = append(class.Members, Member{Token: token, Count: counts[token]})
	}
	if len(members) == 1 {
		class.MinSimilarity, class.MeanSimilarity = 1, 1
		class.MinPositionSimilarity, class.MinLeftContextSimilarity, class.MinRightContextSimilarity = 1, 1, 1
		return class
	}
	class.MinSimilarity, class.MinPositionSimilarity = 1, 1
	class.MinLeftContextSimilarity, class.MinRightContextSimilarity = 1, 1
	pairCount := 0
	for i := 0; i < len(members); i++ {
		for j := i + 1; j < len(members); j++ {
			metrics := pairs[pairKey(members[i], members[j])]
			class.MeanSimilarity += metrics.Similarity
			pairCount++
			class.MinSimilarity = math.Min(class.MinSimilarity, metrics.Similarity)
			class.MinPositionSimilarity = math.Min(class.MinPositionSimilarity, metrics.PositionSimilarity)
			class.MinLeftContextSimilarity = math.Min(class.MinLeftContextSimilarity, metrics.LeftContextSimilarity)
			class.MinRightContextSimilarity = math.Min(class.MinRightContextSimilarity, metrics.RightContextSimilarity)
		}
	}
	class.MeanSimilarity /= float64(pairCount)
	return class
}

func ValidateModel(model Model, counts map[string]int, pairs map[string]PairMetrics, config Config) error {
	seen := make(map[string]struct{}, len(counts))
	for _, class := range model.Classes {
		for i, member := range class.Members {
			count, exists := counts[member.Token]
			if !exists || count != member.Count {
				return fmt.Errorf("class %s has invalid member %q", class.ID, member.Token)
			}
			if _, duplicate := seen[member.Token]; duplicate {
				return fmt.Errorf("token %q belongs to multiple classes", member.Token)
			}
			seen[member.Token] = struct{}{}
			for j := i + 1; j < len(class.Members); j++ {
				metrics, exists := pairs[pairKey(member.Token, class.Members[j].Token)]
				if !exists || !compatible(metrics, model.Threshold, config) {
					return fmt.Errorf("class %s violates complete-link for %q/%q", class.ID, member.Token, class.Members[j].Token)
				}
			}
		}
	}
	if len(seen) != len(counts) {
		return fmt.Errorf("model loses tokens: got %d, want %d", len(seen), len(counts))
	}
	return nil
}

func Mapping(model Model, singletonMode string) map[string]string {
	mapping := make(map[string]string)
	for _, class := range model.Classes {
		for _, member := range class.Members {
			if class.Size > 1 || singletonMode == "class" {
				mapping[member.Token] = class.ID
			} else {
				mapping[member.Token] = member.Token
			}
		}
	}
	return mapping
}

func WriteNormalized(path string, corpus Corpus, mapping map[string]string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	for _, line := range corpus.Lines {
		normalized := make([]string, len(line))
		for i, token := range line {
			replacement, exists := mapping[token]
			if !exists {
				return fmt.Errorf("token %q is absent from normalization map", token)
			}
			normalized[i] = replacement
		}
		if _, err := writer.WriteString(strings.Join(normalized, " ") + "\n"); err != nil {
			return err
		}
	}
	return writer.Flush()
}

func RandomModel(structural Model, corpus Corpus, minTokenCount int, seed int64, run int) Model {
	rng := rand.New(rand.NewSource(seed + int64(math.Round(structural.Threshold*100))*1_000_000 + int64(run)))
	available := make(map[int][]string)
	for token, count := range corpus.Counts {
		if count < minTokenCount {
			continue
		}
		available[frequencyBin(count)] = append(available[frequencyBin(count)], token)
	}
	binKeys := make([]int, 0, len(available))
	for bin := range available {
		binKeys = append(binKeys, bin)
	}
	sort.Ints(binKeys)
	for _, bin := range binKeys {
		sort.Strings(available[bin])
		rng.Shuffle(len(available[bin]), func(i, j int) { available[bin][i], available[bin][j] = available[bin][j], available[bin][i] })
	}
	var groups [][]string
	for _, class := range structural.Classes {
		if class.Size < 2 {
			continue
		}
		group := make([]string, 0, class.Size)
		for _, member := range class.Members {
			token := takeFromBin(available, frequencyBin(member.Count))
			group = append(group, token)
		}
		sort.Strings(group)
		groups = append(groups, group)
	}
	used := make(map[string]struct{})
	for _, group := range groups {
		for _, token := range group {
			used[token] = struct{}{}
		}
	}
	for token := range corpus.Counts {
		if _, exists := used[token]; !exists {
			groups = append(groups, []string{token})
		}
	}
	sortClusters(groups)
	model := Model{Threshold: structural.Threshold, Label: structural.Label}
	nextID := 1
	for _, group := range groups {
		id := ""
		for id == "" {
			candidate := fmt.Sprintf("C%04d", nextID)
			nextID++
			if _, collides := corpus.Counts[candidate]; !collides {
				id = candidate
			}
		}
		class := Class{ID: id, Size: len(group)}
		for _, token := range group {
			class.Members = append(class.Members, Member{Token: token, Count: corpus.Counts[token]})
		}
		model.Classes = append(model.Classes, class)
	}
	classifiedOccurrences := 0
	for _, class := range model.Classes {
		if class.Size < 2 {
			continue
		}
		model.Stats.MultiMemberClasses++
		model.Stats.TokensInMultiClasses += class.Size
		model.Stats.ClassifiedTokens += class.Size
		if class.Size > model.Stats.LargestClass {
			model.Stats.LargestClass = class.Size
		}
		for _, member := range class.Members {
			classifiedOccurrences += member.Count
		}
	}
	model.Stats.RawUniqueTokens = len(corpus.Counts)
	model.Stats.Classes = len(model.Classes)
	model.Stats.SingletonTokens = len(corpus.Counts) - model.Stats.TokensInMultiClasses
	model.Stats.NormalizedUniqueSymbols = len(corpus.Counts) - model.Stats.TokensInMultiClasses + model.Stats.MultiMemberClasses
	model.Stats.TokenOccurrenceCoverage = float64(classifiedOccurrences) / float64(corpus.Occurrences)
	model.Stats.CompressionRatio = float64(model.Stats.NormalizedUniqueSymbols) / float64(model.Stats.RawUniqueTokens)
	return model
}

func frequencyBin(count int) int {
	if count <= 1 {
		return 0
	}
	return int(math.Floor(math.Log2(float64(count))))
}

func takeFromBin(bins map[int][]string, target int) string {
	if values := bins[target]; len(values) > 0 {
		token := values[len(values)-1]
		bins[target] = values[:len(values)-1]
		return token
	}
	for distance := 1; ; distance++ {
		for _, bin := range []int{target - distance, target + distance} {
			if values := bins[bin]; len(values) > 0 {
				token := values[len(values)-1]
				bins[bin] = values[:len(values)-1]
				return token
			}
		}
	}
}

func pairKey(left, right string) string {
	if left > right {
		left, right = right, left
	}
	return strconv.Itoa(len(left)) + ":" + left + right
}

func splitPairKey(key string) (string, string) {
	separator := strings.IndexByte(key, ':')
	length, _ := strconv.Atoi(key[:separator])
	start := separator + 1
	return key[start : start+length], key[start+length:]
}

func mergeMembers(left, right []string) []string {
	result := append(append([]string(nil), left...), right...)
	sort.Strings(result)
	return result
}

func sortClusters(clusters [][]string) {
	for _, cluster := range clusters {
		sort.Strings(cluster)
	}
	sort.Slice(clusters, func(i, j int) bool { return compareMembers(clusters[i], clusters[j]) < 0 })
}

func compareMembers(left, right []string) int {
	for i := 0; i < len(left) && i < len(right); i++ {
		if left[i] < right[i] {
			return -1
		}
		if left[i] > right[i] {
			return 1
		}
	}
	if len(left) < len(right) {
		return -1
	}
	if len(left) > len(right) {
		return 1
	}
	return 0
}
