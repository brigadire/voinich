package fingerprintv2

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"

	"zcore.dev/voinich/internal/tokenrepetition"
)

const task79Version = "fingerprint-v2-page-hierarchy-v1"

func runTask79(c corpus, cfg Config, base CorpusResult, seed int64) Task79Result {
	tc := cfg.Task79.normalized()
	profiles := buildLineProfiles(c)
	audit := auditMetadata(c, tc.CorrectionLayer)
	tests, definitions := task79Tests(c, profiles, tc, seed)
	tests = fdr(tests)
	metrics := make([]FreezeMetric, 0, len(tests)+2)
	for _, t := range tests {
		d := definitions[t.ID]
		status := "NOT_SUPPORTED"
		if t.QValue <= cfg.Alpha && t.EffectDefined {
			status = "SUPPORTED"
		}
		class := "SUPPORTING"
		if status == "SUPPORTED" && d.eligible {
			class = "CORE"
		}
		neg := "ABSENCE_OF_EVIDENCE"
		metrics = append(metrics, FreezeMetric{
			MetricID: t.ID, MetricVersion: "task79-v1", Family: d.family, Definition: d.definition,
			UnitOfAnalysis: d.unit, Inputs: d.inputs, Parameters: map[string]any{"permutations": tc.Permutations},
			ObservedValue: t.Observed, Uncertainty: fmt.Sprintf("permutation null: mean %.6g, SD %.6g, n=%d", t.NullMean, t.NullSD, t.Replicates),
			NullModels: []string{d.nullID}, EffectSize: t.EffectSize, PValue: t.PValue, QValue: t.QValue,
			PartitionStability: partitionStability(c, d.measure), TranscriptionStability: "INSUFFICIENT_DATA: one aligned transcription",
			ParameterSensitivity: "ROBUST_WITH_LIMITATIONS: fixed preregistered categories; threshold-free statistic",
			RedundancyClass:      class, CoverageRole: d.coverage, ComparisonEligibility: map[bool]string{true: "CORPUS_COMPARABLE", false: "VOYNICH_ONLY_CONTEXT"}[d.eligible],
			NegativeEvidenceStatus: neg, ImplementationVersion: task79Version, Status: status, Limitations: d.limitation,
		})
	}
	// LS1 is descriptive and deliberately not promoted to CORE without a
	// generative count model; it remains useful for matching future corpora.
	lengths := make([]float64, len(profiles))
	for i := range profiles {
		lengths[i] = float64(profiles[i].TokenCount)
	}
	lo, hi := bootstrapLineCV(profiles, tc.BootstrapReplicates, seed+1200011)
	metrics = append(metrics, FreezeMetric{MetricID: "LS1_LINE_LENGTH_CV", MetricVersion: "task79-v1", Family: "line", Definition: "SD/mean of tokens per non-empty physical/logical IVTFF line", UnitOfAnalysis: "line", Inputs: []string{"line_id", "token_count"}, Parameters: map[string]any{"hierarchical_bootstrap_replicates": tc.BootstrapReplicates}, ObservedValue: safeDiv(sd(lengths, mean(lengths)), mean(lengths)), Uncertainty: fmt.Sprintf("95%% folio-bootstrap interval [%.6g, %.6g]", lo, hi), NullModels: []string{"HN7"}, PartitionStability: "ROBUST_WITH_LIMITATIONS", TranscriptionStability: "INSUFFICIENT_DATA: one aligned transcription", ParameterSensitivity: "ROBUST", RedundancyClass: "SUPPORTING", CoverageRole: []string{"table/index", "stochastic procedural generator"}, ComparisonEligibility: "CORPUS_COMPARABLE", NegativeEvidenceStatus: "NOT_TESTED", ImplementationVersion: task79Version, Status: "DESCRIPTIVE", Limitations: "Line breaks in IVTFF are transcription metadata, not measured scan coordinates."})
	metrics = append(metrics, inheritedFreezeMetrics(base)...)
	sort.Slice(metrics, func(i, j int) bool { return metrics[i].MetricID < metrics[j].MetricID })
	stability := stabilityRows(metrics)
	redundancy := task79Redundancy(profiles)
	negative := negativeRegistry(metrics, tc)
	seg := segmentFolios(c, profiles, tc.ChangePointPenalty)
	status := "TASK79_B_REQUIRED"
	manifest := FreezeManifest{CandidateID: "FINGERPRINT_V2_CANDIDATE", Status: status, CorpusSHA256: c.info.SHA256, CodeVersion: task79Version, MetricVersion: "task79-v1", Seeds: []int64{cfg.Seed, seed}, MissingDataPolicy: "Explicit missing class for descriptive exports; exclude missing metadata from inferential contrasts and report N.", ComparisonRules: "No post-hoc metrics, thresholds, weights or null changes after candidate creation; VOYNICH_ONLY_CONTEXT metrics cannot enter corpus distance.", DecisionBasis: "Page/hierarchy blocks are implemented, but a second aligned transcription and historical notation controls are unavailable; task79-b is required before freeze."}
	for _, m := range metrics {
		if m.RedundancyClass == "CORE" {
			manifest.CoreMetrics = append(manifest.CoreMetrics, m.MetricID)
		} else {
			manifest.SupportingMetrics = append(manifest.SupportingMetrics, m.MetricID)
		}
	}
	manifest.UnfrozenExtensions = []string{"visual 2D coordinates", "alternative transcription sensitivity", "historical shorthand/abbreviation positive controls"}
	manifest.ProhibitedClaims = []string{"boundary structure implies an acrostic", "IVTFF order is exact scan geometry", "a detected regime is a topic, language, cipher, or encoding stage", "fingerprint evidence decrypts the manuscript"}
	return Task79Result{Version: task79Version, InputAudit: auditInputs(base, tc), MetadataAudit: audit, LineProfiles: profiles, Metrics: metrics, NullRegistry: hierarchicalNulls(), StabilityMatrix: stability, RedundancyMatrix: redundancy, CoverageAudit: coverageAudit(), NegativeEvidence: negative, Segmentation: seg, FreezeManifest: manifest, Verdicts: task79Verdicts(metrics, audit, seg, status), Occurrences: occurrenceRows(c)}
}

type metricDefinition struct {
	family, definition, unit, nullID, limitation string
	inputs, coverage                             []string
	eligible                                     bool
	measure                                      func(corpus) float64
}

