package fingerprintv2

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"

	"zcore.dev/voinich/internal/tokenrepetition"
)

// ---- shared labels (task77 stage 4 variables) ----

func familyLabelOf(familyOf map[string]int, token string) string {
	if id, ok := familyOf[token]; ok {
		return "F" + strconv.Itoa(id)
	}
	return "NONE"
}

func roleLabelOf(roleOf map[string]string, token string) string {
	if r, ok := roleOf[token]; ok {
		return r
	}
	return "NONE"
}

func positionClass(r tokenRecord) (string, bool) {
	if r.LineLength < 2 {
		return "", false
	}
	switch {
	case r.IndexInLine == 0:
		return "INITIAL", true
	case r.IndexInLine == r.LineLength-1:
		return "FINAL", true
	default:
		return "INTERIOR", true
	}
}

func normalizedPositionOf(r tokenRecord) (float64, bool) {
	if r.LineLength < 2 {
		return 0, false
	}
	return float64(r.IndexInLine) / float64(r.LineLength-1), true
}

func locusClass(r tokenRecord) (string, bool) {
	if r.LocusType == "" {
		return "", false
	}
	switch r.LocusType {
	case "P":
		return "TEXT", true
	case "L":
		return "LABEL", true
	default:
		return "SPECIAL", true
	}
}

func regimeLabel(r tokenRecord) (string, bool) {
	if r.Currier == "" || r.Section == "" {
		return "", false
	}
	return r.Currier + "|" + r.Section, true
}

// assignFamiliesAndRoles labels every family-graph member with its family
// index and a CORE/PERIPHERY role from k-core decomposition (coreness>=2 is
// CORE, matching familyStructuralDiagnostics' core/periphery split).
func assignFamiliesAndRoles(g editGraph, families [][]string) (map[string]int, map[string]string) {
	familyOf, roleOf := map[string]int{}, map[string]string{}
	for i, fam := range families {
		core := kCoreDecomposition(g, fam)
		for _, tok := range fam {
			familyOf[tok] = i
			if core[tok] >= 2 {
				roleOf[tok] = "CORE"
			} else {
				roleOf[tok] = "PERIPHERY"
			}
		}
	}
	return familyOf, roleOf
}

// zoneProfileLabels reports, for every family-graph node, the sorted set of
// edit-operation zones (PREFIX/SUFFIX/INTERNAL) it participates in.
func zoneProfileLabels(g editGraph, glyphs map[string][]string) map[string]string {
	zones := map[string]map[string]bool{}
	add := func(tok, zone string) {
		if zones[tok] == nil {
			zones[tok] = map[string]bool{}
		}
		zones[tok][zone] = true
	}
	for _, e := range g.edgeList() {
		rule, ok := ruleFor(e[0], e[1], glyphs)
		if !ok {
			continue
		}
		parts := strings.Split(rule, "|")
		if len(parts) < 2 {
			continue
		}
		add(e[0], parts[1])
		add(e[1], parts[1])
	}
	out := map[string]string{}
	for tok, set := range zones {
		out[tok] = strings.Join(orderedKeys(set), ",")
	}
	return out
}

func zoneLabelOf(zones map[string]string, tok string) string {
	if z, ok := zones[tok]; ok {
		return z
	}
	return "NONE"
}

func lineSequences(c corpus) map[int][]int {
	out := map[int][]int{}
	for i, r := range c.records {
		out[r.Line] = append(out[r.Line], i)
	}
	return out
}

func sortedIntKeys(m map[int][]int) []int {
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}

// ---- stage 4: variable registry ----

