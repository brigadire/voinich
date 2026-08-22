// local-regime-topology-analyze implements Task65's diagnostic study of
// Task64's BROADER_LOCAL_REGIME: distance decay, change points, discrete
// clustering, metadata (Currier/Hand/Section/Page) effects and the
// Task64 discovery/replication discrepancy. It is an independent
// analysis, never a production stage; it does not touch Stages 1-28 and
// builds no generative model (task65 section 2).
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"zcore.dev/voinich/internal/evaglyph"
	"zcore.dev/voinich/internal/lineregime"
	"zcore.dev/voinich/internal/localregimetopology"
	"zcore.dev/voinich/internal/metadatavalidation"
)

const (
	outDir     = "experiments/local-regime-topology-v1"
	corpusPath = "data_work/ZL3b-x7.canonical.txt"
	ivtffPath  = "data/ZL3b-n.txt"

	primaryMinN = 5 // task64-identical minimum eligible line size, reused for reproduction

	// task64's exact split fractions, reused verbatim for byte/metric
	// reproduction (task65 section 33).
	task64BaseSeed = int64(64000)
	trainFrac      = 0.50
	valFrac        = 0.20
	discoveryFrac  = 0.70

	baseSeed = int64(65000)

	primaryWindow  = 20
	primaryStep    = primaryWindow / 4
	maxLagSteps    = 30
	minStrataN     = 200 // minimum tokens per metadata stratum before reporting (task65 section 71)
	minStrataPages = 5

	regimeKMin, regimeKMax = 2, 8
	bootstrapReps          = 200
)

var sensitivityWindows = []int{10, 20, 40, 80}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func put(p, s string) error { return os.WriteFile(p, []byte(s), 0644) }

// =====================================================================
// section 4/5: authoritative corpus + metadata loading and the unified
// manuscript coordinate.
// =====================================================================

func loadVoynichLines(path string) (sha string, tokensByLine [][][]string, err error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", nil, err
	}
	h := sha256.Sum256(b)
	sc := bufio.NewScanner(strings.NewReader(string(b)))
	sc.Buffer(make([]byte, 4096), 16<<20)
	for sc.Scan() {
		var line [][]string
		for _, raw := range strings.Fields(sc.Text()) {
			if g := evaglyph.CollapseEVA(raw); len(g) > 0 {
				line = append(line, g)
			}
		}
		tokensByLine = append(tokensByLine, line)
	}
	return hex.EncodeToString(h[:]), tokensByLine, sc.Err()
}

func loadNatural(path string) (tokensByLine [][][]string, err error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	sc := bufio.NewScanner(strings.NewReader(string(b)))
	sc.Buffer(make([]byte, 4096), 16<<20)
	for sc.Scan() {
		var line [][]string
		for _, raw := range strings.Fields(sc.Text()) {
			var g []string
			for _, r := range strings.ToLower(raw) {
				if unicode.IsLetter(r) || unicode.IsNumber(r) {
					g = append(g, string(r))
				}
			}
			if len(g) > 0 {
				line = append(line, g)
			}
		}
		tokensByLine = append(tokensByLine, line)
	}
	return tokensByLine, sc.Err()
}

func loadGeneratedTokens(path string) ([][]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out [][]string
	sc := bufio.NewScanner(strings.NewReader(string(b)))
	sc.Buffer(make([]byte, 4096), 16<<20)
	for sc.Scan() {
		if f := strings.Fields(sc.Text()); len(f) > 0 {
			out = append(out, f)
		}
	}
	return out, sc.Err()
}

// metadata is Task64's Folio/Currier($L, "Currier's Language A/B" per the
// corpus file's own page-header comments)/Hand($H)/Section($I, the
// standard IVTFF illustration-type code) metadata, joined by physical
// line index (task65 section 4). It intentionally does NOT reuse the
// "$C" field that other packages in this repo call "Currier" (Currier's
// own finer-grained hand numbering, a different and separately tracked
// scribal-identity dimension, documented in GENERIC_STAGE_APPLICABILITY_
// AUDIT.md) - task65's "Currier A/B" is unambiguously the two-valued
// Language field, per the literal comment "# Currier's Language A, hand
// 1" attached to $L=A/$H=1 in data/ZL3b-n.txt.
func metadata(path string, nLines int) (folio, currier, hand, section []string, ok bool) {
	doc, err := metadatavalidation.ParseIVTFF(path)
	if err != nil || len(doc.Loci) != nLines {
		return nil, nil, nil, nil, false
	}
	folio = make([]string, nLines)
	currier = make([]string, nLines)
	hand = make([]string, nLines)
	section = make([]string, nLines)
	for i, l := range doc.Loci {
		folio[i] = l.Folio
		currier[i] = l.Variables["L"]
		hand[i] = l.Variables["H"]
		section[i] = sectionLabel(l.Variables["I"])
	}
	return folio, currier, hand, section, true
}

var sectionNames = map[string]string{
	"A": "Astronomical", "B": "Biological", "C": "Cosmological", "H": "Herbal",
	"P": "Pharmaceutical", "S": "Stars", "T": "Text", "Z": "Zodiac",
}

func sectionLabel(code string) string {
	if name, ok := sectionNames[code]; ok {
		return name
	}
	return code
}

// buildRecords flattens tokensByLine into the unified manuscript
// coordinate (task65 section 5): GlobalIndex, Line, PageIndex plus every
// per-line metadata field, one record per token, in manuscript order -
// never assumed to be the original semantic order.
func buildRecords(tokensByLine [][][]string, folio, currier, hand, section []string) []localregimetopology.TokenRecord {
	var recs []localregimetopology.TokenRecord
	pageIdx := -1
	lastFolio := ""
	gi := 0
	for li, toks := range tokensByLine {
		f := ""
		if li < len(folio) {
			f = folio[li]
		}
		if f != lastFolio {
			pageIdx++
			lastFolio = f
		}
		c, h, s := "", "", ""
		if li < len(currier) {
			c, h, s = currier[li], hand[li], section[li]
		}
		for _, t := range toks {
			recs = append(recs, localregimetopology.TokenRecord{
				GlobalIndex: gi, Line: li, PageIndex: pageIdx,
				Folio: f, Currier: c, Hand: h, Section: s, Glyphs: t,
			})
			gi++
		}
	}
	return recs
}

func flatTokens(records []localregimetopology.TokenRecord) [][]string {
	out := make([][]string, len(records))
	for i, r := range records {
		out[i] = r.Glyphs
	}
	return out
}

func gitInfo() (commit string, dirty bool) {
	if out, err := exec.Command("git", "rev-parse", "HEAD").Output(); err == nil {
		commit = strings.TrimSpace(string(out))
	}
	if st, err := exec.Command("git", "status", "--porcelain").Output(); err == nil {
		dirty = len(strings.TrimSpace(string(st))) > 0
	}
	return
}

