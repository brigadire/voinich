// Command f01-speculum-analyze runs the task76 pre-registered
// experimental protocol for the F01 Speculum reconstruction: baseline
// recovery, prior-knowledge ablation, and state-corruption experiments.
// It never touches the Voynich corpus or any of its fingerprints -- see
// task76's explicit scope exclusions -- and only encodes/decodes the
// fixed message set defined in internal/speculumf01/messages.go.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"zcore.dev/voinich/internal/speculumf01"
)

const (
	numRings    = 12
	readRadius  = 5
	fillerSeed  = int64(20260823)
	controlSeed = int64(74760823)
)

func primaryConfig() speculumf01.Config {
	return speculumf01.Config{
		NumRings:           numRings,
		Alphabet:           speculumf01.Latin23,
		ReadRadius:         readRadius,
		Order:              speculumf01.InnerToOuter,
		RingIdentityMarked: true,
	}
}

func sensitivityConfig() speculumf01.Config {
	c := primaryConfig()
	c.Alphabet = speculumf01.Modern26
	return c
}

func fillerRNG(seed int64) func() int {
	state := seed
	return func() int {
		state = state*6364136223846793005 + 1442695040888963407
		v := int(state >> 33)
		if v < 0 {
			v = -v
		}
		return v
	}
}

// pilotPool is disjoint from NaturalMessages: words the operator has not
// been staring at while writing the experiment code, so a self-experiment
// pilot has a genuine (if weak) blinding property. See HUMAN_PILOT_LOG.md.
var pilotPool = []string{"VERITAS", "SILENTIVM", "CATENA", "MEMENTO", "PRAESENS"}

func main() {
	outDir := flag.String("out", "research/phase2/fontana/f01_speculum", "output directory")
	pilotGen := flag.String("pilot-gen", "", "generate a blinded human-pilot trial into this subdirectory (fresh-encode + fresh-corrupt), then exit")
	pilotCheck := flag.String("pilot-check", "", "reveal ground truth for a prior -pilot-gen directory and score a guess, then exit")
	pilotGuessA := flag.String("pilot-guess-a", "", "operator's decoded guess for trial A (intact state)")
	pilotGuessB := flag.String("pilot-guess-b", "", "operator's decoded guess for trial B (corrupted state)")
	flag.Parse()

	if *pilotGen != "" {
		runPilotGen(*pilotGen)
		return
	}
	if *pilotCheck != "" {
		runPilotCheck(*pilotCheck, *pilotGuessA, *pilotGuessB)
		return
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		panic(err)
	}

	lexicon := speculumf01.Lexicon()
	cfg := primaryConfig()

	messages := append([]speculumf01.TestMessage{}, speculumf01.NaturalMessages...)
	lengths := make([]int, len(speculumf01.NaturalMessages))
	for i, m := range speculumf01.NaturalMessages {
		lengths[i] = len([]rune(m.Text))
	}
	controls := speculumf01.GenerateRandomControls(cfg.Alphabet, lengths, controlSeed, lexicon)
	messages = append(messages, controls...)

	writeMessageSet(*outDir, messages)
	baselineRows := runBaseline(cfg, messages)
	writeBaseline(*outDir, baselineRows)

	ablationRows := runAblation(cfg, messages, lexicon)
	writeAblation(*outDir, ablationRows)

	corruptionRows := runCorruption(cfg, messages, lexicon)
	writeCorruption(*outDir, corruptionRows)

	writeInformationAccounting(*outDir, ablationRows, messages)

	// Reconstruction-profile sensitivity check: repeat baseline + K4 under
	// the Modern26 alternative alphabet profile to test whether the
	// 23-vs-26 letter assumption changes qualitative findings.
	sCfg := sensitivityConfig()
	sBaseline := runBaseline(sCfg, messages)
	sAblation := runAblation(sCfg, messages, lexicon)
	writeSensitivity(*outDir, sBaseline, sAblation)

	writeExample(*outDir, cfg, fillerRNG(fillerSeed))
	writeManifest(*outDir, cfg, len(messages))

	fmt.Println("f01-speculum-analyze: wrote results to", *outDir)
}