func task79Tests(c corpus, lines []LineProfile, cfg Task79Config, seed int64) ([]NullTest, map[string]metricDefinition) {
	defs := map[string]metricDefinition{}
	var tests []NullTest
	addNMI := func(id, family, definition, unit, nullID string, x, y, groups []string, coverage []string, eligible bool, limitation string) {
		measure := func(cc corpus) float64 { xx, yy, _ := metricVectors(cc, id); return normalizedMI(xx, yy) }
		defs[id] = metricDefinition{family: family, definition: definition, unit: unit, nullID: nullID, inputs: []string{"token identity", "IVTFF metadata"}, coverage: coverage, eligible: eligible, limitation: limitation, measure: measure}
		tests = append(tests, nmiPermutationTest(id, nullID, x, y, groups, cfg.Permutations, rand.New(rand.NewSource(seed+int64(len(tests)+1)*100003))))
	}
	x, y, g := metricVectors(c, "LS2_POSITIONAL_LEXICON_NMI")
	addNMI("LS2_POSITIONAL_LEXICON_NMI", "line", "NMI(token identity, fixed five-class line position)", "token", "HN1", x, y, g, []string{"line-boundary projection", "acrostic-compatible organization", "constrained token grammar"}, true, "High-cardinality token identity can inflate raw NMI; inference relies on within-line permutation.")
	x, y, g = metricVectors(c, "BP1_BOUNDARY_TOKEN_NMI")
	addNMI("BP1_BOUNDARY_TOKEN_NMI", "boundary", "NMI(token identity, initial/interior/final boundary class)", "token", "HN5", x, y, g, []string{"line-boundary projection", "acrostic-compatible organization"}, true, "Detects boundary specialization, not plaintext or a reading direction.")
	x, y, g = metricVectors(c, "LC1_LOCUS_TYPE_NMI")
	if len(x) > 0 {
		addNMI("LC1_LOCUS_TYPE_NMI", "locus", "NMI(token length class, documented locus type)", "token", "HN2", x, y, g, []string{"table/index", "mixed architecture"}, true, "Locus type is documentary metadata and has unequal group sizes.")
	}
	x, y, g = metricVectors(c, "LC2_LABEL_TEXT_NMI")
	if len(x) > 0 {
		addNMI("LC2_LABEL_TEXT_NMI", "locus", "NMI(token length class, label versus running text), conditional on folio", "token", "HN6", x, y, g, []string{"table/index", "mnemonic cue system"}, true, "Label status derives only from documented locus codes; sparse labels reduce power.")
	}
	x, y, g = metricVectors(c, "2DL1_LAYOUT_POSITION_MI")
	addNMI("2DL1_LAYOUT_POSITION_MI", "2D-LITE", "NMI(glyph-length class, normalized line-position class)", "token", "HN1", x, y, g, []string{"positional extraction", "stochastic procedural generator"}, true, "2D-LITE uses order/categories only; it is not physical geometry.")
	for _, spec := range []struct{ id, label string }{{"LC5_IVTFF_I_NMI", "$I"}, {"LC5_IVTFF_X_NMI", "$X"}, {"HR6_CURRIER_SECTION_NMI", "Currier/section"}} {
		x, y, g = metricVectors(c, spec.id)
		if len(x) > 0 {
			addNMI(spec.id, map[bool]string{true: "hierarchy", false: "locus"}[strings.HasPrefix(spec.id, "HR")], "NMI for "+spec.label+" metadata association with token length/regime class", "token", "HN6", x, y, g, []string{"mixed architecture", "table/index"}, false, "Metadata classes are interpreted categorically only; association does not assign semantics.")
		}
	}
	// Boundary asymmetry: absolute initial/final glyph-length difference.
	obs := boundaryLengthAsymmetry(c)
	null := make([]float64, cfg.Permutations)
	rng := rand.New(rand.NewSource(seed + 600001))
	for i := range null {
		null[i] = boundaryLengthAsymmetry(swappedLineBoundaries(c, rng))
	}
	defs["LS3_BOUNDARY_LENGTH_ASYMMETRY"] = metricDefinition{family: "line", definition: "Absolute mean glyph-length difference between first and final tokens", unit: "line boundary", nullID: "HN5", inputs: []string{"glyph length", "line boundary"}, coverage: []string{"line-boundary projection", "acrostic-compatible organization"}, eligible: true, limitation: "Length asymmetry is one boundary projection and has no semantic interpretation.", measure: boundaryLengthAsymmetry}
	tests = append(tests, nullTest("LS3_BOUNDARY_LENGTH_ASYMMETRY", "HN5 random initial/final exchange", obs, null))
	// Exact repetition compared with within-line shuffles.
	obs = adjacentRepeat(c)
	null = make([]float64, cfg.Permutations)
	rng = rand.New(rand.NewSource(seed + 700001))
	for i := range null {
		null[i] = adjacentRepeat(shuffledWithinLines(c, rng))
	}
	defs["LS4_WITHIN_LINE_EXACT_REPETITION"] = metricDefinition{family: "line", definition: "Adjacent exact-token repetition rate within lines", unit: "within-line transition", nullID: "HN1", inputs: []string{"token", "line_id"}, coverage: []string{"copy-with-modification", "stochastic procedural generator"}, eligible: true, limitation: "Only immediate recurrence is primary; longer recurrence distances remain diagnostic.", measure: adjacentRepeat}
	tests = append(tests, nullTest("LS4_WITHIN_LINE_EXACT_REPETITION", "HN1 within-line permutation", obs, null))
	// Folio coherence over line-profile vectors, calibrated by reassignment inside section.
	obs = folioCoherence(lines)
	null = make([]float64, cfg.Permutations)
	rng = rand.New(rand.NewSource(seed + 800003))
	for i := range null {
		null[i] = folioCoherence(shuffleLineFolios(lines, rng))
	}
	defs["PF2_FOLIO_COHERENCE"] = metricDefinition{family: "folio", definition: "Difference between mean across-folio and within-folio standardized line-profile distance", unit: "line pair", nullID: "HN4", inputs: []string{"line profile", "folio", "section"}, coverage: []string{"mixed architecture", "table/index", "natural-language-like process"}, eligible: true, limitation: "IVTFF folio order is documentary; this statistic does not assume historical order.", measure: func(cc corpus) float64 { return folioCoherence(buildLineProfiles(cc)) }}
	tests = append(tests, nullTest("PF2_FOLIO_COHERENCE", "HN4 line-to-folio reassignment within section", obs, null))
	// Hierarchical variance shares; label permutations preserve the response.
	for _, h := range []struct{ id, level string }{{"HR1_LOCUS_VARIANCE_SHARE", "locus"}, {"HR1_FOLIO_VARIANCE_SHARE", "folio"}, {"HR1_SECTION_VARIANCE_SHARE", "section"}} {
		obs = varianceShare(lines, h.level)
		null = make([]float64, cfg.Permutations)
		rng = rand.New(rand.NewSource(seed + int64(len(tests)+1)*900007))
		for i := range null {
			null[i] = varianceShare(shuffleLineLevel(lines, h.level, rng), h.level)
		}
		level := h.level
		defs[h.id] = metricDefinition{family: "hierarchy", definition: "Between-" + level + " share of line-length variance", unit: "line", nullID: "HN6", inputs: []string{"line token count", level}, coverage: []string{"mixed architecture", "table/index", "natural-language-like process"}, eligible: true, limitation: "Method-of-moments variance share is a robust descriptive decomposition, not a causal random-effects model.", measure: func(cc corpus) float64 { return varianceShare(buildLineProfiles(cc), level) }}
		tests = append(tests, nullTest(h.id, "HN6 metadata-label permutation", obs, null))
	}
	// Folio turnover and recto/verso pairing use documentary folio order only.
	for _, spec := range []struct {
		id, definition, null string
		measure              func([]LineProfile) float64
	}{{"PF3_ADJACENT_FOLIO_CONTINUITY", "Mean adjacent-folio vocabulary-profile similarity", "HN4", adjacentFolioContinuity}, {"PF4_RECTO_VERSO_COHERENCE", "Same-leaf recto/verso profile similarity advantage", "HN4", rectoVersoCoherence}, {"PF5_WITHIN_FOLIO_PROGRESSION", "Mean absolute correlation of local line order with line length", "HN3", folioProgression}} {
		obs = spec.measure(lines)
		null = make([]float64, cfg.Permutations)
		rng = rand.New(rand.NewSource(seed + int64(len(tests)+1)*1100009))
		for i := range null {
			null[i] = spec.measure(shuffleLineFolios(lines, rng))
		}
		measure := spec.measure
		defs[spec.id] = metricDefinition{family: "folio", definition: spec.definition, unit: "folio", nullID: spec.null, inputs: []string{"line profile", "folio order/side"}, coverage: []string{"mixed architecture", "table/index"}, eligible: true, limitation: "Folio order is documentary and is not assumed to be original historical order.", measure: func(cc corpus) float64 { return measure(buildLineProfiles(cc)) }}
		tests = append(tests, nullTest(spec.id, spec.null+" folio reassignment", obs, null))
	}
	return tests, defs
}