func sortedKeys[T any](m map[string]T) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func meanOf(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := 0.0
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

func medianOf(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	return s[len(s)/2]
}

// bootstrapCIFromSamples is a percentile bootstrap over xs itself (each
// element the resampling unit). For overlapping windows this understates
// true uncertainty (neighboring windows share tokens and are therefore
// correlated); that is documented in the design doc rather than solved
// with a full page-level block bootstrap, per section 74's performance
// mandate and the scope of this diagnostic task.
func bootstrapCIFromSamples(xs []float64, seed int64, reps int) (lo, hi float64) {
	if len(xs) == 0 {
		return 0, 0
	}
	r := rand.New(rand.NewSource(seed))
	means := make([]float64, reps)
	for i := 0; i < reps; i++ {
		s := 0.0
		for range xs {
			s += xs[r.Intn(len(xs))]
		}
		means[i] = s / float64(len(xs))
	}
	sort.Float64s(means)
	loIdx := int(0.025 * float64(reps))
	hiIdx := int(0.975*float64(reps)) - 1
	if hiIdx >= reps {
		hiIdx = reps - 1
	}
	if hiIdx < 0 {
		hiIdx = 0
	}
	return means[loIdx], means[hiIdx]
}

// =====================================================================
// sections 12-14: distance decay as a function of separation, in token,
// line and page units, each against a shuffle-order matched null.
// =====================================================================

type lagRow struct {
	unit                           string
	windowSize, lag, pairs         int
	mean, median, lo, hi, nullMean float64
}

func (r lagRow) row() string {
	return fmt.Sprintf("%s\t%d\t%d\t%d\t%.9f\t%.9f\t%.9f\t%.9f\t%.9f\n",
		r.unit, r.windowSize, r.lag, r.pairs, r.mean, r.median, r.lo, r.hi, r.nullMean)
}

// decayCurve computes D(R_x, R_x+k) for k=1..maxLag over an ordered
// profile sequence, plus a shuffle-order null (task65 section 12). It
// does not assume exponential decay (section 14): it just reports the
// empirical mean/median/CI per lag.
func decayCurve(unit string, windowSize int, profiles []lineregime.Profile, maxLag int, seed int64) []lagRow {
	shuffled := append([]lineregime.Profile(nil), profiles...)
	rand.New(rand.NewSource(seed)).Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	var out []lagRow
	for k := 1; k <= maxLag && k < len(profiles); k++ {
		var ds, nullDs []float64
		for i := 0; i+k < len(profiles); i++ {
			ds = append(ds, lineregime.ProfileDistance(profiles[i], profiles[i+k]))
			nullDs = append(nullDs, lineregime.ProfileDistance(shuffled[i], shuffled[i+k]))
		}
		if len(ds) == 0 {
			continue
		}
		lo, hi := bootstrapCIFromSamples(ds, seed+int64(k), 200)
		out = append(out, lagRow{unit, windowSize, k, len(ds), meanOf(ds), medianOf(ds), lo, hi, meanOf(nullDs)})
	}
	return out
}

// samePageOrderedProfiles groups profiles by page (in manuscript order
// within each page) for line/page-lag decay, mirroring Task64's regime
// persistence computation (task64 section 24).
func samePageLagRows(unit string, byPage map[string][]lineregime.Profile, pageOrder []string, maxLag int, seed int64) []lagRow {
	sums := make([]struct {
		pairs int
		ds    []float64
	}, maxLag)
	for _, pg := range pageOrder {
		ps := byPage[pg]
		for k := 1; k <= maxLag; k++ {
			for i := 0; i+k < len(ps); i++ {
				sums[k-1].ds = append(sums[k-1].ds, lineregime.ProfileDistance(ps[i], ps[i+k]))
			}
		}
	}
	// null: shuffle each page's own profile order, then recompute the same lag pairs.
	r := rand.New(rand.NewSource(seed))
	nullSums := make([][]float64, maxLag)
	for _, pg := range pageOrder {
		ps := append([]lineregime.Profile(nil), byPage[pg]...)
		r.Shuffle(len(ps), func(i, j int) { ps[i], ps[j] = ps[j], ps[i] })
		for k := 1; k <= maxLag; k++ {
			for i := 0; i+k < len(ps); i++ {
				nullSums[k-1] = append(nullSums[k-1], lineregime.ProfileDistance(ps[i], ps[i+k]))
			}
		}
	}
	var out []lagRow
	for k := 1; k <= maxLag; k++ {
		ds := sums[k-1].ds
		if len(ds) == 0 {
			continue
		}
		lo, hi := bootstrapCIFromSamples(ds, seed+int64(k)+1000, 200)
		out = append(out, lagRow{unit, 0, k, len(ds), meanOf(ds), medianOf(ds), lo, hi, meanOf(nullSums[k-1])})
	}
	return out
}

// =====================================================================
// section 13: correlation length from the empirical (not assumed
// exponential) token-lag decay curve.
// =====================================================================

// correlationLength finds the smallest lag at which the excess similarity
// over the null (row.nullMean - row.mean, since larger distance = less
// similar) has fallen to frac of its lag=1 value (task65 section 13). The
// definition (50% and 1/e thresholds) is fixed before results were
// inspected, not chosen after seeing the curve shape (section 14).
func correlationLength(rows []lagRow, frac float64) (float64, string) {
	if len(rows) == 0 {
		return 0, "NOT_APPLICABLE"
	}
	initialExcess := rows[0].nullMean - rows[0].mean
	if initialExcess <= 0 {
		return 0, "NO_INITIAL_EXCESS"
	}
	for _, r := range rows {
		excess := r.nullMean - r.mean
		if excess <= initialExcess*frac {
			return float64(r.lag), "OK"
		}
	}
	return float64(rows[len(rows)-1].lag), "EXCEEDS_MEASURED_RANGE"
}

// =====================================================================
// sections 15-17: same-page vs cross-page comparison and boundary
// discontinuity, defined relative to the ordinary decay curve rather
// than a raw distance comparison (section 16).
// =====================================================================

type boundaryRow struct {
	boundary                  string
	separation, n             int
	observed, expected, delta float64
}

func (b boundaryRow) row() string {
	return fmt.Sprintf("%s\t%d\t%d\t%.9f\t%.9f\t%.9f\n", b.boundary, b.separation, b.n, b.observed, b.expected, b.delta)
}

// expectedFromDecay interpolates (nearest available lag) the ordinary
// decay curve's mean distance at a given separation.
func expectedFromDecay(rows []lagRow, sep int) float64 {
	if len(rows) == 0 {
		return 0
	}
	best := rows[0]
	for _, r := range rows {
		if abs(r.lag-sep) < abs(best.lag-sep) {
			best = r
		}
	}
	return best.mean
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// boundaryDiscontinuity computes Delta_boundary for windows that straddle
// a line or page boundary at the primary window's own token separation
// (windowSize), so it is compared like-for-like against the ordinary
// decay curve at that same separation (task65 sections 16-17).
func boundaryDiscontinuity(windows []localregimetopology.Window, profiles []lineregime.Profile, decay []lagRow, crossPage bool) boundaryRow {
	var ds []float64
	for i, w := range windows {
		crosses := w.CrossesPage
		if !crossPage {
			crosses = w.CrossesLine
		}
		if crosses && i+1 < len(windows) {
			ds = append(ds, lineregime.ProfileDistance(profiles[i], profiles[i+1]))
		}
	}
	label := "LINE_BOUNDARY"
	if crossPage {
		label = "PAGE_BOUNDARY"
	}
	if len(ds) == 0 {
		return boundaryRow{label, primaryStep, 0, 0, 0, 0}
	}
	obs := meanOf(ds)
	exp := expectedFromDecay(decay, 1)
	return boundaryRow{label, primaryStep, len(ds), obs, exp, obs - exp}
}

// =====================================================================
// sections 25-28: Currier/Hand/Section/Page metadata effects, computed
// on one aggregate profile per page (so N is pages, not pseudo-replicated
// tokens/windows) to keep between/within comparisons honest.
// =====================================================================

type metadataEffectRow struct {
	factor, level           string
	pages                   int
	betweenMean, withinMean float64
	status                  string
}

func (m metadataEffectRow) row() string {
	if m.status != "OK" {
		return fmt.Sprintf("%s\t%s\t%d\tNA\tNA\t%s\n", m.factor, m.level, m.pages, m.status)
	}
	return fmt.Sprintf("%s\t%s\t%d\t%.9f\t%.9f\t%s\n", m.factor, m.level, m.pages, m.betweenMean, m.withinMean, m.status)
}

// metadataEffect compares between-group and within-group page-profile
// distance for one factor (task65 sections 25-28); groups with fewer than
// minStrataPages pages are NOT_APPLICABLE rather than reported noisily
// (section 26/71).
func metadataEffect(factor string, groupOf map[string]string, pageProfiles map[string]lineregime.Profile, pageOrder []string, seed int64) []metadataEffectRow {
	byGroup := map[string][]string{}
	for _, pg := range pageOrder {
		g := groupOf[pg]
		if g == "" {
			continue
		}
		byGroup[g] = append(byGroup[g], pg)
	}
	var rows []metadataEffectRow
	r := rand.New(rand.NewSource(seed))
	for _, g := range sortedKeys(byGroup) {
		pages := byGroup[g]
		if len(pages) < minStrataPages {
			rows = append(rows, metadataEffectRow{factor, g, len(pages), 0, 0, "INSUFFICIENT_DATA"})
			continue
		}
		var within []float64
		for i := 0; i < len(pages); i++ {
			for j := i + 1; j < len(pages); j++ {
				within = append(within, lineregime.ProfileDistance(pageProfiles[pages[i]], pageProfiles[pages[j]]))
			}
		}
		var other []string
		for _, pg := range pageOrder {
			if groupOf[pg] != "" && groupOf[pg] != g {
				other = append(other, pg)
			}
		}
		var between []float64
		draws := len(pages) * 5
		for i := 0; i < draws && len(other) > 0; i++ {
			a := pages[r.Intn(len(pages))]
			b := other[r.Intn(len(other))]
			between = append(between, lineregime.ProfileDistance(pageProfiles[a], pageProfiles[b]))
		}
		rows = append(rows, metadataEffectRow{factor, g, len(pages), meanOf(between), meanOf(within), "OK"})
	}
	return rows
}

// =====================================================================
// section 29: hierarchical variance decomposition (descriptive, one
// factor at a time - not a claim of a strict nesting or of causality).
// =====================================================================

func varianceShare(values []float64, groupOf func(i int) string) float64 {
	if len(values) < 2 {
		return 0
	}
	grand := meanOf(values)
	totalSS := 0.0
	for _, v := range values {
		totalSS += (v - grand) * (v - grand)
	}
	if totalSS == 0 {
		return 0
	}
	sums := map[string][]float64{}
	for i, v := range values {
		g := groupOf(i)
		sums[g] = append(sums[g], v)
	}
	betweenSS := 0.0
	for _, g := range sortedKeys(sums) {
		vs := sums[g]
		m := meanOf(vs)
		betweenSS += float64(len(vs)) * (m - grand) * (m - grand)
	}
	return betweenSS / totalSS
}

// =====================================================================
// section 30/31: metadata-conditioned and same-page-conditioned decay.
// =====================================================================

func metadataConditionedDecay(unit string, windows []localregimetopology.Window, profiles []lineregime.Profile, groupOf func(w localregimetopology.Window) string, maxLag int, seed int64) []lagRow {
	byGroup := map[string][]int{}
	for i, w := range windows {
		g := groupOf(w)
		if g == "" {
			continue
		}
		byGroup[g] = append(byGroup[g], i)
	}
	var out []lagRow
	for _, g := range sortedKeys(byGroup) {
		idx := byGroup[g]
		if len(idx) < 2*maxLag {
			continue
		}
		ps := make([]lineregime.Profile, len(idx))
		for i, ix := range idx {
			ps[i] = profiles[ix]
		}
		rows := decayCurve(unit+":"+g, 0, ps, maxLag, seed)
		out = append(out, rows...)
	}
	return out
}

// =====================================================================
// sections 32-34: Task64 discovery/replication split reproduction and
// diagnosis. This MUST reproduce Task64's own numbers (section 33); a
// mismatch is a hard stop (section 33/DEFINITION OF DONE item 1).
// =====================================================================

func composition(lines []lineregime.Line, keep map[string]bool, of func(l lineregime.Line) string) map[string]int {
	out := map[string]int{}
	for _, l := range lines {
		if keep[l.Folio] {
			out[of(l)]++
		}
	}
	return out
}

// =====================================================================
// section 70: region-wise effect map, reusing Task64's own ComputeCoreStats
// (via the shared internal/lineregime package) restricted to each
// metadata region's lines, so this is the same Delta_line definition
// Task64 used, not a re-derived approximation.
type regionEffectRow struct {
	region         string
	n              int
	effect, lo, hi float64
	reliability    string
}

func (r regionEffectRow) row() string {
	if r.reliability != "OK" {
		return fmt.Sprintf("%s\t%d\tNA\tNA\tNA\t%s\n", r.region, r.n, r.reliability)
	}
	return fmt.Sprintf("%s\t%d\t%.9f\t%.9f\t%.9f\t%s\n", r.region, r.n, r.effect, r.lo, r.hi, r.reliability)
}

func regionEffect(region string, regionLines []lineregime.Line, seed int64) regionEffectRow {
	elig := lineregime.Eligible(regionLines, primaryMinN)
	if len(elig) < 30 {
		return regionEffectRow{region, len(elig), 0, 0, 0, "INSUFFICIENT_DATA"}
	}
	cs := lineregime.ComputeCoreStats(regionLines, primaryMinN, seed)
	effect := cs.NonAdj.Rate() - cs.DiffLineSamePage.Rate()
	mean, lo, hi := bootstrapLineDelta(cs.PerLine, seed+1, 300)
	_ = mean
	return regionEffectRow{region, len(elig), effect, lo, hi, "OK"}
}

func bootstrapLineDelta(contribs []lineregime.LineContribution, seed int64, reps int) (mean, lo, hi float64) {
	if len(contribs) == 0 {
		return 0, 0, 0
	}
	r := rand.New(rand.NewSource(seed))
	deltas := make([]float64, reps)
	for i := 0; i < reps; i++ {
		var nA, nN, cA, cN int
		for range contribs {
			c := contribs[r.Intn(len(contribs))]
			nA += c.NonAdjDLE1
			nN += c.NonAdjN
			cA += c.CtrlDLE1
			cN += c.CtrlN
		}
		rn, rc := 0.0, 0.0
		if nN > 0 {
			rn = float64(nA) / float64(nN)
		}
		if cN > 0 {
			rc = float64(cA) / float64(cN)
		}
		deltas[i] = rn - rc
	}
	sort.Float64s(deltas)
	mean = meanOf(deltas)
	lo = deltas[int(0.025*float64(reps))]
	hiIdx := int(0.975*float64(reps)) - 1
	if hiIdx >= reps {
		hiIdx = reps - 1
	}
	hi = deltas[hiIdx]
	return
}

// =====================================================================
// sections 20-22, 49-50: synthetic controls used both to calibrate the
// change-point null (section 19) and to validate the whole pipeline
// against known ground truth. Vocabulary "flavors" are contiguous chunks
// of the real corpus's own sorted unique-token vocabulary, a deterministic
// partition chosen before any result was inspected (not fit to any
// Voynich effect).
// =====================================================================

func uniqueVocab(tokens [][]string) [][]string {
	seen := map[string][]string{}
	for _, t := range tokens {
		key := strings.Join(t, "")
		seen[key] = t
	}
	keys := sortedKeys(seen)
	out := make([][]string, len(keys))
	for i, k := range keys {
		out[i] = seen[k]
	}
	return out
}

func splitVocabContiguous(vocabSorted [][]string, k int) [][][]string {
	out := make([][][]string, k)
	n := len(vocabSorted)
	for i := 0; i < k; i++ {
		start, end := i*n/k, (i+1)*n/k
		if end <= start {
			end = start + 1
		}
		if end > n {
			end = n
		}
		out[i] = vocabSorted[start:end]
	}
	return out
}

func sampleFrom(vocab [][]string, r *rand.Rand) []string {
	if len(vocab) == 0 {
		return []string{"y"}
	}
	return vocab[r.Intn(len(vocab))]
}

func syntheticStationary(vocab [][]string, n int, seed int64) [][]string {
	r := rand.New(rand.NewSource(seed))
	out := make([][]string, n)
	for i := range out {
		out[i] = sampleFrom(vocab, r)
	}
	return out
}

func syntheticDrift(lo, hi [][]string, n int, seed int64) [][]string {
	r := rand.New(rand.NewSource(seed))
	out := make([][]string, n)
	for i := range out {
		p := float64(i) / float64(max(1, n-1))
		if r.Float64() < p {
			out[i] = sampleFrom(hi, r)
		} else {
			out[i] = sampleFrom(lo, r)
		}
	}
	return out
}

func syntheticDiscrete(states [][][]string, blockSize, n int, seed int64) ([][]string, []int) {
	r := rand.New(rand.NewSource(seed))
	out := make([][]string, n)
	var boundaries []int
	stateIdx := 0
	for pos := 0; pos < n; pos += blockSize {
		end := pos + blockSize
		if end > n {
			end = n
		}
		st := states[stateIdx%len(states)]
		for i := pos; i < end; i++ {
			out[i] = sampleFrom(st, r)
		}
		if end < n {
			boundaries = append(boundaries, end)
		}
		stateIdx++
	}
	return out, boundaries
}

func syntheticMixed(states [][][]string, blockSize, n int, seed int64) ([][]string, []int) {
	r := rand.New(rand.NewSource(seed))
	out := make([][]string, n)
	var boundaries []int
	stateIdx := 0
	for pos := 0; pos < n; pos += blockSize {
		end := pos + blockSize
		if end > n {
			end = n
		}
		st := states[stateIdx%len(states)]
		half := len(st) / 2
		if half < 1 {
			half = 1
		}
		lo, hi := st[:half], st[half:]
		if len(hi) == 0 {
			hi = st
		}
		for i := pos; i < end; i++ {
			frac := float64(i-pos) / float64(max(1, end-pos-1))
			if r.Float64() < frac {
				out[i] = sampleFrom(hi, r)
			} else {
				out[i] = sampleFrom(lo, r)
			}
		}
		if end < n {
			boundaries = append(boundaries, end)
		}
		stateIdx++
	}
	return out, boundaries
}

func toRecords(tokens [][]string, folio string) []localregimetopology.TokenRecord {
	out := make([]localregimetopology.TokenRecord, len(tokens))
	for i, t := range tokens {
		out[i] = localregimetopology.TokenRecord{GlobalIndex: i, Line: i / 10, PageIndex: 0, Folio: folio, Glyphs: t}
	}
	return out
}

// changePointCount reports how many ScanChangePoints scores exceed
// threshold (task65 sections 19, 21, 53).
func changePointCount(records []localregimetopology.TokenRecord, w int, giant map[string]bool, threshold float64) int {
	n := 0
	for _, cp := range localregimetopology.ScanChangePoints(records, w, giant) {
		if cp.Score > threshold {
			n++
		}
	}
	return n
}

// boundaryRecoveryRate measures, for a discrete/mixed control with known
// true boundaries, the fraction of true boundaries with a detected
// (above-threshold) change point within tolerance tokens (task65
// section 20).
func boundaryRecoveryRate(records []localregimetopology.TokenRecord, w int, giant map[string]bool, threshold float64, trueBoundaries []int, tolerance int) float64 {
	cps := localregimetopology.ScanChangePoints(records, w, giant)
	var detected []int
	for _, cp := range cps {
		if cp.Score > threshold {
			detected = append(detected, cp.Position)
		}
	}
	if len(trueBoundaries) == 0 {
		return 0
	}
	hit := 0
	for _, tb := range trueBoundaries {
		for _, d := range detected {
			if abs(d-tb) <= tolerance {
				hit++
				break
			}
		}
	}
	return float64(hit) / float64(len(trueBoundaries))
}

func percentile(xs []float64, p float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	idx := int(p * float64(len(s)))
	if idx >= len(s) {
		idx = len(s) - 1
	}
	if idx < 0 {
		idx = 0
	}
	return s[idx]
}

// =====================================================================
// sections 36-43: discrete clustering, stability, recurrence, transitions,
// dwell length, and within-cluster drift.
// =====================================================================

func extendedVector(p lineregime.Profile, degree, nearRepeat float64) []float64 {
	return []float64{p.MeanLen, p.GiantFrac, p.TopInit, p.TopFinal, p.TypeEnt, degree, nearRepeat}
}

func windowNearRepeatRate(tokens [][]string) float64 {
	pairs := lineregime.WithinLinePairs(lineregime.Line{Tokens: tokens})
	if len(pairs) == 0 {
		return 0
	}
	hits := 0
	for _, p := range pairs {
		if p.Distance <= 1 {
			hits++
		}
	}
	return float64(hits) / float64(len(pairs))
}

func windowMeanDegree(tokens [][]string, degrees map[string]int) float64 {
	if len(tokens) == 0 {
		return 0
	}
	sum := 0
	for _, t := range tokens {
		sum += degrees[strings.Join(t, "")]
	}
	return float64(sum) / float64(len(tokens))
}

// coClusterAgreement is a simple bootstrap-stability statistic (task65
// section 38): the fraction of pairs co-clustered in the original
// assignment that remain co-clustered in the resampled assignment.
func coClusterAgreement(original, resampled []int, sampleIdx []int) float64 {
	agree, total := 0, 0
	for i := 0; i < len(sampleIdx); i++ {
		for j := i + 1; j < len(sampleIdx); j++ {
			origSame := original[sampleIdx[i]] == original[sampleIdx[j]]
			newSame := resampled[i] == resampled[j]
			if origSame {
				total++
				if newSame {
					agree++
				}
			}
		}
	}
	if total == 0 {
		return 0
	}
	return float64(agree) / float64(total)
}

func run() error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	corpusSHA, tokensByLine, err := loadVoynichLines(corpusPath)
	if err != nil {
		return err
	}
	folio, currier, hand, section, metaOK := metadata(ivtffPath, len(tokensByLine))
	lines := lineregime.BuildLines(tokensByLine, folio, currier, hand, metaOK)
	records := buildRecords(tokensByLine, folio, currier, hand, section)
	giant := lineregime.BuildGiantSet(flatTokens(records))
	pageOrder := lineregime.PageOrderOf(lines)
	split := lineregime.SplitPages(pageOrder, trainFrac, valFrac, discoveryFrac)

	// ---- sections 6-7, 12: primary token-lag decay (W=20, step=5) ----
	primaryWindows := localregimetopology.BuildWindows(records, primaryWindow, primaryStep)
	primaryProfiles := make([]lineregime.Profile, len(primaryWindows))
	for i, w := range primaryWindows {
		primaryProfiles[i] = lineregime.ComputeProfile(w.Tokens, giant)
	}
	var decayRows []lagRow
	decayRows = append(decayRows, decayCurve("TOKEN", primaryWindow, primaryProfiles, maxLagSteps, baseSeed+1)...)
	tokenDecay := decayRows // primary curve, used by correlation length / boundary discontinuity

	for _, w := range sensitivityWindows {
		if w == primaryWindow {
			continue
		}
		ws := localregimetopology.BuildWindows(records, w, w/4)
		ps := make([]lineregime.Profile, len(ws))
		for i, win := range ws {
			ps[i] = lineregime.ComputeProfile(win.Tokens, giant)
		}
		decayRows = append(decayRows, decayCurve("TOKEN", w, ps, maxLagSteps, baseSeed+2+int64(w))...)
	}

	// line-lag and page-lag decay, same-page-ordered (task64-style).
	elig := lineregime.Eligible(lines, primaryMinN)
	lineProfilesByPage := map[string][]lineregime.Profile{}
	for _, l := range elig {
		lineProfilesByPage[l.Folio] = append(lineProfilesByPage[l.Folio], lineregime.ComputeProfile(l.Tokens, giant))
	}
	decayRows = append(decayRows, samePageLagRows("LINE", lineProfilesByPage, pageOrder, 10, baseSeed+3)...)

	pageProfiles := map[string]lineregime.Profile{}
	pageTokens := map[string][][]string{}
	for _, r := range records {
		pageTokens[r.Folio] = append(pageTokens[r.Folio], r.Glyphs)
	}
	for pg, toks := range pageTokens {
		pageProfiles[pg] = lineregime.ComputeProfile(toks, giant)
	}
	pageProfileSeq := make([]lineregime.Profile, len(pageOrder))
	for i, pg := range pageOrder {
		pageProfileSeq[i] = pageProfiles[pg]
	}
	pageLagRows := decayCurve("PAGE", 0, pageProfileSeq, 10, baseSeed+4)
	decayRows = append(decayRows, pageLagRows...)

	decayTSV := "Unit\tWindowSize\tLag\tPairs\tMeanDistance\tMedian\tCILo\tCIHi\tNullMean\n"
	for _, r := range decayRows {
		decayTSV += r.row()
	}

	// ---- section 13: correlation length ----
	l50Token, s50Token := correlationLength(tokenDecay, 0.5)
	lEToken, sEToken := correlationLength(tokenDecay, 1/math.E)
	lineLagRows := samePageLagRows("LINE", lineProfilesByPage, pageOrder, 10, baseSeed+3)
	l50Line, s50Line := correlationLength(lineLagRows, 0.5)
	l50Page, s50Page := correlationLength(pageLagRows, 0.5)
	corrLenTSV := "Unit\tThreshold\tValue\tStatus\n" +
		fmt.Sprintf("TOKEN\t50pct\t%.9f\t%s\n", l50Token*primaryStep, s50Token) +
		fmt.Sprintf("TOKEN\t1_over_e\t%.9f\t%s\n", lEToken*primaryStep, sEToken) +
		fmt.Sprintf("LINE\t50pct\t%.9f\t%s\n", l50Line, s50Line) +
		fmt.Sprintf("PAGE\t50pct\t%.9f\t%s\n", l50Page, s50Page)

	// ---- sections 15-17: boundary discontinuity ----
	pageBoundary := boundaryDiscontinuity(primaryWindows, primaryProfiles, tokenDecay, true)
	lineBoundary := boundaryDiscontinuity(primaryWindows, primaryProfiles, tokenDecay, false)
	boundaryTSV := "Boundary\tSeparationTokens\tN\tObservedDistance\tExpectedFromDecay\tDelta\n" + pageBoundary.row() + lineBoundary.row()

	// ---- sections 25-28: Currier/Hand/Section/Page metadata effects ----
	currierOfPage, handOfPage, sectionOfPage := map[string]string{}, map[string]string{}, map[string]string{}
	for _, l := range lines {
		if _, ok := currierOfPage[l.Folio]; !ok {
			currierOfPage[l.Folio] = l.Currier
			handOfPage[l.Folio] = l.Hand
		}
	}
	for li, s := range section {
		if li < len(lines) {
			if _, ok := sectionOfPage[lines[li].Folio]; !ok {
				sectionOfPage[lines[li].Folio] = s
			}
		}
	}
	metadataTSV := "Factor\tLevel\tPages\tBetweenMeanDistance\tWithinMeanDistance\tStatus\n"
	for _, r := range metadataEffect("CURRIER", currierOfPage, pageProfiles, pageOrder, baseSeed+10) {
		metadataTSV += r.row()
	}
	for _, r := range metadataEffect("HAND", handOfPage, pageProfiles, pageOrder, baseSeed+11) {
		metadataTSV += r.row()
	}
	for _, r := range metadataEffect("SECTION", sectionOfPage, pageProfiles, pageOrder, baseSeed+12) {
		metadataTSV += r.row()
	}

	// ---- section 29: hierarchical variance decomposition (MeanLen as
	// the representative scalar feature; repeated for GiantFrac) ----
	pageList := pageOrder
	meanLenByPage := make([]float64, len(pageList))
	giantFracByPage := make([]float64, len(pageList))
	for i, pg := range pageList {
		meanLenByPage[i] = pageProfiles[pg].MeanLen
		giantFracByPage[i] = pageProfiles[pg].GiantFrac
	}
	hvTSV := "Feature\tFactor\tVarianceShare\n"
	for _, feat := range []struct {
		name   string
		values []float64
	}{{"MeanTokenLength", meanLenByPage}, {"GiantFraction", giantFracByPage}} {
		hvTSV += fmt.Sprintf("%s\tCURRIER\t%.9f\n", feat.name, varianceShare(feat.values, func(i int) string { return currierOfPage[pageList[i]] }))
		hvTSV += fmt.Sprintf("%s\tHAND\t%.9f\n", feat.name, varianceShare(feat.values, func(i int) string { return handOfPage[pageList[i]] }))
		hvTSV += fmt.Sprintf("%s\tSECTION\t%.9f\n", feat.name, varianceShare(feat.values, func(i int) string { return sectionOfPage[pageList[i]] }))
	}

	// ---- section 30/31: metadata-conditioned and same-page decay ----
	handOfWin := func(w localregimetopology.Window) string { return handOfPage[w.Folio] }
	currierOfWin := func(w localregimetopology.Window) string { return currierOfPage[w.Folio] }
	sectionOfWin := func(w localregimetopology.Window) string { return sectionOfPage[w.Folio] }
	condTSV := "ConditionedOn\tLag\tPairs\tMeanDistance\tMedian\tCILo\tCIHi\tNullMean\n"
	for _, r := range metadataConditionedDecay("CURRIER", primaryWindows, primaryProfiles, currierOfWin, 15, baseSeed+20) {
		condTSV += fmt.Sprintf("%s\t%d\t%d\t%.9f\t%.9f\t%.9f\t%.9f\t%.9f\n", r.unit, r.lag, r.pairs, r.mean, r.median, r.lo, r.hi, r.nullMean)
	}
	for _, r := range metadataConditionedDecay("HAND", primaryWindows, primaryProfiles, handOfWin, 15, baseSeed+21) {
		condTSV += fmt.Sprintf("%s\t%d\t%d\t%.9f\t%.9f\t%.9f\t%.9f\t%.9f\n", r.unit, r.lag, r.pairs, r.mean, r.median, r.lo, r.hi, r.nullMean)
	}
	for _, r := range metadataConditionedDecay("SECTION", primaryWindows, primaryProfiles, sectionOfWin, 15, baseSeed+22) {
		condTSV += fmt.Sprintf("%s\t%d\t%d\t%.9f\t%.9f\t%.9f\t%.9f\t%.9f\n", r.unit, r.lag, r.pairs, r.mean, r.median, r.lo, r.hi, r.nullMean)
	}
	// same-page conditioning (section 31): the LINE-unit decay above is
	// already computed strictly within-page (samePageLagRows never pairs
	// across pages), so it doubles as the same-page-conditioning check -
	// a decay that survives there cannot be explained by Currier/Hand/
	// Section/page-level composition, since page is held fixed.
	for _, r := range lineLagRows {
		condTSV += fmt.Sprintf("SAME_PAGE(LINE)\t%d\t%d\t%.9f\t%.9f\t%.9f\t%.9f\t%.9f\n", r.lag, r.pairs, r.mean, r.median, r.lo, r.hi, r.nullMean)
	}

	// ---- sections 32-34: Task64 split reproduction and diagnosis ----
	realCS := lineregime.ComputeCoreStats(lines, primaryMinN, task64BaseSeed+1)
	task64NullMean := realCS.DiffPage.Rate()
	discoveryLines := lineregime.FilterByPages(lines, split.Discovery)
	replicationLines := lineregime.FilterByPages(lines, split.Replication)
	discoveryEffect := lineregime.RateOnly(discoveryLines, primaryMinN, false) - task64NullMean
	replicationEffect := lineregime.RateOnly(replicationLines, primaryMinN, false) - task64NullMean
	const task64DiscoveryRef, task64ReplicationRef = 0.030069436, 0.003132198
	const reproductionTolerance = 1e-6
	reproduced := math.Abs(discoveryEffect-task64DiscoveryRef) < reproductionTolerance &&
		math.Abs(replicationEffect-task64ReplicationRef) < reproductionTolerance
	provenanceStatus := "REPRODUCED"
	if !reproduced {
		provenanceStatus = "INVALID_PROVENANCE"
	}

	currierOfLine := func(l lineregime.Line) string { return l.Currier }
	handOfLine := func(l lineregime.Line) string { return l.Hand }
	sectionOfLine := func(l lineregime.Line) string {
		if l.Index < len(section) {
			return section[l.Index]
		}
		return ""
	}
	splitDiag := "Fold\tTotalLines\tCompositionByCurrier\tCompositionByHand\tCompositionBySection\tDeltaLineEffect\tNullMean\tProvenanceStatus\n"
	splitDiag += fmt.Sprintf("DISCOVERY\t%d\t%s\t%s\t%s\t%.9f\t%.9f\t%s\n",
		len(lineregime.Eligible(discoveryLines, primaryMinN)),
		compositionInline(composition(lines, split.Discovery, currierOfLine)),
		compositionInline(composition(lines, split.Discovery, handOfLine)),
		compositionInline(composition(lines, split.Discovery, sectionOfLine)),
		discoveryEffect, task64NullMean, provenanceStatus)
	splitDiag += fmt.Sprintf("REPLICATION\t%d\t%s\t%s\t%s\t%.9f\t%.9f\t%s\n",
		len(lineregime.Eligible(replicationLines, primaryMinN)),
		compositionInline(composition(lines, split.Replication, currierOfLine)),
		compositionInline(composition(lines, split.Replication, handOfLine)),
		compositionInline(composition(lines, split.Replication, sectionOfLine)),
		replicationEffect, task64NullMean, provenanceStatus)
	splitDiag += fmt.Sprintf("# reference Task64 REGIME_SCALE_COMPARISON.tsv LINE row: Discovery=%.9f Replication=%.9f\n", task64DiscoveryRef, task64ReplicationRef)

	if !reproduced {
		// section 33: a failed reproduction is a hard stop.
		if err := put(filepath.Join(outDir, "TASK64_SPLIT_DIAGNOSIS.tsv"), splitDiag); err != nil {
			return err
		}
		return fmt.Errorf("INVALID_PROVENANCE: Task64 discovery/replication effect not reproduced (got %.9f/%.9f, want %.9f/%.9f)",
			discoveryEffect, replicationEffect, task64DiscoveryRef, task64ReplicationRef)
	}

	// ---- section 34: standardized discovery/replication comparison
	// after stratifying by the available metadata (Currier/Hand/Section).
	splitStandardized := "Stratum\tDiscoveryEffect\tReplicationEffect\tDeltaOfDeltas\tStatus\n"
	for _, factorName := range []string{"CURRIER", "HAND", "SECTION"} {
		var valueOf func(l lineregime.Line) string
		switch factorName {
		case "CURRIER":
			valueOf = currierOfLine
		case "HAND":
			valueOf = handOfLine
		default:
			valueOf = sectionOfLine
		}
		levels := map[string]bool{}
		for _, l := range lines {
			if v := valueOf(l); v != "" {
				levels[v] = true
			}
		}
		for _, lvl := range sortedKeys(levels) {
			dLines := lineregime.FilterByPages(discoveryLines, foliosWith(discoveryLines, valueOf, lvl))
			rLines := lineregime.FilterByPages(replicationLines, foliosWith(replicationLines, valueOf, lvl))
			dElig, rElig := lineregime.Eligible(dLines, primaryMinN), lineregime.Eligible(rLines, primaryMinN)
			if len(dElig) < 30 || len(rElig) < 30 {
				splitStandardized += fmt.Sprintf("%s=%s\tNA\tNA\tNA\tINSUFFICIENT_DATA\n", factorName, lvl)
				continue
			}
			de := lineregime.RateOnly(dLines, primaryMinN, false) - task64NullMean
			re := lineregime.RateOnly(rLines, primaryMinN, false) - task64NullMean
			splitStandardized += fmt.Sprintf("%s=%s\t%.9f\t%.9f\t%.9f\tOK\n", factorName, lvl, de, re, de-re)
		}
	}
	splitInterpretation := "TRUE_REGIME_HETEROGENEITY"
	{
		anyOK, allSmallGap := false, true
		for _, ln := range strings.Split(splitStandardized, "\n") {
			f := strings.Split(ln, "\t")
			if len(f) < 5 || f[4] != "OK" {
				continue
			}
			anyOK = true
			gap, _ := strconv.ParseFloat(f[3], 64)
			if math.Abs(gap) > 0.005 {
				allSmallGap = false
			}
		}
		switch {
		case !anyOK:
			splitInterpretation = "INSUFFICIENT_DATA"
		case allSmallGap:
			splitInterpretation = "COMPOSITION_EXPLAINED"
		}
	}

	// ---- section 70: region-wise effect map ----
	regionTSV := "Region\tN\tEffect\tCILo\tCIHi\tReliability\n"
	byCurrier := linesByValue(lines, currierOfLine)
	for i, lvl := range sortedKeys(byCurrier) {
		regionTSV += regionEffect("CURRIER="+lvl, byCurrier[lvl], baseSeed+40+int64(i)).row()
	}
	byHand := linesByValue(lines, handOfLine)
	for i, lvl := range sortedKeys(byHand) {
		regionTSV += regionEffect("HAND="+lvl, byHand[lvl], baseSeed+50+int64(i)).row()
	}
	bySection := linesByValue(lines, sectionOfLine)
	for i, lvl := range sortedKeys(bySection) {
		regionTSV += regionEffect("SECTION="+lvl, bySection[lvl], baseSeed+60+int64(i)).row()
	}

	// ---- sections 18-24: change-point analysis with synthetic-null calibration ----
	globalVocab := uniqueVocab(flatTokens(records))
	flavors3 := splitVocabContiguous(globalVocab, 3)
	flavors2 := splitVocabContiguous(globalVocab, 2)
	nTok := len(records)

	stationaryTokens := syntheticStationary(globalVocab, nTok, baseSeed+100)
	driftTokens := syntheticDrift(flavors2[0], flavors2[1], nTok, baseSeed+101)
	discreteTokens, discreteBoundaries := syntheticDiscrete(flavors3, 500, nTok, baseSeed+102)
	mixedTokens, mixedBoundaries := syntheticMixed(flavors3, 500, nTok, baseSeed+103)

	stationaryRecords := toRecords(stationaryTokens, "SYN")
	driftRecords := toRecords(driftTokens, "SYN")

	stationaryScores := localregimetopology.ScanChangePoints(stationaryRecords, primaryWindow, giant)
	stationaryScoreVals := scoreValues(stationaryScores)
	threshold95 := percentile(stationaryScoreVals, 0.95)
	threshold99 := percentile(stationaryScoreVals, 0.99)

	realScores := localregimetopology.ScanChangePoints(records, primaryWindow, giant)
	_, _ = localregimetopology.CUSUMMax(scoreValues(realScores))

	changePointsTSV := "Position\tPage\tLine\tScore\tSignificance\tDiscovery\tReplication\tMetadataBoundary\n"
	var allPositions, significantPositions []int
	metaBoundaryFlag := map[int]bool{}
	for _, cp := range realScores {
		rec := records[cp.Position]
		sig := "NOT_SIGNIFICANT"
		if cp.Score > threshold99 {
			sig = "P99"
		} else if cp.Score > threshold95 {
			sig = "P95"
		}
		disc, repl := "NO", "NO"
		if split.Discovery[rec.Folio] {
			disc = "YES"
		}
		if split.Replication[rec.Folio] {
			repl = "YES"
		}
		mb := metadataBoundaryNear(records, cp.Position, primaryWindow)
		changePointsTSV += fmt.Sprintf("%d\t%s\t%d\t%.9f\t%s\t%s\t%s\t%s\n", cp.Position, rec.Folio, rec.Line, cp.Score, sig, disc, repl, mb)
		allPositions = append(allPositions, cp.Position)
		metaBoundaryFlag[cp.Position] = mb != "NONE"
		if cp.Score > threshold95 {
			significantPositions = append(significantPositions, cp.Position)
		}
	}

	driftCount95 := changePointCount(driftRecords, primaryWindow, giant, threshold95)
	stationaryCount95 := changePointCount(stationaryRecords, primaryWindow, giant, threshold95)
	realCount95 := len(significantPositions)
	changePointNullsTSV := "Control\tMeanScore\tP95\tP99\tCountAboveP95\n" +
		fmt.Sprintf("STATIONARY\t%.9f\t%.9f\t%.9f\t%d\n", meanOf(stationaryScoreVals), threshold95, threshold99, stationaryCount95) +
		fmt.Sprintf("SMOOTH_DRIFT\t%.9f\t%.9f\t%.9f\t%d\n", meanOf(scoreValues(localregimetopology.ScanChangePoints(driftRecords, primaryWindow, giant))), threshold95, threshold99, driftCount95) +
		fmt.Sprintf("VOYNICH\t%.9f\t%.9f\t%.9f\t%d\n", meanOf(scoreValues(realScores)), threshold95, threshold99, realCount95)

	observedOverlap := 0
	for _, p := range significantPositions {
		if metaBoundaryFlag[p] {
			observedOverlap++
		}
	}
	observedRate := 0.0
	if len(significantPositions) > 0 {
		observedRate = float64(observedOverlap) / float64(len(significantPositions))
	}
	rPerm := rand.New(rand.NewSource(baseSeed + 200))
	nullRates := make([]float64, 500)
	for i := range nullRates {
		perm := rPerm.Perm(len(allPositions))
		hits, denom := 0, min(len(significantPositions), len(perm))
		for j := 0; j < denom; j++ {
			if metaBoundaryFlag[allPositions[perm[j]]] {
				hits++
			}
		}
		if denom > 0 {
			nullRates[i] = float64(hits) / float64(denom)
		}
	}
	nullMeanRate := meanOf(nullRates)
	enrichment := 0.0
	if nullMeanRate > 0 {
		enrichment = observedRate / nullMeanRate
	}
	changePointOverlapTSV := "SignificantChangePoints\tObservedOverlapRate\tNullMeanOverlapRate\tEnrichment\n" +
		fmt.Sprintf("%d\t%.9f\t%.9f\t%.9f\n", len(significantPositions), observedRate, nullMeanRate, enrichment)

	// ---- sections 36-43: discrete clustering ----
	degrees := lineregime.ComputeD1Degrees(flatTokens(records))
	extVectors := make([][]float64, len(primaryWindows))
	for i, w := range primaryWindows {
		extVectors[i] = extendedVector(primaryProfiles[i], windowMeanDegree(w.Tokens, degrees), windowNearRepeatRate(w.Tokens))
	}
	var discoveryIdx, validationIdx []int
	for i, w := range primaryWindows {
		if split.Discovery[w.Folio] {
			discoveryIdx = append(discoveryIdx, i)
		}
		if split.Validation[w.Folio] {
			validationIdx = append(validationIdx, i)
		}
	}
	stdVectors := localregimetopology.StandardizeColumns(extVectors, discoveryIdx)

	// K-medoids' medoid-update step is O(n^2) per iteration; with ~5500
	// discovery windows that is billions of operations per K, so fitting
	// uses a deterministic, evenly-spaced subsample (task65 section 74's
	// performance mandate) while dw/vw below are still measured against
	// every discovery/validation window using the fitted medoid vectors.
	const maxFitVectors = 600
	fitIdx := subsampleIndices(discoveryIdx, maxFitVectors)
	fitVectors := make([][]float64, len(fitIdx))
	for i, ix := range fitIdx {
		fitVectors[i] = stdVectors[ix]
	}
	discoveryVectorsAll := make([][]float64, len(discoveryIdx))
	for i, ix := range discoveryIdx {
		discoveryVectorsAll[i] = stdVectors[ix]
	}

	type kResult struct {
		k                                            int
		discoveryWithin, validationWithin, stability float64
	}
	var kResults []kResult
	for k := regimeKMin; k <= regimeKMax; k++ {
		_, fitMedoids := localregimetopology.KMedoids(fitVectors, k, 8)
		var medoidVecs [][]float64
		for _, m := range fitMedoids {
			medoidVecs = append(medoidVecs, fitVectors[m])
		}
		dw := meanDistanceToNearest(discoveryVectorsAll, medoidVecs)
		vw := 0.0
		if len(validationIdx) > 0 {
			var vds []float64
			for _, ix := range validationIdx {
				vds = append(vds, nearestDistance(stdVectors[ix], medoidVecs))
			}
			vw = meanOf(vds)
		}
		stability := clusterStability(fitVectors, k, baseSeed+300+int64(k))
		kResults = append(kResults, kResult{k, dw, vw, stability})
	}
	bestK, bestVal := 0, math.Inf(1)
	for _, kr := range kResults {
		if kr.stability >= 0.5 && kr.validationWithin < bestVal {
			bestK, bestVal = kr.k, kr.validationWithin
		}
	}
	clusteringSupported := bestK > 0

	clusterSelTSV := "K\tDiscoveryWithinDistance\tValidationWithinDistance\tStability\tSelected\n"
	clusterStabTSV := "K\tStability\tStatus\n"
	for _, kr := range kResults {
		sel := "NO"
		if kr.k == bestK {
			sel = "YES"
		}
		clusterSelTSV += fmt.Sprintf("%d\t%.9f\t%.9f\t%.9f\t%s\n", kr.k, kr.discoveryWithin, kr.validationWithin, kr.stability, sel)
		status := "STABLE"
		if kr.stability < 0.5 {
			status = "UNSTABLE"
		}
		clusterStabTSV += fmt.Sprintf("%d\t%.9f\t%s\n", kr.k, kr.stability, status)
	}
	if !clusteringSupported {
		clusterSelTSV += "NOT_APPLICABLE\tNOT_APPLICABLE\tNOT_APPLICABLE\tNOT_APPLICABLE\tDISCRETE_REGIMES_NOT_SUPPORTED\n"
	}

	strideForNonOverlap := primaryWindow / primaryStep
	var nonOverlapIdx []int
	for i := 0; i < len(primaryWindows); i += strideForNonOverlap {
		nonOverlapIdx = append(nonOverlapIdx, i)
	}

	assignTSV := "WindowIndex\tPositionGlobal\tFolio\tCluster\n"
	transTSV := "From\tTo\tCount\tNullMeanCount\n"
	dwellTSV := "Cluster\tMeanDwellLength\tRuns\tNullMeanDwellLength\n"
	withinDriftTSV := "Cluster\tLag\tMeanDistance\n"
	var finalAssign []int
	if clusteringSupported {
		_, medoids := localregimetopology.KMedoids(fitVectors, bestK, 8)
		var medoidVectors [][]float64
		for _, m := range medoids {
			medoidVectors = append(medoidVectors, fitVectors[m])
		}
		finalAssign = make([]int, len(stdVectors))
		for i, v := range stdVectors {
			best, bestD := 0, math.Inf(1)
			for mi, mv := range medoidVectors {
				if d := euclid(v, mv); d < bestD {
					bestD, best = d, mi
				}
			}
			finalAssign[i] = best
		}
		for i, w := range primaryWindows {
			assignTSV += fmt.Sprintf("%d\t%d\t%s\t%d\n", i, w.StartGlobal, w.Folio, finalAssign[i])
		}

		seq := make([]int, len(nonOverlapIdx))
		for i, ix := range nonOverlapIdx {
			seq[i] = finalAssign[ix]
		}
		trans := map[[2]int]int{}
		for i := 0; i+1 < len(seq); i++ {
			trans[[2]int{seq[i], seq[i+1]}]++
		}
		rShuf := rand.New(rand.NewSource(baseSeed + 400))
		nullTrans := map[[2]int][]int{}
		for rep := 0; rep < 50; rep++ {
			shuf := append([]int(nil), seq...)
			rShuf.Shuffle(len(shuf), func(a, b int) { shuf[a], shuf[b] = shuf[b], shuf[a] })
			counts := map[[2]int]int{}
			for i := 0; i+1 < len(shuf); i++ {
				counts[[2]int{shuf[i], shuf[i+1]}]++
			}
			for k := range trans {
				nullTrans[k] = append(nullTrans[k], counts[k])
			}
		}
		for a := 0; a < bestK; a++ {
			for b := 0; b < bestK; b++ {
				key := [2]int{a, b}
				nm := meanOf(intsToFloats(nullTrans[key]))
				transTSV += fmt.Sprintf("%d\t%d\t%d\t%.9f\n", a, b, trans[key], nm)
			}
		}

		dwells := runLengths(seq)
		shufOnce := append([]int(nil), seq...)
		rShuf.Shuffle(len(shufOnce), func(a, b int) { shufOnce[a], shufOnce[b] = shufOnce[b], shufOnce[a] })
		nullDwells := runLengths(shufOnce)
		for c := 0; c < bestK; c++ {
			dwellTSV += fmt.Sprintf("%d\t%.9f\t%d\t%.9f\n", c, meanOf(intsToFloats(dwells[c])), len(dwells[c]), meanOf(intsToFloats(nullDwells[c])))
		}

		for c := 0; c < bestK; c++ {
			var members []lineregime.Profile
			for i, w := range primaryWindows {
				if finalAssign[i] == c {
					members = append(members, primaryProfiles[i])
				}
				_ = w
			}
			if len(members) < 10 {
				continue
			}
			for lag := 1; lag <= 5 && lag < len(members); lag++ {
				var ds []float64
				for i := 0; i+lag < len(members); i++ {
					ds = append(ds, lineregime.ProfileDistance(members[i], members[i+lag]))
				}
				withinDriftTSV += fmt.Sprintf("%d\t%d\t%.9f\n", c, lag, meanOf(ds))
			}
		}
	} else {
		assignTSV += "NOT_APPLICABLE\tNOT_APPLICABLE\tNOT_APPLICABLE\tNOT_APPLICABLE\n"
		transTSV += "NOT_APPLICABLE\tNOT_APPLICABLE\tNOT_APPLICABLE\tNOT_APPLICABLE\n"
		dwellTSV += "NOT_APPLICABLE\tNOT_APPLICABLE\tNOT_APPLICABLE\tNOT_APPLICABLE\n"
		withinDriftTSV += "NOT_APPLICABLE\tNOT_APPLICABLE\tNOT_APPLICABLE\n"
	}

	// ---- section 46-47: distant recurrence without clustering ----
	recurrenceThresholdPages := 5
	rRec := rand.New(rand.NewSource(baseSeed + 500))
	sampleSize := 300
	if sampleSize > len(primaryWindows) {
		sampleSize = len(primaryWindows)
	}
	pageIdxLookup := map[string]int{}
	for i, pg := range pageOrder {
		pageIdxLookup[pg] = i
	}
	pageIndexOfWindow := make([]int, len(primaryWindows))
	for i, w := range primaryWindows {
		pageIndexOfWindow[i] = pageIdxLookup[w.Folio]
	}
	// The null must compare a MINIMUM to a MINIMUM over the same-size
	// candidate pool, not a minimum to one random draw: the minimum of
	// many distances is always far smaller than a single random distance
	// even under pure noise, purely as an order-statistic artifact. The
	// null here instead reassigns profiles to positions at random (once,
	// globally) and recomputes the identical nearest-distant search, so
	// the only thing removed is genuine spatial recurrence, not the
	// minimum-of-many effect (task65 section 47).
	shuffledProfiles := append([]lineregime.Profile(nil), primaryProfiles...)
	rand.New(rand.NewSource(baseSeed+502)).Shuffle(len(shuffledProfiles), func(i, j int) {
		shuffledProfiles[i], shuffledProfiles[j] = shuffledProfiles[j], shuffledProfiles[i]
	})
	var nearestDs, nullDistantDs []float64
	for _, i := range rRec.Perm(len(primaryWindows))[:sampleSize] {
		best, bestNull := math.Inf(1), math.Inf(1)
		for j := range primaryWindows {
			if abs(pageIndexOfWindow[i]-pageIndexOfWindow[j]) < recurrenceThresholdPages {
				continue
			}
			if d := lineregime.ProfileDistance(primaryProfiles[i], primaryProfiles[j]); d < best {
				best = d
			}
			if d := lineregime.ProfileDistance(primaryProfiles[i], shuffledProfiles[j]); d < bestNull {
				bestNull = d
			}
		}
		if !math.IsInf(best, 1) {
			nearestDs = append(nearestDs, best)
		}
		if !math.IsInf(bestNull, 1) {
			nullDistantDs = append(nullDistantDs, bestNull)
		}
	}
	recurrenceTSV := "Metric\tMean\tMedian\n" +
		fmt.Sprintf("ObservedNearestDistantDistance\t%.9f\t%.9f\n", meanOf(nearestDs), medianOf(nearestDs)) +
		fmt.Sprintf("NullNearestUnderShuffledPositions\t%.9f\t%.9f\n", meanOf(nullDistantDs), medianOf(nullDistantDs)) +
		fmt.Sprintf("EffectSize_NullMinusObserved\t%.9f\tNA\n", meanOf(nullDistantDs)-meanOf(nearestDs))

	// ---- sections 48-50: natural, synthetic and Task62 controls ----
	evalControl := func(name string, toks [][]string, trueBoundaries []int) string {
		recs := toRecords(toks, "SYN")
		ws := localregimetopology.BuildWindows(recs, primaryWindow, primaryStep)
		profs := make([]lineregime.Profile, len(ws))
		for i, w := range ws {
			profs[i] = lineregime.ComputeProfile(w.Tokens, giant)
		}
		decay := decayCurve("TOKEN", primaryWindow, profs, min(maxLagSteps, max(1, len(profs)/3)), baseSeed+600)
		cl, _ := correlationLength(decay, 0.5)
		cpCount := changePointCount(recs, primaryWindow, giant, threshold95)
		recovery := 0.0
		if len(trueBoundaries) > 0 {
			recovery = boundaryRecoveryRate(recs, primaryWindow, giant, threshold95, trueBoundaries, primaryWindow*2)
		}
		degs := lineregime.ComputeD1Degrees(toks)
		vecs := make([][]float64, len(ws))
		for i, w := range ws {
			vecs[i] = extendedVector(profs[i], windowMeanDegree(w.Tokens, degs), windowNearRepeatRate(w.Tokens))
		}
		stdVecs := localregimetopology.StandardizeColumns(vecs, allIndices(len(vecs)))
		stability := clusterStability(stdVecs, 3, baseSeed+601)
		return fmt.Sprintf("%s\t%.9f\t%d\t%.9f\t%.9f\n", name, cl*float64(primaryStep), cpCount, recovery, stability)
	}
	header := "Corpus\tCorrelationLengthTokens\tChangePointCountAboveP95\tBoundaryRecoveryRate\tClusterStabilityK3\n"
	syntheticControlTSV := header +
		evalControl("STATIONARY", stationaryTokens, nil) +
		evalControl("SMOOTH_DRIFT", driftTokens, nil) +
		evalControl("DISCRETE", discreteTokens, discreteBoundaries) +
		evalControl("MIXED", mixedTokens, mixedBoundaries)

	naturalControlTSV := header
	for _, ns := range []struct{ name, path string }{
		{"Doyle", "data_test/pg2097-2.txt"}, {"Longfellow", "data_test/pg30795-mod.txt"}, {"Astafiev", "data_test/astafiev-1000-culinar-receipts-prepared.txt"},
	} {
		tb, lerr := loadNatural(ns.path)
		if lerr != nil {
			continue
		}
		var flat [][]string
		for _, line := range tb {
			flat = append(flat, line...)
		}
		naturalControlTSV += evalControl(ns.name, flat, nil)
	}

	task62TSV := header
	if gTokens, lerr := loadGeneratedTokens("experiments/token-formation-v1/generated/POSITION_MARKOV_1-000.txt"); lerr == nil && len(gTokens) > 0 {
		task62TSV += evalControl("TASK62_G_ONLY", gTokens, nil)
	}

	// ---- section 64: final classification ----
	localStructure := "NOT_CONFIRMED"
	if len(tokenDecay) > 0 && tokenDecay[0].nullMean > tokenDecay[0].mean {
		localStructure = "CONFIRMED"
	}
	samePageExcess := 0.0
	if len(lineLagRows) > 0 {
		samePageExcess = lineLagRows[0].nullMean - lineLagRows[0].mean
	}
	topology := "UNRESOLVED"
	switch {
	case clusteringSupported && withinClusterDriftMaterial(withinDriftTSV):
		topology = "MIXED_DRIFT_AND_STATES"
	case clusteringSupported:
		topology = "DISCRETE_REGIMES"
	case localStructure == "CONFIRMED" && realCount95 <= 2*driftCount95+2:
		topology = "CONTINUOUS_DRIFT"
	case localStructure == "NOT_CONFIRMED":
		topology = "STATIONARY"
	}
	metadataClass := "NOT_METADATA_EXPLAINED"
	switch {
	case samePageExcess <= 0:
		metadataClass = "MOSTLY_METADATA_EXPLAINED"
	case samePageExcess < (tokenDecay[0].nullMean-tokenDecay[0].mean)*0.3:
		metadataClass = "PARTIALLY_METADATA_EXPLAINED"
	}

	manifest := map[string]any{
		"task": "Task65", "corpus_path": corpusPath, "corpus_sha256": corpusSHA, "ivtff_path": ivtffPath,
		"metadata_ok": metaOK, "primary_window": primaryWindow, "primary_step": primaryStep,
		"sensitivity_windows": sensitivityWindows, "max_lag_steps": maxLagSteps,
		"regime_k_range": []int{regimeKMin, regimeKMax}, "bootstrap_reps": bootstrapReps,
		"task64_base_seed": task64BaseSeed, "base_seed": baseSeed,
		"train_frac": trainFrac, "val_frac": valFrac, "discovery_frac": discoveryFrac,
		"task64_reproduction": provenanceStatus, "task64_discovery_effect": discoveryEffect, "task64_replication_effect": replicationEffect,
		"selected_k": bestK, "clustering_supported": clusteringSupported,
		"change_point_threshold_p95": threshold95, "change_point_threshold_p99": threshold99,
		"local_structure": localStructure, "topology": topology, "metadata_class": metadataClass, "task64_split_interpretation": splitInterpretation,
	}
	commit, dirty := gitInfo()
	manifest["git_commit"] = commit
	manifest["git_dirty"] = dirty
	manifestBytes, _ := json.MarshalIndent(manifest, "", "  ")

	design := designDoc()
	var kSummaryB strings.Builder
	for _, kr := range kResults {
		fmt.Fprintf(&kSummaryB, "K=%d stability=%.3f validationWithin=%.3f\n", kr.k, kr.stability, kr.validationWithin)
	}
	report := reportDoc(localStructure, topology, metadataClass, splitInterpretation, discoveryEffect, replicationEffect, samePageExcess, bestK, clusteringSupported, kSummaryB.String(), meanOf(nearestDs), meanOf(nullDistantDs))

	files := map[string]string{
		"LOCAL_REGIME_TOPOLOGY_DESIGN.md":   design,
		"DESIGN_FROZEN":                     "frozen\n",
		"manifest.json":                     string(manifestBytes) + "\n",
		"LOCAL_REGIME_DECAY.tsv":            decayTSV,
		"CORRELATION_LENGTH.tsv":            corrLenTSV,
		"BOUNDARY_DISCONTINUITY.tsv":        boundaryTSV,
		"METADATA_EFFECTS.tsv":              metadataTSV,
		"HIERARCHICAL_VARIANCE.tsv":         hvTSV,
		"METADATA_CONDITIONED_DECAY.tsv":    condTSV,
		"LOCAL_EFFECT_BY_REGION.tsv":        regionTSV,
		"TASK64_SPLIT_DIAGNOSIS.tsv":        splitDiag + "\n# standardized comparison\n" + splitStandardized + fmt.Sprintf("\n# interpretation: %s\n", splitInterpretation),
		"CHANGE_POINTS.tsv":                 changePointsTSV,
		"CHANGE_POINT_METADATA_OVERLAP.tsv": changePointOverlapTSV,
		"CHANGE_POINT_NULLS.tsv":            changePointNullsTSV,
		"REGIME_CLUSTER_SELECTION.tsv":      clusterSelTSV,
		"REGIME_CLUSTER_STABILITY.tsv":      clusterStabTSV,
		"REGIME_ASSIGNMENTS.tsv":            assignTSV,
		"REGIME_TRANSITIONS.tsv":            transTSV,
		"REGIME_DWELL.tsv":                  dwellTSV,
		"WITHIN_REGIME_DRIFT.tsv":           withinDriftTSV,
		"DISTANT_REGIME_RECURRENCE.tsv":     recurrenceTSV,
		"NATURAL_CONTROL_TOPOLOGY.tsv":      naturalControlTSV,
		"SYNTHETIC_CONTROL_TOPOLOGY.tsv":    syntheticControlTSV,
		"TASK62_STATIONARY_CONTROL.tsv":     task62TSV,
		"REPORT.md":                         report,
	}
	for name, content := range files {
		if err := put(filepath.Join(outDir, name), content); err != nil {
			return err
		}
	}
	return nil
}

func metadataBoundaryNear(records []localregimetopology.TokenRecord, pos, tolerance int) string {
	lo, hi := pos-tolerance, pos+tolerance
	if lo < 0 {
		lo = 0
	}
	if hi >= len(records) {
		hi = len(records) - 1
	}
	base := records[pos]
	for i := lo; i <= hi; i++ {
		if records[i].Folio != base.Folio {
			return "PAGE"
		}
	}
	for i := lo; i <= hi; i++ {
		if records[i].Currier != base.Currier {
			return "CURRIER"
		}
		if records[i].Hand != base.Hand {
			return "HAND"
		}
		if records[i].Section != base.Section {
			return "SECTION"
		}
	}
	return "NONE"
}

func scoreValues(cps []localregimetopology.ChangePoint) []float64 {
	out := make([]float64, len(cps))
	for i, c := range cps {
		out[i] = c.Score
	}
	return out
}

func euclid(a, b []float64) float64 {
	s := 0.0
	for i := range a {
		d := a[i] - b[i]
		s += d * d
	}
	return math.Sqrt(s)
}

// subsampleIndices deterministically thins idx to at most capN entries,
// evenly spaced, so K-medoids' O(n^2) medoid-update step stays bounded
// regardless of corpus size (task65 section 74).
func subsampleIndices(idx []int, capN int) []int {
	if len(idx) <= capN || capN <= 0 {
		return idx
	}
	out := make([]int, 0, capN)
	step := float64(len(idx)) / float64(capN)
	for i := 0; i < capN; i++ {
		out = append(out, idx[int(float64(i)*step)])
	}
	return out
}

func nearestDistance(v []float64, medoids [][]float64) float64 {
	best := math.Inf(1)
	for _, m := range medoids {
		if d := euclid(v, m); d < best {
			best = d
		}
	}
	return best
}

func meanDistanceToNearest(vectors, medoids [][]float64) float64 {
	var ds []float64
	for _, v := range vectors {
		ds = append(ds, nearestDistance(v, medoids))
	}
	return meanOf(ds)
}

func clusterStability(vectors [][]float64, k int, seed int64) float64 {
	if len(vectors) < k*2 {
		return 0
	}
	orig, _ := localregimetopology.KMedoids(vectors, k, 8)
	r := rand.New(rand.NewSource(seed))
	var agreements []float64
	for rep := 0; rep < 15; rep++ {
		idx := make([]int, len(vectors))
		sample := make([][]float64, len(vectors))
		for i := range sample {
			idx[i] = r.Intn(len(vectors))
			sample[i] = vectors[idx[i]]
		}
		resAssign, _ := localregimetopology.KMedoids(sample, k, 8)
		agreements = append(agreements, coClusterAgreement(orig, resAssign, idx))
	}
	return meanOf(agreements)
}

func intsToFloats(xs []int) []float64 {
	out := make([]float64, len(xs))
	for i, x := range xs {
		out[i] = float64(x)
	}
	return out
}

func runLengths(seq []int) map[int][]int {
	out := map[int][]int{}
	i := 0
	for i < len(seq) {
		j := i
		for j+1 < len(seq) && seq[j+1] == seq[i] {
			j++
		}
		out[seq[i]] = append(out[seq[i]], j-i+1)
		i = j + 1
	}
	return out
}

func allIndices(n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = i
	}
	return out
}