func crossScaleVariables(c corpus) []CrossScaleVariable {
	hasPage, hasLocus, hasRegime, hasLine := false, false, false, false
	for _, r := range c.records {
		if r.Page != "" {
			hasPage = true
		}
		if r.LocusType != "" {
			hasLocus = true
		}
		if r.Currier != "" && r.Section != "" {
			hasRegime = true
		}
		if r.LineLength > 0 {
			hasLine = true
		}
	}
	v := func(scale, name, origin, domain, missing string, available bool) CrossScaleVariable {
		return CrossScaleVariable{Scale: scale, Name: name, Origin: origin, Domain: domain, MissingPolicy: missing, Available: available}
	}
	return []CrossScaleVariable{
		v("token", "token_identity", "corpus token stream", "vocabulary types", "none (always present)", true),
		v("token", "length", "glyph count", "positive integers", "none", true),
		v("token", "frequency", "corpus token counts", "positive integers", "none", true),
		v("token", "edit_family", "EF1/LP2-productive connected component", "family index or NONE", "NONE for isolates/small components/non-productive tokens", true),
		v("token", "family_role", "k-core decomposition of the family graph", "{CORE, PERIPHERY, NONE}", "NONE for tokens outside a family", true),
		v("local_sequence", "previous_token", "corpus order", "vocabulary types", "none within-line/corpus start has no previous token and is excluded", true),
		v("local_sequence", "transformation_zone_profile", "LP1 rule zone classification on family-graph edges", "subset of {PREFIX, SUFFIX, INTERNAL} or NONE", "NONE for tokens outside a family", true),
		v("line", "index_in_line", "IVTFF strict alignment", "0..line_length-1", "unavailable without ivtff_path; corpus-wide NOT_APPLICABLE then", hasLine),
		v("line", "line_length", "IVTFF strict alignment", "positive integers", "unavailable without ivtff_path", hasLine),
		v("line", "position_class", "index_in_line vs line_length", "{INITIAL, INTERIOR, FINAL}", "lines of length 1 excluded (ambiguous)", hasLine),
		v("locus", "locus_type", "IVTFF locus-type code", "{TEXT, LABEL, SPECIAL}", "unavailable without ivtff_path", hasLocus),
		v("locus", "paragraph_start", "IVTFF paragraph marker", "boolean", "unavailable without ivtff_path", hasLocus),
		v("folio", "folio_id", "IVTFF folio identifier", "folio strings", "unavailable without ivtff_path (falls back to no page metadata)", hasPage),
		v("folio", "folio_side", "trailing recto/verso letter of the folio id", "{r, v, unknown}", "unknown if the folio id has no r/v suffix", hasPage),
		v("corpus", "currier", "IVTFF $C page variable", "language labels", "excluded from regime tests if empty", hasRegime),
		v("corpus", "section", "IVTFF $I page variable (illustration-type code)", "section codes", "excluded from regime tests if empty", hasRegime),
	}
}

// ---- cross-scale metric construction helper ----

func csMetric(id, hypothesis, unit string, variables, conditioning, confounders []string, t NullTest, n int, interpretation, limitations, sensitivity string) CrossScaleMetric {
	return CrossScaleMetric{
		MetricID: id, MetricVersion: MetricVersion, Hypothesis: hypothesis, UnitOfAnalysis: unit,
		Variables: variables, ConditioningVariables: conditioning, Confounders: confounders,
		ObservedStatistic: t.Observed, EffectSize: t.EffectSize, EffectDefined: t.EffectDefined,
		Uncertainty: fmt.Sprintf("empirical permutation null, %d replicates", t.Replicates),
		NullModel:   t.NullModel, NullMean: t.NullMean, NullSD: t.NullSD, EmpiricalP: t.PValue,
		N: n, Interpretation: interpretation, Limitations: limitations, Sensitivity: sensitivity,
		AnalysisType: "CONFIRMATORY",
	}
}

func csNotApplicable(id, hypothesis, reason string) CrossScaleMetric {
	return CrossScaleMetric{
		MetricID: id, MetricVersion: MetricVersion, Status: "NOT_APPLICABLE", Hypothesis: hypothesis,
		Limitations: reason, AnalysisType: "CONFIRMATORY",
	}
}

func csInconclusive(id, hypothesis, reason string, n int) CrossScaleMetric {
	return CrossScaleMetric{
		MetricID: id, MetricVersion: MetricVersion, Status: "INCONCLUSIVE", Hypothesis: hypothesis,
		Limitations: reason, N: n, AnalysisType: "CONFIRMATORY",
	}
}

// ---- CS1: family x line position ----