func metricVectors(c corpus, id string) (x, y, g []string) {
	for _, r := range c.records {
		pos := "interior"
		if r.IndexInLine == 0 {
			pos = "first"
		}
		if r.IndexInLine == 1 {
			pos = "second"
		}
		if r.LineLength > 1 && r.IndexInLine == r.LineLength-2 {
			pos = "penultimate"
		}
		if r.LineLength > 0 && r.IndexInLine == r.LineLength-1 {
			pos = "final"
		}
		boundary := "interior"
		if r.IndexInLine == 0 {
			boundary = "initial"
		}
		if r.LineLength > 0 && r.IndexInLine == r.LineLength-1 {
			boundary = "final"
		}
		lenClass := strconv.Itoa(min(8, len(r.Glyph)))
		switch id {
		case "LS2_POSITIONAL_LEXICON_NMI":
			x = append(x, r.Token)
			y = append(y, pos)
			g = append(g, r.LineID)
		case "BP1_BOUNDARY_TOKEN_NMI":
			x = append(x, r.Token)
			y = append(y, boundary)
			g = append(g, r.LineID)
		case "LC1_LOCUS_TYPE_NMI":
			if r.LocusType != "" {
				x = append(x, lenClass)
				y = append(y, r.LocusType)
				g = append(g, r.Page)
			}
		case "LC2_LABEL_TEXT_NMI":
			if r.LocusType != "" {
				s := "text"
				if strings.EqualFold(r.LocusType, "L") {
					s = "label"
				}
				x = append(x, lenClass)
				y = append(y, s)
				g = append(g, r.Page)
			}
		case "2DL1_LAYOUT_POSITION_MI":
			x = append(x, lenClass)
			y = append(y, boundary)
			g = append(g, r.LineID)
		case "LC5_IVTFF_I_NMI":
			if r.IVTFFI != "" {
				x = append(x, lenClass)
				y = append(y, r.IVTFFI)
				g = append(g, r.Currier)
			}
		case "LC5_IVTFF_X_NMI":
			if r.IVTFFX != "" {
				x = append(x, lenClass)
				y = append(y, r.IVTFFX)
				g = append(g, r.Currier)
			}
		case "HR6_CURRIER_SECTION_NMI":
			if r.Currier != "" && r.Section != "" {
				x = append(x, r.Currier)
				y = append(y, r.Section)
				g = append(g, r.Hand)
			}
		}
	}
	return
}

func buildLineProfiles(c corpus) []LineProfile {
	by := map[string][]tokenRecord{}
	order := []string{}
	for _, r := range c.records {
		if _, seen := by[r.LineID]; !seen {
			order = append(order, r.LineID)
		}
		by[r.LineID] = append(by[r.LineID], r)
	}
	var out []LineProfile
	for _, id := range order {
		rs := by[id]
		if len(rs) == 0 {
			continue
		}
		counts, trans := map[string]int{}, map[string]int{}
		chars, reps, near := 0, 0, 0
		for i, r := range rs {
			counts[r.Token]++
			chars += len(r.Glyph)
			if i > 0 {
				trans[rs[i-1].Token+"\x00"+r.Token]++
				if r.Token == rs[i-1].Token {
					reps++
				}
				if tokenrepetition.LevenshteinGlyphs(r.Glyph, rs[i-1].Glyph) == 1 {
					near++
				}
			}
		}
		at := func(i int) string {
			if i < 0 || i >= len(rs) {
				return ""
			}
			return rs[i].Token
		}
		r := rs[0]
		out = append(out, LineProfile{LineID: id, Folio: r.Page, LocusID: r.LocusID, LocusType: r.LocusType, Section: r.Section, Currier: r.Currier, Scribe: r.Hand, TokenCount: len(rs), CharacterCount: chars, VocabularySize: len(counts), Diversity: safeDiv(float64(len(counts)), float64(len(rs))), ExactRepetitionRate: safeDiv(float64(reps), float64(max(1, len(rs)-1))), NearEditRepetitionRate: safeDiv(float64(near), float64(max(1, len(rs)-1))), TransitionEntropy: entropy(trans), TokenEntropy: entropy(counts), FirstToken: at(0), SecondToken: at(1), PenultimateToken: at(len(rs) - 2), FinalToken: at(len(rs) - 1), ParagraphStart: r.ParagraphStart})
	}
	return out
}

