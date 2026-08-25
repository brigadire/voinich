// task85-analyze builds Task85's deterministic hierarchical corpus split
// (GRAMMAR_CORPUS_SPLIT.tsv / GRAMMAR_CORPUS_SPLIT_MANIFEST.json) and
// authoritative corpus-scale statistics from the Task83b-frozen ZL3b/IT2a
// corpora (task85 sections 10, 12-16). It is a one-shot design-freeze audit
// tool, like research/phase2/task83b-analyze and task83r-analyze: it does
// not touch Stages 1-28, builds no grammar, and does not read any HELDOUT
// content for a selection decision (task85 section 41/53).
//
// The split unit is the physical leaf ("folio" in task85's own vocabulary):
// recto/verso sides of the same leaf (and multi-part foldout sides such as
// f102r1/f102r2) are always assigned together, so no partition boundary can
// separate two sides of the same physical page (task85 section 14). Leaves
// are stratified by (Currier language $L, section $I) and, within each
// stratum, sorted by leaf number and assigned by a fixed positional pattern
// (index%5 in {0,1,2}->DEVELOPMENT, 3->VALIDATION, 4->HELDOUT) - a rule, not
// a pseudo-random draw, so the split needs no seed and cannot be re-run into
// a different answer (task85 section 50 determinism inheritance).
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"zcore.dev/voinich/internal/evaglyph"
	"zcore.dev/voinich/internal/genericsegmentation"
	"zcore.dev/voinich/internal/metadatavalidation"
)

func sha256Sum(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

const outDir = "research/phase3/task85"

type source struct {
	name      string
	canonical string
	ivtff     string
}

var sources = []source{
	{"ZL3b", "data_work/ZL3b-x7.canonical.txt", "data/ZL3b-n.txt"},
	{"IT2a", "data_work/IT2a-x7.canonical.txt", "data/IT2a-n.txt"},
}

var sectionNames = map[string]string{
	"A": "Astronomical", "B": "Biological", "C": "Cosmological", "H": "Herbal",
	"P": "Pharmaceutical", "S": "Stars", "T": "Text", "Z": "Zodiac",
}

func sectionLabel(code string) string {
	if code == "" {
		return "UNKNOWN"
	}
	if name, ok := sectionNames[code]; ok {
		return name
	}
	return code
}

func currierLabel(code string) string {
	if code == "" {
		return "UNKNOWN"
	}
	return code
}

var leafRe = regexp.MustCompile(`^(f[0-9]+)`)

func leafOf(pageID string) string {
	m := leafRe.FindStringSubmatch(pageID)
	if m == nil {
		return pageID
	}
	return m[1]
}

func leafNumber(leaf string) int {
	n, _ := strconv.Atoi(strings.TrimPrefix(leaf, "f"))
	return n
}

// tokenRec is one aligned corpus token joined to its IVTFF locus metadata,
// via the frozen token-level metadatavalidation.Align machinery already
// used by internal/fingerprintv2 (task77 cross-scale block) - not a naive
// canonical-line-index-equals-locus-index assumption, which does not hold
// for IT2a (5215 loci fold into 5207 physical lines because some loci are
// same-line continuation fragments).
type tokenRec struct {
	glyphs  []string
	folio   string
	currier string
	section string
	lineKey string // Folio + "\x00" + LineID, one physical line
}

type corpusData struct {
	name            string
	canonicalPath   string
	ivtffPath       string
	canonicalSHA256 string
	ivtffSHA256     string
	tokens          []tokenRec
}

func loadCorpus(s source) (corpusData, error) {
	rawTokens, _, sha, err := genericsegmentation.ReadCorpus(s.canonical)
	if err != nil {
		return corpusData{}, err
	}
	doc, err := metadatavalidation.ParseIVTFF(s.ivtff)
	if err != nil {
		return corpusData{}, err
	}
	aligned, err := metadatavalidation.Align(doc, rawTokens, sha)
	if err != nil {
		return corpusData{}, fmt.Errorf("%s: strict IVTFF alignment: %w", s.name, err)
	}
	if len(aligned.Records) != len(rawTokens) {
		return corpusData{}, fmt.Errorf("%s: aligned %d records for %d tokens", s.name, len(aligned.Records), len(rawTokens))
	}
	ivtffSHA, _, err := sha256File(s.ivtff)
	if err != nil {
		return corpusData{}, err
	}
	toks := make([]tokenRec, len(rawTokens))
	for i, raw := range rawTokens {
		m := aligned.Records[i]
		toks[i] = tokenRec{
			glyphs:  evaglyph.CollapseEVA(raw),
			folio:   m.Folio,
			currier: currierLabel(m.Variables["L"]),
			section: sectionLabel(m.Variables["I"]),
			lineKey: m.Folio + "\x00" + m.LineID,
		}
	}
	return corpusData{
		name: s.name, canonicalPath: s.canonical, ivtffPath: s.ivtff,
		canonicalSHA256: sha, ivtffSHA256: ivtffSHA, tokens: toks,
	}, nil
}

func sha256File(path string) (string, []byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", nil, err
	}
	h := sha256Sum(b)
	return h, b, nil
}

