package structuralprojection

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func readCorpus(path string) (corpus, error) {
	f, e := os.Open(path)
	if e != nil {
		return corpus{}, e
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
	if e = s.Err(); e != nil {
		return corpus{}, e
	}
	if len(c.Tokens) == 0 {
		return corpus{}, fmt.Errorf("corpus is empty")
	}
	return c, nil
}

func readEdges(path string) ([]Edge, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 64<<10), 4<<20)
	if !s.Scan() {
		return nil, fmt.Errorf("empty structural pair TSV")
	}
	h := strings.Split(s.Text(), "\t")
	col := map[string]int{}
	for i, x := range h {
		col[x] = i
	}
	req := []string{"token_a", "token_b", "count_a", "count_b", "position_similarity", "left_similarity", "right_similarity", "raw_similarity", "position_reliability", "left_reliability", "right_reliability", "evidence_strength"}
	for _, x := range req {
		if _, ok := col[x]; !ok {
			return nil, fmt.Errorf("structural pair TSV lacks %s", x)
		}
	}
	var out []Edge
	for s.Scan() {
		r := strings.Split(s.Text(), "\t")
		num := func(k string) float64 { v, _ := strconv.ParseFloat(r[col[k]], 64); return v }
		integer := func(k string) int { v, _ := strconv.Atoi(r[col[k]]); return v }
		out = append(out, Edge{A: r[col["token_a"]], B: r[col["token_b"]], CountA: integer("count_a"), CountB: integer("count_b"), Position: num("position_similarity"), Left: num("left_similarity"), Right: num("right_similarity"), Similarity: num("raw_similarity"), PositionReliability: num("position_reliability"), LeftReliability: num("left_reliability"), RightReliability: num("right_reliability"), Reliability: num("evidence_strength")})
	}
	return out, s.Err()
}

type previousMetric struct {
	Distance                     int `yaml:"distance"`
	JS, Overlap, Jaccard         float64
	ObservationsA, ObservationsB int
	Reliability                  float64
}

func (m *previousMetric) UnmarshalYAML(n *yaml.Node) error {
	var x struct {
		Distance      int     `yaml:"distance"`
		JS            float64 `yaml:"js_similarity"`
		Overlap       float64 `yaml:"weighted_overlap"`
		Jaccard       float64 `yaml:"jaccard_support_overlap"`
		ObservationsA int     `yaml:"observations_a"`
		ObservationsB int     `yaml:"observations_b"`
		Reliability   float64 `yaml:"reliability"`
	}
	if e := n.Decode(&x); e != nil {
		return e
	}
	*m = previousMetric{x.Distance, x.JS, x.Overlap, x.Jaccard, x.ObservationsA, x.ObservationsB, x.Reliability}
	return nil
}

type previousPair struct {
	TokenA string           `yaml:"token_a"`
	TokenB string           `yaml:"token_b"`
	Right  []previousMetric `yaml:"right_context"`
	Left   []previousMetric `yaml:"left_context"`
}
type previousFile struct {
	Pairs []previousPair `yaml:"pairs"`
}

func readPrevious(path string, limit int) ([]previousPair, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	var x previousFile
	if e = yaml.Unmarshal(b, &x); e != nil {
		return nil, e
	}
	if limit > 0 && len(x.Pairs) > limit {
		x.Pairs = x.Pairs[:limit]
	}
	return x.Pairs, nil
}

type familyFile struct {
	Families []family `yaml:"families"`
}

func readFamilies(path string) ([]family, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	var x familyFile
	e = yaml.Unmarshal(b, &x)
	return x.Families, e
}

func uniqueTokens(c corpus) []string {
	out := make([]string, 0, len(c.Counts))
	for t := range c.Counts {
		out = append(out, t)
	}
	sortStrings(out)
	return out
}
func sortStrings(x []string) { // local insertion-free wrapper keeps call sites terse
	for i := 1; i < len(x); i++ {
		for j := i; j > 0 && x[j] < x[j-1]; j-- {
			x[j], x[j-1] = x[j-1], x[j]
		}
	}
}