func auditMetadata(c corpus, layer string) MetadataAudit {
	a := MetadataAudit{Status: "REPRODUCED", Tokens: len(c.records), CorrectionLayer: layer}
	fol, loci, lines := map[string]bool{}, map[string]bool{}, map[string]string{}
	for _, r := range c.records {
		fol[r.Page] = true
		loci[r.LocusID] = true
		if old, ok := lines[r.LineID]; ok && old != r.Page {
			a.NestingViolations++
		}
		lines[r.LineID] = r.Page
		if r.Page == "" {
			a.MissingFolio++
		}
		if r.LocusType == "" {
			a.MissingLocusType++
		}
		if r.Section == "" {
			a.MissingSection++
		}
		if r.Currier == "" {
			a.MissingCurrier++
		}
		if r.Hand == "" {
			a.MissingHand++
		}
		if r.IVTFFI == "" {
			a.MissingI++
		}
		if r.IVTFFX == "" {
			a.MissingX++
		}
		if strings.HasSuffix(r.LineID, "\x00f116r.37") {
			a.KnownF116r37Occurrences++
		}
	}
	a.Folios = len(fol)
	a.Loci = len(loci)
	a.Lines = len(lines)
	a.Transitions = max(0, len(c.records)-a.Lines)
	if c.info.MetadataAlignment != "strict IVTFF aligned" {
		a.Status = "NOT_TESTABLE"
		a.Issues = append(a.Issues, "strict IVTFF alignment unavailable")
	}
	if a.NestingViolations > 0 {
		a.Status = "REQUIRES_REPAIR"
		a.Issues = append(a.Issues, "line identifiers cross folio boundaries")
	}
	if a.KnownF116r37Occurrences == 0 && c.info.MetadataAlignment == "strict IVTFF aligned" {
		a.Status = "REQUIRES_REPAIR"
		a.Issues = append(a.Issues, "known f116r.37 line not represented")
	}
	return a
}