func withinClusterDriftMaterial(withinDriftTSV string) bool {
	lines := strings.Split(strings.TrimSpace(withinDriftTSV), "\n")
	for _, ln := range lines[1:] {
		f := strings.Split(ln, "\t")
		if len(f) < 3 {
			continue
		}
		if f[1] == "1" {
			if v, err := strconv.ParseFloat(f[2], 64); err == nil && v > 0.15 {
				return true
			}
		}
	}
	return false
}

func designDoc() string {
	return fmt.Sprintf(`# Task65 frozen design (LOCAL_REGIME_TOPOLOGY_DESIGN)

Analysis units: the unified manuscript coordinate x (GlobalIndex, Line,
PageIndex, Folio/Currier/Hand/Section) built from %s
in file order, exactly as Tasks58-64 read it; manuscript order is
explicitly NOT assumed to be the original semantic order (section 5).

Currier A/B is IVTFF $L ("Currier's Language A/B", per the literal
comment attached to $L/$H in data/ZL3b-n.txt's page headers); Hand is
$H; Section is $I (the standard illustration-type code: H=Herbal,
P=Pharmaceutical, S=Stars, A=Astronomical, B=Biological, C=Cosmological,
T=Text, Z=Zodiac). This intentionally differs from the "$C" field other
packages in this repo call Currier (a separate, finer scribal-hand
numbering) - see the code comment on metadata() for the full rationale.

Feature set: the authoritative Task64 Profile (MeanLen, GiantFrac,
TopInit, TopFinal, TypeEnt over internal/lineregime.ComputeProfile) is
the PRIMARY feature/distance for decay, correlation length and boundary
discontinuity (section 11, reusing Task64's distance rather than
inventing a new one). Clustering/hierarchical-variance use a SECONDARY,
7-dimensional extended vector (the 5 Profile dims plus mean d1
neighborhood degree and local near-repeat rate), z-scored using
DISCOVERY-fold windows only (section 10).

Window sizes: primary W=%d tokens, step=W/4=%d; sensitivity W=10/40/80,
same step rule; additional physical-line and page scales reuse Task64's
own Line/page grouping. Fixed before Voynich results were inspected
(section 6).

Lag range: token lag 1..%d steps; line/page lag 1..10 (matching Task64's
persistence range).

Change-point algorithms: (1) a deterministic non-overlapping-window
profile-distance scan (ScanChangePoints); (2) the classical CUSUM
maximum-cumulative-deviation statistic applied to that same score series.
Significance is calibrated against a STATIONARY synthetic null's P95/P99
score, never against the Voynich data itself (section 19).

Clustering: deterministic k-medoids (PAM-lite), K=2..8, medoids fit on
DISCOVERY-fold windows only; K is selected by held-out (VALIDATION-fold)
mean distance-to-medoid among K values whose bootstrap co-cluster
stability is >=0.5, never by agreement with Currier/Hand/Section/Page
(section 37/39 - metadata is compared only after this freeze).

Null models: shuffle-order nulls for decay curves; label-preserving
shuffle nulls for transition/dwell; STATIONARY/SMOOTH_DRIFT/DISCRETE/
MIXED synthetic corpora (deterministic contiguous partitions of the
corpus's own sorted vocabulary into "flavors", never hand-picked or
fit to any Voynich effect) for change-point calibration and pipeline
validation (sections 19-22).

Metadata controls: page-level aggregate profiles (not per-window) for
between/within Currier/Hand/Section comparisons, so N is pages, not
pseudo-replicated windows; strata under %d pages or %d tokens are
NOT_APPLICABLE / INSUFFICIENT_DATA (section 71).

Discovery/replication protocol: IDENTICAL to Task64 - contiguous folio
blocks in manuscript order, train/validation/test = %.0f%%/%.0f%%/%.0f%%
of pages, discovery = train+validation, replication = test, seed %d+1
for the shared nullMean baseline (section 33's byte/metric reproduction
requirement).

Acceptance criteria: as listed in task65 sections 53-56/64; thresholds
(0.5 stability, 30%% variance-explained boundary for
PARTIALLY_METADATA_EXPLAINED, +-0.005 Delta gap for COMPOSITION_EXPLAINED)
are fixed here, before Voynich results were inspected.
`, corpusPath, primaryWindow, primaryStep, maxLagSteps, minStrataPages, minStrataN, trainFrac*100, valFrac*100, (1-trainFrac-valFrac)*100, task64BaseSeed)
}

