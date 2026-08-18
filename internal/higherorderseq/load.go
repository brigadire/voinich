package higherorderseq

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

	"zcore.dev/voinich/internal/genericsegmentation"
)

func readTSV(path string) ([]map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(bufio.NewReaderSize(f, 1<<20))
	r.Comma = '\t'
	r.FieldsPerRecord = -1
	head, err := r.Read()
	if err != nil {
		return nil, err
	}
	var out []map[string]string
	for {
		row, e := r.Read()
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, e
		}
		m := make(map[string]string, len(head))
		for i, k := range head {
			if i < len(row) {
				m[k] = row[i]
			}
		}
		out = append(out, m)
	}
	return out, nil
}

func atoi(s string) int     { v, _ := strconv.Atoi(s); return v }
func atof(s string) float64 { v, _ := strconv.ParseFloat(s, 64); return v }

func known(x string) bool {
	x = strings.TrimSpace(x)
	return x != "" && x != "?" && strings.ToLower(x) != "null"
}

// loadCorpusAndBlocks reads the tokenized corpus and its per-token metadata
// map, builds the physical-block segmentation (contiguous joint-metadata
// runs, exactly as every earlier confirmatory stage defines it), and returns
// per-line token counts used by Part K's line-position control.
func loadCorpusAndBlocks(corpusPath, metadataPath string, generic bool) (tokens []Token, blocks []Block, lineLength map[string]int, corpusSHA, metaSHA string, err error) {
	raw, err := os.ReadFile(corpusPath)
	if err != nil {
		return nil, nil, nil, "", "", err
	}
	sum := sha256.Sum256(raw)
	corpusSHA = hex.EncodeToString(sum[:])
	var words []string
	s := bufio.NewScanner(strings.NewReader(string(raw)))
	s.Buffer(make([]byte, 64*1024), 16*1024*1024)
	for s.Scan() {
		words = append(words, strings.Fields(s.Text())...)
	}
	if err = s.Err(); err != nil {
		return nil, nil, nil, "", "", err
	}

	tokens = make([]Token, len(words))
	lineLength = map[string]int{}
	if generic {
		_, lineOfToken, corpusSHA2, e := genericsegmentation.ReadCorpus(corpusPath)
		if e != nil {
			return nil, nil, nil, "", "", e
		}
		infos, e := genericsegmentation.Build(lineOfToken)
		if e != nil {
			return nil, nil, nil, "", "", e
		}
		if len(infos) != len(words) {
			return nil, nil, nil, "", "", fmt.Errorf("generic segmentation/corpus token count mismatch: %d != %d", len(infos), len(words))
		}
		for i, info := range infos {
			t := Token{Position: i, Text: words[i], Line: info.LineID, Currier: info.Group, Hand: genericsegmentation.Sentinel, TokenIndexLine: info.IndexInLine}
			t.Joint = t.Currier + "/" + t.Hand
			tokens[i] = t
			if t.TokenIndexLine+1 > lineLength[t.Line] {
				lineLength[t.Line] = t.TokenIndexLine + 1
			}
		}
		metaSHA = corpusSHA2
	} else {
		metaRaw, e := os.ReadFile(metadataPath)
		if e != nil {
			return nil, nil, nil, "", "", e
		}
		msum := sha256.Sum256(metaRaw)
		metaSHA = hex.EncodeToString(msum[:])
		rows, e := readTSV(metadataPath)
		if e != nil {
			return nil, nil, nil, "", "", e
		}
		if len(rows) != len(words) {
			return nil, nil, nil, "", "", fmt.Errorf("token metadata map has %d tokens but corpus has %d", len(rows), len(words))
		}
		for i, r := range rows {
			if atoi(r["token_position"]) != i || r["token"] != words[i] {
				return nil, nil, nil, "", "", fmt.Errorf("metadata mismatch at token %d", i)
			}
			t := Token{Position: i, Text: r["token"], Line: r["line_id"], Currier: r["currier"], Hand: r["hand"], TokenIndexLine: atoi(r["token_index_in_line"])}
			t.Joint = t.Currier + "/" + t.Hand
			tokens[i] = t
			if t.TokenIndexLine+1 > lineLength[t.Line] {
				lineLength[t.Line] = t.TokenIndexLine + 1
			}
		}
	}

	seen := map[string]int{}
	for i := 0; i < len(tokens); {
		j := i + 1
		for j < len(tokens) && tokens[j].Joint == tokens[i].Joint {
			j++
		}
		if known(tokens[i].Currier) && known(tokens[i].Hand) {
			id := fmt.Sprintf("%s#%d", tokens[i].Joint, seen[tokens[i].Joint])
			seen[tokens[i].Joint]++
			blocks = append(blocks, Block{ID: id, Currier: tokens[i].Currier, Hand: tokens[i].Hand, Joint: tokens[i].Joint, Tokens: append([]Token(nil), tokens[i:j]...)})
		}
		i = j
	}
	return tokens, blocks, lineLength, corpusSHA, metaSHA, nil
}

