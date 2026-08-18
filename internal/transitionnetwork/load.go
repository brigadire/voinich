package transitionnetwork

import (
	"bufio"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"zcore.dev/voinich/internal/genericsegmentation"
)

// genericTokensAndBlocks is metadata-validate's generic-mode substitute
// (see GENERIC_STAGE_APPLICABILITY_AUDIT.md): Currier carries the
// deterministic resampling-fold Group label from
// internal/genericsegmentation, and Hand is always
// genericsegmentation.Sentinel, never a fabricated hand identity.
// corpusSHA is passed in already computed, so the corpus is read from disk
// exactly once regardless of mode.
func genericTokensAndBlocks(corpusPath string, words []string, corpusSHA [32]byte) ([]Token, []Block, string, string, error) {
	_, lineOfToken, metaSHA, e := genericsegmentation.ReadCorpus(corpusPath)
	if e != nil {
		return nil, nil, "", "", e
	}
	infos, e := genericsegmentation.Build(lineOfToken)
	if e != nil {
		return nil, nil, "", "", e
	}
	if len(infos) != len(words) {
		return nil, nil, "", "", fmt.Errorf("generic segmentation/corpus token count mismatch: %d != %d", len(infos), len(words))
	}
	toks := make([]Token, len(words))
	for i, info := range infos {
		toks[i] = Token{i, words[i], info.Group, genericsegmentation.Sentinel, info.Group + "/" + genericsegmentation.Sentinel}
	}
	seen := map[string]int{}
	var blocks []Block
	for i := 0; i < len(toks); {
		j := i + 1
		for j < len(toks) && toks[j].Joint == toks[i].Joint {
			j++
		}
		if known(toks[i].Currier) && known(toks[i].Hand) {
			id := fmt.Sprintf("%s#%d", toks[i].Joint, seen[toks[i].Joint])
			seen[toks[i].Joint]++
			blocks = append(blocks, Block{id, toks[i].Currier, toks[i].Hand, toks[i].Joint, append([]Token(nil), toks[i:j]...)})
		}
		i = j
	}
	return toks, blocks, hex.EncodeToString(corpusSHA[:]), metaSHA, nil
}

func known(s string) bool {
	s = strings.TrimSpace(s)
	return s != "" && s != "?" && strings.ToLower(s) != "null"
}
func loadCorpusAndBlocks(corpusPath, metaPath string, generic bool) ([]Token, []Block, string, string, error) {
	raw, e := os.ReadFile(corpusPath)
	if e != nil {
		return nil, nil, "", "", e
	}
	cs := sha256.Sum256(raw)
	var words []string
	s := bufio.NewScanner(strings.NewReader(string(raw)))
	s.Buffer(make([]byte, 65536), 16<<20)
	for s.Scan() {
		words = append(words, strings.Fields(s.Text())...)
	}
	if e = s.Err(); e != nil {
		return nil, nil, "", "", e
	}
	if generic {
		return genericTokensAndBlocks(corpusPath, words, cs)
	}
	mraw, e := os.ReadFile(metaPath)
	if e != nil {
		return nil, nil, "", "", e
	}
	ms := sha256.Sum256(mraw)
	f, e := os.Open(metaPath)
	if e != nil {
		return nil, nil, "", "", e
	}
	defer f.Close()
	r := csv.NewReader(bufio.NewReaderSize(f, 1<<20))
	r.Comma = '\t'
	r.FieldsPerRecord = -1
	head, e := r.Read()
	if e != nil {
		return nil, nil, "", "", e
	}
	idx := map[string]int{}
	for i, h := range head {
		idx[h] = i
	}
	get := func(row []string, k string) string {
		i, ok := idx[k]
		if !ok || i >= len(row) {
			return ""
		}
		return row[i]
	}
	var toks []Token
	for {
		row, e := r.Read()
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, nil, "", "", e
		}
		p, _ := strconv.Atoi(get(row, "token_position"))
		tok := get(row, "token")
		if p != len(toks) || p >= len(words) || tok != words[p] {
			return nil, nil, "", "", fmt.Errorf("metadata mismatch at token %d", len(toks))
		}
		c, h := get(row, "currier"), get(row, "hand")
		toks = append(toks, Token{p, tok, c, h, c + "/" + h})
	}
	if len(toks) != len(words) {
		return nil, nil, "", "", fmt.Errorf("metadata has %d tokens, corpus has %d", len(toks), len(words))
	}
	seen := map[string]int{}
	var blocks []Block
	for i := 0; i < len(toks); {
		j := i + 1
		for j < len(toks) && toks[j].Joint == toks[i].Joint {
			j++
		}
		if known(toks[i].Currier) && known(toks[i].Hand) {
			id := fmt.Sprintf("%s#%d", toks[i].Joint, seen[toks[i].Joint])
			seen[toks[i].Joint]++
			blocks = append(blocks, Block{id, toks[i].Currier, toks[i].Hand, toks[i].Joint, append([]Token(nil), toks[i:j]...)})
		}
		i = j
	}
	return toks, blocks, hex.EncodeToString(cs[:]), hex.EncodeToString(ms[:]), nil
}