func reportDoc(localStructure, topology, metadataClass, splitInterpretation string, discoveryEffect, replicationEffect, samePageExcess float64, bestK int, clusteringSupported bool, kSummary string, recurrenceObserved, recurrenceNull float64) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Task65 report: local regime topology, drift and change-point analysis\n\n")
	fmt.Fprintf(&b, "LOCAL_STRUCTURE: **%s**\nTOPOLOGY: **%s**\nMETADATA: **%s**\nTASK64_SPLIT: **%s**\n\n",
		localStructure, topology, metadataClass, splitInterpretation)
	fmt.Fprintf(&b, "R is called a *regime* or *local distributional state* throughout, never a\n"+
		"topic, language state or cipher key (task65's own safeguard, echoing task64\n"+
		"section 65). No decipherment, language or semantic claim is made. A discrete\n"+
		"or hierarchical finding does not rule out the scribal/layout alternative\n"+
		"(line filling, glyph-choice drift, copying batches, page planning) - it is\n"+
		"discussed explicitly below rather than assumed away (section 51).\n\n")

	fmt.Fprintf(&b, "## 1-2. Does local similarity depend on distance, and does it decay smoothly?\n\n")
	fmt.Fprintf(&b, "LOCAL_REGIME_DECAY.tsv reports mean/median/CI/null-mean profile distance for "+
		"every token/line/page lag, for the primary W=%d window and the W=10/40/80 "+
		"sensitivity windows, against a shuffle-order null - LOCAL_STRUCTURE=%s answers "+
		"question 1 directly (whether the lag=1 curve sits below its null). Section 14's "+
		"instruction not to assume exponential decay is honored: no functional form is "+
		"fit here, only the empirical curve.\n\n", primaryWindow, localStructure)

	fmt.Fprintf(&b, "## 3. Are there statistically supported abrupt change points?\n\n")
	fmt.Fprintf(&b, "CHANGE_POINTS.tsv reports every non-overlapping-window boundary's score, "+
		"significance against the STATIONARY synthetic null's P95/P99 (CHANGE_POINT_NULLS.tsv), "+
		"discovery/replication fold and nearest metadata boundary. CHANGE_POINT_METADATA_OVERLAP.tsv "+
		"gives the enrichment of significant change points near real Currier/Hand/Section/page "+
		"boundaries over a position-permutation null.\n\n")

	fmt.Fprintf(&b, "## 4. Do similar regimes recur in distant parts of the manuscript?\n\n")
	fmt.Fprintf(&b, "DISTANT_REGIME_RECURRENCE.tsv compares, for a sample of windows, the nearest "+
		"profile distance among windows >=5 pages away against the same nearest-distance search "+
		"repeated after globally shuffling which profile sits at which position (both are a "+
		"minimum over the same-size candidate pool, so the comparison isn't biased by the fact "+
		"that a minimum over many candidates is always smaller than a single random draw) - "+
		"this is independent of the discrete-clustering result and answers the recurrence "+
		"question even if clustering itself is unstable (section 46). Observed mean nearest-"+
		"distant distance is %.6f against a properly-calibrated null of %.6f: %s.\n\n",
		recurrenceObserved, recurrenceNull, recurrenceVerdict(recurrenceObserved, recurrenceNull))

	fmt.Fprintf(&b, "## 5-6. How much does Currier/Hand/Section/Page explain, and does a residual "+
		"local regime remain after conditioning?\n\n")
	fmt.Fprintf(&b, "METADATA_EFFECTS.tsv (between vs within-group page-profile distance) and "+
		"HIERARCHICAL_VARIANCE.tsv (variance share per factor on MeanTokenLength/GiantFraction) "+
		"give METADATA=%s. METADATA_CONDITIONED_DECAY.tsv recomputes the token-lag decay "+
		"restricted to windows sharing the same Currier/Hand/Section; the SAME_PAGE(LINE) rows "+
		"in that same file are Task64's own within-page line-lag decay, i.e. the strict "+
		"same-page conditioning of section 31. Its lag=1 excess similarity is %.9f: %s.\n\n",
		metadataClass, samePageExcess, samePageInterpretation(samePageExcess))

	fmt.Fprintf(&b, "## 7. What explains Task64's discovery (~0.030) vs replication (~0.003) gap?\n\n")
	fmt.Fprintf(&b, "TASK64_SPLIT_DIAGNOSIS.tsv first reproduces Task64's own numbers "+
		"independently: discovery=%.9f, replication=%.9f (compare to Task64's published "+
		"0.030069436/0.003132198). It then reports each fold's Currier/Hand/Section/page "+
		"composition and a standardized per-stratum comparison; TASK64_SPLIT=%s. "+
		"LOCAL_EFFECT_BY_REGION.tsv gives the region-wise map (Currier/Hand/Section, N/effect/CI/"+
		"reliability) requested by section 70, so if the effect concentrates in one region that "+
		"is visible directly rather than averaged away.\n\n", discoveryEffect, replicationEffect, splitInterpretation)

	fmt.Fprintf(&b, "## Discrete clustering (H2) and mixed drift+states (H4)\n\n")
	if clusteringSupported {
		fmt.Fprintf(&b, "A stable K=%d clustering was selected on held-out (VALIDATION-fold) "+
			"distance-to-medoid among bootstrap-stable K values (REGIME_CLUSTER_SELECTION.tsv/"+
			"REGIME_CLUSTER_STABILITY.tsv), fit on DISCOVERY only and only compared to metadata "+
			"afterward (no label leakage). REGIME_TRANSITIONS.tsv/REGIME_DWELL.tsv compare "+
			"transition/dwell structure to a label-shuffled null on non-overlapping windows "+
			"(avoiding the overlapping-window pseudoreplication warned about in section 7); "+
			"WITHIN_REGIME_DRIFT.tsv tests whether decay still exists inside each cluster "+
			"(section 43/H4). K summary: %s\n\n", bestK, kSummary)
	} else {
		fmt.Fprintf(&b, "No K in 2..8 reached the preregistered >=0.5 bootstrap stability "+
			"threshold on DISCOVERY-fold windows (REGIME_CLUSTER_SELECTION.tsv/"+
			"REGIME_CLUSTER_STABILITY.tsv): DISCRETE_REGIMES is NOT supported by this run "+
			"(section 38). REGIME_ASSIGNMENTS.tsv/REGIME_TRANSITIONS.tsv/REGIME_DWELL.tsv/"+
			"WITHIN_REGIME_DRIFT.tsv are still created, explicitly marked NOT_APPLICABLE "+
			"(section 61), rather than omitted.\n\n")
	}

	fmt.Fprintf(&b, "## Controls\n\n")
	fmt.Fprintf(&b, "SYNTHETIC_CONTROL_TOPOLOGY.tsv validates the pipeline itself: STATIONARY "+
		"should show near-zero correlation length and a change-point count near the "+
		"calibration control's own count (it IS that control); SMOOTH_DRIFT should show a real "+
		"correlation length without an inflated change-point count (section 21); DISCRETE and "+
		"MIXED report boundary-recovery rate against their known, injected boundaries (sections "+
		"20/22). NATURAL_CONTROL_TOPOLOGY.tsv applies the identical pipeline to Doyle/"+
		"Longfellow/Astafiev (section 48, not to classify genre but to see whether drift/regime "+
		"topology is a generic property of any finite natural-language corpus). "+
		"TASK62_STATIONARY_CONTROL.tsv applies it to Task62's frozen G-only generation, which "+
		"should read as close to the STATIONARY control (section 49).\n\n")

	fmt.Fprintf(&b, "## Scope\n\n")
	fmt.Fprintf(&b, "This is a diagnostic study only: no G+R or other generative model was built "+
		"(section 2 - that is explicitly left to a future task once this topology is settled). "+
		"Stages 1-28 were not touched and no Stage29 was added. No claim is made here about "+
		"language, semantics, a specific cipher, or decipherment.\n")
	return b.String()
}

