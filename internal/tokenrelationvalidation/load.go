package tokenrelationvalidation

import (
	"bufio"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
	"zcore.dev/voinich/internal/genericsegmentation"
)

func fileSHA(path string) (string, error) {
	f, e := os.Open(path)
	if e != nil {
		return "", e
	}
	defer f.Close()
	h := sha256.New()
	if _, e = io.Copy(h, f); e != nil {
		return "", e
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func loadCorpus(path string) ([]string, string, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, "", e
	}
	h := sha256.Sum256(b)
	var out []string
	s := bufio.NewScanner(strings.NewReader(string(b)))
	s.Buffer(make([]byte, 64*1024), 1024*1024)
	for s.Scan() {
		out = append(out, strings.Fields(s.Text())...)
	}
	return out, hex.EncodeToString(h[:]), s.Err()
}

func loadMetadata(path string, corpus []string) ([]Token, []Block, int, string, error) {
	sha, e := fileSHA(path)
	if e != nil {
		return nil, nil, 0, "", e
	}
	f, e := os.Open(path)
	if e != nil {
		return nil, nil, 0, "", e
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.Comma = '\t'
	r.FieldsPerRecord = -1
	head, e := r.Read()
	if e != nil {
		return nil, nil, 0, "", e
	}
	idx := map[string]int{}
	for i, x := range head {
		idx[x] = i
	}
	required := []string{"token_position", "token", "line_id", "currier", "hand", "token_index_in_line"}
	for _, x := range required {
		if _, ok := idx[x]; !ok {
			return nil, nil, 0, "", fmt.Errorf("metadata map lacks %q", x)
		}
	}
	var tokens []Token
	for {
		row, e := r.Read()
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, nil, 0, "", e
		}
		p, e := strconv.Atoi(row[idx["token_position"]])
		if e != nil {
			return nil, nil, 0, "", e
		}
		li, _ := strconv.Atoi(row[idx["token_index_in_line"]])
		c, h := row[idx["currier"]], row[idx["hand"]]
		tokens = append(tokens, Token{Position: p, Text: row[idx["token"]], Line: row[idx["line_id"]], LineIndex: li, Currier: c, Hand: h, Joint: c + "/" + h})
	}
	if len(tokens) != len(corpus) {
		return nil, nil, 0, "", fmt.Errorf("metadata/corpus token count mismatch: %d != %d", len(tokens), len(corpus))
	}
	for i := range tokens {
		if tokens[i].Position != i || tokens[i].Text != corpus[i] {
			return nil, nil, 0, "", fmt.Errorf("metadata mismatch at token %d: %q != %q", i, tokens[i].Text, corpus[i])
		}
	}
	blocks, unknown := ExtractBlocks(tokens)
	return tokens, blocks, unknown, sha, nil
}

// loadGenericMetadata is metadata-validate's generic-mode substitute: it
// derives the same Token/Block shape genericsegmentation.md documents as
// Class B/C-safe (see GENERIC_STAGE_APPLICABILITY_AUDIT.md), from the
// corpus alone rather than a real IVTFF-derived token_metadata_map.tsv.
// Currier carries the deterministic resampling-fold Group label; Hand is
// always genericsegmentation.Sentinel, never a fabricated hand identity.
func loadGenericMetadata(corpusPath string, corpus []string) ([]Token, []Block, int, string, error) {
	_, lineOfToken, sha, e := genericsegmentation.ReadCorpus(corpusPath)
	if e != nil {
		return nil, nil, 0, "", e
	}
	infos, e := genericsegmentation.Build(lineOfToken)
	if e != nil {
		return nil, nil, 0, "", e
	}
	if len(infos) != len(corpus) {
		return nil, nil, 0, "", fmt.Errorf("generic segmentation/corpus token count mismatch: %d != %d", len(infos), len(corpus))
	}
	tokens := make([]Token, len(corpus))
	for i, info := range infos {
		tokens[i] = Token{
			Position:  i,
			Text:      corpus[i],
			Line:      info.LineID,
			LineIndex: info.IndexInLine,
			Currier:   info.Group,
			Hand:      genericsegmentation.Sentinel,
			Joint:     info.Group + "/" + genericsegmentation.Sentinel,
		}
	}
	blocks, unknown := ExtractBlocks(tokens)
	return tokens, blocks, unknown, sha, nil
}

func knownMetadata(v string) bool {
	v = strings.TrimSpace(v)
	return v != "" && v != "?" && strings.ToLower(v) != "null"
}

// ExtractBlocks assigns a new physical block to every contiguous Currier x hand run.
// Unknown runs receive IDs for auditability but are excluded from the returned primary blocks.
func ExtractBlocks(tokens []Token) ([]Block, int) {
	var all []Block
	unknown := 0
	for i := 0; i < len(tokens); {
		j := i + 1
		for j < len(tokens) && tokens[j].Joint == tokens[i].Joint {
			j++
		}
		id := fmt.Sprintf("%s#%d", tokens[i].Joint, len(all)+1)
		b := Block{ID: id, Currier: tokens[i].Currier, Hand: tokens[i].Hand, Joint: tokens[i].Joint, Start: i, End: j}
		b.Tokens = append([]Token(nil), tokens[i:j]...)
		for k := range b.Tokens {
			b.Tokens[k].BlockID = id
		}
		for k := i; k < j; k++ {
			tokens[k].BlockID = id
		}
		if knownMetadata(b.Currier) && knownMetadata(b.Hand) {
			all = append(all, b)
		} else {
			unknown += j - i
		}
		i = j
	}
	// Renumber within each joint class exactly as residual diagnostics describe it.
	seen := map[string]int{}
	for i := range all {
		all[i].ID = fmt.Sprintf("%s#%d", all[i].Joint, seen[all[i].Joint])
		seen[all[i].Joint]++
		for k := range all[i].Tokens {
			all[i].Tokens[k].BlockID = all[i].ID
		}
	}
	return all, unknown
}

type beginDoc struct {
	Meta struct {
		TokenOccurrences int `yaml:"token_occurrences"`
	} `yaml:"meta"`
	Parameters struct {
		MaxWindow int `yaml:"max_window"`
	} `yaml:"parameters"`
	Candidates []struct {
		A string `yaml:"begin_candidate"`
		B string `yaml:"end_candidate"`
	} `yaml:"candidates"`
}
type distanceDoc struct {
	TokenCount int `yaml:"token_count"`
	Parameters struct {
		MaxDistance int `yaml:"max_distance"`
	} `yaml:"parameters"`
	Pairs []struct {
		A string `yaml:"token_a"`
		B string `yaml:"token_b"`
	} `yaml:"pairs"`
}
type sequenceItem struct {
	Tokens []string `yaml:"tokens"`
	N      int      `yaml:"n"`
}
type sequenceDoc struct {
	Meta struct {
		TokenOccurrences int `yaml:"token_occurrences"`
	} `yaml:"meta"`
	Repeated map[int][]sequenceItem `yaml:"repeated_ngrams"`
}
type reliabilityDoc struct {
	Meta struct {
		TokenOccurrences int `yaml:"token_occurrences"`
	} `yaml:"meta"`
	Parameters struct {
		Threshold float64 `yaml:"threshold"`
	} `yaml:"parameters"`
	Pairs []struct {
		A string  `yaml:"token_a"`
		B string  `yaml:"token_b"`
		S float64 `yaml:"full_similarity"`
	} `yaml:"reference_pairs"`
}
type classesDoc struct {
	Models []struct {
		Threshold float64 `yaml:"threshold"`
		Classes   []struct {
			Members []struct {
				Token string `yaml:"token"`
			} `yaml:"members"`
		} `yaml:"classes"`
	} `yaml:"models"`
}
type softDoc struct {
	Parameters struct {
		GraphMin float64 `yaml:"graph_min_similarity"`
		MinCount int     `yaml:"min_token_count"`
	} `yaml:"parameters"`
}
type tokenCountDoc struct {
	Meta struct {
		TokenOccurrences int `yaml:"token_occurrences"`
	} `yaml:"meta"`
}

func readYAML(path string, v any) error {
	b, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	return yaml.Unmarshal(b, v)
}
func addCandidate(m map[string]*Candidate, c Candidate) {
	if c.A == "" || c.B == "" || c.A == c.B {
		return
	}
	if !c.Directed && c.B < c.A {
		c.A, c.B = c.B, c.A
	}
	key := c.Family + "\x00" + c.A + "\x00" + c.B + "\x00" + c.Sequence
	if old := m[key]; old != nil {
		src := map[string]bool{}
		for _, s := range strings.Split(old.Sources, ",") {
			src[s] = true
		}
		src[c.Sources] = true
		var ss []string
		for s := range src {
			ss = append(ss, s)
		}
		sort.Strings(ss)
		old.Sources = strings.Join(ss, ",")
		if c.FrozenThreshold > old.FrozenThreshold {
			old.FrozenThreshold = c.FrozenThreshold
		}
		return
	}
	c.ID = fmt.Sprintf("%s:%s:%s", c.Family, c.A, c.B)
	if c.Sequence != "" {
		c.ID = c.Family + ":" + strings.ReplaceAll(c.Sequence, " ", "_")
	}
	m[key] = &c
}

func loadCandidates(dir string) ([]Candidate, []InventoryFile, int, error) {
	m := map[string]*Candidate{}
	var files []InventoryFile
	maxD := 20
	inv := func(name string, count int) error {
		p := filepath.Join(dir, name)
		h, e := fileSHA(p)
		if e != nil {
			return e
		}
		files = append(files, InventoryFile{Path: p, SHA256: h, StoredTokenCount: count})
		return nil
	}
	var bd beginDoc
	p := filepath.Join(dir, "begin_end_candidates.yaml")
	if e := readYAML(p, &bd); e != nil {
		return nil, nil, 0, e
	}
	if e := inv("begin_end_candidates.yaml", bd.Meta.TokenOccurrences); e != nil {
		return nil, nil, 0, e
	}
	for _, x := range bd.Candidates {
		addCandidate(m, Candidate{Family: "directional", A: x.A, B: x.B, Directed: true, Sources: "begin_end_candidates.yaml", StoredTokenCount: bd.Meta.TokenOccurrences, FrozenThreshold: float64(bd.Parameters.MaxWindow)})
	}
	var dd distanceDoc
	p = filepath.Join(dir, "distance_context_pairs.yaml")
	if e := readYAML(p, &dd); e != nil {
		return nil, nil, 0, e
	}
	if dd.Parameters.MaxDistance > 0 {
		maxD = dd.Parameters.MaxDistance
	}
	if e := inv("distance_context_pairs.yaml", dd.TokenCount); e != nil {
		return nil, nil, 0, e
	}
	for _, x := range dd.Pairs {
		addCandidate(m, Candidate{Family: "distance-profile", A: x.A, B: x.B, Sources: "distance_context_pairs.yaml", StoredTokenCount: dd.TokenCount, FrozenThreshold: float64(maxD)})
	}
	var sd sequenceDoc
	p = filepath.Join(dir, "sequence_analysis.yaml")
	if e := readYAML(p, &sd); e != nil {
		return nil, nil, 0, e
	}
	if e := inv("sequence_analysis.yaml", sd.Meta.TokenOccurrences); e != nil {
		return nil, nil, 0, e
	}
	for n, items := range sd.Repeated {
		for _, x := range items {
			if len(x.Tokens) != n || n < 2 {
				continue
			}
			seq := strings.Join(x.Tokens, " ")
			addCandidate(m, Candidate{Family: "sequence", A: x.Tokens[0], B: x.Tokens[len(x.Tokens)-1], Sequence: seq, Directed: true, Sources: "sequence_analysis.yaml", StoredTokenCount: sd.Meta.TokenOccurrences})
			if n == 2 {
				addCandidate(m, Candidate{Family: "directional", A: x.Tokens[0], B: x.Tokens[1], Directed: true, Sources: "sequence_analysis.yaml", StoredTokenCount: sd.Meta.TokenOccurrences, FrozenThreshold: 1})
			}
		}
	}
	var rd reliabilityDoc
	p = filepath.Join(dir, "structural_reliability.yaml")
	if e := readYAML(p, &rd); e != nil {
		return nil, nil, 0, e
	}
	if e := inv("structural_reliability.yaml", rd.Meta.TokenOccurrences); e != nil {
		return nil, nil, 0, e
	}
	for _, x := range rd.Pairs {
		addCandidate(m, Candidate{Family: "structural", A: x.A, B: x.B, Sources: "structural_reliability.yaml", StoredTokenCount: rd.Meta.TokenOccurrences, FrozenThreshold: rd.Parameters.Threshold})
	}
	var cd classesDoc
	p = filepath.Join(dir, "structural_classes.yaml")
	if e := readYAML(p, &cd); e != nil {
		return nil, nil, 0, e
	}
	if e := inv("structural_classes.yaml", 0); e != nil {
		return nil, nil, 0, e
	}
	for _, model := range cd.Models {
		if model.Threshold != .7 {
			continue
		}
		for _, cl := range model.Classes {
			for i := 0; i < len(cl.Members); i++ {
				for j := i + 1; j < len(cl.Members); j++ {
					addCandidate(m, Candidate{Family: "structural", A: cl.Members[i].Token, B: cl.Members[j].Token, Sources: "structural_classes.yaml", FrozenThreshold: model.Threshold})
				}
			}
		}
	}
	var soft softDoc
	p = filepath.Join(dir, "soft_structural_space.yaml")
	if e := readYAML(p, &soft); e != nil {
		return nil, nil, 0, e
	}
	if e := inv("soft_structural_space.yaml", 0); e != nil {
		return nil, nil, 0, e
	}
	sp := filepath.Join(dir, "soft_structural_pairs.tsv")
	f, e := os.Open(sp)
	if e != nil {
		return nil, nil, 0, e
	}
	r := csv.NewReader(bufio.NewReaderSize(f, 1<<20))
	r.Comma = '\t'
	head, e := r.Read()
	if e != nil {
		f.Close()
		return nil, nil, 0, e
	}
	ix := map[string]int{}
	for i, x := range head {
		ix[x] = i
	}
	for {
		row, e := r.Read()
		if e == io.EOF {
			break
		}
		if e != nil {
			f.Close()
			return nil, nil, 0, e
		}
		s, _ := strconv.ParseFloat(row[ix["raw_similarity"]], 64)
		if s >= soft.Parameters.GraphMin {
			addCandidate(m, Candidate{Family: "structural", A: row[ix["token_a"]], B: row[ix["token_b"]], Sources: "soft_structural_pairs.tsv", FrozenThreshold: soft.Parameters.GraphMin})
		}
	}
	f.Close()
	h, e := fileSHA(sp)
	if e != nil {
		return nil, nil, 0, e
	}
	files = append(files, InventoryFile{Path: sp, SHA256: h})
	// Tabular rankings and the remaining pre-metadata validation manifests do
	// not expand the candidate universe, but their hashes belong in the frozen
	// input audit trail when they are present.
	for _, extra := range []struct {
		name  string
		count int
	}{
		{"begin_end_top.tsv", bd.Meta.TokenOccurrences}, {"distance_context_top.tsv", dd.TokenCount},
		{"structural_validation.yaml", 0}, {"structural_profile_stability.yaml", 0},
	} {
		path := filepath.Join(dir, extra.name)
		if _, err := os.Stat(path); err == nil {
			count := extra.count
			if strings.HasSuffix(extra.name, ".yaml") {
				var doc tokenCountDoc
				if err := readYAML(path, &doc); err != nil {
					return nil, nil, 0, err
				}
				count = doc.Meta.TokenOccurrences
			}
			hash, err := fileSHA(path)
			if err != nil {
				return nil, nil, 0, err
			}
			files = append(files, InventoryFile{Path: path, SHA256: hash, StoredTokenCount: count})
		}
	}
	out := make([]Candidate, 0, len(m))
	for _, x := range m {
		out = append(out, *x)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Family != out[j].Family {
			return out[i].Family < out[j].Family
		}
		return out[i].ID < out[j].ID
	})
	return out, files, maxD, nil
}