func cs1Test(c corpus, familyOf map[string]int, roleOf map[string]string, repetitions int, rng *rand.Rand) (famTest, roleTest NullTest, n int, diag map[string]float64) {
	var idxs []int
	for i, r := range c.records {
		if _, ok := familyOf[r.Token]; !ok {
			continue
		}
		if r.LineLength < 2 {
			continue
		}
		idxs = append(idxs, i)
	}
	x := make([]string, len(idxs))
	yFam := make([]string, len(idxs))
	yRole := make([]string, len(idxs))
	group := make([]string, len(idxs))
	for j, i := range idxs {
		r := c.records[i]
		pc, _ := positionClass(r)
		x[j] = pc
		yFam[j] = familyLabelOf(familyOf, r.Token)
		yRole[j] = roleLabelOf(roleOf, r.Token)
		group[j] = r.LineID
	}
	famTest = nmiPermutationTest("cs1/family-line-position", "N2 within-line shuffle restricted to family-bearing occurrences", x, yFam, group, repetitions, rng)
	roleTest = nmiPermutationTest("cs1/role-line-position", "N2 within-line shuffle restricted to family-bearing occurrences", x, yRole, group, repetitions, rng)
	sums := map[string][2]float64{}
	for _, i := range idxs {
		r := c.records[i]
		np, _ := normalizedPositionOf(r)
		lbl := familyLabelOf(familyOf, r.Token)
		s := sums[lbl]
		s[0] += np
		s[1]++
		sums[lbl] = s
	}
	diag = map[string]float64{}
	for lbl, s := range sums {
		if s[1] > 0 {
			diag[lbl] = s[0] / s[1]
		}
	}
	return famTest, roleTest, len(idxs), diag
}

// cs1FamilyLabelPermutationNull is N8 applied to CS1: it asks whether
// which specific tokens carry which family label matters, holding the
// family-size distribution (and each token's own frequency) fixed, as a
// robustness check on top of the primary within-line positional null.
func cs1FamilyLabelPermutationNull(c corpus, familyOf map[string]int, repetitions int, rng *rand.Rand) NullTest {
	var idxs []int
	for i, r := range c.records {
		if _, ok := familyOf[r.Token]; !ok {
			continue
		}
		if r.LineLength < 2 {
			continue
		}
		idxs = append(idxs, i)
	}
	x := make([]string, len(idxs))
	y := make([]string, len(idxs))
	for j, i := range idxs {
		pc, _ := positionClass(c.records[i])
		x[j] = pc
		y[j] = familyLabelOf(familyOf, c.records[i].Token)
	}
	observed := normalizedMI(x, y)
	null := make([]float64, repetitions)
	for r := range null {
		permuted := familyLabelPermutation(familyOf, rng)
		yy := make([]string, len(idxs))
		for j, i := range idxs {
			yy[j] = familyLabelOf(permuted, c.records[i].Token)
		}
		null[r] = normalizedMI(x, yy)
	}
	return nullTest("cs1/family-line-position/n8", "N8 family-label permutation (fixed family sizes and token frequencies)", observed, null)
}

// ---- CS2: transformation x local context ----

func cs2Test(c corpus, familyOf map[string]int, zoneOf map[string]string, repetitions int, rng *rand.Rand) (famAdjTest, zoneTest NullTest, n int) {
	lines := lineSequences(c)
	lineIDs := sortedIntKeys(lines)
	observe := func(seqOf func(lineID int) []int) (float64, float64, int) {
		var pf, cf, cz []string
		for _, lid := range lineIDs {
			seq := seqOf(lid)
			for pos := 1; pos < len(seq); pos++ {
				prevTok := c.records[seq[pos-1]].Token
				curTok := c.records[seq[pos]].Token
				curLbl := familyLabelOf(familyOf, curTok)
				if curLbl == "NONE" {
					continue
				}
				pf = append(pf, familyLabelOf(familyOf, prevTok))
				cf = append(cf, curLbl)
				cz = append(cz, zoneLabelOf(zoneOf, curTok))
			}
		}
		return normalizedMI(pf, cf), normalizedMI(pf, cz), len(cf)
	}
	observedFam, observedZone, n := observe(func(lid int) []int { return lines[lid] })
	nullFam, nullZone := make([]float64, repetitions), make([]float64, repetitions)
	for r := 0; r < repetitions; r++ {
		shuffled := map[int][]int{}
		for lid, seq := range lines {
			perm := rng.Perm(len(seq))
			out := make([]int, len(seq))
			for i, p := range perm {
				out[i] = seq[p]
			}
			shuffled[lid] = out
		}
		nullFam[r], nullZone[r], _ = observe(func(lid int) []int { return shuffled[lid] })
	}
	famAdjTest = nullTest("cs2/prev-family-current-family", "N2 within-line sequence shuffle", observedFam, nullFam)
	zoneTest = nullTest("cs2/prev-family-current-zone", "N2 within-line sequence shuffle", observedZone, nullZone)
	return famAdjTest, zoneTest, n
}