func recurrenceVerdict(observed, null float64) string {
	if observed < null*0.8 {
		return "the real manuscript finds closer distant matches than the shuffled-position null, supporting recurring regimes beyond chance (critical result C)"
	}
	return "the real manuscript does not find closer distant matches than the shuffled-position null once the minimum-of-many-candidates bias is corrected for, so this analysis does NOT support recurring discrete regimes beyond chance on its own"
}

func samePageInterpretation(excess float64) string {
	if excess > 0 {
		return "decay survives strictly within one page, which cannot be explained by Currier/Hand/Section/page-level composition since page is held fixed (task65 section 31, critical result B)"
	}
	return "no measurable same-page excess survives; the broader local regime is consistent with page-level (or coarser) composition rather than genuine sub-page structure (critical result A)"
}

func compositionInline(c map[string]int) string {
	var parts []string
	for _, k := range sortedKeys(c) {
		lbl := k
		if lbl == "" {
			lbl = "UNKNOWN"
		}
		parts = append(parts, fmt.Sprintf("%s=%d", lbl, c[k]))
	}
	return strings.Join(parts, ";")
}

// foliosWith returns the set of folios (pages) among lines whose valueOf
// equals level, for restricting a fold's lines to one metadata stratum.
func foliosWith(lines []lineregime.Line, valueOf func(l lineregime.Line) string, level string) map[string]bool {
	out := map[string]bool{}
	for _, l := range lines {
		if valueOf(l) == level {
			out[l.Folio] = true
		}
	}
	return out
}

func linesByValue(lines []lineregime.Line, valueOf func(l lineregime.Line) string) map[string][]lineregime.Line {
	out := map[string][]lineregime.Line{}
	for _, l := range lines {
		v := valueOf(l)
		if v == "" {
			continue
		}
		out[v] = append(out[v], l)
	}
	return out
}