// loadFrozenCandidates extracts the primary confirmatory and secondary
// descriptive ABC candidates programmatically from the previous audit's own
// outputs (task22 sections 2-3): n>=3 sequences from strict_replicated_sequences.tsv
// joined on sequence with sequence_null_validation.tsv for markov_block_p.
// No new sequence discovery ever happens here.
func loadFrozenCandidates(auditDir string) ([]Candidate, error) {
	strict, err := readTSV(filepath.Join(auditDir, "strict_replicated_sequences.tsv"))
	if err != nil {
		return nil, err
	}
	nullRows, err := readTSV(filepath.Join(auditDir, "sequence_null_validation.tsv"))
	if err != nil {
		return nil, err
	}
	markovP := map[string]float64{}
	for _, r := range nullRows {
		markovP[r["sequence"]] = atof(r["markov_block_p"])
	}
	var out []Candidate
	for _, r := range strict {
		n := atoi(r["n"])
		if n < 3 {
			continue
		}
		q := atof(r["shuffle_block_fdr_q"])
		if q > 0.05 {
			continue
		}
		seq := r["sequence"]
		mp, ok := markovP[seq]
		if !ok {
			return nil, fmt.Errorf("sequence %q missing markov_block_p in sequence_null_validation.tsv", seq)
		}
		family := "secondary"
		if mp <= 0.05 {
			family = "primary"
		}
		out = append(out, Candidate{
			Sequence: seq, Tokens: strings.Fields(seq), Family: family,
			CanonicalOccurrences: atoi(r["canonical_occurrences"]),
			PhysicalBlocks:       atoi(r["physical_blocks"]),
			JointClasses:         atoi(r["joint_classes"]),
			ShuffleFDRQ:          q,
			MarkovBlockP:         mp,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Sequence < out[j].Sequence })
	if len(out) == 0 {
		return nil, fmt.Errorf("no frozen n>=3 candidates with shuffle_block_fdr_q<=0.05 found in %s", auditDir)
	}
	return out, nil
}

// structuralRelatives loads the frozen structural-normalize classes (the
// most permissive, lowest-threshold model in structural_classes.yaml) and
// returns, for every multi-member class, the set of tokens sharing that
// class. Singleton classes (no frozen relative) are simply absent from the
// map. This is a lightweight line-oriented scan rather than a full YAML
// unmarshal because the file is tens of megabytes and only the first model
// section is needed.
func structuralRelatives(discoveryDir string) (map[string][]string, error) {
	path := filepath.Join(discoveryDir, "structural_classes.yaml")
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 64*1024), 4*1024*1024)
	relatives := map[string][]string{}
	seenFirstModel := false
	var current []string
	flush := func() {
		if len(current) > 1 {
			for _, tok := range current {
				var others []string
				for _, o := range current {
					if o != tok {
						others = append(others, o)
					}
				}
				relatives[tok] = others
			}
		}
		current = nil
	}
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		switch {
		case strings.HasPrefix(line, "- threshold:"):
			if seenFirstModel {
				flush()
				return relatives, nil
			}
			seenFirstModel = true
		case strings.HasPrefix(line, "- id:"):
			flush()
		case strings.HasPrefix(line, "- token:"):
			tok := strings.TrimSpace(strings.TrimPrefix(line, "- token:"))
			tok = strings.Trim(tok, `'"`)
			current = append(current, tok)
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	flush()
	return relatives, nil
}