// ---- CS3: family x locus type ----

func cs3Test(c corpus, familyOf map[string]int, repetitions int, rng *rand.Rand) (NullTest, int, bool) {
	var x, y, group []string
	for _, r := range c.records {
		lc, ok := locusClass(r)
		if !ok {
			continue
		}
		x = append(x, familyLabelOf(familyOf, r.Token))
		y = append(y, lc)
		group = append(group, r.Page)
	}
	if len(x) < 20 {
		return NullTest{}, len(x), false
	}
	return nmiPermutationTest("cs3/family-locus-type", "N4 within-folio shuffle", x, y, group, repetitions, rng), len(x), true
}

// ---- CS4: family x folio/section/Currier regime ----

func folioLevelLabels(c corpus, get func(tokenRecord) (string, bool)) map[string]string {
	counts := map[string]map[string]int{}
	for _, r := range c.records {
		if r.Page == "" {
			continue
		}
		v, ok := get(r)
		if !ok {
			continue
		}
		if counts[r.Page] == nil {
			counts[r.Page] = map[string]int{}
		}
		counts[r.Page][v]++
	}
	out := map[string]string{}
	for _, folio := range orderedKeys(counts) {
		best, bestN := "", -1
		for _, k := range orderedKeys(counts[folio]) {
			if counts[folio][k] > bestN {
				best, bestN = k, counts[folio][k]
			}
		}
		out[folio] = best
	}
	return out
}

func permuteFolioLabels(folioLabel map[string]string, rng *rand.Rand) map[string]string {
	folios := orderedKeys(folioLabel)
	labels := make([]string, len(folios))
	for i, f := range folios {
		labels[i] = folioLabel[f]
	}
	perm := rng.Perm(len(labels))
	out := map[string]string{}
	for i, f := range folios {
		out[f] = labels[perm[i]]
	}
	return out
}

func cs4MetadataTest(c corpus, familyOf map[string]int, get func(tokenRecord) (string, bool), id, label string, repetitions int, rng *rand.Rand) (NullTest, int, bool) {
	folioLabel := folioLevelLabels(c, get)
	if len(folioLabel) < 4 {
		return NullTest{}, 0, false
	}
	var x, y []string
	for _, r := range c.records {
		if r.Page == "" {
			continue
		}
		v, ok := folioLabel[r.Page]
		if !ok {
			continue
		}
		if _, famOK := familyOf[r.Token]; !famOK {
			continue
		}
		x = append(x, familyLabelOf(familyOf, r.Token))
		y = append(y, v)
	}
	if len(x) < 20 {
		return NullTest{}, len(x), false
	}
	observed := normalizedMI(x, y)
	null := make([]float64, repetitions)
	for r := range null {
		shuffledFolio := permuteFolioLabels(folioLabel, rng)
		yy := make([]string, 0, len(y))
		for _, rec := range c.records {
			if rec.Page == "" {
				continue
			}
			if _, ok := folioLabel[rec.Page]; !ok {
				continue
			}
			if _, famOK := familyOf[rec.Token]; !famOK {
				continue
			}
			yy = append(yy, shuffledFolio[rec.Page])
		}
		null[r] = normalizedMI(x, yy)
	}
	return nullTest(id, label, observed, null), len(x), true
}

// ---- CS5: local context x larger regime (interaction) ----