// leafStats accumulates per-leaf (physical-page-pair) statistics within one
// transcription.
type leafStats struct {
	pages         map[string]bool
	tokenCount    int
	lineCount     int
	sectionCounts map[string]int
	currierCounts map[string]int
	lengthSum     int64
	lengthSumSq   int64
	lengthHist    map[int]int
}

func newLeafStats() *leafStats {
	return &leafStats{pages: map[string]bool{}, sectionCounts: map[string]int{}, currierCounts: map[string]int{}, lengthHist: map[int]int{}}
}

func buildLeafStats(c corpusData) map[string]*leafStats {
	out := map[string]*leafStats{}
	seenLine := map[string]map[string]bool{}
	for _, t := range c.tokens {
		leaf := leafOf(t.folio)
		st, ok := out[leaf]
		if !ok {
			st = newLeafStats()
			out[leaf] = st
			seenLine[leaf] = map[string]bool{}
		}
		st.pages[t.folio] = true
		if !seenLine[leaf][t.lineKey] {
			seenLine[leaf][t.lineKey] = true
			st.lineCount++
		}
		st.sectionCounts[t.section]++
		st.currierCounts[t.currier]++
		st.tokenCount++
		n := len(t.glyphs)
		st.lengthSum += int64(n)
		st.lengthSumSq += int64(n) * int64(n)
		st.lengthHist[n]++
	}
	return out
}

// dominant returns the key with the largest count, breaking ties by sorted
// key order so the result does not depend on map iteration order (project
// determinism convention: sort map keys before accumulation/selection).
func dominant(m map[string]int) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	best, bestN := "", -1
	for _, k := range keys {
		if m[k] > bestN {
			best, bestN = k, m[k]
		}
	}
	return best
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

type leafAssignment struct {
	leaf     string
	currier  string
	section  string
	stratum  string
	partition string
}

const (
	development = "DEVELOPMENT"
	validation  = "VALIDATION"
	heldout     = "HELDOUT"
)

// partitionOf implements the fixed, seed-free assignment rule (task85
// section 50): index 0,1,2 mod 5 -> DEVELOPMENT (60%), 3 -> VALIDATION
// (20%), 4 -> HELDOUT (20%), applied per stratum after leaves are sorted by
// numeric leaf order.
func partitionOf(indexInStratum int) string {
	switch indexInStratum % 5 {
	case 3:
		return validation
	case 4:
		return heldout
	default:
		return development
	}
}

type corpusStats struct {
	CanonicalPath   string `json:"canonical_path"`
	IVTFFPath       string `json:"ivtff_path"`
	CanonicalSHA256 string `json:"canonical_sha256"`
	IVTFFSHA256     string `json:"ivtff_sha256"`
	LineCount       int    `json:"line_count"`
	TokenCount      int    `json:"token_count"`
	VocabSize       int    `json:"vocab_size"`
	GlyphInventory  int    `json:"glyph_inventory_size"`
	LeafCount       int    `json:"leaf_count"`
}

