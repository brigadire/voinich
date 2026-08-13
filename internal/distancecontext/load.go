package distancecontext

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type corpus struct {
	Lines  [][]string
	Tokens []string
	Counts map[string]int
}
type pair struct{ A, B string }
type familyInput struct {
	ID     int      `yaml:"id"`
	Tokens []string `yaml:"tokens"`
	Edges  []struct {
		A string `yaml:"token_a"`
		B string `yaml:"token_b"`
	} `yaml:"edges"`
}
type familyFile struct {
	Families []familyInput `yaml:"families"`
}
type controlInput struct {
	TargetA, TargetB, A, B string
	Rank                   int
}

func readCorpus(path string) (corpus, error) {
	f, err := os.Open(path)
	if err != nil {
		return corpus{}, fmt.Errorf("read corpus: %w", err)
	}
	defer f.Close()
	c := corpus{Counts: map[string]int{}}
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 64<<10), 4<<20)
	for s.Scan() {
		x := strings.Fields(s.Text())
		if len(x) == 0 {
			continue
		}
		c.Lines = append(c.Lines, x)
		c.Tokens = append(c.Tokens, x...)
		for _, t := range x {
			c.Counts[t]++
		}
	}
	if err = s.Err(); err != nil {
		return corpus{}, err
	}
	if len(c.Tokens) == 0 {
		return corpus{}, fmt.Errorf("corpus is empty")
	}
	return c, nil
}

func readPairTSV(path string, limit int) ([]pair, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	if !s.Scan() {
		return nil, fmt.Errorf("empty pair TSV")
	}
	h := strings.Split(s.Text(), "\t")
	ai, bi := -1, -1
	for i, x := range h {
		if x == "token_a" {
			ai = i
		}
		if x == "token_b" {
			bi = i
		}
	}
	if ai < 0 || bi < 0 {
		return nil, fmt.Errorf("pair TSV lacks token columns")
	}
	var out []pair
	for s.Scan() {
		r := strings.Split(s.Text(), "\t")
		if ai >= len(r) || bi >= len(r) {
			return nil, fmt.Errorf("malformed pair TSV")
		}
		out = append(out, pair{r[ai], r[bi]})
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, s.Err()
}

func readFamilies(path string) ([]familyInput, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	var x familyFile
	e = yaml.Unmarshal(b, &x)
	return x.Families, e
}

func readControls(path string) ([]controlInput, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	if !s.Scan() {
		return nil, fmt.Errorf("empty controls TSV")
	}
	h := strings.Split(s.Text(), "\t")
	col := map[string]int{}
	for i, x := range h {
		col[x] = i
	}
	for _, x := range []string{"target_a", "target_b", "control_rank", "control_a", "control_b"} {
		if _, ok := col[x]; !ok {
			return nil, fmt.Errorf("controls TSV lacks %s", x)
		}
	}
	var out []controlInput
	for s.Scan() {
		r := strings.Split(s.Text(), "\t")
		rank, e := strconv.Atoi(r[col["control_rank"]])
		if e != nil {
			return nil, e
		}
		out = append(out, controlInput{r[col["target_a"]], r[col["target_b"]], r[col["control_a"]], r[col["control_b"]], rank})
	}
	return out, s.Err()
}

func canon(a, b string) pair {
	if b < a {
		a, b = b, a
	}
	return pair{a, b}
}