func cs5Test(c corpus, familyOf map[string]int, repetitions int, rng *rand.Rand) (NullTest, int, map[string]float64, bool) {
	lines := lineSequences(c)
	lineIDs := sortedIntKeys(lines)
	type pair struct {
		page  string
		match float64
	}
	var pairs []pair
	for _, lid := range lineIDs {
		seq := lines[lid]
		for pos := 1; pos < len(seq); pos++ {
			prevTok := c.records[seq[pos-1]].Token
			curRec := c.records[seq[pos]]
			curLbl := familyLabelOf(familyOf, curRec.Token)
			if curLbl == "NONE" {
				continue
			}
			match := 0.0
			if familyLabelOf(familyOf, prevTok) == curLbl {
				match = 1
			}
			pairs = append(pairs, pair{page: curRec.Page, match: match})
		}
	}
	folioLabel := folioLevelLabels(c, regimeLabel)
	if len(pairs) < 20 || len(folioLabel) < 4 {
		return NullTest{}, len(pairs), nil, false
	}
	statistic := func(folioToLabel map[string]string) (float64, map[string]float64) {
		sums := map[string][2]float64{}
		for _, p := range pairs {
			lbl, ok := folioToLabel[p.page]
			if !ok {
				continue
			}
			s := sums[lbl]
			s[0] += p.match
			s[1]++
			sums[lbl] = s
		}
		rates := map[string]float64{}
		for k, s := range sums {
			if s[1] >= 5 {
				rates[k] = s[0] / s[1]
			}
		}
		if len(rates) < 2 {
			return 0, rates
		}
		mn, mx := math.Inf(1), math.Inf(-1)
		for _, v := range rates {
			if v < mn {
				mn = v
			}
			if v > mx {
				mx = v
			}
		}
		return mx - mn, rates
	}
	observed, rates := statistic(folioLabel)
	if len(rates) < 2 {
		return NullTest{}, len(pairs), rates, false
	}
	null := make([]float64, repetitions)
	for r := range null {
		null[r], _ = statistic(permuteFolioLabels(folioLabel, rng))
	}
	return nullTest("cs5/local-adjacency-x-regime", "folio-level regime-label permutation", observed, null), len(pairs), rates, true
}

// ---- CS6: family composition x line structure ----

func cs6Test(c corpus, familyOf map[string]int, repetitions int, rng *rand.Rand) (NullTest, int, bool) {
	lines := lineSequences(c)
	lineIDs := sortedIntKeys(lines)
	if len(lineIDs) < 10 {
		return NullTest{}, len(lineIDs), false
	}
	diversity := func(tokenAt func(idx int) string) []float64 {
		out := make([]float64, len(lineIDs))
		for li, lid := range lineIDs {
			counts := map[string]int{}
			for _, idx := range lines[lid] {
				counts[familyLabelOf(familyOf, tokenAt(idx))]++
			}
			out[li] = entropy(counts)
		}
		return out
	}
	lengths := make([]float64, len(lineIDs))
	names := make([]string, len(lineIDs))
	for li, lid := range lineIDs {
		lengths[li] = float64(len(lines[lid]))
		names[li] = strconv.Itoa(lid)
	}
	observedDiv := diversity(func(idx int) string { return c.records[idx].Token })
	observed := math.Abs(spearman(lengths, observedDiv, names))
	null := make([]float64, repetitions)
	for r := range null {
		perm := rng.Perm(len(c.records))
		shuffledToken := make([]string, len(c.records))
		for i, p := range perm {
			shuffledToken[i] = c.records[p].Token
		}
		div := diversity(func(idx int) string { return shuffledToken[idx] })
		null[r] = math.Abs(spearman(lengths, div, names))
	}
	return nullTest("cs6/family-diversity-x-line-length", "N1 global token shuffle (fixed line-length sequence)", observed, null), len(lineIDs), true
}

// ---- CS7: edit distance x structural distance ----

func minLineDistance(a, b []int) int {
	if len(a) == 0 || len(b) == 0 {
		return math.MaxInt32
	}
	sort.Ints(a)
	sort.Ints(b)
	best := math.MaxInt32
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		d := a[i] - b[j]
		if d < 0 {
			d = -d
		}
		if d < best {
			best = d
		}
		if a[i] < b[j] {
			i++
		} else {
			j++
		}
	}
	return best
}