type baselineRow struct {
	Message  string
	Category string
	Natural  bool
	Decoded  string
	Exact    bool
}

func runBaseline(cfg speculumf01.Config, messages []speculumf01.TestMessage) []baselineRow {
	var rows []baselineRow
	for _, m := range messages {
		s, err := cfg.Encode(m.Text, fillerRNG(fillerSeed))
		if err != nil {
			rows = append(rows, baselineRow{Message: m.Text, Category: m.Category, Natural: m.Natural, Decoded: "<encode error: " + err.Error() + ">"})
			continue
		}
		decoded, err := cfg.DecodeFull(s, len([]rune(m.Text)))
		if err != nil {
			decoded = "<decode error: " + err.Error() + ">"
		}
		rows = append(rows, baselineRow{Message: m.Text, Category: m.Category, Natural: m.Natural, Decoded: decoded, Exact: decoded == m.Text})
	}
	return rows
}

type ablationRow struct {
	Message   string
	Category  string
	Natural   bool
	Condition speculumf01.Condition
	speculumf01.AblationResult
}

var allConditions = []speculumf01.Condition{
	speculumf01.CondFullKnowledge,
	speculumf01.CondLengthUnknown,
	speculumf01.CondRingOrderUnknown,
	speculumf01.CondReadRadiusUnknown,
	speculumf01.CondOrientUnknown,
	speculumf01.CondDirectionUnknown,
	speculumf01.CondConventionUnknown,
}

func runAblation(cfg speculumf01.Config, messages []speculumf01.TestMessage, lexicon map[string]bool) []ablationRow {
	var rows []ablationRow
	for _, m := range messages {
		s, err := cfg.Encode(m.Text, fillerRNG(fillerSeed))
		if err != nil {
			continue
		}
		lex := lexicon
		if !m.Natural {
			lex = nil // random controls: no language filter, by design
		}
		for _, cond := range allConditions {
			r := speculumf01.Evaluate(cfg, m.Text, s, cond, lex)
			rows = append(rows, ablationRow{Message: m.Text, Category: m.Category, Natural: m.Natural, Condition: cond, AblationResult: r})
		}
		// K8: no instruction at all -- qualitative, not combinatorial; see
		// F01_RECONSTRUCTION_DOSSIER.md. K9 handled by the corruption run
		// (it is state-metadata loss under full K, i.e. a corruption
		// scenario, not a new combinatorial condition).
	}
	return rows
}

type corruptionRow struct {
	Message  string
	Category string
	Natural  bool
	speculumf01.CorruptionMetrics
}

