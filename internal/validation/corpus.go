package validation

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"

	"zcore.dev/voinich/internal/normalization"
)

func LoadCorpus(path string) (Corpus, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Corpus{}, "", err
	}
	result := Corpus{Counts: make(map[string]int)}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	lineID := 0
	for scanner.Scan() {
		lineID++
		tokens := strings.Fields(scanner.Text())
		result.Lines = append(result.Lines, Line{ID: lineID, Tokens: tokens})
		if len(tokens) == 0 {
			continue
		}
		result.Occurrences += len(tokens)
		result.Transitions += len(tokens) - 1
		for _, token := range tokens {
			result.Counts[token]++
		}
	}
	if err := scanner.Err(); err != nil {
		return Corpus{}, "", err
	}
	if result.Occurrences-nonEmptyLines(result.Lines) != result.Transitions {
		return Corpus{}, "", fmt.Errorf("corpus invariant failed")
	}
	sum := sha256.Sum256(data)
	return result, hex.EncodeToString(sum[:]), nil
}

func SplitFolds(lines []Line, folds int, seed int64) ([][]int, error) {
	if folds < 2 {
		return nil, fmt.Errorf("folds must be at least 2")
	}
	if len(lines) < folds {
		return nil, fmt.Errorf("folds cannot exceed physical line count")
	}
	rng := rand.New(rand.NewSource(seed))
	permutation := rng.Perm(len(lines))
	result := make([][]int, folds)
	for shuffledPosition, lineIndex := range permutation {
		result[shuffledPosition%folds] = append(result[shuffledPosition%folds], lineIndex)
	}
	for i := range result {
		sort.Ints(result[i])
	}
	return result, nil
}

func Partition(corpus Corpus, testIndexes []int) (Corpus, Corpus, error) {
	testSet := make(map[int]struct{}, len(testIndexes))
	for _, index := range testIndexes {
		if index < 0 || index >= len(corpus.Lines) {
			return Corpus{}, Corpus{}, fmt.Errorf("line index %d is out of range", index)
		}
		if _, duplicate := testSet[index]; duplicate {
			return Corpus{}, Corpus{}, fmt.Errorf("duplicate line index %d", index)
		}
		testSet[index] = struct{}{}
	}
	train := Corpus{Counts: make(map[string]int)}
	test := Corpus{Counts: make(map[string]int)}
	for index, line := range corpus.Lines {
		target := &train
		if _, isTest := testSet[index]; isTest {
			target = &test
		}
		addLine(target, line)
	}
	return train, test, nil
}

func addLine(corpus *Corpus, line Line) {
	corpus.Lines = append(corpus.Lines, line)
	if len(line.Tokens) == 0 {
		return
	}
	corpus.Occurrences += len(line.Tokens)
	corpus.Transitions += len(line.Tokens) - 1
	for _, token := range line.Tokens {
		corpus.Counts[token]++
	}
}

func partitionStats(corpus Corpus, includeIDs bool) PartitionStats {
	ids := make([]int, len(corpus.Lines))
	for i, line := range corpus.Lines {
		ids[i] = line.ID
	}
	result := PartitionStats{
		PhysicalLines: len(corpus.Lines), NonEmptyLines: nonEmptyLines(corpus.Lines),
		TokenOccurrences: corpus.Occurrences, Transitions: corpus.Transitions,
		LineIDsSHA256: hashLineIDs(ids),
	}
	if includeIDs {
		result.LineIDs = ids
	}
	return result
}

func hashLineIDs(ids []int) string {
	hash := sha256.New()
	for _, id := range ids {
		hash.Write([]byte(strconv.Itoa(id)))
		hash.Write([]byte{'\n'})
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func nonEmptyLines(lines []Line) int {
	count := 0
	for _, line := range lines {
		if len(line.Tokens) > 0 {
			count++
		}
	}
	return count
}

func normalizationCorpus(corpus Corpus) normalization.Corpus {
	result := normalization.Corpus{
		Counts: corpus.Counts, Occurrences: corpus.Occurrences,
		NonEmpty: nonEmptyLines(corpus.Lines), Transitions: corpus.Transitions,
	}
	for _, line := range corpus.Lines {
		result.Lines = append(result.Lines, line.Tokens)
	}
	return result
}

func applyMapping(corpus Corpus, mapping map[string]string) Corpus {
	result := Corpus{Counts: make(map[string]int), Occurrences: corpus.Occurrences, Transitions: corpus.Transitions}
	for _, line := range corpus.Lines {
		tokens := make([]string, len(line.Tokens))
		for i, token := range line.Tokens {
			tokens[i] = token
			if replacement, exists := mapping[token]; exists {
				tokens[i] = replacement
			}
			result.Counts[tokens[i]]++
		}
		result.Lines = append(result.Lines, Line{ID: line.ID, Tokens: tokens})
	}
	return result
}