func cs7Test(c corpus, glyphs map[string][]string, freq map[string]int, sampleSize, repetitions int, rng *rand.Rand) (NullTest, int, bool) {
	vocab := vocabulary(c)
	if len(vocab) < 20 {
		return NullTest{}, 0, false
	}
	occLines := map[string][]int{}
	for _, r := range c.records {
		if len(occLines[r.Token]) < 50 {
			occLines[r.Token] = append(occLines[r.Token], r.Line)
		}
	}
	type samplePair struct {
		a, b string
		bin  int
	}
	seen := map[string]bool{}
	pairs := make([]samplePair, 0, sampleSize)
	attempts, maxAttempts := 0, sampleSize*20+1000
	for len(pairs) < sampleSize && attempts < maxAttempts {
		attempts++
		a := vocab[rng.Intn(len(vocab))]
		b := vocab[rng.Intn(len(vocab))]
		if a == b {
			continue
		}
		if a > b {
			a, b = b, a
		}
		key := a + "\x00" + b
		if seen[key] {
			continue
		}
		seen[key] = true
		pairs = append(pairs, samplePair{a, b, frequencyBin(freq[a]) + frequencyBin(freq[b])})
	}
	if len(pairs) < 20 {
		return NullTest{}, len(pairs), false
	}
	editDist := make([]float64, len(pairs))
	structDist := make([]float64, len(pairs))
	byBin := map[int][]int{}
	for i, p := range pairs {
		editDist[i] = float64(tokenrepetition.LevenshteinGlyphs(glyphs[p.a], glyphs[p.b]))
		structDist[i] = float64(minLineDistance(occLines[p.a], occLines[p.b]))
		byBin[p.bin] = append(byBin[p.bin], i)
	}
	partialSpearman := func(structOf func(i int) float64) (float64, int) {
		sum, n := 0.0, 0
		for _, idxs := range byBin {
			if len(idxs) < 5 {
				continue
			}
			ex, sx, names := make([]float64, len(idxs)), make([]float64, len(idxs)), make([]string, len(idxs))
			for j, i := range idxs {
				ex[j], sx[j], names[j] = editDist[i], structOf(i), strconv.Itoa(i)
			}
			sum += spearman(ex, sx, names) * float64(len(idxs))
			n += len(idxs)
		}
		return sum, n
	}
	sum, n := partialSpearman(func(i int) float64 { return structDist[i] })
	if n == 0 {
		return NullTest{}, len(pairs), false
	}
	observed := math.Abs(sum / float64(n))
	null := make([]float64, repetitions)
	for r := range null {
		shuffled := append([]float64(nil), structDist...)
		for _, idxs := range byBin {
			perm := rng.Perm(len(idxs))
			orig := make([]float64, len(idxs))
			for j, i := range idxs {
				orig[j] = structDist[i]
			}
			for j, i := range idxs {
				shuffled[i] = orig[perm[j]]
			}
		}
		s, nn := partialSpearman(func(i int) float64 { return shuffled[i] })
		if nn > 0 {
			null[r] = math.Abs(s / float64(nn))
		}
	}
	return nullTest("cs7/edit-distance-x-structural-distance", "N6 frequency-bin-matched structural-distance permutation", observed, null), len(pairs), true
}

// ---- CS8: conditional persistence of CS1 across strata ----

func cs8Test(c corpus, familyOf map[string]int, repetitions int, rng *rand.Rand) []StabilityRun {
	strata := []struct {
		name string
		keep func(tokenRecord) bool
	}{
		{"CURRIER_A", func(r tokenRecord) bool { return r.Currier == "A" }},
		{"CURRIER_B", func(r tokenRecord) bool { return r.Currier == "B" }},
		{"LOCUS_TEXT", func(r tokenRecord) bool { return r.LocusType == "P" }},
	}
	out := make([]StabilityRun, 0, len(strata))
	for _, st := range strata {
		sub := corpus{info: c.info}
		for _, r := range c.records {
			if st.keep(r) {
				sub.records = append(sub.records, r)
			}
		}
		if len(sub.records) < 40 {
			out = append(out, StabilityRun{Perturbation: "cs1_conditioning", Value: st.name, Status: "INSUFFICIENT_DATA"})
			continue
		}
		famTest, _, n, _ := cs1Test(sub, familyOf, nil, repetitions, rng)
		if n < 20 {
			out = append(out, StabilityRun{Perturbation: "cs1_conditioning", Value: st.name, Status: "INSUFFICIENT_DATA", ComparableNodes: n})
			continue
		}
		status := "UNSTABLE"
		if famTest.PValue <= 0.05 {
			status = "PARTITION_SPECIFIC"
		}
		out = append(out, StabilityRun{
			Perturbation: "cs1_conditioning", Value: st.name, ARI: famTest.EffectSize, NMI: famTest.Observed,
			ComparableNodes: n, Status: status, Note: fmt.Sprintf("p=%.4f (within-stratum within-line null)", famTest.PValue),
		})
	}
	return out
}