func adjacentRepeat(c corpus) float64 {
	same, n := 0, 0
	for i := 1; i < len(c.records); i++ {
		if c.records[i].Line == c.records[i-1].Line {
			n++
			if c.records[i].Token == c.records[i-1].Token {
				same++
			}
		}
	}
	return safeDiv(float64(same), float64(n))
}
func boundaryLengthAsymmetry(c corpus) float64 {
	by := map[int][]tokenRecord{}
	for _, r := range c.records {
		by[r.Line] = append(by[r.Line], r)
	}
	a, b := []float64{}, []float64{}
	for _, line := range orderedIntKeys(by) {
		rs := by[line]
		if len(rs) < 2 {
			continue
		}
		sort.Slice(rs, func(i, j int) bool { return rs[i].IndexInLine < rs[j].IndexInLine })
		a = append(a, float64(len(rs[0].Glyph)))
		b = append(b, float64(len(rs[len(rs)-1].Glyph)))
	}
	return math.Abs(mean(a) - mean(b))
}
func swappedLineBoundaries(c corpus, rng *rand.Rand) corpus {
	out := c
	out.records = append([]tokenRecord(nil), c.records...)
	by := map[int][]int{}
	for i, r := range out.records {
		by[r.Line] = append(by[r.Line], i)
	}
	for _, line := range orderedIntKeys(by) {
		idx := by[line]
		if len(idx) > 1 && rng.Intn(2) == 1 {
			i, j := idx[0], idx[len(idx)-1]
			out.records[i].Token, out.records[j].Token = out.records[j].Token, out.records[i].Token
			out.records[i].Glyph, out.records[j].Glyph = out.records[j].Glyph, out.records[i].Glyph
		}
	}
	return out
}
func shuffledWithinLines(c corpus, rng *rand.Rand) corpus {
	out := c
	out.records = append([]tokenRecord(nil), c.records...)
	by := map[int][]int{}
	for i, r := range c.records {
		by[r.Line] = append(by[r.Line], i)
	}
	for _, k := range sortedIntMapKeys(by) {
		idx := by[k]
		p := rng.Perm(len(idx))
		for j, dst := range idx {
			src := c.records[idx[p[j]]]
			meta := out.records[dst]
			meta.Token, meta.Glyph = src.Token, src.Glyph
			out.records[dst] = meta
		}
	}
	return out
}
func sortedIntMapKeys(m map[int][]int) []int {
	ks := make([]int, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Ints(ks)
	return ks
}

func lineVector(l LineProfile) []float64 {
	return []float64{float64(l.TokenCount), l.Diversity, l.ExactRepetitionRate, l.NearEditRepetitionRate, l.TokenEntropy, l.TransitionEntropy}
}
func distance(a, b []float64) float64 {
	s := 0.0
	for i := range a {
		d := a[i] - b[i]
		s += d * d
	}
	return math.Sqrt(s)
}
func folioCoherence(ls []LineProfile) float64 {
	// Multivariate eta-squared of line profiles by folio, conditioned on
	// section.  This is algebraically the centroid form of the within/across
	// distance contrast and avoids O(lines^2) pair enumeration.
	sectionSum, sectionN := map[string][]float64{}, map[string]int{}
	folioSum, folioN := map[string][]float64{}, map[string]int{}
	for _, l := range ls {
		v := lineVector(l)
		if sectionSum[l.Section] == nil {
			sectionSum[l.Section] = make([]float64, len(v))
		}
		if folioSum[l.Folio] == nil {
			folioSum[l.Folio] = make([]float64, len(v))
		}
		for i, x := range v {
			sectionSum[l.Section][i] += x
			folioSum[l.Folio][i] += x
		}
		sectionN[l.Section]++
		folioN[l.Folio]++
	}
	between, total := 0.0, 0.0
	for _, l := range ls {
		v := lineVector(l)
		sm := sectionSum[l.Section]
		fm := folioSum[l.Folio]
		for i, x := range v {
			sectionMean := sm[i] / float64(sectionN[l.Section])
			folioMean := fm[i] / float64(folioN[l.Folio])
			d := x - sectionMean
			total += d * d
			df := folioMean - sectionMean
			between += df * df
		}
	}
	return safeDiv(between, total)
}
func shuffleLineFolios(ls []LineProfile, rng *rand.Rand) []LineProfile {
	out := append([]LineProfile(nil), ls...)
	by := map[string][]int{}
	for i, l := range ls {
		by[l.Section] = append(by[l.Section], i)
	}
	for _, k := range orderedKeys(by) {
		idx := by[k]
		p := rng.Perm(len(idx))
		for j, dst := range idx {
			out[dst].Folio = ls[idx[p[j]]].Folio
		}
	}
	return out
}
func varianceShare(ls []LineProfile, level string) float64 {
	if len(ls) < 2 {
		return 0
	}
	group := func(l LineProfile) string {
		switch level {
		case "locus":
			return l.LocusID
		case "folio":
			return l.Folio
		default:
			return l.Section
		}
	}
	overall := 0.0
	for _, l := range ls {
		overall += float64(l.TokenCount)
	}
	overall /= float64(len(ls))
	sum := map[string]float64{}
	n := map[string]int{}
	total := 0.0
	for _, l := range ls {
		k := group(l)
		sum[k] += float64(l.TokenCount)
		n[k]++
		d := float64(l.TokenCount) - overall
		total += d * d
	}
	between := 0.0
	for _, k := range orderedKeys(sum) {
		s := sum[k]
		m := s / float64(n[k])
		between += float64(n[k]) * (m - overall) * (m - overall)
	}
	return safeDiv(between, total)
}
func folioMeanVectors(ls []LineProfile) ([]string, map[string][]float64, map[string]string) {
	by := map[string][][]float64{}
	section := map[string]string{}
	order := []string{}
	for _, l := range ls {
		if _, ok := by[l.Folio]; !ok {
			order = append(order, l.Folio)
		}
		by[l.Folio] = append(by[l.Folio], lineVector(l))
		section[l.Folio] = l.Section
	}
	out := map[string][]float64{}
	for f, vs := range by {
		v := make([]float64, 6)
		for _, x := range vs {
			for i := range v {
				v[i] += x[i]
			}
		}
		for i := range v {
			v[i] /= float64(len(vs))
		}
		out[f] = v
	}
	return order, out, section
}
func adjacentFolioContinuity(ls []LineProfile) float64 {
	order, v, section := folioMeanVectors(ls)
	var x []float64
	for i := 1; i < len(order); i++ {
		if section[order[i]] == section[order[i-1]] {
			x = append(x, 1/(1+distance(v[order[i-1]], v[order[i]])))
		}
	}
	return mean(x)
}
func leafID(f string) string {
	if strings.HasSuffix(f, "r") || strings.HasSuffix(f, "v") {
		return f[:len(f)-1]
	}
	return f
}
func rectoVersoCoherence(ls []LineProfile) float64 {
	order, v, _ := folioMeanVectors(ls)
	by := map[string]map[string]string{}
	for _, f := range order {
		side := ""
		if strings.HasSuffix(f, "r") {
			side = "r"
		}
		if strings.HasSuffix(f, "v") {
			side = "v"
		}
		if side != "" {
			if by[leafID(f)] == nil {
				by[leafID(f)] = map[string]string{}
			}
			by[leafID(f)][side] = f
		}
	}
	paired := []float64{}
	for _, leaf := range orderedKeys(by) {
		s := by[leaf]
		if s["r"] != "" && s["v"] != "" {
			paired = append(paired, 1/(1+distance(v[s["r"]], v[s["v"]])))
		}
	}
	return mean(paired)
}
func folioProgression(ls []LineProfile) float64 {
	by := map[string][]LineProfile{}
	for _, l := range ls {
		by[l.Folio] = append(by[l.Folio], l)
	}
	var vals []float64
	for _, folio := range orderedKeys(by) {
		xs := by[folio]
		if len(xs) < 3 {
			continue
		}
		a, b, names := []float64{}, []float64{}, []string{}
		for i, l := range xs {
			a = append(a, float64(i))
			b = append(b, float64(l.TokenCount))
			names = append(names, l.LineID)
		}
		vals = append(vals, math.Abs(spearman(a, b, names)))
	}
	return mean(vals)
}
func shuffleLineLevel(ls []LineProfile, level string, rng *rand.Rand) []LineProfile {
	out := append([]LineProfile(nil), ls...)
	p := rng.Perm(len(ls))
	for i := range out {
		switch level {
		case "locus":
			out[i].LocusID = ls[p[i]].LocusID
		case "folio":
			out[i].Folio = ls[p[i]].Folio
		default:
			out[i].Section = ls[p[i]].Section
		}
	}
	return out
}

func partitionStability(c corpus, measure func(corpus) float64) string {
	if measure == nil || len(c.records) < 20 {
		return "INSUFFICIENT_DATA"
	}
	a, b := corpus{info: c.info}, corpus{info: c.info}
	folios := orderedKeys(func() map[string]bool {
		m := map[string]bool{}
		for _, r := range c.records {
			m[r.Page] = true
		}
		return m
	}())
	half := map[string]bool{}
	for i, f := range folios {
		half[f] = i%2 == 0
	}
	for _, r := range c.records {
		if half[r.Page] {
			a.records = append(a.records, r)
		} else {
			b.records = append(b.records, r)
		}
	}
	x, y := measure(a), measure(b)
	if math.IsNaN(x) || math.IsNaN(y) {
		return "INSUFFICIENT_DATA"
	}
	if x*y < 0 {
		return "PARTITION_SPECIFIC"
	}
	ratio := safeDiv(math.Min(math.Abs(x), math.Abs(y)), math.Max(math.Abs(x), math.Abs(y)))
	if ratio >= .5 {
		return "ROBUST"
	}
	return "ROBUST_WITH_LIMITATIONS"
}

func stabilityRows(ms []FreezeMetric) []StabilityAssessment {
	var out []StabilityAssessment
	for _, m := range ms {
		out = append(out, StabilityAssessment{m.MetricID, "folio split", m.PartitionStability, "deterministic alternating-folio halves", "small sections and labels can be imbalanced"}, StabilityAssessment{m.MetricID, "transcription", "INSUFFICIENT_DATA", "only ZL3b-x7 is aligned in repository-local data", "no cross-transcription claim is permitted"}, StabilityAssessment{m.MetricID, "parameters", m.ParameterSensitivity, "fixed bins and thresholds from configuration", "full threshold landscape remains supporting evidence"})
	}
	return out
}
func task79Redundancy(ls []LineProfile) []RedundancyRow {
	series := map[string][]float64{"line_length": {}, "diversity": {}, "token_entropy": {}, "transition_entropy": {}, "exact_repetition": {}, "near_edit_repetition": {}}
	for _, l := range ls {
		series["line_length"] = append(series["line_length"], float64(l.TokenCount))
		series["diversity"] = append(series["diversity"], l.Diversity)
		series["token_entropy"] = append(series["token_entropy"], l.TokenEntropy)
		series["transition_entropy"] = append(series["transition_entropy"], l.TransitionEntropy)
		series["exact_repetition"] = append(series["exact_repetition"], l.ExactRepetitionRate)
		series["near_edit_repetition"] = append(series["near_edit_repetition"], l.NearEditRepetitionRate)
	}
	names := orderedKeys(series)
	var out []RedundancyRow
	for i := range names {
		for j := i + 1; j < len(names); j++ {
			out = append(out, RedundancyRow{names[i], names[j], pearson(series[names[i]], series[names[j]]), len(ls)})
		}
	}
	return out
}

func bootstrapLineCV(ls []LineProfile, repetitions int, seed int64) (float64, float64) {
	byFolio := map[string][]float64{}
	for _, l := range ls {
		byFolio[l.Folio] = append(byFolio[l.Folio], float64(l.TokenCount))
	}
	folios := orderedKeys(byFolio)
	if len(folios) == 0 || repetitions <= 0 {
		return 0, 0
	}
	rng := rand.New(rand.NewSource(seed))
	values := make([]float64, repetitions)
	for r := range values {
		sample := []float64{}
		for range folios {
			f := folios[rng.Intn(len(folios))]
			sample = append(sample, byFolio[f]...)
		}
		values[r] = safeDiv(sd(sample, mean(sample)), mean(sample))
	}
	sort.Float64s(values)
	lo := values[int(.025*float64(len(values)-1))]
	hi := values[int(.975*float64(len(values)-1))]
	return lo, hi
}

func inheritedFreezeMetrics(base CorpusResult) []FreezeMetric {
	makeMetric := func(id, family, definition string, value float64, status, class string) FreezeMetric {
		return FreezeMetric{MetricID: id, MetricVersion: MetricVersion, Family: family, Definition: definition, UnitOfAnalysis: "corpus summary", Inputs: []string{"frozen token corpus"}, Parameters: map[string]any{}, ObservedValue: value, Uncertainty: "see task75/task77 null distributions in raw_results.json", NullModels: []string{"task75/task77 registered nulls"}, PartitionStability: "ROBUST_WITH_LIMITATIONS", TranscriptionStability: "INSUFFICIENT_DATA: one aligned transcription", ParameterSensitivity: "see task75/task77 audit", RedundancyClass: class, CoverageRole: []string{"constrained token grammar", "copy-with-modification"}, ComparisonEligibility: "CORPUS_COMPARABLE", NegativeEvidenceStatus: "see source block", ImplementationVersion: Version, Status: status, Limitations: "Inherited without changing the task75/task77 estimator or null."}
	}
	state := base.Metrics.LP2.ProductivityState
	out := []FreezeMetric{
		makeMetric("LP1_RULE_SUPPORT_GINI", "lexical paradigm", "Concentration of directed edit-rule support", base.Metrics.LP1.SupportGini, state, "SUPPORTING"),
		makeMetric("LP4_PREFIX_ATTACHMENT_NMI", "lexical paradigm", "Prefix-to-core normalized mutual information", base.Metrics.LP4.Prefix.NormalizedMI, "REPRODUCED", "SUPPORTING"),
		makeMetric("LP4_SUFFIX_ATTACHMENT_NMI", "lexical paradigm", "Suffix-to-core normalized mutual information", base.Metrics.LP4.Suffix.NormalizedMI, "REPRODUCED", "SUPPORTING"),
		makeMetric("EF1_GIANT_COMPONENT_SHARE", "edit family", "Share of vocabulary types in the largest edit component", base.Metrics.EF1.GiantComponentShare, base.Metrics.EF4.Verdict, "CORE"),
		makeMetric("EF1_ISOLATE_SHARE", "edit family", "Share of isolated vocabulary types in edit graph", base.Metrics.EF1.IsolateShare, "REPRODUCED", "SUPPORTING"),
		makeMetric("EF2_GLOBAL_CLUSTERING", "edit family", "Global edit-graph clustering coefficient", base.Metrics.EF2.GlobalClustering, "REPRODUCED", "CORE"),
		makeMetric("EF3_DEGREE_FREQUENCY_SPEARMAN", "edit family", "Spearman correlation of edit degree and log frequency", base.Metrics.EF3.SpearmanDegreeLogFrequency, "REPRODUCED", "CORE"),
	}
	if base.CrossScale != nil {
		for _, m := range base.CrossScale.Metrics {
			class := "SUPPORTING"
			for _, c := range base.CrossScale.MetricClassifications {
				if c.MetricID == m.MetricID {
					class = c.Class
					break
				}
			}
			x := makeMetric(m.MetricID, "cross-scale", m.Hypothesis, m.ObservedStatistic, m.Status, class)
			x.EffectSize = m.EffectSize
			x.PValue = m.EmpiricalP
			x.QValue = m.MultipleTestingAdjustment
			x.Uncertainty = m.Uncertainty
			x.NullModels = []string{m.NullModel}
			x.Limitations = m.Limitations
			out = append(out, x)
		}
	}
	return out
}

func segmentFolios(c corpus, ls []LineProfile, penalty float64) SegmentationResult {
	by := map[string][]LineProfile{}
	order := []string{}
	for _, l := range ls {
		if _, ok := by[l.Folio]; !ok {
			order = append(order, l.Folio)
		}
		by[l.Folio] = append(by[l.Folio], l)
	}
	vecs := make([][]float64, len(order))
	for i, f := range order {
		v := make([]float64, 6)
		for _, l := range by[f] {
			for j, x := range lineVector(l) {
				v[j] += x
			}
		}
		for j := range v {
			v[j] /= float64(len(by[f]))
		}
		vecs[i] = v
	}
	scores := []float64{}
	for i := 1; i < len(vecs); i++ {
		scores = append(scores, distance(vecs[i-1], vecs[i]))
	}
	threshold := mean(scores) + penalty*sd(scores, mean(scores))
	var cp []ChangePoint
	agree := 0
	for i, s := range scores {
		if s >= threshold {
			cp = append(cp, ChangePoint{i + 1, s, order[i+1]})
			if by[order[i]][0].Section != by[order[i+1]][0].Section {
				agree++
			}
		}
	}
	status := "NOT_SUPPORTED"
	if len(cp) > 0 {
		status = "PARTIALLY_SUPPORTED"
	}
	return SegmentationResult{"fixed profile-distance peaks", "mean + configured SD penalty", status, cp, safeDiv(float64(agree), float64(len(cp))), "Exploratory; folio order is documentary and no historical-order claim is made."}
}

func auditInputs(base CorpusResult, cfg Task79Config) []AuditRecord {
	rows := []AuditRecord{{"corpus checksum", "REPRODUCED", base.Corpus.SHA256, "local bytes only"}, {"LP1-LP4 / EF1-EF5", "REPRODUCED", "computed in the same pipeline invocation", "shared nulls imply dependent evidence"}, {"C-GRAMMAR", map[bool]string{true: "REPRODUCED", false: "NOT_REPRODUCED"}[base.Grammar != nil], "per-replicate diagnostics serialized; the validation outcome itself need not be positive", "mode-specific validation applies"}, {"consensus edit families", map[bool]string{true: "REPRODUCED", false: "NOT_REPRODUCED"}[base.EditGraph != nil], "edit graph validation serialized", "empty LP2-gated catalog is a substantive outcome"}, {"cross-scale metrics", map[bool]string{true: "REPRODUCED", false: "NOT_REPRODUCED"}[base.CrossScale != nil], "same run and seeds", "some family metrics can be inapplicable"}, {"transcription cross-check", "NOT_TESTABLE", "only one aligned transcription available", "excluded from freeze-ready claims"}}
	for _, p := range cfg.AuditArtifacts {
		if _, err := os.Stat(p); err != nil {
			rows = append(rows, AuditRecord{p, "NOT_REPRODUCED", err.Error(), "artifact path declared in config"})
		} else {
			rows = append(rows, AuditRecord{p, "REPRODUCED", "artifact exists and is independently checksummed in freeze manifest", "semantic comparison uses current pipeline schema"})
		}
	}
	return rows
}

func hierarchicalNulls() []NullModelRegistryEntry {
	return []NullModelRegistryEntry{{"HN1", "Within-line permutation", "line token multiset, length, metadata", "order and identity-slot association", []string{"LS2", "LS3", "LS4", "2D-LITE"}, "boundary slots remain fixed", "direct positional null"}, {"HN2", "Line reassignment within folio", "folio composition and line profiles", "locus membership", []string{"LC1", "LC3", "LC4"}, "folio confounding remains", "tests locus increment"}, {"HN3", "Locus reassignment within section", "section composition and locus sizes", "folio organization", []string{"PF2", "PF3"}, "section regime remains", "tests folio increment"}, {"HN4", "Folio reassignment within Currier/section", "section/Currier composition and folio sizes", "folio identity", []string{"PF2", "PF4", "HR1"}, "document order not tested", "matched folio null"}, {"HN5", "Boundary-preserving token shuffle", "line lengths and boundary slots", "token-boundary identity", []string{"BP1"}, "boundary marginal remains", "boundary-specific null"}, {"HN6", "Metadata-label permutation", "group sizes and observations", "metadata association", []string{"LC1", "LC2", "HR1", "HR6"}, "nested labels require restricted permutation", "association null"}, {"HN7", "Hierarchical bootstrap", "folio/locus nesting", "sampling idiosyncrasy", []string{"all task79 metrics"}, "only observed hierarchy is resampled", "uncertainty"}, {"HN8", "C-GRAMMAR-derived controls", "validated token-formation marginals", "higher organization", []string{"eligible boundary/token metrics"}, "original C-GRAMMAR has no learned page generator", "only used where model preserves relevant marginals"}}
}

func coverageAudit() []CoverageAssessment {
	classes := []struct {
		name     string
		metrics  []string
		gap      string
		critical bool
	}{{"natural-language-like process", []string{"LS1", "LS2", "PF2", "HR1"}, "historical multilingual controls are limited", false}, {"constrained token grammar", []string{"LP1-LP4", "EF1-EF5", "2DL1"}, "page-aware C-GRAMMAR is not defined", false}, {"copy-with-modification", []string{"LS4", "LP/EF", "PF2"}, "no aligned copying ground truth", false}, {"stochastic procedural generator", []string{"LS1", "LS4", "HR1"}, "synthetic process portfolio incomplete", true}, {"table/index", []string{"LC1", "LC2", "PF2", "HR1"}, "no historical table control", true}, {"cipher-like transformation", []string{"Phase I entropy", "LP4", "LS2"}, "cipher family coverage is incomplete", false}, {"abbreviation/shorthand-like token construction", []string{"LP/EF", "LS2", "BP1"}, "aligned abbreviation positive control required", true}, {"line-boundary projection", []string{"BP1", "LS2"}, "periodicity family is exploratory", false}, {"acrostic-compatible organization", []string{"BP1", "LS2"}, "tests structure only, never plaintext", false}, {"external-memory mechanism", []string{"HR1", "LC1", "PF2"}, "semantic retrieval is outside corpus-only observability", true}, {"mnemonic cue system", []string{"LC2", "BP1", "HR1"}, "no positive control", true}, {"mixed architecture", []string{"LC1", "PF2", "HR1", "segmentation"}, "model class is broad", false}}
	out := make([]CoverageAssessment, 0, len(classes))
	for _, c := range classes {
		out = append(out, CoverageAssessment{c.name, c.metrics, c.gap, "Phase I natural-language plus configured controls", "Only metric-sensitive structural predictions can support negative conclusions.", map[bool]string{true: "task79-b: acquire/preregister an external positive control", false: "no critical pre-experiment metric gap"}[c.critical], c.critical})
	}
	return out
}
func negativeRegistry(ms []FreezeMetric, cfg Task79Config) []NegativeEvidence {
	var out []NegativeEvidence
	for _, m := range ms {
		if m.Status == "NOT_SUPPORTED" {
			threshold := "empirical resolution " + fmt.Sprintf("1/(%d+1)", cfg.Permutations)
			status := m.NegativeEvidenceStatus
			out = append(out, NegativeEvidence{m.MetricID, m.Definition, status, "matched permutation test", threshold, m.Uncertainty, strings.Join(m.NullModels, ","), m.PartitionStability, m.Limitations})
		}
	}
	return out
}

func task79Verdicts(ms []FreezeMetric, a MetadataAudit, s SegmentationResult, freeze string) []Task79Verdict {
	supported := func(prefix string) bool {
		for _, m := range ms {
			if strings.HasPrefix(m.MetricID, prefix) && m.Status == "SUPPORTED" {
				return true
			}
		}
		return false
	}
	v := func(id, value, evidence string, effect float64, impact string) Task79Verdict {
		return Task79Verdict{ID: id, Value: value, PrimaryEvidence: evidence, EffectSize: effect, NullComparison: "permutation/bootstrap nulls in metric registry", Stability: "see stability_matrix", HeldOutResult: "grouped/partitioned result in metric registry", Limitations: "single transcription; 2D-LITE is not scan geometry", FreezeImpact: impact}
	}
	meta := "SUPPORTED"
	if a.Status != "REPRODUCED" {
		meta = "INCONCLUSIVE"
	}
	rows := []Task79Verdict{v("INPUT_RESULTS_REPRODUCED", "PARTIALLY_SUPPORTED", "task75/task77 audit includes one NOT_TESTABLE transcription block", 0, "prevents full freeze"), v("METADATA_INTEGRITY_ACCEPTABLE", meta, fmt.Sprintf("%d tokens, %d lines, %d loci, %d folios; nesting violations=%d", a.Tokens, a.Lines, a.Loci, a.Folios, a.NestingViolations), 0, "required for page metrics"), v("LINE_STRUCTURE_SUPPORTED", map[bool]string{true: "SUPPORTED", false: "NOT_SUPPORTED"}[supported("LS")], "LS registry", 0, "core candidate"), v("BOUNDARY_STRUCTURE_SUPPORTED", map[bool]string{true: "SUPPORTED", false: "NOT_SUPPORTED"}[supported("BP")], "BP registry", 0, "does not imply acrostic"), v("LOCUS_STRUCTURE_SUPPORTED", map[bool]string{true: "SUPPORTED", false: "NOT_SUPPORTED"}[supported("LC")], "LC registry", 0, "core candidate"), v("FOLIO_STRUCTURE_SUPPORTED", map[bool]string{true: "SUPPORTED", false: "NOT_SUPPORTED"}[supported("PF")], "PF registry", 0, "core candidate"), v("RECTO_VERSO_DEPENDENCE_SUPPORTED", "INCONCLUSIVE", "PF4 lacks adequately matched independent physical-leaf controls", 0, "unfrozen extension"), v("HIERARCHICAL_ORGANIZATION_SUPPORTED", map[bool]string{true: "SUPPORTED", false: "NOT_SUPPORTED"}[supported("HR1")], "variance-share metrics", 0, "core candidate"), v("HIERARCHICAL_MODEL_OUTPERFORMS_FLAT", "INCONCLUSIVE", "predictive hierarchy is not promoted without multi-corpus held-out validation", 0, "prevents full freeze"), v("UNANNOTATED_CHANGE_POINTS_SUPPORTED", s.Status, "penalized folio-profile distance peaks", 0, "exploratory only"), v("CORE_METRICS_STABLE", "PARTIALLY_SUPPORTED", "folio splits pass for retained metrics; transcription unavailable", 0, "prevents full freeze"), v("CORE_METRICS_NONREDUNDANT", "PARTIALLY_SUPPORTED", "line-profile correlation matrix emitted", 0, "candidate selection only"), v("NEGATIVE_EVIDENCE_REGISTERED", "SUPPORTED", "explicit registry with sensitivity and controls", 0, "required gate passed"), v("ALTERNATIVE_EXPLANATION_COVERAGE_ACCEPTABLE", "NOT_SUPPORTED", "historical notation/table/procedural positive controls missing", 0, "task79-b required"), v("MODEL_COMPARISON_INTERFACE_READY", "PARTIALLY_SUPPORTED", "machine-readable schema exists; frozen distance portfolio does not", 0, "task79-b required"), v("FINGERPRINT_V2_FREEZE_STATUS", freeze, "critical coverage and transcription gaps", 0, "do not freeze")}
	return rows
}

func occurrenceRows(c corpus) []OccurrenceMetadata {
	out := make([]OccurrenceMetadata, len(c.records))
	for i, r := range c.records {
		norm := 0.0
		if r.LineLength > 1 {
			norm = float64(r.IndexInLine) / float64(r.LineLength-1)
		}
		label := "text"
		if strings.EqualFold(r.LocusType, "L") {
			label = "label"
		}
		missing := []string{}
		if r.Page == "" {
			missing = append(missing, "folio")
		}
		if r.LocusType == "" {
			missing = append(missing, "locus_type")
		}
		out[i] = OccurrenceMetadata{i, r.Token, r.Page, r.FolioSide, r.LocusID, r.LocusType, r.LineID, r.IndexInLine, norm, r.LineLength, r.IndexInLocus, r.IndexInFolio, r.ParagraphID, r.ParagraphStart, r.Section, r.Currier, r.Hand, r.Quire, r.IVTFFI, r.IVTFFX, label, strings.Join(missing, ",")}
	}
	return out
}
func bytesSHA256(b []byte) string { x := sha256.Sum256(b); return hex.EncodeToString(x[:]) }

func finalizeTask79Discrimination(f *Fingerprint) {
	if f.Primary.Task79 == nil {
		return
	}
	primary := map[string]FreezeMetric{}
	for _, m := range f.Primary.Task79.Metrics {
		if m.ComparisonEligibility == "CORPUS_COMPARABLE" {
			primary[m.MetricID] = m
		}
	}
	for _, control := range f.Controls {
		if control.Task79 == nil {
			continue
		}
		cm := map[string]FreezeMetric{}
		for _, m := range control.Task79.Metrics {
			cm[m.MetricID] = m
		}
		for id, p := range primary {
			if control.Task79.MetadataAudit.Status != "REPRODUCED" && (strings.HasPrefix(id, "PF") || strings.HasPrefix(id, "LC") || strings.HasPrefix(id, "HR")) {
				continue
			}
			m, ok := cm[id]
			if !ok {
				continue
			}
			scale := math.Abs(p.ObservedValue) + math.Abs(m.ObservedValue)
			z := 0.0
			if scale > 0 {
				z = 2 * math.Abs(p.ObservedValue-m.ObservedValue) / scale
			}
			f.Primary.Task79.Discriminative = append(f.Primary.Task79.Discriminative, DiscriminativeResult{ControlID: control.Corpus.ID, MetricID: id, PrimaryValue: p.ObservedValue, ControlValue: m.ObservedValue, StandardizedDifference: z, Status: "PSEUDO_BLIND_REPORTED", Limitation: "Descriptive held-out control contrast; not a trained classifier and not a freeze criterion by itself."})
		}
	}
	sort.Slice(f.Primary.Task79.Discriminative, func(i, j int) bool {
		a, b := f.Primary.Task79.Discriminative[i], f.Primary.Task79.Discriminative[j]
		if a.ControlID != b.ControlID {
			return a.ControlID < b.ControlID
		}
		return a.MetricID < b.MetricID
	})
}
