package propertytrajectory

import (
	"bufio"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
	"strings"
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
		return c, e
	}
	if len(c.Tokens) == 0 {
		return c, fmt.Errorf("corpus is empty")
	}
	return c, nil
}
func readPrevious(path string, limit int) ([]pair, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	var x struct {
		Pairs []struct {
			A string `yaml:"token_a"`
			B string `yaml:"token_b"`
		} `yaml:"pairs"`
	}
	if e = yaml.Unmarshal(b, &x); e != nil {
		return nil, e
	}
	var out []pair
	for _, p := range x.Pairs {
		out = append(out, pair{p.A, p.B})
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}
func readControls(path string) (map[pair][]pair, error) {
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
	for i, v := range h {
		col[v] = i
	}
	req := []string{"target_a", "target_b", "control_rank", "control_a", "control_b"}
	for _, v := range req {
		if _, ok := col[v]; !ok {
			return nil, fmt.Errorf("controls TSV lacks %s", v)
		}
	}
	out := map[pair][]pair{}
	seen := map[string]bool{}
	for s.Scan() {
		r := strings.Split(s.Text(), "\t")
		rank, _ := strconv.Atoi(r[col["control_rank"]])
		k := pair{r[col["target_a"]], r[col["target_b"]]}
		id := k.A + "\x00" + k.B + "\x00" + strconv.Itoa(rank)
		if !seen[id] {
			out[k] = append(out[k], pair{r[col["control_a"]], r[col["control_b"]]})
			seen[id] = true
		}
	}
	return out, s.Err()
}

type structuralEdge struct {
	A, B                    string
	Similarity, Reliability float64
}

func readStructural(path string) ([]structuralEdge, error) {
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
	for i, v := range h {
		col[v] = i
	}
	for _, v := range []string{"token_a", "token_b", "raw_similarity", "evidence_strength"} {
		if _, ok := col[v]; !ok {
			return nil, fmt.Errorf("structural TSV lacks %s", v)
		}
	}
	var out []structuralEdge
	for s.Scan() {
		r := strings.Split(s.Text(), "\t")
		sim, _ := strconv.ParseFloat(r[col["raw_similarity"]], 64)
		rel, _ := strconv.ParseFloat(r[col["evidence_strength"]], 64)
		out = append(out, structuralEdge{r[col["token_a"]], r[col["token_b"]], sim, rel})
	}
	return out, s.Err()
}
