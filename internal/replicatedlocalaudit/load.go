package replicatedlocalaudit

import (
	"bufio"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
		m := map[string]string{}
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

func loadInputs(c Config) ([]token, []block, []distanceCandidate, []sequenceCandidate, string, error) {
	b, err := os.ReadFile(c.CorpusPath)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}
	h := sha256.Sum256(b)
	corpusSHA := hex.EncodeToString(h[:])
	var corpus []string
	s := bufio.NewScanner(strings.NewReader(string(b)))
	s.Buffer(make([]byte, 65536), 1<<20)
	for s.Scan() {
		corpus = append(corpus, strings.Fields(s.Text())...)
	}
	if s.Err() != nil {
		return nil, nil, nil, nil, "", s.Err()
	}
	rows, err := readTSV(c.MetadataPath)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}
	if len(rows) != len(corpus) {
		return nil, nil, nil, nil, "", fmt.Errorf("metadata/corpus token count mismatch: %d != %d", len(rows), len(corpus))
	}
	tokens := make([]token, len(rows))
	known := func(x string) bool {
		x = strings.TrimSpace(x)
		return x != "" && x != "?" && strings.ToLower(x) != "null"
	}
	for i, r := range rows {
		if atoi(r["token_position"]) != i || r["token"] != corpus[i] {
			return nil, nil, nil, nil, "", fmt.Errorf("metadata mismatch at token %d", i)
		}
		tokens[i] = token{Text: r["token"], Line: r["line_id"], Currier: r["currier"], Hand: r["hand"]}
		tokens[i].Joint = tokens[i].Currier + "/" + tokens[i].Hand
	}
	var blocks []block
	seen := map[string]int{}
	for i := 0; i < len(tokens); {
		j := i + 1
		for j < len(tokens) && tokens[j].Joint == tokens[i].Joint {
			j++
		}
		if known(tokens[i].Currier) && known(tokens[i].Hand) {
			id := fmt.Sprintf("%s#%d", tokens[i].Joint, seen[tokens[i].Joint])
			seen[tokens[i].Joint]++
			z := block{ID: id, Currier: tokens[i].Currier, Hand: tokens[i].Hand, Joint: tokens[i].Joint, Tokens: append([]token(nil), tokens[i:j]...)}
			for k := range z.Tokens {
				z.Tokens[k].Block = id
			}
			blocks = append(blocks, z)
		}
		i = j
	}

	sumRows, err := readTSV(filepath.Join(c.RelationDir, "distance_profile_summary.tsv"))
	if err != nil {
		return nil, nil, nil, nil, "", err
	}
	invRows, err := readTSV(filepath.Join(c.RelationDir, "frozen_candidate_inventory.tsv"))
	if err != nil {
		return nil, nil, nil, nil, "", err
	}
	threshold := map[string]float64{}
	for _, r := range invRows {
		threshold[r["candidate_id"]] = atof(r["frozen_threshold"])
	}
	var distances []distanceCandidate
	for _, r := range sumRows {
		if r["family"] != "distance-profile" {
			continue
		}
		distances = append(distances, distanceCandidate{ID: r["candidate_id"], A: r["token_a"], B: r["token_b"], Classification: r["classification"], Eligible: atoi(r["eligible_blocks"]), Joint: atoi(r["joint_classes"]), Currier: atoi(r["currier_classes"]), Hands: atoi(r["hands"]), Mean: atof(r["profile_mean"]), Median: atof(r["profile_median"]), Min: atof(r["profile_min"]), Transfer: atof(r["transfer_success"]), RawP: atof(r["raw_empirical_p"]), Q: atof(r["fdr_q"]), Threshold: threshold[r["candidate_id"]], Fraction: atof(r["fraction_above_frozen_threshold"])})
	}
	classRows, err := readTSV(filepath.Join(c.RelationDir, "relation_classification.tsv"))
	if err != nil {
		return nil, nil, nil, nil, "", err
	}
	var sequences []sequenceCandidate
	recurrenceRows, err := readTSV(filepath.Join(c.RelationDir, "sequence_block_recurrence.tsv"))
	if err != nil {
		return nil, nil, nil, nil, "", err
	}
	recurrence := map[string]bool{}
	for _, r := range recurrenceRows {
		recurrence[r["candidate_id"]] = true
	}
	for _, r := range classRows {
		if r["family"] == "sequence" && r["classification"] == "UNIVERSAL" {
			seq := r["sequence"]
			if !recurrence[r["candidate_id"]] {
				return nil, nil, nil, nil, "", fmt.Errorf("UNIVERSAL sequence %q missing from sequence_block_recurrence.tsv", r["candidate_id"])
			}
			sequences = append(sequences, sequenceCandidate{ID: r["candidate_id"], Sequence: seq, Tokens: strings.Fields(seq), Classification: "UNIVERSAL"})
		}
	}
	if len(sequences) == 0 {
		return nil, nil, nil, nil, "", fmt.Errorf("no UNIVERSAL sequence candidates in frozen classification")
	}
	return tokens, blocks, distances, sequences, corpusSHA, nil
}

