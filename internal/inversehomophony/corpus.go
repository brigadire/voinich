package inversehomophony

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"zcore.dev/voinich/internal/genericsegmentation"
)

// RelabelResult is the deterministic opaque relabeling of one corpus
// (task57 section 5/8). ToOpaque/ToOriginal are evaluator-only: the
// recovery engine only ever sees Tokens.
type RelabelResult struct {
	Tokens     []string // relabeled, in original corpus order
	ToOpaque   map[string]string
	ToOriginal map[string]string
}

// Relabel assigns every distinct token in tokens a deterministic
// "x%06d" ID in sorted-token order and substitutes it everywhere. This is
// the mandatory anti-leakage step before any feature is computed, for
// every input (synthetic or Voynich) - task57 section 5 requires it even
// though Task46/55 ciphertext already happens to be x-prefixed, and
// Voynich EVA tokens are not opaque at all without this step.
func Relabel(tokens []string) RelabelResult {
	seen := make(map[string]struct{})
	for _, t := range tokens {
		seen[t] = struct{}{}
	}
	distinct := make([]string, 0, len(seen))
	for t := range seen {
		distinct = append(distinct, t)
	}
	sort.Strings(distinct)

	toOpaque := make(map[string]string, len(distinct))
	toOriginal := make(map[string]string, len(distinct))
	for i, t := range distinct {
		opaque := fmt.Sprintf("x%06d", i+1)
		toOpaque[t] = opaque
		toOriginal[opaque] = t
	}
	out := make([]string, len(tokens))
	for i, t := range tokens {
		out[i] = toOpaque[t]
	}
	return RelabelResult{Tokens: out, ToOpaque: toOpaque, ToOriginal: toOriginal}
}

// LoadedCorpus is a corpus read for recovery: relabeled tokens, the
// original-order line index of every token, and lines grouped for
// structural evaluation.
type LoadedCorpus struct {
	Path        string
	SHA256      string
	Relabel     RelabelResult
	LineOfToken []int
	Lines       [][]string // relabeled tokens grouped by natural line
}

// LoadCorpus reads a plain whitespace-tokenized corpus and relabels it.
// This is the only entry point the recovery engine uses - it never reads
// a mapping/allocation file or a filename that encodes H/model.
func LoadCorpus(path string) (LoadedCorpus, error) {
	tokens, lineOfToken, sha, err := genericsegmentation.ReadCorpus(path)
	if err != nil {
		return LoadedCorpus{}, err
	}
	r := Relabel(tokens)
	lines := groupLines(r.Tokens, lineOfToken)
	return LoadedCorpus{Path: path, SHA256: sha, Relabel: r, LineOfToken: lineOfToken, Lines: lines}, nil
}

func groupLines(tokens []string, lineOfToken []int) [][]string {
	if len(tokens) == 0 {
		return nil
	}
	nLines := lineOfToken[len(lineOfToken)-1] + 1
	lines := make([][]string, nLines)
	for i, t := range tokens {
		lines[lineOfToken[i]] = append(lines[lineOfToken[i]], t)
	}
	return lines
}

// OracleMapping is the evaluator-only Task46/55 plaintext<->cipher mapping,
// loaded from a corpus-transform "<output>.mapping.tsv" file. It must never
// be passed to Recover, Features, or any similarity/clustering code.
type OracleMapping struct {
	// CipherToPlaintext maps an *original* (pre-relabel) cipher token to
	// its plaintext token.
	CipherToPlaintext map[string]string
}

// LoadOracleMapping reads a corpus-transform mapping.tsv
// ("plaintext_token\tcipher_token\tprobability" per line, header first).
func LoadOracleMapping(path string) (OracleMapping, error) {
	f, err := os.Open(path)
	if err != nil {
		return OracleMapping{}, err
	}
	defer f.Close()
	m := OracleMapping{CipherToPlaintext: make(map[string]string)}
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 64*1024), 4*1024*1024)
	first := true
	for s.Scan() {
		line := s.Text()
		if first {
			first = false
			if strings.HasPrefix(line, "plaintext_token\t") {
				continue
			}
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 2 {
			continue
		}
		m.CipherToPlaintext[fields[1]] = fields[0]
	}
	if err := s.Err(); err != nil {
		return OracleMapping{}, err
	}
	return m, nil
}

// OraclePartitionForRelabel builds the evaluator-only oracle partition in
// *relabeled* ID space, so it can be compared against a recovered
// Partition (which is always in relabeled ID space) via
// ClassRecoveryMetrics. relabel must be the same RelabelResult used to
// produce the ciphertext the recovery engine ran on.
func (m OracleMapping) OraclePartitionForRelabel(relabel RelabelResult) Partition {
	p := make(Partition, len(relabel.ToOpaque))
	for original, opaque := range relabel.ToOpaque {
		plaintext, ok := m.CipherToPlaintext[original]
		if !ok {
			// Should not happen for a well-formed Task46/55 corpus: every
			// cipher token in the ciphertext has a mapping entry.
			plaintext = "unmapped:" + original
		}
		p[opaque] = plaintext
	}
	return p
}