// ---- stage 8: grouped held-out predictive validation ----

func laplaceProb(counts map[string]int, key string, total, categories int) float64 {
	if categories == 0 {
		categories = 1
	}
	return (float64(counts[key]) + 1) / (float64(total) + float64(categories))
}

// groupedKFoldLogLoss compares an M0 (marginal target distribution)
// baseline against an M1 (target given feature) model, evaluated
// out-of-fold with folds assigned at the group level (folio) so no
// occurrence's fold membership leaks information about a neighboring
// occurrence in the same folio.
func groupedKFoldLogLoss(records []tokenRecord, groupOf func(tokenRecord) string, featureOf func(tokenRecord) (string, bool), targetOf func(tokenRecord) (string, bool), folds int, seed int64) HeldOutResult {
	groupSet := map[string]bool{}
	for _, r := range records {
		if g := groupOf(r); g != "" {
			groupSet[g] = true
		}
	}
	groups := orderedKeysBool(groupSet)
	if len(groups) < 4 {
		return HeldOutResult{Scheme: "grouped_folio_kfold", Note: "fewer than 4 distinct groups (folios) available"}
	}
	if folds > len(groups) {
		folds = len(groups)
	}
	rng := rand.New(rand.NewSource(seed))
	perm := rng.Perm(len(groups))
	foldOf := map[string]int{}
	for i, p := range perm {
		foldOf[groups[p]] = i % folds
	}
	targetCats := map[string]bool{}
	for _, r := range records {
		if t, ok := targetOf(r); ok {
			targetCats[t] = true
		}
	}
	numCats := len(targetCats)
	var baselineLosses, modelLosses []float64
	for f := 0; f < folds; f++ {
		trainCounts, trainTotal := map[string]int{}, 0
		trainJoint := map[string]map[string]int{}
		trainJointTotal := map[string]int{}
		for _, r := range records {
			g := groupOf(r)
			if g == "" || foldOf[g] == f {
				continue
			}
			tgt, ok := targetOf(r)
			if !ok {
				continue
			}
			trainCounts[tgt]++
			trainTotal++
			if feat, fok := featureOf(r); fok {
				if trainJoint[feat] == nil {
					trainJoint[feat] = map[string]int{}
				}
				trainJoint[feat][tgt]++
				trainJointTotal[feat]++
			}
		}
		if trainTotal == 0 {
			continue
		}
		baseLoss, modelLoss, testN := 0.0, 0.0, 0
		for _, r := range records {
			g := groupOf(r)
			if g == "" || foldOf[g] != f {
				continue
			}
			tgt, ok := targetOf(r)
			if !ok {
				continue
			}
			pBase := laplaceProb(trainCounts, tgt, trainTotal, numCats)
			baseLoss += -math.Log(pBase)
			pModel := pBase
			if feat, fok := featureOf(r); fok && trainJoint[feat] != nil {
				pModel = laplaceProb(trainJoint[feat], tgt, trainJointTotal[feat], numCats)
			}
			modelLoss += -math.Log(pModel)
			testN++
		}
		if testN == 0 {
			continue
		}
		baselineLosses = append(baselineLosses, baseLoss/float64(testN))
		modelLosses = append(modelLosses, modelLoss/float64(testN))
	}
	if len(baselineLosses) == 0 {
		return HeldOutResult{Scheme: "grouped_folio_kfold", Note: "no fold produced held-out observations"}
	}
	improvement := make([]float64, len(baselineLosses))
	for i := range improvement {
		improvement[i] = baselineLosses[i] - modelLosses[i]
	}
	return HeldOutResult{
		Scheme: "grouped_folio_kfold", Folds: len(baselineLosses),
		BaselineLogLoss: mean(baselineLosses), ModelLogLoss: mean(modelLosses),
		Improvement: mean(improvement), ImprovementSD: sd(improvement, mean(improvement)), N: len(baselineLosses),
	}
}