func runCorruption(cfg speculumf01.Config, messages []speculumf01.TestMessage, lexicon map[string]bool) []corruptionRow {
	var rows []corruptionRow
	for _, m := range messages {
		s, err := cfg.Encode(m.Text, fillerRNG(fillerSeed))
		if err != nil {
			continue
		}
		L := len([]rune(m.Text))
		mid := L / 2
		lex := lexicon
		if !m.Natural {
			lex = nil
		}

		scenarios := map[string]speculumf01.State{
			"single_position_substitution": speculumf01.SinglePositionDamage(s, cfg.RingPos(mid), (s.Offsets[cfg.RingPos(mid)]+7)%cfg.Alphabet.Size()),
			"random_ring_shift":            speculumf01.RandomRingShift(cfg, s, cfg.RingPos(mid), 5),
			"orientation_mark_loss":        speculumf01.LoseOrientationMark(cfg, s, 6),
		}
		if L >= 2 {
			scenarios["swap_two_elements"] = speculumf01.SwapTwoRings(s, cfg.RingPos(0), cfg.RingPos(1))
		}
		if numRings-2 >= 0 {
			scenarios["outer_contour_loss_2rings"] = speculumf01.LoseOuterContour(s, 2)
		}
		if L >= 4 {
			multi := s
			for _, wi := range []int{0, L / 3, (2 * L) / 3} {
				ring := cfg.RingPos(wi)
				multi = speculumf01.SinglePositionDamage(multi, ring, (multi.Offsets[ring]+11)%cfg.Alphabet.Size())
			}
			scenarios["multiple_independent_damages_3"] = multi
		}

		for name, damaged := range scenarios {
			collapse := false
			decoded := cfg.DecodeWithGap(damaged, L, collapse)
			rows = append(rows, corruptionRow{Message: m.Text, Category: m.Category, Natural: m.Natural,
				CorruptionMetrics: speculumf01.Measure(name, m.Text, decoded, lex, m.Natural)})
		}

		if L >= 3 {
			del := speculumf01.DeleteRing(s, cfg.RingPos(mid))
			localDecoded := cfg.DecodeWithGap(del, L, false)
			rows = append(rows, corruptionRow{Message: m.Text, Category: m.Category, Natural: m.Natural,
				CorruptionMetrics: speculumf01.Measure("deletion_ring_identity_marked", m.Text, localDecoded, lex, m.Natural)})
			cascDecoded := cfg.DecodeWithGap(del, L, true)
			rows = append(rows, corruptionRow{Message: m.Text, Category: m.Category, Natural: m.Natural,
				CorruptionMetrics: speculumf01.Measure("deletion_physical_collapse", m.Text, cascDecoded, lex, m.Natural)})
		}
	}
	// deterministic order: sort by message then scenario name
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Message != rows[j].Message {
			return rows[i].Message < rows[j].Message
		}
		return rows[i].Scenario < rows[j].Scenario
	})
	return rows
}

func writeMessageSet(dir string, messages []speculumf01.TestMessage) {
	var b strings.Builder
	b.WriteString("message\tlength\tcategory\tnatural\n")
	for _, m := range messages {
		fmt.Fprintf(&b, "%s\t%d\t%s\t%t\n", m.Text, len([]rune(m.Text)), m.Category, m.Natural)
	}
	writeFile(dir, "MESSAGE_SET.tsv", b.String())
}

func writeBaseline(dir string, rows []baselineRow) {
	var b strings.Builder
	b.WriteString("message\tcategory\tnatural\tdecoded\texact_recovery\n")
	exact := 0
	for _, r := range rows {
		fmt.Fprintf(&b, "%s\t%s\t%t\t%s\t%t\n", r.Message, r.Category, r.Natural, r.Decoded, r.Exact)
		if r.Exact {
			exact++
		}
	}
	fmt.Fprintf(&b, "# summary: %d/%d exact (%.4f), CI computed in TASK76_REPORT.md\n", exact, len(rows), float64(exact)/float64(len(rows)))
	writeFile(dir, "BASELINE_RESULTS.tsv", b.String())
}

func writeAblation(dir string, rows []ablationRow) {
	var b strings.Builder
	b.WriteString("message\tcategory\tnatural\tcondition\tn_candidates_raw\tn_candidates_lex\tenumeration_capped\texact_blind_p\texact_lex_p\tentropy_raw_bits\ttrue_in_set\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "%s\t%s\t%t\t%s\t%d\t%d\t%t\t%.6f\t%.6f\t%.4f\t%t\n",
			r.Message, r.Category, r.Natural, r.Condition, r.NCandidatesRaw, r.NCandidatesLex, r.EnumerationCapped,
			r.ExactBlindP, r.ExactLexP, r.EntropyRawBits, r.TrueMessageInSet)
	}
	writeFile(dir, "ABLATION_RESULTS.tsv", b.String())
}

func writeCorruption(dir string, rows []corruptionRow) {
	var b strings.Builder
	b.WriteString("message\tcategory\tnatural\tscenario\tdecoded\texact_recovery\tchar_error_rate\tfraction_after_first_error\terror_class\tdetectable\tcorrectable_without_m\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "%s\t%s\t%t\t%s\t%s\t%t\t%.4f\t%.4f\t%s\t%t\t%t\n",
			r.Message, r.Category, r.Natural, r.Scenario, r.Decoded, r.ExactRecovery, r.CharacterErrorRate,
			r.FractionAfterFirstError, r.ErrorClass, r.Detectable, r.CorrectableWithoutM)
	}
	writeFile(dir, "CORRUPTION_RESULTS.tsv", b.String())
}