type partitionStats struct {
	LeafCount        int             `json:"leaf_count"`
	TokenCount       int             `json:"token_count"`
	LineCount        int             `json:"line_count"`
	MeanTokenLength  float64         `json:"mean_token_length"`
	SDTokenLength    float64         `json:"sd_token_length"`
	SectionComposition map[string]int `json:"section_composition"`
	CurrierComposition map[string]int `json:"currier_composition"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	corpora := map[string]corpusData{}
	for _, s := range sources {
		c, err := loadCorpus(s)
		if err != nil {
			return err
		}
		corpora[s.name] = c
	}

	leafStatsBySource := map[string]map[string]*leafStats{}
	for name, c := range corpora {
		leafStatsBySource[name] = buildLeafStats(c)
	}

	// Union of leaf IDs, ZL3b preferred for authoritative currier/section
	// when a leaf exists in both (ZL3b has the larger folio coverage).
	leafSet := map[string]bool{}
	for _, m := range leafStatsBySource {
		for leaf := range m {
			leafSet[leaf] = true
		}
	}
	allLeaves := make([]string, 0, len(leafSet))
	for l := range leafSet {
		allLeaves = append(allLeaves, l)
	}
	sort.Slice(allLeaves, func(i, j int) bool { return leafNumber(allLeaves[i]) < leafNumber(allLeaves[j]) })

	authCurrier := map[string]string{}
	authSection := map[string]string{}
	for _, leaf := range allLeaves {
		if st, ok := leafStatsBySource["ZL3b"][leaf]; ok {
			authCurrier[leaf] = dominant(st.currierCounts)
			authSection[leaf] = dominant(st.sectionCounts)
			continue
		}
		st := leafStatsBySource["IT2a"][leaf]
		authCurrier[leaf] = dominant(st.currierCounts)
		authSection[leaf] = dominant(st.sectionCounts)
	}

	// Stratify by (currier, section); within each stratum sort by leaf
	// number and assign deterministically.
	strataLeaves := map[string][]string{}
	for _, leaf := range allLeaves {
		key := authCurrier[leaf] + "/" + authSection[leaf]
		strataLeaves[key] = append(strataLeaves[key], leaf)
	}
	strataKeys := make([]string, 0, len(strataLeaves))
	for k := range strataLeaves {
		strataKeys = append(strataKeys, k)
	}
	sort.Strings(strataKeys)

	assignment := map[string]leafAssignment{}
	strataSizes := map[string]int{}
	for _, key := range strataKeys {
		leaves := strataLeaves[key]
		sort.Slice(leaves, func(i, j int) bool { return leafNumber(leaves[i]) < leafNumber(leaves[j]) })
		strataSizes[key] = len(leaves)
		for i, leaf := range leaves {
			assignment[leaf] = leafAssignment{
				leaf: leaf, currier: authCurrier[leaf], section: authSection[leaf],
				stratum: key, partition: partitionOf(i),
			}
		}
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	if err := writeSplitTSV(allLeaves, assignment, leafStatsBySource); err != nil {
		return err
	}

	if err := writeManifest(corpora, leafStatsBySource, allLeaves, assignment, strataSizes); err != nil {
		return err
	}

	fmt.Println("task85-analyze: wrote GRAMMAR_CORPUS_SPLIT.tsv and GRAMMAR_CORPUS_SPLIT_MANIFEST.json")
	return nil
}

func writeSplitTSV(allLeaves []string, assignment map[string]leafAssignment, leafStatsBySource map[string]map[string]*leafStats) error {
	var b strings.Builder
	b.WriteString("folio_id\tcurrier\tsection\tstratum\tpartition\tzl3b_page_ids\tzl3b_line_count\tzl3b_token_count\tit2a_page_ids\tit2a_line_count\tit2a_token_count\n")
	for _, leaf := range allLeaves {
		a := assignment[leaf]
		zl := leafStatsBySource["ZL3b"][leaf]
		it := leafStatsBySource["IT2a"][leaf]
		zlPages, zlLines, zlTok := "-", "0", "0"
		if zl != nil {
			zlPages = strings.Join(sortedPages(zl.pages), ";")
			zlLines = strconv.Itoa(zl.lineCount)
			zlTok = strconv.Itoa(zl.tokenCount)
		}
		itPages, itLines, itTok := "-", "0", "0"
		if it != nil {
			itPages = strings.Join(sortedPages(it.pages), ";")
			itLines = strconv.Itoa(it.lineCount)
			itTok = strconv.Itoa(it.tokenCount)
		}
		fmt.Fprintf(&b, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			leaf, a.currier, a.section, a.stratum, a.partition,
			zlPages, zlLines, zlTok, itPages, itLines, itTok)
	}
	return os.WriteFile(outDir+"/GRAMMAR_CORPUS_SPLIT.tsv", []byte(b.String()), 0644)
}

func sortedPages(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func computeCorpusStats(c corpusData, leaves map[string]*leafStats) corpusStats {
	vocab := map[string]bool{}
	glyphs := map[string]bool{}
	lineKeys := map[string]bool{}
	for _, t := range c.tokens {
		vocab[strings.Join(t.glyphs, "")] = true
		for _, g := range t.glyphs {
			glyphs[g] = true
		}
		lineKeys[t.lineKey] = true
	}
	return corpusStats{
		CanonicalPath: c.canonicalPath, IVTFFPath: c.ivtffPath,
		CanonicalSHA256: c.canonicalSHA256, IVTFFSHA256: c.ivtffSHA256,
		LineCount: len(lineKeys), TokenCount: len(c.tokens),
		VocabSize: len(vocab), GlyphInventory: len(glyphs), LeafCount: len(leaves),
	}
}

func computePartitionStats(leaves []string, assignment map[string]leafAssignment, part string, stats map[string]*leafStats) partitionStats {
	ps := partitionStats{SectionComposition: map[string]int{}, CurrierComposition: map[string]int{}}
	var lengthSum, lengthSumSq int64
	var n int64
	for _, leaf := range leaves {
		if assignment[leaf].partition != part {
			continue
		}
		st, ok := stats[leaf]
		if !ok {
			continue
		}
		ps.LeafCount++
		ps.TokenCount += st.tokenCount
		ps.LineCount += st.lineCount
		lengthSum += st.lengthSum
		lengthSumSq += st.lengthSumSq
		n += int64(st.tokenCount)
		for _, k := range sortedKeys(st.sectionCounts) {
			ps.SectionComposition[k] += st.sectionCounts[k]
		}
		for _, k := range sortedKeys(st.currierCounts) {
			ps.CurrierComposition[k] += st.currierCounts[k]
		}
	}
	if n > 0 {
		mean := float64(lengthSum) / float64(n)
		variance := float64(lengthSumSq)/float64(n) - mean*mean
		if variance < 0 {
			variance = 0
		}
		ps.MeanTokenLength = mean
		ps.SDTokenLength = math.Sqrt(variance)
	}
	return ps
}

func adjacentBoundaryCrossings(allLeaves []string, assignment map[string]leafAssignment) (crossings, total int) {
	for i := 1; i < len(allLeaves); i++ {
		prevN, curN := leafNumber(allLeaves[i-1]), leafNumber(allLeaves[i])
		if curN != prevN+1 {
			continue
		}
		total++
		if assignment[allLeaves[i-1]].partition != assignment[allLeaves[i]].partition {
			crossings++
		}
	}
	return
}

func gitHead() string {
	out, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func writeManifest(corpora map[string]corpusData, leafStatsBySource map[string]map[string]*leafStats, allLeaves []string, assignment map[string]leafAssignment, strataSizes map[string]int) error {
	totals := map[string]corpusStats{}
	for name, c := range corpora {
		totals[name] = computeCorpusStats(c, leafStatsBySource[name])
	}

	partitions := []string{development, validation, heldout}
	balance := map[string]map[string]partitionStats{}
	for name := range corpora {
		balance[name] = map[string]partitionStats{}
		for _, p := range partitions {
			balance[name][p] = computePartitionStats(allLeaves, assignment, p, leafStatsBySource[name])
		}
	}

	smallStrata := []string{}
	for _, k := range sortedKeys(strataSizes) {
		if strataSizes[k] < 5 {
			smallStrata = append(smallStrata, fmt.Sprintf("%s (n=%d)", k, strataSizes[k]))
		}
	}

	crossings, totalAdjacent := adjacentBoundaryCrossings(allLeaves, assignment)

	manifest := map[string]any{
		"schema":  "GRAMMAR_CORPUS_SPLIT_MANIFEST_V1",
		"version": "task85-v1",
		"git_commit_at_generation": gitHead(),
		"split_unit": "physical leaf (folio); recto/verso and multi-part foldout sides of the same leaf are always co-assigned",
		"stratification_key": "(Currier language $L, section $I), ZL3b-authoritative when a leaf is present in ZL3b, else IT2a",
		"assignment_rule": "deterministic positional rule, no PRNG/seed: within each stratum, leaves sorted by numeric leaf order; index%5 in {0,1,2}->DEVELOPMENT, 3->VALIDATION, 4->HELDOUT (60/20/20 target)",
		"small_strata_lt_5_leaves_all_or_mostly_development": smallStrata,
		"adjacent_leaf_partition_boundary_crossings": crossings,
		"adjacent_leaf_pairs_total": totalAdjacent,
		"leaf_count_total": len(allLeaves),
		"corpus_totals": totals,
		"partition_balance": balance,
	}
	b, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outDir+"/GRAMMAR_CORPUS_SPLIT_MANIFEST.json", b, 0644)
}
