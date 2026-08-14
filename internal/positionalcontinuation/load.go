package positionalcontinuation

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
// per-line token counts.
func loadCorpusAndBlocks(corpusPath, metadataPath string) (tokens []Token, blocks []Block, lineLength map[string]int, corpusSHA, metaSHA string, err error) {
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

	metaRaw, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, nil, nil, "", "", err
	}
	msum := sha256.Sum256(metaRaw)
	metaSHA = hex.EncodeToString(msum[:])
	rows, err := readTSV(metadataPath)
	if err != nil {
		return nil, nil, nil, "", "", err
	}
	if len(rows) != len(words) {
		return nil, nil, nil, "", "", fmt.Errorf("token metadata map has %d tokens but corpus has %d", len(rows), len(words))
	}
	tokens = make([]Token, len(rows))
	lineLength = map[string]int{}
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

// higherOrderInputFiles are the frozen higher-order-sequence-validate outputs
// this program reads as read-only context (never recomputed).
var higherOrderInputFiles = []string{
	"higher_order_candidate_inventory.tsv",
	"higher_order_occurrences.tsv",
	"conditional_probability_by_block.tsv",
	"conditional_dependence.tsv",
	"conditional_cmi.tsv",
	"continuation_distributions.tsv",
	"continuation_entropy.tsv",
	"first_vs_second_order_lobo.tsv",
	"conditional_context_controls.tsv",
	"conditional_context_rank.tsv",
	"higher_order_cross_block.tsv",
	"higher_order_jackknife.tsv",
	"higher_order_position.tsv",
	"higher_order_validation.tsv",
	"higher_order_sequence_analysis.yaml",
}

// higherOrderFingerprint hashes every frozen input file this program reads
// from the previous stage, so both the checkpoint fingerprint and the
// reproducibility record change the moment any upstream frozen input changes.
func higherOrderFingerprint(dir string) (string, error) {
	h := sha256.New()
	for _, name := range higherOrderInputFiles {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return "", err
		}
		h.Write(b)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// priorResult reads the previous experiment's approximate s-aiin-chey
// statistics from higher_order_sequence_analysis.yaml's frozen candidate
// list, purely for documentation in the report (never as a computation
// input - task23 section 103 requires a fresh recount).
func priorResultExists(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "higher_order_sequence_analysis.yaml"))
	return err == nil
}

func stringKeysInt(m map[string]int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func stringKeysFloat(m map[string]float64) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