func fingerprint(c Config, corpusSHA string) (string, error) {
	h := sha256.New()
	fmt.Fprintf(h, "v1\n%s\n%d\n%d\n", corpusSHA, c.Permutations, c.Seed)
	for _, p := range []string{
		c.MetadataPath,
		filepath.Join(c.RelationDir, "frozen_candidate_inventory.tsv"),
		filepath.Join(c.RelationDir, "distance_profile_block_validation.tsv"),
		filepath.Join(c.RelationDir, "distance_profile_summary.tsv"),
		filepath.Join(c.RelationDir, "sequence_block_recurrence.tsv"),
		filepath.Join(c.RelationDir, "relation_classification.tsv"),
		filepath.Join(c.RelationDir, "relation_controls.tsv"),
		filepath.Join(c.RelationDir, "leave_one_block_out_transfer.tsv"),
		filepath.Join(c.RelationDir, "metadata_transfer_matrix.tsv"),
		filepath.Join(c.RelationDir, "token_relation_validation.yaml"),
		filepath.Join(c.DiscoveryDir, "distance_context_pairs.yaml"),
		filepath.Join(c.DiscoveryDir, "sequence_analysis.yaml"),
	} {
		b, e := os.ReadFile(p)
		if e != nil {
			return "", e
		}
		h.Write(b)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func loadFrozenDistanceDiagnostics(c Config) (map[string]float64, map[string]bool, map[string]bool, map[string]bool, error) {
	similarity := map[string]float64{}
	rows, err := readTSV(filepath.Join(c.RelationDir, "distance_profile_block_validation.tsv"))
	if err != nil {
		return nil, nil, nil, nil, err
	}
	for _, r := range rows {
		if r["family"] == "distance-profile" {
			similarity[r["candidate_id"]+"\x00"+r["block"]] = atof(r["combined_similarity"])
		}
	}
	crossCurrier, crossHand := map[string]bool{}, map[string]bool{}
	rows, err = readTSV(filepath.Join(c.RelationDir, "metadata_transfer_matrix.tsv"))
	if err != nil {
		return nil, nil, nil, nil, err
	}
	for _, r := range rows {
		if r["family"] != "distance-profile" || r["training_metadata"] == r["heldout_metadata"] || atof(r["success_fraction"]) < .67 {
			continue
		}
		if r["dimension"] == "Currier" {
			crossCurrier[r["candidate_id"]] = true
		}
		if r["dimension"] == "hand" {
			crossHand[r["candidate_id"]] = true
		}
	}
	loboSuccess := map[string]bool{}
	rows, err = readTSV(filepath.Join(c.RelationDir, "leave_one_block_out_transfer.tsv"))
	if err != nil {
		return nil, nil, nil, nil, err
	}
	for _, r := range rows {
		if r["family"] == "distance-profile" {
			loboSuccess[r["candidate_id"]+"\x00"+r["heldout_block"]] = r["success"] == "true"
		}
	}
	return similarity, crossCurrier, crossHand, loboSuccess, nil
}
