package pairdecomposition

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type dictionaryEntry struct {
	Token     string `yaml:"token"`
	Count     int    `yaml:"count"`
	Positions []struct {
		Position int `yaml:"position"`
		Count    int `yaml:"count"`
	} `yaml:"position_in_string"`
	Before []struct {
		Token string `yaml:"token"`
		Count int    `yaml:"count"`
	} `yaml:"word_before"`
	After []struct {
		Token string `yaml:"token"`
		Count int    `yaml:"count"`
	} `yaml:"word_after"`
	LineStartCount int `yaml:"line_start_count"`
	LineEndCount   int `yaml:"line_end_count"`
}
type profile struct {
	Count, Start, End int
	Positions         map[int]int
	Left, Right       map[string]int
}
type familiesFile struct {
	Families []FamilyInput `yaml:"families"`
}

func readDictionary(path string) (map[string]profile, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read dictionary: %w", err)
	}
	var in []dictionaryEntry
	if err = yaml.Unmarshal(b, &in); err != nil {
		return nil, fmt.Errorf("decode dictionary: %w", err)
	}
	out := make(map[string]profile, len(in))
	for _, x := range in {
		if x.Token == "" || x.Count < 0 {
			return nil, fmt.Errorf("invalid dictionary token %q", x.Token)
		}
		if _, ok := out[x.Token]; ok {
			return nil, fmt.Errorf("duplicate dictionary token %q", x.Token)
		}
		p := profile{Count: x.Count, Start: x.LineStartCount, End: x.LineEndCount, Positions: map[int]int{}, Left: map[string]int{}, Right: map[string]int{}}
		for _, v := range x.Positions {
			if v.Count < 0 {
				return nil, fmt.Errorf("negative position count for %q", x.Token)
			}
			p.Positions[v.Position] += v.Count
		}
		for _, v := range x.Before {
			if v.Count < 0 {
				return nil, fmt.Errorf("negative predecessor count for %q", x.Token)
			}
			p.Left[v.Token] += v.Count
		}
		for _, v := range x.After {
			if v.Count < 0 {
				return nil, fmt.Errorf("negative successor count for %q", x.Token)
			}
			p.Right[v.Token] += v.Count
		}
		out[x.Token] = p
	}
	return out, nil
}

func readPairs(path string) ([]PairSource, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 65536), 2<<20)
	if !s.Scan() {
		return nil, fmt.Errorf("empty pair file")
	}
	h := strings.Split(s.Text(), "\t")
	col := map[string]int{}
	for i, n := range h {
		col[n] = i
	}
	req := []string{"token_a", "token_b", "count_a", "count_b", "structural_similarity", "reliability", "normalized_grapheme_distance", "position_similarity", "left_similarity", "right_similarity", "position_reliability", "left_reliability", "right_reliability"}
	for _, n := range req {
		if _, ok := col[n]; !ok {
			return nil, fmt.Errorf("missing pair column %q", n)
		}
	}
	var out []PairSource
	line := 1
	for s.Scan() {
		line++
		r := strings.Split(s.Text(), "\t")
		if len(r) != len(h) {
			return nil, fmt.Errorf("line %d: wrong column count", line)
		}
		atoi := func(n string) (int, error) { return strconv.Atoi(r[col[n]]) }
		atof := func(n string) (float64, error) { return strconv.ParseFloat(r[col[n]], 64) }
		p := PairSource{TokenA: r[col["token_a"]], TokenB: r[col["token_b"]]}
		var e error
		if p.CountA, e = atoi("count_a"); e != nil {
			return nil, fmt.Errorf("line %d: %w", line, e)
		}
		if p.CountB, e = atoi("count_b"); e != nil {
			return nil, e
		}
		vals := []*float64{&p.Structural, &p.Reliability, &p.Graphemic, &p.Position, &p.Left, &p.Right, &p.PositionReliability, &p.LeftReliability, &p.RightReliability}
		names := req[4:]
		for i, n := range names {
			if *vals[i], e = atof(n); e != nil {
				return nil, fmt.Errorf("line %d %s: %w", line, n, e)
			}
		}
		out = append(out, p)
	}
	return out, s.Err()
}

func readFamilies(path string) ([]FamilyInput, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	var f familiesFile
	e = yaml.Unmarshal(b, &f)
	return f.Families, e
}

func readDistant(path string) ([][2]string, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	if !s.Scan() {
		return nil, fmt.Errorf("empty distant file")
	}
	h := strings.Split(s.Text(), "\t")
	a, b := -1, -1
	for i, n := range h {
		if n == "token_a" {
			a = i
		}
		if n == "token_b" {
			b = i
		}
	}
	if a < 0 || b < 0 {
		return nil, fmt.Errorf("distant file lacks token columns")
	}
	var out [][2]string
	for s.Scan() {
		r := strings.Split(s.Text(), "\t")
		if a >= len(r) || b >= len(r) {
			return nil, fmt.Errorf("malformed distant row")
		}
		out = append(out, [2]string{r[a], r[b]})
	}
	return out, s.Err()
}