func writeSensitivity(dir string, baseline []baselineRow, ablation []ablationRow) {
	var b strings.Builder
	b.WriteString("# Modern26 alphabet-profile sensitivity check (task76 sec. 'source allows multiple operational interpretations')\n")
	exact := 0
	for _, r := range baseline {
		if r.Exact {
			exact++
		}
	}
	fmt.Fprintf(&b, "baseline_exact_recovery\t%d/%d\n\n", exact, len(baseline))
	b.WriteString("condition\tmean_n_candidates_raw\n")
	sums := map[speculumf01.Condition]int{}
	counts := map[speculumf01.Condition]int{}
	for _, r := range ablation {
		sums[r.Condition] += r.NCandidatesRaw
		counts[r.Condition]++
	}
	for _, cond := range allConditions {
		if counts[cond] == 0 {
			continue
		}
		fmt.Fprintf(&b, "%s\t%.2f\n", cond, float64(sums[cond])/float64(counts[cond]))
	}
	writeFile(dir, "ALPHABET_SENSITIVITY.tsv", b.String())
}

func writeInformationAccounting(dir string, rows []ablationRow, messages []speculumf01.TestMessage) {
	type agg struct {
		sumRaw, sumLex float64
		n              int
	}
	byCondNatural := map[speculumf01.Condition]*agg{}
	byCondRandom := map[speculumf01.Condition]*agg{}
	for _, r := range rows {
		m := byCondNatural
		if !r.Natural {
			m = byCondRandom
		}
		a, ok := m[r.Condition]
		if !ok {
			a = &agg{}
			m[r.Condition] = a
		}
		a.sumRaw += float64(r.NCandidatesRaw)
		a.sumLex += float64(r.NCandidatesLex)
		a.n++
	}
	var b strings.Builder
	b.WriteString("condition\tmean_N_raw_natural\tmean_N_lex_natural\tmean_N_raw_random\tlanguage_narrowing_factor\n")
	for _, cond := range allConditions {
		nAgg, rAgg := byCondNatural[cond], byCondRandom[cond]
		if nAgg == nil || rAgg == nil || nAgg.n == 0 || rAgg.n == 0 {
			continue
		}
		meanRawNat := nAgg.sumRaw / float64(nAgg.n)
		meanLexNat := nAgg.sumLex / float64(nAgg.n)
		meanRawRand := rAgg.sumRaw / float64(rAgg.n)
		narrowing := 1.0
		if meanLexNat > 0 {
			narrowing = meanRawNat / meanLexNat
		}
		fmt.Fprintf(&b, "%s\t%.2f\t%.2f\t%.2f\t%.2f\n", cond, meanRawNat, meanLexNat, meanRawRand, narrowing)
	}
	writeFile(dir, "INFORMATION_ACCOUNTING.tsv", b.String())
}

func writeExample(dir string, cfg speculumf01.Config, filler func() int) {
	s, err := cfg.Encode("MEMORIA", filler)
	if err != nil {
		panic(err)
	}
	writeFile(dir, "example_state.txt", cfg.RenderASCII(s))
	writeFile(dir, "example_state.svg", cfg.RenderSVG(s))
	yamlBytes, err := speculumf01.MarshalState(cfg, s)
	if err != nil {
		panic(err)
	}
	writeFile(dir, "example_state.yaml", string(yamlBytes))
}

func writeManifest(dir string, cfg speculumf01.Config, nMessages int) {
	manifest := map[string]any{
		"num_rings":            cfg.NumRings,
		"alphabet":             cfg.Alphabet.Name,
		"alphabet_size":        cfg.Alphabet.Size(),
		"read_radius":          cfg.ReadRadius,
		"order":                "inner_to_outer",
		"ring_identity_marked": cfg.RingIdentityMarked,
		"filler_seed":          fillerSeed,
		"control_seed":         controlSeed,
		"n_natural_messages":   len(speculumf01.NaturalMessages),
		"n_total_messages":     nMessages,
		"lexicon_size":         len(speculumf01.BaseLexicon),
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		panic(err)
	}
	writeFile(dir, "manifest.json", string(data))
}

func writeFile(dir, name, content string) {
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		panic(err)
	}
}

// runPilotGen realizes task76 Block 6's blinded self-experiment setup. It
// picks trial words using math/rand seeded from real wall-clock time --
// deliberately NOT reproducible -- so the operator running the CLI cannot
// have predicted the draw from the source code. Trial A is an intact
// baseline read; Trial B applies one single-position substitution to a
// second, independently drawn word so the operator must also notice and
// characterize an error. Ground truth is written to a separate file the
// operator must not open before recording a guess.
func runPilotGen(dir string) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		panic(err)
	}
	cfg := primaryConfig()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	wordA := pilotPool[r.Intn(len(pilotPool))]
	wordB := pilotPool[r.Intn(len(pilotPool))]
	for wordB == wordA {
		wordB = pilotPool[r.Intn(len(pilotPool))]
	}

	sA, err := cfg.Encode(wordA, fillerRNG(time.Now().UnixNano()))
	if err != nil {
		panic(err)
	}
	sB, err := cfg.Encode(wordB, fillerRNG(time.Now().UnixNano()+1))
	if err != nil {
		panic(err)
	}
	mid := len([]rune(wordB)) / 2
	ring := cfg.RingPos(mid)
	damageDelta := 1 + r.Intn(cfg.Alphabet.Size()-1)
	sB = speculumf01.SinglePositionDamage(sB, ring, (sB.Offsets[ring]+damageDelta)%cfg.Alphabet.Size())

	writeFile(dir, "TRIAL_A_STATE.txt", cfg.RenderASCII(sA))
	writeFile(dir, "TRIAL_A_STATE.svg", cfg.RenderSVG(sA))
	writeFile(dir, "TRIAL_A_KNOWN_PARAMETERS.txt", fmt.Sprintf(
		"alphabet=%s read_radius=%d order=inner_to_outer message_length=%d\n(full knowledge K1; state intact)\n",
		cfg.Alphabet.Name, cfg.ReadRadius, len([]rune(wordA))))

	writeFile(dir, "TRIAL_B_STATE.txt", cfg.RenderASCII(sB))
	writeFile(dir, "TRIAL_B_STATE.svg", cfg.RenderSVG(sB))
	writeFile(dir, "TRIAL_B_KNOWN_PARAMETERS.txt", fmt.Sprintf(
		"alphabet=%s read_radius=%d order=inner_to_outer message_length=%d\n(full knowledge K1; state MAY be corrupted -- unknown to operator whether/where)\n",
		cfg.Alphabet.Name, cfg.ReadRadius, len([]rune(wordB))))

	writeFile(dir, "GROUND_TRUTH_DO_NOT_OPEN_BEFORE_GUESSING.txt", fmt.Sprintf(
		"trial_a_message=%s\ntrial_b_message=%s\ntrial_b_damaged_ring=%d\ntrial_b_damage_delta=%d\n",
		wordA, wordB, ring, damageDelta))

	fmt.Println("pilot trial generated in", dir, "-- read TRIAL_A_STATE.txt and TRIAL_B_STATE.txt, record guesses, THEN run -pilot-check")
}

func runPilotCheck(dir, guessA, guessB string) {
	data, err := os.ReadFile(filepath.Join(dir, "GROUND_TRUTH_DO_NOT_OPEN_BEFORE_GUESSING.txt"))
	if err != nil {
		panic(err)
	}
	fmt.Println("--- ground truth ---")
	fmt.Print(string(data))
	fmt.Println("--- operator guesses ---")
	fmt.Println("trial_a_guess:", guessA)
	fmt.Println("trial_b_guess:", guessB)
}
